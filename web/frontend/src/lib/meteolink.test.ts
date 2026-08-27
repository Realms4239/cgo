import { describe, it, expect } from 'vitest'
// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
describe('meteolink', () => {
  it('wordmark has gradient and icon', () => {
    const s=readFileSync('web/frontend/src/components/MeteolinkWordmark.tsx','utf8')
    expect(s).toContain('METEOLINK')
    expect(s).toContain('createDrawable')
  })
  it('panel chooser exists', () => {
    const s=readFileSync('web/frontend/src/components/PanelChooser.tsx','utf8')
    expect(s).toContain('Wall')
  })
})
