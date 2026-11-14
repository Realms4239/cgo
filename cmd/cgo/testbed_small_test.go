package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestTestbedSmallOK — l'objet du banc est servi tel quel : 200,
// Content-Length fidèle, octets EXACTS. La preuve conteneur live (/small=200)
// appartient au gate Phase 2 ; ici le handler est prouvé in-repo.
func TestTestbedSmallOK(t *testing.T) {
	obj := filepath.Join(t.TempDir(), "obj16.bin")
	want := make([]byte, 16*1024)
	for i := range want {
		want[i] = byte(i % 251)
	}
	if err := os.WriteFile(obj, want, 0o644); err != nil {
		t.Fatal(err)
	}
	// testbedMux lit l'env À LA CONSTRUCTION : poser CGO_TESTBED_OBJ avant.
	t.Setenv("CGO_TESTBED_OBJ", obj)

	srv := httptest.NewServer(testbedMux())
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/small")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("code=%d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Length"); got != "16384" {
		t.Fatalf("Content-Length=%q, want 16384", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(want) {
		t.Fatalf("corps %d octets ≠ objet %d octets", len(body), len(want))
	}
}

// TestTestbedSmallMissingObj500 — objet manquant : 500 par design (une
// absence de service n'est jamais une mesure).
func TestTestbedSmallMissingObj500(t *testing.T) {
	t.Setenv("CGO_TESTBED_OBJ", filepath.Join(t.TempDir(), "absent.bin"))

	srv := httptest.NewServer(testbedMux())
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/small")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("code=%d, want 500", resp.StatusCode)
	}
}
