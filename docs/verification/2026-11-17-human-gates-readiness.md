# Human gates — readiness audit (2026-11-17T10:47Z)

Engineering is done and verified (`d61695e`, gates 6/6). This audit answers one
question per human gate: **what is ready, what is missing, who must act.**
Nothing here was executed: no book/ write, no push/tag/merge, no secret, no
force-purge, no certificate.

## Gate 1 — Case study / DSI countersign (plan 4.5, VERIFY row 74): BLOCKED on you

The contract `docs/superpowers/specs/2026-10-30-cgo-case-study-contract.md`
requires a **written operator grant** before anything is read. It names four
inputs only you can provide:

1. a non-sensitive configuration and link dump (Nginx/edge config, `tc qdisc`,
   link rate, interface counters at two instants);
2. the service list with cadences and deadlines (classes A/B/C);
3. naming or pseudonym (default: *the national forecasting server*);
4. sign-off on what may appear publicly.

Then the R2 action (operator applies the remediation) and the DSI countersign
are yours. CGO itself must not connect, probe, load, write, root, scan or query.
Already available from this campaign: measured before/after p95, cost in Ariary,
real RouterOS script (CHR proof `91d67458…`), 146-cell corpus. These feed the
deliverables but do **not** replace the grant or the countersign.

**I can do next, without the grant:** draft the case-study skeleton outside
`book/`, with every claim labelled *declared / observed / inferred*, measured
numbers slotted in, and explicit `EN ATTENTE` markers where operator facts go.
That is preparation, not a countersigned case study.

## Gate 2 — Signed release (plan 1.3): BLOCKED on secrets + publish decision

Wiring exists in `.github/workflows/release.yml`: `id-token: write`,
`attestations: write`, syft SBOMs per module, `attest-build-provenance@v1`,
`cosign attest-blob` gated on secret `CGO_IMAGE_DIGEST`.

**Finding to decide, not silently change:** the cosign step ends with
`|| true`. A release can therefore be published even if signing fails. That
may be deliberate (keyless OIDC unavailable in some contexts), but it means
"signed release" is not actually enforced by this workflow. Strengthening it
is your call.

**Needs from you:** whether to remove the `|| true`, the `CGO_IMAGE_DIGEST`
value (via the secret tool, never pasted in chat), and the decision to trigger
a tag run.

## Gate 3 — Docker hardening (plan 1.5): BLOCKED on environment

`docker` CLI is **absent** on this host. The Dockerfile/gates can be reviewed
and `bash -n`/config linted here, but `docker build` and the non-root/TLS/token
checks cannot run. **Needs from you:** a host with Docker, or approval to treat
this as a named environment skip.

## Gate 4 — Thesis PDF (plan 6.1–6.7): READY TO BUILD, needs frozen-data approval

`book/` is git-clean, all audits pass, Word COM is available (v14.0),
`book/tools/build-book.ps1` drives the full chain, and `cgo figures` is wired.
But `book/` is **human-frozen** under AGENTS.md and the final plan requires
regenerating figures from the v2 corpus before the PDF. **Needs from you:**
explicit approval to write to `book/` (content and figures), and confirmation
of the v2 numbers as the thesis source. I have not touched it.

## Gate 5 — `.dist` calibration (plan 4.2): BLOCKED on the clock

Field trace PID 3442899 is alive until **2026-11-17T16:22:27Z** (~5.5 h away).
No sample count is readable through the safe channel (operator-owned output
dir). After the seal: collect `flows.csv`/`manifest.json`, require ≥30 samples,
then run the KS calibration. I cannot accelerate this honestly.

## Gate 6 — Tag / push / merge to main (plan 6.7): pure decision

No action taken. Needs your explicit `go` for each of: tag `v2.0.0`, push,
merge to `main`. Never done by the builder.

## Order that makes sense

1. **Now, no prerequisites:** I draft the case-study skeleton (Gate 1 prep) and
   the release-workflow finding (Gate 2) as reviewable artifacts.
2. **You supply:** grant + service list (Gate 1); signing decision/secret (Gate 2).
3. **16:22Z:** `.dist` calibration after the seal (Gate 5).
4. **Then:** thesis build, tag, push, merge (Gates 4–6), in that order.
