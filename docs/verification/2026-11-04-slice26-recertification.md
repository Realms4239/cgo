# Verifier packet — slice 26′, exhaustive re-certification 20A–25

*Prepared 2026-11-04 by the builder. The builder may not certify its own
numbers: this packet lists what the verifier runs, what must appear, where the
evidence lands, and the builder's proposed disposition for every row. The
signature block stays empty until an independent verifier fills it.*

## Scope

Slice 26′ (spec `docs/superpowers/specs/2026-11-04-cgo-slice26-recertification-design.md`,
plan `docs/superpowers/plans/2026-11-04-cgo-slice-26-recertification.md`)
re-binds every `PENDING` claim from slices 20A, 20C and 21–25 to fresh output,
adds the missing visual evidence for 22–25 at 1024 and 1920, and reconciles the
ledger. No product code changed in this slice.

- Evidence commit: `4712775` — screenshots, capture script, plan.
- Code under test is unchanged since `9df16a8` (last code commit `7dcf502`);
  `4712775` adds only `docs/img/shots/2026-11-04/` and
  `web/frontend/slice26-capture.mjs`.
- Branch: `rebuild/plan1-economics-audit`
- Host: DESKTOP-B6K9UCM (Windows 10, MINGW64), Go 1.25, Node 22.
- Date: 2026-11-04, 20:02–23:20 UTC (`date -u`).
- Accepted desktop range: 1024–1920 px. No mobile claim.
- Book: `book/content/09-annexes.js` carries one paragraph in the working tree
  only; `book/` is never committed on this branch.

## Commands observed by the builder

| # | Command | Observed on 2026-11-04 |
|---|---|---|
| 1 | `go test ./... -count=1` | exit 0; every package `ok` (`pkg/api 64.579s`, `pkg/audit 189.836s`) |
| 2 | `python tools/verify.py` | **6 passed, 0 failed, 0 not run** (`tests 76.9 s`) |
| 3 | `python book/tools/audit-archive.py` | both `OK` lines |
| 4 | `cd web/frontend && npm test` | **59 files, 283 tests passed** (`Duration 101.95s`) |
| 5 | `cd web/frontend && npm run build` | exit 0; `TOTAL gz : 353.7 KB` |
| 6 | `go build ./...` | exit 0 |
| 7 | `gofmt -l pkg cmd` | no output |
| 8 | `go vet ./cmd/cgo` | exit 0, no output |

`python tools/verify.py` output, verbatim:

```
PASS frozen      OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid | OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents | docs/frozen-manifests.sha256: 159 manifests
PASS provenance  179 headers checked against 5 frozen eras; 0 valuation sidecars checked; writer covered by tests
PASS tests       go vet + go test in 76.9 s
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     666 tracked files clean; legacy allowlist: 18 backup archives and pkg/results/runs-frozen.tar.gz; 2 named tailnet addresses
PASS links       5 documents scanned; plan and spec links resolve

6 passed, 0 failed, 0 not run
```

Book gate, two `OK` lines:

```
OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid
OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents
```

## Task A — 20A binding (seven rows)

Targeted tests, all `PASS` at `0b7b723`:

| Row | Command | Observed |
|---|---|---|
| A1 verdict cascade | `go test ./pkg/metrics -run "TestVerdictNeverGreenWithoutGrade" -v` | `PASS` |
| A1 sealed report | `go test ./pkg/audit -run "TestReportVerdictRefusesGreenWithoutLoad" -v` | `PASS` |
| A2 rates + assumed LAN | `go test ./pkg/audit -run "TestReportLinuxOnlyDownEmitsLinuxIngressAndNoRouterOS" -v` | `PASS` |
| A3 economics | `go test ./pkg/metrics -run "TestBuildEconomicsNumbersPerModel\|TestBuildEconomicsNumbersNotInventedWhenUnmeasured\|TestEconomicsNumbersJSONKeys" -v` | 3 `PASS` |
| A4 frozen tier | `go test ./pkg/campagne -run "TestMatrixConfigFreezesActiveTier\|TestCostARPerHUsesFrozenTier\|TestRunEventCostUsesFrozenTier\|TestResumeKeepsTheFrozenTier" -v` | 4 `PASS` |
| A4 write guard | `go test ./pkg/api -run "TestCostTierRefusedWhileCampaignRuns" -v` | `PASS` |
| A5 timeline | `go test ./pkg/audit -run "TestTimelineBucketsBySecondAndKeepsGaps\|TestTimelineCapIsTotalPlusFive" -v` | 2 `PASS` |
| A6 duel | `go test ./pkg/api -run "Duel" -v` | 2 `PASS` (`TestDuelPicksBestAndBaseWithGuardrail`, `TestDuelRefusesUnknownProfile`) |
| A6 provenance | `go test ./pkg/results -run "TestGroupRunsExcludeInvalidOnlyRuns" -v` | `PASS` |
| A7 run id | `go test ./pkg/api -run "RunStartRefusesEscaping\|RunStartSubmatrix\|RunOptsReach\|RunStartProvenance" -v` | 5 `PASS` |

