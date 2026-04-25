<template>
  <div class="resource-gauge">
    <div class="gauge-container">
      <svg class="gauge-svg" viewBox="0 0 100 100">
        <!-- 背景圆环 -->
        <circle
          class="gauge-bg"
          cx="50"
          cy="50"
          r="42"
          fill="none"
          stroke-width="8"
        />
        <!-- 进度圆环 -->
        <circle
          class="gauge-progress"
          cx="50"
          cy="50"
          r="42"
          fill="none"
          stroke-width="8"
          :stroke-dasharray="dashArray"
          :stroke-dashoffset="dashOffset"
        />
        <!-- 中心文字 -->
        <text class="gauge-value" x="50" y="45" text-anchor="middle">{{ displayValue }}</text>
        <text class="gauge-unit" x="50" y="60" text-anchor="middle">{{ unit }}</text>
      </svg>
    </div>
    <div class="gauge-label">{{ label }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

/**
 * 可复用的资源仪表盘组件
 * - 使用 SVG 实现圆形仪表盘
 * - 根据使用率自动切换颜色（绿/橙/红）
 * - 支持动画过渡效果
 */

interface Props {
  /** 当前值 */
  value: number
  /** 最大值 */
  max?: number
  /** 标签文字 */
  label: string
  /** 单位 */
  unit?: string
  /** 强制颜色（不提供则根据使用率自动计算） */
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  max: 100,
  unit: '%',
  color: undefined,
})

/**
 * 计算百分比
 */
const percentage = computed(() => {
  if (props.max === 0) return 0
  return Math.min((props.value / props.max) * 100, 100)
})

/**
 * 显示值
 */
const displayValue = computed(() => {
  const val = percentage.value
  return val % 1 === 0 ? val.toFixed(0) : val.toFixed(1)
})

/**
 * SVG 圆环的 dasharray
 * 周长 = 2 * PI * r = 2 * 3.14159 * 42 ≈ 263.89
 */
const circumference = 2 * Math.PI * 42
const dashArray = computed(() => `${circumference.toFixed(2)} ${circumference.toFixed(2)}`)

/**
 * SVG 圆环的 dashoffset
 */
const dashOffset = computed(() => {
  return circumference - (percentage.value / 100) * circumference
})

/**
 * 根据使用率自动计算颜色
 * - < 60%: 绿色
 * - 60-80%: 橙色
 * - > 80%: 红色
 */
const gaugeColor = computed(() => {
  if (props.color) return props.color
  const pct = percentage.value
  if (pct < 60) return '#52c41a'
  if (pct < 80) return '#faad14'
  return '#f5222d'
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.resource-gauge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-2;
}

.gauge-container {
  width: 120px;
  height: 120px;
  position: relative;
}

.gauge-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.gauge-bg {
  stroke: #f0f0f0;
}

.gauge-progress {
  stroke-linecap: round;
  transition: stroke-dashoffset 0.6s cubic-bezier(0.4, 0, 0.2, 1), stroke 0.3s ease;
}

// 应用动态颜色
:deep(.gauge-progress) {
  stroke: v-bind(gaugeColor);
}

.gauge-value {
  font-size: 18px;
  font-weight: $font-weight-bold;
  fill: $color-text-primary;
  transform: rotate(90deg);
  transform-origin: center;
}

.gauge-unit {
  font-size: 11px;
  fill: $color-text-secondary;
  transform: rotate(90deg);
  transform-origin: center;
}

.gauge-label {
  font-size: $font-size-sm;
  color: $color-text-regular;
  text-align: center;
}
</style>
