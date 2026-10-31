import Explain from './Explain'
import type { Capabilities } from '../lib/capabilities'

// Les axes du banc viennent de l'hôte : le client ne recopie ni les files ni
// les contrôles. BBRv3 est nommé absent au lieu d'être omis en silence.
export function BenchCapabilities({ caps }: { caps: Capabilities | null }) {
  const qdiscs = caps?.qdiscs ?? []
  const ccs = caps?.ccs ?? []
  return (
    <>
      <div className="form-row">
        <label>Files testées</label>
        <span className="mono muted" data-testid="bench-qdiscs">
          {caps
            ? qdiscs.map((q, i) => (
                <span key={q}>
                  <Explain term={q}>{q}</Explain>
                  {i < qdiscs.length - 1 ? ' · ' : ''}
                </span>
              ))
            : 'indisponible'}
        </span>
      </div>
      <div className="form-row">
        <label>Contrôles</label>
        <span className="mono muted" data-testid="bench-ccs">
          {caps
            ? ccs.map((c, i) => (
                <span key={c}>
                  <Explain term={c}>{c}</Explain>
                  {i < ccs.length - 1 ? ' · ' : ''}
                </span>
              ))
            : 'indisponible'}
        </span>
        {caps && !caps.bbr3 ? (
          <span
            className="mono muted bench-hint"
            data-testid="bench-bbr3-hint"
            title={caps.kernel ? `noyau ${caps.kernel}` : undefined}
          >
            BBRv3 absent de ce noyau · kit/install-bbr3.sh
          </span>
        ) : null}
      </div>
    </>
  )
}
