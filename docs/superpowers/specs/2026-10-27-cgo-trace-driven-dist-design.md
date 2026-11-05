# CGO Trace-Driven Netem `.dist` — Design Specification

*Date: 2026-10-27 · Status: approved · Slice 2 of the `epbfBBR.md`
overhaul · Branch: `rebuild/plan1-economics-audit` · Target bench: Ubuntu
24.04 LTS, Linux 6.8, amd64.*

This slice turns a captured audit trace into a netem delay distribution.
It is engine-first: the calibrator, the sealed artifact, the netem contract,
and the fail-closed campaign wiring ship now. The first **citable** field
artifact (`yas_4g*.dist`) waits for a real field trace to be committed.
The two local smoke traces (`audit-1792926281`, `audit-1792926346`) are not
citable evidence.

---

## 0. Context

Today the campaign synthesises delay from `model.Profile`: one scalar
`DelayMs` and one `JitterMs` per profile, applied identically to both hops
of the `dualns` bench. That is a declared model, not a measured shape. It
cannot represent a link whose RTT is mostly stable with a thin tail, or one
whose queueing produces a long, asymmetric tail.

`epbfBBR.md` Directive III asks for a trace-driven netem `.dist` calibrator:
measure the link, build netem's native distribution table, and let the bench
replay the measured shape. This specification covers that subsystem only.

The frozen corpus predates the `dualns` bench and stores `DelayMs` as an RTT
(P1 20.5/20, P2 100.2/100, P3 608/600). The current `dualns` `applyPhysics`
puts `DelayMs` on both `w-edge` and `c-wan`, so its scalar RTT is roughly
twice the corpus value. This slice does **not** change the scalar path; it
introduces a separate, opt-in dist path that interprets the profile delay as
RTT and splits it across hops. The scalar divergence is flagged for the
BBRv3 matrix slice, not fixed here.

### 0.1 In scope

- A strict reader for `data/audits/<id>/telemetry.csv`.
- A calibrator that uses **baseline RTT samples only** and refuses short or
  contradictory traces.
- A netem `.dist` writer using the kernel's real format and scale.
- A sealed sidecar JSON carrying provenance, statistics, and the `.dist`
  SHA-256.
- `cgo calibrate --audit ID --profile ID`.
- An opt-in campaign mode `delay_model=dist` that loads by profile and fails
  closed before any qdisc mutation.
- A documented RTT convention: floor split across both hops, empirical table
  on the measured hop only.
- Tests and Windows/Linux-build gates.

### 0.2 Out of scope

