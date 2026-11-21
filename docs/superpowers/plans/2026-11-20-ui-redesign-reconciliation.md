# Reconciliation of the UI redesign into `main` (2026-11-21)

One clean commit integrates the reviewed UI redesign with `main`'s separate
governance/security lineage without discarding either.

## Inputs

- `main` @ `7ab9331` — governance, security, auth, CI, and a deliberate removal
  of 60 legacy files (`cmd/cgo-gui`, `cmd/meteolink`, `npm/`, …).
- `operator-cockpit` @ `0287b4c` — first execution of the plan (T1–T9).
- `sdd-ui-redesign` @ `88078e3` — full SDD redo, 17 commits, all task reviews and
  the final whole-branch review clean after one fix wave. This is the authority
  for the UI tree: it is the most rigorously reviewed implementation.

`main` and the redesign share merge base `f8a1317`; `acaf474` (the redesign base)
is not an ancestor of `main`, so a literal three-way merge is a 761-file conflict.
The clean result therefore keeps `main`'s tree and overlays only the reviewed UI.

## What was done

- New branch/worktree `main-reconciled` from `main` (no worktree removed).
- Overlaid the reviewed UI sources from `sdd-ui-redesign`: `web/frontend/src`,
  `e2e`, `scripts`, `public`, exact-pinned `package.json`/`bun.lock`, Vite/Vitest/
  TS/Playwright configs, `index.html`, `slice-capture.mjs`. Removed 44 main-only
  legacy UI files that the redesign deletes. Kept main's `dist/`, `embed.go`,
  `embed_test.go` and all Go/governance files intact.
- Carried the plan + process notes from `operator-cockpit` (new files).
- Honest adaptations required by `main`'s stricter backend:
  - `src/lib/tui.test.ts` retargeted from the deleted `cmd/meteolink/main.go` to
    the real consolidated `cmd/cgo/tui.go` (`func runTUI`, `tea.NewProgram`).
  - `playwright.config.ts` forwards `CGO_API_TOKEN` as a Bearer header when set.
    Authentication stays fully enabled; no `--dev-no-auth` was used.
  - `redesign-smoke.spec.ts` accepts the two REAL outcomes of ARM on this host:
    observing, or an honest visible refusal (`Échec audit`, HTTP 501 in Windows
    observe mode). It never fabricates a started campaign.
  - Restored main's `audit-verdict-sync.json` / `auditFixtures.ts`, which a Go
    fixture-identity test requires.

## Verification on the integrated tree

- `./node_modules/.bin/tsc --noEmit` → 0
- `bun run build` → ok, TOTAL 310.0 KB gz (≤600)
- `bun run test` → 85 files, 409/409 passed
- Edge Playwright (`PWCHROMIUM`, `--timeout=90000`, `CGO_API_TOKEN`) → 9/9 twice,
  including axe 0 critical/serious at 890/1280/1920 in both modes. One earlier
  target-size failure under concurrent load did not reproduce in two quiet reruns.
- `main`'s native gate runner `sh .githooks/run-gates.sh` → `3 passed, 1 failed,
  2 not run`. The plan's `tools/verify.py` does not exist in this lineage.
  The single failure (`links`, 28 missing paths in `docs/VERIFY.md` /
  `docs/CHECKPOINTS.md`) is PRE-EXISTING: a clean worktree at `main` HEAD reports
  the same failure (27 paths, the difference is duplicate lines). Both ledgers
  are protected from edits, so this is recorded, not papered over. `frozen` and
  `provenance` skip because main removed `book/`.

## Integration

Advance `main` by fast-forward only. The `main` worktree has unrelated local
changes (protected); no force operation and no ledger edit was used. Both
existing worktrees remain in place.
