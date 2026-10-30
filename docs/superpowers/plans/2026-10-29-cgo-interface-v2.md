# CGO interface v2 — implementation plan (slice 15)

> **For agentic workers:** execute task by task, in order. Every task ends
> green (`npm run typecheck && npm test`) and is committed with explicit
> paths. A step that cannot run is recorded as a blocker, never skipped
> silently.

*Spec: `docs/superpowers/specs/2026-10-29-cgo-interface-v2-design.md` ·
Branch: `rebuild/plan1-economics-audit` · 2026-10-29 · UI/copy only; no API,
CSV, corpus, or measurement change · `book/` stays working-tree only*

## Global constraints

- Four modes: `Tableau de bord · Auditer · Banc d'essai · Preuves & Rapports`.
  No nested tab taxonomy; the bench is one linear page with three sections.
- Audit execution state lives at the app root; a mode switch never stops
  polling and never prints “suivi interrompu”.
- Two-layer language: operator French by default (`Délai sous charge`,
  `Alertes à l'heure`, `Instabilité du lien`, `Perte financière`); statistics
  live in a `Détails scientifiques` disclosure.
- Real measurement and simulation are separate, plainly labelled blocks; no
  `twin` / `gelé` / `AQM` / `QDI` as primary labels.
- Design floor: no new hex in `.tsx`, five font-size tokens only, `--sp-*`,
  `aria-pressed` on toggles, `aria-sort` on sortable tables, reduced motion
  respected. `prefers-reduced-motion` paths stay.
- No new dependency. `check-bundle.mjs` stays the budget gate. WebGL mesh
  removed. No production claim, no live audit claimed.
- `PRODUCT.md` is gitignored; the title edit is applied there and in `book/`
  without staging either.
- French user-facing text, English identifiers. Never `git add -A`.

## Task 1 — pure libraries and their tests

**Files**
- Create `web/frontend/src/lib/auditSummary.ts` (+ test)
- Create `web/frontend/src/lib/dashboard.ts` (+ test)
- Create `web/frontend/src/lib/ribbon.ts` (+ test)
- Create `web/frontend/src/lib/bench.ts` (+ test)

**Interfaces**
- `AuditSummary {audit_id, site, link_type, timestamp?, small_p95_ms?,
  deadline_ok_pct?, grade?}`; `summaryOf(report)`;
  `compareAudits(prev, cur) → {comparable, reason?, smallDeltaPct?,
  deadlineDelta?}` (same site **and** link required);
  `loadLastSummary()`, `saveLastSummary(s)` on key `cgo-audit-last`.
- `dashboard.ts`: `delayBand(ms, warnMs, critMs)` → `{word, level}` with
  `faible|modéré|élevé|critique`; `alertsBand(pct)` → `{word, level}`;
  `instabilityBand(qdi)` → `{word, level}`; `qualityWord(grade)`;
  `remediationView(groups)` → simple/best P2 lines with `n`, or `null`;
  `retailBound(report)` → the `emergency` scenario Ar/h or `null`.
- `ribbon.ts`: `corpusFacts(integrity)` → counts string; `sealFacts(hash8)`;
  `auditFacts(report)`; `ribbonMode(healthMode)`.
- `bench.ts`: `BenchStage = 'config'|'observe'|'results'`;
  `stageAfterCampaign(prev, running)`; `stageAfterReplay(prev, replay)`;
  `sectionOpen(stage, which)` (progressive reveal).

- [x] Step 1: tests first (red) — summary compare refuses a different site;
  delay band boundaries; remediation uses only valid P2 cells; ribbon never
  invents a count; stage machine.
- [x] Step 2: implement → green.
- [x] Step 3: commit — `ui: libs pures v2 — résumé d'audit, langage opérateur, ribbon, étapes du banc`

## Task 2 — root audit session and Auditor view

**Files**
- Create `web/frontend/src/store/audit.ts`
- Create `web/frontend/src/views/AuditView.tsx`
- Create `web/frontend/src/components/AuditDeclare.tsx` (form + presets + advanced)
- Modify `web/frontend/src/views/CampagneView.tsx` (remove audit state, card, polling)
- Modify `web/frontend/src/lib/sheet.test.ts` (no audit assertion)
- Reuse `components/AuditLive.tsx`, `AuditReportCard.tsx`, `AuditHistoryModal.tsx`

