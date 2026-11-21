// Règle et pied de champ : oracles DOM pour les faits réels et les absences.
// Aucune valeur fabriquée — un repère vide dit son absence, un champ absent
// affiche « — » + un mot français, jamais un zéro ni un vert.
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField, { type IfaceHonesty } from './TimeField'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

const T0 = Date.UTC(2026, 10, 21, 13, 45, 0)

function filledRings(): InstrumentRings {
  const r: InstrumentRings = {
    small: new TypedMetricRing(20), rtt50: new TypedMetricRing(20),
    rtt95: new TypedMetricRing(20), goodput: new TypedMetricRing(20),
    ifaceRx: new TypedMetricRing(20), ifaceTx: new TypedMetricRing(20),
  }
  for (let i = 0; i < 6; i++) {
    const ts = T0 + i * 1000
    r.small.push(ts, 30 + i)
    r.rtt50.push(ts, 20 + i)
    r.rtt95.push(ts, 25 + i)
    r.goodput.push(ts, 90 + i)
    r.ifaceRx.push(ts, 150 + i)
    r.ifaceTx.push(ts, 40 + i)
  }
  return r
}

function emptyRings(): InstrumentRings {
  return {
    small: new TypedMetricRing(4), rtt50: new TypedMetricRing(4),
    rtt95: new TypedMetricRing(4), goodput: new TypedMetricRing(4),
    ifaceRx: new TypedMetricRing(4), ifaceTx: new TypedMetricRing(4),
  }
}

function absentIface(): IfaceHonesty {
  return { name: '', source: '', valid: false, gap: '', windowMs: 0, rxDrops: 0, txDrops: 0, rxLast: null, txLast: null }
}

function realIface(): IfaceHonesty {
  return { name: 'eth0', source: 'twin', valid: true, gap: 'aucun', windowMs: 1000, rxDrops: 2, txDrops: 0, rxLast: 150, txLast: 40 }
}

function html(node: React.ReactElement): string {
  return renderToStaticMarkup(node as never)
    .replace(/&#x([0-9a-f]+);/gi, (_, c) => String.fromCodePoint(parseInt(c, 16)))
    .replace(/&#(\d+);/g, (_, c) => String.fromCodePoint(Number(c)))
}

describe('règle — repères issus des événements réels', () => {
  it('un anneau d’événements vide dit l’absence, sans repère ni zéro fabriqué', () => {
    const out = html(createElement(TimeField, {
      rings: filledRings(), events: new DomainEventRing(), phaseSince: {},
      iface: absentIface(), jfi: null,
    }))
    expect(out).toContain('aucun repère')
    expect(out).toContain('phase non mesurée')
    expect(out).not.toMatch(/0 repère/)
    expect(out).not.toContain('class="tf-tick"')
    expect(out).not.toContain('class="tf-band"')
  })

  it('des événements réels portent heure, kind et libellé, alignés sur des repères', () => {
    const events = new DomainEventRing()
    events.push({ ts: T0 + 1000, kind: 'phase', label: 'Charge' })
    events.push({ ts: T0 + 3000, kind: 'gates', label: 'Portes fermées' })
    const out = html(createElement(TimeField, {
      rings: filledRings(), events, phaseSince: { charge: T0 + 1000 },
      iface: absentIface(), jfi: null,
    }))
    expect(out).toContain('13:45:01')
    expect(out).toContain('13:45:03')
    expect(out).toContain('phase')
    expect(out).toContain('gates')
    expect(out).toContain('Portes fermées')
    expect((out.match(/class="tf-tick"/g) ?? []).length).toBe(2)
    expect(out).toContain('class="tf-band"')
    expect(out).toContain('Charge')
  })

  it('la légende porte un sens par mot et par forme, jamais par couleur seule', () => {
    const out = html(createElement(TimeField, { rings: emptyRings(), events: new DomainEventRing(), iface: absentIface() }))
    expect(out).toContain('fait horodaté')
    expect(out).toContain('phase mesurée')
    expect(out).toContain('absence')
    expect(out).toContain('▲')
    expect(out).toContain('—')
  })
})

describe('pied de champ — honnêteté interface, équité et provenance', () => {
  it('chaque champ absent affiche « — » et un mot français, jamais un zéro', () => {
    const out = html(createElement(TimeField, {
      rings: emptyRings(), events: new DomainEventRing(), phaseSince: {},
      iface: absentIface(), jfi: null,
    }))
    expect(out).toContain('Interface')
    expect(out).toContain('non renseigné')
    expect(out).toContain('non mesuré')
    expect(out).toContain('Équité')
    expect(out).toContain('équité non mesurée')
    expect(out).toContain('Provenance')
    expect(out).toContain('aucune trace')
    expect(out).not.toContain('0 Mbit/s')
    expect(out).not.toContain('0 pertes')
  })

  it('les champs réels de l’interface et un JFI mesuré portent leurs valeurs', () => {
    const out = html(createElement(TimeField, {
      rings: filledRings(), events: new DomainEventRing(), phaseSince: {},
      iface: realIface(), jfi: 0.83,
    }))
    for (const word of ['eth0', '150', '40', '1000 ms', 'pertes RX', 'twin', 'aucun', 'flux SSE', 'trace vivante', '83 %', 'équité forte']) {
      expect(out, `mot attendu : ${word}`).toContain(word)
    }
    expect(out).toContain('▰')
    expect(out).not.toContain('équité non mesurée')
    expect(out).not.toContain('aucune trace')
  })
})
