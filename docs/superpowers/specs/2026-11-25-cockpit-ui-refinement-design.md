# Cockpit UI refinement — design (2026-11-25)

Authority: operator chat 2026-11-25 asked for a large UI refinement and polishing pass across components, items and placement, then chose "Cockpit first", "Both, with tests", "Keep identity, refine within it", and "All cockpit components", and approved this design.

Out of scope: live bench and live measurement work is a separate scope with its own spec, same-day task-scoped token, named quiet window, exact rollback line and same-day rehearsal. No bench cell is part of this slice. Labo and shared-layer refinements are later slices.

## 1. Scope

Refine every cockpit-visible component within the visual identity already pinned by
`DESIGN.md`: bright institutional portal, white sheet on light gray ground, teal rule,
charcoal heading, no glass/glow/gradient/decorative shadow. No new dependency, no new
route, no data-semantics change, no change to the frozen corpus, gates, permits, ledgers
or protected config.

The named `grid-template-areas` and DOM order stay as they are: keyboard and
screen-reader order must remain truthful. The work is refinement inside that contract.

## 2. Component inventory

Every cockpit-visible component is in scope: `TimeField`, `AnswerBand`,
`DecisionSpine`, `CockpitInstrument`, `JournalRail`, plus the audit/live/report/prompt
components that render on the cockpit surface. Changes are grouped so each step is
reviewable; no unrelated component outside the cockpit is touched.

## 3. Visual contract

- One spacing rhythm: a 4 px grid for padding, gaps and aligned baselines. Spacing
  values come from the existing token layer; no magic numbers in components.
- One type scale across all cockpit components: heading, section label, value,
  supporting text and caveat. Weights and line-heights are declared once in the token
  layer and reused.
- Prominence rule: the measured verdict and its primary action stay first; provenance,
  caveats and method notes remain adjacent and visibly secondary. Absence is never
  rendered as zero or as a green verdict.
- Hairlines and alignment: use the existing `--hairline*` tokens; align labels and
  values to a shared column grid; no new shadow or border treatment.
- Focus rings stay 2 px teal; `prefers-reduced-motion` keeps instant state swaps.

## 4. State contract

Every in-scope component that renders measured data or a verdict exposes the states
that apply to it from: loading, honest absence, stale, error and live. State wording is
plain French, consistent across components, and never invents a measurement. A
component in an absent or error state must not render a verdict chip or a numeric
result. This tightens existing behaviour rather than adding new data.

## 5. Proof and acceptance

- Existing DOM tests stay; no assertion is weakened or deleted.
- Add focused DOM assertions for hierarchy (verdict before provenance), placement
  (shared alignment grid) and the five-state wording on representative components.
- Keep the axe accessibility check and the bundle budget check; no budget raised.
- `sh .githooks/run-gates.sh` (or `make verify` when `make` exists) runs once per
  implementation step; exit 0 means all six gates passed, 1 a named gate failed,
  2 a gate could not run. Builder passes do not promote a VERIFY row.
- On a host without `make`, name the fallback command in the final report.

## 6. Non-goals

No new animation library, no new font, no color or token replacement beyond adding
shared spacing/type declarations, no layout-area reordering, no changes to measurement
logic, API, gate definitions, ledgers, frozen corpus or user documents.
