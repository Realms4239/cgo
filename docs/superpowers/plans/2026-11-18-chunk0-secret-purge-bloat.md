# CGO / Meteolink Chunk 0 — Secret Purge, Bloat Amputation & SSoT Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. **Task 1 is HUMAN-ONLY and must never be dispatched to a subagent.**

**Goal:** Complete Chunk 0 of the MUSTREAD route: remove the confirmed bloat, bind the service to loopback without hosts-file mutation, and align version/brand/title single sources of truth.

**Architecture:** The human owns credential revocation and the history rewrite (force-push). The agent owns the bounded tree edit and the verification gate. The corpus's `pkg/...` paths are mapped to this repo's real top-level paths per `docs/MUSTREAD/INDEX.md` §4 path binding. All Chunk 0 agent work lands in ONE commit, as the corpus mandates.

**Tech Stack:** Go 1.25 multi-module, React/Vite frontend, Git Bash, GoReleaser config, Windows-host verification.

**Source:** `docs/MUSTREAD/10-chunk-0-secret-purge-bloat.md` (verbatim `C:\END\CGO2.md` 1–133). Spec: `docs/superpowers/specs/2026-11-17-mustread-reorganization-design.md`.

## Global Constraints

- MUSTREAD governs where it conflicts with `AGENTS.md`; operator-level constitution bans still bind. Never `git add -A`/`git add .`; stage explicit paths only. No `--no-verify`.
- Part A (PAT revoke, `git filter-repo`, force-push) is human-only. No agent may run it, and no agent may weaken a gate to make its verification pass.
- The frozen `book/` corpus and `book/data/runs` are never touched. The corpus's "enforce title in `book/`" line is a human decision, not an agent step.
- `docs/CHECKPOINTS.md` has ~7.4 KiB headroom under the 400 KiB ceiling: do not append a checkpoint in this plan.
- Do not modify `cmd/gates/main.go`, `cmd/gates/main_test.go`, `.github/workflows/`, `Dockerfile`, or any test outside the paths named in Task 2. If a gate fails because of a deleted path, STOP and report — do not add a permit.
- Do not write to `docs/VERIFY.md` or `docs/CHECKPOINTS.md`.
- English plan/gate text; French user-facing text and French source comments stay in the language they already use.
- One agent commit only, message exactly `chore(reset): purge bloat modules, unify SSoT identity and bind 127.0.0.1`.
- Target version is `2.0.0-rc.1` (not final `2.0.0`).

---

### Task 1: HUMAN-ONLY — Cap 40 purge (decision gate)

**Files:** none in this repo. This task is never dispatched to an agent.

**Interfaces:**
- Consumes: the operator's GitHub credentials and a writable remote.
- Produces: zero commits touching the leaked PAT file, verified by `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md`. Human ruling 2026-11-18: the global `-S'ghp_'` check is unsatisfiable because the gate's token regex/selftest and historical docs legitimately contain the literal; only the leaked file's history is the oracle.

- [ ] **Step 1: Revoke and rotate (human)** — In GitHub UI revoke the classic `ghp_` PAT committed in `54d3f62` (`book/handoff-livrables-2026-09-02.md:23`); review Security/Actions/Releases logs from 2026-09-02 to today; save the log OUTSIDE the repo (e.g. `D:\cgo-audit-evidence\github-token-audit.txt`); rotate keys in local untracked `.env`.
- [ ] **Step 2: Mirror backup (human)** — `git clone --mirror . /d/cgo-mirror-backup-2026-11-18.git` from `C:\cgo` in Git Bash. Confirm the directory exists before continuing.
- [ ] **Step 3: Verify the tool (human)** — `git filter-repo --version`. If absent, install it as a human decision; do not let an agent install dependencies.
- [ ] **Step 4: Purge (human)** — `git filter-repo --invert-paths --path book/handoff-livrables-2026-09-02.md --force`.
- [ ] **Step 5: Verify zero hits (human)** — `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md`. Expected: EMPTY (currently two commits: `54d3f62`, `8c67859`). Record the exact output.
- [ ] **Step 6: Force-push (human, destructive)** — Re-add origin if stripped: `git remote add origin <YOUR_GITHUB_REMOTE_URL> 2>/dev/null || true`; then `git push origin --force --all` and `git push origin --force --tags`. Only after Step 5 is empty.
- [ ] **Step 7: Record** — Save the command output outside the repo and report to the coordinator. Task 3's check 1 stays BLOCKED until this task reports success.
- [ ] **Step 8: Human ruling on the frozen book** — The corpus also asks to enforce the canonical title in `book/` front matter, but `book/` is frozen evidence. Human decision: leave `book/` byte-unchanged and record the ruling outside the repo. No agent may edit `book/` for this title.

