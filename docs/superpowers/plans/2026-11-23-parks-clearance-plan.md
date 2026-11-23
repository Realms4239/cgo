# Non-blocking parks clearance plan (2026-11-23)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the three locally-fixable parks left by the realism/eBPF/interface programme; leave the live kernel-truth park human-gated with a precise reason.

**Architecture:** Three independent one-to-two-file slices: (1) reserve L1/V1 in `profile.Reserved()`, (2) delete two dead CSS selectors and give `.c-pip.offline` a distinct honest dot, (3) final review. No new dependency, accent, motion or component.

**Tech Stack:** Go (`profile`), CSS (`components.css`), Vitest.

## Global Constraints

- P1-P4 byte-identical; L1/V1 numeric values and source citations unchanged; no frozen corpus, book, C:\END, ledger or gate edit.
- CSS: existing tokens only; no new hex/font-size/`transition: all`/`@keyframes`; forced-colors and reduced-motion preserved; `.et-mark` stays (used by Observer event timeline).
- `.c-pip.offline` must be visually distinct from `idle` without implying green — use an existing warning/danger or faint-with-shape token, never a success colour.
- No `git add -A`; explicit paths only; no merge; native gates once; pre-existing `links` FAIL and `book/` SKIPs recorded, never edited.

---

### Task 1: Reserve L1/V1 against CSV import overwrite

**Files:** Modify `profile/profile.go`, `profile/reserved_test.go`.

- [ ] **Step 1:** Add a failing Go test in `profile/reserved_test.go`: `Reserved("L1") === true`, `Reserved("V1") === true`, and `Import(model.Profile{ID:"L1",...})` returns an error naming "réservé" without writing. Run `go -C profile test -count=1 ./...` -> RED.
- [ ] **Step 2:** In `profile/profile.go:27-29`, extend `Reserved` to include `L1` and `V1` (keep P1-P4). Update the error message to reflect the wider reserved set honestly. Run tests -> GREEN.
- [ ] **Step 3:** Verify `go -C api test -count=1 ./...` and `go build ./...`; gates once. Commit explicit paths with `fix(profile): reserve cited L1 and V1 profiles`.

---

### Task 2: CSS park cleanup

**Files:** Modify `web/frontend/src/styles/components.css`.

- [ ] **Step 1:** DOM/style oracle (or existing style test) that fails if `.tf-status` or `.labo-matrix-word` selectors remain, and that `.c-pip.offline .c-pip-dot` has a distinct visual rule. Watch RED.
- [ ] **Step 2:** Delete the two dead selector rules (`.tf-status` at ~235, `.labo-matrix-word` at ~425-427). Add `.c-pip.offline .c-pip-dot` using an existing non-success token with a distinct shape/border (e.g. `background: var(--surface-elevated); border: 1px solid var(--text-faint)`) so it is visually different from `idle` and never green. Keep `.et-mark` untouched. Watch GREEN.
- [ ] **Step 3:** Full vitest at 2 workers, build, unmocked Playwright green-bar, screenshots at 890/1280/1920 read and validated, gates once. Commit explicit paths with `fix(ui): remove dead selectors and distinguish offline pip`.

---

### Task 3: Final review and kernel-truth park statement

- [ ] Run a final whole-diff reviewer over the two commits; one fix wave + scoped re-review if Critical/Important findings appear.
- [ ] Append to the SDD ledger: the live eBPF-vs-tc confrontation stays PARKED because `probe/bpf/bpf_probe.o` is not in this checkout and the operator card bans runtime clang/gcc/llvm on the VM; unblock requires a human-gated object rebuild on a build host or a verifier-supplied real sealed report. No hand-written seal, no third live window.

## Self-Review

- Spec coverage: park 1 -> Task 1; parks 2-3 -> Task 2; park 4 -> Task 3 statement.
- No placeholders: commands, files and exact conditions given.
- Type consistency: `Reserved`, `.c-pip.offline`, `.et-mark` used as defined in source.
