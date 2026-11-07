# pkg/results + pkg/figures as standalone tools — primary-source research

*Date: 2026-11-07 · Scope: `pkg/results` + `pkg/figures` only, read as if extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — FQ-CoDel | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | What RFC 8290 does/does not normatively define (fq_codel defaults, CAKE status) |
| R2 | Prometheus exposition formats | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched | Text-format contract a results exporter must meet |
| R3 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | fetched | Module path/versioning, `go` directive, extraction mechanics |
| R4 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched | Public-API declaration, MAJOR/MINOR/PATCH rules |
| R5 | Go `encoding/csv` package docs | https://pkg.go.dev/encoding/csv | fetched | `FieldsPerRecord=-1` semantics, RFC 4180 reader |
| R6 | Go `embed` package docs | https://pkg.go.dev/embed | fetched | `//go:embed` pattern rules for `SeedRuns` |
| — | CAKE scheduler specification (beyond RFC 8290 informative ref) | — | **unverified — primary source not found** | `cake` label / `HardwareRecommendation` CAKE branch |
| — | MikroTik RouterOS Queue Tree / PCQ translation | — | **unverified — primary source not found** | `HardwareRecommendation` MikroTik branch |
| — | Bootstrap/median statistical-validity references (for `BootstrapMedianCI95` call) | — | **unverified — primary source not found** | CI95 validity beyond local seed-42 code |

Method: read all `pkg/results/*.go` (`csvrows.go`, `hardware.go`, `results.go`, `seed.go`) and `pkg/figures/figures.go` (not tests except to note pinned behaviour), `docs/architecture.md:17-18,76-83`, `docs/VERIFY.md:24-26,51`, `go.mod:1-10`, plus callers in `pkg/api/server.go`, `pkg/api/translate.go`, `pkg/api/duel.go`, `cmd/cgo/main.go`, `cmd/cgo/runcli.go`, `cmd/cgo/server.go`, and grep `results\.` / `figures\.` / `syscall` across `pkg/results` + `pkg/figures`. Every technical claim below cites `file:line` or `(R#)`. Anything else is labelled **unverified — primary source not found**.

## 1. Responsibility

Today `pkg/results` is the CGO **frozen-corpus aggregation core**: `docs/architecture.md:17` describes it as "`Scan` de `data/runs` → médianes par cellule, `Quarantined`, `Best`", and `pkg/results/results.go:49-65` states its contract — aggregate every `aqm_eval.csv` under a data directory, group by `profile|qdisc|cc|direction`, compute medians. Concretely across its four files it does five things:

- **Header-name CSV reading of frozen `aqm_eval.csv`** tolerating 17/18/19-column historic layouts (`pkg/results/csvrows.go:9-45`); never positional (`pkg/results/csvrows.go:11-15` explains the `qdi_ms` insertion bug that caused QDI to be labelled `small_p95`).
- **Cell aggregation with quarantine honesty** — `invalid` rows counted as `quarantined` and excluded from medians, `degraded` counted separately but included in all-rows medians, `valid`-only medians/CIs kept alongside (`pkg/results/results.go:137-162,175-186`; `docs/VERIFY.md:51` pins `valid:208, degraded:74, quarantined:246` over the frozen corpus).
- **Provenance of contributing runs** — `Group.Runs` lists only runs with at least one non-`invalid` row, sorted (`pkg/results/results.go:43-46,146-148,53-63`); a fully-quarantined run cannot cite a median it did not feed (`pkg/results/results_test.go:30-54`).
- **Best-per-profile + hardware recommendation** — smallest `Smallp95Median` per profile marked `Best`, then a French MikroTik/ISP translation string attached to every group in that profile (`pkg/results/results.go:207-234`, `pkg/results/hardware.go:7-12`).
- **Embedded corpus seeding** — extracts `runs-frozen.tar.gz` into `<root>/data/runs` only when no `aqm_eval.csv` exists, with anti-traversal filtering (`pkg/results/seed.go:15-79`).

`pkg/figures` is its **SVG renderer + provenance hasher**: `docs/architecture.md:18` describes it as "Génération SVG (barres `small_p95.svg`, nuage `scatter.svg`) + `ProvenanceSHA` / `ProvenanceHash8`", and `pkg/figures/figures.go:15-37` states it walks runs via `results.Scan` then emits two fixed-size SVGs with RDF provenance metadata. `docs/architecture.md:76-83` frames the pair: freeze writes a SHA-256 manifest, `pkg/figures` computes `ProvenanceSHA` (sha256 of the last `aqm_eval.csv`) and `ProvenanceHash8` (first 8 hex) injected into SVGs, exported report and `/api/integrity`, so "toute figure ou tableau porte l'empreinte exacte des données qui l'ont produit".

