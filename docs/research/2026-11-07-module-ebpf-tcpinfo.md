# eBPF sidecar + BBR/TCP_INFO path — standalone tool

*Date: 2026-11-07 · Scope: `pkg/probe/*.go`, `pkg/probe/bpf/*`, `docs/VERIFY.md` BLOCKED rows — deep research only, no code changed, no commit.*
*Working tree: `C:\Users\ASUS\.local\share\opencode\worktree\1c66ad257b56747880c45bc5d4f7e50f8688e1eb\operator-cockpit` · Host for repo facts: DESKTOP-B6K9UCM (Windows 10, MINGW64).*
*Primary sources fetched for this note are listed in §1; every kernel/BBR/TCP/cilium claim below cites one of them as `[R#]`. Any claim without a fetched primary source is marked `unverified — primary source not found`.*

## 1 Primary-source index (fetched)

| ID | Source | URL fetched | What it proves for this module |
|----|--------|-------------|--------------------------------|
| R1 | Linux kernel — libbpf Overview | https://docs.kernel.org/bpf/libbpf/libbpf_overview.html | BPF app lifecycle Open → Load → Attach → Tear-down; skeleton embeds bytecode (“no extra files to deploy”); CO-RE portability via `/sys/kernel/btf/vmlinux` + `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h`; `bpf_core_read` relocation. |
| R2 | Linux kernel — BTF | https://docs.kernel.org/bpf/btf.html | BTF type/ELF format; `.BTF` / `.BTF.ext` sections; BTF is the contract CO-RE relocations match against at load time. |
| R3 | Linux kernel — BPF maps | https://docs.kernel.org/bpf/maps.html | Maps share data kernel↔userspace via `bpf()` syscall (`BPF_MAP_CREATE` / `BPF_MAP_LOOKUP_ELEM` / `BPF_MAP_UPDATE_ELEM` / `BPF_MAP_DELETE_ELEM`); hash / array / percpu-array types exist. |
| R4 | `bpf-helpers(7)` man page | https://man7.org/linux/man-pages/man7/bpf-helpers.7.html | `bpf_map_lookup_elem`, `bpf_map_update_elem` (`BPF_NOEXIST`), `bpf_ktime_get_ns`, `bpf_probe_read` / `bpf_probe_read_kernel` semantics. |
| R5 | `tcp(7)` man page | https://man7.org/linux/man-pages/man7/tcp.7.html | `TCP_INFO` (since Linux 2.4) returns `struct tcp_info` from `/usr/include/linux/tcp.h`; `TCP_CONGESTION` sets CC per-socket via string option; `/proc/sys/net/ipv4/tcp_available_congestion_control` and `tcp_allowed_congestion_control` bound choices. |
| R6 | Linux UAPI `tcp.h` (torvalds/linux master, `include/uapi/linux/tcp.h`) | https://raw.githubusercontent.com/torvalds/linux/master/include/uapi/linux/tcp.h | `TCP_INFO=11`, `TCP_CONGESTION=13 (0x0d)`; `struct tcp_info` fields incl. `tcpi_total_retrans` (`__u32`), `tcpi_snd_cwnd`, `tcpi_snd_ssthresh`, `tcpi_rtt`, `tcpi_bytes_retrans` (`__u64`), `tcpi_delivery_rate`, `tcpi_min_rtt`. |
| R7 | `capabilities(7)` man page | https://man7.org/linux/man-pages/man7/capabilities.7.html | `CAP_BPF` (since Linux 5.8) = “employ privileged BPF operations”, split from `CAP_SYS_ADMIN`; `CAP_PERFMON` (since 5.8); `CAP_SYS_ADMIN` still implies the same BPF ops but is broader/overloaded; bounding set / `capget` semantics mean euid alone does not prove caps. |
| R8 | `cilium/ebpf` (ebpf-go) | https://github.com/cilium/ebpf → README: “pure Go library to read, modify and load eBPF programs and attach them to various hooks in the Linux kernel”; packages `link`, `btf`, `rlimit`, `features`, `cmd/bpf2go`; MIT; Linux amd64/arm64, CI on kernel.org LTS. | The Go loader the repo currently lacks; its `link`/`btf`/`rlimit` split is the baseline for §6. |
| R9 | IETF `draft-ietf-ccwg-bbr-07` (BBR Congestion Control, BBRv3) | https://datatracker.ietf.org/doc/html/draft-ietf-ccwg-bbr-07 | BBRv3 is model-based: delivery-rate, RTT, loss → `BBR.max_bw`, `BBR.min_rtt`, BDP, pacing rate / `C.cwnd` / `C.send_quantum`; state machine Startup → Drain → ProbeBW / ProbeRTT; `C.SMSS` MUST/SHOULD be payload-only for TCP (`Eff.snd.MSS`, RFC 9293 §3.7.1). |

Repo-internal facts cite `file:line` instead of `[R#]`. Where a repo comment asserts a kernel fact that has no fetched primary source, the claim is kept verbatim as a repo comment and flagged `unverified — primary source not found`.

## 2 Responsibility

`pkg/probe` (+ `pkg/probe/bpf`) is **load + CC + kernel telemetry**, nothing else. It owns:

- **Traffic it sends:** ICMP ping (`probe.Ping`), small HTTP object (`probe.SmallObject` + DSCP-marked dial), bulk flood (`BulkSend`/`BulkReceive` and N-flow variants) — `pkg/probe/probe.go:28`, `pkg/probe/probe.go:122`, `pkg/probe/bulk.go:20`, `pkg/probe/bulk.go:43`.
- **CC enforcement per socket:** dial-side `DialWithCC` and accept-side `SetConnCC` via `TCP_CONGESTION` — `pkg/probe/bulk_cc_linux.go:14`, `pkg/probe/ccsock_linux.go:15`. On non-Linux both are no-ops keeping the socket working — `pkg/probe/bulk_cc_other.go:11`, `pkg/probe/ccsock_other.go:10`.
- **Retransmit ground truth, two tiers:** (a) TCP_INFO fallback `TCPRetransBytes` (`pkg/probe/tcpinfo_linux.go:20`), (b) eBPF sidecar — CO-RE object `bpf_probe.o` (source `pkg/probe/bpf/bpf_probe.c:1`) + Go contract types/status/hash (`pkg/probe/bpf/types.go`, `status.go`, `assets_*.go`). Fallback ladder `probe.KernelSource` (`pkg/probe/kernelsource.go:34`) chooses `ebpf` vs `tcpinfo`.
- **Aggregation for callers:** `BulkStats` (`pkg/probe/bulkstats.go:11`) carries per-flow bytes + TCP_INFO retrans bytes + `RetransMeasured` flag; `bulkNStats` (`pkg/probe/bulkstats.go:29`) samples TCP_INFO **before** real close via `keepConn` (`pkg/probe/bulkstats.go:23`).

