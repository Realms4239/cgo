# CGO Core Directive

> Source: `C:\END\CGO.md` lines 1–237. Lens: **mechanism** — how this project is governed, executed, debugged and reviewed.

## At a glance

- **Result:** one binding execution constitution for every agent session.
- **Mechanism:** seven iron laws, a 5-step execution loop, a 3-tier speed loop, a 5-layer root-cause playbook and a 30-second reviewer protocol.
- **Contribution:** removes the failure modes the corpus names (`|| true`, exit-2 masking, `SumDrops(nil)=0`, clock-only tests, broken Docker context, empty `modules/`, fake "complete" claims).
- **Usage:** read this first; then `INDEX.md`; then only the chunk doc for the active chunk.
- **Inherited constraints:** none — this is the root directive.
Here is a complete, battle-hardened Agent Execution, Development & Debugging
Guideline built specifically to stop every failure mode your previous agent
committed across 1,780 commits (|| true bypasses, exit 2 -> 0 CI masks,
SumDrops(nil)=0, clock-only synctest_test.go, broken .dockerignore, empty
modules/, CHECKPOINTS.md churn, and fake "complete" claims).

It has 4 operational parts:

1.  A Drop-In AGENTS.md Constitution (copy directly into C:\cgo\AGENTS.md to
    constrain any agent you run),
2.  The High-Speed Development & Self-Audit Loop (how the agent develops and
    verifies in seconds without skipping truth),
3.  Layer-by-Layer Root-Cause Debugging Playbook (exact steps for Probes/dualns,
    Kernel 6.8 eBPF, Go API/AuthZ/SSE, Float64Array/Canvas UI, and Docker/CI),
    and
4.  Your 30-Second Reviewer Protocol (how you catch a weak agent lying before
    accepting any commit).

PART 1 — Drop-In AGENTS.md (Strict Execution Constitution for C:\cgo)

Replace or prepend your repo's AGENTS.md with these exact binding rules:

# AGENTS.md — STRICT EXECUTION & VERIFICATION CONSTITUTION (METEOLINK / CGO)

## 1. MISSION & IDENTITY LOCK
- Product name: **METEOLINK** (`cmd/cgo` single binary + embedded `web/frontend` SPA).
- Canonical M2 Title: *« Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar »*
- Scope boundary: You execute ONLY the active Chunk assigned by the user. Never invent features, extra binaries, extra services, empty directories, or side quests outside the active Chunk's file list.

## 2. THE SEVEN IRON LAWS (VIOLATION = AUTOMATIC REJECTION)
1. **Fail Closed & Never-Zero Law:**
   - NEVER swallow errors into `nil`, `0`, `""`, or `|| true`.
   - Failed `tc -j` / `PollStats` / probe scrapes MUST set `DropsMeasured=false` / `Measured=false` (`null` / `NaN` / `"—"` in UI & CSV) and mark the cell `degraded`/`invalid`. Never emit `0` for unmeasured data.
   - Missing topology files (`CheckBenchTopology`), missing auth tokens (without explicit `--dev-no-auth`), invalid `?iface=`, or corrupt manifests MUST fail closed with explicit errors.
2. **No Test or CI Theater:**
   - NEVER convert `exit 2` / `SKIP` into `exit 0` in CI (always run gates with `--strict`).
   - NEVER write tests that only test Go's stdlib (`synctest_test.go` testing virtual clock) instead of product state machines.
   - NEVER test Docker with `--version` alone; container smoke checks MUST boot `--serve` and assert HTTP `200` on `/`, extracted JS/CSS bundle URLs, and `/api/health`.
   - NEVER claim E2E passes production flows using `page.route('/api/**')` mocks; real E2E must hit the compiled Go server without API stubs.
   - NEVER pass `-count` without `-count=1` on Go tests.
3. **Proof-Before-Claim Law:**
   - You are FORBIDDEN from saying "verified", "complete", "passing", or "fixed" unless you ran the exact command AFTER your final edit and pasted the verbatim command + exit code + output tail.
   - Commit messages and `README` metrics MUST NOT state numbers unless a real executed JSON/CSV/SHA-256 artifact in the tree proves them.
