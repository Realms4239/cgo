# CGO — BBRv3 evaluation matrix (slice 7) — design

*Status: proposed · 2026-10-27 · Branch: `rebuild/plan1-economics-audit`*
*Master prompt: `epbfBBR.md`, Directive II.*
*Umbrella spec: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md`, decision #8.*

---

## 1. Problem

The thesis compares three qdiscs against two congestion controls on the
cellular profile P2. Directive II asks for a **third control, BBRv3**, and a
3×3 matrix:

```
{pfifo_fast, fq_codel, cake} × {cubic, bbr, bbr3}
```

Two facts constrain the design, and the master prompt is wrong about both:

1. **BBRv3 is not a loadable congestion-control module.** `google/bbr`
   publishes a full kernel tree. BBRv3 adds callbacks in core TCP (loss
   repair, `tx.in_flight`, `fast_ack_mode`, ECN), so it cannot be dropped into
   stock 6.8 as a `.ko`.
2. **Mainline 6.8 exposes only `bbr` = BBRv1.** As of 2026-10, BBRv3 is still
   unmerged. A stock Ubuntu 24.04 kernel cannot run the third row.

Decision #8 stays binding: **`bbr` in this project means BBRv1**. BBRv3
enters as a new, additive label, never as a redefinition of `bbr`.

## 2. Scope

In scope:

- the `bbr3` congestion-control label and its availability gate;
- kernel provenance in every new matrix config;
- fail-closed handling when a download socket refuses its congestion control;
- an opt-in host script that builds and installs a BBRv3-capable kernel;
- `cgo doctor` reporting `bbr` and `bbr3` separately;
- the documented 3×3 command and its live acceptance protocol.

Out of scope:

- ECN and L4S as experimental axes. BBRv3 is ECN-capable, but this matrix
  does not enable `tcp_ecn` or `ecn_low`; the cells compare loss/delay
  behaviour only, and the report says so.
- IFB ingress, trace-driven `.dist`, eBPF, UI, and thesis chapters.
- Changing `bbr` rows, the frozen corpus, or the CSV schema.
- Running the 3×3 on a stock 6.8 host.

## 3. Verified upstream facts

| Fact | Source |
|---|---|
| BBRv3 is a kernel tree, not a module | `google/bbr` README: build and install a kernel |
| `bbr-v3-2026-09-16-01` is based on Linux 7.1.0-rc5 | branch `Makefile` |
| That branch keeps `tcp_bbr.c` (`bbr`, v1) and adds `tcp_bbr3.c` (`bbr3`) | upstream files, `.name` fields |
| Head commit pinned for this slice | `674859761ded9f32690c7bdaf22ef585531452ed` |
| The 6.8-era branch has the names inverted (`bbr` = v3, `bbr1` = v1) | branch `Makefile` + Kconfig |
| Mainline still ships BBRv1 only | upstream statements, kernel 6.8 config |

The 7.1 branch is chosen because it keeps the natural names and the latest
algorithm. The matrix is self-contained on that kernel; its values are not
merged with the 6.8 frozen corpus.

## 4. Decisions

| # | Decision | Rationale |
|---|---|---|
| D1 | **`bbr` remains BBRv1; `bbr3` is a new label.** `model.AllCC` stays `{cubic, bbr}`; `model.KnownCC` adds `bbr3`. | Preserves decision #8 and every frozen `cc=bbr` row. |
| D2 | **`bbr3` maps one-to-one to the socket name `bbr3`.** No aliasing, no host-specific remapping. | The kernel's `/proc` list, the socket, the CSV label and the report agree. |
| D3 | **Availability is the exact token in `/proc/sys/net/ipv4/tcp_available_congestion_control`.** | `strings.Contains(..., "bbr")` would read `bbr3` as `bbr`; exact tokens only. |
| D4 | **Default matrix:** `cubic`, `bbr`, plus `bbr3` only when registered. A stock host keeps the historical two-CC default. | The directive's "unlock when available" without silently changing a stock host. |
| D5 | **Explicit `--cc bbr3` on a host without it fails closed** with the installer path. | A requested row is never silently dropped or run under another algorithm. |
| D6 | **A refused server-side CC closes the download source.** The cell fails its gates; no silent cubic fallback. | Current code logs and keeps sending with the host default, mislabelling the cell. |
| D7 | **Every new matrix config records the kernel release** (`kernel` field, additive). | `cc=bbr3` is citable only with the host that exposed it. Old manifests are untouched. |
| D8 | **The 3×3 is a documented command, not a new preset.** P2 × three qdiscs × three CCs × 3 reps, direction `up`, 4 flows = 27 events. | No new matrix code; the existing filtered campaign already expresses it. |
| D9 | **The 3×3 is a new run set, not an extension of Tableau 15.** Tableau 15 comes from the legacy 6.8 bench at one flow; the new run is the hardened `dualns` bench at four flows. | Numeric merging across bench and kernel would be a lie. |
| D10 | **Provisioning is an opt-in shell script, never a `cgo run`/`deploy`/`bootstrap` step:** `kit/install-bbr3.sh`. | The measurement binary stays portable and never compiles a kernel. |
| D11 | **No `cgo kit` action for the build.** The script is run through `cgo kit ssh` or the VM console. | `Config.SSH` caps a remote command at 10 minutes; the build takes 30–90. A wrapper would die mid-build. |
| D12 | **The installer builds BBRv1 and BBRv3 into the kernel (`=y`), not as modules.** | `/proc` lists built-ins from boot; no modprobe race before a cell. |
| D13 | **`cgo doctor` prints `bbr` and `bbr3` as separate lines.** `bbr` may be `fail` in full mode; `bbr3` is never `fail` because it is optional. | Honest capability report; the default slice does not require BBRv3. |
| D14 | **No UI change.** The dashboard displays the CC string it receives; a bbr3 row appears when present. | The 3×3 is a CLI/bench experiment; the UI slice comes later. |
| D15 | **Frozen data untouched.** New runs land under `data/runs`, never `book/data/runs`. | Guard stays 159 runs / 208 valid lines. |

## 5. Behaviour

### 5.1 Label and defaults

```go
const (
    Cubic CC = "cubic"
    BBR   CC = "bbr"   // BBRv1 — decision #8
    BBRv3 CC = "bbr3"  // only when the kernel registers it
)

