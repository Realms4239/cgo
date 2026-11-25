# Plan — repair the stale four-view probe specs (2026-11-25)

Evidence: the full unmocked E2E pass left 15 spec files red. All are pre-existing exploratory/probe specs written for the deleted four-view dashboard. Errors are hard-coded `https://localhost:9090`, removed `[data-panel=...]` selectors, removed nav labels, and one null `getBoundingClientRect`. The live app has only two destinations (`[data-mode="cockpit"]` / `[data-mode="labo"]`), roots `#v-cockpit` / `.labo-view`, and chapter ids `#chapter-*`.

## Global Constraints (binding for every task)

- No assertion or gate is deleted or weakened; no `page.route('/api/**')`, no stub server, no invented value. An assertion whose original target truly no longer exists is reported BLOCKED, not dropped.
- Specs navigate through Playwright `baseURL` (`page.goto('/')`). No hard-coded `C:/cgo`, `localhost:9090`, or other host.
- Real compiled Go binary on 127.0.0.1:19091, `CGO_API_TOKEN` generated in the environment (never written or printed), Edge via `PWCHROMIUM='C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'`, `--timeout=90000`. One spec file at a time when a shared read bucket causes 429; never widen the server limit.
- No new dependency, no manifest change, no production code change, no host contact, no frozen corpus / `book/data/runs` / ledger / gate change.
- French comments and user-facing text; English identifiers and commit messages.
- One logical commit per task with explicit paths. Never `git add -A`, no merge, no force push, no branch deletion.
- Verification per task: named E2E specs green against the real server, `npm -C web/frontend run typecheck`, `npm -C web/frontend test -- --run`, `sh .githooks/run-gates.sh` printing `6 passed, 0 failed, 0 not run`. Exit 2 is never green.

## Task 1 — chart and visual probe family

Files: `e2e/chart-probe.spec.ts`, `e2e/compare-probe.spec.ts`, `e2e/visual-audit.spec.ts`, `e2e/wall.spec.ts`, `e2e/wall-splatter.spec.ts`, `e2e/screenshots.spec.ts`, `e2e/sheet.spec.ts`, `e2e/smoke.spec.ts`.

Retarget each to live selectors and baseURL while preserving the original assertion intent (canvas renders, layout has no horizontal overflow, screenshots exist, keyboard/sheet behaviour). Read the live component source before choosing a selector. The screenshots spec must capture the live two-mode UI, not removed `data-panel` tabs. Report BLOCKED for any assertion whose target no longer exists.

## Task 2 — flow and navigation probe family

Files: `e2e/control-demo.spec.ts`, `e2e/deep-probe.spec.ts`, `e2e/m8.spec.ts`, `e2e/m9.spec.ts`, `e2e/madalink.spec.ts`, `e2e/debug-390.spec.ts`, `e2e/multi-depth.spec.ts` (tests 1-9; test 10 was already retargeted).

Retarget to `[data-mode]`, `#v-cockpit`, `.labo-view`, `.pipeline-rail`, `.pr-step`, `#chapter-*`. For `debug-390`, fix the null `getBoundingClientRect` by selecting an element that actually exists at 390 px; keep the geometry assertion. No test deletion.

## Task 3 — full verification

Rebuild nothing (no Go change). Run the full E2E suite once with a fresh server; if the shared principal bucket exhausts, run spec files in separate invocations and report that shape. Expect all specs green. Run typecheck, frontend suite and six gates; report exact commands, exits and counts. No commit unless a real defect appears, in which case stop and report.
