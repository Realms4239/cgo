import type { EChartsOption } from 'echarts'
import { paletteOnce, alpha } from './palette'

// Grammaire de graphiques en ~100 lignes, pas un framework.
// One parameterized base + craft factories (line | bar | area | scatter) —
// toute surface ECharts de l'app passe par cette grammaire.
// Les couleurs viennent de palette.ts (jetons CSS résolus) : aucun hex ici,
// sauf LIGHT qui recopie littéralement les jetons clairs ratifiés (slice 9).
// Le canvas n'accepte pas var() : tailles de trait littérales, nommées ici pour
// que les composants .tsx n'aient pas de chiffre nu (porte designFloor).
export const CANVAS_FS = { axis: 10, tooltip: 11 } as const

export const LIGHT = {
  canvas: '#fbfbfb',
  surface: '#ffffff',
  text: '#333333',
  textMuted: '#5a5a5a',
  grid: '#e8e8e8',
  live: '#0f6b62', // brand-teal-ink
  ok: '#1e7a3c',
  warn: '#9a7400', // warn-mark — 4,3:1 sur canevas clair
  danger: '#d51600',
  neutral: '#34495e',
} as const

// Ligne d'objectif RÉELLE : yAxis porte la valeur, l'étiquette la nomme.
// Aucune ligne décorative, aucun objectif implicite.
export function objectiveMarkLine(ms: number) {
  const goal = Number.isFinite(ms) && ms > 0 ? ms : 1
  return {
    silent: true,
    symbol: ['none', 'none'],
    lineStyle: { color: LIGHT.warn, type: 'dashed' as const, width: 1.5 },
    label: {
      show: true,
      position: 'insideEndTop' as const,
      formatter: `objectif ${Math.round(goal)} ms`,
      color: LIGHT.warn,
      fontFamily: 'JetBrains Mono',
      fontSize: 10,
    },
    data: [{ yAxis: goal }],
  }
}

