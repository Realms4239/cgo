# Chunk 4 — Blazing-Fast Honest Tests

> Source: `C:\END\CGO5.md` lines 90–167. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** real synctest on product state machines, node + happy-dom split, unmocked Playwright E2E and Axe WCAG gate.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** a fast suite that cannot pass on mocks of the module under test.
- **Usage:** execute after Chunk 3; gates Chunk 5.
- **Inherited constraints:** `01-core-directive.md` §2; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 5 — container, strict CI and signed release.

Hand Part A to your agent, then run Part B yourself to confirm speed and
truthfulness before moving to Chunk 5 (Container, Strict CI & Signed Release).

CHUNK 4 — Blazing-Fast Honest Tests (synctest, happy-dom, Live E2E + Axe)

PART A — Exact Bounded Prompt for Your Coding Agent

# CHUNK 4 EXECUTION TASK — FAST, HONEST, UNMOCKED VERIFICATION SUITE (STRICTLY BOUNDED)

You are executing ONLY Chunk 4 on `C:\cgo`. Do NOT modify production business logic (`pkg/probe/`, `pkg/qdisc/`, `api/auth.go`, `api/ssrf.go`, `Dockerfile`, CI workflows) unless fixing a genuine bug uncovered by these tests.

## 1. Real Go `testing/synctest` Tests (Replace Clock Toys With Real Product Coverage)
Files: `pkg/audit/audit_synctest_test.go` (new), `api/sse_synctest_test.go` (new), plus speeding up existing slow `time.Sleep` tests
1. **`pkg/audit/audit_synctest_test.go`:**
   - Inside `synctest.Test(t, func(t *testing.T) { ... })`, test the real `60 s` (`10s` idle / `40s` load / `10s` drain) phase state machine, `Duration+30s` deadline cancellation, and failed-load `HasValidRemediation == false` using fast in-memory mock probe seams. Must finish in `< 100 ms` of wall clock time!
2. **`api/sse_synctest_test.go`:**
   - Inside `synctest.Test(t, func(t *testing.T) { ... })`, test the real `api/sse.go` hub:
     - `20 Hz` (`50 ms`) ticker emission without wall-clock `time.Sleep` flakes,
     - `4096`-slot O(1) circular ring wrap-around + `Last-Event-ID` deterministic replay,
     - `maxSubs = 64` exact boundary (`64` accepted, `65th` receives `503`),
     - Backpressure behavior: slow reader drops stale `tick` frames while retaining guaranteed milestone `event` frames!
3. **Speed Up Existing Slow Go Tests:**
   - Audit `pkg/audit/*_test.go` and `api/*_test.go` for slow real-time `time.Sleep` waits; convert pure unit timing tests to `testing/synctest` (keeping genuine kernel/integration sleeps strictly under `//go:build real`) so `go test -race -count=1 ./...` completes in `< 10 s`.

## 2. Speed Up Frontend Tests (`jsdom` → `node` + `happy-dom`) & Add Real Coverage
Files: `web/frontend/package.json`, `web/frontend/bun.lock`, `web/frontend/vitest.config.ts`, `web/frontend/src/**/*.test.ts*`
1. Install exact-pinned `happy-dom` and `@axe-core/playwright` (`bun add -D --exact happy-dom @axe-core/playwright`). Verify ZERO `^` carets in `web/frontend/package.json`.
2. Update `web/frontend/vitest.config.ts`:
   - Set default `environment: 'node'` so all pure TypeScript tests (`ringBuffer.test.ts`, `tokens.test.ts`, formatters, stores) execute without DOM boot overhead.
   - Route DOM component tests (`**/*.dom.test.tsx`) to `environment: 'happy-dom'` via `environmentMatchGlobs` (or per-file `@vitest-environment happy-dom`).
   - Add/verify tests for:
     - `TypedMetricRing` & `DomainEventRing` (`NaN` on missing data, wrap-around, `dirty` flag),
     - `ArchiveTable` pluralization (`0 entrée`, `1 entrée`, `2 entrées`),
     - `tokens.test.ts` verifying `>= 4.5:1` WCAG AA contrast on ALL semantic fg/bg pairs (including `--t-danger-text`),
     - Populated `lexicon.test.ts` over loaded/degraded/error fixture states (zero `p95`, `run`, `gel`, `Manifestes`, or English leaks in operator views).

