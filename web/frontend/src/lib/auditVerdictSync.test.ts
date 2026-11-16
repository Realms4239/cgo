// Sync TUI ↔ web — le même rapport scellé doit rendre la même verdict sur
// les deux surfaces. La fixture est partagée, octet pour octet : si l'un
// des deux côtés change sa copie, le test échoue.
import { describe, expect, it } from 'vitest'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { VerdictHero } from '../components/VerdictHero'
import { readFileSync } from 'node:fs'
import fixture from './audit-verdict-sync.json'

function read(path: string): string {
  for (const p of [path, `../${path}`, `../../${path}`]) {
    try { return readFileSync(p, 'utf8') } catch { /* try next */ }
  }
  return readFileSync(path, 'utf8')
}

describe('sync TUI ↔ web, même rapport, même verdict', () => {
  it('la fixture web est octet pour octet celle du TUI Go', () => {
    const go = read('cmd/cgo/testdata/audit-verdict-sync.json')
    const web = read('web/frontend/src/lib/audit-verdict-sync.json')
    expect(web).toBe(go)
  })

  it('le hero web affiche exactement le mot du moteur', () => {
    const html = renderToStaticMarkup(createElement(VerdictHero as any, { report: fixture as any }))
    expect(html).toContain('Lien stable')
  })

  it('le fixture porte le verdict proprement (green, pas un défaut)', () => {
    expect(fixture.operator_verdict.level).toBe('green')
    expect(fixture.operator_verdict.word).toBe('Lien stable')
  })
})
