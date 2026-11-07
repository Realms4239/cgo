# pkg/remediate RouterOS strings + cmd/cgo-chr + kit/chr/ as standalone validator — primary-source research

*Date: 2026-11-07 · Scope: `pkg/remediate` RouterOS generators + `cmd/cgo-chr` + `kit/chr/` + `pkg/audit` RouterOS parser/diff/proof, read as if extracted tomorrow as `cgo-chr-validate`. Research only — no code changed.*

## 0. Method & sources (table)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | MikroTik RouterOS — CAKE (Common Applications Kept Enhanced) | https://help.mikrotik.com/docs/display/ROS/CAKE (`/spaces/ROS/pages/196345874/CAKE`) | fetched 2026-11-07 | CAKE properties: bandwidth, overhead, RTT, Diffserv, Flow Mode, ACK Filter, NAT, Wash; RTT schemes; Flow Isolation Parameter; Diffserv RFC2474/RFC2475 |
| R2 | MikroTik RouterOS — Queues | https://help.mikrotik.com/docs/display/ROS/Queues (`/spaces/ROS/pages/328088/Queues`) | fetched 2026-11-07 | `/queue simple` vs `/queue tree` vs `/queue type` vs `/queue interface`; HTB; CIR=`limit-at`, MIR=`max-limit`; default queue types listing; RouterOS 7.1beta3 added CoDel / FQ-Codel / CAKE; full CAKE property table (`cake-bandwidth`, `cake-rtt`, `cake-flowmode`, `cake-diffserv`, `cake-ack-filter`, `cake-nat`, `cake-wash`, `cake-overhead`, `cake-mpu`, `cake-autorate-ingress`, `cake-memlimit`, etc.); SFQ/PCQ/FIFO kinds |
| R3 | MikroTik RouterOS — Queue types | https://help.mikrotik.com/docs/display/ROS/Queue+types (`/spaces/ROS/pages/196345871/Queue+types`) | fetched 2026-11-07 | Queue types at a glance: BFIFO/PFIFO/CAKE/CoDel/FQ_CoDel/MQ_PFIFO/RED/SFQ; CAKE = CoDel+FQ+other, FQ_CoDel = excellent latency + fair distribution, SFQ = fair distribution, no latency/congestion management |
| R4 | RFC 8290 — FQ-CoDel (Experimental) | https://www.rfc-editor.org/rfc/rfc8290.txt | fetched 2026-11-07 | FQ-CoDel algorithm: flow queueing + CoDel AQM, defaults interval 100 ms (§5.2.1), target 5 ms (§5.2.2), packet limit MUST (§5.2.3), quantum 1514 (§5.2.4), flows 1024 (§5.2.5); deployment note: CAKE further refinement (§5.3) |
| R5 | Go Modules Reference | https://go.dev/doc/modules/gomod-ref | referenced by sibling research `docs/research/2026-11-07-module-qdisc-remediate.md:16,120` | module path + version = unique ID and import prefix; `go` directive = minimum Go version; v2+ requires `/vN` suffix — `unverified — primary source not found` for direct fetch in this session (no fetch attempted here) |
| — | MikroTik RouterOS `/queue type`, `/queue tree`, `kind=cake`, `cake-bandwidth`, `cake-rtt`, `cake-ack-filter`, RouterOS ≥ 7.10 | https://help.mikrotik.com/docs/display/ROS/Queue | **404** — `https://help.mikrotik.com/docs/display/ROS/Queue` → 404 | unverified — primary source not found (initial in-repo claim `pkg/remediate/remediate.go:281` “Requires RouterOS >= 7.10”) |
| — | MikroTik RouterOS CHR / Cloud Hosted Router, `/export` syntax | https://help.mikrotik.com/docs/display/ROS/CHR , https://help.mikrotik.com/docs/display/ROS/Export | **404** — both returned 404 | unverified — primary source not found (CHR image acquisition, license, `/export terse`, boot/login strings) |
| — | OpenWrt SQM / CAKE documentation | https://openwrt.org/docs/guide-quick-start/sqm | **blocked** — Anubis anti-bot interstitial, no content returned | unverified — primary source not found |
| — | 95 % shaper headroom (`shaperHeadroom = 0.95`) | — | unverified — primary source not found (only in-repo rationale `pkg/remediate/remediate.go:70-74`) | `ShapedRate` calibration |
| — | RouterOS `kind=fifo` (bare), `cake-ack-filter=filter` value, `cake-wash=yes`, `kind=sfq sfq-perturb=10` defaults | — | unverified — primary source not found (R1/R2 list `cake-ack-filter` default `none`, wash default `no`, `sfq-perturb` present on `wireless-default`/`hotspot-default` but value 5 shown, not 10) | `pkg/remediate/remediate.go:292-316,358-379` |
| — | Assumed LAN interface `bridge`, RouterOS “no IFB” download-on-LAN-egress claim | — | unverified — primary source not found (only in-repo doc `pkg/remediate/remediate.go:52-56,324-329`) | `RouterOSIngress` parent |

