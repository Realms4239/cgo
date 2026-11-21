# METEOLINK UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild `web/frontend` as the ratified brand-led light "chart-paper" instrument in strictly two modes (Cockpit + Labo), per `docs/research/2026-11-20-spatial-spec.md`, with honesty laws and WCAG 2.2 AA enforced by gates.

**Architecture:** One shell grammar, two readings. Cockpit = pinned left DecisionSpine (Answer→Act→Proof-entry) beside a single-screen shared-time-axis TimeField + JournalRail; Labo = pinned CommandHeader + PipelineRail beside four sequential Chapters. Depth via one right-Drawer primitive + InlineDisclosure. All measured values flow through a Float64Array ring (NaN-for-absent) and one ECharts grammar with a shared x-axis and a real 220 ms markLine.

**Tech Stack:** React 19 + Zustand + ECharts (canvas) + animejs; Tailwind v4 exact-pin (`@tailwindcss/vite`) with `@theme` mapped to CSS custom properties; Inter var + JetBrains Mono (self-hosted, tabular-nums); vitest (node default, happy-dom for `*.dom.test.tsx`); Playwright + @axe-core/playwright (unmocked, real Go binary).

**Ratified inputs (read before any task):** `docs/research/2026-11-20-capability-donors.md` (S1), `...-brand-and-method.md` (S2), `...-task-analysis-and-ia.md` (S3), `...-design-principles.md` + `...-design-system.md` + `...-spatial-spec.md` (S5/S6). The spatial spec is the layout contract; do not inherit the old 4-destination geometry.

## Global Constraints

- Strictly 2 modes: `cockpit`, `labo`. French UI copy. No third destination, no duplicate views.
- Absent NEVER zero: `"—"` + French absence word; measured 0 renders `0`; gaps break lines; no fabricated hash/count/cadence.
- No green/verified badge without a real verification act; state never color-only (word + shape); proof-before-claim; fail-closed with a visible, keyboard-reachable reason.
- WCAG 2.2 AA: ≥4.5:1 every text/control pair; ≥3:1 chart marks; 44×44 px targets; `@media (forced-colors: active)`; `prefers-reduced-motion` collapses motion.
- Supported width 890–1920 CSS px: no horizontal overflow, no clipping, no unreachable control.
- Exact pins: zero `^`/`~` in `package.json`, matching `bun.lock`; zero `@latest`; zero hardcoded `C:/cgo`.
- Bundle gz budget: `scripts/check-bundle.mjs` TOTAL ≤ 600 KB, ECharts ≤ 350 KB.
- Licenses: `web/frontend/public/licenses/` ships Inter OFL + JetBrains Mono OFL + ECharts Apache-2.0 NOTICE.
- Zero ledger churn: NEVER edit `docs/CHECKPOINTS.md` or `docs/VERIFY.md`.
- One commit per verified slice; explicit paths; never `git add -A`.
- Light palette tokens (computed AA, from design-system): canvas `#fbfbfb`, surface `#ffffff`, soft `#f3f3f3`, hairline `#e8e8e8`, hairline-strong `#c8c8c8`; text `#333333`/`#5a5a5a`/`#6d6d6d`; brand-teal `#17938a`, brand-teal-ink `#0f6b62`, emerald `#2dcc70`, gold `#f4c910`, slate `#34495e`; semantic danger `#d51600`, ok `#1e7a3c`, ok-fill `#229b4e`, warn `#806400`, warn-mark `#9a7400`, neutral `#6d6d6d`.
- Host gates: no `make` on this host → run `python tools/verify.py` and say so; exit 2 = named gate could not run, NEVER green. Playwright bundled Chromium absent → use Edge via `PWCHROMIUM`; default 30 s e2e timeout fails here → `--timeout=90000`. Serialize evidence runs (QA timing noise).

## Skill Invocation Contract (extensive; frontend-heavy)

**Law:** every named skill is a COMMAND invoked through the Skill tool and followed literally; evidence of usage = the artifact exists + process notes + the exit gate checked. A missing artifact = the skill did not run = the slice is BLOCKED. Frontend skills are MANDATORY on every UI slice, not optional polish.

