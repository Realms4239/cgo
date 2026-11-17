# CGO Target System

> Source: `C:\END\CGO0.md` lines 1–286. Lens: **result** — the zero-bloat target architecture and the 7-chunk route.

## At a glance

- **Result:** one Go binary, one embedded React cockpit, one eBPF probe, one verification stack, zero dead shells.
- **Mechanism:** amputation list + keep/make-sharp list + Chunk 0→6 dependency chain.
- **Contribution:** a single legible work tree instead of nginx/prometheus/module/GUI sprawl.
- **Usage:** read before any bloat-removal or architecture decision.
- **Inherited constraints:** `01-core-directive.md` §1–2 (iron laws and execution loop).
You are 100% right to take the steering wheel back.

Look at why your agent stalled at 35/100 across 1,780 commits: every time it
faced a problem, it added more moving parts (deploy/nginx/cgo.conf,
deploy/prometheus/rules.yml, empty modules/, cmd/meteolink shim, cmd/gates 1,368
lines, cmd/cgo-gui, toy synctest_test.go, broken Dockerfiles, || true
workflows)—and then failed to wire any of them cleanly.

Let's strip away all of that bloat right now and design the simplest, fastest,
cleanest, most thesis-worthy architecture possible for Météo Madagascar +
ISGEI—then break the exact fixes into small, verifiable Chunks (Chunk 0 →
Chunk 6) that you inspect and apply step by step.

PART 1 — The Ultra-Clean, Zero-Bloat Target System (METEOLINK)

Reading back through the 8 roadmap.sh PDFs without taking any bloat, here is
what a truly ingenious Single-Appliance Observatory looks like:

What We DELETE Permanently (Zero Value, Pure Drag)

1.  deploy/nginx/cgo.conf: Gone. Go's stdlib http.Server already serves
    TLS 1.2+, gzip, security headers, rate limits, and //go:embed SPA natively
    in one binary. Nginx only created port conflicts and broke GET polling with
    1 r/s limits.
2.  deploy/prometheus/rules.yml: Gone (or reduced to optional /metrics endpoint
    only). On a single operator workstation / lab VM, Prometheus + Alertmanager
    rules with zero runbooks are pure enterprise theater.
3.  Empty modules/ dir & "12 modular tools" doc fiction: Gone.
4.  cmd/meteolink/ (54-line shim) & cmd/cgo-gui/ (unused fyne v2.8.1): Gone. One
    single binary (cgo / meteolink)—not 6 func main entrypoints!
5.  api/synctest_test.go toy clock test & docs/CHECKPOINTS.md churn: Gone.
    Replace with real testing/synctest tests on pkg/audit and pkg/api/sse.
6.  Multi-tenant SaaS illusion: This is a single-operator workstation / lab
    appliance, not a multi-user cloud SaaS. Keep AuthN + simple operator_id
    ownership + explicit --dev-no-auth flag—zero bloated user/DB subsystem.
7.  meteolink.dev OS hosts-file hijacking (setup.go): Bind cleanly to
    127.0.0.1:9090 (or localhost:9090).

What We KEEP & Make Razor-Sharp (Enterprise + Academic Perfection)

