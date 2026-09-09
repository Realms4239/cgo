# Meteolink [![version](https://img.shields.io/badge/version-1.2.11-blue)](VERSION) [![license](https://img.shields.io/badge/license-MIT-green)](LICENSE) [![go](https://img.shields.io/badge/go-1.25-%2300ADD8)](go.mod)

## Qu'est-ce que c'est ?

Meteolink est un banc d'essai réseau open source, temps réel et visualiseur interactif qui s'exécute dans un terminal sur les systèmes *nix ou directement dans votre navigateur. Conçu pour les liens d'accès contraints (4G/5G, fibre, VSAT), il fournit à la volée des preuves AQM/BBR rapides et vérifiables. Meteolink audite votre lien côté client, le rejoue sur un banc Linux reproductible et présente les données directement dans le terminal ou via un tableau de bord HTML live — aucune promesse fournisseur, seulement des CSV gelés avec SHA-256.

Plus d'infos sur : [https://github.com/Realms4239/cgo](https://github.com/Realms4239/cgo).

## Fonctionnalités

Meteolink rejoue des profils de lien et affiche les données dans le terminal ou le tableau de bord. Fonctionnalités :

- **Entièrement temps réel**  
  Tous les panneaux et métriques sont rafraîchis toutes les 100 ms sur le flux `SSE` (10 Hz) et toutes les 250 ms sur la `TUI`. Le `live-wall-overlay` affiche l'écart `Figée vs appliqué` instantanément.

- **Configuration minimale nécessaire**  
  Il suffit de le lancer sur votre lien d'accès, choisir les profils (`P1`–`P4`) et laisser Meteolink exécuter la matrice `pfifo_fast / fq_codel / CAKE × CUBIC / BBR` et vous montrer la comparaison.

- **Suivi du temps de réponse applicatif**  
  Suivi du `small p95` — les petits objets critiques (télémétrie, alertes) qui souffrent le plus du bufferbloat. Extrêmement utile si vous voulez protéger le trafic qui compte.

- **Un seul binaire**  
  Meteolink est écrit en `Go`. Pour l'exécuter, seul le binaire est nécessaire — le tableau de bord `React` est embarqué via `go:embed`. Aucune base de données, aucune dépendance de service. Il embarque même son propre serveur `SSE`.

- **Presque tous les scénarios d'accès**  
  Meteolink accepte tout profil de lien (`P1` fibre 80 Mbit/s, `P2` 4G 20 Mbit/s, `P3` VSAT 5 Mbit/s, `P4` Starlink 100 Mbit/s natifs ; profils asymétriques / perte en rafales importables via `POST /api/profile/import`). Les `qdisc` prédéfinis incluent `pfifo_fast`, `fq_codel`, `CAKE` et les `CC` `CUBIC`, `BBR` — en upload comme en download (`--direction down|both`, RRUL séquentiel).

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

