// MetricRow — une étiquette, des créneaux de mesure alignés, repliés en
// colonne sur écran étroit. Le rythme reste identique mesuré ou absent.
import type { ReactNode } from 'react'

export function MetricRow({ label, children }: { label: string; children?: ReactNode }) {
  return (
    <div className="metric-row">
      <span className="metric-row-label">{label}</span>
      <div className="metric-row-values">{children}</div>
    </div>
  )
}
