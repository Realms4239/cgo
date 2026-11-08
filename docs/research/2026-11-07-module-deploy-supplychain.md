# deploy / supply-chain as standalone tools — primary-source research

*Date: 2026-11-07 · Scope: `deploy/` + `kit/*.sh` + `tools/verify.py` + release packaging, read as if released tomorrow. Research only — no code changed.*

Primary sources fetched this pass: (R1) Nginx `ngx_http_limit_req_module` (nginx.org/en/docs/http/ngx_http_limit_req_module.html); (R2) Prometheus exposition formats (prometheus.io/docs/instrumenting/exposition_formats/); (R3) OCI Image Format spec (opencontainers.org → GitHub `opencontainers/image-spec`; Distribution Spec companion); (R4) semver.org 2.0.0; (R5) GoReleaser SBOM docs (goreleaser.com/customization/sbom/); (R6) SLSA v1.0 levels + v1.2 whats-new (slsa.dev); (R7) CycloneDX SBOM capability (cyclonedx.org/capabilities/sbom/). Unverifiable claims are marked **`unverified — primary source not found`**.

## 1. What deploy/ actually is today

Four files in `deploy/` — no CI, no registry publish step, no signing, no SBOM (verified: `glob deploy/**/*` returns only Dockerfile, docker-compose.yml, prometheus/rules.yml, nginx/cgo.conf; no `.goreleaser.yaml`, no `.github/workflows`, no cosign/sigstore config anywhere in-tree):

| File | Responsibility | Exported config surface (file:line) |
|---|---|---|
| `deploy/Dockerfile` | Minimal image: binary into scratch | `FROM scratch` (:1), `COPY bin/cgo /cgo` (:2), `EXPOSE 9090` (:3), `ENTRYPOINT ["/cgo","--serve"]` (:4). Zero build stage, no `USER`, no `HEALTHCHECK`, no `LABEL`. |
| `deploy/docker-compose.yml` | Local run of one digest | image `ghcr.io/realms4239/cgo:latest` (:3), port `127.0.0.1:9090:9090` (:5), `cpus 1.0 / memory 512M` (:9-10), `CGO_API_TOKEN=${CGO_API_TOKEN:?set}` (:12), `restart: unless-stopped` (:13). |
| `deploy/nginx/cgo.conf` | Edge limiter + TLS terminator | `limit_req_zone … cgo_reads:10m 30r/s` (:1), `cgo_mutations:10m 1r/s` (:2), `listen 9090 ssl` (:5), `limit_req_status 503` (:9), per-location `limit_req … burst=10/3 nodelay` (:12,20,25,31), `/metrics` allow-127.0.0.1-deny-all (:17-19), all `proxy_pass http://127.0.0.1:9090` (:13,21,26,32). |
| `deploy/prometheus/rules.yml` | Recording + alert rules | `cgo:slo:deadline_ok_pct_5m` (:4-5), `CGO_deadline_breach_sustained` (:6-12), `CGO_probe_dead` (:13-19), `CGO_iface_invalid` (:20-26). |
| `docs/deploy.md` | Operator install narrative | Windows-observe / Linux-full matrix (:18-28), `kit/install.sh` (:60), `setcap cap_net_admin+ep` (:83), `cgo service install` (:92), `cgo kit …` (:112-118), security line "not multi-tenant hardened" (:159-177), S1 Nginx boundary (:179-180), S4 OCI+rollback (:182-183). |
| `tools/verify.py` | 6-gate verifier (frozen/provenance/tests/exit/hygiene/links) | `GATES` registry (:455-462), `verdict()` 0/1/2 contract (:36,255-261), `SIZE_CEILING` (:13), hygiene selectors (:307-333), `--fast/--staged/--only/--selftest` CLI (:479-484). Zero network/shell by design (:2). |

What it does **not** do: build the image in CI, sign anything, publish an OCI registry artifact with provenance, generate an SBOM, version-tag immutably, or enforce digest pinning beyond a prose claim (`docs/deploy.md:183` says "digest-pinned image" but `docker-compose.yml:3` uses `:latest` — direct contradiction, see §5).

## 2. Per-artifact analysis against primary sources

### 2.1 `deploy/nginx/cgo.conf` — rate limiter

