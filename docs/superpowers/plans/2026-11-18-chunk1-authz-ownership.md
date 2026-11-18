# CGO / Meteolink Chunk 1C — Owner AuthZ & Token Lifecycle (K2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every audit/run/report resource owner-scoped, give IDs unpredictable suffixes, require an explicit `--dev-no-auth` for unauthenticated loopback, replace the static session HMAC with a per-boot random key plus logout revocation, and require `X-Confirm-Action: true` on destructive POSTs.

**Architecture:** `api/auth.go` becomes the single identity seam: `Principal{OperatorID, Role}` is derived from the bearer token, stored as `owner_id` in sealed reports, and checked by a `requireOwner` wrapper on per-resource routes. `api/server.go` and `cmd/cgo/main.go` only generate IDs and attach `owner_id`; they do not reimplement auth logic. `cmd/cgo/main.go` owns the explicit `--dev-no-auth` flag and its loopback-only guard.

**Tech Stack:** Go 1.25 `crypto/rand`, `crypto/sha256`, `crypto/hmac`, `encoding/hex`, `net/http`, `httptest`. Modules: root (`api/`, `cmd/cgo/`), `audit/`.

**Source:** `docs/MUSTREAD/11-chunk-1-security-lockdown.md` §1 (verbatim corpus `CGO3.md` 15–28). Chunk 1 is decomposed; this is sub-plan C.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push, no auth-disable by default.
- Never touch `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, `Dockerfile`, frontend views, live bench, or remote hosts.
- Corpus paths are stale: real files are `api/auth.go`, `api/server.go`, `api/server_audit.go`, `cmd/cgo/main.go`, `audit/report.go`, `pkg/audit` → `audit/`. Keep `validRunID` traversal validation intact wherever it lives.
- Legacy seeded/frozen corpus reports with empty `owner_id` are readable by any authenticated principal but mutable only by role `admin`.
- `--dev-no-auth` must be loopback-only and print a loud warning at startup. Without it, an unconfigured `CGO_API_TOKEN` fails closed at startup with exit code 2.
- French comments/user text; English plan/gate text.
- One commit, exact message: `fix(security): scope resources to owners, randomize sessions and require confirmations`.
- Extensive pre-existing WIP is protected; stage only files named in this plan.

---

### Task 1: Collision-free IDs and explicit dev-no-auth

**Files:**
- Modify: `api/server.go:1601`, `cmd/cgo/main.go:237`, `cmd/cgo/main.go` flag block (`:65-93`)

**Interfaces:**
- Consumes: existing `fmt.Sprintf("audit-%d", time.Now().Unix())` at both sites and existing `fs.Bool` flag registration.
- Produces: IDs of form `audit-<unix>-<8hex>`; a `--dev-no-auth` boolean that disables auth only on loopback.

- [ ] **Step 1: Add a shared ID helper test**

Create `api/auditid_test.go`:

```go
package api

import (
	"regexp"
	"testing"
)

func TestNewAuditIDFormat(t *testing.T) {
	rx := regexp.MustCompile(`^audit-[0-9]{10}-[0-9a-f]{8}$`)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := NewAuditID()
		if !rx.MatchString(id) {
			t.Fatalf("NewAuditID() = %q, want audit-<unix>-<8hex>", id)
		}
		if seen[id] {
			t.Fatalf("duplicate ID %q", id)
		}
		seen[id] = true
	}
}
```

- [ ] **Step 2: Confirm it fails**

Run: `go test -count=1 -run TestNewAuditIDFormat ./api/`
Expected: FAIL with `undefined: NewAuditID`.

- [ ] **Step 3: Implement the helper and use it at both sites**

In `api/auth.go`, add imports `crypto/rand`, `encoding/hex`, `time`, and:

```go
// newAuditID — horodatage + 4 octets aléatoires : deux audits dans la même
// seconde ne collident pas et l'identifiant n'est pas devinable.
func NewAuditID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand indisponible: " + err.Error())
	}
	return fmt.Sprintf("audit-%d-%s", time.Now().Unix(), hex.EncodeToString(b[:]))
}
```

Update the test to call `NewAuditID()` instead of `newAuditID()`. Replace `fmt.Sprintf("audit-%d", time.Now().Unix())` at `api/server.go:1601` with `NewAuditID()`, and at `cmd/cgo/main.go:237` with `api.NewAuditID()` (`cmd/cgo` is a different package).

- [ ] **Step 4: Add the explicit dev flag**

In `cmd/cgo/main.go`, register `devNoAuth := fs.Bool("dev-no-auth", false, "mode dev: désactive l'authentification sur loopback uniquement (jamais avec --expose)")`. When `!*devNoAuth && os.Getenv("CGO_API_TOKEN") == "" && !tokenFilePresent(...)`, print a clear error to stderr and `os.Exit(2)` before serving. When `*devNoAuth`, print `ATTENTION: authentification désactivée (loopback uniquement)` at startup and refuse to combine with `--expose`. Pass the boolean into the API server constructor so `authenticated()` can consult it.

- [ ] **Step 5: Verify**

Run: `go test -count=1 -run TestNewAuditIDFormat ./api/` → PASS.
Run: `go build ./cmd/cgo` → exit 0.
Run: `go test -count=1 ./cmd/cgo/` → all tests ok.

---

### Task 2: Principal, owner_id persistence and ownership checks

**Files:**
- Modify: `api/auth.go`, `api/server.go`, `api/server_audit.go`, `audit/report.go`
- Test: `api/server_owner_test.go` (create)

**Interfaces:**
- Consumes: `Principal` defined in Task 2; reports from `audit.Report`.
- Produces: `func (s *Server) requireOwner(next http.HandlerFunc, ownerOf func(r *http.Request) string) http.HandlerFunc` returning 403 when `ownerID != "" && ownerID != principal.OperatorID && principal.Role != "admin"`.

- [ ] **Step 1: Define the identity and the rejection test**

In `api/auth.go`, add:

```go
type Principal struct {
	OperatorID string
	Role       string // "admin" ou "operator"
}

