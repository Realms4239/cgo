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
	if !strings.Contains(body, "# HELP cgo_jfi_measured") {
		t.Fatalf("missing HELP for cgo_jfi_measured")
	}
	if !strings.Contains(body, "# TYPE cgo_jfi_measured gauge") {
		t.Fatalf("missing TYPE for cgo_jfi_measured")
	}
	if !strings.Contains(body, "# HELP cgo_audit_load_measured_total Load measured flag (0 means absent, never a count)") {
		t.Fatalf("cgo_audit_load_measured_total must be a documented measured-flag, not a fake zero gauge")
	}
	if !strings.Contains(body, `cgo_audit_load_measured_total{site="unknown"} 0`) {
		t.Fatalf("measured-flag series for CGO_probe_dead must be present as flag 0")
	}
	if strings.Contains(body, "cgo_audit_runs_total{") {
		t.Fatalf("unmeasured value gauge cgo_audit_runs_total must be absent, not zero")
	}
	if strings.Contains(body, "cgo_audit_load_attempted_total{") {
		t.Fatalf("unmeasured value gauge cgo_audit_load_attempted_total must be absent, not zero")
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
	if strings.Contains(out, "# HELP cgo_jfi_measured") {
		t.Fatalf("missing help must skip HELP line, never emit empty HELP")
	}
}
