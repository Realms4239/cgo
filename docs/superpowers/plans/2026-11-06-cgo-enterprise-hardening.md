# CGO Enterprise Hardening Slice 28 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the measurement core so it can be operated, audited and deployed as an enterprise platform (S1–S5), with normalization of state/citation/metrics rules.

**Architecture:** Keep the Go measurement core and sealed evidence untouched. Put Nginx (TLS + `limit_req`) in front of every 9090, add authenticated sessions on mutation APIs, expose a Prometheus `/metrics` honoring the exposition contract with the honesty mapping (`valid:false ⇒ absent`), and package the binary as a digest-pinned OCI image with compose/systemd fallback. No live runs, no WIS/Mosquitto/CHR in this plan.

**Tech Stack:** Go 1.25 (stdlib only, no new Go dependency), Nginx (rate limiting), Prometheus exposition format (text), Podman/Docker compose (OCI), systemd, Vitest + Playwright (existing).

## Global Constraints

- No live audit, bench, CHR, eBPF, `.dist`, BBRv3, RouterOS or network traffic as evidence. Tests use fixtures only.
- No mutation is unauthenticated; reads require authentication on any non-loopback bind.
- Rate limiting is enforced at Nginx ingress for all POST mutations.
- `/metrics` honors Prometheus exposition contract: `# HELP`/`# TYPE` before first sample, unique `(name,labels)`, correct `Content-Type` (`text/plain; version=0.0.4`).
- Absent is never `0`; Prometheus absent maps to “no sample”, never to `0`.
- 220 ms / 95% is labelled “operator objective”, never a WMO mandate.
- Frozen corpus (`book/data/runs`, CSV schemas) is never touched; provenance is additive only.
- No new dependency without explicit human approval. Plan adds no Go module; Nginx/Prometheus/Grafana are system packages, not Go modules.
- French comments and user-facing text; English identifiers. No `git add -A`; stage explicit paths. Never merge to `main` without the human's decision.

---

## Source map

### Existing boundaries to preserve
- `pkg/api/server.go` — `New(d Deps) Handler`, `mux.HandleFunc` routes, `crossSiteGuard`, `writeJSON`/`writeErr`.
- `pkg/api/server_audit.go` — audit list/report handlers, `has_report` flag.
- `pkg/audit/audit.go` — honesty invariants, `load_attempted`/`load_measured`, CAP SLA.
- `pkg/metrics/cost.go` — `CAPDeadlineMs = 220`, `CAPBreachThresholdPct = 95.0` (operator objective).
- `pkg/qdisc/qdisc.go` — `ApplyShaper`, Reset-then-apply; FQ-CoDel params to pin.
- `web/frontend/src/lib/viewState.ts`, `health.ts`, `archive.ts`, `citation.ts` — normalization points.
- `docs/deploy.md` — systemd unit + `setcap` deploy path (fallback).
- `docs/research/2026-11-06-enterprise-standard-cgo-platform.md` — primary-source basis for every MUST.

### New boundaries
- `pkg/api/auth.go` — bearer-token auth guard + CSRF.
- `pkg/api/metrics.go` — Prometheus exposition handler.
- `deploy/nginx/cgo.conf` — TLS termination + `limit_req` ingress.
- `deploy/docker-compose.yml` — dashboard + Nginx + exporters, digest-pinned.
- `deploy/Dockerfile` — OCI image.
- `docs/superpowers/plans/2026-11-06-cgo-enterprise-hardening.md` — this plan.

---

### Task 1: Nginx ingress + rate limiting (S1)

**Files:**
- Create: `deploy/nginx/cgo.conf`
- Modify: `docs/deploy.md` (add S1 boundary note)
- Test: `deploy/nginx/cgo.conf` validated by `nginx -t` (manual gate in step 4)

**Interfaces:**
- Consumes: Go handlers at `127.0.0.1:9090` (`GET /api/health`, `POST /api/run/*`, `POST /api/shape`, `POST /api/audit/*`, `POST /api/burst`, `POST /api/cost/tier`).
- Produces: `limit_req_zone` named `cgo`; `limit_req_status 503`.

- [ ] **Step 1: Write the failing config check**

Create `deploy/nginx/cgo.conf` with the exact contract below. Do not add `server_name` tuning beyond this file.

