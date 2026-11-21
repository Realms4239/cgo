// Sceau — « VÉRIFIÉ » n'apparaît qu'après un acte réel : une requête qui répond
// disponible ET une empreinte complète de 64 caractères hexadécimaux. La
// présence d'un champ ne prouve rien toute seule.
import { useState } from 'react'

const SHA256 = /^[0-9a-f]{64}$/i

export function integrityVerified(payload: unknown): boolean {
  if (!payload || typeof payload !== 'object') return false
  const p = payload as { available?: unknown; sha256?: unknown }
  return p.available === true && typeof p.sha256 === 'string' && SHA256.test(p.sha256)
}

export function SealBadge({ verified }: { verified: boolean }) {
  return (
    <span className="seal-badge" data-state={verified ? 'verified' : 'unverified'}>
      {verified ? 'VÉRIFIÉ' : 'À VÉRIFIER'}
    </span>
  )
}

export default function SealInspector() {
  const [verified, setVerified] = useState(false)
  const [hash, setHash] = useState<string | null>(null)
  const [msg, setMsg] = useState('Sceau non vérifié.')

  const verify = async () => {
    setMsg('Vérification en cours…')
    try {
      const r = await fetch('/api/integrity')
      if (!r.ok) {
        setVerified(false)
        setMsg('Sceau injoignable.')
        return
      }
      const j = await r.json()
      const ok = integrityVerified(j)
      setVerified(ok)
      setHash(j && typeof j.sha256 === 'string' ? j.sha256 : null)
      setMsg(ok
        ? 'Intégrité vérifiée sur une empreinte complète.'
        : 'Empreinte incomplète ou source indisponible.')
    } catch {
      setVerified(false)
      setMsg('Sceau injoignable.')
    }
  }

  return (
    <section className="seal-inspector" aria-label="Vérification du sceau">
      <SealBadge verified={verified} />
      <button type="button" className="btn btn-primary seal-verify" onClick={() => void verify()}>
        Vérifier le sceau
      </button>
      {hash ? <code className="seal-hash" title={hash}>{hash.slice(0, 8)}…</code> : null}
      <p className="seal-msg" role="status" aria-live="polite">{msg}</p>
    </section>
  )
}
