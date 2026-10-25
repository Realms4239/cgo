# CGO Enterprise Rebuild — Design Specification

*Date: 2026-10-25 · Status: pending user review · Supersedes: the Oct 24 brainstorm
design (opencode session, exported in
`C:\Users\ASUS\Videos\critical-review-of-cgo-project-as-m2-thesis-2026-10-25.md`),
reconciled against the Oct 22–23 critical review and the working tree.*

---

## 0. Mission

Definitive architectural hardening and codebase reconciliation of the Congestion
Ground-truth Observatory (CGO) for the M2 thesis
(*Conception et évaluation d'un banc d'essai reproductible pour la comparaison de
politiques AQM et BBR sur des liens d'accès contraints*, ISGEI × Météo
Madagascar / DGM). Non-destructive: the frozen corpus (**159 runs / 528 rows /
208 valid / 74 degraded / 246 invalid / 22 cells**, verified by
`book/tools/audit-archive.py` → `OK`) is never re-measured, re-graded, or
rewritten. The rebuild replaces the *implementation* of the five deliverable
modules, pivots the economics from retail prepaid to enterprise contracts, and
harmonizes the thesis book with what the code actually is.

## 1. Decision log (binding)

| # | Fork | Decision |
|---|---|---|
| 1 | Rebuild scope | **Full in-place rebuild of the five modules** (cost, audit/report, campaign, bench, remediation); frozen data + CLI/API/UI contracts stay compatible. Not a greenfield rewrite. |
| 2 | Production topology | **Four-namespace bench is the only new-run path.** Legacy `kit/testbed.sh` survives solely to reproduce the frozen corpus. |
| 3 | Process placement | **Management veth (`m-host` 10.99.0.1 ↔ `m-client` 10.99.0.2); `cgo run/audit/--serve` execute inside `ns_client`;** dashboard on `10.99.0.2:9090`. Fallback if `sudo ip netns exec` from inside the namespace fails on CGO-DEV: host-attached management process, measured traffic stays in namespaces. Proven on VM, not assumed. |
| 4 | Enterprise valuation surface | **Audit + live UI + `valuation.json` sidecar only.** Frozen CSV `cost_ar_per_h` keeps its historical retail/emergency meaning; CSV schema immutable. |
| 5 | Model C | **Measured directly:** retain loaded 16 KB completion samples → `cap_deadline_ok_pct` at 220 ms; breach flag `< 95%`; unavailable (never a pass) without loaded small samples. Operator campaign `deadline_ok_pct` stays a separate metric, unavailable in link audits. |
| 6 | Flow regime | **`flows = 4` default for all new campaigns**; `--flows 1` reproduces the frozen regime; recorded in manifest + matrixConfig. |
| 7 | Bursty channel | **Campaign-level `channel_model=clean|bursty`**, valid only for P2/P3 (else validation error), recorded in manifest and valuation.json. No new profile IDs. |
| 8 | Valuation display | **All scenarios shown, primary marked by declared link type** (fiber→GOLD CIR, 4g/5g→corporate APN 150, vsat→wholesale 81.38, other→GOLD). Primary is a scenario match, not a contract claim. Retail 5,556 Ar/GB stays visible as "emergency out-of-bundle". |
| 9 | Model A retransmit ratio (this session) | **Linux `TCP_INFO` (`tcp_total_retrans` × MSS) on cgo's own bulk sockets** → `measured`. Probe-loss proxy → `estimated`. Neither → `unavailable`. With an external iperf3 bulk: iperf3's retransmit column if parseable, else `estimated`/`unavailable` — basis string always states which. |
| 10 | Field-audit claim (this session) | **Rewrite the thesis claim honestly**: protocol defined; pilot evidence = bench + LAN audits; Annex C rewritten. No field-audit execution planned. |
| 11 | Book source of truth (this session) | **The repo book build** (`book/content/*.js` → `book/build-docx.js` → `memoire.docx`). The Videos-folder `memoire-final.docx` is a downstream submission copy. |
| 12 | Spec format (this session) | Single comprehensive spec file (1M context available). |

## 2. Honesty invariants (testable properties — the rebuild may not regress any)

1. **No load ⇒ no grade.** Regression tests at 31 s and 60 s stay green.
2. **Unmeasured ⇒ `Unavailable` entry, never zero, never a pass** (CAP SLA,
   waste, G7 lineage, Model A ratio).
