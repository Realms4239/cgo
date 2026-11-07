# pkg/iface + cgo-iface as standalone sampler — research

*Date: 2026-11-07 · Scope: `pkg/iface/*.go`, `docs/architecture.md` iface absence, `docs/api.md` `iface_*` fields, `cmd/cgo/server.go` composition, spec `docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md`. Working dir `operator-cockpit`. Research only — no code changed.*

## 0. Method & sources (primary)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | `proc_pid_net(5)` — Linux man page (man-pages 6.19, kernel.org) | https://man7.org/linux/man-pages/man5/proc_pid_net.5.html | fetched 2026-11-07 | `/proc/net/dev` = network device status (received/sent packets, errors, collisions); canonical 2-line header + 16-field rows (`bytes packets errs drop fifo frame compressed multicast \| bytes packets errs drop fifo colls carrier compressed`); `/proc/net` is per-netns via `/proc/self/net` (Linux 2.6.25+) |
| R2 | Prometheus — Exposition formats (Prometheus Text Format) | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched 2026-11-07 | `text/plain; version=0.0.4` Content-Type, UTF-8 + `\n` lines, `# HELP`/`# TYPE` (one per metric, TYPE before first sample), sample EBNF, `counter/gauge/histogram/summary/untyped` vocabulary, grouping/sorting, float/timestamp grammar (`ParseFloat`/`ParseInt`, `NaN`/`+Inf` valid values) |
| R3 | Go Modules Reference | https://go.dev/ref/mod | fetched 2026-11-07 | module = packages released, versioned, distributed together; module path = canonical ID + import prefix; `go` directive = mandatory minimum (≥1.21); `/vN` suffix required for v2+ (import-compatibility rule), forbidden for v0/v1; `require`/`replace`/`exclude`/`retract` semantics; MVS |
| R4 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched 2026-11-07 | §1 MUST declare precise public API; §3 immutability once released; §4 `0.y.z` MAY change; §6 PATCH = compat bug-fix; §7 MINOR = compat functionality; §8 MAJOR = incompat API change |
| — | `tc(8)`, `netlink(7)`, `ip(8)` netns semantics beyond R1, SNMP IF-MIB / RouterOS counters, IFNAMSIZ name limits, counter wrap/reset semantics (64-bit), Mbps decimal-vs-binary convention | — | **unverified — primary source not found** (not fetched) | any shaping enforcement, netns exec form, SNMP OID semantics, colon-in-name legality, NIC counter reset cause beyond "counter decreased" |
| — | Repo | `pkg/iface/iface.go` (200 lines), `iface_test.go`, `sampler_test.go`, `cmd/cgo/server.go`, `docs/api.md`, `docs/architecture.md`, `docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md`, `go.mod`, `kit/run-dualns.sh` | read 2026-11-07 | all `file:line` claims below |

Method: every technical claim below cites `file:line` or `(R#)`; anything else is labelled **unverified — primary source not found**. Prior combined study is `docs/research/2026-11-07-module-api-doctor-iface.md`; this file is the iface-only sampler slice, re-sourced against R1–R4 above.

## 1. Responsibility (as-shipped)

Single responsibility: turn cumulative kernel counters for one named interface into a windowed rate `Sample`, with an explicit validity/gap contract so a zero is never read as a measurement. `pkg/iface/iface.go:1-3` states "débit d'interface observé par les compteurs noyau. stdlib seul : /proc/net/dev sur Linux" (repo comment, `iface.go:1-3`).

Concretely as-shipped (`pkg/iface/iface.go:16-200`):

