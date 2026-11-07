# eBPF Kernel Ground-Truth Seam — `pkg/probe` + `pkg/probe/bpf` as standalone modules

*Date: 2026-11-07 · Scope: `pkg/probe/*`, `pkg/probe/bpf/*`, `docs/VERIFY.md` BLOCKED rows (eBPF, BBRv3, .dist), `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` and `2026-10-27-cgo-ebpf-ground-truth-design.md` · Primary sources: kernel.org BPF docs, `cilium/ebpf` (github.com/cilium/ebpf), libbpf (docs.kernel.org), BTF/CO-RE papers/docs, RFC 8290 · Deep research only — no code changed, no commit.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — FQ-CoDel (Experimental) | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | RFC 8290 is Experimental (§Status) and normative for FQ-CoDel/CoDel vocabulary only; defaults 1024 queues (§1.3), interval 100 ms (§5.2.1), target 5 ms (§5.2.2), quantum 1514 (§5.2.4), flows 1024 (§5.2.5), packet limit MUST (§5.2.3). Not authority for CAKE or eBPF. |
| R2 | BTF kernel docs — BPF Type Format | https://docs.kernel.org/bpf/btf.html | fetched | BTF kernel API (`BPF_BTF_LOAD`, `BPF_MAP_CREATE`, `BPF_PROG_LOAD` §3), ELF `.BTF` / `.BTF.ext` (§4), `/sys/kernel/btf/vmlinux` generation via `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h` via libbpf overview, BTF generation requires llvm ≥8.0 or pahole (§6), function/line info for verifier log (§5.3). |
| R3 | libbpf Overview — kernel docs | https://docs.kernel.org/bpf/libbpf/libbpf_overview.html | fetched | Lifecycle Open → Load → Attach → Tear-down (four phases), skeleton embeds bytecode (“no extra files to deploy”), CO-RE portability via `/sys/kernel/btf/vmlinux` + `vmlinux.h`, `bpf_core_read` relocation. |
| R4 | eBPF verifier — kernel docs | https://docs.kernel.org/bpf/verifier.html | fetched | Two-step verification (DAG + path simulation), register/stack types, `PTR_TO_MAP_VALUE_OR_NULL` NULL-check requirement, direct packet access bounds, error strings (`invalid mem access 'map_value_or_null'`, `misaligned access`). |
| R5 | Linux Kernel Tracepoints — kernel docs | https://docs.kernel.org/trace/tracepoints.html | fetched | Tracepoint is a runtime hook; “off” = branch+space only, “on” = synchronous callback in caller context. |
| R6 | BPF LLVM Relocations — kernel docs (CO-RE) | https://docs.kernel.org/bpf/llvm_reloc.html | fetched | ELF relocations (`R_BPF_64_64` etc.), CO-RE relocation kinds (`BPF_CORE_FIELD_BYTE_OFFSET`, `BPF_CORE_ENUMVAL_VALUE` etc.), record in `.BTF.ext`, `preserve_access_index` examples. |
| R7 | `capabilities(7)` — CAP_BPF / CAP_PERFMON / CAP_NET_ADMIN | https://man7.org/linux/man-pages/man7/capabilities.7.html | fetched | `CAP_BPF` (since 5.8) = privileged BPF; `CAP_PERFMON` (since 5.8) = perf/BPF perf; `CAP_NET_ADMIN` = TOS/interface; root bypasses checks but bounding set still applies; `CAP_SYS_ADMIN` is overloaded legacy. |
| R8 | `github.com/cilium/ebpf` repository | https://github.com/cilium/ebpf | fetched | “pure Go library to read, modify and load eBPF programs and attach them”; packages: `asm`, `cmd/bpf2go`, `link`, `btf`, `rlimit` (RLIMIT_MEMLOCK <5.11), `features`, `pin`; MIT license; supports Linux amd64/arm64, CI on kernel.org LTS. |
| R9 | ebpf-go Getting Started | https://ebpf-go.dev/guides/getting-started/ | fetched | Workflow: C program + `bpf2go` (`//go:generate go tool bpf2go -tags linux counter counter.c`) → generates `*_bpfel.go`/`*_bpfeb.go` + `.o`; Go app must call `rlimit.RemoveMemlock()` for <5.11, `load*Objects()`, `link.Attach*()`, `objs.Close()`; map `Lookup` NULL-gated, atomics for SMP; requires clang, llvm-strip, libbpf-dev, kernel headers, Go; wants `//go:build ignore` on C files. |
| R10 | ebpf-go Loading Objects | https://ebpf-go.dev/concepts/loader/ | fetched | ELF → `CollectionSpec` (`LoadCollectionSpec`) → `NewCollection` (loads maps/progs) or `LoadAndAssign` (selective load into struct); `Collection.Close()` frees maps/progs; `clang -g` emits BTF (`CollectionSpec.Types`, nil if no BTF). |
| R11 | `tcp(7)` — TCP_CONGESTION / TCP_INFO | https://man7.org/linux/man-pages/man7/tcp.7.html | fetched | `TCP_CONGESTION` per-socket string; `TCP_INFO` `getsockopt` returns `struct tcp_info` from `/usr/include/linux/tcp.h`, “not portable”. |
| R12 | `tc(8)`, `tc-cake(8)`, `tc-fq_codel(8)`, `tc-netem(8)`, `tc-tbf(8)` | man7.org | fetched (sibling module) | Shaping = egress, policing = ingress; CAKE `ingress` counts dropped bytes; fq_codel defaults; netem impairs outgoing; TBF pure shaper non-work-conserving. |
| R13 | `go.mod` Go Modules Reference | https://go.dev/doc/modules/gomod-ref | fetched | Module path+version = unique ID + import prefix; `go` directive = minimum version since 1.21; `require path version` SemVer tags immutable; `replace` for migration. |
| — | BBRv3 upstream claims (kernel tree location) | — | `unverified — primary source not found` in this pass (sibling spec quotes `google/bbr` README + branch `Makefile` at `docs/superpowers/specs/2026-10-27-cgo-bbrv3-matrix-design.md:51-60`; not re-fetched here) | BBRv3 matrix provenance. |
| — | Waveform bufferbloat bands, Jain JFI formula, ITU-T G.107 E-model, RouterOS export dialect | — | `unverified — primary source not found` (sibling module `docs/research/2026-11-07-module-probe-audit.md:23-30` already reports same; no primary source fetched) | Inherited unverified neighbours of this seam. |
| — | `ping(8)` output format / Windows `ping -n` semantics | — | `unverified — primary source not found` | `probe.go` parser contract (sibling module repeats; not re-fetched). |
| — | `syscall.TCPInfo.Total_retrans` version invariance | — | `unverified — primary source not found` | `tcpinfo_linux.go` conversion. |

Method: read `pkg/probe/*.go` + `pkg/probe/bpf/bpf_probe.c`, `BUILD.md`, `Makefile:44-59`, `go.mod:1-60`, `docs/VERIFY.md` (BLOCKED rows), `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`, `2026-10-27-cgo-ebpf-ground-truth-design.md`, `2026-10-27-cgo-bbrv3-matrix-design.md`. Every technical claim below cites `file:line` or `(R#)`; anything else is labelled `unverified — primary source not found`.

---

## 1. Why eBPF matters for Master/IP

### 1.1 The thesis needs sender-side ground truth, not a proxy

The frozen corpus is 159 runs / 528 rows / 208 valid / 74 degraded / 246 invalid / 22 cells (`book/tools/audit-archive.py` → OK lines quoted at `docs/VERIFY.md:24` and `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md:16-20`) and is never re-measured (umbrella spec §0, invariant 3). Every *new* measurement must therefore be defensible in front of a Master jury that already knows the proxy flaw:

- The weak link is the retransmit numerator. `pkg/probe/tcpinfo_linux.go:15` hardcodes `retransMSS = 1448` and `pkg/probe/tcpinfo_linux.go:43` returns `uint64(info.Total_retrans) * retransMSS`. The file header says it plainly: “TCP_INFO rend des SEGMENTS réémis ; ×1448 les convertit en octets comme le fait le mémoire (wasted = drops × 1448). Proxy documenté, pas une mesure par paquet.” The same method is pinned by umbrella-spec decision #9: “Linux `TCP_INFO` (`tcp_total_retrans` × MSS) on cgo's own bulk sockets → `measured`. Probe-loss proxy → `estimated`.” The ground-truth spec §0 expands: that proxy “is wrong under offload, fragmentation, or a non-1448 MSS” because it is a segment count multiplied by a guessed MSS and not a byte count per packet. No jury will accept a Master claim whose Model A numerator is a guessed MSS under GSO/TSO/GRO (kernel keeps super-packets up to 64 KiB TSO / ~25 KiB GRO — R1 §5.2.4 notes TSO/GRO “mega-packets severely impact … scheduling” and offload skews packet units).

- Model A/B economics inherit the flaw. Umbrella spec §4 defines `CapacityWasteARPerHour = (wasted / transported) × (MonthlyOpEx / 720)` and `MeteredWasteARPerHour = (wasted_bytes / 2^30) × tariff × (3600 / windowSeconds)` where `wasted` derives from decision #9. If the ratio is a proxy, the Ar/h figure the client pays is a proxy. The 2026-10-27 ground-truth spec §§0.1–0.3 and §§4.2–4.3 therefore supersede decision #9 “for hosts where eBPF is available” while preserving the honesty rule that `measured` (TCP_INFO) is never confused with `estimated` (proxy), and add `source: ebpf | tcpinfo | ""` + `basis: measured | estimated | ""` and a `kernel_telemetry` report block (§4.5). The economics bullet survives only if the source is recorded per window.

- Bufferbloat / AQM claims need drop truth per class. RFC 8290 defines FQ-CoDel as “a combined packet scheduler and AQM algorithm” (R1 §1) with per-queue CoDel (R1 §2 normatively referring to RFC 8289), flow hashing on 5-tuple + salted Jenkins (R1 §4.1), DRR new/old two-tier (R1 §4.2), defaults interval 100 ms (R1 §5.2.1), target 5 ms (R1 §5.2.2), quantum 1514 (R1 §5.2.4), flows 1024 (R1 §5.2.5), and the only MUST in §5.2 is “some packet limit MUST be enforced” (R1 §5.2.3). It defines *none* of CGO’s probes, phases, report schema, or 220 ms/95 % CAP objective. Its legitimate use is framing why loaded latency is measured at all. CAKE has no RFC (R1 §10.2 cites CAKE only informatively as “currently being developed” future work, R1 §7). A Master thesis that cites RFC 8290 as CAKE authority is citation theatre; a thesis that measures EF vs BE under CAKE/fq_codel with kernel drop counters is experiment. eBPF gives that drop leg (see §1.2). A .dist / shape thesis (dwell) cannot be true without per-packet sojourn — which is why eBPF is the *precondition* for the `.dist` slice even though `.dist` itself is blocked (VERIFY row at `docs/VERIFY.md:34`, see §2).

### 1.2 What “ground truth” concretely means here (and what it deliberately does not)

The ground-truth spec (§§2–3) narrows three exactness claims, each anchored in kernel facts:

| Signal | Hook (C source) | Unit claimed | Not claimed |
|---|---|---|---|
| Retransmit bytes | `raw_tp/tcp_retransmit_skb` — `pkg/probe/bpf/bpf_probe.c:111-147`. Comment `bpf_probe.c:6-9`: “`skb->len` est déjà le payload TCP, TSO compris, et `transport_header` porte encore le sentinelle `0xFFFF`. Ne rien soustraire.” TP_PROTO gives `(sk, skb)` so raw TP reads final `skb->len`. Spec §2.1 notes `tcp_transmit_skb(clone_it=1)` pushes headers on the clone; queued skb keeps payload, so `payload = skb->len`. Kernel itself does `tp->bytes_retrans += skb->len` in the same function (independent unit confirmation, spec §2.3). | Submitted retransmit payload bytes, successful transmits only, after fragmentation/collapse (TSO-inclusive). | ACKed/delivered bytes; failed attempts (no tracepoint fire); wire bytes incl. headers (not observable at this hook, spec §2.3 — stored as no counter, not estimated). |
| Socket state | `fentry/tcp_rcv_established` primary + `kprobe/tcp_rcv_established` fallback — `bpf_probe.c:175-185`. `sample_sock` (`bpf_probe.c:149-173`) reads `srtt_us >> 3` (kernel stores left-shifted by 3), `snd_cwnd`, `snd_ssthresh`, `skc_state`, throttled to 50 ms (`SAMPLE_NS 50000000ULL`, `bpf_probe.c:27`). | Sampled `srtt_us` (unshifted), `snd_cwnd`/`snd_ssthresh` in segments, state, ≤20 Hz per socket. | Per-packet RTT. The sample is “monitoring, never per-packet” (spec §2.4). |
| Qdisc drops | `raw_tp/kfree_skb` — `bpf_probe.c:187-251`. Filters `reason == SKB_DROP_REASON_QDISC_DROP` via `bpf_core_enum_value` (`bpf_probe.c:195-197`), parses IPv4/TCP headers (`bpf_probe.c:211-242`), attrs per-flow only if flow entry already exists, else global only (`bpf_probe.c:244-249`). Linux 6.8 has **no** `qdisc_drop` tracepoint (introduced v7.1, commit `ff2998f29f39` — spec §2.1); 6.8’s `__qdisc_drop()` in `include/net/sch_generic.h` defers to `kfree_skb_list_reason()` → `__kfree_skb_reason()` with that reason, so `kfree_skb` is the correct 6.8 hook. On 7.1+ native `qdisc_drop` can replace it in a later change (spec §§2.1, 7.12). | Host-wide qdisc-drop packets/bytes (`global_stat` at `bpf_probe.c:52-58`), per-flow drop packets/bytes when 4-tuple parses and flow exists. | Queue dwell / sojourn / queue delay — “no enqueue/dequeue timestamp programs and no skb-pointer correlation map” (spec §2.6). Drop timing is the free event, possibly batched, not an enqueue timestamp (spec §7.6). |

Tracepoint semantics support this split: when off, only branch+space cost; when on, the probe runs synchronously in caller context (R5 §Purpose). “Zero userspace overhead” in spec §2.1 means *no per-drop syscall* (counters updated in-kernel, read at window end) — not zero cost; the 2 % overhead gate (§6.5) bounds it.

Maps nail the census: `flows` = `BPF_MAP_TYPE_HASH` max 4096 (`bpf_probe.c:60-65`), `globals` = `BPF_MAP_TYPE_PERCPU_ARRAY` max 1 (`bpf_probe.c:67-72`). IPv6/non-IPv4 retransmits and map-full overflows increment `global_stat` so totals never silently drop (`bpf_probe.c:121-126`, `133-141`). Per-flow key omits netns — a same-tuple collision across namespaces merges counters. Spec §3.2 names this a documented bench-only `shortcut` to upgrade to `(netns, tuple)` if background traffic can collide.

### 1.3 Why the BLOCKED rows make this the Master/IP bottleneck

`docs/VERIFY.md` currently freezes three interlocking claims — and eBPF is the only one whose *Go half* already landed but whose *live half* cannot:

