# cgo/api
REST/SSE surface for CGO: `New(d Deps) Handler`, SSE hub, Prometheus `/metrics`, profile and audit routes.
Stdlib only (plus cgo/audit, cgo/capabilities, cgo/figures, cgo/metrics, cgo/model, cgo/probe, cgo/profile, cgo/remediate, cgo/results). Versioned with semver; breaking changes bump MAJOR.

`frontend/` reste un SPA stub — le vrai dist vit dans web/frontend jusqu'à la bascule. Les copies vendored `campagne/` et `profile/` ont été supprimées : le module requiert les modules publiés `capabilities` et `profile` (tags `capabilities/v0.1.0`, `profile/v0.1.0`).
