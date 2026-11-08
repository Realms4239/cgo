# Module kit / vm / profile — deploy & testbench (research)

Date: 2026-11-07. Scope: `internal/kit`, `internal/vm`, `pkg/profile`,
`kit/*.sh`, `docs/deploy.md`, `docs/VERIFY.md` live-blocker rows.
Method: read source, map exported surface, check against primary sources
below. No code, no commit.

## Primary sources (fetched 2026-11-07)

- [L1] Go 1.4 release notes, "Internal packages" — https://go.dev/doc/go1.4#internal
  ("place it in a directory named `internal` … may not be imported by packages
  outside the source subtree … `.../a/b/c/internal/d/e/f` can be imported only
  by code in the directory tree rooted at `.../a/b/c`").
- [L2] Go Modules Reference — https://go.dev/ref/mod (module = "collection of
  packages that are released, versioned, and distributed together"; version
  starts with `v` + SemVer; module path is the prefix for package paths).
- [L3] Semantic Versioning 2.0.0 — https://semver.org/ (rule 1 MUST declare a
  public API; rule 5 "1.0.0 defines the public API"; rule 3 version contents
  MUST NOT be modified after release).
- [L4] QEMU documentation — https://www.qemu.org/docs/master/ + Invocation
  (`qemu-system-x86_64 [options] [disk_image]`; `-machine accel=kvm|…|tcg`,
  `-accel name`, `-smp`, `-m`, `-netdev`/`-device` etc.).
- [L5] Linux Kernel KVM docs — https://www.kernel.org/doc/html/latest/virt/kvm/index.html
  (KVM API, devices, architectures; index page only — no guest-image claim).
- [L6] Ansible documentation — https://docs.ansible.com/ansible/latest/
  (inventory, playbooks, collections, modules/plugins index pages visible).
- [L7] Oracle VirtualBox Manual, Ch.8 VBoxManage —
  https://www.virtualbox.org/manual/ch08.html (`startvm`, `controlvm`,
  `modifyvm --nic1`, `natpf`, `guestproperty`, `snapshot`, `list runningvms`,
  `showvminfo --machinereadable`, `guestcontrol`, `natnetwork`).
- [L8] VMware `vmrun` — `unverified — primary source not found` (Broadcom
  TechDocs portal returned no `vmrun` manual on 2026-11-07; all `vmrun`
  behaviour below is therefore local-source evidence only, not primary).

Local claims cite `path:line`. Where no primary source exists the claim is
marked `unverified — primary source not found`.

## 1. Responsibility

### 1.1 `internal/kit` — deployment engine (operator pilot)

- `internal/kit/kit.go:1-6` declares: "moteur de déploiement en Go : les 11
  actions de l'ancien engine.sh … mêmes codes de sortie (2 usage/déps/build,
  3 scan/pick, 4 hyperviseur absent, 5 timeout SSH, 6 cross-compile/bootstrap,
  7 scp, 8 install/logs) et les mêmes variables d'environnement CGO_*".
- The CLI surface is 30+ actions (counted in `cmd/cgo/kit.go:23-59`, dispatch
  `cmd/cgo/kit.go:94-214`): `doctor, scan, keysetup, ensure, align, build,
  deploy, svc, testbed, dns, tls, shipcheck, package, bootstrap, status, logs,
  tunnel, snapshot, revert, snapshots, ssh, netinfo, nic, vnet, next, guest,
  readme, tui, ps, backup, verify, health, schedule`; exit codes 2–8
  documented at `cmd/cgo/kit.go:58-59`, propagated from `Runner` methods e.g.
  `Build` (`internal/kit/kit.go:1327-1356`), `Deploy`
  (`internal/kit/kit.go:1372-1397`). This is a CLI contract, not a Go API
  ([L3] rule 1: a public API must be declared — here it is declared as CLI
  actions + exit codes, not Go symbols).
