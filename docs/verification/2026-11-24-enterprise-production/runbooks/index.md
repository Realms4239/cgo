# Runbook index — live operations, tokens, rollback, last rehearsal

Authority: design §5 (`docs/superpowers/specs/2026-11-24-enterprise-production-modernisation-design.md`)
— every destructive/live operation keeps a written rollback and evidence
sidecar; this index names each operation, its token, rollback and last
rehearsal date. Builder/narrator pass, 2026-11-24. No gate, CI, ledger,
permit, corpus, book, thesis or protected-config edit.

## Definitions (what counts as a gap)

- **Live operation** = an operation with a named `APPROVED-YYYY-MM-DD-*.md`
  token in this repository's own documents (`docs/**` and `.superpowers/**`).
  Tokens are indexed one row each. A token that exists only as a proposal or a
  draft is still indexed, and marked as such — not dropped.
- **Rollback** = the concrete reversal line quoted in a repo document, with its
  evidence stamp where the document gives one. `ABSENT` means no rollback line
  is quoted in the repo docs; it is a reported gap, not a claim that no rollback
  exists on the host.
- **Last rehearsal** = the most recent rehearsal date actually recorded in a
  repo document. `NOT REHEARSED` means no rehearsal date is recorded there. A
  rehearsal on a *different* operation (e.g. the 2026-11-12 portable-bench
  rehearsal VM) is not borrowed to make this column green.
- `N/A — read-only R0` means the operation performs no mutation, so there is no
  rollback to write; the reason is stated, never left blank.
- No token id, rollback proof or rehearsal date is invented. Where a document
  names a token only by suffix (e.g. `-late` in `2026-11-16-exhaustive/CONSOLIDATION.md`),
  the suffix is quoted as prose under the full token row; it is not expanded into
  a new token id.

## Operation index

