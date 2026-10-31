# CGO — Operator console and engine verdict — design

2026-10-31 · branch `rebuild/plan1-economics-audit` · status: approved in sections, awaiting spec review
Sources: `C:\INTERFACE\CHANGES.md`, `DEEPCHANGES.md`, `MORECHANGES.md`; `C:\IDEA\directive.md`, `IDEA1.md`, `IDEA2.md`, `IDEA3.md`, `Untitled 37.md`; three reviewer passes on this design.

## 1. Problem

The built interface exposes engine internals to a person who has none of the
engine's mental model. Measured facts, not opinions:

- The dashboard offers four entry points to the same intent: mode rail, “Auditer
  de nouveau”, “Prochaine action”, footer shortcut strip (`DashboardView.tsx:75-80`,
  `App.tsx:132-161`).
- Six “absence” sentences occupy a full band when cost and comparison are empty
  (`DashboardView.tsx:194-199`).
- Labels are uppercase monospace French sentences; the page mixes a classical
  serif, a sans, and a mono (`tokens.css:5-9`, `tokens.css:64-66`).
- The bench exposes the matrix axes, and the client hardcodes them
  (`BenchConfig.tsx` before commit 49b5f15).
- There is no single operator verdict. The words are computed in TypeScript
  against adjustable settings (`lib/dashboard.ts:15-75`), and the `LIEN` score is
  normalized against whatever groups are visible (`lib/lien.ts:9-45`), so neither
  is an absolute statement about a link.

## 2. Scope and personas

Primary user: a layman operator at Météo Madagascar. They must answer “is this
link fit to send alerts, and what do I do” without reading a manual. Secondary:
network engineer and thesis jury, served by the proof drawer and the research
mode, not by the default surface.

In scope: shell, typography, surfaces, dashboard, audit, bench, evidence, the
verdict engine, the report fields the UI needs, the live timeline, the duel
endpoint, the campaign-frozen tier, the TUI mapping, DESIGN.md, PRODUCT.md, book
Annex E and F.

Out of scope:
- Production monitoring. `sadcnwp` / 41.207.49.27 appears in IDEA1–3 and
  Untitled 37; `AGENTS.md` forbids touching that host and no production access
  exists. The console measures the link the operator declares, on the machine
  running the binary. No screen may imply it watches the production server.
- `directive.md` §7 “do NOT rebuild or redesign the user interface”. This design
  overrides it by explicit human instruction, and keeps its required additions:
  telemetry tier, capacity-waste rate, WIS 2.0 SLA, exact RouterOS and Linux
  commands including ingress.
- BBRv3 live data, IFB live acceptance, `.dist` field trace. They stay blocked;
  the UI renders availability, never a prediction.

## 3. Architecture

### 3.1 Shell

- Desktop: one 64px header. Wordmark left (Inter, tricolor M), four
  sentence-case destinations with key hints `1–4`, status cluster right
  (connection dot, mode, last audit age, settings). The 232px mode rail and the
  footer shortcut strip are removed. The footer keeps one 28px provenance line.
- Mobile under 700px: header with wordmark and status; destinations become a
  bottom tab bar. No duplicated navigation.
- Content column max 1120px, 24px gutters, one column per destination.
- Views keep the current lazy-mount and root-SSE behaviour; navigation is
  presentation, not lifecycle.

### 3.2 Verdict engine (Go)

New in `pkg/metrics` (pure, no I/O), rendered into the audit report as
`verdict`:

```go
type VerdictLevel string
const (
    VerdictNeutral VerdictLevel = "neutral"
    VerdictRed     VerdictLevel = "red"
    VerdictAmber   VerdictLevel = "amber"
    VerdictGreen   VerdictLevel = "green"
)
type Verdict struct {
    Level    VerdictLevel `json:"level"`
    Word     string       `json:"word"`
    Sentence string       `json:"sentence"`
    Basis    VerdictBasis `json:"basis"` // grade, sla_measured, ok_pct, delta_ms, load_established
}
```

