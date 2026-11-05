# CGO book truth-up and prose restructure — implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the thesis book match the current repository (title, RQs, claims, features), then restructure Chapter III and cut the prose by ~44 % of its old body without losing a number, table, equation, or citation.

**Architecture:** Ten CommonJS content modules in `book/content/` are the book's source of truth. The audit script `book/tools/audit-archive.py` checks the frozen corpus, the aggregates, the required strings, and Table 15. `book/build-docx.js` renders `book/memoire-rebuild.docx`. The work is prose and structure only; no Go, no corpus, no frontend.

**Tech Stack:** Node.js (CommonJS modules), Python 3.11, the existing `book/` pipeline. No new dependency.

## Global Constraints

- `book/` is gitignored and **never staged or committed**. The only commits in this slice are the spec, this plan, the plan index, and the final checkpoint.
- Every number, table, equation, citation, and required audit string stays. Required strings: `159 runs`, `528`, `208 valides`, `74 degraded`, `246 invalid`, `G7`, `non instrumentée`, `sept portes`.
- Required values: `324,25`, `209,25`, `210,3`, `51,75`, `45,75`, `46,15`; the six Table 15 rows.
- Forbidden strings: `160 runs`, `Passent les huit portes`, `au sens strict des huit portes`, `franchit les huit portes`, `Huit portes de qualité contrôlent`.
- No new number unless it is in `docs/VERIFY.md` or the frozen corpus.
- No production claim, hostname, or IP. The production site is unnamed.
- French typography: typographic apostrophes (`’`), non-breaking spaces before `: ; ? !`, « » with inner spaces.
- Run `node book/build-docx.js` only. Never run `book/tools/build-book.ps1`; it copies over `C:\cgo\memoire.docx`.
- After every content task: load the edited module, then run the audit.

## Files

| File | Responsibility |
|---|---|
| `book/content/00-front.js` | Title page, remerciements, résumé, abstract, sigles |
| `book/content/01-intro.js` | Introduction |
| `book/content/015-organisme.js` | Chapter I, institution and field |
| `book/content/02-partie1.js` | Chapter I, links, tariffs, RQs, objectives |
| `book/content/03-partie2.js` | Chapter II, state of the art, method, equations, protocol, instrument |
| `book/content/04-pratique.js` | Chapter III, instrument, campaigns, results, interpretation, recommendations, limits |
| `book/content/06-conclusion.js` | Conclusion |
| `book/content/07-refs.js` | References (no prose) |
| `book/content/08-glossaire.js` | Glossary |
| `book/content/09-annexes.js` | Annexes A–E |
| `book/tools/audit-archive.py` | Hard checks (read-only) |
| `book/build-docx.js` | Renderer (read-only) |

---

### Task 1: Baseline snapshot

**Files:**
- Read: `book/content/*.js`, `docs/VERIFY.md`, `PRODUCT.md`
- Write: `%TEMP%/cgo-book-numbers-before.json` (outside the repo)

**Interfaces:**
- Produces: the before word counts and the before numeric-token set that Task 12 compares against.

- [ ] **Step 1: Record the word counts**

Run:
```bash
node -e "const m=['00-front','01-intro','015-organisme','02-partie1','03-partie2','04-pratique','06-conclusion','07-refs','08-glossaire','09-annexes'];let T=0;for(const x of m){let w=0;for(const e of require('./book/content/'+x+'.js'))if(e.p)w+=e.p.split(/\s+/).length;T+=w;console.log(x.padEnd(14),w)}console.log('TOTAL',T)"
```
Expected: `00-front 848`, `01-intro 985`, `015-organisme 2013`, `02-partie1 2693`, `03-partie2 6871`, `04-pratique 9155`, `06-conclusion 778`, `07-refs 0`, `08-glossaire 0`, `09-annexes 1030`, `TOTAL 24373`.

- [ ] **Step 2: Snapshot every numeric token per module**

