# Module TUI / kit / vm — standalone tools (research)

Date: 2026-11-07. Scope: `cmd/cgo/tui*.go`, `cmd/cgo/kittui*.go`, `cmd/cgo/kit.go`,
`internal/kit/*.go`, `internal/vm/*.go`, `docs/deploy.md`. Revisit of the
kit/vm split in `docs/research/2026-11-07-module-kit-vm-profile.md` (which
covered kit/vm/profile without the TUI). No code, no commit.

Method: read source, inventory exported surface by `file:line`, check against
the primary sources below. Every claim cites `file:line` or `(R#)`; anything
else is labelled **unverified — primary source not found**.

## Primary sources (fetched 2026-11-07)

- [B1] Bubble Tea — charmbracelet/bubbletea (GitHub) —
  https://github.com/charmbracelet/bubbletea : "A Go framework based on The
  Elm Architecture" ; "Bubble Tea programs are comprised of a model that
  describes the application state and three simple methods" —
  `Init` (initial `Cmd`), `Update` (handles events, returns updated model
  + `Cmd`), `View` (renders UI from model) ; `tea.NewProgram(initialModel())`
  then `p.Run()` ; "can't really log to stdout with Bubble Tea because your
  TUI is busy occupying that".
- [B2] `pkg.go.dev` — `github.com/charmbracelet/bubbletea v1.3.10` —
  https://pkg.go.dev/github.com/charmbracelet/bubbletea : `type Cmd func() Msg` ;
  `Tick(d, fn)` , `Batch`, `Quit`, `WithAltScreen`, `WithFPS` ;
  `type KeyMsg` / `KeyMsg.String()` ; `Msg` includes `Quit`, `Suspend`,
  `EnterAltScreen`, `ClearScreen`. v1 tutorial uses `tea.KeyMsg` and
  `func (m model) View() string`.