**Stage skills (contractual sequence):** S1 Explore agents (read-only inventory) · S2 deep-research (Exa-only; firecrawl absent) · S3 brainstorming · S4 dispatching-parallel-agents · S5 design-taste-frontend + impeccable (sequential, main agent) · S6 make-interfaces-feel-better · S7 writing-plans · S8 executing-plans **or** subagent-driven-development + test-driven-development · S9 code-review · S10 verification-before-completion.

**Frontend skills per slice (invoke literally, in addition to the stage skill):**
- **Every UI slice:** `design-taste-frontend` (audit-first; color/shape/theme consistency locks; button+form contrast checks; full state cycles; anti-slop) + `impeccable` (Operate mode; run critique/audit, then polish/harden; typeset/layout/adapt as needed) + `make-interfaces-feel-better` (spacing, type, borders, shadows, motion, hit areas, text wrapping, interaction states; verify 44px / forced-colors / reduced-motion / scrollbars).
- **Slice 2 (shell):** `impeccable layout` + `impeccable typeset`; `make-interfaces-feel-better` (hit areas, optical alignment).
- **Slice 3 (primitives):** `impeccable harden` (states/error/i18n) + `make-interfaces-feel-better` (focus, states, hit areas).
- **Slice 4 (Cockpit):** `design-taste-frontend` + `impeccable` Operate (glance hierarchy) + `make-interfaces-feel-better`.
- **Slice 5 (instrumentation):** `web-perf` (60FPS canvas, rAF off React, INP/LCP/CLS, bundle gz) + `make-interfaces-feel-better` (motion enter/exit, transition scope) + `context7` (ECharts/animejs/vitest API) + `clonedeps` if library internals must be inspected.
- **Slice 6 (Labo):** `impeccable typeset` + `impeccable layout` (dense tables, measure) + `design-taste-frontend` (data-ink) + `make-interfaces-feel-better`.
- **Slice 7 (shared surfaces):** `impeccable clarify` (French copy) + `impeccable harden` + `no-ai-slop` (copy register) + `make-interfaces-feel-better`.
- **Slice 8 (motion/a11y/axe/bundle/captures):** `web-perf` (Core Web Vitals, bundle) + `impeccable audit` (a11y/responsive) + `@axe-core/playwright` + `make-interfaces-feel-better` (scrollbars/forced-colors/reduced-motion).
- **Cross-cutting:** `verification-before-completion` (every "done/verified/passing" claim) + `verification-planning` + `test-driven-development`/`tdd` (red-green-refactor) + `systematic-debugging`/`diagnosing-bugs` on any failure + `simplify` after behavior is understood + `using-git-worktrees` for isolation + `code-review`/`requesting-code-review` at S9 + `technical-writing`/`reviewing-technical-prose` for doc edits.
- **Explicitly NOT used (zero-bloat):** `threejs`/`threejs-webgl` (no decorative WebGL), `sites:*` (not a Qoder Site), `cloudflare/*` (no new services/proxies).

## File Structure

Create: `src/lib/ringBuffer.ts` (+test), `src/components/shell/{Chrome,ModeSwitch,TrustStrip}.tsx`, `src/components/cockpit/{DecisionSpine,AnswerBand,TimeField,JournalRail}.tsx`, `src/components/labo/{CommandHeader,PipelineRail,Chapter,Configurer,Observer,Comparer,Preuves}.tsx`, `src/components/ui/{Drawer,InlineDisclosure,StateSlot,SealInspector,CellLedger,ProfileTile,MetricRow}.tsx`, `src/views/{CockpitView,LaboView}.tsx`, `public/licenses/*`, `e2e/redesign-smoke.spec.ts`, `e2e/redesign-a11y.spec.ts`.
Modify: `package.json`, `bun.lock`, `vite.config.ts`, `tsconfig.json`, `vitest.config.ts`, `src/styles/tokens.css`, `src/styles/index.css`, `src/lib/palette.ts`, `src/lib/sse.ts`, `src/lib/chartGrammar.ts`, `src/store/ui.ts`, `src/App.tsx`, `src/components/ArchiveTable.tsx`, `src/components/AuditDeclare.tsx`, `scripts/check-bundle.mjs`, `docs/superpowers/plans/README.md`.
Delete (as replaced): `TopNav.tsx`, `NavIcon.tsx`, `DashboardView.tsx`, `AuditView.tsx`, `BenchView.tsx`, `EvidenceView.tsx`, `LiveView.tsx`, `ResultatsView.tsx`, `CompareView.tsx`, `InterpretationView.tsx`, `IntegriteView.tsx`, `VerdictHero.tsx`, `PhaseStepper.tsx`, `DonutJFI.tsx`, `EventBadge.tsx`, `Beam.tsx`, `PromptProgressLine.tsx`, `LiveGuide.tsx`, dead anime exports + dead CSS.

