# CGO LIEN Instrument Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Build the LIEN instrument at `C:\cgo` — Go backend (campagne/qdisc/sondes/métriques/archive, SSE 10 Hz, embedded SPA) + fresh four-panel frontend — deployed and verified on the local VMware Ubuntu VM.

**Architecture:** Single Go module `github.com/Realms4239/cgo`; qdisc behind `TCRunner` interface (fake on host); pure metrics package; frozen-CSV archive; 10 Hz SSE hub; go:embed SPA. Frontend: Vite/React/TS, Zustand low-freq store + mutable live rings for 10 Hz data, ECharts via shared grammar builders. Deploy: ported bash engine with vmrun scan/boot/ensure.

**Tech Stack:** Go 1.25, chi, ECharts 5 (tree-shaken), React 19, Zustand, Vite, Vitest, Playwright (host vs VM URL), VMware vmrun, Git Bash engine.

## Global Constraints

- Trace every feature to `LIEN.md`; nothing else ships (`docs/SPEC.md` §5).
- SSE: unnamed message via `onmessage`, named `backpressure`, `Last-Event-ID`, delta structural fields, **10 Hz**.
- No auth/token anywhere. Mutating UI actions arm-then-fire double-confirm only.
- Presentation mode/chapters/command palette do not exist.
- Host unit tests use fake TCRunner — never real `tc`.
- Frozen CSVs immutable after run close; figures only from frozen CSVs.
- Every task ends green: touched-Go → `go vet ./... && go test ./...`; touched-frontend → `bun run typecheck && bun run build && bun run test` (in `web/frontend`).
- Conventional commits; worktree-first per `AGENTS.md`.

---

### Task M0.1: Repo scaffold

**Files:** Create `go.mod` (module github.com/Realms4239/cgo, go 1.25), `cmd/cgo/main.go` (flag parse `--serve|--audit|...` stubs returning "not implemented" exit 2), `pkg/` empty placeholder packages removed later, `Makefile` (build/test/vet/figures/test-real targets), `web/frontend` via `bun create vite` (react-ts), copy fonts from donor `public/fonts/*` into `web/frontend/public/fonts/`, minimal `src/App.tsx` shell rendering sidebar skeleton + four empty panels, fresh `src/styles/tokens.css` (observatory tokens per SPEC §3.3).

- [ ] Scaffold; `go vet ./...` passes; `cd web/frontend && bun install && bun run typecheck && bun run build` pass
- [ ] Commit `feat(m0): scaffold go instrument and frontend shell`

### Task M0.2: Deploy engine scan/boot/ensure

**Files:** Create `deploy/config.sh` (port donor loader verbatim minus cf/route keys), `deploy/engine.sh` (actions: status/scan/boot/ensure/deploy/logs), `deploy/vm-install.sh` (adapted vm-deploy.sh: project `/home/altfloat/cgo`, no negctl/route-pin envs), `deploy/bootstrap-vm.sh` (port; add node not required note), `deploy/cgo-vm.yaml.example`. Generate machine-local `deploy/cgo-vm.yaml` (untracked).

scan algorithm: `"/c/Program Files (x86)/VMware/VMware Workstation/vmrun.exe" list` → default dirs (`%USERPROFILE%/Documents/Virtual Machines`) → shallow sweep each fixed drive root depth ≤3 for `*.vmx`; `--deep` full find. Match `vm_name` or unique result; write `vmx_path:` back to yaml.

ensure algorithm: ssh -o ConnectTimeout=3 reachable? → done : scan → `vmrun start "$vmx" nogui` → poll ssh 60×5 s → done. Optional `snapshot:` key → `vmrun revertToSnapshot` before start.

- [ ] `bash deploy/engine.sh --action scan --config deploy/cgo-vm.yaml` prints discovered vmx (verify against real PC)
- [ ] `--action ensure` reaches SSH OK state (VM currently running → fast path)
- [ ] Commit `feat(m0): deploy engine with vmrun scan boot ensure`

### Task M1.1: Domain types + qdisc module

**Files:** `pkg/model/types.go` (Profile P1/P2 consts w/ capacity/rtt/loss/jitter; Event{RunID,EventID,Profile,Qdisc,CC,Repetition,Phases,GateStatus}; Gate G0..G7; CSVSchemas as struct tags matching SPEC §2.3 exactly), `pkg/qdisc/qdisc.go` (`type TCRunner interface { Run(args ...string) ([]byte,error) }`, `ExecRunner` shells `tc`, `FakeRunner` records), `Apply(netif string, spec Spec) error` emitting exact tc args for netem+TBF+pfifo_fast/fq_codel/cake, `Reset()`, `qdisc_test.go` (arg-vector golden tests per qdisc × profile; fake runner asserts no real exec).

- [ ] Red→green per function; `go test ./pkg/qdisc -v`
- [ ] Commit `feat(m1): model types and qdisc module`

### Task M1.2: Probes + metrics

**Files:** `pkg/probe/ping.go` (wrap system ping, parse p50/p95/p99/loss; injectable command), `pkg/probe/small.go` (HTTP GET small object timer, size param 4–32 KiB, completion ms series), `pkg/probe/bulk.go` (native TCP receiver counting bytes/s for goodput; sender goroutine pool sized by CC under test is kernel-side — bulk uses iperf3 if present else native TCP flood), `pkg/metrics/metrics.go` (Percentile(sorted,p), MedianIQR, DeadlineOK(series,dueMs), WastedBytes(retrans*segSize est), CostAR(wasted)), all pure, `*_test.go` table-driven incl. cost formula golden `wasted=4.5GiB ⇒ 30000 Ar/h`.

