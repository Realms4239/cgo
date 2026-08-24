import type { EChartsOption } from 'echarts'

// ponytail: observatory grammar in ~40 lines, not a chart framework.
export function baseOption(title: string, unit: string): EChartsOption {
  return {
    animation: false,
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'JetBrains Mono', fontSize: 10, color: '#8b9099' },
    title: { text: title, left: 8, top: 4, textStyle: { color: '#f2f2f4', fontFamily: 'Cormorant Garamond', fontSize: 13, fontWeight: 600 } },
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#161618',
      borderColor: '#26262a',
      textStyle: { fontFamily: 'JetBrains Mono', fontSize: 11, color: '#f2f2f4' },
      axisPointer: { type: 'cross' },
    },
    grid: { left: 48, right: 16, top: 32, bottom: 24, containLabel: false },
    xAxis: { type: 'time', axisLabel: { color: '#767b84', fontSize: 10 }, axisLine: { lineStyle: { color: '#26262a' } }, splitLine: { lineStyle: { color: '#141416' } } },
    yAxis: { type: 'value', name: unit, nameTextStyle: { color: '#767b84', fontSize: 10 }, axisLabel: { color: '#8b9099' }, splitLine: { lineStyle: { color: '#141416' } } },
  } as EChartsOption
}

export function lineSeries(name: string, data: [number, number][], color: string, area = false) {
  return {
    name, type: 'line', showSymbol: false, smooth: 0.3, lineStyle: { width: 1.8, color, shadowBlur: 8, shadowColor: color + '60' }, data,
    ...(area ? { areaStyle: { color: color + '18', opacity: 0.3 } } : {}),
    emphasis: { focus: 'series', lineStyle: { width: 2.5 } },
  }
}
