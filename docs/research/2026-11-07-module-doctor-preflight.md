# pkg/doctor + cgo doctor CLI as standalone tool — preflight research

*Date: 2026-11-07 · Scope: `pkg/doctor/*.go`, `cmd/cgo` `doctor` wiring, `docs/api.md` doctor endpoint, `docs/deploy.md` doctor section only. Working dir `operator-cockpit`. Research only — no code changed.*

## 0. Method & sources (primary)

| # | Source (primary) | URL / path | Status | Used for |
|---|---|---|---|---|
| R1 | `capabilities(7)` — Linux man page (man7.org, man-pages 6.19) | https://man7.org/linux/man-pages/man7/capabilities.7.html | fetched 2026-11-07 | `CAP_NET_ADMIN` = interface config / routing / firewall / TOS / promiscuous / multicast-capable `setsockopt` set; privileged (euid 0) bypasses checks only until bounding set / UID change clears caps; thread sets Permitted/Inheritable/Effective/Bounding; file caps (`setcap(8)` semantics in-file) + `capget`/`capset` manipulation |
| R2 | `capget(2)` — Linux man page | https://man7.org/linux/man-pages/man2/capget.2.html | fetched 2026-11-07 | raw `capget`/`capset` = kernel interface for getting/setting thread caps; `effective/permitted/inheritable` bit masks (§Description); portable `cap_get_proc(3)` preferred; probing any pid via `hdrp->pid` |
| R3 | `netlink(7)` — Linux man page | https://man7.org/linux/man-pages/man7/netlink.7.html | fetched 2026-11-07 | `AF_NETLINK` kernel↔userspace datagram service; `NETLINK_ROUTE` = routing/link/qdisc/classifier control (see `rtnetlink(7)`); `NLM_F_ATOMIC` requires `CAP_NET_ADMIN` or euid 0; unreliable — may drop on OOM, resync on `ENOBUFS` |
| R4 | `ip(8)` — Linux man page (iproute2) | https://man7.org/linux/man-pages/man8/ip.8.html | fetched 2026-11-07 | `netns` is a first-class `OBJECT` (show/manipulate routing/devices/interfaces/tunnels); `-n <NETNS>` = shorthand for `ip netns exec NETNS ip …`; exit 0 ok / 1 syntax / 2 kernel error (§EXIT STATUS) |
| R5 | `ethtool(8)` — Linux man page | https://man7.org/linux/man-pages/man8/ethtool.8.html | fetched 2026-11-07 | `-k/--show-features/--show-offload` = query offload state; `-K/--features/--offload` built-ins `tso` (TCP seg), `gso` (generic seg), `gro` (generic recv), `lro` etc. — query vs set are distinct verbs |
| R6 | Linux IP Sysctl (`docs.kernel.org`) | https://docs.kernel.org/networking/ip-sysctl.html | fetched 2026-11-07 | `tcp_available_congestion_control` = STRING "Shows the available congestion control choices that are registered. More … may be available as modules, but not loaded."; `tcp_allowed_congestion_control` = subset; `tcp_congestion_control` sets for new connections |
| R7 | Go Modules Reference | https://go.dev/ref/mod | fetched 2026-11-07 | module = packages released/versioned/distributed together; module path = canonical ID + import prefix; `go 1.21+` directive is a mandatory minimum; `/vN` suffix for v2+; MVS resolution |
| R8 | Semantic Versioning 2.0.0 | https://semver.org/spec/v2.0.0.html | fetched 2026-11-07 | §1 MUST declare precise public API; §3 immutability once released; §4 `0.y.z` MAY change; §6 PATCH = compat bug-fix; §7 MINOR = compat functionality; §8 MAJOR = incompat API change |
| — | `capsh(8)` manual page | https://man7.org/linux/man-pages/man8/capsh.8.html | **unverified — primary source not found** (404 fetch 2026-11-07) | any `capsh --print/--decode/--caps` tooling behaviour — not cited; only R1+R2 capability model is cited |
| — | `tc(8)` netem/cake behaviour, `ping(8)` round-trip semantics, BBR/BBRv3 algorithm internals | — | **unverified — primary source not found** (not fetched; `tc(8)` fetch transport-error 2026-11-07) | shaping enforcement, probe usability, CC algorithm properties beyond R6 STRING listing |
| — | Repo | `pkg/doctor/doctor.go` (183 lines), `cc_test.go`, `offload_test.go`, `pkg/model/cc.go`, `cmd/cgo/main.go`, `cmd/cgo/server.go`, `cmd/cgo/setup.go`, `cmd/cgo/kit.go`, `internal/kit/kit.go`, `pkg/api/server.go`, `docs/api.md`, `docs/deploy.md`, `go.mod` | read 2026-11-07 | all `file:line` claims below |

