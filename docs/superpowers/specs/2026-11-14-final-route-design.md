# CGO / Meteolink — Final Route Design (the last plan)

*Date: 2026-11-14 · Branch/trunk: `main` (multi-module) · Status: approved design, pre-plan*
*Language: English (spec/plan/gate convention, AGENTS.md). User-facing product text: French.*

## 1. Goal

One final route that finishes the project: every audit kill-list item closed or
honestly risk-accepted, the three security caps remediated, the four thesis
evidence gaps filled, a streamlined decision-first interface, and a signed
`v2.0.0` release plus the Master 2 thesis PDF. **Full V2 closure, gated.**
Open-ended to closure; anything a human has not cleared stays an honest
`BLOCKED` row, never a fake `VERIFIED`.

**Thesis title (locked, title 2):**
« Optimisation temps réel de la qualité de service des liens d'accès contraints :
une architecture d'observabilité eBPF et de régulation AQM/BBR à preuves scellées
pour les flux critiques de Météo Madagascar ».
Aligned to ISGEI MSTI-STR « Sciences et Techniques de l'Information et des
Systèmes Temps Réel » (Annexe D.2). The old « Conception et évaluation d'un banc
d'essai… » title is rejected as not master-worthy.

## 2. Locked decisions (2026-11-14)

1. **Done bar:** full V2 closure, gated; human-only actions are named decision
   gates with owners and fallbacks.
2. **Trunk:** `main` multi-module architecture is kept (18 submodules + root). The
   5 unique `operator-cockpit` commits are ported. M0 reconciles every audit
   finding against `main` before any fix (the audit is HEAD-scoped; `main`
   already closed some items and introduced regressions).
3. **Corpus:** frozen 159-run corpus stays byte-identical; legacy-stacked and
   pre-GE-fix bursty cells labeled via a quarantine **sidecar** (no CSV schema
   change); clean dualns runs become a **new sealed v2 corpus** with its own
   anchor.
4. **Cap 40:** full purge — revoke/rotate the PAT, mirror backup, `filter-repo`,
   gitleaks = 0 across history, human-approved force-push + origin GC, remote
   secret-scan gate, fork check.
