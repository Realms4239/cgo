// Entrée ECharts élaguée — importer ici, jamais depuis 'echarts' racine.
import * as echarts from 'echarts/core'
import { LineChart, ScatterChart } from 'echarts/charts'
import { BrushComponent, DataZoomComponent, GraphicComponent, GridComponent, MarkAreaComponent, MarkPointComponent, ToolboxComponent, TooltipComponent, VisualMapComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

// Brush + Scatter pour la comparaison; Toolbox si besoin.
echarts.use([LineChart, ScatterChart, GridComponent, TooltipComponent, MarkAreaComponent, MarkPointComponent, DataZoomComponent, VisualMapComponent, GraphicComponent, BrushComponent, ToolboxComponent, CanvasRenderer])

export { echarts }
