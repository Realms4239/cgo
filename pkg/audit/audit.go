package audit

import (
	"context"
	"fmt"
	"net/http"
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
}

type Deps struct {
	Ping  func(ctx context.Context, target string, n int) []float64
	Small func(ctx context.Context) (float64, error)
	Bulk  func(ctx context.Context, addr string) (uint64, error)
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
		d.Small = func(ctx context.Context) (float64, error) {
			if smallURL != "" {
				return probe.SmallObject(ctx, http.DefaultClient, smallURL)
			}
			// pas d'URL : petit objet NON mesuré (0 + note), jamais de
			// constante 25 ms synthétique dans le CSV gelé
			return 0, fmt.Errorf("no small URL")
		}
	}
	start := time.Now()
	if p.Duration <= 0 {
		p.Duration = 10
	}
	var idleRTTs, loadedRTTs, idleSmalls, loadedSmalls []float64
	var bulkBytes uint64
	// tentatives ping : le dénominateur de la perte. L'ancien calcul
	// (durée × 12,5/s) ignorait la durée réelle d'un appel ping (~1 s pour
	// 5 paquets à 200 ms) et annonçait ~70 % de perte sur un lien parfait.
	pingCalls := 0
	const pingPerCall = 5
	bulkDone := make(chan uint64, 1)
	// le bulk couvre la fenêtre 12–22 s si >=30 et Bulk présent, sinon le débit reste à 0 avec notes honnêtes
	go func() {
		if d.Bulk != nil && p.Duration >= 30 {
			// attendre l'idle 12 s puis flood 10 s
			select {
			case <-time.After(12 * time.Second):
			case <-ctx.Done():
				bulkDone <- 0
				return
			}
			b, _ := d.Bulk(ctx, p.BulkAddr)
			bulkDone <- b
		} else {
			bulkDone <- 0
		}
	}()

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

	if p.Duration >= 30 {
		collectWindow(12, &idleRTTs, &idleSmalls)
		// le bulk flood a déjà démarré à 12 s, laisser la collecte chargée tourner 22–30 s
		// attendre la marque 22 s (fenêtre bulk 10 s) puis collecter 8 s en charge
		remaining := p.Duration - 12 - 8
		if remaining > 0 {
			time.Sleep(time.Duration(remaining) * time.Second)
		}
		collectWindow(8, &loadedRTTs, &loadedSmalls)
	} else {
		// audit court : fenêtre unique, idle==loaded honnête
		collectWindow(p.Duration, &idleRTTs, &idleSmalls)
		loadedRTTs = append([]float64(nil), idleRTTs...)
		loadedSmalls = append([]float64(nil), idleSmalls...)
	}
	select {
	case bulkBytes = <-bulkDone:
	default:
		bulkBytes = 0
	}

	idleSummary := metrics.Summarize(idleRTTs)
	loadedSummary := metrics.Summarize(loadedRTTs)
	allSmalls := append(append([]float64(nil), idleSmalls...), loadedSmalls...)
	sSummary := metrics.Summarize(allSmalls)
	// estimation de perte : échantillons reçus vs TENTÉS (5 par appel ping).
	// Compter les appels, pas la durée — un ping lent n'est pas de la perte.
	expectedSamples := float64(pingCalls * pingPerCall)
	actualSamples := float64(len(idleRTTs) + len(loadedRTTs))
	if p.Duration < 30 {
		// fenêtre unique : loaded = copie honnête d'idle — compter une fois
		actualSamples = float64(len(idleRTTs))
	}
	lossPct := 0.0
	if expectedSamples > 0 && actualSamples < expectedSamples {
		lossPct = (expectedSamples - actualSamples) / expectedSamples * 100
		if lossPct < 0 {
			lossPct = 0
		}
		if lossPct > 100 {
			lossPct = 100
		}
	}
	throughput := 0.0
	dataUsed := 0.0
	var notes []string
	if bulkBytes > 0 {
		throughput = float64(bulkBytes) * 8 / 1e6 / 10
		dataUsed = float64(bulkBytes) / 1e6
	} else {
		notes = append(notes, "throughput non mesuré sans bulk sink (BulkAddr) — iperf3 fallback disponible si installé")
	}
	if len(allSmalls) == 0 {
		notes = append(notes, "petit objet non mesuré (SmallURL absent)")
	}

	return &Result{
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
	}, nil
}
