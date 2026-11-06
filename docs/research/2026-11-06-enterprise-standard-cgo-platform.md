# Enterprise-Standard CGO Platform — Primary-Source Research

**Question:** What does a standardized, enterprise-grade reference architecture for CGO (Météo Madagascar / DGM, Master 2 thesis context) look like when judged by primary sources, not secondary write-ups?

**Date:** 2026-11-06
**Scope:** Research only. No code modified, no commit. Repo root: `operator-cockpit` (Meteolink / cgo).
**Rule applied:** every factual claim cites a primary source (URL or repo file). Claims that could not be sourced are labelled `unverified — primary source not found` explicitly. No unsourced marketing numbers.

---

## 1. Method & sources (table: claim → source → consulted date)

Consulted 2026-11-06 unless otherwise noted. First-party specs and official docs only; blogs, vendor marketing, and pasted write-ups were treated as claims to verify, not sources.

| # | Claim / topic | Primary source consulted | Consulted |
|---|---|---|---|
| R1 | Repo: single Go binary, `go:embed` React dashboard, no DB/service deps | `README.md:36-37` ("Un seul binaire… dashboard React embarqué via go:embed. Aucune base de données, aucune dépendance de service") + `docs/architecture.md:1-4` + `go.mod:1-10` | 2026-11-06 |
| R2 | Repo: frozen corpus counts, CSV schemas, never-touch rules | `docs/VERIFY.md:24` (159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells, VERIFIED) + `AGENTS.md` (§ Never touch) | 2026-11-06 |
| R3 | Repo: gates, ledger discipline (`VERIFIED/RECORDED/PENDING/BLOCKED/REFUTED`) | `docs/VERIFY.md:1-18`, `docs/VERIFY.md:83-91` (how a row moves) | 2026-11-06 |
| R4 | Repo: API surface (routes, validation, 501/409/400 semantics, 1 MiB cap, `validRunID`) | `docs/api.md:1-12` + `pkg/api/server.go:277-312` (`shapeApply`), `pkg/api/server.go:316-320` (`validRunID`), `pkg/api/server.go:611-742` (`/api/run/start`) | 2026-11-06 |
| R5 | Repo: SSE hub (10 Hz ticker, delta compression, 2048 ring, 64 subs, replay, backpressure) | `docs/architecture.md:23-41` + `README.md:27-29` + `docs/api.md:21` | 2026-11-06 |
| R6 | Repo: tc/qdisc topology (netem root 1: + child shaper, reset-then-apply, `ApplyShaper` fallback) | `pkg/qdisc/qdisc.go:1-13,66-143` + `docs/architecture.md:43-58` | 2026-11-06 |
| R7 | Repo: dual-namespace bench (4 netns, 3 veth pairs, netem-only WAN, CAKE-only edge, offload fail-closed, mgmt veth 10.99.0.0/30) | `kit/cgo-bench-dualns.sh:1-39,42-55,115-256,258-341` + `kit/run-dualns.sh:1-72` + `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:28-29,69-112` | 2026-11-06 |
| R8 | Repo: audit phases, directions, honesty invariants (no-load⇒no-grade, unmeasured⇒Unavailable) | `pkg/audit/audit.go:15-38,40-145,204-717` + `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:40-67` | 2026-11-06 |
| R9 | Repo: remediation (measured-only, 95% shaping, Linux tc + RouterOS v7 CAKE strings, IFB, `Validate`) | `pkg/remediate/remediate.go:1-101,131-260,262-384` | 2026-11-06 |
| R10 | Repo: results aggregation (profile×qdisc×cc×direction, valid-only, quarantine excluded) | `pkg/results/results.go:12-47,65-236` | 2026-11-06 |
| R11 | Repo: campaign matrix, resume, tier freeze, 409 on tier change mid-campaign | `pkg/campagne/matrix.go:24-258` + `pkg/api/server.go:460-472` + `docs/VERIFY.md:66` | 2026-11-06 |
| R12 | Repo: economics (GOLD 1M Ar/mo 200 Mb/s CIR, PLATINE 2.5M Ar/mo 30 TB, APN 150 Ar/GB quoted, wholesale 81.38 Ar/GB, emergency 5556 Ar/GB, DGM 220 ms / 95%) | `pkg/metrics/cost.go:1-251` + `docs/data-prices.md` (cited) + `docs/VERIFY.md:48` | 2026-11-06 |
| R13 | Repo: security posture (no authN/Z, no CSRF/rate-limit, Sec-Fetch-Site guard only, mgmt bind) | `docs/deploy.md:159-177` + `docs/api.md:3-11` | 2026-11-06 |
| R14 | Repo: deploy (systemd unit, `setcap cap_net_admin`, releases via GoReleaser, `cgo kit` 21 actions) | `docs/deploy.md:55-157` + `.goreleaser.yaml:1-64` + `README.md:133-153` | 2026-11-06 |
| R15 | Repo: frontend (4 modes, lazy Bench/Audit/Evidence, SSE connect, bundle budget 600 KB gz) | `web/frontend/src/App.tsx:1-178` + `docs/VERIFY.md:55` (349.9 KB gz initial 99.1 KB) | 2026-11-06 |
| A1 | FQ-CoDel / CoDel semantics (DRR + CoDel, 1024 queues default, target 5 ms, interval 100 ms, quantum 1514 B, limit 10240, hash salt, TSO/GRO warning) | RFC 8290 (IETF, Experimental) §1.3, §2–§5 — https://www.rfc-editor.org/rfc/rfc8290.txt (fetched full text 2026-11-06) | 2026-11-06 |
| A2 | CAKE requirements (Linux tc obligations, ingress mode, ack-filter, diffserv, nat, triple-isolate) | Linux kernel CAKE man/doc — **unverified — primary source not found** (repo cites `pkg/remediate/remediate.go:148-166` and bench script; no kernel doc fetched this session) | 2026-11-06 |
| A3 | RouterOS v7 `kind=cake` + `/queue tree` + LAN egress for download | MikroTik docs — **unverified — primary source not found** (fetch to `help.mikrotik.com/docs/spaces/ROS/pages/8396805/CAKE` returned 404; repo strings at `pkg/remediate/remediate.go:262-384` are claims, not primary verification) | 2026-11-06 |
| A4 | WMO WIS 2.0 + CAP 1.2 spec (alert lifecycle, publish/discover, SLA) | WMO / OASIS CAP — **unverified — primary source not found** (fetch to `docs.oasis-open.org/cap/v1.2/cap-v1.2-spec.html` returned 404; repo only cites `cost.go:75-79`: WMO-No. 1061 "a few seconds to a few minutes", no 220 ms mandate — that citation itself is unverified this session) | 2026-11-06 |
| O1 | Prometheus exposition contract (HELP/TYPE, metric_name{labels}, histogram/summary conventions, scrape negotiation, Content-Type required) | Prometheus official docs — https://prometheus.io/docs/instrumenting/exposition_formats/ (fetched 2026-11-06) | 2026-11-06 |
| O2 | Prometheus scraping, retention, alerting, Grafana dashboards, Blackbox/Node Exporter | **unverified — primary source not found** (only exposition format fetched; scraping/retention/alerting docs not fetched this session) | 2026-11-06 |
| M1 | MQTT/Mosquitto (topics, QoS, TLS, auth) | Eclipse Mosquitto docs index — https://mosquitto.org/documentation/ (fetched index 2026-11-06; `mosquitto.conf`, `mqtt`, auth pages linked from index but not fetched) — **partial** | 2026-11-06 |
| N1 | Nginx rate-limiting (leaky bucket, `limit_req_zone`, `limit_req`, burst/nodelay, 503 default) | Nginx official docs — https://nginx.org/en/docs/http/ngx_http_limit_req_module.html (fetched 2026-11-06) | 2026-11-06 |
| S1 | cgroups / CPU bandwidth control (quota/period/burst, throttling stats, hierarchical caveats) | Linux kernel docs — https://www.kernel.org/doc/html/latest/scheduler/sched-bwc.html (fetched 2026-11-06, pages note v2 in `Documentation/admin-guide/cgroup-v2.rst`) | 2026-11-06 |
| S2 | Docker / OCI image-content requirements, NIST/ISO/IEC 25010 quality model | **unverified — primary source not found** (no Docker/NIST doc fetched this session) | 2026-11-06 |
| P1 | Pasted proposal (CHR+QEMU, bench dual-ns, Mosquitto, Nginx, Prometheus/Grafana, cgroups) + claims (RAM footprints, "microsecond" probing, 3-file scope) | Task brief only. **No pasted-context file found in repo** (glob `**/*.txt` returned only `kit/chr/export-*.txt` fixtures + tooltip shot). Evaluated as task-described claims against repo evidence; specific numbers are **unverified — primary source not found** | 2026-11-06 |

