# Final fix report — END-corpus interface programme (2026-11-22)

Wave: single final fix wave (fresh agent, owns the wave). Worktree `C:/cgo/.worktrees/main-reconciled`, branch `main-reconciled`, base HEAD `ab45448`. Corpus `C:/END` untouched. No host/bench contact, no corpus edit, no merge, no fabricated value/seal/PASS/measurement, no new dependency/hex/font-size/`transition: all`. Protected `docs/CHECKPOINTS.md`, `docs/VERIFY.md` and frozen corpus untouched.

## Finding 1 — "Les deux" unreachable → FIXED (honest engine path confirmed)

**Systematic-debugging Phase 1 — trace, not guess.** The important finding assumed direction `both` was dead because `link_direction` offered only `up|down`. Tracing the value showed a real campaign-level path:

- `runreq/runreq.go:29-31,77` — `DirectionBoth = "both"`; `ValidDirection` accepts `up|down|both`; `CampaignRequest.ValidateAll()` enforces it. This is the single canonical gate used by `POST /api/run/start` (`api/server.go:817`).
- `api/run_direction_test.go:31` — `POST /api/run/start` already returns 200 for `"direction":"both"`, 400 for an unknown direction.
- `cmd/cgo/server.go:246-265` — for `o.Direction == "both"` the engine runs the up matrix, then chains a second matrix via `campagne.ProdDepsDown()` (or `ProdDepsDualNS("down")`) **in the same run** — genuine sequential up AND down.
- `pkg/campagne/prod.go:29-32` + `prod_down_test.go:17-19` — `ProdDepsDown()` sets `Direction="down"`; `pkg/campagne/matrix.go:199` labels each cell from `deps.Direction`.
- `audit/audit.go:203,415-440,556` — `audit.Run` with `"both"` executes two load windows (up then down) on one shared baseline.

**Corpus-safety check (the stated constraint).** `model/types.go:136-139` keeps `model.Event.Direction` frozen as `up|down` — without it, historical download cells merge with upload and corrupt 150 runs. The `both` value is a *run-level campaign selector*, never written into a cell: the chained down matrix emits `Direction="down"`, the up matrix `Direction="up"`. The frozen field is preserved. Adding `both` to the schema is therefore a real control, not a fake one.

**Fixes.**
- `api/server.go:142` — `link_direction` enum now `["up","down","both"]`, default still `up`; description states up/down mirror the frozen field and `both` is a RRUL campaign (up matrix then down matrix, cells still up/down).
- `api/schema_contract_test.go:44` — updated enum assertion; new `TestLinkDirectionBothIsExecutable` (`:122`) proves (a) every enum value is accepted by `runreq.ValidDirection` (so the UI advertises only executable directions) and (b) `POST /api/run/start` with `direction:"both"` reaches `StartFn` as `RunOpts.Direction=="both"`.
- `web/frontend/src/components/labo/Configurer.capabilities.dom.test.tsx:28,83-96` — fixture enum now carries `both`; the old `not.toContain('Les deux')` assertion is **flipped** to `toContain('Les deux')` plus an explicit 3-button count. The empty-enum absence case (`— direction indisponible`, no buttons) is retained.
- `docs/superpowers/specs/2026-11-22-end-corpus-interface-design.md:17` — §3 reconciled: `both` documented as campaign-level with the engine execution path and the frozen-cell invariant. §1 already required "Les deux" and stays unchanged — the UI now satisfies it.

**RED→GREEN.** With production files temporarily reverted to `ab45448` (tests kept): `go -C api test -run 'TestSchemaEndCorpusParams|TestLinkDirectionBothIsExecutable'` → FAIL (`link_direction.enum = [up down]`; `both doit être annoncé et exécutable`), exit 1. After fix → exit 0. Note honestly: the DOM `(c)` case is a *contract flip* (fixture + assertion), not a production-code RED — `DIRECTION_LABELS.both` already existed and rendered; the new production behavior proven here is the schema/engine contract in Go.

## Finding 2 — `dist_name` enum wording → FIXED (design wording reconciled)

