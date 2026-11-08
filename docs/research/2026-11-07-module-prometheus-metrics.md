# pkg/api/metrics.go + deploy/prometheus/rules.yml as standalone Prometheus tool — primary-source research

*Date: 2026-11-07 · Scope: `pkg/api/metrics.go`, `pkg/api/metrics_test.go`, `deploy/prometheus/rules.yml`, `docs/api.md`, `pkg/api/server.go` (wiring line only) — read as if extracted tomorrow into `cgo-exporter`. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL | Status | Used for |
|---|---|---|---|---|
| R1 | Prometheus — Exposition formats | https://prometheus.io/docs/instrumenting/exposition_formats/ | fetched | Text-format contract: Content-Type, HELP/TYPE ordering, escaping, grouping, histogram/summary conventions |
| R2 | Prometheus — Recording rules (+ `promtool check rules`) | https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/ | fetched | Record syntax, record-vs-alert names, `promtool check rules` exit contract |
| R3 | Prometheus — Alerting rules | https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/ | fetched | `alert`/`expr`/`for`/`keep_firing_for` semantics, pending→firing |
| R4 | Prometheus — Data model | https://prometheus.io/docs/concepts/data_model/ | fetched | Metric/label name regex, colon reserved for recording rules, empty label == absent label |
| R5 | Prometheus — Metric types | https://prometheus.io/docs/concepts/metric_types/ | fetched | Counter=monotonic, gauge=goes up/down, histogram/summary `_bucket`/`_count`/`_sum` shapes |
| R6 | Prometheus — Metric and label naming (best practices) | https://prometheus.io/docs/practices/naming/ | fetched | App prefix, single unit, base units, `_total`/`_ratio`/unit suffixes, label cardinality caution |
| R7 | Prometheus — Recording rules naming/aggregation | https://prometheus.io/docs/practices/rules/ | fetched | `level:metric:operations` shape, `without` clause, never average a ratio |
| R8 | Prometheus — Alerting (best practices) | https://prometheus.io/docs/practices/alerting/ | fetched | CamelCase alert names, alert on symptoms, metamonitoring |
| R9 | Prometheus — Unit testing for rules (`promtool test rules`) | https://prometheus.io/docs/prometheus/latest/configuration/unit_testing_rules/ | fetched | Input-series / `alert_rule_test` / `promql_expr_test` harness |
| R10 | Grafana — Grafana Alerting (overview) | https://grafana.com/docs/grafana/latest/alerting/ | fetched (overview only) | Queries/expressions from multiple data sources; consolidated alert view |
| R11 | Prometheus — Query functions (`rate`, `absent`, `absent_over_time`, `increase`, `avg_over_time`, aggregations) | https://prometheus.io/docs/prometheus/latest/querying/functions/ | fetched | Which functions apply to which metric type; `rate` counters-only |

*Method: read `pkg/api/metrics.go:1-64`, `pkg/api/metrics_test.go:1-43`, `deploy/prometheus/rules.yml:1-26`, `docs/api.md:1-116`, `pkg/api/server.go:397-399` (wiring), `deploy/docker-compose.yml:1-13`. Every technical claim below cites `file:line` or `(R#)`; anything not traceable to a fetched primary is marked **unverified — primary source not found**.*

## 1. Responsibility

Today `pkg/api/metrics.go` is **not** a real exporter — it is a hand-rolled exposition *builder* plus a hard-coded demo handler:

- **Pure builder** `renderMetrics(samples map[string][]metricSample, help map[string]string)` (`pkg/api/metrics.go:17-41`) — the only Prometheus text emitter in the product. Doc comment states the contract: "Absent is never emitted; the caller only pushes measured values" (`pkg/api/metrics.go:15-16`). It sorts metric names (`pkg/api/metrics.go:19-23`) and label parts (`pkg/api/metrics.go:36`), emits `# HELP` (`pkg/api/metrics.go:25`) and `# TYPE` (`pkg/api/metrics.go:30`), then one sample per line (`pkg/api/metrics.go:37`).
- **HTTP handler** `handleMetrics(getBuildInfo func() map[string]string)` (`pkg/api/metrics.go:43-63`) — serves `GET /metrics` (wired at `pkg/api/server.go:397-399`). It sets `text/plain; version=0.0.4` (`pkg/api/metrics.go:45`), then writes **six static samples, all `"0"`, all with `site="unknown"` or `status="completed"`** (`pkg/api/metrics.go:46-53`, help text at `pkg/api/metrics.go:54-61`). It never reads the audit/iface/campaign state, never increments anything, and **ignores `getBuildInfo` entirely** — the parameter is dead code (`pkg/api/metrics.go:43` accepted, never called in `pkg/api/metrics.go:44-63`).