- [ ] `go test ./pkg/probe ./pkg/metrics`
- [ ] Commit `feat(m1): probes and pure metrics`

### Task M1.3: Campaign orchestrator (single event)

**Files:** `pkg/campaign/event.go` state machine baseline(30s)→charge(120s)→récup(30s) with tick 100 ms driving probes; gate evaluators G0–G7 per SPEC §2.2 thresholds (G6: baseline p95 jitter <10%; G7: cpu>90% flag); `pkg/campaign/writer.go` append row buffer → freeze `aqm_eval.csv` + manifest.json (sha256 of csv+config); `campaign_test.go` fake clock + FakeTC + httptest small-object target completes event <2 s virtual.

- [ ] Tests green incl. quarantine path (force G1 fail → gate_status=invalid)
- [ ] Commit `feat(m1): single-event campaign orchestration`

### Task M1.4: SSE hub + REST + embed

**Files:** `pkg/api/server.go` chi routes per SPEC §2.5; `pkg/api/sse.go` hub: 10 Hz ticker broadcasting latest snapshot JSON (unnamed), delta-compress structural fields vs last frame, named backpressure on slow client, Last-Event-ID ring(2048); `cmd/cgo/main.go` wire `--serve` :9090 with `//go:embed web/frontend/dist`; graceful stop persists resume marker.

- [ ] Integration test: httptest subscribe, assert 10 frames/sec ±20%, structural omission when unchanged
- [ ] `go test ./pkg/api ./cmd/cgo`; commit `feat(m1): sse hub rest surface embedded spa`

### Task M1.5: Frontend live core + Campagne/Temps réel

**Files:** `src/lib/types.ts` (SSE frame mirror), `src/lib/live.ts` (mutable rings 4096 cap), `src/lib/sse.ts`+worker (donor pattern, new frame), `src/store/ui.ts` (panel 1-4, structural state), `src/lib/chartGrammar.ts` (+tests: units/tooltips/no-animation-under-reduced-motion), `src/components/Sidebar.tsx` (nav+status+gates strip), `src/views/CampagneView.tsx` (selectors, phase timeline, arm-confirm Button component), `src/views/LiveView.tsx` (RTT traces + charge rail shading + small p95 + goodput area + drops bars, captions/provenance, LIVE/STALE banner), styles per tokens.

- [ ] Vitest: grammar tests, arm-confirm resets on blur, gates strip renders 8 states
- [ ] trio green; commit `feat(m1): live panels wired to 10hz stream`

### Task M1.6: Vertical slice on VM

- [ ] Cross-compile `cgo-linux`; `engine.sh --action deploy`; VM smoke: `./cgo-linux --serve &`, curl /api/state, subscribe 3 s stream
- [ ] Playwright smoke spec (host→VM url): sidebar nav, Campagne renders, Live receives ≥10 frames in 2 s
- [ ] Commit `test(m1): vertical slice verification on vm`

### Task M2.1–M2.3: Matrix + Résultats + Intégrité

M2.1 `pkg/campaign/matrix.go` iterate profiles×qdiscs×cc×reps (36/18 via flag), sequential events, resume marker skip-done, run dir freeze; test reduced-matrix order + resume. Commit `feat(m2): matrix runner with resume`.

M2.2 `GET /api/results` server-side medians/IQR grouping; `RésultatsView.tsx` table + best-config highlight (client picks metric among existing columns only) + quarantine list + Export report button hitting `/api/report/export`. Vitest render rows from fixture CSV. Commit `feat(m2): results panel and report export`.

M2.3 `pkg/figures/svg.go` static SVG small-p95 bar chart + latency/goodput scatter from frozen CSV (text/template, zero deps); `make figures`; `IntégritéView.tsx`: manifest chain, counts, Verify btn (calls CLI-equivalent endpoint POST /api/figures/regen + GET integrity), Replay picker → `/api/replay/stream` feeding same rings with REPLAY banner. Commit `feat(m2): integrity panel svg figures replay`.

### Task M3: Audit CLI + import

`./cgo audit` runs ping/small/throughput loaded/unloaded sequence writing link_audit.csv; `POST /api/profile/import` accepts json/csv profile → validated into profiles registry file. Tests with injected probe fakes. Commit `feat(m3): link audit cli and profile import`.

### Task M4: Design authority + Impeccable pass

Rewrite `DESIGN.md` as observatory system authority (SPEC §3.3 content expanded). Run detector once over `src/`, fix actionable findings, one screenshot batch desktop/mobile vs VM app, fix batch, confirm batch, stop. README + RUNBOOK (deploy/verify flows). Commit `docs(m4): observatory design authority and runbook`.

## Verification Ledger (per milestone close)

```text
host:   go vet ./... && go test ./...
        cd web/frontend && bun run typecheck && bun run build && bun run test
vm:     bash deploy/engine.sh --action ensure
        bash deploy/engine.sh --action deploy
        ssh altfloat@192.168.174.128 'cd ~/cgo && make test-real'
        playwright: npx playwright test --project=chromium (baseURL=http://192.168.174.128:9090)
```
