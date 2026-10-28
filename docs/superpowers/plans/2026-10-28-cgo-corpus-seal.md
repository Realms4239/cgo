# CGO Frozen Corpus Seal Anchor — Implementation Plan

*Date: 2026-10-28 · Status: approved by the human · Branch:
`rebuild/plan1-economics-audit` · Spec:
`docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md`*

---

## 0. Scope guard

This plan adds one tracked anchor and one check in the existing `frozen` gate.
It MUST NOT:

- modify `book/data/runs`, `book/data/stats.json`, `book/content/`, the audit
  script, or any frozen run;
- change `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, or a Go CSV header;
- add a seventh gate, a registry file, or a generator;
- touch eBPF, IFB, BBRv3, `.dist`, the UI, or the thesis;
- run a campaign, the live bench, Graylog, or the remote PC;
- commit `epbfBBR.md`, `backup/`, `data/`, `book/`, `testbed-bootstrap/`,
  `unp*/`, `*.docx`, `*.pdf`, `cgo.exe`, or any user artifact;
- use `git add -A`; merge to `main`; or push.

## 1. File map

| Path | Kind | Task |
|---|---|---|
| `docs/frozen-manifests.sha256` | new, anchor | 1 |
| `tools/verify.py` | edit, `frozen` gate and proof | 2–3 |
| `tools/verify_test.py` | edit, one unit test | 3 |
| `docs/superpowers/plans/README.md` | edit, slice index | 0 |
| `docs/VERIFY.md` | edit, one row | 4 |
| `docs/CHECKPOINTS.md` | edit, one entry | 4 |

## 2. Task 1 — Cut the anchor

```bash
python - <<'PY'
import hashlib, pathlib
runs = pathlib.Path("book/data/runs")
lines = [f"{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.as_posix()}"
         for p in sorted(runs.glob("*/manifest.json"))]
pathlib.Path("docs/frozen-manifests.sha256").write_bytes(
    ("\n".join(lines) + "\n").encode("utf-8"))
print(len(lines), "manifests")
PY
```

Expect `159 manifests`. Check: 159 lines, sorted, every path under
`book/data/runs/`, no duplicate run id.

## 3. Task 2 — Extend the `frozen` gate

Add `seal_findings(root=ROOT)` to `tools/verify.py`: parse the anchor, hash each
`book/data/runs/*/manifest.json`, and return one finding per missing, extra,
malformed, or mismatched entry. `gate_frozen` calls it after the audit's two
`OK` lines and FAILs on any finding; the PASS detail names the anchor.

The gate never writes the anchor (D4). `book/` absent still returns `SKIP`
before the anchor is read.

## 4. Task 3 — Fault proof and unit test

`prove_frozen` gains a seal control before the existing data-file mutation:

1. copy `book/` to a temp root;
2. write the temp root's own `docs/frozen-manifests.sha256`;
3. control: `seal_findings(tmp)` is empty and the copied audit exits 0;
4. append a newline to one copied manifest;
5. the audit still exits 0 (valid JSON) but `seal_findings(tmp)` names the run —
   the anchor catches what the audit cannot;
6. restore that manifest, then run the existing data-file mutation proof.

`tools/verify_test.py` gets one `test_seal` using a two-file temp fixture:
match → no findings; changed manifest → finding; absent anchor → finding.

## 5. Task 4 — Ceiling and ledgers

- `MAX_GATE_LINES` 600 → 650 (D6), with the reason in the commit message.
- `docs/VERIFY.md`: one `PENDING` row for the anchor, then the independent
  verifier's `VERIFIED` row with commands and host.
- `docs/CHECKPOINTS.md`: one entry with the command, date, host, result, and
  what stays unproven (nothing live; this is a static check).

## 6. Task 5 — Verify and close

```bash
python tools/verify_test.py
python tools/verify.py --only frozen
python tools/verify.py --selftest
python tools/verify.py
```

Expect: tests green; `frozen` PASS naming the anchor; six `caught`; six gates
`6 passed, 0 failed, 0 not run`. `make` is absent on this host, so the targets
run as `python tools/verify.py`; name that, do not waive it.

Then dispatch an independent verifier for the delta. The builder may not
certify its own row.

## Acceptance summary

- The anchor is committed, 159 lines, and matches the tree.
- A manifest rewritten together with its file is caught by `frozen` while the
  audit still passes.
- Six gates green; one fault proof added to `frozen`.
- Live bench, corpus, schemas, and product code untouched.
