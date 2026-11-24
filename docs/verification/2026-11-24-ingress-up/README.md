# 2026-11-24 — ingress shaping and up-direction validation

Builder pass under `APPROVED-2026-11-24-ingress-up.md` plus three controller
amendments (launcher env, engine boundary, egress/up cell). All host contact via
`safe_write.py`; topology restored to egress at the end. No gate, ledger,
corpus or frozen-file edit; no VERIFY-row promotion.

## Preflight

`run-dualns.sh up --ingress on` produced IFB0 UP with a cake ingress qdisc:
`bandwidth 20Mbit ... ingress ... rtt 100ms`; `ip -o link show ifb0`,
`tc -s qdisc show dev ifb0` and `tc -s filter show dev c-cli parent ffff:`
all returned 0; netem on `c-wan` was `delay 50ms`. The generic bench check
returned `CHECK_RC=1` because e-wan had no cake qdisc - expected in ingress mode
(the shaper is on IFB), not a missing IFB. Evidence `8a460bc7…`.

The campaign reads `CGO_SHAPE_MODE` (default `egress`), so a guest launcher with
`CGO_SHAPE_MODE=ingress` was staged (evidence `303df8ce…`); otherwise all cells
would have been refused on topology mismatch.

## Results

| Run | Mode | Direction | rtt_p50 / p95 | drops | goodput | gate |
|---|---|---|---:|---:|---:|---|
| `1795509754` | ingress cake | down | 101.0 / 120.0 ms | 6 | 3.7 Mbps | valid |
| `1795509122` | ingress fq_codel | down | 101.0 / 120.0 ms | 560 | 1.5 Mbps | valid |
| `1795510344` | egress cake | up | 100.0 / 120.0 ms | 10 | 5.2 Mbps | **invalid** |
| `1795507817`, `1795508506` | ingress | up | none | - | - | engine-refused |

Engine-enforced boundary: `shape_mode ingress invalide en direction "up" -
down uniquement (le shaper ingress ne corrige pas l'upload)`. Both ingress/up
cells were refused before measurement; this is a correct guard, not a defect.
The egress/up cell did produce RTT samples but `gate_status: invalid`, so it is
NOT a valid measurement and no SLA or physics claim is made from it.

## What this proves and what it does not

- Ingress shaping can be configured and checked on the real bench with IFB and
  cake/fq_codel; both down cells produced valid measurement-integrity rows.
- Cake ingress and fq_codel ingress have very different drop behaviour on the
  same profile (6 vs 560 drops) - consistent with the documented ingress
  semantics (drops counted as transferred, >=2 packets queued per flow). This
  is a shape-mode difference, not an AQM ranking.
- The up direction remains unproven: ingress is engine-refused and the only
  egress/up row is invalid. A valid up-direction measurement needs a separate
  corrected cell design, not a relaxation of `gate_status`.
- p99.9 is not claimed here: the histogram fields are not yet wired into the
  production CLI path, and no per-packet tail source was available in these runs.