- Parse `/proc/net/dev` (or a same-format file) into per-interface `Counters{RXBytes, RXDrops, TXBytes, TXDrops}` (`iface.go:20-25`, `iface.go:62-100`). Field map per R1 header: RX `f[0]` bytes, `f[3]` drop; TX `f[8]` bytes, `f[11]` drop (R1: rx 0–7 `bytes packets errs drop fifo frame compressed multicast`, tx 8–15 `bytes packets errs drop fifo colls carrier compressed`).
- Derive `RXMbps`/`TXMbps` over measured windows (`iface.go:169-170`: `bytes*8/1e6/dt.Seconds()`), carrying `WindowMs`, `SampleTS`, drop counters, and `Valid`/`Gap` (`iface.go:30-41`).
- Expose two-method seam `Reader{Read(); Latest()}` (`iface.go:44-47`) so a future SNMP/RouterOS adapter returns the same `Sample` (spec target `…-design.md:591`; adapter itself **unverified — primary source not found**).
- Declare provenance, never guess it: `NormalizeSource` maps only `twin|prod|unknown` through, everything else to `unknown` (`iface.go:51-57`).
- Feed the live frame additively: `cmd/cgo/server.go:27-30` embeds `iface.Sample` in `liveFrame{campagne.Snapshot; iface.Sample}`, `cmd/cgo/server.go:63-73` wires `GetSnap→Latest()` and `NextSnap→Read()`; `docs/api.md:28-36` documents the `iface_*` fields and `iface_valid:false` invalidating all numerics, with `iface_gap ∈ {warmup, absent, reset, slow, read, clock}`.

