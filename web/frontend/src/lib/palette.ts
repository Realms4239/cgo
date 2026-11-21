// palette — la SEULE maison des couleurs concrètes (canvas ECharts, badges
// tiers). Les valeurs viennent des tokens CSS à l'exécution ; hors DOM
// (tests, agrégation Node) le repli gelé ci-dessous fait foi — et un test
// vérifie qu'il égale tokens.css. Les composants .tsx n'écrivent plus de hex :
// soit `TOKEN.x` (var() pour les styles CSS), soit `palette()` (canvas).

export const TOKEN = {
  canvas: 'var(--canvas-solid)',
  surfaceSoft: 'var(--surface-soft)',
  surfaceCard: 'var(--surface-card)',
  surfaceElevated: 'var(--surface-elevated)',
  hairline: 'var(--hairline)',
  hairlineStrong: 'var(--hairline-strong)',
  hairlineFaint: 'var(--hairline-faint)',
  textPrimary: 'var(--text-primary)',
  textBodyStrong: 'var(--text-body-strong)',
  textBody: 'var(--text-body)',
  textMuted: 'var(--text-muted)',
  textFaint: 'var(--text-faint)',
  live: 'var(--t-live)',
  threshold: 'var(--t-threshold)',
  danger: 'var(--t-danger)',
  dangerText: 'var(--t-danger-text)',
  ok: 'var(--t-ok)',
  bbr: 'var(--t-bbr)',
  neutral: 'var(--t-neutral)',
  brand1: 'var(--brand-1)',
  brand2: 'var(--brand-2)',
  brand3: 'var(--brand-3)',
} as const

export type TokenName = keyof typeof TOKEN

// Le genre d'une notification désigne un jeton sémantique, jamais un hex :
// la bascule claire (T9) est appliquée dans la feuille, pas dans les composants.
export const KIND_TOKEN: Record<'' | 'ok' | 'err' | 'blue', TokenName> = {
  '': 'textMuted',
  ok: 'ok',
  err: 'danger',
  blue: 'live',
}

// Teinte alpha d'une couleur résolue — ECharts/SVG n'acceptent pas var().
// N'accepte qu'un hex (tous les jetons concrets le sont).
export function alpha(hex: string, a: number): string {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${a})`
}

// Rampe des gris de texte verrouillés — les gris hors jeton y convergent
// (jamais une 4e teinte : #333333 / #5a5a5a / #6d6d6d).
export const OFF_TOKEN_GRAYS: Record<string, TokenName> = {
  '#333333': 'textPrimary',
  '#5a5a5a': 'textBody',
  '#6d6d6d': 'textMuted',
}

// Repli hors DOM — doit égaler tokens.css (assertion de test).
export const FALLBACK = {
  '--canvas-solid': '#fbfbfb',
  '--surface-soft': '#f3f3f3',
  '--surface-card': '#ffffff',
  '--surface-elevated': '#ffffff',
  '--hairline': '#e8e8e8',
  '--hairline-strong': '#c8c8c8',
  '--hairline-faint': '#f3f3f3',
  '--text-primary': '#333333',
  '--text-body-strong': '#5a5a5a',
  '--text-body': '#5a5a5a',
  '--text-muted': '#6d6d6d',
  '--text-faint': '#6d6d6d',
  '--t-live': '#17938a',
  '--t-threshold': '#806400',
  '--t-danger': '#d51600',
  '--t-danger-text': '#d51600',
  '--t-ok': '#1e7a3c',
  '--t-bbr': '#34495e',
  '--t-neutral': '#6d6d6d',
  '--brand-1': '#17938a',
  '--brand-2': '#2dcc70',
  '--brand-3': '#f4c910',
  '--chart-cake': '#17938a',
  '--chart-fq-codel': '#34495e',
  '--chart-pfifo': '#6d6d6d',
  '--chart-cubic': '#1e7a3c',
  '--chart-bbr': '#34495e',
  '--chart-bbr3': '#9a7400',
  '--chart-profile-1': '#17938a',
  '--chart-profile-2': '#1e7a3c',
  '--chart-profile-3': '#806400',
  '--chart-profile-4': '#34495e',
  '--chart-profile-5': '#d51600',
  '--chart-profile-6': '#6d6d6d',
  '--chart-on-ok': '#ffffff',
} as const

export function cssVar(name: keyof typeof FALLBACK): string {
  if (typeof window === 'undefined' || typeof document === 'undefined') return FALLBACK[name]
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return v || FALLBACK[name]
}

export interface Palette {
  canvas: string
  surfaceSoft: string
  surfaceElevated: string
  hairline: string
  hairlineStrong: string
  hairlineFaint: string
  textPrimary: string
  textBodyStrong: string
  textBody: string
  textMuted: string
  textFaint: string
  live: string
  threshold: string
  danger: string
  dangerText: string
  ok: string
  bbr: string
  neutral: string
  onOk: string
  qdisc: Record<string, string>
  cc: Record<string, string>
  profile: string[]
}

// Cache module : le canvas lit les tokens une fois (les thèmes ne changent
// pas à chaud) — évite un getComputedStyle par rendu à 2 Hz.
let cached: Palette | null = null
export function paletteOnce(): Palette {
  if (!cached) cached = palette()
  return cached
}

// Couleurs concrètes pour le canvas — jamais un var() dans ECharts.
export function palette(): Palette {
  return {
    canvas: cssVar('--canvas-solid'),
    surfaceSoft: cssVar('--surface-soft'),
    surfaceElevated: cssVar('--surface-elevated'),
    hairline: cssVar('--hairline'),
    hairlineStrong: cssVar('--hairline-strong'),
    hairlineFaint: cssVar('--hairline-faint'),
    textPrimary: cssVar('--text-primary'),
    textBodyStrong: cssVar('--text-body-strong'),
    textBody: cssVar('--text-body'),
    textMuted: cssVar('--text-muted'),
    textFaint: cssVar('--text-faint'),
    live: cssVar('--t-live'),
    threshold: cssVar('--t-threshold'),
    danger: cssVar('--t-danger'),
    dangerText: cssVar('--t-danger-text'),
    ok: cssVar('--t-ok'),
    bbr: cssVar('--t-bbr'),
    neutral: cssVar('--t-neutral'),
    onOk: cssVar('--chart-on-ok'),
    qdisc: {
      cake: cssVar('--chart-cake'),
      fq_codel: cssVar('--chart-fq-codel'),
      pfifo_fast: cssVar('--chart-pfifo'),
    },
    cc: { cubic: cssVar('--chart-cubic'), bbr: cssVar('--chart-bbr'), bbr3: cssVar('--chart-bbr3') },
    profile: [
      cssVar('--chart-profile-1'), cssVar('--chart-profile-2'), cssVar('--chart-profile-3'),
      cssVar('--chart-profile-4'), cssVar('--chart-profile-5'), cssVar('--chart-profile-6'),
    ],
  }
}
