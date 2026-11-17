# Chunk 1 — Security Lockdown

> Source: `C:\END\CGO3.md` lines 1–91. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** owner-scoped AuthZ, dial-time SSRF guard, `?iface=` regex validation, bounded timeouts and a 60 s audit default.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** closes caps K2, K7 and A03 before any measurement work.
- **Usage:** execute after Chunk 0; gates Chunk 2.
- **Inherited constraints:** `01-core-directive.md` §1–4.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 2 — data honesty and integrity.
Proceeding to Chunk 1 (Security Lockdown, Owner AuthZ, SSRF, ?iface= Injection,
Timeouts & 60s Default).

Hand Part A directly to your agent, then run Part B yourself to verify every
security invariant holds before we touch Chunk 2.

CHUNK 1 — AuthZ Ownership (K2), Unified SSRF (K7), Injection (A03) & Timeouts

PART A — Exact Bounded Prompt for Your Coding Agent

# CHUNK 1 EXECUTION TASK — SECURITY LOCKDOWN & CONTRACT HYGIENE (STRICTLY BOUNDED)

You are executing ONLY Chunk 1 on `C:\cgo`. Do NOT modify `pkg/campagne/prod.go`, `pkg/qdisc/stats.go`, `api/corpus.go`, `dualns.go`, `Dockerfile`, CI workflows, or frontend views (except small API header helpers if needed by tests).

## 1. Fix `K2` — Owner-Scoped AuthZ, Non-Colliding IDs, Token Lifecycle & Confirmations
Files: `api/auth.go`, `api/server.go`, `api/server_audit.go`, `cmd/cgo/main.go`, `cmd/cgo/expose.go`, `pkg/audit/`, `pkg/campagne/`
1. **Collision-Free, Unpredictable IDs (`cmd/cgo/main.go:237`, `api/server.go:1601`):**
   - Replace `audit-<unix>` with `audit-<unix>-<8hex>` using 4 random bytes from `crypto/rand` (`encoding/hex`). Keep existing `validRunID` traversal validation intact.
2. **Operator Identity & Object Ownership (`api/auth.go`, `api/server.go:419-426`):**
   - Introduce `Principal{OperatorID string, Role string}` (`"admin"` or `"operator"`). Derive a stable deterministic `OperatorID` from the token (`sha256(token)[:16]`), or support configured named tokens (`CGO_API_TOKENS` / `CGO_API_TOKEN` defaulting to role `"admin"` for single-operator CLI setups and `"operator"` for scoped tokens).
   - Persist `owner_id` inside every newly created audit report (`manifest.json` / `report.json`) and campaign manifest. (Legacy seeded/frozen corpus runs with empty `owner_id` are readable by authenticated users, mutable only by `"admin"`.)
   - Enforce ownership in `api/server.go:419-426` on ALL `/api/audit/*` and `/api/run/*` / `/api/report/*` per-resource reads, mutations, exports, and deletes: allow ONLY if `owner_id == principal.OperatorID || principal.Role == "admin"`; otherwise return HTTP `403`.
3. **Eliminate Silent Loopback Bypass & Static Cookie (`api/auth.go:16-26, 45-59`):**
   - Remove implicit unauthenticated loopback bypass when `CGO_API_TOKEN` is unset. Require explicit `--dev-no-auth` CLI flag (loopback-only, loud warning at startup) or fail closed with exit code `2` if no token is configured.
   - Replace static deterministic `HMAC(token, "cgo-session-v1")` cookie signing with a cryptographically random 32-byte per-boot signing key (`crypto/rand`) plus an in-memory `revokedJTI` map invalidated by `POST /api/auth/logout`.
4. **Enforce Two-Step Confirmations (`api/server.go:658, 683, 705`, `PRODUCT.md:46,97`):**
   - Require `X-Confirm-Action: true` (or `confirm: true` in JSON payload) on destructive/network-mutating POST routes (`shape`, `burst`, `run/start`, `profile/delete`, `cost/tier`), returning `400` with clear message if absent.

## 2. Fix `K7` — Unified Dial-Time SSRF Guard (HTTP + Raw TCP, Zero Double-DNS TOCTOU)
Files: `pkg/probe/ssrf.go`, `api/ssrf.go`, `pkg/probe/probe.go`, `pkg/probe/bulk_cc_linux.go`
1. Create/update **one single shared SSRF dialer** in `pkg/probe/ssrf.go`:
   - Enforce `http`/`https` schemes (no embedded userinfo) for URLs; valid `host:port` for raw bulk TCP addresses.
   - Implement `IsBlockedAddr(addr netip.Addr) bool` calling `addr.Unmap()` first and rejecting ALL reserved/non-public ranges:
     - IPv4: `0.0.0.0/8`, `10.0.0.0/8`, `100.64.0.0/10` (CGNAT), `127.0.0.0/8`, `169.254.0.0/16` (link-local/metadata), `172.16.0.0/12`, `192.0.0.0/24`, `192.0.2.0/24`, `192.168.0.0/16`, `198.18.0.0/15`, `198.51.100.0/24`, `203.0.113.0/24`, `224.0.0.0/4`, `240.0.0.0/4`, `255.255.255.255/32`.
     - IPv6: `::/128`, `::1/128`, `100::/64`, `2001:db8::/32`, `fc00::/7`, `fe80::/10`, `fec0::/10`, `ff00::/8`.
   - Allow explicitly configured lab/twin CIDRs (`CGO_ALLOW_TARGET_CIDRS`) ONLY when explicitly passed at startup.
   - Implement `SafeDialer(allowCIDRs []netip.Prefix)` using `net.Dialer{Timeout: 5 * time.Second, Control: ...}` that parses `address` (`netip.ParseAddrPort`) **at socket connect time** so DNS is resolved once and validated on the exact IP dialed (eliminating TOCTOU!).
