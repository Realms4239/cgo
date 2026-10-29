// Package campagne — orchestration d'une cellule de matrice
// (baseline → charge → récupération, SPEC §2.2). Durations injectable so
// tests rapides; la production utilise ces durées.
package campagne

import (
	"context"
	"fmt"
	"math"
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
	// BulkN — fenêtre de charge multi-flux : N connexions concurrentes vers
	// addr, avec le contrôle de congestion de la cellule. Nil ⇒ un seul flux.
	// Le CC est explicite : un multi-flux CUBIC ne remplace pas un BBR ;
	// l'ignorer rendrait les cellules BBR indiscernables. L'équité inter-flux
	// (JFI) ne se mesure que per-flow.
	BulkN func(ctx context.Context, addr, cc string, n int) ([]uint64, error)
	CPU   func() float64
	Now   func() time.Time
	// StatsFn rend les compteurs par qdisc pour les deltas pertes/octets.
	// nil ⇒ pas de tc -s (pertes à 0, goodput côté émetteur seul).
	StatsFn func() []qdisc.Stats
	// DeadlineMs — objectif small p95: le réglage de l'opérateur voyage
	// avec la campagne, le CSV exporté reflète la configuration.
	DeadlineMs float64
	// ProfileDelayMs — délai du profil courant : câblé par RunEvent pour
	// calibrer le timeout de la sonde small (VSAT 600 ms ⇒ 2,4 s, pas 2 s).
	ProfileDelayMs float64

	// Flows — nombre de flux de charge concurrents (défaut 4). >1 exige
	// BulkN câblé par l'hôte ; mesure l'isolation inter-flux (JFI) que
	// fq_codel/cake promettent et qu'un flux unique ne montre jamais.
	Flows int

	// PhysicsIfaces — ifaces de physique SUPPLÉMENTAIRES (banc durci :
	// c-wan en plus de w-edge) : le netem du profil s'y pose à l'identique,
	// sinon un run « symétrique » mesurerait un lien asymétrique en secret.
	PhysicsIfaces []string
	// ChannelModel — clean (défaut, perte uniforme historique) | bursty
	// (Gilbert-Elliott, P2/P3 seulement — validé avant toute mutation).
	ChannelModel string
	// Bench — étiquette de provenance : "legacy" (testbed.sh, reproduction
	// du corpus gelé) | "dualns" (banc durci, chemin de production).
	Bench string
	// DelayModel — "scalar" (défaut : delay/jitter déclarés) | "dist"
	// (table empirique par profil, chargée fail-closed AVANT toute mutation).
	DelayModel string
	// CalibrationDir — répertoire des <profil>.dist/.json.
	CalibrationDir string

	// Direction — sens de la charge : "" ou "up" = upload (client → sink,
	// shaper sur l'émission cliente, comportement historique) ; "down" =
	// download (source → client, shaper sur l'émission SERVEUR : le
	// shaper doit vivre là où le congestionnement naît).
	Direction string

	// ShapeMode — "egress" (défaut) ou "ingress" (IFB sur l'ingress WAN).
	// ingress exige bench=dualns + direction=down + un preflight non nil.
	ShapeMode string
	// IngressPreflight — lecture seule de ifb0/ingress/mirred, AVANT toute
	// mutation qdisc. nil en mode ingress ⇒ la cellule est refusée.
	IngressPreflight func() error

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
		ProfileCapMbps: prof.CapUp(), ProfileDelayMs: prof.DelayMs,
		ProfileJitterMs: prof.JitterMs, ProfileLossPct: prof.LossPct,
		Gates: gates, Running: true,
	}
	// Coût entreprise live : GOLD (modèle A, nécessite des octets transportés —
	// zéro honnête avant la charge) et APN 150 Ar/Go (modèle B, fenêtre de
	// charge). JFI à côté : l'équité inter-flux est un livrable du banc durci.
	charge := d.ChargeSec
	if charge == 0 {
		charge = model.ChargeSec
	}
	s.JFIPct = ev.JFI
	s.CostPrimaryARPerH = round1(metrics.CapacityWasteARPerHour(
		metrics.PlanGOLD.MonthlyOpExMGA, float64(ev.WastedBytes), float64(ev.Bytes)))
	s.CostB2BARPerH = round1(metrics.MeteredWasteARPerHour(
		float64(ev.WastedBytes), metrics.TariffCorporateAPN.ARPerGB, float64(charge)))
	return s
}

// httpDefault — repli pour hôtes sans SmallClient marqué (audit client-side :
// le petit objet y est best-effort, documenté docs/api.md)
func httpDefault() *http.Client { return &http.Client{Timeout: 2 * time.Second} }

