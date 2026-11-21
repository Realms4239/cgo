import { describe, expect, it } from 'vitest'
import { paletteItems } from './CommandPalette'

// La palette suit MODES : deux destinations, une seule ligne chacune, aucun
// identifiant historique. Les verbes restent des alias de recherche.
describe('palette de commandes', () => {
  it('une seule entrée par mode de la bascule', () => {
    const all = paletteItems('')
    expect(all.map((i) => i.mode).sort()).toEqual(['cockpit', 'labo'])
    expect(new Set(all.map((i) => i.mode)).size).toBe(all.length)
  })

  it('aucun libellé ne crie en capitales', () => {
    for (const i of paletteItems('')) {
      expect(i.label, i.label).not.toBe(i.label.toLocaleUpperCase('fr'))
    }
  })

  it('les verbes retrouvent toujours leur destination', () => {
    expect(paletteItems('verdict').map((i) => i.mode)).toEqual(['cockpit'])
    expect(paletteItems('laboratoire').map((i) => i.mode)).toEqual(['labo'])
    expect(paletteItems('comparer').map((i) => i.mode)).toEqual(['labo'])
  })

  it('une recherche sans correspondance ne rend rien, pas un repli', () => {
    expect(paletteItems('zzz-introuvable')).toEqual([])
  })
})
