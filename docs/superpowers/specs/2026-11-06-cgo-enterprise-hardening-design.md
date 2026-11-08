# CGO — Enterprise Hardening Slice 28 (Design)

**Status:** approved direction, pending implementation plan.
**Date:** 2026-11-06
**Scope:** harden the measurement core so it can be operated, audited and deployed as an enterprise platform for DGM. No new measurement science, no live runs, no new dependency without the human gates listed below.

## Context

The primary-source research in `docs/research/2026-11-06-enterprise-standard-cgo-platform.md` establishes ground truth in the repo:

- Single Go binary with embedded React dashboard; no DB, no service dependency.
- Frozen corpus and sealed evidence (CSV + `manifest.json` + `hash8` + dual audit seals) are the strongest enterprise asset.
- Dual-namespace bench is real (`kit/cgo-bench-dualns.sh`); CHR validation is not (fixture replay only).
- Security posture is a hard **FAIL** by the repo's own honesty docs: no authN/Z, no CSRF, no rate limit (`docs/deploy.md:159-177`).
- No `/metrics` endpoint, no alerting, no container packaging, no multi-site story.
- 220 ms / 95% is a DGM operator objective, not a WMO mandate (`pkg/metrics/cost.go:75-83`).
- Three pasted-proposal numbers are rejected as unverifiable: microsecond probing, RAM footprints, 3-file scope.

Slice 27 (operator cockpit) is the current HEAD lineage. Slice 28 does not rewrite it. It puts an enterprise perimeter around it.

## Problem statement

An enterprise buyer will not procure a system where anyone reaching the port can reshape the edge, where there is no scrapeable health contract, and where deployment is a tarball + systemd unit with no image, no rollback and no runbook. That is the whole of Slice 28.

## Users

- **Primary: DSI / NOC operator** — needs authenticated access, rate-limited mutation APIs, scrapeable health, and a rollback that works at 3 a.m.
- **Secondary: Master 2 jury / evaluator** — needs provenance and invariants, not demo theatre.
- **Stakeholder: DGM** — needs a system that can be patched, audited and supported after the author leaves.

## Scope

### In scope
- **S1 Ingress hardening:** Nginx in front of every 9090 (TLS termination, `limit_req_zone` for mutations). Config-only, no Go change.
- **S2 AuthN/Z:** authenticated sessions (mTLS or OIDC bearer) on all mutation APIs; CSRF on cookie flows; unauthenticated mutation → `401`.
- **S3 Prometheus `/metrics`:** emission honoring the exposition contract; Blackbox + Node exporters as official images; Grafana dashboards as versioned JSON.
- **S4 Container packaging + cgroups:** OCI image digest-pinned, `cpu.max`/`memory.max`, compose + systemd fallback, rollback by digest.
- **S5 SLO/runbooks:** recording/alerting rules for breach / dead-probe / invalid-iface; runbooks beyond host-acceptance.

### Out of scope
- New measurement science (AQM/BBR matrix, eBPF, BBRv3, `.dist`, IFB live acceptance).
- WIS 2.0 / CAP publication; Mosquitto multi-site fan-out.
- Live bench runs, CHR boot, any traffic mutation outside the sealed audit path.
- Replacing the cockpit UI or the frozen-corpus format.
- New backend endpoints beyond `/metrics` and the auth guard.

## Product constraints

- A missing measurement is **absent**, never `0`, never NaN-as-zero. Prometheus mapping preserves `iface_valid:false` ⇒ absent.
- `load_attempted` vs `load_measured` are always exported as a pair; breach is never exported without its measured flag.
- 220 ms / 95% is labelled “operator objective” everywhere it appears.
- Every mutation requires authentication; reads require authentication on any non-loopback bind.
- No number leaves the repo without a verifier pass (per `AGENTS.md`).
- Live runs, new dependencies, CHR procurement and exposure decisions each require explicit human approval (see “Open decisions”).

## Architecture

### 6.1 MUST (non-negotiable this slice)
1. **AuthN/Z boundary in front of all mutation APIs** — every `POST /api/run/*`, `/api/shape`, `/api/watch`, `/api/burst`, `/api/audit/*`, `/api/cost/tier`, `/api/profile/*`, `/api/figures/regen` requires an authenticated session. CSRF tokens on cookie sessions, or same-site + bearer for API clients. Reuse `pkg/api/server.go` handler seams; no parallel auth server.
2. **Rate limiting at ingress** — Nginx `limit_req_zone $binary_remote_addr zone=cgo:10m rate=…` + per-location `limit_req` for mutations. Mechanism ships first; exact r/s calibrated against measured campaign rates.
3. **Prometheus `/metrics` honoring the exposition contract** — `# HELP`/`# TYPE`, unique `(name,labels)`, correct `Content-Type`. Minimal contract:
   - `cgo_audit_runs_total{site,link_type,status}`
   - `cgo_audit_load_attempted_total{site}` + `cgo_audit_load_measured_total{site}`
   - `cgo_small_p95_ms{site,profile}` or `cgo_rtt_p95_ms{site,direction}` (suffix `_ms` kept for drift safety)
   - `cgo_bloat_delta_ms{site,direction}`, `cgo_jfi_pct{site}` with separate `cgo_jfi_measured{site}`
   - `cgo_cap_deadline_ok_pct{site}`, `cgo_cap_sla_breach{site}`, `cgo_cap_sla_measured{site}`
   - `cgo_qdisc_drops_total{profile,qdisc,cc,direction}`
   - `cgo_iface_rx_mbps`/`cgo_iface_tx_mbps`/`cgo_iface_drops_total` with `cgo_iface_valid` boolean companion
   - `cgo_run_status{run_id,profile,qdisc,cc}` gauge
   - `cgo_build_info{version}`
