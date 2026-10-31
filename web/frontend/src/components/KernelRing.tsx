// KernelRing — ne dessine que des lignes noyau réellement mesurées : un repli
// TCP_INFO n'a pas de flux par connexion, donc rien à représenter. Chaque
// barre est un flux eBPF, sa largeur suit les octets réémis.
import type { KernelTelemetry } from '../lib/types'

const MAX_SLOTS = 12

export function KernelRing({ k }: { k: KernelTelemetry }) {
  if (k.source !== 'ebpf') return null
  const flows = k.flows ?? []
  if (flows.length === 0) return null

  const sumBytes = flows.reduce((n, f) => n + f.retrans_payload_bytes, 0)
  const sumDrops = flows.reduce((n, f) => n + f.qdisc_drop_packets, 0)
  const seen = k.flows_seen ?? flows.length
  const total = k.flows_total ?? flows.length
  const slots = flows.slice(0, MAX_SLOTS)

  return (
    <div
      className="kernel-ring"
      data-testid="kernel-ring"
      role="img"
      aria-label={`Flux noyau : ${seen}/${total} flux, ${sumBytes} octets réémis, ${sumDrops} paquets écartés`}
    >
      <div className="mono kernel-ring-summary" data-testid="kernel-ring-summary">
        {`${seen}/${total} flux · ${sumBytes} o réémis · ${sumDrops} paquets écartés`}
      </div>
      <div className="kernel-ring-track">
        {slots.map((f, i) => (
          <span
            key={`${f.local}-${f.peer}-${i}`}
            className="kernel-slot"
            title={`${f.local} → ${f.peer} : ${f.retrans_payload_bytes} o réémis, ${f.qdisc_drop_packets} écartés`}
            style={{ flexGrow: Math.max(1, f.retrans_payload_bytes) }}
          />
        ))}
      </div>
    </div>
  )
}
