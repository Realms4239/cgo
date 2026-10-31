// La fiche terrain : sur papier, la coquille disparaît. Une porte source parce
// qu'une capture d'écran ne dit rien du rendu d'impression.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const css = readFileSync(join(process.cwd(), 'src/styles/index.css'), 'utf8')
const printBlock = css.slice(css.indexOf('@media print'), css.indexOf('@media print') + 700)

describe('fiche terrain (impression)', () => {
  it('masque la coquille et garde le contenu lisible', () => {
    for (const sel of ['.mode-rail', 'header', '.status-ribbon', '.foot-ticker', '.btn']) {
      expect(printBlock, sel).toContain(sel)
    }
    expect(printBlock).toContain('background: #fff')
    expect(printBlock).toContain('break-inside: avoid')
  })
})
