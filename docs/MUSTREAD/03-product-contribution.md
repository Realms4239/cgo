# CGO Product Contribution

> Source: `C:\END\CGO1.md` lines 1–237. Lens: **contribution** — the problem, the personas, the value proposition and the architecture brief.

## At a glance

- **Result:** a decision-first observatory for Météo Madagascar and a defensible master's thesis for ISGEI.
- **Mechanism:** 60 s stethoscope, 4-netns digital twin, dual-sealed evidence export.
- **Contribution:** names the operational failure (bufferbloat on a constrained link) and the exact product boundaries that prevent overclaiming.
- **Usage:** read when a requirement is ambiguous — the product definition wins.
- **Inherited constraints:** `01-core-directive.md` §1–2 and `02-target-system.md` §1.

PART 1 — The Reworked Narrative & Product Definition

(Grounded in the Product Manager, Software Architect, and System Design
Roadmaps)

1.1 For Whom, What Problem, and Why It Actually Matters to Météo Madagascar

  - Primary User (Persona 1 — Operator / DSI Technician at Météo Madagascar):
    Operates a regional center or headquarters link (Fibre, 4G/LTE, VSAT, or
    LEO) where two incompatible workloads share one constrained WAN bottleneck
    (the classic System Design Noisy Neighbor antipattern):
    1.  Heavy Asynchronous Transfers: WRF / GRIB2 forecast outputs and EUMETSAT
        satellite tiles (50–200 MB bulk TCP flows).
    2.  Life-Critical Real-Time Telemetry: WMO WIS 2.0 / CAP cyclone warnings
        and PULSONIC automatic weather station updates (2–28 KB
        latency-sensitive transactions with a 220 ms operational deadline).
  - The Real Operational Failure: Standard ISP routers use unmanaged FIFO
    (pfifo_fast) queues. The instant a bulk weather file starts downloading or
    uploading, TCP fills the queue buffer (Bufferbloat). Throughput still reads
    18 Mbit/s on iperf, and idle ping still reads 35 ms—yet real-time CAP
    cyclone alerts stall past 1,500–3,000 ms, stations drop offline, and TCP
    retransmissions waste metered Yas Business GOLD / Platine capacity (Ar/h)
    twice.
  - Secondary Reader (Persona 2 — ISGEI MSTI-STR Master 2 Jury): Evaluates
    scientific rigor: Linux Kernel 6.8 eBPF observability
    (SKB_DROP_REASON_QDISC_DROP), deterministic 4-netns kernel emulation (IFB
    ingress + .dist empirical calibration + Gilbert–Elliott wireless fading),
    AQM × Congestion Control (FIFO / fq_codel / CAKE × CUBIC / BBR), atomic
    crash resilience (JIRAMA power cuts), and end-to-end cryptographic
    provenance (SHA-256).

1.2 The Value Proposition (METEOLINK in 3 Sentences)

METEOLINK is a single static Go binary (with an embedded React real-time cockpit
and Linux 6.8 eBPF probe) that runs non-intrusively on an operator workstation
or lab VM—requiring zero database, zero cloud servers, and zero external agents:

1.  Diagnose in 60 Seconds (Cockpit): Runs a safe 60 s (10s idle · 40s load ·
    10s drain) stethoscope test against an authorized target or twin, streaming
    20 Hz kernel telemetry into a zero-jitter Ring Buffer HUD to answer: Do
    critical < 220 ms alerts survive heavy load, and how many Ariary/hour are
    wasted to congestion?
2.  Prove the Fix Safely (Labo — $0 Hardware): Reproduces the site's exact link
    class (P1 Fibre, P2 4G, P3 VSAT, P4 LEO, or an empirical .dist calibrated
    from field traces) inside a 4-namespace Linux Digital Twin—including ifb0
    downlink shaping—proving side-by-side how CAKE keeps alerts < 50 ms without
    starving bulk transfers.
