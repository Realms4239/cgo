# CGO — Console spine — design

2026-11-02 · branch `rebuild/plan1-economics-audit` · status: awaiting spec review

Adopts `2026-10-31-cgo-operator-console-design.md` §3.1–§3.10 as the visual and
UX contract, with the amendments in §9. Supersedes slices 8, 14, 15, 18, 20A,
20B and 20C as *delivery vehicles*; their evidence is kept, their rows are
re-dispositioned in §7.

## 1. Problem

The interface is not the one the approved design describes, and the plan corpus
cannot say so. Measured on this checkout, 2026-11-02:

| Fact | Evidence |
|---|---|
| The console's verdict component is rendered by nothing | `web/frontend/src/components/VerdictHero.tsx` has no importer |
| The old navigation still ships | `components/ModeRail.tsx` is present and routed in `App.tsx` |
| One screen still offers three navigations | `ModeRail` + `.syn-actions` (`DashboardView.tsx:301`) + the footer shortcut strip |
| The primary action falls below the fold | `web/frontend/shots-console-2026-11-02/01-dashboard.png` at 1920×1080: “Auditer le lien” is clipped |
| The plan is 9 480 words for nine tasks and carries its own review history | `2026-10-31-cgo-operator-console-ui.md` |
| Five slices are marked in progress, plus 20B in flight — against a cadence limit of three | `docs/superpowers/plans/README.md` |
| 12 of 46 claims are `PENDING`, 8 `BLOCKED` | `docs/VERIFY.md` |

Diagnosis: the design was reviewed three times and is sound. The **delivery
contract** was not. It grew a review transcript, it sequenced invisible work
first (tokens before shell; a hero nobody mounts), and it had no acceptance
artifact for the one thing the operator judges — the screen.

## 2. Scope

Deliverable, in one sentence: a layman operator opens the console, reads one
verdict with four facts, and copies the exact command for their own router —
with every number behind it sealed and re-runnable.

**In scope:** the shell; the verdict hero and its four facts; the remediation
card; the proof drawer; the dashboard, audit and evidence destinations; the
live trace on the observer; the bench operator surface; the operator lexicon;
and the evidence tail (verifier packet, figures, book annex).

**Out of scope, each with its reason:**

| Excluded | Reason |
|---|---|
| Any new measurement | A live run needs the human's explicit word; none was given (§8) |
| eBPF (4), `.dist` (5), IFB (6), BBRv3 (7) live acceptance | Deferred with their slices and their named unblocks (§7) |
| The TUI (20C) | Built and tested; only its verifier row moves |
| New engine capability | 20A is done; this slice renders and proves what exists. The only engine edit is the two additive, sealed `cap_sla` fields of T2 |
| `book/` commits, the frozen corpus, CSV schemas | Standing constraints in `AGENTS.md` |
| Any production host | `sadcnwp` is out of scope by contract; no screen may imply it is watched |

## 3. The four rules

Each rule names the failure it kills.

**R1 — Visible first, wiring mandatory.** A task is done when a route renders
it. `VerdictHero` exists and no file imports it: that is the failure mode this
rule ends. Any component a task creates must appear in the destination list of
that task's acceptance run, or the task is not done.

**R2 — The plan is a contract, not a review log.** At most ten tasks, each at
most one page: files, the one assertion that could fail, the command, the
screenshot, the commit. Rejected alternatives, earlier drafts and review
findings go to the checkpoint, never into the plan. This is why the previous
plan reached 9 480 words for nine tasks.

**R3 — The screenshot is an acceptance artifact.** Every task that changes a
surface ends with `node capture-fresh.mjs` over the destination it touched at
1920×1080 and 390×844, saved under `docs/img/shots/<date>/` (the `hygiene` gate
rejects a tracked file above 384 KB outside `docs/img/`), plus four measured
gates in `qa-walk.mjs`:

1. the primary action of each destination is inside the first viewport;
2. no horizontal overflow at 1920, 1366 or 390;
3. zero console errors and zero resource errors;
4. no long task above 400 ms after a navigation (the existing check).

Every test was green while the interface looked like `01-dashboard.png`. The
gate is the answer to that, and it is the only new mechanism this slice adds.

**R4 — Every task names what it does not prove.** One line in the task report.
A row moves in `docs/VERIFY.md` only on a named command, on a named host, with
the output quoted.

## 4. The three cuts

All three are visible on the 1920 capture.

1. **The status ribbon dies.** `OBSERVATION · SSE connecté · dernier audit ·
   VÉRIFIER` is developer telemetry on an operator screen. The header keeps the
   connection dot, one state word and the last-audit age (design §3.1); the
   ribbon's per-mode counters, seal and proof entry are deleted, not moved — the
   last audit is already on the dashboard and the proof entry is destination 4.
   The component is deleted, not restyled.
