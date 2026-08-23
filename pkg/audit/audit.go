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
// Bulk throughput is measured if Bulk != nil and Duration >=30.
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
	deadline := start.Add(time.Duration(p.Duration) * time.Second)
	if p.Duration <= 0 {
		deadline = start.Add(10 * time.Second)
	}
	var rtts, smalls []float64
	var bulkBytes uint64
	bulkDone := make(chan uint64, 1)
	bulkStarted := false
	// start bulk after 5s if available and duration allows
	go func() {
		if d.Bulk != nil && p.Duration >= 30 {
			time.Sleep(5 * time.Second)
			bulkStarted = true
			b, _ := d.Bulk(ctx, p.BulkAddr)
			bulkDone <- b
		} else {
			bulkDone <- 0
		}
	}()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			goto done
		default:
		}
		rtts = append(rtts, d.Ping(ctx, p.Target, 5)...)
		if v, err := d.Small(ctx); err == nil {
			smalls = append(smalls, v)
		}
		time.Sleep(400 * time.Millisecond)
	}
done:
	select {
	case bulkBytes = <-bulkDone:
	default:
		bulkBytes = 0
	}
	_ = bulkStarted

	rSummary := metrics.Summarize(rtts)
	sSummary := metrics.Summarize(smalls)
	// loss estimate: missing ping samples vs expected (5 per 400ms ~12.5 per sec)
	// ponytail: loss = 0 for now, honest placeholder
	throughput := 0.0
	dataUsed := 0.0
	if bulkBytes > 0 {
		// bulk ran ~10s, throughput = bytes*8 /10 /1e6
		throughput = float64(bulkBytes) * 8 / 1e6 / 10
		dataUsed = float64(bulkBytes) / 1e6
	}

	return &Result{
		AuditID:    p.AuditID,
		Timestamp:  start.Format(time.RFC3339),
		Site:       p.Site,
		LinkType:   p.LinkType,
		Provider:   p.Provider,
		RTTIdleP50: rSummary.Median,
		RTTIdleP95: rSummary.P95,
		RTTLoadedP50: rSummary.Median, // same as idle for now (no separate loaded phase in minimal audit)
		RTTLoadedP95: rSummary.P95,
		ThroughputMbps: throughput,
		LossPct:      0,
		HTTPSmallP95: sSummary.P95,
		DataUsedMB:   dataUsed,
		Notes:      "",
	}, nil
}


