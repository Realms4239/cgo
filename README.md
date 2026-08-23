# CGO — LIEN Instrument

**Protecting critical operational traffic on constrained access links.**
Non-intrusive QoS audit + reproducible AQM/BBR evaluation for the Meteo
Madagascar context. Self-contained Go binary with an embedded React dashboard.

Thesis source of truth: `LIEN.md`. Scope contract: `docs/SPEC.md`.
Build plan: `docs/PLAN.md`. Visual authority: `DESIGN.md`.

## What it does

1. **Audit** accessible links from a client workstation (ping p50/p95,
   small-object p95, bulk goodput) → `data/link_audit.csv` (Tableau 6).
2. **Evaluate** qdiscs `pfifo_fast / fq_codel / CAKE` × CC `CUBIC / BBR`
   on profiles P1/P2 × 3 reps = 36 events (18 reduced), each event
   baseline 30 s → charge 120 s → récupération 30 s, gated G0–G7
   → `data/runs/<run>/aqm_eval.csv` + SHA-256 `manifest.json`.
3. **Serve** four panels over SSE @10 Hz: Campagne · Temps réel ·
   Résultats · Intégrité (LIEN Partie III.IV).

## Quick start

```bash
# host (dev)
make build            # frontend dist + go binary → bin/cgo
./bin/cgo --serve     # http://localhost:9090

# VM (real tc/netem/BBR) — see RUNBOOK.md
bash deploy/engine.sh --action ensure   # scan+boot VM if needed
bash deploy/engine.sh --action deploy   # cross-compile, push, install, health-check

# CLI (LIEN Partie III.III)
./bin/cgo audit --link-type 5g --site "Dept X" --duration 300
./bin/cgo verify      # manifest SHA-256 check
./bin/cgo figures     # regenerate SVGs from frozen CSVs
```

## Layout

```
cmd/cgo/           CLI entry (--serve | audit | verify | figures)
pkg/model/         LIEN domain types, CSV schemas (Tableaux 6–7)
pkg/qdisc/         tc arg builders behind TCRunner (fake for host tests)
pkg/probe/         ping · small-object HTTP · bulk TCP
pkg/metrics/       percentiles, deadline %, cost Ar/h — pure functions
pkg/campagne/      event state machine, gates G0–G7, matrix runner, freeze
pkg/results/       medians/IQR aggregation over frozen CSVs
pkg/figures/       static SVG generation (no node toolchain needed)
pkg/api/           REST + SSE hub (10 Hz, delta-structural, Last-Event-ID)
web/frontend/      React/Vite/TS — four LIEN panels, embedded via go:embed
deploy/            VMware scan/boot/ensure/deploy engine + VM installer
docs/              SPEC · PLAN
```

## Security model

No auth tokens by design: the instrument is VM-local. Every mutating action
(start run/stop/audit/import) is guarded client-side by an arm-then-fire
double confirmation (`DÉMARRER` → `CONFIRMER ?`, 5 s timeout).

## Verification gates

```bash
go vet ./... && go test ./...
cd web/frontend && bun run typecheck && bun run build
# on the VM:
ssh altfloat@192.168.174.128 'cd ~/cgo && ./cgo-linux verify'
npx playwright test    # against http://192.168.174.128:9090
```

Nothing is fabricated: empty panels say so; quarantined events stay counted;
figures regenerate only from frozen CSVs.
