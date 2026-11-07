package exporter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandlerHelpTypeContentType — HELP/TYPE avant échantillon + Content-Type.
func TestHandlerHelpTypeContentType(t *testing.T) {
	h := Handler(func() (map[string][]Sample, map[string]string) {
		return map[string][]Sample{
				"cgo_audit_runs_total": {{Labels: map[string]string{"status": "completed"}, Value: "1"}},
			}, map[string]string{
				"cgo_audit_runs_total": "Total audits by status",
			}
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain; version=0.0.4") {
		t.Fatalf("bad Content-Type: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# HELP cgo_audit_runs_total") ||
		!strings.Contains(body, "# TYPE cgo_audit_runs_total counter") {
		t.Fatalf("HELP/TYPE manquant : %q", body)
	}
}

// TestAbsentIsNotZero — série absente = aucun échantillon zéro.
func TestAbsentIsNotZero(t *testing.T) {
	out := Render(map[string][]Sample{
		"cgo_jfi_pct":      {},
		"cgo_jfi_measured": {{Labels: map[string]string{"site": "unknown"}, Value: "0"}},
	}, map[string]string{"cgo_jfi_pct": "JFI percent, only when measured"})
	if strings.Contains(out, "cgo_jfi_pct{") {
		t.Fatal("série absente : aucun échantillon ne doit être émis")
	}
	if !strings.Contains(out, "cgo_jfi_measured{") {
		t.Fatal("cgo_jfi_measured=0 (absent) doit être émis")
	}
}
