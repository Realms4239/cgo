# Two-mode craft plan (2026-11-22)

Direction (Operate, design read): trust-first operational instrument for DSI operators and an M2 jury; calm light Météo chart-paper system. Dials `DESIGN_VARIANCE 3 / MOTION_INTENSITY 2 / VISUAL_DENSITY 5`. Authority: ratified S5/S6 design system + 2026-11-20 spatial spec + the 2026-11-21 amendment. No new dependency, no new accent, no decorative motion. French copy. Absent NEVER zero; word+shape states; WCAG AA, 44px, forced-colors, reduced-motion; 890–1920 no overflow/no page scroll. Zero ledger churn; one commit per task explicit paths; never `git add -A`; no merge.

Skill gates: after the final edit run `node C:\Users\ASUS\.config\opencode\skills\impeccable\scripts\detect.mjs --json <changed targets>` once and record its output. Frontend: tsc/build/test + Edge Playwright unmocked (`PWCHROMIUM`, `--timeout=90000`, real Go binary). Native gates have a pre-existing `links` false positive on protected ledgers — record, never edit.

### Task 1: Synthèse — earn the verdict, remove the kicker

Delete the `RÉPONSE` eyebrow above the verdict (craft floor: a kicker above a heading is a ban). Give the verdict a single calm authored emphasis: larger quiet word, 1px hairline beneath, optical alignment of the state mark. Facts: label above value, tabular numerals, unit subordinate, absence word muted; no card boxes. Keep one primary action and one secondary with distinct intent; keep the honest empty invitation. Files: `src/components/cockpit/DecisionSpine.tsx`, `src/styles/components.css`, tests. Verification: DOM test asserts no kicker and the single verdict word remains `#verdict-word`; full gates.

### Task 2: Labo — question-led craft and honest states

Refine chapter heads (question lede under title, generous space above heading, none below), rail step vocabulary (current vs other with word+shape, no fabricated "done"), timeline/observer empty states that teach, and consistent button states (hover/focus/active/disabled) across the panel actions. No new animation beyond the ratified 180 ms panel slide. Files: `src/components/labo/{Chapter,PipelineRail,Observer,Comparer,Preuves}.tsx`, `src/styles/components.css`, tests. Verification: full frontend gates + Playwright + detector; one commit.

### Task 3: Final whole-branch review

Reviewer over both tasks; one fix wave + scoped re-review if needed.
