// Bandeau de réponse sous 1024 px : c'est le même nœud de DecisionSpine qui se
// replie, jamais un second mot de verdict. Au-dessus, l'enveloppe ne change
// rien à la disposition (display: contents).
import type { ReactNode } from 'react'

export default function AnswerBand({ children }: { children: ReactNode }) {
  return <div className="answer-band-stacked">{children}</div>
}
