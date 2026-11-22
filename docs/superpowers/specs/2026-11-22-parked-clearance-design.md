# Parked-clearance programme — design (2026-11-22)

Authority: operator ordered "extensively, best smartest, most bold" clearance of all six named parks. All work local and evidence-backed. The C:\END corpus, protected ledgers, gates, frozen data, thesis and tag/freeze are untouched. No host, bench, sadcnwp, SSH key or live kernel trace is used.

## 1. SSE epoch marker (closes the Important residual)

The current fix resyncs only when replay is empty/stale. A reconnect whose `Last-Event-ID` lands mid-ring on a restarted hub (IDs restart at 1) receives non-empty deltas from a different generation and keeps stale structural state. Add a monotonically generated `epoch` to the `Hub`, include it in `lastFull` and in every published frame, and have the SSE client store it. On `epoch` change, the client clears the rings/live state and applies the full snapshot before processing deltas. Prove with a Go wire-level test (mid-ring, non-empty replay, different epoch → full frame emitted) and a DOM oracle for client-side reset. No fabricated frame.

## 2. BBRv3 dependency

`api/go.mod` has eight sibling `replace` lines but none for `capabilities`, so it compiles against cached `v0.1.0` (`CCPresent("bbr3")`) instead of the local module with `BBRv3Token()`/`StubProvisioned` — a proven false negative on the pinned google/bbr host whose v3 token is `bbr`. Add `replace github.com/Realms4239/cgo/capabilities => ../capabilities`, then prove locally with the real seam test that a provisioned host reports `BBRv3: true` and that `api/schema` exposes `bbr3` in that state. The real-kernel-host verifier re-run stays a named human gate; local correctness is fully proven.

## 3. E2E DOM failures

`TrustStrip.tsx` does render `data-slot="connection"` and `.trust-word`, and `App.tsx` mounts it. The two `fullstack-realtime.spec.ts` failures were attributed to a stale dist on pristine HEAD. Rebuild SPA + Go binary in the correct order, re-run the spec, and classify: if green, the park was a stale-artifact false positive — record it; if red, fix the real boundary with a genuine oracle. Never weaken the selector assertion to make it pass.

## 4. BenchConfig load flake

The `BenchConfig.dom.test.tsx` suite uses `vi.resetModules()` + dynamic import and occasionally exceeds the 5 s default under load. Root-cause by timing evidence, then fix with condition-based determinism (pre-import modules once, or explicit awaited setup) — not by raising the timeout to hide a defect. Prove with repeated runs.

## 5. Replay counters and identifiers

`replay.ts` still uses `parseInt(d.drops) || 0`, `repetition`, `event_id` — the same absence→measured-zero collapse already fixed for metrics. Introduce `archivedInt(v)` returning `NaN` for absent/blank/malformed values, use it for those three fields, and prove with a hostile DOM/node oracle that absence is NaN and a real `0` stays `0`. Preserve `LiveFrame` numeric types (NaN is a valid number) — no cascade of nullable types.

## 6. Kernel truth (local maximum)

Produce a deterministic local oracle for the `--confront-out` contract using the already-sealed 226-event evidence shape, and write a verifier runbook that names the exact rebuilt-eBPF-object + real-kernel-trace steps and the human approvals required. No host contact, no reinterpreting `qdisc_drop_packets=0` as loss. The live trace step remains parked with a named unblock condition; everything locally engineerable is done.

## Slice order and rules

Six sequential SDD slices; each: system/root-cause investigation → failing oracle → minimal fix → full relevant verification → explicit-path commit → fresh reviewer. No parallel subagents. Three-state exits, honest parks, no gate weakening, no fabricated PASS.
