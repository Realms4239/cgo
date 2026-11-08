# pkg/profile + model profile types as a standalone registry tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/profile` + `pkg/model` profile types (`Profile`, `Profiles`, `ProfilesMu`, `BuiltinProfileIDs`) read as if extracted tomorrow into `cgo-profiles`. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | Go Modules Reference | https://go.dev/ref/mod | fetched 2026-11-07 | module vs package, versioning, `go.mod`, nested modules, import rules |
| R2 | Semantic Versioning 2.0.0 | https://semver.org/ | fetched 2026-11-07 | public-API declaration, MAJOR/MINOR/PATCH, deprecation cycle |
| R3 | RFC 8290 — FQ-CoDel | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched 2026-11-07 | what `fq_codel` means; §5.2 parameters; and (negatively) that RFC 8290 defines **no** link-profile registry |
| R4 | Prometheus data model | https://prometheus.io/docs/concepts/data_model/ | fetched 2026-11-07 | metric/label identity, label naming, cardinality |
| — | CAKE scheduler documentation | https://www.kernel.org/doc/html/latest/networking/scheduler/sch_cake.html | **unverified — primary source not found** at that path (see `docs/research/2026-11-07-module-model.md:13`) | `Cake` label provenance — not claimed |
| — | `tc`/`netem` normative man page | — | **unverified — primary source not found** (not fetched in this pass) | `netem`/`tbf` invocation semantics behind `pkg/qdisc/qdisc.go:76-135` — code is cited locally only |
| — | Starlink LEO values / pricing (P4) | `pkg/model/types.go:47-50` (comment only) | **unverified — primary source not found** (no pricing source fetched) | P4 provenance — not claimed |

Method: read `pkg/profile/profile.go` + all `pkg/profile/*_test.go`, `pkg/model/types.go`, `pkg/model/cc.go`, `pkg/model/profile_asym_test.go`, `pkg/model/cc_test.go`, `docs/api.md` profile endpoints (§ Profils, lines 38–44), `pkg/api/server.go` handlers (`GET /api/profiles`, `POST /api/profile/import|delete`, `GET /api/profile/list`), downstream consumers (`pkg/campagne/config.go`, `pkg/audit/match.go`, `cmd/cgo/calibrate.go`, `cmd/cgo/server.go`); fetch R1–R4. Every technical claim below cites either `file:line` or `(R#)`. Anything else is labelled **unverified — primary source not found**. Naming note: this task brief says “RFC 8290 link profiles” — R3 verified by full fetch is the FQ-CoDel memo and contains no link-profile registry concept (R3 §1, §5.2 list scheduler parameters `interval`, `target`, `limit`, `quantum`, `flows`, not capacity/delay/jitter/loss profiles). The “link profile” is therefore treated in this note as a CGO-local invention, not an RFC 8290 requirement (§4.3).

## 1. Responsibility

**Today** the two packages together form a small *registry + persistence* layer: `pkg/model` owns the value object and the in-memory global state; `pkg/profile` owns the only mutations through a single JSON file. That split is already documented as “dynamic registry in `pkg/profile`, which imports `model`” (`docs/research/2026-11-07-module-model.md:28`).

Concretely:

