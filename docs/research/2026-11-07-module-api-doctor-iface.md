# pkg/api + pkg/doctor + pkg/iface as standalone tools — primary-source research

*Date: 2026-11-07 · Scope: `pkg/api`, `pkg/doctor`, `pkg/iface` only. Working dir `operator-cockpit`. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | Prometheus — Exposition formats (Prometheus Text Format) | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched (webfetch, full page 2026-11-07) | `text/plain; version=0.0.4` Content-Type, UTF-8 `\n` lines, `# HELP`/`# TYPE` rules (one per metric, TYPE before first sample), sample EBNF, `counter/gauge/histogram/summary/untyped` TYPE vocabulary, grouping/ordering, float/timestamp grammar |
| R2 | Go Modules Reference | https://go.dev/ref/mod | fetched (webfetch, 2568 lines, 2026-11-07) | module = packages released/versioned/distributed together, module path = ID + import prefix, `go`/`require`/`replace`/`exclude`/`retract` semantics, major-version suffix `/vN` for v2+ and import-compatibility rule, canonical semver versions (`v` + SemVer 2.0.0) |
| R3 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | §1 MUST declare public API; §2 `X.Y.Z` form; §3 immutability once released; §6 PATCH = backward-compatible bug fix; §7 MINOR = backward-compatible functionality; §8 MAJOR = backward-incompatible API changes; §4 `0.y.z` MAY change |
| R4 | `proc_pid_net(5)` — Linux man page (man-pages 6.19, 2026-02-08) | https://man7.org/linux/man-pages/man5/proc_pid_net.5.html | fetched | `/proc/net/dev` is network-device status (received/sent packets, errors, collisions) with the canonical 16-field header (`bytes packets errs drop fifo frame compressed multicast | bytes packets errs drop fifo colls carrier compressed`); `/proc/net` is per-netns via `/proc/self/net` |
| R5 | Linux Kernel docs — IP Sysctl | https://docs.kernel.org/networking/ip-sysctl.html | fetched | `tcp_available_congestion_control` — STRING, "Shows the available congestion control choices that are registered"; `tcp_allowed_congestion_control` is a subset; `tcp_congestion_control` sets the algorithm for new connections |
| R6 | `ethtool(8)` — Linux man page (man7.org) | https://man7.org/linux/man-pages/man8/ethtool.8.html | fetched (webfetch text, 2026-11-07) | `-k/--show-features/--show-offload` queries state of protocol offload and other features; `-K/--features/--offload` built-in names include `tso` (TCP segmentation), `gso` (generic segmentation), `gro` (generic receive offload), `lro`, etc. |
| R7 | HTML Standard — 9.2 Server-sent events (WHATWG, Living Standard 2026-11-06) | https://html.spec.whatwg.org/multipage/server-sent-events.html | fetched | `text/event-stream` MIME, UTF-8 always, wire fields `data`/`event`/`id`/`retry` + `:` comments, blank-line dispatch, `Last-Event-ID` header on reconnect, `retry` field = ASCII digits only (reconnection time), authoring notes (comment keepalive vs. proxy timeout) |
| R8 | Go stdlib `net/http` reference (pkg.go.dev, go1.27.3) | https://pkg.go.dev/net/http | fetched (webfetch, 4545 lines) | `ServeMux` `Handle`/`HandleFunc(pattern)` with `"METHOD /path"` patterns, `MaxBytesReader`, `Flusher`, `Status*` codes, `FileServer`/`StripPrefix` |
| — | `tc(8)`, `ip netns(8)`, `capabilities(7)` (CAP_NET_ADMIN), BBR/BBRv3 algorithm internals, Fetch/Sec-Fetch-Site spec, RouterOS/SNMP | — | unverified — primary source not found (not fetched in this pass) | any `tc`/`ip netns` behaviour beyond binary-presence; privilege semantics; congestion-control algorithm properties beyond the R5 STRING listing |
| — | Repo | `pkg/api/*.go` (non-test), `pkg/doctor/*.go`, `pkg/iface/*.go`, `docs/api.md`, `go.mod:1-3`, `cmd/cgo/server.go` (composition) | read 2026-11-07 | all `file:line` claims below |

Method: read `pkg/api/server.go` (1524 lines), `sse.go` (235), `metrics.go` (64), `translate.go` (125), `auth.go` (26), `duel.go` (136), `server_audit.go` (499), `kernel_bridge.go` (30), `pkg/doctor/doctor.go` (183), `pkg/iface/iface.go` (200), plus tests (`pkg/api/metrics_test.go`, `sse_test.go`, `pkg/doctor/cc_test.go`, `offload_test.go`, `pkg/iface/iface_test.go`, `sampler_test.go`), `docs/api.md` (116 lines), `go.mod:1-3` (`module github.com/Realms4239/cgo`, `go 1.25`). Every technical claim below cites `file:line` or `(R#)`; anything else is labelled **unverified — primary source not found**.

## 1. Responsibility (as-shipped)

### 1.1 `pkg/iface` — interface throughput observer (stdlib-only)

Single responsibility: turn cumulative kernel counters into a rate `Sample` for one named interface, with an explicit validity/gap contract so a zero is never read as a measurement. `pkg/iface/iface.go:1-3` states "débit d'interface observé par les compteurs noyau. stdlib seul : /proc/net/dev sur Linux". Concretely:

- Parse `/proc/net/dev` (or a same-format file) into per-interface `Counters{RXBytes, RXDrops, TXBytes, TXDrops}` (`iface.go:20-25`, `iface.go:62-100`).
- Derive `RXMbps`/`TXMbps` over real elapsed windows (`iface.go:169-170`: `bytes*8/1e6/dt.Seconds()`), carrying `WindowMs`, `SampleTS`, drop counters, and `Valid`/`Gap` (`iface.go:30-41`).
- Expose a two-method seam `Reader{Read(); Latest()}` (`iface.go:44-47`) so a future SNMP/RouterOS adapter can return the same `Sample` (spec reference `docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md:591`; the adapter itself is unverified — primary source not found).
- Declare provenance, never guess it: `NormalizeSource` maps only `twin|prod|unknown` through, everything else to `unknown` (`iface.go:51-57`).
- Feed the live frame additively: `cmd/cgo/server.go:27-30` embeds `iface.Sample` in `liveFrame{campagne.Snapshot; iface.Sample}`, `cmd/cgo/server.go:66-73` wires `GetSnap→Latest()` and `NextSnap→Read()`; `docs/api.md:28-36` documents the `iface_*` fields and `iface_valid:false` invalidating all numerics, with `iface_gap ∈ {warmup, absent, reset, slow, read, clock}`.

What it deliberately does not do: no packets exposure (`iface.go:19` "les paquets ne sont pas exposés, personne ne les affiche"), no auto-detection of the interface (`cmd/cgo/server.go:32-33` "Pas d'auto-détection : une mauvaise interface serait une mesure mal étiquetée"), no shaping/probing, no persistence.

### 1.2 `pkg/doctor` — read-only capability report

Single responsibility: report the same capability truth the dashboard badge shows, without mutating the host. `pkg/doctor/doctor.go:1-3` states "capability report (grilling Q11): one line per capability, la même vérité que le badge du tableau de bord. Rien ici ne mute l'hôte". Concretely:

- Resolve mode: `Mode(override)` (`doctor.go:24-34`) — explicit `observe|full` wins, else Windows→`observe`, else `full`.
- Report checks: `os`, `tc` (LookPath), `cap_net_admin` (Windows warn / euid 0 ok / else warn), `bbr`+`bbr3` (exact-token parse of the R5 file, `doctor.go:152-183` via `model.ParseCCList` in `pkg/model/cc.go:8-18`), `bench-offloads` (GRO/GSO/TSO on six bench transit interfaces, `doctor.go:123-147`), `ping` (LookPath) (`doctor.go:37-65`), with `Print()` CLI form (`doctor.go:68-74`).
- Parse `ethtool -k` output tolerantly (`ParseOffloadFeatures`, `doctor.go:81-106`: `feature: on|off` lines, `[fixed]`/`[requested on]` suffixes ignored by taking the first state word).

