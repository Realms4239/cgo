# Chunk 2 — Data Honesty, Integrity & Lab Physics

> Source: `C:\END\CGO4.md` lines 1–97. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** failed scrapes never become zero, `/api/integrity` verifies real files, topology fails closed, RTT is `DelayMs/2` per hop, writes are atomic and ordered.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** removes fake zeros and orphan manifests from the evidence chain.
- **Usage:** execute after Chunk 1; gates Chunk 3.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 3 — real-time triple-ring pipeline.
Proceeding to Chunk 2 (Data Honesty K4, Real /api/integrity K6, Manifest Repair
K8, Fail-Closed Topology, Single RTT Regime & Atomic Persistence).

Hand Part A to your agent, then run Part B yourself to verify every single
physics and integrity fix before touching Chunk 3.

CHUNK 2 — Data Honesty (K4), Real Integrity (K6, K8), Fail-Closed dualns & Atomic Writes

PART A — Exact Bounded Prompt for Your Coding Agent

# CHUNK 2 EXECUTION TASK — DATA HONESTY, REAL INTEGRITY & LAB PHYSICS (STRICTLY BOUNDED)

You are executing ONLY Chunk 2 on `C:\cgo`. Do NOT modify auth (`api/auth.go`), SSRF (`ssrf.go`), `Dockerfile`, CI workflows, or React views in this chunk.

## 1. Fix `K4` — Enforce "Absent Is Never a Zero" in Scrape & CSV Paths
Files: `pkg/campagne/prod.go:76-86`, `pkg/campagne/dualns.go`, `pkg/qdisc/stats.go:133-139`, `pkg/campagne/campagne.go:637`, `pkg/results/csvrows.go:19-44`
1. **Propagate `PollStats` Errors (`prod.go:76-86`, `dualns.go`, `qdisc/stats.go:133-139`):**
   - Never swallow `PollStats` (`tc -j`) errors and return `nil`! `StatsFn` MUST return `(StatsSnapshot, error)` or a struct with `Measured bool`.
   - In `pkg/qdisc/stats.go:133-139`, change `SumDrops` / `SumBytes` so a `nil`/unmeasured snapshot returns `(0, false)` or an explicit error—NEVER returning `0` disguised as a measured value.
   - In `pkg/campagne/campagne.go:637`: if `StatsFn` fails or `!measured`, set `DropsMeasured = false`, `WastedBytesMeasured = false`, `CostMeasured = false`, emit `""` / `"—"` (never `0`) in CSV/report outputs, and mark the sample `degraded`/`invalid` per gate contract.
   - Similarly, if CPU utilization cannot be read (`CPUMeasured == false`), never write `0.0`; represent it as unmeasured (`nil`/`""`).
2. **Surface Malformed CSV Rows (`pkg/results/csvrows.go:19-44`):**
   - Do not silently skip malformed CSV rows! Track and return `MalformedRows int` and `RowErrors []string` alongside valid rows, surfacing them in `/api/quarantine/summary` and `/api/integrity`.

## 2. Fix `K6` & `K8` — Real `/api/integrity` Verification & Manifest Cleanup
Files: `api/corpus.go:71-73, 99-109`, `docs/verification/2026-11-16-v2-corpus/run-1794853778/manifest.json:9-12`, `docs/verification/2026-11-12-ebpf-probe-bringup/README.md:12`
1. **Real `/api/integrity` (`api/corpus.go:71-73, 99-109`):**
   - Wire real `verifyManifests()` inside `/api/integrity`: for every run directory, read `manifest.json`, verify EVERY declared artifact file physically exists on disk AND its live SHA-256 matches the manifest SHA-256, AND verify `aqm_eval.csv` parses with `MalformedRows == 0`.
   - Return explicit per-run `status` (`"verified"`, `"missing_file"`, `"hash_mismatch"`, `"malformed_csv"`) and top-level `ok: false` / `"degraded"` if ANY check fails.
2. **Fix Manifest & Hash Drift (`K8`):**
   - Inspect `docs/verification/2026-11-16-v2-corpus/run-1794853778/manifest.json:9-12`: remove or regenerate the missing `valuation.json` reference honestly so 100% of files listed in every committed `manifest.json` actually exist and match SHA-256 on disk.
   - Update `docs/verification/2026-11-12-ebpf-probe-bringup/README.md:12` to note/align with the active eBPF object hash (`8612f578…`, clearly separating historical bring-up hash if archived).
   - Add a `--confront-out <path>` JSON export helper in `cmd/cgo` / `pkg/probe/bpf` that records exact side-by-side `ebpf` vs `tc -s` counters into a structured JSON artifact (used in Chunk 6 to satisfy `docs/VERIFY.md:43`).

