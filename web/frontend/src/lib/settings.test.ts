import { describe, expect, it } from 'vitest'
import { DEFAULTS } from './settings'

// Le défaut produit suit l'objectif opérateur CAP/WIS 2.0 (220 ms) —
// pas la valeur de confort 1000 que la thèse argumente comme non discriminant
// en deçà. Une valeur persistée par l'utilisateur n'est pas écrasée.
describe('settings defaults', () => {
  it('ships the CAP/WIS 2.0 deadline as the product default', () => {
    expect(DEFAULTS.deadlineMs).toBe(220)
  })
})