---

### Task 2: Agent implementation — bloat, loopback bind, SSoT

**Files:**
- Delete: `deploy/nginx/`, `deploy/prometheus/`, `modules/`, `cmd/meteolink/`, `api/synctest_test.go`
- Modify: `.goreleaser.yaml`, `go.mod`, `go.sum`, `cmd/cgo/main.go`, `cmd/cgo/setup.go`, `cmd/cgo/server.go`, `cmd/cgo/cert.go`, `cmd/cgo/cert_test.go`, `cmd/cgo/kit.go`, `cmd/cgo/kittui.go`, `cmd/cgo/tui.go`, `VERSION`, `web/frontend/package.json`, `web/frontend/index.html`, `README.md`, `PRODUCT.md`, `docs/api.md`, `docs/superpowers/plans/README.md`

**Interfaces:**
- Consumes: Task 1 only as a gate for final verification; no code dependency.
- Produces: one commit `chore(reset): purge bloat modules, unify SSoT identity and bind 127.0.0.1`; `go build ./cmd/cgo` and `go test ./cmd/cgo ./pkg/... -count=1` green.

- [ ] **Step 1: Amputate the bloat**

Run: `git rm -r deploy/nginx deploy/prometheus modules cmd/meteolink api/synctest_test.go`
Expected: 5 paths staged for deletion. `cmd/cgo-gui` is already absent; do not recreate or reference it.

- [ ] **Step 2: Remove the `meteolink` build from GoReleaser**

In `.goreleaser.yaml`, delete the whole block starting at the comment `# le wrapper meteolink re-exécute cgo` through the build entry ending at `ldflags: - -s -w` (the `id: meteolink` / `main: ./cmd/meteolink` / `binary: meteolink` build).
Change the archive comment line `# archive unique : cgo-<os>-<arch> contient cgo + meteolink + kit/` to `# archive unique : cgo-<os>-<arch> contient cgo + kit/`.
Delete the release-note line ``    - `meteolink` — alias de compatibilité (re-exécute cgo)``.
Verify: `grep -n 'cmd/meteolink\|alias de compatibilité' .goreleaser.yaml` prints nothing.

- [ ] **Step 3: Tidy Go dependencies**

Run: `go mod tidy`
Expected: `fyne.io/fyne/v2` removed from `go.mod`. Verify: `grep -i 'fyne' go.mod || echo CLEAN_NO_FYNE` prints `CLEAN_NO_FYNE`. If any `fyne` reference remains in a tracked Go file, STOP and report — do not hand-edit `go.mod` further.

- [ ] **Step 4: Replace the canonical bind with loopback**

Apply these exact replacements (same file may need several):

