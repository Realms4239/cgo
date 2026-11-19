# Link audit — audit-1795091086-252026ff

- Site: bench-calibration
- Link: fiber ()
- Measured: 2026-11-19T15:24:46+03:00
- Sens: descendant (down) (charge: 4 flux)
- Phases: 10 s baseline / 41 s load / 10 s drain
- Reproduction: ` cgo audit -site='bench-calibration' -link-type='fiber' -duration=60 -direction='down' -flows=4 -target='10.30.0.2' -small='http://10.30.0.2:8081/small' -bulk='10.30.0.2:5201' `
- Preuve: ` /home/bench/bench-data/audits/audit-1795091086-252026ff/audit_report.json `

## Measured

| Metric | Value |
|---|---|
| Idle RTT p50 | 100.0 ms |
| Idle RTT p95 | 101.0 ms |
| Loaded RTT p50 | 100.0 ms |
| Loaded RTT p95 | 101.0 ms |
| Queue delay impairment (QDI) | 1.0 ms |
| Bufferbloat delta (p95) | 0.0 ms |
| Queue drain after load | 0.0 ms above rest |
| Throughput | 18.59 Mbit/s |
| VoIP R-score | 91.9 |
| Équité inter-flux (JFI) | 85.4 % |
| Loss | 0.00 % |
| Small-object p95 | 134.0 ms |
| Data consumed by audit | 94.13 MB |

**Bufferbloat grade: A+** — Latence excellente sous charge : l'AQM tient les files, appel et pilotage restent nets.

**Verdict opérateur : Lien sain** — Sous charge, la latence reste sous l'objectif et 100.0 % des alertes passent à l'heure.

**SLA CAP (objectif DGM 220 ms, cadre WIS 2.0) : 100.0 % sous échéance — CONFORME**

## Not measured

Each of these is absent for a stated reason. None is zero-filled.

- **deadline_ok_pct** — cap_deadline_ok_pct covers the fixed 220 ms operator objective; the campaign's operator-defined deadline_ok_pct stays a separate, campaign-only metric
- **remediation** — no interface was named, so no command set could be generated

## Provenance

- Série brute (sonde par sonde): `/home/bench/bench-data/audits/audit-1795091086-252026ff/telemetry.csv`
- SHA-256 télémétrie: `4c1f94a687ffe1be15dc14b517534cbff9655dea1896afbe1d583c0fd892abb9`
- Source: `/home/bench/bench-data/link_audit.csv`
- Rows: 1
- SHA-256: `5944c0cc87a0b9835c672e5caca1842e83339af90d47160b7391ce61a8fdaff4`

## Impact économique (scénarios contractuels)

Ratio de réémission : 0.0000 (base measured) · 0.00 Mo gaspillés sur 94.13 Mo transportés en 41 s de charge.

| Scénario | Ar/h | Base |
|---|---|---|
| enterprise-gold-cir (A) ← PRIMAIRE (fiber) | 0 | (part réémise 0.0000) × 1 000 000 Ar/720 h |
| enterprise-platine1 (A) | 0 | (part réémise 0.0000) × 2 500 000 Ar/720 h |
| b2b-platine-bulk (B) | 0 | 0.00 Mo gaspillés / fenêtre 41 s × 81.38 Ar/Go |
| b2b-m2m-apn (B) | 0 | 0.00 Mo gaspillés / fenêtre 41 s × 150.00 Ar/Go |
| retail-outofbundle (emergency) | 0 | hors-forfait d'urgence (Net Month 4,5 Go) — pénalité, pas un modèle d'infrastructure |

Scénarios contractuels (yas.mg 31/08/2026 ; APN 150 Ar/Go sur devis), pas des mesures : le chiffre retenu dépend du contrat du site, pas de la physique du lien.