What it deliberately does not do: no `tc` application, no `modprobe`, no offload mutation, no benchmark run. Its only dependency is `model` (for `ParseCCList`), otherwise stdlib (`doctor.go:6-14`).

### 1.3 `pkg/api` — REST + SSE surface (stdlib transport, campaign-coupled)

Single responsibility: expose the campaign/audit/profile/cost state over HTTP + SSE with honest-empty semantics (an absent measurement is absent, never a zero). `pkg/api/sse.go:1-2` states "surface REST + SSE (docs/SPEC.md §2.4–2.5). stdlib seul." Concretely:

- Route table: 40+ `mux.HandleFunc("METHOD /path")` registrations in `New` (`server.go:380-1432`), documented in `docs/api.md:13-116`. Split: health/state/schema/cost/profiles/shape/watch/run/results/duel/compare/integrity/report/figures/replay/audit/remediate/quarantine/events/SPA + `GET /metrics` (`server.go:394-399`) + `GET /api/stream` via `hub.SSE` (`server.go:1275-1276`).
- JSON contract helpers: `writeJSON` (always `application/json`, `server.go:1464-1470`) vs `writeErr` (negotiates `application/json` if Accept/Content-Type mentions it, else `text/plain` via `http.Error`, `server.go:1474-1483`); all JSON bodies capped `http.MaxBytesReader(..., 1<<20)` (e.g. `server.go:459`, `487`, `578`, `601`, `619`) as `docs/api.md:4` states.
- SSE hub at 10 Hz (`frameHz=10`, `sse.go:14`) with structural-delta compression, 2048-frame ring, 64-subscriber cap, `Last-Event-ID` replay, stale-client full resync, and a named `backpressure` event (`sse.go:71-124`, `sse.go:127-215`).
- Prometheus-shaped `/metrics` exposition builder (`metrics.go:17-64`) — see §6 for why the endpoint is currently a toy.
- Mutation hardening: `requireAuth` Bearer token + loopback-only dev mode (`auth.go:11-26`), `crossSiteGuard` refusing `Sec-Fetch-Site: cross-site` on non-safe verbs (`server.go:1438-1450`, `docs/api.md:5-7`), `observeBlocked` 501 for control verbs in observe mode (`server.go:220-229`, `docs/api.md:8-11`), `validRunID` anti-traversal guard (`server.go:320`).
- Audit orchestration lifecycle: `handleAuditStart/Status/Live/Stop/Report` + `handleRemediate/Presets/List/ToProfile` (`server_audit.go:26-499`), with `auditParamsFromRequest` mapping (`server.go:1488-1524`) and `readAuditReport`/`auditReportExists` artifact lookup (`server.go:44-86`).
- Translation + duel + kernel bridge: `HandleTranslate` (`translate.go:16-125`), `handleDuel` (`duel.go:62-120`), `kernelTelemetryFromSummary` (`kernel_bridge.go:12-30`).

What it deliberately does not do: it applies no shaping itself (`Deps.ShapeFn`, `server.go:186`), runs no probe itself (`Deps.BurstFn/WatchFn/StartFn`, `server.go:178`, `195`, `206`), computes no percentile itself (delegates to `results`/`metrics`, e.g. `server.go:334-378`, `779-923`). Composition (snapshots, engine wiring) lives in `cmd/cgo/server.go:294-337` (`api.New(api.Deps{...})`).

### 1.4 Composition reality (why they are not standalone today)

All three packages are wired together only in `cmd/cgo` (`cmd/cgo/server.go:15-23` imports `api`, `doctor`, `iface` together; `cmd/cgo/main.go:68` calls `doctor.Mode`; `cmd/cgo/setup.go:126`, `217` call `doctor.Mode`/`doctor.Print`). `pkg/api` imports seven sibling packages plus `web/frontend` (`server.go:21-30`: `audit`, `campagne`, `figures`, `metrics`, `model`, `probe`, `profile`, `results`, `frontend`) and `probe/bpf` transitively via `kernel_bridge.go:5`. `pkg/doctor` imports only `model` (`doctor.go:13`). `pkg/iface` imports stdlib only (`iface.go:6-14`). Module facts: one `go.mod`, `go 1.25` (`go.mod:1-3`); under R2 a module is "released, versioned, and distributed together" and its path is its ID, so today `api`/`doctor`/`iface` cannot be versioned independently (R2 §Modules, §Module paths).

## 2. Exported API (file:line)

Complete inventory of exported identifiers in non-test sources. Unexported identifiers are summarized per file and are not API.

### 2.1 `pkg/api` — `server.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Param` | struct | `server.go:92-102` | Single-source schema registry element `{Key,Label,Min,Max,Default,Unit,Step,Enum,Desc}`; served by `GET /api/schema` (`server.go:413-415`) and read by server validators via `paramBounds` (`server.go:124-131`). 16 entries (`server.go:104-121`). |
| `SchemaParams() []Param` | func | `server.go:134` | Composition reads the `deadline_ms` default from the same registry (`cmd/cgo/server.go:76-88`). |
| `ShapeReq` | struct | `server.go:137-143` | POST `/api/shape` body `{qdisc, capacity_mbps, delay_ms, jitter_ms, loss_pct}`. |
| `RunOpts` | struct | `server.go:150-170` | POST `/api/run/start` body `{profiles, qdiscs, ccs, reps, deadline_ms, target, direction, run_id, flows, bench, channel_model}`. |
| `Deps` | struct | `server.go:173-216` | 14 fields: `GetSnap/NextSnap func() any`, `StartFn/StopFn/SkipFn`, `ShapeFn`, `RunningFn`, `WatchFn`, `Mode`, `Version`, `DoctorFn func() any`, `BurstFn`, `SinksFn func() audit.Sinks`, `AuditDepsFn func() audit.Deps`. `SinksFn`/`AuditDepsFn` leak `audit` types (see §6). |
| `func (d Deps) observeBlocked` | unexported method on exported type | `server.go:220-229` | Exact-`observe` gate → 501 JSON; any other `Mode` string = full. |
| `RecordEvent(kind, msg string)` | func | `server.go:233` | Exported wrapper over `recordEvent` (50-event ring, `server.go:262-269`); used by `cmd/cgo` quarantine/skip/alert callbacks (`cmd/cgo/server.go:91-103`). |
| `Handler` | struct | `server.go:237-240` | `{http.Handler; hubClose func()}`. |
| `func (h Handler) CloseHub()` | method | `server.go:243-247` | Idempotent hub-ticker stop; called on server shutdown (`cmd/cgo/server.go:386`). |
| `New(d Deps) Handler` | func | `server.go:380-1432` | Builds `Hub` + `ServeMux` (R8 `"METHOD /path"` patterns), 40+ routes listed in §1.3; wraps with `crossSiteGuard` (`server.go:1431`). Nil `GetSnap`→`{"running":false}`, nil `NextSnap`→`GetSnap` (`server.go:385-392`). |
| `maxSubs = 64` | const | `server.go:314` | SSE cap → 503 (`sse.go:134-138`). |
| `validRunID(s string) bool` | unexported func | `server.go:320` | Rejects `/`, `\`, `.` anywhere; shared by results/replay/run-rows/quarantine/run-start (`server.go:626`, `766`, `1104`, `1282`, `1314`). |
| `shapeApply(...)` | unexported func | `server.go:277-312` | Validates qdisc enum, capacity 1–1000, delay 0–600, jitter 0–100, loss 0–10 (from `schemaParams` bounds), runs `ShapeFn`, records event. |
| `readValidMetric(...)` | unexported func | `server.go:334-378` | Valid-only `small_p95_ms`/`deadline_ok_pct` by column name, direction-aware. |
| `auditParamsFromRequest(...)` | unexported func | `server.go:1488-1524` | `map[string]any`→`audit.Params`; `num` accepts `float64` or `int` (JSON vs. direct Go). |

Unexported remainder (not API): `schemaParams` var (`server.go:104`), `compareCache` (`server.go:326-330`), `shapeQdisc/Cap/Since` + `shapeMu` (`server.go:272-275`), `eventsRing` + `eventsMu` (`server.go:259-260`), `recordEvent` (`server.go:262`), `auditMu/lastAudit/lastAuditReport/auditRunning/auditCancel/auditStartedAt/auditTotalSec/auditProgress` (`server.go:249-256`), `paramBounds` (`server.go:124`), `truncateUTF8` (`server.go:1454-1462`), `writeJSON` (`server.go:1464`), `writeErr` (`server.go:1474`), `crossSiteGuard` (`server.go:1438`), `readAuditReport` (`server.go:44`), `auditReportExists` (`server.go:71`).

### 2.2 `pkg/api` — `sse.go` (Hub)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Hub` | struct | `sse.go:18-28` | `{mu, lastID, ring []frame (cap 2048), subs, last map, lastFull string, ticker, done, closeOnce}`. |
| `NewHub() *Hub` | func | `sse.go:41-44` | Constructs; no ticker until `Serve`. |
| `func (h *Hub) Serve(next func() any)` | method | `sse.go:47-63` | Starts `time.NewTicker(1s/frameHz)` (10 Hz, R7-agnostic cadence); second call no-ops; runs until `Close`. |
| `func (h *Hub) Close()` | method | `sse.go:67-69` | Idempotent (`sync.Once`); `Serve` never restarts after `Close`. |