```nginx
limit_req_zone $binary_remote_addr zone=cgo_reads:10m rate=30r/s;
limit_req_zone $binary_remote_addr zone=cgo_mutations:10m rate=1r/s;

server {
    listen 9090 ssl;
    server_name _;

    # TLS terminitation is configured by the operator; see docs/deploy.md S1.
    limit_req_status 503;

    location /api/health {
        limit_req zone=cgo_reads burst=10 nodelay;
        proxy_pass http://127.0.0.1:9090;
        proxy_set_header Host $host;
    }

    location /metrics {
        allow 127.0.0.1;
        deny all;
        limit_req zone=cgo_reads burst=10 nodelay;
        proxy_pass http://127.0.0.1:9090;
    }

    location /api/ {
        limit_req zone=cgo_mutations burst=3 nodelay;
        proxy_pass http://127.0.0.1:9090;
        proxy_set_header Host $host;
    }

    location / {
        limit_req zone=cgo_reads burst=10 nodelay;
        proxy_pass http://127.0.0.1:9090;
        proxy_set_header Host $host;
    }
}
```

- [ ] **Step 2: Validate syntax**

Run: `nginx -t -c deploy/nginx/cgo.conf`
Expected: `syntax is ok` / `test is successful`.

- [ ] **Step 3: Add deploy note**

Append to `docs/deploy.md` under a new `## S1 — Nginx boundary` heading:

```markdown
## S1 — Nginx boundary
All 9090 traffic enters through `deploy/nginx/cgo.conf` (TLS termination + `limit_req`). Direct `--tls=false` binds are loopback-only in CI. `POST` mutations that exceed `1r/s burst=3` receive `503`.
```

- [ ] **Step 4: Commit**

```bash
git add deploy/nginx/cgo.conf docs/deploy.md
git commit -m "feat(deploy): ingress Nginx + rate limiting"
```

---

### Task 2: AuthN/Z on mutation APIs (S2)

**Files:**
- Create: `pkg/api/auth.go`
- Create: `pkg/api/auth_test.go`
- Modify: `pkg/api/server.go` (wrap mutations with `requireAuth`)
- Test: `pkg/api/auth_test.go`

**Interfaces:**
- Produces: `func requireAuth(next http.HandlerFunc) http.HandlerFunc` — checks `Authorization: Bearer <token>` where token comes from `CGO_API_TOKEN` env var; empty env var = dev mode with loopback-only warning (still `401` for non-loopback `RemoteAddr`).
- Consumes: existing `mux.HandleFunc` POST routes in `New(d Deps)`.

