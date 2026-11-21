import { useUIStore } from '../store/ui'

// Copie honnête : une API absente ou un rejet affiche toujours un toast d'erreur.
// Une seule source pour le tiroir de preuve et les autres boutons de copie.
export function copyText(label: string, text: string) {
  const done = () => useUIStore.getState().pushToast(`${label} copié`, 'ok')
  const fail = () => useUIStore.getState().pushToast('Copie impossible — sélectionnez le texte', 'err')
  const clip = typeof navigator !== 'undefined' ? navigator.clipboard : undefined
  if (!clip || typeof clip.writeText !== 'function') { fail(); return }
  clip.writeText(text).then(done, fail)
}
