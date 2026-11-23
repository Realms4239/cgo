# Measurement Physics and Observation Truth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** make the observation source explicit and fail-closed everywhere, add a ring-buffer event lane with an out-of-band loss counter, enrich profile parameters with provenance, and add configured-vs-realised physics metrics with a three-state fidelity verdict.

**Architecture:** four sequential tasks. Task 1 makes `probe.KernelSourceInfo` and `audit.KernelTelemetry` carry an explicit observation block and adds a read-only `preflight` verb; the seal refuses a non-eBPF source. Task 2 adds a ring-buffer event lane behind an injectable reader interface with a reserve-failure counter OUTSIDE the saturated stream. Task 3 adds optional profile fields with byte-identical defaults. Task 4 adds the delay model and fidelity verdict in the pure `metrics` module. No new dependency.

**Tech Stack:** Go modules `probe`, `probe/bpf`, `audit`, `api`, `model`, `qdisc`, `pkg/campagne`, `cmd/cgo`, `metrics`; cilium/ebpf v0.19.0 pinned in the repo root; operator ruling 2026-11-23 allows the same pinned version to be required by probe/go.mod in Task 2; C BPF object source; Vitest/Playwright only if UI copy changes (not in this plan).

## Global Constraints

- Spec: `docs/superpowers/specs/2026-11-23-measurement-physics-observation-design.md`.
- P1-P4, L1, V1, frozen corpus, book, thesis, C:\END, gates, permits, ledgers, protected config: never edited.
- No `git add -A`; explicit paths only; no merge. No new dependency VERSION; operator ruling 2026-11-23 permits the already-pinned `github.com/cilium/ebpf v0.19.0` require in `probe/go.mod` for Task 2. No runtime clang.
- Absent or zero new fields => byte-identical tc vectors and existing JSON, never a placeholder hash or a fabricated zero.
- Ring-buffer loss is exported OUTSIDE the ring; `reserve_failures > 0` forces the explicit "timeline is a lower bound" note.
- `source != ebpf` => no eBPF claim anywhere; the seal refuses.
- RED first with hostile oracles; gate_status stays a measurement-integrity verdict, never an SLA.
- No live bench in these tasks without a separate same-day token; no VERIFY-row promotion.
- Native gates once per task; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Explicit observation truth (Go only)

**Files:**
- Modify `probe/kernelsource.go`
- Modify `audit/report.go`
- Modify `cmd/cgo/seal.go`
- Modify `cmd/cgo/main.go` (new `preflight` case)
- Test `probe/kernelsource_test.go` (new), `cmd/cgo/seal_test.go`, `cmd/cgo/preflight_test.go` (new)

**Interfaces:**
- Consumes: `bpf.Status(goos, btf, caps, objectErr) (source, reason string)` and `bpf.ObjectSHA256() (string, error)`.
- Produces: `probe.KernelSourceInfo{Source, Reason, ObjectSHA256, ReserveFailures}`; `audit.KernelTelemetry.ReserveFailures`, `.LowerBound`; `cgo preflight` writes that struct as JSON to stdout.

- [ ] **Step 1: RED oracles.** In `probe/kernelsource_test.go`, override `kernelFacts` so `bpf.Status` yields `tcpinfo`: assert `ObjectSHA256 == ""` and `ReserveFailures == 0` (no placeholder hash). With `ebpf` and a real `bpf.ObjectSHA256()` result, assert the hash is populated. In `cmd/cgo/seal_test.go`, add an audit whose `kernel_telemetry.source` is `tcpinfo` and assert the existing refusal plus a new assertion that a missing `object_sha256` with `source=ebpf` refuses. Run `go -C probe test -count=1 ./...` and `go -C cmd/cgo test -count=1 ./...` -> RED; capture output.

- [ ] **Step 2: Implement the observation block.**
```go
// probe/kernelsource.go
type KernelSourceInfo struct {
    Source           string `json:"source"`
    Reason           string `json:"reason,omitempty"`
    ObjectSHA256     string `json:"object_sha256,omitempty"`
    ReserveFailures  uint64 `json:"reserve_failures,omitempty"`
}
func KernelSource() KernelSourceInfo {
    goos, btf, caps, objectErr := kernelFacts()
    source, reason := bpf.Status(goos, btf, caps, objectErr)
    info := KernelSourceInfo{Source: source, Reason: reason}
    if source == "ebpf" {
        if sha, err := bpf.ObjectSHA256(); err == nil { info.ObjectSHA256 = sha }
    }
    return info
}
```
```go
// audit/report.go — add to KernelTelemetry
ReserveFailures uint64 `json:"reserve_failures,omitempty"`
LowerBound      bool   `json:"lower_bound,omitempty"`
```
In `cmd/cgo/seal.go`, immediately after the existing `if kt.Source != "ebpf"` refusal, add:
```go
if kt.ObjectSHA256 == "" {
    return fmt.Errorf("kerneltruth seal: object_sha256 absent pour source ebpf (%s)", auditPath)
}
```

