# Profile realism, eBPF-primary measurement and interface honesty — design (2026-11-23)

Authority: operator ordered extensive interface improvements plus deeper bench/kernel-truth research and "actually edit and improve" the qdisc/CC profiles so they reflect real life, with eBPF as the major main measurement and `tc` as cross-check. SDD required. Research: `last30days` engine + supplements, saved under `C:/Users/ASUS/Documents/Last30Days/`; social clusters were "nothing solid", so only the citable web supplement set is used. No frozen corpus, ledger, gate, dependency, accent or new motion.

## 1. Research-grounded profiles (add, never rewrite P1-P4)

Citable anchors from the supplements:
- LEO: operator median ~26 ms and <1% above 55 ms (starlink.com), independent 20-100 ms (IIUM 2025), ITU LEO trials <60 ms at >=50 Mbit/s (itu.int), Ookla Q3 2025 country median up to 282 ms (ookla.com), down >200 / up >30 Mbps (starlink.com).
- GEO/VSAT: measured 600-3000 ms, intrinsic ~480 ms, C-band 580-720 / Ku 550-700 ms (IIUM 2025, IJSRA 2025), ITU Kenya VSAT <=750 ms at >=5 Mbit/s (itu.int).

Add **new** profiles with explicit provenance, keeping P1-P4 byte-identical so the frozen corpus and existing E2E oracles stay comparable:
- `L1` (LEO-real): capacity 200 down / 30 up, delay 26 ms, jitter 15 ms, loss 0.1%.
- `V1` (VSAT-real): capacity 10 down / 2 up, delay 600 ms, jitter 30 ms, loss 1%.
Add a `source` field to `Profile` (`json:"source,omitempty"`) carrying the citation string; no change to the frozen CSV import format (new optional field is ignored on old rows). `BuiltinProfileIDs` must include the new IDs; `/api/profiles` returns them with source. Tests must fail if a new profile is missing or its source absent. Never edit `book/data/runs` or claim these are field measurements - they are emulation targets grounded in cited literature.

## 2. eBPF as primary measurement, tc as cross-check

The audit already carries `kernel_telemetry` (eBPF source, object sha, per-flow retrans events, qdisc drops, srtt). Today the operator-facing facts and health word come from userspace RTT alone. Change the shared computation (single source, two presentations) so that when `kernel_telemetry.available && source == "ebpf"` and flows exist, eBPF retrans events and srtt are the **primary** evidence for the loss/queueing word, with `tc` drop delta as corroboration only; when eBPF is absent or degraded, render `—` + the existing absence word and the real `reason`, never a fabricated zero or a green from userspace alone. The HUATUO caveat is mandatory: `tcp_retransmit_skb` reason is a grouping label, not a verified root cause - the copy must say "retransmissions observées", never "cause confirmée". Keep tc/qdisc as a separate named line, never merged into the eBPF number.

## 3. Interface honesty and feel-better

The `EmptyState` and `StatusPip` primitives already encode `—` + word and word+shape but are used by few live surfaces. Apply them consistently to the live Labo chapters (Matrix, Gates, CampaignTimeline, Observer, Comparer, Preuves) and Synthèse (TimeField footer, JournalRail) for loading/empty/stale/offline/error - no spinners, no fabricated value. Apply feel-better rules: `text-wrap: pretty` on absence sentences, `tabular-nums` on all live counters, concentric radii (outer = inner + padding), explicit transition properties (never `all`), 44px hit areas, reduced-motion and forced-colors preserved. No new hex, font-size, accent or component duplication.

## Slice order

1. Profiles + provenance (model/API, Go oracles).
2. eBPF-primary facts (shared lib, Go/TS oracles).
3. Interface honesty + polish (components/CSS, DOM + Playwright).
Each slice: spec -> plan -> fresh implementer -> fresh reviewer -> gates -> explicit-path commit. No parallel subagents, no frozen/ledger/gate edit, no host contact.
