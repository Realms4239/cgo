# Wall-kit Reunite — NOC Observatory Heavy Reassembly (M8/M9/Wall-kit Reunited)

**Date:** 2026-08-27
**Status:** Approved (Sections 1–7) — 56 QAs heavy discovery
**Base:** `m8@2e7f131` + `wall-kit@2026-08-26` + `main@cbc99a3` (`docs/SPEC.md §4` VMware, `DESIGN.md` observatory authority, `LIEN.md` 1379 lines + `ARG.md` 70% real)
**Source:** Atelier furniture splattered: charts non-functional, views badly laid, Integrité scroll error; GoAccess `rt.goaccess.io` panel chooser liveliness + ArtificialAnalysis leaderboard dense comparative; user grills: wall-kit adopted, hybrid lively, triple chooser, triple ECharts clean, C→B→A layout fix order, 3 instrumentation seams, both lineages Lab 36 + audit 30s, progressive Traduction/Kit, selective HD sparkline, functional Beam/Donut, hidden gates, config-driven + P3 heterogenous, worker SSE, honest EmptyState, Arm 5s, full ledger C, kit build-then-embed, VMware detect-fallback VirtualBox + Cloudflare tunnel cfut_… (env), bumped caps TOTAL 600KB/echarts 350KB
**Approach:** **A debug-first + debug-after double surgery**: Layer 0 fix splatter via 3 seams + Playwright clip+surgical logs+embed analysis → Layers 1–4 build heavy (backend P3 + shell + Wall HD + kit) → Layer 5 heavy debug-after whole execution (both lineages frozen triple-provenance + 12 shots)

## 1. Architecture — Wall-kit NOC observatory as Kit

**Goal:** One Wall `/` that in 5 minutes lets a DSI technician `Auditer 30s Yas 4G at Dpt → Comparer live pfifo grey dashed 73.9 vs CAKE cyan solid 46.9 same scale −60% diff pill → Exporter 1-page MD/CSV cost 0.90 Ar/h` — jury sees `baseline vs algorithm` as thesis proof `H2` (LIEN Tableau 1), not toys. Lab `50→25` heterogenous diurnal collapsed is `P3` import, not hardcoded.

**Stack:** `Go 1.25 bin cgo embed dist` + `React 19 Vite 6 Zustand low-freq + Jotai-like granular selectors (mix B+C)` + `ECharts 5 tree-shaken Line/Bar/Scatter Grid/Tooltip/DataZoom/VisualMap/Graphic Canvas useDirtyRect` + `D3 scale/shape/axis/brush` + `animejs 4.5 svg/text/animatable layout via context7 1983 snippets` + `woff2 Cormorant/Inter/JetBrains` self-hosted (CDN fallback only if offline flexible) + `kit/engine.sh` ops (VMware/VBox/Cloudflare) + `cloudflared` + `VBoxManage` + `meteolink top` TUI.

**Structure — Wall-kit, not 4 equal toys:**
- `/` **Wall** NOC live floor: `header 48 METEOLINK lockup + PanelChooser tri-toggle metric groups/chart craft/temporal source + density airy/dense + FootTicker 28 hash + rail 56↔232 var(--rail-w) + main wall bento max 1280 + footer` — `PanelChooser` persists `localStorage panel-visibility` drives both Wall and Drawer via `CustomEvent` without React re-render
- `Sheet Cmd-K 380px backdrop blur` **control cockpit**: `P1/P2/P3 reps×qdisc×CC + Auditer 30s InlineField 10–600 + Import INLINE Id/cap>0/delay>0 + ArmButton 5s progress line 1px amber 100%→0%` → auto-minimizes to pill `Event 3/6 — charge 4/10s` at top after `CONFIRMER`
- `Drawer History` **comparative appendix**: `56px peek` → full `12 groups` (`bar grey→green ★ + scatter D3 pareto brush rect + diff pill −60% + provenance hash 8-char + EXPORT MD/CSV + hardware inline`) — auto-peeks when `live.running`, `row append + scatter pop anim stagger` when live
- `/archives` **Archives** footer link **provenance**: `Recommandations — Traduction Matérielle progressive disclosure (Sheet>Wall>Archives) + RDF provenance sha256 chain + runs ul hover peek 4 rows + quarantine yellow invalid-only + Figures bar+scatter RDF + Replay → Live REPLAY banner`

**Why this split:** Wall is GoAccess real-time + ArtificialAnalysis dense reconciled as one live instrument (`10Hz SSE → live.ts rings 1800 phase-driven → useRafLoop 4Hz → ECharts`) while Sheet is the `Kit Portable` driver and Drawer is the `Table 10` comparative; `kit/engine.sh` makes `bin/cgo` the single portable artifact (`embed dist` + `manifest.json` chain) — thesis survives without oracle worship.