Unexported: `frameHz=10` (`sse.go:14`), `frame`/`sub` types (`sse.go:30-39`), `toMap` (`sse.go:219`), `contains` (`sse.go:228`), `Publish`/`SSE` (unexported methods, `sse.go:74`, `sse.go:127` — see §2.2 note), `structuralKeys` (`sse.go:72`).

> Note — `Publish(snap any)` and `SSE(w, r)` are unexported methods on the exported `Hub` type (`sse.go:74`, `sse.go:127`). They are reachable only through `Hub` values obtained via `NewHub()`/`New()` within package tests (`sse_test.go:44`, `sse_hub_test.go:26`). An external module consumer cannot call `Publish`/`SSE` directly; the only externally usable hub surface is `api.New(...).CloseHub()` plus HTTP `GET /api/stream`. This is a primary packaging gap (§6.3).

### 2.3 `pkg/api` — `metrics.go`, `translate.go`, `auth.go`, `duel.go`, `server_audit.go`, `kernel_bridge.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `HandleTranslate(w, r)` | func | `translate.go:16-125` | `GET /api/hardware/translate?profile=` (default P2, `translate.go:17-20`); best-cell vs `pfifo_fast` interpretation + `cli` recipe + `next_steps`. |
| (none exported) | — | `metrics.go`, `auth.go`, `duel.go`, `server_audit.go`, `kernel_bridge.go` | `renderMetrics`/`handleMetrics` (`metrics.go:17`, `43`), `requireAuth` (`auth.go:11`), `handleDuel`/`qdiscRank`/`eligible`/`cellOf`/`ptr`/`round1` (`duel.go:46`, `58`, `62`, `122`, `132`, `136`), `handleAudit*`/`resolveArchivedReport`/`auditParamsFromRequest` (`server_audit.go:26`, `228`, `238`, `272`, `293`, `323`, `366`, `376`, `390`, `435`), `kernelTelemetryFromSummary` (`kernel_bridge.go:12`) are all unexported. `duelDir` var (`duel.go:17`), `duelProfile` regex + `duelMinReps=3` (`duel.go:19`, `24`), `duelCell`/`duelResponse` types (`duel.go:26`, `35`) are unexported. |

### 2.4 `pkg/doctor`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Check` | struct | `doctor.go:16-20` | `{Name, Status (ok\|warn\|fail), Detail}` — JSON shape for `GET /api/doctor` (`server.go:405-411`, `cmd/cgo/server.go:304-307`). |
| `Mode(override string) string` | func | `doctor.go:24-34` | `observe|full` passthrough; `""`/unknown → auto (Windows observe, else full). |
| `Report(mode string) (string, []Check)` | func | `doctor.go:37-65` | 7+ checks: `os`, `tc`, `cap_net_admin`, 2× `ccChecks`, `bench-offloads`, `ping`. No context/timeout (see §6.5). |
| `Print()` | func | `doctor.go:68-74` | CLI text; hardcodes `Mode("auto")` (ignores any operator `mode` already resolved). |
| `ParseOffloadFeatures(out string) map[string]bool` | func | `doctor.go:81-106` | Pure parser: state = first word after `:` (`on`→true, `off`→false), ignoring `[fixed]`/`[requested on]` suffixes; non-`on/off` states ignored. |

Unexported: `benchOffloadPairs` (6 pairs, `doctor.go:110-117`), `benchOffloadCheck` (`doctor.go:123`), `ccChecks`/`ccChecksFrom` (`doctor.go:152`, `160`).

### 2.5 `pkg/iface`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Counters` | struct | `iface.go:20-25` | `{RXBytes, RXDrops, TXBytes, TXDrops uint64}` — no packets (§1.1). |
| `Sample` | struct | `iface.go:30-41` | JSON `iface_name/iface_source/iface_rx_mbps/iface_tx_mbps/iface_rx_drops/iface_tx_drops/iface_window_ms/iface_sample_ts/iface_valid/iface_gap`. `Valid=false` invalidates all numerics. |
| `Reader` | interface | `iface.go:44-47` | `{Read() Sample; Latest() Sample}` — the future SNMP/RouterOS seam. |
| `NormalizeSource(s string) string` | func | `iface.go:51-57` | Total function to `twin|prod|unknown`. |
| `ReadCounters(path string) (map[string]Counters, error)` | func | `iface.go:62-100` | Strict: skips 2 header lines + blank lines; `LastIndexByte(':')` split; requires ≥16 fields; truncated line = error (not a skip). Field map f[0]/f[3]/f[8]/f[11] (§5.1). |
| `Sampler` | struct | `iface.go:109-117` | `{mu, path, name, source, now func(), prev *prev, last Sample}`; `prev`/`last` unexported. |
| `New(name, source string) *Sampler` | func | `iface.go:120-122` | Production: `/proc/net/dev` + `time.Now`. |
| `func (s *Sampler) Read() Sample` | method | `iface.go:134-179` | Advances base; gaps: read/absent/warmup/clock(≤0)/slow(>500 ms)/reset(counter decreased). Rate = `delta*8/1e6/dt` (decimal Mbps, no clamp/floor). |
| `func (s *Sampler) Latest() Sample` | method | `iface.go:182-186` | Never advances base; lock-guarded copy of `last`. |

Unexported: `procNetDev="/proc/net/dev"` (`iface.go:16`), `prev` (`iface.go:102-105`), `newSampler(path,name,source,now)` (`iface.go:125-129`), `gap`/`emit`/`round1` (`iface.go:188`, `195`, `200`).

## 3. Dependencies

### 3.1 Direct imports (evidence)

