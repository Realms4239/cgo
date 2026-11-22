// T3 — vocabulaire d'absence unique sur les surfaces LIVE du laboratoire et de
// la synthèse. Oracles DOM indépendants : on rend le vrai composant et on lit
// le DOM produit (aucune mock du composant testé). Pas de chiffre ni de vert
// dans une absence ; le mot porte l'état ; chiffres vivants en tabular-nums.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { EmptyState } from './ui/EmptyState'
import { StatusPip } from './ui/StatusPip'
import Matrix from './labo/Matrix'
import Gates from './labo/Gates'
import CampaignTimeline from './labo/CampaignTimeline'
import { EventTimeline } from './labo/Observer'
import Comparer from './labo/Comparer'
import Preuves from './labo/Preuves'
import TimeField from './cockpit/TimeField'
import JournalRail from './cockpit/JournalRail'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../lib/ringBuffer'

function mount(node: ReactNode): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}

function rings(withData: boolean): InstrumentRings {
  const mk = () => new TypedMetricRing(withData ? 4 : 1)
  const r: InstrumentRings = {
    small: mk(), rtt50: mk(), rtt95: mk(),
    goodput: mk(), ifaceRx: mk(), ifaceTx: mk(),
  }
  if (withData) {
    r.small.push(1000, 42)
    r.goodput.push(1000, 88)
  }
  return r
}

/** Une absence honnête : le primitive, « — », un mot, aucun chiffre, aucun vert. */
function absence(host: HTMLElement, kind: string): HTMLElement {
  const el = host.querySelector<HTMLElement>(`.c-empty.${kind}`)
  expect(el, `EmptyState kind=${kind}`).not.toBeNull()
  const text = el!.textContent ?? ''
  expect(text, 'le tiret d’absence').toContain('—')
  expect(text, 'aucun chiffre fabriqué').not.toMatch(/\d/)
  expect(el!.className, 'aucune classe verte').not.toMatch(/(?:^|\s)(?:ok|success|green|pass)(?:\s|$)/i)
  return el!
}

describe('T3 — primitives d’absence sur les chapitres Labo', () => {
  it('les primitives seules portent le tiret, le mot et la forme', () => {
    const e = mount(<EmptyState kind="loading" hint="chargement réel" />)
    expect(absence(e, 'loading').textContent).toContain('chargement')
    const pip = mount(<StatusPip state="idle" label="— non évaluées" />)
    expect(pip.querySelector('.c-pip.idle')).not.toBeNull()
    expect(pip.textContent).toContain('non évaluées')
  })

  it('Matrice sans plan : EmptyState + pip « indisponible », pas de compte', () => {
    const host = mount(<Matrix plan={{ profiles: [], qdiscs: [], ccs: [], reps: 0 }} progress={null} />)
    const e = absence(host, 'empty')
    expect(e.textContent).toContain('Plan non disponible')
    const pip = host.querySelector('.labo-matrix .c-pip.idle')
    expect(pip, 'mot d’indisponibilité en pip').not.toBeNull()
    expect(pip!.textContent).toContain('indisponible')
  })

  it('Portes non évaluées : le mot est un pip, jamais une pastille verte au rang', () => {
    const host = mount(<Gates gates={[]} />)
    const pip = host.querySelector('.labo-gates .c-pip.idle')
    expect(pip, 'pip d’absence des portes').not.toBeNull()
    expect(pip!.textContent).toContain('non évaluées')
    expect(host.querySelector('.c-pip.ok')).toBeNull()
  })

  it('Chronologie sans phase : EmptyState en attente, sans borne ni vert', () => {
    const host = mount(<CampaignTimeline phaseSince={{}} running={false} />)
    const e = absence(host, 'idle')
    expect(e.textContent).toContain('aucune phase enregistrée')
  })

  it('Observer sans repère : EmptyState en attente, crochets historiques conservés', () => {
    const host = mount(<EventTimeline events={new DomainEventRing()} />)
    const e = absence(host, 'idle')
    expect(e.classList.contains('et-empty')).toBe(true)
    expect(e.querySelector('.et-mark-empty')).not.toBeNull()
    expect(e.textContent).toContain('aucun repère')
  })

  it('Comparer sans mesure : EmptyState « aucune donnée » dans le classement vide', () => {
    const host = mount(<Comparer rows={[]} />)
    const e = host.querySelector('.rank-empty .c-empty.empty')
    expect(e, 'EmptyState du classement vide').not.toBeNull()
    expect(e!.textContent).toContain('Aucune mesure')
    expect(e!.textContent).not.toMatch(/\d/)
  })

  it('Preuves sans service : EmptyState en chargement avec le mot « Preuve manquante »', () => {
    const host = mount(<Preuves archiveRows={[]} integrity={null} />)
    const e = absence(host, 'loading')
    expect(e.classList.contains('archive-wait')).toBe(true)
    expect(e.textContent).toContain('Preuve manquante')
  })
})

