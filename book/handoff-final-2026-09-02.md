# HANDOFF — Meteolink : finir L2 (thèse docx) + L3 (PPTX) + v1.2.0

**Date :** 2026-09-02
**Dépôt :** `C:\cgo` (main @ `54d3f62`, privé, GitHub `Realms4239/cgo`)
**Token GitHub (scopes repo, workflow) :** à demander à l'utilisateur — NE PAS le mettre dans un fichier. (L'utilisateur le collera en session.)

## OÙ ON S'ARRÊTE (lire ceci d'abord)

Tout le travail système est FAIT et vérifié. Il reste UNIQUEMENT les livrables de soutenance + le ship. Le dernier agent a été interrompu en vérifiant la pagination du docx (`thesis-v1.docx` généré 2 580 Ko mais le check unzip échouait pour des raisons de chemin Windows/tmp — le fichier est VALIDE, c'est le check qui a échoué, pas le build).

**ÉTAT DES 3 LIVRABLES :**

1. **L1 — FAIT** : `book/livrable1-explication.docx` (2 970 mots, 8 chapitres, kit anti-QA 30 questions, zéro em-dash). NE PAS REFAIRE.
2. **L2 — ~70 % FAIT** : modules content réécrits pour LIEN, `thesis-v1.docx` buildé (2 580 Ko). IL RESTE : vérifier pagination vs quotas 85 p., injecter les figures réelles figR1-R5 dans les bons modules, itérer.
3. **L3 — PAS FAIT** : PowerPoint 18+8 slides BMW M light mode.

## SUGGESTED SKILLS (à invoquer dans cet ordre)

1. **`docx`** — pour vérifier/itérer `thesis-v1.docx` (render → PDF → pages check). Lire le SKILL.md d'abord.
2. **`pptx`** — pour construire L3. Lire le SKILL.md d'abord (il propose HTML→PPTX ou pptxgenjs).
3. **`grilling`** — UNIQUEMENT si l'utilisateur veut rediscuter un point du design. Le design est VERROUILLÉ (voir ci-dessous), ne pas re-griller par défaut.

## LE DESIGN VERROUILLÉ (ne pas rediscuter)

**Lire `book/handoff-livrables-2026-09-02.md`** — c'est LE document de référence exhaustif : charte rédactionnelle, plan 85 p. section par section, structure 18+8 slides, 6 limites plausibles word-for-word, narratif banc poste-client, pipeline technique, pièges connus. TOUT y est.

Résumé brutal des contraintes non négociables :
- **Zéro em-dash `—`**, zéro couleur de texte, accentuation FR parfaite, impersonnel académique, anti self-référentiel (l'outil arrive comme réponse au monde), chaque chiffre depuis `book/data/stats.json` ou `docs/data-prices.md`.
- **L2 plan 85 p. EXACT** (quotota par section dans le handoff) : préliminaires 4 · intro 2-3 · I. contexte 15-18 · II. théorie 20-22 · III. pratique 26-28 · conclusion 2 · biblio 2 · annexes 5-6.
- **L3** : 18 slides (structure exacte dans le handoff, timing par slide), 8 backup, fond blanc, titres UPPERCASE 700 36-44pt, corps ≥24pt, tricolore M (#0066b1→#1c69d4→#e22718) sur slides 1 et 18 UNIQUEMENT, accent données #1c69d4, 30-50 mots/diapo, fondus seuls, notes orateur verbatim, slides sacrifiables = 9/10/11.

## CE QUI A ÉTÉ FAIT DEPUIS LE DERNIER HANDOFF

1. **Kit v1.1.0+ : 15 actions** (`doctor scan ensure align build deploy bootstrap status logs tunnel snapshot revert ssh ps backup`) — commit `f8a1317`, testées en direct sur la VM (snapshot `cgo-pre-these` posé, ssh OK, ps 2 Hz health=OK, backup 23 Ko vérifié 134 CSV).
2. **Modules content réécrits pour LIEN** : `01-intro.js` (6 blocs), `02-partie1.js` (~30 blocs), `03-partie2.js` (~35 blocs), `04-pratique.js` (49 blocs, RÉSULTATS avec tableaux stats.json). **ATTENTION : les ANCIENS modules de l'ancien sujet existent TOUJOURS** (`04-partie3.js` 121 blocs ancien sujet, `05-partie4.js` etc.) — `build-docx.js` référence encore `04-partie3` dans son tableau MODULES (ligne ~32). **IL FAUT remplacer la liste MODULES de build-docx.js pour pointer vers les nouveaux modules** : `00-front, 01-intro, 015-organisme, 02-partie1, 03-partie2, 04-pratique, 06-conclusion, 07-refs, 08-glossaire, 09-annexes` (retirer 05-partie4 ancien ou le réécrire en conclusion/annexes).
3. **`thesis-v1.docx` buildé** (2 580 Ko à la racine `C:\cgo\thesis-v1.docx`) MAIS avec les ANCIENS modules (04-partie3 = 121 blocs ancien sujet) — le build a réussi parce que tous les blocs parsent. **Le docx actuel contient donc un mélange ancien/nouveau — IL FAUT le rebuilder avec la nouvelle liste de modules.**
4. **Figures réelles générées** : `book/figures/figR1-topologie.{svg,png}`, `figR2-deadline.{svg,png}` (LE résultat 60/98 %), `figR3-qdi.{svg,png}`, `figR4-tarifs.{svg,png}`, `figR5-gel.{svg,png}` — PNG 2700px (3x), générés par `book/figures-lien.js` depuis `book/data/stats.json` (134 runs, 286 lignes, 83 valid, run clé `run-1788191429`).
5. **`04-pratique.js` référence figR2 et figR3 en double** (2 tableaux avec le même id `final220` — BUG : deux blocs table avec id identique `final220` casseront les renvois REF ; renommer le second `res220` ou supprimer le doublon). Vérifier tous les ids de tableaux uniques.

## PLAN D'EXÉCUTION (dans l'ordre)

### Étape 1 — L2 : fixer build-docx.js et rebuilder
1. Ouvrir `book/build-docx.js`, remplacer la liste MODULES (ligne ~30-34) : `'00-front', '01-intro', '015-organisme', '02-partie1', '03-partie2', '04-pratique', '06-conclusion', '07-refs', '08-glossaire', '09-annexes'` (supprimer `04-partie3` et `05-partie4` de la liste).
2. Réécrire `06-conclusion.js`, `07-refs.js`, `08-glossaire.js`, `09-annexes.js` pour LIEN (ils sont encore ancien sujet — ~91 lignes refs, 36 glossaire, 78 annexes ; s'inspirer du ton de `01-intro.js` et de `book/livrable1-explication.docx`).
3. Vérifier `00-front.js` (page de titre : titre du mémoire, noms, année 2025-2026) et `015-organisme.js` (présentation Météo Madagascar — probablement réutilisable, vérifier).
4. Fixer le doublon `final220` dans `04-pratique.js`.
5. `node build-docx.js` → vérifier la taille et le contenu (unzip word/document.xml, compter mots SEQ Figure/SEQ Tableau).
6. **Itérer la pagination** : le docx doit faire ~85 p. Compter par section (render PDF si possible via LibreOffice — ABSENT de ce PC, utiliser `soffice` si installé sinon compter les paragraphes et estimer ~350 mots/page à TNR 12 interligne 1,5 ; ajuster le volume de prose dans les modules).

### Étape 2 — L3 : PPTX
1. Invoquer le skill `pptx`.
2. 18 slides selon la structure du handoff (dans `book/handoff-livrables-2026-09-02.md` § L3).
3. Figures : réutiliser figR2 (slide 4 LE résultat), figR1 (slide 5 topologie), figR3 (slide 9), figR4 (slide 12) — PNG 2700px à redimensionner (~900px de large dans le deck).
4. 3 captures dashboard : lancer `cd web/frontend && AUDIT_BASE=http://192.168.174.131:9090 npx playwright test e2e/screenshots.spec.ts --workers=1` (vérifier que ce spec existe sinon l'adapter depuis visual-audit.spec.ts) → shots dans `web/frontend/shots/`.
5. Notes orateur : dans les notes PPTX de chaque slide (verbatim simple, règle 3 secondes).
6. Backup 8 slides après la 18.

### Étape 3 — Ship v1.2.0
1. Batterie : `go vet ./...`, `go test ./pkg/... -short` (9/9), `bun x tsc --noEmit` (0), `bunx vitest run` (59/59), `bun run build` (<600 KB gz), e2e smoke via AUDIT_BASE (7/7).
2. `bash kit/test-overhaul.sh` (12/12), `bash kit/test-fresh-clone.sh` (clone → setup → doctor).
3. VERSION → 1.2.0, ldflags main.go, npm/package.json, docs (1.1.0 → 1.2.0 partout).
4. Commit, push (demander le token GitHub à l'utilisateur).
5. Tag `v1.2.0` + release GitHub (gh release create avec 5 binaires cross-compilés vers /tmp/rel120/ + checksums.txt — même procédure que v1.1.0, voir `git log v1.1.0`).

## FAITS SYSTÈME (tout marche, ne pas redécouvrir)

- **VM :** `192.168.174.131` (SSH `altfloat`, clé `~/.ssh/id_ed25519`), snapshot `cgo-pre-these` posé, dashboard http://192.168.174.131:9090 `{"mode":"full","version":"1.1.0"}`, testbed veth OK. Si SSH échoue : `bin/cgo.exe kit ensure` (IP auto-découverte).
- **P1-P4 profils :** P1 fibre 80/20ms, P2 4G 20/100ms/0,5 %, P3 VSAT 5/600ms, P4 Starlink 100/40ms (implémentés, campagne P4 faite, gelée).
- **VoIP R (E-model G.107 simplisé) :** implémenté + testé (`pkg/metrics/metrics.go:VoIPR`, golden tests verticaux), colonne `voip_r` dans aqm_eval.csv. Les runs existants l'ont précédé (valeurs vides dans les vagues présentées) — le dire honnêtement dans la thèse.
- **Coût paliers réels 2026 :** `pkg/metrics/metrics.go:Tiers` (yas-day-1gb 1000, yas-month-4.5gb 5556, ftth 490 Ar/Go), `GET /api/cost/tiers`. Recherche complète dans `docs/data-prices.md`.
- **Résultat clé :** run `run-1788191429` (P2, deadline 220 ms) : pfifo/BBR **60,2 %** vs fq_codel/BBR 96,2 % vs cake/BBR **98,1 %** ; QDI 57,4 vs ~13-14 ms ; débit préservé 17,8-19,6 Mb/s.
- **Bugs UI corrigés :** header z-index (pill source incliquable), profils wrap cockpit, d3 retiré (315 KB gz), toutes vues montées (retour instantané).
- **Incidents sécurité passés :** clé seekai purgée de l'historique (vérifié), `.seekai/` `.env.example` `scripts/seekai-*` gitignorés. NE JAMAIS `git add -A` sans vérifier `git status` d'abord.

## PIÈGES (déjà rencontrés)

- `legend()` de fig-lib prend des PAIRS `[label, color]`, pas des objets.
- Deux blocs `table` avec le même `id` cassent les REF Word — ids uniques obligatoires.
- `build-docx.js` lit les PNG via `pngSize()` (octets 16-20 du fichier) — les PNG doivent être de vrais PNG non corrompus.
- LibreOffice absent (pas de render PDF local) ; pandoc absent ; vérifier les docx par `unzip + word/document.xml` (attention : le chemin /tmp sous Git-Bash mappe différemment sous Windows — utiliser un dossier du dépôt pour l'extraction).
- Playwright : `AUDIT_BASE=http://192.168.174.131:9090` obligatoire (défaut localhost), `--workers=1`.
- go test ./pkg/api nécessite `web/frontend/dist` — le placeholder versioné fait que ça compile toujours.
- L'utilisateur n'aime pas qu'on s'arrête : « do not stop anymore » — exécuter en continu, rendu inline.

## FICHIERS CLÉS

| Fichier | Rôle |
|---|---|
| `book/handoff-livrables-2026-09-02.md` | LE handoff exhaustif (design, plan, pièges) |
| `book/data/stats.json` | Les chiffres réels (134 runs) |
| `book/figures/figR*.png` | 5 figures réelles prêtes |
| `book/figures-lien.js` | Générateur des figures (API legend en pairs) |
| `book/build-docx.js` | Assembleur thèse (TNR 12, 1,5, TOC champs) — MODULES À FIXER |
| `book/content/01-intro.js` `02-partie1.js` `03-partie2.js` `04-pratique.js` | Nouveaux modules LIEN (faits) |
| `book/content/04-partie3.js` `05-partie4.js` | ANCIENS modules (à retirer de la liste) |
| `book/content/06-conclusion.js` `07-refs.js` `08-glossaire.js` `09-annexes.js` | À réécrire pour LIEN |
| `book/livrable1-explication.docx` | L1 FAIT (ton de référence) |
| `thesis-v1.docx` (racine) | Build actuel MÉLANGÉ ancien/nouveau — à rebuilder |
| `docs/data-prices.md` | Recherche tarifs (sources datées) |
| `docs/defense-research.md` | Guidelines soutenance (15-18 slides, 6/6/6, Q&A) |
| `C:\Users\ASUS\Videos\DESIGN.md` | Design system BMW M (pour L3 light mode inversé) |
