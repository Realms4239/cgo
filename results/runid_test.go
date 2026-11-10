package results

import (
	"errors"
	"strings"
	"testing"
)

func TestValidRunID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"", false},
		{"run-1700000000", true},
		{"abc", true},
		{"café-2024", true},            // unicode other than controls
		{"run%20", true},              // % is not a separator
		{"run-x_1-2.3", false},        // point
		{"run/evil", false},           // séparateur unix
		{`run\evil`, false},           // séparateur Windows
		{"..", false},                 // traversal pur
		{".", false},                  // point
		{".hidden", false},            // point en tête
		{"a..b", false},               // points internes
		{"run\x00evil", false},        // NUL (C0)
		{"run\nligne", false},         // LF (C0)
		{"run\tcol", false},           // TAB (C0)
		{"run\x7fevil", false},        // DEL — 0x7f
		{"run\u0085evil", false},      // C1 — 0x85
		{strings.Repeat("a", 128), true},
		{strings.Repeat("a", 129), false},
		{strings.Repeat("é", 64), true},   // 2 octets chacune = 128 exactement
		{strings.Repeat("é", 65), false},  // 130 octets
	}
	for _, c := range cases {
		if got := ValidRunID(c.id); got != c.valid {
			t.Errorf("ValidRunID(%q) = %v, veut %v", c.id, got, c.valid)
		}
	}
}

func TestValidateRunIDErrorIsTyped(t *testing.T) {
	for _, id := range []string{"", ".", "..", "a/b", "a\\b", "a.b", "\x00", "\x7f"} {
		if err := ValidateRunID(id); !errors.Is(err, ErrInvalidRunID) {
			t.Errorf("ValidateRunID(%q) = %v, veut ErrInvalidRunID", id, err)
		}
	}
	if err := ValidateRunID("run-valid"); err != nil {
		t.Errorf("ValidateRunID(run-valid) = %v, veut nil", err)
	}
}

func TestScanRejectsHostileFilter(t *testing.T) {
	dir := t.TempDir()
	_, err := Scan(dir, "../outside")
	if !errors.Is(err, ErrInvalidRunID) {
		t.Fatalf("Scan filtre hostile : %v, veut ErrInvalidRunID", err)
	}
	_, err = Scan(dir, `a\b`)
	if !errors.Is(err, ErrInvalidRunID) {
		t.Fatalf("Scan filtre Windows : %v, veut ErrInvalidRunID", err)
	}
	out, err := Scan(dir, "")
	if err != nil {
		t.Fatalf("Scan filtre vide : %v, veut nil", err)
	}
	if len(out) != 0 {
		t.Fatalf("Scan filtre vide : %d groupes, veut 0", len(out))
	}
}

func TestValidRunIDRejectsInvalidUTF8(t *testing.T) {
	raw := strings.Repeat("a", 7) + "\xff\xfe"
	if ValidRunID(raw) {
		t.Fatal("UTF-8 invalide accepté")
	}
}
