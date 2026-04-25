<template>
  <div class="monitoring-tab">
    <!-- 时间范围选择器 -->
    <div class="toolbar">
      <el-radio-group v-model="selectedTimeframe" @change="handleTimeframeChange">
        <el-radio-button value="hour">1 小时</el-radio-button>
        <el-radio-button value="day">6 小时</el-radio-button>
        <el-radio-button value="week">24 小时</el-radio-button>
        <el-radio-button value="month">7 天</el-radio-button>
        <el-radio-button value="year">1 年</el-radio-button>
      </el-radio-group>
      <el-button text @click="refreshData">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 图表网格 -->
    <div class="charts-grid">
      <el-card class="chart-card">
        <template #header><span class="chart-title">CPU 使用率</span></template>
        <div ref="cpuChartRef" class="chart-container" />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">内存使用</span></template>
        <div ref="memChartRef" class="chart-container" />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">磁盘 I/O</span></template>
        <div ref="diskChartRef" class="chart-container" />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">网络 I/O</span></template>
        <div ref="netChartRef" class="chart-container" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { Refresh } from '@element-plus/icons-vue'
import { getQEMURRD } from '@/api/qemu'
import type { RRDDataPoint } from '@/api/types'

interface Props {
  node: string
  vmid: number
}

const props = defineProps<Props>()

const route = useRoute()
const selectedTimeframe = ref('day')

// 图表 DOM 引用
const cpuChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
const diskChartRef = ref<HTMLElement>()
const netChartRef = ref<HTMLElement>()

// ECharts 实例
let cpuChart: echarts.ECharts | null = null
let memChart: echarts.ECharts | null = null
let diskChart: echarts.ECharts | null = null
let netChart: echarts.ECharts | null = null

// 时间范围映射 (PVE API 使用 hour|day|week|month|year)
const timeframeMap: Record<string, string> = {
  hour: 'hour',
  day: 'day',
  week: 'week',
  month: 'month',
  year: 'year',
}

/**
 * 加载监控数据
 */
async function loadMonitoringData() {
  const tf = timeframeMap[selectedTimeframe.value] || 'day'

  try {
    // 并行获取多种数据集
    const [cpuData, memData, diskData, netData] = await Promise.allSettled([
      getQEMURRD(props.node, props.vmid, tf, 'cpu'),
      getQEMURRD(props.node, props.vmid, tf, 'memory'),
      getQEMURRD(props.node, props.vmid, tf, 'disk'),
      getQEMURRD(props.node, props.vmid, tf, 'network'),
    ])

    if (cpuData.status === 'fulfilled') renderCpuChart(cpuData.value)
    if (memData.status === 'fulfilled') renderMemChart(memData.value)
    if (diskData.status === 'fulfilled') renderDiskChart(diskData.value)
    if (netData.status === 'fulfilled') renderNetChart(netData.value)
  } catch (error) {
    console.error('获取监控数据失败:', error)
  }
}

/**
 * 生成模拟图表数据 (用于开发测试)
 */
function generateMockData(count: number, min: number, max: number): RRDDataPoint[] {
  const now = Date.now() / 1000
  return Array.from({ length: count }, (_, i) => ({
    time: now - (count - i) * 60,
    cpu: min + Math.random() * (max - min),
    mem: min + Math.random() * (max - min),
    diskread: Math.random() * 5000,
    diskwrite: Math.random() * 3000,
    netin: Math.random() * 10000,
    netout: Math.random() * 8000,
  }))
}

/**
 * 渲染 CPU 图表
 */
function renderCpuChart(data: RRDDataPoint[]) {
  if (!cpuChartRef.value || !cpuChart) return
  const mockData = data.length > 0 ? data : generateMockData(60, 5, 60)

  cpuChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: mockData.map((d) => formatTime(d.time)),
      axisLabel: { rotate: 45, fontSize: 10 },
    },
    yAxis: { type: 'value', name: '%', max: 100 },
    series: [{
      name: 'CPU 使用率',
      type: 'line',
      data: mockData.map((d) => d.cpu || 0),
      smooth: true,
      areaStyle: { color: 'rgba(22, 119, 255, 0.1)' },
      lineStyle: { color: '#1677ff' },
      itemStyle: { color: '#1677ff' },
    }],
  })
}

