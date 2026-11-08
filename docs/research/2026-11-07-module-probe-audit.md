# pkg/probe + pkg/audit as standalone tools — primary-source research

*Date: 2026-11-07 · Scope: `pkg/probe` (incl. `pkg/probe/bpf`) + `pkg/audit` only, read as if extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — FQ-CoDel (Experimental) | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | What FQ-CoDel/CoDel does/does not normatively define; defaults (target 5 ms, interval 100 ms, quantum 1514, flows 1024, packet limit MUST) |
| R2 | RFC 3246 — Expedited Forwarding PHB (DSCP EF) | https://www.rfc-editor.org/rfc/rfc3246.txt | fetched | DSCP 46 / codepoint 101110 authority for `dscp_linux.go` |
| R3 | tc(8) — Linux traffic control | https://man7.org/linux/man-pages/man8/tc.8.html | fetched | qdisc root/handle/egress, shaping/scheduling/policing, classless vs classful |
| R4 | tc-cake(8) | https://man7.org/linux/man-pages/man8/tc-cake.8.html | fetched | CAKE bandwidth/ingress/egress, diffserv tins incl. EF, `ingress` counts dropped bytes |
| R5 | tc-fq_codel(8) | https://man7.org/linux/man-pages/man8/tc-fq_codel.8.html | fetched | fq_codel limit/flows/target/interval/quantum/ecn defaults |
| R6 | tc-netem(8) | https://man7.org/linux/man-pages/man8/tc-netem.8.html | fetched | netem delay/jitter/loss/corrupt/duplicate/reorder/rate semantics |
| R7 | tc-tbf(8) | https://man7.org/linux/man-pages/man8/tc-tbf.8.html | fetched | TBF pure-shaper, tokens, peakrate, burst/limit |
| R8 | tcp(7) — Linux TCP | https://man7.org/linux/man-pages/man7/tcp.7.html | fetched | `TCP_CONGESTION` per-socket string, `TCP_INFO` via `getsockopt` returning `struct tcp_info` |
| R9 | BPF Type Format (BTF) — kernel docs | https://docs.kernel.org/bpf/btf.html | fetched | `/sys/kernel/btf/vmlinux`, `vmlinux.h` via `bpftool btf dump`, CO-RE relocation |
| R10 | libbpf Overview — kernel docs | https://docs.kernel.org/bpf/libbpf/libbpf_overview.html | fetched | Open/Load/Attach/Tear-down lifecycle, skeleton, CO-RE portability |
| R11 | eBPF verifier — kernel docs | https://docs.kernel.org/bpf/verifier.html | fetched | Verify-before-load, register/stack safety, map-lookup NULL checks |
| R12 | Linux Kernel Tracepoints — kernel docs | https://docs.kernel.org/trace/tracepoints.html | fetched | Tracepoint probe semantics (on/off, synchronous callback) |
| R13 | capabilities(7) — CAP_BPF / CAP_PERFMON / CAP_NET_ADMIN | https://man7.org/linux/man-pages/man7/capabilities.7.html | fetched | CAP_BPF (since 5.8) covers privileged BPF; CAP_NET_ADMIN covers TOS/interface config; root approximates caps |
| R14 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | fetched | Module path + version = unique ID + import prefix; `go` directive = minimum version; `require` forms |
| — | Waveform bufferbloat bands (A+..F thresholds) | — | unverified — primary source not found (only local comment `pkg/metrics/bufferbloat.go:6-7` and `pkg/audit/audit.go:70-74` cite it) | `BufferbloatGrade` bands provenance |
| — | Jain fairness index (JFI) formula | — | unverified — primary source not found | `metrics.JFI` / `JFIPct` provenance |
| — | ITU-T G.107 E-model (VoIP R) | — | unverified — primary source not found (only local comment `pkg/metrics/metrics.go:179-189`) | `Measured.VoIPR` provenance |
| — | RouterOS v7 `/queue type` / `/queue tree` export syntax | — | unverified — primary source not found (only local parser `pkg/audit/routeros_config.go:33-85`) | RouterOS compare/proof dialect |
| — | `ping(8)` output format (`time=` / `time<1ms`, FR `temps=`) and Windows `ping -n` semantics | — | unverified — primary source not found (only local parser `pkg/probe/probe.go:45-85`) | `Ping` / `parsePing` contract |
| — | `syscall.TCPInfo.Total_retrans` field semantics across kernel versions | — | unverified — primary source not found (only local use `pkg/probe/tcpinfo_linux.go:35-43` + comment about old-kernel zero-fill) | `TCPRetransBytes` conversion |
| — | 220 ms / 95 % CAP objective authority | — | unverified — primary source not found at primary level; local framing only: operational DGM objective, not WMO-mandated (`pkg/metrics/cost.go:75-83`, `pkg/audit/audit.go:114-122`) | `CAPSLA` framing |

Method: read `pkg/probe/*.go` + `pkg/probe/bpf/*.go`/`bpf_probe.c`/`BUILD.md`, all `pkg/audit/*.go`, `docs/architecture.md`, `go.mod`. Every technical claim below cites `file:line` or `(R#)`; anything else is labelled unverified — primary source not found.

## 1. Responsibility

`docs/architecture.md:15` describes `pkg/probe` as “Sondes concurrentes : ping, small, bulk” and `docs/architecture.md:19` describes `pkg/audit` as “Audit client d’un lien réel (idle/loaded), append `data/link_audit.csv`”. In source terms:

### 1.1 `pkg/probe` — active measurement primitives (no shaping, no report)

`pkg/probe/probe.go:1-4` states the contract: “sondes (ping, petits objets, transfert de masse). Toutes les sondes prennent leurs dépendances en interfaces pour tester partout ; l’exécution réelle se fait sur la VM.” Concretely across the package it does five things:

- **ICMP ping by exec** — `Ping` shells `ping -c/-i` on Unix, `ping -n` on Windows with `count×interval+5000ms` timeout, then parses `time=` lines (`pkg/probe/probe.go:28-43`, `pkg/probe/probe.go:49-85`). Sorting in `parsePing` (`pkg/probe/probe.go:56`) discards wire order. `<1ms` → half-interval heuristic (`pkg/probe/probe.go:81-83`).
- **HTTP small-object completion** — `SmallClient` (shared 2 s client, `pkg/probe/probe.go:93-96`) vs `SmallClientTimeout` (fresh client, `pkg/probe/probe.go:104-106`); `SmallObject` times `http.Get` to `io.Discard` in ms (`pkg/probe/probe.go:122-139`). Transport marks DSCP EF on Linux via `IP_TOS = 46<<2` (`pkg/probe/dscp_linux.go:12-30`); non-Linux is unmarked best-effort (`pkg/probe/dscp_other.go:11-17`). DSCP 46 is EF 101110 per (R2) §2.7 “Codepoint 101110 is RECOMMENDED for the EF PHB” and §4 allocates 101110 in Pool 1; (R4) §PRIORITY QUEUE lists EF in the Voice tin (“Voice (CS7, CS6, EF, VA, CS5, CS4)”).
- **TCP bulk flood (up/down, single/multi-flow)** — `BulkSend` (64 KiB zeroes, 2 s write deadline, until `ctx.Done`, `pkg/probe/bulk.go:20-40`) and its mirror `BulkSource` (`pkg/probe/bulksource.go:14-31`); receive side `BulkReceive` with cancel-via-read-deadline (`pkg/probe/bulk.go:43-75`); download hello `D` / `D:<cc>\n` via `BulkRecvSource`/`BulkDownloadTo`/`ParseDownloadHello` (`pkg/probe/bulksource.go:39-103`). Multi-flow: `BulkSendNFlows`/`BulkSendNTo` (`pkg/probe/bulkflows.go:23-64`), `BulkDownloadNTo` (`pkg/probe/bulksource.go:68-90`), stats variants (`pkg/probe/bulkstats.go:60-73`). Per-flow bytes feed JFI upstream (`pkg/probe/bulkflows.go:11-16`).
- **Per-socket congestion control + retransmission truth** — `DialWithCC` (`pkg/probe/bulk_cc_linux.go:14-28`: `TCP_CONGESTION = 0x0d` via `SetsockoptString`; (R8) §`TCP_CONGESTION` allows per-socket CC string) and `SetConnCC` for accepted sockets (`pkg/probe/ccsock_linux.go:15-34`); non-Linux stubs are no-ops (`pkg/probe/bulk_cc_other.go:11-14`, `pkg/probe/ccsock_other.go:7-10`). `TCPRetransBytes` (`pkg/probe/tcpinfo_linux.go:20-44`) raw `getsockopt(TCP_INFO)` (`syscall.TCPInfo`, `SYS_GETSOCKOPT`) and converts segments→bytes via `×1448` (`pkg/probe/tcpinfo_linux.go:15`, `pkg/probe/tcpinfo_linux.go:43`); non-Linux returns `ErrTCPInfoUnavailable` (`pkg/probe/tcpinfo_other.go:10-14`). (R8) §`TCP_INFO` says the kernel returns `struct tcp_info` as defined in `/usr/include/linux/tcp.h`; exact `Total_retrans` field presence across kernels is unverified — primary source not found.
- **Kernel-tier ladder (eBPF vs TCP_INFO)** — `KernelSource` (`pkg/probe/kernelsource.go:34-38`) probes facts (`pkg/probe/kernelsource.go:19-31`: `/sys/kernel/btf/vmlinux` presence, `euid==0` approx, embedded object) and delegates to `bpf.Status` (`pkg/probe/bpf/status.go:11-25`): non-Linux→`tcpinfo`, missing object→`tcpinfo`, no BTF→`tcpinfo`, no caps→`tcpinfo`, else `ebpf`. `bpf.ObjectBytes` is build-tag-gated: with `ebpf_object` embeds `bpf_probe.o` (`pkg/probe/bpf/assets_embedded.go:7-15`), without it returns `BuildError` (`pkg/probe/bpf/assets_absent.go:7-9`). C source hooks are `raw_tp/tcp_retransmit_skb`, `fentry/tcp_rcv_established` + `kprobe/tcp_rcv_established` fallback, `raw_tp/kfree_skb` filtered on `SKB_DROP_REASON_QDISC_DROP` (`pkg/probe/bpf/bpf_probe.c:111-251`, `pkg/probe/bpf/BUILD.md:30-41`). Per (R9) BTF is exposed at `/sys/kernel/btf/vmlinux` and generated to `vmlinux.h` via `bpftool btf dump`; per (R10) libbpf lifecycle is Open→Load→Attach→Tear-down and CO-RE resolves `BPF_CORE_READ` offsets against runtime BTF. Maps are `flows` (HASH, 4096) + `globals` (PERCPU_ARRAY) (`pkg/probe/bpf/bpf_probe.c:60-72`).

What it deliberately does **not** do: it applies no `tc` shaping, computes no percentiles/grades, writes no CSV, renders no report — shaping is `pkg/qdisc` and summaries are `pkg/metrics` (`docs/architecture.md:14,16`). That negative scoping is what makes a probe-only module plausible in principle; §3 shows the two couplings that still break it.

### 1.2 `pkg/audit` — phased link audit + sealed operator report

`pkg/audit/audit.go:199-204` states `Run` executes a `Duration`-second audit collecting ping/small samples; default shape is 60 s = 10 s quiescent + 40 s saturated + 10 s drain (`pkg/audit/audit.go:719-748` via `phasePlan`). Concretely it does six things:

- **Three-window measurement harness** — shared baseline, then direction-aware load windows (`up`/`down`/`both` with 5 s mini-drain + 3 s settle, `pkg/audit/audit.go:414-424`) plus drain windows; `collectWindow` probes ping (5 per call, `pkg/audit/audit.go:309`) + small every 400 ms (`pkg/audit/audit.go:344-373`); `runLoadWindow` starts `BulkLoad` in a goroutine, sleeps 300 ms, collects during load, cancels, waits 5 s (`pkg/audit/audit.go:380-409`). Per (R1) §5.2.1/§5.2.2 defaults (interval 100 ms, target 5 ms) explain why a loaded window must be long enough to fill a queue; the audit’s 10/40/10 split is an operational choice, not an RFC requirement — RFC 8290 normatively requires only “some packet limit MUST be enforced” ((R1) §5.2.3).
- **Honesty invariants (the core of this audit)** — zero ping probes → error, never frozen zeros (`pkg/audit/audit.go:478-483`); no unmeasured grade (`pkg/audit/audit.go:562-565`, pinned by `pkg/audit/honesty_test.go:14-32`); `JFIMeasured` false for <2 or all-zero per-flow (`pkg/audit/audit.go:184-197`, `pkg/audit/audit.go:633-637`); `DrainMeasured` false when recovery uncollected (`pkg/audit/audit.go:567-607`); `CAPSLAMeasured` false → CSV blank columns, never “0 %” (`pkg/audit/audit.go:114-122`, `pkg/audit/writer.go:62-68`); `RetransRatio.Basis` ∈ `measured`/`estimated`/`""` with measured-biased `pickRatio` (`pkg/audit/audit.go:826-865`); `Telemetry` absent-probe = absent line, never 0 ms (`pkg/audit/audit.go:140-143`, `pkg/audit/telemetry.go:15-26`); loss = received vs *attempted* (`pingCalls×5`, `pkg/audit/audit.go:305-309`, `pkg/audit/audit.go:505-514`, `pkg/audit/audit.go:786-801`); throughput = real bytes / real `Elapsed`, fallback to window only when unmeasured (`pkg/audit/audit.go:812-824`); UTF-8 gate before any probe (`pkg/audit/audit.go:205-214`); cancelled context → no partial `Result` (`pkg/audit/audit.go:660-666`).
- **Raw evidence + double seal** — `Result.Telemetry []TelemetryRow` (`pkg/audit/audit.go:140-144`) → `telemetry.csv` with exact header (`pkg/audit/telemetry.go:28`, `pkg/audit/telemetry.go:45-74`) and strict reader rejecting header drift / empty-probe lines (`pkg/audit/telemetry.go:80-161`); `Report.Provenance` seals *both* CSV row file and telemetry (`pkg/audit/report.go:197-213`, `pkg/audit/report.go:508-532`); `ArtifactPointer` is `data/audits/<id>/audit_report.json` with slash separators, empty on bad id (`pkg/audit/artifacts.go:49-55`).
- **Reproducibility seal** — `Invocation` built from *normalized* `Params` (`pkg/audit/invocation.go:53-89`, `pkg/audit/audit.go:674`) quoting POSIX `cgo audit -site=… -duration=… -direction=… -flows=…`; empty command stays empty, never invented (`pkg/audit/invocation.go:91-123`).
- **Operator report + remediation generation** — `BuildReport` (`pkg/audit/report.go:255-504`) overwrites caller capacity from measurement (`pkg/audit/report.go:252-255`), refuses remediation without idle RTT + measured capacity (`pkg/audit/report.go:423-428`), emits Linux `tc` egress (`LinuxTC`) + ingress IFB (`LinuxTCIngress`) and RouterOS scripts via `pkg/remediate` (`pkg/audit/report.go:439-500`), and lists `Unavailable` reasons instead of zero-filling (`pkg/audit/report.go:368-408`). tc-side invariants: shaping is egress per (R3) §DESCRIPTION (“Shaping occurs on egress”; policing on ingress); CAKE `ingress` counts already-traversed dropped bytes per (R4) §OTHER PARAMETERS; netem impairs outgoing packets per (R6) §DESCRIPTION; TBF is a pure non-work-conserving shaper per (R7) §DESCRIPTION.
- **Bench discovery + profile matching** — `ResolveSinks` probes `10.30.0.2:5201` (dualns) → `10.200.0.1:5201` (legacy) → external `8.8.8.8`, no bulk (`pkg/audit/sinks.go:34-48`); `EnvOrBulk` honours `CGO_BULK_ADDR` (`pkg/audit/sinks.go:52-57`); `MatchProfile` requires measured throughput, two-axis distance (RTT p95/delay + throughput/capacity) over P1–P4 only (`pkg/audit/match.go:16-56`); `Presets` are three hardcoded Yas/Madagascar drills to `8.8.8.8` (`pkg/audit/presets.go:18-39`).

