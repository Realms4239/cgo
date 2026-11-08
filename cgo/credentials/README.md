# CGO-DEV credentials (local-only, never committed)

- Host: 100.72.251.32 (PC2 nemo, Airis via nemo_lab)
- VM: CGO-DEV (Ubuntu 6.8.0-147-generic), NAT 2222->22, 9090->9090
- Keys (public only in repo docs; private keys stay in ~/.ssh/ on operator host):
  - nemo_lab : /c/Users/ASUS/.ssh/nemo_lab.pub ssh-ed25519 AAAAC3... opencode-pc1-to-nemo
  - id_ed25519 : /c/Users/ASUS/.ssh/id_ed25519.pub nemo@DESKTOP-B6K9UCM
- VM guest user: TBD (console install required per `docs/usage.fr.md`: `mkdir -p ~/.ssh && echo '<pubkey>' >> ~/.ssh/authorized_keys`)
- Fresh VM spec: 1 vCPU, 3072 MB RAM, Ubuntu (64-bit), SATA, NAT forwarding ssh 2222->22, web 9090->9090, dynamic VDI on D: (minimal, reuse CGO3-clone.vdi base; no Ubuntu ISO download if possible)

This folder contains no private keys. Add `cgo/credentials/` to .gitignore before any commit.
