import { useUIStore } from '../store/ui'

// Copie honnête : une API absente ou un rejet affiche toujours un toast d'erreur.
// Une seule source pour le tiroir de preuve et les autres boutons de copie.
// Le verdict est renvoyé pour que l'appelant puisse le dire aussi en `role=status`.
export function copyText(label: string, text: string): Promise<boolean> {
  const done = () => { useUIStore.getState().pushToast(`${label} copié`, 'ok'); return true }
  const fail = () => { useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err'); return false }
  const clip = typeof navigator !== 'undefined' ? navigator.clipboard : undefined
  if (!clip || typeof clip.writeText !== 'function') { fail(); return Promise.resolve(false) }
  return clip.writeText(text).then(done, fail)
}