describe('T3 — primitives d’absence sur la Synthèse', () => {
  it('TimeField : pip vivant quand la trace existe, pip inerte sans signal', () => {
    const live = mount(<TimeField rings={rings(true)} />)
    const on = live.querySelector('.tf-header .c-pip.on')
    expect(on, 'pip « en direct »').not.toBeNull()
    expect(on!.textContent).toContain('En direct')
    const idle = mount(<TimeField rings={rings(false)} />)
    const off = idle.querySelector('.tf-header .c-pip.idle')
    expect(off, 'pip « aucun signal »').not.toBeNull()
    expect(off!.textContent).toContain('Aucun signal')
  })

  it('TimeField : le pied garde « — » + mot, sans pip vert fabriqué', () => {
    const host = mount(<TimeField rings={rings(false)} />)
    const footer = host.querySelector('.tf-footer')
    expect(footer).not.toBeNull()
    for (const v of Array.from(footer!.querySelectorAll('.tf-ifact-v'))) {
      expect(v.textContent ?? '').not.toMatch(/\d/)
      expect(v.className).not.toMatch(/tf-ok|tf-green|tf-success/)
    }
  })

  it('JournalRail vide : EmptyState « aucune donnée », aucun compte zéro', () => {
    const host = mount(<JournalRail events={new DomainEventRing()} />)
    const e = absence(host, 'empty')
    expect(e.textContent).toContain('Aucun événement')
    expect(e.textContent).not.toContain('0 événement')
  })
})

describe('T3 — tokens de ressenti (feuille de styles réelle)', () => {
  const css = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')
  const block = (selector: string) => {
    const esc = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    return css().match(new RegExp(`${esc}[^{]*\\{[^}]*\\}`))?.[0] ?? ''
  }

  it('aucune transition globale « all »', () => {
    expect(css(), 'transition: all interdit').not.toMatch(/transition:\s*all\b/)
  })

  it('les compteurs vivants alignent leurs chiffres', () => {
    for (const sel of ['.labo-matrix-table', '.labo-tl-legend', '.rank-table', '.tf-fill', '.jr-live']) {
      expect(block(sel), `${sel} en tabular-nums`).toMatch(/font-variant-numeric:\s*tabular-nums/)
    }
  })

  it('les phrases d’absence se composent en text-wrap pretty', () => {
    for (const sel of ['.c-empty-hint', '.labo-absent', '.archive-wait', '.jr-empty']) {
      expect(block(sel), `${sel} en text-wrap pretty`).toMatch(/text-wrap:\s*pretty/)
    }
  })

  it('rayons concentriques : extérieur = intérieur + padding, mouvements et couleurs forcés préservés', () => {
    expect(block('.c-empty'), 'rayon de la surface vide').toMatch(/border-radius:\s*var\(--sp-4\)/)
    expect(block('.c-empty.sm'), 'rayon de la surface vide compacte').toMatch(/border-radius:\s*var\(--sp-3\)/)
    expect(css()).toMatch(/@media \(forced-colors: active\)/)
    expect(readFileSync(join(process.cwd(), 'src/styles/tokens.css'), 'utf8'))
      .toMatch(/@media \(prefers-reduced-motion: reduce\)/)
  })
})
