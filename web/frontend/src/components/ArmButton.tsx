import { useState, useEffect, useRef } from 'react'
import { animateArmButton } from '../lib/anime'

export function ArmButton({ label, confirmLabel = 'CONFIRMER ?', onConfirm, disabled }: { label: string; confirmLabel?: string; onConfirm: () => void; disabled?: boolean }) {
  const [armed, setArmed] = useState(false)
  const btnRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    if (!armed) return
    const t = setTimeout(() => setArmed(false), 5000)
    return () => clearTimeout(t)
  }, [armed])
  useEffect(() => {
    if (armed && btnRef.current) animateArmButton(btnRef.current)
  }, [armed])
  if (armed) {
    return (
      <button ref={btnRef} className="btn btn-danger" onClick={() => { setArmed(false); onConfirm() }} onBlur={() => setArmed(false)}>
        {confirmLabel}
      </button>
    )
  }
  return <button className="btn btn-primary" disabled={disabled} onClick={() => setArmed(true)}>{label}</button>
}
