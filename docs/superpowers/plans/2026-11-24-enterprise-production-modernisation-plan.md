# Enterprise Production Modernisation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Checkbox (`- [ ]`) syntax for tracking.

**Goal:** make the measured network stack reproducible, observable with declared SLOs, and safe at its API boundary without changing the frozen corpus or native gates.

**Architecture:** four sequential slices. Task 1 adds a declarative topology contract and read-only drift check. Task 2 extends the existing `/metrics` exporter with real audit/seal/fidelity series and a declared SLO/error budget. Task 3 adds documented auth/ownership/rate-limit policy with hostile route tests. Task 4 adds supply-chain/runbook/version proof. No new runtime dependency, no Kubernetes, no gate edit.

**Tech Stack:** Go (root module + `api`, `metrics`, `audit`), existing React frontend, existing CI workflows, shell bench executor.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-11-24-enterprise-production-modernisation-design.md`.
- Frozen corpus, book, thesis, C:\END, gates, permits, ledgers, protected config: never edited.
- No `git add -A`; explicit paths only; no merge; no new dependency VERSION; no runtime clang.
- Absence is explicit; no fabricated metric, SLO green, rate-limit pass or readiness claim.
- Rate limits: 60/min read, 10/min mutate, burst 20/5; declared here, never widened at runtime.
- Six gates green after each slice; pre-existing false positives recorded, never repaired by gate edit.
- No live bench or host token work in these tasks without a separate same-day token.

---

### Task 1: reproducible topology contract and drift check

**Files:**
- Create `kit/topology.contract.json`, `cmd/topocheck/main.go`, `cmd/topocheck/main_test.go`
- Modify `pkg/campagne/writer.go` (add contract fingerprint to manifest, additive)

**Interfaces:**
- Consumes: existing `kit/run-dualns.sh` topology facts and `model.Profile`.
- Produces: `type Contract struct { Namespaces []string; Ifaces []string; Qdisc string; CapMbps float64; DelayMs float64; ShapeMode string }`; `func Load(path string) (Contract, error)`; `func Check(c Contract, facts Facts) []Drift`.

- [ ] **Step 1: RED oracles.** A contract with `shape_mode=ingress` and live facts `egress` yields a Drift; a missing contract file refuses; a contract fingerprint is stable across runs. Run `go test -count=1 ./cmd/topocheck/...` -> RED.
- [ ] **Step 2: Implement** `Load` with strict JSON decoding and `Check` comparing namespace/interface/qdisc/capacity/delay/shape-mode. No mutation, no shell exec.
- [ ] **Step 3: Wire manifest** with additive `topology_contract_sha256` field; no existing field changed.
- [ ] **Step 4: GREEN, gates once, commit** `feat(repro): declarative topology contract and drift check`.

### Task 2: observability and declared SLO

**Files:**
- Modify `api/metrics.go`, `api/server.go` (metrics handler only), `metrics/slo.go` (new), `metrics/slo_test.go` (new)

**Interfaces:**
- Consumes: audit report fields, seal artifact presence, `metrics.FidelityEMDMax`.
- Produces: `func ErrorBudget(observedOK, observedTotal uint64, target float64) (burnRate, remaining float64, ok bool)`.

- [ ] **Step 1: RED oracles.** Zero observations yields `ok=false` and no green; 95/100 at target 0.99 yields burn>1; NaN refuses. Run `go -C metrics test -count=1 ./...` -> RED.
- [ ] **Step 2: Implement** the SLO helper and add exact series from the spec: `cgo_audit_info`, `cgo_kernel_truth_source_info`, `cgo_seal_present`, `cgo_emd`, `cgo_emd_ceiling`, `cgo_gate_status`. Omit a series when its source is absent.
- [ ] **Step 3: GREEN, gates once, commit** `feat(observability): real audit and fidelity metrics with SLO`.

### Task 3: API boundary policy

**Files:**
- Create `api/ratelimit.go`, `api/ratelimit_test.go`, `api/boundary_test.go`
- Modify `api/server.go` (wrap mux with rate limiter only)

**Interfaces:**
- Consumes: `Principal` from `api/auth.go`, route patterns from `api/server.go`.
- Produces: `func RateLimit(next http.Handler, readsPerMin, mutatesPerMin, readBurst, mutateBurst int) http.Handler`.

- [ ] **Step 1: RED oracles.** Unauthenticated mutate -> 401 before handler; cross-owner -> 403 before handler; 11th mutate in a minute -> 429 with Retry-After; read limit 61 -> 429; no secret/path in error bodies. Run `go -C api test -count=1 ./...` -> RED.
- [ ] **Step 2: Implement** a per-principal token bucket with the declared limits; wrap only the mux, no handler logic change.
- [ ] **Step 3: GREEN, gates once, commit** `feat(api): gateway-style auth and rate-limit boundary`.

### Task 4: production hardening proof

**Files:**
- Create `docs/verification/2026-11-24-enterprise-production/README.md`, `runbooks/index.md`

**Interfaces:**
- Consumes: CI workflow facts, native gates, existing token/rollback artifacts.
- Produces: a runbook index naming each live operation, token, rollback and last rehearsal date.

- [ ] **Step 1: RED oracle.** A live operation without token, rollback or rehearsal date is listed as a gap, never silently omitted; a doc test asserts the index covers all known operations.
- [ ] **Step 2: Implement** the docs only; no gate/CI/ledger edit; no invented rehearsal date.
- [ ] **Step 3: GREEN, gates once, commit** `docs(prod): runbook and supply-chain proof`.

## Self-Review

- Spec sections 2-5 -> Tasks 1-4; acceptance -> RED oracles in each.
- No placeholder: exact files, signatures, metric names and limits.
- Type consistency: `Contract`, `ErrorBudget`, `RateLimit` each defined once.
- No schema bump, dependency, gate/ledger/corpus edit; absence explicit.
