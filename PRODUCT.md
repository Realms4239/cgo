# PRODUCT.md

**Meteolink** — banc reproductible pour l'audit client-side et la comparaison des
politiques AQM/BBR sur liens d'accès contraints, appliqué aux liens de
Météo Madagascar.

Thèse : « Conception et évaluation d'un banc reproductible pour l'audit
client-side et la comparaison des politiques AQM/BBR sur liens d'accès
contraints — application aux liens de Météo Madagascar ».

## Product truth

Météo Madagascar's DSI runs critical traffic (telemetry, alerts) over
constrained access links (4G/5G, VSAT, fiber) where bulk transfers — satellite
imagery, model downloads — bloat ISP buffers and stall everything. The DSI has
**no admin access to the ISP core**; the only lever it controls is the edge.

Meteolink is the instrument for that lever:

1. **Audit** — 30 s on any access link, no admin rights, real RTT/small-object
   measurements frozen with a SHA-256 provenance chain.
2. **Shape & compare** — lock a baseline, apply `cake`/`fq_codel` to the
   gateway edge, watch the two traces fight on the same scale, export a signed
   before/after report.
3. **Prove** — lab matrix (profiles × qdiscs × CC) frozen to CSV + manifest,
   results translated into the hardware the DSI actually owns (MikroTik Queue
   Tree, Linux `tc`, edge gateway).

## Users

- **DSI technician** (primary): drives audits and shaping on site, needs
  answers in 5 minutes without reading the source.
- **Network engineer** (secondary): reads traces, medians, gates; wants
  numbers, not decoration.
- **Jury / director** (audience): reads the constat and the cost in Ariary.

## Mode

**Operate** (per DESIGN.md): scanability, truthful state, native expectations.
Brand lives in precise details — hairlines, tabular figures, one tricolor
accent. The memorable element is the live wall: one hero trace, one comparison,
one lever.

## Non-goals

Production-core control (never — client-side by design), oracle worship,
fictional data. Every number on screen comes from a frozen, verifiable CSV or
from the live SSE stream — labeled which is which.
