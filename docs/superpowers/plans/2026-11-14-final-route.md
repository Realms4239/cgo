# CGO / Meteolink Final Route — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development`
> (recommended) or `executing-plans` to implement this plan task-by-task. Steps
> use checkbox (`- [ ]`) syntax. A very weak agent must be able to execute each
> step from its text alone: exact paths, exact commands, no judgment calls.

**Goal:** Finish CGO/Meteolink — caps remediated, measurement physics proven on a
live bench, eBPF kernel truth, real DGM field evidence, a decision-first
interface, and a signed v2.0.0 + Master 2 thesis PDF (title 2).

**Architecture:** Repair the `main` multi-module trunk first (stub SPA, facade
metrics, unwired duplicates), then security/delivery, measurement physics, eBPF,
field evidence, interface, thesis. Evidence comes from a passive eBPF observatory
(A) + an LXD container bench (B) on the sadcnwp host; BBRv3 stays BLOCKED.

**Tech Stack:** Go 1.25 (multi-module), React 19 + Vite + Zustand + ECharts,
cilium/ebpf (CO-RE, CI-built), iproute2/netem, LXD, QEMU (CHR), GitHub Actions,
goreleaser + syft + cosign. Spec: `docs/superpowers/specs/2026-11-14-final-route-design.md`.

## Global Constraints (every task inherits these)

- Trunk = `main` multi-module; reconcile against the spec's finding→status map before fixing.
- Frozen 159-run corpus byte-identical; quarantine via sidecar; new v2 corpus separate.
- No compiler on the host/VMs; eBPF objects CI-built + sha256-pinned; `-race` on CI.
- LXD safety envelope (spec §4): no host mutation/reboot/auto-start; `network=none`; bounded; rehearsed teardown; least privilege.
- sadcnwp: NO REBOOT; ftp2 untouched; fresh same-day R2 tokens; pseudonymize; `safe_read.py`/`safe_write.py`.
- Secrets never echoed/committed; never touch `book/data/runs`, `backup/`, `data/`, `epbfBBR.md`, remote PC `100.114.101.13`, Graylog.
- Ledgers append-only, never frozen. French user-facing; English specs/plans/gates.
- `make` absent → `python tools/verify.py` (named, never waived); exit 2 ≠ green.
- Builder ≠ verifier; no VERIFIED without a fresh re-run; evidence runs serialized on a quiet host.
- Sequential by default; parallel only when provably shared-no-state.

---

## Phase 0 — Trunk repair & containment

### Task 0.1: Disk triage (day-0, prerequisite)

**Files:** none (environment only).
**Interfaces:** Produces: ≥10 GB free on C:, Go caches on D:.

- [ ] **Step 1: Measure** — Run: `df -h /c /d`. Record free space.
- [ ] **Step 2: Move reproducible caches to D:** set `GOCACHE=/d/gocache-cgo`,
  `GOTMPDIR=/d/gotmp-cgo`; move npm/bun/Playwright caches to D:. Delete ONLY
  reproducible caches, never user data.
- [ ] **Step 3: Verify** — Run: `df -h /c` → ≥10 GB free; `go env GOCACHE` → D:.
- [ ] **Step 4: Commit** the env note to the checkpoint (no code).

### Task 0.2: sadcnwp R0 bench-feasibility read (day-0, human-gated)

**Files:** `docs/verification/<stamp>-sadcnwp-r0.md` (create).
**Interfaces:** Produces: a written bench-feasibility verdict (B usable? privileged needed?).

- [ ] **Step 1:** Obtain a fresh R0 token (human). Through `safe_read.py`, read
  `lxc list`/`lxc info`, `test -r /dev/kvm && test -w /dev/kvm` (as sadc), `uptime`,
  `test -s /sys/kernel/btf/vmlinux`, `df -h /home`. One scope line; evidence + sha256.
- [ ] **Step 2:** Write the verdict: LXD initialized? KVM ACL grants sadc? load
  acceptable? BTF present? Record privileged-vs-unprivileged decision.
- [ ] **Step 3: Verify** — evidence sidecar re-verifies; verdict committed.

### Task 0.3: Stub SPA repair (the deepest regression)

