import { describe, expect, it } from 'vitest'
import { paletteItems } from './CommandPalette'

// La palette proposait sept lignes dont trois menaient exactement là où trois
// autres menaient déjà (« Auditer » et « Auditer le lien réel », « Banc d'essai »
// et « Ouvrir le banc d'essai »). Une doublure de navigation est le défaut que la
// console supprime ailleurs : elle ne se le permet pas ici non plus.
describe('palette de commandes', () => {
  it('une seule entrée par destination', () => {
    const all = paletteItems('')
    expect(all.map((i) => i.mode).sort()).toEqual(['cockpit', 'labo'])
    expect(new Set(all.map((i) => i.mode)).size).toBe(all.length)
  })

  it('aucun libellé ne crie en capitales', () => {
    // La règle est « phrase case » : un libellé entièrement en capitales est une
    // phrase française hurlée, ce que la console interdit sur une surface.
    for (const i of paletteItems('')) {
      expect(i.label, i.label).not.toBe(i.label.toLocaleUpperCase('fr'))
    }
  })

  it('les verbes du vieux QUICK retrouvent toujours leur destination', () => {
    // L'alias de recherche remplace la ligne dupliquée : « preuve » mène au
    // Cockpit (verdict + preuves en ligne), « banc » au Labo.
    expect(paletteItems('preuve').map((i) => i.mode)).toEqual(['cockpit'])
    expect(paletteItems('labo').map((i) => i.mode)).toEqual(['labo'])
    expect(paletteItems('auditer le lien réel').map((i) => i.mode)).toEqual(['cockpit'])
  })

  it('une recherche sans correspondance ne rend rien, pas un repli', () => {
    expect(paletteItems('zzz-introuvable')).toEqual([])
  })
})