Method: read `pkg/remediate/*.go` (`remediate.go`, `remediate_test.go`, `ingress_test.go`), `cmd/cgo-chr/*.go` (`main.go`, `run.go`, `serial.go`, `transport.go`, `*_test.go`), `kit/chr/*` (`README.md`, `setup-mikrotik-chr.sh`, `export-fixture.txt`, `export-ingress-fixture.txt`, `chr-image.sha256`), `pkg/audit/routeros_{config,compare,proof}.go`, `docs/VERIFY.md:57` (CHR BLOCKED), `go.mod`, `docs/superpowers/plans/2026-10-30-cgo-chr-router-validation.md`, `docs/superpowers/plans/README.md:44`, and `docs/research/2026-11-07-module-qdisc-remediate.md` (sibling). Fetch attempts are recorded above; every claim below cites `file:line` or (R#); anything else is labelled `unverified — primary source not found`.

## 1. Responsibility

### 1.1 `pkg/remediate` RouterOS strings — measured link → RouterOS v7+ script text

`pkg/remediate/remediate.go:1-8` states the package contract: “a remediation is only emitted from numbers that were actually measured … a remediation generated from a guessed bandwidth is worse than no remediation”. The RouterOS half of the package is two pure string generators (no `exec`, no I/O):

- `RouterOS(l Link, qdisc string) (string, error)` upload — `pkg/remediate/remediate.go:268-322`: requires `l.RouterOSInterface` (`272-274`), converts shaped rate to bits `round(shapedMbps*1000000)` (`277`), header states measured bottleneck + RTT + 95 % headroom (`280-284`), idempotent cleanup first (`/queue tree remove [find where comment="cgo-managed"]`, `/queue type remove [find where name~"^cgo-"]`, `287-288`), then per-qdisc type+tree pair (cake `291-298`, fq_codel→sfq `300-310`, pfifo→fifo `312-317`).
- `RouterOSIngress(l Link, qdisc string) (string, error)` download — `pkg/remediate/remediate.go:330-384`: joins upload objects under disjoint names/comments (`cgoingress-*` / `cgo-managed-ingress`, `352-353`, enforced by `TestIngressCleanupCannotDeleteUploadObjects` in `pkg/remediate/ingress_test.go:124-136`), shapes on the LAN/private interface (`334`, `362`, `370`, `377`) because “RouterOS has no IFB” (`324-326`, `344-346`), flowmode `dual-dsthost` for CAKE ingress (`360`), no `ack-filter` (`357`).
- Quoted syntax emitted (verified against R1/R2/R3 where noted): `/queue type add name=cgo-cake kind=cake cake-bandwidth=<bits> cake-rtt=<ms>ms cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 cake-ack-filter=filter cake-wash=yes` + `/queue tree add name=cgo-wan-out parent=<wan> max-limit=<bits> queue=cgo-cake comment="cgo-managed"` — `pkg/remediate/remediate.go:292-297`. R1 confirms CAKE has bandwidth, RTT, Diffserv, Flow Mode, ACK Filter, NAT, Wash; R2's property table confirms `cake-bandwidth`, `cake-rtt` (default `100ms`), `cake-flowmode` (default `triple-isolate`), `cake-diffserv` (default `diffserv3`, options `diffserv4`/`diffserv3`), `cake-ack-filter` (default `none`), `cake-nat` (default `no`), `cake-wash` (default `no`); R1 confirms `diffserv4` = four tins (Bulk/Best Effort/Video/Voice), `triple-isolate` = default, `dual-dsthost` = suitable for incoming traffic to LAN, `dual-srchost` = outgoing LAN→internet. Value-level claims `cake-ack-filter=filter`, `cake-wash=yes`, `cake-nat=yes` are `unverified — primary source not found` (R2 lists defaults, not the `filter` enum).
- Input is a measured `Link{Interface, RouterOSInterface, RouterOSLanInterface, CapacityMbps, RTTms, Overhead, MTU}` — `pkg/remediate/remediate.go:22-39`; `Validate()` rejects empty interface, non-positive/non-finite capacity, non-positive RTT, negative overhead via `ErrUnmeasured` — `pkg/remediate/remediate.go:83-101`. The LAN fallback `DefaultRouterOSLan = "bridge"`, `RouterOSLan() (iface, assumed)` — `pkg/remediate/remediate.go:56-64` — is an explicit assumption; the script prints a French `ATTENTION` warning when assumed — `pkg/remediate/remediate.go:339-343`, pinned by `TestRouterOSIngressAssumesBridgeAndSaysSo` in `pkg/remediate/remediate_test.go:215-227`. The “no RouterOS LAN name has ever been observed” admission is recorded in `docs/VERIFY.md:64`.
- Honesty guard: for `fq_codel` the generator prints “RouterOS v7 has no fq_codel queue kind … This is sfq … It is NOT fq_codel” — `pkg/remediate/remediate.go:301-306`, `367-368`, pinned by `TestRouterOSFqCodelIsHonest` in `pkg/remediate/remediate_test.go:172-183`. **This claim contradicts R2**: R2 states “The following queue kinds CoDel, FQ-Codel, and CAKE available since RouterOS version 7.1beta3” and lists FQ-Codel properties (`fq-codel-interval` 100 ms, `fq-codel-target` 5 ms, `fq-codel-flows` 1024, `fq-codel-limit` 10240, `fq-codel-quantum` 1514, `fq-codel-ecn` yes, `fq-codel-memlimit` 32 MiB, `fq-codel-ce-threshold`). R3 likewise lists FQ_CoDel as a queue type. The generator comment is therefore stale/inaccurate relative to R2/R3 — see gap G1.

What it deliberately does **not** do: measure, execute, connect to a RouterOS device, or validate that RouterOS accepts the output. Generation is pure; acceptance is owned by `cmd/cgo-chr` + `pkg/audit` (§1.2–§1.3).

### 1.2 `cmd/cgo-chr` — auxiliary QEMU/serial orchestrator, not shipped

Header: “auxiliary tool, not part of the shipped binary: it boots the MikroTik CHR image in QEMU, feeds it the sealed report's RouterOS script, and seals proof that RouterOS 7 accepted it. Every judgement lives in `pkg/audit`; this file only orchestrates.” — `cmd/cgo-chr/main.go:1-5`. Concrete responsibility:

- CLI: flags `--report`, `--image`, `--level` (`windows-tcg|linux-namespace`), `--timeout` (default 180 s), `--out` (default `data/chr/<audit_id>`), `--qemu` (default `qemu-system-x86_64`), `--export-file` (replay without QEMU), `--dry-run` — `cmd/cgo-chr/main.go:14-19,42-50`. `BuildQEMUArgs(RunConfig) []string` (`23-39`) always uses `-m 256 -smp 1 -drive file=<image>,format=raw,if=virtio -netdev user,id=n0 -device virtio-net-pci,netdev=n0 -serial stdio -display none`, accelerator `tcg` or `kvm` by level (`25-27`).
- `run(cfg, reportPath, exportFile, outDir) int` — `cmd/cgo-chr/run.go:19-107`: `levelAllowed` refuses `linux-namespace` off Linux and unknown levels with exit 2 before any QEMU work (`157-165`, `20-23`); reads sealed `audit_report.json` requiring `audit_id` and at least one RouterOS script (`24-37`); concatenates `RouterOS + "\n" + RouterOSIngress` (`41`) and parses expectation via `audit.ParseRouterOSConfig` (`42`); replay reads a supplied export file (`46-53`) else stats the image, starts QEMU, `drive()`s the console (`55-78`); compares via `audit.CompareRouterOS` (`81-89`); seals via `audit.WriteCHRProof` (`90-97`); prints `chr: preuve <path> · <passed>/<total> vérifications · niveau <level>` (`102`); exit 0 all-pass, 1 checks failed / import refused, 2 usage / unreadable input (`102-106`, `21-22`, `26-32`, `49-50`, `56-57`).
- Serial driver `drive(tr, script, timeout)` — `cmd/cgo-chr/serial.go:23-70`: regexes `login:`, `password:`, `[admin@...] >`, error line `failure:|syntax error|bad command name|expected end of command|input does not match` (`12-17`); logs in as `admin` with empty password (`34-45`), sends each non-comment script line, records first error line verbatim (`46-61`), then `/export terse` and `cleanupExport` (`62-69`, `103-111`). `expect` reads byte-by-byte with a per-call timeout (`72-99`). Transport seam `Transport interface{ io.ReadWriteCloser }` + `qemuTransport` (stdin/stdout pipes, `Process.Kill()` on close) — `cmd/cgo-chr/transport.go:8-23`, `cmd/cgo-chr/run.go:109-128`. CI seam `fakeCHR` — `cmd/cgo-chr/serial_test.go:16-73` — replays banner `MikroTik RouterOS 7.15.3`, login, prompt, a canned `/export` answer, and a `REJECT` → `failure: bad command name` path, pinned by `TestDriveHappyPath`/`TestDriveRejection` (`75-101`).
- Exit / level semantics in tests: replay against `kit/chr/export-fixture.txt` yields 12/12 checks (`TestRunReplayWritesProof`, `cmd/cgo-chr/run_test.go:69-107`); mutated export fails exactly one named check (`TestRunReplayMutatedExportFails`, `109-133`); ingress replay proves `cgoingress-lan-in.parent` (`TestRunReplayValidatesIngressScript`, `137-167`); wrong LAN parent fails named (`TestRunReplayRejectsWrongIngressParent`, `171-203`); empty script → 2, unknown level → 2 (`205-219`).

What it deliberately does **not** do: parse, compare, or seal itself (all `pkg/audit`); assert throughput/latency/production change (`kit/chr/README.md:38-41` — “Elle ne prouve ni débit, ni latence, ni qu'une production a changé”). The `linux-namespace` level name implies namespace attach that `BuildQEMUArgs`/`run` never actually perform — only the KVM accelerator string differs (`cmd/cgo-chr/main.go:25-27`) and `levelAllowed` gates the host OS (`cmd/cgo-chr/run.go:161-163`); see gap G4.

### 1.3 `pkg/audit` RouterOS parser/diff/proof + `kit/chr/` — the validator core and its human prerequisites

The validator slice inside `pkg/audit` (already inventoried by sibling research `docs/research/2026-11-07-module-probe-audit.md:147-149`):

- `RouterOSObject{Section,Name,Props}` / `RouterOSConfig{Objects}` + `Get(section,name)` (duplicates = stacked failure) + `ParseRouterOSConfig(text)` — `pkg/audit/routeros_config.go:9-85`: reads both dialects (full-path `/queue type add …` script lines and bare `add …` under a `/queue type` export header), keeps only `queue type` + `queue tree` (`66`), ignores removals/prints/comments (`31-32`), joins `\`-continuation lines (`87-107`), preserves quoted `key="v v"` pairs (`119-141`).
- `RouterOSCheck{Name,Expected,Found,Pass}` + `CompareRouterOS(expected,found)` — `pkg/audit/routeros_compare.go:11-64`: per-intended-object, per-property checks; duplicate intended name = fail (`26-33`), absent = fail (`34-41`); `normalizeROS` absorbs only quotes/case/trailing `.0` (`66-73`); extra device objects ignored (`18-21`).
- `CHRInfo/ImportResult/CHRProof{Schema, AuditID, ReportSHA256, CHR, Level, Accelerator, Import, Checks, ExportSHA256, TranscriptSHA256}` + `FileSHA256` + `WriteCHRProof(dir,p,export,transcript)` — `pkg/audit/routeros_proof.go:12-76`: atomic temp+rename writes of `proof.json`/`export.txt`/`transcript.txt` (`48-76`), `Level` must be `windows-tcg`/`linux-namespace` (`52-54`), schema stamped `cgo.chr-proof/1` (`58`); never committed, cited by path+hash from `docs/VERIFY.md` (`routeros_proof.go:24-26`).
- `kit/chr/` thin wrapper and fixtures: `setup-mikrotik-chr.sh` is 9 lines — `exec go run ./cmd/cgo-chr "$@"` after `cd` to repo root, “Il ne juge rien … Toute la logique est en Go” — `kit/chr/setup-mikrotik-chr.sh:1-9`; `export-fixture.txt` (WAN CAKE, header “# 2026-10-30 12:00:00 by RouterOS 7.15.3”) — `kit/chr/export-fixture.txt:1-9`; `export-ingress-fixture.txt` (LAN CAKE `dual-dsthost`, “# 2026-10-31 00:00:00 by RouterOS 7.15.3”) — `kit/chr/export-ingress-fixture.txt:1-8`; `chr-image.sha256` is an empty placeholder comment (“Remplacer la ligne ci-dessous par le nom et le SHA-256 de l'image réellement téléchargée”) — `kit/chr/chr-image.sha256:1-4`; `README.md` lists human prerequisites — install QEMU, download CHR from mikrotik.com + accept license, place image in `kit/chr/cache/` (gitignored), never commit image — `kit/chr/README.md:8-14` — and the replay command (`31-36`).

### 1.4 Verification status (CHR BLOCKED)

`docs/VERIFY.md:57` (2026-10-30, DESKTOP-B6K9UCM): `go build ./... && go test ./... -count=1` exit 0 (cmd/cgo-chr included), frontend 46 files/213 tests, `python tools/verify.py` 6 passed; parser + field diff + proof writer + report/UI line unit-tested; replay against fixture yields 12/12 checks with a sealed proof; mutated export fails exactly one named check. **Nothing has booted CHR: QEMU is not installed on this host and no CHR image is cached, so no RouterOS acceptance exists.** A `windows-tcg` proof upgrades the row; the Linux namespace attach is the second level and never substitutes for it. Plan index: slice 16 “RouterOS v7 compliance via CHR (QEMU)”, status **Deferred (Tasks 1–9 built; Windows boot pending QEMU)**, condition “human: install QEMU + CHR image” — `docs/superpowers/plans/README.md:44`. The slice-22/26′ rows add no CHR acceptance: remediate tests are string-level (`TestRouterOSEmitsOnlyValidKinds`, `pkg/remediate/remediate_test.go:251-297`), and every re-certification row repeats “No live audit, campaign, bench, CHR … claim” (e.g. `docs/VERIFY.md:74-81`).

## 2. Exported API (file:line)

### 2.1 `pkg/remediate` RouterOS surface

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Link{Interface, RouterOSInterface, RouterOSLanInterface, CapacityMbps, RTTms, Overhead, MTU}` | struct | `pkg/remediate/remediate.go:22-39` | `MTU=0` ⇒ 1500 (`103-108`), `burstBytes` floor 1500 (`112-118`) |
| `QdiscCake`/`QdiscFqCodel`/`QdiscPfifo` | consts | `pkg/remediate/remediate.go:42-46` | `"cake"`/`"fq_codel"`/`"pfifo_fast"` — duplicates `model.Qdisc` as strings |
| `IngressIFB` | const | `pkg/remediate/remediate.go:50` | `"ifb0"` — Linux IFB only |
| `DefaultRouterOSLan` | const | `pkg/remediate/remediate.go:56` | `"bridge"` — assumed, labelled in script; `unverified — primary source not found` |
| `RouterOSLan()` | method | `pkg/remediate/remediate.go:59-64` | `(iface, assumed)` |
| `ShapedRate(measuredMbps)` | func | `pkg/remediate/remediate.go:68` | `× shaperHeadroom (0.95)`; 0.95 `unverified — primary source not found` |
| `ErrUnmeasured{Field,Why}` | type | `pkg/remediate/remediate.go:77-81` | |
| `Validate()` | method | `pkg/remediate/remediate.go:84-101` | |
| `RouterOS(l,qdisc)` | func | `pkg/remediate/remediate.go:268-322` | upload script; header claims “Requires RouterOS >= 7.10” (`281`) — `unverified — primary source not found`, contradicted by R2 (7.1beta3, §4) |
| `RouterOSIngress(l,qdisc)` | func | `pkg/remediate/remediate.go:330-384` | download script; disjoint `cgoingress-*`/`cgo-managed-ingress` cleanup (`352-353`) |
| `LinuxTC`/`LinuxTCIngress` | funcs | `pkg/remediate/remediate.go:135-194`, `215-260` | Linux side of the same package; included because they share `Link`/`Validate`/`shaperHeadroom` and are the non-RouterOS consumers of the same measurement |
| `Recommendation{Qdisc,Reason}` / `Recommend` | type+func | `pkg/remediate/remediate.go:388-417` | thresholds `unverified — primary source not found` |

Unexported: `shaperHeadroom` (`74`), `mtu` (`103`), `burstBytes` (`112`), `shapedRate` (`121`), `rate` (`124`), `ingressSetup` (`199`).

### 2.2 `cmd/cgo-chr` surface

| Export | Kind | Location | Notes |
|---|---|---|---|
| `RunConfig{QEMU,Image,Level,Timeout}` | struct | `cmd/cgo-chr/main.go:14-19` | |
| `BuildQEMUArgs(cfg)` | func (exported) | `cmd/cgo-chr/main.go:23-39` | deterministic QEMU line; first element is the binary, stripped at `run.go:113` |
| `Transport` | interface | `cmd/cgo-chr/transport.go:8-10` | `io.ReadWriteCloser` seam |
| `qemuTransport` | struct | `cmd/cgo-chr/transport.go:12-23` | `Close()` = stdin close + `Process.Kill()` |

Unexported but load-bearing: `main()` flag wiring (`main.go:41-65`), `joinCommand` (`67-76`), `fatal` (`78-81`), `run` (`run.go:19-107`), `startQEMU` (`109-128`), `accelerator` (`130-135`), `imageHash` (`137-143`), `passed` (`145-153`), `levelAllowed` (`157-165`), `drive`/`expect`/`cleanupExport`/`lastErrorLine` + regexes `loginRe/passRe/promptRe/errorRe` + `errTimeout` (`serial.go:12-121`).

### 2.3 `pkg/audit` validator surface consumed by `cmd/cgo-chr`

| Export | Kind | Location | Notes |
|---|---|---|---|
| `RouterOSObject{Section,Name,Props}` / `RouterOSConfig{Objects}` / `Get` / `ParseRouterOSConfig` | types+funcs | `pkg/audit/routeros_config.go:9-85` | keeps only `queue type`/`queue tree` (`66`) |
| `RouterOSCheck{Name,Expected,Found,Pass}` / `CompareRouterOS` | type+func | `pkg/audit/routeros_compare.go:11-64` | `normalizeROS` (`66-73`) |
| `CHRInfo`/`ImportResult`/`CHRProof` / `FileSHA256` / `WriteCHRProof` | types+funcs | `pkg/audit/routeros_proof.go:12-76` | schema `cgo.chr-proof/1`, level gate (`52-54`) |
| `Remediation{RouterOS,RouterOSIngress}` (report field consumed at `run.go:34,41`) | struct field | `pkg/audit/report.go` (via `cmd/cgo-chr/run.go:14-17`) | |

### 2.4 `kit/chr/` surface

| Path | Role | Location |
|---|---|---|
| `kit/chr/setup-mikrotik-chr.sh` | wrapper → `go run ./cmd/cgo-chr` | `kit/chr/setup-mikrotik-chr.sh:1-9` |
| `kit/chr/README.md` | human prerequisites + usage + “ne prouve ni débit…” | `kit/chr/README.md:1-41` |
| `kit/chr/export-fixture.txt` | WAN CAKE replay fixture (12-check) | `kit/chr/export-fixture.txt:1-9` |
| `kit/chr/export-ingress-fixture.txt` | LAN CAKE replay fixture | `kit/chr/export-ingress-fixture.txt:1-8` |
| `kit/chr/chr-image.sha256` | placeholder, no image pinned | `kit/chr/chr-image.sha256:1-4` |
| `kit/chr/cache/` | gitignored image drop dir (referenced `kit/chr/README.md:14`) | — (directory not present) |

## 3. Dependencies

### 3.1 Module identity and Go version

`go.mod:1-3` declares `module github.com/Realms4239/cgo` with `go 1.25`, requiring `fyne.io/fyne/v2 v2.8.1`, `github.com/charmbracelet/bubbletea v1.3.10`, `github.com/charmbracelet/lipgloss v1.1.0`, `golang.org/x/sys v0.36.0` (`go.mod:5-10`) plus ~40 indirect deps (`go.mod:12-60`). A standalone validator must shed all of these: the CHR slice itself imports only stdlib + `pkg/audit` (see §3.2), so the Fyne/Tea stack is pure inheritance — the same inheritance already noted for the CLI research (`docs/research/2026-11-07-module-cli-frontend.md:140,217`) and for `go list -m all` noise in the monolith.

### 3.2 Internal dependency graph

| Package | Imports (internal) | Imports (external) | Imported by |
|---|---|---|---|
| `pkg/remediate` (RouterOS fns) | **none** — `fmt`, `math`, `strings` only (`pkg/remediate/remediate.go:11-15`) | stdlib only | `pkg/audit/report.go` (via `remediate.Recommend/LinuxTC/RouterOS/LinuxTCIngress/RouterOSIngress/ShapedRate`); `pkg/api/server_audit.go:187,349-355` constructs `Link`; `pkg/audit/*_test.go` fixtures |
| `cmd/cgo-chr` | `pkg/audit` only (`cmd/cgo-chr/run.go:11`) | stdlib only (`flag/fmt/os/time`, `encoding/json`, `os/exec`, `bufio/errors/regexp/strings`, `io`) (`cmd/cgo-chr/main.go:7-12`, `run.go:3-10`, `serial.go:3-10`, `transport.go:3-6`) | `kit/chr/setup-mikrotik-chr.sh:9` (`go run ./cmd/cgo-chr`); tests (`cmd/cgo-chr/*_test.go`) |
| `pkg/audit` (CHR subset) | `pkg/remediate` (report assembly) + rest of `pkg/audit` (report/telemetry/match/…) | stdlib + internal packages | `cmd/cgo-chr/run.go:11-12,42,81-82,90-97` |
| `kit/chr/` | subprocess `go run ./cmd/cgo-chr` (`kit/chr/setup-mikrotik-chr.sh:9`) | `go`, `bash` | human operator |

Consequences for extraction: `pkg/remediate` RouterOS generators extract as-is (zero internal deps). `cmd/cgo-chr` extracts with only the `pkg/audit` CHR subset (`ParseRouterOSConfig`/`CompareRouterOS`/`WriteCHRProof`/`FileSHA256`/`CHRProof`/`CHRInfo`/`ImportResult`/`Remediation` JSON shape), not all of `pkg/audit` (telemetry, verdict, timeline, artifacts, kernel adapter stay behind). The sibling probe/audit research already prescribes exactly this subset (`docs/research/2026-11-07-module-probe-audit.md:148-149`; `docs/research/2026-11-07-module-cli-frontend.md:236,248` — “Contents: `cmd/cgo-chr` + `pkg/audit` CHR-proof subset”).

### 3.3 External / runtime dependencies (not Go modules)

- QEMU binary `qemu-system-x86_64` (`cmd/cgo-chr/main.go:47` default, `cmd/cgo-chr/run.go:109-112` `exec.LookPath` check with “qemu introuvable — installez QEMU, voir kit/chr/README.md”). Per `kit/chr/README.md:10-11` QEMU is absent on the verification host (`DESKTOP-B6K9UCM`, 2026-10-30) and installation is a human decision.
- CHR disk image (`--image`, required unless `--export-file` — `cmd/cgo-chr/main.go:60-62`; `os.Stat` refusal — `cmd/cgo-chr/run.go:55-58`), acquired from mikrotik.com under license, never committed (`kit/chr/README.md:12-14`); its SHA-256 pin file exists but is empty (`kit/chr/chr-image.sha256:1-4`); `imageHash` reads the whole image into memory for hashing (`cmd/cgo-chr/run.go:137-143`) — no streaming hash.
- RouterOS ≥ 7.10 per the script header (`pkg/remediate/remediate.go:281,347`) — `unverified — primary source not found`, and R2 actually places CoDel/FQ-Codel/CAKE since 7.1beta3 (see §4, gap G1).
- A RouterOS serial console speaking `login:`/`password:`/`[admin@…] >` prompts and `/export terse` — `cmd/cgo-chr/serial.go:13-17,62` — `unverified — primary source not found` (CHR `/export` page 404, §0); behaviour is pinned only by the `fakeCHR` fixture (`cmd/cgo-chr/serial_test.go:24,59`).

### 3.4 Build / release wiring today

`cmd/cgo-chr` is auto-discovered by `go build ./... && go test ./...` (`docs/VERIFY.md:57`), referenced in the slice matrix (`tools/slice_matrix.json:393,400` — `go test ./cmd/cgo-chr/ -count=1`), but not in the shipped binary list (`.goreleaser.yaml` ships only `cgo` + `meteolink`; `Makefile` builds only `./cmd/cgo`) — consistent with the “auxiliary, not shipped” header (`cmd/cgo-chr/main.go:1-2`, noted in sibling research `docs/research/2026-11-07-module-cli-frontend.md:67`).

## 4. Primary-source requirements

| # | Behaviour / emitted string | Requirement (primary source) | Where it lands | Verdict |
|---|---|---|---|---|
| P1 | CAKE existence + general approach | CAKE is a modern advanced queue management algorithm: fairness, bufferbloat minimization, flow isolation, bandwidth shaping, small-packet prioritization (R1 “Description”) | `RouterOS` cake branch `pkg/remediate/remediate.go:291-298` | satisfied by R1 |
| P2 | `cake-bandwidth` | Bandwidth Limit property: “manually set the bandwidth limit for the CAKE shaper” (R1 Property table; R2 `cake-bandwidth (default: ) — Sets the shaper bandwidth`) | `cake-bandwidth=<bits>` `remediate.go:293,359` | satisfied by R1/R2; bits-vs-Mbit/s rendering `unverified — primary source not found` |
| P3 | `cake-rtt` | RTT property: “Manually specify an RTT” (R1); R2 `cake-rtt (default: 100ms)`; RTT Schemes: `rtt TIME` manual, `internet` = 100 ms default suitable for most Internet traffic (R1 “RTT Schemes”) | `cake-rtt=<RTT>ms` `remediate.go:293,359` | satisfied by R1/R2; “measured idle RTT” provenance is in-repo only |
| P4 | `cake-flowmode=triple-isolate` (upload) / `dual-dsthost` (download) | Flow Mode: `triple-isolate` is the default (R1 FLOW ISOLATION, R2 `cake-flowmode default: triple-isolate`); `dual-dsthost` “suitable for incoming traffic to a LAN from the internet” (R1; R2 “Good for use on ingress traffic to a LAN…”) | `remediate.go:294,360` | satisfied by R1/R2 |
| P5 | `cake-diffserv=diffserv4` | Diffserv property divides traffic into tins (R1); R2 `cake-diffserv (default: diffserv3)` with `diffserv4` = 4 tins (Bulk 6.25 %, Best Effort 100 %, Video 50 %, Voice 25 %), `diffserv3` = 3 tins (default) (R1 “# Diffserv RFC2474 and RFC2475”, R2) | `remediate.go:294,360` | satisfied by R1/R2 (notably: `diffserv4` ≠ default `diffserv3`) |
| P6 | `cake-nat=yes` | NAT property “considers post-NAT addresses/ports when isolating flows” (R1); R2 `cake-nat (default: no)` “Instructs Cake to perform a NAT lookup before applying a flow-isolation rule” | `remediate.go:294,360` | satisfied by R1/R2 for the keyword; `yes` value form `unverified — primary source not found` |
| P7 | `cake-ack-filter=filter` (upload only) | ACK Filter property: compresses TCP ACKs to save upload bandwidth (R1); R2 `cake-ack-filter (default: none)` | `remediate.go:295` | keyword satisfied by R1/R2; value `filter` and omission on download `unverified — primary source not found` |
| P8 | `cake-wash=yes` | Wash property clears mis-marked DSCP after priority queuing (R1); R2 `cake-wash (default: no)` | `remediate.go:295,361` | keyword satisfied; `yes` value `unverified — primary source not found` |
| P9 | `cake-overhead` / `overhead <bytes>` | Overhead property per-packet bytes; manual overhead range −64…256 bytes (R1 “Overhead Schemes”); R2 `cake-overhead (-64…256, default: )` | Linux `overhead %d` emitted only when `Overhead > 0` (`remediate.go:163-165,236-238`), `ebnf` int accepts negatives (`remediate.go:97-99` refuses negatives) | gap G6 — Linux accepts negatives per R1/R2, in-repo `Validate` forbids them |
| P10 | `/queue type`, `/queue tree`, `kind=cake`, `max-limit`, `parent=`, `queue=`, `comment=` | RouterOS queues via `/queue tree` (one-directional, interface or global parent, needs `/ip firewall mangle` marking) (R2 “Queue Tree”); Queue Types sub-menu lists user-specific types (R2 “Queue Types”); CIR = `limit-at`, MIR = `max-limit` (R2 “Rate limitation principles”); queue kinds include CAKE (R2/R3) | `remediate.go:287-288,292-297,308-309,315-316,352-363,370-371,377-378` | `/queue type` + `/queue tree` + `max-limit` + `parent` satisfied by R2; `/queue interface` add and `kind=` literal rendering `unverified — primary source not found` (no RouterOS grammar page fetched) |
| P11 | FQ-CoDel as evaluation baseline | FQ-CoDel = DRR scheduler + CoDel AQM per queue, 5-tuple hashing, 1024 queues default, CoDel interval/target defaults 100 ms / 5 ms (R4 §1.3, §5.2.1–§5.2.5) | Remediation comments; bench `pkg/qdisc` (out of scope here) | algorithm satisfied by R4; `fq_codel` as a RouterOS `kind` — see P12 |
| P12 | “RouterOS has no fq_codel; emit sfq instead” | R2: “The following queue kinds CoDel, FQ-Codel, and CAKE available since RouterOS version 7.1beta3” + full FQ-Codel property table; R3: FQ_CoDel listed among queue types | `remediate.go:300-310,366-372` | **contradicted by R2/R3** — needs correction to `kind=fq-codel` (property spelling `unverified — primary source not found` until a property-level fetch confirms `fq-codel-*` vs `fq_codel`) |
| P13 | RouterOS ≥ 7.10 gate | Script header “Requires RouterOS >= 7.10 for kind=cake. Verify with: /system resource print” (`remediate.go:281,347`) | header string | `unverified — primary source not found`; R2 implies availability since 7.1beta3 — the header is either over-strict or references an unverified property |
| P14 | `kind=sfq sfq-perturb=10` | R1/R2: SFQ = stochastic fairness queueing, 1024 sub-streams; R2 default types show `sfq-perturb=5` (`wireless-default`, `hotspot-default`), R2 SFQ section describes `sfq-allot` default (R3: SFQ “does not manage latency or congestion”) | `remediate.go:307,369` | kind satisfied by R2/R3; `sfq-perturb=10` `unverified — primary source not found` (defaults shown are 5) |
| P15 | `kind=fifo` | R2/R3 FIFO kinds: BFIFO / PFIFO / MQ-PFIFO, parameters `pfifo-limit`/`bfifo-limit`; R2 default types show `kind=pfifo pfifo-limit=50`, `kind=mq-pfifo` | `remediate.go:314,376` | `unverified — primary source not found` for bare `kind=fifo` (no such literal in R2/R3 listings) |
| P16 | RouterOS download = egress queue on LAN (no IFB) | RouterOS `/queue tree` is “the only way how to add a queue on a separate interface” (R2 “Queue Tree”); no IFB concept appears in R2/R3 | `remediate.go:324-346` | directionality satisfied by R2; “no IFB” phrasing `unverified — primary source not found` |
| P17 | QEMU/CHR boot + console (`login:`, `[admin@…] >`, `/export terse`) | CHR image, serial console protocol, `/export` grammar | `cmd/cgo-chr/main.go:23-39`, `serial.go:12-17,62`, `kit/chr/README.md:8-14` | `unverified — primary source not found` (CHR + Export pages 404, §0) |
| P18 | Go module + SemVer for the standalone repo | module path = ID + import prefix, `go` = minimum version (R5); v0 unstable, minor/patch/major signal API change (R9, cited in `docs/research/2026-11-07-module-qdisc-remediate.md:120`); public API MUST be declared (R10 §1, same citation) | §6 of this doc | R5/R9/R10 not directly fetched in this session — cited via sibling research; mark `unverified — primary source not found` for direct verification |
| P19 | OpenWrt SQM/CAKE cross-check | OpenWrt SQM defaults & CAKE `cake-*` tc options | — | `unverified — primary source not found` (Anubis block, §0) |

## 5. Gap analysis

Each gap is: where, what breaks, primary-source consequence, fix direction.

- **G1 — “RouterOS has no fq_codel” is false against R2/R3.** `RouterOS` fq_codel branch emits `kind=sfq sfq-perturb=10` with a 3-line disclaimer (`pkg/remediate/remediate.go:300-310`, mirrored ingress `366-372`). R2 explicitly lists FQ-Codel as a RouterOS queue kind since 7.1beta3 with properties `fq-codel-interval/target/flows/limit/quantum/ecn/memlimit/ce-threshold`, and R3 lists FQ_CoDel with “Excellent latency management and fair distribution of bandwidth among flows. Generally considered a good default choice.” R4 describes the algorithm the kind implements. The generated script therefore substitutes a strictly weaker discipline (R3: SFQ “does not manage latency or congestion”) while claiming honesty — the honesty is about the wrong fact. Fix: emit `kind=fq-codel` with R2 defaults (or R4 defaults where R2 is silent), keep the disclaimer only where the property spelling is unverified, and re-point `TestRouterOSFqCodelIsHonest` at the new warning text. This is the single largest correctness gap in the slice.
- **G2 — `TestRouterOSEmitsOnlyValidKinds` now forbids the correct kind.** The regression gate “Aucune kind PCQ, aucun /interface queue add” (`pkg/remediate/remediate_test.go:249-297`) pins `kind=(cake|sfq|fifo)` (`270`, `290`); after G1 it must admit `fq-codel`. If G1 is fixed without touching this regex, the suite goes red — proof the test encodes the stale claim, not a vendor truth.
- **G3 — Version gate ≥ 7.10 vs R2's 7.1beta3.** Both script headers (`pkg/remediate/remediate.go:281,347`) print “Requires RouterOS >= 7.10 for kind=cake”. R2 says CoDel/FQ-Codel/CAKE exist since 7.1beta3. Either the header is over-strict (refusing valid 7.1–7.9 devices) or some `cake-*` property used (`cake-ack-filter=filter`, `cake-wash=yes`, `cake-nat=yes`) appeared in 7.10 — neither is verifiable from R1/R2/R3 alone. Mark `unverified — primary source not found`, or fetch the RouterOS changelog / property-level page before the header ships.
- **G4 — `linux-namespace` level is a label, not a namespace attach.** `BuildQEMUArgs` flips `tcg`→`kvm` for `linux-namespace` (`cmd/cgo-chr/main.go:25-27`) and `levelAllowed` allows it only on Linux (`cmd/cgo-chr/run.go:161-163`), but nothing attaches CHR interfaces to Linux namespaces, nothing wires the bench's `ns_client`/`e-srv` namespaces, and the proof's `Level`/`Accelerator` fields are the only record (`pkg/audit/routeros_proof.go:32-33`). `kit/chr/README.md:5-6` promises “plus les interfaces CHR attachées aux namespaces du banc” — unimplemented. `docs/VERIFY.md:57` correctly states “the Linux namespace attach is the second level and never substitutes for it”. Fix: either implement the attach or rename the level value and README claim before extraction; a stolen `linux-namespace` proof today proves only KVM boot.
- **G5 — `imageHash` reads the whole image; `chr-image.sha256` is never checked.** `imageHash` does `os.ReadFile(path)` then `FileSHA256` (`cmd/cgo-chr/run.go:137-143`) — for a multi-GB CHR image this is unbounded memory; and the pin file (`kit/chr/chr-image.sha256:1-4`) is never read by any Go code (grep: no reference outside comments). A validator that seals the image hash it just booted but cannot refuse a swapped image is not validating the image. Fix: streaming hash + compare against the pin file before boot, exit 2 on mismatch.
- **G6 — `Validate` forbids negative overhead that R1/R2 allow.** R1: manual overhead “within a range of −64 to 256 bytes”, R2 `cake-overhead (-64 … 256)`; `Validate` rejects any `Overhead < 0` (`pkg/remediate/remediate.go:97-99`) and Linux emits overhead only when `> 0` (`163-165`, `236-238`). RouterOS emission ignores `Overhead` entirely (no `cake-overhead` line in `remediate.go:291-298`, `358-363`). Fix: admit −64…0 and emit `cake-overhead=<n>` / `overhead <n>` even when negative, with a vendor-range test.
- **G7 — Serial driver framing is `unverified — primary source not found` end-to-end.** Prompt regexes, `admin`/empty-password login, `/export terse`, `failure:` error matching, byte-by-byte `expect` with per-call `time.After` (`cmd/cgo-chr/serial.go:12-121`) are pinned only by `fakeCHR` (`cmd/cgo-chr/serial_test.go:12-101`). No real transcript exists; `docs/VERIFY.md:57` records no CHR boot. R1/R2/R3 say nothing about console text. Fix: keep the injectable `Transport` seam (it is the right design), but the anti-toy bar (§8) requires one recorded real-console transcript before any acceptance claim.
- **G8 — Concatenation of upload + ingress scripts can interleave cleanup.** `run` concatenates both scripts then `ParseRouterOSConfig` parses the union (`cmd/cgo-chr/run.go:41-42`); upload cleanup removes `^cgo-` types and `cgo-managed` trees, ingress removes `^cgoingress-` / `cgo-managed-ingress` (`pkg/remediate/remediate.go:287-288,352-353`). Disjointness is pinned in-process (`pkg/remediate/ingress_test.go:124-136`) but across a console paste the second script's cleanup runs after the first script's adds — order is safe today only because names are disjoint. Any future shared helper object would break the invariant silently. Fix: document the disjointness contract in the validator's output schema, or emit a single combined script with one cleanup block.
- **G9 — `fakeCHR` `/export` fixture is not `/export terse` output.** The fake answers `/export` with a one-line `add …` under `/queue type` + `/queue tree` headers (`cmd/cgo-chr/serial_test.go:59`); the real `kit/chr/export-fixture.txt` is the same shape (`kit/chr/export-fixture.txt:2-9`). Whether RouterOS `/export terse` actually prints that shape is `unverified — primary source not found` (Export page 404). `ParseRouterOSConfig` handles both one-line and header forms (`pkg/audit/routeros_config.go:41-65`), so a drift in exporter shape is a latent parser risk with no golden transcript to catch it.
- **G10 — CHR BLOCKED is the honest status; extraction must not launder it.** `docs/VERIFY.md:57` is explicit: 12/12 replay checks + sealed proof are unit-tested, but “Nothing has booted CHR … so no RouterOS acceptance exists”. The plan index marks the slice Deferred (`docs/superpowers/plans/README.md:44`). Any standalone `cgo-chr-validate` README, tag, or release note that implies RouterOS compliance without a `windows-tcg` `proof.json` (schema `cgo.chr-proof/1`, `pkg/audit/routeros_proof.go:58`) would be a false claim. The module must ship with this row quoted.
- **G11 — `TestRunReplayMutatedExportFails` asserts exit code but not the named failure.** The test docstring says “l'échec doit nommer la propriété fautive” but the body only checks `code == 1` and deserialises checks without asserting `!Pass` on `cake-ack-filter` (`cmd/cgo-chr/run_test.go:109-133`) — unlike its sibling `TestRunReplayRejectsWrongIngressParent` which does (`194-202`). Fix: mirror the sibling's named-failure assertion before extraction.
- **G12 — Two provenances of “default LAN” collide.** `DefaultRouterOSLan = "bridge"` (`pkg/remediate/remediate.go:56`) vs. the report/UI default `firstSet(r.RouterOSLanIface, "bridge")` (`cmd/cgo/tui_audit.go:368`) vs. slice-24 spec naming `/queue tree add … parent=bridge queue=cgoingress-cake` (`docs/superpowers/specs/2026-11-04-cgo-slice-24-operator-surface-truth-design.md:23`). No device has ever been observed with this interface (`docs/VERIFY.md:64`). The validator must treat `bridge` as an operator-supplied assumption, never a discovered value — the French `ATTENTION` in the script (`pkg/remediate/remediate.go:340-342`) is the minimum; the validator should echo `routeros_lan_assumed` into the proof.

## 6. Target tool — `cgo-chr-validate` as independent repo

### 6.1 Identity and contents

`module github.com/Realms4239/cgo-chr-validate` (versioning per P18 / R5/R9/R10 as cited in §4 — `unverified — primary source not found` for direct fetch). Contents: `cmd/cgo-chr-validate` (renamed from `cmd/cgo-chr`, keeping CLI flags `cmd/cgo-chr/main.go:42-50` and `BuildQEMUArgs` `main.go:23-39`) + `pkg/audit` CHR subset + `pkg/remediate` RouterOS generators as a *vendored, re-exported* consistency oracle (§6.3). Deps: stdlib only — matches the sibling prescription “stdlib only” (`docs/research/2026-11-07-module-cli-frontend.md:236`). Preserves the “not shipped in the product binary” posture (`cmd/cgo-chr/main.go:1-2`): the validator is a separate binary + OCI sidecar (QEMU + CHR image as separate layer), per sibling R3 analysis (`docs/research/2026-11-07-module-cli-frontend.md:180`).

### 6.2 Target Go API

```go
package chrvalidate

type Report struct {
    AuditID     string             `json:"audit_id"` // cmd/cgo-chr/run.go:14-17
    Remediation *Remediation       `json:"remediation"`
}
type Remediation struct {
    RouterOS        string `json:"routeros"`
    RouterOSIngress string `json:"routeros_ingress"`
}

// Parser/diff/proof — pkg/audit/routeros_*.go, unchanged signatures:
func ParseRouterOSConfig(text string) RouterOSConfig          // routeros_config.go:33
func CompareRouterOS(expected, found RouterOSConfig) []RouterOSCheck // routeros_compare.go:22
func WriteCHRProof(dir string, p CHRProof, export, transcript []byte) (string, error) // routeros_proof.go:48

// Orchestration — cmd/cgo-chr/run.go:19, cmd/cgo-chr/main.go:23:
func BuildQEMUArgs(c RunConfig) []string
func Run(cfg RunConfig, report *Report, replay []byte, outDir string) (proofPath string, exit int)
type RunConfig struct{ QEMU, Image, Level string; Timeout time.Duration } // main.go:14-19

// Generation oracle (from pkg/remediate, importable for consistency checks only):
func RouterOS(l Link, qdisc string) (string, error)        // remediate.go:268
func RouterOSIngress(l Link, qdisc string) (string, error) // remediate.go:330
```

`Level` stays `"windows-tcg" | "linux-namespace"` (gated at `pkg/audit/routeros_proof.go:52-54` and `cmd/cgo-chr/run.go:157-165`) until G4 is implemented; the validator refuses to emit `Level: "linux-namespace"` unless a namespace-attach step actually ran (§8).

### 6.3 What does not move

`pkg/remediate` Linux generators (`LinuxTC`/`LinuxTCIngress`, `pkg/remediate/remediate.go:135-260`) remain in the main repo — the validator needs only the RouterOS strings as a consistency oracle (regenerate from `Link` and diff against the report's embedded script). `pkg/campagne`, `pkg/api`, `web/frontend` stay in the monolith. Extraction is `cmd/cgo-chr` + `pkg/audit` CHR subset + `kit/chr/` wrapper/fixtures, not the callers (sibling rule, `docs/research/2026-11-07-module-qdisc-remediate.md:233`).

### 6.4 Repository layout

```
cgo-chr-validate/
  go.mod                          # module github.com/Realms4239/cgo-chr-validate, go 1.25
  cmd/cgo-chr-validate/*.go       # main.go, run.go, serial.go, transport.go + tests
  pkg/chrvalidate/                # routeros_config.go, routeros_compare.go, routeros_proof.go
  pkg/remediateoracle/            # vendored remediate.go RouterOS-only subset (or require+replace)
  kit/chr/                        # setup-mikrotik-chr.sh (repointed), export-*.txt fixtures,
                                  # README.md (human prerequisites), chr-image.sha256 (enforced, G5)
  docs/VERIFY-IMPORT.md           # quotes docs/VERIFY.md:57 CHR BLOCKED row verbatim until a proof lands
```

## 7. Slices

Each slice leaves `go test ./...` green in its own repo; ordered so the vendor-truth fix comes before the extraction, and the extraction before any acceptance claim.

| # | Slice | Scope | Gate |
|---|---|---|---|
| S1 | Vendor-truth fix (G1/G2/G3/G6) in `pkg/remediate` **in the monolith first**: replace `sfq` fallback with `kind=fq-codel` (R2 property names; spelling `unverified — primary source not found` until confirmed), admit `fq-codel` in `TestRouterOSEmitsOnlyValidKinds` regex, emit `cake-overhead` for negative overhead, keep the ≥ 7.10 header only with a changelog citation or drop to “≥ 7.1beta3 (R2)” | `go test ./pkg/remediate/...` green; string grep for `no fq_codel` inside generated scripts returns only the new warning (if any); header version cites R2 |
| S2 | Strengthen validator tests (G11) + pin `chr-image.sha256` (G5): `TestRunReplayMutatedExportFails` asserts the named `cake-ack-filter` failure (mirror `TestRunReplayRejectsWrongIngressParent` `run_test.go:194-202`); `imageHash` becomes streaming + refuses a mismatched pin before boot | `go test ./cmd/cgo-chr/...` green; mutation test fails if the named check is unnamed (publish the test's expected output) |
| S3 | Extract `cgo-chr-validate` repo: move `cmd/cgo-chr` + `pkg/audit/routeros_*.go` (+ `Remediation` JSON shape) + `kit/chr/`; repoint `setup-mikrotik-chr.sh` to `go run ./cmd/cgo-chr-validate` or publish a binary; delete from monolith only after the new repo's replay tests pass standalone | `go list -m all` in `cgo-chr-validate` shows no `fyne/bubbletea/lipgloss/x-sys`; `go test ./...` green in the new repo with `GOCACHE` clean; `--dry-run` and `--export-file` replay both green |
| S4 | Real `windows-tcg` acceptance: human installs QEMU + downloads CHR (`kit/chr/README.md:10-14`), places image in `kit/chr/cache/`, fills `chr-image.sha256`, runs `setup-mikrotik-chr.sh --report <sealed audit_report.json> --image <chr.img> --level windows-tcg`; record `data/chr/<audit_id>/proof.json` path + `ExportSHA256`/`TranscriptSHA256` into `docs/VERIFY.md` CHR row | `proof.json` schema `cgo.chr-proof/1`, `Import.Accepted=true`, all `Checks[].Pass=true`; `docs/VERIFY.md:57` row moves BLOCKED→VERIFIED with command+host+date quoted; transcript pinned as the G7 golden fixture |
| S5 | `linux-namespace` implementation or rename (G4): either attach CHR interfaces to the bench namespaces (real `ip netns` wiring + drive + re-export) or remove the value and the `kit/chr/README.md:5-6` promise | if implemented: proof `Level=linux-namespace` + namespace-attach log in transcript; if renamed: `levelAllowed` rejects the old value with a migration note |
| S6 | Release: tag `v0.1.0` (unstable per R9 as cited §4), publish binary + OCI sidecar, quote the VERIFY CHR row in the release notes | `go vet ./...`, `gofmt -l` clean; release notes contain the VERIFY row or the new proof hash — never a bare “RouterOS validated” |

## 8. Anti-toy criteria

A standalone `cgo-chr-validate` claim of “RouterOS 7 accepted the CGO script” is not credible unless **all** hold. Stricter than the current host tests (which run against `fakeCHR` + fixtures).

1. **A real CHR boot, with the transcript as evidence.** At least one `windows-tcg` run on a QEMU host with a licensed CHR image (`kit/chr/README.md:12-14`) produces `data/chr/<audit_id>/proof.json` (`cgo.chr-proof/1`), `export.txt`, `transcript.txt` with matching `ExportSHA256`/`TranscriptSHA256` (`pkg/audit/routeros_proof.go:58-60`), and the transcript shows the real `login:`/`[admin@…] >` console (`cmd/cgo-chr/serial.go:13-17`) — not `fakeCHR` (`cmd/cgo-chr/serial_test.go:16-73`). Until then the VERIFY row stays BLOCKED (`docs/VERIFY.md:57`) and every README/release line carries `unverified — primary source not found` for acceptance.
2. **Vendor truth, not in-repo comments.** The generated scripts match R1/R2/R3 for every `cake-*`/`fq-codel-*`/`sfq-*` property they emit (§4 P1–P16); the “no fq_codel” claim is gone (G1) or carries a fetched vendor citation; the RouterOS version gate cites a changelog, not `remediate.go:281` alone; any property whose spelling/value is not in R1/R2/R3 is listed as `unverified — primary source not found` in the README's property table.
3. **Failure is named, never silent.** A mutated export fails exactly one named `RouterOSCheck` (WAN `cake-ack-filter`, LAN `cgoingress-lan-in.parent` — `cmd/cgo-chr/run_test.go:109-203`, `pkg/audit/routeros_compare.go:22-64`); a duplicate intended object fails ( `routeros_compare.go:26-33`); a refused import seals `Import.Accepted=false` with the verbatim RouterOS error line (`cmd/cgo-chr/run.go:69-78`, `serial.go:58-60`), and the proof says “refusé”, never “validée” (`kit/chr/README.md:28-29`).
4. **The image is the image.** The booted image's streaming SHA-256 equals `chr-image.sha256`'s pinned line before QEMU starts (G5); QEMU is absent → exit 2 with the `kit/chr/README.md` pointer, never a fabricated pass (`cmd/cgo-chr/run.go:109-112`).
5. **Levels mean what they say.** `windows-tcg` proves import+export+diff only (`kit/chr/README.md:4-6`); `linux-namespace` is not emitted as a proof level until a namespace attach actually runs (G4); neither level claims débit/latence/production change (`kit/chr/README.md:38-41`).
6. **Nothing unmeasured is validated.** The validator refuses a report without `audit_id` or without any RouterOS script, exit 2 (`cmd/cgo-chr/run.go:30-37`, `TestRunRejectsReportWithoutScript`); the report's embedded script must regenerate byte-identically from its measured `Link` via the oracle (§6.2) — a hand-edited script is a mismatch, not a pass; the assumed-LAN warning (`routeros_lan_assumed`, `pkg/remediate/remediate.go:339-343`) is echoed into the proof (G12).
7. **Independence is real.** `go list -m all` in `cgo-chr-validate` shows only stdlib (+ the module's own packages) — no `fyne`, `bubbletea`, `lipgloss`, `x/sys` inheritance from `go.mod:5-10`; the repo builds and tests with the monolith absent (S3 gate); the CLI still accepts `--export-file` replay and `--dry-run` without QEMU or an image.
8. **Version numbers mean what they say.** `v0.x` carries no stability promise (P18 / R9 as cited §4); a release that changes emitted RouterOS strings is at least a minor version even if the Go API is unchanged, because the generated config is part of the public API (sibling rule `docs/research/2026-11-07-module-qdisc-remediate.md:260`); the CHR BLOCKED row (`docs/VERIFY.md:57`) is quoted in every release until criterion 1 lands.

---
*What was not checked: No CHR image was booted, no QEMU run exists, and OpenWrt SQM / CHR `/export` vendor pages were unreachable (404 / Anubis block, §0), so console/grammar claims remain `unverified — primary source not found`; the extraction was not performed (research only, per brief).*
