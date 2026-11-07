# Module auth + rate-limit — `cgo-guard` ironing (deep research, no code)

**Date:** 2026-11-07 · **Scope:** `pkg/api/auth.go` (`requireAuth`), `pkg/api/server.go`
(`crossSiteGuard`, routes, `writeErr`/`writeJSON`), `deploy/nginx/cgo.conf`
(`limit_req`), `docs/deploy.md` §6/S1 — as a standalone extraction candidate.
**Status:** research only, no code change, no commit.
**Primary sources fetched 2026-11-07:**
RFC 9110 *HTTP Semantics* (IETF, Standards Track, June 2022) — https://www.rfc-editor.org/rfc/rfc9110.html ;
RFC 6585 *Additional HTTP Status Codes* (IETF, Standards Track, April 2012, for 429) — https://www.rfc-editor.org/rfc/rfc6585.html ;
OWASP *Cross-Site Request Forgery Prevention Cheat Sheet* — https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html ;
Nginx *ngx_http_limit_req_module* — https://nginx.org/en/docs/http/ngx_http_limit_req_module.html ;
Go `net/http` package reference — https://pkg.go.dev/net/http .
Anything below that could not be traced to one of those (or to the repo file:line cited
next to it) is marked `unverified — primary source not found`.

> Note on the brief: 429 is **not** in RFC 9110. RFC 9110 §15 enumerates 401/403/503
> and does not register 429; Go `net/http` lists `StatusTooManyRequests = 429 // RFC 6585, 4`
> (Go `net/http` — Constants) and RFC 6585 §4 defines 429. The brief's “RFC 9110 (… 403/429)”
> grouping is therefore treated as RFC 9110 for 401/403/503 and RFC 6585 §4 for 429.

---

## 1. Responsibility (as-found vs. as-required)