What it deliberately does not do: no packets exposure (`iface.go:19` "les paquets ne sont pas exposés, personne ne les affiche"), no auto-detection of the interface (`cmd/cgo/server.go:32-33` "Pas d'auto-détection : une mauvaise interface serait une mesure mal étiquetée"), no shaping/probing, no persistence, no ticker of its own (driven by the hub's 10 Hz clock, `iface.go:107-108` + spec `…-design.md:165`).

Architecture-doc gap: `docs/architecture.md:10-21` lists ten behaviours (`api/campagne/qdisc/probe/metrics/results/figures/audit/profile/model`) but has no `pkg/iface` (nor `pkg/doctor`) row; the iface contract lives only in spec `…-design.md` + `docs/api.md:28-36`. A standalone release must add the architecture row (see §7 S3).

## 2. Exported API (file:line)

Complete non-test inventory for `pkg/iface`. Unexported identifiers listed separately; they are not API.

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Counters` | struct | `iface.go:20-25` | `{RXBytes, RXDrops, TXBytes, TXDrops uint64}` — bytes + drops only; packets parsed per R1 but deliberately not exposed (`iface.go:18-19`). |
| `Sample` | struct | `iface.go:30-41` | JSON `iface_name/iface_source/iface_rx_mbps/iface_tx_mbps/iface_rx_drops/iface_tx_drops/iface_window_ms/iface_sample_ts/iface_valid/iface_gap`. `Valid=false` invalidates all numerics (`iface.go:28-29`). |
| `Reader` | interface | `iface.go:44-47` | `{Read() Sample; Latest() Sample}` — the future SNMP/RouterOS seam. |
| `NormalizeSource(s string) string` | func | `iface.go:51-57` | Total function to `twin\|prod\|unknown`; callers log refusals (`cmd/cgo/server.go:44-46`). |
| `ReadCounters(path string) (map[string]Counters, error)` | func | `iface.go:62-100` | Strict: skips 2 header lines + blanks; `LastIndexByte(':')` split; requires ≥16 fields; truncated line = error, missing colon = error, missing file = error. |
| `Sampler` | struct | `iface.go:109-117` | `{mu, path, name, source, now func(), prev *prev, last Sample}`; `prev`/`last` unexported. One sampler = one interface, hub-10 Hz read (`iface.go:107-108`). |
| `New(name, source string) *Sampler` | func | `iface.go:120-122` | Production: `/proc/net/dev` (`iface.go:16`) + `time.Now`. |
| `func (s *Sampler) Read() Sample` | method | `iface.go:134-179` | Advances base; gaps `read`/`absent`/`warmup`/`clock`(`dt≤0`)/`slow`(`dt>500ms`)/`reset`(counter decreased); rate `delta*8/1e6/dt` decimal Mbps, no clamp/floor. |
| `func (s *Sampler) Latest() Sample` | method | `iface.go:182-186` | Never advances base; lock-guarded copy of `last`. Used by `/api/state` (`cmd/cgo/server.go:67-69`, `docs/api.md:35-36`). |

Unexported (not API): `procNetDev="/proc/net/dev"` (`iface.go:16`), `prev` (`iface.go:102-105`), `newSampler(path,name,source,now)` (`iface.go:125-129`), `gap`/`emit`/`round1` (`iface.go:188`, `195`, `200`).

Tests that pin semantics (evidence, not API):

- `pkg/iface/iface_test.go:25-40` — `lo`/`eth0`/colon-name `eth:0` (max-`uint64` `18446744073709551615`) parse; `iface_test.go:42-53` — short-line + missing-file rejection.
- `pkg/iface/sampler_test.go:35-66` — warmup→valid rate math (`125000 B / 100 ms = 10 Mbps`); `sampler_test.go:68-109` — reset/slow/clock rebaseline; `sampler_test.go:111-155` — absent→warmup, read-error, `Latest` initial-warmup; `sampler_test.go:157-183` — concurrent `Read`/`Latest` under `-race`, `New` normalises unknown source.

## 3. Dependencies

### 3.1 Direct imports (evidence)

- `pkg/iface/iface.go:6-14`: stdlib only (`fmt`, `math`, `os`, `strconv`, `strings`, `sync`, `time`). No internal import; no external module. Leaf-package property stated in spec `…-design.md:269`.
- `go.mod:1-10`: `module github.com/Realms4239/cgo`, `go 1.25`, direct `require` = `fyne.io/fyne/v2`, `charmbracelet/bubbletea`, `lipgloss`, `golang.org/x/sys` (GUI/TUI — none imported by `pkg/iface`). Under R3, a module is "released, versioned, and distributed together" and its path is its ID, so today `iface` cannot be versioned independently (R3 §Modules, §Module paths); `go` directive `1.25` would be the extracted module's floor (R3 §`go` directive mandatory minimum).
- Composition deps (not library imports): `cmd/cgo/server.go:15-23` imports `api/campagne/doctor/iface/model/probe/qdisc/results` together; `cmd/cgo/server.go:34-48` reads `CGO_IFACE`/`CGO_IFACE_SOURCE`, refuses non-Linux (`cmd/cgo/server.go:39-42`), logs unknown source (`cmd/cgo/server.go:44-46`) and calls `iface.New(name, src)` (`cmd/cgo/server.go:47`); `kit/run-dualns.sh:61-62` sets `CGO_IFACE=c-cli` / `CGO_IFACE_SOURCE=twin` for the twin.

### 3.2 Host / platform dependencies

| Dependency | Usage | Severity binding |
|---|---|---|
| Linux `/proc/net/dev` | production parse target (`iface.go:16`, `120-122`) | R1: format + per-netns semantics; on non-Linux `envSampler` returns nil + log (`cmd/cgo/server.go:39-42`); package still compiles/tests via temp file (`iface.go:2-3`) |
| `/proc/self/net` netns identity | which namespace's counters are read | R1: "since Linux 2.6.25, `/proc/net` is a symbolic link to `/proc/self/net`" — sampler reads the server's own namespace; cross-namespace interfaces report `absent` (`iface.go:144-148`) by design, not by misdetection |
| `CGO_IFACE` / `CGO_IFACE_SOURCE` env | interface name + declared provenance | repo policy (`cmd/cgo/server.go:32-48`, spec `…-design.md:344-346`); SNMP/RouterOS poll cadence **unverified — primary source not found** |

Hardcoded surface: `/proc/net/dev` in `New` (`iface.go:120-122` via `iface.go:16`) is the only blocker without an exported seam today — `newSampler(path, ..., now func())` (`iface.go:125`) already injects path + clock but is unexported, so external embedders cannot use it (gap §5.1).

## 4. Primary-source requirements

Minimum external authorities any standalone `cgo-iface` release must re-verify (with the exact claim each governs):

| ID | Authority | Claim it governs | Repo binding |
|---|---|---|---|
| R1 | `proc_pid_net(5)` | `/proc/net/dev` 16-field header (`bytes packets errs drop fifo frame compressed multicast \| bytes packets errs drop fifo colls carrier compressed`); 2-line header + per-device lines; `/proc/net` per-netns via `/proc/self/net` | `iface.go:16`, `62-100` (f[0]/f[3]/f[8]/f[11]); `iface_test.go:9-14` fixture; `cmd/cgo/server.go:39-42` non-Linux disable |
| R2 | Prometheus exposition formats | `Content-Type: text/plain; version=0.0.4`; HELP/TYPE one-per-metric, TYPE-before-sample; `counter/gauge` vocabulary; sample EBNF; `\n`-terminated last line; grouping/sorting; float (`NaN`/`+Inf` valid) + timestamp grammar (`ParseInt`, ms-since-epoch) | target `/metrics` exposition for `cgo_iface_*` (see §7 S3); today's `pkg/api/metrics.go` static-zero endpoint is the anti-pattern to avoid (prior study §6.1) |
| R3 | Go Modules Reference | What counts as a releasable module; module-path lexical rules; `/vN` required for v2+, forbidden for v0/v1; `go` directive = mandatory minimum; `require`/`replace` semantics; MVS | `go.mod:1-3`; split plan §7 |
| R4 | SemVer 2.0.0 | §1 MUST declare precise public API; §3 immutability; §6–§8 PATCH/MINOR/MAJOR discipline; §4 `0.y.z` instability | version policy §7 |

Fetch-or-it's-unverified rule: any claim about `tc(8)` qdisc byte plane vs NIC byte plane equivalence, netlink/eBPF/SNMP/RouterOS counters, IFNAMSIZ / colon-in-name legality, 64-bit counter wrap or NIC-reset cause, decimal-vs-binary Mbps convention beyond the code's own `/1e6` — label **unverified — primary source not found** until the authoritative man page / RFC / spec is fetched and quoted. Today's `iface.go:72` colon handling (`LastIndexByte`, tested `eth:0` at `iface_test.go:37-39`) is a defensible superset; whether kernel names can contain `:` is **unverified — primary source not found**.

## 5. Conformity details (spot checks against primary sources)

### 5.1 Field indices match R1 exactly; parser strictness is the honesty contract

R1 header decodes to rx 0–7 (`0 bytes, 1 packets, 2 errs, 3 drop, 4 fifo, 5 frame, 6 compressed, 7 multicast`) and tx 8–15 (`8 bytes, 9 packets, 10 errs, 11 drop, 12 fifo, 13 colls, 14 carrier, 15 compressed`). `iface.go:81-96` reads f[0]=rx bytes, f[3]=rx drops, f[8]=tx bytes, f[11]=tx drops — conformant (R1). Parser strictness (short line = error `iface.go:78-80`; missing colon = error `iface.go:73-75`; missing file = error `iface.go:63-66`) implements the comment "Une ligne tronquée est une erreur, pas un saut" (`iface.go:60-61`) — covered `iface_test.go:42-53`.

Rate semantics: `bits*8/1e6/dt` (`iface.go:169-170`) is decimal Mbps with no floor/clamp — matches comments (`iface.go:131-133`), spec D3 (`…-design.md:166`), and `sampler_test.go:49-65`. Decimal-vs-binary Mbps convention as an external standard is **unverified — primary source not found**; only the code's `/1e6` is claimed.

### 5.2 Gap taxonomy + hub-driven cadence

`slow` at `dt>500ms` resets the base and reports `gap:"slow"` rather than a long-window average (`iface.go:160-163`); `reset` on counter decrease (`iface.go:164-167`); `clock` on `dt≤0` (`iface.go:156-159`); `warmup` on first sight (`iface.go:149-152`, initial `last` at `iface.go:127`); `absent`/`read` clear `prev` (`iface.go:140-148`). All six gaps are the full spec vocabulary (`…-design.md:167`, `docs/api.md:32-34`). `Latest()` never advances (`iface.go:182-186`) so `/api/state` is side-effect-free while the hub's `Read()` advances — matches `docs/api.md:35-36` and spec D2 (`…-design.md:165`). Counter-reset cause (NIC bounce vs driver reload vs 64-bit wrap) is a single `reset` arm today; attributing it to any of those is **unverified — primary source not found**.

### 5.3 Prometheus mapping: rate = gauge, drops = counter, `valid` = honest-empty flag

R2: types are `counter/gauge/histogram/summary/untyped` with HELP/TYPE lines first and a unique name+labels per sample. Target mapping (see §7): `cgo_iface_rx_mbps`/`cgo_iface_tx_mbps` as `gauge` (windowed rate, goes down), `cgo_iface_rx_drops_total`/`cgo_iface_tx_drops_total` as `counter` (cumulative since creation, never decreases within a boot — a decrease folds into the `reset` gap per R2 counter semantics), `cgo_iface_valid` boolean companion. R2 warns "Types and docstrings not integral part of the syntax, meaning little-to-nonexistent metric contract validation" — so a scraper accepts fabricated zeros; the project's absent≠zero rule (`iface.go:28-29`) must become "omit the rate sample when `Valid=false`, still emit `cgo_iface_valid 0` and `cgo_iface_gap{reason=...}`" rather than emitting `0` as a measurement.

## 6. Gap analysis (what blocks standalone release)

Ordered by severity. Each gap cites the code and the primary source it would violate or the policy it breaks.

### 6.1 Injection seam is unexported (BLOCKER)

`newSampler(path, name, source string, now func() time.Time)` (`iface.go:125-129`) already parameterises file path + clock for Windows tests, but only `New(name, source)` (`iface.go:120-122`) is exported, hardcoding `/proc/net/dev` + `time.Now`. An external embedder cannot point the sampler at a fixture, a per-netns mount, or a fake clock without forking the file. Against R4 §1 (public API must be precise and comprehensive — a consumer cannot discover the testable surface), this is a release blocker. Fix direction: export `NewSampler(opener/clock, name, source)` (see §7). The same class of bug as `pkg/api`'s `Publish`/`SSE` unexported-on-exported-type pattern (prior study §6.3).

### 6.2 Single-interface, single-file, no opener (MAJOR)

`Sampler` holds exactly one `name`/`path` (`iface.go:109-117`); `ReadCounters` re-reads the whole `/proc/net/raw` file (`iface.go:63`) per `Read()` at 10 Hz. Two consequences for a standalone tool: (a) multi-interface observation (WAN+LAN, router two-end) needs N samplers + N file reads — the Phase-2 spec (`…-design.md:575-577`) names operator-chosen plural interfaces, but the API cannot serve them in one sample; (b) a whole-file `os.ReadFile` per tick is waste but not incorrect under R1 (format is small); no `ReadCounters`-with-`Opener` seam exists to let a future SNMP/RouterOS adapter (`iface.go:43` comment) reuse the rate/gap logic without the file. Fix: `Opener func() (map[string]Counters, error)` + `Sampler` per name stays, but surface a `MultiSampler`/`Snapshot()` that fans out in one open (see §7).

### 6.3 Untyped `Gap` + unchecked `Source` strings (MAJOR)

`Gap` is a bare `string` (`iface.go:40`) and `Source` a bare `string` (`iface.go:32`); `NormalizeSource` (`iface.go:51-57`) does the only validation. A typo like `gap="wramup"` compiles and renders verbatim (spec explicitly wants unknown reasons shown verbatim, `…-design.md:231-233`, so the risk is server-side production of a misspelled gap, not client hiding). R4 §1 precision argues for `type Gap string` + exported consts (`GapWarmup … GapClock`) and `type Source string` (`SourceTwin/SourceProd/SourceUnknown`) so the six reasons are compiler-checked. Changing `Sample.Gap` from `string` to a string type is PATCH-compatible if the underlying type stays `string` for JSON (R4 §6); adding the consts is MINOR (R4 §7); changing `Sample` field names/tags is MAJOR + `/v2` (R3 §Major version suffixes, R4 §8).

### 6.4 No `context`, no scrape/exposition surface (MAJOR — anti-toy)

`Read()`/`Latest()` take no `context.Context` (`iface.go:134`, `182`); the file read is blocking `os.ReadFile` (`iface.go:63`) with no deadline — a hung FUSE/`/proc` mount would stall the hub tick. `Sample` carries no Prometheus exposition: there is no `cgo_iface_rx_mbps`/`_total`/`_valid` output, no HELP/TYPE, no `Content-Type: text/plain; version=0.0.4` (R2). The only exposition in the repo (`pkg/api/metrics.go:46-63`, six static zeros with `site="unknown"`) is the static-zero anti-pattern (prior study §6.1/§5.4): a scraper accepts it (R2 gives little-to-nonexistent contract validation for types/docstrings) but every value is fabricated. A standalone sampler that ships without a wired exposition repeats the toy's sins — §7 S3 adds exposition as a release gate, §8 item 5.

### 6.5 `docs/architecture.md` absence + module mono-repo (MINOR, but release-facing)

`docs/architecture.md:10-21` has no `pkg/iface` row (and no `doctor` row), while `docs/api.md:28-36` documents the fields fully. Under R3 ("module is released/versioned/distributed together") the single `go.mod:1` (`module github.com/Realms4239/cgo`, `go 1.25`) cannot version `iface` independently; under R4 §1 the public API is not documented at the architecture level. Neither blocks extraction, but both must close before tagging `iface@v0.1.0` (§7 S4–S5), or the first release inherits an undocumented surface. No `replace`/`exclude`/`retract` directives exist today (`go.mod:5-60` has only `require` blocks) — clean birth per R3 §`retract`.

### 6.6 Smaller gaps (MINOR / polish)

- `Sample` JSON tags (`iface.go:31-40`) double the snapshot size at 10 Hz with no `omitempty` on `Valid=false` gaps — delta compression skips only profile/qdisc/cc/repetition/event_id/phase (prior study §6.6). For a library this is a consumer choice; document it, don't re-shape silently.
- `round1` (`iface.go:200`) rounds `WindowMs` to 0.1 ms but leaves rates unrounded (`iface.go:169-170`) — deterministic rounding policy is undocumented; pin it in the extracted package doc (PATCH if changed).
- `ReadCounters` uses `LastIndexByte(':')` (`iface.go:72`) — correct for `eth:0` (`iface_test.go:37-39`) but colon-in-name legality is **unverified — primary source not found**; keep as documented superset.
- `Latest()` copies `last` under lock (`iface.go:182-186`) but `Read()` re-reads the entire file while holding the lock (`iface.go:135`) — a slow read serialises `Latest()` too. Fine at 10 Hz; flag in extracted docs so consumers size the lock accordingly.

## 7. Target tool: `cgo-iface` as independent sampler repo

Proposed identity (R3): new module `github.com/Realms4239/cgo/iface` (or a `cgo-iface` repo with the module at its root), `go 1.25` floor from `go.mod:3` (R3 §`go` mandatory minimum), MIT `LICENSE` per repo root. The repo module is `github.com/Realms4239/cgo` (`go.mod:1`); a standalone path must not collide — append `/iface` per R3 §Module paths (subdirectory = tag prefix).

```go
package iface

// Keep (already conformant, R1 §5.1):
type Counters struct { RXBytes, RXDrops, TXBytes, TXDrops uint64 }  // iface.go:20

// Tighten (R4 §1 precision; R3 `/vN` only later):
type Source string
const ( SourceTwin Source = "twin"; SourceProd Source = "prod"; SourceUnknown Source = "unknown" ) // iface.go:51
type Gap string
const ( GapWarmup Gap = "warmup"; GapAbsent Gap = "absent"; GapReset Gap = "reset"; GapSlow Gap = "slow"; GapRead Gap = "read"; GapClock Gap = "clock" ) // iface.go:134-179

type Sample struct { Name, Source string; RXMbps, TXMbps float64; /* + drops/window/ts/valid/gap, iface.go:30 */ Valid bool; Gap Gap }

// Seams the standalone needs (gap §6.1–6.2):
type Reader interface { Sampler() Sample; Latest() Sample }        // iface.go:44 renamed only if `/v0.1.0` -> breaking; keep Read()
type Opener func() (map[string]Counters, error)                    // replaces os.ReadFile in ReadCounters (iface.go:63)
type Clock  func() time.Time                                        // iface.go:114 now func()

func NewSampler(op Opener, clk Clock, name string, source Source) *Sampler // new exported injectable ctor (ex-`newSampler`, iface.go:125)
func New(name, source string) *Sampler                                   // keeps /proc/net/dev production default (iface.go:120)
func ReadCounters(path string) (map[string]Counters, error)              // keep; document R1 field map (iface.go:62-100)
func NormalizeSource(s string) Source                                    // iface.go:51 widened to typed Source (PATCH-compatible if `Source` is `string`)

func (s *Sampler) Read(ctx context.Context) Sample                       // ctx added for hanging-read honesty (§6.4); MINOR if `/v0.x`
func (s *Sampler) Latest() Sample                                        // iface.go:182 — never advances
func (s *Sampler) Expose() []PromSample                                  // Prometheus exposition (R2), see §7 S3
```

Prometheus exposition (R2-gated, part of the target API):

| Metric | TYPE | Labels | Sample rule |
|---|---|---|---|
| `cgo_iface_rx_mbps` / `cgo_iface_tx_mbps` | `gauge` | `iface`, `source` | emit only when `Valid`; omit (absent) on gaps — never `0` as measurement |
| `cgo_iface_rx_drops_total` / `cgo_iface_tx_drops_total` | `counter` | `iface`, `source` | cumulative `RXDrops`/`TXDrops` (`iface.go:35-36`); present on valid samples |
| `cgo_iface_valid` | `gauge` | `iface`, `source` | `1` valid, `0` gap — the honest-empty companion |
| `cgo_iface_gap` | `gauge` | `iface`, `source`, `reason` | `1` for the active gap (`warmup/absent/reset/slow/read/clock`), absent otherwise |

HELP/TYPE precede first sample, one per metric, `Content-Type: text/plain; version=0.0.4`, `\n`-terminated (R2) — see §8 item 5.

Linking rule (R3 + R4 §1): the standalone module imports **nothing** from `github.com/Realms4239/cgo` (stdlib-only per `iface.go:6-14`); `cmd/cgo` then *depends on* `cgo/iface` (`cmd/cgo/server.go:18`, `34-48`, `63-73` becomes `iface.NewSampler(…)` + typed `Reader`). No `replace`, no `exclude`/`retract` at birth (R3 §`retract` only applies in main-module `go.mod`; `go.mod:5-10` currently has none). SemVer: extracting `include` is `v0.1.0` (R4 §4 `0.y.z` MAY change); typed `Gap`/`Source` is PATCH (`string` underpinning, R4 §6); exported `NewSampler`/`Opener`/`Expose` is MINOR (R4 §7); any `Sample` field rename/tag drop is MAJOR + `/v2` suffix (R3 §Major version suffixes, R4 §8).

## 8. Slices (ordered, one-line each)

| Slice | Content | Depends on | Acceptance |
|---|---|---|---|
| S0 | Extract: copy `pkg/iface/*.go` into `cgo/iface`, new `go.mod` (`module …/iface`, `go 1.25`), `LICENSE` (MIT per repo root); no behaviour change | — | `go test ./...` passes `iface_test.go:25-53` + `sampler_test.go:35-183` unmodified; `GOPROXY=off go list -m all` resolves stdlib-only |
| S1 | Seams: export `NewSampler(Opener, Clock, …)` + `Opener`/`Clock` types; type `Gap`/`Source` + consts; `Read(ctx)` deadline; `ReadCounters` R1 field-map doc; keep `New`/`NormalizeSource` behaviour | S0 | existing tests pass with fake `Opener`/clock (no `/proc` on test host); blocking-`Opener` test asserts `Read(ctx)` returns `ctx.Err()` within 2 s |
| S2 | Multi-name fan-out: `Sampler` per name stays; add `SnapshotRange(names)` that opens once per tick and returns one `Sample` per name (Phase-2 WAN+LAN ready, spec `…-design.md:575-577`); `SampleTS` stays per-read wall-clock ms (`iface.go:176`) | S1 | one-open fixture yields 2 valid samples with shared `WindowMs`; absent name yields `absent` while present name stays valid |
| S3 | Exposition (`Expose()`) + wiring: gauge/counter/`_valid`/`_gap` per §7, R2 HELP/TYPE + `version=0.0.4` + `\n` termination; add `docs/architecture.md` `pkg/iface` row; `docs/api.md:28-36` cross-link to `cgo/iface` | S1 | golden exposition test (R2 sample EBNF) + absent-on-gap assertion; `curl /metrics` against a wired binary has no zero-as-measurement when `iface_valid:false` |
| S4 | Re-point CGO at it: `cmd/cgo/server.go:18,34-48,63-73` import `cgo/iface`; delete `pkg/iface` (or thin alias); frontend `live.ts`/`LiveView.tsx` unchanged (wire shape preserved); `kit/run-dualns.sh:61-62` unchanged | S1–S3 | `grep -r pkg/iface` outside `docs/research` shows zero non-test hits; `make verify` green (all six gates) |
| S5 | Release train: tag `iface@v0.1.0`; `go.mod` `require` bump in CGO; SemVer discipline (R4 §6–§8) + R3 tag-prefix (`/iface`) attested | S4 | `GOPROXY=off go list -m all` per module resolves; `v0.1.0` builds from tag alone; release notes state PATCH/MINOR/MAJOR mapping for `Gap`/`NewSampler`/`Sample` |

Slice notes: S0 is zero-dependency and cannot break other slices. S1 removes the only standalone blocker (§6.1) and is the honest-empty hinge — it carries the 500 ms `slow`/`reset` rebaseline unchanged (`iface.go:160-167`). S3 is the anti-toy hinge: without it the extraction ships a sampler nobody can scrape (R2). Total new external deps: **zero** (stdlib only).

## 9. Anti-toy criteria (release gates — all must pass)

1. **Honest emptiness end-to-end.** `Valid=false` emits no rate as measurement: `TestSampler*` gap cases (`sampler_test.go:40-43,76-77,95-96,105-108,115-116,139-141`) + exposition test assert *absence* of `cgo_iface_rx_mbps`/`_tx_mbps` on gaps, while `cgo_iface_valid 0` + `cgo_iface_gap{reason=…}` are present (R2 companion-flag pattern; R2 warns scraper accepts zeros, so the test must assert omission, not zero).
2. **R1-conformant field map, strict parser.** `ReadCounters` reads `f[0]/f[3]/f[8]/f[11]` per R1 header; short line / missing colon / missing file are errors, not skips (`iface.go:60-80`, `iface_test.go:42-53`). Any future widening of exposed fields (packets, errs) is MINOR (R4 §7); any remapping of indices is MAJOR + `/v2` (R3, R4 §8) and must cite a new R1 fetch.
3. **Measured windows only.** `WindowMs` is the real `dt` (`iface.go:154-155`, `round1` `iface.go:200`); `dt>500ms` yields `slow` + rebaseline, never a long-window average (`iface.go:160-163`); `dt≤0` yields `clock` (`iface.go:156-159`). A test with `dt=600ms` asserts `Valid=false` + `Gap=slow` + no rate sample (preserved verbatim from `sampler_test.go:87-109`).
4. **Provenance declared, never guessed.** `NormalizeSource`/`NewSampler` accepts only `twin|prod|unknown` (`iface.go:51-57`, `sampler_test.go:175-183`); the exposition label `source` carries the declared value, never a detected one. Unknown input ⇒ `unknown` + caller log.
5. **Prometheus exposition valid and wired.** Every emitted metric has exactly one `# HELP` + one `# TYPE` before its first sample (`Content-Type: text/plain; version=0.0.4`, `\n`-terminated, unique name+label sets — R2); label values escape `\`, `"`, `\n` per R2 `escaped_string` (needs a `\\`/`\"` fixture test); `cgo_iface_*_total` are `counter` (cumulative), `cgo_iface_rx/tx_mbps` + `cgo_iface_valid` are `gauge`; an absent measurement is never emitted as `0` (the `pkg/api/metrics.go` static-zero anti-pattern, prior study §5.4, is the documented counter-example).
6. **Timeouts and exact seams.** Every `Opener` call carries `context.Context` with a deadline; `Read(ctx)` returns `ctx.Err()` on hang (§6.4). `Latest()` never advances the baseline (`iface.go:182-186`, `sampler_test.go:144-155`) — a `Latest`-between-`Read` test asserts the next `Read` window is unchanged. No ticker, no goroutine, no shaping in the module.

*Skipped: SNMP/RouterOS adapter (Phase 2, spec `…-design.md:567-595`, SNMP semantics **unverified — primary source not found**), netlink/eBPF alternative samplers (rejected `…-design.md:360-366`), twin live re-derivation beyond the published `docs/VERIFY.md:38` twin numbers; no code or commit per task scope. Risk: extracted `Sample` JSON shape is frozen by `cmd/cgo/server.go:27-30` + frontend `live.ts`/`LiveView.tsx` — any field rename is MAJOR (R3/R4) and needs a coordinated CGO + dashboard bump; keep wire shape stable through S4.*
