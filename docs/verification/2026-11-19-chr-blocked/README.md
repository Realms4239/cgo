# 2026-11-19 — RouterOS CHR boot: BLOCKED (root cause confirmed)

## Status — RESOLVED later the same day

This block is superseded: the pinned pristine image was found locally and one bounded
`cgo-chr-v13` cycle sealed 11/11 checks. See `../2026-11-19-chr-verified/README.md`.
The root-cause analysis below is retained as historical evidence.

### Original blocked status

BLOCKED. Named unblock condition: a pristine `chr-7.24.6.img` with SHA-256
`c6abf87622c5bdbd86f7b5730c8e83f86f2335ee2fbe97d008da176ffbf97878` (the pin in
`kit/chr/chr-image.sha256`) deposited at
`/home/sadc/labo-cgo/chr/chr-7.24.6.img`, then one bounded `cgo-chr-v13` run.

## Root cause (systematic debugging, no fix attempted after confirmation)

- Symptom: `./cgo-chr-v13 … --timeout 180s` and `--timeout 360s` both returned
  `chr: chr: délai dépassé` (exit 1), with the console warning
  `CPUID.80000001H:ECX.svm [bit 2]` (cosmetic on this Intel host).
- Evidence at the serial boundary: a minimal diagnostic that spawns the same
  QEMU line with `-serial stdio`, sends exactly the driver's login sequence, and
  captures the console returned:

  ```
  CHR Login: admin+t500w
  Password:
  Login failed, incorrect username or password
  CHR Login:
  ```

- Interpretation: `cmd/cgo-chr/serial.go:119-121` sends an empty password,
  assuming a pristine first boot. This image was already booted on 2026-11-16;
  that first boot set a throwaway admin password (runbook:
  `docs/verification/2026-11-16-chr-boot/README.md`). The driver's
  `loginAfterRe` never matches, so `expect()` burns the full timeout.
- Image state: host SHA-256 `c2dd927938a29ad73c1e06597a1cfab9d733cc740e8664850c413ce371048f7c`
  (R0 evidence `20261119-120106-read-window.txt`), pin
  `c6abf87622c5bdbd86f7b5730c8e83f86f2335ee2fbe97d008da176ffbf97878`.
- No pristine copy was found under `/home/sadc` (R0 evidence
  `20261119-120106-read-window.txt`). Download from
  `https://download.mikrotik.com/routeros/7.24.6/chr-7.24.6.img.zip` failed from
  this workstation (`curl` connect failure) and is refused by the guarded host
  write window.
- Two timeout-doubling attempts were exhausted before the two-strike halt; the
  third investigation step was evidence gathering at the serial boundary, not
  another fix. No password guessing, no auth disabling, no gate change.

## Evidence paths

- `C:\WEBSERVER\evidence\20261119-113740-write-window.txt` (attempt 1)
- `C:\WEBSERVER\evidence\20261119-114040-write-window.txt` (attempt 2)
- `C:\WEBSERVER\evidence\20261119-120106-read-window.txt` (image hash + search)
- `C:\WEBSERVER\evidence\20261119-120538-read-window.txt` (prior good transcript)
- Host diagnostic: `/home/sadc/labo-cgo/chr/chr-diag-login.py` output above; the
  QEMU warning is from `diag-qemu-20261119.log`.

## What was and was not proven

Proven: the CHR guest boots to `CHR Login:` and rejects the empty password;
the image is mutated relative to the pin; no pristine copy or reachable download
exists. Not proven: any RouterOS import/export claim for 2026-11-19. The
2026-11-16 proof remains its own dated artifact and is not extended here.
