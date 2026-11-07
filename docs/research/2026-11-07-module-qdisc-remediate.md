# pkg/qdisc + pkg/remediate as standalone tools — primary-source research

*Date: 2026-11-07 · Scope: `pkg/qdisc` + `pkg/remediate` only, read as if extracted tomorrow. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | RFC 8290 — FQ-CoDel (Experimental) | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched | FQ-CoDel/CoDel normative defaults: interval 100 ms (§5.2.1), target 5 ms (§5.2.2), packet limit MUST (§5.2.3), quantum 1514 (§5.2.4), flows 1024 (§5.2.5) |
| R2 | tc(8) — Linux traffic control | https://man7.org/linux/man-pages/man8/tc.8.html | fetched | shaping on egress, policing on ingress, `replace` = atomic remove/add, classless attach at root only, `-s` = more statistics, `ingress` = special qdisc that can be filtered/policed |
| R3 | tc-cake(8) | https://man7.org/linux/man-pages/man8/tc-cake.8.html | fetched | CAKE is shaping-capable (bandwidth, deficit shaper), `rtt` presets (internet = 100 ms default), `overhead` ∈ [-64,256], `ack-filter`/`nat`/`diffserv4`/`triple-isolate`/`dual-dsthost`/`ingress`/`wash` parameters |
| R4 | tc-fq_codel(8) | https://man7.org/linux/man-pages/man8/tc-fq_codel.8.html | fetched | fq_codel `limit` default 10240, `flows` 1024, `target` 5 ms, `interval` 100 ms, `quantum` 1514 defaults |
| R5 | tc-netem(8) | https://man7.org/linux/man-pages/man8/tc-netem.8.html | fetched | netem `delay TIME [JITTER]`, `loss gemodel p [r [1-h [1-k]]]`, `distribution {uniform\|normal\|pareto\|paretonormal}`; Gilbert-Elliott implementers paper named in SOURCES |
| R6 | tc-tbf(8) | https://man7.org/linux/man-pages/man8/tc-tbf.8.html | fetched | TBF is a pure, non-work-conserving shaper; `burst` = bucket size; `limit`/`latency` mutually exclusive; inner qdisc attached at `parent 10:1` |
| R7 | tc-pfifo_fast(8) | https://man7.org/linux/man-pages/man8/tc-pfifo_fast.8.html | fetched | pfifo_fast is the automatic default, 3-band TOS queue, "does not maintain statistics and does not show up in tc qdisc ls" |
| R8 | Go Modules Reference (`go.mod` file reference) | https://go.dev/doc/modules/gomod-ref | fetched | module path + version = unique ID and import prefix; `go` directive = minimum Go version; v2+ module path must end `/v2` |
| R9 | Go Module version numbering | https://go.dev/doc/modules/version-numbers | fetched | v0 = unstable, minor = backward-compatible API change, patch = no public API change, major = backward-incompatible, major >1 requires new module path suffix |
| R10 | Semantic Versioning 2.0.0 | https://semver.org/ | fetched | MAJOR/MINOR/PATCH increment rules (§6–§8) and "MUST declare a public API" (§1) |
| — | MikroTik RouterOS `/queue type`, `/queue tree`, `kind=cake`, `cake-bandwidth`, `cake-ack-filter`, RouterOS ≥ 7.10 | help.mikrotik.com | **404** — `https://help.mikrotik.com/docs/display/ROS/Queue`, `.../Queue+Tree`, `.../search?q=queue+tree+cake` all returned 404 | unverified — primary source not found (all RouterOS syntax claims below) |
| — | 95 % shaper headroom (`shaperHeadroom = 0.95`) | — | unverified — primary source not found (only in-repo rationale `pkg/remediate/remediate.go:70-74`) | `ShapedRate` calibration |
| — | `Recommend` thresholds (asymmetric → CAKE, symmetric+CPU → fq_codel, else CAKE) | — | unverified — primary source not found (only in-repo doc `pkg/remediate/remediate.go:393-399`) | decision policy |
| — | Custom named netem distribution file (`.dist` by profile name) | — | unverified — primary source not found (fetched tc-netem(8) lists only `uniform\|normal\|pareto\|paretonormal` for `delay`; a `FILE` form appears only under `slot distribution`) | `ApplyNetemDistribution` argument |
| — | netem `.dist` table encoding (base-0 ints, `mu + (sigma/8192)·t`, `NETEM_DIST_MAX` bounds one entry not the 16384 table) | — | unverified — primary source not found (only in-repo comment `pkg/qdisc/dist.go:20-22`; kernel `q_netem.c` not fetched) | `Calibration` constants |
| — | Gilbert-Elliott mean-loss formula `π = p/(p+r)` | — | unverified — primary source not found (paper named in tc-netem(8) SOURCES but not fetched; only in-repo formula `pkg/qdisc/ge.go:26-37`) | `GEMeanLoss` |
| — | `tc filter ... parent ffff:` ingress handle convention | — | unverified — primary source not found (fetched tc(8) names `ingress` a special qdisc but does not state the `ffff:` handle) | `CheckIngress` |

