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

// escapeMetricText escapes \, " and \n for label values / HELP text.
func escapeMetricText(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return r.Replace(s)
}

func validMetricName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		switch {
		case c == '_' || c == ':' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
		case c >= '0' && c <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

func validLabelName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		switch {
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
		case c >= '0' && c <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// renderMetrics — pure Prometheus exposition builder.
// Absent is never emitted; invalid metric/label names are skipped and
// label values / HELP are escaped (no exposition injection).
func renderMetrics(samples map[string][]metricSample, help map[string]string) string {
	var b strings.Builder
	names := make([]string, 0, len(samples))
	for n := range samples {
		if validMetricName(n) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, escapeMetricText(help[name]))
		typ := "gauge"
		if strings.HasSuffix(name, "_total") {
			typ = "counter"
		}
		fmt.Fprintf(&b, "# TYPE %s %s\n", name, typ)
		for _, s := range samples[name] {
			var parts []string
			for k, v := range s.labels {
				if validLabelName(k) {
					parts = append(parts, fmt.Sprintf(`%s="%s"`, k, escapeMetricText(v)))
				}
			}
			sort.Strings(parts)
			fmt.Fprintf(&b, "%s{%s} %s\n", name, strings.Join(parts, ","), s.value)
		}
	}
	return b.String()
}

// FACADE METRICS — no live audit counters are wired yet, so value gauges
// are omitted (absent is never a zero). Only *_measured / cgo_iface_valid
// absent flags (0) are emitted; only those may drive alerts.
// Named condition for closure:
//
//	BLOCKED(facade-metrics): wire these series from the audit engine
//	(cgo_audit_runs_total, cgo_audit_load_attempted/measured_total,
//	cgo_cap_sla_measured, cgo_iface_valid, cgo_jfi_measured) so omitted
//	value gauges become measured values.
func handleMetrics(getBuildInfo func() map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		samples := map[string][]metricSample{
			"cgo_cap_sla_measured": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_iface_valid":      {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_jfi_measured":     {{labels: map[string]string{"site": "unknown"}, value: "0"}},
		}
		help := map[string]string{
			"cgo_cap_sla_measured": "CAP SLA measured flag",
			"cgo_iface_valid":      "Interface sampling valid flag",
			"cgo_jfi_measured":     "JFI measured flag (0 means absent, never a zero value)",
		}
		_, _ = w.Write([]byte(renderMetrics(samples, help)))
	}
}
