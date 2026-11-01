# Operator console (slice 20B) implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the four destinations as a layman operator console: one sans, surfaces instead of borders, one primary action per screen, a Go-computed verdict with four facts, a platform-true remediation copy, and a two-tab proof drawer.

**Architecture:** The engine (slice 20A) supplies every number and the verdict; the frontend renders, selects a platform, and never recomputes a measurement. New pieces: `VerdictHero`, `RemediationCard`, `ProofDrawer`, `LiveLatencyCanvas`, `DuelCard`, plus a token and shell rewrite. `data-mode` navigation attributes and the root SSE session survive unchanged.

**Tech Stack:** React 19, TypeScript, Vite 6, Zustand 5, vitest 3 + jsdom, one `<canvas>` for the live trace, no chart library on operator surfaces.

**Spec:** `docs/superpowers/specs/2026-10-31-cgo-operator-console-design.md` §3.1, §3.3–§3.10, §4–§7, §8 (frontend), §9.
**Depends on:** slice 20A merged — `operator_verdict`, `load_attempted`, `up_shaped_mbps`, `routeros_lan_assumed`, `economics.*.numbers`, `timeline`, `/api/duel`.
**Dependency status, decided 2026-11-01:** 20A is *not* merged and cannot be — a merge to `main` is the human's decision, and its seven `VERIFY` rows stay `PENDING` for an independent verifier. All six engine fields above are committed on this branch (`5d58116`, `f478ef6`, `49f76d0`, `eb11ca9`, `2183492`, `23ffdb1`, `df89c82`), and 20C already built on that tree the same way. 20B therefore proceeds here, with two consequences the operator must know: every frontend assertion about an engine field is written against the committed field name, and the `VERIFY` rows this slice adds stay `PENDING` too — 20B cannot promote a row 20A has not earned. If a verifier later renames a 20A field, 20B's fixtures break loudly rather than silently.

## Global Constraints

- Inter is the only UI face. JetBrains Mono only for numbers, units, IPs and commands. Cormorant leaves the app bundle; it stays in `pkg/figures` and the book.
- No uppercase French sentence anywhere on a surface. Sentence case for labels, buttons and titles.
- Scale, exact: hero verdict 40px, metric value 32px, page title 28px, body 16px, label 14px, fine print 13px inside the drawer only. Buttons 16px on 48px targets.
- Surfaces, exact: canvas `#0a0b0e`, card `#12141a`, elevated `#181b22`. Borders only on table rows, dividers and the focused element. Radius 0. No shadows, gradients or glass on the four destinations.
- Colour: green/amber/red carry verdict and status with a word plus a dot; cyan marks live and the active destination; violet stays in research charts. The tricolor stays identity-only.
- Banned on any default surface: `p50`, `p95`, `qdisc`, `CAKE`, `fq_codel`, `pfifo_fast`, `BBR`, `CUBIC`, `JFI`, `QDI`, `SLA`, `bufferbloat`, `eBPF`, `manifeste`, `corpus`, `gel`, `twin`, `run`, hex strings, `n=`, `garanti`, `garantie`, `prouvé`. They belong in the drawer, except the bench research mode where protocol names are the subject.
- No number is computed in the browser except the arithmetic of two already-displayed values, and those come from `/api/duel`.
- One primary action per screen. Secondary actions are text links or outline buttons.
- Absent data is a word: “Pas encore mesuré”, “Non chiffré”, “non mesurée”. Never a bare dash, never a zero.
- Keep `button[data-mode]` on the destinations: `qa-walk.mjs` and `capture-fresh.mjs` click them.
- No live bench run. Tests use fixtures and `renderToStaticMarkup`; the canvas is tested through a pure function.
- Never `git add -A`; never commit `book/` changes. Every task ends with `npm run typecheck && npm test` green, then a commit.

---

### Task 1: Tokens, faces, surfaces

**Files:**
- Modify: `web/frontend/src/styles/tokens.css`
- Modify: `web/frontend/src/lib/palette.ts` (`FALLBACK` must equal the sheet)
- Modify: every rule using `--font-display` (find with `grep -rn "font-display" web/frontend/src`)
- Modify: `web/frontend/src/lib/tokens.test.ts`
- Delete: `web/frontend/public/fonts/cormorant-garamond-500.woff2`, `cormorant-garamond-600.woff2` if `git ls-files` tracks them

**Interfaces:**
- Produces: tokens `--fs-hero`, `--fs-metric`, `--fs-title`, `--fs-body`, `--fs-label`, `--fs-fine`, `--font-ui`; `--font-display` removed. `paletteOnce().cc.bbr3` keeps working.

- [ ] **Step 1: Write the failing test**

Append to `web/frontend/src/lib/tokens.test.ts`:

```ts
describe('échelle opérateur', () => {
  const sheet = read('web/frontend/src/styles/tokens.css')
  it('porte l’échelle-body et bannit la serif d’interface', () => {
    for (const [tok, px] of [['--fs-hero', '40px'], ['--fs-metric', '32px'], ['--fs-title', '28px'], ['--fs-body', '16px'], ['--fs-label', '14px'], ['--fs-fine', '13px']]) {
      expect(sheet, tok).toContain(`${tok}: ${px}`)
    }
    expect(sheet).not.toContain('Cormorant')
    expect(sheet).not.toContain('--font-display')
  })
  it('les surfaces de console sont figées', () => {
    expect(sheet).toContain('--canvas-solid: #0a0b0e')
    expect(sheet).toContain('--surface-card: #12141a')
    expect(sheet).toContain('--surface-elevated: #181b22')
  })
})
```

The helper is `read(path)`, already defined at the top of that file — it takes a
repo-relative path (`read('web/frontend/src/styles/tokens.css')`), not a bare
filename. There is no `readCss`; the earlier draft of this step invented it.

- [ ] **Step 1b: Reconcile the assertions this task invalidates**

The same file already pins the old skin. Read the whole file before changing
anything, then update each of these in place — do not delete an assertion to make
the suite green, and do not leave a rule asserting a gradient this task removes:

- the `radial-gradient` assertion on `tokens.css`: Step 3 flattens `--canvas`, so
  this passes only if another rule still legitimately uses a radial tint. Check
  with `grep -n "radial-gradient" src/styles/tokens.css`. If none remains, replace
  the assertion with the flat-canvas contract (`--canvas` contains no `gradient`)
  and say which you chose in the report;
- the `linear-gradient` and `box-shadow: inset` and `.card:hover` assertions:
  these read `index.css`/`components.css`, which Task 2 owns. Leave them green —
  Task 1 must not break them, so verify with `npx vitest run src/lib/tokens.test.ts`
  after Step 3, not only after Step 6;
- the two `.mode-rail` media assertions: they survive Task 1 and break in Task 2,
  which handles them. Do not touch them here.

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/lib/tokens.test.ts`
Expected: FAIL on `--fs-hero` and on the Cormorant assertion.

- [ ] **Step 3: Rewrite the token block**

In `web/frontend/src/styles/tokens.css`:
- delete the two `@font-face` rules for `'Cormorant Garamond'` (lines 8-9);
- delete `--font-display` (line 65, not 64 — measured) and add `--font-ui` as an alias, per the bullet at the end of this step; do not re-declare the Inter stack a second time.
- replace the five `--fs-*` declarations with the six listed in Step 1, keeping
  all four retired names as aliases so no rule loses its `font-size`. The earlier
  draft aliased only `--fs-micro` and `--fs-tag` and dropped `--fs-cap` and
  `--fs-head` — measured against the tree, those two are still live at 32 and 15
  `var()` sites (`index.css`, `components.css`, `CompareView.tsx`, `Explain.tsx`,
  `ResultatsView.tsx`), so deleting them would leave 47 rules with an unresolvable
  size. `--fs-base` has zero uses and can go:

```css
  --fs-fine: 13px;
  --fs-label: 14px;
  --fs-body: 16px;
  --fs-title: 28px;
  --fs-metric: 32px;
  --fs-hero: 40px;
  --fs-micro: var(--fs-fine); /* alias de transition, supprimé en fin de slice */
  --fs-tag: var(--fs-label);
  --fs-cap: var(--fs-fine);   /* 13px : même rendu qu'avant, aucun changement visuel */
  --fs-head: var(--fs-title); /* 22px → 28px : la promotion du titre est voulue */
```

  After Step 3, run `npx vitest run src/lib/tokens.test.ts` and `npm run
  typecheck` before touching any other file — a broken alias shows up here, not at
  Step 6.
- `--font-ui` duplicates `--font-body`: both are Inter with the same fallback
  stack, and `--font-body` is the one already referenced across the sheets. Make
  `--font-body` the canonical name, add `--font-ui` as its alias (not the other
  way round), so this task renames nothing it does not have to:
  `--font-ui: var(--font-body);`

- set `--canvas-solid: #0a0b0e;`, `--surface-soft: #0e1015;`, `--surface-card: #12141a;`, `--surface-elevated: #181b22;` and keep `--canvas` as the same flat colour without the radial tints;
- keep `--hairline` and friends; borders stop being the main separator in Task 2, not here.

- [ ] **Step 4: Keep the palette contract**