- **The value object.** `Profile{ID, CapacityMbps, CapacityUpMbps?, DelayMs, JitterMs, LossPct, LossBurst?P/R/H/K}` (`pkg/model/types.go:11-28`). `CapacityMbps` is the historical reference; `CapacityUpMbps` is optional and absent means symmetric up = down (`pkg/model/types.go:7-11`, `pkg/model/types.go:30-41`). `CapUp()` clamps silently: only `0 < up < down` is honoured, otherwise down is returned (`pkg/model/types.go:36-41`). The burst fields are marked deprecated and are kept only for the external CSV import format and the profiles API response (`pkg/model/types.go:18-23`).
- **The built-in reference set.** Four literals: P1 80/20/2/0, P2 20/100/15/0.5, P3 5/600/30/1, P4 100/40/20/0.3 in Mbps/ms/ms/% (`pkg/model/types.go:43-51`). Values are literals in source with no attached measurement provenance in this package (P4 comment cites research dated 2026 without a fetchable source — **unverified — primary source not found** for the P4 provenance claim).
- **The global state.** `Profiles map[string]Profile` (`pkg/model/types.go:43`), guarded by `ProfilesMu sync.RWMutex` with the discipline “RLock en lecture, Lock en écriture” (`pkg/model/types.go:53-57`), and `BuiltinProfileIDs` frozen at init from the map (`pkg/model/types.go:59-68`) so `/api/profiles` marks `imported = id non natif (natifs : P1–P4)` (`docs/api.md:42`).
- **The registry operations.** `Reserved` guards P1–P4 against import and delete (`pkg/profile/profile.go:14-20`); `Import` writes through to memory and to `data/profiles.json` under lock (`pkg/profile/profile.go:22-46`); `Load` merges persisted entries into memory, skipping reserved ids (`pkg/profile/profile.go:48-67`); `Delete` removes from both (`pkg/profile/profile.go:69-93`). The file is `const file = "data/profiles.json"` (`pkg/profile/profile.go:12`), i.e. CWD-relative.
- **The HTTP surface.** `GET /api/profiles` (sorted, `imported` flag) (`pkg/api/server.go:531-559`, contract `docs/api.md:42`); `GET /api/profile/list` (full map) (`pkg/api/server.go:1252-1263`, `docs/api.md:43`); `POST /api/profile/import` (CSV positional `id,capacity_mbps,delay_ms,jitter_ms,loss_pct[,capacity_up_mbps,loss_burst_p,r,h,k]` or JSON `model.Profile`) (`pkg/api/server.go:1158-1230`, `docs/api.md:44`); `POST /api/profile/delete` (`pkg/api/server.go:1233-1251`, absent from `docs/api.md`).

**Negative scope** (what it does not do): it applies no shaping (`pkg/qdisc` owns `ApplyNetem` from `DelayMs/JitterMs/LossPct`, `pkg/qdisc/qdisc.go:76-80`); it runs no probes; it does not resolve profiles for the campaign matrix itself (`pkg/campagne/config.go:71-84` reads `model.Profiles` directly); it does not match audits to the reference set (`pkg/audit/match.go:16-56`). That negative framing — registry + persistence, not engine — is what makes extraction plausible, but §5 shows the current shape is unsafe to publish as-is.

## 2. Exported API (file:line)

Complete inventory (tests are not API but pin behaviour; listed last).

### 2.1 `pkg/model/types.go` — profile types only

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Profile` | struct | `pkg/model/types.go:11-28` | `ID`, `CapacityMbps`, `CapacityUpMbps,omitempty`, `DelayMs`, `JitterMs`, `LossPct`, deprecated `LossBurstP/R/H/K,omitempty` (`pkg/model/types.go:18-27`). JSON tags are the wire schema. |
| `Profile.CapDown()` | method | `pkg/model/types.go:31` | returns `CapacityMbps` (the historical reference, `pkg/model/types.go:30`). |
| `Profile.CapUp()` | method | `pkg/model/types.go:36-41` | returns `CapacityUpMbps` iff `0 < up < down`, else down (`pkg/model/types.go:33-40`). Silent cap, no error. |
| `Profiles` | var `map[string]Profile` | `pkg/model/types.go:43-51` | P1–P4 literals. Exported mutable global. |
| `ProfilesMu` | var `sync.RWMutex` | `pkg/model/types.go:53-57` | discipline documented in comment, not enforced by the type. |
| `BuiltinProfileIDs` | var `map[string]bool` | `pkg/model/types.go:59-68` | frozen at init; determines `imported` (`docs/api.md:42`). |

Out of scope for this note but adjacent: `Qdisc/CC/Gate/Event/AQMEvalHeader/LinkAuditHeader` (`pkg/model/types.go:70-184`) and `ParseCCList` (`pkg/model/cc.go:8-19`) — see `docs/research/2026-11-07-module-model.md` for the full inventory. The registry tool needs `Profile`/`Profiles`/`ProfilesMu`/`BuiltinProfileIDs` and must not drag the rest.

### 2.2 `pkg/profile/profile.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Reserved(id string) bool` | func | `pkg/profile/profile.go:18-20` | true iff `P1`–`P4`. Duplicated literal set (see §5.3). |
| `Import(p model.Profile) error` | func | `pkg/profile/profile.go:23-46` | Empty `ID` returns `nil` silently (`:24-26`); reserved returns error (`:27-29`); else locks, writes map, read-modify-write of `data/profiles.json` (`:33-45`). No value validation. |
| `Load()` | func | `pkg/profile/profile.go:50-67` | Reads file, merges non-reserved entries (`:61-66`); silent no-op on missing/unparseable file (`:53-60`). No error return. |
| `Delete(id string) error` | func | `pkg/profile/profile.go:71-93` | Reserved → error (`:72-74`); unknown id → error (`:81-85`); deletes from map + file (`:86-92`). |

### 2.3 Tests (pinning behaviour)

- P3 literal values 5/600/30/1 (`pkg/profile/profile_test.go:11-19`); import register + persist + `Load` round-trip (`pkg/profile/profile_test.go:22-44`).
- Reserved refused, ghost-file P3 fake ignored, unknown-id delete refused (`pkg/profile/reserved_test.go:25-53`); delete round-trip no-resurrection (`pkg/profile/reserved_test.go:56-72`).
- Concurrent import/load/read safety must hold (8 goroutines × 25 iterations) (`pkg/profile/profile_concurrent_test.go:15-51`).
- Asymmetry rule incl. silent cap: P2 symmetric, 20/5 asymmetric, 10/50 capped to 10 (`pkg/model/profile_asym_test.go:8-24`).
- API level: CSV import (`pkg/api/profile_import_test.go:16-51`), extended columns `capacity_up_mbps` + burst round-trip through `GET /api/profiles` (`pkg/api/profile_fullcols_test.go:15-52`), P4 not flagged as imported (`pkg/api/csvmapping_test.go:99-126`).

## 3. Dependencies

### 3.1 Outgoing (what profile + model profile types need)

- `profile.go` imports only `encoding/json`, `fmt`, `os`, `path/filepath` plus `github.com/Realms4239/cgo/pkg/model` (`pkg/profile/profile.go:3-10`; the only non-stdlib import is the sibling package, `profile.go:9`). Standard-library paths contain no dot in the first element and resolve locally per R1 §“Module paths”.
- `types.go` imports only `"sync"` (`pkg/model/types.go:5`). No `require` line in `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed by these two packages.
- A module is defined by a `go.mod` file in its root (R1 §“`go.mod` files”). Today there is exactly one `go.mod`, at the root: `module github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1-3`). Neither `pkg/profile` nor `pkg/model` has its own `go.mod` (glob `**/go.mod` returns only the root — same finding as `docs/research/2026-11-07-module-model.md:88`).

