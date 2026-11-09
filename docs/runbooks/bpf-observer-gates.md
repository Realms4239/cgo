# Runbook — eBPF observer kernel gates (`cgo-bpf-observer`) — APPROVED (a)+(b) 2026-11-09

> **État : APPROVED (a)+(b) 2026-11-09 — probes + object + loader pending.**
> Gate pack only in this document — no code, no loader, no eBPF object
> built here. Human granted (1) the Linux 6.8 build host and
> (2) `github.com/cilium/ebpf v0.22.0` (2026-11-09, render choice);
> Slices 2+ dispatch on the probes in §1 below.
>
> Contexte : `docs/runbooks/bpf-observer.md` (design, DEFERRED),
> `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §8
> (eBPF/BTF/CO-RE/cilium-ebpf, R1–R11, slices 0–6),
> `docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md`
> (pinned build), `docs/VERIFY.md:29-30` (BLOCKED rows: no committed
> `bpf_probe.o`, no Linux 6.8 run).
>
> Ligne BLOCKED « eBPF retransmit / RTT / qdisc drops » (`docs/VERIFY.md:29`)
> is the bottleneck: until the object (`make bpf` on the 6.8 host) and the
> loader (`session_linux.go`) exist, every `source=ebpf` claim is toy.
> Windows cannot build the object — Windows rows say "Windows body only",
> never "vérifié" where the condition is a Linux kernel feature
> (`bpf-observer.md` § Gate de construction 6.8; spec §8.12).
>
> Rappel AGENTS.md (stop-and-ask): adding a new dependency, a live bench
> run, a frozen-data touch, a merge to `main`, or scope growth all need
> the human before acting. Never run runtime `clang`, `gcc`, `llvm`, or
> `linux-headers` on the bench VM; the eBPF object is prebuilt.

---

## 1. Build-host gate (Linux 6.8) — probe before any `make bpf`

Pinned stack (do not substitute): Ubuntu 24.04 LTS, kernel 6.8.0,
clang-18 (`BPF_CLANG=clang-18`, `Makefile:46`), libbpf-dev headers
(`<bpf/*.h>`, `BUILD.md:18`), bpftool (generates `vmlinux.h`,
`BUILD.md:19`, `Makefile:51-52`), BTF at `/sys/kernel/btf/vmlinux`
(`CONFIG_DEBUG_INFO_BTF`). Targets at `Makefile:44-59`;
procedure at `pkg/probe/bpf/BUILD.md:21-41`.

If any probe below fails, the gate is BLOCKED — record the blocker with
owner + unblock condition and stop. Do not fabricate an object or a hash
(`BUILD.md:8-12`).

### Probes (run ON the Linux build host, not Windows)

| # | Precondition | Command | Expected output |
|---|---|---|---|
| H1 | Kernel 6.8, amd64 | `uname -r; uname -m` | `6.8.*-generic` (or `6.8.*`) + `x86_64`. A 6.18 WSL kernel (PC2, `VERIFY.md:29`) fails this row. |
| H2 | BTF present + parsable | `ls -l /sys/kernel/btf/vmlinux` | file exists, readable by builder (`root` or group). Presence ≠ parsable — next probe proves parsability. |
| H3 | BTF parses via bpftool | `bpftool btf dump file /sys/kernel/btf/vmlinux format c > /tmp/vmlinux.h && head -5 /tmp/vmlinux.h && wc -l /tmp/vmlinux.h` | exit 0; header starts with `#ifndef __VMLINUX_H__`-style guard and `wc -l` is thousands of lines. Non-zero exit = BTF absent/unparsable → BLOCKED. |
| H4 | Kernel UAPI headers present (documented, not substitute for H3) | `ls /usr/include/linux/tcp.h` | file exists (`TCP_INFO=11`, `TCP_CONGESTION=13` live in `/usr/include/linux/tcp.h`; spec §8.6.7 sibling notes). **Note:** the build itself does NOT need the `linux-headers` package — including generated `vmlinux.h` eliminates the system-header dependency (`bpf-observer.md` § BTF / CO-RE). Probed here only to record the host's UAPI baseline; never install runtime `linux-headers` on the bench VM (AGENTS.md "Never touch"). |
| H5 | clang-18 with BPF target | `clang-18 --version` and `clang-18 -target bpf -E - </dev/null; echo $?` | `--version` shows `clang version 18.*`; second command exits `0` (BPF backend present). `.BTF`/`.BTF.ext` require clang ≥8.0 or pahole (spec §8.5) — 18 satisfies it. |
| H6 | libbpf headers | `ls /usr/include/bpf/bpf_helpers.h /usr/include/bpf/bpf_tracing.h /usr/include/bpf/bpf_core_read.h /usr/include/bpf/bpf_endian.h` | all four files exist (`bpf_probe.c:17-20` includes them). `dpkg -l \| grep libbpf-dev` shows `libbpf-dev` installed. |
| H7 | bpftool usable | `bpftool --version` | exits 0 and prints a version. (Same binary as H3; probed separately because H3 needs both.) |
| H8 | Maintainer targets present | `grep -n "bpf\|BPF" Makefile` + `sed -n '44,59p' Makefile` | shows `BPF_CLANG ?= clang-18` (line 46), `BPF_CFLAGS := -O2 -g -target bpf -D__TARGET_ARCH_x86 -Wall -Werror -I$(BPF_DIR)` (47), `bpf:` (54-56) and `bpf-check:` (58-59) exactly as committed; `vmlinux.h:` (51-52) runs `bpftool btf dump …`. No target = BLOCKED (target missing on Windows host, `VERIFY.md:30`). |
| H9 | Build + hash check passes | `make bpf && make bpf-check; echo $?` | `make bpf` runs: `vmlinux.h` generation → `clang-18 -O2 -g -target bpf … -c pkg/probe/bpf/bpf_probe.c -o pkg/probe/bpf/bpf_probe.o` → `sha256sum … > bpf_probe.o.sha256` (no `vmlinux.h` committed — it is gitignored, `BUILD.md:29-31`). `make bpf-check` exits `0`. If the rebuilt SHA differs from the pinned one, investigate — DWARF embeds source paths, so another work directory changes the hash (`BUILD.md:45-48`); never silently accept a mismatched object. |
| H10 | Hook sections present | `readelf -S pkg/probe/bpf/bpf_probe.o \| grep -E 'raw_tp\|fentry\|kprobe'` | shows `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`, `kprobe/tcp_rcv_established`, and `raw_tp/kfree_skb` (`BUILD.md:35-41`). Missing section = BLOCKED. |
| H11 | Object licence kept | `head -5 pkg/probe/bpf/bpf_probe.c` | `SPDX-License-Identifier: GPL-2.0` (line 4). Repo MIT + object GPL-2.0 both stay with their files; never relicense the object (spec §8.5 licence). |

Install command (host only, after H1–H7 probes fail the same way):
```sh
sudo apt-get update && sudo apt-get install -y clang-18 libbpf-dev bpftool
```
per `BUILD.md:23-27`.

### What H1–H11 do not do

- They do not prove attach. `Check()` (verify SHA, then load-only) must
  never be sold as attach capability: "loadable (BTF + CAP_BPF); attach
  is verified by audit" (`bpf-observer.md` § capget).
- They do not prove BTF is magic. BTF is metadata/debug: it does not
  make a verifier-unsafe program load, nor make `tcpinfo` byte-exact.
  A missing CO-RE field fails the whole collection load and falls back
  to TCP_INFO — honest behaviour, not silence (`bpf-observer.md` § BTF).
- They need no runtime toolchain on the bench VM: no clang, gcc, llvm,
  or linux-headers at runtime (CO-RE object is prebuilt).

---

## 2. New-dependency approval ask — `github.com/cilium/ebpf`

**Stop-and-ask per AGENTS.md** (no new dependency without explicit human
approval). Today `go.mod:5-10` has no loader: `github.com/cilium/ebpf`
is absent and no loader code exists (spec §8.5, §8.7 gap 1).
The pin below is the one already current in the primary repo:
`docs/superpowers/plans/2026-10-27-cgo-ebpf-ground-truth.md:9,57`
(`go get github.com/cilium/ebpf@v0.22.0`, `go.mod`/`go.sum` gain the pin).

### Exact ask (paste into `go.mod` of the observer module)

```text
module github.com/Realms4239/cgo-bpf-observer
go 1.25

require github.com/cilium/ebpf v0.22.0

require golang.org/x/sys v0.36.0 // indirect — keep or bump only if cilium/ebpf demands it
```

(`golang.org/x/sys v0.36.0` is the current direct pin in `go.mod:9`;
`cilium/ebpf v0.22.0` must not drag new direct requires beyond it.)
Command once approved: `go get github.com/cilium/ebpf@v0.22.0 && go mod tidy`.
Expected: `go list -m all` in the observer shows stdlib + `cilium/ebpf`
only (plus `x/sys` transitively) — zero `metrics`/`audit`/`remediate`/`model`
imports (spec §8.9).

### Justification

- **Pure-Go loader, no cgo/libbpf binding.** `cilium/ebpf` (MIT Linux
  amd64/arm64, CI on kernel.org LTS) provides the full lifecycle the CGO
  Go side lacks: ELF → `LoadCollectionSpec` → `NewCollection` /
  `LoadAndAssign` → `link.Attach*` → `Collection.Close()`, plus `btf`
  (parse runtime BTF for CO-RE patching + honest `kernelFacts`), `rlimit`
  (`RemoveMemlock()` for pre-5.11, best-effort on 6.8 memcg accounting),
  `features` (capget/`CAP_BPF`/`CAP_PERFMON` check), `cmd/bpf2go`.
- **It replaces the lying heuristics.** `kernelFacts`
  (`kernelsource.go:19-31`) approximates caps by `os.Geteuid() == 0`
  (over-claims when the bounding set dropped the cap) and BTF by
  `os.Stat` (presence ≠ `CONFIG_DEBUG_INFO_BTF` parseable). The loader
  path replaces both with BTF-parse + `capget` and names the missing
  fact in `Check()`/`doctor warn` (spec §8.7 gap 2, §8.10 slice 2).
- **Licences stay separable.** `cilium/ebpf` MIT; repo MIT; BPF object
  GPL-2.0 (`bpf_probe.c:4`). No relicensing of the object.

### Fallback — vendored `cilium/ebpf` (only if the human accepts vendoring)

If module proxy access is unavailable on the build host but vendoring
is acceptable for this repo, the builder may run:

```sh
go get github.com/cilium/ebpf@v0.22.0 && go mod tidy && go mod vendor
```

and commit the generated `vendor/` tree **with** `vendor/modules.txt`
(`go build -mod=vendor` offline then works). Conditions:

- Vendoring is a **fallback, not the default**: the repo currently has
  **no `vendor/` directory** (verified: no matches under `vendor/**`),
  so adopting it is itself a human decision — record it in the
  checkpoint (proxy blocked, vendoring accepted, date).
- `vendor/` inflates every clone and every `git add` review; the
  vendored copy must match the pinned `v0.22.0` exactly — never hand-edit
  under `vendor/`. On upgrade, re-run `go get` + `go mod vendor` in the
  same commit as consumer code.
- If vendoring is **not** acceptable, the only honest path is the
  proxy-backed pin above. No third option: an unpinned loader is a
  supply-chain hole (spec §8.6.5) and stays BLOCKED.

---

## 3. Version matrix — what the design refuses to claim

| Claim | Design position | Unblocks when |
|---|---|---|
| **BBR implementation** | **Refused.** Observer only selects congestion control by socket-option name (`DialWithCC`/`SetConnCC`) and measures outcomes. It never implements BBR and must never claim BBRv3 exactness; sender-internal `max_bw`/`min_rtt`/BDP state is `unverified` from this module's vantage (spec §8.9 non-goals, §8.7 sibling notes, §8.11 #8). | Never in this scope. BBRv3 3×3 / engine remain separate BLOCKED rows until a 7.1 kernel with BBRv3 registered + human-approved run (`VERIFY.md:52`; spec §8.12). |
| **`dualns` 4/4 coverage** (`source=ebpf`, all 4 flows observed) | **Refused until Slice 3.** Coverage is an all-or-nothing gate: every dialed flow observed with ≥1 sample/event, else whole window falls to `tcpinfo` with coverage `n/N` + named reason (spec §8.10 slice 3). Missing ≠ zero; exact zero requires a sample. | Slice 3 passes on a human-approved `dualns` run (upload + download both `source=ebpf`). Until then no 4/4 number leaves the repo. |
| **Live attach** (and attach-derived numbers: overhead, retransmit payload, RTT, qdisc drops) | **Refused until the first successful capture.** Sequence: H1–H11 green → object + `.sha256` committed → `TestEmbeddedObjectPinned` passes → loader `Open` implements Open/Load/Attach/Tear-down with verifier-log capture → `doctor ok` on 6.8 with BTF + `CAP_BPF` → `Snapshot()` shows `samples>0` + forced retransmit `retrans_payload>0` same order as TCP_INFO (spec §8.10 slice 2, §8.11 #1–2). Load ≠ attach; unit-only `Status` without attach does not count. | First capture recorded in a checkpoint (host, date, command, output) and a **verifier who is not the builder** re-runs it (spec §8.11 #7). "Windows-green" is not "done". |
| **7.1+ `qdisc_drop` native hook** | **Refused on 6.8.** Linux 6.8 has no `qdisc_drop` tracepoint (introduced in v7.1); the 6.8-correct hook is `raw_tp/kfree_skb` filtered on `SKB_DROP_REASON_QDISC_DROP` via `bpf_core_enum_value`. Never ship `qdisc_drop` on 6.8. | Target kernel ≥7.1 with a new object + new hash (spec §8.12). |
| **Cross-kernel numeric merging** (6.8 frozen corpus ↔ 7.1 runs) | **Refused.** Each kernel generation is its own run set labelled `bench=dualns`/`kernel=…` (spec §8.10 slice 5). BBRv3 3×3 + `.dist` (30-sample floor) also wait for 7.1 + field trace (spec §8.12). | Never merged; 7.1 run is a fresh corpus set. |
| **BTF field names as magic offsets** | **Refused.** `vmlinux.h` + CO-RE relocations patch offsets at load; no hardcoded offset survives a kernel change. | N/A — build discipline, enforced by `-g` + relocations + reproducible SHA (H9–H10). |

Rows that stay **BLOCKED** until the two approvals + first capture land:
`docs/VERIFY.md:29` (eBPF retransmit/RTT/qdisc drops — no object, no 6.8
run), `docs/VERIFY.md:28` (Linux live acceptance), `docs/VERIFY.md:52`
(BBRv3 engine — needs ≥7.1 + BBRv3), `docs/VERIFY.md:53` (`.dist` —
needs a ≥30-sample field trace). A BLOCKED row moves to PENDING only
when its named condition clears, to VERIFIED only after the command
runs on the named host (`VERIFY.md` "How a row moves").

---

## 4. Dispatch prompt — copy-paste ready (one page)

Hand this verbatim to a future agent **only after the human has
approved (1) the Linux 6.8 build host and (2) `github.com/cilium/ebpf
v0.22.0`** (record both approvals in the dispatch message thread).
Shell out the fenced block below.

```text
You are the builder for slice 2 (object + loader) of the CGO eBPF
observer. The design is docs/runbooks/bpf-observer.md (DEFERRED) and
docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md §8.
The gate pack is docs/runbooks/bpf-observer-gates.md §1–§3 — read it
first; the artifact (Makefile:44-59 + pkg/probe/bpf/BUILD.md) wins over
any prose.

Human approvals (both already granted before this dispatch):
 (1) Build host: <Ubuntu 24.04 / kernel 6.8.0 x86_64 / host id>.
 (2) New dependency: github.com/cilium/ebpf v0.22.0 (go.mod line in
     docs/runbooks/bpf-observer-gates.md §2).

Scope (one line): Build bpf_probe.o + .sha256 on the pinned 6.8 host,
pin cilium/ebpf v0.22.0, and implement the loader Open / Check / Lookup
/ Snapshot / Globals / Attach / Close with verifier-log capture.

Do first — probe the build host (docs/runbooks/bpf-observer-gates.md
§1, H1–H8): uname -r, BTF file + bpftool dump, /usr/include/linux/
tcp.h baseline, clang-18 -target bpf, libbpf headers, bpftool version,
Makefile bpf targets. If any probe fails: STOP, record BLOCKED with
owner + unblock condition, commit nothing. Do not fabricate an object
or a hash.

Build: sudo apt-get install -y clang-18 libbpf-dev bpftool (host only,
BUILD.md:23-27), then make bpf && make bpf-check (exit 0), then
readelf -S pkg/probe/bpf/bpf_probe.o | grep -E 'raw_tp|fentry|kprobe'
(must show raw_tp/tcp_retransmit_skb, fentry/tcp_rcv_established,
kprobe/tcp_rcv_established, raw_tp/kfree_skb). Rebuilt SHA must equal
the pinned SHA or be investigated (DWARF path sensitivity, BUILD.md:
45-48) — never accept a mismatch. Keep vmlinux.h gitignored; keep the
GPL-2.0 object licence (bpf_probe.c:4) distinct from repo MIT.

Loader (only after the object exists): go get github.com/cilium/ebpf
@v0.22.0 && go mod tidy. go list -m all must show stdlib + cilium/ebpf
only (x/sys transitive). Implement internal/probe/bpf/session_linux.go
(//go:build linux && amd64: LoadCollectionSpec → NewCollection/
LoadAndAssign → link.Attach → Collection.Close, fresh maps per
session, concurrent-safe Lookup/Snapshot) and session_stub.go
(!linux || !amd64: identical API, typed UnavailableError). Replace
kernelFacts stat/euid heuristics with BTF-parse (cilium/ebpf/btf) +
capget CAP_BPF/CAP_PERFMON; Check() = verify SHA then load-only (never
sell load as attach). Capture the verifier log.

Never do: touch book/ or CSV schemas; run runtime clang/gcc/llvm/
linux-headers on the bench VM; implement BBR (CC is a socket-option
name only); ship qdisc_drop on 6.8 (use kfree_skb + SKB_DROP_REASON_
QDISC_DROP); claim dualns 4/4, live attach, overhead, or any ebpf
number before the first real capture; git add -A (stage explicit
paths only); merge to main.

Acceptance (run on the 6.8 host, quote outputs in the checkpoint):
 doctor ok (BTF + CAP_BPF); loopback flow in Snapshot() with
 samples>0; forced retransmit yields retrans_payload>0 same order as
 TCP_INFO; without caps → tcpinfo with named reason. make bpf-check
 exit 0; TestEmbeddedObjectPinned passes. go vet ./... + go test
 ./... -count=1 green; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go
 build ./... exit 0.

Close-out: write a checkpoint (command, date, host, result, what
stayed unproven), propose VERIFY row movement (builder never certifies;
independent verifier re-runs), never promote BLOCKED by confidence.
```

---
