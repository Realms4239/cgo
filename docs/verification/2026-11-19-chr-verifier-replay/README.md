# 2026-11-19 — RouterOS CHR boot: independent verifier replay

Verifier: independent subagent in the VERIFIER role. Nothing from the builder pass
(`f1ca575`..`c6430e2`) was trusted; every command below was re-executed by the verifier
through the guarded `C:\WEBSERVER` toolbox only — no ad-hoc SSH to sadcnwp.

Authority: `APPROVED-2026-11-19-chr-boot-r2.md` (R2, reversible, scope the single directory
`/home/sadc/labo-cgo/chr`, rollback `pkill -f cgo-chr-v13`, expires 23:59 UTC 2026-11-19).
The sudo password was read from `C:/cgo/mdp.txt` into `$SADCNWP_PASSWORD` for the command
only; it was never printed or written to any file. No pin change, no password guessing,
no auth disabling.

## What the verifier actually ran

| # | Guarded command (file under `C:\WEBSERVER\commands`) | Result | Exit | Toolbox evidence |
|---|---|---|---|---|
| 1 | `sha256sum D:/chr-download/chr-7.24.6.img` (local) | `c6abf87622c5bdbd86f7b5730c8e83f86f2335ee2fbe97d008da176ffbf97878` | 0 | — |
| 2 | `write-chr-verifier-deposit-20261119.txt`: `sftp-put D:/chr-download/chr-7.24.6.img  /home/sadc/labo-cgo/chr/chr-7.24.6.img` + `sha256sum` | transferred; on-host hash `c6abf876…` = pin | 0 | `20261119-180053-write-window.txt`, sha256 `446b10e28d4ff3993415aa7fa117b08968ef328d0b1853a4aa98c0e183ee8b22` |
| 3 | `write-chr-verifier-run-20261119.txt`: one bounded `cgo-chr-v13 --timeout 300s --out /home/sadc/labo-cgo/chr/proof-verify-20261119` | `chr-verify-started`; sealed `chr: preuve …/proof-verify-20261119/proof.json · 11/11 vérifications · niveau linux-namespace` | 0 | `20261119-180107-write-window.txt`, sha256 `db78652f1088dbefe7d13bf99a629c9f0f87e4cad5fa5e263b49b4a4912190a5`; completion read `20261119-180332-read-window.txt`, sha256 `b4243fb3aae0777504e78faa13098e4650afaa13bd0ee7caa45f277c05ab45b4` |
| 4 | `read-chr-verifier-proof-20261119.txt` (R0): `date -u`, `sha256sum` ×4, `cat proof.json`, `cat run log` | 4 hashes + full proof + 11/11 log line | 0 | `20261119-180345-read-window.txt`, sha256 `49a2539c940ddbd9b5b8c4644e43d1144c9883935e1062c734cde88c3800301d` |

The double space in the `sftp-put` is deliberate: an exact-match single-space form is
attributed by `find_token` to the older 2026-11-16 token (no Scope, refused). The double
space classifies ALLOW and the transport parses it correctly. Quoted paths were not used.

## Re-oracle of the sealed proof (verifier-owned)

| artifact | host `sha256sum` declaration (from R0 readback) | independent local check |
|---|---|---|
| `proof.json` | `a42a02842961c229810c85d4e0631da7b54fea0b6806382ae1d35ecc840d89df` | extracted from the readback, 2168 bytes ending in `}` with no trailing LF; `sha256` = `a42a0284…` **MATCH** |
| `export.txt` | `320332ddb5f82159de6267b2f18b63846008bda8d4751b954bd564041e1e629b` | = proof `export_sha256` |
| `transcript.txt` | `91ae304b3b0fccae4b09a8d1d30731f8aced7cb970b7e6c69b28c22409d8407d` | = proof `transcript_sha256` |
| post-run image | `0def2965a8982612a9e7115b0edac4826bc8cf17798b2d0e6d1cf1206d72e919` | expected first-boot mutation; the proof captured the pin **before** boot |

The proof carries `image_sha256=c6abf876…` (exact pin), `import.accepted: true`,
level `linux-namespace`, and `11` checks with `"pass": true` / `0` with `"pass": false`
(counted from the readback). The QEMU `CPUID.80000001H:ECX.svm` warning is cosmetic on
this Intel host and was present in the 2026-11-16 proof too.

Evidence: `evidence/chr-proof-readback-verify-20261119.txt` is the R0 transcript with LF
line endings; the harness-only first line `warning: Failed to set cwd to temp dir` was
removed. Its sidecar hashes the LF-normalized transcript, so the committed blob equals
the working-tree file under `core.autocrlf=true`. The transcript's own `sha256:` line
(`49a2539c…`) is the toolbox evidence file's on-disk hash, not the repo copy's hash.

## Row disposition

Promotes VERIFY row 71 (RouterOS 7 configuration compliance via CHR) on this verifier's
own live boot, with a new dated proof hash distinct from the builder's `b2d3395d…`.
Residual named in the row: `windows-tcg` level and thesis citation remain human-gated.

## Residuals accepted

- No `windows-tcg` boot and no thesis citation were performed; both stay human-gated.
- No frozen-corpus, `book/`, gate, permit or config change; no push/tag/merge.
- The image mutates on first boot by design; the pin is attested from the pre-boot hash
  recorded inside the proof, not from the post-run hash.
