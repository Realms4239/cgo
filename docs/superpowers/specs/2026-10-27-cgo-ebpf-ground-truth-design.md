# CGO eBPF Ground-Truth Telemetry — Design Specification

*Date: 2026-10-27 · Status: approved · Slice 1 of the
`epbfBBR.md` overhaul · Branch: `rebuild/plan1-economics-audit` ·
Target: Ubuntu 24.04 LTS, Linux 6.8, amd64.*

The reference bench stays on 6.8. Linux 7.0 does **not** add the native
`qdisc_drop` tracepoint (first 7.1), and changing the kernel mid-slice would
invalidate the frozen bench. A 7.x portability run is a later gate, not a
reason to move the reference host.

---

## 0. Context

`epbfBBR.md` asks for five subsystems: eBPF ground-truth telemetry, a
trace-driven netem `.dist` calibrator, IFB/RouterOS ingress shaping, a BBRv3
3×3 matrix, and UI/thesis surfacing. This specification covers **only the
first subsystem**. The other four remain separate slices; BBRv3 is not
touched here.

The current audit already computes enterprise Models A/B/emergency, the
220 ms CAP SLA, the Gilbert–Elliott bursty channel, the four-namespace
`dualns` bench, and RouterOS remediation. The weak link is the retransmission
numerator: `pkg/probe/tcpinfo_linux.go` multiplies `TCP_INFO.Total_retrans`
by a fixed 1448-byte MSS. That is a documented proxy, not a kernel-exact byte
count, and it is wrong under offload, fragmentation, or a non-1448 MSS.

This slice replaces that proxy **where the sender runs on our kernel**, keeps
the proxy as an explicit fallback, and records which source produced every
number. Nothing in the frozen corpus is re-measured or rewritten. The
existing guard reports **159 runs / 208 valid lines**; the master prompt's
phrase “208 valid runs” is not restated as a run count here.

### 0.1 In scope

- A CO-RE eBPF object embedded in the single Go binary.
- Exact submitted retransmit payload counters (successful TCP transmit calls),
  sampled socket state, and qdisc-drop counters.
- Both-end visibility when both endpoints share one kernel (`dualns`,
  `legacy`, loopback).
- A userspace TCP_INFO fallback with a typed reason.
- Additive audit/report provenance: `source` and `kernel_telemetry`.
- A real `cgo doctor` capability check.
- Build, reproduction, and test tooling for the object.

### 0.2 Out of scope

Queue dwell time, ECN, IPv6 drop attribution, BBRv3, the trace-driven
`.dist` calibrator, IFB/RouterOS ingress, UI redesign, and the thesis chapter.
The dashboard header badge is deferred to the UI slice; this slice exposes the
mode through the report and `cgo doctor`.

Remote field endpoints are also out of scope for **sender-side** exactness:
when the load is downlink and the sender is not on our kernel, its socket is
invisible. The audit must fall back and say so. It must never present a local
receiver's counters as the remote sender's retransmissions.

### 0.3 Success criteria

1. On the `dualns` bench, an upload and a download audit both attach eBPF and
   report exact submitted sender-side retransmit payload bytes, with flow
   coverage shown.
2. On a host without BTF, privileges, or a supported architecture, the audit
   completes with `source=tcpinfo` or an empty ratio and a stated reason.
3. The report states the telemetry source, attach mode, object SHA-256, flow
   coverage, retransmit counters, and qdisc-drop counters.
4. `telemetry.csv`, `link_audit.csv`, their headers, and the frozen corpus are
   byte-identical to today.
5. Windows and `GOOS=linux GOARCH=amd64` builds stay green; no runtime clang
   is required on the VM.
6. The feature is not called “verified” until the opt-in Linux integration
   test runs on 6.8 with BTF. The current CGO-DEV VM instability is a known
   blocker, not a reason to fabricate a result.

---

## 1. Decision log (this slice)

| # | Fork | Decision |
|---|---|---|
| D1 | First slice | **eBPF ground truth first.** Calibrator, IFB, BBRv3, and UI follow in separate specs. |
| D2 | Coverage | **Both ends plus qdisc drops.** On a shared kernel one probe sees the client and the sink; no per-packet dwell claim. |
| D3 | Build artifact | **Commit a prebuilt `bpf_probe.o`** with a pinned SHA-256 and a `BUILD.md`. `make bpf` is maintainer-only; Go build and the VM never need clang. |
| D4 | Provenance | **Additive `source` field** and a `kernel_telemetry` block. `basis: measured|estimated` keeps its current meaning; frozen CSVs unchanged. |
| D5 | 6.8 qdisc hook | **`skb:kfree_skb` raw tracepoint filtered to `SKB_DROP_REASON_QDISC_DROP`.** Linux 6.8 has no `qdisc_drop` tracepoint; see §2.1. |
| D6 | Mixed windows | **All-or-nothing per window.** If any flow is missing or unobserved, the window uses TCP_INFO and records coverage; sources are never summed. |
| D7 | UI badge | **Deferred.** This slice exposes the mode in `report.json` and `cgo doctor`; the header badge belongs to the UI slice. |
| D8 | Scope guard | **No BBRv3, calibrator, IFB, or thesis chapter here.** Decision #8 of the 2026-10-25 spec (BBRv1 in the thesis) is not silently changed. |
| D9 | Platform | Real loader only on `linux && amd64`; every other target gets a stub that returns a typed unavailable error. |
| D10 | Exactness claim | Count **successful retransmit submissions** at the `tcp_retransmit_skb` raw tracepoint, not ACKed bytes. The tracepoint fires after fragmentation/collapse; failed attempts are not counted. No queue dwell, no ECN, no ACK pairing. |
| D11 | Source values | `retrans_ratio.source` is `"ebpf"`, `"tcpinfo"`, or omitted; `basis` stays `"measured"`, `"estimated"`, or `""`. |
| D12 | Verification | Windows gates and a Linux cross-build are merge gates; live attach is an opt-in integration test and a manual VM acceptance gate. |