### 3.2 Incoming (who needs the registry)

| Consumer | What it uses | Evidence |
|---|---|---|
| `pkg/api/server.go` | `profile.Load` per `GET /api/profiles` and `GET /api/profile/list`; `profile.Reserved`, `profile.Import`, `profile.Delete`; `model.ProfilesMu`, `model.Profiles`, `model.BuiltinProfileIDs` | `server.go:531-558`, `1252-1262`, `1158-1230`, `1233-1251` |
| `pkg/campagne/config.go` | `model.Profiles` under `ProfilesMu.RLock` to resolve matrix profiles into the frozen manifest | `config.go:71-84` |
| `pkg/audit/match.go` | `profile.Load`, `model.Profiles` under `RLock` for `MatchProfile` | `match.go:16-56` |
| `pkg/campagne/matrix.go`, `watch.go` | `ProfilesMu.RLock` + `Profiles[id]` per cell | `matrix.go:179-181`, `watch.go:82-84` |
| `cmd/cgo/calibrate.go` | `ProfilesMu.RLock` + `Profiles[id]` lookup | `calibrate.go:37-39` |
| `cmd/cgo/server.go` | shaping from resolved `r.DelayMs/r.JitterMs/r.LossPct` via `qdisc.ApplyNetem` | `server.go:287-288` |
| Frontend | `profileShort`, tiles, import via `/api/profile/import` | `web/frontend/src/lib/profiles.ts:3-12`, `BenchConfig.tsx:90-94` |

### 3.3 Transitive weight

A consumer importing only the registry still resolves the monolith’s build list, because Go versions a package by its module: “a module is a collection of packages that are released, versioned, and distributed together” (R1 §“Modules, packages, and versions”). The root `require` set (`go.mod:5-10`, GUI/terminal deps) is the module’s regardless of subpackage imported. An independent `go.mod` for the registry removes exactly that weight (R1 §“`go.mod` files”). The nested-module pattern (a `go.mod` in a subdirectory still under one repo) is the standard R1 route for this (R1 §“Module directories within a repository”); whether to nest or split the repo is a human decision (§8).

## 4. Primary-source requirements

For the registry to become a standalone tool under `cgo-profiles`, four primary sources impose concrete, checkable obligations:

**4.1 Go Modules (R1).** The module path is declared by the `module` directive and is “the prefix for package paths within the module” (R1 §“Module paths”). A version “identifies an immutable snapshot of a module” (R1 §“Versions”). Consequence: while the registry lives at `github.com/Realms4239/cgo/pkg/profile` inside the single root module (`go.mod:1-3`), it cannot be versioned independently. Publication requires its own `go.mod` with its own module path (nested or separate repo), which changes the import path — a breaking change under R2 (see below). A replacement directive “only applies in the main module’s `go.mod`” (R1 §“`replace` directive”) and must not appear in a release tag’s consumer examples.

**4.2 Semantic Versioning (R2).** “Software using Semantic Versioning MUST declare a public API … SHOULD be precise and comprehensive” (R2 rule 1). Today the API is declared in code (`§2`) plus a monolith HTTP table (`docs/api.md:38-44`) — there is no library API document. “Major version zero … Anything MAY change at any time” (R2 rule 4); so the honest start is `v0.1.0`. After `v1.0.0`: MAJOR on incompatible change, MINOR on backward-compatible addition (including deprecations), PATCH on compatible fixes (R2 rules 6–8, restated for Go in R1 §“Versions”). Deprecation needs at least one minor cycle before removal (R2 FAQ “How should I handle deprecating functionality?”), which directly governs the `LossBurst*` removal and the `Profiles`/`ProfilesMu` shim in §6.

**4.3 RFC 8290 (R3) — for `fq_codel`, and negatively for profiles.** R3 defines FQ-CoDel as a hybrid of DRR and CoDel shipped “in the mainline Linux kernel as the fq_codel queueing discipline” (R3 §1). The §5.2 parameters are scheduler parameters (`interval`, `target`, `limit`, `quantum`, `flows`) — R3 §5.2.1, §5.2.2, §5.2.5. **R3 contains no link-profile registry, no `capacity_mbps`/`delay_ms`/`jitter_ms`/`loss_pct` schema, and no P1–P4.** Any claim that the registry is “RFC 8290 based” for profiles is therefore false; the RFC governs at most the `FqCodel` label (`pkg/model/types.go:74`), which the registry does not need (it can stay type-only and name the disciplined note). The registry’s `DelayMs` is a netem/regime parameter (`pkg/qdisc/qdisc.go:76-80`), not the CoDel `target` sojourn time (R3 §5.2.2) — docs must state that distinction rather than borrow RFC language. CAKE/PFIFO normative parameters: **unverified — primary source not found** (§0).

**4.4 Prometheus labels (R4).** Every time series is “uniquely identified by its metric name and optional key-value pairs called labels” (R4 §“Metric names and labels”). Consequences the registry must respect:
- Label names “SHOULD match the regex `[a-zA-Z_][a-zA-Z0-9_]*`” and names starting `__` are reserved (R4). A machine-facing label `profile = "P2"` is legal; the free-form imported IDs (`PX-...`, `P-site-fibre` suggested in `pkg/profile/profile.go:28`) are legal *values* but unbounded.
- “The change of any label's value, including adding or removing labels, will create a new time series” (R4). Imported profiles with unbounded IDs therefore create unbounded series — a cardinality hazard. The registry CLI/exporter must ship a bounded label policy (allowlist, or transform `imported` to `builtin`/`imported` enum) instead of raw IDs.
- “Labels with an empty label value are considered equivalent to labels that do not exist” (R4). Encoding `CapacityUpMbps=0` as a label value silently drops the dimension — exporter must use explicit `symmetric=true` or omit the label by design, and document which.

**4.5 Immutability and honesty (R1 + R2).** “Once a versioned package has been released, the contents of that version MUST NOT be modified” (R2 rule 3); retraction, when needed, uses `retract` with rationale (R1 §“`retract` directive”), never a silent rewrite. Registry schema changes obey the same rule: reordering JSON tags or CSV columns is a MAJOR bump with a migration note (R2 rule 8), not a patch.

## 5. Gap analysis (why it is / is not a standalone tool)