- [L1] Lip Gloss — charmbracelet/lipgloss (GitHub) —
  https://github.com/charmbracelet/lipgloss : "Style definitions for nice
  terminal layouts" ; `lipgloss.NewStyle().Bold().Foreground().Background().
  Padding().Border().Render(...)` ; automatic downsampling of color profiles
  (TrueColor → 256 → 16) ; `lipgloss.Wrap`, `lipgloss.Width/Height/Size`.
- [Q1] QEMU documentation — https://www.qemu.org/docs/master/system/introduction.html :
  "Supported Accelerators" table rows `KVM | Linux | Arm, MIPS, PPC, RISC-V,
  s390x, x86`, `Hypervisor Framework | MacOS`, `Windows Hypervisor Platform |
  Windows`, `TCG` fallback ; "The QEMU Monitor Protocol (QMP) is a well
  defined, versioned, machine usable API that presents a rich interface to
  other tools to create, control and manage Virtual Machines."
- [G1] Go `go.mod` file reference — https://go.dev/doc/modules/gomod-ref :
  module path is the unique identifier + import prefix, v2+ must end `/v2` ;
  `go` directive is the mandatory minimum Go version since Go 1.21 ;
  `require`/`replace`/`exclude`/`retract` semantics ; single-versioning per
  module.
- [G2] Go 1.4 release notes, "Internal packages" — https://go.dev/doc/go1.4 :
  "place it in a directory named `internal` … may not be imported by packages
  outside the source subtree … `.../a/b/c/internal/d/e/f` can be imported only
  by code in the directory tree rooted at `.../a/b/c`".
- [C1] Google Chrome Help, "Speed up Google Chrome" —
  https://support.google.com/chrome/answer/1385029 : "Task Manager" via
  More → More tools → Task Manager ; "Select 'Memory' to sort tasks by memory
  usage" ; "Select the task you want to close. Select End Process. Note: Any
  unsaved work in your task will be lost." No design specification or
  interaction contract beyond these user steps was fetched; any stronger
  Task-Manager analogy is therefore **unverified — primary source not found**.
- VMware `vmrun` manual : **unverified — primary source not found** (Broadcom
  TechDocs portal returned no `vmrun` manual on 2026-11-07; all `vmrun`
  behaviour below is local-source evidence only, not primary).

Local claims cite `path:line`. Where no primary source exists the claim is
marked `unverified — primary source not found`.

## 1. Responsibility

### 1.1 `cmd/cgo` operator TUI — live dashboard + sealed-audit mirror (Bubble Tea)

- Two files, one `package main` model. `cmd/cgo/tui.go:1-405` defines the
  5-tab dashboard (`[1 Setup] [2 Kit] [3 Campagne] [4 Live] [5 Résultats]`,
  `tui.go:33`) and the `main` entry `runTUI` at `tui.go:398-405` :
  `tea.NewProgram(initialModelTUI(), tea.WithAltScreen())` then `p.Run()`
  ([B1] canonical `NewProgram`+`Run` shape; [B2] `ProgramOption`
  `WithAltScreen`). `cmd/cgo/main.go:53-54` dispatches `tui|top` → `runTUI`.
- `cmd/cgo/tui_audit.go:1-577` defines the audit mirror rendered above the
  campaign on tabs 4–5 and the two full-screen panes. The header comment
  (`tui_audit.go:18-23`) states the contract: tabs 4 and 5 are the campaign,
  not an audit; the mirror gives "le verdict et les quatre faits du dernier
  gel, sans sixième onglet". `OK` distinguishes "aucun audit" from "je n'ai
  pas pu lire" (`tui_audit.go:22-23`) — a transport failure must never read
  as an unmeasured link.
- Data flow is polling only, per [B2] `Cmd`/`Tick` : `pollFrames`
  (`tui.go:127-139`, 1 s tick on `/api/state`), `pollResults`
  (`tui.go:141-159`, on `/api/results`), `pollAudit`
  (`tui_audit.go:105-129`, on `/api/audit/status`), `httpPost`
  (`tui.go:260-278`, `POST /api/run/start|stop`). No WebSocket, no SSE.
  `modelTUI.Update` (`tui.go:180-258`) is the [B1] `Update` method: key
  handling (`q`/`ctrl+c` → `tea.Quit`, `h/l|1-5` tabs, `j/k` kit selection,
  `enter/o` exec, `d/p/c/tab` panes on tabs ≥3), message re-issue
  (`frameMsg` → `pollFrames`, `resultsMsg` → 5 s `Tick`, `auditMsg` → 10 s
  `Tick`). `modelTUI.View` (`tui.go:282-393`) is the [B1] pure render of the
  current model; styling is [L1] `lipgloss.NewStyle().Padding.Foreground.
  Border().Render(...)` (`tui.go:20-31`, e.g. `stTab`, `stCell`,
  `stCellOn` with `RoundedBorder`).
- Kit delegation is by subprocess, not by import : `kitExec`
  (`tui.go:161-167`) runs `exec.Command(os.Args[0], "kit", action)` and
  returns `kitDoneMsg` with combined output. `Update` (`tui.go:223-229`)
  maps cursor to `["doctor","scan","ensure","align","build","deploy",
  "bootstrap","status","logs"]`. The TUI therefore never calls `kit.Runner`
  directly; it re-enters the binary ([B2] `Cmd func() Msg` = "IO operation
  that returns a message when it's complete").
- Audit panes obey an 80-column budget proved by tests: `wrap78`/`wrapWidth`
  (`tui_audit.go:204-257`, 78-col panes, hard-cut long words, never
  horizontal scroll), `auditMirror` (`tui_audit.go:511-543`) wraps hint+status
  at `paneWidth-2`, `TestRenderedViewsStayWithinEightyColumns`
  (`tui_audit_test.go:250-273`) renders all tabs + panes and fails on any
  line >80 runes. Four facts are exactly four lines, zero never stands in
  for absent (`auditFactLines`, `tui_audit.go:152-188`: `RTTIdleP50==0 &&
  RTTIdleP95==0` → "Repos : non mesuré"; `LoadAttempted/LoadMeasured`
  distinguishes "non tentée" / "non établie"; `LossPct *float64` pointer so
  a true 0 % renders as "0.00 %", nil renders as "non mesuré").
- Script writing is path-safe: `safeAuditID` (`tui_audit.go:433-448`) rejects
  empty/`.`/`..`/slash/backslash/`..` substring (same rule as `validRunID`),
  `scriptFor` (`tui_audit.go:454-477`) returns verbatim RouterOS or one
  `tc` command per line with ingress behind a comment, `writeRemediation`
  (`tui_audit.go:485-506`) writes under `data/audits/<id>/` only
  (`auditScriptRoot = data/audits`, `tui_audit.go:481`, verified by
  `TestWriteRemediationStaysUnderDataAudits`), never `/tmp`.
  `copyScript` (`tui_audit.go:548-571`) writes the file then attempts OSC 52
  (`osc52`, `tui_audit.go:575-577`); status says "copie OSC 52 tentée",
  never "copiée" (no terminal ACK exists — comment at `tui_audit.go:573-574`).
- The TUI holds no measurement code: `measuredView`/`verdictLine`
  (`tui_audit.go:135-146`) render the frozen `metrics.Verdict.Word/Sentence`
  verbatim ("Jamais un chiffre recalculé : la cascade vit dans pkg/metrics
  et son résultat est gelé", `tui_audit.go:133-134`). `sensorTier`
  (`tui_audit.go:328-339`) reports `ebpf|tcpinfo|standard` from the report's
  kernel block, never probes the local kernel.

### 1.2 `cmd/cgo` kit centre de contrôle — 5-step Bubble Tea wizard

- `cmd/cgo/kittui.go:1-557` + `cmd/cgo/kittui_run.go:1-961` implement the
  second TUI, dispatched from `cmd/cgo/kit.go:170-171` (`kit tui` →
  `runKitTUI`), run at `kittui_run.go:944-961` as
  `tea.NewProgram(initialModelKT(cfgPath, version))` ([B1]/[B2] same shape
  as §1.1). File header (`kittui.go:3-8`): "Un PC Windows + une VM Ubuntu :
  ce TUI scanne les hyperviseurs, fait choisir la VM au clavier, installe
  les dépendances, pose la clé SSH, pousse le binaire et sert le dashboard."
- State machine is five `ktStep` (`kittui.go:23-31`): `ktDeps → ktVM →
  ktAcces → ktDeploy → ktControle`, labels at `kittui.go:33`. Each step
  exposes `items()` (`kittui.go:180-253`, stable `id` + display `label/hint`)
  navigated by `j/k` + `enter` ([B1] cursor pattern). Header
  (`kittui.go:480-505`) renders one line on every screen: machine
  (`● allumée / ○ éteinte / ? inconnue` + net mode), SSH, dashboard, IP +
  access path — the [C1] Task-Manager analogy is relevant here: a single
  status line over a sortable/actionable list, where the list is steps and
  the actions are `End Process`-like (start/stop/deploy). The analogy is
  limited to that presentation shape; **unverified — primary source not
  found** for any claim about Task Manager internals or real-time metering
  ([C1] documents only the user steps "sort by Memory" / "End Process").
- Heavy work runs off the UI thread: `kittui_run.go` defines `ktLog`,
  `ktDone`, `ktVMs`, `ktDepsMsg`, `ktSSH`, `ktDash`, `ktDiagMsg` messages
  and runs `Runner` calls in goroutines that send messages back — the [B2]
  `Cmd` pattern ("IO operation that returns a message"). `keysetup`
  suspends the TUI for the native password prompt then resumes (file header
  `kittui.go:7-8`). Journal is capped at 200 lines in memory
  (`kittui.go:106-112`) and appended to `kit-tui-<date>.log` beside the
  binary (`kittui.go:114-158`), never stdout — consistent with [B1] "can't
  really log to stdout with Bubble Tea".
- This TUI *does* import `internal/kit` and `internal/vm` directly
  (`kittui.go:18`, `kittui_run.go:18-19`) and reads `kit.Config` in the
  render path (`sshTarget` `kittui.go:430-435`, `dashURL` `kittui.go:255-266`,
  `lockedVM` `kittui.go:166-177`). `kittui.go:94-100` caches VM live state
  outside render ("jamais par render (showvminfo/vmrun coûteraient par
  touche)") — a performance rule, not a [B1]/[B2] requirement.
- `cmd/cgo/kit.go:15-215` is the CLI shim for both TUIs: `runKit` parses
  `--config/--deep/--shallow/--public/--yes`, dispatches 32 actions
  (`kit.go:94-214`), documents exit codes `2 usage/build, 3 scan ambigu,
  4 hyperviseur absent, 5 timeout SSH, 6 cross-compile/bootstrap, 7 scp,
  8 install/logs` (`kit.go:58-59`). `cgo tui` (`main.go:53-54`) and
  `cgo kit tui` (`kit.go:171`) are two distinct binaries' entries inside one
  `package main`.

### 1.3 `internal/kit` — deployment engine (operator pilot)

- `internal/kit/kit.go:1-6`: "moteur de déploiement en Go : les 11 actions
  de l'ancien engine.sh … mêmes codes de sortie … mêmes variables
  d'environnement CGO_*". The engine is a CLI contract: 32 actions, exit
  codes 2–8, `CGO_*` env vars and `kit/cgo-vm.yaml` keys (`kit.go:29-47,74-212`).
  `docs/deploy.md:106-128` states the contract: host→VM deployment is driven
  by `cgo kit` (Go engine, same exit codes as legacy bash; `kit/engine.sh`
  is a shim), binary-first and idempotent, SPA embedded, VM needs only
  `iproute2 curl bc` + `tcp_bbr`.
- `internal/kit/kit.go:32-47` defines `Config` (SSH user/host/port/key, VM
  paths, hypervisor, snapshot, project/dashboard dirs, NAT host port);
  `LoadConfig`/`LoadConfigFast` (`kit.go:74-83`) load the YAML subset then
  env (env wins, `kit.go:171-192`); "auto" host resolves via the
  path-selected hypervisor driver (`kit.go:196-205`), skipped in Fast mode
  (no subprocess on UI thread). `KitVersion = "1.3.5"` at `kit.go:52` is the
  single version string (`cmd/cgo/main.go:41` `version = kit.KitVersion`).
- `Runner` (`kit.go:368-372`, `NewRunner` at `kit.go:374-383`) carries
  `Stdout/Stderr/Root` (mockable in tests). Every action returns `int` exit
  code and prints via `out/errf` (`kit.go:385-386`). Methods:
  `Doctor` (`kit.go:822`), `Scan` (`kit.go:910`), `Ensure` (`kit.go:982`),
  `Build` (`kit.go:1327`), `Deploy` (`kit.go:1372`), `Bootstrap`
  (`kit.go:1546`), `Status` (`kit.go:1562`), `Logs` (`kit.go:1590`),
  `Align` (`kit.go:1601`), `Tunnel` (`kit.go:1660`); `actions2.go`:
  `Snapshot/Revert/SSHInteractive/Ps/Svc/Backup/Snapshots/Verify/Health`
  (`actions2.go:26,73,131,152,187,314,397,461,476`); plus `DNS` (`dns.go:15`),
  `TLS` (`tls.go:19`), `KeySetup` (`keysetup.go:23`), `Guest` (`guest.go:16`),
  `Testbed` (`testbed.go:13`), `Package` (`package.go:21`), `Schedule`
  (`schedule.go:54`), `Next/NextVer` (`next.go:359,364`), `Netinfo`
  (`netinfo.go:10`), `Nic` (`netnic.go:11`), `Vnet` (`vnet.go:23`),
  `ShipCheck` (`shipcheck.go:18`).
- SSH/SCP/health primitives live on `Config`: `SSH` (`kit.go:469-488`,
  `BatchMode`, key, `ConnectTimeout=4`, `ServerAliveInterval=10`,
  `StrictHostKeyChecking=accept-new`, 10 min `context.WithTimeout`, stdin
  NUL), `SSHUp` (`kit.go:521-524`), `SCP/SCPOut/SCPOutErr`
  (`kit.go:663-710`, POSIX shell-quoting `shq`), `Health`
  (`kit.go:752-764`, `InsecureSkipVerify` — LAN tool, see `kit tls`).
  `NATForwardTarget` (`kit.go:562-582`) is the single NAT root (explicit
  loopback/`auto`/empty → no NAT; guest `10.0.2.x` → `127.0.0.1:natPort`;
  else driver-selected `NatForwarder` + `NetMode == "nat"`); `ensureNATForward`
  (`kit.go:587-600`) poses the rule, `Ensure` (`kit.go:982-1128`) only poses
  in `nat` mode (`ponté`/`hôte-only` skip with explanation,
  `kit.go:1010-1014`) and verifies the listener (`kit.go:1027-1029`).
- Config writers are three only: `writeFileAtomic` (`kit.go:218-246`, tmp +
  rename), `setYAMLKey` (`kit.go:1273-1324`), `SaveVMX`/`saveConfigValue`/
  `SaveSSHTarget` (`kit.go:307-365,330-340`) — all go through
  `writeFileAtomic`. `requireLockUser` (`kit.go:295-305`) gates late actions
  before any write, pointing at `LinearGuide` (`kit.go:285-289`).

### 1.4 `internal/vm` — hypervisor drivers

- `internal/vm/vm.go:1-5`: "détection et pilotage des hyperviseurs, pur Go
  multi-OS … cherche .vmx/.vbox". Two drivers only: VMware (`vmrun`) and
  VirtualBox (`VBoxManage`) — `Detect()` (`vm.go:835-852`) searches
  `vmrun[.exe]` then `VBoxManage[.exe]` via `candidates` (`vm.go:794-817`).
- `Hypervisor` interface (`vm.go:22-45`): `Name, Exe, Running, Start,
  StartGUI, Stop, GuestIP, NetMode, SetNetMode` — every method has an
  external side effect (start/stop a VM, rewrite `.vmx`, `modifyvm`). It is
  not a query library. `NatForwarder` (`vm.go:49-55`): `EnsureNatSSH(vmx,
  hostPort, dashPort)`, `NatHostAddr()`.
- `ScanVMs`/`ScanVMsProgress` (`vm.go:931-1008`): hypervisor inventory first,
  then `<drive>/VMs`, `~/VirtualBox VMs`, shallow root, deep scan;
  `extraRoots` (`vm.go:897-926`). `HypForPath` (`vm.go:1015-1023`) decides
  by extension (`.vbox` → `virtualbox`, `.vmx` → `vmware`) before config;
  `Pick` (`vm.go:1042-1054`) is exact-name-else-unique-else-ambiguous;
  `IsRunning` (`vm.go:318-331`) is the single state root;
  `SameVM` (`vm.go:696-698`) dedupes by normalized path.
- VirtualBox NAT guest space `10.0.2.x` is unreachable from host by
  construction; the fix is a host port-forward (`EnsureNatSSH`,
  `vm.go:562-642`, rules `cgo-ssh,tcp,,2222,,22` and
  `cgo-dashboard,tcp,,9090,,9090` at `vm.go:614,617`). VMware NAT keeps a
  guest IP reachable via vmnet8 (`NetMode` reads `ethernet0.connectionType`,
  `vm.go:130-162`). `SetNetMode` requires the VM stopped on both drivers
  (`vm.go:197-211,775-787`).
- Subprocess surface: `vmrun list/start/stop/getGuestIPAddress`
  (`vm.go:64-125`), `VBoxManage list runningvms/showvminfo/guestproperty/
  controlvm natpf1/modifyvm` (`vm.go:220-231,248-269,484-544,562-642`).
  `vmrun` CLI truth: **unverified — primary source not found**; VBoxManage
  maps to the public VirtualBox manual (not re-fetched here; see prior
  research `2026-11-07-module-kit-vm-profile.md:L7`). No QEMU/KVM driver
  exists: grep `qemu` in `internal/vm` = 0, while [Q1] lists KVM as the
  Linux accelerator and QMP as the management API — the missing driver is
  the gap for CHR/Windows-TCG acceptance.

### 1.5 `docs/deploy.md` — deployment contract

- `docs/deploy.md:106-128` (VM de banc) is the host→VM contract restated
  above; `docs/deploy.md:1-34` is the Windows-observation vs Linux-full
  matrix (`tc`/`netns` `n/a — hors Linux`, offloads `warn` not fail on
  Windows, hard fail on Linux when bench mounted). Security boundary
  (`docs/deploy.md:159-177`): dashboard is single-operator, no auth, no CSRF,
  1 MiB JSON bound is not access control; shaping modifies the edge;
  nginx boundary S1 (`deploy/nginx/cgo.conf`, `limit_req 1r/s burst=3` →
  503) and OCI/rollback S4 are stated at `deploy.md:179-183`.

## 2. Exported API (file:line)

### 2.1 TUI — nothing is exported (all `package main`)

`cmd/cgo/tui.go`, `tui_audit.go`, `kittui.go`, `kittui_run.go` are
`package main` (`tui.go:1`, `tui_audit.go:1`, `kittui.go:1`,
`kittui_run.go:1`). No symbol is importable outside the binary. The
"exported" surface is therefore the CLI + HTTP + filesystem contract:

| Surface | Evidence |
|---|---|
| `cgo tui` / `cgo top` → `runTUI` | `cmd/cgo/main.go:53-54`, `cmd/cgo/tui.go:398-405` |
| `cgo kit tui` → `runKitTUI` | `cmd/cgo/kit.go:170-171`, `cmd/cgo/kittui_run.go:944-961` |
| `GET /api/state`, `GET /api/results`, `GET /api/audit/status`, `POST /api/run/start`, `POST /api/run/stop`, `CGO_API` default `https://meteolink.dev:9090` | `tui.go:89-95,127-159,260-278`, `tui_audit.go:105-129` |
| Script files `data/audits/<audit_id>/remediation.routeros.rsc` / `remediation.linux.sh` | `tui_audit.go:424-506` |
| Kit log `kit-tui-<YYYYMMDD>.log` beside binary (fallback temp) | `kittui.go:114-158` |
| Tab order Setup/Kit/Campagne/Live/Résultats; panes deployment/proof; keys `h/l 1-5 q d p c tab esc j/k enter` | `tui.go:33,182-240`, `tui_audit.go:93-101` |
| Kit steps Dépendances/Machine/Accès SSH/Déployer/Contrôle; `items()` ids | `kittui.go:23-33,180-253` |

