// Le hero est le premier consommateur du verdict du moteur. Ces tests figent la
// seule règle qui compte ici : le mot et la phrase viennent de Go ; l'interface
// ne classe rien, ne re-formate rien, n'invente rien, et ne montre jamais une
// couleur sans le mot qui la porte.
// La prose relue sort du DOM (textContent), pas de la chaîne HTML : une
// apostrophe droite s'y échappe en &#x27; sans que l'opérateur lise autre chose
// que le texte du moteur. Vérifier la chaîne HTML laisserait passer une
// véritable réécriture de prose.
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { VerdictHero } from './VerdictHero'
import type { AuditReport, AuditVerdictLevel } from '../lib/types'

function read(p: string): string {
  for (const q of [p, `../${p}`, `../../${p}`, `C:/cgo/${p}`]) {
    try { return readFileSync(q, 'utf8') } catch { /* tente suivante */ }
  }
  return readFileSync(p, 'utf8')
}

const base: AuditReport = {
  audit_id: 'a1', site: 'Département X', link_type: '4g',
  measured: { load_attempted: true, load_measured: true, http_small_p95_ms: 209, small_measured: true },
  operator_verdict: { level: 'green', word: 'Lien sain', sentence: "Sous charge, la latence reste sous l'objectif." },
  cap_sla: { measured: true, ok_pct: 96.7, breach: false },
  bloat_grade: 'A',
}

function mount(r: AuditReport | null | undefined): HTMLElement {
  const host = document.createElement('div')
  host.innerHTML = renderToStaticMarkup(<VerdictHero report={r} />)
  return host
}

const markup = (r: AuditReport | null | undefined) => renderToStaticMarkup(<VerdictHero report={r} />)
const text = (r: AuditReport | null | undefined) => mount(r).textContent ?? ''
const wordText = (r: AuditReport | null | undefined) => mount(r).querySelector('.verdict-word')?.textContent ?? ''
const sentenceText = (r: AuditReport | null | undefined) => mount(r).querySelector('.verdict-sentence')?.textContent ?? ''