Per R1, `limit_req_zone key zone=name:size rate=rate` defines a shared-memory leaky bucket (R1, `limit_req_zone` directive); `limit_req zone=name burst=n nodelay` delays or rejects once burst is exceeded, and rejected requests return `limit_req_status` code, default 503 (R1, `limit_req` + `limit_req_status`). The config is conformant in form: two zones at 10 m (:1-2), per-IP key `$binary_remote_addr` — R1 notes `$binary_remote_addr` (4 bytes IPv4 / 16 bytes IPv6) is preferred over `$remote_addr` and 10 MB ≈ 8k 64-bit states (R1, `limit_req_zone` §"Usage example"), so 10 m per zone is generous for one edge host. `limit_req_status 503` (:9) matches R1 default (R1, `limit_req_status`), and `docs/deploy.md:180`'s "exceeding 1r/s burst=3 receives 503" is exactly R1 behaviour for `nodelay` burst exhaustion.

Gaps, with R1 in hand:
- **TLS is fake as committed.** `listen 9090 ssl;` (:5) with no `ssl_certificate`/`ssl_certificate_key` will fail `nginx -t`; the comment (:8) defers to `docs/deploy.md S1`, which only restates "TLS termination" (docs/deploy.md:180). R1 gives no TLS guidance (TLS is core nginx, not the limit module) — operator must supply cert paths outside this file. As a standalone tool the conf cannot be pasted into a server block as-is.
- **`/metrics` ACL (:17-19) is belt-and-braces but inconsistent with the rest** — every other location proxies to `127.0.0.1:9090` identically, so the only layer distinguishing metrics is this ACL; defense is fine, but `allow 127.0.0.1/deny all` with no `allow` for the monitoring network means off-host Prometheus cannot scrape. Intended (loopback-only) — must be documented as such.
- **No `limit_req_dry_run` for tuning** — R1 defines it (1.17.1+) as count-without-limiting (R1, `limit_req_dry_run`); a released tool should ship with a commented dry-run example for capacity tuning. Absent.
- **`proxy_pass` hardcodes `127.0.0.1:9090`, colliding with the `listen 9090` port** (:5 vs :13) — upstream and listener on the same port implies nginx and cgo share the loopback port namespace only if cgo binds elsewhere; in the compose file cgo is the sole 9090 binder (docker-compose.yml:5). The intended topology (nginx fronting cgo on the same host) needs nginx on a different port or cgo on a unix socket / other loopback port. **unverified — primary source not found** for which port cgo actually binds behind nginx (no `proxy_pass` variable, no second port in compose or docs).

### 2.2 `deploy/docker-compose.yml` — runner

Per R3, an OCI image is self-describing via its config (runtime args, env, ports) and distributed through the Distribution Spec API (R3, image-spec README "Distributing an OCI Image"); a Compose file is a local convenience on top. Surface: image tag `:latest` (:3), bind `127.0.0.1:9090:9090` (:5), CPU/memory limits (:9-10), token via `:?set` guard (:12), `unless-stopped` (:13).

Gaps, with R3/R6/R7 in hand:
- **`:latest` is mutable; R4 (semver) forbids modifying a released version's contents** (semver.org §3: released version contents MUST NOT be modified, must be released as a new version) (R4). A `:latest` tag that silently moves violates that invariant operationally; the spec-correct pin is `repo@sha256:<digest>` (R3, Distribution Spec content-addressing) — and `docs/deploy.md:183` already claims digest-pinning that the file does not do.
- **No `cap_add: NET_ADMIN`, no `network_mode`/`cap_dac`**, yet Linux mode requires `CAP_NET_ADMIN` for `tc`/netem shaping (docs/deploy.md:24,79-87). The compose service as written can only run observation mode; a returned image that cannot push packets fails its own §2.1 capability matrix.
- **No `read_only`, `user`, `healthcheck`, or TLS volume** — the Dockerfile sets no `USER` either, so the container runs as root by default. This is a supply-chain hygiene point, not an OCI violation; OCI itself does not mandate non-root (R3) — claim is best practice, marked **unverified — primary source not found** at the OCI level.
- **`CGO_API_TOKEN` is the only secret pattern good practice in the tree** (`:12` with `:?set` fail-fast), matching the runbook rotation procedure (docs/runbooks/smoke-alerts.md:3-6). Keep.

### 2.3 `deploy/Dockerfile` — image

