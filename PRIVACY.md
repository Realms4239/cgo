# Confidentialité — Meteolink (CGO)

*Dernière révision : 2026-11-14. Correspond au binaire 1.3.5.*

Meteolink mesure un lien réseau que l'opérateur possède. Ce document dit ce
qui est traité, ce qui ne l'est pas, et comment l'effacer. Rien de ce qui
suit n'est une promesse commerciale : c'est ce que fait le code, vérifiable.

## Données personnelles : l'essentiel

- **Le binaire n'acquiert pas et ne traite pas de données personnelles
  d'utilisateurs finaux.** Il mesure des caractéristiques de lien (RTT,
  latence de petits objets, débit, pertes, files `tc`). Aucun contenu de
  trafic n'est inspecté ; aucun utilisateur n'est identifié.
- **Les labels de site sont du texte libre choisi par l'opérateur.** Ils
  peuvent être commercialement sensibles (site, département, offre). Ils
  restent sur la machine qui exécute Meteolink, dans les artefacts d'audit
  (`data/audits/<id>/audit_report.{json,md}`, `telemetry.csv`). Ni le code
  ni la chaîne de build ne les exfiltrent.
- **Le cookie de session `cgo_session` n'est pas une donnée personnelle** :
  c'est le HMAC-SHA256 du jeton, rien de plus. Le jeton lui-même vit dans
  `/etc/cgo/auth.token` (`0600`) ou dans `CGO_API_TOKEN`.
- **Aucune télémétrie sortante.** Le binaire n'émet que vers les
  cibles/sinks configurés par l'opérateur (IP, URL, host:port des champs de
  mesure) et vers le tableau de bord qu'il sert lui-même. Il n'existe pas
  de service de ping, de balise d'usage ou de comptage côté éditeur.
- **Le corpus gelé ne contient pas de données personnelles.** 159 runs,
  528 lignes : configuration de campagne (qdiscs, contrôles de congestion,
  profils, répétitions, seaux de mesure) et agrégats — aucune identité,
  aucun contenu, aucune donnée d'utilisateur final.

## Rétention : entièrement gérée par l'opérateur

- Les sessions d'audit vivent sous `data/audits/<audit-id>/` et s'effacent
  avec **`cgo audit purge`** :
  `cgo audit purge [--keep N] [--before YYYY-MM-DD] [--dry-run] [DIR]`.
  - `--keep N` (défaut 10) : les N sessions scellées les plus récentes sont
    **toujours conservées**.
  - `--before YYYY-MM-DD` : seules les sessions scellées **strictement
    antérieures** à cette date sont supprimables.
  - Sans `--before`, **rien n'est supprimé** (`purge — rien à faire`).
  - `--dry-run` liste les suppressions proposées et n'écrit rien.
  - Seules les sessions scellées (`audit-*` + `audit_report.json`) sont
    supprimables ; le corpus gelé (`data/runs`, anchors/manifests) et les
    dossiers sans sceau sont toujours conservés.
- Suppression totale : arrêtez le service, puis supprimez le répertoire de
  données (`CGO_DATA_DIR`, défaut `./data`) — les artefacts scellés partent
  avec. Aucune copie ailleurs n'existe côté code.
- La `telemetry.csv` scellée et le `link_audit.csv` append-only sont
  l'évidence de mesure ; les purger revient à `cgo audit purge` (sessions)
  ou à une suppression manuelle du répertoire — l'opérateur décide.

## Packaging / revue DSI-DPIA

| Question | Réponse |
|---|---|
| Responsable du traitement | l'opérateur qui installe et exécute Meteolink |
| Sous-traitant | aucun — binaire local, pas de backend éditeur |
| Catégories de données | métadonnées de campagne + mesures agrégées de lien ; labels de site (texte libre opérateur) |
| Sujets de données | **N/A** — aucune personne n'est identifiée ni identifiable par ces données |
| Base légale / finalité | mesures d'ingénierie réseau par l'opérateur sur son propre lien ; pas de profilage, pas de suivi inter-sites |
| DPIA | non requise en l'état (pas de traitement de données personnelles) ; ce document tient lieu de notice pour la revue DSI |
| Transferts hors UE | aucun — aucun endpoint extérieur n'est contacté par défaut |
| Conservation | durée opérateur ; mécanisme : `cgo audit purge` (voir ci-dessus) |
| Sécurité | loopback + TLS par défaut, jeton `0600`, auth Bearer/cookie, en-têtes, limite de débit (voir `SECURITY.md`) |
| Minimisation | aucune donnée au-delà des mesures et du label opérateur ; labels à renseigner avec la granularité voulue (département, site), jamais avec des données tierces |

## Ce dont il faut être conscient

- **Un label de site peut révéler une information commerciale** (sites,
  en-cours de déploiement, noms de fournisseurs). Utilisez le niveau de
  détail dont vous avez réellement besoin.
- **`link_audit.csv` et les rapports** consignés montrent la capacité et les
  défauts du lien mesuré : ce sont des documents internes avant de devenir
  des preuves externes ; leur diffusion suit la politique de l'opérateur.
- **Le banc et les mesures diffèrent** : rien de ce qui est écrit ici ne
  couvre les journaux système du poste (`journalctl`, shell history) —
  hors périmètre du binaire.

## Contact

- Questions de confidentialité : le suivi d'issues du repo
  (<https://github.com/Realms4239/cgo/issues>), label `question`.
- Le canal opérateur/DGM passe par le contrat d'étude de cas (lecture seule,
  pseudonyme) : `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md`.
  Aucun contact DGM personnel n'est publié dans ce dépôt.
