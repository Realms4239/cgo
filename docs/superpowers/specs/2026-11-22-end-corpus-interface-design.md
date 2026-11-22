# END-corpus interface capabilities — design (2026-11-22)

Authority: operator chose "Include functional controls" 2026-11-22 after the C:\END corpus (7 files, 1491 lines, read-only) and the live UI were audited. Applies `make-interfaces-feel-better`. No corpus file is modified. Bench execution, frozen corpus, thesis and tag/freeze stay human-gated.

## 1. Audit finding

C:\END requires the interface to expose: BBRv3 (`CGO0:269`, `CGO5:296,355`), Canal Statique vs Empirique `.dist` and Direction Egress/IFB/Les deux (`CGO1:184`), field trace with SHA-256 sidecars verified through `/api/integrity` (`CGO5:307-312`), and eBPF kernel truth (`CGO5:290`, `CGO0:44`). Current live UI (Synthèse + Labo) surfaces none of them. `BenchCapabilities.tsx` already renders an honest BBRv3 line from real `/api/capabilities`, but it is mounted only in the retired `BenchConfig`, not in `App.tsx`.

## 2. Goal

Make these capabilities reachable and actionable from the live Labo, driven only by real host/bench sources. Where a source is absent, render `—` + a French absence word and name the missing source. No fabricated control, value, seal or campaign.

## 3. Backend source of truth (local, testable without a live bench)

- Extend the schema registry (`api/server.go` param list) with:
  - `link_channel` enum `["statique","empirique"]`, default `statique`;
  - `link_direction` enum `["up","down"]`, default `up` — matches the frozen `model.Event.Direction`;
  - `dist_name` — a dynamic enum populated only by real sidecar-verified calibrations.
- Make `burst_cc` dynamic: include `bbr3` only when `capabilities.BBRv3` is true, with the corpus wording rule (native v3 vs CUBIC+BBRv1 on stock 6.8 with detection ready) exposed as a separate honest `bbr3_state` string.
- New read-only `GET /api/calibrations` listing real files under the calibration directory: `{name, sha256, samples, p50_ms, verified}`. Empty list is honest, not an error.

## 4. Frontend surface

- Mount the existing `BenchCapabilities` capability strip in the live Labo `Configurer` (reuse, do not duplicate): qdiscs, ccs, kernel, and the honest BBRv3 word.
- Add Canal, Direction and Distribution controls rendered from `/api/schema` + `/api/calibrations`; empirical options exist only when a verified sidecar is present. A control with no real option is disabled and accompanied by `—` + the absence word — never a fabricated placeholder option.
- Preserve the honesty contract: no default that claims a measurement, no green from an unconfigured value, no seal from an unverified file.
- Feel-better: `tabular-nums` on samples/p50, `text-wrap: pretty` on absence sentences, 44px hit areas, calm hairlines, concentric radii where nested; no new hex, font-size, accent, keyframes or `transition: all`.

## 5. Field trace and eBPF provenance

No functional controls. Add an honest provenance strip reading `/api/integrity` (hash status, `verified`/`degraded` only from real checks) and `/api/doctor` (host checks when present). Where no real field trace or eBPF confrontation exists, render `—` + `non mesuré` and name the missing source (authorized ≥30-sample capture; `ebpf_vs_tc_report.json`). No value is inferred from the live bench in observe mode.

## 6. Slice order

1. Backend schema + calibrations endpoint + dynamic CC/BBRv3, with Go unit tests.
2. Labo capability strip + Canal/Direction/Distribution controls, with DOM oracles and unmocked Playwright.
3. Provenance strip for field trace / eBPF, with DOM oracles.
4. Final feel-better pass and whole-programme review.

Each slice: spec → plan → fresh implementer → fresh reviewer → gates → one explicit-path commit. No parallel subagents.

## 7. Out of scope

Bench execution, hardware/VM writes, sadcnwp, frozen corpus, thesis, v2.0.0 tag/freeze, ledger edits, gate weakening, new dependency. Functional controls select and persist configuration; they do not claim a measurement until a real run produces one.