- [ ] **Step 1: Write the failing test**

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRequireAuth_UnauthenticatedMutationIs401(t *testing.T) {
	os.Setenv("CGO_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CGO_API_TOKEN")
	h := requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/shape", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `go test ./pkg/api -run TestRequireAuth_UnauthenticatedMutationIs401 -v`
Expected: FAIL — `requireAuth` undefined.

- [ ] **Step 3: Implement `pkg/api/auth.go`**

```go
package api

import (
	"net/http"
	"os"
	"strings"
)

// requireAuth — mutations require a bearer token when CGO_API_TOKEN is set.
// Empty token is dev mode: loopback only, still 401 for non-loopback callers.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		want := os.Getenv("CGO_API_TOKEN")
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if want != "" && got == want {
			next(w, r)
			return
		}
		if want == "" && strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeErr(w, r, "authentification requise", http.StatusUnauthorized)
	}
}
```

- [ ] **Step 4: Wire mutations**

In `New(d Deps)`, wrap exactly these `POST` handlers with `requireAuth` (no GET changes):
`POST /api/cost/tier`, `POST /api/burst`, `POST /api/watch`, `POST /api/shape`, `POST /api/run/start`, `POST /api/run/stop`, `POST /api/run/skip`, `POST /api/audit/start`, `POST /api/audit/stop`, `POST /api/audit/toprofile`, `POST /api/profile/import`, `POST /api/profile/delete`, `POST /api/figures/regen`.

Example (applied to each line):

```go
mux.HandleFunc("POST /api/shape", requireAuth(func(w http.ResponseWriter, r *http.Request) {
	// existing body unchanged
}))
```

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/api -run TestRequireAuth -v && go test ./pkg/api/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/api/auth.go pkg/api/auth_test.go pkg/api/server.go
git commit -m "feat(api): auth sur les mutations"
```

---

### Task 3: Prometheus `/metrics` endpoint (S3)

**Files:**
- Create: `pkg/api/metrics.go`
- Create: `pkg/api/metrics_test.go`
- Modify: `pkg/api/server.go` (add `GET /metrics` route next to `GET /api/health`)
- Test: `pkg/api/metrics_test.go`

**Interfaces:**
- Produces: `GET /metrics` returning `text/plain; version=0.0.4` with `# HELP`/`# TYPE` before first sample for each metric; unique `(name,labels)`; histograms with `le="+Inf"`; no duplicate label sets.
- Consumes: honesty fields from audit/results (`load_attempted`, `load_measured`, `iface_valid`, `cap_sla_measured`).

- [ ] **Step 1: Write the failing test**

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics_EmitsHelpTypeAndContentType(t *testing.T) {
	h := New(Deps{Mode: "observe"})
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain; version=0.0.4") {
		t.Fatalf("bad Content-Type: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# HELP cgo_audit_runs_total") {
		t.Fatalf("missing HELP for cgo_audit_runs_total")
	}
	if !strings.Contains(body, "# TYPE cgo_audit_runs_total counter") {
		t.Fatalf("missing TYPE for cgo_audit_runs_total")
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `go test ./pkg/api -run TestMetrics_EmitsHelpTypeAndContentType -v`
Expected: FAIL — `/metrics` not found (404).

- [ ] **Step 3: Implement `pkg/api/metrics.go`**

```go
package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type metricSample struct {
	labels map[string]string
	value  string
}

// renderMetrics — pure Prometheus exposition builder.
// Absent is never emitted; the caller only pushes measured values.
func renderMetrics(samples map[string][]metricSample, help map[string]string) string {
	var b strings.Builder
	names := make([]string, 0, len(samples))
	for n := range samples {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "# HELP %s %s\n", name, help[name])
		typ := "gauge"
		if strings.HasSuffix(name, "_total") {
			typ = "counter"
		}
		fmt.Fprintf(&b, "# TYPE %s %s\n", name, typ)
		for _, s := range samples[name] {
			var parts []string
			for k, v := range s.labels {
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(parts)
			fmt.Fprintf(&b, "%s{%s} %s\n", name, strings.Join(parts, ","), s.value)
		}
	}
	return b.String()
}

func handleMetrics(getBuildInfo func() map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		samples := map[string][]metricSample{
			"cgo_audit_runs_total": {{labels: map[string]string{"status": "completed"}, value: "0"}},
			"cgo_audit_load_attempted_total": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_audit_load_measured_total": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_cap_sla_measured": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
			"cgo_iface_valid": {{labels: map[string]string{"site": "unknown"}, value: "0"}},
		}
		help := map[string]string{
			"cgo_audit_runs_total":              "Total audits by status",
			"cgo_audit_load_attempted_total":    "Load windows attempted",
			"cgo_audit_load_measured_total":     "Load windows measured",
			"cgo_cap_sla_measured":              "CAP SLA measured flag",
			"cgo_iface_valid":                   "Interface sampling valid flag",
		}
		_, _ = w.Write([]byte(renderMetrics(samples, help)))
	}
}
```

Note: initial values are `0` only because the test fixture has no live rows; the honesty rule is that *absent* series use a separate `*_measured` gauge rather than imputing `0` into the value gauge. Full live wiring to `pkg/audit` state is Task 4.

- [ ] **Step 4: Route**

In `New(d Deps)` after `GET /api/health`:

```go
mux.HandleFunc("GET /metrics", handleMetrics(func() map[string]string {
	return map[string]string{"version": d.Version}
}))
```

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/api -run TestMetrics -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/api/metrics.go pkg/api/metrics_test.go pkg/api/server.go
git commit -m "feat(api): expose /metrics Prometheus"
```

---

### Task 4: Live metrics wiring (honesty contract)

**Files:**
- Modify: `pkg/api/metrics.go`
- Test: `pkg/api/metrics_test.go`

**Interfaces:**
- Consumes: existing audit store state (`load_attempted`, `load_measured`, `iface_valid`, `cap_sla_measured`).
- Produces: `cgo_audit_load_attempted_total` vs `cgo_audit_load_measured_total` as a pair; `cgo_iface_valid` boolean; `cgo_cap_sla_measured` separate from breach.

- [ ] **Step 1: Write the failing test**

```go
func TestMetrics_AbsentIsNotZero(t *testing.T) {
	samples := map[string][]metricSample{
		"cgo_jfi_pct": {},
	}
	help := map[string]string{"cgo_jfi_pct": "JFI percent, only when measured"}
	out := renderMetrics(samples, help)
	if strings.Contains(out, "cgo_jfi_pct{") {
		t.Fatalf("absent JFI must not emit a zero sample")
	}
	if !strings.Contains(out, "cgo_jfi_measured") || !strings.Contains(out, "cgo_jfi_measured{") {
		t.Fatalf("absent JFI must still emit cgo_jfi_measured=0")
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `go test ./pkg/api -run TestMetrics_AbsentIsNotZero -v`
Expected: FAIL — `cgo_jfi_measured` not in builder.

- [ ] **Step 3: Implement**

In `handleMetrics`, emit `cgo_jfi_measured` with `0` when unmeasured and omit `cgo_jfi_pct` entirely. Apply the same pattern to `cgo_small_p95_ms`/`cgo_rtt_p95_ms` (omit value, emit `*_measured=0`), `cgo_cap_sla_measured`, and `cgo_iface_valid` (boolean, never a numeric zero masquerading as measured).

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/api -run TestMetrics -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/api/metrics.go pkg/api/metrics_test.go
git commit -m "feat(api): honesty mapping in /metrics"
```

---

### Task 5: Normalize citation + health on frontend

**Files:**
- Modify: `web/frontend/src/lib/citation.ts`
- Modify: `web/frontend/src/lib/citation.test.ts`
- Modify: `web/frontend/src/components/ArchiveTable.tsx` (gate `CITER` on `source_sha256`)
- Test: `web/frontend/src/lib/citation.test.ts`

**Interfaces:**
- Produces: `auditCitation(report)` returns `""` when `!report.provenance?.source_sha256` (single guard, shared by card and archive).
- Consumes: `AuditReport.provenance.source_sha256`.

- [ ] **Step 1: Write the failing test**

```ts
it('returns empty when the seal is missing, never a partial citation', () => {
  expect(auditCitation(report({ provenance: undefined }))).toBe('')
  expect(auditCitation(report({ provenance: { source_sha256: '' } }))).toBe('')
})
```

- [ ] **Step 2: Run and confirm failure**

Run: `cd web/frontend && npx vitest run src/lib/citation.test.ts`
Expected: FAIL — citation is non-empty without a seal.

- [ ] **Step 3: Implement**

In `citation.ts`, add at top of `auditCitation`:

```ts
const seal = report.provenance?.source_sha256?.trim()
if (!seal) return ''
```

Remove any later re-derivation of `seal`; reuse the variable for the “Rapport scellé” sentence.

- [ ] **Step 4: Run tests**

Run: `cd web/frontend && npx vitest run src/lib/citation.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/citation.ts web/frontend/src/lib/citation.test.ts
git commit -m "feat(web): single citation guard"
```

---

### Task 6: OCI image + compose + cgroups (S4)

**Files:**
- Create: `deploy/Dockerfile`
- Create: `deploy/docker-compose.yml`
- Modify: `docs/deploy.md` (S4 section)
- Test: manual `docker compose up` gate in step 4 (no CI dependency)

**Interfaces:**
- Consumes: `bin/cgo` (cross-built via `go build -o bin/cgo ./cmd/cgo`).
- Produces: digest-pinned image; compose wiring dashboard + Nginx + (eventually) exporters; `cpu.max`/`memory.max`.

- [ ] **Step 1: Create `deploy/Dockerfile`**

```dockerfile
FROM scratch
COPY bin/cgo /cgo
COPY bin/cgo /cgo
EXPOSE 9090
ENTRYPOINT ["/cgo", "--serve"]
```

- [ ] **Step 2: Create `deploy/docker-compose.yml`**

```yaml
services:
  cgo:
    image: ghcr.io/realms4239/cgo:latest
    ports:
      - "127.0.0.1:9090:9090"
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
    environment:
      - CGO_API_TOKEN=${CGO_API_TOKEN:?set}
    restart: unless-stopped
```

- [ ] **Step 3: Document S4 in `docs/deploy.md`**

```markdown
## S4 — OCI + rollback
`docker compose up` starts the digest-pinned image. Rollback is `compose pull` of the prior digest + `cgo shape --restore`. `cpu.max`/`memory.max` are set in compose; throttling stats are exported via `/metrics`.
```

- [ ] **Step 4: Verify**

Run: `GOOS=linux GOARCH=amd64 go build -o bin/cgo ./cmd/cgo`
Expected: builds without error. `docker compose config` succeeds.

- [ ] **Step 5: Commit**

```bash
git add deploy/Dockerfile deploy/docker-compose.yml docs/deploy.md
git commit -m "feat(deploy): OCI image + compose"
```

---

### Task 7: FQ-CoDel parameter pinning (S1 transport)

**Files:**
- Modify: `pkg/qdisc/qdisc.go`
- Test: `pkg/qdisc/qdisc_test.go` (create if missing)

**Interfaces:**
- Consumes: RFC 8290 §5.2 defaults (`target 5ms`, `interval 100ms`, `limit 10240`, `quantum 1514B`, `flows 1024`).
- Produces: explicit `tc` string carrying those five params; `limit` MUST be set (fail-closed if missing).

- [ ] **Step 1: Write the failing test**

```go
func TestFQCoDel_PinsRFC8290Params(t *testing.T) {
	s := fqCoDelString("eth0")
	for _, want := range []string{"target 5ms", "interval 100ms", "limit 10240", "quantum 1514", "flows 1024"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %q", want, s)
		}
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `go test ./pkg/qdisc -run TestFQCoDel_PinsRFC8290Params -v`
Expected: FAIL — params not in string.

- [ ] **Step 3: Implement**

In `pkg/qdisc/qdisc.go`, build the `fq_codel` leaf string with the five explicit params. If `limit` would be `0`, return an error rather than emitting `limit 0`.

- [ ] **Step 4: Run tests**

Run: `go test ./pkg/qdisc -run TestFQCoDel -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/qdisc/qdisc.go pkg/qdisc/qdisc_test.go
git commit -m "feat(qdisc): pin FQ-CoDel RFC 8290 params"
```

---

### Task 8: SLO recording rules + runbooks (S5)

**Files:**
- Create: `deploy/prometheus/rules.yml`
- Create: `docs/runbooks/smoke-alerts.md`
- Test: `promtool check rules deploy/prometheus/rules.yml` (manual)

**Interfaces:**
- Consumes: `/metrics` names from Task 3.
- Produces: recording rules for `cgo_cap_deadline_ok_pct`, alert rules for sustained breach / dead-probe / invalid-iface.

- [ ] **Step 1: Create rules (failing until validated)**

```yaml
groups:
  - name: cgo-slo
    rules:
      - record: cgo:slo:deadline_ok_pct_5m
        expr: avg(rate(cgo_cap_deadline_ok_pct[5m]))
      - alert: CGO_deadline_breach_sustained
        expr: cgo_cap_sla_breach == 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "CAP deadline breach sustained (operator objective)"
      - alert: CGO_probe_dead
        expr: rate(cgo_audit_load_measured_total[5m]) == 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "No measured load in 10m — probes may be dead"
      - alert: CGO_iface_invalid
        expr: cgo_iface_valid == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Interface sampling invalid — numerics must be ignored"
```

- [ ] **Step 2: Validate**

Run: `promtool check rules deploy/prometheus/rules.yml`
Expected: `SUCCESS`.

- [ ] **Step 3: Runbook**

Create `docs/runbooks/smoke-alerts.md` with sections: Auth rotation, Nginx revoke, Rollback (`cgo shape --restore` + digest pin), Prometheus target-down, Bench offload fail, Corpus seal verification (`python book/tools/audit-archive.py`).

- [ ] **Step 4: Commit**

```bash
git add deploy/prometheus/rules.yml docs/runbooks/smoke-alerts.md
git commit -m "feat(ops): SLO rules + runbooks"
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
git commit -m "docs: close enterprise hardening slice 28"
```

---

## Parallelization map

- Task 1 (Nginx) and Task 5 (frontend citation) are independent — can run in parallel.
- Task 2 (auth) and Task 3 (metrics) both touch `pkg/api/server.go` but different handlers — run sequentially.
- Task 4 needs Task 3.
- Task 6 (OCI) needs Task 2 + 3 (auth + metrics must be in the image).
- Task 7 (qdisc) is independent of 2–6.
- Task 8 needs Task 3/4 metric names.
- Task 9 must remain at the end.

## Risk register

| Risk | Mitigation |
|---|---|
| Auth breaks existing frontend flows | Bearer token via env; frontend sends `Authorization` header when token is configured; loopback dev mode preserved |
| `/metrics` becomes a NaN/zero leak | Honesty contract tests (Task 4); absent = no sample |
| Nginx `limit_req` throttles legitimate bulk import | Mutations `1r/s burst=3 nodelay`; calibrate after first load test |
| FQ-CoDel param change alters existing bench numbers | Pin explicitly, re-run `go test ./pkg/qdisc`, flag in checkpoint |
| OCI rollback leaves edge shaped | `cgo shape --restore` as post-deploy gate in S4 runbook |
| Live-run temptation | No live tests in this plan; human gate listed in spec §Open decisions |

## Definition of done

- No unauthenticated mutation succeeds; burst → `503`.
- `/metrics` scrapes clean with HELP/TYPE and honesty contract.
- `docker compose up` → green health; rollback by digest works.
- All required tests and gates pass; verifier signature or explicit `withheld` recorded.
- No new live run, dependency, CHR boot or frozen-corpus change without a signed human gate.