Run:
```bash
node -e "const fs=require('fs'),p='book/content';const o={};for(const f of fs.readdirSync(p).filter(x=>x.endsWith('.js'))){const t=fs.readFileSync(p+'/'+f,'utf8');o[f]=[...new Set(t.match(/\d+(?:[.,]\d+)?/g)||[])].sort()}fs.writeFileSync(process.env.TEMP+'/cgo-book-numbers-before.json',JSON.stringify(o,null,1))"
```
Expected: the file exists; no output.

- [ ] **Step 3: Confirm the audit is green before touching anything**

Run:
```bash
python book/tools/audit-archive.py
```
Expected: exit 0, both lines:
```
OK — corpus 159 runs, 528 lignes, 208 valides, 74 degraded, 246 invalid
OK — 22 cellules, 159 manifestes, inventaire, texte et tableau 15 cohérents
```

---

### Task 2: Front matter — title and English abstract

**Files:**
- Modify: `book/content/00-front.js:7` (title), `:23-24` (English abstract)

**Interfaces:**
- Consumes: the title from `PRODUCT.md:7-9`; the French abstract's `sept portes mesurées (G0–G6; G7 non instrumentée sur le gel)`.
- Produces: the title string every later task assumes.

- [ ] **Step 1: Replace the title**

In `00-front.js`, replace the `title:` value with:
```
Conception et évaluation d’un banc reproductible pour l’audit client-side et la comparaison des politiques AQM/BBR sur liens d’accès contraints — application aux liens de Météo Madagascar
```

- [ ] **Step 2: Fix the English abstract's gate count**

Replace `submits every event to eight quality gates before freezing the data` with:
```
submits every event to seven measured quality gates (G0–G6; G7 defined but not instrumented on the freeze) before freezing the data
```
Replace `208 valid under strict valid-only filtering` with:
```
208 valid under the seven measured gates
```

- [ ] **Step 3: Verify the module loads and the old claim is gone**

Run:
```bash
node -e "require('./book/content/00-front.js')" && grep -c "eight quality gates" book/content/00-front.js; grep -c "banc reproductible pour l’audit client-side" book/content/00-front.js
```
Expected: the require exits 0; the first grep prints `0`; the second prints `1`.

- [ ] **Step 4: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 3: Introduction — RQs, field claim, bench versus field

**Files:**
- Modify: `book/content/01-intro.js` (whole file, 8 paragraphs, 985 words → ~700)

**Interfaces:**
- Consumes: the three RQ wordings from the spec §4.
- Produces: the introduction's RQ sentence that Chapter I's Table 5 must match.

- [ ] **Step 1: Rewrite the research-question paragraph**

Replace the paragraph that begins `La question de recherche se formule en une phrase` with a version that names the three questions in one sentence each, using the spec §4 wording, and keeps the two-volet sentence (audit then bench). Target 90 words.

- [ ] **Step 2: Remove the field overclaim**

Replace `Les campagnes de mesure se sont déroulées dans les conditions réelles de l’institution.` with:
```
Les campagnes ont subi les conditions matérielles du stage : coupures d’alimentation, délestages et arrêts de poste involontaires.
```
Keep the sentence about resuming without losing a frozen line. Delete any implication that the bench measured real fibre, 4G, or satellite links.

- [ ] **Step 3: Keep the chapter roadmap, shorten it**

Keep the last paragraph's promise that Chapter III confronts the instrument with the real, and that the annexes hold the interface captures. Cut the list of every annex to one clause.

- [ ] **Step 4: Apply the prose rules**

Delete meta-discourse, restating closes, and triads. Split sentences over ~35 words. Cap paragraphs at ~120 words.

- [ ] **Step 5: Verify**

