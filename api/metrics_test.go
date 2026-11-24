package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/audit"
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

// writeSealedReport dépose un artefact de rapport réel dans le dossier conventionnel.
func writeSealedReport(t *testing.T, dir, id, raw string) {
	t.Helper()
	p := filepath.Join(dir, audit.DataDir(), "audits", id, audit.ReportArtifactName)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
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

// TestMetrics_AuditSealAndFidelityComeFromArtifact — les quatre séries d'audit,
// de sceau et de fidélité ne sont émises que depuis un rapport scellé présent.
func TestMetrics_AuditSealAndFidelityComeFromArtifact(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	stamp := time.Now().UTC().Format(time.RFC3339)
	raw := fmt.Sprintf(`{"audit_id":"audit-obs","timestamp":%q,`+
		`"invocation":{"origin":"http","sink_mode":"dualns"},`+
		`"provenance":{"source_file":"aqm_eval.csv","source_sha256":"aaaa","telemetry_file":"telemetry.csv","telemetry_sha256":"bbbb"},`+
		`"fidelity":{"emd":0.2807,"live_samples":30,"emulated_samples":30,"emd_ceiling":0.10,"basis":"live=srtt_us/1000 ms; emulated=samples_rtt_ms"}}`, stamp)
	writeSealedReport(t, dir, "audit-obs", raw)

	body := scrapeMetrics(t, Deps{})
	for _, want := range []string{
		"cgo_audit_info{",
		`audit_id="audit-obs"`,
		`source="dualns"`,
		`age_bucket="fresh"`,
		"cgo_seal_present{audit_id=\"audit-obs\"} 1",
		"cgo_emd{",
		`live_basis="live=srtt_us/1000 ms; emulated=samples_rtt_ms"`,
		`cgo_emd{live_basis="live=srtt_us/1000 ms; emulated=samples_rtt_ms"} 0.2807`,
		"cgo_emd_ceiling{} 0.1",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in /metrics:\n%s", want, body)
		}
	}
}

// TestMetrics_SealZeroComesFromRealUnsealedReport — un zéro n'est fabriqué que
// quand l'artefact réel porte l'absence de sceau, jamais quand il manque.
func TestMetrics_SealZeroComesFromRealUnsealedReport(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeSealedReport(t, dir, "audit-nu", `{"audit_id":"audit-nu","timestamp":"2026-11-24T10:00:00Z","provenance":{}}`)

	body := scrapeMetrics(t, Deps{})
	if !strings.Contains(body, "cgo_seal_present{audit_id=\"audit-nu\"} 0") {
		t.Fatalf("rapport réel non scellé doit publier 0, pas une absence :\n%s", body)
	}
	if strings.Contains(body, "cgo_emd{") {
		t.Fatalf("rapport sans fidélité ne doit publier aucune mesure EMD :\n%s", body)
	}
}

// TestMetrics_AbsenceOmitsSeries — sans artefact, aucune valeur d'audit, de
// sceau, de fidélité ou de porte n'est inventée : seules les déclarations
// HELP/TYPE restent, et le plafond déclaré reste lui une constante réelle.
func TestMetrics_AbsenceOmitsSeries(t *testing.T) {
	t.Chdir(t.TempDir())
	body := scrapeMetrics(t, Deps{})
	for _, banned := range []string{
		"cgo_audit_info{",
		"cgo_seal_present{",
		"cgo_emd{",
		"cgo_gate_status{",
		"cgo_kernel_truth_source_info{",
	} {
		if strings.Contains(body, banned) {
			t.Fatalf("%s émis sans artefact réel :\n%s", banned, body)
		}
	}
	for _, declared := range []string{
		"# TYPE cgo_audit_info gauge",
		"# TYPE cgo_seal_present gauge",
		"# TYPE cgo_emd gauge",
		"# TYPE cgo_emd_ceiling gauge",
		"# TYPE cgo_gate_status gauge",
		"# TYPE cgo_kernel_truth_source_info gauge",
		"cgo_emd_ceiling{} 0.1",
	} {
		if !strings.Contains(body, declared) {
			t.Fatalf("déclaration absente %q :\n%s", declared, body)
		}
	}
}