## 3. Fail-Closed Topology, Unified RTT Halving & `dualns` Default
Files: `pkg/campagne/topology.go:34-53`, `pkg/campagne/dualns.go:64-71`, `kit/cgo-bench-dualns.sh:48-51,232-238`, `kit/run-dualns.sh:76-78`, `cmd/cgo/server.go:180-186`, `docs/verification/2026-11-15-bench-vm/README.md:95`
1. **Fail-Closed `CheckBenchTopology` (`pkg/campagne/topology.go:34-53`):**
   - If `bench-topology.json` is missing (`os.IsNotExist`), return a clear error (NEVER `return nil`!).
   - Write and validate all required fields: `shape_mode`, `rtt_mode` (`"per_hop_half"`), `delay_ms`, `rate_mbps`, and `ifb_active` (must match campaign direction!).
2. **Single RTT Convention (`DelayMs / 2` per hop everywhere):**
   - Profile `DelayMs` = total path RTT (e.g., P2 = `100 ms`).
   - Fix `kit/cgo-bench-dualns.sh:48-51, 232-238` so `w-edge` and `c-wan` each apply `DELAY_MS / 2` one-way (matching `pkg/campagne/dualns.go:64-71`).
   - Parameterize `kit/run-dualns.sh:76-78` cleanly from CLI args/profile instead of hardcoding `20/100/0`, and export `CGO_SHAPE_MODE` + write `bench-topology.json` consistently.
   - Correct `docs/verification/2026-11-15-bench-vm/README.md:95` so the documented path math matches `50 ms + 50 ms = 100 ms` RTT (and note any historical 200 ms run explicitly as pre-fix).
3. **Default to `dualns` (`cmd/cgo/server.go:180-186`):**
   - Make `dualns` the default bench mode (matching the enterprise rebuild design spec), keeping `legacy` only behind an explicit opt-in `--legacy-bench` flag.

## 4. Probe Order, `SmallObject` Status Check & Atomic Ordered Writes
Files: `pkg/probe/probe.go`, `cmd/cgo/main.go:254-300,431-482`, `api/server_audit.go:196-231`
1. Verify `SmallObject` rejects any HTTP status `< 200 || >= 300` as an error (never timing 404/500 responses) and `ProbePing` preserves chronological RTT slice order (sorting only a copy during percentile calculation).
2. Implement `WriteFileAtomic(path, data, perm)` (temp file in same dir → `f.Sync()` → `Close()` → `os.Rename()` → sync parent dir) + cross-process `flock` for CSV appends (`cmd/cgo/main.go:431-482`), and commit multi-file audit results strictly in order (`telemetry.csv` → `report.json` → `manifest.json` written last, so a crash never leaves a manifest pointing to missing files).

## Constraints
- Add unit tests proving:
  1. Failed `PollStats` returns `DropsMeasured=false` / error (NEVER `0` drops).
  2. `/api/integrity` detects missing files (simulating deleted `valuation.json`), hash mismatches, and malformed CSV rows.
  3. `CheckBenchTopology` fails when `bench-topology.json` is missing or `rtt_mode != per_hop_half` or `shape_mode` mismatches `direction`.
  4. `WriteFileAtomic` and ordered commit handle partial/simulated write errors cleanly.
- Commit all Chunk 2 changes in ONE clean commit: `fix(engine): enforce non-zero absence, real integrity checks, fail-closed topology and atomic writes`.

PART B — Your Hard Verification Gate (Run Before Chunk 3)

Run these commands locally after the agent finishes Chunk 2:

# 1. Run package tests for campagne, qdisc, probe, results, audit, api
go test -v -count=1 ./pkg/campagne/... ./pkg/qdisc/... ./pkg/probe/... ./pkg/results/... ./pkg/audit/... ./api/...

# 2. Verify SumDrops(nil) can no longer silently return 0 in prod/stats
git grep -n "SumDrops" pkg/

# 3. Verify CheckBenchTopology no longer returns nil on missing file
git grep -n -A 8 "func CheckBenchTopology" pkg/campagne/topology.go

# 4. Verify verifyManifests is actually called by /api/integrity
git grep -n "verifyManifests" api/

# 5. Verify no committed manifest.json points to a missing file
python -c "
import json, glob, os, hashlib
bad = []
for mf in glob.glob('docs/**/manifest.json', recursive=True) + glob.glob('data/**/manifest.json', recursive=True):
    d = os.path.dirname(mf)
    m = json.load(open(mf))
    for item in m.get('files', []):
        p = os.path.join(d, item['path'] if isinstance(item, dict) else item)
        if not os.path.exists(p):
            bad.append((mf, p))
print('MISSING_FILES:', bad)
assert not bad
"

# 6. Verify dualns halving matches in both Go and Bash script
git grep -n "DELAY_MS\|DelayMs" kit/cgo-bench-dualns.sh pkg/campagne/dualns.go

Once every check passes cleanly, say "Proceed with Chunk 3" and we will install
the ultra-fast eBPF Ringbuf → 20 Hz Go SSE Ring → Float64Array TypedMetricRing +
RingBufferHUD pipeline into the clean 2-Mode UI!
