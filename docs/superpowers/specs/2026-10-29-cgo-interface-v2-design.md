# CGO interface v2 — four-mode operator console

*Status: approved by user 2026-10-29 · Branch: `rebuild/plan1-economics-audit` ·
Slice 15 · UI only; no API, CSV, corpus, or book change*

---

## 1. Why

The current interface has six panels, a top tab row and a left rail. The audit
lives in a sidebar inside Campagne; the same proof/status appears in Synthèse,
the ribbon and Provenance; parameters such as target, provider and interfaces
are visible up front; and every view mounts at boot. The openchamber slice-10
session and the older UI plans converge on the same diagnosis: the product has
three verbs (audit, shape/compare, prove), but the interface exposes six
intents and a wall of controls.

The redesign is a **feature-preserving rearrangement**, not a feature cut.
Every current screen, modal, chart, table, control, replay, compare, citation
and history surface remains reachable. Only the duplicated navigation chrome
and the decorative WebGL background are removed.

## 2. Sources and decisions carried over

| Source | What it settles |
|---|---|
| `PRODUCT.md` | Three product verbs, primary DSI technician, no production control |
| `DESIGN.md` | Dark instrument world, tokens, Inter/mono discipline |
| 2026-09-03 leaderboard design | Benchmark banner, IQR/n, compare, method drawer |
| 2026-09-05 proof design | Valid-only, CI/MW/Cliff, results pinning, TUI parity |
| Plan 3 (2026-10-25) | Staged audit, report card, cockpit, results dépouillés, design floor |
| Slice 10 session (openchamber) | One-screen “what is protected / what it costs / what changed”; twin/observed labels; no replay buffer or interpolation; 10 Hz sampler |
| Slice 14 Synthèse | Proof counts, aggregate P2 constat, next action, last audit |

The IDEA memos are used for the **operator story** (protected classes, cost,
change) only. Their production-control claims, hostname/IP and unmeasured
latency numbers stay excluded, as slice 12/13 established.

## 3. The four modes

`Aperçu · Audit · Lab · Evidence`. One mode rail, keyboard 1–4. `Aperçu` is the
landing; its primary action is `Auditer un lien`.

**Aperçu** — the one-screen operator story, three columns/sections:
- **Ce qui est protégé** — the last sealed audit's critical-traffic deadline
  rate and the frozen P2 comparison (file simple vs best active discipline),
  each labelled `gelé`, `twin`, or `observé`. No production-protection claim.
- **Ce que ça coûte** — enterprise primary scenario and the retail overage
  bound from the last report; contractual-scenario disclaimer.
- **Ce qui a changé** — last audit vs the previous audit on the same site/link
  when both exist; otherwise an explicit “aucune comparaison disponible”.
Primary action: Audit. Secondary links: Lab, Evidence.

**Audit** — presets, four visible decisions, advanced disclosure, phase
timeline, full report, citation, remediation, history. See §6.

**Lab** — sub-tabs `Campagne · Live · Résultats`. All current cockpit, matrix,
live wall, controls, journal, leaderboard, filters, compare, export and replay
features remain; advanced controls collapse. See §7.

**Evidence** — current Provenance: proof banner, source explanations, runs
table, quarantine, manifest verification. See §9.

## 4. Feature preservation map

| Current feature | New home |
|---|---|
| Synthèse proof counts / P2 constat / next action / last audit | Aperçu; counts also in the ribbon |
| Audit presets, form, validation, live phases, stop, report, citation, remediation, history modal | Audit |
| Campagne cockpit, matrix, SSE state, gates, campaign start/stop | Lab › Campagne |
| Live cards, charts, comparison wall, journal, shaping/watch controls | Lab › Live |
| Résultats leaderboard, filters, IQR/CI/n, pin/compare, export, method drawer | Lab › Résultats |
| Provenance proof banner, runs, quarantine, verify, replay | Evidence |
| Command palette, settings drawer, explain popovers, toasts, flash | Global |
| TUI, CLI, API | Unchanged |
| PanelChooser tri controls (`groupes/craft/source`) | Lab header / `Affichage` |
| Rail gate strip, sparkline, phase/SSE status | Status ribbon or Lab › Live header |
| WebGL mesh | Removed; decoration only |

No feature surface is deleted. `AuditHistoryModal`, `CompareView`,
`SettingsDrawer`, `CommandPalette`, `Explain`, `LiveGuide`, `Interpretation`,
`Replay`, `Toasts`, `FlashBanner` and all tables/charts stay in the tree.

## 5. Shell and visual system

- **Mode rail**: three-state active marker, text labels when pinned, icons when
  collapsed; bottom bar under 700 px. Authored SVG icons only.
- **Status ribbon**: observation/full · SSE · last audit (site + grade or none)
  · corpus `runs / valid / degraded / quarantined` · seal `hash8` · `Vérifier`.
  Fed once by `/api/health`, `/api/integrity`, `/api/audit/status`; refreshed on
  reconnect/mode change; expandable provenance drawer. It never polls.
- **Visual**: existing dark tokens; hairline sections instead of nested cards;
  Inter for UI, JetBrains Mono for data, Cormorant only in the wordmark; accent
  colors only for state; no gradient text, no WebGL, no decorative motion.
- **Parameters**: only decisions that change the result are visible; the rest
  live under `Avancé` with localStorage persistence per mode.
- **Motion**: 150–250 ms, state only; reduced-motion paths preserved.

## 6. Audit

