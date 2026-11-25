# Plan — E2E stack repair and real Playwright demo (2026-11-25)

Evidence: full unmocked E2E against the real Go binary + Edge. Task 1 fixed a real defect
(`api/server.go` rate-limited the SPA document, so a refresh returned JSON 429); after
rebuilding the binary, the live-spec subset went from 30 failures to 9, then to 5 when each
spec file ran with a fresh server. All remaining failures are stale four-view specs or the
documented 429 transient from the declared per-principal rate limit (60 reads/min, burst 20;
10 mutations/min, burst 5). No app assertion is weakened and no API is mocked.

## Global Constraints (binding for every task)

- No assertion, gate, rate limit, permit, ledger, frozen corpus or protected document is weakened, deleted or widened. A 429 may only be treated as a transient with a bounded, Retry-After-respecting retry — never as a pass.
- No mocked API: no `page.route('/api/**')`, no stub server. E2E runs against the real compiled Go binary on 127.0.0.1:19091, auth active via `CGO_API_TOKEN` from the environment (throwaway value, never written to a file or printed). The Go binary MUST be rebuilt after any `api/` change.
- Edge only, via `PWCHROMIUM='C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'`, `--timeout=90000`.
- No new dependency. No hard-coded `C:/cgo` or `https://localhost:9090` in specs. No `^`, `~`, `@latest` in manifests.
- French comments and user-facing text; English identifiers and commit messages.
- One logical commit per task, explicit paths only. Never `git add -A`. Never merge to `main`. Never force push. Never delete a branch. No live bench.
- Do not touch `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, `book/`, `data/`, `epbfBBR.md`, or any user document.
- Verification per task: `npm -C web/frontend run typecheck`, `npm -C web/frontend test -- --run`, and `sh .githooks/run-gates.sh` (must print `6 passed, 0 failed, 0 not run`). Exit 2 is never green.

## Task 1 — Scope rate limiting to `/api/*` only — COMPLETE

Commit `04e2da3`, reviewer PASS. Non-API paths now bypass the limiter; API values, buckets, Retry-After and error body are unchanged.

## Task 2 — Retarget stale four-view specs to the live two-mode UI

Confirmed failures (each reproduced alone):
- `e2e/accessibility.spec.ts:52` — clicks `#chapter-preuves .archive-table button "CITER"`; the cell intercepts the pointer and the four-view chapter no longer exists.
- `e2e/archives.spec.ts:5` — waits for button "Provenance"; removed with the PanelChooser.
- `e2e/archives.spec.ts:15` — expects `.panel-chooser`; removed.
- `e2e/audit-instrumented.spec.ts:38` — expects header text "connecté"; the live chrome says `EN LIGNE` via TrustStrip, and the four-view `VIEWS` loop targets removed nav buttons.
- `e2e/observatory2.spec.ts:9` — waits for nav button "Campagne"; removed.
- `e2e/multi-depth.spec.ts:168` — hard-codes `https://localhost:9090` and clicks removed `[data-panel="live"]`.

Implement: preserve each test's *intent and assertion strength* using only live selectors (`[data-mode]`, `#v-cockpit`, `.labo-view`, `.trust-word`, `[data-slot]`, live chapter ids). For the CITER case, verify the current live control before rewriting; if the equivalent action genuinely no longer exists, report BLOCKED rather than deleting the assertion. No `page.route`, no invented state, no hard-coded host.

Done when: each named test passes against the real binary with Edge; `src/lib/e2eContract.test.ts` stays green; typecheck/frontend suite/six gates green. One explicit-path commit.

## Task 3 — Treat 429 as a transient in E2E assertions (no gate change)

Confirmed failures:
- `e2e/fullstack-api.spec.ts:156` — 14 mutation calls exceed the mutate burst of 5; the 6th gets 429 instead of the expected 501.
- `e2e/fullstack-interface.spec.ts:265` — the console collector records 429 resource responses accumulated from earlier tests in the same file; running the axe tests alone passes 3/3, proving the false positive.

Implement: add a small bounded E2E fixture helper that retries a real request on 429, honouring `Retry-After`, and use it for the mutation loop and direct read assertions that must receive a real status. In the console collector, exclude **only** 429 responses with an explicit comment naming the verified reason (declared per-principal limiter, not an application error); every other console error still fails the test. Do not change the server, the limits, or any expected final status.

Done when: both named tests pass against the real binary; all other assertions in those files still pass; typecheck/frontend suite/six gates green. One explicit-path commit.

## Task 4 — Full real E2E pass and inline demo captures

After Tasks 2–3, rebuild the Go binary and run the full E2E suite unmocked with Edge and a generated `CGO_API_TOKEN`. If a single long run still exhausts the shared principal bucket, run spec files in separate invocations (fresh server per file) and report that explicitly — do not widen the limit. Expect all specs green. Then capture the real app at widths 890, 1280 and 1920 in both `Synthèse` and `Labo` into a temp directory; verify each PNG exists and is non-trivial. Report exact commands, exit codes, counts and screenshot paths. Do not commit screenshots to the repo.

Done when: every spec passes (with the execution shape stated), six captures exist, six gates green. No code commit unless a real defect appears — then stop and report it.
