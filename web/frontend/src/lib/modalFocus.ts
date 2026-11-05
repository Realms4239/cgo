import { useEffect, useRef, type RefObject } from 'react'

// Contrat d'un dialogue modal : l'arrière-plan (#shell) devient inerte, le
// focus entre dans le panneau, y reste (Tab/Shift-Tab), Échap ferme, et le
// focus revient au déclencheur. Le panneau doit vivre HORS de #shell
// (createPortal) : sinon il s'inertait lui-même — panneau infocusable,
// inclickable, constaté sur la palette avant ce correctif.
export function useModalFocus(
  open: boolean,
  onClose: () => void,
  panel: RefObject<HTMLElement | null>,
) {
  // onClose change d'identité à chaque rendu : on lit la dernière version
  // sans relancer l'effet (sinon ré-inertisation + vol de focus en boucle).
  const closeRef = useRef(onClose)
  closeRef.current = onClose

  useEffect(() => {
    if (!open) return
    const previous = document.activeElement as HTMLElement | null
    const shell = document.getElementById('shell')
    if (shell) (shell as unknown as { inert: boolean }).inert = true
    const timer = window.setTimeout(() => {
      const first = panel.current?.querySelector<HTMLElement>(
        'input:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])',
      )
      first?.focus()
    }, 0)
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        closeRef.current()
        return
      }
      if (e.key !== 'Tab' || !panel.current) return
      const els = Array.from(
        panel.current.querySelectorAll<HTMLElement>(
          'input:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ),
      )
      if (els.length === 0) return
      const i = els.indexOf(document.activeElement as HTMLElement)
      if (e.shiftKey && i <= 0) {
        els[els.length - 1].focus()
        e.preventDefault()
      } else if (!e.shiftKey && i === els.length - 1) {
        els[0].focus()
        e.preventDefault()
      }
    }
    document.addEventListener('keydown', onKey)
    return () => {
      window.clearTimeout(timer)
      document.removeEventListener('keydown', onKey)
      if (shell) (shell as unknown as { inert: boolean }).inert = false
      previous?.focus?.()
    }
  }, [open, panel])
}
