import { afterEach, describe, expect, it, vi } from 'vitest'
import { param } from './schema'

// Le client n'a plus aucun repli local : il doit se comporter honnêtement
// quand le contrat est absent (réseau coupé, vieux binaire, 404) — une liste
// vide, pas des bornes inventées.
const params: import('./schema').Param[] = [
  { key: 'deadline_ms', label: 'Deadline', min: 200, max: 5000, default: 220, unit: 'ms', desc: 'objectif' },
]

afterEach(() => {
  vi.unstubAllGlobals()
  vi.resetModules()
})

describe('loadSchema', () => {
  it('fetches the server registry and returns its params', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({ params }) })))
    vi.resetModules()
    const fresh = await import('./schema')
    expect(await fresh.loadSchema()).toEqual(params)
  })

  it('returns an empty list when the server is unreachable — never a fabricated registry', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => { throw new Error('ECONNREFUSED') }))
    vi.resetModules()
    const fresh = await import('./schema')
    expect(await fresh.loadSchema()).toEqual([])
  })

  it('returns an empty list when params is not an array — never a fabricated registry', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({ params: 42 }) })))
    vi.resetModules()
    const fresh = await import('./schema')
    expect(await fresh.loadSchema()).toEqual([])
  })

  it('returns an empty list when params is not an array — never a fabricated registry', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: false, json: async () => ({}) })))
    vi.resetModules()
    const fresh = await import('./schema')
    expect(await fresh.loadSchema()).toEqual([])
  })
})

describe('param', () => {
  it('returns the param when the key exists and undefined otherwise', () => {
    expect(param('deadline_ms', params)?.default).toBe(220)
    expect(param('nope', params)).toBeUndefined()
  })
})
