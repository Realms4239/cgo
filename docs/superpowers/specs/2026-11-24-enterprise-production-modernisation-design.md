# Enterprise production modernisation — design (2026-11-24)

Authority: operator chat 2026-11-24 asked to use `last30days` and the
`C:\spectral-rd\batches\network-engineer` corpus (especially IaC, observability
and APIs for networking) for a new plan that modernises this stack into
enterprise production, then approved all four slices.

Research: `last30days` deep run (134.3 s, 61 items; social clusters off-topic
because "enterprise" matched AI-agent stories, so treated as absent), with 13
citable sources appended to the saved raw file. Local corpus:
`C:\spectral-rd\batches\network-engineer\module-14.md` (APIs for Networking,
YANG/NETCONF/RESTCONF, Ansible, load balancing) and `module-15.md` (Terraform,
IaC, NetFlow/sFlow, observability, SNMP, Prometheus, Grafana). Primary anchors:
[OpenTofu state/plan encryption](https://opentofu.org/docs/language/state/encryption),
[OBI v1.0](https://opentelemetry.io/blog/2026/obi-v1-release),
[OBI limits](https://opentelemetry.io/docs/zero-code/obi),
[Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway),
[SIG filter reference](https://gateway-api.sigs.k8s.io/reference/api-spec/main/spec).

No frozen corpus, book, thesis, gate, permit, ledger or protected config edit.
No new runtime dependency, no Kubernetes cluster, no fabricated readiness.

## 1. Scope

Four sequential SDD slices, each with spec -> plan -> fresh implementer ->
fresh reviewer -> gates -> explicit-path commit:

1. Reproducible testbed (declarative topology, drift, fail-closed preflight).
2. Observability + SLO (real metrics from audit/seal/fidelity, declared error budget).
3. Networking API boundary (auth, ownership, rate-limit policies, conformance-style checks).
4. Production hardening (CI supply chain, rollback/runbook proof, release versioning).

The existing Go/React service, shell bench executor and frozen corpus stay intact.

## 2. Reproducible testbed

- Add a declarative topology contract under `kit/` describing namespaces,
  interfaces, qdiscs, capacity, delay and the shape mode (egress/ingress).
  The bench shell remains the executor; the contract is the reviewable source.
- A `check` verb compares declared topology to live `ip`/`tc` state and exits
  non-zero on drift, with no mutation. Run it before any campaign cell.
- A versioned fingerprint of the contract is included in the run manifest next
  to `config_sha256`, so a row can be traced to the exact topology contract.
- No OpenTofu/Terraform binary is introduced: the research drives state and
  drift discipline, not a new dependency. Human gate remains for live bench.

## 3. Observability + SLO

- The existing `/metrics` exporter publishes only real values from audits,
  seals and fidelity verdicts. Add exactly these series:
  `cgo_audit_info{audit_id,source,age_bucket}`,
  `cgo_kernel_truth_source_info{source}` (emitted only when a source exists),
  `cgo_seal_present{audit_id}` (0/1 from a real seal artifact),
  `cgo_emd{live_basis}` and `cgo_emd_ceiling` (from `metrics.FidelityEMDMax`),
  and `cgo_gate_status{gate}` per named native gate.
- Absence is a labelled series value or omitted series, never a fabricated zero.
- Declare an SLO and error budget in config: availability of a complete audit
  report, and fidelity within the declared EMD ceiling. Burn rate and budget
  remaining are computed from real observations; no green without a sample.
- No new exporter dependency; keep the existing Go exporter handler.

## 4. Networking API boundary

- Keep the current mux but document and test a gateway-style boundary:
  authentication at the edge, ownership guards on resource routes, explicit
  rate limits per principal with 429 and Retry-After: 60 requests/minute for
  read routes, 10 requests/minute for mutating routes, burst 20 and 5
  respectively. These are declared here and tested, never widened at runtime.
  Structured 401/403 bodies never leak paths or secrets.
- Add a conformance-style test suite that asserts, for every mutating route,
  that unauthenticated, cross-owner and rate-limited requests are refused
  before the handler runs. Read routes remain honest absence, not blank.
- Follow the research distinction: routing/conformance is core; auth and rate
  limiting are implementation policy owned by this service, with their limits
  declared in the spec and tested.

## 5. Production hardening

- CI must keep the six native gates green; add supply-chain checks already
  permitted by the repo (pinned actions, checksum attestation, no new dependency
  without review) without editing gate definitions.
- Every destructive/live operation keeps a written rollback and evidence
  sidecar; add a runbook index that names each operation, its token, rollback
  and last rehearsal date, without changing protected ledgers.
- Release versioning stays human-gated; the service reports its own version
  and build provenance through `/api/capabilities`.

## 6. Acceptance and limits

RED-first hostile oracles for drift refusal, metric absence, auth/ownership/
rate-limit refusal and runbook completeness. Gates run once per slice. No
VERIFY-row promotion from builder passes. Known parks (eBPF object rebuild,
live p99.9, valid up direction) remain explicitly unclaimed and are not silently
folded into this programme.
