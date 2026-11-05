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
	// « .. » seul : Base("..") vaut "..", il écrirait audits/../telemetry.csv
	// (hors de la zone d'audit) — refus au même titre qu'un séparateur.
	if _, err := WriteTelemetry(t.TempDir(), &Result{AuditID: ".."}); err == nil {
		t.Fatal("id .. : erreur attendue")
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

func TestReadTelemetryStrict(t *testing.T) {
	dir := t.TempDir()
	res := &Result{AuditID: "audit-read", Telemetry: []TelemetryRow{
		{Phase: "baseline", OffsetMs: 0, RTTms: 20.5},
		{Phase: "baseline", OffsetMs: 400, RTTms: 21.25, SmallMs: 25},
		{Phase: "load", OffsetMs: 800, RTTms: 30, BulkBytes: 1234},
	}}
	if _, err := WriteTelemetry(dir, res); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadTelemetry(dir, "audit-read")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[1].RTTms != 21.25 || rows[1].SmallMs != 25 || rows[2].BulkBytes != 1234 {
		t.Fatalf("lignes relues : %+v", rows)
	}
}

func TestReadTelemetryRejectsUnknownHeader(t *testing.T) {
	dir := t.TempDir()
	auditDir := filepath.Join(dir, "audits", "audit-bad")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "phase,offset,rtt_ms,small_ms,bulk_bytes\nbaseline,0,20,0,\n"
	if err := os.WriteFile(filepath.Join(auditDir, "telemetry.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTelemetry(dir, "audit-bad"); err == nil {
		t.Fatal("en-tête inconnu accepté — les colonnes glisseraient en silence")
	}
}

func TestReadTelemetryRejectsEmptyProbe(t *testing.T) {
	dir := t.TempDir()
	auditDir := filepath.Join(dir, "audits", "audit-empty")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "phase,offset_ms,rtt_ms,small_ms,bulk_bytes\nbaseline,0,0,0,\n"
	if err := os.WriteFile(filepath.Join(auditDir, "telemetry.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTelemetry(dir, "audit-empty"); err == nil {
		t.Fatal("ligne sans mesure acceptée — un zéro deviendrait une mesure")
	}
}
