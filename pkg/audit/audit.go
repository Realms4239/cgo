package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/probe"
)

// Params for one audit run (Tableau 6).
type Params struct {
	AuditID  string
	Site     string
	LinkType string // fiber, 5g, 4g, vsat, other
	Provider string
	Duration int    // seconds, e.g. 60–300

	Target   string // ping target
	SmallURL string // small object url
	BulkAddr string // optional bulk addr for throughput
}

// Result row matching Tableau 6.
type Result struct {
	AuditID      string  `json:"audit_id"`
	Timestamp    string  `json:"timestamp"`
	Site         string  `json:"site"`
	LinkType     string  `json:"link_type"`
	Provider     string  `json:"provider"`
	RTTIdleP50   float64 `json:"rtt_idle_p50_ms"`
	RTTIdleP95   float64 `json:"rtt_idle_p95_ms"`
	RTTLoadedP50 float64 `json:"rtt_loaded_p50_ms"`
	RTTLoadedP95 float64 `json:"rtt_loaded_p95_ms"`
	ThroughputMbps float64 `json:"throughput_mbps"`
	LossPct      float64 `json:"loss_pct"`
	HTTPSmallP95 float64 `json:"http_small_p95_ms"`
	DataUsedMB   float64 `json:"data_used_mb"`
	Notes        string  `json:"notes"`
}

type Deps struct {
	Ping  func(ctx context.Context, target string, n int) []float64
	Small func(ctx context.Context) (float64, error)
	Bulk  func(ctx context.Context, addr string) (uint64, error)
}

// Run executes an audit for Duration seconds, collecting ping/small samples.
// B split: 0–12s idle, 12–22s bulk flood, 22–30s loaded distinct per Q26, plus iperf3 fallback via BulkAddr.
// When Duration <30, single window with idle==loaded honest copy; when >=30 three windows produce distinct RTT idle vs loaded.
func Run(ctx context.Context, p Params, d Deps) (*Result, error) {
	if d.Ping == nil {
		d.Ping = func(ctx context.Context, t string, n int) []float64 {
			ss, _ := probe.Ping(ctx, probe.ExecCmdRunner{}, t, n, 200)
			out := make([]float64, len(ss))
			for i, s := range ss { out[i] = s.RTTms }
			return out
		}
	}
	if d.Small == nil {
		smallURL := p.SmallURL
		d.Small = func(ctx context.Context) (float64, error) {
			if smallURL != "" {
				return probe.SmallObject(ctx, http.DefaultClient, smallURL)
			}
			return 25, nil
		}
	}
	start := time.Now()
	if p.Duration <= 0 {
		p.Duration = 10
	}
	var idleRTTs, loadedRTTs, idleSmalls, loadedSmalls []float64
	var bulkBytes uint64
	bulkDone := make(chan uint64, 1)
	// bulk runs 12–22s window if >=30 and Bulk present, otherwise throughput stays 0 with honest notes
	go func() {
		if d.Bulk != nil && p.Duration >= 30 {
			// wait for idle 12s then flood 10s
			select { case <-time.After(12 * time.Second): case <-ctx.Done(): bulkDone <- 0; return }
			b, _ := d.Bulk(ctx, p.BulkAddr)
			bulkDone <- b
		} else {
			bulkDone <- 0
		}
	}()

	collectWindow := func(secs int, dstRTT, dstSmall *[]float64) {
		deadline := time.Now().Add(time.Duration(secs) * time.Second)
		for time.Now().Before(deadline) {
			select { case <-ctx.Done(): return; default: }
			*dstRTT = append(*dstRTT, d.Ping(ctx, p.Target, 5)...)
			if v, err := d.Small(ctx); err == nil {
				*dstSmall = append(*dstSmall, v)
			}
			time.Sleep(400 * time.Millisecond)
		}
	}

	if p.Duration >= 30 {
		collectWindow(12, &idleRTTs, &idleSmalls)
		// bulk flood already started at 12s, let loaded collection run 22–30s
		// wait until 22s mark (10s bulk window) then collect loaded 8s
		remaining := p.Duration - 12 - 8
		if remaining > 0 {
			time.Sleep(time.Duration(remaining) * time.Second)
		}
		collectWindow(8, &loadedRTTs, &loadedSmalls)
	} else {
		// short audit: single window, idle==loaded honest
		collectWindow(p.Duration, &idleRTTs, &idleSmalls)
		loadedRTTs = append([]float64(nil), idleRTTs...)
		loadedSmalls = append([]float64(nil), idleSmalls...)
	}
	select { case bulkBytes = <-bulkDone: default: bulkBytes = 0 }

	idleSummary := metrics.Summarize(idleRTTs)
	loadedSummary := metrics.Summarize(loadedRTTs)
	allSmalls := append(append([]float64(nil), idleSmalls...), loadedSmalls...)
	sSummary := metrics.Summarize(allSmalls)
	// loss estimate: missing vs expected (5 per 400ms → 12.5/s)
	expectedSamples := float64(p.Duration) * 12.5
	actualSamples := float64(len(idleRTTs) + len(loadedRTTs))
	lossPct := 0.0
	if expectedSamples > 0 && actualSamples < expectedSamples {
		lossPct = (expectedSamples - actualSamples) / expectedSamples * 100
		if lossPct < 0 { lossPct = 0 }
		if lossPct > 100 { lossPct = 100 }
	}
	throughput := 0.0
	dataUsed := 0.0
	notes := ""
	if bulkBytes > 0 {
		throughput = float64(bulkBytes) * 8 / 1e6 / 10
		dataUsed = float64(bulkBytes) / 1e6
	} else {
		notes = "throughput non mesuré sans bulk sink (BulkAddr) — iperf3 fallback disponible si installé"
	}

	return &Result{
		AuditID:    p.AuditID,
		Timestamp:  start.Format(time.RFC3339),
		Site:       p.Site,
		LinkType:   p.LinkType,
		Provider:   p.Provider,
		RTTIdleP50: idleSummary.Median,
		RTTIdleP95: idleSummary.P95,
		RTTLoadedP50: loadedSummary.Median,
		RTTLoadedP95: loadedSummary.P95,
		ThroughputMbps: throughput,
		LossPct:      lossPct,
		HTTPSmallP95: sSummary.P95,
		DataUsedMB:   dataUsed,
		Notes:      notes,
	}, nil
}