2. Wire this exact `SafeDialer` into BOTH `http.Transport.DialContext` in `pkg/probe/probe.go` AND raw TCP dialing in `pkg/probe/bulk_cc_linux.go:14-27`. Remove duplicate DNS lookup in `api/ssrf.go:92-97` (keep fast static/literal/host syntax check at handler entry, leaving dial-time resolution to `SafeDialer`).

## 3. Fix `A03` — Strict Interface Name Validation (`?iface=`)
Files: `api/server_audit.go:363-372`, `pkg/remediate/remediate.go:85-102, 137-186`
1. Add strict Linux `IFNAMSIZ` interface validation: `^[a-zA-Z0-9_.-]{1,15}$`.
2. Enforce it in `Link.Validate()` inside `pkg/remediate/remediate.go` AND inside `GET /api/remediate` (`api/server_audit.go:363-372`), returning HTTP `400` if `?iface=` contains whitespace, `;`, `|`, `&`, `$`, backticks, quotes, `/`, newlines, or exceeds 15 chars.

## 4. Fix Server Timeouts, `writeJSON`, `60s` Default & `docs/api.md`
Files: `api/server.go`, `api/sse.go`, `pkg/audit/audit.go`, `pkg/probe/probe.go`, `pkg/campagne/modprobe_linux.go`, `cmd/cgo/expose.go`, `cmd/cgo/server.go`, `docs/api.md`
1. **`writeJSON` (`api/server.go:1550-1556`):** Encode JSON into `bytes.Buffer` BEFORE calling `w.WriteHeader(status)`. On encode error, log safely and respond `500`.
2. **`WriteTimeout` (`api/server.go:346-364`, `api/sse.go`):** Set `WriteTimeout: 30 * time.Second` on `http.Server`. Inside `/api/stream` and `/api/replay/stream`, clear write deadline on that response via `http.NewResponseController(w).SetWriteDeadline(time.Time{})`.
3. **Bound Unbounded Calls:**
   - Attach `context.WithTimeout` (`60s`) to `/api/report/export` (`:962-1067`) & `/api/figures/regen` (`:1073-1081`); attach `r.Context()` + `15m` max cap to `/api/replay/stream` (`:1141`).
   - Switch `pkg/campagne/modprobe_linux.go:15` from `exec.Command` to `exec.CommandContext(ctx, ...)` (`10s` timeout).
   - Add `5s` context timeout to `net.LookupIP` in `cmd/cgo/expose.go:47`.
   - Fix `pkg/probe/probe.go:98-106`: set default `SmallClientTimeout` floor to `max(8 * time.Second, 4 * delay)` so VSAT profiles (`600 ms+`) never falsely time out at `2s`.
   - Fix `cmd/cgo/server.go:247-249` `both` polling loop so it respects context cancellation cleanly instead of blind 18-minute sleep loops.
4. **Unify `60s` Default & Docs (`api/server.go:106`, `pkg/audit/audit.go:740-764`, `docs/api.md:3-4`):**
   - Set API default audit duration to **`60 s`** (`10s` idle / `40s` load / `10s` drain), matching `PRODUCT.md:27`.
   - Update `docs/api.md:3-4` to explicitly document `4 KiB` limit on `/api/auth/login` and `1 MiB` elsewhere.

## Constraints
- Add comprehensive unit tests covering: cross-owner `403`, unauth `401` (when `--dev-no-auth` not passed), session logout revocation, missing `X-Confirm-Action` returning `400`, SSRF blocking `100.64.0.1` + `169.254.169.254` + `2001:db8::1` + `::ffff:127.0.0.1` across BOTH HTTP and raw TCP bulk dialers, `?iface=eth0;id` returning `400`, and `writeJSON` failing cleanly with `500`.
- Commit all Chunk 1 changes in ONE clean commit: `fix(security): enforce owner authz, dial-time ssrf, iface validation and timeouts`.

PART B — Your Hard Verification Gate (Run Before Chunk 2)

Run these commands yourself after the agent finishes Chunk 1:

# 1. Run package tests for api, probe, remediate, audit, cmd/cgo
go test -v -count=1 ./api/... ./pkg/probe/... ./pkg/remediate/... ./pkg/audit/... ./cmd/cgo/...

# 2. Verify SSRF covers CGNAT, v4-mapped IPv6, benchmark, doc, and bulk TCP dialer
git grep -n "100.64.0.0/10" pkg/probe/
git grep -n "SafeDialer\|DialContext\|Control:" pkg/probe/bulk_cc_linux.go

# 3. Verify iface regex validation exists in both remediate and api
git grep -n "a-zA-Z0-9" pkg/remediate/ api/

# 4. Verify WriteTimeout is non-zero in api/server.go and ResponseController is in SSE
git grep -n "WriteTimeout" api/
git grep -n "ResponseController" api/

# 5. Verify no deterministic "cgo-session-v1" static cookie key remains
git grep "cgo-session-v1" api/ || echo "CLEAN_RANDOM_SESSION_KEY"

# 6. Verify binary builds and default audit duration is 60s
go build ./cmd/cgo
git grep -n "60" api/server.go | head -n 20

Once every check passes cleanly, say "Proceed with Chunk 2" and we will lock
down K4, K6, K8, failed-scrape honesty, /api/integrity, and fail-closed dualns
RTT symmetry!

