# V2 corpus — full dualns matrix (2026-11-16, sadcnwp)

**Status: builder pass, IN PROGRESS.** Run ID `run-1794853778` (guest
`/home/bench/bench-data/run-1794853778/`). 13 of 24 cells sealed and `valid` at
last read (P1 × {cubic,bbr} ×3, P2 × {cubic,bbr} ×3, P3/cubic rep1). The remaining
11 cells are being resumed detached into the same run ID with `--run-id`
(already-frozen cells are skipped). Nothing is claimed as complete until 24/24
are read back, the anchor is sealed and the quarantine sidecar is present.

## Why this is not the frozen corpus

`book/data/runs` (159 runs / 528 rows) is the **frozen evidence corpus** and is
never written. This run lives under the bench VM's `/home/bench/bench-data` and
will be copied into this directory only after sealing, with its own hash8
anchor and quarantine sidecar.

## The authority path (operator-approved, 2026-11-16)

The operator explicitly authorized the full matrix. The guarded tool refuses
the bare campaign command:

```
REFUSE  sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 ip netns exec ns_client /home/bench/cgo run …  (sudo outside the named set)
```

No permit was added to `safe_write.py`. The run goes through `C:\WEBSERVER`
with a same-day scoped token (`APPROVED-2026-11-16-v2-matrix.md`), a staged
0600 disposable guest key, and the operator's SSH path to the bench VM. Every
write file carries the rollback pointer to `labo-cgo-bench-create.sh`.

## Three real defects found on the live bench and fixed in the repo

1. **Stale puits blocks namespace recreation** — `kit/cgo-bench-dualns.sh down`
   deleted namespaces while the puits still held `ns_server`, producing
   `Cannot create namespace file "/run/netns/ns_client": File exists` when the
   campaign's preflight re-ran `up`. Commit `d8a9a38`: `down` kills the puits
   **before** deleting namespaces.
2. **Silent empty measurement** — `kit/run-dualns.sh up` only pinged the puits.
   Without `CGO_TESTBED_OBJ` it serves HTTP 500 on `/small`; the campaign then
   records a cell with `small 0.0 ms / goodput 0.0 Mb` and gates
   `G1BulkStarted`, `G2ProbesProducing`, `G4ThroughputCoherent` invalid. Commit
   `d8a9a38`: `up` verifies `/small` from `ns_client`.
3. **Detached launcher dropped sudo / stale run lock** — `ip netns exec` needs
   root (`Operation not permitted` without `sudo -n`), and a killed campaign
   leaves `.run-<id>.lock`; `rm` is banned by global constitution, so commit
   `82dbfb2` adds `kit/v2-matrix-run.sh`, which renames the lock to
   `.stale-<date>` (evidence kept), enforces a single writer and launches
   `sudo -n` + `setsid` + `nohup`.

## Replication on another bench/VM

```bash
# 1. build the binary and stage kit scripts (committed versions)
# 2. bring up topology + puits in egress mode:
sudo -n bash kit/run-dualns.sh up --ingress off --bin /home/bench/cgo
# 3. launch or resume detached (script handles stale lock + sudo):
sudo -n bash kit/v2-matrix-run.sh --run-id run-1794853778 \
  --profiles P1,P2,P3,P4 --cc cubic,bbr --flows 4 --reps 3
# 4. poll the guest CSV, never write a second campaign while the first runs
awk -F, 'NR>1 {print $3"/"$5"/rep"$7"="$23}' \
  /home/bench/bench-data/run-1794853778/aqm_eval.csv
```

## Measurement shape

`--profiles P1,P2,P3,P4 --qdiscs cake --cc cubic,bbr --direction down
--flows 4 --reps 3 --deadline 220`. P1/P2 are expected to satisfy the 220 ms
deadline; P3 (600 ms RTT) and P4 (40 ms RTT, orbital jitter) may legitimately
produce `invalid` cells — those belong in the quarantine sidecar, not in a
"fix" that turns them green.

## Honesty notes

- Builder pass only; verifier confrontation pending. No VERIFIED row from this
  document.
- The passive field trace (sadcnwp, until 2026-11-17T16:22:27Z) was live during
  these campaigns; bench-induced load is a known confounder for its samples and
  must be named by the verifier.
- No figure from this run enters the thesis before the seal + independent
  verification.