3.  Export Verified Router Remediation + Sealed Proof: Only when real load and
    real congestion are measured (HasValidRemediation), generates ready-to-apply
    MikroTik RouterOS v7 CAKE (fasttrack-safe) and Linux tc/IFB scripts
    alongside a dual-sealed (SHA-256 CSV + Manifest) evidence bundle.

Explicit Product Boundaries (What METEOLINK Refuses to Claim)

  - No Unauthorized Production Probing: Never scrapes or stresses
    41.207.49.27:9100 or live DGM servers; external tests run strictly against
    explicitly authorized endpoints or the local Digital Twin.
  - No Fake Remote Magic: Never claims a browser click shapes an upstream ISP or
    remote router; shaping requires Linux CAP_NET_ADMIN locally or pasting the
    generated RouterOS v7 config onto the site's edge router.
  - No Multi-Tenant Cloud SaaS: Single-appliance workstation tool with clean
    operator_id ownership & token security—no Postgres, no microservices, no
    Nginx reverse-proxy circus, no fake "two-step confirm" claims unless
    actually enforced in code.
  - Strict "Absent Is Never Zero" Law: If a probe or tc scrape fails, the UI and
    sealed manifest record Mesure absente (null / NaN /
    DropsMeasured=false)—never 0.

PART 2 — Master Architectural & Simplification Design Brief

(The strict rules the coding agent is NOT allowed to violate during Chunks 0–6)

2.1 Radical Work-Tree Cleanup (Reject All Bloat)

Before fixing anything else, the work tree is pruned to only what actively runs
and adds value:

  - DELETE permanently:
      - deploy/nginx/ (broken port/rate-limit theater — Go's stdlib http.Server
        already handles TLS 1.2+, gzip, rate limits, security headers, and
        static assets natively).
      - deploy/prometheus/ (unmapped alert rules with no runbooks — keep only
        the minimal Go /metrics handler if needed, or drop Prometheus complexity
        altogether; structured JSON slog + SSE covers observability cleanly).
      - modules/ (empty folder contradicting "12 modular tools" doc fiction).
      - cmd/cgo-gui/ + fyne v2.8.1 (zero imports, broken Linux cross-build,
        excluded from releases).
      - cmd/meteolink/ (redundant 54-line re-exec shim — ship one binary: cgo).
      - api/synctest_test.go (fake test of Go's clock — replaced with real
        testing/synctest tests of pkg/audit and pkg/api/sse).
      - Orphan docs/verification/ dead files (valuation.json mismatch, stale
        b6de79ac eBPF README, old CHECKPOINTS.md churn), 18 legacy backup
        tarballs (after 1 off-disk restore test), and meteolink.dev OS
        hosts-file hijacking (cmd/cgo/setup.go).
  - SIMPLIFY cmd/gates: Keep only lean, strict checks (frozen, provenance,
    manifest integrity, hygiene, doc/route parity) with --strict always on in CI
    (exit 2 = FAIL, never converted to 0).
  - KEEP & CONSOLIDATE:
      - 1 Go binary (cmd/cgo) + 1 embedded React/TS/ECharts/Zustand UI
        (web/frontend) + 1 Linux dualns kit (kit/cgo-bench-dualns.sh +
        run-dualns.sh) + 1 multi-stage deploy/Dockerfile + 1 CI workflow
        (.github/workflows/ci.yml) + 1 signed release workflow
        (.github/workflows/release.yml).

2.2 Core Security & Integrity Invariants (Lifts Every Cap & Critical Bug)

1.  Zero Secrets in History (K1): ghp_ PAT revoked on GitHub + git filter-repo
    scrubbed from all refs/tags + gitleaks enforced on full history in CI.
2.  Owner-Scoped AuthN/AuthZ + Non-Colliding IDs (K2):
      - Loopback default 127.0.0.1:9090 with generated token (or explicit
        --dev-no-auth flag—never silent bypass!). --expose requires TLS + Bearer
        token.
      - Audit/Run IDs use audit-<unix>-<4hex> (zero 1-second collision).
      - Token maps to (operator_id, role); every created audit/run stores
        owner_id; requireAuth checks ownership (owner_id == caller.ID || role ==
        "admin").
      - Random per-boot session HMAC key + /api/auth/logout revocation list.
      - Real X-Confirm-Action: true guard on destructive mutations (PRODUCT.md
        parity) + writeJSON buffers JSON before sending 200.
