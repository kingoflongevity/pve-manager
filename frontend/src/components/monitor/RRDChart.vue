<template>
  <div class="rrd-chart-wrapper">
    <div class="rrd-chart" ref="chartRef" :style="{ height: height }"></div>
    
    <!-- 暂停/继续控制按钮 -->
    <div v-if="props.refreshInterval > 0" class="chart-controls">
      <el-button
        size="small"
        :type="autoRefreshPaused ? 'primary' : ''"
        @click="toggleAutoRefresh"
        class="pause-btn"
      >
        <el-icon>
          <VideoPlay v-if="autoRefreshPaused" />
          <VideoPause v-else />
        </el-icon>
        {{ autoRefreshPaused ? '继续' : '暂停' }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * RRDChart - 可复用的 RRD 监控图表组件
 * 
 * 支持多种图表类型（line、bar、area），自动从 PVE API 获取 RRD 数据。
 * 支持自动刷新、数据缩放、缺失数据处理等特性。
 * 提供暂停/继续控制按钮以符合无障碍要求（减少闪烁）。
 */
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { VideoPlay, VideoPause } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import { getNodeRRD } from '@/api/node'
import { getQEMURRD } from '@/api/qemu'
import { getLXCRRD } from '@/api/lxc'
import type { RRDDataPoint, RRDTimeframe, RRDDataSet } from '@/api/types'

interface ChartSeries {
  name: string
  dataKey: string
  color: string
  unit?: string
}

const props = withDefaults(defineProps<{
  /** 节点名称 */
  node: string
  /** 目标资源标识（QEMU VMID 或 LXC CTID，节点级监控可省略） */
  target?: number
  /** 资源类型：node | qemu | lxc */
  resourceType?: 'node' | 'qemu' | 'lxc'
  /** 时间范围 */
  timeframe?: RRDTimeframe
  /** 数据集 */
  dataset?: RRDDataSet
  /** 图表类型 */
  chartType?: 'line' | 'bar' | 'area'
  /** 图表高度 */
  height?: string
  /** 图表标题 */
  title?: string
  /** Y 轴单位 */
  unit?: string
  /** Y 轴最大值 */
  yAxisMax?: number
  /** 数据系列配置（覆盖默认） */
  series?: ChartSeries[]
  /** 自动刷新间隔（秒），0 表示不自动刷新 */
  refreshInterval?: number
}>(), {
  resourceType: 'node',
  timeframe: 'hour',
  dataset: 'cpu',
  chartType: 'line',
  height: '300px',
  title: '',
  unit: '',
  yAxisMax: undefined,
  series: () => [],
  refreshInterval: 300,
})

const chartRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null
let refreshTimer: ReturnType<typeof setInterval> | null = null
const loading = ref(false)
/** 自动刷新暂停状态 */
const autoRefreshPaused = ref(false)

/** 获取默认的数据系列配置 */
function getDefaultSeries(): ChartSeries[] {
  const root = document.documentElement
  const isLight = root.classList.contains('theme-light')
  
  const defaults: Record<RRDDataSet, ChartSeries[]> = {
    cpu: [
      { name: 'CPU 使用率', dataKey: 'cpu', color: isLight ? '#3B82F6' : '#409EFF', unit: '%' },
    ],
    memory: [
      { name: '内存使用', dataKey: 'mem', color: isLight ? '#16A34A' : '#67C23A', unit: 'MB' },
      { name: '交换空间', dataKey: 'swap', color: isLight ? '#F97316' : '#E6A23C', unit: 'MB' },
    ],
    network: [
      { name: '接收', dataKey: 'netin', color: isLight ? '#3B82F6' : '#409EFF', unit: 'B/s' },
      { name: '发送', dataKey: 'netout', color: isLight ? '#16A34A' : '#67C23A', unit: 'B/s' },
    ],
    disk: [
      { name: '读取', dataKey: 'diskread', color: isLight ? '#3B82F6' : '#409EFF', unit: 'B/s' },
      { name: '写入', dataKey: 'diskwrite', color: isLight ? '#F97316' : '#E6A23C', unit: 'B/s' },
    ],
    system: [
      { name: '系统负载', dataKey: 'loadavg', color: isLight ? '#4B5563' : '#909399' },
    ],
  }
  return defaults[props.dataset] || defaults.cpu
}

/** 格式化字节数为人类可读格式 */
function formatBytes(value: number): string {
  if (value === 0 || value === undefined || isNaN(value)) return '0'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  const i = Math.floor(Math.log(Math.abs(value)) / Math.log(k))
  return (value / k ** i).toFixed(2) + ' ' + units[i]
}

/** 格式化时间戳为可读时间 */
function formatTime(timestamp: number): string {
  const date = new Date(timestamp * 1000)
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

/** 获取完整时间格式（用于 tooltip） */
function formatFullTime(timestamp: number): string {
  const date = new Date(timestamp * 1000)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${month}-${day} ${hours}:${minutes}`
}

/** 获取 RRD 数据 */
async function fetchRRDData(): Promise<RRDDataPoint[]> {
  switch (props.resourceType) {
    case 'qemu':
      if (!props.target) throw new Error('QEMU 资源需要提供 target (vmid)')
      return getQEMURRD(props.node, props.target, props.timeframe, props.dataset)
    case 'lxc':
      if (!props.target) throw new Error('LXC 资源需要提供 target (ctid)')
      return getLXCRRD(props.node, props.target, props.timeframe, props.dataset)
    default:
      return getNodeRRD(props.node, props.timeframe, props.dataset)
  }
}

/** 加载并渲染图表数据 */
async function loadChartData() {
  loading.value = true
  try {
    const rawData = await fetchRRDData()
    renderChart(rawData)
  } catch (error) {
    console.error('获取 RRD 数据失败:', error)
    // 显示空状态
    renderEmptyChart()
  } finally {
    loading.value = false
  }
}

/** 获取主题颜色 */
function getThemeColors() {
  const root = document.documentElement
  const isLight = root.classList.contains('theme-light')
  return {
    isLight,
    axisColor: isLight ? '#E5E7EB' : '#dcdfe6',
    splitLine: isLight ? '#F3F4F6' : '#ebeef5',
    emptyText: isLight ? '#6B7280' : '#909399',
  }
}

/** 将 RRD 数据渲染到 ECharts */
function renderChart(data: RRDDataPoint[]) {
  if (!chartInstance) return

  const seriesConfig = props.series.length > 0 ? props.series : getDefaultSeries()
  const theme = getThemeColors()
  
  // 提取时间轴数据
  const times = data.map(d => d.time).filter(t => t !== undefined)
  
  // 构建每个系列的数据，处理缺失值
  const series = seriesConfig.map(cfg => {
    const chartData = data.map(point => {
      const val = point[cfg.dataKey]
      // 缺失数据返回 null，ECharts 会自动断开连接或插值
      return val !== undefined && val !== null ? Number(val) : null
    })
    
    return {
      name: cfg.name,
      type: props.chartType === 'area' ? 'line' : props.chartType,
      data: chartData,
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2 },
      itemStyle: { color: cfg.color },
      areaStyle: props.chartType === 'area'
        ? { color: cfg.color, opacity: 0.3 }
        : undefined,
    }
  })

  const option: EChartsOption = {
    title: props.title ? { text: props.title, left: 'center' } : undefined,
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'line' },
      formatter: (params: unknown) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        const ts = (params as any[])[0].axisValue
        let result = formatFullTime(Number(ts)) + '<br/>'
        for (const p of params as any[]) {
          const val = p.value
          const seriesCfg = seriesConfig.find(s => s.name === p.seriesName)
          const unit = seriesCfg?.unit || props.unit
          const displayVal = val === null ? '--' : 
            (unit === 'B/s' || unit === 'B') ? formatBytes(val) : 
            (unit === '%' ? val.toFixed(1) + '%' : val.toFixed(2) + ' ' + unit)
          result += `${p.marker}${p.seriesName}: ${displayVal}<br/>`
        }
        return result
      },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '10%',
      top: props.title ? '15%' : '8%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLabel: {
        formatter: (value: number) => formatTime(value),
        interval: 'auto',
        rotate: 0,
      },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: theme.axisColor } },
    },
    yAxis: {
      type: 'value',
      max: props.yAxisMax,
      axisLabel: {
        formatter: (value: number) => {
          if (props.unit === 'B/s' || props.unit === 'B') return formatBytes(value)
          if (props.unit === '%') return value + '%'
          if (props.unit === 'MB') return value + ' MB'
          return value
        },
      },
      splitLine: { lineStyle: { type: 'dashed', color: theme.splitLine } },
    },
    series,
    dataZoom: [
      {
        type: 'slider',
        bottom: 10,
        height: 20,
        handleSize: 0,
        showDetail: false,
      },
      {
        type: 'inside',
        xAxisIndex: 0,
        zoomOnMouseWheel: false,
        moveOnMouseMove: true,
      },
    ],
    animation: false,
  }

  chartInstance.setOption(option, true)
}

/** 渲染空状态图表 */
function renderEmptyChart() {
  if (!chartInstance) return
  const theme = getThemeColors()
  chartInstance.setOption({
    title: { text: '暂无数据', left: 'center', top: 'middle', textStyle: { color: theme.emptyText, fontSize: 14 } },
    xAxis: { show: false },
    yAxis: { show: false },
    series: [],
    grid: {},
  }, true)
}

/** 启动自动刷新 */
function startAutoRefresh() {
  stopAutoRefresh()
  if (props.refreshInterval > 0 && !autoRefreshPaused.value) {
    refreshTimer = setInterval(() => {
      loadChartData()
    }, props.refreshInterval * 1000)
  }
}

/** 停止自动刷新 */
function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

/** 切换自动刷新暂停/继续状态 */
function toggleAutoRefresh() {
  autoRefreshPaused.value = !autoRefreshPaused.value
  if (!autoRefreshPaused.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

/** 初始化图表实例 */
onMounted(async () => {
  await nextTick()
  if (chartRef.value) {
    chartInstance = echarts.init(chartRef.value)
    window.addEventListener('resize', handleResize)
    loadChartData()
    startAutoRefresh()
  }
})

/** 监听窗口大小变化 */
function handleResize() {
  chartInstance?.resize()
}

/** 监听关键属性变化时重新加载数据 */
watch(
  () => [props.node, props.target, props.timeframe, props.dataset, props.chartType, props.resourceType] as const,
  () => {
    loadChartData()
    startAutoRefresh()
  },
)

/** 组件卸载时清理资源 */
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  stopAutoRefresh()
  chartInstance?.dispose()
  chartInstance = null
})

/** 暴露刷新方法供外部调用 */
defineExpose({ refresh: loadChartData })
</script>

<style scoped lang="scss">
.rrd-chart-wrapper {
  width: 100%;
  position: relative;
  
  .rrd-chart {
    width: 100%;
    background: var(--color-bg-container, #0F172A);
    border-radius: 12px;
    border: 1px solid var(--color-border-base, rgba(30, 41, 59, 0.8));
    padding: 16px;
  }

  .chart-controls {
    position: absolute;
    top: 16px;
    right: 16px;
    display: flex;
    gap: 8px;
    z-index: 3000;

    .pause-btn {
      display: flex;
      align-items: center;
      gap: 8px;
      background: var(--color-bg-elevated, #1E293B);
      border: 1px solid var(--color-border-base, rgba(30, 41, 59, 0.8));
      color: var(--color-text-secondary, #94A3B8);
      transition: all 0.2s ease;
      
      &:hover {
        color: var(--color-primary, #4ADE80);
        border-color: var(--color-border-focus, #22C55E);
      }

      &.el-button--primary {
        background: var(--color-primary-gradient, linear-gradient(135deg, #22C55E, #16A34A));
        border: none;
        color: var(--color-text-on-primary, #020617);

        &:hover {
          background: var(--color-primary-gradient-hover, linear-gradient(135deg, #4ADE80, #22C55E));
        }
      }
    }
  }
}
</style>
