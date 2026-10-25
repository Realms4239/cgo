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
	// Direction — sens de la charge : "" ou "up" (upload), "down"
	// (download — le scénario DGM : l'image EUMETSAT qui déborde la file),
	// "both" (deux fenêtres de charge successives, un seul rapport).
	Direction string
	// Flows — flux de charge concurrents (défaut 4, régime du banc durci).
	Flows int
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

	// Direction — sens de la charge réellement exercé, gelé avec la ligne :
	// "up", "down" ou "both". Les champs headline (débit, note bufferbloat,
	// drain, JFI) portent le sens mesuré : la fenêtre up pour up/both, la
	// fenêtre down pour un audit down ; les champs miroir ci-dessous portent
	// toujours la fenêtre descendante quand elle a tourné.
	Direction string `json:"direction"`
	// FlowCount — flux de charge concurrents effectivement utilisés.
	FlowCount int `json:"flow_count"`

	// Note bufferbloat (méthode Waveform, bandes verbatim citées) : écart
	// p95 idle → chargé. Le sens de la fenêtre headline est dans Direction.
	BloatDeltaMs float64 `json:"bloat_delta_ms"`
	BloatGrade   string  `json:"bloat_grade"`
	BloatVerdict string  `json:"bloat_verdict"`

	// Champs miroir — fenêtre DESCENDANTE (Direction "down" ou "both").
	// Zéro tant que la fenêtre down n'a pas tourné : un zéro ici est une
	// absence de mesure, jamais une mesure nulle.
	RTTLoadedDownP50   float64 `json:"rtt_loaded_down_p50_ms"`
	RTTLoadedDownP95   float64 `json:"rtt_loaded_down_p95_ms"`
	ThroughputDownMbps float64 `json:"throughput_down_mbps"`
	BloatDeltaDownMs   float64 `json:"bloat_delta_down_ms"`
	GradeDown          string  `json:"grade_down"`
	JFIDownPct         float64 `json:"jfi_down_pct"`
	DrainDownMs        float64 `json:"drain_down_ms"`
	DrainDownMeasured  bool    `json:"drain_down_measured"`

	// JFIPct — équité inter-flux de la fenêtre headline (Jain fairness
	// index, 0–100). 0 = un seul flux ou charge non mesurée.
	JFIPct float64 `json:"jfi_pct"`
	// RetransRatio — part des octets réémis sur la charge, base explicite :
	// "measured" (TCP_INFO sur nos sockets), "estimated" (proxy documenté
	// perte sonde × volume), "" (rien à chiffrer). Pour "both", une seule
	// ratio headline : la base mesurée gagne, à base égale la plus grande
	// des deux fenêtres.
	RetransRatio metrics.RetransRatio `json:"retrans_ratio"`

	// Rapprochement au référentiel P1–P4 (calibration par site) : calculé
	// en fin d'audit, gelé avec la ligne comme tout le reste.
	ProfileMatch string `json:"profile_match"`
	MatchDelta   string `json:"match_delta"`

	// DrainMs — temps de retour au repos après arrêt de la charge, en ms
	// au-dessus du seuil de repos. 0 = non mesuré (pas de phase 3), jamais
	// « drain instantané ». Mesure la vitesse à laquelle la file se vide.
	// Pour "both" : mini-drain après la fenêtre up ; la fenêtre down a son
	// propre drain dans DrainDownMs.
	DrainMs       float64 `json:"drain_ms"`
	DrainMeasured bool    `json:"drain_measured"`

	// SLA CAP (modèle C) : complétions 16 KiB CHARGÉES contre l'objectif
	// DGM 220 ms. CAPSLAMeasured=false ⇒ colonnes vides dans le CSV, jamais
	// un « 0 % » qui se lirait comme un manquement mesuré — une absence,
	// pas un zéro. Le deadline_ok_pct de la campagne opérateur reste une
	// métrique séparée, indisponible dans un audit de lien.
	CAPSLAMeasured   bool    `json:"cap_sla_measured"`
	CAPDeadlineOKPct float64 `json:"cap_deadline_ok_pct"`
	CAPSLABreach     bool    `json:"cap_sla_breach"`
	CAPSLALabel      string  `json:"cap_sla_label,omitempty"`

	// LoadSeconds is the measured duration of the bulk transfer, in seconds.
	// The financial projection extrapolates from this rather than from a
	// hardcoded window, so the hourly figure is traceable to a real interval.
	// For Direction "both" it is the sum of both load windows.
	LoadSeconds float64 `json:"load_seconds"`
	// Phases réellement exécutées. 60 s = 10 s baseline, 40 s charge,
	// 10 s drain; les audits courts dimensionnent. Stocké pour que le
	// rapport prouve que la note vient d'une fenêtre chargée.
	BaselineSeconds float64 `json:"baseline_seconds"`
	RecoverySeconds float64 `json:"recovery_seconds"`

	// Telemetry — la série brute par sonde (phase, offset, RTT, complétion
	// 16 KiB). C'est la matière première que le double sceau SHA-256 scelle :
	// l'absence de sonde y est une absence de ligne, jamais un 0 ms.
	Telemetry []TelemetryRow `json:"telemetry,omitempty"`
}

