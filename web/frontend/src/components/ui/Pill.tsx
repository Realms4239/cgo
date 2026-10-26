// Pill — labeled toggle control (header/chooser): "groupes · on", "craft · line".
// Le nom du contrôle reste toujours visible — pas de chrome cryptique.
export function Pill({ label, value, on = true, onClick, title }: {
  label: string
  value?: string
  on?: boolean
  onClick?: () => void
  title?: string
}) {
  return (
    <button
      onClick={onClick}
      title={title ?? `${label}${value ? `: ${value}` : ''}`}
      aria-label={title ?? `${label}${value ? ` ${value}` : ''}`}
      aria-pressed={on}
      style={{
        fontFamily: 'var(--font-mono, JetBrains Mono)',
        fontSize: 'var(--fs-micro)',
        fontVariantNumeric: 'tabular-nums',
        letterSpacing: 'var(--tracking-caps)',
        padding: '4px 8px',
        border: '1px solid var(--hairline)',
        background: on ? 'rgba(90,211,227,0.08)' : 'transparent',
        color: on ? 'var(--t-live)' : 'var(--text-muted)',
        cursor: 'pointer',
        display: 'inline-flex',
        gap: 6,
        alignItems: 'center',
        lineHeight: 1.4,
      }}
    >
      <span style={{ color: 'var(--text-faint)' }}>{label}</span>
      {value ? <b style={{ fontWeight: 700 }}>{value}</b> : null}
    </button>
  )
}
