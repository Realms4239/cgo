// Preuves — les comptes viennent du service d'intégrité, le sceau d'un acte de
// vérification. Tant que la réponse n'est pas là, les créneaux disent leur
// absence ; aucun compte n'est inventé pour remplir la page.
import { useEffect, useState } from 'react'
import { numFr, asArray } from '../../lib/format'
import { CardHead } from '../ui/CardHead'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { MetricCard } from '../ui/MetricCard'
import { ArchiveTable } from '../ArchiveTable'
import { replayLine } from '../ProofDrawer'
import { copyText } from '../../lib/clipboard'
import { auditCitation } from '../../lib/citation'
import type { ArchiveRow } from '../../lib/archive'
import type { AuditReport } from '../../lib/types'
import { useUIStore } from '../../store/ui'
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
  report?: AuditReport | null
  // Points d'entrée de test : la vue vit des endpoints réels en production.
  archiveRows?: ArchiveRow[]
  integrity?: Integrity | null
}

export default function Preuves({ report, archiveRows, integrity }: PreuvesProps) {
  const [data, setData] = useState<Integrity | null>(integrity === undefined ? null : integrity)
  const [rows, setRows] = useState<ArchiveRow[] | null>(archiveRows === undefined ? null : archiveRows)
  const [citation, setCitation] = useState<string | null>(null)
  const openProof = useUIStore(s => s.openProof)

  useEffect(() => {
    let alive = true
    if (integrity === undefined) {
      fetch('/api/integrity').then(r => r.json()).then(j => { if (alive) setData(j) }).catch(() => {})
    }
    if (archiveRows === undefined) {
      fetch('/api/audit/list').then(r => r.json())
        .then(j => { if (alive) setRows(asArray<ArchiveRow>(j?.audits)) })
        .catch(() => {})
    }
    return () => { alive = false }
  }, [])

  const source = report?.provenance?.source_file?.trim()
  const k = report?.kernel_telemetry
  const sensor = !k
    ? 'Preuve manquante — sonde non instrumentée'
    : k.source === 'ebpf' ? 'compteurs du noyau'
      : k.source === 'tcpinfo' ? `TCP_INFO — ${k.reason || 'raison non consignée'}`
        : `standard — ${k.reason || 'sonde non instrumentée'}`

  // CITER ne copie qu'une citation bâtie sur un rapport scellé réellement lu ;
  // sans empreinte source, le mot d'absence remplace la copie.
  const cite = (id: string) => {
    setCitation(null)
    fetch(`/api/audit/report?audit_id=${encodeURIComponent(id)}`)
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then((j: AuditReport) => {
        if (!j.provenance?.source_sha256) {
          setCitation('citation impossible — empreinte source absente')
          return
        }
        const text = auditCitation(j)
        if (!text) {
          setCitation('citation impossible — le rapport ne porte pas d’identifiant')
          return
        }
        void copyText('Citation', text).then(ok => {
          setCitation(ok
            ? 'Citation copiée — dossier scellé.'
            : 'Copie impossible — sélectionnez le texte')
        })
      })
      .catch(() => setCitation('citation impossible — preuve indisponible'))
  }

  return (
    <div className="preuves">
      <p className="preuves-lede">
        Les comptes et l'empreinte viennent du service : rien n'est affirmé avant sa réponse.
      </p>

      <p className="corpus-line">
        Source des mesures : <span className="mono">{source || '—'}</span>
        <span> — {source ? 'source déclarée' : 'Preuve manquante'}</span>
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
        <ArchiveTable rows={rows ?? []} onOpen={(id) => openProof('proof', id)} onCopyCitation={cite} />
        {citation ? <p className="mono muted archive-citation" role="status">{citation}</p> : null}
      </div>

      <div className="preuves-disclosures">
        <InlineDisclosure label="Quarantaine et écarts">
          <p>
            Mesures écartées : <span className="mono">{data?.quarantined != null ? numFr(data.quarantined, 0) : '—'}</span>.
            {' '}Un écart est écarté par le service, jamais effacé : {data?.quarantined != null ? 'le compte est consigné' : 'Preuve manquante — aucun compte reçu'}.
          </p>
        </InlineDisclosure>
        <InlineDisclosure label="Figures et empreintes">
          <p>
            Empreinte courte : <span className="mono">{data?.hash8 || '—'}</span>.
            {' '}L'empreinte complète n'est dite vérifiée qu'après l'acte du sceau : {data?.sha256 ? 'empreinte reçue' : 'Preuve manquante — aucune empreinte'}.
          </p>
        </InlineDisclosure>
        <InlineDisclosure label="Matériel et sonde">
          <p>Sonde : {sensor}</p>
        </InlineDisclosure>
        <InlineDisclosure label="Rejeu RouterOS">
          <p className="mono">{replayLine(report?.routeros_validation)}</p>
        </InlineDisclosure>
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
