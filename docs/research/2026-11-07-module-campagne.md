# pkg/campagne as standalone tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/campagne` only. Working dir `operator-cockpit`. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — The Flow Queue CoDel Packet Scheduler and AQM Algorithm (Experimental) | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | FQ-CoDel/CoDel normative defaults: interval 100 ms (§5.2.1), target 5 ms (§5.2.2), packet limit MUST (§5.2.3), quantum 1514 (§5.2.4), flows 1024 (§5.2.5); flow queueing vs CoDel (§1–§4), implementation §5 |
| R2 | Prometheus — Data model + Exposition formats | https://prometheus.io/docs/concepts/data_model/ ; https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched (data_model direct; exposition_formats via data_model + search index comment §4.2) | Metric naming (`[a-zA-Z_:][a-zA-Z0-9_:]*`), labels (`[a-zA-Z_][a-zA-Z0-9_]*`, `__` reserved), sample = float64 + ms timestamp, metric name = `__name__` label pair |
| R3 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | fetched | module path = ID + import prefix, `/vN` suffix for v2+, `go` directive = mandatory minimum (since Go 1.21), `require`/`replace`/`exclude`/`retract` semantics |
| R4 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | MAJOR/MINOR/PATCH rules, MUST declare public API (§1), immutability once released (§3), 0.y.z MAY change (§4) |
| — | tc(8) shaping/ingress semantics, procfs formats | — | unverified — primary source not found (not fetched in this pass; only in-repo code `pkg/campagne/*.go` and `docs/architecture.md:43-58`) | shaper placement, `/proc/stat` CPU, `/proc/sys/net/ipv4/tcp_available_congestion_control` |
| — | CAKE scheduler spec / tc-cake(8) | — | unverified — primary source not found (RFC 8290 cites CAKE only as informative ref `[CAKE]` §10.2 and "currently being developed" §7) | `cake` cell semantics |
| — | Gilbert-Elliott / netem `gemodel` | — | unverified — primary source not found (only in-repo `pkg/campagne/dualns.go:18-26` constants) | `BurstGE` parameters |

Method: read all non-test `pkg/campagne/*.go` (20 files: `campagne.go`, `config.go`, `plane.go`, `matrix.go`, `baseline.go`, `capabilities.go`, `ccavail*.go`, `clock.go`, `cpu_*.go`, `dualns.go`, `instrument.go`, `prod.go`, `valuation.go`, `watch.go`, `writer.go`, `modprobe_*.go`), `docs/architecture.md:13`, `docs/api.md:59-73` run endpoints, `go.mod:1-3`. Every technical claim below cites `file:line` or `(R#)`; anything else is labelled **unverified — primary source not found**.

## 1. Responsibility

`docs/architecture.md:13` summarises: "`pkg/campagne` | Orchestration de la matrice : `Matrix`, `StartMatrixWithID`, `RunEvent` (baseline → charge → récupération), `Deps` (TC, shaper, sondes, `StatsFn`, `DeadlineMs`)".

`pkg/campagne/campagne.go:1-3` states the contract: "`orchestration d'une cellule de matrice (baseline → charge → récupération)`". Concretely the package does seven things, none of which it explicitly disclaims:

- **Single-cell orchestration** — `RunEvent` (`campagne.go:302-703`) applies physics (`applyPhysics` in `dualns.go:46-62`), resets and applies the shaper (`campagne.go:336-371` via `qdisc.ApplyShaper` / `ApplyIngressShaper`), samples baseline/charge/recovery RTT+small-object (`campagne.go:439-466` `collect`), floods bulk with the cell's CC (`campagne.go:499-544`), samples CPU during charge (`campagne.go:550-582`), evaluates G0–G7 gates (`campagne.go:373-374`, `472-481`, `587-657`, `678-701`), and freezes a `model.Event` line.
- **Matrix orchestration** — `startMatrix` (`matrix.go:103-258`) iterates `profiles × qdiscs × ccs × reps` (`matrix.go:178-255`), skips already-frozen cells via `Writer.seen` (`matrix.go:203-207`), supports per-cell skip without stopping the matrix (`matrix.go:43-51`, `213-250`), and defers `w.Freeze(cfgJSON)` on exit (`matrix.go:176`).
- **Live diffusion surface** — `Live`/`Snapshot`/`NewLive` (`campagne.go:162-216`) + `Deps.Snapshot` (`campagne.go:104-140`) compose the SSE frame; `StartWatch` (`watch.go:23-100`) is the non-intrusive ping+small loop (phase `surveil`, `Running=false`) outside campaigns.
- **Persistence + provenance** — `Writer` (`writer.go:17-200`) appends to `aqm_eval.csv` with G5 dedupe (`writer.go:132-152`) and freezes `manifest.json` + `quarantine.json` with SHA-256 (`writer.go:159-200`); `buildMatrixConfig` (`config.go:71-121`) freezes resolved profiles + params; `writeValuationEntry` (`valuation.go:50-95`) appends `valuation.json` sidecar.
- **Host capability + preflight** — `CapabilitiesNow`/`Capabilities` (`capabilities.go:8-34`), `DefaultCCs` (`ccavail.go:27-36`), `PlaneReady` (`plane.go:19-32`), `BaselineStable` G6 gate (`baseline.go:52-72`), `BurstGE` Gilbert-Elliott selector (`dualns.go:18-39`).
- **Production wiring** — `ProdDeps`/`ProdDepsDown` (`prod.go:24-87`) and `ProdDepsDualNS` (`dualns.go:112-181`) bind `Deps` to veth/netns topology, env vars (`CGO_*`), and direction-aware `StatsFn`.
- **Policy helpers** — G6 jitter-aware threshold (`baseline.go:45`, `65-72`), G4 direction-aware floor (`campagne.go:649-655`), CPU gate nil-on-unmeasured (`campagne.go:146-155`, `656`), cost-at-frozen-tier (`campagne.go:707-712`).

What it deliberately does **not** do: it applies no shaping itself (delegates to `qdisc`), runs no probe itself (delegates to `probe.Ping/SmallObject/BulkSendTo` via `Deps` function fields, `campagne.go:29-31`), computes no percentile itself (delegates to `metrics.Summarize/DeadlineOKPct/JFI/VoIPR`, `campagne.go:135-138`, `404-411`, `539`, `626`), and knows no HTTP/SSE transport (pushes via `OnSnap func(Snapshot)` callback, `campagne.go:97`; `docs/api.md:63-65` maps `POST /api/run/start|stop|skip` to `StartMatrix*`/`Matrix.Stop`/`Matrix.Skip`).

