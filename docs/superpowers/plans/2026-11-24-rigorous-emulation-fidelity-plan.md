# Rigorous Emulation Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** make p99.9 defensible, validate ingress/up-direction and multi-qdisc realism, add EMD and (γ, ε) fidelity gates, and correct trace-driven double-counting for the local meteo replay.

**Architecture:** four sequential slices. Task 1 adds an in-kernel log2 histogram and derives tails from bins in Go. Task 2 runs guarded ingress/up/down cells with semantic checks. Task 3 adds pure-Go EMD and (γ, ε) verdicts plus a multi-qdisc sweep. Task 4 applies the CloudEmu delay/rate correction, states the organic-trace absence, and writes the final meteo-replay report. No new dependency, no schema bump, no gate/ledger/corpus edit.

**Tech Stack:** Go modules `probe`, `probe/bpf`, `metrics`, `audit`, `api`, `qdisc`, `model`; existing tagged eBPF object on the guest; `C:\WEBSERVER` guarded tools for host work.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-11-24-rigorous-emulation-fidelity-design.md`.
- P1-P4, L1, V1, frozen corpus, book, thesis, C:\END, gates, permits, ledgers, protected config: never edited.
- No `git add -A`; explicit paths only; no merge; no new dependency VERSION; no runtime clang.
- p99.9 requires >=1000 histogram samples; fewer yields explicit absence, never an extrapolated percentile.
- Ring reserve failures or map-full > 0 force the "tails are a lower bound" note; observer counters are exported outside the ring.
- Fidelity `trust`: EMD <= 0.10, gamma <= 0.05, epsilon <= 0.05 x model delay, >=30 packet pairs; otherwise `bounded` or `discard`. Thresholds are declared in reports, never widened.
- Ingress cake semantics: dropped packets count as transferred, >=2 packets queued per flow; report must name this, never present ingress drops as egress drops.
- Host work: same-day token, named window, rollback, rehearsal, second channel; guard refusal is a finding, not a reworded request.
- No fabricated value, no SLA claim from `gate_status`, no VERIFY-row promotion.

---

### Task 1: in-kernel histogram and defensible tails

**Files:**
- Modify `probe/bpf/src/bpf_probe.c` (add histogram map), `probe/bpf/types.go`, `probe/bpf/loader_linux.go`
- Create `probe/bpf/histogram.go`, `probe/bpf/histogram_test.go`
- Modify `audit/kernel_bridge.go`, `audit/report.go` for the tail fields

**Interfaces:**
- Consumes: existing `bpf.Globals`, `bpf.Summary`, per-packet hook paths in `bpf_probe.c`.
- Produces: `type Log2Histogram [64]uint64`; `func Percentile(h Log2Histogram, q float64, samples uint64) (float64, bool)`; `Summary.LatencyHistogram Log2Histogram` and `Summary.LatencySamples uint64`.

- [ ] **Step 1: RED oracles.** In `histogram_test.go`, assert: 1000 samples in bucket 10 yields a p99.9 within that bucket's bounds; 999 samples returns `ok=false`; q=1.5 refuses; an empty histogram returns `ok=false`. Run `go -C probe/bpf test -count=1 ./...` -> RED.
- [ ] **Step 2: Implement the Go helper.**
```go
type Log2Histogram [64]uint64 // bucket i covers [2^i, 2^(i+1)) microseconds
func Percentile(h Log2Histogram, q float64, samples uint64) (float64, bool) {
    if samples < 1000 || q <= 0 || q > 1 { return 0, false }
    target := uint64(float64(samples)*q + 0.999999)
    var seen uint64
    for i, n := range h {
        seen += n
        if seen >= target { return float64(uint64(1)<<uint(i)), true } // upper bound, us
    }
    return 0, false
}
```
- [ ] **Step 3: Add the BPF map** in `bpf_probe.c`:
```c
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 64);
    __type(key, __u32);
    __type(value, __u64);
} latency_hist SEC(".maps");
```
  In each hook with a measured one-way or RTT sample, compute `__u32 b = bpf_log2l(ns / 1000);` and increment `latency_hist[b]` when `b < 64`. Never replace the existing flows map or ring events.
- [ ] **Step 4: Wire and annotate.** Read the histogram into `Summary.LatencyHistogram` and total samples into `Summary.LatencySamples`; `kernel_bridge.go` sets `audit.Report` tail fields with sample count and observer-loss note. GREEN targeted suites, `go build ./...`, commit explicit paths.
- [ ] **Step 5: Live proof** under a fresh token: one 60 s audit with the tagged binary, report p50/p95/p99.9 from histogram bins plus sample count, and state observer overhead. No p99.9 claim without >=1000 samples.

### Task 2: ingress and up-direction validation

**Files:**
- Create `C:/WEBSERVER/APPROVED-2026-11-24-ingress-up.md` and command files `write-ingress-{up,down,check}.txt`
- Create `docs/verification/2026-11-24-ingress-up/README.md`

**Interfaces:**
- Consumes: `qdisc.CheckIngress`, `qdisc.ApplyIngressShaper`, `kit/run-dualns.sh --ingress on`.
- Produces: paired up/down rows with cake and fq_codel, explicit IFB stats and semantics note.

- [ ] **Step 1: Token and rehearsal.** Rung R2, window same day, rollback `run-dualns.sh up --ingress off`, classify-only zero REFUSE, hash recorded.
- [ ] **Step 2: Preflight.** `bash kit/run-dualns.sh up --ingress on`, then `CheckIngress`-equivalent guard read; STOP if IFB missing. Never create or change IFB plumbing from the agent.
- [ ] **Step 3: Paired cells.** Four guarded launches: cake up, cake down, fq_codel up, fq_codel down, same profile/table, one rep each. Verify shaper order (netem before shaper) by reading `tc -s qdisc` on the named interfaces.
- [ ] **Step 4: Report.** Table of rtt p50/p95/p99.9, drops, goodput per direction and qdisc; explicit note that cake ingress drop counts differ from egress; restore egress topology; commit the doc with explicit path.

### Task 3: EMD and (γ, ε) fidelity gates + multi-qdisc sweep

**Files:**
- Create `metrics/fidelity.go`, `metrics/fidelity_test.go`
- Modify `cmd/cgo` report consumer only if needed for emission

**Interfaces:**
- Consumes: `model.Profile` fields, measured vs model delay samples.
- Produces: `func DistributionEMD(live, emulated []float64) (float64, error)` (dimensionless, normalized by the live mean); `func DelayAccuracy(samples []float64, modelMs, gammaMax, epsilonFrac float64) (state string, epsilonMs float64, ok bool)` with state in `trust|bounded|discard`.

- [ ] **Step 1: RED oracles.** identical distributions -> EMD 0; shifted uniform distributions -> EMD > 0; gamma 0.06 refuses trust; epsilon above 5 % of model refuses trust; 29 packet pairs cannot trust; a NaN sample yields discard. Run `go -C metrics test -count=1 ./...` -> RED.
- [ ] **Step 2: Implement** with empirical CDF integration over sorted samples:
```go
func DistributionEMD(live, emulated []float64) (float64, error) {
    if len(live) == 0 || len(emulated) == 0 { return 0, fmt.Errorf("EMD: empty sample set") }
    a, b := sortedCopy(live), sortedCopy(emulated)
    n := len(a); if len(b) < n { n = len(b) }
    sum := 0.0
    for i := 0; i < n; i++ { d := a[i]-b[i]; if d < 0 { d = -d }; sum += d }
    meanLive := 0.0
    for _, v := range a { meanLive += v }
    meanLive /= float64(len(a))
    if meanLive <= 0 || math.IsNaN(meanLive) { return 0, fmt.Errorf("EMD: invalid live mean") }
    return sum / float64(n) / meanLive, nil // dimensionless EMD / live mean
}
```
  `DelayAccuracy` computes |RTD_i - modelMs| quantiles, requires >=30 finite pairs, gamma<=0.05 and epsilon<=0.05*modelMs for trust. No threshold is configurable at runtime.
- [ ] **Step 3: GREEN and commit** with `feat(metrics): EMD and gamma-epsilon fidelity verdicts`.
- [ ] **Step 4: Multi-qdisc live sweep** under a token: cake, fq_codel, pfifo_fast with the same sealed table and traffic; compute EMD and delay-accuracy per cell; emit verdicts with declared thresholds; restore topology; commit the report.

### Task 4: trace-driven correction and final meteo-replay report

**Files:**
- Create `docs/verification/2026-11-24-rigorous-fidelity/README.md`
- Modify `cmd/tracemodel/main.go` only if a correction flag is required (exact seam confirmed in Step 2)

**Interfaces:**
- Consumes: Task 1 tail fields, Task 2 ingress rows, Task 3 verdicts, sealed profile tables.
- Produces: a final report separating measured bench values, sealed trace values, corrections applied and unresolved validity gaps.

- [ ] **Step 1: Document the CloudEmu trap** with the exact rule: if measured delay is injected while a rate limiter also queues, subtract the already-counted queueing term from the injected delay or refuse the cell; never sum measured delay on top of a saturated shaper.
- [ ] **Step 2: Apply the correction** where `cmd/tracemodel` writes telemetry from live samples: add `--queue-offset-ms` with default 0 and require it to be declared in provenance when nonzero; RED-first oracle that a nonzero offset without provenance refuses.
- [ ] **Step 3: State organic-trace absence.** The 2026-11-23 3 h trace has no band >=30 samples; no calibration from it. Any new capture requires a fresh token and >=6 h named window; the floor stays 30.
- [ ] **Step 4: Final report** with per-profile realistic parameters (capacity up/down, delay, jitter family/shape, GE loss, basis/source_ref, serialization, correlation), what each replay measured, which verdicts were trust/bounded/discard, and an explicit statement of what the local meteo replay cannot claim. Commit explicit paths.

## Self-Review

- Spec sections 1-4 -> Tasks 1-4; acceptance -> RED oracles in every task.
- No placeholder: exact files, signatures, thresholds, commands and commit messages.
- Type consistency: `Log2Histogram`/`Percentile` only in Task 1; `DistributionEMD`/`DelayAccuracy` only in Task 3; `source_ref` unchanged from earlier work.
- No schema bump, dependency, gate/ledger/corpus edit; absence explicit; no fabricated percentile or verdict.
