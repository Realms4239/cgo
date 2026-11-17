# Thesis PDF build — preflight plan (2026-11-17T10:55Z)

No build was run. Word COM is available (v14.0); `book/` is git-clean; all book
audits pass; `cgo figures` is wired. Three write boundaries need your explicit
approval, because `AGENTS.md` forbids each of them by default.

## Current title (confirmed in both sources)
`PRODUCT.md` and `book/content/00-front.js` agree verbatim:

> Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar

Product name: Meteolink (CGO). The plan calls this "title 2" (Task 0.6 Step 1,
Task 6.6). It is already set on both surfaces. No title edit is proposed.

## Exact commands the plan requires

1. `go run ./cmd/cgo figures` — regenerates figures from the corpus into
   `data/figures/`. **Write boundary A: `data/` is in "Never touch".**
2. `powershell -NoProfile -ExecutionPolicy Bypass -File book\tools\build-book.ps1`
   — runs `node book/build-docx.js` (reads `book/content/*.js`), the `fix-*.py`
   passes, copies the result to `C:\cgo\memoire.docx`, then `finalize.ps1`
   (Word: fields, pagination, PDF) → `C:\cgo\memoire.pdf`.
   **Write boundary B: `book/` build outputs if any, and boundary C below.**
   **Write boundary C: overwrites `C:\cgo\memoire.docx` (user document, 3.3 MB,
   dated 2026-10-23) and creates `C:\cgo\memoire.pdf`.**
3. Re-run `book/tools/audit-*.py` after the build (read-only).

Note: `build-book.ps1` has `-SkipWord` for a non-Word rehearsal, but even that
copy step writes `C:\cgo\memoire.docx`, so it is not a no-write dry run.

## What is NOT proposed
- No edit to `book/content/*.js`, no change to title 2, no change to the frozen
  159-run corpus (`book/data/runs`).
- No commit of `.docx` or `.pdf` (permanent ban in `AGENTS.md`).
- No tag, push, or merge.

## Approval needed, stated plainly

| Boundary | Target | Why forbidden by default |
|---|---|---|
| A | `data/figures/` | `data/` is in "Never touch" |
| C | `C:\cgo\memoire.docx` / `.pdf` | a user document; "any user document" is never touch |

The thesis cannot be built honestly without A and C. I will not assume the
page click covered them, because the page's own note asked a question instead
of granting them.
