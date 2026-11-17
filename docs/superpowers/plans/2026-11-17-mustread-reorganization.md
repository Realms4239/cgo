# CGO / Meteolink MUSTREAD Reorganization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `docs/MUSTREAD/` as a 10-document execution pack derived from the 1,491 lines at `C:\END`, and inject CGO as the project's core directive in `AGENTS.md`.

**Architecture:** Three foundation docs carry the mechanism, result and contribution; seven docs carry one chunk each. Every source line is copied verbatim under a recall header, so the pack can be checked mechanically for zero loss. `INDEX.md` is the single entry point and the SDD ledger. `AGENTS.md` gains a compact precedence block, not the full corpus.

**Tech Stack:** Markdown, Git Bash `sed`, Python 3.11 one-off checker in the approved temp directory. No dependency changes.

**Spec:** `docs/superpowers/specs/2026-11-17-mustread-reorganization-design.md` (commit `1940e5a`).

## Global Constraints

- Every one of the 1,491 source lines in `C:\END\{CGO.md,CGO0.md,CGO1.md,CGO2.md,CGO3.md,CGO4.md,CGO5.md}` must remain present in the pack. No deletion, no phasing out, no loss of a task, phase, command, commit message or expected output.
- Organizing lenses: concrete result, mechanism, contribution, actions, usage. Focus on what the project produces and how an agent executes it, not on claim verification.
- MUSTREAD wins where it conflicts with `AGENTS.md`; existing sections stay verbatim underneath. Operator-level constitution bans cannot be waived by a project document.
- Never use `git add -A` or `git add .`. Stage explicit paths only.
- Original `C:\END` files are never modified.
- The global operator constitution is never edited.
- English in specs/plans/gates; French remains as-is where the source uses it.
- No product code, tests, config, frozen corpus, live bench or remote host is touched by this plan.
- Run all shell steps in Git Bash (MINGW64); `/c/END/…` is the path to `C:\END\…`.

---

### Task 1: Foundation documents (mechanism, result, contribution)

**Files:**
- Create: `docs/MUSTREAD/01-core-directive.md`
- Create: `docs/MUSTREAD/02-target-system.md`
- Create: `docs/MUSTREAD/03-product-contribution.md`

**Interfaces:**
- Consumes: `C:\END\CGO.md` lines 1–237; `CGO0.md` 1–286; `CGO1.md` 1–237.
- Produces: three docs whose source ranges are consumed by Task 4's crosswalk and Task 5's checker. Exact filenames are load-bearing.

- [ ] **Step 1: Create the directory**

Run: `mkdir -p docs/MUSTREAD`
Expected: directory exists; `ls -d docs/MUSTREAD` prints the path.

- [ ] **Step 2: Write the recall header for `01-core-directive.md`**

Create the file with exactly this block as its first lines:

```markdown
# CGO Core Directive

> Source: `C:\END\CGO.md` lines 1–237. Lens: **mechanism** — how this project is governed, executed, debugged and reviewed.

## At a glance

- **Result:** one binding execution constitution for every agent session.
- **Mechanism:** seven iron laws, a 5-step execution loop, a 3-tier speed loop, a 5-layer root-cause playbook and a 30-second reviewer protocol.
- **Contribution:** removes the failure modes the corpus names (`|| true`, exit-2 masking, `SumDrops(nil)=0`, clock-only tests, broken Docker context, empty `modules/`, fake "complete" claims).
- **Usage:** read this first; then `INDEX.md`; then only the chunk doc for the active chunk.
- **Inherited constraints:** none — this is the root directive.
```

- [ ] **Step 3: Append the source verbatim**

Run: `sed -n '1,237p' /c/END/CGO.md >> docs/MUSTREAD/01-core-directive.md`
Expected: `tail -n 3 docs/MUSTREAD/01-core-directive.md` shows the corpus's final "ready to inspect the first Chunk diff" lines. Nothing edited or summarized.

- [ ] **Step 4: Repeat for `02-target-system.md`**

Create the file with this header:

```markdown
# CGO Target System

> Source: `C:\END\CGO0.md` lines 1–286. Lens: **result** — the zero-bloat target architecture and the 7-chunk route.

## At a glance

- **Result:** one Go binary, one embedded React cockpit, one eBPF probe, one verification stack, zero dead shells.
- **Mechanism:** amputation list + keep/make-sharp list + Chunk 0→6 dependency chain.
- **Contribution:** a single legible work tree instead of nginx/prometheus/module/GUI sprawl.
- **Usage:** read before any bloat-removal or architecture decision.
- **Inherited constraints:** `01-core-directive.md` §1–2 (iron laws and execution loop).
```

Then run: `sed -n '1,286p' /c/END/CGO0.md >> docs/MUSTREAD/02-target-system.md`

- [ ] **Step 5: Repeat for `03-product-contribution.md`**

Create the file with this header:

```markdown
# CGO Product Contribution

> Source: `C:\END\CGO1.md` lines 1–237. Lens: **contribution** — the problem, the personas, the value proposition and the architecture brief.

## At a glance

- **Result:** a decision-first observatory for Météo Madagascar and a defensible master's thesis for ISGEI.
- **Mechanism:** 60 s stethoscope, 4-netns digital twin, dual-sealed evidence export.
- **Contribution:** names the operational failure (bufferbloat on a constrained link) and the exact product boundaries that prevent overclaiming.
- **Usage:** read when a requirement is ambiguous — the product definition wins.
- **Inherited constraints:** `01-core-directive.md` §1–2 and `02-target-system.md` §1.
```

Then run: `sed -n '1,237p' /c/END/CGO1.md >> docs/MUSTREAD/03-product-contribution.md`

- [ ] **Step 6: Verify the three files**

Run: `wc -l docs/MUSTREAD/0{1,2,3}-*.md`
Expected: each file has at least its source line count plus its header block; no file is empty.

Run: `grep -c '^' docs/MUSTREAD/01-core-directive.md`
Expected: a number ≥ 237.

- [ ] **Step 7: Commit**

```bash
git add docs/MUSTREAD/01-core-directive.md docs/MUSTREAD/02-target-system.md docs/MUSTREAD/03-product-contribution.md
git commit -m "docs(mustread): add foundation directive, target and contribution docs"
```

---

### Task 2: Chunk 0–2 execution documents

**Files:**
- Create: `docs/MUSTREAD/10-chunk-0-secret-purge-bloat.md`
- Create: `docs/MUSTREAD/11-chunk-1-security-lockdown.md`
- Create: `docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md`

**Interfaces:**
- Consumes: `C:\END\CGO2.md` 1–133; `CGO3.md` 1–91; `CGO4.md` 1–97.
- Produces: three chunk docs with Part A / Part B / Part C headings. Task 4 crosswalks them; Task 5 checks them.

- [ ] **Step 1: Write `10-chunk-0-secret-purge-bloat.md`**

Header:

```markdown
# Chunk 0 — Secret Purge, Bloat Amputation & SSoT

> Source: `C:\END\CGO2.md` lines 1–133. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** Cap-40 PAT gone from history; bloat directories deleted; one binary, one version, one canonical bind address.
- **Mechanism:** Part A human security steps, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** a clean, legible tree before any feature work.
- **Usage:** execute first; it gates Chunks 1–6.
- **Inherited constraints:** `01-core-directive.md` §1–4.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 1 — security lockdown.
```

Then run: `sed -n '1,133p' /c/END/CGO2.md >> docs/MUSTREAD/10-chunk-0-secret-purge-bloat.md`

- [ ] **Step 2: Write `11-chunk-1-security-lockdown.md`**

Header:

```markdown
# Chunk 1 — Security Lockdown

> Source: `C:\END\CGO3.md` lines 1–91. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** owner-scoped AuthZ, dial-time SSRF guard, `?iface=` regex validation, bounded timeouts and a 60 s audit default.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** closes caps K2, K7 and A03 before any measurement work.
- **Usage:** execute after Chunk 0; gates Chunk 2.
- **Inherited constraints:** `01-core-directive.md` §1–4.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 2 — data honesty and integrity.
```

Then run: `sed -n '1,91p' /c/END/CGO3.md >> docs/MUSTREAD/11-chunk-1-security-lockdown.md`

