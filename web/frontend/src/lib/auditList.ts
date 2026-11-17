// Une seule requête /api/audit/list pour tous les panneaux montés en même
// temps. Le cockpit et le panneau Preuves & Rapports demandaient la même liste
// au premier rendu (2 allers-retours, ~300 ms chacun mesurés le 2026-11-17).
// Le vol inflight est libéré à la fin : un rafraîchissement explicite refetch.
type AuditListResponse = { audits?: unknown[] }

let inflight: Promise<AuditListResponse> | null = null

export function auditList(): Promise<AuditListResponse> {
  if (!inflight) {
    const ctl = new AbortController()
    const timer = setTimeout(() => ctl.abort(), 3000)
    inflight = fetch('/api/audit/list', { signal: ctl.signal })
      .then(r => {
        if (!r.ok) throw new Error(String(r.status))
        return r.json() as Promise<AuditListResponse>
      })
      .finally(() => {
        clearTimeout(timer)
        inflight = null
      })
  }
  return inflight
}
