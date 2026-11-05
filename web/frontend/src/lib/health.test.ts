import { describe, expect, it } from 'vitest'
import { evaluateLoadedHealth } from './health'

describe('class-aware health', () => {
  it('fiber at +41ms bloat is not healthy simply because it stays below 220ms', () => {
    expect(evaluateLoadedHealth({ linkType: 'fiber', idleP50Ms: 20, loadedP95Ms: 61 }))
      .toEqual({ kind: 'critical', bloatMs: 41 })
  })

  it('cellular uses its own red boundary', () => {
    expect(evaluateLoadedHealth({ linkType: '4g', idleP50Ms: 100, loadedP95Ms: 201 }))
      .toEqual({ kind: 'critical', bloatMs: 101 })
  })

  it('vsat is not failed by its 600ms propagation floor', () => {
    expect(evaluateLoadedHealth({ linkType: 'vsat', idleP50Ms: 600, loadedP95Ms: 720 }))
      .toEqual({ kind: 'attention', bloatMs: 120 })
  })

  it('absent measurements remain absent', () => {
    expect(evaluateLoadedHealth({ linkType: 'fiber' }))
      .toEqual({ kind: 'absent' })
  })

  it('unknown classes do not inherit a universal red boundary', () => {
    expect(evaluateLoadedHealth({ linkType: 'other', idleP50Ms: 40, loadedP95Ms: 240 }))
      .toEqual({ kind: 'unconfigured', bloatMs: 200 })
  })
})
