# CGO Audit Report Surface & Thesis Citation — Implementation Plan

*Date: 2026-10-28 · Status: proposed, awaiting human approval · Branch:
`rebuild/plan1-economics-audit` · Spec:
`docs/superpowers/specs/2026-10-28-cgo-audit-surface-citation-design.md`*

---

## 0. Scope guard

This plan touches the frontend audit card, its types and tests, and
`docs/api.md`. It MUST NOT:

- edit any `book/` file, the thesis text, Tableau 15, or the book build;
- change `pkg/audit`, `pkg/api`, or any Go response shape;
- change a CSV column, an export format, or the frozen corpus;
- add audit form fields, a new chart, a history view, or a report database;
- touch eBPF, IFB live acceptance, BBRv3, `.dist`, or the live bench;
- add a dependency, a new badge primitive, or a new design token;
- commit `epbfBBR.md`, `backup/`, `data/`, `book/`, `testbed-bootstrap/`,
  `unp*/`, `*.docx`, `*.pdf`, `cgo.exe`, or any user artifact;
- use `git add -A`; merge to `main`; or push.

The slice ships a citation string; the human pastes it into the thesis.

## 1. File map

| Path | Kind | Task |
|---|---|---|
| `web/frontend/src/lib/types.ts` | edit, ingress fields | 1 |
| `web/frontend/src/components/AuditReportCard.tsx` | edit, ingress, phases, provenance, citation | 1–4 |
| `web/frontend/src/lib/citation.ts` | new, pure citation builder | 4 |
| `web/frontend/src/lib/citation.test.ts` | new, unit test | 4 |
| `web/frontend/src/components/AuditReportCard.test.ts` | edit, DOM tests | 1–4 |
| `docs/api.md` | edit, two endpoints + CSV claim | 5 |
| `docs/superpowers/plans/README.md` | edit, slice index | 0 |
| `docs/VERIFY.md` | edit, one row | 6 |
| `docs/CHECKPOINTS.md` | edit, one entry | 6 |

## 2. Task 1 — Ingress remediation is typed and copyable

- `AuditRemediation` gains `linux_tc_ingress?: string[]` and
  `routeros_ingress?: string`.
- In the `Remédiation` section, after the upload block, render a
  `Téléchargement (ingress IFB)` sub-block only when one of those fields is
  non-empty. Copy buttons use the same `onCopy` idiom as the upload scripts.
- Tests: a fixture with ingress shows both scripts and both buttons; a
  fixture without it shows no ingress block.

## 3. Task 2 — The audit window is visible

- Render `report.phases` (baseline / load / recovery seconds) only when at
  least one value is present. Absent values are omitted, never zero-filled.
- Test: present phases render the three labels; absent phases render nothing.

## 4. Task 3 — One provenance idiom

- At the top of `Preuves`, render the shared `Provenance` line with the source
  file, the row count, and a frozen marker (`refresh`), matching the existing
  usages in `IntegriteView` and `ResultatsView`.
- Keep the two click-to-copy seal buttons and the Markdown report link beneath
  it; remove only the duplicated ad-hoc summary text.
- Test: the card contains exactly one `data-testid="provenance"`; both copy
  buttons and the Markdown link remain.

## 5. Task 4 — Cite a sealed report

- New `web/frontend/src/lib/citation.ts`:
  `auditCitation(report: AuditReport): string` returns plain Markdown with
  `audit_id`, site, `link_type`/`provider`, timestamp, the direction label
  (reuse `directionLabel` from `auditFlow.ts`), and the full
  `provenance.source_sha256`. Missing fields are omitted, not invented.
- The card gets a `CITER CE RAPPORT` button that copies it through `onCopy`.
- Tests: the string contains the audit id and the full 64-hex SHA-256; a
  report without a seal omits the seal sentence; no number absent from the
  report appears in the string.

## 6. Task 5 — Correct `docs/api.md`

- Document `GET /api/remediate`: query parameters (`audit_id`, `iface`,
  `routeros_iface`, `routeros_lan`, `qdisc`, `asymmetric`, `cpu_constrained`)
  and the `{audit_id, remediation, unavailable}` response.
- Document `GET /api/audit/presets` and its response.
- Correct the `format=csv` paragraph: the CSV carries header + rows; the
  provenance `hash8` line is Markdown-only today. Do not change the CSV.

## 7. Task 6 — Gates and close

```bash
cd web/frontend && npm test
cd web/frontend && npm run build
go test ./pkg/api/... ./cmd/cgo/...
python tools/verify.py
```

Expect vitest green, `tsc --noEmit` + Vite + bundle check green, Go green, six
gates `6 passed, 0 failed, 0 not run`. `make` is absent on this host; name
that the target ran as `python tools/verify.py`, do not waive it.

The build rewrites the tracked `web/frontend/dist/index.html`; it stays a
known working-tree modification, not a commit.

Record a `PENDING` row in `docs/VERIFY.md`, a checkpoint with the commands and
what stays unproven (no live audit; fixture reports only), then dispatch an
independent verifier. The builder may not certify its own row.

## Acceptance summary

- A download remediation can be read and copied from the report card.
- Phases, seal, and citation render from report fields only.
- The citation carries the report's id and full source SHA-256.
- `docs/api.md` matches the code for the two endpoints and the CSV export.
- No `book/`, Go, CSV, or frozen-corpus change; no live run.
