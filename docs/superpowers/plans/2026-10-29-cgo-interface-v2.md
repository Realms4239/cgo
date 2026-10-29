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

- [ ] Step 1: tests first (red) — summary compare refuses a different site;
  delay band boundaries; remediation uses only valid P2 cells; ribbon never
  invents a count; stage machine.
- [ ] Step 2: implement → green.
- [ ] Step 3: commit — `ui: libs pures v2 — résumé d'audit, langage opérateur, ribbon, étapes du banc`

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

- [ ] Step 1: store + view extraction; typecheck.
- [ ] Step 2: tests green (update `sheet.test.ts`); no audit polling remains in
  `CampagneView`.
- [ ] Step 3: commit — `ui: session d'audit à la racine + vue Auditer (suivi continu, rapport, historique)`

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

- [ ] Step 1: store + shell + placeholder views; update tests (red→green).
- [ ] Step 2: `npm run typecheck && npm test`; keyboard and mode switch
  smoke-checked in the source tests.
- [ ] Step 3: commit — `ui: quatre modes — rail, ribbon d'état, coquille v2 sans WebGL`

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

- [ ] Step 1: failing DOM test — separate labels, no `twin`/`gelé` primary
  text, cost from `economics.primary`, change only with same site/link.
- [ ] Step 2: implement → green; `npm run typecheck && npm test`.
- [ ] Step 3: commit — `ui: tableau de bord — lien réel, simulation P2, coût et changement séparés`

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

- [ ] Step 1: source tests (bench sections present, no `campagne`/`live`
  panels, no fixed 380 px cockpit, tri event consumed).
- [ ] Step 2: implement → green; typecheck + tests.
- [ ] Step 3: commit — `ui: banc d'essai linéaire — configurer, observer, résultats sans sous-onglets`

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

- [ ] Step 1: failing test (empty + one row + citation surface).
- [ ] Step 2: implement → green.
- [ ] Step 3: commit — `ui: preuves — rapports d'audit et citations dans la destination Preuves`

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

- [ ] Step 1: source/DOM assertions on the plain labels and the drawer.
- [ ] Step 2: implement → green.
- [ ] Step 3: commit — `ui: deux niveaux de langue — opérateur d'abord, détails scientifiques repliés`

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

- [ ] Step 1: apply the title; `python book/tools/audit-archive.py` and
  `node book/build-docx.js`; `memoire.docx` hash unchanged.
- [ ] Step 2: final frontend gates — `npm run typecheck`, `npm test`,
  `npm run build`; `go test ./...`; `python tools/verify.py` (six gates).
- [ ] Step 3: screenshots of the four modes and the three bench sections
  (`node capture-fresh.mjs` if the script targets the new shell; otherwise a
  fresh capture script, not a claim).
- [ ] Step 4: ledger rows + checkpoint; commit explicit paths (no `book/`,
  no `PRODUCT.md`, no artifact).

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
