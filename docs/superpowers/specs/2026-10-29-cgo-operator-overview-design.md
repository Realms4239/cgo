# CGO operator overview and navigation — design

*Status: implemented (builder pass; verifier pending) · 2026-10-29 · Slice 14 ·
Branch: `rebuild/plan1-economics-audit`*

## 1. Why

The dashboard opened directly on Campagne. An operator arriving for the first
time saw a matrix form, a live panel, a results table and an archive view with
no single answer to “where do I start?”. All four views were mounted at boot,
so ECharts and the 159-run tables did work before the operator asked for them.
The chart toggles (`groupes`, `craft`, `source`) occupied the header even on
views with no chart.

## 2. Decisions

- **D1 — Synthèse is the landing view.** One screen answers: what is frozen,
  what the P2 corpus says, and what to do next.
- **D2 — The constat is an aggregate, labelled as such.** It uses
  `/api/results` groups with a valid deadline sample, not the final campaign;
  the text says “agrégat de tout le corpus gelé, pas la campagne finale”.
  The book keeps the exact Table 15 campaign values.
- **D3 — Real SVG icons.** The rail’s Unicode glyphs are replaced by one
  authored 24×24 stroke set in `NavIcon`.
- **D4 — Contextual chrome.** `groupes`/`craft`/`source` appear only on Live
  and Résultats, where charts exist.
- **D5 — Deferred mounting.** Heavy views mount on first visit and then stay
  mounted; returning remains instant, but the first paint no longer runs every
  view’s effects.
- **D6 — No new dependency, token, or colour.** Existing primitives and
  tokens only; the new CSS is token-driven.

## 3. Acceptance

- `npm test` green, including the new `synthese` unit tests (41 files).
- `npm run build` green; no new dependency; bundle stays within the existing
  check-bundle budget.
- Screenshots at 1920×1080 and 1366×1080 show the Synthèse with real counts
  (159/208/74/246) and readable labels.
- Full six gates green.

## 4. Out of scope

- A visual redesign of the four existing views.
- New data claims, a new chart, or a new backend endpoint.
- Production access or a live bench run.