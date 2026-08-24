import { useEffect, useRef } from 'react';

// Generic rAF loop hook — the React port of the original paintLoop model.
// The callback reads the mutable `live` singleton and paints imperatively
// (ECharts setOption / canvas / DOM refs), keeping React out of the 10 Hz path.
export function useRafLoop(cb: (ts: number) => void, active = true) {
  const cbRef = useRef(cb);
  cbRef.current = cb;
  useEffect(() => {
    if (!active) return;
    let raf = 0;
    const loop = (ts: number) => {
      cbRef.current(ts);
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, [active]);
}

// Interval hook with the same ref pattern (for 1 Hz clock, 5 s CDF poll, 10 s gates).
export function useInterval(cb: () => void, ms: number, active = true) {
  const cbRef = useRef(cb);
  cbRef.current = cb;
  useEffect(() => {
    if (!active) return;
    const id = setInterval(() => cbRef.current(), ms);
    return () => clearInterval(id);
  }, [ms, active]);
}