Design §3 said "a dynamic enum populated only by real sidecar-verified calibrations", but `api/server.go:143` intentionally serves an empty enum and the client reads `GET /api/calibrations` directly (`Configurer.tsx:36-45`). Function is preserved. `docs/superpowers/specs/2026-11-22-end-corpus-interface-design.md:18` now states the enum is intentionally empty in the static registry, the verified list comes from `/api/calibrations`, and this is an explicit "no static fabrication" boundary — not missing dynamic population. No code change; no fabricated entries.

## Finding 3 — untested `ListCalibrations` fallbacks → FIXED (coverage added)

`api/calibrations_test.go:113-199` adds `writeSidecarFixture` plus table-driven `TestListCalibrationsFallbacks` covering the branches exercisable without a live bench: `sha256` ← `dist_sha256` fallback, `p50_ms` ← `source_p50_rtt_ms` fallback, both combined, and rejection of `samples<=0` / `p50<=0` (zero and negative). Each case writes a real temp `.dist` and a sidecar whose hash seals those exact bytes — no mocked measurement. The `.dist`-read-failure and sha-mismatch branches were already covered by `TestCalibrations` (`bad.json`); nothing needed parking. Honest note: these are characterization tests — the branches existed and already behaved correctly, so they are GREEN on old code too; the finding was missing coverage, not a defect, so there is no RED stage for them.

## Finding 4 — Canal "Empirique .dist" selectable with zero calibrations → FIXED (minimal coupling)

`Configurer.tsx:82-107,138` — `optionControl` takes an optional `unavailable` list; the Canal call passes `['empirique']` when `verified.length === 0`. The button keeps the corpus-required label `Empirique .dist`, is disabled with `aria-disabled="true"`, and carries the honest `— aucune calibration vérifiée` text; `Statique` stays enabled. DOM test `(b2)` (`Configurer.capabilities.dom.test.tsx:70-81`) asserts label retention + disabled state + absence word. `Statique`/`Empirique` both enabled with a verified calibration (test `(b)`); the aria-pressed selection test still passes.

**RED→GREEN:** with old `Configurer.tsx`, `(b2)` FAILS (`expected 'CanalStatiqueEmpirique .dist' to contain 'aucune calibration vérifiée'`), exit 1; with fix, exit 0.

## Finding 5 — "no duplicate request" claim inaccurate → FIXED (report + code)

`task-3-report.md:9` — corrected truthfully: at `ab45448`, `ProvenanceStrip` still called `useProbe('/api/integrity')` unconditionally, so `Preuves.tsx` passing `integrity={data}` did **not** prevent a second `/api/integrity` fetch. The claim was inaccurate.
`ProvenanceStrip.tsx:25-41` — `useProbe(path, injected)` now returns early (no fetch) when the value is injected; both call sites pass `prop !== undefined`. `ProvenanceStrip.dom.test.tsx:26-64` adds two tests: no fetch when both props injected, and exactly `/api/integrity` + `/api/doctor` fetched when nothing is injected.

**RED→GREEN:** with old `ProvenanceStrip.tsx`, the no-fetch test FAILS (`expected "spy" to not be called at all, but actually been called 2 times`), exit 1; with fix, exit 0.

## Verification — exact commands and exit codes