- [ ] **Step 3: Write `12-chunk-2-data-honesty-integrity.md`**

Header:

```markdown
# Chunk 2 — Data Honesty, Integrity & Lab Physics

> Source: `C:\END\CGO4.md` lines 1–97. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** failed scrapes never become zero, `/api/integrity` verifies real files, topology fails closed, RTT is `DelayMs/2` per hop, writes are atomic and ordered.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** removes fake zeros and orphan manifests from the evidence chain.
- **Usage:** execute after Chunk 1; gates Chunk 3.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 3 — real-time triple-ring pipeline.
```

Then run: `sed -n '1,97p' /c/END/CGO4.md >> docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md`

- [ ] **Step 4: Verify**

Run: `grep -n '^## Part' docs/MUSTREAD/1{0,1,2}-*.md`
Expected: `Part A`, `Part B`, `Part C` present in each of the three files.

Run: `wc -l docs/MUSTREAD/1{0,1,2}-*.md`
Expected: no file empty.

- [ ] **Step 5: Commit**

```bash
git add docs/MUSTREAD/10-chunk-0-secret-purge-bloat.md docs/MUSTREAD/11-chunk-1-security-lockdown.md docs/MUSTREAD/12-chunk-2-data-honesty-integrity.md
git commit -m "docs(mustread): add chunk 0, 1 and 2 execution docs"
```

---

### Task 3: Split `CGO5.md` into Chunks 3–6

**Files:**
- Create: `docs/MUSTREAD/13-chunk-3-realtime-pipeline.md`
- Create: `docs/MUSTREAD/14-chunk-4-honest-tests.md`
- Create: `docs/MUSTREAD/15-chunk-5-container-ci-release.md`
- Create: `docs/MUSTREAD/16-chunk-6-live-proof-freeze.md`

**Interfaces:**
- Consumes: `C:\END\CGO5.md` lines 1–89, 90–167, 168–267, 268–410. The four ranges partition all 410 lines exactly once; transition paragraphs stay with the chunk they close.
- Produces: four chunk docs; Task 4 crosswalks them; Task 5 checks them.

- [ ] **Step 1: Split the source with exact line ranges**

Run these four commands in order:

```bash
sed -n '1,89p' /c/END/CGO5.md >> docs/MUSTREAD/13-chunk-3-realtime-pipeline.md
sed -n '90,167p' /c/END/CGO5.md >> docs/MUSTREAD/14-chunk-4-honest-tests.md
sed -n '168,267p' /c/END/CGO5.md >> docs/MUSTREAD/15-chunk-5-container-ci-release.md
sed -n '268,410p' /c/END/CGO5.md >> docs/MUSTREAD/16-chunk-6-live-proof-freeze.md
```

Expected: `wc -l` of the four appended bodies = 89, 78, 100, 143; sum = 410.

- [ ] **Step 2: Prepend each recall header before the appended body**

Use `13-chunk-3-realtime-pipeline.md` as the example; the other three use the same five bullet labels with their own goal text:

```markdown
# Chunk 3 — Real-Time Triple-Ring Pipeline

> Source: `C:\END\CGO5.md` lines 1–89. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** eBPF ringbuf → 20 Hz Go SSE ring → Float64Array ring → RingBufferHUD in a 2-mode UI.
- **Mechanism:** kernel events bypass React; rAF paints only when `ring.dirty`.
- **Contribution:** glass-smooth 20 Hz → 60 FPS telemetry with `NaN` for absent data.
- **Usage:** execute after Chunk 2; gates Chunk 4.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 4 — honest tests.
```

Write the corresponding header into `14-chunk-4-honest-tests.md`:

```markdown
# Chunk 4 — Blazing-Fast Honest Tests

> Source: `C:\END\CGO5.md` lines 90–167. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** real synctest on product state machines, node + happy-dom split, unmocked Playwright E2E and Axe WCAG gate.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** a fast suite that cannot pass on mocks of the module under test.
- **Usage:** execute after Chunk 3; gates Chunk 5.
- **Inherited constraints:** `01-core-directive.md` §2; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 5 — container, strict CI and signed release.
```

