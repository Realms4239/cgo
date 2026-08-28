// CardHead — one mono-caps discipline for every section head (11px caps + 10px sub).
export function CardHead({ label, sub, right }: { label: string; sub?: string; right?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 8, flexWrap: 'wrap', minWidth: 0 }}>
      <div style={{ minWidth: 0 }}>
        <span className="mono" style={{ fontFamily: 'var(--font-mono, JetBrains Mono)', fontSize: 11, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-muted, #8b9099)' }}>{label}</span>
        {sub ? <div className="mono" style={{ fontSize: 10, color: 'var(--text-faint, #767b84)', marginTop: 2 }}>{sub}</div> : null}
      </div>
      {right}
    </div>
  )
}
