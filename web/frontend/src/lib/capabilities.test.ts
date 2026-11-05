import { beforeEach, describe, expect, it, vi } from 'vitest'
import { loadCapabilities, normalizeCapabilities, resetCapabilities } from './capabilities'

const PAYLOAD = {
  qdiscs: ['pfifo_fast', 'fq_codel', 'cake'],
  ccs: ['cubic', 'bbr'],
  known_ccs: ['cubic', 'bbr', 'bbr3'],
  kernel: '6.8.0-45-generic',
  bbr3: false,
}

describe('normalizeCapabilities', () => {
  it('accepts the server payload and rejects empty axes', () => {
    expect(normalizeCapabilities(PAYLOAD)).toEqual(PAYLOAD)
    expect(normalizeCapabilities({ ...PAYLOAD, ccs: [] })).toBeNull()
    expect(normalizeCapabilities({ ...PAYLOAD, qdiscs: 'cake' })).toBeNull()
    expect(normalizeCapabilities(null)).toBeNull()
  })

  it('never infers bbr3 from a label', () => {
    expect(normalizeCapabilities({ ...PAYLOAD, bbr3: 'true' })?.bbr3).toBe(false)
  })
})

describe('loadCapabilities', () => {
  beforeEach(() => {
    resetCapabilities()
    vi.unstubAllGlobals()
  })

  it('reads the host once and shares the snapshot', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => PAYLOAD })
    vi.stubGlobal('fetch', fetchMock)
    const [a, b] = await Promise.all([loadCapabilities(), loadCapabilities()])
    expect(a).toEqual(PAYLOAD)
    expect(b).toEqual(PAYLOAD)
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('does not cache a failure', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: false })
      .mockResolvedValueOnce({ ok: true, json: async () => PAYLOAD })
    vi.stubGlobal('fetch', fetchMock)
    expect(await loadCapabilities()).toBeNull()
    expect(await loadCapabilities()).toEqual(PAYLOAD)
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
