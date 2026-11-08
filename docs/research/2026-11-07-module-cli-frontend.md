# Module CLI + Frontend — `cmd/*` + `web/frontend` as standalone tools

Date: 2026-11-07. Scope: `cmd/cgo`, `cmd/meteolink`, `cmd/cgo-chr`, `cmd/cgo-gui`, `web/frontend` (React+Vite).
Method: read `cmd/*/main*.go`, `web/frontend/src/App.tsx`, `package.json`, `docs/architecture.md` CLI/Modes sections, `go.mod`, `.goreleaser.yaml`, `Makefile`, `internal/kit/kit.go`, `docs/VERIFY.md` / `docs/CHECKPOINTS.md` evidence. External standards only via primary sources below. No code, no commit.

## Primary sources (R)

- R1 Go `go.mod` file reference — https://go.dev/doc/modules/gomod-ref :
  module path is the module's unique identifier + import prefix, v2+ must end `/v2`; `go` is a mandatory minimum since 1.21; `require` is `path version` (SemVer `vX.Y.Z` tags or pseudo-versions); `replace`/`exclude`/`retract` semantics; single `go.mod` versioning.
- R2 Semantic Versioning 2.0.0 — https://semver.org/ :
  `X.Y.Z` where MAJOR = incompatible API, MINOR = backward-compatible feature, PATCH = backward-compatible bug fix; once released contents MUST NOT change; 1.0.0 defines the public API; public API MUST be declared.
- R3 OCI — https://opencontainers.org/ and https://github.com/opencontainers/image-spec :
  OCI is an open governance structure for container formats/runtimes with three specs (Runtime, Image, Distribution); high-level flow is download OCI Image → unpack to OCI Runtime filesystem bundle → run by OCI Runtime; image-spec repo provides Go types + JSON Schema + intra-blob validation; sufficient info to launch app (command, args, env).
- R4 Cobra — https://github.com/spf13/cobra (README):
  library for modern CLIs similar to git/go tools; easy subcommands `app server`, fully POSIX-compliant flags (short & long), nested subcommands, global/local/cascading flags, intelligent suggestions, automatic help generation, automatic shell completion (bash/zsh/fish/powershell), automatic man pages, command aliases; flags are modifiers on `cobra.Command`, provided via pflag fork.
- R5 urfave/cli — https://github.com/urfave/cli (README):
  declarative, simple, fast package for CLIs; commands and subcommands with alias/prefix match, flexible help, dynamic shell completion (bash/zsh/fish/powershell), no dependencies except Go stdlib, flags for simple types/slices/time/duration, compound short flags, man + Markdown docs generation; v3 module path `github.com/urfave/cli/v3`.
- R6 Vite — https://vite.dev/guide/ :
  build tool with dev server (native ES modules, HMR) and build command bundling with Rolldown for optimized static assets; opinionated defaults; requires Node.js 20.19+ / 22.12+; `index.html` is the app entry and part of the module graph; `vite.config` adaptation; official `@vitejs/plugin-react`.
- R7 React — https://react.dev/learn :
  React apps are components (JS functions returning markup); JSX markup; `className`, curly-brace data display, conditional/list rendering, event handlers passed without calling, state via `useState`, hooks only at top level, shared state lifted up and passed down as props.

Repo-internal claims cite `path:line` below. Anything not observed in repo and not in R1–R7 is `unverified — primary source not found`.

## 1. Module identity and versioning today

- Single Go module `module github.com/Realms4239/cgo` and `go 1.25` — `go.mod:1-3`. Per R1, the module path is the unique identifier + import prefix for all packages; there is no nested `go.mod` under `cmd/` or `pkg/`, so all four binaries + frontend embed share one versioned module.
- Direct requires are only four — `go.mod:5-10`: `fyne.io/fyne/v2 v2.8.1`, `github.com/charmbracelet/bubbletea v1.3.10`, `github.com/charmbracelet/lipgloss v1.1.0`, `golang.org/x/sys v0.36.0`; remainder `go.mod:12-60` is indirect (Fyne GL, systray, text, term, etc.). Per R1, `require` records minima for every module providing imported packages; a consumer importing one subpackage still resolves the whole build list.
- Canonical version constant `KitVersion = "1.3.5"` — `internal/kit/kit.go:49-52` with comment: single source, `cmd/cgo version` defaults to it and GUIs compare remote dashboard to it. `VERSION` file is `1.3.5` — `VERSION:1`. Frontend `package.json` `version` is `"1.3.5"` — `web/frontend/package.json:4`. `internal/kit/kit.go:52` and `web/frontend/package.json:4` must be bumped together by hand; `cmd/cgo/main.go:36-41` warns that a `-X main.version` release must equal the tag (= KitVersion) or CLI (`kit next`) and GUI diverge. Per R1, version identity is `module path + version` (tags are `vX.Y.Z`); per R2, `1.3.5` parses as MAJOR 1, MINOR 3, PATCH 5.
- No public-API stability declaration was found in `cmd/*/main*.go`, `web/frontend/package.json`, or `docs/architecture.md`. Per R2 §5, 1.0.0 defines the public API and subsequent increments depend on how it changes; without a declared API (here CLI flags + `pkg/api` HTTP/SSE + frontend `/api` contract), MAJOR/MINOR/PATCH meaning for 1.3.5 cannot be verified — `unverified — primary source not found`.

## 2. Responsibility per binary

### 2.1 `cmd/cgo` — the monolith (CLI + server + TUI + kit + audit + campaign)

- Header declares the roles in French — `cmd/cgo/main.go:1-11`: `cgo --serve` dashboard + API, `setup` wizard, `kit` deployment engine, `audit`, `run`, `tui`, `verify`, `figures`. Architecture doc states the matching ownership: "Binaire unique Go embarquant le dashboard React via `go:embed`" with skeleton `cmd/cgo (CLI + serveur)` and `pkg/` — `docs/architecture.md:1-4`.
- `main()` is a manual dispatch on `os.Args[1]` via `switch` — `cmd/cgo/main.go:43-302`, no CLI framework imported (stdlib `flag.NewFlagSet` per subcommand — `cmd/cgo/main.go:56,77,140,149,280`). This matters for §5 (R4/R5 comparison): the surface is hand-rolled, not generated.
- Sub-responsibilities (each is a separate file in `cmd/cgo/`, all `package main`):
  - `--serve` / `runServer` — HTTP+HTTPS dashboard + REST/SSE API, SSE 10 Hz hub, watch vs campaign mutual exclusion, matrix lifecycle — `cmd/cgo/server.go:50-391` (`runServer` at `cmd/cgo/server.go:54`), flags `--addr/--tls/--http-addr/--mode` at `cmd/cgo/main.go:56-65`, env `CGO_IFACE/CGO_IFACE_SOURCE` documented in `usage()` `cmd/cgo/main.go:360-361` and sampler `cmd/cgo/server.go:34-48`.
  - `setup` — 7-step wizard (detect → deps → build → context → VM config → DNS → verify) — `cmd/cgo/setup.go:17-229`, flags `--yes/--no-vm` via mini-flagset `cmd/cgo/setup.go:324-336`.
  - `kit <action>` — deployment engine, 21 actions claimed in header `cmd/cgo/kit.go:12-14` (`doctor, scan, keysetup, ensure, align, build, deploy, svc, testbed, dns, tls, shipcheck, package, bootstrap, status, logs, tunnel, snapshot, revert, snapshots, ssh, netinfo, nic, vnet, next, guest, readme, tui, ps, backup, verify, health, schedule` — usage string `cmd/cgo/kit.go:23-59`, dispatch `cmd/cgo/kit.go:94-214`); flags `--config/--deep/--shallow/--public/--yes` — `cmd/cgo/kit.go:16-21`; exit codes 2–8 documented `cmd/cgo/kit.go:58-59`, propagated from `internal/kit` `Runner` (e.g. `Deploy` at `internal/kit/kit.go:1372-1397`, `Build` at `internal/kit/kit.go:1327-1356`).
  - `tui|top` — Bubble Tea + Lip Gloss terminal dashboard, 5 tabs Setup/Kit/Campagne/Live/Résultats — `cmd/cgo/tui.go:395-405` (`runTUI`), model `cmd/cgo/tui.go:68-87`, polling `cmd/cgo/tui.go:127-159`; `kit tui` is a distinct interactive centre — `cmd/cgo/kittui.go` + `cmd/cgo/kittui_run.go` (not `cmd/cgo/tui.go`); `tui_audit.go` mirrors last sealed audit.
  - `audit` — field audit of a real link, seals telemetry + `link_audit.csv` + report, exit 2 on partial (unmeasured bufferbloat/economics/throughput) — `cmd/cgo/main.go:148-278`, `printAuditSummary` at `cmd/cgo/main.go:315-357`.
  - `run` — CLI matrix campaign, one progress line per frame, graceful Ctrl-C freeze, résumé from frozen CSV — `cmd/cgo/runcli.go:19-175`, file lock `cmd/cgo/runcli.go:48-65`, preflight `cmd/cgo/runcli.go:183-218`, flags `--profiles/--qdiscs/--cc/--direction/--reps/--flows/--deadline/--target/--data/--run-id/--bench/--channel-model` — `cmd/cgo/main.go:280-293`.
  - `doctor` — capability report — `cmd/cgo/main.go:72-73` via `pkg/doctor`.
  - `version` — `meteolink (cgo) <version> — <goos>/<goarch>, mode auto` — `cmd/cgo/main.go:74-75`, `version` var `cmd/cgo/main.go:41`.
  - `shape --restore` — remove root qdiscs posed by meteolink, refuses on Windows — `cmd/cgo/main.go:76-93`.
  - `service install` — writes systemd unit `ExecStart=/usr/local/bin/cgo --serve --addr 0.0.0.0:9090 --tls=false --http-addr=` — `cmd/cgo/main.go:94-121`, Linux-only guard `cmd/cgo/main.go:99-102`.
  - `figures` — regenerate SVGs from frozen CSVs — `cmd/cgo/main.go:122-127`.
  - `verify` — manifest SHA-256 integrity over `data/runs`, every file must be covered — `cmd/cgo/main.go:128-138`, impl `cmd/cgo/main.go:375-431`.
  - `testbedsrv` — small-object HTTP + bulk sink — `cmd/cgo/main.go:139-147`, impl `cmd/cgo/testbedsrv.go`.
  - `calibrate` — trace → `.dist` netem — `cmd/cgo/main.go:297-298` → `cmd/cgo/calibrate.go:23` (`flag.NewFlagSet("calibrate"...)`).