---

### Task 1: Tailwind v4 exact-pin foundation (zero visual change)

**Files:** Modify `web/frontend/package.json`, `web/frontend/bun.lock`, `web/frontend/vite.config.ts`, `web/frontend/tsconfig.json`, `web/frontend/vitest.config.ts`, `web/frontend/scripts/check-bundle.mjs`; Create `web/frontend/public/licenses/{INTER-OFL.txt,JETBRAINS-OFL.txt,ECHARTS-APACHE-2.0-NOTICE.txt}`.
**Interfaces:** Produces: exact-pin manifest + Tailwind v4 `@theme` bridge + node/happy-dom vitest split + tsconfig covering `e2e/`+`scripts/*.mjs`. Later tasks import `@/styles/tokens.css` tokens unchanged (zero visual change this task).

- [x] **Step 1: Pin every dependency exact (remove ALL `^`).** Edit `package.json`: set `animejs 4.5.0`, `echarts 5.6.0`, `react 19.2.8`, `react-dom 19.2.8`, `zustand 5.0.15`; dev `@playwright/test 1.62.1`, `@types/node 26.6.6`, `@types/react 19.2.8`, `@types/react-dom 19.2.8`, `@vitejs/plugin-react 4.7.0`, `typescript 5.9.3`, `vite 6.4.3`, `vitest 3.2.7`; add exact `tailwindcss 4.1.11`, `@tailwindcss/vite 4.1.11`, `happy-dom 18.0.1`, `@axe-core/playwright 4.10.2`; remove `jsdom` and `@types/d3`. (Values = current `bun.lock` resolutions; verify against lock before editing.)
- [x] **Step 2: Regenerate lock exact.** Run `cd web/frontend && bun install` then confirm `grep -c '"\^' package.json` returns 0 and `bun.lock` matches.
- [x] **Step 3: Wire Tailwind v4 via Vite plugin (no postcss plugin).** Edit `vite.config.ts`: `import tailwindcss from '@tailwindcss/vite'`; `plugins: [react(), tailwindcss()]`. Add `@import "tailwindcss";` at top of `src/styles/index.css` and an `@theme { --color-canvas: #fbfbfb; ... }` block mapping the Global-Constraints tokens (values verbatim). Keep existing CSS so visual output is unchanged this task.
- [x] **Step 4: tsconfig covers e2e + scripts.** Edit `tsconfig.json` `include: ["src", "e2e", "scripts/*.mjs", "vite.config.ts", "vitest.config.ts", "playwright.config.ts"]`.
- [x] **Step 5: vitest node default + happy-dom for DOM tests.** Edit `vitest.config.ts`: `environment: 'node'`, `environmentMatchGlobs: [['**/*.dom.test.tsx', 'happy-dom']]`.
- [x] **Step 6: Ship licenses.** Create the three license files (Inter OFL text, JetBrains Mono OFL text, ECharts Apache-2.0 NOTICE) under `public/licenses/`; reference them from `index.html` footer link.
- [x] **Step 7: Fix stale bundle message.** Edit `scripts/check-bundle.mjs` failure string to cite 600 KB (not 580).
- [x] **Step 8: Verify zero visual change + gates.** Run `cd web/frontend && npx tsc --noEmit && bun run build && bun run test`; then `python tools/verify.py` (say so: no make). Expected: build ok, tests pass, `6 passed, 0 failed, 0 not run`. Run `git diff | grep -E '\^|@latest|C:/cgo'` → no hits.
- [x] **Step 9: Commit.** `git add web/frontend/package.json web/frontend/bun.lock web/frontend/vite.config.ts web/frontend/tsconfig.json web/frontend/vitest.config.ts web/frontend/scripts/check-bundle.mjs web/frontend/public/licenses web/frontend/src/styles/index.css && git commit -m "build(web): Tailwind v4 exact-pin foundation, licenses, tsconfig/vitest coverage (zero visual change)"`.

