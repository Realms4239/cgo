import { describe, expect, it } from 'vitest'
import {
  AUDIT_LINK_TYPES,
  AUDIT_DURATIONS,
  DEFAULT_AUDIT_DIRECTION,
  auditPhasePlan,
  auditPhases,
  auditStep,
  directionLabel,
  nextAuditStage,
  normalizeDirection,
  normalizeDuration,
  phaseLabel,
  phasePlanLabel,
  remainingSeconds,
  validateAuditDeclaration,
  type AuditStage,
} from './auditFlow'

describe('machine des trois temps', () => {
  it('idle → declared → observing → done', () => {
    let stage: AuditStage = 'idle'
    stage = nextAuditStage(stage, 'open')
    expect(stage).toBe('declared')
    stage = nextAuditStage(stage, 'start')
    expect(stage).toBe('observing')
    stage = nextAuditStage(stage, 'finish')
    expect(stage).toBe('done')
    expect(auditStep(stage)).toBe(3)
  })

  it('annuler une observation revient à la déclaration, jamais à un faux rapport', () => {
    expect(nextAuditStage('observing', 'cancel')).toBe('declared')
    expect(nextAuditStage('declared', 'close')).toBe('idle')
    expect(nextAuditStage('done', 'again')).toBe('declared')
  })

  it('ignore les événements hors étape', () => {
    expect(nextAuditStage('idle', 'start')).toBe('idle')
    expect(nextAuditStage('observing', 'again')).toBe('observing')
  })
})

describe('audit fixe', () => {
  it('une seule durée disponible : 60 secondes', () => {
    expect(AUDIT_DURATIONS).toEqual([60])
    const r = validateAuditDeclaration({ site: 'Département X', link_type: 'fiber', duration: 20, direction: 'up' })
    expect(r.valid).toBe(false)
    expect(r.errors.duration).toBe('audit limité à 60 s')
  })

  it('la normalisation refuse toute autre durée', () => {
    expect(normalizeDuration('60')).toBe(60)
    expect(normalizeDuration('9')).toBeNull()
    expect(normalizeDuration('61')).toBeNull()
  })
})

describe('sens', () => {
  it('défaut up, valeurs connues seules', () => {
    expect(DEFAULT_AUDIT_DIRECTION).toBe('up')
    expect(normalizeDirection(undefined)).toBe('up')
    expect(normalizeDirection('inconnu')).toBe('up')
    expect(normalizeDirection('both')).toBe('both')
    expect(directionLabel('down')).toBe('descendant (down)')
  })

  it('les phases suivent le sens, charge-down nommée pour both', () => {
    expect(auditPhases('up')).toEqual(['repos', 'charge', 'drain'])
    expect(auditPhases('down')).toEqual(['repos', 'charge-down', 'drain'])
    expect(auditPhases('both')).toEqual(['repos', 'charge', 'charge-down', 'drain'])
    expect(phaseLabel('charge-down')).toBe('Charge descendante')
  })
})

describe('validation de la déclaration', () => {
  it('accepte quatre décisions complètes et normalise le site', () => {
    const r = validateAuditDeclaration({ site: '  Département X ', link_type: '5g', duration: '60', direction: 'both' })
    expect(r.valid).toBe(true)
    expect(r.value).toEqual({ site: 'Département X', link_type: '5g', duration: 60, direction: 'both' })
  })

  it('refuse site vide, lien inconnu, durée hors bornes', () => {
    const r = validateAuditDeclaration({ site: ' ', link_type: 'radio', duration: '700', direction: 'up' })
    expect(r.valid).toBe(false)
    expect(r.errors.site).toBeDefined()
    expect(r.errors.link_type).toBeDefined()
    expect(r.errors.duration).toBeDefined()
    expect(r.value).toBeNull()
    expect(AUDIT_LINK_TYPES).toContain('vsat')
  })
})

describe('temps restant', () => {
  it('reste la durée totale sans passer sous zéro', () => {
    expect(remainingSeconds({ running: true, total_sec: 60, elapsed_sec: 23 })).toBe(37)
    expect(remainingSeconds({ running: true, total_sec: 60, elapsed_sec: 61 })).toBe(0)
    expect(remainingSeconds({ running: false })).toBeNull()
  })
})

describe('auditPhasePlan — audit unique de 60 s', () => {
  it('60 s = 10 repos · 40 charge · 10 drain', () => {
    expect(auditPhasePlan(60)).toEqual({ idle: 10, loaded: 40, drain: 10 })
    expect(phasePlanLabel(60)).toBe('60 s au total : 10 s repos · 40 s charge · 10 s drain')
  })
})
