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
| Interface observation, twin SSE acceptance (slice 10) | VERIFIED | CGO-DEV, Linux 6.8.0-147-generic. 32 s capture during a 60 MiB upload into the sink: `326` frames, `325` unique `iface_sample_ts` (≈10.2/s); `iface_window_ms` p50 `99.9` / p95 `101.3` / max `112.8`; `iface_tx_mbps` p50 `10.8` / p95 `19.9` / max `23.4`; `iface_rx_mbps` p50 `0.3`; `326/326` `iface_valid:true`, no gap reason; `iface_name=c-cli`, `iface_source=twin`. Commands: `sudo ip netns exec ns_client sh -c 'timeout 32 curl -sN http://10.99.0.2:9090/api/stream > /tmp/iface-sse.txt'` and `sudo ip netns exec ns_client sh -c 'dd if=/dev/zero bs=1M count=60 \| nc 10.30.0.2 5201'`; reductions by `grep -o`/`sort`/`awk`, verbatim output in `docs/CHECKPOINTS.md`. Byte accounting: `c-cli` TX delta `65,804,658` B vs `e-srv` qdisc `Sent` delta `65,783,532` B → `21,126` B difference (`0.032 %`); different layers, not called equal. Gap test: 8 s rename capture `20 absent`, `1 warmup`, `65 valid`; after the route restore `46 valid` frames and a `302 ms` sink ping (the rename flushed the route, E3). | 2026-10-27 | CGO-DEV `100.72.251.32` |
| Interface chart paint rate ≥8/s on the available browser | REFUTED | DESKTOP-B6K9UCM, Playwright 1.62.1 / Edge 151, live wall via `ssh -L 19090:10.99.0.2:9090`, 25 s during a 60 MiB upload: headless `179` paints / 25.288 s = `7.1/s`, 5 long tasks > 50 ms (max 164 ms); headed offscreen with occlusion throttling disabled `181` / 25.202 s = `7.2/s`, 1 long task 65 ms; a second headed run `201` / 25.216 s = `8.0/s`, 0 long tasks. Raw rAF on the same host: `46.1 fps` idle, `52.5 fps` with the live wall. With the approved strict `ts - last >= 100` gate a 46–52 Hz surface paints on the 5th–6th frame (≈120 ms). Ring and chart data advanced `260–263` samples in 25 s (10.3/s), `iface_valid:true`, `0` dropped SSE frames: the data path is not the limit. No 60 Hz surface was available; 10 Hz rendering is **not** claimed here. | 2026-10-27 | DESKTOP-B6K9UCM (Windows 10, MINGW64) |
| Slice 10 Windows gates and race | VERIFIED | `go test ./...` exit 0; `go vet ./...` exit 0; `npm test` 150/150; `npm run build` ok; `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` exit 0. Full Linux `go build ./...` still fails only in `cmd/cgo-gui` (Fyne `go-gl/gl` build constraints), pre-existing and outside the slice. Race: standalone `pkg/iface` copy on CGO-DEV, `go test -race ./...` → `ok ifacerace 1.066s`. | 2026-10-27 | DESKTOP-B6K9UCM / CGO-DEV |

## How a row moves

- Only a named command on a named host, with the date and the output saved or
  quoted next to the row.
- `RECORDED` is not promoted by confidence, repetition, or a later document
  agreeing.
- `REFUTED` stays in the table with the evidence that refuted it.
- A `BLOCKED` row moves to `PENDING` only when the named condition clears, and
  to `VERIFIED` only after the command runs.
