// Preuves & Rapports — le corpus gelé d'abord, les rapports d'audit ensuite.
import AuditReportsPanel from '../components/AuditReportsPanel'
import { ModeHead } from '../components/ModeHead'
import IntegriteView from './IntegriteView'
import { scrollToId } from '../lib/scroll'
import { useModeActive } from '../lib/hooks'

export default function EvidenceView() {
  const active = useModeActive('evidence')
  return (
    <div className="panel-stack evidence-page">
      <ModeHead
        title="Preuves & Rapports"
        sub="Manifestes, quarantaine, empreintes — et les rapports citables."
      />
      <nav className="bench-anchors evidence-anchors" aria-label="Sections des preuves">
        <button className="btn" onClick={() => scrollToId('evidence-corpus')}>CORPUS &amp; PREUVES</button>
        <button className="btn" onClick={() => scrollToId('evidence-reports')}>RAPPORTS D'AUDIT</button>
      </nav>
      <div id="evidence-corpus"><IntegriteView embedded active={active} /></div>
      <div id="evidence-reports"><AuditReportsPanel /></div>
    </div>
  )
}
