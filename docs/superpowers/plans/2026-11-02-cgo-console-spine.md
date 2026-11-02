# CGO — Console spine — plan

2026-11-02 · branch `rebuild/plan1-economics-audit` · status: ready to execute

Spec: `2026-11-02-cgo-console-spine-design.md`. It is the contract; this plan does
not restate it. Visual copy comes from `2026-10-31-cgo-operator-console-design.md`
§3.1–§3.10 as amended by the spec §9.

## How to run this plan

All commands from the repo root unless the step says `web/frontend`.

- Frontend check: `cd web/frontend && npm run typecheck && npm test`.
- Document check, when a task edits `AGENTS.md`, `docs/VERIFY.md`,
  `docs/CHECKPOINTS.md`, `docs/ERRORS.md` or `docs/superpowers/plans/README.md`:
  `python tools/verify.py --only links` → `PASS`.
- Runtime evidence needs a served build, never `vite dev`:
  1. `cd web/frontend && npm run build`
  2. `cd ../.. && go run ./cmd/cgo --serve --addr 127.0.0.1:9091 --mode observe`
  3. `node web/frontend/qa-walk.mjs http://127.0.0.1:9091`
  4. `node web/frontend/capture-fresh.mjs http://127.0.0.1:9091 docs/img/shots/2026-11-02`
- Screenshots go under `docs/img/` because the `hygiene` gate rejects a tracked
  file above 384 KB outside that prefix, and a 1920×1080 PNG exceeds it.
- A screenshot proves a render, not a measurement. A task report says what it
  did not prove (R4); the slice checkpoint is T9.
- Stage explicit paths. Never `git add -A`.
- Tooling is at 649 of a 650-line ceiling (`docs/CHECKPOINTS.md`, last entry).
  No new gate goes into `tools/verify.py`; T8's lexicon gate is a vitest file.

## T0 — One active slice

**Files.** `docs/superpowers/plans/README.md`,
`docs/superpowers/specs/2026-11-02-cgo-console-spine-design.md`, this plan.

**Change.**

1. In the index: add `Superseded` to the status legend; set row 21 to
   `**In progress** (builder pass; verifier pending)` with this plan's path.
2. Re-disposition fifteen rows, nothing deleted:
   - `Superseded by 21`: 8, 14, 15, 18, 20A, 20B, 20C.
   - `Deferred` with the named unblock already in the column: 4, 5, 6, 7, 16, 17, 19.
   - `In progress`: 21.
   - Slices 1, 2, 3, 9, 10, 11, 12, 13, 20 already read `Done`; a closed row
     needs no edit. Say so in the row-21 note.
3. Spec amendments, so the contract matches the plan:
   - §5 table gains a `T0` row (ten tasks) and T1 says the state word, not the
     ribbon content, survives in the header.
   - §4.1 corrected: the ribbon is deleted; nothing moves to the drawer except
     the measurement detail the drawer already owns.
   - §6 screenshot path becomes `docs/img/shots/2026-11-02/`; §9 gains amendment
     7 saying why (the `hygiene` 384 KB ceiling).
   - T9 narrowed to the verifier packet, the twelve dispositions and the book
     annex left in the working tree; no `pkg/figures` change.

**Failing assertion.**
`awk -F'|' '$2 ~ /^ *[0-9]/ && $6 ~ /In progress/' docs/superpowers/plans/README.md | wc -l`
prints `1`. It prints `5` before this task (8, 16, 17, 20A, 20C).

**Run.**

```
awk -F'|' '$2 ~ /^ *[0-9]/ && $6 ~ /In progress/' docs/superpowers/plans/README.md | wc -l   # 1
awk -F'|' '$2 ~ /^ *[0-9]/ && $6 ~ /Superseded/' docs/superpowers/plans/README.md | wc -l    # 7
awk -F'|' '$2 ~ /^ *[0-9]/ && $6 ~ /Deferred/' docs/superpowers/plans/README.md | wc -l      # 7
python tools/verify.py --only links                                                          # PASS
```

