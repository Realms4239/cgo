// TDD — le sceau n'est dit vérifié qu'après un acte de vérification réel :
// réponse disponible ET empreinte complète de 64 caractères hexadécimaux.
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import SealInspector, { integrityVerified, SealBadge } from './SealInspector'

const HASH = 'a'.repeat(64)

describe('sceau — VÉRIFIÉ seulement après une vraie vérification', () => {
  it('une empreinte incomplète n’est jamais vérifiée', () => {
    expect(integrityVerified({ available: true, sha256: 'abc' })).toBe(false)
  })

  it('une source indisponible, même avec empreinte complète, n’est pas vérifiée', () => {
    expect(integrityVerified({ available: false, sha256: HASH })).toBe(false)
  })

  it('seule une réponse disponible avec empreinte complète est vérifiée', () => {
    expect(integrityVerified({ available: true, sha256: HASH })).toBe(true)
    expect(integrityVerified({ available: true, sha256: 'A'.repeat(64) })).toBe(true)
  })

  it('le badge ne dit VÉRIFIÉ que quand l’acte a réussi', () => {
    expect(renderToStaticMarkup(<SealBadge verified />)).toContain('VÉRIFIÉ')
    const pending = renderToStaticMarkup(<SealBadge verified={false} />)
    expect(pending).not.toContain('VÉRIFIÉ')
    expect(pending).toContain('À VÉRIFIER')
  })

  it('l’inspecteur ne se dit pas vérifié au repos', () => {
    const host = document.createElement('div')
    host.innerHTML = renderToStaticMarkup(<SealInspector />)
    expect(host.textContent).not.toContain('VÉRIFIÉ')
  })
})
