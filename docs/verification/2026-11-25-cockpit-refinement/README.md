# Cockpit UI refinement — proof record · 2026-11-25

*Builder record for the cockpit refinement slice (plan `docs/superpowers/plans/2026-11-25-cockpit-ui-refinement-plan.md`, design `docs/superpowers/specs/2026-11-25-cockpit-ui-refinement-design.md`). Host: local worktree on Windows (git bash). No field measurement, no screenshot and no VERIFY-row promotion here — this is an SDD implementation proof, not a product claim.*

## Scope and commits

Baseline: `836a916fcc6723d5b34476b7cc1753791f846bbe` (`docs: plan cockpit UI refinement`).

Three implementation commits (the "three task commits"):

| Task | Commit | Subject |
|---|---|---|
| 1 — shared cockpit type/rhythm tokens | `dd791d2696a90558a2938310774adca42a2d02bc` | `refactor(ui): shared cockpit type and rhythm tokens` |
| 2 — decision column placement | `831fe325071a27df850d9275ace7053991f032cb` | `refactor(ui): align cockpit decision column` |
| 3 — instrument field placement | `70b879fd24828f070d575b2a3a1910032349cdc4` | `refactor(ui): align cockpit instrument field` |

Two mechanical follow-ups are part of Tasks 1 and 3 respectively, not separate tasks:
`6b3dc09fa73e917b0cd6732c7e0d7f0704301cff` (`test(ui): redirect hero assertion to shared cockpit alias`) and
`d9b80aa5ccf23c0fbe9d6a7a55ac651cbd5275fd` (`test(ui): remove unused symbols from cockpit field oracle`).

Across all five commits the touched files are exactly:
`web/frontend/src/styles/tokens.css`, `web/frontend/src/styles/components.css`,
`web/frontend/src/components/cockpit/CockpitHierarchy.dom.test.tsx`,
`web/frontend/src/components/cockpit/CockpitField.dom.test.tsx`,
`web/frontend/src/components/cockpit/DecisionSpine.dom.test.tsx`.

## Commands, exact exits

Commands from brief Step 1, run in order from the repo root:

| # | Exact command | Exit | Result (verbatim summary) |
|---|---|---|---|
| 1 | `npm -C web/frontend run typecheck` | **0** | `run tsc --noEmit` — no diagnostics |
| 2 | `npm -C web/frontend test -- --run` | **1** | `Test Files 2 failed / 101 passed (103)`; `Tests 3 failed / 581 passed (584)` |
| 3 | `sh .githooks/run-gates.sh` | **0** | `6 passed, 0 failed, 0 not run` |

Additional focused proof (run to isolate the cockpit contract from the unrelated full-suite failures):

| Exact command | Exit | Result |
|---|---|---|
| `npm -C web/frontend test -- --run src/components/cockpit/CockpitHierarchy.dom.test.tsx src/components/cockpit/CockpitField.dom.test.tsx src/components/cockpit/DecisionSpine.dom.test.tsx src/components/HonestStates.dom.test.tsx` | **0** | `Test Files 4 passed (4)`; `Tests 43 passed (43)` |

Gate 3 detail, each line verbatim: `PASS frozen` (159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid; 22 cellules, 159 manifestes; `docs/frozen-manifests.sha256`: 159 manifests), `PASS provenance` (159 headers, 5 eras, 0 valuation sidecars), `PASS tests` (`go vet + go test in all modules`), `PASS exit` (5 safe invocations; verdict contract ok), `PASS hygiene` (1255 tracked files clean; legacy allowlist 18 backup archives and `results/runs-frozen.tar.gz`; 4 named tailnet addresses), `PASS links` (5 documents scanned; plan and spec links resolve; 12 named historical absent-path allowances).

**The full frontend suite is RED — not claimed green.** Exit 1 is reported, never suppressed. The three failures are:

1. `src/lib/e2eContract.test.ts` → `accessibility.spec.ts ne cible que les vues vivantes` — `accessibility.spec.ts contient le jeton supprimé v-evidence`.
2. `src/lib/e2eContract.test.ts` → `operator-cockpit.spec.ts ne cible que les vues vivantes` — `operator-cockpit.spec.ts contient le jeton supprimé v-dashboard`.
3. `src/lib/title.test.ts` → `PRODUCT.md porte le titre et le sous-titre` — `PRODUCT.md` carries the adopted academic title, not the older string the test pins.

