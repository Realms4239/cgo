import { describe, expect, it } from 'vitest'
import { alertsBand, bothAbsent, delayBand, instabilityBand, operatorFacts, operatorSummary, qualityWord, remediationView, retailBound } from './dashboard'
import type { AuditReport } from './types'

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
  it('deux absences tiennent en une bande', () => {
    expect(bothAbsent(false, false)).toBe(true)
    expect(bothAbsent(true, false)).toBe(false)
    expect(bothAbsent(false, true)).toBe(false)
  })
})

// Les quatre faits de la console, écrits une fois pour le tableau de bord et
// pour l'audit : mêmes libellés, mêmes mots d'absence.
describe('operatorFacts', () => {
  const report = (patch: Partial<AuditReport> = {}): AuditReport => ({
    audit_id: 't3',
    measured: { load_attempted: true, load_measured: true, small_measured: true, http_small_p95_ms: 209 },
    cap_sla: { measured: true, ok_pct: 96.7, breach: false, target_pct: 90, deadline_ms: 300 },
    bloat_grade: 'A',
    ...patch,
  })

  it('rend quatre faits aux libellés de la console', () => {
    const facts = operatorFacts(report(), 40, 100)
    expect(facts.map(f => f.key)).toEqual(['alerts', 'delay', 'quality', 'cost'])
    expect(facts.map(f => f.label)).toEqual([
      'Alertes à l’heure', 'Réponse sous charge', 'Qualité du lien', 'Capacité gaspillée',
    ])
    expect(facts.map(f => f.value)).toEqual(['96,7', '209', 'A', 'Non chiffré'])
    expect(facts[0].unit).toBe('%')
    expect(facts[1].unit).toBe('ms')
    expect(facts[0].target).toBe('objectif 90 %')
    expect(facts[0].band).toEqual({ word: "à l'heure", level: 'ok' })
    expect(facts[1].target).toBe('objectif 300 ms')
    expect(facts[1].band).toEqual({ word: 'élevé', level: 'warn' })
    expect(facts[2].target).toBe('note A')
    expect(facts[2].band).toEqual({ word: 'excellent', level: 'ok' })
    expect(facts[3].target).toBe('aucun scénario retenu')
    expect(facts[3].band).toBeNull()
  })

  it('un coût absent se dit Non chiffré', () => {
    // Le rapport archivé porte "primary": null (structure Go émise à vide), pas
    // une clé absente : l'enchaînement optionnel doit absorber les deux formes.
    const noEconomics = operatorFacts(report(), 40, 100)[3]
    expect(noEconomics.value).toBe('Non chiffré')
    expect(noEconomics.unit).toBe('')
    expect(noEconomics.band).toBeNull()
    const primaryNull = operatorFacts(
      report({ economics: { primary: null } as unknown as AuditReport['economics'] }), 40, 100)[3]
    expect(primaryNull.value).toBe('Non chiffré')
  })

  it('un coût chiffré nomme le scénario en français', () => {
    const economics = {
      primary: { model: 'A', name: 'enterprise-gold-cir', ar_per_hour: 72000, basis: 'part réémise', numbers: { model: 'A', cir_mbps: 10 } },
    }
    const cost = operatorFacts(report({ economics }), 40, 100)[3]
    // U+202F : l'espace fine insécable du groupement français.
    expect(cost.value).toBe('72 000,00')
    expect(cost.unit).toBe('Ar/h')
    expect(cost.target).toBe('scénario ligne dédiée')
    expect(cost.band?.word).toBe('estimation contractuelle')
  })

  it('un pourcentage illisible ne devient pas une cellule vide', () => {
    // cap_sla.measured vrai avec ok_pct illisible : le mot d'absence s'affiche,
    // jamais une chaîne vide que l'opérateur lirait comme un zéro.
    const facts = operatorFacts(report({ cap_sla: { measured: true, ok_pct: Number.NaN } }), 40, 100)
    expect(facts[0].value).toBe('Pas encore mesuré')
    expect(facts[0].unit).toBe('')
    expect(facts.every(f => f.value.length > 0)).toBe(true)
  })

  it('sans note, le fait qualité ne dit pas « non noté » à côté d’une note', () => {
    // qualityWord rend toujours une chaîne : la porte sur s.grade est ce qui
    // empêche une note absente de peser comme une note mesurée.
    const quality = operatorFacts(report({ bloat_grade: '' }), 40, 100)[2]
    expect(quality.value).toBe('Pas encore mesuré')
    expect(quality.target).toBe('aucune note')
    expect(quality.band).toBeNull()
  })

  it('sans rapport, les quatre faits sont absents', () => {
    const facts = operatorFacts(null, 40, 100)
    expect(facts.map(f => f.value)).toEqual(['Pas encore mesuré', 'Pas encore mesuré', 'Pas encore mesuré', 'Non chiffré'])
    expect(facts.every(f => f.band === null)).toBe(true)
    // Un rapport antérieur aux deux objectifs scellés n'en invente pas.
    expect(facts[0].target).toBe('')
    expect(facts[1].target).toBe('')
  })
})