export function baseOption(title: string, unit: string, opts?: { idle?: boolean; light?: boolean }): EChartsOption {
  const p = paletteOnce()
  const light = !!opts?.light
  const c = light
    ? { text: LIGHT.textMuted, title: LIGHT.text, subtle: LIGHT.textMuted, grid: LIGHT.grid, tick: LIGHT.neutral, faint: LIGHT.grid, surface: LIGHT.surface, live: LIGHT.live }
    : { text: p.textMuted, title: p.textPrimary, subtle: p.textFaint, grid: p.hairline, tick: p.hairlineStrong, faint: p.hairlineFaint, surface: p.surfaceElevated, live: p.live }
  return {
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'JetBrains Mono', fontSize: 10, color: c.text },
    animation: false,
    animationDuration: 0,
    animationDurationUpdate: 0,
    // Le canvas n'accepte pas var() : nom de face littéral. La face d'interface
    // est unique (slice 20B) et la serif de titre est sortie de l'application.
    title: { text: title, left: 8, top: 4, textStyle: { color: c.title, fontFamily: 'Inter', fontSize: 13, fontWeight: 600 } },
    tooltip: {
      trigger: 'axis',
      backgroundColor: alpha(c.surface, 0.92),
      borderColor: alpha(c.title, 0.08),
      borderWidth: 1,
      padding: [16, 20] as unknown as number,
      extraCssText: 'backdrop-filter: blur(16px); box-shadow: 0 12px 40px rgba(51,51,51,0.18), 0 0 0 1px rgba(51,51,51,0.06) inset; border-radius: 0;',
      textStyle: { fontFamily: 'JetBrains Mono', fontSize: 11, color: c.title, lineHeight: 18 },
      axisPointer: { type: 'cross' },
    } as unknown as EChartsOption['tooltip'],
    grid: { left: 64, right: 32, top: 48, bottom: 40, containLabel: true, backgroundColor: alpha(c.title, 0.005), borderWidth: 1, borderColor: alpha(c.title, 0.02) },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: c.grid } },
      axisTick: { show: true, length: 4, lineStyle: { color: c.tick } },
      minorTick: { show: true, splitNumber: 4 },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: c.faint, cap: 'round' as const } },
      // x-labels tiennent à 390 : recouvrement masqué, 9 px, heures courtes
      axisLabel: { color: c.text, fontSize: 9, fontFamily: 'JetBrains Mono', margin: 12, hideOverlap: true, formatter: (v: number) => { const d = new Date(v); return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}` } },
      axisPointer: { type: 'shadow', shadowStyle: { color: alpha(c.live, 0.04), shadowBlur: 12 } },
    } as unknown as EChartsOption['xAxis'],
    yAxis: {
      type: 'value',
      name: unit,
      nameTextStyle: { color: c.subtle, fontSize: 10, fontFamily: 'JetBrains Mono' },
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: c.grid } },
      axisLabel: { color: c.text, fontSize: 10, fontFamily: 'JetBrains Mono' },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: c.faint, cap: 'round' as const } },
    } as unknown as EChartsOption['yAxis'],
    dataZoom: [
      { type: 'inside', filterMode: 'none', zoomOnMouseWheel: true, moveOnMouseMove: true, preventDefaultMouseMove: true },
    ] as unknown as EChartsOption['dataZoom'],
    // semantic overload visualMap piecewise — STRIPPED per §3 clean triple
    // (its gradient pipeline crashed LineView/getVisualGradient mid-run)
    // filigrane seulement hors idle — il éclat dans le vide sur un graphe vide
    graphic: opts?.idle ? [] : [
      { type: 'text', left: 'center', top: 10, style: { text: 'METEOLINK \u00B7 LIEN', fill: alpha(c.title, 0.03), font: '600 28px Inter', textAlign: 'center' }, silent: true },
      { type: 'image', left: 'center', top: 'center', style: { image: 'data:image/svg+xml;base64,PHN2Zz4=', width: 300, height: 300, opacity: 0.015 }, silent: true },
    ] as unknown as EChartsOption['graphic'],
  } as unknown as EChartsOption
}

// palette — source unique des couleurs (pas de hex dans les vues).
// Getters : la valeur est résolue au premier rendu (paletteOnce), jamais au
// chargement du module — le canvas reçoit un hex, pas un var().
export const CRAFT = {
  get live() { return paletteOnce().live },
  get ok() { return paletteOnce().ok },
  get threshold() { return paletteOnce().threshold },
  get bbr() { return paletteOnce().bbr },
  get steel() { return paletteOnce().neutral },
  get danger() { return paletteOnce().danger },
}

// fabriques — markPoint en opt-in, formateur à l'unité (pas d'encombrement)
export function lineSeries(name: string, data: [number, number][], color: string, area = false, maxMarkUnit?: string) {
  return {
    name,
    type: 'line',
    showSymbol: false,
    smooth: 0.4,
    smoothMonotone: 'x',
    sampling: 'lttb' as const,
    lineStyle: { width: 2, cap: 'round' as const, join: 'round' as const, shadowBlur: 12, shadowColor: color + '66', shadowOffsetY: 2, color },
    // aire translucide plate — le LinearGradient décoratif plantait ECharts
    // (getVisualGradient 'coord') — interdit par la grammaire propre
    areaStyle: area ? { color: color + '26', opacity: 0.8 } : undefined,
    emphasis: { focus: 'series', lineStyle: { width: 3 }, itemStyle: { borderWidth: 2 } },
    blur: { lineStyle: { opacity: 0.2 } },
    ...(maxMarkUnit ? { markPoint: { data: [{ type: 'max', label: { formatter: `max {c} ${maxMarkUnit}` } }] } } : {}),
    data,
  }
}

// rawLineSeries — série brute : aucun lissage, aucun LTTB, les trous coupent.
// Le débit d'interface est une mesure échantillonnée ; une courbe lissée
// inventerait des points entre deux fenêtres réelles.
export function rawLineSeries(name: string, data: [number, number | null][], color: string) {
  return {
    name,
    type: 'line' as const,
    showSymbol: false,
    smooth: false,
    connectNulls: false,
    lineStyle: { width: 2, cap: 'round' as const, join: 'round' as const, color },
    emphasis: { focus: 'series' as const, lineStyle: { width: 3 } },
    blur: { lineStyle: { opacity: 0.2 } },
    data,
  }
}

// bar craft — pareto/grouped bars (Résultats, A/B, any future panel multiplicity)
export function barSeries(name: string, data: [string | number, number][], color: string, unit = '') {
  return {
    name,
    type: 'bar',
    barMaxWidth: 26,
    itemStyle: { color, borderColor: paletteOnce().hairline, borderWidth: 1, borderRadius: [2, 2, 0, 0] as [number, number, number, number] },
    emphasis: { focus: 'series', itemStyle: { color, opacity: 0.85 } },
    blur: { itemStyle: { opacity: 0.2 } },
    ...(unit ? { markPoint: { data: [{ type: 'max', label: { formatter: `max {c} ${unit}` } }] } } : {}),
    data,
  }
}

// scatter craft — compromise frontier (goodput vs latency), best flagged
export function scatterSeries(name: string, data: [number, number][], color: string, bestIndexes: number[] = []) {
  return {
    name,
    type: 'scatter',
    symbol: 'circle' as const,
    symbolSize: (_val: unknown, params: { dataIndex: number }) => (bestIndexes.includes(params.dataIndex) ? 12 : 8),
    itemStyle: { color, borderColor: paletteOnce().textPrimary, borderWidth: 1, shadowBlur: 8, shadowColor: color + '55' },
    emphasis: { itemStyle: { color: paletteOnce().threshold } },
    blur: { itemStyle: { opacity: 0.2 } },
    data,
  }
}

// paretoFrontier — enveloppe réelle : indices des points non dominés
// (aucun autre point meilleur-ou-égal sur X ET Y avec un strict).
// dir 'down' = plus petit = mieux (latences, coût), 'up' = plus grand =
// mieux (goodput, échéances, indice). Fonction pure testée — le tracé
// TradeSpace n'affiche que ce qu'elle rend, jamais de ligne décorative.
export function paretoFrontier(
  pts: { x: number; y: number }[],
  dirX: 'down' | 'up',
  dirY: 'down' | 'up',
): number[] {
  const betterX = (a: number, b: number) => (dirX === 'down' ? a < b : a > b)
  const betterY = (a: number, b: number) => (dirY === 'down' ? a < b : a > b)
  const out: number[] = []
  for (let i = 0; i < pts.length; i++) {
    let dominated = false
    for (let j = 0; j < pts.length; j++) {
      if (i === j) continue
      const xOk = pts[j].x === pts[i].x || betterX(pts[j].x, pts[i].x)
      const yOk = pts[j].y === pts[i].y || betterY(pts[j].y, pts[i].y)
      if (xOk && yOk && (betterX(pts[j].x, pts[i].x) || betterY(pts[j].y, pts[i].y))) {
        dominated = true
        break
      }
    }
    if (!dominated) out.push(i)
  }
  return out
}

// markArea CHARGE — une seule par graphique, pilotée par la phase. Fenêtres
// (ce<=cs, empty rings at phase boundary) render empty — ECharts markArea
// sinon 'coord' indéfini, ce qui tuait toute la boucle de rendu.
export function chargeMarkArea(cs: number, ce: number, show: boolean) {
  const ok = show && Number.isFinite(cs) && Number.isFinite(ce) && ce > cs
  const amber = paletteOnce().threshold
  return {
    itemStyle: { color: alpha(amber, 0.04), borderColor: alpha(amber, 0.12), borderWidth: 1, borderType: 'dashed' as const },
    label: { show: true, color: amber, fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop' as const, padding: [4, 8] as unknown as number[], backgroundColor: alpha(amber, 0.08), formatter: 'CHARGE' },
    data: ok ? [[{ xAxis: cs }, { xAxis: ce }]] as any : [],
  }
}
