package api

import (
	"math"
	"strings"
	"testing"
)

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