**Not proved.** The dispositions are a decision, not a re-measurement. No
`VERIFY` row changes state, and the deferred rows keep their named unblock.

**Commit.** `docs: le plan de la tranche 21 ouvre une seule tranche active`

## T1 — Shell: one header, one navigation

**Files.** `web/frontend/src/components/TopNav.tsx` (new),
`components/ModeRail.tsx` and `components/StatusRibbon.tsx` (deleted),
`App.tsx`, `store/ui.ts`, `styles/index.css`, `lib/tokens.test.ts`,
`lib/designFloor.test.ts`, `lib/benchLayout.test.ts`, `lib/ribbon.ts` and
`lib/ribbon.test.ts` (deleted), `qa-walk.mjs`.

**Change.**

- `TopNav` renders `MODES` as four `<button data-mode=…>` inside
  `<nav className="topnav" aria-label="Destinations">`, `aria-current="page"` on
  the active one, `preloadMode` on hover and focus, `<kbd>` key hint. `data-mode`
  survives, so `qa-walk` and `capture-fresh` keep clicking the same selector.
- Header: wordmark, `TopNav`, right cluster = connection dot
  (`useUIStore(s => s.connected)`), `#hd-state` (`live?.phase ?? 'au repos'`),
  last-audit age via `relativeTime(report?.timestamp)`, settings button. Footer
  keeps one provenance line; the second line and the shortcut strip go.
- `store/ui.ts`: delete `railPinned` and `setRailPinned` (only `App.tsx` and the
  deleted rail read them). Delete the `mode-rail` / `mode-btn` / `mode-pin` /
  `mode-foot` / `status-ribbon` / `ribbon-*` rules, `#shell.rail-min`, and the
  rail branches of the 899 px and 700 px media blocks.
- Under 700 px the same `<nav>` becomes the bottom bar: fixed, 56 px, four equal
  targets, labels visible, key hints hidden. One element, no duplicate nav.
- `.btn`: `font: 600 var(--fs-body) var(--font-ui)`, `min-height: 48px`,
  `padding: 0 var(--sp-5)`, `letter-spacing: 0`, `text-transform: none`. The
  44 px mobile override becomes 48 px.
- `App.tsx` keeps `MODES.find(m => m.key === e.key)` untouched.

**Failing assertion.** `tokens.test.ts` currently requires `.mode-rail` to turn
into a row and to be fixed at 390 px. Replace both with `.topnav` rules;
`designFloor.test.ts` reads `ModeRail.tsx` and must read `TopNav.tsx`, asserting
`MODES`, `aria-current` and `preloadMode`. `benchLayout.test.ts:67` reads the same
file. Add one case: `TopNav.tsx` exists and `ModeRail.tsx` does not.

**Run.**

```
cd web/frontend
npx vitest run src/lib/tokens.test.ts src/lib/designFloor.test.ts src/lib/benchLayout.test.ts
npm run typecheck
```

**Screenshot.** All four destinations at 1920, dashboard at 390: no 232 px rail,
one 64 px header, four destinations, no ribbon. `qa-walk` mobile step asserts the
bottom-bar target is ≥ 48 px and the destination count is four.

**Not proved.** The primary action is not yet above the fold; that is T5. `.btn`
at 48 px changes density on every surface; the screenshots are the check, and the
bench density is re-checked in T7.

**Commit.** `ui: un seul en-tete, une seule navigation`

## T2 — The verdict, its four facts, and sealed objectives

**Files.** `pkg/metrics/cost.go` (`CAPSLA` struct), `pkg/audit/report.go:345`
(the reassembly), `web/frontend/src/lib/types.ts` (`AuditCapsla`),
`lib/dashboard.ts` (`operatorFacts`), `components/VerdictHero.tsx`,
`views/DashboardView.tsx`, `views/AuditView.tsx`, `lib/dashboard.test.ts`,
`components/VerdictHero.dom.test.tsx`, `pkg/metrics/cost_test.go`.

**Change.**

