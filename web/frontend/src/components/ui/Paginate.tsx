// Paginate — pagination intelligente 1 … 4 5 6 … n (U6h) : anti-overflow
// pour les longues listes. Exporte aussi le slice zéro-index.
export function paginate<T>(arr: T[], page: number, size: number): T[] {
  return arr.slice((page - 1) * size, page * size)
}

export function pageList(page: number, totalPages: number): (number | '…')[] {
  const out = new Set<number>([1, totalPages, page, page - 1, page + 1])
  const list: (number | '…')[] = []
  let prev = 0
  for (let i = 1; i <= totalPages; i++) {
    if (!out.has(i)) continue
    if (i - prev > 1) list.push('…')
    list.push(i)
    prev = i
  }
  return list
}

export function Paginate({ total, page, pageSize, onPage }: {
  total: number
  page: number
  pageSize: number
  onPage: (p: number) => void
}) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  if (totalPages <= 1) return null
  const btn = (label: string, p: number | null, active: boolean, disabled: boolean) => (
    <button
      key={label} className="btn mono" onClick={() => p != null && onPage(p)}
      disabled={disabled}
      style={{
        padding: '3px 10px', fontSize: 11, fontFamily: 'var(--font-mono)',
        border: '1px solid ' + (active ? '#3a3a40' : 'var(--hairline)'),
        background: active ? 'rgba(90,211,227,0.12)' : 'transparent',
        color: active ? '#7fd6e8' : disabled ? 'var(--text-faint)' : '#a9aeb6',
        cursor: disabled ? 'default' : 'pointer',
      }}
    >{label}</button>
  )
  return (
    <div style={{ display: 'flex', gap: 4, alignItems: 'center', flexWrap: 'wrap', marginTop: 8 }} data-testid="paginate">
      {btn('← précédent', page - 1, false, page <= 1)}
      {pageList(page, totalPages).map((p, i) =>
        p === '…'
          ? <span key={`e${i}`} className="mono" style={{ fontSize: 11, color: 'var(--text-faint)', padding: '0 2px' }}>…</span>
          : btn(String(p), p, p === page, false))}
      {btn('suivant →', page + 1, false, page >= totalPages)}
      <span className="mono" style={{ fontSize: 10, color: 'var(--text-faint)', marginLeft: 6, fontVariantNumeric: 'tabular-nums' }}>{total} au total</span>
    </div>
  )
}
