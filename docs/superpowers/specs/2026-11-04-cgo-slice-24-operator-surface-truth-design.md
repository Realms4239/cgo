# Slice 24 — Vérité des surfaces : le vocabulaire opérateur sur le bord vivant et les fixtures

*Spec écrite le 2026-11-04 par le builder. Statut : **projet** (brainstorming de la slice 24, HEAD `7f3f70c`). Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi

NEW.md §5.3 nomme l'« interactive sandbox » (Image 12) : `FIGER L'AVANT` et `bord non façonné` sont du jargon de banc, incompréhensible pour un technicien IT en service. La slice 21 (T8) a porté le vocabulaire opérateur sur les surfaces par défaut, mais pas sur la surface vivante du banc.

Vérifié à HEAD `7f3f70c` :

- `FIGER L'AVANT` subsiste (`web/frontend/src/views/LiveView.tsx:554`) : `lockBaseline`, `wall-baseline` localStorage. Le commentaire ligne 547 dit « trois sections calmes », mais le bouton reste du jargon.
- `bord non façonné (file simple)` subsiste (`LiveView.tsx:576`) ; « façonnage » apparaît 5+ fois dans LiveView.
- Les fixtures de test enseignent encore `/interface queue add` : `ProofDrawer.dom.test.tsx:43` (`routeros_ingress: '/interface queue add name=cgo-ingress'`) et `RemediationCard.dom.test.tsx:15` (même chaîne). Le générateur émet le LAN queue tree, figé par `TestRouterOSEmitsOnlyValidKinds` ; les fixtures DOM contredisent le générateur figé.
- NEW1 cockpit (3 cartes + CTA Fogg) est refusé en slice 22 §3.5 et inchangé : une réécriture DashboardView de 2–3 jours, pas une correction de surface. Différé par design, avec preuve.
- La troncature des tuiles profils est réfuite aux largeurs 1024–1920 : `web/frontend/src/lib/profiles.ts` porte des libellés courts explicites, et `.bench-profile small` élide seulement sous ~200px. Le mobile est abandonné par décision humaine.

## 2. Contrat

**Invariant** : tout texte opérateur ajouté aux surfaces vivantes est du français courant de service ; toute fixture DOM reflète la sortie du générateur figé.

1. `LiveView.tsx:554` : `FIGER L'AVANT` → `Maxi avant` (libellé : mémoriser la médiane de référence). Le tooltip explique le geste.
2. `LiveView.tsx:576` : `bord non façonné (file simple)` → `sans réglage (file simple)`. Le `plainQueue` reste inchangé.
3. Fixtures ingress : `ProofDrawer.dom.test.tsx:43` et `RemediationCard.dom.test.tsx:15` → LAN queue tree (`/queue tree add … parent=bridge queue=cgoingress-cake`, compte tenu des noms réels de `remediate.RouterOSIngress`). Les asserts de test portent la nouvelle syntaxe, jamais `/interface queue add`.
4. Tests : `ProofDrawer.dom.test.tsx` + `RemediationCard.dom.test.tsx` verts ; nouveau test LiveView ne portant que sur l'étiquette (DOM statique ou régression du libellé).

## 3. Non-objectifs

- Cockpit NEW1 (3 cartes, CTA Fogg, audit 60 s) : refusé en slice 22, inchangé.
- Tuiles profils, largeurs mobiles, renommage `PREUVE VÉRIFIABLE` : réfuté avec preuve en slice 23, inchangé.
- Aucun changement Go : le générateur et l'invariant sont déjà figés.
- Aucune valeur live, CHR, bench, audit, eBPF, `.dist`, BBRv3.
