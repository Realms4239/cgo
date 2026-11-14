# cgo/api
REST/SSE surface for CGO: `New(d Deps) Handler`, SSE hub, Prometheus `/metrics`, profile and audit routes.
Stdlib only (plus cgo/audit, cgo/capabilities, cgo/figures, cgo/metrics, cgo/model, cgo/probe, cgo/profile, cgo/remediate, cgo/results). Versioned with semver; breaking changes bump MAJOR.

Le SPA réel (`web/frontend/dist`) est injecté via `Deps.StaticFS` par `cmd/cgo` — le module api ne peut pas importer le module racine (cycle) ; `api/frontend` a été supprimé. Les copies vendored `campagne/` et `profile/` ont été supprimées : le module requiert les modules publiés `capabilities` et `profile` (tags `capabilities/v0.1.0`, `profile/v0.1.0`).
