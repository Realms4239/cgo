// T2 — parcs CSS refermés : sélecteurs morts retirés, pip offline distinct et honnête.
// Oracle : on lit la feuille de styles réelle et on rend le vrai StatusPip (aucun mock).
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { StatusPip } from './ui/StatusPip'

const css = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')

const block = (selector: string) => {
  const esc = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return css().match(new RegExp(`${esc}[^{]*\\{[^}]*\\}`))?.[0] ?? ''
}

const mount = (node: ReactNode) => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}

describe('T2 — parcs CSS refermés', () => {
  it('les sélecteurs morts tf-status et labo-matrix-word ont disparu', () => {
    expect(css()).not.toMatch(/\.tf-status\b/)
    expect(css()).not.toMatch(/\.labo-matrix-word\b/)
  })

  it('.et-mark reste (repères de la chronologie Observer)', () => {
    expect(block('.et-mark')).toMatch(/background:\s*var\(--text-muted\)/)
    expect(css()).toMatch(/\.et-mark\s*\{[^}]*transform:\s*rotate\(45deg\)/)
  })

  it('.c-pip.offline .c-pip-dot a une règle distincte, jamais verte', () => {
    const rule = block('.c-pip.offline .c-pip-dot')
    expect(rule, 'règle offline explicite').not.toBe('')
    expect(rule).toMatch(/background:\s*var\(--surface-elevated\)/)
    expect(rule).toMatch(/border:\s*1px solid var\(--text-faint\)/)
    expect(rule).not.toMatch(/var\(--success\)|--series-1|--t-ok|#[0-9a-fA-F]{3,8}\b/)
    expect(block('.c-pip.idle .c-pip-dot')).toMatch(/background:\s*var\(--text-faint\)/)
  })

  it('aucun mouvement ni couleur crue ajouté ; contraste forcé préservé', () => {
    expect(css()).not.toMatch(/transition:\s*all\b/)
    expect(css()).not.toMatch(/@keyframes\b/)
    expect(css()).toMatch(/@media \(forced-colors: active\)/)
    expect(readFileSync(join(process.cwd(), 'src/styles/tokens.css'), 'utf8'))
      .toMatch(/@media \(prefers-reduced-motion: reduce\)/)
  })

  it('le pip « interface absente » porte réellement la classe offline', () => {
    const host = mount(<StatusPip state="offline" label="— interface absente" />)
    expect(host.querySelector('.c-pip.offline .c-pip-dot')).not.toBeNull()
    expect(host.querySelector('.c-pip.idle')).toBeNull()
    expect(host.querySelector('.c-pip.err')).toBeNull()
    expect(host.textContent).toContain('interface absente')
  })
})