Unexported constructors (useful as test seams, not API): `initialModelTUI`
(`tui.go:89-95`), `initialModelKT` (`kittui.go:101-104`),
`modelTUI.Init/Update/View` (`tui.go:176-393`),
`modelKT.Init/View` (`kittui.go:280-327`), `auditFactLines/verdictLine/
deploymentLines/proofLines/scriptFor/writeRemediation/copyScript/osc52`
(`tui_audit.go:135-577`). Tests pin behaviour, not compatibility:
`cmd/cgo/tui_audit_test.go:44-356` (16 tests) and
`cmd/cgo/kittui_test.go:14-229`.

### 2.2 `internal/kit` (`internal/kit/kit.go` + satellites)

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

### 2.3 `internal/vm` (`internal/vm/vm.go`)

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
`GuestIP`, `EnsureNatSSH`, `NetMode`/`SetNetMode` are interface methods,
not free functions.

## 3. Dependencies

- Module root: `module github.com/Realms4239/cgo`, `go 1.25`
  (`go.mod:1-3`); per [G1] the module path is the unique identifier +
  import prefix, and the `go` directive is the mandatory minimum toolchain
  (enforced since Go 1.21). Direct requires (`go.mod:5-10`):
  `fyne.io/fyne/v2 v2.8.1`, `github.com/charmbracelet/bubbletea v1.3.10`,
  `github.com/charmbracelet/lipgloss v1.1.0`, `golang.org/x/sys v0.36.0`
  ([B2]/[L1]/[G1] pin the TUI legs).
