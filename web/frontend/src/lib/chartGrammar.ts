import type { EChartsOption } from 'echarts'
import { paletteOnce, alpha } from './palette'

// Grammaire de graphiques en ~100 lignes, pas un framework.
// One parameterized base + craft factories (line | bar | area | scatter) —
// toute surface ECharts de l'app passe par cette grammaire.
// Les couleurs viennent de palette.ts (jetons CSS résolus) : aucun hex ici.
export function baseOption(title: string, unit: string, opts?: { idle?: boolean }): EChartsOption {
  const p = paletteOnce()
  return {
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'JetBrains Mono', fontSize: 10, color: p.textMuted },
    animation: false,
    animationDuration: 0,
    animationDurationUpdate: 0,
    title: { text: title, left: 8, top: 4, textStyle: { color: p.textPrimary, fontFamily: 'Cormorant Garamond', fontSize: 13, fontWeight: 600 } },
    tooltip: {
      trigger: 'axis',
      backgroundColor: alpha(p.surfaceElevated, 0.92),
      borderColor: alpha(p.textPrimary, 0.08),
      borderWidth: 1,
      padding: [16, 20] as unknown as number,
      extraCssText: 'backdrop-filter: blur(16px); box-shadow: 0 12px 40px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.06) inset; border-radius: 0;',
      textStyle: { fontFamily: 'JetBrains Mono', fontSize: 11, color: p.textPrimary, lineHeight: 18 },
      axisPointer: { type: 'cross' },
    } as unknown as EChartsOption['tooltip'],
    grid: { left: 64, right: 32, top: 48, bottom: 40, containLabel: true, backgroundColor: alpha(p.textPrimary, 0.005), borderWidth: 1, borderColor: alpha(p.textPrimary, 0.02) },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: p.hairline } },
      axisTick: { show: true, length: 4, lineStyle: { color: p.hairlineStrong } },
      minorTick: { show: true, splitNumber: 4 },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: p.hairlineFaint, cap: 'round' as const } },
      // x-labels tiennent à 390 : recouvrement masqué, 9 px, heures courtes
      axisLabel: { color: p.textMuted, fontSize: 9, fontFamily: 'JetBrains Mono', margin: 12, hideOverlap: true, formatter: (v: number) => { const d = new Date(v); return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}` } },
      axisPointer: { type: 'shadow', shadowStyle: { color: alpha(p.live, 0.04), shadowBlur: 12 } },
    } as unknown as EChartsOption['xAxis'],
    yAxis: {
      type: 'value',
      name: unit,
      nameTextStyle: { color: p.textFaint, fontSize: 10, fontFamily: 'JetBrains Mono' },
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: p.hairline } },
      axisLabel: { color: p.textMuted, fontSize: 10, fontFamily: 'JetBrains Mono' },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: p.hairlineFaint, cap: 'round' as const } },
    } as unknown as EChartsOption['yAxis'],
    dataZoom: [
      { type: 'inside', filterMode: 'none', zoomOnMouseWheel: true, moveOnMouseMove: true, preventDefaultMouseMove: true },
    ] as unknown as EChartsOption['dataZoom'],
    // semantic overload visualMap piecewise — STRIPPED per §3 clean triple
    // (its gradient pipeline crashed LineView/getVisualGradient mid-run)
    // filigrane seulement hors idle — il éclat dans le vide sur un graphe vide
    graphic: opts?.idle ? [] : [
      { type: 'text', left: 'center', top: 10, style: { text: 'METEOLINK \u00B7 LIEN', fill: alpha(p.textPrimary, 0.03), font: '600 28px Cormorant Garamond', textAlign: 'center' }, silent: true },
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
