# CGO Supply-Chain Gaps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the supply-chain gaps spec (SLSA L2 + CycloneDX SBOM + signed OCI, per-module tags/cutover, cgo-gui separate repo, WIS2 probe deferred, eBPF observer deferred) as bite-sized, independently testable tasks.

**Architecture:** Keep the measurement core and frozen evidence untouched. Add GoReleaser SBOM generation for pure modules, signed OCI digest-pinned releases, per-module semver tags with cutover procedure, and separate `cgo-gui` repo scaffolding. No live runs, no WIS/Mosquitto/CHR/eBPF implementation — each deferred tool is design-only with a named human gate.

**Tech Stack:** Go 1.25 (stdlib only, no new Go dependency without approval), GoReleaser (syft/cosign), CycloneDX, SLSA, OCI, Nginx, Prometheus

## Global Constraints

- No live audit, bench, CHR, eBPF, `.dist`, BBRv3, RouterOS or network traffic as evidence. Tests use fixtures only.
- No mutation is unauthenticated; reads require authentication on any non-loopback bind.
- Rate limiting is enforced at Nginx ingress for all POST mutations.
- `/metrics` honors Prometheus exposition contract: `# HELP`/`# TYPE` before first sample, unique `(name,labels)`, correct `Content-Type` (`text/plain; version=0.0.4`).
- Absent is never `0`; Prometheus absent maps to “no sample”, never to `0`.
- 220 ms / 95% is labelled “operator objective”, never a WMO mandate.
- Frozen corpus (`book/data/runs`, CSV schemas) is never touched; provenance is additive only.
- No new dependency without explicit human approval. Plan adds no Go module beyond the extractions listed; GoReleaser/syft/cosign are system tools, not Go modules.
- French comments and user-facing text; English identifiers. No `git add -A`; stage explicit paths. Never merge to `main` without the human's decision.

---

## Source map

### Existing boundaries to preserve
- `go.mod` — `module github.com/Realms4239/cgo`, `go 1.25` floor
- `modules/model/`, `modules/metrics/`, `modules/results/`, `modules/figures/`, `modules/qdisc/` — pure modules already extracted (Tasks 1–4, Slice 29)
- `deploy/Dockerfile`, `deploy/docker-compose.yml`, `deploy/nginx/cgo.conf`, `deploy/prometheus/rules.yml` — Slice 28 S1/S4/S5 ops plane
- `docs/deploy.md` — S1 + S4 boundary notes
- `docs/research/2026-11-07-module-deploy-supplychain.md` + `docs/research/2026-11-07-module-airgap-distribution.md` — primary-source basis

### New boundaries
- `.goreleaser.yaml` — GoReleaser SBOM + signing config (already created in Slice 28 S6, now extended)
- Per-module `go.mod` tag procedure documented in `docs/deploy.md`
- `cgo-gui` separate repo scaffolding path documented (no code moved in this plan)

---

### Task 1: GoReleaser SBOM for pure modules

**Files:**
- Modify: `.goreleaser.yaml` (add `sboms` section for `modules/model`, `modules/metrics`, `modules/results`, `modules/figures`, `modules/qdisc`)
- Modify: `docs/deploy.md` (SLO/SBOM procedure)

**Interfaces:**
- Consumes: GoReleaser upstream (`goreleaser.com/customization/sbom/`), CycloneDX (`cyclonedx.org/capabilities/sbom/`), SLSA (`slsa.dev/spec/v1.0/requirements`)
- Produces: `sboms:` five-field pipeline per artifact (`id`, `kind`, `cmd`, `env`, `artifacts`)

- [ ] **Step 1: Add SBOM generation**

```yaml
sboms:
  - id: cgo
    documents:
      - id: cgo
        format: cyclonedx-json
        cmd: syft
    # full config from 2026-11-08 supplychain-gaps spec §17
```

- [ ] **Step 2: Verify config**

Run: `goreleaser check` (if binary present) else record config-complete in `docs/deploy.md`.
Expected: `goreleaser check` passes or documents missing binary.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yaml docs/deploy.md
git commit -m "feat(deploy): GoReleaser CycloneDX SBOM for pure modules"
```

---

### Task 2: Signed OCI image (digest-pinned)

**Files:**
- Modify: `deploy/Dockerfile`
- Modify: `deploy/docker-compose.yml` (fail-closed digest)

**Interfaces:**
- Consumes: OCI Image/Distribution specs
- Produces: `ghcr.io/realms4239/cgo@sha256:${CGO_IMAGE_DIGEST:?}` — never `:latest`

- [ ] **Step 1: Pin digest**

Verify `deploy/docker-compose.yml:3` uses `@sha256:${CGO_IMAGE_DIGEST:?set release digest}` and `deploy/Dockerfile` is `FROM scratch` with `org.opencontainers.image.*` labels.

- [ ] **Step 2: Verify compose**

Run: `docker compose -f deploy/docker-compose.yml config` (if Docker present) else record config-complete.
Expected: digest required, `:latest` rejected.

- [ ] **Step 3: Commit**

```bash
git add deploy/Dockerfile deploy/docker-compose.yml
git commit -m "feat(deploy): digest-pinned OCI image"
```

---

### Task 3: Per-module tags/cutover (SemVer)

**Files:**
- Modify: `docs/deploy.md` (add cutover procedure section)
- Create: `docs/runbooks/tag-cutover.md` (procedure checklist)

**Interfaces:**
- Consumes: SemVer 2.0.0 (`semver.org`), Go Modules canonical versions (`go.dev/ref/mod`)
- Produces: 8-step cutover (tag → release → verify → replace removal → monorepo pin)

- [ ] **Step 1: Write runbook**

Create `docs/runbooks/tag-cutover.md` with the 8 steps from spec §17.

- [ ] **Step 2: Verify procedure**

Manual walk: `git tag v0.1.0 -m "cgo/model v0.1.0"` in a dry-run clone, `go list -m all`, confirm `replace` removal is `no-op` (record result in runbook).

- [ ] **Step 3: Commit**

```bash
git add docs/runbooks/tag-cutover.md docs/deploy.md
git commit -m "docs(deploy): per-module tag/cutover procedure"
```

---

### Task 4: `cgo-gui` separate repo scaffolding

**Files:**
- Create: `docs/runbooks/gui-repo.md` (scaffolding spec)
- Create: `cmd/cgo-gui/README-SEPARATE.md` (pointer to new repo)

**Interfaces:**
- Consumes: Fyne build constraints (`//go:build linux && cgo` guard)
- Produces: separate-repo structure with no `internal/` import, stub for non-GUI builds

