# Cockpit UI Refinement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** refine every cockpit-visible component within the pinned bright institutional identity, with shared type/spacing declarations, clearer hierarchy and provable placement.

**Architecture:** four sequential tasks. Task 1 declares the shared cockpit type/rhythm tokens and the hierarchy oracle. Task 2 applies them to the decision spine and answer band. Task 3 applies them to the time field and journal rail. Task 4 runs the full proof and records the result. No new dependency, no layout-area reorder, no data-semantics change.

**Tech Stack:** React 19, TypeScript 5.9, Vite 6, Tailwind 4 tokens already in `src/styles/tokens.css`, hand-written CSS in `src/styles/components.css`, Vitest 3 with happy-dom for `*.dom.test.tsx`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-11-25-cockpit-ui-refinement-design.md`.
- Keep the pinned bright institutional identity: no glass, glow, gradient or decorative shadow.
- Keep named `grid-template-areas`, DOM order and keyboard/screen-reader order unchanged.
- Spacing uses the existing `--sp-1..5` scale; type uses the existing `--fs-*` scale. Add only alias declarations named below.
- Absence is never zero or green; no measurement, verdict or provenance value is fabricated.
- No dependency version, bundle budget, gate, permit, ledger, frozen corpus or protected config change.
- No live bench or host work in this plan — that is a separate scope.
- `npm -C web/frontend test -- --run <file>` is the DOM test command; `npm -C web/frontend run typecheck` and `sh .githooks/run-gates.sh` are the other checks. On a host without `make`, name the fallback command.
- Stage explicit paths only; never `git add -A`. One logical commit per task.

---

### Task 1: shared cockpit type/rhythm tokens and the hierarchy oracle

**Files:**
- Modify `web/frontend/src/styles/tokens.css` (after the `--sp-1..5` line, currently line 92)
- Modify `web/frontend/src/styles/components.css` (`.answer-word`, `.answer-sentence`, `.health-label`)
- Create `web/frontend/src/components/cockpit/CockpitHierarchy.dom.test.tsx`

**Interfaces:**
- Consumes: existing `--fs-hero`, `--fs-title`, `--fs-fine`, `--fs-body`, `--font-ui`, `--font-mono`, `--sp-2`, `--text-*` tokens.
- Produces: `--cockpit-font-hero`, `--cockpit-font-value`, `--cockpit-font-label`, `--cockpit-rhythm`; the DOM test `CockpitHierarchy.dom.test.tsx` used by Tasks 2-4.

- [ ] **Step 1: write the failing test** `CockpitHierarchy.dom.test.tsx`:

```tsx
// Oracle indépendant : la hiérarchie est lue dans le CSS réel et dans le DOM
// produit par le vrai DecisionSpine — aucun mock du composant testé.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { act } from 'react'
import { createRoot } from 'react-dom/client'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it } from 'vitest'
import DecisionSpine from './DecisionSpine'
import type { AuditReport } from '../../lib/types'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const REPORT = {
  audit_id: 'audit-t1', site: 'Siège', link_type: 'fiber',
  measured: { rtt_idle_p50_ms: 20, rtt_loaded_p95_ms: 78, small_measured: true,
    http_small_p95_ms: 64, load_measured: true },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: 'Sous charge, la latence reste sous objectif.' },
  cap_sla: { measured: true, ok_pct: 99.1, target_pct: 95, deadline_ms: 220 },
  bloat_grade: 'A', remediation_state: 'available', remediation_available: true,
} as unknown as AuditReport

let cleanup: (() => void) | null = null
function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  act(() => { root.render(node) })
  cleanup = () => { act(() => root.unmount()); host.remove() }
  return host
}
afterEach(() => { cleanup?.(); cleanup = null })

const css = (name: string) => readFileSync(join(process.cwd(), 'src/styles', name), 'utf8')

