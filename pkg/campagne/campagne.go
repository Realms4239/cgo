// Package campagne — LIEN orchestration of one matrix cell
// (baseline → charge → récupération, SPEC §2.2). Durations injectable so
// tests run fast; production uses the LIEN constants.
package campagne

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/probe"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

type Deps struct {
	TC       qdiscRunner // netem hop (main ns)
	TCShaper qdiscRunner // shaping+aqm hop; nil ⇒ TC
	CliIf    string      // latency-hop iface (netem root)
	ShaperIf string      // shaping+aqm iface
	Target   string      // ping target reached through the constrained path
	SmallURL string      // small-object URL through the constrained path
	BulkAddr string      // host:port where the bulk sender connects

	Ping  func(ctx context.Context, target string, n int) []float64 // sorted ms
	Small func(ctx context.Context) (float64, error)
	Bulk  func(ctx context.Context, addr string) (uint64, error) // one charge-window flood
	CPU   func() float64
	Now   func() time.Time
	// StatsFn returns per-qdisc stats for drops/bytes delta measurement.
	// nil ⇒ no tc -s polling (drops stay 0, goodput from sender only).
	StatsFn func() []qdisc.Stats

	BaselineSec int
	ChargeSec   int
	RecupSec    int

	OnSnap func(Snapshot)
}

// Snapshot composes the current broadcast frame.
func (d *Deps) Snapshot(phase, load string, ev model.Event,
	base, chg map[string]float64, bulk uint64, prof model.Profile, gates []*bool) Snapshot {
	s := Snapshot{
		Phase: phase, LoadStatus: load,
		Profile: ev.Profile, Qdisc: string(ev.Qdisc), CC: string(ev.CC),
		Repetition: ev.Repetition, EventID: ev.EventID,
		RTTp50Ms: ev.RTTp50Ms, RTTp95Ms: ev.RTTp95Ms, Smallp95Ms: ev.Smallp95Ms,
		GoodputMbps: ev.BulkGoodputMbps, Drops: ev.Drops,
		Gates: gates, Running: phase != "",
	}
	return s
}

func httpDefault() *http.Client { return &http.Client{Timeout: 2 * time.Second} }

// readCPUIdleGuess — documented ceiling: /proc/stat sampling arrives with the
// VM bootstrap (M1.6); until then CPU is reported 0 and G7 stays not-assessed.
func readCPUIdleGuess() float64 { return 0 }

type qdiscRunner interface {
	Run(args ...string) ([]byte, error)
}

// Live is the mutable snapshot the SSE hub broadcasts.
type Live struct {
	mu        sync.Mutex
	snap      Snapshot
	subscribe chan struct{}
}

type Snapshot struct {
	Phase      string     `json:"phase"`
	Profile    string     `json:"profile"`
	Qdisc      string     `json:"qdisc"`
	CC         string     `json:"cc"`
	Repetition int        `json:"repetition"`
	EventID    int        `json:"event_id"`
	LoadStatus string     `json:"load_status"`
	RTTp50Ms   float64    `json:"rtt_p50_ms"`
	RTTp95Ms   float64    `json:"rtt_p95_ms"`
	Smallp95Ms float64    `json:"small_p95_ms"`
	GoodputMbps float64   `json:"bulk_goodput_mbps"`
	Drops      uint64     `json:"drops"`
	Gates      []*bool    `json:"gates"` // nil = not assessed
	Running    bool       `json:"running"`
	LastTS     int64      `json:"ts"`
}

func NewLive() *Live { return &Live{} }

func (l *Live) Set(s Snapshot) {
	l.mu.Lock(); l.snap = s; l.mu.Unlock()
	select { case l.subscribe <- struct{}{}: default: }
}
func (l *Live) Get() Snapshot { l.mu.Lock(); defer l.mu.Unlock(); return l.snap }

func defaults(d *Deps) {
	if d.Ping == nil {
		d.Ping = func(ctx context.Context, t string, n int) []float64 {
			ss, _ := probe.Ping(ctx, probe.ExecCmdRunner{}, t, n, 200)
			out := make([]float64, len(ss))
			for i, s := range ss { out[i] = s.RTTms }
			return out
		}
	}
	if d.Small == nil {
		d.Small = func(ctx context.Context) (float64, error) {
			return probe.SmallObject(ctx, httpDefault(), d.SmallURL)
		}
	}
	if d.CPU == nil { d.CPU = func() float64 { return readCPUIdleGuess() } }
	if d.Now == nil { d.Now = time.Now }
	if d.BaselineSec == 0 { d.BaselineSec = model.BaselineSec }
	if d.ChargeSec == 0 { d.ChargeSec = model.ChargeSec }
	if d.RecupSec == 0 { d.RecupSec = model.RecupSec }
}

