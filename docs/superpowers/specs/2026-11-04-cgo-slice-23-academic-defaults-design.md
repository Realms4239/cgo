# Slice 23 — Defaults académiques : la deadline CAP/WIS 2.0 et le pill BBRv3

*Spec écrite le 2026-11-04 par le builder. Statut : **projet** (brainstorming de la slice 23, HEAD `8477d20`). Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi

NEW.md §5.1 : « The 1,000 ms Deadline Absurdity » — la thèse repose sur la deadline de 220 ms (objectif opérateur CAP/WIS 2.0, `metrics.CAPDeadlineMs`), mais le front propose 1000 comme défaut de réglage. NEW.md §5.2 : le chemin de script `kit/install-bbr3.sh` fuit dans un label de formulaire.

Vérifié à HEAD `8477d20` :

- Le jury a raison sur la **deadline par défaut** : `web/frontend/src/lib/settings.ts:19` (`deadlineMs: 1000`), `web/frontend/src/lib/schema.ts:18` (`default: 1000`) et le repli `|| 1000` de `web/frontend/src/components/BenchConfig.tsx:265` contredisent `duel?.deadline_ms ?? 220` (`BenchConfig.tsx:145`) et `metrics.CAPDeadlineMs = 220`. Go ne consomme pas cette valeur : c'est un défaut d'interface produit, initial, alors que la thèse argumente sur 220 ms comme non discriminant en deçà.
- Aucune chaîne `PREUVE VÉRIFIABLE` ne subsiste dans le DOM : l'objet de NEW.md §3.2 est **réfuté** — `ProofDrawer.tsx` n'imprime que les hashes existants (`source_sha256 || 'aucune empreinte'`), et `RenderMarkdown` dit `Preuve: non enregistrée` quand `Artifact == ""`. Rien à corriger, à documenter seulement.
- La plainte « truncation » (P1FIBRE 8…) est **réfutée** au niveau du banc : les noms de profils ne tronquent pas ; seule la ligne de paramètres `.bench-profile small` élide (par conception, 1024–1920 px accepts). Le mobile est abandonné par décision humaine.
- Le chemin `kit/install-bbr3.sh` est **toujours** imprimé : `BenchCapabilities.tsx:42`.

## 2. Contrat

**Invariant de défaut** : tout défaut produit de deadline vaut 220 ms (objectif CAP/WIS 2.0, `metrics.CAPDeadlineMs`). Les bornes serveur restent autoritaires ; le champ peut être édité, mais le client fraîchement installé démarre à 220.

1. `settings.ts` : `deadlineMs: 1000 → 220`.
2. `schema.ts` : `default: 1000 → 220` sur `deadline_ms` (repli local du schéma servi par `/api/schema`).
3. `BenchConfig.tsx:265` : `|| 1000 → || 220`.
4. `BenchCapabilities.tsx:42` : `BBRv3 absent de ce noyau · kit/install-bbr3.sh` → `BBRv3 non détecté`. Le chemin du script reste documenté dans le code et le runbook serveur, jamais dans l'interface.
5. Tests : `settings.test.ts` et `schema.test.ts` assertent 220 ; `BenchCapabilities.dom.test.tsx` assertent la nouvelle étiquette et l'absence de `install-bbr3.sh`.

## 3. Non-objectifs

- Cockpit NEW1 (3 cartes, CTA Fogg, audit 60 s) : refusé en slice 22, inchangé.
- Largeurs mobiles, tiles profils, renommage `PREUVE VÉRIFIABLE`, audit live, bancaire, CHR : inchangés, toujours hors scope §3.5 slice 22.
- Aucun changement Go : `metrics.CAPDeadlineMs` existe déjà à 220.
- Les réglages déjà persistés en `localStorage` restent chez l'utilisateur (un défaut produit n'écrase pas un choix explicite) ; le test pin le **défaut produit**, pas la persistance.