4. **Zero-Bloat Law:**
   - NO new reverse proxies (`nginx`), external alert rule files (`prometheus/rules.yml`), empty folders (`modules/`), CLI wrapper shims (`cmd/meteolink`), unused GUIs (`fyne`), or duplicate views.
   - Keep the UI strictly 2 modes (`Cockpit` + `Labo`). Every `manifest.json` MUST list ONLY files that physically exist on disk and pass SHA-256 verification.
5. **Exact Pinning & Hygiene Law:**
   - NO `@latest` in GitHub workflows/Dockerfiles/tools; NO `^` or `~` carets in `package.json`; NO hardcoded machine paths (`C:/cgo`); NO secrets, tokens, or `hosts`-file hijacking (`meteolink.dev`).
   - All network dials (HTTP probes AND raw TCP bulk) MUST route through the single `net.Dialer.Control` SSRF guard blocking all reserved IPv4/IPv6 ranges (`100.64/10`, `169.254/16`, `192.0.0/24`, `198.18/15`, `224/4`, `240/4`, `fec0::/10`, `2001:db8::/32`).
6. **Zero Ledger Churn:**
   - NEVER modify `docs/CHECKPOINTS.md` or `docs/VERIFY.md` unless explicitly instructed in the final release/documentation Chunk.
7. **One Commit Per Verified Chunk:**
   - Do not create dozens of exploratory commits. Keep edits clean, minimal, and focused on the single Chunk's acceptance criteria.

## 3. MANDATORY EXECUTION LOOP FOR EVERY TASK
For every Chunk, execute these 5 steps in strict order:
1. **READ FIRST:** Read the exact target files and callers before writing code. Never assume a function signature or file exists.
2. **PLAN IN 5 LINES:** State which files will be modified/deleted and what invariant each edit enforces. Stop and ask if a required change falls outside the Chunk's allowlist.
3. **IMPLEMENT SMALL & CLEAN:** Make minimal, reversible edits; preserve `//go:build` cross-platform separation (`linux` vs `other`).
4. **SELF-AUDIT THE DIFF:** Run the forbidden-pattern check on `git diff` (zero `|| true`, zero `SumDrops(nil)`, zero `^`, zero `@latest`, zero `C:/cgo`, zero swallowed errors).
5. **RUN HARD GATES & PASTE OUTPUT:** Run targeted unit tests -> `go test -race -count=1 ./...` -> `tsc --noEmit` & `bun run test` & `bun run build` -> `--strict` gates. Paste exact raw results.

## 4. HARD STOP CONDITIONS (STOP & ASK THE HUMAN IMMEDIATELY)
STOP immediately without hacking around it if:
- A task requires revoking secrets, rewriting git history (`filter-repo`), or force-pushing.
- A Linux kernel/netns/eBPF test needs `sudo`/VM access you do not have on Windows (report `BLOCKED_ON_LINUX_HOST` instead of faking a mock pass!).
- A test reveals an architectural contradiction between two docs/specs.
- An external target URL/IP lacks explicit authorization.

PART 2 — High-Speed Development & Self-Audit Workflow

To avoid wasting 7 minutes per iteration and prevent silent regressions, enforce
this 3-tier speed loop:

Tier 1 — Sub-5-Second Inner Loop (Run After Every Edit)

  - Go logic: go test -count=1 ./<modified_pkg>/... (use testing/synctest for
    timers/phases/SSE so tests finish in < 100 ms).
  - Frontend logic: cd web/frontend && bun test <file> (pure TS in node,
    components in happy-dom, never jsdom for math/ring buffers).
  - Frontend types: cd web/frontend && npx tsc --noEmit.

Tier 2 — Pre-Commit Full Local Validation (< 20 Seconds)

Before claiming a Chunk is ready, the agent MUST run:

go vet ./...
go test -race -count=1 ./...          # On Windows without gcc, run -count=1 and state explicitly "-race deferred to Linux CI"
cd web/frontend && npx tsc --noEmit && bun run test && bun run build && node scripts/check-bundle.mjs && cd ../..
go build ./cmd/cgo

Tier 3 — Mandatory Self-Audit Diff Sweep (Before Reporting "Done")

The agent MUST inspect its own diff (git diff) against these exact forbidden
patterns:

# Must return 0 unexpected hits in the diff:
git diff | grep -E '\|\| true|C:/cgo|SumDrops\(nil\)|"@latest"|\^19\.|\^6\.|cgo-session-v1|meteolink\.dev'
# Verify every file created is actually imported/referenced and not a dead shell:
git status --short