- `pkg/api/server.go:3-30`: stdlib (`context`, `encoding/csv`, `encoding/json`, `fmt`, `io`, `log`, `math`, `net/http`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `sync`, `sync/atomic`, `time`) + internal `audit`, `campagne`, `figures`, `metrics`, `model`, `probe`, `profile`, `results` + `web/frontend`. `pkg/api/sse.go:4-12`: stdlib only (`encoding/json`, `fmt`, `net/http`, `strconv`, `strings`, `sync`, `time`). `pkg/api/metrics.go:3-8`: stdlib only (`fmt`, `net/http`, `sort`, `strings`). `pkg/api/duel.go:3-13`: stdlib + `metrics`, `results`. `pkg/api/server_audit.go:3-24`: stdlib + `audit`, `metrics`, `model`, `profile`, `remediate`. `pkg/api/kernel_bridge.go:3-6`: `audit` + `probe/bpf`. `pkg/api/translate.go:3-9`: stdlib + `results`.
- `pkg/doctor/doctor.go:6-14`: stdlib (`fmt`, `os`, `os/exec`, `runtime`, `strings`) + `model` (only `model.ParseCCList`, `pkg/model/cc.go:8`).
- `pkg/iface/iface.go:6-14`: stdlib only (`fmt`, `math`, `os`, `strconv`, `strings`, `sync`, `time`).
- `go.mod:1-10`: `module github.com/Realms4239/cgo`, `go 1.25`, direct `require` only `fyne.io/fyne/v2`, `charmbracelet/bubbletea`, `lipgloss`, `golang.org/x/sys` (TUI/GUI — `pkg/api|doctor|iface` use none of them; their only externals are those four plus stdlib). Under R2, `go.mod` records module identity + dependency minima, and `go list`/`go mod tidy` resolve the build list via minimal version selection (R2 §`go.mod` file, §Minimal version selection, §Automatic updates).
- Process-global state (standalone-blockers): `auditMu/lastAudit/lastAuditReport/auditRunning/auditCancel/auditStartedAt/auditTotalSec/auditProgress` (`server.go:249-256`), `eventsRing/eventsMu` (`server.go:259-260`), `shapeQdisc/Cap/Since/shapeMu` (`server.go:272-275`), `compareCache` (`server.go:326-330`), `schemaParams` (`server.go:104`), `duelDir` (`duel.go:17`). Each `api.New` shares the same globals — two `New` instances race on audit lifecycle.
- Filesystem hardcoding (standalone-blockers): `"data/runs"` (`server.go:335`, `766`, `794`, `925`, `982`, `1004`, `1092`, `1110`, `1286`, `1331`, `1363`; `duel.go:17`), `"data/figures"` (`server.go:1085`, `1091`), `"data/link_audit.csv"` (`server_audit.go:184`, `393`, `437`), `audit.DefaultDataDir` artifact paths (`server.go:45`, `75`). `pkg/doctor` hardcodes `"/proc/sys/net/ipv4/tcp_available_congestion_control"` (`doctor.go:153`), `exec.LookPath("tc"/"ping")` (`doctor.go:41`, `59`), `exec "ip netns list"` (`doctor.go:127`), `exec "sudo -n ip netns exec … ethtool -k"` (`doctor.go:132-133`), and 6 bench pairs (`doctor.go:110-117`). `pkg/iface` hardcodes `/proc/net/dev` in `New` but allows injection via `newSampler(path, ..., now func())` (`iface.go:125`) — the only already-seamed file path.

### 3.2 Toolchain / host dependency (doctor + iface)

- `tc` presence distinguishes `warn` vs `fail` by mode (`doctor.go:44-48`); underlying `tc(8)` shaping behaviour is unverified — primary source not found.
- `ethtool -k` gating assumes `tcp-segmentation-offload`/`generic-segmentation-offload`/`generic-receive-offload` names match R6 `-K` built-ins (`doctor.go:139`). R6 lists `tso/gso/gro` as `-K` feature names for `ethtool -K`; the `-k` output names in `doctor.go:139` are the long forms (`tcp-segmentation-offload` etc. as in `offload_test.go:9-11`). Equivalence of short/long names is unverified — primary source not found (R6 documents `tso/gso/gro` as `-K` names, output spelling unverified).
- `ping` presence → `ok`/`fail` (`doctor.go:59-63`); audit usability beyond presence is unverified — primary source not found (no round-trip, no SOCKS/proxy awareness).
- `cap_net_admin` via `os.Geteuid()==0` (`doctor.go:52`); capabilities(7) semantics, `setcap` helpers suggested in the warn string (`doctor.go:55`), and file capabilities (`cap_net_admin+ep`) are unverified — primary source not found.

## 4. Primary-source requirements

Minimum external authorities any split-module release must re-verify (with the exact claim each governs):

| ID | Authority | Claim it governs | Repo binding |
|---|---|---|---|
| R1 | Prometheus exposition formats | `Content-Type: text/plain; version=0.0.4`; HELP/TYPE cardinality + ordering; `counter/gauge` vocabulary; sample EBNF; `\n`-terminated last line; grouping | `metrics.go:25`, `30`, `45`; `metrics_test.go:18-27` |
| R2 | Go Modules Reference | What counts as a releasable module; module-path lexical rules; `/vN` suffix; `go` directive = mandatory minimum fd Go 1.21; `require`/`replace` semantics | `go.mod:1-3`; split plan §7 |
| R3 | SemVer 2.0.0 | When a split is MAJOR/MINOR/PATCH; MUST declare public API; released-version immutability | version policy §7 |
| R4 | `proc_pid_net(5)` | `/proc/net/dev` 16-field layout (rx f[0..7], tx f[8..15]); 2-line header + per-device lines | `iface.go:62-100` (f[0]/f[3]/f[8]/f[11]) |
| R5 | Linux IP Sysctl (docs.kernel.org) | `tcp_available_congestion_control` = registered-choices STRING; subset rule for allowed; kernel-module presence affects listing ("More … may be available as modules, but not loaded") | `doctor.go:153-158`; `model.ParseCCList` (`pkg/model/cc.go:8-18`) |
| R6 | `ethtool(8)` | `-k` = query offload state; `-K` built-ins `tso/gso/gro` exist; check must read real `ethtool -k` output, not guess | `doctor.go:132-139`; `offload_test.go:5-15` |
| R7 | WHATWG Server-sent events | `text/event-stream`; fields `data/event/id/retry`/comments; blank-line dispatch; `Last-Event-ID` on reconnect; `retry` = digits-only reconnection time | `sse.go:112`, `119`, `139-142`, `147-167` |
| R8 | Go `net/http` (pkg.go.dev) | `"METHOD /path"` mux patterns; `MaxBytesReader` cap semantics; `Flusher` streaming; status-code vocabulary | `server.go:380-1432` route table; `1<<20` caps; `sse.go:128-132` |

Fetch-or-it's-unverified rule: any claim about `tc(8)` netem/cake, `ip-netns` bench topology, capabilities(7), BBR vs BBRv3 mechanics, `Sec-Fetch-Site` semantics, Bearer/RFC 6750, RouterOS/SNMP adapters — label **unverified — primary source not found** until the authoritative man page / RFC / spec is fetched and quoted. Today's docs (`docs/api.md`) assert some of these without external fetch; §6 marks each.

## 5. Conformity details (spot checks against primary sources)

### 5.1 `iface` — field indices match R4 exactly

R4 header: `|bytes packets errs drop fifo frame compressed multicast|bytes packets errs drop fifo colls carrier compressed` → rx indices 0–7 (0 bytes, 1 packets, 2 errs, 3 drop, 4 fifo, 5 frame, 6 compressed, 7 multicast), tx indices 8–15 (8 bytes, 9 packets, 10 errs, 11 drop, 12 fifo, 13 colls, 14 carrier, 15 compressed). `iface.go:81-96` reads f[0]=rx, f[3]=rx_drops, f[8]=tx, f[11]=tx_drops — conformant (R4). Parser strictness (short line = error, `iface.go:78-80`; missing colon = error, `iface.go:73-75`; missing file = error, `iface.go:63-66`) matches the honesty contract "a truncated line would produce a false measurement" (`iface.go:60-61`) and is covered by `iface_test.go:42-53` (short-line rejection, missing-file rejection) and `sampler_test.go` warmup/reset/slow/clock/absent/read/lastest cases. Colon-in-name handling (`LastIndexByte`, `iface.go:72`, tested `eth:0` in `iface_test.go:37-39`) is a defensible superset; whether kernel iface names can contain `:` is unverified — primary source not found (IFNAMSIZ allowance not fetched).

Rate semantics: `bits*8/1e6/dt` (`iface.go:169-170`) is decimal Mbps with no floor/clamp — matches comments (`iface.go:131-133`) and the `sampler_test.go:49-65` expectation (125000 B / 100 ms = 10 Mbps, 25000 B / 100 ms = 2 Mbps). `slow` at `dt>500ms` resets the base and reports `gap:"slow"` rather than a quarter-second average (`iface.go:160-163`); `reset` on counter decrease (`iface.go:164-167`); `clock` on `dt≤0` (`iface.go:156-159`); `warmup` on first sight (`iface.go:149-152`); `absent`/`read` clear `prev` (`iface.go:140-148`). All six gaps are the full R7-agnostic vocabulary; `docs/api.md:32-34` lists the same enum. `Latest()` never advances (`iface.go:182-186`) so `/api/state` (`server.go:527-529` via `cmd/cgo/server.go:67-69`) is side-effect-free while the hub's `Read()` (`cmd/cgo/server.go:70-72`) is advancing — matches `docs/api.md:35-36`.

### 5.2 `doctor` — CC checks match R5; offload gating matches R6 surface