- `CAPSLA` gains `TargetPct float64 \`json:"target_pct"\`` and
  `DeadlineMs int \`json:"deadline_ms"\``; `AssessCAPSLA` fills them from
  `CAPBreachThresholdPct` and `CAPDeadlineMs`; `report.go:345` copies the same
  constants when it rebuilds the struct from `res`. Additive, sealed with the
  report. An unmeasured SLA stays `CAPSLA{}`.
- `operatorFacts`: the two hardcoded strings at `lib/dashboard.ts:128` and `:136`
  become `objectif ${report.cap_sla.target_pct} %` and
  `objectif ${report.cap_sla.deadline_ms} ms`; absent field → `''`.
- `VerdictHero`: render the `<em>` only when target or band is non-empty, joined
  by ` · `.
- Mount `<VerdictHero report={report} warnMs={settings.warnMs}
  critMs={settings.critMs} />` on the dashboard (it replaces the `syn-hero`
  heading) and at the top of `audit-read` (`AuditView.tsx:96`). This is the task
  that ends `VerdictHero.tsx` having no importer. `<OperatorStats>` stays on both
  surfaces for now: T4 removes it once the drawer's tab 2 carries the same
  detail, so no commit ships the numbers absent.

**Failing assertion.** A fixture whose `cap_sla.target_pct` is `90` must render
`objectif 90 %`. Asserting `95 %` would pass with the literal still in place, so
the test must use a value the constant does not hold. Same for `deadline_ms:
300`. A second case: `cap_sla` absent → `facts[0].target === ''` and no empty
`<em>`.

**Run.**

```
go test ./pkg/metrics ./pkg/audit
cd web/frontend && npx vitest run src/lib/dashboard.test.ts src/components/VerdictHero.dom.test.tsx
```

**Screenshot.** Dashboard and audit read state: the 40 px word, one sentence,
four facts, no objective line on the six archived reports (they predate the
field) — the honest absence the spec §8 predicts.

**Not proved.** No live audit. The green/amber/red cascade is proved by unit
fixtures only. No archived report gains a verdict retroactively.

**Commit.** `feat: le verdict scelle ses deux objectifs`

## T3 — The remediation card

**Files.** `web/frontend/src/components/RemediationCard.tsx` (new),
`lib/types.ts` (`AuditRemediation`), `views/AuditView.tsx` (replace
`.audit-remedy`), `views/DashboardView.tsx` (mount under the hero),
`components/RemediationCard.dom.test.tsx` (new).

**Change.**

- `AuditRemediation` gains the fields the engine already emits
  (`pkg/audit/report.go:209-240`): `up_measured_mbps`, `up_shaped_mbps`,
  `down_measured_mbps`, `down_shaped_mbps`, `linux_iface`,
  `routeros_wan_iface`, `routeros_lan_iface`, `routeros_lan_assumed`. No Go edit.
- The card renders design §3.3: title `Protéger ce lien`, one line per measured
  direction (`Envoi` / `Réception`) with measured and shaped rate; an unmeasured
  direction reads `non mesurée — auditer le sens montant` (or `descendant`) and
  contributes no script. A `radiogroup` lists only platforms the report carries;
  the choice persists in `localStorage` under `meteolink-remediation-platform`.
- `RouterOSLanAssumed` renders `interface supposée « <name> » — à vérifier avant
  de coller`; when no ingress script exists, no assumption line is rendered,
  because none is labelled.
- Primary action `Copier la configuration`; secondary `Voir la commande exacte`
  opens the T4 drawer on tab 1. Until T4 lands, the secondary button is absent;
  T4 adds it and edits this file — do not ship a dead button.

**Failing assertion.** Fixture A: down measured, `routeros_lan_assumed: true`,
`routeros_ingress` non-empty → the rendered markup contains `interface supposée`
and the interface name. Fixture B: same rates, no `routeros_ingress` → the
markup contains no `interface supposée`. Fixture C: down absent → no `Réception`
rate and no ingress command.

**Run.**

