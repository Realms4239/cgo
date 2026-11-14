# sadcnwp R0 — CGO bench feasibility read (2026-11-14)

**Scope line:** R0 read-only on sadcnwp for the CGO bench (B) — LXD/snap state,
/dev/kvm, BTF, headroom. Nothing started, installed, or signaled. The ftp2
compromise and production services untouched.
**Window:** 2026-11-14T10:35:41Z (evidence stamp `20261114-103541`).
**Tool:** `python -X utf8 C:\WEBSERVER\safe_read.py --run --file commands/r0-cgo-bench-2026-11-14.txt`
(`-X utf8` required on this host — the tool's buffered pipe decodes with the
console ANSI page otherwise and dies; the guard itself was not edited).
**Identity:** uid=1000(sadc) groups `…,108(lxd),991(docker)`.

## Evidence

- `commands/r0-cgo-bench-2026-11-14.txt` — exact classified command set (all ALLOW).
- `evidence/20261114-103541-read-window.txt` — redacted transcript (sha256 sidecar
  verified byte-for-byte against the file on disk).

## Facts (from the window; load snapshotted live)

| Fact | Observation |
|---|---|
| Host | 2×20c Xeon Gold 6230, 80 threads, 2 NUMA nodes, VT-x |
| Load average | **41.99 / 41.91 / 41.98** — heavy; 6 users on a shared production host |
| Memory | 62 Gi installed; 17 Gi used; 43 Gi buff/cache; ~44 Gi available |
| Swap | 99 Gi, 365 Mi used |
| `/home` | 590 G total, 386 G used, **175 G free** (69%) |
| sadc groups | lxd, docker, sudo, adm, dip |
| `/dev/kvm` | `crw-rw----+ root kvm` — ACL present (contents not readable with the read allow-list; `getfacl` refused) |
| BTF | `/sys/kernel/btf/vmlinux` present, 6 101 415 bytes; `file` = data ✓ eBPF object can be loaded |
| LXD | snap tree present (`lxcfs` running); **`lxd` snap service: inactive**; `snap.lxd.daemon`: inactive; cgroup v2 rw mounted |
| Docker | **active**; many production proxies (mail stack, graylog/lit, nginx/apache, monitoring exports) — shared tenancy |
| Filesystems | cgroup2 rw,nosuid,nodev,noexec; bpf listed in /proc/filesystems |

## Verdict

**B (LXD container bench) is technically possible on this host — not now, and
not without a token.**

- sadc is in the `lxd` and `docker` groups; the LXD snap exists and lxcfs runs;
  cgroups v2 and BTF are in place. A container bench would not need a
  privileged VM.
- **Two hard stops, both human decisions:** (a) `lxd` snap service is inactive —
  starting/configuring anything (bridges, netns, tc, netem, iptables) is R2 with
  a same-day named token, not an R0; (b) load ≈ 42/80 threads is a noisy
  neighborhood for latency measurement — the operator must name a quiet window.
- `/dev/kvm` ACL could not be read with the allow-list (`getfacl` refused) —
  irrelevant for an LXD bench, decisive only if a VM path were required.
- The host is **production** (mail, websites, monitoring, the committed-mining
  ftp2 incident): any bench is scheduled, bounded, and abortable, with the
  operator watching.

**Not run / not decided:** bench start, image creation, netem config, eBPF load.
Those are R1→R2 artifacts with named tokens and a same-day rehearsal — the
2026-11-11 discovery packet and `APPROVED-2026-11-11-labo-cgo-bench.md` stay the
precedent.

## Next

Phase 2 begins with the R2 bench window (named token + rollback + rehearsal) in
a quiet load window — the human names both. Until then Phase 1 proceeds without
touching this host.