**Interfaces**
- Store owns: `stage`, form fields (site/link/provider/duration/direction/
  target/iface/rosIface), `advanced`, `presets`, `presetNote`, `historyN`,
  `msg`, `live`, `report`, `previous`, `comparison`; actions `open`, `close`,
  `restart`, `applyPreset`, `start`, `stop`, `sync`, `loadPresets`, setters.
- Polling: 2 s on `/api/audit/live` + `/api/audit/status`, timer at module
  scope, stops only on finish/cancel/fetch failure. Finish computes the
  summary, loads the previous one, saves the current.
- `restore()` at boot: `/api/audit/status`; if no in-memory report, one
  `/api/audit/report` call for the archived last report. This feeds the
  dashboard and the ribbon.
- `AuditView` renders Déclarer → Observer (`AuditLive`) → Lire (operator
  summary + `Détails scientifiques` drawer holding `AuditReportCard`) and the
  history modal. Presets never auto-launch.
- `CampagneView` keeps campaign + matrix + profile import only.

- [x] Step 1: store + view extraction; typecheck.
- [x] Step 2: tests green (update `sheet.test.ts`); no audit polling remains in
  `CampagneView`.
- [x] Step 3: commit — `ui: session d'audit à la racine + vue Auditer (suivi continu, rapport, historique)`

## Task 3 — four-mode shell

**Files**
- Modify `web/frontend/src/store/ui.ts` (`MODES`, `ModeId`, `mode`, `setMode`,
  `benchStage`, `setBenchStage`; remove `panel`/`PANELS`/`railPinned`)
- Create `web/frontend/src/components/ModeRail.tsx`
- Create `web/frontend/src/components/StatusRibbon.tsx`
- Create `web/frontend/src/components/ScienceDetails.tsx`
- Create `web/frontend/src/views/DashboardView.tsx` (first composition of the
  current Synthèse content, rewritten in Task 4)
- Create `web/frontend/src/views/BenchView.tsx` (three sections composing the
  existing campaign/live/results surfaces)
- Create `web/frontend/src/views/EvidenceView.tsx` (composes `IntegriteView`)
- Modify `web/frontend/src/App.tsx` (mode rail, ribbon, four sections,
  keyboard 1–4, no WebGL, no PanelChooser/Rail)
- Modify `web/frontend/src/components/CommandPalette.tsx` (modes + actions)
- Modify `web/frontend/src/components/NavIcon.tsx` (dashboard/audit/bench/
  evidence icons)
- Delete `components/PanelChooser.tsx`, `components/Rail.tsx`,
  `components/WebGLMesh.tsx`, `views/SyntheseView.tsx`, `lib/mesh.ts`
- Modify tests: `lib/designFloor.test.ts`, `lib/meteolink.test.ts`,
  `lib/beam.test.ts`, `views/LiveView.test.ts` (unchanged),
  delete `lib/mesh.test.ts`, `components/Rail.test.ts`; create
  `components/ModeRail.test.ts`

**Interfaces**
- `MODES = [{id:'dashboard', label:'Tableau de bord', key:'1'},
  {id:'audit', label:'Auditer', key:'2'}, {id:'bench', label:"Banc d'essai",
  key:'3'}, {id:'evidence', label:'Preuves & Rapports', key:'4'}]`.
- `ModeRail`: one source `MODES`, `aria-pressed`, keyboard hint, pinned labels
  persisted `mode-rail-pinned`; horizontal under 899 px, bottom bar under
  700 px.
- `StatusRibbon`: one-shot `/api/health`, `/api/integrity`, audit store state;
  refreshed on reconnect and after an audit; `—` on failure + retry; shows
  observation/full, SSE, “Audit en cours”, last audit site/grade, corpus
  counts, seal `hash8`, `Vérifier` → evidence.
- App: first-visit mounting per mode; audit store `restore()` once; campaign
  end reveals Résultats only when the bench is the active mode.
- `ScienceDetails`: `<details>`-style disclosure labelled `Détails
  scientifiques`, styles in CSS, no hex.

