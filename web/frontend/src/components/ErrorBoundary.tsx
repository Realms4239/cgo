import { Component, ErrorInfo, ReactNode } from 'react';

interface Props { children: ReactNode; }
interface State { hasError: boolean; error?: Error; }

const CHUNK_ERR = 'Failed to fetch dynamically imported module';
const isChunkLoadError = (err: unknown): boolean =>
  !!err && typeof (err as Error).message === 'string' && (err as Error).message.includes(CHUNK_ERR);
const RELOAD_GUARD_KEY = 'cgo.reloaded';

class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };
  static getDerivedStateFromError(error: Error): State { return { hasError: true, error }; }
  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary:', error, info);
    // Un chunk d'UI qui échoue (502/flux coupé côté tunnel) se répare seul :
    // un rechargement de page, une seule fois par session (pas de boucle).
    if (isChunkLoadError(error) && !sessionStorage.getItem(RELOAD_GUARD_KEY)) {
      sessionStorage.setItem(RELOAD_GUARD_KEY, '1');
      window.setTimeout(() => window.location.reload(), 150);
    }
  }
  render() {
    if (this.state.hasError) {
      return <div className="error-boundary"><h1>Erreur</h1><pre>{this.state.error?.message}</pre>
        <div className="eb-actions">
          <button className="side-btn" onClick={() => this.setState({ hasError: false })}>Réessayer</button>
          <button className="side-btn" onClick={() => window.location.reload()}>Recharger la page</button>
        </div></div>;
    }
    return this.props.children;
  }
}
export default ErrorBoundary;