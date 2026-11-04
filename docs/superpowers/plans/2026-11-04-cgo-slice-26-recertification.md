# Slice 26′ — Re-certification exhaustive 20A–25

*Plan écrit le 2026-11-04 par le builder. Spec : `docs/superpowers/specs/2026-11-04-cgo-slice26-recertification-design.md`. Exécution sur la branche `rebuild/plan1-economics-audit`, HEAD de départ `9df16a8`.*

## Contrat

Une seule tranche, cinq tâches, aucune fonctionnalité neuve. Le builder produit
l'évidence ; le vérificateur indépendant re-exécute et signe ; le builder
transcrit la signature sans la modifier.

Toute commande est une commande de test ou de porte. Aucun audit, campagne,
banc, CHR, eBPF, `.dist` ou BBRv3 live. Aucune capture n'invente un état :
chaque image est prise sur l'application servie localement en mode observation.

## T0 — Spec, plan, index

Fait : spec `d0ee18d`. Ce plan et sa ligne d'index sont committés ensemble.

```bash
git add docs/superpowers/plans/2026-11-04-cgo-slice-26-recertification.md docs/superpowers/plans/README.md
git commit -m 'docs: le plan slice 26′ re-certification exhaustive'
```

## T1 — Tâche A : lier les sept lignes 20A

**But.** Chaque ligne `PENDING` de 20A reçoit le nom de son test, la commande
exacte, la sortie fraîche à HEAD et la limite inchangée (fixtures seulement).

**Commandes (hôte Windows, Go 1.25) :**

```bash
go build ./...
gofmt -l pkg cmd
go test ./pkg/metrics ./pkg/audit -count=1
go test ./pkg/remediate ./pkg/audit -count=1
go test ./pkg/campagne ./pkg/results ./pkg/api -count=1
go test ./pkg/api -run Duel -v
go test ./pkg/api -run "RunStartRefusesEscaping|RunStartSubmatrix|RunOptsReach|RunStartProvenance" -v
go test ./pkg/campagne -run TestResumeKeepsTheFrozenTier -v
go test ./pkg/results -run TestGroupRunsExcludeInvalidOnlyRuns -v
python tools/verify.py
```

**Preuves par ligne :**

| # | Ligne 20A | Tests qui portent la preuve |
|---|---|---|
| A1 | verdict cascade | `TestVerdictNeverGreenWithoutGrade`, `TestReportVerdictRefusesGreenWithoutLoad` |
| A2 | taux et interface LAN supposée | `TestReportLinuxOnlyDownEmitsLinuxIngressAndNoRouterOS` |
| A3 | nombres économiques | `TestBuildEconomicsNumbersPerModel`, `TestBuildEconomicsNumbersNotInventedWhenUnmeasured`, `TestEconomicsNumbersJSONKeys` |
| A4 | palier gelé + garde d'écriture | `TestMatrixConfigFreezesActiveTier`, `TestCostARPerHUsesFrozenTier`, `TestRunEventCostUsesFrozenTier`, `TestCostTierRefusedWhileCampaignRuns`, `TestResumeKeepsTheFrozenTier` |
| A5 | timeline par seconde | `TestTimelineBucketsBySecondAndKeepsGaps`, `TestTimelineCapIsTotalPlusFive`, `TestAuditLiveAndStop` |
| A6 | duel `n ≥ 3` | `TestGroupRunsExcludeInvalidOnlyRuns` + `go test ./pkg/api -run Duel -v` |
| A7 | `run_id` échappant | `TestRunStartRefusesEscapingRunID` (+ les trois tests `RunStart*`) |

**Limites conservées.** Aucune observation live ; le résidu du kill dur avant
`Freeze` reste consigné ; la ligne de preuve CHR reste non formulée.

**Acceptation.** Les dix commandes sortent 0 ; chaque test nommé apparaît
`PASS` ou dans un paquet `ok` ; `git status --short` est identique avant/après.

## T2 — Tâche B : lier la ligne 20C

```bash
go build ./...
gofmt -l cmd/cgo
go vet ./cmd/cgo
go test ./cmd/cgo -count=1
go test ./cmd/cgo -run "TestAudit|TestVerdictHeader|TestWrap78|TestDeployment|TestProofPane|TestWriteRemediation|TestScriptFor|TestViewShows|TestRenderedViews" -v -count=1
python tools/verify.py
```

**Preuve.** 16 tests nommés, dont `TestRenderedViewsStayWithinEightyColumns`,
`TestWriteRemediationRefusesTraversal` et
`TestAuditViewTransportFailureNamesTheAPI`. Les cinq clauses de la ligne 20C
sont citées une par une dans le paquet.

**Limites conservées.** Miroir rendu depuis une fixture ; OSC 52 tenté, jamais
livré ; aucune ligne 20A ne bouge par cette tâche.

## T3 — Tâche C : réconcilier la tranche 21 et les quatre lignes antérieures

**But.** Le paquet 21 proposait douze dispositions ; aucune n'a été appliquée.
Les quatre lignes antérieures à 20A reçoivent leur disposition motivée :

