# eBPF-vs-`tc -s` confrontation — verifier runbook (2026-11-22)

Status: **the live kernel-trace step is PARKED.** The local contract, its
deterministic oracle and this runbook are complete; the real measurement needs a
bench host, a rebuilt eBPF object and human approval, none available in this
checkout. This document records no new measurement and promotes no VERIFY row.

## 1. Sealed evidence this runbook refers to (read-only — never re-seal)

- **226-event clean confrontation** — `docs/VERIFY.md` row 43,
  `/home/bench/eb-clean-report.json`, audit **`audit-1794860919`**,
  report sha256 `aa1f0b36d306ff5699876b4ed3626f6eb41d39bd4b88dfb42c86e6bdd352ad7c`,
  `source=ebpf`, object `8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`,
  grade A+. eBPF retrans events `1+225 = 226` (327248 B). `tc -s` e-wan cake
  `dropped 1 → 227` = delta **226** (before/after evidence sha256
  `8d60296ae3ac98e181148bfc2953b922766acb406b6a7648e7c58ba5ef439251` /
  `afb114f41e87a12ac13052e96dbc985a5b4f42a8c4a29b166e875edf686df1eb`).
  The two independent counters agree exactly on the causal chain
  (cake overlimit → TCP retransmissions). Documented in `probe/bpf/BUILD.md`
  § « Limitation mesurée ».
- **242-event audit — recorded, NOT interpreted** — `audit-1794798576`,
  `docs/VERIFY.md` row 43 / `docs/verification/2026-11-15-bench-vm/README.md:73-80`.
  `350392 B / 242 retrans events` on flow `2.0.30.10→1.0.10.10` (srtt_us 218806),
  `qdisc_drops=0`, on a 0 % emulated-loss path. Verbatim note: *"242 retransmit
  events on a 0% emulated-loss path — worth confronting with `tc -s` output on
  the shaper before calling it a fact about the pipe; the instrument's honesty
  posture records it, interpretation belongs to the verifier."* No conclusion
  about the pipe is drawn from 242 until the live step below is executed under
  approval.

These are evidence, not scratch space. Do not rewrite, re-seal, or reinterpret
them. `qdisc_drop_packets=0` is the probe's reason-filter blind spot on kernel
6.8 cake (`tc_skb_cb(skb)->drop_reason == 0` → `SKB_NOT_DROPPED_YET`, so the
`SKB_DROP_REASON_QDISC_DROP`-only filter discards every event) — **not** proof
of zero loss.

## 2. Local contract already in this checkout (no live data)

`cmd/cgo/confront.go` implements `cgo run --confront-out <path>`: it reads a
sealed kernel-truth JSON from `CGO_CONFRONT_INPUT`, validates it and copies the
exact values into `ebpf_vs_tc_report.json` (`cgo.ebpf-vs-tc/1`). It **fabricates
no counter** and launches no run. `source` must be `ebpf`; absent/illegal/null
required fields fail closed (exit 2, no file). Verdict `match = retrans_events ==
(tc_drops_after − tc_drops_before)`; `qdisc_drop_packets` is copied verbatim and
never drives the verdict.

Deterministic oracle coverage is **already present**; do not duplicate it:

- `cmd/cgo/confront_test.go` — full JSON shape (schema, provenance,
  `kernel_truth`, `tc_s`, `confrontation`), flag-absent no-op, and fail-closed on
  missing/invalid/non-ebpf/null seals.
- `cmd/cgo/qdisc_drop_faithful_test.go` — 226-event sealed-fixture-like data:
  non-zero `qdisc_drop_packets` copied verbatim, `match=true` when
  retrans==delta, `match=false` when they differ while the drop counter stays
  faithful (proves the drop counter never pilots the verdict).

Run locally: `go -C cmd/cgo test -count=1 ./...`. These use in-test fixture
strings only — no live host, no sealed bench artifact.

## 3. Exact steps for the PARKED live step

Execute only after §4 is satisfied, only through `C:\WEBSERVER` (never ad-hoc
ssh from an agent session), and with a separate verifier re-running the command.

1. Read `C:\WEBSERVER\safe-read-write-admin-guide.md`. Password source is
   `C:\cgo\mdp.txt`: load into the command's environment, never print, never
   copy into this repo, never type into a file. Invoke `safe_read.py` with
   `-X utf8`. Default rung is **R0**: classify with `safe_read.py --file`, stop
   on any refusal, `--run` one pass, keep the evidence file and its sha256
   sidecar.
