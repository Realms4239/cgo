// Oracle de lisibilité des axes et de densité du champ temporel. Aucun canvas
// n'est peint ici : on lit l'option pure et la feuille de styles réelle, comme
// les autres oracles du cockpit. Les faits restent dans le DOM, regroupés sous
// un discloseur existant — jamais supprimés, jamais remplacés par un zéro.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import TimeField, { axisTickLabel, timeFieldOption, type IfaceHonesty } from './TimeField'
import DecisionSpine from './DecisionSpine'
import { TypedMetricRing, DomainEventRing, type InstrumentRings } from '../../lib/ringBuffer'

const T0 = Date.UTC(2026, 10, 21, 13, 45, 0)

function emptyRings(): InstrumentRings {
  return {
    small: new TypedMetricRing(4), rtt50: new TypedMetricRing(4),
    rtt95: new TypedMetricRing(4), goodput: new TypedMetricRing(4),
    ifaceRx: new TypedMetricRing(4), ifaceTx: new TypedMetricRing(4),
  }
}

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

function absentIface(): IfaceHonesty {
  return { name: '', source: '', valid: false, gap: '', windowMs: 0, rxDrops: 0, txDrops: 0, rxLast: null, txLast: null }
}

function html(node: React.ReactElement): string {
  return renderToStaticMarkup(node as never)
    .replace(/&#x([0-9a-f]+);/gi, (_, c) => String.fromCodePoint(parseInt(c, 16)))
    .replace(/&#(\d+);/g, (_, c) => String.fromCodePoint(Number(c)))
}

const css = () => readFileSync(join(process.cwd(), 'src/styles/components.css'), 'utf8')

/** Bloc d'une requête média, borné en profondeur pour ne pas avaler la suite. */
function mediaBlock(sheet: string, query: string): string {
  const head = sheet.indexOf(`@media ${query} {`)
  if (head < 0) return ''
  let depth = 0
  for (let i = sheet.indexOf('{', head); i < sheet.length; i++) {
    if (sheet[i] === '{') depth++
    if (sheet[i] === '}') {
      depth--
      if (depth === 0) return sheet.slice(head, i + 1)
    }
  }
  return ''
}

describe('axe — étiquettes courtes et non chevauchantes', () => {
  it('le formateur borne chaque graduation, sans suites de zéros ni notation scientifique', () => {
    expect([0, 0.2, 0.8, 220, 1000].map(axisTickLabel)).toEqual(['0', '0.2', '0.8', '220', '1000'])
    for (const v of [0.00000003, 1e-8, 0.123456789, Number.NaN, Number.POSITIVE_INFINITY]) {
      const label = axisTickLabel(v)
      expect(label.length, `graduation lisible pour ${v}`).toBeLessThanOrEqual(6)
      expect(label, 'pas de suite de zéros collés').not.toMatch(/000000/)
      expect(label, 'pas de notation exponentielle').not.toMatch(/e[+-]/i)
    }
  })

  it('les trois axes y portent leur unité à part et masquent le recouvrement', () => {
    const opt: any = timeFieldOption(emptyRings(), 220)
    const axes: any[] = opt.yAxis
    expect(axes).toHaveLength(3)
    for (const ax of axes) {
      expect(ax.nameLocation, 'unité au bout de l’axe').toBe('end')
      expect(ax.nameGap, 'unité dégagée des graduations').toBeGreaterThanOrEqual(12)
      expect(ax.axisLabel.hideOverlap, 'graduations qui se chevauchent masquées').toBe(true)
      expect(ax.axisLabel.margin, 'graduations dégagées de l’axe').toBeGreaterThanOrEqual(8)
      expect(ax.axisLabel.formatter).toBe(axisTickLabel)
      expect(Number.isFinite(ax.max)).toBe(true)
      expect(ax.max, 'aucune échelle dégénérée').toBeGreaterThanOrEqual(1)
    }
    expect(axes.map((a: any) => a.name)).toEqual(['ms', 'Mbit/s', 'Mbit/s'])
  })

  it('sans étendue temporelle, aucun repère de temps fabriqué à 00:00:00', () => {
    const empty: any = timeFieldOption(emptyRings(), 220)
    expect(empty.xAxis[2].axisLabel.show).toBe(false)
    const live: any = timeFieldOption(filledRings(), 220)
    expect(live.xAxis[2].axisLabel.show).toBe(true)
  })

  it('la ligne d’objectif reste la vraie valeur verrouillée de 220 ms', () => {
    const opt: any = timeFieldOption(emptyRings(), 220)
    expect(opt.series[0].markLine.data).toEqual([{ yAxis: 220 }])
    expect(String(opt.series[0].markLine.label.formatter)).toContain('220')
    const source = readFileSync(join(process.cwd(), 'src/components/cockpit/TimeField.tsx'), 'utf8')
    expect(source).toContain('markLine: objectiveMarkLine(goal)')
    expect(html(createElement(TimeField, { rings: emptyRings(), events: new DomainEventRing(), iface: absentIface() })))
      .toContain('objectif 220 ms')
  })
})

describe('densité — aucun chevauchement à 890 / 1280 / 1920', () => {
  const sheet = css()

  it('890 : la légende passe sur sa propre ligne, hors de la ligne Temps / objectif', () => {
    const legend = sheet.match(/\.tf-ruler-legend\s*\{[^}]*\}/)?.[0] ?? ''
    expect(legend, 'légende repliée').toMatch(/flex-wrap:\s*wrap/)
    expect(legend, 'légende sur sa propre ligne').toMatch(/flex-basis:\s*100%/)
    expect(legend, 'interligne de la légende').toMatch(/row-gap:\s*var\(--sp-1\)/)
    expect(sheet.match(/\.tf-ruler-top\s*\{[^}]*\}/)?.[0] ?? '')
      .toMatch(/row-gap:\s*var\(--sp-1\)/)
    expect(sheet.match(/\.tf-plot\s*\{[^}]*\}/)?.[0] ?? '').toMatch(/height:\s*420px/)
  })

  it('1280 : le champ prend la hauteur restante et le revealer remplace la colonne dense', () => {
    const wide = mediaBlock(sheet, '(min-width: 1024px)')
    expect(wide, 'revealer visible dès 1024').toMatch(/\.band-controls\s*\{[^}]*display:\s*flex/)
    expect(wide, 'PREUVE non dupliqué en tête').toMatch(/\.band-proof\s*\{[^}]*display:\s*none/)
    expect(sheet, 'faits secondaires derrière le discloseur à toute largeur')
      .toMatch(/\.answer-facts\[data-open="false"\]\s*\{\s*display:\s*none;\s*\}/)
    const mid = mediaBlock(sheet, '(min-width: 1024px) and (max-width: 1599px)')
    expect(mid, 'le champ occupe la première rangée').toMatch(/\.cockpit-field\s*\{[^}]*grid-template-rows:\s*minmax\(0,\s*1fr\)\s+auto/)
    expect(wide, 'le tracé n’est plus une bande de 180 px').toMatch(/\.tf-plot\s*\{[^}]*min-height:\s*240px/)
  })

  it('1920 : le champ s’étend au lieu de laisser un vide sous le tracé', () => {
    const huge = mediaBlock(sheet, '(min-width: 1600px)')
    expect(huge).not.toBe('')
    expect(huge, 'colonne de champ étirée').toMatch(/\.cockpit-field\s*\{[^}]*align-items:\s*stretch/)
    expect(huge, 'pas de retour à une hauteur fixe').not.toMatch(/\.tf-plot\s*\{[^}]*height:\s*180px/)
  })
})

describe('faits — tous conservés et accessibles', () => {
  it('les quatre faits secondaires restent dans le DOM derrière « 4 faits »', () => {
    const out = html(createElement(DecisionSpine, { report: null }))
    expect(out).toContain('aria-controls="answer-facts"')
    expect(out).toContain('aria-expanded="false"')
    expect(out).toContain('4 faits')
    for (const label of ['Alertes à l’heure', 'Réponse sous charge', 'Qualité du lien', 'Capacité gaspillée']) {
      expect(out, `fait conservé : ${label}`).toContain(label)
    }
    expect(out).toContain('—')
    expect(out).not.toContain('0 Mbit/s')
  })
})
