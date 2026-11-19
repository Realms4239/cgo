# 2026-11-19 — RouterOS CHR boot VERIFIED (pinned pristine image)

Operator request: "there's a router OS VM inside the host isn't it a good deal, we
really need to fix it no matter what". Authorization: `APPROVED-2026-11-19-chr-boot-r2.md`
(R2, reversible, scope `/home/sadc/labo-cgo/chr`, expiry 23:59 UTC).

## Resolution of the earlier BLOCKED record

`../2026-11-19-chr-blocked/README.md` recorded the root cause: the first boot on
2026-11-16 set a throwaway admin password on the raw image, so the driver's empty
password was rejected. That diagnosis was correct, but the blocked state is now
resolved: a **pristine image with the exact pinned hash was found on this workstation**
at `D:/chr-download/chr-7.24.6.img` (the same file the operator used on 2026-11-16),
verified locally as `c6abf87622c5bdbd86f7b5730c8e83f86f2335ee2fbe97d008da176ffbf97878`.

No pin change, no password guessing, no auth disabling. Attempts to reconstruct the
throwaway from recorded timestamps failed; the pristine image made them unnecessary.

## What was run

Guarded R2 window (`C:\WEBSERVER\evidence\20261119-151053-write-window.txt`):
re-deposit the pinned image, then verify on-host hash `c6abf876…` (exact pin).

Guarded R2 window (`C:\WEBSERVER\evidence\20261119-151115-write-window.txt`): one
bounded cycle:

```
nohup /home/sadc/labo-cgo/chr/cgo-chr-v13 \
  --report /home/sadc/labo-cgo/chr/audit-1794898684-report.json \
  --image /home/sadc/labo-cgo/chr/chr-7.24.6.img \
  --level linux-namespace --qemu /usr/bin/qemu-system-x86_64 \
  --timeout 300s --out /home/sadc/labo-cgo/chr/proof-2026-11-19-final
```

A first attempt used a quoted local path that the SFTP transport treated literally,
so the upload failed and the driver ran against the mutated image (bounded failure
`chr: délai dépassé`). No artifact from that attempt is used. The image was then
re-deposited correctly and the run above sealed on the first bounded cycle.

## Result

`chr: preuve …/proof-2026-11-19-final/proof.json · 11/11 vérifications · niveau linux-namespace`

| artifact | sha256 |
|---|---|
| `proof.json` | `b2d3395d3e706153326921370f425da09839076f8f814e34f2ac561099e3d754` |
| `export.txt` | `71a3c9df6637b909aca02d11cdb3151dce1c8a7ef7e0c3a2ce2c9ddc591ff700` |
| `transcript.txt` | `70269d129f67b6bbfdd854ec2fc9faa0f816c02577753b9048d98b5a75ad143b` |

- `import.accepted: true`, 11 pass / 0 fail in `checks` (counted independently from
  the readback).
- `chr.image_sha256` in the proof is `c6abf876…` — the pinned pristine value captured
  **before** boot, as the driver requires. The post-run on-host hash differs, which is
  the expected first-boot mutation and is not an evidence defect.
- The embedded `proof.json` extracted from the R0 readback hashes to exactly
  `b2d3395d…`, matching the on-host declaration.

Evidence: `evidence/chr-proof-readback-20261119.txt` (guarded R0 readback,
`sha256 7e85ddce…` with sidecar) and `evidence/proof.json` (extracted, hash verified).
The QEMU `CPUID.80000001H:ECX.svm` warning is cosmetic on this Intel host and was
present in the 2026-11-16 proof too.

## Proves / does not prove

Proves: RouterOS 7.24.6 CHR accepts the remediation import and its `export` matches
the expected structure, on 2026-11-19, from the pinned pristine image, with 11/11
checks passing and hashes verifying independently.

Does not prove: any live link measurement (CHR runs in user-mode QEMU with no netns),
BBRv3, IFB ingress, thesis identity, or the v2.0.0 freeze. No merge, tag, or ledger
change is made by this slice.