Visible form: site, link type, duration (20/60), direction (up/down/both).
`Avancé`: provider, target, ifaces, free duration, RouterOS LAN, asymmetric,
CPU-constrained. Three preset rows fill the form but never auto-launch.
`LANCER L'AUDIT` keeps the double-confirm ArmButton and existing validation.

Stages: `Déclarer` → `Observer` (phase timeline Repos/Charge/Drain, including
`charge-down`; `AuditLive`; cancel) → `Lire` (`AuditReportCard`: constat,
charge, CAP SLA, economics, proofs, citation, remediation, unavailable).
`NOUVEL AUDIT` and the existing history modal remain. Switching mode stops
polling but not the audit; returning re-reads `/api/audit/status`.

## 7. Lab

Sub-tabs mount on first visit and stay mounted.
- **Campagne**: profiles, repetitions, flows, channel, deadline, direction,
  one `LANCER LA CAMPAGNE`; qdiscs/CC/target/bench behind `Avancé`; matrix
  below; SSE/gates detail in a disclosure. Remove the three-card hero and the
  audit sidebar.
- **Live**: hero trace, compact metric strip, comparison wall, collapsible
  journal; `Réglages` disclosure for shaping/watch/essai. Keep the raw RX/TX
  interface chart, `smooth:false`, no interpolation, gap breaks.
- **Résultats**: leaderboard primary; one filter bar; compare toggle per row;
  export, citation, method drawer. Keep IQR/CI/n and the existing columns.

## 8. Aperçu data rules

- Uses only existing endpoints: `/api/integrity`, `/api/results`,
  `/api/audit/status`; for “Ce qui a changé”, `/api/audit/list` and
  `/api/audit/report?audit_id=` to load the previous comparable report. No new
  endpoint.
- A missing audit or missing comparison is stated, never zero-filled.
- “Ce qui est protégé” uses the audit's measured deadline rate and the frozen
  P2 aggregate with `n`; no production claim.
- “Ce que ça coûte” uses the report's enterprise scenario; retail remains the
  overage bound; the disclaimer is verbatim.
- “Ce qui a changé” requires two comparable audits on the same site/link;
  otherwise no delta.

## 9. Evidence

Current Provenance content preserved: proof banner (runs, valid, degraded,
quarantined, seal), runs table with replay, quarantine table, manifest verify.
Primary action `VÉRIFIER LES MANIFESTES`; detail one level down. The ribbon
carries the same counts, but Evidence remains the authoritative table.

## 10. Throughput

- Only the active mode and active Lab sub-tab mount; first-visit cache after.
- Remove `WebGLMesh` from the shell; remove the always-on multi-view mount.
- Status ribbon does three one-shot fetches, no polling.
- Live chart keeps the canvas path, draws on real sample arrival, no replay
  buffer, no interpolation; measure sample-to-pixel, do not claim 10 repaints/s.
- No new npm/Go dependency; `check-bundle.mjs` stays the budget gate.
- Acceptance must measure, not assert: first paint, mode switch, and chart
  advance on this host; record the numbers.

## 11. Errors

- Ribbon fetch failure: `—` + retry; never fabricated counts.
- Audit start 400/409/501/503: keep the form, show the server reason.
- Poll failure: “suivi interrompu”, offer refresh; never present a stale report.
- Stop: no partial CSV; report only when status returns.
- Lab campaign errors unchanged; Evidence verify errors explicit.
- Unmount during an audit: polling stops, audit continues; return recovers it.

## 12. Migration

1. Store: `mode` + `labTab`; replace `panel`. Update App, Rail replacement,
   Campagne/Live/Results/Evidence entry points, tests.
2. Shell: ModeRail + StatusRibbon + Aperçu; `SyntheseView` is reworked into
   `ApercuView` (not deleted); remove PanelChooser/Rail/WebGL.
3. Audit: extract `useAuditSession` from CampagneView; create `AuditView` +
   `AuditStage`; remove audit state/sidebar/history from Campagne.
4. Lab: `LabView` sub-tabs; trim Campagne hero; disclose advanced controls.
5. Evidence: re-shell Provenance as EvidenceView; no data changes.
6. Tests/screenshots; no API change.

## 13. Acceptance

- Four modes reachable at 1920/1366/390; Audit default action visible; no
  horizontal overflow; keyboard 1–4.
- All feature map rows verified by source test or screenshot.
- Audit form shows four decisions; advanced disclosure persists; start/stop/
  report/citation/remediation path unchanged.
- Lab sub-tabs preserve current data and controls; tri controls reachable.
- Evidence counts equal `/api/integrity`; seal/copy/verify work.
- Frontend tests updated and green; typecheck/build; full six gates.
- Screenshots: Aperçu, Audit idle, Lab Campagne, Lab Résultats, Evidence.
- No live audit claimed; no API/CSV/corpus/book change.

## 14. Out of scope

Backend changes, new endpoints, a new audit-comparison backend, live
Linux/eBPF/BBRv3/`.dist`, production access, book/thesis edits, a new chart
library, dark/light theme switch, role-based modes.

## 15. Risks

- The refactor is broad: extract the audit hook first, keep view internals,
  move one mode at a time, and keep the 161-test suite green.
- Aperçu can duplicate Lab/Résultats: keep it summary-only and link to detail.
- Removing chrome can break keyboard/e2e: update `Rail.test.ts`,
  `designFloor.test.ts`, `meteolink.test.ts` and capture all modes.
- Throughput claims must be measured on this host, with the compositor limit
  recorded rather than hidden.
