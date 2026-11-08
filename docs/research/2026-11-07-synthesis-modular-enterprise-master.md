# Synthesis: Master Demonstration + DGM Needs + Modular Toolkit — Critical Insight

*Date: 2026-11-07 · Basis: 36 per-module primary-source files in `docs/research/` (10,091 lines). No code changed. Every claim below cites a module file or repo `file:line`; unverifiable items are marked `unverified — primary source not found`.*

## 0. The one critical insight nobody wants

A Master is not proven by whether you can draw a Grafana graph. It is proven by whether a second engineer can **re-run your experiment on a different host and get the same numbers with the same evidence**. Everything else is theatre.

The pasted markdown ("just use Grafana/Prometheus/Mosquitto/CHR") gets the direction of operability right and the direction of science wrong. It would trade the only examinable IP you have — a **frozen, sealed, falsifiable measurement bench** — for a prettier dashboard. The correct morph is not replacement. It is **extraction**: keep the evidence engine, extract it into versioned tools, and let standard tools do the boring ops work around it.

## 1. What a Master really needs to demonstrate (Dublin 2nd cycle)

From `2026-11-07-module-master-dgm.md`:

- **Original thinking** = reproducible evidence engine (`README.md:5`, `docs/architecture.md:1-4`), not a new qdisc. FQ-CoDel is RFC 8290 — novelty is measurement method.
- **Problem-solving in new environments** = frozen corpus + six gates + seal anchor (`docs/VERIFY.md:24-26`, `docs/frozen-manifests.sha256` 159 manifests) that the jury re-runs on their host.
- **The 5-minute falsifiable arc** (all air-gap-safe, stdlib/Python only):
  1. `python tools/verify.py` → six PASS
  2. `python book/tools/audit-archive.py` → both OK (159 runs / 528 rows)
  3. `sha256sum -c docs/frozen-manifests.sha256` → 159 OK, mutate a temp copy → seal catches it
  4. `hash8` identical across Wall / Résultats / Provenance / `report.md`
  5. No measured load ⇒ `Unavailable`, never green (`TestVerdictNeverGreenWithoutGrade`, fixtures only — stated honestly)
- **ISGEI/MSTI rubric itself: `unverified — primary source not found`.** Confirm word count, examiners, rubric weights, and unpublished-software rules with the secretariat *before* the defence. Do not invent a rubric on stage.

## 2. What Météo Madagascar really needs (verified vs speculation)

**Verified (repo + primaries only):**
1. Protect small CAP/WIS alerts against bulk EUMETSAT downloads on shared 4G/VSAT links. 220 ms is a **DGM operator objective** (`pkg/metrics/cost.go:75-83`), never a WMO mandate. WIS 2.0 = MQTT pub-sub over public Internet (R11); CAP 1.2 = XML payload (R12). CGO measures links; it is not a WIS2 Node.
2. Cost in Ar/h per scenario from sourced tiers (Yas 25,000 Ar/4.5 Go = 5,556 Ar/Go; GOLD 1M Ar/mo 200 Mbps CIR — `docs/data-prices.md`).
3. Router-grade remediation (Linux `tc` + RouterOS `/queue tree` — `pkg/remediate`). Live proof BLOCKED (`VERIFY.md:57`).
4. Read-only evidence with zero operational risk — grant a–d (`2026-10-30-cgo-case-study-contract.md`) still BLOCKED (`VERIFY.md:60`).

**Excluded as speculation:** Ampandrianomby topology, sadcnwp cadences, WIS2 registration state. Do not claim CGO is a Node.

## 3. The essential structure — what this project essentially is

**One evidence engine + five things around it.** CGO = reproducible measurement bench + audit + provenance (`README.md:5`). Around it:

```
model → metrics/probe/profile → qdisc/iface → campagne/remediate → audit/results/figures → api/doctor → cmd/* + embedded React
```

