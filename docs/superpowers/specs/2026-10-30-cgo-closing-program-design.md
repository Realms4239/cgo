# CGO closing program — CHR router validation, host acceptance, verifier packets

*Status: proposed 2026-10-30 · Branch: `rebuild/plan1-economics-audit` ·
Slices 16–19 · One program spec, four slice contracts, one plan per slice ·
No frozen-corpus, CSV, production, or measurement change*

---

## 1. Why

Slices 4–7 are code-complete on Windows but cannot close without a Linux host, a
7.1 kernel, a real field trace, or production read access. Slices 8, 14 and 15
wait on an independent verifier and a human visual review. Two directive
deliverables were never built (`kit/setup-mikrotik-chr.sh`,
`docs/thesis_trace_driven_methodology.md`), and the sadcnwp case study proposed
in session `ses_e89dc0fe1ffewQxG5Nf3d0NlPq` never entered the ledger.

This spec closes the program honestly: it builds the one proof that is possible
today (RouterOS 7 accepts the generated configuration, verified in QEMU with the
CHR image), finishes the Windows-finishable half of slice 4, turns the first
Linux host-hour into a mechanical runbook, and defines the verifier and thesis
artifacts. It does not invent a measurement, does not touch production, and does
not move any `BLOCKED` row to `VERIFIED` without its live command.

### 1.1 Sources and their standing

| Source | What it gives | Standing |
|---|---|---|
| `C:\IDEA\directive.md` | Master prompt: eBPF, BBRv3, trace-driven, IFB, CHR, economics, UI, deliverables | Directional. Several technical claims are stale (see 1.2) |
| `C:\IDEA\directive-moreinfo.md` | Kernel 6.8 correction: no `qdisc_drop` tracepoint; `__qdisc_drop` is inlined; drops go through `kfree_skb_reason(..., SKB_DROP_REASON_QDISC_DROP)` | Authoritative for the 6.8 hook; slice 4 already matches |
| Session `ses_e89dc0fe1ffewQxG5Nf3d0NlPq` | Production read-only rule, "control is fictional", title honesty, 10 Hz/no-buffer decision, BBRv3/.dist deferral | Binding decisions, already reflected in the tree |
| `docs/superpowers/plans/README.md`, `docs/VERIFY.md` | Slice status and blocker rows | Authority for every status |

### 1.2 Reconciliation — directive against the tree (measured 2026-10-30)

| Directive | Tree | Verdict |
|---|---|---|
| §3 hook `qdisc/qdisc_drop` | Not in 6.8; plan uses `skb:kfree_skb` + `SKB_DROP_REASON_QDISC_DROP` | Directive stale; `directive-moreinfo.md` and the plan are right |
| §3 `tcp_retransmit_skb` exact bytes | That tracepoint carries no length; plan uses fentry | Directive stale; plan right |
| §3 object embedded, no runtime `clang` | Source committed (`pkg/probe/bpf/bpf_probe.c`), object absent | Blocked on one Linux 6.8 `make bpf` |
| §4 BBRv3 via `cgo kit install-bbr3` | `kit/install-bbr3.sh` exists, standalone; `bbr3` gated on the exact kernel token | Keep standalone: the SSH wrapper caps at 10 min, the build takes 30–90 |
| §5A `pkg/audit/calibrator.go` | Implemented as `pkg/qdisc/dist.go` + `cmd/cgo/calibrate.go`, tests green | Done, different path |
| §5B Gilbert–Elliott | `pkg/qdisc/ge.go`, bench `bursty` | Done |
| §5C IFB ingress | `pkg/remediate` + campaign ingress, Windows-green | Done; live row blocked |
| §5D CHR validation | `kit/setup-mikrotik-chr.sh` absent | **Slice 16** |
| §6 economics A/B/C, 220 ms SLA | Implemented, dashboard + report | Done |
| §7 "do not rebuild the UI" | Interface v2 rebuilt on the user's later instruction | Superseded by that instruction; only the telemetry-mode badge is missing (slice 17) |
| §8.6 methodology chapter | `docs/thesis_trace_driven_methodology.md` absent | **Slice 18** |
| §8 "208 valid runs" | 208 **valid lines**, 159 runs | Wording fixed in all generated copy |

## 2. Program shape

| # | Slice | Provable where | Closes |
|---|---|---|---|
| 16 | RouterOS v7 compliance via CHR in QEMU | Windows now; Linux namespace attach later | directive §5D; the "config validated" row |
| 17 | Slice-4 Go half + portable host readiness + acceptance runbook + telemetry badge | Windows (stub-tested); Linux for the live row | the Windows-finishable half of slice 4; makes slices 4–7 host-hours mechanical |
| 18 | Verifier packets for slices 8/14/15 + `docs/thesis_trace_driven_methodology.md` | Windows | the documentation half of "done"; directive §8.6 |
| 19 | sadcnwp read-only case study (contract only) | — (BLOCKED) | session step 3, never specced |

