# Runbook — sonde WIS2 (`cgo-wis2-probe`) — DESIGN DIFFÉRÉ

> **État : DEFERRED.** Ce document est un design seul — aucun code,
> aucun broker, aucun chargeur. Aucune publication WIS2 n'a lieu
> avant la fermeture des gates Q1–Q7 et la décision humaine de
> portée de dissémination (§7.11 du design).

Contexte : `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md`
§7 (WIS2 Probe, R1–R14). Rappel du design : CGO mesure si un lien
contraint transporte de petits objets dans les délais (modèle C :
16 KiB chargés vs objectif DGM 220 ms / 95 %, jamais présenté comme
un mandat OMM). WIS 2.0 est le *transport* (MQTT pub-sub sur
Internet public) ; CAP 1.2 est le *vocabulaire* des alertes
tous-dangers. CGO n'est pas un Nœud WIS2 ; la sonde lit un rapport
scellé et émet des notifications — elle ne re-note, ne re-mesure et
ne re-scelle rien.

## Arbre de sujets `cgo/v1/probe/...` (espace lab, par défaut)

Standard WIS2 (R1), jamais utilisé avant la désignation du centre-id :

```text
origin/a/wis2/<centre-id>/data/core
origin/a/wis2/<centre-id>/data/recommended
origin/a/wis2/<centre-id>/metadata
cache/a/wis2/<centre-id>/data/core
```

`origin/a/wis2/<centre-id>/data/core` = accès libre (résolution 1) ;
le Global Cache le télécharge et le ré-annonce sur
`cache/a/wis2/<centre-id>/data/core`. `.../data/recommended` peut
exiger un contrôle d'accès et n'est PAS ingéré par les Global Caches.
`.../metadata` porte le lien canonique WCMP2.

Espace lab PROPOSÉ (défaut jusqu'à l'existence du centre-id) :

```text
cgo/v1/probe/<site-id>/link-quality        # lab seul, jamais sur un Nœud WIS2
cgo/v1/probe/<site-id>/cap-sla              # lab seul
cgo/v1/probe/<site-id>/status/$SYS-like     # préférer le $SYS du broker [R3]
```

Rationale : `origin/a/wis2/...` exige un centre-id proposé par le
Membre et entériné par le Secrétariat OMM [R1]. Publier sous
`origin/a/wis2/<id-inventé>/...` avant l'affectation serait une
fausse revendication WIS2. `cgo/v1/...` garde l'outil lab honnête et
est ce que les ACL Mosquitto peuvent délimiter (`acl_file`,
`pattern` `%u`/`%c`) [R3].

Mapping WIS2 PROPOSÉ (seulement après affectation du centre-id +
évaluation GISC [R1]) :

| Sortie de la sonde | Sujet WIS2 | Classe |
|---|---|---|
| Résumé d'audit scellé (JSON dérivé, §7.5.2 du design) | `origin/a/wis2/<centre-id>/data/recommended/cgo/link-quality` | PROPOSÉ, sous-préfixe de `.../data/recommended` [R1] ; `recommended` car contrôle d'accès et pas d'ingestion Global Cache [R1] |
| Enregistrement WCMP2 du dataset ci-dessus | `origin/a/wis2/<centre-id>/metadata` | standard [R1] ; schéma `unverified` [R11] |
| XML CAP (§ ci-dessous) | URL HTTP référencée depuis le lien canonique WNM, PAS un payload MQTT | séparation PROPOSÉE ; CAP comme granularité WIS2 = `unverified` [R12] |

## Couche de notifications (JSON lab PROPOSÉ)

```json
{
  "schema": "cgo.probe.notification/v1",
  "site_id": "PROPOSÉ — affecté par l'opérateur, jamais un hostname",
  "audit_id": "<id du audit_report.json scellé>",
  "seal_sha256": "<SHA-256 de telemetry.csv — copié, jamais recalculé autrement>",
  "measured_at": "2026-11-07T00:00:00Z",
  "link_quality": {
    "cap_deadline_ms": 220,
    "cap_ok_pct": 96.5,
    "cap_breach": false,
    "basis": "loaded 16 KiB completions"
  },
  "download_url": "https://<probe-http-host>/<audit_id>/audit_report.json",
  "cap_url": "https://<probe-http-host>/<audit_id>/alert.xml"
}
```

`cap_deadline_ms` / `cap_breach` recopient la sémantique
`AssessCAPSLA` (`Measured=false` ⇒ aucune valeur, jamais un succès
par défaut) [R5]. 220/95 restent « objectif DGM, cadre WIS 2.0 » —
jamais mandat OMM [R5]. `seal_sha256` est une COPIE du sceau CGO [R4] ;
la sonde ne re-scelle jamais. Un rapport sans sceau est refusé
(fail-closed), miroir de « non mesuré ⇒ `Unavailable`, jamais zéro,
jamais succès » [R4].

