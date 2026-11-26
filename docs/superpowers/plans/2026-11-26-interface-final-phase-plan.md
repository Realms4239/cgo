# Plan — final interface and system phase (2026-11-26)

Evidence: fresh captures at `C:/Users/ASUS/AppData/Local/Temp/opencode/cgo-demo-shots-v2/` from current HEAD. Operator wants lower cognitive load, working controls, honest real-time/eBPF, correct RouterOS validation and real audit.

## Global Constraints

- Pinned light chart-paper identity from DESIGN.md: no glass, glow, gradients, decorative shadows, new fonts or new dependencies. Refinement preserves identity, copy, behaviour and everything outside scope.
- No fabricated measurement or value. Absence stays a word, never 0. The objective is the locked 220 ms DGM value; do not change it to 200 ms.
- No gate relaxation, no frozen corpus/ledger/book change, no host contact in this plan, no `.dist` applied in a bench cell.
- TDD/DOM test red before each change; reviewer after each SDD task; explicit-path commits; French user-facing text, English identifiers; bundles stay within 600/350 KB gz; axe AA and 44 px targets at 890/1280/1920.
- Six gates must print `6 passed, 0 failed, 0 not run`.

## Task 1 — Synthèse chart and density

Fix the corrupted/overlapping Y-axis labels, the thin chart strip with unused vertical space, the cramped legend line, and the dense left fact column. Reduce clutter by grouping secondary facts into a quiet disclosure while keeping every fact reachable. Keep the 220 ms objective line and its label. Add DOM/spatial tests for axis label readability and no-overflow at 890/1280/1920.

## Task 2 — Labo spacing and real calibration availability

Increase spacing and quiet rhythm in the Configurer/Observer/Comparer/Preuves panels; remove cramped button rows and ellipsised labels. Make the distribution selector show the real sealed P4 calibration with honest provenance (source audit and sample count) without fabricating a calibration. Do not claim P1/P2/P3 calibrated. Add DOM tests.

## Task 3 — real-time/eBPF and functional paths

Prove the live display is fed by SSE from the Go hub with eBPF interface sampling during a real audit: add or repair the test around the truth gate and interface rings; verify RouterOS validation and audit start/stop/report paths with honest 409/501 handling. No live host run in this task.

## Task 4 — final verification

Rebuild, re-capture at all three widths, run typecheck/unit suite/six gates/axe, and final whole-branch review before push. Report exactly what is proven and what still needs a human-approved live run.
