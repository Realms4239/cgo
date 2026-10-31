import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { KernelRing } from './KernelRing'

const ebpfTwoFlows = {
  available: true,
  source: 'ebpf',
  flows_seen: 2,
  flows_total: 2,
  flows: [
    { local: '10.0.0.1:40000', peer: '10.0.0.2:5201', retrans_payload_bytes: 2048, retrans_events: 3, qdisc_drop_packets: 1, srtt_us: 12000 },
    { local: '10.0.0.1:40002', peer: '10.0.0.2:5201', retrans_payload_bytes: 2048, retrans_events: 1, qdisc_drop_packets: 2, srtt_us: 15000 },
  ],
}

describe('KernelRing', () => {
  it('rend un créneau par flux et le résumé des totaux', () => {
    const html = renderToStaticMarkup(<KernelRing k={ebpfTwoFlows} />)
    expect(html.match(/class="kernel-slot"/g)?.length).toBe(2)
    expect(html).toContain('data-testid="kernel-ring-summary"')
    expect(html).toContain('2/2 flux')
    expect(html).toContain('échelle 0 – ')
    expect(html).toContain('4096 o réémis')
    expect(html).toContain('3 paquets écartés')
  })

  it('un repli TCP_INFO ne dessine aucun flux', () => {
    const html = renderToStaticMarkup(<KernelRing k={{ available: false, source: 'tcpinfo' }} />)
    expect(html).toBe('')
  })

  it('une source eBPF sans flux ne dessine rien', () => {
    const html = renderToStaticMarkup(<KernelRing k={{ available: true, source: 'ebpf', flows: [] }} />)
    expect(html).toBe('')
  })
})
