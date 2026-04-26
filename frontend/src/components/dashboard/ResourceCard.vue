<template>
  <div class="resource-card" :style="{ '--resource-accent': cardAccent }">
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

// 根据使用率返回颜色 (OLED dark theme optimized)
const progressColor = computed(() => {
  if (usagePercent.value < 50) return '#22C55E'
  if (usagePercent.value < 75) return '#F97316'
  return '#EF4444'
})

// 趋势样式
const trendClass = computed(() => props.isUp ? 'trend-up' : 'trend-down')

// 趋势图标
const trendIcon = computed(() => props.isUp ? ArrowUp : ArrowDown)

// 图标样式（背景色和颜色）
const iconStyle = computed(() => ({
  background: getIconBackground(),
  color: getIconColor(),
}))

// 卡片顶部强调色
const cardAccent = getAccentColor()

function getAccentColor(): string {
  const percent = usagePercent.value
  if (percent < 50) return '#22C55E'
  if (percent < 75) return '#F97316'
  return '#EF4444'
}

function getIconBackground(): string {
  const percent = usagePercent.value
  if (percent < 50) return 'rgba(34, 197, 94, 0.15)'
  if (percent < 75) return 'rgba(249, 115, 22, 0.15)'
  return 'rgba(239, 68, 68, 0.15)'
}

function getIconColor(): string {
  const percent = usagePercent.value
  if (percent < 50) return '#22C55E'
  if (percent < 75) return '#F97316'
  return '#EF4444'
}
</script>

<style lang="scss" scoped>
.resource-card {
  background: var(--color-bg-container, #0F172A);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow-base, 0 1px 3px 0 rgba(0, 0, 0, 0.1));
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
  overflow: hidden;
  border: 1px solid var(--color-border-light, rgba(30, 41, 59, 0.4));

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: var(--resource-accent);
    opacity: 0.8;
  }

  &:hover {
    box-shadow: var(--shadow-lg, 0 10px 15px -3px rgba(0, 0, 0, 0.1));
    transform: translateY(-2px);
    border-color: var(--resource-accent);
  }
}

.resource-icon {
  width: 52px;
  height: 52px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.resource-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.resource-label {
  color: var(--color-text-secondary, #94A3B8);
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.025em;
  text-transform: uppercase;
}

.resource-value {
  display: flex;
  align-items: baseline;
  gap: 8px;

  .value-number {
    font-size: 30px;
    font-weight: 700;
    color: var(--color-text-primary, #F8FAFC);
    line-height: 1;
    font-family: 'Fira Code', monospace;
  }

  .value-unit {
    font-size: 14px;
    color: var(--color-text-secondary, #94A3B8);
    font-weight: 500;
  }
}

.resource-progress {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;

  :deep(.el-progress-bar) {
    flex: 1;
  }

  .progress-text {
    font-size: 13px;
    color: var(--color-text-regular, #CBD5E1);
    font-weight: 600;
    min-width: 48px;
    text-align: right;
    font-family: 'Fira Code', monospace;
  }
}

.resource-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  margin-top: auto;
  font-weight: 500;

  &.trend-up {
    color: var(--color-danger, #EF4444);
  }

  &.trend-down {
    color: var(--color-success, #22C55E);
  }
}
</style>
