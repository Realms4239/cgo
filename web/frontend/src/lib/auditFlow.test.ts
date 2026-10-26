import { describe, expect, it } from 'vitest'
import {
  AUDIT_LINK_TYPES,
  DEFAULT_AUDIT_DIRECTION,
  auditPhases,
  auditStep,
  directionLabel,
  durationChoice,
  durationFor,
  liveStats,
  nextAuditStage,
  normalizeDirection,
  normalizeDuration,
  phaseLabel,
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

describe('durée segmentée ↔ libre', () => {
  it('20 et 60 sont les choix segmentés', () => {
    expect(durationFor('20', '')).toBe(20)
    expect(durationFor('60', '')).toBe(60)
    expect(durationChoice(20)).toBe('20')
    expect(durationChoice(60)).toBe('60')
  })

  it('durée libre seulement dans 10–600, sans clamp silencieux', () => {
    expect(durationFor('custom', '90')).toBe(90)
    expect(normalizeDuration('600')).toBe(600)
    expect(normalizeDuration('9')).toBeNull()
    expect(normalizeDuration('601')).toBeNull()
    expect(normalizeDuration('abc')).toBeNull()
    expect(normalizeDuration('')).toBeNull()
    expect(durationChoice(90)).toBe('custom')
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

describe('maths live', () => {
  it('p50/p95 par interpolation linéaire (même formule que metrics)', () => {
    const s = liveStats([10, 0, 20, 30, 40], [])
    expect(s.rttP50).toBeCloseTo(25, 6)
    expect(s.rttP95).toBeCloseTo(38.5, 6)
    expect(s.smallP50).toBeNull()
    expect(s.smallP95).toBeNull()
  })

  it('vide ou zéros seuls ⇒ null, jamais 0 ms mesuré', () => {
    expect(liveStats([], [])).toEqual({ rttP50: null, rttP95: null, smallP50: null, smallP95: null })
    expect(liveStats([0, 0], [0]).rttP50).toBeNull()
  })

  it('reste la durée totale sans passer sous zéro', () => {
    expect(remainingSeconds({ running: true, total_sec: 60, elapsed_sec: 23 })).toBe(37)
    expect(remainingSeconds({ running: true, total_sec: 60, elapsed_sec: 61 })).toBe(0)
    expect(remainingSeconds({ running: false })).toBeNull()
  })
})
