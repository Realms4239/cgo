// TDD — rythme calme du Labo et provenance honnête des calibrations vérifiées.
// Aucune calibration n'est fabriquée : la fixture vient du sidecar scellé
// P4.json et l'empreinte est recalculée sur P4.dist avant d'être revendiquée.
import { createHash } from 'node:crypto'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { act } from 'react'
import { createElement, type ReactNode } from 'react'
import { createRoot } from 'react-dom/client'
import { renderToStaticMarkup } from 'react-dom/server'
import { Window } from 'happy-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Configurer, { type Calibration } from './Configurer'
import type { Capabilities } from '../../lib/capabilities'
import type { Param } from '../../lib/schema'

;(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const CAL_DIR = join(process.cwd(), '../../docs/verification/2026-11-26-field-trace-out3/calibrations')

/** Le scellé est prouvé avant d'être affiché : hash du .dist réel, pas une copie. */
function sealedP4(): { cal: Calibration; digest: string } {
  const side = JSON.parse(readFileSync(join(CAL_DIR, 'P4.json'), 'utf8'))
  const digest = createHash('sha256').update(readFileSync(join(CAL_DIR, 'P4.dist'))).digest('hex')
  expect(side.basis).toBe('measured')
  expect(digest).toBe(side.dist_sha256)
  expect(side.samples).toBe(39)
  return {
    cal: {
      // Le nom est le stem du sidecar servi par GET /api/calibrations.
      name: 'P4', sha256: digest, samples: side.samples,
      p50_ms: side.source_p50_rtt_ms, verified: true,
    },
    digest,
  }
}

const CAPS: Capabilities = {
  qdiscs: ['pfifo_fast', 'cake'],
  ccs: ['cubic', 'bbr'],
  known_ccs: ['cubic', 'bbr', 'bbr3'],
  kernel: '6.8.0-45-generic',
  bbr3: false,
}

const param = (key: string, values: string[]): Param => ({
  key, label: key, default: values[0] ?? '', desc: '', enum: values,
})
const SCHEMA: Param[] = [
  param('link_channel', ['statique', 'empirique']),
  param('link_direction', ['up', 'down', 'both']),
  param('dist_name', []),
  { key: 'bbr3_state', label: 'BBRv3', default: 'BBRv3 non détecté', desc: '' },
]

const mount = (node: ReactNode): HTMLElement => {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(node as never)
  return host
}
const rendered = (calibrations: Calibration[]) =>
  mount(createElement(Configurer, { caps: CAPS, params: SCHEMA, calibrations }))

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, json: async () => ({}) }))
})
afterEach(() => vi.unstubAllGlobals())

/** Feuilles réelles dans l'ordre d'import : les surcharges de rhythm vivent
 *  dans index.css, donc la cascade happy-dom doit être la cascade du produit. */
function layout(width: number) {
  const win = new Window({ url: 'http://localhost', width, height: 900 })
  const doc = win.document
  for (const name of ['tokens.css', 'components.css', 'index.css']) {
    const style = doc.createElement('style')
    style.textContent = readFileSync(join(process.cwd(), 'src/styles', name), 'utf8')
    doc.head.appendChild(style)
  }
  const make = (className: string, tag = 'div', parent?: ReturnType<typeof doc.createElement>) => {
    const el = doc.createElement(tag)
    el.className = className
    ;(parent ?? doc.body).appendChild(el)
    return el
  }
  const configurer = make('configurer')
  const options = make('configurer-options')
  const option = make('configurer-option', 'button')
  const control = make('configurer-control')
  const disclosure = make('inline-disclosure', 'div', configurer)
  const trigger = make('inline-disclosure-trigger', 'button', disclosure)
  const depth = make('observer-depth')
  const field = make('observer-field')
  const duel = make('comparer-duel')
  const rank = make('comparer-rank')
  const disclosures = make('preuves-disclosures')
  const cs = (el: ReturnType<typeof make>) => win.getComputedStyle(el)
  return {
    options: cs(options), option: cs(option), control: cs(control),
    trigger: cs(trigger), disclosure: cs(disclosure),
    depth: cs(depth), field: cs(field), duel: cs(duel), rank: cs(rank),
    disclosures: cs(disclosures),
  }
}

