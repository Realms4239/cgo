# Slice 25 — Honneur de l'infobulle Maxi avant : nommer la barre de comparaison

*Spec écrite le 2026-11-04 par le builder. Statut : **projet** (brainstorming de la slice 25, HEAD `f62a5a8`). Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi

Suite de la slice 24 : le bouton `Maxi avant` (ancien `FIGER L'AVANT`) mémorise une médiane de référence, mais son infobulle ne nomme pas ce que le geste déclenche — la barre de comparaison « avant → maintenant, même fenêtre » (`LiveView.tsx` ~560-575). Un opérateur qui clique sans savoir ce qui apparaît ne peut pas interpréter la barre. NEW.md §5.3 (« sandbox incompréhensible ») reste vrai pour cette moitié-là.

Vérifié à HEAD `f62a5a8` :

- Le `title` actuel : `mémoriser la médiane de référence, cliquer pour effacer` — ne mentionne ni la barre, ni « même fenêtre ».
- La barre existe déjà et est honnête (« avant → maintenant, même fenêtre », ±%). Seulement la porte d'entrée (tooltip) ne l'annonce pas.
- NEW1 cockpit (3 cartes + CTA Fogg) est refusé en slice 22 §3.5, différé en 23/24 : une réécriture DashboardView de 2–3 jours, pas une correction de surface. Différé par design, avec preuve.

## 2. Contrat

**Invariant** : toute action visible de la surface vivante annonce, dans son geste, l'effet qu'elle produit — sans nouveau jargon, en français courant.

1. `LiveView.tsx:554` : `title` étendu → `mémoriser la médiane de référence, cliquer pour effacer — la barre compare avant et maintenant, même fenêtre`.
2. Non-objectifs inchangés : cockpit NEW1 différé par design ; tuiles profils, largeurs mobiles, renommage `PREUVE VÉRIFIABLE` : réfuté avec preuve en slice 23, inchangé ; aucun changement Go, live, CHR, bench, audit, eBPF, `.dist`, BBRv3.
