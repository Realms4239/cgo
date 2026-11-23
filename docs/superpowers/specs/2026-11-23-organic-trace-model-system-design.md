# New organic trace and realistic model system — design (2026-11-23)

Authority: operator chat 2026-11-23 "do new organic trace and more model realistic
system and more things to do", followed by "together" (design A and B together)
and approval of the combined design with a bounded 4-hour capture (operator ruling 2026-11-23: same-day token expiry). SDD required: one
slice at a time, fresh implementer then fresh reviewer, sequential.

Research grounding: `last30days` engine v3.27.2 run this session (45.8 s; social
clusters empty/off-topic; 20 citable primary sources appended to the saved raw
file), plus the project's own field evidence:
`../verification/2026-11-19-field-trace/` (48 h passive trace, 1,752,062 rows),
`../verification/2026-11-19-field-trace-p3/` (field-derived P3 `.dist`), and
`../verification/2026-11-23-field-trace-replay/` (live A/B replay validation).

No frozen corpus, book, thesis, gate, permit, ledger or protected config edit.
No fabricated sample, distribution, hash, PASS state or field provenance.

## 1. Scope decomposition

- **Slice A — fresh organic capture (live R2).** Produce a new passive trace on
  the sadcnwp host and retrieve it with provenance.
- **Slice B — trace-to-model system (local code + live validation gate).** Turn
  sealed traces into per-profile `.dist` calibrations and decide, with an
  explicit verdict, whether the emulator matches them.

B consumes A's artifacts. Slice A runs first. "More things" remains out of scope
until the operator names a third sub-project.

## 2. Slice A — fresh organic capture

Host facts verified by R0 read on 2026-11-23: `/home/sadc/labo-cgo/field-trace/`
contains `cgo-observe-x` (owned sadc, executable), `bpf_probe.o`, `observe.log`,
and root-owned 0700 dirs `out/` and `out-readable/` from the prior capture.

1. Same-day task-scoped R2 token with named window, rollback line, rehearsal
   hash and second channel, authored under the persisted executor-card ruling.
2. New disposable output directory
   `/home/sadc/labo-cgo/field-trace/out2`. The existing `out/` and
   `out-readable/` are NEVER modified, read for deletion, or reused.
3. Start the existing tagged binary as root via `sudo -n` inside the exact
   staged verb: `--duration 10800 --interval 5
   --exclude-ports 3333,4444,14444,4555 --out <out2>` (duration is integer
   SECONDS; the CLI hard-clamps at 172800, so 10800 = 3 h). Capture PID to a file.
   Passive raw-tracepoint attach only; no `tc`, netns, service or config change.
4. Stop by exact PID after the named window. Rollback is kill + removal of
   `out2` only. Any guard refusal stops as BLOCKED with exact stderr.
5. Retrieve through `C:\WEBSERVER` R0/R2 passes into
   `docs/verification/2026-11-23-organic-trace/`, keeping evidence files and
   sha256 sidecars. Verify the embedded object sha `8612f578…`, the manifest
   sha, and the `flows.csv` sha before any calibration use.
6. Report the srtt bands actually present. If the 100-199 ms band is again
   empty, say so explicitly — no P2 field-derived `.dist` is synthesised.

## 3. Slice B — trace-to-model system

1. A new tool under `kit/` (Go, no new dependency version) that reads a sealed
   `flows.csv`, buckets `srtt_us` into profile floor bands: P1 20-99 ms,
   P2 100-199 ms, P3 >=600 ms, P4 40-59 ms. A band with fewer than 30 samples
   is an explicit absence, never an interpolated or copied sample.
2. For each present band, emit a per-profile `telemetry.csv` with `basis,
   source_ref` provenance and run the existing `cgo calibrate --profile <P>`
   to produce a sealed `<P>.dist` and sidecar. Sidecars must carry
   `basis: measured`, `source_ref` = the new trace id, sample count and the
   source telemetry sha256. Existing calibration code and schemas are reused,
   not rewritten; no schema version bump.
3. Live A/B gate: for each new `.dist`, run one replay cell and one scalar
   control on the bench with the same binary and topology. Compare rtt p50/p95
   against the band's measured p50/p95; reuse `metrics.FidelityVerdict` with
   a declared epsilon to yield `trust | bounded | discard`. A band with <30
   samples yields `bounded`, never `trust`. Divergence in small-object p95,
   drops or goodput is reported alongside the delay result, never hidden.
4. Outputs live only in `docs/verification/2026-11-23-organic-trace/` and its
   replay subdirectory. No corpus, gate, ledger or config file is edited.

## 4. Error handling and testing

- Capture: a guard refusal, missing object sha or invalid manifest stops Slice A;
  no partial trace is promoted to a calibration source.
- Calibration: fewer than 30 samples, mismatched telemetry sha or invalid
  profile floor fails closed before writing any `.dist`.
- Validation: absence of a control run, ambiguous value or a second writer
  stops the window as BLOCKED; no SLA verdict is claimed from `gate_status`.
- RED-first oracles for band filtering, sample-count refusal, provenance
  fields, verdict precedence and byte-identical defaults; fresh implementer
  and reviewer per slice; native gates once per slice.

## 5. Named parks (unchanged)

Live eBPF object rebuild (48- vs 40-byte globals, human-gated), custom `.dist`
shape generator (JitterShape remains metadata), and the operator-visible
fidelity pipeline beyond these per-profile reports. This programme does not
claim any of them.