Method: every technical claim below cites `file:line` or `(R#)`; anything else is labelled **unverified — primary source not found**. Prior combined study is `docs/research/2026-11-07-module-api-doctor-iface.md` (R1–R8 there); this file is the doctor-only preflight slice of that study, re-sourced against R1–R8 above.

## 1. Responsibility (as-shipped)

Single responsibility: report the same capability truth the dashboard badge shows, read-only, without mutating the host. `pkg/doctor/doctor.go:1-3` states "capability report (grilling Q11): one line per capability, la même vérité que le badge du tableau de bord. Rien ici ne mute l'hôte: lecture et rapport." (repo comment, `doctor.go:1-3`).

Concretely as-shipped (`pkg/doctor/doctor.go:24-74`):

- Resolve mode: `Mode(override)` (`doctor.go:24-34`) — explicit `observe|full` wins; `""`/unknown → auto → Windows→`observe`, else `full` (`doctor.go:28-33`). The Windows/Linux split is the same matrix `docs/deploy.md:7-16` declares (Windows observation vs Linux full) and `docs/api.md:8-11` repeats for HTTP 501 gating.
- Emit checks (`Report`, `doctor.go:37-65`): `os` (`runtime.GOOS/GOARCH → mode`, `doctor.go:39`), `tc` ( `exec.LookPath("tc")`, `doctor.go:41-49`), `cap_net_admin` (Windows warn / `os.Geteuid()==0` ok / else warn, `doctor.go:50-56`), `bbr`+`bbr3` via `ccChecks` (`doctor.go:57`, `152-183`), `bench-offloads` via `benchOffloadCheck` (`doctor.go:58`, `123-147`), `ping` (`doctor.go:59-63`).
- CLI form `Print()` (`doctor.go:68-74`) prints `mode: …` then `[status] name — detail` per check.
- Parse helper `ParseOffloadFeatures` (`doctor.go:81-106`) reads `ethtool -k` output tolerantly: `feature: on|off`, first state word wins, `[fixed]`/`[requested on]` suffixes ignored (`doctor.go:94-97`).

What it deliberately does not do: no `tc` application, no `modprobe`, no offload mutation, no benchmark run (`doctor.go:1-3` header + `benchOffloadCheck` comment `doctor.go:119-122` "Lecture seule : jamais de mutation de l'hôte"). Its only internal dep is `model.ParseCCList` (`doctor.go:13`, `pkg/model/cc.go:8-18`); otherwise stdlib (`doctor.go:6-14`).

Distinct neighbours (do not conflate as standalone):

- `cgo doctor` (this scope) vs `cgo kit doctor` (different verb): `cmd/cgo/main.go:72-73` dispatches bare `doctor` to `doctor.Print()`; `cmd/cgo/kit.go:26` documents `kit doctor` as "dépendances locales + config (informatif…)" and `cmd/cgo/kit.go:95-96` dispatches it to `Runner.Doctor` (`internal/kit/kit.go:822-907`) which probes `go/bun/node/ssh/scp/curl` paths, SSH key, port 22 / NAT-forward, hypervisors, VM path — none of which is in `pkg/doctor`. A standalone preflight ships the former, not the latter.
- HTTP mirror `GET /api/doctor` (`docs/api.md:18`, handler `pkg/api/server.go:405-411`, closure `cmd/cgo/server.go:304-307`) renders the same `Check` list as JSON (`server.go:305-306`). `docs/deploy.md:30-31` states both render the same `n/a` rule.

## 2. Exported API (file:line)

Complete non-test inventory for `pkg/doctor`. Unexported identifiers listed separately; they are not API.

