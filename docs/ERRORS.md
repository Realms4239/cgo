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

## E3. Renaming `c-cli` flushes its default route (observed 2026-10-27)

- Symptom: after `ip link set c-cli name c-cli-gap` and back,
  `ping 10.30.0.2` from `ns_client` returns "Network is unreachable"; the bench
  itself is up.
- Cause: Linux drops an interface's routes when its name changes; the twin's
  default route was added by hand
  (`ip route add default via 10.10.0.2 dev c-cli initcwnd 50`) and nothing
  re-adds it.
- Fix: re-add the same route after the rename; the sink answers again (302 ms).
  A future gap test should capture the route before the rename and restore it.
- Proof: the gap capture shows `20 absent`, `1 warmup`, `65 valid`; after the
  restore, `46 valid` frames and the ping succeed. Recorded in `docs/VERIFY.md`.
- Lesson: an interface rename is not L3-transparent; measure, then restore.

## E4. `run-dualns.sh down` leaves root-owned sink and dashboard (observed 2026-10-27)

- Symptom: after `down`, `ip netns list` is empty but `pgrep` still shows
  `cgo-linux testbedsrv` and `cgo-linux --serve`.
- Cause: `down` runs an unprivileged `pkill -f "testbedsrv -http 10.30.0.2"`
  against a root process started with `$SUDO`, and it never stops the
  dashboard, which runs via `exec $SUDO ip netns exec ns_client ...`.
- Fix: stop them with privilege, e.g.
  `sudo pkill -f "[c]go-linux --serve"; sudo pkill -f "[t]estbedsrv -http 10.30.0.2"`;
  the bracket keeps the invoking shell from matching itself. Script fix
  deferred: pre-existing and outside slice 10's measurement contract.
- Proof: after the manual stop, `pgrep`, `ip netns list` and `ss -ltnp` are
  empty on CGO-DEV; the VM stays up.
- Lesson: deleting namespaces does not kill the processes inside them.

## E5. `TestSkipJournalsViaHook` flakes on Windows temp cleanup (observed 2026-10-28)

- Symptom: `go test ./...` fails intermittently with
  `TempDir RemoveAll cleanup: unlinkat ...\run-skip-j: Le répertoire n'est pas
  vide` (the directory is not empty), in a clone run of the `tests` gate; the
  same gate passed before and after.
- Cause: the test waits at most 3 s for `IsRunning()` to become false, but the
  next cell's `Small` blocks until the global context is cancelled, so the
  loop cannot succeed; the deferred `m.Stop()` only cancels the context and
  does not wait for the run goroutine, whose `defer w.Freeze(...)` is still
  writing when `t.TempDir()` cleanup starts. Windows refuses to unlink an open
  file, so cleanup fails. Linux tolerates the unlink.
- Fix: none in this slice — slice 9's scope guard forbids product code.
  Proposed one-line test fix for a later slice: after `m.Stop()`, poll
  `m.IsRunning()` before returning (or give `Matrix` a done channel that
  `Stop` waits on).
- Proof: 1 failure in 4 full-set runs on 2026-10-28 (clone feature-push
  proof, `go test exit 1`, 12.2 s); 5 targeted re-runs of
  `go test ./pkg/campagne -run TestSkipJournalsViaHook -count=5` → `ok`,
  17.4 s.
- Lesson: a gate is only as honest as its flakes are loud; a fire-and-forget
  `Stop` plus `t.TempDir()` is a Windows race, not a product failure.
- Update 2026-10-28: fixed in the test only — `TestSkipJournalsViaHook` now
  waits after `m.Stop()` for `IsRunning()` to clear before `t.TempDir()`
  cleanup. Proof: `go test ./pkg/campagne -run TestSkipJournalsViaHook
  -count=20` → `ok`, 64.6 s; `go test -count=1 ./...` → exit 0,
  `pkg/campagne` `ok` 10.9 s; `python tools/verify.py --only tests` → `PASS`,
  75.0 s. No product code changed.
- Update 2026-10-28 (2): the wait moved into a shared test helper `stopWait`
  (`pkg/campagne/campagne_test.go`), used by `TestSkipJournalsViaHook` and
  `TestFailureIsNotASkip`, and it now reports a bound expiry instead of
  swallowing it. The skip test waits on the journal channel, not on a
  matrix-finished loop that could never end (cells 2..9 block in `Small` too),
  which removes ~3 s of dead wait: fresh `pkg/campagne` 8.1 s, down from
  10.9-11.1 s. Test-only.
- Update 2026-10-29 (recurrence, different test): the full `tests` gate failed
  once in `TestMatrixFilteredSubmatrix` with the same Windows cleanup shape:
  `TempDir RemoveAll cleanup: unlinkat ...\run-1793221204: Le répertoire n'est
  pas vide`. This is the E5 family, not a product regression; slice 8 changed
  no Go file. Proof: targeted
  `go test ./pkg/campagne -run TestMatrixFilteredSubmatrix -count=1` → `PASS`,
  4.2 s; `python tools/verify.py --only tests` → `PASS`, 29.8 s; full
  `python tools/verify.py` re-run → `6 passed, 0 failed, 0 not run`. The
  slice-9 fix covered the skip test, not every matrix test that starts a run;
  the general fix (a done channel that `Stop` waits on) remains open.
