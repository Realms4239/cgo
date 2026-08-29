# PRODUCT.md

**Meteolink** — banc reproductible pour l'audit client-side et la comparaison des politiques
AQM/BBR sur liens d'accès contraints, appliqué aux liens de Météo Madagascar.

Binaire Go autonome (tableau de bord React embarqué), dérivé du mémoire :
« Conception et évaluation d'un banc reproductible pour l'audit client-side et la
comparaison des politiques AQM/BBR sur liens d'accès contraints — application aux
liens de Météo Madagascar ».

## Utilisateur cible

Opérateur réseau ou ingénieur de Météo Madagascar qui doit vérifier qu'un lien
d'accès contraint (4G/5G, fibre) transporte correctement le trafic critique
(télémétrie, alertes, tableaux de bord) malgré les transferts de masse, et
choisir une politique AQM/CC pour les équipements sous contrôle local.

## Boucle opérateur

1. **Auditer le lien** — mesures client-side non intrusives (ping p50/p95,
   petits objets p95, goodput de masse) → `data/link_audit.csv`.
2. **Reproduire sur le banc** — rejouer le profil de lien sur le banc Linux
   (`tc`/`netem`), avec Façonnage du bord manuel (qdisc, capacité, délai,
   gigue, perte).
3. **Comparer** — matrice P1/P2 × `pfifo_fast`/`fq_codel`/CAKE × CUBIC/BBR,
   événements normalisés (30 s baseline → 120 s charge → 30 s récupération),
   portes de validité G0–G7, gel CSV + manifeste SHA-256.
4. **Constat** — lecture des résultats et recommandations dans le tableau de
   bord, preuves vérifiables (`cgo verify`, `cgo figures`).

## Quatre vues (tableau de bord, SSE 10 Hz, en français)

- **Campagne de mesure** — préparation et lancement des événements.
- **Tableau live** — suivi temps réel de l'événement en cours.
- **Résultats + Comparaison** — médianes/IQR, matrice de comparaison des configs.
- **Provenance & archives** — manifestes, intégrité, journal opérateur
  (ring de 50, `GET /api/events`), surveillance continue.

Complété par : audit autonome (`cgo audit`) et mode observation sous Windows
(`--mode observe|full|auto`).

## Non-goals

Pas d'alerting, pas de flotte multi-sites, pas de stockage longue durée, pas
d'authentification, pas d'outil de configuration de routeurs : Meteolink produit
les preuves mesurées qui éclairent ces décisions — il ne les prend pas.