describe('cockpit — hiérarchie et rythme partagés', () => {
  it('déclare une seule échelle de type et un seul rythme', () => {
    const tokens = css('tokens.css')
    expect(tokens).toContain('--cockpit-font-hero: 600 var(--fs-hero)/1.15 var(--font-ui);')
    expect(tokens).toContain('--cockpit-font-value: 400 var(--fs-body)/1.5 var(--font-ui);')
    expect(tokens).toContain('--cockpit-font-label: 500 var(--fs-fine)/1.3 var(--font-mono);')
    expect(tokens).toContain('--cockpit-rhythm: var(--sp-2);')
  })
  it('le verdict précède les faits dans le DOM rendu', () => {
    const host = mount(<DecisionSpine report={REPORT} />)
    const word = host.querySelector('#verdict-word')
    const facts = host.querySelector('.answer-facts')
    expect(word).not.toBeNull()
    expect(facts).not.toBeNull()
    expect(word!.compareDocumentPosition(facts!) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(host.querySelectorAll('#verdict-word')).toHaveLength(1)
  })
  it('les déclarations de type utilisent les alias, pas des valeurs en double', () => {
    const components = css('components.css')
    expect(components).toContain('.answer-word { font: var(--cockpit-font-hero);')
    expect(components).toContain('.answer-sentence { font: var(--cockpit-font-value);')
    expect(components).toContain('.health-label { font: var(--cockpit-font-label);')
  })
})
```

- [ ] **Step 2: verify RED**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitHierarchy.dom.test.tsx`
Expected: FAIL — the four tokens and three alias declarations do not exist yet.

- [ ] **Step 3: implement the shared declarations**

In `tokens.css`, immediately after `--sp-1: 4px; --sp-2: 8px; --sp-3: 12px; --sp-4: 16px; --sp-5: 24px;`, add exactly:

```css
  --cockpit-font-hero: 600 var(--fs-hero)/1.15 var(--font-ui);
  --cockpit-font-value: 400 var(--fs-body)/1.5 var(--font-ui);
  --cockpit-font-label: 500 var(--fs-fine)/1.3 var(--font-mono);
  --cockpit-rhythm: var(--sp-2);
```

In `components.css`, replace exactly:
- `.answer-word { display: flex; align-items: center; gap: var(--sp-3); font: 600 var(--fs-hero)/1.15 var(--font-ui); color: var(--text-primary); margin: 0; text-wrap: balance; overflow-wrap: anywhere; }` → same declaration with `font: var(--cockpit-font-hero);`
- `.answer-sentence { font: 400 var(--fs-body)/1.5 var(--font-ui); color: var(--text-body); max-width: 68ch; margin: 0; text-wrap: pretty; }` → same with `font: var(--cockpit-font-value); color: var(--text-body);`
- `.health-label { font: 500 var(--fs-fine) var(--font-mono); letter-spacing: var(--tracking-caps); text-transform: uppercase; color: var(--text-muted); }` → same with `font: var(--cockpit-font-label);`

Keep all other properties, media-query overrides and French comments unchanged.

- [ ] **Step 4: verify GREEN**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitHierarchy.dom.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 5: commit**

```bash
git add web/frontend/src/styles/tokens.css web/frontend/src/styles/components.css web/frontend/src/components/cockpit/CockpitHierarchy.dom.test.tsx
git commit -m "refactor(ui): shared cockpit type and rhythm tokens"
```

---

### Task 2: decision spine and answer band placement

**Files:**
- Modify `web/frontend/src/styles/components.css` (`.spine-view`, `.decision-spine`, `.answer-band`, `.answer-facts .state-slot`, `.act-slot`, `.proof-slot`)
- Modify `web/frontend/src/components/cockpit/CockpitHierarchy.dom.test.tsx` (add two assertions)

**Interfaces:**
- Consumes: `--cockpit-rhythm` from Task 1; existing `#verdict-word`, `.answer-facts`, `.act-slot`, `.proof-slot` DOM.
- Produces: one aligned rhythm for the decision column, asserted in the shared DOM test.

- [ ] **Step 1: extend the failing test** with this case inside the same `describe`:

```tsx
  it('la colonne de décision utilise le rythme partagé et un seul filet', () => {
    const components = css('components.css')
    expect(components).toContain('.decision-spine { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); }')
    expect(components).toContain('.answer-facts .state-slot { min-height: 72px; border-top: 1px solid var(--hairline-faint); padding-top: var(--cockpit-rhythm); }')
    expect(components).toContain('.act-slot { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); border-top: 1px solid var(--hairline); padding-top: var(--sp-3); }')
  })
```

- [ ] **Step 2: verify RED**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitHierarchy.dom.test.tsx`
Expected: FAIL on the three new declarations.

- [ ] **Step 3: implement the placement pass** — in `components.css`, replace the existing declarations with exactly:

```css
.spine-view { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); min-width: 0; }
.decision-spine { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); }
.answer-band { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); padding: var(--sp-4) 0; border-bottom: 1px solid var(--hairline-faint); }
.answer-facts .state-slot { min-height: 72px; border-top: 1px solid var(--hairline-faint); padding-top: var(--cockpit-rhythm); }
.act-slot { display: flex; flex-direction: column; gap: var(--cockpit-rhythm); border-top: 1px solid var(--hairline); padding-top: var(--sp-3); }
.proof-slot { border-top: 1px solid var(--hairline-faint); padding-top: var(--cockpit-rhythm); }
```

Do not touch any `@media` block, any `data-*` selector or any component TSX.

- [ ] **Step 4: verify GREEN and DOM neighbours**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitHierarchy.dom.test.tsx src/components/cockpit/AnswerBand.compact.dom.test.tsx src/components/cockpit/DecisionSpine.dom.test.tsx`
Expected: all PASS; the existing single-verdict and density assertions stay intact.

- [ ] **Step 5: commit**

```bash
git add web/frontend/src/styles/components.css web/frontend/src/components/cockpit/CockpitHierarchy.dom.test.tsx
git commit -m "refactor(ui): align cockpit decision column"
```

---

### Task 3: time field and journal rail placement

**Files:**
- Modify `web/frontend/src/styles/components.css` (`.cockpit-field`, `.time-field`, `.tf-header`, `.journal-rail`, `.jr-row`)
- Create `web/frontend/src/components/cockpit/CockpitField.dom.test.tsx`

