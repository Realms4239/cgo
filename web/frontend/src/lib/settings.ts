// Settings utilisateur — 10 paramètres, chaque borne est un vrai choix de
// mesure. Persistés en localStorage; les bornes serveur restent autoritaires.
// Chaque sauvegarde diffuse `meteolink-settings` — les vues abonnées se
// re-rendent immédiatement (les seuils de sévérité ne sont jamais périmés).
export type Settings = {
  warnMs: number        // seuil « dégradé » latence
  critMs: number        // seuil « critique » latence
  deadlineMs: number    // deadline small p95 (passée à run/start)
  ringMax: number       // fenêtre du tableau live
  watchMs: number       // cadence surveillance
  shapeCap: number      // capacité du bord par défaut
  linkDelayMs: number   // délai du lien (levier)
  linkJitterMs: number  // gigue (levier)
  linkLossPct: number   // perte (levier)
  target: string        // cible de mesure
}

export const DEFAULTS: Settings = {
  warnMs: 40, critMs: 100, deadlineMs: 1000, ringMax: 1800, watchMs: 500,
  shapeCap: 20, linkDelayMs: 20, linkJitterMs: 2, linkLossPct: 0, target: '1.1.1.1',
}

const KEY = 'meteolink-settings'

export function loadSettings(): Settings {
  try {
    const raw = localStorage.getItem(KEY)
    if (!raw) return { ...DEFAULTS }
    return { ...DEFAULTS, ...JSON.parse(raw) }
  } catch { return { ...DEFAULTS } }
}

export function saveSettings(s: Settings) {
  try { localStorage.setItem(KEY, JSON.stringify(s)) } catch { }
  try { window.dispatchEvent(new CustomEvent<Settings>('meteolink-settings', { detail: s })) } catch { }
}

export type Level = 'ok' | 'warn' | 'crit'

// Taxonomie de sévérité uniforme — une seule source pour toutes les cartes.
export function latencyLevel(ms: number, s: Settings): Level {
  if (ms <= s.warnMs) return 'ok'
  if (ms <= s.critMs) return 'warn'
  return 'crit'
}

export function deadlineLevel(pct: number): Level {
  if (pct >= 95) return 'ok'
  if (pct >= 80) return 'warn'
  return 'crit'
}

export function jfiLevel(v: number): Level {
  if (v > 0.95) return 'ok'
  if (v >= 0.8) return 'warn'
  return 'crit'
}

export function goodputLevel(mbps: number, capacity: number): Level {
  const pct = capacity > 0 ? (mbps / capacity) * 100 : 0
  if (pct >= 50) return 'ok'
  if (pct >= 20) return 'warn'
  return 'crit'
}

export function dropsLevel(drops: number): Level {
  if (drops <= 0) return 'ok'
  if (drops <= 100) return 'warn'
  return 'crit'
}

export const LEVEL_COLOR: Record<Level, string> = {
  ok: '#1fa348', warn: '#f4b400', crit: '#e84a3a',
}