Write the corresponding header into `15-chunk-5-container-ci-release.md`:

```markdown
# Chunk 5 — Container, Strict CI, Signed Release & Ops

> Source: `C:\END\CGO5.md` lines 168–267. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** a working 3-stage Docker build, `--strict` gates with no exit-2 masking, pinned signed release with SBOM and SLSA L2, plus NTP, cert, log and backup basics.
- **Mechanism:** Part A human gate, Part B bounded agent prompt, Part C hard verification gate.
- **Contribution:** closes caps K3, K5, K9, K10 and K11.
- **Usage:** execute after Chunk 4; gates Chunk 6.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Chunk 6 — live proof and final freeze.
```

Write the corresponding header into `16-chunk-6-live-proof-freeze.md`:

```markdown
# Chunk 6 — Live Proof, Thesis & v2.0.0 Freeze

> Source: `C:\END\CGO5.md` lines 268–410. Lens: **actions + usage**.

## At a glance

- **Goal (concrete result):** Linux kernel truth, authorized field trace, `.dist` calibration, RouterOS CHR proof, synchronized thesis, signed v2.0.0 tag and hard freeze.
- **Mechanism:** Part A human-controlled runbook with explicit gates, Part B bounded packaging prompt, Part C terminal definition-of-done checklist.
- **Contribution:** closes every cap with a named artifact and then stops.
- **Usage:** execute last, only after Chunks 0–5 are green.
- **Inherited constraints:** `01-core-directive.md` §1–4; `02-target-system.md` §1.

## Part A — Human gate
## Part B — Agent execution steps
## Part C — Verification gate
## What it unlocks

Nothing — this is the final chunk.
```

- [ ] **Step 3: Verify the split is lossless**

Run:

```bash
for f in 13 14 15 16; do echo "--- $f ---"; head -n 1 docs/MUSTREAD/${f}-*.md; done
```

Expected: four distinct titles; no file starts with a blank line from the source.

- [ ] **Step 4: Commit**

```bash
git add docs/MUSTREAD/13-chunk-3-realtime-pipeline.md docs/MUSTREAD/14-chunk-4-honest-tests.md docs/MUSTREAD/15-chunk-5-container-ci-release.md docs/MUSTREAD/16-chunk-6-live-proof-freeze.md
git commit -m "docs(mustread): split corpus chunk 3-6 into one doc per chunk"
```

---

### Task 4: `INDEX.md` — recall surface, crosswalk and SDD ledger

**Files:**
- Create: `docs/MUSTREAD/INDEX.md`

**Interfaces:**
- Consumes: the exact filenames and source ranges from Tasks 1–3.
- Produces: the file named in the `AGENTS.md` injection (Task 6) and the crosswalk table read by Task 5's checker.

- [ ] **Step 1: Write the index with all seven required sections**

Create `docs/MUSTREAD/INDEX.md` with this exact skeleton, filling the tables as shown — no `TBD`, no omitted rows:

