// Feel-better contracts: hit areas and reduced-motion coverage for press states.
// Oracles read the shipped CSS files, not a mock of them.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const indexCss = () => readFileSync(join(process.cwd(), 'src/styles/index.css'), 'utf8')
const componentsCss = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')

function ruleBlock(css: string, selector: string): string {
  const at = css.indexOf(selector)
  if (at < 0) return ''
  const open = css.indexOf('{', at)
  const close = css.indexOf('}', open)
  return css.slice(open + 1, close)
}

function reducedMotionBlock(css: string): string {
  const at = css.indexOf('@media (prefers-reduced-motion: reduce)')
  if (at < 0) return ''
  const open = css.indexOf('{', at)
  let depth = 1
  let i = open + 1
  for (; i < css.length && depth > 0; i++) {
    if (css[i] === '{') depth++
    if (css[i] === '}') depth--
  }
  return css.slice(open + 1, i - 1)
}

describe('feel-better: tactile cibles et mouvement réduit', () => {
  it('les champs texte ont une cible d’au moins 40px sur desktop', () => {
    const input = ruleBlock(componentsCss(), '.c-input')
    const min = input.match(/min-height:\s*(\d+)px/)
    expect(min, 'min-height de .c-input').toBeTruthy()
    expect(Number(min![1]), 'plancher tactile desktop').toBeGreaterThanOrEqual(40)
  })

  it('le mouvement réduit annule TOUS les états :active, pas seulement .btn', () => {
    const reduced = reducedMotionBlock(indexCss())
    for (const selector of ['.btn', '.inline-disclosure-trigger', '.ui-drawer-close']) {
      expect(reduced, `reduced-motion couvre ${selector}`).toContain(selector)
    }
    expect(reduced).toMatch(/transform:\s*none/)
  })

  it('aucune transition ou will-change globale', () => {
    const all = indexCss() + componentsCss()
    expect(all).not.toMatch(/transition:\s*all\b/)
    expect(all).not.toMatch(/will-change:\s*all\b/)
  })
})
