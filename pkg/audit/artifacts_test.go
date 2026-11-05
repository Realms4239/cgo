package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
)

func TestArtifactPathsAreTheOneConvention(t *testing.T) {
	dir, err := ArtifactDir("data", "audit-7")
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join("data", "audits", "audit-7") {
		t.Errorf("dossier : %q", dir)
	}
	file, err := ArtifactFile("data", "audit-7", ReportArtifactName)
	if err != nil {
		t.Fatal(err)
	}
	if file != filepath.Join(dir, "audit_report.json") {
		t.Errorf("fichier du rapport : %q, attendu sous %q", file, dir)
	}
	for _, bad := range []string{"", " ", "../evade", "a/b", ".."} {
		if _, err := ArtifactDir("data", bad); err == nil {
			t.Errorf("identifiant %q accepté : la porte E11 est ouverte ici", bad)
		}
	}
}

// TestArtifactPointerIsPortableAndRefusesBadIDs — le chemin scellé se relit de
// la même façon quel que soit l'hôte, et un identifiant refusé rend une absence.
func TestArtifactPointerIsPortableAndRefusesBadIDs(t *testing.T) {
	if got := ArtifactPointer("audit-7"); got != "data/audits/audit-7/audit_report.json" {
		t.Errorf("pointeur : %q", got)
	}
	for _, bad := range []string{"", " ", "../evade", "a/b", "..", "/absolu"} {
		if got := ArtifactPointer(bad); got != "" {
			t.Errorf("pointeur pour %q : %q, attendu vide", bad, got)
		}
	}
}

func TestSealedReportPointsAtItsOwnArtifact(t *testing.T) {
	res := &Result{
		AuditID: "audit-pointeur", Timestamp: "2026-11-02T10:00:00Z",
		Site: "tana", LinkType: "4g", RTTIdleP95: 49, RTTLoadedP95: 240,
		LoadAttempted: true, Invocation: invocationFrom(Params{Duration: 60, Flows: 4, Direction: "up", Site: "tana", Target: "8.8.8.8"}),
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Artifact != "data/audits/audit-pointeur/audit_report.json" {
		t.Errorf("artifact scellé : %q", rep.Artifact)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "data/audits/audit-pointeur/") {
		t.Errorf("markdown sans chemin de preuve : %s", md)
	}
	// Le dossier et le fichier sont la même vérité : la console n'a pas à
	// reconstruire un chemin à partir de l'identifiant.
	if !strings.Contains(md, "data/audits/audit-pointeur") || strings.Contains(md, `data\audits`) {
		t.Errorf("chemin non portable ou incohérent dans le markdown : %s", md)
	}
}

func TestArtifactAbsenceStaysAnAbsence(t *testing.T) {
	res := &Result{
		AuditID: "../evade", Timestamp: "2026-11-02T10:00:00Z",
		Site: "tana", LinkType: "4g", LoadAttempted: true,
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Artifact != "" {
		t.Fatalf("un identifiant refusé ne doit pas produire de chemin : %q", rep.Artifact)
	}
	if md := rep.RenderMarkdown(); !strings.Contains(md, "Preuve: non enregistrée") {
		t.Errorf("markdown sans absence nommée : %s", md)
	}
}

// TestSealedPointerIsTheFileTheWriterWrites — la couture qui compte : le pointeur
// scellé dans le rapport est LE fichier que WriteArtifacts produit, et le rapport
// relu depuis ce disque porte le même pointeur. Sans elle, la console pourrait
// afficher un chemin que l'écriture ne produit pas — et le prouverait quand même
// en ouvrant un 404.
func TestSealedPointerIsTheFileTheWriterWrites(t *testing.T) {
	dir := t.TempDir()
	res := &Result{
		AuditID: "audit-couture", Timestamp: "2026-11-02T10:00:00Z",
		Site: "tana", LinkType: "4g", RTTIdleP95: 49, LoadAttempted: true,
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	jsonPath, mdPath, err := WriteArtifacts(rep, dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "audits", "audit-couture", ReportArtifactName); jsonPath != want {
		t.Errorf("rapport écrit à %q, attendu %q", jsonPath, want)
	}
	if want := filepath.Join(dir, "audits", "audit-couture", ReportMarkdownName); mdPath != want {
		t.Errorf("markdown écrit à %q, attendu %q", mdPath, want)
	}
	if rep.Artifact != ArtifactPointer("audit-couture") {
		t.Errorf("pointeur scellé instable : %q", rep.Artifact)
	}
	b, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var reread Report
	if err := json.Unmarshal(b, &reread); err != nil {
		t.Fatal(err)
	}
	if reread.Artifact != rep.Artifact {
		t.Errorf("le rapport relu ne pointe pas au même endroit : %q vs %q", reread.Artifact, rep.Artifact)
	}
	// Et la télémétrie de ce même audit se relit par la même convention.
	if _, err := ArtifactFile(dir, reread.AuditID, TelemetryArtifactName); err != nil {
		t.Errorf("chemin télémétrie refusé pour un identifiant accepté : %v", err)
	}
}