- Per-file responsibilities (comments cite the file header of each):
  - `internal/kit/kit.go` — `Config`, YAML+env loading, SSH/SCP/health
    primitives, `Doctor`, `Scan`, `Ensure`, `Build`, `Deploy`,
    `Bootstrap`, `Status`, `Logs`, `Align`, `Tunnel` (`kit.go:822,910,982,
    1327,1372,1546,1562,1590,1601,1660`).
  - `internal/kit/actions2.go` — snapshots/revert, interactive SSH, `ps`,
    `svc` start/stop/restart/status, `backup` of `data/runs`, `verify`,
    `health` (`actions2.go:26,73,131,152,187,314,397,461,476`).
  - `internal/kit/keysetup.go` — idempotent SSH key install via password
    prompt, never stored (`keysetup.go:13-22`).
  - `internal/kit/guest.go` — guest prep via `guest-setup.sh` scp+ssh
    (`guest.go:9-15`).
  - `internal/kit/testbed.go` — guest measurement bench up/check/down via
    `testbed.sh` (`testbed.go:7-12`).
  - `internal/kit/next.go` — 11-step pipeline checklist `GatherNextVer`
    (`next.go:18-22,44-273`); `internal/kit/shipcheck.go` — DNS→TCP→TLS→
    cert→health→OS-trust release gate (`shipcheck.go:12-17`).
  - `internal/kit/package.go` — operator zip/tar packaging, binary-first
    (`package.go:15-20,21-34`); `internal/kit/schedule.go` — systemd timer +
    cron fallback for nightly campaigns (`schedule.go:8-12`).
  - `internal/kit/vnet.go` — host-side network doctor/repair with hard
    guards (never write 10.0.2.0/24 to a host adapter, never overwrite a
    healthy /24) (`vnet.go:14-22,245-248,306-308`).
  - `internal/kit/dns.go:15`, `tls.go:19`, `netinfo.go:10`, `netnic.go:11`,
    `exec_*.go` — DNS mapping, OS trust store, guest net read-only, NIC
    mode, process-execution shims.

### 1.2 `internal/vm` — hypervisor drivers

- `internal/vm/vm.go:1-5`: "détection et pilotage des hyperviseurs, pur Go
  multi-OS … cherche .vmx/.vbox". Two drivers only: VMware (`vmrun`) and
  VirtualBox (`VBoxManage`) — `Detect()` at `vm.go:835-852` searches
  `vmrun[.exe]` then `VBoxManage[.exe]` via `candidates()` at
  `vm.go:794-817` (Windows `C:\Program Files\VMware\VMware Workstation`,
  `C:\Program Files\Oracle\VirtualBox`; Linux `/usr/bin`, `/usr/local/bin`;
  macOS Fusion paths).
- `Hypervisor` interface at `vm.go:22-45`: `Name, Exe, Running, Start,
  StartGUI, Stop, GuestIP, NetMode, SetNetMode` — every method has an
  external side effect (start/stop a VM, rewrite `.vmx`, `modifyvm`). It is
  not a query library.
- VirtualBox NAT guest space `10.0.2.x` is unreachable from host by
  construction; the fix is a host port-forward (`EnsureNatSSH`,
  `vm.go:546-642`, rules `cgo-ssh,tcp,,2222,,22` and
  `cgo-dashboard,tcp,,9090,,9090` at `vm.go:614,617`). VMware NAT keeps a
  guest IP reachable via vmnet8 (`vm.go:150-162` reads
  `ethernet0.connectionType`).
- Disk scan: `ScanVMs`/`ScanVMsProgress` at `vm.go:931-1008` (hypervisor
  inventory first, then `<drive>/VMs`, `~/VirtualBox VMs`, shallow root scan,
  deep scan; `extraRoots()` at `vm.go:897-926`).
- VMware `vmrun` subcommands used: `list`, `start … nogui`,
  `-T ws start … gui`, `stop … soft|hard`, `getGuestIPAddress … -wait 5`
  (`vm.go:77-125`). VMware behaviour itself: `unverified — primary source
  not found` [L8]. VirtualBox subcommands used (`list runningvms`,
  `showvminfo --machinereadable`, `guestproperty get`, `controlvm natpf1`,
  `modifyvm --natpf1/--nic1`, `snapshot … take/restore` in `actions2.go`)
  match [L7] Ch.8 (`startvm`, `controlvm natpf<N>`, `modifyvm --nic1`,
  `guestproperty get`, `snapshot take/restore`).

### 1.3 `pkg/profile` — profile registry persistence

- `pkg/profile/profile.go:1-12`: single file, const `file = "data/profiles.json"`,
  imports only `pkg/model` (`profile.go:9`). `Reserved` guards P1–P4
  (`profile.go:14-20`, rationale: a file "P3" would mask the reference VSAT
  profile in live card and dashboard campaigns).
- Built-ins live in `pkg/model/types.go:43-51` (P1 80 Mb/s, P2 20/100 ms,
  P3 5/600/30/1, P4 100/40/20/0.3) with `ProfilesMu` guard at
  `types.go:53-57`. `profile.Import` (`profile.go:23-46`) writes through
  `model.ProfilesMu` to memory + `data/profiles.json`; `Load`
  (`profile.go:50-67`) merges file over memory skipping reserved IDs;
  `Delete` (`profile.go:71-93`) refuses reserved/unknown IDs.
