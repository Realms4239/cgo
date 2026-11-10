// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
import { contrastRatio, wcagAaNormalText } from './a11y'

// Audit de contraste des boutons d'appel. Le test lit les DÉCLARATIONS
// réelles (index.css) et résout les var() jusqu'à un hex — il ne recopie
// pas une palette à la main, donc il voit une dérive future.

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* try next */ }
  }
  return readFileSync(p, 'utf8')
}

const tokensCss = read('web/frontend/src/styles/tokens.css')
const indexCss = read('web/frontend/src/styles/index.css')

function tokenValue(name: string, depth = 0): string | null {
  if (depth > 5) return null
  const m = tokensCss.match(new RegExp(`^\\s*--${name}:\\s*([^;]+);`, 'm'))
  if (!m) return null
  const v = m[1].trim()
  const alias = v.match(/^var\(\s*--([a-z0-9-]+)\s*\)$/i)
  if (alias) return tokenValue(alias[1], depth + 1)
  return v
}

// declaration d'une règle CSS simple : `.sel { prop: value; ... }`
function rule(selector: string, prop: string): string | null {
  const m = indexCss.match(new RegExp(`\\${selector}\\s*\\{([^}]*)\\}`))
  if (!m) return null
  const decl = m[1].match(new RegExp(`(?:^|;)\\s*${prop}\\s*:\\s*([^;]+)`))
  if (!decl) return null
  const v = decl[1].trim()
  const alias = v.match(/^var\(\s*--([a-z0-9-]+)\s*\)$/i)
  if (alias) return tokenValue(alias[1])
  return v
}

function resolved(value: string | null): string {
  expect(value, 'déclaration CSS résolue').toBeTruthy()
  return String(value)
}

describe('contraste des boutons d\'appel — AA sur texte courant', () => {
  it('bouton principal : cyan plein, texte canvas', () => {
    const bg = resolved(rule('.btn-primary', 'background'))
    const fg = resolved(rule('.btn-primary', 'color'))
    const r = contrastRatio(fg, bg)
    expect(wcagAaNormalText(r), `btn-primary ${fg} sur ${bg} = ${r.toFixed(2)}:1`).toBe(true)
  })

  it('bouton de danger : rouge profond, texte primaire (le rouge vif ne tient pas le texte)', () => {
    const bg = resolved(rule('.btn-danger', 'background'))
    const fg = resolved(rule('.btn-danger', 'color'))
    const r = contrastRatio(fg, bg)
    expect(wcagAaNormalText(r), `btn-danger ${fg} sur ${bg} = ${r.toFixed(2)}:1`).toBe(true)
  })

  it('bouton secondaire : texte body sur le canvas', () => {
    const fg = resolved(rule('.btn', 'color'))
    const bg = resolved(tokenValue('canvas-solid'))
    const r = contrastRatio(fg, bg)
    expect(wcagAaNormalText(r), `btn ${fg} sur ${bg} = ${r.toFixed(2)}:1`).toBe(true)
  })

  it('destination active : texte canvas sur cyan', () => {
    const fg = resolved(rule('.topnav-btn.on', 'color'))
    const bg = resolved(rule('.topnav-btn.on', 'background'))
    const r = contrastRatio(fg, bg)
    expect(wcagAaNormalText(r), `topnav actif ${fg} sur ${bg} = ${r.toFixed(2)}:1`).toBe(true)
  })
})