What it deliberately does **not** do: it does not shape the link itself (no `qdisc` import in `pkg/audit`); shaping stays in `pkg/qdisc` (`docs/architecture.md:14`) and remediation *generation* is delegated to `pkg/remediate`. The audit measures; the report recommends.

## 2. Exported API (file:line)

Complete inventory of exported identifiers in non-test sources. Unexported helpers (`parsePing`, `jfiComputable`, `phasePlan`, `collectWindow`, `runLoadWindow`, `shellQuote`, `countCSVRows`, `logicalLines`, `splitProps`, `isCommandVerb`, `normalizeROS`, `writeAtomic`, `probeTCP`, `resolveSinks`) are not API.

### 2.1 `pkg/probe` (non-test)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `PingSample{RTTms}` | struct | `pkg/probe/probe.go:22` | One completed ping RTT |
| `Ping(ctx,target,count,intervalMs)` | func | `pkg/probe/probe.go:28-43` | Exec ping; `-c/-i` Unix, `-n` Windows; `count×interval+5s` timeout; output parsed, error only if `len(out)==0` |
| `SmallClient()` | func | `pkg/probe/probe.go:93-96` | Shared 2 s client (`once`+`markedClient`, `pkg/probe/probe.go:108-111`) |
| `SmallClientTimeout(timeout)` | func | `pkg/probe/probe.go:104-106` | Fresh client per call (bench use); transport marked EF on Linux |
| `SmallObject(ctx,client,url)` | func | `pkg/probe/probe.go:122-139` | Times GET→`io.Discard`; returns ms as `us/1000.0` |
| `BulkSendTo(ctx,addr,cc)` | func | `pkg/probe/bulk.go:11-17` | Dial+`BulkSend` |
| `BulkSend(ctx,conn)` | func | `pkg/probe/bulk.go:20-40` | 64 KiB zeroes, 2 s write deadline, until `ctx.Done`; `Write` error → `return total, nil` (receiver-closed = end) |
| `BulkReceive(ctx,conn,onBytes)` | func | `pkg/probe/bulk.go:43-75` | 256 KiB reads; cancel via `SetReadDeadline(pastTime())` (`pkg/probe/bulk.go:77`) |
| `BulkSource(ctx,conn)` / `BulkRecvSource(ctx,conn,cc)` | funcs | `pkg/probe/bulksource.go:14-49` | Down-direction source + `"D"/"D:<cc>\n"` hello then `BulkReceive` |
| `BulkDownloadTo(ctx,addr,cc)` | func | `pkg/probe/bulksource.go:55-61` | Dial + `BulkRecvSource` |
| `BulkDownloadNTo(ctx,addr,cc,n)` | func | `pkg/probe/bulksource.go:68-90` | N parallel dials, per-flow return `[]uint64` in dial order |
| `ParseDownloadHello(b)` | func | `pkg/probe/bulksource.go:94-103` | Tolerant `"D"/"D:<cc>"`, spaces/`\n`, prefix case |
| `FlowResult{Bytes}` | struct | `pkg/probe/bulkflows.go:14-16` | Per-flow contribution for JFI |
| `BulkSendNFlows(ctx,conns,_)` | func | `pkg/probe/bulkflows.go:23-36` | N conns flooded in parallel; third arg ignored |
| `BulkSendNTo(ctx,addr,cc,n)` | func | `pkg/probe/bulkflows.go:42-64` | N dials, `""`→`model.Cubic` (`pkg/probe/bulkflows.go:43-45`); `n<=1` single-flow |
| `BulkStats{Bytes,RetransBytes,RetransMeasured}` | struct | `pkg/probe/bulkstats.go:11-15` | Per-flow bytes + summed retrans + whether any TCP socket sampled |
| `BulkSendNStats(ctx,addr,cc,n)` / `BulkDownloadNStats(ctx,addr,cc,n)` | funcs | `pkg/probe/bulkstats.go:60-73` | Shared `bulkNStats`: dial-failures skipped (`pkg/probe/bulkstats.go:36-38`), `keepConn` withholds `Close` so `TCP_INFO` sees live fd (`pkg/probe/bulkstats.go:17-25`) |
| `DialWithCC(ctx,addr,cc)` | func | `pkg/probe/bulk_cc_linux.go:14-28`; stub `pkg/probe/bulk_cc_other.go:11-14` | Linux: `net.Dialer{5s}` + `TCP_CONGESTION` control; non-Linux: plain dial, `cc` ignored |
| `SetConnCC(conn,cc)` | func | `pkg/probe/ccsock_linux.go:15-34`; stub `pkg/probe/ccsock_other.go:10` | Accepted-socket CC; non-TCP/`""` → `nil` silently |
| `TCPRetransBytes(c)` | func | `pkg/probe/tcpinfo_linux.go:20-44`; stub `pkg/probe/tcpinfo_other.go:14` | Linux: `*net.TCPConn`→`File()`→`getsockopt(TCP_INFO)`→`Total_retrans×1448`; non-TCP/non-Linux → error |
| `ErrTCPInfoUnavailable` | var | `pkg/probe/tcpinfo_other.go:12` | `errors.New("tcpinfo: TCP_INFO indisponible hors Linux")` |
| `KernelSourceInfo{Source,Reason}` | struct | `pkg/probe/kernelsource.go:12-15` | `Source` ∈ `ebpf`/`tcpinfo`/`unavailable` |
| `KernelSource()` | func | `pkg/probe/kernelsource.go:34-38` | Runs `kernelFacts` (`pkg/probe/kernelsource.go:19-31`) → `bpf.Status` |

### 2.2 `pkg/probe/bpf` (non-test)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Flow{Local,Peer,RetransPayload,RetransEvents,DropPackets,DropBytes,SRTTus,SndCwnd,SndSsthresh,State,Samples}` | struct | `pkg/probe/bpf/types.go:4-16` | One sender-side flow as kernel maps saw it |
| `Summary{Available,Source,Attach,FallbackReason,ObjectSHA256,FlowsSeen,FlowsTotal,RetransPayload,RetransEvents,QdiscDrops,QdiscDropBytes,QdiscAvailable,MapFull,Flows}` | struct | `pkg/probe/bpf/types.go:20-35` | One telemetry window; `Source` ∈ `ebpf`/`tcpinfo`/`""` |
| `Globals{RetransPayload,RetransEvents,DropPackets,DropBytes,MapFull,QdiscAvailable}` | struct | `pkg/probe/bpf/types.go:40-47` | Host globals (IPv6/non-IPv4 + full-table fallback) |
| `Status(goos,btf,caps,objectErr)` | func | `pkg/probe/bpf/status.go:11-25` | Fallback ladder; injected args so Windows tests each rung |
| `ObjectSHA256()` | func | `pkg/probe/bpf/status.go:28-35` | `sha256(ObjectBytes)` hex, for provenance |
| `ObjectBytes()` | func | `pkg/probe/bpf/assets_embedded.go:10-15` (`ebpf_object`); stub `pkg/probe/bpf/assets_absent.go:7-9` | Embedded vs absent object |
| `UnavailableError{Reason}` / `BuildError{Reason}` | structs + `Error()` | `pkg/probe/bpf/errors.go:4-11` | Typed host-limitation vs object-defect errors |

