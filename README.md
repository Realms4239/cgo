# Meteolink [![version](https://img.shields.io/badge/version-1.0.6-blue)](VERSION) [![license](https://img.shields.io/badge/license-MIT-green)](LICENSE) [![go](https://img.shields.io/badge/go-1.25-%2300ADD8)](go.mod)

## What is it?

Meteolink is an open source, real-time network benchmark and interactive viewer that runs in a terminal in *nix systems or directly in your browser. Designed for constrained access links (4G/5G, fibre, VSAT), it delivers fast, verifiable AQM/BBR evidence on the fly. Meteolink audits your link from the client, replays it on a reproducible Linux bench, and presents the data directly in the terminal or via a live HTML dashboard — no vendor claims, only frozen CSVs with SHA-256.

More info at: [https://github.com/Realms4239/cgo](https://github.com/Realms4239/cgo).

## Features

Meteolink replays link profiles and outputs the data to the terminal or dashboard. Features include:

- **Completely Real Time**  
  All live panels and metrics are timed to be updated every 100 ms on the SSE stream (10 Hz) and every 250 ms on the TUI. The `live-wall-overlay` shows the delta `Figée vs appliqué` instantly.

- **Minimal Configuration needed**  
  You can just run it against your access link, pick the profiles `P1/P2` and let Meteolink run the matrix `pfifo_fast / fq_codel / CAKE × CUBIC / BBR` and show you the comparison.

- **Track Application Response Time**  
  Track `small p95` — the critical small objects (telemetry, alerts) that suffer most under bufferbloat. Extremely useful if you want to protect the traffic that matters.

- **Only one binary**  
  Meteolink is written in Go. To run it, you only need the binary — the React dashboard is embedded via `go:embed`. No database, no service dependency. It even features its own SSE server.

- **Nearly All Access Scenarios**  
  Meteolink allows any link profile (`P1` fibre 80 Mbit/s, `P2` 4G 20 Mbit/s, `P3` VSAT 5 Mbit/s importable via `POST /api/profile/import`). Predefined qdiscs include `pfifo_fast`, `fq_codel`, `CAKE` and CCs `CUBIC`, `BBR`.

- **Incremental Campaign Processing**  
  Need data persistence? Meteolink freezes every event to `data/runs/<run>/aqm_eval.csv` + `manifest.json` (SHA-256). `cgo verify` checks them, `cgo figures` regenerates the SVGs without Node.

- **Verifiable Provenance**  
  Every result shows `hash8 = sha256(dernier aqm_eval.csv)[:8]` — same hash in `Wall`, `Résultats`, `Provenance` and `report.md`. Quarantined `invalid` events stay counted, never hidden.

- **Edge Shaping Lever**  
  `Façonnage du bord` (`qdisc` + `capacity 1–1000 Mbit/s` + `delay/jitter/loss`) composes with `Surveillance continue` (`POST /api/watch`) and `Burst` (`POST /api/burst` CUBIC/BBR) — the live delta is the product.

- **Docker Support**  
  Ability to run the dashboard in a container; mount `data/runs` to keep the frozen evidence.

## Why Meteolink?

Meteolink was designed to be a fast, terminal-based link auditor. Its core idea is to quickly audit and compare AQM/BBR policies in real time without touching your routers (*great if you want to do a quick analysis of your 4G link via SSH, or if you simply love working in the terminal*).

It also serves as a practical tool for field diagnostics, making it easy to spot bufferbloat, unfair sharing (`JFI`), and wasted capacity directly from your link. While the terminal output (`meteolink top`) is the default, it has the capability to generate a complete, self-contained, real-time [`HTML`](http://192.168.174.128:9090) dashboard, as well as a [`CSV`](http://192.168.174.128:9090/api/report/export?format=csv) and [`Markdown`](http://192.168.174.128:9090/api/report/export?format=md) report.

You can see it more of a `monitor` command for your access link than anything else.

## Installation

### Build from release

Meteolink can be compiled and used on *nix systems. Download, extract and run the single binary with:

```
$ wget https://github.com/Realms4239/cgo/releases/download/v1.0.6/cgo-linux-amd64.tar.gz
$ tar -xzvf cgo-linux-amd64.tar.gz
$ ./cgo --serve              # http://127.0.0.1:9090
# or meteolink --serve
```

Verify with `checksums.txt` (SHA-256).

### Build from GitHub (Development)

```
$ git clone https://github.com/Realms4239/cgo.git
$ cd cgo
$ cd web/frontend && bun install && bun run build && cd ../..
$ go build -o bin/cgo ./cmd/cgo
$ ./bin/cgo --serve
```

### Distributions

It is easiest to install Meteolink using the preferred package manager:

#### Go

```
$ go install github.com/Realms4239/cgo/cmd/cgo@latest
$ go install github.com/Realms4239/cgo/cmd/meteolink@latest
```

#### npm (wrapper)

```
$ npm i -g meteolink
$ meteolink --serve
```

#### Windows (observation)

```
> cgo.exe --serve   # 127.0.0.1:9090, Windows = observe (no tc)
```

#### Docker

A Docker image can run the dashboard; mount the frozen runs to keep evidence:

```
$ docker run -p 9090:9090 -v ./data/runs:/data/runs meteolink --serve --addr 0.0.0.0:9090
```

### VM bench (the real `tc` bench)

The bench is an Ubuntu VM (VMware or VirtualBox). Convention: store VMs under `D:\VMs\` or `C:\VMs\` (e.g. `D:\VMs\ubuntu\ubuntu.vmx`); exception `D:\ubuntu.vmx` is still found. The engine scans `C:`/`D:` shallow ≤3 (`--deep` for full) via `vmrun list` + `inventory.vmls` + `VBoxManage`.

```
$ cp kit/cgo-vm.yaml.example kit/cgo-vm.yaml  # fill ssh/vmx
$ bash kit/engine.sh --action scan            # find .vmx/.vbox
$ bash kit/engine.sh --action ensure          # boot if SSH down
$ bash kit/engine.sh --action deploy          # build → cross → push → health
```

`kit/cgo-vm.yaml` is gitignored — never commit it.

## Storage

#### Default Frozen Archives

In-memory `live` rings provide better performance (180 s, 1800 pts at 10 Hz). For persistence, Meteolink freezes every event to `data/runs/<run_id>/aqm_eval.csv` + `manifest.json` (SHA-256). This storage has support for `cgo verify` and `cgo figures` as well.

#### On-disk CSV + Manifest

Each run freezes the matrix; `manifest.json` lists `file` + `sha256`. `GET /api/integrity` exposes `hash8`.

## Command Line / Config Options

See [options](docs/api.md) that can be supplied to the command or specified in `GET /api/schema`. If specified in the configuration file, long options need to be used without prepending `--`.

```
$ cgo --serve --addr 127.0.0.1:9090 --mode auto   # auto: Linux full, Windows observe
$ cgo audit --link-type 5g --site "Dept X" --duration 300
$ cgo verify
$ cgo figures
$ cgo doctor
$ cgo shape --restore
$ meteolink top --addr http://localhost:9090 --interval 250ms
```

## Usage / Examples

**Note:** The dashboard binds `127.0.0.1:9090` by default; pass `--addr 0.0.0.0:9090` to expose on LAN (as `kit/vm-install.sh` does).

### Getting Started

To output to a terminal and generate a live dashboard:

```
$ cgo --serve
# open http://localhost:9090
```

To audit your link from this workstation (non-intrusive, no admin):

```
$ cgo audit --link-type 5g --site "Dept X" --duration 300
# → data/link_audit.csv (p50/p95, small p95, goodput)
```

To generate a CSV report to stdout:

```
$ curl "http://localhost:9090/api/report/export?format=csv"
```

Meteolink also allows great flexibility for real-time filtering. To quickly diagnose bufferbloat on the live wall:

```
$ curl -X POST http://localhost:9090/api/shape -H 'Content-Type: application/json' -d '{"qdisc":"cake","capacity_mbps":20}'
# watch the live-wall-overlay Figée vs appliqué — -41% is the product
```

### Multiple Profiles

There are several ways to run multiple profiles with Meteolink. The simplest is to pass multiple profiles to the campaign:

```
$ curl -X POST http://localhost:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P1","P2"],"reps":3}'
```

It's even possible to import a custom profile from the UI (`Campagne → Profil personnalisé → P3`) or via pipe:

```
$ echo '{"id":"P3","capacity_mbps":5,"delay_ms":600}' | curl -X POST http://localhost:9090/api/profile/import -H 'Content-Type: application/json' -d @-
```

### Real-time Dashboard

Meteolink has the ability to output real-time data in the HTML dashboard. You can even email the `data/runs` folder since it is composed of single CSVs with no external file dependencies.

The process of generating a real-time dashboard is very similar to the process of creating a static report. Only `--serve` is needed.

```
$ cgo --serve --addr 0.0.0.0:9090
```

To view the report you can navigate to `http://<ip>:9090`. By default, Meteolink listens on port `9090`, to use a different port:

```
$ cgo --serve --addr 0.0.0.0:9870
```

And to bind to a different address other than `127.0.0.1`:

```
$ cgo --serve --addr 127.0.0.1:9090
```

### Filtering

#### Working with profiles

Another useful filter is to compare only one profile or one qdisc. On `Résultats`, use the chips `tous → P1` or `cake`. The `filtered` table and the `rank-verdict` recalculate instantly — all from frozen CSVs.

#### Burst Tests

To compare `CUBIC` vs `BBR` through the shaped edge without a full campaign:

```
$ curl -X POST http://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"bbr","seconds":4}'
$ curl -X POST http://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"cubic","seconds":4}'
# watch goodput + RTT on Tableau live
```

### Tips

Also, it is worth pointing out that if you want to run Meteolink at lower priority, you can run it as:

```
$ nice -n 19 cgo --serve
```

and if you don't want to install it on your server, you can still run the audit from your local machine:

```
$ ssh -n altfloat@192.168.174.128 'cgo audit --link-type 5g --site "Site X" --duration 30' | cat
```

### Troubleshooting

We receive many questions. Check first:

- `cgo doctor` — `tc` present, `CAP_NET_ADMIN`, `BBR`, `ping` — all green before a campaign.
- `cgo shape --restore` — clears stale `qdisc`s after a crash.
- `GET /api/health` → `{"mode":"full","version":"1.0.6"}` — `observe` on Windows is normal, campagne returns `501`.
- `go vet ./...` needs `web/frontend/dist` — `bun run build` first, otherwise `embed.go` fails.
- ECharts: never reintroduce `visualMap piecewise` nor `LinearGradient` area — it crashes `LineView` (`coord`) and freezes neighbours. `ChartSurface` `init` in `useEffect`, `dispose` on cleanup.

### Incremental Campaign Processing

Meteolink has the ability to process campaigns incrementally through its frozen storage. It works in the following way:

1. A matrix must be run first with `POST /api/run/start`, then the same dataset is frozen to `data/runs/<run>`.
2. `GET /api/results` scans the frozen CSVs; `GET /api/integrity` shows `hash8`.

To read persisted data only (without a new campaign):

```
$ cgo verify
$ cgo figures
```

## Contributing

Any help on Meteolink is welcome. The most helpful way is to try it out and give feedback. Feel free to use the GitHub issue tracker and pull requests to discuss and submit code changes.

## About

Meteolink is a terminal-friendly, self-contained network benchmark for auditing constrained access links and comparing AQM/BBR policies — runs in *nix systems or through your browser.