### Task 2: Shell + navigation (2 modes) + trust strip

**Files:** Create `src/components/shell/{Chrome,ModeSwitch,TrustStrip}.tsx` + `ModeSwitch.dom.test.tsx`, `TrustStrip.dom.test.tsx`; Modify `src/store/ui.ts`, `src/App.tsx`, `src/styles/index.css`; Delete `TopNav.tsx`, `NavIcon.tsx`.
**Interfaces:** Consumes Task 1 tokens. Produces: `MODES = [{id:'cockpit',label:'Cockpit',key:'1'},{id:'labo',label:'Labo',key:'2'}]`; `useUIStore.mode/setMode`; `<Chrome/>`, `<TrustStrip connection source freshness seal backpressure/>` (words, never dots).

- [x] **Step 1: Failing test for 2-mode switch.** `ModeSwitch.dom.test.tsx`: render, assert exactly 2 buttons labelled Cockpit/Labo, `aria-current` on active, clicking Labo calls `setMode('labo')`, each ≥44px (assert computed min-height). Run `bun run test ModeSwitch` → FAIL.
- [x] **Step 2: Implement ModeSwitch + store.** `store/ui.ts`: replace MODES with the 2-entry const; `ModeSwitch.tsx` segmented control, `aria-current="page"`, keys 1–2 guarded by `isTextEntryTarget`. Run test → PASS.
- [x] **Step 3: Failing test for TrustStrip words.** `TrustStrip.dom.test.tsx`: given `connection='offline'`, assert visible text "HORS LIGNE" (not only a dot); given `backpressure=3` assert "3 trames perdues"; given seal absent assert "ABSENT". Run → FAIL.
- [x] **Step 4: Implement TrustStrip.** Bottom fixed 44px strip, slot order connection·source·freshness·seal·backpressure, each a WORD + shape; abnormal adds saturated left block. Run → PASS.
- [x] **Step 5: Chrome + App shell.** `Chrome.tsx` = logo + ModeSwitch + palette/settings triggers (56px). `App.tsx`: render Chrome + (mode==='cockpit'? CockpitView : LaboView) + TrustStrip + Drawer host; lazy-load the two views; delete TopNav/NavIcon imports.
- [x] **Step 6: Verify.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. e2e overflow at 890/1920 for the shell: `npx playwright test e2e/redesign-smoke.spec.ts --timeout=90000` (create minimal smoke in Task 8; for now assert via `bun run test`).
- [x] **Step 7: Commit.** `git add web/frontend/src/components/shell web/frontend/src/store/ui.ts web/frontend/src/App.tsx web/frontend/src/styles/index.css && git rm web/frontend/src/components/TopNav.tsx web/frontend/src/components/NavIcon.tsx && git commit -m "feat(web): two-mode shell + word-coded trust strip"`.

### Task 3: Shared primitives (Drawer, InlineDisclosure, StateSlot, MetricRow, ProfileTile)

**Files:** Create `src/components/ui/{Drawer,InlineDisclosure,StateSlot,MetricRow,ProfileTile}.tsx` + `Drawer.dom.test.tsx`, `StateSlot.dom.test.tsx`; Modify `src/lib/modalFocus.ts` (Esc topmost-only).
**Interfaces:** Produces: `<Drawer open onClose title width>` (right overlay, focus trap, Esc topmost, focus return, never covers spine ≥1280); `<InlineDisclosure label>` (aria-expanded, pushes content); `<StateSlot label value unit word empty>` (permanent height; absent "—" + word; measured 0 shows 0); `<MetricRow>`; `<ProfileTile tile selected onToggle>` (unclipped, selected=border+check+word).