| Claim | Verdict | Evidence |
|---|---|---|
| It is a *registry* (records + mutations + persistence). | **Yes, narrowly.** | `Import/Load/Delete/Reserved` (`pkg/profile/profile.go:18-93`) + the four built-ins (`pkg/model/types.go:43-51`) + three read endpoints (`docs/api.md:42-44`). Matches `docs/research/2026-11-07-modular-toolkit-morph.md:45` (“Dynamic profile registry”). |
| It is a standalone *tool* (versioned, documented, runnable alone). | **No.** | No `go.mod` of its own (only root `go.mod:1-3`); no library `README.md`, no `LICENSE`, no schema version, no `cmd/` binary; the only documented surface is the monolith HTTP table (`docs/api.md:38-44`, which omits `POST /api/profile/delete` entirely). Under R1 it has no independent version; under R2 it has no declared library API (R2 rule 1). |
| Its semantics are safe for an external consumer. | **No.** | Silent behaviours: empty-ID import returns `nil` (`pkg/profile/profile.go:24-26` while the API rejects it at `pkg/api/server.go:1217-1220` — library and HTTP disagree); `CapUp` silently caps up>down to down (`pkg/model/types.go:36-41`); CSV numbers parse with errors discarded to `0.0` (`pkg/api/server.go:1181` `num := func…; v, _ := ParseFloat`, so `PX,20,100,15,abc` silently imports loss 0); unreadable JSON file silently ignored (`pkg/profile/profile.go:37-39,53-60`). No `Validate()` exists anywhere in the two packages (grep `func.*Validate` in `pkg/profile` + `pkg/model` returns nothing). The HTTP bounds (capacity 1–1000, delay 0–600, jitter 0–100, loss 0–10, `docs/api.md:51`) are enforced only at the shape endpoint, not at import. |
| It is concurrent-safe as designed. | **Eventually, at a cost.** | The race is real (map write during iteration is fatal Go — pinned by `pkg/profile/profile_concurrent_test.go:15-51`) and the `ProfilesMu` discipline (`pkg/model/types.go:53-57`) does fix it. But `Load` (a write) is called on every `GET /api/profiles` (`pkg/api/server.go:532`) and every `GET /api/profile/list` (`:1253`) and inside `MatchProfile` (`pkg/audit/match.go:20`), and `Import/Delete` hold the write lock across full file read-modify-write (`pkg/profile/profile.go:33-45`, `75-92`) — readers block during disk IO. Safe, but not a publishable concurrency contract. |
| Reserved-reference protection is coherent. | **Brittle duplication.** | The P1–P4 set is hardcoded twice: as map literals (`pkg/model/types.go:43-51`) and as a predicate (`pkg/profile/profile.go:19`). They agree today, but nothing enforces agreement — the architecture doc already drifted to “P1/P2/P3” (`docs/research/2026-11-07-module-model.md:20`), and the production ghost-P3 incident (`pkg/profile/profile.go:14-17`, `pkg/profile/reserved_test.go:38-49`) shows the cost. A registry that can contain a conflicting duplicate set is not an authority. |
| Its file format is stable/versioned. | **No.** | `data/profiles.json` is CWD-relative (`pkg/profile/profile.go:12`), unversioned (no `schema_version` key, no migration), and CWD-dependent enough that every test must `chdir` into `t.TempDir()` first (`pkg/profile/profile_test.go:23-31`, `pkg/profile/reserved_test.go:10-21`, `pkg/api/profile_import_test.go:15-25`). The CSV import is positional: header detection is only `row[0] == "id" || "ID"` on the first row (`pkg/api/server.go:1173-1176`), a missing column silently means a default, and columns 5–9 are optional as a block (`pkg/api/server.go:1182-1212`). |
| Its deprecated fields are contained. | **Leaking.** | `LossBurst*` are “deprecated” in prose (`pkg/model/types.go:18-23`) but are still exported, still imported, still returned by `GET /api/profiles` (`pkg/api/server.go:546-555`; round-trip pinned by `pkg/api/profile_fullcols_test.go:40-51`), while the runtime ignores them — the channel is declared per-campaign (`types.go:18-20`). External consumers inherit zombie fields. Keeping them for one deprecation cycle is correct (R2 FAQ); keeping them forever is not. |
| It is ready for Prometheus exposure. | **No.** | No label policy, no bounded cardinality, no label-name linting (R4). Imported IDs become label values 1:1 and every new ID is a new series (R4: label-value change creates a new time series). There is no `registry → exposition` seam at all. |

**Net judgement:** the registry is the *smallest* CGO surface that is already recognizably a tool — three mutations + one predicate + a persisted map — but today it is an embedded registry inside a single module, with silent failures, duplicated authority, an unversioned CWD-relative file, and zombie fields. The gap is lifecycle and safety (validation, versioning, file injection, bounded exposure), not code size. The RFC 8290 / Prometheus / Go-module obligations in §4 are the acceptance bar, not decoration.