### 2.3 `pkg/audit` — engine (audit.go + helpers)

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Params{…}` | struct | `pkg/audit/audit.go:16-38` | `AuditID,Site,LinkType,Provider,Duration,Target,SmallURL,BulkAddr,Direction(""/up/down/both),Flows,SinkMode,Origin` |
| `Result{…}` | struct | `pkg/audit/audit.go:41-144` | 30+ JSON fields: headline + down-mirrors + `Invocation` + `Telemetry`; see honesty invariants §1.2 |
| `Deps{Ping,Small,BulkLoad,OnProgress}` | struct | `pkg/audit/audit.go:146-152` | Injection seams; `nil`→real probe defaults (`pkg/audit/audit.go:215-268`) |
| `Progress{Phase,ElapsedSec,PhaseRemainSec,LastRTTms,LastSmallMs,SamplesRTT,SamplesSmall,Timeline}` | struct | `pkg/audit/audit.go:157-169` | Live ≤32-tail snapshot; `phase` ∈ `repos/charge/charge-down/drain` |
| `LoadOutcome{Bytes,PerFlow,RetransBytes,RetransMeasured,Err,Elapsed}` | struct | `pkg/audit/audit.go:175-182` | One load window truth; `Elapsed` feeds throughput |
| `Run(ctx,p,d)` | func | `pkg/audit/audit.go:204-717` | The harness; see §1.2 + §5 gaps |
| `Invocation{Origin,SinkMode,Duration,Flows,Direction,Site,LinkType,Provider,Target,SmallURL,BulkAddr,Command}` | struct | `pkg/audit/invocation.go:15-34` | Sealed replay (`pkg/audit/invocation.go:58-89`); `Command` empty = pre-field archive |
| `KernelSample{Source,Attach,Reason,ObjectSHA256,FlowsSeen,FlowsTotal,RetransPayload,QdiscDrops,Flows}` | struct | `pkg/audit/kernel_adapter.go:6-16` | Neutral view so audit never imports `probe/bpf` (`pkg/audit/kernel_adapter.go:3-5`) |
| `KernelTelemetryFrom(s)` | func | `pkg/audit/kernel_adapter.go:20-35` | `""`→`nil` (absent stays absent, never defaults to `tcpinfo`) |
| `TelemetryRow{Phase,OffsetMs,RTTms,SmallMs,BulkBytes}` | struct | `pkg/audit/telemetry.go:20-26` | One received probe; all-zero row forbidden (`pkg/audit/telemetry.go:157-160`) |
| `ArtifactID(id)` | func | `pkg/audit/telemetry.go:33-40` | Basename-only, rejects `..`, separators, edge spaces |
| `WriteTelemetry(baseDir,res)` / `ReadTelemetry(baseDir,id)` | funcs | `pkg/audit/telemetry.go:45-74`; `pkg/audit/telemetry.go:80-117` | Writer emits `bulk_bytes` only when non-zero; reader requires exact header (`pkg/audit/telemetry.go:97-99`) |
| `TimelinePoint{T,Phase,N,RTTP50,RTTP95,SmallP95}` | struct | `pkg/audit/timeline.go:7-14` | Pointer nil = no probe that second, never 0 |
| `AppendLinkAudit(dataDir,r)` | func | `pkg/audit/writer.go:13-84` | Appends to `dataDir/link_audit.csv`, migrating 14→17→full header (`pkg/audit/writer.go:18-43`); unmeasured CAP → blank cols |
| `MatchProfile(r)` | func | `pkg/audit/match.go:16-56` | Two-axis P1–P4 match; `Throughput<=0`→`""` |
| `Preset{ID,Label,Site,LinkType,Provider,Duration,Target,Notes}` / `Presets()` | type + func | `pkg/audit/presets.go:7-39` | Three Yas/Madagascar drills |
| `Sinks{Mode,Target,SmallURL,BulkAddr}` / `ResolveSinks()` / `EnvOrBulk(s)` | type + funcs | `pkg/audit/sinks.go:14-57` | Bench discovery (see §1.2) |

### 2.4 `pkg/audit` — report, remediation state, RouterOS

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Report{…}` | struct | `pkg/audit/report.go:25-88` | `Measured,Phases,Unavailable,Grade/Verdict,OperatorVerdict,Provenance,Invocation,Artifact,Remediation,RemediationState/Available,Economics,CAPSLA,RouterOSValidation,KernelTelemetry` |
| `RouterOSValidation{Version,Level,ExportSHA256,ProofPath,Checks}` | struct | `pkg/audit/report.go:93-99` | CHR evidence carrier |
| `KernelTelemetry{…}` / `KernelFlow{…}` | structs | `pkg/audit/report.go:104-125` | eBPF/TCP_INFO tier as report saw it |
| `Phases{BaselineSeconds,LoadSeconds,RecoverySeconds}` | struct | `pkg/audit/report.go:130-134` | Executed windows (proves loaded window length) |
| `Measured{…}` | struct | `pkg/audit/report.go:137-189` | `QDI = loaded p95−loaded p50` (`pkg/audit/report.go:292`); `LoadMeasured` etc. |
| `Unavailable{Metric,Reason}` | struct | `pkg/audit/report.go:192-195` | “Honest half … never empty in practice” (`pkg/audit/report.go:36-38`) |
| `Provenance{SourceFile,SourceSHA256,RowCount,TelemetryFile,TelemetrySHA256,RowFileSHA256}` | struct | `pkg/audit/report.go:200-213` | Two seals: row file + telemetry |
| `Remediation{Qdisc,Reason,LinuxTC,RouterOS,LinuxTCIngress,RouterOSIngress,Up/DownMeasured/Shaped,LinuxIface,RouterOSWan/LanIface,RouterOSLanAssumed}` | struct | `pkg/audit/report.go:216-248` | Egress + ingress command sets, generated only from measurement |
| `BuildReport(res,link,qdisc,asymmetric,cpuConstrained)` | func | `pkg/audit/report.go:255-504` | See §1.2; `link.CapacityMbps/RTTms` overwritten from measurement |
| `Report.SealSource(path)` / `Report.SealTelemetry(path)` | methods | `pkg/audit/report.go:508-532` | SHA-256 file bytes at render time |
| `Report.UnmarshalJSON` / `Report.RenderJSON` / `Report.RenderMarkdown` / `Report.SortUnavailable` | methods | `pkg/audit/report.go:549-565`; `pkg/audit/report.go:563-565`; `pkg/audit/report.go:582-783`; `pkg/audit/report.go:800-804` | Back-compat default `observation_only`; ticket-pasteable MD |
| `WriteArtifacts(rep,baseDir)` | func | `pkg/audit/report.go:810-847` | `data/audits/<id>/audit_report.{json,md}` + `data/audit_report.{json,md}` mirror |
| `DefaultDataDir`, `ReportArtifactName`, `ReportMarkdownName`, `TelemetryArtifactName` | consts | `pkg/audit/artifacts.go:11-21` | `"data"`, `"audit_report.json"`, `"audit_report.md"`, `"telemetry.csv"` |
| `ArtifactDir(baseDir,id)` / `ArtifactFile(...)` / `ArtifactPointer(id)` | funcs | `pkg/audit/artifacts.go:28-55` | One convention (`filepath.Join(baseDir,"audits",clean)`) |
| `RemediationState` + `RemediationAvailableState/RemediationObservationState` | type + consts | `pkg/audit/remediation_state.go:8-13` | `"available"` / `"observation_only"`; `UnmarshalJSON` defaults missing→`observation_only` (`pkg/audit/remediation_state.go:20-34`); `Report.IsAvailable()` (`pkg/audit/remediation_state.go:37-39`) |
| `RouterOSObject{Section,Name,Props}` / `RouterOSConfig{Objects}` + `Get` + `ParseRouterOSConfig` | types + funcs | `pkg/audit/routeros_config.go:9-85` | Parses script + export dialects, keeps only `/queue type` + `/queue tree` (`pkg/audit/routeros_config.go:66`) |
| `RouterOSCheck{Name,Expected,Found,Pass}` / `CompareRouterOS(expected,found)` | type + func | `pkg/audit/routeros_compare.go:11-64` | Duplicate intended name = fail; `normalizeROS` absorbs quotes/case/`.0` only (`pkg/audit/routeros_compare.go:66-73`) |
| `CHRInfo/ImportResult/CHRProof` + `FileSHA256` + `WriteCHRProof(dir,p,export,transcript)` | types + funcs | `pkg/audit/routeros_proof.go:12-76` | Atomic temp+rename; `Level` must be `windows-tcg`/`linux-namespace` (`pkg/audit/routeros_proof.go:52-54`); schema `cgo.chr-proof/1` |

## 3. Dependencies

### 3.1 Outgoing (what the two packages need)

**`pkg/probe` — stdlib + two local edges.**