2. **Rebuild the eBPF object with a reasoning fix** — on a pinned build host,
   *never on the bench VM* (repo rule: no `clang`/`gcc`/`llvm`/`linux-headers`
   on the VM, object is prebuilt). Toolchain per `probe/bpf/BUILD.md`: Ubuntu
   24.04, kernel 6.8 BTF, clang-18, libbpf-dev, bpftool. Widen the
   `on_kfree_skb` filter beyond `SKB_DROP_REASON_QDISC_DROP` to count
   reason-less cake drops, **or** expose `tc -s` `qstats.drops` as a separate
   named source. `make bpf` → `make bpf-check`; verify `readelf -S` contains
   `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`,
   `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb`. Pin the new sha256; rebuild
   must be reproducible on identical toolchain/BTF/path (`BUILD.md`
   § Reproductibilité). Do not commit the object before live 6.8 acceptance and
   the human's decision (`BUILD.md` § Statut).
3. **Real kernel trace** (R2/R3 writes: token + rollback line + rehearsal
   required): after the same-day rehearsal, in the operator-named quiet bench
   window and `ip netns exec ns_client`, attach the rebuilt loader, run one
   controlled shaped dualns cell, capture `tc -s` on e-wan cake **before and
   after**, and seal a kernel-truth JSON with `schema=cgo.kernel-truth-seal/1`,
   `source=ebpf`, the new `object_sha256`, and the exact before/after and eBPF
   counters. Keep both `tc -s` captures with sha256 sidecars and the tool's
   redactions.
4. Produce the confrontation report locally from that seal:
   `CGO_CONFRONT_INPUT=<sealed.json> cgo run --confront-out ebpf_vs_tc_report.json`.
   Success = report schema `cgo.ebpf-vs-tc/1`, provenance copied, `drops_delta`
   == `tc -s` delta, `match` computed from retrans vs delta only, and the
   independent counters agreeing. An independent verifier re-runs step 4 and
   owns any VERIFY-row promotion. If `qdisc_drop_packets` is still 0, that is
   the named blind spot — report it as such, never as zero loss.
5. Promote no row from this runbook. Any new evidence lands as a new checkpoint
   entry and a verifier-owned VERIFY update, with command, date, host, result
   and what stayed unproven.

## 4. Human approvals required before any host contact

All four artifacts must exist, same-day, or **no connection that can write**:

- a **named bench window** (quiet window the operator names) for any live load;
- a valid, non-expired **`APPROVED-*` token** in `C:\WEBSERVER`, scoped to this
  single task;
- a **rollback line** (how the qdisc/loader change is undone);
- a same-day **rehearsal**.

Without token + window + rollback + rehearsal, contact with sadcnwp/bench is
prohibited. `ftp2` stays out of scope until the operator rewrites the standing
2026-10-27 direction (as a new checkpoint, never an in-place edit).

## 5. Explicit prohibitions

- No presenting `qdisc_drop_packets=0` as "zero drop"/zero loss.
- No `clang`/`gcc`/`llvm`/`linux-headers` installation or runtime on the VM.
- No host/bench/sadcnwp contact without all four §4 artifacts; no SSH key from
  this session; no ad-hoc ssh bypassing `C:\WEBSERVER`.
- No touching the frozen corpus, `book/`, `C:\END`, `docs/CHECKPOINTS.md`,
  `docs/VERIFY.md`, gate definitions or permits; no fabricated measurement,
  seal or PASS; no gate weakening.
- No new VERIFY claim from a builder pass — a separate verifier re-run is
  mandatory for any number leaving the repository.

## 6. Unblock condition (exact)

**PARKED** until: (a) the rebuilt-object sha256 is pinned and its live 6.8
acceptance is verified; (b) all four §4 human artifacts exist same-day and the
operator names the bench window; (c) the real kernel trace and `tc -s`
before/after captures are executed through `C:\WEBSERVER` and preserved with
sha256 sidecars; and (d) an independent verifier re-runs `cgo run
--confront-out` on the sealed report and owns the VERIFY-row promotion. Until
then this task's honest DONE is the local contract, its deterministic oracle
and this runbook — not a live trace.
