// T9 round 1 — gardes source pour la classe « premier plan sombre sur fond
// sombre/saturé » exposée par la bascule claire. Chaque garde lit les vrais
// fichiers et calcule un vrai ratio WCAG depuis les jetons : elle rougit si un
// premier plan sombre revient, sans toucher aux seuils d'axe.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
import { FALLBACK } from './palette'

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* suivante */ }
  }
  return readFileSync(p, 'utf8')
}

const comp = read('web/frontend/src/styles/components.css')
const idx = read('web/frontend/src/styles/index.css')
const sheet = read('web/frontend/src/styles/tokens.css')
const prompt = read('web/frontend/src/components/QuickActionsPrompt.tsx')

// --t-danger-solid n'est pas dans FALLBACK (palette() ne le consomme pas) :
// on le lit dans la feuille réelle, comme le fait le navigateur.
const fromSheet = (name: string): string => {
  const m = sheet.match(new RegExp(`^\\s*${name}:\\s*(#[0-9a-fA-F]{6})\\s*;`, 'm'))
  expect(m, `${name} absent de tokens.css`).not.toBeNull()
  return m![1]
}

const lin = (c: number) => { const x = c / 255; return x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4 }
const lum = (hex: string) => {
  const n = parseInt(hex.slice(1), 16)
  return 0.2126 * lin((n >> 16) & 255) + 0.7152 * lin((n >> 8) & 255) + 0.0722 * lin(n & 255)
}
const ratio = (a: string, b: string) => {
  const [hi, lo] = [lum(a), lum(b)].sort((x, y) => y - x)
  return (hi + 0.05) / (lo + 0.05)
}
const rule = (css: string, selector: string): string => css.match(new RegExp(`${selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')} \\{[^}]*\\}`))?.[0] ?? ''

describe('T9 round 1 — premier plan sur fond sombre ou saturé', () => {
  it('la pastille QuickActionsPrompt n’a plus de fond charcoal', () => {
    expect(prompt, 'aucun fond rgba(16,16,18,.92)').not.toContain('rgba(16,16,18,0.92)')
    expect(prompt, 'surface claire pour l’overlay').toContain("background: 'var(--surface-elevated)'")
    expect(prompt, 'bordure de cheveu claire').toContain("border: '1px solid var(--hairline)'")
  })

  it('les trois textes de la pastille tiennent AA sur la surface claire', () => {
    const bg = FALLBACK['--surface-card']
    for (const token of ['--text-primary', '--t-threshold', '--t-neutral'] as const) {
      const r = ratio(FALLBACK[token], bg)
      expect(r, `${token} sur ${bg}`).toBeGreaterThanOrEqual(4.5)
    }
    // oracle hostile : l'ancien fond charcoal faisait échouer les mêmes textes
    expect(ratio(FALLBACK['--text-primary'], '#232325')).toBeLessThan(4.5)
  })

  it('.c-badge--danger porte un premier plan clair, pas --text-primary', () => {
    const badge = rule(comp, '.c-badge--danger')
    expect(badge, 'règle .c-badge--danger présente').not.toBe('')
    expect(badge).toContain('color: var(--chart-on-ok)')
    expect(badge).not.toContain('color: var(--text-primary)')
  })

  it('.btn-danger porte un premier plan clair, pas --text-primary', () => {
    const btn = rule(idx, '.btn-danger')
    expect(btn, 'règle .btn-danger présente').not.toBe('')
    expect(btn).toContain('color: var(--chart-on-ok)')
    expect(btn).not.toContain('color: var(--text-primary)')
  })

  it('blanc sur rouge danger tient AA ; encre sombre échoue (mesure réelle)', () => {
    for (const red of [FALLBACK['--t-danger'], fromSheet('--t-danger-solid')]) {
      expect(ratio(FALLBACK['--chart-on-ok'], red), `${red} + blanc`).toBeGreaterThanOrEqual(4.5)
      expect(ratio(FALLBACK['--text-primary'], red), `${red} + encre`).toBeLessThan(4.5)
    }
  })
})