```
cd web/frontend && npx vitest run src/components/RemediationCard.dom.test.tsx
npm run typecheck
```

**Screenshot.** Audit read state with a fixture report: two rate lines, platform
radio, one primary action.

**Not proved.** No script was executed on a router. Copy is proven, delivery is
not; `routeros_validation` wording stays `non validée` / `aucune preuve`.

**Commit.** `feat: la carte de remediation nomme ce qui est mesure`

## T4 — The proof drawer

**Files.** `web/frontend/src/components/ProofDrawer.tsx` (new),
`store/ui.ts` (proof slice), `App.tsx` (mount once), `views/DashboardView.tsx`,
`views/AuditView.tsx`, `components/RemediationCard.tsx` (the secondary action),
`styles/index.css` (sheet rules), `qa-walk.mjs`,
`components/ProofDrawer.dom.test.tsx` (new).

**Change.**

- `store/ui.ts`: `proof: { tab: 'commands' | 'proof' } | null`,
  `openProof(tab)`, `closeProof()`. `App.tsx` renders `<ProofDrawer />` once,
  reading `useAuditStore(s => s.report)`.
- `ProofDrawer` uses `useModalFocus` and `createPortal` like `SettingsDrawer`;
  `role="dialog" aria-modal="true"`, `tablist` with two tabs, 85vh, max 720 px,
  full width under 700 px, Escape closes, focus returns to the trigger.
- Tab `Déploiement & scripts`: the selected platform's commands in `<pre>`, one
  copy button per block, sizing basis (measured rate, `× 0,95`, idle RTT, qdisc),
  rollout and verification commands. Tab `Preuves & télémétrie brute`: the
  `operatorFacts` figures, the `OperatorStats` detail (RTT p50/p95 idle and
  loaded, worst delta, QDI, loss, throughput, `n`, direction), sensor tier and
  reason, source and telemetry SHA-256, link to
  `/api/audit/report?audit_id=…`.
- Once tab 2 renders that detail, remove the `<OperatorStats>` calls from
  `DashboardView.tsx:109` and `AuditView.tsx:104`; this is the commit where the
  numbers stop appearing twice.
- The border router's rejeu line reads `validée par rejeu`, `non validée` or
  `aucune preuve`. Never `appliquée`.

**Failing assertion.** Static render with `proof: { tab: 'proof' }`: the proof tab
has `aria-selected="true"` and the panel contains `provenance.source_sha256`;
with an upload-only report, tab 1 shows no ingress block. The focus contract is
tested where it can be: `qa-walk.mjs` opens the drawer, asserts focus is inside,
presses Escape, and asserts focus is back on the trigger — the same three steps
it already runs for the settings drawer.

**Run.**

```
cd web/frontend && npx vitest run src/components/ProofDrawer.dom.test.tsx
node web/frontend/qa-walk.mjs http://127.0.0.1:9091
```

**Screenshot.** Drawer open on each tab at 1920 and 390.

**Not proved.** The drawer renders a hash it did not verify; the seal check is
the `frozen` gate's job. No archived report is re-fetched here.

**Commit.** `feat: le tiroir de preuve porte les deux lectures`

## T5 — Three destinations, one column, action above the fold

**Files.** `web/frontend/src/views/DashboardView.tsx`, `views/AuditView.tsx`,
`views/EvidenceView.tsx`, `styles/index.css`, `qa-walk.mjs`,
`lib/designFloor.test.ts`.

**Change.**

- Delete `dash-grid`, `syn-proof`, `syn-next`, `dash-onboarding`, `syn-actions`
  and the three inline `ScienceDetails`; each is a second navigation, a developer
  count, or a duplicate of the drawer. Delete their rules from
  `styles/index.css` in the same commit. `syn-proof` takes the `/api/integrity`
  fetch, the `Integrity` type and the `headline`/`next` helpers with it, so the
  landing page no longer scans 159 manifests. The dashboard becomes one column:
  verdict, remediation, cost, change. The corpus counts stay on `Preuves`.
