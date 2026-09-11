export type ProvState = 'live' | 'err' | 'wait';
export interface ProvInfo {
  source: string;
  refresh?: string | number;
  refreshSec?: number;
  n?: number | string;
  state?: ProvState;
  extra?: string;
}
export function formatProvenance(p: ProvInfo): string {
  const parts = [`source : ${p.source}`];
  const raw = p.refresh ?? p.refreshSec;
  if (raw !== undefined) {
    const r = typeof raw === 'number' ? `${raw} s` : String(raw).trim();
    // mot gelé verbatim (`au gel`) — le ` s` ne se colle qu'aux nombres
    const rNorm = /\d/.test(r) ? (r.endsWith('s') ? r : `${r} s`) : r;
    parts.push(`rafraîchi ${rNorm}`);
  } else {
    parts.push(`rafraîchi 10 s`);
  }
  if (p.n !== undefined) parts.push(`n=${p.n}`);
  if (p.extra) parts.push(p.extra);
  return parts.join(' · ');
}
export function Provenance(props: ProvInfo & { className?: string }) {
  const state = props.state ?? 'wait';
  // lisible même sans connaître le banc : 11px (pas micro), corps (pas faint),
  // mots français — la provenance est une garantie, pas un code secret.
  return <p className={`c-prov c-provenance ${state} ${props.className ?? ''}`.trim()} data-testid="provenance" aria-label="source de données" style={{ fontSize: 11, color: state === 'live' ? '#7fd6e8' : '#a9aeb6', letterSpacing: '0.03em', marginTop: 10 }}>{formatProvenance(props)}</p>;
}
