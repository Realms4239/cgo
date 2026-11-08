# pkg/model as a standalone open-source tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/model` only, read as if it were extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | Go Modules Reference | https://go.dev/ref/mod | fetched | module vs package, versioning, `go.mod`, nested modules, import rules |
| R2 | Semantic Versioning 2.0.0 | https://semver.org/ | fetched | public-API declaration, MAJOR/MINOR/PATCH rules |
| R3 | RFC 8290 — FQ-CoDel | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | `fq_codel` semantics behind `FqCodel` |
| — | RFC 8289 — CoDel (normative ref of R3) | https://www.rfc-editor.org/info/rfc8289 | **unverified — primary source not found** (not fetched in this pass; cited only as referenced by R3 §2) | CoDel control law for context |
| — | CAKE scheduler documentation | https://www.kernel.org/doc/html/latest/networking/scheduler/sch_cake.html | **unverified — primary source not found** at that path (see `docs/research/2026-11-07-modular-toolkit-morph.md:28`) | `Cake` label provenance |
| — | Penetration / naming of profiles P1–P4 as Madagascar link classes | `docs/methodology.fr.md:59,70`, `pkg/model/types.go:47-50` (Starlink comment) | local source only; external Starlink-pricing claim **unverified — primary source not found** | profile semantics |

Method: read `pkg/model/types.go`, `pkg/model/cc.go`, `pkg/model/cc_test.go`, `pkg/model/profile_asym_test.go`, `docs/architecture.md`, `docs/api.md`, `go.mod`; grep importers of `pkg/model`; fetch R1–R3 above. Every technical claim below cites either `file:line` or `(R#)`. Anything else is labelled **unverified — primary source not found**.

## 1. Responsibility

`pkg/model` today is the CGO **domain-types package**, not an application: `docs/architecture.md:21` defines it as “Types domaine : profils P1/P2/P3, qdiscs, CC, portes G0–G7, événements, en-têtes CSV”. That sentence is now stale on one detail: the source defines four built-in profiles, P1–P4, with P4 documented as Starlink LEO (`pkg/model/types.go:43-51`), while the architecture row still says P1/P2/P3 (`docs/architecture.md:21`).

Concretely, its two source files do three things and nothing more:

- Declares the link-profile value object with an asymmetric-capacity rule: `CapacityMbps` is the historical reference, `CapacityUpMbps` is optional, and absent means up = down (`pkg/model/types.go:7-11`, `pkg/model/types.go:30-41`). The rule is capped so an `up` greater than the reference is incoherent and falls back to the reference (`pkg/model/types.go:33-40`).
- Declares the experiment axes and validity vocabulary: qdisc labels (`pfifo_fast`, `fq_codel`, `cake`) (`pkg/model/types.go:70-78`), congestion-control labels (`cubic`, `bbr`, `bbr3`) (`pkg/model/types.go:80-92`), gate enum G0–G7 (`pkg/model/types.go:95-108`), gate-status strings (`pkg/model/types.go:111-115`), phase names and durations (`pkg/model/types.go:118-126`), the matrix-cell `Event` record (`pkg/model/types.go:128-165`), and the two frozen CSV header orders (`pkg/model/types.go:171-177`, `pkg/model/types.go:180-184`).
- Parses the kernel congestion-control list token-wise, in kernel order, dropping empties and duplicates (`pkg/model/cc.go:5-19`).

What it deliberately does **not** do: it applies no shaping (`docs/architecture.md:43-53` places shaping in `pkg/qdisc`), it runs no probes (`docs/architecture.md:15` places probes in `pkg/probe`), it stores no profiles (`docs/architecture.md:20` places the dynamic registry in `pkg/profile`, which imports `model` at `pkg/profile/profile.go:9`), and it exposes no HTTP surface (`docs/architecture.md:12` places HTTP/SSE in `pkg/api`). That negative scoping is the strongest evidence that `pkg/model` is the “shared kernel” candidate in the modularization sketch: `docs/research/2026-11-07-modular-toolkit-morph.md:43` already classifies it as “Domain types … Yes as Go package (zero external deps beyond stdlib), but same module — no own version”, and `docs/research/2026-11-07-modular-toolkit-morph.md:107` recommends promoting it first because “nothing else moves first”.

## 2. Exported API (file:line)

