# CGO Enterprise Rebuild — Plan 3: UI / API Surface Implementation Plan (révisé après brainstorm design)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

> **État au 2026-10-27 (branche `rebuild/plan1-economics-audit`) :** T1–T8 livrés (`c45f6fb`, `09d117f`, `2764200`, `d277c7f`, `5fa92be`, `1ddab5c`, `33e9356`, `3c6de9b`), T9 `5c13fcd`, T10 `a759557`, garde M2 + alignement `flows` `99f6351`, puis revue critique de fermeture : correctifs Go `eabf0db`, UI `968026e`. Portes mesurées : `go test ./... -count=1` PASS ; frontend `typecheck` + 37 fichiers / 139 tests + `build` PASS ; `audit-archive.py` PASS (159 runs, 208 valides, 74 degraded, 246 invalides) ; détecteur impeccable = 2 avertissements pinnés (Inter, dégradé METEOLINK), 0 transition de layout. Revue **locale** (subagent indisponible — limite fournisseur) : constats réels corrigés plutôt que documentés — annulation sans résultat partiel, rapport effacé au démarrage, cible d'audit laissée au banc détecté, `aria-sort`/`aria-pressed` sur les résultats, garde d'id partagée par les deux artefacts. Rulings et écarts mesurés : ledger `.superpowers/sdd/2026-10-25-cgo-plan3-ui-api-surface/progress.md`.

**Goal:** Spec §8's API slice PLUS the design overhaul decided in the 2026-10-26 brainstorm (critique 23/40, snapshot `web/frontend/.impeccable/critique/`): run provenance through `/api/run/start`; `GET /api/cost/enterprise`; bench-aware audit sinks (M4/M5); the `both`-mute report fix (M6); **the audit becomes observable** (`/api/audit/live` + stop); the audit card is rebuilt as a staged experience (Déclarer → Observer → Lire) that renders the full report; the results section is depouillée; and a system-wide extract pass enforces the existing token system (« modeste, spacieux, élégant », Inter conservé).

**Architecture:** Go API first (T1–T5), frontend in priority order (audit card → cockpit/live features → results → system pass). The audit engine gains an `OnProgress` seam (phase + live samples) mirroring the campaign wall's `publishLive` pattern; sink detection moves to `pkg/audit`. Frontend: primitives (`Card`, `Stat`, `Badge`, `.c-input`) are adopted in T6 FIRST and the extract pass (T9) generalizes them — no view is rebuilt on inline styles then migrated twice.