describe('VerdictHero', () => {
  it('affiche le mot du moteur, sa phrase et les quatre faits', () => {
    const out = text(base)
    expect(out).toContain('Lien sain')
    expect(out).toContain("Sous charge, la latence reste sous l'objectif.")
    expect(markup(base)).toContain('data-level="green"')
    expect(out).toContain('Alertes à l’heure')
    expect(out).toContain('96,7')
    expect(out).toContain('Réponse sous charge')
    expect(out).toContain('209')
    expect(out).toContain('Qualité du lien')
    expect(out).toContain('note A')
    expect(out).toContain('Capacité gaspillée')
    expect(out).toContain('Non chiffré')
    // aucune absence là où le moteur a mesuré
    expect(out).not.toContain('Pas encore mesuré')
    // surface par défaut : les noms techniques restent dans le tiroir
    for (const forbidden of ['p95', 'p50', 'SLA', 'JFI', 'QDI', 'n=', 'bufferbloat']) {
      expect(out, `interdit sur une surface par défaut : ${forbidden}`).not.toContain(forbidden)
    }
  })

  it('ne transforme jamais une absence en chiffre', () => {
    const noLoad: AuditReport = {
      ...base,
      measured: { load_attempted: true, load_measured: false },
      operator_verdict: { level: 'neutral', word: 'Charge non établie', sentence: 'Le flux saturant n’a pas pu démarrer.' },
      cap_sla: { measured: false, ok_pct: 0, breach: false },
      bloat_grade: '',
    }
    const out = text(noLoad)
    expect(out).toContain('Charge non établie')
    expect(out).toContain('Le flux saturant n’a pas pu démarrer.')
    // trois faits non mesurés, un seul mot d'absence : jamais un zéro lu comme mesuré
    expect(out.match(/Pas encore mesuré/g)).toHaveLength(3)
    expect(out).toContain('Non chiffré')
    expect(out).not.toContain('—')
    expect(out).not.toContain('0,0')
  })

  it('n’invente pas de verdict si le moteur n’en a pas produit', () => {
    const noVerdict: AuditReport = { ...base, operator_verdict: undefined }
    const out = text(noVerdict)
    expect(out).toContain('Pas encore mesuré')
    expect(wordText(noVerdict)).toBe('Pas encore mesuré')
    expect(markup(noVerdict)).toContain('data-level="neutral"')
    expect(out).toContain('Ce rapport ne porte pas de verdict')
    // le motif d'absence ne prend aucune couleur de verdict
    for (const color of ['var(--t-ok)', 'var(--t-threshold)', 'var(--t-danger-text)']) {
      expect(markup(noVerdict), `absence rendue avec ${color}`).not.toContain(color)
    }
  })

  it('un rapport scellé avant le verdict rend le motif d’absence', () => {
    // omitempty ne supprime pas une structure : relire un rapport antérieur à
    // la cascade donne {level:"",word:"",sentence:""}, pas un champ absent.
    const archived: AuditReport = { ...base, operator_verdict: { level: '', word: '', sentence: '' } }
    const out = text(archived)
    expect(wordText(archived)).toBe('Pas encore mesuré')
    expect(markup(archived)).toContain('data-level="neutral"')
    expect(out).toContain('Ce rapport ne porte pas de verdict')
    for (const color of ['var(--t-ok)', 'var(--t-threshold)', 'var(--t-danger-text)']) {
      expect(markup(archived), `absence rendue avec ${color}`).not.toContain(color)
    }
  })

  it('ne montre jamais la phrase du moteur quand le niveau est vide', () => {
    const contradictory: AuditReport = {
      ...base,
      operator_verdict: { level: '', word: '', sentence: 'Sous charge, la latence reste sous l’objectif.' },
    }
    expect(text(contradictory)).not.toContain('Sous charge')
    expect(text(contradictory)).toContain('Pas encore mesuré')
    expect(markup(contradictory)).toContain('data-level="neutral"')
  })

  it('un niveau hors vocabulaire retombe sur le neutre, le mot reste entier', () => {
    // Si Go ajoute un cinquième niveau, le hero ne rend pas une couleur
    // inconnue ni un point sans style : il dit le mot, en neutre.
    const unknownLevel: AuditReport = {
      ...base,
      operator_verdict: { level: 'violet' as unknown as AuditVerdictLevel, word: 'Lien nouveau', sentence: 'Phrase du moteur pour Lien nouveau.' },
    }
    expect(markup(unknownLevel)).toContain('data-level="neutral"')
    expect(wordText(unknownLevel)).toBe('Lien nouveau')
    expect(markup(unknownLevel)).toContain('var(--text-faint)')
  })

  it('rend la phrase du moteur mot pour mot', () => {
    // Phrase verte de la cascade, apostrophe droite et virgule décimale : une
    // prose qui vient de Go traverse le navigateur sans réécriture ni coupe.
    const sentence = "Sous charge, la latence reste sous l'objectif et 96,7 % des alertes passent à l'heure."
    const greenCase: AuditReport = { ...base, operator_verdict: { level: 'green', word: 'Lien sain', sentence } }
    expect(sentenceText(greenCase)).toBe(sentence)
    expect(text(greenCase)).not.toContain('…')
    expect(text(greenCase)).not.toContain('96.7 % des alertes')
  })

  it('porte un mot à côté de chaque couleur — jamais la couleur seule', () => {
    // Les six mots que la cascade peut émettre, sur ses huit bras (note vide et
    // vocabulaire de note inconnu rendent le même neutre, par deux chemins).
    const CASES = [
      { level: 'neutral', word: 'Pas encore mesuré', dot: 'var(--text-faint)' },
      { level: 'neutral', word: 'Charge non établie', dot: 'var(--text-faint)' },
      { level: 'neutral', word: 'Note indisponible', dot: 'var(--text-faint)' },
      { level: 'red', word: 'Alertes en retard', dot: 'var(--t-danger-text)' },
      { level: 'red', word: 'Saturation critique', dot: 'var(--t-danger-text)' },
      { level: 'amber', word: 'Lien fragile', dot: 'var(--t-threshold)' },
      { level: 'green', word: 'Lien sain', dot: 'var(--t-ok)' },
    ] as const
    for (const c of CASES) {
      const r: AuditReport = {
        ...base,
        operator_verdict: { level: c.level, word: c.word, sentence: `Phrase du moteur pour ${c.word}.` },
      }
      const out = markup(r)
      expect(out, c.word).toContain(`data-level="${c.level}"`)
      // le mot occupe seul l'étage hero : le point de couleur ne porte que le
      // decoratif, le jugement reste écrit.
      expect(wordText(r), c.word).toBe(c.word)
      expect(out, c.word).toContain(c.dot)
      expect(sentenceText(r), c.word).toBe(`Phrase du moteur pour ${c.word}.`)
    }
  })

  it('sans rapport, le hero parle encore', () => {
    const out = text(null)
    expect(wordText(null)).toBe('Pas encore mesuré')
    expect(out).toContain('Aucun audit sur ce poste.')
    expect(out).toContain('Non chiffré')
    expect(out).not.toContain('—')
    expect(markup(null)).toContain('data-level="neutral"')
  })
})

describe('VerdictHero — échelle', () => {
  const sheet = read('web/frontend/src/styles/components.css')
  const index = read('web/frontend/src/styles/index.css')

  it('le mot du verdict occupe l’étage hero, les faits l’étage metric', () => {
    expect(sheet).toMatch(/\.verdict-word\s*\{[^}]*font-size:\s*var\(--fs-hero\)/)
    expect(sheet).toMatch(/\.verdict-fact dd b\s*\{[^}]*font-size:\s*var\(--fs-metric\)/)
  })

  it('rien du hero n’emprunte l’étage 13 px, réservé au tiroir', () => {
    const heroRules = [...sheet.matchAll(/\.verdict[a-z-]*(?=[^}]*\{)[^}]*\}/g)].map(m => m[0]).join('\n')
    expect(heroRules, 'le hero n’est pas un tiroir').not.toContain('--fs-fine')
    expect(heroRules).not.toBe('')
  })

  it('les règles du hero vivent dans components.css', () => {
    // index.css est en réécriture simultanée (Tâche 2) : une règle du hero mise
    // là-bas serait perdue ou dupliquée.
    expect(index).not.toMatch(/\.verdict-word\s*\{/)
    expect(index).not.toMatch(/\.verdict-fact\s*\{/)
  })
})
