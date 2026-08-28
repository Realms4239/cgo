# Meteolink NOC Rebuild — Frontend Full Visual Rebuild + Backend Live-Publish

**Date:** 2026-08-28
**Status:** Approved (user, full-rebuild scope)
**Base:** `main@9e3130b`
**Evidence:** playwright full-page screenshots inspected (shots/audit-*.png) — charts empty during live run (SSE frames carry zeros mid-charge), internal spec labels rendered as UI, invisible empty-chart grammar, merged table cells, dead JFI column, 80% dead Campagne viewport, gates FAIL at idle, modal blocking content, cryptic unlabeled controls, duplicate SSE badges, watermark glaring on idle charts.

## Root cause (backend, §5 violation)

Design §5: `Live 10Hz: OnSnap {phase, rtt/small/goodput/drops/wasted/cost/deadline, gates, running} → Live.Set`.
`RunEvent.collect()` gathers samples but publishes nothing until phase end → hub broadcasts zeros for the whole charge window → every chart flatlines. **Fix at source:** after each probe round (~300ms), publish running truth via OnSnap: p50/p95-so-far, small p95-so-far, goodput from `tc -s` SumBytes delta / elapsed, drops delta, wasted = drops×1448, cost, deadlineOK-so-far. Gates continue folding live.

## Visual system contract (§3 authority)

- Tokens only: `var(--t-*)`, `--hairline`, `--gap`, `--rail-w`, `--surface-*` — **zero raw hex in views** (charts keep craft colors as constants in grammar).
- One radius (0), one gap system, mono-caps discipline: `card-head` 11px caps, `view-title` 20 centered, provenance 10 right.
- Watermark `kept only when !idle` (§3) — idle charts show NO watermark.
- Empty chart = visible hairline grid + axes + centered ghost message (never a black void).
- Motion: state changes only, `prefers-reduced-motion` collapses (already wired).

## Wall (LiveView)

1. Hero 300: head `Petits objets p95` + sub `p95 · fenêtre 180 s`; one status dot + state text; NO spec labels.
2. Charts through grammar; y-axis fits data; single CHARGE band phase-driven.
3. **One 6-col bento**: 8 MetricCards + QDI + JFI on the same grid (no 4+2 misalignment). `DROPS DETAIL` bar removed (drops is a card).
4. Timeline 48 under charts; single SSE badge (footer); `Actions rapides` only while running; OnboardingNudge inside Sheet.
5. Header controls labeled: `groupes [on] · craft [line] · source [both]`.
6. Honest idle: rings cleared at run start; stale snapshot not shown as live.

## Résultats (Drawer)

- Fix merged cells (padding), remove dead JFI column (needs detail=1), bytes formatted (`MiB`), scatter visualMap OFF (opt-in), axis name visible, bars grey→green ★ best (§2).

## Campagne (Sheet cockpit → workspace)

- Main area: phase stepper + Timeline + audit quick-form (left) — no dead void.
- Sheet 380: matrix + cost preview (unchanged role).
- Gates idle = `—`, never FAIL.

## Intégrité

- Ordered progressive disclosure per §4 inside the scroll container; figures side-by-side.

## Verification bar (every layer + final)

- Screenshots 1920/390 idle + live — **agent personally inspects each round**.
- Ledger: `go vet + go test (-tags=instrumentation)` · `tsc` · `vitest` · `build + check-bundle 600/350` · `playwright --workers=1` full · deploy → health + SSE 10Hz ±20% · provenance hash8 chain.
- Tunnel: durable Cloudflare named tunnel pending valid token (user re-issue).

## Out of scope

New metrics, kit changes, replay machinery, mobile app layout beyond 390 screenshots, auth.
