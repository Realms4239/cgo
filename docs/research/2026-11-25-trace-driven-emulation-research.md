# last30days research - trace-driven emulation, netem distributions, CAKE AQM and deadline SLA (2026-11-25)

Engine: last30days v3.27.2, run 2026-11-25, window 2026-10-26 to 2026-11-25.
Engine result: exit 0, 82.1 s, 3 active sources (GitHub, Hacker News, Reddit), 20 Reddit threads, 10 HN stories, 3 GitHub items, 0 YouTube videos, 0 Polymarket markets.
Raw artifact: `C:/Users/ASUS/Documents/Last30Days/trace-driven-network-emulation-with-netem-delay-distributions-and-cake-aqm-for-deadline-sla-raw-v3.md` (with this run's WebSearch Supplement Results appended).

## Honest coverage finding

The ranked social clusters were all off-topic: "trace" matched trace fossils, vulnerability tracing, LLM call tracing and schematic annotation. None of the 33 social items concerned network emulation, netem, CAKE or deadline SLA. Per the skill's nothing-solid floor, these items are not used as evidence. The useful grounding comes from the engine's web supplements and the post-engine WebSearch pass.

## What the evidence actually supports

**Trace-driven emulation is a live research method, not a novelty.** Recent work uses measured channel conditions directly at the traffic-control layer: an eBPF-based satellite emulation method feeds real delay/loss traces into netem (arXiv, "An eBPF-Based Trace-Driven Emulation Method for Satellite Networks"), and CloudEmu replays field-measured user-plane traces through Linux emulator nodes (arXiv, CloudEmu). CellTwin goes further, learning 5G scheduler behaviour from high-resolution in-situ measurement (arXiv, CellTwin). The design direction is the same as this project's `cmd/tracemodel` to `cgo calibrate` to `.dist` path: measured distribution, not a hand-written scalar.

**netem has hard physical limits that no config can remove.** The tc-netem manual states timer granularity and clock interrupts bound delay and rate accuracy, and mixing reordering forms can produce unexpected results. Practical write-ups add a second constraint: netem's default RB-tree limit of 1000 queued packets caps how much delayed traffic it can hold under load (linkem, GitHub; loadtest.app). This matters for a 600 ms RTT VSAT profile, where a 1000-packet budget is reached quickly. These limits are properties of the kernel mechanism, not defects in CGO's gate logic.

**CAKE should not be stacked under netem.** The Bufferbloat CAKE mailing list records the same lesson this repo learned in 2026-11: a child qdisc hidden behind netem's internal FIFO makes CAKE's RTT estimator see queueing it did not create. CAKE's own layers - shaper, priority queue, flow isolation, AQM, packet management (CakeTechnical, bufferbloat.net) - must sit on a separate interface from link physics. This confirms the dual-namespace topology in `kit/cgo-bench-dualns.sh` is the correct architecture; recent 2026 kernel work on qdisc backlog accounting (CAKE list, 2026-11-08) reinforces that backlog correctness is still being actively fixed upstream.

**A 220 ms deadline is a policy, not physics.** Public 2026 measurements put Madagascar's median latency around 232 ms (SpeedOf.me) and operator latency in rapid change (nPerf Madagascar mobile barometer, operator comparisons). A fixed 220 ms objective cannot be a cross-profile SLA: on a 600 ms RTT VSAT profile the propagation delay alone exceeds the deadline, so a 0% result is deterministic. The WIS 2.0 frame has no numeric per-bulletin threshold; CGO's 220 ms is a DGM operator objective, correctly documented as such in `metrics/cost.go`.

**Real field data exists but is not yet calibratable.** The 2026-11-23 organic eBPF trace has 109,550 rows, yet banding against the ratified floors yields P1 0, P2 0, P3 3, P4 11 samples - all below the 30-sample floor. No `.dist` was produced. The blocker is traffic composition (host-local ~0.013 ms and a 200-600 ms gap), not the capture tool or the floor.

## Enhancement implications for this repository

1. Keep link physics and AQM on separate interfaces; do not regress to a stacked qdisc. The kernel and community evidence both support the current split.
2. When reporting `deadline_ok_pct`, pair it with the profile's propagation floor. A 0% P3 result is an expected consequence of `600 ms RTT > 220 ms deadline`; it must not be read as a CAKE or BBR failure, and it is not a gate (the eight gate names in `pkg/campagne/campagne.go:698-702` contain no deadline gate).
3. A real Malagasy trace needs a passive capture whose SRTT samples actually land in a profile band with at least 30 samples. The current organic trace does not qualify; lowering the floor or inventing samples would break the evidence chain.
4. Distribution replay via `.dist` is the right mechanism once the sample floor is met; netem's timer and 1000-packet limits must be recorded as measurement caveats, not tuned away.

## Engine footer

✅ All agents reported back!
├─ 🟠 Reddit: 20 threads │ 298 upvotes │ 192 comments
├─ 🟡 HN: 10 storys │ 84 points │ 32 comments
├─ 🐙 GitHub: 3 items │ 75 comments
├─ 🗣️ Top voices: r/networking, r/mikrotik, r/homelab
└─ 📎 Raw results saved to ~/Documents/Last30Days/trace-driven-network-emulation-with-netem-delay-distributions-and-cake-aqm-for-deadline-sla-raw-v3.md
