# cgo/attest
Vérificateur de scellés SHA-256 pour les preuves CGO (stdlib seulement).

- `attest.Parse`, `attest.VerifyManifest`, `attest.VerifyFile` — bibliothèque.
- `cmd/cgo-attest` — CLI : `cgo-attest --manifest <fichier>` ; sortie 0 si
  tous les scellés passent, 1 au premier écart.

Manifeste : une ligne par fichier, format `<chemin> <sha256hex>`.
Le corpus gelé (`book/data/runs`, schémas CSV) n'est jamais modifié :
la vérification est en lecture seule, la provenance additive seulement.

Versionné avec semver ; changements cassants = MAJOR.
