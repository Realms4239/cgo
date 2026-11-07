# deploy/Dockerfile + deploy/docker-compose.yml + docs/deploy.md S4 as standalone cgroups/container tool — primary-source research

*Date: 2026-11-07 · Scope: `deploy/Dockerfile`, `deploy/docker-compose.yml`, `docs/deploy.md` §S4 (`docs/deploy.md:182-183`) read as if extracted tomorrow into `cgo-isolate` / an OCI distribution package. Research only — no code changed, no live bench run.*

## 0. Method & sources (table)

| # | Source (primary) | URL | Status | Used for |
|---|---|---|---|---|
| R1 | Linux Kernel — Control Group v2 | https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html | fetched (full page, `cpu.max`/`cpu.stat`/`memory.max`/`memory.events` sections) | `cpu.max` format/default, `cpu.stat` throttling fields, `memory.max` hard limit/OOM, `memory.events`, single hierarchy, `subtree_control` |
| R2 | Docker — Running containers (`docker run` runtime constraints) | https://docs.docker.com/engine/containers/run/ | fetched | `--cpus=0.000` no-limit, `--cpu-period`/`--cpu-quota`, `--memory` min/format, `--memory-swap` semantics, `--memory-reservation`, `--oom-kill-disable`, `NET_ADMIN` capability, `EXPOSE`/`-P` publish semantics |
| R3 | Docker — Resource constraints | https://docs.docker.com/config/containers/resource_constraints/ | fetched | `--cpus=1.5` ⇔ `--cpu-period=100000 --cpu-quota=150000`, `--memory` min 6m, OOM-muting guidance, swap/reservation semantics |
| R4 | Docker — Compose Deploy Specification (`deploy.resources`) | https://docs.docker.com/reference/compose-file/deploy | fetched (via search extract) | `limits` MUST prevent / `reservations` MUST guarantee, `cpus`/`memory`/`pids` fields, deploy optional — ignored if not implemented |
| R5 | OCI Runtime Spec — Linux configuration (`config-linux.md`) | https://github.com/opencontainers/runtime-spec/blob/main/config-linux.md | fetched (via search extract) | `resources.cpu{shares,quota,burst,period,cpus,mems}`, `resources.memory{limit,reservation,swap,swappiness,disableOOMKiller,checkBeforeUpdate}` |
| R6 | OCI Image Spec — Content Descriptors / Digests | https://github.com/opencontainers/image-spec/blob/main/descriptor.md | fetched (via search extract) | Descriptor `mediaType`/`digest`/`size`, digest grammar `sha256:<64 hex>`, verification before consuming untrusted content |
| R7 | OCI Distribution Spec — Pull | https://github.com/opencontainers/distribution-spec/blob/main/spec.md | fetched (via search extract) | Manifest reference MUST be digest or tag, `Docker-Content-Digest` header, clients SHOULD verify returned manifest/blob against requested digest |
| R8 | Open Container Initiative (overview) | https://opencontainers.org/ | fetched | Three specs exist: Runtime (run a filesystem bundle), Image, Distribution — implementation downloads image, unpacks to runtime bundle, runtime runs it |
| R9 | Docker — Dockerfile reference | https://docs.docker.com/reference/dockerfile/ | fetched (via search extract) | `FROM`/`COPY`/`EXPOSE` (documentation, does not publish)/`ENTRYPOINT` exec form preferred, `FROM scratch` = first filesystem layer, binary must have no runtime deps |
| R10 | Docker — Base images / `FROM scratch` | https://docs.docker.com/build/building/base-images | fetched (via search extract) | `scratch` cannot be pulled/run/tagged, only referenced in Dockerfile; image works only if binary has no runtime dependencies (no loader, no CA certs, no shell) |

*Method: read `deploy/Dockerfile:1-4`, `deploy/docker-compose.yml:1-13`, `docs/deploy.md:1-183` (esp. `docs/deploy.md:182-183` S4, `docs/deploy.md:159-177` security line), `pkg/api/metrics.go:1-64`, `deploy/nginx/cgo.conf:1-35`, `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:15,37,72,79`. Every technical claim below cites `file:line` or `(R#)`; anything not traceable to a fetched primary is marked **unverified — primary source not found**.*

## 1. Responsibility

Today the trio is a **packaging claim, not a tested isolate**:

