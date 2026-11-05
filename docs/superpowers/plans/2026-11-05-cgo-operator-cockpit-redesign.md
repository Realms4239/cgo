# CGO — Operator Cockpit Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the four Meteolink destinations as a usable, honest desktop cockpit for a DSI operator, without inventing measurements, dressing absence as success, or changing the measurement engine.

**Architecture:** A shared shell/token layer provides one 12-column layout, Inter prose, contact-scale geometry and common loading/absent/error states. Each destination consumes a strict discriminated-union view state; the dashboard consumes a class-aware health helper; the archive consumes a windowed table; all behaviour is proven by component, integration and Playwright tests rather than green snapshots. The Unix boundary remains intact: Go owns measurement, the browser renders evidence and never invents a value.

**Tech Stack:** React 19, TypeScript, Zustand 5, ECharts, Vitest, React Testing Library, Playwright, CSS custom properties, existing Python `tools/verify.py` gates.

## Global Constraints

- Supported desktop/laptop range is **890–1920 px CSS width**; tested at **890×768**, **1280×800** and **1920×1080**.
- Mobile 390×844 is out of scope for this slice; existing mobile CSS must not regress while this range is widened.
- Layouts use a 12-column grid with 16–24 px gutters; at 890 px cards and toolbars wrap with `flex-wrap`, never clip or overflow horizontally.
- Inter is used for prose; JetBrains Mono is used only for values, timestamps, identifiers, data, commands and counts.
- A missing measurement is **absent**, never a zero, pass, cost or verdict. A stale measurement must be labelled historical; an archived item is never presented as current.
- The audit is strictly **60 seconds**: 10 s Repos → 40 s Charge → 10 s Vidange. No 20-second audit option and no reduced-load path on the Auditer screen.
- A separate, explicitly named `Test de connectivité` may exist outside the Auditer route; it is not an audit and cannot produce a verdict, report, remediation or archive citation.
- CAP/health verdicts are class-aware (`Δp95 = p95(loaded) − p50(base)`), never a blanket `220 ms`. Initial red boundaries: fibre `> 30 ms`, cellular `> 100 ms`, geostationary VSAT `> 150 ms`; VSAT reports explain a 220 ms target is physically incompatible with a 600 ms propagation floor. LEO/P4 values come from its configured profile.
- All existing ledgers, provenance records and slice-22 remediation boundaries remain intact; no audit, bench, topology or corpus mutation belongs in this slice.
- No live audit, bench, CHR, eBPF, `.dist`, BBRv3, RouterOS or network traffic is allowed as evidence. Existing tests use fixtures only.
- No measurement, verdict, SHA, byte count, cost or alert percentage may be invented for a rendered state.
- Global keyboard shortcuts 1–4 ignore `input`, `textarea`, `select`, content-editable targets and events originating in a dialog.
- Every destination and the shared archive export a strict discriminated union; parallel `isLoading` / `isError` / `data` booleans are forbidden.
- Archived rows are virtualized or CSS-contained; details/proofs are fetched lazily only for the opened row; no per-row request waterfall on first paint.
- `docs/img/shots/2026-11-05/` contains only real captures from the app at the three viewports; no fabricated passing screenshots.
- No new dependency is added without explicit human approval. `@tanstack/react-virtual` is not required; use CSS containment or the project-local windowing helper.
- French copy for operator UI. English names for source files, functions, tests and spec language. No cast between isolated hidden windows and the sealed audit.

## Scope boundaries

### In

- Shared shell, grid, tokens, interaction states and focus helpers.
- Tableau de bord, Auditer, Banc d’essai, Preuves & Rapports.
- Class-aware presentation helper (presentation only, sourced from report payloads or explicit UI fixtures).
- Virtualized/contained archive table and lazy detail opening.
- Updates to unit/integration/Playwright tests and `DESIGN.md`, `PRODUCT.md`.
- Independent verifier and ledger only after the implementation passes.

### Out

- Measurement engine, remediation generation, CHR/QEMU, live Linux bench, RouterOS, BBRv3, `.dist` and eBPF.
- A new backend endpoint or schema; implementation uses existing list and report APIs.
- Adding monitoring, alerts to mail/SMS/webhook, or external telemetry.
- Mobile layout redesign.
- Tokenizing all console surfaces or rewriting existing bench charts.
- Any claim that a sealed report proves what it does not contain.

---

## Source map

### Existing boundaries to preserve

- `web/frontend/src/store/audit.ts` — audit session owner; tests use it via mocks.
- `web/frontend/src/store/ui.ts` — mode taxonomy (`MODES`), toast/flash, proof drawer.
- `web/frontend/src/store/bench.ts` — bench selection contract; do not fork it.
- `web/frontend/src/lib/auditFlow.ts` — stage machine and duration helpers; 60-second contract owns `Repos`, `Charge`, `Vidange`.
- `web/frontend/src/lib/dashboard.ts`, `auditSummary.ts` — report summary and comparison inputs.
- `web/frontend/src/lib/lexicon.test.ts` — operator wording gate; keep default surfaces off the banned lexicon.
- `web/frontend/src/lib/designFloor.test.ts` — source-tree guard; update only for deliberate, documented design decisions.

### New boundaries

- `web/frontend/src/lib/viewState.ts` — discriminated unions and pure selectors.
- `web/frontend/src/lib/health.ts` — class-aware `Δp95` verdict helper, fixture-driven.
- `web/frontend/src/lib/archive.ts` — archive row typing and window calculator.
- `web/frontend/src/components/ArchiveTable.tsx` — accessible, windowed archive table.
- `web/frontend/src/components/PhaseStepper.tsx` — accessible phase indicator (if the existing `Timeline` cannot support the state contract cleanly).
- `web/frontend/e2e/operator-cockpit.spec.ts` — real-browser behaviour evidence.

---

### Task 1: Strict view-state unions and selectors

**Files:**
- Create: `web/frontend/src/lib/viewState.ts`
- Create: `web/frontend/src/lib/viewState.test.ts`
- Modify: `web/frontend/src/store/audit.ts` (add read-only selector import if no store shape changes)
- Test: `web/frontend/src/lib/viewState.test.ts`

**Interfaces:**
- Consumes: `AuditStage`, `AuditLive`, `AuditReport` from `src/lib/` modules.
- Produces:
  - `AuditSummary`-compatible references unchanged.
  - `export type AuditViewState = ...`
  - `export function selectAuditViewState(input: { stage: AuditStage; live: AuditLive | null; report: AuditReport | null; failure: AuditRequestFailure | null }): AuditViewState`
  - `export type RequestFailure = { code: string; message: string }`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, expect, it } from 'vitest'
