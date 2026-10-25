package audit

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/probe"
)

// Paramètres d'un audit de lien.
type Params struct {
	AuditID  string
	Site     string
	LinkType string // fiber, 5g, 4g, vsat, other
	Provider string
	Duration int // seconds, e.g. 60–300

	Target   string // ping target
	SmallURL string // small object url
	BulkAddr string // optional bulk addr for throughput
}

// Ligne de résultat d'audit.
type Result struct {
	AuditID        string  `json:"audit_id"`
	Timestamp      string  `json:"timestamp"`
	Site           string  `json:"site"`
	LinkType       string  `json:"link_type"`
	Provider       string  `json:"provider"`
	RTTIdleP50     float64 `json:"rtt_idle_p50_ms"`
	RTTIdleP95     float64 `json:"rtt_idle_p95_ms"`
	RTTLoadedP50   float64 `json:"rtt_loaded_p50_ms"`
	RTTLoadedP95   float64 `json:"rtt_loaded_p95_ms"`
	ThroughputMbps float64 `json:"throughput_mbps"`
	LossPct        float64 `json:"loss_pct"`
	HTTPSmallP95   float64 `json:"http_small_p95_ms"`
	DataUsedMB     float64 `json:"data_used_mb"`
	Notes          string  `json:"notes"`

	// Note bufferbloat (méthode Waveform, bandes verbatim citées) :
	// écart p95 idle → chargée. L'audit charge en upload ; le sens
	// descendant reste non mesuré jusqu'à la campagne download.
	BloatDeltaMs float64 `json:"bloat_delta_ms"`
	BloatGrade   string  `json:"bloat_grade"`
	BloatVerdict string  `json:"bloat_verdict"`

	// Rapprochement au référentiel P1–P4 (calibration par site) : calculé
	// en fin d'audit, gelé avec la ligne comme tout le reste.
	ProfileMatch string `json:"profile_match"`
	MatchDelta   string `json:"match_delta"`

	// DrainMs — temps de retour au repos après arrêt de la charge, en ms
	// au-dessus du seuil de repos. 0 = non mesuré (pas de phase 3), jamais
	// « drain instantané ». Mesure la vitesse à laquelle la file se vide.
	DrainMs       float64 `json:"drain_ms"`
	DrainMeasured bool    `json:"drain_measured"`

	// LoadSeconds is the measured duration of the bulk transfer, in seconds.
	// The financial projection extrapolates from this rather than from a
	// hardcoded window, so the hourly figure is traceable to a real interval.
	LoadSeconds float64 `json:"load_seconds"`
	// Phase windows actually executed. 60 s = 10 s baseline, 40 s load,
	// 10 s drain; short audits scale down. Stored so the report can prove
	// the grade came from a loaded window, not from an idle sample.
	BaselineSeconds float64 `json:"baseline_seconds"`
	RecoverySeconds float64 `json:"recovery_seconds"`
}

type Deps struct {
	Ping  func(ctx context.Context, target string, n int) []float64
	Small func(ctx context.Context) (float64, error)
	Bulk  func(ctx context.Context, addr string) (uint64, error)
}

// bulkResult — octets réellement émis, erreur et durée de la charge.
// La durée sert au débit : diviser par une constante 10 s inventait un
// débit faux quand la fenêtre mesurée durait 8 s.
type bulkResult struct {
	bytes   uint64
	err     error
	elapsed time.Duration
}