- Stubs file exists for cross-compile — `cmd/cgo/stubs.go` (build-tagged stubs; content not needed for surface beyond proving `package main` builds on Windows).

### 2.2 `cmd/meteolink` — compat alias, re-executes `cgo`

- Header: "alias du binaire unique cgo (même noyau, nom compatible)" — `cmd/meteolink/main.go:1-5`. The alias gains backward-compatible script/npm names; usage lists `top/tui/setup/kit/run/serve/version … tout cgo, même moteur` — `cmd/meteolink/main.go:1-5`.
- `main()` resolves `os.Executable()` dir, appends `cgo.exe` on Windows, checks sibling then `PATH`, else errors honestly — `cmd/meteolink/main.go:18-39`; `passExit` re-executes with identical args/stdin/stdout/stderr and propagates exit code — `cmd/meteolink/main.go:41-54`.
- No business logic, no `version` var: `.goreleaser.yaml:20-33` builds it with `ldflags: -s -w` only, with comment that linker would refuse `-X main.version` because this main has no such var — `.goreleaser.yaml:20-22`.
- Claim "meteolink becomes an embedded alias (`cgo top` historic + wrapper)" — `.goreleaser.yaml:5-7` shows historical intent, but the shipped binary is still a separate re-exec wrapper, not embedded in `cgo`. Treat the comment as intent, not fact.

### 2.3 `cmd/cgo-chr` — auxiliary, not shipped: MikroTik CHR RouterOS proof

- Header: "auxiliary tool, not part of the shipped binary: it boots the MikroTik CHR image in QEMU, feeds it the sealed report's RouterOS script, and seals proof that RouterOS 7 accepted it. Every judgement lives in `pkg/audit`; this file only orchestrates." — `cmd/cgo-chr/main.go:1-5`.
- `RunConfig` (QEMU/Image/Level/Timeout) — `cmd/cgo-chr/main.go:14-19`; exported `BuildQEMUArgs` returns the exact QEMU command line for both backends, accelerator = evidence — `cmd/cgo-chr/main.go:21-39` (tcg vs kvm by level `linux-namespace`).
- Flags — `cmd/cgo-chr/main.go:42-50`: `--report`, `--image`, `--level` (`windows-tcg|linux-namespace`), `--timeout` (180s), `--out` (default `data/chr/<audit_id>`), `--qemu` (`qemu-system-x86_64`), `--export-file` (replay mode without QEMU), `--dry-run` (prints command). `levelAllowed` refuses `linux-namespace` off Linux early to avoid mislabeled proof — `cmd/cgo-chr/run.go:155-165`.
- `run()` — validates level, reads sealed `audit_report.json`, requires at least one RouterOS script, runs QEMU transport or replay, compares expected vs found export via `audit.ParseRouterOSConfig`/`audit.CompareRouterOS`, seals `CHRProof` via `audit.WriteCHRProof` — `cmd/cgo-chr/run.go:19-107`, `startQEMU` — `cmd/cgo-chr/run.go:109-128`; not in `.goreleaser.yaml:8-33` (only `cgo` + `meteolink`) and not in `Makefile:6-7` (`go build ./cmd/cgo` only) — consistent with "not shipped".

### 2.4 `cmd/cgo-gui` — native control centre, dual view (Win32 + Fyne)

- Shared core, two views: header of `core.go` states "Coeur applicatif PARTAGÉ (Windows + Linux) : aucune API Win32, aucun widget — que du pur Go (kit, vm, stdlib). Les deux vues pilotent les mêmes actions" — `cmd/cgo-gui/core.go:1-4`.
- Windows view is `//go:build windows` — `cmd/cgo-gui/main.go:1`, raw Win32 via `golang.org/x/sys/windows` and LazyDLL `user32.dll`/`gdi32.dll` — `cmd/cgo-gui/win32.go:3-41` (comment: x/sys does not wrap all of user32/gdi32), window creation + pump on one OS thread with `runtime.LockOSThread()` — `cmd/cgo-gui/window.go:20-73` (comment: without it Go migrates goroutine and PostMessage freezes). Window size `winW=940/winH=700` and control IDs — `cmd/cgo-gui/main.go:21-34`. Actions emit `cgo.exe kit …` next to GUI and stream output — `cmd/cgo-gui/main.go:3-7`; interactive actions open visible console — `cmd/cgo-gui/actions.go:217-270`.
- Linux view is `//go:build linux` — `cmd/cgo-gui/fygui_linux.go:1`, using `fyne.io/fyne/v2` + `app` + `container` + `widget` — `cmd/cgo-gui/fygui_linux.go:19-22`; same buttons/journal/anti-double-click, only rendering changes — `cmd/cgo-gui/fygui_linux.go:3-6`; refuses without `DISPLAY`/`WAYLAND_DISPLAY`, redirecting to `cgo kit tui` — `cmd/cgo-gui/fygui_linux.go:731-735`.
- Shared button inventory (IDs + verb mapping): groups `groupActions/accessActions/deployActions/controlActions` — `cmd/cgo-gui/core.go:75-103`; reverse map `buttonAction(id) → (kind,args)` — `cmd/cgo-gui/core.go:247-313` with kinds `direct:`, `bg:`, `console:`; `isKnownSuiteVerb` — `cmd/cgo-gui/core.go:112-123`. Windows dispatch `runKind` — `cmd/cgo-gui/actions.go:381-518`, Linux dispatch `dispatchKind` — `cmd/cgo-gui/fygui_linux.go:420-562`; both share `suiteStep/suiteContinue` anti-loop logic — `cmd/cgo-gui/actions.go:520-589` + `cmd/cgo-gui/fygui_linux.go:564-624`.
- GUI is not a CLI: no `flag` surface; entrypoint only locates `cgo.exe` beside it (fatals if absent) — `cmd/cgo-gui/main.go:79-107`; Linux `main()` locates sibling `cgo` similarly — `cmd/cgo-gui/fygui_linux.go:731-747` (`cgoExe` set at `cmd/cgo-gui/fygui_linux.go:643-645`).

### 2.5 `web/frontend` — React+Vite SPA, embedded in `cgo`

