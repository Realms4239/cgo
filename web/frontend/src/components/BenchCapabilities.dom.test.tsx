import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { BenchCapabilities } from './BenchCapabilities'
import type { Capabilities } from '../lib/capabilities'

const base: Capabilities = {
  qdiscs: ['pfifo_fast', 'fq_codel', 'cake'],
  ccs: ['cubic', 'bbr'],
  known_ccs: ['cubic', 'bbr', 'bbr3'],
  kernel: '6.8.0-45-generic',
  bbr3: false,
}

const html = (caps: Capabilities | null) => renderToStaticMarkup(<BenchCapabilities caps={caps} />)

describe('BenchCapabilities', () => {
  it('renders the host lists and names bbr3 absence', () => {
    const out = html(base)
    expect(out).toContain('pfifo_fast')
    expect(out).toContain('cubic')
    expect(out).toContain('BBRv3 absent de ce noyau')
    expect(out).toContain('install-bbr3.sh')
  })

  it('renders bbr3 without the absence hint when the kernel exposes it', () => {
    const out = html({ ...base, ccs: [...base.ccs, 'bbr3'], bbr3: true })
    expect(out).toContain('bbr3')
    expect(out).not.toContain('absent de ce noyau')
  })

  it('says unavailable instead of inventing a list', () => {
    const out = html(null)
    expect(out).toContain('indisponible')
    expect(out).not.toContain('cubic')
  })
})