**Method note:** Repo ground truth was established by reading files directly, not by trusting task-brief claims. Where an external doc could not be fetched (404/transport), the requirement is marked unverified rather than paraphrased from memory. Where a fetch succeeded, the table gives the exact URL and section.

---

## 2. Ground truth in repo (evidence)

**Product identity.** Meteolink / cgo is a Go single binary with the React dashboard embedded via `go:embed` (`README.md:36-37`, `docs/architecture.md:1-4`). Go 1.25, stdlib-only HTTP/SSE surface, no web framework (`docs/architecture.md:95-103`). Releases: `.goreleaser.yaml:1-64` (cgo + meteolink, linux/darwin/windows, tar.gz/zip, sha256 checksums). Version state: `VERSION` referenced in README badge 1.3.5 (`README.md:1`), API examples still show 1.2.3 (`docs/api.md:17`) — version string drift exists but is cosmetic.

**Frozen corpus.** 159 runs / 528 lines / 208 valid / 74 degraded / 246 invalid / 22 cells, VERIFIED 2026-10-27 (`docs/VERIFY.md:24`). Seal anchor `docs/frozen-manifests.sha256`: 159 manifests, VERIFIED 2026-10-28 (`docs/VERIFY.md:26`). `book/` is gitignored with 0 tracked files (`docs/VERIFY.md:25`) — evidence, not scratch space per `AGENTS.md`.

**Measurement stack.** Audit (`pkg/audit/audit.go`) probes RTT (ping), small-object completion (16 KiB via `probe.SmallObject`), bulk load (BBR, 4 flows default `audit.go:269-271`, directions up/down/both `audit.go:272-278`). Results freeze to `data/runs/<run>/aqm_eval.csv` + `manifest.json` SHA-256 (`docs/architecture.md:76-83`). `pkg/results/results.go:65-236` groups by profile×qdisc×cc×direction, excludes `invalid` from medians, counts degraded separately, exposes valid-only medians + bootstrap CIs. Campaign matrix (`pkg/campagne/matrix.go`) iterates profiles×qdiscs×ccs×reps; gaps are resumable by run_id; skip isolates to current cell (`matrix.go:213-250`).

**Shaping.** `pkg/qdisc/qdisc.go:66-143`: `Reset` then `ApplyNetem` (root `1:`) then `ApplyShaper` as child `1:`/fallback root; CAKE carries `bandwidth`+`rtt`, fq_codel/pfifo under TBF. Dual-ns bench (`kit/cgo-bench-dualns.sh`) splits physics (netem on `w-edge`/`c-wan`) from policy (CAKE on `e-srv`/`e-wan` or IFB), enforces GRO/GSO/TSO off with fail-closed `check_offloads` (`cgo-bench-dualns.sh:241-248`). Mgmt veth `m-host 10.99.0.1 ↔ m-client 10.99.0.2` carries the dashboard, never measured traffic (`run-dualns.sh:57-63`, `cgo-bench-dualns.sh:154-163`).

