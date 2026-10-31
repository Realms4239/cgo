import { useEffect, useRef, useState } from 'react';
import { useUIStore, type ModeId } from '../store/ui';
import { loadCapabilities, type Capabilities } from './capabilities';

// Une destination masquée ne doit rien peindre ni rien sonder : seul le
// service racine (session d'audit, SSE) continue.
export function useModeActive(mode: ModeId): boolean {
  return useUIStore(s => s.mode === mode);
}

// Capacités de l'hôte (files, contrôles, noyau) — lecture unique partagée.
export function useCapabilities(): Capabilities | null {
  const [caps, setCaps] = useState<Capabilities | null>(null)
  useEffect(() => {
    let alive = true
    loadCapabilities().then(c => { if (alive) setCaps(c) })
    return () => { alive = false }
  }, [])
  return caps
}

// Hook de boucle rAF générique — port React du paintLoop d'origine.
// Le callback lit le singleton `live` mutable et peint de façon impérative
// (ECharts setOption / canvas / refs DOM): React reste hors du chemin 10 Hz.
export function useRafLoop(cb: (ts: number) => void, active = true) {
  const cbRef = useRef(cb);
  cbRef.current = cb;
  useEffect(() => {
    if (!active) return;
    let raf = 0;
    let dead = false;
    const loop = (ts: number) => {
      try {
        cbRef.current(ts);
      } catch (e) {
        // une frame de rendu ratée ne doit jamais tuer la boucle — les graphes geleraient
        if (!dead) {
          dead = true;
          const st = (e as Error)?.stack ?? String(e);
          console.error('[raf] paint failed:', st.split('\n').slice(0, 4).join(' | '));
        }
      }
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, [active]);
}

// Hook d'intervalle, même motif de ref (horloge 1 Hz, CDF 5 s, portes 10 s).
export function useInterval(cb: () => void, ms: number, active = true) {
  const cbRef = useRef(cb);
  cbRef.current = cb;
  useEffect(() => {
    if (!active) return;
    const id = setInterval(() => cbRef.current(), ms);
    return () => clearInterval(id);
  }, [ms, active]);
}