API surface (`docs/api.md:59-73` run endpoints): `POST /api/run/start` accepts `{profiles, qdiscs?, ccs?, reps, deadline_ms, target, direction?, flows?, bench?, channel_model?}` and starts a matrix (`docs/api.md:63`); `POST /api/run/stop`, `POST /api/run/skip`, `GET /api/results?run=`, `GET /api/results/delta`, `GET /api/run/rows?run=`, `GET /api/events` are the read-side (`docs/api.md:64-69`). `cc=bbr3` is gated on the kernel exposing exact token `bbr3` (`docs/api.md:71-73`, implemented in `ccavail_linux.go:12-18` + `matrix.go:138-140`).

## 2. Exported API (file:line)

Complete inventory of exported identifiers in non-test sources. Unexported helpers are listed after each file and are not API.

### 2.1 `campagne.go` — cell + live frame

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Deps` | struct | `campagne.go:20-98` | 30+ fields: hops `TC`, `TCShaper qdiscRunner`, `CliIf`, `ShaperIf`; addresses `Target`, `SmallURL`, `BulkAddr`; probes `Ping func(ctx,target,n)[]float64`, `Small func(ctx)(float64,error)`, `Bulk func(ctx,addr)(uint64,error)`, `BulkN func(ctx,addr,cc,n)([]uint64,error)`; `CPU func()float64`, `Now func()time.Time`, `StatsFn func()[]qdisc.Stats`; policy `DeadlineMs`, `ProfileDelayMs`, `Flows`, `PhysicsIfaces`, `ChannelModel`, `Bench`, `Tier`, `DelayModel`, `CalibrationDir`, `Direction`, `ShapeMode`, `IngressPreflight func()error`; windows `BaselineSec`, `ChargeSec`, `RecupSec`; progress `TotalEvents`, `DoneEvents`; `OnSnap func(Snapshot)`. Nil funcs default in `defaults` (`campagne.go:218-273`). |
| `func (d *Deps) Snapshot(phase,load string, ev model.Event, base,chg map[string]float64, bulk uint64, prof model.Profile, gates []*bool) Snapshot` | method | `campagne.go:104-140` | `base`/`chg`/`bulk` unused (dead params); computes `CostPrimaryARPerH` (GOLD) + `CostB2BARPerH` (APN) + `JFIPct` (`campagne.go:134-138`). |
| `Live` | struct | `campagne.go:162-165` | `{mu sync.Mutex; snap Snapshot}` — mutable snapshot behind mutex. |
| `Snapshot` | struct | `campagne.go:167-198` | 22 JSON-tagged fields: costs (`cost_primary_ar_per_h`, `cost_b2b_ar_per_h`, `jfi_pct`), structure (`phase`, `profile`, `qdisc`, `cc`, `repetition`, `event_id`, `total_events`, `done_events`, `phase_total_s`, `load_status`), measurements (`rtt_p50_ms`, `rtt_p95_ms`, `small_p95_ms`, `bulk_goodput_mbps`, `drops`, `wasted_bytes`, `cost_ar_per_h`, `deadline_ok_pct`, `profile_capacity_mbps`, `profile_delay_ms`, `profile_jitter_ms`, `profile_loss_pct`), `gates []*bool` (nil = not assessed), `running bool`, `LastTS int64` (`ts`). |
| `NewLive() *Live` | func | `campagne.go:200` | Constructor. |
| `func (l *Live) Set(s Snapshot)` / `Get() Snapshot` / `SetRunning(v bool)` | methods | `campagne.go:202-216` | `SetRunning` exists to avoid Get+Modify+Set lost-update race with `pumpSnapshots` (`campagne.go:209-211`). |
| `RunEvent(ctx, ev model.Event, prof model.Profile, d Deps) (model.Event, error)` | func | `campagne.go:302-703` | Full cell; returns frozen `ev` with `GateStatus` (`valid`/`degraded`/`invalid`) + `FailedGates`. |

Unexported: `qdiscRunner` (`campagne.go:157-159`), `httpDefault` (`144`), `cpuGate` (`149-155`), `defaults` (`218`), `validateShapeMode` (`278-299`), `costARPerH` (`707-712`), `round1` (`714`), `loadFor` (`716-721`).

### 2.2 `baseline.go` — G6 gate

| Export | Kind | Location | Notes |
|---|---|---|---|
| `BaselineStable(baseRTT []float64, prof model.Profile) (bool, float64)` | func | `baseline.go:52-61` | Returns `(ok, thresholdMs)`; `<5` samples fails (`baseline.go:53-57`). Threshold = `max(5, 0.2*median) + 2.0*JitterMs` (`baseline.go:65-72`, `jitterAllowance=2.0` at `baseline.go:45`). |

### 2.3 `capabilities.go` — host view

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Capabilities` | struct | `capabilities.go:8-14` | `Qdiscs`, `CCs`, `KnownCCs []string`, `Kernel string`, `BBRv3 bool` (JSON tags `qdiscs`, `ccs`, `known_ccs`, `kernel`, `bbr3`). |
| `CapabilitiesNow() Capabilities` | func | `capabilities.go:18-34` | `CCs` = executable only (`DefaultCCs()`), `KnownCCs` = accepted labels. |

### 2.4 `ccavail.go` (+ `ccavail_linux.go` / `ccavail_other.go`) — CC discovery

| Export | Kind | Location | Notes |
|---|---|---|---|
| `DefaultCCs() []string` | func | `ccavail.go:27-36` | `cubic`+`bbr` always, `bbr3` only if `ccPresent` exact-token match (`ccavail.go:15-22`). |

Platform impls (unexported vars): `readCCList`/`readKernelRelease` vars (`ccavail.go:9-11`) → `readHostCCList`/`readHostKernelRelease`. Linux reads `/proc/sys/net/ipv4/tcp_available_congestion_control` + `/proc/sys/kernel/osrelease` (`ccavail_linux.go:12-26`); non-Linux returns nil/"" (`ccavail_other.go:6-7`). Formats are **unverified — primary source not found**.

### 2.5 `matrix.go` — campaign orchestration