- TUI imports: `tea "github.com/charmbracelet/bubbletea"` +
  `"github.com/charmbracelet/lipgloss"` (`tui.go:14-15`, `tui_audit.go:13`,
  `kittui.go:19-20`, `kittui_run.go:20`); `tui_audit.go:15` also imports
  `github.com/Realms4239/cgo/pkg/metrics` (for `metrics.Verdict`,
  `metrics.CAPDeadlineMs` at `tui_audit.go:33,173`). No Bubbles, no
  Harmonica, no ntcharts — [B1]'s "Libraries we use with Bubble Tea" list is
  not a requirement, only a pointer.
- `internal/kit` imports: stdlib + `github.com/Realms4239/cgo/internal/vm`
  (`kit.go:26`, `actions2.go:16`, `guest.go:6`, `next.go:15`, `vnet.go:11`)
  + `golang.org/x/sys/windows` on Windows only (`exec_windows.go:9`).
  Callers: `cmd/cgo/kit.go:9`, `cmd/cgo/main.go:28`, `cmd/cgo/setup.go:12`,
  `cmd/cgo/kittui*.go`, `cmd/cgo-gui/*`. No external hypervisor SDK.
- `internal/vm` imports: stdlib only + `golang.org/x/sys/windows`
  (`drives_windows.go:5`, `execcmd_windows.go:9`). All control is subprocess
  (`vmrun`, `VBoxManage`, `arp`, `ip`, `ping`) e.g. `vm.go:64-75,220-231,
  1183-1266`.
