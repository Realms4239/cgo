# Meteolink

**A terminal-friendly, self-contained network benchmark for auditing constrained access links and comparing AQM/congestion-control policies — with provenance-backed evidence.**

Meteolink is a single Go binary (with an embedded React dashboard) built around a simple idea: audit a constrained access link from the client side, replay it on a reproducible bench, and compare queue-management and congestion-control policies using measurements you can verify — not vendor claims. Developed for the operational context of Météo Madagascar, where critical small traffic (telemetry, alerts, dashboards) competes with bulk transfers (satellite imagery, model downloads) on 4G/5G and fiber links prone to bufferbloat.

## Why Meteolink?

Most network "quality" conversations happen without evidence. Meteolink produces the evidence:

- **Client-side audit, non-intrusive.** Measure a production access link from a workstation without touching the infrastructure: ping p50/p95, small-object retrieval p95, bulk goodput — exported to `data/link_audit.csv`.
- **Reproducible AQM/CC evaluation.** A Linux bench (`tc`/`netem`) replays link profiles P1/P2 × qdiscs `pfifo_fast` / `fq_codel` / CAKE × congestion controls CUBIC / BBR across repetitions. Each event follows a fixed protocol: 30 s baseline → 120 s load → 30 s recovery, validated by gates G0–G7, then frozen to CSV with a SHA-256 manifest.
- **Verifiable provenance.** Every result can be re-checked: `cgo verify` validates the frozen CSVs against their manifest, and `cgo figures` regenerates SVGs only from frozen data. Nothing is fabricated; quarantined events stay counted.
- **Operator-oriented live view.** A four-view French dashboard (Campagne de mesure, Tableau live, Résultats + Comparaison, Provenance & archives) served over SSE at 10 Hz, plus a manual Façonnage du bord lever (qdisc, capacity, delay/jitter/loss), continuous surveillance, and an operator journal (ring buffer of 50 events).
- **One binary.** Frontend embedded via `go:embed`; no external services, no database, no runtime dependencies. Built and maintained as part of a master's thesis on reproducible benchmarking for constrained access links.

## Key Features

- Client-side link audit → `data/link_audit.csv` (p50/p95 latency, small-object p95, bulk goodput)
- Full AQM/CC test matrix: profiles P1/P2 × `pfifo_fast` / `fq_codel` / CAKE × CUBIC / BBR × repetitions
- Fixed event protocol (30 s / 120 s / 30 s) with validity gates G0–G7
- Frozen CSV archives + SHA-256 manifest for reproducibility
- Four-view French dashboard over SSE at 10 Hz (campagne, live, results/comparison, provenance)
- Façonnage du bord: manual shaping of qdisc, capacity, delay, jitter and loss
- Surveillance continue and operator journal (ring of 50, `GET /api/events`)
- Mode awareness: **Linux = full mode** (bench + tc control); **Windows = observation mode** (audit + dashboard only); override with `--mode observe|full|auto`
- Static SVG figure regeneration without a Node toolchain

## Non-goals

Meteolink is deliberately narrow. It does **not**:

- Send alerts or integrate with notification systems
- Manage a multi-site fleet or remote agents
- Provide long-term historical storage
- Offer authentication or multi-user access (the instrument is VM-local; mutating actions are guarded by an arm-then-fire double confirmation in the UI)
- Configure your routers or production network devices

It produces the *measured evidence* that informs those decisions — the audit report, the comparison matrix, the recommendations — but the decisions and their operationalization stay yours.

## Quick start

**Linux — full mode (bench with real `tc`/`netem`/BBR):**

```bash
make build            # frontend dist + Go binary → bin/cgo
./bin/cgo --serve     # dashboard at http://localhost:9090
./bin/cgo audit --link-type 5g --site "Dept X" --duration 300
./bin/cgo run         # execute the AQM/CC matrix
./bin/cgo verify      # SHA-256 check of frozen CSVs
```

**Deploy to an Ubuntu VM (the real bench):**

```bash
bash kit/engine.sh --action deploy   # cross-compile, push, install, health-check
```

**Windows — observation mode (audit + dashboard only):**

```powershell
cgo.exe --serve       # auto-detected observation mode
cgo.exe audit --link-type 5g --site "Dept X" --duration 300
cgo.exe --mode observe --serve   # force the mode explicitly
```

## CLI reference

| Command | Description |
|---|---|
| `cgo --serve` | Start the dashboard (REST + SSE, port 9090) |
| `cgo audit` | Run the client-side link audit → `data/link_audit.csv` |
| `cgo run` | Execute the reproducible AQM/CC evaluation matrix |
| `cgo verify` | Validate frozen CSVs against the SHA-256 manifest |
| `cgo figures` | Regenerate SVG figures from frozen CSVs |
| `cgo doctor` | Diagnose the environment (capabilities, tc, BBR availability) |
| `cgo shape --restore` | Apply/reset edge shaping (Façonnage du bord: qdisc, capacity, delay/jitter/loss) |
| `cgo service install` | Install as a system service |
| `cgo testbedsrv` | Run the bench-side testbed service |
| `cgo version` | Print version |
| `--mode observe\|full\|auto` | Override the execution mode (Windows defaults to observe) |

## Documentation

Full documentation is in French:

- [docs/usage.fr.md](docs/usage.fr.md) — operator manual
- [docs/architecture.md](docs/architecture.md) — internals and design
- [docs/api.md](docs/api.md) — REST + SSE API reference
- [docs/deploy.md](docs/deploy.md) — VM deployment guide
- [docs/methodology.fr.md](docs/methodology.fr.md) — measurement methodology and validity gates

## License

MIT — patches, suggestions, and comments are welcome.

