import { describe, it, expect } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync, existsSync } from 'node:fs'
// @ts-ignore node:path
import { resolve, join } from 'node:path'
// @ts-ignore process global
declare const process: any

function read(p: string): string {
  const cands = [
    p,
    resolve(p),
    join(process.cwd(), p),
    join(process.cwd(), '..', p),
    join(process.cwd(), '..', '..', p),
    p.replace(/^web\/frontend\//, ''),
    join('web', 'frontend', p.replace(/^web\/frontend\//, '')),
  ]
  for (const c of cands) try { if (existsSync(c)) return readFileSync(c, 'utf8') } catch {}
  return readFileSync(p, 'utf8')
}

describe('tui', () => {
  it('the consolidated TUI entry point exists', () => {
    // main retires cmd/meteolink in favour of the single cgo binary TUI.
    const s = read('cmd/cgo/tui.go')
    expect(s).toContain('func runTUI')
    expect(s).toContain('tea.NewProgram')
  })
})