### 1.1 Relationship to the existing spec

- This spec **supersedes the mechanism** of decision #9 of
  `2026-10-25-cgo-enterprise-rebuild-design.md` (TCP_INFO only) for hosts
  where eBPF is available. It preserves that decision's honesty rule: a
  measured/TCP_INFO basis is never confused with an estimated proxy.
- It does **not** change honesty invariant #8: the current thesis statement
  still means BBRv1, and BBRv3 remains unmerged. The later BBRv3 slice must
  explicitly supersede that invariant with its kernel-provisioning plan.
- It does **not** change the frozen-corpus counts, the CSV schemas, or the
  existing `dualns` topology.

---

## 2. Kernel facts and hooks

### 2.1 The 6.8 correction

The master prompt names `tracepoint/qdisc/qdisc_drop`. That tracepoint does
not exist in Linux 6.8. It was introduced later:

- commit `ff2998f29f39` — “net: sched: introduce qdisc-specific drop reason
  tracing”, Jesper Dangaard Brouer;
- author date 2026-02-26, merge date 2026-02-28;
- first release tag containing it: **v7.1**; absent from v6.8 through v6.19
  and v7.0.

The 6.8 kernel already has `skb:kfree_skb`, whose raw arguments are
`(skb, location, reason)`. `__qdisc_drop()` is a static inline in
`include/net/sch_generic.h`: it queues the skb on the `to_free` list instead of
freeing it under the qdisc lock. `__dev_xmit_skb()` then frees that list via
`kfree_skb_list_reason()` → `__kfree_skb_reason()`, with
`SKB_DROP_REASON_QDISC_DROP` set by `tcf_set_drop_reason()` before enqueue.
The tracepoint therefore fires with that reason. On 6.8 this is the correct
qdisc-drop hook, and it needs no per-kernel recompilation because the object is
CO-RE and prebuilt. On 7.1+ the native `qdisc_drop` tracepoint can replace it
in a later change; this slice does not carry two paths.

The drop program updates counters in-kernel; userspace performs **no per-drop
syscall** and reads aggregated map values at window end. That is what “zero
userspace overhead” means here. It is not zero total cost: the tracepoint and
map update still run in kernel context, and the 2% overhead acceptance test is
the bound.

