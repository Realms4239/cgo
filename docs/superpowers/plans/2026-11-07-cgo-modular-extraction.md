# CGO Modular Extraction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract twelve independently versioned tools from the monolith in dependency order, each with its own `go.mod`, README, semver, tests, SBOM, and signed provenance.

**Architecture:** Keep the Go measurement core and sealed evidence untouched. Extract pure-logic first (`model` → `metrics` → `results`/`figures`), then Linux-gated (`qdisc`/`remediate`), then experiment (`probe`/`audit`), then ops surface (`api`/`doctor`/`iface`), then presentation (`cite`), then distribution (`dist`/`isolate`) and defence (`demo`). Monorepo `require` + `replace` bridge until cutover.

**Tech Stack:** Go 1.25 (stdlib only, no new Go dependency without approval), Nginx, Prometheus exposition format (text), Podman/Docker compose (OCI), systemd, Vitest + Playwright (existing).

## Global Constraints

- No live audit, bench, CHR, eBPF, `.dist`, BBRv3, RouterOS or network traffic as evidence. Tests use fixtures only.
- No mutation is unauthenticated; reads require authentication on any non-loopback bind.
- Rate limiting is enforced at Nginx ingress for all POST mutations.
- `/metrics` honors Prometheus exposition contract: `# HELP`/`# TYPE` before first sample, unique `(name,labels)`, correct `Content-Type` (`text/plain; version=0.0.4`).
- Absent is never `0`; Prometheus absent maps to “no sample”, never to `0`.
- 220 ms / 95% is labelled “operator objective”, never a WMO mandate.
- Frozen corpus (`book/data/runs`, CSV schemas) is never touched; provenance is additive only.
- No new dependency without explicit human approval. Plan adds no Go module beyond the extractions listed in each task; Nginx/Prometheus/Grafana are system packages, not Go modules.
- French comments and user-facing text; English identifiers. No `git add -A`; stage explicit paths. Never merge to `main` without the human's decision.

---

## Source map

### Existing boundaries to preserve
- `pkg/model/*.go` — profiles P1–P4, qdiscs, CCs, gates G0–G7 (kernel, stdlib only)
- `pkg/metrics/*.go` — `Percentile`, `Summarize`, `CostARPerH`, tiers, JFI, verdicts
- `pkg/results/*.go`, `pkg/figures/*.go` — frozen scan, SVG, `ProvenanceSHA/Hash8`
- `pkg/qdisc/*.go`, `pkg/remediate/*.go` — `tc` layer, Linux + RouterOS generators
- `pkg/probe/*.go`, `pkg/audit/*.go` — sondes, 60 s phased audit, sealed reports
- `pkg/api/*.go` — REST/SSE surface (stdlib only), `New(d Deps) Handler`, `crossSiteGuard`
- `internal/kit/*.go`, `internal/vm/*.go` — deployment engine + VM discovery (shell, never library)
- `cmd/cgo/*.go`, `cmd/meteolink/*.go`, `cmd/cgo-gui/*.go`, `cmd/cgo-chr/*.go` — binaries
- `web/frontend/src/lib/citation.ts`, `archive.ts`, `auditFlow.ts` — normalization points
- `docs/deploy.md`, `docs/VERIFY.md`, `docs/CHECKPOINTS.md` — ledger + deploy fallback
- `docs/research/2026-11-07-modular-toolkit-morph.md` + 36 per-module files — primary-source basis

### New boundaries
- `modules/model/go.mod` — `cgo/model` submodule (first extraction)
- `modules/metrics/go.mod` — `cgo/metrics` submodule
- `modules/results/go.mod` — `cgo/results` submodule
- `modules/figures/go.mod` — `cgo/figures` submodule
- `modules/qdisc/go.mod` — `cgo/qdisc` submodule
- `modules/remediate/go.mod` — `cgo/remediate` submodule
- `modules/probe/go.mod` — `cgo/probe` submodule
- `modules/audit/go.mod` — `cgo/audit` submodule (+ `cgo-audit` CLI)
- `modules/api/go.mod` — `cgo/api` submodule
- `modules/doctor/go.mod` — `cgo/doctor` submodule
- `modules/iface/go.mod` — `cgo/iface` submodule
- `modules/cite/go.mod` — `cgo/cite` submodule (Go+TS)
- `modules/attest/go.mod` — `cgo/attest` CLI
- `modules/exporter/go.mod` — `cgo/exporter` (Prometheus honesty exporter)
- `modules/guard/go.mod` — `cgo/guard` (auth + rate limit)
- `modules/demo/go.mod` — `cgo-demo` (doctor-verify arc)
- `deploy/nginx/cgo.conf`, `deploy/Dockerfile`, `deploy/docker-compose.yml`, `deploy/prometheus/rules.yml` — already present from Slice 28 (S1/S4/S5)

