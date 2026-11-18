package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/audit"
)

// TestAuditHTTPSealsTelemetryAndSource — le chemin HTTP scelle les DEUX
// preuves comme la CLI (cmd/cgo/main.go) : la série brute (telemetry_sha256)
// et la ligne CSV ajoutée par cet audit (sceau source). L'audit tourne de
// bout en bout avec des dépendances simulées ; le rapport scellé est relu
// depuis le disque. CGO_DATA_DIR/chdir épinglent la racine et écartent tout
// ./data réel du poste.
func TestAuditHTTPSealsTelemetryAndSource(t *testing.T) {
	withDevNoAuth(t)
	dir := chdirTemp(t)
	t.Setenv("CGO_DATA_DIR", dir)
	t.Setenv("CGO_BULK_ADDR", "")
	t.Setenv("CGO_SMALL_URL", "")
	site := "seal-" + time.Now().Format("150405.000000")
	h := New(Deps{
		SinksFn: func() audit.Sinks {
			return audit.Sinks{Mode: "external", Target: "127.0.0.1"}
		},
		AuditDepsFn: func() audit.Deps {
			return audit.Deps{
				Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{42, 43, 44, 45, 46} },
				Small: func(context.Context) (float64, error) { return 12.5, nil },
				BulkLoad: func(_ context.Context, _ string, _ string, _ int) audit.LoadOutcome {
					return audit.LoadOutcome{Bytes: 1_000_000, PerFlow: []uint64{1_000_000}, Elapsed: time.Second}
				},
			}
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	body, _ := json.Marshal(map[string]any{
		"site": site, "link_type": "fiber", "duration": 10, "direction": "up", "flows": 4,
	})
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start : %d", resp.StatusCode)
	}
	waitAuditIdle(t, srv)

	// le dossier d'artefact de CET audit est le seul enfant de audits/
	auditRoot := filepath.Join(dir, "audits")
	entries, err := os.ReadDir(auditRoot)
	if err != nil {
		t.Fatalf("dossier audits absent alors que l'audit a terminé : %v", err)
	}
	found := ""
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "audit-") {
			found = e.Name()
		}
	}
	if found == "" {
		t.Fatalf("aucun artefact pour %s sous %s (entrées : %v)", site, auditRoot, entries)
	}

	reportPath := filepath.Join(auditRoot, found, "audit_report.json")
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("audit_report.json illisible : %v", err)
	}
	var sealed struct {
		Provenance map[string]any `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &sealed); err != nil {
		t.Fatalf("audit_report.json ne se parse pas : %v", err)
	}
	if sealed.Provenance == nil {
		t.Fatal("rapport scellé sans provenance")
	}
	tel, _ := sealed.Provenance["telemetry_sha256"].(string)
	if tel == "" {
		t.Fatal("provenance.telemetry_sha256 vide — la série brute n'est pas scellée (HTTP)")
	}
	// le sceau de la ligne CSV porte le nom source_sha256 sur le fil
	// (et row_file_sha256 comme alias explicite depuis le même sceau) ;
	// row_file_or_source_sha256 n'existe pas dans le schéma publié.
	src, _ := sealed.Provenance["source_sha256"].(string)
	row, _ := sealed.Provenance["row_file_sha256"].(string)
	if src == "" && row == "" {
		t.Fatal("ni provenance.source_sha256 ni provenance.row_file_sha256 — la ligne CSV n'est pas scellée (HTTP)")
	}

	// les fichiers eux-mêmes existent sous la racine épinglée
	for _, want := range []string{
		filepath.Join(auditRoot, found, "telemetry.csv"),
		filepath.Join(dir, "link_audit.csv"),
	} {
		st, err := os.Stat(want)
		if err != nil {
			t.Fatalf("fichier attendu absent : %s : %v", want, err)
		}
		if st.Size() == 0 {
			t.Fatalf("fichier attendu vide : %s", want)
		}
	}
}

// GET /api/audit/report sans identifiant — une archive vide répond 200 +
// available:false (maison), pas un 404 qui pollue la console du dashboard.
func TestHandleAuditReportEmptyArchiveReturns200AvailableFalse(t *testing.T) {
	withDevNoAuth(t)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/audit/report")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	buf, _ := io.ReadAll(resp.Body)
	var body map[string]any
	if err := json.Unmarshal(buf, &body); err != nil {
		t.Fatal(err)
	}
	if v, ok := body["available"]; !ok || v != false {
		t.Fatalf("available must be false, got %v (body=%s)", v, buf)
	}
	if _, ok := body["reason"]; !ok {
		t.Fatalf("reason missing: %s", buf)
	}
}