describe('Configurer — rangées de boutons calmes, libellés entiers', () => {
  it('le bouton empirique ne porte que son libellé ; l’absence est dite à côté', () => {
    const chan = rendered([]).querySelector('.configurer-control[data-control="link_channel"]')!
    const buttons = Array.from(chan.querySelectorAll<HTMLButtonElement>('button'))
    expect(buttons.map(b => b.textContent)).toEqual(['Statique', 'Empirique .dist'])
    expect(buttons[1].disabled).toBe(true)
    expect(buttons[1].getAttribute('aria-disabled')).toBe('true')
    const note = chan.querySelector('.configurer-option-note')
    expect(note, 'mot d’absence hors du bouton').not.toBeNull()
    expect(note!.textContent).toContain('aucune calibration vérifiée')
    expect(note!.textContent).not.toMatch(/\d/)
  })

  it('Distribution reste désactivée et sans chiffre sans sidecar vérifié', () => {
    const dist = rendered([]).querySelector('.configurer-control[data-control="distribution"]')!
    expect(dist.querySelector<HTMLSelectElement>('select')!.disabled).toBe(true)
    expect(dist.textContent).toContain('— aucune calibration vérifiée')
    expect(dist.textContent).not.toMatch(/\d/)
  })
})

describe('Labo — rythme mesuré à 890 / 1280 / 1920', () => {
  it('rangées aérées, cibles 44 px, libellés non tronqués', () => {
    for (const width of [890, 1280, 1920]) {
      const l = layout(width)
      expect(Number.parseFloat(l.options.gap), `${width} : écart de rangée`).toBeGreaterThanOrEqual(8)
      expect(Number.parseFloat(l.control.gap), `${width} : écart de contrôle`).toBeGreaterThanOrEqual(12)
      expect(Number.parseFloat(l.option.minHeight), `${width} : cible tactile`).toBeGreaterThanOrEqual(44)
      expect(Number.parseFloat(l.option.paddingTop), `${width} : souffle vertical`).toBeGreaterThanOrEqual(8)
      expect(l.option.whiteSpace, `${width} : libellé sur plusieurs lignes`).not.toBe('nowrap')
      expect(l.option.textOverflow, `${width} : pas d’ellipse`).not.toBe('ellipsis')
    }
  })

  it('le discloseur Profils de campagne prend sa ligne pleine largeur', () => {
    for (const width of [890, 1280, 1920]) {
      const l = layout(width)
      expect(l.trigger.display, `${width} : ligne pleine`).toBe('flex')
      expect(Number.parseFloat(l.disclosure.marginTop), `${width} : souffle avant le discloseur`)
        .toBeGreaterThanOrEqual(12)
    }
  })

  it('Observer / Comparer / Preuves respirent sans masquer d’information', () => {
    for (const width of [890, 1280, 1920]) {
      const l = layout(width)
      expect(Number.parseFloat(l.depth.gap), `${width} : Observer`).toBeGreaterThanOrEqual(16)
      expect(Number.parseFloat(l.field.gap), `${width} : Observer champ`).toBeGreaterThanOrEqual(16)
      expect(Number.parseFloat(l.duel.gap), `${width} : Comparer duel`).toBeGreaterThanOrEqual(12)
      expect(Number.parseFloat(l.rank.gap), `${width} : Comparer classement`).toBeGreaterThanOrEqual(12)
      expect(Number.parseFloat(l.disclosures.gap), `${width} : Preuves`).toBeGreaterThanOrEqual(12)
    }
  })
})

describe('Distribution — seulement le sidecar scellé vérifié', () => {
  it('n’expose que P4, jamais P1/P2/P3 non calibrés', () => {
    const { cal } = sealedP4()
    const others: Calibration[] = ['P1', 'P2', 'P3'].map(name => ({
      name, sha256: '0'.repeat(64), samples: 0, p50_ms: 0, verified: false,
    }))
    const dist = rendered([...others, cal])
      .querySelector('.configurer-control[data-control="distribution"]')!
    expect(Array.from(dist.querySelectorAll('option')).map(o => o.textContent)).toEqual(['P4'])
    for (const name of ['P1', 'P2', 'P3']) expect(dist.textContent).not.toContain(name)
    expect(dist.textContent).not.toContain('VÉRIFIÉ')
  })

  it('la provenance porte l’empreinte recalculée et le compte réel d’échantillons', () => {
    const { cal, digest } = sealedP4()
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    try {
      act(() => {
        root.render(createElement(Configurer, { caps: CAPS, params: SCHEMA, calibrations: [cal] }))
      })
      const sel = host.querySelector<HTMLSelectElement>('select')!
      act(() => {
        sel.value = 'P4'
        sel.dispatchEvent(new Event('change', { bubbles: true }))
      })
      const detail = host.querySelector('.configurer-calib-detail')
      expect(detail, 'détail affiché après sélection').not.toBeNull()
      const text = detail!.textContent ?? ''
      expect(text).toContain('P4')
      expect(text).toContain('39 échantillons')
      expect(text, 'empreinte courte de l’audit source').toContain(digest.slice(0, 8))
      expect(detail!.getAttribute('title'), 'empreinte complète consultable').toContain(digest)
      expect(text).not.toContain('VÉRIFIÉ')
    } finally {
      act(() => root.unmount())
      host.remove()
    }
  })
})
