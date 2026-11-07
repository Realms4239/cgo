# cgo/guard
Bibliothèque d'authentification pour le plan opérateur CGO (stdlib seulement).

- `RequireAuth(next http.HandlerFunc)` — jeton depuis `CGO_API_TOKEN`.
- `RequireAuthToken(want string, next http.HandlerFunc)` — jeton explicite.

Règle : aucune mutation sans jeton sur bind non-loopback ; jeton vide =
mode dev loopback uniquement (401 pour les appelants distants). Le rate
limiting reste à la charge de Nginx (`deploy/nginx/cgo.conf`, zones
`cgo_reads` / `cgo_mutations`, `limit_req_status 503`).

Versionné avec semver ; changements cassants = MAJOR.
