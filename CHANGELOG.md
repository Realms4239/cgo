# Changelog

## 2.0.0 — 2026-11-25

Release candidate `2.0.0-rc.1` promoted to `2.0.0` after a full green verification pass:
`npm -C web/frontend run typecheck` exit 0, `npm -C web/frontend test -- --run`
584/584 passed, `sh .githooks/run-gates.sh` `6 passed, 0 failed, 0 not run`.

### Enterprise production modernisation

- Reproducible testbed: declarative topology contract, read-only drift check, additive
  manifest fingerprint (`kit/topology.contract.json`, `cmd/topocheck`).
- Observability and SLO: real audit, seal and fidelity series on `/metrics`, SLO and
  error-budget helper; no fabricated metric or green without a sample.
- Networking API boundary: per-principal read/mutate rate limits, structured 401/403
  bodies, cross-owner refusal asserted on every mutating route.
- Production hardening: runbook index with explicit gaps and supply-chain proof; no
  rehearsal date, token id or CI result was invented.

### Reliability and interface

- Root-cause fix for the native `tests` gate: `waitAuditIdle` now treats HTTP 429 as a
  busy audit and waits `Retry-After` instead of misreading the error body as idle.
- Cockpit refinement inside the pinned institutional identity: shared type/rhythm
  tokens, aligned decision column and instrument field, with DOM oracles.
- E2E suite retargeted to the live two-mode UI (`Synthèse` / `Labo`); removed legacy
  four-view selectors, mocks narrowed, and a static oracle now forbids their return.
- Title oracle aligned with the adopted `PRODUCT.md` title; the frozen book title is
  kept as a separate, honest oracle.

### Evidence and limitations

- Host-side Go verification recorded under
  `docs/verification/2026-11-25-host-go-verification/` with sha256 sidecars.
- No frozen corpus, gate, permit, ledger or protected config change.
- Live bench work remains a separate human-gated scope; no live cell is part of this
  release. Browser E2E must be re-run against the real Go server before the release
  can claim a live-browser demonstration; see
  `docs/verification/2026-11-25-cockpit-refinement/README.md`.