- `pkg/probe/probe.go:6-19` imports stdlib (`context,fmt,io,net,net/http,os/exec,runtime,sort,strconv,strings,sync,time`); `pkg/probe/bulk.go:3-8`, `bulkstats.go:3-7`, `bulksource.go:3-8`, `bulkflows.go:3-9`, `bulk_cc_linux.go:5-10`, `tcpinfo_linux.go:5-10`, `dscp_linux.go:5-10`, `ccsock_linux.go:5-8` are all stdlib-only (the `tcpinfo_linux.go` comment explicitly rejects `golang.org/x/sys` in favour of raw `syscall`, `pkg/probe/tcpinfo_linux.go:30-35`). No `require` in `go.mod:5-10` (`fyne`, `bubbletea`, `lipgloss`, `golang.org/x/sys`) is needed for the probe itself.
- Local edges: `pkg/model` only for the `Cubic` default string (`pkg/probe/bulkflows.go:8`, `pkg/probe/bulkflows.go:44`); `pkg/probe/bpf` for the ladder (`pkg/probe/kernelsource.go:7`, `pkg/probe/kernelsource.go:36`). `pkg/probe/bpf/status.go:3-6` imports only `crypto/sha256,encoding/hex`; `assets_embedded.go:5` imports `embed`.
- Build axes: `linux` vs `!linux` for CC/DSCP/TCP_INFO (`bulk_cc_linux.go:1`/`bulk_cc_other.go:1`, `ccsock_linux.go:1`/`ccsock_other.go:1`, `dscp_linux.go:1`/`dscp_other.go:1`, `tcpinfo_linux.go:1`/`tcpinfo_other.go:1`); `ebpf_object` tag for embedded vs absent BPF object (`assets_embedded.go:1`/`assets_absent.go:1`).

**`pkg/audit` — stdlib + four local edges.**

- Stdlib: `audit.go:3-10`, `report.go:3-12`, `telemetry.go:3-13`, `writer.go:3-11`, `routeros_proof.go:3-10`, etc. (csv, json, sha256, http-free, `path/filepath`, `unicode/utf8`).
- Local edges (grep `metrics\.|remediate\.` + imports): `pkg/metrics` — `Summarize` (`pkg/audit/audit.go:467-504`), `BufferbloatGrade/Verdict` (`pkg/audit/audit.go:544-555`), `JFI` (`pkg/audit/audit.go:626-627`), `AssessCAPSLA` (`pkg/audit/audit.go:650`), `VoIPR/BuildEconomics/AssessVerdict` (`pkg/audit/report.go:318-366`), `RetransRatio/Economics/CAPSLA/Verdict` types (`pkg/audit/report.go:46-79`); `pkg/remediate` — `Link`, `Recommend`, `ShapedRate`, `LinuxTC/LinuxTCIngress/RouterOS/RouterOSIngress`, `QdiscCake` (`pkg/audit/report.go:13-14`, `pkg/audit/report.go:431-470`); `pkg/model` — `Profiles/ProfilesMu` (`pkg/audit/match.go:6`), `LinkAuditHeader` (`pkg/audit/writer.go:10,19`); `pkg/profile` — `Load()` (`pkg/audit/match.go:7`); `pkg/probe` — default `Ping/Small/BulkLoad` (`pkg/audit/audit.go:11-12`, `pkg/audit/audit.go:215-268`).
- The whole module is `github.com/Realms4239/cgo` at `go 1.25` (`go.mod:1-3`). Per (R14) §"go", since Go 1.21 the `go` line is a mandatory minimum and toolchains refuse newer modules; per (R14) §"module", module path + version is the unique ID and import prefix. A standalone extraction needs its own `go.mod` + path (R14 §"module" Notes) before first publish.

### 3.2 Incoming (who needs them)

`docs/architecture.md:13-19` positions `pkg/campagne` `Deps` on “TC, shaper, sondes, `StatsFn`, `DeadlineMs`” (probe feeding campaign) and `pkg/audit` as client audit; `docs/architecture.md:60-74` gates G0–G7 (`TargetReachable` … `CPUNotSaturated`) on probe outputs (`BulkStarted`, `ProbesProducing`, …). Grep shows `probe.` callers in bench/campaign sinks and `audit.` callers in CLI/API/TUI/report sealing (all under `cmd/cgo` + `pkg/api`). Extraction must keep every one of these call sites compiling or provide an adapter (see §7).

### 3.3 Licence

`LICENSE` at repo root is MIT (per `LICENSE:3` as recorded for the sibling metrics module research; not re-read in this pass — the claim “MIT” for *this* pair inherits that evidence; if the licence file has changed, this row is stale). A standalone module inherits it cleanly only if the extracted repo keeps the same licence file.

## 4. Primary-source requirements

What each primary source **legitimately** requires of a standalone probe+audit module — and what it does not:

### 4.1 RFC 8290 (R1) — FQ-CoDel / CoDel framing

- RFC 8290 is **Experimental, “not an Internet Standards Track specification”** ((R1) Status of This Memo). It describes FQ-CoDel as “a combined packet scheduler and AQM algorithm” ((R1) §1) with per-queue CoDel ((R1) §2, referring normatively to RFC 8289), flow hashing on the 5-tuple with salted Jenkins hash ((R1) §4.1), new/old queue two-tier DRR ((R1) §4.2), and defaults interval 100 ms ((R1) §5.2.1), target 5 ms with MTU floor ((R1) §5.2.2), quantum 1514 bytes ((R1) §5.2.4), flows 1024 ((R1) §5.2.5). The only MUST in the parameter section is “some packet limit MUST be enforced” ((R1) §5.2.3).
- CAKE has **no RFC**: RFC 8290 cites it only informatively ((R1) §10.2 `[CAKE]`) and as “currently being developed” future work ((R1) §7). A module must never cite RFC 8290 as authority for CAKE behaviour.
- RFC 8290 defines **none** of the probe/audit interfaces: no ping parsing, no HTTP small-object, no bulk hello, no phased audit, no report schema, no RouterOS, no A+..F bands, no 220 ms/95 % objective. Its legitimate use is framing only — sojourn/target/interval vocabulary for *why* loaded latency is measured. Anything stronger is citation theatre.

### 4.2 tc(8) family (R3–R7) — shaping vocabulary for remediation, not for probes

- (R3) §DESCRIPTION: shaping is egress, scheduling egress, policing ingress, dropping both; qdiscs/classes/filters are the three object kinds. §CLASSLESS lists `codel`, `fq_codel`, `netem`, `tbf`, `pfifo_fast`; §CONFIGURING says classless qdiscs attach at root (`tc qdisc add dev DEV root QDISC …`). `docs/architecture.md:45-58` reset-then-apply topology (netem root `1:`, CAKE/fq_codel/TBF child) is consistent with (R3) root/parent syntax, but the audit itself runs no `tc` — it only *generates* commands via `pkg/remediate` (`pkg/audit/report.go:443-470`).
- (R4) §SHAPER PARAMETERS: CAKE deficit shaper bursts exactly as needed; §OTHER PARAMETERS `ingress` counts dropped bytes as transferred because they already traversed the link; §PRIORITY QUEUE `diffserv3` Voice tin includes EF; §FLOW ISOLATION default `triple-isolate`. (R5) §PARAMETERS: fq_codel defaults `limit 10240p`, `flows 1024`, `target 5ms`, `interval 100ms`, `quantum 1514`, `ecn` on. (R6) §DESCRIPTION: netem adds delay/loss/duplication/corruption; §OPTIONS loss `random PERCENT` etc. (R7) §DESCRIPTION/§ALGORITHM: TBF is pure shaper, non-work-conserving, token bucket with `peakrate` second bucket.
- Consequence for the module: a standalone probe/audit library must **document generated `tc` strings against (R3)–(R7) syntax** and must not claim to apply shaping; the `qdisc` package owns application.

### 4.3 TCP sockets (R8, R13) — probe socket truth