---

### Task 1: Extract `cgo/model` (S-A kernel)

**Files:**
- Create: `modules/model/go.mod`
- Create: `modules/model/README.md`
- Copy: `pkg/model/*.go` → `modules/model/` (identical content, package name unchanged)

**Interfaces:**
- Consumes: nothing (stdlib only: `sync`, `strings`)
- Produces: `package model` with `type Profile`, `type Qdisc string`, `type Event`, `func DefaultCCs() []string`, gate constants G0–G7, CSV header constants

- [ ] **Step 1: Create module skeleton**

```go
// go.mod
module github.com/Realms4239/cgo/model

go 1.25
```

```markdown
<!-- modules/model/README.md -->
# cgo/model
Domain types for CGO: profiles P1–P4, qdiscs, CCs, gates G0–G7, CSV headers.
Stdlib only. Versioned with semver; breaking changes bump MAJOR.
```

- [ ] **Step 2: Copy source verbatim**

Run: `cp pkg/model/*.go modules/model/`
Expected: `modules/model` now contains every `.go` file from `pkg/model` with zero edits to logic, comments kept in French.

- [ ] **Step 3: Verify standalone build**

Run: `cd modules/model && go test ./... -count=1`
Expected: PASS (all `pkg/model` tests pass standalone).

- [ ] **Step 4: Bridge the monorepo**

In root `go.mod` add:

```
require github.com/Realms4239/cgo/model v0.1.0
replace github.com/Realms4239/cgo/model => ./modules/model
```