import { selectAuditViewState } from './viewState'

const report = { audit_id: 'a1' } as any

describe('viewState', () => {
  it('a completed report cannot coexist with a running state', () => {
    const s = selectAuditViewState({
      stage: 'observing',
      live: { running: true, phase: 'charge', elapsed_sec: 11, total_sec: 60 } as any,
      report,
      failure: null,
    })
    expect(s.status).toBe('running')
    expect('report' in s).toBe(false)
  })

  it('a report becomes complete only when the stage closes and report exists', () => {
    const s = selectAuditViewState({ stage: 'done', live: null, report, failure: null })
    expect(s).toEqual({ status: 'complete', report })
  })

  it('a request failure never exposes a previous verdict as current', () => {
    const s = selectAuditViewState({
      stage: 'observing',
      live: null,
      report: null,
      failure: { code: 'HTTP_500', message: 'lecture impossible' },
    })
    expect(s).toEqual({
      status: 'error',
      code: 'HTTP_500',
      message: 'lecture impossible',
      retryAction: expect.any(Function),
    })
  })
})
```

- [ ] **Step 2: Run the test and confirm it fails**

Run: `cd web/frontend && npx vitest run src/lib/viewState.test.ts`
Expected: FAIL — module `./viewState` does not exist.

- [ ] **Step 3: Implement the strict union and selectors**

```ts
// src/lib/viewState.ts
import type { AuditStage } from './auditFlow'
import type { AuditLive, AuditReport } from './types'

export interface RequestFailure { code: string; message: string; retry?: () => void }

export type AuditViewState =
  | { status: 'idle'; lastAudit: AuditReport | null }
  | { status: 'loading_stale'; previous: AuditReport; startedAt: number }
  | { status: 'running'; phase: 'repos' | 'charge' | 'vidange'; elapsedSec: number; progressPct: number }
  | { status: 'error'; code: string; message: string; retryAction: () => void }
  | { status: 'complete'; report: AuditReport }

const noop = () => undefined

function toPhase(value: string | undefined): 'repos' | 'charge' | 'vidange' {
  if (value === 'drain') return 'vidange'
  if (value === 'charge-down' || value === 'charge') return 'charge'
  return 'repos'
}

export function selectAuditViewState(input: {
  stage: AuditStage
  live: AuditLive | null
  report: AuditReport | null
  failure: RequestFailure | null
}): AuditViewState {
  if (input.failure) {
    return {
      status: 'error',
      code: input.failure.code,
      message: input.failure.message,
      retryAction: input.failure.retry ?? noop,
    }
  }
  if (input.stage === 'done' && input.report) {
    return { status: 'complete', report: input.report }
  }
  if (input.stage === 'observing' && input.live) {
    const total = input.live.total_sec ?? 60
    const elapsed = Math.max(0, Math.min(total, input.live.elapsed_sec ?? 0))
    return {
      status: 'running',
      phase: toPhase(input.live.phase),
      elapsedSec: elapsed,
      progressPct: Math.round((elapsed / total) * 100),
    }
  }
  if (input.stage === 'observing' && input.report) {
    // A prior report is historical until the closure state arrives.
    return { status: 'loading_stale', previous: input.report, startedAt: Date.now() }
  }
  return { status: 'idle', lastAudit: input.report }
}
```

- [ ] **Step 4: Run test and confirm it passes**

Run: `cd web/frontend && npx vitest run src/lib/viewState.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/viewState.ts web/frontend/src/lib/viewState.test.ts
git commit -m 'feat(web): ajoute les unions d’état de vue'
```

---

### Task 2: Class-aware health verdict helper

**Files:**
- Create: `web/frontend/src/lib/health.ts`
- Create: `web/frontend/src/lib/health.test.ts`
- Test: `web/frontend/src/lib/health.test.ts`

**Interfaces:**
- Produces: `evaluateLoadedHealth(input: HealthInput): HealthVerdict`; `healthCopy(verdict): string`; `type HealthInput = { linkType: string; idleP50Ms?: number; loadedP95Ms?: number; targetMs?: number | null }`; `type HealthVerdict = { kind: 'absent' } | { kind: 'ok'; bloatMs: number } | { kind: 'attention'; bloatMs: number } | { kind: 'critical'; bloatMs: number }`.
- Consumers always show `absent` if either input is missing; the UI never substitutes a zero.

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, expect, it } from 'vitest'
import { evaluateLoadedHealth } from './health'

describe('class-aware health', () => {
  it('fiber at +41ms bloat is not healthy simply because it stays below 220ms', () => {
    expect(evaluateLoadedHealth({ linkType: 'fiber', idleP50Ms: 20, loadedP95Ms: 61 }))
      .toEqual({ kind: 'critical', bloatMs: 41 })
  })

  it('cellular uses its own red boundary', () => {
    expect(evaluateLoadedHealth({ linkType: '4g', idleP50Ms: 100, loadedP95Ms: 201 }))
      .toEqual({ kind: 'critical', bloatMs: 101 })
  })

  it('vsat is not failed by its 600ms propagation floor', () => {
    expect(evaluateLoadedHealth({ linkType: 'vsat', idleP50Ms: 600, loadedP95Ms: 720 }))
      .toEqual({ kind: 'attention', bloatMs: 120 })
  })

  it('absent measurements remain absent', () => {
    expect(evaluateLoadedHealth({ linkType: 'fiber' }))
      .toEqual({ kind: 'absent' })
  })
})
```

- [ ] **Step 2: Run the test and confirm it fails**

Run: `cd web/frontend && npx vitest run src/lib/health.test.ts`
Expected: FAIL — module `./health` does not exist.

- [ ] **Step 3: Implement helper**

