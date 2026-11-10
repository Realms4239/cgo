---
leg: B (detector / build / browser evidence)
date: 2026-11-10
head: 740aa49 (branch `main`) — cgo-console@1.3.5
scope: evidence only; no product file edited by this leg
commands: detector (impeccable), npx vite build, npx vitest run, npx vite preview, curl
---

# Evidence — CGO cockpit UI, 2026-11-10

Evidence-only companion to the 2026-10-25 critique (`2026-10-25T21-24-59Z__web-frontend-src-views.md`). No heuristic scores, no taste judgements beyond what the detector/tests print. Nothing below was re-verified in a browser; see « Not verified ».

## Git state before / after this leg

Pre-existing, untouched by this leg (present when I started, not mine):

- `M web/frontend/src/App.tsx`
- `M web/frontend/src/components/CompareView.tsx`
- `M web/frontend/src/styles/index.css`
- `M web/frontend/src/views/DashboardView.tsx`
- `M web/frontend/src/views/IntegriteView.tsx`
- `M web/frontend/src/views/LiveView.tsx`

Mine, disclosed:

- `M web/frontend/dist/index.html` — tracked build artifact. First `vite build` rewrote it (asset hash bump); I restored it with `git checkout --` to keep the tree clean; that made it reference `assets/index-giCC3vkG.js` / `index-BWEj-VIY.css`, which the build had already deleted (vite `emptyOutDir`) and which are not in git (`git ls-files web/frontend/dist` → only `index.html`). To leave the served page working I rebuilt; `dist/index.html` now references `assets/index-i65bQEAe.js` / `index-BXhUG2Kz.css` and is modified vs HEAD. This modification is unavoidable post-build state, not a source change. Old hashed assets are unrecoverable from git.
- No other files created or edited. Scratch (raw `--no-inline-ignores` JSON) left at `C:/Users/ASUS/AppData/Local/Temp/opencode/detect-noignore.json`. Two scratch files I had briefly created under `web/frontend/` were removed.

`bin/cgo` (compiled Go binary) embeds an earlier frontend build: it still references the old asset hash. Out of scope; not rebuilt.

## 1. Mechanical detector (impeccable)

Invocation: `node C:/Users/ASUS/.config/opencode/skills/impeccable/scripts/detect.mjs --json web/frontend/src/styles web/frontend/src/views web/frontend/src/components`

- Result: **`[]` — zero findings, exit 0** (advisory findings included; none emitted).
- Scope: 80 files under the three roots (`src/styles` 3 CSS, plus `views`/`components` `.tsx`/`.ts`).
- Self-test (in temp dir, not the repo): a fixture containing `font-family: Inter` produced `overused-font`, exit 2 — the detector is functional; the empty result is not a broken scan.
- Waiver audit, `--no-inline-ignores` (same roots): **2 warnings, exit 2**:
  1. `gradient-text` — `web/frontend/src/styles/index.css:37` (`.wordmark-text`, `background-clip: text` + `linear-gradient` on the Meteolink wordmark; inline disable cites « identity wordmark pinned by DESIGN.md »).
  2. `overused-font` — `web/frontend/src/styles/tokens.css:7` (`@font-face 'Inter'` variable face; inline disable cites « Inter pinned by DESIGN.md, self-hosted variable face »).
  Both waivers are inline `impeccable-disable-line` comments with a stated DESIGN.md authority.
- **Cross-check gap:** no `DESIGN.md` exists anywhere in this worktree (`find . -name DESIGN.md` returns nothing), although AGENTS.md references one. The « pinned by DESIGN.md » justifications could not be verified against the design doc here.
- Plain-text mode (`detect.mjs <dirs>` without `--json`): no output, exit 0. Exit-code semantics observed: clean → 0, warnings → 2.
- Detector mode note: non-HTML inputs use regex pattern matching; there is no DOM/computed-style pass in this run.

## 2. Build and test (`web/frontend/`)

- `npx vite build`: **OK** (twice). 731 modules transformed; `✓ built in 23.30s` / `45.40s`. One warning, not acted on: `dist/assets/BenchView-*.js` 729.48 kB minified (241.71 kB gzip) exceeds the 500 kB chunk-size hint; suggests dynamic import / manualChunks / chunkSizeWarningLimit. CSS 73.09 kB (13.31 kB gzip).
- `npx vitest run`: **68 files passed, 321 tests passed** (Duration ~326 s). Exit 0.
- `make verify` / Go gates: **not run** (out of scope for this leg; frontend-only evidence).

## 3. Preview server and HTTP probe