Failures 1–2 are the known deliberate RED oracle for the separate stale-E2E-spec debugging task (`web/frontend/src/lib/e2eContract.test.ts` is uncommitted; it was not read, edited, staged or committed). Failure 3 is a pre-existing stale title assertion: `git diff --name-only 836a916 d9b80aa` lists only the five cockpit files above, so `title.test.ts` / `PRODUCT.md` are untouched by this slice. None of the three is attributed to the cockpit refinement. Per brief Step 1 a non-zero exit would normally stop the task; the harness direction for this task instead required running the focused cockpit oracles and reporting the full-suite red honestly, which is what this record does.

## New DOM assertions

Net-new `expect(...)` assertions added across the five commits: **22**, counted from the committed test diffs (`+` minus `-` on `expect(` lines):

- `dd791d2` → `CockpitHierarchy.dom.test.tsx`: **11** (4 token declarations, 4 rendered-DOM verdict-before-facts + single-verdict checks, 3 alias-consumption checks).
- `6b3dc09` → `DecisionSpine.dom.test.tsx`: **+1 net** (the stale raw-var assertion was split into an alias-consumption check plus a token-target check; stronger, no weakened assertion).
- `831fe32` → `CockpitHierarchy.dom.test.tsx`: **3** (shared-rhythm / single-hairline case).
- `70b879f` → `CockpitField.dom.test.tsx`: **7** (3 CSS-placement assertions, 4 honest-absence assertions).
- `d9b80aa` → `CockpitField.dom.test.tsx`: **0** (removed two unused TS symbols only; no assertion change).

All 22 assertions pass in the focused run (43 tests across four files, exit 0). No existing assertion was weakened or deleted.

## Untouched identity and layout constraints

Per spec §1, §3, §4 and §6, these were deliberately left unchanged:

- Visual identity: bright institutional portal, white sheet on light-gray ground, teal rule, charcoal heading; no new colour, no glass/glow/gradient/decorative shadow, no new font, no focus-ring or `prefers-reduced-motion` change.
- Layout: named `grid-template-areas` and DOM order unchanged (verdict word before facts, single `#verdict-word`); no layout-area reordering, no keyboard/screen-reader order change.
- Spacing: 4 px grid via existing token layer; no magic numbers. Existing `@media` breakpoints and overrides were not edited (the base-layer shared-rhythm alias is overridden at those breakpoints, as before).
- Hairlines: existing `--hairline*` tokens and `min-height: 72px` / `padding-top: var(--sp-3)` / `padding: var(--sp-4) 0` values kept verbatim.
- State contract: absent/error states still render honest wording, never a verdict chip, fabricated number or green; French operator-facing strings unchanged.
- Boundaries: no gate, permit, ledger, corpus, config or frozen-data edit; no dependency added; no new route, API or measurement-logic change; no VERIFY-row promotion. Explicit paths only — no `git add -A`.

## What stayed unproven

- **Browser-only E2E not run.** Playwright specs (`accessibility.spec.ts`, `operator-cockpit.spec.ts`, etc.) were not executed in a real browser; the two `e2eContract.test.ts` failures are static token scans of those spec files, not browser results. Real-browser layout, keyboard focus order and WCAG contrast are not proven by this record.
- **Axe accessibility check not run** (spec §5 names it; no axe result was produced here — no fabricated report).
- **Bundle budget check not run** — no budget raised and no measured gz size produced here.
- **Full frontend suite is red** (3 out-of-scope failures above); a green full suite is not claimed.
- **Visual polish not screenshot-verified** — no screenshot was taken or fabricated.
- **Gate 3 `tests` covers Go only** (`go vet + go test in all modules`); it does not re-run the frontend suite, so its PASS does not overwrite the full-suite red.
- **Non-independent controller verification.** Tasks 1–3 were implemented and verified by the same controller flow (reviewer dispatch was blocked by provider free-usage limits); no independent reviewer signature exists for the three commits.
- **Follow-up `6b3dc09` was verified by the controller, not an independent agent**, and is recorded explicitly as such.

Builder pass only. A verifier must re-run the commands above and own any VERIFY-row change; this record promotes no row.