## 6. Target tool: `cgo-profiles` as an independent registry repo

One repo, two deliverables: a **Go library** (`Registry` + schema v1) and a **thin CLI** (`cgo-profiles list|import|export|delete|validate`). Lattice rules from §4: own `go.mod` (R1), `v0.1.0` start (R2 rule 4), public API declared before `v1.0.0` (R2 rule 1), schema changes are MAJOR (R2 rule 8), Prometheus label policy bounded (R4).

| Target export | Source | Change vs today |
|---|---|---|
| `type Profile struct { ID; CapacityMbps; CapacityUpMbps?; DelayMs; JitterMs; LossPct }` + `schema_version` | `pkg/model/types.go:11-17` | Drop `LossBurst*` from the v1 struct; ship `LegacyBurst` separately for the CGO importer only, with a migration note (R2 deprecation cycle for `types.go:18-23`). JSON stays tag-stable; any reorder is MAJOR (R2 rule 8). |
| `func (p Profile) CapDown() / CapUp() float64` (semantics verbatim) + `func (p Profile) Validate() error` | `pkg/model/types.go:31`, `:36-41`; bounds `docs/api.md:51` | Keep the `0 < up < down` rule but return an error for `up > down` instead of silently capping — or keep the cap, document it as contract, and make `Validate` report the clamp. Promote the HTTP bounds (1–1000, 0–600, 0–100, 0–10) into the type; reject zeros/negatives, reject `""` ID (fixing the `profile.go:24-26` silent-nil). |
| `type Registry` with `Get(id) (Profile, bool)`, `Set(p) error`, `Delete(id) error`, `List() []Profile` (sorted), `Reserved(id) bool`, `BuiltinIDs() map[string]bool` | `pkg/model/types.go:43-68`, `pkg/profile/profile.go:18-93` | Replace exported `Profiles` + `ProfilesMu` with an encapsulated type (kills the manual RLock/Lock discipline and the init-frozen `BuiltinProfileIDs`). Single authority for the P1–P4 set — the `types.go` literals and the `profile.go` predicate merge into one `Builtins` slice. Keep a deprecated global shim for one minor cycle (R2) because `pkg/api`, `pkg/campagne`, `pkg/audit`, `cmd/cgo` all write through the globals today (§3.2). |
| `type Store interface { Read() ([]byte, error); Write([]byte) error }` + `FileStore(path)` default, `NewRegistry(store Store, builtins []Profile)` | `pkg/profile/profile.go:12`, `:37-45`, `:53-67`, `:78-92` | Kill the CWD-relative const and the silent-ignore. Path is injected; missing file = empty store (honest); corrupt JSON = returned error, not a no-op. Tests inject an in-memory store instead of `chdir` (§2.3). |
| `func ParseCSV(r io.Reader) ([]Profile, error)` (header-based, name-matched columns) + `func (r *Registry) Import(r io.Reader) (*ImportResult, error)` | `pkg/api/server.go:1167-1216`, `docs/api.md:44` | Replace positional parsing: column names, not positions; `num()` error discarded (`server.go:1181`) becomes a per-line error in `ImportResult` (line, raw value, reason). Header is a first-class schema, not `row[0]=="id"` sniffing. Keep the two accepted shapes (CSV + JSON) named explicitly; reject unknown extra columns instead of ignoring. |
| `func (r *Registry) Export(w io.Writer, format Format) error` (JSON + CSV) | — (currently read-only GETs) | The missing round-trip: `Import` without `Export` is a write-only registry. `Export` is byte-stable for schema v1 so scripts can diff; CSV export is name-headed (inverse of `ParseCSV`). |
| `cgo-profiles` CLI: `list`, `import <file>`, `export`, `delete <id>`, `validate` | `docs/research/2026-11-07-modular-toolkit-morph.md:139` sketch | Thin wrapper over the library; exit non-zero on any line error; prints the per-line report `ImportResult` carries. No HTTP, no shaping, no probing — those stay in CGO. |
| Prometheus helper: `LabelValues() ([]string, error)` returning only bounded sets (`profile ∈ {P1..P4}` or `profile_kind ∈ {builtin, imported}`), plus lint check `ValidateLabelName` | R4 naming/cardinality rules | No raw imported IDs as label values by default (R4: each new value = new series). `CapacityUpMbps=0` exported as explicit `symmetric` boolean, never as an empty label that vanishes (R4 empty-value rule). FQ-CoDel notes: `FqCodel` label documentation cites R3 §1; the registry never claims to implement the algorithm. |

