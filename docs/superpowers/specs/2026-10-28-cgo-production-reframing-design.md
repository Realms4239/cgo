# CGO production-facing reframing and economic reconciliation — design

*Status: approved by direct user instruction · 2026-10-28 · Branch:
`rebuild/plan1-economics-audit` · Slice 13. Scope: current-facing documents and
`book/content/*.js` only. `book/` is never committed.*

---

## 1. Why this slice exists

Slice 12 reconciled the book with `PRODUCT.md` and deliberately kept the
PRODUCT title and the retail-first cost example. The human now instructs a
production-facing reframing drawn from `C:\IDEA`, especially the title, and
more execution of `epbfBBR.md`. This slice adopts the strongest IDEA title,
adds an honest subtitle, and writes the enterprise Model A/B/C and WIS 2.0 SLA
material that the code already implements but the book does not describe.

The memos remain excluded as evidence: their numbers contradict each other,
they publish a hostname/IP that the live-observation spec forbids, and they
claim production measurement that was never performed. Only their title
direction is adopted.

## 2. The title

Main title, from the latest IDEA draft (`Untitled 37`):

> Architecture d’ordonnancement et de régulation de trafic pour la sécurisation
> des flux temps réel du système d’information de Météo Madagascar

Subtitle, written for what the work actually establishes:

> Audit client-side, banc reproductible et comparaison des politiques AQM/BBR
> sur les liens d’accès contraints

Full string used on the cover, the title page, and the metadata:

> Architecture d’ordonnancement et de régulation de trafic pour la sécurisation
> des flux temps réel du système d’information de Météo Madagascar — audit
> client-side, banc reproductible et comparaison des politiques AQM/BBR sur les
> liens d’accès contraints

The main title states the architectural aim; the subtitle states the evidence.
The production limitation paragraph in chapter I remains: no production
measurement, no production control, no server name or IP.

## 3. Sources of truth

| Source | What it settles |
|---|---|
| This spec §2 | The title |
| `PRODUCT.md` | Product description and users |
| `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` §4 | Model A/B/C formulas, tariffs, primary-scenario mapping |
| `pkg/metrics/cost.go` | Implemented constants and labels: GOLD 1 000 000 Ar/200 Mb/s, PLATINE 2 500 000 Ar/30 To, wholesale 81,38 Ar/Go, APN 150 Ar/Go on quote, retail 5 556 Ar/Go emergency, 220 ms / 95 % |
| `epbfBBR.md` §6 | Enterprise framing and WIS 2.0 SLA intent |
| `docs/VERIFY.md` | Claim status; the tariffs are `RECORDED` contractual scenarios, not measurements |

## 4. Book changes

1. **Title page and cover.** Replace the slice-12 title in
   `book/content/00-front.js` and `book/content/00-couverture.xml` with the
   string in §2.
2. **Framing.** One sentence in the introduction and one in the conclusion:
   the architecture acts at the edge, where the DSI has a lever, and is
   validated on a reproducible bench. No production claim.
3. **Enterprise economics.** Add one paragraph in chapter II after equation
   (7): Model A (capacity depreciation), Model B (metered volume), Model C
   (WIS 2.0 SLA), the primary scenario by link type, and the contractual
   disclaimer. Keep equation (7), the ×20 special case, and every tariff number.
4. **Directive status.** Add one paragraph in the limits: eBPF, BBRv3, and
   trace-driven `.dist` remain perspectives because their prerequisites are
   absent; none is claimed as a measurement.
5. **Headers.** Remove the stale « extrait automatiquement … ne pas éditer à la
   main » headers from the content modules; every module is a source of truth.

## 5. Invariants

- No new measurement; no live bench, namespace, qdisc, or listener.
- No production hostname, IP, latency, or control claim.
- Tariffs are labelled contractual scenarios, not measurements.
- Every number added here is in the umbrella spec §4 or `pkg/metrics/cost.go`.
- The slice-12 required strings and Table 15 rows stay.
- `book/` is never staged or committed.
- French typography: typographic apostrophes, « » with inner spaces.

## 6. Acceptance

- `python book/tools/audit-archive.py` → exit 0, both `OK` lines.
- All ten content modules load; `node book/build-docx.js` writes
  `book/memoire-rebuild.docx`; `memoire.docx` is not overwritten.
- `python book/tools/audit-biblio.py` → 47/47 cited.
- The new title appears in `word/document.xml`; the old title does not.
- Numeric-token diff: every lost token listed and justified; added tokens are
  the Model A/B/C values above.
- Body ≤ 15,300 words; total ≤ 17,000.
- `python tools/verify.py --fast` → 5 passed.

## 7. Out of scope

- eBPF, BBRv3, trace-driven `.dist`, RouterOS CHR, or any live acceptance.
- UI redesign; the IDEA executive dashboard is not adopted (`epbfBBR.md` §7
  keeps the existing interface).
- Production access or measurement.
- A new table or annex; the economics is prose only, to preserve Word table
  numbering and the five-annex promise.