| Command (cwd) | Exit | Result |
| --- | --- | --- |
| `go -C api test -count=1 ./...` (repo root) | 0 | `ok .../api 120.471s` |
| `go build ./...` (repo root) | 0 | clean |
| `go -C api test -count=1 -run 'TestSchemaEndCorpusParams\|TestLinkDirectionBothIsExecutable\|TestListCalibrationsFallbacks' .` | 0 (RED 1 before fix) | targeted green |
| `npx vitest run Configurer.capabilities.dom.test.tsx ProvenanceStrip.dom.test.tsx --maxWorkers=2` (web/frontend) | 0 (RED 1 before fix) | `2 passed`, `15 tests passed` |
| `npm test -- --maxWorkers=2` (web/frontend) | 0 | `95 passed (95)`, `498 passed (498)` |
| `npm run build` (web/frontend) | 0 | `304.0 KB gz` / `TOTAL gz 320.4 KB` (budget 400 KB) |
| `go build -o web/frontend/cgo-test-server.exe ./cmd/cgo` (repo root, after npm build, before dist restore) | 0 | 16,783,360-byte binary |
| `PWCHROMIUM=...Edge... CGO_API_TOKEN=... npx playwright test e2e/redesign-smoke.spec.ts --reporter=line` | 0 | `1 passed (17.7s)` — unmocked real Go server |
| `PWCHROMIUM=...Edge... CGO_API_TOKEN=... npx playwright test e2e/redesign-layout.spec.ts e2e/redesign-a11y.spec.ts e2e/redesign-feel.spec.ts --reporter=line` | 0 | `19 passed (2.5m)` |
| `git checkout -- web/frontend/dist/index.html` (repo root, AFTER Go rebuild + Playwright) | 0 | tracked dist clean |
| `sh .githooks/run-gates.sh` (repo root, once) | 1 | `3 passed, 1 failed, 2 not run` — pre-existing `FAIL links` (28 missing paths in protected `docs/VERIFY.md`/`docs/CHECKPOINTS.md`) and `SKIP provenance`/`SKIP frozen` (`book/` absent). Not edited, not waived. |

Order respected: `npm run build` → Go rebuild embedding fresh dist → Playwright → restore tracked `dist/index.html`.

**First full npm run:** exit 1 — `BenchConfig.dom.test.tsx > expose les quatre profils sans mode recherche` timed out at 5000 ms under full-suite load. Diagnosis: the file is untouched by this wave; run alone it passes at 2045 ms with maxWorkers=2 (`EXIT_SOLO=0`). Load-sensitive flake, not a regression; the immediate retry passed 498/498. No timeout or gate was changed.

**Commit hook — recorded override (not precautionary).** The normal `git commit -m ...` was attempted FIRST and failed exit 1: the pre-commit hook runs `sh .githooks/run-gates.sh --staged`, which returned `1 passed, 1 failed, 0 not run` — `PASS hygiene` plus the same pre-existing `FAIL links` (28 stale path references inside the protected `docs/VERIFY.md` / `docs/CHECKPOINTS.md`). An independent `sh .githooks/run-gates.sh --staged` confirmed exit 1 with only `links` failing. No gate, permit, regex or protected ledger was edited; the `links` failure is a real gate result, not a skipped check. The commit was therefore made with the documented override `git -c core.hooksPath=C:/Users/ASUS/AppData/Local/Temp/opencode/no-hooks commit ...` (`docs/research/2026-11-22-chunks0-6-gap-audit.md:130-138`), used only because that pre-existing gate blocked the hook. All code/test gates had already passed independently (table above).

## Self-review

- Root cause first (systematic-debugging): traced `both` through `runreq` → `POST /api/run/start` → `cmd/cgo/server.go` chained up+down matrices → frozen per-event `up|down` invariant before editing. No guess-fix.
- Every advertised schema value is executable (`runreq.ValidDirection` cross-check in Go); no fake control, no fabricated direction, value, seal, measurement or PASS.
- Corpus, ledgers, host/bench, gates untouched; protected untracked files (`web/frontend/data/`, `feel-audit.mjs`, `shots-audit/`, `ui-audit.json`, `ui-audit.mjs`) never read into or overwritten by the commit.
- Skipped/not checked: no live bench run or sadcnwp contact (explicitly out of scope); no re-run of the full Playwright suite beyond redesign-smoke/layout/a11y/feel (wave minimum); no change to the pre-existing `links` gate FAIL or `book/` SKIPs.
- Concerns: (1) `npm test -- --maxWorkers=2` has a pre-existing load-sensitive 5 s timeout in `BenchConfig.dom.test.tsx` (passed on retry; not modified — a gate-timeout change would be tampering). (2) The `both` control changes configuration only; it does not itself trigger a run — consistent with §7 ("controls select and persist configuration; they do not claim a measurement until a real run produces one") and with the existing `Configurer` pattern. (3) `final-fix-report.md` lives under `.superpowers/sdd/` which is gitignored; it is force-added by explicit path only.
