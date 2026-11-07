# Runbook — smoke alerts (S5)

## Auth rotation
1. Generate a new token: `openssl rand -hex 32`
2. Set `CGO_API_TOKEN` in the compose `.env`, `docker compose up -d`
3. Verify `curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:9090/api/health` → 200; without → 401.

## Nginx revoke
Reload Nginx after token rotation: `sudo systemctl reload nginx`. Breach → 429 under mutation burst (1r/s burst=3).

## Rollback
`docker compose pull` prior digest, `docker compose up -d`, then `cgo shape --restore` to clear any edge shaped by the failed release. Verify `/metrics` HELP/TYPE still served.

## Prometheus target-down
Check scrape `GET /metrics` returns `text/plain; version=0.0.4`. If target down, inspect `docker compose logs cgo`. No data is invented — gap, not zero.

## Bench offload fail
`kit/cgo-bench-dualns.sh` fail-closed check must pass before any campaign; if it fails, stop — do not bypass.

## Corpus seal verification
`python book/tools/audit-archive.py` (when `book/` present). Frozen corpus is never edited in place; provenance is additive only.