`FROM scratch` + one `COPY` + `EXPOSE 9090` + `ENTRYPOINT` (:1-4). Per R3, the OCI Image Format's job is to carry launch info (cmd/args/env) on the target platform (R3, image-spec README "Running an OCI Image"); a scratch image with a static Go binary satisfies the letter of that — no runtime, no shell, minimal attack surface.

Gaps:
- **No `USER`, no `read_only` hints, no OCI `org.opencontainers.image.*` labels** (version, revision, created, source) — R3 defines these as the standard annotation surface; their absence means `docker inspect` cannot provenance the image. Labels convention source: **unverified — primary source not found** in this pass beyond the image-spec repo layout (annotations.md exists at R3 but was not read line-by-line).
- **No `HEALTHCHECK`** — OCI/Docker do not require it (best practice, **unverified — primary source not found** at spec level); the runbook instead polls `/api/health` externally (docs/runbooks/smoke-alerts.md:6). Consistent, but the health path duplicated in three places (runbook, deploy.md:16, compose-missing) invites drift.
- **Architecture is implicit** — the binary is `bin/cgo` whatever the build host produced; no `GOOS/GOARCH` pin or multi-arch manifest. R3's image index concept exists for multi-platform (R3, image-index.md in repo layout) — not used here; a Windows-built `bin/cgo.exe` copied in would produce a broken image silently.

### 2.4 `deploy/prometheus/rules.yml` — alerting

Per R2, the text exposition contract requires: `Content-Type: text/plain; version=0.0.4` (R2, "Basic info" table), `# HELP`/`# TYPE` lines with TYPE before first sample and exactly one each per metric (R2, "Comments, help text, and TYPE"), `\`/`"`/LF escaping of label values (R2, `escaped_string` / `label_value`), `_sum`/`_count`/bucket conventions for summaries/histograms, and `NaN`/`+Inf`/`-Inf` as valid float values (R2).

