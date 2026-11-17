# CGO / Meteolink — MUSTREAD Reorganization Design

*Date: 2026-11-17 · Branch: `main` (worktree `operator-cockpit`) · Status: approved design, pre-plan*
*Language: English (specs, plans, gates). User-facing product text: French.*

## 1. Goal

Turn the seven-file blueprint at `C:\END` into a canonical MUSTREAD pack that is
an execution system, not an archive: one overarching index, one document per
chunk, and a compact injection into the project's core instructions so that every
future session is governed by it.

Organizing lenses, in the operator's words: **concrete result, concrete
mechanism, concrete contribution, actions and usage**. The focus is what the
project produces, how it works, what it contributes and what an agent does next —
not a claims/evidence trustworthiness regime.

### Non-goals

- No code, tests, config or product behaviour changes in this design's own
  implementation phase beyond the new docs, the compact `AGENTS.md` block and
  the one-line pointer in `docs/superpowers/plans/README.md`.
- No deletion, no phasing out, no summarization that loses a task, phase,
  command, commit message or expected output from the source corpus. The
  corpus's own deletion directives are preserved as tasks; they are not
  executed by this design's implementation.
- No edits to the original `C:\END` files. They stay frozen as the source.
- No edits to the global operator constitution.

## 2. Source corpus

| Source file | Lines | Dominant lens |
|---|---:|---|
| `CGO.md` | 1–237 | Mechanism — constitution, execution loop, debug playbook, reviewer protocol |
| `CGO0.md` | 1–286 | Result — zero-bloat target system and the 7-chunk route |
| `CGO1.md` | 1–237 | Contribution — problem, personas, value proposition, architecture brief |
| `CGO2.md` | 1–133 | Actions — Chunk 0 (secret purge, bloat, SSoT) |
| `CGO3.md` | 1–91 | Actions — Chunk 1 (security lockdown) |
| `CGO4.md` | 1–97 | Actions — Chunk 2 (data honesty, integrity, physics) |
| `CGO5.md` | 1–410 | Actions — Chunks 3, 4, 5, 6 and the definition of done |

Total: 1,491 lines. All of them remain present in the pack.

## 3. Locked decisions

1. **MUSTREAD wins on conflict.** Existing rules are retained verbatim as
   subordinate operating detail; they are not deleted or rewritten away.
2. **Nothing is removed.** Every task, phase, human gate, agent prompt, command,
   commit message and verification step survives, with its source line range.
3. **One document per chunk.** The four chunks embedded in `CGO5.md` are split
   into four separate documents.
4. **Uniform document template** across all seven chunk docs, so each can be
   dispatched to SDD independently.
5. **Two-stage handoff:** first a plan that builds this pack and injects CGO into
   the core instructions; then one implementation plan per chunk, executed by
   subagent-driven-development in dependency order 0→6, one subagent at a time
   with a review after each.
6. **Index is the recall surface.** It carries the overarching approach, the
   five-lens map, the source crosswalk, canonical constants and the SDD ledger.

## 4. Pack structure

`docs/MUSTREAD/` contains `INDEX.md` plus ten documents.

### Foundation (3)

| Doc | Source | Purpose |
|---|---|---|
| `01-core-directive.md` | `CGO.md` 1–237 | Mechanism: seven iron laws, 5-step execution loop, 3-tier speed loop, 5-layer root-cause playbook, 30-second reviewer protocol |
| `02-target-system.md` | `CGO0.md` 1–286 | Result: delete/keep lists, zero-bloat target architecture, Chunk 0→6 overview |
| `03-product-contribution.md` | `CGO1.md` 1–237 | Contribution: personas, operational problem, 3-sentence value proposition, explicit product boundaries, architecture brief, chunk acceptance table |

### Execution (7 — one per chunk)

| Doc | Source ranges | Chunk result |
|---|---|---|
| `10-chunk-0-secret-purge-bloat.md` | `CGO2.md` 1–133 | Cap-40 purge, bloat amputation, SSoT alignment, 127.0.0.1 default |
| `11-chunk-1-security-lockdown.md` | `CGO3.md` 1–91 | Owner AuthZ, dial-time SSRF, `?iface=` validation, timeouts, 60 s default |
| `12-chunk-2-data-honesty-integrity.md` | `CGO4.md` 1–97 | Absent-is-never-zero, real integrity, fail-closed topology, atomic writes |
| `13-chunk-3-realtime-pipeline.md` | `CGO5.md` 1–89 | eBPF ringbuf → 20 Hz SSE → Float64Array → RingBufferHUD, 2-mode polish |
| `14-chunk-4-honest-tests.md` | `CGO5.md` 90–167 | Real synctest, node/happy-dom split, unmocked Playwright + Axe |
| `15-chunk-5-container-ci-release.md` | `CGO5.md` 168–267 | 3-stage Docker, `--strict` CI, pinned signed release, NTP/log/DR ops |
| `16-chunk-6-live-proof-freeze.md` | `CGO5.md` 268–410 | Linux proof, field trace, `.dist`, CHR, thesis, v2.0.0, DoD checklist |

