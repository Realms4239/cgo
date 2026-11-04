# Slice 26′ — Re-certification exhaustive 20A–25 : re-lier chaque affirmation à sa sortie fraîche

*Spec écrite le 2026-11-04 par le builder. Statut : **projet** (brainstorming de la redo, HEAD `9df16a8`). Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi

L'exécution des tranches 20–25 a laissé une dette d'évidence, pas de code :

- **20A** (operator truth engine, 7 tâches) a livré le moteur (`metrics.AssessVerdict`, `operator_verdict` scellé, `load_attempted`, taux par direction, `numbers` économiques, `timeline` capée, `/api/duel` n≥3, `validRunID`). Le paquet `docs/verification/2026-11-01-slice-20a.md` existe, mais les 7 lignes `PENDING` de `docs/VERIFY.md` n'ont jamais été déplacées par un vérificateur. La slice 21 les a absorbées sans réconciliation.
- **20C** (TUI truth panes) idem : code + paquet existent, 1 ligne `PENDING` jamais déplacée.
- **21** a un paquet, mais 12 propositions intouchées, et ses captures incluent 390 (abandonné par décision humaine).
- **22–25** ont des paquets, 19 propositions intouchées, et **zéro capture** : le packet discipline était de ne pas les revendiquer, mais l'utilisateur demande désormais la preuve visuelle aux largeurs acceptées (1024/1920, pas 390).

La slice 26′ est une **re-certification**, pas une feature : re-lier chaque affirmation de 20A–25 à sa sortie fraîche sur HEAD, capturer les preuves manquantes à 1024/1920, et réconcilier l'index. Aucun Go/engine/UI nouveau, sauf si une affirmation est réfutée (auquel cas le fix atterrit inline avec fault proof).

## 2. Ce qui existe déjà (vérifié en source)

- Paquets : `docs/verification/2026-11-01-slice-20a.md`, `2026-11-01-slice-20c.md`, `2026-11-02-slice-21.md`, `2026-11-04-slice-22.md`, `-23.md`, `-24.md`, `-25.md`.
- Lignes `PENDING` : 7 (20A) + 1 (20C) + 1 (21) + 7 (22) + 5 (23) + 4 (24) + 3 (25) = 28 propositions.
- Captures 21 : `docs/img/shots/2026-11-02/` (37 fichiers, dont 4 à 390).
- HEAD `9df16a8` : `go test ./...` ok (tous packages), `python tools/verify.py` 6/6, `npm test` 59 fichiers / 283 tests, build 353.7 KB gz (vérifié pendant le brainstorming).

## 3. Design

### 3.1 Tâche A — 20A packet binding

Re-run `2026-11-01-slice-20a.md` commandes 1–8 sur HEAD. Pour chaque ligne `PENDING`, binder le nom du test (`TestVerdictNeverGreenWithoutGrade`, `TestReportVerdictRefusesGreenWithoutLoad`, `TestReportLinuxOnlyDownEmitsLinuxIngressAndNoRouterOS`, `TestBuildEconomicsNumbersPerModel`, `TestBuildEconomicsNumbersNotInventedWhenUnmeasured`, `TestEconomicsNumbersJSONKeys`, `TestMatrixConfigFreezesActiveTier`, `TestTimelineBucketsBySecondAndKeepsGaps`, `TestTimelineCapIsTotalPlusFive`, etc.) + sortie fraîche. Déplacer vers `VERIFIED` seulement si chaque test passe à HEAD (pas à `df89c82`).

### 3.2 Tâche B — 20C packet binding

Idem pour la ligne TUI (16 tests).

### 3.3 Tâche C — 21 packet reconciliation

Re-run `2026-11-02-slice-21.md` commandes 1–5. Réconcilier les 12 propositions contre HEAD (certaines ont pu devenir obsolètes après 22–25). Archiver les 4 captures 390 hors du set revendiqué (mobile abandonné).

### 3.4 Tâche D — 22–25 evidence sweep + captures

Re-run les quatre paquets en une passe sur HEAD. Capturer `docs/img/shots/2026-11-04/slice22-*`, `slice23-*`, `slice24-*`, `slice25-*` à **1024 et 1920 uniquement** : (22) suppression de `Protéger ce lien`, (23) défaut 220 ms en réglages + pill BBRv3, (24) `Maxi avant` / `sans réglage`, (25) tooltip nommant la barre. Binder les 19 propositions.

### 3.5 Tâche E — Index + registre + annexes

20A/20B/20C cessent de lire « Superseded (builder pass) » en ambiguïté ; 22–25 passent à `Done (vérifié par 26′)` ; les paragraphes annexe manquants pour 20A/20C sont émis ; une ligne `PENDING` pour 26′ elle-même.

## 3.6 Non-objectifs

- Cockpit NEW1 (3 cartes + CTA Fogg) : refusé en slice 22 §3.5, inchangé.
- Largeurs mobiles, tuiles profils, renommage `PREUVE VÉRIFIABLE` : réfuté avec preuve en slice 23, inchangé.
- Aucun Go/engine/UI nouveau sauf réfutation (fix inline + fault proof).
- Aucune valeur live, CHR, bench, audit, eBPF, `.dist`, BBRv3.
