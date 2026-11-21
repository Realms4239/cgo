import { describe, expect, it } from 'vitest'
import { LOCAL_FALLBACK, param } from './schema'

// Le schéma local est le repli servi quand /api/schema est absent : il doit
// porter le même défaut produit que le serveur (220 ms, objectif CAP/WIS 2.0).
describe('schema defaults', () => {
  it('ships the CAP/WIS 2.0 deadline as the fallback default', () => {
    expect(param('deadline_ms', LOCAL_FALLBACK)?.default).toBe(220)
  })
})
