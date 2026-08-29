# Architecture — Meteolink (cgo)

> Binaire unique Go embarquant le dashboard React (`web/frontend` via `go:embed`).
> Squelette : `cmd/cgo` (CLI + serveur) et `pkg/` — dix paquets à responsabilité unique.

## Composants

| Paquet | Rôle |
|---|---|
| `pkg/api` | Surface HTTP + SSE : `server.go` (routes REST), `sse.go` (hub 10 Hz), `translate.go` (recommandation matérielle). stdlib uniquement. |
| `pkg/campagne` | Orchestration de la matrice : `Matrix`, `StartMatrixWithID`, `RunEvent` (baseline → charge → récupération), `Deps` (TC, shaper, sondes, `StatsFn`, `DeadlineMs`). |
| `pkg/qdisc` | Couche tc : `Reset` (suppression du qdisc root), `ApplyNetem`, `ApplyShaper`, lecture des stats (`SumBytes` feuille uniquement, `SumDrops`). |
| `pkg/probe` | Sondes concurrentes : ping, small, bulk. |
| `pkg/metrics` | `Percentile`, `Summarize`, `DeadlineOKPct`, `CostARPerH`. |
| `pkg/results` | `Scan` de `data/runs` → médianes par cellule, `Quarantined`, `Best`. |
| `pkg/figures` | Génération SVG (barres `small_p95.svg`, nuage `scatter.svg`) + `ProvenanceSHA` / `ProvenanceHash8`. |
| `pkg/audit` | Audit client d'un lien réel (idle/loaded), append `data/link_audit.csv`. |
| `pkg/profile` | Registre dynamique de profils importés (CSV/JSON). |
| `pkg/model` | Types domaine : profils P1/P2/P3, qdiscs, CC, portes G0–G7, événements, en-têtes CSV. |

## Flux SSE (`GET /api/stream`)

Le hub diffuse un snapshot de l'état de campagne à **10 Hz** (`time.Ticker`, démarré
par `hub.Serve`, arrêté par `Handler.CloseHub` — le `Close` du serveur HTTP seul ne
suffit pas). Points clés (implémentés dans `pkg/api/sse.go`) :

- **Compression delta** : les clés structurelles (`profile`, `qdisc`, `cc`,
  `repetition`, `event_id`, `phase`) sont omises tant qu'elles ne changent pas.
- **Reprise** : anneau de 2048 frames ; un client qui renvoie `Last-Event-ID`
  rejoue tout ce qui suit. Sans cet en-tête, replay limité aux **5 dernières
  frames** (~0,5 s de warm-up, pas de tempête de 2 s).
- **Client frais** : reçoit d'abord une frame **complète** (dernier snapshot
  intégral) pour resynchroniser les champs structurels absents des deltas.
- **Backpressure** : abonné lent → compteur `dropped` + événement nommé
  `backpressure` (jamais plus d'une notice empilée) ; la mesure n'est jamais
  bloquée par un client.
- **Plafond** : 64 abonnés (`maxSubs`) ; au-delà, `503 too many streams`.
- `retry: 2000` ; pas d'en-tête `Connection` manuel (illégal en HTTP/2 —
  aurait cassé le SSE derrière un edge type cloudflared).

## Couche tc (façonnage du bord)

Topologie veth : le client subit le **netem** (`veth-c`), le shaper + AQM
s'appliquent côté serveur. Sémantique **reset-then-apply** :

1. `Reset` supprime tout qdisc root sur l'interface (idempotent) ;
2. `ApplyNetem` pose `netem` en **root** (`handle 1:`) avec délai/gigue/perte
   (perte omise quand nulle) ;
3. `ApplyShaper` empile le qdisc choisi comme **enfant du netem 1:** (cake ou
   fq_codel ; tbf pour la cellule `pfifo_fast`) — délai et débit s'appliquent
   ainsi au même egress. Repli sur root si le netem 1: est absent.

Le levier manuel `POST /api/shape` est **refusé (409)** tant qu'une campagne est
active : matrice et façonnage manuel se disputent le même shaper. `qdisc: none`
efface le façonnage. `cgo shape --restore` supprime les qdiscs root (exécuté
aussi au démarrage pour repartir d'un bord propre).

## Portes G0–G7 (validité d'une cellule)

Chaque événement est marqué `valid` / `degraded` / `invalid` (`gate_status` de
`aqm_eval.csv`) :

| Porte | Condition |
|---|---|
| G0 | Cible atteignable (`TargetReachable`) |
| G1 | Bulk démarré (`BulkStarted`) |
| G2 | Sondes productives (`ProbesProducing`) |
| G3 | Latence plausible (`LatencyPlausible`) |
| G4 | Débit cohérent (`ThroughputCoherent`) |
| G5 | Pas de lignes dupliquées (`NoDuplicateRows`) |
| G6 | Baseline stable (`BaselineStable`) |
| G7 | CPU non saturé (`CPUNotSaturated`) |

## Gel, manifest et provenance

En fin de matrice, le writer **gèle** le run (`Freeze("config")`) : les CSV sous
`data/runs/<run_id>/` deviennent immuables et un **manifest SHA-256** est écrit.
`pkg/figures` calcule `ProvenanceSHA` (sha256 du dernier `aqm_eval.csv`) et
`ProvenanceHash8` (8 premiers hex), injectés dans les SVG, le rapport exporté et
`/api/integrity` : toute figure ou tableau porte l'empreinte exacte des données
qui l'ont produit.

## Modes d'exécution et diagnostics

- **Linux = mode complet** : mesure, façonnage tc, campagne.
- **Windows = mode observation** : les leviers actifs (shape/run/watch) sont
  refusés (501) ; l'observation reste possible. `--mode observe|full|auto`.
- Journal opérateur : anneau des **50 derniers événements** (`recordEvent`),
  exposé par `GET /api/events`.
- `GET /api/diagnostics` répond `{hub: "ok", time}` ; health check sur
  `GET /api/health`.

## Pourquoi Go

Un **binaire unique statique** (dashboard React embarqué, zéro runtime à
installer sur la passerelle), **cross-compilable** sans dépendances (stdlib
seule pour toute la surface HTTP/SSE — pas de framework web), et des
**goroutines** qui rendent naturelles les sondes concurrentes (ping/small/bulk
en parallèle d'un bulk pendant 120 s) et le fan-out SSE à 10 Hz vers 64
abonnés. Le compromis CGO est évité : tout le façonnage passe par `tc`
exécuté en sous-processus, ce qui garde le binaire cross-compilable.