| Export | Kind | Location | Notes |
|---|---|---|---|
| `OnQuarantine func(runID string, eventID int, profile, qdisc, cc string)` | var | `matrix.go:14` | Package hook set by host (`cmd/cgo`); nil ⇒ silent. Fired when `GateStatus == invalid` (`matrix.go:235-237`). |
| `OnSkip func(runID string, eventID int, profile, qdisc, cc string)` | var | `matrix.go:19` | Fired on per-cell skip (not global stop) (`matrix.go:244-246`). |
| `Matrix` | struct | `matrix.go:24-35` | Exported `RunID string`, `Total int`, `Done int`, `Running bool`; unexported `mu`, `cancel`, `skipCancel`. |
| `func (m *Matrix) Skip()` | method | `matrix.go:44-51` | Cancels current cell only. |
| `func (m *Matrix) IsRunning() bool` | method | `matrix.go:54-61` | Nil-safe (`m==nil` ⇒ false). |
| `func (m *Matrix) Stop()` | method | `matrix.go:260-267` | Cancels whole matrix. |
| `StartMatrix(base ctx, profiles []string, reps int, deps Deps, dataDir string) (*Matrix, error)` | func | `matrix.go:65-68` | Full matrix (`allQdiscStrings()` × `DefaultCCs()`). |
| `StartMatrixFiltered(base, profiles, qdiscs, ccs []string, reps, deps, dataDir)` | func | `matrix.go:74-77` | Submatrix; empty axes or unknown values refused (`matrix.go:115-141`). |
| `StartMatrixWithID(base, runID, profiles, reps, deps, dataDir)` | func | `matrix.go:79-82` | Resume with given run ID. |
| `StartMatrixFilteredWithID(base, runID, profiles, qdiscs, ccs, reps, deps, dataDir)` | func | `matrix.go:87-93` | Resume filtered; empty `runID` refused. |

Unexported: `startMatrix` (`matrix.go:103-258`), `allQdiscStrings` (`95-101`), `newRunID` (`269-271`, `run-<unix>` via `clock.go:5`).

### 2.6 `plane.go` — preflight

| Export | Kind | Location | Notes |
|---|---|---|---|
| `PlaneReady(ctx, d Deps) error` | func | `plane.go:19-32` | 6 s budget; probes `SmallURL` (4 s) + TCP dial `BulkAddr` (3 s); refuses with `kit/testbed.sh up` hint. |

### 2.7 `prod.go` / `dualns.go` — production wiring

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ProdDeps() Deps` | func | `prod.go:24-26` | Upload: shaper on `veth-c` emission (client), netem same hop. |
| `ProdDepsDown() Deps` | func | `prod.go:31-33` | Download: shaper on `veth-s` via `NsRunner(cgo-srv)`, stats on server hop. |
| `BurstGE(profileID string) (qdisc.GEModel, bool)` | func | `dualns.go:18-26` | P2 (`0.001/0.1/0.5/0.0005`) P3 (`0.002/0.1/0.5/0.001`) only; others `false`. Parameters **unverified — primary source not found**. |
| `ProdDepsDualNS(direction string) Deps` | func | `dualns.go:112-181` | Hardened bench: netem in `ns_wan_emulator` (`w-edge`+`c-wan`), shaper in `ns_edge_router` (`e-srv`/`e-wan` or `ifb0` ingress), `Bench="dualns"`. |

### 2.8 `writer.go` — frozen CSV + manifest

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Writer` | struct | `writer.go:17-23` | Exported `Dir string`; unexported `seen`, `fh *os.File`, `dupes`, `quarantine`. |
| `OpenRun(dir string) (*Writer, error)` | func | `writer.go:35-77` | `MkdirAll`, replays `seen` from existing `aqm_eval.csv` (cell key `profile/qdisc/cc/direction/rep`, cols 2-6, `writer.go:54-61`), writes header if new. |
| `func (w *Writer) Append(ev model.Event) error` | method | `writer.go:132-152` | G5 dedupe via `cellKey` (`writer.go:83-85`); quarantines non-`valid` rows. 23-col format string (`writer.go:145-149`). |
| `func (w *Writer) Freeze(configJSON []byte) error` | method | `writer.go:159-200` | Closes CSV, refuses empty/invalid JSON (`writer.go:163-165`), writes `quarantine.json` + `manifest.json` (`config_sha256`, `config`, `files[]` with per-file SHA-256). |

### 2.9 `valuation.go` — enterprise sidecar

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ValuationLabels` | struct | `valuation.go:14-19` | `Flows`, `ChannelModel`, `Bench`, `ChargeSec` (JSON). `set()` unexported (`valuation.go:21`). |
| `EventValuation` | struct | `valuation.go:27-34` | `Cell`, `Bytes`, `Drops`, `Ratio metrics.RetransRatio`, `Primary *EconomicsScenario`, `Scenarios []EconomicsScenario`. |
| `ValuationSidecar` | struct | `valuation.go:39-44` | `Labels`, `Note`, `Disclaimer`, `Entries`. |
| `WriteValuation(dir string, ev model.Event, labels ValuationLabels) error` | func | `valuation.go:99-111` | Public wrapper; fills defaults (`flows 4`, `clean`, `legacy`, `model.ChargeSec`). |

Unexported: `writeValuationEntry` (`valuation.go:50-95`), `labelsFromDeps` (`116-134`).

### 2.10 `watch.go` — passive surveillance

| Export | Kind | Location | Notes |
|---|---|---|---|
| `OnWatchAlert func(p95, deadlineMs float64)` | var | `watch.go:15` | Package hook; fired after 10 consecutive p95 > deadline, ≤1/5 min (`watch.go:63-70`). |
| `StartWatch(base ctx, d Deps) (stop func())` | func | `watch.go:23-100` | 500 ms ticker, `Ping(3)` + `Small`, phase `surveil`, `Running=false` (`watch.go:89`). |

### 2.11 `config.go`, `clock.go`, `cpu_*.go`, `modprobe_*.go`, `instrument.go`

No exported symbols. `matrixConfig` (`config.go:15-36`) is the frozen manifest schema; `storedTier`/`tierForRun`/`buildMatrixConfig` (`config.go:41-121`) are unexported. `nowUnix` (`clock.go:5`), `readCPUUsage` (Linux `/proc/stat` 200 ms sample `cpu_linux.go:16-39`; non-Linux NaN `cpu_other.go:9`), `ensureQdiscModules` (Linux `sudo -n modprobe sch_netem sch_fq_codel sch_cake` `modprobe_linux.go:10-12`; no-op elsewhere), `logTC` (`instrument.go:8`, `//go:build instrumentation` only) are unexported.

### 2.12 Tests (pin behaviour, not API)

~60 `Test*` functions across 18 `*_test.go` files. They pin: G6 jitter-aware thresholds (`baseline_test.go`), BBRv3 exact-token gating (`ccavail_test.go`), live publish + counter-reset clamping (`campagne_test.go:187-370`), skip-vs-failure classification (`failclass_test.go`, `skip_journal_test.go`), ingress topology refusal (`ingress_test.go`), direction-aware G4 floor (`g4down_test.go`), `PlaneReady` refusal (`plane_test.go`), `ProdDepsDown` server-hop polling (`prod_down_test.go`), tier freeze on resume (`tier_freeze_test.go`), valuation idempotence (`valuation_test.go`), writer baseline columns (`writer_baseline_test.go`).

## 3. Dependencies

### 3.1 Outgoing (what `pkg/campagne` needs)

