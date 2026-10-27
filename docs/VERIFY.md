# CGO — claim ledger

One row per claim a reader could rely on. A number in a document cites a row
here, or is labelled `RECORDED` or a hypothesis. The row is the source of
truth; prose is not. Every row carries a date and an evidence path or command.

## Statuses

- **VERIFIED**: the named command ran on the named host and date, and its
  output is saved or quoted in the checkpoint.
- **RECORDED**: a prior document states it; it has not been re-run here. It is
  not "verified".
- **PENDING**: a check exists and has not run.
- **BLOCKED**: it cannot run until a named condition clears (host, kernel,
  artifact, approval).
- **REFUTED**: a check showed it false. The row stays, with the evidence.

## Claims

| Claim | Status | Evidence (command or path) | Date | Host |
|---|---|---|---|---|
| Frozen corpus: 159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells | VERIFIED | `python book/tools/audit-archive.py` → exit 0:<br>`OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid`<br>`OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents` | 2026-10-27 | DESKTOP-B6K9UCM (Windows 10, MINGW64) |
| Frozen corpus unchanged since the seal | PENDING | The `frozen` gate does not exist yet; Phase A ran the audit but did not compare against a seal | 2026-10-27 | — |
| Windows engine: `go test ./...`, `go vet ./...`, Linux cross-build | RECORDED | Plan index 2026-10-27; to be re-run in Phase B, not re-run in this session | 2026-10-27 | Windows |
| Linux live acceptance (G0–G7, campaign) | BLOCKED | Needs the stable Linux 6.8 bench; no live run recorded in this system | 2026-10-27 | — |
| eBPF retransmit / RTT / qdisc drops | BLOCKED | No committed `bpf_probe.o` and no Linux 6.8 run; Windows cannot build the object | 2026-10-27 | — |
| IFB ingress shaping | BLOCKED | Live acceptance pending; Windows only generates commands | 2026-10-27 | — |
| BBRv3 3×3 matrix | BLOCKED | No 7.1 kernel; the VM exposes only reno, cubic, and bbr | 2026-10-27 | — |
| `.dist` netem | BLOCKED | No field trace at or above the 30-sample floor; the two `SMOKE-TEST` traces are 19 and 25 rows | 2026-10-27 | — |
| Tableau 15 CUBIC cells | RECORDED | Limitation: P2, `direction=up`, BBRv1, 1-flow legacy bench; CUBIC cells are invalid on G4 and MUST NOT be cited as a valid comparison | 2026-10-27 | — |
| Thesis book numbers | RECORDED | Thesis text; the corpus audit passed on 2026-10-27, but the individual figures were not re-derived in this system | 2026-10-27 | — |
| Second copy of `book/` and the frozen corpus | PENDING | A dated off-tree copy and a restore drill are required before submission; neither is recorded | 2026-10-27 | — |

## How a row moves

- Only a named command on a named host, with the date and the output saved or
  quoted next to the row.
- `RECORDED` is not promoted by confidence, repetition, or a later document
  agreeing.
- `REFUTED` stays in the table with the evidence that refuted it.
- A `BLOCKED` row moves to `PENDING` only when the named condition clears, and
  to `VERIFIED` only after the command runs.