- `deploy/Dockerfile:1-4` — four lines: `FROM scratch`, `COPY bin/cgo /cgo`, `EXPOSE 9090`, `ENTRYPOINT ["/cgo", "--serve"]`. Its entire job is "ship one static binary with no filesystem". No user, no healthcheck, no capabilities, no build stage.
- `deploy/docker-compose.yml:1-13` — one service `cgo` (`deploy/docker-compose.yml:2-3`): image `ghcr.io/realms4239/cgo:latest`, loopback-only publish `127.0.0.1:9090:9090` (`deploy/docker-compose.yml:4-5`), `deploy.resources.limits` `cpus: "1.0"` + `memory: 512M` (`deploy/docker-compose.yml:6-10`), `CGO_API_TOKEN` required env (`deploy/docker-compose.yml:11-12`), `restart: unless-stopped` (`deploy/docker-compose.yml:13`). Its job is "run that image with a CPU/memory ceiling and a token".
- `docs/deploy.md:182-183` (S4) — the *story*: "`docker compose up` starts the digest-pinned image. Rollback is `compose pull` of the prior digest + `cgo shape --restore`. `cpu.max`/`memory.max` are set in compose; throttling stats are exported via `/metrics`."

What the trio deliberately does **not** do: it does not write any cgroup file itself (no `cpu.max`/`memory.max` path, no `/sys/fs/cgroup/...` open — grep `sys/fs/cgroup|cpu\.stat|memory\.events|throttled` in `pkg/` and `cmd/` returns no matches), it does not read `cpu.stat`/`memory.events`, it does not emit any throttling series (`pkg/api/metrics.go:46-53` emits six static `"0"` samples, none cgroup-derived), it does not pin a digest (`:latest` at `deploy/docker-compose.yml:3`), it does not grant `NET_ADMIN` (required for `tc`/netem shaping per `docs/deploy.md:79-87`), and it does not include Nginx/Prometheus/Grafana services despite S1/S3/S5 implying them (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:34-38`).

Read as a standalone tool, the job would be: **produce a reproducible, digest-pinned OCI image of `cgo` and enforce + observe kernel-enforced CPU/memory isolation for it, with rollback by digest.** §5 shows the current files assert that job in prose but do not implement it.

## 2. Exported config surface (file:line)

### 2.1 `deploy/Dockerfile` — image build surface

| Instruction | Value | Location | Primary-source meaning |
|---|---|---|---|
| `FROM scratch` | empty base | `deploy/Dockerfile:1` | First filesystem layer; no pulled base image; binary must be self-contained (R9, R10) |
| `COPY bin/cgo /cgo` | one binary | `deploy/Dockerfile:2` | Copies from build context to `/cgo` (R9 `COPY`) |
| `EXPOSE 9090` | TCP 9090 documented | `deploy/Dockerfile:3` | Documentation only — does not publish; `-P` publishes only exposed ports (R9 `EXPOSE`, R2) |
| `ENTRYPOINT ["/cgo", "--serve"]` | exec form | `deploy/Dockerfile:4` | Preferred exec form: binary is PID 1, receives signals directly (R9 `ENTRYPOINT`) |

Absent from the file (not API, but part of the surface a standalone tool would need): no `USER`, no `HEALTHCHECK`, no `LABEL`, no `STOPSIGNAL` — all are Dockerfile instructions (R9 table lists `USER`/`HEALTHCHECK`/`LABEL`/`STOPSIGNAL` as supported instructions). No `--platform`, no multi-stage build, no digest on `FROM` (R9 `FROM` allows `<image>[@<digest>]`).

### 2.2 `deploy/docker-compose.yml` — runtime config surface

| Key | Value | Location | Primary-source meaning |
|---|---|---|---|
| `services.cgo.image` | `ghcr.io/realms4239/cgo:latest` | `deploy/docker-compose.yml:3` | Mutable tag reference, not a digest reference (`name:tag[@digest]` form, R2/R7) |
| `ports` | `127.0.0.1:9090:9090` | `deploy/docker-compose.yml:4-5` | Explicit publish, host loopback only (R2 `-p` semantics) |
| `deploy.resources.limits.cpus` | `"1.0"` | `deploy/docker-compose.yml:9` | Fractional CPU limit; platform MUST prevent exceeding it (R4) |
| `deploy.resources.limits.memory` | `512M` | `deploy/docker-compose.yml:10` | Byte-string memory limit; platform MUST prevent exceeding it (R4) |
| `environment` | `CGO_API_TOKEN=${CGO_API_TOKEN:?set}` | `deploy/docker-compose.yml:11-12` | Required-at-up env var; compose interpolation failure if unset — compose behaviour `unverified — primary source not found` in this pass (not fetched) |
| `restart` | `unless-stopped` | `deploy/docker-compose.yml:13` | Top-level service key, not `deploy.restart_policy` (R4 deploy vs service keys) |

Absent: no `deploy.resources.reservations`, no `deploy.resources.limits.pids` (R4 defines `pids` as a `limits` field), no `cap_add: [NET_ADMIN]`, no `devices`, no `healthcheck`, no second service (no Nginx, no Prometheus).

### 2.3 `docs/deploy.md` S4 — claimed surface (`docs/deploy.md:182-183`)

Four claims, each checked in §5:

1. `docker compose up` starts the **digest-pinned** image.
2. Rollback = `compose pull` of the **prior digest** + `cgo shape --restore`.
3. `cpu.max`/`memory.max` are **set in compose**.
4. Throttling stats are **exported via `/metrics`**.

## 3. Dependencies

### 3.1 Outgoing (what the tool needs)

- Build-time: one file `bin/cgo` on the build host (`deploy/Dockerfile:2`), plus Docker/BuildKit to execute `FROM`/`COPY` (R9). **Whether `bin/cgo` is statically linked is `unverified — primary source not found`** (no `go.mod`/`goreleaser` build flags were read for the binary in this pass; R10 requires a `scratch` binary to have no runtime dependencies). If it is dynamically linked, the image cannot start — R10 is explicit that a `scratch` image "can only be successfully executed as long as the `hello` binary doesn't have any runtime dependencies".
- Runtime: an OCI-compliant container runtime to honour `deploy.resources.limits` (R4: "the platform must prevent the container from allocating more resources" — the *platform*, not compose, enforces it) and a Linux kernel with cgroup v2 controllers enabled (R1: a controller is available only when "supported by the kernel (i.e., compiled in, not disabled and not attached to a v1 hierarchy) and listed in `cgroup.controllers`").
- Registry: `ghcr.io` reachable at pull time (R7 pull model: client fetches manifest by reference, then blobs by digest).

### 3.2 Incoming (who needs it)

- `docs/deploy.md:130-147` (§4 mini-PC edge recipe) and the hardening spec `docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:37,72` (S4 MUST: "multi-arch OCI image digest-pinned; compose wiring dashboard + Nginx (+ exporters); systemd preserved as fallback; `cgo verify` + `cgo doctor` + `cgo shape --restore` as post-deploy gates. Rollback = prior digest").
- `deploy/nginx/cgo.conf:17-22` proxies `/metrics` loopback-only — the consumer that *would* scrape throttling stats if they existed.
- `pkg/api/metrics.go:43-63` — the endpoint S4 claim 4 points at; it currently emits no cgroup series (§5.4).

### 3.3 Licence

Repo root `LICENSE` is MIT ("Copyright (c) 2026 Meteolink contributors", per sibling research) — **licence text not re-read in this pass, portability beyond this repo is `unverified — primary source not found`.**

## 4. Primary-source requirements

### 4.1 cgroup v2 — `cpu.max` (R1)

R1 defines `cpu.max` as a read-write two-value file on non-root cgroups, default `"max 100000"`, format `$MAX $PERIOD` meaning "the group may consume up to $MAX in each $PERIOD duration", `"max"` for `$MAX` = no limit; writing one number updates `$MAX` only (R1, `cpu.max`). All durations are microseconds (R1, CPU interface files preamble). It affects only fair-class/BPF processes (R1).

Consequence: `cpus: "1.0"` (R4) maps to `cpu.max = "100000 100000"` (1.0 × 100000µs period) — R3 gives the exact equivalence for 1.5 CPUs (`--cpu-period=100000 --cpu-quota=150000`), and R2 gives `--cpus=0.5` ⇔ `--cpu-period=100000 --cpu-quota=50000`. The mapping *through* Docker is real; the *file* `cpu.max` is never written by any repo code (§3.1).

Throttling observation (R1, `cpu.stat`): when the controller is enabled, `cpu.stat` reports `nr_periods`, `nr_throttled`, `throttled_usec`, `nr_bursts`, `burst_usec` — "non-hierarchical; they only account for throttling caused by this cgroup's own bandwidth limit, not including throttling inherited from ancestor cgroups" (R1). `cpu.stat.local` reports the local `throttled_usec` *including* inherited throttling (R1, `cpu.stat.local`). A standalone tool that wants "throttling stats" must pick one and label it — mixing them double-counts.

### 4.2 cgroup v2 — `memory.max` (R1)

R1 defines `memory.max` as a read-write single-value file, default `"max"`, "Memory usage hard limit. This is the main mechanism to limit memory usage of a cgroup. If a cgroup's memory usage reaches this limit and can't be reduced, the OOM killer is invoked in the cgroup" (R1, `memory.max`). Usage may temporarily exceed the limit (R1). `memory.events` counts `low`/`high`/`max`/`oom`/`oom_kill`/`oom_group_kill` (R1, `memory.events`) — hierarchical by default, local variant is `memory.events.local` (R1).

Consequence: `memory: 512M` (R4) is a *hard* limit, not a soft one. Docker's own docs warn that without limits the kernel OOM-kills "any process… including Docker"; Docker "adjusts the OOM priority on the Docker daemon so that it's less likely to be killed" but "the OOM priority on containers isn't adjusted" (R3, "Understand the risks"). A 512M kill is the *expected* behaviour at the ceiling, not an error — the tool must surface `memory.events.max`/`oom_kill` rather than treat exit 137 as a mystery. Exit-code-137 semantics itself: **unverified — primary source not found** in the fetched Docker pages (no `137` in R2/R3 extracts).

### 4.3 Docker CPU/memory flags (R2, R3)

- `--cpus` default `0.000` means **no limit** (R2, R3). `cpu-period` default is 100000µs (R3).
- `--memory` minimum is **6M** (R2 table "Minimum is 6M", R3 "`-m` or `--memory=`… minimum allowed value is `6m`"). 512M clears it comfortably.
- `--memory-swap` semantics are total (memory+swap), not additive: `--memory=300m --memory-swap=1g` ⇒ 300M RAM + 700M swap; unset swap ⇒ swap = memory (R2, R3). `docker-compose.yml` sets neither swap key — so on a swap-enabled host the effective ceiling is **1024M (512M RAM + 512M swap)**, not 512M. Whether the target host has swap: **unverified — primary source not found** (host-dependent).
- `--memory-reservation` is a **soft** limit, "doesn't guarantee that the container doesn't exceed the limit" (R2, R3). Compose `reservations` likewise: platform MUST guarantee at least that amount (R4) — a different guarantee direction from the kernel's soft-limit semantics. The repo sets no reservation — correct to omit rather than mis-set, but a standalone tool should declare the choice.
- `--oom-kill-disable` "only disable… on containers where you have also set the `-m/--memory` option. If the `-m` flag is not set, this can result in the host running out of memory" (R2). Repo does not set it — correct.

### 4.4 Compose `deploy.resources` (R4)

R4: `limits` = "The platform must prevent the container from allocating more resources"; `reservations` = "The platform must guarantee the container can allocate at least the configured amount" (R4, `resources`). `cpus` "configures a limit or reservation for how much of the available CPU resources, as number of cores"; `memory` "set as a string expressing a byte value" (R4).

Critical caveat, verbatim from the Compose specification as fetched: "Deploy support is an optional aspect of the Compose Specification… **If not implemented the `deploy` section is ignored and the Compose file is still considered valid**" (R4, "Define services… `deploy`"). So `deploy.resources.limits` is a *request to the platform*, not a kernel guarantee. A standalone tool that claims "limits enforced" must verify with `docker inspect` (`HostConfig.Memory`/`NanoCpus`) or by reading the cgroup files — compose alone does not prove enforcement. Whether current `docker compose up` (v2 plugin) applies `deploy.resources` on a non-Swarm host: **unverified — primary source not found** in the fetched R4 extract (the historical v1/Swarm caveat is not in the fetched text; treat "applied" as needing the inspect check in §7).

### 4.5 OCI Runtime Spec (R5)

The runtime config that *actually* carries limits is `linux.resources`: `cpu{shares,quota,burst,period,realtimeRuntime,realtimePeriod,cpus,mems}` and `memory{limit,reservation,swap,kernel,kernelTCP,swappiness,disableOOMKiller,useHierarchy,checkBeforeUpdate}` (R5). Notable: `checkBeforeUpdate` exists precisely because "when cgroup v2 is used, an OOM killer is invoked" [for a limit below current usage] (R5). `quota`/`burst`/`period` are microseconds (R5), matching R1 `cpu.max` units. R5 also: "Runtimes MAY attach the container process to additional cgroup controllers beyond those necessary" (R5) — so the container's cgroup tree may contain more than `cpu`/`memory`.

### 4.6 OCI Image + Distribution Specs (R6, R7, R8)

- OCI is three specs — Runtime, Image, Distribution (R8). The runtime runs a "filesystem bundle"; an implementation "would download an OCI Image then unpack that image into an OCI Runtime filesystem bundle" then run it (R8). The Dockerfile's `scratch`+`COPY` output must therefore be a valid image manifest: `config` descriptor + ordered `layers` descriptors (R6, Image Manifest: "The array MUST have the base layer at index 0… The final filesystem layout MUST match the result of applying the layers to an empty directory").
- A descriptor's `digest` is `algorithm:encoded`, e.g. `sha256:<64 hex>` (R6, Digests). "Retrieved content SHOULD be verified against this digest when consumed via untrusted sources" (R6). "Compliant implementations SHOULD use SHA-256… Implementations MUST implement SHA-256 digest verification" (R6).
- Distribution pull: the reference "`MUST be either (a) the digest of the manifest or (b) a tag. The `MUST NOT be in any other format`" (R7). "If the ` part of a manifest request is a digest, clients SHOULD verify the returned manifest matches this digest" (R7). **`:latest` is case (b), a mutable tag — by R7 the registry may return a different manifest for the same tag tomorrow. "Digest-pinned" means case (a) `@sha256:…` and nothing else.**
- Tag vs digest: "A manifest digest may have zero, one, or many tags referencing it" (R7) — so pinning by digest does not require the tag to exist at pull time.

### 4.7 Dockerfile semantics (R9, R10)

- `EXPOSE` "doesn't actually publish the port… functions as a type of documentation" (R9, `EXPOSE`). Publish happens via `-p`/`-P` (R2). `deploy/docker-compose.yml:4-5` does the real publish; `deploy/Dockerfile:3` is inert documentation. Both are correct — but a standalone tool must not claim `EXPOSE` *secures* anything. Loopback binding (`127.0.0.1:9090`) is the actual access control, consistent with `docs/deploy.md:172-174` ("never exposed on Internet").
- `FROM scratch` "signals… that you want the next command in the `Dockerfile` to be the first filesystem layer" and the resulting image "can only be successfully executed as long as the `hello` binary doesn't have any runtime dependencies" (R10).
- Exec-form `ENTRYPOINT` is "the preferred form" (R9) — `deploy/Dockerfile:4` complies.

## 5. Gap analysis (current → releasable)

Ordered by severity. Each gap cites local `file:line` and the R# it breaks.

1. **`:latest` is not digest-pinned — S4 claim 1 is false.** `deploy/docker-compose.yml:3` uses `ghcr.io/realms4239/cgo:latest`. By R7 a tag reference is mutable; by R7 digest references are the only immutable form. The hardening spec's own MUST (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:72` "multi-arch OCI image digest-pinned") is violated by the very file that claims to implement it. No `@sha256:` appears anywhere in `deploy/`.
2. **No prior digest exists to roll back to — S4 claim 2 is inoperable.** Rollback by digest requires (a) the current digest recorded somewhere machine-readable and (b) the prior digest recorded before upgrade. Neither is in the repo: no digest file, no `docker inspect` output, no manifest annotation. `docs/deploy.md:183` says "compose pull of the prior digest" but names no source for that digest. By R6/R7 the rollback value *could* be recorded as a descriptor; it currently is not.
3. **No shaping capability — the container cannot do the product's core job.** Shaping (`tc`/netem) requires `CAP_NET_ADMIN` (`docs/deploy.md:79-87`: "Le façonnage exige CAP_NET_ADMIN"). R2 confirms `NET_ADMIN` is not in Docker's default capability set and must be added via `--cap-add=NET_ADMIN`. `deploy/docker-compose.yml` has no `cap_add`. As shipped, the container runs in observation-mode-equivalent for shaping — the 501 path, not the shaping path. This also breaks the systemd-fallback story: the binary path (`setcap`, `docs/deploy.md:82-87`) and the container path diverge silently.
4. **No throttling stats are exported — S4 claim 4 is false.** `pkg/api/metrics.go:46-53` emits six static `"0"` series; none reads `/sys/fs/cgroup/.../cpu.stat` or `memory.events`. R1 gives the exact fields (`cpu.stat`: `nr_throttled`/`throttled_usec`; `memory.events`: `max`/`oom_kill`). S4's "throttling stats are exported via `/metrics`" describes a feature that does not exist. `deploy/nginx/cgo.conf:17-22` proxies `/metrics` but there is nothing cgroup-derived to proxy.
5. **`cpu.max`/`memory.max` are "set in compose" only by delegation, and may be silently ignored.** Compose expresses intent (`deploy.resources.limits`, R4); the platform enforces it (R4), and R4 explicitly allows implementations to ignore `deploy` while the file remains valid. No `docker inspect` check, no cgroup-file read-back, and no `cap_add`/device declaration exist. Additionally, unset swap means effective ceiling is 2× the stated RAM on swap-enabled hosts (R2/R3 swap semantics, §4.3). A standalone tool must prove enforcement, not assert it.
6. **`scratch` + no `USER` = root by default, with no discussion.** R9 lists `USER` as a supported instruction; the file sets none, so the process runs as root inside the container. Combined with gap 3 (no `NET_ADMIN`), this is the worst of both: root without the one capability the product needs, and root remains a risk if any future `cap_add` *is* granted. `docs/deploy.md:159-177` already documents "no authN/Z, anyone reaching the port can reshape" — running as root widens that blast radius. A non-root `USER` (numeric UID, since `scratch` has no `/etc/passwd`) is a one-line fix in a multi-stage build.
7. **No healthcheck — "green health" cannot be checked by the platform.** The hardening spec's acceptance is "`docker compose up` → green health" (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:146`). R9 lists `HEALTHCHECK` as a supported instruction and R2 documents healthcheck override semantics; the image has none. Compose has no `healthcheck:` key either. `GET /api/health` exists (`docs/deploy.md:16`), but nothing in the container stack calls it.
8. **No multi-arch build — "multi-arch OCI image" (spec) vs single `bin/cgo` (file).** `deploy/Dockerfile:2` copies one host-built `bin/cgo`. R6's image index / platform descriptor mechanism exists precisely for multi-arch (R6, "allow multi-architecture images"). A `cgo-isolate` that can only build for the build host's arch is a toy for edge deployment (mini-PC `amd64` vs bench VM — `docs/deploy.md:130-147`).
9. **`restart: unless-stopped` at top level ≠ `deploy.restart_policy`.** R4: `deploy` groups "deployment strategy"; `restart` at service level is a compose-local concern. If the same file is ever used with `docker stack deploy` (Swarm), `restart` is not supported there (**stack/Swarm `restart` key behaviour: unverified — primary source not found** in this pass). Minor, but a standalone package should use the key its target platform honours.
10. **No `CGO_API_TOKEN` rotation / no secret hygiene note.** `deploy/docker-compose.yml:12` requires the token at `up` time (`${…:?set}`). That is good fail-fast behaviour. What is missing: any statement of where the token comes from, its lifetime, or whether the *image* embeds it. R2 notes environment variables set via `-e` are visible in the container — the compose file does not mark the variable sensitive. (Compose secret-file semantics: **unverified — primary source not found**.)
11. **`docs/deploy.md` S4 is two bullet lines with no runbook.** No `docker inspect` command, no cgroup read-back (`cat /sys/fs/cgroup/.../cpu.stat`), no `docker stats`, no digest-recording step. The hardening spec wants runbooks (`docs/superpowers/specs/2026-11-06-cgo-enterprise-hardening-design.md:80`); S4 as written cannot be executed by a 3 a.m. operator.

