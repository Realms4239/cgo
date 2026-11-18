# CGO / Meteolink Chunk 3C — Frontend Zero-GC Rings Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a four-series `TypedMetricRing` backed by `Float64Array` with NaN-for-absent semantics and a `dirty` flag, plus a 100-slot `DomainEventRing`, without removing the existing `FloatRing`.

**Architecture:** Both classes live in `web/frontend/src/lib/ringBuffer.ts`, alongside `FloatRing`. `TypedMetricRing` keeps four parallel preallocated arrays (`ts`, `throughputMbps`, `rttMs`, `dropsBytes`), an O(1) modulo head, and a `dirty` boolean that `rAF` consumers can check before repainting. `DomainEventRing` stores `{id, tsMonoMs, kind, severity, title, detail}` in a fixed 100-slot circular array.

**Tech Stack:** TypeScript, Vitest (`vitest run`), `Float64Array`.

**Source:** `docs/MUSTREAD/13-chunk-3-realtime-pipeline.md` §2 (verbatim corpus `CGO5.md` 29–36). Chunk 3 is decomposed; this is sub-plan C1.

## Global Constraints

- MUSTREAD governs over `AGENTS.md`; operator bans bind. No `git add -A`/`git add .`, no `--no-verify`, no history rewrite, no force-push.
- Never touch `book/`, frozen corpus/schemas, `docs/CHECKPOINTS.md`, `docs/VERIFY.md`, CI, Dockerfile, live bench, remote hosts.
- `FloatRing` and `entries()` are existing public API — do not modify or remove them.
- Missing telemetry MUST be stored as `NaN`, never `0`.
- French comments in source; English test names.
- One commit, exact message: `feat(realtime): add four-series typed metric ring and domain event ring`.
- Extensive pre-existing WIP is protected; stage only files named here.

---

### Task 1: `TypedMetricRing` with NaN honesty and dirty flag

**Files:**
- Modify: `web/frontend/src/lib/ringBuffer.ts`
- Test: `web/frontend/src/lib/ringBufferTyped.test.ts` (create)

**Interfaces:**
- Consumes: existing file conventions (`FloatRing` capacity default 1800 = 90 s @ 20 Hz).
- Produces: `class TypedMetricRing` with `capacity`, `dirty`, `length`, `push(sample)`, `at(i)`, `tail(n)`, `measured(i)`, `markPainted()`, `reset()`.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/lib/ringBufferTyped.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { TypedMetricRing } from './ringBuffer'

describe('TypedMetricRing', () => {
  it('stores absent metrics as NaN, never zero', () => {
    const r = new TypedMetricRing(4)
    r.push({ ts: 1, throughputMbps: NaN, rttMs: 42, dropsBytes: null })
    expect(Number.isNaN(r.at(0)?.throughputMbps)).toBe(true)
    expect(Number.isNaN(r.at(0)?.dropsBytes)).toBe(true)
    expect(r.at(0)?.rttMs).toBe(42)
    expect(r.measured(0, 'rttMs')).toBe(true)
    expect(r.measured(0, 'throughputMbps')).toBe(false)
  })

  it('wraps in O(1) and keeps the last capacity samples', () => {
    const r = new TypedMetricRing(3)
    for (let i = 1; i <= 5; i += 1) r.push({ ts: i, throughputMbps: i, rttMs: i, dropsBytes: i })
    expect(r.length).toBe(3)
    expect(r.at(0)?.ts).toBe(3)
    expect(r.at(-1)?.ts).toBe(5)
  })

  it('tracks dirty until the painter acknowledges', () => {
    const r = new TypedMetricRing(2)
    expect(r.dirty).toBe(false)
    r.push({ ts: 1, throughputMbps: 1, rttMs: 2, dropsBytes: 0 })
    expect(r.dirty).toBe(true)
    r.markPainted()
    expect(r.dirty).toBe(false)
  })
})
```

Use the exact public shape you implement; `push` accepts `dropsBytes: number | null` and normalizes `null`/`undefined` to `NaN`.

- [ ] **Step 2: Confirm FAIL** — `cd web/frontend && npx vitest run src/lib/ringBufferTyped.test.ts` → FAIL (`TypedMetricRing` not exported).

- [ ] **Step 3: Implement**

Append to `web/frontend/src/lib/ringBuffer.ts`:

```ts
export interface MetricSample {
  ts: number
  throughputMbps: number | null
  rttMs: number | null
  dropsBytes: number | null
}

/** MetricSample — quatre séries préallouées ; NaN = absence, jamais un 0 inventé. */
export class TypedMetricRing {
  readonly capacity: number
  private readonly ts: Float64Array
  private readonly throughputMbps: Float64Array
  private readonly rttMs: Float64Array
  private readonly dropsBytes: Float64Array
  private head = 0
  private count = 0
  dirty = false