What they deliberately do **not** do: no shaping, no probing, no HTTP, no Prometheus exposition, no live measurement — shaping is `pkg/qdisc`, probes are `pkg/probe` (`docs/architecture.md:14-15`), and the only exposition builder in-repo is hand-rolled in `pkg/api/metrics.go:17-41`. That negative scoping is what makes extraction plausible in principle: both packages are filesystem-in/value-out (plus one embed), and stdlib-only except for the internal chain `metrics → results → figures` (§3), so they are leaves of the call graph, not hubs — except that `pkg/api` calls `results.Scan` at least seven times per request path (`pkg/api/server.go:770,942,982,1004`, `pkg/api/duel.go:77`, `pkg/api/translate.go:25`) and `figures.Generate` on `POST /api/figures/regen` (`pkg/api/server.go:1084-1090`) and CLI `cgo figures` (`cmd/cgo/main.go:122-127`).

The frozen corpus they aggregate is **evidence, not scratch space** per `AGENTS.md` ("`book/data/runs` and the frozen corpus: no new run, rewrite, or reshape"), and its verified shape is **159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells** (`docs/VERIFY.md:24`), sealed by per-run manifests plus a corpus-wide anchor `docs/frozen-manifests.sha256` over all 159 manifests (`docs/VERIFY.md:26`). `Scan`'s job is to reproduce that shape faithfully; any standalone module that changes counting rules silently re-writes history.

## 2. Exported API (file:line)

Complete inventory of exported identifiers in the five source files. Unexported helpers (`runIDs` `pkg/results/results.go:53-63`, `provenanceMeta` `pkg/figures/figures.go:63-79`, `writeBar` `pkg/figures/figures.go:81-116`, `writeScatter` `pkg/figures/figures.go:118-156`, `runsFrozen` `pkg/results/seed.go:16`) are not API.

### 2.1 `pkg/results/csvrows.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `ReadAQM(path)` | func | `pkg/results/csvrows.go:19-45` | CSV reader with `FieldsPerRecord=-1` (`pkg/results/csvrows.go:26`); header = first record, empty/blank rows skipped (`pkg/results/csvrows.go:35-43`). Err on open/parse; short rows kept, whole file never dropped for one bad row (`pkg/results/csvrows.go:16-18`). |
| `ColIndex(header, name)` | func | `pkg/results/csvrows.go:49-56` | Trimmed exact-match header lookup, `-1` if absent (`pkg/results/csvrows.go:47-48`). |
| `Field(header, row, name)` | func | `pkg/results/csvrows.go:59-65` | Named field of one row, `""` if absent/out-of-range (`pkg/results/csvrows.go:58`). |

### 2.2 `pkg/results/results.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Group` | struct | `pkg/results/results.go:13-47` | Cell `profile|qdisc|cc|direction` (`pkg/results/results.go:131`), `Count/Quarantined/Degraded` (`pkg/results/results.go:20-22`), all-rows medians+IQRs for RTT/small/goodput/deadline/wasted/cost (`pkg/results/results.go:24-31`), valid-only `Smallp95*` + `Deadline*` with N/median/IQR/CI95 (`pkg/results/results.go:34-40`), `Best` (`pkg/results/results.go:41`), `HardwareRecommendation` (`pkg/results/results.go:42`), sorted `Runs` provenance (`pkg/results/results.go:46`). JSON tags on every field. |
| `Scan(dataDir, runFilter)` | func | `pkg/results/results.go:65-236` | Glob `dataDir/*/aqm_eval.csv` (or single `runFilter`) (`pkg/results/results.go:66-70`); empty → `nil,nil` (`pkg/results/results.go:71-73`). Skips `run-smoke*` (`pkg/results/results.go:88-90`), skips unreadable files / missing `profile/qdisc/cc` headers silently (`pkg/results/results.go:93-100`). Direction defaults `"up"` for historic rows without the column, up/down never merge (`pkg/results/results.go:125-131`), sorted output with `up`-before-`down` tiebreak (`pkg/results/results.go:188-206`). Best = min `Smallp95Median` per profile, `>0` only (`pkg/results/results.go:208-220`). |

### 2.3 `pkg/results/hardware.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `HardwareRecommendation(bestQdisc, profile)` | func | `pkg/results/hardware.go:7-12` | `fq_codel`/`cake` → `"Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour %s — %s prouvé en lab"`; else `"Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour %s — pilote isolé d'abord"`. Hardcoded French operator string is part of the returned value (§5). CAKE/MikroTik provenance: **unverified — primary source not found**. |

### 2.4 `pkg/results/seed.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `SeedRuns(root)` | func | `pkg/results/seed.go:24-79` | No-op `(false,0,nil)` if any `aqm_eval.csv` already exists (`pkg/results/seed.go:26-28`). Else extracts embedded `runs-frozen.tar.gz` (`pkg/results/seed.go:15-16`) accepting only regular files under `data/runs/` with no `..`/absolute/`..`-containing components and inside `runsDir` (`pkg/results/seed.go:43-54`), `O_CREATE|O_EXCL` per file (`pkg/results/seed.go:58-64`); `n==0` → error (`pkg/results/seed.go:75-77`). Per R6, `//go:embed` variable must be `string`/`[]byte`/`FS` at package scope with patterns relative to the package dir (`pkg/results/seed.go:15` is `[]byte`, R6). |