- (R8) §`TCP_CONGESTION` (since 2.6.13): per-socket string CC; unprivileged restricted to `tcp_allowed_congestion_control`, privileged (`CAP_NET_ADMIN`-adjacent) to any available. Probe’s `SetsockoptString(TCP_CONGESTION)` (`pkg/probe/bulk_cc_linux.go:20`) must therefore check/report errors (today `SetConnCC` returns `serr`, `DialWithCC` returns `serr` — good) and declare non-Linux a no-op (today silent, §5).
- (R8) §`TCP_INFO` (since 2.4): kernel returns `struct tcp_info` from `/usr/include/linux/tcp.h`, “should not be used in code intended to be portable”. Probe’s non-Linux error (`pkg/probe/tcpinfo_other.go:10-14`) is exactly that declaration; the `×1448` conversion (`pkg/probe/tcpinfo_linux.go:15`) is a bench-MTU proxy, not a kernel unit — must stay labelled as proxy.
- (R13) §Capabilities: `CAP_BPF` (since 5.8) separates BPF from `CAP_SYS_ADMIN`; `CAP_NET_ADMIN` covers interface/TOS (`set type-of-service (TOS)`) and `SO_PRIORITY`-adjacent ops. Probe’s `euid==0` approx for caps (`pkg/probe/kernelsource.go:29-30`) over-claims vs (R13) — root bypasses checks but a uid-0 process without the capability in its bounding set is still gated; the ladder must not promote `ebpf` on euid alone (see §5).

### 4.4 Kernel probe / eBPF (R9–R12) — what an eBPF-capable module must do

- (R9) §5.1/§5.2/§5.3: BTF enables map pretty-print, prog dump annotations, verifier line info; §6 BTF generation via llvm ≥8.0 or pahole; §3.1–3.3 `BPF_BTF_LOAD` → `BPF_MAP_CREATE`/`BPF_PROG_LOAD` with `prog_btf_fd`/`func_info`/`line_info`. `/sys/kernel/btf/vmlinux` is the running kernel’s authoritative BTF ((R10) §BPF CO-RE, `bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h`).
- (R10) lifecycle: Open (parse, discover maps/progs) → Load (create maps, resolve relocations, verify+load) → Attach (to tracepoints/kprobes/cgroup/packet hooks) → Tear-down (detach, unload, destroy maps). Skeleton embeds bytecode so “no extra files to deploy” (`<name>__open/__load/__attach/__destroy`). Today the Go side has **only** `ObjectBytes` (`pkg/probe/bpf/assets_embedded.go:10-15`) — none of the four phases exists in Go.
- (R11): verifier does DAG + path simulation, tracks register/stack types, requires NULL-checked `map_lookup_elem` before access, rejects uninitialized reads. The C code’s `flow_get_or_create` handles `BPF_NOEXIST` races + `map_full` global (`pkg/probe/bpf/bpf_probe.c:80-99`) — consistent with (R11), but no verifier log is ever captured in Go.
- (R12): tracepoints are lightweight hooks; when off, only branch+space cost; when on, probe runs synchronously in caller context. `raw_tp/tcp_retransmit_skb` + `raw_tp/kfree_skb` in `bpf_probe.c` are therefore *tracepoint programs*, requiring BPF load/attach privilege per (R13) — unattainable without the (R10) loader.
- `pkg/probe/bpf/BUILD.md:8-12` is honest: “L’objet n’est pas encore construit : aucun hôte Linux 6.8 stable n’était joignable.” Until `bpf_probe.o` + `.sha256` are produced on the pinned toolchain (`BUILD.md:16-20`, `BUILD.md:29-31`) and verified (`readelf` sections, `BUILD.md:35-41`), every eBPF claim is aspirational — `KernelSource` will always fall to `tcpinfo` (`pkg/probe/bpf/status.go:15-17`).

### 4.5 Go modules (R14)

- A module’s path + version is its unique ID and import prefix ((R14) §"module"); v2+ must end `/v2`. Today’s `github.com/Realms4239/cgo` (`go.mod:1`) is a monorepo path — a standalone needs a new controlled path before first publish ((R14) §"module" Notes). The `go` directive is a mandatory minimum since 1.21 ((R14) §"go"). `require` is `path version` (SemVer `vX.Y.Z` tags, not SemVer strings) with pseudo-versions for untagged commits ((R14) §"require"); tags are immutable once pushed ((R14) §"retract" Notes). `embed` (`assets_embedded.go:5-8`) and `syscall` stdlib use need no `require` — only the two local edges (§3.1) force a `replace` during migration if paths change.

## 5. Gap analysis (current → releasable)

Ordered by severity. Each gap cites local evidence; the primary side cites R# or is unverified — primary source not found.

1. **eBPF tier is object-less and loader-less.** `BUILD.md:8-12` says the object is not built; `go.mod` has no eBPF loader dependency; Go has only `ObjectBytes`/`Status`/`ObjectSHA256` (`pkg/probe/bpf/status.go`, `assets_embedded.go:10-15`) and no Open/Load/Attach/map-read code per (R10). `KernelSample`/`KernelTelemetryFrom` (`pkg/audit/kernel_adapter.go:6-35`) is a pure adapter with no producer. Fix: build `bpf_probe.o` on the pinned 6.8/clang-18/libbpf/bpftool stack (`BUILD.md:16-31`), add a Go loader (skeleton or cilium/ebpf — new dependency, stop-and-ask per `AGENTS.md`), attach `raw_tp/tcp_retransmit_skb` + `fentry/tcp_rcv_established` (+kprobe fallback) + `raw_tp/kfree_skb`, read `flows`/`globals` maps, and pin the verifier log + `ObjectSHA256` in tests. Until then every `ebpf` claim is toy (see §8).
2. **`pkg/audit/audit.go` is an 865-line god-harness couched in four other packages.** `Run` mixes phase planning, probing, load orchestration, drain math, loss, JFI, CAP, retrans-ratio, economics wiring, and `MatchProfile` (`pkg/audit/audit.go:204-717`) while importing `metrics` (`pkg/audit/audit.go:11`), `probe` (`pkg/audit/audit.go:12`), and indirectly `remediate` (via `report.go`), `model`/`profile` (via `match.go`/`writer.go`). A standalone audit cannot be built without first extracting a `metrics`-free engine + a `report` that takes an injected remediation provider (see §6). Slices in §7 sequence this.
3. **Bench/testbed identities are hardcoded in library code.** `ResolveSinks` hardcodes `10.30.0.2:5201` / `10.200.0.1:5201` / `8.8.8.8` + `:8081/small` URLs (`pkg/audit/sinks.go:41-48`); `Presets` hardcode Yas/Madagascar sites + `8.8.8.8` (`pkg/audit/presets.go:18-39`); report economics realise Yas/Airtel/Orange tariffs via `metrics` (`pkg/audit/report.go:327-334`). A releasable module must take sinks/presets/tariffs as caller data; the current constants are a deployment, not an API.
4. **Global mutable registries are read inside measurement.** `MatchProfile` calls `profile.Load()` and takes `model.ProfilesMu.RLock()` over the global `model.Profiles` map (`pkg/audit/match.go:20-22`). Concurrent audit + profile import can interleave; library users cannot inject a reference set. Same class as the metrics-module global-tier gap, but here it corrupts a *measurement verdict*, not just pricing.
5. **Probe measurement fidelity gaps (each is a “toy-signal” risk):**
   - *Ping by exec + locale parsing.* `parsePing`/`parsePingLine` (`pkg/probe/probe.go:49-85`) tolerates `temps=`/`temps<` and `<1ms`→`v/2` (`pkg/probe/probe.go:81-83`), but the ping-output contract has no primary source (unverified — primary source not found); `Ping` returns partial output as success when `len(out)>0` even on non-zero exit (`pkg/probe/probe.go:38-43`), and `parsePing` sorts, losing per-packet order. Windows `-n` has no `-i` equivalent (`pkg/probe/probe.go:31-34`) — interval is silently ignored.
   - *Small-object timeout split.* Shared 2 s client (`pkg/probe/probe.go:93-96`) vs fresh bench client (`pkg/probe/probe.go:104-106`); audit defaults to 8 s (`pkg/audit/audit.go:231`). Three timeouts, no documented rule for picking one; `SmallClientTimeout` allocates a fresh `Transport` per call (`pkg/probe/probe.go:105`), defeating connection reuse.
   - *Bulk payload + deadlines are heuristics.* Zero-filled 64 KiB writes (`pkg/probe/bulk.go:22`) measure goodput of compressible zeros (no entropy, fill policy unverified — primary source not found); 2 s write deadline (`pkg/probe/bulk.go:33`) is “≈20× VSAT margin” by comment only; `BulkSend` swallows `Write` errors as clean end (`pkg/probe/bulk.go:36-38`). `bulkNStats` skips dial failures (`pkg/probe/bulkstats.go:36-38`) so `Bytes` length < `n` silently; `RetransMeasured` is true if *any* socket sampled (`pkg/probe/bulkstats.go:48-56`) while `RetransBytes` sums only sampled sockets — ratio basis is then partial.
   - *RETRAN conversion.* `×1448` (`pkg/probe/tcpinfo_linux.go:15,43`) is a veth-MTU proxy; `skb->len` in eBPF counts payload TSO-inclusive (`pkg/probe/bpf/bpf_probe.c:6-11`) — the two tiers count different things, and no code reconciles them.