| Export | Kind | Location | Notes |
|---|---|---|---|
| `Check` | struct | `doctor.go:16-20` | `{Name, Status (ok\|warn\|fail), Detail string}` with `json:"name"/"status"/"detail"` tags — the JSON shape served at `GET /api/doctor` (`pkg/api/server.go:405-411`, `cmd/cgo/server.go:304-307`, `docs/api.md:18`). |
| `Mode(override string) string` | func | `doctor.go:24-34` | `observe|full` passthrough; `""`/unknown → auto (Windows `observe`, else `full`). Callers: `cmd/cgo/main.go:68` (`--serve --mode`), `cmd/cgo/main.go:75` (`version` line), `cmd/cgo/setup.go:126` (wizard context). |
| `Report(mode string) (string, []Check)` | func | `doctor.go:37-65` | 7+ checks: `os`, `tc`, `cap_net_admin`, 2× `ccChecks`, `bench-offloads`, `ping`. No context/timeout (gap §5). Server caller: `cmd/cgo/server.go:304-306`. |
| `Print()` | func | `doctor.go:68-74` | CLI text; hardcodes `Mode("auto")` (`doctor.go:69`). Callers: `cmd/cgo/main.go:72-73` (`cgo doctor`), `cmd/cgo/setup.go:217` (step 7). |
| `ParseOffloadFeatures(out string) map[string]bool` | func | `doctor.go:81-106` | Pure parser: state = first word after `:` (`on`→true, `off`→false); `[fixed]`/`[requested on]` suffixes ignored; non-`on/off` states ignored; tested `offload_test.go:17-44`. |

Unexported (not API): `benchOffloadPairs` (6 pairs, `doctor.go:110-117`), `benchOffloadCheck` (`doctor.go:123-147`), `ccChecks` (`doctor.go:152-158`), `ccChecksFrom` (`doctor.go:160-183`).

Tests that pin semantics (evidence, not API):

- `pkg/doctor/cc_test.go:5-26` — exact-token `bbr` vs `bbr3` separation; unreadable file → 2×`warn`.
- `pkg/doctor/offload_test.go:17-44` — `on/off/[fixed]/empty/header` parse cases; `offload_test.go:46-66` requires `bench-offloads` check present in every mode.

## 3. Dependencies

### 3.1 Direct imports (evidence)

- `pkg/doctor/doctor.go:6-14`: stdlib (`fmt`, `os`, `os/exec`, `runtime`, `strings`) + `github.com/Realms4239/cgo/pkg/model` (only `model.ParseCCList`, `pkg/model/cc.go:8-18`: order-preserving dedupe tokeniser over `strings.Fields`).
- `go.mod:1-10`: `module github.com/Realms4239/cgo`, `go 1.25`, direct `require` = `fyne.io/fyne/v2 v2.8.1`, `charmbracelet/bubbletea v1.3.10`, `charmbracelet/lipgloss v1.1.0`, `golang.org/x/sys v0.36.0`. None of these is imported by `pkg/doctor` (GUI/TUI only); `pkg/doctor`'s build list is stdlib + `model` (R7 §`go.mod` file: `go.mod` records module identity + minima; R7 §Module paths: path is ID).
- Wiring deps (composition only, not library import): `cmd/cgo/main.go:31`, `cmd/cgo/server.go:17`, `cmd/cgo/setup.go:14` import `pkg/doctor` together with `api/campagne/iface/model/probe/qdisc/results` (`cmd/cgo/server.go:15-23`) — standalone must break this co-import (R7 §Modules: packages distributed together in one module cannot be versioned independently today).

### 3.2 Host / toolchain dependencies

