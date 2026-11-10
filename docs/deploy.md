# Déploiement — installation machine propre (Meteolink 1.3.5)

Ce document couvre l'installation sur une machine vierge : Windows en
**mode observation**, Linux en **mode complet**, la VM de banc et la
recette mini-PC edge.

Les deux modes :

- **Windows = mode observation** : audit du lien + consultation
  (résultats, tableau live, provenance, journal). Campagne, façonnage
  et surveillance renvoient HTTP 501 avec un message clair.
- **Linux = mode complet** : tout est disponible (campagne, façonnage
  du bord, surveillance continue).

Le mode se force via `--mode observe|full|auto` (défaut `auto`).
Vérification : `GET /api/health` → `{ok, version, mode}`.

### 1.1 Matrice des capacités (Windows vs Linux)

| Capacité | Windows (observation) | Linux (complet) |
|---|---|---|
| Audit du lien (RTT, petits objets, charge) | ✅ | ✅ |
| Campagne / matrice AQM × CC | 501 | ✅ |
| Façonnage du bord (`tc`/netem) | 501 | ✅ (CAP_NET_ADMIN) |
| Surveillance continue | 501 | ✅ |
| `TCP_INFO` (RTT noyau, retransmissions) | n/a — hors Linux | ✅ |
| `netns` / banc `dualns` | n/a — hors Linux | ✅ (`kit/run-dualns.sh`) |
| Porte offloads du banc (GRO/GSO/TSO) | n/a — `warn`, pas un échec | ✅ (échec franc si actif) |

**n/a ≠ échec** : `cgo doctor` et `/api/doctor` rendent
`n/a — hors Linux (mode observation)` pour ces portes. Un `warn` Windows
ne doit pas être lu comme une régression ; la porte offloads échoue
franchement sur Linux quand le banc est monté et que GRO/GSO/TSO restent
actifs (commit `4968745`).

---

## 0. La racine du corpus — `CGO_DATA_DIR`

Depuis 2026-11-10, la racine des données n'est plus un littéral cwd-relatif (`data/`). Défaut inchangé pour les postes de travail (relatif) ; **tout déploiement doit l'épingler** :