R5: `tcp_available_congestion_control` is a STRING listing registered choices; more may exist as unloaded modules. `doctor.go:152-158` reads the file, delegates tokenization to `ParseCCList` (dedupe + order-preserving, `pkg/model/cc.go:8-18`), and `ccChecksFrom` (`doctor.go:160-183`) does exact-token membership (`present["bbr"]` / `present["bbr3"]`), never substring — the shipped fix for the "contains(`bbr`) declared bbr on a bbr3-only kernel" bug noted in `doctor.go:149-151`. Status semantics: readable+`bbr`→`ok`, readable+`bbr3`→`ok`, missing `bbr3`→`warn` (optional), missing `bbr`→`warn` (observe) or `fail` (full), unreadable→2×`warn` — covered by `cc_test.go:5-26`. The "absent from list ⇒ fail in full" rule is a policy choice, not a kernel requirement; R5 does not state that a listed algorithm must be loaded — it explicitly says unloaded-module algorithms are absent from the list (R5 quote above). So "absent ⇒ bench falsified" is unverified — primary source not found as a kernel rule; as a bench-integrity policy it is documented in `doctor.go:119-122`.

Offload gating: R6 `-k` "Queries the specified network device for the state of protocol offload and other features"; `-K` built-ins include `tso` (TCP segmentation offload), `gso` (generic segmentation offload), `gro` (generic receive offload). `doctor.go:132-139` runs `ethtool -k` per bench interface inside `ip netns exec` and fails if any of `tcp-segmentation-offload`/`generic-segmentation-offload`/`generic-receive-offload` is on — the gate direction (any-on ⇒ fail) is intentional bench honesty (`doctor.go:119-122` "a bench that lies falsifies every measurement"). Parser tolerance for `[fixed]`/`[requested on]` suffixes is the first-word cast (`doctor.go:94-97`); header line `Features for c-cli:` is skipped only because its post-colon state is empty (`doctor.go:89-93` `continue` on `state==""`), not by explicit header filtering — fragile to future ethtool output changes (R6 output format changes are unverified — primary source not found beyond the `-k/-K` option text). `offload_test.go:17-44` covers on/off/fixed/empty/header cases; `offload_test.go:46-66` requires the `bench-offloads` check to exist in every mode (never silent).

Honesty limits: `benchOffloadCheck` returns `warn` when not on Linux, when `ip netns list` is empty/fails, or when `ethtool -k` is unreadable (`doctor.go:124-137`) — i.e. "bench absent ⇒ warn", not `ok`. `Report` has no timeout context on any `exec` (`doctor.go:41`, `59`, `127`, `132` use bare `exec.LookPath`/`exec.Command.Output()`), so a hung `ip`/`ethtool` blocks `GET /api/doctor` indefinitely — gap §6.5. The six bench pairs (`doctor.go:110-117`) are hardcoded from `kit/cgo-bench-dualns.sh` (cited `doctor.go:108-109`); bench topology changes silently drift the gate (unverified — primary source not found; no manifest consumed).

### 5.3 `api` transport — ServeMux/JSON/SSE conform to R8/R7/R1 surfaces

- Mux patterns: `mux.HandleFunc("GET /api/health", …)`, `mux.HandleFunc("POST /api/run/start", …)`, etc. (`server.go:394-1432`) use R8 method+path patterns. The `/api/` catch-all 404 JSON (`server.go:1269-1273`) precedes the SPA `mux.Handle("/", spa)` (`server.go:1430`) so unknown `/api/*` never falls through to `FileServer` (R8 `FileServer`/`StripPrefix`, `server.go:1278`, `1091`). SPA fallback serves embedded `web/frontend.FS` (`server.go:1278`).
- `MaxBytesReader` 1 MiB caps: every JSON decode path in `server.go`/`server_audit.go` wraps `r.Body` with `http.MaxBytesReader(w, r.Body, 1<<20)` (`server.go:459`, `487`, `578`, `601`, `619`, `1159`; `server_audit.go:45`; R8 `MaxBytesReader(w, r, n)`). The CSV profile import also reads bounded (`server.go:1159`). `docs/api.md:4` correctly states "Tout corps JSON est limité à 1 Mio".
- `writeJSON`/`writeErr` negotiation: `writeJSON` always sets `Content-Type: application/json` and encodes (`server.go:1464-1470`, R8 status/header semantics); `writeErr` returns JSON only when `Accept` or `Content-Type` contains `application/json`, else `text/plain` via `http.Error` (`server.go:1474-1483`, as `docs/api.md` implies for curl vs. fetch). Shortcoming: `writeJSON` logs encode errors (`server.go:1467-1469`) but cannot change the already-200 status/partial body — a `NaN`/`Inf` value yields a truncated 200, not an error response (the `metrics`/`results` NaN history cited in the comment `server.go:1466`). JSON float non-finite handling is a Go stdlib property (unverified — primary source not found in this pass beyond the in-code comment).
- SSE wire (R7): `Content-Type: text/event-stream`, `Cache-Control: no-cache`, no manual `Connection` header (hop-by-hop, HTTP/2-illegal, `sse.go:139-142`, `server.go:1124-1126`), `retry: 2000` first line (`sse.go:142`; R7 `retry` must be ASCII digits — digits-only, conformant), frames `id: <uint>` + `data: <json>` + blank line (`sse.go:112`, `143`, `177`; R7 `id` must not contain NUL/LF/CR — numeric IDs conformant), named `event: backpressure` (`sse.go:118-121`; R7 event-type buffer is free-form), `Last-Event-ID` read only if numeric, else treated as fresh client (`sse.go:148-167`; R7 `Last-Event-ID` is an opaque UTF-8 string — numeric-only parsing is stricter than the spec but safe), initial replay capped at last 5 frames when no header (`sse.go:171-173`), stale ID (pre-ring) → full-frame resync (`sse.go:161-163`, `198-200`). The 5-frame cap + `lastFull` resync is an application-level freshness protocol layered on top of R7 primitives; R7 itself defines no delta compression (`structuralKeys` delta omission, `sse.go:72-91`) — clients must implement the delta merge per the app contract (`sse_test.go:69`, `80-82` asserts `profile` appears exactly once across the stream). No `:`-comment heartbeat is sent on idle connections; R7's authoring note recommends a comment every ~15 s against proxy timeouts (R7 §9.2.7) — `sse.go` sends only the initial `retry: 2000`; idle-stream proxy-drop behaviour is unverified — primary source not found in-repo (no keepalive code), and R7 buffer/keepalive guidance is non-normative.
- Subscriber accounting: `len(h.subs) >= 64` → `503` (`sse.go:134-138`, `maxSubs` `server.go:314`); `503 Service Unavailable` for overload is application policy, R7 defines no status for this; flusher absence → `500` (`sse.go:128-132`), replay flusher absence → `500` (`server.go:1119-1123`). Status-code vocabulary (`400/401/403/404/409/422/500/501/503`) aligns with R8 constants but the per-route mapping is app policy (`docs/api.md` table) — e.g. `422` for audit-without-measured-capacity (`server_audit.go:489`), `501` for observe-mode control verbs (`server.go:224-228`), `409` for campaign-running conflicts (`server.go:280`, `466`, `513`, `587`), `403` for cross-site browser mutations (`server.go:1444`).

### 5.4 `api` `/metrics` — exposition-valid, semantically empty (toy)

`renderMetrics` (`metrics.go:17-41`) emits `# HELP <name> <help>` + `# TYPE <name> gauge|counter` (`_total` suffix → `counter`, `metrics.go:26-30`) + `name{labels} value` lines with sorted names and sorted label pairs — conformant with R1 line ordering ("HELP and TYPE lines first (in no particular order)", "reproducible sorting preferred"), one HELP/TYPE per metric, TYPE-before-sample, `value` as float string. `handleMetrics` sets `Content-Type: text/plain; version=0.0.4` (`metrics.go:45`; R1 requires exactly this value). `metrics_test.go:10-28` asserts HELP/TYPE/Content-Type presence; `metrics_test.go:30-43` asserts an absent sample is not emitted (empty `metricSample` slice → no `cgo_jfi_pct{` line) while the `cgo_jfi_measured=0` flag still appears.

