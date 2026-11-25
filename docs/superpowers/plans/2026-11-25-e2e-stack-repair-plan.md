# Plan — E2E stack repair and real Playwright demo (2026-11-25)

Source evidence: full E2E run against the real Go binary + Edge produced 6 failures;
focused re-runs and direct probes reduced them to one real server defect and two stale specs.
The two 1920px layout failures passed on focused re-run (order/state dependent, not a defect).
Everything below is bounded to those confirmed causes.

## Global Constraints (binding for every task)

- No assertion, gate, rate limit, permit, ledger, frozen corpus or protected document is weakened, deleted or widened.
- No mocked API: no `page.route('/api/**')`, no stub server. E2E runs against the real compiled Go binary on 127.0.0.1:19091, auth active via `CGO_API_TOKEN` from the environment (throwaway value, never written to a file or printed).
- Edge only, via `PWCHROMIUM='C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'`, `--timeout=90000`.
- No new dependency. No hard-coded `C:/cgo` or `https://localhost:9090` in specs. No `^`, `~`, `@latest` in manifests.
- French comments and user-facing text; English identifiers and commit messages.
- One logical commit per task, explicit paths only. Never `git add -A`. Never merge to `main`. Never force push. Never delete a branch. No live bench.
- Do not touch `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, `book/`, `data/`, `epbfBBR.md`, or any user document.
- Verification per task: `npm -C web/frontend run typecheck`, `npm -C web/frontend test -- --run`, `go test -count=1 ./...` for touched Go packages, and `sh .githooks/run-gates.sh` (must print `6 passed, 0 failed, 0 not run`). Exit 2 is never green.

## Task 1 — Scope rate limiting to `/api/*` only

**Root cause (already proven):** `api/server.go:508` calls `s.limited.ServeHTTP(w, r)` for every request, so the SPA document and static assets consume the read bucket. In the isolated smoke test, `page.reload()` received `429 {"error":"trop de requêtes — réessayez plus tard"}` instead of the HTML, producing an empty document title. `docs/MUSTREAD` and `api/boundary_test.go` define the limiter as an API boundary; serving a JSON error for a page refresh is a real user-facing defect.

**Implement:** in `Server.ServeHTTP`, route `/api/` requests through `s.limited`, and all other paths directly through `s.mux`. Keep the edge auth check and ownership guard exactly as they are. Do not change any rate value, bucket, retry logic, or error body.

**Test first (red):** add a Go test in `api/` that (a) performs more than the read burst of GETs to a non-API route and proves none returns 429, and (b) proves the 61st authenticated `/api/health` read still returns 429 exactly as before. Use the repo's existing test helpers and exact declared limits (60/min, read burst 20). Do not relax an existing assertion.

**Done when:** the new test is red before the edit and green after; `go test -count=1 ./...` in `api/` is green; all six gates are green. Commit explicit paths.

## Task 2 — Retarget `observatory2.spec.ts` to the live two-mode UI

**Failure:** `e2e/observatory2.spec.ts:9` waits for `getByLabel('Navigation').getByRole('button', {name:'Campagne'})`, a navigation destination deleted by the ratified two-mode redesign (`Synthèse` / `Labo` only). It times out against the real app; the app is correct.

**Implement:** rewrite the flow to preserve its intent (command palette opens/closes; the audit form is reachable; the site field exists; the optional live sparkline must not crash) using only live selectors: `[data-mode="cockpit"]` / `[data-mode="labo"]`, `#v-cockpit`, `.labo-view`, and the actual Audit-du-lien form control. Keep the existing palette assertions. No `page.route`, no invented state, no new assertion that something is visible when the spec itself says it may be absent.

**Done when:** the spec passes against the real Go server with Edge; the static E2E contract oracle `src/lib/e2eContract.test.ts` stays green; typecheck/frontend suite/six gates green. Commit explicit paths.

## Task 3 — Retarget the stale mobile case in `multi-depth.spec.ts`

**Failure:** `e2e/multi-depth.spec.ts:168-172` navigates to a hard-coded `https://localhost:9090/` (no server there) and clicks `[data-panel="live"]` (removed). Every other test in that file passed against the real harness.

**Implement:** change only that test to use Playwright's `baseURL` (`page.goto('/')`) and the live two-mode selectors; keep the same assertion intent (390px mobile, single-column layout, no horizontal overflow). Do not rewrite the whole file, do not weaken the overflow bound, do not introduce a hard-coded host.

**Done when:** that test passes against the real Go server with Edge; typecheck/frontend suite/six gates green. Commit explicit paths.

## Task 4 — Full real E2E pass and inline demo captures

**Implement:** after Tasks 1–3, run the full E2E suite once against the freshly built dist + Go binary with Edge and a generated `CGO_API_TOKEN`, `--timeout=90000`, unmocked. Expect all specs green. Then capture the real running app with Playwright at widths 890, 1280 and 1920, in both `Synthèse` and `Labo`, into a temp directory; verify each PNG exists and is non-trivial. Report exact commands, exit codes, counts and screenshot paths. Do not commit screenshots to the repo.

**Done when:** full E2E exits 0 with count of passes stated; six captures exist; six gates green. No code commit needed unless a real defect appears (then stop and report it).
