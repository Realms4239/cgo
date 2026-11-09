# Verifier packet — per-module cutover (§5.4–5.8, runbook `tag-cutover.md` step 8)

Rôle : le **builder** ne certifie pas son propre chiffre — le
**vérifieur** relance ces commandes et possède les lignes `VERIFIED`.
Colonne "attendu" : sorties minimales que le vérifieur doit voir ;
toute déviation = refus de la signature. Une ligne ne passe `PENDING` →
`VERIFIED` qu'après ce re-run (git `main` HEAD vérifié, `audit`/`api` suite
décastante lente incluse).

**Prérequis :** accès réseau au proxy `proxy.golang.org` (ou `GOPROXY=direct`)
pour les téléchargements de modules déjà publiés ; toolchain `go 1.25+`.

## 0. État du dépôt

```bash
git log --oneline -1   # attendu : le HEAD publié de main
```

## 1. Les 15 tags sont publiés et fetchables (règle des sous-répertoires)

```bash
for m in model metrics qdisc results figures cite exporter guard iface \
          remediate attest doctor probe audit api; do
  go list -m github.com/Realms4239/cgo/$m@latest 2>&1 | grep -E '^(github.com/Realms4239/cgo/\S+ v0\.1\.(0|1))$'
done
```

Attendu (15 modules, aucune flèche) : chaque ligne
`github.com/Realms4239/cgo/<nom> v0.1.N`. `model/v0.1.0` est le seul
rémité (tag dud pré-move, jamais consommé) — `model` résout en `v0.1.1`,
c'est le comportement attendu.
Le tag dud `model/v0.1.0` (git object `f7c2baf`) n'est pas supprimé ;
vérifier qu'il existe sur le remote vaut preuve d'honnêteté :

```bash
git ls-remote origin 'refs/tags/model/v0.1.0'
# attendu : 1 ligne `<hash>\trefs/tags/model/v0.1.0`
```

## 2. Aucun `replace` ne survit dans le chemin du module

```bash
grep -rn "=> .*/" go.mod $(ls -d */go.mod 2>/dev/null) 2>/dev/null || echo "aucun replace local"
```

Attendu : `aucun replace local`.

## 3. Suite de chaque module (isolée, sans `replace`)

```bash
for m in api audit cite doctor exporter figures guard iface metrics \
          model probe qdisc remediate results attest; do
  (cd $m && go test ./... -count=1) || echo "ECHEC $m"
done
```

Attendu (15 modules) : `ok github.com/Realms4239/cgo/<nom>` pour chacun.
(`audit` met en général 3 minutes.)
< !-- NB: les `ok` ci-dessus incluent les téléchargements des dépendances
      déjà publiées (`model v0.1.1`, `metrics v0.1.0`, …) — par exemple
      `audit` et `api` faut `probe v0.1.0` du proxy. -->

## 4. Porte racine

```bash
python tools/verify.py             # 6 passed, 0 failed
go test ./...                      # racine verte
```

Attendu : `6 passed, 0 failed, 0 not run`; `ok` pour tous les paquets
racine.

## 5. Ce que ce packet ne prouve PAS

- Les runs live du banc (aucun bench/CHR/BBR lancé — grants pendants).
- La suppression du tag dud `model/v0.1.0` (décision du vérifieur :
  ne jamais supprimer un tag sans analyse fraîche du cache proxy).
- Les tags plus bas que `v0.1.0` (semver `0.y.z` = API publique non
  garantie, SemVer §4) ; `v1.0.0` est la promesse.
- La culture `book/` (gelée prescrite, intacte — `python tools/verify.py`
  le scelle).

## Signature (par le vérifieur uniquement)

`docs/VERIFY.md` : signer la ligne `15 modules` (et lignes report → RAPPORT
si présentes) en y écrivant la commande, la date, l'hôte et le résultat
rencontré. Ne pas réécrire le corps des entrées existantes.
