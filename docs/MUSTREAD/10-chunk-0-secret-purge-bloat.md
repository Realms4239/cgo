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
We lock the Design Brief as our constitution. Let’s execute Chunk 0 now.

Because K1 (Cap 40 PAT history purge) involves revoking credentials and
rewriting public Git history (git filter-repo + --force), you run Part A
yourself—never trust an agent with credential revocation or --force pushes. Then
hand Part B verbatim to the agent, and verify Part C before touching Chunk 1.

CHUNK 0 — Secret Purge (K1), Bloat Amputation & SSoT Alignment

PART A — Human Security Steps (Run This First on Your Machine)

1. Revoke the Leaked PAT & Rotate Local Keys

2.  Open GitHub → Settings → Developer settings → Personal access tokens and
    Revoke/Delete the classic ghp_ PAT committed in 54d3f62
    (book/handoff-livrables-2026-09-02.md:23).
3.  Check GitHub Settings → Security log and your repository Actions + Releases
    history from 2026-09-02 to today; verify no unrecognized workflows or tags
    ran, and save the log locally (e.g.
    D:\cgo-audit-evidence\github-token-audit.txt, outside the repo).
4.  Rotate any active API keys inside your untracked local .env, and wipe any
    hardcoded values in .env.example so only variable names remain.

5. Purge book/handoff-livrables-2026-09-02.md From All Git History

(Run from MINGW64 / Git Bash in C:\cgo. Make a full mirror backup on D: first!)

# 1. Backup full mirror first (outside C:\cgo)
git clone --mirror . /d/cgo-mirror-backup-2026-11-17.git

# 2. Check git-filter-repo is available (install via pip/winget if needed)
git filter-repo --version

# 3. Purge the file across EVERY commit, branch, and tag
git filter-repo --invert-paths --path book/handoff-livrables-2026-09-02.md --force

# 4. Verify ZERO commits contain the ghp_ prefix anywhere in history
git log --all -S'ghp_' --oneline
# Expected output: EMPTY (0 lines)

# 5. Re-add origin remote if filter-repo stripped it, then force-push all refs & tags
git remote add origin <YOUR_GITHUB_REMOTE_URL> 2>/dev/null || true
git push origin --force --all
git push origin --force --tags

(Note: If you use GitHub, open a quick ticket or let GC run if old commit SHAs
remain cached in PR views; the unreachable blob is gone from your branches and
all 47+ tags.)

PART B — Exact Task Spec for Your Coding Agent (Chunk 0 Only)

Copy and paste this exact bounded prompt to your agent. Notice the strict File
Allowlist and Forbidden List so it cannot wander or invent new architecture:

# CHUNK 0 EXECUTION TASK — BLOAT AMPUTATION & IDENTITY SSoT (STRICTLY BOUNDED)

You are executing ONLY Chunk 0 on `C:\cgo`. Do NOT touch authentication (`api/auth.go`), SSRF (`api/ssrf.go`), probes (`pkg/probe/`), `Dockerfile`, CI YAML, or UI views unless explicitly named below.