What it deliberately does **not** do: pull from `pkg/probe`/`pkg/campagne`/`pkg/audit`, read `data/link_audit.csv` or `data/runs`, honour `--mode`/token checks, or expose the frozen-corpus quantities (`cgo_small_p95_ms`, `cgo_rtt_p95_ms`, `cgo_bloat_delta_ms`, `cgo_qdisc_drops_total`, `cgo_run_status`, `cgo_build_info`) that the hardening spec names as the contract (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:61-70`). And it is **undocumented**: `docs/api.md:3` declares itself verified against `server.go`/`sse.go`/`translate.go` only, and has no `/metrics` row at all (`docs/api.md:15-116`).

`deploy/prometheus/rules.yml` is the consumer side: one recording rule plus three alerts (`deploy/prometheus/rules.yml:4-26`) — a small SLO/aliveness pack for CAP SLA, dead probes, and invalid interface sampling. Nothing loads it: `deploy/docker-compose.yml:1-13` defines only the `cgo` service, with no Prometheus service, no `rule_files` mount, and no `promtool` step — the file is inert in the current deploy.

Read as a standalone tool, the pair's job would be: **emit a Prometheus-scrapable text exposition of measured link/campaign facts, and evaluate SLO/probe/iface alert expressions against that exposition.** The extraction question is whether the current builder + rules can satisfy that job against the primary sources. §5 shows they cannot as-is.

## 2. Exported metric surface (file:line)

### 2.1 `pkg/api/metrics.go` — emitted series

| Metric | TYPE emitted | Labels emitted | Location | Notes |
|---|---|---|---|---|
| `cgo_audit_runs_total` | `counter` (via `_total` suffix) | `status="completed"` only | `pkg/api/metrics.go:47`, help `pkg/api/metrics.go:55` | Spec contract is `{site,link_type,status}` (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:62`) — `site`/`link_type` absent here. |
| `cgo_audit_load_attempted_total` | `counter` | `site="unknown"` | `pkg/api/metrics.go:48`, help `pkg/api/metrics.go:56` | — |
| `cgo_audit_load_measured_total` | `counter` | `site="unknown"` | `pkg/api/metrics.go:49`, help `pkg/api/metrics.go:57` | Only rule-referenced metric that **is** emitted (`deploy/prometheus/rules.yml:14`). |
| `cgo_cap_sla_measured` | `gauge` (default branch) | `site="unknown"` | `pkg/api/metrics.go:50`, help `pkg/api/metrics.go:58` | 0/1 flag; no companion value metric emitted. |
| `cgo_iface_valid` | `gauge` | `site="unknown"` | `pkg/api/metrics.go:51`, help `pkg/api/metrics.go:59` | Lived metric referenced by `deploy/prometheus/rules.yml:21`. |
| `cgo_jfi_measured` | `gauge` | `site="unknown"` | `pkg/api/metrics.go:52`, help `pkg/api/metrics.go:60` | Help text says "0 means absent, never a zero value" (`pkg/api/metrics.go:60`) — endpoint emits `0` anyway. |

Unexported pieces (not API, but part of the surface): `metricSample` struct (`pkg/api/metrics.go:10-13`), `renderMetrics` (`pkg/api/metrics.go:17`), `handleMetrics` (`pkg/api/metrics.go:43`, wired at `pkg/api/server.go:397`). `renderMetrics` is test-covered through `pkg/api/metrics_test.go:10-43` (HELP/TYPE/Content-Type presence; absent≠zero at builder level).