- `go:embed` package — `web/frontend/embed.go:1-21`: `//go:embed all:dist`, `FS` serves dist at root, `dist/.gitkeep` guarantees build before first frontend build — `web/frontend/embed.go:1-3`.
- Vite build — `web/frontend/vite.config.ts:1-8`: `@vitejs/plugin-react`, dev proxy `'/api' → https://localhost:9090, secure:false` (comment: backend local self-signed), `build.outDir='dist', sourcemap:false`. Per R6, Vite's build bundles with Rolldown for optimized static assets; here the output is consumed by Go embed, not deployed as standalone static site.
- Router (not react-router): four destinations modelled as `MODES` — `web/frontend/src/store/ui.ts:7-12` (`dashboard/audit/bench/evidence`, keys 1–4). `App.tsx` lazy-loads heavy views only on first visit and keeps them mounted — `web/frontend/src/App.tsx:21-25` (comment: landing only in first render, Bench carries ECharts, audit/proofs carry panels) + `web/frontend/src/App.tsx:45-49` (`mounted` state).
- Shell (`App.tsx` default export `App` — `web/frontend/src/App.tsx:35-178`): `connectSSE/disconnectSSE` on mount — `web/frontend/src/App.tsx:51-55`; document title + focus + `animateViewEnter` on mode change — `web/frontend/src/App.tsx:57-70`; campaign end ghost cleanup (`clearLive`, reset live frame, `stageAfterCampaign`, flash/toast) — `web/frontend/src/App.tsx:72-98`; `stageAfterReplay` — `web/frontend/src/App.tsx:100-104`; global keys `Ctrl/Cmd+K` palette, `1–4` modes, ignore text-entry targets — `web/frontend/src/App.tsx:106-122`; error boundary + header (`MeteolinkWordmark`, `TopNav`, connection dot, phase, last audit, settings) — `web/frontend/src/App.tsx:124-142`; views hidden not unmounted — `web/frontend/src/App.tsx:147-164`; footer provenance (`profile·qdisc·cc`) — `web/frontend/src/App.tsx:166-169`; drawers/toasts — `web/frontend/src/App.tsx:171-174`.
- Per R7, this is standard React: function components, hooks (`useState/useEffect/useRef`), props drilling, composed components (`TopNav, FlashBanner, ProofDrawer, CommandPalette, QuickActionsPrompt`, `ui/` primitives `Stat/MetricCard/Card/...`). Entry is `StrictMode` + `createRoot` — `web/frontend/src/main.tsx:1-10`.
- View/component inventory (from `src/` listing): views `DashboardView, AuditView, BenchView, LiveView, EvidenceView, ResultatsView, IntegriteView`; `lib/` (`types, sse, live, bench, format, anime, modalFocus, preload, trace, validation, viewState, explain, echarts, designFloor, tokens, title, statStrip, tui`); `store/` (`ui, audit, bench`); `components/ui/` primitives + 40+ domain components (`Audit*, Bench*, Duel*, Proof*, VerdictHero, KernelRing, TraceCanvas...`). This is task-relevant only as surface size: frontend is not "a page" but a 4-destination app mirroring `cgo --serve` API.

## 3. Exported API / CLI surface (file:line)

### 3.1 `cgo` CLI (`cmd/cgo/main.go`)

| Verb | Entry | Flags / args | Exit codes |
|---|---|---|---|
| (none) | `usage(); exit 2` — `cmd/cgo/main.go:44-47` | — | 2 |
| `kit` | `runKit(args[2:])` — `cmd/cgo/main.go:49-50` → `cmd/cgo/kit.go:15` | `--config/--deep/--shallow/--public/--yes` + action name — `cmd/cgo/kit.go:16-62`; sub-actions dispatch `cmd/cgo/kit.go:94-214`; `svc <start\|stop\|restart\|status>`, `schedule --at/--profiles/--qdiscs/--cc/--direction/--reps` — `cmd/cgo/kit.go:186-209` | 0 ok, 2 usage/build, 3 scan, 4 hypervisor, 5 SSH timeout, 6 cross/bootstrap, 7 scp, 8 install/logs — `cmd/cgo/kit.go:12-14,58-59` |
| `setup` | `runSetup` — `cmd/cgo/main.go:51-52` → `cmd/cgo/setup.go:21` | `--yes/--no-vm` — `cmd/cgo/setup.go:324-336` | 0 ok, 2 missing deps/declined — `cmd/cgo/setup.go:76-78,91-99` |
| `tui` / `top` | `runTUI` — `cmd/cgo/main.go:53-54` → `cmd/cgo/tui.go:398` | none (env `CGO_API`, default `https://meteolink.dev:9090` — `cmd/cgo/tui.go:89-95`) | 1 on tea error — `cmd/cgo/tui.go:399-403` |
| `--serve` | `runServer` — `cmd/cgo/main.go:55-71` → `cmd/cgo/server.go:54` | `--addr` (default `meteolink.dev:9090` via `CGO_DASHBOARD__ADDR` — `cmd/cgo/main.go:57-61`), `--tls` (default true), `--http-addr` (default `127.0.0.1:9080`), `--mode` (`auto/observe/full`, OS decides) — `cmd/cgo/main.go:62-64` | 1 on serve error — `cmd/cgo/main.go:68-71` |
| `doctor` | `doctor.Print()` — `cmd/cgo/main.go:72-73` | none | 0 |
| `version` | `fmt.Printf("meteolink (cgo) %s — %s/%s, mode %s\n",...)` — `cmd/cgo/main.go:74-75` | none | 0 |
| `shape` | manual qdisc delete — `cmd/cgo/main.go:76-93` | `--restore` (only value; without it usage exit 2 — `cmd/cgo/main.go:80-83`); Windows refusal — `cmd/cgo/main.go:84-87` | 1 Windows, 2 usage |
| `service` | systemd unit write — `cmd/cgo/main.go:94-121` | only `install` accepted — `cmd/cgo/main.go:95-98`; Linux-only — `cmd/cgo/main.go:99-102` | 1/2 |
| `figures` | `figures.Generate("data/runs","data/figures")` — `cmd/cgo/main.go:122-127` | none (hardcoded paths) | 1 |
| `verify` | `verifyManifests("data/runs")` — `cmd/cgo/main.go:128-138` | none (hardcoded) | 1 (details `cmd/cgo/main.go:375-431`: missing/illegitimate manifest, sha mismatch, uncovered file) |
| `testbedsrv` | `testbedSrvFn(*http,*bulk)` — `cmd/cgo/main.go:139-147` | `--http` default `10.200.0.1:8081`, `--bulk` default `10.200.0.1:5201` | 1 |
| `audit` | `audit.Run` + seal + report — `cmd/cgo/main.go:148-278` | operator four (`--link-type/--site/--duration/--direction/--provider/--flows`) — `cmd/cgo/main.go:151-156`; advanced overrides (`--target/--small/--bulk`) — `cmd/cgo/main.go:161-163`; remediation (`--iface/--routeros-iface/--ros-lan/--report/--qdisc/--asymmetric/--cpu-constrained`) — `cmd/cgo/main.go:167-173` | 0 complete, 1 error, **2 partial** (report exists but bufferbloat/economics/throughput unavailable — `cmd/cgo/main.go:266-278`) |
| `run` | `runCLI(...)` — `cmd/cgo/main.go:279-296` → `cmd/cgo/runcli.go:23` | `--profiles/--qdiscs/--cc/--direction/--reps/--flows/--deadline/--target/--data/--run-id/--bench/--channel-model` — `cmd/cgo/main.go:281-292` | 2 usage (qdiscs+cc mismatch `cmd/cgo/runcli.go:40-43`, direction `cmd/cgo/runcli.go:70-75`, bench `cmd/cgo/runcli.go:81-88`, flows `cmd/cgo/runcli.go:103-106`), 3 double-run lock — `cmd/cgo/runcli.go:48-65`, 4 bench not ready — `cmd/cgo/runcli.go:183-218` |
| `calibrate` | `runCalibrate` — `cmd/cgo/main.go:297-298` → `cmd/cgo/calibrate.go:23` | `flag.FlagSet("calibrate")` — `cmd/cgo/calibrate.go:23` (full flag set not enumerated here; usage string `cmd/cgo/main.go:365`) | per calibrate.go |
| `usage` | printed on unknown/none — `cmd/cgo/main.go:359-373` (canonical `--serve/audit/run/verify/calibrate/figures/doctor/shape/service/version/testbedsrv` list) | — | 2 |

