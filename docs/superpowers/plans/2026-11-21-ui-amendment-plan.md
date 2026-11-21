# UI amendment plan (2026-11-21)

Source of truth: `docs/research/2026-11-21-spatial-spec-amendment.md` (human-directed amendment to the ratified spatial spec). 21st MCP patterns ingested: weather/stat cards (label·value·unit), single-line status bar, content stepper, incident timeline; decorative marquee rejected.

Global constraints: French copy; absent NEVER zero; word+shape state; WCAG AA, 44px targets, forced-colors, reduced-motion; 890–1920 no page overflow; bundle ≤600 KB gz; exact pins; zero ledger churn; one commit per task with explicit paths; never `git add -A`; no merge. Host: no make, `python tools/verify.py` is replaced on this lineage by `sh .githooks/run-gates.sh` (pre-existing `links` failure is a verified false positive on protected ledgers). Frontend gates: `./node_modules/.bin/tsc --noEmit && bun run build && bun run test`, plus Edge Playwright with `PWCHROMIUM`, `--timeout=90000`, real Go binary, no `page.route`.

Workspace: `C:/cgo/.worktrees/main-reconciled` (branch `main-reconciled` → fast-forward `main` after each verified task).

### Task 1: Labo swipe panels

Replace the sequential vertical document with four full-viewport panels switched by the pinned rail. Files: `src/components/labo/{PipelineRail,Chapter}.tsx`, `src/views/LaboView.tsx`, `src/styles/components.css`, affected tests. Requirements: rail buttons with `aria-current="step"`; panel track `translateX(-index*100%)`; 180 ms transform-only transition; reduced-motion instant; only active panel visible/focusable (`inert` + `aria-hidden`); focus moves to active heading; no page scroll; no new dependency; 44px rail targets. Verification: tsc/build/test; `playwright test e2e/redesign-layout.spec.ts e2e/redesign-a11y.spec.ts e2e/redesign-smoke.spec.ts --timeout=90000`; assert panel switch has no vertical scroll and hidden panels are inert.

### Task 2: Footer status line

Redesign `TrustStrip` into a one-glance status line (identity + version left, quiet status words centre, help/settings right if not duplicated in chrome). Keep words `EN LIGNE`/`HORS LIGNE`, `ABSENT`, `AUCUNE PERTE`, `n trames perdues`; word+shape; quiet-until-abnormal via one slot, not a whole-bar banner; 44px controls. Files: `src/components/shell/TrustStrip.tsx`, `src/styles/index.css`, tests. Verification: tsc/build/test; axe at 890/1280/1920; DOM tests for the preserved words and abnormal flag.

### Task 3: Synthèse rename + card mechanics, full verification

Finish the `Cockpit` → `Synthèse` label rename across app/tests/e2e (ids stay `cockpit`), apply 21st-derived stat-card mechanics (label·value·unit, tabular nums, quiet borders) to the Synthèse facts without new dependencies, and run the complete gate set. Verification: full suite, native gates, forbidden diff, one commit.

### Task 4: Final whole-branch review

Dispatch a final code reviewer over the whole amendment diff; address findings in one fix wave + scoped re-review.
