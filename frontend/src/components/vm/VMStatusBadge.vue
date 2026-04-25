<template>
  <div class="vm-status-badge" :class="statusClass">
    <span class="status-dot"></span>
    <span class="status-text">{{ statusText }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  /** 虚拟机状态 */
  status: 'running' | 'stopped' | 'error' | 'paused' | 'unknown'
}

const props = defineProps<Props>()

// 状态映射
const statusMap = {
  running: { text: '运行中', className: 'status-running' },
  stopped: { text: '已停止', className: 'status-stopped' },
  error: { text: '错误', className: 'status-error' },
  paused: { text: '已暂停', className: 'status-paused' },
  unknown: { text: '未知', className: 'status-unknown' },
}

const statusClass = computed(() => statusMap[props.status]?.className || 'status-unknown')
const statusText = computed(() => statusMap[props.status]?.text || '未知')
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.vm-status-badge {
  display: inline-flex;
  align-items: center;
  gap: $spacing-2;
  padding: $spacing-1 $spacing-3;
  border-radius: $radius-full;
  font-size: $font-size-xs;
  font-weight: $font-weight-medium;
  line-height: 1.5;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
    position: relative;
  }
}

// 运行中 - 绿色
.status-running {
  background: $success-1;
  color: $success-7;

  .status-dot {
    background: $success-6;

    &::after {
      content: '';
      position: absolute;
      inset: -2px;
      border-radius: 50%;
      background: rgba(82, 196, 26, 0.4);
      animation: pulse 2s ease-in-out infinite;
    }
  }
}

// 已停止 - 灰色
.status-stopped {
  background: $gray-3;
  color: $gray-8;

  .status-dot {
    background: $gray-6;
  }
}

// 错误 - 红色
.status-error {
  background: $danger-1;
  color: $danger-7;

  .status-dot {
    background: $danger-6;
  }
}

// 已暂停 - 橙色
.status-paused {
  background: $warning-1;
  color: $warning-7;

  .status-dot {
    background: $warning-6;
  }
}

// 未知
.status-unknown {
  background: $info-1;
  color: $info-7;

  .status-dot {
    background: $info-6;
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.8);
    opacity: 0;
  }
}
</style>