Explicitly **out of** `cgo-profiles`: shaping (`pkg/qdisc`), campaign orchestration (`pkg/campagne`), metrics (`pkg/metrics`), HTTP/SSE (`pkg/api`). R1 keeps the module stdlib-only (no `require`s) if the profile types stay self-contained — the whole point of extracting this surface first.

## 7. Slices

| Slice | Delivers | Depends on | Definition of done (checkable now) |
|---|---|---|---|
| PR-1 — Freeze & document | godoc for every export in §2; decision on module path (nested `go.mod` vs separate repo, §8.1); `README.md` stating “registry + persistence only, no shaping/probing” with the R3/R4 notes | — | `go vet ./pkg/profile/... ./pkg/model/...` + `go test ./pkg/profile/... ./pkg/model/... -count=1` green; README states what it is and is not (§1); open-decision list §8 routed to the human. |
| PR-2 — Extract the library | `go.mod` (module path, `go 1.25` floor from `go.mod:3`), `LICENSE`, `v0.1.0` (R2 rule 4), `Registry` + deprecated `Profiles` shim (§6), `Validate`, `Store` injection, moved tests (`profile_test.go`, `reserved_test.go`, `profile_concurrent_test.go`, `profile_asym_test.go`) | PR-1; human decision on module path + license | `go test ./...` in the new module green without network; each CGO importer (§3.2) still compiles via `require` with no `replace` in the release tag (R1 §“Versions”); shim documented as one-cycle deprecation (R2 FAQ); `Reserved` has exactly one authority. |
| PR-3 — Store + CLI round-trip | `FileStore`, `ParseCSV`/`ImportResult`, `Export`, `cgo-profiles` CLI (§6) | PR-2 | `import → export → import` on a fixture with asymmetric up and burst-legacy fields is byte-stable; malformed CSV (bad number, missing id, reserved id, up>down without cap policy) exits non-zero with a named line; `data/profiles.json` schema has `schema_version: 1` and a migration note for pre-v1 files. |
| PR-4 — Re-point CGO at it | Root `go.mod` requires the new module; `pkg/api`, `pkg/campagne`, `pkg/audit`, `cmd/cgo` import the new path; `docs/api.md:42-44` (+ add `POST /api/profile/delete`) updated | PR-2 (PR-3 for the file format) | Full `make verify` green (all six gates per project policy); `grep "pkg/profile"` shows no stale intra-module import; frozen CSV headers byte-identical (`pkg/model/types.go:171-184`); CWD-chdir tests removed. |
| PR-5 — Harden the contract | Bounded Prometheus label helper + lint, `LegacyBurst` migration path, `SchemaVersion` bump policy, example program (`examples/registry/main.go`) | PR-2 | Example builds `GOFLAGS=-mod=readonly` offline; README documents breaking-change policy (R2 rules 6–8) and retraction policy (R1 `retract`); FQ-CoDel/CAKE caveats read as **unverified — primary source not found** wherever normative text is absent. |

PR-2 on its own is the minimum for “standalone library” status; PR-3 makes it a *tool* (runnable + round-trip). PR-4 is the consumer migration; PR-5 is hardening, not gating.

## 8. Anti-toy criteria

A “toy” here means: code that compiles but cannot be safely relied on by an external consumer. Each gate cites its primary source; failure of A1–A5 means the extraction is a toy regardless of demo quality.

