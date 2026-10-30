# CGO — project slice ledger

*Last updated: 2026-10-29 · Branch: `rebuild/plan1-economics-audit`*

This is the one-page history of every slice and every historical plan. It is a
map, not a claim ledger: the authoritative status of a number remains
`docs/VERIFY.md`, and the dated evidence remains `docs/CHECKPOINTS.md`.

## Status legend

| Status | Meaning |
|---|---|
| **Done** | Implemented, committed, and the slice's gates passed. |
| **Windows-green** | Engine/tests pass on Windows; live Linux acceptance still pending. Never call this “verified”. |
| **Builder pass** | The builder ran the slice gates; an independent verifier has not re-run them. |
| **Blocked** | A named external condition is missing. |
| **Deferred** | Deliberately not executed now. |
| **Historical** | Pre-rebuild plan; kept for provenance, not to be re-executed. |
| **Approved** | Design approved; implementation not started. |

## Slices

| # | Slice | Status | Key commits / evidence | What remains |
|---|---|---|---|---|
| 1 | Enterprise economics & audit | **Done** | `f2d4ae7`…`1af7d96`; plan 1 ledger | — |
| 2 | Hardened bench & campaign | **Done** | `a40b4b2`…`e01281d`; plan 2 ledger | — |
| 3 | UI/API surface | **Done** | `c45f6fb`, `09d117f`, `2764200`, `d277c7f`, `5fa92be`, `1ddab5c`, `33e9356`, `3c6de9b`, `5c13fcd`, `a759557`, `99f6351`, `eabf0db`, `968026e` | — |
| 4 | eBPF ground truth | **Blocked** | CO-RE source, `BUILD.md`, `make bpf` target at `1798f25` | No `bpf_probe.o`: CGO-DEV SSH key/password missing; PC2 WSL is 6.18 without clang/bpftool. |
| 5 | Trace-driven `.dist` | **Windows-green** | `41ef618`; focused tests + full gates | Real field trace ≥30 samples, then Linux 6.8 Task 8. |
| 6 | Ingress IFB shaping | **Windows-green** | `ad203f9`; independent verifier pass | Linux 6.8 Task 9. |
| 7 | BBRv3 matrix | **Windows-green** | `41ef618`; exact-token gating, kernel provenance, installer | 7.1 kernel boot + Task 8. |
| 8 | Audit report surface & citation | **Builder pass** | `eb4d01d`; ingress, phases, provenance, citation, API docs | Independent verifier. |
| 9 | Engineering system | **Done** | `04fad4d`; six gates, hooks, ledgers | — |
| 10 | Live interface observation | **Done** | `7424eaa`, `f33dbe2`, `3556493`; twin acceptance | Production read-only adapter awaits operator access; repaint rate host-bound (7.97/s), data path 10.27/s. |
| 11 | Frozen corpus seal anchor | **Done** | `7fa0ed4`, `53a2cc6`, `06bc310`; 159 manifests | Follow-up: FAIL instead of SKIP on a partial `book/`. |
| 12 | Book truth-up & restructure | **Done** | `31ad4b1`; book working tree, never committed | Human review of rendered book; placeholders. |
| 13 | Production-facing reframing | **Done** | `5293799`; title + Model A/B/C + WIS SLA | Book working tree; production access remains blocked. |
| 14 | Operator overview (Synthèse) | **Builder pass** | `ca09893`, `9a59276`; 161 frontend tests, 6 gates | Independent verifier. |
| 15 | Interface v2 + academic title | **Builder pass** | `66e0877`, `8f9af74`, `7b79a30`, `7303ede`; 46 files / 202 tests, 345.4 KB gz, six gates, book rebuilt | Independent verifier; human visual review of the four modes. |

## Historical plans and what survived