// TestMetrics_UnsafeAuditIDNeverBecomesLabel — un identifiant traversant dans
// un artefact importé est refusé : aucun label ne peut porter un chemin.
func TestMetrics_UnsafeAuditIDNeverBecomesLabel(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeSealedReport(t, dir, "audit-evil", `{"audit_id":"../escape","timestamp":"2026-11-24T10:00:00Z"}`)
	body := scrapeMetrics(t, Deps{})
	if strings.Contains(body, `../escape`) || strings.Contains(body, "cgo_audit_info{") {
		t.Fatalf("identifiant traversant publié comme label :\n%s", body)
	}
}

// TestMetrics_KernelTruthSourceOnlyWhenReal — la source de vérité noyau n'est
// publiée que lorsqu'un artefact scellé porte available=true. Sur un hôte sans
// noyau, le probe étiquette « tcpinfo » par défaut : ce n'est PAS une source.
func TestMetrics_KernelTruthSourceOnlyWhenReal(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeSealedReport(t, dir, "audit-kt", `{"audit_id":"audit-kt","timestamp":"2026-11-24T10:00:00Z","kernel_telemetry":{"available":false,"source":"ebpf","reason":"BTF absent"}}`)
	if body := scrapeMetrics(t, Deps{}); strings.Contains(body, "cgo_kernel_truth_source_info{") {
		t.Fatalf("source non disponible publiée :\n%s", body)
	}
	writeSealedReport(t, dir, "audit-kt", `{"audit_id":"audit-kt","timestamp":"2026-11-24T10:00:00Z","kernel_telemetry":{"available":true,"source":"ebpf"}}`)
	body := scrapeMetrics(t, Deps{})
	if !strings.Contains(body, `cgo_kernel_truth_source_info{source="ebpf"} 1`) {
		t.Fatalf("source disponible non publiée :\n%s", body)
	}
}

// TestMetrics_GateStatusFromRealArtifactOnly — les statuts 0/1 viennent d'un
// artefact écrit par le runner ; skip et inconnu restent des absences, et le
// corps du rapport n'invente aucun verdict.
func TestMetrics_GateStatusFromRealArtifactOnly(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	p := filepath.Join(dir, audit.DataDir(), "gate_status.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"frozen":"pass","links":"fail","tests":"skip","inconnu":"pass"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	body := scrapeMetrics(t, Deps{})
	for _, want := range []string{
		`cgo_gate_status{gate="frozen"} 1`,
		`cgo_gate_status{gate="links"} 0`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q :\n%s", want, body)
		}
	}
	for _, banned := range []string{`gate="tests"`, `gate="inconnu"`} {
		if strings.Contains(body, banned) {
			t.Fatalf("statut absent publié %q :\n%s", banned, body)
		}
	}

	// Oracle hostile : un artefact illisible ne produit aucun statut.
	if err := os.WriteFile(p, []byte(`{pas du json`), 0o644); err != nil {
		t.Fatal(err)
	}
	body = scrapeMetrics(t, Deps{})
	if strings.Contains(body, "cgo_gate_status{") {
		t.Fatalf("artefact invalide a produit des statuts :\n%s", body)
	}
}

// TestMetrics_LatestSealedReportWins — deux rapports réels : c'est le plus
// récent par horodatage qui est publié, pas l'ordre de lecture du disque.
func TestMetrics_LatestSealedReportWins(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeSealedReport(t, dir, "audit-old", `{"audit_id":"audit-old","timestamp":"2026-01-01T00:00:00Z"}`)
	writeSealedReport(t, dir, "audit-new", `{"audit_id":"audit-new","timestamp":"2026-12-01T00:00:00Z"}`)
	body := scrapeMetrics(t, Deps{})
	if !strings.Contains(body, `audit_id="audit-new"`) || strings.Contains(body, `audit_id="audit-old"`) {
		t.Fatalf("mauvais rapport sélectionné :\n%s", body)
	}
	// Un corps JSON illisible dans un autre dossier ne doit pas masquer le bon.
	writeSealedReport(t, dir, "audit-broke", `{`)
	body = scrapeMetrics(t, Deps{})
	if !strings.Contains(body, `audit_id="audit-new"`) {
		t.Fatalf("artefact illisible a masqué le rapport valide :\n%s", body)
	}
}

// TestMetrics_JSONShapeStillDecodes — garde-fou : la charge /metrics reste un
// corps Prometheus, pas un JSON de rapport potentialisé.
func TestMetrics_JSONShapeStillDecodes(t *testing.T) {
	body := scrapeMetrics(t, Deps{})
	var doc map[string]any
	if err := json.Unmarshal([]byte(body), &doc); err == nil {
		t.Fatalf("/metrics doit rester du texte Prometheus, pas un JSON :\n%s", body)
	}
}
