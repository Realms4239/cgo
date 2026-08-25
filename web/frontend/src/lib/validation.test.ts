import { describe, it, expect } from 'vitest'
import { validate } from './validation'
describe('validation', () => {
  it('rejects empty site and duration out of range', () => {
    const r=validate({site:'', duration:5}, {site:{required:true}, duration:{min:10, max:600}})
    expect(r.valid).toBe(false)
    expect(r.errors.site).toBeDefined()
  })
  it('accepts valid', () => {
    const r=validate({site:'Dago', duration:60}, {site:{required:true}, duration:{min:10, max:600}})
    expect(r.valid).toBe(true)
  })
})