- [ ] **Step 3: Add the preflight verb.** In `cmd/cgo/main.go` beside `case "version":`:
```go
case "preflight":
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    if err := enc.Encode(probe.KernelSource()); err != nil {
        fmt.Fprintln(os.Stderr, "preflight:", err)
        os.Exit(2)
    }
    return
```
Test asserts JSON contains `"source"` and never a non-empty `object_sha256` for a fake tcpinfo source.

- [ ] **Step 4: GREEN + commit.** Run `go -C probe test -count=1 ./...`, `go -C audit test -count=1 ./...`, `go -C cmd/cgo test -count=1 ./...`, `go build ./...`. Commit: `feat(observe): explicit source and preflight proof`.

### Task 2: Ring-buffer lane with out-of-band loss counter

**Files:**
- Modify `probe/bpf/src/bpf_probe.c`
- Modify `probe/bpf/types.go`
- Modify `probe/bpf/loader_linux.go`
- Test `probe/bpf/ringbuf_test.go` (new)

**Interfaces:**
- Consumes: `bpf.Globals` and `bpf.Summary` from Task 1's module.
- Produces: `type RingReader interface { Read() ([]byte, error); Close() error }`; `func DrainEvents(r RingReader, g *Globals) (int, error)`; `Globals.ReserveFailures uint64`.

- [ ] **Step 1: RED oracle.** `ringbuf_test.go` uses a fake `RingReader` returning two events then `io.EOF`; assert `DrainEvents` returns 2 and does not alter `ReserveFailures`. A second fake returns an error on the second read: assert the error propagates and the count already read is returned. Run `go -C probe/bpf test -count=1 ./...` -> RED.

- [ ] **Step 2: Add the C counter and ring map.**
```c
struct global_stat {
    __u64 retrans_payload;
    __u64 retrans_events;
    __u64 drop_packets;
    __u64 drop_bytes;
    __u64 map_full;
    __u64 reserve_failures; /* hors du ring : preuve de perte d'observateur */
};
struct cgo_event { __u32 kind; __u32 pad; __u64 ts_ns; __u64 payload; };
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 20);
} events SEC(".maps");
```
In each of `on_retransmit`, `tcp_rcv_established` and `kfree_skb`, after the existing aggregate update, reserve and submit:
```c
struct cgo_event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
if (!e) { struct global_stat *g = globals_get(); if (g) __sync_fetch_and_add(&g->reserve_failures, 1); }
else { e->kind = 1; e->ts_ns = bpf_ktime_get_ns(); e->payload = payload_len; bpf_ringbuf_submit(e, 0); }
```
Do not remove the existing HASH/PERCPU_ARRAY aggregate maps; the ring is additive.

- [ ] **Step 3: Implement the Go drain behind an interface.**
```go
// probe/bpf/ring.go
type RingReader interface { Read() ([]byte, error); Close() error }
func DrainEvents(r RingReader, g *Globals) (int, error) {
    n := 0
    for { if _, err := r.Read(); err != nil { if errors.Is(err, io.EOF) { return n, nil }; return n, err }; n++ }
}
```
In `loader_linux.go`, construct a real ringbuf reader from the loaded map and call `DrainEvents`; add `ReserveFailures` to `Globals` and to `Summary` with `json:"reserve_failures,omitempty"`. The audit bridge sets `KernelTelemetry.ReserveFailures` and sets `LowerBound = failures > 0`; the report note is emitted from that flag.

- [ ] **Step 4: GREEN + commit.** Run `go -C probe/bpf test -count=1 ./...`, `go -C probe test -count=1 ./...`, `go -C audit test -count=1 ./...`, `go build ./...`. Commit: `feat(bpf): ring-buffer events with out-of-band loss counter`.

### Task 3: Profile parameters with provenance

**Files:**
- Modify `model/types.go`, `model/profile_params_test.go` (new)
- Modify `qdisc/ge.go`, `qdisc/ge_test.go`
- Modify `pkg/campagne/dualns.go`, `pkg/campagne/dualns_test.go`
- Modify `api/server.go` (CSV trailing columns + DTO)

**Interfaces:**
- Consumes: `model.ValidateJitterCorrelation`, `model.ValidateSerial`, `qdisc.ApplyNetemPhysics`.
- Produces: `Profile.DelayDistribution`, `.JitterShape`, `.LossModel`, `.LossP/LossR/LossBad/LossGood`, `.Basis`, `.SourceRef`; `ValidateProfileParams(p Profile) error`.

- [ ] **Step 1: RED oracles.** Assert: absent fields leave existing tc vectors byte-identical; `basis=measured` with empty `SourceRef` refuses; `JitterShape` with `uniform` refuses; out-of-range GE fractions refuse before any `FakeRunner` call; CSV rows with <=12 columns keep all new fields zero. Run `go -C model test -count=1 ./...` and `go -C qdisc test -count=1 ./...` -> RED.

