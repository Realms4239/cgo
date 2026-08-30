# Meteolink

**Banc d'audit et de comparaison AQM/BBR pour liens d'accès contraints — un seul binaire Go + dashboard React intégré, preuves gelées et vérifiables.**

[![version](https://img.shields.io/badge/version-1.0.6-blue)](VERSION)
[![license](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![go](https://img.shields.io/badge/go-1.25-%2300ADD8)](go.mod)
[![bun](https://img.shields.io/badge/bun-1.x-black)](web/frontend/bun.lock)
[![dashboard](https://img.shields.io/badge/dashboard-React%20%2B%20ECharts-5ad3e3)](#tableau-de-bord)

Meteolink audite un lien depuis le poste client, rejoue le lien sur un banc Linux reproductible (`tc`/`netem`), et compare les disciplines de file (`pfifo_fast` / `fq_codel` / `CAKE`) × contrôles de congestion (`CUBIC` / `BBR`) avec un protocole fixe (30 s baseline → 120 s charge → 30 s récupération, portes G0–G7) et des CSV gelés + manifeste SHA-256. Contexte d'origine : Météo Madagascar — petit trafic critique (télémétrie, alertes) vs. transferts massifs (imagerie satellite) sur liens 4G/5G/fibre sujets au bufferbloat.

## Pourquoi Meteolink ?

La plupart des discussions "qualité réseau" se font sans preuve. Meteolink produit la preuve :

- **Audit côté client, non intrusif.** `ping` p50/p95, `small p95`, `bulk goodput` → `data/link_audit.csv`.
- **Évaluation reproductible.** Matrice `P1/P2 × qdisc × CC × répétitions`, gel `data/runs/<run>/aqm_eval.csv` + `manifest.json`.
- **Provenance vérifiable.** `cgo verify` (SHA-256) + `cgo figures` (SVG depuis gelés) + `GET /api/integrity` `hash8`.
- **Pilotage opérateur.** Dashboard 4 vues en français sur SSE 10 Hz + levier `Façonnage du bord` + `Surveillance continue` + journal 50 évènements.
- **Un seul binaire.** `go:embed` embarque le frontend ; pas de DB, pas de service externe.

## Fonctionnalités

- Audit lien → `data/link_audit.csv`
- Matrice `P1/P2 × pfifo_fast/fq_codel/CAKE × CUBIC/BBR × reps`
- Protocole 30/120/30 + portes G0–G7 + quarantaine `invalid`
- Archives gelées + `GET /api/run/rows` + `GET /api/results` (médianes/IQR, `best ★`)
- Dashboard : `Campagne`, `Tableau live`, `Résultats` (classement complet, `A/B` pin), `Provenance`
- Façonnage du bord (`POST /api/shape` `qdisc/capacity/delay/jitter/loss` 1–1000 Mbit/s) + `Surveillance` (`POST /api/watch`) + `Burst` (`POST /api/burst` CUBIC/BBR) + `Journal` (`GET /api/events`)
- Modes : `Linux=full` (`tc`), `Windows=observe` (audit seul) ; forçage `--mode observe|full|auto`
- `GET /api/schema` source unique des bornes (10 params rendus côté Réglages)

## Non-objectifs

- Alertes / notifications
- Flotte multi-sites / agents distants
- Stockage historique long terme
- Auth multi-utilisateur (instrument VM-local, double confirmation `ArmButton`)
- Configuration de routeurs de prod

## Démarrage rapide

**Linux — full :**
```bash
./bin/cgo --serve              # http://127.0.0.1:9090
./bin/cgo audit --link-type 5g --site "Dept X" --duration 300
./bin/cgo run                  # via API : POST /api/run/start {profiles:["P2"],reps:3}
./bin/cgo verify
```

Le dashboard écoute `127.0.0.1:9090` par défaut ; `0.0.0.0:9090` pour exposer en LAN (cf. `kit/vm-install.sh` et `cgo service install`).

**VM Ubuntu — le vrai banc :**
```bash
cp kit/cgo-vm.yaml.example kit/cgo-vm.yaml   # renseigner ssh/vmx
bash kit/engine.sh --action deploy           # build → cross-compile → push → health
```

**Windows — observation :**
```powershell
cgo.exe --serve                # auto observe
cgo.exe audit --link-type 5g --site "Site X" --duration 300
```

## Installation

| Voie | Commande | Notes |
|---|---|---|
| **Release** | Télécharger `cgo-<os>-<arch>.tar.gz` depuis [Releases](https://github.com/Realms4239/cgo/releases) | 5 plateformes, `ldflags -X main.version`, `checksums.txt` |
| **Go** | `go install github.com/Realms4239/cgo/cmd/cgo@latest` / `.../cmd/meteolink@latest` | Go ≥1.25 |
| **npm** | `npm i -g meteolink` | wrapper `npm/` télécharge le binaire de plateforme à `postinstall` |
| **Source** | `make build` ou `bun run build && go build` | voir § Construction |

## Construction depuis les sources

**Prérequis :** Go 1.25, Bun ≥1.x (ou Node ≥18), `git`. Linux full : `iproute2`, `curl`, `bc`, noyau avec `tcp_bbr`.

```bash
git clone https://github.com/Realms4239/cgo.git && cd cgo

# 1) Frontend (TypeScript + Vite, <600KB gz)
cd web/frontend
bun install
bun x tsc --noEmit
bun run build          # → web/frontend/dist (vérifié par scripts/check-bundle.mjs)
bun x vitest run       # 59/59
cd ../..

# 2) Binaire (embed du dist)
go vet ./...
go test ./... -short
go build -o bin/cgo ./cmd/cgo
go build -o bin/meteolink ./cmd/meteolink

# ou tout-en-un :
make build             # frontend dist + go vet/test + binaires
make build-frontend    # seulement le frontend
```

Cross-compilation VM (depuis Windows) :
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kit/cgo-linux.new ./cmd/cgo
```

## Convention VM

- **Emplacement standard :** `D:\VMs\` ou `C:\VMs\` (ex. `D:\VMs\ubuntu\ubuntu.vmx`, `C:\VMs\debian\debian.vbox`). Le moteur scanne `C:` et `D:` en profondeur ≤3 (`--deep` pour balayage complet), `vmrun list` + `inventory.vmls` + `VirtualBox VMs`.
- **Exception historique :** `D:\ubuntu.vmx` (banc actuel) reste supportée ; le scan la trouve et `kit/engine.sh --action scan` met à jour `vmx_path`/`vbox_path` dans `kit/cgo-vm.yaml`.
- **Config machine-locale :** `kit/cgo-vm.yaml` (gitignoré, copier depuis `kit/cgo-vm.yaml.example` / `deploy/cgo-vm.yaml.example`) :

```yaml
ssh: {user: altfloat, host: 192.168.174.128, port: 22, key: ~/.ssh/id_ed25519}
vm_name: "ubuntu"
vmx_path: "D:/VMs/ubuntu/ubuntu.vmx"  # ou C:/VMs/...
project_dir: /home/altfloat/cgo
dashboard_port: 9090
```

Ne jamais committer `kit/cgo-vm.yaml` (contient chemins locaux).

## Configuration et déploiement

**Linux complet :**
```bash
bash kit/install.sh              # iproute2/curl/bc, modprobe tcp_bbr, test CAP_NET_ADMIN (ne change jamais le CC par défaut)
cgo doctor                       # tc, CAP_NET_ADMIN/root, BBR, ping — tout vert avant de continuer
sudo setcap cap_net_admin+ep ./cgo   # ou sudo ./cgo --serve
cgo service install              # /etc/systemd/system/meteolink.service (0.0.0.0:9090)
cgo shape --restore              # nettoyage qdiscs (aussi au démarrage serveur)
```

**VM (hôte → VM, idempotent, binaire-first) :**
```bash
bash kit/engine.sh --action scan       # localise le .vmx/.vbox
bash kit/engine.sh --action ensure     # boot si SSH down
bash kit/engine.sh --action deploy     # build+cross+push+install+health (http://192.168.174.128:9090)
bash kit/engine.sh --action status     # SSH + process + /api/health
bash kit/engine.sh --action logs       # tail /tmp/cgo.log
bash kit/engine.sh --action bootstrap  # apt + veth-c/veth-s (une fois)
```

Prérequis VM une fois : `sudo apt update && sudo apt install -y iproute2 curl bc && sudo modprobe tcp_bbr`.

**Edge mini-PC (Debian) :** `kit/install.sh` → `setcap` → `cgo service install` → `ufw allow 9090/tcp` → `curl http://<ip>:9090/api/health` → `{"ok":true,"version":"1.0.6","mode":"full"}` (`docs/deploy.md:4`).

## Utilisation

**Audit :**
```bash
cgo audit --link-type 5g --site "Dept X" --provider Yas --duration 300 --target 8.8.8.8
# → data/link_audit.csv (Tableau 6)
```

**Campagne :** UI `Campagne → DÉMARRER → CONFIRMER ?` ou API :
```bash
curl -X POST http://192.168.174.128:9090/api/run/start -H 'Content-Type: application/json' -d '{"profiles":["P2"],"reps":3}'
curl -X POST http://192.168.174.128:9090/api/run/stop
```

Chaque évènement : baseline 30 s → charge 120 s → récup 30 s, `G0–G7` ; arrêt gracieux, reprise par `reps`.

**Live :** `GET /api/state` + SSE `GET /api/stream` 10 Hz (`live.rtt95`, `live.small`, `phase`, `gates`). `Surveillance` `POST /api/watch {on:true}` compose avec `Façonnage`, `409` si campagne en cours.

**Façonnage / Burst :**
```bash
curl -X POST http://192.168.174.128:9090/api/shape -H 'Content-Type: application/json' -d '{"qdisc":"cake","capacity_mbps":50,"delay_ms":20,"jitter_ms":2,"loss_pct":0}'
curl -X POST http://192.168.174.128:9090/api/burst -H 'Content-Type: application/json' -d '{"cc":"bbr","seconds":4}'
```

**Résultats / Provenance :**
```bash
curl http://192.168.174.128:9090/api/results           # groupes médianes/IQR, best
curl http://192.168.174.128:9090/api/integrity         # runs, manifests, valid/quarantined, hash8
curl "http://192.168.174.128:9090/api/report/export?format=csv"  # ou md
curl -X POST http://192.168.174.128:9090/api/figures/regen
./cgo verify && ./cgo figures   # hors ligne
```

## Référence CLI

| Commande | Description |
|---|---|
| `cgo --serve [--addr host:port] [--mode observe\|full\|auto]` | Dashboard + API (défaut `127.0.0.1:9090`, VM `0.0.0.0:9090`) |
| `cgo audit --link-type --site --duration` | Audit lien → `data/link_audit.csv` |
| `cgo run` | Via API `POST /api/run/start` (CLI renvoie vers l'API) |
| `cgo verify` | SHA-256 des CSV gelés vs `manifest.json` |
| `cgo figures` | SVG `data/figures` depuis gelés |
| `cgo doctor` | Capacités `tc`, `CAP_NET_ADMIN`, `BBR`, `ping` |
| `cgo shape --restore` | Supprime qdiscs root |
| `cgo service install` | Unité systemd `meteolink.service` |
| `cgo testbedsrv` | Service banc `10.200.0.1:8081/5201` |
| `cgo version` | `meteolink (cgo) 1.0.6` |
| `meteolink top [--addr]` | TUI 8 cartes, sparklines `live` |

## Tableau de bord

4 vues françaises : `Campagne` (cockpit 380 fixe, P3 sticky, `① Auditer → ② Comparer → ③ Exporter`), `Tableau live` (bento 5 + duo RTT/goodput, `live-wall-overlay` `Figée vs appliqué`), `Résultats` (classement complet `small p95`/`RTT`/`goodput`/`coût`, filtres, `A/B` pin → `Comparaison`), `Provenance` (RDF `hash8`). `Rail` replié par défaut (56px, `▶` pour épingler), `Explain` survol + clic-pin (`lib/explain.ts`).

## API

`GET /api/health` `{ok,version,mode}`, `GET /api/schema` (10 params), `GET /api/state`, `GET /api/stream` (SSE), `GET /api/results`, `GET /api/run/rows?run=ID`, `GET /api/integrity`, `GET /api/events` (anneau 50), `POST /api/run/start|stop`, `POST /api/shape`, `POST /api/watch`, `POST /api/burst`, `POST /api/audit/start`, `GET /api/audit/status`, `POST /api/profile/import`, `GET /api/profiles` (`docs/api.md`).

## Structure

```
cmd/cgo        # binaire principal (serve + audit + verify + figures + doctor)
cmd/meteolink  # TUI + alias serve
pkg/api        # REST + SSE + hub 64, CloseHub
pkg/campagne   # Matrix, instrument, watch, writer, gates G0–G7
pkg/qdisc      # tc/qdisc, stats feuille, netem
pkg/results    # Scan + hardwareRecommendation (MikroTik v7)
web/frontend   # React + Vite + ECharts (init useEffect, dispose), <600KB gz
kit/           # engine.sh (VMware→VBox fallback), vm-install.sh, install.sh
deploy/        # moteur historique (donor thesis)
npm/           # wrapper meteolink (postinstall télécharge release)
data/runs/*/aqm_eval.csv + manifest.json  # gelés, trame modèle types.go
```

## Développement

```bash
go vet ./... && go test ./... -short
# web/frontend/dist requis pour embed.go ; sinon `go vet` échoue → `bun run build` d'abord
bun x tsc --noEmit
bunx vitest run          # 59/59
bun run build && node scripts/check-bundle.mjs  # TOTAL <600KB gz
npx playwright test --workers=1  # 10 sondes (chart-probe, compare-probe, control-demo …)
bash kit/test-kit.sh && bash kit/test-hypervisor.sh
```

ECharts : ne jamais réintroduire `visualMap piecewise` ni `LinearGradient` area (crash `LineView coord`). `ChartSurface` `init` dans `useEffect`, `dispose` au cleanup, base `baseOption` primée avant données.

## Reproductibilité

`data/runs` gelés + `manifest.json` SHA-256 (`cgo verify`), `cgo figures` régénère les SVG sans Node, `GET /api/report/export` inclut `provenance sha256:hash8`.

## Documentation

- [docs/usage.fr.md](docs/usage.fr.md) — manuel opérateur
- [docs/architecture.md](docs/architecture.md) — internes & décisions
- [docs/api.md](docs/api.md) — REST + SSE
- [docs/deploy.md](docs/deploy.md) — installation machine propre & VM
- [docs/methodology.fr.md](docs/methodology.fr.md) — protocole & portes
- [LIEN.md](LIEN.md) — mémoire source, [DESIGN.md](DESIGN.md), [PRODUCT.md](PRODUCT.md)

## Licence

MIT — contributions bienvenues. `VERSION` `1.0.6`, `.goreleaser.yaml` v2 (binaires `cgo`+`meteolink`, 5 plateformes, `checksums.txt`).
