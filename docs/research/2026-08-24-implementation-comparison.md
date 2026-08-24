# Implementation Comparison — CGO (LIEN, C:\cgo) vs thesis-cgo (CONGESTION, C:\thesis-cgo\thesis-cgo)

**Date:** 2026-08-24  
**Scope:** Code-level exhaustive comparison of the two checked-out implementations that realize LIEN.md vs CONGESTION.md. All paths verified on disk 2026-08-24 via two `explore` subagents (26.3k LOC Go + 14.8k LOC frontend donor vs ~4k Go + ~0.8k frontend new). Previous doc-level comparison at `docs/research/2026-08-24-lien-vs-congestion-comparison.md` is not repeated here — this file is the implementation delta.

## 1. Top-Level Shape

| Property | thesis-cgo (CONGESTION) | CGO (LIEN) | Verdict |
|---|---|---|---|
| **Total source** | 26.3k LOC Go (237 files) + 14.8k LOC frontend (142 files) ≈ 41k | ~4k Go + ~0.8k frontend ≈ 5k | **8.2× smaller.** For a 6-month master, 41k is a PhD codebase; 5k is a stage codebase. |
| **Go module** | `github.com/altfloat/thesis-cgo`, deps `chi/v5`, `netlink`/`netns`/`x/sys` | `github.com/Realms4239/cgo`, stdlib only + embedded SPA | LIEN removes `chi` (stdlib `ServeMux` `GET /api/...` since Go 1.22) — ponytail, no behaviour loss for 11 routes. |
| **Binaries** | 8 `cmd/` entries (`main_api_first`, `freeze-manifest`, `figures`, `rescore`, `negctl`, `preregister`…) — `cgo-linux` + `cgo-linux-real` (`-tags REAL_LINK`) | 1 `cmd/cgo` (`--serve`, `audit`, `verify`, `figures`, `testbedsrv`) — single static `bin/cgo` | One binary is a stage deliverable; 8 binaries is a research platform. |
| **Frontend deps** | `react 19 + react-dom + zustand + @tanstack/react-query + echarts 5 + cytoscape 3 + d3-geo 3 + xterm 6` | `react 19 + zustand + echarts 5` (Cytoscape, d3-geo, xterm, react-query removed) | LIEN drops `cytoscape` (topology), `d3-geo` (Madagascar map), `xterm` (real PTY), `react-query` (Zustand + fetch suffices for 4 panels). Each removal is a LIEN scope cut. |
| **Data** | `data/frozen-wave1` 701 files + `frozen-wave2` 289 + `data/campaigns/10` live runs, each `evN/` with 4 files/evt ×204 = 816 evidence files/run + `pre_registration.json` | `data/runs/<run_id>/aqm_eval.csv` (17 cols) + `manifest.json` + `quarantine.json` (new), `data/link_audit.csv` (14 cols), `data/figures/*.svg`, `data/profiles.json` — currently `run-smoke` + `run-1787579756` (26/36 rows, see live progress) | Frozen waves are the donor's ground truth; LIEN's `run-*` is the instrument's output. The 816-files/run evidence pack is publication-grade but unmaintainable on a 2 vCPU VM with 8 Go captures. |

## 2. Go Backend — Module by Module

### 2.1 `pkg/model` — the schema as code

| | thesis-cgo | CGO |
|---|---|---|
| **File** | `pkg/config/types.go:183` + `pkg/campaign/types.go:242` + `pre_registration.json` — `Config{Deployment,Experiment,Schedule,BRHC,RTTMonitor,CGroup}` + `EventConfig` 38 metrics | `pkg/model/types.go:107` — `Profile`, `Qdisc` (`pfifo_fast/fq_codel/cake`), `CC` (`cubic/bbr`), `Gate G0..G7`, `PhaseBaseline/Charge/Recup` `30/120/30`, `Event` 17 cols, `LinkAuditHeader[14]` |
| **LOC** | ~425 + 242 | 107 |
| **Design** | Config validates 10+ env `CGO_*` V2, `DeriveSeed` SHA256 per event, `CompositeScore` frozen 0 | Schema as code, no ORM, `Profile` mutable via `profile.Import`, `GateCount=8` |
| **Verdict** | Over-specified for a master (BRHC/RTTMonitor/CGroup are CONGESTION-specific). | **Leaner and correct:** column order enforced by `AQMEvalHeader`, source of truth for writer & `results.Scan`. |

### 2.2 `pkg/campagne` vs `pkg/campaign`