Released as **SemVer Go modules with SLSA provenance, CycloneDX SBOMs, and OCI images**, demonstrated via the six-gate + seal + hash8 arc. WIS 2.0/CAP are *neighbours* (transport/payload standards CGO's measurements must satisfy), not parts.

## 4. The modular toolkit — 12 independent tools

Each tool below is one research file in `docs/research/2026-11-07-module-*.md`, each with its own target repo, API, slices, and anti-toy gates. Extract in dependency order; keep `internal/` and `pkg/campagne`/`pkg/api` as the shell.

| # | Tool (repo) | What it does | Why it must exist separately |
|---|---|---|---|
| 1 | `cgo/model` | Domain types (profiles P1–P4, qdiscs, gates G0–G7) | Stdlib-only kernel; every other module depends on it first |
| 2 | `cgo/metrics` | `Percentile`, `CostARPerH`, `JFI`, `AssessVerdict` | Pure logic; Prometheus honesty contract lives here |
| 3 | `cgo/probe` + `cgo/audit` (+ `cgo-audit` CLI) | Ping/small/bulk sondes; 60 s phased audit → sealed report | The experiment; `cgo-audit` is the field tool |
| 4 | `cgo/qdisc` + `cgo/remediate` | `tc` Reset/ApplyNetem/ApplyShaper; Linux + RouterOS generators | Linux-gated; `cgo-chr-validate` reuses `cmd/cgo-chr` + parser |
| 5 | `cgo/results` + `cgo/figures` | Frozen scan → medians; SVG; `hash8` | The jury-proof; no TSDB can do this |
| 6 | `cgo/api` (sse/metrics) + `cgo/doctor` + `cgo/iface` | REST/SSE hub; preflight checks; `/proc/net/dev` sampler | Ops surface; `/metrics` must emit honest gauges, never static zeros |
| 7 | `cgo/cite` + `@cgo/proof-panel` | Seal-gated citation (`""` without `source_sha256`) + accessible proof UI | Pure TS+Go core; React panel as peer-dependency |
| 8 | `cgo/exporter` | Prometheus exposition (`# HELP`/`# TYPE`, `Content-Type` correct) | Static-zero `/metrics` is currently a ship-blocker |
| 9 | `cgo/guard` | Bearer auth + Nginx `limit_req` (1r/s burst=3 → 429, CSRF) | The security FAIL that blocks enterprise |
| 10 | `cgo/attest` | `manifest.json` → corpus digest + signed anchor + in-toto bundle | Turns the frozen seal from static anchor into SLSA L1 attestation |
| 11 | `cgo-lite` stack | OCI digest-pin + `cpu.max`/`memory.max` + rollback drill (replaces `:latest` + dead rules) | The distribution anti-toy gap (no SBOM/signing/`NET_ADMIN` today) |
| 12 | `cgo-demo` | Doctor-verify + falsifiable 5-min arc as a CLI | The defence is a tool, not a slide deck |

**Do-not-extract verdicts (from file:line evidence):**
- `internal/kit` + `internal/vm` → keep as internal implementation of a standalone `cgo-kit` **binary**, never a Go library (Go `internal/` rule + SemVer).
- `pkg/campagne` → orchestrator of the monolith; extract only if a second runner appears.
- `cmd/cgo-gui` (Fyne) → **own repo** — pre-existing Linux build failure (`VERIFY.md:27`) must never block the core build.

**Build-what-we-lack (each is a real gap in the tree):**
- `cgo-wis2-probe` — lab MQTT observer on `cgo/v1/...` topics (no MQTT in `pkg/` at all; WIS2 is neighbour, not part).
- `cgo-cost` — tier calculator with mandatory source/date (Yas 200k tier is now stale at 250 Go; Airtel grid superseded).
- `cgo-bpf-observer` — object + loader for `kernelTelemetry` (today `ebpf` tier is object-less and loader-less; BLOCKED `VERIFY.md:29`).

## 5. Hard truths (the crush)

1. **Toy status is bimodal.** UI toy was true; evidence core was never toy. Slice 27 fixed the UI; Slice 28 fixed ops. The "toy" charge does not survive either.
2. **The standard-tools charge is half-right.** Buy Grafana/Prometheus/Mosquitto/Nginx for ops. Do not buy them *instead of* measurement science.
3. **eBPF is the real research frontier** — and the real blocker. Object-less + loader-less + BLOCKED on Linux 6.8. That is where the thesis IP actually lives; everything else is engineering hygiene.
4. **Supply chain is the shallowest enterprise gap** — `tar.gz` + `.sha256` only, no SBOM/signing (R11–R13, R15). Fix with GoReleaser + SLSA + CycloneDX before any Grafana dashboard.

## 6. Slices in dependency order (do not skip)

S-A `model` → S-B `metrics` → S-C `results/figures` → S-D `qdisc/remediate` → S-E `probe/audit` → S-F `api/doctor/iface` → S-G `cite/proof-panel` → S-H `guard` → S-I `attest` → S-J `exporter` → S-K `lite stack` → S-L `demo`. Each slice = new `go.mod`, README, semver, tests, SBOM. At most three active (AGENTS.md cadence).

## 7. Open decisions (human-gated, do not guess)

Live bench (Linux 6.8) · eBPF build host · CHR image licence · WIS domain fetch · case-study grant a–d · port exposure · tariff re-source · microsecond/RAM/3-file sentence deletion. Until each is signed, it stays `BLOCKED` on screen — that honesty *is* the defence.