It does **not** own: bench topology / netns orchestration (kit/), campaign matrix (`pkg/campagne`), metrics/economics (`pkg/metrics`), audit/report rendering (`pkg/audit`, `pkg/api`), qdisc shaping, or the marks/TUN policy beyond the small-object DSCP bit. An independent tool must therefore extract only `pkg/probe` (+ `pkg/probe/bpf`) and re-expose typed seams so callers never import `probe/bpf` loader internals directly (same layering rule as `pkg/audit/kernel_adapter.go:3` and `pkg/api/kernel_bridge.go:8`).

**BLOCKED context (`docs/VERIFY.md`):** the live eBPF row, the Linux 6.8 live row, BBRv3 engine/matrix, IFB, `.dist`, PC2 bench access, and CHR rows are all `BLOCKED` (`docs/VERIFY.md:28`, `:29`, `:33`, `:52`, `:56`, etc.). The Go-half row is `VERIFIED` but explicitly “No object exists and nothing was attached: the slice-4 live row stays BLOCKED” (`docs/VERIFY.md:58`). The object row is `RECORDED` — source + `make bpf` target landed, “`make` is absent on this Windows host, so the target was not run. No object, no hash, no loader yet” (`docs/VERIFY.md:30`). Any standalone tool inherits these blockers as its own acceptance gates; it cannot be called verified from Windows alone.

## 3 Exported API (file:line, semantics)

All line numbers are on this working tree. Signatures are exact; comments are paraphrased to one line. Build tags noted where they change the API.

