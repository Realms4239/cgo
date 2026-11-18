# CGO / Meteolink Chunk 1B — Interface-Name Validation (A03) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reject any interface name that is not a strict Linux `IFNAMSIZ` token in both the remediation module and the `GET /api/remediate` handler, returning HTTP 400 at the API boundary.

**Architecture:** One shared rule, `^[a-zA-Z0-9_.-]{1,15}$`, is applied in two places because the packages are independent modules: `remediate.Link.Validate()` (defense in depth for every script generator) and an early guard in `handleRemediate` (fail before report lookup, so hostile input cannot reach the audit session).

**Tech Stack:** Go 1.25 `regexp`, `net/http/httptest`, separate `remediate` module (`remediate/go.mod`), `api` module.

**Source:** `docs/MUSTREAD/11-chunk-1-security-lockdown.md` §3 (verbatim corpus `CGO3.md` 40–45). Chunk 1 is decomposed; this is sub-plan B.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, `Dockerfile`, frontend, live bench, or any network target.
- The corpus's `pkg/remediate/remediate.go` path is stale: the real module is `remediate/remediate.go` (top-level). API handler is `api/server_audit.go:347` (`handleRemediate`).
- French comments/user text stay French; English plan/gate text.
- Hostile forms that MUST be rejected: `eth0;id`, whitespace, `|`, `&`, `$`, backticks, quotes, `/`, newlines, and names longer than 15 chars. Valid forms: `eth0`, `ifb0`, `ether1-wan`, `bridge`.
- One commit per sub-plan, exact message: `fix(security): reject invalid interface names at remediate and api boundary`.
- Extensive pre-existing WIP is protected; stage only the files named in this plan.

---

### Task 1: `remediate.Link.Validate()` rejects invalid interface names

**Files:**
- Modify: `remediate/remediate.go`
- Test: `remediate/remediate_test.go`

**Interfaces:**
- Consumes: existing `Link.Validate() error` at `remediate/remediate.go:85`, called by all four script generators (`:131`, `:219`, `:272`, `:334`).
- Produces: any non-empty `Interface`, `RouterOSInterface` or `RouterOSLanInterface` failing `^[a-zA-Z0-9_.-]{1,15}$` returns an error before a script can be emitted.

- [ ] **Step 1: Write the failing test**

Append to `remediate/remediate_test.go` (add `"regexp"` only if the test needs it — this test does not; use no new imports):

```go
func TestLinkValidateRejectsInvalidInterfaceNames(t *testing.T) {
	hostile := []string{
		"eth0;id", "eth0 ls", "eth0|id", "eth0&id", "eth0$id",
		"eth0`id`", "eth0'id", "eth0\"id", "eth0/id", "eth0\nid", "abcdefghijklmnop",
	}
	for _, name := range hostile {
		link := Link{Interface: name, CapacityMbps: 20, RTTms: 100}
		if err := link.Validate(); err == nil {
			t.Errorf("Validate accepted interface %q, want rejection", name)
		}
	}
	for _, name := range []string{"eth0", "ifb0", "ether1-wan", "bridge"} {
		link := Link{Interface: name, CapacityMbps: 20, RTTms: 100}
		if err := link.Validate(); err != nil {
			t.Errorf("Validate rejected valid interface %q: %v", name, err)
		}
	}
	ros := Link{RouterOSInterface: "ether1;ifb0", CapacityMbps: 20, RTTms: 100}
	if err := ros.Validate(); err == nil {
		t.Error("Validate accepted hostile RouterOSInterface")
	}
	lan := Link{RouterOSLanInterface: "../etc/passwd", CapacityMbps: 20, RTTms: 100}
	if err := lan.Validate(); err == nil {
		t.Error("Validate accepted hostile RouterOSLanInterface")
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd remediate && go test -count=1 -run TestLinkValidateRejectsInvalidInterfaceNames ./...`
Expected: FAIL — hostile names currently pass `Validate` (no name check exists).

- [ ] **Step 3: Implement the guard**

In `remediate/remediate.go`, add `"regexp"` to imports and insert before `Validate`:

```go
// ifaceRx — nom d'interface Linux strict (IFNAMSIZ) : ni espace, ni métacaractère
// shell, ni séparateur de chemin. Le script de remédiation est collé dans un
// routeur de production : un nom hostile ne doit jamais atteindre le générateur.
var ifaceRx = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,15}$`)