2. **Two navigations become one.** The 232 px rail, the footer shortcut strip
   and the dashboard's quick-action strip all point at the same destinations.
   One 64 px header, four destinations, keys `1–4`, a bottom bar under 700 px.
3. **The primary action cannot sit below the fold.** Today it is clipped at
   1080 px. R3 makes that a failing gate rather than a matter of taste.

## 5. Tasks

The order is the rework. T1 changes the screen before anything invisible is
touched.

| # | Task | What appears on screen |
|---|---|---|
| **T0** | Ledger: the plan lands, the index re-dispositions fifteen rows so exactly one reads active, and the spec absorbs T0, the corrected ribbon cut, the screenshot path and the narrowed T9 | Nothing; the work order changes |
| **T1** | Shell: `TopNav` replaces `ModeRail`; the ribbon is deleted, the header keeps one state word and the last-audit age; `.btn` becomes 16 px on a 48 px target; bottom bar under 700 px | The 232 px rail is gone; one header |
| **T2** | Verdict hero **and** its four facts, mounted on the dashboard and the audit read state, fed by the sealed report; `cap_sla` gains `target_pct` and `deadline_ms` so the two objective lines are sealed with the measurement instead of retyped in the browser (`lib/dashboard.ts:128,136`) | The 40 px word, its sentence, four facts |
| **T3** | Remediation card: platform radio, measured and shaped rate per direction, “interface supposée” only where the script exists | “Protéger ce lien” |
| **T4** | Proof drawer: two tabs, platform-aware copy, seals and raw telemetry in tab 2 | “Voir la commande exacte” |
| **T5** | Dashboard, audit and evidence rewired; `dash-grid`, `syn-proof`, `syn-next`, `dash-onboarding` and the inline `ScienceDetails` deleted | Three destinations change at once |
| **T6** | Live trace on the observer: canvas, gap breaks, objective line | A running audit draws itself |
| **T7** | Bench operator surface: preset duel and research mode; the matrix axes and the leaderboard leave the default screen | The bench reads as one decision |
| **T8** | Lexicon gate: the banned-word list becomes a test over rendered surfaces; `DESIGN.md`, `PRODUCT.md` reconciled | No jargon on any default surface |
| **T9** | Evidence tail: verifier packet, the twelve `PENDING` dispositions, and one book-annex paragraph left in the working tree; no `pkg/figures` change | The thesis cites the console |

T5 is deliberately one task although it touches three destinations. Split, it
leaves half-updated screens, which is the exact failure this slice exists to
end.

Each task's acceptance is the same shape: `npm run typecheck && npm test` for
the files it owns, the destination rendered in a DOM test, the screenshot, and
`python tools/verify.py --only links` when it writes a document.

## 6. Evidence artifacts

- Screenshots: `docs/img/shots/2026-11-02/<destination>-<width>.png`, produced by
  `node web/frontend/capture-fresh.mjs http://127.0.0.1:9091 docs/img/shots/2026-11-02`
  against a served fresh bundle — never against `vite dev`. The `docs/img/`
  prefix is the one place the `hygiene` gate does not apply its 384 KB ceiling.
- Runtime gates: `node web/frontend/qa-walk.mjs http://127.0.0.1:9091`, extended
  with the three checks in R3.
- Each task's checkpoint quotes the command, the date, the host and the result,
  and names what it did not prove.

Screenshots are evidence of a render, not of a measurement. No screenshot may be
cited as proof that a link behaves in a given way.

## 7. Ledger

Twenty-four rows become **one active slice**.

| Disposition | Rows | Count |
|---|---|---|
| **Active** | 21 Console spine | 1 |
| **Superseded by 21** | 8, 14, 15, 18, 20A, 20B, 20C | 7 |
| **Deferred, named unblock** | 4 → a built `bpf_probe.o` on Linux 6.8 · 5 → a field trace of at least 30 baseline samples · 6 → a stable Linux 6.8 bench · 7 → a 7.1 BBRv3 kernel · 16 → QEMU and a CHR image · 17 → a Linux 6.8 host · 19 → the operator grant, items a–d | 7 |
| **Closed** | 1, 2, 3, 9, 10, 11, 12, 13, 20 | 9 |

The 22 existing plan files are kept. Their content is the record; only their
status changes. A closed row keeps its evidence and takes no further work.

### 7.1 Disposition of the 12 `PENDING` claims

T9 produces this table; the verifier decides each row, not the builder.