- `systemd` (`cgo service install`) : crée `/var/lib/cgo` via `StateDirectory=cgo`, pose `WorkingDirectory=/var/lib/cgo` et `Environment=CGO_DATA_DIR=/var/lib/cgo` dans l'unité.
- `Docker` : l'image pose `WORKDIR /data`, `VOLUME /data`, `ENV CGO_DATA_DIR=/data` ; `docker-compose.yml` monte `cgo-data:/data`.
- La variable vaut pour tout (runs gelés, figures, `link_audit.csv`, artefacts d'audit, profils, calibrations).

Le contexte du changement et la revérification sont dans le registre des checkpoints (2026-11-10).

## 1. Windows — mode observation

Deux voies, au choix :

```powershell
# Binaire précompilé depuis GitHub Releases
#    Télécharger l'archive de plateforme, décompresser, lancer cgo.exe
```

Aucune dépendance système : le mode observation n'a pas besoin de `tc`
ni de privilèges réseau. L'audit du lien (client-side) fonctionne tel
quel :

```powershell
cgo audit --link-type 5g --site "Site X" --duration 300
```

## 2. Linux — mode complet

### 2.1 Installation des prérequis

```bash
bash kit/install.sh
```

Le script installe `iproute2`, `curl`, `bc`, charge le module
`tcp_bbr`, vérifie la présence de `tc` et signale l'absence de
CAP_NET_ADMIN. Il **ne change jamais** l'algorithme de contrôle de
congestion par défaut du système.

### 2.2 Diagnostic

```bash
cgo doctor
```

Vérifie : mode OS, `tc` dans le PATH, CAP_NET_ADMIN/root, BBR
disponible, `ping`. Ne passez pas à la suite tant qu'un point reste rouge.

### 2.3 Privilèges de façonnage

Le façonnage (`tc`/netem) exige CAP_NET_ADMIN. Deux options :

```bash
# Option a : capability sur le binaire (recommandé, pas de root au runtime)
sudo setcap cap_net_admin+ep ./cgo

# Option b : lancer le serveur en root
sudo ./cgo --serve
```

### 2.4 Service systemd

```bash
cgo service install
```

Installe l'unité `/etc/systemd/system/meteolink.service` puis active le
démarrage automatique (`systemctl enable --now`). Le service démarre en
mode `auto` : Linux complet.

### 2.5 Nettoyage du façonnage

`cgo shape --restore` garantit le **nettoyage des qdiscs** appliqués
par le façonnage (netem, cake, fq_codel…). Il est aussi exécuté
automatiquement **au démarrage du serveur** : même après un arrêt
brutal, le bord ne reste jamais façonné à l'insu de l'opérateur.

## 3. VM de banc (développement / mesure contrôlée)

Le déploiement hôte → VM est piloté par `cgo kit` (moteur Go, mêmes codes de
sortie que l'ancien bash ; `kit/engine.sh --action X` reste un shim de compat) :

```bash
cgo kit scan      # trouver le .vmx/.vbox, sauvegarder l'unique
cgo kit ensure    # démarrer la VM si SSH est down, attendre SSH
cgo kit deploy    # push + install + health-check (recompile si sources dispo, sinon binaire précompilé)
cgo kit status    # SSH + process + santé du dashboard
cgo kit logs      # tail du log serveur sur la VM
cgo kit align     # NIC vmxnet3 + CPU/mémoire mini du banc (à froid)
```

Le déploiement est **binaire-first et idempotent** : la SPA est
embarquée dans le binaire, aucune toolchain Go n'est requise sur la VM.

Prérequis VM (une fois) :

```bash
sudo apt update && sudo apt install -y iproute2 curl bc
sudo modprobe tcp_bbr || true
```

## 4. Recette mini-PC edge (Debian)

Pour un banc fixe sur site :

1. **Debian** vierge, accès sudo.
2. Installer les prérequis : `bash kit/install.sh`.
3. Installer le binaire (archive de la Release GitHub : `cgo-linux-amd64.tar.gz`).
4. Diagnostic : `cgo doctor` — tout vert.
5. Privilèges : `sudo setcap cap_net_admin+ep ./cgo`.
6. Service : `cgo service install`.
7. **Réseau** : ouvrir le **port 9090** (dashboard + API) —
   `sudo ufw allow 9090/tcp` ou équivalent.
8. Vérifier : `curl -k https://<ip-edge>:9090/api/health` →
   `{"ok":true,"version":"1.3.5","mode":"full"}` (-k : certificat auto-signé).

Le mini-PC est alors prêt pour la boucle opérateur complète
(auditer → campagner → façonner → surveiller → comparer → constat),
voir `docs/usage.fr.md`.

## 5. Checklist de mise en service

```text
[ ] /api/health renvoie {ok:true, version, mode attendu}
[ ] cgo doctor passe (Linux : tc, CAP_NET_ADMIN, bbr, ping)
[ ] port 9090 ouvert (accès distant)
[ ] cgo shape --restore exécuté une fois (bord propre au départ)
[ ] service systemd actif : systemctl status meteolink
```

## 6. Ligne de sécurité — dashboard mono-opérateur

`kit/run-dualns.sh serve` lie le dashboard à `10.99.0.2:9090` dans le
netns `ns_client`, en HTTP brut. Ce n'est **pas** un service multi-tenant
durci :

- **pas d'authentification ni d'autorisation** : quiconque atteint le port
  (ou le loopback de l'hôte) peut lancer/arrêter une campagne, appliquer ou
  retirer un façonnage, changer le palier tarifaire ;
- **pas de jeton CSRF ni de rate limit** ; les mutations refusent seulement
  `Sec-Fetch-Site: cross-site` (garde navigateur, contournable par un client
  direct) ; le corps JSON est borné à 1 Mio, ce n'est pas un contrôle
  d'accès ;
- **le façonnage modifie réellement le bord** : banc/labo ou mini-PC DSI,
  jamais exposé sur Internet ; n'ouvrir le 9090 que sur le réseau
  d'administration.

Ce qui est protégé : l'intégrité des mesures (gel SHA-256) et la validation
des entrées. Ce qui ne l'est pas : l'accès au levier de façonnage.

## S1 — Nginx boundary
All 9090 traffic enters through `deploy/nginx/cgo.conf` (TLS termination + `limit_req`). Direct `--tls=false` binds are loopback-only in CI. `POST` mutations that exceed `1r/s burst=3` receive `503`.

## S4 — OCI + rollback
`docker compose up` starts the digest-pinned image. Rollback is `compose pull` of the prior digest + `cgo shape --restore`. `cpu.max`/`memory.max` are set in compose; throttling stats are exported via `/metrics`.

---

## 7. Chaîne d'approvisionnement — SBOM + signatures + SLSA L2

Un binaire sans provenance est un jouet. Ce § est le contrat
d'expédition : tout artefact publié est nommé, énumérable et vérifiable
hors ligne, ou il ne part pas.

### 7.1 Ce que produit `.goreleaser.yaml`

| Artefact | Contenu | Référence primaire |
|---|---|---|
| `cgo-<os>-<arch>.tar.gz` / `.zip` | Binaires `cgo` + `meteolink` (le wrapper re-exécute `cgo`) | `semver.org` §2 (déclarer l'API, nommer la version) |
| `cgo-<os>-<arch>.tar.gz.sbom.json` | SBOM CycloneDX par archive, graphe hierarchique complet (`syft … --output cyclonedx-json`, `--enrich all`) | CycloneDX « Inventory … hierarchical relationships » |
| `checksums.txt` + `checksums.txt.sigstore.json` | Somme SHA-256 de chaque archive, signée `cosign sign-blob` keyless (workflow hébergé) | SLSA v1.0 L2 (authentic, signé par le builder hébergé) |
| `digests.txt` | Digests OCI publiés par le builder, attestés `actions/attest` (SLSA L2) | SLSA v1.0 « Providable, digest-identified » |
| SBOM par module pur (`cgo-<module>.sbom.json`) | Produit en CI depuis `modules/<module>/go.mod` par le même `syft`, même format, même release | R2 GoReleaser (limite : l'image OCI n'est pas catalogable par le pipe SBOM) |

Les cinq modules purs sont `model`, `metrics`, `results`, `figures`,
`qdisc`. Ce sont des bibliothèques sans binaire. `goreleaser` ne peut
pas les compiler en archives ; leur SBOM est produit par le CI depuis
leur `go.mod` et rattaché à la même release. L'image OCI, pareillement,
se documente hors du pipe SBOM (`syft <image>@<digest>` + attestation).

### 7.2 Épingles + cutover par module — aucun `replace` ne survit à un tag

Le monorepo vit en `require` + `replace` local (`go.mod` racine,
`modules/*/go.mod`). Le `replace` ne lie que le module principal : un
tag publié en gardant un `replace` embarque le chemin local au lieu du
module publié — le tag ne vaut rien.

Procédure, **dans l'ordre, module par module** (commencer par `model`,
puis `metrics`, `qdisc`, `results`, `figures` — ordre topologique) :

1. Module vert isolé :
   ```bash
   go test ./...   # depuis model/, puis metrics/, etc. (modules migrés
   # hors de modules/ au fur et à mesure des cutovers, voir tag-cutover.md)
   ```
   Les 15 modules doivent passer ensemble ; `go test ./...` depuis la
   racine doit rester vert après chaque cutover.
2. Tag signé, semver `v0.1.0` en nom `<module>/v0.1.0` (ex. `model/v0.1.0`,
   `metrics/v0.1.0` — un tag nu `v0.1.0` est impossible par module dans un
   même dépôt ; le `0` signale une API publique non garantie, `semver.org`
   §4 ; `1.0.0` est la promesse) :
   ```bash
   git tag -s model/v0.1.0 -m "cgo-model v0.1.0"   # répéter : metrics/v0.1.0, qdisc/v0.1.0, results/v0.1.0, figures/v0.1.0
   ```
   Un tag signé est immuable : son contenu ne change jamais (semver §3).
   On ne déplace pas un tag publié ; on publie `<module>/v0.1.1`.
3. Retirer le `replace` de CE module dans `go.mod` racine (et dans les
   `go.mod` des modules consommateurs), puis :
   ```bash
   go mod download github.com/Realms4239/cgo/model
   go test ./...
   python tools/verify.py
   ```
   Le consommateur doit résoudre le module publié, pas le chemin local.
   Tant qu'un `replace` survit, le cutover n'est pas terminé.
4. Vérifier avant de passer au suivant : `git grep "=>" go.mod
   modules/*/go.mod` ne doit plus nommer le module coupé, et la suite
   racine reste verte.

La version du monolithe (`VERSION`, `1.3.5`) et les versions des modules
(`v0.1.0` via `require`) sont deux lignes indépendantes : les modules ne
suivent pas `VERSION`, et inversement.

### 7.3 SBOM + SLSA en CI (L2)

Le niveau d'assurance est déclaré, jamais sous-entendu (SLSA) :

- **L1** (provenance existe) : `provenance.json` publié avec chaque
  artefact — digest cryptographique, commit source, builder. C'est le
  plancher.
- **L2** (authentic) : provenance signée par un builder hébergé.
  `.goreleaser.yaml` signe `checksums.txt` avec `cosign sign-blob`
  (keyless OIDC), `digests.txt` est attesté par `actions/attest`. C'est
  la cible de chaque release `<module>/v0.1.0` (ex. `model/v0.1.0`).

Vérification par le consommateur, hors ligne sauf la consultation du log
de transparence pour cosign keyless :

```bash
# 1. Vérifier la signature du fichier de checksums (valider l'identité
#    attendue du workflow + l'émetteur OIDC de GitHub Actions)
cosign verify-blob \
  --certificate-identity "https://github.com/Realms4239/cgo/.github/workflows/release.yml@refs/tags/v0.1.0" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle dist/checksums.txt.sigstore.json dist/checksums.txt

# 2. Vérifier chaque archive contre ce fichier signé
sha256sum --check --ignore-missing dist/checksums.txt

# 3. Vérifier la provenance SLSA L2
gh attestation verify --owner Realms4239 dist/cgo-linux-amd64.tar.gz
gh attestation verify --owner Realms4239 ghcr.io/realms4239/cgo@sha256:<digest>

# 4. SBOM CycloneDX : graphe hiérarchique, pas une liste de fichiers
#    (ex. avec l'outil de consultation de votre choix, local, sans réseau)
syft dist/cgo-linux-amd64.tar.gz --output cyclonedx-json=module.sbom.json
```

### 7.3.1 SBOM par module pur — `config-complete` (S2)

Les cinq modules purs (`model`, `metrics`, `results`, `figures`, `qdisc`)
n'ont pas de binaire : GoReleaser ne catalogue que les artefacts qu'il
produit (les archives, §7.1). Leurs SBOM sont donc catalogués par `syft`
depuis `modules/<nom>/go.mod`, avec le même format CycloneDX et rattachés
à la même release :

```bash
for m in model metrics results figures qdisc; do
  if [ -d "modules/$m" ]; then d="modules/$m"; else d="$m"; fi
  syft "$d" --output "cyclonedx-json=cgo-$m.sbom.json" --enrich all
done
```

Choix I1 : il n'y a PAS d'entrée `sboms:` par module dans `.goreleaser.yaml`.
Le pipe `sboms:` ne catalogue que les artefacts GoReleaser (les archives) ;
une entrée `artifacts: archive` avec un répertoire fixe `modules/<nom>`
tournerait une fois par archive et écraserait le même document
(`cgo-model.sbom.json` réécrit par chaque archive). La boucle CI ci-dessus
est donc la voie canonique des SBOM par module pur : même `syft`, même format
CycloneDX, même release, sans `goreleaser check` à faire passer sur ce pipe.
`goreleaser` est absent de cet hôte — vérifié par `command -v goreleaser` —
ce qui rend la validation du pipe chez CI sans objet pour ces modules.

Vérification, hors ligne :

```bash
python -c 'import json,sys; d=json.load(open("cgo-model.sbom.json")); assert d["bomFormat"]=="CycloneDX", d.get("bomFormat"); print("CycloneDX OK")'
sha256sum --check --ignore-missing dist/checksums.txt
```

Note S2 : le pipe `sboms:` catalogue les archives ; l'image OCI se
documente hors du pipe (`syft <image>@<digest>`, §7.4 — limite R2
vérifiée : les images ne sont pas catalogables par le pipe SBOM).

### 7.4 Image OCI signée, épinglée par digest — jamais de tag mutable

L'image se construit depuis `deploy/Dockerfile` (`FROM scratch`,
un binaire, zéro surface) avec les étiquettes
`org.opencontainers.image.*` (`title`, `version`, `revision`, `source`,
`licenses`) — `docker inspect` retrouve donc la provenance sans réseau.
Elle se publie par digest :

```bash
docker build --build-arg VERSION=1.3.5 --build-arg REVISION=$(git rev-parse HEAD) \
  -f deploy/Dockerfile -t ghcr.io/realms4239/cgo:<semver> .
cosign sign ghcr.io/realms4239/cgo@sha256:<digest>
# SBOM de l'image, hors du pipe GoReleaser (limite R2) :
syft ghcr.io/realms4239/cgo@sha256:<digest> --output cyclonedx-json=cgo.sbom.json
```

Le runner ne prend qu'un digest (`deploy/docker-compose.yml`) :

```bash
export CGO_IMAGE_DIGEST=<sha256-hex-du-manifest>
export CGO_API_TOKEN=<jeton>   # paramètre, jamais une valeur committée
docker compose -f deploy/docker-compose.yml config   # échoue franc sans digest/jeton
docker compose -f deploy/docker-compose.yml up -d
cosign verify ghcr.io/realms4239/cgo@sha256:${CGO_IMAGE_DIGEST}
```

État de vérification compose : `config-complete` — Docker absent de
l'hôte de build (`docker: command not found`), `docker compose config`
n'a pas pu tourner ici ; `deploy/docker-compose.yml` utilise déjà
`ghcr.io/realms4239/cgo@sha256:${CGO_IMAGE_DIGEST:?set release digest}`
et `deploy/Dockerfile` est `FROM scratch` avec les étiquettes
`org.opencontainers.image.*`. Valider en CI où Docker existe.

Mode image : par défaut le conteneur ne façonne pas le bord (pas de
`NET_ADMIN`). Pour le mode complet Linux, ajouter localement
`cap_add: [NET_ADMIN]` dans le compose et utiliser `cgo shape --restore`
au rollback : le bord ne reste jamais façonné à l'insu de l'opérateur.

Bundle hors ligne (mini-PC, banc sans egress) : charger le tar image
sans registre, SBOM et provenance en fichiers dans le bundle — un SBOM
derrière une URL est inutile hors ligne (CycloneDX, machine-readable).

### 7.5 `cmd/cgo-gui` — interne, hors chaîne d'expédition

`cmd/cgo-gui` est le client Windows natif (tags `//go:build windows` +
`golang.org/x/sys/windows`). Il reste **interne au monorepo** : il n'est
pas construit par `.goreleaser.yaml`, pas archivé, pas dans l'image OCI.
C'est volontaire : le cross-build Linux de ce paquetage Fyne casse déjà
(voir `docs/VERIFY.md`, « pré-existing `cmd/cgo-gui` cross-build
failure »). L'inclure casserait chaque release Linux.

Si un jour il doit être distribué, il sort du monorepo vers son propre
dépôt (chemin retenu : `github.com/Realms4239/cgo-gui`, `go 1.25`,
fenêtre Windows seule) avec sa propre `.goreleaser.yaml`, son propre
SBOM et sa propre signature — pas un artefact du pipeline `cgo`.

### 7.6 Runbook de bascule par module

La liste de contrôle complète (les 8 étapes atomiques : tag →
`require` → `tidy` → `list` → dry-run → lock → verify-fast →
vérifieur, plus la procédure d'exécution à sec et le rollback) est
`docs/runbooks/tag-cutover.md`. Le dry-run du 2026-11-08 y est consigné :
le retrait d'un `replace` sans publication préalable casse la
résolution (`unknown revision`) — publier d'abord, retirer ensuite.
