import { formatProvenance } from './Provenance'
import { describe, expect, it } from 'vitest'

describe('formatProvenance', () => {
  it('mot gelé verbatim, sans ` s` inventé', () => {
    expect(formatProvenance({ source: 'mesures gelées', refresh: 'au gel' }))
      .toBe('source : mesures gelées · rafraîchi au gel')
  })
  it('nombres gardent leur ` s`', () => {
    expect(formatProvenance({ source: 'x', refreshSec: 10 }))
      .toBe('source : x · rafraîchi 10 s')
    expect(formatProvenance({ source: 'x', refresh: '10' }))
      .toBe('source : x · rafraîchi 10 s')
  })
})
