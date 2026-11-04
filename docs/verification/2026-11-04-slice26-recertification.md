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
ledger. No product behaviour changed in this slice; nine Go files were
formatted after the first verifier pass (whitespace only).

- Evidence commit: `4712775` — screenshots, capture script, plan.
- Code under test is unchanged since `9df16a8` (last code commit `7dcf502`);
  `4712775` adds only `docs/img/shots/2026-11-04/` and
  `web/frontend/slice26-capture.mjs`.
- Branch: `rebuild/plan1-economics-audit`
- Host: DESKTOP-B6K9UCM (Windows 10, MINGW64), Go 1.26.5, Node 22.23.2.
- Date: 2026-11-04, 20:02–20:30 UTC (`date -u`; verifier start 20:20:09,
  end 20:30:01 UTC).
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
| 7 | `gofmt -l pkg cmd` | after the fix below, no output; before it, 9 files (finding 1) |
| 8 | `go vet ./cmd/cgo` | exit 0, no output |

`python tools/verify.py` output, verbatim:

```
PASS frozen      OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid | OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents | docs/frozen-manifests.sha256: 159 manifests
PASS provenance  179 headers checked against 5 frozen eras; 0 valuation sidecars checked; writer covered by tests
PASS tests       go vet + go test in 76.9 s
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     679 tracked files clean; legacy allowlist: 18 backup archives and pkg/results/runs-frozen.tar.gz; 2 named tailnet addresses
PASS links       5 documents scanned; plan and spec links resolve

6 passed, 0 failed, 0 not run
```

Book gate, two `OK` lines:

```
OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid
OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents
```

### Independent verifier, first pass (2026-11-04 20:20–20:30 UTC)

The verifier re-ran on `319f4e2`, fresh, not cached:

| Command | Exit | Observed |
|---|---|---|
| `go test ./... -count=1` | 0 | 19 `ok`; `pkg/audit 192.310s`, `pkg/api 67.046s` |
| `python tools/verify.py` | 0 | `6 passed, 0 failed, 0 not run`; `PASS tests ... 75.7 s`; hygiene 679 files |
| `python book/tools/audit-archive.py` | 0 | two `OK` lines |
| `cd web/frontend && npm test` | 0 | 59 files, 283 tests, `Duration 132.32s` |
| `cd web/frontend && npm run build` | 0 | `TOTAL gz : 353.7 KB` |

That pass found the ten defects listed in the next section. This packet now
records their fixes; the second pass is the one that signs.

## Task A — 20A binding (seven rows)

Targeted tests, all `PASS` at `319f4e2`:

| Row | Command | Observed |
|---|---|---|
| A1 verdict cascade | `go test ./pkg/metrics -run "TestVerdictNeverGreenWithoutGrade" -v` | `PASS` |
| A1 sealed report | `go test ./pkg/audit -run "TestReportVerdictRefusesGreenWithoutLoad" -v` | `PASS` |
| A2 rates + assumed LAN | `go test ./pkg/audit -run "TestReportLinuxOnlyDownEmitsLinuxIngressAndNoRouterOS" -v` | `PASS` |
| A3 economics | `go test ./pkg/metrics -run "TestBuildEconomicsNumbersPerModel|TestBuildEconomicsNumbersNotInventedWhenUnmeasured|TestEconomicsNumbersJSONKeys" -v` | 3 `PASS` |
| A4 frozen tier | `go test ./pkg/campagne -run "TestMatrixConfigFreezesActiveTier|TestCostARPerHUsesFrozenTier|TestRunEventCostUsesFrozenTier|TestResumeKeepsTheFrozenTier" -v` | 4 `PASS` |
| A4 write guard | `go test ./pkg/api -run "TestCostTierRefusedWhileCampaignRuns" -v` | `PASS` |
| A5 timeline | `go test ./pkg/audit -run "TestTimelineBucketsBySecondAndKeepsGaps|TestTimelineCapIsTotalPlusFive" -v` | 2 `PASS` |
| A6 duel | `go test ./pkg/api -run "Duel" -v` | 2 `PASS` (`TestDuelPicksBestAndBaseWithGuardrail`, `TestDuelRefusesUnknownProfile`) |
| A6 provenance | `go test ./pkg/results -run "TestGroupRunsExcludeInvalidOnlyRuns" -v` | `PASS` |
| A7 run id | `go test ./pkg/api -run "RunStartRefusesEscaping|RunStartSubmatrix|RunOptsReach|RunStartProvenance" -v` | 5 `PASS` |

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
(eight files). Five historical captures at 390 px exist in the same folder and
are **outside the claimed range**; they are kept as history, not deleted:
`auditer-lire-390.png`, `auditer-lire-fixture-390.png`, `banc-resultats-390.png`,
`remediation-fixture-390.png`, `tableau-de-bord-390.png`. The packet
`2026-11-02-slice-21.md` carries the same list.

