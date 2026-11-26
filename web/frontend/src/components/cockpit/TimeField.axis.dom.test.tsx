// Oracle de lisibilité des axes et de densité du champ temporel. Aucun canvas
// n'est peint ici : on lit l'option pure et la feuille de styles réelle, comme
// les autres oracles du cockpit. Les faits restent dans le DOM, regroupés sous
// un discloseur existant — jamais supprimés, jamais remplacés par un zéro.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Window } from 'happy-dom'
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

/** Feuilles réelles dans l'ordre d'import de main.tsx : index.css importe
 *  components.css puis ajoute .btn — la cascade doit être la vraie, sinon une
 *  règle moins spécifique peut perdre sans que le source grep ne le voie. */
function layout(width: number) {
  const win = new Window({ url: 'http://localhost', width, height: 900 })
  const doc = win.document
  for (const name of ['tokens.css', 'components.css', 'index.css']) {
    const style = doc.createElement('style')
    style.textContent = readFileSync(join(process.cwd(), 'src/styles', name), 'utf8')
    doc.head.appendChild(style)
  }
  const make = (className: string, tag = 'div') => {
    const el = doc.createElement(tag)
    el.className = className
    doc.body.appendChild(el)
    return el
  }
  const legend = make('tf-ruler-legend')
  const ruler = make('tf-ruler')
  const footer = make('tf-footer')
  const proof = make('band-proof btn', 'button')
  const controls = make('band-controls')
  const trigger = make('facts-disclosure-trigger', 'button')
  const track = make('tf-ruler-track')
  const cs = (el: ReturnType<typeof make>) => win.getComputedStyle(el)
  return {
    legend: cs(legend), ruler: cs(ruler), footer: cs(footer), proof: cs(proof),
    controls: cs(controls), trigger: cs(trigger), track: cs(track),
    overflow: legend.scrollWidth > legend.clientWidth,
  }
}

/** Règles hors de tout @media, accolades équilibrées : prouve la portée globale
 *  sans confondre un @media sans rapport rencontré plus haut dans la feuille. */
function topLevelRules(sheet: string): string {
  let out = ''
  let i = 0
  while (i < sheet.length) {
    const at = sheet.indexOf('@media', i)
    if (at < 0) return out + sheet.slice(i)
    out += sheet.slice(i, at)
    let depth = 0
    let j = sheet.indexOf('{', at)
    for (; j < sheet.length; j++) {
      if (sheet[j] === '{') depth++
      else if (sheet[j] === '}' && --depth === 0) { j++; break }
    }
    i = j
  }
  return out
}

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
    axes.forEach((ax, i) => {
      if (i === 0) {
        expect(ax.nameLocation, 'unité au bout de l’axe primaire').toBe('end')
        expect(ax.nameGap, 'unité dégagée des graduations').toBeGreaterThanOrEqual(12)
      } else {
        expect(ax.nameLocation, 'unité des axes secondaires à mi-hauteur').toBe('middle')
        expect(ax.nameRotate, 'unité secondaire tournée, hors des graduations').toBe(90)
        expect(ax.nameGap, 'unité secondaire séparée des graduations').toBeGreaterThanOrEqual(36)
        expect(ax.nameGap - ax.axisLabel.margin, 'séparation réelle des deux étiquettes')
          .toBeGreaterThanOrEqual(20)
      }
      expect(ax.axisLabel.hideOverlap, 'graduations qui se chevauchent masquées').toBe(true)
      expect(ax.axisLabel.margin, 'graduations dégagées de l’axe').toBeGreaterThanOrEqual(8)
      expect(ax.axisLabel.formatter).toBe(axisTickLabel)
      expect(Number.isFinite(ax.max)).toBe(true)
      expect(ax.max, 'aucune échelle dégénérée').toBeGreaterThanOrEqual(1)
    })
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

  it('layout réel à 890 / 1280 / 1920 : aucun débordement, phrase d’absence séparée du pied', () => {
    for (const width of [890, 1280, 1920]) {
      const l = layout(width)
      expect(l.legend.display, `${width} : légende en flex`).toBe('flex')
      expect(l.legend.flexWrap, `${width} : légende repliable`).toBe('wrap')
      expect(l.legend.flexBasis, `${width} : légende sur sa propre ligne`).toBe('100%')
      expect(Number.parseFloat(l.track.marginLeft) + Number.parseFloat(l.track.marginRight),
        `${width} : piste dans le viewport`).toBeLessThan(width)
      expect(l.overflow, `${width} : pas de débordement horizontal`).toBe(false)
      // La phrase « aucun repère » ne doit pas coller l'en-tête INTERFACE.
      expect(Number.parseFloat(l.ruler.paddingBottom), `${width} : souffle sous la règle`)
        .toBeGreaterThanOrEqual(8)
      expect(Number.parseFloat(l.footer.paddingTop), `${width} : souffle avant le pied`)
        .toBeGreaterThanOrEqual(8)
    }
  })

  it('cascade réelle : PREUVE masqué en tête à 1280/1920, « Preuves » du bas conservé', () => {
    for (const width of [1280, 1920]) {
      const l = layout(width)
      expect(l.proof.display, `${width} : PREUVE en tête masqué`).toBe('none')
      expect(l.controls.display, `${width} : revealer visible`).toBe('flex')
      expect(l.trigger.display, `${width} : bouton 4 faits visible`).toBe('inline-flex')
    }
    expect(layout(890).proof.display, 'PREUVE reste accessible sous 1024').toBe('inline-flex')
    expect(html(createElement(DecisionSpine, { report: null })), 'discloseur du bas conservé')
      .toContain('Preuves')
  })

  it('1280 : le champ prend la hauteur restante et le revealer remplace la colonne dense', () => {
    const wide = mediaBlock(sheet, '(min-width: 1024px)')
    expect(wide, 'revealer visible dès 1024').toMatch(/\.band-controls\s*\{[^}]*display:\s*flex/)
    expect(wide, 'PREUVE non dupliqué en tête : sélecteur aussi spécifique que .btn')
      .toMatch(/\.band-proof\.btn\s*\{[^}]*display:\s*none/)
    const topLevel = topLevelRules(sheet)
    expect(topLevel, 'faits derrière le discloseur par une règle globale, pas seulement sous 1023px')
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
