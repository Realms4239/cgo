# Thesis identity — title and concrete project (2026-10-31)

*Status: decision, new phase. Replaces the Option A title of 2026-10-30.*

## 1. The finding

Read against the tree, the current title claims three things the project cannot
defend:

- **« garantie d'échéances »** — the corpus shows 60.2 % of deadlines met under
  the untuned queue; a “guarantee” is refuted by the project's own table.
- **« application au système d'information de Météo Madagascar »** — the case
  study is BLOCKED on the operator's grant; nothing of that infrastructure was
  read, loaded, or changed.
- **BBR** adjacent to “garantie” — the frozen corpus holds BBRv1 only; BBRv3 is
  BLOCKED on a 7.1 kernel.

The IDEA titles are worse on the same axis (“sécurisation”, “priorisation”,
one-click control); the master prompt's own title is the only one whose nouns
are all reproducible.

## 2. Decision — the title becomes

> **Conception et évaluation d'un banc d'essai reproductible pour la
> comparaison de politiques AQM et BBR sur des liens d'accès contraints**

> **Le cas des liens d'accès du service national de prévision météorologique
> de Madagascar**

Main title = the master prompt's wording verbatim: it names what was built
(bench), what was compared (AQM × BBR), and the constraint (access links). No
verb of production, no guarantee, no deployment.

Subtitle = the institution as **case**, not as claim. The book's Chapter I does
document the institution from public facts; the twin is calibrated on the same
link classes. Nothing in the subtitle asserts that the operator's infrastructure
was measured, secured or modified.

Rejected: “garantie” (refuted by 60.2 %), “sécurisation/priorisation” (no
production read), “observatoire” (implies production observation the case study
never obtained).

## 3. Consequences to apply in the same pass

1. Title strings (PRODUCT.md, title page, cover XML, running header, document
   metadata) replaced verbatim; the previous title absent from the rendered
   book.
2. **Abstract numeric fix**: the abstract must quote the corpus values, not a
   transcription drift — 328.2 → 210.5 ms and 57.4 → 14.0 ms QDI (methodology
   §5), never 324.3/210.3.
3. The word “garantie” is removed from the thesis identity everywhere it
   describes the project's outcome; where the 220 ms objective is discussed, it
   stays an operator objective evaluated on the bench.
4. The concrete-project paragraph (synthesis §5) enters the introduction as the
   one-sentence project definition: audit client-side + sealed reports + twin
   bench + reviewable configs, with the four explicit non-claims.
5. Rebuild the book, run the five audits, confirm `memoire.docx` unchanged.

## 4. What this does not change

No measurement, no corpus row, no gate, no BLOCKED row. The adjustment is
identity and copy; the evidence ledger is untouched.

## 5. Acceptance

- `grep -R "garantie d" book/content PRODUCT.md` returns only passages that
  explicitly discuss the rejected claim (none in the title/abstract).
- The new title is in `word/document.xml`; the old one is absent.
- The abstract's millisecond values match the corpus-rendered table.
- `audit-archive`, `audit-nombres`, `audit-renvois`, `audit-sigles`,
  `audit-biblio` green; `memoire.docx` SHA-256 unchanged.