| Import | Used by | Notes |
|---|---|---|
| `github.com/Realms4239/cgo/pkg/model` | all files (`campagne.go:15`, `matrix.go:9`, `baseline.go:7`, `capabilities.go:3`, `config.go:9`, `watch.go:9`, `writer.go:11`, `valuation.go:9`) | `Event`, `Profile`, `Qdisc`/`CC` enums, `AllQdiscs`/`KnownCC`/`KnownProfiles`, gates `G0–G7`/`GateCount`/`GatePass|Degraded|Invalid`, phases `PhaseBaseline|Charge|Recup`, `BaselineSec=30`/`ChargeSec=120`/`RecupSec=30` (`pkg/model/types.go:108-125`), `AQMEvalHeader`, `ParseCCList` |
| `github.com/Realms4239/cgo/pkg/metrics` | `campagne.go:14`, `baseline.go:6`, `config.go:8`, `valuation.go:8`, `watch.go:8` | `Summarize` (`campagne.go:404`, `baseline.go:58`), `DeadlineOKPct` (`campagne.go:411`), `JFI` (`campagne.go:539`), `VoIPR` (`campagne.go:626`), `TierByName`/`CostARPerHTier`/`CostARPerH` (`campagne.go:708-711`), `ActiveTierName` (`config.go:68`), `CapacityWasteARPerHour`/`MeteredWasteARPerHour`/`PlanGOLD`/`TariffCorporateAPN` (`campagne.go:135-138`), `BuildEconomics`/`RetransRatio`/`EconomicsScenario` (`valuation.go:31-33`, `64-76`) |
| `github.com/Realms4239/cgo/pkg/probe` | `campagne.go:16`, `plane.go:9`, `dualns.go:9`, `prod.go:9` | `Ping` (`campagne.go:221`), `SmallObject`+`SmallClientTimeout` (`campagne.go:240`, `plane.go:23`), `BulkSendTo`/`BulkDownloadTo` (`campagne.go:506-510`), `BulkDownloadNTo` (`campagne.go:530`), `BulkSendNTo` (`prod.go:73`, `dualns.go:150`) |
| `github.com/Realms4239/cgo/pkg/qdisc` | `campagne.go:17`, `dualns.go:10`, `prod.go:10`, `watch.go:10` | `ApplyShaper`/`ApplyIngressShaper` (`campagne.go:360-368`), `ApplyNetemPhysics`/`ApplyNetemDistribution` (`dualns.go:57-101`), `SumDrops`/`SumBytes` (`campagne.go:391`, `campagne.go:415`), `PollStats` (`prod.go:80`, `dualns.go:155`), `LoadCalibration` (`dualns.go:78`), `CheckIngress` (`dualns.go:170`), `GEModel` (`dualns.go:18`), `ExecRunner`/`NsRunner`/`TCRunner` (`prod.go:41-53`, `dualns.go:120-123`) |
| stdlib | all | `context`, `fmt`, `math`, `net`, `net/http` (`campagne.go:6-18`), `sync`, `time`, `encoding/json`, `os`, `path/filepath`, `crypto/sha256`, `encoding/hex`, `log`, `os/exec` (`modprobe_linux.go:3`), `strings`/`strconv` (`cpu_linux.go:5-11`) |

No third-party Go module is imported directly by `pkg/campagne`: `go.mod:5-10` requires `fyne.io/fyne/v2`, `charmbracelet/bubbletea`, `lipgloss`, `golang.org/x/sys` — none is imported by this package (verified by grep `syscall|golang.org/x/sys|os/exec` → only `os/exec` in `modprobe_linux.go:3`). Transitive Go dependencies arrive only via the four internal packages above.

### 3.2 Module identity

`go.mod:1-3` declares `module github.com/Realms4239/cgo` with `go 1.25`. Per (R3) the module path is the unique identifier combined with the version number and the import prefix of every package; per (R3) the `go` directive is the mandatory minimum Go version (since Go 1.21 toolchains refuse newer-than-declared modules). Per (R3) a v2+ module path must end `/v2`. Per (R4) §1 a released module MUST declare its public API — §2 of this document inventories it — and per (R4) §3 a released version's contents MUST NOT be modified.

### 3.3 Runtime / OS dependencies (not Go modules)

- Linux-only shaping path: `tc` binary + kernel qdiscs `netem`, `fq_codel`, `cake`, `tbf`, `pfifo` (R1 §1: FQ-CoDel is the Linux `fq_codel` qdisc; `modprobe_linux.go:11` preloads `sch_netem|sch_fq_codel|sch_cake` via `sudo -n`). Non-Linux builds stub `readHostCCList`/`readHostKernelRelease`/`ensureQdiscModules`/`readCPUUsage` to nil/"" /no-op/NaN (`ccavail_other.go:6-7`, `modprobe_other.go:6`, `cpu_other.go:9`). `docs/architecture.md:87-89` formalises: Linux = full mode, Windows = observe (501 for shape/run/watch).
- `/proc` pseudo-files: `/proc/stat` CPU (`cpu_linux.go:42`), `/proc/sys/net/ipv4/tcp_available_congestion_control` CC list (`ccavail_linux.go:13`), `/proc/sys/kernel/osrelease` (`ccavail_linux.go:21`). Formats and semantics **unverified — primary source not found**.
- Network namespaces + veth topology: `veth-c`/`veth-s` (`prod.go:39-47`), `ns_wan_emulator`/`ns_edge_router`/`ns_client`/`ns_server` + `w-edge`/`c-wan`/`e-srv`/`e-wan`/`c-cli`/`ifb0` (`dualns.go:112-179`). Layout is in-repo convention; no primary source fetched.
- Filesystem: `data/runs/<run_id>/{aqm_eval.csv,manifest.json,quarantine.json,valuation.json}` (`writer.go:40`, `valuation.go:55`, `config.go:42`); calibrations in `CGO_CALIBRATION_DIR` default `data/calibrations` (`prod.go:65`, `dualns.go:122`). CSV schema is frozen by `AGENTS.md` ("`telemetry.csv`, `link_audit.csv`, `aqm_eval.csv` schemas" never touch) — no primary source beyond the repo.

## 4. Primary-source requirements

What a standalone `campagne` module must satisfy (or explicitly refuse) per each primary source:

### 4.1 RFC 8290 (R1) — AQM

