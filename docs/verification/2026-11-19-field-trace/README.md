# 2026-11-19 — 48 h passive field trace retrieval and `.dist` attempt

## Provenance

The passive eBPF observatory authorized 2026-11-15 ran to a clean stop at
2026-11-17T19:22:32Z: 1,752,062 rows, 0 excluded, object SHA-256
`8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`,
manifest SHA-256 `66bcd654d2b6c33c31b3d7e4bd5ce27c39739100ce6a938abd3fd3b6567c5619`,
flows CSV SHA-256 `2639f06b727a86e673a68658cf77d382504818eeabbb15f2159ad0f27c257beb`.

Retrieved on 2026-11-19 through `C:\WEBSERVER` only:

- R0 permission diagnosis `evidence/20261119-073217-read-window.txt`, sha256 `b304309e…`.
- R2 guarded copy `evidence/20261119-083355-write-window.txt`, sha256 `1a03ed27…` (exit 0).
- R0 export read `evidence/20261119-083625-read-window.txt`, sha256 `673cada6…` (exit 0).
- R0 high-RTT samples `evidence/20261119-084054-read-window.txt`, sha256 `76446bf4…`.
- R0 full-trace srtt max `evidence/20261119-084201-read-window.txt`, sha256 `3098480f…`.
- R0 P2-band attempt `evidence/20261119-084249-read-window.txt`, sha256 in tool output.

Tokens: `APPROVED-2026-11-19-field-trace-read.md` (R0),
`APPROVED-2026-11-19-field-trace-readback.md` (R2, reversible),
`APPROVED-2026-11-19-field-trace-export-read.md` (R0). R1 proposal:
`R1-proposal-field-trace-readback-2026-11-19.md`.

## Result: calibration BLOCKED — trace matches no defined profile

- Head/tail slices: 80 flow samples, `srtt_us` 52–92 µs (~0.05–0.09 ms) — host-local flows.
- High-RTT slice: 80 flow samples, `srtt_us` 226,489–962,642 µs (227–963 ms).
- No `1xxxxx` µs band (100–199 ms) was found in the sampled windows.
- `cgo calibrate --profile P2` on the low band: **refused**, p50 0.08 ms ≤ 100 ms floor, exit 1.
- `cgo calibrate --profile P3` on the high band: **refused**, p50 241.79 ms ≤ 600 ms floor, exit 1.

The engine correctly failed closed. The passive trace observes host-local traffic
plus high-latency flows; it is not a P1/P2/P3/P4 profile trace. No `.dist` was
produced and none is claimed. The derived `data/audits/*` inputs under this
folder are labelled experiments, not sealed evidence.

Named unblock condition: an authorized audit of the actual constrained link
whose baseline `rtt_ms` matches a defined profile (P2 ≈ 100 ms or P3 ≈ 600 ms),
then `cgo calibrate --audit <id> --profile <P>` with a sealed sidecar.

The live Linux bench proof was **not** run in this pass: it requires a separate
same-day R2 token with a named rollback and rehearsal for bench commands.