func validIface(name string) bool { return ifaceRx.MatchString(name) }
```

Keep the existing French comment above `Validate` unchanged, then insert this as the first check inside `Validate` (before the `Interface == ""` check):

```go
	for _, name := range []string{l.Interface, l.RouterOSInterface, l.RouterOSLanInterface} {
		if name != "" && !validIface(name) {
			return fmt.Errorf("nom d'interface invalide %q", name)
		}
	}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd remediate && go test -count=1 -run TestLinkValidateRejectsInvalidInterfaceNames ./...`
Expected: PASS.

- [ ] **Step 5: Full remediate suite**

Run: `cd remediate && go test -count=1 ./...`
Expected: all tests ok (the new guard must not break existing fixtures).

---

### Task 2: `GET /api/remediate` returns 400 before any report lookup

**Files:**
- Modify: `api/server_audit.go`
- Test: `api/server_audit_iface_test.go` (create)

**Interfaces:**
- Consumes: `handleRemediate(s *auditSession) http.HandlerFunc` at `api/server_audit.go:347`; `writeErr(w, r, message, status)` already used in this file.
- Produces: hostile `?iface=`, `?routeros_iface=`, `?routeros_lan=` answered 400 without touching `auditSession`; valid names proceed unchanged.

- [ ] **Step 1: Write the failing test**

Create `api/server_audit_iface_test.go` with:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRemediateRejectsInvalidIface(t *testing.T) {
	for _, q := range []string{
		"?iface=eth0%3Bid", "?iface=eth0%20ls", "?iface=eth0%7Cid",
		"?routeros_iface=ether1%3Bid", "?routeros_lan=..%2Fetc%2Fpasswd",
		"?iface=abcdefghijklmnop",
	} {
		r := httptest.NewRequest(http.MethodGet, "/api/remediate"+q, nil)
		w := httptest.NewRecorder()
		handleRemediate(nil)(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET /api/remediate%s = %d, want 400", q, w.Code)
		}
	}
}

func TestValidIfaceTokens(t *testing.T) {
	for _, name := range []string{"eth0", "ifb0", "ether1-wan", "bridge"} {
		if !validIface(name) {
			t.Errorf("validIface(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"eth0;id", "eth0 id", "../etc/passwd", "abcdefghijklmnop", ""} {
		if validIface(name) {
			t.Errorf("validIface(%q) = true, want false", name)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -count=1 -run 'TestHandleRemediateRejectsInvalidIface|TestValidIfaceTokens' ./api/`
Expected: FAIL — `validIface` undefined and hostile queries currently reach the session logic.

- [ ] **Step 3: Implement the early guard**

In `api/server_audit.go`, add `"regexp"` to imports and insert near the top of the file:

```go
var ifaceRx = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,15}$`)

func validIface(name string) bool { return ifaceRx.MatchString(name) }
```

At the very start of the closure returned by `handleRemediate` (before `q := r.URL.Query()` is used for anything), insert:

```go
		q := r.URL.Query()
		for _, key := range []string{"iface", "routeros_iface", "routeros_lan"} {
			if v := strings.TrimSpace(q.Get(key)); v != "" && !validIface(v) {
				writeErr(w, r, "nom d'interface invalide : "+v, http.StatusBadRequest)
				return
			}
		}
```

Then remove the later duplicate `q := r.URL.Query()` line if the edit leaves two declarations (keep exactly one).

- [ ] **Step 4: Run the targeted and full API tests**

Run: `go test -count=1 -run 'TestHandleRemediateRejectsInvalidIface|TestValidIfaceTokens' ./api/`
Expected: PASS.
Run: `go test -count=1 ./api/...`
Expected: all tests ok.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:**
- Consumes: Tasks 1–2.
- Produces: one commit with the exact message `fix(security): reject invalid interface names at remediate and api boundary`.

- [ ] **Step 1: Full module tests**

Run: `cd remediate && go test -count=1 ./...` → PASS.
Run: `go test -count=1 ./api/...` → PASS.

- [ ] **Step 2: Static proof**

Run: `git grep -n 'a-zA-Z0-9_.-' remediate/ api/`
Expected: the regex appears in both `remediate/remediate.go` and `api/server_audit.go`.

- [ ] **Step 3: Commit once with explicit paths**

Stage exactly `remediate/remediate.go`, `remediate/remediate_test.go`, `api/server_audit.go`, `api/server_audit_iface_test.go`. Commit with message `fix(security): reject invalid interface names at remediate and api boundary`. No other files. If a gate fails, STOP and report — do not add a permit.

- [ ] **Step 4: Record**

Write verbatim outputs and exit codes to the SDD report. Do not append to `docs/CHECKPOINTS.md`.