  constructor(capacity = 1800) {
    if (!Number.isFinite(capacity) || capacity < 1 || capacity % 1 !== 0) {
      throw new RangeError(`ringBuffer: capacity invalide (${capacity})`)
    }
    this.capacity = capacity
    this.ts = new Float64Array(capacity)
    this.throughputMbps = new Float64Array(capacity)
    this.rttMs = new Float64Array(capacity)
    this.dropsBytes = new Float64Array(capacity)
    this.reset()
  }

  push(s: MetricSample): void {
    const i = (this.head + this.count) % this.capacity
    this.ts[i] = s.ts
    this.throughputMbps[i] = s.throughputMbps ?? NaN
    this.rttMs[i] = s.rttMs ?? NaN
    this.dropsBytes[i] = s.dropsBytes ?? NaN
    if (this.count < this.capacity) this.count += 1
    else this.head = (this.head + 1) % this.capacity
    this.dirty = true
  }

  get length(): number { return this.count }

  at(i: number): MetricSample | undefined {
    const k = i < 0 ? this.count + i : i
    if (k < 0 || k >= this.count) return undefined
    const j = (this.head + k) % this.capacity
    return {
      ts: this.ts[j],
      throughputMbps: this.throughputMbps[j],
      rttMs: this.rttMs[j],
      dropsBytes: this.dropsBytes[j],
    }
  }

  tail(n: number): MetricSample[] {
    const take = Math.min(Math.max(n, 0), this.count)
    const out: MetricSample[] = new Array(take)
    for (let i = 0; i < take; i += 1) out[i] = this.at(this.count - take + i) as MetricSample
    return out
  }

  measured(i: number, key: 'throughputMbps' | 'rttMs' | 'dropsBytes'): boolean {
    const k = i < 0 ? this.count + i : i
    if (k < 0 || k >= this.count) return false
    const j = (this.head + k) % this.capacity
    return Number.isFinite(this[key][j])
  }

  markPainted(): void { this.dirty = false }

  reset(): void {
    this.head = 0
    this.count = 0
    this.ts.fill(0)
    this.throughputMbps.fill(NaN)
    this.rttMs.fill(NaN)
    this.dropsBytes.fill(NaN)
    this.dirty = false
  }
}
```

- [ ] **Step 4: Verify** — targeted vitest PASS; full frontend suite `cd web/frontend && npx vitest run` exit 0.

---

### Task 2: `DomainEventRing` (100 slots)

**Files:**
- Modify: `web/frontend/src/lib/ringBuffer.ts`
- Test: `web/frontend/src/lib/ringBufferEvents.test.ts` (create)

**Interfaces:** consumes `TypedMetricRing` file conventions; produces `class DomainEventRing`.

- [ ] **Step 1: Write the failing test**

```ts
import { describe, expect, it } from 'vitest'
import { DomainEventRing } from './ringBuffer'

describe('DomainEventRing', () => {
  it('wraps at 100 slots, oldest first', () => {
    const r = new DomainEventRing()
    for (let i = 1; i <= 102; i += 1) r.push({ id: i, tsMonoMs: i, kind: 'SLA_BREACH_220MS', severity: 'warn', title: `e${i}`, detail: '' })
    expect(r.length).toBe(100)
    expect(r.at(0)?.id).toBe(3)
    expect(r.at(-1)?.id).toBe(102)
  })
})
```

- [ ] **Step 2: Confirm FAIL** — targeted vitest → FAIL (`DomainEventRing` not exported).
- [ ] **Step 3: Implement** a fixed 100-slot array with `head`/`count`, `push(e)`, `at(i)` (negative from end), `tail(n)`, `reset()`. Preserve insertion order oldest-first.
- [ ] **Step 4: Verify** — targeted + full frontend suite green.

---

### Task 3: Verification gate and commit

**Files:** none new.

**Interfaces:** consumes Tasks 1–2.

- [ ] **Step 1: Suites** — `cd web/frontend && npx vitest run` → exit 0; `npx tsc --noEmit` → exit 0.
- [ ] **Step 2: Static proof** — `git grep -n 'TypedMetricRing\|DomainEventRing' web/frontend/src/lib/ringBuffer.ts` ≥2 hits; `git grep -n 'null ?? NaN\|?? NaN' web/frontend/src/lib/ringBuffer.ts` ≥1.
- [ ] **Step 3: Commit once** — stage only `web/frontend/src/lib/ringBuffer.ts` and the two new test files by explicit path; message `feat(realtime): add four-series typed metric ring and domain event ring`. If a gate fails, STOP and report.
- [ ] **Step 4: Record** — verbatim outputs and exit codes to the SDD report; explicitly record that React/rAF integration and real milestone producers remain 3C follow-ups. No `docs/CHECKPOINTS.md` append.
