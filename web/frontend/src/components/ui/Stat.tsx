// Stat — discipline du chiffre tabulaire (valeur + unité + label).
export function Stat({ label, value, unit, color = 'var(--text-body)', size = 20 }: {
  label?: string
  value: string
  unit?: string
  color?: string
  size?: number
}) {
  return (
    <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, minWidth: 0 }}>
      <span className="mono" style={{ fontFamily: 'var(--font-mono, JetBrains Mono)', fontSize: size, fontWeight: 700, color, fontVariantNumeric: 'tabular-nums', letterSpacing: 'var(--tracking-tight)' }}>{value}</span>
      {unit ? <span className="mono" style={{ fontSize: 'var(--fs-tag)', color: 'var(--text-faint)' }}>{unit}</span> : null}
      {label ? <span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)', marginLeft: 'auto' }}>{label}</span> : null}
    </div>
  )
}
