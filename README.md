# Meteolink [![version](https://img.shields.io/badge/version-1.0.6-blue)](VERSION) [![license](https://img.shields.io/badge/license-MIT-green)](LICENSE) [![go](https://img.shields.io/badge/go-1.25-%2300ADD8)](go.mod)

## Qu'est-ce que c'est ?

Meteolink est un banc d'essai réseau open source, temps réel et visualiseur interactif qui s'exécute dans un terminal sur les systèmes *nix ou directement dans votre navigateur. Conçu pour les liens d'accès contraints (4G/5G, fibre, VSAT), il fournit à la volée des preuves AQM/BBR rapides et vérifiables. Meteolink audite votre lien côté client, le rejoue sur un banc Linux reproductible et présente les données directement dans le terminal ou via un tableau de bord HTML live — aucune promesse fournisseur, seulement des CSV gelés avec SHA-256.

Plus d'infos sur : [https://github.com/Realms4239/cgo](https://github.com/Realms4239/cgo).

## Fonctionnalités

Meteolink rejoue des profils de lien et affiche les données dans le terminal ou le tableau de bord. Fonctionnalités :

- **Entièrement temps réel**  
  Tous les panneaux et métriques sont rafraîchis toutes les 100 ms sur le flux `SSE` (10 Hz) et toutes les 250 ms sur la `TUI`. Le `live-wall-overlay` affiche l'écart `Figée vs appliqué` instantanément.

- **Configuration minimale nécessaire**  
  Il suffit de le lancer sur votre lien d'accès, choisir les profils `P1/P2` et laisser Meteolink exécuter la matrice `pfifo_fast / fq_codel / CAKE × CUBIC / BBR` et vous montrer la comparaison.

- **Suivi du temps de réponse applicatif**  
  Suivi du `small p95` — les petits objets critiques (télémétrie, alertes) qui souffrent le plus du bufferbloat. Extrêmement utile si vous voulez protéger le trafic qui compte.

- **Un seul binaire**  
  Meteolink est écrit en `Go`. Pour l'exécuter, seul le binaire est nécessaire — le tableau de bord `React` est embarqué via `go:embed`. Aucune base de données, aucune dépendance de service. Il embarque même son propre serveur `SSE`.

- **Presque tous les scénarios d'accès**  
  Meteolink accepte tout profil de lien (`P1` fibre 80 Mbit/s, `P2` 4G 20 Mbit/s, `P3` VSAT 5 Mbit/s importable via `POST /api/profile/import`). Les `qdisc` prédéfinis incluent `pfifo_fast`, `fq_codel`, `CAKE` et les `CC` `CUBIC`, `BBR`.

- **Traitement incrémental des campagnes**  
  Besoin de persistance ? Meteolink fige chaque évènement vers `data/runs/<run>/aqm_eval.csv` + `manifest.json` (SHA-256). `cgo verify` les vérifie, `cgo figures` régénère les `SVG` sans `Node`.

- **Provenance vérifiable**  
  Chaque résultat affiche `hash8 = sha256(dernier aqm_eval.csv)[:8]` — même hash dans `Wall`, `Résultats`, `Provenance` et `report.md`. Les évènements `invalid` mis en quarantaine restent comptés, jamais masqués.

- **Levier de façonnage du bord**  
  `Façonnage du bord` (`qdisc` + `capacity 1–1000 Mbit/s` + `delay/jitter/loss`) compose avec `Surveillance continue` (`POST /api/watch`) et `Burst` (`POST /api/burst` CUBIC/BBR) — l'écart live est le produit.

- **Support Docker**  
  Possibilité d'exécuter le tableau de bord dans un conteneur ; montez `data/runs` pour conserver les preuves gelées.

## Pourquoi Meteolink ?

Meteolink a été conçu pour être un auditeur de lien rapide, basé sur le terminal. Son idée centrale est d'auditer et comparer rapidement les politiques AQM/BBR en temps réel sans toucher à vos routeurs (*idéal si vous voulez analyser vite votre lien 4G via SSH, ou si vous aimez simplement travailler dans le terminal*).

