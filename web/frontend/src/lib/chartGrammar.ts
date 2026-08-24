import type { EChartsOption } from 'echarts'
import { echarts } from './echarts'

// ponytail: observatory grammar in ~60 lines, not a chart framework.
export function baseOption(title: string, unit: string): EChartsOption {
  return {
    useDirtyRect: true,
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099' },
    animationDuration: 900,
    animationDurationUpdate: 600,
    animationEasing: 'cubicInOut' as unknown as string,
    animationDelay: (idx: number) => idx * 15,
    animationThreshold: 2000,
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
      {
        type: 'slider',
        height: 24,
        handleIcon: 'path://M-5,0 L5,0 L5,20 L-5,20 Z',
        handleSize: '100%',
        handleStyle: { color: '#f2f2f4', borderColor: '#26262a' },
        backgroundColor: '#0b0b0c',
        fillerColor: 'rgba(90,211,227,0.12)',
        borderColor: 'transparent',
        selectedDataBackground: { lineStyle: { color: '#5ad3e3' }, areaStyle: { color: 'rgba(90,211,227,0.15)' } },
        emphasis: { handleStyle: { borderColor: '#5ad3e3' } },
      },
    ] as unknown as EChartsOption['dataZoom'],
    visualMap: {
      show: false,
      type: 'piecewise',
      dimension: 1,
      pieces: [
        { gt: 100, color: '#e22718' },
        { gt: 40, color: '#f4b400' },
        { lte: 40, color: '#5ad3e3' },
      ],
      outOfRange: { color: '#9aa3ad' },
    } as unknown as EChartsOption['visualMap'],
    graphic: [
      { type: 'text', left: 'center', top: 12, style: { text: 'MadaLink \u00B7 LIEN \u2014 observatoire', fill: 'rgba(255,255,255,0.025)', font: '600 56px Cormorant Garamond', textAlign: 'center' }, silent: true },
      { type: 'image', left: 'center', top: 'center', style: { image: 'data:image/svg+xml;base64,PHN2Zz4=', width: 400, height: 400, opacity: 0.015 }, silent: true },
    ] as unknown as EChartsOption['graphic'],
    markArea: {
      itemStyle: { color: 'rgba(244,180,0,0.04)', borderColor: 'rgba(244,180,0,0.12)', borderWidth: 1, borderType: 'dashed' },
      label: { color: '#f4b400', fontFamily: 'JetBrains Mono', fontSize: 10, position: 'insideTop', padding: [4, 8] as unknown as number, backgroundColor: 'rgba(244,180,0,0.08)', borderRadius: 0 },
      data: [[{ xAxis: 0 }, { xAxis: 0 }]],
    } as unknown as EChartsOption['markArea'],
  } as unknown as EChartsOption
}

export function lineSeries(name: string, data: [number, number][], color: string, area = false) {
  return {
    name,
    type: 'line',
    showSymbol: false,
    smooth: 0.4,
    smoothMonotone: 'x',
    sampling: 'lttb' as const,
    lineStyle: { width: 2, cap: 'round' as const, join: 'round' as const, shadowBlur: 12, shadowColor: color + '66', shadowOffsetY: 2, color },
    areaStyle: area
      ? { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: color + '26' }, { offset: 1, color: color + '00' }]), opacity: 0.8 }
      : undefined,
    emphasis: { focus: 'series', lineStyle: { width: 3 }, itemStyle: { borderWidth: 2 } },
    blur: { lineStyle: { opacity: 0.2 } },
    markPoint: { data: [{ type: 'max', label: { formatter: 'max {c} ms' } }] },
    data,
  }
}