The endpoint body (`metrics.go:46-53`) is six hard-coded samples: `cgo_audit_runs_total{status=completed} 0`, `cgo_audit_load_attempted_total{site=unknown} 0`, `cgo_audit_load_measured_total{site=unknown} 0`, `cgo_cap_sla_measured{site=unknown} 0`, `cgo_iface_valid{site=unknown} 0`, `cgo_jfi_measured{site=unknown} 0`. `help` strings (`metrics.go:54-61`) are present but `handleMetrics(getBuildInfo)` ignores `getBuildInfo` entirely (`metrics.go:43-63` — parameter never read). No counter ever increments; no audit/iface state is read; `site` is always `"unknown"`. R1 itself warns "Types and docstrings not integral part of the syntax, meaning little-to-nonexistent metric contract validation" — so a scraper will accept these zeros, but the values are fabricated, not measured. The project's own absent≠zero invariant (stated in `renderMetrics` comment `metrics.go:15-16` "Absent is never emitted; the caller only pushes measured values" and `metrics_test.go:38` "absent JFI must not emit a zero sample") is violated at the endpoint level: absent measurements are emitted as zeros, and `cgo_jfi_measured` help text literally says "0 means absent, never a zero value" (`metrics.go:60`) while the endpoint emits `0`. See §6.1/§6.6 for the full gap.

## 6. Gap analysis (what blocks standalone release)

Ordered by severity. Each gap cites the code and the primary source it would violate or the policy it breaks.

### 6.1 `/metrics` is a static-zero facade (BLOCKER)