**Files:**
- Modify: `api/server.go:18,1231` (import + SPA fallback)
- Delete: `api/frontend/` (stub)
- Modify: `web/frontend/embed.go`, `cmd/cgo/server.go:295`, `cmd/cgo/setup.go:86,315`
- Modify: `kit/test-fresh-clone.sh`
- Test: `api/frontend_test.go` (create)

**Interfaces:**
- Consumes: `web/frontend.FS` (`//go:embed all:dist`).
- Produces: `api` serves the real SPA; `ValidateEmbeddedAssets() error`.

- [ ] **Step 1: Write the failing test**

```go
// api/frontend_test.go
package api

import "testing"

func TestEmbeddedAssetsPresent(t *testing.T) {
    if err := ValidateEmbeddedAssets(); err != nil {
        t.Fatalf("embedded SPA missing: %v", err)
    }
}
```

- [ ] **Step 2: Run it to verify it fails** — Run: `go test ./api/ -run TestEmbeddedAssetsPresent -count=1`.
  Expected: FAIL (`ValidateEmbeddedAssets` undefined).
- [ ] **Step 3: Wire the real embed** — in `api/server.go`, replace the
  `api/frontend` import with `web/frontend`; serve `web/frontend.FS` at `/`.
  Add `ValidateEmbeddedAssets()` in `web/frontend/embed.go` asserting
  `dist/index.html` AND ≥1 `dist/assets/*.js` exist in the embed.FS; call it at
  server startup. Delete `api/frontend/`.
- [ ] **Step 4: Run the test to verify it passes** — Run: `go test ./api/ -run TestEmbeddedAssetsPresent -count=1`. Expected: PASS.
- [ ] **Step 5: Fix setup** — `setup.go:86` `needsBuild` checks real index.html +
  ≥1 asset; `--yes` actually builds (`bun install --frozen-lockfile && bun run
  build`, fallback `npm ci && npm run build`); delete the `runServe` no-op (`:315`).
- [ ] **Step 6: Rewrite the fresh-clone gate** — `kit/test-fresh-clone.sh`:
  clone to temp → `make build` → serve `127.0.0.1:19090` with EXIT trap → curl
  `/`, extract asset URLs, assert 200 on `/` + every JS/CSS + `/api/health`.
- [ ] **Step 7: Verify in a disposable clone** — Run: `bash kit/test-fresh-clone.sh`.
  Expected: exit 0, all assets 200.
- [ ] **Step 8: Commit** — `git add api/server.go web/frontend/embed.go cmd/cgo/setup.go kit/test-fresh-clone.sh api/frontend_test.go && git rm -r api/frontend && git commit -m "fix(api): serve the real embedded SPA, delete the stub, gate the fresh clone"`.

### Task 0.4: Duplicate-module disposition (deletion test)

**Files:** delete `guard/`; modify `api/server.go:399` + `api/metrics.go` (wire
`exporter/`, delete facade); delete-or-wire `cite/`; wire `attest/` into verify.
**Interfaces:** Produces: one auth seam (`api/auth.go`), honest `/metrics`.

- [ ] **Step 1: Auth** — confirm `api/auth.go` is the wired seam; `git grep guard`
  → only `go.mod` replace; delete `guard/` + its replace line. Run `go build ./...` → exit 0.
- [ ] **Step 2: Metrics failing test** — write `api/metrics_test.go` asserting
  `/metrics` emits a measured series and never a hardcoded `0` flag. Run → FAIL.
- [ ] **Step 3: Wire exporter** — replace the `api/metrics.go` facade with
  `exporter/` at the `/metrics` seam; re-key `deploy/prometheus/rules.yml`. Run → PASS.
- [ ] **Step 4: Cite** — `git grep` for a server-side consumer of `cite/`; if none,
  delete it (record the deletion test); else wire. Build green.
- [ ] **Step 5: Attest** — wire `attest/` into `cgo verify`; a mutated seal fails
  the named check (test). Run → PASS.
- [ ] **Step 6: Commit** — explicit paths, one logical group per module.

### Task 0.5: Cap-40 purge (human force-push gate)

**Files:** none (git history). **Interfaces:** Produces: gitleaks = 0 across all refs.