| Export | Location | Signature / shape | Contract |
|--------|----------|-------------------|----------|
| `PingSample` | `pkg/probe/probe.go:22` | `type PingSample struct{ RTTms float64 }` | One ICMP round-trip in ms. |
| `Ping` | `pkg/probe/probe.go:28` | `func Ping(ctx context.Context, target string, count int, intervalMs int) ([]PingSample, error)` | Spawns system `ping`; Unix `-c/-i`, Windows `-n`; timeout `count*interval+5s` (1s/pkt on Windows). Parses `time=`/`temps=` tolerant EN+FR. |
| `SmallClient` | `pkg/probe/probe.go:93` | `func SmallClient() *http.Client` | Shared client, 2s timeout, DSCP-EF transport. |
| `SmallClientTimeout` | `pkg/probe/probe.go:104` | `func SmallClientTimeout(timeout time.Duration) *http.Client` | Fresh client with caller timeout (bench-only, VSAT). |
| `SmallObject` | `pkg/probe/probe.go:122` | `func SmallObject(ctx context.Context, client *http.Client, url string) (float64, error)` | One HTTP GET, returns completion ms. |
| `BulkSendTo` / `BulkSend` / `BulkReceive` | `pkg/probe/bulk.go:11`, `:20`, `:43` | `BulkSendTo(ctx, addr, cc) (uint64,error)`; `BulkSend(ctx, conn) (uint64,error)`; `BulkReceive(ctx, conn, onBytes func(uint64)) (uint64,error)` | Flood zeros until ctx done; 64 KiB buf, 2s write deadline; receive closes on EOF/deadline. |
| `DialWithCC` (linux) | `pkg/probe/bulk_cc_linux.go:14` | `func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error)` | `net.Dialer.Control` → `SetsockoptString(IPPROTO_TCP, 0x0d, cc)`. `0x0d==13==TCP_CONGESTION` [R6]. |
| `DialWithCC` (!linux) | `pkg/probe/bulk_cc_other.go:11` | same signature, `cc` ignored | Plain dial; doc: inter-CC download only interpretable on Linux bench. |
| `SetConnCC` (linux) | `pkg/probe/ccsock_linux.go:15` | `func SetConnCC(conn net.Conn, cc string) error` | Server-side accept socket; `SyscallConn` + `SetsockoptString`. No-op (`nil`) on non-TCP or empty `cc`. |
| `SetConnCC` (!linux) | `pkg/probe/ccsock_other.go:10` | `func SetConnCC(_ net.Conn, _ string) error { return nil }` | Documented no-op. |
| `FlowResult` / `BulkSendNFlows` / `BulkSendNTo` | `pkg/probe/bulkflows.go:14`, `:23`, `:42` | `type FlowResult struct{ Bytes uint64 }`; `BulkSendNFlows(ctx, conns []net.Conn, _ int) ([]FlowResult,error)`; `BulkSendNTo(ctx, addr, cc string, n int) ([]uint64,error)` | Parallel N `BulkSend`; `cc==""`→`cubic`; `n<=1` single. Ignores per-flow errors (live flows count). |
| `BulkSource` / `BulkRecvSource` / `BulkDownloadTo` / `BulkDownloadNTo` / `ParseDownloadHello` | `pkg/probe/bulksource.go:14`, `:39`, `:55`, `:68`, `:94` | Mirror of upload; hello `"D"` / `"D:<cc>\n"` | Download = source floods, client receives; CC travels in hello. |
| `BulkStats` | `pkg/probe/bulkstats.go:11` | `type BulkStats struct{ Bytes []uint64; RetransBytes uint64; RetransMeasured bool }` | Per-flow bytes + TCP_INFO retrans total + measured flag. **No `Source`/`Kernel` field yet** (see §5.3). |
| `BulkSendNStats` / `BulkDownloadNStats` | `pkg/probe/bulkstats.go:60`, `:68` | `func BulkSendNStats(ctx, addr, cc string, n int) (BulkStats,error)`; `...DownloadNStats` | `dial=DialWithCC` + `bulkNStats` (sample-before-close). |
| `KernelSourceInfo` / `KernelSource` | `pkg/probe/kernelsource.go:12`, `:34` | `type KernelSourceInfo struct{ Source string; Reason string }`; `func KernelSource() KernelSourceInfo` | Delegates to `bpf.Status(goos,btf,caps,objectErr)`. Source ∈ `ebpf`/`tcpinfo`. |
| `TCPRetransBytes` (linux) | `pkg/probe/tcpinfo_linux.go:20` | `func TCPRetransBytes(c net.Conn) (uint64,error)` | `*net.TCPConn` → `File()` → raw `SYS_GETSOCKOPT(IPPROTO_TCP, TCP_INFO)` → `info.Total_retrans * 1448` (`pkg/probe/tcpinfo_linux.go:15`, `retransMSS`). Non-TCP → error. |
| `TCPRetransBytes` / `ErrTCPInfoUnavailable` (!linux) | `pkg/probe/tcpinfo_other.go:14`, `:12` | `func TCPRetransBytes(net.Conn) (uint64,error)` returning typed error | `TCP_INFO` is Linux-only; caller must declare non-measured, never invent. |
| `bpf.Flow` | `pkg/probe/bpf/types.go:4` | `Local, Peer string; RetransPayload, RetransEvents uint64; DropPackets, DropBytes uint64; SRTTus, SndCwnd, SndSsthresh, State, Samples uint32` | One sender-side flow as read from kernel maps. |
| `bpf.Summary` | `pkg/probe/bpf/types.go:20` | `Available bool; Source string; Attach string; FallbackReason string; ObjectSHA256 string; FlowsSeen, FlowsTotal int; RetransPayload, RetransEvents uint64; QdiscDrops, QdiscDropBytes uint64; QdiscAvailable bool; MapFull uint64; Flows []Flow` | One telemetry window; `Source` ∈ `ebpf`/`tcpinfo`/`""`. |
| `bpf.Globals` | `pkg/probe/bpf/types.go:40` | `RetransPayload, RetransEvents, DropPackets, DropBytes, MapFull uint64; QdiscAvailable bool` | Host-global fallback for IPv6/non-IPv4 + full-map (see §4.2). |
| `bpf.Status` | `pkg/probe/bpf/status.go:11` | `func Status(goos string, btf, caps bool, objectErr error) (source, reason string)` | Ladder: non-Linux→`tcpinfo`; objectErr→`tcpinfo`; !btf→`tcpinfo`; !caps→`tcpinfo`; else `ebpf`. Pure, injectable. |
| `bpf.ObjectSHA256` / `bpf.ObjectBytes` | `pkg/probe/bpf/status.go:28`; `pkg/probe/bpf/assets_embedded.go:10`; `pkg/probe/bpf/assets_absent.go:7` | `ObjectSHA256() (string,error)` = hex sha256 of embedded bytes; `ObjectBytes() ([]byte,error)` | `//go:build ebpf_object` embeds `bpf_probe.o`; else `BuildError{"objet non embarqué ..."}`. Empty embed → `BuildError{"objet vide"}`. |
| `bpf.UnavailableError` / `bpf.BuildError` | `pkg/probe/bpf/errors.go:4`, `:9` | `type UnavailableError struct{ Reason string }` / `type BuildError struct{ Reason string }` | Host limitation vs object defect. |
| BPF programs (C) | `pkg/probe/bpf/bpf_probe.c:112`, `:176`, `:182`, `:188` | `raw_tp/tcp_retransmit_skb` `on_retransmit`; `fentry/tcp_rcv_established` `on_ack_fentry`; `kprobe/tcp_rcv_established` `on_ack_kprobe`; `raw_tp/kfree_skb` `on_kfree_skb` | Hooks + maps `flows` (HASH 4096) + `globals` (PERCPU_ARRAY 1) — `pkg/probe/bpf/bpf_probe.c:60`. |
| `make bpf` / `make bpf-check` | `Makefile:49`–`:59` | `vmlinux.h` via `bpftool btf dump file /sys/kernel/btf/vmlinux format c`; `clang-18 -O2 -g -target bpf` → `bpf_probe.o` + `.sha256`; check = sha256 equality | Maintainer-only; output documented in `pkg/probe/bpf/BUILD.md:21`. |

**Not present on disk** (designed in the ground-truth spec, absent in code): `bpf.Start (*Session,error)`, `(*Session).Lookup/Snapshot/Globals/Attach/Close`, `bpf.Check() error`, `probe.KernelTelemetry = bpf.Summary` alias, `BulkStats.Source/Kernel`, `kernelSession` interface — see spec excerpts in `docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md:393`–`:433`, `:482`–`:537`. A standalone tool must implement them (or explicitly refuse them) — they cannot be imported from this tree.

## 4 Dependencies

### 4.1 Go module reality

- Module `github.com/Realms4239/cgo`, `go 1.25` (`go.mod:1`–`:3`). Requires `fyne.io/fyne/v2 v2.8.1`, `github.com/charmbracelet/bubbletea v1.3.10`, `lipgloss v1.1.0`, `golang.org/x/sys v0.36.0` (`go.mod:5`–`:10`). **No eBPF loader dependency** — `github.com/cilium/ebpf` is absent. That is the whole loader gap: the Go side has `ObjectBytes`/`Status`/`ObjectSHA256` only and cannot Open/Load/Attach per [R1][R8].
- `pkg/probe` imports: stdlib (`context`, `net`, `net/http`, `os/exec`, `runtime`, `sync`, `syscall`, `unsafe`) + `pkg/probe/bpf` (`pkg/probe/kernelsource.go:7`) + `pkg/model` (`pkg/probe/bulkflows.go:8`, for `model.Cubic` default). No `golang.org/x/sys/unix` — `tcpinfo_linux.go:30`–`:32` deliberately uses raw `SYS_GETSOCKOPT` because stdlib `syscall` has no TCP_INFO helper and `x/sys/unix` was forbidden by comment.
- `pkg/probe/bpf` imports: stdlib only (`crypto/sha256`, `encoding/hex`, `embed`) — `pkg/probe/bpf/status.go:3`, `assets_embedded.go:5`. Zero bpf(2) syscalls in Go.
- Build axes: `linux` vs `!linux` (CC/DSCP/TCP_INFO) and `ebpf_object` for object-embedded vs absent. No `linux && amd64` loader axis on disk; the spec's loader axis is planned, not built.

