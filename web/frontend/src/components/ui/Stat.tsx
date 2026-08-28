// Stat — one tabular number discipline (value + unit + label), no boxes-in-boxes.
export function Stat({ label, value, unit, color = 'var(--text-body, #f2f2f4)', size = 20 }: {
  label?: string
  value: string
  unit?: string
  color?: string
  size?: number
}) {
  return (
    <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, minWidth: 0 }}>
      <span className="mono" style={{ fontFamily: 'var(--font-mono, JetBrains Mono)', fontSize: size, fontWeight: 700, color, fontVariantNumeric: 'tabular-nums', letterSpacing: '-0.02em' }}>{value}</span>
      {unit ? <span className="mono" style={{ fontSize: 11, color: 'var(--text-faint, #767b84)' }}>{unit}</span> : null}
      {label ? <span className="mono" style={{ fontSize: 10, color: 'var(--text-faint, #767b84)', marginLeft: 'auto' }}>{label}</span> : null}
    </div>
  )
}