- RFC 8290 is **Experimental, not Internet Standards Track** (R1, Status of This Memo) and defers the AQM half to CoDel (R1, §2: "we refer the reader to [RFC8289]"). The only MUST in the parameter section is the packet limit: "some packet limit MUST be enforced" (R1, §5.2.3). Defaults a module may cite: `interval 100 ms` (R1, §5.2.1, "SHOULD be on the order of worst-case RTT"), `target 5 ms` with low-bandwidth MTU-transmission floor (R1, §5.2.2, e.g. "~15 ms for 1 Mbps and MTU 1500"), `quantum 1514` (R1, §5.2.4), `flows 1024` (R1, §5.2.5, Linux range 1–65535), ECN enabled by default / `noecn` to disable (R1, §5.2.6), `ce_threshold` disabled by default (R1, §5.2.7).
- FQ-CoDel mixes DRR [DRR] + CoDel per queue, hashes the 5-tuple (Jenkins, salt-perturbed, R1 §4.1) into 1024 queues by default (R1, §3, §5.2.5), uses two-tier new/old round-robin with byte credits/quantum (R1, §1.3, §4.2), and drains the longest queue on overload (R1, §4.1). A standalone module measuring `fq_codel` cells inherits these semantics without reimplementing them — it only needs to *name* them honestly (see §5 G2).
- CAKE has **no RFC**: R1 cites it only as `[CAKE]` informative (§10.2) and "currently being developed" (§7). A module must never cite RFC 8290 as authority for CAKE behaviour; CAKE provenance stays **unverified — primary source not found** in this pass.
- RFC 8290 defines **none** of the campaign's own vocabulary: no G0–G7 gates, no baseline/charge/recovery split, no Gilbert-Elliott, no DANIEL/IFB, no tier/cost. Its legitimate use is framing only — sojourn-time / bufferbloat ("dark buffers", R1 Abstract; queue delay vs link busy, R1 §2) — not gate design.

### 4.2 Prometheus exposition (R2)

- Prometheus fundamentally stores time series: streams of timestamped values with a metric name + labels (R2, Data model). Every time series is uniquely identified by metric name + optional label pairs (R2). Metric names SHOULD match `[a-zA-Z_:][a-zA-Z0-9_:]*` with `:` reserved for recording rules (R2); label names SHOULD match `[a-zA-Z_][a-zA-Z0-9_]*` with `__` reserved for internal use (R2); label values MAY contain any UTF-8; empty label value ≡ absent (R2). Samples are float64 (or native histogram) + ms-precision timestamp (R2).
- Consequence for this package: `pkg/campagne` today emits **no Prometheus exposition at all**. `Snapshot` is a JSON struct for SSE (`campagne.go:167-198`, pushed via `OnSnap` `campagne.go:97`, rendered by `pkg/api/sse.go` per `docs/architecture.md:25-41`), not a text-format scrape target. A standalone module that claims Prometheus compatibility must add a purpose-built `RenderText`/`Exposition` function written against R2 rules above (HELP/TYPE before first sample, unique metric+labels per line, escaping, `_sum`/`_count`, `{quantile}`/`{le}` conventions) — copying `Snapshot` JSON is not exposition. The module must pick Prometheus text vs OpenMetrics and declare it (cf. `docs/research/2026-11-07-module-metrics.md:138-144`).

### 4.3 Go modules / semver (R3, R4)

- Module path = unique ID + import prefix (R3, §"module"); v2+ requires `/v2` suffix (R3). Today's `github.com/Realms4239/cgo` (`go.mod:1`) is a monorepo path — a standalone `campagne` needs a new path it controls, chosen before first publish to avoid rename (R3, "Notes").
- `go` directive is the mandatory minimum Go version (R3, since Go 1.21). Extraction must bump or pin `go` deliberately; `go 1.25` today (`go.mod:3`) — a consumer on Go 1.21 cannot build it.
- Software using SemVer MUST declare a public API (R4 §1). Versioning: MAJOR = incompatible change, MINOR = backward-compatible addition, PATCH = backward-compatible fix (R4 §§7–9); released versions MUST NOT be modified (R4 §3); `0.y.z` MAY change and is not stable (R4 §4). Go versions are `vX.Y.Z` tags, not SemVer strings (R4 FAQ); `require` form is `module-path v1.2.3` or pseudo-version (R3 §"require").

## 5. Gap analysis (current → standalone)

Ordered by severity. Each gap cites local evidence; the primary-source side cites R# or is marked unverified.

