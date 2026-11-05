# Case study contract — read-only service map of the national forecasting server

*Status: BLOCKED on the operator's written grant. Nothing in this document has
been executed, and nothing executes until the grant exists. Written
2026-10-30 from the decision recorded in session `ses_e89dc0fe1ffewQxG5Nf3d0NlPq`
(“the honest product is: simulate in the twin, show the evidence, generate the
reviewed configuration, let the operator apply it”).*

## 1. Why this exists

The bench is reproducible and the audit is client-side, but neither names a
real service. This case study would anchor the thesis in the operator's actual
link — read-only, without load and without writes — and feed the twin with the
service mix it currently lacks.

## 2. What would be read (and nothing else)

| Item | How | Why |
|---|---|---|
| Nginx/edge configuration | a configuration dump provided by the operator | which service classes exist and where they terminate |
| `tc qdisc` / link rate | command output provided by the operator | the real control point and the real capacity |
| Interface counters | output of a read-only command at two instants | whether a bottleneck is on the access link |
| Service inventory | the operator's own list | A (alerts), B (telemetry), C (bulk) classes |
| Incident notes | operator-provided excerpts | deadlines that were missed, without inferring causes |

Every item is supplied by the operator or read by them. CGO itself connects to
nothing.

## 3. What is never done

- No packet capture, no probe, no load, no write, no root, no login from CGO.
- No port scan, no SNMP walk, no API call, no DNS query against the operator's
  infrastructure.
- No hostname, address or topology in any artifact, unless the operator grants
  it in writing; the default is the pseudonym *the national forecasting
  server*.
- No number from the operator's dashboards is copied into a figure without its
  provenance and the operator's approval.

## 4. Deliverables if the grant exists

1. **Service map** — classes A/B/C, where each terminates, which direction it
   traverses, and the confidence level of each statement (declared by the
   operator vs observed in a dump).
2. **Workload model** — for each class: size distribution, cadence, deadline
   or completion window, and what the operator considers “late”.
3. **SLA candidates** — the 220 ms CAP objective confirmed or replaced by the
   operator's own target, with the source of the target named.
4. **Experiment matrix** — which cells of the existing bench reproduce the
   observed mix, and which cells cannot be reproduced and why.
5. **Remediation plan** — generated config for the identified control point,
   labelled with its evidence level (twin-measured vs generated-only).

## 5. Acceptance

- Every statement in the service map is labelled *declared*, *observed* or
  *inferred*; no unlabelled claim.
- No artifact contains a hostname, address, credential or topology element
  beyond what the grant allows.
- The experiment matrix names at least one cell that the twin cannot reproduce;
  a matrix that claims total fidelity is refused.
- `docs/VERIFY.md` gains rows only for what was actually read or measured; the
  rest stay `BLOCKED` with the condition named.

## 6. The request to put to the operator (exact text to adapt)

> To anchor the study in your infrastructure without touching it, we ask for:
> (a) a configuration and link dump you consider non-sensitive, (b) your
> service list with cadences and deadlines, (c) whether we may name the
> service or must use a pseudonym, (d) who signs off on what may appear in a
> public thesis. We will run no load, install nothing, and send nothing to
> your network; every artifact will be shown to you before it leaves our
> machines.

## 7. Unblock condition

A written grant covering §6(a)–(d). Until then this slice executes nothing, and
the twin remains the only source of results.
