import type { EChartsOption } from 'echarts'

// Grammaire de graphiques en ~100 lignes, pas un framework.
// One parameterized base + craft factories (line | bar | area | scatter) —
// toute surface ECharts de l'app passe par cette grammaire.
export function baseOption(title: string, unit: string, opts?: { idle?: boolean }): EChartsOption {
  return {
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099' },
    animation: false,
    animationDuration: 0,
    animationDurationUpdate: 0,
    title: { text: title, left: 8, top: 4, textStyle: { color: '#f2f2f4', fontFamily: 'Cormorant Garamond', fontSize: 13, fontWeight: 600 } },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(16,16,18,0.92)',
      borderColor: 'rgba(255,255,255,0.08)',
      borderWidth: 1,
      padding: [16, 20] as unknown as number,
      extraCssText: 'backdrop-filter: blur(16px); box-shadow: 0 12px 40px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.06) inset; border-radius: 0;',
      textStyle: { fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4', lineHeight: 18 },
      axisPointer: { type: 'cross' },
    } as unknown as EChartsOption['tooltip'],
    grid: { left: 64, right: 32, top: 48, bottom: 40, containLabel: true, backgroundColor: 'rgba(255,255,255,0.005)', borderWidth: 1, borderColor: 'rgba(255,255,255,0.02)' },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: '#2a2a30' } },
      axisTick: { show: true, length: 4, lineStyle: { color: '#3a3a40' } },
      minorTick: { show: true, splitNumber: 4 },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: '#1a1a1e', cap: 'round' as const } },
      axisLabel: { color: '#8b9099', fontSize: 10, fontFamily: 'JetBrains Mono', margin: 12 },
      axisPointer: { type: 'shadow', shadowStyle: { color: 'rgba(90,211,227,0.04)', shadowBlur: 12 } },
    } as unknown as EChartsOption['xAxis'],
    yAxis: {
      type: 'value',
      name: unit,
      nameTextStyle: { color: '#767b84', fontSize: 10, fontFamily: 'JetBrains Mono' },
      axisLine: { lineStyle: { width: 1.5, cap: 'round' as const, color: '#2a2a30' } },
      axisLabel: { color: '#8b9099', fontSize: 10, fontFamily: 'JetBrains Mono' },
      splitLine: { lineStyle: { type: [4, 4] as unknown as string, color: '#1a1a1e', cap: 'round' as const } },
    } as unknown as EChartsOption['yAxis'],
    dataZoom: [
      { type: 'inside', filterMode: 'none', zoomOnMouseWheel: true, moveOnMouseMove: true, preventDefaultMouseMove: true },
    ] as unknown as EChartsOption['dataZoom'],
    // semantic overload visualMap piecewise — STRIPPED per §3 clean triple
    // (its gradient pipeline crashed LineView/getVisualGradient mid-run)
    // filigrane seulement hors idle — il éclat dans le vide sur un graphe vide
    graphic: opts?.idle ? [] : [
      { type: 'text', left: 'center', top: 10, style: { text: 'METEOLINK \u00B7 LIEN', fill: 'rgba(255,255,255,0.03)', font: '600 28px Cormorant Garamond', textAlign: 'center' }, silent: true },
      { type: 'image', left: 'center', top: 'center', style: { image: 'data:image/svg+xml;base64,PHN2Zz4=', width: 300, height: 300, opacity: 0.015 }, silent: true },
    ] as unknown as EChartsOption['graphic'],
  } as unknown as EChartsOption
}

// palette — source unique des couleurs (pas de hex dans les vues)
export const CRAFT = {
  live: '#5ad3e3',
  ok: '#1fa348',
  threshold: '#f4b400',
  bbr: '#b48ae0',
  steel: '#9aa3ad',
  danger: '#e22718',
} as const

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
    itemStyle: { color, borderColor: '#26262a', borderWidth: 1, borderRadius: [2, 2, 0, 0] as [number, number, number, number] },
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
    itemStyle: { color, borderColor: '#f2f2f4', borderWidth: 1, shadowBlur: 8, shadowColor: color + '55' },
    emphasis: { itemStyle: { color: '#f4b400' } },
    blur: { itemStyle: { opacity: 0.2 } },
    data,
  }
}

// markArea CHARGE — une seule par graphique, pilotée par la phase. Fenêtres
// (ce<=cs, empty rings at phase boundary) render empty — ECharts markArea
// sinon 'coord' indéfini, ce qui tuait toute la boucle de rendu.
export function chargeMarkArea(cs: number, ce: number, show: boolean) {
  const ok = show && Number.isFinite(cs) && Number.isFinite(ce) && ce > cs
  return {
    itemStyle: { color: 'rgba(244,180,0,0.04)', borderColor: 'rgba(244,180,0,0.12)', borderWidth: 1, borderType: 'dashed' as const },
    label: { show: true, color: '#f4b400', fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop' as const, padding: [4, 8] as unknown as number[], backgroundColor: 'rgba(244,180,0,0.08)', formatter: 'CHARGE' },
    data: ok ? [[{ xAxis: cs }, { xAxis: ce }]] as any : [],
  }
}
