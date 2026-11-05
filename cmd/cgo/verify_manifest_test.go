package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testFile struct {
	File   string `json:"File"`
	Sha256 string `json:"Sha256"`
}

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func makeRun(t *testing.T, root, name string, config []byte, configHash string) (string, map[string][]byte) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"aqm_eval.csv":    []byte("run_id,event_id\nr1,1\n"),
		"quarantine.json": []byte("[]"),
	}
	var entries []testFile
	for fn, body := range files {
		if err := os.WriteFile(filepath.Join(dir, fn), body, 0o644); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, testFile{File: fn, Sha256: sha(body)})
	}
	doc, err := json.Marshal(map[string]any{
		"config_sha256": configHash,
		"config":        string(config),
		"files":         entries,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), doc, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, files
}

func TestVerifyManifestsRequiresEveryRunAndRealConfig(t *testing.T) {
	root := t.TempDir()
	cfg := []byte(`{"profiles":["P1"],"reps":1}`)
	dir, _ := makeRun(t, root, "run1", cfg, sha(cfg))
	if ok, err := verifyManifests(root); !ok || err != nil {
		t.Fatalf("manifeste valide refusé: ok=%v err=%v", ok, err)
	}

	// Le hash factice historique doit échouer.
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"config_sha256":"config","config":"{\"profiles\":[\"P1\"]}","files":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := verifyManifests(root); ok || err == nil || !strings.Contains(err.Error(), "config_sha256") {
		t.Fatalf("hash factice accepté: ok=%v err=%v", ok, err)
	}

	// Un run sans manifeste doit échouer (ancien comportement : ignoré).
	dir, _ = makeRun(t, root, "run1", cfg, sha(cfg))
	if err := os.Remove(filepath.Join(dir, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	if ok, err := verifyManifests(root); ok || err == nil || !strings.Contains(err.Error(), "manifeste manquant") {
		t.Fatalf("run sans manifeste accepté: ok=%v err=%v", ok, err)
	}

	// Un fichier non couvert doit échouer.
	dir, _ = makeRun(t, root, "run1", cfg, sha(cfg))
	if err := os.WriteFile(filepath.Join(dir, "extra.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := verifyManifests(root); ok || err == nil || !strings.Contains(err.Error(), "non couvert") {
		t.Fatalf("fichier non couvert accepté: ok=%v err=%v", ok, err)
	}
}

func TestVerifyManifestsDetectsTamper(t *testing.T) {
	root := t.TempDir()
	cfg := []byte(`{"profiles":["P1"],"reps":1}`)
	dir, files := makeRun(t, root, "run1", cfg, sha(cfg))
	files["aqm_eval.csv"] = []byte("run_id,event_id\nr1,2\n")
	if err := os.WriteFile(filepath.Join(dir, "aqm_eval.csv"), files["aqm_eval.csv"], 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := verifyManifests(root); ok || err == nil || !strings.Contains(err.Error(), "sha mismatch") {
		t.Fatalf("fichier modifié accepté: ok=%v err=%v", ok, err)
	}
}
