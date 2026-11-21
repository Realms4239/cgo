# Chunks 0–6 acceptance gap audit (2026-11-22)

P0 of the chunks 0–6 engineering programme. Read-only audit: this document and
the programme progress ledger are the only writes; no source, gate, frozen
corpus, `docs/VERIFY.md` or `docs/CHECKPOINTS.md` was edited. It consumes the
approved spec `docs/superpowers/specs/2026-11-22-chunks0-6-engineering-design.md`
and the seven MUSTREAD chunk docs `docs/MUSTREAD/10-…` → `16-…`. Every claim
carries a `file:line` or a command with its numeric exit code. This record
promotes no `VERIFY.md` row.

Worktree: `C:/cgo/.worktrees/main-reconciled`, branch `main-reconciled`, base
`5a3caff`. All commands ran in that worktree unless stated.

## Criteria

Verbatim output of the Step 1 command
`for f in docs/MUSTREAD/1[0-6]-chunk-*.md; do echo "=== $f"; grep -nE 'Goal|Produces|Verifies|Expected|Definition of done' "$f"; done`
(exit 0). This is the authoritative criterion list; no criterion was invented.

```
=== docs/MUSTREAD/10-chunk-0-secret-purge-bloat.md
7:- **Goal (concrete result):** Cap-40 PAT gone from history; bloat directories deleted; one binary, one version, one canonical bind address.
57:# Expected output: EMPTY (0 lines)
=== docs/MUSTREAD/11-chunk-1-security-lockdown.md
7:- **Goal (concrete result):** owner-scoped AuthZ, dial-time SSRF guard, `?iface=` regex validation, bounded timeouts and a 60 s audit default.
=== docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md
7:- **Goal (concrete result):** failed scrapes never become zero, `/api/integrity` verifies real files, topology fails closed, RTT is `DelayMs/2` per hop, writes are atomic and ordered.
=== docs/MUSTREAD/13-chunk-3-realtime-pipeline.md
7:- **Goal (concrete result):** eBPF ringbuf → 20 Hz Go SSE ring → Float64Array ring → RingBufferHUD in a 2-mode UI.
=== docs/MUSTREAD/14-chunk-4-honest-tests.md
7:- **Goal (concrete result):** real synctest on product state machines, node + happy-dom split, unmocked Playwright E2E and Axe WCAG gate.
62:   - Verifies `/api/health`, switches cleanly between `Cockpit` and `Labo`, checks `/api/integrity` real manifest status, verifies live SSE `RingBufferHUD` frame progression, triggers a confirmation-guarded (`ArmButton` + `X-Confirm-Action`) action, and persists state cleanly across reload.
=== docs/MUSTREAD/15-chunk-5-container-ci-release.md
7:- **Goal (concrete result):** a working 3-stage Docker build, `--strict` gates with no exit-2 masking, pinned signed release with SBOM and SLSA L2, plus NTP, cert, log and backup basics.
=== docs/MUSTREAD/16-chunk-6-live-proof-freeze.md
7:- **Goal (concrete result):** Linux kernel truth, authorized field trace, `.dist` calibration, RouterOS CHR proof, synchronized thesis, signed v2.0.0 tag and hard freeze.
```

The goal lines above frame the programme; the operational criteria audited below
are the exact markers named in the brief's Steps 2–5 and the chunk 6 readiness
rows. `GapRow` = `{chunk, criterion, state, evidence}`, state ∈ `met|gap|blocked`.
A criterion is `met` only when the symbol **and** its test both exist (Step 2).

## Gap table

| chunk | criterion | state | evidence |
|---|---|---|---|
| 0 | 0.1 Cap-40 PAT purge from all history | blocked | `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md` → exit 0, two commits `54d3f62`, `8c67859`; `docs/CHECKPOINTS.md:3568` — BLOCKED until the human completes the file purge |
| 1 | 1.1 `IsBlockedAddr` dial-time SSRF guard | met | `probe/ssrf.go:59`; test `probe/ssrf_test.go:154` → grep exit 0 |
| 1 | 1.2 `guardedDialContext` shared SSRF dialer | met | `probe/ssrf.go:147`; test `probe/ssrf_test.go:102` → grep exit 0 |
| 2 | 2.1 real `/api/integrity` statuses (`no_manifest`/`missing_file`/`hash_mismatch`) | met | `api/corpus.go:150`, `:165`, `:171`; test `api/corpus_manifest_test.go:61`, `:64`, `:110` → grep exit 0 |
| 2 | 2.2 `writeJSON` encode-before-header | met | `api/server.go:1618`; test `api/server_writejson_test.go:12`, `:23` → grep exit 0 |
| 3 | 3.1 `Float64Array` / `DomainEventRing` rings | met | `grep -rl "Float64Array\|DomainEventRing" web/frontend/src --include=*.ts --include=*.tsx` → exit 0; 9 files incl. `web/frontend/src/lib/ringBuffer.ts` |
| 4 | 4.1 real `synctest` on product state | met | `grep -rl "synctest" api --include=*.go` → exit 0; `api/sse_cadence_test.go` |
| 4 | 4.2 Axe WCAG gate | met | `grep -n axe-core web/frontend/package.json` → exit 0; `web/frontend/package.json:23` (`@axe-core/playwright` 4.10.2) |
| 5 | 5.1 CI + release workflows | met | `ls .github/workflows/ci.yml .github/workflows/release.yml` → exit 0; both present |
| 6 | 6.1.1 environment readiness | met | `docs/verification/2026-11-19-chunk6-readiness/README.md:27` (doctor output lines 8–16) |
| 6 | 6.1.2 race and real-tag suites | gap | `README.md:28` — not run in that pass; Linux `-tags=real` lane remains a named gap |
| 6 | 6.1.3 eBPF-vs-tc confrontation JSON | gap | `README.md:29` — no `--confront-out` flag in `cmd/cgo`; kernel truth sealed at `docs/VERIFY.md` row 43 |
| 6 | 6.1.4 clean dualns matrix | met | `README.md:30` — `run-1794853778`, 24 cells all valid, plus 146-cell exhaustive campaign |
| 6 | 6.2 field audit and `.dist` | met | `README.md:31` — done with caveats; integrity surface `README.md:38-40` (200/159/208/74/246) |
| 6 | 6.3 RouterOS CHR | met | `README.md:32` — VERIFY row 71 promoted by an independent verifier |
| 6 | 6.4 corpus, screenshots, thesis | blocked | `README.md:33` — human-gated; `book/` and `C:\cgo\memoire.docx` stay human-owned |
| 6 | 6.5 v2.0.0 tag and hard freeze | blocked | `README.md:34` — human-gated; no tag, no freeze, no force-push |