PART 3 — Root-Cause Debugging Playbook (By Layer)

When a bug occurs, the agent must reproduce with a minimal failing test first,
trace the exact data boundary, and fix the root cause—never suppress the
symptom.

Layer 1: Probes, dualns & Linux tc Physics Debugging

  - Symptom: Latency looks doubled (e.g., 202 ms on 100 ms profile), too fast
    (< 2 ms on 4G), or drops show 0.
  - Root-Cause Checklist:
    1.  Check RTT symmetry: Inspect both kit/cgo-bench-dualns.sh AND
        pkg/campagne/dualns.go. Confirm each transit hop (w-edge, c-wan)
        receives DelayMs / 2 (never DelayMs on both!). Verify with raw tc qdisc
        show inside both namespaces.
    2.  Check HTTP status codes: Verify SmallObject in pkg/probe/probe.go checks
        200 <= resp.StatusCode < 300 and bytesRead > 0. Verify testbedsrv has
        obj16.bin provisioned (curl -i http://<sink>/small must NOT return 500
        "obj missing"!).
    3.  Check PollStats & tc -j errors: Inspect what tc -j -s qdisc show dev
        <iface> outputs manually inside the netns. Ensure JSON decode errors or
        non-zero exit codes propagate as DropsMeasured=false (NEVER
        SumDrops(nil)=0).
    4.  Check Ingress IFB: Verify bench-topology.json matches CGO_SHAPE_MODE
        (ingress vs egress), verify e-wan does NOT also hold a root CAKE shaper
        in ingress mode (no double bottleneck), and verify ethtool -k <iface>
        shows gro/gso/tso: off.
    5.  Check iproute2 gemodel: Verify tc qdisc show prints exact expected
        percentages (e.g. 0.1%, NOT 0.001%) and confirm measured loss rate
        on 10,000 test packets matches π = p / (p + r).

