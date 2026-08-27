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
	handler := api.New(api.Deps{
		GetSnap: func() any { return live.Get() },
		StartFn: startFn,
		StopFn:  stopFn,
	})
	srv := &http.Server{Addr: addr, Handler: handler}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("cgo dashboard on %s", addr)
	select {
	case <-ctx.Done():
		stopFn()
		return srv.Close()
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