## 1. Delete Confirmed Bloat & Dead Shells
Remove these exact paths from git (`git rm -rf` / `rm -rf`) and update any build/script/test references broken by removing them:
- `deploy/nginx/` (Go stdlib `http.Server` handles TLS/gzip/rate-limit natively; Nginx config is broken bloat)
- `deploy/prometheus/` (unmapped alert rules with no runbooks; drop Nginx/Prometheus compose references if any)
- `modules/` (empty directory)
- `cmd/cgo-gui/` (unused Fyne GUI excluded from releases)
- `cmd/meteolink/` (redundant 54-line re-exec shim; `cmd/cgo` is the single binary)
- `api/synctest_test.go` (toy test of Go's virtual clock; real CGO synctest tests belong in Chunk 4)
- Any leftover `npm/` directory or `cmd/cgo-chr/` mock harness if still present on disk

Update `.goreleaser.yaml`, `Makefile`, `deploy/docker-compose.yml`, and `cmd/gates/` (only where referencing deleted `cmd/cgo-gui` or `cmd/meteolink` or `deploy/nginx`/`deploy/prometheus`). Do NOT rewrite `Dockerfile` or CI workflows yet (that is Chunk 5).

Clean Go module dependencies:
- Remove unused `fyne.io/fyne/v2` (and any unused `go-i18n`, `nfnt/resize`, `oksvg`, `glfw`, etc.) by running `go mod tidy` after deleting `cmd/cgo-gui`.
- Verify `grep -rn "fyne" go.mod pkg/ cmd/ api/` returns 0 hits.

## 2. Eliminate `meteolink.dev` OS `hosts`-File Hijacking
- In `cmd/cgo/main.go` (`:60-68`), `cmd/cgo/setup.go`, `cmd/cgo/server.go`, and `cmd/cgo/cert.go`:
  - Change the default bind address from `meteolink.dev:9090` to `127.0.0.1:9090`.
  - Remove automatic mutation of the OS `hosts` file during default setup/run.
  - Ensure TLS SANs in `cmd/cgo/cert.go` target `localhost`, `127.0.0.1`, and `::1` cleanly.
  - If an HTTP->HTTPS redirect listener on `:9080` exists, redirect cleanly to `https://127.0.0.1:9090` (never `meteolink.dev`), or make `:9080` optional.

## 3. Lock Single Source of Truth (Brand, Version, Canonical Thesis Title, Dead Docs)
1. **Version SSoT:**
   - Set `VERSION` to `2.0.0-rc.1`.
   - Ensure `cmd/cgo` embeds `VERSION` (or shares single embedded `VERSION` string — no hardcoded `"1.3.5"` or `"1.2.3"` drift).
   - Sync `"version": "2.0.0-rc.1"` in `web/frontend/package.json`, `README.md:303` health example, and `docs/api.md:14`.
2. **Product Name & Binary:**
   - Standardize user-facing product name to `METEOLINK` across `web/frontend/index.html`, `README.md`, `PRODUCT.md`, and `DESIGN.md`.
   - Update `README.md:36-37` to state truthfully: one Go binary (`cgo`, built from `cmd/cgo`) + gate runner (`cmd/gates`, if retained for CI). Delete `"three skins, one engine"` / `"12 independently versioned tools"` prose from docs.
3. **Canonical Master 2 Title:**
   - Verify and enforce this exact title verbatim in `README.md`, `PRODUCT.md`, `book/` front matter, and `docs/VERIFY.md`:
     `Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar`
4. **Clean Stale Pointers:**
   - Update `docs/superpowers/plans/README.md:3` branch/date header, and remove references to deleted `deploy/nginx`, `deploy/prometheus`, `modules/`, `cmd/cgo-gui`, `cmd/meteolink`, and `docs/SPEC.md` across active docs.

## Constraints
- Do NOT write to `docs/CHECKPOINTS.md` except adding a single freeze notice at top: `> Frozen post-2026-11-17 audit reset.`
- Do NOT claim anything is verified unless you actually ran the verification command and paste the exact output.
- Commit all Chunk 0 work in ONE clean commit: `chore(reset): purge bloat modules, unify SSoT identity and bind 127.0.0.1`.

PART C — Your Hard Verification Gate (Run Before Advancing to Chunk 1)

Run these commands yourself after the agent finishes. Do not advance to Chunk 1
unless every check passes:

# 1. Confirm Cap 40 PAT is completely gone from reachable history
git log --all -S'ghp_' --oneline
# Must print NOTHING

# 2. Confirm bloated directories/files are really deleted
ls -d deploy/nginx deploy/prometheus modules cmd/cgo-gui cmd/meteolink api/synctest_test.go 2>&1
# Every single one must report "No such file or directory"

# 3. Confirm fyne is gone from go.mod
grep -i "fyne" go.mod || echo "CLEAN_NO_FYNE"

# 4. Confirm meteolink.dev bind default is gone from cmd/cgo
git grep "meteolink.dev" -- cmd/ pkg/ api/ || echo "CLEAN_NO_METEOLINK_DEV"

# 5. Confirm VERSION SSoT matches across files
cat VERSION
grep '"version"' web/frontend/package.json
grep -n "2.0.0-rc.1" README.md docs/api.md

# 6. Confirm single binary builds and unit tests still pass cleanly
go build ./cmd/cgo
go test ./cmd/cgo ./pkg/...

Once all 6 checks pass cleanly (and note: if 16:22Z / 19:22 EAT is coming up in
~55 minutes and you want to run your scheduled field capture safely in parallel
without touching C:\cgo files, keep the field trace artifacts separate until
Chunk 6!), paste the output or say "Chunk 0 verified — proceed with Chunk 1".