- Callers: `pkg/api/server.go:27`, `pkg/api/server_audit.go:22`,
  `pkg/audit/match.go:7`, `cmd/cgo/calibrate.go:15` (grep `pkg/profile`).
  So `pkg/profile` is importable by consumers inside and outside the module
  ([L1] internal rule does not apply — it is under `pkg/`, not `internal/`);
  it is the only one of the three that is a library by shape.

### 1.4 `kit/*.sh` — guest/bench shell layer

| Script | Role (header) |
|---|---|
| `kit/engine.sh` | Compat shim: `--action X` → `cgo kit X` (`engine.sh:2-4`) |
| `kit/config.sh` | Machine-local YAML loader, same keys as `kit.go` `Config` (`config.sh:2-5`) |
| `kit/testbed.sh` | On-VM bench: veth + netns `cgo-srv` + `cgo testbedsrv` :8081/:5201, `up\|down\|check` (`testbed.sh:2-9`) |
| `kit/cgo-bench-dualns.sh` | Decoupled bench: link physics vs queue policy, netns topology (`cgo-bench-dualns.sh:2-9`) |
| `kit/run-dualns.sh` | Production path: hardened bench + sink + dashboard in `ns_client` on `10.99.0.2:9090` (`run-dualns.sh:2-9`) |
| `kit/guest-setup.sh` | Manual on-VM prep, idempotent, `--check` read-only (`guest-setup.sh:2-11`) |
| `kit/vm-install.sh` | On-VM installer: validate `cgo-linux.new`, install, health-check (`vm-install.sh:2-7`) |
| `kit/install.sh` | Linux dep bootstrap, never switches default CC (`install.sh:2-6`) |
| `kit/install-bbr3.sh` | Opt-in BBRv3 full-kernel build, never called by kit (`install-bbr3.sh:2-9`) |
| `kit/host-readiness.sh` | Read-only host doctor, `--require-linux` exits 2 on Windows (`host-readiness.sh:2-8`) |
| `kit/stress_testbed.sh` | Adversarial bench check: offloads OFF, traversal, teardown (`stress_testbed.sh:2-10`) |
| `kit/test-kit.sh`, `test-hypervisor.sh`, `test-overhaul.sh`, `test-fresh-clone.sh` | Static gates: grep `check-bundle`/`vitest` in `kit.go`, `virtualbox`/`vmware` in `vm.go`, no `powershell` in `vm.go` (`test-hypervisor.sh:5-7`, `test-kit.sh:5-8`) |
| `kit/chr/setup-mikrotik-chr.sh` | Thin wrapper → `go run ./cmd/cgo-chr`, "Toute la logique est en Go" (`setup-mikrotik-chr.sh:2-5`) |
| `kit/fix-vnet-admin.ps1`, `host-tunnel.ps1`, `setup-meteolink-dev.*` | Operator Windows helpers (called by `vnet.go:265-291` `findKitFile`) |

`docs/deploy.md:106-128` states the contract: host→VM deployment is driven
by `cgo kit` (Go engine, same exit codes as legacy bash; `engine.sh` is a
shim), binary-first and idempotent, SPA embedded, VM needs only
`iproute2 curl bc` + `tcp_bbr`.

## 2. Exported API (file:line)

### 2.1 `internal/vm` (`internal/vm/vm.go`)

| Symbol | Line |
|---|---|
| `type Hypervisor interface` | `vm.go:22` |
| `type NatForwarder interface` | `vm.go:49` |
| `func IsRunning(h Hypervisor, vbx string) bool` | `vm.go:318` |
| `func SameVM(a, b string) bool` | `vm.go:696` |
| `func Detect() []Hypervisor` | `vm.go:835` |
| `func Primary() Hypervisor` | `vm.go:855` |
| `func ScanVMs(deep bool) []string` | `vm.go:931` |
| `func ScanVMsProgress(deep bool, fn func(string)) []string` | `vm.go:937` |
| `func HypForPath(p string) string` | `vm.go:1015` |
| `func Pick(vms []string, name string) string` | `vm.go:1042` |
| `func (v *virtualbox) EnsureRegistered` (exported method, unexported type) | `vm.go:433` |

`vmware`/`virtualbox` implementations are unexported (`vm.go:59,215`);
`RunCtx` boxing, `GuestIP`, `EnsureNatSSH`, `NetMode`/`SetNetMode` are
interface methods, not free functions.