Method: read `pkg/qdisc/*.go`, `pkg/remediate/*.go`, `docs/api.md` remediation endpoints, `pkg/model/types.go`, `go.mod`, and callers in `pkg/campagne`, `pkg/audit`, `pkg/api`. Every technical claim below cites `file:line` or `(R#)`; anything else is labelled `unverified — primary source not found`.

## 1. Responsibility

### 1.1 `pkg/qdisc` — bench-side `tc` execution behind a runner seam

`pkg/qdisc/qdisc.go:1-13` states the contract: "`tc` configuration behind a TCRunner seam", host tests use `FakeRunner`, the VM uses `ExecRunner`, and the topology is two hops (`veth client ← netem · veth shaping+aqm → veth server`) so two classless roots never contend on one interface (R2: classless qdiscs attach only at root). Concretely:

- **Runner seam** — `TCRunner{Run(args...)}` (`pkg/qdisc/qdisc.go:23-25`); `ExecRunner` runs `sudo tc …` (`pkg/qdisc/qdisc.go:27-32`); `NsRunner{Ns, Env}` runs `sudo env … ip netns exec NS tc …` (`pkg/qdisc/qdisc.go:37-54`); `FakeRunner` records and always succeeds (`pkg/qdisc/qdisc.go:57-62`). A second seam, `IngressRunner{Run(prog, args...)}` + `NsCommand` (`pkg/qdisc/ingress.go:13-23`), exists because the ingress preflight needs `ip` as well as `tc`.
- **Egress shaper+AQM** — `ApplyShaper` (`pkg/qdisc/qdisc.go:88-143`): CAKE carries its own bandwidth (R3) so it is the qdisc (child `10:` of netem `1:`, falling back to root `1:`); `fq_codel` has no shaper of its own so it is attached as child `20:` of a `tbf` (`10:`) with fq_codel pinned to `target 5ms interval 100ms limit 10240 quantum 1514 flows 1024` (`pkg/qdisc/qdisc.go:126`); the default (pfifo_fast) cell is TBF alone (R7: pfifo_fast is the magic default and cannot be attached as a child, per `pkg/qdisc/qdisc.go:10-12`). All three are `qdisc replace` (R2: nearly atomic remove/add, creating the node if absent).
- **Impairment hop** — `ApplyNetem`/`ApplyNetemPhysics`/`ApplyNetemGE`/`ApplyNetemDistribution` (`pkg/qdisc/qdisc.go:79-81`, `pkg/qdisc/ge.go:41-92`): root `netem` at handle `1:` with `delay+jitter` (R5), uniform `loss N%`, or 4-argument `loss gemodel p r 1-h 1-k` (R5), or a named `distribution`; jitter ≤ 0 with a distribution is refused before `tc` (`pkg/qdisc/ge.go:66-68`).
- **Ingress shaper** — `ApplyIngressShaper` (`pkg/qdisc/ingress.go:28-56`): CAKE root with `ingress` (R3: ingress mode counts dropped packets as already transferred), fq_codel/pfifo as `tbf` root + child; `CheckIngress` (`pkg/qdisc/ingress.go:61-88`) is read-only and runs before any mutation (IFB up, ingress qdisc present on the WAN, `mirred` filter to the IFB) — matching R2's "ingress … allowing for it to be filtered and policed" and the code comment that an ingress qdisc cannot itself build an AQM queue (`pkg/qdisc/ingress.go:25-27`).
- **Stats** — `PollStats` (`pkg/qdisc/stats.go:22-28`) runs `tc -s qdisc show dev` (R2: `-s` = more statistics) and parses `Sent`/`backlog` lines into `Stats{Kind,Handle,Bytes,Packets,Drops,Overlimits,Backlog}` (`pkg/qdisc/stats.go:10-18`, `36-66`); `SumDrops` sums all layers, `SumBytes` returns only the leaf (`pkg/qdisc/stats.go:86-105`, see §5).
- **Measured calibration** — `BuildCalibration`/`WriteCalibration`/`LoadCalibration`/`Render` (`pkg/qdisc/dist.go:73-303`): RTT baseline → 1024-entry netem table, sealed by a JSON sidecar with `sha256` of the `.dist` bytes; load is fail-closed (missing/divergent sidecar or `.dist` refuses before the first qdisc mutation, `pkg/qdisc/dist.go:261-263`).

What it deliberately does **not** do: it generates no production remediation (that is `pkg/remediate`), measures no traffic itself (that is `pkg/probe`, per `docs/architecture.md:14-16`), and knows no HTTP/CSV/report formats. It executes `tc` on a bench the caller owns.

### 1.2 `pkg/remediate` — measured link → production queueing config

`pkg/remediate/remediate.go:1-8` states the contract: "a remediation is only emitted from numbers that were actually measured … a remediation generated from a guessed bandwidth is worse than no remediation". Concretely:

- **Refusal of unmeasured input** — `Link{Interface, RouterOSInterface, RouterOSLanInterface, CapacityMbps, RTTms, Overhead, MTU}` (`pkg/remediate/remediate.go:22-39`); `Validate` rejects empty interface, `CapacityMbps ≤ 0`/non-finite, `RTTms ≤ 0`, negative overhead with `ErrUnmeasured` (`pkg/remediate/remediate.go:77-101`). Rate actually written is `CapacityMbps × 0.95` (`pkg/remediate/remediate.go:68-74`, `120-121`) — the 95 % figure is unverified — primary source not found.
- **Four generators** — `LinuxTC` egress (`pkg/remediate/remediate.go:135-194`), `LinuxTCIngress` download via IFB (`215-260`), `RouterOS` upload script (`268-322`), `RouterOSIngress` download script (`330-384`). Each returns ordered shell commands / a script string; all begin from a clean root (`tc qdisc del … || true`, `pkg/remediate/remediate.go:145`) because a stale qdisc is "the most common cause of the fix did nothing" (`142-144`). CAKE is root with `bandwidth`+`rtt`+`ack-filter`+`nat`+`diffserv4`+`triple-isolate` (R3 lists all of these parameters); ingress CAKE uses `ingress`+`dual-dsthost` and deliberately drops `ack-filter` (`226-239`). fq_codel is `tbf` root (rate = shaped rate, `latency 400ms`) with `fq_codel` child (`168-178`, `241-247`). RouterOS scripts are idempotent (remove `cgo-managed` objects first) and explicitly state fq_codel is emitted as `sfq` "NOT fq_codel" (`300-310`); all RouterOS syntax is unverified — primary source not found.
- **Decision** — `Recommend` (`pkg/remediate/remediate.go:400-417`) returns one-line `{Qdisc, Reason}`: asymmetric → CAKE, symmetric+cpuConstrained → fq_codel, else CAKE. Thresholds are unverified — primary source not found.

What it deliberately does **not** do: it never executes anything (no `exec`, no `os/exec` import — generators return strings), never measures, and never shapes the bench. It is a pure string generator from a validated struct, which is exactly what makes it the easiest of the two to extract.

### 1.3 API surface (`docs/api.md` remediation endpoints)

- `GET /api/remediate?audit_id=&iface=&routeros_iface=&routeros_lan=&qdisc=&asymmetric=&cpu_constrained=` → `{audit_id, remediation, unavailable}`; `404` when no audit report exists (`docs/api.md:101`). Implementation: `handleRemediate` (`pkg/api/server_audit.go:323-363`) returns the sealed report's remediation verbatim, or rebuilds via `audit.BuildReport` from the in-memory measured result when `iface` is supplied; `asymmetric` is false only on `=false` (`pkg/api/server_audit.go:354`). `POST /api/audit/start` accepts the same shaping hints (`routeros_iface`, `routeros_lan`, `qdisc`, `asymmetric`, `cpu_constrained`) and seals them into the report (`pkg/api/server_audit.go:187-192`).
- Related but distinct: `POST /api/shape` applies a bench shaper via `pkg/qdisc` (`docs/api.md:51`, bounds capacity 1–1000 Mbit/s, delay 0–600 ms, jitter 0–100 ms, loss 0–10 %); `GET /api/hardware/translate` recommends a qdisc from measured `Best` else `cake` for P1 / `fq_codel` (`docs/api.md:110`); `GET /api/report/export?format=sh` emits a replayable `tc` script derived from the frozen best cell (`docs/api.md:81`). The remediation endpoint is the only one that calls `pkg/remediate` generators, through `pkg/audit/report.go:431-500` (`remediate.Recommend`, `LinuxTC`, `RouterOS`, `LinuxTCIngress`, `RouterOSIngress`, `ShapedRate`).

## 2. Exported API (file:line)

Complete inventory of exported identifiers in non-test sources. Unexported helpers are listed separately and are not API.