| # | Criterion | Primary source | How `cgo-profiles` passes |
|---|---|---|---|
| A1 | Declared, precise public API (godoc + README) covering every export in §2, so “nobody knows … what methods are safe to call” is impossible | R2 rule 1 (“MUST declare a public API … SHOULD be precise and comprehensive”); R1 §“Versions” (public interface drives MAJOR bumps) | PR-1 godoc covers §2; README states the negative scope (§1) and the R3/R4 caveats (§4). |
| A2 | Independent SemVer: `vMAJOR.MINOR.PATCH`; MAJOR on incompatible change, MINOR on backward-compatible addition (including deprecations), PATCH on compatible fixes; `0.y.z` until `v1.0.0` | R2 rules 4, 6–8; R1 §“Versions” (same three bullets) | PR-2 ships `v0.1.0`; `v1.0.0` only after the `Profiles` shim has a full deprecation cycle (R2 FAQ). |
| A3 | Immutable releases; retraction via `retract` with rationale, never a silent rewrite | R2 rule 3; R1 §“`retract` directive” | Tagged module versions; schema `v1` documented before any `v2` (§6). |
| A4 | Own `go.mod`, zero external `require`s (stdlib-only), reproducibly buildable offline | R1 §“`go.mod` files”; §3.1 (stdlib-only imports incl. `sync` only) | PR-2 `go.mod` requires nothing beyond the Go toolchain; `GOFLAGS=-mod=readonly go build ./...` works with no network. |
| A5 | No exported mutable global state without a documented synchronization contract | Go concurrency evidence in-repo (`pkg/model/types.go:53-57`, writers `pkg/profile/profile.go:33-34,51-52,75-76`; hazard pinned by `pkg/profile/profile_concurrent_test.go:15-51`) | `Registry` encapsulates state (§6); the shim documents the old RLock/Lock discipline for exactly one cycle (R2 FAQ). |
| A6 | Validation at the boundary; silent normalization forbidden or explicit | R2 rule 1 (precise API); in-repo silent-failure evidence (`profile.go:24-26`, `types.go:36-41`, `server.go:1181`) | `Validate` + per-line `ImportResult` errors (§6); any clamp is contract, not surprise. |
| A7 | Normative semantics cited, unknowns labelled; RFC 8290 / Prometheus claims never overreached | R3 §1, §5.2; R4 naming/cardinality; R2 rule 1 | FQ-CoDel label cites R3 §1; no “RFC 8290 link profile” claim exists (R3 contains none — §0, §4.3); CAKE/PFIFO/P4 provenance stay **unverified — primary source not found** until fetched. |
| A8 | Bounded machine-facing exposure (Prometheus) | R4 (label-value change = new series; `__` reserved; empty value = absent) | Default label policy emits `profile ∈ {P1..P4}` or `profile_kind ∈ {builtin, imported}` — never raw unbounded IDs (§6); lint enforces `[a-zA-Z_][a-zA-Z0-9_]*` (R4). |
| A9 | Tests run per-module in CI, not only in the monolith | R1 (module as distribution unit); project gate policy (`make verify`, six gates) | PR-2 moves the §2.3 tests into the new module; PR-4 keeps `make verify` green over the consumer. |
| A10 | License, provenance, and supply-chain records; file format versioned | R1 (module distribution); release signing/SBOM practice is **unverified — primary source not found** in this pass (not fetched) | PR-2 ships `LICENSE` + `CHANGELOG.md` + checksums; `data/profiles.json` gains `schema_version: 1` (§6). Signed attestations/SBOM are explicitly *not* claimed until GoReleaser/SLSA tooling is fetched and cited. |

## Open decisions requiring human approval (not decided here)

1. **Module path and repo shape** — nested `go.mod` inside `github.com/Realms4239/cgo` (R1 §“Module directories within a repository”) vs a separate `github.com/Realms4239/cgo-profiles` repository. Breaking import-path change under R2 rule 8 + R1 major-version-suffix rules; needs a named decision before PR-2.
2. **License** — none present next to `pkg/profile`/`pkg/model` today; needed before any publish (R1 distributes modules from version control; rights must be stated). CGO-wide license claim: **unverified — primary source not found** in this pass.
3. **Cap policy for `up > down`** — keep the silent cap (`pkg/model/types.go:36-41`) as documented contract, or return an error and fix importers. Affects API shape and the CLI exit codes; decide before PR-2 so tests pin one behaviour.
4. **CAKE / PFIFO / `netem` normative pages** — fetch and cite primary sources, or keep the README caveats as **unverified — primary source not found**. Docs only.
5. **P1–P4 value provenance** — decide whether the literals (`pkg/model/types.go:43-51`) are presented as measured link classes, illustrative defaults, or both, and attach evidence or downgrade the claim.

---

*Evidence base: `pkg/profile/profile.go`, `pkg/profile/profile_test.go`, `pkg/profile/profile_concurrent_test.go`, `pkg/profile/reserved_test.go`, `pkg/model/types.go`, `pkg/model/cc.go`, `pkg/model/profile_asym_test.go`, `pkg/model/cc_test.go`, `pkg/api/server.go` (profile handlers), `pkg/campagne/config.go`, `pkg/audit/match.go`, `cmd/cgo/calibrate.go`, `cmd/cgo/server.go`, `docs/api.md:38-44`, `docs/architecture.md`, `go.mod`, plus primary docs R1–R4 fetched 2026-11-07. 404/NOTFETCHED items and unverifiable claims are labelled inline per project rule.*