In `web/frontend/src/lib/palette.ts`, update `FALLBACK` values for the four surfaces to the new hexes. `--chart-bbr3` stays.

- [ ] **Step 5: Retire the serif from the app**

Run: `grep -rn "font-display" web/frontend/src`
Replace each hit's `var(--font-display)` with `var(--font-ui)`, and where the rule also set `text-transform: uppercase` on French text, drop the transform. Then:

Measured against the tree: the `--font-display` call sites are 14 (`index.css` ×11,
`components.css` ×1 in `.viz-empty`, `MeteolinkWordmark.tsx` ×1, plus the token
declaration itself). Three of them also carry `text-transform: uppercase` on
French text (`.view-title`, `.mode-head h1`, `.syn-hero h1`) — drop those three
transforms here, because a serif-free title in caps is still an uppercase French
sentence.

The wordmark is the one hit that needs a decision, not a substitution:
`MeteolinkWordmark.tsx:38` reads `fontFamily: 'var(--font-display, "Cormorant Garamond")'`.
Replacing only the token leaves the literal `"Cormorant Garamond"` fallback, which
still renders a serif on any machine that has one installed locally — the app would
claim one face and ship another. Delete the fallback string too, so the wordmark
resolves to `var(--font-ui)`. Finish with `grep -rn "Cormorant" web/frontend/src`
empty; if it is not, the serif is still in the app.

```bash
git rm web/frontend/public/fonts/cormorant-garamond-500.woff2 web/frontend/public/fonts/cormorant-garamond-600.woff2
```

Only if `git ls-files` lists them; otherwise delete the untracked copies and say so.

- [ ] **Step 6: Verify, measure, commit**

Run: `npm run typecheck && npx vitest run src/lib/tokens.test.ts src/lib/palette.test.ts && npm run build`
Expected: tests pass; the bundle check prints a `TOTAL gz` lower than the 351.0 KB baseline — record the new number, do not estimate it.

```bash
git add web/frontend/src/styles/tokens.css web/frontend/src/lib/palette.ts web/frontend/src/lib/tokens.test.ts web/frontend/src web/frontend/dist/index.html
git commit -m "ui: une seule police d'interface, une échelle de lecture, des surfaces"
```

---

### Task 2: Shell — one header, no duplicated navigation

**Files:**
- Create: `web/frontend/src/components/TopNav.tsx`
- Delete: `web/frontend/src/components/ModeRail.tsx`
- Modify: `web/frontend/src/App.tsx`, `web/frontend/src/store/ui.ts`, `web/frontend/src/styles/index.css`
- Modify: `web/frontend/src/lib/benchLayout.test.ts` (it reads `ModeRail.tsx`), `web/frontend/src/components/StatusRibbon.tsx`

**Interfaces:**
- Consumes: `MODES`, `setMode`, `preloadMode`.
- Produces: `<nav class="topnav">` with `button[data-mode]` per destination, `aria-current`, key hints `1–4`; `.topnav-status` cluster; footer keeps only `.foot-ticker`.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/components/TopNav.dom.test.tsx`:

```tsx
import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import TopNav from './TopNav'

describe('TopNav', () => {
  const html = renderToStaticMarkup(<TopNav />)
  it('expose les quatre destinations, en phrase case, avec leur touche', () => {
    for (const m of ['dashboard', 'audit', 'bench', 'evidence']) {
      expect(html).toContain(`data-mode="${m}"`)
    }
    expect(html).toContain('Tableau de bord')
    expect(html).toContain('Preuves & Rapports')
  })
  it('ne duplique pas la navigation', () => {
    expect(html).not.toContain('Auditer un lien')
    expect(html).not.toContain('mode-pin')
  })
})
```

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/components/TopNav.dom.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement**

Create `web/frontend/src/components/TopNav.tsx`:

```tsx
import { MODES, useUIStore } from '../store/ui'
import { preloadMode } from '../lib/preload'
import { NavIcon } from './NavIcon'

// Une ligne, quatre destinations, aucune barre de raccourcis en pied : le
// tableau de bord comptait quatre chemins vers le même audit.
export default function TopNav() {
  const mode = useUIStore(s => s.mode)
  const setMode = useUIStore(s => s.setMode)
  return (
    <nav className="topnav" aria-label="Navigation par destination">
      {MODES.map(m => (
        <button
          key={m.id}
          type="button"
          className={'topnav-item' + (mode === m.id ? ' on' : '')}
          data-mode={m.id}
          aria-pressed={mode === m.id}
          aria-current={mode === m.id ? 'page' : undefined}
          title={`${m.label} — touche ${m.key}`}
          onClick={() => setMode(m.id)}
          onMouseEnter={() => preloadMode(m.id)}
          onFocus={() => preloadMode(m.id)}
        >
          <NavIcon id={m.id} size={18} />
          <span className="topnav-lbl">{m.label}</span>
          <span className="topnav-key" aria-hidden="true">{m.key}</span>
        </button>
      ))}
    </nav>
  )
}
```

- [ ] **Step 4: Rewire the shell**

In `web/frontend/src/App.tsx`: replace `<ModeRail />` with `<TopNav />` inside the header row, drop the `ModeRail` import, and keep `<footer className="foot-ticker">` unchanged. Delete `web/frontend/src/components/ModeRail.tsx`. In `web/frontend/src/store/ui.ts`, remove `railPinned`, `setRailPinned` and the `railPinned()` localStorage reader, and the `rail-min` class use in `App.tsx`.

In `web/frontend/src/styles/index.css`:
- add `.topnav { display:flex; gap:var(--sp-1); align-items:center; }`, `.topnav-item { min-height:48px; padding:0 var(--sp-3); font-size:var(--fs-body); color:var(--text-body); background:transparent; border:0; }`, `.topnav-item.on { color:var(--text-primary); box-shadow: inset 0 -2px 0 var(--t-live); }`, `.topnav-key { font-family:var(--font-mono); font-size:var(--fs-fine); color:var(--text-faint); margin-left:var(--sp-1); }`;
- under `@media (max-width: 700px)`, make `.topnav` a fixed bottom bar: `position:fixed; bottom:0; left:0; right:0; background:var(--surface-soft); border-top:1px solid var(--hairline-faint); justify-content:space-around; z-index:40;` and hide `.topnav-key`;
- delete the `.mode-rail`, `.mode-btn`, `.mode-pin`, `.syn-actions` and `.syn-action` rules and their media overrides;
- give `#shell` a grid without the sidebar column and reserve `padding-bottom:72px` under 700px so the bottom bar never covers content.

- [ ] **Step 5: Fix the source-reading tests**

In `web/frontend/src/lib/benchLayout.test.ts`, replace the `ModeRail.tsx` reads with `TopNav.tsx` and the assertion `rail.toContain('preloadMode(m.id)')` with the same call in the new file.

`src/lib/tokens.test.ts` has the same problem and the earlier draft missed it: two
of its media-query cases assert that `.mode-rail` rules exist — one that
`.mode-rail { … flex-direction: row` appears under `max-width: 899px`, one that
`.mode-rail { … position: fixed` appears under `max-width: 700px`. Step 4 deletes
both rules, so rewrite those two cases against `.topnav` and the new breakpoint,
keeping them as media-query existence assertions rather than dropping them.

Then run `npx vitest run` and fix only assertions that named the deleted rail.
Before committing, check the Playwright walk is not silently broken: `qa-walk.mjs`
clicks by text (`button:has-text("RÉSULTATS")`, `.bench-anchors`,
`.evidence-anchors`), not by `data-mode`, and Playwright's `:has-text()` is
case-insensitive substring, so the sentence-case labels survive — but the
`.bench-anchors` / `.evidence-anchors` clusters must still exist after this task.
If Step 4 removes either, that is a Task 7 decision, not a Task 2 one; leave them
and say so in the report.

- [ ] **Step 6: Verify and commit**

Run: `npm run typecheck && npx vitest run && npm run build`
Expected: all green.

```bash
git add web/frontend/src/App.tsx web/frontend/src/components/TopNav.tsx web/frontend/src/components/TopNav.dom.test.tsx web/frontend/src/store/ui.ts web/frontend/src/styles/index.css web/frontend/src/lib/benchLayout.test.ts web/frontend/dist/index.html
git rm web/frontend/src/components/ModeRail.tsx
git commit -m "ui: une barre de destinations remplace le rail et la barre de raccourcis"
```

---

### Task 3: Verdict hero and four facts

