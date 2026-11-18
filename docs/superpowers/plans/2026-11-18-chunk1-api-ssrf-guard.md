# CGO / Meteolink Chunk 1E — API SSRF Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `api/ssrf.go` delegate its address decisions to the single `probe` guard, so CGNAT, benchmark, documentation and v4-mapped IPv6 ranges are rejected at the API boundary with the same rule as at dial time.

**Architecture:** The `api` module already replaces `cgo/probe => ../probe`, so it can import the local `probe` package. `api/ssrf.go` keeps its own three error types (the handler maps them to 403/400) and its URL/address shape checks, but the repeated IP-range table and the second DNS resolution are removed: `checkHost` calls `probe.GuardHost` and translates `*probe.PrivateTargetError` into the API error type. The real connection is already dialed through `probe`'s guarded transport, so there is one rule and no double-resolution TOCTOU.

**Tech Stack:** Go 1.25 `net/netip`, `errors`, `net/url`, `net/http/httptest`; modules `api/` and `probe/` (local replace already present).

**Source:** `docs/MUSTREAD/11-chunk-1-security-lockdown.md` §2 (verbatim corpus `CGO3.md` 29–38). Chunk 1 is decomposed; this is the final sub-plan.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, `Dockerfile`, frontend views, live bench, remote hosts.
- Only `api/ssrf.go` and `api/ssrf_test.go` are in scope. `api/go.mod` already has `replace .../probe => ../probe` — do not edit it.
- Preserve the API error types (`privateTargetError`, `ssrfRefusedError`, `ssrfArgError`) and the handler's 403-for-private/refused, 400-for-argument shape. Do not weaken an existing assertion.
- The bench opt-in remains `CGO_ALLOW_PRIVATE_TARGETS=1`, read by `probe.AllowPrivateTargets()` at dial time. No new env switch.
- French comments/user text; English plan/gate text.
- One commit, exact message: `fix(api): delegate SSRF checks to the shared probe guard`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Delegate host checks to `probe.GuardHost`, with a real CGNAT oracle

**Files:**
- Modify: `api/ssrf.go`
- Test: `api/ssrf_test.go`

**Interfaces:**
- Consumes: `probe.GuardHost(ctx, host string, allowPrivate bool) error` and `*probe.PrivateTargetError{Host, IP, Range string}` from `probe/ssrf.go`.
- Produces: `checkHost(ctx, host, allowPrivate)` keeps its signature and returns the API error types; handler status mapping is unchanged.

- [ ] **Step 1: Write the failing test**

Append to `api/ssrf_test.go`:

```go
func TestCheckHostBlockedRangesMatchProbe(t *testing.T) {
	for _, host := range []string{
		"100.64.0.1", "169.254.169.254", "198.18.0.1", "::ffff:127.0.0.1", "2001:db8::1",
	} {
		err := checkHost(context.Background(), host, false)
		if err == nil {
			t.Errorf("checkHost(%q) = nil, want refusal", host)
			continue
		}
		var pe *privateTargetError
		var re *ssrfRefusedError
		if !errors.As(err, &pe) && !errors.As(err, &re) {
			t.Errorf("checkHost(%q) error %v, want API SSRF error type", host, err)
		}
	}
}
```

`errors` and `context` are already imported in this test file.

- [ ] **Step 2: Confirm it fails**

Run: `go -C api test -count=1 -run TestCheckHostBlockedRangesMatchProbe ./...`
Expected: FAIL — `checkHost("100.64.0.1")` currently returns nil because the API range table lacks CGNAT/benchmark/doc prefixes.

- [ ] **Step 3: Replace the duplicate range logic**

In `api/ssrf.go`:
- Add `"errors"` and `"github.com/Realms4239/cgo/probe"` imports.
- Delete `blockedIP` entirely (lines 46–77).
- Replace `checkHost` (lines 82–103) with:

```go
func checkHost(ctx context.Context, host string, allowPrivate bool) error {
	if host == "" {
		return &ssrfRefusedError{Why: "hôte vide — refusé"}
	}
	if err := probe.GuardHost(ctx, host, allowPrivate); err != nil {
		var blocked *probe.PrivateTargetError
		if errors.As(err, &blocked) {
			return &privateTargetError{IP: blocked.IP, Range: blocked.Range}
		}
		return &ssrfRefusedError{Why: err.Error()}
	}
	return nil
}
```

Keep `guardSmallURL`/`guardBulkAddr` shape checks and their `ssrfArgError` returns unchanged.

- [ ] **Step 4: Verify** — targeted test PASS; `go -C api test -count=1 ./...` ok.

---

### Task 2: Reject embedded userinfo in URLs

**Files:**
- Modify: `api/ssrf.go` (`guardSmallURL`)
- Test: `api/ssrf_test.go`

**Interfaces:**
- Consumes: `net/url` parsing already in `guardSmallURL`.
- Produces: `http://user:pass@host/` returns `ssrfArgError` → handler 400.

- [ ] **Step 1: Write the failing test**

```go
func TestGuardSmallURLRejectsUserinfo(t *testing.T) {
	if err := guardSmallURL(context.Background(), "http://user:pass@8.8.8.8/", false); err == nil {
		t.Fatal("URL with userinfo must be refused")
	}
	var ae *ssrfArgError
	if err := guardSmallURL(context.Background(), "http://user:pass@8.8.8.8/", false); !errors.As(err, &ae) {
		t.Fatalf("userinfo error = %v, want *ssrfArgError", err)
	}
}
```

- [ ] **Step 2: Confirm FAIL** — `go -C api test -count=1 -run TestGuardSmallURLRejectsUserinfo ./...` → FAIL (currently accepted).

- [ ] **Step 3: Implement**

In `guardSmallURL`, after parsing and before the scheme check:

```go
	if u.User != nil {
		return &ssrfArgError{Why: "small_url avec userinfo refusée"}
	}
```

- [ ] **Step 4: Verify** — targeted PASS; full api suite ok.

---

### Task 3: Remove the duplicate resolver and verify single-rule delegation

**Files:**
- Modify: `api/ssrf.go`
- Test: none new (static proof + existing suite)

**Interfaces:**
- Consumes: Tasks 1–2.
- Produces: no `net.DefaultResolver` / `LookupIP` call remains in `api/ssrf.go`; all address decisions come from `probe`.

- [ ] **Step 1: Static proof**

Run: `git grep -n 'LookupIP\|DefaultResolver' -- api/ssrf.go`
Expected: EMPTY.
Run: `git grep -n 'probe.GuardHost' -- api/ssrf.go`
Expected: one hit.

- [ ] **Step 2: Keep the API error contract**

Run: `go -C api test -count=1 -run 'TestAuditStartSSRFPrivateRefused|TestAuditStartSSRFArgsMalformed|TestCheckHostBlockedRangesMatchProbe|TestGuardSmallURLRejectsUserinfo' ./...`
Expected: PASS — private targets still 403, malformed arguments still 400.

- [ ] **Step 3: Full suites**

Run: `go -C api test -count=1 ./...` and `cd probe && go test -count=1 ./...` — both exit 0.

---

### Task 4: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–3.

- [ ] **Step 1: Commit once** — stage exactly `api/ssrf.go` and `api/ssrf_test.go` by explicit path; message `fix(api): delegate SSRF checks to the shared probe guard`. If a pre-commit gate fails, STOP and report.
- [ ] **Step 2: Record** — verbatim outputs and exit codes to the SDD report; no `docs/CHECKPOINTS.md` append.