Order: 16 → 17 → 18. Slice 19 is a blocked contract and does not count against
the three-active-slice cadence (16 and 17 active, 18 starts when 16's report
surface lands).

### 2.1 Guards (every slice)

- No production write, load, root, or capture; production stays read-only and
  unnamed (pseudonym) until the operator grants otherwise.
- The frozen corpus, `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv` and
  `data/runs/` are untouched. New provenance is additive only.
- No VM image, ISO, CHR disk, or archive is committed; the CHR cache lives
  under a gitignored path and its SHA-256 is recorded in the proof artifact.
- No new Go runtime dependency. QEMU, CHR and host checks live in `kit/` and
  `docs/`; the single static binary stays single.
- No eBPF, BBRv3 or IFB claim before its live row runs; a stub is a stub.
- UI changes additive only: one validation line and one telemetry badge.
- French for user-facing text and comments; English for identifiers, this spec,
  plan and gate messages.

## 3. Slice 16 — RouterOS v7 compliance via CHR

### 3.1 What the proof is, and is not

It proves that a RouterOS 7.x instance **accepts** the generated remediation
script without error and that its `export` shows the queue type and queue tree
that the sealed report intended. It does **not** prove throughput, latency, or
that production changed. Two evidence levels:

| Level | Where | What it adds |
|---|---|---|
| `windows-tcg` | This host, QEMU with TCG | Import accepted, export structure matches, artifact sealed |
| `linux-namespace` | Linux 6.8 bench | Same, plus CHR interfaces attached to the dualns namespaces and queue counters moving under bench traffic |

A Linux run that cannot attach stays `windows-tcg`; it is never upgraded.

### 3.2 Components

| Path | Role |
|---|---|
| `kit/chr/setup-mikrotik-chr.sh` | One script, two backends. `--dry-run` prints the exact QEMU command line; `--image <path>` and `--report <path>` are the only required inputs |
| `kit/chr/chr-image.sha256` | Pinned image name and hash; the script refuses a different image |
| `kit/chr/README.md` | Human steps: download the CHR image, accept the licence (free/limited or evaluation), place it in the cache, never commit it |
| `kit/chr/export-fixture.txt` | Sanitized RouterOS `export` used by the unit tests (no addresses, no secrets) |
| `pkg/audit/routeros_validate.go` | Parser + structure diff + proof artifact writer; pure Go, no shell |
| `pkg/audit/routeros_validate_test.go` | Fixture diff, mutated-export fault proof, artifact schema |

The shell script orchestrates QEMU and the serial console; every judgement
(what the export contains, whether it matches) is Go code with tests.

**Host prerequisite, measured 2026-10-30**: `qemu-system-x86_64` is **not
installed** on DESKTOP-B6K9UCM (no binary on `PATH`, no `C:\Program Files\qemu`).
Installing QEMU is the human's decision — a system tool, not a Go dependency,
nothing committed — and until it is installed the Windows acceptance exits 2
with that instruction. The Go validator, fixtures and tests proceed regardless,
so the slice is never fully blocked on the installer.

### 3.3 Input and expected structure

Input is the **sealed report** `data/audits/<audit_id>/audit_report.json`
(never regenerated, never re-measured). The script extracts
`remediation.routeros` and `remediation.routeros_ingress`. The expected
structure is derived from the same report:

- queue type name(s) and their parameters (`cake-bandwidth`, `cake-rtt`,
  `cake-diffserv`, `cake-ack-filter`, `cake-nat`, `cake-wash`);
- queue tree entries: parent, target interface, packet marks, max-limit;
- the interface names the report named (`routeros_iface`, `routeros_lan`).

### 3.4 Flow

1. Verify the cached image against `chr-image.sha256`; absent or mismatched →
   exit 2 with the README step named.
2. Boot CHR headless with a bounded timeout; serial console captured to a
   transcript file from the first byte.
3. Wait for the login banner, log in, record `/system resource print` (version,
   architecture) into the artifact.
4. Send the report's script line by line with a short settle delay; capture
   every RouterOS reply.
5. `/export terse`; capture the full export.
6. Parse the export with the Go validator, diff against the expected
   structure, write `proof.json` + transcript + SHA-256 of both.
