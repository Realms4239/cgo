# CGO — LIEN Instrument Specification

Source of truth: `LIEN.md` (thesis "Protection du trafic opérationnel critique sur liens d'accès contraints"). Every requirement below traces to a LIEN section. Anything not in LIEN.md does not exist in this project.

Approved session decisions (2026-08-23) are binding: fresh repo at `C:\cgo`, donor `C:\thesis-cgo\thesis-cgo` supplies infrastructure patterns only, **presentation mode removed**, **no API token (double-confirm instead)**, **SSE at 10 Hz**, **no old views copied**.

## 1. Product

A self-contained Go instrument (`./cgo`) with an embedded React/Vite/TypeScript UI that:

1. audits accessible links non-intrusively from a client workstation (LIEN II);
2. replays constrained-link profiles in a controlled Linux testbed comparing qdiscs (`pfifo_fast`, `fq_codel`, `CAKE`) × CC (CUBIC, BBR) on profiles P1/P2, 3 repetitions = **36 events** default, **18** reduced (LIEN III.III);
3. archives results reproducibly and regenerates figures via `make figures` (LIEN II.V);
4. serves exactly four supervision panels (LIEN III.IV): Campagne · Temps réel · Résultats · Intégrité.

## 2. Backend (Go)

### 2.1 Modules (LIEN III.II)

| Module | Responsibility |
|---|---|
| campagne | orchestrates events: prepare → qdisc config → flux launch → collect → CSV → close; matrix runner; resume |
| qdisc | applies netem/TBF/qdisc via `tc` behind a `TCRunner` interface (fake for host tests) |
| sondes | ping (p50/p95/p99), small objects (4–32 KiB HTTP completion times), bulk TCP goodput, drops |
| métriques | pure functions: percentiles, deadline_ok_pct, bulk_goodput_mbps, wasted_bytes, cost_ar_per_h = wasted_bytes / 4.5·1024³ · 30000 |
| archive | writes frozen `aqm_eval.csv` + `link_audit.csv`, SHA-256 manifest, logs, quarantine journal |

### 2.2 Event model (LIEN III.III)

Phases per event: **baseline 30 s → charge 120 s → récupération 30 s**. Gates evaluated per event:

| Gate | Check | Fail action |
|---|---|---|
| G0 | target reachable | invalid |
| G1 | bulk load actually started | invalid |
| G2 | critical probes producing samples | degraded |
| G3 | latency physically plausible | invalid |
| G4 | throughput coherent with profile | invalid |
| G5 | no duplicate rows | invalid |
| G6 | baseline stable before charge | degraded |
| G7 | CPU not saturating measurement | covariable |

Gate status lands in `gate_status` column and the SSE frame.

### 2.3 Data schemas (LIEN Tableaux 6–7, exact columns)

`link_audit.csv`: audit_id, timestamp, site, link_type, provider, rtt_idle_p50_ms, rtt_idle_p95_ms, rtt_loaded_p50_ms, rtt_loaded_p95_ms, throughput_mbps, loss_pct, http_small_p95_ms, data_used_mb, notes

`aqm_eval.csv`: run_id, event_id, profile, qdisc, cc, repetition, rtt_p50_ms, rtt_p95_ms, small_p95_ms, deadline_ok_pct, bulk_goodput_mbps, drops, retransmissions, wasted_bytes, cost_ar_per_h, cpu_pct, gate_status

Frozen CSVs live under `data/runs/<run_id>/`; figures regenerate only from frozen CSVs.

### 2.4 SSE contract (10 Hz)

`GET /api/stream`, unnamed message consumed via `onmessage`:

```json
{"ts":0,"phase":"charge","profile":"P2","qdisc":"fq_codel","cc":"bbr",
 "repetition":2,"event_id":17,"load_status":"bulk-on",
 "rtt_p50_ms":82.4,"rtt_p95_ms":131.0,"small_p95_ms":140.2,
 "bulk_goodput_mbps":18.9,"drops":12,"gates":[true,true,true,null,true,true,true,false]}
```

Structural fields (profile/qdisc/cc/repetition/event_id/phase) may be delta-encoded — client retains last value. Named `backpressure` event retained. `Last-Event-ID` replay retained. Ring window: 120 s at 10 Hz = 1200 points, cap 4096.

### 2.5 REST surface (complete list)

```
GET  /api/state                  phase/profile/qdisc/cc/repetition/event_id/gates/load_status
POST /api/run/start              {profiles:["P1","P2"], reps:3}   double-confirm client-side
POST /api/run/stop               graceful stop + resume marker
POST /api/audit/start            {site,link_type,provider,duration_s}
POST /api/profile/import         JSON/CSV profile upload
GET  /api/results?run=           medians/IQR rows computed server-side from frozen CSV
GET  /api/integrity              archive_id, manifest sha256, valid/quarantined counts
POST /api/figures/regen          runs figure pipeline over frozen CSVs
GET  /api/replay/list            archived runs
GET  /api/replay/stream?run=     SSE re-stream of archived event series
GET  /api/report/export          short markdown report
GET  /*                          embedded SPA fallback
```

No auth middleware, no token. Security model: VM-local network locality + deliberate double-confirm UX for every mutating action.