HTTP API surface (server, not CLI): routes live in `pkg/api/server.go` (REST) + `pkg/api/sse.go` (hub 10 Hz) per `docs/architecture.md:10-12`; server wiring `cmd/cgo/server.go:294-337` (`GetSnap/NextSnap/StartFn/StopFn/SkipFn/ShapeFn/RunningFn/Mode/Version/DoctorFn/BurstFn/WatchFn`). Per `docs/architecture.md:23-41`: `GET /api/stream` delta compression, 2048-frame ring, `Last-Event-ID` replay, 5-frame warm-up without header, 64-sub cap → 503, `retry:2000`, no manual `Connection` header (illegal on HTTP/2). Per `docs/architecture.md:85-93`: Linux=full, Windows=observe (shape/run/watch 501), `--mode observe|full|auto`; `GET /api/events` (50 ring), `GET /api/diagnostics` `{hub,time}`, `GET /api/health`. Manual shape `POST /api/shape` refused 409 during campaign — `docs/architecture.md:55-58`.

### 3.2 `meteolink` CLI (`cmd/meteolink/main.go`)

Keeles: no flags, no subcommands; `exec.Command(path, os.Args[1:]...)` with stdio forwarded — `cmd/meteolink/main.go:42-46`, exit code propagated via `exec.ExitError.ExitCode()` — `cmd/meteolink/main.go:47-52`. Discovery order: sibling `cgo` → sibling `cgo.exe` (Windows) → `cgo` in PATH → error — `cmd/meteolink/main.go:24-38`.

### 3.3 `cgo-chr` CLI (`cmd/cgo-chr/main.go`)

Flags `cmd/cgo-chr/main.go:42-50` (stdlib `flag`): `--report=""`, `--image=""`, `--level="windows-tcg"`, `--timeout=180s`, `--out=""`, `--qemu="qemu-system-x86_64"`, `--export-file=""`, `--dry-run=false`. Validation: `--report` required unless dry-run — `cmd/cgo-chr/main.go:57-62`; `--image` required without `--export-file` — `cmd/cgo-chr/main.go:60-62`. Exported `BuildQEMUArgs(RunConfig) []string` — `cmd/cgo-chr/main.go:23-39` (deterministic QEMU line used as evidence; first element is binary, stripped at `cmd/cgo-chr/run.go:113`). `levelAllowed` — `cmd/cgo-chr/run.go:157-165`.

### 3.4 `cgo-gui` surface (no CLI)

No Go `flag`/`cobra`/`urfave` surface. Surface = window controls with fixed IDs — `cmd/cgo-gui/main.go:26-34`: `idVMList=101, idLog=102, idStatus=103, idLocked=104, idSSH=105, idDash=106, idUser=107, idNext=108`; layout built in `cmd/cgo-gui/window.go:177-237` (Win32) and `cmd/cgo-gui/fygui_linux.go:648-729` (Fyne, `winW/H 940×700` — `cmd/cgo-gui/fygui_linux.go:650`). Logical actions are the `controlActions` list — `cmd/cgo-gui/core.go:94-103` (22 buttons: État/Start/Stop/Restart/Logs/DNS/TLS/Vérifier/Backup/Snapshot/Réseau invité/Démarrer VM/Arrêter VM/Carte NAT-pont/Tunnel hôte/Invité/Réseau hôte/Guide/Suite/Banc). Button→verb mapping `cmd/cgo-gui/core.go:247-313`. Single-flight background runners `runKit` (Win32 `cmd/cgo-gui/actions.go:188-215`), `runBg` (Linux `cmd/cgo-gui/fygui_linux.go:200-241`), interactive `runKitConsole`/`runTerm` (`cmd/cgo-gui/actions.go:217-270`, `cmd/cgo-gui/fygui_linux.go:243-317`). Log mirroring `cgo-gui-<date>.log` beside exe else `%TEMP%` — `cmd/cgo-gui/core.go:23-66`.

### 3.5 `web/frontend` surface

- npm scripts — `web/frontend/package.json:6-13`: `dev=vite`, `build=tsc --noEmit && vite build && node scripts/check-bundle.mjs`, `typecheck=tsc --noEmit`, `test=vitest run`, `test:e2e=playwright test`.
- Dependencies — `web/frontend/package.json:15-21`: `react ^19.0.0`, `react-dom ^19.0.0`, `zustand ^5.0.0`, `echarts ^5.5.0`, `animejs ^4.5.0`. Dev — `web/frontend/package.json:22-33`: `vite ^6.0.0`, `@vitejs/plugin-react ^4.0.0`, `typescript ^5.7.0`, `vitest ^3.2.7`, `@playwright/test ^1.62.1`, `jsdom`, `@types/...`. Per R6, Vite requires Node 20.19+/22.12+ (compatibility note); `setup.go` checks `go/bun/node/ssh` — `cmd/cgo/setup.go:48-58` and `kit Doctor` checks `go/bun/node/ssh/scp/curl` — `internal/kit/kit.go:823-832`.
- Bundle budget gate — `web/frontend/scripts/check-bundle.mjs:35-36`: `TOTAL_GZ_MAX=600*1024`, `ECHARTS_GZ_MAX=350*1024`; build fails above budget — `web/frontend/scripts/check-bundle.mjs:65-69`. Comment calibration history — `web/frontend/scripts/check-bundle.mjs:1-36` (floor 542.5 KB gz, recalibrated to 600 KB for wall-kit + heterogenous P3 breadth). ECharts watchdog is weakened: since single-chunk wall-kit build, echarts lives inside main chunk — `web/frontend/scripts/check-bundle.mjs:52-58`.
- API contract (unversioned): dev proxy only — `web/frontend/vite.config.ts:6` `'/api' → https://localhost:9090, secure:false`; production is same-origin embed (`FS` at root — `web/frontend/embed.go:14-21`). No `/api` version in URL in observed code; server `Version` is passed as `version` var — `cmd/cgo/server.go:303`. Exact REST paths are in `pkg/api`, not frontend `package.json`; frontend calls them via `lib/sse` (`connectSSE`) + `lib/api`-adjacent clients and Zustand stores (`store/ui.ts:38-56`, `store/audit.ts`, `store/bench.ts`).

## 4. Deps (who needs what)

### 4.1 Go (per R1: `module` + `go` + `require`)

Joined constraints from R1: (a) module path = import prefix; (b) `go` is mandatory minimum; (c) every non-stdlib import needs a matching `require`. Observed:

| Component | Direct imports (stdlib vs require) | Consequence |
|---|---|---|
| `cmd/cgo` (monolith) | stdlib `flag/syscall/os/exec/...` + `internal/kit`, `pkg/*`; **plus** `bubbletea` (`cmd/cgo/tui.go:14`, `cmd/cgo/kittui.go:19`, `cmd/cgo/tui_audit.go:13`, `cmd/cgo/kittui_run.go:20`) + `lipgloss` (`cmd/cgo/tui.go:15`, `cmd/cgo/kittui.go:20`) | Whole `cmd/cgo` binary carries TUI deps even for `--serve`/`audit`/`run` which never use them — because single module. Per R1, minimal version selection resolves `go.mod:7-8` for the module, not per-subcommand. |
| `cmd/meteolink` | stdlib `fmt/os/os-exec/path/runtime` only — `cmd/meteolink/main.go:8-14` | Zero `require` needed, but still inherits module's four requires via `go.mod`. Standalone module would drop all four. |
| `cmd/cgo-chr` | stdlib `flag/fmt/os/time` + `pkg/audit` — `cmd/cgo-chr/main.go:7-12`, `cmd/cgo-chr/run.go:3-12` | Zero `require` needed. Same inheritance as above under monolith. |
| `cmd/cgo-gui` (Windows) | stdlib + `internal/kit` + `golang.org/x/sys/windows` — `cmd/cgo-gui/main.go:10-19`, `cmd/cgo-gui/win32.go:7-11`, `cmd/cgo-gui/window.go:5-13`, `cmd/cgo-gui/elevate.go:9` | Needs `x/sys v0.36.0` (`go.mod:9`). No fyne on Windows path (files are `//go:build windows`). |
| `cmd/cgo-gui` (Linux) | stdlib + `internal/kit` + `fyne.io/fyne/v2{,/app,/container,/widget}` — `cmd/cgo-gui/fygui_linux.go:9-25` | Needs `fyne v2.8.1` (`go.mod:6`) + its transitive `go-gl/gl`, `glfw`, `image`, `oksvg`, `systray` (`go.mod:12-60`). Per R1, this is why full `go.mod` carries GL deps for every binary. |
| `internal/kit` + `internal/vm` (shared by cgo/gui) | stdlib + `golang.org/x/sys/windows` on Windows files (`internal/kit/exec_windows.go:9`, `internal/vm/execcmd_windows.go:9`, `internal/vm/drives_windows.go:5`) | Windows build also carries `x/sys`; Linux build of `kit` drops it. |
| `pkg/probe` (Linux TCP_INFO) | stdlib `syscall` + raw `SYS_GETSOCKOPT` — `pkg/probe/tcpinfo_linux.go:30-37` (comment rejects `x/sys` in favour of raw syscall) | Deliberately avoids `x/sys` for TCP_INFO; shows preference for stdlib-only Linux primitives. |
| `pkg/api|doctor|iface|campagne|qdisc` | stdlib only per architecture doc — `docs/architecture.md:10-12` ("stdlib uniquement"), campaign execs `tc` via subprocess (`docs/architecture.md:95-103`: no CGO, cross-compilable) | These are the only candidates for standalone server module with zero `require`. |

