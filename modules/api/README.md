# cgo/api
REST/SSE surface for CGO: `New(d Deps) Handler`, SSE hub, Prometheus `/metrics`, profile and audit routes.
Stdlib only (plus cgo/audit, cgo/figures, cgo/metrics, cgo/model, cgo/probe, cgo/remediate, cgo/results). Versioned with semver; breaking changes bump MAJOR.

Compat shims (vendored until the composition layer is promoted): `campagne/` (CapabilitiesNow subset), `profile/` (full copy), `frontend/` (SPA stub — the real dist stays in web/frontend until cutover).