**Remediation.** `pkg/remediate/remediate.go:83-101` refuses unmeasured/NaN/zero capacity/RTT; shaping at 95% of measured rate (`remediate.go:66-74`); Linux egress CAKE `ack-filter nat diffserv4 triple-isolate` (`remediate.go:148-166`), IFB ingress CAKE `dual-dsthost` w/o ack-filter (`remediate.go:225-260`); RouterOS v7 `kind=cake` strings with explicit sfq-fallback labelling for fq_codel (`remediate.go:300-310`), download via LAN egress `/queue tree` because RouterOS has no IFB (`remediate.go:324-384`).

**API / security.** JSON bodies capped at 1 MiB via `http.MaxBytesReader` (`docs/api.md:3-4`, used e.g. `pkg/api/server.go:456,484`). Path traversal blocked by `validRunID` (no `/\ .`) (`pkg/api/server.go:316-320`). Mutations reject `Sec-Fetch-Site: cross-site` for browsers only (`docs/api.md:5-6`). Observe mode returns 501 for shape/run/watch/burst (`docs/api.md:8-11`, `pkg/api/server.go:218-229`). **No authN/Z, no CSRF token, no rate limit** — stated honestly in `docs/deploy.md:159-177`: "pas d'authentification ni d'autorisation… pas de jeton CSRF ni de rate limit… jamais exposé sur Internet." `docs/deploy.md:140-143` instructs opening port 9090 on the admin network — an explicit lab/DSI boundary that is a GAP for enterprise (see §4).

**Economics.** `pkg/metrics/cost.go`: PlanGOLD 1,000,000 Ar/mo 200 Mb/s CIR (`cost.go:21`), PlanPLATINE 2,500,000 Ar/mo 30 TB FUP (`cost.go:23`), wholesale 81.38 Ar/GB derived (`cost.go:54`), APN 150 Ar/GB `Quoted:true` (`cost.go:56`), emergency `RetailEmergencyARPerGB = 5556` (`cost.go:62`). DGM objective `CAPDeadlineMs = 220`, `CAPBreachThresholdPct = 95.0` (`cost.go:80-83`), explicitly **not** a WMO mandate (`cost.go:75-79`, enterprise spec §2 invariant 7). Tier freezes per run; mid-campaign `POST /api/cost/tier` answers 409 (`pkg/api/server.go:460-472`).

**Observability today.** Live SSE at 10 Hz with delta compression, 2048-frame ring, 64-subscriber cap + 503, `Last-Event-ID` replay, backpressure events (`docs/architecture.md:23-41`). Operator journal: 50-event ring at `GET /api/events` (`docs/architecture.md:90-91`). Interface sampling: `CGO_IFACE`-gated `iface_*` fields with `iface_valid`/`iface_gap` semantics (`docs/api.md:28-36`). **There is no `/metrics` Prometheus endpoint, no alerting rules, no Grafana dashboards, no Blackbox/Node exporters in the repo** — verification by absence: no such files found in `pkg/api/server.go` route list (routes enumerated `docs/api.md:13-116`; `/metrics` absent), no `prometheus`/`grafana` in `go.mod` deps (`go.mod:1-60`). Retained live window is 180 s / 1800 pts at 10 Hz (`README.md:159`) — ephemeral, not long-term retention.

**Blocked / unverified live claims (ledger).** Linux live acceptance G0–G7: BLOCKED (`docs/VERIFY.md:28`). eBPF retransmit/RTT/qdisc drops: BLOCKED (`docs/VERIFY.md:29`). IFB ingress live acceptance: BLOCKED (`docs/VERIFY.md:31`). BBRv3 3×3: BLOCKED (`docs/VERIFY.md:33`). `.dist` netem: BLOCKED (`docs/VERIFY.md:34`). CHR (RouterOS) compliance: BLOCKED — parser/proof writer tested on fixtures, nothing booted (`docs/VERIFY.md:57`). Case study (SSH to forecasting server): BLOCKED awaiting operator grant (`docs/VERIFY.md:60`). These BLOCKED rows are load-bearing for §5: the proposal cannot claim them as proven.

---

## 3. Primary-source requirements by dimension

For each dimension: what the primary source **REQUIRES** (not what a blog says). Citations are exact. Where the source was not fetched, the row says so — the requirement is then unverified and must not be assumed.

### 3.1 Reliability / SLO

- **RFC 8290 (FQ-CoDel, Experimental)** does not define SLOs; it defines algorithm parameters the deployment must choose: `target` default 5 ms (tune ≥ MTU transmission time, e.g. ~15 ms at 1 Mbps), `interval` default 100 ms (SHOULD be on order of worst-case RTT), `limit` global packet cap (default 10240; "some packet limit MUST be enforced" §5.2.3), `quantum` default 1514 B, `flows` default 1024 (§5.2). Security: hash MUST be salted (32-bit random in Linux) to prevent targeted DoS (§8). **Implication:** enterprise CGO MUST pin these four values in config and MUST fail-closed if `limit` is unset; SLOs (e.g. 220 ms) are operator objectives layered on top, not RFC mandates.
- **Enterprise SLO / NIST / ISO quality model:** **unverified — primary source not found.** No ISO/IEC 25010 or NIST doc fetched. Repo-internal proxy: `pkg/metrics/cost.go:80-83` (220 ms / 95%) + `pkg/audit/audit.go:114-122` (CAP SLA measured only from loaded samples) + six gates (`docs/VERIFY.md:42`). Treat 220 ms/95% as **operator objective**, never as WMO requirement, until primary WMO text is fetched.

### 3.2 Security / authN/Z