**Implication for a standalone repo:** add exactly one new dependency — `github.com/cilium/ebpf` (pinned; ground-truth plan says `v0.22.0`) — plus `golang.org/x/sys` bump to its minimum if required. Everything else stays stdlib. That needs stop-and-ask per `AGENTS.md` (new dependency) and a `go.mod` with its own module path; do not carry fyne/bubbletea/lipgloss into the observer.

### 4.2 External tooling (build host only)

`pkg/probe/bpf/BUILD.md:16` pins Ubuntu 24.04, kernel 6.8.0 (BTF `/sys/kernel/btf/vmlinux`), clang 18, `libbpf-dev`, `bpftool`. `Makefile:46`–`:56` pins `BPF_CLANG=clang-18`, `-O2 -g -target bpf -D__TARGET_ARCH_x86 -Wall -Werror`. Runtime host needs **no clang/gcc/llvm/linux-headers** — object is prebuilt and embedded [R1: CO-RE “compiled once and run across different kernel versions”]. `vmlinux.h` is gitignored (`BUILD.md:30`–`:31`); only `bpf_probe.o` (+ `.sha256`) is the deliverable.

### 4.3 Lineage for audit/report (out of scope, noted as consumers)

`pkg/audit/kernel_adapter.go:6` defines neutral `KernelSample`; `pkg/audit/kernel_adapter.go:20` maps it to `KernelTelemetry`, returning `nil` when `Source==""` (absent stays absent). `pkg/api/kernel_bridge.go:12` maps `bpf.Summary` → `KernelSample`. The observer must keep this neutrality: return `nil`/typed errors/fallback reasons, never zero-fill missing flows.

## 5 Primary-source requirements

For each requirement: what the primary source demands, what the repo does, and whether it holds.

### 5.1 CO-RE object + BTF (kernel.org) [R1][R2]

- **Demand:** CO-RE programs are compiled once against `vmlinux.h` generated from the running kernel’s BTF (`/sys/kernel/btf/vmlinux`) and relocated at load time by the loader against the runtime BTF [R1]. BTF metadata is the verbatim contract for relocations [R2].
- **Repo:** `bpf_probe.c` includes `vmlinux.h`, `bpf/bpf_helpers.h`, `bpf/bpf_tracing.h`, `bpf/bpf_core_read.h`, `bpf/bpf_endian.h` (`pkg/probe/bpf/bpf_probe.c:13`–`:17`); `Makefile:51`–`:52` generates `vmlinux.h` exactly via `bpftool btf dump ... format c` [R1]. Build doc pins clang 18 / libbpf-dev / bpftool [R1 build deps]. **Missing:** the object + `.sha256` do not exist (`BUILD.md:8`: “L’objet n’est pas encore construit”); verifier/CO-RE load has never run (`VERIFY.md:30`). Until then [R1] lifecycle cannot be entered — the tool is a source tree, not a runnable observer.

### 5.2 Maps, helpers, lifecycle (kernel.org + cilium/ebpf) [R1][R3][R4][R8]

- **Demand:** BPF apps go Open → Load → Attach → Tear-down; Skeleton embeds bytecode; maps are created by `BPF_MAP_CREATE` with type/key/value/max_entries [R1][R3]. Helpers `bpf_map_lookup_elem` / `bpf_map_update_elem` / `bpf_ktime_get_ns` have the exact signatures in [R4]. Go-side equivalents (`LoadCollectionSpec` → `NewCollection` / `LoadAndAssign` → `Close`, `link.Attach*`, `rlimit.RemoveMemlock`) are provided by `cilium/ebpf` [R8].
- **Repo C side (all used helpers exist in [R4]):** `flows` = `BPF_MAP_TYPE_HASH`, 4096 entries (`bpf_probe.c:60`); `globals` = `BPF_MAP_TYPE_PERCPU_ARRAY`, 1 entry (`bpf_probe.c:67`). Uses `bpf_map_lookup_elem`/`bpf_map_update_elem(BPF_NOEXIST)` (`bpf_probe.c:83`–`:89`), `bpf_ktime_get_ns` (`:160`), `bpf_probe_read_kernel` (`:220`, `:234`), `__sync_fetch_and_add` for counters. Per-CPU globals require userspace sum-across-CPUs [R3]; the `Globals` struct (`types.go:40`) anticipates it but no reader exists.
- **Repo Go side:** none. No `cilium/ebpf` import, no `LoadCollectionSpec`, no `link.AttachRawTracepoint`, no `rlimit.RemoveMemlock`. The ground-truth design specifies exactly that stack (`docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md:321`–`:336`, `:383`–`:380`) but the code was never written. A standalone observer must write it from [R1][R8], not stub it.

### 5.3 Hooks: retransmit, socket sample, qdisc drops [R1][R4][R9 — partially fetched]

