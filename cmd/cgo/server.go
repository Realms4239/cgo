package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/api"
	"github.com/Realms4239/cgo/pkg/campagne"
	"github.com/Realms4239/cgo/pkg/doctor"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/probe"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// runServer starts dashboard+API with the campagne core attached.
// mode: "full" (Linux bench/edge) or "observe" (Windows workstation).
func runServer(ctx context.Context, addr, mode string) error {
	live := campagne.NewLive()

	// quarantined cells land in the operator journal (Q13)
	campagne.OnQuarantine = func(runID string, eventID int, profile, qdisc, cc string) {
		api.RecordEvent("quarantaine", fmt.Sprintf("%s évènement %d %s/%s/%s — cellule invalidée par les portes", runID, eventID, profile, qdisc, cc))
	}

	if mode == "full" {
		// startup self-check (Q18): a crashed session can leave a root qdisc
		// behind; reset-then-apply semantics start from a clean shaper.
		deps := campagne.ProdDeps()
		_, _ = deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root")
	}

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

	// watchCtl — the surveillance toggle shared between the campagne path
	// (startFn auto-stops watch) and the API path (POST /api/watch). Both
	// mutate the same state; one mutex or nothing (a bare bool here raced
	// between the two call paths).
	watchCtl := struct {
		mu   sync.Mutex
		stop func()
		on   bool
	}{stop: func() {}}
	watchStart := func() error {
		watchCtl.mu.Lock()
		defer watchCtl.mu.Unlock()
		if watchCtl.on {
			return nil
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		stop := campagne.StartWatch(ctx, deps)
		watchCtl.stop = stop
		watchCtl.on = true
		return nil
	}
	watchStop := func() {
		watchCtl.mu.Lock()
		defer watchCtl.mu.Unlock()
		if watchCtl.on {
			watchCtl.stop()
			watchCtl.on = false
		}
	}
	startFn := func(o api.RunOpts) error {
		watchStop() // campagne and surveillance are mutually exclusive
		if m := getMtx(); m != nil {
			m.Stop()
			// drain — the old matrix's cells may still be mid-tc-call; starting
			// under them races the shaper and fails the new cells wholesale
			for i := 0; i < 30 && m.IsRunning(); i++ {
				time.Sleep(100 * time.Millisecond)
			}
		}
		deps := campagne.ProdDeps()
		deps.OnSnap = func(s campagne.Snapshot) { live.Set(s) }
		m, err := campagne.StartMatrix(ctx, o.Profiles, o.Reps, deps, live, "data/runs")
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
	shapeFn := func(r api.ShapeReq) error {
		deps := campagne.ProdDeps()
		// reset first — stale handles from a campagne cell make 'replace'
		// fail; the lever owns a deterministic clean state
		_, _ = deps.TCShaper.Run("qdisc", "del", "dev", deps.ShaperIf, "root")
		if r.Qdisc == "none" {
			return nil
		}
		// conditions du lien (Q8): netem at root, shaper stacked as child
		if r.DelayMs > 0 || r.JitterMs > 0 || r.LossPct > 0 {
			if err := qdisc.ApplyNetem(deps.TCShaper, deps.ShaperIf, r.DelayMs, r.JitterMs, r.LossPct); err != nil {
				return err
			}
		}
		return qdisc.ApplyShaper(deps.TCShaper, deps.ShaperIf, model.Qdisc(r.Qdisc), r.CapMbps, 100)
	}
	handler := api.New(api.Deps{
		GetSnap:   func() any { return live.Get() },
		StartFn:   startFn,
		StopFn:    stopFn,
		ShapeFn:   shapeFn,
		RunningFn: func() bool { return getMtx() != nil && getMtx().IsRunning() },
		Mode:      mode,
		Version:   version,
		DoctorFn: func() any {
			m, checks := doctor.Report(mode)
			return map[string]any{"mode": m, "checks": checks}
		},
		BurstFn: func(cc string, seconds int) error {
			deps := campagne.ProdDeps()
			ctxB, cancel := context.WithTimeout(ctx, time.Duration(seconds+3)*time.Second)
			defer cancel()
			b, err := probe.BulkSendTo(ctxB, deps.BulkAddr, cc)
			if err != nil && b == 0 {
				return err
			}
			g := float64(b) * 8 / 1e6 / float64(seconds)
			if g > 2500 { // same ring clamp as the campagne cells
				g = 2500
			}
			live.Set(campagne.Snapshot{
				Phase: "burst", Profile: "burst", CC: cc,
				GoodputMbps: math.Round(g*10) / 10,
				Drops:       0, Running: false,
			})
			return nil
		},
		WatchFn: func(on bool) error {
			if on {
				if err := watchStart(); err != nil {
					return err
				}
			} else {
				watchStop()
			}
			return nil
		},
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
