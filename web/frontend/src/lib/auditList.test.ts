import { describe, expect, it, vi, beforeEach } from 'vitest'
import { auditList } from './auditList'

describe('auditList', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('partage un seul aller-retour entre panneaux montés ensemble', async () => {
    const fetchMock = vi.fn(() => Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ audits: [{ audit_id: 'a' }] }),
    } as unknown as Response))
    vi.stubGlobal('fetch', fetchMock)

    const [a, b] = await Promise.all([auditList(), auditList()])
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(a).toEqual(b)
    expect(a.audits).toHaveLength(1)
  })

  it('refetch après résolution — un rafraîchissement explicite reste possible', async () => {
    const fetchMock = vi.fn(() => Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ audits: [] }),
    } as unknown as Response))
    vi.stubGlobal('fetch', fetchMock)

    await auditList()
    await auditList()
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