| Roadmap Pillar          | Clean Zero-Bloat Implementation in METEOLINK                                                                                                                                                                                                                                                                              | Why Both Météo Madagascar & ISGEI Love It                                                                                             |
| :---------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------ |
| **UX & Product**        | **Exact existing React + Zustand + ECharts + `tokens.css` UI**, locked to **2 crisp modes: `Cockpit` + `Labo`**. Zero bloat views, `60s` (`10/40/10`) everywhere, single-CTA empty state, 3 executive cards (`Santé du lien`, `Alertes WIS 2.0 < 220 ms`, `Gaspillage Yas Ar/h`), strict `HasValidRemediation` invariant. | A meteorologist reads the link verdict in 5 seconds; zero jargon on landing; no fake zeros or ghost scripts.                          |
| **Real-Time Speed**     | **Kernel 6.8 `eBPF BPF_MAP_TYPE_RINGBUF`** (`SKB_DROP_REASON_QDISC_DROP` + `tcp_retransmit_skb` + `srtt_us`) → **Go `O(1)` Priority SSE Ring (`4096` slots, `20 Hz`)** → **Frontend `Float64Array` `TypedMetricRing` (`NaN` = absent) + `60 FPS` Canvas + `RingBufferHUD`**.                                              | Zero-copy kernel events straight to glass-smooth browser rendering (`0` GC stutter). Literally connects Linux Kernel 6.8 to live UX\! |
| **Networking**          | **4-netns `dualns` bench (`KVM VM` + `ip netns`)**: single unified RTT halving (`50+50 = 100 ms`), **Ingress `ifb0`**, **Gilbert–Elliott fading**, **`.dist` empirical calibration (`calibrator.go`)**, **CAKE / `fq_codel` / FIFO × CUBIC / BBR**, and verified **RouterOS v7 CAKE (`fasttrack`-safe)** scripts.         | Zero fake `0` on failed `tc -j` scrapes (`K4`), real `TC_LIB_DIR` `.dist` replay from actual Madagascar traces, real RouterOS proof.  |
| **Security & Delivery** | Clean `git filter-repo` history (`0` PATs), loopback+TLS default, `owner_id` AuthZ + non-colliding IDs, unified `Dialer.Control` SSRF guard, regex-checked `?iface=`, **1 multi-stage `Dockerfile`**, **1 `--strict` GitHub Actions CI**, **1 pinned `Cosign + Syft + SLSA L2` release workflow**.                        | Every single hard cap (`35`, `40`, `45`) vanishes cleanly.                                                                            |
| **Blazing Test Stack**  | **Go `testing/synctest`** (`<3s` Go suite), **`bun test` + `happy-dom`** (`<4s` JS suite), **1 real unmocked Playwright + `@axe-core/playwright` check**.                                                                                                                                                                 | Total local verification runs in **seconds**, not 7 minutes.                                                                          |

PART 2 — Your Controlled, Step-by-Step Execution Plan (Chunk 0 → Chunk 6)

We do this one small Chunk at a time. You verify each Chunk's command output
before moving to the next—no agent running wild for 800 commits.

CHUNK 0 — Secret Purge (K1) & Radical Bloat Removal

Goal: Lift Cap 40 PAT permanently and delete all dead clutter so the repo
becomes small and legible.

  - Mod 0.1 (Human — GitHub UI): Revoke the leaked ghp_ PAT (54d3f62). Audit
    GitHub Security/Actions/Release logs (2026-09-02 → today) and save locally
    outside git. Rotate any keys in local untracked .env.
  - Mod 0.2 (Git History Rewrite):
    git clone --mirror <origin-url> /d/cgo-mirror-backup.git
    git filter-repo --invert-paths --path book/handoff-livrables-2026-09-02.md --force
    git log --all -S'ghp_'   # MUST return empty (0 commits)
    git push --force --all && git push --force --tags
  - Mod 0.3 (Delete Bloat Files):
      - Delete deploy/nginx/, deploy/prometheus/, empty modules/, cmd/cgo-gui/,
        cmd/meteolink/, api/synctest_test.go (we write the real synctest tests
        in Chunk 4!), unused fyne in go.mod (go mod tidy).
      - Keep cmd/cgo as the one binary (Makefile / .goreleaser.yaml build only
        cgo + gates test tool if kept—or fold verify into cgo verify and delete
        cmd/gates/ bloat too!).
  - Mod 0.4 (Unify Title, Version & 127.0.0.1 Default):
      - Lock VERSION (1.4.0 or 2.0.0) across VERSION, web/frontend/package.json,
        README.md:303, docs/api.md:14.
      - Set canonical title in README.md, PRODUCT.md, book/, memoire.docx.
      - Change default bind in cmd/cgo/main.go:60-68 from meteolink.dev:9090 (OS
        hosts hijack) to 127.0.0.1:9090.
  - Verify Chunk 0: git log --all -S'ghp_' is empty; go build ./cmd/cgo
    succeeds; go mod tidy clean.

CHUNK 1 — Security Lockdown & Clean Contracts (K2, K7, A03, Timeouts)