7. On Linux with root and the bench namespaces available, attach the CHR
   interfaces and run the bench's traffic; record counter deltas. Otherwise
   stamp the level `windows-tcg` with the reason.

### 3.5 Proof artifact (`data/chr/<audit_id>/proof.json`)

Additive, versioned, cited from `docs/VERIFY.md`; never committed (lives under
`data/`). Example values only:

```json
{
  "schema": "cgo.chr-proof/1",
  "audit_id": "<report audit_id>",
  "report_sha256": "<sha256 of the sealed report>",
  "chr": {"version": "<7.x.y as printed by /system resource>", "arch": "x86_64", "image_sha256": "<pinned>"},
  "level": "windows-tcg | linux-namespace",
  "accelerator": "tcg | kvm",
  "import": {"accepted": true, "error_lines": []},
  "checks": [
    {"name": "queue type cake-diffserv4", "expected": "…", "found": "…", "pass": true}
  ],
  "export_sha256": "…",
  "transcript_sha256": "…",
  "counters": {"rx_packets": 0, "tx_packets": 0, "drops": 0}
}
```

### 3.6 Failure ladder (all explicit, none silent)

| Condition | Result |
|---|---|
| Image absent or hash mismatch | exit 2; README step named; no artifact |
| QEMU absent | exit 2; install instruction, no artifact |
| Boot timeout | exit 1; serial transcript kept; `proof.json` written with `import: {"accepted": false, "reason": "boot timeout"}` |
| RouterOS rejects a line | `import.accepted=false`, RouterOS's error line stored **verbatim**, no retry loop beyond one |
| Export mismatch | `checks[].pass=false` with expected/found per field; exit 1 |
| Linux attach fails | level stays `windows-tcg`, reason recorded; exit 0 only if the Windows-level checks pass |

### 3.7 Report and UI surface

- Report gains an additive `routeros_validation` block: `{version, level,
  export_sha256, checks, proof_path}`.
- UI: one line in the audit report card — `Configuration RouterOS validée
  (7.x, CHR) · export <hash8>` or `non validée` with the reason. Never
  "appliquée", never "protégée".
- The dashboard does not gain a new card; the line lives where the remediation
  script already is.

### 3.8 Acceptance

- Windows: `bash -n`, `--dry-run` prints both backends' command lines; a real
  CHR boot imports a fixture report's script and produces a sealed
  `proof.json` with `level: windows-tcg`; a mutated export fails the diff
  (fault proof); the report and UI render the line from the artifact.
- Linux (runbook row): same, `level: linux-namespace`, counters non-zero after
  bench traffic; `docs/VERIFY.md` row moves only then.

### 3.9 Tests

Fixture-based unit tests (no VM): parser on `export-fixture.txt`, structure
diff pass, mutated export fail, artifact schema, hash pinning. The script is
exercised through a fake `qemu-system-x86_64` on `PATH` that replays a scripted
serial transcript, including the rejection and timeout ladders. The real VM is
never part of CI.

## 4. Slice 17 — slice-4 Go half, portability, runbook, badge

### 4.1 Go half (no kernel claim)

- `pkg/probe` exposes `Load() (Probe, error)` and a stub probe; the report gains
  its additive `kernel_telemetry` block with `source: ebpf | tcpinfo |
  unavailable` and the fallback ladder from the slice-4 plan (Windows/arm64 →
  stub; no BTF → TCP_INFO; no `CAP_BPF`/`CAP_PERFMON` → TCP_INFO; fentry fails →
  kprobe; partial window → whole window TCP_INFO with the reason).
- The embed is guarded (`//go:build` variant or generated file) so the binary
  builds without the object; `make bpf` on Linux adds it without Go changes.
- Tests: ladder order, provenance strings, all-or-nothing windows, and the
  hard rule that `source: ebpf` cannot be emitted without a loaded object and
  a recorded object SHA.

### 4.2 Portability — `kit/host-readiness.sh`

One script, one table, exit non-zero naming what is missing: kernel ≥ 6.8 and
BTF present, `CAP_BPF`/`CAP_NET_ADMIN`, `ip`/`tc`/`ethtool`, `bpftool`/`clang`
(build only), KVM, QEMU, CHR image cached, bench script present. On Windows the
Linux-only rows print `n/a`; the script still runs.

### 4.3 Runbook — `docs/runbooks/2026-10-30-host-acceptance.md`

Per slice 4, 5, 6, 7: prerequisite, exact commands, expected output lines,
sealed artifact path, and **what must not be claimed** after the step. Slice 5
names the human prerequisite (a real field trace of ≥30 baseline samples);
slice 7 names the 7.1 kernel build and keeps `install-bbr3.sh` standalone with
the 10-minute SSH reason. A "first hour" checklist orders the four slices so
they never mutate the same namespaces at once.

