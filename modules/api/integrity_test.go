package api

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// Triple provenance (registre maître) : /api/integrity expose le sha256 du
// dernier aqm_eval.csv (complet + 8 caractères) pour que le tiroir Wall + la pastille
// Archives + le Rapport + les figures RDF portent le MÊME hash8. Chdir vers un
// dossier temp — ne jamais polluer le dépôt.
// TestIntegrityCounts — valid strict, degraded and quarantined sont trois
// comptes distincts ; « valid » n'inclut jamais les lignes dégradées.
func TestIntegrityCounts(t *testing.T) {
	dir := chdirTemp(t)
	writeCSVRun(t, dir, "run-counts", header18,
		"run-counts,1,P2,cake,bbr,1,100,114,14,209.2,98.1,18.1,81,0,117288,12.10,0.0,valid",
		"run-counts,2,P2,cake,bbr,1,101,115,14,209.3,98.0,18.0,82,0,117300,12.2,0.0,degraded",
		"run-counts,3,P2,cake,bbr,1,102,116,14,209.4,97.9,17.9,83,0,117400,12.3,0.0,invalid",
	)
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/integrity")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{`"valid":1`, `"degraded":1`, `"quarantined":1`, `"rows":3`} {
		if !strings.Contains(s, want) {
			t.Fatalf("integrity counts missing %s in %s", want, s)
		}
	}
}

func TestIntegrityProvenance(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	run := filepath.Join(dir, "data", "runs", "run-x")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	csv := strings.Join(model.AQMEvalHeader, ",") + "\nrun-x,1,P1,pfifo_fast,bbr,1,20,22.5,46.9,100,235.1,0,0,1,11.6,12,valid\n"
	if err := os.WriteFile(filepath.Join(run, "aqm_eval.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(csv))
	full := hex.EncodeToString(sum[:])

	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/integrity")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	if !strings.Contains(s, `"hash8":"`+full[:8]+`"`) {
		t.Fatalf("integrity hash8 mismatch: %s", s)
	}
	if !strings.Contains(s, `"sha256":"`+full+`"`) {
		t.Fatalf("integrity sha256 missing: %s", s)
	}
}
