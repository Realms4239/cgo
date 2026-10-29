# CGO operator overview and navigation — implementation plan

*2026-10-29 · Slice 14 · Spec:
`docs/superpowers/specs/2026-10-29-cgo-operator-overview-design.md`*

## Task 1 — Synthèse view

- Add `views/SyntheseView.tsx`: proof counts, aggregate P2 constat with
  valid-only samples, next action, four shortcuts.
- Add `lib/synthese.ts` with `pickHeadline` and `nextStep`, plus unit tests.
- Fetch `/api/integrity`, `/api/results`, `/api/health`; a missing corpus is
  said, never zero-filled.

## Task 2 — Navigation

- Add `synthese` to `PANELS`, default the store to it, and keep the keyboard
  keys 1–5 aligned.
- Add `components/NavIcon.tsx` (authored SVG set) and use it in `Rail`.
- Show `groupes`/`craft`/`source` only on `live` and `resultats`.

## Task 3 — Deferred mounting

- In `App.tsx`, mount `campagne`, `live`, `resultats`, `integrite` on first
  visit, then keep them mounted. `synthese` mounts at boot.

## Task 4 — Styles

- Add token-driven `.syn*` rules to `styles/index.css`; 4-column proof grid,
  two-column constat/next action, four shortcuts, responsive at 900px.

## Task 5 — Gates

```bash
cd web/frontend && npm test
cd web/frontend && npm run build
python tools/verify.py
```

Record the screenshots, the test count and the six-gate result. No live run.