### 2.5 `pkg/figures/figures.go`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Generate(dataDir, outDir)` | func | `pkg/figures/figures.go:17-37` | `results.Scan(dataDir,"")`; `len==0` → `fmt.Errorf("no data")` (`pkg/figures/figures.go:22-24`). Writes `small_p95.svg` (bars) and `scatter.svg` (goodput×small) into `outDir` (`pkg/figures/figures.go:28-35`). |
| `ProvenanceSHA(dataDir)` | func | `pkg/figures/figures.go:41-53` | SHA-256 hex of the **lexically last** `aqm_eval.csv` (`pkg/figures/figures.go:42-47`); `""` if none/unreadable. Not a corpus hash (§5). |
| `ProvenanceHash8(dataDir)` | func | `pkg/figures/figures.go:56-61` | First 8 hex of `ProvenanceSHA`; `""` if empty (`pkg/figures/figures.go:55`). |

### 2.6 Tests (pin behaviour, not API)

`pkg/results`: `TestScanReadsByName` (positional-read regression, `pkg/results/results_test.go:70-121`), `TestScanMixedLayouts` (17/18/19 merge, `pkg/results/results_test.go:126-160`), `TestGroupRunsExcludeInvalidOnlyRuns` (`pkg/results/results_test.go:33-54`), `TestValidOnlyFields` (`pkg/results/validonly_test.go:7-40`), `TestScanSkipsSmokeFixture` + direction split/default (`pkg/results/direction_test.go:9-78`), `TestScanDirectionOrderStable` 30-iteration (map-randomness regression, `pkg/results/direction_order_test.go:13-33`), `TestHardwareSingle` (`pkg/results/hardware_test.go:9-22`), `TestSeedRunsExtracts/NoOp/RejectsEvil` (`pkg/results/seed_test.go:12-96`). `pkg/figures`: `TestFiguresRDF` (honest error + RDF in both SVGs, `pkg/figures/figures_test.go:16-51`), `TestProvenanceHash8` (`pkg/figures/figures_test.go:55-73`).

## 3. Dependencies

### 3.1 Outgoing (what these packages need)