- **G1 — `Deps` is a god-object god-constructor.** `Deps` has 30+ fields (`campagne.go:20-98`) mixing hops, addresses, probe funcs, policy, windows, progress, and the `OnSnap` callback. `defaults()` (`campagne.go:218-273`) silently fills Ping/Small/CPU/Now/windows/channel/bench/tier/delay/shape defaults. `Snapshot()` takes dead params `base, chg map[string]float64, bulk uint64` (`campagne.go:104-105`) that are never read. Any caller constructing `Deps` by hand must know which nils are fatal vs defaulted — unwritten contract. Standalone: split into `CellConfig` (policy) + `Probes` (funcs) + `Topology` (hops), or document every nil's fate.
- **G2 — AQM semantics are inherited, not asserted.** `campagne` never names RFC 8290 parameters; it calls `qdisc.ApplyShaper` which pins `target 5ms interval 100ms limit 10240 quantum 1514 flows 1024` for fq_codel-under-tbf (per `docs/research/2026-11-07-module-qdisc-remediate.md:36`), matching R1 §5.2 defaults. But `campagne` itself exposes no `FqCodelParams`; if a consumer needs a different `target` they cannot express it — the cell is either default fq_codel or not. Conversely CAKE cells carry only `bandwidth+rtt` in the bench while the production recipe adds classification knobs (R3; **unverified — primary source not found**). The module must decide whether it measures "fq_codel as RFC 8290 describes it" or "whatever `qdisc` emits today" and expose that as a typed param.
- **G3 — No Prometheus exposition; `Snapshot` is the de-facto metrics surface.** `Snapshot` JSON fields (`rtt_p50_ms`, `small_p95_ms`, `bulk_goodput_mbps`, `drops`, `cost_*`, `jfi_pct` — `campagne.go:167-198`) are gauge-like, but there is no `HELP`/`TYPE`, no timestamps beyond `LastTS` (`ts`), no label model (R2), no `_total`/`_count`/`_sum` companions, and `gates []*bool` encodes tristate as JSON null/false/true with no exposition equivalent. A Prometheus scrape of `Snapshot` would need a fresh mapping written against R2; today `/api/metrics` (elsewhere in-repo) is the only exposition and `campagne` does not feed it. Claiming "metrics" for the standalone tool is therefore **unverified — primary source not found** until that mapping exists.
- **G4 — Package-level mutable hooks are global state.** `OnQuarantine`, `OnSkip` (`matrix.go:14`, `19`) and `OnWatchAlert` (`watch.go:15`) are exported vars any importer can overwrite, set by host (`cmd/cgo`) per comments. Two matrices in one process share them; tests run "tranquilles" only because they never set them. Per (R4 §1) the public API must be precise — global hooks are not. Standalone: replace with a `Observer`/`Hooks` struct field on `Deps`/`Matrix`.
- **G5 — Internal packages are hard deps, extraction drags the whole repo.** `campagne` imports `model`, `metrics`, `probe`, `qdisc` (`campagne.go:14-17`). `model` is the domain vocabulary (`Profile`, `Event`, gates, phases); `metrics`/`probe`/`qdisc` are the other standalone candidates. Extracting `campagne` alone without those three is impossible without redefining their types — the module graph is a diamond, not a leaf (cf. `docs/research/2026-11-07-module-qdisc-remediate.md:126-129`). The `model.Profile` global `ProfilesMu` + `Profiles` map (`matrix.go:179-181`, `config.go:73-84`) is read without injection, so a standalone `campagne` changes behaviour when another package mutates the registry.
- **G6 — Linux-only surface hidden behind a cross-platform API.** `PlaneReady`, `StartMatrix*`, `RunEvent`, `StartWatch`, `ProdDeps*` all compile on any GOOS (stubs in `*_other.go`), but only Linux can actually shape (`docs/architecture.md:87-89`: Windows = observe, 501). No `//go:build linux` gate on the orchestrator; the failure shows up as a `tc` error mid-cell, not a compile error. `ensureQdiscModules` (`modprobe_linux.go:10-12`) even shells `sudo -n modprobe` best-effort and ignores failure. A standalone library should fail closed at construction (`ErrUnsupportedPlatform`) instead of at first `tc` call.
- **G7 — Filesystem layout + frozen-schema coupling.** `Writer` hardcodes `aqm_eval.csv` + `manifest.json` + `quarantine.json` (`writer.go:40`, `159-200`) and a 23-col `fmt.Sprintf` (`writer.go:145-149`) that must stay byte-compatible with `model.AQMEvalHeader` and the frozen corpus (`AGENTS.md`: never reshape `aqm_eval.csv` schema, never touch `book/data/runs`). `startMatrix` builds `dataDir + "/" + m.RunID` (`matrix.go:163`) with string concat, not `filepath.Join`. `runID` is `run-<unix>` (`matrix.go:269-271`) — collision-prone under fast successive starts. Standalone: the writer must be a `RunStore` interface (file impl + in-memory test impl), or the module ships CGO's directory convention as API.
- **G8 — Numbers that must be right are computed with fragile heuristics.** `WastedBytes = Drops × 1448` (`campagne.go:637`) assumes a 1448-byte MSS for veth MTU 1500 — stated in comment but nowhere verified against the actual interface MTU. `costARPerH` (`campagne.go:707-712`) silently falls back to the global tier when the frozen tier name is unknown (`campagne.go:710-711` returns `CostARPerH` without error), and `tier_freeze_test.go:48-49` pins exactly that fallback. `G4` floor switches from 50 % of `CapUp` to 1 % of `CapDown` by direction (`campagne.go:649-655`) with a bench-observed rationale in comment. `liveTransported` accumulates `tc -s` byte deltas with a `g <= 2500` clamp (`campagne.go:424`) and a `int64` wrap guard (`campagne.go:419`, `601-609`). Each is defensible in-repo; none carries a primary source — all **unverified — primary source not found** as standalone claims.
- **G9 — Dead / duplicate code at the seams.** `Writer.dupes` (`writer.go:21`, incremented `writer.go:135`) is never read or exported. `matrixConfig.Kernel` (`config.go:32`) is frozen but never consumed on load. `Deps.Snapshot`'s `base`/`chg`/`bulk` params are dead (§2.1). `ccPresent` token match (`ccavail.go:15-22`) and `DefaultCCs` (`ccavail.go:27-36`) duplicate the `bbr3` gate in `matrix.go:138-140` and `docs/api.md:71-73`. `cpuGate` nil-on-NaN (`campagne.go:149-155`) and `readCPUUsage` NaN-on-!Linux (`cpu_other.go:9`) agree, but `G7CPUNotSaturated` is set via `gates[G7] = cpuGate(cpuAvg)` (`campagne.go:656`) while other gates use `set()` — bypassing the helper.
- **G10 — Error taxonomy collapses into strings.** `startMatrix` and `RunEvent` return bare `fmt.Errorf` strings (`matrix.go:90`, `105-141`, `campagne.go:315`, `353-355`, `361-370`), including user-facing French hints with CLI remediation (`plane.go:24`, `28`). `RunEvent` returns `ctx.Err()` for skip (`campagne.go:674-676`) but the matrix classifies skip via `evCtx.Err() != nil && ctx.Err() == nil` (`matrix.go:238`) — a second, fragile channel. `Writer.Append` returns `fmt.Errorf("duplicate event row %s")` (`writer.go:136`) with no sentinel. A Go module needs typed errors (`ErrBurstInvalid`, `ErrIngressPreflight`, `ErrDuplicateCell`, …) per (R4 §1) precise API, not string matching.
- **G11 — AQM/queueing claims in comments outrun sources.** `campagne.go:621-626` freezes `QDIPctMs` as "P95 chargé − P50 chargé" and calls it "La métrique queue-delay de référence" — but R1 defines queue delay via sojourn time / CoDel target (R1 §5.2.2), not this P95−P50 proxy. `BurstGE` P2/P3 parameters (`dualns.go:20-23`) cite "Cibles vérifiées par `qdisc.GEMeanLoss`" — the Gilbert-Elliott mean-loss formula itself is **unverified — primary source not found** (see `docs/research/2026-11-07-module-qdisc-remediate.md:24`). `BaselineStable`'s `jitterAllowance=2.0` (`baseline.go:45`) is justified by a host-noise argument in comment, not by any RFC/procfs primary — **unverified — primary source not found** as a standalone threshold.

## 6. Target module API

Proposed contract for a standalone module. Per (R3) it lives behind a new module path with its own `go.mod` and `go` directive; per (R4) `v0.1.0` is unstable (`0.y.z` MAY change, R4 §4), minor = backward-compatible addition, major (with `/v2` per R3) = breaking change. Per (R4 §1) this declaration *is* the public API once published.