## Mapping CAP 1.2 (PROPOSÉ, seulement si l'opérateur veut des alertes CAP)

Standard (R2) : XML CAP 1.2, espace `urn:oasis:names:tc:emergency:cap:1.2`
[§3.2.1]. Requis `<alert>` : `identifier`, `sender`, `sent`,
`status`, `msgType`, `scope` [§3.2.1]. Requis `<info>` : `category`,
`event`, `urgency`, `severity`, `certainty` [§3.2.2]. `status` ∈
{Actual, Exercise, System, Test, Draft} ; `msgType` ∈ {Alert, Update,
Cancel, Ack, Error} ; `scope` ∈ {Public, Restricted, Private}
[§3.2.1]. `sender` globalement unique (ex. basé sur le domaine)
[§3.2.1]. Producteur §4.3, consommateur §4.4 [R2].

| Élément CAP [R2] | Valeur PROPOSÉE | Source / règle |
|---|---|---|
| `<identifier>` | `<audit_id>-<préfixe-sha256>` (unique par message, sans espaces/virgules) | PROPOSÉ ; unicité [§3.2.1] |
| `<sender>` | id basé sur le domaine opérateur, ex. `cgo-probe/<site-id>.dgm.example` | PROPOSÉ ; unicité globale [§3.2.1] |
| `<sent>` | `measured_at` de l'audit en DateTime, UTC en `-00:00` | règle [§3.2.1] ; format [§3.2.2 `effective`] |
| `<status>` | `Actual` | PROPOSÉ ; domaine [§3.2.1] |
| `<msgType>` | `Alert` à la première brèche, `Update` + `<references>` au rétablissement, `Cancel` pour rétracter | séquence PROPOSÉE ; domaine + forme `<references>` `sender,identifier,sent` [§3.2.1] |
| `<scope>` | `Private` (adresses nommées opérateur) ou `Restricted` | PROPOSÉ ; domaine [§3.2.1] — jamais `Public` sans visa opérateur |
| `<info><category>` | `Infra` (infrastructure télécom) | PROPOSÉ depuis le domaine R2 [§3.2.2] |
| `<info><event>` | `Link quality below operator objective` | texte PROPOSÉ ; élément requis [§3.2.2] |
| `<info><urgency>/<severity>/<certainty>` | dérivés des bandes `cap_ok_pct` — **bandes PROPOSÉES et devant être signées par l'opérateur (Q5)** | domaines [§3.2.2] ; aucune table de bandes dans R2 — en inventer une en silence = jouet |
| `<info><parameter>` | `valueName="CGO_CAP_OK_PCT"`, `value="<cap_ok_pct>"` ; `valueName="CGO_SEAL_SHA256"`, `value="<sceau>"` | usage PROPOSÉ du `<parameter>` optionnel [§3.2.2] |
| `<resource>` | `audit_report.json` / `telemetry.csv` scellés via `<uri>` | PROPOSÉ ; `<resource>` = référence optionnelle d'actif numérique [§1.3.3] |

Frontière dure : une sonde saine (aucune brèche) ne publie AUCUN
message CAP. `msgType=Alert` porte « l'information initiale exigeant
l'attention » [§3.2.1] ; la télémétrie routinière n'est pas une
alerte. Le canal d'état stable est la notification JSON ci-dessus,
pas CAP.

## Profil Mosquitto TLS/ACL (lab, PROPOSÉ)

