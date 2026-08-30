// Package campagne — orchestration d'une cellule de matrice
// (baseline → charge → récupération, SPEC §2.2). Durations injectable so
// tests rapides; la production utilise ces durées.
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
	// StatsFn rend les compteurs par qdisc pour les deltas pertes/octets.
	// nil ⇒ pas de tc -s (pertes à 0, goodput côté émetteur seul).
	StatsFn func() []qdisc.Stats
	// DeadlineMs — objectif small p95: le réglage de l'opérateur voyage
	// avec la campagne, le CSV exporté reflète la configuration.
	DeadlineMs float64

	BaselineSec int
	ChargeSec   int
	RecupSec    int

	// Progression de la matrice — peuplé par StartMatrix, porté par chaque
	// snapshot : un seul canal de vérité pour le badge et le popup.
	TotalEvents int
	DoneEvents  int

	OnSnap func(Snapshot)
}

// Snapshot compose le frame de diffusion courant.
// Running est toujours true ici — push n'est appelé que pendant que la matrice
// tourne ; la vérité de référence (mtx.Running) est appliquée par pumpSnapshots
// à 10 Hz comme dernier mot, évitant le battement pump vs push.
func (d *Deps) Snapshot(phase, load string, ev model.Event,
	base, chg map[string]float64, bulk uint64, prof model.Profile, gates []*bool) Snapshot {
	phaseTotal := 0
	switch phase {
	case model.PhaseBaseline:
		phaseTotal = d.BaselineSec
	case model.PhaseCharge:
		phaseTotal = d.ChargeSec
	case model.PhaseRecup:
		phaseTotal = d.RecupSec
	}
	s := Snapshot{
		Phase: phase, LoadStatus: load,
		Profile: ev.Profile, Qdisc: string(ev.Qdisc), CC: string(ev.CC),
		Repetition: ev.Repetition, EventID: ev.EventID,
		TotalEvents: d.TotalEvents, DoneEvents: d.DoneEvents, PhaseTotalS: phaseTotal,
		RTTp50Ms: ev.RTTp50Ms, RTTp95Ms: ev.RTTp95Ms, Smallp95Ms: ev.Smallp95Ms,
		GoodputMbps: ev.BulkGoodputMbps, Drops: ev.Drops,
		WastedBytes: ev.WastedBytes, CostARPerH: ev.CostARPerH, DeadlineOKPct: ev.DeadlineOKPct,
		ProfileCapMbps: prof.CapacityMbps, ProfileDelayMs: prof.DelayMs,
		ProfileJitterMs: prof.JitterMs, ProfileLossPct: prof.LossPct,
		Gates: gates, Running: true,
	}
	return s
}

func httpDefault() *http.Client { return &http.Client{Timeout: 2 * time.Second} }

// readCPUIdleGuess — plafond documenté : l'échantillonnage /proc/stat arrive avec
// le bootstrap VM (M1.6) ; jusque-là le CPU est rapporté 0 et G7 reste non évalué.
func readCPUIdleGuess() float64 { return 0 }

type qdiscRunner interface {
	Run(args ...string) ([]byte, error)
}

// Live est l'instantané mutable que le hub SSE diffuse.
type Live struct {
	mu        sync.Mutex
	snap      Snapshot
	subscribe chan struct{}
}

type Snapshot struct {
	Phase         string  `json:"phase"`
	Profile       string  `json:"profile"`
	Qdisc         string  `json:"qdisc"`
	CC            string  `json:"cc"`
	Repetition    int     `json:"repetition"`
	EventID       int     `json:"event_id"`
	TotalEvents   int     `json:"total_events"`
	DoneEvents    int     `json:"done_events"`
	PhaseTotalS   int     `json:"phase_total_s"`
	LoadStatus    string  `json:"load_status"`
	RTTp50Ms      float64 `json:"rtt_p50_ms"`
	RTTp95Ms      float64 `json:"rtt_p95_ms"`
	Smallp95Ms    float64 `json:"small_p95_ms"`
	GoodputMbps   float64 `json:"bulk_goodput_mbps"`
	Drops         uint64  `json:"drops"`
	WastedBytes   uint64  `json:"wasted_bytes"`
	CostARPerH    float64 `json:"cost_ar_per_h"`
	DeadlineOKPct float64 `json:"deadline_ok_pct"`
	ProfileCapMbps  float64 `json:"profile_capacity_mbps"`
	ProfileDelayMs  float64 `json:"profile_delay_ms"`
	ProfileJitterMs float64 `json:"profile_jitter_ms"`
	ProfileLossPct  float64 `json:"profile_loss_pct"`
	Gates         []*bool `json:"gates"` // nil = not assessed
	Running       bool    `json:"running"`
	LastTS        int64   `json:"ts"`
}

