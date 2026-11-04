// La carte de remédiation : deux directions mesurées, la plateforme que le
// rapport porte, et un texte copié qui ne contient que ce qui a été mesuré.
import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { RemediationCard, scriptOf } from './RemediationCard'
import type { AuditRemediation } from '../lib/types'

const full: AuditRemediation = {
  qdisc: 'fq_codel',
  up_measured_mbps: 5.06,
  up_shaped_mbps: 4.81,
  down_measured_mbps: 18.5,
  down_shaped_mbps: 17.58,
  routeros: '/queue type add name=cgo-shape',
  routeros_ingress: '/interface queue add name=cgo-ingress',
  routeros_wan_iface: 'ether1',
  routeros_lan_iface: 'bridge',
  routeros_lan_assumed: true,
  linux_tc: ['tc qdisc add dev eth0 root handle 1: htb'],
}

const markup = (r: AuditRemediation | null | undefined) => renderToStaticMarkup(<RemediationCard remediation={r} state="available" />)
const text = (r: AuditRemediation | null | undefined) => markup(r).replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ')

describe('RemediationCard', () => {
  it('nomme les deux directions mesurées et leurs taux façonnés', () => {
    const out = text(full)
    expect(out).toContain('Protéger ce lien')
    expect(out).toContain('Envoi (montant)')
    expect(out).toContain('4,8')
    expect(out).toContain('Mbit/s')
    expect(out).toContain('sur 5,1 mesurés')
    expect(out).toContain('Réception (desc.)')
    expect(out).toContain('17,6')
    expect(out).toContain('sur 18,5 mesurés')
    expect(out).toContain('Copier la configuration')
  })

  it('fixture A — une interface supposée est nommée quand le script d’entrée existe', () => {
    const out = text(full)
    expect(out).toContain('interface supposée')
    expect(out).toContain('bridge')
    expect(out).toContain('à vérifier avant de coller')
  })

  it('fixture B — sans script d’entrée, aucune hypothèse n’est étiquetée', () => {
    const out = text({ ...full, routeros_ingress: undefined })
    expect(out).not.toContain('interface supposée')
    expect(out).not.toContain('bridge')
  })

  it('fixture C — une direction absente le dit et n’entre dans aucun script', () => {
    const noDown: AuditRemediation = {
      ...full,
      down_measured_mbps: 0,
      down_shaped_mbps: 0,
      routeros_ingress: undefined,
      routeros_lan_assumed: undefined,
    }
    const out = text(noDown)
    expect(out).toContain('non mesurée — auditer le sens descendant')
    expect(out).not.toContain('17,6')
    expect(out).not.toContain('18,5')
    expect(scriptOf(noDown, 'routeros')).not.toContain('cgo-ingress')
  })

  it('ne liste que les plateformes que le rapport porte', () => {
    const linuxOnly = text({ ...full, routeros: '', routeros_ingress: undefined })
    expect(linuxOnly).toContain('Linux tc')
    expect(linuxOnly).not.toContain('MikroTik RouterOS 7')
    const routerosOnly = text({ ...full, linux_tc: undefined })
    expect(routerosOnly).toContain('MikroTik RouterOS 7')
    expect(routerosOnly).not.toContain('Linux tc')
  })

  it('le texte copié garde le script de la plateforme choisie', () => {
    expect(scriptOf(full, 'routeros')).toContain('/queue type add name=cgo-shape')
    expect(scriptOf(full, 'routeros')).toContain('/interface queue add name=cgo-ingress')
    expect(scriptOf(full, 'linux')).toContain('tc qdisc add dev eth0 root handle 1: htb')
    expect(scriptOf(full, 'linux')).not.toContain('/queue type')
  })

  it('reprend la dernière plateforme choisie', () => {
    localStorage.setItem('meteolink-remediation-platform', 'linux')
    const out = markup(full)
    localStorage.removeItem('meteolink-remediation-platform')
    expect(out).toMatch(/aria-checked="true"[^>]*>Linux tc/)
    expect(out).not.toMatch(/aria-checked="true"[^>]*>MikroTik/)
  })

  it('sans remédiation, la carte ne rend rien', () => {
    expect(markup(null)).toBe('')
    expect(markup({ qdisc: '' })).toBe('')
  })

  it('masque la carte quand l\'état est observation_only', () => {
    const html = renderToStaticMarkup(
      <RemediationCard remediation={full} state="observation_only" />
    )
    expect(html).toContain('Aucune configuration émise')
    expect(html).not.toContain('Copier la configuration')
  })
})
