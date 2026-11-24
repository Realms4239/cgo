# Enterprise production modernisation — Task 4 verification (2026-11-24)

Authority: approved plan `docs/superpowers/plans/2026-11-24-enterprise-production-modernisation-plan.md`,
Task 4 brief, and design §5 (`docs/superpowers/specs/2026-11-24-enterprise-production-modernisation-design.md`).
Pass: builder (one SDD implementer, no parallel workers). Scope: the two
documents under this directory only. No gate, CI, ledger, permit, corpus, book,
thesis or protected-config edit; no live host or token work; no new dependency.

## RED → GREEN (one line)

RED: the doc test run against a copy with one live-operation row removed and one
rollback cell blanked exited 1 (`missing=1 bad_cells=1`, naming
`APPROVED-2026-11-24-stack-debug.md` as uncovered) → GREEN: the same test against
`runbooks/index.md` exited 0 (`discovered=28 rows=28 indexed=28 missing=0 bad_cells=0`),
so no operation is silently omitted and every gap is explicit.

## What landed

- `runbooks/index.md` — one row per named live operation (`APPROVED-*` token),
  with operation, scope, rollback, last rehearsal date and evidence/status.
  28 tokens re-discovered from the repository's own `docs/**` and `.superpowers/**`
  and indexed one-for-one, including draft and proposal-only tokens.
  A **Gap register** names every `ABSENT` rollback, every `NOT REHEARSED`
  rehearsal, the three draft/proposal tokens and the three read-only R0 tokens.
  Suffix-only token references (`-late`) are quoted, never expanded into an
  invented token id.
- This README — the coverage doc test (below), the supply-chain declarations
  read from the repository, the honest gate result and the residual gaps.

No rehearsal date, token id, rollback proof or CI result is invented. Where a
repo document records no rehearsal, the cell says `NOT REHEARSED`; where it
records no rollback line, the cell says `ABSENT`; where an operation is read-only,
the cell says `N/A — read-only R0`.

## Supply-chain and release declarations (read from the repo, not run here)

These are facts declared in workflow files, read during this pass. **No CI run
was executed or observed by this builder pass; no CI result is asserted.**

- `.github/workflows/ci.yml` — `gates` job runs `sh .githooks/run-gates.sh --strict`
  (a named `SKIP` counts as failure and is never rewritten green), `go build ./...`
  and `go vet ./...`. Other jobs: `bpf` builds the object with clang-18 on
  ubuntu-24.04 and prints its sha256 — the pin itself stays a human decision;
  `frontend` uses `bun install --frozen-lockfile`, typecheck, build and vitest;
  `docker-build` smoke-tests the served container and asserts `GET /api/results`
  returns 401 without a token; `fresh-clone` runs `kit/test-fresh-clone.sh`;
  `secret-scan` runs `gitleaks/gitleaks-action@v2`.
- `.github/workflows/release.yml` — `goreleaser check`, gates `--strict`,
  `goreleaser release --clean`, per-module CycloneDX SBOMs from `syft@v1.24.0`,
  publication via `softprops/action-gh-release@v2`, build-provenance attestation
  via `actions/attest-build-provenance@v1` over `dist/checksums.txt`, and a
  conditional `cosign@v2.4.0 attest-blob` over `dist/digests.txt` when the
  `CGO_IMAGE_DIGEST` secret is set. `attestations: write` is declared.
- Actions are pinned to **major-version tags** (`@v4`, `@v5`, `@v2`, `@v1`); Go
  tools are pinned to exact versions. Full commit-SHA pinning of actions is not
  declared in these workflows — recorded as a residual, not claimed as done.
- Release versioning stays human-gated (design §5). `/api/capabilities` is
  registered in `api/server.go`; this doc-only slice did not re-verify its
  runtime version/provenance payload and makes no claim about it.

## Doc test (reproducible, embedded — no third file)

Run from the repository root with the index at its committed path:

```bash
awk '/^```python$/{f=1;next} /^```$/{if(f)exit} f' \
  docs/verification/2026-11-24-enterprise-production/README.md | python -
```

The program (extract between the markers) is:

<!-- # --- DOC TEST BEGIN --- -->
```python
#!/usr/bin/env python3
"""Doc test T4: the runbook index must cover every live operation token named
in the repository's own documents, with no silent omission.

Usage: python check-index.py [path/to/index.md]
Exit 0 = complete. Exit 1 = a gap or omission. Exit 2 = input unreadable.
"""
import re
import sys
from pathlib import Path

root = Path.cwd()
default = root / "docs/verification/2026-11-24-enterprise-production/runbooks/index.md"
index_path = Path(sys.argv[1]) if len(sys.argv) > 1 else default
new_dir = root / "docs/verification/2026-11-24-enterprise-production"
TOKEN = re.compile(r"APPROVED-\d{4}-\d{2}-\d{2}-[A-Za-z0-9_-]+\.md")

discovered = set()
for base in (root / "docs", root / ".superpowers"):
    for p in base.rglob("*"):
        if not p.is_file() or p.suffix not in (".md", ".txt"):
            continue
        if str(p.resolve()).startswith(str(new_dir.resolve())):
            continue
        discovered.update(TOKEN.findall(p.read_text(encoding="utf-8", errors="replace")))

try:
    body = index_path.read_text(encoding="utf-8")
except OSError as exc:
    print(f"EXIT 2: cannot read index: {exc}")
    sys.exit(2)

indexed = set(TOKEN.findall(body))
missing = sorted(discovered - indexed)
rows = [ln for ln in body.splitlines() if ln.startswith("| APPROVED-")]
row_tokens = set(TOKEN.findall("\n".join(rows)))

fail = []
if not discovered:
    fail.append("no tokens discovered from repo docs")
if missing:
    fail.append("uncovered tokens: " + ", ".join(missing))
if row_tokens != discovered:
    fail.append("row/token mismatch: missing=" + str(sorted(discovered - row_tokens))
                + " extra=" + str(sorted(row_tokens - discovered)))
if len(rows) != len(discovered):
    fail.append(f"row count {len(rows)} != discovered {len(discovered)}")

bad_cells = []
for ln in rows:
    cells = [c.strip() for c in ln.strip().strip("|").split("|")]
    if any(not c for c in cells):
        bad_cells.append(ln[:80] + " (empty cell)")
    if len(cells) >= 5:
        rehearsal = cells[4]
        if not (re.search(r"\d{4}-\d{2}-\d{2}", rehearsal)
                or re.search(r"NOT REHEARSED|ABSENT|not recorded|N/A", rehearsal)):
            bad_cells.append(ln[:80] + " (rehearsal not explicit)")
if bad_cells:
    fail.append("implicit/omitted fields:\n  " + "\n  ".join(bad_cells))

for required in ("## Gap register", "ABSENT", "NOT REHEARSED"):
    if required not in body:
        fail.append("missing gap marker: " + required)

print(f"discovered={len(discovered)} rows={len(rows)} indexed={len(indexed & discovered)} "
      f"missing={len(missing)} bad_cells={len(bad_cells)}")
if fail:
    for f in fail:
        print("FAIL:", f)
    sys.exit(1)
print("PASS: no live operation silently omitted; gaps explicit")
sys.exit(0)
```
<!-- # --- DOC TEST END --- -->

Recorded results (this host, 2026-11-24):

- GREEN, committed index, exit 0:
  `discovered=28 rows=28 indexed=28 missing=0 bad_cells=0` /
  `PASS: no live operation silently omitted; gaps explicit`.
- RED, hostile copy with the `APPROVED-2026-11-24-stack-debug.md` row removed
  and the `APPROVED-2026-11-07-ssh-key.md` rollback cell blanked, exit 1:
  `discovered=28 rows=27 indexed=27 missing=1 bad_cells=1`, followed by
  `FAIL: uncovered tokens: APPROVED-2026-11-24-stack-debug.md`,
  `FAIL: row count 27 != discovered 28`, `FAIL: … (empty cell)`.

## Gates (run once, after the final edit)

Host: this worktree (`operator-cockpit`). `make` is not installed here (recorded
in the earlier task reports), so the documented fallback was used with the same
contract — no gate skipped, widened or edited:

```
PASS frozen      OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid | OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents | docs/frozen-manifests.sha256: 159 manifests
PASS provenance  159 headers checked against 5 frozen eras; 0 valuation sidecars checked; writer covered by tests
PASS tests       go vet + go test in all modules
PASS exit        5 safe invocations; verdict contract ok
PASS hygiene     1222 tracked files clean; legacy allowlist: 18 backup archives and results/runs-frozen.tar.gz; 4 named tailnet addresses
PASS links       5 documents scanned; plan and spec links resolve; 12 named historical absent-path allowance(s)

6 passed, 0 failed, 0 not run
```

Exit code **0**. The gate suite was run exactly once for this slice. The
pre-existing `TestAuditHTTPSealsTelemetryAndSource` temp-audit flake recorded in
Task 3 (disk 99 % full) did **not** reproduce in this run; no gate was re-run to
chase it and no VERIFY row was promoted by this builder pass.

## Concerns / what was skipped

- The index reflects only what the repository's own documents record. 7 tokens
  have no quoted rollback line, 25 have no recorded rehearsal date and 3 are
  draft/proposal-only — all named in the Gap register rather than omitted.
  Nothing here proves those operations were rehearsed on the host.
- Rehearsal dates for the 2026-11-11 pair are described as "named in the
  discovery evidence set, no distinct date" rather than a fabricated timestamp.
- The doc test discovers tokens by filename pattern in `docs/**` and
  `.superpowers/**`. A future live operation recorded only outside those trees
  would not be discovered; that is the stated scope of "known from the repo's
  own documents".
- Action pinning is by major tag, not full SHA — recorded above, not changed.
- No live host contact, token authoring, rehearsal evidence, ledger promotion,
  gate/CI edit or dependency change was performed by this slice.