Builder transcript: `.perch/slice26/taskA-1.log`, `taskA-2.log`,
`taskA-3.log` (local, not committed).

Residuals unchanged and carried into the rows: fixtures only, no live audit;
a hard kill before `Freeze` leaves no frozen tier; the CHR proof line stays
unworded.

## Task B — 20C binding (one row)

`go test ./cmd/cgo -run "TestAudit|TestVerdictHeader|TestWrap78|TestDeployment|TestProofPane|TestWriteRemediation|TestScriptFor|TestViewShows|TestRenderedViews" -v -count=1`
→ exit 0, **16 tests PASS**, `ok ... 0.462s`.

The five clauses are covered by:
`TestViewShowsAuditMirrorOnTabs4And5`,
`TestAuditFactLinesNeverInventAValue` + `TestAuditFactLinesNameChargeNonEtablie`,
`TestWrap78NeverExceedsTheColumn` + `TestDeploymentPaneOneCommandPerLine` +
`TestRenderedViewsStayWithinEightyColumns`,
`TestWriteRemediationRefusesTraversal` + `TestWriteRemediationStaysUnderDataAudits`
+ `TestScriptForRefusesMissingPlatform`,
`TestAuditViewNoneIsNotAFailure` + `TestAuditViewTransportFailureNamesTheAPI`.

Residuals unchanged: fixture mirror only, OSC 52 attempted not delivered.

## Task C — slice 21 and the four earlier rows

### Correction to the slice-21 packet

The 21 packet proposed `SUPERSEDED` for **Slice 8 audit report surface and
citation**. That proposal is wrong and is withdrawn here. The surface still
exists and is reachable: `EvidenceView` → `AuditReportsPanel` →
`AuditReportCard`, with `src/lib/citation.ts`. At HEAD, three targeted files
pass: **14 tests** (`citation.test.ts` 3, `AuditReportCard.dom.test.tsx` 7,
`AuditReportsPanel.test.tsx` 4). Builder proposal: **PROMOTE**, not supersede.

The 20A packet's bullet "the TUI mapping has no plan and is not built" was true
on 2026-11-01 and was superseded hours later by slice 20C; it is historical,
not a live claim.

### Four earlier rows — builder proposals

| Row | Proposal | Basis |
|---|---|---|
| Slice 8 audit report surface and citation | **PROMOTE** | Corrected above; 14 targeted tests pass at HEAD; surface reachable. |
| Interface v2 (four modes) and the academic title | **SUPERSEDED** | The rail/ribbon, the 390 px ribbon behaviour and the perf figures certified a surface removed by the console spine. The four destinations and the academic title persist and are covered by the 21 row. |
| BBRv3 engine | **BLOCKED** | No Linux 7.1 kernel on this host or the 6.8 bench; unit tests exist but no socket observation. Named unblock: a 7.1 host. |
| Trace-driven `.dist` engine | **BLOCKED** | No field trace with ≥ 30 baseline samples; the two `SMOKE-TEST` traces are 19 and 25 rows. Named unblock: a real field trace. |

### Slice 21 screenshots

The claimed set stays `docs/img/shots/2026-11-02/t8-{dashboard,audit,bench,evidence}-{1920,1024}.png`
(eight files). Four historical captures at 390 px exist in the same folder and
are **outside the claimed range**; they are kept as history, not deleted. The
packet `2026-11-02-slice-21.md` carries a post-script saying so.

## Task D — 22–25 evidence and screenshots

Common commands: `npm test` → 59 files / **283 tests passed**; `npm run build`
→ `TOTAL gz : 353.7 KB`; `go test ./...` → all `ok`; `verify` → 6/6.

Screenshots at 1024×768 and 1920×1080, `docs/img/shots/2026-11-04/`:

| Slice | File | Assertion captured |
|---|---|---|
| 22 | `slice22-audit-remediation-suppressed-{1024,1920}.png` | Archived report `Département X` reads « Pas encore mesuré »; « Aucune remédiation émise — capacité ou RTT repos non mesurés. » present; « Protéger ce lien » absent. |
| 23 | `slice23-settings-deadline-{1024,1920}.png` | Settings drawer, `Deadline small p95` = **220**. |
| 23 | `slice23-bench-bbrv3-{1024,1920}.png` | Research mode, controls row reads `cubic · bbr` + `BBRv3 non détecté`. |
| 24 | `slice24-bench-edge-labels-{1024,1920}.png` | Edge control reads `Maxi avant` and `sans réglage (file simple)`. |
| 25 | `slice25-bench-tooltip-{1024,1920}.png` | The `Maxi avant` button in place; the native `title` is not painted in a screenshot, so the exact string is extracted into `slice25-tooltip.txt`. |

`docs/img/shots/2026-11-04/evidence.json` records the machine checks for both
viewports: `slice22.hasSuppressedLine=true`, `hasProtectCta=false`;
`slice23.deadline="220"`, `hasBbrv3Hint=true`; `slice24.hasMaxiAvant=true`,
`hasSansReglage=true`; `slice25.titleMatches=true`. The capture script is
`web/frontend/slice26-capture.mjs`; it starts no audit and no bench.

## Builder proposals for the remaining rows

| # | Row (short) | Proposal | Evidence to cite |
|---|---|---|---|
| 1 | Slice 8 report surface and citation | **PROMOTE** | 14 targeted frontend tests at HEAD. |
| 2 | BBRv3 engine | **BLOCKED** | No 7.1 kernel; named unblock. |
| 3 | Trace-driven `.dist` | **BLOCKED** | No field trace ≥ 30 samples; named unblock. |
| 4 | Interface v2 (four modes + title) | **SUPERSEDED** | Surface replaced by 21; modes/title covered by row 13. |
| 5 | 20A verdict cascade | **PROMOTE** | A1 commands above. |
| 6 | 20A rates + assumed LAN | **PROMOTE** | A2 command above. |
| 7 | 20A economics numbers | **PROMOTE** | A3 command above. |
| 8 | 20A frozen tier + guard | **PROMOTE** | A4 commands above. |
| 9 | 20A timeline | **PROMOTE** | A5 commands above. |
| 10 | 20A duel | **PROMOTE** | A6 commands above. |
| 11 | 20A run id | **PROMOTE** | A7 command above. |
| 12 | 20C TUI mirror | **PROMOTE** | 16 tests above. |
| 13 | Slice 21 console spine | **PROMOTE** | 59 files / 283 tests; build 353.7 KB; eight captures at 1024/1920. |
| 14 | Slice 22 remediation invariant | **PROMOTE** | Go suite + slice22 screenshots. |
| 15 | Slice 23 academic defaults | **PROMOTE** | Go suite + deadline 220 and BBRv3 screenshots. |
| 16 | Slice 24 operator surface truth | **PROMOTE** | 283 tests + edge-label screenshots. |
| 17 | Slice 25 tooltip honesty | **PROMOTE** | Title text in `slice25-tooltip.txt`, `titleMatches=true`, 283 tests. |

The verifier decides every disposition. A `PROMOTE` here is a proposal, not a
status. The builder has not edited any row's status.

## What this slice does NOT prove

- **No live observation.** No audit, campaign, bench, CHR boot, Linux 6.8
  acceptance, eBPF attach, IFB ingress, `.dist` calibration or BBRv3 socket was
  run. Everything is a unit test, a gate, or a local render in observe mode.
- **Screenshots prove a render, not a measurement.** The archived report and
  the duel figures shown in the captures come from the frozen local data, not
  from a fresh run.
- **The slice-25 tooltip is not painted.** The PNG shows the button; the exact
  tooltip string is the DOM extraction in `slice25-tooltip.txt`.
- **The 390 px historical captures are not re-validated.** Mobile remains
  dropped by human amendment; 1024–1920 is the accepted range.
- **Residuals stay open**: hard kill before `Freeze`, OSC 52 delivery, CHR
  proof wording, the NEW1 cockpit (deferred by design), the refuted
  `PREUVE VÉRIFIABLE` and tile truncation items.
- **No merge to `main`** is proposed or performed.

## Commands expected from the verifier

Re-run 1–8 above on a clean tree, plus the targeted commands in Tasks A–C, and
compare. Any mismatch is a finding. Then check:

- `git status --short` shows no tracked modification before or after; `book/`
  untouched except the one annex paragraph.
- The ten PNG files exist at the stated dimensions and are not blank.
- `evidence.json` matches the assertions above.
- No `VERIFIED` row for 20A–25 exists until the verifier writes it.

## Signature block

```
Verifier: ______________________
Date (UTC): ____________________
Rows promoted: _________________
Rows blocked: __________________
Rows superseded: _______________
Screenshots checked (Y/N): ______
Residuals accepted (Y/N): ______
Notes: _________________________
```