Run:
```bash
node -e "require('./book/content/01-intro.js')" && node -e "const m=require('./book/content/01-intro.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; the count is ≤ 750.

- [ ] **Step 6: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 4: Chapter I, institution — remove the three sites, add the limitation

**Files:**
- Modify: `book/content/015-organisme.js` (2,013 words → ~1,300)

**Interfaces:**
- Consumes: the limitation paragraph from the spec §8; the three-preset truth from `pkg/audit/presets.go`.
- Produces: Chapter I's honest field status, which the conclusion and Annex C reuse.

- [ ] **Step 1: Replace the three-sites claim**

Replace the paragraph that begins `Les commandes de l’instrument mobilisées pendant le stage` and contains `Le protocole d’audit terrain a été éprouvé sur trois sites` with:
```
Les commandes de l’instrument mobilisées pendant le stage sont celles du {tabRef:tbl402}. `cgo setup` conduit l’installation guidée, `cgo doctor` diagnostique l’environnement, `cgo kit` déploie la machine du banc, `cgo run` lance les campagnes, `cgo tui` supervise en terminal, `cgo audit` mesure un lien accessible, `cgo verify` recalcule les empreintes et `cgo figures` régénère les graphiques depuis les archives. Le protocole d’audit terrain est implémenté sous la forme de trois préréglages opérateur. Il a été éprouvé sur le banc et sur le réseau local du stage ; aucune campagne de terrain n’a été exécutée. Les profils du banc reprennent les quatre classes de liens décrites au chapitre I, sans prétendre mesurer un site réel.
```

- [ ] **Step 2: Separate the bench from the field**

Replace the paragraph that begins `Les campagnes se sont déroulées dans les conditions réelles de l’institution` with:
```
Les campagnes ont subi les conditions matérielles du stage : coupures d’alimentation, délestages et arrêts de poste involontaires. Chaque incident a été repris sans perdre une ligne déjà gelée, ce qui a fait de la reprise après interruption une propriété constatée. Les quatre profils du banc couvrent la fibre, la 4G, le satellite géostationnaire et l’orbite basse ; ils reconstituent ces liens, ils ne les mesurent pas.
```

- [ ] **Step 3: Add the production limitation**

At the end of the section `Du service au mémoire : besoins formulés et réponses apportées`, insert the paragraph from the spec §8 verbatim.

- [ ] **Step 4: Apply the prose rules and cut repetition**

The DEM description is repeated across the chapter; keep the first full description and reduce later mentions to a clause. Target 1,300 words.

- [ ] **Step 5: Verify**

Run:
```bash
node -e "require('./book/content/015-organisme.js')" && grep -c "trois sites" book/content/015-organisme.js; node -e "const m=require('./book/content/015-organisme.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; `trois sites` prints `0`; the count is ≤ 1,350.

- [ ] **Step 6: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 5: Chapter I, links and RQs — new Table 5, economic framing

**Files:**
- Modify: `book/content/02-partie1.js` (2,693 words → ~1,600)

**Interfaces:**
- Consumes: the RQ wordings from the spec §4; the tariff data already in the file.
- Produces: the RQ identifiers and wording that Tasks 8 and 9 must match.

- [ ] **Step 1: Replace the three RQ rows**

In the table `tbl285` (Tableau 5), replace the three `rows` with:

```
["RQ1", "Quels symptômes de dégradation un audit non intrusif peut-il établir sur un lien d’accès que l’équipe ne contrôle pas, et quelles limites cette position impose-t-elle à l’interprétation ?", "Audit non intrusif depuis un poste client ; trois préréglages ; pilotes sur banc et réseau local"],
["RQ2", "À quelles conditions un banc reproductible reproduit-il le mécanisme du bufferbloat assez fidèlement pour départager pfifo_fast, fq_codel et CAKE sous CUBIC et BBR, et que montrent les mesures sur les quatre profils ?", "Banc reproductible, matrice profil × discipline × contrôle, portes de qualité, corpus gelé"],
["RQ3", "Quelle configuration recommander aux équipements sous contrôle de la direction des systèmes d’information, par classe de lien, et quel coût en ariary la protection représente-t-elle ?", "Interprétation des mesures, tarifs entreprise et prépayés, équipements maîtrisés"]
```

- [ ] **Step 2: Rewrite the problématique paragraph to match RQ1**

The paragraph that begins `La problématique s’énonce alors ainsi` must ask what a client-side audit can establish, not whether the links are degraded. Keep the `sans modifier l’infrastructure de production` clause.

- [ ] **Step 3: Demote the retail tier**

In the tariff section, keep the 5,556 Ar/GB comparison and the inversion, but make the enterprise CIR/depreciation the primary reading. Replace any sentence that presents the 4.5 GB retail tier as the default with one that presents it as the emergency overage. Do not change a single tariff number.