| File | Old text | New text |
|---|---|---|
| `cmd/cgo/main.go:62` | `def = "meteolink.dev:9090" // interface canonique (hosts → 127.0.0.1) ; --tls=false + 0.0.0.0 pour la VM` | `def = "127.0.0.1:9090" // interface canonique en loopback ; --tls=false + 0.0.0.0 pour la VM` |
| `cmd/cgo/main.go:3` | `https://meteolink.dev:9090 par défaut` | `https://127.0.0.1:9090 par défaut` |
| `cmd/cgo/main.go:65` | `HTTPS avec le certificat local meteolink.dev` | `HTTPS avec le certificat local (SAN localhost/127.0.0.1)` |
| `cmd/cgo/server.go:421` | `target := "https://meteolink.dev:" + port` | `target := "https://127.0.0.1:" + port` |
| `cmd/cgo/server.go:414` | `redirectHTTPS renvoie tout vers https://meteolink.dev:<port> + chemin.` | `redirectHTTPS renvoie tout vers https://127.0.0.1:<port> + chemin.` |
| `cmd/cgo/server.go:390` | `log.Printf("cgo redirect http://%s → https (meteolink.dev)", httpAddr)` | `log.Printf("cgo redirect http://%s → https (127.0.0.1)", httpAddr)` |
| `cmd/cgo/server.go:393` | `log.Printf("cgo dashboard on https://%s (certificat local meteolink.dev)", addr)` | `log.Printf("cgo dashboard on https://%s (certificat local)", addr)` |
| `cmd/cgo/tui.go:92` | `api = "https://meteolink.dev:9090"` | `api = "https://127.0.0.1:9090"` |
| `cmd/cgo/tui.go:97` | `HTTPS : racines système + certificat local meteolink.dev` | `HTTPS : racines système + certificat local` |

In `cmd/cgo/server.go`, delete the hosts hint block: the comment at `:373` and the `if host, _, herr := net.SplitHostPort(addr); herr == nil && host == "meteolink.dev" { hint = ... }` branch at `:377-378`. Keep the surrounding error handling intact.

- [ ] **Step 5: Remove hosts-file mutation from setup**

In `cmd/cgo/setup.go`, delete the numbered DNS block: the comment at `:187` (`// 6 — DNS local ...`), the `if confirm("Ajouter meteolink.dev/meteolink.vm au hosts local ?") { addHosts(...) }` block at `:189-191`, the Windows Admin note at `:205`, and change the launch hint at `:215` to `→ lancez ensuite : cgo --serve  (https://127.0.0.1:9090, certificat local auto-signé : accepter une fois dans le navigateur)`.
Leave `addHosts()` and `hostsPath()` defined but unreferenced (deleting them is out of scope); if `go vet` reports them as dead code, STOP and report.
In `cmd/cgo/kit.go:35`, change `dns mappe meteolink.dev vers la VM (hosts local, lien stable)` to `dns réservé — le hosts n'est plus modifié (127.0.0.1 par défaut)`.
In `cmd/cgo/kittui.go:243`, change the label `"Mapper meteolink.dev"` to `"DNS (désactivé — loopback)"`.

- [ ] **Step 6: Restrict the local certificate SANs to loopback**

In `cmd/cgo/cert.go`: change `localCertDNS` at `:30` to `[]string{"localhost"}`; change the SAN presence check at `:92` from `n == "meteolink.dev"` to `n == "localhost"`; change CommonName at `:111` to `"localhost"`; update the three header comments at `:3,6,7` and the comment at `:68` to name `localhost`/`127.0.0.1`/`::1` instead of `meteolink.dev`. Keep `IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}` unchanged.
In `cmd/cgo/cert_test.go`: replace `meteolink.dev` with `localhost` at `:50`, `:55`, `:107-108`; at `:113` change `redirectHTTPS("meteolink.dev:9090")` to `redirectHTTPS("127.0.0.1:9090")`; at `:120` change the expected URL to `https://127.0.0.1:9090/api/health`.

- [ ] **Step 7: Align the single sources of truth**

Set the contents of `VERSION` to exactly `2.0.0-rc.1` (no other line).
In `web/frontend/package.json:4`, change `"version": "1.3.5"` to `"version": "2.0.0-rc.1"`.
In `README.md:1`, change the badge fragment `version-1.3.5-blue` to `version-2.0.0-rc.1-blue`.
In `docs/api.md:16`, change `"version": "1.2.3"` to `"version": "2.0.0-rc.1"`.
In `README.md`, add the canonical title line immediately below the H1:
`**Optimisation temps réel de la qualité de service des liens d’accès contraints : une architecture d’observabilité eBPF et de régulation AQM/BBR à preuves scellées pour les flux critiques de Météo Madagascar**`
`PRODUCT.md:5` already carries that exact title — verify, do not duplicate.
In `web/frontend/index.html`, if a user-facing product name differs from `METEOLINK`, change it to `METEOLINK`; if already `Meteolink`, change it to `METEOLINK`.
In `docs/superpowers/plans/README.md:3`, change the header line to `*Last updated: 2026-11-18 · Branch: main (worktree) · Chunk 0 reset*`; keep the MUSTREAD pointer line from the pack build directly below it.

