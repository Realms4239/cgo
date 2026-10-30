# Trace-driven methodology, kernel telemetry and the BBRv1 → BBRv3 comparison

*Chapter draft, 2026-10-30. Every number below is either a `docs/VERIFY.md`
row, a value rendered from the frozen corpus and checked by
`book/tools/audit-nombres.py`, or explicitly marked `non mesuré` / `bloqué`.
The chapter introduces no number of its own.*

## 1. What the bench is, and what it refuses to be

The bench is a four-namespace Linux topology (`ip netns` + `veth` + `netem` +
`tc`) on which four link classes P1–P4 are replayed, with the file that the
ISP actually applies (`pfifo_fast`, `fq_codel`, CAKE) and the congestion
control the sender actually runs (`cubic`, `bbr`, and later `bbr3`).

It refuses three shortcuts:

- **No synthetic delay parameters dressed as measurements.** The delay model is
  either a scalar sized from the measured RTT or an empirical `.dist` built
  from a field trace (§2). When neither exists, the report says so.
- **No userspace timing claimed as a kernel fact.** Retransmission counters come
  from TCP_INFO until the eBPF probe is attached; the report carries which tier
  answered and why (§4).
- **No transfer of bench results to the production link.** The bench reproduces
  a regime; it does not measure the operator's link.

## 2. Trace-driven delay (`netem ... distribution`)

What is built: a strict telemetry reader, a baseline-only quantile table, and a
sealed native netem distribution file plus its sidecar
(`pkg/qdisc/dist.go`, `cmd/cgo/calibrate.go`). The campaign loads it only when
`CGO_DELAY_MODEL=dist`, and refuses a malformed or orphaned `.dist`.

What a usable field trace must contain — and does not exist today: at least
**30 baseline samples** from a real link, taken before any load, so the
distribution describes path delay rather than queue delay. With no such trace,
no `.dist` is cited in this thesis; the scalar model stands and the limitation
is stated.

## 3. Wireless loss as a two-state Markov chain (Gilbert–Elliott)

Uniform random loss cannot represent tropical rain attenuation, which arrives
in bursts. The bench uses Linux netem's `gemodel` form:

```text
tc qdisc add dev <veth-wan> root netem loss gemodel 1% 20% 90% 5%
```

Four parameters: the loss probability in the good state, the loss probability
in the bad state, and the two transition probabilities that govern how long
each state lasts. The implementation and its mean-loss expression live in
`pkg/qdisc/ge.go`; the campaign selects it with `channel_model=bursty`, which
the engine accepts only for P2 and P3 and refuses elsewhere, naming the
offending profiles.

This is a model, not a measurement: it reproduces a plausible temporal
structure of loss, and every cell that uses it says `channel_model: bursty` in
its provenance.

## 4. Kernel ground truth on Linux 6.8 — and the honest detour

The master prompt asked for the `qdisc/qdisc_drop` tracepoint. **It does not
exist in Linux 6.8.** The 6.8 source defines only `qdisc_dequeue`,
`qdisc_reset` and `qdisc_destroy` in `include/trace/events/qdisc.h`; the
dedicated drop tracepoint arrived later (first shipped in v7.1). Moreover,
`__qdisc_drop` is a `static inline` helper, inlined into `sch_cake.c` and
`sch_fq_codel.c`, so it has no symbol a kprobe or fentry could attach to.

On Ubuntu 24.04's 6.8 kernel, every qdisc discard reaches
`kfree_skb_reason(skb, SKB_DROP_REASON_QDISC_DROP)` inside
`net/sched/sch_generic.c`. The probe therefore attaches to
`tracepoint:skb:kfree_skb` and filters strictly on
`SKB_DROP_REASON_QDISC_DROP`, which isolates queue discards from firewall,
routing and socket noise without a kernel rebuild.

Exact retransmitted bytes need a second hook: the `tcp_retransmit_skb`
tracepoint carries only `skbaddr` and `skaddr`, no length. The design reads the
socket and skb through fentry (with kprobe as the fallback) to obtain the
payload length, and counts bytes rather than assuming a 1,448-byte segment.

**Status of this paragraph's subject:** the CO-RE source
(`pkg/probe/bpf/bpf_probe.c`), its build instructions and the `make bpf`
target are committed; the object has never been built, because this host
cannot build it and the Linux bench has not been available. The Go contract,
the fallback ladder, the report block and the badge are implemented and tested
(slice 17, `VERIFY` row “eBPF Go half”); nothing has been attached. The
ground-truth row stays `BLOCKED` until an audit runs on the bench with
`source: ebpf`.

The fallback ladder, in order: non-Linux host → TCP_INFO; object absent →
TCP_INFO with that reason; no BTF → TCP_INFO; no `CAP_BPF`/`CAP_PERFMON` →
TCP_INFO; fentry rejected → kprobe; any partial window → the whole window
falls back to TCP_INFO with its reason. A fallback is never silent and never
mixed into one ratio.

## 5. BBRv1, BBRv3 and CAKE — what is measured, what is not

