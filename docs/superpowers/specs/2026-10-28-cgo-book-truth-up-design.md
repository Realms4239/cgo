# CGO book truth-up and prose restructure — design

*Status: approved · 2026-10-28 · Branch: `rebuild/plan1-economics-audit`*
*Slice 12. Scope: `book/content/*.js` only. `book/` is never committed.*

---

## 1. Why this slice exists

The book was extracted from `memoire.docx` on 2026-10-19/23. The rebuild of
2026-10-25/28 changed the title, the research questions, the economic framing,
the field-audit claim, and added four instrument capabilities. The book still
describes the earlier state. A prose pass alone would polish claims that are no
longer true.

The slice has two workstreams, in this order:

1. **Truth-up** — make every claim match `PRODUCT.md`, the umbrella spec, and
   `docs/VERIFY.md`.
2. **Restructure and shorten** — Chapter III becomes an evaluation, the manual
   moves to the annexes, and the prose loses 40–45 % of its old body.

The order matters: rewriting prose around a false claim wastes the rewrite.

## 2. Sources of truth

| Source | What it settles |
|---|---|
| `PRODUCT.md` | Thesis title, product truth, users, non-goals |
| `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` §9 | Book harmonization workstream, economic pivot, field-audit rewrite, regime bridge, honesty anchors |
| `docs/VERIFY.md` | Claim status: `VERIFIED`, `RECORDED`, `PENDING`, `BLOCKED`, `REFUTED` |
| `docs/superpowers/specs/2026-10-27-cgo-ifb-ingress-design.md` | IFB ingress scope and its live blocker |
| `docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md` | Live interface chart, pseudonym rule, no production claim before access |
| `docs/superpowers/specs/2026-10-28-cgo-corpus-seal-design.md` | Corpus-wide SHA-256 anchor over the 159 manifests |
| `book/tools/audit-archive.py` | Hard numeric constraints and required strings |

The memos in `C:\IDEA` are **not** a source of truth. They are four drafts of one
argument. Their numbers contradict each other (protected latency 24/25/28/42–48 ms;
bloated latency 1,800/1,850/2,000/3,400/3,800 ms), they publish a hostname and IP
that the live-observation spec forbids publishing, and they claim production
measurement that the project has not performed. None of their numbers, names, or
titles may enter the book.

## 3. The title

`PRODUCT.md:7-9` carries the current title. Use it verbatim on the title page,
with the book's typographic apostrophes:

> Conception et évaluation d’un banc reproductible pour l’audit client-side et la
> comparaison des politiques AQM/BBR sur liens d’accès contraints — application
> aux liens de Météo Madagascar

No other title variant survives. The running header and the metadata follow the
same string.

## 4. The research questions

These replace Table 5 and every RQ reference in the introduction, the chapter
answers, the conclusion, and the glossary.

**RQ1 — Audit client-side.**

> Quels symptômes de dégradation un audit non intrusif peut-il établir sur un
> lien d’accès que l’équipe ne contrôle pas, et quelles limites cette position
> impose-t-elle à l’interprétation ?

**RQ2 — Bench and comparison.**

> À quelles conditions un banc reproductible reproduit-il le mécanisme du
> bufferbloat assez fidèlement pour départager pfifo_fast, fq_codel et CAKE sous
> CUBIC et BBR, et que montrent les mesures sur les quatre profils ?

**RQ3 — Recommendation and cost.**

> Quelle configuration recommander aux équipements sous contrôle de la direction
> des systèmes d’information, par classe de lien, et quel coût en ariary la
> protection représente-t-elle ?

The answers must match the evidence:

- **RQ1** — the protocol is implemented as three operator presets; pilots ran on
  the bench and the LAN; no field campaign was executed. Annex C is a startup
  guide, not field data.
- **RQ2** — the frozen corpus is a 1-flow legacy bench. The current bench is
  4-flow dual-ns. The regime-bridge experiment was not run; the book says so.
- **RQ3** — enterprise CIR/depreciation is the primary scenario; retail prepaid
  is the emergency overage. P3 and P4 remain exploratory bounds.

## 5. Claims to correct

| # | Book says | Current truth | Resolution |
|---|---|---|---|
| 1 | Title: *banc d’essai reproductible… comparaison de politiques AQM et BBR…* | `PRODUCT.md` title above | Rewrite title page |
| 2 | "Le protocole d’audit terrain a été éprouvé sur trois sites, consignés en annexe C" | No field audit executed; Annex C is the instrument guide | Rewrite: protocol implemented, pilots bench/LAN, field not executed |
| 3 | "Les campagnes se sont déroulées… Fibre, 4G et satellite y figurent" | The corpus is a bench corpus on four profiles | Separate bench profiles from field conditions |
| 4 | Cost default: 4.5 GB retail tier at 5,556 Ar/GB | Enterprise CIR is the primary scenario; retail is overage | Rewrite the cost framing, keep the retail comparison |
| 5 | English abstract: "eight quality gates" | French and corpus: seven measured gates (G0–G6; G7 not instrumented) | Fix |
| 6 | UI: four views, five tabs, no interface chart | Live view has a raw RX/TX interface chart (slice 10) | Add to Annex E, labelled twin |
| 7 | Instrument: egress only | IFB ingress shaping exists (slice 6); Windows engine verified, live blocked | Add one sentence and one limit |
| 8 | Reproducibility: per-run manifests | Corpus-wide SHA-256 anchor over 159 manifests (slice 11) | Add one paragraph |
| 9 | No regime label | Frozen corpus is 1-flow legacy; current bench is 4-flow dual-ns; bridge not run | State the regime; no bridge claim |
| 10 | No mention of eBPF, BBRv3, `.dist` | All blocked or deferred | Perspectives only, never features |