var AllCC   = []CC{Cubic, BBR}          // historical default
var KnownCC = []CC{Cubic, BBR, BBRv3}   // accepted labels
```

`DefaultCCs()` returns `cubic, bbr` and appends `bbr3` when the kernel list
contains that exact token. All three default-expansion sites (`StartMatrix`,
the CLI, the dashboard start) use it, so they cannot disagree.

### 5.2 Explicit request

`startMatrix` validates against `KnownCC`. For `bbr3` it additionally
requires host availability:

```
cc bbr3 indisponible sur cet hôte : le noyau n'expose pas bbr3
(kit/install-bbr3.sh) — refus plutôt qu'exécuter un autre contrôle
sous cette étiquette
```

The refusal happens before any qdisc mutation or run directory creation.

### 5.3 Download path

The download source sets the cell's CC on the accepted socket. A refusal
currently logs `(défaut hôte)` and continues. It becomes:

- log the refusal;
- return immediately, closing the connection;
- the client reads zero bytes, G1/G4 fail, the cell is quarantined.

No protocol change is needed; a closed source is already a failed bulk.

### 5.4 Provenance

`matrixConfig` gains:

```go
Kernel string `json:"kernel,omitempty"`
```

filled from `/proc/sys/kernel/osrelease`. The manifest already embeds the
config JSON and its SHA-256, so the kernel is frozen with the run. The CSV
schema and column order do not change.

A `bbr3` row is therefore citable as:

- `cc=bbr3` in the row;
- `kernel` in the manifest config;
- the installer's `/var/lib/cgo/bbr3-kernel.json` on the host (branch and
  commit), for the thesis appendix.

## 6. The matrix

```bash
CGO_SHAPE_MODE= CGO_BENCH=dualns cgo run \
  --bench dualns --direction up --profiles P2 \
  --qdiscs pfifo_fast,fq_codel,cake \
  --cc cubic,bbr,bbr3 \
  --reps 3 --flows 4 --target 10.30.0.2