Cascade, strictly ordered, first match wins. Inputs are `grade`, `cap_sla`,
`load_established` (the engine's existing `load_measured`) and a new
`load_attempted` recorded in `measured`, which says the audit was configured with
a load window:

| # | Gate | Level | Word |
|---|---|---|---|
| 1 | load not attempted | neutral | Pas encore mesuré |
| 2 | load attempted, bulk never established | neutral | Charge non établie |
| 3 | `cap_sla.measured && cap_sla.breach` | red | Alertes en retard |
| 4 | load established, grade empty or delta not finite | neutral | Note indisponible |
| 5 | grade D or F | red | Saturation critique |
| 6 | grade C | amber | Lien fragile |
| 7 | grade A+, A or B | green | Lien sain |

A measured mission failure outranks a missing grade: gate 3 precedes gate 4, so
an audit that proves the alert target was missed is never softened to
“indisponible”.

Rules that make this safe:
- Gate 6 and 7 read the same `bloat_grade` the “Qualité du lien” fact displays, so
  the verdict and the fact cannot disagree. The earlier `delta > 50 ms` amber is
  deleted; it turned a grade B into an amber alert.
- No gate may return green without a grade. The reviewer's hazard was reachable:
  `capSamples` pools the load-window small probes (`pkg/audit/audit.go:580`), so a
  bulk that transfers zero bytes while probes succeed yields a passing SLA with no
  grade. Gate 4 catches it.
- Thresholds are the existing ones only: the Waveform bands
  (`pkg/metrics/bufferbloat.go:11`) and `CAPDeadlineMs`/`CAPBreachThresholdPct`
  (`pkg/metrics/cost.go:81-82`). The verdict reads `cap_sla.breach`, never a
  re-typed 95.0.
- Sentences (French, fixed strings):
  - Charge non établie — “Le flux saturant n'a pas pu démarrer ; la tenue sous
    congestion n'est pas mesurée.”
  - Pas encore mesuré — “Aucun test sous charge n'a été réalisé sur ce lien.”
  - Note indisponible — “La charge a tourné mais l'écart de latence n'a pas pu
    être calculé.”
  - Alertes en retard — “X % des alertes passent sous 220 ms (objectif 95 %).”
  - Saturation critique — “Le lien sature sous charge (note D) ; les
    télétransmissions risquent d'échouer.”
  - Lien fragile — “La latence monte nettement sous charge ; les alertes peuvent
    partir en retard.”
  - Lien sain — “Sous charge, la latence reste sous l'objectif” plus, when the
    alert probe ran, “et X % des alertes passent à l'heure”.
- The report Markdown prints the same verdict line, so the sealed artifact, the
  web UI and the TUI never diverge.

### 3.3 Recommendation card

```
Protéger ce lien
Régulation active, calibrée sur ce qui a été mesuré (95 % du goulot).

Envoi (montant)     4,8 Mbit/s   sur 5,1 mesurés      script prêt
Réception (desc.)   17,6 Mbit/s  sur 18,5 mesurés      IFB actif

Plateforme   (•) MikroTik RouterOS 7    ( ) Linux tc
[ Copier la configuration ]              Voir la commande exacte ▸
```

- The engine already sizes each direction separately and records a missing
  direction as unavailable (`pkg/audit/report.go:337-395`). The card shows both
  lines; an unmeasured direction reads “non mesurée — auditer le sens descendant”
  and contributes no script.
- New report fields so the UI never parses a command string:
  `up_measured_mbps`, `up_shaped_mbps`, `down_measured_mbps`, `down_shaped_mbps`,
  `linux_iface`, `routeros_wan_iface`, `routeros_lan_iface`. Shaped rate is
  `0.95 × measured` (`pkg/remediate/remediate.go:56`), stated on screen.
- The platform toggle lists only platforms the report carries, persists the last
  choice, and is a `radiogroup`. Copy reads the active platform and the measured
  directions, and the toast names what was copied.
- RouterOS has no IFB. When the ingress script exists, the card and the drawer
  name the LAN interface actually used and add the line “RouterOS façonne la
  réception sur l'interface locale (LAN ou bridge)”.
- Half-remediation is refused. Today `RouterOSIngress` returns
  `ErrUnmeasured` when no LAN interface is declared
  (`pkg/remediate/remediate.go:314`), so an operator who never opened the
  advanced field copies an upload-only script and believes the link is fixed
  while the next 200 MB downlink bloats it. New behaviour: when the download
  window was measured and the LAN interface is undeclared, the generator emits
  the ingress script against `bridge` and leads it with a warning block —
  “ATTENTION : remplacez « bridge » par votre interface LAN si différent —
  vérifiez avec /interface print brief”. The report records
  `routeros_lan_assumed: true`, the card reads “Réception 17,6 Mbit/s ·
  interface supposée « bridge » — à vérifier avant de coller”, and the drawer
  repeats the assumption. An assumed name is visible and checkable; an absent
  script is neither.

### 3.4 Proof drawer, two tabs

Sheet at 85vh, max 720px on desktop, full width on mobile, focus-trapped,
`Escape` closes, opened to the requested tab.

- **Déploiement & scripts** — selected platform's exact commands per measured
  direction, in `<pre>` with per-block copy; sizing basis (measured rate, 0.95
  factor, idle RTT, qdisc name); rollout and verification commands
  (`/queue type print`, `/queue tree print`, `tc qdisc show`); the RouterOS rejeu
  line worded “validée par rejeu”, “non validée”, or “aucune preuve”, never
  “appliquée”.
- **Preuves & télémétrie brute** — p50/p95 idle and loaded, worst delta, QDI,
  loss, throughput, `n`, direction; sensor tier (eBPF, TCP_INFO, standard) with
  its reason; source and telemetry SHA-256; link to the sealed report
  (`/api/audit/report?audit_id=…`); the 4G reference duel with its run IDs;
  BBRv3 availability from `/api/capabilities`.

TUI mapping: tabs stay five (`cmd/cgo/tui.go:33`). Tabs 4 and 5 print a one-line
verdict header and the four facts as four lines. Keys `d` and `p` open the two
panes full screen, text wrapped at 78 columns, one command per line, no
horizontal scroll. `c` writes the selected script to
`data/audits/<audit_id>/remediation.routeros.rsc` or `remediation.linux.sh`, next
to the sealed report, and prints the path in the status line. It attempts OSC 52
as well and says plainly when the terminal refused it. Nothing is written to
`/tmp`.

### 3.5 Live timeline and canvas

`/api/audit/live` currently ships the engine's capped per-window ring
(`pkg/api/server_audit.go:244`), which is why a client-side chart can only show
the last few seconds. The audit progress gains a full-run, per-second series:

```json
"timeline": [{"t": 0, "phase": "baseline", "n": 5, "rtt_p50": 24.1,
              "rtt_p95": 31.0, "small_p95": null}, …]
```

- One bucket per elapsed second for the whole run, capped at `total_sec + 5`.
  `n` is the number of probes that answered in that second; a value is `null`
  when nothing answered. A null is never coerced to 0.
- The canvas breaks the line at every null bucket and marks the gap, so a VSAT
  retransmission timeout reads as missing data, not as perfect latency. A
  single-pixel legend line states “sonde sans réponse”.
- The Observer canvas is 120px tall, single `<canvas>`, X axis 0 → `total_sec`,
  Y range always covering both 220 ms and the highest observed value, with the
  220 ms objective drawn as a labelled rule. Phase bands shade baseline, load and
  drain. Points are the per-second loaded latency, so the step from idle to
  loaded is still visible at second 30.
- Redrawn on each poll, no per-sample DOM nodes, no ECharts, no animation.
- Accessible name plus the three live facts as its text alternative.
- During the run the facts are temps de réponse, réponse sous charge against the
  220 ms objective, and débit; the percentile wording stays inside the drawer.
  The alert percentage appears only after the freeze, because the engine computes
  it at the freeze.

### 3.6 Economic basis, and the tier where it actually lives

Correction of record, made after checking the consumers: the audit's Ariary
figure never reads a price tier. `BuildEconomics` composes model A from the two
enterprise plans, model B from the two B2B tariffs, and the emergency bound from
`RetailEmergencyARPerGB`, then marks the primary by declared link type
(`pkg/metrics/cost.go:138-196`). `DefaultTier` is read only by the campaign
(`pkg/campagne/campagne.go:429` and `:631`). An audit-scoped tier selector would
therefore control a number the audit does not use, so it is dropped.

What the audit needs is a numeric basis, so the UI never parses a French sentence
to explain a figure. Added to `EconomicsScenario`:

```go
type EconomicsNumbers struct {
    MonthlyOpExMGA float64 `json:"monthly_opex_mga,omitempty"` // model A
    CIRMbps        float64 `json:"cir_mbps,omitempty"`         // model A
    ARPerGB        float64 `json:"ar_per_gb,omitempty"`        // model B, emergency
    RetransRatio   float64 `json:"retrans_ratio"`
    WastedBytes    float64 `json:"wasted_bytes"`
    LoadSeconds    float64 `json:"load_seconds"`
}
// EconomicsScenario gains: Numbers EconomicsNumbers `json:"numbers"`
```

- Model A is never forced through volume arithmetic: it reports the monthly
  charge and the CIR. Models B and emergency report the per-gigabyte price. The
  reviewer's `cir_depreciation` / `metered_volume` labels are not adopted, because
  the engine and the book already speak A, B and emergency.
- The operator's real lever is the declared link type, already a primary field of
  the audit form and already the input to `primaryName`. The card states it:
  “scénario retenu pour un lien 4G : APN entreprise”. The drawer lists every
  scenario with its numbers, as it does today.
- When the waste cannot be valued the fact reads “Non chiffré”, never zero.

The mutation hazard is real, but it lives in the campaign. `POST /api/cost/tier`
changes the global `DefaultTier`, and the campaign reads it per event at the
freeze, so one run can carry two cost bases. Fix:

- `buildMatrixConfig` freezes the active tier into the run config, next to
  profiles, qdiscs, ccs, reps, kernel and bench.
- The campaign cost reads the tier from its own run config, not from the global.
- `POST /api/cost/tier` returns 409 while a campaign runs, naming the run.
- The results proof and the duel proof name the tier frozen with the run, so a
  cost column is traceable to one tariff.

### 3.7 Duel endpoint

`GET /api/duel?profile=P2` returns the pair, the deltas and the evidence:

```json
{"profile":"P2","deadline_ms":220,
 "base":{"qdisc":"pfifo_fast","cc":"bbr","deadline_pct":48.8,"small_p95_ms":328.2,"n":22,"runs":["…"]},
 "best":{"qdisc":"fq_codel","cc":"bbr","deadline_pct":96.7,"small_p95_ms":209.0,"n":18,"runs":["…"]},
 "delta_deadline_pts":47.9,"delta_small_ms":-119.2,"min_reps":3}
```

- Valid-only cells, `n ≥ 3` for both sides, otherwise no duel and the card says
  “aucune base exploitable”.
- Winner is the highest valid `deadline_median`; ties break on lower
  `small_p95_valid_median`, then qdisc order `pfifo_fast < fq_codel < cake`, then
  cc name. The base is the `pfifo_fast` cell with the largest `n` among eligible
  cells.
- The browser subtracts nothing; the deltas come from the endpoint, so the web,
  the TUI and the report agree.

### 3.8 Bench

Framed as a lab, never as the operator's link.

- Default card “Duel de référence — profil P2 (4G chargée : 20 Mbit/s, 100 ms,
  perte 0,5 %)” rendering the `/api/duel` pair with protocol names, percentages,
  latency and `n` visible. The winner is whatever the data says.
- One action “Lancer la comparaison”, still arm-then-confirm. Preset: P2, 3 reps,
  4 flows, clean channel, objective 220 ms, full P2 matrix server-side; the UI
  curates the pair.
- Running: phase, cell 7/18, remaining, stop, and the same canvas pattern as the
  audit observer.
- “Mode recherche” (secondary, remembered) reveals the matrix, leaderboard,
  filters, TradeSpace, cost columns, burst test, and BBRv3 when the kernel
  exposes it.

### 3.9 Preuves & Rapports

- Four counts at 32px, one line “sceau 523956be · maj …”, “Copier l'empreinte”,
  full SHA-256 inside the drawer.
- One primary “Vérifier l'intégrité”, answered in one line.
- Report rows: site, date, verdict word with its dot, “Ouvrir”, “Preuve”. No hex,
  no run IDs on the row.
- “Détails du corpus” folds the runs table, replay, quarantine, manifests.

## 4. Visual system

- One sans: Inter for every label, button and heading. JetBrains Mono only for
  numbers, units, IPs and commands. Cormorant leaves the app bundle; it stays in
  the book and in `pkg/figures` output.
- Scale: hero verdict 40px, metric value 32px, page title 28px, body 16px,
  label 14px, fine print 13px inside the drawer only. Buttons 16px on 48px
  targets. No uppercase French sentences.
- Surfaces: canvas `#0a0b0e`, card `#12141a`, elevated `#181b22`. Borders only
  for table rows, dividers and the focused element. Radius stays 0; no shadows,
  no gradients, no glass.
- Colour: green, amber and red carry verdict and status words with a dot, never
  colour alone. Cyan marks live and active navigation. Violet stays inside
  charts and research. The tricolor remains identity-only.
- Motion: 120ms fade on destination change; charts do not animate;
  `prefers-reduced-motion` collapses durations.
- Accessibility: 4.5:1 for text on the new surfaces, focus ring, landmarks,
  `radiogroup` for the platform toggle, `tablist` for the drawer, polite live
  region for the verdict and the audit phase, 48px touch targets.
- Print: the field sheet prints verdict, four facts, recommendation basis and the
  exact command; no chrome.

## 5. Language

| Surface | Drawer and report | Engine field |
|---|---|---|
| Alertes à l’heure | complétions 16 Kio sous 220 ms | `cap_sla.ok_pct` |
| Réponse sous charge | petit p95 chargé | `http_small_p95_ms` |
| Qualité du lien | note bufferbloat Waveform | `bloat_grade` |
| Capacité gaspillée | Ar/h, scénario et base numérique | `economics.primary` |
| Protéger ce lien | script exact par plateforme | `remediation` |
| Envoi / Réception | egress / ingress IFB | `direction` |
| Régulation active | CAKE, fq_codel, pfifo_fast | `qdisc` |
| Mesure exacte du noyau | eBPF, TCP_INFO, standard | `kernel_telemetry.source` |
| Lien 4G de référence | profil P2 | `profile` |
| Mesure archivée | run, manifeste | `run` |
| Empreinte de preuve | SHA-256 | `provenance` |

Banned on any default surface: p50, p95, qdisc, CAKE, fq_codel, BBR, CUBIC, JFI,
QDI, SLA, bufferbloat, eBPF, manifeste, corpus, gel, twin, run, hex strings,
`n=` (say “3 mesures”), and the words “garanti”, “garantie”, “prouvé” unless a
sealed proof is named next to them.

## 6. Honesty invariants

1. No green without a grade A+, A or B.
2. No command without measured capacity and idle RTT; no ingress script without
   a named LAN interface or a measured down window.
3. Every displayed rate shows measured and shaped values.
4. Every Ariary figure names its model, its scenario and its numbers. A bench
   cost column names the tier frozen with its run.
5. The duel winner is valid-only with `n ≥ 3` and cites run IDs.
6. BBRv3 appears only when `/api/capabilities` reports it.
7. The bench never describes the audited link; the simulation offers no
   production copy.
8. RouterOS wording is “validée par rejeu” or “non validée”, never “appliquée”.
9. Absent data is a word, never a dash or a zero.
10. Every number on a surface traces to the report JSON, `/api/results`,
    `/api/duel` or the live stream.
11. A name the operator did not declare is labelled as assumed, on the card, in
    the script comment and in the report field.

Each invariant gets a test in section 8.

## 7. Errors and empty states

- No report: the hero is the empty state, one sentence, one action.
- Bulk failed: neutral “Charge non établie” with the engine's reason; the audit
  action stays primary.
- No archived campaign: the duel card is replaced by one line and the bench
  action; no empty table.
- Integrity failure: one red line naming the gate and the file; no stack trace.
- Clipboard refused: the toast says so and points at the artifact path.

## 8. Testing and evidence

Go:
- `pkg/metrics` table test for all seven gates, including bulk-failed-with-passing-probes
  and grade-absent-with-SLA, asserting no green without a grade.
- `pkg/audit` report test: `verdict` present in JSON and Markdown; per-direction
  rate and interface fields; `economics.primary.numbers` carrying the model's own
  basis.
- `pkg/api` tests: `/api/duel` rejects `n < 3` and degraded cells, deterministic
  tie-break; `/api/audit/live` timeline length matches `total_sec` and never
  exceeds the cap; `POST /api/cost/tier` returns 409 while a campaign runs and
  names the run.
- `pkg/campagne` test: the run config freezes the active tier, and a row's cost
  derives from the frozen tier rather than the global one.
- `pkg/remediate` tests: a measured download with an undeclared LAN interface
  still yields the ingress script, naming `bridge` and carrying the warning line,
  with `routeros_lan_assumed` true; an unmeasured download yields no ingress
  script at all, assumed interface or not.
- Economics numbers test: model A records `monthly_opex_mga` and `cir_mbps` with
  no `ar_per_gb`; model B and the emergency bound record `ar_per_gb` with no
  monthly charge.
- Existing `/api/capabilities` test stays the guard for BBRv3 honesty.

Frontend:
- Verdict component: four levels, exact strings, no dash for absence.
- Recommendation card: toggle copies the selected platform only; unmeasured
  direction contributes nothing; rates match the report fields.
- Drawer: `tablist` semantics, focus trap, restore, independent scroll.
- Canvas: full-duration axis, phase bands, 220 ms rule, empty timeline renders
  the idle state, and a null bucket breaks the polyline with its gap marker — no
  unanswered probe may be plotted as 0 ms.
- Duel card renders the endpoint pair verbatim.
- Lexicon gate: a source test asserting no banned surface term appears in the
  four view components or their labels.
- QA walk (`qa-walk.mjs`): one primary action per destination, no horizontal
  overflow at 390/1024/1366/1920, drawer keyboard path, copy toast, verdict
  announced once per change.

Gates: `make verify` six gates, bundle budget re-measured after removing the
serif, captures at 1920/1366/390. No live bench run in this work.

## 9. Documentation

- `DESIGN.md` rewritten for the console: type scale, surfaces, navigation,
  verdict, drawer, lexicon, honesty rules.
- `PRODUCT.md`: persona order corrected to layman operator first; the verdict
  field, the duel endpoint, the economic numbers on the primary scenario and the
  campaign-frozen tier documented; the stale “no redesign in this wave” line
  removed; the economics paragraph made precise about which figure comes from a
  contract scenario and which from a mobile tier.
- Book Chapter III needs content edits, not a terminology bridge. The four
  destinations are already the ones this console keeps: `book/content/04-pratique.js:10`
  states “L’interface de supervision expose quatre modes — Tableau de bord,
  Auditer, Banc d’essai, Preuves & Rapports”, and Tableau 13 (`tbl412`) lists
  them. What is now wrong is what each row claims: the dashboard row describes
  three raw metrics instead of the verdict and four facts, the bench row
  describes the full matrix instead of the preset duel, and no row mentions the
  two-tab drawer. Those rows and the surrounding prose are rewritten in place.
- One live contradiction found while checking this: `04-pratique.js:73` says the
  shaped capacity is set at 90 % of the measured capacity, while the generator
  uses 95 % (`pkg/remediate/remediate.go:56`). The book text moves to 95 % with
  the reason, or states both figures with their scope. The code is not changed to
  match prose.
- Book Annex E rewritten screen by screen with the new captures; Annex F gains the
  duel eligibility rule and the rejeu wording. Annexes stay uncommitted per
  `AGENTS.md`.
- Ledgers: `VERIFY.md` rows for the verdict, duel, live timeline, economic
  numbers and the campaign-frozen tier as builder-pass with verifier pending;
  `CHECKPOINTS.md` entry; `ERRORS.md` entry for the false-green hazard (a passing
  alert probe with no grade) and for the hardcoded bench axes fixed in 49b5f15.

## 10. Phasing

Two slices, one dependency.
- Slice 20A, engine truth: verdict, report fields, economic numbers, live
  timeline, duel endpoint, campaign-frozen tier, Go tests, ledger rows.
- Slice 20B, operator console: shell, typography, surfaces, four destinations,
  recommendation card, drawer tabs, canvas, bench preset, evidence, lexicon gate,
  QA walk, DESIGN/PRODUCT and book rewrites. Depends on 20A.

## 11. Reviewer decisions

Accepted: kill the serif in the app; one primary action; surfaces instead of
borders; verdict in Go; platform toggle; restore the Ariary fact; split the
drawer; duration toggle; name protocols in the lab; restore a live trace; TUI
mapping; drop the 50 ms amber gate; per-direction rates; duel guardrail; file
fallback for TUI copy; no half-remediation on RouterOS; numeric economics fields
instead of a prose-only basis; gap-break rendering for unanswered probes.

Rejected with evidence: the claim that the RouterOS generator hardcodes `bridge`
(it refuses to generate without a named LAN interface,
`pkg/remediate/remediate.go:314`); the proposed audit-scoped tier selector,
because the audit never reads a tier at all — `BuildEconomics` works from
contractual scenarios keyed to the declared link type
(`pkg/metrics/cost.go:138-196`), and `DefaultTier` is consumed only by the
campaign (`pkg/campagne/campagne.go:429`, `:631`). The mutation hazard is real but
belongs to the campaign, and §3.6 fixes it there. The claim that the Ariary
figure was fabricated is also rejected: six real tiers exist
(`pkg/metrics/metrics.go:85`) and the scenarios are named in the report. What was
true and is accepted is narrower — the numbers behind a scenario were only
readable from a prose string. The reviewer's duel numbers and CAKE winner are
rejected because the corpus selects the winner from data. The earlier draft's
“SLA passes, therefore green, even with no grade” is removed by gate 4. `/tmp`
is rejected as the TUI export target. The formula `p95 loaded − p50 idle` is
wrong for this engine, which uses p95 − p95 (`pkg/audit/audit.go:440`). The
proposed thesis desynchronization is stale: Chapter III and Tableau 13 already
name the four modes (`book/content/04-pratique.js:10`). The proposed
`cir_depreciation` and `metered_volume` model labels are not adopted, because the
engine and the book
already speak A, B and emergency.

## 12. Standing blockers

Unchanged by this design: `make bpf` and eBPF attach on Linux 6.8, IFB live
acceptance, a field trace of at least 30 baseline samples for `.dist`, a 7.1
kernel for BBRv3, a booted CHR for live RouterOS acceptance, the operator grant
for the case study, production read access, and the merge to `main`.
