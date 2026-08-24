export type ProvState = 'live' | 'err' | 'wait';
export interface ProvInfo {
  source: string;
  refresh?: string | number;
  refreshSec?: number;
  n?: number | string;
  state?: ProvState;
  extra?: string;
}
function formatProvenance(p: ProvInfo): string {
  const parts = [`source : ${p.source}`];
  const raw = p.refresh ?? p.refreshSec;
  if (raw !== undefined) {
    const r = typeof raw === 'number' ? `${raw} s` : String(raw).trim();
    const rNorm = r.endsWith('s') ? r : `${r} s`;
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
  return <p className={`c-prov c-provenance ${state} ${props.className ?? ''}`.trim()} data-testid="provenance" aria-label="source de données">{formatProvenance(props)}</p>;
}