- [ ] **Step 4: Apply the prose rules**

Cut the operator-by-operator catalogue to the two facts that change the cost: the facial price is equal, the effective price per GB is not. Target 1,600 words.

- [ ] **Step 5: Verify**

Run:
```bash
node -e "require('./book/content/02-partie1.js')" && grep -c "Quels symptômes de dégradation" book/content/02-partie1.js; node -e "const m=require('./book/content/02-partie1.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; the grep prints `1`; the count is ≤ 1,700.

- [ ] **Step 6: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 6: Chapter II, method — cost, regime, seal, ingress, status

**Files:**
- Modify: `book/content/03-partie2.js` (6,871 words → ~3,900)

**Interfaces:**
- Consumes: `docs/VERIFY.md` rows for the seal, the IFB engine, and the blocked live acceptance; `pkg/qdisc/ingress.go`, `pkg/campagne/dualns.go`, `pkg/qdisc/ge.go`.
- Produces: the method vocabulary (regime, seal, ingress, bursty) that Chapter III and the annexes reuse.

- [ ] **Step 1: State the regime**

In `Protocole des campagnes de mesure`, add one paragraph: the frozen corpus was produced on the legacy 1-flow bench; the current bench is 4-flow dual-ns; the bridge experiment between the two regimes was not run, so the book reports the frozen regime and does not translate its absolutes to the new one.

- [ ] **Step 2: Add the corpus seal**

In `Empreinte de preuve`, add one paragraph: the per-run manifests are themselves anchored by `docs/frozen-manifests.sha256`, 159 lines, checked by the `frozen` gate; a manifest rewritten with its files is caught. Label it `VERIFIED` in prose only if the sentence names the command; otherwise say "vérifié par la porte `frozen`".

- [ ] **Step 3: Add ingress shaping, labelled**

In `Architecture logicielle de l’instrument` or the remediation passage, add one sentence: the instrument can shape ingress through an IFB, with a RouterOS LAN queue tree for the DSI's equipment; the command generation is verified on Windows, the Linux 6.8 live acceptance is blocked. Do not claim a live ingress result.

- [ ] **Step 4: Add the bursty channel as a perspective**

In the same section or in the limits, add one sentence: the bench implements a Gilbert-Elliott bursty channel (`gemodel`) for P2/P3, but the frozen corpus was produced on the clean channel; no bursty result is claimed.

- [ ] **Step 5: Demote retail in the cost equation section**

Keep the formula, the 20× extrapolation, and every tariff number. Change the default-tier sentence so the enterprise CIR is the primary scenario and the 4.5 GB retail tier is the overage.

- [ ] **Step 6: Apply the prose rules and cut the textbook passages**

The state-of-the-art section loses the history that the citations already carry. Keep the mechanism sentences that the results use (G4 Mathis bound, BBR fill behaviour, CAKE isolation). Keep every equation and every symbol definition. Target 3,900 words.

- [ ] **Step 7: Verify**

Run:
```bash
node -e "require('./book/content/03-partie2.js')" && node -e "const m=require('./book/content/03-partie2.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; the count is ≤ 4,100.

- [ ] **Step 8: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 7: Chapter III — retitle, open with the result, move the manual

**Files:**
- Modify: `book/content/04-pratique.js` (9,155 words → ~5,300 after the move and the cut)
- Modify: `book/content/09-annexes.js` (Annex C and E receive the moved material)

**Interfaces:**
- Consumes: the Chapter III opening from the spec §6; the UI block in `04-pratique.js` from `{ h2: "Interface de supervision du poste" }` to the end of `Tableau terminal, onglet par onglet`.
- Produces: a Chapter III that opens with the campaign and keeps the full Table 15 in *Résultats*; an Annex E with the UI prose and its five figures.

- [ ] **Step 1: Retitle the chapter**

Replace `{ h1: "CHAPITRE III : CONCEPTION DU BANC D’ESSAI" }` with:
```
{ h1: "CHAPITRE III : MESURES ET ÉVALUATION" }
```

- [ ] **Step 2: Replace the opening paragraph**

Replace the paragraph that begins `Ce chapitre décrit la réalisation du banc` with the opening from the spec §6, verbatim.

