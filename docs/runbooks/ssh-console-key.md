# CGO-DEV — SSH console key install (operator console step)

**Scope:** unblocks the eBPF H1–H11 probes + Linux live acceptance from
PC1 (`DESKTOP-B6K9UCM`) — only a human with VirtualBox console access
on the host can do this. No autonomous step can install a key into
a VM console.

**Host:** PC2 `nemo` (`tailscale` `100.72.251.32`, owner `Airis`)
**VM:** `CGO-DEV-FRESH` (1 vCPU / 3072 MB, NAT `2222→22` + `9090→9090`,
SATa `D:\VMs\CGO-DEV\CGO-DEV.vdi` reused — no new VDI, no ISO download)
**Guest (recorded):** Ubuntu 24.04 `6.8.0-147-generic`, x86_64
**Key:** `nemo_lab` — private key stays `~/.ssh/nemo_lab` on PC1;
the **public** key is the binding artifact below (also at
`cgo/credentials/nemo_lab.pub`, git-kept, never a private key).

---

## 0. Pre-call — confirm the VM is reachable (still BLOCKED here)

From PC1 now (2026-11-09): `tailscale status` shows
`nemo … offline` and `tailscale ping` times out. That has to change
first: PC2 **must be powered on**, else every step below is moot.

```powershell
# Power PC2 (unless it is already on), then on PC2:
tailscale status
ssh Airis@100.72.251.32                        # host reachable
VBoxManage startvm "CGO-DEV-FRESH" --type headless
```

If `PC2` is off, that is the first unblock condition and it is also
human-only.

Once the VM is up, keep it headless until the key is installed;
switch to the GUI console for step 2 below.

---

## 1. Open the VM console on PC2

```powershell
VBoxManage list vms
VBoxManage startvm "CGO-DEV-FRESH"  # or click "Start" in the GUI
```

Inside VirtualBox: click the **running** CGO-DEV-FRESH window
(non-headless), log in at the console as the guest user.

## 2. Install the key (guest console, paste verbatim)

```bash
mkdir -p ~/.ssh && echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDN8D1Ddl53RB1H590KPHlLmq7cS3plVYMNWtFRImBFT opencode-pc1-to-nemo' >> ~/.ssh/authorized_keys
chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys
```

Notes:

- The line above is the **full** `ssh-ed25519` text from
  `cgo/credentials/nemo_lab.pub` (line-wrap for display only).
- Do not run this from this Windows shell again: SSH is PC1→VM only; this
  section runs **inside the VM console**.
- Guest user must already exist in the VM's GUI login; the command
  above does not create it.

## 3. Verify from PC1 (this host)

```bash
ssh -i ~/.ssh/nemo_lab -p 2222 -o BatchMode=yes -o StrictHostKeyChecking=accept-new <guest-user>@100.72.251.32 \
  "uname -r; uname -m; whoami"
```

Full `ssh-ed25519` line of both keys in `cgo/credentials/` (`*.pub` only;
private keys live in `~/.ssh/` and are **never committed** — the tree
is hygiene-clean (1024 files).

### Success = comes back cleanly

Expected (probe H1, docs/runbooks/bpf-observer-gates.md §1):

```text
6.8.0-147-generic
x86_64
<guest-user>
```

If you see `Permission denied (publickey,password)`, the install
world is red; re-do step 2. Do not special-case passwords — none are
recorded.

## 4. Enabler for the rest

Once step 3 prints, in the same SSH session:

```bash
ls -l /sys/kernel/btf/vmlinux            # H2
bpftool btf dump file /sys/kernel/btf/vmlinux format c > /tmp/vmlinux.h && head -5 /tmp/vmlinux.h   # H3
clang-18 --version; clang-18 -target bpf -E - </dev/null; echo $?  # H5
ls /usr/include/bpf/bpf_helpers.h        # H6
make bpf && make bpf-check               # H9 (repo root, Linux 6.8)
readelf -S pkg/probe/bpf/bpf_probe.o | grep -E 'raw_tp|fentry|kprobe'  # H10
```

record the exact outputs in `docs/CHECKPOINTS.md` and move the BLOCKED
rows to PENDING only after H1–H11 pass.

## 5. Guest user is TBD — record it once known

`guest_user` lives in `cgo/credentials/vm-connection.json`; fill the
field after the console login succeeds and note it in
`docs/CHECKPOINTS.md` with the date. Until then every H1 probe is
blocked on it.