The `CGO5.md` partition sums to 89 + 78 + 100 + 143 = 410 lines. Transition
paragraphs are assigned whole to the chunk they close, so no sentence is cut in
half and no line is double-counted.

### Uniform chunk-doc template

Every chunk doc begins with a short recall header (`Goal`, `Mechanism`,
`Contribution`, `Usage`, `Source`), then:

1. **Part A — Human gate** (credential revocation, force-push, live runs,
   approvals), labelled with its owner.
2. **Part B — Agent execution steps** with exact file paths, exact commands,
   forbidden list and the named commit message.
3. **Part C — Verification gate** with commands and expected outputs.
4. **What it unlocks** — the dependency link to the next chunk.

Original prose is kept verbatim under retitled headings. Where the same
constraint appears in several chunks, the text stays in each place and gains a
one-line pointer such as "inherited from `01-core-directive.md` §2" — smart
navigation without deletion.

## 5. `INDEX.md` contents

1. **Overarching approach** — one paragraph: what METEOLINK is and the 7-chunk
   route from this corpus.
2. **Five-lens map** — result / mechanism / contribution / actions / usage →
   which document answers each.
3. **Document map** — filename, one-line purpose, source line range.
4. **Persistent-recall block** — locked constants (`2.0.0`, 60 s = 10/40/10,
   220 ms, `127.0.0.1`, canonical title), the precedence rule, and a path-binding
   note mapping the corpus's `pkg/...` paths to this repo's real top-level paths
   (`qdisc/`, `results/`, `remediate/`, `probe/`, `metrics/`, …).
5. **Completeness crosswalk** — one row per source line range → destination
   section.
6. **Execution status ledger** — task, doc, owner, state — the ledger required
   by SDD, kept here so no extra file is needed.
7. **Handoff** — which doc feeds which implementation plan.

## 6. Core-instruction injection

Only the worktree `AGENTS.md` is edited, with a compact block immediately under
the title:

```markdown
## CGO Core Directive (MUSTREAD)

`docs/MUSTREAD/INDEX.md` is the canonical source of direction for this project.
Where this card conflicts with the MUSTREAD pack, MUSTREAD governs; the sections
below are retained as subordinate operating detail. Operator-level constitution
bans cannot be waived by a project document.
```

A short **Precedence** subsection under that block names each concrete conflict
and its resolution:

| Conflict | Resolution |
|---|---|
| Corpus forbids `CHECKPOINTS.md`/`VERIFY.md` edits outside the final chunk | Wins; the existing checkpoint cadence becomes subordinate, files stay untouched by chunk work |
| Chunk 6 uses `git add -A` | Cannot be waived by a project document; the commit task stays, staging becomes explicit paths |
| Chunk 0 deletes `cmd/gates` | Directive stands, but a replacement for the six-gate runner must exist before removal — no gate tampering |
| Canonical title `Optimisation temps réel…`, 2 modes, version `2.0.0` | Locked; present repo values are labelled historical, not deleted |
| Corpus `pkg/...` paths | Kept verbatim; the INDEX path-binding note gives this repo's real paths |

All pre-existing `AGENTS.md` sections stay verbatim underneath. One line is
added to `docs/superpowers/plans/README.md` pointing to `docs/MUSTREAD/INDEX.md`.

## 7. Completeness mechanism

During implementation, a one-off script checks that the crosswalk partitions
all 1,491 source lines exactly once: for each source file, the sorted ranges are
contiguous from line 1 to its last line, with no gap and no overlap. The check
is run once during the build; it is not installed as a permanent gate. No
claims-verification regime is added.

## 8. Execution handoff

- `writing-plans` first produces the plan that builds `docs/MUSTREAD/` and
  injects the core directive.
- It then produces one implementation plan per chunk. Mapping: Part A → a
  decision task; Part B → tasks with Files / Interfaces / `- [ ]` steps; Part C →
  expected outputs.
- SDD executes chunks 0→6 in dependency order, fresh implementer subagent per
  task, task review after each, one subagent at a time, ledger rows in
  `INDEX.md`.

## 9. Acceptance criteria

1. `docs/MUSTREAD/` has `INDEX.md` and exactly ten numbered docs.
2. The one-off crosswalk check reports all seven source files fully covered,
   zero gaps, zero overlaps.
3. No task, phase, command, commit message or expected output from `C:\END` is
   absent from the pack; `C:\END` itself is byte-unchanged.
4. `AGENTS.md` contains the core-directive block and precedence subsection,
   with every prior section intact below it.
5. The first implementation plan is written by `writing-plans` only after the
   operator has reviewed and approved this spec.

## 10. Retained risks

The corpus contains internal contradictions (title, 2 vs 4 UI modes, `cmd/gates`
delete vs. use, stale `pkg/...` paths, `2.0.0` vs `1.3.5`). Per the locked
decisions they are preserved, made navigable by the INDEX notes and the
precedence table, and resolved chunk-by-chunk at plan time — never by silently
deleting text.
