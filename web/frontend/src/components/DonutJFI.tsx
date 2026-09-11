// Arc 0–360 (56×56, épaisseur 8) pour JFI 0–1. Pas de chiffre au centre :
// la valeur vit déjà à côté (carte JFI) — l'anneau seul, jamais en double.
export function DonutJFI({ value }: { value: number | null }) {
  if (value == null || !Number.isFinite(value)) {
    return (
      <svg width={56} height={56} viewBox="0 0 56 56" aria-label="JFI —">
        <circle cx={28} cy={28} r={22} fill="none" stroke="rgba(154,163,173,0.12)" strokeWidth={8} />
      </svg>
    )
  }
  const v = Math.max(0, Math.min(1, value))
  const circ = 176
  const dash = `${v * circ} ${circ}`
  const color = v > 0.95 ? '#1fa348' : '#9aa3ad'
  return (
    <svg width={56} height={56} viewBox="0 0 56 56" aria-label={`JFI ${v.toFixed(2)}`}>
      <circle cx={28} cy={28} r={22} fill="none" stroke="rgba(154,163,173,0.15)" strokeWidth={8} />
      <circle
        cx={28}
        cy={28}
        r={22}
        fill="none"
        stroke={color}
        strokeWidth={8}
        strokeLinecap="round"
        strokeDasharray={dash}
        transform="rotate(-90 28 28)"
        style={{ transition: 'stroke-dasharray 0.4s ease' }}
      />
    </svg>
  )
}
