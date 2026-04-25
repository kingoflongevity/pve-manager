<template>
  <div class="cluster-summary-bar">
    <!-- 集群概览统计 -->
    <div class="summary-stats">
      <div class="stat-card" v-for="stat in stats" :key="stat.label">
        <div class="stat-icon" :style="{ background: stat.bgColor, color: stat.color }">
          <el-icon :size="20"><component :is="stat.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import { Monitor, VideoPlay, Box, Coin } from '@element-plus/icons-vue'

interface SummaryStat {
  label: string
  value: number
  icon: Component
  bgColor: string
  color: string
}

interface Props {
  /** 集群节点总数 */
  totalNodes?: number
  /** 在线节点数 */
  onlineNodes?: number
  /** 总虚拟机数 */
  totalVMs?: number
  /** 运行中虚拟机数 */
  runningVMs?: number
  /** 总容器数 */
  totalCTs?: number
  /** 运行中容器数 */
  runningCTs?: number
  /** 总存储数 */
  totalStorages?: number
}

const props = withDefaults(defineProps<Props>(), {
  totalNodes: 0,
  onlineNodes: 0,
  totalVMs: 0,
  runningVMs: 0,
  totalCTs: 0,
  runningCTs: 0,
  totalStorages: 0,
})

/** 概览统计项 */
const stats = computed<SummaryStat[]>(() => [
  {
    label: '集群节点',
    value: props.totalNodes,
    icon: Monitor,
    bgColor: '#e8f3ff',
    color: '#1677ff',
  },
  {
    label: '在线节点',
    value: props.onlineNodes,
    icon: Monitor,
    bgColor: '#f6ffed',
    color: '#52c41a',
  },
  {
    label: '虚拟机',
    value: props.totalVMs,
    icon: VideoPlay,
    bgColor: '#f6ffed',
    color: '#52c41a',
  },
  {
    label: '运行中 VM',
    value: props.runningVMs,
    icon: VideoPlay,
    bgColor: '#f0f5ff',
    color: '#4096ff',
  },
  {
    label: '容器',
    value: props.totalCTs,
    icon: Box,
    bgColor: '#fffbe6',
    color: '#faad14',
  },
  {
    label: '运行中 CT',
    value: props.runningCTs,
    icon: Box,
    bgColor: '#fff7e6',
    color: '#d48806',
  },
  {
    label: '存储',
    value: props.totalStorages,
    icon: Coin,
    bgColor: '#f9f0ff',
    color: '#722ed1',
  },
])
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.cluster-summary-bar {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-5 $spacing-6;
  box-shadow: $shadow-card;
  margin-bottom: $spacing-6;
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: $spacing-4;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  padding: $spacing-3;
  border-radius: $radius-sm;
  background: $gray-2;
  transition: $transition-fast;

  &:hover {
    background: $gray-3;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    border-radius: $radius-base;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-info {
    display: flex;
    flex-direction: column;

    .stat-value {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      line-height: 1.2;
    }

    .stat-label {
      font-size: $font-size-xs;
      color: $color-text-secondary;
    }
  }
}
</style>