Fyne Linux system prerequisites (GCC, X11/Wayland, OpenGL, pkg-config dev packages) are **not documented in-repo** and Fyne docs were not fetched — `unverified — primary source not found`.

### 4.2 Frontend (per R6/R7)

- Runtime: React 19 + ReactDOM 19 — `web/frontend/package.json:18-19`. Per R7, React 19 components are functions returning JSX; `StrictMode` in `main.tsx` is a React development aid, not a build step. Zustand `^5.0.0` (`web/frontend/package.json:20`) is the sole state library; no router/state server package is present.
- Charts/animation: `echarts ^5.5.0` (`web/frontend/package.json:17`), `animejs ^4.5.0` (`web/frontend/package.json:16`) for `animateViewEnter` (`App.tsx:69` via `lib/anime`).
- Toolchain: `vite ^6.0.0` + `@vitejs/plugin-react ^4.0.0` (`web/frontend/package.json:28,31`), `typescript ^5.7.0`, `vitest ^3.2.7`, `playwright ^1.62.1` (`web/frontend/package.json:22-32`). `bun.lock` present (`web/frontend/bun.lock` listing) and `Makefile:39` / `setup.go:91-96` invoke `bun` (`bun install`, `bun run build`, `bun x tsc/vitest`). Per R6, Vite Node requirement (20.19+/22.12+) applies to whoever runs `vite`; bun substitution is a local choice — `unverified — primary source not found` for bun↔node Vite parity beyond observed scripts.
- Embed coupling: `dist/.gitkeep` placeholder guarantees `go embed` succeeds pre-build — `web/frontend/embed.go:1-3`; `Makefile:6-7` and `setup.go:82-122` rebuild `dist` before `go build ./cmd/cgo`.

## 5. Primary-source requirements (what a standalone tool must satisfy)

### 5.1 CLI framework (Cobra R4 vs urfave R5 vs stdlib `flag`)

- Today: stdlib `flag.NewFlagSet(..., flag.ExitOnError)` per subcommand (`cmd/cgo/main.go:56,77,140,149,280`, `cmd/cgo/kit.go:16`, `cmd/cgo-chr/main.go:42-50`) + manual `switch os.Args[1]` + hand-written `usage()` (`cmd/cgo/main.go:359-373`, `cmd/cgo/kit.go:22-61`). Consequences observed: no autogenerated help, no suggestions (`did you mean`), no shell completion, no man pages, duplicated usage strings, inconsistent flag styles (long-only, `--flag=value` vs `--flag value`, `kit` `--yes` propagated via env `CGO_YES` — `cmd/cgo/kit.go:70-75`).
- Per R4, Cobra would require: one `cobra.Command` tree, flags via pflag (POSIX short+long), global/local/cascading flags, auto help/completion/man, suggestions. Cost: new `require` (Cobra + pflag) per R1; module path stays same, but build list grows; adopters (Kubernetes/Hugo/GitHub CLI per R4 README) gain completion for free. Cobra's "best applications read like sentences" pattern (`APPNAME VERB NOUN --ADJECTIVE`) matches existing `cgo kit <action>`, `cgo audit --link-type ...`, `cgo run --profiles ...` surface.
- Per R5, urfave/cli would require: declarative `Command` structs, alias support, dynamic completion (bash/zsh/fish/powershell), man/Markdown gen, **zero non-stdlib deps**, input lookup from env/files (via altsrc module). Cost: zero build-list growth, but v3 module path is `github.com/urfave/cli/v3` (per Go Reference badge in R5) — a `/v3` suffix is required by R1 for v2+ and would need pinning in `go.mod` + `go.sum`. Help templating can be disabled via `urfave_cli_no_template` build tag (R5 Build tags).
- Staying on stdlib `flag` is allowed (no R1 violation), but any standalone claiming POSIX-compliance or completion must adopt one of R4/R5. Decision is open; what is non-optional per R1 is that any new framework must appear in `require` with an immutable version, and any major bump must follow the module-path `/vN` rule.

### 5.2 Go modules (R1)

- Standalone tools need independent `go.mod` (or at minimum independent `require` sets) so `meteolink`/`cgo-chr` stop inheriting `fyne`, `bubbletea`, `lipgloss`. Per R1 §"Modules, packages, and versions", a module is the versioned unit; a package cannot be versioned independently inside one module. Today's single module is why `go list -m all` includes GL deps for the alias and CHR aux even though they never import them.
- Every standalone must declare `module <unique-controlled-path>`, `go <minimum>`, `require <path> <vX.Y.Z>`. Per R1 §"module" Notes, the path must uniquely identify the module and ideally be the repo location; do not reuse `github.com/Realms4239/cgo` for a split module without a `/v2` or new path — that would collide under R1's unique-ID rule.
- Tags are immutable once pushed; retraction needs a newer tag (R1 §"retract"). Releases today use GoReleaser `name_template: cgo-{{ .Os }}-{{ .Arch }}`, `checksums.txt` SHA-256 — `.goreleaser.yaml:35-51`. Any new module layout must keep tag = version semantics per R1/R2.

### 5.3 SemVer (R2)

- `1.3.5` (KitVersion, VERSION, frontend) places the product at MAJOR 1. Per R2 §5, "Version 1.0.0 defines the public API" — so 1.x signals a stable public API whose breaking changes must be MAJOR, backward-compatible features MINOR, bug fixes PATCH (§§6–8). Candidate public APIs: CLI flags/exit codes (table §3.1), `pkg/api` REST/SSE (paths, status codes 409/501/503, delta semantics `docs/architecture.md:23-41`), frontend `/api` + `go:embed` layout. None of these is formally declared as "the public API" in-repo — see §1.
- Practical requirement: before splitting, declare which surface is public (CLI + `/api/health|state|results|stream` + audit report JSON/MD) and freeze it under R2 §3 (released contents MUST NOT be modified; fix = new version). Minor additions (new `kit` action, new `/api` field additive) increment MINOR; breaking renames require MAJOR + `/v2` module path per R1.
- Build metadata / pre-release (R2 §§9–10) are not used in observed versions (`VERSION:1` is bare `1.3.5`); introducing `-0` or `+sha` would need a precedence policy per R2 §11.

### 5.4 OCI (R3)