| VERIFY row | Status | What it gates |
|---|---|---|
| “eBPF retransmit / RTT / qdisc drops” (`docs/VERIFY.md:29`) | **BLOCKED** — “No committed `bpf_probe.o` and no Linux 6.8 run; Windows cannot build the object. The CO-RE source and `make bpf` target landed 2026-10-29, but the object is not built: CGO-DEV VM SSH rejects every recorded key/user, PC2 WSL is 6.18 without clang/bpftool/tc.” | Every `source=ebpf` sentence in a report. |
| “eBPF CO-RE source and maintainer build target” (`docs/VERIFY.md:30`) | **RECORDED** — `bpf_probe.c`, `BUILD.md`, `vmlinux.h` gitignore, `make bpf`/`make bpf-check` committed 2026-10-29; `make` absent on Windows host; “No object, no hash, no loader yet; do not commit an object until `make bpf` and `make bpf-check` run on Linux 6.8.” | Provenance chain. |
| “Linux live acceptance (G0–G7, campaign)” (`docs/VERIFY.md:28`) + “IFB ingress shaping — Linux 6.8 live acceptance” (`docs/VERIFY.md:31`) + “PC2 / nemo live bench access” (`docs/VERIFY.md:56`) | **BLOCKED** (6.8 bench unreachable; CGO-DEV SSH reject; PC2 WSL 6.18.54.1-WSL2 lacks `tc`/BBR/clang/bpftool) | Bench truth for *all* live cells. |
| “BBRv3 3×3 matrix” (`docs/VERIFY.md:33`) | **BLOCKED** — “No 7.1 kernel; VM exposes only reno, cubic, bbr.” Engine row (`docs/VERIFY.md:52`) is BLOCKED on the same 7.1 kernel, and umbrella invariant #8 (BBR = BBRv1 in 6.8 mainline) stays binding until that slice supersedes it. | Third CC row. |
| “`.dist` netem” (`docs/VERIFY.md:34`) + engine row (`docs/VERIFY.md:53`) | **BLOCKED** — “No field trace at or above the 30-sample floor; two SMOKE-TEST traces are 19 and 25 rows.” | Trace-driven calibrator. |
| “eBPF Go half — contract, fallback ladder, report block, health source, ribbon badge, readiness script” (`docs/VERIFY.md:58`) | **VERIFIED 2026-10-30** — `go build ./... && go test ./pkg/probe/ ./pkg/probe/bpf/ ./pkg/audit/ ./pkg/api/ -count=1` exit 0; ladder tests cover every rung (Windows, absent object, no BTF, no CAP_BPF, ready); `/api/health` returns `{"source":"tcpinfo","reason":"hôte non-Linux"}`; `bash kit/host-readiness.sh` prints n/a rows on Windows. “No object exists and nothing was attached: the slice-4 live row stays BLOCKED.” | Windows side is done. |

Consequence for the Master: the thesis is currently *legible* as honest engineering (contract + fallback + report block verified) but *illegible* as measurement science (no `source=ebpf` cell has ever run). The eBPF ground-truth seam is therefore the critical path: it is the only BLOCKED row whose clearing turns a `tcpinfo` proxy column into a kernel byte column; BBRv3 and `.dist` both ride on top of it (BBRv3’s 27-event 3×3 at `docs/superpowers/specs/2026-10-27-cgo-bbrv3-matrix-design.md:148-164` needs per-flow retransmit truth under loss; `.dist` needs a drop/sojourn corpus that is not SMOKE-TEST-sized). IP-wise, the seam’s product is the *fallback ladder + provenance* pattern (typed `UnavailableError` vs `BuildError`, SHA-256-pinned object, coverage-ranked ratio), which is the part a jury can reuse even if the kernel never cooperates.

---

## 2. Exported API (file:line)

Complete inventory of exported identifiers in non-test sources under `pkg/probe` (incl. `bpf`). Unexported helpers (`parsePing`, `parsePingLine`, `smallMarkedTransport`, `keepConn`, `bulkNStats`, `kernelFacts`, `socket_key`, `flow_get_or_create`, etc.) are not API.

### 2.1 `pkg/probe` (non-test)

| Export | Kind | Location | Contract (as written) |
|---|---|---|---|
| `PingSample{RTTms}` | struct | `pkg/probe/probe.go:22` | One completed ping RTT, ms. |
| `Ping(ctx,target,count,intervalMs)` | func | `pkg/probe/probe.go:28-43` | Exec ping; `-c/-i` Unix, `-n` Windows (1/s, no `-i`); timeout = `count×interval+5000ms`; returns `parsePing(out)` even on non-zero exit when output exists (`len(out)==0` only error). |
| `SmallClient()` | func | `pkg/probe/probe.go:93-96` | Shared 2 s client (`once`+`markedClient` at `probe.go:108-111`); transport marks DSCP EF on Linux. |
| `SmallClientTimeout(timeout)` | func | `pkg/probe/probe.go:104-106` | Fresh client per call (bench use); same marked transport; `max(2s,4×delay)` bench rule in comment. |
| `SmallObject(ctx,client,url)` | func | `pkg/probe/probe.go:122-139` | Times GET→`io.Discard` in ms (`us/1000.0`); returns `0,err` on request/do/copy error. |
| `BulkSendTo(ctx,addr,cc)` | func | `pkg/probe/bulk.go:11-17` | `DialWithCC` + `BulkSend`. |
| `BulkSend(ctx,conn)` | func | `pkg/probe/bulk.go:20-40` | 64 KiB zeroes, 2 s write deadline, until `ctx.Done`; `Write` error → `return total, nil` (receiver-closed = end). |
| `BulkReceive(ctx,conn,onBytes)` | func | `pkg/probe/bulk.go:43-75` | 256 KiB reads; cancel via `SetReadDeadline(pastTime())` (`bulk.go:77`, `pastTime` = epoch). |
| `BulkSource(ctx,conn)` | func | `pkg/probe/bulksource.go:14-31` | Download-direction mirror of `BulkSend`. |
| `BulkRecvSource(ctx,conn,cc)` | func | `pkg/probe/bulksource.go:39-49` | Writes `"D"` or `"D:<cc>\n"` then `BulkReceive`. |
| `BulkDownloadTo(ctx,addr,cc)` | func | `pkg/probe/bulksource.go:55-61` | Dial + `BulkRecvSource`. |
| `BulkDownloadNTo(ctx,addr,cc,n)` | func | `pkg/probe/bulksource.go:68-90` | N parallel dials, per-flow `[]uint64` in dial order. |
| `BulkDownloadNStats(ctx,addr,cc,n)` (with `BulkSendNStats`) | func | `pkg/probe/bulkstats.go:60-73` | `bulkNStats`: dial-failures skipped (`bulkstats.go:36-38`); `keepConn` withholds `Close` so `TCP_INFO` sees live fd (`bulkstats.go:17-25`); sums `RetransBytes`, sets `oneTCP` if ≥1 socket sampled (`bulkstats.go:43-56`). |
| `ParseDownloadHello(b)` | func | `pkg/probe/bulksource.go:94-103` | Tolerant `"D"/"D:<cc>"`, spaces/`\n`, prefix case. |
| `FlowResult{Bytes}` | struct | `pkg/probe/bulkflows.go:14-16` | Per-flow contribution (JFI). |
| `BulkSendNFlows(ctx,conns,_)` | func | `pkg/probe/bulkflows.go:23-36` | N conns flooded; third arg ignored. |
| `BulkSendNTo(ctx,addr,cc,n)` | func | `pkg/probe/bulkflows.go:42-64` | `""`→`model.Cubic` (`bulkflows.go:44`); `n<=1` single-flow. |
| `BulkStats{Bytes,RetransBytes,RetransMeasured}` | struct | `pkg/probe/bulkstats.go:11-15` | Per-flow bytes + summed retrans + whether any TCP socket sampled. No `Source`/`Kernel` field yet — ground-truth spec §4.2 proposes adding `Source string` + `Kernel *KernelTelemetry`. |
| `DialWithCC(ctx,addr,cc)` | func | `pkg/probe/bulk_cc_linux.go:14-28`; stub `pkg/probe/bulk_cc_other.go:11-14` | Linux: `net.Dialer{5s}` + `TCP_CONGESTION` control (`tcpCongestion=0x0d`); else: plain dial, `cc` ignored. (R11) §`TCP_CONGESTION` allows per-socket string. |
| `SetConnCC(conn,cc)` | func | `pkg/probe/ccsock_linux.go:15-34`; stub `pkg/probe/ccsock_other.go:10` | Accepted-socket CC; non-TCP/`""` → `nil` silently. |
| `TCPRetransBytes(c)` | func | `pkg/probe/tcpinfo_linux.go:20-44`; stub `pkg/probe/tcpinfo_other.go:14` | Linux: `*net.TCPConn`→`File()`→`getsockopt(TCP_INFO)`→`Total_retrans×1448`; non-TCP/non-Linux → error. Raw syscall only (stdlib comment rejects `golang.org/x/sys`, `tcpinfo_linux.go:30-35`). |
| `ErrTCPInfoUnavailable` | var | `pkg/probe/tcpinfo_other.go:12` | Non-Linux error. (R11) §`TCP_INFO` says kernel returns `struct tcp_info` from `/usr/include/linux/tcp.h`; exact `Total_retrans` field presence across kernels is `unverified — primary source not found`. |
| `KernelSourceInfo{Source,Reason}` | struct | `pkg/probe/kernelsource.go:12-15` | `Source` ∈ `ebpf`/`tcpinfo`/`""` (impl. also `unavailable` in sibling docs — current code returns only `ebpf`/`tcpinfo` via `bpf.Status`). |
| `KernelSource()` | func | `pkg/probe/kernelsource.go:34-38` | Runs `kernelFacts` (`kernelsource.go:19-31`) → `bpf.Status`. `kernelFacts` checks `/sys/kernel/btf/vmlinux` via `os.Stat` and approxes caps as `os.Geteuid()==0` (`kernelsource.go:26-30`). |

