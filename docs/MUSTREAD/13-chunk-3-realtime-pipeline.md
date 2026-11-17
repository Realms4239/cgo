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
Proceeding to Chunk 3 (Triple-Ring Real-Time Pipeline: Kernel 6.8 eBPF Ringbuf
-> 20 Hz Go SSE Ring -> Float64Array TypedMetricRing + RingBufferHUD +
Clean 2-Mode UI & Exact Pins).

Hand Part A to your agent, then run Part B yourself before moving to Chunk 4
(Tests).

CHUNK 3 — Real-Time Triple-Ring Pipeline, RingBufferHUD, 2-Mode Cleanup & Pins

PART A — Exact Bounded Prompt for Your Coding Agent

# CHUNK 3 EXECUTION TASK — TRIPLE-RING PIPELINE & STREAMLINED 2-MODE UI (STRICTLY BOUNDED)

You are executing ONLY Chunk 3 on `C:\cgo`. Do NOT modify auth (`api/auth.go`), SSRF (`ssrf.go`), `Dockerfile`, CI workflows, or rewrite tests yet (Chunk 4 handles tests).

## 1. Kernel 6.8 eBPF Ring Buffer & `20 Hz` Go Priority SSE Ring
Files: `pkg/probe/bpf/`, `pkg/audit/`, `api/sse.go`, `api/server.go`
1. **eBPF `BPF_MAP_TYPE_RINGBUF` & Target Filter (`pkg/probe/bpf/`):**
   - Ensure `BPF_MAP_TYPE_RINGBUF` emits structured kernel events (`ts_ns`, `kind`, `ifindex`, `skb_len`, `srtt_us`, `cwnd`) from `tracepoint/skb/kfree_skb` (`ctx->reason == SKB_DROP_REASON_QDISC_DROP`) and `raw_tp/tcp_retransmit_skb`, scoped cleanly via the target filter map (`target_ifindex` / `target_dport`) so SSH/Tailscale traffic is excluded.
   - Ensure `fentry/tcp_rcv_established` vs `kprobe/tcp_rcv_established` attach mutually exclusively (never both).
   - If you recompile the `.o` object, immediately update its `.sha256` sidecar and `docs/verification/2026-11-12-ebpf-probe-bringup/README.md` hash in the same commit!
2. **`20 Hz` O(1) Priority Circular SSE Ring (`api/sse.go`, `pkg/audit/`):**
   - Upgrade `api/sse.go` from `10 Hz` + `2048` slice to a fixed `[4096]StreamFrame` O(1) circular ring (`head`, `nextSeq`, `Last-Event-ID` resync, `maxSubs = 64`).
   - Emit **both** live Audits and Bench runs onto `/api/stream` using two frame types:
     - `type: "tick"` (`20 Hz` = every `50 ms`): `{seq, ts_mono_ms, phase, mbps_down, mbps_up, rtt_ms, ebpf_drop_bytes, sla_ok_pct}` — derived from atomic `io.Reader`/`Writer` byte counters + monotonic `time.Since` (never `0` for absent values: use `null`!). Coalesces/drops stale ticks on slow client channels (`cap 32`) while counting `backpressure_drops`.
     - `type: "event"` (Guaranteed milestone frames, NEVER dropped ahead of ticks): `CALIBRATION_BASELINE`, `BULK_SATURATION_START`, `SLA_BREACH_220MS`, `QDISC_DROP_BURST`, `AQM_STABILIZED`, `DUAL_SEAL_COMMITTED`.

## 2. Frontend Zero-GC `Float64Array` Ring Buffers & `60 FPS` Canvas Pipeline
Files: `web/frontend/src/lib/ringBuffer.ts` (new), `web/frontend/src/lib/sse.ts`, `web/frontend/src/store/audit.ts`, chart components
1. **Create `src/lib/ringBuffer.ts`:**
   - Implement `TypedMetricRing` (default `capacity = 1800` = 90 s @ 20 Hz) backed by `Float64Array` (`ts`, `throughputMbps`, `rttMs`, `dropsBytes`), O(1) `head` modulo pointer, `dirty` boolean flag. Missing/null inputs MUST write `NaN` (NEVER `0`!).
   - Implement `DomainEventRing` (`capacity = 100`, O(1) circular log) storing `{id, tsMonoMs, kind, severity, title, detail}`.
2. **Bypass React Virtual DOM on `20 Hz` Ticks (`sse.ts`, `store/audit.ts`):**
   - `sse.ts` writes incoming `tick` frames straight into the singleton `TypedMetricRing` **without triggering React/Zustand re-renders**; sync numeric KPIs to Zustand at `4 Hz` (`250 ms`) with `tabular-nums`, and push `event` frames immediately into `DomainEventRing`.
   - When SSE is connected during an active audit, eliminate the sluggish `2000 ms` `/api/audit/live` poll (`store/audit.ts:88-89`), keeping polling strictly as a degraded reconnect fallback.
   - Configure ECharts strictly with `{ renderer: 'canvas' }` and `animation: false`, updating series inside `requestAnimationFrame` ONLY when `ring.dirty === true`.

