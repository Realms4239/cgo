export type Capabilities = {
  qdiscs: string[]
  ccs: string[]
  known_ccs: string[]
  kernel: string
  bbr3: boolean
}

const strings = (v: unknown): string[] =>
  Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string' && x.length > 0) : []

export function normalizeCapabilities(raw: unknown): Capabilities | null {
  if (!raw || typeof raw !== 'object') return null
  const j = raw as Record<string, unknown>
  const qdiscs = strings(j.qdiscs)
  const ccs = strings(j.ccs)
  if (!qdiscs.length || !ccs.length) return null
  return {
    qdiscs,
    ccs,
    known_ccs: strings(j.known_ccs),
    kernel: typeof j.kernel === 'string' ? j.kernel : '',
    bbr3: j.bbr3 === true,
  }
}

async function fetchCapabilities(): Promise<Capabilities | null> {
  try {
    const r = await fetch('/api/capabilities')
    if (!r.ok) return null
    return normalizeCapabilities(await r.json())
  } catch {
    return null
  }
}

let cached: Capabilities | null = null
let pending: Promise<Capabilities | null> | null = null

// Une seule lecture par session : le banc et l'aperçu lisent le même
// instantané. Un échec n'est pas mémorisé — un remontage peut réessayer.
export function loadCapabilities(): Promise<Capabilities | null> {
  if (cached) return Promise.resolve(cached)
  if (!pending) {
    pending = fetchCapabilities().then(
      c => {
        if (c) cached = c
        pending = null
        return c
      },
      () => {
        pending = null
        return null
      },
    )
  }
  return pending
}

export function resetCapabilities() {
  cached = null
  pending = null
}
