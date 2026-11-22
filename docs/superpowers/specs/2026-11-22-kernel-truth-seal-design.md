# Kernel-truth seal producer — design (2026-11-22)

Authority: operator approved the full confrontation runbook and the SDD slice for its missing producer, 2026-11-22, with "do not wait for my approval anymore" for the subsequent steps. All work local except the final guarded window, which keeps token + rehearsal artifacts under the persisted 2026-11-22 ruling. No corpus, ledger, gate, frozen data or permit changes.

## 1. Problem

`cmd/cgo/confront.go` consumes `cgo.kernel-truth-seal/1` but nothing in the repository produces it. The runbook's step 3 asks for a real eBPF audit plus `tc -s` before/after and a sealed kernel-truth JSON — but no tool converts those real inputs into the seal schema. Two guarded windows failed: stale guest binary, then an audit report used as the seal input (`schema absent`). Hand-writing the seal from the numbers in `docs/VERIFY.md` would be assembling evidence and is forbidden.

## 2. Goal

Add a deterministic command that emits `cgo.kernel-truth-seal/1` **only** from real, already-sealed inputs, fail-closed on any missing or ambiguous value. It is a producer of a verification artifact, never a source of measurement.

## 3. Inputs and mapping (all real)

- `--audit <path>`: a sealed audit report. Read `audit_id`, `kernel_telemetry.source` (must be `ebpf`), `.available` (must be true), `.object_sha256`, `.retrans_payload_bytes`, `.qdisc_drop_packets`, and sum `kernel_telemetry.flows[].retrans_events` into `retrans_events`. No flow rows → fail closed.
- `--tc-before <path>` and `--tc-after <path>`: plain `tc -s qdisc show dev <iface>` text captures. Parse the `dropped N` counter on the target qdisc line; require exactly one unambiguous match per file, else fail closed. `tc_drops_after >= tc_drops_before`, else fail closed.
- `--out <path>`: atomic write (temp + rename) of the seal with `schema = cgo.kernel-truth-seal/1`, `source = ebpf`, the copied real values, `tc_interface`, `tc_qdisc`, and a `note` that `qdisc_drop_packets=0` is the named kernel reason-filter blind spot when it is zero — never a claim of zero loss.

## 4. Command shape

`cgo kerneltruth seal --audit <path> --tc-before <path> --tc-after <path> --iface e-wan --qdisc cake --out <path>`. Register as a new top-level command in `cmd/cgo/main.go` beside `run`/`audit`; no new dependency. Exit 0 only on a written, validated seal; exit 2 with a named reason on any failure. No default input path, no embedded fixture, no fabricated counter.

## 5. Verification

Table-driven tests with hostile oracles: missing audit/source/telemetry/flows, non-ebpf source, duplicate/ambiguous `dropped` lines, after < before, null counters, and a positive case with exact copied values and the zero-drop note. Tests must fail if any validation is deleted. `go test ./cmd/cgo`, `go build ./...`, gates once; commit explicit paths.

## 6. Live continuation (after the slice)

Build a linux binary, deploy as a new guest file, run a fresh eBPF audit through `CGO_ALLOW_PRIVATE_TARGETS=1` with the real prebuilt object, capture real `tc -s` before/after into guest files, run `kerneltruth seal` on those files, then `cgo run --confront-out` on the produced seal, preserving sha256 sidecars. Any refusal or ambiguity stops and rolls back. No VERIFY-row promotion from this builder pass.

## 7. Out of scope

No kernel/hardware changes, no clang/gcc/llvm on the VM, no corpus/ledger/gate/permit edit, no fabricated measurement or seal, no host contact without a fresh token + rehearsal under the persisted ruling.