## 2. Data & Metrics — Tableau 4+5 heterogenous, both lineages, gates honest

**Primary Tableau 4 + secondary Tableau 5 — every metric has baseline vs optimized delta on same scale:**

| Metric | Baseline unmanaged `pfifo_fast/cubic` median | Optimized `fq_codel/bbr` or `cake` | Viz (hybrid lively) |
|---|---|---|---|
| `rtt_p95` | `191 ms` pfifo median | `76 ms` cake best `→ -60%` | Overlay grey dashed vs cyan solid + QDI amber fill 60×12 clipPath brush |
| `small_p95` hero 300px | `738 ms` P1 pfifo | `20 ms` cake `→ -82%` | Bar grey→green `★ best` + hero live line lttb400 |
| `bulk_goodput` violet | `77.7 Mbps` leaf-only valid G4 | `—` idle → valid | Area 25% violet |
| `cost Ar/h` | `0.90 Ar/h` | `0` idle `→ — steel` | Mono `f4b400→767b84` honest, not red flatline |
| `JFI` | `—` idle | `0.98` | Donut 56px steel→green, — when idle |
| `wasted/cost drops` | `0` honest idle | `wasted×1448 /4.5GiB×30000` | Numeric + trend, no spark when idle |

**Storage & truth — both lineages frozen `data/frozen-wave3` no template:** Lab `tc netem 20/100/600ms distribution normal + tbf stacked parent 1: leaf-only 80/20/5 Mbit + DialWithCC per socket + StatsFn PollStats leaf-only delta` → `metrics Summarize Median/P50/P95/IQR 25/75 + DeadlineOKPct 1000 fixed + CostAR + JFI detail=1 gated` → `writer aqm_eval.csv 17 cols + manifest sha256 + quarantine.json (only invalid) + hardwareRec per profile best single hardware.go source` **and** Edge `laptop→Yas 1.1.1.1 ping 5Hz + Small 4–32K http.DefaultClient + bulk BulkSendTo/iperf3 3 windows idle 0–12s bulk 12–22s loaded 22–30s split + LossPct missing vs expected + DataUsedMB/Throughput` → `link_audit.csv 14 cols` `rtt idle 20→375 small 52→754` both frozen with `manifest sha chain`.

**Gates G0–G7:** `G0 target, G1 bulk, G2 probes, G3 plausible (<delay*10+200), G4 throughput coherent [0.5*cap,1.1*cap+0.5], G5 dedupe writer seen map, G6 baseline stable P95-Median<max(5,0.2*Median), G7 cpu<90` — `quarantined` only `invalid` (G0/G1/G3/G4/G5) `→4 not 40`, `degraded G2/G6 amber` stays valid; tiny mono `PASS/FAIL/—` in rail only (Q13 C), no large strip, `quarantine yellow` only in Archives.

## 3. Visual System — tokens, bento, charts clean-lively, TUI, guarded motion

**Tokens:** `DESIGN.md sharp 0 #070707 #101012 hairlines #26262a --rail-w --gap 24 airy/16 dense --dur 80/150 cubicBezier + radial cyan/violet noise 1% WebGLMesh 1%` + `METEOLINK lockup Cormorant 600 gradient text-clip #f2f2f4 fallback + 16×16 NOC icon satellite→wave stroke1.5 round + 4px tricolor stripe + shimmer hover splitText grid[4,2] + svg drawable 800`. `telemetry cyan #5ad3e3/threshold #f4b400/danger #e22718/ok #1fa348/bbr #b48ae0/steel #9aa3ad` separate from identity M tricolor. Single source `src/styles/tokens.css` consumed by all via `var(--t-*)`, no hardcoded `#070707` in `LiveView`.

**Layout:** `header 48 glass sticky + rail 56↔232 animate width 400 var(--rail-w) + main wall bento max 1280 container-type:inline-size + footer 28 ticker tabular hash 8-char hairline top` `tidy+breathing panel-stack gap24 airy/16 dense card-head 11 caps view-title 20 centered provenance 10 right-aligned equalized subgrid` `bento scoped #wall 2-col @1100px` + `@900px rail 100% row 48 + header wordmark + scatter full width` + `prefersReducedMotion 1ms`.

