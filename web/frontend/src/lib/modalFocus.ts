import { useEffect, useRef, type RefObject } from 'react'

// Les raccourcis de destinations ne doivent jamais partir d’un champ de saisie.
export function isTextEntryTarget(target: EventTarget | null): boolean {
  if (!target || !(target instanceof HTMLElement)) return false
  if (target.closest('input, textarea, select, [contenteditable="true"], [role="dialog"]')) return true
  return target.isContentEditable === true
}

// Contrat d'un dialogue modal : l'arrière-plan (#shell) devient inerte, le
// focus entre dans le panneau, y reste (Tab/Shift-Tab), Échap ferme le
// panneau du dessus seulement, et le focus revient au déclencheur. Empilés,
// l'arrière-plan reste inerte jusqu'à la fermeture du dernier panneau.
// Le panneau doit vivre HORS de #shell
// (createPortal) : sinon il s'inertait lui-même — panneau infocusable,
// inclickable, constaté sur la palette avant ce correctif.
const FOCUSABLE = 'input:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])'

type ModalEntry = {
  panel: RefObject<HTMLElement | null>
  closeRef: { current: () => void }
  trigger: HTMLElement | null
}

// Une pile de dialogues : Échap et Tab ne parlent qu'au dernier ouvert,
// l'arrière-plan reste inerte tant qu'il reste au moins un panneau.
const stack: ModalEntry[] = []

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
    const trigger = document.activeElement as HTMLElement | null
    const shell = document.getElementById('shell')
    const entry: ModalEntry = { panel, closeRef, trigger }
    stack.push(entry)
    if (shell) (shell as unknown as { inert: boolean }).inert = true

    const timer = window.setTimeout(() => {
      if (stack[stack.length - 1] !== entry) return
      const first = panel.current?.querySelector<HTMLElement>(FOCUSABLE)
      first?.focus()
    }, 0)
    const onKey = (e: KeyboardEvent) => {
      if (stack[stack.length - 1] !== entry) return
      if (e.key === 'Escape') {
        closeRef.current()
        return
      }
      if (e.key !== 'Tab' || !panel.current) return
      const els = Array.from(panel.current.querySelectorAll<HTMLElement>(FOCUSABLE))
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
      const i = stack.indexOf(entry)
      if (i >= 0) stack.splice(i, 1)
      window.clearTimeout(timer)
      document.removeEventListener('keydown', onKey)
      if (stack.length === 0 && shell) (shell as unknown as { inert: boolean }).inert = false
      trigger?.focus?.()
    }
  }, [open, panel])
}