| Dependency | Usage | Severity binding |
|---|---|---|
| `tc` binary in PATH | `exec.LookPath("tc")` gates `ok` vs `warn`/`fail` by mode (`doctor.go:41-49`) | **unverified — primary source not found** for `tc(8)` shaping behaviour beyond presence; R1 lists `CAP_NET_ADMIN` operations (interface config, routing, firewall) but not `tc` subcommand semantics |
| `ping` binary in PATH | `exec.LookPath("ping")` → `ok`/`fail` (`doctor.go:59-63`) | **unverified — primary source not found** for audit usability beyond presence (no round-trip, no SOCKS/proxy awareness) |
| `ip netns list` + `sudo -n ip netns exec <ns> ethtool -k <iface>` | bench gate (`doctor.go:127`, `132-133`) | R4 (`ip(8)`: `netns` is a first-class object; `-n` shorthand) + R3 (`NLM_F_ATOMIC` needs `CAP_NET_ADMIN`/euid 0) govern the substrate; `sudo -n` non-interactive behaviour **unverified — primary source not found** (not fetched) |
| `ethtool` (implicit) | `-k` read per bench iface (`doctor.go:132-139`) | R5: `-k` = query offload state; `-K` built-ins `tso/gso/gro` — short/long name equivalence (§4) |
| `/proc/sys/net/ipv4/tcp_available_congestion_control` | read + tokenise (`doctor.go:153`, `pkg/model/cc.go:8`) | R6: STRING = registered choices; unloaded-module algorithms absent by definition |
| euid / file caps | `os.Geteuid()==0` → `ok` (`doctor.go:52`); warn text suggests `sudo setcap cap_net_admin+ep ./cgo` (`doctor.go:55`, `docs/deploy.md:79-87`) | R1: euid 0 bypass is real but bounding-set narrowing and UID-change clearing can still deny; file caps via `setcap` are the mechanism (`docs/deploy.md:82-85`); R2: `capget` is the authoritative probe, not euid alone (gap §5) |

Hardcoded filesystem/exec surface (standalone blockers): `/proc/sys/net/ipv4/tcp_available_congestion_control` (`doctor.go:153`), `LookPath("tc"/"ping")` (`doctor.go:41`, `59`), `ip netns list` (`doctor.go:127`), `sudo -n ip netns exec … ethtool -k` (`doctor.go:132-133`), 6 bench pairs (`doctor.go:110-117` citing `kit/cgo-bench-dualns.sh` at `doctor.go:108-109`).

## 4. Primary-source requirements

Minimum external authorities any standalone `cgo-doctor` release must re-verify (with the exact claim each governs):

| ID | Authority | Claim it governs | Repo binding |
|---|---|---|---|
| R1 | `capabilities(7)` | `CAP_NET_ADMIN` operation set (interface config / routing / firewall / TOS / promiscuous / multicast / `SO_DEBUG/MARK/PRIORITY/RCVBUFFORCE/SNDBUFFORCE`); privileged bypass vs bounding-set narrowing; per-thread Permitted/Inheritable/Effective/Bounding; file caps stored in `security.capability` xattr, written with `CAP_SETFCAP`; `capget`/`capset` manipulation | `doctor.go:50-56` (`cap_net_admin` check + `setcap` suggestion), `docs/deploy.md:77-87` (privilege model) |
| R2 | `capget(2)` | `capget`/`capset` = raw kernel interface; `effective/permitted/inheritable` bit masks; `cap_get_proc(3)` portable alternative; probing any pid via `hdrp->pid` | gap §5 (euid heuristic must become `capget`-based); `internal/kit/kit.go` **not** in scope |
| R3 | `netlink(7)` | `NETLINK_ROUTE` = routing/link/qdisc/classifier control; `NLM_F_ATOMIC` needs `CAP_NET_ADMIN`/euid 0; unreliability + `ENOBUFS` resync | `tc` absence/`ip netns` substrate; `benchOffloadCheck` topology assumption (`doctor.go:110-117`) |
| R4 | `ip(8)` | `netns` object + `-n <NETNS>` shorthand; exit-status vocabulary (0/1/2) | `doctor.go:127`, `132` exec forms |
| R5 | `ethtool(8)` | `-k` = query; `-K` built-ins `tso/gso/gro/lro/rx/tx/sg/ufo/rxvlan/txvlan/ntuple/rxhash` | `doctor.go:132-139` (long output names `tcp-segmentation-offload` etc. vs R5 short `-K` names — equivalence **unverified — primary source not found**) |
| R6 | Linux IP Sysctl | `tcp_available_congestion_control` STRING semantics; subset rule for `tcp_allowed_congestion_control`; unloaded-module absence | `doctor.go:152-183`, `pkg/model/cc.go:8-18`, `cc_test.go:5-26`, `docs/api.md:71-73` (bbr3 exact-token rule) |
| R7 | Go Modules Reference | What counts as a releasable module; module-path lexical rules; `/vN` suffix; `go` directive = mandatory minimum (≥1.21); `require`/`replace`/`exclude`/`retract` semantics; MVS | `go.mod:1-3`; split plan §6 |
| R8 | SemVer 2.0.0 | §1 MUST declare precise public API; §3 immutability; §6–§8 PATCH/MINOR/MAJOR increments; §4 `0.y.z` instability | version policy §6 |