- `AuditView`: the read state keeps verdict, remediation, proof entry, one
  `Nouvel audit`; the inline `ScienceDetails` wrapping `AuditReportCard` is
  replaced by the drawer's tab 2.
- `EvidenceView`: four counts at 32 px, one primary `Vérifier l'intégrité`, the
  report rows' `Ouvrir` unchanged, and the runs/replay/quarantine tables folded
  behind one `Détails du corpus`. `ScienceDetails` survives here — this is the
  corpus drawer, not a second copy of the report.
- `qa-walk.mjs` gains two measured checks and keeps two:
  1. the first `.btn-primary` of each destination has
     `box.y + box.height <= viewport height` at 1440×900;
  2. no horizontal overflow at 1920, 1366 and 390
     (`document.documentElement.scrollWidth <= innerWidth` after each resize);
  3. zero console and resource errors (existing, exit code);
  4. no long task above 400 ms (existing).
- `designFloor.test.ts` gains one case: the five deleted class names appear in no
  view.

**Failing assertion.** The fold check fails on this branch before the deletion at
1440×900, because the rail and the stacked cards push `Auditer un lien` below
900 px. It must fail first and pass after, on all four destinations.

**Run.**

```
cd web/frontend && npm run build
cd ../.. && go run ./cmd/cgo --serve --addr 127.0.0.1:9091 --mode observe
# second terminal
node web/frontend/qa-walk.mjs http://127.0.0.1:9091        # 0 failures
cd web/frontend && npx vitest run src/lib/designFloor.test.ts
```

**Screenshot.** Four destinations at 1920×1080 and dashboard at 390×844, plus one
1366×768 overflow capture.

**Not proved.** "Above the fold" is proved at 1440×900 and 1920×1080 only. No
tablet width is claimed. The screenshots still show `Pas encore mesuré` on the
archived reports.

**Commit.** `ui: trois destinations, une colonne, l'action visible`

## T6 — The live trace draws itself

**Files.** `web/frontend/src/lib/types.ts` (`AuditTimelinePoint`, `timeline`,
`deadline_ms` on `AuditLive`), `lib/trace.ts` (new, pure),
`components/TraceCanvas.tsx` (new), `components/AuditLive.tsx`,
`styles/index.css`, `pkg/api/server_audit.go` (one key),
`lib/trace.test.ts` (new), `pkg/api/audit_live_test.go`.

**Change.**

- The engine already ships `timeline` (`pkg/audit/timeline.go`, one bucket per
  second, `null` when no probe answered). Add `deadline_ms` to the live payload
  from `metrics.CAPDeadlineMs`, so the browser never retypes 220.
- `lib/trace.ts`: `buildTrace(points, deadlineMs)` returns `segments` (a new
  segment after every `null`), `yMax` (max of observed and the objective), and
  `objectiveY`. Pure; the canvas only draws it.
- `TraceCanvas`: one `<canvas>`, 120 px tall, X `0 → total_sec`, phase bands,
  the objective as a labelled rule, a gap at every break, single-pixel legend
  `sonde sans réponse`, accessible name and the two live facts as its text
  alternative. No ECharts, no per-sample DOM, no animation.
- `AuditLive` replaces the four percentile tiles with `temps de réponse` and
  `réponse sous charge` against the objective; `débit` is absent during the run
  because `audit.Progress` carries no byte counter (spec §9.8), and it appears
  after the freeze. The alert percentage appears only after the freeze too.

**Failing assertion.** `buildTrace([{t:0,rtt_p50:30},{t:1,rtt_p50:null},
{t:2,rtt_p50:220}], 220)` returns two segments, `yMax >= 220`, and
`objectiveY` inside the plot. A trace with no `null` returns one segment. This is
the gap property the slice exists to show; a render-only test cannot see it.

**Run.**

```
go test ./pkg/api -run TestAuditLiveAndStop
cd web/frontend && npx vitest run src/lib/trace.test.ts
```