- [ ] **Step 3: Move the UI walkthrough to Annex E**

Cut every element from `{ h2: "Interface de supervision du poste" }` through the last paragraph of `Tableau terminal, onglet par onglet`, stopping before `{ h2: "Conduite des campagnes de mesure" }`. Paste it into `09-annexes.js` under Annex E. Compress to ~900 words: one paragraph per view (Campagne, Live, Résultats, Provenance), one paragraph for the TUI, and the five existing figure entries. Add the live interface chart to the Live paragraph, labelled `twin` only.

- [ ] **Step 4: Move the installer detail to Annex C**

Cut the seven-step installer, the two-stage build, and the deployment-engine detail from `Instrument logiciel` and paste it into Annex C, compressed. Keep one paragraph in Chapter III: the single binary, no dependency, the build, the deployment engine.

- [ ] **Step 5: Update the annex count and titles**

Annex E's title becomes `Annexe E. L’interface de supervision, écran par écran`. The introduction's "cinq annexes" stays five.

- [ ] **Step 6: Verify the move**

Run:
```bash
node -e "require('./book/content/04-pratique.js')" && node -e "require('./book/content/09-annexes.js')"
grep -c "Vue Campagne\|Vue Live\|Vue Résultats\|Vue Provenance" book/content/04-pratique.js
grep -c "Vue Campagne\|Vue Live\|Vue Résultats\|Vue Provenance" book/content/09-annexes.js
```
Expected: both modules load; the first grep prints `0`; the second prints a number ≥ `1`.

- [ ] **Step 7: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 8: Chapter III — results, interpretation, recommendations, limits

**Files:**
- Modify: `book/content/04-pratique.js`

**Interfaces:**
- Consumes: Table 15 and the audit's required strings; the RQ wordings from Task 5.
- Produces: the RQ answers that Task 9's conclusion summarizes.

- [ ] **Step 1: Rewrite the headings as claims**

Apply at least these:
- `Campagnes menées et périmètre couvert` → `159 runs, 528 lignes, 208 valides`
- `Résultat central : respect d’échéance sous charge` → `Sous BBR, la file simple tient 60 % des échéances, CAKE 98 %`
- `Gel des événements et traçabilité des archives` → `Chaque ligne gelée porte son empreinte`
- `Instrument logiciel et architecture déployée` → `Un binaire unique, sans dépendance`
- `Réponse à RQ1` → `RQ1 : l’audit mesure sans modifier le lien`
- `Réponse à RQ2 : les disciplines actives protègent le trafic critique` → `RQ2 : les disciplines actives protègent le trafic critique`
- `Réponse à RQ3 : la prescription et son coût` → `RQ3 : la prescription et son coût`

Keep labels for `Limites et perspectives` and the annex headings.

- [ ] **Step 2: Rewrite the RQ answers to the new questions**

RQ1's answer states what the audit can and cannot establish; it names the three presets and the bench/LAN pilots, and does not claim a field campaign. RQ2's answer states the regime and the comparison. RQ3's answer states the recommendation per link class and the cost.

- [ ] **Step 3: Apply the prose rules**

Delete `La lecture de ce tableau constitue le cœur du mémoire` and every equivalent. Delete restating closes. Split the 273-word paragraph. Keep the 746 ms outlier, the 12,486 Ar/h anomaly, and the G4 mechanism. Target 5,300 words for the whole chapter after the move.

- [ ] **Step 4: Verify the numbers**

Run:
```bash
node -e "require('./book/content/04-pratique.js')" && grep -c "159 runs\|528\|208 valides\|74 degraded\|246 invalid" book/content/04-pratique.js
```
Expected: exit 0; the grep prints at least `1` for each required string present in this module.

- [ ] **Step 5: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 9: Conclusion — answer the three new RQs

**Files:**
- Modify: `book/content/06-conclusion.js` (778 words → ~500)

**Interfaces:**
- Consumes: the RQ answers from Task 8; the economic framing from Task 6.
- Produces: the final RQ summary that the reader remembers.

- [ ] **Step 1: Replace the RQ answers**

