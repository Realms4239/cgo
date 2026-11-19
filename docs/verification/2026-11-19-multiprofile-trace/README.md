# 2026-11-19 — guest-side observatory capture, multi-profile attempt

The capture produced a real P2 population of 151 samples and a sealed P2 `.dist`.
It did **not** produce a P3 population and only 15 sub-100 ms samples, so the
goal of one profile-matched sample set per profile was not reached.

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

The configured P3 and P4 delays did not appear in the observed srtt values.
The capture therefore cannot calibrate P1, P3 or P4. This is a named gap, not
a claim that those profiles were observed.

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
of the qcow2 by 5 GB was approved and is documented in `C:\WEBSERVER`
(`PROPOSED-2026-11-19-bench-disk-expand.md` and
`APPROVED-2026-11-19-bench-disk-expand.md`) and executed after this capture.

## Proves / does not prove

Proves: an eBPF observatory capture of shaped bench traffic contains 151 real
P2-band samples and calibrates a sealed `P2.dist`.

Does not prove: organic passive field evidence; P1, P3 or P4 calibration; a
`TC_LIB_DIR` replay of this new `P2.dist`; any production or thesis claim. No
VERIFY row is promoted by this builder pass.