**Charts HD clean-lively (triple strip Q5):** `ECharts canvas useDirtyRect min(dpr,2) at init, grid 64/32 inset hairline #141416, thin traces, hairline axes 10px JetBrains #8b9099, crosshair tooltip blur16` **stripped** `navigational chrome dataZoom slider handleIcon + brush toolbox slider + graphic watermark 28px 3% kept only when !idle + decorative LinearGradient blur12 + watermark` + `semantic overload visualMap piecewise + markArea per-series` kept single `markArea CHARGE amber dashed` same scale baseline vs CAKE; `rtt 180 + goodput 180 + hero 300 full-width date Last updated pulse` share `cs/ce` from `live.phase` not estimate; `MetricCard 60×12 clipPath rx4 lttb40` selective (continuous only, Q11 B), `QDI 60×12 amber brush`, `DonutJFI 56px`.

**TUI+Web:** `meteolink --serve` Web Wall + `meteolink top` TUI 8 cards ASCII sparklines `60×12 ▁▂▃▄▅▆▇█` via same `live rings 1800` + `kit/engine.sh npm/brew/go` cross-platform `postinstall` detects `linux/darwin/win32 x64/arm64` downloads `cgo-linux/macos/win` from releases.

**Motion guarded:** `anime 4.5 svg/text/animatable layout 1983 snippets` only on **state changes** (`rail width animatable, wordmark shimmer hover, panel enter stagger40, prompt 6s linear pause on hover, toast y[16,0] stagger20, shake x[-4,4,0] elastic, Beam dashOffset -40 infinite when running, Donut arc, bars scaleX left`) all `prefersReducedMotion() → return` early, `animateGrid` single `createTimeline grid[4,2]→[2,3]` pause on panel change.

## 4. Components & Views — layered primitives, PanelChooser tri-toggle, Sheet cockpit

**Carried:** `Rail WebGLMesh Prompt Palette PeekPopover Timeline QDI/JFI MetricCard DataTable Provenance EmptyState Toasts ErrorBoundary Skeleton PhaseStep lttb hooks validation` all `data-testid` for `playwright clip`.

**New primitives:**
- `MeteolinkWordmark.tsx` `METEOLINK·LIEN 600 + NOC icon + stripe + shimmer`
- `PromptProgressLine.tsx` bottom `1px #f4b400 100%→0% linear 6s pause on hover` used in `QuickActionsPrompt` `Event 3/6 — charge 4/10s` when running (not `Actions rapides` choices) + `ArmButton` `5s progress line` above button when armed `spring scale 0.96→1`
- `Beam.tsx` `animated-beam` 1 SVG `path linearGradient strokeDasharray 4 dashOffset -40 linear infinite` between Wall and Sheet when `live.running` + `ResizeObserver` hover-sync `Campagne ↔ Wall`
- `DonutJFI.tsx` `donut-chart` arc `0–360 56px stroke 8 steel→green` `— when idle` + `QDI 60×12` clipPath rx4
- `PanelChooser.tsx` GoAccess-inspired pill `Wall|History|Archives` tri-toggle `metric groups / chart craft line|bar|area / source live|frozen|both overlay` persisted `localStorage panel-visibility` drives Wall and Drawer via `CustomEvent` without React re-render
- `kit/engine.sh` `kit/cgo-vm.yaml` + `pkg/results/hardware.go` single source + `pkg/api/translate.go` `GET /api/hardware/translate?profile=P2`
- `TuiTop` via `cmd/meteolink/main.go` `meteolink top` shares `live` rings

**Enhanced views:**
- **Wall `LiveView`:** `8 MetricCards selective HD + QDI amber + DonutJFI + hero 300 date pulse + RTT/goodput 180 + Timeline 48 brushX + Beam + provenance hash pill`
- **Drawer `ResultatsView`:** `ab-bento 3-col Avant/Après diff pill −60% + table 5–12 groups n small p95 bar grey→green ★ rtt p95 goodput JFI/deadline/wasted/cost toggle + scatter D3 goodput vs small 220 brush rect pareto + provenance`
- **Archives `IntegriteView`:** ordered `Traduction progressive + RDF sha256 chain frozen-wave + runs ul hover peek 4 rows medians + quarantine yellow invalid-only + Figures bar+scatter RDF side-by-side Regen + Replay → Live` inside `max-height calc(100vh -48-28) overflow-y auto` container fixing scroll lock
- **Sheet `CampagneView`:** `Sheet 380px backdrop blur` `Timeline 48 hasData gate + InlineField Site/LinkType/Provider/Duration 10–600 + Import INLINE Id/cap/delay + bento 2-col + Side peek 32×12 + cost preview wasted×cost` `Démarrer → Arm 5s → Sheet auto-minimize to pill`

