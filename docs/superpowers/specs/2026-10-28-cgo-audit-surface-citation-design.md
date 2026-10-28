# CGO Audit Report Surface & Thesis Citation — Design Specification

*Date: 2026-10-28 · Status: proposed, awaiting human approval · Slice 8 ·
Branch: `rebuild/plan1-economics-audit`*

This slice finishes the audit report surface and makes a sealed report citable.
It does **not** edit the thesis: `book/` is untracked user material and is never
committed (AGENTS.md). The thesis edit stays the human's.

---

## 0. Context

Slice 3 built `AuditReportCard` and the audit API. Slice 6 added ingress
remediation to `audit.Report`. Slice 10 added the live interface chart. The
card has not caught up:

- `AuditRemediation` in `web/frontend/src/lib/types.ts` has only `qdisc`,
  `reason`, `linux_tc`, `routeros`. The JSON already carries
  `linux_tc_ingress` and `routeros_ingress`; the TypeScript model drops them,
  so a download remediation cannot be copied from the UI.
- `AuditReportCard` never renders `phases` (baseline / load / recovery), even
  though the report carries them.
- The card uses ad-hoc seal buttons instead of the existing `Provenance`
  component, so the same seal is rendered two different ways in the product.
- The thesis cites the frozen campaign corpus (Tableau 15) but has no path to
  cite a sealed `audit.Report`. A reader cannot copy an audit id and its
  SHA-256 into the book.
- `docs/api.md` omits `GET /api/remediate` and `GET /api/audit/presets`, and
  claims a provenance `hash8` line in the CSV export that the code does not
  write.

## 1. Decision log

- **D1 — No `book/` edit.** The slice ships the citation text; the human pastes
  it into the thesis. `book/` stays untracked and uncommitted.
- **D2 — Reuse the existing primitives.** `Provenance`, `Badge`, `Card`,
  `CardHead`, `Stat`, and `ArmButton` only. No new badge, card, or token.
- **D3 — Ingress only when present.** Render `linux_tc_ingress` /
  `routeros_ingress` only when the arrays are non-empty. An upload-only report
  shows no empty download section.
- **D4 — Phases only when present.** Render baseline / load / recovery seconds
  from `report.phases`; absent values stay absent, never zero-filled.
- **D5 — The citation invents nothing.** It carries `audit_id`, site, link
  type, provider, timestamp, direction, and the full `source_sha256` from the
  report's provenance. Every number in it is copied from the report.
- **D6 — Fix the API document, not the CSV shape.** `docs/api.md` is corrected
  to describe what the code does. No CSV column or export format changes.
- **D7 — No premature ground truth.** No eBPF, BBRv3, or live-IFB badge. A
  `source` value is shown only when a report actually carries one.

## 2. What the operator gets

1. **Download remediation is usable.** When a report has ingress commands, the
   card shows the Linux ingress script and the RouterOS ingress script with
   copy buttons, next to the existing upload scripts.
2. **The audit window is visible.** Baseline / load / recovery seconds appear
   in the report card, so the reader can see what was measured.
3. **One provenance idiom.** The card's seal block uses the shared
   `Provenance` line (source file, telemetry file, row count, SHA-256) with the
   existing copy action, matching `IntegriteView` and `ResultatsView`.
4. **A citable report.** A "Citer ce rapport" action copies Markdown:

   ```
   Audit CGO <audit_id> — <site>, <link_type>/<provider>, <timestamp>,
   direction <up|down|both>. Rapport scellé : source_sha256 <64 hex>.
   ```

   The copied text is plain Markdown; nothing is written to `book/`.

## 3. API document corrections

- Document `GET /api/remediate` (query parameters and response shape).
- Document `GET /api/audit/presets`.
- Correct the `format=csv` claim: the CSV carries header + rows; the provenance
  `hash8` line is Markdown-only today.

## 4. Acceptance

- `cd web/frontend && npm test` — vitest green, including new DOM tests:
  ingress shown when present and absent otherwise; phases shown when present;
  the citation contains `audit_id` and the full `source_sha256`; the seal line
  renders once, not twice.
- `cd web/frontend && npm run build` — `tsc --noEmit`, Vite build, bundle check.
- `go test ./pkg/api/... ./cmd/cgo/...` — unchanged Go surface stays green.
- `python tools/verify.py` — six gates, exit 0 (`make` absent on this host; the
  target runs as the Python entry point, named not waived).
- No live audit, bench, namespace, qdisc, or listener is run. The card is
  exercised with fixture reports, not a live measurement.

## 5. Out of scope

- Any `book/` file, the thesis text, Tableau 15, or the book build pipeline.
- New audit form fields (`flows`, `qdisc`, `asymmetric`, `cpu_constrained`,
  `routeros_lan`) in `CampagneView`; that is a separate slice if wanted.
- New charts, a new audit history view, a report database, or a download
  endpoint for `audit_report.json`.
- eBPF, IFB live acceptance, BBRv3, `.dist`, or any new measurement.
- CSV schema, export format, or frozen corpus change.