## Task D — 22–25 evidence and screenshots

Common commands: `npm test` → 59 files / **283 tests passed**; `npm run build`
→ `TOTAL gz : 353.7 KB`; `go test ./...` → all `ok`; `verify` → 6/6.

Eight screenshots at 1024×768 and 1920×1080, `docs/img/shots/2026-11-04/`,
plus the extracted tooltip text:

| Slice | File | Assertion captured |
|---|---|---|
| 22 | `slice22-audit-remediation-suppressed-{1024,1920}.png` | Archived report `Département X` reads « Pas encore mesuré »; the card is not rendered at all (no `remediation` object), so « Aucune remédiation émise — capacité ou RTT repos non mesurés. » is present. The source `RemediationCard` keeps its heading when `state=observation_only`; only the shaped script is suppressed. |
| 23 | `slice23-settings-deadline-{1024,1920}.png` | Settings drawer, `Deadline small p95` = **220**. |
| 23 | `slice23-bench-bbrv3-{1024,1920}.png` | Research mode, controls row reads `cubic · bbr` + `BBRv3 non détecté`. |
| 24 | `slice24-bench-edge-labels-{1024,1920}.png` | Edge control reads `Maxi avant` and `sans réglage (file simple)`. |
| 25 | `slice25-tooltip.txt` | A native `title` is not painted, so no separate PNG is claimed; the exact string is extracted from the live DOM into `slice25-tooltip.txt`. The button itself is visible in the slice-24 capture. |

`docs/img/shots/2026-11-04/evidence.json` records the machine checks for both
viewports: `slice22.hasSuppressedLine=true`, `hasProtectCta=false`;
`slice23.deadline="220"`, `hasBbrv3Hint=true`; `slice24.hasMaxiAvant=true`,
`hasSansReglage=true`; `slice25.titleMatches=true`. The capture script is
`web/frontend/slice26-capture.mjs`; it starts no audit and no bench, and it
writes `slice25-tooltip.txt` on every run.

## Corrections after the first verifier pass (2026-11-04 20:30 UTC)

The first independent verifier pass found ten defects. Each is fixed here; the
verifier re-checks the deltas before signing.

| # | Finding | Fix |
|---|---|---|
| 1 | `gofmt -l pkg cmd` listed 9 files, not empty | All nine formatted; `gofmt -l pkg cmd` now empty. Whitespace only, no behaviour change. |
| 2 | `slice25-tooltip.txt` referenced but absent | File created; `slice26-capture.mjs` now writes it on every run. |
| 3 | Slice-25 PNGs byte-identical to slice 24 | Both deleted; slice 25 claims no separate PNG, only the extracted text. |
| 4 | A3/A4/A5/A7 used `\|` in double quotes, selecting zero tests | Commands corrected to `|`; the verifier's corrected runs showed 3, 4, 2 and 5 `PASS`. |
| 5 | `npm run build` dirties `web/frontend/dist/index.html` | File restored before commit; this packet states the build dirties it and that the restore is mandatory. |
| 6 | Five 390 px captures, not four | All five named in this packet and in the slice-21 post-script. |
| 7 | Slice-22 wording said the heading was suppressed | Corrected: the archived report renders no card at all; `RemediationCard` keeps `<h2>Protéger ce lien</h2>` when `state=observation_only` and suppresses the shaped script. The invariant is the script, not the heading. |
| 8 | Local +03 timestamps labelled UTC | Corrected to 20:02–20:30 UTC (builder) and 20:20–20:30 UTC (verifier). |
| 9 | Stale hygiene count 666 | Quoted block refreshed to the verifier's 679 tracked files. |
| 10 | Plan T3 still said Slice 8 `SUPERSEDED` | Plan corrected to `PROMOTE`, matching the packet and the 14 passing tests. |

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
- **The slice-25 tooltip is not painted.** The button is visible in the
  slice-24 capture; the exact tooltip string is the DOM extraction in
  `slice25-tooltip.txt`.
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
- The eight PNG files exist at the stated dimensions and are not blank; the
  tooltip text file is present and byte-exact.
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
