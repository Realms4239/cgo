# CGO — Operator Cockpit Redesign (Design)

**Status:** approved by the operator for specification; not implemented.
**Date:** 2026-11-05
**Scope:** UX/UI redesign of the four Meteolink modes on desktop, 1024–1920 px.

## Context

The 2026-11-04 evidence set (`docs/img/shots/2026-11-04/`) shows a working instrument whose UI reads like a debugging panel to an operator who is not a network engineer. The main problems visible in the evidence are:

- large areas of empty black space inside a narrow, centrally constrained column;
- prose and labels rendered in a terminal face, including long French sentences;
- absent measurements replaced by repeated negative wording instead of one clear state and one useful next action;
- configuration cards with clipped identifiers;
- archives displayed as raw run identifiers rather than human-readable audits;
- technical terms appearing too early instead of in proofs or research mode.

The product already supports the data truth boundary described in `PRODUCT.md`: missing measurements remain absent and are never rewritten as zero. The redesign must make that boundary visible and usable without turning every absent state into an apology wall.

## Problem statement

An operator familiar with the four workflows but not the source code cannot reliably answer, in under five minutes:

1. Is this site currently healthy?
2. Is an alert deadline at risk, and how much?
3. What exactly do I need to do next?
4. Which audit or archive supports that decision?

The current interface exposes the underlying system rather than supporting those decisions. It is visually and textually optimized for inspection tooling, not for a DSI operator at a desk or beside a link.

## Users

- **Primary: DSI technician / operator** — needs a timely answer, site context, actionable evidence and a safe path to produce or consult an audit.
- **Secondary: network engineer** — needs numbers, series and raw proof details available in a proof drawer or research mode.
- **Stakeholder: M2 jury / director** — needs the current state, the accepted evidence and the cost, not unexplained instrumentation.

## Scope

### In scope

- Four modes: Tableau de bord, Auditer, Banc d'essai, Preuves & Rapports.
- One shared visual system and interaction model.
- Layout, typography, spacing, surface contrast, labels, empty/loading/error/stale states.
- Desktop acceptance range: 1024×768 and 1920×1080, with 1280/1600 considered during implementation.
- Existing front-end components and tests.
- Existing documentation refresh in `DESIGN.md` and `PRODUCT.md`.
- Accessibility, keyboard interaction, focus management and color-independent state labels.

### Out of scope

- RouterOS/CAKE remediation generation or parsing.
- Domain invariants for remediation eligibility.
- Any backend/API change unrelated to rendering, interaction, accessibility or existing UI state.
- Live audit, bench, Linux, CHR, BBRv3, eBPF or `.dist` runs.
- Fabricated measurements, placeholder metrics, or demo-mode output mixed into operator views.
- Mobile 390×844 work; that remains explicitly out of this slice.
- A visual redesign of the as-yet-unmeasured Windows service itself.

## Product constraints

- A missing measurement is **absent**, never rendered as zero.
- A stale reading is explicitly labelled stale.
- A failed load is explicitly labelled and leaves the previous reading unusable unless clearly marked as archived/stale.
- Technical artefacts stay available but do not dominate the primary operator surface.
- Dense evidence is allowed in **Preuves** and in proof drawers; it does not belong in operator prose.
- Copy never claims a measurement occurred if it did not.
- No colour-only state, no ambiguous icon-only status.

## Design principles

1. **One primary question per screen.** The dashboard answers status and next action; Auditer executes a measurement; the bench compares policies; Preuves provides evidence.
2. **Absence must be actionable.** If the system cannot fill a panel, it names the missing input and offers the strongest safe next step.
3. **Inter for language, mono for records.** Long sentences never use JetBrains Mono. Metrics, identifiers, timestamps, hashes, commands and units use monospace.
4. **Full-width desktop construction.** Avoid a narrow floating column with dead margins; use a 12-column grid.
5. **Readability before decoration.** No shadow-heavy glass, gradients as meaning, decorative animation or repeated alarms.
6. **Every state has a cause and an action.** Absent, loading, loaded, stale and failed are separate states.
7. **No fake sophistication.** Do not invent data, benchmarks, costs or verdicts to fill a lively-looking empty screen.

