# Wall-kit Reunite — NOC Observatory Heavy Reassembly Master Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reassemble splattered atelier furniture into Wall-kit NOC observatory (Wall hybrid lively + Sheet cockpit + Drawer History + Archives) with both lineages Lab P1/P2/P3×3×2×3 + audit 30s 3-window frozen triple-provenance, debug-first + debug-after double surgery inline heavy.

**Architecture:** 5 layered plans executed sequentially inline: **0 fix splatter + 3 seams → 1 backend robust heterogenous + audit split → 2 shell var(--rail-w) + PanelChooser tri-toggle + Sheet → 3 Wall HD selective + Beam/Donut + clean triple strip phase-driven 1800 → 4 kit build-then-embed + VMware/VBox + tunnel + npm/brew + TUI → 5 debug-after whole execution both lineages frozen + playwright trace + 12 shots + embed analysis**. Each layer ships green before next (`go vet + bun typecheck strict + build TOTAL 600/echarts 350 + vitest + playwright clip + test-real 77.7 G4`).

**Tech Stack:** Go 1.25, React 19 Vite 6 Zustand low-freq + live.ts mutable 1800 lttb400, ECharts tree-shaken useDirtyRect + D3 brush + anime 4.5 svg/text/animatable, Vitest jsdom, Playwright 1.62, kit/engine.sh, cloudflared, VBoxManage, woff2

## Global Constraints

