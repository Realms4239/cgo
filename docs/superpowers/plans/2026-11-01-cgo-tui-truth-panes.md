# Slice 20C — TUI truth panes

**Spec:** `docs/superpowers/specs/2026-10-31-cgo-operator-console-design.md` §3.4,
paragraph “TUI mapping”.
**Depends on:** 20A (the report fields) — builder-complete.
**Scope guard:** no new tab, no frontend change, no live bench, no `/tmp` write.

## Decision of record (2026-11-01)

The spec's sentence had two incompatible readings, because the TUI's tabs 4 and
5 are the **campaign live frame** and the **results matrix**, not an audit:

1. mirror the last audit above those two views, keeping five tabs; or
2. add a sixth audit tab.

The human chose **1**. Five tabs stay. Tabs 4 and 5 (indices 3 and 4) gain a
one-line verdict header and four fact lines, read from the last sealed audit.
Keys `d`, `p`, `c` operate on that audit.

## The four facts (defined here, because the spec names them without a list)

The spec says “the four facts as four lines” and, in the drawer bullet, lists
what a fact may be. This slice fixes the four lines explicitly so the verifier
can check them against a report:

| Line | Report fields | When unmeasured |
|---|---|---|
| `Repos` | `measured.rtt_idle_p50_ms`, `rtt_idle_p95_ms` | `non mesuré` |
| `Sous charge` | `measured.rtt_loaded_p50_ms`, `rtt_loaded_p95_ms`, `measured.bloat_delta_ms`, objective `metrics.CAPDeadlineMs` | `charge non établie` when `load_attempted` and not `load_measured`; `non mesuré` otherwise |
| `Débit` | `measured.throughput_mbps` | `non mesuré` |
| `Perte` | `measured.loss_pct` | `non mesuré` |

A zero is only printed when the engine recorded a zero, never as a stand-in for
“absent” — the report's own rule.

## Task 1 — read the last audit, and tell “none” from “fetch failed”

**Files:** new `cmd/cgo/tui_audit.go`, new `cmd/cgo/tui_audit_test.go`.

`GET /api/audit/status` returns `{running, last, report}`. Decode only the subset
this slice renders (`audit_id`, `site`, `link_type`, `timestamp`,
`bloat_grade`, `operator_verdict`, `measured`, `provenance`, `remediation`).

- `auditView` carries `OK bool` and `Why string` so the TUI can say
  “aucun audit scellé” (no report), “API injoignable — <err>” (transport), and
  “rapport illisible” (decode) as three different lines. Collapsing them into one
  is the E10/E11 defect class: an unreadable thing must not read as an absent
  thing.
- Poll every 10 s alongside the existing 1 s frame poll, chained like
  `pollResults`.

**Tests:** `TestAuditViewNoneIsNotAFailure`, `TestAuditViewTransportFailureNamesTheAPI`,
`TestAuditViewDecodesVerdictAndFacts`.

## Task 2 — verdict header and four fact lines on tabs 4 and 5

**Files:** `cmd/cgo/tui_audit.go`, `cmd/cgo/tui.go` (the two view cases).

- `verdictHeader(v)` returns one line: `<word> — <sentence>` from
  `operator_verdict`, or `Verdict indisponible` when the report carries none.
- `auditFactLines(v)` returns exactly four strings from the table above.
- Tabs 4 and 5 print the header and then the four lines, before the existing
  cards/table. Nothing is removed; the campaign views keep their own data.

**Tests:** `TestAuditFactLinesNeverInventAValue` (empty measured ⇒ four
`non mesuré`), `TestAuditFactLinesNameChargeNonEtablie`,
`TestVerdictHeaderUsesTheReportWord`.

## Task 3 — panes `d` and `p`

**Files:** `cmd/cgo/tui_audit.go`, `cmd/cgo/tui.go` (`Update`, `View`).

- `d` opens the deployment pane: sized rates (`up_measured_mbps` →
  `up_shaped_mbps`, same for down), qdisc and reason, the platform's commands,
  the RouterOS replay line worded `validée par rejeu` / `non validée` /
  `aucune preuve`, and `routeros_lan_assumed` when true.
- `p` opens the proof pane: p50/p95 idle and loaded, worst delta, QDI, loss,
  throughput, direction, sensor tier (`kernel_telemetry.available` →
  `eBPF`, else `sources`/`tcp_info`, else `standard`), source and telemetry
  SHA-256, `/api/audit/report?audit_id=…`, and the duel link.
- Both panes render full screen: the tab bar is not drawn, `esc` returns,
  `q` still quits. Text wraps at 78 columns (`wrap78`), one command per line,
  no horizontal scroll.
- `tab` toggles the pane's platform (`routeros` ↔ `linux`) when both exist;
  the pane always shows which platform `c` would write.

**Tests:** `TestWrap78NeverExceedsTheColumn` (long command, long sentence, CJK-free),
`TestDeploymentPaneOneCommandPerLine`, `TestProofPaneNamesSensorTier`.

## Task 4 — `c` writes the script beside the sealed report

**Files:** `cmd/cgo/tui_audit.go`, `cmd/cgo/tui.go`.

- Path: `data/audits/<audit_id>/remediation.routeros.rsc` or
  `remediation.linux.sh`. `MkdirAll` the directory; never `/tmp`, never the
  home directory.
- `audit_id` comes from the server, so it is validated before it becomes a path
  (`safeAuditID`: letters, digits, `-`, `_`, `.`; rejects `..`, separators and
  empty). Same rule the API applies to `run_id` (E11).
- Linux script = `linux_tc` lines joined with newlines plus a trailing newline;
  RouterOS = the script verbatim.
- `c` attempts OSC 52 and prints, in the status line, the written path **and**
  that the copy was attempted. It does not claim the clipboard took it: there is
  no reliable refusal signal in a terminal. The spec asks to “say plainly when
  the terminal refused it”; this slice words the line so no success is asserted
  (`copie OSC 52 tentée — le terminal peut l'ignorer`), and records that
  deviation here rather than faking detection.

**Tests:** `TestWriteRemediationRefusesTraversal`, `TestWriteRemediationWritesBothPlatforms`,
`TestWriteRemediationStaysUnderDataAudits`.

## Task 5 — help line, gates, ledger

- The tab bar help gains `d/p/c` on tabs 4-5 only, worded once.
- `gofmt -l cmd`, `go vet ./cmd/...`, `go test ./cmd/... -count=1`, then
  `python tools/verify.py`.
- `docs/VERIFY.md` row (builder pass, `PENDING`), `docs/CHECKPOINTS.md` entry,
  index row 20C → **In progress** with the verifier named as the next owner.

## What this slice will not claim

- Nothing is observed on a live audit. The TUI is exercised against fixtures and
  the JSON shapes above.
- OSC 52 delivery is not verified; only the attempt is.
- Slice 20B's console work (web) is untouched.