The master prompt also says `tcp_retransmit_skb` gives the exact segment
length. The formatted tracepoint payload has only `skbaddr`/`skaddr`, no
length. But the tracepoint's `TP_PROTO` is `(const struct sock *sk, const
struct sk_buff *skb)`, so a raw tracepoint program receives both pointers and
can read the final `skb->len` after `__tcp_retransmit_skb` has fragmented or
collapsed the skb. No function probe is needed for exact bytes.

### 2.2 Hook table

| Hook | Attach | Arguments / fields used | Purpose |
|---|---|---|---|
| `tcp_retransmit_skb` | raw tracepoint | `(sk, skb)`, `skb->len` (TCP payload) | Exact submitted retransmit payload per flow, after fragmentation/collapse; fires on successful `tcp_transmit_skb` only. |
| `tcp_rcv_established` | fentry, kprobe fallback | `struct sock *sk`, `struct sk_buff *skb` | Sampled `srtt_us >> 3`, `snd_cwnd`, `snd_ssthresh`, socket state. |
| `skb:kfree_skb` | raw tracepoint | `args[0]=skb`, `args[1]=location`, `args[2]=reason` | Qdisc drops only when `reason == SKB_DROP_REASON_QDISC_DROP`; parse IPv4/TCP headers for per-flow bytes. |

`tcp_rcv_established` is `void tcp_rcv_established(struct sock *, struct
sk_buff *)` in 6.8; it is global and attachable with BTF. fentry is preferred;
kprobe is compiled into the same object as a fallback.

A fentry on `tcp_retransmit_skb` would run **before** `__tcp_retransmit_skb`,
which may fragment or collapse the skb to the congestion-window allowance. Its
`skb->len` is then the queued super-packet, not the transmitted piece, and the
remainder can be counted again later. The tracepoint fires after that
adjustment, inside `__tcp_retransmit_skb`, on the success path. That is why the
retransmit hook is a raw tracepoint, not fentry.

### 2.3 Byte accounting

At the `tcp_retransmit_skb` tracepoint, the `skb` is the retransmit-queue skb,
not the clone that carries the headers. `tcp_transmit_skb(..., clone_it=1)`
pushes the TCP/IP headers on a clone; the queued skb keeps its payload. So
`skb->len` is exactly the TCP payload of the segment or GSO super-packet:

```
payload = skb->len
```

No header is subtracted, and there is no header-parse failure. The kernel uses
the same value in this function (`tp->bytes_retrans += skb->len`), which is an
independent confirmation of the unit. The economic ratio uses **payload** bytes
over payload bytes transported, matching the existing `Bytes` counter. A
`retrans_events` counter records one per successful tracepoint event; it is
diagnostic and is not a GSO segment count.

Wire bytes including headers are not observable at this hook. The slice stores
no wire counter rather than estimating one from a guessed header size.

The tracepoint can fire for SYN/FIN or a zero-payload retransmit; those count
as submitted events with zero payload. Failed attempts do not fire the
tracepoint, so they are not counted. The report says “réémissions transmises”,
never “ACKed” or “delivered”. ACK pairing and success-after-ACK accounting are
deferred.

### 2.4 Socket state sampling

`tcp_rcv_established` fires on the established receive path. Writing a map
entry for every ACK would turn telemetry into overhead, so each socket is
sampled at most once per **50 ms**:

```
now = bpf_ktime_get_ns()
if now - last_sample_ns < 50_000_000: return
last_sample_ns = now
srtt_us      = tp->srtt_us >> 3
snd_cwnd     = tp->snd_cwnd
snd_ssthresh = tp->snd_ssthresh
state        = sk->__sk_common.skc_state
```

This is a monitoring sample, never a per-packet claim. `srtt_us` is stored
left-shifted by 3 in the kernel, hence the shift. `snd_cwnd` and
`snd_ssthresh` are in segments.

### 2.5 Address and port encoding

For IPv4, **every key field is host order**. The eBPF side converts the
socket's raw `__be32`/`__be16` fields before writing the key:

```
local = bpf_ntohl(sk->__sk_common.skc_rcv_saddr) : sk->__sk_common.skc_num
peer  = bpf_ntohl(sk->__sk_common.skc_daddr)     : bpf_ntohs(skc_dport)
```

`skc_dport` is `__be16`; `skc_num` is host-order `__u16`. Packet parsing uses
the same conversion (`bpf_ntohl(iph->saddr)`, `bpf_ntohs(tcph->source)`), so a
socket key and a packet key agree. The Go side uses
`binary.BigEndian.Uint32(addr.As4())` (which yields the same host-order
number) and the port as-is; a unit test pins this representation. A mixed
byte order would silently miss every flow, so this is tested explicitly.

IPv6 is not in the flow key for this slice; an IPv6 flow is counted only in
the global retransmit totals if the socket family is not `AF_INET`, and it
never claims per-flow coverage. The bench is IPv4.

For `skb:kfree_skb`, the 4-tuple is parsed from the packet headers
(`skb->head + skb->network_header`), with bounds checks against
`skb_headlen(skb)` for the linear header and `skb->len` for the packet length.
IPv4/TCP only. Non-IP, fragments, and malformed headers increment global drop
counters only.

### 2.6 No dwell time

There are no enqueue/dequeue timestamp programs and no skb-pointer correlation
map. The tracepoint gives drop events, not queue residence. The report must
not call drop counts “dwell”, “sojourn”, or “queue delay”.

---

## 3. Object, maps, and loader

### 3.1 Source and sections

`pkg/probe/bpf/bpf_probe.c` contains:

```c
SEC("raw_tp/tcp_retransmit_skb") int on_retransmit(struct bpf_raw_tracepoint_args *ctx);
SEC("fentry/tcp_rcv_established") int BPF_PROG(on_ack, ...);
SEC("kprobe/tcp_rcv_established") int BPF_KPROBE(on_ack_kprobe, ...);
SEC("raw_tp/kfree_skb") int on_kfree_skb(struct bpf_raw_tracepoint_args *ctx);
```

The raw tracepoints use plain functions, not `BPF_PROG`: their context is the
args array. For `tcp_retransmit_skb`, `ctx->args[0]` is `sk` and
`ctx->args[1]` is `skb`; for `kfree_skb`, `ctx->args[0]` is `skb`. The fentry
and kprobe variants for `tcp_rcv_established` share `static __always_inline`
bodies so the socket sampling cannot diverge. The object is compiled against a
generated `vmlinux.h` with BTF (`-g`) and CO-RE relocations. The drop program
resolves the reason with `bpf_core_enum_value(enum skb_drop_reason,
SKB_DROP_REASON_QDISC_DROP)` rather than a hard-coded integer.

### 3.2 Flow map

One per-flow hash map:

```c
struct flow_key {
    __u32 saddr;   /* local IPv4, host order */
    __u32 daddr;   /* peer IPv4, host order */
    __u16 sport;   /* host order */
    __u16 dport;   /* host order */
    __u16 family;  /* AF_INET only in this slice */
    __u16 pad;
};

