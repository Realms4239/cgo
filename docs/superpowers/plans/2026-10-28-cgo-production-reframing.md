# CGO production-facing reframing and economic reconciliation — implementation plan

*2026-10-28 · Slice 13 · Spec:
`docs/superpowers/specs/2026-10-28-cgo-production-reframing-design.md`*

## Global constraints

- `book/` is gitignored and never staged. `PRODUCT.md` is also gitignored and
  is updated in the working tree only. Only the spec, this plan, the index, the
  ledgers, and the checkpoint are committed.
- No measurement, live bench, namespace, qdisc, listener, hostname, or IP.
- Tariffs are contractual scenarios, not measurements.
- Every added number exists in the umbrella spec §4 or `pkg/metrics/cost.go`.
- Required slice-12 strings and Table 15 rows stay unchanged.

## Task 1 — Title in PRODUCT.md and the book

- `PRODUCT.md`: description becomes « architecture de régulation et
  d’observabilité des liens d’accès contraints… »; thesis line becomes the
  full title string from the spec §2.
- `book/content/00-front.js`: `titlePage.title` becomes the full string.
- `book/content/00-couverture.xml`: replace the slice-12 title (two
  occurrences) with the full string, typographic apostrophes.
- Verify: `grep` finds the new main title once in PRODUCT.md and the book
  title page; the old slice-12 title is absent.

## Task 2 — Framing in the introduction and conclusion

- Introduction: after the economic constraint, add one sentence: the
  architecture acts at the edge, where the DSI keeps a lever, and is validated
  on a reproducible bench.
- Conclusion: replace « Deux instruments y répondent. » with « L’architecture
  y répond par deux instruments. »
- No production claim; the chapter I limitation paragraph stays verbatim.

## Task 3 — Enterprise Model A/B/C in chapter II

Insert one paragraph after the paragraph that follows equation (7), before
Table 7:

- Model A: `(wasted/transported) × (monthly OpEx / 720)`; GOLD 1 000 000 Ar /
  200 Mb/s → 1 388,89 Ar/h.
- Model B: `(wasted bytes / 2³⁰) × tariff Ar/Go × (3600 / window seconds)`;
  the ×20 is the 180-second special case.
- Model C: loaded 16 KiB completions under 220 ms, breach below 95 %; 220 ms is
  the DGM operator objective in the WIS 2.0 frame, not a WMO mandate.
- Primary scenario: fiber → GOLD, 4G/5G → APN 150 Ar/Go, VSAT → wholesale
  81,38 Ar/Go, other → GOLD; retail 5 556 Ar/Go is the emergency overage.
- End with the contractual-scenario disclaimer.
- Verify: module loads; audit green; the paragraph contains `1 388,89`,
  `81,38`, `150`, `95 %`, and `220`.

## Task 4 — Directive status in the limits

Add one paragraph before the perspectives paragraph:

- eBPF needs a committed object and a stable Linux 6.8 host; BBRv3 needs a 7.1
  kernel; trace-driven `.dist` needs a field trace of at least thirty samples.
- All three are perspectives, not corpus features, and none is claimed as a
  measurement.
- Verify: module loads; audit green.

## Task 5 — Clean module headers

Replace the stale three-line extraction header in `015-organisme`,
`03-partie2`, `04-pratique`, `07-refs`, `08-glossaire`, and `09-annexes` with:

```js
// <module> — source de vérité du livre (avec les autres modules content/).
// Le pipeline est : book/content/*.js -> book/build-docx.js -> C:\cgo\memoire.docx.
```

Verify: no `extrait automatiquement` or `Ne pas éditer à la main` remains in
`book/content/*.js`.

## Task 6 — Gates and close

```bash
python book/tools/audit-archive.py
python book/tools/audit-biblio.py
node book/build-docx.js
python tools/verify.py --fast
```

- Word counts: body ≤ 15,300; total ≤ 17,000. Record the new counts.
- Numeric diff against `%TEMP%/cgo-book-numbers-before.json`: list every lost
  token; added tokens are the Model A/B/C values.
- Confirm the new title in `word/document.xml` and no old title; `memoire.docx`
  hash unchanged.
- Append `docs/CHECKPOINTS.md`; add `docs/VERIFY.md` rows: tariffs `RECORDED`,
  slice 13 acceptance `VERIFIED` (builder; independent verifier pending).
- Update the plan index row 13 and section.
- Stage only the spec, the plan, the index, and the two ledgers. Never `book/`
  or `PRODUCT.md` (both gitignored).
