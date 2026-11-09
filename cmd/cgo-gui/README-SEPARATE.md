# cgo-gui — dépôt séparé (pointeur)

La GUI vit dans son propre dépôt : [Realms4239/cgo-gui](https://github.com/Realms4239/cgo-gui)
(module `github.com/Realms4239/cgo-gui`, `go 1.25`, créé le 2026-11-09,
`v0.1.0` publié, CI stub + Windows verte). Ce dossier ici est un pointeur :
aucun code n'a été déplacé hors de cette section de travail.

Spécification complète : [`docs/runbooks/gui-repo.md`](../../docs/runbooks/gui-repo.md)
(design `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §6).

Rappels rapides (détail = runbook) :

- `go.mod` propre : Fyne `v2.8.1`, `golang.org/x/sys v0.36.0`, sans
  `replace` ; le stub n’a aucun `require` (stdlib seule).
- `core.go` partagé sans tag ; vues `windows` (Win32 brut,
  `LockOSThread` obligatoire) et `linux && cgo` (Fyne).
- Stub `//go:build !linux || !cgo` : redirection honnête + `exit 2`,
  jamais d’import Fyne — c’est ce qui rend `CGO_ENABLED=0` vert.
- Aucun import `internal/` ni `github.com/Realms4239/cgo/*` : le
  dépôt séparé parle à `cgo`/`cgo-kit` en sous-processus.
- SemVer indépendant : `v0.1.0` au départ, sans lien avec `VERSION`
  du monolithe.

Aujourd’hui, `cmd/cgo-gui` reste dans le monolithe. Son échec
préexistant en `GOOS=linux CGO_ENABLED=0` est isolé à ce paquet seul
(`fyne.io/fyne/v2/app` → `go-gl/gl` sans fichiers) ; `cmd/cgo`,
`cmd/cgo-chr` et `cmd/meteolink` compilent en `CGO_ENABLED=0`.