struct flow_stat {
    __u64 retrans_payload;
    __u64 retrans_events;
    __u64 drop_packets;
    __u64 drop_bytes;
    __u32 srtt_us;
    __u32 snd_cwnd;
    __u32 snd_ssthresh;
    __u32 state;
    __u64 last_sample_ns;
    __u32 samples;
};
```

Maximum 4096 entries. Counters use atomic adds; scalar socket samples are
last-writer-wins. A `map_full` counter is incremented when insertion fails, and
the window falls back rather than reporting an incomplete total. Global
counters (`retrans_total`, `drop_total`, `drop_bytes_total`, `map_full`) live
in a per-CPU array and are summed by userspace.

`drop_bytes` is `skb->len` of the dropped skb (wire length), not TCP payload;
`drop_packets` counts free events. The qdisc totals are **host-wide**: they
include unrelated traffic and every namespace. Per-flow drop attribution
exists only for a parsed IPv4/TCP 4-tuple *and* an existing flow entry;
otherwise the drop is counted globally only. The report presents drops as a
diagnostic counter, never as a per-flow exact claim.

The key omits the network namespace. On the bench, the client/server 4-tuples
are unique; a same-tuple collision across namespaces would merge counters.
This is a documented bench-only limit, not a production claim:

```
shortcut: flow key has no netns; same 4-tuple in two namespaces merges,
upgrade to (netns, tuple) if background traffic can collide.
```

### 3.3 Loader and lifecycle

`pkg/probe/bpf/session_linux.go` uses `cilium/ebpf`:

1. `rlimit.RemoveMemlock()` best-effort (memcg accounting on 6.8).
2. Load the embedded object with `ebpf.NewCollection`.
3. Attach the `tcp_retransmit_skb` raw tracepoint. This is core: if it fails,
   the session cannot claim retransmit bytes and the window falls back to
   TCP_INFO.
4. Attach `tcp_rcv_established` with fentry. Attach mode is all-or-nothing: if
   fentry attach fails, detach and attach the kprobe variant. If kprobe also
   fails, socket sampling is off and the window falls back to TCP_INFO.
   `attach` always names one mode, never a mix.
5. Attach the `kfree_skb` raw tracepoint. This attach is independent: if it
   fails, retransmit/socket telemetry stays active and qdisc drops are marked
   unavailable.
6. Detach and close at session end.

Each `Start()` creates its own collection and maps; there is no global
session state, so concurrent audits cannot reset each other's counters. The
session is created once per bulk load, not per flow. `bulkNStats` queries each
flow **before** the real socket close; the map entry survives a close, but
reading before close avoids ephemeral-port reuse within the same window.
`Lookup` and `Snapshot` are safe for concurrent use because the bulk engine
queries them from N flow goroutines.

A collection **load** is all-or-nothing: if a CO-RE field is missing, the
verifier rejects a program, or BTF is absent, the whole session falls back to
TCP_INFO. Only an attach failure of `tcp_rcv_established` tries kprobe. Since
the raw tracepoint program is part of the same collection, a load failure there
also disables socket telemetry; splitting the object into multiple collections
is a possible later refinement, not part of this slice.

### 3.4 Privileges and BTF

Required: `CONFIG_DEBUG_INFO_BTF`, `/sys/kernel/btf/vmlinux`, `CAP_BPF`,
`CAP_PERFMON`; `CAP_NET_ADMIN` is already required for `tc`. Root is the
supported path. The doctor detail suggests:

```
sudo setcap cap_bpf,cap_perfmon,cap_net_admin+ep ./cgo
```

File capabilities are build-sensitive; root remains the documented fallback.
No runtime clang, gcc, llvm, or linux-headers.

### 3.5 Build and provenance

Committed:

- `pkg/probe/bpf/bpf_probe.c`
- `pkg/probe/bpf/bpf_probe.o` (CO-RE ELF, embedded)
- `pkg/probe/bpf/bpf_probe.o.sha256`
- `pkg/probe/bpf/BUILD.md` — clang version, libbpf headers, BTF source,
  exact command, object SHA-256.

`vmlinux.h` is generated by `make bpf` from
`/sys/kernel/btf/vmlinux` and gitignored. The committed object is the release
artifact; rebuilding is a maintainer action. `make bpf-check` verifies the
embedded object against the pinned SHA. A maintainer rebuild with `make bpf`
on the pinned toolchain and BTF must reproduce that SHA; if it does not, the
build is not reproducible and must be investigated, not silently accepted.
The exact clang is pinned in `BUILD.md` (expected: clang 18 on Ubuntu 24.04).

`go.mod` pins a specific `github.com/cilium/ebpf` release. The real loader is
`//go:build linux && amd64`; all other targets use a stub with the same
exported API.

---

## 4. Go API and audit integration

### 4.1 `pkg/probe/bpf` API

```go
type UnavailableError struct{ Reason string } // host limitation
type BuildError struct{ Reason string }       // embedded object defect

func (e *UnavailableError) Error() string
func (e *BuildError) Error() string

func ObjectSHA256() string
func Check() error                    // verify hash, then load-only for doctor
func Start() (*Session, error)        // load + attach; fresh maps
func (s *Session) Lookup(local, peer netip.AddrPort) (Flow, bool)
func (s *Session) Snapshot() []Flow
func (s *Session) Globals() (Globals, error)
func (s *Session) Attach() string     // fentry | kprobe (socket sample)
func (s *Session) Close() error

type Globals struct {
    DropPackets    uint64
    DropBytes      uint64
    MapFull        uint64
    QdiscAvailable bool
}
```

`Globals` reads the per-CPU diagnostic map: host-wide qdisc drops, `map_full`,
and whether the drop probe is attached. It carries no per-flow claim.
`Attach` names the socket-sample attach mode for `Summary.Attach`.

`Lookup` matches a flow by its local/peer tuple. The caller passes the
**sender's** local and peer: the client socket for upload, the server socket
for download. For a download flow, userspace constructs the server-side key
from the client connection's `RemoteAddr()` (local) and `LocalAddr()` (peer).
No server file descriptor is needed. If the download sender is remote, no
local socket produces that key and the lookup misses; the window falls back
and says “émetteur distant”.