| Ligne | Proposition builder | Preuve |
|---|---|---|
| Slice 8 audit report surface and citation | `SUPERSEDED` | La surface a été supprimée par la refonte 21 ; le code de la vue n'existe plus. |
| Interface v2 (four modes) and the academic title | `SUPERSEDED` | Rail et ribbon remplacés par TopNav et le cluster d'état ; le titre reste mais la surface n'existe plus. |
| BBRv3 engine | `BLOCKED` | Aucun noyau 7.1 sur cet hôte ni sur le banc 6.8 ; rien observé dans un socket. |
| Trace-driven `.dist` engine | `BLOCKED` | Aucune trace de terrain ≥ 30 échantillons baseline ; les deux traces `SMOKE-TEST` font 19 et 25 lignes. |

**Commandes :**

```bash
cd web/frontend && npm run build && npm test
python tools/verify.py
python book/tools/audit-archive.py
```

**Captures 21.** Le set `docs/img/shots/2026-11-02/` contient quatre images à
390 px (`*-390.png`). La plage acceptée est 1024–1920 : ces quatre fichiers
sortent du set revendiqué. Ils restent sur disque comme historique, et le
paquet 21 est amendé d'une ligne qui le dit — aucune image n'est supprimée.

**Acceptation.** Les douze propositions ont une disposition signée ; aucune
ligne 21–25 ne porte encore `PENDING` après la signature du vérificateur.

## T4 — Tâche D : sweep 22–25 et captures 1024/1920

**Commandes communes :**

```bash
go test ./... -count=1
python tools/verify.py
python book/tools/audit-archive.py
cd web/frontend && npm test && npm run build
```

**Captures.** Servir l'application locale :

```bash
npm run build                      # web/frontend
go build -o cgo.exe ./cmd/cgo      # racine
./cgo.exe --serve --addr 127.0.0.1:9091 --tls=false --mode observe
```

Puis, avec `playwright-core` (déjà installé) et un script de capture en
lecture seule, produire dans `docs/img/shots/2026-11-04/` :

| Tranche | Fichier | Ce que l'image doit montrer |
|---|---|---|
| 22 | `slice22-audit-remediation-suppressed-{1024,1920}.png` | Un rapport archivé sans charge mesurée : aucune carte « Protéger ce lien », l'état « observation seule » visible. |
| 23 | `slice23-settings-deadline-{1024,1920}.png` | Tiroir Réglages : `220 ms` comme défaut. |
| 23 | `slice23-bench-bbrv3-{1024,1920}.png` | Banc : pastille `BBRv3 non détecté`. |
| 24 | `slice24-bench-edge-labels-{1024,1920}.png` | Bord vivant : `Maxi avant` et `sans réglage (file simple)`. |
| 25 | `slice25-bench-tooltip-{1024,1920}.png` | Le bouton `Maxi avant` ; le texte de l'infobulle est extrait du DOM dans `slice25-tooltip.txt` (un `title` natif n'est pas peint dans une capture). |

**Limites des captures.** Mode observation Windows ; aucune mesure live n'est
provoquée ; aucun bouton qui démarre un banc ou un audit n'est cliqué ; les
images prouvent un rendu, pas une mesure.

**Acceptation.** Dix PNG réels, non vides, à 1024×768 et 1920×1080 ; le texte
du `title` de slice 25 est dans le `.txt` et correspond exactement à la chaîne
committée ; aucun fichier `dist/index.html` stagé.

## T5 — Tâche E : index, registre, checkpoints, annexes

- `docs/VERIFY.md` : les dix-sept lignes `PENDING` reçoivent la disposition du
  vérificateur (sept 20A, une 20C, une 21, quatre 22–25, quatre antérieures).
- `docs/superpowers/plans/README.md` : 20A/20B/20C passent à
  `Superseded (verified by 26′)` ou `Blocked` selon la signature ; 21–25 à
  `Done (builder pass + verifier signature)` ; ligne 26′ ajoutée.
- `docs/VERIFY.md` : une ligne 26′ elle-même.
- `docs/CHECKPOINTS.md` : une entrée avec commande, date, hôte, résultat et
  non-prouvés.
- `book/content/09-annexes.js` : un paragraphe pour 26′, working tree seulement.
- `docs/verification/2026-11-04-slice26-recertification.md` : le paquet consolidé
  (commandes, sorties, captures, dispositions, bloc signature).

## T6 — Vérificateur indépendant

Un agent distinct reçoit le paquet consolidé, re-exécute les commandes sur
l'arbre courant et remplit le bloc signature. Le builder ne remplit jamais ce
bloc ; il transcrit la signature telle quelle dans `docs/VERIFY.md`.

## Acceptation finale

- `python tools/verify.py` → `6 passed, 0 failed, 0 not run`.
- `go test ./... -count=1` → tout `ok`.
- `npm test` → tous les fichiers/tests passés ; `npm run build` → exit 0.
- `python book/tools/audit-archive.py` → deux lignes `OK`.
- Dix captures 1024/1920 pour 22–25.
- Zéro ligne `PENDING` pour 20A–25 ; chaque ligne porte date, hôte, commande
  et signature du vérificateur.
- Aucune modification de `book/data`, aucun schéma CSV, aucun merge `main`.

## Ce que cette tranche ne prouve pas

- Aucune valeur live : les tests et les captures sont des fixtures et un
  rendu local en mode observation.
- Le cockpit NEW1 reste différé par conception ; les largeurs mobiles restent
  hors périmètre ; `PREUVE VÉRIFIABLE` et la troncature des tuiles restent
  réfutés, inchangés.
- Les résidus déjà consignés (kill dur avant `Freeze`, OSC 52, CHR) restent
  des résidus : cette tranche ne les ferme pas.
