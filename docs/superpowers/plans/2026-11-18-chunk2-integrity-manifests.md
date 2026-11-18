# CGO / Meteolink Chunk 2B — Real Manifest Integrity (K6/K8) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/api/integrity` verify that every file declared in each run's `manifest.json` exists on disk and matches its recorded SHA-256, reporting per-run status and a top-level `ok: false` on any failure.

**Architecture:** `api/corpus.go` gains `verifyRunManifests()` called from `buildIntegrity()`. It reads each `manifest.json`, iterates declared files, hashes each with `sha256.Sum256`, and emits a per-run status array plus `ok`/`integrity` summary. The existing counting/ETag/cache path is preserved; the cache key already reflects corpus fingerprints. No frozen file is modified — this is read-only verification.

**Tech Stack:** Go 1.25 `encoding/json`, `crypto/sha256`, `os`, `filepath`; module `api/`; `results` module for `RunsRoot()`.

**Source:** `docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md` §2 (verbatim corpus `CGO4.md` 25–33). Chunk 2 is decomposed; this is sub-plan B.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- **Frozen corpus and schemas are read-only**: `book/data/runs`, `results.RunsRoot()` contents, `aqm_eval.csv`, `docs/frozen-manifests.sha256`, `VERSION`. Tests use `t.TempDir()` + chdir, never the committed corpus.
- Never touch auth, SSRF, CI, Dockerfile, frontend, live bench, remote hosts, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`.
- Real file path: `api/corpus.go` (`buildIntegrity` at `:54`); corpus's `pkg/api` is stale.
- French comments/user text; English plan/gate text.
- One commit, exact message: `fix(api): verify manifest files and hashes in integrity endpoint`.
- Malformed-CSV row counting (`ReadAQM` tolerant short rows) is explicitly a separate follow-up — do not change that behavior here.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Manifest verification helper with hostile oracles

**Files:**
- Modify: `api/corpus.go`
- Test: `api/corpus_manifest_test.go` (create)

**Interfaces:**
- Consumes: each run directory's `manifest.json` with a `files` object mapping filename → `{sha256: "..."}` (inspect the real JSON shape first; use its actual structure, not an assumed one).
- Produces: `func verifyRunManifests() (ok bool, statuses []map[string]any)` — status values `verified`, `missing_file`, `hash_mismatch`, `no_manifest`.

- [ ] **Step 1: Write the failing test**

Create `api/corpus_manifest_test.go` using the repo's `chdirTemp(t)` helper (from `api/integrity_test.go`) and a manifest fixture built from a real file's SHA-256:

```go
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func writeManifestFixture(t *testing.T, dir, run, content string) string {
	t.Helper()
	p := filepath.Join(dir, "runs", run)
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

func TestVerifyRunManifestsDetectsMissingAndHashMismatch(t *testing.T) {
	chdirTemp(t)
	good := writeManifestFixture(t, ".", "run-good", "a,b\n1,2\n")
	bad := writeManifestFixture(t, ".", "run-bad", "a,b\n1,2\n")
	if err := os.WriteFile(filepath.Join(bad, "aqm_eval.csv"), []byte("a,b\n9,9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(good, "aqm_eval.csv"))

	ok, statuses := verifyRunManifests()
	if ok {
		t.Fatal("integrity must be false with a missing file and a hash mismatch")
	}
	got := map[string]string{}
	for _, s := range statuses {
		got[s["run"].(string)] = s["status"].(string)
	}
	if got["run-good"] != "missing_file" {
		t.Errorf("missing file status = %q, want missing_file", got["run-good"])
	}
	if got["run-bad"] != "hash_mismatch" {
		t.Errorf("hash mismatch status = %q, want hash_mismatch", got["run-bad"])
	}
}
```

Inspect the real manifest JSON shape before finalizing the fixture — if the `files` value is a bare hash string rather than an object, encode that exact shape. Use `results.RunsRoot()` semantics: `chdirTemp(t)` must place `runs/` where `RunsRoot()` looks (copy the pattern from `api/integrity_test.go`).

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestVerifyRunManifestsDetectsMissingAndHashMismatch ./...` → FAIL (`undefined: verifyRunManifests`).

- [ ] **Step 3: Implement the helper**

In `api/corpus.go`, add imports `crypto/sha256`, `encoding/hex`, `encoding/json`, and:

```go
type manifestFile struct {
	File   string `json:"File"`
	SHA256 string `json:"Sha256"`
}

type runManifest struct {
	Files []manifestFile `json:"files"`
}

// verifyRunManifests — lecture seule : chaque fichier déclaré doit exister et
// son SHA-256 doit correspondre. Un run sans manifest est signalé, pas ignoré.
func verifyRunManifests() (bool, []map[string]any) {
	runs, _ := filepath.Glob(filepath.Join(results.RunsRoot(), "*"))
	sort.Strings(runs)
	statuses := []map[string]any{}
	ok := true
	for _, p := range runs {
		st := map[string]any{"run": filepath.Base(p), "status": "verified"}
		data, err := os.ReadFile(filepath.Join(p, "manifest.json"))
		if err != nil {
			st["status"] = "no_manifest"
			ok = false
			statuses = append(statuses, st)
			continue
		}
		var m runManifest
		if err := json.Unmarshal(data, &m); err != nil {
			st["status"] = "invalid_manifest"
			ok = false
			statuses = append(statuses, st)
			continue
		}
		for _, f := range m.Files {
			body, err := os.ReadFile(filepath.Join(p, f.File))
			if err != nil {
				st["status"] = "missing_file"
				ok = false
				break
			}
			h := sha256.Sum256(body)
			if hex.EncodeToString(h[:]) != f.SHA256 {
				st["status"] = "hash_mismatch"
				ok = false
				break
			}
		}
		statuses = append(statuses, st)
	}
	return ok, statuses
}
```

The shape above is verified against `docs/verification/2026-11-16-v2-corpus/run-1794853778/manifest.json`: top-level `config`, `config_sha256`, `files`; `files` is an array of objects with keys `File` and `Sha256`. Keep the status vocabulary as written.

- [ ] **Step 4: Verify** — targeted test PASS; `go -C api test -count=1 ./...` ok.

---

### Task 2: Expose verification in `/api/integrity`

**Files:**
- Modify: `api/corpus.go`
- Test: `api/corpus_manifest_test.go` (extend)

**Interfaces:**
- Consumes: `verifyRunManifests()` from Task 1.
- Produces: `buildIntegrity()` returns `ok`, `integrity` status and `manifest_statuses`; existing counts unchanged.

- [ ] **Step 1: Write the failing assertion**

Append to `TestVerifyRunManifestsDetectsMissingAndHashMismatch`:

```go
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
```

- [ ] **Step 2: Confirm FAIL** — targeted test → FAIL (keys absent).

- [ ] **Step 3: Wire it**

In `buildIntegrity()`, call `ok, statuses := verifyRunManifests()` before building the returned map and add:

```go
	integrityStatus := "verified"
	if !ok {
		integrityStatus = "degraded"
	}
```
and to the returned map: `"ok": ok, "integrity": integrityStatus, "manifest_statuses": statuses`.

- [ ] **Step 4: Verify** — targeted PASS; full api suite ok. Existing `TestIntegrityCounts`/`TestIntegrityProvenance` must still pass unchanged.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `go -C api test -count=1 ./...` → exit 0; `cd results && go test -count=1 ./...` → exit 0 (results module untouched but verified).
- [ ] **Step 2: Static proof** — `git grep -n 'verifyRunManifests' api/` → ≥2 hits; `git diff --stat -- book/ results/ docs/frozen-manifests.sha256` → EMPTY.
- [ ] **Step 3: Commit once** — stage only `api/corpus.go` and `api/corpus_manifest_test.go` by explicit path; message `fix(api): verify manifest files and hashes in integrity endpoint`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