- [ ] **Step 1 (human):** revoke/rotate the PAT; export token-use logs outside the repo.
- [ ] **Step 2:** `git clone --mirror <origin> /d/cgo-mirror-backup.git`.
- [ ] **Step 3:** on the mirror, `git filter-repo --invert-paths --path book/handoff-livrables-2026-09-02.md --force`.
- [ ] **Step 4:** `gitleaks detect --source .` → 0 findings across full history.
- [ ] **Step 5 (human gate):** user approves `git push --force --all` + `--tags`; origin GC; `gh api` fork check.
- [ ] **Step 6:** land the remote CI secret-scan gate (with Task 1.3).
- [ ] **Step 7: Verify** — `git rev-list --all` no longer reaches the blob; CI scan green.

### Task 0.6: Identity, version SSoT, dead-surface cuts

**Files:** `book/content/00-front.js`, `PRODUCT.md`, `README.md`,
`web/frontend/index.html`, `VERSION`, `web/frontend/package.json`, `docs/api.md`,
`tools/verify.py`; delete `cmd/cgo-gui`, `npm/`, dead `.woff2`, orphan config lines.
**Interfaces:** Produces: title 2 + single VERSION verified by the gate.

- [ ] **Step 1:** set title 2 across the surfaces (human sign-off); lock name METEOLINK.
- [ ] **Step 2:** VERSION single source (`//go:embed VERSION`); add a `verify.py`
  version-parity check (VERSION == package.json == api.md).
- [ ] **Step 3:** delete `cmd/cgo-gui`, `npm/`, 3 dead `.woff2`, orphan
  `.gitattributes:3`/`.gitignore:3-4`; **keep `cmd/cgo-chr`**.
- [ ] **Step 4:** triage the 37 product TODOs (implement/delete/ledger); purge
  `C:/cgo` from `pkg cmd web` only, never ledgers.
- [ ] **Step 5: Verify** — `python tools/verify.py` version-parity PASS;
  `go build ./cmd/cgo ./cmd/meteolink` exit 0; `git grep "C:/cgo" -- pkg cmd web` → 0.
- [ ] **Step 6: Commit** — explicit paths.

### Task 0.7: Ledger & gate hygiene

**Files:** `docs/superpowers/plans/README.md`, `tools/verify.py:202`, `Makefile`,
`web/frontend/scripts/check-bundle.mjs`, `web/frontend/vite.config.ts`.

- [ ] **Step 1:** update the plan index row (do NOT freeze ledgers).
- [ ] **Step 2:** add `-count=1` to `verify.py:202`; add `--strict` (fail on SKIP).
- [ ] **Step 3:** fix phantom gates — `test-real` selects real `-tags=real` tests
  or is dropped; `bpf-check` compares the CI-built object + sha256.
- [ ] **Step 4:** `vite.config.ts` `manualChunks:{echarts:['echarts']}`;
  `check-bundle.mjs` honest budget (TOTAL_GZ_MAX=400, ECHARTS_GZ_MAX=350).
- [ ] **Step 5: Verify** — `python tools/verify.py --selftest` → 6 caught; `--strict` PASS.
- [ ] **Step 6: Commit.**

**Phase 0 gate (verifier re-runs):** fresh clone serves real SPA (assets 200);
gates green `-count=1`; gitleaks 0; one auth seam; honest `/metrics`; title 2 +
version SSoT; reconciliation map committed.

---

## Phase 1 — Security & delivery

### Task 1.1: Exposure lockdown
**Files:** `cmd/cgo/main.go:113`, `cmd/cgo/server.go`, `api/auth.go`, `api/server.go`, `docs/deploy.md`.
**Acceptance:** unauth `/api/*`→401; non-loopback `--tls=false` without `--expose`+token→exit 2; browser SSE cookie→200; headers present; burst→429.
Steps: default loopback+TLS unit → startup guard → auto auth.token → two-adapter auth (Bearer + HttpOnly cookie login) → complete observeBlocked → timeouts → headers → rate limit → deploy.md CIDR. Each: failing test → implement → run → commit.