**Screenshot.** No live audit, and the canvas mounts only while
`stage === 'running'`, so the capture shows the declaration step. The canvas is
proved by `trace.test.ts` and by a static render of `TraceCanvas` asserting its
accessible name and the `sonde sans réponse` legend. Say that in the task report.

**Not proved.** No live run, so no real gap is drawn. The canvas is proved by a
pure-function test and a static render, not by a VSAT timeout.

**Commit.** `feat: la trace vivante casse la ligne sur une sonde muette`

## T7 — The bench reads as one decision

**Files.** `web/frontend/src/components/DuelCard.tsx` (new),
`components/BenchConfig.tsx`, `views/BenchView.tsx`, `lib/bench.ts`,
`store/bench.ts`, `styles/index.css`, `capture-fresh.mjs`,
`components/DuelCard.dom.test.tsx` (new), `components/BenchConfig.dom.test.tsx`
(new).

**Change.**

- `store/bench.ts` gains `research: boolean`, `setResearch`, persisted under
  `meteolink-bench-research`.
- `lib/bench.ts` gains the `DuelPair` type matching `pkg/api/duel.go`,
  `loadDuel(profile)`, and `startCampaign(opts)` extracted verbatim from
  `BenchConfig.start` (`components/BenchConfig.tsx:90-118`) so the duel's action
  posts the same body instead of a second copy. `BenchConfig.start` calls it.
- `DuelCard` is presentational: it takes `pair: DuelPair | null` and renders
  profile P2, protocol names, percentages, latency, `n`,
  `Lancer la comparaison` (arm-then-confirm, preset
  `P2 × 3 reps × 4 flows × clean × 220 ms`), and `aucune base exploitable` when
  `base` or `best` is null. `BenchConfig` loads `/api/duel` and passes it. The
  browser subtracts nothing: it renders `delta_deadline_pts` and
  `delta_small_ms` from the response.
- `BenchConfig` wraps `Profils`, `Flux de charge`, `Canal`, `Deadline` and the
  matrix controls in `research ? … : null`; a secondary `Mode recherche` toggle
  (`data-research`) reveals them and is remembered.
- `BenchView` renders the results body (`ResultatsView`) only when `research`,
  keeping the `#bench-results` section; `openResults` sets `research` true.
  Default bench = duel, observer, no matrix. `capture-fresh.mjs` clicks
  `button[data-research]` before the results shot.

**Failing assertion.** Static render `BenchConfig` twice: with
`research: false`, `Profils` and `Deadline` are absent and `Mode recherche` is
present; after `useBenchStore.setState({ research: true })`, they are present.
`DuelCard` with `pair={null}` renders `aucune base exploitable` and no delta.

**Run.**

```
cd web/frontend && npx vitest run src/components/DuelCard.dom.test.tsx src/components/BenchConfig.dom.test.tsx
```

**Screenshot.** Bench default at 1920 and 390; research mode at 1920.

**Not proved.** No bench run; the card renders `/api/duel` over archived results.
The P2 numbers are the corpus's, not new measurements.

**Commit.** `ui: le banc montre un duel, la matrice passe en mode recherche`

## T8 — The operator lexicon

**Files.** `web/frontend/src/lib/lexicon.test.ts` (new), plus every string the
test catches in the four default destinations; `DESIGN.md`, `PRODUCT.md`.

**Change.**

- The test renders the four default surfaces with `renderToStaticMarkup` — the
  pattern already used by `AuditReportCard.dom.test.tsx` — after seeding the
  stores: `DashboardView` and `AuditView` with a fixture report and
  `stage: 'done'`, `BenchView` with `research: false`, `EvidenceView` in its
  initial state. Effects do not run under SSR, so no fetch and no 9 s scan.
- The rendered text is stripped of tags and scanned against the design §5 banned
  list: `p50`, `p95`, `qdisc`, `CAKE`, `fq_codel`, `BBR`, `CUBIC`, `JFI`, `QDI`,
  `SLA`, `bufferbloat`, `eBPF`, `manifeste`, `corpus`, `gel`, `gelé`, `twin`,
  `run`, `n=`, hex strings, `garanti`, `garantie`, `prouvé` without a named
  sealed proof beside them. The drawer and research panels are not rendered, so
  they need no allowlist: the gate covers exactly the default surface.
