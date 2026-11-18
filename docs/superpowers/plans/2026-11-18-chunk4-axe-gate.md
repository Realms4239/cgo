# CGO / Meteolink Chunk 4 — Live Axe WCAG Gate Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an exact-pinned `@axe-core/playwright` and a real, unmocked Playwright spec that scans the served SPA at 390 px and 1920 px and fails on any critical/serious WCAG 2.2 AA violation.

**Architecture:** `playwright.config.ts` keeps its real-`cgo` webServer but adds a test token so the now-fail-closed auth permits the harness. `e2e/axe-live.spec.ts` uses `AxeBuilder` against that server with **no `page.route` / no fetch rewrites**, the same principle as the existing `unmocked.spec.ts`.

**Tech Stack:** TypeScript, Playwright, `@axe-core/playwright`, Go binary served by the existing webServer.

**Source:** `docs/MUSTREAD/14-chunk-4-honest-tests.md` §3 (verbatim corpus `CGO5.md` 63–75). Chunk 4 is decomposed; most of §1–2 already exists; this is the remaining axe slice.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- The user explicitly approved this ONE new dependency, exact-pinned. No other dependency may be added.
- Never touch `book/`, frozen corpus/schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, live bench, remote hosts.
- No `page.route`, no mocked `/api/**` in the new spec — it must hit the real served binary.
- The 14 pre-existing `^` carets elsewhere in `package.json` are out of scope for this slice; record them as a follow-up, do not silently rewrite them.
- French comments/user text; English spec names.
- One commit, exact message: `test(e2e): add unmocked axe wcag 2.2 aa gate`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Exact-pinned dependency

**Files:**
- Modify: `web/frontend/package.json`, `web/frontend/bun.lock`

**Interfaces:** consumes existing bun project; produces an exact (no `^`) `@axe-core/playwright` devDependency and matching lockfile entry.

- [ ] **Step 1: Install**

Run from `web/frontend`: `bun add -D --exact @axe-core/playwright`.
Expected: package.json gains `"@axe-core/playwright": "<exact version>"` with no caret; bun.lock records the same version.

- [ ] **Step 2: Verify** — `grep -n '"@axe-core/playwright"' package.json` → one hit with no `^`; `grep -c axe-core bun.lock` ≥1; `bun install --frozen-lockfile` → exit 0 (lock consistent). If the registry is unreachable, STOP and report BLOCKED with the exact error — do not hand-edit the lockfile.

---

### Task 2: Unmocked axe spec + test token in webServer

**Files:**
- Modify: `web/frontend/playwright.config.ts`
- Create: `web/frontend/e2e/axe-live.spec.ts`

**Interfaces:** consumes the webServer-spawned real `cgo` binary at `127.0.0.1:4173`.

- [ ] **Step 1: Give the harness a token**

In `playwright.config.ts`, add an `env` block to `webServer` supplying `CGO_API_TOKEN: 'cgo-e2e-axe-test-token-32-chars-minimum'` and `CGO_DATA_DIR` set to a temp path under `tmp/` (the server must not write into tracked data). Keep the existing command/cwd/url/timeout untouched. Explain in a French comment that fail-closed auth requires a token.

- [ ] **Step 2: Write the spec**

Create `e2e/axe-live.spec.ts`:

```ts
// UNMOCKED — aucun page.route : le vrai binaire cgo sert la SPA et /api/*.
import AxeBuilder from '@axe-core/playwright'
import { expect, test } from '@playwright/test'

for (const width of [390, 1920]) {
  test(`axe wcag 2.2 aa @ ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 844 })
    await page.goto('/')
    await expect(page.locator('body')).toBeVisible()
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'])
      .analyze()
    const serious = results.violations.filter((v) => v.impact === 'critical' || v.impact === 'serious')
    expect(serious.map((v) => `${v.id}: ${v.help}`)).toEqual([])
  })
}
```

No mocks, no `--dev-no-auth`, no screenshot-only assertions.

- [ ] **Step 3: Verify** — `cd web/frontend && npx playwright test e2e/axe-live.spec.ts` → exit 0 if browser binaries and the server are available. If browsers are missing, report BLOCKED with the exact Playwright error; do not install browsers without a separate approval.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `npx vitest run` → 0; `./node_modules/.bin/tsc --noEmit` → 0; the axe spec result from Task 2 (or BLOCKED status).
- [ ] **Step 2: Static proof** — `grep -n 'page.route' web/frontend/e2e/axe-live.spec.ts` → EMPTY; `git grep -n 'AxeBuilder' web/frontend/e2e/` ≥1; `git diff --name-only` contains only the four named files.
- [ ] **Step 3: Commit once** — stage `web/frontend/package.json`, `web/frontend/bun.lock`, `web/frontend/playwright.config.ts`, `web/frontend/e2e/axe-live.spec.ts` by explicit path; message `test(e2e): add unmocked axe wcag 2.2 aa gate`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes and limitations to the SDD report. Record the 14 pre-existing carets as a follow-up. No `docs/CHECKPOINTS.md` append.
