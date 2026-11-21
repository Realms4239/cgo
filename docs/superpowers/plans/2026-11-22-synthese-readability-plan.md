# Synthèse Readability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Synthèse field footer and ruler skimmable by regrouping the honesty fields into three labelled groups, without changing any value, state or honesty rule.

**Architecture:** Presentation-only change in `TimeField.tsx` and `components.css`. The footer becomes three labelled label→value grids; the ruler legend adopts the same typography. The existing ECharts grids, deterministic ruler inset, ring buffers and all data sources are untouched.

**Tech Stack:** React/TypeScript, CSS custom properties, Vitest + happy-dom, Playwright on Edge.

## Global Constraints

- French user-facing copy; English identifiers. Absence stays `—` + a French word — never zero, green, or a fabricated value.
- No new dependency, hex, accent, font size, motion or component file. Use existing tokens only.
- Keep `containLabel: false` and grid `left:64 / right:16`; the CSS ruler-track inset must stay derived from those values.
- WCAG 2.2 AA, 44px N/A (no new control), forced-colors, reduced-motion; no overflow at 890/1280/1920.
- Protected `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, frozen corpus never edited. No `git add -A`; explicit paths only. No merge.
- Native gates once; pre-existing `links` false positive and `book/` SKIPs recorded, never edited.

---

### Task 1: Regroup footer and align ruler legend

**Files:**
- Modify: `web/frontend/src/components/cockpit/TimeField.tsx` (footer JSX ~lines 309-345, ruler legend ~272-277)
- Modify: `web/frontend/src/styles/components.css` (`.tf-ruler-*` ~223-233, `.tf-footer*` ~246-260)
- Test: new `web/frontend/src/components/cockpit/TimeField.readability.dom.test.tsx`

**Interfaces:**
- Consumes: existing `live.iface`, JFI and provenance values already read by `TimeField.tsx`; existing `ifaceField`, `equityWord`, `trace` helpers.
- Produces: three labelled groups with class names `.tf-footer-group` (already exists), `.tf-footer-title`, `.tf-footer-grid`; no new exported API.

- [ ] **Step 1: Write the failing DOM oracle**

Create `TimeField.readability.dom.test.tsx` using the same render setup as `TimeField.spatial.dom.test.tsx`. Assert: (a) exactly three `.tf-footer-group` elements whose `.tf-footer-title` text is `Interface`, `Équité`, `Provenance`; (b) with empty instrument state, every `.tf-footer-grid` value contains `—` and a French absence word, and no `\d` digit appears in any absent value; (c) with real `live.iface` data, the RX/TX/name values appear inside the Interface group; (d) no group is empty and no `.tf-footer-group` carries a fabricated-green class. Run it and watch it fail (missing groups).

- [ ] **Step 2: Read the current footer and CSS completely**

Read `TimeField.tsx:220-350` and `components.css:220-262` in full. List the exact existing fields and their helper functions before changing markup. No field may be dropped.

- [ ] **Step 3: Restructure the footer markup**

In `TimeField.tsx`, replace the run-on footer line with three `.tf-footer-group` sections. Keep the existing field order within Interface (name, RX, TX, window, drops RX, drops TX, source, gap), Équité (index, gauge) and Provenance (trace). Each value keeps its current helper output unchanged. Use a `<dl>`-style label/value pair or the existing `.tf-footer-grid` with explicit label spans — choose the one that preserves AT reading order. No `aria-hidden` on meaningful text.

- [ ] **Step 4: Apply typography tokens in CSS**

In `components.css`, set `.tf-footer-title` to uppercase with `var(--tracking-caps)`, `.tf-footer-grid` label/value pairs to a two-column grid with `column-gap: var(--sp-3)`, values to `font-variant-numeric: tabular-nums`, and absence text to `text-wrap: pretty`. Apply the same label/value typography to `.tf-ruler-legend > span`. Do not add any new hex, font-size declaration or `@keyframes`. Keep the forced-colors block and the `.tf-ruler-track` inset rule untouched.

- [ ] **Step 5: Run the covering tests**

Run `npx vitest run TimeField.readability.dom.test.tsx TimeField.spatial.dom.test.tsx` → exit 0. Run `npm test` → all existing tests green (472 + new). Run `npm run build` → exit 0, bundle within budget. Restore tracked `dist/index.html` with `git checkout -- web/frontend/dist/index.html` after the build.

- [ ] **Step 6: Unmocked Playwright validation**

Rebuild the Go binary `go build -o web/frontend/cgo-test-server.exe ./cmd/cgo` so it embeds the new dist. Run `CGO_API_TOKEN=cgo-local-dev-7f3a91d2 CGO_SMOKE_REAL=1 PWCHROMIUM='C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe' ./node_modules/.bin/playwright test e2e/redesign-smoke.spec.ts e2e/redesign-layout.spec.ts e2e/redesign-a11y.spec.ts e2e/redesign-feel.spec.ts --timeout=90000` → all green. Capture 890×512 / 1280×768 / 1920×1080 screenshots, read them, and record the visual verdict.

- [ ] **Step 7: Gates and commit**

Run `sh .githooks/run-gates.sh` once; record the pre-existing `links` FAIL and `book/` SKIPs without editing them. Commit with explicit paths: `git add web/frontend/src/components/cockpit/TimeField.tsx web/frontend/src/styles/components.css web/frontend/src/components/cockpit/TimeField.readability.dom.test.tsx && git commit -m "feat(cockpit): skimmable field footer and ruler legend"`. Use the documented no-hooks override only if a pre-existing gate blocks the hook, and record why.

## Self-Review

- Spec coverage: §3 structure → Steps 3-4; §4 honesty → Step 1+3; §5 a11y → Steps 1+6; §6 verification → Steps 5-7; §7 out of scope respected.
- No placeholders, no `TBD`, no "similar to Task N", no unspecified edge-case handling.
- Type consistency: existing helper names and existing CSS class names are reused; `GapRow`-style new types are not introduced.
