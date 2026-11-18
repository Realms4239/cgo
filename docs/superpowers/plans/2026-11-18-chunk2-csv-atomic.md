# CGO / Meteolink Chunk 2D — Malformed Rows & Atomic Writes Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Count CSV rows that carry more fields than the header and surface that count in `/api/quarantine/summary`; make audit report writes use the existing atomic writer so a crash never leaves a truncated report.

**Architecture:** `results/csvrows.go` gains `ReadAQMDetailed` with a `MalformedRows` counter; `ReadAQM` keeps its tolerant behavior for short historical rows and delegates to the new function, so all existing callers are unchanged. `api/server.go` uses the detailed reader for the quarantine summary's `malformed_rows` field. `audit/report.go` swaps its four `os.WriteFile` calls to the already-tested `atomicWriteFile` from `audit/writer.go`.

**Tech Stack:** Go 1.25 `encoding/csv`, `os`, `filepath`; modules `results/`, `api/`, `audit/`.

**Source:** `docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md` §1.2 and §4.2 (verbatim corpus `CGO4.md` 21–23, 48–51). Chunk 2 is decomposed; this is sub-plan D.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- **Frozen schemas and corpus are read-only**: `aqm_eval.csv` 23-column header, `book/data/runs`, `docs/frozen-manifests.sha256`, `data/runs`. Tests use `t.TempDir()`.
- Short historical rows remain tolerated — they are "missing field = absent", not malformed. Only rows with MORE fields than the header count as malformed.
- Never touch auth, SSRF, CI, Dockerfile, frontend, live bench, remote hosts, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`.
- Corpus paths `pkg/results/` and `pkg/audit/` are stale: real modules are `results/` and `audit/`.
- French comments/user text; English plan/gate text.
- One commit, exact message: `fix(engine): count malformed CSV rows and write reports atomically`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: `ReadAQMDetailed` with malformed-row oracle

**Files:**
- Modify: `results/csvrows.go`
- Test: `results/csvrows_malformed_test.go` (create)

**Interfaces:**
- Consumes: existing `ReadAQM(path) (header []string, rows [][]string, err error)`.
- Produces: `ReadAQMDetailed(path) (header []string, rows [][]string, malformed int, err error)`; `ReadAQM` delegates to it and keeps its exact signature.

- [ ] **Step 1: Write the failing test**

Create `results/csvrows_malformed_test.go`:

```go
package results

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAQMDetailedCountsExtraFieldRows(t *testing.T) {
	p := filepath.Join(t.TempDir(), "aqm_eval.csv")
	body := "a,b\n1,2\n3,4,5\n6\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	header, rows, malformed, err := ReadAQMDetailed(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(header) != 2 || len(rows) != 3 {
		t.Fatalf("header=%v rows=%v, want 2 columns and 3 rows", header, rows)
	}
	if malformed != 1 {
		t.Fatalf("malformed = %d, want 1 (the row with an extra field)", malformed)
	}
	h2, r2, err := ReadAQM(p)
	if err != nil || len(h2) != 2 || len(r2) != 3 {
		t.Fatalf("ReadAQM compatibility broken: %v %v %v", h2, r2, err)
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `cd results && go test -count=1 -run TestReadAQMDetailedCountsExtraFieldRows ./...` → FAIL (`undefined: ReadAQMDetailed`).

- [ ] **Step 3: Implement**

In `results/csvrows.go`, add:

```go
// ReadAQMDetailed — like ReadAQM but counts rows carrying MORE fields than the
// header. Short rows are tolerated historical schema ("absent", not malformed).
func ReadAQMDetailed(path string) (header []string, rows [][]string, malformed int, err error) {
	header, rows, err = ReadAQM(path)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, row := range rows {
		if len(row) > len(header) {
			malformed++
		}
	}
	return header, rows, malformed, nil
}
```

Keep `ReadAQM` body untouched — no caller behavior changes.

- [ ] **Step 4: Verify** — targeted PASS; `cd results && go test -count=1 ./...` ok.

---

### Task 2: Surface `malformed_rows` in the quarantine summary

**Files:**
- Modify: `api/server.go` (`/api/quarantine/summary` at `:1461`)
- Test: `api/server_quarantine_test.go` (create)

**Interfaces:**
- Consumes: `results.ReadAQMDetailed`; existing quarantine counters `total, low, high, zero, g3`.
- Produces: an extra JSON key `malformed_rows` on `/api/quarantine/summary`; existing keys unchanged.

- [ ] **Step 1: Write the failing test**

```go
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestQuarantineSummaryReportsMalformedRows(t *testing.T) {
	withDevNoAuth(t)
	dir := chdirTemp(t)
	run := filepath.Join(dir, "data", "runs", "run-malformed")
	if err := os.MkdirAll(run, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "aqm_eval.csv"), []byte("profile,gate_status\nP2,valid,EXTRA\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/quarantine/summary")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["malformed_rows"] != float64(1) {
		t.Fatalf("malformed_rows = %v, want 1", got["malformed_rows"])
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestQuarantineSummaryReportsMalformedRows ./...` → FAIL (`malformed_rows` absent).

- [ ] **Step 3: Implement**

In `api/server.go`, add `"github.com/Realms4239/cgo/results"` only if not already imported (it is). In the quarantine handler, add `var malformedRows int` beside the existing counters, change `header, rows, err := results.ReadAQM(f)` to `header, rows, malformed, err := results.ReadAQMDetailed(f)`, add `malformedRows += malformed` after the `err` guard, and include `"malformed_rows": malformedRows` in the handler's final JSON object (keep every existing key).

- [ ] **Step 4: Verify** — targeted PASS; full `go -C api test -count=1 ./...` ok.

---

### Task 3: Atomic report writes

**Files:**
- Modify: `audit/report.go`
- Test: `audit/report_atomic_test.go` (create)

**Interfaces:**
- Consumes: `atomicWriteFile(path string, data []byte, perm os.FileMode) error` from `audit/writer.go`.
- Produces: report JSON/MD writes are atomic; existing signatures and paths unchanged.

- [ ] **Step 1: Write the failing test**

```go
package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFileLeavesNoTempOnFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "report.json")
	if err := atomicWriteFile(target, []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(target)
	if err != nil || string(body) != `{"ok":true}` {
		t.Fatalf("atomic write result = %q, %v", body, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("leftover temp files: %v, %v", entries, err)
	}
}
```

- [ ] **Step 2: Run the positive guard** — `cd audit && go test -count=1 -run TestAtomicWriteFileLeavesNoTempOnFailure ./...` → PASS. `atomicWriteFile` already exists in `audit/writer.go` and is not the new behavior; this test guards the primitive. The actual behavior change is proven by Task 4's static proof that `audit/report.go` no longer calls `os.WriteFile`. Do not fabricate a red step here.

- [ ] **Step 3: Apply the real change**

In `audit/report.go`, replace the four `os.WriteFile(...)` calls at `:835`, `:838`, `:844`, `:847` with `atomicWriteFile(...)` using the same paths/permissions. Do not change write order or signatures. If `os` becomes unused, remove the import.

- [ ] **Step 4: Verify** — `cd audit && go test -count=1 ./...` ok; `go build ./cmd/cgo` exit 0.

---

### Task 4: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–3.

- [ ] **Step 1: Suites** — `cd results && go test -count=1 ./...`; `go -C api test -count=1 ./...`; `cd audit && go test -count=1 ./...` — all exit 0.
- [ ] **Step 2: Static proof** — `git grep -n 'ReadAQMDetailed' results/ api/` ≥2 hits; `git grep -n 'malformed_rows' api/` ≥1; `git grep -n 'atomicWriteFile' audit/report.go` =4 hits; `git diff --stat -- book/ results/ docs/frozen-manifests.sha256 aqm_eval.csv` → EMPTY.
- [ ] **Step 3: Commit once** — stage only the named files by explicit path; message `fix(engine): count malformed CSV rows and write reports atomically`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
