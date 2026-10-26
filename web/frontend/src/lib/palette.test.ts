// palette.ts est la seule maison des couleurs concrètes : ce test verrouille
// le contrat (jetons, repli hors DOM, mapping des gris, cache canvas).
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
import { TOKEN, OFF_TOKEN_GRAYS, FALLBACK, cssVar, paletteOnce, alpha } from './palette'

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* suivante */ }
  }
  return readFileSync(p, 'utf8')
}

describe('palette — contrat', () => {
  it('TOKEN ne porte que des var(--…) — un style, pas une valeur', () => {
    for (const [name, value] of Object.entries(TOKEN)) {
      expect(value, name).toMatch(/^var\(--[a-z0-9-]+\)$/)
    }
  })

  it('chaque gris hors jeton converge vers un nom de jeton connu', () => {
    for (const [gray, name] of Object.entries(OFF_TOKEN_GRAYS)) {
      expect(gray, name).toMatch(/^#[0-9a-f]{6}$/)
      expect(TOKEN[name], `${gray} → ${name}`).toBeDefined()
    }
  })

  it('le repli hors DOM égale tokens.css (aucune dérive silencieuse)', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    for (const [name, value] of Object.entries(FALLBACK)) {
      const m = css.match(new RegExp(`${name}:\\s*([^;]+);`))
      expect(m, `${name} absent de tokens.css`).not.toBeNull()
      expect(m![1].trim(), name).toBe(value)
    }
  })

  it('cssVar retombe sur le repli gelé hors feuille de style', () => {
    expect(cssVar('--t-live')).toBe(FALLBACK['--t-live'])
  })

  it('paletteOnce est mémorisé — le canvas ne relit pas le DOM à chaque frame', () => {
    expect(paletteOnce()).toBe(paletteOnce())
  })

  it('alpha teinte un hex résolu', () => {
    expect(alpha('#5ad3e3', 0.5)).toBe('rgba(90,211,227,0.5)')
  })
})