// cpuGate — nil = CPU non mesuré : G7 reste non évaluée et ne peut pas passer.
// Une valeur nulle est le sentinelle historique "non instrumenté" ; 0 % réel
// n'est pas distinguable et ne doit jamais devenir un feu vert.
func cpuGate(cpu float64) *bool {
	if math.IsNaN(cpu) || cpu <= 0 {
		return nil
	}
	ok := cpu < 90
	return &ok
}

type qdiscRunner interface {
	Run(args ...string) ([]byte, error)
}

// Live est l'instantané mutable que le hub SSE diffuse.
type Live struct {
	mu   sync.Mutex
	snap Snapshot
}

type Snapshot struct {
	// Coût entreprise (live) : le scénario GOLD et le B2B 150 Ar/Go à côté
	// du cost_ar_per_h historique (qui garde sa signification détail gelée).
	CostPrimaryARPerH float64 `json:"cost_primary_ar_per_h"`
	CostB2BARPerH     float64 `json:"cost_b2b_ar_per_h"`
	JFIPct            float64 `json:"jfi_pct"`
	Phase             string  `json:"phase"`
	Profile           string  `json:"profile"`
	Qdisc             string  `json:"qdisc"`
	CC                string  `json:"cc"`
	Repetition        int     `json:"repetition"`
	EventID           int     `json:"event_id"`
	TotalEvents       int     `json:"total_events"`
	DoneEvents        int     `json:"done_events"`
	PhaseTotalS       int     `json:"phase_total_s"`
	LoadStatus        string  `json:"load_status"`
	RTTp50Ms          float64 `json:"rtt_p50_ms"`
	RTTp95Ms          float64 `json:"rtt_p95_ms"`
	Smallp95Ms        float64 `json:"small_p95_ms"`
	GoodputMbps       float64 `json:"bulk_goodput_mbps"`
	Drops             uint64  `json:"drops"`
	WastedBytes       uint64  `json:"wasted_bytes"`
	CostARPerH        float64 `json:"cost_ar_per_h"`
	DeadlineOKPct     float64 `json:"deadline_ok_pct"`
	ProfileCapMbps    float64 `json:"profile_capacity_mbps"`
	ProfileDelayMs    float64 `json:"profile_delay_ms"`
	ProfileJitterMs   float64 `json:"profile_jitter_ms"`
	ProfileLossPct    float64 `json:"profile_loss_pct"`
	Gates             []*bool `json:"gates"` // nil = not assessed
	Running           bool    `json:"running"`
	LastTS            int64   `json:"ts"`
}

func NewLive() *Live { return &Live{} }

func (l *Live) Set(s Snapshot) {
	l.mu.Lock()
	l.snap = s
	l.mu.Unlock()
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
			ss, _ := probe.Ping(ctx, t, n, 200)
			out := make([]float64, len(ss))
			for i, s := range ss {
				out[i] = s.RTTms
			}
			return out
		}
	}
	if d.Small == nil {
		d.Small = func(ctx context.Context) (float64, error) {
			// client marqué DSCP EF : la sonde vit la classe temps réel que
			// l'AQM doit protéger (diffserv exercée, pas supposée).
			// Timeout calibré sur le profil : le client partagé (2 s) tue la
			// sonde VSAT par construction (16 Ko à 600 ms > complétion 2 s,
			// G2 vide) — max(2 s, 4×delay) rend au satellite sa marge.
			to := 2 * time.Second
			if d.ProfileDelayMs > 0 {
				to = time.Duration(math.Max(2000, 4*d.ProfileDelayMs) * float64(time.Millisecond))
			}
			return probe.SmallObject(ctx, probe.SmallClientTimeout(to), d.SmallURL)
		}
	}
	if d.CPU == nil {
		d.CPU = func() float64 { return readCPUUsage() }
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
	if d.ChannelModel == "" {
		d.ChannelModel = "clean"
	}
	if d.Bench == "" {
		d.Bench = "legacy"
	}
	if d.DelayModel == "" {
		d.DelayModel = "scalar"
	}
	if d.ShapeMode == "" {
		d.ShapeMode = "egress"
	}
}

