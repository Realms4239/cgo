package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/api"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// runServer starts dashboard+API with the campagne core attached.
func runServer(ctx context.Context, addr string) error {
	live := campagne.NewLive()

	var mtx *campagne.Matrix
	var mtxMu sync.Mutex
	getMtx := func() *campagne.Matrix {
		mtxMu.Lock()
		defer mtxMu.Unlock()
		return mtx
	}
	setMtx := func(m *campagne.Matrix) {
		mtxMu.Lock()
		mtx = m
		mtxMu.Unlock()
	}

	// Single pump goroutine for the lifetime of the server (H1 fix):
	// reads current mtx via getMtx each tick, no leak on restart, no flap.
	go pumpSnapshots(ctx, live, getMtx)

	startFn := func(profiles []string, reps int) error {
		if getMtx() != nil {
			getMtx().Stop()
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		m, err := campagne.StartMatrix(ctx, profiles, reps, deps, live, "data/runs")
		if err != nil {
			return err
		}
		setMtx(m)
		return nil
	}
	stopFn := func() {
		if getMtx() != nil {
			getMtx().Stop()
		}
		// pump will notice mtx.IsRunning()==false on next tick (≤100ms) and
		// set live false atomically; we also set immediately for snappier UX.
		live.SetRunning(false)
	}
	// ARG.md edge control — the shaping lever on the gateway's egress hop,
	// same primitives the campagne cells use (manual mode for the DSI).
	shapeFn := func(qdiscName string, capMbps float64) error {
		deps := campagne.ProdDeps()
		// reset first — stale handles from a campagne cell make 'replace'
		// fail; the lever owns a deterministic clean state (manual mode has
		// no netem: pure AQM+bandwidth at root, that is the DSI's lever)
		_, _ = deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root")
		if qdiscName == "none" {
			return nil
		}
		return qdisc.ApplyShaper(deps.TCShaper, deps.ShaperIf, model.Qdisc(qdiscName), capMbps, 100)
	}
	handler := api.New(api.Deps{
		GetSnap:   func() any { return live.Get() },
		StartFn:   startFn,
		StopFn:    stopFn,
		ShapeFn:   shapeFn,
		RunningFn: func() bool { return getMtx() != nil && getMtx().IsRunning() },
	})
	srv := &http.Server{Addr: addr, Handler: handler}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("cgo dashboard on %s", addr)
	select {
	case <-ctx.Done():
		stopFn()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		handler.CloseHub() // stop the 10 Hz broadcast ticker — Server.Close never does
		return nil
	case err := <-errCh:
		return err
	}
}

// pumpSnapshots mirrors matrix progress into the broadcast snapshot at 10 Hz.
// Uses Live.SetRunning atomically to avoid Get+Modify+Set lost-update race (H2).
// getMtx is a func to read the current matrix pointer under lock (single pump, H1).
func pumpSnapshots(ctx context.Context, live *campagne.Live, getMtx func() *campagne.Matrix) {
	tk := time.NewTicker(time.Second / 10)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			live.SetRunning(mtxRunning(getMtx()))
		}
	}
}

func mtxRunning(m *campagne.Matrix) bool { return m.IsRunning() }

var _ = os.Getenv // keep os import for future env-driven config
