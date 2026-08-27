// Tree-shaken ECharts entry — import from here, never from 'echarts' root.
import * as echarts from 'echarts/core'
import { LineChart, ScatterChart } from 'echarts/charts'
import { BrushComponent, DataZoomComponent, GraphicComponent, GridComponent, MarkAreaComponent, MarkPointComponent, ToolboxComponent, TooltipComponent, VisualMapComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

// ponytail: Brush+Scatter added for Task 6 scatter compare — Toolbox rect if cheap, full brush toolbox if UX needs
echarts.use([LineChart, ScatterChart, GridComponent, TooltipComponent, MarkAreaComponent, MarkPointComponent, DataZoomComponent, VisualMapComponent, GraphicComponent, BrushComponent, ToolboxComponent, CanvasRenderer])

export { echarts }