6. **Audit defaults are invisible tunables.** `Flows=4` (`pkg/audit/audit.go:269-271`), `Direction ""→up` (`pkg/audit/audit.go:272-278`), `Duration<=0→10s` (`pkg/audit/audit.go:280-282`), BBR in default loader (`pkg/audit/audit.go:243-248`), `pingPerCall=5` + 400 ms sleep (`pkg/audit/audit.go:309,371`), 300 ms load-start delay + 5 s grace (`pkg/audit/audit.go:395,405`), mini-drain 5 s + settle 3 s (`pkg/audit/audit.go:415-424`), CAP 220 ms via `metrics.CAPDeadlineMs` (`pkg/audit/report.go:352-355`). A library needs these as an explicit `Config`; today callers must pass zero `Deps` and accept all of them or fork `Run`.
7. **Loss denominator depends on a fake-friendly contract.** `expected = pingCalls×5` (`pkg/audit/audit.go:507`) assumes `Deps.Ping` returns exactly the requested `n`; injected fakes in tests return 5 (`pkg/audit/honesty_test.go:16`), so the contract is test-shaped, not interface-stated. A hostile or wrapped `Ping` breaks loss %. Fix: `Deps.Ping` returns `(samples, attempted, err)` or the engine counts from a probe-acknowledged attempts value.
8. **Throughput fallback reintroduces the invented duration the code warns against.** `throughputOf` falls back to `windowSecs` when `Elapsed<=0` (`pkg/audit/audit.go:815-824`) — exactly the “constant window invented a false throughput” failure the `LoadOutcome` comment warns about (`pkg/audit/audit.go:171-182`). The fallback must either be deleted or labelled `ThroughputEstimated:true`.
9. **Report mixes measurement, economics, and remediation in one `BuildReport`.** `BuildReport` computes `VoIPR` (`pkg/audit/report.go:317-319`), `Economics` (`pkg/audit/report.go:327-334`), `CAPSLA`/`OperatorVerdict` (`pkg/audit/report.go:352-366`), *and* generates `LinuxTC`/`RouterOS` scripts via `remediate` (`pkg/audit/report.go:431-500`). The `link remediate.Link` parameter (`pkg/audit/report.go:255`) drags the remediation vocabulary into the measurement type. A standalone audit should end at `Measured+Provenance+Invocation`; remediation is a separate module consuming `Measured`.
10. **CSV/artifact schema is frozen by fiat, not by version.** `AppendLinkAudit` migrates short headers itself (`pkg/audit/writer.go:24-43`) and depends on `model.LinkAuditHeader` (`pkg/audit/writer.go:19`); `link_audit.csv` schema is listed “Never touch” in `AGENTS.md`. Any standalone module needs its own versioned schema file + `Schema:"cgo.link-audit/1"` stamp; today there is no version string anywhere in `writer.go:13-84`.
11. **Minor correctness/robustness notes.** `timeline.points()` allocates `max+1` points even for sparse buckets (`pkg/audit/timeline.go:72-94`); `SealSource` reads the whole CSV into memory (`pkg/audit/report.go:509-519`); `kernelFacts` claims BTF present on `stat` alone (`pkg/probe/kernelsource.go:26-28`) and caps as `euid==0` (`pkg/probe/kernelsource.go:30`) — both over-claim vs (R9)/(R13); `BulkSendNFlows` third arg is ignored (`pkg/probe/bulkflows.go:23`); `Presets`/`Sinks`/`Remediation` French operator strings are returned values (same class as metrics §5.2). None blocks extraction alone; all belong in the target API hygiene pass.

## 6. Target module API

Proposed contract for two standalone modules (illustrative names; the shape is the deliverable). Per (R14) each lives behind its own `go.mod` + controlled path; pre-1.0 (`0.y.z`) it MAY break API freely, 1.0.0 is the compatibility promise.

```go
// ── Module probe: github.com/<org>/cgo-probe ─────────────────────
// No audit, no metrics, no remediate, no model. Stdlib + optional loader dep.
// Transport truth (what the kernel actually did):
type PingResult struct{ Attempts, Received int; Samples []PingSample } // Attempts in the type, loss is computable
func Ping(ctx context.Context, target string, count int, interval time.Duration, runner PingRunner) (PingResult, error)
// PingRunner is the exec seam (testable without a real `ping`); production impl
// shells `ping`, test impl returns canned output. Locale/pathology is in the runner,
// not in the API.

type SmallClient struct{ Timeout time.Duration; MarkEF bool } // MarkEF is a NAME, not silent Linux-only
func SmallObject(ctx context.Context, c *http.Client, url string) (time.Duration, error) // no builder hid in the name

type BulkOptions struct{ CC string; Flows int } // "" CC = host default (named), not cubic-by-global
func BulkUpload(ctx context.Context, addr string, opt BulkOptions) (BulkStats, error)
func BulkDownload(ctx context.Context, addr string, opt BulkOptions) (BulkStats, error) // "D:<cc>" hello internal
type BulkStats struct{ PerFlow []uint64; RetransBytes uint64; RetransMeasured bool; RetransProxy bool } // proxy labelled

// Socket truth (TCP, Linux-only where the kernel says so — R8):
func TCPRetransBytes(c net.Conn) (uint64, error) // ErrNotTCP / ErrNotLinux are the honest returns

// Kernel tier (R10 lifecycle): object + loader in this module.
type KernelTier interface{ Open() (closer io.Closer, err error); Summary() (bpf.Summary, error) }
func KernelSource() (KernelSourceInfo, error)   // still a ladder, but caps are real (capget), BTF is parsed (not stat-only)
func TCPInfoFallback() KernelTier               // today's `tcpinfo` rung, named as a tier

// DELETE: the `model.Cubic` default (`pkg/probe/bulkflows.go:44`); CC is always an explicit BulkOptions field.
// DELETE: `DialWithCC`/`SetConnCC` as public API — fold into BulkOptions; keep unexported helpers.
// KEEP (renamed): Hello parse/build, FlowResult (or fold into BulkStats.PerFlow).

// ── Module audit: github.com/<org>/cgo-audit ─────────────────────
// Depends on cgo-probe (types only) + an injected Stats interface (today's pkg/metrics).
// No remediate, no model, no profile, no hardcoded bench.
type Config struct {
  Duration time.Duration; Flows int; Direction Direction // Duration/Flows/Direction are values, not Run-normalised mysteries
  Target string; SmallURL string; BulkAddr string
  Ping PingFunc; Small SmallFunc; Load LoadFunc               // Deps, now in Config with documented signatures
  OnProgress func(Progress); Kernel KernelTier                // KernelTier nil = absent, never `tcpinfo` by default
}
type PingFunc func(ctx context.Context, target string, n int) (PingResult, error) // Attempts travel with samples (gap §5.7)
type Engine struct{ Stats Stats }   // Stats = Summarize/Grade/JFI/CAP/VoIPR surface currently in pkg/metrics
func Run(ctx context.Context, cfg Config, meta Meta) (*Result, error)  // Meta{Site,Provider,LinkType,AuditID,Origin,Sink}
// Result: headline + mirrors + Telemetry + Invocation, as today (pkg/audit/audit.go:41-144),
// minus throughput fallback (gap §5.8) — Throughput carries Estimated bool when window-fallback fires.
// Report: measured/ unavailable/ provenance/ invocation only. Economics & remediation are separate modules:
func BuildReport(res *Result, src SourceSeal) (*Report, error) // SourceSeal = the two SHA-256 seals; no remediate.Link param
// CSV: versioned schema
const LinkAuditSchema = "cgo.link-audit/1"
func AppendLinkAudit(w io.Writer, schemaVersion string, res *Result) error // caller owns the file; no model.LinkAuditHeader, no self-migration
```

