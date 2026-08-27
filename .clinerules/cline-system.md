# Cline system copy

This is Cline's own operating system, copied into the repo so it travels with
the code — Cline's equivalent of the `AGENTS.md` that opencode uses. Cline
auto-loads every `.md` file in this `.clinerules/` folder at the start of each
task. Sibling files here (e.g. `ponytail.md`) are always-on and bind every
task in this repo.

- **What Cline is:** an agentic coding system — it plans, edits files, runs
  commands, verifies results and reports back in a loop, using tools.
  Underlying model: GLM (trained by Z.ai).
- **Modes:** `plan` (explore, analyze, align on a plan — no edits and no
  state-changing commands) and `act` (implementation allowed). The newest
  message's mode governs; a mode notice marks exactly when a switch happened.

## Working style

1. Gather all context before acting: read the relevant files, search the
   codebase, and learn the conventions, frameworks, libraries and the
   commands used to run and test. Never guess or invent.
2. Present a short plan before executing any task.
3. Batch independent tool calls (reads, searches, commands) in one turn;
   sequence only true dependencies.
4. Use absolute paths for all file references. Read a file before editing it.
5. Re-verify every edit or creation at the end of the task (read it back,
   run the relevant checks).
6. Deliver complete, functional output — no placeholders, no omissions.
   Be explicit about assumptions and limitations.
7. Ask a clarifying question only when a decision truly blocks progress;
   otherwise make the reasonable call and state it.

## Project workflow — AGENTS.md is canonical

`AGENTS.md` at the repo root holds the binding workflow for every code
change. Summary (full text lives there):

- **Isolation first:** worktree per task — `git worktree add
  .worktrees/<task-slug> -b agent/<task-slug>`; work only inside the
  worktree; never checkout/switch branches in the main checkout; frontend
  deps per worktree via `cd web/frontend && bun install`.
- **Branch & commit discipline:** conventional commits
  (`feat:` / `fix:` / `refactor:` / `docs:` / `test:` / `chore:`); never
  commit directly to `main` (exception: trivial `chore:` config such as
  `.gitignore` / `AGENTS.md` / this `.clinerules` folder); never commit
  secrets or generated dirs; commit only task-relevant files.
- **Verification before done** (failing checks = not done; show real
  output): `make test` then `make test-real`; static pass `go vet ./...`;
  `pkg/api` has pre-existing Windows file-lock failures, so run
  `make test-linux` under WSL/Linux for oracle/detect/stats/negctl;
  frontend (`web/frontend/`, when touched): `bun run typecheck` then
  `bun run build`. Never modify `data/frozen-wave1|2`.
- **Protected operations (ask before doing):** `push --force`,
  `reset --hard` on shared refs, branch deletion, dependency version bumps
  (`go.mod`, `package.json`), anything touching `deploy/deploy.env`, VM
  configs or secrets, changes to frozen experiment data or provenance
  manifests.
- **Session hygiene:** no session exports or summary files inside the repo;
  the final message reports goal, files changed, commands run and
  verification results; on completion merge to main, then remove the
  worktree and delete the merged `agent/*` branch.

## Skills

Cline loads skills on demand via its `skills` tool; when a request matches a
skill, invoking it is a blocking requirement before responding. The global
skill set on this machine is shared with opencode and lives in mirrored
directories (`~` = `C:\Users\ASUS`):

- `~\.config\opencode\skills\` — opencode global skills (canonical)
- `~\.claude\skills\` — Claude mirror (Cloudflare pack)
- `~\.agents\skills\` — agents mirror (superpowers + mattpocock pack)

Project-level overrides would live in `.opencode/skill/` or
`.claude/skills/`; this repo has none (its `.superpowers/` dir holds SDD
session artifacts — data, not skills).

### Mandated by the user for this repo

- **impeccable** — award-director design skill. MUST be invoked for any
  frontend/UI task in `web/frontend/` (commands: init, shape, polish,
  harden, animate, colorize, typeset, layout, delight, overdrive, clarify,
  adapt, optimize, live). Installed globally at
  `~\.config\opencode\skills\impeccable\` (this repo has no local
  `.opencode/` copy, so run its `scripts/context.mjs` from the global path).
  It reads `PRODUCT.md`, `DESIGN.md` and surface briefs — `DESIGN.md`
  already exists at the repo root.
- **ponytail** — the lazy-senior-dev ruleset (DietrichGebert/ponytail, MIT;
  https://ponytail.dev). On Cline it ships as an instruction-only adapter:
  the always-on ruleset, no chat commands (the `/ponytail*` commands need
  skill-capable hosts). Its canonical Cline file is copied verbatim into
  this folder as `ponytail.md` (source: `.clinerules/ponytail.md` in the
  ponytail repo, fetched 2026-08-27). Intensity default `lite`; change via
  optional `~\.config\ponytail\config.json` or `PONYTAIL_DEFAULT_MODE`
  (`lite` / `full` / `ultra` / `off`).

### Routing — invoke the matching skill before acting

- UI/design work in `web/frontend/` → `impeccable`
- Any bug, test failure or unexpected behavior → `systematic-debugging`
- Multi-step feature from a spec → `writing-plans`, then `executing-plans`
- Test-first feature/bugfix → `tdd`
- About to claim work is complete → `verification-before-completion`
- Library/framework internals → `clonedeps`; current docs → `context7`
- Go backend work → `golang-backend-development` (global opencode skill)

### Global skill index (verified on disk 2026-08-27)

- **Cloudflare/infra:** agents-sdk, cloudflare, cloudflare-email-service,
  cloudflare-one, cloudflare-one-migrations, durable-objects, sandbox-next,
  sandbox-stable, sandbox-migrate-to-next, turnstile-spin, web-perf,
  workers-best-practices, wrangler, orca-cli
- **Discovery/meta:** find-skills, using-superpowers, writing-skills,
  writing-for-agents, ask-matt, computer-use, wizard, scaffold-exercises,
  setup-matt-pocock-skills
- **Planning/writing:** brainstorming, writing-plans, writing-fragments,
  writing-shape, writing-beats, to-questionnaire, to-spec, to-tickets,
  triage, wayfinder, deepwork, orchestration, loop-me, prototype, research
- **Execution:** implement, executing-plans, subagent-driven-development,
  dispatching-parallel-agents, task-management, tdd, test-driven-development,
  verification-planning, verification-before-completion
- **Quality:** code-review, requesting-code-review, receiving-code-review,
  simplify, systematic-debugging, diagnosing-bugs, grilling, grill-me,
  grill-with-docs, wait-what, reflect, caveman, teach
- **Git:** using-git-worktrees, worktrees, finishing-a-development-branch,
  resolving-merge-conflicts, git-guardrails-claude-code, claude-handoff,
  handoff
- **Architecture:** codebase-design, domain-modeling,
  improve-codebase-architecture, setup-ts-deep-modules, migrate-to-shoehorn,
  setup-pre-commit, clonedeps, codemap, context7
- **Design/frontend:** impeccable, frontend-design, design-taste-frontend,
  high-end-visual-design
- **Go:** golang-backend-development
- **Also on disk (opencode global):** oh-my-opencode-slim

Full instructions for each: `<skill-dir>\SKILL.md` in the global dirs above.

