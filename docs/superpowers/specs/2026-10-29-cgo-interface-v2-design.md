# CGO interface v2 and academic title

*Status: revised after critical review, awaiting user spec review · 2026-10-29 ·
Branch: `rebuild/plan1-economics-audit` · Slice 15 · UI and product copy only;
no API, CSV, corpus, or measurement change*

---

## 1. Why

The current interface has six panels, a top tab row and a left rail. The audit
lives in a sidebar inside Campagne; proof/status is repeated in Synthèse, the
ribbon and Provenance; parameters such as target, provider and interfaces are
visible up front; every view mounts at boot. The operator cannot answer three
questions in one place: what is the real link doing, what would remediation
change, and what does it cost?

This spec replaces the navigation shell, flattens the benchmark workspace, and
sets a two-layer language: plain operational French by default, the academic
statistics one disclosure down. It preserves every current feature.

## 2. Academic title (adopted)

The thesis title becomes:

> **Optimisation du transport réseau et garantie d’échéances temps réel pour la
> diffusion d’alertes hydrométéorologiques sur des liens d’accès contraints**

Subtitle:

> **Régulation du bufferbloat par gestion active de file (AQM), contrôle de
> congestion BBR et application au système d’information de Météo Madagascar**

The title states the optimisation and the real-time deadline objective; the
subtitle states the mechanism and the institutional case. “Garantie” is the
evaluated deadline objective on the bench, not a production SLA. The chapter I
limitation paragraph stays: no production measurement, no production control.
The string is applied verbatim to `PRODUCT.md`, the book title page, the cover
XML, the running header and the metadata; no other title variant survives.

## 3. Four modes, no nested tabs

`Tableau de bord · Auditer · Banc d’essai · Preuves & Rapports`. One mode rail,
keyboard 1–4. Each mode is a single screen; the benchmark is not a second tab
taxonomy.

- **Tableau de bord** — the operator story: what the real link is doing, what
  remediation would change, what it costs, what changed since the last audit.
- **Auditer** — presets, four visible decisions, live phases, full report,
  citation, remediation, history.
- **Banc d’essai** — one linear workspace: `Configurer → Observer → Résultats`.
  All campaign/live/results features sit in those three sections; no sub-tabs.
- **Preuves & Rapports** — corpus proof, runs, quarantine, audit reports and
  exportable citations.

The mode rail replaces the old `PanelChooser` + `Rail` pair. The decorative
WebGL mesh is removed. `AuditHistoryModal`, `CompareView`, `SettingsDrawer`,
`CommandPalette`, `Explain`, `LiveGuide`, `Interpretation`, replay, toasts and
the flash banner stay.

## 4. Tableau de bord: real vs simulated

The dashboard separates the two evidence worlds in plain French. It never uses
`twin`, `gelé`, `AQM`, `QDI` or gate IDs as primary labels.

**État du lien réel — audit**
- Shows the last sealed audit: site/link, timestamp, loaded small-object delay
  (“Délai sous charge : 328 ms — élevé”), critical-traffic deadline rate
  (“Alertes à l’heure : 60 % — danger”), and the audit verdict.
- Label: **« Mesuré sur le lien réel, sans modification »**.
- If no audit exists: “Aucun audit sur ce poste” with the audit action.

**Remédiation recommandée — simulation**
- Shows the frozen P2 comparison as a projection, not a measurement of the
  audited link: file simple vs best active discipline, with `n`.
- Plain labels: “Avec la file simple : 60 % des alertes à l’heure” vs “Avec
  CAKE : 98 %”; “Délai sous charge : 328 ms → 210 ms”.
- Label: **« Simulation de remédiation sur profil P2 — pas une mesure du lien »**.
- A link to the full results; no claim that the production link was shaped.

**Ce que ça coûte**
- Enterprise primary scenario and retail overage bound from the last report;
  “Perte financière : X Ar/h”; contractual-scenario disclaimer verbatim.

**Ce qui a changé**
- Compares the current audit with the previous audit stored in the local
  session summary (site, link, deadline rate, loaded delay, timestamp).
