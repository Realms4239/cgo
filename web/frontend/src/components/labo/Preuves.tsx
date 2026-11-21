// Preuves — les comptes viennent du service d'intégrité, le sceau d'un acte de
// vérification. Tant que la réponse n'est pas là, les créneaux disent leur
// absence ; aucun compte n'est inventé pour remplir la page.
import { useEffect, useState } from 'react'
import { numFr } from '../../lib/format'
import { CardHead } from '../ui/CardHead'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { MetricCard } from '../ui/MetricCard'
import SealInspector from './SealInspector'

interface Integrity {
  available?: boolean
  reason?: string
  runs?: number
  valid?: number
  degraded?: number
  quarantined?: number
  hash8?: string
  sha256?: string
}

export interface PreuvesProps {
  report?: { audit_id?: string } | null
}

export default function Preuves({ report }: PreuvesProps) {
  const [data, setData] = useState<Integrity | null>(null)
  useEffect(() => {
    let alive = true
    fetch('/api/integrity').then(r => r.json()).then(j => { if (alive) setData(j) }).catch(() => {})
    return () => { alive = false }
  }, [])

  return (
    <div className="preuves">
      <p className="preuves-lede">
        Les comptes et l'empreinte viennent du service : rien n'est affirmé avant sa réponse.
      </p>
      <SealInspector />

      <div className="archive">
        <CardHead label="Archives" sub={data?.hash8 ? `sceau ${data.hash8}` : 'sceau en attente'} />
        <div className="archive-counts">
          <MetricCard label="Campagnes archivées" value={data?.runs != null ? numFr(data.runs, 0) : '—'} />
          <MetricCard label="Mesures valides" value={data?.valid != null ? numFr(data.valid, 0) : '—'} />
          <MetricCard label="Mesures écartées" value={data?.quarantined != null ? numFr(data.quarantined, 0) : '—'} />
        </div>
        {!data?.available ? (
          <p className="archive-wait" role="status">Preuve manquante — les comptes et l'empreinte du service d'intégrité s'afficheront ici dès que le service répondra.</p>
        ) : null}
      </div>

      <InlineDisclosure label="D'où viennent ces chiffres">
        <p>
          Les comptes proviennent du service d'intégrité du serveur. L'empreinte complète est vérifiée
          par l'acte de vérification du sceau, jamais par sa seule présence.
        </p>
      </InlineDisclosure>
      <InlineDisclosure label="Rapport d'audit">
        {report?.audit_id
          ? <p className="mono">Rapport {report.audit_id}</p>
          : <p>Preuve manquante — aucun rapport d'audit.</p>}
      </InlineDisclosure>
    </div>
  )
}
