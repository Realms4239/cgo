import { useState, useEffect, useRef } from 'react'
import { animate } from 'animejs'
import { animateArmButton, prefersReducedMotion } from '../lib/anime'

export function ArmButton({ label, confirmLabel = 'CONFIRMER ?', onConfirm, disabled }: { label: string; confirmLabel?: string; onConfirm: () => void; disabled?: boolean }) {
  const [armed, setArmed] = useState(false)
  const btnRef = useRef<HTMLButtonElement>(null)
  const progressRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!armed) return
    const t = setTimeout(() => setArmed(false), 5000)
    return () => clearTimeout(t)
  }, [armed])
  useEffect(() => {
    if (!armed || !btnRef.current) return
    // primary white → danger red spring scale 0.96→1 rotate 0.5→-0.5 + shadow pulse
    animateArmButton(btnRef.current)
    if (!prefersReducedMotion()) {
      animate(btnRef.current, { rotate: [0.5, -0.5, 0], duration: 400, ease: 'outElastic(1, .6)' } as any)
      animate(btnRef.current, { boxShadow: ['0 0 0 rgba(226,39,24,0)', '0 0 16px rgba(226,39,24,0.35)', '0 0 0 rgba(226,39,24,0)'], duration: 800, ease: 'linear' } as any)
    }
    if (progressRef.current) {
      progressRef.current.style.transformOrigin = 'left center'
      progressRef.current.style.transform = 'scaleX(1)'
      if (!prefersReducedMotion()) {
        animate(progressRef.current, { scaleX: [1, 0], duration: 5000, ease: 'linear' } as any)
      } else {
        // mouvement réduit: ligne statique 1 px, sans drainage
      }
    }
  }, [armed])
  if (armed) {
    return (
      <span style={{ position: 'relative', display: 'inline-block' }}>
        <div
          ref={progressRef}
          className="arm-progress"
          aria-hidden="true"
          style={{
            position: 'absolute',
            top: -4,
            left: 0,
            width: '100%',
            height: 1,
            background: 'var(--t-danger, #e22718)',
            transformOrigin: 'left center',
          }}
        />
        <button ref={btnRef} className="btn btn-danger" onClick={() => { setArmed(false); onConfirm() }} onBlur={() => setArmed(false)}>
          {confirmLabel}
        </button>
      </span>
    )
  }
  return <button className="btn btn-primary" disabled={disabled} onClick={() => setArmed(true)}>{label}</button>
}
