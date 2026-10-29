import { describe, expect, it } from 'vitest'
import { sectionOpen, stageAfterCampaign, stageAfterReplay } from './bench'

describe('étapes du banc', () => {
  it('démarrer ouvre l’observation', () => {
    expect(stageAfterCampaign('config', true)).toBe('observe')
    expect(stageAfterCampaign('results', true)).toBe('observe')
  })

  it('arrêter depuis l’observation ouvre les résultats', () => {
    expect(stageAfterCampaign('observe', false)).toBe('results')
  })

  it('un arrêt depuis config ne saute pas l’observation', () => {
    expect(stageAfterCampaign('config', false)).toBe('config')
  })

  it('un replay ouvre l’observation', () => {
    expect(stageAfterReplay('config', true)).toBe('observe')
    expect(stageAfterReplay('results', false)).toBe('results')
  })

  it('configurer reste toujours ouvert', () => {
    expect(sectionOpen('config', 'config')).toBe(true)
    expect(sectionOpen('results', 'config')).toBe(true)
  })

  it('observer et résultats apparaissent progressivement', () => {
    expect(sectionOpen('config', 'observe')).toBe(false)
    expect(sectionOpen('observe', 'observe')).toBe(true)
    expect(sectionOpen('observe', 'results')).toBe(false)
    expect(sectionOpen('results', 'results')).toBe(true)
  })
})