### 2.2 `pkg/probe/bpf` (non-test)

| Export | Kind | Location | Contract |
|---|---|---|---|
| `Flow{Local,Peer,RetransPayload,RetransEvents,DropPackets,DropBytes,SRTTus,SndCwnd,SndSsthresh,State,Samples}` | struct | `pkg/probe/bpf/types.go:4-16` | One sender-side flow as kernel maps saw it (JSON tags snake_case). |
| `Summary{Available,Source,Attach,FallbackReason,ObjectSHA256,FlowsSeen,FlowsTotal,RetransPayload,RetransEvents,QdiscDrops,QdiscDropBytes,QdiscAvailable,MapFull,Flows}` | struct | `pkg/probe/bpf/types.go:20-35` | One telemetry window; `Source` ∈ `ebpf`/`tcpinfo`/`""` (types.go:19). |
| `Globals{RetransPayload,RetransEvents,DropPackets,DropBytes,MapFull,QdiscAvailable}` | struct | `pkg/probe/bpf/types.go:40-47` | Host globals (IPv6/non-IPv4 + full-table fallback). |
| `Status(goos,btf,caps,objectErr)` | func | `pkg/probe/bpf/status.go:11-25` | Fallback ladder, injected args so Windows tests each rung. Returns `tcpinfo` with French reason (`hôte non-Linux` / object error / `BTF absent` / `CAP_BPF absent`) or `ebpf`. |
| `ObjectSHA256()` | func | `pkg/probe/bpf/status.go:28-35` | `sha256(ObjectBytes)` hex (provenance). |
| `ObjectBytes()` | func | `pkg/probe/bpf/assets_embedded.go:10-15` (`//go:build ebpf_object`, `//go:embed bpf_probe.o`); stub `pkg/probe/bpf/assets_absent.go:7-9` (`//go:build !ebpf_object` returns `BuildError`) | Embedded vs absent object. |
| `UnavailableError{Reason}` / `BuildError{Reason}` | structs + `Error()` | `pkg/probe/bpf/errors.go:4-11` | Typed host-limitation vs object-defect errors (`ebpf indisponible : …` / `ebpf : objet invalide : …`). |

### 2.3 Build assets (`pkg/probe/bpf/`)

| File | Role | Location |
|---|---|---|
| `bpf_probe.c` | CO-RE C source (GPL-2.0). Hooks `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established` + `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb`. Uses `vmlinux.h`, `<bpf/bpf_helpers.h>`, `<bpf/bpf_tracing.h>`, `<bpf/bpf_core_read.h>`, `<bpf/bpf_endian.h>`. | `bpf_probe.c:1-253` |
| `BUILD.md` | Maintainer build pin: “Source et cible posées 2026-10-29. L’objet n’est pas encore construit” (`BUILD.md:8-12`), clang-18 / libbpf-dev / bpftool / Ubuntu 24.04 / kernel 6.8.0 `/sys/kernel/btf/vmlinux` (`BUILD.md:16-19`), `make bpf` generates `vmlinux.h` → CO-RE compile → `bpf_probe.o.sha256` (`BUILD.md:29-31`), verify via `readelf -S … \| grep -E 'raw_tp\|fentry\|kprobe'` (`BUILD.md:35-41`), DWARF embeds source path so other dirs change hash (`BUILD.md:45-48`). | `BUILD.md:1-48` |
| `assets_embedded.go` / `assets_absent.go` | Build-tag gate for embedding. | `assets_embedded.go:1-15`, `assets_absent.go:1-9` |
| `Makefile` targets | `bpf` (bpftool dump → clang-18 `-O2 -g -target bpf -D__TARGET_ARCH_x86` → `sha256sum` → `.sha256`) and `bpf-check` (sha256 compare). | `Makefile:44-59` |

---

## 3. Dependencies

### 3.1 Outgoing (what the probe seam needs today)