- **Deploy boundary (repo honesty):** `docs/deploy.md:159-177` requires: mgmt-network-only exposure, never internet-facing, because shaping mutates the edge. **What a primary security source would require** (OWASP/CWE/NIST 800-207 — not fetched): authenticated sessions, CSRF tokens on state-changing POSTs, rate limiting, TLS with verifiable identity. CGO today enforces only `Sec-Fetch-Site: cross-site` browser guard (`docs/api.md:5-6`) — explicitly "contournable par un client direct" (`docs/deploy.md:165-170`). **Requirement gap is therefore primary-sourced by the repo itself.**
- **Nginx `limit_req` (fetched):** rate limiting uses leaky-bucket `limit_req_zone` + `limit_req zone=… burst=… [nodelay|delay=…]`; excessive requests are delayed until burst is exceeded, then terminated (default 503 via `limit_req_status`). Shared-memory zone, LRU eviction on exhaustion. **Requirement:** any enterprise ingress MUST place a `limit_req_zone` in front of `POST /api/run/*`, `/api/shape`, `/api/audit/*`, `/api/burst`, `/api/cost/tier` with named r/s budgets; "no rate limit" is non-compliant.

### 3.3 Observability

- **Prometheus exposition (fetched):** text format requires `# HELP`/`# TYPE` (counter|gauge|histogram|summary|untyped), `metric_name{labels}` lines, unique (name,labels) pairs, HELP/TYPE before first sample, histogram buckets `le="+Inf"` equal to `x_count`, summary `quantile` labels. Negotiation via `Content-Type`; **Prometheus 3.0 scrape fails if Content-Type is missing/unparsable/unsupported**. Targets MUST return valid metrics Content-Type. **Requirement:** a `/metrics` endpoint MUST emit HELP/TYPE, MUST NOT emit duplicate label sets, MUST set `text/plain; version=0.0.4` (or OpenMetrics `application/openmetrics-text`) correctly.
- **Scraping, retention, alerting, Grafana, Blackbox/Node Exporter:** **unverified — primary source not found** (not fetched). Common enterprise expectation (scrape interval, retention, alert rules, dashboards) cannot be cited this session; §6 marks them SHOULD with explicit re-verification note.
- **Repo contract to preserve:** `docs/api.md:28-36` `iface_*` fields with `valid:false` invalidating numerics; `docs/architecture.md:36-41` backpressure accounting. Any Prometheus mapping MUST preserve "false means not measured" semantics — Prometheus `NaN` in exposition maps to absent, never 0.

### 3.4 Scalability / multi-site

- **cgroups bandwidth control (kernel docs, fetched):** quota/period managed via cgroupfs; `-1` = unconstrained; period bounds 1 ms–1 s; burst ≤ quota; throttling statistics in `cpu.stat` (`nr_periods`, `nr_throttled`, `throttled_time`, `nr_bursts`, `burst_time`); hierarchical quota enforcement where child can be throttled by parent consumption. **Requirement:** containerized bench/dashboard MUST set `cpu.max`/`memory.max` (cgroup v2 — v1 files shown in fetched page; v2 doc linked, not fetched — **partial**), MUST export throttling stats to `/metrics`, MUST NOT oversubscribe parent quota silently.
- **Multi-site / MQTT fan-out (Mosquitto index fetched, pages not):** topics/QoS/TLS/auth specifics **unverified — primary source not found** except index existence of `mosquitto.conf`, `mqtt`, `Authentication methods`, `Dynamic Security`, `Persistence`. Index proves these are first-party configurable areas, not their required values. **Requirement:** multi-site aggregation MUST define topic namespace, QoS, retained-message policy, TLS client-cert policy from those pages — to be fetched before implementation. Until then, DEFERRED (§6).
- **Repo scale today:** single-host binary, 64 SSE subs cap (`pkg/api/server.go:314`), in-memory live ring 180 s (`README.md:159`), file-backed `data/runs` scans (`pkg/results/results.go:66` glob). **No sharding, no federated aggregation, no horizontal scaling in repo** — enterprise multi-site is therefore a GAP, not an assumption.

### 3.5 Operability / deploy / rollback

- **Repo deploy (primary):** `docs/deploy.md:89-97` systemd unit via `cgo service install`; `docs/deploy.md:77-87` privileges via `setcap cap_net_admin+ep` (preferred) or root; `docs/deploy.md:101-104` shape-restore at startup. Releases carry `checksums.txt` SHA-256 (`.goreleaser.yaml:49-51`). Kit: 21 actions, `cgo kit deploy` binaire-first idempotent (`docs/deploy.md:108-121`).
- **OCI/compose/Ansible, rollback, blue-green:** **unverified — primary source not found** (no OCI/Docker/Ansible doc fetched; no compose/Dockerfile/Ansible in repo glob — verified by absence in `go.mod`, `.goreleaser.yaml`, `kit/` listing). **Requirement:** enterprise deploy MUST ship OCI image + digest pin + compose/systemd unit + idempotent `kit/` scripts + `cgo verify` + `cgo shape --restore` rollback hooks; current release-tarball + systemd path is the only primary-grounded baseline.

### 3.6 Supportability

- **Repo supportability (primary):** six gates exit 0/1/2 discipline (`AGENTS.md`: "Exit 0 = every gate passed. Exit 1 = a named gate failed. Exit 2 = a gate could not run; 2 is not green"). Hooks block secrets/`book/` paths, refuse push to main (`docs/VERIFY.md:43`). Ledger promotion only by named command on named host+date (`docs/VERIFY.md:83-91`). Runbooks present: `docs/runbooks/2026-10-30-host-acceptance.md` (listed). **Requirement:** supportability MUST keep: ledger + gates + hooks + deploy checklist (`docs/deploy.md:149-157`) + `cgo doctor` / `GET /api/doctor` capability matrix.

### 3.7 Data provenance

- **Repo provenance (primary):** per-run `manifest.json` with `file`+`sha256`, `hash8` in Wall/Résultats/Provenance/report (`README.md:45-46`, `docs/architecture.md:76-83`, `pkg/figures` `ProvenanceSHA`). Corpus seal `docs/frozen-manifests.sha256` (`docs/VERIFY.md:26`). Audit dual seals: SHA-256 of `telemetry.csv` + appended `link_audit.csv` row file (enterprise spec §5). `invocation` freezes effective audit settings (`pkg/audit/audit.go:66-68` + `docs/api.md:100`). `validRunID` + `run-smoke` exclusion (`pkg/results/results.go:88-90`) prevent fixture pollution.
- **WIS 2.0 provenance / CAP lifecycle:** **unverified — primary source not found.** CGO's provenance is repo-internal and strong for measurement evidence; it makes **no** WIS 2.0 publication claim that can be verified this session. WIS/CAP integration MUST therefore be DEFERRED until WMO primary docs are fetched (§8).

