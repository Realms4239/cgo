# Task 4.5 service map packet

Nothing in this packet has been executed. CGO has not contacted, read from, or
written to any production system. This document contains the inputs the operator
and the DSI reviewer must supply, plus the blank templates that turn those inputs
into the countersigned case study. The governing contract is
`docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md`.

## Request to the operator

Send this text unchanged. It is the written grant the contract requires.

> To anchor the study in your infrastructure without touching it, we ask for:
> (a) a configuration and link dump you consider non-sensitive, (b) your service
> list with cadences and deadlines, (c) whether we may name the service or must
> use a pseudonym, (d) who signs off on what may appear in a public thesis. We
> will run no load, install nothing, and send nothing to your network; every
> artifact will be shown to you before it leaves our machines.

Acceptance of items (a) to (d) is the unblock condition. Until the written grant
arrives, Task 4.5 stays BLOCKED with that named condition.

## Evidence the operator supplies

The operator acquires every item. CGO names no host, address, credential or
topology element until the grant permits it. The default subject name is
*the national forecasting server*.

| Item | Form | Used for |
|---|---|---|
| Edge configuration | Operator-provided dump | Service classes and termination points |
| Link configuration | Operator-provided `tc qdisc` output | Control point and real capacity |
| Interface counters | Two read-only snapshots | Whether the bottleneck is on the access link |
| Service inventory | Operator's own list | Classes A (alerts), B (telemetry), C (bulk) |
| Incident notes | Operator-provided excerpts | Missed deadlines, without inferred causes |

## Service map template

Each statement carries one confidence label. *Declared* means the operator wrote
it. *Observed* means it appears in a supplied dump. *Inferred* means CGO derived
it and the operator must confirm it before publication.

| Class | Service example | Termination | Direction | Confidence |
|---|---|---|---|---|
| A | [operator input] | [operator input] | [up / down / both] | [declared / observed / inferred] |
| B | [operator input] | [operator input] | [up / down / both] | [declared / observed / inferred] |
| C | [operator input] | [operator input] | [up / down / both] | [declared / observed / inferred] |

## Workload model template

| Class | Size distribution | Cadence | Deadline or completion window | What the operator calls late |
|---|---|---|---|---|
| A | [operator input] | [operator input] | [operator input] | [operator input] |
| B | [operator input] | [operator input] | [operator input] | [operator input] |
| C | [operator input] | [operator input] | [operator input] | [operator input] |

## SLA candidate

The standing P2 objective is a 220 ms CAP small-object p95, recorded in
`docs/MUSTREAD/INDEX.md`. The operator either confirms it for the observed link
or replaces it with their own target.

| Field | Value |
|---|---|
| Objective | [220 ms confirmed / operator target stated below] |
| Operator target | [value in ms] |
| Source of the target | [contract / dashboard / incident note, named by the operator] |

## Experiment matrix template

Map the operator's classes to cells of the existing bench. At least one row must
remain unreproducible, with the reason stated. A matrix claiming total fidelity
is refused by the contract.

| Class | Reproducing cell | Profile | Qdisc | Congestion control | Reproducible | Reason if not |
|---|---|---|---|---|---|---|
| A | [cell or none] | [P1 to P4] | [qdisc] | [cc] | [yes / no] | [operator or builder input] |
| B | [cell or none] | [P1 to P4] | [qdisc] | [cc] | [yes / no] | [operator or builder input] |
| C | [cell or none] | [P1 to P4] | [qdisc] | [cc] | [yes / no] | [operator or builder input] |

## Remediation plan template

The operator applies the remediation at R2. Generated configuration is labelled
*generated-only* until a twin cell or an operator-applied change measures it.

| Field | Value |
|---|---|
| Control point | [from supplied `tc qdisc` output] |
| Generated configuration | [reviewed script, no live target] |
| Evidence level | [twin-measured / generated-only] |
| Applier | [named operator role] |
| Rollback | [operator-defined] |

## Before and after measurement template

The operator fills both instants. CGO computes no number from an operator
dashboard without its provenance and approval.

| Metric | Before | After | Source | Operator approval |
|---|---|---|---|---|
| Small-object p95, ms | [input] | [input] | [measurement name] | [pending / granted] |
| Bulk goodput, Mbit/s | [input] | [input] | [measurement name] | [pending / granted] |
| Cost, Ariary per hour | [input] | [input] | [contract or invoice] | [pending / granted] |

## DSI countersign

| Field | Value |
|---|---|
| Service map reviewed by | [DSI name and role] |
| Figures approved for the thesis | [list or none] |
| Pseudonym required | [yes / no] |
| Signature | [DSI signature] |
| Date | [DSI date] |

## Acceptance checks

1. Every service map row has a confidence label.
2. No artifact contains a hostname, address, credential or topology element
   outside the grant.
3. The experiment matrix names at least one cell the twin cannot reproduce.
4. `docs/VERIFY.md` gains rows only for what was actually read or measured.
   Everything else stays BLOCKED with its condition named.
5. The DSI signature is present before any figure leaves the machines.

## What CGO does not do

CGO sends no packet to the operator's infrastructure. It runs no probe, load,
port scan, SNMP walk, API call or DNS query. It installs nothing, asks for no
root on the production host and copies no dashboard number into a figure without
provenance and approval.
