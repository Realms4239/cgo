# Link audit — audit-1795091574-1fd85cca

- Site: bench-calibration
- Link: fiber ()
- Measured: 2026-11-19T15:32:54+03:00
- Sens: descendant (down) (charge: 4 flux)
- Phases: 10 s baseline / 0 s load / 10 s drain
- Reproduction: ` cgo audit -site='bench-calibration' -link-type='fiber' -duration=60 -direction='down' -flows=4 -target='10.30.0.2' -small='http://10.30.0.2:8081/small' -bulk='10.30.0.2:5201' `
- Preuve: ` /home/bench/bench-data/audits/audit-1795091574-1fd85cca/audit_report.json `

## Measured

| Metric | Value |
|---|---|
| Idle RTT p50 | 110.0 ms |
| Idle RTT p95 | 111.0 ms |
| Loss | 0.95 % |
| Data consumed by audit | 0.00 MB |

**Bufferbloat grade: not measured.** Bufferbloat non mesuré — aucune charge bulk exploitable.

**Verdict opérateur : Charge non établie** — Le flux saturant n'a pas pu démarrer ; la tenue sous congestion n'est pas mesurée.

**SLA CAP : non mesuré** — aucune complétion 16 KiB chargée ; jamais un succès par défaut.

## Not measured

Each of these is absent for a stated reason. None is zero-filled.

- **economics** — économie entreprise — fenêtre de charge vide ; aucun scénario n'est chiffré plutôt qu'un zéro fabriqué
- **bufferbloat grade** — no bulk load ran, so the loaded window was never under load; a grade here would be fabricated
- **throughput** — no bulk sink reachable; supply -bulk host:port or install iperf3
- **voip_r** — requires a measured headline load window; not computed from an empty headline window
- **http_small_p95** — no small-object URL supplied; ping RTT is not a substitute for object completion time
- **drain_ms** — audit shorter than 20 s, so no recovery window was collected; drain is not the same as zero
- **deadline_ok_pct** — cap_deadline_ok_pct covers the fixed 220 ms operator objective; the campaign's operator-defined deadline_ok_pct stays a separate, campaign-only metric
- **remediation** — capacity or idle RTT unmeasured; a queue config sized from a guess would be deployed blindly

## Provenance

- Série brute (sonde par sonde): `/home/bench/bench-data/audits/audit-1795091574-1fd85cca/telemetry.csv`
- SHA-256 télémétrie: `e130dc842261bf06caebbef77ab2044141d6e7766b1e1e06d877797adbfad434`
- Source: `/home/bench/bench-data/link_audit.csv`
- Rows: 3
- SHA-256: `8587023c0ecf2a5d8c850c40da425daeffd1a46dbab20d7c492a433815ae3399`

