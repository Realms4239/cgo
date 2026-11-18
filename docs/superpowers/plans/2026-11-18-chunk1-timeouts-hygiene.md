# CGO / Meteolink Chunk 1D — Timeout & Write Hygiene Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bound every unbounded server and probe path, make `writeJSON` fail closed with HTTP 500 on an encode error, and align the audit-duration default to 60 s (10/40/10).

**Architecture:** `api/server.go` owns response encoding and per-route contexts; `cmd/cgo/server.go` owns `http.Server` timeouts; `api/sse.go` clears the write deadline only on long-lived streams; leaf modules (`probe`, `campagne`, `cmd/cgo/expose.go`) bound their own dials and execs. No new abstraction — each layer uses stdlib primitives already present.

**Tech Stack:** Go 1.25 `bytes.Buffer`, `http.ResponseController`, `context.WithTimeout`, `exec.CommandContext`, `net.Resolver`, `httptest`. Modules: root (`api/`, `cmd/cgo/`), `probe/`, `campagne/` is a package under root (`pkg/campagne/`).

**Source:** `docs/MUSTREAD/11-chunk-1-security-lockdown.md` §4 (verbatim corpus `CGO3.md` 45–57). Chunk 1 is decomposed; this is sub-plan D.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, `Dockerfile`, frontend views, live bench, or remote hosts.
- Stale corpus paths: real files are `probe/probe.go`, `pkg/campagne/modprobe_linux.go`, `api/server.go`, `api/sse.go`, `cmd/cgo/server.go`, `cmd/cgo/expose.go`.
- SSE must keep working past 30 s: `WriteTimeout = 30 * time.Second` on the server, with `http.NewResponseController(w).SetWriteDeadline(time.Time{})` on `/api/stream` and `/api/replay/stream` only.
- 60 s default audit = 10 s idle / 40 s load / 10 s drain. `api/server.go:128` currently `Default: 30` → `60`; the CLI already uses `num("duration", 60)`.
- French comments/user text; English plan/gate text.
- One commit, exact message: `fix(server): bound contexts, fail writeJSON closed and unify 60s default`.
- Extensive pre-existing WIP is protected; stage only files named in this plan.

---

### Task 1: `writeJSON` buffers and fails closed with 500

**Files:**
- Modify: `api/server.go` (`writeJSON` at `:1599-1605`)
- Test: `api/server_writejson_test.go` (create)

**Interfaces:**
- Consumes: existing `writeJSON(w http.ResponseWriter, v any)` — signature MUST stay unchanged so no caller is touched.
- Produces: on encode error, HTTP 500 with no partial body; on success, identical JSON as before.

- [ ] **Step 1: Write the failing test**

Create `api/server_writejson_test.go`:

```go
package api

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONEncodesBeforeWriting(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]any{"ok": true})
	if w.Code != http.StatusOK {
		t.Fatalf("valid value: code = %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("valid value produced no JSON body")
	}
}

func TestWriteJSONFailsClosedOnEncodeError(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]any{"bad": math.NaN()})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unencodable value: code = %d, want 500", w.Code)
	}
}
```

- [ ] **Step 2: Confirm it fails**

Run: `go -C api test -count=1 -run 'TestWriteJSONEncodesBeforeWriting|TestWriteJSONFailsClosedOnEncodeError' ./...`
Expected: FAIL — unencodable value currently yields 200 with a partial/empty body.

- [ ] **Step 3: Implement buffering**

In `api/server.go`, add `"bytes"` to imports and replace `writeJSON` with:

```go
func writeJSON(w http.ResponseWriter, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
		http.Error(w, `{"error":"échec d'encodage JSON"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}
```

- [ ] **Step 4: Verify** — targeted tests PASS; `go -C api test -count=1 ./...` ok.

---

### Task 2: `WriteTimeout` 30 s with SSE deadline exemption

**Files:**
- Modify: `cmd/cgo/server.go:346-355`, `api/sse.go` (`Hub.SSE` at `:134`), `api/server.go` (replay handler at `:1147`)
- Test: `api/sse_deadline_test.go` (create)

**Interfaces:**
- Consumes: `Hub.SSE(w, r)` and the replay handler; `http.NewResponseController(w)`.
- Produces: server applies a 30 s write timeout to non-stream responses; both stream handlers clear their own deadline.

- [ ] **Step 1: Write the failing test**

Create `api/sse_deadline_test.go`:

```go
package api

import (
	"net/http/httptest"
	"testing"
)

func TestSSEClearsWriteDeadline(t *testing.T) {
	h := NewHub()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/stream", nil)
	c := &deadlineSpyResponseWriter{ResponseWriter: w}
	h.SSE(c, r)
	if !c.deadlineCleared {
		t.Error("Hub.SSE must clear the write deadline for long-lived streams")
	}
}
```

Add the spy in the same file (imports `net/http` and `time`):

```go
type deadlineSpyResponseWriter struct {
	http.ResponseWriter
	deadlineCleared bool
}

func (s *deadlineSpyResponseWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func (s *deadlineSpyResponseWriter) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		s.deadlineCleared = true
	}
	return nil
}

func (s *deadlineSpyResponseWriter) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestSSEClearsWriteDeadline ./...` → FAIL.

- [ ] **Step 3: Implement**

In `cmd/cgo/server.go`, change `WriteTimeout: 0` to `WriteTimeout: 30 * time.Second` and update the comment to state that only `/api/stream` and `/api/replay/stream` clear the deadline.

In `api/sse.go`, at the start of `func (h *Hub) SSE(...)`, before the first write:

```go
	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil {
		// Certains ResponseWriter (tests, proxies) ne supportent pas la deadline :
		// on continue sans échouer le flux pour autant.
		_ = err
	}
