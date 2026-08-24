// Tree-shaken ECharts entry — import from here, never from 'echarts' root.
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { DataZoomComponent, GraphicComponent, GridComponent, MarkAreaComponent, MarkPointComponent, TooltipComponent, VisualMapComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, GridComponent, TooltipComponent, MarkAreaComponent, MarkPointComponent, DataZoomComponent, VisualMapComponent, GraphicComponent, CanvasRenderer])

export { echarts }