Observed `metrics.go:46-63`: six zero samples, `site=unknown` constant, `getBuildInfo` ignored, no wiring to `auditRunning`/`lastAudit`/`cgo_iface_valid` (the sampler exists only in `cmd/cgo`, never passed to `api`), no `cgo_build_info` (the ignored callback's evident intent). A standalone observability module cannot ship zeros as measurements. Under R3 §1 (MUST declare a precise public API), the declared API includes metrics that never change — false precision. Test `metrics_test.go:30-43` proves the *builder* respects absent≠zero but cannot detect the endpoint's fabrication (it tests `renderMetrics`, not `handleMetrics` state). Prior research already noted this: `docs/research/2026-11-06-enterprise-standard-cgo-platform.md:64` ("There is no /metrics Prometheus endpoint" — now present but still static; the `go.mod` grep there found no prometheus deps, still true: no prometheus client in `go.mod:5-10`, `metrics.go` hand-rolls exposition).

### 6.2 Shared globals + CWD-relative paths (BLOCKER)

`api.New` creates a fresh `Hub` per call (`server.go:381`) but audits, events, shape state, compare cache, and schema are package globals (§3.1). Two servers in one process share `lastAuditReport` (cleared on every audit start, `server_audit.go:130`), `eventsRing` (50-cap, `server.go:266-268`), `shapeQdisc` (`server.go:303`), and `compareCache` (`server.go:919-921`). Standalone, an embedder cannot run two ledgers, nor point at a non-CWD data dir — every read/write is hardcoded `"data/…"` (§3.1). `duelDir` is test-mutable (`duel.go:17`) but all other paths are constants. For a library this is a testability and multi-tenant failure (failure mode: embedder in `/tmp` writes into caller's CWD; second `New` observes first `New`'s audit). R2 module-release readiness requires a parameterized store or FS root — absent.

### 6.3 Hub exportedness mismatch (MAJOR)

`Hub`/`NewHub`/`Serve`/`Close` are exported (`sse.go:18`, `41`, `47`, `67`) but `Publish`/`SSE` — the only methods that give `Hub` meaning — are unexported (`sse.go:74`, `127`). An external consumer can create a hub, start a ticker that publishes `nil` from a supplied `next`, and close it, but cannot feed it, cannot serve it, and cannot inspect `lastID`/`ring`. The intended external surface is opaque: `GET /api/stream` over a full `New` server is the only path. Either export or shrink to unexported; shipping a half-exported type violates R3 §1 (public API must be precise — a caller discovering `Hub.Serve` cannot complete the lifecycle). The same pattern recurs for `Handler` (`server.go:237`) — exported, but its `hubClose` field makes it unconstructable outside `New`, with one method (`CloseHub`) — acceptable but undocumented as "obtain only via New".

### 6.4 Untyped `any` snapshots + `Deps` leakage (MAJOR)

`Deps.GetSnap/NextSnap func() any` (`server.go:174-177`), `DoctorFn func() any` (`server.go:203`), `SinksFn func() audit.Sinks`, `AuditDepsFn func() audit.Deps` (`server.go:211-215`). Consequences: (a) no compile-time snapshot contract — `Hub.Publish` does `toMap` via `json.Marshal`/`Unmarshal` round-trip (`sse.go:219-226`) and silently drops marshal failures (`sse.go:75-77`), so a type error becomes a silent 10 Hz gap; (b) `api` cannot be built without `audit` (and transitively `results`, `metrics`, etc.) — `Deps` recompilation coupling; (c) `Mode`/`Version` are unchecked strings (`server.go:199-201`; `observeBlocked` treats anything ≠`"observe"` as full, `server.go:221`), so `"Observe"`/`"obs"` silently enables control verbs. A standalone module needs typed `Snapshot`/`Sinks`/`AuditDeps` interfaces owned by `api` (or a `core` contract package), with `Mode` as a validated enum.

### 6.5 Unbounded / platform-fragile checks + auth edge cases (MAJOR)

- `Report` blocks on bare `exec` (`doctor.go:41`, `59`, `127`, `132-133`) with no `context.Context`/timeout; `GET /api/doctor` (`server.go:405-411`) can hang behind `sudo -n` prompts (non-interactive auth failure → warn, but a `sudo` timestamp-refresh hang is unverified — primary source not found). `sudo -n` inside a library is also a deployment assumption (requires NOPASSWD or cached creds) with no fallback documented.
- `requireAuth` (`auth.go:11-26`): loopback check is `strings.HasPrefix(r.RemoteAddr, "127.0.0.1")` (`auth.go:19`) — misses IPv6 loopback `::1` (remote addr `[::1]:port`), so an empty-token dev server on IPv6 loopback denies the legitimate dev caller with `401` (observed, not primary-sourced; TCP remote-addr format is a Go stdlib property — unverified — primary source not found in this pass beyond R8 `net/http` package scope). Token compare is `got == want` (`auth.go:15`), not constant-time (timing side channel; constant-time-compare guidance unverified — primary source not found). Always sets `WWW-Authenticate: Bearer` before `writeErr` (`auth.go:23-24`); `writeErr` then sets `Content-Type` — order is safe in Go (header map before write), but if `writeErr` is later changed to stream, auth challenge could be lost.
- `crossSiteGuard` (`server.go:1438-1450`) checks only `Sec-Fetch-Site: cross-site` on non-safe methods; `same-site` browser mutations from a malicious sibling subdomain are allowed, and non-browser agents (no header) bypass — by design per `docs/api.md:5-7`, but `Sec-Fetch-Site` semantics per se are unverified — primary source not found (Fetch Metadata spec not fetched).
- `build` env defaults in `api` leak host env into a library: `os.Getenv("CGO_SMALL_URL")` (`server_audit.go:103`), `os.Getenv("CGO_AUDIT_IFACE/ROUTEROS")` (`server_audit.go:109-113`), `audit.EnvOrBulk` (`server_audit.go:106`). A library reading the embedder's environment violates least-surprise; env must move to `Deps` or composition.
- `Print()` hardcodes `Mode("auto")` (`doctor.go:69`) even though `Report(mode)` takes a mode; CLI `--mode full` on Windows would still print observe — divergence between CLI truth and server truth.

### 6.6 Smaller gaps (MINOR / polish)

- `writeJSON` partial-body-on-encode-error (§5.3) — silent 200 with truncated JSON; should pre-marshal or buffer, or set 500 before writing.
- `figure`/`results` reads tolerate missing dirs silently in some branches (e.g. `results.Scan` ignored errors `server.go:770`, `1004`) — empty groups masquerade as "no data" vs. "corrupt data".
- `validRunID` forbids `.` everywhere (`server.go:320`), so run IDs containing dots (e.g. `run-2026.11.07`) are rejected even when safe; and allows empty strings for "all runs" (documented, `server.go:317-319`) — the two meanings of empty (all vs. missing) are documented per-route (`docs/api.md:69`).
- `handleAuditToProfile` hard-fails `422` when capacity/RTT are unmeasured (`server_audit.go:489`) — correct honesty, but the error is `UnprocessableEntity` while neighboring "no audit" is `404` (`server_audit.go:439`); client must distinguish two different "cannot import" cases by status+message.
- `iface.Sample` JSON tags (`iface.go:31-40`) double the snapshot size when embedded in every 10 Hz frame; no `omitempty` on `Valid=false` gaps — delta hub re-sends gaps too (`sse.go:83-91` structural compression skips only profile/qdisc/cc/repetition/event_id/phase).
- `renderMetrics` escapes nothing: label values containing `\`, `"`, `\n` would break R1 `escaped_string` rules (`metrics.go:33-36` interpolates raw). Only today's constant labels save it; any future `site` label from user input (e.g. `server_audit.go:59` site string interpolated with `truncateUTF8`) would need R1 escaping.
- Renaming: `Rename` discipline — the repo module is `github.com/Realms4239/cgo` (`go.mod:1`); a third-party fork shipping `cgo/api/doctor/iface` as its own module would collide on this path (R2 module-path = canonical name; collision requires a distinct path or `/vN` per R2 §Module paths / §Major version suffixes).

## 7. Target module API (proposed)

Design spine (R3 §1 MUST declare a precise public API; R2 modules released/versioned together; `/v2` suffix when breaking): keep the current import path `github.com/Realms4239/cgo`, split into three versioned modules, one SHA-pinned `go.mod` each, `go 1.25` floor from `go.mod:3`.

### 7.1 `cgo/iface` — standalone module (no internal deps)

Proposed `github.com/Realms4239/cgo/iface` (v0.1.0, MIT per repo `LICENSE` — verified in repo root; R2 requires LICENSE in the initial commit):

```go
package iface

// Keep (already conformant):
type Counters struct { RXBytes, RXDrops, TXBytes, TXDrops uint64 }       // iface.go:20
type Sample struct { Name, Source string; RXMbps, TXMbps float64; ... } // iface.go:30

// Export the seams the hub needs:
type Reader interface { Read() Sample; Latest() Sample }                  // iface.go:44 — now exported-useful
type Clock func() time.Time
type Opener func() (map[string]Counters, error) // replaces hardcoded file read for tests/embedders

// Fix the ctor seam (currently only unexported newSampler, iface.go:125):
func NewSampler(op Opener, name, source string, clk Clock) *Sampler      // new exported injectable ctor
func New(name, source string) *Sampler                                    // keeps /proc/net/dev production default (iface.go:120)
func ReadCounters(op Opener) (map[string]Counters, error)                  // or ReadCounters(path) kept + documented — external path is embedder-chosen
func NormalizeSource(s string) string                                      // iface.go:51
func (s *Sampler) Read() Sample
func (s *Sampler) Latest() Sample
```

Rationale: (a) stdlib-only today — trivially releasable; (b) externalize `procNetDev` (`iface.go:16`) for `newSampler` parity; (c) keep `Gap` string enum but document it as a string type (`type Gap string`) so `warmup/absent/reset/slow/read/clock` is compiler-checked. Semver: PATCH for gap-string addition (backward-compatible), MINOR for `Opener`/`NewSampler` (new functionality, R3 §7), MAJOR/`/v2` only if `Sample` fields change (R3 §8, R2 §Major version suffixes).

### 7.2 `cgo/doctor` — standalone module (decouple `model`)

Proposed `github.com/Realms4239/cgo/doctor` (v0.1.0). The only internal dep is `model.ParseCCList` (`doctor.go:13`, `pkg/model/cc.go:8`); absorb it:

```go
package doctor

type Check struct { Name, Status, Detail string }                          // doctor.go:16
type Mode string; const (Observe Mode = "observe"; Full Mode = "full")     // replaces stringly Mode (doctor.go:24)
type Report struct { Mode string; Checks []Check }                           // instead of (string, []Check) tuple (doctor.go:37)

type Runner interface { // single seam for all exec (replaces 4 bare exec sites, doctor.go:41,59,127,132)
    LookPath(file string) (string, error)
    Output(ctx context.Context, name string, args ...string) ([]byte, error)
}
type FileReader func(path string) ([]byte, error)                           // for tcp_available_congestion_control (doctor.go:153)

func Mode(override string) string                                           // keep; document "" = auto (doctor.go:28-33)
func Report(ctx context.Context, mode string, r Runner, fr FileReader, pairs []BenchPair) ([]Check, error)
func ParseOffloadFeatures(out string) map[string]bool                        // doctor.go:81 (pure, keep)
func Print(ctx context.Context, mode string)                                  // fix hardcoded auto (doctor.go:69)
```

`BenchPair{Ns, Iface}` exported so `kit/` can inject topology instead of the hardcoded 6 (`doctor.go:110-117`). All `exec` flows through `Runner` with `ctx` + deadline (fixes §6.5 hang). `ccChecks`/`benchOffloadCheck` become unexported methods on a `Checker` struct (not package funcs), so tests inject `Runner`/`FileReader` without `sudo`/`root`. Keep `ParseOffloadFeatures` exact semantics (R6) — PATCH-eligible only if header tolerance changes.

### 7.3 `cgo/api` — transport module (store interfaces, typed snapshots)

Proposed `github.com/Realms4239/cgo/api` (v0.1.0). Two sub-packages to keep the hub/metrics independently importable (R2: "keep its functionality focused and discrete" — Go's module guidance cited in R2 summary):

```
github.com/Realms4239/cgo/api      — HTTP router, writeJSON/writeErr, auth, crossSite, schema, audit lifecycle (server.go)
github.com/Realms4239/cgo/api/sse  — Hub + SSE (sse.go), Publish/SSE exported
github.com/Realms4239/cgo/api/metrics — exposition builder + wired collectors (metrics.go)
```

Target API:

```go
package api

// Typed contract (replaces func() any, server.go:174-177,203):
type Snapshot struct { /* iface.Sample-compatible base; structural keys explicit */ }
type SnapshotProvider interface { Latest() Snapshot; Next() Snapshot }
type Engine interface {
    StartCampaign(ctx context.Context, o RunOpts) error
    StopCampaign(); SkipCell()
    ApplyShape(r ShapeReq) error; Running() bool
    Watch(on bool) error; Burst(ctx context.Context, cc string, secs int) error
}
type DoctorSource interface { Report(mode string) []doctor.Check } // server.go:203 typed
type AuditSink interface { /* audit start/stop/report verbs */ }   // replaces audit.* leakage (server.go:211-215)
type Store interface { // replaces "data/…" CWD reads (server.go:335)
    Runs() RunIndex; RunRows(runID string) ([]Row, error)
    FiguresRoot() string; AuditCSV() string; ArtifactRoot() string
}
type ServerOpts struct {
    Engine Engine; Store Store; Doctor DoctorSource; Audit AuditSink
    Snapshot SnapshotProvider; Mode, Version string; FS fs.FS // SPA root
    Auth AuthConfig; Metrics metrics.Registry               // §6.1 wired metrics
}
func New(opts ServerOpts) *Server   // *Server owns Hub (CloseHub), per-instance state (no globals, §6.2)
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (s *Server) Close() error
```

Plus exported `sse.NewHub`/`Hub.Publish`/`Hub.SSE`/`Hub.Close` (fix §6.3) and `metrics.NewRegistry`/`Registry.Observe(name, labels, value)` with real collectors bound to `Engine`/`SnapshotProvider` state (fix §6.1 — no static zeros; a metric with no collector is never emitted, per R1 grouping and `metrics.go:15-16` intent). `writeJSON`/`writeErr` become unexported methods on `*Server` with `Err` type carrying code+message, pre-marshaling so encode failures surface as 500 not truncated 200 (§6.6). `Mode` becomes `type Mode string` with `Observe/Full` consts + `ParseMode(string)(Mode, error)` (fix §6.4). `MaxBytes` (1 MiB, R8) becomes an opt with the same default.

Version policy (R3): v0.1.0 ships the transport + hub + real `/metrics`; any breaking `Snapshot`/`RunOpts` change → MINOR if fields only added (R3 §7), MAJOR + `/v2` (R2 §Major version suffixes) if fields removed/renamed or `validRunID` rule tightens. Retraction via `retract` directive (R2 §`retract`) for any premature tag.

## 8. Slices (proposed sequence)

Smallest-standalone-first (each slice releasable, green gates; in-repo gate is `make verify` per `AGENTS.md` — the *one command* `make verify`/`verify-fast`/`verify-selftest`; `verify` exit 0 = green, 1 = named gate failed, 2 = gate could not run — not green. Gate command quoted here from the harness/AGENTS note, not primary-fetched).

| Slice | Scope | Contents (files) | Acceptance (anti-toy probe) |
|---|---|---|---|
| S1 | `iface` module | `iface.go` + `iface_test.go` + `sampler_test.go` → new `go.mod` (`module …/iface`, `go 1.25`); export `NewSampler`/`Opener`; keep `New`/`ReadCounters`/`NormalizeSource` | `go test ./...` in module root passes without `cmd/` present; `ReadCounters` exercised against temp fixture (existing `iface_test.go:25-53` proves independence from `/proc`) |
| S2 | `doctor` decouple | Absorb `ParseCCList`; `Runner`/`FileReader` seams; `Mode` enum; `Report(ctx,…)`; `BenchPair` injection | `cc_test.go:5-26` + `offload_test.go:17-66` pass with fake `Runner` only (no `tc`/`ping`/`sudo`/`ip netns` on the test host); `Report` returns within 2 s when `Runner.Output` blocks until `ctx` cancel |
| S3 | `sse` sub-package | `sse.go` → `api/sse` (`NewHub/Publish/SSE/Close/Serve` exported); `delta` merge documented as app protocol | External test imports only `api/sse`, starts server, connects two `Last-Event-ID` clients, one stale (pre-ring ID) → asserts full-frame resync; no `pkg/api` import |
| S4 | `api` transport core | `writeJSON/writeErr`, `ServeMux` (`server.go:380-1432`), `auth.go`, `crossSiteGuard`, `validRunID`, `schemaParams`; `*Server` with per-instance state; `Engine/Store` interfaces; no `data/…` reads | `go test` runs in a temp CWD, asserts no file is created outside the injected `Store`; two `New` instances in one process do not share audit/event state (fails before slice, passes after) |
| S5 | Handlers decoupling | results/duel/compare/integrity/report/figures/replay/quarantine via `Store` (currently `results.Scan`/`figures.*` direct, `server.go:334`, `779`, `924`, `977`, `1085`); audit handlers via `AuditSink` (`server_audit.go`); `translate`/`duel` defaults injected | Suite under `-run TestStore` with stub store returns real `groups`/`Best` from stub, no `data/runs` access (assert via cwd empty) |
| S6 | Real `/metrics` | `metrics.go` collectors bound to `Engine`+`SnapshotProvider`; delete `handleMetrics` zeros (`metrics.go:46-61`); R1 escaping for label values | `GET /metrics` after a stub campaign emits `cgo_audit_runs_total … 1` and `cgo_iface_valid` from the live sampler; before any campaign, scrape emits NO `cgo_audit_runs_total` sample (absent, not zero) — fails on today's static zeros |
| S7 | Release train | Tag `iface@v0.1.0`, `doctor@v0.1.0`, `api@v0.1.0` (+`api/sse`, `api/metrics`); `LICENSE`/`go.mod` per R2; `/api/doctor`, `/api/stream`, `iface_*` field doc bump (`docs/api.md:17-36`); `retract` slot reserved | `GOPROXY=off go list -m all` per module resolves; `v0.1.0` binaries build from the tags alone (no `git` HEAD); `make verify` + `verify-selftest` green |

Slice notes: S1 is zero-dependency and cannot break other slices. S2 removes `model` from `doctor`'s import list (currently the only internal dep, `doctor.go:13`). S3–S5 must land together with the `Deps`→interfaces refactor (§6.4) or the split simply re-exports the coupling. S6 is the anti-toy hinge — it replaces the only fabricated output in these three packages. S7 applies R2 (`module`/`go`/`require`/`exclude`/`retract` full reference) + R3 (immutability, MINOR/MAJOR discipline) — no tag is moved, ever.

## 9. Anti-toy criteria (ship-blockers for any split release)

A split may not be tagged until all of these hold. Each criterion names the failing today behavior where applicable.

1. **No fabricated measurements.** Every exported counter/gauge is bound to a collector reading live or frozen evidence. A scrape of `/metrics` with no audits and no sampler emits no `cgo_audit_runs_total` sample at all (absent ≠ zero, `metrics.go:15-16`; R1 grouping/ordering). Today's static `0` (`metrics.go:47-52`, `site=unknown`) fails this criterion. (`metrics_test.go:38` already asserts the builder side of this rule.)
2. **No CWD-relative persistence.** No handler opens a path not obtained from injected `Store`/`FS`. Test runs in an empty temp CWD and asserts no `data/` directory appears. Today's ~15 hardcoded `"data/…"` call sites (§3.1) fail this criterion.
3. **No process-global mutable behavior.** Two `api.New(opts)` instances in one process do not share audit lifecycle, event ring, shape state, or compare cache. Today's `auditMu`/`eventsRing`/`shapeQdisc`/`compareCache` globals (`server.go:249-330`) fail this criterion. (`dup: sse_hub_test.go` pattern — construct two hubs in one process — must succeed.)
4. **Complete or minimal public API.** Every exported type's lifecycle is completable by an external importer: `Hub` must expose `Publish`/`SSE` or be unexported (§6.3); `Deps` must not leak `audit` types (§6.4); `Mode` must be a validated type. Per R3 §1, the README API table must list every exported identifier in §2 and the slice that introduced it.
5. **Timeouts everywhere.** Every `exec` (doctor `ip`/`ethtool`/`tc`/`ping`) and every handler read/write carries a caller-provided `context.Context` with a deadline. `Report(ctx)` returns `ctx.Err()` rather than hanging. Today's bare `exec.Command.Output()` (`doctor.go:41`, `59`, `127`, `132`) fails this criterion.
6. **Wire-conformance tests, not code-review claims.** SSE replay/delta/503/backpressure covered by an HTTP-level test with two concurrent subscribers (not just `Publish`-driven `sse_test.go:42-83` cadence); `/metrics` re-parsed by a strict Prometheus text parser asserting R1 HELP/TYPE/sample EBNF + `text/plain; version=0.0.4`; `/proc/net/dev` parser tested against the R4 header verbatim (today's `iface_test.go:9-14` fixture does this — keep); CC checks tested against R5 exact-token (today's `cc_test.go` does — keep).
7. **Version discipline.** Module tags are immutable once published (R3 §3); a breaking `Snapshot`/`RunOpts`/`Sample` change bumps MAJOR and forks the module path `/v2` (R2 §Major version suffixes, R3 §8); docs (`docs/api.md:13-116`) bump in the same commit as the wire change.
8. **Honest emptiness end-to-end.** `GET /api/results` with no frozen runs returns `{"available": false}` (`server.go:771-774`), `GET /api/audit/live` when idle returns `{"running":false}` alone (`server_audit.go:243-246`), `GET /api/hardware/translate` with no data returns `{"available": false}` (`translate.go:26-32`), `iface.Valid=false` emits no numerics as measurements (`iface.go:28-29`). These four honest-empty proofs must each have a dedicated test asserting ABSENCE (not zero) — the `/metrics` slice (S6) extends the same rule to exposition.

### Suggested first failure to reproduce (before writing S1)

```text
$ (cd $(mktemp -d) && \
  GOPROXY=off go run github.com/Realms4239/cgo/cmd/cgo@latest serve ...)
# today: fails — api requires campaign+audit+results+figures+frontend to
# link at all (server.go:21-30); iface alone cannot be imported without
# the whole binary. S1 makes `go test ./...` inside an iface-only module green.
```

---

*Sources end.* Repo citations are `file:line` against the files listed in §0. External citations are `(R1)`–`(R8)`. Any behaviour not traceable to those is marked **unverified — primary source not found** above.