// Run exécute un audit de Duration secondes, en collectant des échantillons ping/small.
// Découpe B : 0–12 s idle, 12–22 s bulk flood, 22–30 s chargé distinct selon Q26, plus repli iperf3 par BulkAddr.
// Si Duration <30, fenêtre unique avec idle==loaded copie honnête ; si >=30, trois fenêtres produisent des RTT idle vs chargé distincts.
func Run(ctx context.Context, p Params, d Deps) (*Result, error) {
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
		smallURL := p.SmallURL
		// Bounded client: the old http.DefaultClient had no timeout, so a
		// stalled 16 KB fetch could pin the audit goroutine after every phase
		// had finished. 8 s covers a 600 ms VSAT path with retransmissions
		// without letting a dead peer hang the process.
		smallClient := probe.SmallClientTimeout(8 * time.Second)
		d.Small = func(ctx context.Context) (float64, error) {
			if smallURL != "" {
				return probe.SmallObject(ctx, smallClient, smallURL)
			}
			// pas d'URL : petit objet NON mesuré (0 + note), jamais de
			// constante 25 ms synthétique dans le CSV gelé
			return 0, fmt.Errorf("no small URL")
		}
	}
	// sans Bulk câblé, un BulkAddr suffit à activer la charge réelle :
	// c'est le cas nominal du CLI et de l'API, qui n'ont pas à connaître
	// l'implémentation du puits. BBR = contrôle de congestion de la
	// campagne de référence (RQ1) ; un flux cubic sur 100 ms de RTT ne
	// remplit pas la capacité et ne crée aucune file — l'audit mesurerait
	// un lien vide puis noterait A+ à tort.
	if d.Bulk == nil && p.BulkAddr != "" {
		d.Bulk = func(ctx context.Context, addr string) (uint64, error) {
			return probe.BulkSendTo(ctx, addr, "bbr")
		}
	}
	start := time.Now()
	if p.Duration <= 0 {
		p.Duration = 10
	}
	// Timeline honnête : la charge bulk couvre la fenêtre « chargée », qui
	// occupe la FIN de l'audit. L'ancienne version lançait le bulk à t=12 s
	// puis mesurait à t=Duration-8 s : à 60 s, la charge était finie depuis
	// 30 s et le p95 « chargé » était un p95 idle déguisé.
	//
	// 60 s audit = 10 s quiescent + 40 s saturating + 10 s drain. Short
	// audits keep a usable baseline/load split without stealing the drain
	// window they cannot afford.
	withLoad := d.Bulk != nil && p.Duration >= 2
	idleSecs, loadedSecs, recSecs := phasePlan(p.Duration, withLoad)

	// One audit-wide deadline: a caller may pass context.Background() (the
	// HTTP handler does), and a stuck peer must not leave a goroutine and a
	// half-written CSV behind forever.
	auditCtx, auditCancel := context.WithTimeout(ctx, time.Duration(p.Duration+30)*time.Second)
	defer auditCancel()
	ctx = auditCtx

	var idleRTTs, loadedRTTs, idleSmalls, loadedSmalls, recRTTs, recSmalls []float64
	var bulkBytes uint64
	var bulkErr error
	var bulkElapsed time.Duration
	var notes []string
	// tentatives ping : le dénominateur de la perte. L'ancien calcul
	// (durée × 12,5/s) ignorait la durée réelle d'un appel ping (~1 s pour
	// 5 paquets à 200 ms) et annonçait ~70 % de perte sur un lien parfait.
	pingCalls := 0
	const pingPerCall = 5

	collectWindow := func(secs int, dstRTT, dstSmall *[]float64) {
		deadline := time.Now().Add(time.Duration(secs) * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			pingCalls++
			*dstRTT = append(*dstRTT, d.Ping(ctx, p.Target, pingPerCall)...)
			if v, err := d.Small(ctx); err == nil {
				*dstSmall = append(*dstSmall, v)
			}
			time.Sleep(400 * time.Millisecond)
		}
	}

	if withLoad {
		bulkStart := make(chan struct{})
		bulkDone := make(chan bulkResult, 1)
		bulkCtx, bulkCancel := context.WithCancel(ctx)
		defer bulkCancel()
		go func() {
			select {
			case <-bulkStart:
			case <-ctx.Done():
				bulkDone <- bulkResult{}
				return
			}
			t0 := time.Now()
			b, err := d.Bulk(bulkCtx, p.BulkAddr)
			bulkDone <- bulkResult{bytes: b, err: err, elapsed: time.Since(t0)}
		}()

		collectWindow(idleSecs, &idleRTTs, &idleSmalls)
		close(bulkStart)
		// laisser le dial/flood démarrer avant de compter les pings chargés
		time.Sleep(300 * time.Millisecond)
		collectWindow(loadedSecs, &loadedRTTs, &loadedSmalls)
		bulkCancel()
		select {
		case br := <-bulkDone:
			bulkBytes, bulkErr, bulkElapsed = br.bytes, br.err, br.elapsed
		case <-time.After(5 * time.Second):
			notes = append(notes, "bulk : pas de retour après annulation")
		}
		// Phase 3 — drain. Après l'arrêt de la charge, on continue de sonder :
		// le temps que le p95 revienne sous le seuil de repos est la vitesse à
		// laquelle la file se vide. C'est la mesure qui distingue un AQM qui
		// tient la file d'un AQM qui l'a seulement vidée à temps.
		if recSecs > 0 {
			collectWindow(recSecs, &recRTTs, &recSmalls)
		}
	} else {
		// audit court / sans bulk : fenêtre unique, idle==loaded honnête
		collectWindow(p.Duration, &idleRTTs, &idleSmalls)
		loadedRTTs = append([]float64(nil), idleRTTs...)
		loadedSmalls = append([]float64(nil), idleSmalls...)
	}

	idleSummary := metrics.Summarize(idleRTTs)
	loadedSummary := metrics.Summarize(loadedRTTs)
	// garde anti-ligne-fantôme : zéro sonde = cible injoignable, on refuse
	// de geler des zéros (un audit aveugle qui passerait pour une mesure
	// serait pire que pas d'audit — même discipline que les portes G0–G7).
	if len(idleRTTs) == 0 {
		return nil, fmt.Errorf("audit %s : aucune sonde ping — cible %s injoignable ?", p.AuditID, p.Target)
	}
	// La note porte sur l'écart p95 (queue de latence sous charge), pas sur
	// la médiane : le mémoire cite p95, et c'est la queue qui dégrade la
	// voix/le jeu. Sans charge réellement mesurée, aucune note : un écart
	// 0 ms n'est pas un A+, c'est une absence de mesure.
	bloatDelta := 0.0
	loadMeasured := withLoad && bulkBytes > 0 && len(loadedRTTs) > 0
	if loadMeasured {
		bloatDelta = loadedSummary.P95 - idleSummary.P95
	}
	allSmalls := append(append([]float64(nil), idleSmalls...), loadedSmalls...)
	sSummary := metrics.Summarize(allSmalls)
	// estimation de perte : échantillons reçus vs TENTÉS (5 par appel ping).
	// Compter les appels, pas la durée — un ping lent n'est pas de la perte.
	expectedSamples := float64(pingCalls * pingPerCall)
	actualSamples := float64(len(idleRTTs) + len(loadedRTTs) + len(recRTTs))
	if !withLoad {
		// fenêtre unique : loaded = copie honnête d'idle — compter une fois
		actualSamples = float64(len(idleRTTs))
	}
	lossPct := lossPercent(expectedSamples, actualSamples)
	throughput := 0.0
	dataUsed := 0.0
	if bulkBytes > 0 {
		secs := bulkElapsed.Seconds()
		if secs <= 0 {
			secs = float64(loadedSecs)
		}
		throughput = float64(bulkBytes) * 8 / 1e6 / secs
		dataUsed = float64(bulkBytes) / 1e6
	} else if withLoad {
		msg := "charge bulk non mesurée — sink injoignable ou erreur"
		if bulkErr != nil {
			msg += " (" + bulkErr.Error() + ")"
		}
		notes = append(notes, msg)
	} else {
		notes = append(notes, "throughput non mesuré sans bulk sink (BulkAddr) — iperf3 fallback disponible si installé")
	}
	grade, verdict := "", ""
	if loadMeasured {
		grade = metrics.BufferbloatGrade(bloatDelta)
		verdict = metrics.BufferbloatVerdict(grade)
	} else {
		notes = append(notes, "bufferbloat non mesuré — aucune charge bulk exploitable")
		verdict = "Bufferbloat non mesuré — aucune charge bulk exploitable."
	}

	// Drain : combien de temps la latence met à revenir sous le seuil de repos
	// après l'arrêt de la charge. Mesuré seulement si la fenêtre de récupération
	// a produit des échantillons ; sinon 0 = non mesuré, jamais « drain instantané ».
	drainMs := 0.0
	drainMeasured := false
	if loadMeasured && len(recRTTs) > 0 {
		drainMeasured = true
		recSummary := metrics.Summarize(recRTTs)
		// Le seuil de retour au repos : le p95 de repos plus la tolérance G6 de
		// l'audit (5 ms ou 20 % de la médiane, le plus grand des deux).
		restThreshold := idleSummary.P95 + math.Max(5, 0.2*idleSummary.Median)
		drainMs = recSummary.P95 - restThreshold
		if drainMs < 0 {
			drainMs = 0 // revenu sous le seuil : drain effectivement terminé
		}
	}
	if len(allSmalls) == 0 {
		notes = append(notes, "petit objet non mesuré (SmallURL absent)")
	}

	res := &Result{
		AuditID:        p.AuditID,
		Timestamp:      start.Format(time.RFC3339),
		Site:           p.Site,
		LinkType:       p.LinkType,
		Provider:       p.Provider,
		RTTIdleP50:     idleSummary.Median,
		RTTIdleP95:     idleSummary.P95,
		RTTLoadedP50:   loadedSummary.Median,
		RTTLoadedP95:   loadedSummary.P95,
		ThroughputMbps: throughput,
		LossPct:        lossPct,
		HTTPSmallP95:   sSummary.P95,
		DataUsedMB:     dataUsed,
		Notes:          strings.Join(notes, " ; "),
		// Note Waveform : écart p95 idle → chargé. L'audit charge en
		// upload (bulk client → sink) ; le sens descendant reste non mesuré
		// jusqu'à la campagne download.
		BloatDeltaMs:    bloatDelta,
		BloatGrade:      grade,
		BloatVerdict:    verdict,
		DrainMs:         drainMs,
		DrainMeasured:   drainMeasured,
		LoadSeconds:     bulkElapsed.Seconds(),
		BaselineSeconds: float64(idleSecs),
		RecoverySeconds: float64(recSecs),
	}
	res.ProfileMatch, res.MatchDelta = MatchProfile(res)
	return res, nil
}

// phasePlan returns the baseline, loaded and drain windows for an audit.
// The thesis default is the non-intrusive 60 s shape: 10 s quiescent,
// 40 s saturated, 10 s drain. Shorter runs scale the first two windows and
// keep a drain only when the budget can afford one.
func phasePlan(duration int, withLoad bool) (idle, loaded, rec int) {
	if duration <= 0 {
		duration = 10
	}
	if !withLoad {
		return duration, 0, 0
	}
	switch {
	case duration >= 60:
		idle, rec = 10, 10
		loaded = duration - idle - rec
	case duration >= 20:
		idle, rec = duration/4, duration/4
		loaded = duration - idle - rec
	default:
		idle = duration / 2
		loaded = duration - idle
	}
	if idle < 1 {
		idle = 1
	}
	if loaded < 1 {
		loaded = 1
	}
	return idle, loaded, rec
}

// lossPercent — probes received vs probes attempted. Recovery probes are
// attempts too: excluding them from the numerator turned every drain phase
// into invented packet loss.
func lossPercent(expected, actual float64) float64 {
	if expected <= 0 || actual >= expected {
		return 0
	}
	pct := (expected - actual) / expected * 100
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}
