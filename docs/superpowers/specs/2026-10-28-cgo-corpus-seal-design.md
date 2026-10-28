# CGO Frozen Corpus Seal Anchor — Design Specification

*Date: 2026-10-28 · Status: approved by the human · Slice 11 · Branch:
`rebuild/plan1-economics-audit`*

This slice adds the missing root above the frozen corpus manifests. It is a
seal, not a signature: it detects any rewrite of a manifest, including one made
together with the file that manifest covers. It does not defend against an
attacker who can rewrite the anchor too.

---

## 0. Context

`book/tools/audit-archive.py` verifies every run's `manifest.json` against the
files it names: file hashes, the config JSON, and `config_sha256`. That chain
proves manifest → file. It has no root: a file rewritten together with its own
manifest still passes.

The independent verifier narrowed the slice 9 frozen row for exactly this
reason: the corpus "matches its 159 per-run SHA-256 manifests", but there is no
corpus-wide anchor, and `book/` is gitignored. A committed anchor outside
`book/` closes the chain: anchor → manifest → file.

## 1. Decision log

- **D1 — Hash the manifests, not the data files.** The audit already proves
  manifest → file. Hashing 159 small manifests instead of 1 019 files keeps the
  anchor readable and the check cheap. Together the two layers cover the corpus.
- **D2 — Anchor at `docs/frozen-manifests.sha256`,** tracked. `book/` stays
  untracked and unchanged. Format is `sha256sum` compatible so the same file
  can be checked by hand on any host.
- **D3 — Fail closed when `book/` is present.** The gate FAILs if the anchor is
  absent, malformed, missing a run, listing an extra run, or mismatched. It
  SKIPs when `book/` is absent (a fresh clone). The pre-existing gate also
  SKIPs on a partial `book/` (missing audit or `data/runs`); that path exits 2
  and is a recorded follow-up, not a green result.
- **D4 — The gate never writes the anchor.** Re-cutting is a deliberate act
  with a documented command, so a corrupted corpus cannot re-seal itself in a
  gate run.
- **D5 — No new gate.** `frozen` gains one check; the six-gate count is
  unchanged.
- **D6 — Ceiling moves 600 → 650.** Spec §4.12 allows raising it in a commit;
  the new total stays inside the measured + 10 % band (596 measured, 650
  rounded to 25). Nothing is deleted because no gate is added.

## 2. Anchor format

```
<sha256>  book/data/runs/<run>/manifest.json
```

- Lowercase hex, two spaces, LF endings, UTF-8, one final newline.
- Sorted by run id (byte order), one line per manifest, 159 lines.
- `sha256sum -c docs/frozen-manifests.sha256` works from the repo root on
  Linux; the gate parses the same file in Python.

## 3. Gate behaviour

| Condition | Result |
|---|---|
| `book/` absent | `SKIP` (unchanged) |
| `book/` present, anchor absent or malformed | `FAIL`, names the anchor |
| anchor lists a run that is gone | `FAIL`, names the run |
| a run exists that the anchor does not list | `FAIL`, names the run |
| a manifest's bytes differ from the anchor | `FAIL`, names the run |
| anchor matches, audit exits 0 with two `OK` lines | `PASS` |

The `frozen` detail line names the anchor and the manifest count.

## 4. Failure modes covered

- File changed, manifest not → the existing audit fails.
- File and its manifest changed together → the audit passes, the anchor fails.
- A run added or removed → the anchor fails.
- A manifest reindented with no semantic change → the anchor fails. That is a
  corpus change and must be a human decision, not a silent edit.

## 5. Acceptance

- `docs/frozen-manifests.sha256` has exactly 159 lines, every run once, and
  every hash matches the tree.
- `python tools/verify.py --only frozen` → `PASS`, detail names the anchor.
- Fault proof: a copied corpus with a manifest rewritten (audit still exits 0)
  is caught by the seal; the untouched copy passes.
- `python tools/verify_test.py` → green.
- `python tools/verify.py` → `6 passed, 0 failed, 0 not run`.
- `python tools/verify.py --selftest` → six `caught`.

## 6. Out of scope

- No change to `book/`, its audit, the CSV schemas, or any frozen run.
- No signing key, remote timestamp, transparency log, git LFS, or dependency.
- No new gate, registry file, or generator.
- No UI, thesis, eBPF, IFB, BBRv3, or `.dist` work.
