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
	var pumpCancel context.CancelFunc
	var pumpMu sync.Mutex

	startPump := func(parent context.Context, cur *campagne.Matrix) {
		pumpMu.Lock()
		if pumpCancel != nil {
			pumpCancel()
		}
		pCtx, cancel := context.WithCancel(parent)
		pumpCancel = cancel
		pumpMu.Unlock()
		go pumpSnapshots(pCtx, live, cur)
	}

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
		startPump(ctx, m)
		return nil
	}
	stopFn := func() {
		if getMtx() != nil {
			getMtx().Stop()
		}
		pumpMu.Lock()
		if pumpCancel != nil {
			pumpCancel()
			pumpCancel = nil
		}
		pumpMu.Unlock()
		// ensure Live reflects stopped state immediately (no 100ms lag)
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
func pumpSnapshots(ctx context.Context, live *campagne.Live, mtx *campagne.Matrix) {
	tk := time.NewTicker(time.Second / 10)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			live.SetRunning(mtxRunning(mtx))
		}
	}
}

func mtxRunning(m *campagne.Matrix) bool { return m.IsRunning() }

var _ = os.Getenv // keep os import for future env-driven config