- [x] **Step 1: Failing StateSlot test.** assert absent renders "—" + word and keeps row height; measured 0 renders "0"; never color-only. Run → FAIL.
- [x] **Step 2: Implement StateSlot + MetricRow.** Run → PASS.
- [x] **Step 3: Failing Drawer test.** assert focus trapped, Esc closes topmost only (stack two drawers), focus returns to trigger, background inert. Run → FAIL.
- [x] **Step 4: Implement Drawer + InlineDisclosure + modalFocus Esc-topmost fix.** Run → PASS.
- [x] **Step 5: ProfileTile unclipped + selected word.** test asserts no `text-overflow: ellipsis`, full name present, selected shows word "sélectionné". Implement. Run → PASS.
- [x] **Step 6: Verify + commit.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. `git add web/frontend/src/components/ui web/frontend/src/lib/modalFocus.ts && git commit -m "feat(web): drawer/disclosure/state-slot primitives with honest absence"`.

### Task 4: Cockpit regions (DecisionSpine + AnswerBand + Act states)

**Files:** Create `src/components/cockpit/{DecisionSpine,AnswerBand}.tsx` + `DecisionSpine.dom.test.tsx`; Create `src/views/CockpitView.tsx`; Modify `src/lib/dashboard.ts` (keep operatorFacts), `src/lib/health.ts` (remove Math.max clamp so improvement shows signed); Delete `DashboardView.tsx`, `AuditView.tsx`, `VerdictHero.tsx`, `PhaseStepper.tsx`.
**Interfaces:** Consumes Task 2 MODES, Task 3 StateSlot/Drawer. Produces: `<CockpitView/>` = DecisionSpine (A1 verdict word, A2 sentence, A3 class health, A4–A7 facts, B Act, C Proof-entry) + TimeField (Task 5) + JournalRail; AnswerBand at <1024.

- [x] **Step 1: Failing test: single verdict slot, no duplicate id.** render CockpitView measured; assert exactly one `#verdict-word`; absent shows "MESURE ABSENTE" neutral; facts absent "—". Run → FAIL.
- [x] **Step 2: Implement DecisionSpine + AnswerBand + CockpitView skeleton (TimeField placeholder slot).** Act slot four exclusive states (idle invitation+declare+Arm / measuring ANNULER / measured gated remediation / quiet sentence). Run → PASS.
- [x] **Step 3: Signed improvement.** test `evaluateLoadedHealth` returns negative bloat when loaded<idle (no clamp); Dashboard shows signed Δ. Implement health.ts change. Run → PASS.
- [x] **Step 4: Verify + commit.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. `git add web/frontend/src/components/cockpit web/frontend/src/views/CockpitView.tsx web/frontend/src/lib/health.ts && git rm .../DashboardView.tsx .../AuditView.tsx .../VerdictHero.tsx .../PhaseStepper.tsx && git commit -m "feat(web): Cockpit decision spine with permanent honest slots"`.

### Task 5: Instrumentation (Float64Array rings, NaN honesty, shared-axis TimeField, 220 ms markLine, JournalRail)

**Files:** Create `src/lib/ringBuffer.ts` + `ringBuffer.test.ts`; Create `src/components/cockpit/{TimeField,JournalRail}.tsx` + `TimeField.dom.test.tsx`; Modify `src/lib/sse.ts`, `src/lib/chartGrammar.ts`, `src/lib/trace.ts`; Delete `LiveView.tsx`, `DonutJFI.tsx`, `EventBadge.tsx`, `Beam.tsx`.
**Interfaces:** Produces: `class TypedMetricRing { constructor(capacity=1800); push(ts:number, v:number|null):void; series():Float64Array; dirty:boolean }` (null→NaN, NEVER 0); `class DomainEventRing { push(e); list() }`; `chartGrammar.baseOption` light + `objectiveMarkLine(ms)`; `<TimeField rings events/>` shared x-axis; `<JournalRail events/>` live during run.

