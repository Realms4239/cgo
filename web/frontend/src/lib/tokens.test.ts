// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* try next */ }
  }
  // last attempt lets error throw with original path for diagnostics
  return readFileSync(p, 'utf8')
}

describe('tokens', () => {
  it('MadaLink wordmark exists in App.tsx', () => {
    const app = read('web/frontend/src/App.tsx')
    expect(app).toContain('MadaLink')
  })
  it('MadaLink wordmark has glow', () => {
    const app = read('web/frontend/src/App.tsx')
    expect(app).toMatch(/textShadow.*rgba\(90,211,227,0\.4\)/)
  })
  it('tokens.css has radial glow', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('radial-gradient')
  })
  it('tokens.css has noise overlay', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('noise.png')
  })
  it('tokens.css has surface-card gradient and backdrop-filter', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    expect(css).toContain('linear-gradient')
    expect(css).toContain('backdrop-filter')
  })
  it('tokens.css has hairline fading', () => {
    const css = read('web/frontend/src/styles/tokens.css')
    // hairline fading is a linear-gradient with transparent edges
    expect(css).toMatch(/linear-gradient.*transparent.*#26262a/)
  })
  it('index.css has fadeIn and card hover and nav-btn inset', () => {
    const css = read('web/frontend/src/styles/index.css')
    expect(css).toContain('@keyframes fadeIn')
    expect(css).toContain('.card:hover')
    expect(css).toContain('box-shadow: inset')
  })
})