```go
// Module path: github.com/<org>/campagne  (R3; chosen before first publish)
package campagne

// --- Configuration (replaces Deps god-object, G1) ---
type Topology struct {
    // NSS are opaque handles; Env is the only hook for TC_LIB_DIR etc.
    NetemRunner  Runner            // netem hop (veth-c / w-edge)
    ShaperRunner Runner            // shaping+AQM hop; nil => NetemRunner
    CliIf, ShaperIf string
    PhysicsIfaces []string
    Direction     Direction          // Up | Down | Both (both = two cells, never one)
    ShapeMode     ShapeMode          // Egress | Ingress
    IngressPreflight func() error    // nil + Ingress => construction error (G6)
}
type Direction string; const ( DirectionUp Direction = "up"; DirectionDown Direction = "down" )
type ShapeMode string; const ( ShapeModeEgress ShapeMode = "egress"; ShapeModeIngress ShapeMode = "ingress" )

type Probes struct {
    Ping func(ctx context.Context, target string, n int) []float64
    Small func(ctx context.Context) (float64, error)
    Bulk func(ctx context.Context, addr string) (uint64, error)
    BulkN func(ctx context.Context, addr, cc string, n int) ([]uint64, error) // nil => single flow
    CPU func() float64   // NaN => not measured (G7 nil), never 0-as-pass
    Now func() time.Time
    StatsFn func() []Stats // []qdisc.Stats by value; nil => no tc -s
}
type CellPolicy struct {
    Profile Profile         // local copy: replaces model.Profile global read (G5)
    Qdisc Qdisc             // cake | fq_codel | tbf-fifo (no pfifo_fast label; cf. qdisc research G4)
    CC CC                    // cubic | bbr | bbr3 (bbr3 exact-token gated, ErrCCUnavailable)
    Repetition int
    Flows int               // <=0 => 4; 1 = historical single-flow
    Channel ChannelModel    // Clean | Bursty (bursty only P2/P3, else ErrBurstInvalid before mutation)
    Delays  DelayModel      // Scalar | Dist (dist requires dualns + calibration, fail-closed)
    DeadlineMs float64
    Windows Windows         // BaselineSec, ChargeSec, RecupSec (30/120/30 defaults)
    Tier TierRef            // frozen at run start; unknown tier => error, never silent fallback (G8)
}
type Hooks struct { // replaces package vars OnQuarantine/OnSkip/OnWatchAlert (G4)
    OnQuarantine func(runID string, eventID int, cell CellID)
    OnSkip func(runID string, eventID int, cell CellID)
    OnSnap func(Snapshot)
}
type Deps struct { Topology; Probes; Hooks } // Node graph is Topology+Probes+Hooks; CellPolicy travels per cell.

// --- Cell execution (replaces RunEvent signature, G10) ---
type GateStatus string; const ( GateValid GateStatus = "valid"; GateDegraded GateStatus = "degraded"; GateInvalid GateStatus = "invalid" )
type Event struct { // local mirror of model.Event; frozen fields documented per-field
    RunID, Direction string; EventID, Repetition int
    Profile string; Qdisc Qdisc; CC CC
    RTTp50Ms, RTTp95Ms, Smallp95Ms, BulkGoodputMbps float64
    Drops, WastedBytes, Bytes uint64; CostARPerH float64
    CPUPct, DeadlineOKPct, JFIPct, QDIPctMs, VoIPR float64 // QDIPctMs documented as P95-P50 proxy, NOT RFC sojourn (G11)
    RTTBaseP50Ms, RTTBaseP95Ms float64
    G7 *bool // nil = not assessed; 0% never passes
    GateStatus GateStatus; FailedGates []string
}
func RunEvent(ctx context.Context, ev Event, pol CellPolicy, d Deps) (Event, error) // typed errors: ErrBurstInvalid, ErrIngressPreflight, ErrShapeInvalid, ErrCCUnavailable

// --- Matrix ---
type Matrix struct { RunID string; Total, Done int; Running bool }
func StartMatrix(base context.Context, profiles []Profile, qdiscs []Qdisc, ccs []CC, reps int, deps Deps, store RunStore) (*Matrix, error)
func ResumeMatrix(base context.Context, runID string, profiles []Profile, qdiscs []Qdisc, ccs []CC, reps int, deps Deps, store RunStore) (*Matrix, error)
func (m *Matrix) Skip(); func (m *Matrix) Stop(); func (m *Matrix) IsRunning() bool

// --- Persistence (G7) ---
type RunStore interface { OpenRun(dir string) (*Writer, error) } // file impl + in-memory test impl; schema version field required
func OpenRun(dir string) (*Writer, error)
func (w *Writer) Append(ev Event) error   // ErrDuplicateCell on G5 violation
func (w *Writer) Freeze(configJSON []byte) error // refuses invalid JSON; manifest carries schema+sha256

// --- Live + capabilities ---
type Snapshot struct { /* same JSON fields as today, plus Schema int `json:"schema"` */ }
type Live struct{ /* mutex + current Snapshot */ }
func NewLive() *Live; func (l *Live) Set(Snapshot); func (l *Live) Get() Snapshot; func (l *Live) SetRunning(bool)
type Capabilities struct { Qdiscs, CCs, KnownCCs []string; Kernel string; BBRv3 bool }
func CapabilitiesNow() Capabilities
func DefaultCCs() []string   // executable only
func PlaneReady(ctx context.Context, d Deps) error  // ErrPlaneUnreachable with bench hint
func BaselineStable(baseRTT []float64, prof Profile) (bool, float64) // threshold 2.0 documented as heuristic, unverified — primary source not found
func BurstGE(profileID string) (GEModel, bool) // parameters unverified — primary source not found

// --- Production wiring (topology as data, not env-reading constructors) ---
func ProdDeps(cfg ProdConfig) Deps          // ProdConfig carries env-resolved values; no os.Getenv inside
func ProdDepsDualNS(cfg DualNSConfig) Deps  // Bench="dualns", DelayModel, ShapeMode as fields

// --- Prometheus exposition (NEW, written against R2; G3) ---
type Sample struct { Metric string; Labels map[string]string; Value float64; TimestampMs int64 }
func RenderSnapshotText(samples []Sample) (string, error) // HELP/TYPE first, unique metric+labels, escaping per R2, LF-terminated; or declare OpenMetrics + `# EOF`
```

Design notes: (a) `model.Profile`/`model.Event` become local copies — the `ProfilesMu` global read (`matrix.go:179`) and `model.Profiles` map read (`config.go:73`) disappear (G5); (b) hooks are per-`Deps`, not package vars (G4); (c) `Tier` unknown ⇒ error, not fallback (G8); (d) `Writer` is behind `RunStore` so the 23-col CSV is one impl among others (G7). What does **not** move: `pkg/qdisc` keeps owning `tc` execution, `pkg/probe` keeps owning ping/small/bulk, `pkg/api` keeps owning HTTP/SSE and `POST /api/run/*` mapping (`docs/api.md:63-65`).

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it.

| Slice | Scope | Gate |
|---|---|---|
| S1 | Decouple from globals: local `Profile`/`Qdisc`/`CC`/`Event` copies; `CellPolicy.Profile` replaces `model.ProfilesMu`+`Profiles` reads (`matrix.go:179-181`, `config.go:73-84`); `Hooks` struct replaces `OnQuarantine`/`OnSkip`/`OnWatchAlert` vars (`matrix.go:14`, `19`, `watch.go:15`) | `grep -rn 'pkg/model\|pkg/metrics\|pkg/probe\|pkg/qdisc' pkg/campagne` narrowed to interfaces only; two matrices in one process with different `Hooks` do not cross-fire (new test) |
| S2 | Split `Deps` god-object: `Topology` / `Probes` / `Hooks` / `CellPolicy`; drop dead `Snapshot(base,chg,bulk)` params (`campagne.go:104-105`); document every nil's fate | `go vet` clean; `TestRunEventHappyPath`/`TestRunEventLivePublish` pass unchanged on the new struct |
| S3 | Typed errors + fail-closed construction: `ErrBurstInvalid`, `ErrIngressPreflight`, `ErrShapeInvalid`, `ErrCCUnavailable`, `ErrDuplicateCell`, `ErrUnsupportedPlatform`; remove `sudo -n modprobe` silent best-effort (`modprobe_linux.go:10-12`) or make it an explicit `PreloadModules` opt-in | `TestFailureIsNotASkip` (`failclass_test.go`) asserts error identity, not string; non-Linux construction fails before first `tc` call |
| S4 | `RunStore` interface + in-memory impl; `filepath.Join` for all run paths; `runID` uniqueness (monotonic + pid, not `run-<unix>` `matrix.go:269-271`); 23-col CSV behind schema-version field | round-trip test: in-memory store receives the same 23 cols; `openchamber`/`make verify` green |
| S5 | AQM assertion layer: expose `FqCodelParams` (defaults = R1 §5.2 values) and `CAKEParams` as `CellPolicy` fields; document `QDIPctMs` as P95−P50 proxy, NOT RFC sojourn (G11) | arg-vector test: fq_codel emitted with all five R1 values; CAKE delta vs production recipe named in docs |
| S6 | Prometheus exposition: `RenderSnapshotText` written against R2 (HELP/TYPE order, unique metric+labels, escaping, LF termination, `+Inf` bucket if histogram chosen) with golden tests; declare Prometheus-text vs OpenMetrics in README | R2-conformant scrape of the module's own exposition succeeds; `Snapshot` JSON path kept for SSE back-compat |
| S7 | Module split: new `go.mod` (`module github.com/<org>/campagne`, `go` directive per R3), `v0.1.0` tag (R4 §4 unstable), README declares public API (§2) per R4 §1; in-repo callers migrate via `require` + `replace` during migration, then drop `replace` | `go list -m all` in the new module shows no `github.com/Realms4239/cgo/pkg/*` internals; `make verify` green with the module consumed by `pkg/api`/`cmd/cgo` |
| S8 | Verification slice: re-run verifier on `go test ./...` + `python tools/verify.py` per `AGENTS.md` Roles; `VERIFIED` rows in `docs/VERIFY.md` only after verifier pass, never builder self-certification | verifier owns the row; "Windows-green" is not "done" (`AGENTS.md` Cadence) |

Each slice is independently reviewable; S5–S6 are the only ones that change behaviour-visible semantics and each should be verifier-re-run before its `VERIFIED` row is claimed.

## 8. Anti-toy criteria (what would make this a toy — and the bar for "not a toy")

A standalone claim of "works" for `campagne` is not credible unless all of the following hold. They are deliberately stricter than the current host tests (which run against `FakeRunner` and function-field probes, e.g. `campagne_test.go:53-101`):

1. **Real `tc` + real namespaces on a Linux VM.** At least one full `RunEvent` and one full `StartMatrix` (2+ cells) executed against real `tc`/netns (not `FakeRunner`), with `tc -s qdisc show` output recorded; ingress preflight (`plane.go:19-32`) and stale-qdisc reset (`campagne.go:347-354`) exercised on the VM, not only in `ingress_test.go`/`campagne_test.go`. No silent `sudo modprobe` failure can hide a missing qdisc.
2. **RFC 8290 parameters asserted, not commented.** A test fails if an `fq_codel` cell is emitted without `target 5ms interval 100ms limit 10240 quantum 1514 flows 1024` (R1 §5.2), and the test names R1 §5.2 as its source. CAKE cells either emit the production classification knobs or the delta is documented as measured — not "close enough".
3. **Gates mean what they say.** `BaselineStable` passes a synthetic LEO-wide baseline (≥17 ms spread, `baseline_test.go:27-53` scenario) only because of the `2.0×jitter` term; a thrashing host still fails. `cpuGate` returns nil (never `true`) on NaN/0/unmeasured (`campagne.go:149-155`); `WastedBytes = Drops×1448` and the `×20`/`2500` heuristics are labelled heuristics with MTU/threshold provenance, or replaced by measured values. No `0` reads as "measured 0" — absent is nil / error / explicit `Measured:false`.
4. **No global mutable state.** Two matrices, two `Hooks`, two `RunStore`s in one process cannot cross-fire (G4) and cannot corrupt each other's tier or profile registry (G5). `grep -rn 'var On' pkg/campagne` returns nothing; `go vet` + `race` clean.
5. **Frozen means frozen.** `Freeze` refuses non-JSON config (`writer.go:163-165`), `Append` refuses duplicate cells (`writer.go:134-136`), `StartMatrix*` refuses empty axes / unknown qdisc/CC / missing `bbr3` (`matrix.go:105-141`, `docs/api.md:63`), `PlaneReady` refuses before the first cell (`plane.go:19-32`), and resume re-reads the frozen tier (`config.go:64-69`) instead of the live global. A run that has frozen cannot be silently re-frozen with a different config hash.
6. **A real consumer + a real scrape.** At least one out-of-tree importer (not only `pkg/api`/`cmd/cgo` in this repo) drives a matrix and reads a frozen run, AND one R2-conformant Prometheus scrape succeeds against the module's own exposition (`RenderSnapshotText`, §6) — otherwise the "metrics" claim is decoration. The SSE JSON path alone does not count.
7. **Version discipline.** Declared public API (R4 §1), tags `vX.Y.Z` immutable once pushed (R3 §"retract"), changelog in MAJOR/MINOR/PATCH terms (R4 §§7–9). A release that changes emitted `tc` lines or the CSV schema is at least a minor version even if the Go API is unchanged, because the public API here includes the generated config and the frozen files (R4 §1). A repo that edits a released tag, or ships no declared API, is a toy regardless of code quality.

---
*Teams note: slices S1–S8 are proposals in this research file only. No `pkg/campagne` source was modified in this pass; implementation waits for human approval per `AGENTS.md` ("Stop and ask" — live bench, new dependency, scope growth). Report verified against primary sources R1–R4; everything else is labelled `unverified — primary source not found` at the claim site.*
