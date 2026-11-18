# CGO / Meteolink Chunk 2A — Absent Is Never a Zero (K4) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Never let a failed or empty `tc -j` scrape be published as a measured zero: expose a measured flag from the qdisc helpers and have the campaign mark such cells `invalid` via the JSON-only `failed_gates` field.

**Architecture:** `qdisc/stats.go` gains measured variants of `SumDrops`/`SumBytes` that return `(value, false)` for an empty/unmeasured snapshot. `pkg/campagne/campagne.go` uses them at both ends of the load phase; when either end is unmeasured it stops deriving drops/goodput from `tc` and appends the JSON-only gate name `G8StatsMeasured`, forcing `gate_status = invalid`. The frozen `aqm_eval.csv` schema is untouched because `FailedGates` carries `json:"failed_gates,omitempty"` and no `csv` tag.

**Tech Stack:** Go 1.25, modules `qdisc/` and root `pkg/campagne/`, `httptest`-free unit oracles.

**Source:** `docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md` §1 (verbatim corpus `CGO4.md` 15–22). Chunk 2 is decomposed; this is sub-plan A.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- **Frozen schemas are untouchable**: `aqm_eval.csv`, `telemetry.csv`, `link_audit.csv`, `book/data/runs`, `docs/frozen-manifests.sha256`.
- Never touch auth, SSRF, `Dockerfile`, CI, frontend, live bench, or remote hosts.
- Corpus paths `pkg/qdisc/` and `pkg/results/` are stale: real modules are `qdisc/` and `results/`; `pkg/campagne/` is correct.
- French comments; English plan/gate text.
- One commit, exact message: `fix(engine): never publish unmeasured tc counters as zero drops`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Measured variants in `qdisc/stats.go`

**Files:**
- Modify: `qdisc/stats.go`
- Test: `qdisc/stats_measured_test.go` (create)

**Interfaces:**
- Consumes: existing `SumDrops(stats []Stats) uint64`, `SumBytes(stats []Stats) uint64`.
- Produces: `SumDropsMeasured(stats []Stats) (uint64, bool)`, `SumBytesMeasured(stats []Stats) (uint64, bool)` — `false` iff the snapshot is empty/unmeasured.

- [ ] **Step 1: Write the failing test**

Create `qdisc/stats_measured_test.go`:

```go
package qdisc

import "testing"

func TestSumDropsMeasuredReportsAbsence(t *testing.T) {
	if v, ok := SumDropsMeasured(nil); ok || v != 0 {
		t.Fatalf("SumDropsMeasured(nil) = (%d, %v), want (0, false)", v, ok)
	}
	if v, ok := SumDropsMeasured([]Stats{}); ok || v != 0 {
		t.Fatalf("SumDropsMeasured(empty) = (%d, %v), want (0, false)", v, ok)
	}
	v, ok := SumDropsMeasured([]Stats{{Drops: 7}})
	if !ok || v != 7 {
		t.Fatalf("SumDropsMeasured(7) = (%d, %v), want (7, true)", v, ok)
	}
}

func TestSumBytesMeasuredReportsAbsence(t *testing.T) {
	if v, ok := SumBytesMeasured(nil); ok || v != 0 {
		t.Fatalf("SumBytesMeasured(nil) = (%d, %v), want (0, false)", v, ok)
	}
	v, ok := SumBytesMeasured([]Stats{{Bytes: 11}})
	if !ok || v != 11 {
		t.Fatalf("SumBytesMeasured(11) = (%d, %v), want (11, true)", v, ok)
	}
}
```

The `Stats` fields are `Bytes` and `Drops` (`qdisc/stats.go:9-16`), so the literals above compile as written.

- [ ] **Step 2: Confirm FAIL** — `cd qdisc && go test -count=1 -run 'TestSumDropsMeasuredReportsAbsence|TestSumBytesMeasuredReportsAbsence' ./...` → FAIL (`undefined`).

- [ ] **Step 3: Implement**

In `qdisc/stats.go`, add after `SumDrops`:

```go
// SumDropsMeasured — like SumDrops, but says whether a scrape actually happened.
// Un snapshot vide ne doit jamais devenir un 0 mesuré dans un rapport scellé.
func SumDropsMeasured(stats []Stats) (uint64, bool) {
	if len(stats) == 0 {
		return 0, false
	}
	return SumDrops(stats), true
}
```

