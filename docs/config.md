# Référence `CGO_*` — Meteolink (CGO)

*Dernière révision : 2026-11-14. Vaut pour le binaire 1.3.5, la branche
`main` et les scripts `kit/*.sh`.*

## Méthode d'énumération

Chaque ligne ci-dessous a été trouvée dans le code **aujourd'hui** par :

```sh
grep -rhoE 'CGO_[A-Z0-9_]+' --include='*.go' --include='*.sh' \
  --include='*.ps1' --include='*.mjs' . | sort | uniq -c | sort -rn
```

→ **48 variables**, auxquelles s'ajoute **`CGO_ALLOW_MAIN`**, lue dans
`.githooks/pre-push` (fichier sans extension, hors du glob ci-dessus —
documentée ici pour complétude). Soit **49 entrées**. Toute variable absente
de ce tableau n'est lue nulle part ; il n'existe pas de magic globale autre.

Les cinq variables **critiques** sont marquées **★** — celles dont dépendent
l'exposition, la cible mesurée, la racine de données ou la porte de push.
`CGO_ENABLED` est une variable de la chaîne d'outillage Go, listée pour
mémoire (pas un réglage d'application).

Forme : `Var` — ce qu'elle fait (fichier où elle est lue) — défaut — quand
la poser.

## ★ Variables critiques

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_API_TOKEN` ★ | Jeton d'authentification (Bearer CLI/API ; cookie `cgo_session` HMAC du jeton navigateur). Sans lui, aucun bind hors loopback (`--expose` refusé via exit 2), y compris avec TLS. Lu par `api/auth.go`, `cmd/cgo/expose.go`, `cmd/cgo/main.go`. | vide → dev loopback uniquement | **Toujours** pour un accès non loopback ; jamais dans une unité systemd (voir `/etc/cgo/auth.token`, 0600) |
| `CGO_IFACE` ★ | Nom de l'interface dont le débit descend à 10 Hz dans `/api/state` et `/api/stream`. Linux seulement (`/proc/net/dev`). Aucune auto-détection : une mauvaise interface serait une mesure mal étiquetée. Lu par `cmd/cgo/server.go`, `kit/run-dualns.sh`. | vide → échantillonneur inactif (champs `iface_*` absents) | Dès que vous voulez voir le débit réel du bord (Linux) |
| `CGO_DATA_DIR` ★ | Racine des données : `audits/`, `runs/`, `figures/`, `link_audit.csv`, `calibrations/`. Lu par `results/corpus.go`, `audit/artifacts.go`, `cmd/cgo/*`. | `data` (relatif au cwd ; `StateDirectory=cgo` en systemd → `/var/lib/cgo`) | Dès que les artefacts ne vivent pas dans `./data` (VM, systemd) |
| `CGO_ALLOW_PRIVATE_TARGETS` ★ | Opt-in banc : autorise les cibles RFC1918/link-local (10.30.0.2 du banc dualns) côté SSRF. **Sans lui, toute cible privée est refusée (403)** même avec un audit légitime. Lu par `api/ssrf.go`, `probe/ssrf.go`. | vide → refus (fail-closed) | **Uniquement** sur le banc dualns ou pour des tests sur réseau privé ; jamais comme interrupteur d'administration |
| `CGO_ALLOW_MAIN` ★ | Override de la porte `.githooks/pre-push` : autorise `git push origin main`. Lu par `.githooks/pre-push`. | vide → push main refusé | Quand — et seulement quand — le mainteneur décide de pousser `main` |

## Service et console

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_DASHBOARD__ADDR` | Adresse par défaut du `--serve` (**double underscore**, volontaire). Lu par `cmd/cgo/main.go`, `kit/run-dualns.sh`. | `meteolink.dev:9090` | Pour un autre nom/port d'écoute sans flag (rare — préférez `--addr`) |
| `CGO_BENCH` | Étiquette de banc pour la provenance (`legacy` = reproduction du corpus gelé, `dualns` = banc durci production). Lu par `cmd/cgo/runcli.go`, `cmd/cgo/server.go`, `cmd/cgo/main.go`, `api/server.go`. | vide → `legacy` | `dualns` sur le banc durci ; laisser vide pour reproduire le corpus |
| `CGO_CHANNEL_MODEL` | Canal : `clean` ou `bursty` (P2/P3 uniquement — refus sinon). Lu par `pkg/campagne/dualns.go`, `cmd/cgo/main.go`. | `clean` | Pour des courses en canal pertes en rafales (validation) |
| `CGO_API` | URL de base de l'API que le TUI appelle. Lu par `cmd/cgo/tui.go`. | `https://meteolink.dev:9090` | TUI hors localhost ou contre une autre instance (HTTPS requis, sinon `tui` refuse) |
| `CGO_VERSION` | Version auto-déclarée du binaire (repli du packaging). Lu par `internal/kit/package.go`. | `inconnue` | Ne pas poser : réservée à la chaîne de release |
| `CGO_ENABLED` | Variable de la chaîne Go (`CGO_ENABLED=0` en cross-compile statique). **Pas un réglage Meteolink.** Lu par `internal/kit/kit.go`, `internal/kit/package.go`. | défaut Go (= 1 sur Linux avec toolchain gcc) | Ne pas poser : posé par la chaîne de build (`GOOS=linux go build`) |
| `CGO_GUI_CONSOLE` | `=1` : garde la console Windows ouverte après une erreur du launcher (flash évité). Lu par `internal/kit/keysetup.go`. | vide | Lancement via double-clic sous Windows |
| `CGO_YES` | `=1` : réponses par défaut en non-interactif (kit, keysetup). Lu par `cmd/cgo/kit.go`, `internal/kit/keysetup.go`. | vide → interactif | Scripts, CI, `--yes` |

## Interface de mesure (échantillonneur 10 Hz)

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_IFACE_SOURCE` | Provenance déclarée de l'interface : `twin` \| `prod` \| `unknown` (toute valeur inconnue retombe sur `unknown`). Lu par `cmd/cgo/server.go`, `iface/iface.go`. | non déclarée → `unknown` | Dès que `CGO_IFACE` est posé (jumeau vs production) |

## Banc de mesures (legacy 10.200.0.0/24 et dualns 10.30.0.0/24)

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_TARGET` | Cible ping du banc. Lu par `pkg/campagne/prod.go`, `pkg/campagne/dualns.go`, `kit/stress_testbed.sh`. | legacy `10.200.0.1` · dualns `10.30.0.2` | Topologie de banc non standard |
| `CGO_SMALL_URL` | URL du petit objet (HTTP). Lu par `pkg/campagne/prod.go`, `pkg/campagne/dualns.go`, `api/server_audit.go`, `kit/stress_testbed.sh`. | legacy `http://10.200.0.1:8081/small` · dualns `http://10.30.0.2:8081/small` | Puits HTTP terrain ; externe → obligatoire sinon pas de charge mesurée |
| `CGO_BULK_ADDR` | Hôte:port du puits bulk (iperf-like). Lu par `audit/sinks.go`, `pkg/campagne/*`, `api/server_audit.go`, `kit/stress_testbed.sh`. | legacy `10.200.0.1:5201` · dualns `10.30.0.2:5201` | Puits terrain / externe — explicite ou pas de charge |
| `CGO_SHAPER_IF` | Interface dont `root` reçoit le façonnage (upload : émission du client ; download : émission du serveur). Lu par `pkg/campagne/prod.go`, `pkg/campagne/dualns.go`. | legacy = `CGO_CLI_IF` · dualns `e-srv` (up) / `e-wan` (down) | Topologie non standard |
| `CGO_CLI_IF` | Interface côté client (veth/w-edge). Lu par `pkg/campagne/prod.go`, `pkg/campagne/dualns.go`. | legacy `veth-c` · dualns `w-edge` | Topologie non standard |
| `CGO_SRV_IF` | Interface côté serveur (mode download legacy). Lu par `pkg/campagne/prod.go`. | `veth-s` | Download sur banc legacy non standard |
| `CGO_SRV_NS` | Namespace netns côté serveur. Lu par `pkg/campagne/prod.go`. | `cgo-srv` | Banc legacy refondu |
| `CGO_CLIENT_NS` | Namespace netns client (mode `ingress`). Lu par `pkg/campagne/dualns.go`. | `ns_client` | Façonnage descendant IFB |
| `CGO_EDGE_NS` | Namespace du routeur d'égouttement dualns. Lu par `pkg/campagne/dualns.go`. | `ns_edge_router` | Banc dualns non standard |
| `CGO_WAN_NS` | Namespace d'émulation WAN (netem). Lu par `pkg/campagne/dualns.go`. | `ns_wan_emulator` | Banc dualns non standard |
| `CGO_INGRESS_IF` | Interface IFB côté client (mode `ingress`). Lu par `pkg/campagne/dualns.go`. | `c-cli` | Façonnage descendant |
| `CGO_INGRESS_DEV` | Périphérique IFB. Lu par `pkg/campagne/dualns.go`. | `ifb0` | Façonnage descendant |
| `CGO_DOWNLINK_PHYS_IF` | Interface physique downstream (sondes physiques). Lu par `pkg/campagne/dualns.go`. | `c-wan` | Banc dualns instruments |
| `CGO_SHAPE_MODE` | Mode de façonnage : `egress` (défaut) ou `ingress` (IFB). Lu par `pkg/campagne/dualns.go`, `kit/cgo-bench-dualns.sh`. | `egress` | `ingress` pour le download sur IFB ; le banc lance alors l'IFB côté client |
| `CGO_OFFLOADS` | Recette GRO/GSO/TSO du banc `kit/testbed.sh` : `on` (recette gelée, par défaut) ou `off` (durcissement). Lu par `kit/testbed.sh`, `kit/stress_testbed.sh`. | `on` | `off` **explicitement** pour une passe de durcissement |

## Modèle de délai, canal et étalonnage

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_DELAY_MODEL` | `scalar` (délai fixe) ou `dist` (distribution netem via `.dist`). Lu par `pkg/campagne/dualns.go`, `pkg/campagne/prod.go`. | `scalar` | `dist` après `cgo calibrate` (trace → `.dist` scellée) |
| `CGO_CALIBRATION_DIR` | Dossier des `.dist` (lu via `TC_LIB_DIR`). Lu par `pkg/campagne/dualns.go`, `pkg/campagne/prod.go`. | `<DataRoot>/calibrations` (`filepath.Join(results.DataRoot(), "calibrations")`) | `dist` avec calibrations déplacées |
| `CGO_BASELINE_SEC` | Fenêtre baseline raccourcie (runs de stress). **0 = défauts de thèse (30 s).** Lu par `pkg/campagne/prod.go`, `pkg/campagne/dualns.go`, `kit/stress_testbed.sh`. | `0` | Stress/CI — jamais pour une campagne citée |
| `CGO_CHARGE_SEC` | Fenêtre de charge raccourcie. **0 = défauts de thèse (120 s).** Lu par les mêmes fichiers. | `0` | Stress/CI |
| `CGO_RECUP_SEC` | Fenêtre de récupération raccourcie. **0 = défauts de thèse (30 s).** Lu par les mêmes fichiers. | `0` | Stress/CI |

## Audits (démarrage, cibles, remédiation)

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_ALLOW_PRIVATE_TARGETS` ★ | Voir « Variables critiques ». | refus | Banc privé uniquement |
| `CGO_AUDIT_IFACE` | Interface par défaut de la remédiation côté API (audits lancés via `POST /api/audit/start`). Lu par `api/server_audit.go`. | vide → champ API requis | Machines où l'audit part toujours via la même iface |
| `CGO_AUDIT_ROUTEROS_IFACE` | Interface RouterOS par défaut (script de remédiation). Lu par `api/server_audit.go`. | vide → champ API requis | Audits API sur une même box RouterOS |
| `CGO_SMALL_URL` / `CGO_BULK_ADDR` | Voir « Banc ». En audit **externe**, ils priment sur la détection banc (`audit.EnvOrBulk`). Lu par `audit/sinks.go`, `api/server_audit.go`, `cmd/cgo/main.go`. | banc ; vide → pas de charge en externe | Terrain : toujours, sinon pas de charge mesurée |

## Kit, machine virtuelle et outillage shell

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_BIN` | Chemin du binaire `cgo` utilisé par les scripts de banc. Lu par `kit/run-dualns.sh`, `kit/stress_testbed.sh`, `kit/testbed.sh`. | `cgo` (PATH) ou `$HOME/cgo/cgo-linux` selon le script | Poste sans binaire sur le PATH |
| `CGO_SSH_HOST` | Hôte SSH de la VM banc ; `auto` = découverte. Lu par `internal/kit/kit.go`, `kit/config.sh`. | `auto` | VM à adresse fixe / DHCP ; forcer la cible |
| `CGO_SSH_PORT` | Port SSH de la VM. Lu par `internal/kit/kit.go`, `kit/config.sh`. | `22` | NAT/VirtualBox (voir `NatHostPort` 2222 côté kit) |
| `CGO_SSH_USER` | Utilisateur SSH de la VM. Lu par `internal/kit/kit.go`. | vide (demandé à la configuration) | Non-interactif |
| `CGO_SSH_KEY` | Clé privée SSH. Lu par `internal/kit/kit.go`. | `~/.ssh/id_ed25519` | Clé ailleurs |
| `CGO_VM_IP` | IP invitée en dernier ressort (aucune découverte possible). Lu par `internal/kit/kit.go`, `kit/config.sh`. | vide | Découverte en panne / sans hyperviseur détectable |
| `CGO_PROJECT_DIR` | Dossier du checkout **côté invité** (deploy kit : binaire + installateur). Lu par `internal/kit/kit.go`. | `$HOME` réel de l'invité, sinon `/home/<user>/cgo` | Layout distant non standard |
| `CGO_DASHBOARD_HOST` | Nom d'hôte du dashboard (kit : DNS, URLs). Lu par `internal/kit/kit.go`. | `meteolink.dev` | Déploiement hors `meteolink.dev` |
| `CGO_DASHBOARD_PORT` | Port du dashboard côté kit. Lu par `internal/kit/kit.go`, `kit/config.sh`. | `9090` | Pare-feu / multi-instances |
| `CGO_TESTBED_OBJ` | Chemin de l'objet binaire servi par `testbedsrv`. Lu par `cmd/cgo/testbedsrv.go`. | `testbed/obj16.bin` | Sink non standard |
| `CGO_TESTBED_LOG` | Journal du banc de stress. Lu par `kit/stress_testbed.sh`. | `$HOME/cgo/testbed/testbedsrv.log` | Diagnostic |
| `CGO_BBR_NO_RETRY` | `=1` : ne pas relancer le build noyau en cas d'échec (install-bbr3). Lu par `kit/install-bbr3.sh`. | vide → une relance auto | Build kernel supervisé |
| `CGO_WRITE_FIXTURE` | `=1` : écrire la fixture du rapport mesuré (test opt-in) — sinon `t.Skip`. Lu par `audit/report_fixture_test.go`. | vide → skip | Quand on veut inspecter la fixture à l'œil |

## Hors glob mais lu par les hooks

| Variable | Ce qu'elle fait | Défaut | Quand la poser |
|---|---|---|---|
| `CGO_ALLOW_MAIN` ★ | Voir « Variables critiques ». | refus | Push main approuvé par le mainteneur |

## Notes honnêtes

- **Aucune valeur ci-dessus n'a été inventée** : chaque défaut vient du
  fichier cité. Deux familles de défauts historiques diffèrent entre le banc
  **legacy** (`10.200.0.0/24`) et **dualns** (`10.30.0.0/24`) — elles sont
  notées.
- **`CGO_DASHBOARD__ADDR` porte deux underscores** : c'est voulu, ne pas
  « corriger » en `CGO_DASHBOARD_ADDR` (cette variable-là n'existe pas).
- **Sens de `--before`** dans `cgo audit purge` : seuls les dossiers
  strictement antérieurs sont supprimables, et la fenêtre `--keep` est un
  plancher absolu (voir `SECURITY.md`).
- Les fichiers `.ps1` et `.mjs` du dépôt ne lisent **aucune** variable
  `CGO_*` ; les hits du glob viennent de `*.go` et `*.sh`.
- Si une variable ne figure pas ici, elle n'est lue nulle part dans le code
  (ou elle relève du toolchain Go comme `CGO_ENABLED`).