5. **Field access:** DGM granted (audit's BLOCKED superseded). `safe_read.py` /
   `safe_write.py` rungs R0–R3; fresh same-day R2 tokens; **NO REBOOT** on
   sadcnwp; ftp2 compromise untouched; pseudonymized artifacts.
6. **Bench platforms (resolved):** no rebootable Linux box exists. **A** passive
   eBPF observatory on real DGM traffic + **B** LXD container synthetic bench,
   both on the sadcnwp host. **BBRv3 BLOCKED** (needs a rebootable kernel).
7. **Interface:** the stale IDEA4 §7 "no redesign" mandate is released to
   reasoning; an evolutionary 2-mode shell is synthesized (not a big-bang
   rewrite, not the stale 6-view status quo).
8. **Execution:** sequential bite-sized weak-agent tasks; parallel only when
   provably shared-no-state; evidence runs serialized; builder ≠ verifier.

## 3. Ground truth that shapes the route (discovery 2026-11-14)

- **Audit:** 34/100; caps 35(ii) (plaintext `0.0.0.0:9090` no auth), 40 (PAT in
  public history), 45 (same fact, transport). Kill list KL-01…KL-20.
- **`main` regressions (V2 never saw these):** shipped binary serves a **stub
  SPA** (`api/frontend`, real `web/frontend` orphaned); `/metrics` is a
  **hardcoded-zero facade** (`api/metrics.go`, honest `exporter/` unwired); four
  **unwired duplicate modules** (`guard/`, `exporter/`, `cite/`, `attest/`).
  Measurement defects Q2/Q4/P1/B1/seal-asymmetry/GE all still open. Auth partial
  (13 POST mutations; GETs/SPA open). CI release-only, 3 partial gates.
- **GE ×100 grounded:** iproute2 `tc/tc_util.c` `parse_percent` does
  `strtod/100` unconditionally; `q_netem.c` gemodel uses it. Repo emits raw
  decimals → 100× under-injection. Fix: emit `p*100 …%`. Legacy bursty corpus
  cells are pre-fix.
- **BBRv3 gate:** honest *with the repo provisioner* (`install-bbr3.sh` registers
  a distinct `bbr3`); would refuse a stock official kernel. Fix: provenance
  detection + keep the provisioner.
- **Hosts:** Windows laptop = i5-5200U 2-core/8 GB, 849 MB free RAM, C: 309 MB
  free → build/CHR/thesis only, NOT a bench. CGO-DEV host (PC2/nemo) OFFLINE,
  1 vCPU/3 GB → marginal fallback. sadcnwp = 80 threads/62 GB/BTF/6.8, all qdiscs
  signed, but compromised (40-core XMRig), production, no KVM group, no reboot.
- **Thesis:** ~60% built; gaps = `.dist` fidelity, eBPF kernel-truth + proxy
  error bound, CHR equivalence, countersigned case study. Zero per-cell
  citations; false "every SVG fingerprinted"; stale "160 runs". **The content is
  extremely stale against the new direction:** it still asserts "no field
  campaign", "eBPF not compiled", "no CHR booted", legacy single-flow numbers,
  the old 4-mode UI, and a monolith architecture — all made false by this route.
  Phase 6 is a major rewrite, not a gap-fill. PDF via
  `book/tools/build-book.ps1` → Word COM; gates `audit-*.py`.

## 4. Execution surfaces & the LXD safety envelope

| Surface | Role |
|---|---|
| Windows laptop | builds, FakeRunner unit tests, CHR (QEMU windows-tcg), thesis PDF |
| sadcnwp host | **A** passive eBPF observatory + passive field trace; **B** LXD container synthetic bench |
| CGO-DEV | offline/weak; optional fallback only |
| BBRv3 | BLOCKED — no rebootable kernel; honest fence |

**LXD safety envelope (every clause a gate):** shares the running 6.8 kernel (no
new kernel/bootloader/reboot/kexec); `boot.autostart=false`, no host
systemd/cron/timer/symlink created, manual start/stop, nothing fires after the
pass; `network=none` (dualns is self-contained in internal netns/veth — no
`lxdbr0`, no host iptables/route/NIC change); no group/sudoers/firewall/sshd/
production-service change; `limits.cpu`/`limits.memory` bounded + quiet-window
preflight; least privilege (unprivileged nested where possible, privileged only
if eBPF requires it, risk named + operator-accepted); rehearsed clean teardown
before any data matters, zero host residue; SHA-256-pinned Ubuntu image + audited
binary; R2 discipline (fresh token, rollback, rehearsal, second channel, operator
watching, one scope line). Residual impacts named: eBPF tracepoints attach
host-wide (transient, read-only, filtered, detach on stop); bounded CPU/RAM.

## 5. The route (reality-first, dependency-ordered)

Day 0 launches the longest-lead items in parallel (shared-no-state): passive
field trace, cap-40 revoke + purge prep, disk triage, sadcnwp R0 read, DGM R2
token requests. The field trace runs passively through Phases 0–3.

### Phase 0 — Trunk repair & containment
- WP0.1 Day-0 launch (parallel): cap-40 revoke (human); disk triage (caches→D:);
  sadcnwp R0 read (LXD state, `/dev/kvm` ACL, load, BTF); field trace start; R2
  token requests.
- WP0.2 Main reconciliation + trunk lock: finding→status→task map; port the 5
  HEAD commits; branch/worktree cleanup (human gate).
- WP0.3 Stub SPA repair: wire `web/frontend` embed at the api seam, delete
  `api/frontend/`, `ValidateEmbeddedAssets()`, fix `setup.go`, rewrite
  `kit/test-fresh-clone.sh` to build+serve+curl assets.
- WP0.4 Duplicate-module disposition (deletion test): delete `guard/` (keep
  `api/auth.go`); wire `exporter/`, delete the `/metrics` facade; delete-or-wire
  `cite/`; wire `attest/` into verify/release; keep `runreq/`, `capabilities/`.
- WP0.5 Cap-40 purge (human force-push gate): mirror→filter-repo→gitleaks 0→
  force-push→GC→fork check→remote scan gate.
- WP0.6 Identity/version SSoT + dead-surface cuts: title 2; VERSION single
  source; delete `cmd/cgo-gui`, `npm/`, dead fonts, orphan configs; **keep
  `cmd/cgo-chr`**; triage 37 TODOs; purge `C:/cgo` from code/tests not ledgers.
- WP0.7 Ledger/gate hygiene: update plan index (do NOT freeze ledgers);
  `-count=1`; fix phantom `test-real`/`bpf-check`; real `check-bundle` guard;
  full-history secret scan in CI.
- **Exit:** clean clone serves real SPA (assets 200); gates green `-count=1`;
  gitleaks 0; one auth seam; honest `/metrics`; title 2 + version SSoT.

### Phase 1 — Security & delivery (caps 35(ii)/45)
- WP1.1 Exposure: default loopback+TLS unit; non-loopback refuses without
  `--expose`+token+TLS; auto-generate `/etc/cgo/auth.token`; **two-adapter auth**
  (Bearer CLI/API + HttpOnly cookie session for browser/SSE); complete
  `observeBlocked`; timeouts; security headers; per-IP rate limit (429);
  `deploy.md` CIDR fix.
- WP1.2 SSRF (`probe/ssrf.go`, Dialer.Control resolved-IP guard); root-CA SPKI
  pin + uninstall + 365d; scoped `cgo-netctl` sudoers; guest-setup Tar-Slip/exit/
  sink/m-host/have_net; host-readiness exit-2; PS1 IPv4 regex.
- WP1.3 CI/CD: push/PR `ci.yml` (security/go-linux-race/web-e2e/gates); frozen
  gate = named host-only skip (exit 2 ≠ green); harden `release.yml` (full gates,
  real SLSA L2 attestation, all-module SBOM, cosign without `|| true`, versioned
  archives).
- WP1.4 Retention/governance: `cgo audit purge`; SECURITY.md/PRIVACY.md/config.md.
- WP1.5 Docker hardening: non-root, token env, scoped caps, nginx TLS template.
- **Exit:** unauth→401; exposure refused; SSRF→403; CI green; signed release
  verifies on a fresh host.

### Phase 2 — Measurement physics (container RUN evidence)
- WP2.1 Probe fidelity: P1 StatusCode→absence; ping chronological (remove sort)
  + `-W/-s/-M do`; bulk error propagation + calibrated deadline; hello framing.
- WP2.2 Channel physics: Q2 `DelayMs/2`; GE emit `p*100…%` (grounded); `.dist`
  TC_LIB_DIR `<128` bare-name; live readback ≈ π.
- WP2.3 Qdisc parity: Q4 full CAKE flags; dynamic tbf burst; reproducible
  fq_codel; dead-code removal; 4 root-fallback branch tests; `tc -j -s` parsing.
- WP2.4 IFB/lab: I1 `CGO_SHAPE_MODE` export + topology.json; I2 ethtool off;
  T1 `set -euo pipefail`+trap+kill dashboard; T2 numeric validation; 0 orphans.
- WP2.5 Results B1–B4 + one verdict: Best None-state + `(profile,direction)`;
  full-corpus ProvenanceSHA; sorted median; MW tie-correction; ETag/pagination;
  class-aware Δp95 + `<95% Critical` in `metrics/verdict.go`; delete `health.ts`.
- WP2.6 Durability/budgets/seal: WriteFileAtomic+flock; shutdown ctx; sink
  `>600s`; unified duration; clean `both`; HTTP dual seal; R1 measured asymmetric.
- WP2.7 BBRv3 gate+provisioner (engine): provenance detection; keep
  `install-bbr3.sh`; honest 3×2 fallback. Live matrix BLOCKED.
- **Exit:** container P2≈100 ms; GE≈π; `.dist` loads; IFB 0 orphans; kill -9
  recovers; dual seals; Best correct P1–P4; one verdict.

### Phase 3 — Kernel truth (eBPF)
- WP3.1 TCP_INFO proxy first: exact `tcpi_bytes_retrans`; quantify ×1448 error.
- WP3.2 C-source corrections (IDEA5/Annexe F): kfree_skb+QDISC_DROP reason;
  fentry XOR kprobe; target filter; ringbuf 16 MiB; srtt/cwnd/**ssthresh**; netns
  filter; CI-built object + sha256 (cilium/ebpf dep = human gate).
- WP3.3 Loader + ladder honesty: E1 ebpf only on full success; E4 window deltas;
  ringbuf overflow counted.
- WP3.4 Cross-validation: retrans≈tcpi_bytes_retrans; drops≈`tc -s`; srtt≈ping;
  overhead <2% under flood.
- WP3.5 Passive observatory (A): scoped CAP_BPF on real DGM traffic, read-only,
  miner-excluded, pseudonymized, "compromised host" label.
- WP3.6 Report integration: wire `kernelTelemetryFromSummary`; honest tier+reason.
- **Exit:** object attaches; three witnesses agree; overhead <2%; real-traffic
  observatory; honest kernel telemetry.

### Phase 4 — Field evidence
- WP4.1 Passive field trace on the host (sadc), authorized public target,
  CAP_NET_RAW R2 grant, ≥30 sealed samples.
- WP4.2 `.dist` calibration: `twin calibrate`; container KS/quantile match; the
  fidelity figure.
- WP4.3 New v2 corpus: clean dualns matrix in container; separate anchor;
  quarantine sidecar; frozen corpus untouched.
- WP4.4 CHR real transcript: QEMU boot, SHA-256 image, real export diff (fix
  C1/C2/C3/C5); `proof.json accepted:true`; transcript committed.
- WP4.5 Countersigned case study: operator-mediated service map; operator-applied
  remediation (R2); before/after p95 + Ariary; DSI countersign.
- WP4.6 VERIFY rows moved by the verifier; BBRv3 stays BLOCKED.
- **Exit:** ≥30 real samples; `.dist` KS-verified; v2 corpus anchored; CHR real;
  countersigned case study.

### Phase 5 — Interface synthesis
- WP5.1 Evolutionary 2-mode shell (Cockpit/Labo) from verified components;
  root-persistent SSE.
- WP5.2 TypedMetricRing (Float64Array, NaN) + DomainEventRing; event-driven SSE
  flush (20 ms coalescing, TCP_NODELAY); stale≠live badge.
- WP5.3 Cockpit decision-first (verdict→money figure→prescription→mechanism→
  seal footer); one engine verdict.
- WP5.4 Labo (unclipped P1–P4, Statique/Empirique `.dist`, Egress/Ingress/both,
  DGM twin comparator, honest seals).
- WP5.5 Honesty F1–F18; surgical additions (badge, MGA/h, SLA `<95% Critical`,
  Copy-RouterOS Ingress+Egress).
- WP5.6 WCAG AA (44px, contrast ≥4.5, forced-colors, chart data-table,
  OFL/NOTICE); populated lexicon; ArmButton on 8 uncovered mutations.
- WP5.7 Real tests (testing-library, unmocked Playwright 390/890/1920,
  deterministic SSE flake fix, TUI sync).
- **Exit:** tests + unmocked e2e green; populated lexicon; contrast AA; 0 flakes
  quiet.

### Phase 6 — Thesis rewrite & freeze (MAJOR work; the content is extremely stale)

The thesis source (`book/content/*.js` → `build-docx.js` → `memoire.docx`)
describes the OLD direction and is stale against everything this route produces.
It still says "no field campaign executed", "eBPF object not compiled", "no CHR
booted", the legacy single-flow frozen-corpus numbers, the old 4-mode UI, and a
monolith architecture. The new direction (title 2: real-time QoS optimization +
kernel observability + sealed evidence) and the route's new evidence make all of
these false. Phase 6 is a substantial rewrite/reconciliation, not a light
gap-fill. `book/` is untracked/gitignored — the rewrite is local, never
committed; the PDF is the deliverable.

- WP6.1 Thesis audit & reconciliation map: inventory every stale claim in
  `book/content/*.js` against the new direction + the route's evidence; produce a
  claim→status→rewrite map (old title, "no field campaign", "eBPF not compiled",
  "no CHR", legacy numbers, old UI, monolith, G7 uninstrumented, dualns bridge
  "not executed").
- WP6.2 Re-frame around title 2 + the seven contributions: rewrite the
  introduction/problem statement around real-time QoS optimization + kernel
  observability + the decision loop (measure→grade→remediate→validate→re-measure→
  prove ms+Ariary); foreground the seven contributions (honesty-invariant
  measurement, quarantine taxonomy, trace-calibrated emulation, kernel-truth +
  proxy error bound, Linux↔RouterOS equivalence, costed waste, countersigned case
  study); position against Flent/RRUL, perfSONAR, Smokeping (cite Flent); the M2
  critical-reflection descriptors.
- WP6.3 Method chapter rewrite (`03-partie2.js` +
  `docs/thesis_trace_driven_methodology.md`): trace-driven `.dist` (Mahimahi),
  GE 4-param Markov math, eBPF CO-RE precision (kfree_skb reason,
  srtt/cwnd/ssthresh, the ×1448→tcpi_bytes_retrans proxy error bound), gates
  G0–G7 (G7 now instrumented), the honesty algebra (absent≠zero, stale≠live), the
  quarantine taxonomy (46% invalid as a finding).
- WP6.4 Results chapter rewrite (`04-pratique.js`): the NEW v2 corpus numbers
  (clean dualns 4-flow) with per-cell run_id/seal citations; label legacy
  single-flow numbers historical; the `.dist` fidelity figure (KS distance); the
  eBPF three-witness cross-validation; the BBRv3 BLOCKED fence; the countersigned
  before/after case study (ms + Ariary); the offloads-era finding; the GE ×100 and
  Q2 RTT-doubling discoveries as critical reflection.
- WP6.5 Annexes rewrite: Annexe E (the new 2-mode Cockpit/Labo interface, fresh
  captures); Annexe F (kernel telemetry NOW compiled+attached, CHR NOW booted —
  real evidence, not "not yet"); SVG-fingerprint truth; the v2 corpus schema +
  quarantine sidecar.
- WP6.6 Front matter + citations: title 2 across surfaces; abstract/résumé
  rewritten for the new direction; keywords + sigles updated (eBPF, CO-RE, BTF,
  LXD…); per-cell citations; 160→159; ISGEI MSTI-STR alignment (Annexe D).
- WP6.7 Build, audit, freeze: `build-book.ps1` → Word COM → PDF;
  `audit-archive/biblio/renvois/nombres.py` green; figures regenerated from the v2
  corpus (Go-generated, hash8); `v2.0.0` tag; signed release verified on a fresh
  host; maintenance-mode archive; final VERIFY/checkpoint; honest BLOCKED
  (BBRv3); human merge decision.
- **Exit:** the thesis reflects the new direction with cited new evidence; PDF
  builds; audits green; release verified; ledgers honest.

## 6. Execution model (weak-agent-optimized)

- Long sequence of **bite-sized tasks** (2–5 min steps: failing test → run →
  minimal code → run → commit); exact paths/code/commands; no placeholders; no
  judgment left to the agent. Each task is the smallest independently-testable
  unit with an `Interfaces: Consumes/Produces` block and its own evidence path.
- **Sequential by default** (shared trunk/binary/gates); parallel only when
  provably shared-no-state; evidence runs serialized on a quiet host.
- **Builder ≠ verifier:** a fresh reviewer re-runs the named command and owns
  each VERIFIED row (verification-before-completion: no claim without fresh
  evidence).
- Skills governing execution: `subagent-driven-development` (fresh subagent per
  task + two-stage review) or `executing-plans` (inline checkpoints);
  `systematic-debugging` (root cause first); `codebase-design` (deep seams);
  `verification-planning` (evidence path per claim).

## 7. Human decision gates (named owners)

| Gate | Owner | Phase |
|---|---|---|
| PAT revoke/rotate + token-use audit | user | 0 |
| Force-push + origin GC approval | user | 0 |
| Branch/worktree cleanup + merge direction | user | 0 |
| `cilium/ebpf` dependency approval | user | 3 |
| sadcnwp R2 tokens (CAP_BPF, CAP_NET_RAW, LXD) | user + DGM operator | 0/3/4 |
| Operator-applied remediation (case study) | DGM operator | 4 |
| DSI countersignature | DSI | 4 |
| Thesis title sign-off | user | 6 |
| Merge to main / v2.0.0 release | user | 6 |

## 8. Standing laws (every task inherits)

Frozen corpus byte-identical (sidecar quarantine + separate v2 corpus); no
compiler on VMs/host (eBPF CI-built, `-race` on CI); the LXD safety envelope; no
reboot on sadcnwp; ftp2 untouched; pseudonymize DGM artifacts; secrets never
echoed/committed; never touch `book/data/runs`, `backup/`, `data/`, `epbfBBR.md`,
the remote PC `100.114.101.13`, Graylog; ledgers append-only, never frozen;
French user-facing text, English specs/plans/gates; `make` absent → `python
tools/verify.py` named, never waived; exit 2 ≠ green.

## 9. Closure matrix (audit finding → phase)

KL-01 cap40 → P0 · KL-02 exposure → P1 · KL-03 clean-clone → P0 · KL-04 Q2 → P2 ·
KL-05 corpus → P4 · KL-06 P1 → P2 · KL-07 CI → P1 · KL-08 SSRF → P1 · KL-09
root-CA → P1 · KL-10 CHR → P4 · KL-11 R1 → P2 · KL-12 deploy.md → P1 · KL-13 seal
→ P2 · KL-14 verdict → P2/P5 · KL-15 retention → P1 · KL-16 a11y/lexicon → P5 ·
KL-17 IFB/fasttrack → P2 · KL-18 phantom gates → P0 · KL-19 README numbers → P6 ·
KL-20 MTU/IPv6/overlap → P2. Q1/Q4/P2/P3/P4/B1–B5/F1–F18/C1–C5/E1–E4/R1–R3/I1–I2/
S1–S5/T1–T2 → phases as above. Stub-SPA/facade-metrics/unwired-modules (main
regressions) → P0.

## 10. Unanswerables kept honest (BLOCKED with condition)

BBRv3 live matrix (no rebootable kernel); module-wide `-race` (runs on CI);
field validity beyond the granted trace; thesis-title history (resolved to title
2 by user decision); cap-40 token validity (unknowable offline, not required).