- Today distribution is **not OCI**: GoReleaser tar.gz (zip on Windows) containing `cgo + meteolink + kit/*.example + kit/*.sh` — `.goreleaser.yaml:35-47` + `checksums.txt` — `.goreleaser.yaml:49-51`; deploy is `scp` of `cgo-linux` built with `CGO_ENABLED=0` — `internal/kit/kit.go:1358-1397` (`crossCompile env GOOS=linux GOARCH=amd64 CGO_ENABLED=0` at `internal/kit/kit.go:1361`). No image layout, manifest, or registry flow was observed.
- Per R3, an OCI-packaged server would be: download OCI Image → unpack to Runtime bundle → run by OCI Runtime, with config carrying command/args/env. Applied here, `cgo --serve` would become an Image whose config `Cmd/Args/Env` includes `CGO_IFACE`, `CGO_DASHBOARD__ADDR`, `CGO_BENCH`, `CGO_CHANNEL_MODEL` (envs already read at `cmd/cgo/server.go:34-48,173-179` and `cmd/cgo/main.go:57,292`). This would replace versioned tar + scp with content-addressable digests (the OCI Image Format provides Go types + JSON Schema + intra-blob validation per R3) and would be the natural home for `cgo-server` and `cgo-chr` (QEMU + CHR image as separate layer).
- What OCI would *not* give: running unprivileged shaping on the host (tc/netem need root/CAP_NET_ADMIN — `docs/architecture.md:85-89`, `cmd/cgo/main.go:99-102`, `pkg/doctor` mode). Image choice (host-network vs privileged vs dualns namespaces) is a deployment trade-off, not an R3 requirement — exact OCI Image/D Manifest JSON fields beyond R3's README summary are `unverified — primary source not found`.

### 5.5 Vite + React (R6/R7)

- Standalone frontend must satisfy R6: `index.html` entry at project root, `vite build` producing optimized static assets in `dist/` (today `outDir:'dist'` — `web/frontend/vite.config.ts:7`), React plugin via `@vitejs/plugin-react` (today `web/frontend/vite.config.ts:1-5`), Node 20.19+/22.12+ for the build host. ECharts tree-shaking is a Rolldown concern now (R6 lists Rolldown as bundler); today's "single self-contained index chunk" note (`check-bundle.mjs:52-58`) must be re-validated against R6's current defaults if Vite upgrades from `^6.0.0`.
- Per R7, frontend apps are component trees with lifted state + props. Today's topology (Zustand stores as lifted state — `store/ui.ts:38-56`, props drilling into 40+ components) is a valid R7 pattern; a split frontend must preserve the 4-destination IA (`store/ui.ts:7-12`) and the lazy/Suspense boundaries (`App.tsx:23-25,149-163`) unless the API contract changes. Styling via `className` + CSS files (`src/styles/*.css`) matches R7's React-prescribed approach.

## 6. Gap analysis — including the GUI Linux build failure

### 6.1 GUI Linux build failure (blocking `go build ./...` on Linux)

Evidence (primary, repo):

- `docs/VERIFY.md:27`: "`GOOS=linux GOARCH=amd64 go build ./...` → exit 1, sole failing package `github.com/Realms4239/cgo/cmd/cgo-gui` (`fyne.io/fyne/v2/app` → `go-gl/gl` v2.1 'build constraints exclude all Go files'; `CGO_ENABLED=0`, no `x86_64-linux-gnu-gcc`); `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` → exit 0. `cmd/cgo-gui/fygui_linux.go` last touched in `35bb021`, an ancestor of HEAD: pre-existing."
- `docs/VERIFY.md:32` (Slice 6 verifier re-run) and `docs/VERIFY.md:41` (Slice 10) repeat the same sole-failure with `pkgs=$(go list ./... | grep -v '/cmd/cgo-gui$')` → exit 0.
- `docs/CHECKPOINTS.md:256-260`: "`GOOS=linux GOARCH=amd64 go build ./...` → exit 1, only `cmd/cgo-gui` (`fyne.io/fyne/v2` → `github.com/go-gl/gl/v2.1/gl: build constraints exclude all Go files`); ... `CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/...` → exit 0."
- `docs/VERIFY.md:27` header interpretation: "`cmd/cgo-gui` excluded (needs cgo/go-gl)" — i.e. gates deliberately exclude it, not that it builds.

Root cause chain (repo → R1):

1. `cmd/cgo-gui/fygui_linux.go:1` is `//go:build linux` with no exclusion tag, unconditionally importing `fyne.io/fyne/v2` — `cmd/cgo-gui/fygui_linux.go:19-22`. On a Windows host cross-compiling to Linux, Go disables CGO when it cannot find a suitable cross toolchain (`CGO_ENABLED=0`, no `x86_64-linux-gnu-gcc` per VERIFY). 
2. Fyne's Linux GL backend (`go-gl/gl/v2.1`) provides zero Go files when CGO is disabled — hence "build constraints exclude all Go files". This is a Fyne/go-gl behaviour; Fyne's exact apt/pkg-config prerequisites were not fetched — `unverified — primary source not found`.
3. The Windows build never hits this because `cmd/cgo-gui/main.go:1` is `//go:build windows` and imports no Fyne — `cmd/cgo-gui/main.go:10-19`. The Linux file is a parallel `main()` (`cmd/cgo-gui/fygui_linux.go:731-747`) chosen solely by GOOS. There is no intermediate `//go:build linux && cgo` guard or stub `main` for the CGO-free Linux case.

Why it matters (not just cosmetic):

- `go test ./...` / `go vet ./...` / `go build ./...` are the project's own gates (`Makefile:9-15`, `verify.py` gates per `docs/VERIFY.md`). A red `./...` on Linux means cross-build cannot be `VERIFIED`; `docs/VERIFY.md:27` records it as VERIFIED only by excluding the package. This is the correct "recorded, not hidden" pattern, but the gate stays red.
- The failure is pre-existing (ancestor `35bb021`) — any new Linux CI job running `./...` fails from HEAD backwards, masking new regressions.
- `.goreleaser.yaml:8-33` never builds `cgo-gui` for any target, and `Makefile:6-7` builds only `cmd/cgo`. So the GUI is effectively undeployed via release automation on both Windows and Linux — the failure is invisible in shipped artifacts but visible in every `./...` invocation.

Minimal fix shape (no new dep, per R1): isolate the CGO-dependent Linux view behind `//go:build linux && cgo` + a `//go:build !linux || !cgo` stub `main()` that prints the honest no-display redirect (`fygui_linux.go:732-735` logic) and exits 2; or extract `cmd/cgo-gui` into its own `go.mod` excluded from the monolith `./...`. The stub needs no `require` (stdlib only), satisfying R1 §"require" (only imported paths need listing).

### 6.2 Other gaps blocking standalone tools

| Gap | Evidence | Why it blocks standalone |
|---|---|---|
| Monolith `go.mod` forces all binaries to resolve Fyne+Tea+LipGloss+x/sys | `go.mod:6-9` + R1 module concept | `meteolink` (stdlib only `cmd/meteolink/main.go:8-14`) and `cgo-chr` (stdlib+pkg/audit `cmd/cgo-chr/main.go:7-12`) cannot be built alone without the GUI/terminal stack. Violates least-privilege for audit. |
| Manual CLI (no Cobra/urfave) | `cmd/cgo/main.go:43-302` switch, `flag.NewFlagSet` per verb, `usage()` `cmd/cgo/main.go:359-373` | No POSIX short flags, no completion, no man, duplicated help. Per R4/R5 this is the entire feature set a framework would provide; staying manual means every standalone re-invents it. |
| Version triple (KitVersion + VERSION + package.json) hand-bumped | `internal/kit/kit.go:52`, `VERSION:1`, `web/frontend/package.json:4`, `cmd/cgo/main.go:36-41` | Drift risk already warned in code. Per R2, release identity is the version; three writers can diverge without a generator. |
| `cgo-gui` duplication (Win32 vs Fyne) with only `core.go` shared | `cmd/cgo-gui/core.go:1-4`, `actions.go:381-518` vs `fygui_linux.go:420-562`, `window.go:177-237` vs `fygui_linux.go:648-729` | Behaviour drift (button IDs `core.go:75-103`, Suite chain logic) must be kept in sync by hand. Tested `buttonAction`/`isKnownSuiteVerb` pure functions mitigate, but two event loops (Win32 pump `window.go:63-73` vs Fyne `pump()` `fygui_linux.go:78-117`) remain. |
| `meteolink` sibling-binary dependency | `cmd/meteolink/main.go:24-38`, archive layout `.goreleaser.yaml:36-44` | Standalone `meteolink` without archive breaks. Per R1, a module's binaries should be buildable from their own `main` without a sibling artifact. |
| Frontend bundle watchdog weakened (in-chunk echarts) | `web/frontend/scripts/check-bundle.mjs:52-58` | TOTAL budget (600 KB gz) still catches blowups — `check-bundle.mjs:65-69` — but the echarts-specific regression (+137 KB gz per comment `check-bundle.mjs:10-12`) is no longer isolated. Standalone frontend needs the chunk boundary restored or a different invariant. |
| No OCI; deploy via scp of prebuilt `cgo-linux` | `.goreleaser.yaml:35-47`, `internal/kit/kit.go:1358-1397`, R3 flow | No content-addressable, reproducible, immutable distribution. Bench bring-up (`kit bootstrap`, `testbed up`) is script + scp, not `download → unpack → run` per R3. |
| API not declared stable | No stability doc found (§1); SSE semantics only in prose `docs/architecture.md:23-41` | Per R2, MAJOR/MINOR cannot be assigned without a declared API. Split modules would version together blindly. |
| `docs/architecture.md` CLI coverage is partial | `docs/architecture.md` has Flux SSE, couche tc, Portes G0–G7, Gel/manifest, Modes — `docs/architecture.md:8-103` — but no `kit`/`audit`/`run`/`chr`/`gui` flag tables | Each binary's surface must be reverse-engineered from `main.go` (this file) — onboarding cost for standalone extraction. |
| Windows/Linux mode split is implicit in `doctor.Mode("auto")` | `cmd/cgo/main.go:64,75,126`, `docs/architecture.md:85-93` (Linux=full, Windows=observe, 501 for shape/run/watch) | Standalone server on Windows silently degrades; a container image must encode the mode in its config/env per R3, not rely on OS detection. |

