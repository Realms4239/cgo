// TDD — clarté des étapes : chaque chapitre dit ce qu'il montre, jamais un
// résultat, un chiffre ou un verdict ; la déclaration garde ses mots d'absence.
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Chapter from './Chapter'
import Configurer from './Configurer'
import LaboView from '../../views/LaboView'

let uiState: Record<string, unknown>
let auditState: Record<string, unknown>

vi.mock('../../store/ui', () => ({
  useUIStore: (selector: (s: Record<string, unknown>) => unknown) => selector(uiState),
}))
vi.mock('../../store/audit', () => ({
  useAuditStore: (selector: (s: Record<string, unknown>) => unknown) => selector(auditState),
}))

function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}

const OUTCOMES = [
  'Vous verrez la déclaration réelle du lien et les profils disponibles.',
  "Vous verrez la matrice de progression, les portes et la chronologie d'une campagne.",
  "Vous verrez les mesures classées, le duel et l'espace de décision.",
  "Vous verrez l'état des fichiers de preuve, les archives scellées et les preuves vérifiables.",
]

beforeEach(() => {
  uiState = { live: { running: false } }
  auditState = { report: null, stage: 'idle', start: vi.fn(), stop: vi.fn() }
})

describe('Labo step clarity — oracle DOM', () => {
  it('(a) un Chapter avec outcome rend p.chapter-outcome avec le texte exact', () => {
    const p = mount(<Chapter id="x" n="01" title="X" lede="L" active outcome={OUTCOMES[0]} />)
      .querySelector('p.chapter-outcome')
    expect(p).not.toBeNull()
    expect(p!.textContent).toBe(OUTCOMES[0])
  })

  it('(b) sans outcome, aucun noeud .chapter-outcome', () => {
    const host = mount(<Chapter id="x" n="01" title="X" lede="L" active />)
    expect(host.querySelector('.chapter-outcome')).toBeNull()
    expect(host.querySelector('.chapter-lede')).not.toBeNull()
  })

  it('(c) les quatre chapitres exposent une phrase sans chiffre ni verdict', () => {
    const nodes = mount(<LaboView />).querySelectorAll('.chapter-outcome')
    expect(nodes).toHaveLength(4)
    const texts = Array.from(nodes).map(n => n.textContent ?? '')
    expect(texts).toEqual(OUTCOMES)
    for (const text of texts) {
      expect(text).not.toMatch(/\d/)
      expect(text).not.toMatch(/PASS|FAIL|VÉRIFIÉ|verdict|sceau|classement/i)
    }
  })

  it('(d) Configurer porte une ligne de contexte et garde les mots d’absence', () => {
    const host = mount(<Configurer />)
    const ctx = host.querySelector('.configurer-context')
    expect(ctx).not.toBeNull()
    expect(ctx!.textContent).toContain('rapport réel')
    const text = host.textContent ?? ''
    expect(text).toContain('site non déclaré')
    expect(text).toContain('lien non déclaré')
    expect(text).toContain('durée non fixée')
  })
})
