# MUSTREAD — CGO / Meteolink canonical pack

> This index is the single entry point. `docs/MUSTREAD/01-core-directive.md` is the governing directive; the seven chunk docs are the execution units. Source corpus: `C:\END` (1,491 lines), never modified.

## 1. Overarching approach

METEOLINK is one Go binary with an embedded React cockpit and a Linux eBPF probe for a constrained meteorological link. This pack turns the seven-file CGO blueprint into a governed, chunk-by-chunk execution route: purge and simplify, lock security, enforce data honesty, install the real-time pipeline, prove it with honest tests, ship it in a signed container, then close with live evidence, the thesis and the v2.0.0 freeze. MUSTREAD governs where it conflicts with the operating card; operator-level bans still bind.

## 2. Five-lens map

| Lens | Document(s) |
|---|---|
| Result | `02-target-system.md`, `16-chunk-6-live-proof-freeze.md` |
| Mechanism | `01-core-directive.md`, `13-chunk-3-realtime-pipeline.md` |
| Contribution | `03-product-contribution.md` |
| Actions | `10-…` through `16-…` (one per chunk) |
| Usage | each doc's "At a glance" → "Usage" bullet |

## 3. Document map

| Doc | Source | Purpose |
|---|---|---|
| `01-core-directive.md` | `CGO.md` 1–237 | Governance: laws, loop, debug playbook, reviewer protocol |
| `02-target-system.md` | `CGO0.md` 1–286 | Zero-bloat target and 7-chunk route |
| `03-product-contribution.md` | `CGO1.md` 1–237 | Problem, personas, value, architecture brief |
| `10-chunk-0-secret-purge-bloat.md` | `CGO2.md` 1–133 | Purge + bloat + SSoT |
| `11-chunk-1-security-lockdown.md` | `CGO3.md` 1–91 | AuthZ, SSRF, injection, timeouts |
| `12-chunk-2-data-honesty-integrity.md` | `CGO4.md` 1–97 | Absent-never-zero, integrity, RTT |
| `13-chunk-3-realtime-pipeline.md` | `CGO5.md` 1–89 | eBPF → SSE → Float64Array → HUD |
| `14-chunk-4-honest-tests.md` | `CGO5.md` 90–167 | synctest, happy-dom, live Playwright |
| `15-chunk-5-container-ci-release.md` | `CGO5.md` 168–267 | Docker, strict CI, signed release, ops |
| `16-chunk-6-live-proof-freeze.md` | `CGO5.md` 268–410 | Live proof, field trace, thesis, freeze |

## 4. Persistent recall

- **Locked constants:** version `2.0.0`; audit `60 s` = 10/40/10; SLA `220 ms`; bind `127.0.0.1:9090`; canonical title « Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar ».
- **Precedence:** MUSTREAD > `AGENTS.md` operating card > skill defaults. Operator-level constitution bans cannot be waived by a project document.
- **Path binding:** corpus `pkg/qdisc/` → repo `qdisc/`; `pkg/results/` → `results/`; `pkg/remediate/` → `remediate/`; `pkg/probe/` → `probe/`; `pkg/audit/` → `audit/`; `cmd/gates` stays until a replacement six-gate runner exists.
- **Retained contradictions:** title, 2 vs 4 UI modes, `cmd/gates`, `2.0.0` vs `1.3.5`, stale `pkg/...` paths. All are preserved and resolved per chunk at plan time, never deleted.
- **Quoted source:** corpus text may quote operator-ban commands (`git add -A`, `|| true`, `@latest`, live-bench and unauthorized-target instructions). These are preserved verbatim as source material, not executable instructions; the precedence table governs what an agent may actually do.

## 5. Completeness crosswalk

| Source | Lines | Destination |
|---|---|---|
| `CGO.md` | 1–237 | `01-core-directive.md` |
| `CGO0.md` | 1–286 | `02-target-system.md` |
| `CGO1.md` | 1–237 | `03-product-contribution.md` |
| `CGO2.md` | 1–133 | `10-chunk-0-secret-purge-bloat.md` |
| `CGO3.md` | 1–91 | `11-chunk-1-security-lockdown.md` |
| `CGO4.md` | 1–97 | `12-chunk-2-data-honesty-integrity.md` |
| `CGO5.md` | 1–89 | `13-chunk-3-realtime-pipeline.md` |
| `CGO5.md` | 90–167 | `14-chunk-4-honest-tests.md` |
| `CGO5.md` | 168–267 | `15-chunk-5-container-ci-release.md` |
| `CGO5.md` | 268–410 | `16-chunk-6-live-proof-freeze.md` |

## 6. Execution status (SDD ledger)

| Task | Doc | Owner | State |
|---|---|---|---|
| Pack build + injection | `INDEX.md` | coordinator | pending |
| Chunk 0 | `10-…` | implementer + reviewer | pending |
| Chunk 1 | `11-…` | implementer + reviewer | pending |
| Chunk 2 | `12-…` | implementer + reviewer | pending |
| Chunk 3 | `13-…` | implementer + reviewer | pending |
| Chunk 4 | `14-…` | implementer + reviewer | pending |
| Chunk 5 | `15-…` | implementer + reviewer | pending |
| Chunk 6 | `16-…` | implementer + reviewer | pending |

## 7. Handoff

`writing-plans` derives one implementation plan per chunk from Part A / B / C of its doc. SDD executes chunks 0→6 in order, one subagent at a time, with a task review after each and ledger rows updated here.