```

- 9 cells × 3 reps = 27 events.
- Direction `up` matches the thesis campaign; the download side keeps its
  existing egress path (IFB is slice 6 and not required here).
- The command is run inside `ns_client` on the bench, as the other
  `dualns` campaigns are.
- `--flows 4` is the hardened bench regime. The thesis text must not compare
  these 27 rows numerically with the one-flow legacy cells of Tableau 15.

## 7. Provisioning

`kit/install-bbr3.sh` is the only provisioning artifact.

Contract:

1. Refuse if `bbr3` is already registered (idempotent success).
2. Require Debian/Ubuntu, `apt-get`, `git`, root or `sudo -n`.
3. Refuse Secure Boot enabled (unsigned kernel would not boot).
4. Require ≥ 20 GiB free under the source parent.
5. Install build dependencies (headers toolchain, perl-free build deps,
   `dwarves`, `debhelper`, `fakeroot`).
6. Fetch the pinned commit, verify `HEAD == COMMIT`.
7. Start from `/boot/config-$(uname -r)`, then force:
   - `CONFIG_TCP_CONG_BBR=y` (v1)
   - `CONFIG_TCP_CONG_BBR3=y` (v3)
   - `CONFIG_TCP_CONG_CUBIC=y`
   - `CONFIG_DEFAULT_TCP_CONG="cubic"` (the campaign sets per socket)
   - `CONFIG_LOCALVERSION="-cgo-bbr3"` (visible in `uname -r`)
   - no trusted/revocation keys (upstream tree has none)
   - netem, CAKE, fq_codel, `matchall`, `mirred`, IFB (the bench needs them)
8. Verify every forced option after `olddefconfig`; die if one was dropped.
9. Build `bindeb-pkg`, install the matching `linux-image` deb.
10. Write `/var/lib/cgo/bbr3-kernel.json` with branch, commit, release, date.
11. Print the reboot command; `--reboot` is opt-in.

The script never runs from `cgo run`, `cgo deploy`, or `cgo kit bootstrap`.
It is invoked explicitly, on the VM, after a snapshot.

## 8. Doctor

```
[bbr]  disponible (BBRv1, mainline)
[bbr3] disponible (BBRv3)           # or: absent — optionnel : kit/install-bbr3.sh
```

`bbr` keeps its full-mode `fail` when absent. `bbr3` is `warn` when absent,
`ok` when present. An unreadable `/proc` file marks both `indéterminé`
(`warn`), never "absent".

## 9. Acceptance

### 9.1 Windows (engine proof)

- `go test ./... -count=1`
- `go vet ./...`
- `GOOS=linux go build ./...`
- `bash -n kit/install-bbr3.sh`
- `python book/tools/audit-archive.py` → 159 runs / 208 valid lines

Tests cover: the CC list parser, default selection with and without `bbr3`,
explicit-request refusal and acceptance, kernel provenance in the config,
and the doctor's token-exact separation.

Windows cannot prove that a 7.1 kernel boots, that `/proc` lists `bbr3`, or
that sockets actually use it. That is live acceptance.

### 9.2 Linux (blocked)

1. Snapshot the VM, then `sudo bash kit/install-bbr3.sh`.
2. Reboot; `uname -r` contains `-cgo-bbr3`.
3. `cgo doctor` shows `bbr` and `bbr3` available.
4. Run the 27-event command inside `ns_client`.
5. Verify: 27 rows; `ccs` and `kernel` in the manifest config; at least one
   `bbr3` row per qdisc; `cc=bbr3` never appears without the BBRv3 kernel.
6. Record the exact commands and outputs.

No cell is called "valid" because it ran; the G0–G7 gates decide, and CUBIC
cells may be invalid on a lossy P2 exactly as in the thesis.

## 10. Risks

| Risk | Mitigation |
|---|---|
| 7.1 kernel does not boot or lacks a VM driver | old kernel stays in GRUB; snapshot before install; `--reboot` is opt-in |
| Build fails late (30–90 min) | forced options verified before the build; full log path printed |
| Secure Boot rejects the unsigned kernel | checked before building |
| 7.1 changes netem/CAKE behaviour vs 6.8 | the 3×3 is self-contained; cross-kernel numeric merging is forbidden |
| Operator lies about the host | `/proc` token + kernel release in the manifest; provenance file on the host |
| The branch moves | commit pinned; the script verifies `HEAD` |
| Download CC silently falls back | closed source + failed gates (D6) |
| Disk fills | 20 GiB preflight; build dir path printed |

## 11. Frozen corpus

No file under `book/` changes. `book/tools/audit-archive.py` keeps counting
159 runs, 528 rows, 208 valid. New runs go to `data/runs/<run-id>/`, and the
thesis must label them as a separate 7.1-kernel run set.

---

## 12. Open questions

None blocking. The two choices this spec depends on were answered by the
author: additive `bbr3` gated by the kernel, and an opt-in 7.1 kernel build.
