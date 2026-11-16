# Kit V2 — bench path for a replicated VM

Three committed scripts own the v2 dualns path. Run them in this order on the
bench VM; never edit `safe_write.py` or add wildcard sudo grants.

1. `run-dualns.sh up --ingress off --bin /home/bench/cgo`
   Builds netem 100 ms + CAKE 20 Mbit (egress) **and starts the puits**. It now
   checks `/small` from `ns_client`, so a missing pinned 16 KiB object fails
   closed instead of producing a "valid" cell with 0.0 ms / 0.0 Mb.
2. `v2-matrix-run.sh --run-id <id> --profiles P1,P2,P3,P4 --cc cubic,bbr --flows 4 --reps 3`
   Launches/resumes the campaign detached with `sudo -n`; renames a stale
   `.run-<id>.lock` to `.stale-<date>` (never deletes evidence); refuses a
   second writer. Already-frozen cells are skipped by the binary's `--run-id`.
3. `check-v2-run.sh <run-dir> 24` then `seal-v2-run.sh <run-dir>`
   Asserts 24 cells, a manifest, the gate vocabulary and all four profiles;
   then prints the hash8 and quarantine state. The frozen corpus
   `book/data/runs` is never touched.

Full authority path, root causes and honesty notes:
`docs/verification/2026-11-16-v2-corpus/README.md`.
