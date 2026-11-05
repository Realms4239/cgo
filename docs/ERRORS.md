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
- Update 2026-10-29 (fix): `TestMatrixFilteredSubmatrix` and the new
  `TestExplicitBBRv3AcceptedWhenPresent` now `defer stopWait(t, m)`, so the
  matrix is stopped and drained before `t.TempDir()` cleanup. Stress:
  `go test ./pkg/campagne -run
  'TestMatrixFilteredSubmatrix|TestExplicitBBRv3AcceptedWhenPresent' -count=10`
  → `ok`, 11.0 s. Full suite green. The generic done-channel remains a design
  improvement, not a current failure.

## E6. Host blocks IPv4 ICMP loopback (observed 2026-10-29)

- Symptom: `go test ./pkg/probe -run TestPingLiveLoopback` fails with “live ping
  returned no samples”; a manual `ping 127.0.0.1` loses 100 %.
- Diagnosis: `ping -4 8.8.8.8` and `ping -6 ::1` both answer; only IPv4 loopback
  is filtered on this Windows host (public firewall profile enabled, rule
  inspection denied to the current user). The product's external IPv4 path is
  not implicated.
- Fix: the smoke test uses `::1` and keeps the real `ping` binary and parser in
  the path; the comment records the host quirk. No product code changed.
- Lesson: a live integration test must name the environment it depends on; a
  blocked loopback is not a parser regression. Do not “fix” it by fabricating
  samples.

## E7. Full C: disk and a cold Go cache made the exit gate time out (observed 2026-10-29)

- Symptom: `python tools/verify.py` → `tests` failed with `link.exe: resize
  output file failed: Espace insuffisant sur le disque`, then the `exit` gate's
  `go build ./cmd/cgo` timed out after 300 s.
- Cause: C: was at 99 % (1.2 GB free) and `go clean -cache` had emptied the
  build cache; the first link plus the test binaries did not fit in the wall.
- Fix: removed only regenerable caches and our own temp artifacts (`go clean
  -cache`, `/tmp/cgo-ui`, old screenshots), then warmed `go build ./cmd/cgo`
  before rerunning. Full verify → `6 passed, 0 failed, 0 not run`. No product
  code changed.
- Lesson: run the disk check before a full gate on this host; a cold cache needs
  one warm build and ~2.5 GB free.

## E8. CGO-DEV boots but no recorded SSH identity is accepted (observed 2026-10-29)

- Symptom: after `VBoxManage startvm CGO-DEV --type headless`, port 2222 opens
  but every recorded user/key pair returns `Permission denied
  (publickey,password)`; host SSH to PC2 itself is fine.
- Tried: users `altfloat`, `cgo`, `ubuntu`, `meteolink`, `airis`, `root`; keys
  `id_ed25519`, `nemo_lab`, `nemo_ed25519`. PC2’s VirtualBox NAT forwards
  `0.0.0.0:2222 → guest:22`; the guest IP property is `10.0.2.15`.
- State: the VM was powered off again (`VMState="poweroff"`); no guest command
  ran, no file was changed, no Graylog component was touched.
- Unblock: install a known public key through the VirtualBox console (or record
  the `altfloat` password), then `cgo kit keysetup`; or provide a stable Linux
  6.8 host. A 6.18 WSL is not a substitute for the 6.8 acceptance.

## E9. The API gate made bbr3 unreachable over HTTP (observed 2026-10-31)

- Symptom: the engine accepts `bbr3` only when `/proc` exposes it and refuses
  otherwise with the `install-bbr3` path, but `/api/run/start` rejected
  `ccs:["bbr3"]` as `cc inconnu` even on a kernel that exposes it; the bench UI
  also hard-coded `['cubic','bbr']`.
- Cause: the handler validated the matrix axes against the settings enums
  (`shape_qdisc`, `burst_cc`) instead of the model lists the engine uses; the
  UI duplicated those lists. Three copies drifted. The qdisc enum also
  accepted `none`, which the engine rejects.
- Fix: `campagne.CapabilitiesNow()` + `GET /api/capabilities` (qdiscs, runnable
  CCs, known CCs, kernel release, bbr3); `/api/run/start` validates
  `model.AllQdiscs` / `model.KnownCC`, leaving availability to the engine's
  fail-closed gate; the frontend reads the payload once (`loadCapabilities`) and
  renders it in Configurer and the matrix preview; results recognize `bbr3`.
- Proof: `go test ./pkg/campagne ./pkg/api ./pkg/model` ok; frontend 52 files /
  231 tests; typecheck; build 351.0 KB gz; on 9092 `/api/capabilities` returned
  `ccs:["cubic","bbr"]`, `known_ccs` with `bbr3`, `bbr3:false`; the captured
  Configurer names the missing kernel and the installer. No 7.1 kernel booted.
- Lesson: a whitelist copied from a settings enum is not a capability; one
  payload from the host is the seam.

## E10. A passing alert probe could have produced a green verdict (observed 2026-10-31)

- Symptom: with the bulk sink unreachable, `capSamples` still collects load-window
  completions, so `cap_sla` can read 100 % while no load was ever established.
- Cause: a verdict derived from the alert percentage alone would treat that as a
  pass. `load_measured` is the fact that says whether congestion existed.
- Fix: `metrics.AssessVerdict` gates on `LoadAttempted` and `LoadEstablished`
  before any green, and the report freezes `load_attempted`.
- Proof: `TestVerdictNeverGreenWithoutGrade`, `TestReportVerdictRefusesGreenWithoutLoad`.
- Lesson: a metric measured in the absence of its own stress is not a pass.

## E11. A client-supplied `run_id` could write outside `data/runs` (observed 2026-11-01)

- Symptom: `POST /api/run/start` accepted `run_id: "../evade"`, forwarded it to
  the engine, and `OpenRun` created a run directory above `data/runs`, writing
  `aqm_eval.csv` there. The same string later reaches `storedTier` as a path.
- Cause: the handler validated every field except `run_id`; every other route
  that reads a run already used `validRunID` (no `/`, `\`, `.`).
- Fix: the handler refuses a non-empty `run_id` that fails `validRunID` with 400
  before the engine gate; `TestRunStartRefusesEscapingRunID` covers `../`,
  `..\`, `a/b`, `run.1` and `.`, and asserts the engine is never called.
- Proof: `go test ./pkg/api -run "RunStartRefusesEscaping|RunStartSubmatrix|RunOptsReach|RunStartProvenance" -v` → PASS.
- Lesson: an identifier that becomes a filesystem path is validated at the
  boundary, not at the third consumer that happens to notice.

## E12. A resumed campaign rewrote its cost basis (observed 2026-11-01)

- Symptom: stop a run, change the price tier, resume it with the same `run_id`:
  new rows were valued at the new tier while rows already written kept the old
  one, and `manifest.json` was rewritten with the new config.
- Cause: `OpenRun` rebuilds resume state from the CSV only, and `startMatrix`
  built `matrixConfig` from `metrics.ActiveTierName()`; `Freeze` stores the
  config but nothing read it back.
- Fix: `storedTier` reads `manifest.json`'s `config.tier`, and `tierForRun`
  prefers it over the global; a fresh run still takes the current tier.
- Proof: `TestResumeKeepsTheFrozenTier` (frozen `yas-day-1gb` wins over a global
  changed to `yas-ftth-100gb`; a new run takes the global).
- Lesson: "frozen with the run" needs the resume path to read the freeze, not
  just the write path to produce it.