- Rewrite the caught strings: `corpus` → `mesures archivées`, `run` → `mesure`,
  `manifeste` → `empreinte`, `p50/p95` → `médiane / haut de gamme` or a plain
  word, `SLA` → `objectif d'alertes`, `gelé` → `scellé`. Keep `SHA-256` under
  `Empreinte de preuve`; a full hash moves into the drawer.
- `DESIGN.md` and `PRODUCT.md`: replace any copy that still promises the rail or
  the ribbon, and align the lexicon table with design §5 plus the T2 objective
  fields.

**Failing assertion.** Before the rewrite the test prints each offending
`file: phrase`; after, zero. A control case runs the same scanner over the
literal `'le corpus gelé'` and expects it caught, so an empty render cannot pass.

**Run.**

```
cd web/frontend && npx vitest run src/lib/lexicon.test.ts
npm run typecheck && npm test
python tools/verify.py --only links
```

**Screenshot.** Four destinations, default state: no banned word visible.

**Not proved.** A word inside the drawer or a research panel is not covered,
because those surfaces do not render in the test; design §5 is their rule, not
this gate. `13 px is drawer-only` (spec §9.4) stays a rule until a CSS-level type
scale test exists; this task does not add one.

**Commit.** `ui: le vocabulaire operateur devient une porte`

## T9 — The evidence tail

**Files.** `docs/verification/2026-11-02-slice-21.md` (new), `docs/VERIFY.md`,
`docs/CHECKPOINTS.md`, `book/content/09-annexes.js` (working tree only, never
staged).

**Change.**

- The packet: the exact commands, host, date, commit, the screenshot paths, the
  four runtime checks' output, and one table with the twelve `PENDING` claims and
  the builder's proposed disposition (spec §7.1) — `SUPERSEDED` for slice 8 and
  Interface v2, `DEFER` for BBRv3 and `.dist`, `re-run, promote` for the seven
  20A rows and 20C. The verifier decides each row; the builder only proposes.
- `docs/VERIFY.md`: one new row, `Slice 21 console spine (builder pass)`,
  `PENDING`, with the packet path and the four gates' results. No existing row
  changes status.
- `docs/CHECKPOINTS.md`: one entry with the command, date, host, result and what
  stayed unproven.
- The book annex gets one paragraph citing the console and the packet; it stays
  in the working tree because `book/` is never committed.

**Failing assertion.** The packet contains twelve disposition lines, one per
`PENDING` claim, and the set of claim texts equals the set in `docs/VERIFY.md`.
A missing row fails the check.

**Run.**

```
python tools/verify.py                    # 6 passed, 0 failed, 0 not run
cd web/frontend && npm run build && npm test
cd ../.. && go test ./... -count=1
python book/tools/audit-archive.py        # two OK lines; book is not staged
```

**Screenshot.** Final set under `docs/img/shots/2026-11-02/`, four destinations at
1920 and 390.

**Not proved.** No row is `VERIFIED`: the builder may not certify its own number.
The slice closes at `builder pass complete, packet ready`; a separate verifier
re-runs the packet and owns the rows. No live audit, no bench run, no merge.

**Commit.** `docs: le packet de la tranche 21, sans auto-certification`

## Ledger after this plan

| Task | Surfaces touched | Screenshot | Row moves |
|---|---|---|---|
| T0 | index, spec | — | 15 rows re-dispositioned |
| T1 | shell | dashboard 1920/390 | — |
| T2 | hero + engine | dashboard, audit | — |
| T3 | remediation | audit | — |
| T4 | drawer | drawer ×2 | — |
| T5 | three destinations | four ×2 widths | — |
| T6 | live trace | audit observer | — |
| T7 | bench | bench ×2 | — |
| T8 | lexicon | four | — |
| T9 | packet | final set | one `PENDING` row added |