## 5. Data Flow & Real-time — phases, heterogenous P3, audit split, SSE worker

```
Lab VM reproducible                         Edge audit 3-window split
 tc netem 20/100/600ms distribution + TBF stacked parent 1: leaf-only   laptop→Yas 1.1.1.1 ping 5Hz + Small 4–32K http + bulk BulkSendTo/iperf3
      ↓ Probe collect 400ms loop                      ↓ metrics Summarize + Loss missing vs expected
      ↓ metrics.Summarize Median/P95/IQR + DeadlineOK 1000 + CostAR   ↓ split idle 0–12s vs loaded 22–30s distinct rtt_idle/loaded
      ↓ Gates G0–G7 valid/invalid only → writer Freeze manifest sha256 + quarantine.json → both frozen data/frozen-wave3
      → Scan profile|qdisc|cc median/IQR + Best + hardwareRec single source → figures bar+scatter RDF + report md/csv
                          ↓ both converge
Live 10Hz: OnSnap {phase baseline|charge|recup from campagne, rtt/small/goodput/drops/wasted/cost/deadline, gates, running} → Live.Set
  → Hub 10Hz delta structural Last-Event-ID 2048 backpressure 16→dropped → sse.ts worker retry:2000 frameCount%5+gatesEqual→2Hz
  → live.ts mutable rings 1800 (≈180s one event) lttb400 → useRafLoop 4Hz → ECharts lazy + HD sparklines + Beam + Timeline brush → Wall
  → rings 1800 → Timeline 48 brushX scrubs live.max 1800→60, markArea CHARGE from live.phase, scrubber hidden when empty via dataZoom:[]
Replay /api/replay/list|stream?run 100ms 10Hz SSE → live.rings with banner REPLAY, Sheet pill Event 3/6, Drawer append
```

**Instrumentation 3 seams:** Go `--tags=instrumentation` `TCRunner Run args → kit/logs/tc.log + StatsFn deltas → kit/logs/sse.log ringLen/lastID/dropped` + Frontend `data-testid + data-metric + window.__CGO_LIVE/__CGO_SSE` + Embed `kit/logs/build.log dist sha + kit/logs/embed.log dist sha==buildSha==integrity sha8` for `playwright trace on + clip screenshots 12` surgical diff. Rings mutable off-React via `live.ts` singleton + granular Zustand slices (mix B+C) avoids `10Hz` React churn.

## 6. Error Handling & Empty States — honest emptiness, validation, ledger double

| State | Where | Operator sees | Handling (Wall-kit table) |
|---|---|---|---|
| Loading | Wall/Drawer/Archives 2s | `EmptyState loading + Skeleton 3-line shimmer + Provenance wait n=0` | `fetch null → loading` |
| Empty | Wall hero, Drawer !groups | `EmptyState empty "en attente — Démarrer depuis Campagne pour alimenter le Live" + MetricCard — not 0% red + dataZoom:[] scrubber hidden + Timeline hidden idle hasData rings-empty` | `available:false` honest, no `700` hardcoded |
| Offline | Wall/Campagne | `EmptyState offline "backend injoignable — vérifiez VM 192.168.174.128:9090" + banner OFFLINE red + Toast err` | `EventSource.onerror → Flash` |
| Stale | Wall 5s no frame | `banner STALE amber` | `useInterval 5s ts` |
| Error | POST run/audit/import 409/400 | `EmptyState error + Toast err + Flash danger + shake + Arm retry` | `!r.ok throw→catch setErr+pushToast` |
| Invalid/Degraded | Drawer quar 4 vs 40 | `StatusPip tiny mono PASS/FAIL/— rail only + quarantine yellow invalid-only + table gate_status` | `Scan only invalid` |
| Validation | Sheet Site/link/dur 10–600 + Import Id/cap>0/delay>0 | `InlineField gap2 shake + Arm disabled until valid + helper 'requis/min'` | `validate.ts required/min/max string trim (fixes 0) + animateShake` All+optimistic |

**Validation:** `validate.ts` `required/min/max string trim` + `animateShake` on `InlineField` fail + `ArmButton 5s` disabled until valid + `Esc` clears flash.