Goal: Lift the second Cap 40 (K2 AuthZ), fix SSRF (K7), fix command injection
(A03), and fix server error/timeout hygiene.

  - Mod 1.1 (K2 AuthZ, Collision & Revocation — api/auth.go, api/server.go,
    cmd/cgo/main.go):
      - Change ID generation (cmd/cgo/main.go:237, api/server.go:1601) from
        audit-<unix> → audit-<unix>-<4_random_hex_bytes>.
      - Derive owner_id (HMAC/SHA256 of operator token) & role (admin /
        operator) in api/auth.go, embed owner_id inside every created audit/run
        manifest, and enforce owner_id == caller.ID || caller.Role == "admin" in
        requireAuth (api/server.go:419-426).
      - Remove silent unauthenticated loopback bypass (api/auth.go:45-59) unless
        --dev-no-auth is explicitly passed on CLI.
      - Replace deterministic HMAC(token, "cgo-session-v1") (:16-26) with a
        random per-boot signing key + server-side session revocation set on
        /api/auth/logout.
      - Either implement real confirmation tokens (X-Confirm-Action nonce) on
        api/server.go:658,683,705 OR remove the "two-step confirmation"
        overclaim from PRODUCT.md:46,97 (recommendation: keep ArmButton UI +
        server X-Confirm-Action: true header on destructive POSTs—5 lines of
        Go!).
  - Mod 1.2 (K7 Unified SSRF + A03 Injection):
      - Update api/ssrf.go:47-77 using netip.Prefix covering ALL reserved
        IPv4/IPv6 ranges (100.64.0.0/10, 192.0.0.0/24, 198.18.0.0/15,
        224.0.0.0/4, 240.0.0.0/4, fec0::/10, 2001:db8::/32).
      - Move net.Dialer.Control into shared pkg/probe/ssrf.go and pass that same
        single dialer into BOTH HTTP SmallObject AND TCP
        probe/bulk_cc_linux.go:14-27 (eliminating the double-DNS TOCTOU race and
        unguarded bulk dialer!).
      - In api/server_audit.go:363-372 & pkg/remediate/remediate.go:85-102:
        reject any iface not matching ^[a-zA-Z0-9_.-]{1,15}$ with HTTP 400.
  - Mod 1.3 (Timeouts, writeJSON, Unifying 60s & 4 KiB Docs):
      - Buffer JSON in writeJSON (api/server.go:1550-1556) before calling
        w.WriteHeader(status).
      - Set WriteTimeout: 30s on http.Server, clearing it selectively inside SSE
        handlers via
        http.NewResponseController(w).SetWriteDeadline(time.Time{}). Add
        context.WithTimeout to /api/replay/stream, /api/report/export,
        /api/figures/regen, modprobe_linux.go:15 (exec.CommandContext), and
        expose.go:47 (net.Resolver).
      - Unify default audit duration to 60 s (10/40/10) across
        api/server.go:106, audit/audit.go, and PRODUCT.md:27 (adjust ProbePing /
        SmallObject VSAT timeout probe/probe.go:98-106 to max(8s, 4*delay)).
      - Document 4 KiB /login body cap cleanly in docs/api.md:3.
  - Verify Chunk 1: Unit tests verifying cross-owner 403, unauth 401, SSRF
    blocking 100.64.0.1 + raw TCP dial, ?iface=eth0;id returning 400, and
    writeJSON error returning 500.

CHUNK 2 — Data Honesty, Integrity Verification & Lab Physics (K4, K6, K8, Topology)

Goal: Eliminate every false-zero / fake-seal / RTT contradiction in the Go
engine and verification corpus.

  - Mod 2.1 (K4 — Stop Turning Failed Scrapes into 0):
      - pkg/campagne/prod.go:76-86 + pkg/qdisc/stats.go:133-139: When PollStats
        (tc -j) fails, return (Stats{}, err) / set DropsMeasured = false and
        StatsMeasured = false, marking the cell degraded/invalid—never pass nil
        to SumDrops(nil) yielding 0!
      - pkg/results/csvrows.go:19-44: Track and surface malformed CSV row count
        instead of silently dropping corrupted rows.
  - Mod 2.2 (K6 + K8 — Real /api/integrity & Clean Corpus Manifests):
      - Wire verifyManifests() inside api/corpus.go:71-73, 99-109 so
        /api/integrity checks that every file listed in manifest.json physically
        exists on disk and matches its SHA-256 hash.
      - Fix
        docs/verification/2026-11-16-v2-corpus/run-1794853778/manifest.json:9-12
        (either generate/commit valuation.json or re-seal cleanly without orphan
        reference).
      - Sync docs/verification/2026-11-12-ebpf-probe-bringup/README.md:12 to
        active eBPF hash (8612f578…) and prepare to commit the live eBPF-vs-tc
        -s JSON/capture artifact (VERIFY.md:43) in Chunk 6.
  - Mod 2.3 (Fail-Closed Topology & Single RTT Regime):
      - pkg/campagne/topology.go:34-53 (CheckBenchTopology): Return an explicit
        error when bench-topology.json is absent (fail closed!), and validate
        rtt_mode, delay_ms, and rate_mbps in addition to shape_mode.
      - Standardize RTT convention across
        kit/cgo-bench-dualns.sh:48-51, 232-238, kit/run-dualns.sh:76-78,
        pkg/campagne/dualns.go:64-71, and
        docs/verification/2026-11-15-bench-vm/README.md:95: profile DelayMs =
        total path RTT; each hop (w-edge, c-wan) ALWAYS applies DelayMs / 2
        one-way!
      - Default cmd/cgo/server.go:180-186 to dualns (matching spec).
  - Verify Chunk 2: Mocked failing PollStats produces DropsMeasured: false
    (never 0 drops); /api/integrity fails when any manifest file is deleted from
    a temp test dir; CheckBenchTopology fails when bench-topology.json is
    missing.