- Committing any calibration built from the smoke traces.
- Changing `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, their headers,
  or the frozen corpus.
- Changing the scalar `applyPhysics` path or the profile registry.
- IFB/RouterOS ingress, BBRv3, eBPF, UI, or thesis work.
- Multi-hop or per-direction distribution fitting; `both` is refused in dist
  mode.
- Runtime clang, libnl, or any new dependency.

### 0.3 Success criteria

1. `cgo calibrate --audit <field-id> --profile <id>` produces
   `data/calibrations/<id>.dist` and `<id>.json`; the sidecar seals the
   `.dist` bytes and the source telemetry hash.
2. A trace with fewer than 30 baseline RTT samples, or whose median RTT is
   at or below the profile floor, is rejected with a stated reason.
3. Loaded-phase samples never enter the table.
4. `delay_model=dist` loads the calibration by profile ID, verifies the
   sealed `.dist`, and refuses a missing, stale, or tampered artifact
   **before** the first `tc` call of the event.
5. On `dualns` up, `w-edge` receives `delay floor/2 jitter scale_ms ...
   distribution <profile>` and `c-wan` receives `delay floor/2 0ms`; on down
   the measured hop is `c-wan`. Reconstructed RTT ≈ floor + residual.
6. Scalar campaigns and the frozen guard are unchanged
   (`159 runs / 208 valid lines`).
7. Windows tests/vet and `GOOS=linux GOARCH=amd64 go build ./...` stay green.
8. No claim of live-kernel acceptance until the Linux bench applies a
   generated `.dist` and a campaign completes with `delay_model=dist`.

---

## 1. Decision log (this slice)

| # | Fork | Decision |
|---|---|---|
| D1 | Slice order | **Trace-driven `.dist` next.** eBPF is specced and planned but blocked on a stable Linux 6.8 host; IFB, BBRv3, and UI remain later slices. |
| D2 | Input window | **Baseline only.** A loaded residual belongs to the AQM, not to the link model. Never mix phases. |
| D3 | Sample floor | **30 baseline RTT samples.** Shorter traces cannot support a tail table; the two smoke traces (19 and 25 rows) fall below it. |
| D4 | Delay convention | **Profile `DelayMs` is an RTT.** Dist mode applies `floor/2` per hop and the empirical table on the measured hop only, so RTT = floor + residual. |
| D5 | Measured hop | **`w-edge` for up, `c-wan` for down.** Dist mode requires exactly one `PhysicsIfaces` entry and refuses `both`. |
| D6 | Table format | **Native netem:** decimal `int16` entries, scale 8192, max 32767, 1024 quantile entries. |
| D7 | Scale | **`scale_ms = p99.9_residual · 8192 / 32767`.** Entry `i = round(q_i / scale_ms · 8192)`, clamped `[0, 32767]`. |
| D8 | Artifact | **`<profile>.dist` + `<profile>.json`** in `data/calibrations/`. The sidecar is the provenance; the `.dist` is the kernel input. |
| D9 | Loading | **Fail closed by profile ID.** Missing sidecar, hash mismatch, profile mismatch, or floor mismatch stops the event before qdisc mutation. No synthetic fallback. |
| D10 | Provenance | **Additive and sealed.** The sidecar records audit ID, telemetry/report hashes, site/link type when present, sample count, floor, scale, clip/below-floor fractions, and round-trip quantiles. |
| D11 | Bench | **`dualns` only.** The dist path needs `TC_LIB_DIR` inside `ip netns exec`; legacy keeps the scalar path. |
| D12 | Frozen data | **No CSV changes.** `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, and the frozen corpus stay byte-identical. |

---

## 2. Pinned netem facts

These are contract facts, verified against the tc and kernel sources; they
drive the format and the tests.

1. `tc qdisc ... netem delay <base> <jitter> distribution <name>` reads
   `<TC_LIB_DIR>/<name>.dist`. `TC_LIB_DIR` defaults to `/usr/lib/tc/`.
2. The parser skips `#` lines. Each remaining line is one table entry, read
   with `strtol` base 0. **Never zero-pad** (`010` would be read as octal).
   The table is capped at 16384 entries.
3. The kernel samples the table uniformly and computes
   `delay = mu + (sigma / 8192) · t`, where `mu` is the delay base and
   `sigma` is the jitter argument.
4. `distribution` is rejected by tc unless **both** `delay` and `jitter` are
   present. A zero or negative jitter is therefore invalid.
5. `tc qdisc show` prints the delay/jitter. It does **not** print the
   table, and the distribution name is a userspace lookup key, not kernel
   state, so `show` cannot echo it either. The table is proven by the
   sealed file and by the RTT response, not by `show`.

The file name is the profile ID because the campaign loads by profile:
`data/calibrations/P2.dist` and `data/calibrations/P2.json`. `dist/` is
gitignored in this repository; `data/calibrations/` is not, so a deliberate
field artifact can be committed later.

---

## 3. Calibration pipeline

### 3.1 Input selection

`audit.ReadTelemetry(baseDir, id)` reads `data/audits/<id>/telemetry.csv`
with the exact header produced by `WriteTelemetry`:

```
phase,offset_ms,rtt_ms,small_ms,bulk_bytes
```

It rejects unknown headers, invalid numbers, negative values, and rows with
no measurement at all. An absent probe stays absent; it never becomes a
0 ms sample.

The calibrator keeps rows where:

- `phase == "baseline"`, exactly;
- `RTTms > 0`.