- [x] Step 1: store + shell + placeholder views; update tests (red→green).
- [x] Step 2: `npm run typecheck && npm test`; keyboard and mode switch
  smoke-checked in the source tests.
- [x] Step 3: commit — `ui: quatre modes — rail, ribbon d'état, coquille v2 sans WebGL`

## Task 4 — dashboard: real link vs simulated remediation

**Files**
- Rewrite `web/frontend/src/views/DashboardView.tsx`
- Modify `web/frontend/src/styles/index.css` (dashboard classes; drop `.syn-*`)
- Modify `web/frontend/src/lib/synthese.ts` + test (next step returns a mode)
- Test `web/frontend/src/views/DashboardView.test.tsx` (SSR DOM, no browser)

**Interfaces**
- **État du lien réel — audit**: last report from the audit store;
  `Délai sous charge` = `http_small_p95_ms` when measured else
  `rtt_loaded_p95_ms`, with `delayBand`; `Alertes à l'heure` =
  `cap_sla.ok_pct` when measured; `Qualité du lien` = grade + word; label
  « Mesuré sur le lien réel, sans modification »; drawer with p50/p95 idle and
  loaded, Δ, drain, QDI, JFI, loss, rows (n).
- **Remédiation recommandée — simulation**: falsifiable P2 comparison from
  `/api/results` (`remediationView`), label « Simulation de remédiation sur
  profil P2 — pas une mesure du lien »; simple vs best deadline rate, small
  p95 medians, n; drawer with qdisc/cc and hash; link to the leaderboard.
- **Ce que ça coûte**: `Perte financière : X Ar/h` from
  `economics.primary`, retail bound from the `emergency` scenario, verbatim
  disclaimer, badge « scénario contractuel »; `non mesuré` when absent.
- **Ce qui a changé**: `compareAudits(previous, current)` — local only;
  absent/different site ⇒ « Aucune comparaison disponible ».
- **Preuve gelée**: counts + seal + updated (ribbon also shows them).
- **Prochaine action** + four shortcuts with `setMode`.
- If no audit: “Aucun audit sur ce poste” + audit CTA; if no P2 groups:
  “Aucune campagne gelée — la simulation n'a pas de données.”

- [x] Step 1: failing DOM test — separate labels, no `twin`/`gelé` primary
  text, cost from `economics.primary`, change only with same site/link.
- [x] Step 2: implement → green; `npm run typecheck && npm test`.
- [x] Step 3: commit — `ui: tableau de bord — lien réel, simulation P2, coût et changement séparés`

## Task 5 — bench: one linear workspace

**Files**
- Create `components/BenchConfig.tsx` (campaign form + profile import +
  audit→profile; no audit run)
- Create `components/BenchProgress.tsx` (matrix + G0–G7 gates + timeline from
  the old CampagneView)
- Create `components/DisplayControls.tsx` (metric/chart/source, dispatches
  `panel-chooser-tri`; replaces PanelChooser)
- Rewrite `views/BenchView.tsx` (anchor nav `Configurer · Observer ·
  Résultats`, progressive reveal via `benchStage`)
- Modify `views/LiveView.tsx` (render `DisplayControls` in Observer; no other
  behaviour change)
- Modify `views/ResultatsView.tsx` (optional `embedded` prop hides its own h1)
- Delete `views/CampagneView.tsx`; replace `lib/sheet.test.ts` with
  `lib/benchLayout.test.ts`
- Modify `lib/beam.test.ts` (DisplayControls instead of PanelChooser)
- Modify `styles/index.css` (bench sections, `#wall` grid scoped, cockpit
  sheet removed, `#v-campagne` → `#v-bench`)

**Interfaces**
- `BenchConfig` keeps every campaign control: profiles P1–P4, qdisc list,
  CC list, repetitions, flows (default 4), clean/bursty gate, deadline, cost
  preview, DÉMARRER/ARRÊTER, custom profile import, audit→profile.
- `BenchProgress` keeps matrix order/progress, G0–G7 semantics, honest
  `attente` when not running.
- `BenchView` reveals Observer on campaign start/replay; Résultats on finish
  or explicit open; anchor buttons scroll to sections; no nested tab.