```markdown
# MUSTREAD — CGO / Meteolink canonical pack

> This index is the single entry point. `docs/MUSTREAD/01-core-directive.md` is the governing directive; the seven chunk docs are the execution units. Source corpus: `C:\END` (1,491 lines), never modified.

## 1. Overarching approach

METEOLINK is one Go binary with an embedded React cockpit and a Linux eBPF probe for a constrained meteorological link. This pack turns the seven-file CGO blueprint into a governed, chunk-by-chunk execution route: purge and simplify, lock security, enforce data honesty, install the real-time pipeline, prove it with honest tests, ship it in a signed container, then close with live evidence, the thesis and the v2.0.0 freeze. MUSTREAD governs where it conflicts with the operating card; operator-level bans still bind.

## 2. Five-lens map

| Lens | Document(s) |
|---|---|
| Result | `02-target-system.md`, `16-chunk-6-live-proof-freeze.md` |
| Mechanism | `01-core-directive.md`, `13-chunk-3-realtime-pipeline.md` |
| Contribution | `03-product-contribution.md` |
| Actions | `10-…` through `16-…` (one per chunk) |
| Usage | each doc's "At a glance" → "Usage" bullet |

## 3. Document map

| Doc | Source | Purpose |
|---|---|---|
| `01-core-directive.md` | `CGO.md` 1–237 | Governance: laws, loop, debug playbook, reviewer protocol |
| `02-target-system.md` | `CGO0.md` 1–286 | Zero-bloat target and 7-chunk route |
| `03-product-contribution.md` | `CGO1.md` 1–237 | Problem, personas, value, architecture brief |
| `10-chunk-0-secret-purge-bloat.md` | `CGO2.md` 1–133 | Purge + bloat + SSoT |
| `11-chunk-1-security-lockdown.md` | `CGO3.md` 1–91 | AuthZ, SSRF, injection, timeouts |
| `12-chunk-2-data-honesty-integrity.md` | `CGO4.md` 1–97 | Absent-never-zero, integrity, RTT |
| `13-chunk-3-realtime-pipeline.md` | `CGO5.md` 1–89 | eBPF → SSE → Float64Array → HUD |
| `14-chunk-4-honest-tests.md` | `CGO5.md` 90–167 | synctest, happy-dom, live Playwright |
| `15-chunk-5-container-ci-release.md` | `CGO5.md` 168–267 | Docker, strict CI, signed release, ops |
| `16-chunk-6-live-proof-freeze.md` | `CGO5.md` 268–410 | Live proof, field trace, thesis, freeze |

## 4. Persistent recall

- **Locked constants:** version `2.0.0`; audit `60 s` = 10/40/10; SLA `220 ms`; bind `127.0.0.1:9090`; canonical title « Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar ».
- **Precedence:** MUSTREAD > `AGENTS.md` operating card > skill defaults. Operator-level constitution bans cannot be waived by a project document.
- **Path binding:** corpus `pkg/qdisc/` → repo `qdisc/`; `pkg/results/` → `results/`; `pkg/remediate/` → `remediate/`; `pkg/probe/` → `probe/`; `pkg/audit/` → `audit/`; `cmd/gates` stays until a replacement six-gate runner exists.
- **Retained contradictions:** title, 2 vs 4 UI modes, `cmd/gates`, `2.0.0` vs `1.3.5`, stale `pkg/...` paths. All are preserved and resolved per chunk at plan time, never deleted.

## 5. Completeness crosswalk

| Source | Lines | Destination |
|---|---|---|
| `CGO.md` | 1–237 | `01-core-directive.md` |
| `CGO0.md` | 1–286 | `02-target-system.md` |
| `CGO1.md` | 1–237 | `03-product-contribution.md` |
| `CGO2.md` | 1–133 | `10-chunk-0-secret-purge-bloat.md` |
| `CGO3.md` | 1–91 | `11-chunk-1-security-lockdown.md` |
| `CGO4.md` | 1–97 | `12-chunk-2-data-honesty-integrity.md` |
| `CGO5.md` | 1–89 | `13-chunk-3-realtime-pipeline.md` |
| `CGO5.md` | 90–167 | `14-chunk-4-honest-tests.md` |
| `CGO5.md` | 168–267 | `15-chunk-5-container-ci-release.md` |
| `CGO5.md` | 268–410 | `16-chunk-6-live-proof-freeze.md` |

## 6. Execution status (SDD ledger)

| Task | Doc | Owner | State |
|---|---|---|---|
| Pack build + injection | `INDEX.md` | coordinator | pending |
| Chunk 0 | `10-…` | implementer + reviewer | pending |
| Chunk 1 | `11-…` | implementer + reviewer | pending |
| Chunk 2 | `12-…` | implementer + reviewer | pending |
| Chunk 3 | `13-…` | implementer + reviewer | pending |
| Chunk 4 | `14-…` | implementer + reviewer | pending |
| Chunk 5 | `15-…` | implementer + reviewer | pending |
| Chunk 6 | `16-…` | implementer + reviewer | pending |

## 7. Handoff

`writing-plans` derives one implementation plan per chunk from Part A / B / C of its doc. SDD executes chunks 0→6 in order, one subagent at a time, with a task review after each and ledger rows updated here.
```

- [ ] **Step 2: Verify the index**

Run: `grep -c '^## ' docs/MUSTREAD/INDEX.md`
Expected: `7`.

