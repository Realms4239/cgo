package results

import (
	"strings"
	"testing"
)

// La règle est la même partout (HTTP, CLI, banc) — ses tests hostiles aussi.

func TestValidRunID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"", false},
		{"run-1700000000", true},
		{"abc", true},
		{"run-x_1-2.3", false}, // point
		{"run/evil", false},    // séparateur unix
		{`run\evil`, false},    // séparateur Windows
		{"..", false},          // traversal pur
		{".", false},           // point
		{".hidden", false},     // point en tête
		{"a..b", false},        // points internes
		{"run%20", true},       // percent-encodé en clair : pas de séparateur
		{"run\x00evil", false}, // NUL
		{"run\nligne", false},  // \n
		{"run\tcol", false},    // \t
		{"café-2024", true},    // unicode, pas de séparateur
		{strings.Repeat("a", 128), true},
		{strings.Repeat("a", 129), false},
	}
	for _, c := range cases {
		if got := ValidRunID(c.id); got != c.valid {
			t.Errorf("ValidRunID(%q) = %v, veut %v", c.id, got, c.valid)
		}
	}
}

func TestValidateRunIDErrorNamesTheRule(t *testing.T) {
	for _, id := range []string{"", ".", "..", "a/b", "a\\b", "a.b"} {
		if err := ValidateRunID(id); err == nil {
			t.Errorf("ValidateRunID(%q) : aucune erreur rendue", id)
		} else if !strings.Contains(err.Error(), "invalide") {
			t.Errorf("ValidateRunID(%q) : message %q doit nommer la règle", id, err)
		}
	}
}