func principalFrom(r *http.Request) Principal {
	if !authenticated(r) {
		return Principal{}
	}
	token := os.Getenv("CGO_API_TOKEN")
	sum := sha256.Sum256([]byte(token))
	return Principal{OperatorID: hex.EncodeToString(sum[:])[:16], Role: "admin"}
}
```

Create `api/server_owner_test.go` with a direct unit oracle (no HTTP server needed):

```go
func TestAuthorizeOwner(t *testing.T) {
	admin := Principal{OperatorID: "aaaaaaaaaaaaaaaa", Role: "admin"}
	op := Principal{OperatorID: "bbbbbbbbbbbbbbbb", Role: "operator"}
	if !authorizeResource(admin, "cccccccccccccccc") {
		t.Error("admin must access any resource")
	}
	if !authorizeResource(op, "bbbbbbbbbbbbbbbb") {
		t.Error("owner must access own resource")
	}
	if authorizeResource(op, "cccccccccccccccc") {
		t.Error("cross-owner access must be refused")
	}
	if !authorizeResource(op, "") {
		t.Error("legacy unowned resource must be readable when authenticated")
	}
}
```

- [ ] **Step 2: Confirm it fails** — `go test -count=1 -run TestAuthorizeOwner ./api/` → FAIL (`undefined: authorizeResource`).

- [ ] **Step 3: Implement the check**

In `api/auth.go`, add:

```go
// authorizeResource — lecture/mutation d'une ressource possédée. Les rapports
// historiques sans owner_id restent lisibles ; leur mutation reste réservée à
// l'admin (contrôlée par l'appelant via requireOwnerMutate).
func authorizeResource(p Principal, ownerID string) bool {
	if p.Role == "admin" {
		return true
	}
	return ownerID == "" || ownerID == p.OperatorID
}
```

- [ ] **Step 4: Persist owner_id**

In `audit/report.go`, add `OwnerID string \`json:"owner_id"\`` to `Report`. In `api/server.go` and `cmd/cgo/main.go`, set `OwnerID: principalFrom(r).OperatorID` (CLI: derive from the configured token, or empty in explicit dev mode) on every newly created report. Add a test in `audit/` asserting the JSON key `owner_id` is present.

- [ ] **Step 5: Wrap per-resource routes**

In `api/server.go`, wrap every route that reads/mutates a specific audit/run/report by id (`/api/audit/*`, `/api/run/*`, `/api/report/*`) with a `requireOwner` wrapper that loads the report's `owner_id` and returns 403 via `writeErr(..., http.StatusForbidden)` when `authorizeResource` is false. Mutation routes (delete/export/stop) additionally refuse `ownerID == ""` for non-admin. Keep all existing `requireAuth` wrappers intact.

- [ ] **Step 6: Verify**

Run: `go test -count=1 -run TestAuthorizeOwner ./api/` → PASS.
Run: `go test -count=1 ./api/... ./cmd/cgo/... ./audit/...` → all ok.

---

### Task 3: Random per-boot session key and logout revocation

**Files:**
- Modify: `api/auth.go`, `api/server.go`
- Test: `api/auth_test.go` (modify)

**Interfaces:**
- Consumes: existing `sessionToken(secret)`, `validSession(cookieVal, secret)`, `Set-Cookie` at `api/server.go:450`.
- Produces: per-boot random key, `revokeSession(token)` invalidating sessions, `POST /api/auth/logout`.

- [ ] **Step 1: Replace deterministic token tests**

In `api/auth_test.go`, replace the determinism assertions (`TestSessionToken...` at lines 29–55) with:

```go
func TestSessionKeyIsRandomAndRevocable(t *testing.T) {
	a := newSessionToken()
	b := newSessionToken()
	if a == b {
		t.Fatal("two sessions must differ")
	}
	if !validSession(a) {
		t.Fatal("fresh session must validate")
	}
	revokeSession(a)
	if validSession(a) {
		t.Fatal("revoked session must not validate")
	}
	if !validSession(b) {
		t.Fatal("revoking one session must not revoke another")
	}
}
```

- [ ] **Step 2: Confirm it fails** — `go test -count=1 -run TestSessionKeyIsRandomAndRevocable ./api/` → FAIL (undefined symbols).

- [ ] **Step 3: Implement**

In `api/auth.go`, replace `sessionMsg`/`sessionToken` with a package-level `var bootKey [32]byte` initialised in `init()` by `rand.Read`, plus:

```go
func newSessionToken() string {
	var nonce [16]byte
	_, _ = rand.Read(nonce[:])
	mac := hmac.New(sha256.New, bootKey[:])
	_, _ = mac.Write(nonce[:])
	return hex.EncodeToString(nonce[:]) + "." + hex.EncodeToString(mac.Sum(nil))
}

var revokedJTI = map[string]bool{}

func revokeSession(token string) { revokedJTI[token] = true }

func validSession(token string) bool {
	jti, sig, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	mac := hmac.New(sha256.New, bootKey[:])
	_, _ = mac.Write([]byte(jti))
	return subtle.ConstantTimeCompare([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) == 1 && !revokedJTI[token]
}
```

Guard `revokedJTI` with a `sync.Mutex` (concurrent HTTP handlers). Update `authenticated()` and `api/server.go:450` to the no-argument forms. Register `POST /api/auth/logout` behind `requireAuth`, reading the cookie and calling `revokeSession`.

- [ ] **Step 4: Verify** — targeted test PASS; full `go test -count=1 ./api/...` ok.

---

### Task 4: Require X-Confirm-Action on destructive POSTs

**Files:**
- Modify: `api/server.go` (routes at `:532`, `:559`, `:683`, `:705`, `:1240`)
- Test: `api/server_confirm_test.go` (create)

**Interfaces:**
- Consumes: `writeErr(w, r, msg, status)`.
- Produces: `requireConfirm(next http.HandlerFunc) http.HandlerFunc` returning 400 when the header is absent.

- [ ] **Step 1: Write the failing test**

```go
func TestRequireConfirm(t *testing.T) {
	called := false
	h := requireConfirm(func(w http.ResponseWriter, r *http.Request) { called = true })
	r := httptest.NewRequest(http.MethodPost, "/api/shape", nil)
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusBadRequest || called {
		t.Fatalf("missing X-Confirm-Action: code=%d called=%v, want 400 and no call", w.Code, called)
	}
	r2 := httptest.NewRequest(http.MethodPost, "/api/shape", nil)
	r2.Header.Set("X-Confirm-Action", "true")
	w2 := httptest.NewRecorder()
	h(w2, r2)
	if !called {
		t.Fatal("confirmed request must reach the handler")
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go test -count=1 -run TestRequireConfirm ./api/` → FAIL (`undefined: requireConfirm`).

- [ ] **Step 3: Implement and wrap**

```go
func requireConfirm(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Confirm-Action") != "true" {
			writeErr(w, r, "confirmation requise : X-Confirm-Action: true", http.StatusBadRequest)
			return
		}
		next(w, r)
	}
}
```

Wrap `POST /api/shape`, `/api/burst`, `/api/run/start`, `/api/profile/delete`, `/api/cost/tier` with `requireConfirm` inside the existing `requireAuth` wrapper. Do not wrap read routes.

- [ ] **Step 4: Verify** — targeted PASS; full api suite ok.

---

### Task 5: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–4.

- [ ] **Step 1: Full suites**

Run: `go test -count=1 ./api/... ./cmd/cgo/... ./audit/...` → all ok.
Run: `go build ./cmd/cgo` → exit 0.

- [ ] **Step 2: Static proofs**

Run: `git grep -n 'dev-no-auth' cmd/cgo/` → at least one hit.
Run: `git grep -n 'X-Confirm-Action' api/` → at least one hit.
Run: `git grep -n 'cgo-session-v1' api/` → expected EMPTY.
Run: `git grep -n 'audit-%d' api/ cmd/cgo/` → expected EMPTY.

- [ ] **Step 3: Commit once**

Stage only the named files by explicit path; commit message `fix(security): scope resources to owners, randomize sessions and require confirmations`. If a pre-commit gate fails, STOP and report.

- [ ] **Step 4: Record**

Write verbatim outputs and exit codes to the SDD report. Do not append to `docs/CHECKPOINTS.md`.
