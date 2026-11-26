package campagne

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
	"github.com/Realms4239/cgo/model"
)

// TestEffectiveDeadlineIsProfileRelative — l'échéance vient des vraies valeurs
// P1-P4 (délai + sérialisation 16 Ko à la capacité du profil + marge), jamais
// du seuil fixe 220 ms partagé entre tous les profils.
func TestEffectiveDeadlineIsProfileRelative(t *testing.T) {
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		prof := model.Profiles[id]
		want, ok := effectiveDeadline(prof, "up", 50)
		if !ok {
			t.Fatalf("profil %s : échéance refusée à tort", id)
		}
		reference, err := metrics.EffectiveDeadlineMs(prof.DelayMs, deadlineObjectBytes, prof.CapUp(), 50)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(want-reference) > 1e-9 || want == metrics.CAPDeadlineMs {
			t.Fatalf("profil %s : échéance = %v, référence = %v (pas un seuil plat 220 ms)", id, want, reference)
		}
		d := fastDeps()
		d.DeadlineMs = 50
		ev := model.Event{RunID: "r1", EventID: 1, Profile: id, Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
		got, err := RunEvent(context.Background(), ev, prof, d)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.DeadlineDueMs-want) > 1e-9 || got.DeadlineBasis != "profile-relative" {
			t.Fatalf("profil %s : due=%v basis=%q, voulu due=%v basis=profile-relative", id, got.DeadlineDueMs, got.DeadlineBasis, want)
		}
		if !got.DeadlineMeasured {
			t.Fatalf("profil %s : complétions présentes mais measured=false", id)
		}
	}
}

// TestEffectiveDeadlineInvalidRateFailsClosed — capacité nulle/négative/NaN :
// aucun nombre fabriqué, le marqueur d'absence reste faux.
func TestEffectiveDeadlineInvalidRateFailsClosed(t *testing.T) {
	for _, rate := range []float64{0, -5, math.NaN(), math.Inf(1)} {
		prof := model.Profile{ID: "BAD", CapacityMbps: rate, DelayMs: 10}
		due, ok := effectiveDeadline(prof, "up", 50)
		if ok || due != 0 {
			t.Fatalf("rate=%v : due=%v ok=%v, attendu (0,false)", rate, due, ok)
		}
	}
}

// TestEmptyCompletionsMarkedAbsent — la source non gardée (campagne.go:649)
// transformait NaN en -9.2e17 via int(NaN). Le drapeau porte l'absence, le JSON
// ne contient jamais NaN.
func TestEmptyCompletionsMarkedAbsent(t *testing.T) {
	d := fastDeps()
	d.DeadlineMs = 50
	d.Small = func(context.Context) (float64, error) { return 0, errors.New("sonde indisponible") }
	ev := model.Event{RunID: "r1", EventID: 1, Profile: "P3", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P3"], d)
	if err != nil {
		t.Fatal(err)
	}
	if got.DeadlineMeasured || got.DeadlineOKPct != 0 {
		t.Fatalf("absence mal encodée : measured=%v pct=%v", got.DeadlineMeasured, got.DeadlineOKPct)
	}
	if got.DeadlineDueMs <= 0 {
		t.Fatalf("provenance absente : due=%v", got.DeadlineDueMs)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("JSON d'événement non sérialisable (NaN) : %v", err)
	}
	if strings.Contains(string(raw), "NaN") || !strings.Contains(string(raw), `"deadline_measured":false`) {
		t.Fatalf("JSON d'absence incorrect : %s", raw)
	}
}

// TestWriterDeadlineAbsenceIsEmptyCell — en-tête 23 colonnes byte pour byte,
// cellule vide pour une absence, format %.1f inchangé pour une mesure.
func TestWriterDeadlineAbsenceIsEmptyCell(t *testing.T) {
	dir := t.TempDir()
	w, err := OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	measured := eventForCampagne(1, model.GatePass)
	measured.DeadlineMeasured = true
	absent := eventForCampagne(2, model.GatePass)
	absent.DeadlineMeasured = false
	absent.DeadlineOKPct = 0
	for _, ev := range []model.Event{measured, absent} {
		if err := w.Append(ev); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.fh.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "aqm_eval.csv"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(b))
	if lines[0] != strings.Join(model.AQMEvalHeader, ",") {
		t.Fatalf("en-tête gelé modifié : %s", lines[0])
	}
	cells := splitCSV(lines[1])
	if len(cells) != 23 || cells[13] != "99.0" {
		t.Fatalf("cellule mesurée : %d colonnes, deadline=%q (attendu 99.0)", len(cells), cells[13])
	}
	cells = splitCSV(lines[2])
	if len(cells) != 23 || cells[13] != "" {
		t.Fatalf("cellule absente : %d colonnes, deadline=%q (attendu vide)", len(cells), cells[13])
	}
}

// TestMatrixConfigSealsDeadlineBasisAndDue — le manifeste scellé porte le basis
// et l'échéance effective par profil ; un taux invalide n'écrit aucune valeur.
func TestMatrixConfigSealsDeadlineBasisAndDue(t *testing.T) {
	raw, err := buildMatrixConfig([]string{"P1", "P2", "P3", "P4"}, []string{"cake"}, []string{"bbr"}, 1, Deps{DeadlineMs: 50})
	if err != nil {
		t.Fatal(err)
	}
	var cfg matrixConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.DeadlineBasis != "profile-relative" {
		t.Fatalf("deadline_basis = %q", cfg.DeadlineBasis)
	}
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		prof := model.Profiles[id]
		want, _ := effectiveDeadline(prof, "up", 50)
		if got := cfg.DeadlineDueMs[id]; math.Abs(got-want) > 1e-9 {
			t.Fatalf("profil %s : due gelé = %v, voulu %v", id, got, want)
		}
	}
}
