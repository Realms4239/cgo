package main

import "testing"

func TestClampDuration(t *testing.T) {
	if got := clampDuration(0); got != 60 {
		t.Fatalf("defaut = %d, want 60", got)
	}
	if got := clampDuration(200000); got != maxDurationSec {
		t.Fatalf("plafond dur = %d, want %d", got, maxDurationSec)
	}
	if got := clampDuration(3600); got != 3600 {
		t.Fatalf("legitime = %d, want 3600", got)
	}
}

func TestPseudonymStablePerWindow(t *testing.T) {
	salt := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	a, b := pseudonym("10.0.0.1", salt), pseudonym("10.0.0.1", salt)
	c := pseudonym("10.0.0.2", salt)
	other := pseudonym("10.0.0.1", []byte("other-salt-0123456789abcdef01234"))
	if a != b {
		t.Fatal("meme IP/sel doit rendre le meme label")
	}
	if a == c || a == other {
		t.Fatal("labels distincts requis — le sel change tout")
	}
	if len(a) != 16 {
		t.Fatalf("label = %d hex, want 16 (8 octets)", len(a))
	}
}

func TestPortExcluded(t *testing.T) {
	if !portExcluded("3333,4444", 4444) {
		t.Fatal("port distant 4444 exclut")
	}
	if !portExcluded("3333,4444", 8080, 3333) {
		t.Fatal("port local 3333 exclut")
	}
	if portExcluded("", 4444) {
		t.Fatal("liste vide exclut rien")
	}
	if portExcluded("3333", 4444) {
		t.Fatal("port inconnu non exclu")
	}
}