| Claim | Proposed |
|---|---|
| Slice 8 audit report surface | `SUPERSEDED` — the drawer of T4 replaces its surface |
| Interface v2 (four modes) | `SUPERSEDED` — the destinations survive; the surface does not |
| BBRv3 engine | `DEFER` with slice 7 |
| Trace-driven `.dist` engine | `DEFER` with slice 5 |
| 20A verdict cascade | re-run, promote |
| 20A remediation rates | re-run, promote |
| 20A economics numbers | re-run, promote |
| 20A campaign-frozen tier | re-run, promote |
| 20A live timeline | re-run, promote, plus the T6 render line |
| 20A duel endpoint | re-run, promote, plus the T7 render line |
| 20A run-start `run_id` guard | re-run, promote |
| 20C TUI mirror | re-run, promote from the packet at `docs/verification/2026-11-01-slice-20c.md` |

## 8. What this slice does not prove

Stated here so no reader has to infer it from a green suite.

- **No live audit was run.** Every archived report predates `operator_verdict`,
  so the hero renders “Pas encore mesuré” on all of them. Screenshots show that
  absence, and the green/amber/red cascade is proven by unit fixtures only.
- **No bench run.** T7 renders existing results; it does not produce new ones.
- **No row is `VERIFIED` by this slice.** The builder may not certify its own
  number. T9 prepares the packet; a separate verifier re-runs it and owns the
  rows. Until then the slice is “builder pass complete, packet ready”.
- **A screenshot proves a render, not a measurement.** Colour, spacing and
  copy are proven by it; a latency claim is not.
- **The mobile bar is proven at 390×844 only**, the width `capture-fresh.mjs`
  covers. No tablet width is claimed.

## 9. Amendments to the console design

1. **Verdict field name.** `2026-10-31-cgo-operator-console-design.md` §3.2
   shows `verdict`; the report carries `operator_verdict`
   (`pkg/audit/report.go:44`). The report wins; this spec is the corrected
   name, so a verifier does not chase a field that does not exist.
2. **The status cluster keeps four things, and the fourth is the audit age.**
   Connection, mode word, last-audit age and settings stay in the header; the
   ribbon's corpus counts, seal and proof button do not, because the dashboard
   already shows the audit and destination 4 is the proof entry. §3.1 lists four
   and leaves “last audit” unplaced; this is the placement.
3. **Hero and facts are one task.** §3.2 defines them as one visual unit; the
   previous plan split them across two tasks, which is how the hero shipped
   unmounted.
4. **13 px stays a drawer rule, not yet a gate.** The scale bullet reserves 13 px
   for the drawer and then had 168 default-surface rules at that size. T8 gates
   the lexicon only; the type-scale gate is named as unproven in T8's report
   rather than promised. `designFloor.test.ts` already rejects numeric font
   sizes in `.tsx`; a CSS-level 13 px gate needs its own allowlist and is not in
   this slice.
5. **`qa-walk.mjs` references three ribbon selectors** (`.ribbon-corpus`,
   `.ribbon-seal`, `.status-ribbon .btn`, lines 48 and 222–226). T1 rewrites
   them against the new header; deleting the ribbon without that edit breaks
   the walk silently.
6. **`cap_sla` grows two additive fields.** `target_pct` and `deadline_ms` are
   copied from `metrics.CAPBreachThresholdPct` and `metrics.CAPDeadlineMs` at
   seal time. Until now the browser retyped both as literal strings
   (`lib/dashboard.ts:128,136`) that a test asserts, so a change to the engine
   constant would leave the UI quoting a target nobody applies. A report
   archived before this change renders its fact with no objective line rather
   than a remembered one.
7. **Screenshots live under `docs/img/shots/`.** The `hygiene` gate rejects a
   tracked file above 384 KB outside `docs/img/`, and a 1920×1080 PNG crosses
   that. Putting the evidence where the gate already allows it avoids both a
   gate edit and an uncommitted artifact that the `links` gate could not check.
8. **The live trace shows two facts, not three.** §3.5 lists débit among the
   running facts, but `audit.Progress` carries no byte counter: the bulk outcome
   exists only at the end of the window. T6 renders temps de réponse and réponse
   sous charge; débit appears in the frozen report and the drawer. Reporting a
   throughput during the run would mean inventing one.

## 10. Decision log

| Date | Decision |
|---|---|
| 2026-11-02 | One active slice; fifteen rows change status, nine closed rows stay untouched, none deleted |
| 2026-11-02 | No live audit; absence renders as a word, not a zero |
| 2026-11-02 | The screenshot becomes an acceptance artifact (R3) |
| 2026-11-02 | The builder cannot promote a `VERIFY` row; T9 prepares, a verifier signs |
| 2026-11-02 | The ribbon is deleted rather than restyled |