### 2.1 `pkg/qdisc`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `TCRunner` | interface | `pkg/qdisc/qdisc.go:23-25` | `Run(args ...string) ([]byte, error)` — the seam |
| `ExecRunner` | struct | `pkg/qdisc/qdisc.go:27-32` | `sudo tc …` |
| `NsRunner{Ns, Env}` | struct | `pkg/qdisc/qdisc.go:37-44` | `sudo env … ip netns exec … tc` |
| `FakeRunner{Calls}` | struct | `pkg/qdisc/qdisc.go:57-62` | records, always succeeds |
| `Reset(r, iface)` | func | `pkg/qdisc/qdisc.go:67-74` | `qdisc del dev root`; tolerates "Cannot find specified qdisc"/"No such file" |
| `ApplyNetem(r, iface, delayMs, jitterMs, lossPct)` | func | `pkg/qdisc/qdisc.go:79-81` | uniform loss, historical byte-exact corpus form |
| `ApplyShaper(r, iface, q model.Qdisc, capMbps, rttMs)` | func | `pkg/qdisc/qdisc.go:88-143` | CAKE / fq_codel-under-tbf / pfifo=tbf; parent-`1:` then root fallback, tc output kept in error |
| `GEModel{GoodToBad, BadToGood, LossBad, LossGood}` | struct | `pkg/qdisc/ge.go:19-24` | 4-arg gemodel (R5) |
| `GEMeanLoss(ge)` | func | `pkg/qdisc/ge.go:30-37` | mean loss %; formula unverified — primary source not found |
| `ApplyNetemGE(r, iface, delayMs, jitterMs, ge)` | func | `pkg/qdisc/ge.go:41-46` | |
| `ApplyNetemPhysics(r, iface, delayMs, jitterMs, lossPct, ge *GEModel)` | func | `pkg/qdisc/ge.go:52-57` | uniform XOR GE, never both |
| `ApplyNetemDistribution(r, iface, delayMs, jitterMs, lossPct, ge, dist)` | func | `pkg/qdisc/ge.go:62-73` | refuses empty dist / `jitterMs ≤ 0` before `tc` |
| `Calibration` | struct | `pkg/qdisc/dist.go:33-67` | `Table []int16` is `json:"-"` (`.dist` carries it) |
| `BuildCalibration(prof model.Profile, auditID, rtts)` | func | `pkg/qdisc/dist.go:73-149` | refuses `<30` samples, `delay ≤ 0`, p50 ≤ floor, zero residual p999 |
| `Render()` | method | `pkg/qdisc/dist.go:187-194` | one decimal int per line, no padding/comments |
| `WriteCalibration(dir, cal)` | func | `pkg/qdisc/dist.go:199-240` | atomic `.dist` + sidecar, seals sha256 |
| `LoadCalibration(dir, profile)` | func | `pkg/qdisc/dist.go:264-303` | fail-closed, hash + line-count verified |
| `Stats{Kind, Handle, Bytes, Packets, Drops, Overlimits, Backlog}` | struct | `pkg/qdisc/stats.go:10-18` | cumulative counters |
| `PollStats(r, iface)` | func | `pkg/qdisc/stats.go:22-28` | `tc -s qdisc show dev` (R2 `-s`) |
| `SumDrops(stats)` | func | `pkg/qdisc/stats.go:86-92` | sums drops over all qdiscs |
| `SumBytes(stats)` | func | `pkg/qdisc/stats.go:99-105` | leaf (last) qdisc bytes only |
| `IngressRunner` | interface | `pkg/qdisc/ingress.go:13-15` | `Run(prog, args...)` |
| `NsCommand{Ns}` | struct | `pkg/qdisc/ingress.go:18-23` | `sudo ip netns exec …` |
| `ApplyIngressShaper(r, iface, q model.Qdisc, capMbps, rttMs)` | func | `pkg/qdisc/ingress.go:28-56` | CAKE `ingress` / tbf+fq_codel / tbf at root |
| `CheckIngress(r, ingressIf, ifb)` | func | `pkg/qdisc/ingress.go:61-88` | read-only preflight, never mutates |

Unexported: `nsTCArgs` (`qdisc.go:46`), `mbps` (`qdisc.go:64`), `fqCoDelString` (`qdisc.go:145`), `oneLine` (`qdisc.go:149`), `applyNetemArgs` (`ge.go:75`), `quantile`/`recomputeRoundTrip`/`writeFileAtomic`/`safeCalibrationName` (`dist.go:151-259`), `parseStats`/`parseU64` (`stats.go:36-83`), `execCmdRunner` (`stats.go:108-112`, test helper in a non-test file), `linkUp` (`ingress.go:92-103`).

### 2.2 `pkg/remediate`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Link{Interface, RouterOSInterface, RouterOSLanInterface, CapacityMbps, RTTms, Overhead, MTU}` | struct | `pkg/remediate/remediate.go:22-39` | measured shape; `MTU=0` ⇒ 1500 (`103-108`) |
| `QdiscCake`/`QdiscFqCodel`/`QdiscPfifo` | consts | `pkg/remediate/remediate.go:42-46` | `"cake"`/`"fq_codel"`/`"pfifo_fast"` — duplicates `model.Qdisc` (`pkg/model/types.go:72-76`) as plain strings |
| `IngressIFB` | const | `pkg/remediate/remediate.go:50` | `"ifb0"` |
| `DefaultRouterOSLan` | const | `pkg/remediate/remediate.go:56` | `"bridge"` — assumed, labelled in script |
| `RouterOSLan()` | method | `pkg/remediate/remediate.go:59-64` | returns `(iface, assumed)` |
| `ShapedRate(measuredMbps)` | func | `pkg/remediate/remediate.go:68` | `× 0.95` (unverified — primary source not found) |
| `ErrUnmeasured{Field, Why}` | type | `pkg/remediate/remediate.go:77-81` | error for unmeasured input |
| `Validate()` | method | `pkg/remediate/remediate.go:84-101` | refuses zero/negative/non-finite capacity, `RTT ≤ 0`, negative overhead |
| `LinuxTC(l, qdisc)` | func | `pkg/remediate/remediate.go:135-194` | egress command set |
| `LinuxTCIngress(l, qdisc)` | func | `pkg/remediate/remediate.go:215-260` | download command set via IFB |
| `RouterOS(l, qdisc)` | func | `pkg/remediate/remediate.go:268-322` | upload script; RouterOS syntax unverified — primary source not found |
| `RouterOSIngress(l, qdisc)` | func | `pkg/remediate/remediate.go:330-384` | download script; assumed-LAN warning |
| `Recommendation{Qdisc, Reason}` | struct | `pkg/remediate/remediate.go:388-391` | |
| `Recommend(l, asymmetric, cpuConstrained)` | func | `pkg/remediate/remediate.go:400-417` | policy unverified — primary source not found |