3. **No retroactive re-grading.** `cgo verify` + `audit-archive.py` output must
   remain byte-identical (`OK — corpus 159 runs, 528 lignes, 208 valides, 74
   degraded, 246 invalid` / `OK — 22 cellules, 159 manifestes, inventaire,
   texte et tableau 15 cohérents`); thesis Table 15 byte-identical.
4. **Scenario ≠ claim.** Valuation primary is a link-type match, labelled as
   such; tariffs are contractual scenarios, not measurements — the report says so.
5. **Fail-closed bench.** Offloads that cannot be disabled *and read-back
   verified* abort `up()`. `set_offloads` loses `|| true` on the four testbed
   veth pairs; `up()` ends with `check_offloads` on all six interfaces and dies
   on mismatch (today the read-back lives only in the separate `check`
   subcommand — a fail-open hole, fixed).
6. **Bursty identity is demonstrated, not asserted.** Theoretical mean-loss unit
   test + VM empirical calibration before any bursty campaign is legal.
7. **220 ms is the DGM operator objective**, never presented as a WMO mandate
   (WMO-No. 1061 defines WIS 2.0 "real time" as "a few seconds to a few
   minutes", no numeric per-alert threshold — cite it as supporting context).
8. **BBR in the thesis means BBRv1** (`tcp_bbr`, kernel 6.8 mainline; BBRv3
   never merged). The book says so explicitly.
9. **Never a 0 ms RTT record.** Probe timeouts scale with measured baseline
   (VSAT 600 ms ⇒ no 2 s cap killing legitimate 640 ms probes); timeouts are
   recorded as failures.
10. **"Achieved saturation throughput", not "link capacity"**, in audit reports
    and book wording.

## 3. Measurement topology (production path)

```
host browser ── m-host 10.99.0.1 ── m-client 10.99.0.2 ┐
                                                       │ ns_client
                                      c-cli 10.10.0.1 ─┘
                                           │
                                      c-wan 10.10.0.2
                                       ns_wan  ── netem delay/jitter/loss (both directions)
                                           │
                                      w-edge 10.20.0.1
                                           │
                                      e-wan 10.20.0.2
                                       ns_edge ── CAKE / TBF+pfifo / fq_codel
                                           │
                                      e-srv 10.30.0.1
                                           │
                                      s-edge 10.30.0.2
                                       ns_server ── testbedsrv :8081 / iperf3 :5201
```

- `kit/cgo-bench-dualns.sh` rebuilt as **topology owner, not policy owner**:
  namespaces, veths, routes, forwarding, management veth, offload enforcement
  (fail-closed per invariant 5). A default netem/CAKE may remain for standalone
  checks; production campaigns replace policy per cell.
- `LinkPhysics` (evolution of the existing `Deps.TC`/`TCShaper` pair —
  `SkipNetem` does not exist in the tree; per-cell netem already lives at
  `campagne.go:225`) applies netem per cell to both `w-edge` and `c-wan` via
  `ip netns exec ns_wan`.
- Shaping/AQM per cell on `e-srv` (upload) or `e-wan` (download) through
  `qdisc.NsRunner{Ns:"ns_edge"}`; stats read from the same egress.
- `kit/run-dualns.sh` (new): bench up → sink in ns_server → optional
  `cgo --serve` in ns_client → clean teardown.
- `cgo doctor` verifies offloads in all four namespaces (independent second
  layer — catches post-startup drift). Not running ⇒ `n/a`, never a pass.
- Bench `check` gains a **positive four-hop traversal assertion** (traffic
  provably crosses ns_wan → ns_edge → ns_server; no loopback shortcut) —
  generalizing `kit/stress_testbed.sh` item 2 to the dual-ns path; the stress
  suite gets a dual-ns edition (its header currently admits it targets legacy
  only and that claiming campaign results from dual-ns "would be false" — the
  rebuild makes that statement obsolete).
- Frozen-data guard: nothing under `book/data/runs`,
  `pkg/results/runs-frozen.tar.gz`, or the 159-run tables is rewritten. New
  runs are labelled `bench=dualns`, `flows=4`, `channel_model=clean|bursty`.

## 4. Enterprise valuation (`pkg/metrics/cost.go` + `cost_test.go`)

**Tariff table with provenance** (sourced `docs/data-prices.md`, extended):

| Tier | Type | Value | Source |
|---|---|---|---|
| `enterprise-gold-cir` | Model A plan | 1,000,000 Ar/month, 200 Mbps CIR | yas.mg Grande Entreprise, 31/08/2026 |
| `enterprise-platine1` | Model A plan | 2,500,000 Ar/month, 30 TB FUP | same |
| `b2b-platine-bulk` | Model B tariff | 81.38 Ar/GB | derived: 2.5 M ÷ 30,720 GB |
| `b2b-m2m-apn` | Model B tariff | 150 Ar/GB | operator quote (devis) — labeled as such, not a public grid price |
| `retail-outofbundle` | penalty | 5,556 Ar/GB | Yas Net Month 4.5 GB — demoted to emergency overage |

**Formulas:**
- Model A: `AmortizedARPerSecond = MonthlyOpEx / (720 × 3600)`;
  `CapacityWasteARPerHour = (wasted / transported) × (MonthlyOpEx / 720)`.
  The ratio comes from decision 9 (TCP_INFO / proxy / unavailable).
- Model B: `MeteredWasteARPerHour = (wasted_bytes / 2^30) × tariff × (3600 /
  windowSeconds)` — the directive's ×20 is the 180 s-window special case; the
  audit uses its measured load window; the campaign keeps ×20 (180 s charge
  window) for sidecar continuity with the historical meaning.
- Model C: `cap_deadline_ok_pct` from retained loaded 16 KB completions at
  220 ms; breach flag `< 95%`; label « ALERTE OPÉRATIONNELLE — MANQUEMENT SLA
  (WIS 2.0) » with the operator-objective framing of invariant 7.

**Report presentation:** all scenarios (GOLD, PLATINE, wholesale 81.38, APN
150, retail emergency) with the primary marked per decision 8. JSON schema
versioned (`economics.schema = 2`).

**Campaign compatibility:** frozen CSV untouched; new runs emit
`valuation.json` sidecar (Model A/B/C + channel_model + flows + bench labels);
live snapshot and dashboard show enterprise numbers.

## 5. Audit rebuild (`pkg/audit`)

**Phases (kept):** 10 s quiescent / 40 s load / 10 s drain; short audits scale.
Small-probe timeout scales with measured baseline (invariant 9).

**Direction (new):** `-direction up|down|both` (default `up`). Download = bulk
FROM the sink through shaped `e-wan` — the actual DGM scenario (200 MB EUMETSAT
downlink vs CAP egress). `both` = one audit with two back-to-back load windows
(up then down) and per-direction sections in a single report. Report and 80×24
label direction; drain measured per exercised direction.

**Parameter simplification (new):** primary surface collapses to
`cgo audit --site NAME --link fiber|4g|5g|vsat|other [--duration 60]
[--direction up|down|both]`. Auto-detection ladder: (1) dual-ns bench up →
full internal four-hop audit with bench sink (reliable default on the VM);
(2) no bench → external target, honest degradation matrix (Windows: no
TCP_INFO, no netns — labelled). `-bulk`, `-small`, `-target`, `-iface`,
`-routeros-iface`, `-qdisc`, `-asymmetric`, `-cpu-constrained` remain as
documented advanced overrides. External-link audits carry a **data budget
guard** (bounded bulk volume — the non-intrusive promise). Exit codes: 0 full
report, 2 partial (Unavailable entries present), non-zero hard failure.

**Blackout tolerance (new validation):** netem → 100% loss mid-phase-2 must
produce graceful exit, honest Unavailable/failure entries, zero crashes, zero
0 ms records.

**Artifacts per audit:**
- `data/audits/<id>/telemetry.csv` — raw per-probe timeseries (phase,
  timestamp, RTT, small completion, bulk bytes)
- `audit_report.json`, `audit_report.md`; latest mirrors at `data/audit_report.*`
- **Two seals:** SHA-256 of `telemetry.csv` (raw evidence) + SHA-256 of the
  appended `link_audit.csv` row file.

**Report contents:** grade (A+..F, Δp95 Waveform bands) + verdict, QDI, drain
ms, VoIP R (ITU-T G.107 simplified, loaded conditions), throughput (achieved
saturation), loss, **JFI** (multi-flow fairness — surfaces the per-flow data
`bulkflows.go` already produces; the metric that separates CAKE from pfifo
under flows=4), CAP SLA (cap_deadline_ok_pct + breach flag), Model A/B +
emergency valuation + primary scenario, direction, remediation, explicit
`Unavailable` entries.

**`link_audit.csv`:** appends `cap_deadline_ok_pct`, `cap_sla_breach` via the
existing migration path; old rows padded, never re-graded.

**80×24 CLI summary:** grade, Δp95, QDI, drain, VoIP R, JFI, CAP SLA,
direction, enterprise primary Ar/h, B2B Ar/h, seal prefix, RouterOS script
path + copy hint.

## 6. Campaign rebuild (`pkg/campagne`)

- `ProdDepsDualNS()` selects the hardened topology; `CGO_BENCH=dualns` enables it.
- `Deps` gains `LinkPhysics` (applies/resets netem on `w-edge` + `c-wan`, clean
  or bursty) and `ChannelModel`; the existing `TC`/`TCShaper` pair evolves into
  this (no `SkipNetem` exists to remove).
- `flows = 4` default (decision 6); G4 evaluates the multi-flow aggregate
  (C ≥ 0.85 × cap target). If the VM proof shows 4 CUBIC flows cannot reach
  the threshold on P2 (per-flow Mathis bound ≈ 2 Mbit/s ⇒ 4 flows ≈ 8 Mbit/s
  is *below* 0.85 × 20; CUBIC beats Mathis in practice but not provably), the
  matrix overrides `flows = 8` for CUBIC-under-loss cells — still inside the
  directive's 4–8 range. G6 keeps the profile-aware jitter allowance
  (`baseline.go`, `floor + 2.0 × declaredJitter`).
- **P4/G6 positive re-validation (new):** one P4 4-flow run on the dual-ns
  bench must pass G6 with declared ±20 ms jitter present (frozen P4 rows stay
  degraded by design; the book explains why).
- **Bursty gemodel fix:** `ApplyNetemBurst` rebuilt to the documented form
  `loss gemodel p r 1-h 1-k` where (per the netem implementers' paper — Carry,
  Briscoe & Kulykin) **p = P(good→bad), r = P(bad→good), 1-h = loss in bad
  state, 1-k = loss in good state**, decimals as in the paper's worked example
  (5% mean: `0.01 0.19 0.81 0.01`). The current five-argument form and the
  `H=200` test are removed. Parameters, re-derived under correct semantics
  (verified: mean loss = π_bad·(1-h) + π_good·(1-k), π_bad = p/(p+r)):
  - P2: `p=0.001, r=0.1, 1-h=0.5, 1-k=0.0005` → 0.544% ≈ 0.55% mean, bursts
    (bad-state sojourn ≈ 10 pkts)
  - P3: `p=0.002, r=0.1, 1-h=0.5, 1-k=0.001` → 1.08% mean
  - Unit test computes theoretical mean loss; VM step measures achieved loss
    over a calibration window before bursty campaigns are legal (invariant 6).
- Live snapshot gains enterprise cost fields + JFI; CSV columns stay frozen.

## 7. Remediation (`pkg/remediate`)

- Rebuilt around `Link{CapacityMbps, RTTMs, Iface, AuditProvenance}`; refuses
  zero/NaN capacity, zero RTT, no interface.
- Linux: CAKE at 95% with `diffserv4 ack-filter nat triple-isolate overhead
  rtt`; TBF + fq_codel and TBF + pfifo baselines preserved.
- **RouterOS v7+: real CAKE queue type** (research-verified: native since
  RouterOS 7.x — official MikroTik docs list `cake-rtt`, `cake-diffserv`
  incl. `diffserv4`, `cake-ack-filter`): idempotent `/queue type` + `/queue
  tree` with `cake-bandwidth`, `cake-rtt`, `cake-diffserv=diffserv4`,
  `cake-ack-filter=filter`, `cake-nat=yes`, `cake-wash=yes`. This upgrades the
  current `sfq` fallback (which stays only for pre-CAKE RouterOS, stated in
  the script).
- Header comments embed audit ID + evidence SHA when present.
- Asymmetric scenarios validated (RQ3): generated bandwidth must match
  measured capacity in symmetric and asymmetric cases.

## 8. UI / API / doctor

- **Existing layout and design tokens only** (`--surface-soft`, `--hairline`,
  `.btn`, `.card-head`, portal modal pattern). The **impeccable** skill governs
  the implementation quality pass.
- **Audit card simplification:** the 7-field form collapses to Site + Type de
  lien (select) + Durée (segmented 20/60) + Sens (up/down/both) + AUDITER;
  advanced parameters move into a collapsed details block. Phase progress
  (Repos → Charge → Drain) with live Δp95.
- **Result card:** grade badge (existing green/amber convention), Δp95, drain,
  R-score, JFI, CAP SLA badge, primary + B2B Ar/h, seal (truncated, click to
  copy), existing COPIER LINUX TC / COPIER ROUTEROS buttons (CampagneView
  :443-445 — extended, not replaced).
- **Campaign cockpit:** channel-model selector (clean/bursty, P2/P3-gated),
  4-flow default shown.
- **API:** audit report exposes `valuation` + `cap_sla` + `jfi` + `direction`;
  `/api/cost/enterprise` returns plans/tariffs; live snapshot exposes
  enterprise cost.
- **Doctor:** Linux — enumerate visible testbed interfaces + the four known
  namespaces, `ethtool -k`, fail in full mode if any GRO/GSO/TSO is on; bench
  not running ⇒ `n/a`, never a pass.
- **Security line (new):** dashboard binds `10.99.0.2` only; documentation
  states that anyone on the host can trigger shaping (lab/DSI context).
- **Windows capability matrix (new, documented):** audit works (ping + HTTP +
  optional bulk), grading honest; TCP_INFO, netns bench, doctor offload check
  = Linux-only, labelled `unavailable`/`n/a`.

## 9. Book harmonization workstream (repo build is truth — decision 11)

Chapter-level edits in `book/content/*.js`, rebuilt via `book/build-docx.js`;
`book/build/stats.json` regenerated; `book/build-qa.js` run; `memoire.docx/pdf`
regenerated. The user ports to `memoire-final.docx` downstream.

1. **Economic pivot:** rewrite retail-5,556 passages in `02-partie1` (×3),
   `04-pratique` (×4), `03-partie2` (×1), `08-glossaire` (×1): enterprise
   Models A/B, primary-scenario table, retail demoted to emergency overage,
   tariff provenance (GOLD/PLATINE sourced yas.mg; 150 Ar/GB = devis).
2. **Field-audit claim rewrite (decision 10):** "three sites / annexe C"
   becomes "protocol defined; pilot executions on bench + LAN"; Annex C
   rewritten; abstract + RQ1 answer aligned with limitations (review fix #3+#7).
3. **Front matter (review fix #7):** verify/correct supervisor, jury, date,
   Master vs Licence LISISA ambiguity.
4. **G4/Mathis multi-flow:** `03-partie2`, `04-pratique`, `09-annexes` —
   flows=4 regime, Mathis escape math (4 flows × ~2 Mbit/s ≥ 0.85 × 20 Mbit/s),
   JFI as the CAKE-vs-pfifo discriminator.
5. **G6/Starlink:** the 8 files referencing G6 get the profile-aware allowance
   (`floor + 2×declaredJitter`), frozen P4 rows stay degraded, justification.
6. **Regime bridge (new, crucial):** the book presents frozen (1-flow, legacy
   bench) and new (4-flow, dual-ns) as connected evidence via a planned
   **bridge experiment**: P2 reference cells (cake × BBR × CUBIC) run in BOTH
   regimes; the measured shift (Δp95, goodput, G-verdicts) is reported. Without
   it, two regimes = two worlds for the jury.
7. **New sections (zero coverage today):** offload hygiene (GRO/GSO/TSO and
   why 64 KB super-packets corrupt drop accounting); Gilbert-Elliott bursty
   channel (paper cited, parameters, mean-loss math); 3-phase audit protocol +
   dual seals + artifact set; download-direction audit as the DGM scenario.
8. **Honesty anchors:** BBR = BBRv1 (kernel 6.8); 220 ms = DGM operator
   objective with WMO-No. 1061 "few seconds to minutes" as supporting context;
   "achieved saturation throughput" wording throughout.
9. **Stale fixes:** `doc-final-build.cjs` "160 runs" → 159 (and its retail
   framing paragraph); corpus counts everywhere pinned to the checker's OK
   lines.

## 10. `docs/thesis_harmonization_m2.md` (deliverable 5)

Concise academic brief, written after VM validation cites fresh measured
numbers: (a) the economic pivot from retail prepaid to enterprise CIR
depreciation — rationale, formulas, provenance; (b) mathematical defense of
the G4/Mathis multi-flow update without invalidating historical tables; (c)
the G6/Starlink profile-aware jitter defense; (d) the regime-bridge argument;
(e) what remains unavailable and why that is honest.

## 11. Validation plan

1. Unit tests: every formula (Model A/B/C, amortization, window
   generalization), gemodel argument form + theoretical mean loss, CAP SLA,
   valuation primary mapping, artifact dual seals, flows default,
   channel-model validation (P2/P3-only), doctor parser, remediation
   provenance + RouterOS CAKE emission, JFI, direction handling, exit codes.
2. `go test ./...` on Windows host and CGO-DEV; frontend typecheck + build;
   `bash -n` on all kit scripts.
3. **VM batch (single session):** bench up (fail-closed offloads asserted;
   deliberate re-enable ⇒ startup dies) → four-hop leak proof → sink → 60 s
   audit (up + down) → inspect telemetry, dual seals, CAP SLA, valuation,
   JFI, remediation; API/UI smoke over `10.99.0.2:9090`; dashboard-inside-
   ns_client risk proven or fallback engaged.
4. **VM campaigns:** P2 cake CUBIC ×4 flows passes G4 (fallback per §6:
   matrix override ×8); P2 cake BBR same; P4
   4-flow passes G6 with ±20 ms jitter; clean vs bursty in separate campaigns
   (bursty gated by empirical mean-loss calibration ≈ 0.55% P2 / 1.08% P3);
   **regime-bridge cells** (P2 cake BBR + CUBIC, legacy 1-flow vs dual-ns
   4-flow); manifests record `flows`, `channel_model`, `bench=dualns`.
5. **Délestage proof (new):** SIGKILL the campaign mid-charge → rerun → resume
   journal completes the matrix → `cgo verify` + `audit-archive.py` pass on
   the result. Converts the thesis's délestage-robustness assertion into
   evidence.
6. **Blackout test (new):** audit under mid-phase 100% loss — graceful, honest,
   no zero-RTT records.
7. Teardown check: no namespaces, listeners, processes left.
8. Frozen corpus: checker output byte-identical; Table 15 byte-identical.
9. Book: rebuild, build-qa, stats regen; review `git status`; commit only
   source/docs/kit/dist — never backups, binaries, `data/`, or `book/data`.

## 12. Risks accepted

- Dashboard inside `ns_client` depends on `sudo ip netns exec` from within the
  namespace — proven on CGO-DEV first, host-attached fallback defined.
- flows=4 changes new-run absolutes; frozen tables stay 1-flow and labelled;
  the bridge experiment quantifies the shift.
- Enterprise tariffs are contractual scenarios, not measurements — report says so.
- CAP SLA depends on the 16 KB endpoint; missing small object ⇒ unavailable.
- gemodel decimal-vs-percent parsing is verified empirically on the VM before
  bursty runs (paper form is decimals; kernel `%` suffix handling confirmed on
  hardware, not assumed).

## 13. Research citations (this session)

- Netem gemodel semantics & worked example: Carry, Briscoe, Kulykin —
  *Definition of a general and intuitive loss model for packet networks and
  its implementation in the Linux kernel Netem* (the implementers' paper).
- RouterOS CAKE queue type (cake-rtt/diffserv/ack-filter): MikroTik official
  documentation (help.mikrotik.com, CAKE page, 2024) + community practice on
  RouterOS 7.3+.
- WIS 2.0 real-time framing: WMO-No. 1061, *Guide to the WMO Information
  System* — "a few seconds to a few minutes", no numeric per-alert threshold.
- BBRv1 in mainline 6.8; BBRv3 unmerged: google/bbr repo + IETF-117 CCWG
  slides.
- Tariff provenance: `docs/data-prices.md` (yas.mg Grande Entreprise,
  31/08/2026; 150 Ar/GB APN = operator quote).

## 14. Out of scope / explicitly not claimed

- Live OTA interface binding (SO_BINDTODEVICE) — unchanged from harmonization
  matrix §6: do not claim.
- Field audits (decision 10) — claim rewritten, not executed.
- RouterOS fq_codel — still does not exist; sfq fallback labelled for
  pre-CAKE RouterOS only.
- SQLite/WAL journaling — CSV append + manifest remains.