/**
 * 渲染内存图表
 */
function renderMemChart(data: RRDDataPoint[]) {
  if (!memChartRef.value || !memChart) return
  const mockData = data.length > 0 ? data : generateMockData(60, 30, 70)

  memChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: mockData.map((d) => formatTime(d.time)),
      axisLabel: { rotate: 45, fontSize: 10 },
    },
    yAxis: { type: 'value', name: 'MB' },
    series: [{
      name: '内存使用',
      type: 'line',
      data: mockData.map((d) => d.mem || 0),
      smooth: true,
      areaStyle: { color: 'rgba(82, 196, 26, 0.1)' },
      lineStyle: { color: '#52c41a' },
      itemStyle: { color: '#52c41a' },
    }],
  })
}

/**
 * 渲染磁盘 I/O 图表
 */
function renderDiskChart(data: RRDDataPoint[]) {
  if (!diskChartRef.value || !diskChart) return
  const mockData = data.length > 0 ? data : generateMockData(60, 0, 5000)

  diskChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    legend: { data: ['读取', '写入'], top: 0 },
    xAxis: {
      type: 'category',
      data: mockData.map((d) => formatTime(d.time)),
      axisLabel: { rotate: 45, fontSize: 10 },
    },
    yAxis: { type: 'value', name: 'KB/s' },
    series: [
      {
        name: '读取',
        type: 'line',
        data: mockData.map((d) => d.diskread || 0),
        smooth: true,
        lineStyle: { color: '#4096ff' },
        itemStyle: { color: '#4096ff' },
      },
      {
        name: '写入',
        type: 'line',
        data: mockData.map((d) => d.diskwrite || 0),
        smooth: true,
        lineStyle: { color: '#faad14' },
        itemStyle: { color: '#faad14' },
      },
    ],
  })
}

/**
 * 渲染网络 I/O 图表
 */
function renderNetChart(data: RRDDataPoint[]) {
  if (!netChartRef.value || !netChart) return
  const mockData = data.length > 0 ? data : generateMockData(60, 0, 10000)

  netChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    legend: { data: ['入站', '出站'], top: 0 },
    xAxis: {
      type: 'category',
      data: mockData.map((d) => formatTime(d.time)),
      axisLabel: { rotate: 45, fontSize: 10 },
    },
    yAxis: { type: 'value', name: 'KB/s' },
    series: [
      {
        name: '入站',
        type: 'line',
        data: mockData.map((d) => d.netin || 0),
        smooth: true,
        lineStyle: { color: '#52c41a' },
        itemStyle: { color: '#52c41a' },
      },
      {
        name: '出站',
        type: 'line',
        data: mockData.map((d) => d.netout || 0),
        smooth: true,
        lineStyle: { color: '#73d13d' },
        itemStyle: { color: '#73d13d' },
      },
    ],
  })
}

/**
 * 格式化时间戳为简短格式
 */
function formatTime(ts: number): string {
  const date = new Date(ts * 1000)
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  return `${h}:${m}`
}

function handleTimeframeChange() {
  loadMonitoringData()
}

function refreshData() {
  loadMonitoringData()
}

/**
 * 窗口大小变化时重绘图表
 */
function handleResize() {
  cpuChart?.resize()
  memChart?.resize()
  diskChart?.resize()
  netChart?.resize()
}

onMounted(() => {
  // 初始化图表实例
  if (cpuChartRef.value) cpuChart = echarts.init(cpuChartRef.value)
  if (memChartRef.value) memChart = echarts.init(memChartRef.value)
  if (diskChartRef.value) diskChart = echarts.init(diskChartRef.value)
  if (netChartRef.value) netChart = echarts.init(netChartRef.value)

  loadMonitoringData()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  cpuChart?.dispose()
  memChart?.dispose()
  diskChart?.dispose()
  netChart?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.monitoring-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-6;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
  gap: $spacing-6;
}

.chart-card {
  :deep(.el-card__header) {
    padding: $spacing-3 $spacing-6;
  }
}

.chart-title {
  font-weight: $font-weight-semibold;
  font-size: $font-size-sm;
}

.chart-container {
  width: 100%;
  height: 250px;
}
</style>
