# CGO / Meteolink Chunk 2C — Fail-Closed Topology & Single RTT Convention Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refuse a campaign cell whose bench topology is missing or contradictory, enforce one `DelayMs/2`-per-hop RTT convention in Go and bash, and default the server to `dualns` with `legacy` behind an explicit flag.

**Architecture:** `pkg/campagne/topology.go` validates the full written topology contract (shape mode, rtt mode, delay, rate, IFB) and fails closed when the file is absent. The dualns bash script applies half the profile delay on each shaped hop. `cmd/cgo/server.go` chooses `dualns` by default unless `--legacy-bench` is passed.

**Tech Stack:** Go 1.25, Git Bash, `kit/cgo-bench-dualns.sh`, `docs/verification/2026-11-15-bench-vm/README.md`.

**Source:** `docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md` §3 (verbatim corpus `CGO4.md` 34–42). Chunk 2 is decomposed; this is sub-plan C.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, frontend, live bench, remote hosts, frozen corpus/schemas.
- Corpus paths `kit/run-dualns.sh` and `cmd/cgo/server.go` are real; `pkg/campagne/topology.go` is real. The bench script is `kit/cgo-bench-dualns.sh`.
- French comments in Go; the bash script keeps its existing English comments; English plan/gate text.
- One commit, exact message: `fix(engine): fail closed on topology and use one RTT convention`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: `CheckBenchTopology` fails closed and validates the contract

**Files:**
- Modify: `pkg/campagne/topology.go`, `pkg/campagne/topology_test.go`
- Test: updated existing tests are the oracle

**Interfaces:**
- Consumes: `benchTopology` struct currently only `{ShapeMode string}`; `CheckBenchTopology(expected string) error`.
- Produces: fails on missing file, invalid `rtt_mode` (must be `"per_hop_half"`), missing/invalid `delay_ms` or `rate_mbps`, and `ifb_active` mismatching `expected=="ingress"`.

- [ ] **Step 1: Change the stale test first**

In `pkg/campagne/topology_test.go`, replace `TestCheckBenchTopologyAbsentFileIsLegal` with:

```go
func TestCheckBenchTopologyAbsentFileFailsClosed(t *testing.T) {
	topoEnv(t)
	if err := CheckBenchTopology("ingress"); err == nil {
		t.Fatal("missing bench topology must fail closed")
	}
	if err := CheckBenchTopology(""); err == nil {
		t.Fatal("missing bench topology must fail closed for the default direction too")
	}
}
```

Update the other topology fixtures to include `"rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20` so they remain valid under the tightened contract.

- [ ] **Step 2: Confirm FAIL** — `go test -count=1 -run 'TestCheckBenchTopology' ./pkg/campagne/` → FAIL (absent file still legal, new fields ignored).

- [ ] **Step 3: Implement**

Extend `benchTopology`:

```go
type benchTopology struct {
	ShapeMode string  `json:"shape_mode"`
	RTTMode   string  `json:"rtt_mode"`
	DelayMs   float64 `json:"delay_ms"`
	RateMbps  float64 `json:"rate_mbps"`
	IFBActive bool    `json:"ifb_active"`
}
```

In `CheckBenchTopology`, change the `os.IsNotExist` branch to return `fmt.Errorf("topologie banc absente (%s): fail closed", path)`. After unmarshal and the existing `shape_mode` validation, add:

```go
	if topo.RTTMode != "per_hop_half" {
		return fmt.Errorf("topologie banc (%s): rtt_mode %q invalide (per_hop_half requis)", path, topo.RTTMode)
	}
	if !(topo.DelayMs > 0) || !(topo.RateMbps > 0) {
		return fmt.Errorf("topologie banc (%s): delay_ms et rate_mbps doivent être > 0", path)
	}
	if topo.IFBActive != (expected == "ingress") {
		return fmt.Errorf("topologie banc (%s): ifb_active=%v contredit la direction %q", path, topo.IFBActive, expected)
	}
```

- [ ] **Step 4: Verify** — targeted tests PASS; `go test -count=1 ./pkg/campagne/...` ok.

---

### Task 2: One `DelayMs/2` RTT convention in the bench script and doc

**Files:**
- Modify: `kit/cgo-bench-dualns.sh`, `kit/run-dualns.sh`, `docs/verification/2026-11-15-bench-vm/README.md`
- Test: `bash -n` on both scripts + static grep (no live bench)