### 2.2 `internal/kit`

Functions/consts: `KitVersion` `kit.go:52`; `LoadConfig` `kit.go:74`;
`LoadConfigFast` `kit.go:81`; `AutoLockSingle` `kit.go:252`;
`LinearGuide` `kit.go:285`; `SaveSSHTarget` `kit.go:330`;
`SaveVMX` `kit.go:346`; `NewRunner` `kit.go:374`;
`NATForwardTarget` `kit.go:562`; `BgCmd/BgCmdCtx` `exec_export.go:10,15`;
`ScheduleUnits` `schedule.go:13`; `ScheduleCronLine` `schedule.go:101`;
`Readme` `package.go:220`; `NextStep` type `next.go:23`;
`GatherNext` `next.go:37`; `GatherNextVer` `next.go:44`;
`FirstOpen` `next.go:277`.

Types/methods: `Config` `kit.go:32` (+ `VMPath` `kit.go:58`, `SSH`
`kit.go:469`, `SSHUp` `kit.go:521`, `SCP` `kit.go:663`, `SCPOut` `kit.go:682`,
`SCPOutErr` `kit.go:707`, `Health` `kit.go:752`); `Runner` `kit.go:368` (+
`Doctor` `kit.go:822`, `Scan` `kit.go:910`, `Ensure` `kit.go:982`,
`Build` `kit.go:1327`, `Deploy` `kit.go:1372`, `Bootstrap` `kit.go:1546`,
`Status` `kit.go:1562`, `Logs` `kit.go:1590`, `Align` `kit.go:1601`,
`Tunnel` `kit.go:1660`, `Snapshot/Revert/SSHInteractive/Ps/Svc/Backup/
Snapshots/Verify/Health` `actions2.go:26,73,131,152,187,314,397,461,476`,
`DNS` `dns.go:15`, `TLS` `tls.go:19`, `KeySetup` `keysetup.go:23`,
`Guest` `guest.go:16`, `Testbed` `testbed.go:13`, `Package` `package.go:21`,
`Schedule` `schedule.go:54`, `Next/NextVer` `next.go:359,364`, `Netinfo`
`netinfo.go:10`, `Nic` `netnic.go:11`, `Vnet` `vnet.go:23`, `ShipCheck`
`shipcheck.go:18`).

### 2.3 `pkg/profile`

`Reserved` `profile.go:18`; `Import` `profile.go:23`; `Load` `profile.go:50`;
`Delete` `profile.go:71`. `file` (`profile.go:12`) is unexported.

## 3. Dependencies

- Module root: `module github.com/Realms4239/cgo`, `go 1.25`
  (`go.mod:1-3`) [L2] (module path = prefix for all package paths, so
  `internal/kit` is `github.com/Realms4239/cgo/internal/kit`).
- `internal/kit` imports: stdlib + `github.com/Realms4239/cgo/internal/vm`
  (`kit.go:26`, `next.go:15`, `guest.go:6`, `actions2.go:16`, `vnet.go:11`)
  + `golang.org/x/sys/windows` on Windows only (`exec_windows.go:9`).
  Callers: `cmd/cgo/kit.go:9`, `cmd/cgo/main.go:28`,
  `cmd/cgo/setup.go:12`, `cmd/cgo/kittui*.go`, `cmd/cgo-gui/*.go`
  (grep `internal/kit` — 6+ call sites).
- `internal/vm` imports: stdlib only + `golang.org/x/sys/windows`
  (`drives_windows.go:5`, `execcmd_windows.go:9`). No external hypervisor
  SDK; all control is subprocess (`vmrun`, `VBoxManage`, `arp`, `ip`, `ping`)
  e.g. `vm.go:64-75,220-231,1183-1266`.
- `pkg/profile` imports: stdlib + `github.com/Realms4239/cgo/pkg/model`
  (`profile.go:9`); `go.mod:5-10` shows `x/sys v0.36.0` as the only direct
  `require` besides Fyne/BubbleTea (used by `cmd/cgo-gui`/`cmd/cgo`, not by
  these three modules except `x/sys/windows` build-tagged files).
- File-IO deps: `kit/cgo-vm.yaml` (machine-local, gitignored per
  `cgo-vm.yaml.example:1`) via hand-rolled parser `kit.go:100-212` (no
  `gopkg.in/yaml.v3` import in `kit` — `yaml.v3` is indirect in `go.mod:59`);
  `data/profiles.json` (`profile.go:12`); `kit/*.sh` located at runtime by
  `findKitFile` (`vnet.go:345-358`).

