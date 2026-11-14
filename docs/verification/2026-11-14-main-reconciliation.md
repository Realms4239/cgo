# Main reconciliation — audit findings vs `main`

**Date:** 2026-11-14 · **Trunk under review:** `main` @ `3ef0fb5` · **Host:** Europe/Moscow Windows
**Audit under reconciliation:** `critical-review-ground-truth-2026-11-13.md` (34/100, cap
35(ii) governing) + `.slim/deepwork/ground-truth-audit-2026-11-13.md` (node tables),
both HEAD-scoped at `operator-cockpit` @ `5ba9434`.

**Rule of this document.** No finding is marked closed without (a) commit evidence on
`main`, (b) a command re-run by the controller on 2026-11-14, or (c) a named plan task
that owns it. Anything else is **UNCLAIMED** and printed, not silently absorbed.
This document indexes the plan; it does not re-scope it, and it certifies nothing for
production. The frozen 159-run corpus is untouched, byte-identical, and out of scope.

**Legend:** `CLOSED` = fixed on `main` with commit/re-run evidence · `PENDING` =
assigned to a named task in `docs/superpowers/plans/2026-11-14-final-route.md` ·
`BLOCKED` = human gate required · `UNCLAIMED` = no plan row exists (needs a decision) ·
`NO-TRIGGER` = audit confirmed no cap/fabrication breach.

## A. The caps

| Cap | Audit evidence | Status on `main` @ `3ef0fb5` | Owner |
|---|---|---|---|
| **35(ii)** exposure (0.0.0.0, plaintext, no AuthN) | governing cap; `cmd/cgo/main.go` systemd `--tls=false`; `docs/deploy.md` ufw-any; zero AuthN repo-wide at audit time | **OPEN.** One auth *seam* now exists canonically (`api/auth.go`) and the duplicate `guard/` module is deleted (Task 0.4), but no exposure lockdown has shipped. `0.0.0.0` default unchanged. | Phase 1 · WP1.1 (Task 1.1, Task 1.2) |
| **35(i)** clean clone fails | `kit/test-fresh-clone.sh` certified a dead dashboard; README false | **CLOSED (code).** `c4d8325` + `2f6f9ba`: `api/frontend/` deleted, `StaticFS` seam, `ValidateEmbeddedAssets`, `setup.go` needsBuild fix, fresh-clone gate rewritten. **Field proof PENDING:** no `make` on this host; fresh-clone was not re-proven serving real assets 200 since the fix; CI re-proof assigned to Task 1.3. Tracked `web/frontend/dist/index.html` workaround still present (see F). | Task 0.3 (done, code) → 1.3 (proof) |
| **45** plaintext transport / user data | same fact, transport reading | **OPEN.** Service-path TLS not enforced; wizard path (ECDSA P-256, 0600) untouched and still the only HTTPS-by-default path. | Phase 1 · WP1.1 (Task 1.1) |
| **40** committed live PAT in public history | `54d3f62` → deleted `8c67859`, permanent in ≥7 refs / 984 revs | **BLOCKED (human).** MANDATORY: revoke/rotate first, then filter-repo + force-push + origin GC, or signed written risk acceptance. Nothing has been done to the history. No mirror backup taken. Validity unknown offline by design. | Task 0.5 (human-gated) + 1.3 (remote secret-scan gate) |
| 30 — core value mocked | probes real, I/O real; canned-output claim rejected | **NO-TRIGGER** (audit §1/§16) — unchanged; not re-scored here. | — |
| §2.4 fabrication | none found in Phases 1–3 | **NO-TRIGGER** — unchanged. | — |

**No cap was lifted by anything in this document.** 35(ii) still governs ≤ 35 until
Phase 1 lands; 40 still requires the human gate; 45 still binds the transport reading.

## B. Drift table disposition (audit §13, rows 1–34)