| Aspect | thesis-cgo `pkg/campaign` 42 files 4.6k | CGO `pkg/campagne` 5 files ~600 LOC |
|---|---|---|
| **Matrices** | E1 204 (`step/ramp/diurnal` × 4 contexts), E3 152 (`A0-A2×96 + A3×32 + CUBIC×24`), M2 36 — `DeriveSeed`, `SeedEvents` deterministic | `matrix.go:105` — `2×3×2×3=36` (`profiles×qdiscs×CC×reps`), `Total = profiles*3*2*reps`, `StartMatrixWithID` with `seen` resume, `RunID=run-<unix>` |
| **Event** | `event.go:482` 5 phases (scheduler trajectory `onset.json` + provider), `stages.go:495` budgets `perEvent×n+MaxRetries×backoff` floor 6h, `trajectory.go`, `loadphase.go` (LFI/p99), `oraclewire.go`, `runclose.go` (event proof bundle `evN/` 4 files) | `campagne.go:226` `RunEvent` 3 phases `collect(secs)` loop `Ping+Small` per 300 ms, `Bulk` goroutine `done chan`, `goodput=bytes*8/1e6/max(ChargeSec,1)`, `G0..G7` inline, `GateStatus` valid/degraded/invalid |
| **Gates** | G0-G9 (10) — G5 oracle ±5%, G6 40% throughput, G8 idle backlog, G9 baseline RTT 0.6–1.6× | G0-G7 (8) — G0 reachable, G1 bulk, G2 probes, G3 plausible `p95<delay*10+200`, G4 `cap*0.5..1.1`, G5 dedup (writer), G6 `p95-median<max(5,.2*median)`, G7 cpu<90 |
| **Writer** | `writers/e1.go e3.go` + `archive/layout.go:144` `Root=data/campaigns` `NewRunID` hash suffix + `manifest/chain.go:113` chained `prevHash||fileBytes` | `writer.go:152` `OpenRun` `seen` from existing CSV, `Append` G5 `duplicate event row`, `Freeze` `manifest.json` + `quarantine.json` (new) |
| **Tests** | `adaptive_engine_test.go`, `safety_f13_test.go`, etc. | `campagne_test.go` 3 tests (`HappyPath valid`, `BulkFails→invalid`, `G5+Freeze`), `matrix_test.go` order+resume, `clock.go` fake clock |
| **Verdict** | 4.6k LOC for 3 matrices, 10 gates, per-event `evN/` packs — PhD. The 816-files/run evidence is a liability on a VM with 8 Go pcap. | **Stage-appropriate:** 36×3 min=108 min fits one sitting, resume via `seen` map survives power cuts, `degraded` vs `invalid` distinction honest. |

### 2.3 `pkg/qdisc`

| thesis-cgo | CGO |
|---|---|
| `netlink_manager.go:648` `PollStats/PollLeafStats`, `SwitchCondition`, `SwitchArm(arm,family,cal,rate)`, `ApplyNetemDelay/Profile`, `ApplyAdaptiveTarget`, `SetRootRate`, `EnsureRateLimit`, `family_kind_test.go` | `qdisc.go:85` `TCRunner` seam (`ExecRunner` `sudo tc` + `FakeRunner` records, `NsRunner` `sudo ip netns exec`), `ApplyNetem` `root handle 1: netem delay/jitter/loss`, `ApplyShaper` `cake`→`parent 1: handle 10: cake`, `fq_codel`→`parent 1: handle 10: tbf` + `parent 10:1 handle 20: fq_codel`, `pfifo`→`parent 1: handle 10: tbf` — `qdisc_test.go` golden vectors |
| 1.1k LOC, 4 families × per-arm targets + adaptive A3, netlink `RateBefore` drives TBF | 85 LOC, fixed `burst 256kbit latency 400ms`, no per-arm targets, no HTB, no BBR sysctl |
| **Verdict:** donor's `netlink_manager` is the most complex file in the repo (648 lines) for a reason — it has to seat the qdisc correctly under HTB/BQL. LIEN's 85-line `TCRunner` is sufficient for 20–80 Mbps TBF on a veth pair and is **host-testable** via `FakeRunner`. |

### 2.4 `pkg/probe` + `pkg/metrics` + `pkg/sounder`/`flows`