Il sert aussi d'outil pratique pour le diagnostic terrain, facilitant la détection du bufferbloat, du partage inéquitable (`JFI`) et de la capacité gaspillée directement depuis votre lien. Bien que la sortie terminal (`meteolink top`) soit la sortie par défaut, il peut générer un tableau de bord [`HTML`](http://192.168.174.128:9090) temps réel complet et autonome, ainsi qu'un rapport [`CSV`](http://192.168.174.128:9090/api/report/export?format=csv) et [`Markdown`](http://192.168.174.128:9090/api/report/export?format=md).

Voyez-le plutôt comme une commande `monitor` pour votre lien d'accès.

## Installation

### Compilation depuis une release

Meteolink peut être compilé et utilisé sur les systèmes *nix. Téléchargez, extrayez et exécutez le binaire unique avec :

```
$ wget https://github.com/Realms4239/cgo/releases/download/v1.0.6/cgo-linux-amd64.tar.gz
$ tar -xzvf cgo-linux-amd64.tar.gz
$ ./cgo --serve              # http://127.0.0.1:9090
# ou meteolink --serve
```

Vérifiez avec `checksums.txt` (SHA-256).

### Compilation depuis GitHub (Développement)

```
$ git clone https://github.com/Realms4239/cgo.git
$ cd cgo
$ cd web/frontend && bun install && bun run build && cd ../..
$ go build -o bin/cgo ./cmd/cgo
$ ./bin/cgo --serve
```

### Distributions

Il est plus simple d'installer Meteolink via le gestionnaire de paquets préféré :

#### Go

```
$ go install github.com/Realms4239/cgo/cmd/cgo@latest
$ go install github.com/Realms4239/cgo/cmd/meteolink@latest
```

#### npm (wrapper)

```
$ npm i -g meteolink
$ meteolink --serve
```

#### Windows (observation)

```
> cgo.exe --serve   # 127.0.0.1:9090, Windows = observe (sans tc)
```

#### Docker

Une image `Docker` peut exécuter le tableau de bord ; montez les `runs` gelés pour conserver les preuves :

```
$ docker run -p 9090:9090 -v ./data/runs:/data/runs meteolink --serve --addr 0.0.0.0:9090
```

### Banc VM (le vrai banc `tc`)

Le banc est une VM `Ubuntu` (`VMware` ou `VirtualBox`). Convention : stocker les VMs sous `D:\VMs\` ou `C:\VMs\` (ex. `D:\VMs\ubuntu\ubuntu.vmx`) ; exception `D:\ubuntu.vmx` toujours trouvée. Le moteur scanne `C:`/`D:` en profondeur ≤3 (`--deep` pour complet) via `vmrun list` + `inventory.vmls` + `VBoxManage`.

```
$ cp kit/cgo-vm.yaml.example kit/cgo-vm.yaml  # renseigner ssh/vmx
$ bash kit/engine.sh --action scan            # trouve les .vmx/.vbox
$ bash kit/engine.sh --action ensure          # démarre si SSH coupé
$ bash kit/engine.sh --action deploy          # build → cross → push → health
```

`kit/cgo-vm.yaml` est gitignoré et portable — `host: auto` découvre l'IP invité via `vmrun getGuestIPAddress`, `vmx_path` auto-rempli par `scan` (`D:/VMs`/`C:/VMs`), surcharge possible via `CGO_SSH_HOST`/`CGO_DASHBOARD_PORT`. Ne jamais le committer.

**DNS local portable :** `bash kit/install.sh --hosts` (Admin) ajoute `127.0.0.1 meteolink.dev` (host) et `192.168.174.128 meteolink.vm` (VM) — `http://meteolink.dev:9090` et `http://meteolink.vm:9090`. `.dev` est `HSTS` (force `https`) : en local `http` reste OK via `hosts` + `mkcert meteolink.dev` si `https` requis, sinon préférer `http://localhost:9090` (secure context).

## Stockage

#### Archives gelées par défaut

Les anneaux `live` en mémoire offrent de meilleures performances (180 s, 1800 pts à 10 Hz). Pour la persistance, Meteolink fige chaque évènement vers `data/runs/<run_id>/aqm_eval.csv` + `manifest.json` (SHA-256). Ce stockage supporte aussi `cgo verify` et `cgo figures`.

#### CSV sur disque + Manifest

Chaque campagne fige la matrice ; `manifest.json` liste `file` + `sha256`. `GET /api/integrity` expose `hash8`.

## Ligne de commande / Options de configuration

Voir les [options](docs/api.md) passables à la commande ou dans `GET /api/schema`. Si spécifiées dans le fichier de configuration, les options longues doivent être utilisées sans `--`.

```
$ cgo --serve --addr 127.0.0.1:9090 --mode auto   # auto: Linux full, Windows observe
$ cgo audit --link-type 5g --site "Dept X" --duration 300
$ cgo verify
$ cgo figures
$ cgo doctor
$ cgo shape --restore
$ meteolink top --addr http://localhost:9090 --interval 250ms
```

## Utilisation / Exemples

**Note :** le tableau de bord écoute `127.0.0.1:9090` par défaut ; passez `--addr 0.0.0.0:9090` pour exposer en LAN (comme le fait `kit/vm-install.sh`).

### Démarrage

Pour afficher dans un terminal et générer un tableau de bord live :

```
$ cgo --serve
# ouvrir http://localhost:9090
```

Pour auditer votre lien depuis ce poste (non intrusif, sans admin) :

```
$ cgo audit --link-type 5g --site "Dept X" --duration 300
# → data/link_audit.csv (p50/p95, small p95, goodput)
```

Pour générer un rapport `CSV` sur la sortie standard :

```
$ curl "http://localhost:9090/api/report/export?format=csv"
```

Meteolink permet aussi une grande flexibilité de filtrage temps réel. Pour diagnostiquer vite le bufferbloat sur le mur live :

```
$ curl -X POST http://localhost:9090/api/shape -H 'Content-Type: application/json' -d '{"qdisc":"cake","capacity_mbps":20}'
# observez le live-wall-overlay Figée vs appliqué — -41% est le produit
```

### Multiples profils

Il existe plusieurs façons d'exécuter plusieurs profils avec Meteolink. La plus simple est de passer plusieurs profils à la campagne :

```
$ curl -X POST http://localhost:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P1","P2"],"reps":3}'
```

Il est même possible d'importer un profil personnalisé depuis l'UI (`Campagne → Profil personnalisé → P3`) ou via pipe :

```
$ echo '{"id":"P3","capacity_mbps":5,"delay_ms":600}' | curl -X POST http://localhost:9090/api/profile/import -H 'Content-Type: application/json' -d @-
```

### Tableau de bord temps réel

Meteolink peut afficher les données temps réel dans le tableau `HTML`. Vous pouvez même envoyer le dossier `data/runs` par email puisqu'il est composé de simples `CSV` sans dépendance externe.

Le processus de génération d'un tableau temps réel est très similaire à celui d'un rapport statique. Seul `--serve` est nécessaire.

```
$ cgo --serve --addr 0.0.0.0:9090
```

Pour voir le rapport, naviguez vers `http://<ip>:9090`. Par défaut, Meteolink écoute sur le port `9090`, pour utiliser un autre port :

```
$ cgo --serve --addr 0.0.0.0:9870
```

Et pour lier le serveur `WebSocket` à une autre adresse que `127.0.0.1` :

```
$ cgo --serve --addr 127.0.0.1:9090
```

### Filtrage

#### Travail avec les profils

Un autre filtre utile est de comparer un seul profil ou une seule `qdisc`. Sur `Résultats`, utilisez les puces `tous → P1` ou `cake`. Le tableau `filtered` et le `rank-verdict` se recalculent instantanément — tout depuis les `CSV` gelés.

#### Tests Burst

Pour comparer `CUBIC` vs `BBR` à travers le bord façonné sans campagne complète :

```
$ curl -X POST http://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"bbr","seconds":4}'
$ curl -X POST http://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"cubic","seconds":4}'
# observez goodput + RTT sur Tableau live
```

### Astuces

Il vaut aussi noter que si vous voulez exécuter Meteolink en basse priorité, vous pouvez le lancer comme :

```
$ nice -n 19 cgo --serve
```

et si vous ne voulez pas l'installer sur votre serveur, vous pouvez encore exécuter l'audit depuis votre machine locale :

```
$ ssh -n altfloat@192.168.174.128 'cgo audit --link-type 5g --site "Site X" --duration 30' | cat
```

### Dépannage

Nous recevons beaucoup de questions. Vérifiez d'abord :

- `cgo doctor` — `tc` présent, `CAP_NET_ADMIN`, `BBR`, `ping` — tout vert avant une campagne.
- `cgo shape --restore` — nettoie les `qdisc` périmés après un crash.
- `GET /api/health` → `{"mode":"full","version":"1.0.6"}` — `observe` sur `Windows` est normal, la campagne renvoie `501`.
- `go vet ./...` a besoin de `web/frontend/dist` — `bun run build` d'abord, sinon `embed.go` échoue.
- `ECharts` : ne jamais réintroduire `visualMap piecewise` ni `LinearGradient` area — cela plante `LineView` (`coord`) et fige les voisins. `ChartSurface` `init` dans `useEffect`, `dispose` au cleanup.

### Traitement incrémental des campagnes

Meteolink peut traiter les campagnes de façon incrémentale via son stockage gelé. Fonctionnement :

1. Une matrice doit d'abord être exécutée avec `POST /api/run/start`, puis le même jeu est gelé vers `data/runs/<run>`.
2. `GET /api/results` scanne les `CSV` gelés ; `GET /api/integrity` affiche `hash8`.

Pour lire seulement les données persistées (sans nouvelle campagne) :

```
$ cgo verify
$ cgo figures
```

## Contribuer

Toute aide sur Meteolink est la bienvenue. Le plus utile est de l'essayer et donner votre retour. N'hésitez pas à utiliser le suivi d'issues `GitHub` et les `pull requests` pour discuter et proposer des changements de code.

## À propos

Meteolink est un banc d'essai réseau autonome et visualiseur interactif pour auditer les liens d'accès contraints et comparer les politiques AQM/BBR — s'exécute dans un terminal sur les systèmes *nix ou directement dans votre navigateur.