Each answer is one paragraph, one claim, one number. RQ1: the audit measures what a client can see, and its limits. RQ2: the bench separates the policies under the frozen regime, and the regime is stated. RQ3: the recommendation per class, with the enterprise cost and the retail overage.

- [ ] **Step 2: Remove the duplicated result narration**

The conclusion must not re-copy the Chapter III tables. Keep one number per RQ answer.

- [ ] **Step 3: Apply the prose rules and target 500 words**

- [ ] **Step 4: Verify**

Run:
```bash
node -e "require('./book/content/06-conclusion.js')" && node -e "const m=require('./book/content/06-conclusion.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; the count is ≤ 550.

- [ ] **Step 5: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 10: Annexes — C and E absorb the manual, add the seal

**Files:**
- Modify: `book/content/09-annexes.js` (1,030 words → ~1,900)

**Interfaces:**
- Consumes: the moved UI and installer material from Task 7; the seal paragraph from Task 6.
- Produces: the annex set the introduction promises (five annexes).

- [ ] **Step 1: Rewrite Annex C**

Annex C keeps the startup guide, gains the compressed installer/build detail, and states the field-audit status: three presets implemented, pilots on bench/LAN, no field campaign. Remove any pointer that presents Annex C as field data.

- [ ] **Step 2: Rewrite Annex E**

Annex E carries the five figures and the compressed UI prose: one paragraph per view, one for the TUI, the interface chart labelled `twin`, and the provenance view's seal. No new number.

- [ ] **Step 3: Add the seal to Annex A or B**

If the seal fits the artifact description, add one sentence to Annex A: the manifests are anchored by `docs/frozen-manifests.sha256` and checked by the `frozen` gate. Otherwise leave it in Chapter II only.

- [ ] **Step 4: Verify**

Run:
```bash
node -e "require('./book/content/09-annexes.js')" && node -e "const m=require('./book/content/09-annexes.js');let w=0;for(const e of m)if(e.p)w+=e.p.split(/\s+/).length;console.log(w)"
```
Expected: exit 0; the count is between 1,700 and 2,100.

- [ ] **Step 5: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 11: Glossary and sigles — new terms, new RQs

**Files:**
- Modify: `book/content/08-glossaire.js`
- Modify: `book/content/00-front.js` (sigles block)

**Interfaces:**
- Consumes: the vocabulary introduced in Tasks 6–8.
- Produces: no undefined term in the body.

- [ ] **Step 1: Add the new entries**

Add, in alphabetical order, entries for:
- `Meteolink` — nom produit de l'instrument CGO.
- `IFB` — Intermediate Functional Block, interface virtuelle pour la mise en forme entrante.
- `GEModel` — modèle de perte en rafales Gilbert-Elliott de `netem`.
- `Ancre de corpus` — empreinte SHA-256 des 159 manifestes, vérifiée par la porte `frozen`.

- [ ] **Step 2: Update the RQ entries**

The sigles block's `RQ1 à RQ3` entry must match the new questions' topics: audit, bench, recommendation.

- [ ] **Step 3: Verify**

Run:
```bash
node -e "require('./book/content/08-glossaire.js')" && node -e "require('./book/content/00-front.js')"
```
Expected: both exit 0.

- [ ] **Step 4: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines.

---

### Task 12: Final verification and checkpoint

**Files:**
- Read: all ten modules
- Write: `docs/CHECKPOINTS.md` (append)
- Write: `%TEMP%/cgo-book-numbers-after.json` (outside the repo)

**Interfaces:**
- Consumes: the before snapshot from Task 1; the audit's two `OK` lines.
- Produces: the slice's evidence for the checkpoint.

- [ ] **Step 1: Run the audit**

Run: `python book/tools/audit-archive.py`
Expected: exit 0, both `OK` lines verbatim.

- [ ] **Step 2: Load every module**

Run:
```bash
for m in 00-front 01-intro 015-organisme 02-partie1 03-partie2 04-pratique 06-conclusion 07-refs 08-glossaire 09-annexes; do node -e "require('./book/content/$m.js')" || exit 1; done
```
Expected: exit 0, no output.

- [ ] **Step 3: Build the reconstruction**

Run: `node book/build-docx.js`
Expected: exit 0; `book/memoire-rebuild.docx` is written; `C:\cgo\memoire.docx` is untouched.

- [ ] **Step 4: Compare the numeric tokens**

Run:
```bash
node -e "const fs=require('fs'),p='book/content';const o={};for(const f of fs.readdirSync(p).filter(x=>x.endsWith('.js'))){const t=fs.readFileSync(p+'/'+f,'utf8');o[f]=[...new Set(t.match(/\d+(?:[.,]\d+)?/g)||[])].sort()}fs.writeFileSync(process.env.TEMP+'/cgo-book-numbers-after.json',JSON.stringify(o,null,1))"
node -e "const fs=require('fs');const a=JSON.parse(fs.readFileSync(process.env.TEMP+'/cgo-book-numbers-before.json'));const b=JSON.parse(fs.readFileSync(process.env.TEMP+'/cgo-book-numbers-after.json'));for(const f of Object.keys(a)){const lost=a[f].filter(x=>!b[f].includes(x));if(lost.length)console.log(f,'LOST:',lost.join(' '))}"
```
Expected: every `LOST` token is either a duplicate removed with a deleted paragraph or a number deliberately replaced; list each in the checkpoint. No unexplained loss.

- [ ] **Step 5: Record the final word counts**

Run:
```bash
node -e "const m=['00-front','01-intro','015-organisme','02-partie1','03-partie2','04-pratique','06-conclusion','07-refs','08-glossaire','09-annexes'];let T=0;for(const x of m){let w=0;for(const e of require('./book/content/'+x+'.js'))if(e.p)w+=e.p.split(/\s+/).length;T+=w;console.log(x.padEnd(14),w)}console.log('TOTAL',T)"
```
Expected: body (00–06) ≤ 15,000; total ≤ 17,000.

- [ ] **Step 6: Verify the claims table**

Open the spec §5 and check each row against the edited book. Every row must be resolved. Record any unresolved row in the checkpoint as a blocker.

- [ ] **Step 7: Confirm nothing in `book/` is staged**

Run: `git status --short`
Expected: `M web/frontend/dist/index.html` and the pre-existing untracked files only; no `book/` path staged.

- [ ] **Step 8: Append the checkpoint**

Append to `docs/CHECKPOINTS.md`:
```markdown
## 2026-10-28 HH:MM UTC · slice 12 book truth-up and prose restructure

