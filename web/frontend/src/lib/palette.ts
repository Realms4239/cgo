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

// Teinte alpha d'une couleur résolue — ECharts/SVG n'acceptent pas var().
// N'accepte qu'un hex (tous les jetons concrets le sont).
export function alpha(hex: string, a: number): string {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${a})`
}

// Rampe de 5 textes — les gris hors jeton y convergent (jamais une 6e teinte).
export const OFF_TOKEN_GRAYS: Record<string, TokenName> = {
  '#f2f2f4': 'textPrimary',
  '#d6d8dd': 'textBodyStrong',
  '#d6dade': 'textBodyStrong',
  '#d8dde6': 'textBodyStrong',
  '#c3c9d1': 'textBodyStrong',
  '#a9aeb6': 'textBody',
  '#a8aeb7': 'textBody',
  '#9aa0a8': 'textBody',
  '#9aa3ad': 'neutral',
  '#8b9099': 'textMuted',
  '#7e838c': 'textFaint',
  '#767b84': 'textFaint',
  '#6b7078': 'textFaint',
}

// Repli hors DOM — doit égaler tokens.css (assertion de test).
export const FALLBACK = {
  '--canvas-solid': '#0a0b0e',
  '--surface-soft': '#0e1015',
  '--surface-card': '#12141a',
  '--surface-elevated': '#181b22',
  '--hairline': '#26262a',
  '--hairline-strong': '#3a3a40',
  '--hairline-faint': '#141416',
  '--text-primary': '#f2f2f4',
  '--text-body-strong': '#d6d8dd',
  '--text-body': '#a9aeb6',
  '--text-muted': '#8b9099',
  '--text-faint': '#7e838c',
  '--t-live': '#5ad3e3',
  '--t-threshold': '#f4b400',
  '--t-danger': '#e22718',
  '--t-danger-text': '#e84a3a',
  '--t-ok': '#1fa348',
  '--t-bbr': '#b48ae0',
  '--t-neutral': '#9aa3ad',
  '--brand-1': '#0066b1',
  '--brand-2': '#1c69d4',
  '--brand-3': '#e22718',
  '--chart-cake': '#1fd4a0',
  '--chart-fq-codel': '#5ad3e3',
  '--chart-pfifo': '#9aa3ad',
  '--chart-cubic': '#e08a4c',
  '--chart-bbr': '#4c9be8',
  '--chart-bbr3': '#b48ae0',
  '--chart-profile-1': '#5ad3e3',
  '--chart-profile-2': '#1fa348',
  '--chart-profile-3': '#f4b400',
  '--chart-profile-4': '#b48ce8',
  '--chart-profile-5': '#e2635c',
  '--chart-profile-6': '#8b9099',
  '--chart-on-ok': '#04120a',
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