- [x] **Step 1: Failing ring test.** `ringBuffer.test.ts`: push(null) stores NaN (not 0); push(0) stores 0; wrap-around at capacity; dirty flag set. Run `bun run test ringBuffer` → FAIL.
- [x] **Step 2: Implement TypedMetricRing + DomainEventRing (Float64Array).** Run → PASS.
- [x] **Step 3: sse writes rings with NaN.** test: a frame missing rtt_p50 leaves NaN in rtt50 (no fabricated 0); idle frames not pushed. Modify `sse.ts` to write rings directly (no React re-render per tick); store sync at 4 Hz. Run → PASS.
- [x] **Step 4: chartGrammar light + real 220 markLine + gap-breaking.** test: `objectiveMarkLine(220)` present with label; `rawLineSeries` connectNulls:false; palette = light semantic hues. Implement. Run → PASS.
- [x] **Step 5: TimeField + JournalRail.** test: primary trace tallest with markLine; goodput+iface share x-extent; ruler is the shared axis; JournalRail refreshes during run (not frozen). Implement; rAF paint only when ring.dirty; animation:false. Run → PASS.
- [x] **Step 6: Verify + commit.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. `git add web/frontend/src/lib/ringBuffer.ts web/frontend/src/lib/ringBuffer.test.ts web/frontend/src/lib/sse.ts web/frontend/src/lib/chartGrammar.ts web/frontend/src/components/cockpit/TimeField.tsx web/frontend/src/components/cockpit/JournalRail.tsx && git rm .../LiveView.tsx .../DonutJFI.tsx .../EventBadge.tsx .../Beam.tsx && git commit -m "feat(web): Float64Array NaN-honest rings + shared-axis instrument with 220ms line"`.

### Task 6: Labo regions (CommandHeader + PipelineRail + 4 Chapters)

**Files:** Create `src/components/labo/{CommandHeader,PipelineRail,Chapter,Configurer,Observer,Comparer,Preuves}.tsx` + `Comparer.dom.test.tsx`; Create `src/views/LaboView.tsx`; Delete `BenchView.tsx`, `EvidenceView.tsx`, `ResultatsView.tsx`, `CompareView.tsx`, `InterpretationView.tsx`, `IntegriteView.tsx`.
**Interfaces:** Consumes Task 3 Drawer/InlineDisclosure/ProfileTile/CellLedger, Task 5 rings for replay. Produces: `<LaboView/>` = CommandHeader (recipe+state+ARM/STOP pinned) + PipelineRail + Chapters Configurer/Observer/Comparer/Preuves; unmeasured sorts LAST; equality-on-empty refused; seal VÉRIFIÉ only after real verify act.

- [x] **Step 1: Failing test: stop never scrolls + unmeasured last.** render LaboView running; assert ARRÊTER visible (pinned header); leaderboard with an unmeasured row sorts it last with "—"; A/B with empty side shows refusal words not "égalité". Run → FAIL.
- [x] **Step 2: Implement CommandHeader + PipelineRail + Chapter + Configurer + Observer.** Run → PASS.
- [x] **Step 3: Implement Comparer + Preuves + SealInspector (real verify).** Comparer: constat lede → duel → leaderboard → TradeSpace disclosure; Preuves: corpus+verify act → archive → disclosures. SealInspector calls `/api/integrity` and only shows VÉRIFIÉ on a passing hash-verification act. Run → PASS.
- [x] **Step 4: Verify + commit.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. `git add web/frontend/src/components/labo web/frontend/src/views/LaboView.tsx && git rm .../BenchView.tsx .../EvidenceView.tsx .../ResultatsView.tsx .../CompareView.tsx .../InterpretationView.tsx .../IntegriteView.tsx && git commit -m "feat(web): Labo proof document with pinned command header and honest ranking"`.

### Task 7: Shared surfaces (drawers wiring, palette, toasts, ArchiveTable plurals, AuditDeclare NFC)

**Files:** Modify `src/components/ArchiveTable.tsx` (+test), `src/components/AuditDeclare.tsx` (+test), `src/components/Toasts.tsx`, `src/lib/palette.ts`, `src/components/ProofDrawer.tsx` (single component two entries); Create `src/components/ui/CellLedger.tsx` if not in Task 6.
**Interfaces:** Produces: ArchiveTable count `0 entrée / 1 entrée / n entrées`; AuditDeclare site/provider `maxLength={80}` + `.normalize('NFC')`; toasts kind word + dismiss; palette light single house.