## Visual language

### Canvas

Dark but not pure black and not empty:

- background: cool near-black canvas;
- panel surfaces one step above the canvas, differentiated by hairlines and fill tones;
- no gradients, no colored glows, no 3D depth;
- full-width shell with controlled readable measure (~1600 px useful width).

### Typography

- **Inter**: navigation, titles, labels, prose, helper text, table headings and buttons.
- **JetBrains Mono**: values, units, timestamps, CSV/IP/SHA values, command text, counts and raw record identifiers.
- Heading hierarchy uses face contrast and tone, not a gigantic scale.
- Body text target: 14–16 px; table content 12–14 px; captions 11–12 px.
- Line length target: 55–85 characters for prose blocks.
- No forced all-caps for readable sentences; uppercase is reserved for eyebrows and very short status labels.

### Spacing and geometry

- 8 px base unit with 4 px micro-steps where needed.
- Card radius: 8–10 px.
- Spacing increases above headings, decreases below.
- 1024 px layout must not clip labels or shrink controls below intention.
- 1920 px layout must use its width for useful information and controls, not a single narrow column.

### Colour semantics

| Token role | Meaning |
|---|---|
| green | measured healthy / successful / safe within configured threshold |
| amber | attention / stale / degraded / objective at risk |
| red | measured failure / breach / required intervention |
| blue/cyan accent | active control or selected state, not a verdict of success |
| gray | absent / inactive / not available |

Every coloured state includes a text label. No success is communicated through colour alone.

## Shared shell

### Header

- Four destinations only: `Tableau de bord`, `Auditer`, `Banc d’essai`, `Preuves & Rapports`.
- Keyboard shortcuts remain visible where they fit and woven into accessible names.
- A wider desktop allows full labels and status text.
- Compact status cluster:
  - service state (`en ligne`, `données périmées`, `service arrêté`, `lecture impossible`);
  - last recorded audit age;
  - settings/help.
- Links to reports, copies and downloads belong in content; the header never holds a fake action.

### Main grid

- 12-column grid, column gap 16–24 px.
- Breakpoints:
  - 1024–1280: two-column, compact gaps;
  - 1280–<1600: three-column capable;
  - 1600–1920: wide two-zone compositions allowed, max content width ~1680 px.

### Footer / status line

- Source and evidence orientation only.
- No invented counts; a stale or missing source is labelled.

## Mode specifications

### 1. Tableau de bord

**Question:** “En deux secondes, où en est mon lien, et que dois-je faire ?”

Layout order:

1. Page heading and concise site/link context.
2. Three state cards:
   - `Santé du lien` — status, measured latency/bloat summary or explicit absence;
   - `Ponctualité des alertes` — percentage against the 220 ms target when measured; never a fabricated 0;
   - `Capacité gaspillée` — Ariary/month when economic inputs exist; otherwise unavailable, with the missing input named.
3. `Action recommandée` card:
   - when valid measured evidence supports it, one instruction plus one safe action;
   - if no valid action can be derived, explain precisely why and offer observation/audit instead;
   - never display a ready-to-copy script with no supporting measurement.
4. Recent measurements/elements (3–5): site, link type, date, status, verdict, action (`Ouvrir` / `Preuves`).

For a site with no recent measurement, the content collapses to one invitation:
- short statement: no recent audit for this site;
- what an audit proves;
- one primary CTA: `Lancer l’audit de 60 secondes` (or 20-second observation labelled as such);
- secondary CTA: `Voir un rapport archivé` only when the archive is reachable.

No repeating wall of negations.

### 2. Auditer

**Question:** “Comment je lance la mesure, et qu'est-ce que j'obtiens à la fin ?”

- Start with 3 decisions:
  1. Site/preset;
  2. Link type;
  3. Duration: 20 s (limited observation) or 60 s (full measurement), with the limitation explicitly named.
