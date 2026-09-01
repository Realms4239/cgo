# HANDOFF — Meteolink : 3 livrables de soutenance (L2 thèse + L3 PowerPoint)

**Date :** 2026-09-02
**Dépôt :** `C:\cgo` (main @ `f8a1317`, privé GitHub Realms4239/cgo)
**À qui :** agent frais, prends le relais SANS rien redécouvrir — ce document est la seule source de vérité du travail restant.

## MISSION

Trois livrables de soutenance pour un étudiant qui présente devant jury strict (20 min PowerPoint + 40 min QA) :
1. **L1 — FAIT** : `book/livrable1-explication.docx` (2 970 mots, 8 chapitres, kit anti-QA 30 questions, zéro em-dash, accentuation FR vérifiée). NE PAS REFAIRE.
2. **L2 — EN COURS (~15 %)** : thèse docx 85 p. exactes dans `book/` (pipeline existant, figures réelles générées, contenu à écrire).
3. **L3 — PAS COMMENCÉ** : PowerPoint 18 slides + 8 backup, design BMW M light mode.

Après les livrables : push final + tag v1.2.0.

## ÉTAT SYSTÈME (vérifié en direct, tout marche)

- **VM :** `192.168.174.131` (bail DHCP peut changer !), user `altfloat`, clé `~/.ssh/id_ed25519`, snapshot `cgo-pre-these` posé, `ens160` réparé via netplan `en*` (fichier `/etc/netplan/50-cgo.yaml` pattern `match: name: "en*"`).
- **Dashboard :** http://192.168.174.131:9090 → `{"mode":"full","ok":true,"version":"1.1.0"}`.
- **Batterie :** go test 9/9, vitest 59/59, playwright 7/7 (smoke+wall+archives) via `AUDIT_BASE=http://192.168.174.131:9090`.
- **Kit v1.1.0+ :** 15 actions (`doctor scan ensure align build deploy bootstrap status logs tunnel snapshot revert ssh ps backup`), testées en direct. `bin/cgo.exe` build local. **Le binaire sur la VM ne connaît PAS encore snapshot/revert/ps/backup/ssh** (pas re-déployé depuis) — pas nécessaire pour les livrables.
- **Données réelles :** `book/data/runs/` = 134 runs rapatriés (backup vérifié 23 Ko), `book/data/stats.json` extrait par `book/extract-stats.js` : 286 lignes, 83 valid, 154 quarantaine, 14 cellules. **Run clé = `run-1788191429`** (P2 deadline 220ms : pfifo/cubic 98%, pfifo/bbr **60.2%** vs fq_codel/bbr 96.2%, cake/bbr 98.1%, QDI 57.4 vs ~13-14). VoIP R vide dans les données (implémenté après ces runs — le dire honnêtement dans la thèse).
- **Sécurité :** clé seekai `sk-iyaUN...` purgée de l'historique ; `.seekai/`, `.env.example`, `scripts/seekai-*` gitignorés. GitHub token : `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` (scopes repo, workflow).

## LE DESIGN APPROUVÉ (verrouillé par grilling — ne pas rediscuter)

### Charte transversale (Q8 a+b+c)
- **Anti self-référentiel** : chaque chapitre part du monde (Météo Madagascar, ARTEC oct 2025, tarifs Yas 2026) et arrive à l'outil comme réponse. L'outil est utilisé, jamais célébré. Zéro adjectif d'auto-promotion (innovant, puissant, complet, robuste).
- **Humain/académique** : impersonnel ("l'étude", "le banc", "les mesures" ; "nous" réservé à la démarche). Chaque affirmation porte une source (RFC 8290, ITU-T G.107, yas.mg) ou un chiffre des CSV gelés. **AUCUN EM-DASH (—)**, pas de "il est important de noter", pas de phrases uniformes. Accentuation française parfaite. Test : lisible à voix haute sans sonner récité.
- **PAS de couleur de texte** autre que défaut Word noir.
- **Langue** : français intégral, noms de technologies en anglais (CAKE, BBR, p95).