Unexported: `shaperHeadroom` (`remediate.go:74`), `mtu` (`103`), `burstBytes` (`112`), `shapedRate` (`121`), `rate` (`124`), `ingressSetup` (`199`).

## 3. Dependencies

### 3.1 Module identity and Go version

`go.mod:1-3` declares `module github.com/Realms4239/cgo` with `go 1.25`. Per (R8) the module path is the unique identifier combined with the version number and the import prefix of every package; per (R8) the `go` directive is the minimum Go version a consumer needs. Per (R9) a standalone re-issue must choose a version whose number signals the API change (v0 = unstable; minor = backward-compatible addition; major = breaking, and major > 1 requires a `/vN` path suffix per (R8)); per (R10) §1 the module MUST declare its public API, which §2 of this document inventories.

### 3.2 Internal dependency graph

| Package | Imports (internal) | Imports (external) | Imported by |
|---|---|---|---|
| `pkg/qdisc` | `pkg/model` only (`Qdisc` in `qdisc.go:20`/`ingress.go:8`, `Profile` in `dist.go:17`) | stdlib only (`os/exec`, `fmt`, `strings`, …) | `pkg/campagne` (`campagne.go:360-368,391,415,428,495,596-597`, `dualns.go:46-173`, `prod.go:41-80`, `watch.go:34,77`), `pkg/api` doc reference only (`server.go:1048`) |
| `pkg/remediate` | **none** — `fmt`, `math`, `strings` only (`remediate.go:11-15`); the `model.Qdisc` values are re-declared as string consts (`remediate.go:42-46`) | stdlib only | `pkg/audit/report.go:431-500` (`Recommend`, `ShapedRate`, `LinuxTC`, `RouterOS`, `LinuxTCIngress`, `RouterOSIngress`), `pkg/api/server_audit.go:187,349-355` (constructs `Link`), `pkg/audit/*_test.go` fixtures |

`pkg/model/types.go:70-78` defines `type Qdisc string` with `PfifoFast/FqCodel/Cake` and `AllQdiscs`; `pkg/model/types.go:11-28` defines `Profile{ID, CapacityMbps, CapacityUpMbps, DelayMs, JitterMs, LossPct, …}` with the deprecated `LossBurst*` fields kept only for the CSV import format (`pkg/model/types.go:18-23`). Consequences for extraction: `pkg/remediate` is already dependency-free and extracts as-is; `pkg/qdisc` drags `pkg/model` for exactly two types (`Qdisc`, `Profile`) and must either take them or redefine them (see §7, slice S1).

### 3.3 External / runtime dependencies (not Go modules)

- `tc` binary and the `cake`, `fq_codel`, `tbf`, `netem`, `pfifo` kernel qdiscs (R2–R7), invoked via `exec.Command("sudo", "tc", …)` (`qdisc.go:31`) and `ip netns exec` (`qdisc.go:43`, `ingress.go:22`); requires `CAP_NET_ADMIN`-equivalent privilege, exercised by `ExecRunner` on the VM only.
- IFB device + `act_mirred`/`cls_matchall` modules for the ingress path, created idempotently by `ingressSetup` (`remediate.go:199-209`) and preflighted by `CheckIngress` (`ingress.go:61-88`).
- RouterOS ≥ 7.10 per the generated script header (`remediate.go:281`) — unverified — primary source not found.

## 4. Primary-source requirements

For each behaviour, the primary source a standalone module must satisfy (or explicitly refuse):

| Behaviour | Requirement (primary source) | Where it lands |
|---|---|---|
| fq_codel parameters | defaults `target 5 ms`, `interval 100 ms`, `limit 10240`, `quantum 1514`, `flows 1024`; packet limit MUST be enforced (R1 §5.2.1–§5.2.3, R4) | bench `ApplyShaper` pins all five (`qdisc.go:126`); production `LinuxTC` emits bare `fq_codel` (R4 defaults) — see gap G1 |
| CAKE shape + classify + ingress | CAKE is shaping-capable with `bandwidth`, `rtt` (internet default = 100 ms), `overhead` ∈ [−64, 256], `ingress` counts dropped bytes, `diffserv4`/`triple-isolate`/`dual-dsthost`/`nat`/`ack-filter`/`wash` are the available knobs (R3) | bench sets only `bandwidth`+`rtt` (`qdisc.go:89-96`); production sets the full knobs (`remediate.go:154-166`, `227-239`) — see gap G2 |
| `tc` mechanics | shaping on egress only, policing on ingress; classless qdiscs attach at root; `replace` is remove/add; a child qdisc attaches under a parent class `1:1`/`10:1` (R2, R6 inner-qdisc example) | handle topology `1:`/`10:`/`20:` (`qdisc.go:88-143`, `ingress.go:28-56`) |
| netem | `delay`+`jitter`, `loss gemodel p r 1-h 1-k` 4-arg, `distribution` named set (R5) | `ge.go:75-92`; custom `.dist` name unverified — primary source not found |
| TBF | pure non-work-conserving shaper; `burst` bucket; `limit`/`latency` mutually exclusive; too-small buffer drops (R6) | `rate … burst 256kbit latency 400ms` (`qdisc.go:97-104`) vs `burst <MTU bytes> latency 400ms` (`remediate.go:172-178`) — see gap G3 |
| pfifo_fast baseline | 3-band TOS default, magic, no stats, invisible in `tc qdisc ls` (R7) | bench "pfifo_fast cell" is TBF+default fifo (`qdisc.go:133-142`); production emits `pfifo limit 1000` child (`remediate.go:188`) — see gap G4 |
| Stats | `tc -s` shows more statistics (R2 FORMAT `-s`); `-j` JSON is also available (R2 FORMAT `-j`) | text parser `stats.go:36-66` ignores `-j` — see gap G5 |
| Ingress queueing | ingress qdisc can filter/police but cannot queue; shaping occurs on egress (R2 DESCRIPTION, `ingress.go:25-27`) | IFB redirect plumbing (`ingress.go:28-56`, `remediate.go:199-260`) |
| Go module + SemVer | module path = ID + import prefix, `go` = minimum version (R8); v0 unstable, minor/patch/major signal API change (R9); public API MUST be declared (R10 §1) | §7 target API + §8 slices, versioned per (R9)/(R10) |
| RouterOS | any `/queue type`/`/queue tree`/`kind=cake` syntax | unverified — primary source not found (404, §0) — see gap G6 |