---

## 4. Gap analysis: current CGO vs required

**Legend:** PASS = requirement met with evidence. GAP = partially met / missing enterprise hardening. FAIL = contradicted by evidence or missing non-negotiable.

| Dimension | Verdict | Evidence |
|---|---|---|
| Reliability/SLO | **GAP** | Measurement honesty PASS: `pkg/audit/audit.go:488-565` (no grade without loaded measurement), `pkg/campagne/matrix.go` gates G0–G7, `docs/VERIFY.md:28-34` BLOCKED rows honestly held. Enterprise SLO GAP: no SLO definition outside 220 ms operator objective (`pkg/metrics/cost.go:75-83`); no `limit_req` / SLO alert rules; 180 s live ring not a multi-site SLO store (`README.md:159`). FQ-CoDel params GAP: repo does not pin `target/interval/limit/quantum/flows` per RFC 8290 §5.2 in `pkg/qdisc/qdisc.go` (defaults inherited from kernel, not declared). |
| Security/authN/Z | **FAIL** | `docs/deploy.md:159-177` self-declares: no authN/Z, no CSRF, no rate limit, shaping mutable by anyone reaching the port. `docs/api.md:5-6` guard is browser-only. Nginx `limit_req` requirement (fetched primary) unmet. No TLS-by-default outside bench HTTP (`kit/run-dualns.sh:57-63` `--tls=false`). For enterprise multi-user, this is a release blocker, not a nit. |
| Observability | **GAP** | SSE/journal PASS internally (`docs/architecture.md:23-41,90-91`). Enterprise GAP: no `/metrics` (Prometheus contract unmet — no HELP/TYPE emission anywhere in `pkg/api/server.go` routes), no retention, no alertmanager rules, no dashboards, no Blackbox/Node exporters (verified by absence in route list `docs/api.md:13-116` + `go.mod`). `iface_*` honesty model (`docs/api.md:28-36`) is a strong local contract to preserve in any exporter. |
| Scalability/multi-site | **GAP** | Single-host PASS for lab: dual-ns bench 4 netns (`kit/cgo-bench-dualns.sh:42-55`), 64-sub cap (`pkg/api/server.go:314`), in-memory rings. Multi-site GAP: no aggregation, no MQTT, no federation, no cgroup limits (no `cpu.max`/`memory.max` in repo). Scale ceiling is one bench host + operator laptop. |
| Operability/deploy/rollback | **GAP** | Deploy PASS for lab: systemd (`docs/deploy.md:89-97`), `setcap` (`docs/deploy.md:81-83`), checksummed releases (`.goreleaser.yaml:49-51`), idempotent kit (`docs/deploy.md:108-121`), startup `shape --restore` (`docs/deploy.md:99-104`). Enterprise GAP: no OCI image, no compose, no Ansible, no blue-green/rollback beyond binary replace, no digest pinning. `docs/deploy.md:140-143` port-9090 opening is the only scaling story. |
| Supportability | **PASS (lab) / GAP (enterprise)** | Ledger+gates+hooks PASS and unusually strong (`docs/VERIFY.md:1-18,42-43,83-91`). Enterprise GAP: runbook coverage is host-acceptance only; no multi-site incident, backup/restore (beyond corpus second copy `docs/VERIFY.md:37`), or on-call SLO runbooks. |
| Data provenance | **PASS (measurement) / DEFERRED (WIS)** | Measurement provenance PASS: freeze+manifest+hash8+dual seals+invocation+quarantine honesty (`docs/architecture.md:76-83`, `pkg/audit/audit.go:66-68`, `pkg/results/results.go:137-162`). WIS/CAP provenance DEFERRED: **unverified — primary source not found**; no WIS publication path exists in repo. |

**Bottom line:** CGO is a credible **lab-grade measurement platform** with exceptional evidence hygiene. It is **not enterprise-grade** today because authN/Z + rate limiting + operability packaging + Prometheus observability are missing — precisely the headline enterprise blockers, all primary-sourced by the repo's own honesty docs.

---

## 5. Gap analysis: pasted proposal vs required

Proposal as described in task: CHR+QEMU, bench dual-ns, Mosquitto, Nginx, Prometheus/Grafana, cgroups. Claims flagged: RAM footprints, "microsecond" probing, 3-file scope. **No pasted file exists in repo** (glob evidence §1/P1). Each element is therefore judged against repo evidence, not against an unseen write-up.

