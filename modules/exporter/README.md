# cgo/exporter
Collecteur Prometheus honnête pour le plan opérateur CGO (stdlib seulement).

- `Render(samples, help) string` — exposition pure : `# HELP` / `# TYPE`
  avant le premier échantillon ; suffixe `_total` ⇒ `counter`, sinon `gauge`.
- `Handler(collect)` — sert avec `Content-Type: text/plain; version=0.0.4`.

Contrat d'honnêteté : absent n'est jamais `0` — une série absente n'émet
aucun échantillon (Prometheus = « no sample », jamais `0`). Voir
`pkg/api/metrics.go` et les tests `TestMetrics_*` pour le collecteur réel.

Versionné avec semver ; changements cassants = MAJOR.
