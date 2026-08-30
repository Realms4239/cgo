// EmptyChart — vide honnête: indication centrée sur la grille
// hairline visible. Jamais de trou noir, jamais de fausses données.
export function EmptyChart({ hint }: { hint: string }) {
  return (
    <div
      aria-hidden
      style={{
        position: 'absolute',
        inset: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        pointerEvents: 'none',
      }}
    >
      <span
        className="mono"
        style={{
          fontFamily: 'var(--font-mono, JetBrains Mono)',
          fontSize: 10,
          letterSpacing: '0.1em',
          textTransform: 'uppercase',
          color: 'var(--text-faint, #767b84)',
          border: '1px dashed var(--hairline, #26262a)',
          padding: '8px 14px',
          background: 'rgba(7, 7, 7, 0.6)',
        }}
      >
        {hint}
      </span>
    </div>
  )
}
