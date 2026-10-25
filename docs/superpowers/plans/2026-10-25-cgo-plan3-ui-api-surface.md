# CGO Enterprise Rebuild — Plan 3: UI / API Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver spec §8's API + UI slice: run provenance (`bench`, `channel_model`) through `/api/run/start` with fail-fast validation and the decision-6 `flows=4` default; `GET /api/cost/enterprise` catalog; bench-aware audit sink detection shared CLI/API (closes re-review M4/M5); the `both`-mute report completeness fix (M6); and the dashboard surfaces — live enterprise cost/JFI, campaign channel-model cockpit, simplified audit card, enterprise result card.

**Architecture:** Go API first (`pkg/api`, `pkg/audit`, `cmd/cgo` wiring), then the React frontend consumes it (`web/frontend/src`). Sink detection moves from `cmd/cgo` (package main) to `pkg/audit` so CLI and API share one ladder; the API server gains a `SinksFn` seam (nil ⇒ real detection) so handler tests stay offline. Frontend touches `types.ts` (LiveFrame mirror), `LiveView`, `CampagneView` (cockpit + audit card + result card), one new lib module for the P2/P3 canal gate. Existing design tokens/layout only (`--surface-soft`, `--hairline`, `.btn`, `.card-head`, portal modal, LiveGuide collapse pattern, LiveView cubic/bbr button-group) — the impeccable skill governs the final UI quality pass.