## 4. Primary-source requirements

- [L1] `internal/` may not be imported outside the subtree rooted at the
  parent of `internal/` — so `internal/kit`+`internal/vm` are usable only
  by packages under `github.com/Realms4239/cgo/` (tested: `cmd/cgo`,
  `cmd/cgo-gui` import them; nothing outside that module can). Promoting
  kit/vm to a publishable Go module **requires** moving them out of
  `internal/` — and then every host-side side effect becomes a public API
  commitment ([L3] rule 5).
- [L2] A module is "released, versioned, and distributed together"; the
  `go` directive (`go 1.25` at `go.mod:3`) sets the minimum Go toolchain
  and is enforced since Go 1.21. Any split must keep this floor and the
  `require` graph consistent ([L2] `require`/`go` directive sections).
- [L3] SemVer requires a declared public API (rule 1) and forbids modifying
  released version contents (rule 3). The current de-facto public APIs are
  (a) CLI actions + exit codes 2–8 (`cmd/cgo/kit.go:23-59`), (b) `CGO_*`
  env vars + `kit/cgo-vm.yaml` keys (`kit.go:29-47,74-212`), (c)
  `KitVersion = "1.3.5"` (`kit.go:52`) compared by GUIs against the remote
  dashboard (`kit.go:49-51`). Any library extraction that changes the Go
  signatures without the CLI/YAML/env contract breaks operators.
- [L4] QEMU/KVM [L5]: the current `vm` package has **no** QEMU/KVM driver.
  [L4] Invocation shows the canonical interface is `qemu-system-x86_64`
  with `-machine accel=kvm`, `-m`, `-smp`, `-netdev/-device` — none of
  which appears in `vm.go` (grep `qemu` = 0 in `internal/vm`). A "testbench
  module" claiming KVM/QEMU support must add a `Hypervisor` implementation
  driving QEMU (QMP/monitor + `qemu-img`) — `unverified — primary source
  not found` for any narrower claim about KVM ioctl-level control from Go
  (userspace QEMU is the documented path).
- [L6] Ansible: `guest-setup.sh`/`testbed.sh`/`vm-install.sh` are ad-hoc
  idempotent shell; Ansible's documented unit is inventory + playbook +
  module with return values ([L6] playbook/inventory/module index pages).
  Mapping kit actions to Ansible modules/collections is a design gap:
  no `ansible/` directory, no inventory, no playbook exists in the repo
  (glob `**/*.yml` outside `kit/cgo-vm.yaml.example` is empty).
- [L7] VirtualBox control is via `VBoxManage` subcommands (`startvm`,
  `controlvm`, `modifyvm`, `natpf`, `guestproperty`, `snapshot`);
  `vm.go`'s usage matches [L7] Ch.8, including `--machinereadable` parsing
  (`vm.go:248-269,484-544`). VMware `vmrun` usage: `unverified — primary
  source not found` [L8].

## 5. Gap analysis

### 5.1 Tool vs library (structural)

| Fact | Evidence |
|---|---|
| kit/vm live under `internal/` → unimportable outside module by Go rule | [L1]; `internal/kit/kit.go` path |
| Consumers are executables only: `cmd/cgo`, `cmd/cgo-gui` | grep `internal/kit` in `cmd/` (14 hits), `cmd/cgo-gui` |
| Every `Runner`/`Hypervisor` method mutates host/VM state (yaml writes via `writeFileAtomic` `kit.go:218-246`, `setYAMLKey` `kit.go:1273-1324`, `SaveVMX` `kit.go:346-365`) | `kit.go`, `vm.go:197-211,775-787` |
| Return type is `int` exit code + prints to `Stdout/Stderr` (`Runner.out/errf` `kit.go:385-386`) — the caller contract is process exit, not a Go value | `kit.go:822-1660`, `actions2.go` |
| Version identity is `-ldflags`-free string `KitVersion` (`kit.go:52`), hand-bumped with `VERSION` + frontend `package.json` per repo notes | `kit.go:49-52` |

### 5.2 Live-blocker rows (`docs/VERIFY.md`) and what they block