## 7. Target tool decomposition (minimal, R1/R2/R3-compliant)

Principles (from R1/R2/R3): **one module per versioned unit**, **no binary imports what it does not execute**, **public API declared before 1.x split**, **distribution is OCI for server-side, tar+checksums only for operator workstations**. Numbers in brackets refer to slices §8.

- **`cgo-server`** (new module, e.g. `github.com/Realms4239/cgo-server` or `/v2` on the monorepo path — `unverified — primary source not found` for which path the org prefers; R1 requires it be unique-controlled). Contents: `cmd/cgo --serve` only (`server.go` + `pkg/api` + `pkg/campagne` + `pkg/qdisc`+`pkg/probe`+`pkg/results`+`pkg/figures`+`pkg/doctor`+`pkg/iface`+ frontend `embed.go`). Deps: stdlib only (per `docs/architecture.md:10-12` + `docs/architecture.md:95-103` subprocess-tc). Artefact: static binary + OCI Image (entrypoint `--serve`, env `CGO_IFACE/CGO_BENCH/CGO_CHANNEL_MODEL/CGO_DASHBOARD__ADDR`, `ReadHeaderTimeout 5s` already in `server.go:342`). This is the bench/edge tool.
- **`cgo-kit`** (new module). Contents: `cmd/cgo kit` + `setup` thin wrappers + `internal/kit` + `internal/vm` (becomes `pkg/kit|vm` or stays `internal` per consumer — R1 "internal" is a Go convention, not an R1 requirement). Deps: stdlib + `x/sys/windows` on Windows only. Replaces the 21-action table (§3.1) as a single deployment tool; `setup`'s 7 steps (`setup.go:42-229`) move here verbatim. Exit codes 2–8 preserved.
- **`cgo-run` / **`cgo-audit`** (either one module `cgo-measure` or two). Contents: `run` (`runcli.go`), `audit` (`pkg/audit`+`pkg/remediate`), `verify`/`figures`/`testbedsrv`/`calibrate` as subcommands. Deps: stdlib only today (run/audit import only `pkg/campagne|model|results|audit|remediate`). No TUI/HTTP needed — separates measurement from serving. Standalone use: SSH campaign without opening dashboard (already the use-case in `runcli.go:19-22` header).
- **`cgo-tui`** (new module). Contents: `tui/top` (`tui.go`+`tui_audit.go`) + `kit tui` (`kittui*.go`). Deps: `bubbletea`+`lipgloss` (only place that needs them). Talks to `cgo-server` via `/api/state|results` (`tui.go:127-159`) + `exec os.Args[0] kit <action>` (`tui.go:161-167`) — second half must become a direct `cgo-kit` exec or RPC in the split.
- **`cgo-chr`** (new module, trivial split). Contents: `cmd/cgo-chr` + `pkg/audit` CHR-proof subset (`ParseRouterOSConfig/CompareRouterOS/WriteCHRProof`). Deps: stdlib only. Ships as separate binary + OCI sidecar (QEMU + CHR image layers).
- **`cgo-gui`** (new module, CGO-split). Contents: `core.go`+`actions.go`(windows)+`fygui_linux.go` with the `//go:build linux && cgo` guard from §6.1, `pkg/kit` client. This isolates Fyne+`x/sys` to exactly this module; the monolith `./...` goes green on `CGO_ENABLED=0`. Keeps dual-view parity (shared `buttonAction`/`isKnownSuiteVerb`).
- **`meteolink`** (deleted as binary). Replaced by (a) symlink/hardlink to `cgo` (argv[0] check) or (b) a Cobra alias command — either removes the sibling-presence failure at `cmd/meteolink/main.go:37-38`. If a wrapper is kept for npm compat, it becomes a 5-line shell script, not a Go module.
- **`cgo-frontend`** (kept as Vite app, versioned API contract). Single `package.json` version generated from KitVersion (§8 S5); `vite.config.ts` proxy stays dev-only; production `dist/` is embedded by `cgo-server` (`embed.go:11`) and also published as a versioned tar/OCI layer for CDN/static hosting. Bundle budget 600 KB gz retained until echarts chunk isolation is restored (§8 S6).

Module-path sketch (paths themselves are placeholders — R1 requires choosing unique controlled paths before first publish; `/v2` suffix mandatory if reusing `github.com/Realms4239/cgo` for v2):

```text
github.com/Realms4239/cgo-server   (pkg/api,campagne,qdisc,probe,results,figures,doctor,iface + embed)
github.com/Realms4239/cgo-kit      (kit,vm,doctor)
github.com/Realms4239/cgo-measure  (run,audit,verify,figures)
github.com/Realms4239/cgo-tui      (bubbletea,lipgloss)
github.com/Realms4239/cgo-chr      (audit-CHR subset)
github.com/Realms4239/cgo-gui      (fyne,x/sys,kit client)
github.com/Realms4239/cgo-model    (model,metrics,results shared — extracted first, zero-dep)
```

Shared libraries (`model`, `metrics`, `results`, `figures`) extract **before** binaries, because every binary imports them (server `cmd/cgo/server.go:15-22`, run `runcli.go:14-16`, chr `run.go:11`).

## 8. Slices (ordered, one scope each)

Each slice is shippable alone; later slices assume earlier ones. Acceptance = command + exit, not prose.