## 5. Gap analysis

Each gap is: where, what breaks, primary-source consequence.

- **G1 — fq_codel params pinned in the bench, defaulted in production.** Bench emits the RFC 8290 set verbatim (`qdisc.go:126`, matching R1 §5.2/R4), production emits bare `fq_codel` (`remediate.go:177`, `246`, `ingress.go:43`) and relies on kernel defaults. R4 confirms the defaults equal the RFC values, so today they agree — but the agreement is implicit: a kernel changing a default silently drifts production away from the bench cell, and nothing in the emitted production command records which parameters were intended. Extract as one shared `FqCodelParams` value.
- **G2 — CAKE knobs differ bench vs production.** Bench CAKE sets `bandwidth`+`rtt` only (`qdisc.go:89-96`); production CAKE adds `ack-filter`, `nat`, `diffserv4`, `triple-isolate` (egress) or `dual-dsthost` (ingress) (`remediate.go:154-166`). Per R3, `diffserv4` vs the default `diffserv3` changes the tin set (4 tins vs 3), and `ack-filter`/`nat` change classification — so the bench cell and the production recipe are not the same qdisc. A module must decide whether it measures the bench cell or the production recipe, or expose both explicitly.
- **G3 — TBF burst sized two different ways.** Bench uses a fixed `burst 256kbit` (`qdisc.go:99`); production uses `burst = max(MTU, 1500)` bytes (`remediate.go:112-118`, `172-176`) with a comment that one-MTU is the minimum that does not throttle a full segment. R6 warns larger rates need larger buffers and a too-small buffer drops — at 1000 Mbit/s one MTU is far below `rate/HZ`, so production TBF may under-deliver at high rates while the bench's fixed 256 kbit is rate-blind at low rates. Neither sizes from the rate; both use `latency 400ms` (R6: `latency` is derived from bucket+rate, so the two knobs interact).
- **G4 — "pfifo_fast" means three different things.** Bench cell = `tbf` + kernel default fifo (`qdisc.go:133-142`, R7 magic default); production `LinuxTC` = `tbf` + `pfifo limit 1000` child (`remediate.go:186-189`); API accepts `pfifo_fast` as a qdisc name (`docs/api.md:51`). R7 says pfifo_fast is 3-band TOS with txqueuelen-sized bands and invisible in `tc qdisc ls` — none of the three implementations is that. The label is a local convention; a standalone module must rename it (e.g. `tbf-fifo`) or implement actual pfifo_fast.
- **G5 — stats parsing is text scraping with a fragile leaf heuristic.** `parseStats` matches `Sent `/`backlog ` line prefixes (`stats.go:36-66`); R2 offers `-j` JSON which is ignored. `SumBytes` returns the last qdisc's bytes and `SumDrops` sums all layers (`stats.go:86-105`), with a comment admitting the summing would inflate goodput 2–3× (P1 80 Mbit observed 231) and that the leaf heuristic fails on "a qdisc exotique". R7 compounds this: a real pfifo_fast default never appears in `tc qdisc ls`, so a pfifo_fast cell yields zero rows and both sums are 0 — indistinguishable from "no traffic". Extract as: parse `-j`, or document the exact `tc -s` line grammar as the module's contract and add a VM golden test.
- **G6 — RouterOS output is unverified end-to-end.** All of `kind=cake`, `cake-bandwidth`, `cake-rtt`, `cake-flowmode`, `cake-ack-filter=filter`, `sfq-perturb`, `/queue tree … max-limit`, the ≥ 7.10 gate, and the `sfq`-for-fq_codel substitution (`remediate.go:268-384`) rest on in-repo comments only; help.mikrotik.com returned 404 for every attempted page (§0), so every RouterOS claim is `unverified — primary source not found`. Shipping this as a standalone tool means shipping untested config for a device the maintainer may not own. Either verify against a real RouterOS or mark the generators experimental.
- **G7 — Calibration `.dist` is loadable but its format is unverified.** `WriteCalibration`/`LoadCalibration` seal sha256 + a newline count (`dist.go:228-303`), so corruption is caught; but a sidecar whose table values are out of range, or a `.dist` rendered from a table mutated after sealing, is caught only by the hash, while the external consumer of the file is `tc-netem` whose custom-distribution form is unverified — primary source not found (§4). `LoadCalibration` also counts lines rather than parsing values (`dist.go:299-301`), so it cannot detect an empty line pair that preserves the count but changes content — the hash covers this, which makes the count redundant rather than load-bearing.
- **G8 — Dead code and duplicated constants at the seam.** `fqCoDelString` (`qdisc.go:145-147`) and the `cakeArgs`/`tbfArgs` closures (`qdisc.go:89-104`) are constructed and discarded (`_ = cakeArgs`, `_ = tbfArgs`); `execCmdRunner` (`stats.go:108-112`) is a test helper shipped in a non-test file; the fq_codel string is duplicated between `qdisc.go:126` and `fqCoDelString`; `remediate` re-declares `model.Qdisc` values as strings (`remediate.go:42-46`) instead of importing `model`. None of these is a bug today; all of them become drift the moment one copy is edited.
- **G9 — `Reset` error matching is string matching.** `Reset` tolerates only two English `tc` error substrings (`qdisc.go:69-70`); a localised or reworded `tc` error becomes a hard failure. R2 does not specify `tc`'s error text, so this tolerance is unverified — primary source not found.
- **G10 — Ingress preflight checks names, not effect.** `CheckIngress` verifies the ingress qdisc exists and a `mirred` filter mentions the IFB (`ingress.go:72-87`), but never verifies traffic actually arrives at the IFB or that the IFB qdisc is the one shaping it; combined with G5, a mis-wired ingress path reports zero stats rather than an error. The fail-closed claim (`ingress.go:58-60`) covers absence of plumbing, not a silent no-op.

