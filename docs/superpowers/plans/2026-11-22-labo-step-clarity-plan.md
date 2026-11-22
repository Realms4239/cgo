# Labo Step Clarity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give each Labo chapter one honest outcome sentence and explain where Configurer's declaration values come from, without changing any value or adding any control.

**Architecture:** Add an optional `outcome` prop to the presentational `Chapter` component, supply four method-output sentences from `LaboView`, add one context line to `Configurer`, and style both with existing CSS tokens. No state, store, API or rail changes.

**Tech Stack:** React/TypeScript, CSS custom properties, Vitest + happy-dom, Playwright on Edge.

## Global Constraints

- French user-facing copy; English identifiers. Outcome sentences describe method output — never a measurement, PASS, seal, ranking or verdict. Absence words stay exact.
- No new dependency, hex, font-size, accent, motion, component file or interactive control. Existing tokens only.
- WCAG 2.2 AA, forced-colors, reduced-motion; no overflow/page scroll at 890/1280/1920.
- Protected `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, frozen corpus never edited. No `git add -A`; explicit paths only. No merge.
- Native gates once; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Outcome lines and declaration context

**Files:**
- Modify: `web/frontend/src/components/labo/Chapter.tsx` (props + one `<p>`)
- Modify: `web/frontend/src/views/LaboView.tsx` (four `outcome` props)
- Modify: `web/frontend/src/components/labo/Configurer.tsx` (one context `<p>`)
- Modify: `web/frontend/src/styles/components.css` (`.chapter-outcome`, `.configurer-context`)
- Test: new `web/frontend/src/components/labo/Chapter.clarity.dom.test.tsx`

**Interfaces:**
- Consumes: existing `ChapterProps` and the four existing chapters in `LaboView`.
- Produces: `ChapterProps.outcome?: string` — optional so existing callers stay valid.

- [ ] **Step 1: Write the failing DOM oracle**

Create `Chapter.clarity.dom.test.tsx` using the same `renderToStaticMarkup` setup as the other labo DOM tests. Assert: (a) a `Chapter` with `outcome="…"` renders `p.chapter-outcome` with the exact text; (b) without `outcome`, no `.chapter-outcome` node; (c) the outcome text contains no digit and none of `PASS|FAIL|VÉRIFIÉ|verdict|sceau|classement`; (d) `Configurer` renders a `.configurer-context` line whose text states the values come from the real report and keeps absence words. Run it and watch it fail on missing nodes.

- [ ] **Step 2: Add the optional outcome to `Chapter`**

In `Chapter.tsx`, add `outcome?: string` to `ChapterProps`, destructure it, and render `<p className="chapter-outcome">{outcome}</p>` immediately after `.chapter-lede` only when defined. No `aria-hidden`, no new wrapper div.

- [ ] **Step 3: Supply four real sentences from `LaboView`**

Add one `outcome` prop to each `Chapter` call:
- Configurer: `« Vous verrez la déclaration réelle du lien et les profils disponibles. »`
- Observer: `« Vous verrez la matrice de progression, les portes et la chronologie d'une campagne. »`
- Comparer: `« Vous verrez les mesures classées, le duel et l'espace de décision. »`
- Preuves: `« Vous verrez l'état des fichiers de preuve, les archives scellées et les preuves vérifiables. »`

These describe what is displayed, never a result. Do not add digits, verdicts or promises of measurement.

- [ ] **Step 4: Add the Configurer context line**

In `Configurer.tsx`, add one `<p className="configurer-context">` before the `MetricRow`: `« Ces valeurs viennent du rapport réel ; une déclaration absente garde son mot. »`. Do not change any `StateSlot` value or `word` prop.

- [ ] **Step 5: Style with existing tokens only**

In `components.css`, add `.chapter-outcome { margin: var(--sp-1) 0 0; font-size: var(--fs-fine); color: var(--text-muted); text-wrap: pretty; }` and `.configurer-context` with the same typography. No new hex, font-size, `@keyframes` or `transition`.

- [ ] **Step 6: Run covering and full tests**

Run `npx vitest run Chapter.clarity.dom.test.tsx` plus all existing labo DOM tests (`Observer.dom.test.tsx`, `Observer.depth.dom.test.tsx`, `Comparer.dom.test.tsx`, `Preuves.dom.test.tsx`, `PipelineRail.dom.test.tsx`, `laboCraft.test.ts`) → all green. Run `npm test` with 2 workers → all green. Run `npm run build` → exit 0; restore tracked `dist/index.html` with `git checkout --`.

- [ ] **Step 7: Rebuild, Playwright, screenshots**

`go build -o web/frontend/cgo-test-server.exe ./cmd/cgo` so the embedded SPA carries the change. Run unmocked Edge: `CGO_API_TOKEN=cgo-local-dev-7f3a91d2 CGO_SMOKE_REAL=1 PWCHROMIUM='C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe' ./node_modules/.bin/playwright test e2e/redesign-smoke.spec.ts e2e/redesign-layout.spec.ts e2e/redesign-a11y.spec.ts e2e/redesign-feel.spec.ts --timeout=90000` → all green. Capture and read 890×512 / 1280×768 / 1920×1080 screenshots of all four chapters; record the visual verdict.

- [ ] **Step 8: Gates and commit**

Run `sh .githooks/run-gates.sh` once; record `links` FAIL and `book/` SKIPs without editing. Commit: `git add web/frontend/src/components/labo/Chapter.tsx web/frontend/src/views/LaboView.tsx web/frontend/src/components/labo/Configurer.tsx web/frontend/src/styles/components.css web/frontend/src/components/labo/Chapter.clarity.dom.test.tsx && git commit -m "feat(labo): chapter outcome lines and declaration context"`. Use the documented no-hooks override only if a pre-existing gate blocks the hook, and record why.

## Self-Review

- Spec coverage: §3 → Steps 2-5; §4 honesty → Steps 1+3+4; §5 a11y → Steps 1+7; §6 verification → Steps 6-8; §7 out of scope respected.
- No placeholders, no `TBD`, no "similar to Task N", no unspecified handling.
- Type consistency: `ChapterProps.outcome?: string` is used identically in `Chapter.tsx` and all four `LaboView` call sites; CSS class names match the DOM oracle.