`Check()` first compares the embedded object's SHA-256 with
`bpf_probe.o.sha256`; a mismatch is a `BuildError`. On `linux && amd64` it
then loads and closes the collection without attaching. A host limitation
(missing BTF, `EPERM`, verifier/CO-RE rejection) is an `UnavailableError`.
Loading is not attaching: the doctor must not claim attach capability.

The report-facing data types are platform-neutral and live in
`pkg/probe/bpf/types.go` (no build tag):

```go
type Flow struct {
    Local          string `json:"local"`
    Peer           string `json:"peer"`
    RetransPayload uint64 `json:"retrans_payload_bytes"`
    RetransEvents  uint64 `json:"retrans_events"`
    DropPackets    uint64 `json:"qdisc_drop_packets"`
    DropBytes      uint64 `json:"qdisc_drop_bytes"`
    SRTTus         uint32 `json:"srtt_us"`
    SndCwnd        uint32 `json:"snd_cwnd"`
    SndSsthresh    uint32 `json:"snd_ssthresh"`
    State          uint32 `json:"state"`
    Samples        uint32 `json:"samples"`
}

type Summary struct {
    Available       bool   `json:"available"`
    Source          string `json:"source,omitempty"` // ebpf | tcpinfo | ""
    Attach          string `json:"attach,omitempty"` // fentry | kprobe (socket sample)
    FallbackReason  string `json:"fallback_reason,omitempty"`
    ObjectSHA256    string `json:"object_sha256,omitempty"`
    FlowsSeen       int    `json:"flows_seen"`
    FlowsTotal      int    `json:"flows_total"`
    RetransPayload  uint64 `json:"retrans_payload_bytes"`
    RetransEvents   uint64 `json:"retrans_events"`
    QdiscDrops      uint64 `json:"qdisc_drop_packets"`
    QdiscDropBytes  uint64 `json:"qdisc_drop_bytes"`
    QdiscAvailable  bool   `json:"qdisc_available"`
    MapFull         uint64 `json:"map_full,omitempty"`
    Flows           []Flow `json:"flows,omitempty"`
}
```

`Source` is the ratio source (`ebpf`, `tcpinfo`, or empty); `Available` says
whether the eBPF session attached; `FallbackReason` says why the ratio did not
use eBPF. A summary can therefore be `available=true, source=tcpinfo` when one
flow was missing. `QdiscAvailable` is independent.

`RetransPayload` and `RetransEvents` are sums over the sender-side flows
selected for this window; an event is one successful tracepoint hit, not a GSO
segment. `QdiscDrops`, `QdiscDropBytes`, and `MapFull` are host-wide
diagnostics; the report labels the drop counters as host-wide and never mixes
them into the retransmission ratio.

`pkg/probe` aliases the report type as `probe.KernelTelemetry = bpf.Summary`
so `pkg/audit` does not import the loader package directly.

### 4.2 Bulk load

`probe.BulkStats` gains:

```go
Source string               `json:"source,omitempty"` // ebpf | tcpinfo | ""
Kernel *KernelTelemetry     `json:"kernel,omitempty"`
```

`bulkNStats` accepts a sender-side key function and does:

1. Start a session (or record the typed unavailable reason).
2. For each dialed connection, derive the sender's local/peer tuple, wrap the
   connection with `keepConn`, run the send/receive, then query the
   sender-side flow before the real close.
3. A flow is **observed** only if it exists and has at least one socket sample
   or retransmit event from the socket programs; a qdisc drop alone does not
   prove the retransmit accounting. A missing flow is never treated as zero.
   A flow with zero retransmits but a sample is an exact zero.
4. Require **all** dialed flows to be observed and parsed. If so, sum their
   `retrans_payload`, set `RetransBytes` exactly, set
   `RetransMeasured=true`, `Source="ebpf"`, and copy the summary. Otherwise
   use the existing TCP_INFO sample for every flow, keep
   `RetransMeasured` from TCP_INFO, set `Source="tcpinfo"` when measured,
   and keep the summary as diagnostic coverage with the reason.
5. If the session could not start, `Source` is `"tcpinfo"` when TCP_INFO
   measured at least one socket, otherwise `""`; the summary carries the
   reason.
6. Failed dials are not counted as flows; `flows_total` is the number of
   successfully dialed sockets, matching `Bytes`.

`BulkSendNStats` uses the client's `LocalAddr`/`RemoteAddr`;
`BulkDownloadNStats` uses the reverse (`RemoteAddr`/`LocalAddr`) for the
server socket. `BulkSendNTo` (campaign) is unchanged in this slice; its
retransmission integration belongs to the campaign/BBRv3 slice.

`RetransMeasured` is the existing “a measurement exists” flag. It is true for
a complete eBPF window and for a TCP_INFO read; `windowRatio` keys on it, so a
measured zero is not mistaken for a missing counter.

The test seam is an unexported interface, so tests can fake partial coverage:

```go
type kernelSession interface {
    Lookup(local, peer netip.AddrPort) (bpf.Flow, bool)
    Snapshot() []bpf.Flow
    Globals() (bpf.Globals, error)
    Attach() string
    Close() error
}

var startKernel = func() (kernelSession, error) { return bpf.Start() }
```

### 4.3 Metrics

`metrics.RetransRatio` gains:

```go
Source string `json:"source,omitempty"` // ebpf | tcpinfo | ""
```

`basis` is unchanged. `windowRatio`:

| Condition | Ratio | Basis | Source |
|---|---|---|---|
| eBPF, all flows observed, payload known | `RetransBytes / Bytes` | `measured` | `ebpf` |
| TCP_INFO measured (`RetransMeasured`) | `RetransBytes / Bytes` | `measured` | `tcpinfo` |
| loss proxy only | `LossPct / 100` | `estimated` | `` |
| no load / no counter / no loss | — | `` | `` |

A measured zero is a measured zero: the branch keys on `RetransMeasured`, not
on `RetransBytes > 0`. `pickRatio` ranks `ebpf` above `tcpinfo` above
`estimated`; on equal rank it keeps the larger ratio. Both window ratios are
stored separately in the result for the report; the non-headline window never
enters the headline economics.

### 4.4 Audit result and report

`audit.LoadOutcome` gains:

```go
RetransSource string
Kernel        *probe.KernelTelemetry
```

`audit.Result` gains JSON-only fields (never CSV columns):

```go
KernelUp         *probe.KernelTelemetry `json:"kernel_up,omitempty"`
KernelDown       *probe.KernelTelemetry `json:"kernel_down,omitempty"`
RetransRatioUp   *metrics.RetransRatio  `json:"retrans_ratio_up,omitempty"`
RetransRatioDown *metrics.RetransRatio  `json:"retrans_ratio_down,omitempty"`
```

`Run` copies each window's `Kernel` into `KernelUp`/`KernelDown` and stores both
window ratios; a window that did not run leaves its fields nil. The headline
`Result.RetransRatio` keeps its current meaning and source.

`audit.Report` gains:

```go
KernelTelemetry *KernelTelemetry `json:"kernel_telemetry,omitempty"`

type KernelTelemetry struct {
    Up   *KernelWindow `json:"up,omitempty"`
    Down *KernelWindow `json:"down,omitempty"`
}

type KernelWindow struct {
    *probe.KernelTelemetry
    RetransRatio *metrics.RetransRatio `json:"retrans_ratio,omitempty"`
}
```

`Up`/`Down` are nil when that window did not run. A window with a summary but
no ratio omits `retrans_ratio`; there is no zero-filled placeholder.
`BuildReport` copies `KernelUp/Down` and `RetransRatioUp/Down` into the report
block; the existing headline `Result.RetransRatio` is unchanged. Example:

```json
"kernel_telemetry": {
  "up": {
    "source": "ebpf",
    "available": true,
    "attach": "fentry",
    "object_sha256": "…",
    "flows_seen": 4,
    "flows_total": 4,
    "retrans_payload_bytes": 12345,
    "retrans_events": 9,
    "qdisc_drop_packets": 3,
    "qdisc_drop_bytes": 4344,
    "qdisc_available": true,
    "retrans_ratio": {"ratio": 0.0012, "basis": "measured", "source": "ebpf"},
    "flows": [ … ]
  }
}
```

The Markdown report gains one section, in the report's existing French voice:

```
## Télémétrie noyau

- Mode : eBPF (échantillon socket fentry) — 4/4 flux · 12 345 octets de
  réémission transmis (payload) · 9 événements.
- Drops qdisc (hôte, tous flux) : 3 paquets / 4 344 octets (raison QDISC_DROP).
- Objet : sha256:…
```

or, on fallback:

```
- Mode : repli TCP_INFO — CAP_BPF absent ; le ratio reste mesuré par TCP_INFO.
```

or, when no kernel counter exists at all:

```
- Mode : non mesuré — linux/amd64 requis ; aucun ratio noyau.
```

If no ratio exists at all, the existing `Unavailable` economics entry remains;
the telemetry section says “non mesuré” and why.

### 4.5 Doctor

`pkg/doctor/doctor.go` adds a check named `ebpf`:

- `ok` — `bpf.Check()` verified the SHA and loaded the embedded object; the
  detail says “objet chargeable (BTF + CAP_BPF) ; l'attache est vérifiée par
  l'audit”. It does not claim attach.
- `warn` — observe mode, unsupported OS/arch, missing BTF, missing
  capabilities, or a verifier/CO-RE load error; detail is the typed reason and
  says TCP_INFO fallback.
- `fail` — `Check()` returned `BuildError` (embedded object empty or SHA
  mismatch), a build defect rather than a host limitation.

The check is read-only and does not attach probes. `/api/doctor` includes it
automatically because it renders `doctor.Report`.

### 4.6 Compatibility

- `telemetry.csv`: header and rows unchanged; no new columns.
- `link_audit.csv`: header and columns unchanged; new `Result` fields are not
  written by `AppendLinkAudit`.
- Report JSON: additive optional fields; existing consumers that ignore them
  keep working.
- Report Markdown: one new section only when a window has a kernel summary or
  ratio.
- No new HTTP endpoint.
- `epbfBBR.md` remains an untracked user artifact.

---

## 5. Failure ladder

Every row is a normal path, not an exception. The audit never panics because
telemetry is unavailable.

