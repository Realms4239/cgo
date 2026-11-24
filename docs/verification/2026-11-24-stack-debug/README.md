# 2026-11-24 — systematic stack debugging and E2E check

Builder pass using `systematic-debugging`: root-cause investigation first, no
symptom fixes. Host work under `APPROVED-2026-11-24-stack-debug.md` (R2,
reversible, evidence files in `C:\WEBSERVER\evidence\`). No gate, ledger,
corpus, frozen-file or schema edit; no VERIFY-row promotion.

## Phase 1 - evidence by layer

| Layer | Evidence | Result |
|---|---|---|
| Repo build | `go build ./...` | exit 0 |
| Gates | `sh .githooks/run-gates.sh` | 5 PASS (frozen, provenance, tests, exit, hygiene), 1 FAIL (links) |
| Links FAIL | `docs/CHECKPOINTS.md:2808` references `web/frontend/src/lib/types.gen.ts`; `git ls-files` count 0 | stale protected-ledger link to a generated file; protected file, human disposition only |
| Frontend typecheck | `@tailwindcss/vite` declared in devDependencies but absent from `node_modules` | stale local install, not a code defect |
| Guest state | testbedsrv running; `cgo-ebpf` (19,064,804 B) and `cgo-physval-20261123` (16,410,606 B) present; evidence `770f42c7…` | healthy |
| Audit action | detached 15 s audit with the current binary; evidence `c02283dd…` | `audit-1795516495-537832da`, real p50 840 / p95 963, small 1502.9, goodput 3.37 Mbps, A+ note, honest SLA-failure line |
| Seal | `kerneltruth seal` on that fresh audit | exit 2, exact stderr `kernel_telemetry absent ou null`; fail-closed as designed (untagged binary) |
| Bench cell | P2/cake/cubic/down, run `1795516553` | valid row, rtt_p95 119.0, eight gates ✓, evidence `858720be…` |

The audit first hit `transport:` (exit 255) only because `safe_write.py` caps a
single SSH command at 20 s; a 15 s audit plus setup exceeds it. Detaching the
audit inside the guest reproduced the ratified pattern and worked - the tool
timeout, not the application, was the cause.

## Phase 2 - up-direction gate comparison

Working example: down cell `1795509754` (ingress cake), goodput 3.7 Mbps,
`gate_status: valid`. Broken example: egress/up `1795510344`, goodput 5.2 Mbps,
`gate_status: invalid` despite RTT p95 120 ms.

`pkg/campagne/campagne.go:668-676` computes `g4Cap := prof.CapUp()` and
`g4Floor := g4Cap * 0.5` for up, but `g4Floor := CapDown * 0.01` for down.
P2 has no explicit `capacity_up_mbps`, so `CapUp()` falls back to 20 Mbps:
the up floor is 10 Mbps and the observed 5.2 Mbps fails G4. The down floor is
0.2 Mbps and 3.7 Mbps passes. The code comment states this is deliberate
(up-direction data traverses the shaper, so 50 % is expected).

## Phase 3 - hypotheses tested

- H1 "audit action is broken": rejected. It produced a complete audit with real numbers and honest SLA verdict; only kernel telemetry is absent, by build tag.
- H2 "seal is broken": rejected. It refused with the precise absence error, which is the designed fail-closed path.
- H3 "up cell was a bad measurement": rejected as a tool failure. The measurement completed; G4 correctly rejected an up goodput below its declared 50 % floor. This is a policy/threshold question, not a bug to patch without operator approval.
- H4 "npm typecheck failure is a source defect": rejected; declared devDependency is not installed locally. Install from the lockfile is the root-cause fix (in progress at writing time).

## Open rulings (no silent fixes)

1. The `links` gate points at a generated file referenced only from the protected `CHECKPOINTS.md`. Fixing requires a human decision on the protected ledger or the generated-file policy.
2. The up-direction G4 floor of 50 % of `CapUp` is a project rule. Lowering it to accept 5.2 Mbps would be gate tampering without explicit approval.
3. The histogram-enabled eBPF object needs clang on a separate build host; none exists here (no clang, no WSL distro) and the card forbids clang on the VM. No live p99.9 until then.
4. The real EMD over sealed sets is 0.2807 with a 0.10 ceiling (`discard`): the emulator does not yet match the field distribution within the declared tolerance. This is a physics finding, not a tool failure.

## Restoration

Topology was restored to egress after the E2E cell; all run data and evidence
were kept. No service, sudoers, guard or permit file was changed by this pass.