- **Retransmit:** `raw_tp/tcp_retransmit_skb` reads `(sk, skb)` off `bpf_raw_tracepoint_args`, uses `skb->len` as payload (`bpf_probe.c:111`–`:146`). Repo comment states the tracepoint fires post-fragmentation/collapse on the success path and `skb->len` is payload because the tracepoint sees the queued skb, not the header-carrying clone (`bpf_probe.c:6`–`:9`). The configured tracepoint arg layout `(sk, skb)` for `tcp_retransmit_skb` and the `TP_PROTO` claim are **not corroborated by any fetched primary source** — the fetched [R1]–[R9] do not list the TCP tracepoint prototype. Mark: `unverified — primary source not found` for that arg-order detail (the code itself is the only evidence; see §7.3).
- **Socket sample:** `fentry/tcp_rcv_established` + `kprobe/tcp_rcv_established` fallback sample socket state at ≤50 ms (`SAMPLE_NS=50_000_000`, `bpf_probe.c:27`), `srtt_us>>3`, `snd_cwnd`, `snd_ssthresh`, `skc_state` (`bpf_probe.c:149`–`:185`). The 50 ms throttle is a design choice, not a kernel requirement; the `srtt_us` left-shift-by-3 convention is a kernel-internal detail **not in any fetched source** — `unverified — primary source not found` as a kernel fact (it is correct only as a repo-convention claim).
- **Qdisc drops:** `raw_tp/kfree_skb` filtered on `SKB_DROP_REASON_QDISC_DROP` via `bpf_core_enum_value` (`bpf_probe.c:187`–`:251`), then IPv4/TCP header parse with bounds checks against `skb_headlen`, fragment check, else global-only. The `kfree_skb` raw args `(skb, location, reason)` and the 6.8-vs-7.1 `qdisc_drop` tracepoint history are **not in any fetched primary source** (they come from the ground-truth spec §2.1, which itself cites kernel commits not fetched here). Mark the kernel-history claim `unverified — primary source not found`; the code pattern (reason-filtered free) is consistent with [R4] map/helper usage but the specific tracepoint choice is unverified here.

### 5.4 BBR / CC (IETF draft + man tcp) [R5][R9]

- **Demand:** BBRv3 (`draft-ietf-ccwg-bbr-07`) models delivery rate / RTT / loss into `max_bw`, `min_rtt`, BDP and paces via `C.pacing_rate` + `C.cwnd` + `C.send_quantum`; TCP `C.SMSS` SHOULD be payload-only (RFC 9293 §3.7.1) [R9 §Abstract, §2.2]. `TCP_CONGESTION` (13) sets CC per-socket by string; unprivileged restricted to `tcp_allowed_congestion_control`, privileged to any in `tcp_available_congestion_control` [R5].
- **Repo:** observer does **not** implement BBR — it only selects CC by name (`DialWithCC`/`SetConnCC` via `0x0d` [R6] = `TCP_CONGESTION` [R5]) and measures outcomes. That is honest and sufficient: the draft requires sender-side model changes [R9 §3.6], which live in the kernel (+ `kit/install-bbr3.sh` building `google/bbr` branch `bbr-v3-2026-09-16-01` as kernel 7.1 with `CONFIG_TCP_CONG_BBR3=y`). Repo `VERIFY.md:33`/`VERIFY.md:52` track the 7.1/BBRv3 blockers separately. The observer must **never** claim BBRv3 exactness; it can only claim the requested CC name reached the socket (a `setsockopt` success) — anything about `min_rtt`/`max_bw` model state inside the sender is `unverified — primary source not found` from this module’s vantage point.

### 5.5 TCP_INFO (man tcp + UAPI tcp.h) [R5][R6]

- **Demand:** `TCP_INFO` (since Linux 2.4) returns `struct tcp_info` from `/usr/include/linux/tcp.h`; “should not be used in code intended to be portable” [R5]. `TCP_INFO=11`, `TCP_CONGESTION=13` [R6]. `struct tcp_info` includes `tcpi_total_retrans`, `tcpi_bytes_retrans`, `tcpi_snd_cwnd`, etc. [R6].
- **Repo:** `TCPRetransBytes` does `SYS_GETSOCKOPT(IPPROTO_TCP, TCP_INFO)` into Go `syscall.TCPInfo` and returns `Total_retrans * 1448` (`pkg/probe/tcpinfo_linux.go:20`). That matches [R5]+[R6] on *what* is read. The multiplier is the documented bench proxy: “TCP_INFO rend des SEGMENTS réémis ; ×1448 les convertit en octets comme le fait le mémoire” (`tcpinfo_linux.go:12`–`:15`). `*1448` is exact only if MSS is 1448 for the path; the choice 1448 (vs the usual 1460 for MTU 1500 IPv4 − 40) is bench-specific and **not derivable from any fetched primary source** — mark the 1448 number itself as repo-convention, not kernel fact. Behaviour on short struct: comment says kernel fills `min(struct, kernel)` and extra fields stay zero (`tcpinfo_linux.go:32`–`:34`) — kernel-prefix behaviour is **not in [R5][R6]**; `unverified — primary source not found`.
- Additional modern TCP_INFO fields the repo ignores (`tcpi_bytes_retrans` `__u64`, `tcpi_delivery_rate`, `tcpi_min_rtt` [R6]) should be surfaced by the standalone tool as optional exact-byte alternatives; today only the segment-count × MSS path is used.

### 5.6 Privileges (capabilities + BTF + architecture) [R1][R7]

- **Demand:** privileged BPF needs `CAP_BPF` (5.8+) or `CAP_SYS_ADMIN`; perf monitoring needs `CAP_PERFMON` (5.8+) [R7]. BTF presence at `/sys/kernel/btf/vmlinux` is required for CO-RE [R1]. Architecture/`linux/amd64` is a loader-build choice, not a kernel requirement — the kernel does not restrict eBPF to amd64; the repo’s `linux && amd64` axis is a toolchain choice (unverified as kernel requirement; keep as build-axis, not kernel claim).
- **Repo:** `kernelFacts` (`pkg/probe/kernelsource.go:19`) uses `os.Stat("/sys/kernel/btf/vmlinux")` + `os.Geteuid()==0` as “BTF + caps”. Per [R7], euid 0 ≠ effective `CAP_BPF` when the bounding set was narrowed, and per [R1] file *presence* ≠ parseable BTF. The file itself admits “root approxime CAP_BPF/CAP_PERFMON” (`kernelsource.go:29`). A standalone tool must replace the heuristic with BTF-parse + `capget` (via `cilium/ebpf/features` [R8] or `capget(2)`) and keep the ladder’s reason strings honest (`status.go:11`).

## 6 Target tool

Single independent repo **cgo-bpf-observer** (name candidates `cgo-bpf-observer` / `cgo-tcpinfo` collapse into one module with two entry points — ponytail rule 1: one repo, thin CLI; two repos would duplicate the module boundary). The library is the product; the CLI is sugar. No metrics/audit/remediate/model imports — `go list -m all` must show stdlib + `cilium/ebpf` only.

### 6.1 Repository layout

