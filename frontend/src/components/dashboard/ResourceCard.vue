<template>
  <div class="resource-card" :style="cardStyle">
    <!-- 图标 -->
    <div class="resource-icon" :style="iconStyle">
      <el-icon :size="28"><component :is="icon" /></el-icon>
    </div>

    <!-- 内容区 -->
    <div class="resource-content">
      <div class="resource-label">{{ label }}</div>
      <div class="resource-value">
        <span class="value-number">{{ currentValue }}</span>
        <span class="value-unit">{{ unit }}</span>
      </div>

      <!-- 进度条 -->
      <div class="resource-progress">
        <el-progress
          :percentage="usagePercent"
          :color="progressColor"
          :show-text="false"
          :stroke-width="8"
        />
        <span class="progress-text">{{ usagePercent.toFixed(1) }}%</span>
      </div>
    </div>

    <!-- 趋势指示 -->
    <div v-if="trend" class="resource-trend" :class="trendClass">
      <el-icon><component :is="trendIcon" /></el-icon>
      <span>{{ trend }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import type { Component } from 'vue'

interface Props {
  /** 资源标签 */
  label: string
  /** 当前值 */
  currentValue: string | number
  /** 总值 */
  totalValue: number
  /** 单位 */
  unit?: string
  /** 图标组件 */
  icon: Component
  /** 趋势描述 */
  trend?: string
  /** 是否上升 */
  isUp?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  unit: '',
  trend: '',
  isUp: true,
})

// 计算使用百分比
const usagePercent = computed(() => {
  const current = typeof props.currentValue === 'number' ? props.currentValue : parseFloat(props.currentValue as string)
  return Math.min((current / props.totalValue) * 100, 100)
})

// 根据使用率返回颜色
const progressColor = computed(() => {
  if (usagePercent.value < 50) return '#52c41a'
  if (usagePercent.value < 75) return '#faad14'
  return '#f5222d'
})

// 趋势样式
const trendClass = computed(() => props.isUp ? 'trend-up' : 'trend-down')

// 趋势图标
const trendIcon = computed(() => props.isUp ? ArrowUp : ArrowDown)

// 卡片样式
const cardStyle = computed(() => ({
  '--resource-bg': getBackgroundColor(),
}))

const iconStyle = computed(() => ({
  background: getIconBackground(),
  color: getIconColor(),
}))

function getBackgroundColor(): string {
  const percent = usagePercent.value
  if (percent < 50) return '#f6ffed'
  if (percent < 75) return '#fffbe6'
  return '#fff2f0'
}

function getIconBackground(): string {
  const percent = usagePercent.value
  if (percent < 50) return '#f6ffed'
  if (percent < 75) return '#fffbe6'
  return '#fff2f0'
}

function getIconColor(): string {
  const percent = usagePercent.value
  if (percent < 50) return '#52c41a'
  if (percent < 75) return '#faad14'
  return '#f5222d'
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.resource-card {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-6;
  box-shadow: $shadow-card;
  transition: $transition-base;
  display: flex;
  flex-direction: column;
  gap: $spacing-4;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: var(--resource-bg);
  }

  &:hover {
    box-shadow: $shadow-card-hover;
    transform: translateY(-2px);
  }
}

.resource-icon {
  width: 52px;
  height: 52px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
}

.resource-content {
  display: flex;
  flex-direction: column;
  gap: $spacing-2;
}

.resource-label {
  color: $color-text-secondary;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
}

.resource-value {
  display: flex;
  align-items: baseline;
  gap: $spacing-1;

  .value-number {
    font-size: $font-size-3xl;
    font-weight: $font-weight-bold;
    color: $color-text-primary;
    line-height: $line-height-xs;
  }

  .value-unit {
    font-size: $font-size-base;
    color: $color-text-secondary;
  }
}

.resource-progress {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  margin-top: $spacing-2;

  :deep(.el-progress-bar) {
    flex: 1;
  }

  .progress-text {
    font-size: $font-size-sm;
    color: $color-text-regular;
    font-weight: $font-weight-medium;
    min-width: 48px;
    text-align: right;
  }
}

.resource-trend {
  display: flex;
  align-items: center;
  gap: $spacing-1;
  font-size: $font-size-xs;
  margin-top: auto;

  &.trend-up {
    color: $danger-6;
  }

  &.trend-down {
    color: $success-6;
  }
}
</style>