- [x] **Step 1: Failing plural test.** `ArchiveTable` with 0/1/3 rows asserts "0 entrée"/"1 entrée"/"3 entrées" (aria-live). Run → FAIL.
- [x] **Step 2: Implement plurals.** Run → PASS.
- [x] **Step 3: Failing NFC/maxLength test.** AuditDeclare site input has maxLength 80 and normalizes NFC on change. Run → FAIL. Implement. Run → PASS.
- [x] **Step 4: Toasts word + dismiss + hover-pause; palette light.** test toast kind carries a word and a dismiss button. Implement; palette.ts returns light tokens. Run → PASS.
- [x] **Step 5: ProofDrawer single component, two entry points; useCopy honest failure (no silent no-op).** test copy failure shows a word when clipboard undefined. Implement. Run → PASS.
- [x] **Step 6: Verify + commit.** `npx tsc --noEmit && bun run build && bun run test`; `python tools/verify.py`. `git add web/frontend/src/components/ArchiveTable.tsx web/frontend/src/components/AuditDeclare.tsx web/frontend/src/components/Toasts.tsx web/frontend/src/lib/palette.ts web/frontend/src/components/ProofDrawer.tsx && git commit -m "feat(web): honest shared surfaces, entrée plurals, NFC declare, light palette"`.

### Task 8: Motion / a11y / axe / bundle / captures

**Files:** Create `e2e/redesign-smoke.spec.ts`, `e2e/redesign-a11y.spec.ts`; Modify `src/styles/index.css` (forced-colors, reduced-motion, scrollbars, remove dead CSS), `src/lib/anime.ts` (remove dead exports), `playwright.config.ts`; Create `web/frontend/slice-capture.mjs`.
**Interfaces:** Consumes all prior. Produces: unmocked Playwright booting real Go binary; axe wcag2a/2aa/21aa/22aa at 890/1280/1920 with 0 critical/serious; 44px bounding boxes; no horizontal overflow; captures at 890/1280/1920.

- [x] **Step 1: forced-colors + reduced-motion + scrollbars CSS.** add `@media (forced-colors: active)` (borders+words, system Highlight focus), reduced-motion collapse, scrollbar thumb/track/hover tokens; delete dead `.gates/.rail-*/.brush` rules. Verify via `bun run test` (tokens/designFloor gates).
- [x] **Step 2: Unmocked smoke e2e.** `redesign-smoke.spec.ts`: webServer boots compiled `cgo --serve --addr 127.0.0.1:19091 --tls=false --insecure-allow-plaintext` with test token + temp data dir; NO `page.route('/api/**')`; assert `/api/health`, switch Cockpit↔Labo, arm+confirm guarded action, reload persists mode. Run `npx playwright test e2e/redesign-smoke.spec.ts --timeout=90000` (Edge via PWCHROMIUM) → PASS.
- [x] **Step 3: axe a11y e2e.** `redesign-a11y.spec.ts`: AxeBuilder tags wcag2a/wcag2aa/wcag21aa/wcag22aa at 890×512, 1280×768, 1920×1080; assert 0 critical/serious, 0 horizontal overflow, ≥44px targets, unclipped tiles. Run → PASS.
- [x] **Step 4: Bundle + captures.** `bun run build` (check-bundle ≤600/350); `node web/frontend/slice-capture.mjs` writes 890/1280/1920 captures; `node web/frontend/scripts/validate-images.mjs`.
- [x] **Step 5: Full gates + forbidden diff.** `python tools/verify.py` → `6 passed, 0 failed, 0 not run`; `git diff | grep -E '\|\| true|C:/cgo|@latest|\^|page\.route'` → no hits.
- [x] **Step 6: Commit.** `git add web/frontend/e2e web/frontend/src/styles/index.css web/frontend/src/lib/anime.ts web/frontend/playwright.config.ts web/frontend/slice-capture.mjs && git commit -m "test(web): unmocked e2e + axe AA + forced-colors/reduced-motion + captures"`.

---

## Dependencies
Task 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 (strictly sequential; each builds on prior interfaces). Task 4's TimeField slot is filled by Task 5; Task 6 replay borrows Task 5 rings. No task spans two scopes.

## Self-review (spec coverage)
2-mode shell (T2), trust words (T2), primitives/drawer/state-slot (T3), Cockpit spine/Act/absence (T4), rings NaN + shared axis + 220 line + journal live (T5), Labo header/rail/chapters + honest ranking + real verify (T6), plurals/NFC/toasts/palette/proof (T7), forced-colors/reduced-motion/scrollbars/axe/bundle/captures (T8). Licenses+pins+tsconfig+vitest (T1). Zero ledger churn honored (no CHECKPOINTS/VERIFY edits). Every step has a verbatim run command; one commit per slice.