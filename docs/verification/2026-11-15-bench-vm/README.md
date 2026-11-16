# Bench VM (Lane A) — bring-up log, 2026-11-15

Host sadcnwp (sadc). Image: noble-server-cloudimg-amd64 (build 20260926.20261017, Ubuntu 24.04.5 LTS),
sha256 04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94 — matches Ubuntu's release manifest;
byte-identical to the local D: copy. qcow2 overlay on that image (~0 disk).

Operator-run sequence (his ssh session, after the guarded tool refused the rollback whitelist — see .superpowers/sdd/2026-11-14-final-route/progress.md):

- xorriso built nocloud-seed.iso from host files nocloud-user-data/nocloud-meta-data
  → ISO contains /nocloud-user-data and /nocloud-meta-data (NOT user-data/meta-data)
  → xorriso -volid warning: "cidata" does not comply with ECMA 119
- qemu launch OK: pid file bench-qemu.pid, serial → bench-console.log, hostfwd 127.0.0.1:10022→22
- BOOT OK: `Linux version 6.8.0-147-generic (Ubuntu 6.8.0-147.147-generic 6.8.12)`
  Ubuntu 24.04.5 LTS, kernel 6.8 GA line ✓ (bench parity; rehearsal VM is 6.8.0-149 — SRU delta only)
- **cloud-init datasource = DataSourceNone** — the seed was NOT mounted; nothing configured:
  ci-info: no authorized SSH keys fingerprints found for user ubuntu
  Cloud-init v. 26.1-0ubuntu1~24.04.1 finished … Datasource DataSourceNone … cc_final_message.py WARNING: Used fallback datasource
  (`Ubuntu 24.04.5 LTS ubuntu ttyS0` login prompt; default ubuntu user, no keys, bench user NOT created)
- Root cause: NoCloud looks for exactly `user-data` / `meta-data` at the ISO root;
  xorriso maps host basenames as-is, so the ISO carried `nocloud-user-data` / `nocloud-meta-data`.
  Fix prescribed (operator runs): kill the VM, rebuild the seed from a directory
  `/home/sadc/labo-cgo/nocloud-seed/` containing `user-data` + `meta-data`, volid CIDATA,
  recreate overlay, relaunch. Then verify: Datasource NoCloud [cidata], cloud-init final message,
  `Ubuntu 24.04.5 LTS labo-cgo-bench ttyS0`.

Kernel verdict: noble GA 6.8 ✓ — acceptable for the bench (6.8 parity with rehearsal VM).
Guest unusable until the seed is rebuilt and relaunched.

## Fix round (2026-11-15T19:55Z) — operator rebuilt the seed
- Operator re-ran the bring-up with a directory-based seed: `nocloud-seed/` containing `user-data` + `meta-data`, volid CIDATA (no ECMA warning), `xorriso -rock` over the directory, fresh overlay, relaunched (pid file rewritten 19:55Z).
- Boot log: kernel banner again 6.8.0-147-generic; cloud-init now downloading the noble apt indexes (Get:… lines) → datasource NoCloud MOUNTED, our user-data is being applied (package_update + packages from the draft yaml).
- Expected final lines: Datasource NoCloud [seed=/dev/sr0] … instance-id labo-cgo-bench-2026-11-15-r5, cc_final_message "labo-cgo-bench boot complete", ttyS0 prompt `Ubuntu 24.04.5 LTS labo-cgo-bench ttyS0`.
- Verification pass: after cloud-init completes (Get: lines drain), re-check Datasource line + final message; SSH path = operator's session (bench@127.0.0.1:10022, disposable 2026-11-11 key).

## SUCCESS (2026-11-15T19:56Z) — bench VM up on sadcnwp
- cloud-init completed from the NoCloud seed: `bench` user created with the disposable 2026-11-11 ed25519 key
  (ci-info: fingerprint 4c:03:c7:f9:c3:7c:86:ba:47:67:bd:b1:44:f3:e3:b8 · comment « labo-cgo-bench disposable guest key 2026-11-11 » ✓)
- Hostname: `labo-cgo-bench`; ssh host keys at install (root@labo-cgo-bench); `serial-getty@ttyS0` enabled
- Packages from the user-data installed (openssh-server, iproute2, iputils-ping, tcpdump, ethtool, traceroute, curl)
- Final message: « labo-cgo-bench boot complete » · reached cloud-init.target
- Kernel: 6.8.0-147-generic (noble GA 6.8, build 20260926.20261017 image) — bench parity line recorded vs rehearsal VM 6.8.0-149 (SRU delta only)
- SSH path (operator only): `ssh -p 10022 bench@127.0.0.1` from the sadcnwp host, key 2026-11-11 (`sudo` NOPASSWD inside the guest for the bench)
- Evidence (R0, sha256 sidecars in C:\WEBSERVER\evidence\):
  20261115-195349-read-window.txt · 195601-read-window.txt · 195610-read-window.txt · 195646-read-window.txt · 195656-read-window.txt