| Proposal element | Verdict | Evidence / reason |
|---|---|---|
| Bench dual-ns (4 netns, decoupled netem/CAKE, offload fail-closed) | **PASS — already repo truth** | `kit/cgo-bench-dualns.sh:1-256` + enterprise spec §3. The proposal adds nothing new here; it restates the strongest part of CGO. No vendor prose needed. |
| CHR + QEMU (RouterOS validation) | **GAP — not proven** | Current state: parser + proof writer + fixture replay only; nothing booted (`docs/VERIFY.md:57` BLOCKED). A proposal claiming CHR validation as done would be **false**. QEMU+CHR is a reasonable SHOULD for RouterOS CAKE verification, but current RouterOS strings (`pkg/remediate/remediate.go:268-384`) remain **unverified vs primary MikroTik docs** (404 this session). Requires: CHR image procurement + human approval (§8). |
| Mosquitto (multi-site ingestion) | **DEFERRED** | No MQTT in repo; index fetch proves Mosquitto has TLS/auth/QoS pages but their required values are **unverified** this session. Adding Mosquitto before authN/Z on CGO itself (§4 FAIL) would compound exposure: a second unauthenticated plane. Must sequence **after** authN/Z + rate limiting (see §7). |
| Nginx (TLS termination / reverse proxy / `limit_req`) | **MUST — closes a FAIL** | Directly addresses the `docs/deploy.md:159-177` gap: `limit_req_zone` leaky-bucket rate limiting (primary fetched, §3.2) + TLS + single ingress. This is the cheapest enterprise win: no Go change, closes rate-limit + exposure findings. Recommendation: Nginx in front of every 9090. |
| Prometheus/Grafana (+ Blackbox/Node Exporter) | **MUST — closes a GAP** | Repo has no `/metrics` (§4). Prometheus exposition contract is primary-fetched (§3.3). Proposal direction is correct, but scraping/retention/alerting specifics are **unverified** this session and MUST be fetched before pinning numbers. Grafana dashboards SHOULD reuse existing `iface_*` honesty (`docs/api.md:28-36`) — never render absent as 0. |
| cgroups resource isolation | **SHOULD — cheap, standard** | Kernel docs fetched (§3.4). Repo has no cpu/memory limits today. Setting `cpu.max`/`memory.max` on dashboard + bench processes is low-risk and expected by any OCI deploy. Low priority vs authN/Z. |
| "Microsecond" probing claim | **REJECT — unverifiable as stated** | Repo probes are RTT (ping, ms), small-object completion (ms), bulk goodput (Mbit/s) — `pkg/audit/audit.go:40-55`, `pkg/probe/` — all ms-scale. No microsecond-resolution probe exists in repo evidence. "Microsecond" would require eBPF/timestamping RFC 8290 §5.5 territory; eBPF is BLOCKED (`docs/VERIFY.md:29`). Any microsecond claim is **unverified — primary source not found** and contradicts the ms-scale evidence chain. |
| RAM footprint claims ("X MB per service") | **REJECT — unverifiable** | No memory measurements in repo; no cgroup memory numbers, no OCI resource manifest. Exact footprints without a `memory.max` manifest + measurement are marketing. **Unverified — primary source not found.** §6 instead requires declaring `memory.max`/`cpu.max` and measuring, not asserting. |
| "3-file scope" claim | **REJECT — false** | Repo footprint is orders of magnitude beyond 3 files: `pkg/api/server.go` alone is 1300+ lines, plus `pkg/audit`, `pkg/campagne`, `pkg/qdisc`, `pkg/remediate`, `pkg/results`, `pkg/metrics`, `pkg/probe`, `web/frontend`, `kit/`, `tools/verify.py`, `docs/` ledgers. A 3-file characterization is factually false against the tree. |

**Overall:** Proposal direction is sound on dual-ns/Nginx/Prometheus, but it overstates CHR readiness, understates authN/Z sequencing, and contains three factually false/unverifiable quantitative claims. The standardized target in §6 keeps the sound parts, drops the false claims, and sequences the rest.

---

## 6. Standardized target architecture (MUST / SHOULD / DEFERRED)

Names are literal. Numbers are pinned to primary sources above or marked `unverified — primary source not found` and excluded from MUST.

### 6.1 MUST (non-negotiable for enterprise)

1. **AuthN/Z boundary in front of all mutation APIs.** Exact contract: every `POST /api/run/*`, `/api/shape`, `/api/watch`, `/api/burst`, `/api/audit/*`, `/api/cost/tier`, `/api/profile/*`, `/api/figures/regen` requires an authenticated session (mTLS or OIDC bearer); `GET` observation APIs require authentication on any non-loopback bind. CSRF tokens on all state-changing POSTs (cookie sessions) or same-site + bearer for API clients. Rationale: closes `docs/deploy.md:159-177` FAIL. Bearer validation MUST reuse `pkg/api/server.go` handler seams (`newHandler`) — no parallel auth server.
2. **Rate limiting at ingress.** Nginx `limit_req_zone $binary_remote_addr zone=cgo:10m rate=…` + per-location `limit_req` for mutation endpoints (RFC primary, §3.2). Default policy proposal: reads `30r/s burst=10`, mutations `1r/s burst=3 nodelay` returning 503. Exact r/s to be calibrated against measured campaign rates — **rates themselves unverified until load-tested**, mechanism MUST ship first.
3. **Prometheus `/metrics` honoring exposition contract.** Minimal metric contract (names + labels, all MUST carry `site` + `bench` + `run_id` where applicable):
   - `cgo_audit_runs_total{site,link_type,status}` (counter: started/completed)
   - `cgo_audit_load_attempted_total{site}` + `cgo_audit_load_measured_total{site}` (honesty pair — attempted vs measured, `pkg/audit/audit.go:124-127,488-492`)
   - `cgo_small_p95_ms{site,profile}` histogram or `cgo_rtt_p95_ms{site,direction}` gauge (units in name per Prometheus best practice — **naming strength unverified for this project, MUST keep `_ms` suffix for drift safety**)
   - `cgo_bloat_delta_ms{site,direction}` gauge; `cgo_jfi_pct{site}` gauge with `jfi_measured` as separate `cgo_jfi_measured{site}` gauge (never encode absence as 0 — `pkg/audit/audit.go:91-93`)
   - `cgo_cap_deadline_ok_pct{site}` gauge + `cgo_cap_sla_breach{site}` gauge (0/1) + `cgo_cap_sla_measured{site}` gauge (`pkg/metrics/cost.go:97-112`)
   - `cgo_qdisc_drops_total{profile,qdisc,cc,direction}` counter (from `pkg/qdisc/stats.go` leaf stats)
   - `cgo_iface_rx_mbps`/`cgo_iface_tx_mbps`/`cgo_iface_drops_total` with `cgo_iface_valid` boolean companion (`docs/api.md:28-36` — MUST preserve `valid:false ⇒ ignore numerics`)
   - `cgo_run_status{run_id,profile,qdisc,cc}` gauge (1=running, 0=idle)
   - Emission MUST include `# HELP`/`# TYPE`, MUST set `Content-Type: text/plain; version=0.0.4`, MUST NOT duplicate (name,labels) sets (Prometheus primary §3.3). Retention/alerting config: **unverified — primary source not found**, so no interval/days pinned in MUST; placeholder SHOULD in §6.2.