```

In `api/server.go`, apply the same call at the top of the `/api/replay/stream` handler closure (after the `run` query validation).

- [ ] **Step 4: Verify** — targeted PASS; `go -C api test -count=1 ./...` ok; `go build ./cmd/cgo` exit 0.

---

### Task 3: Bound every remaining unbounded call

**Files:**
- Modify: `api/server.go`, `pkg/campagne/modprobe_linux.go`, `cmd/cgo/expose.go`, `probe/probe.go`
- Test: `probe/smallclient_timeout_test.go` (create)

**Interfaces:**
- Consumes: existing handlers `/api/report/export` (`:1011`), `/api/figures/regen` (`:1122`), replay stream (`:1147`).
- Produces: explicit deadlines everywhere; no call can hang for the process lifetime.

- [ ] **Step 1: Write the probe timeout oracle**

Create `probe/smallclient_timeout_test.go`:

```go
package probe

import (
	"testing"
	"time"
)

func TestSmallClientTimeoutUsesCalibratedFloor(t *testing.T) {
	c := SmallClientTimeout(4 * 600 * time.Millisecond)
	if c.Timeout != 2400*time.Millisecond {
		t.Fatalf("P2 VSAT client timeout = %v, want 2.4s", c.Timeout)
	}
	if got := smallClientFloor(10 * time.Millisecond); got != 8*time.Second {
		t.Fatalf("floor(10ms) = %v, want 8s", got)
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `cd probe && go test -count=1 -run TestSmallClientTimeoutUsesCalibratedFloor ./...` → FAIL (`undefined: smallClientFloor` if helper absent, or wrong timeout if the floor is missing).

- [ ] **Step 3: Implement the floor**

In `probe/probe.go`, add:

```go
// smallClientFloor — VSAT : jamais moins de 8 s, et au moins 4× le délai
// du profil. Un client 2 s sur un lien 600 ms échoue G2 par construction.
func smallClientFloor(delay time.Duration) time.Duration {
	floor := 4 * delay
	if floor < 8*time.Second {
		floor = 8 * time.Second
	}
	return floor
}
```

Keep `SmallClientTimeout(timeout)` returning `&http.Client{Timeout: timeout, Transport: sharedMarkedTransport()}`; callers pass `smallClientFloor(profileDelay)`.

- [ ] **Step 4: Bound server calls**

In `api/server.go`:
- `/api/report/export`: after `requireOwnerMutate`, create `ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second); defer cancel()` and use `r.WithContext(ctx)` for the handler body.
- `/api/figures/regen`: same with 60 s.
- `/api/replay/stream`: `ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute); defer cancel()`; use `r.WithContext(ctx)`.

- [ ] **Step 5: Bound exec and DNS**

In `pkg/campagne/modprobe_linux.go`, change `exec.Command(...).Run()` to:

```go
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = exec.CommandContext(ctx, "sudo", "-n", "/usr/local/bin/cgo-netctl", "modprobe", m).Run()
		cancel()
```

In `cmd/cgo/expose.go:47`, replace `net.LookupIP(h)` with a `net.Resolver.LookupIPAddr` under `context.WithTimeout(ctx, 5*time.Second)` (keep the same returned error shape).

- [ ] **Step 6: Verify** — `cd probe && go test -count=1 ./...` ok; `go -C api test -count=1 ./...` ok; `go test -count=1 ./pkg/campagne/... ./cmd/cgo/...` ok.

---

### Task 4: 60 s audit default and API docs

**Files:**
- Modify: `api/server.go:128`, `docs/api.md`
- Test: `api/duration_default_test.go` (create)

**Interfaces:**
- Consumes: the settings table entry `audit_duration_s`.
- Produces: default 60 s in the API settings; doc states 4 KiB login cap / 1 MiB elsewhere.

- [ ] **Step 1: Write the failing test**

```go
package api

import "testing"

func TestAuditDurationDefaultIs60(t *testing.T) {
	for _, s := range schemaParams {
		if s.Key == "audit_duration_s" {
			if s.Default != 60 {
				t.Fatalf("audit_duration_s default = %v, want 60", s.Default)
			}
			return
		}
	}
	t.Fatal("audit_duration_s setting missing")
}
```

The table is `api/server.go:124` `var schemaParams = []Param{...}` in package `api`, so the test reaches it directly.

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestAuditDurationDefaultIs60 ./...` → FAIL (default 30).

- [ ] **Step 3: Implement** — change `Default: 30` to `Default: 60` on `audit_duration_s` in `api/server.go:128`. In `docs/api.md`, add to the existing header note: `Le corps de /api/auth/login est limité à 4 KiB ; les autres corps JSON à 1 MiB.` Keep existing French text intact.

- [ ] **Step 4: Verify** — targeted PASS; full api suite ok.

---

### Task 5: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–4.

- [ ] **Step 1: Full suites** — `go -C api test -count=1 ./...`; `cd probe && go test -count=1 ./...`; `go test -count=1 ./pkg/campagne/... ./cmd/cgo/...`; `go build ./cmd/cgo` — all exit 0.
- [ ] **Step 2: Static proofs** — `git grep -n 'WriteTimeout: 30 \* time.Second' cmd/cgo/`; `git grep -n 'SetWriteDeadline' api/`; `git grep -n 'Default: 60' api/server.go`; `git grep -n 'CommandContext' pkg/campagne/modprobe_linux.go` — each must have a hit.
- [ ] **Step 3: Commit once** — stage only the files named in Tasks 1–4 by explicit path; message `fix(server): bound contexts, fail writeJSON closed and unify 60s default`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs and exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
