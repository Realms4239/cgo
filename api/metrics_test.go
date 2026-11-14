package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func scrapeMetrics(t *testing.T, d Deps) string {
	t.Helper()
	rec := httptest.NewRecorder()
	New(d).ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain; version=0.0.4") {
		t.Fatalf("bad Content-Type: %s", ct)
	}
	return rec.Body.String()
}

func TestMetrics_EmitsMeasuredSeries(t *testing.T) {
	body := scrapeMetrics(t, Deps{Version: "v9.9.9", Mode: "observe"})
	for _, want := range []string{
		"# TYPE cgo_build_info gauge",
		`version="v9.9.9"`,
		`mode="observe"`,
		"cgo_uptime_seconds{} ",
		"cgo_sse_subscribers{} 0",
		"cgo_campaign_running{} 0",
		"cgo_kernel_source_info{",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in /metrics:\n%s", want, body)
		}
	}
}

func TestMetrics_CampaignRunningIsMeasured(t *testing.T) {
	body := scrapeMetrics(t, Deps{RunningFn: func() bool { return true }})
	if !strings.Contains(body, "cgo_campaign_running{} 1") {
		t.Fatalf("running campaign must read 1, body:\n%s", body)
	}
}

func TestMetrics_NoFabricatedFlags(t *testing.T) {
	body := scrapeMetrics(t, Deps{})
	for _, banned := range []string{
		"cgo_audit_load_measured_total",
		"cgo_cap_sla_measured",
		"cgo_iface_valid",
		"cgo_jfi_measured",
	} {
		if strings.Contains(body, banned) {
			t.Fatalf("%s was a hardcoded 0 flag, never measured; it must not be emitted", banned)
		}
	}
}