## 6. Target module API

Two modules, one per package, because their owners, executions, and test gates differ (bench execution vs pure generation). Versioned per (R9)/(R10): start `v0.1.0` (unstable, API may change), bump minor only for backward-compatible additions, major (with `/v2` path per (R8)) only for breaking changes.

### 6.1 `qdisc` module — `module github.com/Realms4239/cgo/qdisc`

Replaces `model.Qdisc` with a local `Kind` (breaks the `pkg/model` coupling, §3.2); all parameters become an explicit struct so G1–G4 drift is a type error, not a comment.

```go
package qdisc

type Kind string
const ( KindCake Kind = "cake"; KindFqCodel Kind = "fq_codel"; KindTBF Kind = "tbf-fifo" ) // G4: no pfifo_fast label

type TCRunner interface{ Run(args ...string) (stdout []byte, err error) }
type ExecRunner struct{}
type NsRunner struct{ Ns string; Env []string }
type FakeRunner struct{ Calls [][]string }

type ShaperParams struct {
    Kind        Kind
    CapacityMbps float64 // must be > 0, finite
    RTTms       float64 // CAKE rtt; must be > 0
    CAKE        CAKEParams   // FlowMode, Diffserv, AckFilter, NAT, Overhead  (R3)
    FqCodel     FqCodelParams // Target, Interval, Limit, Quantum, Flows — default = R1 §5.2 values
    TBF         TBFParams    // BurstBytes derived from rate (R6), LatencyMs
}
func Reset(r TCRunner, iface string) error
func ApplyShaper(r TCRunner, iface string, p ShaperParams) error
func ApplyIngressShaper(r TCRunner, iface string, p ShaperParams) error
func CheckIngress(r IngressRunner, ingressIf, ifb string) error
func ApplyNetem(r TCRunner, iface string, p NetemParams) error // Delay, Jitter, LossPct XOR GE, Distribution
func PollStats(r TCRunner, iface string) ([]Stats, error)       // parses tc -j (R2), text parser kept behind a flag
func SumLeafBytes(stats []Stats) uint64                          // renamed: it is not a sum
type Stats struct{ Kind, Handle string; Bytes, Packets, Drops, Overlimits, Backlog uint64 }
```

`Calibration` moves unchanged except: `.dist` format constant names the kernel source it targets, `LoadCalibration` parses and range-checks values (not only counts lines, G7), and the custom-distribution argument is documented as unverified — primary source not found until a `tc-netem` source confirms it.

### 6.2 `remediate` module — `module github.com/Realms4239/cgo/remediate`

No internal dependencies today (§3.2); kept that way. The only change: generators return a typed command, not a string, so shell-injection via an interface name is a construction error, and RouterOS is gated behind a build tag / explicit `Experimental` marker (G6).

```go
package remediate

type Link struct{ Interface, RouterOSInterface, RouterOSLanInterface string; CapacityMbps, RTTms float64; Overhead, MTU int }
type ErrUnmeasured struct{ Field, Why string }
func (l Link) Validate() error
func ShapedRate(measuredMbps float64) float64 // 0.95 documented as unverified — primary source not found

type Command struct{ argv []string } // not a shell string
func LinuxTC(l Link, qdisc string) ([]Command, error)
func LinuxTCIngress(l Link, qdisc string) ([]Command, error)
func RouterOS(l Link, qdisc string) (string, error)        // marked experimental until RouterOS is verified
func RouterOSIngress(l Link, qdisc string) (string, error)

type Recommendation struct{ Qdisc, Reason string }
type Policy struct{ Asymmetric, CPUConstrained bool } // thresholds documented as heuristic, unverified — primary source not found
func Recommend(l Link, p Policy) Recommendation
```