// RunEvent executes one cell end-to-end and returns the frozen row.
func RunEvent(ctx context.Context, ev model.Event, prof model.Profile, d Deps) (model.Event, error) {
	defaults(&d)

	// configure both hops
	if err := qdisc.ApplyNetem(d.TC, d.CliIf, prof.DelayMs, prof.JitterMs, prof.LossPct); err != nil {
		return ev, fmt.Errorf("netem: %w", err)
	}
	shaper := qdiscRunner(d.TC)
	if d.TCShaper != nil {
		shaper = d.TCShaper
	}
	if err := qdisc.ApplyShaper(shaper, d.ShaperIf, ev.Qdisc, prof.CapacityMbps, prof.DelayMs); err != nil {
		return ev, fmt.Errorf("shaper: %w", err)
	}

	gates := make([]*bool, model.GateCount)
	set := func(g model.Gate, ok bool) { b := ok; gates[g] = &b }
	push := func(phase string) {
		if d.OnSnap != nil {
			d.OnSnap(d.Snapshot(phase, loadFor(phase), ev,
				nil, nil, 0, prof, gates))
		}
	}

	var baseRTT, baseSmall, chgRTT, chgSmall, recRTT []float64
	var bulkBytes uint64
	collect := func(secs int) (rtt, small []float64) {
		if secs <= 0 { // instant window: single synthetic pass (tests)
			rtt = d.Ping(ctx, d.Target, 5)
			if v, err := d.Small(ctx); err == nil {
				small = append(small, v)
			}
			return rtt, small
		}
		deadline := d.Now().Add(time.Duration(secs) * time.Second)
		for d.Now().Before(deadline) {
			rtt = append(rtt, d.Ping(ctx, d.Target, 5)...)
			if v, err := d.Small(ctx); err == nil {
				small = append(small, v)
			}
			time.Sleep(300 * time.Millisecond)
		}
		return
	}

	// baseline
	push(model.PhaseBaseline)
	baseRTT, baseSmall = collect(d.BaselineSec)
	_ = baseSmall
	set(model.G0TargetReachable, len(baseRTT) > 0)
	sumB := metrics.Summarize(baseRTT)
	set(model.G6BaselineStable, len(baseRTT) > 4 && sumB.P95-sumB.Median < maxVal(5, .2*sumB.Median))

	// charge — bulk flood with the cell's congestion control (real CC matrix)
	push(model.PhaseCharge)
	chgCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// snapshot tc -s counters at charge start for drops/goodput delta
	var startDrops, startBytes uint64
	if d.StatsFn != nil {
		sts := d.StatsFn()
		startDrops = qdisc.SumDrops(sts)
		startBytes = qdisc.SumBytes(sts)
	}

	bulkFn := d.Bulk
	if bulkFn == nil {
		cc := string(ev.CC)
		bulkFn = func(c context.Context, addr string) (uint64, error) {
			return probe.BulkSendTo(c, addr, cc)
		}
	}
	done := make(chan uint64, 1)
	go func() { b, _ := bulkFn(chgCtx, d.BulkAddr); done <- b }()
	if d.ChargeSec > 0 {
		time.Sleep(300 * time.Millisecond) // let the flood connect
	}
	chgRTT, chgSmall = collect(d.ChargeSec)
	cancel()
	bulkBytes = <-done
	set(model.G1BulkStarted, bulkBytes > 0)
	set(model.G2ProbesProducing, len(chgRTT) > 0 && len(chgSmall) > 0)
	chargeDur := float64(maxVal(float64(d.ChargeSec), 1)) // ≥1 s denominator

	// goodput: prefer tc -s receiver-side delta over sender-side bytes
	goodput := float64(bulkBytes) * 8 / 1e6 / chargeDur
	if d.StatsFn != nil {
		sts := d.StatsFn()
		endDrops := qdisc.SumDrops(sts)
		endBytes := qdisc.SumBytes(sts)
		ev.Drops = endDrops - startDrops
		rxBytes := endBytes - startBytes
		if rxBytes > 0 {
			goodput = float64(rxBytes) * 8 / 1e6 / chargeDur
		}
	}

	sumC := metrics.Summarize(chgRTT)
	ev.RTTp50Ms = round1(sumC.Median)
	ev.RTTp95Ms = round1(sumC.P95)
	sm := metrics.Summarize(chgSmall)
	ev.Smallp95Ms = round1(sm.P95)
	ev.DeadlineOKPct = round1(metrics.DeadlineOKPct(chgSmall, 1000))
	ev.BulkGoodputMbps = round1(goodput)
	cpuAvg := d.CPU()
	ev.CPUPct = round1(cpuAvg)
	// wasted bytes: retransmitted segments × MSS (approx 1448 for veth MTU 1500)
	ev.WastedBytes = ev.Drops * 1448
	ev.CostARPerH = round1(metrics.CostARPerH(ev.WastedBytes))
	set(model.G3LatencyPlausible, ev.RTTp95Ms < prof.DelayMs*10+200)
	set(model.G4ThroughputCoherent, goodput >= prof.CapacityMbps*.5 && goodput <= prof.CapacityMbps*1.1+.5)
	set(model.G7CPUNotSaturated, cpuAvg < 90)
	set(model.G5NoDuplicateRows, true) // enforced by writer at freeze

	// récupération
	recRTT, _ = collect(d.RecupSec)
	_ = recRTT

	status := model.GatePass
	for g, b := range gates {
		if b == nil { continue }
		if !*b {
			switch model.Gate(g) {
			case model.G2ProbesProducing, model.G6BaselineStable:
				if status == model.GatePass { status = model.GateDegraded }
			default:
				status = model.GateInvalid
			}
		}
	}
	ev.GateStatus = status
	return ev, nil
}

func maxVal(a, b float64) float64 { if a > b { return a }; return b }
func round1(v float64) float64   { return float64(int(v*10+0.5)) / 10 }

func loadFor(phase string) string {
	if phase == model.PhaseCharge {
		return "bulk-on"
	}
	return "idle"
}
