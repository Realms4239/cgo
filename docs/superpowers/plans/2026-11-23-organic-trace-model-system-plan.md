# Organic Trace and Model System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** capture a fresh bounded 4-hour passive eBPF trace on sadcnwp, turn its profile bands into sealed `.dist` calibrations, and A/B validate each against a scalar control.

**Architecture:** three sequential tasks. Task 1 is a guarded R2 capture and retrieval whose outputs are hashed evidence under `docs/verification/2026-11-23-organic-trace/`. Task 2 is a new root-module Go command that reads the sealed `flows.csv`, buckets `srtt_us` into profile bands and writes per-profile `telemetry.csv`; the existing `cgo calibrate` then seals `.dist` files. Task 3 is a guarded live A/B replay that reuses the three-state fidelity verdict and records honest divergence. No corpus, gate, permit, ledger or schema edit.

**Tech Stack:** Go 1.25 root module (`cmd/tracemodel`); existing `cgo calibrate`; host-side `cgo-observe-x` with prebuilt `bpf_probe.o`; `C:\WEBSERVER/safe_read.py` and `safe_write.py` for every host contact.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-11-23-organic-trace-model-system-design.md`.
- P1-P4, L1, V1, frozen corpus, book, thesis, C:\END, gates, permits, ledgers, protected config: never edited.
- No `git add -A`; explicit paths only; no merge; no new dependency VERSION; no runtime clang.
- Capture is 10800 integer SECONDS = 3 h (`--duration`; CLI clamps at 172800), bounded so stop/retrieval fit the same-day token window. Output dir must be NEW (`out2`); existing `out/` and `out-readable/` are never modified.
- Every host pass needs a same-day task-scoped token, named window, exact rollback line, classify-only rehearsal and second channel. A guard refusal is a finding, not a reason to reword the request.
- Bands with fewer than 30 samples are explicit absence, never synthesised. `basis=measured`, `source_ref` = new trace id, source telemetry sha256 in every sidecar.
- `gate_status` stays measurement-integrity, never an SLA. No fabricated value, no widened threshold, no VERIFY-row promotion.
- Native gates once per task; known `links`/SKIP issues recorded, never edited.

---

### Task 1: Fresh organic capture and retrieval

**Files:**
- Create `C:/WEBSERVER/APPROVED-2026-11-23-organic-capture.md`
- Create `C:/WEBSERVER/commands/write-organic-start.txt`, `write-organic-stop.txt`, `read-organic-state.txt`, `read-organic-retrieve.txt`
- Create `docs/verification/2026-11-23-organic-trace/README.md`

**Interfaces:**
- Consumes: host binary `/home/sadc/labo-cgo/field-trace/cgo-observe-x`, object `bpf_probe.o` sha `8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`.
- Produces: `docs/verification/2026-11-23-organic-trace/flows.csv` + `manifest.json` + evidence/sidecars; trace id `organic-20261123`.

- [ ] **Step 1: R0 state proof.** Run `safe_read.py --file commands/read-observatory-state.txt --run` (already proven 2026-11-23: evidence `cb9c7307…`). Do not use redirection or globs — the read guard refuses both.
- [ ] **Step 2: Author the token** with Rung R2, Window `2026-11-23 19:30-23:59 UTC`, scope limited to the four command files, `sudo_grant: none` (guest root attach via exact staged `sudo -n` verb), rollback `kill <exact PID>; rm -rf /home/sadc/labo-cgo/field-trace/out2`, rehearsal hash from Step 3, second channel, expiry.
- [ ] **Step 3: Write and rehearse command files.** `write-organic-start.txt` contains exactly:
```bash
cd /home/sadc/labo-cgo/field-trace && sudo -n /home/sadc/labo-cgo/field-trace/cgo-observe-x --duration 10800 --interval 5 --exclude-ports 3333,4444,14444,4555 --out ./out2
```
  `write-organic-stop.txt` kills the exact PID recorded at launch. Classify every line with `safe_write.py --check`; require ALLOW and zero REFUSE; record the sha256 of classify output in the token.
- [ ] **Step 4: Launch, record PID, and stop after the named window.** Use `safe_write.py --run --file ...`; keep the evidence file and sha256 sidecar. On any refusal or ambiguous PID, STOP and report BLOCKED; do not kill by pattern.
- [ ] **Step 5: Retrieve and verify.** R0 read of `out2/manifest.json` and `out2/flows.csv`; copy through `safe_write.py` into the repo verification dir; verify object sha, manifest sha and flows sha before use. Write `README.md` with those hashes, the observed srtt bands, and explicit absence for any empty profile band. Commit those paths with `feat(verification): fresh organic eBPF trace`.

### Task 2: Trace-to-model tool and sealed calibrations

**Files:**
- Create `cmd/tracemodel/main.go`, `cmd/tracemodel/bands.go`, `cmd/tracemodel/bands_test.go`
- Create `docs/verification/2026-11-23-organic-trace/calibrations/` outputs (generated)

**Interfaces:**
- Consumes: `flows.csv` header `kind,window_unix,local_label,peer_label,local_port,peer_port,retrans_payload_bytes,retrans_events,qdisc_drop_packets,qdisc_drop_bytes,srtt_us,snd_cwnd,snd_ssthresh,state,samples`.
- Produces: `BucketFlows(rows []FlowRow, floors map[string]float64) map[string][]float64`; `WriteTelemetry(dir, traceID, profile string, samples []float64) (string, error)` returning the telemetry sha256; `ValidateBand(samples []float64, min int) error`.

- [ ] **Step 1: RED oracle.** In `bands_test.go`, assert a row with `srtt_us=839657` goes to P3 and not P1/P2/P4; a row with `srtt_us=150000` goes to P2; 29 samples causes `ValidateBand` to refuse and 30 passes; NaN/negative/zero srtt is skipped, never coerced to zero. Run `go test -count=1 ./cmd/tracemodel/...` -> RED.
- [ ] **Step 2: Implement the bucketer.** `floors` are exact: `P1=20, P2=100, P3=600, P4=40`; upper bounds are 100, 200, +Inf, 60 ms respectively. `BucketFlows` uses half-open bands `[floor, upper)` and skips invalid values. `ValidateBand` refuses `<30` samples with an error naming the profile and count.
- [ ] **Step 3: Implement telemetry writer.** Emit exactly the format `cgo calibrate` reads:
```csv
phase,offset_ms,rtt_ms,small_ms,bulk_bytes
baseline,0,839.657,,
baseline,5000,839.657,,
```
  one row per sample with a 5000 ms offset; write to `<auditRoot>/data/audits/<traceID>-<profile>/telemetry.csv`; return sha256 of the bytes written.
- [ ] **Step 3b: Wire the CLI in `main.go` (no extra dependency).**
```go
var (
    flowsPath = flag.String("flows", "", "sealed flows.csv to read")
    outDir    = flag.String("out", "", "verification directory for audit tree")
    traceID   = flag.String("trace-id", "", "trace identifier for provenance")
)
flag.Parse()
if *flowsPath == "" || *outDir == "" || *traceID == "" {
    fmt.Fprintln(os.Stderr, "tracemodel: --flows, --out and --trace-id are required")
    os.Exit(2)
}
```
  A band that fails `ValidateBand` writes no directory and the tool prints an explicit `absent: <profile> has N samples (< 30)` line, then exits 0 for the remaining present bands.
- [ ] **Step 4: GREEN and build.** `go test -count=1 ./cmd/tracemodel/...`, `go build ./...` exit 0. Commit `cmd/tracemodel/` with `feat(tracemodel): profile bands from sealed srtt samples`.
- [ ] **Step 5: Produce sealed calibrations.** For each present band, run:
```bash
go run ./cmd/tracemodel --flows docs/verification/2026-11-23-organic-trace/flows.csv --out docs/verification/2026-11-23-organic-trace --trace-id organic-20261123
go run ./cmd/cgo calibrate --audit organic-20261123-P3 --profile P3 --data docs/verification/2026-11-23-organic-trace --out docs/verification/2026-11-23-organic-trace/calibrations
```
  Repeat for each present profile. Verify each sidecar has `basis=measured`, `source_ref=organic-20261123`, sample count and matching `.dist` sha. Absent bands get a written note, not a `.dist`. Commit the generated files with explicit paths.

### Task 3: Live A/B validation gate

**Files:**
- Create `C:/WEBSERVER/APPROVED-2026-11-23-organic-ab.md`, command files `write-organicab-{deploy,launch,poll,result}.txt`
- Create `docs/verification/2026-11-23-organic-trace/replay/README.md`

**Interfaces:**
- Consumes: Task 2's sealed `.dist` files with verified shas; existing guest launchers pattern from `commands/eb-physval.stage` and `commands/fieldcontrol.stage`.
- Produces: per-profile dist vs scalar rows, verdict `trust|bounded|discard`, evidence hashes.

- [ ] **Step 1: Token and rehearsal.** Same-day R2 token, window `2026-11-23 20:00-23:59 UTC`, exact rollback (stop PID, restore topology with `run-dualns.sh up --ingress off`, remove only staged files), classify-only rehearsal with zero REFUSE, hash recorded in token.
- [ ] **Step 2: Stage and verify.** sftp-put each new `.dist`/`.json` to sadc, then explicit `scp` to the guest (avoid the known wrong-host path bug), copy into `/home/bench/dist` with `sudo -n cp`, and verify every sha256 guest-side.
- [ ] **Step 3: Run one dist campaign and one scalar campaign sequentially** (never overlap; the launcher's single-writer guard must not be bypassed). Use `--profiles` with only the present bands, one rep each, same qdisc/CC/direction/flows as earlier replays.
- [ ] **Step 4: Gate.** For each profile, compare rtt p50/p95 to the band's measured p50/p95 with a declared epsilon; reuse `metrics.FidelityVerdict`; fewer than 30 samples yields `bounded`. Report small-object p95, drops and goodput divergence alongside. Any guard refusal or ambiguous value stops as BLOCKED with exact stderr.
- [ ] **Step 5: Record.** Write `replay/README.md` with the table, verdicts, evidence hashes and explicit statement that bench-derived or organic provenance differs by profile. Restore topology; commit the doc with explicit path.

## Self-Review

- Spec §2 -> Task 1; spec §3 -> Tasks 2-3; spec §4 -> RED oracles in Tasks 2-3; spec §5 parks untouched.
- No placeholder: exact paths, flags, headers, bands, commands and commit messages given.
- Type consistency: `BucketFlows`, `WriteTelemetry`, `ValidateBand` defined once in Task 2 and used only there; `source_ref=organic-20261123` identical in Task 2 and Task 3.
- No schema bump, no dependency, no gate/ledger/corpus edit; absence is explicit; no fabricated sample or verdict.