What the frozen corpus measured is the **legacy regime**: P2 cellular,
`direction=up`, BBRv1, one flow. In that regime, at a 220 ms deadline:

| Discipline | CC | small p95 (ms) | Deadline met (%) | QDI (ms) | Bulk (Mb/s) | Status |
|---|---|---|---|---|---|---|
| `pfifo_fast` | CUBIC | 213.9 | 98.0 | 14.0 | 2.0 | invalid |
| `pfifo_fast` | BBR | 328.2 | 60.2 | 57.4 | 18.2 | valid |
| `fq_codel` | CUBIC | 209.3 | 97.1 | 14.0 | 2.2 | invalid |
| `fq_codel` | BBR | 205.4 | 96.2 | 14.0 | 19.5 | valid |
| CAKE | CUBIC | 211.9 | 95.1 | 13.0 | 2.0 | invalid |
| CAKE | BBR | 210.5 | 98.1 | 14.0 | 18.5 | valid |

Reading rule, which the thesis must state beside the table: the **CUBIC cells
are invalid on G4 and are not a valid comparison** (`VERIFY` row “Tableau 15
CUBIC cells”, `RECORDED`). The comparison the thesis defends is the BBR column
across disciplines: an unmanaged FIFO under BBR fills until latency reaches
328.2 ms and misses 40 % of deadlines, while the active disciplines keep the
queue short — 96–98 % of deadlines met at comparable bulk goodput.

Mechanism, in one paragraph: loss-based CUBIC treats the regime's baseline loss
as congestion and settles at ~2 Mb/s, so the load never establishes; BBR sizes
inflight from bandwidth and RTT and fills the FIFO, which is exactly the
bufferbloat the thesis is about; an active discipline drops or marks the
excess early, keeps the standing queue short, and converts the difference into
deadlines met.

**BBRv3.** The third control is the algorithmic answer to the first column:
loss-responsiveness rather than pure model-following. It is **not measured
here**: BBRv3 ships with a full kernel tree and requires a 7.1-class kernel
build on the bench; `kit/install-bbr3.sh` exists and the matrix accepts `bbr3`
only when `/proc/sys/net/ipv4/tcp_available_congestion_control` lists it. Until
that host exists, every `bbr` in the corpus means BBRv1, and no cell claims
otherwise. The evaluation matrix (P2 × three qdiscs × {cubic, bbr, bbr3} ×
three repetitions) is defined and blocked on the same host.

## 6. Interfaces and control point

The download direction is the harder half: server-side egress shaping cannot
touch traffic that arrives from the ISP. The engine therefore also generates
ingress shaping — IFB on Linux, and on RouterOS a LAN-interface queue tree,
with the explicit note that RouterOS has no IFB and that this is a different
control point, not a byte-equivalent translation. The generated configuration
is machine-checked where it can be: a RouterOS 7 instance in QEMU (CHR)
imports the sealed report's script and exports it back for a field-level diff
(slice 16). On this host that proof is a replay: 12/12 checks against a
committed export fixture, and a sealed artifact at level `windows-tcg`; no CHR
has booted yet. Neither the IFB live row nor the CHR live row is claimed.

## 7. Sources for every number in this chapter

| Claim | Where it comes from |
|---|---|
| 159 runs, 528 lines, 208 valid, 74 degraded, 246 invalid, 22 cells | `VERIFY` row “Frozen corpus…”, `python book/tools/audit-archive.py` |
| P2 comparison table (328.2 / 60.2 / 96.2 / 98.1 / 210.5 / 57.4 / 14.0 / 18.2 / 19.5) | rendered from the frozen corpus, checked by `audit-nombres`; limitation row “Tableau 15 CUBIC cells” |
| CUBIC cells invalid on G4 | same limitation row (`RECORDED`) |
| 220 ms deadline | CAP objective used by the corpus and the audit report; not a WMO mandate |
| Gilbert–Elliott parameters 1 % / 20 % / 90 % / 5 % | `pkg/qdisc/ge.go`, `kit/cgo-bench-dualns.sh` |
| 6.8 hook facts (`qdisc_drop` absent, `__qdisc_drop` inlined, `kfree_skb_reason`) | kernel source review recorded in the session and in the slice-4 plan |
| eBPF status (`BLOCKED`, no object) | `VERIFY` row “eBPF retransmit / RTT / qdisc drops” |
| CHR replay 12/12, no boot | `VERIFY` row “RouterOS 7 configuration compliance via CHR” |
| Enterprise tariffs (1,388.89 Ar/h; 81.38 / 150 / 5,556 Ar/GB) | `VERIFY` row “Enterprise tariff scenarios”, `RECORDED` — contractual, not measured |

## 8. What this chapter cannot say

It cannot say that the production link is congested, that a remediation was
deployed, that BBRv3 behaves like its paper, or that the RouterOS translation
was accepted by a real device. Each of those needs a row this chapter does not
have, and the thesis defends the method — a reproducible twin, an honest
provenance chain, and a decision procedure — rather than a deployed result.
