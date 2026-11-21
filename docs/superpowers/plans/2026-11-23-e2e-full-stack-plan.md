# Full-stack E2E plan (2026-11-23)

Authorized boundary (human): **local real stack only** — the real compiled Go binary on 127.0.0.1, real API/SSE, no sadcnwp, no bench hardware, no frozen-corpus writes, no `/api/**` stubbing in Playwright. The local binary runs in observe mode: mutation routes return 501 by design and `/api/stream` emits real idle frames at 20 Hz. Tests must assert that honesty, never fabricate a campaign, PASS gate, seal, or measured value.

Global rules: `PWCHROMIUM` Edge, `--timeout=90000`, `CGO_API_TOKEN` header (auth stays enabled), `reuseExistingServer:false`, rebuild the Go binary after any `dist` change, kill stale servers by PID/port. Evidence runs serialized. One commit per task with explicit paths; never `git add -A`; no ledger/gate/frozen data edits; no merge. Zero console errors on every flow.

### Task 1: API contract and auth E2E

New spec `e2e/fullstack-api.spec.ts`: against the real binary assert `/api/health` contract; unauthenticated `/api/state` → 401 while Bearer → 200; login cookie path; read endpoints `/api/state`, `/api/events`, `/api/integrity`, `/api/results`, `/api/quarantine`, `/api/capabilities`, `/api/profiles`, `/api/cost/tiers`, `/api/figures/`, `/api/replay/stream` return valid JSON/honest absence and no fabricated data; observe-mode mutation routes (`/api/audit/start|stop`, `/api/run/start|stop`, `/api/shape`, `/api/watch`, `/api/audit/toprofile`) return 501 with no side effects. No page.route. Verification: spec passes; tsc/build; one commit.

### Task 2: Real-time SSE and UI live-state E2E

New spec `e2e/fullstack-realtime.spec.ts`: open real `/api/stream` from the browser, assert event-stream frames arrive with `running:false` idle contract (no measured values presented as zero), assert 20 Hz cadence envelope on raw `request` stream, assert the UI reflects connection/live state and that an ARM click produces the honest 501 refusal word (no fabricated "en cours"). Cover reconnect/backpressure status without stubbing. Verification: spec passes; full gates.

### Task 3: Whole-interface interaction E2E

New spec `e2e/fullstack-interface.spec.ts`: one unmocked walk through both modes — mode switch + reload persistence, Synthèse compact/disclosure at 890 and spine geometry at 1280/1920, Labo rail keyboard (arrows/Home/End + Enter) and all four panels, inline disclosures and ProofDrawer focus trap/Esc/focus return, 44px targets, no page overflow, axe WCAG 2.2 AA at 890/1280/1920, forced-colors and reduced-motion. Verification: spec + existing axe/layout/feel specs all green.

### Task 4: Cross-surface workflow E2E

New spec `e2e/fullstack-workflow.spec.ts`: real-data Labo → Comparer → Preuves workflow — unmeasured-last ranking, empty-duel refusal, corpus line from real `/api/integrity`, seal inspector real verify act, archive `OUVRIR` loads the clicked row's dossier via the real `/api/audit/report` endpoint (honest failure when absent), MD/CITER row-deep-linked, toasts with honest copy. No fabricated seals or counts. Verification: spec + full gates.

### Task 5: Evidence aggregation, gates and final review

Run the complete Playwright suite serially plus tsc/build/vitest and native `sh .githooks/run-gates.sh` (record the pre-existing `links` false positive, never edit ledgers); forbidden diff check; capture screenshots at 890/1280/1920 and validate; final whole-branch reviewer over the E2E diff, one fix wave + scoped re-review if needed.