- Branch/HEAD: `rebuild/plan1-economics-audit` / `<HEAD>`
- Changed: `book/content/*.js` (working tree, never committed); spec, plan, index, this checkpoint.
- Proven: `python book/tools/audit-archive.py` → exit 0, both `OK` lines; all ten modules load; `node book/build-docx.js` → `book/memoire-rebuild.docx`; numeric diff: `<n>` tokens lost, all justified; word count body `<n>`, total `<n>`.
- Unproven/blocked: production measurement (access pending); field audit (not executed); Linux live acceptance (blocked); eBPF, BBRv3, `.dist` (blocked/deferred).
- Next: human review of the rendered book; fill supervisor/jury/date placeholders; decide Master vs Licence LISIME wording.
```

- [ ] **Step 9: Commit the docs only**

Run:
```bash
git add docs/superpowers/specs/2026-10-28-cgo-book-truth-up-design.md docs/superpowers/plans/2026-10-28-cgo-book-truth-up.md docs/superpowers/plans/README.md docs/CHECKPOINTS.md
git commit -m "docs: approve slice 12 book truth-up and prose restructure"
```
Expected: one commit; `git status --short` still shows no `book/` path.

---

## Self-review

- Spec §3 title → Task 2.
- Spec §4 RQs → Tasks 3, 5, 8, 9, 11.
- Spec §5 claims 1–10 → Tasks 2–8, 10.
- Spec §6 Chapter III → Tasks 7, 8.
- Spec §7 prose rules → every content task.
- Spec §8 production limitation → Task 4.
- Spec §9 invariants → Global Constraints and Task 12.
- Spec §10 acceptance → Task 12.
- No task commits `book/`.
- No task adds a number that is not in `VERIFY.md` or the frozen corpus.
