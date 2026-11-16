# V2 corpus — sealed (2026-11-16T19:56Z)

- Run: `run-1794853778` (guest `/home/bench/bench-data/run-1794853778/`).
- **24/24 cells `gate_status=valid`** — P1/P2/P3/P4 × cake × {cubic,bbr} × 3 reps,
  down, 4 flows, deadline 220 ms. No quarantine sidecar: no cell was invalid.
- Own anchor (separate from frozen corpus): `hash8 = 6adc83d9`
  (`aqm_eval.csv` SHA-256 `6adc83d9eba4295ff32d252dee2ebf59a529c7eb470b81be401ca58c78363f74`).
- Manifest SHA-256 `db777e73447bb38f1b17d78f8017464a22be269247fa8059` (verbatim from guest).
- Verified locally with the committed gates: `bash kit/check-v2-run.sh … 24` (OK) and
  `bash kit/seal-v2-run.sh …` (OK). `make`-equivalent gates 6/6 after the seal.
- Frozen corpus `book/data/runs` untouched: still 159 runs / 528 rows / 208 valid,
  `PASS frozen` at HEAD. This run lives only under this verification directory.
- Builder pass only — verifier confrontation pending. The passive field trace was live
  on sadcnwp until 2026-11-17T16:22:27Z; bench load is a named confounder for its samples.
- Provenance of fixes that made this run possible: `d8a9a38` (puits teardown + /small
  preflight), `82dbfb2` (detached launcher + stale lock rename), `e11d4cd` (run checker),
  `b5d892b` (seal boundary note), `a60778f` (replication guide).