Fetch-or-it's-unverified rule: any claim about `tc(8)` netem/cake enforcement, `ip netns` bench topology freshness, `ping(8)` usability, `capsh(8)` output format, BBR/BBRv3 algorithm internals — label **unverified — primary source not found** until the authoritative man page / spec is fetched and quoted. Today's `docs/deploy.md` asserts some of these without external fetch (§5 marks each).

## 5. Gap analysis (what blocks standalone release)

Ordered by severity. Each gap cites the code and the primary source it would violate or the policy it breaks.

### 5.1 `cap_net_admin` is an euid heuristic, not a capability check (BLOCKER)

Observed `doctor.go:52`: `ok` iff `os.Geteuid()==0`. Per R1, euid 0 historically bypasses permission checks, but post-2.2 Linux divides superuser power into capabilities that can be independently enabled/disabled, and the bounding set can limit what `execve` grants; UID-change rules clear permitted/effective/ambient sets when all UIDs become nonzero. Per R2, `capget` returning the effective/permitted bit masks is the kernel's own query interface (`cap_get_proc(3)` portable). Consequence: a root process with a narrowed bounding set, or a `setcap cap_net_admin+ep` binary run as non-root, is misclassified in both directions by euid alone (false `ok` for narrowed-root, false `warn` for capped non-root). `docs/deploy.md:82-85` documents the correct `setcap` path, but the check does not verify it. Fix direction: `capget`-based `CAP_NET_ADMIN ∈ effective` (R2 structure `effective/permitted/inheritable` masks) with the euid shortcut kept only as a display detail. `capsh(8)` verification is **unverified — primary source not found** (R-table) so the tool must not shell out to `capsh` as its sole source.

### 5.2 Bare `exec` with no context/timeout — preflight can hang (BLOCKER)

Observed `doctor.go:41` (`LookPath`), `59` (`LookPath`), `127` (`Command("ip"…).Output()`), `132-133` (`sudo -n ip netns exec … ethtool -k …`). No `context.Context`, no deadline on any site. `GET /api/doctor` (`pkg/api/server.go:405-411` → `cmd/cgo/server.go:304-307`) inherits the hang: one slow `ip`/`ethtool`/`sudo -n` (timestamp-refresh prompt, stale netns) blocks the HTTP handler indefinitely. Per R3, netlink itself is unreliable and requires `ENOBUFS` detection + resync; per R4, `ip` exit 2 = kernel-reported error (distinct from syntax 1) — neither signal is surfaced today (error → generic `warn` at `doctor.go:134-137` / `127-129`). Standalone preflight must carry `ctx` + per-exec deadline and return `ctx.Err()` rather than hang; `--timeout` flag is the CLI surface.

### 5.3 `Print()` / CLI ignores the resolved mode (MAJOR)