- [ ] **Step 8: Clean stale pointers**

Run: `git grep -n 'deploy/nginx\|deploy/prometheus\|cmd/meteolink\|docs/SPEC.md' -- docs/ README.md PRODUCT.md DESIGN.md .goreleaser.yaml Makefile deploy/docker-compose.yml`
Expected: no hits in those files. Do not touch `cmd/gates/main.go` or `cmd/gates/main_test.go` — their `cmd/cgo-gui` entries are historical absent-path allowances, not build references.

- [ ] **Step 9: Verify the build and unit tests**

Run: `go build ./cmd/cgo` → exit 0.
Run: `go test ./cmd/cgo ./pkg/... -count=1` → all packages ok.
Run: `git grep 'meteolink.dev' -- cmd/ pkg/ api/` → expected EMPTY. If any hit remains, fix it with the same replacement approach and re-run.

- [ ] **Step 10: Commit once**

Stage explicit paths only (the five deletions and every modified file listed in this task's Files block), using one `git add <path> <path> …` command. `git add -A` and `git add .` are forbidden.
Run: `git commit -m "chore(reset): purge bloat modules, unify SSoT identity and bind 127.0.0.1"`
Expected: one commit; `git show --stat` lists only Chunk 0 paths. If pre-commit hygiene fails because of a deleted path, STOP and report the exact gate finding — do not add a permit.

---

### Task 3: Part C — verification gate

**Files:** none modified (read-only).

**Interfaces:**
- Consumes: Task 2's commit and Task 1's human report.
- Produces: verbatim output of the six checks, with check 1 reported `BLOCKED` until Task 1 confirms an empty history.

- [ ] **Step 1: Cap-40 history check**

Run: `git log --all --no-textconv --oneline -- book/handoff-livrables-2026-09-02.md`
Expected: EMPTY **only after Task 1**. If not empty, record `BLOCKED — Task 1 incomplete` and do not claim Chunk 0 verified. At plan time this returns two commits (`54d3f62`, `8c67859`); the global `-S'ghp_'` form is deliberately not used — the gate's token regex/selftest and historical docs keep that literal forever. No agent may rewrite history.

- [ ] **Step 2: Bloat deletion check**

Run: `ls -d deploy/nginx deploy/prometheus modules cmd/cgo-gui cmd/meteolink api/synctest_test.go 2>&1`
Expected: every path reports "No such file or directory".

- [ ] **Step 3: Fyne removal check**

Run: `grep -i "fyne" go.mod || echo CLEAN_NO_FYNE`
Expected: `CLEAN_NO_FYNE`.

- [ ] **Step 4: Loopback default check**

Run: `git grep "meteolink.dev" -- cmd/ pkg/ api/ || echo CLEAN_NO_METEOLINK_DEV`
Expected: `CLEAN_NO_METEOLINK_DEV`.

- [ ] **Step 5: Version SSoT check**

Run: `cat VERSION; grep '"version"' web/frontend/package.json; grep -n "2.0.0-rc.1" README.md docs/api.md`
Expected: `2.0.0-rc.1` in all four places.

- [ ] **Step 6: Build and tests**

Run: `go build ./cmd/cgo && go test ./cmd/cgo ./pkg/... -count=1`
Expected: exit 0, all packages ok.

---

### Task 4: Acceptance record

**Files:** none modified.

**Interfaces:**
- Consumes: Task 3's verbatim outputs.
- Produces: a short completion report in SDD workspace only. No `docs/CHECKPOINTS.md` entry (ceiling headroom), no `docs/VERIFY.md` edit.

- [ ] **Step 1: Report state honestly**

Write the six check results with exact output and exit codes. State `BLOCKED` for check 1 if Task 1 has not reported success. State explicitly: "Chunk 0 agent work complete; full Chunk 0 verification awaits the human Cap-40 purge."
- [ ] **Step 2: Handoff**

Report to the coordinator. Do not advance to Chunk 1 until check 1 is empty.