**Tech Stack:** Go 1.25 stdlib only; React 19 + zustand + vitest (jsdom); `npm run typecheck` / `test` / `build` from `web/frontend`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` (§2 honesty invariants, §8, decisions 6–8) + brainstorm decisions (2026-10-26): full staged audit, extract pass toutes vues, Inter conservé.
- Go 1.25 stdlib only, no new dependencies (frontend deps frozen); comments French, identifiers English.
- Windows-green: `go build ./...`, `go vet ./...`, `go test ./pkg/api/... ./pkg/audit/... ./cmd/cgo/... -count=1`; frontend `npm run typecheck && npm run test && npm run build`.
- Frozen-data guard: `book/data/runs`, `pkg/results/runs-frozen.tar.gz` untouched; CSV schemas immutable; `book/tools/audit-archive.py` byte-identical after every task.
- Never fabricate: unmeasured ⇒ stated (badge « non mesuré », `Unavailable`, dashed tiles), never a zero that reads as measured; scenario ≠ claim (labels carry « scénario contractuel »).
- **Design floor (brainstorm)**: no new hardcoded hex in `.tsx` (tokens/`var()` or the palette module only); no font size outside the 5 tokens (`--fs-micro..--fs-head`; nothing below 10px); spacing from `--sp-1..5`; primitives before raw divs; `aria-pressed` on every toggle, `aria-sort` on sortable tables; `prefers-reduced-motion` respected (existing anime.js gating). The impeccable skill's craft-floor reference governs every UI task.
- Bursty is P2/P3-only (decision 7): enforced at the API (400 before mutation) and in campagne (last word).
- Commit after every task with the given message.

---

### Task 1: `/api/run/start` — provenance `bench` + `channel_model`, défaut `flows=4`

**Files:**
- Modify: `pkg/api/server.go` (`RunOpts`, run/start validation), `cmd/cgo/server.go` (`startFn`), `cmd/cgo/runcli.go` (précédence `CGO_CHANNEL_MODEL`)
- Modify: `pkg/api/run_start_filters_test.go` (new cases)

**Interfaces:**
- `RunOpts` gains `Bench string \`json:"bench"\`` and `ChannelModel string \`json:"channel_model"\``.
- Validation (before the engine-wired check): `bench` ∈ {"", "legacy", "dualns"} else 400; `channel_model` ∈ {"", "clean", "bursty"} else 400 naming the rule; `channel_model=="bursty"` ⇒ every profile ∈ {P2, P3} else 400 naming the offender(s) + the rule.
- `startFn`: `o.Flows == 0 ⇒ 4` (decision 6 — today's code lands on 1); bench resolution mirrors `runCLI` (`o.Bench` → `CGO_BENCH` → `"legacy"`); `"dualns"` ⇒ `ProdDepsDualNS(o.Direction)` (direction-aware) replaces `ProdDeps()/ProdDepsDown()`; `deps.ChannelModel = o.ChannelModel`.
- `runcli.go`: `--channel-model` default becomes `envDefault("CGO_CHANNEL_MODEL", "clean")` (env no longer dead; explicit flag still wins — same pattern as `--bench`).

- [x] **Step 1: Failing tests** — bursty+P1 ⇒ 400 naming `P1`; canal inconnu ⇒ 400; banc inconnu ⇒ 400; bursty+P2 sans moteur ⇒ 503 (gate order); provenance reaches `StartFn` intact (`got.ChannelModel/Bench/Flows`).
- [x] **Step 2: red → Step 3: implement → Step 4: green** — `go test ./pkg/api/ ./cmd/cgo/ -count=1`; audit-archive byte-identical.
- [x] **Step 5: Commit** — `api: run/start accepte bench + channel_model (bursty gate P2/P3), flows=4 par défaut`

---

### Task 2: `GET /api/cost/enterprise` — catalogue plans/tarifs/provenance

**Files:**
- Modify: `pkg/api/server.go` (endpoint)
- Create: `pkg/api/enterprise_cost_test.go`

**Interfaces:** `GET /api/cost/enterprise` → 200 JSON: `{plans:[{name,monthly_opex_mga,cir_mbps|fup_bytes}], tariffs:[{name,ar_per_gb,quoted}], emergency_ar_per_gb, primary_mapping{fiber,other,4g,5g,vsat}, disclaimer}` — every value serialized FROM the `pkg/metrics` vars and `primaryName` mapping (single source; a tariff change updates both surfaces). `quoted:true` on the APN (devis opérateur, invariant 4).

- [x] **Step 1: Failing test** — names present; APN `quoted:true`, wholesale `false`; emergency 5556; disclaimer verbatim; five mapping keys.
- [x] **Step 2: red → implement → green** — `go test ./pkg/api/ -count=1`.
- [x] **Step 3: Commit** — `api: GET /api/cost/enterprise — plans, tarifs, provenance, mapping primaire`

---

### Task 3: audit/start — puits décidés par le banc (ResolveSinks partagé) + branches 400 testées

**Files:**
- Create: `pkg/audit/sinks.go`, `pkg/audit/sinks_test.go`
- Modify: `cmd/cgo/auditsinks.go` (thin alias), `pkg/api/server.go` (`Deps.SinksFn`, defaults, dead-branch removal)
- Create: `pkg/api/audit_start_sinks_test.go`

**Interfaces:**
- `pkg/audit` exports `Sinks{Mode,Target,SmallURL,BulkAddr}`, `ResolveSinks()` (dualns 10.30.0.2 → legacy 10.200.0.1 → external), `EnvOrBulk(s)`. `cmd/cgo` keeps delegating wrappers.
- `api.Deps` gains `SinksFn func() audit.Sinks` (nil ⇒ real). audit/start: empty `target` ⇒ `sinks.Target` (external keeps 8.8.8.8); empty `small_url` ⇒ `sinks.SmallURL` (**empty stays empty in external — jamais de charge implicite**); empty `bulk_addr` ⇒ `EnvOrBulk`; env precedence unchanged. The `recordEvent` start line names the detected bench.
- M5: remove the dead `Flows < 1` arm; pin 400 branches (direction invalide, flows 9, cible > 64).

- [x] **Step 1: Failing tests** — dualns stub ⇒ params carry the 10.30.0.2 trio; external stub ⇒ small stays "" + bulk via env only; the 400 branches.
- [x] **Step 2: red → implement → green** — `go test ./pkg/audit/ ./pkg/api/ ./cmd/cgo/ -count=1`; audit-archive byte-identical.
- [x] **Step 3: Commit** — `api/audit: puits détectés par le banc (ResolveSinks partagé), branches 400 testées`

---

### Task 4: rapport d'audit `both` — miroirs mesurés rendus, économie absente dite

**Files:**
- Modify: `pkg/audit/report.go`, `pkg/audit/report_test.go`

**Context:** re-review M6 (`.superpowers/sdd/2026-10-25-cgo-plan1-economics-audit/re-review-t6-t11.md`) — read it first; it carries file:line specifics.

**Interfaces:** mute direction ⇒ the OTHER direction's measured mirrors still render (labelled); `Economics == nil` ⇒ explicit `Unavailable` entry (« économie entreprise — fenêtre de charge vide »); the « voir grade_down » note references what the report actually carries.

- [x] **Step 1: Failing test** — both fixture, mute up + measured down: down mirrors in Markdown; Unavailable names economics; no dangling reference.
- [x] **Step 2: red → implement → green** — `go test ./pkg/audit/ -count=1`; audit-archive byte-identical.
- [x] **Step 3: Commit** — `audit: rapport both — miroirs down rendus, économie absente dite Unavailable`

---

### Task 5: audit observable — `OnProgress`, `GET /api/audit/live`, `POST /api/audit/stop`

**Files:**
- Modify: `pkg/audit/audit.go` (`Deps.OnProgress`, calls in `collectWindow` + phase boundaries), `pkg/audit/audit_test.go`
- Modify: `pkg/api/server.go` (endpoints, atomic state, cancel wiring), `pkg/api/audit_live_test.go` (create)

**Interfaces:**
- `audit.Progress{Phase string, ElapsedSec, PhaseRemainSec int, LastRTTms, LastSmallMs float64, SamplesRTT, SamplesSmall []float64}` — `Phase` ∈ {`repos`,`charge`,`charge-down`,`drain`} (+ `both` up/down windows distinguishable); `Deps.OnProgress func(Progress)` (nil ⇒ no-op) called each collect round (~1 Hz) and at each phase start. Sample slices carry the current window's last ≤32 values (live p50/p95 computable client-side exactly like the campaign wall).
- Server: the audit goroutine stores the latest Progress in an `atomic.Value` + start time; `GET /api/audit/live` → `{running, phase, elapsed_sec, total_sec, phase_remain_sec, last_rtt_ms, last_small_ms, samples:{rtt_ms:[],small_ms:[]}}`; not running ⇒ `{running:false}` (honest, no zeros-as-measured). `POST /api/audit/stop` cancels the run context (stored cancel func; `auditMu`-guarded) — the goroutine's error path records « annulé par l'opérateur » in `recordEvent` (distinguish cancellation from failure), no partial CSV row is appended.
- `both`: the live phase names the active window (`charge` then `charge-down`).

- [x] **Step 1: Failing tests** — audit test with a scripted Deps (instant fakes, Duration minimal): OnProgress fires with `repos` then `charge` phases, monotone ElapsedSec, LastRTT equals the fake's value; API test: stub Deps running ⇒ live returns phase/elapsed; stop ⇒ subsequent live `running:false` and the goroutine's event says annulé (assert via injected RecordEvent seam if one exists, else via status polling).
- [x] **Step 2: red → implement → green** — `go test ./pkg/audit/ ./pkg/api/ -count=1`.
- [x] **Step 3: Commit** — `audit: observable — OnProgress, /api/audit/live (phase + échantillons), /api/audit/stop`

---

### Task 6: CampagneView — la carte audit en trois temps (Déclarer → Observer → Lire)

**Files:**
- Modify: `web/frontend/src/views/CampagneView.tsx` (audit card 297–498 rebuilt)
- Create: `web/frontend/src/lib/auditFlow.ts` (+ test) — stage machine + validation/normalisation
- Create: `web/frontend/src/components/AuditLive.tsx` (Observer) + `AuditReportCard.tsx` (Lire)
- Modify: `web/frontend/src/lib/types.ts` (AuditLive, report mirror types), tests colocated

**Interfaces:**
- **Déclarer** — four decisions visible: Site, Type de lien (select), Durée (segmented 20/60), Sens (up/down/both, three-option group — POST gains `direction`). Presets stay as chips. Everything else (Fournisseur, Cible ping, ifaces, durée libre) collapses behind « avancé » (LiveGuide pattern: `useState` + `localStorage('audit-advanced-open')`, replié par défaut). Validation lives in `auditFlow.ts` (pure, tested).
- **Observer** — replaces the blind « audit en cours… »: polls `/api/audit/live` (2 s), renders the phase timeline Repos → Charge → Drain (with `charge-down` named for `both`), a remaining-time counter, live RTT/small readouts (p50/p95 over the window, same math as the wall), and **ANNULER** (`POST /api/audit/stop`, confirm-once via ArmButton). Reduced to a quiet state when `/api/audit/live` says not running and status flips.
- **Lire** — `AuditReportCard` consumes the FULL report: **Constat** (grade badge via `LEVEL_COLOR`, Δp95, drain, sens), **Charge** (débit atteint, perte, flux, JFI), **SLA CAP** (badge: breach⇒danger, pass⇒success, `measured:false`⇒steel « non mesuré » — NEVER green), **Économie** (primaire + B2B Ar/h prominent via `formatCostAR`, five scenarios behind a details toggle, disclaimer verbatim, « scénario contractuel » label — **fixes the P0: no more `waste.*` read, economics from `economics.*`**), **Preuves** (seals truncated 12, click-to-copy via `copyText`, telemetry + markdown report link), **Remédiation** (existing copy buttons). Every `unavailable[]` entry rendered as a visible muted line.
- Primitives adopted HERE first: `Card`/`CardHead` for the card shells, `Badge` for verdicts, `.c-input` for all fields (kills the copy-pasted `'6px 8px'` idiom in this view), `Stat` for the hero numbers. No new hex; tokens only.

- [x] **Step 1: Failing tests** — `auditFlow.test.ts`: stage machine (idle→declared→observing→done + cancel path), duration segmented↔free-text normalization (20/60/other within 10–600), direction default `up`; `AuditReportCard` tone mapping (breach/pass/unmeasured; economics binding reads `economics.primary.ar_per_hour`); live math (p50/p95 from samples).
- [x] **Step 2: red → implement → green** — `npm run typecheck && npm run test`; manual capture pass (`node capture-fresh.mjs`) — the three stages screenshotted and eyeballed.
- [x] **Step 3: Commit** — `ui: audit en trois temps — déclarer 4 décisions, observer (live+annuler), lire le rapport complet`

---

### Task 7: cockpit campagne + mur live — canal, flux 4, coûts entreprise

**Files:**
- Create: `web/frontend/src/lib/canal.ts` (+ test)
- Modify: `web/frontend/src/views/CampagneView.tsx` (cockpit 310–354), `web/frontend/src/lib/types.ts` (`LiveFrame`), `web/frontend/src/views/LiveView.tsx` (bento)

**Interfaces:** `canalGate(profiles)` — bursty legal iff **tous** les profils ∈ {P2,P3} et la liste n'est pas vide (règle exacte du serveur) ; `reason` nomme la règle et les profils fautifs. Cockpit : `flows` default **4** ; clean/bursty two-option group (LiveView cubic/bbr pattern) gated by `canalGate` (blocked ⇒ disabled + reason, snaps to clean); POST body gains `channel_model`. `LiveFrame` += `cost_primary_ar_per_h`, `cost_b2b_ar_per_h`, `jfi_pct` (mirror of Go tags, keys test); bento gains « Coût entreprise » + « Coût B2B » cards, existing « Coût » relabelled « Coût détail (gelé) »; `jfi_pct` shown when > 0.

- [x] **Step 1: Failing tests** — canal gate ([P2]✓, [P1,P4]✗+reason); LiveFrame keys mirror; bento terms list contains the new cards.
- [x] **Step 2: red → implement → green** — `npm run typecheck && npm run test`.
- [x] **Step 3: Commit** — `ui: cockpit — canal clean|bursty (P2/P3), flux 4; mur — coûts entreprise + B2B + JFI`

---

### Task 8: Résultats dépouillés — identité, puces, un encodage

**Files:**
- Modify: `web/frontend/src/views/ResultatsView.tsx`, `web/frontend/src/views/CompareView.tsx`, `web/frontend/src/styles/index.css` (chip-row layout)
- Tests colocated / extracted pure helpers

**Interfaces:**
- Identity cell: 4 stacked lines + two always-visible A/B buttons ⇒ 2 lines (cellule + sous-libellé) + ONE compare toggle revealed on hover/focus (`aria-pressed`, keyboard reachable — A/B becomes a two-state of that single control).
- Chip row: one `nowrap` scroll of ~20 ⇒ three labeled groups (PROFIL / FILE / CC) that WRAP, legend into a « légende » popover; rank chips keep their place.
- Metric cells: keep bar+whisker, DROP the alpha-heat background; ±half·n and secondary stats move into the expanded detail row; numbers `font-variant-numeric: tabular-nums`, one size, units in `--text-faint`.
- Expanded detail: 5-col re-grid → definition-list rhythm; LIEN decomposition unchanged in content.
- CompareView: visual only (tokens), behavior frozen.

- [x] **Step 1: Failing tests** — pure helpers extracted where they exist (chip grouping, encoding choice); snapshot-ish term assertions per existing test style.
- [x] **Step 2: red → implement → green** — `npm run typecheck && npm run test`; capture pass (1920 + 1366) eyeballed.
- [x] **Step 3: Commit** — `ui: résultats dépouillés — identité deux lignes, puces groupées, un seul encodage par cellule`

---

### Task 9: passe système — tokens partout, primitives adoptées, dettes effacées

**Files:**
- Modify: `web/frontend/src/styles/tokens.css` (chart palette vars; alias block **conservé** comme couche de compatibilité CSS — ruling HM-1), `index.css`, `components.css` (dead classes removed), all views/components (`LiveView` control box, `IntegriteView`, `Rail`, `App`, modals, `InterpretationView`, `MetricCard`, `LiveGuide`)
- Create: `web/frontend/src/lib/palette.ts` (+ test) — the only place color constants live, fed from CSS vars

**Interfaces & gates (measurable):**
- Hardcoded hex in `.tsx`: **0** (grep gate: `grep -rEn '#[0-9a-fA-F]{6}' src --include='*.tsx'` empty; chart configs take colors from `palette.ts` reading CSS vars). The 9 off-token grays (`#767b84`, `#c3c9d1`, `#a8aeb7`, …) map to the 5-step text ramp — a mapping table in `palette.ts`, asserted by test.
- Inline styles: dynamic values only (widths/transforms/positions computed at runtime); everything static becomes classes. `'6px 8px'` input idiom gone (`.c-input`).
- Type scale: 12 inline sizes ⇒ the 5 tokens; `fontSize:9` gone; hero numbers via `Stat`; `letterSpacing`/weights from tokens.
- Alias token generation (tokens.css 91–139) **conservée comme couche de compatibilité CSS** (38 consommateurs réels hors `tokens.css` mesurés ; ruling HM-1) : la porte exige zéro alias consommé dans les `.tsx`, tous convertis aux jetons canoniques ; dead CSS (`c-lamp`, `c-arm`, `c-phase`, `c-errp`, `camp-chip`, `conduite-*`) deleted, **`c-kpi` conservé** (7 consommateurs) ; `Skeleton`/`DataTable` deleted — no dead exports remain.
- Écart mesuré (ruling HM-8) : les styles inline **statiques** ont été convertis dans les fichiers nommés ci-dessus ; `ResultatsView`/`CampagneView` (T8/T6) conservent leurs objets statiques (165 objets, aucun nouveau) hors périmètre de cette passe. La porte `designFloor.test.ts` n'encode pas cette règle ; l'écart est écrit ici plutôt que masqué.
- LiveView control box: ~12 controls ⇒ three quiet sections (Surveillance / Façonnage / Essai), every toggle `aria-pressed`; journal collapsible.
- Detector findings: 7 `width` transitions ⇒ transform/opacity (or `grid-template-rows`); Rail fake sparkline (hardcoded points) removed or fed real last-N (honesty).
- Motion/labels: `prefers-reduced-motion` paths intact; focus-visible everywhere; top-nav/rail active-state disconnect (vision finding) fixed in `App.tsx`.

- [x] **Step 1: Gates first** — write the grep-gate script (hex=0 in tsx, alias consumers=0 in `.tsx` (CSS compat kept), `'6px 8px'`=0, `fontSize:9`=0) as a vitest test reading sources (precedent: `Rail.test.ts` asserts CSS by reading files); it fails.
- [x] **Step 2: implement → Step 3: green** — `npm run typecheck && npm run test && npm run build` (bundle check); `node capture-fresh.mjs` all views eyeballed; impeccable detector re-run — layout-transition findings 0.
- [x] **Step 4: Commit** — `ui: passe système — palette depuis les tokens, primitives adoptées, alias et CSS morts effacés`

---

### Task 10: docs + passe qualité impeccable

**Files:**
- Modify: `docs/api.md` (new endpoints + defaults), `docs/deploy.md` (security line, Windows matrix)
- Quality pass (impeccable `polish` reference): re-read T6–T9 diffs against the craft floor; fix in place.

**Content:** `/api/cost/enterprise`, `/api/audit/live`, `/api/audit/stop`, run/start `bench`/`channel_model`, audit/start banc-aware defaults; security line (dashboard binds 10.99.0.2 via `run-dualns.sh`; anyone on the host can trigger shaping — lab/DSI context); Windows capability matrix (audit works; TCP_INFO/netns/doctor-offload Linux-only, labelled).

- [x] **Step 1: docs. Step 2: polish pass. Step 3: Full gates** — `go test ./... -count=1`; `npm run typecheck && npm run test && npm run build`; `python book/tools/audit-archive.py`; `git status` clean of junk.
- [x] **Step 4: Commit** — `docs: surface API plan 3, sécurité dashboard, matrice Windows + passe qualité`

---

## Self-review (author's, kept in-document)

**Spec coverage (§8 + brainstorm):** provenance API (T1), cost catalog (T2), sinks partagés M4/M5 (T3), M6 (T4), audit observable (T5), audit card staged + P0 fix + full report (T6), cockpit canal/flux + live enterprise costs (T7), results dépouillés (T8), system extract pass (T9), docs+security+matrix (T10). Brainstorm decisions honored: full staged audit ✓, all-views extract ✓, Inter kept ✓. Deferred-minor routing: P2 API-half + CGO_CHANNEL_MODEL → T1; P1 M4/M5 → T3; M6 → T4.

**Risk:** T5 touches `audit.Run`'s collect loop — keep OnProgress calls allocation-light (reuse the window slice; ≤32 cap) and nil-guarded so CLI behavior is unchanged. T9 is the largest diff — gates are grep-based so partial completion is visible, not vibes; charts keep their colors (moved to palette.ts, same values). Frozen views: CompareView behavior, frozen tables' content, the honesty doctrine — inviolable.

**Executor notes:** T6 needs T5's endpoints (order fixed); capture-fresh.mjs exists (`node capture-fresh.mjs http://127.0.0.1:9091 shots-fresh` — server: `go build -o $TEMP/cgo-serve.exe ./cmd/cgo && $TEMP/cgo-serve.exe --serve --addr 127.0.0.1:9091 --tls=false --http-addr= --mode full`, stop after captures). The critique snapshot (`.impeccable/critique/2026-10-25*`) and the code-scan findings are the design contract — re-read before T6/T8/T9. Design floor violations are re-review blockers, not style nits.
