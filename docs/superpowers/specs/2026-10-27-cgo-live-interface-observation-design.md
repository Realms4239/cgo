# CGO — live interface throughput observation (slice 10) — design

*Status: approved · 2026-10-27 · Branch: `rebuild/plan1-economics-audit`*
*Motivating case: the operator QoS case under discussion (pseudonym `sadcnwp`). This is an engineering capability, not a case-study pivot decision.*
*Target bench: CGO-DEV twin, Ubuntu 24.04 LTS, Linux 6.8.0-147, 1 vCPU.*

---

## 1. Problem

The live wall has a "Bulk goodput" chart and a 10 Hz SSE transport, but
neither shows interface throughput at 10 Hz.

Three code facts break the promise:

1. `pkg/campagne/campagne.go` derives `bulk_goodput_mbps` from `tc -s` qdisc
   byte deltas inside `publishLive`. That function runs once per probe round,
   and only updates the value when the round gap exceeds 0.2 s. The SSE hub
   then repeats the last snapshot at 10 Hz (`pkg/api/sse.go`), so the chart is
   a staircase of probe rounds decorated with a 10 Hz transport.
2. `StartWatch` (`pkg/campagne/watch.go`) never computes bulk goodput. Outside
   a campaign the goodput chart has no new bulk data at all.
3. The qdisc counter lives on the shaped interface, not on the NIC. It counts
   the policy plane and misses traffic that bypasses the qdisc. It cannot
   answer "what did the link carry?"

A rendering-only fix — interpolation, smoothing, replay, zero-fill — would
make stale data look continuous. That is forbidden by the umbrella spec's
honesty invariants (invariant 2: no fabricated or stale-as-live value).

The operator asked for the real interface rate: RX/TX at the observed end,
real timestamps, no stutter, no artificial latency. This slice adds that
measurement path and a chart that plots it without interpolation.

## 2. Scope

### 2.1 In scope

- a Linux `/proc/net/dev` sampler driven by the existing 10 Hz hub clock;
- an additive `iface_*` field set on the API/SSE live frame;
- `CGO_IFACE` / `CGO_IFACE_SOURCE` server configuration;
- the twin `serve` path passing `c-cli` / `twin` by default;
- a "Débit interface" chart with RX/TX, real timestamps, gap breaks and
  cumulative drop counters;
- unit tests, the index's Windows gates, a twin live acceptance, and a browser
  performance trace;
- a ready-to-send read-only production request (appendix; not sent here).

### 2.2 Out of scope

- any production write, control action, configuration change, packet capture,
  or root access;
