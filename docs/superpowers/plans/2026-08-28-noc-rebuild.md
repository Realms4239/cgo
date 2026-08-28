# NOC Rebuild — Layered Implementation Plan

**Spec:** `docs/superpowers/specs/2026-08-28-noc-rebuild-design.md`
**Order:** A backend live-publish → B primitives → C Wall → D Résultats+Campagne+Intégrité → E visual+ledger+ship. Each layer ships green (tests + build + my own screenshot inspection) before next.

---

## Layer A — backend live-publish (root cause)

**Files:** `pkg/campagne/campagne.go`, `pkg/campagne/campagne_test.go`
**Task A1 (TDD):**
1. Test `TestRunEventLivePublish` — fake deps (Ping 20-30ms values, Small ~50ms, StatsFn growing bytes, ChargeSec 1, BaselineSec 1, RecupSec 0): capture OnSnap snapshots while inside the charge window; assert ≥1 mid-charge snapshot has `RTTp50Ms > 0` AND `Smallp95Ms > 0` AND phase `charge`. RED (today: zeros).
2. Implement: in `RunEvent`, rework `collect(secs, phase)` to publish per round: build `liveEv := ev` copy, set fields from `metrics.Summarize` of samples-so-far (+ deadline pct), goodput = SumBytes delta×8/elapsed via StatsFn (fall back 0), `d.OnSnap(d.Snapshot(phase, loadFor(phase), liveEv, nil, nil, 0, prof, gates))`.
3. GREEN → `go vet + go test ./pkg/campagne ./pkg/api` → commit `fix(campagne): publish running measurements per probe round — SSE carries truth mid-phase (§5)`.

## Layer B — component primitives (kill inline soup)

**Files:** new `web/frontend/src/components/ui/{Card,CardHead,Stat,Pill,EmptyChart}.tsx`, touch views in C/D.
1. `Card {head, sub, right, children, testid}` — border hairline, surface, padding 12; `CardHead` 11 caps mono + sub 10 muted; `Stat {label, value, unit, color}` tabular; `Pill {on, label, onClick}` labeled toggle; `EmptyChart {hint}` ghost overlay (visible grid handled by grammar axes).
2. Grammar: `baseOption(title, unit, opts?: {idle?: boolean})` — graphic watermark only when `!idle`; chart craft colors as exported constants (`CRAFT = {live:'#5ad3e3', ok:'#1fa348', threshold:'#f4b400', bbr:'#b48ae0', steel:'#9aa3ad', danger:'#e22718'}`).
3. Tests: file-content asserts (primitives consume var(--t-)/no raw hex in views; grammar watermark conditional). GREEN → commit.

## Layer C — Wall rebuild

**Files:** `LiveView.tsx`, `PanelChooser.tsx`, `App.tsx` (nudge/modal placement), `index.css` (single badge), `chart-probe/visual-audit` specs updated.
1. Hero head via CardHead (`Petits objets p95`, sub `p95 · fenêtre 180 s`), single status dot+text.
2. 6-col bento: 8 cards + QDI + JFI; drops card absorbs DROPS DETAIL (remove bar).
3. Idle honesty: rings cleared when running flips true (start), '—' when idle-empty; banner states per §6 (OFFLINE/IDLE/STALE skip STALE — keep OFFLINE/IDLE/CHARGE/BASELINE/RÉCUP).
4. Labeled header pills via Pill primitive; `Actions rapides` gated to `running`; nudge → Sheet.
5. Visual round: screenshot idle+live, I inspect, fix, repeat until it reads as NOC instrument. Commit per green round.

## Layer D — Résultats + Campagne + Intégrité

1. Résultats: td padding fix, JFI column removed, bytes `MiB` formatting, scatter `visualMap: {show:false}` override + axis name visible, ab-bento rhythm.
2. Campagne: left workspace (phase stepper + Timeline + audit quick form) + sheet matrix; gates idle `—`.
3. Intégrité: ordered sections per spec (already mostly ordered — fix gates/idle + spacing).
4. Visual rounds as C. Commits per view.

## Layer E — ship

1. Full ledger (vet, go test + instrumentation, tsc, vitest, build+bundle, playwright --workers=1 all specs incl. visual-audit idle/live/tunnel).
2. Deploy → health + SSE cadence + provenance hash8 → merge --ff-only → worktree/branch cleanup → report.
3. Tunnel: keep supervisor alive; wire named tunnel when token arrives.