- Not done / not claimed: no ssh login from any machine verified (operator's session required); bench binary NOT deployed into the guest; nothing about Phase 3.4/4.3 proven yet.

## DEPLOYED + VERIFIED (2026-11-15T20:2xZ) — bench binary in the guest
- Transport: scp workstation→host (direct, `sadcnwp:9922`, silent success 2m45s) → host→guest via forwarded agent (`-A`) to `bench@127.0.0.1:10022` over slirp hostfwd; host key verification for BOTH hops used keys read from the VM's own console log (process-substituted known_hosts / mktemp KH, no files on disk outside /tmp, removed by trap; dedicated KH file on the workstation kept).
  - Note: inner scp takes `-P` (capital), inner ssh takes `-p` (lowercase) — the 127.0.0.1:22 « connection refused » was exactly this parser pitfall, not network.
- sha256 /home/bench/cgo = 89999ffa45ada688493dbd1a5c2342f91c0585c89854a1419e92447d6e787956 — matches the local pinned build ✓
- uname -m = x86_64 · `cgo version` → meteolink (cgo) 1.3.5 — linux/amd64, mode full ✓
- Verified via the operator's waiver (2026-11-15, "Waive the rule for this one deployment"): no sudo, no other files on prod disk besides the staged copy below, agent killed after each hop.
- Staged copy on prod disk: /home/sadc/labo-cgo/cgo-linux-amd64 (16 MB) — operator's call to remove (not touched by the deploy).
- Not done: cgo doctor (read-only), first live smoke cell — both need the operator's explicit go (load events on a shared prod box).

## BENCH SMOKE CELL SEALED (2026-11-16T00:15Z) — first live measurement on the sadcnwp bench VM
- Preconditions laid down this window (all systematic-debugged root causes):
  1. cgo-netctl ré-entrant exec passed a literal `--` to iproute2 6.1.0 → « exec of "--" failed » → every ré-entrant wrapper call failed → campaign physics step died at applyPhysics (« netem w-edge: exit status 1 »). Fixed in `kit/cgo-netctl` (3 sites, one line each, syntax-checked).
  2. `run-dualns.sh down` used unprivileged `pkill`; the puits/dashboard are launched under sudo (root) → never killed → « orphelins survivants » every down. Fixed: `$SUDO pkill` for both, plus `CGO_TESTBED_OBJ=/var/lib/cgo-testbed/obj16.bin` passed into the puits NS env (puits was returning HTTP 500 on /small → all small-object probes were absent ⇒ 0.0ms ⇒ G2ProbesProducing fail).
  3. `/small` probe from the machine works (HTTP 200, 16384 B); before the fix: HTTP 500 with 12 B body.
  4. Cell invocation requires `CGO_ALLOW_PRIVATE_TARGETS=1` (SSRF guard from 2026-11-14 hardening refuses RFC1918 by default — after that land, every bench cell on private targets fails closed with absent probes). Docs/config.md:35 documents the opt-in.
- Evidence: run-1794776677 — `P2/cake/cubic/down/rep=1, flows=1, deadline=220ms`
  rtt_p50=101.0, rtt_p95=120.3, qdi=19.3, small_p95=254.7, deadline_ok=92.7%, bulk_goodput=2.0Mb (campaign bulk-sender anomaly — pre-existing, raw-socket control test earlier measured 2.6/26 Mbit), drops=0, retrans=0, cpu=4.4%, gate_status=**valid** (all G0–G7). No quarantine dir (nothing invalid).
- Cell loop invariant discovered: **a finished cell deletes its shaper root** (e-wan noqueue) → the NEXT cell's preflight fails until `down`+`up` rebuilds the topology (down+up is seconds; the engine itself tries to re-up and fails on existing namespaces — a fail-closed abort, not damage). Every cell ⇒ `down`+`up` first.
- Cosmetic defect (logged, not fixed): the console recap prints `small p95 0.0 / deadline 0%` while the sealed row and valuation.json say `254.7/92.7` — the recap render reads stale zeros; the artifacts are authoritative. Queued for a task with a verifier.
- Committed: kit/cgo-netctl (iproute2 fix) + kit/run-dualns.sh (sudo pkill + obj16 env) — see git log.

## KERNEL-TRUTH AUDIT SEALED (2026-11-16T06:0xZ) — `audit-1794798576` (grade A+)
- Binary: `/home/bench/cgo-ebpf` = local build `-tags ebpf_object` — sha256 **6522e45084756e10aca87d99b1e4b887fb86f23dae4227cffa5ecc32d652b578**
  (previous staged copy `513a4e62…` was pre-restoration; overwritten by this seal's build).
- Restoration commit `9501b79` (local, unpushed): `cmd/cgo/kernel_ebpf_linux.go` + `kernel_ebpf_other.go` committed, and the `attachKernelProbe(&d)` wiring restored in `cmd/cgo/main.go` (deleted at f5a173a for being in untracked WIP — cause of the audit running silent tcpinfo).
- Invocation: `sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 ip netns exec ns_client /home/bench/cgo-ebpf audit --link-type 5g --site bench-sadcnwp-2026-11-16e --duration 45 --direction down --flows 1` — from ns_client.
- Report: `data/audits/audit-1794798576/audit_report.{json,md}` — grade A+ , `replay_command` embedded in `invocation`, seals: telemetry `7d34bcc62f767626…`, lines `23db39fd96e52303…`.
- **kernel_telemetry is sealed:**
  source=ebpf · object_sha256=**8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f** (the CI/rehearsal-pinned CO-RE object) ·
  retrans_payload_bytes=351840 · flows_seen=4/4 · qdisc_drops=0
  flows: {2.0.30.10→1.0.10.10 retrans 1448 (1 ev) srtt_us=204254} · {reverse 0 retrans srtt_us=200547} ·
         {2.0.30.10→1.0.10.10 retrans 350392 (242 ev) srtt_us=218806} · {reverse 0 retrans srtt_us=201019}
  — kernel-exact fields (`tcpi_bytes_retrans` path via kfree_skb/fentry), not the ×1448 proxy.
- Note for the verifier: 242 retransmit events on a 0% emulated-loss path — worth confronting with `tc -s` output on the shaper before calling it a fact about the pipe; the instrument's honesty posture records it, interpretation belongs to the verifier.
- Also landed: `9501b79` (the wiring restoration) — local only, NOT pushed; the S1-era history is intact.

## MATRIX 4.3 (A-opt) SEALED (2026-11-16T03:32Z) — `run-1794798866` — six rows, all valid
- Invocation: `sudo -n env CGO_ALLOW_PRIVATE_TARGETS=1 ip netns exec ns_client /home/bench/cgo run --bench dualns --profiles P2 --reps 3 --qdiscs cake --cc cubic,bbr --direction down --target 10.30.0.2 --deadline 220 --flows 4 --data /home/bench/bench-data`
- Binary: /home/bench/cgo sha256 89999ffa45ada688493dbd1a5c2342f91c0585c89854a1419e92447d6e787956. Bench state at run start: down→up verified (e-wan had been removed by the prior cell — the down/up invariant held).
- Verbatim rows:
```
run-1794798866,1,P2,cake,cubic,down,1,99.8,119.0,19.2,91.3,59.3,205.8,95.5,2.8,0,0,0,0.00,4.5,98.0,118.1,valid
run-1794798866,2,P2,cake,cubic,down,2,100.0,121.0,21.0,91.2,89.8,219.8,94.5,6.2,0,0,0,0.00,4.1,103.0,124.0,valid
run-1794798866,3,P2,cake,cubic,down,3,99.7,120.0,20.3,91.3,56.2,244.2,90.7,3.3,0,0,0,0.00,7.7,101.0,120.8,valid
run-1794798866,4,P2,cake,bbr,down,1,100.0,121.0,21.0,91.2,66.7,257.0,91.7,16.5,0,0,0,0.00,8.9,100.0,123.0,valid
run-1794798866,5,P2,cake,bbr,down,2,101.0,122.0,21.0,91.2,71.1,246.1,88.9,17.1,0,0,0,0.00,6.4,101.0,124.4,valid
run-1794798866,6,P2,cake,bbr,down,3,101.0,121.0,20.0,91.2,64.0,279.0,91.7,17.2,0,0,0,0.00,5.8,99.1,122.4,valid
```
- Legit-checking per-row: RTT p95 flat 119–122 ms (emulated 100+100 ms path); small_p95 205.8–279.0 ms at 0% loss; P2/cake/CUBIC goodput 2.8–6.2 Mb/s vs BBR 16.5–17.2 Mb/s (the cubic bulk-sender anomaly and its absence for bbr — record as such, interpret only with the verifier); drops/retrans 0 everywhere; deadline_ok 88.9–95.5% at objective 220 ms; gates G0–G7 valid on all six; quarantine sidecar absent.
- Frozen corpus untouched; post-freeze register = 1 (sealed cell 2026-11-16) + 6 (this matrix). Local evidence for the thesis: sealed artifacts referenced by VERIFY.md (row added 2026-11-16).