### 2.6 CLI (LIEN III.III)

```
./cgo --serve                                   dashboard + API
./cgo audit --link-type 5g --site "Dept X" --duration 300
./cgo run  --matrix full|reduced --profiles P1,P2 --reps 3
./cgo verify                                    manifest integrity check
./cgo figures                                   regenerate SVGs from frozen CSVs
```

Figures are Go-generated static SVG (no node toolchain required on the VM).

## 3. Frontend

Fresh codebase. Nothing copied from the old views. Donor supplies only: worker-parsed SSE client pattern, ring-buffer live model pattern, `useLiveEchart` lifecycle wrapper, tree-shaken ECharts setup, fonts (Inter var, JetBrains Mono 400/700, Cormorant Garamond 500/600 woff2).

### 3.1 Shell

Persistent left sidebar (rebuilt from scratch): identity block (M stripe identity-only + CGO wordmark), four nav entries with keys `1–4`, campaign/event status block beneath navigation, G0–G7 gate strip at bottom. Collapses to icon rail. Mobile: stacked status header, one active panel, no horizontal overflow.

**Presentation mode, chapters, command palette: removed** (session decision). Keyboard shortcuts: `1–4` panels only.

### 3.2 Panels (exactly LIEN III.IV)

1. **Campagne** — selected profile/qdisc/cc/repetition selectors; current event state machine display (baseline/charge/récup with progress); gate strip detail; actions Start run / Stop / Start audit / Import profile, all double-confirm.
2. **Temps réel** — shared chart grammar: RTT p50+p95 traces with charge-phase rail shading, small-object p95 trace, bulk goodput area, drops bars; load status banner; LIVE/STALE/OFFLINE states; captions + provenance under each chart.
3. **Résultats** — comparative table (medians ± IQR per cell from `/api/results`), best-config highlight per chosen metric, quarantined-events list; Export report action.
4. **Intégrité** — archive id, manifest SHA-256 chain, valid vs quarantined counts, Verify action, Replay picker (list → re-stream into Temps réel charts with REPLAY banner), figures regeneration button.

Actions bar semantics: read-only actions single-click (Verify, Replay view, Export report, figures); mutating actions arm-then-fire (`DÉMARRER` → `CONFIRMER ?` ≤5 s or blur resets): Start run, Stop, Start audit, Import profile, qdisc apply.

### 3.3 Design system (carried from approved observatory direction)

Near-black canvas, hairline separators, sharp geometry (radius 0), uppercase display headings, light body copy. Fonts: Cormorant Garamond (display moments), Inter (body/nav), JetBrains Mono (measurements/provenance). M tricolor stripe identity-only; semantic telemetry colors separate: cyan primary/live, amber threshold/caution, red drops/errors, green passed, violet BBR, white neutral. Observatory texture = CSS calibration rules/phase rails only. No gradients, no rounded dashboard cards, no decorative imagery implying state.

Shared chart grammar builders: dark canvas, thin high-contrast traces, restrained fills, crisp event rails, explicit units, crosshair tooltips, reduced-motion honored. Every chart paired with DOM caption + provenance. Truthful states only: LIVE / REPLAY / STALE / BACKPRESSURE / OFFLINE / INCOMPLETE / NOT ASSESSED as text+marker, never color-only.

Accessibility floor: skip link, focus-visible, semantic landmarks, keyboard nav, aria labels, reduced motion.

## 4. Deployment & tests

Deploy engine (ported from donor `deploy/`, extended): VMware Workstation via `vmrun` primary; scan = `vmrun list` + default VM dirs + shallow drive sweep matching `*.vmx` (`--deep` full sweep); `ensure` = boot if SSH down (`vmrun start <vmx> nogui`) → poll SSH 5 min → deploy. Optional named-snapshot revert via yaml. Machine-local `deploy/cgo-vm.yaml`: user `altfloat`, host `192.168.174.128`, key `~/.ssh/id_ed25519`, project `/home/altfloat/cgo`, port 9090. Fresh secrets; old token retired. Cloudflare `--public` path kept but out of v1 flow.

Test topology:
- host: `go test ./...` unit (fake TCRunner), `bun run typecheck && bun run build && bun run test`
- VM: `make test-real` (real tc/netem/BBR), live SSE verification
- Playwright on Windows host against `http://192.168.174.128:9090` at 1366×768, 1920×1080, mobile

Verification gates before any "done": frontend trio, `go vet ./...`, `make test`, VM `make test-real`, Impeccable detector pass on changed UI.

## 5. Out of scope (dies with CONGESTION)

Detector bank, E1/E2/E3 matrices, oracle tc -s worship, negctl/preregister protocol chains, Madagascar map, Cytoscape topology, terminal view/FAB, frozen-wave replay machinery, API tokens/auth, presentation mode/chapters/command palette, old views/CSS.

## 6. Milestones

M0 scaffold + deploy engine(scan/boot/ensure) + gates green → M1 vertical slice (one smoke event end-to-end: qdisc→probes→CSV→manifest→SSE@10Hz→Campagne+Temps réel) → M2 full matrix + Résultats + Intégrité + figures + verify/replay/report → M3 audit CLI + import profile → M4 DESIGN.md rewrite as observatory authority + Impeccable pass + docs.