**Testing ledger double — debug-before + debug-after:**
- `every commit:` `go vet && go test -tags=instrumentation -timeout 60s -race` + `cd web/frontend && bun run typecheck strict noUnusedLocals && bun run build && node scripts/check-bundle.mjs TOTAL 600KB watchdog echarts 350KB && bunx vitest run jsdom + data-testid`
- `every deploy:` `kit/engine.sh --action build → --action ensure --config kit/cgo-vm.yaml (VMware detect-fallback VirtualBox poll SSH 60×5s) → --action deploy [--public --tunnel-token $CLOUDFLARE_TUNNEL_TOKEN cloudflared] → health + curl -N /api/stream 10Hz ±20% delta once + go test -tags=real veth-c leaf-only 77.7 G4 + playwright wall/sheet/archives clip 8/8 + screenshots 12 1920/1366/390 rail collapsed/expanded + hero + beam + donut + bento tidy + embed analysis dist sha==buildSha==integrity sha8 → kit/logs/e2e.log trace on`
- `full:` `Lab 36 P1/P2/P3×3×2×3 →12 groups overnight + audit 30s split idle→loaded Yas 3-window → link_audit.csv 14 cols real 20→375 + GET /api/hardware/translate?profile=P2 MikroTik + POST /api/figures/regen SVG RDF + GET /api/report/export md/csv + ./cgo verify + a11y land`

## 7. Build/Deploy Kit & Rollout — general great tool

**Kit rename `deploy/` → `kit/`:** `kit/config.sh` (loader) + `kit/engine.sh` (build|scan|boot|bootstrap|ensure|deploy|tunnel|logs) + `kit/bootstrap-vm.sh` + `kit/vm-install.sh` + `kit/cgo-vm.yaml.example` → machine-local `kit/cgo-vm.yaml` gitignored (`hypervisor: auto, vm_name, vmx_path|vbox_path, user altfloat, host 192.168.174.128, key ~/.ssh/id_ed25519, project /home/altfloat/cgo, port 9090, snapshot?`).

- `--action scan [--deep]` `vmrun list + ~/Documents/Virtual Machines + C:/D:/ depth≤3 *.vmx --deep full + VBoxManage list vms + ~/VirtualBox VMs + *.vbox` writeback `vmx_path:`
- `--action ensure` `ssh -o ConnectTimeout=3` else `vmrun/VBoxManage start headless` → `poll SSH 60×5s`
- `--action bootstrap` `iproute2 curl bc tcp_bbr, sysctl net.ipv4.tcp_congestion_control=bbr, veth-c/veth-s pair netem/tbf stacked parent 1: up` idempotent, node not required
- `--action build` **strict** `go vet + bun typecheck + build + check-bundle TOTAL 600KB/echarts 350KB + vitest` fail-fast before `go build -ldflags="-s -w" -o bin/cgo-linux` + `sha256 bin/cgo* → kit/logs/build.log`
- `--action deploy [--public --tunnel-token $CLOUDFLARE_TUNNEL_TOKEN]` **always build first** → `scp cgo-linux.new → install → health /api/health` → `--public` `cloudflared tunnel run` quick tunnel for jury without VM net
- `npm/brew/go` `meteolink` `postinstall.js` ponytail minimal `linux/darwin/win32 x64/arm64 try meteolink-plat-arch then cgo-plat-arch from releases/latest/download chmod 0755 fallback go install` + `bin meteolink → scripts/npm-postinstall.js` + `meteolink top` TUI 8 cards ASCII sparklines same `live rings 1800`

**Rollout — huge layered operation (inline, no subagents):**
Layer 0 **fix splatter** `rail jank C → bento breathing B → scroll lock A` + `TCRunner leaf-only + Instrumentation 3 seams` via `playwright trace + screenshots 12 + kit/logs + embed` prove `nothing renders` fixed (debug-before). Layer 1 **backend** `P3 + audit split B+C + hardware single source + quarantine 4 + Gates` green. Layer 2 **shell** `var(--rail-w) + PanelChooser tri-toggle + Sheet auto-minimize`. Layer 3 **Wall HD** `selective sparklines + Beam/Donut guarded + clean triple strip + hero 300 phase-driven 1800 + D3 Timeline 48`. Layer 4 **kit** `build→ensure→deploy→tunnel→npm general`. Layer 5 **debug-after** whole execution `both lineages frozen manifest sha chain + quarantine 4 + Figures triple-provenance + report md/csv + playwright 8/8 + 12 shots + test-real 77.7` green = robust atelier reassembled, measured, provable.

**Verification before done:** `kit/engine.sh --action build + scan --deep C:/ D:/ + ensure + deploy --public + curl /api/health + curl -N /api/stream | head -n 30 (watch 10Hz) + POST /api/run/start {profiles:["P2"],reps:2} smoke pilot → Wall drawer | npx playwright test --project=chromium baseURL http://192.168.174.128:9090 + SHOT_DIR=shots npx playwright test screenshots | ./cgo verify + make figures`
