<template>
  <el-card class="cluster-node-card" shadow="hover" :class="statusClass">
    <!-- 卡片头部：节点名称和状态 -->
    <template #header>
      <div class="node-header">
        <div class="node-identity">
          <div class="node-icon" :style="{ background: statusColor }">
            <el-icon :size="20" color="#fff"><Monitor /></el-icon>
          </div>
          <div class="node-info">
            <span class="node-name" :title="node.name">{{ node.name }}</span>
            <span class="node-ip">{{ node.ip || '--' }}</span>
          </div>
        </div>
        <el-tag :type="statusTagType" size="small" effect="dark" class="status-tag">
          {{ statusText }}
        </el-tag>
      </div>
    </template>

    <!-- 资源使用率 -->
    <div class="node-resources">
      <!-- CPU 使用率 -->
      <div class="resource-item">
        <div class="resource-label">
          <el-icon :size="14"><Cpu /></el-icon>
          <span>CPU</span>
          <span class="resource-value">{{ cpuPercent.toFixed(1) }}%</span>
        </div>
        <el-progress
          :percentage="cpuPercent"
          :color="getResourceColor(cpuPercent)"
          :show-text="false"
          :stroke-width="6"
        />
      </div>

      <!-- 内存使用率 -->
      <div class="resource-item">
        <div class="resource-label">
          <el-icon :size="14"><Memo /></el-icon>
          <span>内存</span>
          <span class="resource-value">{{ memoryPercent.toFixed(1) }}%</span>
        </div>
        <el-progress
          :percentage="memoryPercent"
          :color="getResourceColor(memoryPercent)"
          :show-text="false"
          :stroke-width="6"
        />
      </div>

      <!-- 磁盘使用率 -->
      <div class="resource-item">
        <div class="resource-label">
          <el-icon :size="14"><Coin /></el-icon>
          <span>磁盘</span>
          <span class="resource-value">{{ diskPercent.toFixed(1) }}%</span>
        </div>
        <el-progress
          :percentage="diskPercent"
          :color="getResourceColor(diskPercent)"
          :show-text="false"
          :stroke-width="6"
        />
      </div>
    </div>

    <!-- 状态信息 -->
    <div class="node-stats">
      <div class="stat-item">
        <span class="stat-label">虚拟机</span>
        <span class="stat-value">
          <span class="stat-running">{{ node.vmCount }}</span>
          <span class="stat-total">/{{ node.vmTotal }}</span>
        </span>
      </div>
      <div class="stat-item">
        <span class="stat-label">网络</span>
        <span class="stat-value network-stats">
          <span class="net-in" title="入站">
            <el-icon :size="12"><Bottom /></el-icon>
            {{ formatNetworkRate(node.netin) }}
          </span>
          <span class="net-out" title="出站">
            <el-icon :size="12"><Top /></el-icon>
            {{ formatNetworkRate(node.netout) }}
          </span>
        </span>
      </div>
      <div class="stat-item">
        <span class="stat-label">运行时间</span>
        <span class="stat-value">{{ formatUptime(node.uptime) }}</span>
      </div>
    </div>

    <!-- 节点操作 -->
    <div class="node-actions">
      <el-dropdown trigger="click" @command="handleAction">
        <el-button text size="small">
          <el-icon><More /></el-icon>
          操作
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="detail">
              <el-icon><View /></el-icon>
              查看详情
            </el-dropdown-item>
            <el-dropdown-item command="console">
              <el-icon><Monitor /></el-icon>
              打开终端
            </el-dropdown-item>
            <el-dropdown-item divided command="refresh">
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { formatBytes, formatUptime, formatNetworkRate } from '@/utils/format'
import {
  Monitor,
  Cpu,
  Memo,
  Coin,
  Bottom,
  Top,
  More,
  View,
  Refresh,
} from '@element-plus/icons-vue'
import type { ClusterNode } from '@/api/taskTypes'

interface Props {
  /** 节点数据 */
  node: ClusterNode
}

const props = defineProps<Props>()

const emit = defineEmits<{
  action: [command: string, node: ClusterNode]
}>()

// ===== 计算属性 =====

/** CPU 使用百分比 (0-100) */
const cpuPercent = computed(() => Math.round((props.node.cpu || 0) * 1000) / 10)

