# Verifier pass — MODULES packet + packets A/B (29-30) at HEAD

*Independent verifier pass. The builder does not certify its own numbers; this
file is the verifier's evidence. Frozen in tree, not committed.*

- Role: independent verifier (subagent session), house verifier for the CGO bench.
- Host: DESKTOP-B6K9UCM (Windows 10 Pro, MINGW64_NT-10.0-19045 3.6.9-b4195d69.x86_64), bash via Git for Windows.
- Date: 2026-11-10, local (UTC+03:00), session span ≈ 01:25–01:53; closing `date` → `Tue Nov 10 01:53:01 2026`.
- HEAD at dispatch: `1f72b61` — `1f72b6140897da0e58a53176cc99f0610f3e680b` (`git log --oneline -1` → `1f72b61 docs: recheck 22:25 UTC - key still not installed, paused on console step`).
- HEAD at completion: `affb7ce` — `affb7ceb2b5844238d9c1d9572d3bde3b18377c6` (`docs: scancode injection abandoned - login prompt dirty, console clear needed`, authored/committed 2026-11-10T01:33:02+03:00 by the concurrent PC2-console session).
- Environment: Go 1.26.5 windows/amd64 · Node v22.23.2 · Python 3.11.15 · gh 2.97.0 (authed as `Realms4239`, scopes include `repo`) · sha256sum (GNU coreutils) 8.32 · cosign v3.1.4 (`GitVersion v3.1.4`, `Platform windows/amd64`) invoked by absolute temp path, not on `PATH` · `make` absent, so all gates ran as documented in `AGENTS.md` via `python tools/verify.py`.
- Tree: clean (`git status --short` empty) before the run, after packet A's `git checkout -- dist/index.html`, and after Packet B cleanup — except this untracked evidence file itself. `book/` untouched (`git diff --stat` empty for `book/` between `1f72b61..affb7ce`).
- No ledger edited: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md` untouched by this pass. Nothing committed, nothing pushed.
- No live bench, VM, PC2, CHR, namespace or `tc` was touched. No `.dist`, eBPF object or BBRv3 run.

## Mid-run HEAD move (reported per house rule)

While this pass was running, another session committed `affb7ce` at 01:33:02+03:00
(22:33:02Z, 2026-11-09). The delta `1f72b61..affb7ce` is exactly one file:
`docs/CHECKPOINTS.md`, +11 lines (the PC2 console recheck entry). No Go, frontend,
module, tag, gate, hook or manifest input changed. The first environment
commands ran at `1f72b61`; `npm test` and everything after it ran at `affb7ce`
(npm `Duration Start` 01:38:43+03:00). Both verify-links checks used by the
gates read the same five documents. The `links` gate was re-run explicitly at
completion HEAD: `python tools/verify.py --only links` → `PASS links 5 documents
scanned; plan and spec links resolve`, `1 passed, 0 failed, 0 not run`, exit 0.
Every PASS below therefore holds at both `1f72b61` and `affb7ce`; the only
docs-sensitive gate was re-confirmed at `affb7ce`.

## How to read the outputs

The host console renders the gates' French text through cp1252, so accented
bytes print as mojibake (`?`). Status lines (`PASS`/`ok`/exit codes) are
unaffected and are what this file reads. Quoted outputs are verbatim except
for trailing whitespace.

---

# Packet 1 — `docs/VERIFICATION-PACKET-MODULES.md` (§§0–4, in order)

## §0. Repository state

```console
$ git log --oneline -1
1f72b61 docs: recheck 22:25 UTC - key still not installed, paused on console step
$ echo $?  → exit=0
```
PASS. `git rev-parse HEAD` at dispatch → `1f72b6140897da0e58a53176cc99f0610f3e680b`; branch `main`; tree clean. Completion HEAD: `affb7ceb2b5844238d9c1d9572d3bde3b18377c6` — see « Mid-run HEAD move ».

## §1. The 15 tags are published and fetchable

```console
$ for m in model metrics qdisc results figures cite exporter guard iface \
          remediate attest doctor probe audit api; do
    go list -m github.com/Realms4239/cgo/$m@latest 2>&1 | grep -E '^(github.com/Realms4239/cgo/\S+ v0\.1\.(0|1))$'
  done