- Label: “Comparaison locale avec l’audit précédent”. If absent or the
  site/link differs: “Aucune comparaison disponible” — no multi-round-trip
  archaeology, no blank wait.

## 5. Two-layer language

**Layer 1 — operator.** Plain French, no statistics:
- `small_p95` → “Délai sous charge” with a qualitative band (faible/modéré/
  élevé/critique).
- `deadline_ok_pct` → “Alertes à l’heure” with a danger/ok state.
- `qdi_ms` → “Instabilité du lien” (stable/modérée/sévère).
- `cost_ar_per_h` → “Perte financière” (Ar/h).
- grade → “Qualité du lien” (A–F with words).
- remediation → one primary button “Copier la configuration MikroTik” and one
  “Copier la configuration Linux”; raw scripts behind the drawer.

**Layer 2 — researcher.** A per-mode `Détails scientifiques` disclosure:
- percentiles, IQR, IC95, n, Mann-Whitney, Cliff’s delta, QDI;
- gates G0–G7 and `Unavailable` entries;
- full provenance chain, source/telemetry SHA-256, raw scripts;
- the frozen campaign table and figure links.

The default screen must remain readable by a DSI technician or director; the
drawer must retain every number the thesis uses.

## 6. Auditer: global execution state

The audit session lives at the application root, not inside the view. It owns
presets, form, validation, start/stop, polling and the final report. The
`Auditer` view is a presentation layer over that state.

- Visible decisions: site, link type, duration (20/60), direction.
- `Avancé`: provider, target, ifaces, free duration, RouterOS LAN, asymmetric,
  CPU-constrained.
- Presets fill the form; they never auto-launch.
- Start → `Observer` (phase timeline, live cards, cancel) → `Lire` (report,
  citation, remediation, history).
- Switching to Tableau de bord or Banc d’essai while an audit runs does not
  stop polling or lose the run; the ribbon shows “Audit en cours” and the view
  is restored on return. No “suivi interrompu” unless the backend actually
  stops.
- Start refusals keep the form and show the server reason; stop keeps the
  no-partial-CSV rule.

## 7. Banc d’essai: one linear workspace

No sub-tabs. One page with three sections:

1. **Configurer** — profiles, repetitions, flows, channel, deadline, direction,
   and `LANCER LA CAMPAGNE`. `Avancé`: qdiscs, CCs, target, bench.
2. **Observer** — the live hero trace, compact metric strip, comparison wall,
   journal and `Réglages` (shaping/watch/essai). Hidden/collapsed until a run
   or replay starts; the campaign state is already global SSE.
3. **Résultats** — the leaderboard, one filter bar, IQR/CI/n in the scientific
   drawer, compare pin, export, citation and method drawer.

All current campaign/live/results features are preserved inside these sections.
Replay from Preuves & Rapports lands in `Observer`; `CompareView` opens from
`Résultats`; no feature requires a nested tab.

## 8. Preuves & Rapports

Corpus proof banner (runs, valid, degraded, quarantined, seal), runs table with
replay, quarantine table, manifest verification. Add the existing audit list /
last report and its citation/export surface, so “rapports” is a real destination
rather than a hidden modal. No new endpoint; `/api/audit/list` and
`/api/audit/report` already exist.

## 9. Shell and status ribbon

- **Mode rail**: four modes, authored SVG icons, labels when pinned, bottom bar
  under 700 px. Keyboard 1–4.
- **Status ribbon**: observation/full · SSE · “Audit en cours” when running ·
  last audit site/grade · corpus counts · seal `hash8` · `Vérifier`. Fed once by
  `/api/health`, `/api/integrity`, `/api/audit/status`; refreshed on reconnect
  and after an audit; no polling.
- **Visual**: existing dark tokens; hairline sections; Inter for UI, JetBrains
  Mono for data, Cormorant only in the wordmark; no gradient text, no WebGL,
  no decorative motion.
- **Parameters**: only decisions that change the result are visible; advanced
  controls persist per mode in localStorage.

## 10. Feature preservation map

