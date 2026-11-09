# WIS2 sign-off form — operator signature (Q4, Q5, dissemination scope)

> **Status:** ready for the operator to sign. Nothing below is implemented
> until the bracketed fields are filled and signed here. This document
> closes the human-autonomous part of `docs/runbooks/wis2-q-research.md`;
> the design (`docs/runbooks/wis2-probe.md`) stays DEFERRED until this form
> is signed and Q1–Q3/Q6-Q7 (already PROVEN-a) are folded into the slice.
>
> Date: 2026-11-09 · Prepared by builder (no signature implied).

## 1. Q4 — centre-id status

| Field | Entry (fill one) |
|---|---|
| Have a centre-id today? | [ ] Yes → `mg-<name>`: ______________  [ ] No → see written decision below |
| Anticipated centre-id (if pending assignment) | `mg-<name>`: ______________ |
| Written non-registration decision (if not registering) | The operator decides the probe remains a **lab-only** sender under the namespace `cgo/v1/probe/<site-id>/…` and never requests a centre-id. Signed: ______________ date: ______ |

- Source: Codes Registry snapshot 06 Jan 2026 — no `mg-*` present
  (negative finding only, `wis2-q-research.md` Q4).
- With a centre-id: the mapping in `wis2-probe.md` uses
  `origin/a/wis2/<centre-id>/data/recommended/cgo/link-quality`.
- Without: the lab namespace is permanent and the probe is never a
  WIS2 Node.

## 2. Q5 — CAP severity bands (`cap_ok_pct` → urgency/severity/certainty)

Domains per CAP 1.2 §3.2.2 (PROVEN-a); **no numeric mapping exists in any
primary source** — the boundaries below are operator policy, to be signed
before implementation.

| `cap_ok_pct` band | `<urgency>` | `<severity>` | `<certainty>` | `<category>` | `<scope>` |
|---|---|---|---|---|---|
| ≥ 95 % | *(no CAP, JSON only)* | — | — | — | — |
| < 95 % → 90 % | [ ] Immediate  [ ] Expected  [ ] Future | [ ] Extreme  [ ] Severe  [ ] Moderate  [ ] Minor | [ ] Observed  [ ] Likely  [ ] Possible | [ ] Infra  [ ] Other: ______ | [ ] Private  [ ] Restricted  [ ] Public |
| < 90 % → 80 % | [ ] Immediate  [ ] Expected  [ ] Future | [ ] Extreme  [ ] Severe  [ ] Moderate  [ ] Minor | [ ] Observed  [ ] Likely  [ ] Possible | [ ] Infra  [ ] Other: ______ | [ ] Private  [ ] Restricted  [ ] Public |
| < 80 % | [ ] Immediate  [ ] Expected  [ ] Future | [ ] Extreme  [ ] Severe  [ ] Moderate  [ ] Minor | [ ] Observed  [ ] Likely  [ ] Possible | [ ] Infra  [ ] Other: ______ | [ ] Private  [ ] Restricted  [ ] Public |

Other CAP fields to sign: `<sender>` domain `__________________`,
`<senderName>` `__________________`, recovery → `Update` + `<references>`?
[ ] Yes  [ ] No, retraction → `Cancel` [ ] Yes  [ ] No.

Signed (bands + fields): ______________ date: ______
<!-- Example shape only. `95` matches the `cap_sla.target_pct = 95`
     default seen in DashboardView; bands are NOT implied by it. -->

## 3. Dissemination-scope decision

| Question | Answer (fill one) |
|---|---|
| Probe transport | [ ] Lab only (`cgo/v1/…`, own broker)  [ ] wis2box reference Node (Q6) |
| Data policy per dataset | `link-quality`: [ ] core  [ ] recommended |
| Public Internet stance | [ ] stays lab  [ ] registers a real Node after Q4 = Yes + GISC |

Signed (transport + data policy + stance): ______________ date: ______

## Effect of signing

- S0 (reads) becomes implementable; S1 (lab broker) and S2 (producer) need
  one more human approval: the new Go dependency (`github.com/eclipse/paho.golang`
  recommended, Q7 PROVEN-a) per the no-new-dep rule.
- S3 (CAP producer) starts only after Q5; S4 (WIS2 publication) only after Q4.
