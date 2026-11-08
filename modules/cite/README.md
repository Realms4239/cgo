# cgo/cite
Citation guard for sealed CGO audit reports.
Canonical implementation: `web/frontend/src/lib/citation.ts`
(`auditCitation(report) string`); Go port `cite.go` (`AuditCitation`)
mirrors the same rule and is covered by `cite_test.go`.

## Rule (exact)

The citation is `""` when `!report.provenance?.source_sha256`;
no partial citation is ever returned without a seal.

- Inputs: `AuditReport.provenance.source_sha256`, `AuditReport.audit_id`.
- Output: `auditCitation(report) string` — `""` without
  `source_sha256`; otherwise an RFC-style citation carrying the full
  `source_sha256` seal (`Rapport scellé : source_sha256 <seal>.`).
- Empty or missing `audit_id` also returns `""`.

## Seal format

In `cgo-cite` v2 the seal format `/^[0-9a-f]{64}$/i` becomes required.
Until v2, only presence (after trim) is enforced by the guard;
the 64-hex check is the documented contract, not yet a runtime gate.

## Verification

`cd web/frontend && npx vitest run src/lib/citation.test.ts` — 4/4 PASS
(seal guard, empty-id guard, sealed path).
Go: `cd modules/cite && go test ./...` — sealed / omits / missing-seal /
empty-id cases mirror the TS tests.

Versioned with semver; breaking changes bump MAJOR.