Findings against the only exposition builder, `pkg/api/metrics.go:17-41` (full analysis in the parallel `docs/research/2026-11-07-module-metrics.md`):
- Builder sets `text/plain; version=0.0.4` (metrics.go:45) — conformant with R2.
- No label-value escaping (metrics.go:34 `fmt.Sprintf(`%s="%s"`)` raw) — violates R2 `escaped_string` rule (must escape `\`, `"`, LF as `\\`, `\"`, `\n`). Current labels are `site=unknown` (metrics.go:48-52) so harmless today; a future `site` with a quote breaks exposition parsing.
- TYPE inferred from `_total` suffix (metrics.go:27-29) — R2 requires explicit TYPE, not suffix inference; `cgo_cap_deadline_ok_pct` (rules.yml:5) would be mis-typed `gauge` correctly by luck, but any `_total` that is actually a gauge is mis-declared.
- Metrics referenced by rules (`cgo_cap_deadline_ok_pct`, `cgo_cap_sla_breach`) are **never emitted** by `handleMetrics` (metrics.go:46-53 emits only `*_total` progress counters + zero flags) — the recording rule `cgo:slo:deadline_ok_pct_5m` (rules.yml:4-5) and `CGO_deadline_breach_sustained` (rules.yml:6-12) therefore never fire from the real endpoint; only `cgo_audit_load_measured_total` (rules.yml:14, emitted metrics.go:49) and `cgo_iface_valid` (rules.yml:21, emitted metrics.go:51) are live. Rules wired to dead metrics = silent-green dashboards unless `CGO_probe_dead` (rules.yml:13-19) catches total absence.
- `rate(cgo_cap_deadline_ok_pct[5m])` (rules.yml:5) applies `rate()` to a percentage gauge — R2 defines `rate` semantics nowhere (it's PromQL, not exposition), so this is **unverified — primary source not found** for whether the ratio is a counter-meaningful range; as recorded, a gauge percentage fed to `rate()` is semantically wrong (should be `avg_over_time` or the percentage emitted as a summary). At minimum the metric must exist.
- Rules lack `severity` consistency (`critical` vs `warning`, rules.yml:10,17,24 — fine) but have no `runbook_url`/drill annotations; harmless.

### 2.5 `docs/deploy.md` + `kit/*.sh` + `tools/verify.py`

- `docs/deploy.md` is the strongest artifact: explicit capability matrix (:18-28), the "n/a ≠ failure" rule (:30-34), preflight (`cgo doctor`, :68-75), shaping privilege model (`setcap` preferred, :79-87), systemd install (:89-97), shape-restore guarantee (:99-104), bench VM flow via `cgo kit` (:108-118), edge mini-PC recipe (:130-147), and a brutal honesty security line "not multi-tenant hardened; never expose 9090 on the internet" (:159-177). As a standalone deploy doc it needs: what changed in each version (no changelog pointer), where the release artifact comes from (URL + checksum), and rollback beyond shape-restore (see §5).
- `kit/*.sh`: 15 scripts. Production-critical: `install.sh` (apt bootstrap, idempotent, never switches default CC — kit/install.sh:57-69, honest warn-only network), `cgo-bench-dualns.sh` (decoupled netem/CAKE namespaces; the design comment kit/cgo-bench-dualns.sh:1-35 explains *why* — invisible FIFO attribution — which is the single best engineering comment in the deploy surface), `run-dualns.sh` (up/serve/down, sink inside `ns_server`, kit/run-dualns.sh:46-49). `engine.sh` is a documented compat shim reading `bin/cgo` first (kit/engine.sh:1-4). `config.sh` is a YAML-subset parser with zero hardcoded identities (kit/config.sh:1-17). `host-readiness.sh` is read-only diagnostic with fail-1 contract. The `test-*.sh` scripts are stubs (`test-kit.sh` 489 B, `test-hypervisor.sh` 455 B) — fine as smoke tests, not as gates.
- `tools/verify.py`: exactly six gates (verify.py:455-462), stdlib-only, exit contract 0/1/2 (verify.py:36), every gate has a fault proof (`prove_*` + `--selftest` verify.py:510-516). This is the supply-chain's best friend: it already pins frozen corpus hashes, secret scanning, and link integrity. It is **infra, not an artifact** — a standalone deploy must publish it as a verifier binary/script separately from the product (see §7).

## 3. Primary-source requirements table

| Req | Source | What it demands of a released cgo deploy |
|---|---|---|
| Leaky bucket `limit_req_zone`/`limit_req`, 503 on reject, `nodelay` | R1 lines "limit_req", "limit_req_zone", "limit_req_status" (fetched verbatim) | zones+status already conformant; TLS certs and port topology must be supplied outside `cgo.conf`; ship commented `limit_req_dry_run` for tuning (R1) |
| Text exposition: `text/plain; version=0.0.4`, HELP/TYPE ordering, `\`/`"`/LF escaping, `_sum/_count`, unique metric+labels | R2 "Basic info", "Details > Comments/HELP/TYPE", "Grouping and sorting", "Histograms and summaries" | `renderMetrics` must escape label values, emit explicit TYPE per metric (not suffix-inferred), emit the metrics the rules reference or delete the dead rules; add `CGO_iface_valid` parity fixtures |
| Image contains launch info; distinguish image vs runtime vs distribution | R3 README ("Running an OCI Image", "Distributing an OCI Image", Distribution Spec API) | keep scratch+entrypoint shape; add `org.opencontainers.image.*` labels + created/revision; publish via a registry speaking the Distribution Spec, pull by digest |
| Version immutability `X.Y.Z`, MAJOR/MINOR/PATCH bump rules, pre-1.0 instability, "1.0.0 defines the public API" | R4 §§2-8; §1 "MUST declare a public API"; FAQ "v1.2.3 is not a SemVer" | declare public API (CLI flags, REST, /metrics names); tag `vX.Y.Z` ≠ version `X.Y.Z`; never move a released tag; current `1.3.5` (internal/kit/kit.go:52) already claims a stable API — declare it |
| SBOM per archive via pluggable cmd (default `syft`), formats via `--output` (spdx-json / cyclonedx-json) | R5 "Usage" (documents/artifacts/args templates; limitation: container images not catalogable by the SBOM step) | add `.goreleaser.yaml` `sboms:` block targeting archives+binaries; note GoReleaser's image non-catalog limitation — SBOM for OCI must come from `syft <image>` / cosign-attest separately, not this pipe |
| Provenance per artifact; L1 exists, L2 signed-on-hosted-builder, L3 hardened-builder; Source track added v1.2 | R6 "Security levels" + "What's new" v1.2 (Source track, VSA guidance) | L1: publish provenance (what built what, from which commit) with every release; L2: sign via hosted CI; state achieved level per artifact in release notes |
| SBOM in machine-readable form: component inventory + dependency hierarchy; vulnerability/compliance integration | R7 "Capabilities > SBOM" ("Inventory software components and services and the dependency relationships"; "Supports end-to-end transparency") | publish CycloneDX (or SPDX) per release asset + OCI image; the hierarchical dependency graph is the requirement, not a bare package list |

## 4. Dependencies

- **Nginx** ≥ 1.3.15 (first `limit_req_status` version per R1) — realistically ≥ 1.27 for `$binary_remote_addr` maturity; **unverified — primary source not found** for a current minimum version beyond R1's inline version stamps.
- **Prometheus-compatible server + Alerting rules loader** for rules.yml; needs at least the version that understands recording rules + `for:` — **unverified — primary source not found** (not fetched in this pass).
- **Docker / OCI runtime** supporting Compose v2 `deploy.resources.limits` (docker-compose.yml:6-10) and `FROM scratch` + digest pull (R3).
- **Go 1.25 toolchain** (`go.mod:3`, `kit/config.sh` `CFG_GO_MIN_VERSION=1.25`) for building `bin/cgo`.
- **Syft** (or equivalent SBOM tool) if GoReleaser default SBOM pipeline used (R5 default `cmd: syft`).
- **cosign / sigstore** for image signing — not yet a dep; proposed (§7). SLSA L2 requires a hosted builder that signs provenance (R6) — cosign keyless signing is one route; key storage choice is **unverified — primary source not found** for this org (no policy doc read).
- `kit/*.sh` runtime deps: bash, `iproute2` (`tc`), `curl`, `bc`, `tcp_bbr` kernel module (kit/install.sh:22-55) — documented, idempotent, warn-don't-fail where touching the host.

## 5. Gap analysis (ranked)

1. **Image identity contradiction.** `docs/deploy.md:183` claims digest-pinning + rollback by prior digest; `docker-compose.yml:3` ships `:latest` (mutable). R4 §3 forbids moving a released version. Fix is a versioned tag + digest in compose and a rollback note naming *which* prior release.
2. **Compose cannot shape.** No `NET_ADMIN` capability (docker-compose.yml) vs Linux-full shaping requirement (docs/deploy.md:79-87). Container as shipped = observation mode only, contradicting the Linux-full promise on the same page. One line fix (`cap_add: [NET_ADMIN]`) plus a README note — or restrict compose to observe mode explicitly.
3. **TLS conf is undeployable as pasted.** `listen 9090 ssl` with no cert directives (cgo.conf:5). nginx `-t` fails. Either ship `ssl_certificate` placeholders pointing at `/etc/nginx/certs/` with a docs step, or demote the file to "intent documentation" and template the cert paths.
4. **Port collision.** `listen 9090` and `proxy_pass 127.0.0.1:9090` (cgo.conf:5,13) — nginx and cgo cannot both own 9090 on loopback. Needs upstream port divergence documented (e.g. cgo on 9091 / unix socket).
5. **Dead alert rules.** `cgo_cap_sla_breach` / `cgo_cap_deadline_ok_pct` (rules.yml:5,7) are not emitted by `pkg/api/metrics.go:46-53`; dashboards stay green. Either emit them (measured, never invented — the house rule) or delete the rules and record the gap.
6. **Exposition non-conformance (labelling/typing)** — label-value escaping missing, TYPE suffix-inferred (§2.4). Any `site` label with a quote/backslash breaks R2 parsing. Small fix in one function.
7. **No SBOM, no signing, no provenance, no air-gap bundle.** Everything §7 proposes is absent. That is the whole supply-chain gap; the repo is currently a from-source build + manual copy deployment (docs/deploy.md:42-61, 108-118), which is honest but not distributable at scale.
8. **Version sprawl.** `KitVersion` (internal/kit/kit.go:52) is source of truth, `cmd/cgo/main.go:41` defaults to it, frontend pins its own string (per CLI-frontend research), and there is no `.goreleaser.yaml` to derive `-ldflags -X` consistently. `cgo version` vs tag drift already warned about in code (main.go:36-40).
9. **Bench scripts are privileged and unguarded.** `run-dualns.sh` requires root/sudo-NOPASSWD and mutates netns; it is correctly documented as lab-only (docs/deploy.md:172-174) but ships in the same repo directory as install scripts. Separate or clearly gate.
10. **rules.yml has no deploy target.** Nothing loads it into a Prometheus; the file is inert. A standalone tool must name the target (sidecar volume mount, remote-write, or `promtool` validation) and wire it in compose.

## 6. Target packaging

Three artifact families, all built by one GoReleaser config (absent today, proposed).

### 6.1 Signed binaries + archives (Windows observe + Linux full)

- `.goreleaser.yaml` at repo root (GoReleaser reads this filename by canonical convention — R5 example config is exactly `.goreleaser.yaml`): `builds` for `cmd/cgo` (linux/amd64, windows/amd64) with `-ldflags "-X main.version={{.Version}}"` matching the `version` var (cmd/cgo/main.go:41) so `cgo version` equals the tag (R4: name the version, don't confuse `vX.Y.Z` tag with `X.Y.Z` version), `archives` naming `cgo-<os>-<arch>.tar.gz` / `.zip`, `checksum` (sha256), `sboms:` per R5 (documents `{{ .ArtifactName }}.sbom.json`, `args: syft … --output cyclonedx-json` or spdx-json), `signs:` for archives/checksums, GitHub release publish with the SBOM + checksum files as first-class assets.
- Version single-sourcing: generate `internal/kit/KitVersion`, `VERSION`, and frontend `package.json:version` from the git tag in CI rather than three manual bumps (drift already warned about in code).

### 6.2 Signed OCI images (the edge story)

- Multi-arch image index (R3 image-index concept) from the scratch Dockerfile, with `org.opencontainers.image.title/version/revision/source/created` labels set at build time (revision = git SHA; unverified detail: exact label key namespace should be opencontainers.org annotations.md, fetched only at repo level in this pass).
- Publish to `ghcr.io/realms4239/cgo` **by digest**, and change compose to `image: ghcr.io/realms4239/cgo@sha256:…` (or at minimum `:<semver>`), so `docs/deploy.md:183`'s rollback sentence becomes literally true.
- Sign with cosign (keyless OIDC or a rotated key) and attach provenance as an OCI referrer / attestation so SLSA consumers can verify (R6 L2); the GoReleaser-docker SBOM limitation (R5 "Limitations": container images not catalogable by the SBOM step) means OCI SBOM must be produced off-image: `syft ghcr.io/…/cgo@sha256:… --output cyclonedx-json` and pushed as an attestation (referrer) — or simply documented as a release asset alongside the image if referrer support is unavailable.
- Advertise the SLSA level per release: L1 (provenance exists: which commit, which builder, which commit of bench scripts) with the release; L2 (signed provenance from hosted CI) once a workflow builds+signs; L3 (hardened builder) is a stated non-goal until the platform is hardened — claiming L3 without the platform is exactly what R6 warns provenance may be bypassed at L1.

### 6.3 Air-gap bundle (one download, `docker load`)

For the mini-PC edge recipe (docs/deploy.md:130-147) and bench VMs with no egress:
- Bundle = `docker save` OCI tarball (image @ digest) + the compose file pinned to that digest + `cgo.conf` + `rules.yml` + `bin/cgo` archive + release SBOM + `SHA256SUMS`. Loader: `docker load`, `docker compose up`, `nginx -t && systemctl reload nginx`, `promtool check rules rules.yml` if promtool is present.
- If a registry is unavailable at the edge, the SBOM and provenance must be in the bundle as files, not behind a URL — R7's machine-readable SBOM requirement is useless if it cannot be read offline.
- Verify-before-run: each file's sha256 against `SHA256SUMS` in the bundle plus the signed checksums from the release (GoReleaser `checksum` + `signs` output). That is the verifier story told for Terraform/Ansible; for the container bundle the checksum verification is a loader script step, not manual.

## 7. Slices (proposed, each with acceptance)

- **S1 — Make the deploy docs true** (no signature infra). (a) compose → `image: …@sha256:<digest>` + `cap_add: [NET_ADMIN]`, or split into `compose.observe.yml` vs `compose.full.yml`; (b) cgo.conf → template cert paths + non-colliding upstream port, comment `limit_req_dry_run`; (c) rollback section names an actual prior digest. Acceptance: `docker compose config` succeeds from the repo, `nginx -t` passes with a test cert, and the rollback command in docs matches compose. Priority: the two lies (digest, shaping).
- **S2 — Exposition conformance.** Escape label values, explicit TYPE per metric, emit `cgo_cap_deadline_ok_pct` + `cgo_cap_sla_breach` measured or remove rules.yml:4-12 and record the gap; fix `rate()`-on-percentage (rules.yml:5). Acceptance: exposition validates (byte-identical round trip through a real parsers' text parse — which one is **unverified — primary source not found** for an in-repo validator; use Prometheus' parser via a vendored lib or a golden-file test), `rules.yml` referenced metrics all present in `GET /metrics`.
- **S3 — `.goreleaser.yaml` (binaries+archives+SBOM).** GoReleaser with checksum + SBOM (R5, default `syft` path) for archives and binaries; GoReleaser's image non-catalog limitation means OCI SBOM goes in S4. Acceptance: cut a pre-release, `goreleaser check` passes, SBOM JSON parses as CycloneDX/SPDX and lists the direct deps from `go.mod:5-10`, checksums verify by hand.
- **S4 — OCI sign + provenance (SLSA L1→L2).** Digest-pinned publish, `org.opencontainers.image.*` labels, cosign sign, SLSA provenance attestation behind hosted CI. Acceptance: `cosign verify <image>@<digest>` + provenance check green in CI; compose pulls the digested image.
- **S5 — Air-gap bundle script.** `make bundle` produces the §6.3 tarball + `SHA256SUMS`; tested by `docker load` on a clean VM/CI container. Acceptance: bundle files verify with `sha256sum -c`, compose up works offline.
- **S6 — Release engineering.** Version generate-from-tag, CHANGELOG/attestations publish, rules.yml wiring (compose service with volume mount + `promtool check rules` in CI — promtool availability **unverified — primary source not found** in this environment). Acceptance: `cgo version` == tag == release assets; `get/metrics` and rules consistent on a fresh deploy.

Out of scope for the deploy module: the bench itself (`kit/cgo-bench-dualns.sh` measurement code), the in-binary verifier (`tools/verify.py` ships alongside and gates releases but is not the release artifact), and the frontend bundle (separate module research).

## 8. Anti-toy criteria (what would make this unacceptable as a standalone tool)

A deploy/supply-chain submission fails review if any of these holds:
1. **Mutable version identity.** `:latest`, branch-name tags, or a `.goreleaser.yaml` absent so releases are hand-copied binaries (current state; R4 §3). A released artifact's digest must be immutable once published.
2. **Unverifiable payload.** No checksum or SBOM downloadable beside the binary/image (R7). "Trust the zip" fails.
3. **PDF-ish identity.** The image cannot be inspected for provenance — no image labels, no `cgo version` == tag, no revision embedded. R6 L1 minimum is provenance *exists*; without it the tool is indistinguishable from a random binary.
4. **Dead observability.** Alert rules referencing metrics the endpoint never emits (current rules.yml:4-12), or TYPE/escaping violations that let a parser silently misread (R2). Monitoring that can't fire is not monitoring.
5. **Capability lies.** Compose/docs promise Linux-full shaping the container cannot exercise (no NET_ADMIN), or TLS-termination conf that fails `nginx -t`. The tool's own docs must not describe an install path that doesn't run.
6. **Silent mutation of the host.** Install/bench scripts that switch the host default congestion control, rewrite `/etc/hosts` outside the `--hosts` opt-in, or shape traffic without a `--restore` escape — current `install.sh` correctly never touches default CC (kit/install.sh:57-69) and `cgo shape --restore` exists (docs/deploy.md:99-104); keeping those properties is the line.
7. **Secret in the image.** Token committed, baked `COPY .env`, or token logged — current `CGO_API_TOKEN:?set` pattern (docker-compose.yml:12) plus hygiene gate (verify.py:307-372) already enforce; the release must keep both green.
8. **No rollback.** A release without a named prior digest + a one-command rollback that also restores the edge shape is a trap — `docs/deploy.md:183`'s rollback sentence must become literally executable.

---
*Teams checklist: every quantitative claim above traces to `file:line` or R1-R7; everything else carries `unverified — primary source not found` (OCI label-key namespace, PromQL `rate()` semantics, promtool availability, cosign key strategy, current Nginx minimum version).*