// validateShapeMode refuses an ingress cell whose topology or ownership is
// wrong, before any qdisc mutation. The bench owns ifb0 and the redirect; the
// campaign owns only the qdisc on the IFB.
func validateShapeMode(d *Deps) error {
	switch d.ShapeMode {
	case "egress":
		return nil
	case "ingress":
		if d.Bench != "dualns" {
			return fmt.Errorf("shape_mode ingress invalide sur bench %q — dualns uniquement", d.Bench)
		}
		if d.Direction != "down" {
			return fmt.Errorf("shape_mode ingress invalide en direction %q — down uniquement (le shaper ingress ne corrige pas l'upload)", d.Direction)
		}
		if d.TCShaper == nil || d.ShaperIf == "" {
			return fmt.Errorf("shape_mode ingress sans shaper/iface — ifb0 et son runner sont requis")
		}
		if d.IngressPreflight == nil {
			return fmt.Errorf("shape_mode ingress sans preflight — ifb0/ingress/mirred non vérifiés, cellule refusée")
		}
		return nil
	default:
		return fmt.Errorf("shape_mode %q invalide (egress|ingress)", d.ShapeMode)
	}
}

// RunEvent exécute une cellule de bout en bout et rend la ligne gelée.
func RunEvent(ctx context.Context, ev model.Event, prof model.Profile, d Deps) (model.Event, error) {
	d.ProfileDelayMs = prof.DelayMs // calibrage sonde small (VSAT) AVANT defaults
	defaults(&d)
	ensureQdiscModules()

	// Canal en salves : P2/P3 seulement, et l'erreur part AVANT toute
	// mutation qdisc — un repli silencieux en uniforme changerait l'identité
	// de la cellule mesurée (bursty annoncé, uniforme exécuté).
	switch d.ChannelModel {
	case "clean", "":
		d.ChannelModel = "clean"
	case "bursty":
		if _, ok := BurstGE(prof.ID); !ok {
			return ev, fmt.Errorf("channel_model bursty invalide pour %s — rafales sur P2/P3 uniquement (choisir clean)", prof.ID)
		}
	default:
		return ev, fmt.Errorf("channel_model %q invalide (clean|bursty)", d.ChannelModel)
	}
	ge := burstGEOrNull(prof.ID, d.ChannelModel)

	if err := validateShapeMode(&d); err != nil {
		return ev, err
	}
	if d.ShapeMode == "ingress" {
		if err := d.IngressPreflight(); err != nil {
			return ev, fmt.Errorf("ingress preflight: %w", err)
		}
	}

	// Physique : le saut de latence ET chaque iface supplémentaire (banc
	// durci : w-edge + c-wan, les deux sens voient délai et perte).
	if err := applyPhysics(&d, prof, ge); err != nil {
		return ev, fmt.Errorf("netem: %w", err)
	}
	shaper := qdiscRunner(d.TC)
	shaperIf := d.CliIf
	if d.TCShaper != nil {
		shaper = d.TCShaper
		shaperIf = d.ShaperIf
	}
	// reset du saut shaper AUSSI : en down, veth-s vit dans le netns et
	// n'est jamais nettoyé — un netem manuel oublié y survit (vu en prod :
	// RTT doublé 1,2 s, small 3,6 s > timeout, G2 vide "par construction",
	// 2 cellules cake invalidées par une topologie sale, pas par le lien).
	// del root emporte toute la hiérarchie (netem + enfants parent 1:).
	if shaperIf != d.CliIf {
		_, _ = shaper.Run("qdisc", "del", "dev", shaperIf, "root")
		// le saut serveur ne revoit JAMAIS les cellules up : nettoyer en
		// sortie, même en échec (defer) — sinon un cake-down survit et
		// étrangle les ACKs/réponses des cellules up suivantes (vu en prod :
		// RTT 807 ms et small 1843 ms sur P1-fibre pourtant à 20 ms).
		defer func() { _, _ = shaper.Run("qdisc", "del", "dev", shaperIf, "root") }()
	}
	// sens download : le shaper se pose sur l'émission SERVEUR — le
	// congestionnement du download naît côté source. ProdDeps câble
	// TCShaper sur NsRunner(cgo-srv)/veth-s pour ce cas.
	capDown := prof.CapDown()
	if d.ShapeMode == "ingress" {
		if err := qdisc.ApplyIngressShaper(shaper, shaperIf, ev.Qdisc, capDown, prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper ingress: %w", err)
		}
	} else if d.Direction == "down" {
		if err := qdisc.ApplyShaper(shaper, shaperIf, ev.Qdisc, capDown, prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper down: %w", err)
		}
	} else {
		if err := qdisc.ApplyShaper(shaper, shaperIf, ev.Qdisc, prof.CapUp(), prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper: %w", err)
		}
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
	// octets transportés cumulés (live) : la somme des deltas SURVIT à un
	// remplacement de qdisc (compteurs tc remis à zéro) — base honnête du
	// coût entreprise publié en cours de charge, avant le gel final.
	liveTransported := uint64(0)
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
			if db := int64(nowB) - int64(liveLastBytes); db > 0 {
				liveTransported += uint64(db)
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
		live.Bytes = liveTransported
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
	// G6 : le repos est-il assez calme pour que la charge qui suit soit
	// interprétable ? Le seuil tient compte du jitter DÉCLARÉ du profil : sur
	// un lien LEO, l'étalement du repos est la physique du lien, pas une
	// instabilité de l'hôte. Voir baseline.go — sur le corpus gelé, les 42
	// lignes P4 échouaient toutes G6 avec l'ancien seuil statique.
	g6ok, g6thr := BaselineStable(baseRTT, prof)
	set(model.G6BaselineStable, g6ok)
	_ = g6thr
	// latence au repos gélée avec la ligne : la note bufferbloat Waveform
	// (écart chargé − repos) exige la référence, pas seulement la charge
	ev.RTTBaseP50Ms = round1(sumB.Median)
	ev.RTTBaseP95Ms = round1(sumB.P95)

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
		if d.Direction == "down" {
			bulkFn = func(c context.Context, addr string) (uint64, error) {
				// la CC de la CELLULE pilote l'émetteur download (pas la CC
				// par-défaut de l'hôte) — sinon bbr/cubic sont indiscernables
				return probe.BulkDownloadTo(c, addr, string(ev.CC))
			}
		} else {
			bulkFn = func(c context.Context, addr string) (uint64, error) {
				return probe.BulkSendTo(c, addr, cc)
			}
		}
	}
	// charge — N flux concurrents si BulkN câblé (défaut 4, régime du banc
	// durci ; --flows 1 reproduit le gel historique), per-flow pour le JFI
	flows := d.Flows
	if flows <= 0 {
		flows = 4
	}
	type chargeOut struct {
		total uint64
		jfi   float64
	}
	done := make(chan chargeOut, 1)
	go func() {
		if d.BulkN != nil && flows > 1 {
			var per []uint64
			var err error
			if d.Direction == "down" {
				per, err = probe.BulkDownloadNTo(chgCtx, d.BulkAddr, string(ev.CC), flows)
			} else {
				per, err = d.BulkN(chgCtx, d.BulkAddr, string(ev.CC), flows)
			}
			_ = err
			var total uint64
			for _, b := range per {
				total += b
			}
			done <- chargeOut{total: total, jfi: metrics.JFI(per)}
			return
		}
		b, _ := bulkFn(chgCtx, d.BulkAddr)
		done <- chargeOut{total: b}
	}()
	if d.ChargeSec > 0 {
		time.Sleep(300 * time.Millisecond) // laisser le flood se connecter
	}
	// Échantillonnage CPU PENDANT la charge (G7) : une mesure après coup
	// verrait le calme revenu. Chaque échantillon /proc/stat couvre 200 ms.
	cpuSamples := make([]float64, 0, 64)
	cpuDone := make(chan struct{})
	go func() {
		defer close(cpuDone)
		for {
			select {
			case <-chgCtx.Done():
				return
			default:
			}
			v := d.CPU()
			if math.IsNaN(v) || v <= 0 {
				return
			}
			cpuSamples = append(cpuSamples, v)
			select {
			case <-chgCtx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}()
	chgRTT, chgSmall = collect(d.ChargeSec, model.PhaseCharge)
	cancel()
	<-cpuDone
	cpuAvg := math.NaN()
	if len(cpuSamples) > 0 {
		sum := 0.0
		for _, v := range cpuSamples {
			sum += v
		}
		cpuAvg = sum / float64(len(cpuSamples))
	}
	out := <-done
	bulkBytes = out.total
	ev.JFI = round1(out.jfi * 100) // 0..100 gelé avec la ligne ; 0 = mono-flux/n.a.
	_ = out
	set(model.G1BulkStarted, bulkBytes > 0)
	set(model.G2ProbesProducing, len(chgRTT) > 0 && len(chgSmall) > 0)
	chargeDur := float64(max(float64(d.ChargeSec), 1)) // dénominateur ≥ 1 s

	// goodput : préférer le delta côté récepteur de tc -s aux octets côté émetteur
	goodput := float64(bulkBytes) * 8 / 1e6 / chargeDur
	rxBytes := uint64(0)
	if d.StatsFn != nil {
		sts := d.StatsFn()
		endDrops := qdisc.SumDrops(sts)
		endBytes := qdisc.SumBytes(sts)
		// les compteurs tc repartent de zéro si le qdisc est remplacé en
		// cours de cellule — un delta négatif (wrap uint64) invaliderait la
		// cellule à tort (même garde int64 que publishLive)
		if dd := int64(endDrops) - int64(startDrops); dd > 0 {
			ev.Drops = uint64(dd)
		}
		if db := int64(endBytes) - int64(startBytes); db > 0 {
			rxBytes = uint64(db)
		}
		if rxBytes > 0 {
			goodput = float64(rxBytes) * 8 / 1e6 / chargeDur
		}
	}
	// octets transportés de la cellule (JSON seul) : récepteur si instrumenté,
	// sinon émetteur — base du ratio estimé drops×1448/octets du sidecar.
	ev.Bytes = bulkBytes
	if rxBytes > 0 {
		ev.Bytes = rxBytes
	}

	sumC := metrics.Summarize(chgRTT)
	ev.RTTp50Ms = round1(sumC.Median)
	ev.RTTp95Ms = round1(sumC.P95)
	// QDI — dégradation de délai de file : P95 chargé − P50 chargé (ms).
	// La métrique queue-delay de référence, gelée avec la ligne.
	ev.QDIPctMs = round1(sumC.P95 - sumC.Median)
	// VoIP R (E-model G.107 simplifié) — délai aller (RTT/2), gigue (p95−p50),
	// perte du profil : les conditions de charge vécues par un flux voix.
	ev.VoIPR = round1(metrics.VoIPR(sumC.Median/2, sumC.P95-sumC.Median, prof.LossPct))
	sm := metrics.Summarize(chgSmall)
	ev.Smallp95Ms = round1(sm.P95)
	ev.DeadlineOKPct = round1(metrics.DeadlineOKPct(chgSmall, d.DeadlineMs))
	ev.BulkGoodputMbps = round1(goodput)
	if math.IsNaN(cpuAvg) || cpuAvg <= 0 {
		ev.CPUPct = 0 // non mesuré : la colonne reste 0, G7 est nil
	} else {
		ev.CPUPct = round1(cpuAvg)
	}
	// octets gaspillés : segments retransmis × MSS (environ 1448 pour veth MTU 1500)
	ev.WastedBytes = ev.Drops * 1448
	ev.CostARPerH = round1(metrics.CostARPerH(ev.WastedBytes))
	set(model.G3LatencyPlausible, ev.RTTp95Ms < prof.DelayMs*10+200)
	// G4 juge le goodput du SENS MESURÉ : montant → CapUp, descendant →
	// CapDown (l'asymétrie est la règle, pas l'exception).
	// Plancher direction-aware, prouvé sur le banc (2026-09-04) : en
	// download mono-flux, le goodput est horlogé par les ACK (Mathis sur la
	// perte ACK 0,5 % : ~2,5 Mb/s BBR, ~0,5 cubic — IDENTIQUE tous qdiscs,
	// les ACK ne traversent pas le shaper download). Le plancher down (1 %
	// de CapDown) ne détecte que le tuyau mort (0, dial cassé, sink absent) ;
	// le verdict download se joue sur la LATENCE (small/deadline), pas le
	// débit. En montant, les données traversent le shaper : 50 % pleins.
	g4Cap := prof.CapUp()
	g4Floor := g4Cap * 0.5
	if d.Direction == "down" {
		g4Cap = prof.CapDown()
		g4Floor = g4Cap * 0.01
	}
	set(model.G4ThroughputCoherent, goodput >= g4Floor && goodput <= g4Cap*1.1+.5)
	gates[model.G7CPUNotSaturated] = cpuGate(cpuAvg)
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

	// annulation de CELLULE (skip) : rendre l'annulation en erreur — sans
	// elle la cellule se gélait complète malgré le skip (bug : le skip ne
	// coupait rien). L'arrêt global (Stop) n'annule pas la matrice entière
	// par cette voie : la boucle matrix vérifie ctx avant chaque cellule.
	if ctx.Err() != nil {
		return ev, ctx.Err()
	}

	status := model.GatePass
	// noms indexés par Gate (quarantaine lisible, spec §quarantaine).
	var gateNames = [8]string{
		"G0TargetReachable", "G1BulkStarted", "G2ProbesProducing",
		"G3LatencyPlausible", "G4ThroughputCoherent", "G5NoDuplicateRows",
		"G6BaselineStable", "G7CPUNotSaturated",
	}
	for g, b := range gates {
		if b == nil {
			continue
		}
		if !*b {
			ev.FailedGates = append(ev.FailedGates, gateNames[g])
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

func round1(v float64) float64 { return float64(int(v*10+0.5)) / 10 }

func loadFor(phase string) string {
	if phase == model.PhaseCharge {
		return "bulk-on"
	}
	return "idle"
}
