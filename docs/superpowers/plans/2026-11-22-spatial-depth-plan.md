# Spatial-spec depth plan (2026-11-22)

Authority: ratified S5 `docs/research/2026-11-20-spatial-spec.md` §§1.2–2.4, S6 design system, and the 2026-11-21 amendment. The chart already implements the causal pair (three ECharts grids: primary latency, goodput, interface RX/TX) and the 220 ms markLine — do not duplicate them. Target only the missing depth, in order. No new dependency, accent, card, or decorative motion. French copy; absent NEVER zero; word+shape states; WCAG AA, 44px, forced-colors, reduced-motion; 890–1920 no overflow/no page scroll. Zero ledger churn; one commit per task explicit paths; never `git add -A`; no merge.

### Task 1: Field ruler and honesty footer

Make the `.tf-ruler` the literal shared axis: phase bands, event ticks from real `instrument.events` only, and a text legend (no color-only channel). Add the field footer: interface honesty (name, RX/TX, window, drops, source, gap) and JFI equity word + trace provenance — each from real data, else `—` + absence word, never a fabricated zero or green. Files: `src/components/cockpit/TimeField.tsx`, `src/styles/components.css`, tests. Verification: DOM/e2e oracles for empty vs real events and for absent footer values; tsc/build/test + Playwright.

### Task 2: 890 compact AnswerBand

At <1024, condense to a pinned ~112px band under chrome carrying verdict word + one key number + status word + PREUVE (44px), with the four facts behind a `4 faits ▾` disclosure and the Act adjacent below. Preserve the permanent verdict slot id/words, no page overflow, focus order, and reduced-motion. Files: `src/components/cockpit/{AnswerBand,DecisionSpine}.tsx`, `src/views/SynthèseView.tsx`, `src/styles/components.css`, tests. Verification: 890×512 Playwright geometry + axe; full gates.

### Task 3: Labo CH2 depth — matrix, gates, campaign timeline

Implement the progress matrix (profiles × qdiscs × cc, n/reps and state WORDS per cell, server order), tri-state gates G0–G7 as PASS/FAIL/— text with inline science disclosure, and a campaign timeline sharing one real clock. Idle = honest `attente` skeleton + one sentence; post-run = FINAL words, never reset-to-attente. No fabricated cells. Files: new `src/components/labo/{Matrix,Gates,CampaignTimeline}.tsx`, `Observer.tsx`, CSS, tests. Verification: DOM oracles for tri-state and absence; full gates.

### Task 4: Labo CH3 and CH4 depth

CH3: constat lede (narrative, no number), duel with shared scale and signed Δ + named winner, TradeSpace disclosure with real frontier summary, leaderboard with sortable 44px headers, honest ranking (unmeasured LAST with `—`), and always-visible pin A/B → drawer that refuses equality-on-empty with words. CH4: real corpus line + verify act, windowed 44px archive table with sealed-only actions and `Preuve manquante`, plus quarantine/figures/hardware/replay disclosures. No duplicate ProofDrawer. Files: `Comparer.tsx`, `Preuves.tsx`, CSS, tests. Verification: DOM/e2e oracles for unmeasured-last and refusal words; full gates.

### Task 5: Final whole-branch review

Final reviewer over the whole depth diff; one fix wave + scoped re-review if needed.