**Not emitted at all** (but named by spec/rules): `cgo_cap_deadline_ok_pct`, `cgo_cap_sla_breach`, `cgo_small_p95_ms`/`cgo_rtt_p95_ms`, `cgo_bloat_delta_ms`, `cgo_jfi_pct`, `cgo_qdisc_drops_total`, `cgo_iface_rx_mbps`/`cgo_iface_tx_mbps`/`cgo_iface_drops_total`, `cgo_run_status`, `cgo_build_info{version}` (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:61-70`; `deploy/prometheus/rules.yml:5,7`). The `getBuildInfo` closure that could have produced `cgo_build_info` is never called (`pkg/api/server.go:397-399` passes `version`, `pkg/api/metrics.go:43` drops it).

### 2.2 `deploy/prometheus/rules.yml` — evaluation surface

| Rule | Kind | Expr | Location |
|---|---|---|---|
| `cgo:slo:deadline_ok_pct_5m` | record | `avg(rate(cgo_cap_deadline_ok_pct[5m]))` | `deploy/prometheus/rules.yml:4-5` |
| `CGO_deadline_breach_sustained` | alert (`for: 5m`, `severity: critical`) | `cgo_cap_sla_breach == 1` | `deploy/prometheus/rules.yml:6-12` |
| `CGO_probe_dead` | alert (`for: 10m`, `severity: warning`) | `rate(cgo_audit_load_measured_total[5m]) == 0` | `deploy/prometheus/rules.yml:13-19` |
| `CGO_iface_invalid` | alert (`for: 5m`, `severity: warning`) | `cgo_iface_valid == 0` | `deploy/prometheus/rules.yml:20-26` |

## 3. Dependencies

### 3.1 Outgoing (what the tool needs)

- `pkg/api/metrics.go:3-8` imports stdlib only: `fmt`, `net/http`, `sort`, `strings`. No `github.com/prometheus/client_golang` import, no `require` line consumed (`go.mod:5-10` carries `fyne`/`bubbletea`/`lipgloss`/`golang.org/x/sys` — none is touched by this file). For a standalone repo this is a **zero-dependency** start: the builder compiles with `net/http` alone. The cost is that exposition correctness is hand-rolled instead of delegated to a maintained client library (see §5).
- One internal seam: `pkg/api/server.go:397-399` builds the handler with `func() map[string]string { return {"version": d.Version} }`. That closure is the whole coupling — an extracted exporter needs only a version string, not the `Deps` struct.

### 3.2 Incoming (who needs it)

- `pkg/api/server.go:397` `GET /metrics` route. Test `pkg/api/metrics_test.go:10-28` exercises it through `New(Deps{Mode:"observe"})`.
- `deploy/prometheus/rules.yml:5,7,14,21` references four metric names; only two of those four are emitted (§2.1). No compose service, Grafana dashboard, or CI step consumes `rules.yml` (`deploy/docker-compose.yml:1-13` has no Prometheus/Grafana service).
- `docs/api.md` does **not** document the endpoint (§1) — documentation consumers cannot discover its contract.

### 3.3 Licence

Repo root `LICENSE` is MIT ("Copyright (c) 2026 Meteolink contributors", `LICENSE:3`). An independent exporter repo inherits it cleanly if it ships the same file — **a claim about this repo's licence, not about any other licence's terms; the licence text itself was not re-read in this pass, so portability beyond this repo is `unverified — primary source not found`.**

## 4. Primary-source requirements

### 4.1 HELP / TYPE (R1)

R1 requires: lines starting with `#` are comments unless the first token is `HELP` or `TYPE` (R1, "Comments, help text, and type information"). Only one `HELP` line may exist per metric name; only one `TYPE` line per metric name; the `TYPE` line **must appear before the first sample** of that name; absent `TYPE` means `untyped` (R1, same section). The `TYPE` token must be one of `counter`, `gauge`, `histogram`, `summary`, `untyped` (R1). `HELP` docstrings may contain any UTF-8 except that backslash and line-feed must be escaped as `\\` and `\n` (R1).

Current builder satisfies HELP/TYPE *placement*: it writes `# HELP` then `# TYPE` then samples, per metric (`pkg/api/metrics.go:25,30,37`), and metric names are sorted so the group is stable (`pkg/api/metrics.go:19-23`). What it violates is *content*: it infers type from the name suffix — `_total` → `counter`, otherwise `gauge` (`pkg/api/metrics.go:26-29`) — which R1/R5 do not sanction: type is declared metadata, not a name-derived default, and R5 defines counter as "a cumulative metric that represents a single monotonically increasing counter" and gauge as "a single numerical value that can arbitrarily go up and down" (R5, "Counter"/"Gauge"). A future `_total` that can decrease would be mis-declared.

### 4.2 Content-Type (R1)

R1's table fixes the text exposition's `Content-Type` as `text/plain` with parameters `version=0.0.4` (missing `version` falls back to latest) (R1, "Basic info"). R1 also requires scrape targets to return a valid `Content-Type` from Prometheus 3.0 onward: "If the Content-Type is missing, unparsable, or not a supported media type, the scrape will fail" (R1, "HTTP Content-Type requirements").

`pkg/api/metrics.go:45` sets exactly `text/plain; version=0.0.4` — conformant. Test pins the prefix (`pkg/api/metrics_test.go:18-20`). No `charset`/encoding parameter is added; R1 declares optional `Content-Encoding: gzip` only, so the bare header is the safe shape. What R1 does not define: any requirement about TLS/auth for `/metrics` — **authentication requirements for a metrics endpoint are `unverified — primary source not found`** in the exposition spec; this repo's own edge ACL (`deploy/nginx/cgo.conf:17-19`, loopback-only) is the local decision, not a Prometheus requirement.

### 4.3 Naming (R4, R6)

R4 requires/recommends: metric names SHOULD match `[a-zA-Z_:][a-zA-Z0-9_:]*` with any UTF-8 allowed but quoting needed outside it; label names SHOULD match `[a-zA-Z_][a-zA-Z0-9_]*`; names beginning with `__` are reserved; colons are "reserved for user-defined recording rules" and "SHOULD NOT be used by exporters" (R4, "Metric names and labels"). Labels with an empty value are equivalent to absent labels (R4).

R6 (best practice, not a parser requirement) adds: SHOULD have a single-word application prefix; MUST refer to a single unit and a single quantity; SHOULD use base units (seconds, not milliseconds; bytes, not megabytes), SHOULD carry a plural unit suffix, and "an accumulating count has `total` as a suffix, in addition to the unit if applicable" (R6, "Metric names"). Base units table: Time→seconds, Bits→bytes, Percent→ratio with values 0–1, suffix `ratio` (R6, "Base Units"). Labels: do not encode the label name in the metric name; every unique label-value combination is a new time series, so "do not use labels to store dimensions with high cardinality (many different label values), such as user IDs, email addresses, or other unbounded sets" (R6, "Labels").

Mapping: `cgo_audit_runs_total` / `cgo_audit_load_*_total` satisfy prefix + `_total` (R6). `cgo_cap_sla_measured`, `cgo_iface_valid`, `cgo_jfi_measured` carry no unit suffix — not a violation (R6 is SHOULD-level), but they are boolean flags, and R6's ratio convention would name a percentage `*_ratio` in 0–1 rather than `*_pct` in 0–100 (R6, "Base Units"). The spec names `cgo_cap_deadline_ok_pct`, `cgo_jfi_pct`, `cgo_small_p95_ms` (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:64-66`) — all three use the non-base `pct`/`ms` forms R6 advises against (SHOULD, not MUST — renaming is a consistency decision, noted in §5, not a conformance break). All emitted names match the R4 regex, none uses `:`, the `cgo` prefix is the application prefix R6 asks for (R6), and the emitted label keys (`status`, `site`) match the R4 label regex (R4).

### 4.4 Histogram / summary (R1, R5)

R1 fixes the only legal text representation of the complex types (R1, "Histograms and summaries"): a summary/histogram named `x` exposes `_sum` as `<x>_sum`, `_count` as `<x>_count`, each summary quantile as `x{quantile="y"}`, each histogram bucket as `x_bucket{le="y"}`; a histogram **must** have `{le="+Inf"}` whose value equals `x_count`; buckets/quantiles must appear in increasing numeric order. R5 gives the matching type definitions: histogram counts observations in configurable buckets (`_bucket{le}`, `_sum`, `_count`), and a summary additionally yields streaming φ-quantiles, `_sum`, `_count` (R5, "Histogram"/"Summary").

Current builder emits **no histogram or summary at all** — every series is counter or gauge. That is legal today (R1 permits counter/gauge-only exposition), but it means a latency percentile (`p95`) cannot be expressed as a histogram/summary without rewriting the builder: the builder has no `_bucket`/`_sum`/`_count`/`{quantile}`/`{le}` code path, no `+Inf` bucket obligation handling, and no ordering of buckets/quantiles. The spec's `cgo_small_p95_ms{site,profile}` gauge idea (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:64`) is a *precomputed gauge*, not a histogram — the exporter must decide (document, R4-spec §1-style API declaration) whether p95 is a sampled gauge or a real histogram exposed per R1, because the two query differently (`histogram_quantile` vs plain gauge).

### 4.5 Escaping, grouping, sample grammar (R1)

R1 sample grammar: `metric_name_or_labels value [timestamp]`; `label_value` is UTF-8 where `\`, `"`, and line-feed must be escaped as `\\`, `\"`, `\n` (R1, "Line format"/`escaped_string`). `value` follows Go `ParseFloat`, with `NaN`, `+Inf`, `-Inf` explicitly valid (R1). Optional timestamp is int64 milliseconds since epoch (R1). All lines for a metric must be one group with HELP/TYPE first; reproducible sorting is preferred but not required; "Each line must have a unique combination of a metric name and labels. Otherwise, the ingestion behavior is undefined" (R1, "Grouping and sorting").

Builder gaps against R1: label values are interpolated raw via `fmt.Sprintf(`%s="%s"`, k, v)` (`pkg/api/metrics.go:34`) with **no escaping** — any future label containing `"`, `\`, or newline breaks the exposition. Today's labels are `site=unknown`/`status=completed` so nothing breaks yet. It always emits `{...}` even for label-less series (`pkg/api/metrics.go:37`) — R1 permits `metric_name[ "{" labels "}" ]` with braces omitted when label-less (R1 grammar), so always-braced is legal, though noisy. It never emits a timestamp (R1 makes it optional, so legal) and never emits `NaN`/`+Inf`/`-Inf` (R1-valid values; current values are all the string `"0"` — R1-valid, but semantically fabricated, §5). Uniqueness within a metric is caller-enforced by map key; the builder itself does not dedupe or sort samples inside a metric — R1 prefers reproducible order (R1, "Grouping and sorting") but does not require it, so intra-metric order is **unverified** as a conformance risk beyond scrape-noise (primary source silent).

### 4.6 Alerting, recording, and `promtool` (R2, R3, R7, R8, R9, R11)

R2: recording and alerting rules live in YAML rule files with `groups: [{name: <string>, rules: [...]}]`; record names must be valid metric names, alert names must be valid label values (R2, "Recording rules"). `promtool check rules <file>` checks syntax without starting Prometheus: valid → prints parsed rules, exit 0; syntax error → error to stderr, exit 1 (R2, "Syntax-checking rules"). R9 adds `promtool test rules <test.yml>` with `input_series`, `alert_rule_test`, `promql_expr_test` (R9, "Test file format") — the mechanism for asserting that an alert fires/doesn't fire on synthetic series.

R3: an alerting rule is `alert: <label-value>`, `expr: <PromQL>`, optional `for: <duration>` (pending until the expr stays true for that duration), optional `keep_firing_for`, plus `labels:`/`annotations:` (templated) (R3, "Defining alerting rules"). No `for` → fires on first evaluation (R3).

R7: recording-rule names should be `level:metric:operations` (newest operation first), metric name unchanged except stripping `_total` after `rate`/`irate`; with aggregation use `without (…)` preserving other labels; "Do not take the average of a ratio or average of an average" — aggregate numerator and denominator separately then divide (R7, "Naming"/"Aggregation").

R11: `rate(v[range])` "should only be used with counters (for both floats and histograms)" and computes per-second average increase with counter-reset handling (R11, "`rate()`"). `absent(v)` "returns a 1-element vector with the value 1 if the vector passed to it has no elements … useful for alerting on when no time series exist" (R11, "`absent()`"); `absent_over_time` does the same over a range (R11, "`absent_over_time()`").

R8: alert names have no strict restriction but "the community has rallied around using CamelCase" (R8, "Naming"); alert on symptoms, not causes; and "It is important to have confidence that monitoring is working" via metamonitoring alerts (R8, "What to alert on", "Metamonitoring").

`rules.yml` check-in: `record: cgo:slo:deadline_ok_pct_5m` uses colons — **correct** for recording rules (R4 reserves `:` for recording rules) and matches `level:metric:operations` (R7); group name `cgo-slo` is unique within file (R2). Three alerts are well-formed `alert`/`expr`/`for`/`labels`/`annotations` blocks (R3). Deviations: (a) `avg(rate(..._pct[5m]))` takes `rate` of a percentage gauge — violates R11's counters-only rule and R7's never-average-a-ratio rule; (b) `rate(cgo_audit_load_measured_total[5m]) == 0` (`deploy/prometheus/rules.yml:14`) is a category أيضاًне correct: a missing series yields *no elements*, so `== 0` matches nothing — the R11-prescribed shape for "no series" is `absent()` (R11); (c) no `keep_firing_for` is set — R3 defaults it to 0s, so alerts deactivate on first non-matching evaluation (R3) — flap-prone but legal; (d) SCREAMING_SNAKE alert names deviate from R8's CamelCase convention (SHOULD-level). Whether `deploy/prometheus/rules.yml` parses at all is **unverified in this pass** — `promtool check rules` was not executed (no code executed, per brief) and the file's YAML/PromQL well-formedness is therefore `unverified — primary source not found` *for this file* until run.

### 4.7 Grafana alerting (R10)

R10 frames Grafana Alerting as: monitor incoming metrics/logs, "create queries and expressions from multiple data sources", and "create, manage, and take action on your alerts from a single, consolidated view" (R10, Overview). That is the only claim R10 supports in this pass. Anything about Grafana's own alert-rule evaluation, `for`/`keep_firing_for` equivalence, or notification routing is **unverified — primary source not found** (the Grafana alert-rule sub-pages were not fetched). Practically, a `cgo-exporter` that emits Prometheus text format (R1) is queryable from Grafana through its Prometheus data source on the same contract (R10) — that composition is the only Grafana-side requirement the overview actually states; provisioning details are **unverified — primary source not found**.

## 5. Gap analysis (current → releasable)

Ordered by severity. Each gap cites local `file:line` and the R# it breaks.

1. **Four of six spec metrics are static zeroes, not measurements.** `handleMetrics` returns fixed `"0"` for all six names (`pkg/api/metrics.go:46-53`) and reads no state. R1's "Types and docstrings not integral part of the syntax, meaning little-to-nonexistent metric contract validation" (R1, "Basic info") means a scraper accepts these zeros — so they read as measured zeros in every dashboard. The builder's own invariant ("Absent is never emitted", `pkg/api/metrics.go:15-16`; pinned at `pkg/api/metrics_test.go:37-38`) is violated one layer up: *absent* measurements are emitted as `0`. R5 defines gauge as values that go up/down (R5) — a counter that never moves is not a counter (R5), and an absent-iface `cgo_iface_valid=0` (R1-valid value) is semantically "invalid" when the truth is "not sampled". This is the largest gap: the endpoint prints telemetry-shaped lies.
2. **Rules reference metrics that are never emitted.** `cgo_cap_deadline_ok_pct` and `cgo_cap_sla_breach` appear in `deploy/prometheus/rules.yml:5,7` but nowhere in `pkg/api/metrics.go:46-53`. The recording rule and `CGO_deadline_breach_sustained` can never fire from the real endpoint — a violation of R8 metamonitoring ("confidence that monitoring is working", R8), producing silent-green SLO dashboards. Either emit both measured or delete `deploy/prometheus/rules.yml:4-12` and record the gap.
3. **`rate()` on a percentage gauge.** `avg(rate(cgo_cap_deadline_ok_pct[5m]))` (`deploy/prometheus/rules.yml:5`) applies a counters-only function (R11) to a 0–100 percentage (R5 gauge) and then averages the ratio (forbidden by R7). Even if the metric existed, the expression is wrong twice. Percentage should be emitted as a 0–1 ratio gauge and recorded with `avg_over_time` (R11) or aggregated numerator/denominator (R7), never `rate()`.
4. **`CGO_probe_dead` cannot detect absence.** `rate(...)[5m]) == 0` (`deploy/prometheus/rules.yml:14`) fires only when the counter exists and is flat — including the always-`0` static counter (§5.1), so it fires constantly while idle and never when the series is missing (empty vector vs `== 0`, R11). R11's `absent_over_time(...[10m])` is the absence shape. Additionally, a never-incrementing counter makes `rate`=0 the *normal* state, so this alert is inverted: it should be tied to a real heartbeat counter that the bench increments.
5. **TYPE inferred from suffix.** `_total` → counter, else gauge (`pkg/api/metrics.go:26-29`). R1 requires explicit per-metric TYPE (R1); R5 ties type to semantics (monotonic vs up/down). Any future `_total` gauge is mis-declared silently.
6. **No label-value / HELP escaping.** Raw interpolation (`pkg/api/metrics.go:33-34`, HELP at `:25`) violates R1 `escaped_string` (must escape `\`, `"`, LF). Latent: a `site` label with a quote breaks parsing. The builder needs the three replacements, plus a golden test with adversarial labels.
7. **No histogram/summary path.** If p95 is to be a real histogram, R1's `_bucket`/`_sum`/`_count`/`{le="+Inf"}`/ordering rules (R1, R5) have no code. Today p95 doesn't exist at all — the spec's `cgo_small_p95_ms` gauge is unimplemented.
8. **`getBuildInfo` dead / `cgo_build_info` missing.** The version closure (`pkg/api/server.go:397-399`) is accepted and ignored (`pkg/api/metrics.go:43-63`), so the spec's `cgo_build_info{version}` (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:70`) — the `*_build_info` pseudo-metric R6 cites as metadata convention (R6, "Metric names") — never appears. Either emit it or remove the parameter.
9. **Percent/centisecond naming vs base units.** `*_pct` (and spec `*_p95_ms`) vs R6's ratio 0–1 with `ratio` suffix and seconds base unit (R6, "Base Units"). SHOULD-level, but a standalone exporter should rename to `*_ratio` and seconds before first tag, while renaming is free. Changing later is a breaking rename for every PromQL/dashboard.
10. **No `promtool` gate.** Nothing runs `promtool check rules` (R2) or `promtool test rules` (R9) against `deploy/prometheus/rules.yml`; no test file provides `input_series` fixtures. `rules.yml` parse-status is therefore `unverified — primary source not found` *for this file* until R2/R9 are actually run.
11. **Docs and deploy wiring absent.** `/metrics` is undocumented (`docs/api.md:15-116` has no row); `rules.yml` has no Prometheus service / volume / `rule_files` entry (`deploy/docker-compose.yml:1-13`). R8 metamonitoring demands alerts on the monitor itself; an inert rule file + undocumented endpoint cannot meet that. Also `docs/api.md:3` header claim ("Vérifiée contre server.go, sse.go et translate.go") silently excludes this endpoint.