### Per-chunk row counts

| chunk | rows | met | gap | blocked |
|---|---|---|---|---|
| chunk 0 | 1 | 0 | 0 | 1 |
| chunk 1 | 2 | 2 | 0 | 0 |
| chunk 2 | 2 | 2 | 0 | 0 |
| chunk 3 | 1 | 1 | 0 | 0 |
| chunk 4 | 2 | 2 | 0 | 0 |
| chunk 5 | 1 | 1 | 0 | 0 |
| chunk 6 | 8 | 4 | 2 | 2 |
| **total** | **17** | **12** | **2** | **3** |

Only rows with `state != "met"` are consumed by later slices: 5 rows —
0.1 (blocked), 6.1.2 (gap), 6.1.3 (gap), 6.4 (blocked), 6.5 (blocked).

## No-change contract

These 12 `met` criteria are already satisfied in the current tree. P1–P3 must
not churn them; each is independently evidenced above:

- 1.1 `IsBlockedAddr` + test
- 1.2 `guardedDialContext` + test
- 2.1 `/api/integrity` statuses + test
- 2.2 `writeJSON` + test
- 3.1 `Float64Array` / `DomainEventRing` rings
- 4.1 `synctest` on product state
- 4.2 Axe WCAG gate
- 5.1 CI + release workflows
- 6.1.1 environment readiness
- 6.1.4 clean dualns matrix
- 6.2 field audit and `.dist` (with its recorded caveats)
- 6.3 RouterOS CHR

## Evidence log (exact commands and exit codes)

All commands below ran in `C:/cgo/.worktrees/main-reconciled` and exited `0`.
A zero exit means the marker was found; absence would have been `gap` with the
same command cited. Test-presence greps enforce Step 2's "symbol **and** test"
rule.

```
grep -rn "func IsBlockedAddr" probe/ssrf.go                    -> exit 0  (line 59)
grep -rn "func guardedDialContext" probe/ssrf.go               -> exit 0  (line 147)
grep -rn "missing_file\|hash_mismatch\|no_manifest" api/corpus.go -> exit 0 (lines 150,165,171)
grep -rn "func writeJSON" api/server.go                        -> exit 0  (line 1618)
grep -rn "IsBlockedAddr" probe --include=*_test.go             -> exit 0  (ssrf_test.go:154+)
grep -rn "guardedDialContext\|SafeDialer" probe --include=*_test.go -> exit 0 (ssrf_test.go:102+)
grep -rn "missing_file\|hash_mismatch\|no_manifest\|verifyManifests" api --include=*_test.go -> exit 0 (corpus_manifest_test.go:61+)
grep -rn "writeJSON" api --include=*_test.go                   -> exit 0  (server_writejson_test.go:12+)
grep -rl "Float64Array\|DomainEventRing" web/frontend/src --include=*.ts --include=*.tsx | head -20 -> exit 0 (9 files)
grep -rl "synctest" api --include=*.go | head -20              -> exit 0  (api/sse_cadence_test.go)
grep -n axe-core web/frontend/package.json                     -> exit 0  (line 23)
ls .github/workflows/ci.yml .github/workflows/release.yml      -> exit 0  (both present)
git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md -> exit 0 (54d3f62, 8c67859)
grep -n "2026-11-18\|purge\|Cap.40\|Cap 40\|ghp_\|human" docs/CHECKPOINTS.md -> exit 0 (ruling at 3564-3568)
grep -c '|' docs/research/2026-11-22-chunks0-6-gap-audit.md    -> exit 0 (38; self-review in task-1-report.md)
```

Chunks 0 and 6 were not locally testable: chunk 0 check 1 is a human history
purge (`docs/CHECKPOINTS.md:3568`: "check 1 BLOCKED until the human completes
the file purge. Chunk 0 agent work is complete but the chunk is NOT verified";
"no agent may rewrite history"). Chunk 6 rows are transcribed from the
2026-11-19 readiness record; no live host, bench, sadcnwp, frozen write or
thesis build was touched.

The brief's "record why" for the no-hooks override: the normal commit hook
failed exit 1 on the pre-existing `links` false positive — 28 stale path
references inside the protected `docs/VERIFY.md` and `docs/CHECKPOINTS.md`.
No gate, permit, regex or protected ledger was edited; the `links` failure is
a real gate result, not a skipped check. The commit was therefore made with the
documented `core.hooksPath=C:/Users/ASUS/AppData/Local/Temp/opencode/no-hooks`
override. The full command list and exit codes are in `task-1-report.md`.