- `npx vite preview --host 127.0.0.1 --port 4173` → **ports 4173, 4174, 4175 already in use**; server fell back and is listening on **`http://127.0.0.1:4176/`** (PID 14172). **Left running, as instructed.** Nothing was killed.
- Occupants of 4173/4174/4175 are pre-existing `vite preview --host 127.0.0.1 --port 4173` processes from this same worktree (PIDs 12336, 16248, 7284; identical command lines). I did not start or touch them.
- `curl -s -o /dev/null -w '%{http_code}'`: `4173 → 200` (pre-existing process, same worktree path, same `dist/`), `4176 → 200` (this leg's server). Current `dist/index.html` asset URLs return **200 on both** ports (`index-i65bQEAe.js`, `index-BXhUG2Kz.css`).
- This is an HTTP-level probe only. **No page was rendered or inspected** (no Puppeteer/Playwright/screenshot pass in this leg). Contrast, overflow, focus order and layout were **not** visually verified; nothing below should be read as « looks right in the browser ».

## 4. Code-visible signals (grep-level, `src/`)

Contrast (from `tokens.css` comments + `tokens.test.ts`, not re-measured here):

- `--text-faint #8a8a92` — documented measured ratios 5.75 (canvas) / 5.37 (card) / 5.03 (elevated), vs WCAG AA 4.5 floor for normal text. `tokens.test.ts` (« contraste des surfaces de console ») recomputes the ratio from the hex literals in the stylesheet on every suite run; the suite passes.
- `--t-danger #e22718` is documented as fills/dots only (4.11/3.90 on card/elevated); text uses `--t-danger-text #e84a3a` (4.95/4.71). Inline comment in `tokens.css` states the pair split.
- `--text-body #a9aeb6`, `--text-body-strong #d6d8dd`, `--text-primary #f2f2f4` on `#0a0b0e/#12141a/#181b22` — no low-contrast hex pairs flagged by the detector.

Focus / keyboard:

- Global `:focus-visible { outline: 1px solid var(--t-live); outline-offset: 2px }` — `index.css:14`.
- `.skip-link:focus-visible` reveals the skip link — `index.css:25`. Skip-link target `main:focus { outline: none }` — `index.css:85` (standard skip-target pattern).
- `.c-input { outline: none }` — `components.css:54`, but `:focus` restores border + outline (`components.css:56`) and `:focus-visible` adds outline-offset 2px (`:57`). Note: `outline: none` with only a border-color change on `:focus` would fail focus-visible; as written, an outline is present on keyboard focus.
- 8 `:focus-visible` rules across styles; 120 `aria-*` occurrences across 52 `.tsx` files; 45 `role=`; 6 `tabIndex`; 10 `aria-pressed`; **1 `aria-sort`** (live tables largely lack it — consistent with the 2026-10-25 critique's Sam note); 7 `scope=` / `role="columnheader"`.
- `prefers-reduced-motion` honored in CSS (`index.css:207` no-preference block, `index.css:287` reduce block, `components.css:63`) and in JS (`lib/anime.ts:12`, `lib/scroll.ts:6`, `Beam` via `lib/beam.test.ts`, `designFloor.test.ts`).
- Only 2 `<img>` in `src`, both with `alt` (`views/IntegriteView.tsx:289`, `:293`).

Overflow / truncation (code patterns, not rendered proof):

- 15 `overflow: hidden`, 8 `text-overflow: ellipsis`, 19 `white-space: nowrap`, 7 `overflow-x: auto` (scroll containers), 52 `min-width: 0` flex guards (the pattern the `header` comment at `index.css:27-30` describes for wrap/truncation safety).
- `.hd-audit` truncates with ellipsis (`index.css:39`); `.int-hash` truncates (`:667`); `.int-rdf-note` / `.int-hw-*` use `overflow-wrap: anywhere`. `.header`/`header` and `.app-header` both carry `min-width: 0`.
- `dist/assets/BenchView-*.js` 729 kB is a load-weight signal, not a layout one; unchanged by this leg.

Other detector-adjacent notes:

- Wordmark gradient + `drop-shadow(0 0 8px rgba(90,211,227,0.35))` at `index.css:37` is the one waived decorative gradient-text; the drop-shadow glow is not itself flagged by the detector.
- Two `impeccable-disable-line` comments exist; both cite DESIGN.md (see §1 gap).

## 5. Not verified / not run

- No browser render, screenshot, or interaction pass (the leg was HTTP probe only); no mobile viewport pass; no `--viewport 390x844` detector URL scan (would need Puppeteer; not requested, not run).
- `npx playwright` / e2e suites: not run. `make verify`, Go gates, bench: not run.
- DESIGN.md cross-check: file absent in this worktree; waiver authority unverified.
- Contrast ratios: taken from stylesheet comments + `tokens.test.ts` (which recomputes them); not independently re-measured by me.
- The pre-existing `src/` modifications (§ header) were not reviewed; they may interact with the detector results (e.g., `index.css:37` lives in a modified file).
- Old (untracked) `dist/assets` content was destroyed by `vite build`'s emptyOutDir; nothing in git could restore it. `dist/index.html` is intentionally left in its rebuilt state.

## 6. Raw outputs for traceability

- Detector default: stdout `[]`, stderr empty, exit 0.
- Detector `--no-inline-ignores`: JSON with the 2 findings quoted in §1; exit 2.
- Build log tail: `731 modules transformed`, `✓ built in 23.30s` (first) / `45.40s` (rebuild), BenchView chunk-size warning.
- Vitest tail: `Test Files 68 passed (68)`, `Tests 321 passed (321)`.
- Preview log: `Port 4173 is in use … Port 4175 is in use … ➜ Local: http://127.0.0.1:4176/`.