Run: `grep -c 'CGO' docs/MUSTREAD/INDEX.md`
Expected: at least 10 (document map + crosswalk rows).

- [ ] **Step 3: Commit**

```bash
git add docs/MUSTREAD/INDEX.md
git commit -m "docs(mustread): add recall index, source crosswalk and SDD ledger"
```

---

### Task 5: One-off completeness check

**Files:**
- Create (outside the repo): `C:\Users\ASUS\AppData\Local\Temp\opencode\check_mustread.py`
- Modify: none unless the check fails — then repair the affected doc and re-commit it in this task.

**Interfaces:**
- Consumes: the crosswalk in `INDEX.md` and the seven source files under `C:\END`.
- Produces: exit 0 with `MISSING: 0`, or a numbered list of missing source lines to repair.

- [ ] **Step 1: Write the checker**

Write this exact script to the temp path:

```python
from pathlib import Path

SRC = Path(r"C:\END")
PACK = Path(r"C:\Users\ASUS\.local\share\opencode\worktree\1c66ad257b56747880c45bc5d4f7e50f8688e1eb\operator-cockpit\docs\MUSTREAD")
RANGES = {
    "CGO.md": [(1, 237, "01-core-directive.md")],
    "CGO0.md": [(1, 286, "02-target-system.md")],
    "CGO1.md": [(1, 237, "03-product-contribution.md")],
    "CGO2.md": [(1, 133, "10-chunk-0-secret-purge-bloat.md")],
    "CGO3.md": [(1, 91, "11-chunk-1-security-lockdown.md")],
    "CGO4.md": [(1, 97, "12-chunk-2-data-honesty-integrity.md")],
    "CGO5.md": [(1, 89, "13-chunk-3-realtime-pipeline.md"),
                (90, 167, "14-chunk-4-honest-tests.md"),
                (168, 267, "15-chunk-5-container-ci-release.md"),
                (268, 410, "16-chunk-6-live-proof-freeze.md")],
}

missing = []
for name, ranges in RANGES.items():
    lines = (SRC / name).read_text(encoding="utf-8").splitlines()
    cursor = 1
    for start, end, target in ranges:
        if start != cursor:
            missing.append((name, f"gap/overlap before {start}"))
        cursor = end + 1
        body = (PACK / target).read_text(encoding="utf-8")
        for n in range(start, end + 1):
            text = lines[n - 1].strip().lstrip("#").strip()
            if text and text not in body:
                missing.append((name, n, text[:70]))
    if cursor != len(lines) + 1:
        missing.append((name, f"tail gap: ends at {cursor - 1}, file has {len(lines)}"))

print("MISSING:", len(missing))
for item in missing[:30]:
    print(item)
assert not missing, "source lines lost or ranges not contiguous"
```

- [ ] **Step 2: Run it**

Run: `python "C:\Users\ASUS\AppData\Local\Temp\opencode\check_mustread.py"`
Expected: `MISSING: 0`, exit 0.

If it reports missing lines: append the exact reported lines to the named destination doc with a `> Source line <n>` prefix, re-run, then commit the repaired doc with an explicit path. Do not edit `C:\END`.

- [ ] **Step 3: Remove the temp checker**

Run: `rm "C:\Users\ASUS\AppData\Local\Temp\opencode\check_mustread.py"`
Expected: file absent. The checker is one-off, not a permanent gate.

- [ ] **Step 4: Commit only if a repair was needed**

If Step 2 changed a doc: `git add <the-repaired-doc>` and `git commit -m "docs(mustread): repair source-line completeness gap"`.
If no repair: no commit in this step.

---

### Task 6: Inject CGO into core instructions

**Files:**
- Modify: `AGENTS.md` (insert after line 1 only; all existing lines below remain verbatim)
- Modify: `docs/superpowers/plans/README.md` (one line after the `*Last updated:*` header)

**Interfaces:**
- Consumes: `docs/MUSTREAD/INDEX.md` from Task 4.
- Produces: the core directive that every future session reads first.

- [ ] **Step 1: Insert the core-directive block**

After the first line (`# CGO — agent operating card`) of `AGENTS.md`, insert exactly:

```markdown

## CGO Core Directive (MUSTREAD)

`docs/MUSTREAD/INDEX.md` is the canonical source of direction for this project.
Where this card conflicts with the MUSTREAD pack, MUSTREAD governs; the sections
below are retained as subordinate operating detail. Operator-level constitution
bans cannot be waived by a project document.

### Precedence

| Conflict | Resolution |
|---|---|
| Corpus forbids `CHECKPOINTS.md`/`VERIFY.md` edits outside the final chunk | Wins; the existing checkpoint cadence becomes subordinate, files stay untouched by chunk work |
| Chunk 6 uses `git add -A` | Cannot be waived by a project document; the commit task stays, staging becomes explicit paths |
| Chunk 0 deletes `cmd/gates` | Directive stands, but a replacement for the six-gate runner must exist before removal — no gate tampering |
| Canonical title, 2 modes, version `2.0.0` | Locked; present repo values are labelled historical, not deleted |
| Corpus `pkg/...` paths | Kept verbatim; the INDEX path-binding note gives this repo's real paths |
```

- [ ] **Step 2: Add the plan-index pointer**

Immediately after the line beginning `*Last updated:` in `docs/superpowers/plans/README.md`, insert:

```markdown
*Canonical execution direction: `docs/MUSTREAD/INDEX.md` (CGO directive wins on conflict).*
```

- [ ] **Step 3: Verify existing instructions survive**

Run: `grep -n '^## ' AGENTS.md`
Expected: all original headings still present below `## CGO Core Directive (MUSTREAD)`, including `## Read first`, `## The one command`, `## Never touch`, `## Never commit`, `## Workflow`, `## Cadence`, `## Roles`, `## Language`, `## Stop and ask`, `## WEBSERVER (sadcnwp) access`, `## FAST TOOLS — standing arsenal (ingested 2026-11-14)`, `## Pointers`.

Run: `git diff --stat -- AGENTS.md docs/superpowers/plans/README.md`
Expected: insertions only; no deletions in `AGENTS.md`.

- [ ] **Step 4: Verify the pack file count**

Run: `ls docs/MUSTREAD/*.md | wc -l`
Expected: `11` (INDEX + 10 docs).

Run: `git status --short`
Expected: only the two files from this task are newly modified; all unrelated protected paths are unchanged by this plan.

- [ ] **Step 5: Commit**

```bash
git add AGENTS.md docs/superpowers/plans/README.md
git commit -m "docs(mustread): inject CGO core directive into project instructions"
```

---

### Task 7: Final acceptance

**Files:** none new.

**Interfaces:**
- Consumes: all six tasks' outputs.
- Produces: a recorded acceptance result against the spec's five criteria.

- [ ] **Step 1: Check the document count and names**

Run: `ls docs/MUSTREAD/`
Expected: `INDEX.md`, `01-…`, `02-…`, `03-…`, `10-…`, `11-…`, `12-…`, `13-…`, `14-…`, `15-…`, `16-…`.

- [ ] **Step 2: Confirm the corpus is untouched**

Run: `ls -l /c/END/*.md`
Expected: seven files with their original byte sizes — `CGO.md` 15056, `CGO0.md` 19430, `CGO1.md` 18070, `CGO2.md` 7661, `CGO3.md` 8544, `CGO4.md` 7939, `CGO5.md` 31460. No file was edited.

- [ ] **Step 3: Confirm no operator-level ban was introduced**

Run: `git grep -n 'git add -A' docs/MUSTREAD/ AGENTS.md`
Expected: the only hits are the precedence rows that name the ban as forbidden. No executable `git add -A` step.

- [ ] **Step 4: Record the result**

Write one checkpoint entry in `docs/CHECKPOINTS.md` (append only, newest at bottom) with the exact commands from Tasks 5–7, their exit codes and the resulting file count. Note what stayed unproven: no chunk was executed; each chunk still needs its own `writing-plans` plan and SDD run.

- [ ] **Step 5: Commit the checkpoint**

```bash
git add docs/CHECKPOINTS.md
git commit -m "docs(mustread): record pack-build acceptance checkpoint"
```