Design notes: (a) probe module has zero `metrics`/`audit`/`remediate`/`model` imports — its `go list -m all` shows stdlib (+ loader) only; (b) audit module never imports `probe/bpf` directly (keeps `pkg/audit/kernel_adapter.go:3-5` decoupling via `KernelTier`); (c) every “absent” is a flag/error/`nil`, never a silent zero (honesty invariants preserved); (d) sinks/presets/tariffs/CC/DSCP-marking are explicit values, so the library is bench-agnostic and the bench is configuration, not code.

## 7. Slices (extraction order)

No slice touches code in this pass; each names its acceptance check.

1. **Slice 1 — Freeze + declare.** Snapshot current exported API (§2) into each module README as `0.1.0` declared public API; create the two `go.mod` files (R14); keep `LICENSE`; port all existing tests unchanged (probe: `probe_test.go`, `bulkflows_test.go`, `bulksource_test.go`, `bulkstats_test.go`, `kernelsource_test.go`, `bpf/*_test.go`; audit: `audit_test.go`, `honesty_test.go`, `report_test.go`, etc.). Acceptance: `go test ./...` green on each extracted tree; README lists every §2 identifier.
2. **Slice 2 — Probe de-globalise + honest sockets.** Delete `model.Cubic` default (`pkg/probe/bulkflows.go:44`) in favour of explicit `BulkOptions{CC}`; make DSCP marking a named option (`MarkEF`) instead of silent Linux-only (`pkg/probe/dscp_linux.go:12-30` vs `dscp_other.go:11-17`); make `PingRunner` the seam and carry `Attempts` in the result (gap §5.7); delete or label the throughput-window fallback (gap §5.8). Acceptance: `grep -rn "pkg/model" pkg/probe` empty; non-Linux `Ping`/`SmallObject`/`TCPRetransBytes` return named errors in tests; `-race` clean.
3. **Slice 3 — Kernel truth.** Build `bpf_probe.o` + `.sha256` on the pinned stack (`BUILD.md:16-41`); add a Go loader implementing Open/Load/Attach against (R10) + verifier-log capture per (R11); replace `kernelFacts` stat/euid heuristics (`pkg/probe/kernelsource.go:19-31`) with real BTF-parse + `capget`; wire `KernelTier` → `bpf.Summary` → `KernelSample`, read maps, attribute drops per (R12) tracepoints. Acceptance: on a 6.8 VM with BTF + CAP_BPF, `KernelSource()==ebpf` and a load produces non-zero `retrans_payload` matching `TCP_INFO` order-of-magnitude; without caps it falls to `tcpinfo` with `Reason` naming the missing fact. Until this lands, `ebpf` is never claimed in the README.
4. **Slice 4 — Audit de-bench + de-globalise.** Move hardcoded sinks/presets (`pkg/audit/sinks.go:37-48`, `pkg/audit/presets.go:18-39`) into example configuration (bench module or README); inject profile reference set instead of `model.Profiles`+`profile.Load()` in `MatchProfile` (`pkg/audit/match.go:20-22`); split `BuildReport` into `BuildReport` (measured/provenance only) + separate remediation/economics packages consuming `Measured` (gap §5.9); version the CSV schema (gap §5.10). Acceptance: `grep -rn "10\.30\|10\.200\|8\.8\.8\.8" pkg/audit` empty in non-test code; `go list -deps ./...` in audit shows no `model`/`profile`/`remediate`.
5. **Slice 5 — First tag.** Resolve both module paths (R14), tag `v0.1.0` for probe then audit, publish; in-repo `pkg/campagne`/`pkg/api`/`cmd/cgo` migrated to the new module versions (or dual-path with `replace` until migration lands, then deleted). Acceptance: `go list -m all` shows the new modules; in-repo `go test ./...` + `python tools/verify.py` green; no `replace` directives left.

Each slice is independently reviewable; Slices 2–4 change behaviour-visible semantics and each should be verifier-re-run before its `VERIFIED` row is claimed (repo builder/verifier split in `AGENTS.md` Roles).

## 8. Anti-toy criteria (what would make these toys — and the bar for “not a toy”)

A probe/audit library is a toy if it re-describes its own outputs instead of measuring against an external check. These extractions are **not** toys only if all of the following hold:

1. **Live-bench proof per probe family.** Ping/small/bulk run on the dual-namespace bench (`sinks.go` dualns `10.30.0.2` today) *and* on a real link with a `tc` impairment whose semantics are cited: netem delay ± jitter with known distribution (R6 §OPTIONS), TBF rate (R7), fq_codel `target/interval` (R5). A green run that never touched a bench or a shaping qdisc is a toy. The `data/runs` frozen corpus stays untouched — live proof is new runs, never rewritten history.
2. **Kernel-tier proof or honest fallback.** Either (a) `ebpf` attaches on a BTF+CAP_BPF host and per-flow retrans/drop counts are cross-checked against `TCP_INFO` + `tc -s qdisc show` drops (R3 §FORMAT `-s`), *or* (b) the module honestly reports `tcpinfo`/`unavailable` with the exact missing fact (object/BTF/caps) — as `KernelSource` does today (`pkg/probe/kernelsource.go:34-38`). A module that claims eBPF without ever reading a map is a toy. Until Slice 3 lands, (b) is the only honest state.
3. **Honesty invariants as tests, not comments.** The existing honesty suite (`pkg/audit/honesty_test.go:14-90`, `pkg/audit/audit_test.go` loss/CAP/JFI cases, `pkg/audit/report_test.go` unavailable cases) must pass in the extracted module *unmodified* first (Slice 1), then keep passing after Slices 2–4. A refactor that greens by deleting a honesty test is a toy-generator. Concretely: no-load ⇒ no grade (`pkg/audit/audit.go:562-565`), <2 flows ⇒ `JFIMeasured=false`, short audit ⇒ `DrainMeasured=false` with `Unavailable`, unmeasured CAP ⇒ blank CSV cols, empty ping ⇒ error.
4. **Provenance honesty.** Every constant carries source + date or is caller-supplied: bench IPs removed (gap §5.3), DSCP 46 cited to (R2) §2.7/§4, fq_codel defaults to (R5), CAKE ingress to (R4), TCP socket semantics to (R8), BTF path to (R9)/(R10). Waveform/Jain/G.107/RouterOS/`ping` output/`Total_retrans×1448` stay labelled unverified — primary source not found until fetched — the exact phrase required by the brief. A module presenting an operator quote or a bench IP as a universal constant is a toy.
5. **No silent zero, no self-migration.** Absent telemetry/probe is an absent row or a `Measured:false` flag, never a 0 ms ( `pkg/audit/telemetry.go:157-160`, `pkg/audit/report.go:36-38`); CSV schema is version-stamped and caller-owned, never self-migrating (`pkg/audit/writer.go:24-43` deleted). A library that fabricates a throughput duration or invents a replay command on decode is a toy.
6. **A real consumer + a real `tc` remediation check.** At least one out-of-tree importer (not only in-repo `cmd/cgo`/`pkg/campagne`) and one remediation path validated end-to-end: generated `tc` commands applied on a CHR/namespace bench and `CompareRouterOS` proving exports match intent (`pkg/audit/routeros_compare.go:22-64`) — RouterOS dialect itself stays unverified — primary source not found until a vendor doc is fetched. A module that only renders its own Markdown is a toy.
7. **Version discipline.** A declared API per module (R14 + README §1), tags `vX.Y.Z` treated as immutable, changelogs in MAJOR/MINOR/PATCH terms (as in the sibling metrics research `docs/research/2026-11-07-module-metrics.md` §8). A repo that edits a released tag, or ships no declared API, is a toy regardless of code quality.

---
*Teams note: slices 1–5 are proposals in this research file only. No `pkg/probe` or `pkg/audit` source was modified in this pass; implementation waits for human approval per `AGENTS.md` (“Stop and ask” — new dependency (eBPF loader), scope growth, frozen-data/schema touch).*