4. **Auth boundary for shaping plane.** `POST /api/shape` + `cgo shape --restore` MUST require the same auth as (1); Nginx MUST NOT proxy `/api/shape` to loopback without auth. Benches (`10.99.0.2:9090` mgmt path `kit/run-dualns.sh:57-63`) MUST move behind Nginx TLS; raw `--tls=false` only on loopback in CI.
5. **Deployment package: OCI + compose/systemd.** OCI image (multi-arch per `.goreleaser.yaml:12-17` targets), digest-pinned, plus compose file wiring dashboard + Nginx + (eventually) Prometheus. Systemd unit preserved (`docs/deploy.md:89-97`) as fallback. `cgo verify` + `cgo doctor` + `cgo shape --restore` as post-deploy gates. Rollback = previous digest + `cgo shape --restore`. Ansible: SHOULD, not MUST — compose+systemd covers DGM single-site first.
6. **Provenance contract (extend, don't replace).** Keep frozen CSV schema + `manifest.json` + `hash8` + dual audit seals + `invocation` (all §3.7). Add: OCI image digest + `--version` stamp to `GET /api/health` (already present) and to `manifest.json` at freeze; add Prometheus `cgo_build_info{version}` gauge. No relabeling of frozen 159 runs.
7. **FQ-CoDel parameter pinning.** Declare `target/interval/limit/quantum/flows` explicitly per RFC 8290 §5.2 in bench + remediation config (don't inherit kernel defaults silently). `limit` MUST be set (MUST-enforced) — defaults in `pkg/qdisc` and `kit/cgo-bench-dualns.sh` today are implicit.

### 6.2 SHOULD (desirable, follow MUST quickly)

- **Blackbox Exporter** for probe-free link checks (ping/HTTP/TLS) with `site` labels; **Node Exporter** for host metrics (CPU/memory/disk) + `node_cpu_seconds_total` and `node_memory_MemTotal_bytes` (exact exporter doc pages **unverified — primary source not found**; deploy the official exporters, don't write custom ones).
- **Grafana dashboards:** (a) Operator wall (live SSE-equivalent panels), (b) Audit comparison (valid-only medians like `pkg/results/results.go:175-186`), (c) SLO/CAP SLA (220 ms objective, breach flag — labelled "operator objective" per `pkg/metrics/cost.go:75-79`), (d) Provenance (run counts, hash8, seal status). Dashboards MUST be versioned JSON in repo.
- **Alerting:** Prometheus recording + Alertmanager rules for (a) `cgo_cap_sla_breach == 1` sustained, (b) `rate(cgo_audit_load_measured_total) == 0` (probes dead), (c) `cgo_iface_valid == 0` sustained. Thresholds/durations: to be pinned after fetch — **unverified this session**.
- **cgroups:** `cpu.max`/`memory.max` on dashboard container + bench namespace processes; export `cpu.stat` throttling counters (kernel primary §3.4).
- **Runbooks:** extend `docs/runbooks/` beyond host-acceptance with: auth-rotation, Nginx revoke, rollback (`cgo shape --restore` + digest pin), Prometheus target-down, bench offload fail (`cgo-bench-dualns.sh` check), corpus seal verification (`python book/tools/audit-archive.py`).

### 6.3 DEFERRED (explicitly not now)

- **WIS 2.0 / CAP publication** (alert lifecycle, topic federation, CAP XML signing): DEFERRED until WMO primary docs fetched (§8). CGO MUST NOT claim WIS compliance today; 220 ms stays DGM objective.
- **Mosquitto multi-site fan-out** (topics/QoS/TLS/auth specifics): DEFERRED until authN/Z + Nginx land (§7 sequence). No second unauthenticated plane.
- **CHR/QEMU CI validation:** DEFERRED until CHR image license/procurement + human approval (§8). Fixture replay (`kit/chr/export-*.txt`, `docs/VERIFY.md:57`) stays the gate.
- **eBPF ground truth, BBRv3, `.dist` netem:** DEFERRED — BLOCKED rows stand (`docs/VERIFY.md:29,33,34`). No enterprise claim depends on them.
- **SQLite/Postgres** for runs: DEFERRED. CSV+manifest+seal is the provenance contract; DB migration would break it. Keep file store.

### 6.4 SLOs (honest proposal — threshold numbers beyond primary sources framed as operator choices)

- **Service SLO (dashboard):** 99% monthly availability of `GET /api/health` 200 on mgmt interface; p95 `GET /api/state` < 200 ms on loopback. **Both thresholds are operator proposals, not primary-sourced** — to be baselined after first enterprise deploy, then moved to VERIFIED only by measurement.
- **Measurement SLO:** every sealed audit MUST have `invocation` + dual seals + `load_attempted`/`load_measured` pair; every campaign cell MUST be `valid/degraded/invalid` under G0–G7 (`docs/architecture.md:60-74`). Exported as `cgo_audit_seal_ok` gauge. This is the SLO that matters for the thesis: evidence completeness, not uptime theatre.
- **CAP SLA:** retain 220 ms / 95% as **operator objective** (`pkg/metrics/cost.go:80-83`), surfaced as `cgo_cap_deadline_ok_pct` + breach, never labelled WMO-mandated.

---

## 7. Slices in dependency order

Ordered by dependency: each slice unblocks the next; no slice skips auth.

1. **S1 — Ingress hardening (Nginx + TLS + rate limit).** Nginx in front of 9090 everywhere (compose + systemd sample). `limit_req_zone` for all POSTs. Closes rate-limit + exposure half of §4 FAIL. Smallest complete change: config only, no Go. *Acceptance:* `limit_req_status 503` under synthetic burst; TLS handshake verifies; `docs/deploy.md:159-177` updated to describe the new boundary.
2. **S2 — AuthN/Z on mutation APIs.** Session/MTLS bearer on handlers in `pkg/api/server.go` (reuse `Deps` pattern). CSRF tokens on cookie flow. Unit tests per handler (401 without creds, 403 CSRF). *Acceptance:* unauthenticated `POST /api/shape` → 401; `GET /api/health` still public on loopback only.
3. **S3 — Prometheus `/metrics` + Blackbox/Node exporters.** Go handler emitting §6.1 contract with HELP/TYPE; compose wiring for exporters; dashboards as versioned JSON. *Acceptance:* `promtool check` (or scrape) passes; Grafana panels render valid-only medians; `iface_valid:false` renders as gap, not zero.
4. **S4 — Container packaging + cgroups.** OCI image + digest pin + `cpu.max`/`memory.max` through compose/systemd. *Acceptance:* `docker compose up` → health green; `cpu.stat nr_throttled` visible in `/metrics`; rollback by digest works.
5. **S5 — SLO/alert runbooks.** Recording/alerting rules for breach/dead-probe/invalid-iface; runbooks in `docs/runbooks/`. *Acceptance:* alert fires on synthetic breach fixture; runbook drill passes.
6. **S6 — RouterOS CHR validation (human-gated).** QEMU + CHR image → real boot → fixture diff becomes live acceptance. Moves `docs/VERIFY.md:57` BLOCKED→PENDING→VERIFIED only on booted evidence. Depends on S1–S4 (needs authenticated bench access).
7. **S7 — Multi-site aggregation (Mosquitto).** Only after S1–S5. Fetch Mosquitto primary docs first; define topic tree (`cgo/{site}/{metric}`), QoS, retained policy, TLS. *Acceptance:* two-site aggregate doc with per-site `site` label intact.
8. **S8 — WIS 2.0 / CAP alignment.** Only after WMO primary docs fetched. Gate on §8 procurement. Until then, DEFERRED.

Cadence constraint from `AGENTS.md`: at most three active slices. Suggested first three: S1, S2, S3.

---

## 8. Open decisions requiring human approval (live runs, deps, CHR, procurement)

1. **Live bench runs.** Current BLOCKEDs require Linux 6.8 host + human-approved live runs (`docs/VERIFY.md:28,31,56`). No enterprise claim (G0–G7 live, IFB acceptance, `.dist`, eBPF) can close without that approval. Human must grant bench access + run window.
2. **CHR procurement + QEMU.** RouterOS validation needs CHR image (license, version pin) + QEMU availability on the bench host. Plinks: `docs/VERIFY.md:57` (QEMU not installed, no CHR cached), `pkg/remediate/remediate.go:268-384` strings unverified vs MikroTik docs (404). Human must approve download, version, and any license cost.
3. **Dependencies (new).** Proposal adds Mosquitto, Nginx, Prometheus, Grafana, Blackbox/Node Exporters, OCI runtime if not present. Each is a new `overhead`: Nginx + Prometheus are MUST (§6.1) because they close FAIL/GAP; Mosquitto is DEFERRED; Ansible is SHOULD. Per `AGENTS.md` "Stop and ask" (new dependency), human must approve each addition before scaffolding.
4. **WIS 2.0 / CAP scope.** Any WIS publication, CAP XML generation, or "WIS-compliant" wording requires WMO primary docs (`warnings.wmo.int`, WIS 2.0 guide, CAP 1.2 spec) — all **unverified this session** (404/fetch fail). Human (DGM) must decide whether WIS integration is in thesis scope at all; 220 ms stays operator objective regardless.
5. **Case-study grant.** Read-only service map of the forecasting server needs the operator's written grant (`docs/VERIFY.md:60`, contract `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md`). Human + DGM operator sign-off required; no capture/scan/write.
6. **Port / network exposure.** `docs/deploy.md:140-143` (open 9090) is lab-only. Enterprise exposure (which sites, which mgmt subnets, TLS cert issuance) needs DSI decision. Default stays mgmt-only behind Nginx TLS.
7. **Tariff refresh.** Enterprise scenarios (`pkg/metrics/cost.go:19-62`) cite yas.mg 31/08/2026 + operator quote. Re-use requires fresh sourcing. Human must confirm quote still valid before any number leaves the repo (per `AGENTS.md` verifier rule: numbers leaving the repo need a separate verification pass).
8. **Microsecond / RAM / scope claims.** If the thesis or slides state microsecond probing, RAM footprints, or 3-file scope, those sentences need deletion or primary measurement. Current verdict (§5): all three rejected as stated. Human must approve removal from prose before defense.

---

## Methodology note

This document was produced from direct file reads (repo paths + line numbers) and fetches of first-party URLs (RFC editor, Prometheus, Mosquitto index, Nginx, kernel.org). Fetches that returned 404 are recorded as unverified, not backfilled from memory. No secondary write-ups, marketing copy, or unsourced numbers were allowed into MUST. Anything that cannot be cited is labelled `unverified — primary source not found` in place.

## Repository file references (for verifier spot-checks)

- `README.md:36-37,45-46,159` · `docs/architecture.md:1-4,23-58,60-83,90-103` · `docs/api.md:1-116` · `docs/deploy.md:55-177` · `docs/VERIFY.md:24-60,72-115`
- `go.mod:1-10` · `.goreleaser.yaml:1-64` · `pkg/audit/audit.go:15-38,66-68,124-127,269-278,488-565,639-650` · `pkg/remediate/remediate.go:66-101,148-260,262-384` · `pkg/results/results.go:12-47,65-236` · `pkg/metrics/cost.go:19-83,97-112,170-188` · `pkg/qdisc/qdisc.go:66-143` · `pkg/campagne/matrix.go:103-258` · `pkg/api/server.go:218-320,460-472,611-742`
- `kit/cgo-bench-dualns.sh:42-55,115-256,258-341` · `kit/run-dualns.sh:46-63` · `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:28-29,40-67,69-112`
- `web/frontend/src/App.tsx:1-178` · `web/frontend/src/views/{DashboardView,AuditView,BenchView,EvidenceView}.tsx`
