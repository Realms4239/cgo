package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// sourceOf lit le champ via le contrat JSON : le test compile avant l'ajout du
// champ Source et échoue avec un message explicite (pas une erreur de build).
func sourceOf(t *testing.T, id string) string {
	t.Helper()
	b, err := json.Marshal(Profiles[id])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	s, _ := m["source"].(string)
	return s
}

// TestProfileSourceField — champ optionnel, ignoré par les anciennes lignes CSV.
func TestProfileSourceField(t *testing.T) {
	f, ok := reflect.TypeOf(Profile{}).FieldByName("Source")
	if !ok {
		t.Fatal("Profile sans champ Source")
	}
	if f.Type.Kind() != reflect.String {
		t.Fatalf("Source: type %v, want string", f.Type)
	}
	if got := f.Tag.Get("json"); got != "source,omitempty" {
		t.Fatalf("Source json tag = %q, want %q", got, "source,omitempty")
	}
	// omitempty : un profil historique (sans source) ne sérialise pas la clé.
	b, _ := json.Marshal(Profile{ID: "OLD", CapacityMbps: 5})
	if strings.Contains(string(b), `"source"`) {
		t.Fatalf("profil sans source sérialise la clé: %s", b)
	}
}

// TestCitedProfilesExist — cibles d'émulation L1/V1 (littérature citée, jamais
// des mesures de terrain). Valeurs exactes de la spec §1.
func TestCitedProfilesExist(t *testing.T) {
	want := map[string]Profile{
		"L1": {ID: "L1", CapacityMbps: 200, CapacityUpMbps: 30, DelayMs: 26, JitterMs: 15, LossPct: 0.1},
		"V1": {ID: "V1", CapacityMbps: 10, CapacityUpMbps: 2, DelayMs: 600, JitterMs: 30, LossPct: 1},
	}
	for id, exp := range want {
		got, ok := Profiles[id]
		if !ok {
			t.Fatalf("profil cité %s absent de Profiles", id)
		}
		if got.ID != exp.ID || got.CapacityMbps != exp.CapacityMbps ||
			got.CapacityUpMbps != exp.CapacityUpMbps || got.DelayMs != exp.DelayMs ||
			got.JitterMs != exp.JitterMs || got.LossPct != exp.LossPct {
			t.Fatalf("%s: got %+v, want %+v", id, got, exp)
		}
	}
}

// TestCitedProfilesHaveSource — chaque cible porte sa provenance et ses
// ancrages bibliographiques ; absent = oracle hostile qui échoue fermé.
func TestCitedProfilesHaveSource(t *testing.T) {
	for id, markers := range map[string][]string{
		"L1": {"starlink.com", "IIUM", "ITU"},
		"V1": {"IIUM", "IJSRA", "ITU"},
	} {
		s := sourceOf(t, id)
		if strings.TrimSpace(s) == "" {
			t.Fatalf("%s: source vide (cible d'émulation non citée)", id)
		}
		for _, m := range markers {
			if !strings.Contains(s, m) {
				t.Fatalf("%s: source %q sans ancrage %q", id, s, m)
			}
		}
	}
}

// TestCitedProfilesBuiltin — natifs du binaire, donc jamais marqués importés.
func TestCitedProfilesBuiltin(t *testing.T) {
	for _, id := range []string{"L1", "V1"} {
		if !BuiltinProfileIDs[id] {
			t.Fatalf("%s absent de BuiltinProfileIDs", id)
		}
	}
}

// TestLegacyProfilesUnchanged — P1-P4 figés : le corpus gelé et les oracles
// E2E existants dépendent de ces valeurs au bit près.
func TestLegacyProfilesUnchanged(t *testing.T) {
	want := map[string]Profile{
		"P1": {ID: "P1", CapacityMbps: 80, DelayMs: 20, JitterMs: 2, LossPct: 0},
		"P2": {ID: "P2", CapacityMbps: 20, DelayMs: 100, JitterMs: 15, LossPct: 0.5},
		"P3": {ID: "P3", CapacityMbps: 5, DelayMs: 600, JitterMs: 30, LossPct: 1},
		"P4": {ID: "P4", CapacityMbps: 100, DelayMs: 40, JitterMs: 20, LossPct: 0.3},
	}
	for id, exp := range want {
		got, ok := Profiles[id]
		if !ok {
			t.Fatalf("profil figé %s absent", id)
		}
		if got != exp {
			t.Fatalf("%s modifié: got %+v, want %+v", id, got, exp)
		}
	}
}