CHUNK 3 — Lightning-Fast Real-Time Pipeline (eBPF Ringbuf → 20 Hz Go SSE → Float64Array UI)

Goal: Deliver glass-smooth 20 Hz → 60 FPS throughput & latency streaming +
pretty RingBufferHUD inside the clean 2-Mode UI (Cockpit + Labo).

  - Mod 3.1 (Backend eBPF + 20 Hz Priority SSE Ring):
      - Verify pkg/probe/bpf/ emits real-time BPF_MAP_TYPE_RINGBUF events
        (SKB_DROP_REASON_QDISC_DROP, tcp_retransmit_skb, srtt_us) scoped to
        target flow/ifindex (cilium/ebpf).
      - Upgrade pkg/api/sse.go to a 4096-slot O(1) priority circular array
        streaming both Audits and Bench runs at 20 Hz (50 ms coalescable tick
        frames + guaranteed milestone event frames: SLA_BREACH_220MS,
        QDISC_DROP_BURST, CAKE_ENGAGED, DUAL_SEAL_COMMITTED), removing the 2,000
        ms polling lag in store/audit.ts:88-89!
  - Mod 3.2 (Frontend TypedMetricRing + RingBufferHUD + Polish):
      - Add web/frontend/src/lib/ringBuffer.ts: Float64Array circular
        TypedMetricRing (NaN = absent, zero 0-fills, zero GC) + DomainEventRing
        (100 slots).
      - Bypass React on 20 Hz SSE ticks (sse.ts writes straight into
        TypedMetricRing; rAF updates ECharts Canvas with animation: false; KPI
        readouts sync at 4 Hz with tabular-nums).
      - Drop in crisp RingBufferHUD.tsx (radial buffer fill + 20 Hz cadence +
        eBPF drops + live event feed) inside CockpitView.
      - Fix ArchiveTable.tsx:48 pluralization, lock exact pins in
        web/frontend/package.json (^ removed matching bun.lock), and bundle OFL
        + ECharts Apache-2.0 NOTICE in public/licenses/.
  - Verify Chunk 3: Live UI streams 20 Hz smooth ticks with 0 React re-render
    stalls and 0 0-substitutions on missing frames.

CHUNK 4 — Blazing-Fast, Honest Tests (synctest, bun/happy-dom, Live Playwright + Axe)