Capacités standard du design (R3) : listeners séparés par classe de
confiance (`listener <port> [bind]`) ; listener TLS
(`cafile`/`certfile`/`keyfile`) ; `require_certificate true` +
`use_identity_as_username true` (le CN du certificat devient
l'utilisateur ACL, mot de passe inutilisé) ou `password_file` +
`allow_anonymous false` ; `acl_file` (`topic [read|write|readwrite|deny]
<topic>`, `user <username>`, `pattern` `%c`/`%u`) ; `max_qos` ;
`retain_available` ; `persistence` + `persistence_location` ;
`sys_interval` (hiérarchie `$SYS`) ; ponts `connection` avec
`topic <pattern> [[in|out|both] qos] [préfixes]` + TLS de pont
(`bridge_cafile`/`bridge_certfile`/`bridge_keyfile`) et notification
`$SYS/broker/connection/<id>/state`.

```ini
# Fragment mosquitto.conf PROPOSÉ — lab seul, pas une config de Nœud WIS2
listener 8883
cafile /etc/cgo-wis2/ca.crt
certfile /etc/cgo-wis2/broker.crt
keyfile /etc/cgo-wis2/broker.key
require_certificate true
use_identity_as_username true
allow_anonymous false
max_qos 1
retain_available true
persistence true
persistence_location /var/lib/cgo-wis2/
sys_interval 10
acl_file /etc/cgo-wis2/acl
```

`max_qos 1` et la politique retain sont PROPOSÉS : R1 ne donne pas de
valeur QoS et R3 ne définit que le bouton `max_qos` [R3, R12]. QoS 1
proposé car QoS 0 est « fire-and-forget » et la notification doit
survivre au lien contraint — choix d'ingénierie, pas mandat OMM,
enregistré comme tel. Liaison à une interface/adresse spécifique
(`listener <port> <bind>` [R3]) ; un listener public sur Internet est
une décision SÉPARÉE et ultérieure, conditionnée à l'enregistrement
[R1], jamais le défaut lab.

## Gates Q1–Q7 (avant tout slice WIS2)

| # | Requirement | Pourquoi | Statut |
|---|---|---|---|
| Q1 | Récupérer OMM-N°1060 Vol. II App. D (hiérarchie de sujets) + App. E (WNM, WCMP2) | R1 les lie mais ne reproduit pas les schémas ; JSON/CAP du design non implémentables depuis R1 seul | `unverified` [R11] |
| Q2 | Récupérer le Guide OMM-N°1061 §2.6.1.1 (enregistrement/désinscription de Nœud) | R1 résume l'enregistrement ; étapes normatives/évaluation dans le Guide | `unverified` [R11] |
| Q3 | Établir si les documents CAP sont portés comme granularités WIS2, et à quel QoS/retain | La séparation CAP via le lien canonique HTTP (pas le payload MQTT) reste une proposition jusqu'à réponse | `unverified` [R12] |
| Q4 | Obtenir l'état centre-id opérateur ou une décision écrite de non-enregistrement | Détermine si le mapping WIS2 du design est jamais utilisé ou si l'espace lab est permanent | `unverified` [R13] |
| Q5 | Bandes de sévérité CAP signées par l'opérateur (`cap_ok_pct` → urgency/severity/certainty) | R2 définit les domaines mais aucun mapping depuis la métrique de lien [§3.2.2] ; des bandes non signées = seuils inventés | signature, pas un fetch |
| Q6 | Évaluer la doc `wis2box` (install, payloads acceptés, support CAP) | Décide sonde→`wis2box` vs sonde→broker propre ; R1 le nomme seulement [R1] | `unverified` [R14] |
| Q7 | Choisir une bibliothèque cliente MQTT avec source primaire (notes de release/docs API) | Aucune bibliothèque dans `go.mod` aujourd'hui [R9] ; le choix doit être citable | non commencé |

Règle de construction : aucun slice WIS2 ne démarre avant la
fermeture de ses Q-items. En particulier S4 (publication WIS2) est
bloqué sur Q1–Q4, comme l'exige le DEFERRED du hardening [R8]. S0
ferme Q1–Q4 ; S1 Mosquitto lab ; S2 éditeur lab ; S3 producteur CAP
§4.3 ; S4 publication WIS2 (conditionnel) ; S5 pont/multi-site ; S6
preuve + ledger.

## Décision humaine — portée de dissémination

Seconde moitié de la condition de déblocage du hardening [R8] : les
docs primaires récupérées (première moitié, R1–R3) ne suffisent pas.
Une décision HUMAINE écrite que la dissémination est dans la portée
est requise, plus la subvention opérateur (a)–(d) avant toute lecture
de configuration opérateur [R7]. Sans cette décision, l'espace lab
`cgo/v1/...` reste permanent et aucune publication
`origin/a/wis2/...` n'a lieu — le design l'interdit (test : la
config lab ne contient aucun sujet `origin/a/wis2`).

Limites d'usage : jamais présenter 220 ms comme un mandat OMM ;
jamais présenter CGO comme un Nœud WIS2 (un capteur qui mesure le
lien ne devient pas, par mesure, éligible à publier pour le réseau).
Mosquitto est une dépendance externe nouvelle (stop-and-ask) ; toute
bibliothèque cliente MQTT en Go est une dépendance module nouvelle
(stop-and-ask, candidate non choisie — Q7). La sonde ne touche jamais
le corpus gelé ni les schémas CSV ; les gates CGO existantes sont
inchangées.

## Références primaires

- `docs/superpowers/specs/2026-11-08-cgo-supplychain-gaps-design.md` §7
  (R1 WMO WIS 2.0, R2 CAP 1.2 OASIS, R3 `mosquitto.conf(5)`, Q1–Q7, S0–S6).
- `docs/superpowers/plans/2026-11-08-cgo-supplychain-gaps.md` Task 5.