- `VERIFY.md:28` "Linux live acceptance (G0–G7, campaign) BLOCKED — Needs
  the stable Linux 6.8 bench; no live run recorded" → `kit deploy/ensure/
  testbed` have no live acceptance; any slice touching them must not claim
  bench-green.
- `VERIFY.md:29` eBPF BLOCKED (no `bpf_probe.o`, no Linux 6.8 run; CGO-DEV
  SSH rejects every recorded key/user) → `kit guest/bootstrap` SSH path is
  the blocker, not the scripts.
- `VERIFY.md:31` IFB ingress BLOCKED (Linux 6.8 bench) and `VERIFY.md:33`
  BBRv3 matrix BLOCKED (no 7.1 kernel) → `kit/install-bbr3.sh` is opt-in
  build tooling only; `kit testbed` cannot prove BBRv3.
- `VERIFY.md:34` `.dist` BLOCKED (no field trace ≥30 samples) — unrelated to
  kit/vm but gates `pkg/profile` economics calibration that `schedule`
  nightly campaigns would consume (`schedule.go:13-28` runs `cgo-linux run`).
- `VERIFY.md:56` PC2/nemo live bench BLOCKED (VM SSH rejects all recorded
  users/keys; WSL kernel 6.18, no `tc`/bpftool) → the documented
  `cgo kit ensure` SSH-wait loop (`kit.go:1050-1128`, 60×5 s = 300 s) is
  exactly the blocked path.
- `VERIFY.md:57` CHR RouterOS BLOCKED (QEMU not installed, no CHR image
  cached) → `kit/chr/setup-mikrotik-chr.sh` delegates to `cmd/cgo-chr`,
  but the QEMU/KVM row above ([L4][L5]) shows the missing driver is the
  blocker, not the shell wrapper.

### 5.3 Concrete gaps

1. **No QEMU/KVM hypervisor** — [L4][L5] document QEMU+KVM as the standard
   Linux path; `vm.go:791-852` implements VMware+VirtualBox only. CHR
   acceptance (`VERIFY.md:57`) is blocked on precisely this.
2. **No Ansible surface** — [L6] documents inventory/playbook/collections;
   kit re-implements idempotency in bash (`guest-setup.sh`, `testbed.sh`,
   `vm-install.sh`) + Go. No declarative desired-state exists; drift is
   detected only by `Next` probes (`next.go:44-273`).
3. **Config parser fragility** — hand-rolled YAML subset (`kit.go:100-212`,
   `gopkg.in/yaml.v3` present but unused by kit) plus three writers
   (`saveConfigValue` `kit.go:307-325`, `setYAMLKey` `kit.go:1273-1324`,
   `SaveVMX` `kit.go:346-365`); `vnet.go` duplicates path logic in
   `cmd/cgo-gui` parity test (`vnet_parity_test.go:1`). No schema/version
   field in `cgo-vm.yaml.example`.
4. **Secret handling** — `keysetup.go:116-121` types the password into
   `ssh` itself (good), but `config.sh` documents a `password` key
   (`config.sh:2-5`) and `setYAMLKey` treats `password` as an SSH key
   (`kit.go:1274`); `vnet.go`/`dns.go`/`tls.go` need admin. A library would
   inherit this boundary; an independent tool keeps it at the process edge.
5. **Version single-source** — `KitVersion` (`kit.go:52`) is source of
   truth but must be bumped manually with `VERSION` + frontend version
   (per repo research `2026-11-07-module-cli-frontend.md:29`); [L3] rule 5
   means `1.3.5` already claims a stable public API that is nowhere
   declared as such.
6. **Snapshot granularity** — `actions2.go:26-70` stops the VM for a cold
   snapshot; live snapshots are not attempted (gap vs [L7] `snapshot
   take --live`, flag present in [L7] Ch.8).

## 6. Target tool decomposition

### 6.1 Verdict: keep `kit`/`vm` as internal implementation of an independent tool, NOT as a published library

- **Internal rule** [L1]: `internal/kit` and `internal/vm` cannot be imported
  outside `github.com/Realms4239/cgo/`. Publishing them as Go modules
  requires a move to `pkg/` (or a new module root), which instantly freezes
  their Go signatures as a public API under [L3] rule 5 — including
  side-effecting methods (`Start`, `Stop`, `SetNetMode`, `EnsureNatSSH`)
  and process-global state (`Runner.Stdout`, yaml paths). That freeze is
  the wrong commitment: the operators' contract is the CLI, not the Go API.
- **SemVer** [L3]: a library must declare its public API (rule 1). The
  declared API here is 30 CLI actions + exit codes 2–8 + `CGO_*` env +
  yaml keys (`cmd/cgo/kit.go:23-59`). No Go-doc page declares e.g.
  `Runner.Ensure` stable; GUI callers (`cmd/cgo-gui/core.go`,
  `cmd/cgo-gui/status.go`) call the same symbols but are part of the same
  product — they are co-versioned, not consumers.
- **Modules** [L2]: a module is downloaded and `require`d at a version;
  kit/vm are never downloaded — they ship inside the `cgo` binary
  (binary-first, `package.go:15-20`; `deployPush` scp's `cgo-linux`,
  `kit.go:1396-1427`). Versioning them separately would version the
  *implementation detail* of a binary whose real version is `KitVersion`
  (`kit.go:52`).
- **Side-effect shape**: an importable library must be safe to import
  (no auto-start of VMs, no yaml mutation at init). `vm.Detect()` shells
  out (`vm.go:819-852`), `kit.Should…` writes config, `Ensure` boots a VM
  and waits 300 s (`kit.go:1050-1128`). Importing these gains the caller
  nothing over running `cgo kit ensure` and reading the exit code — and
  leaks host-admin privilege assumptions ([L6] Ansible's module boundary:
  modules run on the target with explicit privilege, not in-process).
- **Profile is different**: `pkg/profile` (`profile.go:18-93`) is already
  under `pkg/`, stdlib+`model` only, and *is* imported by `pkg/api`,
  `pkg/audit`, `cmd/cgo/calibrate.go`. It is the honest library candidate:
  promote it to its own module (`cgo/profile`) with `Import/Load/Delete/
  Reserved` as the declared API, SemVer-versioned per [L2][L3], and keep
  `data/profiles.json` format versioned. It does not start VMs, shell out,
  or mutate host state beyond one JSON file under `model.ProfilesMu`.

### 6.2 Target shape

```
cgo-kit   (standalone binary module, e.g. module cgo/kit)
  cmd/cgo-kit  — thin CLI: flag parsing + exit codes (today cmd/cgo/kit.go:15-214)
  kit          — src/cgo-kit/kit    (today internal/kit, minus GUI-coupled)
  vm           — src/cgo-kit/vm     (today internal/vm + NEW qemu driver, §7 S2)
  bench        — kit/testbed.sh + cgo-bench-dualns.sh + run-dualns.sh as data files
                 embedded or shipped, versioned with the binary, NOT a Go API
  profile      — separate Go module cgo/profile (today pkg/profile), pkg/model stays shared kernel
