import { useEffect, useRef } from 'react';

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
