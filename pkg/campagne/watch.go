package campagne

import (
	"context"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// StartWatch runs the non-intrusive probe loop (ping + small object, no bulk —
// ARG.md: the audit is client-side by design). The wall stays alive outside a
// campagne, which makes the shape lever's effect visible in real time.
// Snapshots carry phase "surveil" and never touch the shaper.
func StartWatch(base context.Context, d Deps) (stop func()) {
	defaults(&d)
	ctx, cancel := context.WithCancel(base)
	var once sync.Once
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
				live.DeadlineOKPct = round1(metrics.DeadlineOKPct(small, 1000))
			}
			if d.StatsFn != nil {
				sts := d.StatsFn()
				drp := int64(qdisc.SumDrops(sts)) - int64(startDrops)
				if drp > 0 {
					live.Drops = uint64(drp)
				}
			}
			d.OnSnap(d.Snapshot("surveil", "", live, nil, nil, 0, model.Profiles["P2"], nil))
			if len(rtt) > 3600 { // ~30 min window at 3 pings / 500 ms
				rtt = rtt[len(rtt)-1800:]
			}
			if len(small) > 1200 {
				small = small[len(small)-600:]
			}
		}
	}()
	return func() { once.Do(cancel) }
}