And after `SumBytes`:

```go
// SumBytesMeasured — même contrat que SumDropsMeasured pour le compteur feuille.
func SumBytesMeasured(stats []Stats) (uint64, bool) {
	if len(stats) == 0 {
		return 0, false
	}
	return SumBytes(stats), true
}
```

- [ ] **Step 4: Verify** — targeted PASS; `cd qdisc && go test -count=1 ./...` ok.

---

### Task 2: Campaign marks unmeasured scraps invalid

**Files:**
- Modify: `pkg/campagne/campagne.go`
- Test: `pkg/campagne/stats_measured_test.go` (create)

**Interfaces:**
- Consumes: `qdisc.SumDropsMeasured`, `qdisc.SumBytesMeasured`; `model.FailedGates` is JSON-only (`model/types.go:165`).
- Produces: when `StatsFn()` yields no stats at either end of the load phase, the event gets `failed_gates: ["G8StatsMeasured"]` and `gate_status: invalid`; no tc-derived goodput is substituted.

- [ ] **Step 1: Write the failing test**

Create `pkg/campagne/stats_measured_test.go` with a unit oracle on the decision helper you add (no full campaign run):

```go
package campagne

import (
	"testing"

	"github.com/Realms4239/cgo/qdisc"
)

func TestStatsUnmeasuredForcesInvalidGate(t *testing.T) {
	_, dropsOK := qdisc.SumDropsMeasured(nil)
	_, bytesOK := qdisc.SumBytesMeasured(nil)
	if !statsUnmeasured(dropsOK, bytesOK) {
		t.Fatal("nil snapshot must count as unmeasured")
	}
	_, dropsOK = qdisc.SumDropsMeasured([]qdisc.Stats{{Drops: 1}})
	_, bytesOK = qdisc.SumBytesMeasured([]qdisc.Stats{{Bytes: 1}})
	if statsUnmeasured(dropsOK, bytesOK) {
		t.Fatal("populated snapshot must count as measured")
	}
}

func TestAppendStatsFailureGateIsJSONOnly(t *testing.T) {
	var failed []string
	failed = appendStatsFailureGate(failed)
	if len(failed) != 1 || failed[0] != "G8StatsMeasured" {
		t.Fatalf("failed gates = %v, want [G8StatsMeasured]", failed)
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go test -count=1 -run 'TestStatsUnmeasuredForcesInvalidGate|TestAppendStatsFailureGateIsJSONOnly' ./pkg/campagne/` → FAIL (`undefined`).

- [ ] **Step 3: Implement**

In `pkg/campagne/campagne.go`, add near the campaign helpers:

```go
func statsUnmeasured(dropsOK, bytesOK bool) bool {
	return !dropsOK || !bytesOK
}

func appendStatsFailureGate(failed []string) []string {
	return append(failed, "G8StatsMeasured")
}
```

In the load-phase block that currently calls `qdisc.SumDrops(sts)` / `qdisc.SumBytes(sts)`, use the measured variants. Track whether the *start* snapshot was measured alongside `startDrops`/`startBytes`; at the end, if either snapshot was unmeasured:
- do not compute `ev.Drops` / `rxBytes` deltas from tc (leave the CSV fields at their existing values, never invent a measured zero),
- append `G8StatsMeasured` to `ev.FailedGates` via `appendStatsFailureGate`,
- force `status = model.GateInvalid` in the final gate loop (the existing `default:` branch already maps a failed gate to invalid; add the explicit case if needed).

- [ ] **Step 4: Verify** — targeted tests PASS; `go test -count=1 ./pkg/campagne/...` ok.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `cd qdisc && go test -count=1 ./...`; `go test -count=1 ./pkg/campagne/... ./pkg/audit/...` — both exit 0.
- [ ] **Step 2: Static proofs** — `git grep -n 'SumDropsMeasured' qdisc/ pkg/campagne/` (≥2 hits); `git grep -n 'G8StatsMeasured' pkg/campagne/` (≥1 hit); `git diff --stat -- book/ results/ docs/frozen-manifests.sha256` → EMPTY (frozen assets untouched).
- [ ] **Step 3: Commit once** — stage only the four named files by explicit path; message `fix(engine): never publish unmeasured tc counters as zero drops`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs and exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
