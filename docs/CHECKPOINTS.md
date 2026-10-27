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

## 2026-10-27 21:30 UTC · slice 10 implementation and twin acceptance

- Branch/HEAD: `rebuild/plan1-economics-audit` / `f7d4639`.
- Changed: slice 10 built in ten commits — `9e6342d` parser, `bed31b9` sampler,
  `3511ed0` hub `NextSnap` seam, `c021aec` live frame, `5449f9c` twin env/API,
  `cf6523d` iface ring + SSE push, `b8877a6` raw chart, `1c8e73e` beam guard;
  then `d21c9ca` acceptance row in `docs/VERIFY.md` + E3/E4 in
  `docs/ERRORS.md`, and `f7d4639` plan index. `web/frontend/dist/index.html`
  is modified by the build and deliberately not committed.
- Proven on CGO-DEV (`100.72.251.32`, Linux 6.8.0-147-generic), 32 s SSE
  capture during a 60 MiB upload: 326 frames, 325 unique `iface_sample_ts`;
  window p50 99.9 / p95 101.3 / max 112.8 ms; TX p50 10.8 / p95 19.9 /
  max 23.4 Mbit/s; RX p50 0.3; 326/326 valid; `c-cli`/`twin`; no gap reason.
  Byte accounting: `c-cli` TX delta 65,804,658 B vs `e-srv` qdisc `Sent`
  delta 65,783,532 B → 21,126 B (0.032 %). Gap test: 20 `absent`, 1 `warmup`,
  65 valid, then 46 valid after the route restore, sink ping 302 ms.
  Race: `go test -race ./...` on a standalone `pkg/iface` copy → `ok 1.066s`.
  Windows: `go test ./...` exit 0, `go vet ./...` exit 0, `npm test`
  150/150, `npm run build` ok, targeted Linux cross-build exit 0.
- Unproven/blocked: the browser paint gate is **REFUTED on this host**, not
  passed. Playwright/Edge 25 s traces during the upload: headless 7.1 paints/s
  (5 long tasks, max 164 ms), headed offscreen 7.2/s, second headed run 8.0/s.
  Raw rAF here is 46.1–52.5 fps, so the approved strict 100 ms gate lands on
  the 5th–6th frame (≈120 ms). The ring and chart data still advanced 10.3
  samples/s with 0 dropped SSE frames. Slice 10 stays **In progress** until a
  ≥60 Hz surface is traced; no code change is pending. Also unproven:
  production numbers (await operator read-only access), full Linux
  `go build ./...` (fails only in pre-existing `cmd/cgo-gui`),
  `go run ./cmd/cgo verify` (untracked `camp-fin3` data), and slice 9 Phase B.
- Bench state: `run-dualns.sh down` left root-owned sink/dashboard processes
  (E4); stopped manually with bracket `pkill`. No namespaces, processes or
  listeners remain; the SSH tunnel is closed; the VM stays running.
- Next: decide whether to certify the paint gate on a 60 Hz browser or to
  relax the gate constant (a design change, not a silent edit); send the
  read-only production request; then slice 9 Phase B.