- [ ] **Step 2: Add fields and validator.**
```go
DelayDistribution string  `json:"delay_distribution,omitempty"`
JitterShape       float64 `json:"jitter_shape,omitempty"`
LossModel         string  `json:"loss_model,omitempty"` // ""|uniform|gemodel
LossP             float64 `json:"loss_p,omitempty"`
LossR             float64 `json:"loss_r,omitempty"`
LossBad           float64 `json:"loss_bad,omitempty"`
LossGood          float64 `json:"loss_good,omitempty"`
Basis             string  `json:"basis,omitempty"` // ""|measured|standard|declared
SourceRef         string  `json:"source_ref,omitempty"`
```
`ValidateProfileParams` enforces: distribution in the closed set; `JitterShape > 0` unless distribution is `uniform` or empty; all four GE fractions in [0,1] when `loss_model=gemodel`; `basis` in the closed set and `SourceRef != ""` when `basis=measured`. Reject NaN/Inf.

- [ ] **Step 3: Wire the emitter.** In `qdisc/applyNetemArgs`, emit `distribution <name>` only when `DelayDistribution != ""` and append the shape token with the documented type-specific meaning; emit `loss gemodel` from the profile when `LossModel == "gemodel"` instead of the campaign GE only when the campaign did not supply one. In `pkg/campagne.applyPhysics`, call `model.ValidateProfileParams(prof)` before the first topology check or tc mutation. Campaign `channel_model` still overrides when explicitly set.

- [ ] **Step 4: CSV + DTO.** In `api/server.go`, parse optional columns 13-20 in order (distribution, shape, loss_model, loss_p, loss_r, loss_bad, loss_good, basis, source_ref would be 9 - use columns 13-21 with `source_ref` last); add the fields to the `/api/profiles` DTO with `omitempty`. Update only shape-enumerating assertions.

- [ ] **Step 5: GREEN + commit.** Run model/qdisc/api/campagne suites and `go build ./...`; commit `feat(profile): validated distribution, GE loss and provenance fields`.

### Task 4: Physics and fidelity metrics

**Files:**
- Create `metrics/physics.go`, `metrics/physics_test.go`
- Modify `metrics/verdict.go`, `metrics/verdict_test.go`
- Modify `api/server.go` only if a report field is added (exact path confirmed during exploration)

**Interfaces:**
- Consumes: `model.Profile` fields from Task 3; `audit.Report` configured values.
- Produces: `ModelDelayMs(propagationMs, packetBytes, rateMbps, queueMs float64) float64`; `FidelityVerdict{Samples int, DeviationMs, EpsilonMs float64, State string}` with state `trust|bounded|discard`.

- [ ] **Step 1: RED oracles.** Assert `ModelDelayMs(100, 1500, 20, 0)` equals the arithmetic `100 + (1500*8/20e6*1000) + 0` within 1e-9; zero rate refuses; a sample set with deviation > epsilon yields `discard`; fewer than 30 samples yields `bounded` never `trust`; no input produces `trust` with an explicit validity bound missing. Run `go -C metrics test -count=1 ./...` -> RED.

- [ ] **Step 2: Implement.**
```go
func ModelDelayMs(propagationMs, packetBytes, rateMbps, queueMs float64) (float64, error) {
    if rateMbps <= 0 || math.IsNaN(rateMbps) { return 0, fmt.Errorf("rate must be > 0") }
    return propagationMs + packetBytes*8/(rateMbps*1e6)*1000 + queueMs, nil
}
```
`FidelityVerdict` computes deviation from measured-vs-model delay samples, requires >=30 samples for `trust`, returns `bounded` when samples are insufficient and `discard` when deviation > epsilon. Add configured-vs-realised rows struct `{Metric, Configured, Realised string}` populated only from actual report values; absent becomes `"—"` with an explicit note.

- [ ] **Step 3: GREEN + commit.** Run `go -C metrics test -count=1 ./...`, `go -C api test -count=1 ./...`, `go build ./...`; commit `feat(metrics): delay model and three-state fidelity verdict`.

## Self-Review

- Spec §1 -> Tasks 1-2; §2 -> Task 3; §3 -> Task 4; named park (object vendoring) stays out of scope with the human gate intact.
- No placeholder: every step names exact files, code and commands. No "TBD", no "similar to Task N".
- Type consistency: `KernelSourceInfo.ObjectSHA256` -> `KernelTelemetry.ObjectSHA256` -> seal check; `Globals.ReserveFailures` -> `Summary.ReserveFailures` -> `KernelTelemetry.ReserveFailures` -> `LowerBound`; `Profile.LossModel` -> `qdisc` emitter -> `campagne` validator; `FidelityVerdict.State` uses the closed set trust|bounded|discard.
- No gate relaxation, no fabricated value, no schema version bump, no dependency added.