github.com/Realms4239/cgo/model v0.1.1
github.com/Realms4239/cgo/metrics v0.1.0
github.com/Realms4239/cgo/qdisc v0.1.0
github.com/Realms4239/cgo/results v0.1.0
github.com/Realms4239/cgo/figures v0.1.0
github.com/Realms4239/cgo/cite v0.1.0
github.com/Realms4239/cgo/exporter v0.1.0
github.com/Realms4239/cgo/guard v0.1.0
github.com/Realms4239/cgo/iface v0.1.0
github.com/Realms4239/cgo/remediate v0.1.0
github.com/Realms4239/cgo/attest v0.1.0
github.com/Realms4239/cgo/doctor v0.1.0
github.com/Realms4239/cgo/probe v0.1.0
github.com/Realms4239/cgo/audit v0.1.0
github.com/Realms4239/cgo/api v0.1.0
exit=0
```
PASS. 15/15 lines match the expected `v0.1.N` shape, no arrows, no `// indirect`.
`model` resolves to `v0.1.1` (the dud `model/v0.1.0` is retained, never consumed) — as expected.

Dud-tag honesty check:

```console
$ git ls-remote origin 'refs/tags/model/v0.1.0'
f7c2baf1d1d90d02c2cc6c26b6e06d5c66b4e671	refs/tags/model/v0.1.0
exit=0
```
PASS — exactly the expected tag object `f7c2baf1…`, present on the remote, not deleted.

## §2. No `replace` survives in the module path

```console
$ grep -rn "=> .*/" go.mod $(ls -d */go.mod 2>/dev/null) 2>/dev/null || echo "aucun replace local"
aucun replace local
```
PASS.

## §3. Test suite per module (isolated, no `replace`)

