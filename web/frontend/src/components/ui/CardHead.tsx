// CardHead — discipline mono-caps pour chaque tête de section (11 px + 10 px).
export function CardHead({ label, sub, right }: { label: React.ReactNode; sub?: string; right?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 8, flexWrap: 'wrap', minWidth: 0 }}>
      <div style={{ minWidth: 0 }}>
        <span className="mono" style={{ fontFamily: 'var(--font-mono, JetBrains Mono)', fontSize: 'var(--fs-tag)', letterSpacing: 'var(--tracking-wide)', textTransform: 'uppercase', color: 'var(--text-muted)' }}>{label}</span>
        {sub ? <div className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)', marginTop: 2 }}>{sub}</div> : null}
      </div>
      {right}
    </div>
  )
}
