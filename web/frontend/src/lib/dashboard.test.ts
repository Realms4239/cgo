import { describe, expect, it } from 'vitest'
import { alertsBand, delayBand, instabilityBand, operatorSummary, qualityWord, remediationView, retailBound } from './dashboard'

describe('bandes opérateur', () => {
  it('délai : quatre mots, seuils des Réglages', () => {
    expect(delayBand(30, 40, 100)?.word).toBe('faible')
    expect(delayBand(80, 40, 100)?.word).toBe('modéré')
    expect(delayBand(328, 40, 100)?.word).toBe('élevé')
    expect(delayBand(900, 40, 100)?.word).toBe('critique')
  })

  it('un délai absent ne devient pas « faible »', () => {
    expect(delayBand(undefined, 40, 100)).toBeNull()
    expect(delayBand(0, 40, 100)).toBeNull()
  })

  it('alertes : à l’heure, limite, en retard', () => {
    expect(alertsBand(98)?.level).toBe('ok')
    expect(alertsBand(85)?.level).toBe('warn')
    expect(alertsBand(60)?.level).toBe('crit')
    expect(alertsBand(undefined)).toBeNull()
  })

  it('instabilité : stable, modérée, sévère', () => {
    expect(instabilityBand(20)?.word).toBe('stable')
    expect(instabilityBand(120)?.word).toBe('modérée')
    expect(instabilityBand(355)?.word).toBe('sévère')
  })

  it('qualité du lien en mots', () => {
    expect(qualityWord('A')).toBe('excellent')
    expect(qualityWord('B+')).toBe('bon')
    expect(qualityWord(undefined)).toBe('non noté')
  })
})

describe('projection de remédiation P2', () => {
  const groups = [
    { profile: 'P2', qdisc: 'pfifo_fast', cc: 'bbr', deadline_median: 60, deadline_valid_n: 22, small_p95_median: 328 },
    { profile: 'P2', qdisc: 'cake', cc: 'bbr', deadline_median: 98, deadline_valid_n: 20, small_p95_median: 210 },
    { profile: 'P1', qdisc: 'cake', cc: 'bbr', deadline_median: 100, deadline_valid_n: 16, small_p95_median: 20 },
  ]

  it('ne retient que les cellules P2 valides', () => {
    const v = remediationView(groups)
    expect(v?.simple.qdisc).toBe('pfifo_fast')
    expect(v?.best.qdisc).toBe('cake')
    expect(v?.best.deadlinePct).toBe(98)
    expect(v?.best.smallP95Ms).toBe(210)
    expect(v?.best.n).toBe(20)
  })

  it('sans cellule valide, aucune projection', () => {
    expect(remediationView([{ profile: 'P2', qdisc: 'cake', cc: 'bbr', deadline_median: 99, deadline_valid_n: 0 }])).toBeNull()
  })
})

describe('borne détail', () => {
  it('lit le scénario emergency, pas le primaire', () => {
    const report = {
      economics: {
        primary: { model: 'A', name: 'enterprise-gold-cir', ar_per_hour: 1200, basis: '' },
        scenarios: [
          { model: 'A', name: 'enterprise-gold-cir', ar_per_hour: 1200, basis: '' },
          { model: 'emergency', name: 'retail-outofbundle', ar_per_hour: 72000, basis: '' },
        ],
      },
    }
    expect(retailBound(report)).toBe(72000)
  })

  it('aucun scénario ⇒ null', () => {
    expect(retailBound({ economics: { scenarios: [] } })).toBeNull()
  })
})


describe('operatorSummary', () => {
  it('préfère le petit objet mesuré et n’invente pas un SLA', () => {
    const s = operatorSummary({
      measured: { small_measured: true, http_small_p95_ms: 328, rtt_loaded_p95_ms: 410 },
      cap_sla: { measured: false },
      bloat_grade: 'D',
    }, 40, 100)
    expect(s.delayMs).toBe(328)
    expect(s.delay?.word).toBe('élevé')
    expect(s.alertsPct).toBeUndefined()
    expect(s.alerts).toBeNull()
    expect(s.quality).toBe('faible')
  })

  it('replie sur le RTT chargé quand le petit objet n’est pas mesuré', () => {
    const s = operatorSummary({ measured: { rtt_loaded_p95_ms: 80 }, bloat_grade: 'B' }, 40, 100)
    expect(s.delayMs).toBe(80)
    expect(s.delay?.word).toBe('modéré')
    expect(s.quality).toBe('bon')
  })

  it('sans rapport, tout reste non mesuré', () => {
    const s = operatorSummary(null, 40, 100)
    expect(s.delayMs).toBeUndefined()
    expect(s.delay).toBeNull()
    expect(s.alertsPct).toBeUndefined()
    expect(s.quality).toBe('non noté')
  })
})