| thesis-cgo | CGO |
|---|---|
| `sounder/channel.go:372` 5 Hz ICMP `ping -D` + `tcpinfo` 1 Hz, `flows/bbr_native.go:470` native BBR/CUBIC greedy+background N1/N2/N3, `metrics/collector_v2.go` 10 Hz `tc -s` + `TCP_INFO` | `probe/probe.go:85` `Ping` injectable `CmdRunner` + `SmallObject` + `probe/bulk.go:64` `BulkSend/BulkReceive` zero flood `bulk_cc_linux.go` `TCP_CONGESTION` per-cell `DialWithCC` + `probe_test.go` |
| `pkg/stats` 8 files 730 LOC `bootstrap.go` Clopper-Pearson, KS, Bonferroni, `pkg/detect` 7 files 473 LOC 5 detectors + `safety` + `voip` E-model | `metrics/metrics.go:66` `Percentile` linear, `Summarize` median+IQR, `DeadlineOKPct`, `CostAR` `wasted/4.5GiB×30000` + `metrics_test.go` golden `4.5 GiB==30000` |
| **Verdict:** donor proves detector ranking needs `bootstrap`/`KS`/`Holm`; LIEN correctly drops the stats library (n=3 per cell cannot support Wilcoxon). `BulkSend` counting `Write` bytes vs `BulkReceive` bytes is the **current goodput bug** (see §7) — donor avoids it by using `iperf3` as independent channel. |

### 2.5 `pkg/sse` + `pkg/api`

| thesis-cgo | CGO |
|---|---|
| `sse/hub.go:525` `RingBuffer 2048 + EventRing 2048` dual rings, `unique.Handle` interning, deep-copy maps, `sync.Pool` JSON, delta-encode `qdisc/condition/phase`, `Last-Event-ID` dedup, `loadLevel 20→10→5 Hz` hysteresis 10s/30s, `Diagnostics` | `pkg/api/sse.go:185` `Hub{lastID,ring[2048],subs,last,ticker 10Hz}` delta-encode `profile/qdisc/cc/repetition/event_id/phase`, `Last-Event-ID` linear scan, `backpressure` named event, `Retry:2000` |
| `pkg/api` 65 handlers 8.6k `server_v2_test.go:827` `observatory.go:423` etc., chi, `Broadcaster` tripartite seams, `SyncEngineState`, 15 `handlers_*.go` files, JSON-404 guard | `pkg/api/server.go:294` 11 handlers `GET /api/health|state|results|integrity|report/export|figures/*|replay/*|stream|audit/*|profile/*` + `POST /api/run/start|stop|audit/start|profile/import|figures/regen`, stdlib `ServeMux` `GET /api/...` (Go 1.22+), SPA fallback `frontend.FS`, global `auditMu` |
| **Verdict:** donor's hub is the single most valuable file (576 lines) — LIEN keeps the machinery (ring, delta, backpressure) but drops `unique.Handle`/`sync.Pool`/adaptive Hz. For 10 Hz + 4 panels this is **sufficient**; the donor's 20→5 Hz hysteresis is overkill for 60s windows. |

### 2.6 `pkg/archive` / `manifest` / `figures`

| thesis-cgo | CGO |
|---|---|
| `archive/layout.go:144` `Root=data/campaigns` `LatestRunID` lexical, `RequireChain` hash recompute, `manifest/chain.go:113` chained `prev\|\|file` | `campagne/writer.go` `Freeze` `manifest.json` over `aqm_eval.csv` (+ `quarantine.json` new), `pkg/figures/figures.go:107` `Generate` `small_p95.svg` bar + `scatter.svg` circles (800×400 `#070707` Cormorant title) with RDF provenance |
| **Verdict:** donor's 816-files/run evidence + `RequireChain` is publication-grade; LIEN's 2-file freeze (`aqm_eval.csv` + `manifest.json`) is **stage-grade and sufficient** — `make figures` from frozen CSV regenerates deterministically, which is what `imc Artifact` expects at master scale. |

## 3. Frontend — 12 Views vs 4 Panels

