# Host acceptance — slices 4, 5, 6, 7

*Written 2026-10-30. One Linux 6.8 host, one hour, four slices in order.*
*Every step states what it proves and what it must not be taken to prove.*

## 0. Before anything — readiness, read-only

```bash
git log --oneline -1
bash kit/host-readiness.sh --require-linux
```

Expected: a table, `résumé : hôte prêt pour l'acceptation`, exit 0.
If exit 2, the host is not Linux. If exit 1, the named `MISSING` lines are the
blockers; stop and fix them before touching a namespace.

Never run two slices' acceptance at the same time: each mutates namespaces and
qdiscs on the same kernel.

## 1. Build the eBPF object (once, on this host)

```bash
make bpf && make bpf-check
sha256sum pkg/probe/bpf/bpf_probe.o
go build -tags ebpf_object ./...
```

Expected: the two `make` targets exit 0; `bpf-check` prints nothing; the hash
matches `pkg/probe/bpf/bpf_probe.o.sha256`; the tagged build exits 0.

**Never commit the object until the live attach below has run.** The object is
a build artifact of this kernel tree, not source.

## 2. Slice 4 — eBPF ground truth

```bash
sudo -E go run -tags ebpf_object ./cmd/cgo --serve --addr 127.0.0.1:9091 &
curl -s localhost:9091/api/health | jq .telemetry
# one audit with the probe, one without
sudo -E go run -tags ebpf_object ./cmd/cgo audit --site lab --link-type fiber --duration 20
go run ./cmd/cgo audit --site lab --link-type fiber --duration 20
jq '.kernel_telemetry' data/audits/<audit_id>/audit_report.json
```

Expected with the probe: `{"source":"ebpf", ...}` and a report whose
`kernel_telemetry.available` is `true` with a non-zero `object_sha256`.
Expected without: `"source":"tcpinfo"` and a `reason` that names the cause.

**Do not conclude**: that the retransmitted bytes are exact for a production
link, or that the qdisc drops match CAKE's own counters — only that the
attachment and the counters moved on this host.

## 3. Slice 6 — IFB ingress acceptance

```bash
sudo bash kit/run-dualns.sh   # bench + sinks inside ns_client
sudo -E CGO_SHAPE_MODE=ingress go run ./cmd/cgo run --bench dualns --direction down
```

Expected: ingress preflight passes, `ifb0` fills, CAKE drops under load, and
the report carries `linux_tc_ingress` / `routeros_ingress`. Record the drop
counters and a before/after throughput pair.

**Do not conclude**: that the operator's router will accept the equivalent
queue tree — that is the CHR proof (`kit/chr/README.md`), a different
evidence.

## 4. Slice 5 — trace-driven `.dist`

Prerequisite, human: a **real field trace of ≥ 30 baseline samples**. Without
it, this slice stays Windows-green and this step is not run.

```bash
go run ./cmd/cgo calibrate --audit <audit_id> --profile P2
CGO_DELAY_MODEL=dist go run ./cmd/cgo run --bench dualns --profiles P2
```

Expected: `data/calibrations/P2.dist` + its sealed `P2.json` sidecar with the
dist bytes hashed; the campaign loads it and the report says `delay_model: dist`.

**Never** cite a `SMOKE-TEST` `.dist` as a field-calibrated model.

## 5. Slice 7 — BBRv3 matrix

```bash
bash kit/install-bbr3.sh --help        # read what it will do before running
sudo bash kit/install-bbr3.sh --yes    # builds and installs; leaves the reboot
sudo reboot                            # manual, then reconnect
cat /proc/sys/net/ipv4/tcp_available_congestion_control
go run ./cmd/cgo run --bench dualns --profiles P2 --cc bbr3
```

Expected: `bbr3` appears in the available list; the run is accepted; the
matrix seals three repetitions per cell.

**Do not conclude**: `bbr` still means BBRv1 everywhere in the corpus; do not
rewrite a frozen label.

## 6. Record

For each slice that ran: command, date, host, artifact path, hash, and the
sentence **« ce que cela ne prouve pas »**. Open the `docs/VERIFY.md` row only
with those four; a slice whose step was skipped stays `BLOCKED` with the
missing condition named.

## 7. Never

- No production write, load, root, or capture.
- No frozen corpus, CSV schema, or `data/runs/` change.
- No claim of "verified" from a Windows-green result.
- No merge to `main` — the human decides.
