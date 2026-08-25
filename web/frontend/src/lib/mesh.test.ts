// @ts-ignore node:fs types not needed for vitest jsdom — ponytail minimal
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

function read(p: string): string {
  for (const cand of [p, p.replace(/^web\/frontend\//, ''), `../${p}`, `../../${p}`]) {
    try { return readFileSync(cand, 'utf8') } catch { /* try next */ }
  }
  return readFileSync(p, 'utf8')
}
describe('mesh', () => {
  it('WebGLMesh exists and index.css has container-type', () => {
    const c=read('web/frontend/src/components/WebGLMesh.tsx')
    expect(c).toContain('canvas')
    const css=read('web/frontend/src/styles/index.css')
    expect(css).toContain('container-type')
  })
})
