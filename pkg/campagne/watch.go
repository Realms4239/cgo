package campagne

import (
	"context"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// OnWatchAlert — sentinelle deadline : la surveillance dépasse l'objectif
// small p95 de façon soutenue. Journal opérateur côté hôte.
var OnWatchAlert func(p95, deadlineMs float64)

// StartWatch exécute la boucle de sonde non intrusive (ping + petit objet, sans bulk —
// l'audit reste côté client). Le mur reste vivant hors
// campagne, ce qui rend l'effet du levier de façonnage visible en temps réel.
// Les instantanés portent la phase "surveil" et ne touchent jamais le shaper.
// DeadlineMs > 0 : sentinelle — 10 dépassements consécutifs de small p95
// journalisent une alerte honnête (pas de notification externe).
func StartWatch(base context.Context, d Deps) (stop func()) {
	defaults(&d)
	ctx, cancel := context.WithCancel(base)
	var once sync.Once
	overDeadline := 0
	var lastAlert time.Time
	go func() {
		defer once.Do(cancel)
		var startDrops uint64
		if d.StatsFn != nil {
			sts := d.StatsFn()
			startDrops = qdisc.SumDrops(sts)
		}
		var rtt, small []float64
		tick := time.NewTicker(500 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}
			rtt = append(rtt, d.Ping(ctx, d.Target, 3)...)
			if v, err := d.Small(ctx); err == nil {
				small = append(small, v)
			}
			if d.OnSnap == nil {
				continue
			}
			live := model.Event{Profile: "surveil"}
			if len(rtt) > 0 {
				s := metrics.Summarize(rtt)
				live.RTTp50Ms = round1(s.Median)
				live.RTTp95Ms = round1(s.P95)
			}
			if len(small) > 0 {
				sm := metrics.Summarize(small)
				live.Smallp95Ms = round1(sm.P95)
				live.DeadlineOKPct = round1(metrics.DeadlineOKPct(small, d.DeadlineMs))
				// sentinelle deadline — journal + pas de spam (1 alerte / 5 min)
				if d.DeadlineMs > 0 && sm.P95 > d.DeadlineMs {
					overDeadline++
					if overDeadline >= 10 && time.Since(lastAlert) > 5*time.Minute {
						lastAlert = time.Now()
						if OnWatchAlert != nil {
							OnWatchAlert(sm.P95, d.DeadlineMs)
						}
					}
				} else {
					overDeadline = 0
				}
			}
			if d.StatsFn != nil {
				sts := d.StatsFn()
				drp := int64(qdisc.SumDrops(sts)) - int64(startDrops)
				if drp > 0 {
					live.Drops = uint64(drp)
				}
			}
			model.ProfilesMu.RLock()
		p2 := model.Profiles["P2"]
		model.ProfilesMu.RUnlock()
		snap := d.Snapshot("surveil", "", live, nil, nil, 0, p2, nil)
			// watch ≠ campagne — Running doit rester false sinon il bat contre la
			// vérité de référence de pumpSnapshots et l'app traite chaque frame comme
			// une fin de campagne (bascule de panneaux en pleine surveillance).
			snap.Running = false
			d.OnSnap(snap)
			if len(rtt) > 3600 { // fenêtre de ~30 min à 3 pings / 500 ms
				rtt = rtt[len(rtt)-1800:]
			}
			if len(small) > 1200 {
				small = small[len(small)-600:]
			}
		}
	}()
	return func() { once.Do(cancel) }
}