## 6. Target tool: `cgo-exporter` as an independent repo

A separate, tiny repo — not a package inside `cgo` — that owns only the exposition contract and the rules. It reads frozen/live facts and never runs the bench (the bench needs explicit human approval per `AGENTS.md`; the exporter must not be a backdoor to it).

### 6.1 Shape

- **Module**: own `go.mod` with its own module path (R2-Go-modules analogue: path + version is identity; pre-1.0 may change freely, 1.0 is a compatibility promise — general versioning discipline). Zero third-party deps to start (stdlib `net/http` suffices, per §3.1); optionally a `prometheus/client_golang` dependency only if the team prefers delegating exposition to the maintained library instead of fixing the hand-rolled builder — decision recorded in README.
- **Binary**: `cmd/cgo-exporter/main.go` — flags `--listen` (default `127.0.0.1:9100`), `--data-dir` (path to `data/` / audit CSVs), `--version`. One endpoint `GET /metrics` returning `text/plain; version=0.0.4` (R1), plus `GET /healthz`. No other HTTP surface.
- **Collectors** (one per metric family, each bound to a reader): audit counters (`cgo_audit_runs_total{site,link_type,status}`, attempted/measured pair), iface gauge + `_valid` companion, CAP gauges (`cgo_cap_deadline_ok_ratio`, `cgo_cap_sla_breach`, `cgo_cap_sla_measured`), latency (either precomputed `*_p95_seconds` gauge **or** a real histogram — README declares the choice per R4-style API declaration), drops (`cgo_qdisc_drops_total{profile,qdisc,cc,direction}`), `cgo_build_info{version}` (emitted from the version string — the dead `getBuildInfo` param becomes real). **Absent ≠ zero everywhere**: a collector with no input emits nothing for that series and flips the companion `*_measured/_valid` to `0` (R1-valid values) — mirroring `pkg/api/metrics_test.go:37-42`, but at the endpoint, not just the builder (§5.1).
- **Rules**: `rules.yml` shipped in-repo with `promtool check rules` in CI (R2) and a `rules_test.yml` harness (R9) covering: firing on sustained breach, firing on probe absence (`absent_over_time`, R11), non-firing on healthy data, recording-rule ratio correctness (R7). Recording rule renamed to `level:metric:operations` with `without` clause (R7).