type Deps struct {
	Ping     func(ctx context.Context, target string, n int) []float64
	Small    func(ctx context.Context) (float64, error)
	BulkLoad func(ctx context.Context, addr, direction string, flows int) LoadOutcome
}

// LoadOutcome — le résultat d'une fenêtre de charge : total, par flux (JFI),
// réémissions (modèle A) et durée réelle. La durée sert au débit : diviser
// par une constante de la fenêtre inventait un débit faux quand la charge
// durait 8 s pour 10 s prévus. Err non-nil = charge échouée.
type LoadOutcome struct {
	Bytes           uint64
	PerFlow         []uint64
	RetransBytes    uint64
	RetransMeasured bool
	Err             error
	Elapsed         time.Duration
}

// Run exécute un audit de Duration secondes, en collectant des échantillons ping/small.
// Découpe B : 0–12 s idle, 12–22 s bulk flood, 22–30 s chargé distinct selon Q26, plus repli iperf3 par BulkAddr.
// Si Duration <30, fenêtre unique avec idle==loaded copie honnête ; si >=30, trois fenêtres produisent des RTT idle vs chargé distincts.
// Direction "down" exerce la fenêtre descendante (champs miroir + headline),
// "both" enchaîne fenêtre up puis fenêtre down sur une seule baseline partagée.
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
	// sans chargeur câblé, un BulkAddr suffit à activer la charge réelle :
	// c'est le cas nominal du CLI et de l'API, qui n'ont pas à connaître
	// l'implémentation du puits. BBR = contrôle de congestion de la
	// campagne de référence (RQ1) ; un flux cubic sur 100 ms de RTT ne
	// remplit pas la capacité et ne crée aucune file — l'audit mesurerait
	// un lien vide puis noterait A+ à tort. Le sens "down" passe par le
	// puits récepteur (le scénario DGM : l'image qui déborde la file), le
	// sens "up" par l'envoi ; 4 flux par défaut, le régime du banc durci.
	if d.BulkLoad == nil && p.BulkAddr != "" {
		d.BulkLoad = func(ctx context.Context, addr, direction string, flows int) LoadOutcome {
			t0 := time.Now()
			var st probe.BulkStats
			var err error
			if direction == "down" {
				st, err = probe.BulkDownloadNStats(ctx, addr, "bbr", flows)
			} else {
				st, err = probe.BulkSendNStats(ctx, addr, "bbr", flows)
			}
			out := LoadOutcome{Bytes: sumU64(st.Bytes), PerFlow: st.Bytes,
				RetransBytes: st.RetransBytes, RetransMeasured: st.RetransMeasured,
				Elapsed: time.Since(t0)}
			if len(st.Bytes) == 0 && err == nil {
				err = fmt.Errorf("aucun flux établi vers %s", addr)
			}
			out.Err = err
			return out
		}
	}
	if p.Flows <= 0 {
		p.Flows = 4
	}
	switch p.Direction {
	case "", "up":
		p.Direction = "up"
	case "down", "both":
	default:
		return nil, fmt.Errorf("audit %s : direction %q invalide (up|down|both)", p.AuditID, p.Direction)
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
	withLoad := d.BulkLoad != nil && p.Duration >= 2
	idleSecs, loadedSecs, recSecs := phasePlan(p.Duration, withLoad)

	// One audit-wide deadline: a caller may pass context.Background() (the
	// HTTP handler does), and a stuck peer must not leave a goroutine and a
	// half-written CSV behind forever.
	auditCtx, auditCancel := context.WithTimeout(ctx, time.Duration(p.Duration+30)*time.Second)
	defer auditCancel()
	ctx = auditCtx

	var idleRTTs, idleSmalls, upRTTs, upSmalls, downRTTs, downSmalls []float64
	var recRTTs, recDownRTTs []float64
	var upOc, downOc LoadOutcome
	var notes []string
	// tentatives ping : le dénominateur de la perte. L'ancien calcul
	// (durée × 12,5/s) ignorait la durée réelle d'un appel ping (~1 s pour
	// 5 paquets à 200 ms) et annonçait ~70 % de perte sur un lien parfait.
	pingCalls := 0
	const pingPerCall = 5
	// les small de récupération sont sondés (le rythme est celui de la
	// fenêtre) mais ne nourrissent pas le p95 — sémantique inchangée depuis
	// la découpe en phases.
	var discardSmalls []float64
	// telemetry — la série brute (une ligne par sonde réellement reçue).
	var telemetry []TelemetryRow

	collectWindow := func(phase string, secs int, dstRTT, dstSmall *[]float64) {
		deadline := time.Now().Add(time.Duration(secs) * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			pingCalls++
			offset := time.Since(start).Milliseconds()
			for _, s := range d.Ping(ctx, p.Target, pingPerCall) {
				*dstRTT = append(*dstRTT, s)
				if s != 0 {
					telemetry = append(telemetry, TelemetryRow{Phase: phase, OffsetMs: offset, RTTms: s})
				}
			}
			if v, err := d.Small(ctx); err == nil {
				*dstSmall = append(*dstSmall, v)
				if v != 0 {
					telemetry = append(telemetry, TelemetryRow{Phase: phase, OffsetMs: offset, SmallMs: v})
				}
			}
			time.Sleep(400 * time.Millisecond)
		}
	}

	// runLoadWindow — UNE fenêtre de charge : démarre le chargeur du sens
	// demandé, attend 300 ms (dial/flood), collecte secs secondes de sondes
	// PENDANT la charge, annule, puis récupère le verdict du chargeur
	// (5 s de grâce après annulation). Renvoie les échantillons chargés de
	// cette fenêtre et le résultat du chargeur.
	runLoadWindow := func(dir string, secs int) ([]float64, []float64, LoadOutcome) {
		var rtt, small []float64
		wctx, wcancel := context.WithCancel(ctx)
		defer wcancel()
		done := make(chan LoadOutcome, 1)
		t0 := time.Now()
		go func() {
			oc := d.BulkLoad(wctx, p.BulkAddr, dir, p.Flows)
			if oc.Elapsed <= 0 {
				// chargeur sans durée : mesurée autour de l'appel
				oc.Elapsed = time.Since(t0)
			}
			done <- oc
		}()
		// laisser le dial/flood démarrer avant de compter les pings chargés
		time.Sleep(300 * time.Millisecond)
		phase := "load"
		if dir == "down" {
			phase = "load-down"
		}
		collectWindow(phase, secs, &rtt, &small)
		wcancel()
		select {
		case oc := <-done:
			return rtt, small, oc
		case <-time.After(5 * time.Second):
			notes = append(notes, "charge "+dir+" : pas de retour après annulation")
			return rtt, small, LoadOutcome{}
		}
	}

	// both : le budget de charge se partage entre les deux sens ; le
	// mini-drain (5 s) et le répit (3 s) entre les sens sont prélevés sur
	// ce budget pour que l'audit reste dans sa durée déclarée.
	loadedUpSecs, loadedDownSecs := loadedSecs, 0
	const bothSettleSecs = 3
	const bothMiniDrainSecs = 5
	if p.Direction == "both" {
		budget := loadedSecs - bothSettleSecs - bothMiniDrainSecs
		if budget < 2 {
			budget = 2
		}
		loadedUpSecs = budget / 2
		loadedDownSecs = budget - loadedUpSecs
	}

	if withLoad {
		// UNE baseline partagée : idle vs chargé se compare toujours au
		// même repos, quel que soit le nombre de fenêtres de charge.
		collectWindow("baseline", idleSecs, &idleRTTs, &idleSmalls)
		switch p.Direction {
		case "both":
			upRTTs, upSmalls, upOc = runLoadWindow("up", loadedUpSecs)
			// Phase 3 (up) — mini-drain : le temps de retour au repos après
			// la fenêtre montante nourrit DrainMs avant d'enchaîner.
			collectWindow("drain", bothMiniDrainSecs, &recRTTs, &discardSmalls)
			// répit : laisser la file vider avant de mesurer le sens
			// descendant, sinon la fenêtre down démarre sur un héritage up.
			time.Sleep(bothSettleSecs * time.Second)
			downRTTs, downSmalls, downOc = runLoadWindow("down", loadedDownSecs)
			// Phase 3 (down) — drain final après la fenêtre descendante.
			if recSecs > 0 {
				collectWindow("drain-down", recSecs, &recDownRTTs, &discardSmalls)
			}
		case "down":
			downRTTs, downSmalls, downOc = runLoadWindow("down", loadedSecs)
			// Phase 3 — drain. Après l'arrêt de la charge, on continue de
			// sonder : le temps que le p95 revienne sous le seuil de repos
			// est la vitesse à laquelle la file se vide. C'est la mesure qui
			// distingue un AQM qui tient la file d'un AQM qui l'a seulement
			// vidée à temps.
			if recSecs > 0 {
				collectWindow("drain-down", recSecs, &recDownRTTs, &discardSmalls)
			}
		default: // up
			upRTTs, upSmalls, upOc = runLoadWindow("up", loadedSecs)
			if recSecs > 0 {
				collectWindow("drain", recSecs, &recRTTs, &discardSmalls)
			}
		}
	} else {
		// audit court / sans bulk : fenêtre unique, idle==loaded honnête
		collectWindow("baseline", p.Duration, &idleRTTs, &idleSmalls)
		upRTTs = append([]float64(nil), idleRTTs...)
		upSmalls = append([]float64(nil), idleSmalls...)
	}

	idleSummary := metrics.Summarize(idleRTTs)
	upSummary := metrics.Summarize(upRTTs)
	downSummary := metrics.Summarize(downRTTs)
	// headline = le sens exercé : la fenêtre up pour up/both, la fenêtre
	// down pour un audit down (MatchProfile et BuildReport lisent les
	// champs headline ; Direction documente d'où ils viennent).
	headlineIsDown := p.Direction == "down"
	loadedSummary := upSummary
	if headlineIsDown {
		loadedSummary = downSummary
	}
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
	loadMeasuredUp := withLoad && upOc.Bytes > 0 && len(upRTTs) > 0
	loadMeasuredDown := withLoad && downOc.Bytes > 0 && len(downRTTs) > 0
	// loadMeasured — VRAI dès qu'UNE fenêtre exercée a mesuré une charge ;
	// les notes par fenêtre restent quand même honnêtes (voir plus bas).
	loadMeasured := loadMeasuredUp || loadMeasuredDown
	bloatDelta, bloatDeltaDown := 0.0, 0.0
	if loadMeasuredUp {
		bloatDelta = upSummary.P95 - idleSummary.P95
	}
	if loadMeasuredDown {
		bloatDeltaDown = downSummary.P95 - idleSummary.P95
	}
	if headlineIsDown {
		bloatDelta = bloatDeltaDown
	}
	allSmalls := append(append(append([]float64(nil), idleSmalls...), upSmalls...), downSmalls...)
	sSummary := metrics.Summarize(allSmalls)
	// estimation de perte : échantillons reçus vs TENTÉS (5 par appel ping).
	// Compter les appels, pas la durée — un ping lent n'est pas de la perte.
	expectedSamples := float64(pingCalls * pingPerCall)
	actualSamples := float64(len(idleRTTs))
	if withLoad {
		// chaque fenêtre (baseline, charges, drains) est comptée une fois ;
		// les fenêtres non exercées restent vides et ne comptent pour rien
		actualSamples = float64(len(idleRTTs) + len(upRTTs) + len(downRTTs) + len(recRTTs) + len(recDownRTTs))
	}
	lossPct := lossPercent(expectedSamples, actualSamples)
	throughput := throughputOf(upOc.Bytes, upOc.Elapsed, loadedUpSecs)
	throughputDown := throughputOf(downOc.Bytes, downOc.Elapsed, loadedDownSecs)
	if headlineIsDown {
		throughput = throughputDown
	}
	dataUsed := float64(upOc.Bytes+downOc.Bytes) / 1e6
	if withLoad {
		noteFailed := func(dir string, oc LoadOutcome) {
			if oc.Bytes > 0 {
				return
			}
			msg := "charge " + dir + " non mesurée — sink injoignable ou erreur"
			if oc.Err != nil {
				msg += " (" + oc.Err.Error() + ")"
			}
			notes = append(notes, msg)
		}
		if p.Direction != "down" {
			noteFailed("up", upOc)
		}
		if p.Direction != "up" {
			noteFailed("down", downOc)
		}
	} else {
		notes = append(notes, "throughput non mesuré sans bulk sink (BulkAddr) — iperf3 fallback disponible si installé")
	}
	grade, verdict, gradeDown := "", "", ""
	if headlineIsDown {
		if loadMeasuredDown {
			gradeDown = metrics.BufferbloatGrade(bloatDeltaDown)
			grade = gradeDown
			verdict = metrics.BufferbloatVerdict(grade)
		}
	} else {
		if loadMeasuredUp {
			grade = metrics.BufferbloatGrade(bloatDelta)
			verdict = metrics.BufferbloatVerdict(grade)
		}
		if loadMeasuredDown {
			gradeDown = metrics.BufferbloatGrade(bloatDeltaDown)
		}
		// both, fenêtre up muette : la note headline suit le sens up —
		// dire pourquoi elle manque alors que grade_down existe.
		if p.Direction == "both" && !loadMeasuredUp && loadMeasuredDown {
			notes = append(notes, "charge up non mesurée — la note headline suit la fenêtre up ; voir grade_down")
		}
	}
	if !loadMeasured {
		notes = append(notes, "bufferbloat non mesuré — aucune charge bulk exploitable")
		verdict = "Bufferbloat non mesuré — aucune charge bulk exploitable."
	}

	// Drain : combien de temps la latence met à revenir sous le seuil de
	// repos après l'arrêt de la charge. Mesuré seulement si la fenêtre de
	// récupération a produit des échantillons ; sinon 0 = non mesuré,
	// jamais « drain instantané ». Le seuil de retour au repos : le p95 de
	// repos plus la tolérance G6 de l'audit (5 ms ou 20 % de la médiane,
	// le plus grand des deux) — sur la baseline partagée.
	restThreshold := idleSummary.P95 + math.Max(5, 0.2*idleSummary.Median)
	drainOf := func(rtts []float64) (float64, bool) {
		if len(rtts) == 0 {
			return 0, false
		}
		d := metrics.Summarize(rtts).P95 - restThreshold
		if d < 0 {
			d = 0 // revenu sous le seuil : drain effectivement terminé
		}
		return d, true
	}
	drainMs, drainMeasured := 0.0, false
	drainDownMs, drainDownMeasured := 0.0, false
	switch p.Direction {
	case "both":
		// mini-drain après la fenêtre up → DrainMs ; drain final après la
		// fenêtre down → DrainDownMs. Chaque drain ne dit que son sens.
		if loadMeasuredUp {
			drainMs, drainMeasured = drainOf(recRTTs)
		}
		if loadMeasuredDown {
			drainDownMs, drainDownMeasured = drainOf(recDownRTTs)
		}
	case "down":
		if loadMeasuredDown {
			drainDownMs, drainDownMeasured = drainOf(recDownRTTs)
			// headline = le sens exercé : le drain descendante voyage
			// aussi dans DrainMs (BuildReport lit le headline).
			drainMs, drainMeasured = drainDownMs, drainDownMeasured
		}
	default: // up
		if loadMeasuredUp {
			drainMs, drainMeasured = drainOf(recRTTs)
		}
	}
	if len(allSmalls) == 0 {
		notes = append(notes, "petit objet non mesuré (SmallURL absent)")
	}

	// Ratio de réémission — APRÈS la perte : le proxy documenté en dépend.
	// Fenêtre par fenêtre : TCP_INFO d'abord ("measured"), sinon perte sonde
	// × volume ("estimated"), sinon rien. Pour "both" une seule ratio
	// headline : la base mesurée gagne, à base égale la plus grande des
	// deux fenêtres (choix conservateur).
	upRatio := windowRatio(upOc, loadMeasuredUp, lossPct)
	downRatio := windowRatio(downOc, loadMeasuredDown, lossPct)
	ratio := upRatio
	if headlineIsDown {
		ratio = downRatio
	} else if p.Direction == "both" {
		ratio = pickRatio(upRatio, downRatio)
	}

	jfiPct := metrics.JFI(upOc.PerFlow) * 100
	jfiDownPct := metrics.JFI(downOc.PerFlow) * 100
	if headlineIsDown {
		jfiPct = jfiDownPct
	}

	// SLA CAP — les complétions 16 KiB SOUS CHARGE : la fenêtre up pour un
	// audit up, la fenêtre down pour un audit down, les deux fenêtres
	// poolées pour « both » (chaque complétion chargée compte, quel que soit
	// le sens qui la provoquait). Sans échantillon chargé : non mesuré.
	capSamples := upSmalls
	switch p.Direction {
	case "down":
		capSamples = downSmalls
	case "both":
		capSamples = append(append([]float64(nil), upSmalls...), downSmalls...)
	}
	sla := metrics.AssessCAPSLA(capSamples)

	loadSeconds := upOc.Elapsed.Seconds()
	switch p.Direction {
	case "down":
		loadSeconds = downOc.Elapsed.Seconds()
	case "both":
		loadSeconds = upOc.Elapsed.Seconds() + downOc.Elapsed.Seconds()
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
		// Note Waveform : écart p95 idle → chargé, sur le sens de la
		// fenêtre headline (Direction). Les champs miroir portent la
		// fenêtre descendante quand elle a tourné.
		BloatDeltaMs:       bloatDelta,
		BloatGrade:         grade,
		BloatVerdict:       verdict,
		Direction:          p.Direction,
		FlowCount:          p.Flows,
		RTTLoadedDownP50:   downSummary.Median,
		RTTLoadedDownP95:   downSummary.P95,
		ThroughputDownMbps: throughputDown,
		BloatDeltaDownMs:   bloatDeltaDown,
		GradeDown:          gradeDown,
		JFIDownPct:         jfiDownPct,
		DrainDownMs:        drainDownMs,
		DrainDownMeasured:  drainDownMeasured,
		JFIPct:             jfiPct,
		RetransRatio:       ratio,
		CAPSLAMeasured:     sla.Measured,
		CAPDeadlineOKPct:   sla.OKPct,
		CAPSLABreach:       sla.Breach,
		CAPSLALabel:        sla.Label,
		DrainMs:            drainMs,
		DrainMeasured:      drainMeasured,
		LoadSeconds:        loadSeconds,
		BaselineSeconds:    float64(idleSecs),
		RecoverySeconds:    float64(recSecs),
		Telemetry:          telemetry,
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

// sumU64 — somme d'une tranche d'octets (total de la fenêtre multi-flux).
func sumU64(xs []uint64) uint64 {
	var s uint64
	for _, x := range xs {
		s += x
	}
	return s
}

// throughputOf — débit d'une fenêtre : octets réels / durée réelle. La
// durée vient du chargeur (LoadOutcome.Elapsed) ; le repli sur la fenêtre
// prévue ne sert que si le chargeur n'a rien mesuré.
func throughputOf(bytes uint64, elapsed time.Duration, windowSecs int) float64 {
	if bytes == 0 {
		return 0
	}
	secs := elapsed.Seconds()
	if secs <= 0 {
		secs = float64(windowSecs)
	}
	return float64(bytes) * 8 / 1e6 / secs
}

// windowRatio — ratio de réémission d'UNE fenêtre : TCP_INFO d'abord (base
// "measured"), sinon le proxy documenté perte sonde × volume ("estimated"),
// sinon "" (rien à chiffrer — pas de charge mesurée ou aucune perte).
func windowRatio(oc LoadOutcome, measured bool, lossPct float64) metrics.RetransRatio {
	if !measured {
		return metrics.RetransRatio{}
	}
	total := float64(oc.Bytes)
	switch {
	case oc.RetransMeasured && total > 0:
		return metrics.RetransRatio{Ratio: float64(oc.RetransBytes) / total, Basis: "measured"}
	case lossPct > 0 && total > 0:
		return metrics.RetransRatio{Ratio: lossPct / 100, Basis: "estimated"}
	}
	return metrics.RetransRatio{}
}

// pickRatio — pour "both", une seule ratio headline : la base mesurée gagne
// sur le proxy ; à base égale, la plus grande des deux fenêtres.
func pickRatio(a, b metrics.RetransRatio) metrics.RetransRatio {
	rank := func(r metrics.RetransRatio) int {
		switch r.Basis {
		case "measured":
			return 2
		case "estimated":
			return 1
		}
		return 0
	}
	if rank(a) != rank(b) {
		if rank(a) > rank(b) {
			return a
		}
		return b
	}
	if b.Ratio > a.Ratio {
		return b
	}
	return a
}