- Observer = banner + hero charts + metric bento + edge control + journal
  (existing `LiveView`) + progress/gates.
- Résultats = existing leaderboard/TradeSpace/compare/export with
  `embedded`.

- [x] Step 1: source tests (bench sections present, no `campagne`/`live`
  panels, no fixed 380 px cockpit, tri event consumed).
- [x] Step 2: implement → green; typecheck + tests.
- [x] Step 3: commit — `ui: banc d'essai linéaire — configurer, observer, résultats sans sous-onglets`

## Task 6 — Evidence: reports and citations

**Files**
- Create `components/AuditReportsPanel.tsx`
- Modify `views/EvidenceView.tsx` (proof first, then reports)
- Modify `views/IntegriteView.tsx` (optional `embedded` to skip its h1)
- Test `components/AuditReportsPanel.test.tsx` (SSR: list, empty, report render)

**Interfaces**
- List from `/api/audit/list`; open one via
  `/api/audit/report?audit_id=`; render `AuditReportCard` with copy/citation;
  Markdown link; close returns to the list. No new endpoint.
- Empty list: “Aucun rapport d'audit archivé.”

- [x] Step 1: failing test (empty + one row + citation surface).
- [x] Step 2: implement → green.
- [x] Step 3: commit — `ui: preuves — rapports d'audit et citations dans la destination Preuves`

## Task 7 — two-layer language pass

