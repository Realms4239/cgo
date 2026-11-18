package api

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// writeManifestFixture — corpus éphémère au format réel (runs/<run>/manifest.json
// avec files:[{File,Sha256}]). Ne jamais toucher au corpus gelé. RunsRoot() = data/runs.
func writeManifestFixture(t *testing.T, dir, run, content string) string {
	t.Helper()
	p := filepath.Join(dir, "data", "runs", run)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "aqm_eval.csv"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(content))
	manifest := `{"files":[{"File":"aqm_eval.csv","Sha256":"` + hex.EncodeToString(h[:]) + `"}]}`
	if err := os.WriteFile(filepath.Join(p, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func manifestStatuses(statuses []map[string]any) map[string]string {
	got := map[string]string{}
	for _, s := range statuses {
		got[s["run"].(string)] = s["status"].(string)
	}
	return got
}

// TestVerifyRunManifestsDetectsMissingAndHashMismatch — oracles hostiles :
// un fichier déclaré disparu et un fichier modifié doivent tous deux faire
// échouer l'intégrité, tandis qu'un run intact reste « verified » (oracle positif).
func TestVerifyRunManifestsDetectsMissingAndHashMismatch(t *testing.T) {
	chdirTemp(t)
	writeManifestFixture(t, ".", "run-ok", "a,b\n1,2\n")
	good := writeManifestFixture(t, ".", "run-good", "a,b\n1,2\n")
	bad := writeManifestFixture(t, ".", "run-bad", "a,b\n1,2\n")
	if err := os.WriteFile(filepath.Join(bad, "aqm_eval.csv"), []byte("a,b\n9,9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(good, "aqm_eval.csv")); err != nil {
		t.Fatal(err)
	}

	ok, statuses := verifyRunManifests()
	if ok {
		t.Fatal("integrity must be false with a missing file and a hash mismatch")
	}
	got := manifestStatuses(statuses)
	if got["run-ok"] != "verified" {
		t.Errorf("clean run status = %q, want verified", got["run-ok"])
	}
	if got["run-good"] != "missing_file" {
		t.Errorf("missing file status = %q, want missing_file", got["run-good"])
	}
	if got["run-bad"] != "hash_mismatch" {
		t.Errorf("hash mismatch status = %q, want hash_mismatch", got["run-bad"])
	}

	v := buildIntegrity()
	if v["ok"] != false {
		t.Errorf("integrity ok = %v, want false", v["ok"])
	}
	if v["integrity"] != "degraded" {
		t.Errorf("integrity status = %v, want degraded", v["integrity"])
	}
	if _, present := v["manifest_statuses"]; !present {
		t.Error("integrity payload missing manifest_statuses")
	}
}

// TestVerifyRunManifestsAllValid — oracle positif : corpus intact ⇒ ok=true.
func TestVerifyRunManifestsAllValid(t *testing.T) {
	chdirTemp(t)
	writeManifestFixture(t, ".", "run-1", "a,b\n1,2\n")
	writeManifestFixture(t, ".", "run-2", "x,y\n3,4\n")

	ok, statuses := verifyRunManifests()
	if !ok {
		t.Fatalf("integrity must be true with valid manifests, got %v", manifestStatuses(statuses))
	}
	got := manifestStatuses(statuses)
	for _, run := range []string{"run-1", "run-2"} {
		if got[run] != "verified" {
			t.Errorf("%s status = %q, want verified", run, got[run])
		}
	}
}

// TestVerifyRunManifestsNoManifest — un run sans manifest est signalé, pas ignoré.
func TestVerifyRunManifestsNoManifest(t *testing.T) {
	chdirTemp(t)
	p := writeManifestFixture(t, ".", "run-no-manifest", "a,b\n1,2\n")
	if err := os.Remove(filepath.Join(p, "manifest.json")); err != nil {
		t.Fatal(err)
	}

	ok, statuses := verifyRunManifests()
	if ok {
		t.Fatalf("integrity must be false with no manifest, got %v", manifestStatuses(statuses))
	}
	if got := manifestStatuses(statuses); got["run-no-manifest"] != "no_manifest" {
		t.Errorf("no-manifest status = %q, want no_manifest", got["run-no-manifest"])
	}
}