**Interfaces:**
- Consumes: `--cockpit-rhythm` from Task 1; `TypedMetricRing`, `DomainEventRing`, `InstrumentRings` from `src/lib/ringBuffer`; `TimeField` and `JournalRail` components.
- Produces: a shared alignment rhythm for the instrument field, proven by a new DOM test.

- [ ] **Step 1: write the failing test** `CockpitField.dom.test.tsx`:

```tsx
// Oracle indépendant : on monte les vrais composants avec des anneaux vides,
// puis on lit le CSS réel. Aucun chiffre n'est fabriqué pour une absence.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField from './TimeField'
import JournalRail from './JournalRail'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

function emptyRings(): InstrumentRings {
  const mk = () => new TypedMetricRing(1)
  return { small: mk(), rtt50: mk(), rtt95: mk(), goodput: mk(), ifaceRx: mk(), ifaceTx: mk() }
}

describe('cockpit field — placement et absence honnête', () => {
  it('le champ et le journal partagent le rythme déclaré', () => {
    const css = readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')
    expect(css).toContain('.cockpit-field { min-width: 0; display: grid; grid-template-columns: minmax(0, 1fr); gap: var(--cockpit-rhythm); }')
    expect(css).toContain('.time-field { min-width: 0; display: flex; flex-direction: column; gap: var(--cockpit-rhythm); }')
    expect(css).toContain('.journal-rail { min-width: 0; display: flex; flex-direction: column; gap: var(--cockpit-rhythm); max-height: 100%; }')
  })
  it('sans événement, le journal dit l’absence sans chiffre ni vert', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<JournalRail events={new DomainEventRing(4)} /> as never)
    const empty = host.querySelector<HTMLElement>('.c-empty')
    expect(empty).not.toBeNull()
    expect(empty!.textContent).toContain('Aucun événement')
    expect(empty!.textContent).not.toMatch(/\d/)
    expect(empty!.className).not.toMatch(/(?:^|\s)(?:ok|success|green|pass)(?:\s|$)/i)
  })
})
```

- [ ] **Step 2: verify RED**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitField.dom.test.tsx`
Expected: FAIL — the three CSS declarations still use `--sp-2`/`--sp-3`.

- [ ] **Step 3: implement the placement pass** — replace those three declarations in `components.css` with exactly the strings asserted above. Change no `@media` block, no data attribute and no component logic.

- [ ] **Step 4: verify GREEN and neighbours**

Run: `npm -C web/frontend test -- --run src/components/cockpit/CockpitField.dom.test.tsx src/components/cockpit/TimeField.dom.test.tsx src/components/HonestStates.dom.test.tsx src/components/cockpit/DecisionSpine.dom.test.tsx`
Expected: all PASS.

- [ ] **Step 5: commit**

```bash
git add web/frontend/src/styles/components.css web/frontend/src/components/cockpit/CockpitField.dom.test.tsx
git commit -m "refactor(ui): align cockpit instrument field"
```

---

### Task 4: full proof and recorded result

**Files:**
- Create `docs/verification/2026-11-25-cockpit-refinement/README.md`

**Interfaces:**
- Consumes: the three task commits, the six native gates, existing DOM tests and axe e2e.
- Produces: a verification record with exact commands, exit codes and what stayed unproven.

- [ ] **Step 1: run the proof** — in order, recording each exact command and exit code:

```bash
npm -C web/frontend run typecheck
npm -C web/frontend test -- --run
sh .githooks/run-gates.sh
```

Expected: typecheck exit 0; full DOM/pure suite exit 0; gates `6 passed, 0 failed, 0 not run`, exit 0. A non-zero exit stops the task and is reported, never suppressed.

- [ ] **Step 2: write the record** `docs/verification/2026-11-25-cockpit-refinement/README.md` with: the three commit hashes, the three commands and exits, the number of new DOM assertions, the untouched identity/layout constraints, and an explicit list of what stayed unproven (e.g. browser-only e2e if not run). No invented PASS, no fabricated screenshot.

- [ ] **Step 3: commit**

```bash
git add docs/verification/2026-11-25-cockpit-refinement/README.md
git commit -m "docs(verify): cockpit refinement proof"
```

## Self-Review

- Spec §2 component inventory -> Tasks 1-3 cover the cockpit CSS contract, decision column and instrument field; surrounding audit/report/proof components keep their existing honest-state wording and are covered by the existing `HonestStates` oracle.
- Spec §3 visual contract -> Task 1 declares type/rhythm aliases; Tasks 2-3 apply them; no new colour, no shadow, no gradient.
- Spec §4 state contract -> Task 3's absence oracle plus the existing `HonestStates.dom.test.tsx`; no verdict chip for absent/error data.
- Spec §5 proof -> Task 4 runs typecheck, the full suite and the six gates once, with an honest record.
- Placeholder scan: none; every step names exact files and contains the code or command to run.
- Type consistency: `--cockpit-font-hero`, `--cockpit-font-value`, `--cockpit-font-label`, `--cockpit-rhythm` are declared once in Task 1 and used verbatim in Tasks 2-3.