| Condition | Behavior | Report |
|---|---|---|
| Windows, or non-amd64 | Stub; no load | `source=tcpinfo` or empty; reason “linux/amd64 required” |
| `--mode observe` | No attach | Doctor warns; audit uses TCP_INFO if run |
| Missing `/sys/kernel/btf/vmlinux` | Load fails | `source=tcpinfo`; reason “BTF missing” |
| Missing `CAP_BPF`/`CAP_PERFMON` | Load fails `EPERM` | `source=tcpinfo`; reason “capabilities” |
| Object verifier/CO-RE failure | Load fails | `source=tcpinfo`; reason “verifier/CO-RE” |
| `tcp_retransmit_skb` attach fails | Core telemetry off | `source=tcpinfo`; reason “attach” |
| `tcp_rcv_established` fentry attach fails | Try kprobe | `attach=kprobe` |
| fentry and kprobe fail | Socket sampling off | `source=tcpinfo`; reason “attach” |
| `kfree_skb` attach fails | Retransmits stay eBPF | `qdisc_available=false`; no drop claim |
| Flow not observed (no sample/event) | Whole window TCP_INFO | coverage `n/N`; reason “flow not observed” |
| Map full | Whole window TCP_INFO | `map_full>0`; reason “map full” |
| Socket closes before read | Query-before-close; if still missing, fallback | coverage/reason |
| TCP_INFO also fails | No ratio | empty; existing economics `Unavailable` |
| External downlink sender | No local server event | `source=tcpinfo` or empty; reason “émetteur distant” |

The report distinguishes “eBPF was not attempted” from “eBPF was attempted and
fell back”. The latter carries the typed reason; the former is the normal
Windows/observe path.

---

## 6. Tests and verification

### 6.1 Windows / default gates

- `go test ./... -count=1`
- `go vet ./...`
- frontend typecheck, 37 files / 139 tests, build
- frozen-corpus guard (`audit-archive.py`) — expected unchanged output
- detector — expected two pinned warnings, exit 2

The eBPF package must compile as a stub on Windows and keep these gates
unchanged.

### 6.2 Cross-build

```
GOOS=linux GOARCH=amd64 go build ./...
```

on the Windows host, so the Linux-only loader and the cilium/ebpf dependency
are type-checked even though they cannot run here. `go vet` may be run with the
same environment where supported.

### 6.3 Unit tests

- Flow-key encoding: host-order address/port representation, IPv4 only,
  canonical local/peer; Go `binary.BigEndian` matches the C helper.
- Sender selection: upload uses client, download uses server; reverse-key
  construction for the sink; remote download falls back.
- Fallback ladder: no session, partial flows, a flow with zero retransmits but
  a sample (exact zero), a flow with no event (not zero), map full, TCP_INFO
  failure.
- `windowRatio`/`pickRatio`: all source/basis combinations; measured zero;
  eBPF never chosen on partial coverage.
- Report JSON: `kernel_telemetry` present/absent; `retrans_ratio_up` and
  `retrans_ratio_down`; `source` propagates through
  `economics.retrans_ratio`; no zero-filled counters.
- Markdown: mode line in eBPF, TCP_INFO fallback, and no-counter cases.
- CSV: `AppendLinkAudit` header and row length unchanged.
- Object: embedded SHA-256 matches `bpf_probe.o.sha256`; object non-empty;
  SHA mismatch yields `BuildError`.
- Doctor: `ok`/`warn`/`fail` statuses with a fake checker or the stub.

### 6.4 Opt-in Linux integration

Build tag `ebpf_integration`, Linux/amd64 only, skipped unless root and BTF are
present:

1. `bpf.Check()` succeeds.
2. `Start()` attaches the retransmit and drop raw tracepoints and the socket
   sample; `Attach` names fentry or kprobe.
3. A loopback TCP flow appears in `Snapshot()` with `samples > 0`.
4. `Close()` detaches cleanly.

Forcing a retransmit is a manual bench step, not a CI assertion: run the
`dualns` bench with `tc netem loss`, upload and download, and verify
`source=ebpf`, coverage 4/4, retransmit payload > 0, and qdisc drops when the
queue overflows. A download from a remote server is a separate check: it must
fall back with “émetteur distant”, not claim a local sender. The VM's current
soft-lockup instability blocks this step; the feature is not marked verified
until it runs.

### 6.5 Overhead acceptance

A 60 s audit with eBPF on and off on a stable Linux host: goodput delta under
2%, all phases executed, no audit error. If the delta exceeds 2%, the 50 ms
socket sampling or the drop tracepoint is the first suspect; the fallback
remains available.

---

## 7. Risks and open questions

1. **Hook availability.** The `tcp_retransmit_skb` raw tracepoint is core; if
   it cannot attach, retransmit telemetry falls back to TCP_INFO.
   `tcp_rcv_established` is global in 6.8, but fentry can be rejected by
   kernel config or BTF; kprobe is the fallback, and both failing disables
   socket sampling.
2. **CO-RE field drift.** `srtt_us`, `snd_cwnd`, `snd_ssthresh`, and the drop
   reason enum exist in 6.8. A future kernel rename makes the collection fail
   to load; the fallback is explicit, not silent.
3. **Raw tracepoint layout.** `tcp_retransmit_skb` raw args are `(sk, skb)` and
   `skb:kfree_skb` raw args are `(skb, location, reason)` in 6.8. A future
   change to either tracepoint requires a new object and a new hash; the old
   object is pinned.
4. **Submitted vs delivered.** The tracepoint fires on a successful
   `tcp_transmit_skb`, but a qdisc drop can still follow. The slice counts
   submitted retransmit bytes, not ACKed or delivered bytes. The report says
   “réémissions transmises”; ACK pairing is deferred.
5. **Netns collisions.** The 4-tuple key omits netns; the bench is unique.
   Background traffic with identical tuples would merge counters. Documented
   shortcut, upgrade to `(netns, tuple)` if needed.
