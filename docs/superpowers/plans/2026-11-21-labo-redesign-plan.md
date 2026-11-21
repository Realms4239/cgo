# Labo redesign plan (2026-11-21)

Authority: `docs/research/2026-11-21-labo-deep-research.md` (Exa-only, 95 sources, 5 deep-read) + the ratified spatial spec amendment. No new dependency. French copy. Absent NEVER zero; no green without a real act; word+shape states; WCAG AA, 44px, forced-colors, reduced-motion; 890–1920 no overflow. Zero ledger churn; one commit per task with explicit paths; never `git add -A`; no merge. Gates: `./node_modules/.bin/tsc --noEmit && bun run build && bun run test`; Edge Playwright with `PWCHROMIUM`, `--timeout=90000`, real Go binary, no `page.route`; native gates `sh .githooks/run-gates.sh` (pre-existing `links` false positive on protected ledgers).

### Task 1: Rail keyboard model and progress word

Add Left/Right/Home/End navigation to the Labo step rail while keeping `aria-current="step"` (stepper semantics, not tablist). Add a visible "Étape n sur 4" word and ensure step labels wrap rather than truncate at 890. Keep rail targets ≥44px and hidden panels inert/aria-hidden with focus moved to the active heading. Files: `src/components/labo/PipelineRail.tsx`, `src/views/LaboView.tsx`, `src/styles/components.css`, DOM/e2e tests. Verification: focused DOM test for keyboard navigation + progress word; `playwright test e2e/redesign-layout.spec.ts e2e/redesign-a11y.spec.ts --timeout=90000`; full suite green.

### Task 2: Question-led panels and honest event timeline

Make each panel answer exactly one question in its heading/lede; keep advanced controls behind disclosures. Add a compact timeline strip to Observer built **only** from real `instrument.events`, with explicit no-data/grey state and no fabricated severity. Files: `src/components/labo/{Observer,Comparer,Preuves}.tsx`, `src/views/LaboView.tsx`, `src/styles/components.css`, tests. Verification: DOM test that an empty event ring renders no-data and a real event renders a marker; full tsc/build/test + all e2e + forbidden diff; one commit.

### Task 3: Final whole-branch review

Final code reviewer over the redesign diff; one fix wave + scoped re-review if needed.
