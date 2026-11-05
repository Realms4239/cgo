---
target: the dashboard views
total_score: 23
max_score: 40
na_heuristics: 
p0_count: 1
p1_count: 4
timestamp: 2026-10-25T21-24-59Z
slug: web-frontend-src-views
---
# Critique — CGO dashboard (views), 2026-10-26

Method: dual-agent (A: vision critiques of fresh 1920px captures + deep Explore code scan · B: impeccable detector agent)

## Heuristic scores (Nielsen, 0-4)
| # | Heuristic | Score | Key issue |
|---|-----------|-------|-----------|
| 1 | Visibility of status | 2 | Audit run shows only « audit en cours… » — no phase, no countdown, no live numbers despite 3-phase protocol |
| 2 | Match real world | 3 | French operator language good; jargon wall (QDISC/CC/gates) softened by Explain tooltips |
| 3 | Control & freedom | 2 | No way to cancel a running audit; A/B compare pin needs 2 clicks/row |
| 4 | Consistency | 1 | Two token generations, 350 hex literals, 584 inline styles, 12 font sizes, dead primitives (DataTable/Stat/Skeleton/Badge + c-input/c-kpi generation) |
| 5 | Error prevention | 2 | Free-text duration 10-600, hidden direction default, 7 visible jargon fields invite mistakes |
| 6 | Recognition vs recall | 3 | Presets, Explain glossary, labels; control box relies on remembered prior state |
| 7 | Flexibility | 3 | Ctrl+K palette, per-panel hotkeys, presets, compare |
| 8 | Minimalist | 2 | Chip soup (~20 chips one row), dual encodings (heat + bar/whisker), 380px fixed cockpit always dense |
| 9 | Error recovery | 2 | Toasts/flash/rollback good; BUT gaspillage card silently always « non mesuré » (broken binding auditReport.waste.* vs economics.*) |
| 10 | Help & docs | 3 | LiveGuide, Explain, runbook pointers |
| **Total** | | **23/40** | Acceptable — significant improvements needed |

## P0
- [P0] Audit economics display never works: CampagneView reads `auditReport.waste.ar_per_hour`; Report JSON has `economics.*` — silent « non mesuré » forever.

## P1
- [P1] Audit card renders ~15% of report: no direction/JFI/CAP-SLA/scenarios/unavailable/down-mirrors/seal-copy/markdown link.
- [P1] No audit progress feedback (60s blind wait); no cancel.
- [P1] Results identity cell = 4 stacked lines + always-visible A/B buttons per row (chip/button soup); chip row ~20 buttons nowrap-scroll.
- [P1] Token system bypassed at scale (350 hex / 584 inline / dead primitives) — consistency ceiling.

## P2/P3
- width transitions ×7 (detector: layout thrash); gradient text ×1 (Meteolink shimmer — intentional?); Inter overuse ×1 (taste rule); 9-10px label sizes; top-nav vs left-rail active-state disconnect (vision); title duplication in Campagne header.

## Personas
- Alex: no cancel-audit, no keyboard submit in form, pin compare = 2 clicks/row.
- Sam: aria-pressed ×2 only, no aria-sort (live tables), color-only heat meaning, 10px mono labels.
- Jordan (DGM tech): jargon fields up-front (Cible ping, ifaces), blind 60s wait, wall-of-numbers report.
