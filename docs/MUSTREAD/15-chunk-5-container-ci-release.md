# Chunk 5 — Container, Strict CI, Signed Release & Ops

> Source: `C:\END\CGO5.md` lines 168–267. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** a working 3-stage Docker build, `--strict` gates with no exit-2 masking, pinned signed release with SBOM and SLSA L2, plus NTP, cert, log and backup basics.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** closes caps K3, K5, K9, K10 and K11.
- **Usage:** execute after Chunk 4; gates Chunk 6.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 6 — live proof and final freeze.
Container K3, Honest Strict CI K5, Pinned Signed Release Pipeline K9, and Ops
Basics K10/K11).

Hand Part A to your agent, then run Part B yourself to verify container boot,
strict CI, pinned releases, and logging/DR before moving to Chunk 6 (Live Linux
Validation & Final Freeze).

CHUNK 5 — Working Container (K3), --strict CI (K5), Signed Release (K9) & Ops (K10, K11)

PART A — Exact Bounded Prompt for Your Coding Agent

# CHUNK 5 EXECUTION TASK — CONTAINER, STRICT CI, SIGNED RELEASE & OPS (STRICTLY BOUNDED)

You are executing ONLY Chunk 5 on `C:\cgo`. Do NOT modify probe math, UI views (except build/bundle paths if needed), or `docs/CHECKPOINTS.md`.

## 1. Fix `K3` — Working 3-Stage `Dockerfile` & Clean `.dockerignore`
Files: `.dockerignore`, `deploy/Dockerfile`, `deploy/docker-compose.yml`, `cmd/cgo/main.go` (add `healthcheck` subcommand)
1. **Fix `.dockerignore` (`.dockerignore:10-11`):**
   - Do NOT ignore `web/frontend/src/`, `web/frontend/public/`, `web/frontend/package.json`, `web/frontend/bun.lock`, `web/frontend/index.html`, `web/frontend/vite.config.ts`, `web/frontend/tsconfig*.json`, `web/frontend/embed.go`, `VERSION`, `pkg/results/runs-frozen.tar.gz`, or `pkg/probe/bpf/*.o`!
   - Ignore ONLY `web/frontend/node_modules/`, `web/frontend/dist/assets/`, `web/frontend/dist/index.html`, `.git/`, `data/`, `book/`, `*.log`, `backup/`.
2. **Add `cgo healthcheck` subcommand (`cmd/cgo/main.go`):**
   - Implement a tiny `cgo healthcheck [--addr https://127.0.0.1:9090]` subcommand that dials `/api/health` (using `InsecureSkipVerify` for local self-signed loopback if `--insecure-local` passed) and exits `0` iff HTTP `200` with `"ok":true`. This lets `distroless` execute a real Docker `HEALTHCHECK` without `curl`!
3. **Rewrite `deploy/Dockerfile` as a 3-Stage Multi-Stage Build:**
   - **Stage 1 (`oven/bun:1.3.14-alpine` or pinned `oven/bun:1`):** Copy `web/frontend/package.json`, `bun.lock`, run `bun install --frozen-lockfile`, copy frontend sources, run `bun run build`.
   - **Stage 2 (`golang:1.25-bookworm`):** Copy `go.mod`, `go.sum`, `go mod download`; copy Go sources + `VERSION` + embedded corpus/eBPF `.o` + Stage 1 `web/frontend/dist`; build static `CGO_ENABLED=0` `/out/cgo`.
   - **Stage 3 (`gcr.io/distroless/static-debian12:nonroot`):** `USER 65532:65532` (non-root), `COPY --from=2 /out/cgo /cgo`, `VOLUME ["/data"]`, `EXPOSE 9090`, `HEALTHCHECK --interval=15s --timeout=3s --start-period=5s CMD ["/cgo", "healthcheck", "--insecure-local"]`, `ENTRYPOINT ["/cgo"]`, `CMD ["--serve", "--addr", "127.0.0.1:9090", "--tls=true"]`.
4. **Clean `deploy/docker-compose.yml:3-10`:**
   - Remove fake signature-verification comments, mount `./data:/data`, pass `CGO_API_TOKEN` via environment/secrets, and verify default config boots cleanly.

## 2. Fix `K5` — Stop Masking Exit 2 in `.github/workflows/ci.yml` & Upgrade Docker Smoke
Files: `.github/workflows/ci.yml:21-30, 83-84`
1. **Enforce `--strict` on Gates (`:21-30`):**
   - Pass `--strict` to `cmd/gates` in `.github/workflows/ci.yml`, and **DELETE** the script lines converting exit `2` / `frozen SKIP` into exit `0`. Note: ensure CI has access to the clean committed frozen corpus (`pkg/results/runs-frozen.tar.gz` / `docs/frozen-manifests.sha256`) so `--strict` genuinely runs and passes 6/6 with `0` skips!
2. **Real Container Boot Smoke Test (`:83-84`):**
   - Replace `docker run ... cgo version` with a real live serve verification:
     - Build the image via `docker build -f deploy/Dockerfile -t cgo:ci .`
     - Boot container in background with `--serve --expose --addr 0.0.0.0:19090 --tls=false --insecure-allow-plaintext` and `CGO_API_TOKEN=ci-smoke-test-token-32-chars-minimum`
     - Wait for `/api/health == 200`
     - Curl `http://127.0.0.1:19090/`, extract `<script src>` & `<link href>` paths from HTML, curl **every** extracted JS/CSS bundle asserting **`HTTP 200`**
     - Assert unauthenticated `GET /api/results` returns **`401`** and `Authorization: Bearer ...` returns **`200`**, then remove container.