## 6. Target tool: `cgo-isolate` + OCI distribution package

Two deliverables, kept separate because they fail differently:

### 6.1 `cgo-isolate` — a tiny cgroup-v2 helper (no Docker required)

A single static Go binary (stdlib only) that the systemd path and the container path share:

- **Flags**: `--cgroup-root` (default `/sys/fs/cgroup`), `--name` (default `cgo`), `--cpus` (float, default `1.0`), `--memory` (bytes or `512M` suffix, default `512M`), `--memory-swap` (explicit, default = memory, i.e. no swap — killing the §5.5 ambiguity), `--pids` (int, default `unverified — primary source not found` value to be chosen, e.g. 256), `--exec` (command to run inside the cgroup).
- **Behaviour** (all R1): create child cgroup under the root (R1 "mkdir $CGROUP_NAME"), enable `cpu`+`memory` via parent `cgroup.subtree_control` (R1 "no controller is enabled by default"; write `+cpu +memory`), write `cpu.max = "<quota> <period>"` with period 100000µs (R1 default, R3 equivalence), write `memory.max = <bytes>` (R1 hard limit), write the child PID to `cgroup.procs` (R1 migration semantics: one PID per write, all threads move together), exec under `nsenter`-free direct fork (no namespace mgmt — **namespace creation semantics: unverified — primary source not found** in this pass; defer to `unshare` if ever needed).
- **Observe**: read `cpu.stat` (`nr_throttled`/`throttled_usec`, R1) and `memory.events` (`max`/`oom_kill`, R1) from the child cgroup, emit as Prometheus exposition on the existing `/metrics` shape. **Absent cgroup ⇒ no series, never zero** (mirrors `pkg/api/metrics.go:15-16` "Absent is never emitted").
- **Cleanup**: on exit, `rmdir` the cgroup only when empty (R1: "a cgroup which doesn't have any children or live processes can be destroyed by removing the directory").

