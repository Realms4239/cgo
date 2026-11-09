# CGO-DEV-FRESH — SSH key install (console)

**Owner:** operator with VirtualBox console access on PC2 nemo (`100.72.251.32`).
**Unblocks:** eBPF probes H1–H11 (`docs/runbooks/bpf-observer-gates.md` §1),
Linux 6.8 live acceptance, IFB ingress Task 9 — all BLOCKED until guest SSH
accepts a recorded key.

## 0. Reachability (checked 2026-11-09)

- `tailscale status` (this host): nemo `100.72.251.32` → **offline**.
- `tailscale ping` → timeout. Guest SSH at `100.72.251.32:2222` (NAT `2222→22`)
  unreachable while the host is offline.
- Bring the host back first: power on PC2 nemo, start `CGO-DEV-FRESH`
  (`VBoxManage startvm "CGO-DEV-FRESH" --type headless` on the host), verify
  guest `6.8.0-147-generic` and `openssh-server` enabled.

## 1. Console install (temporary GUI console on CGO-DEV-FRESH)

1. Open the VM in VirtualBox GUI → Console → log in as the guest user.
2. Paste (replace `<nemo_lab.pub>` with the **full** line from
   `cgo/credentials/nemo_lab.pub` — currently at
   `/c/Users/ASUS/.ssh/nemo_lab.pub` on this operator host; this doc stores
   only the public key):

   ```sh
   mkdir -p ~/.ssh && echo '<nemo_lab.pub full line>' >> ~/.ssh/authorized_keys
   chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys
   ```

3. Idempotent alternative (key already present once — does not duplicate):
   `grep -qxF '<nemo_lab.pub>' ~/.ssh/authorized_keys || echo '<nemo_lab.pub>' >> ~/.ssh/authorized_keys`

## 2. Verify from this host

```sh
ssh -i ~/.ssh/nemo_lab -p 2222 -o BatchMode=yes -o StrictHostKeyChecking=accept-new <guest-user>@100.72.251.32 \
  "uname -r; uname -m; whoami"
# expect: 6.8.0-147-generic / x86_64 / <guest-user>  (probe H1)
```

If the guest user is not `nemo`, use the actual console account name;
record it here: `guest_user = ________________`.

## 3. Record

- Append a `docs/CHECKPOINTS.md` entry (date, host, `uname -r`, probe outputs).
- Move eBPF BLOCKED row to PENDING only after H1–H11 pass; never claim
  live attach without a real capture.

## Never

- No private keys in repo docs — public only (`cgo/credentials/*.pub`,
  private stay in `~/.ssh/`).
- No `cloud-init` guest breakage: this file is console-only recovery, not a
  host-modification recipe.
