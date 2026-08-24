export type EmptyKind = 'idle' | 'loading' | 'empty' | 'stale' | 'offline' | 'error';
const COPY: Record<EmptyKind, string> = {
  idle: 'en attente', loading: 'chargement…', empty: 'aucune donnée',
  stale: 'données périmées', offline: 'backend injoignable', error: 'erreur de source',
};
export function EmptyState({ kind = 'empty', hint }: { kind?: EmptyKind; hint?: string }) {
  return (
    <div className={`c-empty ${kind}`} role="status">
      <span className="c-empty-mark" aria-hidden>—</span>
      <span className="c-empty-lbl">{COPY[kind]}</span>
      {hint ? <span className="c-empty-hint">{hint}</span> : null}
    </div>
  );
}