**Files**
- Modify `views/AuditView.tsx` (operator summary: Délai sous charge, Alertes à
  l'heure, Qualité du lien, Perte financière; full report in the drawer)
- Modify `components/AuditReportCard.tsx` (operator-first section titles:
  `Constat` → `Ce que ça donne`, keep scientific detail inside the existing
  sections; no removed number)
- Modify `views/ResultatsView.tsx` (plain column tooltips, scientific
  drawer around IQR/CI/method, no formula as primary label)
- Modify `views/LiveView.tsx` (label `QDI` → `Instabilité du lien` primary,
  `QDI` in the Explain tooltip; `Délai sous charge` on the small-object card)
- Modify `views/DashboardView.tsx` if any jargon remains.

- [x] Step 1: source/DOM assertions on the plain labels and the drawer.
- [x] Step 2: implement → green.
- [x] Step 3: commit — `ui: deux niveaux de langue — opérateur d'abord, détails scientifiques repliés`

## Task 8 — academic title, docs, ledgers, gates

**Files**
- Modify `PRODUCT.md` (title + subtitle, ignored by git)
- Modify `book/content/00-front.js` (titlePage title), `book/content/00-couverture.xml`
  (two title runs), `book/build-docx.js` metadata title/description (working
  tree only; never staged)
- Modify `docs/superpowers/plans/README.md` (row 15 + evidence)
- Modify `docs/PROJECT-SLICES.md` (slice 15)
- Modify `docs/VERIFY.md` (UI rows), `docs/CHECKPOINTS.md` (checkpoint)

**Title (verbatim)**
> Optimisation du transport réseau et garantie d’échéances temps réel pour la
> diffusion d’alertes hydrométéorologiques sur des liens d’accès contraints

Subtitle:
> Régulation du bufferbloat par gestion active de file (AQM), contrôle de
> congestion BBR et application au système d’information de Météo Madagascar

- [x] Step 1: apply the title; `python book/tools/audit-archive.py` and
  `node book/build-docx.js`; `memoire.docx` hash unchanged.
- [x] Step 2: final frontend gates — `npm run typecheck`, `npm test`,
  `npm run build`; `go test ./...`; `python tools/verify.py` (six gates).
- [x] Step 3: screenshots of the four modes and the three bench sections
  (`node capture-fresh.mjs` if the script targets the new shell; otherwise a
  fresh capture script, not a claim).
- [x] Step 4: ledger rows + checkpoint; commit explicit paths (no `book/`,
  no `PRODUCT.md`, no artifact).

## Task 9 — UX, performance and encapsulation polish (2026-10-30)

- [x] **Lazy bench chunk**: `App` imports `BenchView` with `React.lazy` +
  `Suspense`; hover/focus on the bench destination prefetches it. Initial JS
  fell from **344.2 KB gz in one chunk** to **111.1 KB gz initial + 234.2 KB gz
  bench chunk** (total 345.4 KB gz). Shell readiness measured 2,171 → 1,265 ms,
  DCL 1,264 → 648 ms, LCP 2,208 → 1,952 ms on this host.
- [x] **Hidden destinations stop working**: `LiveView`, `BenchProgress`,
  `ResultatsView`, `IntegriteView` and the dashboard receive an `active` prop;
  the live rAF loop, SSE re-renders and 10 s polls stop while hidden. Only the
  root audit session continues, by design.
- [x] **One header, one summary**: `ModeHead` for Auditer / Banc / Preuves;
  `OperatorStats` and `operatorSummary(report)` shared by the dashboard and the
  audit view (no duplicated thresholds or labels); `relativeTime` for archive
  age.
- [x] **Empty and recovery paths**: first-run dashboard block (3 steps), bench
  matrix folded as a preview at rest, results `RÉINITIALISER` for filters /
  rank / pins.
- [x] **Accessibility**: `aria-current` on the active destination, polite live
  region for audit phases, accessible names on chart surfaces, ≥32 px controls
  and ≥40 px on phones.
- [x] **Detector clean**: two intentional exceptions documented inline
  (Inter pinned by `DESIGN.md`; identity wordmark gradient). No other findings.
- [x] **QA tool**: `perf-probe.mjs` measures shell, LCP, DCL, transfers and
  console errors; `benchLayout.test.ts` pins lazy loading and hidden gating.

## Task 10 — design system, product and book refresh (2026-10-30)

- [x] `DESIGN.md` describes the four-mode shell, ModeHead/OperatorStats/
  ScienceDetails, the no-hidden-work rule and the motion/a11y floor.
- [x] `PRODUCT.md` UI row states the four modes, root audit session, real vs
  simulated split and the code-split performance.
- [x] Book: Table 13 becomes the four modes; Annex E rewritten with five new
  v2 captures (`book-tableau`, `book-auditer`, `book-banc-configurer`,
  `book-banc-resultats`, `book-preuves`); Table/Figure numbers preserved.
  `audit-archive`, `audit-biblio`, `audit-renvois`, `audit-nombres` and
  `build-docx` green; `memoire.docx` hash unchanged.

## Task 11 — quality of life and runtime QA (2026-10-30, builder pass)

- [x] `GET /api/audit/list` carries `has_report` per row (`auditReportExists`,
  `pkg/api/server.go`): the evidence panel says `non archivé` before the click
  instead of letting the first click answer 404. Six of nine archived rows are
  openable; the three that precede folder archiving are named. No CSV schema
  change, no new endpoint.
- [x] Opening the bench results leaves the click task: `startTransition` wraps
  the stage/section state, the scroll stays urgent. Bench entry peak task
  138 ms → 87 ms; the results mount measures 184–246 ms across runs and is now
  a second task, not the click's own.
- [x] `qa-walk.mjs` marks each step in page time and attributes every long
  task to the gesture that produced it; adds a 400 ms budget step and asserts
  the declarative report contract (`ouvrables + non archivées = lignes`,
  first `OUVRIR` opens with zero 404). Final walk: 19 steps, 0 failures,
  0 console errors, 0 resource errors, peak task 276 ms (boot, before the
  first mark), 10 long tasks in the whole walk.
- [x] `perf-probe.mjs` on the rebuilt bundle: shell 1050 ms, DCL 437 ms,
  LCP 1296 ms, bench destination 551 ms, 0 console errors. Initial chunk
  111.5 KB gz, lazy bench chunk 234.2 KB gz, total 345.6 KB gz.
- [x] QoL: `DÉMARRER` disabled with an explanation when no profile is
  selected; the duration field states the phase plan (mirror of
  `pkg/audit` `phasePlan`) instead of “durée libre dans avancé”; results
  reset tooltip typo fixed.
- [x] Abstraction: `useModeActive(mode)` replaces the four duplicated
  `s.mode === …` selectors; `main` takes focus on mode change so the
  destination is announced (`tabIndex={-1}`, `outline: none`).
- [x] Gates on the final tree: 46 files / 209 tests, `tsc --noEmit` clean,
  `go build ./... && go test ./... -count=1` clean (two new API tests),
  `python tools/verify.py` 6 passed / 0 failed / 0 not run.
- [x] Screenshots recaptured for the four modes and the three bench sections
  at 1920/1366/390.
- Unproven: the 184 ms results mount is still one React unit — splitting the
  ECharts subtree out needs the init effect to re-run on deferred mount, which
  is a separate change; the Linux/eBPF acceptance of slice 15 stays open.

## Task 12 — first chunk and the cost of opening results (2026-10-30, builder pass)

- [x] Only the landing destination ships in the first render: `AuditView` and
  `EvidenceView` join `BenchView` behind `lazy()` with one `Suspense` boundary
  per section (a single boundary would let a suspending view blank the others).
- [x] `lib/preload.ts` becomes a table of importers keyed by mode; `ModeRail`
  prefetches on hover/focus and `store/ui.setMode` prefetches on every mode
  change, so the keyboard and the command palette get the same warm start as
  the mouse. One place names the chunk: a new mode cannot be forgotten.
- [x] Initial chunk 111.5 → 99.1 KB gz (350.1 → 306.5 KB raw); new chunks
  `AuditView` 4.7 KB gz and `EvidenceView` 6.4 KB gz plus four shared UI chunks;
  total 349.9 KB gz against the 600 KB budget; JS heap 8.3 → 7.6 MB.
- [x] The TradeSpace card mounts one paint after the ranking (`chartsReady`) and
  reserves its height while waiting (`.ts-loading`, 396 px), so opening the
  results neither blocks the click nor jumps the page; its init effect now lists
  the mount flag so the chart cannot be left uninitialised.
- [x] Measured on the served build: results-anchor interaction task 184–463 ms
  before → 163–232 ms over six runs; bench entry 52–97 ms; `qa-walk` adds
  “le TradeSpace est réellement peint” (canvas 880×320 px) so a deferred mount
  cannot pass as a performance win while broken. Final walk: 20 steps,
  0 failures, 0 console errors, 0 resource errors.
- [x] `perf-probe` medians over five runs on the split build: shell 962 ms,
  DCL 378 ms, LCP 1,180 ms (single-sample pass before the split: 1,050 / 437 /
  1,296 ms). This host spreads ±60% run to run — one split run measured shell
  1,515 ms and bench reveal 1,705 ms — so the defensible claims are the chunk
  and heap figures, not the timings.
- [x] Gates on the final tree: `tsc --noEmit` clean, 46 files / 209 tests,
  `go test ./pkg/api/ -count=1` ok, `python tools/verify.py` → 6 passed,
  0 failed, 0 not run; screenshots recaptured at 1920/1366/390.
- Unproven: the boot task (210–310 ms, React init and first render) is still
  the peak and is untouched by this task; the four extra chunk requests are
  unmeasured on a real link, only on localhost.


## Acceptance mapping

| Spec acceptance | Task |
|---|---|
| Four modes at 1920/1366/390, keyboard 1–4, no nested tabs | 3, 5 |
| Real vs simulated separate, plain labels, no `twin`/`gelé` primary | 4 |
| Audit survives a mode switch, no false “suivi interrompu” | 2, 3 |
| Bench preserves every control, no sub-tabs | 5 |
| Operator labels plain, science in drawer | 4, 7 |
| Evidence lists reports/citations without new endpoints | 6 |
| Tests/typecheck/build, six gates, screenshots | 8 |
| Title in PRODUCT/cover/title page/header/metadata; book audit/build green | 8 |
| No live audit, no API/CSV/corpus change | all |

## Risks and mitigations

- **Mode shell breaks source tests** (Rail/PanelChooser/WebGL assertions):
  update or delete those tests in the same commit as the removal; never leave
  a test asserting a component that no longer exists.
- **Audit polling leaks** across StrictMode remounts: timer at module scope,
  one poller, cleared on finish/cancel; a mode switch never clears it.
- **Bench page length**: progressive sections + anchors, not nested tabs.
- **Book title touches Word XML**: edit the two runs + the JS title; rebuild,
  verify `memeire.docx` hash unchanged and the old title absent.
- **Bundle budget**: no new dependency; WebGL removal offsets the new views.
