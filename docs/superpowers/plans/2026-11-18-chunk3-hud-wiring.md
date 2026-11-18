# CGO / Meteolink Chunk 3D — SSE-to-Ring HUD Wiring Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Feed 20 Hz SSE ticks directly into the frontend rings without React re-renders, feed milestone events into `DomainEventRing`, and render a `RingBufferHUD` whose values update in `requestAnimationFrame` only when `ring.dirty`.

**Architecture:** A new `lib/liveRings.ts` owns singletons (`TypedMetricRing`, `DomainEventRing`) and two pure push functions. `lib/sse.ts` calls them from `onmessage` and from named milestone listeners — no Zustand write per tick. `components/RingBufferHUD.tsx` reads the singletons in a rAF loop and writes text into DOM refs (no `setState` per frame), calling `markPainted()` after each paint.

**Tech Stack:** TypeScript, React 19 refs/effects, Vitest + happy-dom for the DOM test.

**Source:** `docs/MUSTREAD/13-chunk-3-realtime-pipeline.md` §2–3 (verbatim corpus `CGO5.md` 29–49). Chunk 3 is decomposed; this is sub-plan D.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, frozen corpus/schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, live bench, remote hosts.
- Existing `sse.ts` behaviour (delta retention, coalesced store updates every 5 frames, backpressure status, replay pause, connect/disconnect lifecycle) must be preserved — only ADD ring pushes and event listeners.
- Missing telemetry stays `NaN` in the ring and renders as `—`, never `0`.
- French user-facing text and comments; English tests.
- One commit, exact message: `feat(realtime): feed SSE ticks into rings and add RingBufferHUD`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: Singleton rings and pure push functions

**Files:**
- Create: `web/frontend/src/lib/liveRings.ts`
- Test: `web/frontend/src/lib/liveRings.test.ts` (create)

**Interfaces:**
- Consumes: `TypedMetricRing`, `DomainEventRing` from `./ringBuffer`; `LiveFrame` from `./types`.
- Produces: `metricRing`, `eventRing`, `MILESTONE_KINDS`, `pushLiveFrame(f: Partial<LiveFrame>): void`, `pushLiveEvent(kind: string, detail: Record<string, unknown>): void`, `resetLiveRings(): void`.

- [ ] **Step 1: Write the failing test**

```ts
import { beforeEach, describe, expect, it } from 'vitest'
import { eventRing, metricRing, pushLiveEvent, pushLiveFrame, resetLiveRings } from './liveRings'

describe('liveRings', () => {
  beforeEach(() => resetLiveRings())

  it('maps live ticks into the typed ring without inventing zeros', () => {
    pushLiveFrame({ ts: 1, rtt_p95_ms: 120, bulk_goodput_mbps: 9.4, drops: 0 })
    const last = metricRing.at(-1)
    expect(last?.rttMs).toBe(120)
    expect(last?.throughputMbps).toBe(9.4)
    expect(last?.dropsBytes).toBe(0)
    expect(metricRing.dirty).toBe(true)
    pushLiveFrame({ ts: 2 })
    expect(Number.isNaN(metricRing.at(-1)?.rttMs)).toBe(true)
  })

  it('stores milestone events in the domain ring', () => {
    pushLiveEvent('SLA_BREACH_220MS', { rtt_ms: 260 })
    expect(eventRing.at(-1)?.kind).toBe('SLA_BREACH_220MS')
    expect(eventRing.at(-1)?.title).toContain('220')
  })
})
```

- [ ] **Step 2: Confirm FAIL** — targeted vitest → FAIL (module missing).

- [ ] **Step 3: Implement** `liveRings.ts`:

```ts
import { DomainEventRing, TypedMetricRing } from './ringBuffer'
import type { LiveFrame } from './types'

export const MILESTONE_KINDS = [
  'CALIBRATION_BASELINE', 'BULK_SATURATION_START', 'SLA_BREACH_220MS',
  'QDISC_DROP_BURST', 'AQM_STABILIZED', 'DUAL_SEAL_COMMITTED',
] as const

export const metricRing = new TypedMetricRing(1800)
export const eventRing = new DomainEventRing()
let eventID = 0

export function pushLiveFrame(f: Partial<LiveFrame>): void {
  metricRing.push({
    ts: f.ts ?? 0,
    throughputMbps: f.bulk_goodput_mbps ?? null,
    rttMs: f.rtt_p95_ms ?? null,
    dropsBytes: f.drops ?? null,
  })
}

export function pushLiveEvent(kind: string, detail: Record<string, unknown>): void {
  eventID += 1
  eventRing.push({
    id: eventID,
    tsMonoMs: typeof performance !== 'undefined' ? performance.now() : 0,
    kind,
    severity: kind === 'SLA_BREACH_220MS' || kind === 'QDISC_DROP_BURST' ? 'error' : 'info',
    title: kind.replaceAll('_', ' '),
    detail: JSON.stringify(detail),
  })
}

export function resetLiveRings(): void {
  metricRing.reset()
  eventRing.reset()
  eventID = 0
}
```

- [ ] **Step 4: Verify** — targeted PASS; full frontend suite green.

---

### Task 2: Wire SSE and add the HUD component

**Files:**
- Modify: `web/frontend/src/lib/sse.ts`
- Create: `web/frontend/src/components/RingBufferHUD.tsx`
- Test: `web/frontend/src/components/RingBufferHUD.dom.test.tsx` (create, happy-dom)

**Interfaces:** consumes Task 1 exports.

- [ ] **Step 1: Add the SSE bridge**

In `sse.ts`, import `{ pushLiveFrame, pushLiveEvent, MILESTONE_KINDS }` from `./liveRings`. Inside `es.onmessage`, after the delta-fill loop and before the coalesced store schedule, call `pushLiveFrame(data)`. After creating `es`, register one listener per milestone kind:

```ts
  for (const kind of MILESTONE_KINDS) {
    es.addEventListener(kind, (ev) => {
      try { pushLiveEvent(kind, JSON.parse((ev as MessageEvent).data)) }
      catch { /* événement illisible : ignoré, jamais un 0 fabriqué */ }
    })
  }
```

Do not remove or alter any existing store/coalescer/backpressure/connect logic.

- [ ] **Step 2: Add `RingBufferHUD.tsx`**

A functional component with `useEffect` starting a `requestAnimationFrame` loop. It reads `metricRing.at(-1)` and `eventRing.at(-1)`, writes text into `useRef<HTMLSpanElement>` nodes via `textContent` (format `—` when `Number.isNaN`), calls `metricRing.markPainted()` after each paint, and cancels the frame on unmount. Render: throughput (`Mbit/s`), WIS 2.0 RTT (`ms`), eBPF drops, a status badge (`PROTÉGÉ` / `CONGESTION > 220 ms`), and the last domain event title. No `useState` updates inside the rAF loop.

- [ ] **Step 3: DOM test**

Create `RingBufferHUD.dom.test.tsx` with `@vitest-environment happy-dom`: render the HUD, push a frame with `rtt_p95_ms: 260`, advance one animation frame (or call the exported `paintHUD` helper directly if rAF timing is awkward), and assert the RTT text contains `260` — plus an absent frame renders `—`. Follow the pattern of existing `*.dom.test.tsx` files in this repo.

- [ ] **Step 4: Verify** — targeted DOM test PASS; full `npx vitest run` → 0; `./node_modules/.bin/tsc --noEmit` → 0.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `cd web/frontend && npx vitest run` → 0; `./node_modules/.bin/tsc --noEmit` → 0.
- [ ] **Step 2: Static proof** — `git grep -n 'pushLiveFrame' web/frontend/src/lib/sse.ts` ≥1; `git grep -n 'markPainted' web/frontend/src/components/RingBufferHUD.tsx` ≥1; `git grep -n 'useState' web/frontend/src/components/RingBufferHUD.tsx` → EMPTY.
- [ ] **Step 3: Commit once** — stage only the five named files by explicit path; message `feat(realtime): feed SSE ticks into rings and add RingBufferHUD`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs/exit codes to the SDD report; record that React render-isolation at 20 Hz is structural (no setState in rAF) and was not measured with a browser trace. No `docs/CHECKPOINTS.md` append.
