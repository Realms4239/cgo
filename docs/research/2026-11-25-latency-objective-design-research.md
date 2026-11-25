# last30days research - latency objective design for heterogeneous links (2026-11-25)

Engine: last30days v3.27.2. Two runs were executed. The first used the wording "fixed versus baseline-relative" and was auto-parsed as comparison mode with thin, off-topic evidence; it is recorded here for honesty but is not used as evidence. The second used "latency objective design for heterogeneous network links", GENERAL mode, exit 0, with 15 Reddit threads, 20 HN stories, 21 GitHub items and 0 Polymarket markets. YouTube was unavailable (yt-dlp search timeout in this harness).

Raw artifact: `C:/Users/ASUS/Documents/Last30Days/latency-objective-design-for-heterogeneous-network-links-raw-v3.md`, with WebSearch Supplemental Results appended.

## Coverage finding

The social clusters of the second run were software-latency bugs (db latency, crawl latency, 3D viewer latency), not network SLA design. Per the nothing-solid floor they are not used as evidence. Grounding comes from the pre-research and post-engine WebSearch passes.

## What the evidence supports

**Define the indicator before the objective.** SRE guidance says to define the SLI first and then put an upper bound on it; a Google Cloud SLO document defines a latency SLI as the ratio of calls below a threshold to all calls, and Honeycomb distinguishes the success indicator from the allowed failure budget. ClickHouse engineering adds that the target should come from measured past performance rather than aspiration, with a fixed rolling window. This repo already has the indicator (`DeadlineOKPct`, share of completions under a due value); the weakness is that the objective (220 ms) was not derived from the heterogeneous links it is applied to.

**Baseline RTT is the standard reference point.** RFC 6349 defines baseline RTT as the inherent, non-congested round-trip time of a path and requires measuring it before a throughput test. Viavi's RFC 6349 application notes repeat that baseline RTT and bottleneck bandwidth are prerequisites. That is a direct precedent for expressing a latency objective relative to the link's own propagation floor rather than as one absolute number.

**Heterogeneous paths need relative delay terms.** IEEE work on relative forward delay distinguishes the forward link's condition from the whole path, and a separate end-to-end delay distribution study decomposes delay into radio-link and core components. LEO and TSN work similarly sum per-node propagation and residence terms. This is the same decomposition this repo already encodes in `metrics/physics.go` via `ModelDelayMs = propagation + serialization + queue`.

**WIS 2.0 does not supply the 220 ms number.** WMO material describes "reliable and low latency access" and a global cache, but the repo's own `metrics/cost.go:75-82` states explicitly that WIS 2.0 gives no numeric per-bulletin threshold and that 220 ms is a DGM operator objective. Treating it as an WMO requirement is a category error.

## Better approach supported by the evidence

1. Keep the fixed 220 ms DGM objective for the operational CAP transaction, but label it as policy and keep it in the audit path (`CAPSLA`) where it is already correctly scoped.
2. For heterogeneous campaign profiles, derive the due value from the profile's propagation floor plus a bounded queue allowance: `due = prof.DelayMs + small-packet serialization + q`, where `q` is an explicit, declared allowance rather than an invented number. `metrics.ModelDelayMs` already implements the arithmetic and refuses invalid rates.
3. Make absence explicit: `metrics.AbsentValue = "—"` already exists for missing report cells. `DeadlineOKPct([]) == 0` conflates absence with a measured failure; the audit path guards it, the campaign path does not.
4. Do not weaken any gate. Deadline is not one of the eight gates in `pkg/campagne/campagne.go:698-702`; it is a reported metric. Adding or removing a gate is a human decision.
5. All of the above change frozen CSV semantics and therefore need an operator ruling before implementation. No code was changed by this research.

## Engine footer

✅ All agents reported back!
├─ 🟠 Reddit: 15 threads │ 45 upvotes │ 85 comments
├─ 🟡 HN: 20 storys │ 3,677 points │ 1,443 comments
├─ 🐙 GitHub: 21 items │ 17 reactions │ 4,406 comments
├─ 🗣️ Top voices: r/networking, r/homelab
└─ 📎 Raw results saved to ~/Documents/Last30Days/latency-objective-design-for-heterogeneous-network-links-raw-v3.md
