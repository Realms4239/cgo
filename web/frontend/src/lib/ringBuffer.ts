// Anneaux typés de l'instrument. La métrique est une mesure, jamais un zéro de
// fortune : une sonde sans réponse est NaN. Le zéro mesuré reste 0. Les tableaux
// floats et l'index de tête sont réutilisés ; aucune allocation dans le flux SSE.
export const RING_CAPACITY = 1800

export class TypedMetricRing {
  readonly capacity: number
  private readonly values: Float64Array
  private readonly times: Float64Array
  private head = 0 // prochain emplacement d'écriture
  private count = 0
  dirty = false // posé à l'écriture, effacé par le peintre rAF

  constructor(capacity = RING_CAPACITY) {
    this.capacity = capacity
    this.values = new Float64Array(capacity)
    this.times = new Float64Array(capacity)
  }

  get size(): number {
    return this.count
  }

  push(ts: number, v: number | null): void {
    // Nan pour une absence, jamais 0 ; un 0 réel passe tel quel.
    this.values[this.head] = v == null || !Number.isFinite(v) ? NaN : v
    this.times[this.head] = ts
    this.head = (this.head + 1) % this.capacity
    if (this.count < this.capacity) this.count++
    this.dirty = true
  }

  private ordered(src: Float64Array): Float64Array {
    const out = new Float64Array(this.count)
    const start = (this.head - this.count + this.capacity) % this.capacity
    for (let i = 0; i < this.count; i++) out[i] = src[(start + i) % this.capacity]
    return out
  }

  /** Valeurs chronologiques — NaN = absence (jamais 0). */
  series(): Float64Array {
    return this.ordered(this.values)
  }

  /** Horodatages chronologiques, alignés sur series(). */
  timestamps(): Float64Array {
    return this.ordered(this.times)
  }

  /** Paires [ts, valeur|null] pour ECharts — NaN devient null et casse le tracé. */
  pairs(): [number, number | null][] {
    const values = this.series()
    const times = this.timestamps()
    const pairs: [number, number | null][] = new Array(values.length)
    for (let i = 0; i < values.length; i++) {
      pairs[i] = [times[i], Number.isNaN(values[i]) ? null : values[i]]
    }
    return pairs
  }

  clear(): void {
    this.head = 0
    this.count = 0
    this.dirty = true
  }
}

export interface DomainEvent {
  ts: number
  kind: string
  label: string
  detail?: string
}

export class DomainEventRing {
  private readonly items: DomainEvent[] = []
  dirty = false

  constructor(private readonly capacity = 200) {}

  push(e: DomainEvent): void {
    this.items.push(e)
    if (this.items.length > this.capacity) this.items.shift()
    this.dirty = true
  }

  /** Ordre chronologique (le plus récent en fin) ; copie pour ne pas exposer le stockage. */
  list(): DomainEvent[] {
    return this.items.slice()
  }

  get size(): number {
    return this.items.length
  }

  clear(): void {
    this.items.length = 0
    this.dirty = true
  }
}

export interface InstrumentRings {
  small: TypedMetricRing
  rtt50: TypedMetricRing
  rtt95: TypedMetricRing
  goodput: TypedMetricRing
  ifaceRx: TypedMetricRing
  ifaceTx: TypedMetricRing
}

// Une seule source pour l'instrument : le flux SSE écrit ici, les surfaces lisent
// ici. L'anneau interface reste court (60 s à 10 Hz) comme dans lib/live.ts.
export const instrument: InstrumentRings & { events: DomainEventRing } = {
  small: new TypedMetricRing(),
  rtt50: new TypedMetricRing(),
  rtt95: new TypedMetricRing(),
  goodput: new TypedMetricRing(),
  ifaceRx: new TypedMetricRing(600),
  ifaceTx: new TypedMetricRing(600),
  events: new DomainEventRing(),
}
