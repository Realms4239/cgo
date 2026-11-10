# Runbook — dépôt séparé `cgo-gui` (scaffolding)

Ce document est la spécification du dépôt séparé `cgo-gui` : il n’y a
aucun code déplacé ici. Il décrit la structure cible, le `go.mod`,
le cœur partagé dual-view, le stub `CGO_ENABLED=0` et les contraintes
de compilation. Aucune exécution live, aucun run, aucun bench.

Contexte : `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md`
§6 (dual-view, échec bloquant, structure cible). Pointeur dans le
monorepo : `cmd/cgo-gui/README-SEPARATE.md`.

## État (2026-11-09)

**Dépôt créé et vert :** [Realms4239/cgo-gui](https://github.com/Realms4239/cgo-gui),
module `github.com/Realms4239/cgo-gui`, `go 1.25`, départ SemVer `v0.1.0`
(tag `v0.1.0` poussé le 2026-11-09 aux 06:24 UTC). CI GitHub Actions
`ci` (two jobs) : `stub` (`GOOS=linux CGO_ENABLED=0` : `go vet` +
`go build`) et `windows` (`windows-latest` : `go build` + `go test`) —
tous deux verts sur le run `42572824714`. Vue Linux Fyne native :
preuve non faite ici (toolchain CGO + écran, pas d'hôte Linux au poste —
reste non prouvé, documenté dans le README du dépôt).

Structure implémentée, avec deux corrections de vanille portées du
runbook : `win/`, `linux/`, `stubs/` ne sont PAS des sous-dossiers (Go
ne peut pas partager `package main` entre dossiers pour un seul binaire) ;
ce sont des fichiers du paquet `main` à gardes disjointes, exactement
comme dans le monolithe : `main.go`/`win32.go`/… (`//go:build windows`),
`fygui_linux.go` (`//go:build linux && cgo` + miroir de test
`fytest_linux_test.go` à la même garde — CI stub en a été la preuve :
sans elle, `undefined: fyApp` sous `GOOS=linux CGO_ENABLED=0`),
`stub.go` (`//go:build !windows && (!linux || !cgo)`, sortie 2 —
table de couverture corrigée : sous `GOOS=windows`, `main.go` couvre
déjà le binaire, le stub ne doit jamais compiler en parallèle).
`kit/` (client fin par sous-processus `cgo kit …`, zéro import du
monolithe), `vm/` (copié d'`internal/vm`, pur Go) et `pkg/kit-client`
n'existe pas : le client est `kit/` directement.

Le monolithe garde son `cmd/cgo-gui` (copie, pas suppression — runbook :
« il ne déplace pas le code dans cette tâche ») ; les deux codebases
divergent à partir de cette date. La « dual-view parity » (S4) est
portée par fixtures (`gui_test.go`, `chain_test.go`, `vnet_parity_test.go`,
`health_telemetry_test.go` passent dans les deux dépôts) ; le client
sous-processus exécute les mêmes verbes `cgo kit` aux mêmes codes de
sortie que le CLI nativ (vérification fonctionnelle restante : run
live de la VM, jamais lancé ici).

- `cmd/cgo-gui/core.go` : cœur partagé SANS tag (pur Go : `internal/kit`,
  `internal/vm`, stdlib). Deux vues pilotent les mêmes actions.
- Vue Windows (`//go:build windows`) : `main.go`, `win32.go`,
  `window.go`, `actions.go`, `elevate.go`, `status.go` — Win32 brut
  via `golang.org/x/sys/windows` + LazyDLL (`user32`/`gdi32`), aucune
  Fyne. Pompe + fenêtres sur un seul thread OS (`runtime.LockOSThread()`
  dans `window.go:run`).
- Vue Linux (`//go:build linux` SEUL, sans garde `cgo`) :
  `fygui_linux.go`, `fytest_linux_test.go` — `fyne.io/fyne/v2` +
  `app` + `container` + `widget`. Sans `DISPLAY`/`WAYLAND_DISPLAY`,
  refus poli vers `cgo kit tui`.
- Dépendances (monorepo `go.mod`) : `fyne.io/fyne/v2 v2.8.1`,
  `golang.org/x/sys v0.36.0`.
- Échec préexistant (isolé, voir §6.2 du design) : sans garde
  `linux && cgo` ni stub, `GOOS=linux CGO_ENABLED=0 go build ./cmd/...`
  échoue sur `cmd/cgo-gui` seul (`go-gl/gl` : « build constraints
  exclude all Go files »). Les autres paquets `cmd/cgo`,
  `cmd/cgo-chr`, `cmd/meteolink` passent. Ce runbook corrige en
  isolant, il ne déplace pas le code dans cette tâche.

## Dépôt cible

```text
cgo-gui/  (dépôt propre + go.mod propre ; ./... du monolithe vert en CGO_ENABLED=0)
  go.mod                          # module github.com/Realms4239/cgo-gui, go 1.25
  core.go                         # partagé pur Go, SANS tag (les deux vues)
  win/                            # vue Windows (//go:build windows) :
                                  #   main.go, win32.go, window.go, actions.go, elevate.go
  linux/                          # vue Linux Fyne (//go:build linux && cgo) :
                                  #   fygui_linux.go
  stubs/                          # stub sans CGO (//go:build !linux || !cgo) :
                                  #   redirect honnête + exit 2, stdlib seule
  kit/  (ou pkg/kit-client)       # client fin vers cgo/cgo-kit (sous-processus),
                                  #   JAMAIS import internal/kit
  README.md                       # prérequis Fyne (GCC/X11/OpenGL/pkg-config,
                                  #   à documenter une fois récupérés) + CGO
  Makefile                        # go build ./... (CGO_ENABLED=1 pour la vue
                                  #   Linux) ; GOOS=windows pour la vue Win
```

## `go.mod` du nouveau dépôt

```go
// go.mod — module propre de la GUI ; Fyne v2.8.1, go 1.25.
// Le stub n’a pas de require : seuls les chemins importés sont listés.
module github.com/Realms4239/cgo-gui

go 1.25

require (
	fyne.io/fyne/v2 v2.8.1
	golang.org/x/sys v0.36.0
)
```

- Aucun `replace`, aucun `exclude`/`retract` à la naissance.
- Aucun import de `github.com/Realms4239/cgo` ni de ses `internal/`
  (règle Go `internal` : import depuis l’extérieur interdit).
- `core.go` actuel importe `internal/kit` et `internal/vm` : dans le
  dépôt séparé, ces deux usages deviennent composition sous-processus
  (`cgo.exe kit …` à côté de la GUI, déjà la composition actuelle) ou
  client fin public — jamais `internal/`.
- Critère : `go list -m all` sans flèche locale, sans `internal/`,
  sans Fyne dans le monolithe après extraction (spec §6.7).

## `core.go` — cœur partagé, deux vues

- `core.go` : SANS tag de compilation. Aucune API Win32, aucun widget.
  Inventaires `groupActions`/`accessActions`/`deployActions`/
  `controlActions`, `buttonAction(id) → (kind, args)`, `kit.IsKnownSuiteVerb` (la table `internal/kit/verbs.go` — 30 verbes, une seule source),
  `diagGUI`, `probeStatus`, `scanVMRows`, `cfgCache` (45 s, anti-vmrun),
  miroir journal `cgo-gui-<date>.log`.
- Vue Windows (`//go:build windows`) : Win32 brut. Fenêtre 940×700,
  `LockOSThread` OBLIGATOIRE dans `run()` — sans ça Go migre la
  goroutine, la file de messages Win32 se fige (voir `window.go:20-73`).
  `PostMessage` seul pour le retour fond → UI, jamais de touché direct.
- Vue Linux (`//go:build linux && cgo`) : Fyne, mêmes boutons et même
  garde anti-double-clic. `fyne` n’apparaît QUE derrière cette garde.
  Sans écran (`DISPLAY`/`WAYLAND_DISPLAY` vides), refus + redirection
  `cgo kit tui`.
- Parité : `buttonAction` et `kit.IsKnownSuiteVerb` sont partagés et
  testables sans écran ; la table `SuiteVerbs` lue par les deux
  exécutants est épinglée par `cmd/cgo-gui/verbs_parity_test.go`
  (lecture des sources — un verbe ajouté dans un seul exécutant
  fait échouer le test).

## Contraintes de compilation + stub `CGO_ENABLED=0`

| Config | Garde | Résultat |
| --- | --- | --- |
| Windows (tout `CGO_ENABLED`) | `//go:build windows` | vue Win32 (pur Go, pas de Fyne) |
| Linux + CGO | `//go:build linux && cgo` | vue Fyne (GLFW/OpenGL, exige écran) |
| Linux sans CGO, autre GOOS | `//go:build !linux \|\| !cgo` | stub stdlib seule, jamais d’import Fyne |

Correction par rapport à l’existant : `fygui_linux.go` est aujourd’hui
`//go:build linux` seul. Le dépôt séparé porte `linux && cgo` et le
stub ci-dessous (stdlib seule) :

```go
//go:build !linux || !cgo

// Stub sans CGO : pas de Fyne, pas de GL. Redirection honnête + exit 2.
package main

import (
	"fmt"
	"os"
)

func main() {
	// Même message que la vue Fyne sans écran : on renvoie vers la TUI.
	fmt.Fprintln(os.Stderr, "pas d'écran (ni DISPLAY ni WAYLAND_DISPLAY) — ici : ./cgo kit tui")
	os.Exit(2)
}
```

Le `main()` stub coexiste avec les deux vrais `main()` car les gardes
sont disjointes. `go vet`/`go build` en `CGO_ENABLED=0` retombent
sur le stub et passent.

## Versionning SemVer indépendant

- Tags propres du dépôt `cgo-gui` : départ `v0.1.0` (`0.y.z` : API
  non garantie). `v1.0.0` seulement quand le contrat bouton→verbe
  est déclaré.
- Indépendant de `VERSION` du monolithe (`1.3.5`) et des modules
  (`v0.1.0` via `require`). Aucun tag déplacé après publication.
- `KitVersion` == tag == `cgo version` doivent concorder (source
  unique, §5.4 du design).

## Règle de lien (liens entre dépôts)

- `cgo-gui` ne lie rien du monolithe : ni `internal/kit` ni
  `internal/vm`, ni `github.com/Realms4239/cgo/*`.
- Dialogue = sous-processus `cgo`/`cgo-kit` voisin (composition
  existante `cgo.exe kit …`) ou client fin public versionné.
- Tests de la GUI : fixtures seules, jamais de VM live, jamais de
  `book/data/runs` touché.

## Vérification (dépôt séparé)

```bash
# Stub seul (CI headless, cross depuis Windows) :
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go vet ./...
# Vue Windows :
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...
# Vue Linux Fyne (toolchain native + écran requis) :
CGO_ENABLED=1 go build ./...
```

Acceptations (design §6.6) : S1 `./...` vert croisé (Linux + Windows +
vet) ; S2 dépôt séparé, monolithe sans `cmd/cgo-gui` dans `./...`,
build autonome `GOPROXY=off` ; S3 prérequis Fyne documentés avec
source primaire (à récupérer, sinon `unverified`) ; S4 parité
dual-view (22 boutons, fixtures partagées).

## Références primaires

- `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §6
  (analyse Fyne, échec bloquant, structure cible, slices).
- `docs/superpowers/plans/2026-11-08-cgo-supplychain-gaps.md` Task 4.
- `go.dev/ref/mod` — Go Modules Reference (règle `internal`, guards).
- `semver.org` — Semantic Versioning 2.0.0 (§4 `0.y.z`).