| # | Audit finding (short) | Status | Resolution |
|---|---|---|---|
| 1 | README:73 `go run` SPA setup false (DR-1 family; cap 35(i) input) | CLOSED (code) / PENDING (proof) | Task 0.3 · proof → 1.3 |
| 2 | README headline numbers unsubstantiated (38,2 / 241,6 / JFI 1,00) | PENDING | Phase 6 · README rewrite from v2 corpus (Tasks 6.2/6.4) |
| 3 | README « HTTPS uniquement » false | PENDING | Task 1.1 (code+docs) → Task 6.2 (README) |
| 4 | README:265 `--tls=false` contradiction | PENDING | Task 1.1 (flag semantics) → 6.2 |
| 5 | Release channels: goreleaser cgo/meteolink only; versionless archives; npm unpinned | PENDING | Task 1.3 (CI/CD release pipeline) |
| 6 | « Support Docker » false (∅ Dockerfile at audit time) | PENDING | Task 1.5 (Docker hardening); main already carries an unaudited Dockerfile (SLSA L2 commit `60de44b`) — out of scope here |
| 7 | README TUI « 250 ms » vs 1 s tick (`cmd/cgo/tui.go:128`) | PENDING | Task 6.2 (README correction; one-line fix) |
| 8 | docs/api.md stale (« 1.2.3 », « replis 20/100 », 5 undocumented routes) | PENDING | Task 0.6 (docs/api.md in SSoT) → 6.2/6.4; route registry is already single-source (`GET /api/schema`) |
| 9 | plan README branch stale (« rebuild/plan1-economics-audit ») | PENDING | Task 0.7 Step 1 (plan index row) |
| 10 | architecture.md « dix paquets » vs 13 | PENDING | Task 6.2 (doc truth) |
| 11 | architecture.md P1/P2/P3 vs P4 in types.go | PENDING | Task 6.2 |
| 12 | `docs/SPEC.md` cited, absent (DESIGN.md, sse.go §1) | PENDING | Task 6.2 (doc truth) |
| 13 | npm@**1.2.2** postinstall, no hash check | PENDING | Task 0.6 (`npm/` deleted); published 1.2.2 unverification is historical — unpublishing is a human/release call, not this route |
| 14 | `.goreleaser.yaml:39` versionless archive clobbers | PENDING | Task 1.3 |
| 15 | orphan `.gitattributes:3` / `.gitignore:3-4` | PENDING | Task 0.6 Step 3 |
| 16 | Makefile `test-real -tags=real` selects nothing | PENDING | Task 0.7 Step 3 (drop or implement) |
| 17 | `bpf-check` compares nonexistent `bpf_probe.o`+sha256 | PENDING | Task 0.7 Step 3 + Phase 3 (Task 3.2 builds the object in CI) |
| 18 | `kit/test-fresh-clone.sh` certifies dead dashboard | CLOSED | Task 0.3 (gate rewritten; serves, curls, fails on 404) |
| 19 | `web/frontend/dist/index.html` tracked; assets gitignored/absent | **UNCLAIMED** | `.gitignore:35` re-includes it deliberately; no plan row names it. Needs human decision or one-line plan amendment (likely rides Task 0.6 dead-surface cuts). Do not fix silently. |
| 20 | `check-bundle.mjs` budget dishonest (600/350; comments admit ~343 KB regression passes) | PENDING | Task 0.7 Step 4 (400/350) |
| 21 | deploy.md `ufw allow 9090` / `curl -k https://…` contradictions | PENDING | Task 1.1 (deploy unit + docs fixed together) |
| 22 | Dual export seal asymmetry (HTTP source-only; PRODUCT.md false) | PENDING | Task 2.6 (durability + seal parity) |
| 23 | Class-aware Δp95 false; three Δ definitions coexist (blanket 220 SI gate) | PENDING | Task 2.5 (Results B1–B4 + one verdict) → Task 6.4 |
| 24 | Lexicon gate mocks-only; populated surfaces leak (`p95`, hex SHA, « 738 → 20 ms ») | PENDING | Task 5.6 (WCAG + lexicon + ArmButton) |
| 25 | DESIGN.md SPEC.md refs; radius-0 drift (2/4/8 px observed) | PENDING | Task 6.2 (doc truth) |
| 26 | Thesis title ledger conflict (« Conception… » vs « Architecture d'ordonnancement… ») | **BLOCKED (human)** | Spec locked decision §1: old title rejected; title 2 is the master-worthy form (« …Systèmes Temps Réel » family, Annexe D.2) — applied only with human sign-off (Task 0.6 Step 1) |
| 27 | types.go:2 stale `csvFieldOrder` comment | PENDING | Task 6.x doc/comment pass (trivial) |
| 28 | Kit action count: 4 published counts, none true (33 dispatch cases) | PENDING | Task 6.2 (doc truth) |
| 29 | `kit/engine.sh` « Retiré à v1.2 » shim still shipped | PENDING/UNCLAIMED | No plan row names it. Deletion-test candidate; note for Phase 0/1 sweep or human |
| 30 | `go-i18n/v2` declared in go.mod, zero importers | PENDING | Task 0.6 / Phase 6 dep purge (verify before removal) |
| 31 | tests gate without `-count=1` | PENDING | Task 0.7 Step 2 |
| 32 | IntegriteView.tsx:268 « small_p95 738 → 20 ms » uncited | PENDING | Task 6.4 (Results chapter rewrite; ungrounded number removed or cited) |
| 33 | Brand naming triad (index.html « Meteolink — LIEN » / DESIGN « CGO/LIEN » / PRODUCT « Meteolink ») | PENDING (blocked at sign-off) | Task 0.6 Step 1 — lock name **METEOLINK** + title 2 (human sign-off) |
| 34 | DESIGN.md/sse.go → absent SPEC.md (DR-1 family) | PENDING | Task 6.2 |

## C. Kill list disposition (audit §16)

| ID | Sev | Minimal fix (audit) | Status | Resolution |
|---|---|---|---|---|
| KL-01 | Critical | revoke PAT; rewrite+push or risk acceptance; remote scan gate | **BLOCKED** | Task 0.5 (human); Task 1.3 (remote gate) |
| KL-02 | Critical | loopback+TLS default unit; non-loopback hard refusal; deploy.md | PENDING | Task 1.1 |
| KL-03 | Critical | fresh-clone gate; untrack dist; README | CLOSED (code) / UNCLAIMED (untrack) | Task 0.3 (gate) · #19 above |
| KL-04 | High | dualns scalar half-delay (Q2) | PENDING | Task 2.2 (channel physics) |
| KL-05 | High | corpus recalibration on dualns fail-closed path | PENDING | Task 2.2 + 4.3 (v2 corpus, sidecar quarantine) |
| KL-06 | High | `small_url` StatusCode ≠2xx ⇒ absence; provision obj16.bin | PENDING | Task 2.1 (probe fidelity) |
| KL-07 | High | minimal CI: verify.py + builds + npm build + secret scan | PENDING | Task 1.3 |
| KL-08 | High | SSRF allowlist + egress doc (small_url/bulk_addr) | PENDING | Task 1.2 |
| KL-09 | High | SPKI pin + explicit confirm; uninstall root-store path | PENDING | Task 1.2 |
| KL-10 | High | CHR parse vs real golden transcript (not mock↔mock) | PENDING | Task 4.4 |
| KL-11 | High | recommendation inputs from audit data, not constants | PENDING | Task 2.5 (verdict) / Task 5.5 |
| KL-12 | High | deploy.md ufw/curl protocol consistency | PENDING | Task 1.1 |
| KL-13 | High | seal telemetry in HTTP export or label seal scope | PENDING | Task 2.6 |
| KL-14 | High | one engine Δp95 definition; UI reads it | PENDING | Task 2.5 → 6.4 |
| KL-15 | High | retention policy + purge + DPIA block | PENDING | Task 1.4 |
| KL-16 | Med | populated-surface lexicon gate; ≥44 px targets; danger-text AA; NOTICE/license texts | PENDING | Task 5.5, 5.6 + Phase 6 (license files) |
| KL-17 | Med | `CGO_SHAPE_MODE` export in run-dualns; fasttrack or refuse | PENDING | Task 2.4 |
| KL-18 | Med | phantom gates D2/D3; `-count=1` | PENDING | Task 0.7 Step 2–3 |
| KL-19 | Low | strip/cite README numbers | PENDING | Task 6.2/6.4 |
| KL-20 | Low | MTU probe; IPv6 unsupported doc; CIDR overlap assert | PENDING | Task 2.4 (lab scripts) + 6.2 (doc) |

## D. Audit-record dispositions (node-level, unique items)

- **Corpus is frozen and not audit-fixable here:** 159 runs · 528 rows · direction census
  `{up: 82, down: 25, unrecorded: 421}` · `direction=both` = **0** ·
  `gate_status` valid 208 / degraded 74 / invalid 246 · `jfi_pct` 122/122 encoded 0.0.
  The frozen corpus stays byte-identical (AGENTS.md). A quarantine **sidecar** marks
  legacy-stacked and pre-GE-fix cells without editing CSVs. **Tasks 2.2, 4.3.**
- **Best zero-poisoning:** 25 rows `small_p95_ms == 0.0`, all invalid — candidates exist
  but route to the `>`0-gate. Engine fix + one verdict. **Task 2.5.**
- **P2 dualns delay doubling (Q2, 202 ms on P2-100):** scalar profile miscalibration;
  the lab figure is a defect signature, not a validation. **Task 2.2.**
- **Q4 CAKE recipe destroyed by the engine** (bench `ack-filter nat diffserv4
  triple-isolate` vs engine `cake bandwidth X rtt Y`): instrument measures a different
  AQM than it recommends. **Task 2.3.**
- **E-family leak/fatality in dualns scripts** (no `-e`, no trap; invisibility of ns
  root; leaked ns + listeners). **Task 2.4.**
- **eBPF object/loader absent** (`cilium/ebpf` never imported; Makefile compares
  nonexistent files): Phase 3 builds it (CI, prebuilt object allowed — never compiled on
  the host; AGENTS.md). **Tasks 3.1–3.3** — note `cilium/ebpf` itself is a human-gated
  dependency decision (spec §6).
- **CSV durability:** zero `Sync()`/fsync in any writer; read-modify-write full rewrites
  (audit writer). Kill -9 test + recovery. **Task 2.6.**
- **Ping ordering:** `exec ping` samples SORTED ASCENDING — temporal order destroyed;
  spacing evidence gone. **Task 2.1.**
- **SmallObject:** new transport per sample, never checks StatusCode. **Task 2.1.**
- **`/api/results` uncached O(corpus) scans** (0.32–0.36 s observed). **Task 2.5.**

## E. Gate / verification state observed 2026-11-14

| Check | Result | Note |
|---|---|---|
| `go build ./...` @ `3ef0fb5` (detached worktree) | **exit 0** (2026-11-14, post-C4) | run by controller |
| `go vet ./...` @ `3ef0fb5` (same worktree) | **exit 0** | run by controller |
| `cd api && go test ./... -count=1` | **PASS** (`ok 65.204s`) | re-run after C4 |
| `go test ./... -count=1` (root) | **PASS** | implementer, Task 0.4 |
| `python tools/verify.py --only tests` | **PASS** (1 gate) | Task 0.4, via pre-commit |
| `python tools/verify.py --only links` | **FAIL** (7 refs) | pre-existing at HEAD; `.gitignore:35`, `.gitignore:21-22` untracked `data/`; → Task 0.7 |
| `go test ./...` (no `-count=1`) | permitted today | Task 0.7 Step 2 removes this ambiguity |
| `kit/test-fresh-clone.sh` | **NOT RE-PROVEN** | needs `make`; host- absent; → Task 1.3 CI |
| `gitleaks` full history | **NOT RUN** | → Task 0.5 (after revoke) + Task 1.3 (remote) |
| `-race` module-wide | **UNVERIFIED** | needs Linux/gcc; Windows host out of bounds (no gcc per AGENTS.md) |
| Standing WIP = 75 pre-existing paths | uncommitted, preserved | bench/eBPF work; committed only by the human, or by named task with human's word |
| SDD reviews | 0.3 clean; 0.4 APPROVE_WITH_FINDINGS | findings discharged by controller; nits not adopted |

Related named shortfall: `C:` disk free 4.1 GB vs plan target ≥ 10 GB (Task 0.1
deviation; caches moved to `D:`; nothing deleted beyond caches).

## F. Phase 0 exit — where each item stands

| Exit criterion (plan) | Status |
|---|---|
| Clean clone serves real SPA (assets 200) | CLOSED (code) — **proof pending** (host lacks `make`) |
| Gates green `-count=1` | PENDING — Task 0.7 Step 2 |
| `gitleaks` = 0 all refs | BLOCKED — Task 0.5 (human); remote gate → 1.3 |
| One auth seam | **CLOSED** — Task 0.4 (`api/auth.go` only; `guard/` deleted; review passed) |
| Honest `/metrics` | **CLOSED** — Task 0.4 (`exporter/` wired; facade + hardcoded-0 flags deleted; two doc'd alerts) |
| title 2 + version SSoT | PENDING / blocked at Step 1 sign-off — Task 0.6 |
| Reconciliation map committed | **CLOSED on this commit** — this file |
| Task 0.2 R0 read (LXD state read-only) | BLOCKED — needs fresh R0 token (human + DGM operator) |
| Task 0.5 cap-40 purge | BLOCKED — human revoke + force-push approval |

## G. Divergences and unclaimed items (need a human decision)

1. **`web/frontend/dist/index.html` tracked; `.gitignore:35` re-includes it.** The G6
   poison (clone serves 200 for `/`, hashed chunks 404). No plan row names it. Likely
   belongs to Task 0.6 (dead-surface cuts) but was not briefed there.
2. **`kit/engine.sh` « Retiré à v1.2 »** still shipped (compat shim). No plan row.
3. **`go-i18n/v2`** declared, zero importers. No plan row until a dep-purge pass.
4. **75-path WIP** (`go.mod` cilium/ebpf + x/net/x/text; `cmd/cgo/main.go`
   `attachKernelProbe`; bench/eBPF) intersects Phase 2/3 — must be reconciled with the
   human **before** Phase 2 starts; committing it is the human's call.
5. **`github.com/kr/text v0.2.0 // indirect` leftover** in root go.mod would be dropped
   by `go mod tidy`; deliberately untouched (not Task 0.4 scope; tidy forbidden here).
6. **Disk shortfall** (§E) — accepted deviation, recorded, not green, not waived.
7. **HEAD-vs-main divergence of the audit itself:** audit scored `operator-cockpit`
   @ `5ba9434`; this map reconciles against `main`. The main-only delivery surface
   (`deploy/Dockerfile`, compose, SLSA — `60de44b`) is **not audited** and is scored
   nowhere until an independent pass exists.
8. **`direction=both` latent writer bug** (CLI `--direction both` freezes UP-ONLY rows;
   three disagreeing duration contracts: CLI unbounded, HTTP 10–600 s, UI 60 s).
   No corpus rows exist for it. **Task 2.1/2.2 adjacent; explicit row needed.**
9. **Ping sample order** and **small-object StatusCode** are measurement-validity
   critical and touch the *recorder* path: fix must be proven by tests, not by review.

## H. Provenance and rules of engagement

Sources: `critical-review-ground-truth-2026-11-13.md` (VERDICT + §13 drift + §16 kill
list), `.slim/deepwork/ground-truth-audit-2026-11-13.md` (node tables, corpus census),
git history of `main` (`c5941b1`..`3ef0fb5`), `tools/verify.py` @ HEAD, plan
`docs/superpowers/plans/2026-11-14-final-route.md`, spec
`docs/superpowers/specs/2026-11-14-final-route-design.md`.

This file is a routing index. It does not close the caps, does not freeze or reopen
budgets, does not touch the frozen corpus, and does not authorize push, force-push,
mirror rewrite, or any live bench. Human gates remain: 0.2 R0, 0.5 cap-40,
0.6 title sign-off, Phase 2/3 live bench runs, 4.5 countersign, 6 merge/release.