Goal: Turn a slow, partially-mocked suite into a <10s truthful safety net.

  - Mod 4.1 (Real Go testing/synctest Suite — < 3 s):
      - Delete the clock-only api/synctest_test.go toy test and write real
        synctest.Test unit tests for pkg/audit (10/40/10 60s phases + stall
        timeouts) and pkg/api/sse (64-subscriber admission & backpressure
        drops).
  - Mod 4.2 (Fast Frontend Tests — < 4 s):
      - Configure Vitest (or bun test) to run pure logic/store tests in node and
        DOM tests in happy-dom (instead of jsdom everywhere).
  - Mod 4.3 (K12 — Real Unmocked Playwright E2E + @axe-core/playwright):
      - Keep mocked layout specs separate, and add
        web/frontend/e2e/live-smoke.spec.ts booting the real compiled Go server
        (unmocked /api/**), verifying auth login, live audit flow, RingBufferHUD
        updates, and running @axe-core/playwright WCAG 2.2 AA scan!
  - Verify Chunk 4: go test -race -count=1 ./... < 5s; frontend tests < 5s; live
    Playwright + axe-core pass 0 critical/serious violations.

CHUNK 5 — Clean Multi-Stage Container, Strict CI, Real Release Pipeline & Ops (K3, K5, K9, K10, K11)

Goal: Lift Cap 35 (K3), stop CI exit-2 cheating (K5), eliminate || true signing
(K9), and harden operations.

  - Mod 5.1 (K3 — Working 3-Stage deploy/Dockerfile & .dockerignore):
      - .dockerignore: keep web/frontend/src, package.json, bun.lock (ignore
        dist/ and node_modules/).
      - deploy/Dockerfile:
        1.  Stage 1 (oven/bun:1): bun install --frozen-lockfile && bun run build
        2.  Stage 2 (golang:1.25-bookworm): copies Go src + Stage 1 dist, builds
            static cgo
        3.  Stage 3 (gcr.io/distroless/static-debian12:nonroot, USER 1000:1000):
            copies cgo, sets HEALTHCHECK CMD ["/cgo", "healthcheck"] (or
            curl/wget hitting /api/health), ENTRYPOINT ["/cgo"].
      - Clean up deploy/docker-compose.yml (remove fake verification comments,
        mount ./data:/data).
  - Mod 5.2 (K5 — Honest Strict CI .github/workflows/ci.yml):
      - Pass --strict to cmd/gates and delete the bash snippet converting exit
        2/SKIP to 0 (.github/workflows/ci.yml:21-30).
      - Update CI docker smoke (.github/workflows/ci.yml:83-84): run container
        --serve, wait for /api/health == 200, curl / + extracted JS/CSS chunks
        verifying HTTP 200!
      - Add gitleaks (fetch-depth: 0) + govulncheck + go test -race -count=1 +
        Playwright live-smoke + axe-core.
  - Mod 5.3 (K9 — Real Signed Release .github/workflows/release.yml):
      - Pin exact release tags/SHAs for syft, cosign, goreleaser (zero @latest).
      - Delete || true (release.yml:79), emit SPDX SBOM (syft), sign
        checksums.txt with Cosign (--bundle), and emit genuine GitHub SLSA Build
        L2 provenance (actions/attest-build-provenance).
  - Mod 5.4 (K10, K11 — Ops Basics, No Bloat):
      - Add timedatectl/chrony NTP sync check + TLS cert expiry warning (<30d)
        to cgo doctor.
      - Replace 50-entry volatile memory log (api/session.go:22-44) with an
        append-only rotated JSON slog audit log + X-Request-ID correlation IDs.
      - Add minimal docs/runbooks/operations.md (alert/error recovery + cgo.prev
        rollback + retention cgo audit purge), and run 1 off-disk backup/restore
        drill (separate volume/folder, not same Disk 0!).
  - Verify Chunk 5: docker build + docker run --serve returns 200 on / &
    bundles; ci.yml passes --strict cleanly; restore drill log saved.

CHUNK 6 — Live Linux Proof, Field Trace (16:22Z), .dist Calibration, CHR & Final Freeze

Goal: Close every empirical proof cleanly, update the thesis, cut the release
tag, and STOP.

  - Mod 6.1 (Linux Lab Proof on KVM / Upgraded 4-vCPU CGO-DEV):
      - Run clean dualns matrix (P1–P4 × {pfifo_fast, fq_codel, cake} × {cubic,
        bbr[, bbr3]} across Egress and Ingress IFB) + DGM Twin (WRF bulk vs
        WIS 2.0 CAP 16KB <220ms).
      - Capture & commit the 226+ event eBPF-vs-tc -s JSON confrontation
        artifact (VERIFY.md:43) and live gemodel verification log.
  - Mod 6.2 (Live Field Trace + .dist Calibration):
      - Capture the authorized ≥30-sample field trace (16:22Z window), run
        calibrator.go to emit verified .dist + SHA-256 sidecar, and run dualns
        replay with TC_LIB_DIR .dist verified live!
  - Mod 6.3 (Real RouterOS v7 CHR Proof):
      - Archive clean RouterOS v7 CHR transcript (/queue tree export + /ip
        firewall filter print proving fasttrack compatibility) matching
        pkg/remediate.
  - Mod 6.4 (Write book/ + C:\cgo\memoire.docx & Freeze):
      - Now—and only now—write book/ and C:\cgo\memoire.docx using the canonical
        title and exact verified numbers from Chunks 0–6.
      - Tag v2.0.0 (or v1.4.0), verify 100% green CI + Cosign/SBOM/SLSA release
        assets, download & verify from scratch, compile thesis PDF, and STOP.

