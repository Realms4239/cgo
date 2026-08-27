// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'
function read(p:string){ for(const c of [p, p.replace(/^web\/frontend\//,''), `../${p}`, `../../${p}`]){ try{ return readFileSync(c,'utf8') }catch{} } return readFileSync(p,'utf8') }
describe('peek', () => {
  it('PeekPopover exists portal', () => {
    const s=read('web/frontend/src/components/PeekPopover.tsx')
    expect(s).toContain('createPortal')
    expect(s).toContain('160')
  })
})