3. Ensure CI runs: `gitleaks` (`fetch-depth: 0`), `govulncheck ./...`, `go vet ./...`, `go test -race -count=1 ./...`, `npx tsc --noEmit`, `bun run test`, `bun run build`, `check-bundle.mjs`, `npx playwright test` (incl. `live-smoke.spec.ts` + `@axe-core/playwright`), and `cmd/gates --strict`.

## 3. Fix `K9` — Pinned, Honest Release Pipeline (`.github/workflows/release.yml`, `.goreleaser.yaml`)
Files: `.github/workflows/release.yml:23,24,31,79`, `.goreleaser.yaml`, `SECURITY.md:14-15`
1. **Pin Every Tool & Action:** Replace ALL `@latest` in `.github/workflows/release.yml:23,24,31` (`syft`, `cosign`, `goreleaser`) and GitHub Actions with exact pinned version tags/commit SHAs.
2. **Remove `|| true` (`:79`):** Delete `|| true` from `cosign attest-blob` / `cosign sign-blob` so any signing failure fails the release job immediately!
3. **Real SBOM, Cosign & SLSA L2 Attestation:**
   - Generate SPDX SBOMs (`syft`), sign `checksums.txt` with Cosign (`--bundle` `.sigstore.json`), and emit GitHub SLSA Build L2 provenance via `actions/attest-build-provenance` (`permissions: id-token: write, attestations: write, contents: write`).
   - Update `SECURITY.md:14-15` with exact matching verification commands (`cosign verify-blob` + `gh attestation verify` + `sha256sum -c`).

## 4. Fix `K10` & `K11` — NTP Sync Check, Cert Expiry, Durable `slog` Audit Log & Runbooks
Files: `cmd/cgo/doctor.go`, `cmd/cgo/cert.go`, `api/session.go:22-44`, `api/server.go`, `docs/runbooks/operations.md` (new), `PRIVACY.md`
1. **NTP & TLS Lifecycle (`K10`):**
   - Add `chrony` / `timedatectl` clock-sync check to `cgo doctor` (reporting sync status explicitly and warning before eBPF/tc correlation if unsynchronized).
   - Add cert expiry check (`< 30 days` warning) to `cgo doctor` + `cgo kit tls rotate` command.
2. **Durable Audit Log & Request Correlation (`K11`, `api/session.go:22-44`):**
   - Replace the 50-entry volatile in-memory ring with an **append-only size-rotated JSON `slog` security/audit log** in `CGO_DATA_DIR/logs/audit-events.jsonl` (recording `ts`, `req_id`, `operator_id`, `method`, `path`, `status`, `resource_id`, sanitized of secrets), propagating `X-Request-ID` through HTTP -> audit/engine logs!
   - Enforce log retention alongside `cgo audit purge` matching `PRIVACY.md:57-58`.
3. **Concise Runbook (`docs/runbooks/operations.md`):**
   - Document exact recovery steps for: disk exhaustion / `cgo audit purge`, corrupt manifest (`verifyManifests` error), expired cert (`tls rotate`), NTP drift, atomic binary rollback (`cgo.prev` + `systemctl revert`), and off-disk backup/restore drill procedure.

## Constraints
- Verify `docker build` and container serve locally if Docker is available; if Docker daemon is absent on Windows host, validate syntax + verify all non-daemon commands and explicitly note Docker daemon execution awaits CI/Linux runner.
- Commit all Chunk 5 work in ONE clean commit: `ci(ops): fix multi-stage Dockerfile, strict gates, pinned cosign/sbom release and durable slog`.

PART B — Your Hard Verification Gate (Run Before Chunk 6)

Run these commands locally after the agent finishes Chunk 5:

# 1. Verify NO || true or @latest remains in workflows
git grep -n "|| true\|@latest" .github/workflows/ || echo "CLEAN_WORKFLOWS"

# 2. Verify --strict is passed to gates and exit 2 is NOT masked in ci.yml
git grep -n "\-\-strict" .github/workflows/ci.yml
git grep -n "exit 2" .github/workflows/ci.yml || echo "CLEAN_NO_EXIT2_MASK"

# 3. Verify Dockerfile has frontend build stage + non-root USER + HEALTHCHECK
cat deploy/Dockerfile

# 4. Verify .dockerignore does NOT exclude web/frontend/src or VERSION or .o
cat .dockerignore

# 5. Verify durable slog + X-Request-ID + NTP check exist in code
git grep -n "X-Request-ID" api/
git grep -n "timedatectl\|chrony" cmd/cgo/

# 6. Run tests + gates locally (--strict)
go test -race -count=1 ./...
sh .githooks/run-gates.sh

# 7. (If Docker installed locally or in CI) Test container build & boot:
docker build -f deploy/Dockerfile -t cgo:test .
# Boot and verify /, JS/CSS assets = 200, /api/results unauth = 401!

# 8. Perform ONE off-disk backup/restore drill (copy data/ to USB or separate directory/archive,
#    restore to temp dir, run `cgo` + `verify` against restored dir, log RTO, and delete stale legacy tarballs).

Once Chunk 5 passes (including clean GitHub Actions CI green on all jobs!), say
"Proceed with Chunk 6" for the final Linux CGO-DEV live validation, field trace
+ .dist, RouterOS CHR transcript, book/ + memoire.docx build, and v2.0.0 Freeze!