Layer 2: Linux Kernel 6.8 eBPF CO-RE & BPF_MAP_TYPE_RINGBUF Debugging

  - Symptom: eBPF fails to load, reports "ebpf" when unattached, counts
    SSH/Tailscale noise, or numbers disagree with tc -s.
  - Root-Cause Checklist:
    1.  Check Kernel 6.8 Hooks: Confirm kfree_skb is attached at
        tracepoint/skb/kfree_skb filtering ctx->reason ==
        bpf_core_enum_value(enum skb_drop_reason, SKB_DROP_REASON_QDISC_DROP)
        (never nonexistent qdisc:qdisc_drop!).
    2.  Check Exclusive Attach (E2): Verify via sudo bpftool prog show and sudo
        bpftool link show that ONLY ONE of fentry/tcp_rcv_established or
        kprobe/tcp_rcv_established is attached—never both!
    3.  Check Target Filter (E3): Dump the eBPF config map via sudo bpftool map
        dump and verify target_ifindex / target_dport match the active test
        interface/port (zero SSH port 22 / Tailscale traffic leaking in).
    4.  Check .o Hash Parity: Compare sha256sum pkg/probe/bpf/*.o against the
        embedded .sha256 sidecar and bringup README—stale binaries cause silent
        behavior mismatches.
    5.  Verify Non-Root Degradation (E1): Run cgo doctor as an unprivileged user
        (without CAP_BPF/root): it MUST report "tcpinfo" / "userspace" fallback
        cleanly, never "ebpf".

Layer 3: Go API, AuthZ, SSRF, Concurrency & SSE Debugging

  - Symptom: 401/403 mismatches, flaky SSE tests, slow /api/results, panics, or
    blocked dials.
  - Root-Cause Checklist:
    1.  Race & Deadlock Debugging: Run go test -race -count=1 ./api/...
        ./pkg/audit/... (on Linux/WSL/CI with gcc). Never fix a race by adding
        time.Sleep—use explicit sync.Mutex, atomic, channels, or
        testing/synctest barriers!
    2.  AuthZ & IDs: Write a table-driven httptest test verifying: Token A
        creates audit-<ts>-<hex> (201/200) → Token B (role="operator") gets 403
        on GET/DELETE → unauthenticated gets 401 → ?iface=eth0;id gets 400.
    3.  SSRF Single-Resolution Trace: Confirm BOTH SmallObject and
        bulk_cc_linux.go dial through the same SafeDialer (net.Dialer.Control
        checking netip.Addr.Unmap()); test with 100.64.0.1, 169.254.169.254,
        ::ffff:127.0.0.1, and 198.18.0.1!
    4.  Response & Timeout Audit: Check WriteTimeout (30s) doesn't kill
        /api/stream (confirm
        http.ResponseController(w).SetWriteDeadline(time.Time{}) works under
        live 20 Hz traffic for > 35s), and verify writeJSON returns 500 when
        given an unmarshalable struct (e.g., math.NaN() or chan int).

Layer 4: Frontend Float64Array Ring Buffer, Canvas & E2E Debugging

  - Symptom: Chart drops frames, renders 0 ms on gaps, React CPU spikes during
    20 Hz SSE, or Playwright fails layout/WCAG.
  - Root-Cause Checklist:
    1.  Check React Render Isolation: Add a quick console.count('CockpitRender')
        temporarily during dev—at 20 Hz SSE ticks, React MUST NOT re-render 20
        times/sec! Only 4 Hz KPI updates (4 renders/sec) + discrete
        DomainEventRing milestones may trigger React renders; the Canvas updates
        purely inside requestAnimationFrame when ring.dirty === true.
    2.  Check PC11 Gap Rendering: Feed null / undefined into
        TypedMetricRing.push(...) and assert Float64Array stores NaN (visually
        breaking the line cleanly, never dropping to 0 ms!).
    3.  Check Missing Assets (K3) & Live E2E (K12): Never debug UI solely via
        Vite dev server (vite dev) or page.route() mocks. Build the actual
        binary (make build or docker build), run cgo --serve, open DevTools
        Network tab (or Playwright), and verify 0 404s, 0 console errors, 0
        horizontal scrollbars, and 0 axe-core critical/serious WCAG violations
        at 390px and 1920px.

Layer 5: Container & CI Debugging

  - Symptom: Local tests pass, CI or Docker fails.
  - Root-Cause Checklist:
    1.  Inspect Docker Build Context: Run docker build --no-cache -f
        deploy/Dockerfile -t cgo:debug . and verify Stage 1 builds
        web/frontend/dist and Stage 2 embeds non-empty hashed JS/CSS chunks
        (embed.go:22-27).
    2.  Test Container Runtime Exactly Like Prod:
        docker run -d --name cgo-dbg -p 19090:19090 -e CGO_API_TOKEN=test-token-32chars-minimum-value cgo:debug \
          --serve --expose --addr 0.0.0.0:19090 --tls=false --insecure-allow-plaintext
        curl -fsS http://127.0.0.1:19090/api/health
        # Extract script/css links from / and curl each for HTTP 200; verify /api/results returns 401 without Bearer!
    3.  Audit Workflow Exit Codes: Never allow || true or if [ $rc -eq 2 ]; then
        rc=0; fi in .github/workflows/. Always require exact exit 0 from
        --strict gates.

PART 4 — Your 30-Second Reviewer Check (How You Audit the Agent Every Chunk)

Whenever the agent says "Chunk X is done", never trust its summary. Run these 4
checks in 30 seconds before accepting:

1.  Check the Scope: git status --short + git diff --stat → Did it touch ONLY
    files in the Chunk allowlist? If it edited CHECKPOINTS.md or unrelated
    views, reject immediately.
2.  Check for Coward Shortcuts:
    git diff | grep -E '\|\| true|return nil|SumDrops|exit 0|@latest|\^|page\.route|time\.Sleep'
    If you see a newly added || true, return nil on missing config, time.Sleep
    in unit tests, or page.route in live E2E, reject immediately.
3.  Check Real Execution Output: Did it paste actual stdout/stderr from running
    go test -count=1, bun run test, tsc, and the Chunk's specific verification
    gate? If it says "should pass" or "verified via code inspection", reject and
    make it run the command.
4.  Spot-Check One Negative Case Yourself: Run one failure test manually (e.g.,
    pass ?iface=eth0;ls, delete a file from a test manifest and hit
    /api/integrity, call an endpoint without a token, or run docker run without
    cgo version). If the negative case fails open, send it right back.

Save Part 1 directly into AGENTS.md right now as part of Chunk 0, and hold the
agent strictly to Parts 2–4 on every Chunk from Chunk 0 through Chunk 6. Tell me
whenever you've updated AGENTS.md or are ready to inspect the first Chunk diff!