### 4.4 Telemetry badge (directive §7, the one missing item)

One header/ribbon element: `Télémétrie : eBPF | TCP_INFO | non instrumenté`,
read from the report's `kernel_telemetry.source`; absent block → `non
instrumenté`. No colour claim, no green.

### 4.5 Acceptance

`go test ./...` green with the stub; the badge renders all three states from
fixtures; on Windows `host-readiness` prints the table and exits non-zero with
the Linux rows named (expected); the runbook's commands are dry-checked
(`bash -n`, `--help`), not executed.

## 5. Slice 18 — verifier packets and the methodology chapter

### 5.1 Verifier packets

`docs/verification/2026-10-30-slices-8-14-15.md`: per slice, the exact commands
(`npm test`, `npm run build`, `go test ./...`, `python tools/verify.py`,
`node qa-walk.mjs`, `node perf-probe.mjs`), the expected counts, the evidence
paths, and a signature block. The builder prepares it; the verifier runs it and
owns the `VERIFIED` rows. This is not self-certification.

### 5.2 `docs/thesis_trace_driven_methodology.md`

Four sections, every number sourced:

1. Trace-driven `.dist` — what is built (`pkg/qdisc/dist.go`,
   `cmd/cgo/calibrate.go`), what a usable field trace must contain (≥30
   baseline samples), and why no `.dist` is cited today.
2. Gilbert–Elliott as implemented (`pkg/qdisc/ge.go`): the four parameters and
   the mean-loss equation.
3. eBPF precision, including the 6.8 callout from `directive-moreinfo.md`:
   `qdisc_drop` does not exist in 6.8, `__qdisc_drop` is inlined and not
   kprobable, drops are traced through `skb:kfree_skb` filtered on
   `SKB_DROP_REASON_QDISC_DROP`; exact retransmit bytes come from fentry, not
   from the `tcp_retransmit_skb` tracepoint. Status of the object: not built.
4. BBRv1 → BBRv3 → CAKE: what the corpus measured (BBRv1), what BBRv3 would
   test and why it is blocked (7.1 kernel), and what CAKE answers. Every cell is
   a `VERIFY.md` row or the words `non mesuré` / `bloqué`.

Acceptance: links gate, and every numeric token traced to a `VERIFY.md` row.

## 6. Slice 19 — sadcnwp read-only case study (contract, blocked)

Scope if granted: read-only service map (Nginx configuration, `tc qdisc`, link
rate, logs), no load, no writes, no root; operator approval on what may be
named; pseudonym by default ("the national forecasting server"). Deliverables:
service map, workload classes (A alerts, B telemetry, C bulk), SLA candidates,
experiment matrix, remediation plan. Acceptance criteria are written in the
plan; the slice row is **BLOCKED** on the operator's grant and executes nothing.

## 7. Cross-cutting

- Six gates stay six; no new gate. S16's fault proof lives in the unit tests.
- Cadence: 16 and 17 active; 18 starts when 16's report surface lands; 19 is
  blocked and does not count. A fourth active slice is refused.
- Every artifact that leaves the repository (CHR proof, runbook results,
  verifier packet) carries its command, date, host, and the words *what this
  does not prove*.
- Risks and mitigations: TCG is slow → bounded timeout, accelerator recorded;
  RouterOS console automation is fragile without `expect` → deterministic
  prompts, one retry, transcript kept, loud failure; CHR licensing is a human
  step → README, image never committed; QEMU is not installed today → the ask
  is explicit and the Go work does not wait on it; the stub could invite an
  eBPF claim → the ladder test forbids it; the program could sprawl → four
  slice contracts, each with its own plan and acceptance.

## 8. Acceptance mapping

| Spec acceptance | Slice |
|---|---|
| RouterOS 7 accepts the generated script; export structure matches; artifact sealed | 16 |
| Windows and Linux run the same CHR script; level recorded, never upgraded silently | 16 |
| RouterOS error lines stored verbatim; no "validated" on rejection | 16 |
| Report + UI line state the validation level and never claim production | 16 |
| Slice-4 loader/report/fallback green on Windows with no eBPF claim | 17 |
| One portable readiness script; runbook turns host hours mechanical | 17 |
| Telemetry badge honest in all three states | 17 |
| Verifier packets exist for 8/14/15; builder does not self-certify | 18 |
| Methodology chapter exists; every number traced or labelled blocked | 18 |
| Case study contract written, blocked, executes nothing | 19 |
| Frozen corpus, CSV schemas, production untouched; no image committed | all |