- File-IO deps: `kit/cgo-vm.yaml` (machine-local) via hand-rolled parser
  `kit.go:100-212` (no `gopkg.in/yaml.v3` import in `kit` — `yaml.v3` is
  indirect in `go.mod:59`); `data/audits/<id>/` scripts
  (`tui_audit.go:481`); `kit/*.sh` located at runtime by `findKitFile`
  (`vnet.go:345-358`); kit log `kit-tui-<date>.log` (`kittui.go:114-158`).
- Host tools (not Go deps): `ssh`/`scp`/`curl`/`go`/`bun`/`node` probed by
  `Doctor` (`kit.go:822-907`); `iproute2`/`curl`/`bc` on the VM
  (`docs/deploy.md:125-128`); `vmrun[.exe]`/`VBoxManage[.exe]`
  (`vm.go:794-852`). [Q1] names QEMU+KVM as a *future* host tool; it is not
  a current dependency.

## 4. Primary-source requirements

- [G2] `internal/` may not be imported outside the subtree rooted at the
  parent of `internal/` — so `internal/kit`+`internal/vm` are usable only
  by packages under `github.com/Realms4239/cgo/` (observed: `cmd/cgo`,
  `cmd/cgo-gui`; `cmd/cgo` is `package main` so the TUI cannot be reused
  by any other binary). Promoting kit/vm to publishable Go modules
  **requires** moving them out of `internal/` — and then every host-side
  side effect becomes a public API commitment. Same rule applies to the
  TUI: `package main` models (`modelTUI`, `modelKT`, `auditView`) are
  unimportable by construction.
