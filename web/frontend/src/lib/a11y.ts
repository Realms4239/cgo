// WCAG AA pour texte courant : seulement les couleurs calculées, jamais une
// intuition de nuance. L’alpha n’est pas composé ici : un texte transparent
// n’est jamais déclaré conforme.
type Rgb = { r: number; g: number; b: number }

function clamp01(value: number): number {
  return Math.min(255, Math.max(0, value)) / 255
}

function parseRgb(value: string): Rgb {
  const match = value.trim().match(/^(?:rgb\(\s*|rgba\(\s*)(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i)
  if (match) return { r: Number(match[1]), g: Number(match[2]), b: Number(match[3]) }
  const hex = value.trim().match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i)?.[1]
  if (hex) {
    const full = hex.length === 3 ? [...hex].map(c => c + c).join('') : hex
    return { r: Number.parseInt(full.slice(0, 2), 16), g: Number.parseInt(full.slice(2, 4), 16), b: Number.parseInt(full.slice(4, 6), 16) }
  }
  throw new Error(`couleur illisible : ${value}`)
}

function luminance({ r, g, b }: Rgb): number {
  const channel = (n: number) => {
    const c = clamp01(n)
    return c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
  }
  return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b)
}

export function contrastRatio(foreground: string, background: string): number {
  const a = luminance(parseRgb(foreground))
  const b = luminance(parseRgb(background))
  const light = Math.max(a, b)
  const dark = Math.min(a, b)
  return (light + 0.05) / (dark + 0.05)
}

export function wcagAaNormalText(ratio: number): boolean {
  return ratio >= 4.5
}