| Current feature | New home |
|---|---|
| Synthèse counts/P2 constat/next action/last audit | Tableau de bord; counts also in ribbon |
| Audit presets/form/live/report/citation/history | Auditer |
| Campagne cockpit/matrix/SSE/gates | Banc d’essai › Configurer + Observer |
| Live cards/charts/wall/journal/controls | Banc d’essai › Observer |
| Résultats leaderboard/filters/compare/export | Banc d’essai › Résultats |
| Provenance proof/runs/quarantine/verify | Preuves & Rapports |
| Audit list/report/citation | Preuves & Rapports + Auditer |
| PanelChooser tri controls | Banc d’essai › Affichage |
| Rail gates/sparkline/phase/SSE | Ribbon or Banc d’essai › Observer |
| Command palette/settings/explain/toasts/replay | Global |
| WebGL mesh | Removed; decoration only |

No feature surface is deleted. `SyntheseView` is reworked into the dashboard,
not removed.

## 11. Throughput

- One mode mounted at a time; Banc d’essai sections mount when their stage is
  reached. First-visit cache after.
- Audit polling lives at the root and continues across modes; no re-fetch burst
  on return.
- Remove WebGL and duplicated chrome; status ribbon does three one-shot fetches.
- Live chart keeps canvas, sample-arrival drawing, no interpolation, no replay
  buffer; measure sample-to-pixel; do not claim 10 repaints/s.
- No new dependency; `check-bundle.mjs` stays the budget gate. Record first
  paint, mode switch and chart advance on this host.

## 12. Errors

- Ribbon fetch failure: `—` + retry; never invented counts.
- Audit start 400/409/501/503: keep the form, show the server reason.
- Backend stop: “suivi interrompu” only when `/api/audit/status` says not
  running; a mode switch never produces that message.
- Stop: no partial CSV; report only when status returns.
- Campaign errors stay in Banc d’essai; Evidence verification errors explicit.

## 13. Migration

1. Store: `mode: dashboard|audit|bench|evidence`; root audit session state;
   remove `panel`/`labTab`.
2. Shell: `ModeRail`, `StatusRibbon`, dashboard; remove PanelChooser/Rail/WebGL;
   rework `SyntheseView` into the dashboard.
3. Audit: move `useAuditSession` to the root provider; AuditView presentation.
4. Bench: merge Campagne/Live/Results into one page with three sections; remove
   sub-tabs; move tri controls to `Affichage`.
5. Evidence: add audit reports list/citation; keep corpus proof.
6. Title: update PRODUCT.md, book title page, cover XML, running header and
   metadata to the §2 string.

## 14. Acceptance

- Four modes at 1920/1366/390; keyboard 1–4; no nested tabs.
- Tableau de bord shows “Mesuré sur le lien réel” and “Simulation de
  remédiation” as separate, plainly labelled blocks; no `twin`/`gelé` primary
  labels.
- Audit survives a mode switch: polling continues, ribbon shows “Audit en
  cours”, return restores the live/report state; no false “suivi interrompu”.
- Banc d’essai preserves every campaign/live/results control; no sub-tabs.
- Every operator label is plain French; scientific details live in the drawer.
- Preuves & Rapports lists reports and citations without new endpoints.
- Frontend tests updated and green; typecheck/build; full six gates; screenshots
  for the four modes and the three bench sections.
- Title string appears in PRODUCT.md, cover/title page, header and metadata;
  book audit/build still green.
- No live audit claimed; no API/CSV/corpus/measurement change.

## 15. Out of scope

Backend changes, new endpoints (including a diff endpoint), live Linux/eBPF/
BBRv3/`.dist`, production access, field audit, book prose changes beyond the
title, role-based modes, new chart library, dark/light switch.

## 16. Risks

- Flattening the bench may produce a long page: use progressive sections and
  anchor navigation, not nested tabs.
- Root audit state must be created before the view extraction; test mode-switch
  continuity explicitly.
- Removing chrome can break e2e/keyboard tests; update `Rail.test.ts`,
  `designFloor.test.ts`, `meteolink.test.ts`.
- Title change touches the book and PRODUCT; run `audit-archive`, `audit-biblio`
  and `build-docx` after the edit; `book/` stays uncommitted.