### 6.2 Composition

- In this repo it replaces the `pkg/api/server.go:397` hand-off: `pkg/api` keeps its in-process `/metrics` shim (or drops it) and the exporter is scraped as a separate job `cgo-exporter` via `deploy/docker-compose.yml` (new Prometheus service mounting `rules.yml` as `rule_files`). Grafana reads the same Prometheus data source (R10). No dashboard/alert logic lives in the exporter — it only serves exposition (R1) and ships the rule files (R2/R3).

## 7. Slices (proposed, each with acceptance)

1. **S1 — Freeze + declare.** Write the README declaring the public contract: metric names, TYPE per metric (explicit, R1), label sets, unit convention (ratio 0–1, seconds — §5.9 renamed *before* first tag), histogram-vs-gauge choice for p95 (§4.4). Move `pkg/api/metrics.go` builder + `rules.yml` verbatim. Acceptance: README lists every §2 name; `go test ./...` green on the extracted tree.
2. **S2 — Conformance fix.** Explicit TYPE map (kill `_total` inference, R1), label/HELP escaping with golden tests (R1), `cgo_build_info{version}` emitted from `--version` (R6 pseudo-metric). Acceptance: exposition round-trips through a real Prometheus text parser; adversarial label test (`site="a\"b\nc\\"`) passes.
3. **S3 — Bind collectors to evidence.** Wire each collector to `data/link_audit.csv` / snapshot (read-only; never writes frozen data). Delete the six static zeroes (`pkg/api/metrics.go:46-53`). Acceptance: scrape with no audits emits **no** `cgo_audit_runs_total` series (absent, not zero); after one stub audit it emits `... 1`.
4. **S4 — Fixing rules + emit missing metrics.** Either emit `cgo_cap_deadline_ok_ratio` + `cgo_cap_sla_breach` measured (CAP SLA from `pkg/metrics.AssessCAPSLA`) or delete `deploy/prometheus/rules.yml:4-12`. Replace `avg(rate(%))` with numerator/denominator record + `without` clause (R7); replace `rate(...)==0` with `absent_over_time` (R11). Acceptance: `promtool check rules rules.yml` exit 0 (R2); `promtool test rules rules_test.yml` green (R9).
5. **S5 — Histogram decision (if chosen).** If p95 is a histogram, implement `_bucket{le}`/`_sum`/`_count` with mandatory `{le="+Inf"}` (R1) and increasing bucket order; else document the gauge choice. Acceptance: `histogram_quantile` test in `rules_test.yml` (R9) computes the expected p95.
6. **S6 — First tag + wiring.** Tag `v0.1.0`, add Prometheus service to `deploy/docker-compose.yml` mounting `rules.yml`, document `/metrics` in `docs/api.md`. Acceptance: fresh `docker compose up` → scrape `/metrics` → `promtool check rules` + a synthetic firing in Grafana over the Prometheus source (R10).

