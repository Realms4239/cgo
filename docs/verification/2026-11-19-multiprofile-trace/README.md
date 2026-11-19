# 2026-11-19 — guest-side observatory capture, multi-profile attempt

The captures produced a profile-matched population for all four profiles and
sealed a `.dist` for each: P1 40 samples, P2 151, P3 41, P4 42. Every sidecar
hash and every evidence hash verifies. The result is an eBPF observatory
capture of **bench-generated shaped traffic**, not an organic passive field
trace. Claims built on it must say that.

This is an eBPF observatory capture of **bench-generated shaped traffic** on the
guest, not an organic passive field trace. The flows are real kernel samples;
the traffic was injected by the bench. Claims built on it must say that.

## Capture

The same `cgo-observe-x` binary used on the host was cross-built from
`cmd/cgo-observe` with `-tags ebpf_object`, uploaded to the bench guest and run
there, because the host observatory could not see traffic inside the guest's
network namespaces.

```
sudo -n /home/bench/cgo-observe-x --out /home/bench/profile-trace2 \
  --duration 900 --interval 5
```

Clean stop: 7,513 rows, 0 excluded, object
`8612f5785a46390cfead383f5e5d2bcde1e33777dda10331221169cae603d17f`.
The full `flows.csv` sha256 `ee8ce4185ea335d219fc1714e666d3da7afd17b1d880092e491e86baed2ee2e8`
matches the manifest declaration; it stays on the guest at
`/home/bench/profile-trace2/flows.csv` because the repository hygiene gate caps
a single file at 400 KB and the full trace is 799 KB. The committed
`evidence/flows-extract.csv` holds the header plus all 167 flow rows, 15,162
bytes, sha256 `58deb9a2…`; `evidence/manifest.json` sha256 `6a77aa98…`. Both
committed files verify with `sha256sum -c evidence/SHA256SUMS`.

Four audits were run while the capture was live, one per configured profile,
all exit 0: P1 20 ms, P2 100 ms, P3 600 ms, P4 40 ms. Each audit's netem delay
was changed on `w-edge` and `c-wan` in `ns_wan_emulator`.

## Observed srtt bands

Counts over the 167 flow rows of the sealed trace:

| Band | Samples | Result |
|---|---|---|
| below 100 ms | 15 | below the P1 calibration floor of 100 ms |
| 100 to 199 ms | **151** | P2 population found |
| 200 to 599 ms | 1 | no defined profile |
| 600 ms and above | 0 | no P3 population |

The configured P3 and P4 delays did not appear in that capture because the eBPF
tracepoints see TCP and the audits generated little load. A second, focused
capture drove TCP directly with `curl` loops at each configured delay, with the
observatory attached throughout (`profile-trace3`, 4,004 rows). It produced
41 samples at 603.026 ms (P3) and 42 at 43.025 ms (P4). Those were written as
baseline telemetry and calibrated with exit 0:

| Profile | Samples | p50 ms | dist sha256 |
|---|---|---|---|
| P3 | 41 | 603.026 | `cf1712ebdecb30c73294476cb48259da8939247b20a584aa901cb28f68f64a65` |
| P4 | 42 | 43.025 | `619c7a881af8526221633306c1f9a4b1f49290dfc4e1e67892f1dfdf87054ed8` |

Evidence: `evidence/flows-p3p4-extract.csv` (83 flow rows, 7,109 bytes, sha256
`421a1f88…`), `evidence/manifest-p3p4.json` (sha256 `5f8f6d4d…`) and
`evidence/SHA256SUMS-p3p4`, all verified with `sha256sum -c`. The full
4,004-row trace stays on the guest at `/home/bench/profile-trace3/flows.csv`.

A third, P1-only capture used the same TCP-driving method at 10 ms per hop,
with the observatory attached for 200 s (`profile-trace4`, 1,993 rows). It
produced 40 samples at 23.028 ms, above the 20 ms floor, and calibrated:

| Profile | Samples | p50 ms | Floor ms | dist sha256 |
|---|---|---|---|---|
| P1 | 40 | 23.028 | 20 | `fb99def6c36bb00228eccb09ff35ca3d25df05675bae83a469f67b5a91bb79d3` |
| P2 | 151 | 175.282 | 100 | `f9c23c39791f44d45de72befd652f0f722f40c9edb57d37c082b04d0995f5aa2` |
| P3 | 41 | 603.026 | 600 | `cf1712ebdecb30c73294476cb48259da8939247b20a584aa901cb28f68f64a65` |
| P4 | 42 | 43.025 | 40 | `619c7a881af8526221633306c1f9a4b1f49290dfc4e1e67892f1dfdf87054ed8` |

Evidence: `evidence/flows-p1-extract.csv` (40 flow rows, 3,505 bytes, sha256
`4b086ff3…`), `evidence/manifest-p1.json` (sha256 `981ec0b1…`) and
`evidence/SHA256SUMS-p1`. The full 1,993-row trace stays on the guest at
`/home/bench/profile-trace4/flows.csv`. All four `dist_sha256` and
`source_telemetry_sha256` values were recomputed and match their sidecars.

All four profile floors are met, so the goal of one profile-matched sample set
per profile is reached. The honest limitation is the source: shaped bench
traffic observed by eBPF, not passive field traffic.

## P2 calibration

The 151 samples in the 100 to 199 ms band were written as `baseline` rows with
5 s offsets under `audits/field-profile-trace-p2-20261119/telemetry.csv` and
calibrated:

```
go run ./cmd/cgo calibrate --audit field-profile-trace-p2-20261119 --profile P2 \
  --data docs/verification/2026-11-19-multiprofile-trace \
  --out docs/verification/2026-11-19-multiprofile-trace/calibrations
```

Exit 0. `P2.dist` sha256 `f9c23c39791f44d45de72befd652f0f722f40c9edb57d37c082b04d0995f5aa2`,
`source_telemetry_sha256` `0dfa0809…` (both recomputed and matched), 151 samples,
p50 175.282 ms, p95 175.282 ms, scale 18.821 ms, 0 below floor, 0 clipped.

This is separate from the bench-audit-derived `P2.dist` in
`../2026-11-19-field-tune/` and from the field-trace-derived `P3.dist` in
`../2026-11-19-field-trace-p3/`.

## Disk work

The guest disk reached 99 percent during the first capture attempt. Apt,
fwupd and man caches were cleared, returning 269 MB free. A scoped expansion
of the qcow2 by 5 GB was approved in `C:\WEBSERVER`
(`PROPOSED-2026-11-19-bench-disk-expand.md` and
`APPROVED-2026-11-19-bench-disk-expand.md`) and executed after this capture:
rehearsal resize on a copy verified 8.5 GiB virtual size, the live overlay was
backed up to `labo-cgo-bench.qcow2.pre-expand-20261119`, the VM was stopped with
`kill -TERM 3673912`, the live overlay was resized to 8.5 GiB, and the VM was
relaunched with its exact recorded command. Cloud-init grew the partition on
boot, so the guest now reports 7.2 GB available at 29 percent use. Evidence is
in `C:\WEBSERVER\evidence\20261119-200907-*` through `20261119-201154-*`.

## Proves / does not prove

Proves: eBPF observatory captures of shaped bench traffic contain real,
profile-matched populations for P1, P2, P3 and P4, each above the engine's
floor, and calibrate four sealed `.dist` files with hashes that verify.

Does not prove: organic passive field evidence; a `TC_LIB_DIR` replay of these
four new `.dist` files; any production or thesis claim. No VERIFY row is
promoted by this builder pass.
