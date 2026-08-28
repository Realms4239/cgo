// Pill — labeled toggle control (header/chooser): "groupes · on", "craft · line".
// The control name is always visible — no cryptic one-word chrome.
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
      aria-pressed={on}
      style={{
        fontFamily: 'var(--font-mono, JetBrains Mono)',
        fontSize: 10,
        fontVariantNumeric: 'tabular-nums',
        letterSpacing: '0.06em',
        padding: '4px 8px',
        border: '1px solid var(--hairline, #26262a)',
        background: on ? 'rgba(90,211,227,0.08)' : 'transparent',
        color: on ? 'var(--t-live, #5ad3e3)' : 'var(--text-muted, #8b9099)',
        cursor: 'pointer',
        display: 'inline-flex',
        gap: 6,
        alignItems: 'center',
        lineHeight: 1.4,
      }}
    >
      <span style={{ color: 'var(--text-faint, #767b84)' }}>{label}</span>
      {value ? <b style={{ fontWeight: 700 }}>{value}</b> : null}
    </button>
  )
}