6. **Drop timing.** `skb:kfree_skb` reports the free event, which may be
   batched after the enqueue decision. It is a drop event, not an enqueue
   timestamp; no dwell is inferred.
7. **Host-wide drops.** The qdisc drop counters include every interface and
   namespace on the host, and only per-flow attribution is filtered by the
   4-tuple. They are diagnostic, never a per-flow exact claim, and never feed
   the retransmission ratio.
8. **Key representation.** A mixed host/network byte order would silently miss
   every flow and could turn a real retransmission into a false zero. The C
   helpers and the Go encoder must agree; the unit test and the integration
   sample guard this.
9. **External sender.** A remote downlink sender is invisible. The report must
   not claim sender-side exactness for field audits.
10. **VM stability.** Live attach cannot be accepted on the current CGO-DEV VM
    while it soft-locks. The Windows/cross-build gates can merge the code, but
    the thesis claim waits for a stable 6.8 run.
11. **BBRv3 conflict.** The master prompt requires BBRv3; the existing
    invariant says BBRv1. This slice deliberately does not resolve it. The
    BBRv3 spec must present the kernel-provisioning risk and explicitly
    supersede the invariant, not change it in passing.
12. **7.1+ opportunity.** Once the target kernel is 7.1 or newer, the native
    `qdisc_drop` tracepoint can replace the `kfree_skb` filter, with the
    qdisc-specific reason enum. Not in this slice.

---

## 8. Implementation order and commits

Commit per logical group; never `git add -A`; leave `epbfBBR.md`, frozen
corpora, and user artifacts untouched.

1. **`bpf: objet CO-RE, chargeur et repli TCP_INFO`**
   - `pkg/probe/bpf/{bpf_probe.c,bpf_probe.o,bpf_probe.o.sha256,BUILD.md}`
   - `types.go`, `session_linux.go`, `session_stub.go`, object/hash tests
   - `pkg/probe/kernel.go` alias, `go.mod`/`go.sum`, root `Makefile`
     targets, `.gitignore` for `vmlinux.h`
2. **`probe/audit: télémétrie noyau dans le ratio de réémission`**
   - `pkg/probe/bulkstats.go`, `bulkstats_test.go`
   - `pkg/metrics/cost.go`, `cost_test.go`
   - `pkg/audit/audit.go`, `audit_test.go`
3. **`audit/report: bloc kernel_telemetry et provenance`**
   - `pkg/audit/report.go`, `report_test.go`
   - `pkg/doctor/doctor.go`, doctor test
   - `docs/api.md` (doctor check; no endpoint change)
4. **`docs/tests: preuves eBPF et garde-fous`**
   - integration test, acceptance notes, `BUILD.md` finalization
   - any report fixture updates that are strictly additive

No commit may touch `pkg/audit/telemetry.go`'s header, `data/runs/`, the
frozen CSV, or the frontend.

---

## 9. Acceptance checklist

- [ ] `make bpf-check` passes (or, on Windows without make, the Go object test
      passes); `bpf_probe.o.sha256` matches the embedded object.
- [ ] `GOOS=linux GOARCH=amd64 go build ./...` passes on Windows.
- [ ] `go test ./... -count=1`, frontend typecheck/tests/build, frozen guard,
      detector all pass on Windows.
- [ ] `cgo doctor` reports `ebpf ok` with root + BTF; `warn` with a host
      reason; `fail` only on a `BuildError`.
- [ ] `dualns` upload and download audits report `source=ebpf`, coverage
      `4/4`, exact submitted payload bytes, and qdisc drops where present.
- [ ] No-BTF, no-capability, partial-flow, zero-retransmit-without-event, and
      external-downlink runs report `source=tcpinfo` or empty with the reason;
      no fabricated zero.
- [ ] `report.json` contains `kernel_telemetry` with object SHA and per-window
      counters, plus `retrans_ratio_up`/`retrans_ratio_down`;
      `economics.retrans_ratio.source` is correct.
- [ ] Markdown report states the mode and the fallback in French.
- [ ] `telemetry.csv` and `link_audit.csv` headers/columns are unchanged.
- [ ] No frozen corpus, `epbfBBR.md`, or untracked user artifact is modified.

---

## 10. References

- `epbfBBR.md` §3 (eBPF directive), §7 (interface stability), §8 (deliverables).
- `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`
  decisions #9 and honesty invariant #8.
- Linux 6.8: `include/trace/events/tcp.h` (`tcp_retransmit_skb` payload),
  `include/trace/events/skb.h` (`kfree_skb`), `include/linux/tcp.h`
  (`srtt_us`, `snd_cwnd`, `snd_ssthresh`), `include/net/sock.h`
  (`skc_dport`/`skc_num`), `include/net/sch_generic.h` (`qdisc_drop` inline),
  `net/core/dev.c` (`SKB_DROP_REASON_QDISC_DROP`).
- Linux 7.1: commit `ff2998f29f39`, “net: sched: introduce qdisc-specific
  drop reason tracing” (native `qdisc_drop` tracepoint, not used here).
- `pkg/probe/tcpinfo_linux.go` (fallback), `pkg/audit/audit.go`
  (`LoadOutcome`, `windowRatio`), `pkg/audit/report.go`,
  `pkg/metrics/cost.go`, `pkg/doctor/doctor.go`.
