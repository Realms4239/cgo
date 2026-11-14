# Product TODO triage — 2026-11-14 (Task 0.6 Part B)

*Adjudication pass, not a code-churn pass. No code changed; this document is the deliverable.*

## The rule (product = what ships)

In scope (product): `pkg/`, `cmd/`, `web/frontend/src`, `kit/`, `tools/`, `internal/`.
Out of scope: `docs/` prose, plans and specs (including this file's neighbours), the frozen
corpus and its tooling (`book/`), the three ledgers (`docs/VERIFY.md`, `docs/CHECKPOINTS.md`,
`docs/ERRORS.md`), gates' self-test fixtures, test-file strings that merely assert marker text
(they die with their assertions), and dependency/lockfile/binary noise (`node_modules/`,
`go.sum` hashes, `bun.lock` base64, `*.woff2`, `*.png`).

Each marker found gets exactly one disposition:

- **implement** — only a one-line, obvious, already-test-covered fix (e.g. an ancient stale
  comment referencing a removed file).
- **delete** — only unreferenced/dead code whose deletion is safe; anything wider than the
  marker's own file is ledgered instead.
- **ledger** — everything else; each row names what the marker owes.

## How the enumeration was done (and what it found)

Command run on 2026-11-14 at `main` @ `bba4a80` (working tree carries 75 WIP paths from other
work; product files among them are listed below where relevant):

```
git grep -n -E 'TODO|FIXME|XXX|HACK' -- pkg cmd web kit tools internal
```

Output (verbatim, 2 lines, both false positives of the `XXX` pattern):

```
cmd/cgo/main.go:414:        cgo calibrate --audit audit-XXXX --profile P2   # trace baseline → .dist netem
kit/stress_testbed.sh:29: TMP="$(mktemp -d "${TMPDIR:-/tmp}/cgo-stress.XXXXXX")"
```

`cmd/cgo/main.go` line number refers to the working tree (HEAD carries it at line 412–414
depending on WIP hunks; the file is WIP-dirty with the human's uncommitted eBPF work and was
not touched — §S applies).

Broader sweeps, all clean (text files only):

- `grep -rniI -E 'TODO|FIXME|XXX|HACK'` over the whole repo minus `node_modules/`, `.git/`,
  `book/`, `backup/`, `docs/`, `.superpowers/`, `dist/`, `test-results/` → only the two hits
  above plus non-marker noise: `cmd/cgo/purge_test.go:178`
  (`TestPurgeNothingToDoIsZero`, a test name), `go.mod:53-54` (`hack-pad` dependency names),
  `go.sum` (hashes), French strings `rien à faire` in `pkg`-adjacent CLI output and kit logs,
  and `internal/kit/vnet.go` (`-GwIp` flag). None of these are work-item markers.
- `git grep -n -i --text -E 'todo|fixme|hack|xxx' HEAD` → no further product-scope hits;
  `web/frontend/src`, `web/frontend/e2e`, `web/frontend/scripts`, `.github/`, `Makefile`,
  `.githooks/` → 0 hits.
- Untracked working-tree product files (`cmd/cgo/kernel_ebpf_linux.go`,
  `cmd/cgo/kernel_ebpf_other.go`, `probe/bpf/loader_linux.go`, `kit/autoinstall/`) → 0 hits.
- The case-insensitive sweep also surfaced one genuine `TBD` in
  `cgo/credentials/README.md:8`; `cgo/` sits outside the brief's enumeration paths but ships in
  the repository, so it is triaged here as a flagged extra (not counted in the census below).

**The 2026-11-13 census figure (102 repo-wide markers, 37 product-code, 33 in pkg/cmd/web,
zero in core logic) is not reproducible at `bba4a80`.** The tree — HEAD and working tree,
product paths included — contains zero genuine `TODO`/`FIXME`/`HACK` markers. Likely
explanation: the census predates the landed dead-surface deletions (Task 0.6a: `cmd/cgo-gui`,
`npm/` removed 2026-11-14, commit `9a9a4b5`) and other in-flight cleanups, or it counted
dependency/lockfile/binary matches as product. Either way this is recorded as residual risk,
not silently absorbed: the triage below is the complete, reproducible 2026-11-14 state.

## Triage table (every product-scope candidate found)

| path:line | marker text (verbatim) | disposition | reason (one sentence) |
|---|---|---|---|
| `cmd/cgo/main.go:414` (working tree) | `cgo calibrate --audit audit-XXXX --profile P2   # trace baseline → .dist netem` | **ledger** | Not a work item — `XXXX` is the literal audit-id placeholder in user-facing help text, and the file is WIP-dirty (human's uncommitted eBPF work; edits forbidden here), so the row is recorded rather than changed. |
| `kit/stress_testbed.sh:29` | `TMP="$(mktemp -d "${TMPDIR:-/tmp}/cgo-stress.XXXXXX")"` | **ledger** | Not a work item — `XXXXXX` is the `mktemp` template for the throwaway directory; rewriting it would change behaviour for zero benefit. |
| `cgo/credentials/README.md:8` (out of enumeration scope; flagged) | `- VM guest user: TBD (console install required per \`docs/usage.fr.md\`: \`mkdir -p ~/.ssh && echo '<pubkey>' >> ~/.ssh/authorized_keys\`)` | **ledger** | Genuine documentation `TBD` — the value (VM guest user) exists only on the operator's console-installed host and in `docs/usage.fr.md`; it cannot be filled from the repo without inventing a fact, so it owes the operator the console-install identity verbatim. |

## Disposition counts

- **implement: 0** — no candidate met the bar (one-line, obvious, test-covered fix).
- **delete: 0** — no candidate was dead/unreferenced code.
- **ledger: 3** — two enumeration false positives (WIP file / mktemp template) and one
  flagged out-of-scope documentation `TBD`.
- Product markers before: 0 genuine (2 enumeration hits, both false positives). After:
  unchanged (docs-only commit; no code edited).

## Residual risk

1. The 2026-11-13 census (37 product-code markers) could not be reproduced or accounted for;
   if a product marker class existed then (e.g. inside deleted `cmd/cgo-gui`/`npm/` trees or a
   differently-scoped sweep), those candidates are gone with those trees and were never
   individually adjudicated. The enumeration re-run for this task is the authority recorded
   here.
2. `cmd/cgo/main.go` is WIP-dirty; line numbers above match the working tree, not `HEAD`
   (`audit-XXXX` sits at 412–414 depending on the WIP hunks). The file was not edited.
3. `cgo/credentials/README.md` ships a `TBD` and an instruction to add the directory to
   `.gitignore` before any commit; this triage does not re-adjudicate those items (out of
   scope) but flags them: no private material is claimed to exist in the repo — the file
   itself states the directory is meant to stay untracked.
4. No product marker was implemented or deleted, so the six-gate suite is expected green;
   the gates re-run for this task, and the exit code is recorded in the task report.
