package campagne

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// eventForCampagne construit une cellule distincte pour le Freeze/resume.
func eventForCampagne(repetition int, gate string) model.Event {
	return model.Event{
		RunID: "run-durabilite", EventID: repetition, Profile: "P2", Qdisc: model.Cake,
		CC: model.BBR, Direction: "up", Repetition: repetition,
		RTTp50Ms: 20, RTTp95Ms: 24, QDIPctMs: 5, Smallp95Ms: 26, DeadlineOKPct: 99,
		BulkGoodputMbps: 79.5, CostARPerH: 0, CPUPct: 31,
		RTTBaseP50Ms: 18, RTTBaseP95Ms: 21,
		GateStatus: gate,
	}
}

// TestWriterResumeAfterUncleanClose — arrêt brutal simulé (close SANS
// Freeze, analogue d'un kill -9 : le CSV reste, jamais de manifeste) : la
// réouverture reconstruit seen depuis le fichier et G5 refuse de ré-ajouter
// une cellule déjà gelée. C'est le chemin de reprise après arrêt PC/VM en
// pleine matrice.
func TestWriterResumeAfterUncleanClose(t *testing.T) {
	dir := t.TempDir()
	ev := eventForCampagne(1, model.GatePass)

	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	// close SANS Freeze : comme si le process était mort entre l'append et
	// le gel. Le handle est fermé proprement, le gel n'a jamais eu lieu.
	if err := w.fh.Close(); err != nil {
		t.Fatal(err)
	}

	w2, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w2.fh.Close()
	if err := w2.Append(ev); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("G5 muet à la reprise après fermeture brutale : %v", err)
	}
}

// TestWriterResumeRepairTornTail (b2) — la reprise retire le fragment de
// queue AVANT toute relecture : le fichier retombe sur « header\n », la
// cellule déchirée se ré-appende, et le CSV final porte exactement 2 lignes
// non blanches, la dernière en 23 cellules commençant par le run id de la
// cellule déchirée.
func TestWriterResumeRepairTornTail(t *testing.T) {
	dir := t.TempDir()
	ev := eventForCampagne(1, model.GatePass)

	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	if err := w.fh.Close(); err != nil {
		t.Fatal(err)
	}

	// déchirer la queue SANS retour à la ligne : dernier octet ≠ '\n'
	path := filepath.Join(dir, "aqm_eval.csv")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	trimmed := bytes.TrimRight(b, "\n")
	nl := bytes.LastIndexByte(trimmed, '\n')
	row := trimmed[nl+1:]
	torn := trimmed[:nl+1+len(row)/2]
	if err := os.WriteFile(path, torn, 0644); err != nil {
		t.Fatal(err)
	}

	w2, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w2.fh.Close()
	if err := w2.Append(ev); err != nil {
		t.Fatalf("la cellule déchirée doit se ré-ajouter : %v", err)
	}
	final, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(final))
	nonEmpty := 0
	var last string
	for _, ln := range lines {
		if ln == "" {
			continue
		}
		nonEmpty++
		last = ln
	}
	if nonEmpty != 2 {
		t.Fatalf("splitLines = %d lignes non blanches, want 2 (en-tête + la cellule)", nonEmpty)
	}
	parts := splitCSV(last)
	if len(parts) != 23 {
		t.Fatalf("la dernière ligne porte %d cellules, want 23", len(parts))
	}
	if parts[0] != ev.RunID {
		t.Fatalf("la dernière ligne commence par %q, want %q", parts[0], ev.RunID)
	}
}

// TestWriterResumeIgnoresTornTailRow — une ligne coupée en deux par l'arrêt
// (largeur < 23 cellules, pas de retour à la ligne) ne pollue PAS seen :
// la même cellule se ré-appende proprement. Avant le garde-fou, la moitié de
// ligne était lue comme une cellule gelée et G5 la sautait à tort.
func TestWriterResumeIgnoresTornTailRow(t *testing.T) {
	dir := t.TempDir()
	ev := eventForCampagne(1, model.GatePass)

	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(ev); err != nil {
		t.Fatal(err)
	}
	if err := w.fh.Close(); err != nil {
		t.Fatal(err)
	}

	// déchirer la dernière ligne : garder seulement la moitié de son cœur
	path := filepath.Join(dir, "aqm_eval.csv")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	trimmed := bytes.TrimRight(b, "\n")
	nl := bytes.LastIndexByte(trimmed, '\n')
	row := trimmed[nl+1:]
	torn := trimmed[:nl+1+len(row)/2]
	if parts := splitCSV(string(torn[nl+1:])); len(parts) >= 23 {
		t.Fatalf("la déchirure simulée porte encore %d cellules — test invalide", len(parts))
	}
	if err := os.WriteFile(path, torn, 0644); err != nil {
		t.Fatal(err)
	}

	w2, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w2.fh.Close()
	if err := w2.Append(ev); err != nil {
		t.Fatalf("la cellule doit se ré-ajouter après une ligne déchirée : %v", err)
	}
}

// TestFreezeManifestListsAndHashes — le manifest parse, chaque fichier
// listé hash juste (aqm_eval.csv inclus), et quarantine.json y figure dès
// qu'un événement est mis en quarantaine (une rangée invalid suffit).
func TestFreezeManifestListsAndHashes(t *testing.T) {
	dir := t.TempDir()
	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Append(eventForCampagne(1, model.GatePass)); err != nil {
		t.Fatal(err)
	}
	if err := w.Append(eventForCampagne(2, model.GateInvalid)); err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`{"profiles":["P2"],"reps":2}`)
	if err := w.Freeze(cfg); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal("manifest.json absent")
	}
	var doc struct {
		Files []struct {
			File   string `json:"File"`
			Sha256 string `json:"Sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("manifest illisible : %v", err)
	}
	seen := map[string]bool{}
	for _, f := range doc.Files {
		fb, err := os.ReadFile(filepath.Join(dir, f.File))
		if err != nil {
			t.Fatalf("fichier listé illisible : %s : %v", f.File, err)
		}
		sum := sha256.Sum256(fb)
		if got := hex.EncodeToString(sum[:]); got != f.Sha256 {
			t.Fatalf("hash de %s : %s, manifeste dit %s", f.File, got, f.Sha256)
		}
		seen[f.File] = true
	}
	if !seen["aqm_eval.csv"] {
		t.Fatal("aqm_eval.csv absent du manifeste")
	}
	if !seen["quarantine.json"] {
		t.Fatal("quarantine.json absent du manifeste alors qu'un événement était en quarantaine")
	}
	b, err = os.ReadFile(filepath.Join(dir, "quarantine.json"))
	if err != nil {
		t.Fatal("quarantine.json absent sur le disque")
	}
	var quarantine []map[string]any
	if err := json.Unmarshal(b, &quarantine); err != nil {
		t.Fatalf("quarantine.json illisible : %v", err)
	}
	if len(quarantine) != 1 {
		t.Fatalf("quarantine : %d entrées, want 1", len(quarantine))
	}
}
