# Live physics validation - 2026-11-23 (builder pass, PARTIAL / BLOCKED)

Guarded window under `C:/WEBSERVER/APPROVED-2026-11-23-live-physics-validation.md`
(R2, reversible, 09:45-23:59 UTC). All SSH traffic went through `safe_write.py`;
no direct SSH, no host service/config change, no frozen corpus, no ledger or gate
edit, no VERIFY-row promotion.

## What actually ran

- New Linux binary `cgo-physval-20261123` (source `ad6731e`, 16 410 606 bytes,
  guest sha256 `0a56e018280c4c795b822cbd5cda22003bdc64f282744424dac7297dc724c894`)
  staged with launcher and two imported validation profiles (`V-CORR`, `V-SERIAL`),
  evidence `20261123-101959`, sha256 `db1075eb…46ebf`.
- **Baseline cell P2/cake/bbr/down, run `1795429528`** completed on the real bench:
  `UP_RC=0`, `CHECK_RC=0`, valid 3816-byte tc-before capture, 1/1 frozen row,
  log showing `small p95 223.4 ms`, `rtt 122.0 ms`, `G[✓✓✓✓✓✓✓✓] 1/1`.
  Evidence `20261123-102528`, sha256 `05f3ae4b…bf4`.
- The new confrontation schema was executed by the new binary with an **archived,
  real** audit (`audit-1794898684`, NOT this run): `cgo.ebpf-vs-tc/1` carried the
  additive `truth_planes` object (tc = accounting surface, eBPF = temporal
  surface, no hierarchy claim); it reported `retrans_events 291`,
  `tc_drops_delta 0`, `match false`. This is a schema-execution check only — it
  is NOT a measurement of the baseline cell and is not claimed as such.
- Every `run` in this campaign does not emit an audit by itself; the ratified
  recipe calls a separate `audit` command. No fresh per-cell audit was produced,
  so no fresh per-cell kernel-truth seal is claimed.

## BLOCKED: the two new physics knobs cannot pass the guest permit wrapper

- `V-CORR` failed at first tc mutation: `netem: netem w-edge: netem w-edge:
  exit status 2` (evidence `20261123-103344`, sha256
  `8cd8abc3babdbdf26ceeeba793667a78e3a8729c5e49c6851fd280ad6b978ec4`).
- A direct tc probe with the exact vector succeeded and read back
  `qdisc netem 1: root ... delay 50ms 15ms 25% loss 0.5%` (evidence
  `20261123-103328`, sha256 `20f5f783…2471e`). So the kernel and iproute2 6.1.0
  accept the percent-encoded token.
- Root cause: `/usr/local/bin/cgo-netctl` allows only
  `netem delay <TIME> <TIME> [loss <PCT> | loss gemodel ...] [distribution <NOM>]`.
  It has no correlation token and no rate token. `V-SERIAL` was therefore not
  attempted — same blocker. Neither the permit nor any refusal rule was edited.
- A guard refusal is a finding, not a reason to reword the request. Widening
  `cgo-netctl` needs a fresh, explicit operator decision and a separately
  rehearsed R2 change; this builder pass does not do it.

## Disposition

Diagnostic qdisc removed and topology restored (`20261123-103434`, sha256
`50ef0794…23eea1c2d`); all run data and evidence kept. Correlation and
serialisation remain **locally proven, live-BLOCKED** with the named unblock
condition above. No number was hand-written and no failure was edited into a pass.
