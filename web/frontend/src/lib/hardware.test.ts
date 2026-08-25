import { describe, it, expect } from 'vitest'
import { hardwareRecommendation } from './hardware'
describe('hardware', () => {
  it('mikrotik vs isp', () => {
    expect(hardwareRecommendation('fq_codel','P2')).toContain('MikroTik')
    expect(hardwareRecommendation('cake','P1')).toContain('CAKE')
  })
  it('isp gateway', () => {
    expect(hardwareRecommendation('pfifo_fast','P1')).toContain('mini-PC')
  })
})