| Token (APPROVED-*) | Operation | Scope | Rollback | Last rehearsal | Evidence / status |
|---|---|---|---|---|---|
| APPROVED-2026-11-07-ssh-key.md | SSH console key install | Guest `~/.ssh/authorized_keys` on the lab/bench guest | ABSENT — no rollback line quoted in repo docs (`docs/runbooks/ssh-key-install.md`) | NOT REHEARSED — no date recorded | Referenced by discovery write-window `20261111-102129`; runbook `ssh-key-install.md` |
| APPROVED-2026-11-11-labo-cgo-bench.md | Bench VM provisioning (KVM, `/home/sadc/labo-cgo`) | sadcnwp guest + loopback hostfwd 10022 | VM stop / pre-BBRv3 snapshot `labo-cgo-bench-6.8-pre-bbrv3` remains available (2026-11-12 phase2 doc) | 2026-11-11 — rehearsal named in the discovery evidence set (`docs/CHECKPOINTS.md`); no distinct date | `2026-11-11-sadcnwp-bench-discovery/`; `2026-11-14-sadcnwp-r0/README.md` |
| APPROVED-2026-11-11-qemu-kvm-disable.md | qemu/kvm module disable action | Host kernel module state | ABSENT — no rollback line quoted in repo docs | 2026-11-11 — named in the same discovery evidence set; no distinct date | Discovery write-window `20261111-101800`; `docs/CHECKPOINTS.md` |
| APPROVED-2026-11-15-bench-vm-lane-a-DRAFT.md | Lane A bench VM bring-up (draft) | sadcnwp guest bring-up | ABSENT — draft, no rollback line | NOT REHEARSED — draft only | DRAFT token; superseded by `APPROVED-2026-11-15-bench-vm-lane-a.md` |
| APPROVED-2026-11-15-bench-vm-lane-a.md | Lane A bench VM bring-up | sadcnwp guest, cloud-init seed | ABSENT — no rollback command quoted; log records seed rebuild + fresh overlay at 2026-11-15T19:55Z | NOT REHEARSED — boot parity checked against rehearsal VM 6.8.0-149, not this operation | `2026-11-15-bench-vm/README.md` |
| APPROVED-2026-11-16-chr-boot.md | RouterOS CHR first boot | `/home/sadc/labo-cgo/chr` | ABSENT — no rollback line quoted for this token; later boot used pristine image re-deposit | NOT REHEARSED — no date recorded | `2026-11-16-chr-boot/README.md`; `2026-11-19-chr-blocked/README.md` |
| APPROVED-2026-11-16-exhaustive-bench.md | Exhaustive live-bench campaign (single writer) | sadcnwp bench VM | Topology restore; no concurrent campaign (CONSOLIDATION.md) | NOT REHEARSED — no date recorded | `2026-11-16-exhaustive/{PLAN,CONSOLIDATION}.md`; companion suffixes `-exhaustive-continuation`, `-late` are quoted, not expanded |
| APPROVED-2026-11-16-v2-matrix.md | V2 24-cell corpus run | sadcnwp bench VM, own hash anchor | Write file carries rollback pointer to `labo-cgo-bench-create.sh` | NOT REHEARSED — no date recorded | `2026-11-16-v2-corpus/README.md`; `docs/VERIFY.md` V2 row |
| APPROVED-2026-11-17-chr-real-audit.md | CHR real-audit proof | `/home/sadc/labo-cgo/chr` | Pinned image re-deposited (`c6abf876…`) | NOT REHEARSED — no date recorded | `2026-11-17-chr-real-audit/README.md`; `2026-11-17-api-bootstrap/README.md` |
| APPROVED-2026-11-17-exhaustive-continuation.md | Campaign continuation (renewal, exp 06:00 UTC) | sadcnwp bench VM, no new privileges | Same single-writer campaign reversal as `APPROVED-2026-11-16-exhaustive-bench.md` | NOT REHEARSED — no date recorded | `docs/CHECKPOINTS.md`; `2026-11-16-exhaustive/CONSOLIDATION.md` |
| APPROVED-2026-11-19-bench-disk-expand.md | qcow2 +5 GB expansion | Guest overlay `labo-cgo-bench.qcow2` | Overlay backed up to `labo-cgo-bench.qcow2.pre-expand-20261119`; VM stopped `kill -TERM 3673912`; relaunched with recorded command | 2026-11-19 — rehearsal resize on a copy verified 8.5 GiB virtual size | `2026-11-19-multiprofile-trace/README.md` §Disk work; evidence `20261119-200907-*` to `20261119-201154-*` |
| APPROVED-2026-11-19-chr-boot-r2.md | CHR boot R2 (verified), exp 23:59 UTC 2026-11-19 | `/home/sadc/labo-cgo/chr` | `pkill -f cgo-chr-v13` | NOT REHEARSED — no date recorded | `docs/VERIFY.md`; `2026-11-19-chr-verified/README.md`; `2026-11-19-chr-verifier-replay/README.md` |
| APPROVED-2026-11-19-field-trace-read.md | Field trace R0 read | Passive trace files, read-only | N/A — read-only R0, no mutation | NOT REHEARSED — no date recorded | `2026-11-19-field-trace/README.md` |
| APPROVED-2026-11-19-field-trace-readback.md | Field trace R2 re-open / read / restore | Trace file ownership | Ownership restored; rollback named in write-window evidence `20261119-175127` | NOT REHEARSED — no date recorded | `2026-11-19-field-trace/README.md`; `2026-11-19-field-trace-p3/README.md` |
| APPROVED-2026-11-19-field-trace-export-read.md | Field trace export R0 read | Export files, read-only | N/A — read-only R0, no mutation | NOT REHEARSED — no date recorded | `2026-11-19-field-trace/README.md` |
| APPROVED-2026-11-22-exhaustive-bench.md | Exhaustive bench, 2026-11-22 | sadcnwp bench VM | Topology restore; evidence sidecars in `C:\WEBSERVER\evidence\` | NOT REHEARSED — no date recorded | `2026-11-22-exhaustive/README.md` |
| APPROVED-2026-11-23-field-trace-replay.md | Field-derived P3 `.dist` A/B replay | sadcnwp bench VM, P3/cake/cubic/down | Topology restored, evidence `5e17a0d1…` | NOT REHEARSED — no date recorded | `2026-11-23-field-trace-replay/README.md` |
| APPROVED-2026-11-23-field-replay-extended.md | Extended A/B, 3 reps per arm | sadcnwp bench VM, P3 | Topology restored, evidence `39bc077a…` | NOT REHEARSED — no date recorded | `2026-11-23-field-trace-replay/README.md` |
| APPROVED-2026-11-23-more-profiles-replay.md | Bench-derived P1/P2/P4 `.dist` replay | sadcnwp bench VM, P1/P2/P4 | Topology restored, evidence `b75bbf76…` | NOT REHEARSED — no date recorded | `2026-11-23-field-trace-replay/README.md` |
| APPROVED-2026-11-23-live-physics-validation.md | Live physics validation (correlation, serialisation) | sadcnwp bench VM, netem vectors | cgo-netctl permit restored byte-exact `b28092d2…` (`PERMIT_RESTORED`, evidence `20261123-110912`); topology restored `20261123-110941` | NOT REHEARSED — no date recorded | `2026-11-23-live-physics/README.md` |
| APPROVED-2026-11-23-netctl-permit.md | Narrow cgo-netctl permit change | `/usr/local/bin/cgo-netctl`, one branch | Original permit restored byte-exact `b28092d2…`, `PERMIT_RESTORED` evidence `20261123-110912` | NOT REHEARSED — no date recorded | `2026-11-23-live-physics/README.md` |
| APPROVED-2026-11-23-organic-capture.md | Organic 3 h capture (first plan) | sadcnwp host, eBPF observatory | ABSENT — proposal only, no rollback written | NOT REHEARSED — proposal only | `docs/superpowers/plans/2026-11-23-organic-trace-model-system-plan.md` ("Create …"); no execution evidence in repo docs |
| APPROVED-2026-11-23-organic-ab.md | Organic live A/B replay | sadcnwp bench VM | ABSENT — proposal only; plan task BLOCKED, not executed | NOT REHEARSED — proposal only | `docs/superpowers/plans/2026-11-23-organic-trace-model-system-plan.md`; `2026-11-23-organic-trace/README.md` |
| APPROVED-2026-11-23-organic-capture-2.md | Organic 3 h eBPF capture (executed) | sadcnwp host, 109,550 rows | One narrow sudoers addition + three readback commands rolled back; ownership rollback evidence `20261124-060634` / `3aa607cd…` | NOT REHEARSED — no date recorded | `2026-11-23-organic-trace/README.md` |
| APPROVED-2026-11-24-huge-model-sweep.md | Four-profile model sweep (dist vs scalar) | sadcnwp bench VM, P1–P4 | Topology restored after both arms, evidence `4a0a43ed…` | NOT REHEARSED — no date recorded | `2026-11-24-model-sweep/README.md` |
| APPROVED-2026-11-24-ingress-up.md | Ingress shaping / up-direction validation | sadcnwp bench VM, IFB + cake ingress | Topology restored to egress at end | NOT REHEARSED — no date recorded | `2026-11-24-ingress-up/README.md`; three controller amendments recorded there |
| APPROVED-2026-11-24-organic-retrieval.md | Organic trace retrieval | sadcnwp host, bounded extracts | Ownership rollback evidence `20261124-060634` / `3aa607cd…` | NOT REHEARSED — no date recorded | `2026-11-23-organic-trace/README.md` |
| APPROVED-2026-11-24-stack-debug.md | Systematic stack debugging / E2E check | sadcnwp bench VM, detached audit + cells | Topology restored to egress, evidence `2f298e3f…`, `49796fc8…` | NOT REHEARSED — no date recorded | `2026-11-24-stack-debug/README.md` |

## Gap register

Gaps are listed, never silently omitted. Counts are from the table above.

- **Rollback `ABSENT` (7):** `APPROVED-2026-11-07-ssh-key.md`,
  `APPROVED-2026-11-11-qemu-kvm-disable.md`,
  `APPROVED-2026-11-15-bench-vm-lane-a-DRAFT.md`,
  `APPROVED-2026-11-15-bench-vm-lane-a.md`,
  `APPROVED-2026-11-16-chr-boot.md`,
  `APPROVED-2026-11-23-organic-capture.md`,
  `APPROVED-2026-11-23-organic-ab.md`.
  (For the last two the reason is stronger: they are proposals, not executed
  operations.) The remaining rows quote a rollback line or an explicit
  `N/A — read-only R0`.
- **Rehearsal `NOT REHEARSED` (25):** every row except
  `APPROVED-2026-11-11-labo-cgo-bench.md`,
  `APPROVED-2026-11-11-qemu-kvm-disable.md` (both 2026-11-11, named in the
  discovery evidence set, no distinct date) and
  `APPROVED-2026-11-19-bench-disk-expand.md` (2026-11-19, rehearsal resize on a
  copy). No rehearsal date was inferred from a different operation.
- **Draft / proposal, not executed (3):**
  `APPROVED-2026-11-15-bench-vm-lane-a-DRAFT.md`,
  `APPROVED-2026-11-23-organic-capture.md`,
  `APPROVED-2026-11-23-organic-ab.md`.
- **Read-only R0 (3):** the two `field-trace-read*` tokens and the export read;
  rollback is `N/A` by mutation scope, not a blank cell.
- **Suffix-only references (not expanded into token ids):** `-exhaustive-continuation`
  and `-late` in `2026-11-16-exhaustive/CONSOLIDATION.md`. The full
  `APPROVED-2026-11-17-exhaustive-continuation.md` is indexed separately because
  it appears verbatim in `docs/CHECKPOINTS.md`; `-late` has no verbatim full
  token in the repo docs and is therefore not invented here.

## Coverage oracle

The index is checked by the doc test embedded in `../README.md` (§ "Doc test").
It re-discovers every `APPROVED-*` token from the repository's own documents,
asserts one index row per token, asserts no empty cell, and asserts that the
gap markers above are present. Hostile proof: remove one row or blank one cell
in a copy and the same test exits 1.
