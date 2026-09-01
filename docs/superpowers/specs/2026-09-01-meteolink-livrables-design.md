# Design : 3 livrables Meteolink — explication, thèse, PowerPoint

**Date :** 2026-09-01
**Statut :** Approuvé (brainstorming complet, Q1-Q11 validés)
**Auteur :** Équipe Meteolink + encadrant

## 1. Contexte et problème

Météo Madagascar opère sur des liens d'accès inégaux (fibre siège, 4G départementale Yas, VSAT stations isolées, Starlink LEO récent) avec une contrainte prépayée forte (25 000 Ar / 4,5 Go, saisine ARTEC octobre 2025). Le bufferbloat dégrade le trafic critique (télémétrie, tableaux de bord, alertes) pendant les transferts massifs (imagerie satellite, modèles). Le besoin : mesurer, comparer les remèdes (AQM CAKE/fq_codel, CC BBR/CUBIC), et recommander à la DSI avec des preuves vérifiables (gel SHA-256).

Le projet est réel au sein de l'entreprise, présenté en soutenance de 1 heure (20 min PowerPoint + 40 min QA) devant un jury strict. L'étudiant, non-spécialiste réseau, n'aime pas le compliqué et vise une très bonne note.

## 2. Objectifs et critères de succès

- L'étudiant comprend et peut défendre les 10 concepts clés et les chiffres réels (60 % vs 98 % deadline, QDI 57 vs 13, etc.) sans jargon.
- Le mémoire (85 pages) respecte le plan imposé et le contrat de style du book existant (TNR 12, 1,5, 25 mm, légendes tableaux au-dessus/figures en-dessous).
- Le PowerPoint (18 slides + 8 backup) respecte les guidelines soutenance (15-18 slides, résultats 50 %, 6/6/6, titres 36-44pt).
- Les 3 livrables sont exempts de self-référentiel et d'AI slop, accentuation française correcte, couleur de texte par défaut uniquement.

## 3. Approches considérées

### Livrable 1 : explication

- (a) Document unique ~25 pages, 8 chapitres pyramidaux — chaque chapitre : analogie → explication → encadré "si le jury demande", kit anti-QA de 30 questions. **Retenu** : moins de fragmentation, lecture en une soirée, verbatim pour l'oral.
- (b) Trois documents séparés — écarté : fragmentation pour un étudiant qui n'aime pas le compliqué.
- (c) Cartes à réviser la veille — écarté : trop synthétique pour une très bonne note.

### Livrable 2 : thèse

- (a) LIEN.md → docx tel quel — écarté : tableaux résultats vides, pas de chapitre réalisation.
- (c) LIEN.md enrichi + chapitres manquants (résultats réels, VoIP/Starlink, réalisation v1.1/v1.2, position métriques) — **retenu**.

### Livrable 3 : PowerPoint

- (a) Classique linéaire — écarté : résultats trop tard.
- (b) Problème → preuve → explication (résultat à la diapo 4) — **retenu** : capte le jury, 6 diapos résultats + 8 backup.
- (c) Démo-centric — écarté : tour de produit, pas soutenance.

## 4. Design détaillé

### 4.1 Livrable 1 — Explication (8 chapitres, ~25 pages, docx)

| Chapitre | Contenu |
|---|---|
| 1 | Météo Madagascar et ses liens (mission, fibre/4G/VSAT/Starlink, prépayé, ARTEC) |
| 2 | Le péage embouteillé (analogie, bufferbloat) |
| 3 | Les remèdes : CAKE et BBR |
| 4 | Le banc sur le poste client (topologie veth, profils P1-P4, protocole 30/120/30, portes G0-G7) |
| 5 | Les mesures (p95, deadline, QDI, VoIP R, coût) + vrais chiffres |
| 6 | La preuve vérifiable (gel SHA-256) |
| 7 | Raconter au jury (verbatim, règle 30-50 mots, slides sacrifiables) |
| 8 | Kit anti-QA (30 questions + 5 phrases de recadrage) |

Chaque terme technique défini à première apparition. Style : ton oral, phrases courtes. Généré via skill docx, vérifié sans em-dash, accentuation française, couleur par défaut.

### 4.2 Livrable 2 — Thèse (85 pages, book/)