### Task 1.2: SSRF / root-CA / sudoers / guest hardening
**Files:** `probe/ssrf.go` (create), `internal/vm/tls.go:21`, `cmd/cgo/cert.go`, `kit/testbed.sh:57`, `kit/guest-setup.sh`, `kit/host-readiness.sh`, `kit/host-tunnel.ps1`, `kit/fix-vnet-admin.ps1`.
**Acceptance:** SSRF to RFC1918→403; root-CA install needs SPKI confirm; cgo-netctl refuses out-of-scope; Tar-Slip refused; `host-readiness.sh --require-linux` exits 2 on Windows.

### Task 1.3: CI/CD
**Files:** `.github/workflows/ci.yml` (create), `.github/workflows/release.yml`, `.goreleaser.yaml`.
**Acceptance:** CI green on push; tag → signed archives + all-module SBOM + provenance; fresh-host `cosign verify-blob` + `sha256sum -c` + `gh attestation verify` pass. Frozen gate = named host-only skip. Human gate: workflow permissions.

### Task 1.4: Retention & governance
**Files:** `cmd/cgo/` (purge), `SECURITY.md`, `PRIVACY.md`, `docs/config.md`.
**Acceptance:** purge keeps last N (temp-dir test); docs exist; links gate PASS.

### Task 1.5: Docker hardening
**Files:** `deploy/Dockerfile`, `deploy/docker-compose.yml`, `deploy/nginx/cgo.conf`.
**Acceptance:** `docker build` exit 0; non-root; loopback+TLS+token; nginx cert paths valid.

---

## Phase 2 — Measurement physics (container RUN evidence)

### Task 2.1: Probe fidelity
**Files:** `probe/probe.go:128-138,56`, `probe/bulk.go:36-38`, `probe/bulkflows.go:53-55`, `cmd/cgo/testbedsrv.go:57`.
**Acceptance:** 500→absence; write-error→error; ping chronological; container `/small`=200.

### Task 2.2: Channel physics (Q2, GE, .dist)
**Files:** `pkg/campagne/dualns.go:56-60`, `qdisc/ge.go:79-83`, `qdisc/dist.go`, `pkg/campagne/campagne.go:639`.
**Acceptance:** container P2≈100 ms; GE emit `p*100…%`, live loss≈π; `.dist` loads (TC_LIB_DIR<128).

### Task 2.3: Qdisc parity (Q4) + dead code
**Files:** `qdisc/qdisc.go:91,94,107`, `qdisc/stats.go`, `remediate/remediate.go:170-171`.
**Acceptance:** CAKE flags==bench; 4 root-fallback branch tests; `tc -j -s` parsing; ExecRunner `-tags=real` on container.

### Task 2.4: IFB + lab scripts
**Files:** `kit/run-dualns.sh:41-45`, `kit/cgo-bench-dualns.sh`, `remediate/remediate.go:199-209`, `kit/stress_testbed.sh`.
**Acceptance:** `down()` 0 orphans; ingress mismatch refused; `bash -n` clean.

### Task 2.5: Results B1–B4 + one verdict
**Files:** `results/results.go:216-224`, `api/server.go:846-851`, `metrics/stats.go:108`, `metrics/verdict.go`, delete `web/frontend/src/lib/health.ts`.
**Acceptance:** Best None-state + (profile,direction); full-corpus hash; sorted median; MW tie; `<95% Critical`; one engine verdict.

### Task 2.6: Durability / budgets / seal parity
**Files:** `campagne/writer.go`, `audit/writer.go`, `api/server_audit.go:177`, `pkg/campagne/bulksource.go`, `cmd/cgo/runcli.go:70-100`, `api/server.go:240`.
**Acceptance:** kill -9 recovers; `both` up→down; HTTP dual seals; R1 measured asymmetric.

### Task 2.7: BBRv3 gate + provisioner (engine; live BLOCKED)
**Files:** `capabilities/ccavail.go:15-22`, `model/types.go:85`, `kit/install-bbr3.sh`.
**Acceptance:** gate accepts provisioned bbr3 / refuses stock-without-provenance; honest 3×2 fallback; live matrix BLOCKED.

---

## Phase 3 — Kernel truth (eBPF)

### Task 3.1: TCP_INFO proxy first
**Files:** `probe/tcpinfo_linux.go:35-43`. **Acceptance:** exact `tcpi_bytes_retrans`; ×1448 error bound computed.