Run: `go test ./pkg/model -count=1` (still passes via bridge).
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add modules/model/go.mod modules/model/README.md modules/model/*.go go.mod go.sum
git commit -m "feat(model): extract cgo/model submodule"
```

---

### Task 2: Extract `cgo/metrics` (S-A)

**Files:**
- Create: `modules/metrics/go.mod`
- Create: `modules/metrics/README.md`
- Copy: `pkg/metrics/*.go` → `modules/metrics/` (logic verbatim)

**Interfaces:**
- Consumes: `github.com/Realms4239/cgo/model` (via `require` + `replace`)
- Produces: `func Percentile`, `func Summarize`, `func CostARPerH`, `type PriceTier`, `func JFI`, `func AssessVerdict`

- [ ] **Step 1: Create module**

```go
module github.com/Realms4239/cgo/metrics

go 1.25

require github.com/Realms4239/cgo/model v0.1.0

replace github.com/Realms4239/cgo/model => ../model
```

- [ ] **Step 2: Copy source**

Run: `cp pkg/metrics/*.go modules/metrics/`
Expected: no logic change; only import paths updated if they referenced `pkg/model`.

- [ ] **Step 3: Verify**

Run: `cd modules/metrics && go test ./... -count=1`
Expected: PASS.

- [ ] **Step 4: Bridge**

In root `go.mod`: `require github.com/Realms4239/cgo/metrics v0.1.0` + `replace => ./modules/metrics`. Run: `go test ./pkg/metrics -count=1`. Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add modules/metrics/ go.mod go.sum
git commit -m "feat(metrics): extract cgo/metrics submodule"
```

---

### Task 3: Extract `cgo/results` + `cgo/figures` (S-B)

**Files:**
- Create: `modules/results/go.mod`
- Create: `modules/figures/go.mod`
- Create: `modules/results/README.md`, `modules/figures/README.md`
- Copy: `pkg/results/*.go`, `pkg/figures/*.go` → respective modules

**Interfaces:**
- Consumes: `cgo/model`, `cgo/metrics`
- Produces: `func Scan`, `func ProvenanceSHA/Hash8`, `func Generate` (SVG)

- [ ] **Step 1: Create modules (verbatim `go.mod` shape)**

```go
module github.com/Realms4239/cgo/results

go 1.25

require (
  github.com/Realms4239/cgo/metrics v0.1.0
  github.com/Realms4239/cgo/model v0.1.0
)

replace (
  github.com/Realms4239/cgo/metrics => ../metrics
  github.com/Realms4239/cgo/model => ../model
)
```

Same shape for `figures` (depends on `results` + `model`).

- [ ] **Step 2: Copy + verify**

Run: `cp pkg/results/*.go modules/results/ && cp pkg/figures/*.go modules/figures/`
Run: `cd modules/results && go test ./... -count=1 && cd ../figures && go test ./... -count=1`
Expected: PASS in both.

- [ ] **Step 3: Bridge + commit**

Add `require` + `replace` for both in root `go.mod`, run `go test ./pkg/results ./pkg/figures -count=1` (PASS), then:

```bash
git add modules/results/ modules/figures/ go.mod go.sum
git commit -m "feat(results): extract cgo/results + cgo/figures"
```

---

### Task 4: Extract `cgo/qdisc` + `cgo/remediate` (S-C)

**Files:**
- Create: `modules/qdisc/go.mod`
- Create: `modules/remediate/go.mod`
- Copy: `pkg/qdisc/*.go`, `pkg/remediate/*.go`

**Interfaces:**
- Consumes: `cgo/model`
- Produces: `func Reset`, `func ApplyNetem`, `func ApplyShaper`, `func LinuxTC`, `func RouterOS`, `func Recommend`

- [ ] **Step 1: Create modules + copy**

Run: `mkdir -p modules/qdisc modules/remediate && cp pkg/qdisc/*.go modules/qdisc/ && cp pkg/remediate/*.go modules/remediate/`
Write `go.mod` files (stdlib + `cgo/model` only).

- [ ] **Step 2: Pin FQ-CoDel (RFC 8290)**

Ensure `qdisc.go` leaf for `fq_codel` carries explicit `target 5ms interval 100ms limit 10240 quantum 1514 flows 1024`. Add test:

```go
func TestFQCoDel_PinsRFC8290Params(t *testing.T) {
  s := fqCoDelString("eth0")
  for _, want := range []string{"target 5ms", "interval 100ms", "limit 10240", "quantum 1514", "flows 1024"} {
    if !strings.Contains(s, want) { t.Fatalf("missing %q", want) }
  }
}
```

Run: `cd modules/qdisc && go test -run TestFQCoDel -v` → PASS.

- [ ] **Step 3: Bridge + commit**

Root `go.mod` requires/replaces for both; `go test ./pkg/qdisc ./pkg/remediate -count=1` PASS; then commit (same pattern as Task 3).

---

### Task 5: Extract `cgo/probe` + `cgo/audit` (S-D)

**Files:**
- Create: `modules/probe/go.mod`, `modules/audit/go.mod`
- Create: `modules/audit/cmd/cgo-audit/main.go` (thin CLI over `audit.Run`)
- Copy: `pkg/probe/*.go`, `pkg/audit/*.go`

**Interfaces:**
- Consumes: `cgo/model`, `cgo/metrics`, `cgo/qdisc`, `cgo/remediate`
- Produces: `func Ping`, `func SmallObject`, `func BulkSendTo`, `func BulkDownloadTo`; `func Run`, `func BuildReport`

- [ ] **Step 1: Create modules + copy + CLI**

Write `modules/audit/cmd/cgo-audit/main.go`:

```go
package main

import (
  "os"
  "github.com/Realms4239/cgo/audit"
)

func main() {
  // flags: --site, --link-type, --duration (default 60)
  if err := audit.Run(audit.Params{Site: "site", LinkType: "4g", Duration: 60}); err != nil {
    os.Exit(1)
  }
}
```

- [ ] **Step 2: Verify**

Run: `cd modules/probe && go test ./... -count=1` and `cd modules/audit && go test ./... -count=1` → PASS.
Run: `cd modules/audit && go run ./cmd/cgo-audit --help` → prints usage, exit 0.

- [ ] **Step 3: Bridge + commit**

Root requires/replaces; `go test ./pkg/probe ./pkg/audit -count=1` PASS; commit.

---

### Task 6: Extract `cgo/api` + `cgo/doctor` + `cgo/iface` (S-E)

**Files:**
- Create: `modules/api/go.mod`, `modules/doctor/go.mod`, `modules/iface/go.mod`
- Copy: `pkg/api/*.go`, `pkg/doctor/*.go`, `pkg/iface/*.go`

**Interfaces:**
- Consumes: all promoted leaves
- Produces: `func New(d Deps) Handler`, `func NewHub() *Hub`, `func Report(mode string)`, `func New() *Sampler`

- [ ] **Step 1: Create modules + copy**

Same `go.mod` pattern (stdlib + `cgo/model` + `cgo/metrics` as needed).

- [ ] **Step 2: Verify + bridge**

`cd modules/api && go test ./... -count=1` PASS; `cd modules/doctor && go test ./... -count=1` PASS; `cd modules/iface && go test ./... -count=1` PASS.
Root bridge; `go test ./pkg/api ./pkg/doctor ./pkg/iface -count=1` PASS; commit.

---

### Task 7: Extract `cgo/cite` + proof panel (S-F)

**Files:**
- Create: `modules/cite/go.mod` (Go side, if Go wrapper needed) + `web/frontend/src/lib/citation.ts` stays canonical; add `modules/cite/README.md`
- Modify: `web/frontend/src/lib/citation.test.ts` (already updated: seal → `""`)

**Interfaces:**
- Consumes: `AuditReport.provenance.source_sha256`, `AuditReport.audit_id`
- Produces: `auditCitation(report) string` — `""` without `source_sha256`; else RFC-style citation with full seal

- [ ] **Step 1: Verify citation guard**

Run: `cd web/frontend && npx vitest run src/lib/citation.test.ts` → 4/4 PASS (seal guard, empty-id guard, sealed path).

- [ ] **Step 2: Document the contract**

Add to `modules/cite/README.md` the exact rule: citation is `""` when `!report.provenance?.source_sha256`; seal format `/^[0-9a-f]{64}$/i` becomes required in `cgo-cite` v2.

- [ ] **Step 3: Commit**

```bash
git add web/frontend/src/lib/citation.ts web/frontend/src/lib/citation.test.ts modules/cite/
git commit -m "feat(cite): single citation guard"
```

---

### Task 8: Ops plane — `cgo/guard`, `cgo/attest`, `cgo/exporter`, `cgo-lite`, `cgo-demo`

**Files:**
- Create: `modules/guard/go.mod` (auth + rate-limit library)
- Create: `modules/attest/go.mod` + `cmd/cgo-attest/main.go`
- Create: `modules/exporter/go.mod` (real Prometheus collector)
- Already present: `deploy/nginx/cgo.conf`, `deploy/Dockerfile`, `deploy/docker-compose.yml`, `deploy/prometheus/rules.yml`, `docs/runbooks/smoke-alerts.md`

**Interfaces:**
- Consumes: `cgo/api` (promoted), `pkg/results` (frozen scan)
- Produces: `requireAuth`, `attest`, `renderMetrics` (honest: absent ⇒ no sample), SLO recording rules `cgo:slo:deadline_ok_pct_5m`

- [ ] **Step 1: Verify Nginx ingress**

Check `deploy/nginx/cgo.conf` has `limit_req_zone` named `cgo_reads`/`cgo_mutations` and `limit_req_status 503`. Manual `nginx -t` when available; else record in `docs/deploy.md` S1 section (already present).

- [ ] **Step 2: Verify `/metrics` honesty**

Run: `go test ./pkg/api -run TestMetrics -v` → PASS (HELP/TYPE + absent-not-zero tests green).

- [ ] **Step 3: Verify OCI + SLO**

Run: `GOOS=linux GOARCH=amd64 go build -o bin/cgo ./cmd/cgo` → PASS.
Run: `docker compose -f deploy/docker-compose.yml config` (or `docker compose config` if Docker present) → PASS.

- [ ] **Step 4: Commit (explicit)**

```bash
git add deploy/ docs/runbooks/smoke-alerts.md docs/deploy.md
git commit -m "feat(deploy): ingress Nginx + OCI/SLO ops plane"
```

---

### Task 9: Whole-app verification and slice closeout

**Files:**
- Modify: `docs/CHECKPOINTS.md`
- Modify: `docs/VERIFY.md` (PENDING row only; verifier signs separately)
- Modify: `docs/superpowers/plans/README.md` (slice row → Pending verifier)

- [ ] **Step 1: Run the gates**

Run:

```bash
cd web/frontend && npm test && npm run build && cd ..
python tools/verify.py
go build ./...
go test ./... -count=1
```

Expected: tests green, build ~357 KB gz, `tools/verify.py` 6/6 (or SKIP named). If `make` is absent, record `python tools/verify.py` instead.

- [ ] **Step 2: Restore build artifact**

Run: `git checkout -- web/frontend/dist/index.html 2>/dev/null || true`
Run: `git status --short` — only intended source/docs/deploys are present; no secrets, no frozen datasets.

- [ ] **Step 3: Checkpoint**

Append to `docs/CHECKPOINTS.md`: command, date (UTC), host, exit codes, and what stayed unproven (CHR, WIS, live bench, Mosquitto).

- [ ] **Step 4: Independent verifier pass**

Dispatch a separate subagent/verifier with this exact prompt:

> Re-run Go/frontend suites yourself; inspect the four cockpit screens and the S1–S5 artifacts; verify absent has no fake values; verify 890×768/1920×1080 overflow; verify no per-row waterfall on 179 rows; verify `401` on unauthenticated mutation and `503` under burst; return `RESIDUALS ACCEPTED: Y/N` and a signature block or `withheld`.

- [ ] **Step 5: Ledger transcription**

Only after verifier sign: exact commands/host/date/exit codes into `docs/VERIFY.md`; slice row to the verifier's disposition; no rewriting of historical rows.

- [ ] **Step 6: Commit closeout**

```bash
git add docs/CHECKPOINTS.md docs/VERIFY.md docs/superpowers/plans/README.md
git commit -m "docs: close modular extraction slice 29"
```

---

## Parallelization map

- Task 1 (`model`) must be first — everything else depends on it.
- Tasks 2–3 (`metrics`, `results/figures`) depend on Task 1.
- Task 4 (`qdisc/remediate`) depends on Task 1 only — independent of 2–3.
- Task 5 (`probe/audit`) depends on Tasks 1–4.
- Task 6 (`api/doctor/iface`) depends on Tasks 1–5.
- Task 7 (`cite`) is independent of 1–6 — can run in parallel.
- Task 8 (ops plane) depends on Tasks 1–6.
- Task 9 must remain at the end.

## Risk register

| Risk | Mitigation |
|---|---|
| Auth breaks existing frontend flows | Bearer token via env; frontend sends `Authorization` header when token is configured; loopback dev mode preserved |
| `/metrics` becomes a NaN/zero leak | Honesty contract tests (`TestMetrics_AbsentIsNotZero`); absent = no sample |
| Nginx `limit_req` throttles legitimate bulk import | Mutations `1r/s burst=3 nodelay`; calibrate after first load test |
| FQ-CoDel pin alters existing bench numbers | Pin explicitly, re-run `go test ./pkg/qdisc`, flag in checkpoint |
| OCI rollback leaves edge shaped | `cgo shape --restore` as post-deploy gate in S4 runbook |
| Live-run temptation | No live tests in this plan; human gate listed in spec §Open decisions |

## Definition of done

- All twelve extractions merged in sequence, each with green gates, SBOM, and signed provenance.
- Monorepo still builds; no frozen-corpus edit; no live run; no new Go dependency without signed human gate.
- Jury air-gap arc reproduces offline; live claims remain explicitly BLOCKED where unproven.
- Research index (`docs/research/README.md` + 41 files) cited for every MUST above.
