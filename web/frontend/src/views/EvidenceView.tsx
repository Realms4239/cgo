// Preuves & Rapports — le corpus gelé d'abord, les rapports d'audit ensuite.
import AuditReportsPanel from '../components/AuditReportsPanel'
import IntegriteView from './IntegriteView'
import { scrollToId } from '../lib/scroll'

export default function EvidenceView() {
  return (
    <div className="panel-stack evidence-page">
      <header className="mode-head">
        <div>
          <h1>Preuves &amp; Rapports</h1>
          <p>
            Ce qui est archivé et vérifiable : manifestes scellés, quarantaine, empreintes, et les
            rapports d'audit citables.
          </p>
        </div>
      </header>
      <nav className="bench-anchors evidence-anchors" aria-label="Sections des preuves">
        <button className="btn" onClick={() => scrollToId('evidence-corpus')}>CORPUS &amp; PREUVES</button>
        <button className="btn" onClick={() => scrollToId('evidence-reports')}>RAPPORTS D'AUDIT</button>
      </nav>
      <div id="evidence-corpus"><IntegriteView embedded /></div>
      <div id="evidence-reports"><AuditReportsPanel /></div>
    </div>
  )
}
