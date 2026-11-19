# CGO / Meteolink Chunk 5 — Strict CI & Signed Release Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop masking exit-2 in CI, replace the `cgo version` Docker smoke with a real served-container HTTP check, pin all release tools to exact versions, and remove `|| true` from signing.

**Architecture:** `.github/workflows/ci.yml` runs the six-gate runner with exit 2 treated as failure; the docker job boots the real image and curls `/`, bundle URLs and `/api/health`. `.github/workflows/release.yml` installs exact-pinned syft/cosign/goreleaser and fails closed on signing errors.

**Tech Stack:** GitHub Actions YAML, `sh .githooks/run-gates.sh`, Docker, Docker Compose, syft/cosign/goreleaser.

**Source:** `docs/MUSTREAD/15-chunk-5-container-ci-release.md` §2–3 (verbatim corpus `CGO5.md` 76–93). Chunk 5 is decomposed; the multi-stage Dockerfile already exists and is out of scope.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- **No secret, token, live bench, remote host, or production access in CI changes.** Test tokens are literals scoped to the workflow only; never logged.
- The frozen corpus (`book/`) is gitignored, so `frozen` legitimately SKIPs in a fresh CI clone. Exit 2 must NOT be converted to 0 — record the named condition instead of faking green. A human decision is required to commit corpus fixtures or change that gate; do not do it here.
- Never touch `book/`, frozen schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, frontend, or Go business logic.
- Tool versions must be exact tags/SHAs (no `@latest`), recorded verbatim. If a version cannot be verified from the repo, STOP and ask rather than guess.
- English workflow text; French comments where the file already uses French.
- One commit, exact message: `ci(release): fail closed on gates and pin signed release tools`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Strict CI gates

**Files:** modify `.github/workflows/ci.yml`

**Interfaces:** consumes `.githooks/run-gates.sh` exit contract (0/1/2).

- [ ] **Step 1: Remove the exit-2 mask**

Replace the `set +e ... if [ $rc -ne 0 ] ... exit 0` block at `:21-31` with:

```yaml
      - name: gates
        run: sh .githooks/run-gates.sh --strict
```

If `--strict` is not supported by the runner, use `sh .githooks/run-gates.sh` with NO exit-code rewriting, and add a comment that a named `frozen` SKIP is reported honestly (exit 2), never green.

- [ ] **Step 2: Verify** — `sh .githooks/run-gates.sh` locally; record its exact exit code and named skips. Do not claim green if any gate skipped.

---

### Task 2: Real Docker smoke

**Files:** modify `.github/workflows/ci.yml` (`docker-build` job at `:77-84`)

**Interfaces:** consumes the image built by `docker build -f deploy/Dockerfile`.

- [ ] **Step 1: Replace the version smoke**

Replace `docker run --rm --entrypoint /cgo cgo-ci version` with a served-container check using a test token and a loopback port:

```yaml
      - name: smoke (served container)
        run: |
          docker run -d --name cgo-ci-smoke -p 19090:9090 \
            -e CGO_API_TOKEN=ci-smoke-test-token-32-chars-minimum \
            cgo-ci --serve --addr 0.0.0.0:9090 --tls=false
          for i in $(seq 1 30); do
            curl -fsS http://127.0.0.1:19090/api/health && break
            sleep 1
          done
          curl -fsS http://127.0.0.1:19090/ -o /tmp/cgo-index.html
          grep -o '/assets/[^" ]*' /tmp/cgo-index.html | while read -r asset; do
            curl -fsS "http://127.0.0.1:19090${asset}" > /dev/null
          done
          code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:19090/api/results)
          test "$code" = 401
          docker rm -f cgo-ci-smoke
```

Keep `if: always()` cleanup if the job already uses it. Note: the loopback `--addr` with `--expose` rules may require `--expose` for non-loopback binds — verify against `cmd/cgo/expose.go` and adjust the command minimally rather than disabling the guard.

- [ ] **Step 2: Verify** — `docker build` may be unavailable on this host; if so, record BLOCKED with exact error and validate YAML syntax only. Do not claim a green container smoke.

---

### Task 3: Pinned release tools and fail-closed signing

**Files:** modify `.github/workflows/release.yml`

**Interfaces:** consumes existing release steps at `:23,24,31,79`.

- [ ] **Step 1: Pin tools**

Replace each `@latest` with an exact pinned tag verified from the repo's lockfile or an approved version recorded in the plan; no guessed versions. Use the SAME version for all occurrences of a tool.

- [ ] **Step 2: Remove the signing bypass**

Delete `|| true` from the cosign signing line at `:79` so a signing failure fails the job. Keep SBOM/provenance steps unchanged.

- [ ] **Step 3: Verify** — `grep -n '@latest\||| true' .github/workflows/release.yml` → EMPTY; YAML parses. Record any version that could not be verified as BLOCKED, not guessed.

---

### Task 4: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–3.

- [ ] **Step 1: Checks** — YAML parse of both workflows; static greps; local gate run with honest status.
- [ ] **Step 2: Commit once** — stage only `.github/workflows/ci.yml` and `.github/workflows/release.yml` by explicit path; message `ci(release): fail closed on gates and pin signed release tools`. If a gate fails, STOP and report.
- [ ] **Step 3: Record** — verbatim outputs and limitations to the SDD report; explicitly record that frozen-corpus SKIP in CI remains a named condition requiring a human decision. No `docs/CHECKPOINTS.md` append.