## 3. `RingBufferHUD.tsx` & Clean 2-Mode UI Polish
Files: `web/frontend/src/components/RingBufferHUD.tsx` (new), `Cockpit` & `Labo` views/panels, `ArchiveTable.tsx:48`, `AuditDeclare`
1. **Create `RingBufferHUD.tsx`:**
   - Crisp radial SVG ring gauge showing circular buffer fill (`count / capacity`), live `head` pulse (respecting `@media (prefers-reduced-motion: reduce)`), `tabular-nums` throughput (`Mbit/s`) + WIS 2.0 RTT (`ms`), explicit non-color status text badge (`PROTÉGÉ`, `CONGESTION > 220 ms`, `HORS LIGNE`), eBPF drop bytes (`skb->len`), SSE drop counter (`0 perte`), and the scrollable `DomainEventRing` live event feed.
   - Pair alongside the `60 FPS` ECharts Canvas waveform featuring a crisp horizontal `220 ms` threshold line (`Seuil Alerte WIS 2.0 / CAP`) + a collapsible `<details>` accessible summary table below the chart.
2. **Enforce Streamlined 2-Mode Discipline (`Cockpit` + `Labo`):**
   - **`Cockpit`:** 1 welcoming CTA empty state (no `"Pas encore mesuré"` wall) → 60 s (`10/40/10`) audit launcher → 3 verdict cards (`Protection WIS 2.0 < 220 ms`, `Latence sous charge`, `Capacité Gaspillée Ar/h` using Yas GOLD `1,388.89 Ar/h` or `"Mesure absente"` `—`) → `RingBufferHUD` + waveform → Remediation card shown ONLY when `HasValidRemediation` holds, wired to `ArmButton` + `X-Confirm-Action: true`.
   - **`Labo`:** Full unclipped `P1–P4` + `DGM Twin` tiles (no CSS `text-overflow: ellipsis`!), explicit Canal (`Statique` vs `Empirique .dist`) and Direction (`Montant Egress` vs `Descendant Ingress IFB` vs `Les deux`) controls, `220 ms` default SLA, before/after delta, and Seal/Integrity panel backed by Chunk 2's real `/api/integrity` (never displaying green verified badges when files are missing/corrupt).
3. **Fix Specific Inclusivity, Plural & Pin Defects:**
   - `ArchiveTable.tsx:48`: Fix French pluralization cleanly (`0 entrée`, `1 entrée`, `${n} entrées`).
   - `AuditDeclare`: Add `maxLength={80}` and `.normalize('NFC')` on site/provider inputs.
   - Ensure `--t-danger-text` passes `>= 4.5:1` WCAG AA, enforce `min-height: 44px; min-width: 44px` on interactive controls across desktop & mobile, add `@media (forced-colors: active)`, and ship `OFL` font licenses + ECharts `Apache-2.0 NOTICE` in `web/frontend/public/licenses/`.
   - Remove `^` caret ranges in `web/frontend/package.json:17-32`, pinning exact versions matching `web/frontend/bun.lock`.
   - Ensure `tsconfig.json` (or `tsconfig.tools.json`) covers `e2e/` and `scripts/*.mjs`.

## Constraints
- Run `cd web/frontend && npx tsc --noEmit && bun run build` and `go build ./cmd/cgo` before finishing; confirm bundle gz size stays clean.
- Do NOT touch `docs/CHECKPOINTS.md`.
- Commit all Chunk 3 work in ONE clean commit: `feat(realtime): wire eBPF ringbuf to 20Hz SSE ring, Float64Array HUD and 2-mode UI`.

PART B — Your Hard Verification Gate (Run Before Chunk 4)

Run these commands locally after the agent finishes Chunk 3:

# 1. Verify exact pinned versions in package.json (NO ^ carets on deps)
grep -E '"\^' web/frontend/package.json || echo "CLEAN_EXACT_PINS"

# 2. Verify Float64Array TypedMetricRing & NaN honesty exist (no zero-fill)
git grep -n "Float64Array" web/frontend/src/lib/ringBuffer.ts
git grep -n "NaN" web/frontend/src/lib/ringBuffer.ts

# 3. Verify SSE 20 Hz (50ms) and 4096 ring exist in Go + X-Confirm-Action wired
git grep -n "4096" api/sse.go
git grep -n "X-Confirm-Action" web/frontend/src/

# 4. Verify OFL & ECharts Apache-2.0 license files are shipped
ls -la web/frontend/public/licenses/

# 5. Verify eBPF hash matches between binary/sidecar and README
sha256sum pkg/probe/bpf/*.o 2>/dev/null
git grep -n "8612f578\|sha256" docs/verification/2026-11-12-ebpf-probe-bringup/ pkg/probe/bpf/

# 6. Full TS typecheck + frontend build + Go build (bundle must be valid!)
cd web/frontend && npx tsc --noEmit && bun run build && cd ../..
go build ./cmd/cgo

Once every check passes cleanly, say "Proceed with Chunk 4" and we will replace
api/synctest_test.go with blazing-fast real synctest tests, convert Vitest to
node + happy-dom, and add a real unmocked Playwright + @axe-core/playwright live
smoke test! Proceeding to Chunk 4 (Blazing-Fast, Honest Tests: Real Go
testing/synctest, node/happy-dom Frontend Speedup, and Real Unmocked Playwright
E2E + @axe-core/playwright).