Il sert aussi d'outil pratique pour le diagnostic terrain, facilitant la détection du bufferbloat, du partage inéquitable (`JFI`) et de la capacité gaspillée directement depuis votre lien. Bien que la sortie terminal (`meteolink top`) soit la sortie par défaut, il peut générer un tableau de bord [`HTML`](https://localhost:9090) temps réel complet et autonome, ainsi qu'un rapport [`CSV`](https://localhost:9090/api/report/export?format=csv) et [`Markdown`](https://localhost:9090/api/report/export?format=md).

Voyez-le plutôt comme une commande `monitor` pour votre lien d'accès.

## Installation

### La voie unique : `cgo setup` (wizard guidé)

Clonez puis laissez le wizard détecter, construire et configurer — de zéro au dashboard en une commande :

```
$ git clone https://github.com/Realms4239/cgo.git && cd cgo
$ go run ./cmd/cgo setup
```

Sept étapes idempotentes : détection (OS, go/bun/node/ssh, hyperviseurs) → dépendances → build frontend+binaire → contexte (observe/full/VM) → config VM (scan, IP invité auto) → DNS local (`meteolink.dev`) → doctor + menu final (dashboard/TUI). Modes : `--yes` (CI, défauts), `--no-vm` (sans hyperviseur).

### Sans source : les releases (Windows + Ubuntu, zéro toolchain)

Chaque [release](https://github.com/Realms4239/cgo/releases) livre des archives prêtes : `cgo-gui` (centre de contrôle graphique, même 27 boutons sur les deux OS), `cgo kit tui` (équivalent terminal), scripts hôte/invité. Le pipeline est toujours le même — `cgo kit next` l'affiche palier par palier avec la prochaine commande exacte. Détail complet dans le `LISEZ-MOI` de l'archive (aussi : `cgo kit readme`).

### Poste opérateur depuis une release (Ubuntu + VM du banc)

Téléchargez l'archive du poste (`cgo` linux statique + config exemple + installateur VM + mode d'emploi), puis lancez le centre de contrôle — aucune commande à taper :

```
$ wget https://github.com/Realms4239/cgo/releases/download/v1.2.3/cgo-linux-amd64.tar.gz
$ wget https://github.com/Realms4239/cgo/releases/download/v1.2.3/cgo-linux-amd64.tar.gz.sha256
$ sha256sum -c cgo-linux-amd64.tar.gz.sha256
$ tar xzf cgo-linux-amd64.tar.gz -C ~/cgo-op && cd ~/cgo-op
$ ./cgo kit tui    # flèches + entrée : dépendances → scan VM → clé SSH → deploy → dashboard
```

```
$ wget https://github.com/Realms4239/cgo/releases/download/v1.2.3/cgo-linux-amd64.tar.gz
$ wget https://github.com/Realms4239/cgo/releases/download/v1.2.3/cgo-linux-amd64.tar.gz.sha256
$ sha256sum -c cgo-linux-amd64.tar.gz.sha256
$ tar xzf cgo-linux-amd64.tar.gz -C ~/cgo-op && cd ~/cgo-op
$ ./cgo kit doctor     # dépendances (openssh-client installé auto si absent)
$ ./cgo kit scan       # trouve la VM sur tout le PC (vmrun/VBoxManage)
$ ./cgo kit keysetup   # pose la clé SSH (mot de passe tapé dans ssh, jamais stocké)
$ ./cgo kit ensure     # boot VM + SSH actif
$ sudo ./cgo kit dns   # meteolink.dev → VM
$ ./cgo kit deploy     # pousse CE binaire testé, sert le dashboard (sans recompiler)
$ sudo ./cgo kit tls   # confiance HTTPS totale (magasin système)
# → https://meteolink.dev:9090
```

Pré-requis côté VM invitée : `open-vm-tools` (découverte d'IP) + `openssh-server` — `kit ensure` dit la commande console exacte s'il manque. Serveur du dashboard : `./cgo kit svc start|stop|restart|status`.

### Compilation depuis GitHub (développement)

```
$ git clone https://github.com/Realms4239/cgo.git && cd cgo
$ go run ./cmd/cgo setup --yes     # ou manuellement :
$ cd web/frontend && bun install && bun run build && cd ../..
$ go build -o bin/cgo ./cmd/cgo && ./bin/cgo --serve
```

### Distributions

#### Go

```
$ go install github.com/Realms4239/cgo/cmd/cgo@latest
```

#### Windows (observation + pilotage kit)

```
> cgo.exe --serve   # https://meteolink.dev:9090, Windows = observe (sans tc)
```

Voir « Poste opérateur depuis une release » ci-dessus pour le pilotage VM complet depuis Windows (compagnon `cgo-linux` embarqué dans le zip).

### Banc VM (le vrai banc `tc`) — `cgo kit`

Le banc est une VM `Ubuntu` (`VMware` ou `VirtualBox`). Convention : VMs sous `D:\VMs\` ou `C:\VMs\` (exception racine historique tolérée). Le moteur `cgo kit` (pur Go, multi-OS) remplace l'ancien bash :

```
$ cgo kit scan        # trouve les .vmx/.vbox sur TOUT le PC, sauvegarde l'unique (--shallow = conventions seules)
$ cgo kit keysetup    # pose la clé SSH via mot de passe (prompts user/hôte/port + confirmation OUI, zéro GUI)
$ cgo kit ensure      # SSH up, sinon boot + attente
$ cgo kit align       # NIC vmxnet3 + CPU/mémoire mini (à froid)
$ cgo kit deploy      # build → cross → push → health
$ cgo kit svc start|stop|restart|status  # pilote le dashboard distant (sans sudo)
$ cgo kit doctor      # dépendances + config, tout vert avant d'agir
$ cgo kit status · logs · health · bootstrap · build · tunnel
$ cgo kit snapshot · snapshots · revert   # garde-fou : liste + retour arrière
$ cgo kit verify       # empreintes SHA-256 des archives, recalculées sur la VM
$ cgo kit ssh · ps · backup
```

21 actions au total. Mêmes codes de sortie que l'ancien `engine.sh` (2 usage/build, 3 scan ambigu, 4 hyperviseur absent, 5 timeout SSH, 6 cross, 7 scp, 8 install), env `CGO_SSH_HOST`/`CGO_DASHBOARD_PORT`/`CGO_VM_IP` inchangés (l'env **gagne** sur `cgo-vm.yaml` — forcer une IP après un bail DHCP glissant). `kit/engine.sh` reste en shim de compatibilité. **Machine propre sans SSH :** `cgo kit doctor` classe l'échec (sshd absent / clé refusée / machine éteinte) et affiche la remédiation, `cgo kit keysetup` pose la clé sans console (mot de passe tapé dans ssh), `cgo kit ensure` gère boot + découverte d'IP.

**DNS local + HTTPS (obligatoire, HSTS préchargé) :** `sudo cgo kit dns` mappe `meteolink.dev` vers la VM ; le dashboard n'écoute QU'en HTTPS (`https://meteolink.dev:9090`, certificat auto-signé généré au premier lancement) — `sudo cgo kit tls` l'installe dans le magasin système pour un accès sans avertissement. Sans confiance installée : accepter une fois dans le navigateur (HSTS interdit le contournement au clic pour un hôte inconnu — installez le certificat).

## Stockage

#### Archives gelées par défaut

Les anneaux `live` en mémoire offrent de meilleures performances (180 s, 1800 pts à 10 Hz). Pour la persistance, Meteolink fige chaque évènement vers `data/runs/<run_id>/aqm_eval.csv` + `manifest.json` (SHA-256). Ce stockage supporte aussi `cgo verify` et `cgo figures`.

#### CSV sur disque + Manifest

Chaque campagne fige la matrice ; `manifest.json` liste `file` + `sha256`. `GET /api/integrity` expose `hash8`.

## Ligne de commande / Options de configuration

Voir les [options](docs/api.md) passables à la commande ou dans `GET /api/schema`. Si spécifiées dans le fichier de configuration, les options longues doivent être utilisées sans `--`.

```
$ cgo setup                                      # wizard : de zéro au dashboard
$ cgo kit doctor|scan|keysetup|ensure|align|deploy|svc|...    # moteur de déploiement (21 actions)
$ cgo run --profiles P2 --reps 3 --deadline 1000 # campagne CLI réelle
$ cgo tui                                        # terminal 5 onglets
$ cgo --serve --addr meteolink.dev:9090 --mode auto  # auto: Linux full, Windows observe
$ cgo audit --link-type 5g --site "Dept X" --duration 300
$ cgo verify
$ cgo figures
$ cgo doctor
$ cgo shape --restore
```

## Utilisation / Exemples

**Note :** le tableau de bord écoute `https://meteolink.dev:9090` (HTTPS uniquement — `.dev` est HSTS préchargé, aucun HTTP). `cgo kit dns` mappe le nom, `cgo kit tls` installe la confiance.

### Démarrage

Trois peaux, un seul moteur — le web dashboard, le TUI terminal, la CLI :

```
$ cgo --serve          # dashboard web — https://localhost:9090
$ cgo tui              # terminal : 5 onglets (Setup Kit Campagne Live Résultats)
$ cgo run --profiles P2 --reps 3   # campagne CLI : progression, CTRL-C gel, résumé
```

Le TUI pilote tout depuis le terminal : navigation `1-5`/`h-l`, actions par `ENTRÉE` (démarrer/arrêter campagne, exécuter kit), polling live 1 s. Sans API : état vide honnête, jamais de données synthétiques.

Pour auditer votre lien depuis ce poste (non intrusif, sans admin) :

```
$ cgo audit --link-type 5g --site "Dept X" --duration 300
# → data/link_audit.csv (p50/p95, small p95, goodput, note bufferbloat A+..F)
```

### Coût du gaspillage — paliers tarifaires réels

Le coût en Ariary suit le forfait réel de l'institution (`GET /api/cost/tiers`, recherche 2026, docs/data-prices.md) :

```
$ curl -k https://localhost:9090/api/cost/tiers
# yas-day-1gb 1 000 Ar/Go · yas-month-4.5gb 5 556 · yas-month-100gb 2 000 ·
# yas-ftth-100gb 490 · orange-month-5gb 2 000 · airtel-month-4.5gb 5 556
```

Défaut : mobile mensuel 4,5 Go (contexte cellular DSI). La fibre est ~10× moins chère au Go — le palier change le verdict économique.

Pour générer un rapport `CSV` sur la sortie standard :

```
$ curl -k "https://localhost:9090/api/report/export?format=csv"
```

Pour rejouer la meilleure cellule gelée de chaque profil en script `tc` (recette `aqm-recipe.sh`, même règle que la suggestion CLI) :

```
$ curl -k "https://localhost:9090/api/report/export?format=sh" -o aqm-recipe.sh
```

Meteolink permet aussi une grande flexibilité de filtrage temps réel. Pour diagnostiquer vite le bufferbloat sur le mur live :

```
$ curl -k -X POST https://localhost:9090/api/shape -H 'Content-Type: application/json' -d '{"qdisc":"cake","capacity_mbps":20}'
# observez le live-wall-overlay Figée vs appliqué — -41% est le produit
```

### Multiples profils

Il existe plusieurs façons d'exécuter plusieurs profils avec Meteolink. La plus simple est de passer plusieurs profils à la campagne :

```
$ curl -k -X POST https://localhost:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P1","P2"],"reps":3}'
```

Pour rejouer une seule cellule sans la matrice pleine (sous-matrice), ou mesurer le sens download (badge `↓` dans Résultats, jamais comparé à l'upload) :

```
$ cgo run --profiles P2 --qdiscs cake --cc bbr --reps 1 --direction down
$ cgo run --profiles P2 --reps 1 --direction both   # RRUL séquentiel : up puis down
```

Il est même possible d'importer un profil personnalisé depuis l'UI (`Campagne → Profil personnalisé → P3`) ou via pipe :

```
$ echo '{"id":"P3","capacity_mbps":5,"delay_ms":600}' | curl -k -X POST https://localhost:9090/api/profile/import -H 'Content-Type: application/json' -d @-
```

### Tableau de bord temps réel

Le dashboard vit à `https://meteolink.dev:9090` (HTTPS uniquement, HSTS — `cgo kit dns` + `cgo kit tls` sur le poste opérateur). Les campagnes gèlent chaque cellule dans `data/runs/<run_id>/aqm_eval.csv` (+ `manifest.json` SHA-256) ; `Résultats` lit ces CSV gelés, `Provenance` les vérifie.

```
$ cgo --serve --addr meteolink.dev:9090   # poste local (TLS, certificat auto-signé)
```

Pour exposer en LAN (debug seul, jamais en soutenance) : `--tls=false --addr 0.0.0.0:9090`.

### Filtrage

#### Travail avec les profils

Un autre filtre utile est de comparer un seul profil ou une seule `qdisc`. Sur `Résultats`, utilisez les puces `tous → P1` ou `cake`. Le tableau `filtered` et le `rank-verdict` se recalculent instantanément — tout depuis les `CSV` gelés.

#### Tests Burst

Pour comparer `CUBIC` vs `BBR` à travers le bord façonné sans campagne complète :

```
$ curl -k -X POST https://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"bbr","seconds":4}'
$ curl -k -X POST https://localhost:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"cubic","seconds":4}'
# observez goodput + RTT sur Tableau live
```

### Astuces

Il vaut aussi noter que si vous voulez exécuter Meteolink en basse priorité, vous pouvez le lancer comme :

```
$ nice -n 19 cgo --serve
```

et si vous ne voulez pas l'installer sur votre serveur, vous pouvez encore exécuter l'audit depuis votre machine locale :

```
$ ssh -n <user>@<ip-vm> 'cgo audit --link-type 5g --site "Site X" --duration 30' | cat
```

### Dépannage

Nous recevons beaucoup de questions. Vérifiez d'abord :

- `cgo doctor` — `tc` présent, `CAP_NET_ADMIN`, `BBR`, `ping` — tout vert avant une campagne.
- `cgo shape --restore` — nettoie les `qdisc` périmés après un crash.
- `GET /api/health` → `{"mode":"full","version":"1.2.3"}` — `observe` sur `Windows` est normal, la campagne renvoie `501`.
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