## 3. Fix `K12` — Real Unmocked Playwright E2E + `@axe-core/playwright` WCAG 2.2 AA Gate
Files: `web/frontend/playwright.config.ts`, `web/frontend/e2e/live-smoke.spec.ts` (new), `web/frontend/e2e/accessibility.spec.ts`, `web/frontend/tsconfig.json`
1. Rename or clearly label existing fixture-mocked tests (e.g., `accessibility.mock.spec.ts` or keep isolated) while **DO NOT** claim mocked specs prove backend integration.
2. Create `web/frontend/e2e/live-smoke.spec.ts`:
   - Uses `playwright.config.ts` `webServer` (or helper) launching the **real compiled `cgo` Go binary** (`--serve --addr 127.0.0.1:19091 --tls=false --insecure-allow-plaintext`) with an explicit test `CGO_API_TOKEN` and temp `CGO_DATA_DIR` (NO `--dev-no-auth`, NO `page.route('/api/**')` mocks!).
   - Authenticates through the real UI login / `/api/auth/login` flow.
   - Verifies `/api/health`, switches cleanly between `Cockpit` and `Labo`, checks `/api/integrity` real manifest status, verifies live SSE `RingBufferHUD` frame progression, triggers a confirmation-guarded (`ArmButton` + `X-Confirm-Action`) action, and persists state cleanly across reload.
   - Runs **`AxeBuilder` (`@axe-core/playwright`)** checking `wcag2a`, `wcag2aa`, `wcag21aa`, `wcag22aa` across both `Cockpit` and `Labo` at `390x844` and `1920x1080`: asserts `0` critical/serious violations, `0` horizontal overflow, `0` truncated ellipsis on `P1–P4` tiles, and `>= 44x44 px` bounding boxes on main interactive buttons/inputs.
3. Ensure `npx tsc --noEmit` covers `e2e/` and `scripts/*.mjs` cleanly.

## Constraints
- Paste exact execution timings from `go test -race -count=1 ./...`, `bun run test`, and `npx playwright test` in your final report—no estimated timings!
- Commit all Chunk 4 work in ONE clean commit: `test(suite): add real synctest coverage, happy-dom speedup and unmocked playwright+axe e2e`.

PART B — Your Hard Verification Gate (Run Before Chunk 5)

Run these commands locally after the agent finishes Chunk 4:

# 1. Verify Go test suite is fast AND uses synctest on real audit/SSE packages
go test -race -count=1 ./...
# Check wall-clock time: should now be seconds, NOT 221s!

# 2. Verify frontend tests run fast with happy-dom + node and ZERO ^ carets remain
cd web/frontend
grep -E '"\^' package.json || echo "CLEAN_EXACT_PINS"
npx tsc --noEmit
bun run test
# Check wall-clock time: should be ~2–8s, NOT 154–218s!

# 3. Run Playwright (including the unmocked live-smoke + axe-core WCAG scan)
bun run build
npx playwright test
cd ../..

# 4. Verify live-smoke.spec.ts does NOT mock /api/**
git grep "page.route" web/frontend/e2e/live-smoke.spec.ts || echo "CLEAN_UNMOCKED_E2E"

Once Go tests, Vitest, TypeScript, and unmocked Playwright + Axe all pass
cleanly and fast, say "Proceed with Chunk 5" and we will lock down the 3-stage
Dockerfile (K3), --strict CI (K5), pinned signed Release pipeline (K9),
NTP/Cert/Logging/Backup drills (K10, K11)! Proceeding to Chunk 5 (3-Stage Clean
