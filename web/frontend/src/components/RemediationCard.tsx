// La remédiation — ce qui a été mesuré, ce qui sera façonné, et sur quelle
// plateforme. L'interface ne lit jamais un script pour en tirer un chiffre :
// les deux taux viennent des champs du rapport scellé.
import { useState } from 'react'
import { numFr } from '../lib/format'
import type { AuditRemediation } from '../lib/types'
import { useUIStore } from '../store/ui'

const PLATFORM_KEY = 'meteolink-remediation-platform'

type Platform = 'routeros' | 'linux'

const PLATFORM_LABEL: Record<Platform, string> = {
  routeros: 'MikroTik RouterOS 7',
  linux: 'Linux tc',
}

// Le texte copié est celui du rapport : sortie et entrée, jamais recousues.
export function scriptOf(r: AuditRemediation, p: Platform): string {
  const parts = p === 'routeros'
    ? [r.routeros, r.routeros_ingress]
    : [r.linux_tc?.join('\n'), r.linux_tc_ingress?.join('\n')]
  return parts.filter(Boolean).join('\n\n')
}

function platformCarried(r: AuditRemediation, p: Platform): boolean {
  return p === 'routeros' ? !!(r.routeros || r.routeros_ingress) : !!(r.linux_tc?.length || r.linux_tc_ingress?.length)
}

function readChoice(): Platform | null {
  try {
    const v = localStorage.getItem(PLATFORM_KEY)
    return v === 'routeros' || v === 'linux' ? v : null
  } catch { return null }
}

function RateLine({ dir, measured, shaped }: { dir: 'up' | 'down'; measured?: number; shaped?: number }) {
  const label = dir === 'up' ? 'Envoi (montant)' : 'Réception (desc.)'
  const asked = dir === 'up' ? 'montant' : 'descendant'
  if (!measured || measured <= 0) {
    return (
      <div className="remediation-rate" data-empty>
        <dt>{label}</dt>
        <dd>non mesurée — auditer le sens {asked}</dd>
      </div>
    )
  }
  return (
    <div className="remediation-rate">
      <dt>{label}</dt>
      <dd>
        <b>{numFr(shaped && shaped > 0 ? shaped : measured, 1)}</b> <small>Mbit/s</small>
        <span className="remediation-basis">sur {numFr(measured, 1)} mesurés</span>
      </dd>
    </div>
  )
}

export function RemediationCard({ remediation }: { remediation: AuditRemediation | null | undefined }) {
  const carried = (['routeros', 'linux'] as const).filter(p => remediation && platformCarried(remediation, p))
  const [chosen, setChosen] = useState<Platform | null>(() => readChoice())
  if (!remediation || carried.length === 0) return null
  const platform: Platform = chosen && carried.includes(chosen) ? chosen : carried[0]
  const assumed = platform === 'routeros' && !!remediation.routeros_ingress && !!remediation.routeros_lan_assumed
  const script = scriptOf(remediation, platform)

  const pick = (p: Platform) => {
    setChosen(p)
    try { localStorage.setItem(PLATFORM_KEY, p) } catch { /* mode privé : le choix vaut pour la session */ }
  }

  const copy = () => {
    navigator.clipboard?.writeText(script).then(
      () => useUIStore.getState().pushToast(`Configuration ${PLATFORM_LABEL[platform]} copiée`, 'ok'),
      () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'),
    )
  }

  return (
    <section className="remediation" aria-labelledby="remediation-title">
      <h2 id="remediation-title">Protéger ce lien</h2>
      <p className="remediation-sub">
        Régulation active, calibrée sur ce qui a été mesuré (95 % du goulot).
      </p>
      <dl className="remediation-rates">
        <RateLine dir="up" measured={remediation.up_measured_mbps} shaped={remediation.up_shaped_mbps} />
        <RateLine dir="down" measured={remediation.down_measured_mbps} shaped={remediation.down_shaped_mbps} />
      </dl>
      {assumed ? (
        <p className="remediation-assumed">
          interface supposée « {remediation.routeros_lan_iface || 'bridge'} » — à vérifier avant de coller
        </p>
      ) : null}
      <div className="remediation-platform" role="radiogroup" aria-label="Plateforme">
        {carried.map(p => (
          <button
            key={p}
            type="button"
            role="radio"
            aria-checked={platform === p}
            className="remediation-choice"
            onClick={() => pick(p)}
          >
            {PLATFORM_LABEL[p]}
          </button>
        ))}
      </div>
      <div className="remediation-actions">
        <button type="button" className="btn btn-primary" onClick={copy} disabled={!script}>
          Copier la configuration
        </button>
      </div>
    </section>
  )
}
