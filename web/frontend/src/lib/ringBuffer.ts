// ringBuffer — anneau Float64Array pour les séries du banc.
// NaN dans la valeur = slot non mesuré (jamais un 0 fabriqué, SPEC §2.2).
// ts strictement croissants côté producteur ; hors bornes/ordres non triés —
// l'anneau n'évince qu'en périphérie (LRU), jamais au milieu.
export class FloatRing {
  readonly capacity: number
  private readonly ts: Float64Array
  private readonly value: Float64Array
  private head = 0 // index du plus ancien échantillon
  private count = 0

  constructor(capacity = 1800) {
    if (!Number.isFinite(capacity) || capacity < 1 || capacity % 1 !== 0) {
      throw new RangeError(`ringBuffer: capacity invalide (${capacity})`)
    }
    this.capacity = capacity
    this.ts = new Float64Array(capacity)
    this.value = new Float64Array(capacity)
  }

  /** pousse l'échantillon (v NaN = non mesuré). O(1). */
  push(ts: number, v: number): void {
    const i = (this.head + this.count) % this.capacity
    this.ts[i] = ts
    this.value[i] = v
    if (this.count < this.capacity) {
      this.count += 1
    } else {
      this.head = (this.head + 1) % this.capacity
    }
  }

  get length(): number { return this.count }

  /** i relatif — Array.prototype.at : négatif = depuis la fin. */
  at(i: number): [number, number] | undefined {
    if (i < 0) i = this.count + i
    if (i < 0 || i >= this.count) return undefined
    const j = (this.head + i) % this.capacity
    return [this.ts[j], this.value[j]]
  }

  latest(): [number, number] | undefined {
    return this.count === 0 ? undefined : this.at(this.count - 1)
  }

  /** n derniers échantillons, du plus ancien au plus récent. */
  tail(n: number): [number, number][] {
    const take = Math.min(Math.max(n, 0), this.count)
    const out: [number, number][] = new Array(take)
    const start = this.count - take
    for (let i = 0; i < take; i += 1) {
      const j = (this.head + start + i) % this.capacity
      out[i] = [this.ts[j], this.value[j]]
    }
    return out
  }

  slice(start?: number, end?: number): [number, number][] {
    const n = this.count
    let a = start ?? 0
    let b = end ?? n
    if (a < 0) a = Math.max(0, n + a); else a = Math.min(n, a)
    if (b < 0) b = Math.max(0, n + b); else b = Math.min(n, b)
    if (a >= b) return []
    const out: [number, number][] = []
    for (let i = a; i < b; i += 1) out.push(this.at(i) as [number, number])
    return out
  }

  /** true si v n'est pas NaN — le « vu mesuré » sans rien inventer. */
  measured(i: number): boolean {
    const j = (this.head + i) % this.capacity
    return Number.isFinite(this.value[j])
  }

  since(tsLo: number): [number, number][] {
    const out: [number, number][] = []
    for (let i = 0; i < this.count; i += 1) {
      const j = (this.head + i) % this.capacity
      if (this.ts[j] >= tsLo) out.push([this.ts[j], this.value[j]])
    }
    return out
  }

  reset(): void {
    this.head = 0
    this.count = 0
    this.ts.fill(0)
    this.value.fill(NaN)
  }
}

/** vue itérable comme un tableau (pour .map(([, v]) => …)). */
export function entries(r: FloatRing): [number, number][] {
  const out: [number, number][] = new Array(r.length)
  for (let i = 0; i < r.length; i += 1) out[i] = r.at(i) as [number, number]
  return out
}

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

export interface DomainEvent {
  id: number
  tsMonoMs: number
  kind: string
  severity: string
  title: string
  detail: string
}

/** DomainEvent — 100 créneaux fixes ; l'éviction retire le plus ancien en périphérie. */
export class DomainEventRing {
  readonly capacity = 100
  private readonly slots: (DomainEvent | undefined)[]
  private head = 0
  private count = 0

  constructor() {
    this.slots = new Array<DomainEvent | undefined>(this.capacity)
  }

  /** pousse l'événement ; O(1), l'ancienneté est préservée. */
  push(e: DomainEvent): void {
    const i = (this.head + this.count) % this.capacity
    this.slots[i] = e
    if (this.count < this.capacity) this.count += 1
    else this.head = (this.head + 1) % this.capacity
  }

  get length(): number { return this.count }

  /** i relatif — négatif = depuis la fin. */
  at(i: number): DomainEvent | undefined {
    const k = i < 0 ? this.count + i : i
    if (k < 0 || k >= this.count) return undefined
    return this.slots[(this.head + k) % this.capacity]
  }

  /** n derniers événements, du plus ancien au plus récent. */
  tail(n: number): DomainEvent[] {
    const take = Math.min(Math.max(n, 0), this.count)
    const out: DomainEvent[] = new Array(take)
    for (let i = 0; i < take; i += 1) out[i] = this.at(this.count - take + i) as DomainEvent
    return out
  }

  reset(): void {
    this.head = 0
    this.count = 0
    this.slots.fill(undefined)
  }
}
