package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics_EmitsHelpTypeAndContentType(t *testing.T) {
	h := New(Deps{Mode: "observe"})
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain; version=0.0.4") {
		t.Fatalf("bad Content-Type: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# HELP cgo_audit_runs_total") {
		t.Fatalf("missing HELP for cgo_audit_runs_total")
	}
	if !strings.Contains(body, "# TYPE cgo_audit_runs_total counter") {
		t.Fatalf("missing TYPE for cgo_audit_runs_total")
	}
}

func TestMetrics_AbsentIsNotZero(t *testing.T) {
	samples := map[string][]metricSample{
		"cgo_jfi_pct":      {},
		"cgo_jfi_measured": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
	}
	help := map[string]string{"cgo_jfi_pct": "JFI percent, only when measured"}
	out := renderMetrics(samples, help)
	if strings.Contains(out, "cgo_jfi_pct{") {
		t.Fatalf("absent JFI must not emit a zero sample")
	}
	if !strings.Contains(out, "cgo_jfi_measured") || !strings.Contains(out, "cgo_jfi_measured{") {
		t.Fatalf("absent JFI must still emit cgo_jfi_measured=0")
	}
}