- the SNMP/RouterOS production adapter itself (Phase 2, separate slice);
- eBPF, IFB, BBRv3, trace-driven `.dist`, or thesis chapters;
- `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, `data/runs/`, or any
  frozen corpus change;
- new dependencies, a new SSE stream, a new endpoint, or a database;
- replay/playout buffering, interpolation, extrapolation, or a reduced sample
  rate presented as 10 Hz;
- both-end correlation in the UI. One sampler observes one end; the interface
  name and source label say which end. The peer's qdisc is used only as an
  acceptance reference, not as a second UI series.

## 3. Measured basis on the twin

This is a **pre-slice probe**, not the implementation's acceptance. It proves
the counter source and cadence are viable on the target host. It does not prove
the slice code or the UI.

### 3.1 Setup

- CGO-DEV: Ubuntu 24.04, `6.8.0-147-generic`, 1 vCPU, 2.9 GiB RAM.
- Current HEAD `05715f3` built in `/home/lab/cgo-current` with `GOPROXY=off`.
  The old `/home/lab/cgo` copy was not used.
- `bash kit/run-dualns.sh up --bin /home/lab/cgo-current/cgo-linux`.
- Topology: `ns_client` (`c-cli`) → `ns_wan_emulator` (`w-edge`, netem 100 ms)
  → `ns_edge_router` (`e-srv`, CAKE 20 Mbit/s) → `ns_server`.
- A standalone 10 Hz `/proc/net/dev` probe inside `ns_client`, during a 25 s
  local `nc` upload to `10.30.0.2:5201`. Local bench traffic only; no external
  download.

### 3.2 Results

| Metric | Measured |
|---|---|
| samples | 260 |
| window `dt` | 100.4–101.8 ms |
| active samples (`tx > 1 Mbit/s`) | 200 |
| active TX average | 12.56 Mbit/s |
| active TX p50 / p95 / p99 / max | 11.71 / 20.24 / 21.10 / 21.67 Mbit/s |
| `c-cli` TX total | 31.619 MB |
| `e-srv` qdisc `Sent` total | 31.551 MB |
| difference | +0.2 % |
| qdisc drops | 6 |

The p95 near the 20 Mbit/s CAKE ceiling shows the sampler catches the shaped
rate; the p50/avg show the actual flow was bursty and averaged about
12.5 Mbit/s, not a flat 20. That is a property of this short single-flow test,
not a sampler defect.

The byte totals are the important check. `c-cli` counts bytes before the CAKE
drop; `e-srv` counts bytes after it. The +0.2 % is in the expected direction
and magnitude (6 drops plus framing), so the interface counter is accounting
for the same stream. It is not an equality claim: the counters sit at
different layers.

### 3.3 What this proves and does not prove

Proves on the twin:

- `/proc/net/dev` can be read at a stable 10 Hz with a measured window of
  roughly 100 ms;
- the counters track the shaped stream and the byte total is within 0.2 % of
  the qdisc total for the same run;
- the rate series can see the 20 Mbit/s ceiling and the burst peaks.

Does not prove:

- that the CGO sampler or the UI works (neither exists yet);
- that production exposes 10 Hz counters (SNMP/RouterOS cadence is unknown);
- that interface throughput equals application goodput (different planes);
- both-end attribution (one end only);
- any production measurement (no access yet).

## 4. Measurement planes and honesty

### 4.1 Three counters, three questions

| Counter | Plane | Cadence today | Question it answers |
|---|---|---|---|
| `bulk_goodput_mbps` | receiver / qdisc bytes | probe round (seconds) | what did the application receive? |
| `iface_rx_mbps` / `iface_tx_mbps` | observed end's NIC/veth counters | 10 Hz, measured window | what did the link carry? |
| `drops` | qdisc drops | probe round | what did the policy drop? |

They are displayed side by side, never merged, never substituted for one
another, and never compared as if equal. The twin's 0.2 % byte agreement
validates the counter plane; it does not make the metrics interchangeable.

### 4.2 Invariants

- **I1 — real samples only.** Every rate comes from two real counter reads
  with a real, monotonic window. No interpolation, no replay, no synthetic
  ramp.
- **I2 — a gap is a gap.** An invalid window sets `iface_valid=false` and a
  reason; the chart breaks the line. It never emits a valid zero or negative
  rate to cover a failure.
- **I3 — timestamps are the sample's, not the frame's.** The chart plots
  `iface_sample_ts`. A repeated or late SSE frame cannot move a point in time.
- **I4 — the window is measured, not assumed.** `iface_window_ms` is the real
  `dt`; the UI shows it. "10 Hz" is a claim only when the measured window is
  near 100 ms.
- **I5 — provenance is declared.** `iface_source` is `twin`, `prod`, or
  `unknown`. The software never guesses; `unknown` is shown as "source non
  déclarée".
- **I6 — no artificial latency.** There is no playout buffer, no smoothing
  delay, no extrapolation. The unavoidable lag is bounded and stated in §8.3.
- **I7 — no production writes.** The sampler reads a file; the production
  phase is read-only; no `tc`, no config, no root, no capture.

## 5. Decisions

| # | Decision | Rationale |
|---|---|---|
| D1 | **Source is `/proc/net/dev`.** No `tc -s`, netlink, eBPF, or SNMP in this slice. | Cumulative 64-bit byte/drop counters, unprivileged, stable on Linux 6.8, measured at 10 Hz on the twin. |
| D2 | **The sampler is driven by the existing 10 Hz hub clock.** The hub's next function calls `sampler.Read()`; `/api/state` calls `sampler.Latest()`; there is no second ticker. | The hub already ticks at 10 Hz. Driving the read from it gives exactly one fresh sample per frame, with no drift, duplication, or lost sample from two independent clocks. |
| D3 | **Nominal window 100 ms; `dt` is measured.** Valid while `0 < dt ≤ 500 ms`. No minimum byte floor. | Byte counters are exact, so a low byte count is a real low rate; a floor would manufacture zeros for control traffic. A >500 ms gap is a gap, not a 1 s average. |
| D4 | **Invalid windows are explicit gaps.** `iface_valid=false` plus `iface_gap` = `warmup`, `absent`, `reset`, `slow`, `read`, or `clock`. | Honesty invariant I2. The UI breaks the line and names the reason. |
| D5 | **One sampler, explicit interface name; no auto-detection.** | Auto-detect can pick `lo`/`docker0` and mislabel the measurement. Production needs an operator-chosen interface. |
| D6 | **Source is declared, never detected.** `CGO_IFACE_SOURCE` accepts `twin`, `prod`, `unknown`; default `unknown`; anything else logs and stays `unknown`. | A production number must carry its provenance; the twin must not be confused with production. |
| D7 | **The fields are additive at the API frame.** `campagne.Snapshot` and the CSV are untouched; the server composes the snapshot and the sample at the API boundary. | The sampler is independent of campaign state; no campaign code or CSV change; the live frame gains fields without a migration. |
| D8 | **No new SSE stream or event.** The existing 10 Hz hub carries the fields. | The transport already exists; a second stream would add lifecycle and backpressure code for no benefit. |
| D9 | **The client plots each unique `iface_sample_ts` once.** Repeated frames are ignored; `clearLive()` does not clear the interface ring. | The sample timestamp is authoritative; campaign start/stop must not erase continuous observation. |
| D10 | **The interface chart paints at 10 Hz in the existing rAF loop; the other live charts keep their 250 ms gate.** | No global performance regression; the 10 Hz requirement is specific to the interface chart. |
| D11 | **The interface series is raw:** `smooth:false`, no LTTB/sampling, `connectNulls:false`, `animation:false`. | Smoothing overshoots and hides gaps; the line must show the measured points. |
| D12 | **RX/TX only in the chart; drops are cumulative text.** No third series or dual axis. | Packets and Mbit/s do not share an axis; the first slice shows the real counter, not a derived drop rate. |
| D13 | **The existing chart is relabeled "Débit utile applicatif".** Its data and semantics are unchanged. | Two measurement planes must not be confused. |
| D14 | **Disabled safely.** Empty `CGO_IFACE`, non-Linux, or an interface absent from the server's namespace disables or gaps the sampler; the server runs and the UI says "non observée". | No failure on Windows/observe; no fake line. |
| D15 | **Production is Phase 2, read-only, after operator approval.** The adapter implements the same `Sample` seam; no production claim before access. | No writes, no root, no capture; honest provenance. |
| D16 | **Zero display latency is not claimed.** The lag budget is stated, not hidden. | No measurement path can be instantaneous; the design removes artificial delay and bounds the rest. |

Two details from the earlier sketch are deliberately dropped:

- **The 10 KB byte floor.** The counters are exact byte counts, so a 10 KB
  floor would turn a real low rate into a false zero. The validity floors are
  `dt > 0` and `dt ≤ 500 ms` only.
- **A separate invalid first sample.** The sampler reads a baseline on the
  first call and returns a `warmup` gap; the first *rate* is the next 100 ms
  window. That is a baseline, not a fabricated zero, and it is visible in the
  UI as "démarrage".

## 6. Data contract

### 6.1 Fields

The following fields are added to the live frame (`GET /api/state` and each
`/api/stream` frame) when the sampler is configured. They are absent when
`CGO_IFACE` is empty.

| Field | Type | Meaning |
|---|---|---|
| `iface_name` | string | interface observed in the server's network namespace |
| `iface_source` | string | `twin`, `prod`, or `unknown` (declared, not detected) |
| `iface_rx_mbps` | number | receive rate over the measured window |
| `iface_tx_mbps` | number | transmit rate over the measured window |
| `iface_rx_drops` | integer | cumulative RX drops since interface creation |
| `iface_tx_drops` | integer | cumulative TX drops since interface creation |
| `iface_window_ms` | number | real `dt` of the window, in milliseconds |
| `iface_sample_ts` | integer | wall-clock Unix ms at the counter read |
| `iface_valid` | boolean | whether the numeric fields are a measurement |
| `iface_gap` | string | reason when `iface_valid=false`; empty otherwise |

**The `iface_valid` flag gates every numeric field.** When it is false, the
rate and drop fields are zero values in JSON and are not measurements; the UI
must render a gap or "—", never the zeros. This keeps one wire shape while
making the honesty rule enforceable in one place.

`iface_window_ms` is reported on a gap when a `dt` exists (for example
`slow`); it is 0 when no window could be formed (`read`, `absent`).

### 6.2 Gap reasons

| Reason | Meaning | Client behaviour |
|---|---|---|
| `warmup` | baseline read; no rate yet | append `null`, show "démarrage" |
| `absent` | interface not found in `/proc/net/dev` | append `null`, show "interface absente" |
| `reset` | a byte counter decreased | append `null`, show "compteur réinitialisé"; next valid sample rebaselines |
| `slow` | `dt > 500 ms` | append `null`, show "fenêtre trop longue"; no averaged rate |
| `read` | file read or parse error | append `null`, show "lecture impossible" |
| `clock` | monotonic `dt ≤ 0` | append `null`, show "horloge"; rebaseline |

The UI maps the known reasons to French; an unknown reason is shown verbatim,
never hidden. A gap does not reset the wire's cumulative drops on the server;
the next valid sample reports the new counters.

### 6.3 Composition and transport

`cmd/cgo/server.go` composes the frame:

```go
type liveFrame struct {
    campagne.Snapshot
    iface.Sample
}
```

The embedded fields flatten in JSON. The existing `Snapshot` shape, the CSV
export, and the campaign `OnSnap` path are unchanged. The hub's delta encoding
continues to compress only its declared structural keys; the `iface_*` fields
travel every frame (about 150 bytes at 10 Hz, negligible).

No new endpoint, event name, or stream is added. The hub and the state
endpoint get different compositions:

- the hub's `NextSnap` calls `sampler.Read()` — one fresh sample per 10 Hz
  frame;
- `/api/state` calls `sampler.Latest()` — the last sample, without advancing
  the baseline.

`pkg/api.Deps` gains an optional `NextSnap func() any`. When it is nil, the
hub falls back to `GetSnap`, so every existing test and caller keeps its
behaviour. `api.New` passes `NextSnap` to `hub.Serve` and leaves `GetSnap` for
`/api/state`. This is what keeps the 10 Hz stream the only clock that reads
the counters; a monitoring poll of `/api/state` cannot shorten an SSE window.

## 7. Sampler

### 7.1 Read and parse

`pkg/iface` is a leaf package (standard library only). It exports:

- `Sample` — the ten fields above;
- `Counters` — RX/TX bytes and drops;
- `Reader` — `Read() Sample` and `Latest() Sample`, the seam the future
  SNMP/RouterOS adapter implements;
- `New(name, source string) *Sampler` — production constructor;
- `ReadCounters(path string) (map[string]Counters, error)` — parser.

`/proc/net/dev` has two header lines and one line per interface. Each data
line is split at the **last** `:` (Linux interface names may contain a colon),
then into 16 fields. The parser reads:

- RX bytes: field 0
- RX drops: field 3
- TX bytes: field 8
- TX drops: field 11

Counters are parsed with `strconv.ParseUint(..., 10, 64)`. A line with fewer
than 16 fields is a parse error, not a skip. The interface is looked up by the
exact name; a missing name is `absent`, not a zero sample.

The file is per-network-namespace. The twin `run-dualns.sh serve` already runs
the dashboard inside `ns_client`, so `c-cli` is visible. On a host where the
interface is in another namespace, the sampler reports `absent`; the operator
must run the server in the right namespace or use the future adapter.

### 7.2 Window and rate

`Sampler.Read()` runs on the hub tick; `Sampler.Latest()` is read by
`/api/state`. A mutex keeps the pair consistent, and `Read` is the only
baseline-advancing call, so an external state poll cannot shorten an SSE
window.

If the file read or parse fails, the sample is a `read` gap and the baseline
is cleared. If the name is not in the parsed map, it is an `absent` gap and
the baseline is cleared; when the interface returns, the first read is
`warmup`, then the next window is valid.

1. Read the counters and take `now` (wall clock + monotonic).
2. If there is no baseline, store it and return a `warmup` gap.
3. Compute `dt = now - prevT`.
4. If `dt <= 0`, return `clock` and rebaseline.
5. If `dt > 500 ms`, return `slow` and rebaseline.
6. If RX or TX bytes decreased, return `reset` and rebaseline.
7. Otherwise compute:

```
rx_mbps = (rx_bytes - prev_rx_bytes) * 8 / 1e6 / dt_seconds
tx_mbps = (tx_bytes - prev_tx_bytes) * 8 / 1e6 / dt_seconds
```

No clamp is applied. The existing goodput path clamps at 2500 Mbit/s because
its qdisc counter can glitch across a replacement; a NIC counter that did not
decrease is a real rate. An absurd value would indicate a platform bug and
should be visible, not silently clamped.

The measured window, not the nominal 100 ms, is used in the denominator and is
reported in `iface_window_ms`. The rate's timestamp is the read time; the
chart plots it at that time.

**Why no 10 KB floor:** see D3. A floor would make low-rate SSH/control
traffic look idle. The counters are byte-exact, so a small delta is a real
small rate.

**Why the hub drives the read:** the alternative — a second 10 Hz ticker —
gives two independent clocks. When they drift, a sample can be read by the
sampler and overwritten before the hub sees it, or the same sample can be
published twice. Driving the read from the hub's existing tick makes the
sample rate exactly the transport rate and removes a goroutine. `/api/state`
is not the 10 Hz path; it returns the last sample and cannot advance the
baseline. The client dedupes by timestamp as a safety net.

### 7.3 Lifecycle and configuration

- `CGO_IFACE` — interface name. Empty disables the sampler. No auto-detect.
- `CGO_IFACE_SOURCE` — `twin`, `prod`, or `unknown`; default `unknown`.
  Unrecognized values log a warning and stay `unknown`.
- On non-Linux, a set `CGO_IFACE` logs "ignoré — `/proc/net/dev` absent" and
  the server runs normally. The package itself still compiles and its parser
  is tested on Windows with a temp file.
- The sampler has no goroutine and no shutdown path; it lives with the server
  and is read by the hub until the hub is closed.

`kit/run-dualns.sh serve` sets `CGO_IFACE=c-cli` and `CGO_IFACE_SOURCE=twin`
by default (overridable by the environment) inside `ns_client`. The production
systemd unit sets the operator's interface and `prod` explicitly. The
`cgo --serve` usage text gains one line naming both variables.

### 7.4 Why not the alternatives

| Alternative | Why not now |
|---|---|
| `tc -s` | qdisc plane, probe cadence, replacement resets; already the source of the misleading chart |
| netlink / ethtool stats | more code and driver-specific fields; `/proc/net/dev` is enough for byte/drop counters and was measured |
| eBPF | frozen scope, needs a prebuilt object, gives per-flow data this slice does not need |
| SNMP/RouterOS | production only; no access yet; Phase 2 adapter behind the same `Reader` seam |
| pcap | capture overhead and privacy; not needed for counters |

## 8. UI

### 8.1 Chart

A new full-width card, `Débit interface`, sits below the hero chart and above
the metric bento. It reuses `live-hero-card` (300 px) and adds no CSS.

- Head: `Débit interface`.
- Sub: `compteurs noyau · échantillon 10 Hz · fenêtre 60 s` (the 10 Hz and
  60 s wording is shown only while the measured window is near 100 ms; a
  slower source shows its real cadence).
- Right: `RX {x} · TX {y} Mbit/s · fenêtre {ms} ms · pertes IF RX {rx} ·
  TX {tx}` and the source badge (`twin`, `production`, or `source non
  déclarée`).
- Chart title: `Interface`.
- Two series: `RX` (`CRAFT.live`) and `TX` (`CRAFT.ok`). Existing palette
  tokens; no new color.
- Y axis: Mbit/s. The chart auto-scales; no clamp.
- X axis: `iface_sample_ts`, so the point sits at its real measurement time.
- Empty state: if no interface is configured, `interface non observée —
  déclarer CGO_IFACE côté serveur`; if configured but absent, `interface
  absente`. Never a zero line.

The `Charge` mark area used by the other live charts is carried on the
interface chart too, so the campaign phase is visible in the same place.

### 8.2 Ring and dedupe

`web/frontend/src/lib/live.ts` gains a separate interface ring:

```ts
type IfacePt = [number, number | null]
live.iface = {
  rx: [] as IfacePt[], tx: [] as IfacePt[],
  name: '', source: '', valid: false, gap: '',
  windowMs: 0, rxDrops: 0, txDrops: 0,
  ts: 0, seq: 0, max: 600,
}
```

`pushIface(f)`:

1. If `iface_name` changed, clear the ring and reset the last timestamp.
2. If `iface_sample_ts` is missing, equal to the last timestamp, or older by
   1 s or less, ignore it (duplicate or out-of-order frame).
3. If it is older by more than 1 s (wall-clock step), clear the ring and
   restart.
4. If `iface_valid`, append `[ts, rx]` and `[ts, tx]`; otherwise append
   `[ts, null]` to both (the line breaks).
5. Update the header fields (source, window, drops, gap) and increment
   `iface.seq` only when the timestamp advanced.

`live.max` stays 1800 for the campaign rings; the interface ring is capped at
600 samples (60 s at 10 Hz). A slower production source simply holds more
wall-clock time in the same 600 points; the header reports the real window.

`clearLive()` (called when a campaign starts) deliberately does **not** clear
`live.iface`: continuous observation survives campaign transitions.

`web/frontend/src/lib/sse.ts` calls `pushIface(data)` for every frame that
carries `iface_sample_ts`, **before** the campaign `measured` gate and
independent of `running`/`phase`. This is what makes the chart alive during
idle observation, when the old goodput chart has no new data. The replay pause
still suppresses live frames, including interface frames.

### 8.3 Render cadence and latency

The interface chart is painted in the existing `useRafLoop` with its own
100 ms gate and its own `iface.seq` dirty check. The existing charts keep the
250 ms gate:

```ts
if (live.iface.seq !== ifaceSeqRef.current && ts - ifaceLastRef.current >= 100) {
  iface.setData(rawSeries)
  ifaceSeqRef.current = live.iface.seq
  ifaceLastRef.current = ts
}
if (ts - lastRef.current < 250) return
// existing rtt/small/goodput paint
```

The interface series is not passed through the `d()` LTTB downsampler: 600
points at 10 Hz is small enough, and downsampling would drop real samples.
ECharts already uses `renderer:'canvas'`, `useDirtyRect:true`,
`animation:false`.

The lag budget:

| Stage | Bound |
|---|---|
| measurement window | 100 ms nominal, measured |
| transport (SSE tick) | ≤ 100 ms |
| render gate | ≤ 100 ms |
| browser frame | ≤ 16.7 ms at 60 Hz |
| **total, worst case** | **≈ 300 ms; typical ≈ 200 ms** |

This is the truthful version of "no latency": no artificial delay is added,
and the remaining delay is bounded and visible. The chart point is placed at
its real timestamp, so a late frame does not shift the data.

Acceptance requires a browser trace on the browser host (not the 1-vCPU VM):
the interface chart must advance at least 8 times per second during a 25 s
upload, and no long task may exceed 50 ms. If the trace fails, stop and report;
do not reduce the observation rate or reintroduce smoothing to hide it. The
only acceptable fallback is a shorter visible window or fewer points, never a
lower sample rate.

### 8.4 Copy and labels

- The existing card head and chart title change from `Bulk goodput` to
  `Débit utile applicatif`; the metric term `bulk_goodput` and the data are
  unchanged. The metric card keeps its short `Débit utile` label.
- The new card is `Débit interface`, not `goodput`: it is a link counter, not
  application delivery.
- The drop counters are labeled `pertes IF`, cumulative, to distinguish them
  from the existing qdisc `Pertes` metric.
- French UI text; English identifiers and JSON keys.

## 9. Failure modes

| Failure | Server | UI |
|---|---|---|
| `CGO_IFACE` empty | no sampler; no `iface_*` fields | "interface non observée" |
| non-Linux with `CGO_IFACE` set | log and disable | "interface non observée" |
| interface absent at startup | `absent` gap each read | "interface absente"; line broken |
| interface disappears mid-run | `absent` gap; baseline cleared | line broken; resumes after `warmup` |
| counter reset (link down/up) | `reset` gap; rebaseline | line broken; new segment |
| read/parse error | `read` gap | "lecture impossible"; line broken |
| host suspended > 500 ms | `slow` gap; no averaged rate | line broken; no fake smooth bridge |
| wall clock steps backward | sample is valid; server `dt` is monotonic | client clears the ring and restarts |
| `/api/state` read between hub ticks | returns the last sample; no baseline move | not plotted (state only) |
| hub backpressure / slow client | existing SSE backpressure event | existing status; interface samples resume from the ring/full frame |

## 10. Testing and acceptance

### 10.1 Unit and Windows gates

`pkg/iface` tests:

- parse a fixture with `lo`, a normal interface, a name containing `:`, and
  large 64-bit counters;
- reject a short line; skip header lines;
- missing interface → `absent`;
- first read → `warmup`; second read over 100 ms → valid rate;
- rate math for known bytes and `dt`;
- counter decrease → `reset` and rebaseline;
- `dt > 500 ms` → `slow`, no averaged rate;
- read error → `read`;
- source/name pass-through; unknown source normalisation;
- concurrent `Read`/`Latest` calls under `-race`.

Frontend tests:

- `pushIface` dedupes an identical timestamp;
- a gap appends `null` and breaks the series;
- a name change clears the ring;
- a backward clock step clears the ring;
- `clearLive()` preserves `live.iface`;
- the interface frame is pushed while idle (`running=false`, no `phase`).

Windows gates are the ones in `docs/superpowers/plans/README.md`: `go test
./...`, `go vet ./...` (with `web/frontend/dist`), `GOOS=linux GOARCH=amd64 go
build ./...`, the frontend tests/build, and the archive audit if `book/` is
present. If a gate is not runnable in the checkout, record "not runnable
here"; do not claim it passed.

The frozen corpus is untouched. The guard row in `docs/VERIFY.md` (created by
slice 9 Phase A) is authoritative; this slice does not run or rewrite the
frozen corpus, and it does not restate a count from memory.

### 10.2 Twin live acceptance

1. Build the slice branch into `/home/lab/cgo-current` on CGO-DEV with
   `GOPROXY=off`; the old `/home/lab/cgo` is not a source.
2. `bash kit/run-dualns.sh up --bin /home/lab/cgo-current/cgo-linux`.
3. `bash kit/run-dualns.sh serve --bin /home/lab/cgo-current/cgo-linux`
   (inside `ns_client`, `CGO_IFACE=c-cli`, `CGO_IFACE_SOURCE=twin`).
4. From `ns_client`, run a 25 s local upload to `10.30.0.2:5201`; use only
   local bench traffic, no external download.
5. Capture the SSE stream and assert:
   - unique `iface_sample_ts` count is at least 8/s of the run;
   - `iface_window_ms` stays ≤ 150 ms for the active run (p95);
   - every frame has a timestamp; no sample is fabricated or repeated as new;
   - `iface_tx_mbps` is in the expected order of the shaped stream;
   - the `c-cli` TX total and the `e-srv` qdisc `Sent` total for the same run
     are within 1 % (the pre-slice probe measured 0.2 %).
6. Gap test: bring `c-cli` down for ~2 s in `ns_client` and back up (SSH is on
   `enp0s3`, not `c-cli`); assert `absent`/`reset` gaps, a broken line, and a
   clean resume. Do not run this test on a production interface.
7. Browser trace on the host browser: the interface chart advances at 10 Hz
   (≥8 paints/s), no long task > 50 ms, and the gap produces a visible break.
   Save the trace/screenshot with the task report.
8. Record the measured values in `docs/VERIFY.md` (after Phase A) and the
   plan's task report. Teardown: `bash kit/run-dualns.sh down`; no namespaces
   left behind.

No production claim is made by this acceptance. The twin numbers are twin
numbers.

## 11. Production Phase 2 (not this slice)

### 11.1 Read-only request

The request is defined now and sent only with operator approval. It asks for:

- SNMP v2c (community restricted to the OIDs below) or v3 authNoPriv, **or** a
  dedicated read-only RouterOS API user;
- the interface name(s) chosen by the operator (WAN and, if available, the
  LAN side);
- the IF-MIB counters: `ifName`, `ifHCInOctets`, `ifHCOutOctets`,
  `ifInDiscards`, `ifOutDiscards`, `ifHighSpeed`;
- a poll interval of 10 s if the device supports it; otherwise the fastest
  interval the operator permits. The UI will display the real window, not a
  fake 10 Hz.

No writes, no configuration changes, no reboots, no packet capture, no root,
no extra accounts beyond the read-only one. The data is pseudonymized in every
document; no hostname or IP is published without operator approval. The
production site is referred to by the pseudonym `sadcnwp` unless the operator
approves naming.

### 11.2 Adapter seam

The future adapter implements `iface.Reader` and returns the same `Sample`.
The server's `GetSnap`/`NextSnap` composition, the SSE fields, the frontend
ring, and the chart are unchanged. The adapter is a separate slice with its
own tests, credential handling, and read-only proof; it is not smuggled into
this one.

### 11.3 No claims before access

Until the operator grants access and the adapter runs, the production columns
remain "pending"; the twin chart is labeled `twin`; nothing in the UI or the
report implies production measurement. The 10 Hz twin cadence is not presented
as a production capability.

## 12. Rollout

1. **User approval of this spec.** No code before approval.
2. **Slice 9 Phase A first** (engineering system: ledgers, frozen guard,
   entry). Slice 10 depends on the guard/ledger, not on the other live slices.
3. **Write the implementation plan** (separate document) and get approval.
4. **Execute**, in logical groups:
   - `pkg/iface` + tests;
   - server composition + env config + twin script + API docs;
   - frontend ring/SSE/chart/relabel + tests;
   - twin live acceptance + browser trace.
5. **Commit** per logical group with explicit paths; never `git add -A`.
   Leave user artifacts and `epbfBBR.md` untracked. Do not merge to `main`
   unless the user decides.
6. **Production request** only after approval; Phase 2 adapter separate.

## 13. Open inputs

- **User:** approved 2026-10-27 — the spec, the `Débit interface` placement,
  and the French copy.
- **Operator:** choose the production interface(s), grant read-only SNMP/
  RouterOS access, state the permitted poll interval, and approve or deny
  naming.
- **User:** confirm the browser host for the performance trace (the host
  browser, not the 1-vCPU VM).

None of these blocks the twin sampler. The twin acceptance is the only
verification this slice will claim.

---

## Appendix A — read-only production request (draft, not sent)

> Objet : demande d'accès en lecture seule aux compteurs d'interface du site
> `sadcnwp`.
>
> Pour la mesure QoS du site, nous demandons un accès **en lecture seule** aux
> compteurs de l'interface indiquée par vos soins :
>
> - SNMP v2c avec une communauté restreinte aux OID ci-dessous, ou SNMP v3
>   authNoPriv, ou un compte API RouterOS dédié en lecture seule ;
> - `ifName` (`.1.3.6.1.2.1.31.1.1.1.1`), `ifHCInOctets`
>   (`.1.3.6.1.2.1.31.1.1.1.6`), `ifHCOutOctets`
>   (`.1.3.6.1.2.1.31.1.1.1.10`), `ifInDiscards`
>   (`.1.3.6.1.2.1.2.2.1.13`), `ifOutDiscards`
>   (`.1.3.6.1.2.1.2.2.1.19`), `ifHighSpeed`
>   (`.1.3.6.1.2.1.31.1.1.1.15`) ;
> - une collecte toutes les 10 s si l'équipement le permet ; sinon la
>   fréquence maximale qu'il accepte, affichée telle quelle dans l'interface.
>
> Nous ne demandons aucune écriture, aucun changement de configuration, aucun
> redémarrage, aucune capture de paquets et aucun accès root. Un compte dédié,
> limité à ces OID et journalisé, nous convient. Les données seront
> pseudonymisées dans les documents ; aucun nom d'hôte ni adresse IP ne sera
> publié sans votre accord.
