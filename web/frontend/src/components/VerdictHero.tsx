// Le verdict vient du moteur. L'interface ne classe rien : elle montre le mot,
// la phrase et les quatre faits qui les justifient, dans cet ordre et à ces
// tailles. Une absence de verdict se dit par un mot neutre, jamais par une
// couleur empruntée à un niveau que le moteur n'a pas produit.
import { operatorFacts } from '../lib/dashboard'
import type { AuditReport, AuditVerdictLevel } from '../lib/types'

// Le point de couleur accompagne le mot, il ne le remplace pas.
const DOT: Record<AuditVerdictLevel, string> = {
  neutral: 'var(--text-faint)',
  amber: 'var(--t-threshold)',
  red: 'var(--t-danger-text)',
  green: 'var(--t-ok)',
}

// Les trois chemins d'absence : pas de rapport du tout, pas de verdict (rapport
// scellé avant la cascade, ou relu avec ses trois champs vides), niveau hors du
// vocabulaire du moteur.
const ABSENT_WORD = 'Pas encore mesuré'
const NO_REPORT_SENTENCE = 'Aucun audit sur ce poste.'
const NO_VERDICT_SENTENCE = 'Ce rapport ne porte pas de verdict : une absence n’est pas un feu vert.'

// Vocabulaire de niveau figé côté Go ; rien d'autre n'a droit à une couleur.
function levelOf(level: string | undefined): AuditVerdictLevel {
  return level === 'neutral' || level === 'red' || level === 'amber' || level === 'green' ? level : 'neutral'
}

// Une valeur non mesurée porte un mot, pas un chiffre : la cellule passe en
// texte cours (voir .verdict-fact[data-empty] dans components.css).
const IS_ABSENT_VALUE = /Pas encore mesuré|Non chiffré/

export function VerdictHero({ report, warnMs = 40, critMs = 100 }: {
  report: AuditReport | null | undefined
  warnMs?: number
  critMs?: number
}) {
  // Un verdict existe quand le moteur a rendu ET un niveau ET un mot. Sinon
  // — rapport ancien, structure relue à vide, niveau hors vocabulaire — c'est
  // l'absence qui parle, en neutre, sans la phrase du moteur.
  const received = report?.operator_verdict
  const verdict = received && received.word.trim() && received.level ? received : null
  const level = levelOf(verdict?.level)
  const facts = operatorFacts(report, warnMs, critMs)
  return (
    <section className="verdict" data-level={level} aria-labelledby="verdict-word">
      <p className="verdict-word" id="verdict-word">
        <span className="verdict-dot" style={{ background: DOT[level] }} aria-hidden="true" />
        {verdict ? verdict.word : ABSENT_WORD}
      </p>
      <p className="verdict-sentence">
        {verdict ? verdict.sentence : (report == null ? NO_REPORT_SENTENCE : NO_VERDICT_SENTENCE)}
      </p>
      <dl className="verdict-facts">
        {facts.map(f => (
          <div className="verdict-fact" key={f.key} data-empty={IS_ABSENT_VALUE.test(f.value) || undefined}>
            <dt>{f.label}</dt>
            <dd>
              <b>{f.value}</b>
              {f.unit ? <small> {f.unit}</small> : null}
            </dd>
            <em>{f.target}{f.band ? ` · ${f.band.word}` : ''}</em>
          </div>
        ))}
      </dl>
    </section>
  )
}