**Tech Stack:** Go 1.25 stdlib only; React 19 + zustand + vitest (jsdom) existing setup; `npm run typecheck` + `npm run test` + `npm run build` from `web/frontend`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` (§2 honesty invariants, §8 UI/API slice, decisions 6–8).
- Go 1.25 stdlib only, **no new dependencies** (frontend deps frozen too); comments French, identifiers English.
- From repo root `C:\cgo`: `go build ./...`, `go vet ./...`, `go test ./pkg/api/... ./pkg/audit/... ./cmd/cgo/... -count=1` green on Windows; from `web/frontend`: `npm run typecheck`, `npm run test`, `npm run build` green.
- Frozen-data guard: `book/data/runs`, `pkg/results/runs-frozen.tar.gz` untouched; frozen CSV schemas immutable; `book/tools/audit-archive.py` output byte-identical after every task.
- Never fabricate: unmeasured ⇒ stated (badge « non mesuré », `Unavailable`, empty), never a zero that reads as measured; scenario ≠ claim (labels carry « scénario contractuel »).
- Bursty is P2/P3-only (decision 7): enforced twice — API 400 before any mutation, campaign validation as last word.
- Commit after every task with the given message.

---

### Task 1: `/api/run/start` — provenance `bench` + `channel_model`, défaut `flows=4`

**Files:**
- Modify: `pkg/api/server.go` (`RunOpts`, run/start validation), `cmd/cgo/server.go` (`startFn`), `cmd/cgo/runcli.go` (précédence `CGO_CHANNEL_MODEL`)
- Modify: `pkg/api/run_start_filters_test.go` (new cases)

**Interfaces:**
- `RunOpts` gains `Bench string \`json:"bench"\`` and `ChannelModel string \`json:"channel_model"\``.
- Validation (before the engine-wired check, so it holds on observe-only hosts): `bench` ∈ {"", "legacy", "dualns"} else 400 `bench must be legacy|dualns`; `channel_model` ∈ {"", "clean", "bursty"} else 400 naming the rule `clean|bursty`; `channel_model=="bursty"` ⇒ every entry of `opts.Profiles` ∈ {P2, P3} else 400 naming the offending profile(s) + the P2/P3 rule (mirrors campagne's `RunEvent` message).
- `startFn` (cmd/cgo/server.go): `o.Flows == 0 ⇒ 4` (decision 6 — today's `if <1 {=1}` lands on 1); bench resolution mirrors `runCLI`: `o.Bench` → else `CGO_BENCH` env → else `"legacy"`; `"dualns"` + direction ⇒ `ProdDepsDualNS(o.Direction)` (up/down aware — it already takes direction) instead of `ProdDeps()/ProdDepsDown()`; `deps.ChannelModel = o.ChannelModel` ("" stays "" — campagne defaults to clean).
- `runcli.go`: `--channel-model` flag default becomes `envDefault("CGO_CHANNEL_MODEL", "clean")` so the env isn't dead (re-review P2 minor); same precedence audit as `--bench` (flag explicit value must still win over env — flag parsing gives that only if the default absorbs the env, exactly the `--bench` pattern; verify `runCLI`'s bench handling and copy it).

- [ ] **Step 1: Failing tests** (run_start_filters_test.go): `channel_model=bursty` + profiles [P1] ⇒ 400, body names `P1`; `channel_model=tlmn` ⇒ 400; `bench=xyz` ⇒ 400; `channel_model=bursty` + [P2] passes validation ⇒ 503 `run engine not wired` (proves the gate order); `flows` absent (0) no longer 400.
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/api/ ./cmd/cgo/ -count=1`; audit-archive byte-identical.
- [ ] **Step 5: Commit** — `api: run/start accepte bench + channel_model (bursty gate P2/P3), flows=4 par défaut`

---

### Task 2: `GET /api/cost/enterprise` — catalogue plans/tarifs/provenance

**Files:**
- Modify: `pkg/api/server.go` (endpoint)
- Create: `pkg/api/enterprise_cost_test.go`

**Interfaces:** `GET /api/cost/enterprise` → 200 JSON:
```json
{ "plans": [ {"name":"enterprise-gold-cir","monthly_opex_mga":1000000,"cir_mbps":200},
             {"name":"enterprise-platine1","monthly_opex_mga":2500000,"fup_bytes":32985348833280} ],
  "tariffs": [ {"name":"b2b-platine-bulk","ar_per_gb":81.38,"quoted":false},
               {"name":"b2b-m2m-apn","ar_per_gb":150,"quoted":true} ],
  "emergency_ar_per_gb": 5556,
  "primary_mapping": {"fiber":"enterprise-gold-cir","other":"enterprise-gold-cir","4g":"b2b-m2m-apn","5g":"b2b-m2m-apn","vsat":"b2b-platine-bulk"},
  "disclaimer": "<verbatim metrics.BuildEconomics disclaimer>" }
```
Values serialize FROM the `pkg/metrics` vars (PlanGOLD/PlanPLATINE/TariffWholesale/TariffCorporateAPN/RetailEmergencyARPerGB) and `primaryName` mapping — the endpoint never hardcodes a number the package owns (single source; a tariff change updates both surfaces). The `quoted` flag on APN keeps the « devis opérateur » honesty (invariant 4).

- [ ] **Step 1: Failing test** — GET parses; both plan names present; APN entry `quoted:true`, wholesale `quoted:false`; emergency 5556; disclaimer verbatim from `metrics.BuildEconomics(...).Disclaimer`; primary_mapping has all five keys.
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/api/ -count=1`.
- [ ] **Step 5: Commit** — `api: GET /api/cost/enterprise — plans, tarifs, provenance, mapping primaire`

---

### Task 3: audit/start — puits décidés par le banc (ResolveSinks partagé) + branches 400 testées

**Files:**
- Create: `pkg/audit/sinks.go` (moved from `cmd/cgo/auditsinks.go`), `pkg/audit/sinks_test.go`
- Modify: `cmd/cgo/auditsinks.go` (thin alias, CLI unchanged), `pkg/api/server.go` (`Deps.SinksFn`, audit/start defaults, dead-branch removal), `pkg/api/audit_start_sinks_test.go` (create)

**Interfaces:**
- `pkg/audit` exports `Sinks{Mode, Target, SmallURL, BulkAddr string}`, `ResolveSinks() Sinks` (same ladder: dualns 10.30.0.2 trio → legacy 10.200.0.1 trio → external), `EnvOrBulk(s Sinks) string` (CGO_BULK_ADDR override). `cmd/cgo` keeps `ResolveSinks()`/`Sinks` wrappers delegating to the package — CLI call sites untouched.
- `api.Deps` gains `SinksFn func() audit.Sinks` — nil ⇒ `audit.ResolveSinks()`; handler tests inject a stub (offline, deterministic).
- audit/start replaces the hardcoded `10.200.0.1` defaults: `target` empty ⇒ `sinks.Target` (external keeps 8.8.8.8); `small_url` empty ⇒ `sinks.SmallURL` (**empty stays empty in external — jamais de charge implicite**); `bulk_addr` empty ⇒ `EnvOrBulk(sinks)`; env overrides keep today's precedence (`CGO_SMALL_URL`/`CGO_BULK_ADDR`). The `recordEvent("audit", "démarré…")` line names the detected bench mode so the operator sees which path runs.
- M5 cleanups: remove the dead `body.Flows < 1` arm (unreachable after `<=0 ⇒ 4`; keep `> 8`), and pin the 400 branches with tests: direction invalide, flows 9, cible > 64 caractères.

- [ ] **Step 1: Failing tests** — stub SinksFn returning the dualns trio: POST audit/start with empty target/small/bulk ⇒ the started audit params carry 10.30.0.2/8081/5201 (assert via a Params-recording seam or the audit/status payload — whichever the existing tests use); external stub ⇒ small_url stays "" and bulk only via env; 400 branches listed above.
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/audit/ ./pkg/api/ ./cmd/cgo/ -count=1`; audit-archive byte-identical.
- [ ] **Step 5: Commit** — `api/audit: puits détectés par le banc (ResolveSinks partagé), branches 400 testées`

---

### Task 4: rapport d'audit `both` — miroirs mesurés rendus, économie absente dite

**Files:**
- Modify: `pkg/audit/report.go`, `pkg/audit/report_test.go`

**Context:** re-review M6 (`.superpowers/sdd/2026-10-25-cgo-plan1-economics-audit/re-review-t6-t11.md`) — with `direction=both` and a mute up window (no loaded samples up), the Markdown report drops the measured down-mirrors, `Economics` can be absent without an `Unavailable` entry saying why, and the note « voir grade_down » points at a field the report never exposes. Under-reporting measured evidence + silent absence — both honesty-adjacent. Read the re-review entry first; it carries file:line specifics.

**Interfaces:**
- Markdown rendering: when a direction window is mute, the OTHER direction's measured mirrors (Δp95, drain, VoIP R as available) still render, labelled `down`/`up`.
- `Economics == nil` (load window empty) ⇒ an explicit `Unavailable` entry (« économie entreprise — fenêtre de charge vide ») instead of an absent section.
- The « voir grade_down » note: reword to reference what the report actually carries (the per-direction section), or expose the field — whichever is smaller; the note must not promise an absent field.

- [ ] **Step 1: Failing test** — a `both` fixture with mute up + measured down: Markdown contains the down Δp95/drain lines; `Unavailable` names the economics absence; no dangling « voir grade_down » reference.
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/audit/ -count=1`; audit-archive byte-identical (frozen reports untouched — rendering change applies to fresh reports only).
- [ ] **Step 5: Commit** — `audit: rapport both — miroirs down rendus, économie absente dite Unavailable`

---

### Task 5: frontend — LiveFrame porte le coût entreprise, le B2B et le JFI gelé

**Files:**
- Modify: `web/frontend/src/lib/types.ts` (`LiveFrame`), `web/frontend/src/views/LiveView.tsx` (bento strip), tests colocated (`src/lib/`/`src/views/` per existing pattern)

**Interfaces:** `LiveFrame` gains `cost_primary_ar_per_h: number`, `cost_b2b_ar_per_h: number`, `jfi_pct: number` (mirror of Go tags — campagne.go Snapshot). Bento strip (LiveView:402-419) gains « Coût entreprise » (GOLD, `cost_primary_ar_per_h`) and « Coût B2B » (APN 150, `cost_b2b_ar_per_h`) cards via `formatCostAR`; the existing « Coût » card keeps its frozen retail meaning (label it « Coût (détail, gelé) » to prevent misreading — the two regimes side by side was the spec's point). `jfi_pct` shows when > 0 (final cell value; the live client-side JFI from rings stays as-is — different semantics, both honest).

- [ ] **Step 1: Failing tests** — type mirror test (LiveFrame keys vs Go json tags — extend the existing schema/type test if one exists, else a minimal keys test); bento strip renders the two new cards given a frame fixture (component-level if a precedent exists, else extract the strip's term list into a tested lib array).
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `npm run typecheck && npm run test` from `web/frontend`.
- [ ] **Step 5: Commit** — `ui: snapshot live porte coût entreprise + B2B + JFI gelé`

---

### Task 6: frontend — cockpit campagne : canal clean|bursty (P2/P3), flux 4 par défaut

**Files:**
- Create: `web/frontend/src/lib/canal.ts` (+ `canal.test.ts`)
- Modify: `web/frontend/src/views/CampagneView.tsx` (cockpit 310-354)

**Interfaces:** `canal.ts` exports `canalGate(profiles: string[]): { allowed: boolean, reason: string }` — bursty legal iff `profiles ∩ {P2, P3} ≠ ∅`; reason names the rule (« bursty : P2/P3 seulement — le GE y est calibré »). Cockpit: `flows` state default 1 → **4** (input keeps 1–8); a clean/bursty two-option button group (the LiveView:512-519 cubic/bbr pattern) between the flux row and the deadline row; bursty disabled + reason shown when the gate blocks (selection snaps back to clean); POST `/api/run/start` body gains `channel_model`. `bench` stays env-driven (out of the UI — spec asks only canal + flux; the API accepts it for outillage).

- [ ] **Step 1: Failing test** — `canal.test.ts`: gate on [P2], [P2,P3], [P1,P4] (allowed/allowed/blocked+reason).
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `npm run typecheck && npm run test`; manual smoke of the POST payload via the existing fetch seam if testable.
- [ ] **Step 5: Commit** — `ui: cockpit campagne — sélecteur canal clean|bursty (P2/P3), flux 4 par défaut`

---

### Task 7: frontend — carte audit simplifiée + carte résultat entreprise

**Files:**
- Modify: `web/frontend/src/views/CampagneView.tsx` (audit card 297-498), tests per pattern

**Interfaces:**
- **Form:** visible surface = Site, Type de lien (select), Durée segmentée (20/60 — two `.btn` buttons, cubic/bbr pattern; advanced free-text duration moves into the collapsed block), Sens (up/down/both three-option group — POST gains `direction`), AUDITER. Advanced (LiveGuide collapse pattern: `useState` + `localStorage('audit-advanced-open')`, replié par défaut): Fournisseur, Cible ping, Interface Linux, Interface RouterOS, durée libre. `validate()` extended: direction required-with-default (up), duration still 10–600 after free-text override.
- **Result card:** JFI (from report `jfi_pct`), CAP SLA badge — `Badge` tones: breach ⇒ danger + label SLA, pass ⇒ success, `cap_sla.measured=false` ⇒ steel « non mesuré » (NEVER green — invariant 2); primary + B2B Ar/h lines from `economics` (« scénario contractuel » suffix); seal truncated 12 chars, click-to-copy via `copyText`; grade colors migrate from inline hex to `LEVEL_COLOR` (quality alignment, no visual change).

- [ ] **Step 1: Failing tests** — extract the audit-form validation/normalization (duration segmented↔free-text, direction default) into a tested lib module if no component-test precedent exists; CAP SLA badge tone mapping as a pure function test (breach/pass/unmeasured → danger/success/steel).
- [ ] **Step 2: red → Step 3: implement → Step 4: green** — `npm run typecheck && npm run test`.
- [ ] **Step 5: Commit** — `ui: carte audit simplifiée (sens, durée segmentée, avancé replié), résultat entreprise`

---

### Task 8: docs — surface API, ligne de sécurité, matrice Windows + passe qualité

**Files:**
- Modify: `docs/api.md` (endpoints + defaults), `docs/deploy.md` (sécurité), `docs/usage.fr.md` (matrice Windows si c'est son domicile — sinon deploy.md)
- Quality pass over Tasks 5–7 UI (impeccable skill): tokens only (no new hex — `LEVEL_COLOR`/Badge tones), labels français cohérents, aria-pressed/expanded states, focus visible, empty-states honest (« non mesuré »).

**Content:** `/api/cost/enterprise` + run/start `bench`/`channel_model` + audit/start banc-aware defaults documented; security line: dashboard inside ns_client binds `10.99.0.2` (via `run-dualns.sh` / `CGO_DASHBOARD__ADDR`) — documented that anyone on the host can trigger shaping (lab/DSI context, spec §8); Windows capability matrix: audit works (ping+HTTP+bulk optionnel), TCP_INFO/netns/doctor-offload = Linux-only, étiquetés `unavailable`/`n/a`.

- [ ] **Step 1: Write docs. Step 2: Quality pass** — re-read Tasks 5–7 diffs against the checklist; fix in place.
- [ ] **Step 3: Full gates** — `go test ./... -count=1`; `npm run typecheck && npm run test && npm run build`; `python book/tools/audit-archive.py`; `git status` review (no data/dist junk).
- [ ] **Step 4: Commit** — `docs: surface API plan 3, ligne de sécurité dashboard, matrice de capacités Windows`

---

## Self-review (author's, kept in-document)

**Spec coverage (§8):** audit card simplification (T7), result card JFI/CAP-SLA/Ar-h/seal (T7), campaign cockpit canal + flux 4 (T6), API run provenance (T1), `/api/cost/enterprise` (T2), live snapshot enterprise cost (T5 — Snapshot fields landed in Plan 2 T7), doctor (done Plan 2 T5), security line + Windows matrix (T8). Deferred minors landed: P2 API-half-of-T6 + CGO_CHANNEL_MODEL (T1), P2 non-atomic sidecar write stays parked, P1 M4/M5 (T3), P1 M6 (T4). Plan 3 does NOT touch: `/api/cost/tier` POST (legacy surface, untouched), e2e Playwright suite (runs in Plan 5 VM batch).

**Type consistency:** RunOpts ↔ startFn ↔ campagne.Deps (bench/channel_model/flows names match manifest keys); Sinks move is mechanical (same fields, new package); LiveFrame ↔ Snapshot json tags mirrored by test; economics card reads `audit.Report.Economics` (already exposed by Plan 1).

**Risk:** startFn bench selection changes the API default only when `CGO_BENCH=dualns` is set or `bench:"dualns"` posted — legacy hosts byte-identical behavior. Audit/start sink defaults change ONLY when a bench is actually up (probe succeeds) — on a Windows dev host the ladder still lands external/8.8.8.8, same as the CLI today.

**Executor notes:** T1 first (frontend T6 posts fields it defines); T3's audit/status assertion seam — check how existing audit tests observe started params before writing the failing test; T4 — read the M6 re-review entry, it has exact line numbers; T7's collapse must persist via localStorage like LiveGuide (memory across views is the point).
