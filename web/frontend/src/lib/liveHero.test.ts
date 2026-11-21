// Porte « liveHero » retargetée T6 : l'objectif est une ligne réelle et la
// phase vient du flux SSE, pas d'une estimation d'interface.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

const read = (p: string) => readFileSync(join(process.cwd(), 'src', p), 'utf8')

describe('liveHero', () => {
  it('ligne d’objectif nommée et phase issue du flux', () => {
    const t = read('components/cockpit/TimeField.tsx')
    expect(t).toContain('objectiveMarkLine(goal)')
    expect(t).toContain('objectif {objectiveMs} ms')
    expect(read('lib/chartGrammar.ts')).toContain('chargeMarkArea')
    const s = read('lib/sse.ts')
    expect(s).toContain("kind: 'phase'")
    expect(s).toContain('PHASE_FR')
  })
})