**Files:**
- Create: `web/frontend/src/components/VerdictHero.tsx`
- Modify: `web/frontend/src/lib/types.ts`, `web/frontend/src/lib/dashboard.ts`, `web/frontend/src/lib/dashboard.test.ts`
- Create: `web/frontend/src/components/VerdictHero.dom.test.tsx`
- Modify: `web/frontend/src/components/OperatorStats.tsx` (keeps the audit view's dense variant)

**Interfaces:**
- Consumes: `report.operator_verdict.{level,word,sentence}`, `report.measured.{load_attempted,load_measured,http_small_p95_ms}`, `report.cap_sla`, `report.bloat_grade`, `report.economics.primary.{ar_per_hour,name,model,numbers}`.
- Produces: `<VerdictHero report tone …/>`; `operatorFacts(report, warnMs, critMs): Fact[]` where `Fact = {key, label, value, unit, target, band}`.

- [ ] **Step 1: Write the failing tests**

Create `web/frontend/src/components/VerdictHero.dom.test.tsx`:

```tsx
import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { VerdictHero } from './VerdictHero'
import type { AuditReport } from '../lib/types'

const base: AuditReport = {
  audit_id: 'a1', site: 'Département X', link_type: '4g',
  measured: { load_attempted: true, load_measured: true, http_small_p95_ms: 209, small_measured: true },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: "Sous charge, la latence reste sous l'objectif." },
  cap_sla: { measured: true, ok_pct: 96.7, breach: false },
  bloat_grade: 'A',
}

describe('VerdictHero', () => {
  it('affiche le mot du moteur, sa phrase et les quatre faits', () => {
    const out = renderToStaticMarkup(<VerdictHero report={base} />)
    expect(out).toContain('Lien sain')
    expect(out).toContain('data-level="green"')
    expect(out).toContain('Alertes à l’heure')
    expect(out).toContain('96,7')
    expect(out).toContain('Capacité gaspillée')
    expect(out).toContain('Non chiffré')
  })

  it('ne transforme jamais une absence en chiffre', () => {
    const out = renderToStaticMarkup(<VerdictHero report={{
      ...base,
      measured: { load_attempted: true, load_measured: false },
      operator_verdict: { level: 'neutral', word: 'Charge non établie', sentence: 'Le flux saturant n’a pas pu démarrer.' },
      cap_sla: { measured: false, ok_pct: 0, breach: false },
      bloat_grade: '',
    }} />)
    expect(out).toContain('Charge non établie')
    expect(out).toContain('Pas encore mesuré')
    expect(out).not.toContain('—')
  })

  it('n’invente pas de verdict si le moteur n’en a pas produit', () => {
    const out = renderToStaticMarkup(<VerdictHero report={{ ...base, operator_verdict: undefined } as AuditReport} />)
    expect(out).toContain('Pas encore mesuré')
  })
})
```

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/components/VerdictHero.dom.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: One decimal separator for the whole console**

French prose and French numbers use the comma, and the console shows both on one
line. Add to `web/frontend/src/lib/format.ts`:

```ts
// numFr — une seule virgule décimale pour la console. Les nombres technique
// du tiroir gardent leur format d'origine, celui du moteur.
export const numFr = (v: number | null | undefined, digits = 1): string =>
  v == null || !Number.isFinite(v) ? '' : v.toFixed(digits).replace('.', ',')
```

Append to `web/frontend/src/lib/format.test.ts`:

```ts
import { numFr } from './format'

describe('numFr', () => {
  it('écrit les décimales à la française et ne fabrique rien', () => {
    expect(numFr(96.73, 1)).toBe('96,7')
    expect(numFr(209.4, 0)).toBe('209')
    expect(numFr(null)).toBe('')
    expect(numFr(Number.NaN)).toBe('')
  })
})
```

- [ ] **Step 4: Extend the types and the fact builder**

In `web/frontend/src/lib/types.ts`, add to `AuditReport`:

```ts
  operator_verdict?: { level: 'neutral' | 'red' | 'amber' | 'green'; word: string; sentence: string }
```

and to `AuditMeasured`: `load_attempted?: boolean`. Add to `AuditEconomicsScenario`:
`numbers?: { model: string; monthly_opex_mga?: number; cir_mbps?: number; ar_per_gb?: number; retrans_ratio?: number; wasted_bytes?: number; transport_bytes?: number; load_seconds?: number }`. The Go side emits a superset (slice 20A Task 4 review); the UI declares the keys it reads and ignores the rest, and `model` is the key the fact's wording branches on.

In `web/frontend/src/lib/dashboard.ts`, append:

```ts
export interface Fact { key: string; label: string; value: string; unit: string; target: string; band: Band | null }

// Les quatre faits de la console. Un seul calcul, partagé par le tableau de
// bord et l'audit : le verdict et les faits lisent les mêmes champs.
export function operatorFacts(
  report: AuditReport | null | undefined,
  warnMs: number,
  critMs: number,
): Fact[] {
  const s = operatorSummary(report, warnMs, critMs)
  const cap = report?.cap_sla
  const cost = report?.economics?.primary
  return [
    {
      key: 'alerts', label: 'Alertes à l’heure',
      value: s.alertsPct != null ? numFr(s.alertsPct, 1) : 'Pas encore mesuré',
      unit: s.alertsPct != null ? '%' : '', target: 'objectif 95 %', band: s.alerts,
    },
    {
      key: 'delay', label: 'Réponse sous charge',
      value: s.delayMs != null ? numFr(s.delayMs, 0) : 'Pas encore mesuré',
      unit: s.delayMs != null ? 'ms' : '', target: 'objectif 220 ms', band: s.delay,
    },
    {
      key: 'quality', label: 'Qualité du lien',
      value: s.grade || 'Pas encore mesuré', unit: '',
      target: s.grade ? `note ${s.grade}` : 'aucune note', band: s.grade ? { word: s.quality, level: qualityLevel(s.grade) } : null,
    },
    {
      key: 'cost', label: 'Capacité gaspillée',
      value: cost?.ar_per_hour != null ? formatCostAR(cost.ar_per_hour).num : 'Non chiffré',
      unit: cost?.ar_per_hour != null ? formatCostAR(cost.ar_per_hour).unit : '',
      target: cost?.name ? `scénario ${scenarioWord(cost.model, cost.name)}` : 'aucun scénario retenu',
      band: cost?.ar_per_hour != null ? { word: 'estimation contractuelle', level: 'warn' as PlainLevel } : null,
    },
  ]
}

function qualityLevel(grade: string): PlainLevel {
  const g = grade.charAt(0).toUpperCase()
  if (g === 'A' || g === 'B') return 'ok'
  if (g === 'C') return 'warn'
  return 'crit'
}

// Le nom du scénario reste technique dans le tiroir ; ici il devient lisible.
function scenarioWord(model: string | undefined, name: string): string {
  if (model === 'A') return 'ligne dédiée'
  if (model === 'B') return 'forfait entreprise'
  if (model === 'emergency') return 'hors forfait'
  return name
}
```

`formatCostAR` is already imported there from `./format`.

Acceptance requirement carried from slice 20A's review: reports archived before
the verdict landed reload with `operator_verdict: {level: "", word: "",
sentence: ""}` — `omitempty` cannot suppress a struct value. An empty level must
render the absence motif ("Pas encore mesuré" with the neutral dot), never a
coloured badge, and never the engine's French sentence.

- [ ] **Step 5: Implement the hero**

Create `web/frontend/src/components/VerdictHero.tsx`:

```tsx
import { operatorFacts } from '../lib/dashboard'
import type { AuditReport } from '../lib/types'

const DOT: Record<string, string> = { neutral: 'var(--text-faint)', amber: 'var(--t-threshold)', red: 'var(--t-danger-text)', green: 'var(--t-ok)' }

// Le verdict vient du moteur. L'interface ne classe rien : elle montre le mot,
// la phrase et les quatre faits qui les justifient.
export function VerdictHero({ report, warnMs = 40, critMs = 100 }: {
  report: AuditReport | null | undefined
  warnMs?: number
  critMs?: number
}) {
  const v = report?.operator_verdict
  const level = v?.level ?? 'neutral'
  const word = v?.word ?? 'Pas encore mesuré'
  const sentence = v?.sentence ?? 'Aucun audit sur ce poste.'
  const facts = operatorFacts(report, warnMs, critMs)
  return (
    <section className="verdict" data-level={level} aria-labelledby="verdict-word">
      <p className="verdict-word" id="verdict-word">
        <span className="verdict-dot" style={{ background: DOT[level] }} aria-hidden="true" />
        {word}
      </p>
      <p className="verdict-sentence">{sentence}</p>
      <dl className="verdict-facts">
        {facts.map(f => (
          <div className="verdict-fact" key={f.key} data-empty={/Pas encore|Non chiffré/.test(f.value) || undefined}>
            <dt>{f.label}</dt>
            <dd>
              <b>{f.value}</b>
              {f.unit ? <small> {f.unit}</small> : null}
            </dd>
            <em>{f.target}{f.band ? ` · ${f.band.word}` : ''}</em>
          </div>
        ))}
      </dl>
    </section>
  )
}
```

Add to `index.css`: `.verdict { padding: var(--sp-5) 0; }`, `.verdict-word { font-size: var(--fs-hero); font-weight: 600; letter-spacing: -0.01em; margin: 0; }`, `.verdict-dot { width: 12px; height: 12px; display: inline-block; margin-right: 12px; }`, `.verdict-sentence { font-size: var(--fs-body); color: var(--text-body); max-width: 68ch; margin: var(--sp-3) 0 var(--sp-5); }`, `.verdict-facts { display: grid; grid-template-columns: repeat(4, minmax(0,1fr)); gap: var(--sp-5); margin: 0; }`, `.verdict-fact dt { font-size: var(--fs-label); color: var(--text-muted); }`, `.verdict-fact dd { margin: 4px 0 0; }`, `.verdict-fact dd b { font-family: var(--font-mono); font-size: var(--fs-metric); font-variant-numeric: tabular-nums; }`, `.verdict-fact em { font-style: normal; font-size: var(--fs-fine); color: var(--text-faint); }`, `.verdict-fact[data-empty] dd b { font-size: var(--fs-body); font-family: var(--font-ui); color: var(--text-muted); }`, and media rules collapsing to two columns at 1366px and one column at 700px.

- [ ] **Step 6: Verify and commit**

Run: `npm run typecheck && npx vitest run src/components/VerdictHero.dom.test.tsx src/lib/dashboard.test.ts src/lib/format.test.ts`
Expected: pass. Add two cases to `dashboard.test.ts` asserting `operatorFacts` returns four facts with the exact labels and that a missing cost yields `Non chiffré`.

```bash
git add web/frontend/src/components/VerdictHero.tsx web/frontend/src/components/VerdictHero.dom.test.tsx web/frontend/src/lib/dashboard.ts web/frontend/src/lib/dashboard.test.ts web/frontend/src/lib/types.ts web/frontend/src/styles/index.css
git commit -m "ui: le verdict du moteur et ses quatre faits, dont l'ariary"
```

---

### Task 4: Remediation card with a platform that cannot be wrong

**Files:**
- Create: `web/frontend/src/components/RemediationCard.tsx`
- Create: `web/frontend/src/components/RemediationCard.dom.test.tsx`
- Modify: `web/frontend/src/lib/types.ts` (new remediation fields), `web/frontend/src/lib/settings.ts` (persist the platform)

**Interfaces:**
- Consumes: `report.remediation.{up_measured_mbps,up_shaped_mbps,down_measured_mbps,down_shaped_mbps,linux_iface,routeros_wan_iface,routeros_lan_iface,routeros_lan_assumed,linux_tc,linux_tc_ingress,routeros,routeros_ingress}`.
- Produces: `<RemediationCard report onCopy onOpenDrawer />`; `copyTargetFor(rem, platform): string[]` returning the exact command lines.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/components/RemediationCard.dom.test.tsx`:

```tsx
import { describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { RemediationCard, copyTargetFor } from './RemediationCard'
import type { AuditReport } from '../lib/types'

const rem = {
  qdisc: 'cake',
  linux_tc: ['tc qdisc replace dev eth0 root cake bandwidth 4mbit'],
  routeros: '/queue type add name=cgo-wan kind=cake',
  routeros_ingress: '/queue tree add name=CGO-Download parent=bridge',
  linux_tc_ingress: ['tc qdisc replace dev ifb0 root cake bandwidth 17mbit'],
  up_measured_mbps: 5.1, up_shaped_mbps: 4.845,
  down_measured_mbps: 18.5, down_shaped_mbps: 17.575,
  routeros_lan_iface: 'bridge', routeros_lan_assumed: true,
}
const report = { audit_id: 'a1', measured: {}, remediation: rem } as unknown as AuditReport

describe('RemediationCard', () => {
  const card = (p: Platform) => renderToStaticMarkup(
    <RemediationCard report={report} platform={p} onPlatform={vi.fn()} onCopy={vi.fn()} onOpenDrawer={vi.fn()} />,
  )
  it('donne les deux sens avec leurs débits mesurés et façonnés', () => {
    const out = card('routeros')
    expect(out).toContain('Protéger ce lien')
    expect(out).toContain('4,8 Mbit/s')
    expect(out).toContain('sur 5,1 mesurés')
    expect(out).toContain('17,6 Mbit/s')
    expect(out).toContain('interface supposée « bridge »')
  })

  it('copie la syntaxe de la plateforme active, jamais un mélange', () => {
    expect(copyTargetFor(rem, 'routeros').join('\n')).toContain('/queue tree add name=CGO-Download')
    expect(copyTargetFor(rem, 'routeros').join('\n')).not.toContain('tc qdisc')
    expect(copyTargetFor(rem, 'linux').join('\n')).toContain('ifb0')
    expect(copyTargetFor(rem, 'linux').join('\n')).not.toContain('/queue type')
    expect(card('linux')).toContain('Linux tc')
  })

  it('ne propose rien sans remédiation mesurée', () => {
    const out = renderToStaticMarkup(<RemediationCard report={{ audit_id: 'a', measured: {} } as AuditReport} platform="routeros" onPlatform={vi.fn()} onCopy={vi.fn()} onOpenDrawer={vi.fn()} />)
    expect(out).toContain('Pas encore de configuration')
    expect(out).not.toContain('Copier')
  })
})
```

Add `import type { Platform } from './RemediationCard'` to the test file.

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/components/RemediationCard.dom.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement**

Create `web/frontend/src/components/RemediationCard.tsx`:

```tsx
import { numFr } from '../lib/format'
import type { AuditRemediation, AuditReport } from '../lib/types'

export type Platform = 'routeros' | 'linux'

const n1 = (v?: number) => (v == null || !Number.isFinite(v) ? '' : numFr(v, 1))

// Une copie aveugle casse un routeur : le presse-papier suit la plateforme
// active et ne porte que les sens réellement mesurés. La plateforme vit chez le
// parent, pour que la carte et le tiroir montrent la même syntaxe.
export function copyTargetFor(rem: AuditRemediation | undefined, platform: Platform): string[] {
  if (!rem) return []
  if (platform === 'routeros') {
    return [rem.routeros, rem.routeros_ingress].filter(Boolean).flatMap(s => (s as string).split('\n'))
  }
  return [...(rem.linux_tc ?? []), ...(rem.linux_tc_ingress ?? [])]
}

export function RemediationCard({ report, platform, onPlatform, onCopy, onOpenDrawer }: {
  report: AuditReport | null | undefined
  platform: Platform
  onPlatform: (p: Platform) => void
  onCopy: (text: string, label: string) => void
  onOpenDrawer: (tab: 'deploy' | 'proof') => void
}) {
  const rem = report?.remediation
  const hasRouterOS = !!(rem?.routeros || rem?.routeros_ingress)
  const hasLinux = !!(rem?.linux_tc?.length || rem?.linux_tc_ingress?.length)
  if (!rem || (!hasRouterOS && !hasLinux)) {
    return (
      <section className="remedy" data-empty>
        <h2>Pas encore de configuration</h2>
        <p>La capacité ou le temps de réponse au repos n’a pas été mesuré : une file réglée au hasard serait déployée à l’aveugle.</p>
      </section>
    )
  }
  const lines = copyTargetFor(rem, platform)
  const shaped = n1(platform === 'routeros' ? rem.up_shaped_mbps ?? rem.down_shaped_mbps : rem.up_shaped_mbps ?? rem.down_shaped_mbps)
  return (
    <section className="remedy">
      <h2>Protéger ce lien</h2>
      <p>Régulation active, calibrée sur ce qui a été mesuré (95 % du goulot).</p>
      <ul className="remedy-dirs">
        {rem.up_measured_mbps ? (
          <li><span>Envoi (montant)</span><b>{n1(rem.up_shaped_mbps)} Mbit/s</b><em>sur {n1(rem.up_measured_mbps)} mesurés</em></li>
        ) : <li><span>Envoi (montant)</span><b>non mesuré</b></li>}
        {rem.down_measured_mbps ? (
          <li>
            <span>Réception (descendant)</span><b>{n1(rem.down_shaped_mbps)} Mbit/s</b>
            <em>sur {n1(rem.down_measured_mbps)} mesurés · {platform === 'routeros' && rem.routeros_lan_assumed ? `interface supposée « ${rem.routeros_lan_iface} » — à vérifier avant de coller` : 'via IFB sur la passerelle'}</em>
          </li>
        ) : <li><span>Réception (descendant)</span><b>non mesurée</b><em>auditez le sens descendant pour un script de réception</em></li>}
      </ul>
      <div className="remedy-platform" role="radiogroup" aria-label="Plateforme cible">
        {hasRouterOS ? (
          <button type="button" role="radio" aria-checked={platform === 'routeros'} onClick={() => onPlatform('routeros')}>MikroTik RouterOS 7</button>
        ) : null}
        {hasLinux ? (
          <button type="button" role="radio" aria-checked={platform === 'linux'} onClick={() => onPlatform('linux')}>Linux tc</button>
        ) : null}
      </div>
      <div className="remedy-actions">
        <button className="btn btn-primary" onClick={() => onCopy(lines.join('\n'), platform === 'routeros' ? 'Configuration MikroTik' : 'Configuration Linux')}>
          Copier la configuration{shaped ? ` (${shaped} Mbit/s)` : ''}
        </button>
        <button className="btn-link" onClick={() => onOpenDrawer('deploy')}>Voir la commande exacte</button>
      </div>
    </section>
  )
}
```

`onPlatform` in the two views also calls `saveSetting('remediation-platform', p)` and persists to `localStorage['remediation-platform']`; the initial value comes from that key, defaulting to `routeros` when the report carries a RouterOS script and `linux` otherwise. Add `remediationPlatform?: Platform` to the settings type in `lib/settings.ts`. Add `.remedy h2 { font-size: var(--fs-title); }`, `.remedy-dirs li { display:grid; grid-template-columns: 1fr auto; gap: var(--sp-2); min-height:48px; align-items:baseline; }`, `.remedy-dirs b { font-family: var(--font-mono); font-size: var(--fs-metric); }`, `.remedy-platform button[aria-checked="true"] { color: var(--text-primary); box-shadow: inset 0 -2px 0 var(--t-live); }`, `.remedy-platform button { min-height:48px; background:none; border:0; font-size:var(--fs-body); color:var(--text-muted); }`.

- [ ] **Step 4: Verify and commit**

Run: `npm run typecheck && npx vitest run src/components/RemediationCard.dom.test.tsx`
Expected: pass.

```bash
git add web/frontend/src/components/RemediationCard.tsx web/frontend/src/components/RemediationCard.dom.test.tsx web/frontend/src/lib/types.ts web/frontend/src/lib/settings.ts web/frontend/src/styles/index.css
git commit -m "ui: la copie de remédiation choisit sa plateforme et ne promet pas un sens non mesuré"
```

---

### Task 5: Two-tab proof drawer

**Files:**
- Create: `web/frontend/src/components/ProofDrawer.tsx`
- Create: `web/frontend/src/components/ProofDrawer.dom.test.tsx`
- Modify: `web/frontend/src/views/AuditView.tsx`, `web/frontend/src/views/DashboardView.tsx`
- Reuse: `web/frontend/src/lib/modalFocus.ts`, `web/frontend/src/components/AuditReportCard.tsx`

**Interfaces:**
- Produces: `<ProofDrawer open tab report onClose onCopy />` with tabs `deploy` and `proof`, `role="tablist"`, arrow-key move, focus restored on close.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/components/ProofDrawer.dom.test.tsx`:

```tsx
import { describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { ProofDrawer } from './ProofDrawer'
import type { AuditReport } from '../lib/types'

const report = {
  audit_id: 'a1', site: 'Siège', link_type: 'fiber',
  measured: { load_attempted: true, load_measured: true, rtt_idle_p95_ms: 12, rtt_loaded_p95_ms: 40 },
  remediation: { qdisc: 'cake', linux_tc: ['tc qdisc replace dev eth0 root cake bandwidth 76mbit'], routeros: '/queue type add name=cgo kind=cake' },
} as unknown as AuditReport

describe('ProofDrawer', () => {
  it('ouvre sur le déploiement et nomme les deux onglets', () => {
    const out = renderToStaticMarkup(<ProofDrawer open tab="deploy" platform="routeros" report={report} onClose={vi.fn()} onCopy={vi.fn()} />)
    expect(out).toContain('role="tablist"')
    expect(out).toContain('Déploiement & scripts')
    expect(out).toContain('Preuves & télémétrie brute')
    expect(out).toContain('/queue type add name=cgo kind=cake')
    expect(out).not.toContain('tc qdisc replace dev eth0')
    expect(out).not.toContain('sha256')
  })
  it('suit la plateforme choisie sur la carte', () => {
    const out = renderToStaticMarkup(<ProofDrawer open tab="deploy" platform="linux" report={report} onClose={vi.fn()} onCopy={vi.fn()} />)
    expect(out).toContain('tc qdisc replace dev eth0')
    expect(out).not.toContain('/queue type')
  })
  it('met les sceaux et la référence 4G dans l’onglet preuve', () => {
    const out = renderToStaticMarkup(<ProofDrawer open tab="proof" platform="routeros" report={{ ...report, provenance: { source_sha256: 'a'.repeat(64), row_count: 3 } }} onClose={vi.fn()} onCopy={vi.fn()} />)
    expect(out).toContain('aabbbb')
    expect(out).toContain('Duel de référence')
  })
  it('ferme sans rapport', () => {
    expect(renderToStaticMarkup(<ProofDrawer open tab="proof" platform="routeros" report={null} onClose={vi.fn()} onCopy={vi.fn()} />)).toBe('')
  })
})
```

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/components/ProofDrawer.dom.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement**

Create `web/frontend/src/components/ProofDrawer.tsx`:

```tsx
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useModalFocus } from '../lib/modalFocus'
import { AuditReportCard } from './AuditReportCard'
import { copyTargetFor, type Platform } from './RemediationCard'
import { numFr } from '../lib/format'
import type { AuditReport } from '../lib/types'

export type DrawerTab = 'deploy' | 'proof'

// Deux onglets qui défilent chacun : chercher une commande ne doit pas faire
// défiler des sceaux, et vérifier un sceau ne doit pas faire défiler un script.
export function ProofDrawer({ open, tab, platform, report, onClose, onCopy }: {
  open: boolean
  tab: DrawerTab
  platform: Platform
  report: AuditReport | null | undefined
  onClose: () => void
  onCopy: (text: string, label: string) => void
}) {
  const [active, setActive] = useState<DrawerTab>(tab)
  const ref = useModalFocus(open, onClose)
  useEffect(() => setActive(tab), [tab, open])
  if (!open || !report) return null
  const lines = copyTargetFor(report.remediation, platform)
  const rem = report.remediation
  return createPortal(
    <div className="drawer-scrim" onClick={onClose}>
      <aside
        ref={ref}
        className="drawer"
        role="dialog"
        aria-modal="true"
        aria-label="Preuve et déploiement"
        onClick={e => e.stopPropagation()}
      >
        <header className="drawer-head">
          <div className="drawer-tabs" role="tablist" aria-label="Contenu de la preuve">
            <button role="tab" id="tab-deploy" aria-selected={active === 'deploy'} aria-controls="pane-deploy" onClick={() => setActive('deploy')}>Déploiement &amp; scripts</button>
            <button role="tab" id="tab-proof" aria-selected={active === 'proof'} aria-controls="pane-proof" onClick={() => setActive('proof')}>Preuves &amp; télémétrie brute</button>
          </div>
          <button className="btn-link" onClick={onClose}>Fermer</button>
        </header>
        {active === 'deploy' ? (
          <div id="pane-deploy" role="tabpanel" aria-labelledby="tab-deploy" className="drawer-pane">
            <h3>Commande exacte · {platform === 'routeros' ? 'MikroTik RouterOS 7' : 'Linux tc'}</h3>
            <pre className="drawer-code">{lines.join('\n') || 'aucune commande mesurée pour ce lien'}</pre>
            <button className="btn" onClick={() => onCopy(lines.join('\n'), 'Commande exacte')}>Copier ce bloc</button>
            <h3>Base de calibrage</h3>
            <p className="drawer-note">
              {rem?.up_measured_mbps ? `${numFr(rem.up_measured_mbps)} Mbit/s mesurés en envoi, façonnés à ${numFr(rem.up_shaped_mbps)} (95 %). ` : ''}
              {rem?.down_measured_mbps ? `${numFr(rem.down_measured_mbps)} Mbit/s mesurés en réception, façonnés à ${numFr(rem.down_shaped_mbps)}. ` : ''}
              {report.measured?.rtt_idle_p50_ms ? `Temps de réponse au repos ${numFr(report.measured.rtt_idle_p50_ms, 0)} ms.` : ''}
              {platform === 'routeros' && rem?.routeros_lan_assumed ? ` Interface LAN supposée « ${rem.routeros_lan_iface} » : vérifiez avec /interface print brief.` : ''}
            </p>
            <h3>Vérifier après collage</h3>
            <pre className="drawer-code">{platform === 'routeros' ? '/queue type print\n/queue tree print' : 'tc qdisc show'}</pre>
          </div>
        ) : (
          <div id="pane-proof" role="tabpanel" aria-labelledby="tab-proof" className="drawer-pane">
            <AuditReportCard report={report} onCopy={onCopy} />
          </div>
        )}
      </aside>
    </div>,
    document.body,
  )
}
```

`AuditReportCard` already renders the percentiles, both seals, the economics, the kernel ring and the 4G reference; it becomes the content of the proof tab unchanged, minus its `ScienceDetails` wrapper.

Add CSS: `.drawer-scrim { position:fixed; inset:0; background:rgba(0,0,0,.6); display:flex; justify-content:flex-end; z-index:60; }`, `.drawer { width:min(720px,100%); max-height:85vh; overflow:auto; background:var(--surface-card); }`, `.drawer-tabs button[aria-selected="true"] { color:var(--text-primary); box-shadow: inset 0 -2px 0 var(--t-live); }`, `.drawer-tabs button { min-height:48px; background:none; border:0; font-size:var(--fs-body); color:var(--text-muted); }`, `.drawer-code { font-family:var(--font-mono); font-size:var(--fs-fine); white-space:pre-wrap; word-break:break-word; background:var(--surface-elevated); padding:var(--sp-3); }`, `.drawer-pane { padding: var(--sp-4); }`.

- [ ] **Step 4: Wire it into the two views**

In `AuditView.tsx` and `DashboardView.tsx`, hold `const [drawer, setDrawer] = useState<{open:boolean;tab:DrawerTab;platform:Platform}>({open:false,tab:'deploy',platform:'routeros'})`, pass `onOpenDrawer={(t) => setDrawer(d => ({...d, open:true, tab:t}))}` to `RemediationCard`, lift `platform` into the parent, and render `<ProofDrawer …/>` once per view. Remove the old `ScienceDetails` wrapper around `AuditReportCard` in `AuditView.tsx`; the drawer replaces it.

- [ ] **Step 5: Verify and commit**

Run: `npm run typecheck && npx vitest run src/components/ProofDrawer.dom.test.tsx src/lib/modalFocus.test.ts`
Expected: pass, including the `platform="linux"` assertion added in Step 3.

```bash
git add web/frontend/src/components/ProofDrawer.tsx web/frontend/src/components/ProofDrawer.dom.test.tsx web/frontend/src/views/AuditView.tsx web/frontend/src/views/DashboardView.tsx web/frontend/src/styles/index.css
git commit -m "ui: la preuve se divise en déploiement et télémétrie, chacune pour soi"
```

---

### Task 6: Live latency trace on the observer

**Files:**
- Create: `web/frontend/src/lib/liveTrace.ts`
- Create: `web/frontend/src/lib/liveTrace.test.ts`
- Create: `web/frontend/src/components/LiveLatencyCanvas.tsx`
- Modify: `web/frontend/src/components/AuditLive.tsx`, `web/frontend/src/lib/types.ts`

**Interfaces:**
- Consumes: `/api/audit/live` `timeline: TimelinePoint[]`, `total_sec`, `phase`.
- Produces: `traceGeometry(points, width, height, totalSec): { segments: {x:number;y:number}[][]; gaps: number; yMax: number; objectiveY: number; marks: {t:number;label:string}[] }` — a pure function so the canvas is testable without a DOM.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/lib/liveTrace.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { traceGeometry } from './liveTrace'

const pt = (t: number, rtt: number | null, phase = 'baseline') => ({
  t, phase, n: rtt == null ? 0 : 1, rtt_p50: rtt, rtt_p95: rtt, small_p95: null,
})

describe('traceGeometry', () => {
  it('borne l’axe Y sur 220 ms au minimum et sur le max observé', () => {
    const g = traceGeometry([pt(0, 24), pt(1, 400)], 600, 120, 60)
    expect(g.yMax).toBeGreaterThanOrEqual(400)
    expect(g.segments[0].length).toBe(2)
  })
  it('ne relie jamais deux points séparés par une seconde muette', () => {
    const g = traceGeometry([pt(0, 24), pt(1, null), pt(2, 300)], 600, 120, 60)
    expect(g.segments.length).toBe(2)
    expect(g.gaps).toBe(1)
  })
  it('place les marqueurs de phase', () => {
    const g = traceGeometry([pt(0, 24, 'baseline'), pt(10, 300, 'charge')], 600, 120, 60)
    expect(g.marks.map(m => m.label)).toContain('charge')
  })
  it('rend un tracé vide sans inventer de point', () => {
    const g = traceGeometry([], 600, 120, 60)
    expect(g.segments).toHaveLength(0)
    expect(g.gaps).toBe(0)
  })
})
```

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/lib/liveTrace.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the geometry**

Create `web/frontend/src/lib/liveTrace.ts`:

```ts
export interface TracePoint {
  t: number
  phase: string
  n: number
  rtt_p50?: number | null
  rtt_p95?: number | null
  small_p95?: number | null
}

export interface TraceGeometry {
  segments: { x: number; y: number }[][]
  gaps: number
  yMax: number
  objectiveY: number
  marks: { t: number; label: string }[]
}

const OBJECTIVE_MS = 220

// Une seconde muette n'est pas un 0 ms : le tracé se rompt, l'opérateur voit
// l'absence plutôt qu'une latence parfaite.
export function traceGeometry(points: TracePoint[], w: number, h: number, totalSec: number): TraceGeometry {
  const vals = points
    .map(p => p.rtt_p95 ?? p.small_p95)
    .filter((v): v is number => typeof v === 'number' && v > 0)
  const yMax = Math.max(OBJECTIVE_MS, ...vals, 1)
  const span = Math.max(totalSec, 1)
  const x = (t: number) => (t / span) * w
  const y = (v: number) => h - (v / yMax) * h

  const segments: { x: number; y: number }[][] = []
  let current: { x: number; y: number }[] = []
  for (const p of points) {
    const v = p.rtt_p95 ?? p.small_p95
    if (typeof v !== 'number' || v <= 0) {
      if (current.length) segments.push(current)
      current = []
      continue
    }
    current.push({ x: x(p.t), y: y(v) })
  }
  if (current.length) segments.push(current)

  const marks: { t: number; label: string }[] = []
  let last = ''
  for (const p of points) {
    if (p.phase !== last) {
      marks.push({ t: p.t, label: p.phase })
      last = p.phase
    }
  }
  return { segments, gaps: Math.max(0, segments.length - 1), yMax, objectiveY: y(OBJECTIVE_MS), marks }
}
```

- [ ] **Step 4: Draw it**

Create `web/frontend/src/components/LiveLatencyCanvas.tsx`:

```tsx
import { useEffect, useRef } from 'react'
import { traceGeometry, type TracePoint } from '../lib/liveTrace'
import { paletteOnce } from '../lib/palette'

const H = 120

export function LiveLatencyCanvas({ points, totalSec }: { points: TracePoint[]; totalSec: number }) {
  const ref = useRef<HTMLCanvasElement>(null)
  useEffect(() => {
    const c = ref.current
    if (!c) return
    const dpr = Math.min(2, window.devicePixelRatio || 1)
    const w = c.clientWidth
    c.width = Math.max(1, w * dpr)
    c.height = H * dpr
    const ctx = c.getContext('2d')
    if (!ctx) return
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, H)
    const P = paletteOnce()
    const g = traceGeometry(points, w, H, totalSec)
    ctx.strokeStyle = P.hairline
    ctx.beginPath()
    ctx.moveTo(0, g.objectiveY)
    ctx.lineTo(w, g.objectiveY)
    ctx.stroke()
    ctx.strokeStyle = P.live
    ctx.lineWidth = 2
    for (const seg of g.segments) {
      if (seg.length < 2) continue
      ctx.beginPath()
      ctx.moveTo(seg[0].x, seg[0].y)
      for (const p of seg.slice(1)) ctx.lineTo(p.x, p.y)
      ctx.stroke()
    }
  }, [points, totalSec])
  return (
    <canvas
      ref={ref}
      className="live-trace"
      style={{ width: '100%', height: H }}
      role="img"
      aria-label={`Latence sur ${totalSec} secondes, objectif 220 millisecondes`}
    />
  )
}
```

Canvas cannot read CSS variables, which is why the two colours come from `paletteOnce()` — the same rule `lib/echarts.ts` already follows. A one-point segment is skipped rather than drawn as a dot, so a lone sample cannot read as a trend.

- [ ] **Step 5: Put it in the observer**

In `AuditLive.tsx`, add `timeline?: TracePoint[]` to the live type in `lib/types.ts`, render `<LiveLatencyCanvas points={live?.timeline ?? []} totalSec={live?.total_sec ?? 60} />` above the three facts, and rename the fact labels to “Temps de réponse”, “Réponse sous charge”, “Débit”. Keep the existing `ArmButton` for stop and the line “Arrêter n’écrit aucune mesure partielle.”

- [ ] **Step 6: Verify and commit**

Run: `npm run typecheck && npx vitest run src/lib/liveTrace.test.ts`
Expected: pass.

```bash
git add web/frontend/src/lib/liveTrace.ts web/frontend/src/lib/liveTrace.test.ts web/frontend/src/components/LiveLatencyCanvas.tsx web/frontend/src/components/AuditLive.tsx web/frontend/src/lib/types.ts
git commit -m "ui: la trace de latence couvre toute la fenêtre et rompt sur les sondes muettes"
```

---

### Task 7: Dashboard, audit and evidence screens on the new parts

**Files:**
- Modify: `web/frontend/src/views/DashboardView.tsx`, `web/frontend/src/views/AuditView.tsx`, `web/frontend/src/views/EvidenceView.tsx`, `web/frontend/src/components/AuditDeclare.tsx`, `web/frontend/src/components/StatusRibbon.tsx`
- Modify: `web/frontend/src/lib/dashboard.ts` (`bothAbsent` becomes unused — delete it and its test case)

**Interfaces:**
- Consumes: `VerdictHero`, `RemediationCard`, `ProofDrawer`, `/api/duel` via `lib/duel.ts` (Task 8), `loadCapabilities`.

- [ ] **Step 1: Dashboard**

Rewrite `DashboardView.tsx` to: header (title 28px + one subtitle), `<VerdictHero>`, `<RemediationCard>`, one proof line (`Preuve : {runs} mesures archivées · {valid} valides · sceau {hash8}` with a text link to the evidence destination), one primary action `Auditer le lien` / `Auditer de nouveau`, and the drawer. Delete: the two-card `dash-grid`, the cyan and violet tag spans, the `dash-quiet` negative band, the `syn-proof` four-cell grid, the `syn-next` card, the `syn-actions` strip, and the `dash-onboarding` block. Keep the single `/api/integrity` + `/api/results` fetch, and keep `useModeActive` gating it.

- [ ] **Step 2: Audit**

In `AuditView.tsx`: delete the `audit-steps` list and the idle card; the declaration form is the screen. Label the four fields “Où êtes-vous ?”, “Quel type de lien ?”, “Combien de temps ?”, “Quel sens ?”; make the duration a two-button toggle “Complet 60 s” / “Rapide 20 s · aperçu” defaulting to 60; primary button “Lancer l’audit · 60 s” following the toggle. Move `iface`, `routeros_iface`, `routeros_lan`, `target`, `small_url`, `bulk_addr`, `qdisc` under “Réglages avancés”. In the read state, render `<VerdictHero>` + `<RemediationCard>` + `<ProofDrawer>` and drop the inline `OperatorStats`/`ScienceDetails` pair.

- [ ] **Step 3: Evidence**

In `EvidenceView.tsx`: keep the four counts but at `--fs-metric`, one line for the seal with a copy action, primary “Vérifier l’intégrité”, report rows as `site · date · verdict word` with “Ouvrir” and “Preuve”, and fold the runs table plus replay under “Détails du corpus”. Move the full SHA-256 block inside the proof drawer.

- [ ] **Step 4: Ribbon**

In `StatusRibbon.tsx`, keep mode, SSE and the running audit; replace the “VÉRIFIER” button with a text link “Preuves”; the last-audit item shows the verdict word instead of the grade letter.

- [ ] **Step 5: Verify and commit**

Run: `npm run typecheck && npx vitest run && npm run build`
Expected: green. Fix assertions that named deleted classes; do not restore them.

```bash
git add web/frontend/src/views web/frontend/src/components/AuditDeclare.tsx web/frontend/src/components/StatusRibbon.tsx web/frontend/src/lib/dashboard.ts web/frontend/src/lib/dashboard.test.ts web/frontend/src/styles/index.css web/frontend/dist/index.html
git commit -m "ui: tableau de bord, audit et preuves passent au verdict et au tiroir"
```

---

### Task 8: Bench preset duel and research mode

**Files:**
- Create: `web/frontend/src/lib/duel.ts`, `web/frontend/src/lib/duel.test.ts`
- Create: `web/frontend/src/components/DuelCard.tsx`, `web/frontend/src/components/DuelCard.dom.test.tsx`
- Modify: `web/frontend/src/views/BenchView.tsx`, `web/frontend/src/components/BenchConfig.tsx`, `web/frontend/src/store/bench.ts`

**Interfaces:**
- Produces: `fetchDuel(profile: 'P2', run?: string): Promise<Duel | null>`; `Duel = { base, best, deltaDeadlinePts, deltaSmallMs, reason }`; `researchMode: boolean` in the UI store, persisted.

- [ ] **Step 1: Write the failing tests**

Create `web/frontend/src/lib/duel.test.ts`:

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { normalizeDuel } from './duel'

const payload = {
  profile: 'P2', deadline_ms: 220, min_reps: 3,
  base: { qdisc: 'pfifo_fast', cc: 'bbr', deadline_pct: 48.8, small_p95_ms: 328.2, n: 22, runs: ['run-a'] },
  best: { qdisc: 'fq_codel', cc: 'bbr', deadline_pct: 96.2, small_p95_ms: 205.4, n: 18, runs: ['run-b'] },
  delta_deadline_pts: 47.4, delta_small_ms: -122.8,
}

describe('normalizeDuel', () => {
  beforeEach(() => vi.unstubAllGlobals())
  it('garde la paire et ses runs', () => {
    const d = normalizeDuel(payload)
    expect(d?.best?.qdisc).toBe('fq_codel')
    expect(d?.best?.runs).toEqual(['run-b'])
    expect(d?.deltaSmallMs).toBe(-122.8)
  })
  it('refuse une paire sous la garde d’échantillonnage', () => {
    expect(normalizeDuel({ ...payload, best: { ...payload.best, n: 1 } })).toBeNull()
    expect(normalizeDuel({ profile: 'P2', min_reps: 3, reason: 'aucune cellule' })).toBeNull()
  })
})
```

Create `web/frontend/src/components/DuelCard.dom.test.tsx`:

```tsx
import { describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { DuelCard } from './DuelCard'
import type { Duel } from '../lib/duel'

const duel = {
  profile: 'P2', minReps: 3,
  base: { qdisc: 'pfifo_fast', cc: 'bbr', deadlinePct: 48.8, smallP95Ms: 328.2, n: 22, runs: ['run-a'] },
  best: { qdisc: 'fq_codel', cc: 'bbr', deadlinePct: 96.2, smallP95Ms: 205.4, n: 18, runs: ['run-b'] },
  deltaDeadlinePts: 47.4, deltaSmallMs: -122.8,
} as Duel

describe('DuelCard', () => {
  it('nomme les protocoles et cite ses runs', () => {
    const out = renderToStaticMarkup(<DuelCard duel={duel} onLaunch={vi.fn()} onResearch={vi.fn()} onProof={vi.fn()} />)
    expect(out).toContain('pfifo_fast / bbr')
    expect(out).toContain('fq_codel / bbr')
    expect(out).toContain('22 mesures')
    expect(out).toContain('run-a')
    expect(out).toContain('Lancer la comparaison')
  })
  it('dit qu’il n’y a rien à comparer sans inventer de gagnant', () => {
    const out = renderToStaticMarkup(<DuelCard duel={null} onLaunch={vi.fn()} onResearch={vi.fn()} onProof={vi.fn()} />)
    expect(out).toContain('aucune base exploitable')
    expect(out).not.toContain('fq_codel')
  })
})
```

- [ ] **Step 2: Run to verify failure**

Run: `npx vitest run src/lib/duel.test.ts src/components/DuelCard.dom.test.tsx`
Expected: FAIL — modules not found.

- [ ] **Step 3: Implement the reader and the card**

Create `web/frontend/src/lib/duel.ts`:

```ts
export interface DuelCell { qdisc: string; cc: string; deadlinePct: number; smallP95Ms: number; n: number; runs: string[] }
export interface Duel { profile: string; minReps: number; base: DuelCell | null; best: DuelCell | null; deltaDeadlinePts: number; deltaSmallMs: number }

const num = (v: unknown): number | null => (typeof v === 'number' && Number.isFinite(v) ? v : null)

function cell(v: unknown): DuelCell | null {
  if (!v || typeof v !== 'object') return null
  const j = v as Record<string, unknown>
  const n = num(j.n) ?? 0
  const qdisc = typeof j.qdisc === 'string' ? j.qdisc : ''
  if (!qdisc || n < 3) return null
  return {
    qdisc, cc: typeof j.cc === 'string' ? j.cc : '',
    deadlinePct: num(j.deadline_pct) ?? 0, smallP95Ms: num(j.small_p95_ms) ?? 0,
    n, runs: Array.isArray(j.runs) ? j.runs.filter((r): r is string => typeof r === 'string') : [],
  }
}

// Le gagnant vient du moteur, avec sa garde n ≥ 3. Sans paire éligible, la
// carte le dit : aucun repli sur une cellule d'une répétition.
export function normalizeDuel(raw: unknown): Duel | null {
  if (!raw || typeof raw !== 'object') return null
  const j = raw as Record<string, unknown>
  const base = cell(j.base)
  const best = cell(j.best)
  if (!base || !best) return null
  return {
    profile: typeof j.profile === 'string' ? j.profile : 'P2',
    minReps: num(j.min_reps) ?? 3, base, best,
    deltaDeadlinePts: num(j.delta_deadline_pts) ?? 0,
    deltaSmallMs: num(j.delta_small_ms) ?? 0,
  }
}

export async function fetchDuel(profile = 'P2', run = ''): Promise<Duel | null> {
  try {
    const q = new URLSearchParams({ profile })
    if (run) q.set('run', run)
    const r = await fetch(`/api/duel?${q}`)
    if (!r.ok) return null
    return normalizeDuel(await r.json())
  } catch {
    return null
  }
}
```

Create `web/frontend/src/components/DuelCard.tsx`:

```tsx
import type { Duel } from '../lib/duel'

const pct = (v: number) => v.toFixed(1)
const ms = (v: number) => v.toFixed(0)

// Le banc est le laboratoire : ici les noms de protocoles sont le sujet, pas
// du jargon à cacher. Ils viennent de la cellule mesurée, jamais d'un preset.
export function DuelCard({ duel, onLaunch, onResearch, onProof }: {
  duel: Duel | null
  onLaunch: () => void
  onResearch: () => void
  onProof: () => void
}) {
  if (!duel?.base || !duel.best) {
    return (
      <section className="duel" data-empty>
        <h2>Duel de référence — profil P2</h2>
        <p>aucune base exploitable : il faut au moins trois mesures valides par cellule.</p>
        <button className="btn btn-primary" onClick={onLaunch}>Lancer la comparaison</button>
      </section>
    )
  }
  const side = (label: string, c: typeof duel.base) => (
    <div className="duel-side">
      <span>{label}</span>
      <b className="duel-protocols">{c.qdisc} / {c.cc}</b>
      <p><em>{pct(c.deadlinePct)} %</em> des alertes à l’heure · {ms(c.smallP95Ms)} ms sous charge</p>
      <small>{c.n} mesures · runs {c.runs.join(', ')}</small>
    </div>
  )
  return (
    <section className="duel">
      <h2>Duel de référence — profil P2 (4G chargée)</h2>
      <div className="duel-pair">
        {side('Ligne de base', duel.base)}
        {side('Meilleure protection mesurée', duel.best)}
      </div>
      <p className="duel-gain">
        Sur ce scénario : {duel.deltaDeadlinePts > 0 ? `+${pct(duel.deltaDeadlinePts)}` : pct(duel.deltaDeadlinePts)} points d’alertes à l’heure, {duel.deltaSmallMs > 0 ? '+' : ''}{ms(duel.deltaSmallMs)} ms de réponse sous charge.
      </p>
      <div className="duel-actions">
        <button className="btn btn-primary" onClick={onLaunch}>Lancer la comparaison</button>
        <button className="btn-link" onClick={onProof}>Voir la preuve</button>
        <button className="btn-link" onClick={onResearch}>Mode recherche</button>
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Preset and research mode**

In `store/ui.ts` add `researchMode: boolean` initialised from `localStorage['bench-research']` with a setter that persists it. In `BenchConfig.tsx`, when `researchMode` is false render only the profile choice, the repetitions and the launch action, with the preset line “P2 · 3 répétitions · 4 flux · canal propre · objectif 220 ms”, and send `deadline_ms: 220` in the start body; when true, render the current full control set unchanged. In `BenchView.tsx`, put `<DuelCard …>` at the top of the results section and gate the leaderboard, filters and TradeSpace behind `researchMode`.

- [ ] **Step 5: Verify and commit**

Run: `npm run typecheck && npx vitest run src/lib/duel.test.ts src/components/DuelCard.dom.test.tsx src/lib/benchLayout.test.ts`
Expected: pass.

```bash
git add web/frontend/src/lib/duel.ts web/frontend/src/lib/duel.test.ts web/frontend/src/components/DuelCard.tsx web/frontend/src/components/DuelCard.dom.test.tsx web/frontend/src/views/BenchView.tsx web/frontend/src/components/BenchConfig.tsx web/frontend/src/store/ui.ts
git commit -m "banc: le duel nommé par les données et le mode recherche séparés"
```

---

### Task 9: Lexicon gate, QA walk, captures

**Files:**
- Create: `web/frontend/src/lib/lexicon.test.ts`
- Modify: `web/frontend/qa-walk.mjs`

- [ ] **Step 1: Write the gate**

Create `web/frontend/src/lib/lexicon.test.ts`:

```ts
// Le lexicon est une porte, pas une recommandation : un terme de moteur qui
// remonte sur une surface opérateur est une régression.
import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const SURFACES = [
  'views/DashboardView.tsx', 'views/AuditView.tsx', 'views/EvidenceView.tsx',
  'components/VerdictHero.tsx', 'components/RemediationCard.tsx',
  'components/StatusRibbon.tsx', 'components/AuditDeclare.tsx', 'components/TopNav.tsx',
]
const BANNED = ['p50', 'p95', 'qdisc', 'CAKE', 'fq_codel', 'pfifo_fast', 'BBR', 'CUBIC', 'JFI', 'QDI', 'SLA', 'bufferbloat', 'eBPF', 'manifeste', 'corpus', 'twin', 'garanti', 'garantie', 'prouvé']
// Le banc est un laboratoire : ses noms de protocoles sont son sujet.
const ALLOWED_IN = new Set(['views/BenchView.tsx', 'components/BenchConfig.tsx'])

const src = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')
const visibleText = (s: string) => (s.match(/>([^<>{}]+)</g) ?? []).join(' ')

describe('lexique opérateur', () => {
  for (const f of SURFACES) {
    it(`${f} ne montre pas de vocabulaire de moteur`, () => {
      if (ALLOWED_IN.has(f)) return
      const text = visibleText(src(f))
      for (const term of BANNED) {
        expect(text, `${f} affiche « ${term} » sur une surface`).not.toContain(term)
      }
    })
  }
  it('aucune phrase française en capitales sur les surfaces', () => {
    for (const f of SURFACES) {
      const css = src(f)
      expect(css, f).not.toMatch(/text-transform:\s*uppercase/)
    }
  })
})
```

If a legitimate label needs a banned word, move that label into the drawer rather than adding an exception to the list.

- [ ] **Step 2: Extend the QA walk**

In `qa-walk.mjs`, add steps: verdict word visible at 1920/1366/390; exactly one `.btn-primary` per destination; drawer opens on `Voir la commande exacte`, both tabs reachable by arrow keys, focus returns to the trigger; no horizontal overflow at 390px; the live canvas has an accessible name; the duel card names its protocols. Keep the existing long-task attribution step and its 400 ms budget.

- [ ] **Step 3: Run the walk and capture**

Start the built binary on a spare port, then:

```bash
node qa-walk.mjs http://127.0.0.1:9093
node capture-fresh.mjs http://127.0.0.1:9093 shots-console
```
Expected: zero failures, ten captures. Record the `TOTAL gz` from the build.

- [ ] **Step 4: Commit**

```bash
git add web/frontend/src/lib/lexicon.test.ts web/frontend/qa-walk.mjs
git commit -m "qa: le lexique devient une porte et la marche couvre le tiroir"
```

---

### Task 10: Design system, product card, book

**Files:**
- Modify: `DESIGN.md`, `PRODUCT.md`, `docs/api.md`
- Modify (uncommitted): `book/content/04-pratique.js`, `book/content/09-annexes.js`

- [ ] **Step 1: DESIGN.md**

Replace §1 (product truth), §2 (colour), §3 (typography), §4 (layout), §5 (components) with the console: one sans, the six-step scale, the three surfaces, top navigation with a mobile bottom bar, `VerdictHero`, `RemediationCard`, `ProofDrawer`, `DuelCard`, the seven-gate cascade table, and the eleven honesty invariants from the spec §6. Delete the Cormorant row and the mode-rail description.

- [ ] **Step 2: PRODUCT.md**

Reorder Users: layman operator first, network engineer second, jury third. Replace the UI row of “What ships today” with the console description, the verdict field, `/api/duel`, `/api/capabilities`, the frozen campaign tier and the guard. Remove “Existing layout and tokens only; no redesign in this wave.” Make the economics paragraph say which figure comes from a contract scenario and which from a mobile tier.

- [ ] **Step 3: API doc**

Add `/api/capabilities`, `/api/duel`, the `operator_verdict` field, `timeline`, the new remediation fields, `economics.*.numbers`, and the 409 on `POST /api/cost/tier` during a run.

- [ ] **Step 4: Book**

In `book/content/04-pratique.js`: rewrite the Tableau 13 rows (`tbl412`) so each mode's role matches the console, and rewrite the surrounding paragraph at line 10 to name the verdict and the drawer. Fix the shaper sentence at line 73 to 95 %, with the reason, or state both figures with their scope. In `book/content/09-annexes.js`: rewrite Annex E screen by screen against the new captures, add the seven-gate table and the `n ≥ 3` duel rule to Annex F, and keep the traceability note.

- [ ] **Step 5: Rebuild the book and verify nothing drifted**

```bash
cd book && node build-docx.js && node check-numbers.js
```
Expected: the numeric audit reports no changed figure; `memoire.docx` changes only in prose. Do not commit `book/`.

- [ ] **Step 6: Commit the tracked documents**

```bash
git add DESIGN.md PRODUCT.md docs/api.md
git commit -m "docs: le système de design devient la console opérateur"
```

---

### Task 11: Close the slice

- [ ] **Step 1: Full gates.** Run `python tools/verify.py` from `C:\cgo`. Expected `6 passed, 0 failed, 0 not run`; a skipped gate is not green.
- [ ] **Step 2: VERIFY rows** for the console, the lexicon gate, the duel card and the drawer: builder-pass, commands, date, host, and one line each for what stays unproven (no live audit, no live campaign, no eBPF, no booted CHR).
- [ ] **Step 3: CHECKPOINTS entry** with branch, HEAD, changed files, measured bundle size, captures path, and the unproven list.
- [ ] **Step 4: Plan index** — move slice 20B to `Windows-green`, name the verifier pass as the unblock condition, and record that 20A and 20B stay on this branch until the human decides on `main`.
- [ ] **Step 5: Commit** with explicit paths only.

## Self-review notes

- Spec coverage: §3.1 → Tasks 1-2; §3.2 → Task 3; §3.3 → Task 4; §3.4 → Task 5 and Task 8 of 20A for the TUI keys; §3.5 → Task 6; §3.7/§3.8 → Task 8; §3.9 → Task 7; §4 → Tasks 1-2; §5 → Task 9; §6 → Tasks 3-8 tests plus Task 9; §7 → Task 3 and Task 7; §9 → Task 10.
- The TUI pane keys (`d`, `p`, `c` writing to `data/audits/<id>/`) are not in this file: they belong to slice 20A Task 8 or a follow-up TUI task, and the spec §3.4 records them. Do not silently drop them.
- `platform` is owned by the view in Tasks 4 and 5, so the card and the drawer cannot disagree; the drawer test asserts both values.
- Field names match slice 20A exactly: `operator_verdict`, `load_attempted`, `up_measured_mbps`, `up_shaped_mbps`, `down_measured_mbps`, `down_shaped_mbps`, `routeros_lan_iface`, `routeros_lan_assumed`, `numbers`, `timeline`, `/api/duel`.
