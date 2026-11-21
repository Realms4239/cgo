// Lisibilité du pied et de la légende de règle : oracles DOM + tokens CSS.
// Aucune valeur fabriquée — l'absence reste « — » + un mot français, jamais
// un zéro ni un vert. Les tokens sont lus sur disque comme dans l'oracle
// spatial, pour que le contrat de typographie soit prouvé, pas supposé.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField, { type IfaceHonesty } from './TimeField'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

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

function footerGroups(out: string): string[] {
  return out.split('<div class="tf-footer-group">').slice(1)
}

function titleOf(group: string): string {
  return group.match(/<span class="tf-footer-title">([^<]*)<\/span>/)?.[1] ?? ''
}

/** Chaque paire label→valeur du groupe ; la jauge est décorative (aria-hidden). */
function values(group: string): { text: string; cls: string; hidden: boolean }[] {
  const out: { text: string; cls: string; hidden: boolean }[] = []
  const wrapper = /<span class="tf-ifact"( aria-hidden="true")?>([\s\S]*?)<\/span><\/span>/g
  for (const m of group.matchAll(wrapper)) {
    const v = m[2].match(/class="tf-ifact-v([^"]*)">([^<]*)/)
    out.push({ cls: v?.[1] ?? '', text: v?.[2] ?? '', hidden: Boolean(m[1]) })
  }
  return out
}

function block(css: string, selector: string): string {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return css.match(new RegExp(`${escaped}\\s*\\{[^}]*\\}`))?.[0] ?? ''
}

const css = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')

describe('lisibilité — groupes étiquetés du pied de champ', () => {
  const out = html(createElement(TimeField, {
    rings: emptyRings(), events: new DomainEventRing(), phaseSince: {},
    iface: absentIface(), jfi: null,
  }))
  const groups = footerGroups(out)

  it('trois groupes exactement, dans l’ordre Interface / Équité / Provenance', () => {
    expect(groups).toHaveLength(3)
    expect(groups.map(titleOf)).toEqual(['Interface', 'Équité', 'Provenance'])
  })

  it('en absence, chaque valeur porte « — » et un mot français, jamais un chiffre', () => {
    for (const group of groups) {
      const title = titleOf(group)
      const pairs = values(group)
      expect(pairs.length, `${title} garde ses champs`).toBeGreaterThan(0)
      for (const v of pairs) {
        expect(v.text, `${title} : aucune valeur ne contient un chiffre en absence`).not.toMatch(/\d/)
        if (v.hidden) continue // la jauge est un ornement aria-hidden, pas une phrase
        expect(v.text, `${title} : « — » présent`).toContain('—')
        expect(v.text, `${title} : mot d’absence français`).toMatch(/non mesuré|non renseigné|aucune trace/)
      }
    }
  })

  it('aucun groupe n’est vide et aucun ne porte une classe verte fabriquée', () => {
    for (const group of groups) {
      expect(group, `${titleOf(group)} rend des valeurs`).toContain('tf-ifact-v')
      expect(group, `${titleOf(group)} n’invente pas de vert`).not.toMatch(/class="[^"]*(?:tf-ok|tf-green|tf-vert|tf-success)[^"]*"/i)
    }
  })

  it('avec de vraies données, RX/TX/nom restent dans le groupe Interface', () => {
    const real = html(createElement(TimeField, {
      rings: emptyRings(), events: new DomainEventRing(), phaseSince: {},
      iface: realIface(), jfi: 0.83,
    }))
    const ifaceGroup = footerGroups(real)[0]
    for (const word of ['eth0', '150 Mbit/s', '40 Mbit/s', '1000 ms', 'twin', 'aucun']) {
      expect(ifaceGroup, `valeur réelle : ${word}`).toContain(word)
    }
    expect(ifaceGroup, 'aucune absence fabriquée sur une interface valide').not.toContain('— non mesuré')
  })
})

describe('lisibilité — tokens de typographie (feuille de styles)', () => {
  it('la grille de pied est à deux colonnes avec un column-gap ratifié', () => {
    const grid = block(css(), '.tf-footer-grid')
    expect(grid, 'grille, pas une ligne flex continue').toMatch(/display:\s*grid/)
    expect(grid, 'deux colonnes label → valeur').toMatch(/grid-template-columns:/)
    expect(grid, 'gouttière de colonne --sp-3').toMatch(/column-gap:\s*var\(--sp-3\)/)
  })

  it('absences en text-wrap pretty, titres et légende en text-wrap balance', () => {
    expect(block(css(), '.tf-absent'), 'phrases d’absence').toMatch(/text-wrap:\s*pretty/)
    expect(block(css(), '.tf-footer-title'), 'titres de groupe').toMatch(/text-wrap:\s*balance/)
    expect(block(css(), '.tf-legend-title'), 'court titre de légende').toMatch(/text-wrap:\s*balance/)
  })

  it('la légende de règle adopte la même typographie label/valeur', () => {
    const legend = block(css(), '.tf-ruler-legend > span')
    expect(legend, 'chiffres alignés').toMatch(/font-variant-numeric:\s*tabular-nums/)
    expect(legend, 'mots de légence non coupés').toMatch(/text-wrap:\s*pretty/)
  })
})
