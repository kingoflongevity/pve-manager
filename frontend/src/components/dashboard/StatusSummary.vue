<template>
  <div class="status-summary">
    <!-- 标题 -->
    <div class="summary-header">
      <h3 class="summary-title">{{ title }}</h3>
      <el-button v-if="showRefresh" text @click="$emit('refresh')">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <!-- 状态列表 -->
    <div class="status-list">
      <div
        v-for="item in statusItems"
        :key="item.status"
        class="status-item"
        @click="handleClick(item)"
      >
        <div class="status-indicator" :style="{ background: item.color }"></div>
        <span class="status-label">{{ item.label }}</span>
        <span class="status-count">{{ item.count }}</span>
      </div>
    </div>

    <!-- 可视化条 -->
    <div class="status-bar">
      <div
        v-for="item in statusItems"
        :key="item.status"
        class="status-bar-segment"
        :style="{
          width: `${getPercentage(item.count)}%`,
          background: item.color,
        }"
      ></div>
    </div>

    <!-- 统计信息 -->
    <div class="summary-footer">
      <span class="total-text">总计: {{ totalCount }} 台</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Refresh } from '@element-plus/icons-vue'

interface StatusItem {
  status: string
  label: string
  count: number
  color: string
}

interface Props {
  title?: string
  statusItems: StatusItem[]
  showRefresh?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '状态汇总',
  showRefresh: false,
})

defineEmits<{
  refresh: []
  'status-click': [item: StatusItem]
}>()

// 计算总数
const totalCount = computed(() =>
  props.statusItems.reduce((sum, item) => sum + item.count, 0)
)

// 计算百分比
function getPercentage(count: number): number {
  if (totalCount.value === 0) return 0
  return (count / totalCount.value) * 100
}

function handleClick(item: StatusItem) {
  // TODO: 根据状态过滤列表
  console.log('点击状态:', item)
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.status-summary {
  background: $color-bg-container;
  border-radius: $radius-lg;
  padding: $spacing-6;
  box-shadow: $shadow-base;
  transition: $transition-slow;
  border: 1px solid $color-border-light;

  &:hover {
    box-shadow: $shadow-lg;
  }
}

.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-4;

  .summary-title {
    font-size: $font-size-lg;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin: 0;
    letter-spacing: 0.025em;
  }
}

.status-list {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-4;
  margin-bottom: $spacing-4;
}

.status-item {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  cursor: pointer;
  padding: $spacing-3 $spacing-4;
  border-radius: $radius-base;
  transition: $transition-base;
  background: $color-bg-elevated;
  border: 1px solid transparent;
  flex: 1;
  min-width: 120px;

  &:hover {
    background: $color-bg-hover;
    border-color: var(--status-color, $green-500);
  }

  .status-indicator {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 8px var(--status-color, $green-500);
  }

  .status-label {
    color: $color-text-secondary;
    font-size: $font-size-sm;
    font-weight: $font-weight-medium;
    flex: 1;
  }

  .status-count {
    color: $color-text-primary;
    font-weight: $font-weight-bold;
    font-size: $font-size-xl;
    font-family: $font-family-mono;
  }
}

.status-bar {
  height: 8px;
  background: $color-border-light;
  border-radius: $radius-full;
  overflow: hidden;
  margin-bottom: $spacing-4;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);

  .status-bar-segment {
    height: 100%;
    transition: width $transition-slow;
  }
}

.summary-footer {
  text-align: right;

  .total-text {
    color: $color-text-secondary;
    font-size: $font-size-sm;
    font-family: $font-family-mono;
  }
}
</style>
