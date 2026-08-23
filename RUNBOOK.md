# RUNBOOK — CGO on the local VM

Machine-local config: `deploy/cgo-vm.yaml` (gitignored). Copy from
`deploy/cgo-vm.yaml.example` and fill ssh/vmx values.

## 1. One-time VM bootstrap

The Ubuntu VM (24.04, user with passwordless sudo) needs once:

```bash
sudo apt update && sudo apt install -y iproute2 curl bc
sudo modprobe tcp_bbr || true
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr   # if testing BBR cells
```

## 2. Deploy pipeline (host → VM)

```bash
bash deploy/engine.sh --action scan    # find .vmx anywhere on this PC (--deep for full sweep)
bash deploy/engine.sh --action ensure  # boot VM via vmrun if SSH down, wait for SSH
bash deploy/engine.sh --action deploy  # build frontend+binary, push cgo-linux.new, install, health-check
bash deploy/engine.sh --action status  # SSH + process + dashboard health
bash deploy/engine.sh --action logs    # tail /tmp/cgo.log on the VM
```

Deploy is binary-first and idempotent: the SPA is embedded in the binary;
no Go toolchain required on the VM.

## 3. Running a campaign

UI: Campagne panel → select P1/P2 + reps → `DÉMARRER` → `CONFIRMER ?`.
Or API:

```bash
curl -X POST http://192.168.174.128:9090/api/run/start \
     -H 'Content-Type: application/json' \
     -d '{"profiles":["P2"],"reps":3}'        # reduced matrix = 18 events
```

Each event runs baseline 30 s → charge 120 s → récupération 30 s against the
testbed path; rows freeze into `data/runs/<run_id>/aqm_eval.csv`; stopping is
graceful — restarting the same run resumes past completed event IDs (G5).

> Testbed prerequisite (M3+): the veth pair (`veth-c` latency hop with netem,
> `veth-s` shaping/AQM hop) must exist before real campaigns. Without it,
> gates G0/G1 fail honestly and events quarantine as `invalid`.

## 4. Audit (RQ1)

```bash
./cgo-linux audit --link-type 5g --site "Dept A" --provider Yas --duration 300 --target 8.8.8.8
```

Appends one Tableau-6 row to `data/link_audit.csv`. UI equivalent lives in
the Campagne panel under *Audit lien accessible*.

## 5. Results, integrity, figures

```bash
curl 'http://192.168.174.128:9090/api/results'                 # medians/IQR per cell, best ★
curl 'http://192.168.174.128:9090/api/report/export?format=md' # short report
curl -X POST http://192.168.174.128:9090/api/figures/regen     # SVGs from frozen CSVs
ssh altfloat@VM 'cd ~/cgo && ./cgo-linux verify'               # manifest SHA-256 check
```

## 6. Verification checklist (every milestone close)

```text
host:   go vet ./... && go test ./...
        cd web/frontend && bun run typecheck && bun run build && npx playwright test
VM:     deploy engine status → dashboard healthy
        ./cgo-linux verify → manifests ok
```