Complete inventory of exported identifiers in the two source files. Tests are not exported API but are listed last because they pin two behaviours that any extraction must preserve.

### 2.1 `pkg/model/types.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Profile` | struct | `pkg/model/types.go:11-28` | fields `ID`, `CapacityMbps`, `CapacityUpMbps`, `DelayMs`, `JitterMs`, `LossPct`, plus deprecated `LossBurstP/R/H/K` retained only for the documented external CSV import format and the profiles API response (`pkg/model/types.go:18-23`). JSON tags are part of the wire schema (`pkg/model/types.go:12-27`). |
| `Profile.CapDown()` | method | `pkg/model/types.go:31` | returns `CapacityMbps`; the “reference historical” direction (`pkg/model/types.go:30`). |
| `Profile.CapUp()` | method | `pkg/model/types.go:36-41` | returns `CapacityUpMbps` only when `0 < up < down`, else `down` (`pkg/model/types.go:33-40`). |
| `Profiles` | var `map[string]Profile` | `pkg/model/types.go:43-51` | built-ins P1 80/20/2/0, P2 20/100/15/0.5, P3 5/600/30/1, P4 100/40/20/0.3 (Mbps / ms / ms / %). Values are literals in source. |
| `ProfilesMu` | var `sync.RWMutex` | `pkg/model/types.go:53-57` | guards the map: “RLock en lecture, Lock en écriture” (`pkg/model/types.go:53-56`). Exported mutable state — see §5. |
| `BuiltinProfileIDs` | var `map[string]bool` | `pkg/model/types.go:59-68` | frozen at init from `Profiles`; includes P4 so `/api/profiles` does not mislabel it imported (`pkg/model/types.go:59-61`). Exposed via `docs/api.md:42`: `imported` = id non natif (natifs : P1–P4). |
| `Qdisc` | type `string` | `pkg/model/types.go:70` |  |
| `PfifoFast`, `FqCodel`, `Cake` | consts | `pkg/model/types.go:73-76` | values `"pfifo_fast"`, `"fq_codel"`, `"cake"`. `FqCodel` names the Linux `fq_codel` discipline (R3 §1: “included in the mainline Linux kernel as the fq_codel queueing discipline”). `Cake` provenance: **unverified — primary source not found** (§0). |
| `AllQdiscs` | var `[]Qdisc` | `pkg/model/types.go:78` | all three, in declaration order. |
| `CC` | type `string` | `pkg/model/types.go:80` |  |
| `Cubic`, `BBR`, `BBRv3` | consts | `pkg/model/types.go:83-86` | values `"cubic"`, `"bbr"`, `"bbr3"`. Comment: “`bbr` ne veut jamais dire v3” (`pkg/model/types.go:84`). Runtime gating of `bbr3` is outside this package (`docs/api.md:71-73`). |
| `AllCC` | var `[]CC` | `pkg/model/types.go:88` | `{Cubic, BBR}` only — the default matrix axes. |
| `KnownCC` | var `[]CC` | `pkg/model/types.go:90-92` | `{Cubic, BBR, BBRv3}` — accepted labels; `bbr3` stays gated at execution (`pkg/model/types.go:90-91`). |
| `Gate` | type `int` | `pkg/model/types.go:95` |  |
| `G0TargetReachable` … `G7CPUNotSaturated` | consts (iota) | `pkg/model/types.go:97-106` | G0–G7 per `SPEC §2.2`; names map to the gate table in `docs/architecture.md:67-74`. |
| `GateCount` | var `int` | `pkg/model/types.go:108` | value `8`. |
| `GatePass`, `GateDegraded`, `GateInvalid` | consts | `pkg/model/types.go:112-115` | values `"valid"`, `"degraded"`, `"invalid"`; the `gate_status` column of `aqm_eval.csv` (`pkg/model/types.go:110`). |
| `PhaseBaseline`, `PhaseCharge`, `PhaseRecup` | consts | `pkg/model/types.go:119-121` | `"baseline"`, `"charge"`, `"recup"`. |
| `BaselineSec`, `ChargeSec`, `RecupSec` | consts | `pkg/model/types.go:123-125` | `30`, `120`, `30`. |
| `Event` | struct | `pkg/model/types.go:128-165` | one matrix-cell execution. Identity fields (`RunID`, `EventID`, `Profile`, `Qdisc`, `CC`, `Direction`, `Repetition`) carry `json` tags; measured fields carry both `csv` and `json` tags (`pkg/model/types.go:141-156`). Two JSON-only fields exist because the 23-column CSV schema is frozen: `Bytes` (`pkg/model/types.go:157-161`) and `FailedGates` (`pkg/model/types.go:162-164`). `Direction` is frozen into the row so up/down cells cannot merge on scan (`pkg/model/types.go:135-137`). |
| `AQMEvalHeader` | var `[]string` | `pkg/model/types.go:167-177` | 23 columns after identity; order is historical, readers consume by name (`pkg/model/types.go:168`). Includes `rtt_base_p50_ms`, `rtt_base_p95_ms`, `gate_status`. |
| `LinkAuditHeader` | var `[]string` | `pkg/model/types.go:179-184` | `audit_id`, `timestamp`, `site`, `link_type`, `provider`, RTT idle/loaded, throughput, loss, `http_small_p95_ms`, `data_used_mb`, `notes`. |