### Task 3.2: C-source corrections + CI build
**Files:** `probe/bpf/probe.bpf.c`, `probe/bpf/gen.go`, `Makefile` (bpf-check), `.github/workflows/ci.yml`.
**Acceptance:** kfree_skb+QDISC_DROP; fentry XOR kprobe; ringbuf 16 MiB; srtt/cwnd/ssthresh; netns filter; CI object+sha256; bpf-check PASS. Human gate: cilium/ebpf dep.

### Task 3.3: Loader + ladder honesty
**Files:** `probe/bpf/loader_linux.go`, `loader_other.go`, `status.go:11-25`, `kernel_bridge.go`.
**Acceptance:** E1 ebpf only on full success; E4 window deltas; overflow counted; container attaches exclusively.

### Task 3.4: Cross-validation
**Acceptance:** container retrans≈tcpi_bytes_retrans; drops≈`tc -s`; srtt≈ping; overhead <2% under flood; sealed.

### Task 3.5: Passive observatory (A)
**Files:** deployment on sadcnwp (scoped CAP_BPF). **Acceptance:** read-only real-traffic metrics; miner-excluded; pseudonymized; "compromised host" label; production untouched. Human gate: R2 token.

### Task 3.6: Report integration
**Files:** `probe/bpf/kernel_bridge.go:12`, `audit/report.go`. **Acceptance:** sealed report carries real kernel telemetry when attached; honest tier+reason otherwise.

---

## Phase 4 — Field evidence

### Task 4.1: Passive field trace
**Acceptance:** ≥30 sealed samples on the host (sadc), authorized target, CAP_NET_RAW R2 grant, sidecar. Human gate.

### Task 4.2: .dist calibration
**Files:** `audit/calibrator.go`, `qdisc/dist.go`, `api` (twin/calibrate). **Acceptance:** sealed .dist; container KS/quantile match; fidelity figure.

### Task 4.3: New v2 corpus
**Acceptance:** clean dualns matrix in container; separate anchor; quarantine sidecar; frozen corpus untouched.

### Task 4.4: CHR real transcript
**Files:** `cmd/cgo-chr/` (fix C1/C2/C3/C5), `kit/chr/chr-image.sha256`, `docs/validation/chr-v7-transcript.txt`.
**Acceptance:** real CHR boot; `proof.json accepted:true` from real export; transcript committed.

### Task 4.5: Countersigned case study
**Acceptance:** operator-mediated service map; operator-applied remediation (R2); before/after p95 + Ariary; DSI countersign. Human gates.

### Task 4.6: VERIFY rows
**Acceptance:** verifier moves .dist/field/CHR/case-study rows; BBRv3 stays BLOCKED.

---

## Phase 5 — Interface synthesis

### Task 5.1: 2-mode shell
**Files:** `web/frontend/src/App.tsx`, `views/CockpitView.tsx`, `views/LaboView.tsx` (create from verified components). **Acceptance:** mode switch; root-persistent SSE; tests green.

### Task 5.2: Rings + event-driven flush
**Files:** `web/frontend/src/lib/ringBuffer.ts` (create), `lib/sse.ts`, `api/sse.go`. **Acceptance:** Float64Array NaN=unmeasured; 20 ms coalescing; TCP_NODELAY; stale≠live badge.

### Task 5.3: Cockpit decision-first
**Acceptance:** verdict→money figure→prescription→mechanism→seal footer; one engine verdict.

### Task 5.4: Labo
**Acceptance:** unclipped P1–P4; Statique/Empirique .dist; Egress/Ingress/both; DGM twin; honest seals.

### Task 5.5: Honesty + surgical additions
**Acceptance:** F1–F18 fixed; badge/MGA-h/SLA `<95% Critical`/Copy-RouterOS present.

### Task 5.6: WCAG + lexicon + ArmButton
**Acceptance:** 44px; contrast ≥4.5; forced-colors; chart data-table; OFL/NOTICE; populated lexicon; ArmButton on 8 mutations.

### Task 5.7: Real tests
**Acceptance:** testing-library interactions; unmocked Playwright 390/890/1920; deterministic SSE; TUI sync; 0 flakes quiet.

---

## Phase 6 — Thesis rewrite & freeze (MAJOR work; content extremely stale)