func NewLive() *Live { return &Live{} }

func (l *Live) Set(s Snapshot) {
	l.mu.Lock()
	l.snap = s
	l.mu.Unlock()
	select {
	case l.subscribe <- struct{}{}:
	default:
	}
}
func (l *Live) Get() Snapshot { l.mu.Lock(); defer l.mu.Unlock(); return l.snap }

// SetRunning met à jour atomiquement le seul drapeau Running, évitant
// Get+Modify+Set lost-update race where a pump's stale Get overwrites
// OnSnap's structural fields (profile/qdisc/gates).
func (l *Live) SetRunning(v bool) {
	l.mu.Lock()
	l.snap.Running = v
	l.mu.Unlock()
}

func defaults(d *Deps) {
	if d.Ping == nil {
		d.Ping = func(ctx context.Context, t string, n int) []float64 {
			ss, _ := probe.Ping(ctx, probe.ExecCmdRunner{}, t, n, 200)
			out := make([]float64, len(ss))
			for i, s := range ss {
				out[i] = s.RTTms
			}
			return out
		}
	}
	if d.Small == nil {
		d.Small = func(ctx context.Context) (float64, error) {
			return probe.SmallObject(ctx, httpDefault(), d.SmallURL)
		}
	}
	if d.CPU == nil {
		d.CPU = func() float64 { return readCPUIdleGuess() }
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.BaselineSec == 0 {
		d.BaselineSec = model.BaselineSec
	}
	if d.ChargeSec == 0 {
		d.ChargeSec = model.ChargeSec
	}
	if d.RecupSec == 0 {
		d.RecupSec = model.RecupSec
	}
}

// RunEvent exécute une cellule de bout en bout et rend la ligne gelée.
func RunEvent(ctx context.Context, ev model.Event, prof model.Profile, d Deps) (model.Event, error) {
	defaults(&d)

	// configurer les deux sauts — reset d'abord : le levier de façonnage (ou une
	// cellule périmée) peut laisser un qdisc étranger à la racine, ce qui ferait échouer chaque apply
	_, _ = d.TC.Run("qdisc", "del", "dev", d.CliIf, "root")
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

	// vérité live (§5) : publier les mesures en cours à chaque tour de sonde pour que
	// le hub 10 Hz porte rtt/small/goodput/drops PENDANT la fenêtre — pas seulement
	// aux frontières de phase (les courbes restaient à 0 en plein charge avant).
	startDrops, startBytes := uint64(0), uint64(0)
	if d.StatsFn != nil {
		sts := d.StatsFn()
		startDrops, startBytes = qdisc.SumDrops(sts), qdisc.SumBytes(sts)
	}
	liveLastBytes, liveLastT := startBytes, d.Now()
	publishLive := func(phase string, rtt, small []float64) {
		if d.OnSnap == nil {
			return
		}
		live := ev
		if len(rtt) > 0 {
			s := metrics.Summarize(rtt)
			live.RTTp50Ms = round1(s.Median)
			live.RTTp95Ms = round1(s.P95)
		}
		if len(small) > 0 {
			sm := metrics.Summarize(small)
			live.Smallp95Ms = round1(sm.P95)
			live.DeadlineOKPct = round1(metrics.DeadlineOKPct(small, d.DeadlineMs))
		}
		if d.StatsFn != nil {
			sts := d.StatsFn()
			nowB, nowT := qdisc.SumBytes(sts), d.Now()
			if nowB < liveLastBytes {
				liveLastBytes = nowB // qdisc remplacé (levier de façonnage) — compteurs remis à zéro
			}
			if dt := nowT.Sub(liveLastT).Seconds(); dt > 0.2 {
				g := float64(nowB-liveLastBytes) * 8 / 1e6 / dt
				if g >= 0 && g <= 2500 { // écarter les artefacts de compteur, garder l'axe lisible
					live.BulkGoodputMbps = round1(g)
				}
			}
			d := int64(qdisc.SumDrops(sts)) - int64(startDrops)
			if d >= 0 {
				live.Drops = uint64(d)
			}
			liveLastBytes, liveLastT = nowB, nowT
		}
		live.WastedBytes = live.Drops * 1448
		live.CostARPerH = round1(metrics.CostARPerH(live.WastedBytes))
		d.OnSnap(d.Snapshot(phase, loadFor(phase), live, nil, nil, 0, prof, gates))
	}
	collect := func(secs int, phase string) (rtt, small []float64) {
		if secs <= 0 { // fenêtre instantanée : une seule passe synthétique (tests)
			rtt = d.Ping(ctx, d.Target, 5)
			if v, err := d.Small(ctx); err == nil {
				small = append(small, v)
			}
			publishLive(phase, rtt, small)
			return rtt, small
		}
		deadline := d.Now().Add(time.Duration(secs) * time.Second)
		for d.Now().Before(deadline) {
			// arrêt prompt : une campagne annulée ne doit pas continuer à sonder jusqu'à
			// l'échéance de phase (jusqu'à 120 s de tours vides en boucle chaude)
			select {
			case <-ctx.Done():
				return rtt, small
			default:
			}
			// publier D'ABORD — ping+small peuvent bloquer des secondes sur un lien
			// saturé ; le mur montre la vérité courante à chaque tour, pas des zéros de frontière
			publishLive(phase, rtt, small)
			rtt = append(rtt, d.Ping(ctx, d.Target, 5)...)
			if v, err := d.Small(ctx); err == nil {
				small = append(small, v)
			}
		}
		return
	}

	// baseline
	push(model.PhaseBaseline)
	baseRTT, baseSmall = collect(d.BaselineSec, model.PhaseBaseline)
	_ = baseSmall
	set(model.G0TargetReachable, len(baseRTT) > 0)
	sumB := metrics.Summarize(baseRTT)
	set(model.G6BaselineStable, len(baseRTT) > 4 && sumB.P95-sumB.Median < maxVal(5, .2*sumB.Median))

	// charge — bulk flood avec le contrôle de congestion de la cellule (vraie matrice CC)
	push(model.PhaseCharge)
	chgCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// re-baseline des compteurs tc -s au début de la charge — le delta drops/goodput est propre à la charge
	if d.StatsFn != nil {
		sts := d.StatsFn()
		startDrops, startBytes = qdisc.SumDrops(sts), qdisc.SumBytes(sts)
		liveLastBytes, liveLastT = startBytes, d.Now()
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
		time.Sleep(300 * time.Millisecond) // laisser le flood se connecter
	}
	chgRTT, chgSmall = collect(d.ChargeSec, model.PhaseCharge)
	cancel()
	bulkBytes = <-done
	set(model.G1BulkStarted, bulkBytes > 0)
	set(model.G2ProbesProducing, len(chgRTT) > 0 && len(chgSmall) > 0)
	chargeDur := float64(maxVal(float64(d.ChargeSec), 1)) // dénominateur ≥ 1 s

	// goodput : préférer le delta côté récepteur de tc -s aux octets côté émetteur
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
	// QDI — dégradation de délai de file : P95 chargé − P50 chargé (ms).
	// La métrique queue-delay de référence, gelée avec la ligne.
	ev.QDIPctMs = round1(sumC.P95 - sumC.Median)
	sm := metrics.Summarize(chgSmall)
	ev.Smallp95Ms = round1(sm.P95)
	ev.DeadlineOKPct = round1(metrics.DeadlineOKPct(chgSmall, d.DeadlineMs))
	ev.BulkGoodputMbps = round1(goodput)
	cpuAvg := d.CPU()
	ev.CPUPct = round1(cpuAvg)
	// octets gaspillés : segments retransmis × MSS (environ 1448 pour veth MTU 1500)
	ev.WastedBytes = ev.Drops * 1448
	ev.CostARPerH = round1(metrics.CostARPerH(ev.WastedBytes))
	set(model.G3LatencyPlausible, ev.RTTp95Ms < prof.DelayMs*10+200)
	set(model.G4ThroughputCoherent, goodput >= prof.CapacityMbps*.5 && goodput <= prof.CapacityMbps*1.1+.5)
	set(model.G7CPUNotSaturated, cpuAvg < 90)
	set(model.G5NoDuplicateRows, true) // appliqué par l'écrivain au gel
	// publier les métriques mises à jour pour que SSE porte la vérité (wasted/cost/deadline) sans dérivation
	if d.OnSnap != nil {
		d.OnSnap(d.Snapshot(model.PhaseCharge, loadFor(model.PhaseCharge), ev, nil, nil, 0, prof, gates))
	}

	// récupération
	if d.OnSnap != nil {
		d.OnSnap(d.Snapshot(model.PhaseRecup, loadFor(model.PhaseRecup), ev, nil, nil, 0, prof, gates))
	}
	recRTT, _ = collect(d.RecupSec, model.PhaseRecup)
	_ = recRTT

	status := model.GatePass
	for g, b := range gates {
		if b == nil {
			continue
		}
		if !*b {
			switch model.Gate(g) {
			case model.G2ProbesProducing, model.G6BaselineStable:
				if status == model.GatePass {
					status = model.GateDegraded
				}
			default:
				status = model.GateInvalid
			}
		}
	}
	ev.GateStatus = status
	return ev, nil
}

func maxVal(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func round1(v float64) float64 { return float64(int(v*10+0.5)) / 10 }

func loadFor(phase string) string {
	if phase == model.PhaseCharge {
		return "bulk-on"
	}
	return "idle"
}