- **Stdlib only for the probe.** `pkg/probe/probe.go:6-19`, `bulk.go:3-8`, `bulkstats.go:3-7`, `bulksource.go:3-9`, `bulkflows.go:3-9`, `bulk_cc_linux.go:5-10`, `tcpinfo_linux.go:5-10`, `dscp_linux.go:5-10`, `ccsock_linux.go:5-8` are all stdlib. `tcpinfo_linux.go:30-35` explicitly rejects `golang.org/x/sys` in favour of raw `syscall` + `SYS_GETSOCKOPT`. None of `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed for the probe itself.
- **Two local edges.** `pkg/model` only for the `Cubic` default string (`pkg/probe/bulkflows.go:8`, `bulkflows.go:44`) and `pkg/probe/bpf` for the ladder (`pkg/probe/kernelsource.go:7`, `kernelsource.go:36`). `pkg/probe/bpf/status.go:3-6` imports only `crypto/sha256,encoding/hex`; `assets_embedded.go:5` imports `embed`. Module is `github.com/Realms4239/cgo` at `go 1.25` (`go.mod:1-3`) — per (R13) §"module" path+version is the unique ID and import prefix; a standalone needs a new controlled path; per (R13) §"go" the `go` directive is a mandatory minimum since 1.21.
- **Build axes.** `linux` vs `!linux` for CC/DSCP/TCP_INFO (`bulk_cc_linux.go:1`/`bulk_cc_other.go:1`, etc.) and `ebpf_object` for embedded vs absent BPF object (`assets_embedded.go:1`/`assets_absent.go:1`). The spec’s fuller loader axis (`linux && amd64` real loader vs stub, ground-truth spec §3.5/§4.1 `Check()`/`Start()`/`Lookup`/`Snapshot`/`Globals`/`Attach`/`Close`) is **designed but not present** — there is no `session_linux.go`/`session_stub.go` today.
- **Build toolchain (maintainer-only, Linux 6.8).** Per `BUILD.md:16-20` and `Makefile:46-56`: Ubuntu 24.04, kernel 6.8.0, clang-18 (`BPF_CLANG=clang-18`), libbpf-dev headers (`<bpf/*.h>`), bpftool (`vmlinux.h` from `/sys/kernel/btf/vmlinux` per R2/R3), `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h` (R3). `.BTF`/`.BTF.ext` in the object require clang ≥8.0 or pahole (R2 §6). `vmlinux.h` is gitignored (`BUILD.md:29-31`); only the object + `.sha256` is the deliverable.
- **Missing runtime dependency.** `go.mod` has **no** `github.com/cilium/ebpf` (or libbpf binding) — the loader and map-read code do not exist. Ground-truth spec §3.5 says `go.mod` *will* pin a specific cilium/ebpf release for `session_linux.go`; today that pin is absent. Adding it is a new-dependency stop-and-ask per repo `AGENTS.md`.
- **Host capabilities for live attach.** Per (R7) §`CAP_BPF`: `CAP_BPF` (since 5.8) = privileged BPF operations; §`CAP_PERFMON` (since 5.8) covers perf/BPF perf ops; §`CAP_NET_ADMIN` covers `set type-of-service (TOS)` and interface config; §Notes warn `CAP_SYS_ADMIN` is overloaded (“the new root”) and should be avoided for new features. Ground-truth spec §3.4 therefore requires `CONFIG_DEBUG_INFO_BTF` + `/sys/kernel/btf/vmlinux` + `CAP_BPF` + `CAP_PERFMON` (+ `CAP_NET_ADMIN` already for `tc`); root is the supported path, `sudo setcap cap_bpf,cap_perfmon,cap_net_admin+ep ./cgo` is the documented fallback. `kernelFacts` (`kernelsource.go:29-30`) approximating caps as `euid==0` over-claims vs (R7) — root bypasses checks *unless* the bounding set dropped the cap, and a uid-0 process without the cap in its bounding set is still gated (R7 §Programmatically adjusting capability sets; §Capability bounding set). Real capget is a gap (§5.4).

### 3.2 Incoming (who needs the seam)

- `docs/architecture.md:60-74` gates G0–G7 on probe outputs (`BulkStarted`, `ProbesProducing`, …). Umbrella spec §7 notes remediation is a separate layer (`Link{CapacityMbps,RTTMs,Iface,AuditProvenance}`); the audit/report consumes probe stats but must not import the loader directly — ground-truth spec §4.1 aliases `probe.KernelTelemetry = bpf.Summary` so `pkg/audit` never imports the loader.
- Sibling modules (`pkg/audit`, `pkg/metrics`, `pkg/campagne`) call `probe.` / `audit.` — extraction must keep every call site compiling or provide an adapter (sibling `2026-11-07-module-probe-audit.md:169`).

### 3.3 Licence

`LICENSE` at repo root is MIT (per sibling module `2026-11-07-module-probe-audit.md:173`; `cilium/ebpf` is MIT — R8 §License). BPF C object is GPL-2.0 (`bpf_probe.c:1` vs cilium example `Dual MIT/GPL` in R9 §10-11: GPL-licensed helpers require GPL-compatible BPF program licence). A standalone observer must keep both licences with their files; do not relicense the object.

---

## 4. Primary-source requirements

What each primary source *legitimately* requires of a releasable eBPF observer module — and what it does not.

### 4.1 Kernel BTF + CO-RE (R2, R3, R6)

- BTF has two parts — kernel API + ELF file format (R2 §1). Kernel API involves `BPF_BTF_LOAD` → `BPF_MAP_CREATE`/`BPF_PROG_LOAD` with `prog_btf_fd`/`func_info`/`line_info` (R2 §3). ELF carries `.BTF` (§4.1) + `.BTF.ext` (§4.2) with `core_relo` sub-section; see (R6) for LLVM relocation types and CO-RE relocation kinds/records. Function/line info enables “source annotated translated byte code, jited code and verifier log” (R2 §1, §5.2–§5.3).
- CO-RE = Compile Once – Run Everywhere (R3 §BPF CO-RE). Producing it needs the running kernel’s authoritative BTF at `/sys/kernel/btf/vmlinux` → `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h` (R3); including `vmlinux.h` eliminates system-wide kernel-header dependency. libbpf resolves recorded BTF relocations against runtime BTF and patches offsets at load. `BPF_CORE_READ` records field info for relocation (R3 example); generic helpers are in `bpf_helpers.h`; see R9 annotations 2–3 (`__uint`/`__type`/`SEC` from libbpf) on the C program.
- Concrete requirements for this module: (a) commit a prebuilt `bpf_probe.o` + pinned SHA-256 + `BUILD.md` with clang version, libbpf headers, BTF source, exact command (ground-truth spec §§3.5, 8); (b) keep `vmlinux.h` generated + gitignored; (c) use `-g` and CO-RE relocations so `BPF_CORE_READ`/`bpf_core_enum_value` offsets are patched at load instead of hardcoded (the C file already does this — `bpf_probe.c:105-108`, `195-197`); (d) make the build reproducible — same toolchain + same BTF + same workdir must reproduce the pinned SHA, else investigate (spec §3.5); DWARF embeds source paths, so other dirs change the hash (`BUILD.md:45-48`); (e) verify sections `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`, `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb` via `readelf` (`BUILD.md:35-41`) and pin SHA via `TestEmbeddedObjectPinned` without `make`.
- **What BTF does not do**: BTF is debug/metadata, not magic offsets. It does not make a verifier-unsafe program load, nor make `tcpinfo` byte-exact. A missing CO-RE field fails the whole collection load (spec §3.3) and falls back to TCP_INFO — that is the honest behaviour.

### 4.2 libbpf / cilium/ebpf loader lifecycle (R3, R8, R9, R10)

- libbpf’s four phases (R3 §BPF App Lifecycle): **Open** (parse ELF, discover maps/progs/globals) → **Load** (create maps, resolve relocations, verify+load progs; no prog executed yet) → **Attach** (to tracepoints/kprobes/cgroup/packet hooks; work happens here) → **Tear down** (detach, unload, destroy maps). Skeleton (`<name>__open/__load/__attach/__destroy`) embeds bytecode so “no extra files to deploy” (R3 §BPF Object Skeleton).
- ebpf-go mirrors this natively (R10): ELF → `LoadCollectionSpec` → `NewCollection` (loads maps+progs, `Collection.Close()` frees them) or `LoadAndAssign` (selective load into struct, `Close()` pattern). R9 shows the canonical Go app: `rlimit.RemoveMemlock()` for <5.11 (R8 §`rlimit` package), `loadCounterObjects(&objs,nil)`, `link.AttachXDP`/`AttachTracepoint`/… (R8 §`link`), `objs.Close()` + `link.Close()`, map `Lookup` after null-check. R8 also ships `features` (`bpftool feature probe` equivalent), `btf` reader, `pin` (bpffs), and `cmd/bpf2go` (compile C + generate Go scaffolding + embed objects; endianness variants `*_bpfel.go`/`*_bpfeb.go` in R9).
- Concrete requirements: (a) pin a specific cilium/ebpf release in `go.mod` (spec §3.5) — today it is absent (§3.1); (b) implement Open/Load/Attach/Tear-down in Go with a verifier log captured (R4 below); (c) `rlimit.RemoveMemlock()` best-effort on 6.8 memcg accounting (spec §3.3); (d) per-session `Collection` (no global session state) with concurrent-safe `Lookup`/`Snapshot`; (e) build-tag split: real loader `//go:build linux && amd64`, stub elsewhere with identical exported API (spec §3.5, §4.1). R9 also requires `//go:build ignore` on C files co-located with Go to avoid “C source files not allowed when not using cgo”, plus tool dependency `go get -tool github.com/cilium/ebpf/cmd/bpf2go`.
- **What the loader does not do**: libbpf/ebpf-go loads and attaches; it does not author BTF, nor validate BPF C correctness beyond the kernel verifier. See §4.3.

### 4.3 Verifier + tracepoint + capability discipline (R4, R5, R7)

- The verifier does DAG (disallow loops, unreachable insns) + path simulation (register/stack types) (R4 §eBPF verifier). It enforces NULL-checked `map_lookup_elem` before access — `PTR_TO_MAP_VALUE_OR_NULL` becomes `PTR_TO_MAP_VALUE` only when checked `!= NULL` (R4 §Register value tracking); un-checked access yields `R0 invalid mem access 'map_value_or_null'`, misaligned access yields `misaligned access off …` (R4 §Understanding messages). The C code already respects this: `flow_get_or_create` handles `BPF_NOEXIST` + `map_full` (`bpf_probe.c:80-99`) and callers handle NULL (`bpf_probe.c:83-99`, `132-141`).
- Tracepoints are runtime hooks: off = branch+space only; on = synchronous callback in caller context (R5). The three hooks therefore run in transmit/ACK/free context with the target CPU’s preemption state — no per-packet userspace syscall is needed, only end-of-window map reads (spec §2.1).
- Capabilities (R7, §3.1): `CAP_BPF`/`CAP_PERFMON` are the *weakest* caps that cover privileged BPF / perf-BPF; `CAP_SYS_ADMIN` is overloaded and exists for legacy. A loader must *check* them (capget) and fall back with a typed reason, not approximate via euid or `stat` alone. Doctor must not claim attach capability from a load-only `Check()` (spec §4.1 `Check()` = verify hash then load-only; doctor detail says “loadable (BTF + CAP_BPF); attach is verified by audit”).
- **What these do not do**: the verifier proves safety, not metric correctness. ახმიენტობა (verify + load ≠ attach) and register tracking do not prove `skb->len` is payload — that is a *semantic* claim grounded in `tcp_transmit_skb` / `tp->bytes_retrans` code, cited at spec §§2.1–2.3, not in the verifier docs. Keep them separate in the thesis.

### 4.4 RFC 8290 + tc(8) family (R1, R12)

- Per §1.1 above, RFC 8290 legitimately frames *why* FQ-CoDel/CoDel queue control matters ( target/interval vocabulary for loaded-latency measurement). It does not normatively define ping, bulk, audit phases, EFI/DSCP marking, CAP 220 ms, or any report schema. CAKE has no RFC (R1 §10.2 informal). A releasable module must cite RFC 8290 only for FQ-CoDel mechanics, cite `tc-cake(8)` (R12) for CAKE `ingress`/`diffserv4`/`ack-filter`, and cite kernel docs for eBPF — never cross-cite.
- tc(8) vocabulary (R12, sibling-checked): shaping is egress, scheduling egress, policing ingress; CAKE `ingress` counts already-traversed dropped bytes; netem impairs outgoing; TBF is pure non-work-conserving shaper. A standalone observer must *document generated `tc` strings* against (R12) syntax and must *not* claim to apply shaping — `pkg/qdisc` owns application; the observer only *reads* the kernel after `tc` acted.

### 4.5 Go module hygiene (R13)

- A standalone `cgo-ebpf-observer` lives behind its own `go.mod` + controlled path (R13 §"module"); v2+ needs `/v2`; pre-1.0 (`0.y.z`) may break API, 1.0.0 is the promise. `require` is `path version` (SemVer tags, immutable once pushed); `go` directive is a minimum since 1.21 (R13 §"go"). `ebpf_object` `embed` and raw `syscall` need no `require`; only the new loader dep forces `require` + `replace` during migration if paths change.

---

## 5. Gap analysis

Ordered by severity. Local evidence cites `file:line`; primary side cites `R#` or is `unverified — primary source not found`.

1. **Object-less and loader-less — the `ebpf` tier cannot run.** `BUILD.md:8-12`: “L’objet n’est pas encore construit : aucun hôte Linux 6.8 stable n’était joignable.” `go.mod:5-10` has no eBPF loader dependency. Go has only `ObjectBytes` (`assets_embedded.go:10-15` / `assets_absent.go:7-9`), `Status` (`status.go:11-25`), `ObjectSHA256` (`status.go:28-35`) — none of the four libbpf phases (Open→Load→Attach→Tear-down, R3) exists in Go. Spec §4.1’s designed API (`Check`, `Start`, `Lookup`, `Snapshot`, `Globals`, `Attach`, `Close`) is absent on disk (`glob pkg/probe/bpf/*` returns only `types.go`, `status.go`, `errors.go`, `assets_*.go`, `bpf_probe.c`, `BUILD.md`, tests). `KernelSource` (`kernelsource.go:34-38`) will therefore always fall to `tcpinfo` (`status.go:15-17` via `BuildError`, `assets_absent.go:8`). Until the object + SHA are built on the pinned 6.8/clang-18/libbpf/bpftool stack (`BUILD.md:16-31`) and a loader attaches the three hooks and reads `flows`/`globals`, every `ebpf` claim is toy (see §8). This is the sole reason VERIFY `docs/VERIFY.md:29` stays BLOCKED.
2. **Host-fact heuristics lie about caps and BTF.** `kernelFacts` (`kernelsource.go:19-31`) uses `os.Stat("/sys/kernel/btf/vmlinux")` alone for BTF and `os.Geteuid()==0` for caps. Per (R7) root bypasses checks *unless* the bounding set was narrowed (R7 §Capability bounding set, §Programmatically adjusting), so euid==0 over-claims; per (R2)/(R3) BTF *presence* ≠ parseable BTF (the file is authoritative only when readable and the kernel built it — `CONFIG_DEBUG_INFO_BTF` per spec §3.4). Fix: parse BTF (via `cilium/ebpf/btf` R8 package or `bpftool` READ, not stat) and use `capget` to test `CAP_BPF`/`CAP_PERFMON` in the effective set; the doctor `warn` detail must name the missing fact (spec §4.5).
3. **Coverage contract is not yet in code.** `bulkstats.go:29-56` (`bulkNStats`) has no `Source`/`Kernel` plumbing; `BulkStats` (`bulkstats.go:11-15`) has no `Source`/`Kernel` field; there is no `kernelSession` interface seam (`Start()` factory) from spec §4.2, no sender-side key derivation (upload = client `LocalAddr/RemoteAddr`, download = reversed `RemoteAddr/LocalAddr`), no query-before-real-close before `keepConn` release, and no all-or-nothing gate (“require *all* dialed flows observed with ≥1 sample/event; qdisc drop alone does not prove retransmit accounting; missing ≠ zero; exact zero has a sample” — spec §4.2). `BulkSendNStats`/`BulkDownloadNStats` (`bulkstats.go:60-73`) still sum TCP_INFO only. Any extraction must move this logic into the new module (§6), not duplicate it in `pkg/audit`.
4. **Probe fidelity gaps carry into the observer if left as-is.**
   - *Ping by exec + locale parsing.* `parsePing`/`parsePingLine` (`probe.go:49-85`) tolerates `temps=`/`temps<` and `<1ms`→`v/2` (`probe.go:81-83`), but the ping-output contract has no primary source (`unverified — primary source not found`); `Ping` returns partial output as success when output exists even on non-zero exit (`probe.go:38-43`), and `parsePing` sorts, losing per-packet order. Windows `-n` has no `-i` (`probe.go:31-34`) — interval silently ignored.
   - *Small-object timeout triad.* Shared 2 s (`probe.go:93-96`) vs fresh bench client (`probe.go:104-106`) vs audit default 8 s (`pkg/audit/audit.go:231` in sibling) — three timeouts, no rule; `SmallClientTimeout` allocates fresh `Transport` per call (`probe.go:105`), defeating reuse.
   - *Bulk payload heuristics.* Zero-filled 64 KiB writes (`bulk.go:22`) measure compressible-zero goodput (no entropy; fill policy `unverified — primary source not found`); 2 s write deadline (`bulk.go:33`) is comment-justified only; `BulkSend` swallows `Write` errors as clean end (`bulk.go:36-38`); `bulkNStats` skips dial failures (`bulkstats.go:36-38`) so `Bytes` < `n` silently; `RetransMeasured` = true if *any* socket sampled (`bulkstats.go:48-56`) while `RetransBytes` sums only sampled — partial basis.
   - *Byte-unit mismatch.* `×1448` (`tcpinfo_linux.go:15,43`) counts segments × MSS; eBPF `skb->len` counts payload TSO-inclusive (`bpf_probe.c:119`). Two tiers count different things; no code reconciles them (spec §2.3).
5. **Report seam is designed, not wired.** Ground-truth spec §4.2–§4.5 proposes `BulkStats.Source`/`Kernel`, `metrics.RetransRatio.Source`, `audit.LoadOutcome.RetransSource`/`Kernel`, `Result.KernelUp/Down`, `Report.KernelTelemetry{Up,Down}`, Markdown “Télémétrie noyau” block, doctor `ebpf` check (`ok`/`warn`/`fail` on `Check()`), and `pkg/audit/kernel_adapter.go` decoupling — none of which exists in the current `pkg/probe` tree for this seam (and `KernelTelemetryFrom`/adapter belong to the audit module, not here). An observer must publish `bpf.Summary`/`Flow`/`Globals` cleanly so the audit can consume it without importing the loader (spec §4.1 alias `probe.KernelTelemetry = bpf.Summary`).
6. **Bench identities and tunables are hardcoded for the bench, not the API.** `BulkSendNTo` defaults `""`→`model.Cubic` (`bulkflows.go:43-45`); `BulkSendNFlows` third arg ignored (`bulkflows.go:23`); `DialWithCC` `5s` dial timeout (`bulk_cc_linux.go:15`); `SmallClient` 2 s; `SAMPLE_NS` 50 ms (`bpf_probe.c:27`); `MAX_FLOWS` 4096 (`bpf_probe.c:26`). A releasable API needs these as explicit `Options`/`Config`, not globals (sibling §5.6 class of gap).
7. **Versioning/schema vacuum.** No `LinkAuditSchema` version string exists in `pkg/probe`; standalone module needs its own schema stamp so `telemetry.csv`/`link_audit.csv` freeze (`AGENTS.md` “Never touch” + invariant 3) is versioned, not fiat. Today `AppendLinkAudit` self-migrates headers (sibling §5.10) — the observer should keep schemas as files with versions.

---

## 6. Target tool: `cgo-ebpf-observer` as independent module

Proposed contract for a standalone observer (path illustrative; shape is the deliverable). Per (R13) it lives behind its own `go.mod`; pre-1.0 (`0.y.z`) it MAY break API. It contains **only** probe/bpf — no audit, no metrics, no remediate, no model. Stdlib + `cilium/ebpf` only.

```go
// ── Module: github.com/<org>/cgo-ebpf-observer ────────────────────
// Stdlib + github.com/cilium/ebpf. No metrics/audit/remediate/model imports.
// Kernel truth (what the kernel actually did):
type Options struct {
    CC         string        // "" = host default (named), never cubic-by-global
    Flows      int           // >0 = multi-flow; 0 = single
    SampleNS   time.Duration // 0 = 50ms default (bpf_probe.c:27)
    MaxFlows   int           // 0 = 4096 default (bpf_probe.c:26)
    Attach     string        // "" = autofentry→kprobe; or "fentry"|"kprobe"
}
func Ping(ctx context.Context, target string, count int, interval time.Duration, runner PingRunner) (PingResult, error)
// PingRunner is the exec seam (testable without real ping); production shells ping, test returns canned output.
type PingResult struct{ Attempts, Received int; Samples []PingSample } // Attempts in type: loss is computable.

type SmallClient struct{ Timeout time.Duration; MarkEF bool } // MarkEF is a NAME, not silent Linux-only.
func SmallObject(ctx context.Context, c *http.Client, url string) (time.Duration, error)

func BulkUpload(ctx context.Context, addr string, opt Options) (BulkStats, error)
func BulkDownload(ctx context.Context, addr string, opt Options) (BulkStats, error) // "D:<cc>" hello internal.
type BulkStats struct {
    PerFlow []uint64
    RetransBytes uint64; RetransMeasured bool; RetransProxy bool; RetransUnit string // "payload"|"tcpinfo_segments_x_mss"
    Source string // "ebpf"|"tcpinfo"|""
    Kernel *bpf.Summary
}

func TCPRetransBytes(c net.Conn) (uint64, error) // ErrNotTCP / ErrNotLinux are honest returns.

// Kernel tier (R3 lifecycle Open→Load→Attach→Tear-down; R8/R9/R10):
type Session interface {
    Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
    Snapshot() []bpf.Flow
    Globals() (bpf.Globals, error)
    Attach() string // "fentry"|"kprobe" — socket sample only
    Close() error
}
func Open() (*Session, error)          // load + attach; fresh maps per session
func Check() error                     // hash verify then load-only for doctor (no attach)
func ObjectSHA256() (string, error)    // provenance
func KernelSource() (KernelSourceInfo, error) // ladder with real capget + BTF parse, not stat/euid
// bpf types stay as today (types.go:4-47); errors stay typed (errors.go:4-11).

// CLI: `cgo-ebpf-observer --addr 10.30.0.2:5201 --cc cubic --flows 4 --duration 10s --format json`
// Emits Summary JSON + flow table to stdout, exit 2 = fallback (tcpinfo with reason), exit 1 = fatal.
```

Design notes: (a) observer has zero `metrics`/`audit`/`remediate`/`model` imports — `go list -m all` shows stdlib + `cilium/ebpf` only; (b) callers never import `probe/bpf` loader directly — they consume `bpf.Summary`/`Flow`/`Globals` via the alias or a `KernelTelemetry` interface; (c) every “absent” is a flag/error/`nil`, never silent zero (honesty invariants 1–2, umbrella spec §2); (d) CC/DSCP-marking, sinks, and tunables are explicit values, so the bench is configuration, not code; (e) CLI is optional sugar over the library — the library is the product.

---

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check. AGENTS.md stop-and-ask applies to the new-dependency step (adding `cilium/ebpf`) and to any live bench run — ask before acting.

1. **Slice 1 — Freeze + declare.** Snapshot §2 API into the observer README as `0.1.0` declared public API; create `go.mod` (R13); keep `LICENSE` (MIT, R8) + GPL object note; port tests unchanged (`probe_test.go`, `kernelsource_test.go`, `bpf/*_test.go`). Acceptance: `go test ./...` green on extracted tree; README lists every §2 identifier.
2. **Slice 2 — Probe de-globalise + honest sockets.** Delete `model.Cubic` default (`bulkflows.go:44`) for explicit `Options{CC}`; make DSCP marking a named `MarkEF` instead of silent Linux-only; make `PingRunner` the seam and carry `Attempts` (gap §5.4); label `RetransProxy`/unit on `BulkStats` so `×1448` never passes as payload-exact. Acceptance: `grep -rn "pkg/model" pkg/probe` empty; non-Linux returns named errors in tests; `-race` clean.
3. **Slice 3 — Kernel truth (object + loader).** Build `bpf_probe.o` + `.sha256` on the pinned stack (`BUILD.md:16-41`); pin `cilium/ebpf` in `go.mod` (stop-and-ask); implement `Open/Check/Lookup/Snapshot/Globals/Attach/Close` against R3 lifecycle + R10 `CollectionSpec`/`NewCollection`/`LoadAndAssign` + verifier log (R4); replace `kernelFacts` stat/euid with BTF-parse + `capget` (R7); wire `Session` → `bpf.Summary` → `BulkStats.Kernel`, read `flows`/`globals`, attribute drops per R5 tracepoints; pin verifier log + `ObjectSHA256` in tests. Acceptance: on a 6.8 VM with BTF + `CAP_BPF`, `KernelSource()==ebpf` and a load produces non-zero `retrans_payload` matching TCP_INFO order-of-magnitude; without caps it falls to `tcpinfo` with named `Reason`. Until this lands, `ebpf` is never claimed in the README.
4. **Slice 4 — Bench validation (dualns).** Run 60 s upload + download audits with eBPF on/off; require `source=ebpf`, coverage N/N, exact payload >0 when loss present, qdisc drops when queue overflows; measure overhead <2 % (spec §6.5); exercise external-downlink “émetteur distant” fallback. No live run without human approval (AGENTS.md “Stop and ask” + “live bench without explicit human approval”). Acceptance: report shows `kernel_telemetry` with object SHA + per-window counters; otherwise row stays BLOCKED.
5. **Slice 5 — 7.1+ upgrade + `.dist` correlation.** Once target kernel is ≥7.1, swap `kfree_skb` filter for native `qdisc_drop` tracepoint (spec §§2.1, 7.12) with a new object+hash; correlate observer drop corpus with `.dist` traces (30-sample floor) and BBRv3 3×3 (require kernel 7.1 + BBRv3 registered). Cross-kernel numeric merging with frozen 6.8 corpus is forbidden (BBRv3 spec §6). Acceptance: 7.1 run produces `qdisc_available=true` via native hook and a `.dist`-gradeable trace ≥30 samples, each in its own run set labelled `bench=dualns`/`kernel=7.1`.
6. **Slice 6 — Publish.** Tag `0.x.y` on the observer repo; audit module consumes `BulkStats.Source`/`Kernel` (spec §4.1 alias) without re-importing the loader; update VERIFY rows with named command + host + date. Until `VERIFIED`, stay BLOCKED — never promote by confidence (VERIFY §How a row moves).

---

## 8. Anti-toy criteria

A claim of “kernel ground truth” is toy unless all of these hold. Each lists the exact check that would refute it.

| # | Criterion | Check (must pass) | Refuter (toy if…) |
|---|---|---|---|
| A1 | Object provenance | `bpf_probe.o` committed + `bpf_probe.o.sha256` matches `ObjectSHA256()` + `make bpf-check` passes on build host + sections contain `raw_tp`/`fentry`/`kprobe` via `readelf` (`BUILD.md:35-41`) | Object empty/missing, hash mismatch ignored, or `vmlinux.h` claimed as artefact. |
| A2 | Load ≠ attach | `Check()` load-only returns typed error; doctor reports `warn` with named reason, never `ok` on load alone; attach separately verified by audit (spec §4.5) | Load success marketed as attach success; doctor `ok` without BTF/caps. |
| A3 | All-or-nothing coverage | Window = `ebpf` only if *all* dialed flows observed with sample/event; qdisc drop alone ≠ observed; missing ≠ zero (spec §4.2) | Partial flows summed as if complete; missing flow treated as 0. |
| A4 | Submitted vs delivered honesty | Report says “réémissions transmises”; no ACK pairing; no dwell/sojourn claim (spec §§2.3, 2.6) | “Delivered/ACKed” or dwell wording appears. |
| A5 | Unit honesty | `RetransBytes` tag names unit (`payload` for eBPF, `tcpinfo_segments×mss` for fallback) and never mixes them into the ratio (spec §§2.3, 4.2) | `×1448` labelled payload-exact; eBPF and TCP_INFO summed. |
| A6 | Fallback ladder completeness | Each rung tested (Windows / absent object / no BTF / no caps / attach-fail / map-full / external sender) with typed reason (VERIFY `docs/VERIFY.md:58`; spec §5) | Any rung returns `ebpf` or panics; reason empty. |
| A7 | Verifier discipline | Verifier log captured; NULL-checked map access (R4); no uninitialised-stack or misaligned access in the log | Verifier error suppressed or “load elsewhere”. |
| A8 | Overhead bound | 60 s audit eBPF on/off: goodput delta <2 %, all phases run, no audit error (spec §6.5) | Delta ≥2 % shipped without naming the sampling/drop suspect. |
| A9 | Bench coverage | Dualns upload + download both report `source=ebpf`, coverage `4/4` (or N/N), payload>0 under loss, qdisc drops under overflow; remote download falls back “émetteur distant” (spec §6.4) | Only loopback shown; remote sender claimed as local. |
| A10 | Non-regression | `telemetry.csv`/`link_audit.csv` headers + frozen corpus byte-identical (`audit-archive.py` → both OK lines; spec §§0.3, 4.6) | Schema or frozen file drifted. |
| A11 | Concurrency safety | Per-session collection (no global state), concurrent `Lookup`/`Snapshot` from N flow goroutines (spec §3.3) | One audit resets another’s counters. |
| A12 | Right hook for the kernel | 6.8 uses `kfree_skb`+`SKB_DROP_REASON_QDISC_DROP`; 7.1+ may switch to native `qdisc_drop` only with new object+hash (spec §§2.1, 7.12) | `qdisc_drop` claimed on 6.8, or hook swapped without re-hash. |

If any criterion fails, the module is a toy — ship it as contract (`tcpinfo` source, named reason) and keep the VERIFY row BLOCKED. That is the design’s own rule: “The feature is not called ‘verified’ until the opt-in Linux integration test runs on 6.8 with BTF” (spec §0.3), and “`RECORDED` is not promoted by confidence” (VERIFY §How a row moves).

---

## 9. Caveats

- Kernel version floor: object targets 6.8 (`BUILD.md:16`). Behaviour on newer kernels (TOCTOU on `srtt_us`/`snd_cwnd` fields, tracepoint arg layouts, drop-reason enum) is `unverified — primary source not found` in this pass until the Slice-5 7.1 run.
- BBRv3 provenance: `unverified — primary source not found` in this pass (see §0 table); the matrix design’s `google/bbr` claims are quoted only as sibling-spec context.
- Ping / TCPInfo field invariance / BBRv3 / JFI / G.107 / RouterOS dialect remain `unverified — primary source not found` as labelled in §0.
- No live run was performed. Every `ebpf` statement in the repo today is *design* (spec §§1–9) or *contract* (VERIFIED Windows half). That is the point: the seam is the place where honesty and measurement must finally meet on a 6.8 host.

---

## 10. References

- Kernel BPF docs: [BTF](https://docs.kernel.org/bpf/btf.html) (R2), [libbpf overview](https://docs.kernel.org/bpf/libbpf/libbpf_overview.html) (R3), [verifier](https://docs.kernel.org/bpf/verifier.html) (R4), [tracepoints](https://docs.kernel.org/trace/tracepoints.html) (R5), [LLVM relocations / CO-RE](https://docs.kernel.org/bpf/llvm_reloc.html) (R6).
- Capabilities: [`capabilities(7)`](https://man7.org/linux/man-pages/man7/capabilities.7.html) (R7) — `CAP_BPF`/`CAP_PERFMON` since 5.8, `CAP_SYS_ADMIN` overloaded.
- Go loader: [`cilium/ebpf`](https://github.com/cilium/ebpf) (R8, MIT), [Getting Started](https://ebpf-go.dev/guides/getting-started/) (R9), [Loading Objects](https://ebpf-go.dev/concepts/loader/) (R10); R9/R10 are the primary Go-loader sources for §4.2/§6.
- Sockets: [`tcp(7)`](https://man7.org/linux/man-pages/man7/tcp.7.html) (R11) — `TCP_CONGESTION` + `TCP_INFO`.
- AQM framing: [RFC 8290](https://www.rfc-editor.org/rfc/rfc8290.txt) (R1, Experimental) — FQ-CoDel/CoDel only; no CAKE RFC.
- Go modules: [Go Modules Reference](https://go.dev/doc/modules/gomod-ref) (R13).
- Repo ground truth: `pkg/probe/probe.go:1-139`, `pkg/probe/bulk.go:1-77`, `pkg/probe/bulksource.go:1-103`, `pkg/probe/bulkflows.go:1-64`, `pkg/probe/bulkstats.go:1-73`, `pkg/probe/bulk_cc_linux.go:1-28`, `pkg/probe/ccsock_linux.go:1-34`, `pkg/probe/tcpinfo_linux.go:1-44`, `pkg/probe/kernelsource.go:1-38`, `pkg/probe/bpf/bpf_probe.c:1-253`, `pkg/probe/bpf/BUILD.md:1-48`, `pkg/probe/bpf/types.go:1-47`, `pkg/probe/bpf/status.go:1-35`, `pkg/probe/bpf/errors.go:1-11`, `pkg/probe/bpf/assets_embedded.go:1-15`, `pkg/probe/bpf/assets_absent.go:1-9`, `Makefile:44-59`, `go.mod:1-60`, `docs/VERIFY.md` (rows 28–34, 52–58), `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`, `docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md`.