- Start control is a confirm-before-fire action (`DÉMARRER` then `CONFIRMER`), preserving intentionality.
- During the run:
  - stepper: `Repos → Charge → Vidange`;
  - clear phase status and remaining time;
  - exactly two primary visualizations: RTT/latency and throughput;
  - other diagnostics live behind `Détails scientifiques`.
- End state:
  - verdict card with grade/class, metric summary and deadline position;
  - provenance (`mesure de...`, source artefact, sealing identifier) available one click away;
  - safe actions: `Ouvrir le constat`, `Voir les preuves`, `Copier la recommandation` only if it exists;
  - if load phase failed, the page says exactly what failed and offers one concrete next step.

### 3. Banc d’essai

**Question:** “Quelle configuration vaut le coup, et qu’est-ce que le simulateur prouve ?”

- Four profile tiles, not truncated checkboxes:
  - `P1 — Fibre dédiée`: 80 Mbit/s · 20 ms RTT · reference site class;
  - `P2 — 4G dégradée`: 20 Mbit/s · 100 ms RTT · representative degraded channel;
  - `P3 — VSAT géostationnaire`: 5 Mbit/s · 600 ms RTT · satellite;
  - `P4 — Orbite basse`: 100 Mbit/s · 40 ms RTT · LEO.
- Each tile shows capacity, delay, loss assumption and use class on one scannable block.
- Selection state visible but color-independent.
- `Duel` comparison reads as before/after:
  - `Avant — file standard`;
  - `Après — file régulée CAKE`;
  - shown result: p95 delay under load and percentage of alerts delivered on time.
- Exceptional simulation details remain in research mode:
  - qdisc, flow mode, kernel, TCP cubic/BBR, ingress path, channel model;
  - never required to choose a profile.
- Any command or script copy action is available only beside the specific measured/simulated context it supports.

### 4. Preuves & Rapports

**Question:** “Est-ce que je peux citer ce résultat, et comment je le récupère ?”

- Main archive table with columns:
  `Date | Site | Type de lien | Durée | État | Verdict | Actions`
- Default order: most recent.
- Filters: `Tous / Valide / Dégradé / Observation seule`, with plain French labels.
- Actions:
  - `Ouvrir` — open proof/report;
  - `Rapport MD` — download the source report;
  - `Copier la citation` — copy a valid citation.
- Raw IDs and cryptographic hashes are never hidden, but they live in the proof drawer:
  - exact run ID;
  - source and telemetry SHA-256;
  - manifest, gates and corpus evidence;
  - export/print controls.
- An archive whose proof does not exist is not opened as a verified report; it is labelled `Preuve manquante` and cannot claim validity.
- Large corpus counts remain in Preuves, never in the main header.

## Shared components and boundaries

New or refactored components:

- `ShellHeader` — shell/nav/status orchestration.
- `LayoutGrid` — desktop twelve-column placement.
- `StatusCard` — semantic status with explicit label.
- `ActionCard` — one reason and one primary action.
- `EmptyMeasure` — no-measurement invitation.
- `LoadError` — failure + recovery/detail.
- `MetricPair` — paired presenter for compatible metrics.
- `PhaseStepper` — phase state and accessibility.
- `ArchiveTable` — semantic table with accessible sorting and action labels.

Unchanged unless required:

- Go remediation/audit domain code;
- measurement engine;
- corpus shape or frozen read flows;
- unsupported or deferred live targets.

## State model

Each mode maps its most relevant request to exactly one of:

| State | UI response |
|---|---|
| absent | named missing input; one action; no placeholder metric |
| loading | status, expected unit, cancel/retry where available; no prior values asserted |
| stale | value clearly tagged with age; read as historical, not current |
| failed | plain-language failure; cause, retry, fallback/evidence; no fake fallback |
| succeeded | value, source, timeliness and confidence/tier label |

Transitions must not accidentally preserve an old reading when a refresh has started. During refetch, either show the previous value explicitly marked historical or show an empty skeleton, never a blended intermediate.

## Accessibility

