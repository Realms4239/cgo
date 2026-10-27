# CGO — checkpoints

Append-only, newest at the bottom. A session that ends without an entry loses
its detail to compaction; this file is how the next session resumes.

## Resume protocol

1. Read `AGENTS.md`, then the latest entry here, then `docs/VERIFY.md`, then
   the current row of `docs/superpowers/plans/README.md`, then the artifact the
   entry names.
2. Compare the entry's branch and HEAD with `git log --oneline -1`. If they
   differ, say so before acting.
3. The artifact wins over the entry's prose.

## Entry format

```markdown
## YYYY-MM-DD HH:MM UTC · <focus>

- Branch/HEAD:
- Changed:
- Proven: <command → result, host>
- Unproven/blocked:
- Next:
```

UTC comes from a command, never the taskbar. A checkpoint is written at session
end, at a settled decision, and at a handoff.

## 2026-10-27 20:42 UTC · slice 9 Phase A preflight

- Branch/HEAD: `rebuild/plan1-economics-audit` / `05715f3`
- Changed: nothing — read-only preflight.
- Proven: on DESKTOP-B6K9UCM (Windows 10, MINGW64), `python book/tools/audit-archive.py` → exit 0, both lines verbatim:<br>`OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid`<br>`OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents`<br>`python --version` → Python 3.11.15; `go version` → go1.26.5 windows/amd64. Prerequisites present: `book/data/runs`, `book/tools/audit-archive.py`, `web/frontend/dist`. Code size: 194 `*.go` files, 94 `*_test.go` files, excluding `book/` and `.worktrees/`.
- Unproven/blocked: no gate runner yet (`make verify` targets arrive in Phase B); the Windows Go suite was not re-run in this session; Linux live acceptance, eBPF, IFB, BBRv3, and `.dist` remain blocked as listed in `docs/VERIFY.md`.
- Next: write `AGENTS.md`, `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, update the plan index, commit Phase A with explicit paths, then start Phase B (gate runner and six gates).
