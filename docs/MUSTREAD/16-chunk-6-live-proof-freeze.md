# Chunk 6 — Live Proof, Thesis & v2.0.0 Freeze

> Source: `C:\END\CGO5.md` lines 268–410. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** Linux kernel truth, authorized field trace, `.dist` calibration, RouterOS CHR proof, synchronized thesis, signed v2.0.0 tag and hard freeze.
- **Mechanism:** Part A human-controlled runbook with explicit gates, Part B bounded packaging prompt, Part C terminal definition-of-done checklist.
- **Contribution:** closes every cap with a named artifact and then stops.
- **Usage:** execute last, only after Chunks 0–5 are green.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Nothing — this is the final chunk.
Proceeding to Chunk 6 (Final Live Linux Proof, Field Trace + .dist Calibration,
RouterOS v7 CHR Transcript, Corpus Re-Baseline, Thesis Generation & v2.0.0 Hard
Freeze).

Only run Chunk 6 after Chunks 0–5 have passed every verification gate cleanly.

CHUNK 6 — Live Empirical Validation, Thesis Sync & v2.0.0 Hard Freeze

PART A — Step-by-Step Execution Runbook (You Control Every Gate)

Step 6.1 — Linux Kernel 6.8 Lab Validation (CGO-DEV, KVM / 4-vCPU Ubuntu 24.04)

(Do NOT run on the 1-vCPU Windows host; run inside your upgraded 4-vCPU / 4 GB
Ubuntu 24.04 VM.)

1.  Check Environment Readiness (K10):
    sudo ./bin/cgo doctor
    # MUST confirm: NTP synchronized (chrony/timedatectl), eBPF CO-RE attached, sch_cake + ifb available
2.  Run Full Linux -race + Real Kernel Integration Tests:
    sudo -E go test -race -count=1 ./...
    sudo -E go test -tags=real -count=1 ./pkg/qdisc/... ./pkg/campagne/... ./pkg/probe/bpf/...
3.  Generate Real eBPF-vs-tc -s Confrontation Artifact (VERIFY.md:43):
      - Run a controlled shaped bench cell with --confront-out
        docs/verification/v2-ebpf-confrontation/ebpf_vs_tc_report.json capturing
        side-by-side eBPF SKB_DROP_REASON_QDISC_DROP + tcp_retransmit_skb counts
        vs tc -j -s qdisc drops + TCP_INFO counts.
4.  Run Clean dualns Matrix + DGM Twin Scenario (Q1 Replaced):
      - Execute full clean dualns runs across {P1, P2, P3, P4} × {pfifo_fast,
        fq_codel, cake} × {cubic, bbr[, bbr3 if kernel exposes it]} in both
        Egress and Ingress IFB (ifb0), plus the Météo Madagascar DGM Twin (WRF
        bulk vs WIS 2.0 CAP 16 KB under 220 ms SLA).
      - Confirm iproute2 gemodel packet loss rate matches π = p / (p + r) (ge.go
        scaling verified live).
      - Confirm dualns P2 (100 ms) measures ~100 ms idle RTT (50 + 50 ms, Q2
        verified live).

Step 6.2 — Authorized Field Trace + .dist Empirical Calibration

1.  Authorized Field Audit (≥30 Samples):
      - Run your 60 s audit against your authorized constrained WAN
        target/gateway (with written authorization if institutional; never
        unauthorized probing of 41.207.49.27:9100).
      - Verify dual-sealed output (source_sha256 + telemetry_sha256 +
        aqm_eval.csv + report.json) passes GET /api/integrity (status:
        "verified").
2.  Calibrate .dist & Verify Live TC_LIB_DIR Replay:
    ./bin/cgo twin calibrate --audit <AUTHORIZED_AUDIT_ID> --out var/dist/field_empirical
    # Verify sidecar SHA-256, >=30 samples, p50 > floor
      - Run dualns empirical replay on CGO-DEV using TC_LIB_DIR + bare
        distribution name (iproute2 verified!) and save the tc -s + tail latency
        comparison log.

Step 6.3 — One Real MikroTik RouterOS v7 CHR Verification

1.  Boot a real RouterOS v7 CHR VM (image SHA-256 verified) on your host
    QEMU/VirtualBox.
2.  Generate a CAKE + fasttrack-safe script from your verified Step 6.2 audit
    inside METEOLINK (remember: HasValidRemediation must be true, never 0.0
    Mbit/s!).
3.  Apply it inside RouterOS v7 CHR, run /queue type print, /queue tree export,
    and /ip firewall filter print (confirming fasttrack handling), and save the
    raw console output to docs/validation/chr-v7-transcript.txt.

Step 6.4 — Rebuild Frozen Corpus, Refresh Docs & Approve book/ + memoire.docx

(Run ONLY after Steps 6.1–6.3 are archived on disk!)

Copy and hand this final bounded prompt to your agent:

# CHUNK 6 EXECUTION TASK — FINAL CORPUS PACKAGING, DOCS & THESIS BUILD (STRICTLY BOUNDED)

Execute ONLY Chunk 6 final packaging on `C:\cgo`. All code/tests are already locked in Chunks 0–5. Do NOT change Go/TS business logic.

1. **Rebuild Clean Frozen Corpus:**
   - Package ONLY the clean verified Chunk 6 `dualns` runs (zero legacy `prod.go` stacked runs, zero missing `valuation.json` files!) into `pkg/results/runs-frozen.tar.gz` and regenerate `docs/frozen-manifests.sha256`.
   - Run `cmd/gates --strict` (or `sh .githooks/run-gates.sh` with `--strict`) — MUST pass 6/6 with `0` skips and exit `0`.