The thesis source (`book/content/*.js`) describes the OLD direction; this route
makes its core claims false ("no field campaign", "eBPF not compiled", "no CHR",
legacy single-flow numbers, old 4-mode UI, monolith). Phase 6 is a substantial
rewrite around title 2 + the seven contributions, not a gap-fill. `book/` is
untracked — local work, never committed; the PDF is the deliverable. Source of
truth: `book/content/*.js` → `build-docx.js` → memoire.docx.

### Task 6.1: Thesis audit & reconciliation map
**Files:** `docs/verification/<stamp>-thesis-reconciliation.md` (create).
**Acceptance:** every stale claim inventoried (old title, "no field campaign", "eBPF not compiled", "no CHR", legacy numbers, old UI, monolith, G7 uninstrumented, dualns "not executed") → status → rewrite task.

### Task 6.2: Re-frame around title 2 + seven contributions
**Files:** `book/content/01-intro.js`, `015-organisme.js`, `06-conclusion.js`.
**Acceptance:** intro/problem re-framed around real-time QoS optimization + kernel observability + the decision loop (measure→grade→remediate→validate→re-measure→prove ms+Ariary); seven contributions foregrounded; positioned vs Flent/RRUL/perfSONAR/Smokeping (Flent cited); M2 critical-reflection descriptors.

### Task 6.3: Method chapter rewrite
**Files:** `book/content/03-partie2.js`, `docs/thesis_trace_driven_methodology.md`.
**Acceptance:** trace-driven .dist (Mahimahi); GE 4-param Markov math; eBPF CO-RE precision (kfree_skb reason, srtt/cwnd/ssthresh) + ×1448→tcpi_bytes_retrans proxy error bound; G0–G7 (G7 instrumented); honesty algebra (absent≠zero, stale≠live); quarantine taxonomy (46% as finding).

### Task 6.4: Results chapter rewrite
**Files:** `book/content/04-pratique.js`.
**Acceptance:** NEW v2 corpus numbers (clean dualns 4-flow) with per-cell run_id/seal; legacy labeled historical; .dist KS fidelity figure; eBPF three-witness cross-validation; BBRv3 BLOCKED fence; countersigned before/after (ms+Ariary); offloads-era finding; GE×100 + Q2 discoveries as critical reflection.

### Task 6.5: Annexes rewrite
**Files:** `book/content/09-annexes.js`, fresh captures.
**Acceptance:** Annexe E = new 2-mode Cockpit/Labo + fresh captures; Annexe F = kernel telemetry compiled+attached, CHR booted (real evidence, not "not yet"); SVG-fingerprint truth; v2 corpus schema + quarantine sidecar.

### Task 6.6: Front matter + citations
**Files:** `book/content/00-front.js`, `07-refs.js`, `08-glossaire.js`, `doc-final-build.cjs`.
**Acceptance:** title 2; abstract/résumé rewritten for the new direction; keywords+sigles updated (eBPF/CO-RE/BTF/LXD); per-cell citations; 159 not 160; ISGEI MSTI-STR alignment (Annexe D).

### Task 6.7: Build, audit, freeze
**Files:** `book/tools/build-book.ps1`, `book/tools/audit-*.py`, figures.
**Acceptance:** PDF builds (Word COM); audit-archive/biblio/renvois/nombres green; figures regenerated from v2 corpus (Go-generated, hash8); v2.0.0 tag; signed release verified on fresh host; maintenance archive; final VERIFY/checkpoint; honest BLOCKED (BBRv3); human merge.

---

## Self-review (against the spec)

- **Spec coverage:** every spec §5 phase has tasks; closure matrix §9 maps each
  KL/defect to a phase. BBRv3 live, field-validity limits stay BLOCKED (§10).
- **Placeholder scan:** Phase 0 carries full code/commands; Phases 1–6 specify
  exact files + acceptance; the literal implementation code for those tasks is
  written at execution against the spec's design (each task is fully specified,
  not a TBD).
- **Type consistency:** `ValidateEmbeddedAssets() error`, `web/frontend.FS`,
  `exporter` at the `/metrics` seam, `tcpi_bytes_retrans` — names match the spec.
- **Execution:** subagent-driven-development (fresh subagent per task + two-stage
  review) recommended; executing-plans for inline checkpoints.