```ts
// src/lib/health.ts
export type HealthInput = {
  linkType: string
  idleP50Ms?: number
  loadedP95Ms?: number
  targetMs?: number | null
}

export type HealthVerdict =
  | { kind: 'absent' }
  | { kind: 'unconfigured'; bloatMs: number }
  | { kind: 'ok'; bloatMs: number }
  | { kind: 'attention'; bloatMs: number }
  | { kind: 'critical'; bloatMs: number }

type LinkClass = 'fiber' | 'cellular' | 'geo' | 'leo' | 'other'
const RED_BLOAT_MS: Record<LinkClass, number | null> = {
  fiber: 30,
  cellular: 100,
  geo: 150,
  leo: null,
  other: null,
}

function classOf(linkType: string): LinkClass {
  const v = linkType.toLowerCase()
  if (v === 'fiber') return 'fiber'
  if (v === '4g' || v === '5g') return 'cellular'
  if (v === 'vsat') return 'geo'
  return 'other'
}

export function evaluateLoadedHealth(input: HealthInput): HealthVerdict {
  if (input.idleP50Ms == null || input.loadedP95Ms == null) return { kind: 'absent' }
  const bloatMs = Math.max(0, input.loadedP95Ms - input.idleP50Ms)
  const cls = classOf(input.linkType)
  const red = RED_BLOAT_MS[cls]
  if (red == null) return { kind: 'unconfigured', bloatMs } // P4/unknown: no invented verdict
  if (bloatMs > red) return { kind: 'critical', bloatMs }
  if (bloatMs > red * 0.8) return { kind: 'attention', bloatMs }
  return { kind: 'ok', bloatMs }
}
```

- [ ] **Step 4: Run the tests**

Run: `cd web/frontend && npx vitest run src/lib/health.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/health.ts web/frontend/src/lib/health.test.ts
git commit -m 'feat(web): compare la qualité selon la classe de lien'
```

---

### Task 3: Shell, 890–1920 grid and shared surface tokens

**Files:**
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/components/TopNav.tsx`
- Modify: `web/frontend/src/styles/tokens.css`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/styles/components.css`
- Test: `web/frontend/src/lib/tokens.test.ts`

**Interfaces:**
- Produces `.layout-shell`, `.layout-grid`, `.pane`, `.state-card`, `.state-absent`, `.state-error`, `.surface-elevated`.
- Consumers use `className="layout-grid"`; no fixed `max-width: 1180px` remains for operator pages.

- [ ] **Step 1: Write the failing source-token assertions**

```ts
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const tokens = readFileSync('web/frontend/src/styles/tokens.css', 'utf8')
const index = readFileSync('web/frontend/src/styles/index.css', 'utf8')

describe('desktop grid', () => {
  it('declares an 890–1920 range and flex wrapping', () => {
    expect(index).toContain('flex-wrap: wrap')
    expect(tokens).toContain('--grid-gap:')
  })
  it('keeps Inter and mono division', () => {
    expect(tokens).toContain("--font-body: 'Inter'")
    expect(tokens).toContain('--font-mono:')
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/lib/tokens.test.ts`
Expected: FAIL — `--grid-gap` / new layout classes not present.

- [ ] **Step 3: Replace fixed-width shell with 12-column layout**

```css
/* styles/index.css */
#shell { grid-template-rows: auto minmax(0, 1fr) auto; }
main { overflow-y: auto; }
.layout-shell { width: min(100%, 1680px); margin: 0 auto; padding: clamp(16px, 2vw, 32px); }
.layout-grid { display: grid; grid-template-columns: repeat(12, minmax(0, 1fr)); gap: 16px; }
@media (min-width: 1600px) { .layout-grid { gap: 24px; } }
.pane, .state-card, .archive-row { min-width: 0; box-sizing: border-box; }
@media (max-width: 1100px) { .layout-grid { display: flex; flex-direction: column; } .layout-action, .layout-nav, .layout-card { flex: 1 1 100%; } }
```

Set tokens:

```css
/* styles/tokens.css additions */
--grid-gap: 16px;
--surface-1: #10151c;
--surface-2: #151c24;
--canvas: #090d12;
--text-primary: #f5f7fa;
--text-body: #c9d0d9;
--focus: #5ad3e3;
```

Update `App.tsx` to wrap `<main id="main">` content in `.layout-shell` or apply `.layout-shell` to the existing `<main>`. Replace the header link/action placement and remove the global `max-width: 1180px` from `.panel-stack`; keep `#evidence-*` scroll margins.

- [ ] **Step 4: Run tests/build**

Run: `cd web/frontend && npx vitest run src/lib/tokens.test.ts src/lib/designFloor.test.ts && npm run typecheck`
Expected: PASS; suite must show no stale test requiring 1024–1920.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/App.tsx web/frontend/src/components/TopNav.tsx web/frontend/src/styles/tokens.css web/frontend/src/styles/index.css web/frontend/src/styles/components.css web/frontend/src/lib/tokens.test.ts
git commit -m 'feat(web): adopte la grille écran 890–1920'
```

---

### Task 4: Tableau de bord — absent, measured, stale and failed states

**Files:**
- Modify: `web/frontend/src/views/DashboardView.tsx`
- Create: `web/frontend/src/views/DashboardView.test.tsx`
- Modify: `web/frontend/src/lib/dashboard.ts` only if `retailBound()` needs a missing-data guard
- Modify: `web/frontend/src/lib/auditSummary.ts` only for state-mapping text helpers
- Test: `web/frontend/src/views/DashboardView.test.tsx`

**Interfaces:**
- Consumes: `selectAuditViewState()`, `evaluateLoadedHealth()`, `summaryOf()`, `compareAudits()`, `formatCostAR()`.
- Produces: `data-testid="dash-absent" | "dash-gauge" | "dash-action" | "dash-recent"`.
- Dashboard never renders `0 ms`, `0 Mbit/s`, a green verdict or a copy-script button unless the loaded report passes validation.

- [ ] **Step 1: Write failing state tests**

```tsx
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import DashboardView from './DashboardView'

vi.mock('../store/audit', () => ({
  useAuditStore: (selector: any) => selector({ report: null, previous: null, showReport: () => {} }),
}))
vi.mock('../store/ui', () => ({
  useUIStore: (selector: any) => selector({ setMode: () => {} }),
}))