- [G1] A module is versioned as one unit (single `go.mod`, `require` graph,
  `go` floor `1.25` at `go.mod:3`). Splitting `cgo-tui` / `cgo-kit` means
  two new `go.mod` files with their own module paths ([G1] "module path is
  the module's unique identifier"), each declaring its `require` set
  (`bubbletea`/`lipgloss` for the TUI repo, stdlib-only for the VM repo)
  and its `go` floor. A `replace` directive may be used transiently during
  migration ([G1] `replace` section) but must not survive a tag.
- [B1]/[B2] Bubble Tea contract: a program is `Init` (initial `Cmd`),
  `Update` (`Msg` → updated `Model` + `Cmd`), `View` (render from model);
  `Cmd` is `func() Msg` ("IO operation that returns a message"); `Tick`
  sends a single message and must be re-armed by returning another `Tick`
  from `Update` ([B2] `Tick` doc); `WithAltScreen` owns the alternate screen;
  `KeyMsg.String()` is the sanctioned key comparison; stdout belongs to the
  renderer ("can't really log to stdout" [B1]). Any standalone `cgo-tui`
  must preserve this loop exactly — replacing polling `Cmd`s with direct
  HTTP in `View` would violate the architecture.
- [L1] Lip Gloss is a style-only library ("Style definitions for nice
  terminal layouts", `NewStyle().Border().Foreground().Render()`). It
  guarantees color-profile downsampling, not layout semantics, and provides
  `Wrap`/`Width`/`Size` helpers. The 80-column budget is a project rule
  (`paneWidth = 78`, `tui_audit.go:98`), not a [L1] requirement.
- [Q1] QEMU/KVM: KVM is the Linux accelerator (`KVM | Linux | …`) and QMP is
  "the interface used by higher level tools". A `cgo-kit` repo claiming
  QEMU/KVM support must therefore drive QEMU through `qemu-system-x86_64`
  + QMP (`query-status`, `-netdev user,hostfwd=…` mirroring `NatForwarder`),
  not through a KVM ioctl from Go — ioctl-level guest control from a Go
  tool: **unverified — primary source not found** at the granularity needed.
- [C1] Chrome Task Manager: the only fetched primary surface is the user
  flow "More tools → Task Manager", "sort by Memory", "End Process (unsaved
  work lost)". It supports the *analogy* "one status line + sortable list +
  destructive action per row" for the kit centre de contrôle header
  (`kittui.go:480-505`) and the `svc`/`vm-stop` items
  (`kittui.go:234-240`). It does not support claims about live metering,
  per-process CPU accounting, or process-tree semantics — those remain
  **unverified — primary source not found**.
- VMware `vmrun`: **unverified — primary source not found**. Any slice
  touching `vmrun` semantics must treat `internal/vm/vm.go:64-125` as
  local evidence only, or fetch the Broadcom manual first.

## 5. Gap analysis (revisit of the kit/vm split, with the TUI)

### 5.1 Tool vs library (structural, TUI included)

| Fact | Evidence |
|---|---|
| kit/vm live under `internal/` → unimportable outside module by Go rule | [G2]; `internal/kit/kit.go` path |
| TUI lives in `package main` → zero importable symbols, tested only via `package main` tests | `tui.go:1`, `tui_audit_test.go:1-11`, `kittui_test.go:1` |
| Consumers of kit are executables only: `cmd/cgo`, `cmd/cgo-gui`, `cmd/cgo setup.go` | grep `internal/kit` in `cmd/` |
| Consumers of the operator TUI: `runTUI` only (`main.go:53-54`); of the kit TUI: `runKitTUI` only (`kit.go:171`) | `tui.go:398`, `kittui_run.go:945` |
| Every `Runner`/`Hypervisor` method mutates host/VM state (yaml writes via `writeFileAtomic` `kit.go:218-246`, `setYAMLKey` `kit.go:1273-1324`, `SaveVMX` `kit.go:346-365`) | `kit.go`, `vm.go:197-211,775-787` |
| Return type is `int` exit code + prints to `Stdout/Stderr` (`Runner.out/errf` `kit.go:385-386`) — the caller contract is process exit, not a Go value | `kit.go:822-1660`, `actions2.go` |
| TUI polling hardcodes `CGO_API` default `https://meteolink.dev:9090` and per-endpoint timeouts (2 s / 3 s / 5 s) with no backoff | `tui.go:89-95,127-159,260-278`, `tui_audit.go:105-129` |
| Operator TUI execs `os.Args[0] kit <action>` (assumes the binary contains the kit subcommand); kit TUI imports `kit` directly — two different couplings to the same engine | `tui.go:161-167` vs `kittui.go:18-20` |

### 5.2 Including the TUI changes the split

1. **The operator TUI is the easiest extraction, the kit TUI is not.**
   `tui.go` + `tui_audit.go` touch the network (four endpoints), `pkg/metrics`
   (one type + one const, `tui_audit.go:15,33,173`), the filesystem
   (`data/audits/<id>/…`), and Bubble Tea/Lip Gloss — but never `internal/kit`
   or `internal/vm` except via subprocess (`kitExec`). It is already a
   standalone tool wearing a `package main` costume.
   `kittui.go` + `kittui_run.go` import both `internal/kit` and `internal/vm`,
   read `kit.Config` in the render path, and reimplement
   `dashURL`/`sshTarget`/`lockedVM` (`kittui.go:166-177,255-266,430-435`)
   instead of calling `kit` methods — it is a second kit VM client, not a
   view over a shared core.
2. **There is no QEMU/KVM hypervisor.** [Q1] names KVM+QMP as the Linux path;
   `vm.go:791-852` implements VMware+VirtualBox only (grep `qemu` = 0 in
   `internal/vm`). CHR acceptance is blocked on precisely this (prior
   research §5.2; `docs/VERIFY.md` CHR row BLOCKED). A `cgo-kit` repo without
   a QEMU driver cannot claim CHR/Windows-TCG.
3. **Config parser fragility** (carried from prior research §5.3.3):
   hand-rolled YAML subset (`kit.go:100-212`, `gopkg.in/yaml.v3` unused by
   kit) plus three writers; no schema/version field in `cgo-vm.yaml`. The
   kit TUI adds a fourth reader (`initialModelKT` → `kit.LoadConfig`,
   `kittui.go:101-104`) with no validation surface.
4. **Secret handling**: `keysetup` types the password into `ssh` itself
   (good), but `config.sh` documents a `password` key and `setYAMLKey`
   treats `password` as an SSH key (`kit.go:1274`); `dns/tls/vnet` need
   admin. A TUI that stores `inputVal` (`kittui.go:90-92`) in memory must
   stay at the process edge.
5. **Version single-source**: `KitVersion` (`kit.go:52`) is source of truth
   but bumped manually; `kittui.go:69` `version` and `tui.go:51` `Version`
   are two more copies of the same number ([G1] single-versioning per
   module is not respected across the three surfaces today).
6. **TUI honesty rules are the product**: audit facts never invent zero
   (`tui_audit.go:152-188`), verdict never recomputed (`tui_audit.go:135-146`),
   scripts never empty (`scriptFor` honest errors, `tui_audit.go:454-477`),
   paths never escape (`safeAuditID`+prefix check,
   `tui_audit.go:433-506`), OSC 52 never claimed delivered
   (`tui_audit.go:573-577`), polling never fabricates live data
   (`frameMsg{}` on error, `tui.go:130-138`). Any extraction that drops one
   rule is a toy even if it compiles.

### 5.3 Live-blocker rows (`docs/VERIFY.md`) still blocking kit/vm/TUI slices

- Linux live acceptance (G0–G7 campaign) BLOCKED (needs stable Linux 6.8
  bench) → `kit deploy/ensure` and `cgo tui` Live tab have no live acceptance.
- eBPF BLOCKED (no `bpf_probe.o`, no Linux 6.8 run) → `kit guest/bootstrap`
  SSH path is the blocker, not the scripts.
- IFB / BBRv3 / `.dist` / CHR / PC2-nemo rows BLOCKED → as in prior research
  §5.2; the TUI mirror is fixture-tested only (VERIFY row "20C TUI mirrors
  the last sealed audit" notes "Never observed against a live audit").

## 6. Target tools — `cgo-tui`, `cgo-kit` as independent repos

### 6.1 Verdict (revisited)

- **Extract both as independent *tools* (binaries), not as published
  libraries.** [G2] forbids importing `internal/kit`/`internal/vm` outside
  the module, and `package main` forbids importing the TUI at all. A
  published Go library would freeze side-effecting signatures (`Start`,
  `Stop`, `SetNetMode`, `EnsureNatSSH`, `Runner.Ensure` waiting 300 s at
  `kit.go:1050-1128`) as a public API — the wrong commitment. The
  operators' contract is process behaviour: exit codes 2–8, `CGO_*` env,
  `cgo-vm.yaml` keys, `/api/state|results|audit/status` polling, script
  files under `data/audits/<id>/`, `kit-tui-<date>.log`.
- **Split the two TUIs.** The operator TUI (§1.1) and the kit TUI (§1.2)
  share only Bubble Tea/Lip Gloss idioms ([B1] `Init/Update/View`, [L1]
  `NewStyle`). They share no model type, no message type, and opposite
  couplings (subprocess vs direct import). Folding them into one `cgo-tui`
  module would couple a read-only monitor to host-admin actions
  (`SetNetMode`, NAT forwards) — the exact privilege boundary
  `docs/deploy.md:159-177` draws.
- [C1] fits exactly one of the two: the kit centre de contrôle header
  (`kittui.go:480-505`) is the Task-Manager shape (status line + actionable
  list + `End Process`-like `svc-stop`/`vm-stop`). The operator dashboard
  is a different shape (live cells + audit mirror + result table) and must
  not be justified by [C1].

### 6.2 Target shape

```
cgo-tui   (independent repo, module github.com/<org>/cgo-tui, own go.mod + go floor [G1])
  cmd/cgo-tui       — main.go (two sub-commands or two binaries, see S2)
  internal/tui      — today cmd/cgo/tui.go + tui_audit.go (operator dashboard)
                      promoted OUT of package main; Model stays [B1] Init/Update/View
  internal/kittui   — today cmd/cgo/kittui.go + kittui_run.go (kit wizard)
                      promoted OUT of package main; keep the header/steps/items shape
  internal/auditview— auditView + auditFactLines + writeRemediation (today tui_audit.go:24-577)
                      OR fold into internal/tui (single consumer)
  deps: bubbletea v1.x [B2], lipgloss v1.x [L1], stdlib net/http (polling)
        NO internal/kit, NO internal/vm, NO pkg/* — talks to a running cgo
        via HTTP + os.Args[0] kit shim OR via cgo-kit binary (S3)
  data: reads data/audits/<id>/ scripts (contract §2.1); writes ONLY via
        writeRemediation (never /tmp); logs via tea.LogToFile [B1]

cgo-kit   (independent repo, module github.com/<org>/cgo-kit, own go.mod + go floor [G1])
  cmd/cgo-kit       — thin CLI: flag parsing + exit codes (today cmd/cgo/kit.go:15-214)
  kit               — today internal/kit (minus GUI-coupled helpers), promoted OUT of internal/
  vm                — today internal/vm + NEW qemu driver (S3), promoted OUT of internal/
  bench             — kit/testbed.sh + cgo-bench-dualns.sh + run-dualns.sh as data files
                      embedded or shipped, versioned with the binary, NOT a Go API
  consumes: cgo-tui's kit step calls it as `cgo-kit <action>` (or keeps `cgo kit <action>` shim)
  qemu: tracks QEMU upstream releases; QMP via stdlib net + encoding/json ([Q1])
```

Why not "one library `cgo/deploy`" or "one repo `cgo-tools` with two
`cmd/`": the first freezes exit-code semantics as a Go API; the second
keeps the operator monitor coupled to host-admin actions and to the
`KitVersion` string. The two-repo shape keeps the public API where it
actually is: process exit + stdout + yaml diff + HTTP polling + script
files.

### 6.3 Module-path notes ([G1])

- Pick the path once (`github.com/<org>/cgo-tui`, `github.com/<org>/cgo-kit`
  or `cgo.dev/tui`); [G1] warns to use the repository location as the
  import prefix and not to rename after publish. v2+ needs `/v2` suffix —
  start at v0.1.0 (unstable) and do not claim `1.0.0` until the CLI + HTTP
  contract is golden-tested.
- `replace` directives may point at `../cgo-kit` during migration ([G1]
  `replace` section) and must be dropped before tagging.

## 7. Proposed slices (one logical group each, explicit paths)

| Slice | Scope | Acceptance | Gate |
|---|---|---|---|
| S1 — Freeze contracts | Golden-test `cgo tui` keymap + tab order + `auditFactLines` four facts (`tui_audit.go:152-188`), `cgo kit` actions + exit codes 2–8 + `CGO_*` env + `cgo-vm.yaml` keys (`kit.go:23-59` in the shim), script filenames + `data/audits/<id>/` layout (`tui_audit.go:424-506`), kit log name (`kittui.go:140`). Declare API in `docs/cgo-tui-api.md` + `docs/cgo-kit-api.md` per [G1] §"module" (path = ID) | `cmd/cgo -run TestContent` green; `--help` output committed as golden; `go test ./cmd/cgo` in each new repo passes without sibling repos | Full `make verify` six gates; `VERIFY.md` row moves only by independent verifier run |
| S2 — `cgo-tui` extraction (operator first) | Move `cmd/cgo/tui.go` + `tui_audit.go` → `cgo-tui/internal/tui` (promoted out of `package main` per [G2]); `runTUI` becomes `cmd/cgo-tui/main.go`; `CGO_API` env + 1 s/5 s/10 s ticks preserved exactly ([B2] `Tick` re-arm rule); keep 80-col suite (`tui_audit_test.go:250-273`) | Binary `cgo-tui` renders all five tabs + both panes from fixtures; `TestRenderedViewsStayWithinEightyColumns` + `TestWriteRemediation*` green in new repo; no import of `internal/kit`/`internal/vm` | `go test ./...` in new repo; live API row stays BLOCKED (no live audit claimed) |
| S3 — `cgo-kit` extraction | Move `cmd/cgo/kit.go` dispatch + `internal/kit` + `internal/vm` → `cgo-kit` (promoted out of `internal/` per [G2]); `kit`/`vm` import paths updated; `KitVersion` becomes the binary version ([G1] single version); `engine.sh` shim stays in `cgo` repo pointing at `cgo-kit` | `cgo-kit doctor/scan/ensure/deploy` exit codes match frozen table; `cgo kit` shim forwards faithfully; `KitVersion` == tag == `cgo-kit version` | `make verify` six gates; no frozen CSV schema change |
| S4 — Decouple kit TUI | Move `kittui.go` + `kittui_run.go` → `cgo-tui/internal/kittui`; replace direct `internal/kit`/`internal/vm` imports with a `KitClient` interface (`Runner`-shaped, exit-code + log streaming); keep [B1] `Init/Update/View` and header/steps shape ([C1] analogy, limited) | `cgo-tui kit` (or `cgo-tui --kit`) drives a stub `cgo-kit` binary; `kittui_test.go` green against the stub; no `internal/kit` import remains | `go test ./...` in `cgo-tui`; human-approved live run before any bench claim |
| S5 — QEMU/KVM driver | New `vm.Hypervisor` impl `vm/qemu.go` driving `qemu-system-x86_64` ([Q1] `KVM | Linux`, QMP `query-status`, `-netdev user,hostfwd=…` mirroring `NatForwarder`); `Detect` probes `qemu-system-x86_64` presence; CHR cache + Windows-TCG proof per `VERIFY.md` CHR condition | Absent → honest "absent" like `Doctor` (`kit.go:878-897`); present → `Running/Start/GuestIP` via QMP round-trip in test | `go test ./vm -run Qemu` + live CHR boot on a host with QEMU+image (human-approved); BLOCKED rows move only after |
| S6 — TUI hardening | Replace 1 s/5 s/10 s blind ticks with backoff + `WithFPS` cap ([B2]); add `tea.LogToFile` wiring ([B1]); `WindowSize` handling for narrow terminals; keep 80-col hard cut (`tui_audit.go:204-249`) | Fixture re-render at 40/80/120 cols never exceeds 80 in the preserved metrics; no direct HTTP in `View` | `go test ./...` in `cgo-tui`; no live claim |

Live runs need human approval per repo AGENTS.md ("The live bench without
explicit human approval" is never-touch) and `VERIFY.md` how-to-move rules
(rows move to VERIFIED only after the named command runs on the named host).

## 8. Anti-toy criteria (a reviewer applies these; failing any one = toy)

1. **No fake live data.** Tabs 4–5 render only polled `/api/state` +
   `/api/results` + `/api/audit/status`, or the explicit "aucun audit /
   API injoignable" strings (`tui.go:371`, `tui_audit.go:110-121`). A cell
   that shows a number on transport error (`frameMsg{}` → zero frame,
   `tui.go:130-138`) without saying "non mesuré" is a toy ([B1] `View`
   renders the model, never the network).
2. **Four facts, never invented.** `auditFactLines`
   (`tui_audit.go:152-188`) returns "non mesuré / non tentée / non établie"
   strings, never `0.0` from a zero value (proved by
   `TestAuditFactLinesNeverInventAValue`,
   `tui_audit_test.go:101-111`). A renderer that prints `0` for an absent
   latency is a toy.
3. **Verdict never recomputed.** `verdictLine` renders the frozen
   `Verdict.Word/Sentence` (`tui_audit.go:135-146`,
   `TestVerdictHeaderUsesTheReportWord`). Re-scoring in the TUI is a toy.
4. **80 columns hard.** `wrap78`/`wrapWidth` (`tui_audit.go:204-249`) and
   `TestRenderedViewsStayWithinEightyColumns` pin every line ≤80. A pane
   that scrolls horizontally or wraps a `tc` command across two commands
   (`TestDeploymentPaneOneCommandPerLine`) is a toy. ([L1] gives `Wrap`,
   not the budget — the budget is the project's.)
5. **Scripts never empty, paths never escape.** `scriptFor` honest errors
   (`tui_audit.go:454-477`), `safeAuditID` + prefix check
   (`tui_audit.go:433-506`), `data/audits` only
   (`TestWriteRemediationStaysUnderDataAudits`,
   `tui_audit_test.go:320-341`). A `c` key that writes an empty file or a
   file under `/tmp` is a toy.
6. **OSC 52 never claimed.** `copyScript` says "tentée", never "copiée"
   (`tui_audit.go:569-577`). A status that asserts delivery is a toy.
7. **Bubble Tea contract intact.** `Init`/`Update`/`View` ([B1]),
   `Cmd func() Msg` + `Tick` re-arm ([B2]), stdout owned by the renderer
   ([B1] "can't really log to stdout" → `kit-tui-<date>.log` file,
   `kittui.go:114-158`). A `View` that does HTTP or a logger that prints to
   stdout is a toy.
8. **Honest exit codes.** Every `Runner` method returns a documented code
   from the 2–8 table (`kit.go:58-59`, shim). No `return 0` after a caught
   error. Fault proof: stub `ssh` failing → `Ensure` returns 5
   (`kit.go:1126-1127`), stub `vmrun` absent → `Scan` returns 3
   (`kit.go:914-917`).
9. **NAT verified, not assumed.** `EnsureNatSSH` re-reads `showvminfo`
   after posing (`vm.go:638-641`); `Ensure` probes the listener
   (`kit.go:1027-1029`). A rule string without `natRulePresent` is a toy.
   No topology guesses: NAT only when `NetMode == "nat"`
   (`kit.go:1008-1031`); guest `10.0.2.x` never written to a host adapter.
10. **No silent config mutation.** `requireLockUser` (`kit.go:295-305`)
    gates late actions before any write; yaml writes go through
    `writeFileAtomic` (`kit.go:218-246`). A slice that writes
    `cgo-vm.yaml` outside the three writers fails.
11. **Live-bench claims stay BLOCKED until the command runs.** Per
    `VERIFY.md`, Windows unit green never promotes Linux live, eBPF, IFB,
    BBRv3, `.dist`, CHR, or PC2/nemo rows. The verifier re-runs and owns
    the row (builder ≠ verifier).
12. **QEMU is real or absent.** `Detect` returns only executables actually
    present (`vm.go:819-852`, `findExe` + `exec.LookPath`); a stub in tests
    must be labelled `fakeHyp` and never ship as a "driver". Proved by a
    live `VBoxManage list`/`vmrun list` capture in the checkpoint.
13. **Version identity is one thing.** `KitVersion` (`kit.go:52`), git tag,
    and `cgo version` must agree; a release where `cgo version` ≠ tag ≠
    `KitVersion` is a toy release. ([G1] one version per module — the
    current three-copy drift is the gap, not the model.)

## 9. Caveats

- VMware `vmrun` CLI surface: **unverified — primary source not found**.
  All `vmrun` semantics above are local-source evidence only.
- QEMU/KVM beyond [Q1] Introduction (QMP command list, KVM ioctl ABI,
  `-machine accel=help` probing) is **unverified — primary source not
  found** at the granularity needed for slice S5 and must be fetched before
  implementation.
- Chrome Task Manager ([C1]) gives the user flow only; per-process metering,
  tree rendering, and "End Process" semantics beyond the quoted steps are
  **unverified — primary source not found**. The [C1] analogy in this
  document is limited to the header + actionable-list shape.
- `docs/deploy.md` and `docs/VERIFY.md` rows are project-local claims, not
  external primary sources; they are cited by line as evidence, and BLOCKED
  rows remain BLOCKED until the named host condition clears.
- This report revisits `2026-11-07-module-kit-vm-profile.md` for kit/vm and
  adds the missing TUI layer. Proposals in §6–§8 are design options gated
  by the criteria in §8; nothing here is implemented, and no live run,
  measurement, or benchmark was performed for this research.