- **Pipeline** : `book/build-docx.js` (TNR 12, 1,5, 25 mm) + `fig-lib.js` (SVG déterministe, palette encre grise/bleu) + `svg-to-png.js`.
- **Plan (quotas exacts)** : Préliminaires 4p. (couverture, remerciements+résumé, TOC+listes 2p.) + Intro 2-3p. + I. Justification & contexte 15-18p. (institut 3-4, domaine 4-5, problématique 3-4, objectifs 2-3) + II. Considération théorique 20-22p. (méthodo 6-8, équations 8-10, plan/algorithme 4-6, matériels 1-2) + III. Considération pratique 26-28p. (application 8-10, résultats 7-9, interprétation 8-10) + Conclusion 2p. + Biblio 2p. + Annexes 5-6p. Marge 5p.
- **Figures (~20)** : Toutes nouvelles via fig-lib.js depuis données réelles (CSV VM rapatriés) : topologie poste-client, matrice P1-P4×qdisc×CC, protocole+portes, courbes small_p95/deadline/QDI/VoIP R, paliers tarifaires, pipeline gel, architecture, + 3 captures dashboard. Anciennes figures CONGESTION (détecteurs) non réutilisées.
- **Contenu** : LIEN.md enrichi avec résultats réels (133 runs, deadline 60/98%, QDI, VoIP R, P4-starlink, tarifs réels 2026), chapitres réalisation v1.1/v1.2.

### 4.3 Livrable 3 — PowerPoint (18 slides + 8 backup, pptx)

- **Design** : BMW M light mode (fond blanc, typo noire UPPERCASE 700 36-44pt, rectangles 0px, tricolore M #0066b1→#1c69d4→#e22718 sur diapos 1 et 18 seulement, accent #1c69d4). Règle 30-50 mots/diapo, 6/6/6, numérotation.
- **Structure (b)** : 1 Titre, 2 Contexte Météo, 3 Péage, 4 LE RÉSULTAT 60 vs 98%, 5 Banc poste client, 6 Portes, 7 Instrument, 8 Dashboard live, 9 QDI, 10 VoIP R, 11 Starlink, 12 Coût, 13 Résultats+interprétation, 14 Prescription, 15 Contributions, 16 Limites plausibles (métriques ciblées, flux unique, banc ≠ radio), 17 Conclusion, 18 Merci. Backup 8 : équations, portes complètes, matrice, CSV schéma, CLI, outils, métriques, tarifs.
- **Notes orateur** : verbatim par slide, règle 3 secondes, slides sacrifiables 9,10,11 marquées.

## 5. Contraintes transversales

- Anti self-référentiel : faits extérieurs d'abord, outil comme réponse, impersonnel académique thèse ("l'étude"), zéro auto-promotion.
- Humain : une idée par paragraphe (affirmation→preuve→explication→transition), connecteurs variés, chaque claim sourcé (RFC, G.107, yas.mg, ARTEC) ou chiffré CSV.
- Accentuation française, pas de couleur de texte autre que défaut, pas d'em-dash, pas de "il est important de noter".
- Figures maison, légendes Table au-dessus/Figure en-dessous, numérotation séparée.

## 6. Plan d'implémentation

Ordre : L1 → L2 → L3, consécutivement, chaque livrable vérifié avant le suivant (docx via unzip+word/document.xml, pptx via lecture). Utiliser skills docx et pptx, fig-lib, données VM rapatriées.

## 7. Critères d'acceptation

- L1 : 2970 mots, 84 paragraphes, 0 em-dash, accentuation présente, 8 chapitres, kit anti-QA 30 questions.
- L2 : 85 pages ±5, TNR 12, 1,5, 25 mm, TOC champs Word, 20 figures fig-lib + 3 captures, LIEN enrichi.
- L3 : 18 slides + 8 backup, 30-50 mots/diapo, titres 36-44pt, notes orateur, design BMW M light.

## 8. Risques et mitigations

- Étudiant non-technique : L1 comme prérequis, analogies avant définitions.
- Jury strict sur métriques custom : tableau position vs standards (RFC 8290, Flent, E-model) + contribution cost_ar ancrée.
- Temps : 3 livrables consécutifs, chacun avec vérification docx/pptx.