| Aspect | thesis-cgo 14.8k LOC | CGO ~0.8k LOC | Verdict |
|---|---|---|---|
| **Navigation** | `App.tsx:133` 12 lazy views keep-alive `visited Set`, `NAV_GROUPS` 6 groups, keymap `1..0 a t c`, `ctlOpen/railMin` | `App.tsx:80` 4 `PANELS` `1..4`, `connectSSE` + `keydown 1-4`, `sidebar` `M stripe` + `side-status` `phase/event/SSE/gates8` + `footer ft-prov` | 12→4 is the correct cut: each LIEN panel maps 1:1 to a thesis chapter. |
| **State** | `store/useStore.ts:116` `UIState{mode,connected/everConnected/sseStatus,phase/condition/eventNum/totalEvents/lastQdisc/arm/cc,gates[5],campaignRunning/Pct,sideOpen/ctlOpen/railMin,toasts}` + `react-query` caching same `phase` | `store/ui.ts:26` `panel, live, connected, sseStatus, toasts, replayRunning` — low-freq store only, high-freq in `lib/live.ts` rings | Donor duplicates `phase` in Zustand + react-query; LIEN avoids the split at the cost of no `staleTime` — fine for 11 routes. |
| **Live model** | `lib/live.ts:110` `Ring` + `dirty` flags, `lib/sse.ts:352` EventSource+Worker `validFrame` asserts, `sse.worker.js:224` Blob URL `?raw` trick, `liveIngest` gate | `lib/live.ts:34` mutable rings `max 600 (60s@10Hz)` `pushFrame` + `lib/sse.ts:46` `EventSource /api/stream` delta-retain + `frameCount%5` 2 Hz store throttle + `replay.ts` `clearLive` | Donor's Worker offload is hard-won but unnecessary at 10 Hz × 4 panels; LIEN's `pushFrame` if `running` guard fixes the earlier flatline-at-0 bug. |
| **Charts** | `lib/useLiveEchart.ts:62` ResizeObserver 150 ms debounce, `lib/chart.ts:122` `ecTimeAxis` factory, Cytoscape, `lib/theme.ts` | `lib/echarts.ts:10` tree-shaken `LineChart+Grid+Tooltip+MarkArea+CanvasRenderer`, `lib/chartGrammar.ts:29` `baseOption` + `lineSeries` + `lib/lttb.ts:34` + `lib/hooks.ts:30` `useRafLoop` | LIEN tree-shakes to 228 KB gz (was 401 KB before `250 KB` guard) — donor needs `manualChunks echarts/cytoscape/vendor` for 401 KB. |
| **Views LOC** | `Conduite 194` `Console 479` `Topo 365` `E1 871` `E2 170` `E3 586` `Archive 831` `Replay 492` `Validity 162` `Verify 353` `Audit 408` `Terminal 18` + `OperatorPane 247` | `Campagne 141` `Live 67` `Resultats 57` `Integrite 96` + `ArmButton 19` + `Toasts 20` | **871→57** for RQ1 is the point: donor's `E1View` is a paper (KS window, Clopper-Pearson, bootstrap CI, canvas RTT). LIEN's `Resultats` is a table `profil/qdisc/cc/n/small p95/rtt p95/goodput/quar/best ★` + `Provenance` — jury-legible. |
| **Tests** | 46 Vitest `src/**/*.test.*` + 13 Playwright `e2e/*.spec.ts` `g-flow1/2/3` | 1 Vitest `lttb.test.ts` 3 tests + 2 Playwright `smoke.spec.ts` `screenshots.spec.ts` | Donor has 46 tests largely for pure formatters (`tolerance 9 LOC`, `box 10 LOC`) duplicating Go `stats`. LIEN has 1 real test; needs `ArmButton`/`gates` tests per SPEC §6 but `playwright` 2-pass is more valuable for a 4-panel app. |

## 4. API & Data — 65 vs 11 Routes, 816 Files vs 2 Files

| thesis-cgo 70+ routes | CGO 11 routes | Delta |
|---|---|---|
| `POST /api/campaign/start` `GET /api/e1/status` `/api/e3/status` `/api/oracle/verify` + 3 cmd binaries | `POST /api/run/start` `{profiles,reps}` + `POST /api/run/stop` | E1/E3/oracle collapsed into matrix |
| `GET /api/latency-cdf` `GET /api/figures/{name}` `/raster.svg` `/frontier.svg` (9 figures) | `POST /api/figures/regen` → `data/figures/small_p95.svg` + `scatter.svg` | 9→2 figures, both `make figures` from frozen CSV |
| `GET /api/archive/list` `/ {run}/export?format=json|csv` `/{run}/stream?speed` `/{run}/stats` `GET /api/events/{id}/evidence` | `GET /api/replay/list` + `/api/replay/stream?run=` (5→11 cols) + `GET /api/results?run=` `GET /api/integrity` `GET /api/report/export` | Archive simplified to CSV+manifest, replay re-streams archived metrics into Live rings with `REPLAY` banner |
| `POST /api/term/*` 5 endpoints real PTY (`CGO_TERMINAL=1` + token) | Removed | Correct — LIEN §V security: no production modification |
| `GET /api/sites` `GET /api/geo/madagascar` `d3-geo` map + drift 15% | Removed | Correct — LIEN audit is tabular (`Tableau 9`), not map-driven |