- Wall-kit floorplan adopted: Wall / + Sheet Cmd-K 380 + Drawer 56px peek auto-peek + Archives footer → /archives (not 4 equal PANELS)
- Hybrid lively: hero small_p95 300 full-width date pulse + PanelChooser tri-toggle metric groups/chart craft/source live|frozen|both overlay same scale grey dashed vs cyan solid
- Clean ECharts triple strip: dataZoom slider/visualMap watermark blur/LinearGradient removed, markArea CHARGE single, animation:false on Live, thin hairline traces
- Layout C→B→A: rail var(--rail-w) animatable 400 → bento wall scoped airy 24 dense 16 → integrite calc(100vh-48-28) overflow-y auto
- 3 instrumentation seams: Go --tags=instrumentation tc.log + frontend data-testid + window.__CGO_* + embed trace kit/logs/build.log dist sha==buildSha==integrity sha8 gitignored
- Both lineages: Lab P1 80/20/2/0 + P2 20/100/15/0.5 + P3 VSAT 5/600/30/1 distribution normal via profile import JSON/CSV + audit 30s 3-window split idle 0–12s bulk 12–22s loaded 22–30s + iperf3 optional, decently portable not zero-dep
- Receiver-side leaf-only SumBytes last tc -s delta for goodput 77.7 valid G4, SumDrops sum, NsRunner ip netns exec when veth-s
- Metrics Go pure, deadline 1000 fixed, JFI detail=1 gated, CostAR 4.5GiB*30000, gates quarantined only invalid tiny mono rail only, hardwareRec single hardware.go per profile best progressive Sheet>Wall>Archives
- Selective HD: continuous rtt/small/goodput/QDI 60×12 lttb40 clipPath rx4, drops/wasted/cost numeric + trend, DonutJFI — when idle, Beam hover-sync when running guarded anime state changes only
- Worker SSE retry:2000 frameCount%5+gatesEqual→2Hz live 1800 lttb400 useRafLoop 4Hz, honest EmptyState — not 0% red, dataZoom:[] scrubber hidden when empty, Timeline 48 brushX hidden idle via hasData, phase from live.phase
- ArmButton 5s progress line 1px amber + InlineField validation 10–600 0 fix + shake + Esc, Sheet 3-step Kit helpers inline
- Kit general: kit/ not deploy/, hypervisor auto vmware|virtualbox, scan C:/D: depth≤3 *.vmx/*.vbox --deep full, ensure poll SSH 60×5s, bootstrap iproute2 curl bc tcp_bbr veth-c/veth-s parent 1:, build always before deploy, TOTAL 600 echarts 350, --public cloudflared $CLOUDFLARE_TUNNEL_TOKEN not committed, npm postinstall ponytail meteolink top TUI 8 cards ASCII sparklines same live rings 1800
- Overarching NOC observatory C: Wall 8 MetricCards auto-fit 160 + QDI/JFI 1fr1fr + hero 300 + RTT/goodput 180 1fr1fr + Timeline 48 + Drawer bar+scatter D3 pareto + mix B+C extra drops/wasted/cost 180 toggle via PanelChooser
- Inline not subagents, huge layered operation debug 70% — Layer 0 fix before + Layer 5 fix after both via kit/logs + playwright trace + 12 shots 1920/1366/390 + embed analysis + both lineages frozen manifest sha chain quarantine 4 + Figures bar+scatter RDF + Report md/csv + verify
- DESIGN sharp 0 tokens var(--t-*), bento tidy+breathing, a11y land, guarded anime prefersReducedMotion

---

## Layer Index — Separate Plans Per Section (Heavy Layered Operation)

**Why separate plans:** Spec covers 7 sections with independent deliverables (shell vs backend heterogenous vs Wall HD vs kit ops vs provenance). Each layer produces working testable software on its own and ships green before next — as Q30 B debug-first heavy demands. One monolith would be 40+ tasks unreviewable inline. Executing inline, one layer at a time, keeps atelier reassembled floor by floor with `playwright trace + kit/logs` proving each floor before next furniture arrives.

| Layer | Plan File | Delivers | Gate Green |
|---|---|---|---|
| 0 | `2026-08-27-wall-kit-0-splatter-fix.md` | Rail var + bento + scroll lock + honest — + 3 instrumentation seams | `bun typecheck + vitest + playwright wall-splatter clip` |
| 1 | `2026-08-27-wall-kit-1-backend.md` | P3 + audit 3-window + leaf-only + hardware single + gates hidden | `go vet -tags=instrumentation -race + go test` |
| 2 | `2026-08-27-wall-kit-2-shell.md` | App shell 48+rail+wall bento + PanelChooser tri-toggle + Sheet 380 | `bun typecheck + vitest + playwright wall` |
| 3 | `2026-08-27-wall-kit-3-wall-hd.md` | MetricCard selective 60×12 + hero 300 phase-driven 1800 + Beam/Donut clean triple | `bun build TOTAL 600 + vitest + playwright wall clip` |
| 4 | `2026-08-27-wall-kit-4-kit.md` | kit rename + VMware/VBox + cloudflare + npm/brew TUI meteolink top | `kit/engine.sh --action build + scan --deep + ensure + deploy` |
| 5 | `2026-08-27-wall-kit-5-debug-after.md` | Figures triple-provenance + 12 shots + both lineages frozen + report verify | `go test -tags=real + playwright 8/8 + screenshots 12 + embed analysis dist sha==buildSha==integrity sha8` |

**Execution order:** `0 → 1 → 2 → 3 → 4 → 5` strictly sequential inline. Each plan's Task N Produces is next plan's Consumes (e.g., 0's `window.__CGO_*` is 3's `useRafLoop` probe). No subagents — one worktree `agent/wall-kit-reunite` per `AGENTS.md`, conventional commits `feat(kit):, fix(wall):` etc., worktree remove after merge to main per `git worktree remove .worktrees/<slug>`.

**Master verification before done (heavy):**

```
host: go vet ./... && go test ./... -tags=instrumentation -timeout 60s -race
      cd web/frontend && bun run typecheck && bun run build && node scripts/check-bundle.mjs && bunx vitest run
VM:   bash kit/engine.sh --action scan --deep C:/ D:/ --config kit/cgo-vm.yaml
      bash kit/engine.sh --action ensure --config kit/cgo-vm.yaml
      bash kit/engine.sh --action bootstrap --config kit/cgo-vm.yaml
      bash kit/engine.sh --action build && bash kit/engine.sh --action deploy --config kit/cgo-vm.yaml --public --tunnel-token $CLOUDFLARE_TUNNEL_TOKEN
      curl http://192.168.174.128:9090/api/health && curl -N http://192.168.174.128:9090/api/stream | head -n 30 # 10Hz ±20% delta once
      ssh altfloat@192.168.174.128 'cd ~/cgo && make test-real'
      npx playwright test --project=chromium --baseURL=http://192.168.174.128:9090
      SHOT_DIR=shots npx playwright test screenshots # 12 shots 1920/1366/390 rail collapsed/expanded + prompt line + donut + beam + bento tidy
      curl -X POST http://192.168.174.128:9090/api/run/start -d '{"profiles":["P2"],"reps":2}' # smoke pilot → Wall drawer auto-peek
      curl -X POST http://192.168.174.128:9090/api/run/start -d '{"profiles":["P1","P2","P3"],"reps":3}' # full 36 overnight → frozen-wave3
      ./cgo audit --site "Département X" --link-type 5g --duration 30 --target 1.1.1.1 && curl http://192.168.174.128:9090/api/audit/list
      curl -X POST http://192.168.174.128:9090/api/figures/regen && ls data/figures/*.svg && curl http://192.168.174.128:9090/api/report/export?format=md
      ./cgo verify && sha256sum data/runs/*/aqm_eval.csv # hash8 same across Wall/Drawer/Archives/Report
```

**What we don't build:** detector bank / E1-3 356 events / oracle worship / negctl / frozen-wave replay machinery beyond `replay 100ms 10Hz` / API tokens / presentation mode — dies with CONGESTION per `docs/SPEC.md §5`.
