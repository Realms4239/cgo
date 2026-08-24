package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Realms4239/cgo/pkg/api"
	"github.com/Realms4239/cgo/pkg/campagne"
)

// runServer starts dashboard+API with the campagne core attached.
func runServer(ctx context.Context, addr string) error {
	live := campagne.NewLive()

	var mtx *campagne.Matrix
	startFn := func(profiles []string, reps int) error {
		if mtx != nil {
			mtx.Stop()
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		m, err := campagne.StartMatrix(ctx, profiles, reps, deps, live, "data/runs")
		if err != nil {
			return err
		}
		mtx = m
		go pumpSnapshots(ctx, live, mtx)
		return nil
	}
	stopFn := func() {
		if mtx != nil {
			mtx.Stop()
		}
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
func pumpSnapshots(ctx context.Context, live *campagne.Live, mtx *campagne.Matrix) {
	tk := time.NewTicker(time.Second / 10)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s := live.Get()
			s.Running = mtxRunning(mtx)
			live.Set(s)
		}
	}
}

func mtxRunning(m *campagne.Matrix) bool {
	if m == nil {
		return false
	}
	return m.Running
}

var _ = os.Getenv // keep os import for future env-driven config
