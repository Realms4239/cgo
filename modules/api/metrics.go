package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type metricSample struct {
	labels map[string]string
	value  string
}

// renderMetrics — pure Prometheus exposition builder.
// Absent is never emitted; the caller only pushes measured values.
func renderMetrics(samples map[string][]metricSample, help map[string]string) string {
	var b strings.Builder
	names := make([]string, 0, len(samples))
	for n := range samples {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help[name])
		typ := "gauge"
		if strings.HasSuffix(name, "_total") {
			typ = "counter"
		}
		fmt.Fprintf(&b, "# TYPE %s %s\n", name, typ)
		for _, s := range samples[name] {
			var parts []string
			for k, v := range s.labels {
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(parts)
			fmt.Fprintf(&b, "%s{%s} %s\n", name, strings.Join(parts, ","), s.value)
		}
	}
	return b.String()
}

func handleMetrics(getBuildInfo func() map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		samples := map[string][]metricSample{
			"cgo_audit_runs_total": {{labels: map[string]string{"status": "completed"}, value: "0"}},
			"cgo_audit_load_attempted_total": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_audit_load_measured_total":              {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_cap_sla_measured":                         {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_iface_valid":                              {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_jfi_measured":                             {{labels: map[string]string{"site": "unknown"}, value: "0"}},
		}
		help := map[string]string{
			"cgo_audit_runs_total":           "Total audits by status",
			"cgo_audit_load_attempted_total": "Load windows attempted",
			"cgo_audit_load_measured_total":  "Load windows measured",
			"cgo_cap_sla_measured":           "CAP SLA measured flag",
			"cgo_iface_valid":                "Interface sampling valid flag",
			"cgo_jfi_measured":               "JFI measured flag (0 means absent, never a zero value)",
		}
		_, _ = w.Write([]byte(renderMetrics(samples, help)))
	}
}