- `pkg/results/results.go:3-10` imports stdlib `path/filepath`, `sort`, `strconv`, `strings` plus internal `github.com/Realms4239/cgo/pkg/metrics`. `pkg/results/csvrows.go:3-7` imports stdlib `encoding/csv`, `os`, `strings`. `pkg/results/hardware.go:3` imports stdlib `fmt`. `pkg/results/seed.go:3-13` imports stdlib `archive/tar`, `bytes`, `compress/gzip`, `embed`, `fmt`, `io`, `os`, `path/filepath`, `strings`. No third-party import appears in any of the four files; per R3 stdlib import paths contain no dot in the first element and resolve locally, not by download (R3, §"Module paths"), and none of the `require` lines in `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed by `pkg/results`.
- `pkg/figures/figures.go:3-13` imports stdlib `crypto/sha256`, `encoding/hex`, `fmt`, `os`, `path/filepath`, `sort`, `time` plus internal `github.com/Realms4239/cgo/pkg/results`. Again no third-party import; none of `go.mod:5-10` is needed by `pkg/figures`.
- **No `syscall` (or `golang.org/x/sys`) import in either package**: grep `syscall|prometheus|client_golang|exposition` over both directories returns zero hits, and the import blocks above list no syscall package. `golang.org/x/sys v0.36.0` in `go.mod:9` serves other packages in the monorepo, not these two.
- The internal chain is the extraction blocker: `results` needs `metrics.Summarize/Quartiles/BootstrapMedianCI95/Median` (`pkg/results/results.go:166-183` → `pkg/metrics/stats.go:32-73`, `pkg/metrics/metrics.go:34-44`), and `figures` needs `results.Scan` + `results.Group` (`pkg/figures/figures.go:18,81,118`). The module as a whole is `github.com/Realms4239/cgo` at `go 1.25` (`go.mod:1-3`); as standalone these would need their own `go.mod`, and since Go 1.21 the `go` line is a mandatory minimum that toolchains refuse to violate (R3, §"go" directive).

CSV semantics note: `FieldsPerRecord=-1` means "no check is made and records may have a variable number of fields" (R5, `Reader.FieldsPerRecord`), which is exactly what `pkg/results/csvrows.go:26` relies on to keep short historic rows instead of failing the file — a deliberate deviation from the RFC 4180 default that R5 documents as the zero-value behaviour.

### 3.2 Incoming (who needs them)

Grep `results\.|figures\.` across `pkg/`+`cmd/` shows load-bearing use in: `pkg/api/server.go` (`ReadAQM/ColIndex/Field` for delta/replay `pkg/api/server.go:342-349,809-818,1110-1140,1369-1377`; `Scan` for `/api/results`, `/api/integrity`, `/api/report/export` `pkg/api/server.go:770,942,982,1004`; `ProvenanceSHA/Hash8` for compare-cache key `pkg/api/server.go:896`, integrity `pkg/api/server.go:975`, report `pkg/api/server.go:1022,1079`; `Generate` for `/api/figures/regen` `pkg/api/server.go:1084-1090`), `pkg/api/translate.go:12,25,39-44` (`HardwareRecommendation` + `Scan`), `pkg/api/duel.go:58,77-78,122` (`Group` eligibility + `Scan`), `cmd/cgo/server.go:57` (`SeedRuns`), `cmd/cgo/runcli.go:222` (`Scan(dataDir, runID)`), `cmd/cgo/main.go:123` (`figures.Generate("data/runs","data/figures")`). Extraction must keep every one of these call sites compiling, or vendor a compatibility shim (see §7, slices).

### 3.3 Licence

`LICENSE` at repo root is MIT ("Copyright (c) 2026 Meteolink contributors", `LICENSE:3`). A standalone module inherits this cleanly only if the extracted repo keeps the same licence file — required for open-source use, listed as an anti-toy criterion (§8).

## 4. Primary-source requirements

What each primary source **legitimately** requires of a standalone results/figures module — and what it does not require:

### 4.1 RFC 8290 (R1) — FQ-CoDel/CAKE

- RFC 8290 is **Experimental, "not an Internet Standards Track specification"** (R1, Status of This Memo), and it normatively depends on CoDel (RFC 8289) for the AQM half (R1, §2: "we refer the reader to [RFC8289]"). The only MUST in the parameter section is the packet limit: "some packet limit MUST be enforced" (R1, §5.2.3). Default parameter values an extracted module may cite for `fq_codel`: interval 100 ms (R1, §5.2.1), target 5 ms with a low-bandwidth floor tied to MTU transmission time (R1, §5.2.2), packet limit 10240 (R1, §5.2.3), quantum 1514 bytes (R1, §5.2.4), 1024 flows (R1, §5.2.5).
- CAKE has **no RFC**: RFC 8290 cites it only as an informative reference (R1, §10.2 `[CAKE]`) and as "currently being developed as a further refinement" (R1, §7). A module must therefore never cite RFC 8290 as authority for CAKE behaviour — the `cake` branch of `HardwareRecommendation` (`pkg/results/hardware.go:8-10`) and any `cake`-is-better claim stay **unverified — primary source not found** in this pass.
- Critically, **RFC 8290 defines none of this pair's interfaces**: no CSV schema, no `profile|qdisc|cc|direction` cell, no median/IQR/CI95, no `Best`, no SVG, no SHA-256 provenance, no MikroTik translation. Its legitimate use is framing only — sojourn-time/target/interval vocabulary for explaining *why* loaded latency (`small_p95_ms`) is measured (bufferbloat = "dark buffers", R1, Abstract; queue delay vs link busy, R1, §2). Anything stronger is citation theatre.
- The package's actual cell authority is the frozen corpus (`docs/VERIFY.md:24-26`), its counting authority is the G0–G7 gates (`docs/architecture.md:60-74`), its recommendation authority is the lab `Best` computation (`pkg/results/results.go:208-220`) — the MikroTik/RouterOS translation itself is **unverified — primary source not found** and must be labelled as such in the module README until a RouterOS primary is fetched.

### 4.2 Prometheus exposition (R2)

- Since Prometheus 2.0, **all processes exposing metrics must use a text format by default** (R2, intro). The text format is line-oriented, LF-separated, the last line must end with LF, empty lines are ignored (R2, "Prometheus Text Format"). Lines with `#` are comments unless the first token is `HELP`/`TYPE`; only one HELP and one TYPE per metric name; TYPE **must appear before the first sample** of that name; absent TYPE means `untyped` (R2, "Comments, help text, and type information").
- Sample grammar: `metric_name_or_labels value [timestamp]`; `value` is Go `ParseFloat` syntax and **`NaN`, `+Inf`, `-Inf` are valid values** (R2, "The remaining lines describe samples"); timestamp is int64 milliseconds since epoch (R2). All lines for one metric form a single group, HELP/TYPE first, reproducible sorting preferred (R2, "Grouping and sorting"); **each line must have a unique metric-name+labels combination, otherwise ingestion behaviour is undefined** (R2).
- Histogram/summary conventions: `_sum` and `_count` companion samples, `{quantile="y"}` per summary quantile, `{le="y"}` per histogram bucket with a mandatory `{le="+Inf"}` equal to `_count`, buckets/quantiles in increasing order (R2, "Histograms and summaries").
- Consequence for this pair: `pkg/results` emits **JSON `Group` structs** (`pkg/results/results.go:13-47` JSON tags) and `pkg/figures` emits **SVG** (`pkg/figures/figures.go:81-156`) — neither emits any exposition at all. The only builder in-repo is `pkg/api/metrics.go:17-41`, hand-rolled, which already deviates from R2 (no timestamps, no `\`/`"`/LF escaping of label values required by R2 "escaped_string", always emits `{...}` even for label-less samples, no `_sum`/`_count`/quantile conventions, counter/gauge inferred from `_total` suffix). A standalone module that claims Prometheus compatibility must be written against R2 rules above, not by copying `pkg/api/metrics.go` and not by re-labelling JSON/SVG output. OpenMetrics (1.0/2.0 draft, `# EOF` trailer) is the alternative target (R2, "OpenMetrics Text Format") — the module must pick one and declare it.

### 4.3 Go modules / semver (R3, R4)

- A module's **module path is its unique identifier combined with its version, and the prefix of every package path in it** (R3, §"module" directive); v2+ paths must end in `/v2` (R3). Today's `github.com/Realms4239/cgo` (`go.mod:1`) is a monorepo path — a standalone `results`/`figures` needs a new path it controls (R3, "Notes" under `module`), chosen before first publish to avoid a later rename.
- Software using SemVer **MUST declare a public API** in code or documentation, and it SHOULD be precise and comprehensive (R4, spec §1). Versioning rules: MAJOR for incompatible API changes, MINOR for backward-compatible functionality (and deprecations), PATCH for backward-compatible bug fixes (R4, spec §§7–9); **once released, a version's contents MUST NOT be modified** (R4, spec §3); 0.y.z MAY change and is not stable (R4, spec §4) — so a pre-1.0 extraction may break API freely but must say so, and the 1.0.0 release is the moment the §2 table becomes a compatibility promise.
- Go versions are `vX.Y.Z` git tags, **not** SemVer strings (`"v1.2.3" is not a semantic version` — R4, FAQ); the `require` form is `module-path v1.2.3` or a Go-generated pseudo-version (R3, §"require"). Tags must be immutable once pushed, or the checksum database will flag mismatches (R3, §"retract" notes). The embedded `runs-frozen.tar.gz` (`pkg/results/seed.go:15`) is a build artifact, not a versioned module — its contents change without a version bump, which violates the R4 §3 immutability promise if shipped as part of a tagged release without a hash pin (§5).

## 5. Gap analysis (current → releasable)

Ordered by severity. Each gap cites the local evidence; the primary-source side cites R# or is marked unverified.

1. **Dependency chain `metrics → results → figures` blocks one-module extraction.** `pkg/results/results.go:166-183` calls `metrics.Summarize/Quartiles/BootstrapMedianCI95/Median`; `pkg/figures/figures.go:18` calls `results.Scan`. Extracting `results` alone drags in `pkg/metrics` (five files, §2 of `docs/research/2026-11-07-module-metrics.md`); extracting `figures` alone drags in both. Fix: vendor the four needed stats functions into the new module (or depend on the published metrics module once it exists), and make `figures` take `[]Group` instead of a filesystem path (see §6).
2. **No Prometheus exposition in either package.** JSON (`pkg/results/results.go:13-47`) and SVG (`pkg/figures/figures.go:81-156`) are not exposition formats; the only builder (`pkg/api/metrics.go:17-41`) deviates from R2 on timestamps, label-value escaping (`\`, `"`, LF per R2 "escaped_string"), `_sum`/`_count`, and TYPE inference. R2 requires unique metric+labels per line and TYPE-before-first-sample (R2). The module needs a purpose-built `RenderText` function written against R2, with tests covering escaping, `+Inf` bucket, and HELP/TYPE ordering — or must explicitly declare "no Prometheus surface" and stop implying one.
3. **`ProvenanceSHA` is not a corpus provenance.** It hashes only the lexically-last `aqm_eval.csv` (`pkg/figures/figures.go:41-53`), while the frozen corpus is sealed over all 159 manifests (`docs/VERIFY.md:26`, `docs/architecture.md:76-83` describes per-run manifests separately). Changing any of the other 158 files does not move the hash; renaming a run changes which file is "last" without changing content. `docs/architecture.md:80-83` honestly says "sha256 du dernier `aqm_eval.csv`", but the surrounding "triple provenance" framing (`pkg/figures/figures.go:39-40`, `pkg/api/server.go:896,975,1022,1079`) lets readers believe it covers the corpus. Fix: hash a sorted manifest of all files (or reuse the seal), and document that `Hash8` is a label, not a proof.
4. **SVG output is not reproducible and leaks local paths.** `provenanceMeta` embeds `time.Now().UTC()` (`pkg/figures/figures.go:77`) so two runs over identical data produce different bytes, and embeds the local filesystem path as `dc:source` (`pkg/figures/figures.go:76-78`) with hardcoded `dc:creator Meteolink` (`pkg/figures/figures.go:78`). A standalone tool must accept a fixed timestamp/source or omit them; today `TestFiguresRDF` (`pkg/figures/figures_test.go:41-50`) only checks substrings, so byte-instability is invisible to tests.
5. **`SeedRuns` ships an unversioned, unhashed corpus snapshot.** `//go:embed runs-frozen.tar.gz` (`pkg/results/seed.go:15-16`) bakes a tarball into the binary with no hash check on extract, no version, and no relationship to the seal `docs/frozen-manifests.sha256` (`docs/VERIFY.md:26`). Per R6 the embed is legal (single-pattern `[]byte`, package-scope), but per R4 §3 a released version's contents MUST NOT be modified — yet the tarball can be silently replaced between tags. It also couples the library to one operator's historic runs, which is app bootstrap logic, not aggregation. Fix: move to an app-level `cmd` or a separate `seed` subpackage with an explicit corpus version + SHA-256 pin, and delete from the core module.
6. **`HardwareRecommendation` couples aggregation to operator advice — and its authority is unverified.** It is called inside `Scan` for every group (`pkg/results/results.go:221-234`), returns a hardcoded French sentence (`pkg/results/hardware.go:7-12`), duplicates `pkg/api/translate.go:12,39` and the frontend mirror `web/frontend/src/lib/hardware.ts` (comment `pkg/results/hardware.go:6`), and asserts "prouvé en lab" / RouterOS v7+ translation with **unverified — primary source not found** provenance (CAKE: R1 §7/§10.2 informative only; MikroTik: not fetched). A standalone aggregation library must not ship a vendor recommendation as a field of every result row. Fix: delete the field from `Group`; expose an optional, clearly-labelled advisory function for the CGO app only.
7. **Silent skipping makes "empty" indistinguishable from "corrupt".** `Scan` returns `nil,nil` on no files (`pkg/results/results.go:71-73`), `continue`s on unreadable files (`pkg/results/results.go:93-96`) and on missing `profile/qdisc/cc` headers (`pkg/results/results.go:97-100`), and `ReadAQM` keeps short rows with missing fields as `""` (`pkg/results/csvrows.go:16-18`, `pkg/results/csvrows.go:59-65`). A directory with one corrupt file and 158 good ones produces the same shape as a directory with 158 good ones plus silence. Callers in `pkg/api/server.go:770,982` ignore the error entirely (`groups, _ :=`). Fix: return a structured diagnostic (files seen / files skipped with reason) or at minimum count skipped files in the result.
8. **`Best` is computed on all-rows medians, but decisions should be valid-only.** `Best` picks min `Smallp95Median` (`pkg/results/results.go:208-220`), which includes `degraded` rows (`pkg/results/results.go:156-161` adds degraded to `smalls`), while the valid-only fields (`pkg/results/results.go:34-40`) exist precisely to exclude degraded. A cell whose median is low only because of degraded rows can win `Best` and drive `HardwareRecommendation`, the duel, and the exported report. Document the choice or compute `Best` on `Smallp95ValidMedian` with an explicit `n≥` guard.

9. **Chart math has no guards and no scales.** `writeBar` computes `barW = usableW/(len(groups)*2+1)` with integer division (`pkg/figures/figures.go:94`) — zero-width bars when groups are numerous, with no guard; both charts draw axis lines but no tick labels or scale (only max-normalised bar/dot heights, `pkg/figures/figures.go:100,141-142`), fixed 800×400 canvas, hardcoded colours/fonts (`#070707`, Cormorant Garamond, JetBrains Mono, `pkg/figures/figures.go:97,108`). The scatter labels (`qdisc/cc` per point, `pkg/figures/figures.go:148`) will overlap with many groups. This is a demo renderer, not a releasable chart library — it needs data-driven sizing, tick axes, and tests pinning pixel values, or must be declared an example renderer.
10. **Recomputation cost is paid on every call with no cache.** `Scan` runs `BootstrapMedianCI95(...,10000,42)` twice per group (`pkg/results/results.go:173-174`) on every invocation, and HTTP handlers call `Scan` repeatedly per request (`pkg/api/server.go:770,942,982,1004`, `pkg/api/duel.go:77`). Bootstrap statistical validity is **unverified — primary source not found** beyond the local Park-Miller LCG (`pkg/metrics/stats.go:46-73`). A standalone module must either cache by corpus hash or move CI computation out of the hot path / make resamples configurable.
11. **No module identity, no README, no schema declaration.** Today the pair is `github.com/Realms4239/cgo/pkg/results` + `/pkg/figures` on `go 1.25` (`go.mod:1-3`) with no package doc beyond inline comments. The 17/18/19-column tolerance (`pkg/results/csvrows.go:9-10`) is not a declared schema — no column list, no types, no `gate_status` enum (`valid/degraded/invalid` only in prose, `docs/architecture.md:62`), no direction enum (`up` default, `down` explicit, `pkg/results/results.go:125-131`). A standalone release needs its own `go.mod` (R3), a README declaring the public API (R4 §1), the accepted CSV schema, and the exposition format choice (R2).

## 6. Target module API

Proposed contract for a standalone module (illustrative names; the shape is the deliverable, not the spelling). Per R4, this declaration **is** the public API once published, so every item below is a 1.0.0 compatibility commitment (R4, spec §§5–8). Per R3, it would live behind a new module path with its own `go.mod` and `go` directive; pre-1.0 (`0.y.z`) it MAY change (R4, spec §4).

```go
// Core — pure aggregation over an explicit file list (kills gaps §5.1/§5.7).
type Row struct { Profile, Qdisc, CC, Direction string; Gate string; Values map[string]float64 }
type Source interface { Files() ([]File, error) } // or simply: func Scan(files []File) ([]Group, error)
type File struct { Name string; Header []string; Rows [][]string } // ReadAQM stays, but callers pass bytes, not paths
func Scan(files []File) (out []Group, skipped []Skipped, err error) // skipped names the corrupt files (gap §5.7)

type Group struct {
    Profile, Qdisc, CC, Direction string
    Count, Quarantined, Degraded int
    SmallP95, RTTP95, Goodput Summary // Summary{N, Median, Q1, Q3} — degraded-included, labelled as such
    SmallP95Valid, DeadlineValid Summary // valid-only, N + median + CI95; CI resamples configurable, default documented
    Runs []string // sorted provenance (keep pkg/results/results.go:46)
    // DELETE: Best (or compute on Valid with n>= guard, gap §5.8), HardwareRecommendation (gap §5.6)
}

// Provenance — corpus-wide, reproducible (kills gaps §5.3/§5.4).
type Provenance struct { SHA256 string; Hash8 string; Files int; At time.Time } // At is caller-supplied, never time.Now() inside
func ProvenanceOf(files []File) Provenance // hash sorted "name + sha256(bytes)" manifest, not last-file-only

// Figures — renderer over values, not filesystem (kills gaps §5.4/§5.9).
func RenderBar(groups []Group, opts Opts) ([]byte, error)   // data-driven size, tick axes, no hardcoded fonts as API
func RenderScatter(groups []Group, opts Opts) ([]byte, error)
func Generate(dataDir, outDir string) error // DELETE from core; keep as cmd/app adapter only

// Exposition — NEW, written against R2 (gap §5.2).
type Sample struct { Metric string; Labels map[string]string; Value float64; TimestampMs int64 }
func RenderText(groups []Group, prov Provenance) (string, error) // HELP/TYPE first, unique metric+labels enforced, escaping per R2, LF-terminated; or declare OpenMetrics + "# EOF"

// Seed + hardware — DELETE from core (gaps §5.5/§5.6).
// SeedRuns -> app-level cmd with corpus version + SHA-256 pin. HardwareRecommendation -> CGO-app advisory, labelled unverified.
```

Design notes: (a) filesystem and clock are inputs, never ambient (`time.Now`, `filepath.Glob` inside core are the reproducibility bugs); (b) every skipped file is named, never silently dropped; (c) recommendation and seeding are adapters, not fields; (d) one summary type, one valid-only path, CI resamples explicit.

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check so a verifier can close it.

1. **Slice 1 — Freeze + declare.** Snapshot the §2 exported API into the module README as the `0.1.0` declared public API (R4 spec §1); create the new `go.mod` with its own module path (R3); keep `LICENSE` (MIT, `LICENSE:3`); port all existing tests unchanged (they pin behaviour today, §2.6). Acceptance: `go test ./...` green on the extracted tree; README lists every §2 identifier with its `file:line`.
2. **Slice 2 — Break the `metrics` chain.** Vendor `Summarize/Quartiles/Median/BootstrapMedianCI95` (or depend on the published metrics module) into the new module; change `Scan` to accept an explicit file list / `io.Reader` set instead of `dataDir` globbing; return skipped-file diagnostics. Keep in-repo callers compiling via a thin adapter that globs `data/runs`. Acceptance: no import of `github.com/Realms4239/cgo/pkg/metrics` remains in the extracted tree; a test with one corrupt + one good file asserts the skipped name is reported.
3. **Slice 3 — Fix provenance + reproducibility.** Replace last-file `ProvenanceSHA` with a sorted corpus manifest hash; make `provenanceMeta` take caller-supplied timestamp/source; assert byte-identical SVG output for identical input. Acceptance: mutating any non-last file changes the hash; two `Generate` runs over unchanged data are byte-identical.
4. **Slice 4 — Decoupled adapters.** Move `SeedRuns` to an app `cmd` with a corpus version + SHA-256 pin checked on extract; move `HardwareRecommendation` out of `Group` into a CGO-app advisory labelled **unverified — primary source not found** for CAKE/MikroTik until primaries are fetched; decide and document `Best` on valid-only vs all-rows. Acceptance: no `embed`, no French vendor string, and no `Best` field in the core `Group`; in-repo `pkg/api`/`cmd/cgo` still compile via adapters.
5. **Slice 5 — Renderer + exposition.** Rewrite bar/scatter as value-in renderers with tick axes and size guards; add `RenderText` (or OpenMetrics) written against R2. Acceptance: exposition golden tests pass against an R2-conformant parser (escaping, unique metric+labels, HELP/TYPE-before-sample, `+Inf` bucket if histograms are exposed, LF final line); chart tests pin pixel values for a 2-group fixture.
6. **Slice 6 — First tag.** Resolve module path (R3), tag `v0.1.0`, publish; migrate `pkg/api`, `pkg/figures` callers, and `cmd/cgo` to the new module version in-repo. Acceptance: `go list -m all` shows the new module; in-repo `go test ./...` + `python tools/verify.py` green; no `replace` directives left in `go.mod`; frozen-corpus counts still reproduce `159/528/208/74/246` (`docs/VERIFY.md:24`).

Each slice is independently reviewable; Slices 2–5 are the only ones that change behaviour-visible semantics and each should be verifier-re-run before its `VERIFIED` row is claimed (per the repo's builder/verifier split, `AGENTS.md` Roles).

## 8. Anti-toy criteria (what would make this a toy — and the bar for "not a toy")

A results/figures module is a toy if it re-describes its own outputs instead of aggregating against an external check. This extraction is **not** a toy only if all of the following hold:

1. **Independent reproduction of the frozen corpus.** `Scan` over the sealed corpus must reproduce `159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells` (`docs/VERIFY.md:24`) and the live `/api/integrity` shape `{runs:159, manifests:159, valid:208, degraded:74, quarantined:246}` (`docs/VERIFY.md:51`) — not just pass its own fixture tests. If a counting-rule change moves any of those numbers without a named `VERIFIED` row, it is a toy that rewrites history.
2. **Hand-vector correctness.** Medians/IQRs/CI95 against hand-computed vectors on the 17/18/19-column fixtures already in `pkg/results/results_test.go:70-160` (e.g. `Smallp95Median == 328.2` not QDI `57.4`, `GoodputMedian == 18.2` not deadline `60.2`), plus the direction-split (`pkg/results/direction_test.go:30-59`) and 30-iteration order-stability (`pkg/results/direction_order_test.go:13-33`) tests — ported unchanged, not rewritten to match new output.
3. **Provenance honesty.** Every hash claim names what it covers: corpus manifest hash (new) vs last-file hash (old), with `Hash8` labelled a short label, not a proof. The MikroTik/CAKE translation, CAKE behaviour, and bootstrap validity stay labelled **unverified — primary source not found** until primaries are fetched (R1 §7/§10.2 for CAKE; RouterOS and statistics primaries not fetched in this pass). A module that presents a French vendor sentence or a last-file hash as corpus proof is a toy.
4. **No silent skip, no silent zero.** A corrupt or short-row file is named in diagnostics, never dropped; absent data is an explicit empty/skipped result, never a `0` median that a dashboard or Prometheus scrape (R2 valid-`NaN` noted as an option the API must choose) will read as a measurement. If two importers can feed different file sets and get the same `Group` shape with no way to tell which files were skipped, it is an app helper, not a module.
5. **Byte-stable figures + a real scrape.** Two renders of identical input are byte-identical (caller-supplied timestamp, no local path leak, §5.4), at least one out-of-tree importer exists (not only `pkg/api`/`cmd/cgo` in this repo), and one R2-conformant scrape succeeds against the module's own exposition (HELP/TYPE, escaping, unique metric+labels, LF termination) — otherwise the Prometheus and figure claims are decoration.
6. **Version discipline.** A declared API (R4 spec §1), tags `vX.Y.Z` treated as immutable (R3 §"retract"), and a changelog phrased in MAJOR/MINOR/PATCH terms (R4 specs §§7–9). A repo that silently swaps the embedded `runs-frozen.tar.gz` between tags, or ships no declared API, is a toy regardless of code quality.

---
*Teams note: slices 1–6 are proposals in this research file only. No `pkg/results` or `pkg/figures` source was modified in this pass; implementation waits for human approval per `AGENTS.md` ("Stop and ask" — new dependency, scope growth).*

*Skipped / not checked: live bench re-measurement (no Linux 6.8 host in this pass — `docs/VERIFY.md:28` BLOCKED), CAKE/MikroTik/statistics primaries (marked unverified above), and the `book/` frozen corpus itself (never touched). Risk the user must know: `ProvenanceSHA` today covers only the last file, so any figure/report/integrity claim citing it as corpus proof is weaker than it reads (`pkg/figures/figures.go:41-53` vs `docs/VERIFY.md:26`).*