3.  Unified SSRF & Injection Guards (K7, A03):
      - Single pkg/probe/ssrf.go net.Dialer.Control using netip.Prefix covering
        all reserved IPv4/IPv6 ranges (100.64/10, 192.0.0/24, 198.18/15, 224/4,
        240/4, fec0::/10, 2001:db8::/32), eliminating double-DNS TOCTOU and
        guarding both HTTP SmallObject and raw TCP bulk_cc_linux.go!
      - ?iface= validated strictly against ^[a-zA-Z0-9_.-]{1,15}$ (400 on
        invalid) before hitting pkg/remediate.
4.  Honest Zeros & Real Integrity (K4, K6, K8):
      - pkg/campagne/prod.go + qdisc/stats.go: PollStats failure returns error /
        DropsMeasured = false, never SumDrops(nil) = 0. Malformed CSV rows
        quarantined (csvrows.go), not silently dropped.
      - /api/integrity runs full verifyManifests() checking every declared file
        exists on disk AND matches its SHA-256 hash.
5.  Single Topology & Physics Contract:
      - CheckBenchTopology (topology.go) fails closed when bench-topology.json
        is absent.
      - One RTT convention everywhere: Profile DelayMs = total path RTT;
        dualns.go and cgo-bench-dualns.sh both apply DelayMs / 2 per hop (50
        + 50 = 100 ms).
      - 60 s (10/40/10) unified across CLI, API default, and PRODUCT.md.
        WriteTimeout enabled (30s) with http.ResponseController exemption for
        SSE. exec.CommandContext on sudo calls.

2.3 Triple-Ring High-Speed Telemetry Pipeline

1.  Kernel 6.8 eBPF (pkg/probe/bpf/): cilium/ebpf CO-RE (vmlinux.h + bpf2go)
    hooking tracepoint/skb/kfree_skb (SKB_DROP_REASON_QDISC_DROP),
    raw_tp/tcp_retransmit_skb, and exclusive fentry/kprobe tcp_rcv_established →
    pushing exact skb->len drop/RTT events via BPF_MAP_TYPE_RINGBUF
    (target-filtered to exclude SSH/Tailscale noise), with clean TCP_INFO
    fallback on non-root/Windows.
2.  Go O(1) Priority SSE Ring (pkg/api/sse.go): Fixed 4096-slot circular array
    streaming 20 Hz (50 ms) coalescable tick frames + guaranteed milestone event
    frames (SLA_BREACH_220MS, QDISC_DROP_BURST, CAKE_ENGAGED,
    DUAL_SEAL_COMMITTED), eliminating 2s audit polling lag.
3.  Frontend Zero-GC Engine (src/lib/ringBuffer.ts): Pre-allocated Float64Array
    TypedMetricRing (NaN for missing telemetry—zero 0 fabrication!) + 100-slot
    DomainEventRing. SSE writes straight to TypedMetricRing bypassing React
    renders; rAF updates ECharts Canvas (animation: false) cleanly at 60 FPS.

2.4 Simplified 2-Mode Operator UI Brief (Keep Framework, Kill Complexity)

Preserve React 19 + Zustand + ECharts + tokens.css—strictly locked to 2
destinations:

1.  Cockpit (Default / — Zero Jargon, Operator-First):
      - Top bar: METEOLINK · Météo Madagascar · Telemetry Badge ([● eBPF
        Noyau 6.8 · 20 Hz] or [TCP_INFO]) · [ Cockpit | Labo ].
      - Action bar: Target selector (Lien Autorisé / DGM Twin · WRF vs WIS 2.0)
        + [ ▶ Lancer l'audit (60 s) ] + 3-phase bar (Repos 10s → Charge 40s →
        Vidange 10s).
      - Verdict section: Idle = 1 welcoming CTA card (zero "Pas encore mesuré"
        wall). Measured = 3 crisp cards (Protection WIS 2.0 < 220 ms, Latence
        sous charge, Capacité Gaspillée Ar/h using Yas GOLD 1,388.89 Ar/h).
      - Live Observatory (RingBufferHUD + Canvas): Sleek radial ring gauge
        (head/cap, 20 Hz, kernel drops) + smooth Canvas waveform with 220 ms CAP
        threshold + plain-French live event feed (DomainEventRing).
      - Remediation card (Gated): Rendered ONLY when HasValidRemediation =
        (BulkMbps > 0.5 && DeltaP95Measured && RTTMeasured)—exporting verified
        RouterOS v7 CAKE (fasttrack-safe) / Linux tc IFB configs behind
        ArmButton + X-Confirm-Action.
2.  Labo (Digital Twin & Preuves):
      - Unclipped full-text P1–P4 + DGM Twin tiles, explicit Canal (Statique vs
        Empirique .dist) + Direction (Egress WAN vs Ingress IFB vs Les deux),
        220 ms default SLA, before/after comparison, and an honest Seal
        Inspector driven directly by real /api/integrity verification.

2.5 Blazing-Fast Verification, Clean Docker & Real Release Brief

  - Tests: Go testing/synctest on pkg/audit & pkg/api/sse (<5s Go suite); bun
    test / happy-dom (<5s JS suite); 1 real unmocked Playwright +
    @axe-core/playwright E2E test booting the real Go server (0 critical/serious
    WCAG 2.2 AA violations, proper French pluralization in ArchiveTable.tsx).
  - Container (K3): .dockerignore keeps web/frontend/src, package.json,
    bun.lock; 3-stage deploy/Dockerfile builds frontend (oven/bun) → Go binary
    (golang:1.25-bookworm) → distroless:nonroot (USER 1000:1000 + HEALTHCHECK);
    CI boots container --serve and verifies HTTP 200 on /, JS/CSS bundles, and
    /api/health.
  - CI & Release (K5, K9): ci.yml passes --strict to gates (zero exit 2 -> 0
    masking); release.yml pins exact SHA/versions for syft/cosign/goreleaser,
    removes || true, attaches SPDX SBOM + Cosign bundle + GitHub SLSA Build L2
    provenance.
  - Ops (K10, K11): cgo doctor checks chrony/timedatectl NTP sync + cert expiry
    (<30d); durable rotated JSON slog log with X-Request-ID; 1 off-disk
    backup/restore drill.

PART 3 — Master Plan Overview: The 7 Agent-Proof Chunks

Because the coding agent is weak when given broad open-ended goals, we will feed
it one tight, isolated Chunk at a time, with explicit file allowlists, explicit
forbidden actions, and mandatory verification commands that YOU (or I reviewing
the diff) verify before proceeding:

| Chunk       | Scope & Blocker Closed                                                                                 | Files Touched (Bounded\!)                                                                                                                                                                                                                                                | Hard Gate Before Advancing                                                                                                                                       |
| :---------- | :----------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Chunk 0** | **Human Gate (`K1`) + Bloat Amputation + SSoT**                                                        | Git history (`filter-repo`), delete `deploy/nginx`, `deploy/prometheus`, `modules/`, `cmd/cgo-gui`, `cmd/meteolink`, `api/synctest_test.go`; update `VERSION`, `go.mod`, `cmd/cgo/main.go` (`127.0.0.1`), title docs                                                     | `git log --all -S'ghp_'` = empty; `go build ./cmd/cgo` passes; single binary + clean tree                                                                        |
| **Chunk 1** | **Security Lockdown & Clean Contracts (`K2`, `K7`, `A03`, Timeouts)**                                  | `api/auth.go`, `api/server.go`, `api/server_audit.go`, `api/ssrf.go`, `pkg/probe/ssrf.go`, `pkg/probe/bulk_cc_linux.go`, `pkg/remediate/remediate.go`, `pkg/campagne/modprobe_linux.go`, `cmd/cgo/expose.go`, `docs/api.md`, `PRODUCT.md`                                | Cross-owner `403`, unauth `401`, SSRF `100.64.0.1` + raw TCP blocked, `?iface=eth0;id` = `400`, `writeJSON` error = `500`                                        |
| **Chunk 2** | **Data Honesty, Real Integrity & Topology Physics (`K4`, `K6`, `K8`, RTT)**                            | `pkg/campagne/prod.go`, `pkg/qdisc/stats.go`, `pkg/results/csvrows.go`, `api/corpus.go`, `pkg/campagne/topology.go`, `pkg/campagne/dualns.go`, `kit/cgo-bench-dualns.sh`, `kit/run-dualns.sh`, `cmd/cgo/server.go`, `pkg/probe/probe.go`, verification READMEs/manifests | Failed `tc -j` sets `DropsMeasured=false` (never `0`); `/api/integrity` catches missing `valuation.json`; `CheckBenchTopology` fails closed; unified `DelayMs/2` |
| **Chunk 3** | **Triple-Ring Real-Time Engine (`eBPF` → `20 Hz` SSE → `Float64Array` + `RingBufferHUD`) & 2-Mode UI** | `pkg/probe/bpf/`, `pkg/api/sse.go`, `pkg/audit/`, `web/frontend/src/lib/ringBuffer.ts`, `sse.ts`, `CockpitView.tsx`, `LaboView.tsx`, `RingBufferHUD.tsx`, `ArchiveTable.tsx`, `package.json`, `public/licenses/`                                                         | `20 Hz` SSE stream verified; `Float64Array` (`NaN` = absent, `0` fake zeros); 2 clean modes, `60s` default, `HasValidRemediation`, `OFL`/`NOTICE` shipped        |
| **Chunk 4** | **Blazing-Fast Tests (`synctest`, `happy-dom`/`bun`, Live E2E + Axe)**                                 | `pkg/audit/*_test.go`, `pkg/api/sse_test.go`, `web/frontend/vitest.config.ts`, `web/frontend/e2e/live-smoke.spec.ts`                                                                                                                                                     | `go test -race -count=1 ./...` `< 5s`; JS tests `< 5s`; live unmocked Playwright + `@axe-core/playwright` passes `0` serious violations                          |
| **Chunk 5** | **Clean 3-Stage Docker, `--strict` CI, Signed Release & Ops (`K3`, `K5`, `K9`, `K10`, `K11`)**         | `.dockerignore`, `deploy/Dockerfile`, `deploy/docker-compose.yml`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `cmd/cgo/doctor.go`, `api/session.go`, `docs/runbooks/operations.md`                                                                     | `docker build` + `docker run --serve` returns `200` on `/`, JS/CSS, `/api/health`; CI runs `--strict` (`exit 2` fails); pinned release tools (no \`              |
| **Chunk 6** | **Live Linux Proof, Field Trace, `.dist`, CHR Transcript & Thesis Freeze**                             | `CGO-DEV` live runs, `docs/verification/` (`eBPF vs tc -s` JSON, `.dist` sidecar, `chr-v7-transcript.txt`), `book/`, `C:\cgo\memoire.docx`, tag **`v2.0.0`**                                                                                                             | 6/6 `--strict` gates + live Linux `-race` + field trace + `.dist` + CHR verified; `v2.0.0` Cosign/SBOM/SLSA green; thesis compiled & repo **FROZEN**             |

Review & Sign-Off Before Chunk 0

Look over this Design Brief carefully right now:

1.  Do you approve deleting the bloat list (deploy/nginx/, deploy/prometheus/,
    modules/, cmd/cgo-gui/, cmd/meteolink/, api/synctest_test.go, meteolink.dev
    hosts squat)?
2.  Do you confirm keeping the single-operator workstation + Digital Twin
    (Cockpit + Labo) narrative and 60 s (10/40/10) unified audit standard?

Once you say "Design brief approved — write Chunk 0", I will produce the exact,
copy-pasteable Chunk 0 execution spec (commands + agent guardrails +
verification checks) so you execute Chunk 0 cleanly and inspect every single
change before moving to Chunk 1!
