# END-Corpus Interface Capabilities Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make BBRv3, `.dist`, direction, field-trace and eBPF capabilities reachable from the live Labo, driven only by real backend sources with honest absence where no source exists.

**Architecture:** Four sequential SDD slices: (1) backend schema + read-only calibrations endpoint + dynamic CC/BBRv3, (2) Labo capability strip + Canal/Direction/Distribution controls, (3) field-trace/eBPF provenance strip, (4) feel-better pass and final review. Each slice produces its own working, testable change and one explicit-path commit.

**Tech Stack:** Go (`api`, `capabilities`, `model`), React/TypeScript (`web/frontend`), Vitest + happy-dom, Playwright on Edge, `sh .githooks/run-gates.sh`.

## Global Constraints

- Corpus `C:/END` is read-only, never modified. Functional controls select and persist configuration only — they never claim a measurement until a real run produces one.
- Absence stays `—` + a French absence word. No fabricated control, value, seal, campaign or green. No default that claims a measurement.
- No new dependency, hex, font-size, accent, `transition: all` or `@keyframes`. Use existing tokens. 44px hit areas, WCAG 2.2 AA, forced-colors, reduced-motion, no overflow at 890/1280/1920.
- Protected `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, frozen corpus never edited. No `git add -A`; explicit paths only. No merge.
- Native gates once per slice; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.
- Bench execution, sadcnwp, thesis, tag/freeze stay human-gated.

---

### Task 1: Backend source of truth

**Files:**
- Modify: `api/server.go` (schema param list, capabilities-driven CC, new endpoint registration)
- Create: `api/calibrations.go`, `api/calibrations_test.go`

**Interfaces:**
- Consumes: `capabilities.CapabilitiesNow()` → `{qdiscs, ccs, known_ccs, kernel, bbr3}`; `model.Event.Direction ∈ {"up","down"}`; calibration files with SHA-256 sidecars under the real calibration directory.
- Produces: schema keys `link_channel ∈ {"statique","empirique"}`, `link_direction ∈ {"up","down"}`, `dist_name` (dynamic enum), `bbr3_state` (honest string); `GET /api/calibrations` → `{calibrations: [{name, sha256, samples, p50_ms, verified}]}`.

- [ ] **Step 1: Write failing schema test**

In `api/schema_contract_test.go`, assert `paramByKey("link_channel")` exists with enum `["statique","empirique"]`, `link_direction` enum `["up","down"]`, and that `burst_cc` contains `bbr3` only when the injected capabilities stub reports `BBRv3: true`. Run `go -C api test -count=1 -run TestSchemaEndCorpusParams .` → FAIL (keys absent).

- [ ] **Step 2: Implement schema params and dynamic CC**

Add the three params to the registry in `api/server.go` beside `shape_qdisc`/`burst_cc`. Replace the static `burst_cc` enum with a function that reads `capabilities.CapabilitiesNow()`; when `BBRv3` is true append `bbr3`, else expose `["cubic","bbr"]` and set `bbr3_state` to the corpus sentence: `BBRv3 non détecté — CUBIC + BBRv1 sur noyau 6.8, détection bbr3 prête`. Run the same test → PASS.

- [ ] **Step 3: Write failing calibrations endpoint test**

In `api/calibrations_test.go`, create a temp calibration dir with one valid `.dist` + sidecar (`sha256`, `samples=35`, `p50_ms=110`) and one invalid file; assert `GET /api/calibrations` returns only the verified one with exact fields, and that an empty dir returns `{"calibrations":[]}` with 200. Run `go -C api test -count=1 -run TestCalibrations .` → FAIL (endpoint absent).

- [ ] **Step 4: Implement the read-only endpoint**

Create `api/calibrations.go` with `ListCalibrations(dir string) []Calibration` that reads sidecars, rejects malformed/unverified entries, and never synthesizes an entry. Register `GET /api/calibrations` in `api/server.go` returning `writeJSON(w, map[string]any{"calibrations": ListCalibrations(calDir)})`. Run the test → PASS.

- [ ] **Step 5: Full module suite and commit**

Run `go -C api test -count=1 ./...` and `go -C capabilities test -count=1 ./...` → exit 0; `go build ./...` → exit 0; `sh .githooks/run-gates.sh --only tests` → exit 0. Commit explicit paths only: `git add api/server.go api/calibrations.go api/calibrations_test.go api/schema_contract_test.go && git commit -m "feat(api): real channel/direction/calibration sources"`. Use the documented no-hooks override if a pre-existing gate blocks, and record why.

---

### Task 2: Labo capability and configuration controls

**Files:**
- Modify: `web/frontend/src/components/labo/Configurer.tsx`
- Modify: `web/frontend/src/views/LaboView.tsx`
- Modify: `web/frontend/src/styles/components.css`
- Create: `web/frontend/src/components/labo/Configurer.capabilities.dom.test.tsx`

**Interfaces:**
- Consumes: `/api/schema` params from Task 1 (`link_channel`, `link_direction`, `dist_name`, `bbr3_state`), `/api/calibrations`, `useCapabilities()` and the existing `BenchCapabilities` component.
- Produces: visible, honest controls bound to the real params; no new exported API.

- [ ] **Step 1: Write the failing DOM oracle**

Create `Configurer.capabilities.dom.test.tsx`. Assert: (a) capability strip renders qdiscs, ccs, kernel and the honest BBRv3 word from a real `/api/capabilities` fixture; (b) Canal renders `Statique` / `Empirique .dist`; (c) Direction renders `Egress WAN` / `Ingress IFB` / `Les deux` only when the schema enum contains those values — otherwise the absent word; (d) Distribution lists only sidecar-verified calibrations, with `— aucune calibration vérifiée` when empty; (e) no digit, `PASS`, `VÉRIFIÉ` or fabricated value in any absence state. Run `npx vitest run Configurer.capabilities.dom.test.tsx` → RED.

- [ ] **Step 2: Mount the orphaned capability strip**

In `LaboView.tsx` or `Configurer.tsx` (choose the one that keeps chapter lede order), render `<BenchCapabilities caps={caps} />` from the existing `useCapabilities()` hook. Do not duplicate its markup; do not change its honesty wording.

- [ ] **Step 3: Render controls from the real schema**

In `Configurer.tsx`, read the schema params and calibrations through the existing fetch hooks (add a small `useCalibrations` hook only if none exists). Render Canal and Direction as 44px option buttons with `aria-pressed`, and Distribution as a select whose options come only from verified calibrations. When a source is empty, render `—` + the French absence word and disable that control. No placeholder option.

- [ ] **Step 4: Style with existing tokens**

In `components.css`, use `.configurer-*` classes with `var(--sp-*)`, `tabular-nums`, `text-wrap: pretty`, calm hairlines and concentric radii where nested. No new hex, font-size, keyframes or `transition: all`.

- [ ] **Step 5: Verify and commit**

Run `npx vitest run Configurer.capabilities.dom.test.tsx` → GREEN; full `npm test` → green; `npm run build` → exit 0 and restore tracked `dist/index.html`; rebuild `go build -o web/frontend/cgo-test-server.exe ./cmd/cgo`; unmocked Edge Playwright `redesign-smoke`, `redesign-layout`, `redesign-a11y`, `redesign-feel` → green; screenshots at 890/1280/1920 read and validated. Commit explicit paths only with message `feat(labo): real capability and configuration controls`.

---

### Task 3: Field-trace and eBPF provenance strip

**Files:**
- Create: `web/frontend/src/components/labo/ProvenanceStrip.tsx`, `ProvenanceStrip.dom.test.tsx`
- Modify: `web/frontend/src/components/labo/Preuves.tsx`, `web/frontend/src/styles/components.css`

**Interfaces:**
- Consumes: `/api/integrity` → `{integrity, ok, hash8, ...}`; `/api/doctor` → `{mode, checks}`; the `--confront-out` JSON shape from `cmd/cgo/confront.go` when present.
- Produces: three honest lines — field trace, eBPF confrontation, integrity — each real or `—` + word.

- [ ] **Step 1: Write the failing oracle**

`ProvenanceStrip.dom.test.tsx`: with empty responses, assert `— trace de terrain non mesurée`, `— confrontation eBPF absente`, and integrity rendered only from the real `integrity` value; assert no digit/green/seal word in absent states. With a real integrity fixture, assert the exact `hash8` and status word. Run → RED.

- [ ] **Step 2: Implement the strip**

`ProvenanceStrip.tsx` reads real data through existing hooks; each line names its source and renders absence when the source is missing or degraded. No inference from observe-mode frames. Insert below the Preuves chapter content without duplicating `ProofDrawer`.

- [ ] **Step 3: Verify and commit**

Targeted DOM → GREEN; full suite; build + Go rebuild; Playwright + screenshots; gates once. Commit explicit paths with `feat(labo): honest field-trace and eBPF provenance`.

---

### Task 4: Feel-better pass and final programme review

**Files:** `web/frontend/src/styles/components.css` (polish only)

**Interfaces:** consumes all three slices; produces the final visual/a11y verdict.

- [ ] **Step 1: Check the feel-better checklist**

Verify: concentric radii, optical alignment of new controls, tabular numerals, `text-wrap: pretty/balance`, explicit transitions (no `all`), 44px hit areas, no new motion. Fix only concrete failures found.

- [ ] **Step 2: Full evidence run**

Run build, full vitest, unmocked Edge Playwright (all green-bar specs), 890/1280/1920 screenshots, native gates once. Record `links` FAIL and `book/` SKIPs without editing.

- [ ] **Step 3: Final review**

Dispatch a whole-programme reviewer over the combined diff with the parked list. One fix wave + scoped re-review if Critical/Important findings appear. Append final state to the SDD ledger; no merge beyond the local fast-forward.

## Self-Review

- Spec coverage: §3 → Task 1; §4 → Task 2; §5 → Task 3; §6 → Task 4; §7 respected in every Global Constraints block.
- No placeholders: every step names a command, a file or exact copy. No `TBD`, "similar to Task N", or unspecified handling.
- Type consistency: `Calibration {name, sha256, samples, p50_ms, verified}`, schema keys and `Capabilities` field names are used identically across tasks.