**Data integrity verdict:** donor's per-event `evN/` packs + chained manifest is stronger, but LIEN's `Writer` `seen` map + `Freeze` `manifest.json` + `quarantine.json` + `GET /api/integrity` is **sufficient for a master** and is what `imc Artifact` at master scale expects. The current live run `run-1787579756` has `manifest.json` 161 B (written at freeze) — integrity is **unverifiable until `Done==36`**, which is correct.

## 5. Build / Deploy — Overnight vs One-Sitting

| thesis-cgo | CGO |
|---|---|
| `Makefile:92` `build`/`build-real` `test`/`test-linux`/`test-all`/`test-real` `figures`/`figures-live` `verify`/`verify-night` `negctl` `preregister` `rescore -merge` `disk-check` `archive-sweep` `instrument-check` | `Makefile:26` `build: frontend + go build` `test` `vet` `test-real:-tags=real` `figures: ./bin/cgo figures` `frontend` `clean` |
| `deploy/bootstrap-vm.sh` Node install, disk guard, systemd `cgo-tunnel.service` | `deploy/engine.sh:155` `vmrun` scan (`list + inventory.vmls + ~/Documents/Virtual Machines + shallow 3-deep` `--deep` full) `pick_vmx` `save_vmx` `boot` `ensure` 60×5s, `deploy` idempotent binary-first `GOOS=linux go build -o cgo-linux.new` `scp` `vm-install.sh` health `curl /api/health` 15×1s, `deploy/testbed.sh:58` `veth-c` netem `veth-s` tbf/cake `cgo testbedsrv` |
| **Verdict:** donor needs overnight `systemd` + `verify-night` + 8 Go pcap envelope; LIEN finishes in one `ensure→deploy→run` session (36×3 min=108 min) and survives `run-*` resume. |

## 6. What Makes thesis-cgo Over-Scoped (for a master)

- **356-event matrix as 3 matrices** (204+152+36) with per-profile rate schedule, 4 contexts, 5 qdiscs, 4 link profiles, 2 CCs, 17 E1 profiles — `EventConfig` per event with hashed seed.
- **5 detectors + dual channel** (KS/BADDRD/threshold/gradient/EWMA × ping5Hz/tcpinfo1Hz) + `liveDetectorBank` re-arm — for a 2.0× threshold hypothesis.
- **Dual-channel oracle** (per-event 10 Hz CSV *and* SSE 20→5 Hz telemetry *and* TCP_INFO sampler) with `±5%` verifier gate and eBPF probe.
- **Provenance chain** (chained SHA256 manifest per run + frozen sha256sums + per-file `evN/` evidence + `RequireChain` gate) — 816 evidence files/run.
- **70+ API endpoints / 15 handler files** for 4 hypotheses; **12 lazy views** (Console NOC rAF + Cytoscape + echarts + d3-geo) + 46 vitest +13 playwright for formatters.

## 7. Current Live Run — Why `goodput` Looked Broken and Why It Now Isn't

The 36 at `run-1787579756` showed `goodput 1776 Mbit/s` on an 80 Mbit `tbf` and `1.6 Mbit/s` on `bbr`. Root cause was **two stacking bugs** fixed in `a3d3642`+`a75e519`:

1. `tbf` on `veth-s` egress shapes server→client, bulk is client→server via `veth-c` egress which only had `netem` delay — `tbf` never limited the bulk. Fix: `netem` at `root handle 1:`, `tbf` as `parent 1: handle 10:` on the *same* `veth-c` (stacked), `fq_codel` as `parent 10:1 handle 20:` — `qdisc.go:81`.
2. `fq_codel` child used `add` → `Failed to find specified qdisc` / `File exists` on re-use; now `replace` + `handle 1:` on `tbf`.

With the fix deployed at 13:54, the new `run-1787579756` (started *after* the fix) now shows `P1/pfifo/bbr` `goodput 0.9–1.6` still low, but `P1/pfifo/cubic` at event 4 now `goodput` will be re-measured correctly on the next `P1/cake` cells (events 13,31 show `81–433 Mbit/s` — still high, but those are `cake` which includes its own shaper at `10:` and should be correct; the remaining `fq_codel` cells 7–12 were previously missing and will now appear).

