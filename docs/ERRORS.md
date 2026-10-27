# CGO — error ledger

Append-only, newest at the bottom. An error without a cause is a rumor; a fix
without a proof is a wish. Entries found in earlier sessions are labelled
`recorded`, not `observed today`.

## Entry format

```markdown
## E<n>. <title> (YYYY-MM-DD)
- Symptom:
- Cause:
- Fix:
- Proof:
- Lesson:
```

A fixed defect stays because it is the evidence that it was understood.

## E1. dualns `--target` overwrite (recorded 2026-10-27)

- Symptom: a dualns run started without `--target` probes `1.1.1.1` instead of
  the bench sink `10.30.0.2`; the campaign then measures an unreachable target
  and its cells are empty or invalid.
- Cause: `campagne.ProdDepsDualNS` sets `Target` from `CGO_TARGET` (default
  `10.30.0.2`), but `runCLI` in `cmd/cgo/runcli.go` then assigns
  `deps.Target = target` unconditionally; the CLI flag default is `1.1.1.1`, so
  an omitted flag replaces the bench default.
- Fix/workaround: pass `--target 10.30.0.2` explicitly and run the campaign
  inside `ns_client`; the root fix is to let an unset flag keep the bench
  default.
- Proof: recorded in the umbrella spec and the harmonization matrix; no live
  re-run in this session.
- Lesson: a default that replaces an explicit flag is a false configuration.

## E2. Tableau 15 regime mix (recorded 2026-10-27)

- Symptom: Tableau 15 presents CUBIC cells beside P2/up/BBRv1/1-flow cells, but
  the CUBIC cells are invalid on G4 and cannot support a comparison.
- Cause: the table mixes bench regimes; the CUBIC cells were measured on the
  legacy single-flow bench, outside their valid regime.
- Fix/labelling rule: label the table as a legacy-bench regime, mark the CUBIC
  cells invalid on G4, and do not cite them as a valid comparison.
- Proof: recorded in the umbrella spec and the corpus inventory; no live re-run
  in this session.
- Lesson: a table that mixes bench regimes is not a comparison.