### 6.3 What does not move

`Calibration`'s sidecar schema stays versioned by the existing `schema:1` field (`dist.go:34`); `pkg/audit`'s `BuildReport` keeps owning the HTTP/CSV/report assembly (`docs/api.md:101`); `pkg/campagne` keeps owning bench orchestration. Extraction is the two packages plus their tests, not the callers.

## 7. Slices

Each slice is independently reviewable and leaves the tree green; ordered so the dependency break comes first.

| Slice | Scope | Gate |
|---|---|---|
| S1 | Decouple `pkg/qdisc` from `pkg/model`: local `Kind` + local `Profile` fields used by `BuildCalibration`; `remediate` already decoupled (§3.2) | `go test ./pkg/qdisc/...` unchanged; `grep import.*pkg/model pkg/qdisc` empty |
| S2 | One `FqCodelParams` value shared by `ApplyShaper`, `ApplyIngressShaper`, `LinuxTC`, `LinuxTCIngress` (G1); delete `fqCoDelString` dead code (G8) | arg-vector tests assert the five R1 parameters |
| S3 | `PollStats` gains a `tc -j` parser; text parser kept only as fallback; `SumBytes` renamed and documented as leaf-only (G5) | golden `tc -j` fixtures; pfifo_fast-invisible case (R7) returns an explicit "no qdisc rows" error, not 0 |
| S4 | TBF burst derived from rate for both packages; single sizing rule (R6) (G3) | unit tests at 5/20/80/1000 Mbit/s assert `burst ≥ rate/HZ` |
| S5 | CAKE knobs unified: bench and production set the same `CAKEParams` (R3) or the difference is named (G2) | arg-vector diff between `ApplyShaper` CAKE and `LinuxTC` CAKE is empty or explicitly listed |
| S6 | RouterOS: verify against a real RouterOS ≥ 7.10 device, or fence behind `Experimental` and remove from the default report (G6) | `RouterOS` output pasted on a real device and `tc`-equivalent `show` recorded; until then the generator is not in the stable API |
| S7 | Calibration: parse+range-check on load; document or replace the custom `.dist` name (G7) | round-trip test with a corrupted-but-hash-valid sidecar is rejected |
| S8 | Module split: two `go.mod` files, `v0.1.0` tags, public-API docs per (R10 §1); `pkg/campagne`/`pkg/audit`/`pkg/api` become consumers via `require` + `replace` during migration (R8) | `go list -m all` in each module shows no internal packages; `make verify` green |

## 8. Anti-toy criteria

A standalone claim of "works" for these two packages is not credible unless all of the following hold. They are deliberately stricter than the current host tests (which run against `FakeRunner`, `qdisc.go:57-62`).

1. **Real `tc` on a Linux VM.** Every `ApplyShaper`/`ApplyNetem*`/`ApplyIngressShaper` case is executed against real `tc` (not `FakeRunner`) and the resulting `tc -s qdisc show` output is recorded in the test fixture; the pfifo_fast-invisibility case (R7) is one of them. `Reset`'s English-string tolerance (G9) is exercised against the VM's actual error text.
2. **RFC 8290 parameters are asserted, not commented.** A test fails if fq_codel is emitted without `target 5ms interval 100ms limit 10240 quantum 1514 flows 1024` (R1 §5.2, R4), in both packages (G1).
3. **CAKE equivalence is proven or the delta is named.** Bench CAKE and production CAKE either produce identical `tc` lines (R3) or the module documents the delta as a measured difference (G2) — "close enough" is not a state.
4. **Stats do not lie.** `SumBytes`/`PollStats` on a pfifo_fast-default interface (R7) returns an error or an explicit empty, never a silent 0 that reads as "measured 0" (G5, G10); goodput on a stacked netem+tbf+fq_codel is within the leaf-only expectation, not 2–3× inflated (`stats.go:94-98`).
5. **Nothing unmeasured is emitted.** `Validate` rejects zero/negative/non-finite capacity and non-positive RTT (`remediate.go:84-101`); `BuildCalibration` refuses <30 samples and p50 ≤ floor (`dist.go:80-93`); `LoadCalibration` fails closed on any divergence (`dist.go:261-303`); the API returns `404` rather than a fabricated remediation (`docs/api.md:101`).
6. **RouterOS is verified or fenced.** No RouterOS script is presented as production-ready until it has been pasted on a real RouterOS device and the resulting queue shown — until then every RouterOS claim carries `unverified — primary source not found` (G6, §0).
7. **Version numbers mean what they say.** `v0.x` carries no stability promise (R9); a minor bump adds only backward-compatible API (R9, R10 §7); a breaking change bumps major and the module path gains `/v2` (R8). A release that changes emitted `tc` lines is at least a minor version even if the Go API is unchanged, because the public API here includes the generated config (R10 §1).
