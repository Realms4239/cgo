package api

import (
	"math"
	"strings"
	"testing"
)

// nanFrame — struct locale : la production publie des structs (Snapshot,
// liveFrame), pas seulement des maps. Un NaN dans un champ struct ne doit pas
// faire tomber la frame entière.
type nanFrame struct {
	Phase            string  `json:"phase"`
	DeadlineOKPct    float64 `json:"deadline_ok_pct"`
	DeadlineMeasured bool    `json:"deadline_measured"`
}

// TestSSEJSONNeverCarriesNaN — un NaN entré dans un instantané fait échouer
// json.Marshal et ferait tomber toute la frame SSE (silencieusement). La couche
// api doit garantir un JSON sans NaN ; l'absence est portée par le booléen.
func TestSSEJSONNeverCarriesNaN(t *testing.T) {
	h := NewHub()
	h.Publish(map[string]any{
		"phase":             "charge",
		"deadline_ok_pct":   math.NaN(),
		"deadline_measured": false,
	})
	if h.lastFull == "" {
		t.Fatal("aucune frame publiée : la frame entière est tombée sur un NaN")
	}
	if strings.Contains(h.lastFull, "NaN") || strings.Contains(h.lastFull, "Inf") {
		t.Fatalf("JSON SSE avec une valeur non finie : %s", h.lastFull)
	}
	if strings.Contains(h.lastFull, `"deadline_ok_pct"`) {
		t.Fatalf("clé numérique publiée pour une absence : %s", h.lastFull)
	}
	if !strings.Contains(h.lastFull, `"deadline_measured":false`) {
		t.Fatalf("drapeau d'absence manquant : %s", h.lastFull)
	}
}

// TestSSEAbsenceDropsDeadlinePct — un 0 déclaré absent n'est pas un nombre
// publié : la clé disparaît, seul le booléen porte l'absence.
func TestSSEAbsenceDropsDeadlinePct(t *testing.T) {
	h := NewHub()
	h.Publish(map[string]any{"deadline_ok_pct": 0.0, "deadline_measured": false})
	if h.lastFull == "" {
		t.Fatal("aucune frame publiée")
	}
	if strings.Contains(h.lastFull, `"deadline_ok_pct"`) {
		t.Fatalf("0 publié pour une absence : %s", h.lastFull)
	}
	if !strings.Contains(h.lastFull, `"deadline_measured":false`) {
		t.Fatalf("drapeau d'absence manquant : %s", h.lastFull)
	}
}

// TestSSEMeasuredZeroKeepsDeadlinePct — un vrai 0 % mesuré reste publié :
// le nettoyage ne doit pas confondre mesure et absence.
func TestSSEMeasuredZeroKeepsDeadlinePct(t *testing.T) {
	h := NewHub()
	h.Publish(map[string]any{"deadline_ok_pct": 0.0, "deadline_measured": true})
	if !strings.Contains(h.lastFull, `"deadline_ok_pct":0`) {
		t.Fatalf("0 mesuré supprimé à tort : %s", h.lastFull)
	}
}

// TestSSENaNStructStillPublishes — un struct avec un NaN doit produire une
// frame ; le nettoyage ne peut pas se limiter aux maps d'entrée.
func TestSSENaNStructStillPublishes(t *testing.T) {
	h := NewHub()
	h.Publish(nanFrame{Phase: "charge", DeadlineOKPct: math.NaN(), DeadlineMeasured: true})
	if h.lastFull == "" {
		t.Fatal("aucune frame publiée : un NaN dans un struct a fait tomber la frame")
	}
	if strings.Contains(h.lastFull, "NaN") || strings.Contains(h.lastFull, "Inf") {
		t.Fatalf("JSON SSE avec une valeur non finie : %s", h.lastFull)
	}
	if !strings.Contains(h.lastFull, `"phase":"charge"`) {
		t.Fatalf("struct mal sérialisé : %s", h.lastFull)
	}
}

// TestSSENestedNaNStillPublishes — le nettoyage est récursif : un NaN dans une
// map imbriquée ne doit pas faire échouer le marshalling de la frame.
func TestSSENestedNaNStillPublishes(t *testing.T) {
	h := NewHub()
	h.Publish(map[string]any{
		"phase":  "charge",
		"nested": map[string]any{"pct": math.NaN()},
	})
	if h.lastFull == "" {
		t.Fatal("aucune frame publiée : un NaN imbriqué a fait tomber la frame")
	}
	if strings.Contains(h.lastFull, "NaN") || strings.Contains(h.lastFull, "Inf") {
		t.Fatalf("JSON SSE avec une valeur non finie : %s", h.lastFull)
	}
}
