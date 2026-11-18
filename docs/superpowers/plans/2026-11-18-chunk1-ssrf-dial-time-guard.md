# CGO / Meteolink Chunk 1A — Dial-Time SSRF Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every probe dial — HTTP small-object and raw TCP bulk — reject reserved/non-public addresses at connect time, with the full IPv4/IPv6 range set from the MUSTREAD corpus.

**Architecture:** The `probe` module is the single guard owner. `IsBlockedAddr(netip.Addr)` is the shared predicate; `guardedDialContext` resolves once and validates the exact IP again inside `net.Dialer.Control`, closing the double-DNS TOCTOU window. `DialWithCC` (bulk TCP) and `NewSafeTransport` (HTTP) both route through it.

**Tech Stack:** Go 1.25 `netip`, `net.Dialer.Control`, `syscall.RawConn`, separate `probe` module (`probe/go.mod`), `testing`.

**Source:** `docs/MUSTREAD/11-chunk-1-security-lockdown.md` §2 (verbatim corpus `CGO3.md` 29–56). Chunk 1 is decomposed into four self-contained plans; this is the first.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator-level bans still bind. Never `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- No `book/`, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI workflows, `Dockerfile`, frontend views, or live bench touched.
- The `api` module binds the *published* probe module, not this tree: this plan changes only the local `probe/` module. The API-side duplicate (`api/ssrf.go`) is a separate follow-up plan.
- Keep French comments and user-facing text in French; English plan/gate text.
- Keep existing `//go:build` separation (`bulk_cc_linux.go` vs `bulk_cc_other.go`).
- The corpus's single "one commit per Chunk" rule applies per decomposed plan: this plan lands exactly one commit, message `fix(probe): block reserved ranges and enforce dial-time ssrf guard`.
- Chunk 0 check 1 is still BLOCKED on the human Cap-40 purge; this plan proceeds on explicit human instruction.

---

### Task 1: Full reserved-range predicate with failing tests

**Files:**
- Modify: `probe/ssrf.go`
- Test: `probe/ssrf_test.go`

**Interfaces:**
- Consumes: existing `blockedIP(ip net.IP) (string, bool)` used by `GuardHost`, `GuardURL`, `guardedDialContext`.
- Produces: `func IsBlockedAddr(addr netip.Addr) bool` — the shared predicate used by Task 2 and by `blockedIP`.

- [ ] **Step 1: Write the failing test**

Append this exact test to `probe/ssrf_test.go` (add `"net/netip"` to its imports):

```go
func TestIsBlockedAddrCoversCorpusRanges(t *testing.T) {
	blocked := []string{
		"0.0.0.1", "10.0.0.1", "100.64.0.1", "127.0.0.1", "169.254.169.254",
		"172.16.0.1", "192.0.0.1", "192.0.2.1", "192.168.1.1", "198.18.0.1",
		"198.51.100.1", "203.0.113.1", "224.0.0.1", "240.0.0.1", "255.255.255.255",
		"::", "::1", "100::1", "2001:db8::1", "fc00::1", "fe80::1", "fec0::1", "ff00::1",
		"::ffff:127.0.0.1", "::ffff:100.64.0.1",
	}
	for _, s := range blocked {
		a, err := netip.ParseAddr(s)
		if err != nil {
			t.Fatalf("ParseAddr(%q): %v", s, err)
		}
		if !IsBlockedAddr(a) {
			t.Errorf("IsBlockedAddr(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"8.8.8.8", "1.1.1.1", "2001:4860:4860::8888"} {
		a := netip.MustParseAddr(s)
		if IsBlockedAddr(a) {
			t.Errorf("IsBlockedAddr(%q) = true, want false", s)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd probe && go test -count=1 -run TestIsBlockedAddrCoversCorpusRanges ./...`
Expected: FAIL with `undefined: IsBlockedAddr`.

- [ ] **Step 3: Implement the predicate**

In `probe/ssrf.go`, add `"net/netip"` to imports and insert:

```go
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("255.255.255.255/32"),
	netip.MustParsePrefix("::/128"), netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("fec0::/10"), netip.MustParsePrefix("ff00::/8"),
}

// IsBlockedAddr — garde unique : Unmap() d'abord, donc ::ffff:127.0.0.1 est
// bien le loopback v4. Aucune adresse non publique ne sort de ce predicate.
func IsBlockedAddr(addr netip.Addr) bool {
	a := addr.Unmap()
	for _, p := range blockedPrefixes {
		if p.Contains(a) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Delegate the existing predicate to it**

Replace the entire body of `blockedIP` (lines 43–74) with:

```go
func blockedIP(ip net.IP) (string, bool) {
	if ip == nil {
		return "adresse absente", true
	}
	a, ok := netip.AddrFromSlice(ip)
	if !ok {
		return "forme non reconnue", true
	}
	if IsBlockedAddr(a) {
		return "plage réservée ou non publique", true
	}
	return "", false
}
```

Keep `PrivateTargetError` and every caller unchanged.

- [ ] **Step 5: Run the test to verify it passes**

Run: `cd probe && go test -count=1 -run TestIsBlockedAddrCoversCorpusRanges ./...`
Expected: PASS.

---

### Task 2: Dial-time validation and raw TCP bulk wiring

**Files:**
- Modify: `probe/ssrf.go`, `probe/bulk_cc_linux.go`
- Test: `probe/ssrf_test.go`

**Interfaces:**
- Consumes: `IsBlockedAddr(netip.Addr)` from Task 1; existing `AllowPrivateTargets()` in `probe/ssrf.go`.
- Produces: `guardedDialContext` validates inside `Control` at connect time; `DialWithCC` routes through it preserving TCP congestion control.

- [ ] **Step 1: Write the failing test**

Append to `probe/ssrf_test.go` (add `"errors"` to its imports):

```go
func TestGuardedDialRejectsReservedAtConnectTime(t *testing.T) {
	for _, addr := range []string{
		"100.64.0.1:5201", "169.254.169.254:80", "[2001:db8::1]:80", "[::ffff:127.0.0.1]:80",
	} {
		c, err := guardedDialContext(context.Background(), "tcp", addr, false, nil)
		if err == nil {
			c.Close()
			t.Errorf("guardedDialContext(%q) returned a connection, want refusal", addr)
			continue
		}
		var refused *PrivateTargetError
		if !errors.As(err, &refused) {
			t.Errorf("guardedDialContext(%q) error %v, want *PrivateTargetError", addr, err)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd probe && go test -count=1 -run TestGuardedDialRejectsReservedAtConnectTime ./...`
Expected: FAIL — the current guard's range list does not reject CGNAT/benchmark/doc ranges, so a real network error is returned instead of `*PrivateTargetError`.

- [ ] **Step 3: Add the connect-time control**

In `probe/ssrf.go`, insert after `IsBlockedAddr`:

```go
func dialControl(allowPrivate bool, control func(network, address string, c syscall.RawConn) error) func(string, string, syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		ap, err := netip.ParseAddrPort(address)
		if err != nil {
			return fmt.Errorf("adresse de connexion %q illisible — refusée : %w", address, err)
		}
		if !allowPrivate && IsBlockedAddr(ap.Addr()) {
			return &PrivateTargetError{Host: ap.Addr().String(), IP: ap.Addr().String(), Range: "plage réservée/non publique"}
		}
		if control != nil {
			return control(network, address, c)
		}
		return nil
	}
}
```

- [ ] **Step 4: Use it in `guardedDialContext`**

Replace line 144 of `probe/ssrf.go` (`d := &net.Dialer{Timeout: 5 * time.Second, Control: control}`) with:

```go
	d := &net.Dialer{Timeout: 5 * time.Second, Control: dialControl(allowPrivate, control)}
```

The resolution loop above it stays: resolve once, validate each IP, then dial those IPs. The `Control` re-validates the exact IP at socket connect time — that is the TOCTOU fix.

- [ ] **Step 5: Wire bulk TCP through the same guard**

In `probe/bulk_cc_linux.go:14-27`, replace the body of `DialWithCC` with:

```go
func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error) {
	var control func(network, address string, c syscall.RawConn) error
	if cc != "" {
		control = func(_, _ string, c syscall.RawConn) error {
			var serr error
			if err := c.Control(func(fd uintptr) {
				serr = syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, tcpCongestion, socketToken(cc))
			}); err != nil {
				return err
			}
			return serr
		}
	}
	return guardedDialContext(ctx, "tcp", addr, AllowPrivateTargets(), control)
}
```

Do not touch `bulk_cc_other.go` (non-Linux build stays as-is).

- [ ] **Step 6: Run both guards' tests**

Run: `cd probe && go test -count=1 -run 'TestIsBlockedAddrCoversCorpusRanges|TestGuardedDialRejectsReservedAtConnectTime' ./...`
Expected: PASS.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:**
- Consumes: Tasks 1–2.
- Produces: one commit with the exact message `fix(probe): block reserved ranges and enforce dial-time ssrf guard`; verbatim gate output.

- [ ] **Step 1: Full probe module tests**

Run: `cd probe && go test -count=1 ./...`
Expected: all packages ok, exit 0.

- [ ] **Step 2: Static proof of the single dial path**

Run: `git grep -n "100.64.0.0/10" probe/`
Expected: at least one hit in `probe/ssrf.go`.
Run: `git grep -n "guardedDialContext\|dialControl\|Control:" probe/bulk_cc_linux.go probe/ssrf.go`
Expected: `bulk_cc_linux.go` calls `guardedDialContext`; `ssrf.go` defines `dialControl` and uses it in the dialer.

- [ ] **Step 3: Ensure no bypass was introduced**

Run: `git grep -n "allowPrivate.*true\|CGO_ALLOW_PRIVATE_TARGETS" probe/`
Expected: only the existing `AllowPrivateTargets()` opt-in path; no hardcoded `true`.

- [ ] **Step 4: Commit once with explicit paths**

Stage `probe/ssrf.go` and `probe/ssrf_test.go` and `probe/bulk_cc_linux.go` by explicit path. Commit with message `fix(probe): block reserved ranges and enforce dial-time ssrf guard`. No other files.

- [ ] **Step 5: Record the result**

Write the verbatim outputs and exit codes to the SDD workspace report. Do not append to `docs/CHECKPOINTS.md` (near ceiling).