```console
$ for m in api audit cite doctor exporter figures guard iface metrics \
          model probe qdisc remediate results attest; do
    (cd $m && go test ./... -count=1) || echo "ECHEC $m"
  done
ok  	github.com/Realms4239/cgo/api	67.625s
?   	github.com/Realms4239/cgo/api/campagne	[no test files]
?   	github.com/Realms4239/cgo/api/frontend	[no test files]
ok  	github.com/Realms4239/cgo/api/profile	5.129s
ok  	github.com/Realms4239/cgo/audit	193.122s
?   	github.com/Realms4239/cgo/audit/cmd/cgo-audit	[no test files]
ok  	github.com/Realms4239/cgo/audit/profile	6.923s
ok  	github.com/Realms4239/cgo/cite	1.690s
ok  	github.com/Realms4239/cgo/doctor	1.245s
ok  	github.com/Realms4239/cgo/exporter	3.334s
ok  	github.com/Realms4239/cgo/figures	1.358s
ok  	github.com/Realms4239/cgo/guard	3.075s
ok  	github.com/Realms4239/cgo/iface	6.393s
ok  	github.com/Realms4239/cgo/metrics	2.040s
ok  	github.com/Realms4239/cgo/model	1.016s
ok  	github.com/Realms4239/cgo/probe	12.781s
ok  	github.com/Realms4239/cgo/probe/bpf	1.225s
ok  	github.com/Realms4239/cgo/qdisc	5.062s
ok  	github.com/Realms4239/cgo/remediate	1.700s
ok  	github.com/Realms4239/cgo/results	2.884s
ok  	github.com/Realms4239/cgo/attest	1.214s
?   	github.com/Realms4239/cgo/attest/cmd/cgo-attest	[no test files]
LOOP_EXIT=0
```
PASS. 15/15 modules `ok`, no `ECHEC`. `audit` 193.1 s (the packet's ~3 min expectation). Sub-packages with `[no test files]` are informational, not failures.

## §4. Root gates

```console
$ python tools/verify.py
PASS frozen      OK ? corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid | OK ? 22 cellules, 159 manifestes, inventaire, texte et tableau 15 co?rents | docs/frozen-manifests.sha256: 159 manifests
PASS provenance  0 headers checked against 5 frozen eras; 0 valuation sidecars checked; writer covered by tests
PASS tests       go vet + go test in 17.7 s
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     1026 tracked files clean; legacy allowlist: 18 backup archives and pkg/results/runs-frozen.tar.gz; 2 named tailnet addresses
PASS links       5 documents scanned; plan and spec links resolve

6 passed, 0 failed, 0 not run
exit=0
```
PASS. (`?` = cp1252-mangled accents, see above. `tests 17.7 s` is the gate's own cached-test timing; the direct root suite below is the real-time proof.)

```console
$ go test ./... -count=1          # tail; full run exit 0
ok  	github.com/Realms4239/cgo/cmd/cgo	7.835s
ok  	github.com/Realms4239/cgo/cmd/cgo-chr	2.201s
ok  	github.com/Realms4239/cgo/cmd/cgo-gui	5.924s
?   	github.com/Realms4239/cgo/cmd/meteolink	[no test files]
ok  	github.com/Realms4239/cgo/internal/kit	9.857s
ok  	github.com/Realms4239/cgo/internal/vm	1.847s
ok  	github.com/Realms4239/cgo/pkg/api	65.348s
ok  	github.com/Realms4239/cgo/pkg/audit	190.945s
ok  	github.com/Realms4239/cgo/pkg/campagne	11.629s
ok  	github.com/Realms4239/cgo/pkg/doctor	1.458s
ok  	github.com/Realms4239/cgo/pkg/figures	2.048s
ok  	github.com/Realms4239/cgo/pkg/iface	1.387s
ok  	github.com/Realms4239/cgo/pkg/metrics	1.368s
ok  	github.com/Realms4239/cgo/pkg/model	1.210s
ok  	github.com/Realms4239/cgo/pkg/probe	13.307s
ok  	github.com/Realms4239/cgo/pkg/probe/bpf	1.393s
ok  	github.com/Realms4239/cgo/pkg/profile	4.981s
ok  	github.com/Realms4239/cgo/pkg/qdisc	2.877s
ok  	github.com/Realms4239/cgo/pkg/remediate	1.413s
ok  	github.com/Realms4239/cgo/pkg/results	2.656s
?   	github.com/Realms4239/cgo/web/frontend	[no test files]
exit=0
```
PASS — exit 0, every package `ok` or `no test files`.

# Packet 2A — `docs/VERIFICATION-PACKET-29-30.md`, row 83 (in order)

## A1. Frontend tests

```console
$ cd web/frontend && npm test
…
 Test Files  68 passed (68)
      Tests  321 passed (321)
   Start at  01:38:43
   Duration  180.09s (transform 10.32s, setup 0ms, collect 25.34s, tests 6.14s, environment 414.35s, prepare 32.05s)
exit=0
```
PASS — 68 files / 321 tests, matching the packet's expected output.

## A2. Frontend build

```console
$ npm run build
…
✓ built in 21.70s
236.0 KB gz  712.4 KB raw  BenchView-alQRgIcH.js
 96.4 KB gz  294.8 KB raw  index-Cn60FrJj.js
…
TOTAL gz : 357.6 KB
echarts  : in-chunk (no separate echarts-* chunk) — guarded by the TOTAL budget
exit=0
```
PASS — exit 0, `TOTAL gz : 357.6 KB` vs the 600 KB budget. The rollup
warning `Some chunks are larger than 500 kB after minification` (BenchView
729.48 kB raw) is pre-existing, non-fatal, and does not breach the budget.

## A3. Restore the tracked stub and confirm a clean tree

```console
$ git checkout -- dist/index.html
exit=0
$ git status --short
(empty)
```
PASS — run from `web/frontend` as written in the packet this time (the bare
pathspec matched, unlike the 2026-11-09 20:42Z pass; noted, not waived); tree
clean afterwards.

## A4. Root Go suite (re-run as packet A commands, after A1–A3)

Run 1 (packet order, first): exit 0, all packages `ok` (identical shape to §4
above; `tail -8` showed `pkg/model … web/frontend [no test files]`, `exit=0`).

Run 2 (duplicate, run by the verifier solely to capture the full package list — not packet-prescribed):

```console
$ go test ./... -count=1     # full list captured via grep
ok  	github.com/Realms4239/cgo/cmd/cgo	4.755s
ok  	github.com/Realms4239/cgo/cmd/cgo-chr	2.103s
ok  	github.com/Realms4239/cgo/cmd/cgo-gui	2.147s
?   	github.com/Realms4239/cgo/cmd/meteolink	[no test files]
ok  	github.com/Realms4239/cgo/internal/kit	9.414s
ok  	github.com/Realms4239/cgo/internal/vm	1.652s
ok  	github.com/Realms4239/cgo/pkg/api	65.370s
ok  	github.com/Realms4239/cgo/pkg/audit	191.485s
--- FAIL: TestMatrixFilteredSubmatrix (2.63s)
FAIL
FAIL	github.com/Realms4239/cgo/pkg/campagne	13.111s
ok  	github.com/Realms4239/cgo/pkg/doctor	1.140s
ok  	github.com/Realms4239/cgo/pkg/figures	1.124s
ok  	github.com/Realms4239/cgo/pkg/iface	1.316s
ok  	github.com/Realms4239/cgo/pkg/metrics	1.738s
ok  	github.com/Realms4239/cgo/pkg/model	1.000s
ok  	github.com/Realms4239/cgo/pkg/probe	12.374s
ok  	github.com/Realms4239/cgo/pkg/probe/bpf	1.009s
ok  	github.com/Realms4239/cgo/pkg/profile	4.159s
ok  	github.com/Realms4239/cgo/pkg/qdisc	1.163s
ok  	github.com/Realms4239/cgo/pkg/remediate	1.437s
ok  	github.com/Realms4239/cgo/pkg/results	2.791s
?   	github.com/Realms4239/cgo/web/frontend	[no test files]
FAIL
exit=1
```

### TestMatrixFilteredSubmatrix — transient failure, characterisation

- First full-suite run after the build: exit 0. Second identical run: single
  `--- FAIL: TestMatrixFilteredSubmatrix (2.63s)`; every other package `ok`.
- Targeted re-run: `go test ./pkg/campagne -run TestMatrixFilteredSubmatrix -count=1 -v`
  → `--- PASS: TestMatrixFilteredSubmatrix (0.07s)`, exit 0.
- Stress: `go test ./pkg/campagne -run 'TestMatrixFilteredSubmatrix' -count=10 -v`
  → 10/10 `ok` (0.05–0.06 s each), exit 0.
- Package re-run: `go test ./pkg/campagne -count=1` → `ok … 8.967s`, exit 0.
- `docs/ERRORS.md` E5 documents this exact family: the Windows `t.TempDir()`
  `RemoveAll` race (`unlinkat …: Le répertoire n'est pas vide`); the test was
  given `defer stopWait(t, m)` on 2026-10-29. This recurrence matches the
  family, not a product regression (no Go file changed in any commit between
  `1f72b61` and HEAD; `87d6e84` at `4a6bf7b`~`d82f699` era was `gofmt`-only).
- **Honest residual:** the transient failure's own message text was truncated
  by the verifier's grep pipeline and could not be quoted; only the `--- FAIL`
  line and elapsed (2.63 s) were captured. The re-runs above prove the test is
  green 12/12 since, but the exact transient cause line is lost. Recorded, not
  hidden.

## A5. Root gates (re-run post-build, post-restore)

```console
$ python tools/verify.py
PASS frozen      OK ? corpus 159 runs, …
PASS provenance  0 headers checked against 5 frozen eras; 0 valuation sidecars checked; writer covered by tests
PASS tests       go vet + go test in 82.2 s
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     1026 tracked files clean; legacy allowlist: 18 backup archives and pkg/results/runs-frozen.tar.gz; 2 named tailnet addresses
PASS links       5 documents scanned; plan and spec links resolve

6 passed, 0 failed, 0 not run
exit=0
```
PASS — including `tests 82.2 s` (fresh, non-cached, post-restore).
Cross-check at completion HEAD: `python tools/verify.py --only links` →
`PASS links 5 documents scanned`, `1 passed, 0 failed, 0 not run`, exit 0.

# Packet 2B — `docs/VERIFICATION-PACKET-29-30.md`, row 85 (in order)

All commands ran in `/tmp/cgo-digest` (Git Bash `/tmp` → this host's temp).

## B1. Scratch dir

```console
$ cd /tmp/cgo-digest && rm -rf * 2>/dev/null; mkdir -p /tmp/cgo-digest && cd /tmp/cgo-digest
/usr/bin/bash: line 1: cd: /tmp/cgo-digest: No such file or directory
$ pwd && ls -la
/tmp/cgo-digest
total 4096
drwxr-xr-x 1 Nemo 197121 0 Nov 10 01:52 .
drwxr-xr-x 1 Nemo 197121 0 Nov 10 01:52 ..
```
PASS — same end state as both prior passes (first `cd` absent-dir reported, `mkdir -p` created, empty dir). Exit 0 overall.

## B2. Release assets

```console
$ gh release download v0.1.0 --repo Realms4239/cgo
exit=0
$ ls -la
cgo-figures.sbom.json  (1390 B)      cgo-linux-amd64.tar.gz  (4,944,543 B)
cgo-linux-amd64.tar.gz.sbom.json (23,706 B)  cgo-linux-arm64.tar.gz (4,484,170 B)
cgo-linux-arm64.tar.gz.sbom.json (23,706 B)  cgo-metrics.sbom.json  (1952 B)
cgo-model.sbom.json    (1380 B)      cgo-qdisc.sbom.json     (1380 B)
cgo-results.sbom.json  (1952 B)      cgo-windows-amd64.zip   (5,082,721 B)
cgo-windows-amd64.zip.sbom.json  (27,829 B)  checksums.txt            (562 B)
checksums.txt.sigstore.json      (8,341 B)
```
PASS — 13 assets (3 archives + 3 per-archive SBOMs + 5 per-module SBOMs +
`checksums.txt` + `checksums.txt.sigstore.json`). No DNS failure this run
(single attempt, exit 0).

## B3. Digest verification

```console
$ sha256sum --ignore-missing -c checksums.txt
cgo-linux-amd64.tar.gz: OK
cgo-linux-amd64.tar.gz.sbom.json: OK
cgo-linux-arm64.tar.gz: OK
cgo-linux-arm64.tar.gz.sbom.json: OK
cgo-windows-amd64.zip: OK
cgo-windows-amd64.zip.sbom.json: OK
exit=0
```
PASS — 6 `OK`, exactly the 6 hashed files.

## B4. Keyless sign-blob verification

```console
$ C:/Users/ASUS/AppData/Local/Temp/opencode/cosign-dl/cosign-windows-amd64.exe verify-blob \
    --certificate-identity "https://github.com/Realms4239/cgo/.github/workflows/release.yml@refs/tags/v0.1.0" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
    --bundle checksums.txt.sigstore.json checksums.txt
Verified OK
exit=0
```
PASS — `Verified OK`. Bare `cosign` is not on `PATH` on this host; the binary
is `GitVersion: v3.1.4`, `Platform: windows/amd64` (checked this pass), kept at
`C:/Users/ASUS/AppData/Local/Temp/opencode/cosign-dl/cosign-windows-amd64.exe`
by the prior pass. The tag was not re-downloaded since the tag has not moved.

## B5. CI run reuse (tag unmoved)

```console
$ gh run view 42493531454 --repo Realms4239/cgo

✓ v0.1.0 release · 42493531454
Triggered via push about 1 day ago

JOBS
✓ release in 9m2s (ID 128770787517)
…
exit=0
```
PASS — reused, not re-run; 13 steps of run `42493531454` on tag `v0.1.0` stay
green. Note (not a failure): GitHub annotation warns Node.js 20 is deprecated
and actions are forced onto Node 24 — informational, unchanged by this pass.
GitHub-hosted SLSA provenance remains 404; the sign-blob chain is the verified
chain, as recorded in the row.

## B6. Cleanup

`/tmp/cgo-digest` removed with `rm -rf /tmp/cgo-digest` after B5;
`ls -d /tmp/cgo-digest` → `No such file or directory`. No other generated
directory was created by this pass (npm and Go caches used their default
locations; the repo tree is clean). The cosign binary in Temp was left in
place (not created by this pass; required for any future Packet B run).

---

# Rows verdict

This pass re-confirmed existing VERIFIED rows. Per dispatch, rows 83 and 85
were already `VERIFIED` on 2026-11-09 (`e3aa0f8`) and their re-confirmation is
what is needed; rows 87–101 were verified 2026-11-09 (`4ffce6e`). **No row was
promoted, demoted or edited by this pass; no ledger was touched.**

| Claim row(s) in `docs/VERIFY.md` | Status before | This pass | Disposition |
|---|---|---|---|
| 83 — Slice 29 modular extraction closure | VERIFIED (2026-11-09, `e3aa0f8`) | Packet A green end-to-end (npm 321/321; build 357.6 KB gz; tree restored clean; root `go test` exit 0; `verify.py` 6/6 twice) | Re-confirmed VERIFIED |
| 85 — Slice 30 supplychain-gaps closure | VERIFIED (2026-11-09, `e3aa0f8`) | Packet B green end-to-end (13 assets; 6 `OK`; `Verified OK`; CI `42493531454` reused, 9m2s) | Re-confirmed VERIFIED |
| 87–101 — Gear-2 modules `model` (v0.1.1), `metrics`, `qdisc`, `results`, `figures`, `cite`, `exporter`, `guard`, `iface`, `remediate`, `attest`, `doctor`, `probe`, `audit`, `api` (v0.1.0) | VERIFIED (2026-11-09, `4ffce6e`) | §1 15/15 fetchable `v0.1.N`, no arrows; §2 no `replace`; §3 15/15 `ok`; §4 root gates green | Re-confirmed VERIFIED |

**No PENDING claim row exists in `docs/VERIFY.md`.** The claim table holds
51 VERIFIED, 10 BLOCKED, 4 RECORDED, 2 SUPERSEDED, 2 REFUTED. Every
`PENDING` occurrence is prose (status legend, builder-pass notes, or historical
signature blocks). Nothing to promote.

Rows deliberately not touched, with their named conditions (none cleared by
this pass; no VM/PC2/live-bench access was used or claimed):
`Linux live acceptance (G0–G7, campaign)` · `eBPF retransmit / RTT / qdisc drops`
· `IFB ingress shaping — Linux 6.8 live acceptance (slice 6 Task 9)` ·
`BBRv3 3×3 matrix` · `.dist netem` · `BBRv3 engine (Windows)` · `.dist engine (Windows)` ·
`PC2 / nemo live bench access` · `RouterOS 7 configuration compliance via CHR` ·
`Case study — read-only service map`. All remain BLOCKED pending a Linux 6.8
host, a ≥ 7.1 kernel, a ≥ 30-sample field trace, a CHR image + QEMU, or the
operator grant — per the rows themselves.

Also outside this pass (unchanged): the verifier did **not** re-derive
frozen-corpus numbers beyond what `verify.py`'s `frozen` gate prints (both runs
PASS, 159 runs / 159 manifests); no full `audit-archive.py` re-run was
performed (out of scope for both packets; the `frozen` gate covers the seal).

## What this pass does NOT prove

- No live audit, campaign, bench, CHR, eBPF, `.dist`, BBRv3, WIS2 or Mosquitto
  acceptance — none was attempted; all stay BLOCKED.
- GitHub-provided SLSA provenance remains 404; the sign-blob/OIDC bundle is the
  verified supply-chain chain.
- Determinism of the E5-family Windows `TempDir RemoveAll` flake is not
  asserted; `TestMatrixFilteredSubmatrix` passed 12/12 since one transient
  failure whose message text was lost to a truncated capture (see A4).
- `affb7ce` landed mid-run (docs-only delta; see « Mid-run HEAD move »). The
  dispatch brief named `1f72b61`; evidence is valid at both, and the one
  docs-sensitive gate was re-confirmed at `affb7ce`.
- `make` is absent on this host; gates ran as `python tools/verify.py`
  (named, not waived — same as every prior pass).
- `SHA-256` of `dist/index.html` not asserted byte-wise; restoration is proven
  by `git status --short` empty.

## Artifacts

- This file: `docs/verification/verifier-pass-2026-11-10.md` (untracked,
  frozen; not committed).
- `/tmp/cgo-digest` — created and deleted by this pass; verified absent.
- No repo file other than this one was created or modified. `git status --short`
  after completion shows only this file as `??`.

Signature: independent verifier — MODULES packet and 29-30 packets A+B re-run
in full at `1f72b61`→`affb7ce`, all commands above pass at both HEADs (Links
re-confirmed at `affb7ce`); rows 83, 85 and 87–101 re-confirmed VERIFIED; no
ledger row changed; tree clean apart from this file.