describe('DashboardView absent', () => {
  it('renders one CTA and no fake zero', () => {
    render(<DashboardView />)
    expect(screen.getByRole('button', { name: /audit/i })).toBeInTheDocument()
    expect(document.body.textContent).not.toMatch(/0\s*(ms|Mbit\/s|Ar)/i)
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/views/DashboardView.test.tsx`
Expected: FAIL — current output has VerdictHero empty/absence copy rather than `dash-absent` contract.

- [ ] **Step 3: Rewrite the dashboard hierarchy**

Render in this exact order:

1. Header line: site/link context from report; otherwise `Site — mesure en attente` (not `site —`).
2. If `report == null`, render `.state-card.state-absent`:
   - Heading: `Ce site n’a pas encore été mesuré`
   - Body: `L’audit de 60 secondes mesure la latence au repos puis sous charge; il prouve ou échoue, jamais l’inverse.`
   - Primary button: `Lancer l’audit de 60 secondes`; onClick → `setMode('audit')`.
   - Secondary button only if `/api/audit/list` returns a report and the fetch succeeds: `Voir un rapport archivé`.
3. If `report == null` failed/load error, show `.state-error` with `Lecture impossible`, a retry action and `Détail` proof drawer; never reuse prior metrics.
4. If `report != null`, render three cards:
   - `Santé du lien`: `evaluateLoadedHealth(...)` + `Délai chargé Δ … ms`; absent if inputs missing.
   - `Alertes à l’heure`: configured target from report, never hard-coded `220 ms`; unavailable if absent.
   - `Capacité gaspillée`: formatted Ariary only if economic payload exists.
5. One `Action recommandée` card. If no valid remediation exists, include one safe next step; do not render a ready script.
6. Recent archive table (3–5 rows): date, site, link, status, `Ouvrir` with `audit_id`; no raw `run-…` labels in primary UI.
7. `stale` data is badged `Données datées · <age>`; freshness source is shown inside the proof drawer.

Keep `RemediationCard` only behind a `status === 'complete'` route or remove it from this screen after migrating its content into the action card; do not let it render when a load error exists.

- [ ] **Step 4: Run tests and update lexicon**

Run: `cd web/frontend && npx vitest run src/views/DashboardView.test.tsx src/lib/lexicon.test.ts && npm run typecheck`
Expected: PASS; banned lexicon remains absent on default route.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/DashboardView.tsx web/frontend/src/views/DashboardView.test.tsx web/frontend/src/lib/dashboard.ts web/frontend/src/lib/auditSummary.ts
git commit -m 'feat(web): recompose le tableau de bord en cockpit'
```

---

### Task 5: Audit fixed to 60 seconds, full refactor of the declaration flow

**Files:**
- Modify: `web/frontend/src/lib/auditFlow.ts`
- Modify: `web/frontend/src/lib/auditFlow.test.ts`
- Modify: `web/frontend/src/components/AuditDeclare.tsx`
- Modify: `web/frontend/src/components/AuditDeclare.dom.test.tsx` (create if missing)
- Modify: `web/frontend/src/store/audit.ts`
- Modify: `web/frontend/src/components/AuditLive.tsx` only for state-class wiring

**Interfaces:**
- Produces: `AUDIT_DURATIONS = [60] as const`; `auditDuration(): 60`; default store `duration: 60`.
- No user-facing 10–600 seconds control on the Auditer screen. The engine’s CLI-level `duration` may remain for other tools, but the UI cannot select it.

- [ ] **Step 1: Write failing flow tests**

```ts
import { expect, it, describe } from 'vitest'
import { AUDIT_DURATIONS, validateAuditDeclaration, auditPhasePlan } from './auditFlow'

describe('audit duration contract', () => {
  it('offers only the 60-second audit', () => {
    expect(AUDIT_DURATIONS).toEqual([60])
    expect(validateAuditDeclaration({ site: 'A', link_type: 'fiber', duration: 60, direction: 'up' }).valid).toBe(true)
    expect(validateAuditDeclaration({ site: 'A', link_type: 'fiber', duration: 20, direction: 'up' }).valid).toBe(false)
  })
  it('never accounts for a charge-less plan', () => {
    expect(auditPhasePlan(60)).toEqual({ idle: 10, loaded: 40, drain: 10 })
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/lib/auditFlow.test.ts`
Expected: FAIL — `AUDIT_DURATIONS` still contains `[20, 60]`.

- [ ] **Step 3: Make the fixed-duration contract explicit**

```ts
// lib/auditFlow.ts
export const AUDIT_DURATIONS = [60] as const
export const AUDIT_DURATION_SECONDS = 60 as const
export function auditDuration(): 60 { return AUDIT_DURATION_SECONDS }
export function auditPhasePlan(): AuditPhasePlan { return { idle: 10, loaded: 40, drain: 10 } }
```

Update `validateAuditDeclaration()` so a non-60 duration returns `{ valid: false, errors: { duration: 'audit limité à 60 s' }, value: null }`. Delete `durationChoice` / `DurationChoice` or keep them exported only if another caller imports them; first run `grep -n "durationChoice\|AUDIT_DURATIONS\|DEFAULT_AUDIT_DIRECTION" -R src` and remove all unused references.

In `AuditDeclare.tsx`:

1. Remove the duration pill group and its `<InlineField label="Durée">`.
2. Replace it with a static fact line: `Audit unique : 10 s au repos → 40 s sous charge → 10 s de drainage.`
3. Remove “Durée libre (s)” from advanced. Keep direction fixed unless the engine still requires it; direction can remain as advanced only if expanding scope is accepted, otherwise fix direction `up` inherited from the product contract.
4. `site` input keeps `type="text"` and no numeric shortcut affinity.
5. Presets that submit `duration: 20` are filtered by `applyPreset()` to refuse an audit; display `Preset incompatible — audit fixe 60 s` instead of silently changing it.
6. `useAuditStore` initializes `duration: 60`; all `patch({ duration: ... })` UI paths are deleted.
7. The start request sends `duration: 60`; nothing else changes in measurement code.

- [ ] **Step 4: Run tests/build**

Run: `cd web/frontend && npx vitest run src/lib/auditFlow.test.ts src/components/AuditDeclare.dom.test.tsx src/lib/lexicon.test.ts && npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/auditFlow.ts web/frontend/src/lib/auditFlow.test.ts web/frontend/src/components/AuditDeclare.tsx web/frontend/src/store/audit.ts web/frontend/src/components/AuditLive.tsx
git commit -m 'feat(web): ferme la voie audit 60 secondes'
```

---

### Task 6: Audit state rendering, confirm-first start and phase stepper

**Files:**
- Modify: `web/frontend/src/views/AuditView.tsx`
- Modify: `web/frontend/src/components/AuditLive.tsx`
- Create: `web/frontend/src/components/PhaseStepper.tsx`
- Create: `web/frontend/src/components/PhaseStepper.dom.test.tsx`
- Modify: `web/frontend/src/components/ArmButton.tsx` only for stable focus behavior if needed
- Test: `web/frontend/src/components/PhaseStepper.dom.test.tsx`

**Interfaces:**
- Consumes: `selectAuditViewState()`; `useAuditStore`; `AuditLive`.
- Produces: `PhaseStepper` with `role="group"` and `aria-current="step"`; Audit destination accepts only `idle | loading_stale | running | error | complete`, no mixed Boolean props.

- [ ] **Step 1: Write failing DOM tests**

```tsx
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import PhaseStepper from './PhaseStepper'

describe('PhaseStepper', () => {
  it('does not mark an unmeasured phase complete', () => {
    render(<PhaseStepper phase="charge" />)
    expect(screen.getByLabelText('Charge')).toHaveAttribute('aria-current', 'step')
    expect(screen.getByLabelText('Repos')).toHaveAttribute('data-state', 'done')
  })
  it('renders a failed load without a success state', () => {
    render(<PhaseStepper phase="vidange" error="lecture impossible" />)
    expect(screen.getByRole('status')).toHaveTextContent('lecture impossible')
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/components/PhaseStepper.dom.test.tsx`
Expected: FAIL — component does not exist.

- [ ] **Step 3: Implement view rendering and avoid mixed state**

`AuditView` switch:

```tsx
const snapshot = selectAuditViewState({
  stage: s.stage,
  live: s.live,
  report: s.report,
  failure: s.failure ?? null,
})

switch (snapshot.status) {
  case 'idle': return <DeclareView lastAudit={snapshot.lastAudit} />
  case 'loading_stale': return <PreviousReportAsStale previous={snapshot.previous} />
  case 'running': return <RunView state={snapshot} />
  case 'error': return <StateError message={snapshot.message} retry={snapshot.retryAction} />
  case 'complete': return <ReportView report={snapshot.report} />
}
```

If the store does not yet expose `failure`, add one plain object `{ code, message } | null` and set it in the poll catch callback; do not overload `msg` with parsing logic. All controls disabled while `running`; the “Ouvrir le constat” action is unavailable until backend produces `complete`. `PhaseStepper` supports `repos`, `charge`, `vidange` and a visible failed marker. The state-error screen offers `Réessayer` and a collapsed `Détail technique`; audit reports cannot show a green result before close.

- [ ] **Step 4: Run DOM and type tests**

Run: `cd web/frontend && npx vitest run src/components/PhaseStepper.dom.test.tsx src/lib/viewState.test.ts src/lib/lexicon.test.ts && npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/views/AuditView.tsx web/frontend/src/components/PhaseStepper.tsx web/frontend/src/components/PhaseStepper.dom.test.tsx web/frontend/src/components/AuditLive.tsx web/frontend/src/store/audit.ts
git commit -m 'feat(web): unifie les états de la voie audit'
```

---

### Task 7: Bench profile cards, no truncation, non-research visible path

**Files:**
- Modify: `web/frontend/src/components/BenchConfig.tsx`
- Modify or create: `web/frontend/src/components/ProfileTiles.tsx`
- Create: `web/frontend/src/components/ProfileTiles.dom.test.tsx`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/lib/profiles.ts` only to expose `PROFILE_DESCRIPTIONS` if absent
- Test: `web/frontend/src/components/ProfileTiles.dom.test.tsx`

**Interfaces:**
- Produces: `PROFILE_TILES = [{ id:'P1', title:'P1 — Fibre dédiée', meta:'80 Mbit/s · 20 ms RTT', role:'Référence fibre' }, ...]`.
- Tiles are buttons/checkbox cards with `aria-pressed`; no `text-overflow: ellipsis` on names.
- Imported profiles render in a separate `Profils importés` row and never select automatically.

- [ ] **Step 1: Write failing truncation test**

```tsx
import { render, screen } from '@testing-library/react'
import { expect, describe, it } from 'vitest'
import ProfileTiles from './ProfileTiles'

const profiles = [
  { id: 'P1', imported: false }, { id: 'P2', imported: false },
  { id: 'P3', imported: false }, { id: 'P4', imported: false },
] as any

describe('ProfileTiles', () => {
  it('shows full profile names and never clips them', () => {
    render(<ProfileTiles profiles={profiles} selected={['P2']} onToggle={() => {}} />)
    expect(screen.getByText(/Fibre dédiée/i)).toBeInTheDocument()
    expect(screen.getByText(/20 ms RTT/i)).toBeInTheDocument()
    expect(document.body.innerHTML).not.toContain('P1FIBRE')
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/components/ProfileTiles.dom.test.tsx`
Expected: FAIL — component does not exist.

- [ ] **Step 3: Implement four non-truncating tiles**

```tsx
const TILES = [
  ['P1', 'Fibre dédiée', '80 Mbit/s · 20 ms RTT', 'Référence fibre'],
  ['P2', '4G dégradée', '20 Mbit/s · 100 ms RTT', 'Cellulaire contrainte'],
  ['P3', 'VSAT géostationnaire', '5 Mbit/s · 600 ms RTT', 'Satellite'],
  ['P4', 'Orbite basse', '100 Mbit/s · 40 ms RTT', 'LEO'],
] as const
// render each card with id, full title, measured profile values, aria-pressed,
// and a keyboard-selectable control. CSS: grid auto-fit/minmax; no ellipsis.
```

Move profile names/rates into `lib/profiles.ts`; never hard-code a measurement where report data disagrees with the fixture definitions. Keep visual simulation details under a closed `Détails de la simulation` disclosure; research mode still exists, but no profile flow is hidden behind it.

- [ ] **Step 4: Run tests and build**

Run: `cd web/frontend && npx vitest run src/components/ProfileTiles.dom.test.tsx src/components/BenchConfig.dom.test.tsx src/lib/benchLayout.test.ts && npm run typecheck`
Expected: PASS; CSS tree has no `text-overflow: ellipsis` in profile labels.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/ProfileTiles.tsx web/frontend/src/components/ProfileTiles.dom.test.tsx web/frontend/src/components/BenchConfig.tsx web/frontend/src/lib/profiles.ts web/frontend/src/styles/index.css
git commit -m 'feat(web): expose quatre cartes profil lisibles'
```

---

### Task 8: Archive table — windowing, lazy proofs, no request waterfall

**Files:**
- Create: `web/frontend/src/lib/archive.ts`
- Create: `web/frontend/src/lib/archive.test.ts`
- Create: `web/frontend/src/components/ArchiveTable.tsx`
- Create: `web/frontend/src/components/ArchiveTable.dom.test.tsx`
- Modify: `web/frontend/src/components/AuditReportsPanel.tsx`
- Modify: `web/frontend/src/views/EvidenceView.tsx` only for anchor layout
- Test: `web/frontend/src/lib/archive.test.ts`

**Interfaces:**
- Produces `computeWindow(total, scrollTop, rowHeight, viewportHeight, overscan): { start,end }`.
- `ArchiveTable` renders `rows.slice(start, end)` inside a fixed `height` viewport, `contain: content`, with correct row count; details load on expand.
- Archive list calls `/api/audit/list` once per mount and opens details only through `/api/audit/report?audit_id=`.

- [ ] **Step 1: Write failing window calculation tests**

```ts
import { expect, it, describe } from 'vitest'
import { computeWindow } from './archive'

describe('archive window', () => {
  it('renders only visible rows plus overscan', () => {
    expect(computeWindow({ total: 179, scrollTop: 900, rowHeight: 40, viewportHeight: 400, overscan: 3 }))
      .toEqual({ start: 19, end: 36 })
  })
  it('never returns an invalid slice at the end', () => {
    expect(computeWindow({ total: 179, scrollTop: 10000, rowHeight: 40, viewportHeight: 400, overscan: 3 }))
      .toEqual({ start: 172, end: 179 })
  })
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/lib/archive.test.ts`
Expected: FAIL — file missing.

- [ ] **Step 3: Implement archive helper and accessible table**

```ts
// lib/archive.ts
export function computeWindow(input: {
  total: number; scrollTop: number; rowHeight: number; viewportHeight: number; overscan: number
}): { start: number; end: number } {
  const first = Math.floor(input.scrollTop / input.rowHeight)
  const count = Math.ceil(input.viewportHeight / input.rowHeight)
  const start = Math.max(0, first - input.overscan)
  const end = Math.min(input.total, first + count + input.overscan)
  return { start, end }
}
```

`ArchiveTable.tsx` behavior:

1. Full table container `role="table"` with native `<table>`; screen-reader full count uses `aria-rowcount={rows.length}`.
2. `.archive-window` has a fixed height (for 8 rows at 44 px); inner spacer `height: rows.length * rowHeight`.
3. `onScroll` updates `scrollTop`; no network fetch in the scroll handler.
4. `rows.slice(start,end)` are transformed into `<tr>` with `aria-rowindex={i+1}`.
5. Cells: date (`YYYY-MM-DD HH:mm`), site, link, duration, status (`Valide|Dégradé|Observation seule` from sealed values only), verdict (`Non noté` if absent), and actions.
6. Action `Ouvrir` is disabled/labelled `Preuve manquante` when `has_report !== 'true'`; `Rapport MD` is a bytes download button or a disabled cell with reason; `Copier la citation` opens the citation helper; never renders a working button without the required metadata.
7. Row expansion or focus loads proofs lazily and does not unmount/destroy the list.
8. Empty and error states show recovery, not silence.

- [ ] **Step 4: Run tests, including DOM row creation count**

```tsx
const rows = Array.from({ length: 179 }, (_, i) => ({ audit_id: `a-${i}`, site: `site-${i}` }))
render(<ArchiveTable rows={rows} onOpen={() => {}} />)
expect(container.querySelectorAll('tbody tr').length).toBeLessThan(30)
```

Run: `cd web/frontend && npx vitest run src/lib/archive.test.ts src/components/ArchiveTable.dom.test.tsx && npm run typecheck`
Expected: PASS; at most 30 rows exist in the DOM during initial render.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/archive.ts web/frontend/src/lib/archive.test.ts web/frontend/src/components/ArchiveTable.tsx web/frontend/src/components/ArchiveTable.dom.test.tsx web/frontend/src/components/AuditReportsPanel.tsx web/frontend/src/views/EvidenceView.tsx
git commit -m 'feat(web): virtualise l’archive de preuves'
```

---

### Task 9: Preuves wording, report drawer and proof honesty

**Files:**
- Modify: `web/frontend/src/components/AuditReportCard.tsx`
- Modify: `web/frontend/src/components/AuditReportCard.dom.test.tsx`
- Modify: `web/frontend/src/components/ProofDrawer.tsx` only for missing-proof handling
- Modify: `web/frontend/src/components/RemediationCard.tsx` only to keep remediation impossible without a valid complete outcome
- Modify: `web/frontend/src/lib/citation.ts` only if the archive citation action requires missing-data guards
- Test: `web/frontend/src/components/AuditReportCard.dom.test.tsx`

**Interfaces:**
- Produces three labels: `Valide`, `Dégradé`, `Observation seule`; `Preuve manquante` is an action state, never a verdict.
- Copy supports `Ouvrir`, `Rapport MD`, `Copier la citation`; missing SHA displays `Preuve manquante` and disables citation.

- [ ] **Step 1: Add missing-proof tests to DOM suite**

```tsx
it('never renders a green proof without a completed report payload', () => {
  const { container } = render(<AuditReportCard report={{ ...baseReport, remediation: undefined }} />)
  expect(container.textContent).not.toMatch(/PREUVE VÉRIFIABLE|prouvé|vérifiable/i)
})
```

- [ ] **Step 2: Run and confirm failure/extend existing tests**

Run: `cd web/frontend && npx vitest run src/components/AuditReportCard.dom.test.tsx`
Expected: FAIL on missing `Preuve manquante` path (or update expected copy to new exact strings).

- [ ] **Step 3: Replace binary proof badge with provenance verdict**

`AuditReportCard` header shows one of:

- `Valide` — report sealed and proof metadata present;
- `Dégradé` — sealed report with failed/degraded sections;
- `Observation seule` — no load/QoE result or report not remediation-safe;
- `Preuve manquante` — row may be focusable but cannot open proof/citation.

Hash rows only show present files; if telemetry is absent, render `Télémétrie absente` in normal Inter text, no invented SHA. “Copier la citation” returns an RFC-style sentence: site, link, date, measurement availability; if measurement is absent, it says exactly what is missing. Keep all scientific IDs in the proof drawer, never in primary table cells.

- [ ] **Step 4: Run tests**

Run: `cd web/frontend && npx vitest run src/components/AuditReportCard.dom.test.tsx src/lib/citation.test.ts src/lib/lexicon.test.ts && npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/AuditReportCard.tsx web/frontend/src/components/AuditReportCard.dom.test.tsx web/frontend/src/components/ProofDrawer.tsx web/frontend/src/components/RemediationCard.tsx web/frontend/src/lib/citation.ts
git commit -m 'feat(web): honore les preuves manquantes sans faux vert'
```

---

### Task 10: Keyboard, focus, names and 890 px content reflow

**Files:**
- Modify: `web/frontend/src/App.tsx`
- Modify: `web/frontend/src/components/CommandPalette.tsx` only if it leaks modal focus
- Modify: `web/frontend/src/lib/modalFocus.ts`
- Modify: `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/styles/components.css`
- Create: `web/frontend/e2e/operator-cockpit.spec.ts`
- Create: `web/frontend/playwright.config.ts`

**Interfaces:**
- Produces a guarded `isTextEntryTarget(target: EventTarget | null): boolean`.
- Playwright spec covers `890×768` and `1920×1080`, no horizontal scrolling/clipping, and 1–4 inert in text fields.

- [ ] **Step 1: Write failing focus/keyboard tests**

Extend `designFloor.test.ts` or a new `textEntryGuard.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { isTextEntryTarget } from './modalFocus'

describe('text entry guard', () => {
  it('ignores input, textarea, select, contenteditable and dialog targets', () => {
    const editable = { isContentEditable: true } as HTMLInputElement
    expect(isTextEntryTarget(editable)).toBe(true)
  })
})
```

Playwright scenario:

```ts
test('typing 2 in Site does not switch modes at 890px', async () => {
  await page.setViewportSize({ width: 890, height: 768 })
  await page.getByRole('button', { name: 'Auditer' }).click()
  await page.getByLabel('Site').fill('site 2')
  await expect(page).toHaveURL(/auditer|audit/)
  await expect(page.getByRole('textbox', { name: 'Site' })).toHaveValue('site 2')
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/lib/modalFocus.test.ts`
Expected: FAIL — helper is not exported or does not cover editable targets.

- [ ] **Step 3: Implement exact guard and modal focus behaviour**

```ts
export function isTextEntryTarget(target: EventTarget | null): boolean {
  if (!target) return false
  if (!(target instanceof HTMLElement)) return false
  if (target.closest('input, textarea, select, [contenteditable="true"], [role="dialog"]')) return true
  return target.isContentEditable
}
```

In `App.tsx`, global `keydown` exits early when `isTextEntryTarget(e.target)`. `1–4` still works away from inputs. Modals/drawers focus first control on open, trap Tab/Shift+Tab, close with Escape, restore focus to the invoker. At 890 px use one column, touch targets ≥44 px, headings clamp, and table actions wrap; no content relies on hover. Every interactive control is a native button/a/input with an accessible name.

- [ ] **Step 4: Run tests**

Run: `cd web/frontend && npx vitest run src/lib/modalFocus.test.ts src/lib/designFloor.test.ts && npx playwright test e2e/operator-cockpit.spec.ts`
Expected: Vitest PASS; Playwright PASS at both viewports.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/App.tsx web/frontend/src/components/CommandPalette.tsx web/frontend/src/lib/modalFocus.ts web/frontend/src/styles/index.css web/frontend/src/styles/components.css web/frontend/e2e/operator-cockpit.spec.ts web/frontend/playwright.config.ts
git commit -m 'feat(web): fiabilise clavier et 890 px'
```

---

### Task 11: Playwright behaviour matrix and accessibility floor

**Files:**
- Create: `web/frontend/e2e/accessibility.spec.ts`
- Create: `web/frontend/e2e/fixtures.ts`
- Modify: `web/frontend/playwright.config.ts`
- Modify: `web/frontend/package.json` only to add script `test:e2e` if missing
- Create: `web/frontend/src/lib/a11y.ts` (role/name/contrast assertions used by tests)
- Create: `web/frontend/src/lib/a11y.test.ts`

**Interfaces:**
- `test:e2e` launches the built app with deterministic API routes; no live service required.
- Tests mock `/api/integrity`, `/api/audit/list`, `/api/audit/report`, `/api/audit/live`, `/api/audit/status` with fixtures.

- [ ] **Step 1: Write failing negative e2e scenarios**

```ts
test('absent cannot render a green success', async () => {
  await page.route('**/api/audit/report', route => route.fulfill({ status: 404 }))
  await page.goto('/#/')
  await expect(page.locator('[data-testid="operator-verdict"]')).not.toContainText(/stable|sain|OK/i)
})
test('report open can be completed with the keyboard', async () => { /* open row with Enter */ })
test('failure state has a recovery action', async () => { /* route live fails, expect Réessayer button */ })
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx playwright test e2e/accessibility.spec.ts`
Expected: FAIL — routes/selectors not in the app yet.

- [ ] **Step 3: Implement scripts and fixtures**

`package.json` adds (no new dependency):

```json
"test:e2e": "playwright test",
"test:e2e:ui": "playwright test --ui"
```

`e2e/fixtures.ts` exposes raw JSON strings for a valid report, absent measurement, and corrupt payload. Browser tests assert:

- no synthetic zeros in absent fixture;
- no `PREUVE VÉRIFIABLE`/`green` pathway without required hashes;
- focus traverses TopNav → main controls → report controls in logical order;
- colour contrast rule: every `.state-*` computed color/background meets WCAG AA for 14 px text;
- table ARIA rows match windowed row count;
- full page has no horizontal overflow at 890 px.

`axe-core` is not added; if later approved, it can be swapped without changing the states or fixtures.

- [ ] **Step 4: Run browser matrix**

Run: `cd web/frontend && npx playwright test`
Expected: all e2e tests PASS at 890×768 and 1920×1080.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/e2e/accessibility.spec.ts web/frontend/e2e/fixtures.ts web/frontend/playwright.config.ts web/frontend/package.json web/frontend/src/lib/a11y.ts web/frontend/src/lib/a11y.test.ts
git commit -m 'test(web): verrouille accessibilité et états'
```

---

### Task 12: Documentation, no stale claims, and visual-evidence script

**Files:**
- Modify: `DESIGN.md`
- Modify: `PRODUCT.md`
- Modify: `docs/VERIFY.md` only after verifier pass (not in this task) 
- Modify: `docs/img/shots/2026-11-05/` (create captures if absent)
- Modify: `web/frontend/slice26-exhaustive.mjs` or create: `web/frontend/slice-cockpit-capture.mjs`
- Create: `web/frontend/scripts/validate-images.mjs`

**Interfaces:**
- `slice-cockpit-capture.mjs` takes `baseURL`, produces 12+ real PNGs at 890/1280/1920; verifies each has `.textContent` values from expected state and dimensions; no fabricated empty images.
- `validate-images.mjs` refuses images < 30% non-background for text states and detects duplicate hashes.

- [ ] **Step 1: Write failing image manifest check**

```js
import crypto from 'node:crypto'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
const dir = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..', 'docs', 'img', 'shots', '2026-11-05')
const hashes = new Set()
for (const file of readdirSync(dir).filter(f => f.endsWith('.png'))) {
  const bytes = crypto.createHash('sha256').update(readFileSync(join(dir, file))).digest('hex')
  if (hashes.has(bytes)) throw new Error(`duplicate capture ${file}`)
  hashes.add(bytes)
}
```

If current `.png` files are not yet deterministic, first assert non-empty and unique names only.

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && node scripts/validate-images.mjs`
Expected: FAIL — directory or validation absent.

- [ ] **Step 3: Update docs and capture script**

`DESIGN.md` changes:

- Replace fixed 1024–1920 minimum with 890–1920 (scaled 1366 @150%).
- Replace “avatar/edge labels” examples with current cockpit vocabulary.
- Replace `screen covers` and raw profile boxes with full-name tiles and card tokens.
- Replace mono-all prose rule with Inter vs JetBrains Mono division.
- Keep provenance/counters explicit; no generic “measured” claims.

`PRODUCT.md` changes:

- Dashboard: 3 status cards and one recommended action when measured.
- Auditer: fixed 60 second full measurement, no 20 s option.
- Banc: full profile cards.
- Preuves: human table + lazy proofs + 179-row handling.
- Remove the statement “no redesign in this wave”; state the cockpit wave explicitly if accepted.

Capture script names (no invented states):

```
cockpit-dashboard-{890,1280,1920}.png
cockpit-audit-{890,1280,1920}.png
cockpit-bench-{890,1280,1920}.png
cockpit-evidence-{890,1280,1920}.png
```

The app stays in fixture/observation mode. A loaded testimonial image is only valid when hosted fixtures exist; otherwise absent is the honest state.

- [ ] **Step 4: Run validation**

Run: `node web/frontend/slice-cockpit-capture.mjs http://127.0.0.1:9091 docs/img/shots/2026-11-05 && node web/frontend/scripts/validate-images.mjs`
Expected: 12 PNG + manifest; no duplicate hashes; zero live field traffic.

- [ ] **Step 5: Commit**

```bash
git add DESIGN.md PRODUCT.md web/frontend/slice-cockpit-capture.mjs web/frontend/scripts/validate-images.mjs docs/img/shots/2026-11-05
git commit -m 'docs(web): vivifie les captures du cockpit'
```

---

### Task 13: Whole-app verification and slice closeout

**Files:**
- Modify: `docs/CHECKPOINTS.md`
- Modify: `docs/VERIFY.md` only after independent verifier signs
- Modify: `docs/superpowers/plans/README.md`
- Test command outputs in `.perch/slice27/` (gitignored)

**Interfaces:**
- Builder does not assert `VERIFIED`; verifier owns it. This task records closed/blocked state exactly.

- [ ] **Step 1: Run local gates**

Run:

```bash
cd web/frontend && npm test && npm run build && cd ..
python tools/verify.py
python book/tools/audit-archive.py
go build ./...
go test ./... -count=1
```
On this host, if `make` is absent, use `python tools/verify.py` and record why.

Expected: tests green, 6/6 gates, archive checks OK, build output reports ~350 KB gz or a justified delta.

- [ ] **Step 2: Restore build artifact and inspect diff**

Run: `git checkout -- web/frontend/dist/index.html 2>/dev/null || true`
Run: `git status --short` and verify only intended source/docs/images are present; no secrets or freezed datasets.

- [ ] **Step 3: Independent verifier pass**

Use a separate subagent/verifier session. Prompt must:

- re-run Go/frontend suites yourself; do not accept builder output;
- inspect all four screens and all images;
- verify absent has no fake values;
- verify 890×768/1920×1080 screenshots;
- verify no network waterfall on 179 rows;
- verify no 20 s selector on Auditer;
- return explicit `RESIDUALS ACCEPTED: Y/N` and signature block or `withheld`.

- [ ] **Step 4: Ledger transcription and index**

Only after verifier sign:

- write exact commands, host, date, exit codes;
- move the slice row to the verifier’s disposition;
- record unresolved constraints (mobile out; no live field run);
- do not rewrite historical rows.

- [ ] **Step 5: Commit closeout**

```bash
git add docs/CHECKPOINTS.md docs/VERIFY.md docs/superpowers/plans/README.md
git commit -m 'docs: close le cockpit opérateur'
```

---

## Parallelization map

- Task 1 and Task 2 are sequential because both define state/type ownership.
- Task 3 opens CSS/layout work and can run while Task 1 is reviewed.
- Tasks 4, 5 and 7 depend on Task 3 and Task 1 but can be executed in either order after they land.
- Task 8 is independent of Dashboard/Audit UI and can run in parallel once archive fixtures compile.
- Task 9 depends on Task 8.
- Task 10 depends on all UI tasks.
- Task 11 depends on the finished implementation.
- Task 12 and Task 13 must remain at the end.

## Risk register

| Risk | Mitigation |
|---|---|
| UI quietly substitutes stale evidence for live state | `AuditViewState` discriminated union + tests that forbid mixed output |
| Dashboard fills absent state with zeros | `evaluateLoadedHealth()` returns `absent`; tests assert no synthetic units |
| 179 rows freeze low-end laptops | Windowed rendering + DOM count assertion + lazy detail loading |
| A VSAT link permanently fails CAP | Class-aware Δp95 threshold + fixture tests |
| Shortcuts break site names with digits | Typed `TextEntryGuard`; Playwright keyboard check |
| Style churn only fixes colors, not semantics | Strict tests, state contracts, documented design tokens |
| Mobile silently stays broken | out of scope but existing 700 px CSS may not regress; explicit verification at 890/1920 |
| Screenshots fabricate confidence | reproducible captures at three desktop sizes; verifier checks files and bytes |
| Model confidence creates green-only tests | tests assert forbidden outcomes; e2e uses mocked failure paths; no snapshot-only tests |

## Definition of done

- Four destinations render at 890×768 without horizontal scrolling or clipped action labels.
- 1920×1080 uses full width usefully; no narrow floating column with inert margins.
- No synthetic zero or green signal appears for absent measurement.
- Fixed 60-second audit flow is test-enforced.
- Archive performs no per-row waterfall and lazy-loads details.
- Keyboard shortcuts are inert in text fields.
- All required automated tests and gates pass.
- Independent verifier signs a written evidence block.