### L2 — Thèse docx 85 p. EXACTES (Q6, plan de l'utilisateur)
| Section | Pages | Contenu |
|---|---|---|
| Pages préliminaires | 4 | Couverture, remerciements + résumé FR/EN, TOC + listes figures/tableaux |
| Introduction générale | 2-3 | Contexte, problématique, plan |
| I. Justification & contexte | 15-18 | Météo Madagascar 3-4 (organigrammes réels DGM/DEM dans `book/figures/fig-org-*.png`) · contexte 4-5 (liens d'accès malgaches, fibre/4G/VSAT/Starlink 226 000 Ar, tarification prépayée, ARTEC) · problématique 3-4 (bufferbloat vs trafic critique, RQ1-3) · objectifs 2-3 |
| II. Considération théorique | 20-22 | Méthodologie 6-8 (audit non intrusif, banc reproductible sur le POSTE CLIENT, protocole) · équations 8-10 (percentiles, QDI, deadline, E-model R chaque terme défini, coût paliers, SHA-256) · plan/algorithme 4-6 (30/120/30, portes G0-G7, topologie poste-client veth) · matériels 1-2 |
| III. Considération pratique | 26-28 | Application 8-10 (l'instrument : un binaire Go, dashboard 4 vues, TUI 5 onglets, CLI, setup wizard ; les campagnes menées) · Résultats 7-9 (deadline 60/98%, QDI, P4-starlink, coût Ar/Go — depuis stats.json) · Interprétation + discussion 8-10 (prescriptions DSI, CAKE/fq_codel/MikroTik, **6 limites plausibles**, perspectives) |
| Conclusion | 2 | Réponse aux RQ, apports |
| Bibliographie | 2 | RFC 8290/8312, ITU-T G.107, Gettys/Nichols 2011, Jiang IMC 2010, Cardwell BBR 2016, Höiland-Jørgensen CAKE 2018, Ha CUBIC 2008, yas.mg, ARTEC newsmada 24/10/2025, L'Express 26/11/2025, cable.co.uk, docs/data-prices.md |
| Annexes | 5-6 | Schéma aqm_eval.csv, portes complètes, guide cgo, captures dashboard |

**Les 6 limites plausibles (wording exact approuvé) :** métriques ciblées (JFI/TTB écartés car flux bulk unique) · répétitions (1 par cellule finales, 3 systématiques = perspective) · fenêtre 120 s (pas saturation prolongée/diurne) · trafic synthétique (bulk TCP unique, pas multiplexage réel) · banc netem ≠ couche radio · 4 classes de liens (pas calibration par site).

**Narratif banc (Q-révision) :** le banc s'installe sur le POSTE CLIENT de l'agent, aucune modification de l'infrastructure Météo ; le lien d'accès est reconstitué par une paire veth (netem sur l'émission + discipline de file). Figure figR1-topologie.

### L3 — PowerPoint (Q7b, structure approuvée)
- **Design BMW M LIGHT MODE** (DESIGN.md inversé) : fond blanc pur, typographie noire, titres UPPERCASE 700 (36-44pt, Arial/Inter en substitution), corps 300 (≥24pt), légendes 16-18pt, rectangles 0px, **tricolore M (#0066b1→#1c69d4→#e22718) sur slides 1 et 18 UNIQUEMENT** (4px stripe), accent données #1c69d4, pas de rouge sur les résultats (confus avec échec), fondus seuls, numéros de slides, 30-50 mots/diapo, 1 idée/diapo, titre = la conclusion.
- **18 slides** : 1 Titre+tricolore (30s) · 2 Réseau vital, budget prépayé (1min) · 3 Péage embouteillé/bufferbloat schéma (1min) · **4 LE RÉSULTAT : 60% vs 98% deadline** (graphique figR2, 1min) · 5 Banc sur le poste client (figR1, 1min) · 6 Huit portes qualité (1min) · 7 Un seul binaire, tout vérifiable (1min) · 8 Dashboard capture live (1min) · 9 QDI régularité (figR3, 1min) · 10 Voix E-model (1min) · 11 Starlink LEO (1min) · 12 Prix du gaspillage Ar/Go (figR4, 1min) · 13 Résultats+interprétation capture (1min) · 14 Prescription CAKE routeur site (1min) · 15 Contributions (1min) · 16 Limites : métriques ciblées, flux unique, banc ≠ radio (1min) · 17 Conclusion (1min) · 18 Merci+tricolore (30s).
- **8 backup** : équations, portes complètes, matrice 36 events, schéma CSV, CLI, comparaison outils (Flent/iperf3), position métriques/standards (small_p95≈RFC8290 latency-under-load, QDI≈Flent spread, deadline≈SLO, cost_ar=originale justifiée tarifs), sources tarifs.
- **Notes orateur** : verbatim par slide (phrases simples), règle 3 secondes regard notes, slides sacrifiables = 9, 10, 11. Remerciements à la FIN jamais au début. Remplir 80-85 % du temps.

## PIPELINE TECHNIQUE (tout existe, dans book/)

- **`book/build-docx.js`** : assembleur docx-js, contrat de style Times New Roman 12pt, interligne 1,5, justifié, alinéa 12,5mm, marges 25mm, légendes tableaux AU-DESSUS figures EN-DESSOUS (italique centré), TOC champs Word, renvois `{figRef:X}`/`{tabRef:X}` → champs REF. Sortie `thesis-v1.docx`. Modules `content/*.js` (00-front, 01-intro, 015-organisme, 02-partie1, 03-partie2, 04-partie3, 05-partie4, 06-conclusion, 07-refs, 08-glossaire, 09-annexes) — **actuellement l'ANCIEN sujet (CONGESTION.md), à RÉÉCRIRE entièrement pour LIEN** (structure garder : méthode d'export `{h1:...}`, `{p:...}` etc. — lire un module pour le format exact).
- **`book/fig-lib.js`** : lib SVG académique (Svg, chartFrame, linScale, legend, PAL encre/gris/bleu académique). API : `legend(svg, x, y, [[label, color], ...])` (array de pairs, PAS d'objets).
- **`book/figures-lien.js`** : NOUVELLES figures LIEN générées depuis stats.json : figR1-topologie (poste client + veth + serveur), figR2-deadline (LES 60/98%), figR3-qdi, figR4-tarifs (paliers Yas réels), figR5-gel (pipeline SHA-256). Déjà en SVG+PNG (2700px, render vérifié par tailles).
- **`book/svg-to-png.js`** : SVG→PNG 3x via Playwright (chemin corrigé vers C:/cgo/web/frontend/node_modules).
- **`book/figures/`** : 22 SVG ancien sujet (réutiliser fig01-cables, fig02-rtt-phases, fig07-archi si utile, organigrammes DGM/DEM) + 5 figR nouvelles + 24 PNG.
- **`book/data/stats.json`** : les chiffres réels (extraits, voir ci-dessus). **`book/extract-stats.js`** pour régénérer.
- **Captures dashboard réelles** : sur la VM `~/cgo` via e2e existants (`web/frontend/e2e/` a smoke, wall, visual-audit — lancer avec `AUDIT_BASE=http://192.168.174.131:9090 npx playwright test ... --workers=1` ; captures tombent dans `web/frontend/shots/`). Pour L3 : 3 captures nécessaires (campagne, live, resultats) — un spec dédié qui screenshot les 4 panneaux à 1920 existe déjà dans visual-audit.spec.ts, l'adapter.
- **Skill docx** (`C:\Users\ASUS\.agents\skills\docx`) : guidelines docx-js (déjà respectées par build-docx.js). **Skill pptx** (`C:\Users\ASUS\.agents\skills\pptx`) : à utiliser pour L3 (HTML→PPTX ou pptxgenjs selon le skill — LIRE le SKILL.md d'abord).
- **`docs/data-prices.md`** : recherche tarifs complète (Yas/Airtel/Orange/FTTH/Starlink/ARTEC, sources datées) — matière pour II.5 et figR4.
- **`docs/defense-research.md`** : recherche guidelines soutenance (15-18 slides, résultats 50 %, 6/6/6, interdits, Q&A) — S'Y RÉFÉRER pour L3.
- **`book/build-l1.js`** : générateur L1 (exemple de style rédactionnel approuvé — s'en inspirer pour le ton L2/L3).

## CONTRAINTES RÉDACTION (rappel brutal)

1. AUCUN em-dash `—` (utiliser `:` ou reformuler).
2. AUCUNE couleur de texte (noir Word par défaut).
3. Accentuation française PARFAITE (é è à ç ô î û).
4. Pas de self-référence ("notre outil innovant") — l'outil arrive comme réponse au monde.
5. Chaque chiffre depuis stats.json / data-prices.md — ne RIEN inventer. VoIP R : dire que le score est implémenté mais que les vagues présentées l'ont précédé (runs pré-voip), donner le calcul théorique par profil dans II.
6. Quotas de pages EXACTS (tableau ci-dessus), marge ~5 p. max.
7. Lisible à voix haute — phrases variées, connecteurs logiques, pas de listes interminables sans transitions.

## PLAN D'EXÉCUTION RECOMMANDÉ

1. **Lire** : `book/content/015-organisme.js` (format module), `book/build-docx.js` (assembler), `docs/data-prices.md`, 1 chapitre de `book/livrable1-explication.docx` (ton), figR*.png (nommer, pas lire le contenu image).
2. **L2 contenu** : réécrire les 11 modules content/*.js pour LIEN selon le plan 85 p. (utiliser stats.json pour tous les tableaux résultats — Tableau 10 du LIEN.md original = nos vrais chiffres deadline/QDI/cost ; Tableau 9 = audit honnête "méthode prête, déploiement terrain en perspective"). Injecter figR1-R5 + réutiliser fig01/fig02/organigrammes. Attention : LIEN.md racine est STALE (pré-système) — la vérité = le système v1.1.0 actuel + stats.json + docs/.
3. **L2 build** : `node build-docx.js` → vérifier pagination (compter les pages par section dans le docx produit, ajuster le volume de prose pour coller aux quotas — c'est l'itération principale).
4. **L3** : lire skill pptx, puis construire les 18+8 slides (design ci-dessus), figures figR réutilisées + 3 captures dashboard fraches. Notes orateur dans les notes PPTX.
5. **Vérification finale** : go test + vitest + e2e smoke → commit → push → tag v1.2.0 + release GitHub (même procédure que v1.1.0 : binaires 5 plateformes vers /tmp/rel120, `gh release create v1.2.0` avec assets + notes FR).
6. **Ne pas** re-déployer la VM pour les livrables (le dashboard .131 tourne, c'est ce qui compte pour les captures) ; NE PAS toucher au snapshot `cgo-pre-these`.

## PIÈGES CONNUS

- `legend()` prend des pairs `[label, color]`, pas des objets (erreur déjà commise).
- docx npm installé dans `book/node_modules` (racine C:\cgo a aussi un package.json parasite des scripts seekai — ignorer, travailler dans book/).
- Les PNG figR sont en 2700px de large (3x) — pour PPTX les régénérer à 1x ou les redimensionner dans le deck.
- Playwright e2e : TOUJOURS `AUDIT_BASE=http://192.168.174.131:9090` (config par défaut localhost). `--workers=1` obligatoire.
- IP VM .131 peut glisser (DHCP) : si SSH échoue, `cgo kit ensure` (dans bin/cgo.exe) re-découvre l'IP ; vérifier `ping`.
- LibreOffice absent de ce PC (pas de conversion PDF pour vérif visuelle docx) : vérifier via `pandoc` absent aussi — utiliser `unzip + word/document.xml` (méthode déjà éprouvée dans ce handoff, voir git log) ou installer LibreOffice si besoin visuel.
- go test ./pkg/api nécessite web/frontend/dist (bun run build d'abord) — dist/.placeholder versionné fait que ça compile toujours.
