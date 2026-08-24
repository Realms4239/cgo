export type PipState = 'ok' | 'warn' | 'err' | 'idle' | 'on' | 'live' | 'offline' | 'error' | 'success' | 'empty' | 'loading' | 'stale';

const PIP_MAP: Record<string, string> = {
  ok: 'ok', success: 'ok',
  warn: 'warn', stale: 'warn',
  err: 'err', error: 'err', offline: 'err',
  on: 'on', live: 'on', loading: 'on',
  idle: 'idle', empty: 'idle',
};

export function StatusPip({ state, label, title }: { state: PipState; label?: string; title?: string }) {
  const cls = PIP_MAP[state] ?? 'idle';
  return (
    <span className={`c-pip ${cls}`} title={title ?? label} aria-label={title ?? label}>
      <i className="c-pip-dot" aria-hidden />
      {label ? <span className="c-pip-lbl">{label}</span> : null}
    </span>
  );
}

export function Badge({ children, tone = 'steel', className = '' }: { children: React.ReactNode; tone?: 'steel' | 'success' | 'warning' | 'danger' | 'live'; className?: string }) {
  return <span className={`c-badge c-badge--${tone}${className ? ' ' + className : ''}`}>{children}</span>;
}

export function Chip({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return <span className={`c-chip${className ? ' ' + className : ''}`}>{children}</span>;
}
