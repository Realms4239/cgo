// Le verdict vient du moteur : l'interface ne classe rien, elle montre le mot,
// la phrase et les faits qui les justifient. Une absence garde son mot, jamais
// un zéro ni une couleur empruntée à un niveau non produit. Les quatre états de
// l'Act sont exclusifs : une mesure ne se confond jamais avec un constat.
import type { AuditStage } from '../../lib/auditFlow'
import { operatorFacts } from '../../lib/dashboard'
import { evaluateLoadedHealth } from '../../lib/health'
import type { AuditReport, AuditVerdictLevel } from '../../lib/types'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { MetricRow } from '../ui/MetricRow'
import { StateSlot } from '../ui/StateSlot'

export type ActState =
  | { kind: 'idle' }
  | { kind: 'measuring' }
  | { kind: 'remediable' }
  | { kind: 'quiet'; reason?: string }

export function selectActState(input: {
  stage: AuditStage | undefined
  report: AuditReport | null | undefined
  failure?: { message: string } | null
}): ActState {
  if (input.failure) return { kind: 'quiet', reason: input.failure.message }
  if (input.stage === 'observing') return { kind: 'measuring' }
  if (input.stage === 'done' && input.report) {
    const r = input.report
    const ready = !!r.remediation && r.remediation_state === 'available' && r.remediation_available !== false
    return ready ? { kind: 'remediable' } : { kind: 'quiet' }
  }
  return { kind: 'idle' }
}

// Le point de couleur accompagne le mot, il ne le remplace jamais.
const DOT: Record<AuditVerdictLevel, string> = {
  neutral: 'var(--text-faint)',
  amber: 'var(--t-threshold)',
  red: 'var(--t-danger-text)',
  green: 'var(--t-ok)',
}

const HEALTH_WORD = {
  ok: 'Lien prêt',
  attention: 'À surveiller',
  critical: 'Critique sous charge',
  absent: 'Non mesurée',
  unconfigured: 'Seuil non configuré',
} as const

const ABSENT_VALUE = /Pas encore mesuré|Non chiffré/
const signed = (v: number) => `${v > 0 ? '+' : ''}${v}`

function levelOf(level: string | undefined): AuditVerdictLevel {
  return level === 'neutral' || level === 'red' || level === 'amber' || level === 'green' ? level : 'neutral'
}

export interface ActActions {
  declare?: () => void
  arm?: () => void
  cancel?: () => void
  remediate?: () => void
  proof?: () => void
}

export interface DecisionSpineProps {
  report: AuditReport | null | undefined
  stage?: AuditStage
  failure?: { message: string } | null
  warnMs?: number
  critMs?: number
  actions?: ActActions
}

export default function DecisionSpine({
  report,
  stage,
  failure,
  warnMs = 40,
  critMs = 100,
  actions,
}: DecisionSpineProps) {
  // Une mesure en cours rend le rapport précédent illisible : la même
  // exclusivité que selectActState s'applique à toute la réponse, pas à l'Act
  // seulement. Un rapport périmé ne redevient jamais un verdict courant.
  const readableReport = stage === 'observing' ? null : report
  const received = readableReport?.operator_verdict
  const verdict = !failure && received && received.word.trim() && received.level ? received : null
  const level = levelOf(verdict?.level)
  const word = verdict ? verdict.word : readableReport ? 'VERDICT ABSENT' : 'MESURE ABSENTE'
  const sentence = verdict
    ? verdict.sentence
    : stage === 'observing'
      ? 'Mesure en cours. Le verdict arrivera à la fin.'
      : readableReport == null
        ? 'Aucune mesure pour ce lien.'
        : 'Ce rapport ne contient pas de verdict.'

  const health = evaluateLoadedHealth({
    linkType: readableReport?.link_type ?? '',
    idleP50Ms: readableReport?.measured?.rtt_idle_p50_ms,
    loadedP95Ms: readableReport?.measured?.rtt_loaded_p95_ms,
  })
  const healthText = health.kind === 'absent' ? null
    : health.kind === 'unconfigured'
      ? `Δ ${signed(Math.round(health.bloatMs))} ms — aucun seuil de classe fourni`
      : `Δ ${signed(Math.round(health.bloatMs))} ms sous charge`

  const facts = operatorFacts(readableReport, warnMs, critMs)
  const act = selectActState({ stage, report, failure })
  const actSentence = act.kind === 'idle'
    ? 'Prêt à mesurer ce lien pendant 60 secondes.'
    : act.kind === 'measuring'
      ? 'Mesure en cours — le verdict viendra à la fin.'
      : act.kind === 'remediable'
        ? 'Un réglage peut être examiné à partir de cette mesure.'
        : 'Mesure terminée. Aucun réglage proposé.'

  return (
    <section className="decision-spine" aria-label="Décision">
      <div className="answer-band">
        <h1 id="verdict-word" className="answer-word" data-state={verdict ? 'measured' : 'absent'} data-level={level}>
          <span className="answer-dot" style={{ background: DOT[level] }} aria-hidden="true" />
          {word}
        </h1>
        <p className="answer-sentence">{sentence}</p>
      </div>

      <div className="health-slot">
        <span className="health-label">Santé</span>
        <span className="health-word" data-level={health.kind}>{HEALTH_WORD[health.kind]}</span>
        {healthText ? <span className="health-detail">{healthText}</span> : null}
      </div>

      <div className="answer-facts">
        <MetricRow label="Constat">
          {facts.map(f => {
            const absent = ABSENT_VALUE.test(f.value)
            return (
              <StateSlot
                key={f.key}
                label={f.label}
                value={absent ? null : f.value}
                unit={absent ? undefined : f.unit}
                word={f.key === 'cost' ? 'non chiffré' : 'absent'}
                empty={absent}
              />
            )
          })}
        </MetricRow>
      </div>

      <div className="act-slot" data-state={act.kind} aria-label="Action">
        {act.kind === 'measuring' ? <p role="status">Mesure en cours — la parole sera rendue à la fin.</p> : null}
        {act.kind === 'quiet' && failure ? <p role="status">{failure.message}</p> : null}
        <p className="act-sentence">{actSentence}</p>
        <div className="act-actions">
          {act.kind === 'idle' ? (
            <>
              <button type="button" className="btn" onClick={actions?.declare}>Configurer le lien</button>
              <button type="button" className="btn btn-primary" onClick={actions?.arm}>Démarrer la mesure</button>
            </>
          ) : null}
          {act.kind === 'measuring' ? (
            <button type="button" className="btn btn-danger" onClick={actions?.cancel}>ANNULER</button>
          ) : null}
          {act.kind === 'remediable' ? (
            <button type="button" className="btn btn-primary" onClick={actions?.remediate}>VOIR LA REMÉDIATION</button>
          ) : null}
        </div>
      </div>

      <div className="proof-slot">
        <InlineDisclosure label="Preuves">
          <button type="button" className="btn" onClick={actions?.proof}>Voir les preuves</button>
        </InlineDisclosure>
      </div>
    </section>
  )
}