### 2.2 `pkg/model/cc.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ParseCCList(raw string) []string` | func | `pkg/model/cc.go:8-19` | tokenizes `/proc/sys/net/ipv4/tcp_available_congestion_control`, preserves kernel order, discards empties and duplicates. The path is the input contract, documented in the comment (`pkg/model/cc.go:5-6`) and surfaced in `docs/api.md:71-72` for the `bbr3` gate. |

### 2.3 Tests (pinning behaviour, not API)

- `TestParseCCList` covers `"reno cubic bbr\n"`, `"  bbr3   bbr bbr3 "`, and `""` (`pkg/model/cc_test.go:8-22`).
- `TestBBRv3IsKnownButNotDefault` asserts `KnownCC` contains `BBRv3` and `AllCC` does not (`pkg/model/cc_test.go:24-39`).
- `TestProfileAsymmetricCapacity` asserts P2 is symmetric, an explicit 20/5 profile is asymmetric, and a 10/50 profile is capped to 10 (`pkg/model/profile_asym_test.go:8-24`).

## 3. Dependencies

### 3.1 Outgoing (what `pkg/model` needs)

- `types.go` imports only `"sync"` (`pkg/model/types.go:5`). `cc.go` imports only `"strings"` (`pkg/model/cc.go:3`). Both are standard-library packages. Per the Go Modules Reference, standard-library package paths contain no dot in the first element and are resolved locally, not fetched (R1, §“Module paths”). No `require` line in `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed by this package.
- The module as a whole is `github.com/Realms4239/cgo` at `go 1.25` (`go.mod:1-3`). `pkg/model` inherits that module path and that Go version, but adds no requirement of its own.

### 3.2 Incoming (who needs `pkg/model`)

The package path `github.com/Realms4239/cgo/pkg/model` is imported across the repository:

- Registry mutation and persistence: `pkg/profile/profile.go:9` (`"github.com/Realms4239/cgo/pkg/model"`), with `Import`, `Load`, and `Delete` all writing through `model.ProfilesMu` (`pkg/profile/profile.go:33-34`, `pkg/profile/profile.go:51-52`, `pkg/profile/profile.go:75-76`).
- Campaign orchestration, shaping, probing, audit, doctor, figures, and the HTTP surface all import the same path (grep for `pkg/model` shows hits in `pkg/api`, `pkg/campagne`, `pkg/qdisc`, `pkg/audit`, `pkg/doctor`, `pkg/figures`, `pkg/probe`, and `cmd/cgo`).
- The HTTP contract directly serializes these types: `/api/profiles` returns `model.Profile` fields including the `?`-optional ones (`docs/api.md:42`), and `/api/profile/import` accepts either CSV with documented positions or a `model.Profile` JSON body (`docs/api.md:44`).

There is no file `pkg/model/go.mod` in the repository (only `go.mod` at the root; see glob for `**/go.mod`), so there is no independently versioned module for this package.

### 3.3 Transitive weight

A consumer that imports only `pkg/model` today still resolves the whole monolith’s build list, because Go versions a package by its module, and “a module is a collection of packages that are released, versioned, and distributed together” (R1, §“Modules, packages, and versions”). The direct requirements in `go.mod:5-10` (Fyne, Bubble Tea, Lip Gloss, `x/sys`) are mostly GUI/terminal concerns, but they are the module’s requirements regardless of which subpackage a consumer imports. That is precisely the weight a separate `go.mod` for `model` would remove (R1, §“`go.mod` files”).

## 4. Primary-source requirements

For `pkg/model` to become a standalone open-source tool, three primary sources impose concrete, checkable obligations:

1. **Go Modules Reference (R1).** A module path is declared by the `module` directive in `go.mod`, and the module path is “the prefix for package paths within the module” (R1, §“Module paths”). A version is an immutable snapshot identified by `v` + Semantic Versioning (R1, §“Versions”). A package path is the module path joined with the subdirectory, “for example … `golang.org/x/net/html`” (R1, §“Modules, packages, and versions”). Consequence: `pkg/model` cannot be versioned on its own while it lives inside `github.com/Realms4239/cgo`; it needs its own `go.mod` with its own module path (the standard pattern is a nested module directory, covered by R1 §“Module directories within a repository”), or a separate repository with a new module path. Either way the import path of the package changes, which is a breaking change under R2 (see below).
2. **Semantic Versioning 2.0.0 (R2).** “Software using Semantic Versioning MUST declare a public API. This API could be declared in the code itself or exist strictly in documentation. However it is done, it SHOULD be precise and comprehensive” (R2, rule 1). The current public API is declared in code (the exports inventoried in §2) plus a monolith `docs/api.md` that describes the HTTP surface, not a library. Before assigning a `1.0.0` to a `model` module, the API must be documented as such; until then the honest starting version under R2 is `0.y.z` (“initial development … Anything MAY change at any time”, R2 rule 4). Incompatible changes bump MAJOR, backward-compatible additions bump MINOR, backward-compatible bug fixes bump PATCH (R2 rules 6–8; R1 §“Versions” restates the same three rules for Go).
3. **RFC 8290 (R3) for the `fq_codel` label.** R3 defines FQ-CoDel as a hybrid of DRR and the CoDel AQM, shipped in the Linux kernel as the `fq_codel` qdisc (R3 §1). The label `FqCodel = "fq_codel"` (`pkg/model/types.go:74`) therefore carries a defined meaning: per-flow queueing with CoDel AQM on each queue (R3 §1.3). Any standalone documentation for the `model` module must describe `FqCodel` in those terms and must not imply the package implements the algorithm — the package only names it; implementation lives in `pkg/qdisc` per `docs/architecture.md:14`. For `Cake` and `pfifo_fast`, the corresponding normative parameters (CAKE) are **unverified — primary source not found** (§0); the standalone README must say so rather than invent parameters. `pfifo_fast` itself: **unverified — primary source not found** in the fetched set.

## 5. Gap analysis (why it is / is not a tool)

| Claim | Verdict | Evidence |
|---|---|---|
| It is a *library* (importable value objects + one parser). | **Yes, narrowly.** | Zero non-stdlib imports (`pkg/model/types.go:5`, `pkg/model/cc.go:3`); exports are plain structs, string types, consts, and one pure function (`pkg/model/cc.go:8-19`). `docs/research/2026-11-07-modular-toolkit-morph.md:43` records the same judgement. |
| It is a standalone *tool* (versioned, documented, runnable on its own). | **No.** | No `go.mod` of its own (only root `go.mod:1-3`); no `README.md`, no `LICENSE`, no `CHANGELOG`, no example `main`; no binary under `cmd/` (`docs/architecture.md:10-21` lists binaries only in `cmd/cgo`). Under R1 it has no independent version; under R2 it has no declared public API document for library consumers. |
| It is safe to import as-is from outside CGO. | **No, not yet.** | Three concrete hazards: (a) exported mutable global state — `Profiles` (`pkg/model/types.go:43`) plus `ProfilesMu` (`pkg/model/types.go:57`) invite data races and cross-test pollution; (b) deprecated struct fields retained for internal CSV compatibility (`pkg/model/types.go:18-27`) leak into the public JSON schema for any external consumer; (c) `BuiltinProfileIDs` is computed once at init from a mutable map, so an importer that mutates `Profiles` before init-order assumptions break gets a stale registry (`pkg/model/types.go:62-68`). |
| Its semantics are self-contained and normatively grounded. | **Partly.** | `FqCodel` is grounded in R3 (§1, §1.3); `ParseCCList` is grounded in the kernel file path (`pkg/model/cc.go:5-6`, `docs/api.md:71-72`). Profile *values* (P1–P4 literals at `pkg/model/types.go:43-51`) are presented as measured link classes (`docs/methodology.fr.md:59`) with no attached measurement provenance in this package — an external consumer cannot tell which numbers are defaults, which are measurements, and which are illustrative. That distinction must be made explicit in extracted docs. |
| Its CSV headers make it wire-compatible with frozen evidence. | **Yes, for CGO; brittle for outsiders.** | `AQMEvalHeader` order is historical and readers use names (`pkg/model/types.go:167-168`); `Event` deliberately keeps JSON-only fields out of the 23-column CSV (`pkg/model/types.go:157-164`). The schema is frozen by policy, not by a versioned contract — an external tool has no semver protection against a reordering that breaks historical readers. |
| It has no dependencies to extract. | **Yes for code; no for lifecycle.** | Code deps are stdlib-only (§3.1). Lifecycle deps (module identity, release process, docs) do not exist yet (§4). |

**Net judgement:** `pkg/model` is the *closest* CGO package to a standalone tool — pure types, pure parser, stdlib-only — but today it is a shared kernel inside a single module, not a published tool. The gap is lifecycle and hygiene (versioning, docs, global state, deprecation residue), not code weight.

## 6. Target module API surface

Proposed extracted module: `model` (module path to be decided by the human — see §8.4; the sketch in `docs/research/2026-11-07-modular-toolkit-morph.md:107` suggests `github.com/Realms4239/cgo/model`). The target below keeps every current CGO consumer working while closing the §5 hazards. Each line cites the current source it derives from.

| Target export | Source | Change vs today |
|---|---|---|
| `type Profile struct { ID, CapacityMbps, CapacityUpMbps, DelayMs, JitterMs, LossPct }` | `pkg/model/types.go:11-17` | **Drop** the deprecated `LossBurstP/R/H/K` fields from the standalone struct; expose a separate `LegacyBurst` type only for the CGO CSV importer, with a migration note (`pkg/model/types.go:18-27`). |
| `func (p Profile) CapDown() float64`, `func (p Profile) CapUp() float64` | `pkg/model/types.go:31`, `pkg/model/types.go:36-41` | Keep semantics exactly; add a `Validate() error` method that rejects negative/zero capacities and out-of-range delay/jitter/loss (bounds already enforced at the HTTP edge in `docs/api.md:51`: capacity 1–1000, delay 0–600, jitter 0–100, loss 0–10 — promote them into the type). |
| `func ParseCCList(raw string) []string` | `pkg/model/cc.go:8-19` | Keep signature and semantics (kernel order, dedupe). Document the input-file contract (`pkg/model/cc.go:5-6`) in godoc. |
| `type Qdisc string; const PfifoFast, FqCodel, Cake` + `AllQdiscs` | `pkg/model/types.go:70-78` | Unchanged, with godoc for `FqCodel` citing R3 §1 (“mainline Linux kernel as the fq_codel queueing discipline”) and an explicit “label only — algorithm lives in the shaping layer” note. CAKE/PFIFO parameters marked **unverified — primary source not found** until a normative page is fetched and cited. |
| `type CC string; const Cubic, BBR, BBRv3` + `AllCC`, `KnownCC` | `pkg/model/types.go:80-92` | Unchanged; keep the `AllCC` (defaults) vs `KnownCC` (accepted labels) distinction that `TestBBRv3IsKnownButNotDefault` pins (`pkg/model/cc_test.go:24-39`). |
| `type Gate int; const G0…G7` + `GateCount` + `GatePass/Degraded/Invalid` | `pkg/model/types.go:95-115` | Unchanged; map G0–G7 to the `docs/architecture.md:67-74` table in godoc so the gate semantics travel with the type. |
| `PhaseBaseline/Charge/Recup`, `BaselineSec/ChargeSec/RecupSec` | `pkg/model/types.go:118-126` | Unchanged. |
| `type Event struct` | `pkg/model/types.go:128-165` | Unchanged shape; document the frozen-23-column invariant (`pkg/model/types.go:157-164`) and the `Direction` freeze rationale (`pkg/model/types.go:135-137`) as contract, not comment. |
| `AQMEvalHeader`, `LinkAuditHeader` | `pkg/model/types.go:171-177`, `pkg/model/types.go:180-184` | Unchanged content; add `SchemaVersion` const and a `ColumnIndex(name string) (int, error)` helper so external readers bind by name without copying the slice. |
| Registry | `pkg/model/types.go:43-68` | **Replace** exported `Profiles` + `ProfilesMu` with a small `Registry` type (`Get`, `Set`, `List`, `BuiltinIDs`) and a `Default() *Registry` holding P1–P4. Keep a deprecated `Profiles` shim for one release (R2 rule 7: deprecation → MINOR bump with at least one minor cycle before removal) because `pkg/profile/profile.go:33-34,51-52,75-76` and all matrix writers depend on the current globals. Lovely-to-have but optional in slice 1: copy-on-read accessors to kill the RLock/Lock discipline burden. |
| `BuiltinProfileIDs` | `pkg/model/types.go:62-68` | Becomes `Registry.BuiltinIDs()`; the `/api/profiles` `imported` semantics (`docs/api.md:42`) stay identical. |

Explicitly **out of** the `model` module: profile persistence (`pkg/profile`), shaping (`pkg/qdisc`), metrics (`pkg/metrics`), HTTP (`pkg/api`), campaign orchestration (`pkg/campagne`). Under R1, keeping the extracted module stdlib-only keeps its `require` set at zero, which is the whole point of extracting it first.

## 7. Slices

| Slice | Delivers | Depends on | Definition of done (each step now, not “later”) |
|---|---|---|---|
| M-1 — Freeze & document | godoc for every export in §2; decision on module path (nested vs separate repo, see §8.4); `README.md` stating “types + parser only, no shaping/probing” with the R3 citation for `FqCodel` | — | `go vet ./pkg/model/...` + `go test ./pkg/model/... -count=1` green; godoc renders; README states what it is and is not (§1). |
| M-2 — Extract the module | `go.mod` (module path, `go 1.25` floor from `go.mod:3`), `LICENSE`, SemVer `v0.1.0` (R2 rule 4: initial development), `Registry` type + deprecated `Profiles` shim (§6), `Profile.Validate`, `ColumnIndex`, moved tests (`cc_test.go:8-39`, `profile_asym_test.go:8-24`) | M-1; human decision on module path + license | `go test ./...` in the new module green; each CGO importer (`pkg/profile/profile.go:9` et al.) still compiles against the new module via `require` (no `replace` in the release tag, R1 §“Versions”); the shim is documented as one-cycle deprecation (R2 rule 7). |
| M-3 — Re-point CGO at it | Root `go.mod` requires the `model` module; `pkg/profile`, `pkg/campagne`, `pkg/qdisc`, `pkg/api`, `pkg/audit`, `pkg/doctor` import the new path; `docs/architecture.md:21` and `docs/api.md:42-44` updated to the new import path | M-2 | Full `make verify` green (all six gates per project policy); `grep pkg/model` shows no stale intra-module import; frozen CSV schemas untouched (headers byte-identical to `pkg/model/types.go:171-184`). |
| M-4 — Harden the contract (optional, after M-3) | `LegacyBurst` migration path for the deprecated fields; schema-version bump policy for `AQMEvalHeader`/`LinkAuditHeader`; example program (`examples/profiles/main.go`) demonstrating `Registry` + `ParseCCList` without CGO | M-2 | Example builds with `GOFLAGS=-mod=readonly` and no network; README documents the breaking-change policy (R2 rules 6–8). |

M-2 and M-3 are the minimum for “standalone tool” status; M-1 is required because R2 rule 1 demands a declared public API before any version claim; M-4 is hardening, not gating.

## 8. Anti-toy criteria

A “toy” here means: code that compiles but cannot be safely relied on by an external consumer. The following gates distinguish a real tool from a toy; each cites its primary source.

| # | Criterion | Primary source | How `model` passes |
|---|---|---|---|
| A1 | Declared, precise public API (godoc + README), complete enough that “nobody knows … what methods are safe to call” is impossible (R2 FAQ) | R2 rule 1 (“MUST declare a public API … SHOULD be precise and comprehensive”); R1 §“Versions” (public interface drives MAJOR bumps) | M-1 godoc covers every export in §2; README states the negative scope (§1). |
| A2 | Independent SemVer: `vMAJOR.MINOR.PATCH`, MAJOR on incompatible change, MINOR on backward-compatible addition (including deprecations), PATCH on compatible fixes | R2 rules 6–8; R1 §“Versions” (same three bullets) | M-2 ships `v0.1.0` (R2 rule 4: unstable, anything may change) and promotes to `v1.0.0` only when the registry redesign (§6) has one deprecation cycle behind it. |
| A3 | Immutable releases: contents of a released version MUST NOT be modified | R2 rule 3 | Releases are tagged module versions (R1: “version identifies an immutable snapshot”); retractions, if ever needed, use the `retract` directive (R1 §“`retract` directive”), never a silent rewrite. |
| A4 | Own `go.mod`, zero external `require`s, reproducibly buildable offline | R1 §“`go.mod` files” (module defined by `go.mod` in its root); §3.1 (stdlib-only imports) | Extracted `go.mod` requires nothing beyond the Go toolchain; `GOFLAGS=-mod=readonly go build ./...` works with no network. |
| A5 | No exported mutable global state without a documented synchronization contract | Go memory / concurrency discipline (no dedicated primary fetched — the *current* hazard is evidenced locally at `pkg/model/types.go:53-57` and its writers at `pkg/profile/profile.go:33-34,51-52,75-76`) | `Registry` replaces raw `Profiles`/`ProfilesMu` (§6); shim documents the old discipline for one cycle. |
| A6 | Normative semantics cited, unknowns labelled | R3 §1 (FQ-CoDel origin), §1.3 (informal summary); R2 rule 1 (precise API) | `FqCodel` godoc cites R3; CAKE/PFIFO and profile-value provenance marked **unverified — primary source not found** until fetched (§4.3). |
| A7 | Tests run per-module in CI, not only in the monolith | Project gate policy (`make verify`, six gates) vs R1 (module as distribution unit) | M-2 adds module-local `go test ./...`; the moved tests are `pkg/model/cc_test.go:8-39` and `pkg/model/profile_asym_test.go:8-24`. |
| A8 | License, provenance, and supply-chain records | Go Modules Reference (module distribution, R1 §“Modules, packages, and versions”); release signing/SBOM practice is **unverified — primary source not found** in this pass (not fetched) | M-2 ships `LICENSE`, `CHANGELOG.md`, and per-release checksums. Signed attestations/SBOM are explicitly *not* claimed until GoReleaser/SLSA tooling is fetched and cited. |
| A9 | Frozen schemas are versioned, not folklore | R2 (public API) + local freeze evidence (`pkg/model/types.go:157-164`, `pkg/model/types.go:167-168`) | `SchemaVersion` + `ColumnIndex` (§6); any header reorder is a MAJOR bump with a migration note (R2 rule 8). |

Failure of any of A1–A5 means the extraction is a toy, regardless of demo quality.

## Open decisions requiring human approval (not decided here)

1. **Module path and repo shape** — nested `go.mod` inside `github.com/Realms4239/cgo` vs a separate `github.com/Realms4239/cgo-model` repository. This is a breaking import-path change under R2 rule 8 and R1 major-version-suffix rules; it needs a named decision before M-2 (see `docs/research/2026-11-07-modular-toolkit-morph.md:176`).
2. **License** — none present next to `pkg/model` today; the extracted module needs an explicit license before publishing (R1 distributes modules from version control; rights must be stated). **Unverified — primary source not found** for any CGO-wide license claim in this pass.
3. **CAKE / PFIFO normative pages** — fetch and cite the primary CAKE and FIFO documentation, or keep the README caveat as **unverified — primary source not found**. Affects only docs, not code.
4. **Profile-value provenance** — decide whether P1–P4 literals (`pkg/model/types.go:43-51`) are presented as measured link classes, illustrative defaults, or both, and attach the measurement evidence or downgrade the claim.

---

*Evidence base: `pkg/model/types.go`, `pkg/model/cc.go`, `pkg/model/cc_test.go`, `pkg/model/profile_asym_test.go`, `pkg/profile/profile.go`, `docs/architecture.md`, `docs/api.md`, `go.mod`, plus primary docs R1–R3 fetched 2026-11-07. 404/NOTFETCHED items and unverifiable claims are labelled inline per project rule.*