- **S1 — `cmd/cgo-gui` isolated, `./...` green on Linux (fixes §6.1).** Scope: `cmd/cgo-gui/fygui_linux.go` build tags + stub `main` (or separate `go.mod`). Acceptance: `GOOS=linux GOARCH=amd64 go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./cmd/cgo-gui` → exit 0 (Windows untouched); `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go vet ./...` → exit 0. Risk: stub text must match today's no-display message (`fygui_linux.go:732-735`); Fyne CGO dev deps remain `unverified — primary source not found` for Linux CI image.
- **S2 — Extract `cgo-chr` to its own module.** Scope: new `go.mod`, `cmd/cgo-chr` + `pkg/audit` CHR subset only. Acceptance: `go list -m all` in `cgo-chr` shows no `fyne/bubbletea/lipgloss/x-sys`; `go build ./cmd/cgo-chr` binary passes `--dry-run` and `--export-file` replay tests (`cmd/cgo-chr/main_test.go`, `serial_test.go`, `run_test.go`). Proves R1 split mechanism.
- **S3 — Extract `pkg/model|metrics|results|figures` as shared libs, then `cgo-server`.** Scope: `docs/architecture.md` already lists single-responsibility packages `docs/architecture.md:10-21`; move with stable imports first. Acceptance: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./pkg/... ./cmd/cgo/... ./cmd/meteolink/...` → exit 0 (today's scoped green at `VERIFY.md:27`); `cgo --serve` binds `https://meteolink.dev:9090` (`main.go:59`) and serves `dist/index.html` via `embed.go:14-21`; SSE `/api/stream` 10 Hz + 64-sub 503 (`architecture.md:23-41`).
- **S4 — CLI framework decision + migration for `cgo`/`kit`/`chr`.** Scope: adopt Cobra (R4) **or** urfave (R5) for one binary (`cgo-kit`) hoisting `runKit` (`kit.go:15-215`). Acceptance: `cgo kit --help` generated, `cgo kit deploy --help` shows `--config/--deep/--shallow/--public/--yes` from `kit.go:16-21`, unknown action suggests nearest (`did you mean`) per R4; exit codes 2–8 preserved (`kit.go:58-59`). Keep `flag` in `cgo-chr` if urfave's zero-dep is chosen (R5) — decision recorded before coding. Unverified: migration effort for `setup.go:324-336` mini-flagset under either framework — `unverified — primary source not found`.
- **S5 — Version single-source + release all bins.** Scope: generate `VERSION`, `package.json:version`, `KitVersion` from one tag. Acceptance: `go build -ldflags -X main.version=v1.3.5` matches `internal/kit/kit.go:52` and `web/frontend/package.json:4`; `cgo version` (`main.go:75`) == `version` == `VERSION`. Per R2, bump rules documented (MINOR additive CLI, MAJOR breaking).
- **S6 — Frontend standalone hardening.** Scope: re-isolate echarts chunk or replace watchdog, version `/api` contract, CI `bun run build` (today `package.json:8`). Acceptance: `bun run build` → `check-bundle.mjs` checks echarts ≤350 KB gz separately again (`check-bundle.mjs:35-36`) or proves in-chunk invariant; `bunx vitest run` + `playwright test` green (`package.json:10-12`); `App.tsx` 4 modes (`ui.ts:7-12`) render without console errors (Playwright). Per R6, pin Vite `^6.0.0` or upgrade deliberately with Node version recorded.
- **S7 — OCI server image.** Scope: `Dockerfile`/`oci-layout` for `cgo-server` only. Acceptance: `docker run` (or `podman`) image starts `cgo --serve --addr 0.0.0.0:9090 --tls=false --http-addr=` (same args as systemd unit `main.go:109`) and `GET /api/health` returns `ok` (per `architecture.md:92-93`); image digest == release tag. Per R3, image carries command/env; privileged shapes (tc/netem) declared in deployment manifest, not hidden. Exact OCI manifest/config fields beyond R3 README are `unverified — primary source not found`.

## 9. Anti-toy criteria (a standalone must meet all)

A tool is **toy** if any row fails. These are testable, not aspirational.

| # | Criterion | Test | Source |
|---|---|---|---|
| 1 | `./...` builds cross-platform | `GOOS=linux GOARCH=amd64 go build ./...` → 0; `GOOS=windows GOARCH=amd64 go build ./...` → 0; `go vet ./...` → 0 | `Makefile:9-15`, `VERIFY.md:27` (today red on Linux — §6.1) |
| 2 | Declared, frozen public API | CLI/REST/SSE surface file:line-able (§3), version bumps obey R2 §§6–8; released contents never mutated (R2 §3) | R2 §1 "MUST declare a public API" |
| 3 | Honest measurement, never fabricated | No synthetic telemetry; absent data is named (`bufferbloat NON MESURÉ` — `main.go:215-217`; `non mesuré` in audit summary `main.go:312-357`; TUI "sans API affiche déconnecté" — `tui.go:395-397`) | `cmd/cgo/main.go:212-278`, `cmd/cgo/tui.go:395-397` |
| 4 | Gated validity (G0–G7) + quarantine/skip journalled | Cell marked valid/degraded/invalid by gates `architecture.md:60-74`; quarantine/skip recorded `server.go:90-98` | `docs/architecture.md:60-74`, `cmd/cgo/server.go:90-98` |
| 5 | Sealed provenance, verifiable offline | `manifest.json` SHA-256 per file + config_sha256 recomputed (`main.go:375-431`); `ProvenanceSHA/Hash8` in SVGs/reports (`architecture.md:76-83`); `verify` fails on uncovered file | `cmd/cgo/main.go:128-138,375-431` |
| 6 | Real resource lifecycle | tc `Reset→ApplyNetem→ApplyShaper` (`architecture.md:43-58`); manual shape 409 during campaign; shutdown drains `CloseHub` (10 Hz ticker) `server.go:386`; `run` freeze-on-Ctrl-C + run lock (`runcli.go:45-65`) | `docs/architecture.md:43-58`, `cmd/cgo/server.go:62-127,377-391` |
| 7 | CLI usable by a newcomer | `--help` per verb, POSIX flags or documented alternative, completion, man (R4/R5); exit codes documented and tested (2–8 for kit, 2 partial for audit, 4 for bench, 1 mail for chr) | R4/R5; `cmd/cgo/kit.go:58-59`, `cmd/cgo/main.go:266-278`, `cmd/cgo/runcli.go:183-218` |
| 8 | GUI never lies | No-display refuses (`fygui_linux.go:732-735`); stale banner marked périmé instantly (`actions.go:201-207`, `fygui_linux.go:213-218`); key setup verified, not assumed (`actions.go:255-264`, `fygui_linux.go:266-277`); suite chain anti-loop (`actions.go:549-567`) | `cmd/cgo-gui/fygui_linux.go:731-747`, `cmd/cgo-gui/actions.go:188-270` |
| 9 | Frontend production-grade | `tsc --noEmit` + `vite build` + bundle budget (`package.json:8`, `check-bundle.mjs:65-69`); `vitest` + `playwright` green (`package.json:10-12`); ErrorBoundary + lazy/Suspense per view (`App.tsx:124-164`); keyboard/screen-reader path preserved (`App.tsx:106-122,127`) | R6/R7; `web/frontend/package.json:6-13`, `web/frontend/src/App.tsx` |
| 10 | Immutable, reproducible distribution | Version = tag; `checksums.txt` SHA-256 (`.goreleaser.yaml:49-51`) or OCI digest (R3); not a mutable file copy. `VERSION` == binary `version` == frontend version (S5) | R1/R2/R3; `.goreleaser.yaml:35-51`, `internal/kit/kit.go:49-52` |
| 11 | No hidden topology guesses | `CGO_IFACE` explicit, no auto-detect ("une mauvaise interface serait une mesure mal étiquetée" — `server.go:32-48`); bench mode detected, not chosen (`main.go:161-163,190-192`); NAT forward only when NIC is NAT (`kit.go:1008-1031`) | `cmd/cgo/server.go:32-48`, `cmd/cgo/main.go:175-193`, `internal/kit/kit.go:1008-1031` |
| 12 | At most one active problem per slice, failures recorded not waived | `make verify` exit 1 = named gate failed, exit 2 = gate could not run (not green) — per AGENTS §The one command. BLOCKED rows name owner + unblock condition (e.g. BBRv3 `VERIFY.md:52`) | `docs/VERIFY.md` convention; `Makefile:24-31` |

## Appendix — what was *not* fetched (unverified)

- Fyne Linux prerequisites (apt packages, GCC/X11/OpenGL/pkg-config) — `unverified — primary source not found`.
- OCI Image/D Manifest exact JSON keys (`config`, `annotations`, `mediaTypes`) beyond R3 README — `unverified — primary source not found`.
- Cobra/pflag version to pin, and urfave `/v3` migration diff for `setup.go` — `unverified — primary source not found`.
- Bun vs npm parity for Vite 6 builds beyond observed `bun.lock` usage — `unverified — primary source not found`.

## Files consulted (selection)

`cmd/cgo/main.go`, `cmd/cgo/kit.go`, `cmd/cgo/server.go`, `cmd/cgo/runcli.go`, `cmd/cgo/setup.go`, `cmd/cgo/tui.go`, `cmd/cgo/calibrate.go`, `cmd/meteolink/main.go`, `cmd/cgo-chr/main.go`, `cmd/cgo-chr/run.go`, `cmd/cgo-gui/main.go`, `cmd/cgo-gui/win32.go`, `cmd/cgo-gui/window.go`, `cmd/cgo-gui/core.go`, `cmd/cgo-gui/actions.go`, `cmd/cgo-gui/fygui_linux.go`, `internal/kit/kit.go`, `web/frontend/package.json`, `web/frontend/vite.config.ts`, `web/frontend/embed.go`, `web/frontend/scripts/check-bundle.mjs`, `web/frontend/src/App.tsx`, `web/frontend/src/main.tsx`, `web/frontend/src/store/ui.ts`, `go.mod`, `.goreleaser.yaml`, `Makefile`, `VERSION`, `docs/architecture.md`, `docs/VERIFY.md:27,32,41`, `docs/CHECKPOINTS.md:256-260`.
