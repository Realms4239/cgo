# RouterOS replay validation + fail-closed control — 2026-11-17

No-QEMU validation (the live Layer-1 campaign was the single writer, so no CHR
boot was started). Commands run from the repository root with the committed
`cmd/cgo-chr` tool; both consume the real RouterOS 7.24.6 export sealed on
2026-11-16 (`export-2026-11-16m.txt`, 982 bytes, sha `61e51573…`).

## Positive control
```
go run ./cmd/cgo-chr \
  --report docs/verification/2026-11-16-chr-boot/artifacts/chr-fixture-cake-2026-11-16-r2.json \
  --export-file docs/verification/2026-11-16-chr-boot/artifacts/export-2026-11-16m.txt \
  --out <temp>/chr-replay-local
```
Exit **0** · `import.accepted=true` · **23/23** checks pass ·
`export_sha256=61e5157394e458cf…` · schema `cgo.chr-proof/1`.
Reproduces the sealed `proof-2026-11-16m.json` locally from committed evidence.

## Negative control (fail-closed)
One byte-level edit to a COPY of the export:
`name=cgo-wan-out parent=ether1` → `name=cgo-wan-out parent=ether9`.
```
go run ./cmd/cgo-chr \
  --report <same fixture report> \
  --export-file docs/verification/2026-11-17-chr-replay/export-mutated.txt \
  --out <temp>/chr-replay-neg
```
Exit **1** · **22/23** checks · the single failing check is exactly
`queue tree cgo-wan-out.parent`, found `ether9`. The tool still writes a
`proof.json` with `import.accepted=true`.

## Reading rule (important)
**`import.accepted=true` is not a compliance verdict.** RouterOS accepted the
import syntax in both runs; only the check comparison separates them. A proof
is compliant only when exit code is 0 AND every check passes (23/23 here).
This is why row 71 of `docs/VERIFY.md` states both facts, not acceptance alone.

## Scope / residuals
- The report consumed is still the **fixture** (`fixture: true`), not the real
  bench audit. The sealed audit `audit-1794798576` has no RouterOS script
  (`remediation_state=observation_only`, reason: "no interface was named"), so
  it cannot drive a CHR import. Closing that needs a fresh audit with
  `-iface` / `-routeros-iface` / `-ros-lan` in a single-writer gap, then a real
  CHR boot under a separate named window. Not done here.
- No CHR boot, no host config change, no new token; this used the no-QEMU
  replay mode only, so the live campaign could not be contaminated.
- Artifacts: `proof-positive.json`, `proof-negative.json`, `export-mutated.txt`,
  `SHA256SUMS` (sidecar of all three).