Observed `doctor.go:68-74`: `Print()` calls `Report(Mode("auto"))` unconditionally, while `Report(mode)` accepts a mode and `cgo --serve --mode` (`cmd/cgo/main.go:64`) resolves one. There is no `cgo doctor --mode full|observe` flag (`cmd/cgo/main.go:72-73` takes no args) and `cgo doctor` usage line (`cmd/cgo/main.go:367`) documents none. Consequence: on Windows, an operator forcing `full` (or on Linux forcing `observe`) still gets the auto verdict — CLI truth diverges from server truth (`cmd/cgo/server.go:304-307` uses the server's resolved `mode`). Against R8 §1 (public API must be precise — a caller cannot predict CLI output from the documented surface), this is a release defect. Fix: `Print(ctx, mode)` / `Report(ctx, mode, …)` threading the operator's mode through (see §6).

### 5.4 Bench topology is hardcoded, not injected (MAJOR)

Observed `doctor.go:108-117`: six `benchOffloadPairs` (`ns_client/c-cli`, `ns_wan_emulator/c-wan`, `ns_wan_emulator/w-edge`, `ns_edge_router/e-wan`, `ns_edge_router/e-srv`, `ns_server/s-edge`) transcribed from `kit/cgo-bench-dualns.sh`. Per R3/R4 the namespace names/interfaces are deployment facts, not kernel constants; a bench rename silently drifts the gate, and absence yields `warn` (`doctor.go:127-129`) rather than a named "topology unknown" verdict. A standalone tool must accept an injected pair list (file or flag) with the six pairs as the embedded default, so `kit/` owns topology and `doctor` owns the gate.

### 5.5 Honesty-limit semantics (MAJOR — preserve, but document as policy)

- CC checks match R6 exactly: `ccChecksFrom` (`doctor.go:160-183`) does exact-token membership (`present["bbr"]`/`present["bbr3"]`), never substring — the shipped fix for the "contains(`bbr`) declared bbr on bbr3-only kernel" bug (`doctor.go:149-151`). Status rule (readable+`bbr`→`ok`; missing `bbr3`→`warn` optional; missing `bbr`→`warn` in observe / `fail` in full; unreadable→2×`warn`) is covered `cc_test.go:5-26`. The "absent ⇒ fail in full" rule is a bench-integrity policy, not a kernel requirement: R6 explicitly says more algorithms "may be available as modules, but not loaded" and therefore absent from the list. "Absent ⇒ bench falsified" is **unverified — primary source not found** as a kernel rule; keep as documented policy (`doctor.go:119-122`).
- Offload gating matches R5 surface: `-k` queries (`doctor.go:132`), any of `tcp-segmentation-offload`/`generic-segmentation-offload`/`generic-receive-offload` `on` ⇒ `fail` (`doctor.go:139-143`), header `Features for …:` skipped only because post-colon state is empty (`doctor.go:89-93` `continue` on `state==""`), not by explicit header filter — fragile to future `ethtool` output changes (R5 output-format stability **unverified — primary source not found** beyond `-k/-K` text). `offload_test.go:17-44` pins current tolerance; `offload_test.go:46-66` pins "check always present, never silent".
- `benchOffloadCheck` returns `warn` (never `ok`) when not on Linux, when `ip netns list` is empty/fails, or when `ethtool -k` is unreadable (`doctor.go:124-137`) — i.e. "bench absent ⇒ `warn`, not `ok`", matching the `n/a ≠ failure` rule `docs/deploy.md:30-34`.

### 5.6 Smaller gaps (MINOR / polish)

- `tc`/`ping` gating is presence-only (`doctor.go:41-49`, `59-63`); `tc(8)` enforcement semantics **unverified — primary source not found**, `ping(8)` usability **unverified — primary source not found**.
- `GOOS`/`GOARCH` os check (`doctor.go:39`) reports `runtime.GOOS` string, not a kernel release — `uname`/BTF availability is **unverified — primary source not found** in this pass.
- `Check` status is an unchecked string (`doctor.go:16-20`); `"Observe"`/`"fail "` typos compile — against R8 §1 precision, needs a `Status` string-type + validated `Mode` enum (see §6).
- `Mode("auto")` default arm treats any unknown string as auto (`doctor.go:28-33`) — a typo like `--mode ful` silently becomes auto (same silent-degrade class as `docs/api.md:8-11` 501 gating).
- Et/Offload long-name equivalence (`doctor.go:139` long forms vs R5 `-K` short `tso/gso/gro`) is **unverified — primary source not found** (R5 documents short `-K` names; `-k` output spelling unverified).

## 6. Target tool: `cgo-doctor` as standalone CLI repo

Proposed identity (R7): new module `github.com/Realms4239/cgo/doctor` (or `/doctor` subdirectory of a `cgo-doctor` repo), `go 1.25` floor from `go.mod:3` (R7 §`go` directive mandatory minimum), MIT `LICENSE` per repo root (`LICENSE:1-3`, R7 §Modules requires LICENSE in initial commit). The repo module is `github.com/Realms4239/cgo` (`go.mod:1`); a standalone path must not collide — `/doctor` suffix path under the same repo root per R7 §Module paths (subdirectory = prefix for tags).

```go
package doctor

type Check struct { Name, Status, Detail string }                          // doctor.go:16 — keep JSON tags
type Mode string; const (Observe Mode = "observe"; Full Mode = "full")     // replaces stringly Mode (doctor.go:24)
type Report struct { Mode string; Checks []Check }                           // replaces (string, []Check) tuple (doctor.go:37)

type Runner interface { // single seam for all exec (4 bare sites doctor.go:41,59,127,132)
    LookPath(file string) (string, error)
    Output(ctx context.Context, name string, args ...string) ([]byte, error)
}
type FileReader func(path string) ([]byte, error)                           // tcp_available_congestion_control (doctor.go:153)
type BenchPair struct{ Ns, Iface string }                                   // exported; default = doctor.go:110-117 six
type CapReader func(pid int) (effective, permitted, inheritable uint64, err error) // capget(2) seam (R2); default CapGetProc

func Mode(override string) string                                           // keep; document "" = auto (doctor.go:28-33)
func Report(ctx context.Context, mode string, r Runner, fr FileReader, cr CapReader, pairs []BenchPair) (Report, error)
func ParseOffloadFeatures(out string) map[string]bool                        // doctor.go:81 (pure, keep exact semantics)
func Print(ctx context.Context, mode string, w io.Writer)                    // fix hardcoded auto (doctor.go:69)
```

CLI (`cmd/cgo-doctor`):

```text
cgo-doctor [--mode observe|full|auto] [--timeout 2s] [--pairs FILE]
           [--format text|json] [--capnetadmin auto|capget|euid]
```

- `--format json` emits `{"mode": …, "checks": [{name,status,detail}]}` — the exact `GET /api/doctor` contract (`pkg/api/server.go:405-411`, `docs/api.md:18`) so `cgo-doctor --format json` ≡ `curl /api/doctor` (diff-verifiable).
- Exit codes (documented, R8 §1): `0` all `ok` or only `warn` with no `fail`; `1` any `fail`; `2` usage/error (mirrors `cmd/cgo/main.go:46`, `64-70` usage/exit vocabulary of the parent CLI; `ip(8)` R4 exit 0/1/2 is the model).
- `capnetadmin` default `capget` (R2); `euid` kept as opt-in fallback preserving today's behaviour (`doctor.go:52`).

Linking rule (R7 §Modules + R8 §1): the standalone module imports **nothing** from `github.com/Realms4239/cgo` (absorb `ParseCCList` from `pkg/model/cc.go:8-18` verbatim); `cmd/cgo` and `pkg/api` then *depend on* `cgo/doctor` (`DoctorFn` at `cmd/cgo/server.go:304-307` becomes `doctor.Report(ctx, mode, …)` with typed `[]doctor.Check`). No `replace` directives, no `exclude`/`retract` needed at birth (R7 §`retract`/`exclude` only apply in main module `go.mod`; `go.mod:5-10` currently has none).

## 7. Slices (ordered, one-line each)

| Slice | Content | Depends on | Acceptance |
|---|---|---|---|
| S0 | Extract: copy `pkg/doctor/*.go` + `ParseCCList` into `cgo/doctor`, new `go.mod` (`module …/doctor`, `go 1.25`), `LICENSE` (MIT, `LICENSE:1-3`); no behaviour change | — | `go test ./...` in new module passes `cc_test.go:5-26` + `offload_test.go:17-44,46-66` unmodified; `GOPROXY=off go list -m all` resolves stdlib-only |
| S1 | Seams: `Runner`/`FileReader`/`BenchPair`/`CapReader` injection; `Report(ctx,…)` + per-exec deadline; `Mode` enum + validated `Status`; `Print(ctx,mode,w)` fix (§5.3) | S0 | existing tests pass with fake `Runner` (no `tc`/`ping`/`sudo`/`ip netns` on test host); `Report` returns within 2 s when `Output` blocks until `ctx` cancel |
| S2 | Capability truth: `capget`-based `cap_net_admin` (R1+R2) with `euid` fallback; detail names effective-set bit + bounding-set state; `capsh` explicitly not shelled out (**unverified — primary source not found** for `capsh` semantics) | S1 | on narrowed-bounding-set fixture, euid=0 + `CAP_NET_ADMIN ∉ effective` ⇒ `fail`/`warn` (not `ok`); `setcap` fixture (non-root + capped) ⇒ `ok` |
| S3 | `cgo-doctor` CLI (`cmd/cgo-doctor`): `--mode/--timeout/--pairs/--format`, exit 0/1/2, JSON ≡ `/api/doctor` contract (`docs/api.md:18`); `docs/api.md:18` + `docs/deploy.md:68-75` bumped for the new binary name | S1 | `cgo-doctor --format json` diffed against `GET /api/doctor` on same host ⇒ identical `mode`+`checks` (modulo `ping`/`tc` PATH drift); `cgo doctor` (old) delegates or prints redirect |
| S4 | Re-point CGO at it: `cmd/cgo` (`main.go:72-73`, `setup.go:217`), `cmd/cgo/server.go:304-307`, `pkg/api/server.go:202-203` import `cgo/doctor`; delete `pkg/doctor` (or thin alias); `docs/deploy.md:30-34,68-75,137,153` updated | S2, S3 | `grep -r pkg/doctor` outside `docs/research` shows zero non-test hits; `make verify` green (all six gates) |
| S5 | Release train: tag `doctor@v0.1.0` (+`Retract` slot reserved per R7 §`retract`); `go.mod` `require` bump in CGO | S4 | `GOPROXY=off go list -m all` per module resolves; `v0.1.0` builds from tag alone (no `git` HEAD); SemVer discipline R8 §6–§8 attested in release notes |

Slice notes: S0 is zero-dependency and cannot break other slices. S1 removes the only internal dep (`model`, `doctor.go:13`). S2 is the honest-capability hinge — it replaces the only fabricated output (euid heuristic). S3 makes the JSON equivalence the release gate. S5 applies R7 (module/`go`/`require`/`exclude`/`retract`) + R8 (immutability §3, MINOR/MAJOR discipline §7–§8) — no tag is ever moved. Total new external deps: **zero** (stdlib + `golang.org/x/sys` only if `capget` needs raw syscall, per R2 `syscall(SYS_capget, …)`).

## 8. Anti-toy criteria (release gates — all must pass)

1. **Capability truth, not euid.** `cap_net_admin` verdict is derived from `capget` effective-set bit `CAP_NET_ADMIN` (R2), reporting effective/permitted/bounding state in `Detail` (R1). A test fixture with `effective` missing the bit must yield `fail`/`warn`, never `ok`. `capsh(8)` shell-out is forbidden (**unverified — primary source not found** — untrusted as sole source).
2. **Timeouts everywhere.** Every `exec` (`LookPath`/`Output` at `doctor.go:41,59,127,132` today) carries caller `context.Context` with a deadline; `Report` returns `ctx.Err()` rather than hanging. A blocking-stub test asserts `Report` returns ≤2 s (§7 S1). `GET /api/doctor` inherits this once re-pointed (S4).
3. **Mode is threaded, not hardcoded.** CLI `--mode`, server `mode`, and `Print`/`Report` agree; `Print()` no longer forces `Mode("auto")` (`doctor.go:69`). A Windows-forced-`full` fixture must render `full` verdicts (exact-token `bbr` fail-in-full per `cc_test.go:14-17`).
4. **Bench topology injected, gate universal.** `BenchPair` list is injectable (S1) with the six pairs (`doctor.go:110-117`) as default; the `bench-offloads` check is always present in every mode (`offload_test.go:46-66` preserved verbatim) and never returns `ok` when the bench is absent (`doctor.go:124-137` → `warn`).
5. **JSON equivalence.** `cgo-doctor --format json` output is byte-equivalent in shape to `GET /api/doctor` (`pkg/api/server.go:405-411`, `docs/api.md:18`): `{"mode", "checks":[{name,status,detail}]}` with `status ∈ {ok,warn,fail}` (`doctor.go:16-20`). Diff on the same host is empty (modulo PATH drift). No fabricated checks, no silent drops (cf. `n/a ≠ failure`, `docs/deploy.md:30-34`).
6. **Pure parser preserved.** `ParseOffloadFeatures` exact semantics (`doctor.go:81-106`) unchanged unless `ethtool -k` output format is re-sourced (R5); header-tolerance changes are PATCH only (R8 §6), any `Status`/`Mode` enum widening is MINOR (R8 §7), any `Report` signature break is MAJOR + `/v2` suffix (R7 §Major version suffixes, R8 §8).

*Skipped: live bench verification, `tc(8)`/`ping(8)` man-page fetches (`unverified — primary source not found`), `capsh(8)` semantics (`unverified — primary source not found`, 404 fetch). Risk: `capget` on pre-2.6.25 kernels needs version probing per R2 — untested here; keep euid fallback (S2) until a real host confirms.*