- Keyboard focus is visible with a high-contrast cyan outline.
- `:focus-visible` management is consistent across modal drawers and tables.
- All controls have accessible names; icon-only controls are not allowed.
- Status is never color-only; labels and descriptions are visible or screen-reader equivalent.
- Tables have proper headers, caption/name and sorted state.
- Form errors are linked to inputs (`aria-describedby`) and announced appropriately.
- Modals/drawers use focus trap and restore prior focus on close.
- Touch targets ≥44×44 px; minimum contrast AA.
- Live regions for progress and toast changes remain polite and rate-aware.

## Error and failure handling

- Network or service failure: `Lecture impossible — l’interface n’a pas pu lire les dernières preuves.`; visible `Réessayer`; optional `Détail` for HTTP/path error.
- API 404/report missing: `Le constat est absent pour cet identifiant.`; no fabricated card.
- Invalid/quarantine evidence: never shown as valid; excluded from success paths and tables expose status.
- Partial action success (copy for exported md): feedback states exactly what completed.
- Offline: read-only evidence marked as archived-only; measurement controls disabled with explanation, not silently hidden.

## Testing strategy

The implementation must avoid false confidence from snapshot tests that only assert the page still renders.

### Automated behaviour tests

- Vitest + Testing Library for component/state behaviour:
  - absent state renders zero fabricated values and one action;
  - load error preserves no misleading current value and provides retry;
  - loading prevents stale value from being presented as current;
  - keyboard/focus/aria relationships;
  - controls disabled with correct explanation.
- Assert absence of forbidden strings under absent state (for example synthetic `0 ms`, `0 Mbit/s`, or green verified proof text).

### Browser/e2e tests

- Playwright for 1024×768 and 1920×1080:
  - empty-to-audit path;
  - archived proof open path;
  - partial API failure path with mocked response;
  - focus and keyboard navigation.
- axe-core (or equivalent) on each primary screen and state; zero serious/blocking issues.

### Visual verification

- Screenshot matrix:
  - 1024×768 and 1920×1080;
  - all four modes;
  - at least one absent, one loaded, one failure/loading state per mode where reachable without live traffic;
  - no clipped label, horizontal overflow or duplicate apology copy.
- No screenshot is accepted where absence is disguised as a passing/green state.

### Artifacts of acceptance

- screenshot set generated from a reproducible script;
- a generated evidence manifest naming viewport, state and source;
- zero failing browser/integration tests;
- build and typecheck status;
- human five-minute DSI smoke-test recorded (no credentials or sensitive content);
- independent verifier before ledger promotion.

## Rollout and order of work

The implementation plan will split this redesign into stages, each independently verified:

1. Shared shell/tokens/layout and empty/failure patterns.
2. Tableau de bord and Auditer.
3. Banc d'essai.
4. Preuves & Rapports and evidence drawer.
5. Docs and final accessibility/audit pass.

Each stage preserves existing behaviour where it is already correct and removes deprecated overlapping components rather than leaving duplicates.

## Success criteria

- An operator can state, within five minutes, whether the site is healthy or not measured.
- An operator can identify the next safe action without reading the source.
- No unsupported value appears anywhere in operator surfaces.
- All four modes render at 1024 and 1920 without truncation or unintended empty voids.
- Every dangerous action remains explicitly confirmed.
- No stale/archived signal is mistaken for current state.
- The operator can reach and explain the relevant proof for any recommendation.

## Open decisions

No section remains open. The user approved:

- scope: all four modes in one spec, executed in phases;
- visual direction: restrained dark cockpit;
- no fabricated fallback content;
- desktop 1024–1920 only;
- DSI operator five-minute usability as primary success criterion;
- strong behaviour/e2e/a11y evidence, not green-only snapshot tests.

## Design record

- Source of critique: `C:\IDEA\NEW.md`, `C:\IDEA\NEW1.md`, plus attached directive in conversation.
- Evidence reference: `docs/img/shots/2026-11-04/`.
- This document does not assert any live benchmark, remediation or bulletin outcome.