cgo (product)  — require cgo/kit only for the `kit` subcommand shim + Next probes;
                 require cgo/profile for /api/profile/*
```

Why not "one library `cgo/deploy`" : it would couple every consumer to
exit-code semantics, ssh/scp availability, hypervisor presence, and
`KitVersion` — the anti-pattern [L3] warns against (public API defined by
implementation accident). The independent-tool shape keeps the public API
where it actually is: process exit code + stdout + yaml diff.

## 7. Proposed slices

| Slice | Scope (one logical group, explicit paths) | Acceptance | Gate |
|---|---|---|---|
| S1 — Freeze CLI contract | Document+test `cmd/cgo/kit.go:23-59` actions, exit codes 2–8, `CGO_*` env, `cgo-vm.yaml` keys; add `kit contract` golden test; declare API in `docs/kit-api.md` per [L3] rule 1 | `cmd/cgo -run TestKitContract` green; `cgo kit --help` output committed as golden | `make verify` six gates; `VERIFY.md` row promoted only by independent verifier run |
| S2 — QEMU/KVM driver | New `Hypervisor` impl `internal/vm/qemu.go` driving `qemu-system-x86_64` ([L4]) with `-machine accel=kvm`, QMP `query-status`, `-netdev user,hostfwd=…` mirroring `NatForwarder`; CHR image cache + `windows-tcg` proof per `VERIFY.md:57` condition | `vm.Detect` returns qemu where present; `QEMU` not installed → honest "absent" like `Doctor` `kit.go:878-897` | `go test ./internal/vm -run Qemu` + live CHR boot on a host with QEMU+image (human-approved); BLOCKED rows `VERIFY.md:57` moves only after |
| S3 — profile module extraction | Move `pkg/profile` → module `cgo/profile`, `require` from root, keep `ProfilesMu` shim for one cycle per [L3] deprecation (minor → deprecate → major remove); version `data/profiles.json` | `go test` in new module green; root `go.mod` requires it without `replace` in tag ([L2]); all 4 callers compile | Full `make verify`; no frozen CSV schema change (`pkg/model/types.go:171-184`) |
| S4 — Config schema | Version field in `cgo-vm.yaml` (`version: 1`), single writer (`setYAMLKey` only), reject unknown keys in `LoadConfig` (`kit.go:74-83` fast/slow split kept); drop `password` key from `config.sh`+`setYAMLKey` `kit.go:1274` | Round-trip tests `yaml_atomic_test.go` extended; unknown key → exit 2 with remedy | `go vet` + kit tests; `test-kit.sh` updated |
| S5 — Ansible thin layer (optional, last) | Inventory generator (`cgo kit ansible-inventory` → JSON) + two playbooks wrapping `guest-setup.sh`/`testbed.sh` per [L6] playbook shape; no logic moved | `ansible-playbook --check` on bench host (human-approved) | BLOCKED until S2 live bench; never substitute for `VERIFY.md` live rows |

Live runs need human approval per repo AGENTS.md ("The live bench without
explicit human approval" is never-touch) and `VERIFY.md` how-to-move rules
(rows move to VERIFIED only after the named command runs on the named host).

## 8. Anti-toy criteria (a reviewer applies these; failing any one = toy)

1. **No fake hypervisor.** `Detect()` returns only executables actually
   present (`vm.go:819-852`, `findExe` + `exec.LookPath`); a stub `qemu` in
   tests must be labelled `fakeHyp` (`kit_test.go:153-161`) and never ship
   as a "driver". Proved by `kit/test-hypervisor.sh:5-7` greps + a live
   `VBoxManage list`/`vmrun list` capture in the checkpoint.
2. **Honest exit codes.** Every `Runner` method returns a documented code
   from the 2–8 table (`cmd/cgo/kit.go:58-59`); no `return 0` after a
   caught error. Fault proof: stub `ssh` failing → `Ensure` returns 5
   (`kit.go:1126-1127`), stub `vmrun` absent → `Scan` returns 3
   (`kit.go:914-917`).
3. **NAT forward verified, not assumed.** `EnsureNatSSH` re-reads
   `showvminfo` after posing (`vm.go:638-641`); `Ensure` probes the listener
   (`kit.go:1027-1029`). A rule string without `natRulePresent` is a toy.
4. **No silent config mutation.** `requireLockUser` (`kit.go:295-305`)
   gates `Ensure/Deploy/Testbed/Svc/Guest` before any write; `SaveVMX`
   normalises paths (`kit.go:346-365`); yaml writes go through
   `writeFileAtomic` (`kit.go:218-246`, `yaml_atomic_test.go:9-54`).
   A slice that writes `cgo-vm.yaml` outside these three writers fails.
5. **Live-bench claims stay BLOCKED until the command runs.** Per
   `VERIFY.md:28-60`, Windows unit green (e.g. `TestNATForwardTarget`,
   `TestScheduleUnit`) never promotes `Linux live acceptance`, `eBPF`,
   `IFB`, `BBRv3`, `.dist`, `CHR`, or `PC2/nemo` rows. The verifier re-runs
   and owns the row (repo roles: builder ≠ verifier).
6. **No topology guesses.** NAT forward only when `NetMode == "nat"`
   (`kit.go:1008-1031`); guest `10.0.2.x` never written to a host adapter
   (`vnet.go:245-261,306-308`); QEMU driver (S2) must select accel by
   probe (`-machine accel=help` per [L4]), never assume KVM.
7. **Snapshot is cold-and-stopped.** `Snapshot` stops the VM first
   (`actions2.go:34-41`); claiming live-snapshot safety without a
   `--live` implementation and a restore drill is a toy ([L7] `snapshot
   take --live` exists — unverified for VMware [L8]).
8. **Profile isolation.** `Reserved` P1–P4 can never be imported/deleted
   (`profile.go:18-29,71-74`, `reserved_test.go:25-56`); a test that
   writes `P3` to `data/profiles.json` and still sees the reference value
   must pass — masking is the bug this module exists to prevent.
9. **Version identity is one thing.** `KitVersion` (`kit.go:52`), `VERSION`,
   frontend version, and git tag must agree; a release where
   `cgo version` ≠ tag ≠ `KitVersion` is a toy release ([L3] rule 3 + §6.1).

## 9. Caveats

- VMware `vmrun` CLI surface: `unverified — primary source not found` [L8].
  All VMware semantics above are local-source evidence only.
- Ansible (L6), QEMU (L4), KVM (L5) were fetched as index/TOC pages only;
  playbook/module and KVM-ABI details beyond the TOC are
  `unverified — primary source not found` at the granularity needed for
  slices S2/S5 and must be fetched before implementation.
- `docs/deploy.md` and `docs/VERIFY.md` rows are project-local claims, not
  external primary sources; they are cited by line as evidence, and BLOCKED
  rows remain BLOCKED until the named host condition clears.