**Interfaces:**
- Consumes: script variables `DELAY_MS=100`, `RTT_MS=100`, and two `tc ... netem delay "${DELAY_MS}ms"` lines at `:233` and `:238`.
- Produces: each hop applies `DELAY_MS / 2`; the profile's total RTT stays `DELAY_MS`.

- [ ] **Step 1: Add the half-delay derivation**

After `DELAY_MS=100` and `RTT_MS=100`, insert:

```bash
# One RTT convention: DELAY_MS is the TOTAL path RTT. Each shaped hop gets
# half, one-way — two hops then reproduce the profile's RTT, not twice it.
HOP_DELAY_MS=$(( DELAY_MS / 2 ))
```

- [ ] **Step 2: Replace both hop delays**

Change both occurrences of `delay "${DELAY_MS}ms" loss "${LOSS_PCT}%"` (lines ~233 and ~238) to `delay "${HOP_DELAY_MS}ms" loss "${LOSS_PCT}%"`.

Update the `up — netem ...` log line at :286 to print `${HOP_DELAY_MS}ms` per hop and note the total `${DELAY_MS}ms` RTT.

- [ ] **Step 3: Emit the validated contract from `kit/run-dualns.sh`**

At `kit/run-dualns.sh:74-78`, the topology writer currently emits `{"shape_mode":...,"rate_mbit":20,"delay_ms":100,"loss_pct":0,...,"ifb":...,"written_at":...}`. Add the fields required by Task 1 alongside the existing ones (keep `ifb` unchanged for historical consumers):

```bash
    IFB_ACTIVE=false
    if [ "$SHAPE_MODE" = "ingress" ]; then IFB_ACTIVE=true; fi
    printf '{"shape_mode":"%s","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"loss_pct":0,"ifb_active":%s,"namespaces":["ns_client","ns_wan_emulator","ns_edge_router","ns_server"],"ifb":%s,"written_at":"%s"}\n' \
      "${SHAPE_MODE:-egress}" "$IFB_ACTIVE" "$IFB_JSON" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$TOPO.tmp"
```

Verify: `bash -n kit/run-dualns.sh` exit 0; `grep -n 'rtt_mode' kit/run-dualns.sh` → one hit.

- [ ] **Step 4: Correct the doc**

In `docs/verification/2026-11-15-bench-vm/README.md:95`, change the description of the emulated path from `100+100 ms` to `50 ms + 50 ms = 100 ms RTT` and add a sentence that any historical 200 ms run predates this fix.

- [ ] **Step 5: Verify** — `bash -n kit/cgo-bench-dualns.sh && bash -n kit/run-dualns.sh` exit 0; `grep -n 'delay "\${HOP_DELAY_MS}ms"' kit/cgo-bench-dualns.sh` → 2 hits; `git diff --stat -- book/ docs/VERIFY.md docs/frozen-manifests.sha256` → EMPTY. No live bench run.

---

### Task 3: `dualns` default behind `--legacy-bench`

**Files:**
- Modify: `cmd/cgo/server.go`
- Test: existing `cmd/cgo` tests

**Interfaces:**
- Consumes: `bench := o.Bench` / env / `"legacy"` fallback at `:180-186`.
- Produces: default `dualns`; `legacy` only when the operator passes `--legacy-bench` (or explicitly sets `CGO_BENCH=legacy`).

- [ ] **Step 1: Change the fallback**

Replace `bench = "legacy"` with `bench = "dualns"` and update the French comment name.

- [ ] **Step 2: Verify** — `go test -count=1 ./cmd/cgo/...` ok; `go build ./cmd/cgo` exit 0.

---

### Task 4: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–3.

- [ ] **Step 1: Suites** — `go test -count=1 ./pkg/campagne/... ./cmd/cgo/...`; `bash -n kit/cgo-bench-dualns.sh` — exit 0.
- [ ] **Step 2: Static proofs** — `git grep -n 'fail closed' pkg/campagne/topology.go`; `grep -c 'HOP_DELAY_MS' kit/cgo-bench-dualns.sh` ≥ 3; `git grep -n 'bench = "dualns"' cmd/cgo/server.go`.
- [ ] **Step 3: Commit once** — stage only these six files by explicit path: `pkg/campagne/topology.go`, `pkg/campagne/topology_test.go`, `kit/cgo-bench-dualns.sh`, `kit/run-dualns.sh`, `docs/verification/2026-11-15-bench-vm/README.md`, `cmd/cgo/server.go`; message `fix(engine): fail closed on topology and use one RTT convention`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