4. **Shaping-plane auth** — `POST /api/shape` + `cgo shape --restore` require the same auth as (1). Nginx MUST NOT proxy `/api/shape` to loopback without auth. Benches on mgmt `10.99.0.2:9090` move behind Nginx TLS; raw `--tls=false` only on loopback in CI.
5. **OCI + compose/systemd** — multi-arch OCI image digest-pinned; compose wiring dashboard + Nginx (+ exporters); systemd preserved as fallback; `cgo verify` + `cgo doctor` + `cgo shape --restore` as post-deploy gates. Rollback = prior digest.
6. **Provenance extension** — keep CSV schema + `manifest.json` + `hash8` + dual seals + `invocation`; add image digest + version stamp to `GET /api/health` and `manifest.json`. No relabeling of frozen runs.
7. **FQ-CoDel parameter pinning** — declare `target/interval/limit/quantum/flows` per RFC 8290 §5.2; `limit` MUST be set.

### 6.2 SHOULD (follow MUST quickly)
- Blackbox + Node exporters (official images).
- Grafana dashboards as versioned JSON (operator wall, audit comparison valid-only, SLO/CAP SLA labelled “operator objective”, provenance).
- Alerting: recording/alerting rules for sustained breach, dead probes, invalid iface; cgroups `cpu.max`/`memory.max` + `cpu.stat` throttling exported.
- Runbooks: auth-rotation, Nginx revoke, rollback, target-down, offload-fail, corpus seal verification.

### 6.3 DEFERRED (explicitly not now)
- WIS 2.0 / CAP publication; Mosquitto multi-site; CHR/QEMU CI validation; eBPF / BBRv3 / `.dist`; DB migration for runs. Each deferred item names its unblock condition (primary docs fetched, image procured, human grant signed).

## Normalization

Slice 28 replaces ad-hoc honesty checks with one normalization point per concern:

- **State normalization** — one `viewState`/`health`/`archive` rule for `Valide`, `Observation seule`, `Dégradé`, `Preuve manquante`. No component re-derives it.
- **Citation normalization** — one guard: citation is disabled without `provenance.source_sha256`. Archive path and card path share it.
- **Metrics normalization** — Prometheus mapping preserves the existing `valid:false ⇒ absent` semantics; absent never becomes `0`.
- **Token normalization** — cockpit tokens already frozen in slice 27; Slice 28 adds no new visual language.

## Data flow

```
Operator → Nginx (TLS, limit_req, auth) → Go binary (handlers, /metrics, sealed APIs)
                                           ├─ Prometheus scrape → /metrics → Grafana
                                           ├─ Node/Blackbox exporters → Prometheus
                                           └─ Sealed freeze → CSV + manifest + hash8
```

No new measurement is invented by the ops plane. Grafana renders; Go measures; Nginx guards.

## Error handling

- Unauthenticated mutation → `401` with `WWW-Authenticate`; CSRF failure → `403`; rate-limited → `503` (Nginx `limit_req_status`).
- Invalid `run_id` → `400` (`validRunID`); tier change mid-campaign → `409` (existing).
- Scrape failure → Prometheus target-down alert; `iface_valid:false` ⇒ gap, not zero.

## Testing & verification

- Unit: per-handler auth tests (401/403), `/metrics` HELP/TYPE + label uniqueness, `cgo verify` after packaging.
- Integration: Nginx config test (synthetic burst → 503); compose up → health green; rollback by digest.
- Browser: existing 320 frontend tests + Playwright cockpit matrix remain (no UI rewrite in this slice).
- Gates: `make verify` / `python tools/verify.py`, `go test ./...`, `go build ./...`, `python book/tools/audit-archive.py` (when `book/` present). Exit 2 is **not** green; BLOCKED rows stay named.

## Slices in dependency order

1. **S1 Ingress** — Nginx + TLS + rate limit.
2. **S2 AuthN/Z** — mutation APIs.
3. **S3 `/metrics` + exporters + dashboards.**
4. **S4 OCI/cgroups packaging.**
5. **S5 SLO/runbooks.**
6. **S6 CHR (human-gated, deferred).**
7. **S7 Multi-site (deferred).**
8. **S8 WIS (deferred).**

At most three active at once. S1–S3 first.

## Open decisions requiring human approval

1. Live bench window + host (Linux 6.8).
2. CHR image licence + version pin + download approval.
3. Each new dependency: Nginx, Prometheus, Grafana, exporters, OCI runtime, Mosquitto (deferred).
4. WIS 2.0 / CAP scope and primary-doc fetch.
5. Case-study operator grant.
6. Port / network exposure (mgmt subnets, TLS cert issuance).
7. Tariff re-sourcing before any number leaves the repo.
8. Deletion of microsecond / RAM / 3-file sentences from thesis prose.

## Definition of done

- No unauthenticated mutation succeeds; rate limiting demonstrably `503`s a burst.
- `/metrics` scrapes clean with HELP/TYPE and the honesty contract intact.
- `docker compose up` → green health; rollback by digest restores the prior image.
- No new live run, no new dependency, and no frozen-corpus change without a signed human gate.
- Research file at `docs/research/2026-11-06-enterprise-standard-cgo-platform.md` is the cited source for every MUST above.
