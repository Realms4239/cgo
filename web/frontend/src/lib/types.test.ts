// LiveFrame est le miroir exact des tags JSON de pkg/campagne.Snapshot :
// un champ ajouté d'un côté doit exister de l'autre, sinon le mur affiche
// silencieusement un zéro (ou rien) là où le serveur mesure.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

function read(p: string): string {
  for (const q of [p, `../${p}`, `../../${p}`, `C:/cgo/${p}`]) {
    try { return readFileSync(q, 'utf8') } catch { }
  }
  return readFileSync(p, 'utf8')
}

function liveFrameBlock(): string {
  const src = read('web/frontend/src/lib/types.ts')
  const start = src.indexOf('export interface LiveFrame {')
  const end = src.indexOf('\n}', start)
  return src.slice(start, end)
}

describe('LiveFrame — miroir des tags Go', () => {
  it('porte les coûts entreprise/B2B et le JFI mesuré', () => {
    const block = liveFrameBlock()
    const go = read('pkg/campagne/campagne.go')
    for (const key of ['cost_primary_ar_per_h', 'cost_b2b_ar_per_h', 'jfi_pct']) {
      expect(block, `LiveFrame manque ${key}`).toContain(key)
      expect(go, `Snapshot manque ${key}`).toContain(key)
    }
  })
})
