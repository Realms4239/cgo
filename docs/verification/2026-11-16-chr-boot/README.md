# Vérification CHR RouterOS 7 — 2026-11-16 (hôte sadcnwp, côte banc)

**Résultat scellé (attempt 14, 2026-11-16T~10:12Z) — `proof-2026-11-16m.json` :
import.accepted = true · 23/23 vérifications · niveau linux-namespace · accélérateur KVM ·
`image_sha256 == c6abf876…` (octet-en-octet identique à l’image public-depinnée avant
démarrage) · `export_sha256 == 61e51573…` (export.txt = preuve octets, 982 o, hash vérifié) ·
`transcript_sha256 == 88b492d1…` (transcript 5 671 o, conservé sur l’hôte).**

Hôte de banc : sadcnwp (Ubuntu 24.04.5), à côté de la VM de banc (`labo-cgo-bench`,
`127.0.0.1:10022`, pid 3673912 — intouchée tout au long de la fenêtre). Le CHR tourne
sous KVM (`-accel kvm`, 256 Mo / 1 vCPU), image brute isolée du banc. Aucune VM de
production n'a été modifiée.

## Ce que prouve le scellé (avec des limites nommées)

- MikroTik CHR 7.24.6 démarre sous KVM et accepte le script RouterOS généré par le dépôt
  (`remediate.RouterOS` / `remediate.RouterOSIngress`, CAKE, 19 Mbit/s façonnés à 95 % d'un
  goulot mesuré « 20 Mbit/s / 25 ms » **de fixture**, interfaces `ether1`/`ether2`).
- La création de `cgo-wan-out`, `cgoingress-lan-in`, `cgo-cake`, `cgoingress-cake` est
  vérifiable dans `/export terse` après import.
- Une absence de propriété attendue n'est acceptée que si elle vaut le défaut usine
  RouterOS pour ce champ (`cake-flowmode=triple-isolate` — un tiers de la route `/export`
  est omise par RouterOS ; voir `rosDefaults` dans `audit/routeros_compare.go`).

## Ce que ce scellé ne prouve PAS

- **Pas de mesure réelle en jeu.** Le rapport consommé (`chr-fixture-cake-2026-11-16-r2.json`,
  `fixture: true`) ne vient d'aucun audit réel — c'est le même format de fixture que le
  rejeu hors-ligne du 2026-11-12 (`docs/verification/2026-11-12-chr-replay/`). Le rapport
  scellé du bench (`audit-1794798576`) vit dans la VM de banc et n’est pas atteignable par
  la chaîne sécurisée — c’est pourquoi le fixture a été utilisé.
- `level=linux-namespace` désigne ici `KVM` — l’attachement aux namespaces du banc n’est
  PAS implémenté dans `cmd/cgo-chr` (BuildQEMUArgs n’attache aucun netns ; attacher une
  saisie d’interface CHR à un namespace de banc est un jalon ultérieur).
- `transcript_sha256` (88b492d1…) référence le transcript complet conservé sur l’hôte
  (`/home/sadc/labo-cgo/chr/proof-2026-11-16m/transcript.txt`) — 5 671 octets contenant
  des séquences ANSI. **Transfert troué** sur cette chaîne : les captures réseau locales
  ne sont pas octet-exactes pour un fichier ANSI ; la preuve engagée est le sceau lui-même
  plus sha256 côté hôte.
- `linux-namespace` + premier démarrage : un mot de passe jetable local à l’essai a été posé
  par le CHR (« n » au contrat de licence, throwaway). Aucun secret persisté ; l’image
  brute du CHR est mutée par chaque premier démarrage (re-déposer l’image épinglée entre
  chaque tentative, ce que toutes les 14 tentatives ont fait).

## Chronotique (preuve par l’échec)

14 fenêtres le 2026-11-16 (≈ 07:58Z–10:12Z), chacune réutilisant la précédente. Le journal
détaillé — chaque ligne nommée, chaque SHA-256, épisodes 1 à 14 — vit dans
`C:\WEBSERVER\APPROVED-2026-11-16-chr-boot.md` (mandate autonome prolongé par le message
de l’opérateur : « do phase 4.4 on the server alongside the current vm ») et dans les
fenêtres `C:\WEBSERVER\evidence\20261116-*-write-window.txt`. En résumé honnête :

- les premiers essais ont montré que le matériel CHR démarre mais le harnais était faux
  (licence, exigence de mot de passe du premier démarrage, écho masqué, tampon torpillé
  par des lecteurs zombies ou des réponses d’attente mal comptées — chaque défaut nommé
  et corrigé dans le dépôt : `cmd/cgo-chr/serial.go`) ;
- les essais intermédiaires ont montré que RouterOS refuse ce qui est réel :
  interface inexistante sur un CHR frais à un NIC (image brute mutée à chaque démarrage,
  re-déposée entre tentatives) ;
- la comparaison d’export a dû être rendue sémantique sur les débits (`19M ≡ 19.0Mbps ≡
  19000000`) et l’absence acceptée uniquement pour le défaut usine (`triple-isolate`) ;
- chaque correctif est un commit du dépôt, chaque sceau raté est conservé sur l’hôte.


## Fichiers ici

- `artifacts/chr-fixture-cake-2026-11-16-r2.json` — le rapport fixture consommé (sha256
  5c7eb9ab…, `fixture: true`, aucune mesure revendiquée).
- `artifacts/export-2026-11-16m.txt` — l’export réel RouterOS 7.24.6 post-import
  (982 octets, sha256 vérifié contre le sceau).
- `artifacts/proof-2026-11-16m.json` — le sceau complet (23/23, exportations, hashes).
- Sur l’hôte : `/home/sadc/labo-cgo/chr/proof-2026-11-16m/{proof,transcript,export}.txt`
  (transcript brut — 5 671 o).

## Verdict

Le rapport scellé du bench (1794798576, mesuré sur un lien réel) n’est PAS consommé ici —
il vit dans la VM de banc. Cette vérification prouve exclusivement que RouterOS 7.24.6
accepte le script de remédiation émis par l’outil, avec une séquence premier-démarrage
complète et un export lisible machine (982 octets, sha256 vérifié).