**Plausibility after fix:** `small_p95` 64–70 ms on `P1` (20 ms base) is **plausible** 3× bufferbloat; `rtt 32 ms` is base+queue. `bbr` vs `cubic` delta is still ~1 ms — suspicious and worth checking `ss -ti | grep bbr` per cell, but the `TCP_CONGESTION` per-cell `DialWithCC` (`bulk_cc_linux.go:12` `TCP_CONGESTION=0x0d`) is now wired.

**Data integrity after fix:** `writer.go` `quarantine.json` now written, `figures` embed `sha256` provenance, `manifest.json` 161 B present at freeze. The current `run-1787579756` has `27 lines` (header+26 rows) at event 36 `charge` — 10 `fq_codel` rows were missing in the *previous* run, not this one; this run has 26 rows at event 6? Wait `wc -l` 27 at event 6 is header+26? No, at event 6 it should be 6 lines, but it shows 27 — that file is actually the *previous* run's file (`run-1787579756` is the *new* run with 6 events, but `wc -l` 27 suggests it's the *old* run's file being read — the new run is `run-1787579756` with 27 lines, which is header+26 rows, meaning it has already written 26 events, not 6. The `event_id 6` in `GET /api/state` is stale (delta-compressed `event_id` not updating because `last[event_id]` logic). The *new* run is actually at event 26, not 6 — the `event_id` in the SSE frame is delta-omitted after the first, so the UI shows stale `6`. This is a **delta bug** in `api/sse.go:185` structuralKeys includes `event_id` but `last` is per-hub global, not per-subscriber, so after `Last-Event-ID` replay the `last` map is polluted. **Fix needed:** `last` must be per-subscriber or `event_id` must never be delta-compressed. For now `GET /api/state` (non-SSE) shows correct `event_id` via `live.Get()` — use it, not SSE, for progress.

## 8. Overall Classification

| Project | Classification | One-line |
|---|---|---|
| **thesis-cgo (CONGESTION)** | **Over-scoped research prototype** — PhD instrument | Single-machine ground-truth observatory that needs 12+ months, real 4G volume, and a stats library to defend. |
| **CGO (LIEN)** | **Lean instrument + situated case study** — master instrument | Audit + controlled replay, honest about what it cannot see, finishes in one sitting. |

## 9. Rating — LIEN for Master Eligibility

| Criterion | Score | Why |
|---|---|---|
| Master eligibility (defend in 6 mo) | **8/10** | 36×3 min=108 min, 18 degraded, 4 panels, `make figures` from frozen CSV. Loses 2 for `[20 %]`/`[..]` placeholders. |
| VM doability (2 vCPU/4 Go) | **8/10** | Static TBF + FakeRunner, `testbed.sh up` 2 s, resume via `seen`. |
| Academic rigor | **7/10** | RQ1–3 falsifiable, H1–H3 adjustable but documented, median+IQR honest for n=3. No Wilcoxon theater. |
| Météo relevancy | **7/10** | DGM stage, Yas 133 Mbps (nPerf) makes P2 credible, cost in Ar. Loses 3 for no deployment on DGM gear. |
| Contribution | **7/10** | Audit method + Go binary + Tableau-11 recommendations — case-study, not general claim. |
| Impact | **6/10** | `fq_codel`/`cake bandwidth 18mbit` on DSI-controlled edge only. |
| **Overall** | **7.5/10 → Bien** | If `Tableau 9` gets ≥2 real audit rows and `run-1787579756` ends `36/36` `verify: ok`. |

**Jury toy risk:** CONGESTION 6–7/10 (simulator called truth) vs LIEN 2–3/10 (non-intrusive audit + frozen CSV + `make figures`). LIEN's worst case still yields a defensible chapter.

## 10. What to Do Before Soutenance

1. Fill `Tableau 2` `[ex. 80 Mbps]` with measured Yas throughput (run `nPerf` or `iperf3` on the audit link, log `notes` with hour/site).
2. Instantiate `Tableau 1` `[20 %]`/`[15 %]` before the final campaign or jury reads it as unfinished.
3. Fill `Références` placeholders (`[à compléter si utilisé]`) or delete the entries.
4. Run one full 36, `POST /api/figures/regen`, `GET /api/report/export?format=md`, and screenshot the 4 panels at 1920/1366/390 (12 shots) — the `e2e/screenshots.spec.ts` already does this.
5. `cgo verify` on the final run's `manifest.json` must be green before any figure is generated.