2. **Empirical Grounding & Screenshots (`README.md`, `docs/img/shots/`):**
   - Replace obsolete screenshots in `docs/img/shots/` with fresh Playwright captures of `Cockpit` (`RingBufferHUD` + 3 cards + `220 ms` Canvas) and `Labo` (unclipped `P1–P4`, `DGM Twin`, verified Seal Inspector).
   - Update `README.md` metrics to cite exact verified `run_id` / `audit_id`, real `JFI`, and real SHA-256 hashes from Steps 6.1–6.2 (zero unattributed numbers!).
   - Clean up any dangling `slices ... verifier pending` status in `docs/VERIFY.md` so only verified artifacts remain.

3. **Build `book/` & `C:\cgo\memoire.docx`:**
   - Enforce the exact canonical title across `book/` and `C:\cgo\memoire.docx`:
     `Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar`
   - Ensure all chapters cite ONLY real verified Chunk 6 data:
     - Kernel 6.8 `tracepoint/skb/kfree_skb` (`SKB_DROP_REASON_QDISC_DROP`) + `BPF_MAP_TYPE_RINGBUF` + `ebpf_vs_tc_report.json` confrontation,
     - Clean `dualns` `AQM × CC` results (explicitly noting whether `BBRv3` ran natively or `CUBIC + BBRv1` ran on stock Kernel `6.8` with `bbr3` runtime detection ready),
     - Trace-driven `.dist` calibration (`calibrator.go`) & `Gilbert–Elliott` math (`iproute2` verified),
     - Ingress `ifb0` shaping & real `chr-v7-transcript.txt` RouterOS v7 proof,
     - Authorized `≥30`-sample field capture & Yas Business GOLD `1,388.89 Ar/h` institutional economics.

Step 6.5 — Tag v2.0.0, Verify Signed Release & HARD FREEZE

1.  Bump VERSION from 2.0.0-rc.1 → 2.0.0 (matching web/frontend/package.json &
    docs/api.md), commit once:
    git add -A
    git commit -m "chore(release): freeze v2.0.0 verified baseline and thesis artifacts"
    git tag -a v2.0.0 -m "METEOLINK v2.0.0 - M2 Final Verified Release"
    git push origin main --tags
2.  Wait for 100% green GitHub Actions CI + Release workflow (all jobs green,
    --strict gates green, 3-stage Docker smoke green, Playwright + axe-core
    green, pinned syft SBOM + Cosign .sigstore.json + SLSA Build L2 provenance
    attached).
3.  Download the released v2.0.0 artifact onto a fresh folder and run your
    SECURITY.md verification commands (cosign verify-blob, sha256sum -c
    checksums.txt, gh attestation verify, plus booting the downloaded binary and
    checking GET / & /api/health = 200 with "version": "2.0.0").
4.  Export final thesis PDF from book/ / C:\cgo\memoire.docx directly against
    tag v2.0.0.
5.  STOP. Zero commits after v2.0.0.

PART B — Terminal Definition-of-Done Checklist (Every Audit Cap & Blocker Closed)

Before declaring the project finished, verify every single box below has a green
artifact:

- [ ] Zero Active Caps (35, 40, 45): git log --all -S'ghp_' empty + gitleaks 0
  across history (K1); owner-scoped AuthZ + non-colliding audit-<ts>-<hex> IDs +
  explicit --dev-no-auth (K2); 3-stage Dockerfile serves HTTP 200 on / & bundles
  in CI (K3).
- [ ] Zero Measurement or Data Lies: PollStats failure marks DropsMeasured=false
  (K4); /api/integrity runs verifyManifests() cleanly (K6); zero manifests
  reference missing files (K8); CheckBenchTopology fails closed; unified
  DelayMs/2 RTT convention; gemodel & TC_LIB_DIR .dist verified live on CGO-DEV.
- [ ] Security & Supply Chain Clean: Unified SafeDialer SSRF blocks
  CGNAT/reserved/TCP bulk (K7); ?iface= strictly regex-checked (A03);
  Nginx/Prometheus/modules/cgo-gui/meteolink bloat gone; CI enforces --strict
  (K5); Release pins tools + drops || true + attaches SBOM/Cosign/SLSA (K9).
- [ ] Blazing Fast & Effortless 2-Mode Observatory: Kernel 6.8 eBPF
  BPF_MAP_TYPE_RINGBUF → 20 Hz Go SSE ring → Float64Array TypedMetricRing (NaN =
  absent) + RingBufferHUD inside streamlined Cockpit + Labo; Go synctest + JS
  happy-dom run in seconds; unmocked Playwright + @axe-core/playwright pass
  (K12).
- [ ] Thesis & Release Frozen: eBPF vs tc -s JSON, chr-v7-transcript.txt,
  ≥30-sample field capture, clean dualns corpus, and canonical memoire.docx
  all locked at v2.0.0.

You now have the complete, end-to-end blueprint (Design Brief + Chunks 0 → 6)
where you control every boundary and verify every gate. Whenever you're ready to
execute Chunk 0 (starting with the ghp_ PAT revoke + git filter-repo purge),
paste any outputs or errors here and we'll walk through each chunk side by side
until v2.0.0 is tagged and done.