## 6. Chapter III

Title: **CHAPITRE III : MESURES ET ÉVALUATION**.

Order and targets:

| Section | Words |
|---|---|
| Opening: the campaign and the central comparison | 250 |
| Un binaire unique, sans dépendance | 450 |
| Conduite des campagnes | 900 |
| Chaque ligne gelée porte son empreinte | 300 |
| Résultats | 1,800 |
| Interprétation | 1,100 |
| Recommandations | 700 |
| Limites | 350 |

The 1,881-word UI walkthrough moves to **Annexe E**, compressed to ~900 words.
The installer and build detail moves to **Annexe C**, compressed. No evidence is
deleted; it changes address.

The opening is written from Table 15 only:

> Ce chapitre rend compte de la campagne finale. Cent cinquante-neuf runs ont été
> archivés, soit 528 lignes d’événements ; 208 sont valides au sens des sept
> portes mesurées (G7 non instrumentée sur le gel), 74 dégradées et 246
> invalidées. Quatre profils de lien ont été couverts : fibre, cellulaire,
> satellite géostationnaire et orbite basse. Le résultat tient en une
> comparaison. Sur le profil cellulaire P2, à l’échéance de 220 ms, la file
> simple associée à BBR tient 60,2 % des requêtes critiques ; CAKE en tient
> 98,1 %, fq_codel 96,2 %, pour un débit de masse conservé entre 18,2 et
> 19,5 Mb/s. Le small p95 passe de 328,2 ms à 210,5 ms, et l’écart p95-médiane
> de 57,4 ms à 14,0 ms. Ces écarts fondent la recommandation opérationnelle du
> chapitre.

Every value above is a Table 15 value. The full table stays in *Résultats*; the
opening does not duplicate it.

## 7. Prose rules

1. Delete meta-discourse: sentences that announce what the text will do or has
   just done.
2. Delete the closing sentence of a paragraph when it only restates it.
3. Split any sentence over ~35 words at its semicolon or relative clause.
4. Cap paragraphs at ~120 words; one idea per paragraph.
5. Replace nominalizations with the act.
6. Cut triads to the two items that carry information.
7. Keep the specific: dates, names, failures, the 746 ms fibre outlier, the
   12,486 Ar/h CAKE anomaly.

Target: **body 23,343 → ~14,500 words**, including ~1,100 words of new truth-up
material, so about **−44 % of the old prose**. Annexes ~1,900. Total ~16,400,
down from 24,373.

## 8. The production limitation

One paragraph in Chapter I, `Du service au mémoire`:

> La mesure sur le serveur de production lui-même n’entre pas dans ce mémoire.
> L’accès en lecture seule aux compteurs de son interface n’était pas disponible
> pendant les campagnes ; le banc installé sur un poste client en tient lieu. Les
> chiffres qui suivent sont donc des chiffres de banc, produits sur des liens
> reconstitués, et non des mesures du serveur en exploitation.

No server name, no IP, no production number. The live-observation spec's
pseudonym rule stays satisfied because the paragraph names nothing.

## 9. Invariants

- Every number, table, equation, citation, and required audit string stays.
- Required strings: `159 runs`, `528`, `208 valides`, `74 degraded`, `246 invalid`,
  `G7`, `non instrumentée`, `sept portes`.
- Required values: `324,25`, `209,25`, `210,3`, `51,75`, `45,75`, `46,15`; the six
  Table 15 rows.
- Forbidden strings: `160 runs`, `Passent les huit portes`, `au sens strict des
  huit portes`, `franchit les huit portes`, `Huit portes de qualité contrôlent`.
- French typography: typographic apostrophes, non-breaking spaces before
  `: ; ? !`, « » with inner spaces.
- No new number unless it is in `VERIFY.md` or the frozen corpus.
- No production claim; no hostname or IP.
- `book/data/runs` untouched.

## 10. Acceptance

- `python book/tools/audit-archive.py` → exit 0, both `OK` lines.
- All ten content modules load under `node`.
- `node book/build-docx.js` → `book/memoire-rebuild.docx`; `memoire.docx` is not
  overwritten.
- Numeric-token diff before/after: every disappeared number is listed and
  justified.
- Claims table: every new or changed claim maps to a `VERIFY.md` row or carries a
  label (`measured`, `recorded`, `blocked`).
- Word count: body ≤ 15,000; total ≤ 17,000.
- `git status --short` shows no `book/` path staged.

## 11. Out of scope

- New measurements, campaigns, or corpus changes.
- Production access, hostname/IP, or production latency.
- A title other than `PRODUCT.md`.
- A UI redesign; the book describes the interface the code has.
- The memos' numbers, titles, or one-screen dashboard.
- Supervisor, jury, date, and Master/Licence placeholders; these need the
  human's data.

## 12. Risks

| Risk | Mitigation |
|---|---|
| Cutting a method nuance the jury expects | Invariants list; second read of Chapter II before Chapter III results |
| Adding a feature claim the corpus cannot support | Every added feature is `VERIFIED` in `VERIFY.md` or explicitly labelled blocked |
| Losing a number silently | Numeric-token diff, reviewed line by line |
| The RQ rewrite leaving a stale answer somewhere | Grep every `RQ1`/`RQ2`/`RQ3` reference after the edit |
| The build overwriting the deliverable | Run `node book/build-docx.js` only, never `build-book.ps1` |
