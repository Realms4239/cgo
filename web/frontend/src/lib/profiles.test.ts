import { describe, expect, it } from 'vitest'
import { profileShort } from './profiles'

describe('profileShort', () => {
  it('décrit les quatre profils du référentiel', () => {
    expect(profileShort('P1')).toContain('fibre')
    expect(profileShort('P2')).toContain('4G')
    expect(profileShort('P3')).toContain('VSAT')
    expect(profileShort('P4')).toContain('Orbite basse')
  })
  it('un profil importé garde son identifiant, sans description inventée', () => {
    expect(profileShort('P-site-fibre')).toBe('P-site-fibre')
  })
})
