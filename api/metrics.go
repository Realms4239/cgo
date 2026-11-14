package api

import (
	"strconv"
	"time"

	"github.com/Realms4239/cgo/exporter"
)

// metricsSamples — les séries que ce serveur mesure vraiment. Aucune valeur
// inventée : une série absente n'est jamais émise à zéro (contrat exporter).
func (s *Server) metricsSamples() (map[string][]exporter.Sample, map[string]string) {
	running := "0"
	if s.deps.RunningFn != nil && s.deps.RunningFn() {
		running = "1"
	}
	ks := s.kernel
	samples := map[string][]exporter.Sample{
		"cgo_build_info": {{
			Labels: map[string]string{"version": s.deps.Version, "mode": s.deps.Mode},
			Value:  "1",
		}},
		"cgo_uptime_seconds": {{
			Value: strconv.FormatFloat(time.Since(s.started).Seconds(), 'f', 3, 64),
		}},
		"cgo_sse_subscribers": {{
			Value: strconv.Itoa(s.hub.Subscribers()),
		}},
		"cgo_campaign_running": {{
			Value: running,
		}},
		"cgo_kernel_source_info": {{
			Labels: map[string]string{"source": ks.Source, "reason": ks.Reason},
			Value:  "1",
		}},
	}
	help := map[string]string{
		"cgo_build_info":         "Build identity (version, mode); always 1",
		"cgo_uptime_seconds":     "Seconds since the API server started",
		"cgo_sse_subscribers":    "Live SSE dashboard subscribers",
		"cgo_campaign_running":   "1 while a campaign matrix runs, 0 otherwise",
		"cgo_kernel_source_info": "Kernel telemetry source chosen at startup; always 1",
	}
	return samples, help
}
