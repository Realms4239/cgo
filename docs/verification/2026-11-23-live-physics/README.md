# Live physics validation - 2026-11-23 (builder pass)

Two guarded windows under `C:/WEBSERVER/APPROVED-2026-11-23-live-physics-validation.md`
and `APPROVED-2026-11-23-netctl-permit.md`. All host contact went through
`safe_write.py`; no direct SSH, no host service/config change, no frozen corpus,
no ledger or gate edit, no VERIFY-row promotion.

## Result: correlation and serialisation are live-proven; permit restored

| Cell | Run | Kernel-applied netem (read back with `tc -s`) | Frozen row | Last log values |
|---|---|---|---|---|
| baseline P2 | `1795429528` | `delay 50ms 15ms` | 1/1 | small p95 223.4 ms, rtt 122.0 ms, G[✓×8] |
| V-CORR | `1795431568` | `delay 50ms 15ms 25% loss 0.5%` | 1/1 | small p95 205.5 ms, rtt 122.6 ms, G[✓×8] |
| V-SERIAL | `1795431920` | `delay 50ms 15ms 25% loss 0.5% rate 40Mbit packetoverhead 38` | 1/1 | small p95 348.1 ms, rtt 119.9 ms, G[✓×8] |

Both hops carried the same vector; `UP_RC=0`, `CHECK_RC=0` before each launch.
The correlation readback (evidence `20261123-105940`, sha256
`53c222077c6f8cff3148798719b2b7bb6ea3ded1fa8200f3f1f254c172299f6a`) and the
serialisation readback (`20261123-110603`, sha256
`2a7127de00baf884b6008310fc9e33a733e7273e9104ed32d0addc85bc790e28`) are the
kernel-level proof that the new tokens were applied by the new binary.

## The blocker and its resolution

The first V-CORR attempt failed with `netem w-edge: exit status 2`. A direct tc
probe with the identical vector succeeded, proving the kernel and iproute2 6.1.0
accept `25%`. Root cause: `/usr/local/bin/cgo-netctl` allowed only
`netem delay <TIME> <TIME> [loss …] [distribution …]`. The operator explicitly
approved a **narrow** permit change; a fail-closed patch added only validated
delay-CORRELATION and `rate [PACKETOVERHEAD]` tokens to that one branch
(permit sha `f4217068…`). Both cells then ran. The original permit was restored
byte-exact, verified sha
`b28092d2638d96fc9dfa2c1583cdd2374a219c3fc06e922dec019673ed463448` and
`PERMIT_RESTORED` (evidence `20261123-110912`, sha256
`e416ca8a8770c13e3da2fd6a7d8f957312492eeb4fb2c2886826f70139f930bb`). Topology
restored (evidence `20261123-110941`, sha256
`69b4b976ab5fc6f88fdff9d543c26d711bb41baa9d56912f9eaf939b0d4a5bec`).

## What this does and does not prove

- Proves: the shipped code applies percent-encoded correlation and explicit
  per-byte serialisation on the real dual-namespace bench; each configuration
  produces a valid frozen row with intact measurement-integrity gates.
- Does not prove: external validity. Three single cells are not a distribution,
  the small-object figures are non-stationary (serialisation charge 160.2 ms vs
  recup 348.1 ms), and `gate_status` is a measurement-integrity verdict, never
  an SLA. No field trace was captured, so the emulated link is still not shown
  to match a real link.
- The new `truth_planes` object was executed by the new binary with an **archived**
  real audit (`audit-1794898684`), not a fresh per-cell audit; `run` does not emit
  an audit by itself. No fresh per-cell kernel-truth seal is claimed.
- No value was hand-written, no failure edited into a pass, no permit left open.

## Fresh kernel truth (second window, same day)

A fresh eBPF observation was required because `run` emits no audit and the
validation binary was built without the `ebpf_object` tag — its audit had no
`kernel_telemetry`, and the new seal correctly refused with
`kernel_telemetry absent ou null` (no value invented; evidence
`20261123-112509`, sha256 `802a89f6…14b07a918aa`). The fresh audit was therefore
produced by the guest's already-tagged binary `/home/bench/cgo-ebpf`
(sha `6522e45084756e10aca87d99b1e4b887fb86f23dae4227cffa5ecc32d652b578`), and
the NEW binary only sealed and confronted it.

| Artifact | Value |
|---|---|
| Fresh audit | `audit-1795433639`, report sha256 `ef9851995d11245a630e7dd74bded065da863408aede20a45407b1ad1ed84384`, `source=ebpf` |
| Seal | `cgo.kernel-truth-seal/1`, sha256 `db5192ea99a9d45b4ab3cc38721f59ffe7a6d55dd4fcd17d9a9f3a1a1098e020`, object `8612f578…603d17f`, 73 retrans events / 105704 B |
| tc accounting | e-wan cake, drops `1 → 75`, delta 74 (before sha `fef6173e…`, after sha `174e64e7…`) |
| Confrontation | `cgo.ebpf-vs-tc/1`, sha256 `11249bcc46cff18e7103b98ecd54d2cfd4b9f2d6d8f85af2f7ce8f04879ff380`: 73 eBPF retrans vs 74 tc drops, **`match: false`** |
| Explicit absence | `qdisc_drop_packets_absent: true` with the named 6.8-cake reason-filter blind-spot note — never a measured zero |
| Truth planes | tc = "accounting surface (kernel counters: drops, overlimits, backlog, delay)"; eBPF = "temporal surface (per-packet retrans, srtt, bytes at near-wire speed)" |

Window evidence `20261123-113419`, sha256
`df082c59832eab7332d191625800c18c972c567ad9a65f70de597b5f44322d34`.
`match: false` is reported as-is: a 73/74 near-agreement is not agreement, and
no threshold was widened to make it green. Topology restored (`20261123-113500`,
sha256 `7a8bcd96b8fbcd82bc93573f671e150c320911b5d5a4c69367c25f8ac1017a6e`)
and the cgo-netctl permit re-verified byte-exact `b28092d2…` with
`PERMIT_RESTORED`.
