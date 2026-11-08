# Runbook — étiquettes par module + bascule (SemVer)

Ce document est la liste de contrôle du cutover par module : chaque
module (`model`, `metrics`, `results`, `figures`, `qdisc` et, à terme,
chaque outil extrait) suit exactement ces 8 étapes. Aucune étape ne se
saute ; aucun `replace` ne survit à un tag ; aucun tag publié ne se
déplace jamais (SemVer 2.0.0 §3, Go Modules `go.dev/ref/mod`).

Contexte et règles produites : `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md`
§5.4–5.8, §9. Déclaration publique du contrat : `docs/deploy.md` §7.2.

## État du monorepo aujourd'hui

- `go.mod` racine : `require github.com/Realms4239/cgo/<module> v0.1.0`
  + `replace => ./modules/<module>` pour chaque module extrait.
- Chaque `modules/<nom>/go.mod` : `module github.com/Realms4239/cgo/<nom>`,
  `go 1.25`.
- Version du monolithe (`VERSION`, `1.3.5`) et versions des modules
  (`v0.1.0` via `require`) : deux lignes indépendantes.

Ordre de cutover (topologique, dépendances d'abord) :
`model` → `metrics` → `qdisc` → `results` → `figures`, puis chaque outil
extrait dans l'ordre de dépendance. Un module à la fois.

## Les 8 étapes (cutover atomique)

```text
1. Publier le nouveau tag du module (`<module>/v0.1.0`, ex. `model/v0.1.0`, immuable)
   v0.1.0 = première release, API publique non garantie
   (SemVer §4) ; 1.0.0 est la promesse.
   Le tag est signé ; son contenu ne change jamais (§3).
   On ne déplace pas un tag publié : on publie `<module>/v0.1.1`.

2. Monter le require dans le go.mod du consommateur
   (sans aucun replace dans le tag)
   Découpe du replace de CE module dans go.mod racine
   (et dans les go.mod des consommateurs) : la résolution
   doit viser le module publié, jamais le chemin local.

3. go mod tidy

4. Vérifier : go list -m all
   Le consommateur doit résoudre <module> v0.1.0, sans
   replace, sans pkg/* internes, sans fyne sauf s'il est
   la vue GUI (modules purs : stdlib seule).

5. goreleaser release --clean --dry-run
   Vérifie checksums.txt + SBOM CycloneDX présents
   (.goreleaser.yaml, §7.1 de docs/deploy.md).
   Si goreleaser/syft/cosign sont absents sur cet hôte :
   config-complete, première CI tag = preuve.

6. Mettre à jour chaque entrée keep_version
   (mise.lock / package.json / pin Vite, le cas échéant).

7. Lancer make verify-fast + python tools/verify.py --selftest
   Le module isolé passe (go test ./... depuis modules/<nom>) ;
   la suite racine reste verte ; les six gates sont prouvées
   (selftest : les six caught).

8. Le vérifieur relance les mêmes commandes
   (le builder ne certifie pas son propre chiffre) avant
   qu'une ligne VERIFIED ne soit écrite.
```

Le cutover est atomique : `code` + `require` + `tidy` + `list` +
dry-run + lock + verify-fast + vérifieur. Un cutover partiel
n'est pas un commit.

## Garde-fous (FAIL immédiat)

- Un `replace` restant dans un tag publié : FAIL (Go Modules :
  `replace` ne lie que le module principal, il ne peut pas être publié).
- Un tag déplacé ou édité après publication : FAIL.
- Des données gelées réécrites (`book/data/runs`, schémas CSV) : FAIL.
- `go test ./...` rouge dans le module ou à la racine : STOP, aucun tag.
- Un dry-run en réseau sur la VM de banc ou un run live : refusé ;
  ce runbook ne requiert aucun run réel.

## Rollback (une commande)

Revert du commit de bascule + `go mod tidy` + `go list -m all` :
aucun tag ne bouge, la version publiée reste installable, le
monorepo revient au `require` + `replace` local précédent.

## Exécution à sec (dry-run) — procédure

Le dry-run prouve que le retrait d'un `replace` est un no-op :
le consommateur compile et teste à l'identique contre le module
versionné. Il se fait dans un clone temporaire, jamais dans le
checkout de travail, et ne pousse rien.

```bash
# 1. Clone temporaire (branche incluse, aucun push)
git clone --no-local . /tmp/cgo-cutover-dryrun
cd /tmp/cgo-cutover-dryrun

# 2. Tag du module (ex. model/v0.1.0, non signé = suffisant ici)
git tag model/v0.1.0 -m "cgo/model v0.1.0"

# 3. Référence : résolution AVANT retrait du replace
go list -m all > /tmp/cgo-before.txt

# 4. Retirer le replace de CE module dans go.mod racine
#    (ex. la ligne  github.com/Realms4239/cgo/model => ./modules/model)
go mod edit -dropreplace github.com/Realms4239/cgo/model

# 5. Résolution APRÈS + vérification no-op
go list -m all > /tmp/cgo-after.txt
go test ./modules/model/...
go test ./...      # ou la suite racine concernée
diff /tmp/cgo-before.txt /tmp/cgo-after.txt
```

Critère de succès : `diff` vide sur la ligne du module (même version
résolue, sans `=> ./modules/model`), tests verts avant et après.
Sans publication préalable, la résolution échoue après retrait
(`unknown revision`) : c'est le signal que l'étape 1 (publication)
n'est pas faite. Nettoyage : `rm -rf /tmp/cgo-cutover-dryrun`
(le tag meurt avec le clone, jamais poussé).

## Résultat du dry-run

Date UTC : 2026-11-08 · Hôte : Windows (worktree operator-cockpit)
Commit de base : voir `git log --oneline -1` du checkout de travail.
Commandes (clone temporaire, site-packages hors du checkout) :

```bash
git clone --no-local . /tmp/cgo-cutover-dryrun-task3
cd /tmp/cgo-cutover-dryrun-task3
git tag model/v0.1.0 -m "cgo/model v0.1.0"
go list -m all
go mod edit -dropreplace github.com/Realms4239/cgo/model
go list -m all
go test ./modules/model/... -count=1
```

Sortie (abrégée, lignes pertinentes) : `go list -m all` avant retrait
liste `github.com/Realms4239/cgo/model v0.1.0 => ./modules/model` ;
après `-dropreplace`, la résolution échoue :
`go: github.com/Realms4239/cgo/model@v0.1.0: invalid version:
unknown revision model/v0.1.0`.
Le `replace` est donc porteur tant que le tag n'est pas publié et
récupérable via le proxy ; le retrait n'est un no-op qu'APRÈS
publication (`git push origin model/v0.1.0` + `go mod download`
réussissant — étape 1→3 de la procédure, jamais poursuivie ici).
Ce que le dry-run prouve c'est la condition de bascule : publier
d'abord, retirer ensuite, et vérifier `go list -m all` sans flèche
locale. `go test ./...` dans `modules/model` passe avant comme
après (`ok github.com/Realms4239/cgo/model`) : le module isolé n'est
pas affecté par le retrait.
Le clone temporaire a été supprimé ; aucun tag n'a été poussé.

Limites assumées : `go test ./...` complet non lancé sur cette
étape (la cible est le no-op de résolution) ; les gates
complètes appartiennent à la tâche 6 et au vérifieur.

## Références primaires

- `semver.org` — Semantic Versioning 2.0.0 (§1 API déclarée, §3
  immuabilité, §4 `0.y.z`, §6–8 PATCH/MINOR/MAJOR, `vX.Y.Z` = tag).
- `go.dev/ref/mod` — Go Modules Reference (path = ID + préfixe,
  `go` minimum depuis 1.21, `require`/`replace`/`exclude`/`retract`,
  MVS, règle `/vN` pour v2+).
- `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §5 + §9.