It drops `load`, `load-down`, and `drain` rows, and it drops small-object and
bulk rows (they have no RTT). If fewer than 30 RTT samples remain, it fails
with `trace trop courte`.

### 3.2 Math

Let `floor = profile.DelayMs` (RTT, milliseconds) and
`residual_i = max(0, rtt_i - floor)`. Samples below the floor are clamped to
zero and counted; their percentage is recorded in the sidecar.

- `p99.9 = quantile(residual, 0.999)` (nearest-rank on the sorted residual
  vector).
- If the **median RTT** is at or below `floor`, the trace does not match the
  profile: fail with a message telling the operator to fix the profile or
  choose another one.
- If `p99.9 <= 0`, fail: a zero-variance residual cannot form a table.
- `scale_ms = p99.9 · 8192 / 32767`.
- For `i = 0 … 1023`, `q_i = quantile(residual, (i + 0.5) / 1024)` and
  `t_i = round(q_i / scale_ms · 8192)`, clamped to `[0, 32767]`.
- The sidecar records `below_floor_pct`, `clipped_pct`, the source RTT
  p50/p95/p99.9/max, and the reconstructed RTT p50/p95/p99.9 computed as
  `floor + scale_ms · t / 8192`.

1024 entries keep the top entry near p99.9 while staying well under the
kernel's 16384 cap. The table is non-decreasing by construction; a test
asserts it.

### 3.3 Artifacts

`.dist` is one decimal integer per line, no padding, no comments, no
trailing spaces. The sidecar is indented JSON:

```json
{
  "schema": 1,
  "profile": "P2",
  "basis": "measured",
  "hop_model": "rtt_floor_split",
  "source_audit": "audit-…",
  "source_telemetry_sha256": "…",
  "source_report_sha256": "…",
  "source_site": "…",
  "source_link_type": "…",
  "created": "2026-10-27T…",
  "profile_delay_ms": 100,
  "floor_ms": 100,
  "hop_delay_ms": 50,
  "scale_ms": 0.123,
  "table_size": 1024,
  "dist_file": "P2.dist",
  "dist_sha256": "…",
  "samples": 125,
  "below_floor_pct": 0,
  "clipped_pct": 0,
  "source_p50_rtt_ms": 0,
  "source_p95_rtt_ms": 0,
  "source_p999_rtt_ms": 0,
  "source_max_rtt_ms": 0,
  "roundtrip_p50_rtt_ms": 0,
  "roundtrip_p95_rtt_ms": 0,
  "roundtrip_p999_rtt_ms": 0
}
```

`basis` is `"measured"` because the distribution comes from measured RTTs.
`hop_model` names the one-hop application convention so the assumption
travels with the number.

`WriteCalibration` writes the `.dist` first, then the sidecar, each through
a `.tmp` + rename. An orphan `.dist` is refused by `LoadCalibration`; it is
never guessed.

### 3.4 Loading

`LoadCalibration(dir, profile)`:

1. validates the profile ID as a simple file name;
2. reads `<profile>.json`; a missing or unparsable sidecar is an error;
3. requires `schema == 1`, `profile == requested`, `basis == "measured"`,
   `hop_model == "rtt_floor_split"`, `floor_ms > 0`, `scale_ms > 0`,
   `table_size == 1024`;
4. requires `dist_file == "<profile>.dist"`;
5. reads the `.dist`, compares its SHA-256 with `dist_sha256`, and requires
   the line count to equal `table_size`; any mismatch is an error.

The campaign additionally requires `floor_ms == profile.DelayMs`. A profile
edit after calibration therefore invalidates the artifact instead of
silently replaying an old shape.

---

## 4. Campaign dist mode

`Deps` gains:

- `DelayModel string` — `"scalar"` (default) or `"dist"`.
- `CalibrationDir string` — directory holding `<profile>.dist/.json`.

`applyPhysics` branches:

- `scalar`: the existing path, byte-identical; no calibration is read.
- `dist`: `applyPhysicsDist` runs **before** any reset or `tc` call:
  - refuses `direction == "both"`;
  - requires `bench == "dualns"`;
  - loads and verifies the calibration;
  - requires `cal.FloorMs == prof.DelayMs`;
  - requires exactly one `PhysicsIfaces` entry;
  - resolves the measured hop: `CliIf` for up, `PhysicsIfaces[0]` for down;
  - then resets and applies: measured hop gets
    `delay hop_delay scale_ms loss ... distribution <profile>`; the other
    hop gets `delay hop_delay 0ms loss ...`.

`ProdDepsDualNS` reads `CGO_DELAY_MODEL` and `CGO_CALIBRATION_DIR`
(default `data/calibrations`). In dist mode it makes the directory absolute
and sets `NsRunner.Env = ["TC_LIB_DIR=<dir>"]`; `NsRunner` runs
`sudo env TC_LIB_DIR=… ip netns exec <ns> tc …`, so the variable survives
sudo and the namespace exec. The shaper hop is untouched. The legacy
`ProdDeps` carries the same env, so `CGO_DELAY_MODEL=dist` on the legacy
bench fails the dualns-only check before any qdisc mutation instead of
silently falling back to scalar.

The scalar path remains the default and is not re-calibrated by this slice.
The known scalar 2× RTT divergence is recorded here and deferred to the
BBRv3 matrix slice.

---

## 5. Honesty and provenance rules

1. A `.dist` is only as good as its trace. The sidecar names the audit, the
   telemetry hash, the report hash, site, and link type when the report
   exists.
2. Loaded-phase samples are excluded, because they describe the AQM under
   test, not the link.
3. The table is applied to one hop. The sidecar says so (`hop_model`); the
   README/methodology repeats it. The campaign does not claim per-hop
   queueing measurements.
4. A missing or stale calibration is an error, never a synthetic fallback.
5. Smoke traces are engine fixtures, not field evidence. No smoke-derived
   `.dist` is committed as a citable calibration.
6. The first citable `yas_4g*.dist` requires a committed field trace and a
   sidecar whose numbers a reviewer can re-derive.

---

## 6. Failure modes

| Condition | Result |
|---|---|
| `telemetry.csv` missing / bad header / bad number | `cgo calibrate` fails, no artifact |
| < 30 baseline RTT samples | fail: `trace trop courte` |
| median RTT ≤ profile floor | fail: profile/trace mismatch |
| p99.9 residual ≤ 0 | fail: no table possible |
| `.dist` missing or hash mismatch | campaign fails before `tc` |
| sidecar profile/floor mismatch | campaign fails before `tc` |
| `delay_model=dist` with `both` | fail before `tc` |
| `delay_model=dist` on legacy bench | fail before `tc` |
| `TC_LIB_DIR` not passed through sudo | `tc` reports the distribution file missing; Linux acceptance catches it |
| `tc qdisc show` omits the table | expected; the sealed file is the proof |

---

## 7. Gates

Windows / cross-build (must stay green):

```
go test ./...
go vet ./...
GOOS=linux GOARCH=amd64 go build ./...
```

Frozen guard (local corpus): `python book/tools/audit-archive.py` →
`159 runs`, `208 valid lines`. Do not rewrite the corpus to make a gate pass.

Linux acceptance (blocked until the bench is stable; not a Windows gate):

1. Build `cgo-linux` and calibrate a trace.
2. Apply the `.dist` manually with `TC_LIB_DIR` and confirm `tc` accepts it.
3. Run one `dualns` cell with `CGO_DELAY_MODEL=dist` and confirm the RTT
   track and the absence of a synthetic fallback.
4. Remove the sidecar and confirm the campaign fails before any qdisc
   mutation.

A live success is not claimed in the plan or the docs until step 2–4 run on
Linux 6.8.

---

## 8. Deferred

- The scalar `DelayMs` 2× divergence fix (BBRv3 matrix slice).
- Per-direction `.dist` fitting and multi-hop application.
- IFB/RouterOS ingress shaping.
- UI surfacing of the active calibration.
- The first citable field `.dist`.