*Each slice touches only the exporter repo + the two files named in the brief; bench/probe code is untouched. Slices 2–4 change wire behaviour and get verifier-re-run before a `VERIFIED` row is claimed.*

## 8. Anti-toy criteria (what would make this unacceptable as a standalone tool)

1. **No fabricated zeroes.** A scrape with no input series for a metric emits no sample for it (absent ≠ zero, `pkg/api/metrics.go:15-16` honoured at the endpoint). Any hard-coded `"0"` remaining outside the `*_measured/_valid` companion flags fails review — that is the toy signal.
2. **Every rule expression names an emitted metric.** `promtool check rules` (R2) passes *and* every metric in every `expr` appears in a real scrape of `GET /metrics`. A rule wired to a dead metric is a dead alert, and R8 metamonitoring (R8) is the standard it fails.
3. **Types are declared, not inferred.** Each TYPE is hand-set per R5 semantics (counter monotonic, gauge up/down), never derived from a name suffix. A `_total` gauge mislabelled as counter is a scrape that lies to every `rate()` (R11).
4. **Escaping + uniqueness are tested, not hoped.** Golden tests with `"`/`\`/newline in label values and docstrings (R1) and a uniqueness assertion (R1 "Grouping and sorting"). Without them the exposition is one label away from unparsable.
5. **Histogram/summary honesty.** No fake histogram: if a histogram is exposed it has `_bucket{le="+Inf"} == _count` in increasing order (R1); if p95 is precomputed it is named as such and unit-correct (R6). A histogram implementation copied from the gauge code is a toy.
6. **Unit-correct names.** Ratios in 0–1 with `ratio` suffix, seconds not milliseconds (R6). Shipping `*_pct`/`*_ms` and documenting "we know, it's SHOULD" at 1.0 is the frozen-corpus rule: correct it before the tag.
7. **`promtool` is a gate, not an aspiration.** CI runs `promtool check rules` (R2) and `promtool test rules` with firing/non-firing fixtures (R9). A rules file that has never been parsed by `promtool` is §5.10 forever.
8. **Independent consumer + independent scrape.** At least one out-of-tree Prometheus scrape (not the in-repo test) succeeds against `cgo-exporter`, and one real audit drives a counter from absent to `1` in that scrape. If the only proof is `pkg/api/metrics_test.go`'s string-contains checks, the tool is decoration on a hand-rolled builder, not an exporter.

---
*Teams checklist: every quantitative claim traces to `file:line` or R1–R11. Claims that could not be traced to a fetched primary are marked `unverified — primary source not found` inline (§4.2 auth, §4.5 intra-metric order, §4.6 `rules.yml` parse status, §4.7 Grafana sub-pages). Slices 1–6 are proposals in this research file only — no `pkg/api/metrics.go` or `rules.yml` was modified in this pass.*
