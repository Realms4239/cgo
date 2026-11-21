// Portes source de la dernière tranche : mouvement, contraste forcé,
// barres de défilement, et suppression des coquilles mortes. Le test lit
// les feuilles réelles : une régression de style doit rougir sans navigateur.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

function read(p: string): string {
  for (const cand of [p, p.replace('web/frontend/', ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* suite */ }
  }
  return readFileSync(p, 'utf8')
}

const css = read('web/frontend/src/styles/index.css')
const tokens = read('web/frontend/src/styles/tokens.css')
const anime = read('web/frontend/src/lib/anime.ts')

describe('contraste forcé et mouvement réduit', () => {
  it('le focus forced-colors utilise Highlight, pas une couleur de marque', () => {
    const block = css.match(/@media \(forced-colors: active\) \{[\s\S]*?\n\}/)
    expect(block, 'bloc forced-colors présent').not.toBeNull()
    expect(block![0]).toContain('Highlight')
    expect(block![0]).toContain('CanvasText')
  })

  it('reduced-motion collapse toute animation et transition', () => {
    const block = css.match(/@media \(prefers-reduced-motion: reduce\) \{[\s\S]*?\n\}/)
    expect(block, 'bloc reduced-motion présent').not.toBeNull()
    expect(block![0]).toContain('animation-duration')
    expect(block![0]).toContain('transition-duration')
    expect(block![0]).toContain('animation-iteration-count')
  })

  it('les barres de défilement ont des jetons et un état de survol', () => {
    expect(css).toContain('--scrollbar-thumb')
    expect(css).toContain('--scrollbar-track')
    expect(css).toContain('--scrollbar-thumb-hover')
    expect(css).toContain('::-webkit-scrollbar-thumb:hover')
    expect(css).toContain('scrollbar-color')
  })

  it('les jetons clairs sont déclarés sans toucher aux jetons historiques', () => {
    expect(css).toContain('@theme')
    expect(css).toContain('--color-canvas: #fbfbfb')
    expect(tokens).toContain('--canvas-solid: #fbfbfb')
  })
})

describe('coquilles mortes retirées', () => {
  it('ni bandeau de portes, ni rail, ni brosse ne subsistent', () => {
    for (const dead of ['.gates {', '.rail-peek', '.rail-gate', '.brush .handle', '.brush .selection']) {
      expect(css).not.toContain(dead)
    }
  })

  it('exports anime inutilisés retirés', () => {
    for (const dead of [
      'export function animateBar(', 'export function animateBarWidth(',
      'export function animateCardStagger(', 'export function flashRowUp(',
      'export function animateBannerPulse(', 'export function animateLiveEnter(',
      'export function animateBeam(', 'export function animateDonut(',
      'export function animateGrid(', 'export const EASE_OUT',
    ]) {
      expect(anime).not.toContain(dead)
    }
    // Le garde animeSvg reste vrai : la fonction svg/text vit toujours.
    expect(anime).toContain('createDrawable')
    expect(anime).toContain('splitText')
  })
})