/** 内存使用百分比 (0-100) */
const memoryPercent = computed(() => {
  if (!props.node.maxmem) return 0
  return Math.round((props.node.mem / props.node.maxmem) * 1000) / 10
})

/** 磁盘使用百分比 (0-100) */
const diskPercent = computed(() => {
  if (!props.node.maxdisk) return 0
  return Math.round((props.node.disk / props.node.maxdisk) * 1000) / 10
})

/** 根据节点状态返回 CSS 类名 */
const statusClass = computed(() => `node-status-${props.node.status}`)

/** 根据节点状态返回标签类型 */
const statusTagType = computed(() => {
  const map: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
    online: 'success',
    offline: 'danger',
    warning: 'warning',
  }
  return map[props.node.status] || 'info'
})

/** 根据节点状态返回中文文本 */
const statusText = computed(() => {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    warning: '告警',
  }
  return map[props.node.status] || '未知'
})

/** 根据节点状态返回图标背景色 */
const statusColor = computed(() => {
  const map: Record<string, string> = {
    online: '#52c41a',
    offline: '#8c8c8c',
    warning: '#faad14',
  }
  return map[props.node.status] || '#8c8c8c'
})

// ===== 方法 =====

/**
 * 根据使用率返回进度条颜色
 */
function getResourceColor(percent: number): string {
  if (percent < 50) return '#52c41a'
  if (percent < 75) return '#faad14'
  return '#f5222d'
}

/**
 * 处理节点操作命令
 */
function handleAction(command: string) {
  emit('action', command, props.node)
  ElMessage.success(`操作: ${command} - ${props.node.name}`)
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.cluster-node-card {
  border-radius: $radius-base;
  transition: $transition-base;

  &:hover {
    transform: translateY(-2px);
    box-shadow: $shadow-card-hover;
  }

  &.node-status-online {
    border-top: 3px solid $success-6;
  }
  &.node-status-offline {
    border-top: 3px solid $gray-5;
    opacity: 0.7;
  }
  &.node-status-warning {
    border-top: 3px solid $warning-6;
  }

  :deep(.el-card__header) {
    padding: $spacing-4 $spacing-5;
    border-bottom: none;
  }
}

// 卡片头部
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .node-identity {
    display: flex;
    align-items: center;
    gap: $spacing-3;
    min-width: 0;

    .node-icon {
      width: 36px;
      height: 36px;
      border-radius: $radius-base;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    .node-info {
      display: flex;
      flex-direction: column;
      min-width: 0;

      .node-name {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $color-text-primary;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .node-ip {
        font-size: $font-size-xs;
        color: $color-text-secondary;
        font-family: $font-family-code;
      }
    }
  }
}

// 资源使用区
.node-resources {
  display: flex;
  flex-direction: column;
  gap: $spacing-3;
  margin-bottom: $spacing-4;
}

.resource-item {
  .resource-label {
    display: flex;
    align-items: center;
    gap: $spacing-2;
    margin-bottom: $spacing-1;
    font-size: $font-size-xs;
    color: $color-text-secondary;

    .el-icon {
      color: $color-text-secondary;
    }

    .resource-value {
      margin-left: auto;
      font-weight: $font-weight-medium;
      color: $color-text-regular;
    }
  }
}

// 统计信息
.node-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-3;
  padding-top: $spacing-4;
  border-top: 1px solid $color-border-lighter;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-1;

  .stat-label {
    font-size: $font-size-xs;
    color: $color-text-secondary;
  }

  .stat-value {
    font-size: $font-size-sm;
    font-weight: $font-weight-medium;
    color: $color-text-primary;

    .stat-running {
      color: $color-success;
    }

    .stat-total {
      color: $color-text-secondary;
      font-size: $font-size-xs;
    }

    &.network-stats {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: $spacing-1;
      font-size: $font-size-xs;

      .net-in {
        color: $color-success;
        display: flex;
        align-items: center;
        gap: 2px;
      }

      .net-out {
        color: $color-primary;
        display: flex;
        align-items: center;
        gap: 2px;
      }
    }
  }
}

// 操作区
.node-actions {
  margin-top: $spacing-4;
  padding-top: $spacing-3;
  border-top: 1px solid $color-border-lighter;
  display: flex;
  justify-content: flex-end;
}
</style>
