// Provenance — trois lignes honnêtes : trace de terrain, confrontation eBPF et
// intégrité. Aucune valeur n'est déduite des trames du banc en mode observe :
// une source absente garde son tiret, son mot français et le nom du fichier
// ou endpoint manquant. Le sceau reste l'acte de SealInspector, pas d'ici.
import { useEffect, useState } from 'react'

export interface IntegrityProbe {
  available?: boolean
  integrity?: string
  ok?: boolean
  hash8?: string
}

export interface DoctorProbe {
  mode?: string
  checks?: unknown[]
}

export interface ProvenanceStripProps {
  // Points d'entrée de test : la vue vit des endpoints réels en production.
  integrity?: IntegrityProbe | null
  doctor?: DoctorProbe | null
}

function useProbe<T>(path: string, injected: boolean): T | null {
  const [data, setData] = useState<T | null>(null)
  useEffect(() => {
    if (injected) return // valeur déjà fournie (parent ou test) : pas de second appel réseau
    let alive = true
    fetch(path)
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(j => { if (alive) setData(j) })
      .catch(() => {})
    return () => { alive = false }
  }, [path, injected])
  return data
}

export default function ProvenanceStrip({ integrity, doctor }: ProvenanceStripProps) {
  const liveIntegrity = useProbe<IntegrityProbe>('/api/integrity', integrity !== undefined)
  const liveDoctor = useProbe<DoctorProbe>('/api/doctor', doctor !== undefined)
  const probe = integrity === undefined ? liveIntegrity : integrity
  const host = doctor === undefined ? liveDoctor : doctor

  // Seule la valeur réelle « integrity » du service qualifie l'état : un hash
  // présent ou un drapeau ok ne prouvent rien sans ce contrôle explicite.
  const status = probe?.integrity
  const verified = status === 'verified'
  const degraded = status === 'degraded'
  const mode = typeof host?.mode === 'string' && host.mode ? host.mode : null
  const traceSource = `capture de terrain autorisée absente${mode ? ` (mode ${mode})` : ''}`

  return (
    <section className="provenance-strip" aria-label="Provenance des preuves" data-testid="provenance-strip">
      <ul className="provenance-lines">
        <li className="provenance-line" data-line="field-trace">
          <span className="provenance-label">Trace de terrain</span>
          <span className="provenance-value provenance-absent">— trace de terrain non mesurée</span>
          <span className="provenance-source">source : {traceSource}</span>
        </li>
        <li className="provenance-line" data-line="ebpf">
          <span className="provenance-label">Confrontation eBPF</span>
          <span className="provenance-value provenance-absent">— confrontation eBPF absente</span>
          <span className="provenance-source">source : ebpf_vs_tc_report.json absent</span>
        </li>
        <li className="provenance-line" data-line="integrity">
          <span className="provenance-label">Intégrité</span>
          {verified || degraded ? (
            <>
              <span className="provenance-value mono" data-state={status}>
                {probe?.hash8 || '—'} · {verified ? 'vérifiée' : 'dégradée'}
              </span>
              <span className="provenance-source">source : /api/integrity</span>
            </>
          ) : (
            <>
              <span className="provenance-value provenance-absent">— intégrité indisponible</span>
              <span className="provenance-source">source : /api/integrity</span>
            </>
          )}
        </li>
      </ul>
    </section>
  )
}
