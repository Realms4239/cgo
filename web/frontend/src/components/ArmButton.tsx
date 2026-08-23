import { useState, useEffect } from 'react'

export function ArmButton({ label, confirmLabel = 'CONFIRMER ?', onConfirm, disabled }: { label: string; confirmLabel?: string; onConfirm: () => void; disabled?: boolean }) {
  const [armed, setArmed] = useState(false)
  useEffect(() => {
    if (!armed) return
    const t = setTimeout(() => setArmed(false), 5000)
    return () => clearTimeout(t)
  }, [armed])
  if (armed) {
    return (
      <button className="btn btn-danger" onClick={() => { setArmed(false); onConfirm() }} onBlur={() => setArmed(false)}>
        {confirmLabel}
      </button>
    )
  }
  return <button className="btn btn-primary" disabled={disabled} onClick={() => setArmed(true)}>{label}</button>
}