**As-found responsibility (repo's honest account).**

- `requireAuth` — “mutations require a bearer token when `CGO_API_TOKEN` is set.
  Empty token is dev mode: loopback only, still 401 for non-loopback callers.”
  (`pkg/api/auth.go:9-10`) On success it calls `next(w, r)`; on failure it sets
  `WWW-Authenticate: Bearer` and returns 401 via `writeErr`. (`pkg/api/auth.go:11-25`)
- `crossSiteGuard` — “un navigateur venu d'un autre site ne déclenche pas de mutation
  (CSRF): `Sec-Fetch-Site: cross-site` est refusé sur les verbes non sûrs. Les clients
  non-navigateur n'envoient pas cet en-tête et restent inchangés; aucun jeton n'est
  introduit.” (`pkg/api/server.go:1434-1437`) It wraps the whole `mux`
  (`pkg/api/server.go:1431`).
- Nginx `cgo.conf` — TLS boundary + `limit_req` per IP: `cgo_reads 30r/s` and
  `cgo_mutations 1r/s`, `limit_req_status 503`. (`deploy/nginx/cgo.conf:1-9`) Own docs
  describe S1 as “TLS termination + `limit_req`” and “`POST` mutations that exceed
  `1r/s burst=3` receive `503`.” (`docs/deploy.md:179-180`)
- The deployment doc's honesty line (§6) is the baseline: “pas d'authentification ni
  d'autorisation … pas de jeton CSRF ni de rate limit; les mutations refusent seulement
  `Sec-Fetch-Site: cross-site` (garde navigateur, contournable par un client direct);
  le corps JSON est borné à 1 Mio, ce n'est pas un contrôle d'accès … le façonnage
  modifie réellement le bord.” (`docs/deploy.md:159-177`) The enterprise design
  classifies this as a hard FAIL citing those lines. (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:14`)

**As-required responsibility (per primary sources — see §4).**

1. Reject unauthenticated mutations with **401 + `WWW-Authenticate`** (RFC 9110 §15.5.2, §11.6.1).
2. Reject CSRF-forged state-changing requests with **403** by a defense that OWASP
   considers sufficient — synchronizer token, signed double-submit, custom-header +
   CORS preflight, *or* Fetch Metadata **with the mandatory Origin/Referer fallback**
   (OWASP CSRF Cheat Sheet — Introduction; “Fetch Metadata headers”). A bare
   `Sec-Fetch-Site: cross-site` deny with fail-open on absence does **not** satisfy
   that bar (§4.2).
3. Reject over-rate requests with **429 + explanation + optional `Retry-After`,
   not cache-stored** (RFC 6585 §4; `Retry-After` field itself is RFC 9110 §10.2.3).
   Nginx `limit_req` is a per-key “leaky bucket” (Nginx `ngx_http_limit_req_module`
   — Synopsis) and returns `limit_req_status` (default 503) for rejected requests
   (Nginx docs — `limit_req_status`); mapping it to 429 is the target, not the
   current 503 (§5).

The extraction (`cgo-guard`) must own exactly those three behaviors + the error
shape, and nothing of the measurement core. Anything calling a bench, a qdisc, a
profile, or a CSV is out of responsibility for the new module.

---

## 2. Exported / load-bearing surface — `file:line`

Exact pointers at HEAD of this worktree (verify with `git log --oneline -1` before reuse).

### 2.1 `pkg/api/auth.go` (26 lines, whole file)

| Symbol / line | What it does |
|---|---|
| `requireAuth(next http.HandlerFunc) http.HandlerFunc` — `pkg/api/auth.go:11` | Middleware wrapper for mutations only. Imports are only `net/http`, `os`, `strings`. (`pkg/api/auth.go:3-7`) |
| `os.Getenv("CGO_API_TOKEN")` per request — `pkg/api/auth.go:13` | Token read on every request, no startup snapshot. |
| `strings.TrimPrefix(Authorization, "Bearer ")` — `pkg/api/auth.go:14` | Non-case-normalized strip; any header not starting with exactly `Bearer ` yields the raw header as `got` and is compared (and rejected) below. |
| `want != "" && got == want` — `pkg/api/auth.go:15` | Success path → `next(w, r)` (`pkg/api/auth.go:16`). |
| `want == "" && HasPrefix(RemoteAddr, "127.0.0.1")` — `pkg/api/auth.go:19` | Dev-mode loopback bypass → `next(w, r)` (`pkg/api/auth.go:20`). |
| `WWW-Authenticate: Bearer` + `writeErr(..., 401)` — `pkg/api/auth.go:23-24` | Failure path. Order is correct: header set before `WriteHeader` via `writeErr`. |

### 2.2 `pkg/api/server.go` — guards, error shape, server construction

| Symbol / line | What it does |
|---|---|
| `Handler` struct + `CloseHub` — `pkg/api/server.go:237-247` | `Handler` embeds `http.Handler` plus `hubClose`; `New` returns it wrapping `crossSiteGuard(mux)`. (`pkg/api/server.go:1431`) |
| `New(d Deps) Handler` — `pkg/api/server.go:380` | Builds `ServeMux`, registers all routes, then returns `Handler{Handler: crossSiteGuard(mux), ...}`. |
| `crossSiteGuard(next http.Handler) http.Handler` — `pkg/api/server.go:1438-1450` | Allows `GET`/`HEAD`/`OPTIONS` unconditionally (`pkg/api/server.go:1440-1442`); on other methods denies `Sec-Fetch-Site == "cross-site"` with 403 (`pkg/api/server.go:1443-1445`), else `next.ServeHTTP`. |
| `writeErr(w, r, msg, code)` — `pkg/api/server.go:1474-1483` | Content-negotiated error: JSON `{"error": msg}` when `Accept` or `Content-Type` contains `application/json`, else `http.Error` (text/plain). |
| `writeJSON` — `pkg/api/server.go:1464-1470` | `Content-Type: application/json` + `json.NewEncoder(w).Encode` (logs on encode error). |
| `Deps` — `pkg/api/server.go:173-216` | Server seams (`GetSnap`, `StartFn`, `ShapeFn`, `RunningFn`, …). Auth does not go through `Deps` — it reads `os.Getenv` directly. This is an extraction seam: `cgo-guard` must accept a config/validator instead of reading env. |
| `validRunID` — `pkg/api/server.go:320` | Adjacent (not auth) path guard, reused by `run/start`, `results`, `replay/stream`, `run/rows`, `quarantine`. Included here because a standalone guard must not break this invariant. |

### 2.3 Route inventory — `pkg/api/server.go` (authenticated vs. open)

**Mutations wrapped in `requireAuth` (13 routes).**

| Route | Line |
|---|---|
| `POST /api/cost/tier` | `pkg/api/server.go:455` |
| `POST /api/burst` | `pkg/api/server.go:479` |
| `POST /api/watch` | `pkg/api/server.go:571` |
| `POST /api/shape` | `pkg/api/server.go:596` |
| `POST /api/run/start` | `pkg/api/server.go:614` |
| `POST /api/run/stop` | `pkg/api/server.go:746` |
| `POST /api/run/skip` | `pkg/api/server.go:755` |
| `POST /api/figures/regen` | `pkg/api/server.go:1084` |
| `POST /api/audit/start` | `pkg/api/server.go:1149` |
| `POST /api/audit/stop` | `pkg/api/server.go:1152` |
| `POST /api/audit/toprofile` | `pkg/api/server.go:1157` |
| `POST /api/profile/import` | `pkg/api/server.go:1158` |
| `POST /api/profile/delete` | `pkg/api/server.go:1233` |

**Reads / other routes with no `requireAuth` (34 patterns).** Pattern = literal in
`mux.HandleFunc`/`mux.Handle` (Go `ServeMux` method routing per Go `net/http` —
`ServeMux.Handle`, `ServeMux.HandleFunc`).

`GET /api/health` (`:394`), `GET /metrics` (`:397`, metrics handler), `GET /api/capabilities`
(`:402`), `GET /api/doctor` (`:405`), `GET /api/schema` (`:413`), `GET /api/cost/tiers`
(`:417`), `GET /api/cost/enterprise` (`:423`), `GET /api/state` (`:527`),
`GET /api/profiles` (`:531`), `GET /api/shape` (`:561`), `GET /api/results` (`:764`),
`GET /api/results/delta` (`:779`), `GET /api/duel` (`:877`), `GET /api/results/compare`
(`:882`), `GET /api/integrity` (`:924`), `GET /api/report/export` (`:977`),
`/api/figures/` static file server (`:1091`), `GET /api/replay/list` (`:1092`),
`GET /api/replay/stream` (`:1102`), `GET /api/audit/status` (`:1150`),
`GET /api/audit/live` (`:1151`), `GET /api/audit/report` (`:1153`), `GET /api/remediate`
(`:1154`), `GET /api/audit/presets` (`:1155`), `GET /api/audit/list` (`:1156`),
`GET /api/profile/list` (`:1252`), `GET /api/diagnostics` (`:1264`),
`GET /api/hardware/translate` (`:1267`), `/api/` 404 catch-all (`:1269`),
`GET /api/stream` SSE (`:1275`), `GET /api/run/rows` (`:1280`),
`GET /api/quarantine` (`:1312`), `GET /api/quarantine/summary` (`:1362`),
`GET /api/events` (`:1423`), `GET /` SPA (`:1430`).

Consequence: only the 13 mutations are bearer-gated; every read is reachable
without any token. The enterprise target requires reads to be authenticated on any
non-loopback bind (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:52`).

### 2.4 `deploy/nginx/cgo.conf` (35 lines, whole file)

| Line | Directive | Effect |
|---|---|---|
| `:1` | `limit_req_zone $binary_remote_addr zone=cgo_reads:10m rate=30r/s;` | Leaky-bucket zone `cgo_reads`, key = client IPv4/IPv6 binary IP. |
| `:2` | `limit_req_zone $binary_remote_addr zone=cgo_mutations:10m rate=1r/s;` | Zone `cgo_mutations`, same key. |
| `:9` | `limit_req_status 503;` | Rejected requests return 503 (Nginx default is 503 per `ngx_http_limit_req_module` — `limit_req_status`). |
| `:11-15` | `location /api/health` → `cgo_reads burst=10 nodelay` | Health under reads zone. |
| `:17-22` | `location /metrics` → `allow 127.0.0.1; deny all;` + `cgo_reads burst=10 nodelay` | Only loopback may fetch `/metrics`. |
| `:24-28` | `location /api/` → `cgo_mutations burst=3 nodelay` | **All** `/api/` traffic (reads **and** writes) under `1r/s`. |
| `:30-34` | `location /` → `cgo_reads burst=10 nodelay` | SPA under reads zone. |

All locations `proxy_pass http://127.0.0.1:9090`. (`deploy/nginx/cgo.conf:13,21,26,32`)
The `server` block itself listens `9090 ssl` (`deploy/nginx/cgo.conf:5`).

### 2.5 Tests + specs pinned to these lines

- `pkg/api/auth_test.go:10-23` — one test: without token → 401.
- `pkg/api/hardening_test.go:22-58` — `TestCrossSiteGuard`: `POST` with
  `Sec-Fetch-Site: cross-site` → 403, without header → 200, `GET` cross-site → 200.
- `docs/superpowers/plans/2026-11-06-cgo-enterprise-hardening.md:132` — planned
  `requireAuth` contract (env var, loopback warning, 401).
- `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:35,59,107`
  — S2 contract: authenticated sessions on all mutations, CSRF on cookie flows,
  unauthenticated → 401, CSRF failure → 403, rate-limited → 503.

---

## 3. Dependencies

### 3.1 Go imports (direct)

- `pkg/api/auth.go`: `net/http`, `os`, `strings`. (`pkg/api/auth.go:3-7`) No
  third-party import, no config struct, no clock, no store. `net/http` is the stdlib
  HTTP server package (Go `net/http` — Overview: “Package http provides HTTP client
  and server implementations”). `Header.Get`/`Header.Set` and `ServeMux`/
  `HandlerFunc.ServeHTTP` are the stdlib primitives the guards rely on (Go `net/http`
  — `Header.Get/Set`, `HandlerFunc.ServeHTTP`, `ServeMux.Handle/HandleFunc`).
- `pkg/api/server.go`: guard code path adds no new import beyond what `New` already
  imports (`net/http`, `os`, `strings`, …). (`pkg/api/server.go:3-30`)
- `writeErr` uses `strings.Contains`, `json.NewEncoder`, `http.Error` (Go
  `net/http` — `Error(w, error, code)`). (`pkg/api/server.go:1474-1483`)

### 3.2 Environment / process

- `CGO_API_TOKEN` (env, read per request). (`pkg/api/auth.go:13`) No default,
  no file-based secret, no rotation hook in this module.
- Implicit reliance on `r.RemoteAddr` (`pkg/api/auth.go:19`) and on the caller not
  wrapping the handler with `X-Forwarded-For` rewriting. No `XFF` handling exists
  in `pkg/api/auth.go` or `pkg/api/server.go` (grep: zero hits for `X-Forwarded-For`
  in `pkg/api/`).
- No cookie/session store exists. CSRF protection is header-based only
  (`pkg/api/server.go:1438-1450`), matching the design caveat that “aucun jeton
  n'est introduit” (`pkg/api/server.go:1437`).

### 3.3 Nginx

- Module `ngx_http_limit_req_module`, “leaky bucket”, per-key rate limiting
  (Nginx docs — Synopsis). Key `$binary_remote_addr` (Nginx docs — `limit_req_zone`
  example uses the same variable).
- Tuning knobs in use: `rate=…r/s`, `burst=…`, `nodelay`, `limit_req_status`.
  Semantics in §4.3 (Nginx docs — `limit_req`, `limit_req_zone`, `limit_req_status`).

### 3.4 What the guards do **not** depend on (extraction relief)

- `Deps` seams (`StartFn`/`ShapeFn`/…) are **not** consulted by `requireAuth` or
  `crossSiteGuard` — they only wrap handlers. (`pkg/api/auth.go:11-25`; `pkg/api/server.go:1438-1450`)
  The guards are already separable without touching bench/audit/profile code.
- `writeErr`/`writeJSON` are pure `net/http` helpers; extracting them is a
  copy with no hidden coupling. (`pkg/api/server.go:1464-1483`)
- Nginx config is standalone (`deploy/nginx/cgo.conf`) and Go does not parse it.

---

## 4. Primary-source requirements

Citations are to the fetched texts (§“Sources”). Section numbers are the ones that
govern behavior — not a full paraphrase of each RFC.

### 4.1 Bearer auth — 401 + `WWW-Authenticate` (RFC 9110 §11, §15.5.2)

- **401 signals “not authenticated.”** RFC 9110 §15.5.2 defines 401 Unauthorized as
  indicating the request has not been applied because it lacks valid authentication
  credentials for the target resource; a client must not repeat the same credentials
  and should be prompted to present new ones. Current `requireAuth` returns 401 on
  any missing/mismatched token. (`pkg/api/auth.go:23-24`)
- **A 401 MUST carry `WWW-Authenticate` with at least one challenge.**
  RFC 9110 §11.6.1 specifies `WWW-Authenticate` is sent by the server to
  authenticate a user agent to a target resource; §11 (“HTTP Authentication”)
  frames header-based authentication as challenge (`WWW-Authenticate`) followed by
  client credentials (`Authorization`, §11.6.2). Current code sets exactly
  `WWW-Authenticate: Bearer`. (`pkg/api/auth.go:23`)
- **Challenge without `realm` is permitted but minimal.** RFC 9110 §11.5 (Protection
  Space / realm) makes the realm an optional auth-param; a bare scheme token satisfies
  §11.6.1's “at least one challenge.” Current code provides no `realm` param.
  (`unverified — primary source not found` for whether an enterprise buyer/regulator
  will demand a realm param; RFC 9110 does not mandate one.)
- **403 ≠ 401.** RFC 9110 §15.5.4 defines 403 Forbidden as: server understands the
  request but refuses to authorize it, and resubmitting the same request will not help.
  Current CSRF denial uses 403 (`pkg/api/server.go:1444`) — semantically a CSRF
  failure is an authorization refusal, not missing credentials — consistent with
  RFC 9110 §15.5.4 and with the design's “CSRF failure → 403” (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:107`).
- **Bearer wire syntax.** RFC 9110 §11.6.2 governs the `Authorization` field
  (`Credentials` per scheme); the exact `Bearer <token>` grammar and the 401/403
  error payload for token schemes is defined by RFC 6750 (Bearer Token Usage),
  **not fetched in this pass**. Any claim about RFC 6750 §2.1 (`Authorization: Bearer …`),
  invalid-token error codes, or `error="invalid_token"` params is therefore
  `unverified — primary source not found` until RFC 6750 is fetched and pinned.

**Target invariant for `cgo-guard` (bearer):** missing/invalid credentials → 401
with `WWW-Authenticate: Bearer` (optionally `realm="…"`); never 403 and never 200.
On success, proceed exactly as the handler demands. Comparison must be constant-time
(§4.2).

### 4.2 CSRF — OWASP CSRF Cheat Sheet + RFC 9110 §9.2.1 (safe methods)

OWASP calls out the overall defense posture in one line: check for a framework
built-in defense first; otherwise add CSRF tokens to all state-changing requests and
validate server-side; if only modern browsers are targeted, you *may* rely on Fetch
Metadata “together with the fallback options described below” (OWASP CSRF —
Introduction). Three sub-requirements matter for this module:

1. **Fetch Metadata primary signal is `Sec-Fetch-Site`, and thedeny rule is
   cross-site + non-safe.** OWASP (“How to treat Fetch Metadata headers… §1.1”):
   “By default, reject non-safe methods (POST / PUT / PATCH / DELETE) when
   `Sec-Fetch-Site: cross-site`.” Safe set quoted by OWASP:
   `GET, HEAD, OPTIONS`. Current guard allows exactly `GET`/`HEAD`/`OPTIONS` and
   denies `cross-site` otherwise. (`pkg/api/server.go:1440-1445`) Alignment: one-for-one.
2. **Safe means read-only.** RFC 9110 §9.2.1 defines safe methods (GET/HEAD/OPTIONS
   in practice) as typically read-only with respect to resource state. The guard
   treats those three as non-mutating; this matches RFC 9110 §9.2.1 **only if**
   no handler mutates state on GET. `GET /api/run/rows`, `GET /api/report/export`,
   etc. only read frozen data per their handlers (`pkg/api/server.go:764-1500+`
   inspected); `GET /api/profiles` calls `profile.Load()` (merge-persisted state)
   at `pkg/api/server.go:531-532` — a read-with-side-effect that is not a state
   change of the measurement itself. Whether this violates RFC 9110 §9.2.1's
   “typically” wording is a design judgment, not a hard breach (`unverified — primary source not found`
   for a normative MUST on GET side-effects).
3. **Fallback when `Sec-Fetch-*` is absent is mandatory, not optional.**
   OWASP: “Because some legacy browsers do not send `Sec-Fetch-*` headers, a fallback
   to standard origin verification headers **is a mandatory requirement** for any
   Fetch Metadata implementation.” (`Sec-Fetch-*` in all major browsers since
   Mar 2023 per caniuse via the Cheat Sheet.) For missing headers OWASP enumerates
   two choices: (2.1) fail-safe — treat absence as unknown and block (recommended for
   sensitive endpoints); (2.2) fail-open — fall back to Origin/Referer verification,
   CSRF tokens, or additional validation. Current guard **fails open with no
   fallback**: absence → `next.ServeHTTP`. (`pkg/api/server.go:1440-1448`) This is
   the single largest CSRF gap (§5).
4. **`same-origin` vs `same-site` handling must be explicit.** OWASP §1.3: allow
   `same-origin`; treat `same-site` as allowed only if sibling subdomains are
   trusted, otherwise handle conservatively (e.g., require extra validation for
   state-changing methods). Current guard has **no `same-site` branch at all** —
   `same-site` + POST falls through to the handler. (`pkg/api/server.go:1440-1448`)
   Whether the CGO deployment has trusted subdomains is not recorded in
   `docs/deploy.md` (§6) or `deploy/nginx/cgo.conf`; `unverified — primary source not found`
   for the trust decision.
5. **Trustworthy URLs + HSTS are preconditions.** OWASP (“Requirements”):
   Fetch Metadata headers are only sent to *potentially trustworthy URLs*
   (https/wss/file/localhost); HTTPS must be enforced across the app (HSTS helps).
   `deploy/nginx/cgo.conf` terminates TLS on 9090 ssl (`deploy/nginx/cgo.conf:5-6`)
   while `docs/deploy.md:179-180` notes direct `--tls=false` binds are loopback-only
   in CI. On non-localhost plain HTTP, browsers will not send `Sec-Fetch-*` — the
   fail-open then also admits browser CSRF. (`unverified — primary source not found`
   for HSTS presence — no `Strict-Transport-Security` header is set in `deploy/nginx/cgo.conf`
   or `pkg/api/server.go`.)
6. **Operational rollout requirements.** OWASP (“Rollout & testing…”): include
   `Vary: Sec-Fetch-Site, Origin` so caches do not conflate contexts; start in
   log-only mode and review false positives before enforcing; monitor UA coverage;
   document whitelisted cross-origin endpoints. Current guard sets **no `Vary`**,
   has **no log-only mode**, has **no allowlist**, and performs **no UA logging**.
   (`pkg/api/server.go:1438-1450`)
7. **Token and custom-header alternatives (the rest of the Cheat Sheet).**
   OWASP's recommended defenses are, in order: synchronizer token pattern, signed
   double-submit (HMAC, session-bound, `Secure` cookie), or custom request headers
   (which trigger CORS preflight; server rejects if the header is absent). Key
   constraints:
   - Token must be unique per session, secret, unpredictable via a secure random
     method; compare server-side and reject on mismatch (OWASP — Synchronizer Token
     Pattern). Token must not be transmitted in a cookie for synchronizer patterns,
     must not leak in logs/URL/Referer, and should be sent in a custom header rather
     than a hidden form field for AJAX (OWASP — Transmitting CSRF Tokens).
   - HMAC tokens: recompute with the same encoding and compare with a
     **constant-time comparison function** (`hmac.compare_digest` analogue) (OWASP —
     HMAC Token Generation and Validation).
   - Custom header: ANY custom header triggers preflight in modern browsers; the
     server verifies it is present and rejects otherwise; CORS must then allow only
     explicitly-listed origins (OWASP — Employing Custom Request Headers;
     Custom Headers and CORS). The current module has **none** of these — no token,
     no custom-header check, no CORS config (`pkg/api/server.go` has zero `Access-Control-*`).
   - SameSite is defense-in-depth only, with subdomain-lookalike caveats (OWASP —
     SameSite). No session cookie exists in CGO, so SameSite cannot currently bite
     or help. Setting `SameSite` without a cookie is a no-op (trivially true, `unverified — primary source not found`
     for any framework-specific interaction).
8. **Built-in Go alternative exists.** OWASP notes (from Go 1.25): `net/http` provides
   `CrossOriginProtection`, implementing Fetch-Metadata-based CSRF defense including
   `Sec-Fetch-Site` validation, in the standard library (“Built-In Or Existing CSRF
   Implementations”). Go `net/http` exposes `type CrossOriginProtection` with
   `NewCrossOriginProtection()`, `Check(req *Request) error`,
   `Handler(h Handler) Handler`, `AddTrustedOrigin(origin)` (Go `net/http` — Index:
   `CrossOriginProtection`, `Check`, `Handler`, `AddTrustedOrigin`). `cgo-guard`
   should wrap or replicate this type rather than inventing a third Fetch-Metadata
   implementation — OWASP's first principle is “use the built-in if one exists.”

### 4.3 Rate limiting — Nginx `limit_req` leaky bucket + RFC 6585 §4 (429) + RFC 9110 §10.2.3/§15.6.4

**Nginx mechanics (verbatim semantics from `ngx_http_limit_req_module`):**

- The module “limits the request processing rate per a defined key … using the
  ‘leaky bucket’ method.” (Synopsis)
- `limit_req_zone key zone=name:size rate=rate [sync];` creates a shared-memory zone
  storing “the current number of excessive requests” per key. (Nginx docs —
  `limit_req_zone`) “Requests with an empty key value are not accounted.” (id.)
- Rate is expressed in **r/s**, or **r/m** when less than 1 r/s is desired (e.g.,
  `30r/m` = 0.5 r/s). (id.) Current zones: `30r/s` (reads) and `1r/s` (mutations).
  (`deploy/nginx/cgo.conf:1-2`)
- `limit_req zone=name [burst=number] [nodelay | delay=number];` — “If the requests
  rate exceeds the rate configured for a zone, their processing is delayed such that
  requests are processed at a defined rate. Excessive requests are delayed until
  their number exceeds the maximum burst size in which case the request is terminated
  with an error.” Default `burst=0`. With `nodelay`, delaying is suppressed:
  “all excessive requests [beyond burst] are rejected immediately.” (Nginx docs —
  `limit_req`) Current config uses `burst=10 nodelay` (reads/tasks) and
  `burst=3 nodelay` (mutations). (`deploy/nginx/cgo.conf:12,20,25,31`)
- Inherited declarations: “There could be several `limit_req` directives… These
  directives are inherited from the previous configuration level if and only if
  there are no `limit_req` directives defined on the current level.” (Nginx docs —
  `limit_req`) Each `location` re-declares, so no inheritance surprises here.
- **Rejection status:** `limit_req_status code;` “Sets the status code to return in
  response to rejected requests. Default: `limit_req_status 503`.” (Nginx docs —
  `limit_req_status`) Current config pins `limit_req_status 503`. (`deploy/nginx/cgo.conf:9`)
- **State size / exhaustion:** `$binary_remote_addr` is 4 B (IPv4) / 16 B (IPv6);
  zone state is 64 B (32-bit) / 128 B (64-bit); 1 MB ≈ 16k (64 B) or 8k (128 B)
  states; 10 MB ⇒ ~80–160k keys. “If the zone storage is exhausted, the least
  recently used state is removed. If even after that a new state cannot be created,
  the request is terminated with an error.” (Nginx docs — `limit_req_zone`)
  Current zones are `10m` each. (`deploy/nginx/cgo.conf:1-2`)
- **Dry-run / logging:** `limit_req_dry_run on|off` (no limiting, count only);
  `limit_req_log_level` controls refusal/delay logging (Nginx docs —
  `limit_req_dry_run`, `limit_req_log_level`). Current config sets neither —
  refusals log at default `error`. (`deploy/nginx/cgo.conf` — absent)

**HTTP semantics of rejection:**

- RFC 6585 §4 defines **429 Too Many Requests**: “user has sent too many requests
  in a given amount of time (‘rate limiting’).” Semantics: “The response
  representations SHOULD include details explaining the condition, and MAY include a
  `Retry-After` header indicating how long to wait before making a new request.”
  “Note that this specification does not define how the origin server identifies the
  user, nor how it counts requests… [per-resource, per-server, per-credential, or
  cookie].” “Responses with the 429 status code MUST NOT be stored by a cache.”
- The `Retry-After` field itself is defined in RFC 9110 §10.2.3 (delay-seconds or
  HTTP-date). Nginx `limit_req` does **not** emit `Retry-After` by default —
  `unverified — primary source not found` for any Nginx mechanism to set
  `Retry-After` per-rejection without an `error_page`/`proxy_set_header` rewrite
  (not in `ngx_http_limit_req_module` docs fetched).
- RFC 9110 §15.6.4 defines **503 Service Unavailable**: temporarily unable to handle
  the request, typically due to overload or maintenance; the client may retry.
  Current 503-on-rate-limit therefore **mis-communicates** “the server is down”
  instead of “you are over your rate.” Consistent with this, the enterprise design
  embeds the wrong status: “rate-limited → `503` (Nginx `limit_req_status`)”
  (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:107`) —
  correct as a description of the *current* config, incorrect as a target
  (RFC 6585 §4 requires 429 semantics).
- Go `net/http` constant confirms the split: `StatusTooManyRequests = 429 // RFC 6585, 4`
  and `StatusServiceUnavailable = 503 // RFC 9110, 15.6.4` (Go `net/http` —
  Constants: StatusCodes).

**Go-side equivalent:** Go `net/http` provides no built-in rate limiter; it offers
`MaxBytesHandler`, `TimeoutHandler`, `Handler`/`Header` primitives (Go `net/http` —
Functions/Types). An in-process leaky-bucket equivalent (e.g., `golang.org/x/time/rate`)
is a **new dependency** requiring human approval per the design's open decisions
(`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:131-135`)
— treated as out-of-scope for extraction unless the human approves (§6).

### 4.4 Go `net/http` primitives the module rests on

- `http.Handler`, `http.HandlerFunc.ServeHTTP(w, r)`, `http.ServeMux.Handle`/
  `HandleFunc` — the server model the guards wrap (Go `net/http` — Overview:
  Servers; Index).
- `Header.Get(key)` / `Header.Set(key, value)` — reading `Authorization` and
  `Sec-Fetch-Site`, writing `WWW-Authenticate` (Go `net/http` — `Header.Get/Set`).
- `Request.RemoteAddr`, `Request.Header`, `Request.Method`, `Request.Context()` —
  the signals `requireAuth`/`crossSiteGuard` read (Go `net/http` — `Request`).
- `Error(w, error, code)` / `WriteHeader(code)` / `Encode` — the error surface
  `writeErr`/`writeJSON` use (Go `net/http` — `Error`).
- `CrossOriginProtection` (Go ≥1.25 per OWASP) with `Check(req)` /
  `Handler(h)` / `SetDenyHandler(h)` / `AddTrustedOrigin` — the stdlib CSRF
  building block the extraction must use (Go `net/http` — Index:
  `CrossOriginProtection`; OWASP — Built-In Implementations).

---

## 5. Gap analysis (every gap cites its primary source)

| # | Gap (as-found location) | Norm violated | Primary source + current line |
|---|---|---|---|
| G1 | **CSRF fails open on absence.** Absent `Sec-Fetch-Site` → `next` (`pkg/api/server.go:1442-1448`). | OWASP: fallback is mandatory; fail-safe for sensitive endpoints = block on absence. | OWASP CSRF — Fetch Metadata intro (“fallback … is a mandatory requirement”) + §2; `pkg/api/server.go:1440-1448` |
| G2 | **`same-site` unhandled.** `same-site` + POST falls through to handler (`pkg/api/server.go:1440-1448`). | OWASP §1.3 requires explicit same-site policy (trust or conservative reject). | OWASP CSRF — “How to treat Fetch Metadata…” §1.3 |
| G3 | **No Origin/Referer verification, no token, no custom-header, no CORS.** Zero `Origin`/`Referer`/`X-Custom`/`Access-Control-*` handling in `pkg/api/`. | OWASP: Fetch-Metadata-alone is insufficient without fallback; tokens/custom-header are the fallback mechanisms. | OWASP CSRF — Token-Based Mitigation; Employing Custom Request Headers; Using Standard Headers to Verify Origin |
| G4 | **No `Vary`, no log-only, no allowlist, no UA monitoring.** (`pkg/api/server.go:1438-1450`) | OWASP rollout: `Vary: Sec-Fetch-Site, Origin` required for caches; log-only first. | OWASP CSRF — Rollout & testing recommendations |
| G5 | **Trustworthy-URL/HSTS precondition unproven.** TLS on 9090 ssl in Nginx (`deploy/nginx/cgo.conf:5`) but no `Strict-Transport-Security` header found; direct `--tls=false` is loopback-only in CI (`docs/deploy.md:179-180`) — production termination path unverified. | OWASP: Fetch Metadata only sent on trustworthy URLs; HTTPS across app. | OWASP CSRF — Requirements; `deploy/nginx/cgo.conf`; `docs/deploy.md:179-180` (+ `unverified — primary source not found` for HSTS presence) |
| G6 | **Bearer compare is `==`, not constant-time.** (`pkg/api/auth.go:15`) | OWASP HMAC validation: constant-time comparison is the required technique. (Also general token-compare hygiene.) | OWASP CSRF — HMAC Token Generation and Validation (“constant-time comparison”) |
| G7 | **Token read per request via `os.Getenv`, no startup check, rotation, entropy, or versioning.** (`pkg/api/auth.go:13`) | `unverified — primary source not found` — no fetched primary dictates token storage/rotation; design requires auth rotation runbooks (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:80`) but defines no mechanism. |  (no primary; flagged unverified) |
| G8 | **Dev-mode loopback = `HasPrefix( "127.0.0.1")`.** Misses `::1`, other `127/8` (`127.0.0.2`), and is catastrophically wrong behind Nginx: `proxy_pass 127.0.0.1:9090` makes `RemoteAddr` always loopback, so **any** remote client through Nginx bypasses bearer when `CGO_API_TOKEN` is empty. (`pkg/api/auth.go:19`; `deploy/nginx/cgo.conf:13,21,26,32`) | Requirement is *transport-boundary* correctness; no fetched RFC governs `RemoteAddr` parsing (`unverified — primary source not found` for a normative parser rule). Current prefix check is also inconsistent with the Nginx `/metrics` allowlist (`127.0.0.1` only). |  (code-vs-code mismatch; plus `unverified — primary source not found` for a MUST-level parser rule) |
| G9 | **No `realm` in challenge.** `WWW-Authenticate: Bearer` bare. (`pkg/api/auth.go:23`) | RFC 9110 §11.5 realm is optional; §11.6.1 requires ≥1 challenge. Not a breach, but enterprises may require realm/params. (`unverified` for buyer demand.) | RFC 9110 §11.5, §11.6.1 |
| G10 | **Scheme parsing is exact `Bearer ` (case/whitespace sensitive).** (`pkg/api/auth.go:14`) | RFC 9110 §11 token rules govern parsing; whether `bearer` lowercase must be accepted is `unverified — primary source not found` without RFC 6750 fetch. Current behavior rejects non-exact scheme with 401 — safe but potentially interoperability-breaking. | RFC 9110 §11 (general); exact Bearer grammar `unverified — primary source not found` (RFC 6750 not fetched) |
| G11 | **Reads unauthenticated (34 patterns, §2.3).** Only 13 POSTs gated. GET readers see results, profiles, costs. | Design: “reads require authentication on any non-loopback bind.” (design line 52) — repo-internal requirement, not a fetched primary. RFC 9110 does not mandate auth on reads in general. | design (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:52`); RFC 9110 silent — no primary mandates reads-auth |
| G12 | **Mutation rate limit covers reads; no reads zone for `/api/`.** `location /api/` → `cgo_mutations 1r/s burst=3 nodelay` for **all** `/api/` traffic, including `GET /api/state`, SSE `GET /api/stream` (long-lived), `GET /api/results`. (`deploy/nginx/cgo.conf:24-28`) | Zone-name intent (“mutations”) vs. mechanism (path prefix) mismatch; violates least-surprise and the design's own `POST`-only claim (`docs/deploy.md:179-180`). Worse, SSE stream holds a connection that must be renewed — a burst of 3 then steady 1 r/s starves polling. | `deploy/nginx/cgo.conf:24-28` vs. `docs/deploy.md:179-180`; Nginx inheritance quoted in §4.3 explains why the reads zone never applies here |
| G13 | **Rate-limit status is 503, not 429; no `Retry-After`, no details, no cache-busting.** `limit_req_status 503` (`deploy/nginx/cgo.conf:9`). | RFC 6585 §4: 429 SHOULD include details, MAY include `Retry-After`, MUST NOT be cached. RFC 9110 §15.6.4: 503 = temporary overload/maintenance. | RFC 6585 §4; RFC 9110 §15.6.4; Nginx docs — `limit_req_status`; `deploy/nginx/cgo.conf:9` |
| G14 | **Rate identified per IP only (`$binary_remote_addr`).** (Nginx zones `:1-2`) | RFC 6585 §4 explicitly leaves user identification/counting undefined — per-IP is *permitted*, but NAT/CGNAT sharing and distributed spoofing make per-IP-only weak for mutations. Whether per-credential keying is required is a threat-model choice: `unverified — primary source not found` for a MUST. | RFC 6585 §4 (“does not define how … identifies the user”); Nginx docs — `limit_req_zone` |
| G15 | **No per-token brute-force throttle** in Go; only the 1 r/s mutation bucket (which is path-shared with reads per G12). | `unverified — primary source not found` (no fetched primary defines auth-attempt throttling). |  |
| G16 | **Zone exhaustion / silent LRU.** `10m` per zone (~80–160k keys at 128 B state); LRU drop silently weakens limiting under large key cardinality. (`deploy/nginx/cgo.conf:1-2`) | Nginx docs norm: LRU removal, then request termination with error. For a mgmt network this is theoretical; for Internet-facing it is real. | Nginx docs — `limit_req_zone` (state-size paragraph) |
| G17 | **No dry-run, no `$limit_req_status` logging, no refusal metrics.** Config sets neither `limit_req_dry_run` nor `limit_req_log_level`; `$limit_req_status` never logged. (`deploy/nginx/cgo.conf`) | OWASP log-only rollout (CSRF analogue) + observability of rate limiting (`unverified` as a hard MUST — OWASP rollout is guidance). | OWASP CSRF — Rollout; Nginx docs — `limit_req_dry_run`, `limit_req_log_level`, Embedded Variables `$limit_req_status` |
| G18 | **Error shape varies by `Accept`.** `writeErr` returns JSON for `Accept: application/json`, text otherwise. (`pkg/api/server.go:1475-1482`) | Not a spec breach (RFC 9110 allows representation selection), but auth clients asserting JSON will get text/plain for 401/403 when `Accept` is absent (curl, scripts). | RFC 9110 §12 content negotiation (§12.5 Accept); `pkg/api/server.go:1474-1483` |
| G19 | **`/metrics` IP-allow is the only access control; everything else is open at Nginx.** Only `/metrics` has `allow/deny` (`deploy/nginx/cgo.conf:17-22`). `/api/` and `/` accept any remote IP that reaches Nginx. Mutations then rely solely on Go bearer. | Design: “Nginx MUST NOT proxy `/api/shape` to loopback without auth” and raw loopback-only binds (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:71`). Current config proxies mutations without any Nginx-layer auth — Go is the single point. | design line 71; `deploy/nginx/cgo.conf` |
| G20 | **`proxy_pass` port collision / loop is unverified.** `listen 9090 ssl` + `proxy_pass http://127.0.0.1:9090` implies backend also on 9090. How Go binds (9090 plain? different port?) is not documented in `docs/deploy.md` §S1. | `unverified — primary source not found` — no fetched primary clarifies the deployment topology; runtime config required. |  (flagged unverified) |

**Healthy behaviors to preserve (verified against primaries):**

- 401 + `WWW-Authenticate` order and value on mutation miss: compliant with
  RFC 9110 §11.6.1/§15.5.2 (`pkg/api/auth.go:23-24`).
- 403 (not 401) for CSRF deny: compliant with RFC 9110 §15.5.4 semantics
  (`pkg/api/server.go:1444`).
- Safe-method allowlist `GET/HEAD/OPTIONS`: matches OWASP's `SAFE_METHODS` set and
  RFC 9110 §9.2.1 safe-method family (`pkg/api/server.go:1440-1442`).
- `nodelay` + small burst on mutations: matches Nginx leaky-bucket intent for
  mutations that must not queue (`deploy/nginx/cgo.conf:25` + Nginx docs `limit_req`).
- `/metrics` loopback allow: defense-in-depth consistent with observability
  guidance, but **not** a substitute for mutation auth (primary: repo design).

---

## 6. Target tool — `cgo-guard` (independent repo)

**Principle:** extract the smallest behavior-complete guard, zero measurement code,
zero bench/qdisc/profile awareness. The repo is language-consistent (Go module +
Nginx fragment) and ships as a library first, a templated Nginx fragment second.

### 6.1 Layout (proposed, minimal)

```
cgo-guard/
  go.mod                          # module github.com/Realms4239/cgo-guard (example); go >= 1.25 for net/http CrossOriginProtection
  guard.go                        # Bearer middleware, Fetch-Metadata/CSRF middleware, RateLimit middleware, ErrorWriter
  guard_test.go                   # 401/403/429 + fallback + constant-time + Retry-After tests
  nginx/cgo-guard.conf            # 429-mapped Nginx fragment (drop-in for deploy/nginx/cgo.conf)
  README.md                       # wire contract: status codes, headers, config, rollout
```

No new Go dependency is required for v1: `net/http` (`CrossOriginProtection`,
`Header`, `ServeMux`), `crypto/subtle` (constant-time), `os`/`strings` are stdlib
(Go `net/http` — Index/Functions; OWASP HMAC guidance — constant-time). A token-bucket
store is **not** included in v1 — rate limiting stays in Nginx (independent repo
rule: Nginx already implements the leaky bucket per the fetched docs; replicating
it in Go would add a dependency and a second source of truth — see §6.3 Slice 4
for the optional Go fallback if Nginx is unavailable).

### 6.2 Public API (library)

```go
// guard.Config — all knobs, no env reads.
type Config struct {
    Token       string              // required for Bearer; read once at construction
    Realm       string              // optional WWW-Authenticate realm (RFC 9110 §11.5)
    TrustedOrigins []string         // allowed Origin/Referer values for Fetch-Metadata fallback
    LogOnly     bool                // OWASP rollout mode: log would-block, don't block (OWASP CSRF — Rollout)
    AllowLoopbackWithoutToken bool  // dev-mode; defaults false
}

// guard.Bearer — 401 + WWW-Authenticate: Bearer on miss (RFC 9110 §11.6.1/§15.5.2),
// constant-time compare (OWASP HMAC guidance). Uses crypto/subtle.
func Bearer(cfg Config, next http.Handler) http.Handler

// guard.CSRF — wraps net/http CrossOriginProtection (Go >= 1.25) with explicit
// same-site policy + Origin/Referer double-check + Vary. 403 on deny (RFC 9110 §15.5.4).
func CSRF(cfg Config, next http.Handler) http.Handler

// guard.RateLimit — in-process fallback only (NOT default): fixed-window per
// Bearer identity, 429 + Retry-After, never cached (RFC 6585 §4, RFC 9110 §10.2.3).
// Default production path is Nginx (§6.3 Slice 1); this exists for no-Nginx hosts.
func RateLimit(cfg Config, next http.Handler) http.Handler

// guard.ErrorWriter — single shape for 401/403/429 bodies (details per RFC 6585 §4
// §“response representations SHOULD include details”).
func ErrorJSON(w http.ResponseWriter, code int, msg string)
```

**Composition order** (outer → inner): `CSRF` → `Bearer` → `RateLimit` →
`Handler`. Rationale: CSRF rejection is cheapest and identity-agnostic; Bearer
establishes identity before rate counting so the key can be per-credential
(RFC 6585 §4 explicitly allows per-credential counting); rate limit is last so
unauthenticated probes are already rejected with 401 before they consume the
mutation bucket. This order is a design decision — `unverified — primary source not found`
for a normative middleware order — but it satisfies all three status-code contracts
simultaneously on any single request.

**Behavior contracts (testable):**

| Condition | Status | Headers | Primary |
|---|---|---|---|
| Bearer missing/invalid | 401 | `WWW-Authenticate: Bearer` (+optional `realm="…"`) | RFC 9110 §11.6.1, §15.5.2 |
| CSRF: `Sec-Fetch-Site: cross-site` + non-safe | 403 | `Vary: Sec-Fetch-Site, Origin` | OWASP §1.1; RFC 9110 §15.5.4 |
| CSRF: `same-site` + non-safe, trusted-origin false | 403 | same `Vary` | OWASP §1.3 |
| CSRF: `Sec-Fetch-*` absent + sensitive endpoint + no fallback pass | 403 (fail-safe) or 401/400 per config | same `Vary` | OWASP §2.1 |
| Rate exceeded | 429 | `Retry-After`, `Cache-Control: no-store`, body explains condition | RFC 6585 §4 (SHOULD details, MAY Retry-After, MUST NOT cache); RFC 9110 §10.2.3 |
| Success | `next` | — | — |

### 6.3 Nginx fragment (in `cgo-guard`, not in CGO)

```nginx
# nginx/cgo-guard.conf — drop into http {} then server {}, behind TLS termination
limit_req_zone $binary_remote_addr zone=cgo_guard_reads:10m rate=30r/s;
limit_req_zone $binary_remote_addr zone=cgo_guard_mutations:10m rate=1r/s;

# rate-limited → 429, not 503
limit_req_status 429;
log_format cgo_guard '$remote_addr $request $limit_req_status';

server {
    listen 9090 ssl;
    # ...TLS certs...

    # mutations: path+method split is NOT expressible with location prefixes alone;
    # use http-level map OR duplicate zones per location (see Slice 1 note).
    location /api/ {
        limit_req zone=cgo_guard_mutations burst=3 nodelay;
    }
    location / {
        limit_req zone=cgo_guard_reads burst=10 nodelay;
    }
}
```

**Mapping notes (authoritative):**

- `limit_req_status 429` is supported by Nginx docs ("Sets the status code to
  return in response to rejected requests") (Nginx docs — `limit_req_status`).
  Current CGO config uses 503 (`deploy/nginx/cgo.conf:9`); migration to 429 is
  Slice 1 (G13). 429 semantics then match RFC 6585 §4 (SHOULD details — Nginx
  returns a generic page; adding a `proxy_intercept_errors` body requires runtime
  verification — `unverified — primary source not found` for the exact
  `error_page 429 = @rate_limited` recipe on the operator's Nginx version).
- `Retry-After` cannot be emitted by `limit_req` itself. Options: (a) accept 429
  without `Retry-After` (permitted by RFC 6585 §4 "MAY"), or (b) intercept 429
  and add `Retry-After` via `error_page` handler, or (c) move rate limiting
  in-process (`guard.RateLimit`) where `Retry-After` is trivial. (b)'s exact Nginx
  recipe is `unverified — primary source not found` for the target Nginx build.
- The reads-vs-mutations split by *method* cannot be done with `location /api/`
  alone (Nginx `location` matches path, not method). The current G12 mis-split
  is therefore architectural, not a typo. Correct splits require either:
  (i) `if` on `$request_method` inside `location /api/` to choose zone (Nginx
  docs for `limit_req` do not show an `if` example — `unverified — primary source not found`
  for the preferred idiom), or (ii) accept that **all** `/api/` traffic shares
  one zone and size it for reads (e.g., 30r/s reads, 1r/s counted only on the 13
  POSTs via Go `guard.RateLimit`). Slice 1 chooses (ii) for mutations and fixes
  the size mismatch by moving the mutation bucket to Go — see §7.

---

## 7. Slices (dependency order, ≤3 active)

Each slice is shippable independently; acceptance = the listed command/behavior
runs, not a code review.

| Slice | Goal | Changes | Acceptance (all must run green) |
|---|---|---|---|
| **1. Nginx 429 + boundary** | Rate-limit status correctness + remove G13/G12 half. | `cgo-guard/nginx/cgo-guard.conf`: `limit_req_status 429`; add `log_format` with `$limit_req_status`; keep `cgo_reads 30r/s` for SPA/health, resize `/api/` zone to **reads** rate and move mutation bucket to Go (temporary: keep 1r/s on 13 POST paths via `location`-duplication or `if` on `$request_method` — pick after a 30-min runtime spike on the operator's Nginx, recorded in the slice checkpoint). | `nginx -t` passes; synthetic burst: 5th mutation request → **429** (not 503), `GET /api/state` under normal polling never 429s; `$limit_req_status` present in access log. |
| **2. Bearer hardening (`cgo-guard` lib)** | Constant-time, snapshot-once, `realm`, exact documented parsing, loopback fix. | `guard.Bearer`: `crypto/subtle.ConstantTimeCompare`; read token once in constructor; loopback = `127.0.0.0/8` **or** `::1` via `net.ParseIP` + `IsLoopback` (fixes `::1` gap and `127/8`); reject dev-bypass when `AllowLoopbackWithoutToken=false` (default). Keep 401 + `WWW-Authenticate`. | `go test ./...` in `cgo-guard`: 401 on missing/wrong, 200 on correct (incl. loopback dev-mode off → 401); timing test (wrong-length vs wrong-value indistinguishable — `unverified` for a deterministic timing harness, so test by static analysis: `go vet` + `subtle` usage grep) is best-effort. |
| **3. CSRF fallback (`cgo-guard` lib)** | Close G1–G4. | `guard.CSRF` = `net/http.CrossOriginProtection` (`Handler`/`Check`) + explicit `same-site` policy (default: deny non-safe unless `TrustedOrigin` allowlisted), `Origin`/`Referer` double-check when `Sec-Fetch-Site` absent (OWASP §2.2 fallback), `Vary: Sec-Fetch-Site, Origin`, `LogOnly` mode (OWASP rollout). Require Go ≥1.25 (OWASP built-in note). | `go test ./...`: cross-site POST → 403, absent-header POST → 403 (fail-safe) or fallback pass only with trusted Origin; `same-site` POST → 403 by default; `Vary` present; log-only logs without blocking. |
| **4. Go rate-limit fallback** | Per-credential mutation bucket for no-Nginx hosts; 429 + `Retry-After` + `Cache-Control: no-store`. | `guard.RateLimit`: per-identity (Bearer token hash, fallback IP) fixed-window; emits 429 + `Retry-After: <seconds>` + details body; `Cache-Control: no-store` (RFC 6585 §4 MUST NOT cache). **New dependency decision required** if a token-bucket lib is imported — do not add without human approval (see §8). Default implementation: stdlib-only windowed counter (no new dep). | `go test ./...`: 13th mutation under 1r/s → 429 with `Retry-After`; after window → 200; `Cache-Control: no-store` present. |
| **5. Reads auth (policy)** | Close G11 per design. | Wrap reads with `Bearer` on non-loopback bind (bind-aware, not path-blind). Requires a bind decision (loopback-only vs. remote). | Integration test: unauthenticated `GET /api/state` on non-loopback → 401; loopback without token → 200 (dev mode only). |
| **6. Observability** | G17. | Nginx `limit_req_log_level warn`; access log includes `$limit_req_status`; Go `/metrics` exposes `cgo_auth_rejected_total{reason="bearer|csrf|ratelimit"}` (repo design — `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:60-70` metrics contract). | `docker compose` scrape shows counters increment on 401/403/429; log-only CSRF mode visible in logs. |

**Deferred (explicitly not in these slices):**

- OIDC/mTLS session auth — design S2 names it (`mTLS or OIDC bearer`) but Bearer
  adequacy for a single-operator bench is undecided. `unverified — primary source not found`
  for whether mTLS/OIDC is required by an enterprise buyer; defer until the buyer names it.
- WIS 2.0 / multi-site / CHR — out of scope per design §“Deferred” (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:82-83`).

---

## 8. Anti-toy criteria (definition of done for the independent repo)

A change that violates any of these is a toy, not a guard. Each criterion names
its check.

1. **Status-code fidelity.** 401 only for auth miss (RFC 9110 §15.5.2), 403 only for
   CSRF/forbidden (RFC 9110 §15.5.4), 429 only for rate limit (RFC 6585 §4), 503
   only for genuine overload/maintenance (RFC 9110 §15.6.4). *Check:* `guard_test.go`
   asserts each code once per cause; Nginx burst test asserts 429, never 503.
2. **Challenge on every 401.** `WWW-Authenticate: Bearer` present (and `realm` when
   configured) on every 401 response. *Check:* table test asserts header on all 401s.
   (RFC 9110 §11.6.1)
3. **Constant-time token compare, no `==`.** *Check:* `grep -r "== want" --include="*.go"`
   must be empty; `crypto/subtle.ConstantTimeCompare` present. (OWASP HMAC guidance)
4. **Token loaded once, not per request.** *Check:* `os.Getenv` count in `guard.go`
   ≤ 1 (constructor). Secrets rotation = restart/reload, documented.
5. **Loopback is `net.IP.IsLoopback`, not a string prefix.** *Check:* test with
   `::1`, `127.0.0.2`, and proxied `RemoteAddr = 127.0.0.1` + remote `XFF` ignored
   (documented behavior). (Fixes G8)
6. **Fetch-Metadata fallback is implemented, not documented.** Absent-header +
   sensitive endpoint ⇒ deny (fail-safe) **or** Origin/Referer-checked pass;
   `same-site` explicit; `Vary` present; log-only mode works. *Check:* the §7 Slice 3
   test matrix. (OWASP Fetch Metadata §1–§2 + Rollout)
7. **Rate limit is leaky-bucket honest.** Under burst: first `burst` pass immediately,
   rest delayed-to-rate **or** rejected-with-429 per `nodelay`, never silently
   queued forever and never silently dropped without a status. *Check:* Nginx
   burst test + Go window test. (Nginx docs `limit_req`; RFC 6585 §4)
8. **429 carries `Retry-After` and `Cache-Control: no-store` in Go; Nginx 429
   documented if `Retry-After` is intentionally omitted (RFC "MAY").** *Check:*
   header assertion in Go tests. (RFC 6585 §4; RFC 9110 §10.2.3)
9. **Per-identity keying is deliberate.** If key = IP, the limitation (NAT sharing)
   is written in the README with the RFC 6585 §4 quote (“does not define how …
   identifies the user”) so the buyer sees the tradeoff. *Check:* README grep for
   that phrase.
10. **No new Go dependency without human sign-off.** V1 is stdlib-only
    (`net/http`, `crypto/subtle`, `net`, `time`, `strings`). Any `x/time/rate`
    or token-bucket import is a separate approval gate (repo rule: new dependency
    requires human approval; AGENTS.md is the authority, not cited here as a primary).
11. **Rate-limit body explains the condition.** 429 response includes a
    human-readable detail ("per-identity rate exceeded, retry after N s").
    *Check:* substring assertion in tests. (RFC 6585 §4 SHOULD)
12. **Dry-run is available before enforcement.** `LogOnly` (Go) and
    `limit_req_dry_run on` (Nginx) modes let the operator measure false positives
    before blocking. *Check:* log-only test asserts no 403/429 in dry-run while
    logs show would-block. (OWASP Rollout; Nginx `limit_req_dry_run`)

---

## 9. Open / unverifiable items (do not guess these)

1. `unverified — primary source not found` — exact Bearer wire grammar and
   invalid-token error params (needs RFC 6750 fetch + pin).
2. `unverified — primary source not found` — whether `bearer` lowercase / trailing
   whitespace must be accepted (RFC 9110 §11 token rules not disambiguated for
   scheme casing without RFC 6750).
3. `unverified — primary source not found` — token entropy / length / storage /
   rotation policy (no fetched primary governs static API tokens).
4. `unverified — primary source not found` — HSTS presence/absence at runtime
   (no `Strict-Transport-Security` in repo; needs live header check).
5. `unverified — primary source not found` — Nginx `error_page 429` → `Retry-After`
   injection recipe for the operator's Nginx build.
6. `unverified — primary source not found` — chosen `if ($request_method)` vs.
   Go-bucket split for mutations under `limit_req` (needs 30-min runtime spike).
7. `unverified — primary source not found` — Go `CrossOriginProtection` exact
   `Check` internals beyond the type/method surface (pkg.go.dev index fetched;
   source bodies not fetched — pin Go version before relying on behavior).
8. `unverified — primary source not found` — whether Nginx + Go both on 9090
   loops (needs live topology; `deploy/nginx/cgo.conf:5` + `proxy_pass` lines).
9. `unverified — primary source not found` — `/metrics` scrape behavior under
   `limit_req` + auth (Prometheus contract in design §6.1(3) not exercised here).
10. `unverified — primary source not found` — enterprise buyer demand for
    mTLS/OIDC (design names it, no buyer statement in repo).

---

## Sources

- **RFC 9110** — Fielding et al., *HTTP Semantics*, IETF Standards Track, June 2022.
  §9.2.1 (safe methods), §10.2.3 (`Retry-After`), §11–§11.6.1–§11.6.2 (auth,
  challenges, credentials), §15.5.2 (401), §15.5.4 (403), §15.6.4 (503).
  https://www.rfc-editor.org/rfc/rfc9110.html
- **RFC 6585** — Nottingham & Fielding, *Additional HTTP Status Codes*, IETF
  Standards Track, April 2012. §4 (429 Too Many Requests: SHOULD details, MAY
  `Retry-After`, MUST NOT cache, user-identification undefined).
  https://www.rfc-editor.org/rfc/rfc6585.html
- **OWASP CSRF Prevention Cheat Sheet** — Introduction (token fallback doctrine),
  Token-Based Mitigation (synchronizer pattern; HMAC; constant-time), Fetch
  Metadata headers (§1 policy; §2 absence fallback; Requirements; Concerns;
  Rollout), Custom Request Headers + CORS, SameSite, Standard Headers
  (Origin/Referer), Built-In Implementations (Go `CrossOriginProtection` ≥1.25).
  https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
- **Nginx `ngx_http_limit_req_module`** — Synopsis (leaky bucket); `limit_req`
  (burst/nodelay/delay semantics; inheritance); `limit_req_zone` (key, 10m sizing,
  LRU exhaustion, r/s vs r/m); `limit_req_status` (default 503); `limit_req_dry_run`;
  `limit_req_log_level`; `$limit_req_status`.
  https://nginx.org/en/docs/http/ngx_http_limit_req_module.html
- **Go `net/http`** — pkg.go.dev `net/http` (go1.27.3, BSD-3-Clause): Overview (client/
  server), Functions (`Error`, `Handle`, `MaxBytesReader`, `StatusText`), Types
  (`Handler`, `Header.Get/Set`, `Request`, `ServeMux`, `CrossOriginProtection`
  with `NewCrossOriginProtection`/`Check`/`Handler`/`AddTrustedOrigin`),
  Constants (`MethodGet…`, `StatusUnauthorized 401 // RFC 9110, 15.5.2`,
  `StatusForbidden 403 // RFC 9110, 15.5.4`, `StatusTooManyRequests 429 // RFC 6585, 4`,
  `StatusServiceUnavailable 503 // RFC 9110, 15.6.4`).
  https://pkg.go.dev/net/http
- **Repo** — `pkg/api/auth.go:1-26`; `pkg/api/server.go:173-247,320,380,394-1280,
  1312-1500,1431-1483`; `pkg/api/auth_test.go:1-23`; `pkg/api/hardening_test.go:19-58`;
  `deploy/nginx/cgo.conf:1-35`; `docs/deploy.md:159-183`;
  `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:14,35,52,59-71,80,82-83,107,131-135`;
  `docs/superpowers/plans/2026-11-06-cgo-enterprise-hardening.md:28,38,132,148-194,507`.

---

*End of research file. Wrote one file only; no code, no commit, no frozen-data touch.*
