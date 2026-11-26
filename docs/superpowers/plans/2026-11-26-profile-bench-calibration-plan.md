# Plan — targeted field capture and calibration for P1, P2, P3 (2026-11-26)

Operator request 2026-11-26: "do another bench run which explicitly point to all of our other profiles and calibrate them, make it an SDD". Scope: make a real field trace for the profiles that are still below the 30-sample floor, then seal a calibration for each qualifying profile. P4 is already calibrated (`docs/verification/2026-11-26-field-trace-out3/calibrations/P4.dist`) and is out of scope here.

## Global Constraints

- No synthetic sample, no lowered floor, no fabricated value. A profile with fewer than 30 real SRTT samples stays BLOCKED with a named condition.
- Every host interaction goes through `C:\WEBSERVER` guarded tools under a task-scoped token with all four artifacts (token, named quiet window, exact rollback, same-day rehearsal). No ad-hoc SSH, no guard change without explicit operator approval, no retained sudoers change, no password in a document/script/shell history (private card only).
- No `.dist` is applied in a live bench cell; no frozen corpus, gate, ledger, `book/data/runs` or CSV schema change; no branch deletion, merge, or force push.
- SDD: one implementer per task, task review after each, final whole-branch review before push. French comments/user text, English identifiers, explicit-path commits.

## Task 1 — find reachable endpoints in the three RTT bands (read-only)

Using only `safe_read.py` on sadcnwp, identify TCP endpoints whose observed SRTT lands in P1 (20-40 ms, excluding P4's 40-60 band), P2 (100-200 ms) and P3 (>=600 ms). Record exact read commands, evidence hashes and observed values. Do not write to the host, do not change the guard, do not start a capture. Deliver a table of candidate endpoints per band with confidence and the named risks (e.g. an endpoint that is only reachable over TLS or is DNS-dependent).

## Task 2 — scoped capture (controller-owned)

Author the task-scoped R2 token and exact command file for a bounded passive capture whose traffic is directed at the Task 1 endpoints. Rehearse (selftest, exact classify, near-variant refusal), launch, restore any temporary privilege immediately, and stop by exact PID. No `.dist`, no gate change.

## Task 3 — band and calibrate

For each profile with >=30 real samples: run `cgo tracemodel` and `cgo calibrate --profile <P>`, verify the `.dist` SHA-256 against its sealed sidecar, and commit the evidence under `docs/verification/<date>-field-trace-<profile>/`. Profiles below the floor stay BLOCKED with the observed count and the named unblock condition. No synthetic rows.

## Task 4 — verification and report

Run the six gates, frontend typecheck/tests, and a frozen-corpus audit; report exact commands, exits and counts. Final whole-branch review before push. State plainly which profiles were calibrated from real field traces and which remain blocked.
