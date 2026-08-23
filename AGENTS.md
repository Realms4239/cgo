# Agent Workflow Rules

## Isolation first — mandatory for any code change

1. Already isolated? (`git rev-parse --git-dir` != `git rev-parse --git-common-dir`) → work where you are, skip creation.
2. Otherwise create a worktree before editing anything:
   ```
   git worktree add .worktrees/<task-slug> -b agent/<task-slug>
   cd .worktrees/<task-slug>
   ```
3. Work only inside the worktree. NEVER checkout or switch branches in the main checkout — the user edits live there.
4. Frontend deps per worktree: `cd web/frontend && bun install` (hardlinks from cache, no re-download). Go deps come from the shared module cache automatically.

## Branch & commit discipline

- Never commit directly to `main` (exception: trivial `chore:` config like .gitignore/AGENTS.md).
- Conventional commits: `feat:` / `fix:` / `refactor:` / `docs:` / `test:` / `chore:`
- Never commit secrets (`.env*`, keys, pats — see .gitignore) or generated dirs (`node_modules/`, `dist/`, `out/`).
- Commit scope: only files relevant to the task.

## Verification before done — show real output, failing checks = not done

- Go (Windows-safe): `make test` then `make test-real`
  - Note: full `pkg/api` suite has pre-existing Windows file-lock failures; run `make test-linux` under WSL/Linux for oracle/detect/stats/negctl.
  - Static pass: `go vet ./...`
- Frontend (`web/frontend/`, when touched): `bun run typecheck` then `bun run build`
- Experiment data lives under `data/frozen-wave1|2` — never modify frozen waves.

## Protected operations — ask before doing

- `push --force`, `reset --hard` on shared refs, branch deletion
- Dependency version bumps (go.mod, package.json)
- Anything touching `deploy/deploy.env`, VM configs, or secrets
- Changes to frozen experiment data or provenance manifests

## Session hygiene

- No session exports or summary files inside the repo. The final message reports: goal, files changed, commands run, verification results.
- On completion: merge to main, then `git worktree remove .worktrees/<slug>` and delete the merged `agent/*` branch.
