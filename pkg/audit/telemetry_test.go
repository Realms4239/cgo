package audit

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTelemetry(t *testing.T) {
	dir := t.TempDir()
	res := &Result{AuditID: "audit-t", Telemetry: []TelemetryRow{
		{Phase: "baseline", OffsetMs: 0, RTTms: 20.5},
		{Phase: "load", OffsetMs: 10000, RTTms: 140.2, SmallMs: 180.0},
		{Phase: "drain", OffsetMs: 50000, RTTms: 30.1},
	}}
	path, err := WriteTelemetry(dir, res)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	r := csv.NewReader(strings.NewReader(string(raw)))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][0] != "phase" || len(rows) != 4 {
		t.Fatalf("en-tête/lignes : %v", rows)
	}
	if rows[2][1] != "10000" || rows[2][4] != "" { // offset + bulk vide pour la sonde
		t.Fatalf("ligne load : %v", rows[2])
	}
}

func TestWriteTelemetryRejectsEmptyID(t *testing.T) {
	// un id vide écrirait data/audits//telemetry.csv — refus propre, pas
	// une écriture hors zone.
	if _, err := WriteTelemetry(t.TempDir(), &Result{AuditID: ""}); err == nil {
		t.Fatal("id vide : erreur attendue")
	}
	if _, err := WriteTelemetry(t.TempDir(), &Result{AuditID: "../escape"}); err == nil {
		t.Fatal("id traversant : erreur attendue")
	}
}

func TestSealTelemetryAndDualSeals(t *testing.T) {
	// Deux sceaux distincts : la série brute (telemetry.csv) et la ligne
	// CSV gelée — le rapport prouve les deux maillons de la chaîne.
	dir := t.TempDir()
	telPath := filepath.Join(dir, "telemetry.csv")
	os.WriteFile(telPath, []byte("phase,offset_ms\nbaseline,0\n"), 0644)
	rowPath := filepath.Join(dir, "link_audit.csv")
	os.WriteFile(rowPath, []byte("audit_id\nx\n"), 0644)
	rep := &Report{AuditID: "a"}
	if err := rep.SealTelemetry(telPath); err != nil {
		t.Fatal(err)
	}
	if err := rep.SealSource(rowPath); err != nil {
		t.Fatal(err)
	}
	if rep.Provenance.TelemetrySHA256 == "" || rep.Provenance.TelemetrySHA256 == rep.Provenance.SourceSHA256 {
		t.Fatalf("sceaux confondus : tel=%s row=%s", rep.Provenance.TelemetrySHA256, rep.Provenance.SourceSHA256)
	}
	if rep.Provenance.RowFileSHA256 != rep.Provenance.SourceSHA256 {
		t.Fatal("RowFileSHA256 doit sceller le fichier de lignes (alias explicite)")
	}
	// le rendu Markdown cite le sceau télémétrie — preuve visible.
	md := rep.RenderMarkdown()
	if !strings.Contains(md, rep.Provenance.TelemetrySHA256[:16]) {
		t.Fatal("sceau télémétrie absent du rendu")
	}
}
