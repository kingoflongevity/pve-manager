<template>
  <div class="monitoring-tab">
    <div class="toolbar">
      <el-radio-group v-model="selectedTimeframe">
        <el-radio-button value="hour">1 小时</el-radio-button>
        <el-radio-button value="day">1 天</el-radio-button>
        <el-radio-button value="week">1 周</el-radio-button>
        <el-radio-button value="month">1 月</el-radio-button>
        <el-radio-button value="year">1 年</el-radio-button>
      </el-radio-group>
    </div>

    <div class="charts-grid">
      <el-card class="chart-card">
        <template #header><span class="chart-title">CPU 使用率</span></template>
        <RRDChart
          :node="node"
          :target="vmid"
          resource-type="lxc"
          :timeframe="selectedTimeframe"
          dataset="cpu"
          chart-type="line"
          height="250px"
          :yAxisMax="100"
          unit="%"
          :refreshInterval="300"
        />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">内存使用</span></template>
        <RRDChart
          :node="node"
          :target="vmid"
          resource-type="lxc"
          :timeframe="selectedTimeframe"
          dataset="mem"
          chart-type="area"
          height="250px"
          unit="MB"
          :refreshInterval="300"
        />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">磁盘 I/O</span></template>
        <RRDChart
          :node="node"
          :target="vmid"
          resource-type="lxc"
          :timeframe="selectedTimeframe"
          dataset="disk"
          chart-type="bar"
          height="250px"
          unit="B/s"
          :refreshInterval="300"
        />
      </el-card>

      <el-card class="chart-card">
        <template #header><span class="chart-title">网络 I/O</span></template>
        <RRDChart
          :node="node"
          :target="vmid"
          resource-type="lxc"
          :timeframe="selectedTimeframe"
          dataset="net"
          chart-type="area"
          height="250px"
          unit="B/s"
          :refreshInterval="300"
        />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import RRDChart from '@/components/monitor/RRDChart.vue'
import type { RRDTimeframe } from '@/api/types'

interface Props {
  node: string
  vmid: number
}

defineProps<Props>()

const selectedTimeframe = ref<RRDTimeframe>('hour')
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
</style>
