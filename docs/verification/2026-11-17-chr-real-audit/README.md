# CHR validation from a REAL audit — 2026-11-17

**This closes the fixture-only limitation of the earlier CHR seal.**

Input: `audit-1794898684`, produced on the sadcnwp bench VM by the FIXED
`/home/bench/cgo-fixed` binary (kernel telemetry `source=ebpf`, object
`8612f578…`, 421 368 retrans bytes, flows 5/5, `remediation_state=available`).
Its RouterOS script was generated from measured values: 17.52 Mbit/s bottleneck,
16 648 496 bit/s shaped rate (95%), 200 ms idle RTT, interfaces ether1/ether2,
down direction (LAN queue tree — no WAN script, which is correct for download).

Command (host sadcnwp, separate named window
`C:\WEBSERVER\APPROVED-2026-11-17-chr-real-audit.md`, pinned image re-deposited
before boot as required):
```
/home/sadc/labo-cgo/chr/cgo-chr-v13 --report audit-1794898684-report.json \
  --image /home/sadc/labo-cgo/chr/chr-7.24.6.img --level linux-namespace \
  --qemu /usr/bin/qemu-system-x86_64 --timeout 300s \
  --out /home/sadc/labo-cgo/chr/proof-real-v13
```

Verdict: `import.accepted=true`, **11/11 checks pass**, zero failures,
schema `cgo.chr-proof/1`. Export SHA-256 sealed in the proof:
`91d67458f740ada2c9e48364ef2a791e3f812302f2f4b354ca6a922f0ed26672`.
Host-side SHA-256 of `proof.json`, `export.txt` and `transcript.txt` is in
`HOST-SHA256.txt`.

## First attempt and its fix (not hidden)
`cgo-chr-v12` gave 10/11: `queue type cgoingress-cake.cake-bandwidth` expected
`16648496`, found `16.6Mbps`. RouterOS 7.24.6 rounds `/export terse` to one
decimal in Mbps (0.29% display error); the configuration was correct. Fixed in
`audit/routeros_compare.go` with a **rate-only 1% tolerance** plus a regression
test using the exact observed values (`0788835`); rebuilt as `cgo-chr-v13`,
re-ran from a freshly re-deposited pinned image → 11/11.

## Reading rule and residuals
- 11 checks, not 23: the prior seal covered both WAN and LAN scripts (fixture,
  up+down). This proof covers the **LAN/ingress** script only, because the real
  audit was down-direction. Do not claim WAN remediation from this proof.
- `import.accepted=true` alone is not compliance; the 11/11 check result is the
  verdict.
- The local `proof.json`/`export.txt` are **transcriptions** through the safe
  text channel, which is not byte-exact for these files. The authoritative
  hashes are the proof's own `export_sha256` and the host-side `HOST-SHA256.txt`.
  No byte-exactness is claimed for the local copies.
- No campaign was running during this window (all l3*/l4* matrices complete);
  the field trace, bench VM and frozen corpus were untouched.
