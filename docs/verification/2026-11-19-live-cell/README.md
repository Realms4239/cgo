# 2026-11-19 — bounded live Linux cell (KVM guest on sadcnwp)

Operator approval: page answer "I approve the ad hoc ssh and proceed with the bench and calibration", following the earlier "Authorise a bounded smoke cell now" with window "now".

## Environment

- Host: sadcnwp, KVM guest `labo-cgo-bench` (4 vCPU, 8192 MB, user-mode slirp NAT, hostfwd `127.0.0.1:10022 -> guest:22`).
- Access: ad-hoc SSH via jump host with `~/.ssh/sadcnwp`, guest key `~/.ssh/labo-cgo-bench`. The guest host key was pinned from the R0 console-log fingerprint `SHA256:6ct2H6rKP1Hik8XA9hVVh4a25VBEobDLlENiC6YemPo` in a temporary known_hosts file; the user's `known_hosts` was not modified and strict checking stayed on.
- Binary: current HEAD cross-compiled `GOOS=linux CGO_ENABLED=0 go build -trimpath ./cmd/cgo`, uploaded as `/home/bench/cgo-20261119`, guest SHA-256 `fb799649eaf62652b6436d92c6cb3cc7d0e0cf4612d7f39de9578ca0922f0289`.
- Bench prepared with the updated guest scripts: `sudo -n bash kit/run-dualns.sh up --ingress off --bin /home/bench/cgo-20261119` → 50 ms per hop, 100 ms total RTT, CAKE 20 Mbit; `[dualns] ok — link physics and queue policy are isolated`. The topology JSON emitted the fail-closed contract: `rtt_mode=per_hop_half`, `ifb_active=false`.

## Result — `run-1795087328`

Command: `sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 ip netns exec ns_client /home/bench/cgo-20261119 run --bench dualns --run-id run-1795087328 --profiles P2 --reps 1 --qdiscs cake --cc cubic --direction down --target 10.30.0.2 --deadline 220 --flows 4 --data /home/bench/bench-data`; exit 0.

One cell, sealed under `run-1795087328/`: `aqm_eval.csv`, `manifest.json`, `valuation.json`. Both declared artifact hashes verified locally (OK). The row is `gate_status=valid`:

- RTT p50/p95 101.0 / 122.0 ms (base 101.0 / 118.0 ms), QDI 21.0 ms
- small p95 253.4 ms, deadline_ok 92.6 %, goodput 4.3 Mbit/s
- drops 16, retransmissions 0, wasted bytes 23,168, cost 2.40 Ar/h, CPU 8.9 %

The generic `kit/check-v2-run.sh` was not used as the gate: it asserts a P1–P4 matrix and cannot accept a single-cell run. Its failure on "profil P1 absent" is a scope mismatch, not a cell defect; the per-file SHA-256 check above is the applicable oracle.

## Teardown

`sudo -n bash kit/run-dualns.sh down` completed; the guest check then reported no CAKE on e-wan and `NO_CAMPAIGN`. The teardown script also printed a missing `/root/cgo/cgo-linux` warning; the shaping and campaign were nevertheless removed. No live process or shaper left behind.

## What this proves / does not prove

Proves: the current HEAD binary runs a bounded dualns cell on the real Linux 6.8 bench, the 50+50 ms RTT convention is physically instantiated, and the cell seals `valid` with an honest manifest.

Does not prove: BBRv3, IFB ingress, field-tuned `.dist` replay, CHR, thesis identity, or the v2.0.0 freeze. The passive field trace still fails profile calibration (see `../2026-11-19-field-trace/README.md`); this live cell is not a field trace.