This is what makes S4 claim 3 *literally* true: `cpu.max`/`memory.max` files exist, are read back, and are tested.

### 6.2 OCI distribution package — the image + compose fix

- **Dockerfile** (multi-stage, R9): `FROM golang AS build` → `CGO_ENABLED=0 go build -o /cgo` → `FROM scratch` → `COPY --from=build /cgo /cgo` → `USER 65534` → `EXPOSE 9090` → `ENTRYPOINT ["/cgo","--serve"]` + `HEALTHCHECK CMD ["/cgo","healthcheck"]` (R9 instruction table). Multi-arch via `--platform=$BUILDPLATFORM` cross-compile (mechanism: **unverified — primary source not found** for BuildKit specifics; standard pattern, flagged).
- **Compose** (R4): `image: ghcr.io/realms4239/cgo@sha256:<digest>` (R7 case (a)), `cap_add: [NET_ADMIN]` (R2), `deploy.resources.limits.pids`, explicit `memswap_limit`, `healthcheck:` calling `GET /api/health`, plus a `digest.txt` sidecar recording the *previous* digest before any upgrade (R6 descriptor shape: `{"digest":"sha256:…","size":…}`).
- **Rollback runbook** (one page, not two lines): `docker inspect` the running digest → write it to `digest.txt` → edit compose → `docker compose pull && docker compose up -d` → `GET /api/health` → `cgo shape --restore` (host-side, since the container's `NET_ADMIN` is netns-scoped — **whether container `NET_ADMIN` reaches the host qdisc: unverified — primary source not found**, flag for the owner).

## 7. Slices (proposed, each with acceptance)

1. **S1 — Freeze + declare.** README declares the contract: exact `cpu.max`/`memory.max` values, period 100000µs (R1/R3), swap policy, `pids` value, `NET_ADMIN` rationale, digest-pinning rule (R7). Move the three files verbatim. Acceptance: README lists every §2 row; `docker compose config` exits 0.
2. **S2 — Digest-pin + inspect gate.** Replace `:latest` with `@sha256:<digest>` (R7); add `digest.txt` recording prior digest; add CI step `docker inspect --format '{{.HostConfig.Memory}} {{.HostConfig.NanoCpus}}'` asserting 536870912 / 1000000000 (512M, 1.0 CPU in nanos — R3 semantics). Acceptance: `docker compose pull` with a tampered digest fails; inspect output matches §2.2.
3. **S3 — `cgo-isolate` observe-only.** Implement cgroup create + read-back of `cpu.max`/`memory.max`/`cpu.stat`/`memory.events` (R1) with no exec yet; unit test against a fake cgroup root. Acceptance: read-back values equal the requested values; absent cgroup ⇒ no output, never zero.
4. **S4 — Wire throttling to `/metrics`.** Emit `cgo_cgroup_cpu_throttled_usec`, `cgo_cgroup_cpu_nr_throttled`, `cgo_cgroup_memory_max_events`, `cgo_cgroup_memory_oom_kill_total` from R1 fields, labelled by cgroup name — **not** from the static zeroes in `pkg/api/metrics.go:46-53`. Acceptance: `curl localhost:9090/metrics` under CPU load shows non-zero `throttled_usec` (R1 non-hierarchical definition documented); idle host shows absent, not zero.
5. **S5 — Capability + user fix.** Add `cap_add: [NET_ADMIN]` (R2) with a comment citing `docs/deploy.md:79-87`, add `USER 65534` + multi-stage build (R9), add `HEALTHCHECK` (R9). Acceptance: `docker exec cgo tc qdisc show` succeeds from inside the container (**in-container `tc` presence: unverified — primary source not found** — `scratch` has no `tc`; the healthcheck proves liveness, `tc` proves shaping — one of the two must be replaced by a host-side `cgo shape --restore` step and the README must say which).
6. **S6 — Rollback drill.** Execute S4 claim 2 end-to-end on a throwaway host: record digest, upgrade, roll back by digest, verify `/api/health` + `cgo shape --restore`. Acceptance: the drill output (digests before/after, health JSON) is pasted into `docs/CHECKPOINTS.md` — the VERIFY row moves only on this evidence, per `AGENTS.md` roles.

*Each slice touches only `deploy/` + the new `cgo-isolate` tree; `pkg/` measurement code is untouched. Slices 2–5 change runtime behaviour and get verifier-re-run before any `VERIFIED` row is claimed.*

## 8. Anti-toy criteria (what would make this unacceptable as a standalone tool)

1. **Tag is not a pin.** Any `:latest` (or any bare tag) in the released compose file fails review — R7 makes tags mutable, R6 makes digest verification a SHOULD for untrusted sources. The only acceptable reference is `name@sha256:<64 hex>`.
2. **Asserted limits without read-back are decoration.** If `cpu.max`/`memory.max` (or their Docker equivalents) are not read back from `/sys/fs/cgroup/...` or `docker inspect` in CI, the limit is a comment, not a control (R4: platform-enforced, R4: deploy may be ignored).
3. **Swap ambiguity is a toy.** Shipping `memory: 512M` with unset swap and claiming "512M ceiling" is wrong on swap-enabled hosts (R2/R3: effective = memory + swap). Either set `memswap_limit` explicitly or document "1024M effective on this host" with the host's swap stated.
4. **Throttling zeros are lies.** Emitting `cgroup_cpu_throttled_usec 0` when the cgroup does not exist, instead of emitting nothing (R1 fields are only meaningful when the controller is enabled, `pkg/api/metrics.go:15-16`), repeats the `pkg/api/metrics.go:46-53` static-zero failure. Absent ≠ zero, at the series level.
5. **Hierarchical vs local confusion.** `cpu.stat` throttling is non-hierarchical (R1); `cpu.stat.local` includes inherited throttling (R1). Exporting one while labelling it as the other double-counts or under-counts by the ancestor's share. The README must name which file each series reads.
6. **OOM is not an error to hide.** At `memory.max` the kernel OOM-kills (R1). A tool that treats exit 137 as "container crashed, retry" without surfacing `memory.events.max`/`oom_kill` (R1) is hiding the limit doing its job. The metric must exist *before* the first OOM, not after the first post-mortem.
7. **Root + no capability is the worst combo.** Shipping `scratch` root (no `USER`) without `NET_ADMIN` (gap §5.3/§5.6) means the image is privileged where it should not be and unprivileged where it must be. Either the capability is declared and the user is non-root, or the README states plainly "shaping runs host-side via systemd, container is observation-only" — and S4's rollback line is updated to match.
8. **Rollback that has never been drilled is a slogan.** If `docs/CHECKPOINTS.md` has no pasted digest-before/digest-after + health-JSON evidence from an actual rollback, S4 claim 2 remains prose. R7 gives the mechanism (pull by digest); only a drill proves it works on *this* registry, *this* compose version, *this* host.

---
*Teams checklist: every quantitative claim traces to `file:line` or R1–R10. Claims that could not be traced to a fetched primary are marked `unverified — primary source not found` inline (§3.1 static linking, §4.2 exit 137, §4.4 compose-v2 deploy application, §5.9 Swarm `restart`, §5.10 secret files, §6.1 namespace creation, §6.2 BuildKit multi-arch + in-container `tc`, §6.2 container `NET_ADMIN` reach). Slices 1–6 are proposals in this research file only — no `deploy/` file was modified in this pass.*