- [ ] **Step 1: Write scaffolding spec**

Document: new repo `cgo-gui`, `go.mod` with Fyne, `core.go` dual-view (`windows` Win32 vs `linux` Fyne), `LockOSThread`, `CGO_ENABLED=0` fallback stub, SemVer independent.

- [ ] **Step 2: Commit**

```bash
git add docs/runbooks/gui-repo.md cmd/cgo-gui/README-SEPARATE.md
git commit -m "docs(gui): cgo-gui separate repo scaffolding"
```

---

### Task 5: WIS2 probe + eBPF observer (design-only, gated)

**Files:**
- Create: `docs/runbooks/wis2-probe.md` (DEFERRED design)
- Create: `docs/runbooks/bpf-observer.md` (DEFERRED design)

**Interfaces:**
- Consumes: WMO WIS 2.0 overview, CAP 1.2 OASIS, Mosquitto `mosquitto.conf(5)`, kernel eBPF/BTF
- Produces: design docs only — no code, no broker, no loader

- [ ] **Step 1: Write deferred designs**

Document: `cgo/wis2-probe` topic tree (`cgo/v1/probe/...`), CAP mapping, Mosquitto TLS/ACL profile; `cgo-bpf-observer` Session API, BTF/CO-RE, capget discipline, 6.8 build gate. Each with `DEFERRED: needs Q1–Q7 + human gate`.

- [ ] **Step 2: Commit**

```bash
git add docs/runbooks/wis2-probe.md docs/runbooks/bpf-observer.md
git commit -m "docs(wis2/bpf): deferred probe designs"
```

---

### Task 6: Verify + closeout

**Files:**
- Modify: `docs/CHECKPOINTS.md`
- Modify: `docs/VERIFY.md` (PENDING row only)
- Modify: `docs/superpowers/plans/README.md`

- [ ] **Step 1: Run gates**

Run:

```bash
go test ./... -count=1
python tools/verify.py
```

Expected: all PASS, `verify.py` 6/6.

- [ ] **Step 2: Checkpoint + ledger**

Append to `docs/CHECKPOINTS.md`: command, date UTC, host, exit codes, unproven (CHR, WIS, bench, Mosquitto, eBPF live).
Update `docs/VERIFY.md` PENDING row (not VERIFIED — verifier owns it).
Update `docs/superpowers/plans/README.md` slice row → Pending verifier.

- [ ] **Step 3: Commit**

```bash
git add docs/CHECKPOINTS.md docs/VERIFY.md docs/superpowers/plans/README.md
git commit -m "docs: close supplychain gaps slice 30"
```

---

## Parallelization map

- Task 1 (GoReleaser) and Task 4 (GUI scaffolding) are independent — parallel allowed.
- Task 2 (OCI) depends on Task 1 (SBOM digest must match image digest).
- Task 3 (tags/cutover) depends on Tasks 1–2 (tags sign the SBOM + image).
- Task 5 (WIS2/eBPF deferred) is independent of 1–4 — design only.
- Task 6 must remain at the end.

## Risk register

| Risk | Mitigation |
|---|---|
| GoReleaser binary absent on this host | Record config-complete; first CI run with syft/cosign produces the first SBOM/signature |
| Digest pin breaks first pull | Fail-closed with clear error message in compose + runbook |
| Per-module tag released before tests green | Tag procedure requires `go test ./...` exit 0 in module dir first |
| `cgo-gui` build still fails on Linux | Build tag guard `//go:build linux && cgo` + stub; Linux CI image until S3 |
| WIS2/eBPF spec drift into code | Deferred design only; no broker/loader shipped in this plan |

## Definition of done

- GoReleaser produces CycloneDX SBOMs for the five pure modules (config-complete, CI-proven on first tag).
- OCI image digest-pinned (`:latest` rejected); rollback is one-command offline.
- Per-module semver tags + cutover procedure documented and dry-run recorded.
- `cgo-gui` separate-repo structure documented with build-constraint guard.
- WIS2 probe and eBPF observer designs present as DEFERRED with named human gates.
- All required tests and gates pass; verifier signature or explicit `withheld` recorded.