| Plan | Status | What it contributed | What was superseded |
|---|---|---|---|
| `2026-09-03-refonte-leaderboard` | **Historical** | Benchmark banner, dense results with IQR/n, method drawer, integrity tables, compare, campaign matrix, live timeline | “Live as landing / rail of five panels” replaced by slice 15's Aperçu/Audit/Lab/Evidence. |
| `2026-09-05-remontee-16-20` | **Historical** | Valid-only aggregates, IC95/IQR, Mann-Whitney/Cliff, results pinning, TUI parity, six tariff tiers | Extension B (live bench) blocked; not re-executed. |

Both were marked historical after the 2026-10-25 rebuild, but their surviving
ideas are implemented in slices 1–3, 8 and 10.

## Dismissed or potential ideas, reconciled

| Idea | Source | Disposition |
|---|---|---|
| Three service classes A/B/C (alerts, telemetry, bulk) | Slice-10 session | **Adopted as copy/framing**, but only the measured small/bulk series are shown; no invented class metrics. |
| One-screen “what is protected / what it costs / what changed” | `C:\IDEA`; slice-10 session | **Adopted** as Tableau de bord, split into real audit vs remediation simulation. |
| Production control / “one-click protection” | `C:\IDEA` | **Rejected as a claim**; the product exports reviewed Linux/RouterOS configuration and says `twin`/`observé`. |
| Production hostname/IP in the title or UI | `C:\IDEA` | **Rejected**; pseudonymous server, no IP/topology (confidentiality). |
| Production-oriented title | `C:\IDEA` | **Superseded** by the Option A academic title (2026-10-29): optimisation, temps réel, alertes hydrométéorologiques, AQM/BBR. |
| Executive one-screen dashboard replacing the cockpit | `C:\IDEA` | **Partially adopted**: Tableau de bord is the operator story; Banc d’essai keeps the engineering cockpit. |
| Nested Lab sub-tabs (`Campagne/Live/Résultats`) | Slice 15 draft | **Rejected after critique**: one linear Banc d’essai workspace, no nested tabs. |
| Blending real audit and synthetic simulation | Slice 15 draft | **Rejected after critique**: two separate, plainly labelled blocks — “Mesuré sur le lien réel” and “Simulation de remédiation”. |
| eBPF ground truth | Slice 4 | **Kept, blocked** on object + Linux 6.8. |
| Trace-driven `.dist` | Slice 5 | **Delivered engine**, live blocked on a real trace. |
| BBRv3 | Slice 7 | **Delivered engine**, live blocked on a 7.1 kernel. |
| RouterOS CHR QEMU validation | `epbfBBR.md` §5D | **Not planned**; zero-cost potential, no spec. |
| `kit schedule` nightly campaigns | 2026-09-05 extension | **Archived**; not installed, would dominate aggregates. |
| Production read-only collector | Slice-10 appendix | **Drafted, not sent**; awaits operator approval. |
| 10 chart repaints/s | Slice 10 | **REFUTED on this host** (7.97/s, compositor-bound); chart data advances 10.27/s. |
| Six-panel shell + top tabs + rail | Current UI | **Replaced** by four modes in slice 15 (`7b79a30`). |
| WebGL decorative mesh | Current UI | **Removed** for throughput; not a feature. The bench + ECharts chunk is now lazy-loaded (`7303ede`). |

## Current proof boundary

- **Windows-green**: audit engine, campaign/matrix, IFB command generation,
  `.dist` campaign loading, BBRv3 gating, report card, economics, live-interface
  plumbing, frozen seal, six gates.
- **Live/blocked**: Linux 6.8 acceptance for bench/IFB/`.dist`/BBRv3, eBPF
  object, production measurement, field audit.
- **Book**: rebuilt and audited in the working tree; `book/` and `PRODUCT.md`
  are never committed.
- **Interface**: slice 15 ships in the working tree — four modes, root audit session, one linear bench, real-vs-simulation dashboard, lazy bench chunk (111.2 KB gz initial). No live audit was claimed.