```
cgo-bpf-observer/
  go.mod                          # module github.com/Realms4239/cgo-bpf-observer; go 1.25; require cilium/ebpf <pinned>, x/sys
  README.md                       # what it measures, what it refuses to claim, BLOCKED preconditions
  Makefile                        # bpf / bpf-check / test / verify   (same as root Makefile:44-59 shape)
  cmd/cgo-bpf-observer/main.go   # CLI: attach + dump Summary as JSON; `tcpinfo` subcommand
  internal/probe/                 # moved pkg/probe/*.go (Ping/SmallObject/Bulk*/DialWithCC/SetConnCC/TCPRetransBytes/KernelSource)
  internal/probe/bpf/             # moved pkg/probe/bpf/{types.go,status.go,errors.go,assets_*.go,bpf_probe.c,BUILD.md}
  internal/probe/bpf/session_linux.go   # loader: cilium/ebpf Open/Load/Attach/Tear-down [R1][R8]
  internal/probe/bpf/session_stub.go    # //go:build !linux || !amd64 identical API, typed UnavailableError
```

`internal/` (not `pkg/`) so external callers consume **one** public package (`probe`) + returned JSON/Summary values. The `probe/bpf` loader types stay internal; `probe.KernelTelemetry = bpf.Summary` alias lives at the seam (as the ground-truth spec intends).

### 6.2 Public surface (library contract)

Exact names and JSON tags as in §3; new additions marked `+`.

```go
package probe  // stdlib + cilium/ebpf only

// — TCP_INFO tier [R5][R6] —
func TCPRetransBytes(c net.Conn) (uint64, error)
var ErrTCPInfoUnavailable error
// + TCPInfoSnapshot returns the full struct tcp_info [R6] (optional; default bytes-retrans + del. rate)

// — CC selection [R5] —
func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error)
func SetConnCC(conn net.Conn, cc string) error

// — Kernel tier [R1][R8] —
func KernelSource() KernelSourceInfo
type KernelSourceInfo struct{ Source, Reason string }  // Source ∈ ebpf|tcpinfo; Reason names missing fact

// — Session (new; designed in ground-truth spec §4.1) —
type Session interface {
    Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
    Snapshot() []bpf.Flow
    Globals() (bpf.Globals, error)
    Attach() string  // "fentry"|"kprobe"
    Close() error
}
func Open() (*Session, error)   // fresh maps per session [R1]; load+attach, all-or-nothing per load
func Check() error              // SHA-verify then load-only (no attach); BuildError|UnavailableError

// — Bulk with kernel truth (new contract) —
type BulkStats struct {
    Bytes []uint64 `json:"bytes"`
    RetransBytes uint64 `json:"retrans_bytes"`
    RetransMeasured bool `json:"retrans_measured"`
    Source string `json:"source,omitempty"`          // + ebpf|tcpinfo|""
    Kernel *bpf.Summary `json:"kernel,omitempty"`    // + coverage + per-flow
}
func BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)
func BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)
```

CLI (optional sugar):

```
cgo-bpf-observer capture --flows 4 --duration 10s --cc bbr --out summary.json
cgo-bpf-observer tcpinfo --conn tcp://host:port   # single-shot TCPRetransBytes + full tcp_info [R5][R6]
cgo-bpf-observer doctor                            # Check() → ok|warn|fail, names missing fact
```

### 6.3 Non-goals (explicit)

- No BBR/BBRv3 implementation — CC is a socket option name [R5]; the model lives in the kernel [R9].
- No qdisc/tc control, no netns orchestration, no economics/audit/report — those callers stay in CGO.
- No IPv6 per-flow keys, no queue-dwell/ECN/ACK pairing — inherited from the object’s 6.8 scope (see §5.3).

## 7 Gap analysis

### 7.1 Object-less and loader-less — `ebpf` tier cannot run (critical)

- `BUILD.md:8` and `VERIFY.md:30`: object not built, no `.sha256`, `make` never run (no make on Windows host, no stable 6.8 host reachable). `pkg/probe/bpf` has **zero loader calls**: no `cilium/ebpf` import, no `Open/Load/Attach/Tear-down` [R1][R8], no map reader, no `Session`. Only `ObjectBytes` (`assets_embedded.go:10`) / `Status` (`status.go:11`) / `ObjectSHA256` (`status.go:28`) exist. Therefore `KernelSource()` always falls to `tcpinfo` via `BuildError` (`assets_absent.go:8`) on any real host today.
- The C file’s four sections (`raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established`, `kprobe/tcp_rcv_established`, `raw_tp/kfree_skb`) and two maps exist **only as source** (`bpf_probe.c:60`, `:111`–`:251`); ELF sections, verifier log, and SHA have never been observed.

### 7.2 TCP_INFO tier is honest but lossy (important)

- Works where Linux: raw `SYS_GETSOCKOPT(TCP_INFO)` [R5] into Go `syscall.TCPInfo`, `Total_retrans*1448` (`tcpinfo_linux.go:20`). Non-Linux → `ErrTCPInfoUnavailable` (`tcpinfo_other.go:14`), never estimated.
- Gaps: (a) ×1448 is bench-MSS proxy, wrong under offload/non-1448 MSS (repo’s own warning `tcpinfo_linux.go:12`–`:15`); modern alternative `tcpi_bytes_retrans` (`__u64` [R6]) is ignored. (b) `keepConn` + sample-before-close (`bulkstats.go:17`–`:56`) is correct but partial-coverage gate (all-or-nothing, flow-observed definition) is **not in code** — `bulkNStats` sums whatever TCP_INFO returns and marks `RetransMeasured=oneTCP` (`bulkstats.go:55`), with no `Source`/`Kernel` plumbing. (c) No `TCP_CC_INFO` usage; no `delivery_rate`/`min_rtt` exposure [R6].

### 7.3 Unverified kernel facts carried as repo comments (do not promote)

Each of these is stated in the repo as fact, but no fetched primary source [R1]–[R9] corroborates it. They stay `unverified — primary source not found` until a kernel-doc/manual fetch covers them:

