import { useUIStore } from '../store/ui'

// ponytail: fixed beam between Campagne and Live when running — CSS dashOffset -40 linear infinite
// hover-sync (Task 3.3): hovered = chart hover state from LiveView — beam brightens; anime guarded (CSS no-preference)
export function Beam({ hovered = false }: { hovered?: boolean }) {
  const live = useUIStore((s: any) => s.live)
  if (!live?.running) return null
  return (
    <svg
      aria-hidden
      className="beam"
      style={{
        position: 'fixed',
        top: 120,
        left: 'calc(var(--rail-w, 56px) + 24px)',
        width: 'calc(100% - var(--rail-w, 56px) - 48px)',
        height: 2,
        pointerEvents: 'none',
        zIndex: 5,
        opacity: hovered ? 1 : 0.55,
        transition: 'opacity 0.3s ease',
      }}
    >
      <defs>
        <linearGradient id="beam-grad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#5ad3e3" />
          <stop offset="100%" stopColor="#1fa348" />
        </linearGradient>
      </defs>
      <line
        x1={0}
        y1={1}
        x2="100%"
        y2={1}
        stroke="url(#beam-grad)"
        strokeWidth={hovered ? 2 : 1}
        strokeDasharray="4"
        strokeDashoffset={-40}
        style={{ animation: 'beamDash 1s linear infinite' } as any}
      />
    </svg>
  )
}