1. `tcp_retransmit_skb` raw args are `(sk, skb)` and `kfree_skb` raw args are `(skb, location, reason)` — `bpf_probe.c:114` vs `:190`, spec §2.1–§2.2. The fetched kernel docs index [R1] does not list these prototypes.
2. `srtt_us` is stored left-shifted by 3, hence `>>3` (`bpf_probe.c:167`) and spec §2.4. No fetched source shows `struct tcp_sock` `srtt_us` encoding.
3. `skb->len` at the retransmit tracepoint is payload with `transport_header==0xFFFF` sentinel (`bpf_probe.c:6`–`:9`). Behavioural claim about `tcp_transmit_skb` cloning; no fetched source.
4. `SKB_DROP_REASON_QDISC_DROP` timing: `__qdisc_drop()` static-inline + `tcf_set_drop_reason()` + `kfree_skb_list_reason()` path, and the 6.8-vs-7.1 `qdisc_drop` tracepoint history (commit `ff2998f29f39`, v7.1) — spec §2.1. Kernel git history not fetched.
5. `SAMPLE_NS=50ms`, `MAX_FLOWS=4096`, “64-octet `flow_stat` matches Go `bpfFlowStat`” (`bpf_probe.c:26`, `:49`) — design constants, not kernel facts; the Go mirror struct does not exist on disk (`types.go` has no `bpfFlowStat`).

### 7.4 Small-object + bench constants

- `dscpEF=46` (RFC 3246) and `46<<2=184` TOS (`dscp_linux.go:16`, `:24`) — RFC 3246 not fetched; mark `unverified — primary source not found` for the EF value as a primary claim (the code is the claim).
- `retransMSS=1448` (veth MTU 1500) (`tcpinfo_linux.go:15`) — bench-measured proxy; no fetched source derives 1448. Keep as documented proxy, never as kernel exactness.

### 7.5 IPv6 / netns / drop attribution limits (known, documented)

- IPv6/non-IPv4 retransmits go to `globals` only (`bpf_probe.c:121`–`:128`); flow key omits netns (spec §3.2 shortcut `netns,tuple` upgrade noted). Per-flow drop attribution requires existing flow entry (`bpf_probe.c:244`–`:249`); else global only. All faithful to the source but **unverifiable from fetched sources** as production guarantees — they are scope statements, not kernel theorems.

## 8 Slices

Stop-and-ask triggers (AGENTS.md): new dependency (`cilium/ebpf`), live bench run, frozen-data touch — each requires human approval before acting. No slice touches code in this pass.

| Slice | Scope | Acceptance (named command / host) |
|-------|-------|-----------------------------------|
| **0 — Extraction & pin** | Move `pkg/probe`+`bpf` to new module; wire `go.mod` stdlib-only; keep all `file:line` APIs byte-identical; add `make bpf/bpf-check` Makefile shape. No loader, no dep. | `go vet ./...`; `go test ./... -count=1` (Windows); `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` exit 0; `go list -m all` shows stdlib only. |
| **1 — TCP_INFO standalone** | Extract `TCPRetransBytes` + `DialWithCC`/`SetConnCC` + `keepConn`/`bulkNStats` + `BulkStats` as standalone lib+CLI (`tcpinfo` subcommand). Add optional full-`struct tcp_info` snapshot [R6] + `tcpi_bytes_retrans` exact-byte path alongside the 1448 proxy, never silent. | `go test ./...` incl. non-TCP-error + multi-flow `RetransMeasured` cases; live `cgo-bpf-observer tcpinfo` against loopback TCP shows non-zero `tcpi_total_retrans` under `tc netem loss` (manual bench, human-approved). |
| **2 — Object + loader** | Build `bpf_probe.o`+`.sha256` on pinned 6.8/clang-18/[R1] stack; pin `cilium/ebpf`; implement `Open/Check/Lookup/Snapshot/Globals/Attach/Close` [R1][R8]; replace `kernelFacts` stat/euid with BTF-parse + `capget` [R1][R7]; split `session_linux.go` (`//go:build linux && amd64`) / stub. Verifier log captured. | On 6.8 VM with BTF + `CAP_BPF`: `doctor ok`; loopback flow in `Snapshot()` with `samples>0`; forced retransmit yields `retrans_payload>0` matching TCP_INFO order-of-magnitude; without caps → `tcpinfo` with named reason. `make bpf-check` exit 0; `TestEmbeddedObjectPinned` passes. |
| **3 — Coverage contract & audit wiring** | `BulkStats.Source/Kernel`, `kernelSession` seam, sender-key derivation (upload client, download reversed), query-before-close, all-or-nothing gate (require *all* dialed flows observed with ≥1 sample/event; qdisc-drop alone ≠ observed; missing ≠ zero; exact zero has a sample). Wire `KernelSource`→`Summary`→`KernelSample` neutral adapter. | Unit: partial-coverage + map-full + zero-with-sample + zero-without-event fall back to `tcpinfo` with coverage `n/N` + reason. `dualns` upload+download report `source=ebpf`, coverage 4/4 (human-approved live run). |
| **4 — Hardening** | Pin + fuzz key-encoding (host-order addr/port), perf overhead (<2% goodput delta per spec §6.5), `vmlinux.h` reproducibility docs, 7.1+ `qdisc_drop` note (later gate, not this slice). | Overhead run recorded; key-encoding test green; reproducible SHA documented. Human decides 7.1 port. |

## 9 Anti-toy criteria

A build that fails **any** of these is toy — fixtures only, never called `ebpf` in README/report:

1. **Object real:** `bpf_probe.o` + `.sha256` committed, built on the pinned 6.8/clang-18/[R1] stack; `make bpf-check` exit 0; `readelf -S` shows `raw_tp/*`, `fentry/*`, `kprobe/*`; embedded SHA == pinned SHA (`TestEmbeddedObjectPinned`).
2. **Loader real:** `go.mod` pins `cilium/ebpf`; `Open` implements [R1] Open/Load/Attach/Tear-down with verifier-log capture; `Snapshot()` under real traffic returns `samples>0`; `Close()` detaches cleanly. Unit-only `Status` returns without any attach do not count.
3. **Attribution exact:** bulk with `tc netem loss` produces `retrans_payload>0` and `retrans_events>0` on sender-side flows, same order-of-magnitude as TCP_INFO `Total_retrans*1448` on the same window; `srtt_us`/`snd_cwnd` samples non-stale (>50 ms freshness). Zero-flow fixture JSON does not count.
4. **Fallback named, never silent:** every non-`ebpf` window carries `Source=tcpinfo|""` + `FallbackReason` naming the missing fact (non-Linux, missing BTF, no `CAP_BPF`, verifier/CO-RE, attach, partial coverage, map-full, remote sender) — the ladder in `status.go:11` plus the spec §5 table. Missing ≠ zero; exact-zero requires a sample.
5. **No silent MSS:** TCP_INFO path labels its byte count as proxy (`×1448`) unless `tcpi_bytes_retrans` exact path is used; eBPF path labels payload-vs-wire honestly (`skb->len` semantics stated, never “ACKed”).
6. **Cross-build + Windows gates green:** `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` exit 0; `go vet` + `go test ./... -count=1` exit 0 on Windows; `GOOS=linux go vet` where supported.
7. **Independent verification:** live attach, overhead (<2%), and dualns 4/4 claims re-run by a verifier who is not the builder, on a named host, with date + command + output/quoted checkpoint (VERIFY row rule). Windows-green explicitly not “done”.
8. **No BBRv3/retransmodel smuggling:** CC selection success is never presented as BBR model evidence (`max_bw`/`min_rtt` [R9] are not exposed by this tool). BBRv3 matrix/engine remain separate BLOCKED rows until a 7.1 host + human-approved run exists.

## 10 Claims ledger

| Claim | Source | Status |
|-------|--------|--------|
| eBPF app lifecycle is Open→Load→Attach→Tear-down; skeleton embeds bytecode | [R1] §BPF App Lifecycle, §Skeleton | Verified primary |
| CO-RE needs BTF at `/sys/kernel/btf/vmlinux`; `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h` generates `vmlinux.h` | [R1] §CO-RE | Verified primary |
| BTF `types`/`strings` are the locator CO-RE matches against | [R2] §BTF intro | Verified primary |
| Maps are created/looked up via `bpf()` syscall; hash/percpu-array are valid map types | [R3] §Map Types | Verified primary |
| `bpf_map_lookup_elem` / `bpf_map_update_elem(BPF_NOEXIST)` / `bpf_ktime_get_ns` have the [R4] signatures used in `bpf_probe.c` | [R4] helper list | Verified primary |
| `TCP_INFO` returns `struct tcp_info` from `<linux/tcp.h>`; `TCP_CONGESTION` sets CC per-socket by string; allowed vs available CC lists bound it | [R5] §TCP_INFO, §TCP_CONGESTION | Verified primary |
| `TCP_INFO=11`, `TCP_CONGESTION=13 (0x0d)`; `struct tcp_info` fields `tcpi_total_retrans`, `tcpi_bytes_retrans`, `tcpi_snd_cwnd`, `tcpi_delivery_rate`, `tcpi_min_rtt` exist | [R6] UAPI tcp.h | Verified primary |
| `CAP_BPF` (5.8+) governs privileged BPF; `CAP_PERFMON` (5.8+) governs perf; `CAP_SYS_ADMIN` still implies same ops but is overloaded; euid alone ≠ effective caps | [R7] capabilities list | Verified primary |
| `cilium/ebpf` is a pure-Go eBPF loader with `link`/`btf`/`rlimit`/`features`/`cmd/bpf2go` packages, MIT, Linux amd64/arm64 | [R8] repo README | Verified primary |
| BBRv3 is model-based (delivery-rate/RTT/loss → `max_bw`/`min_rtt`/BDP; pacing/`cwnd`/`send_quantum`; Startup→Drain→ProbeBW/ProbeRTT) and TCP `C.SMSS` SHOULD be payload-only | [R9] Abstract, §2, §3 | Verified primary |
| Repo has no `bpf_probe.o`, no `.sha256`, no loader, no `cilium/ebpf` dep; Go half is contract-only | `BUILD.md:8`, `go.mod:5`, `VERIFY.md:29`, `VERIFY.md:30` | Verified repo fact |
| `tcp_retransmit_skb`/`kfree_skb` raw-arg orders + `srtt_us>>3` + `skb->len==payload` + 6.8/7.1 `qdisc_drop` history | Repo comments/spec only; no fetched source | `unverified — primary source not found` |
| `retransMSS=1448` is the correct MSS for all paths; DST/EF `46<<2` | Repo bench proxy (`tcpinfo_linux.go:15`, `dscp_linux.go:16`); no fetched derivation | `unverified — primary source not found` |
| Short-`struct tcp_info` prefix-fill behaviour | Comment `tcpinfo_linux.go:32`–`:34`; not in [R5][R6] | `unverified — primary source not found` |

## 11 Sources

1. Linux kernel documentation — BPF: https://docs.kernel.org/bpf/index.html ; libbpf Overview: https://docs.kernel.org/bpf/libbpf/libbpf_overview.html [R1]; BTF: https://docs.kernel.org/bpf/btf.html [R2]; BPF maps: https://docs.kernel.org/bpf/maps.html [R3].
2. `bpf-helpers(7)`: https://man7.org/linux/man-pages/man7/bpf-helpers.7.html [R4].
3. `tcp(7)`: https://man7.org/linux/man-pages/man7/tcp.7.html [R5]; `capabilities(7)`: https://man7.org/linux/man-pages/man7/capabilities.7.html [R7].
4. Linux UAPI `include/uapi/linux/tcp.h` (torvalds/linux master): https://raw.githubusercontent.com/torvalds/linux/master/include/uapi/linux/tcp.h [R6].
5. `cilium/ebpf`: https://github.com/cilium/ebpf [R8].
6. IETF `draft-ietf-ccwg-bbr-07`: https://datatracker.ietf.org/doc/html/draft-ietf-ccwg-bbr-07 [R9].
7. Repo files cited inline as `file:line` (e.g. `pkg/probe/bpf/bpf_probe.c`, `pkg/probe/bpf/BUILD.md`, `Makefile`, `docs/VERIFY.md`) and the design spec `docs/superpowers/specs/2026-10-27-cgo-ebpf-ground-truth-design.md`.
