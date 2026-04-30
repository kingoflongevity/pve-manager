<template>
  <div class="node-list-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">节点管理</h1>
        <p class="page-description">管理所有 Proxmox VE 节点及其网络配置</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
      </div>
    </div>

    <!-- 汇总统计条 -->
    <div class="summary-bar">
      <div class="summary-item" v-for="item in summaryItems" :key="item.label">
        <div class="summary-icon" :style="{ background: item.bgColor, color: item.color }">
          <el-icon :size="22"><component :is="item.icon" /></el-icon>
        </div>
        <div class="summary-info">
          <span class="summary-value">{{ item.value }}</span>
          <span class="summary-label">{{ item.label }}</span>
        </div>
      </div>
    </div>

    <!-- 筛选控制区 -->
    <div class="filter-bar">
      <el-radio-group v-model="statusFilter" size="default">
        <el-radio-button value="all">全部</el-radio-button>
        <el-radio-button value="online">在线</el-radio-button>
        <el-radio-button value="warning">告警</el-radio-button>
        <el-radio-button value="offline">离线</el-radio-button>
      </el-radio-group>
      <el-input
        v-model="searchQuery"
        placeholder="搜索节点名称..."
        :prefix-icon="Search"
        clearable
        class="search-input"
      />
    </div>

    <!-- 加载/空状态 -->
    <div v-if="loading && filteredNodes.length === 0" class="loading-state">
      <el-skeleton :rows="3" animated />
    </div>
    <el-empty v-else-if="filteredNodes.length === 0" description="暂无节点数据" />

    <!-- 节点卡片网格 -->
    <div v-else class="node-grid">
      <el-card
        v-for="node in filteredNodes"
        :key="node.node"
        class="node-card"
        shadow="hover"
        :class="getNodeStatusClass(node.status)"
        @click="goToNodeDetail(node.node)"
      >
        <template #header>
          <div class="card-header">
            <div class="node-identity">
              <div class="node-icon" :style="{ background: getNodeStatusColor(node.status) }">
                <el-icon :size="18" color="#fff"><Monitor /></el-icon>
              </div>
              <div class="node-info">
                <span class="node-name">{{ node.node }}</span>
                <span class="node-status" :class="getNodeStatusText(node.status)">{{ getNodeStatusText(node.status) }}</span>
              </div>
            </div>
            <el-tag :type="getNodeTagType(node.status)" size="small" effect="dark">
              {{ getNodeStatusText(node.status) }}
            </el-tag>
          </div>
        </template>

        <!-- 资源使用 -->
        <div class="node-resources">
          <div class="resource-item">
            <div class="resource-label">
              <el-icon :size="14"><Cpu /></el-icon>
              <span>CPU</span>
              <span class="resource-value">{{ getResourcePercent(node.cpu, node.maxcpu) }}%</span>
            </div>
            <el-progress
              :percentage="getResourcePercent(node.cpu, node.maxcpu)"
              :color="getResourceColor(getResourcePercent(node.cpu, node.maxcpu))"
              :show-text="false"
              :stroke-width="6"
            />
          </div>
          <div class="resource-item">
            <div class="resource-label">
              <el-icon :size="14"><Memo /></el-icon>
              <span>内存</span>
              <span class="resource-value">{{ getResourcePercent(node.mem, node.maxmem) }}%</span>
            </div>
            <el-progress
              :percentage="getResourcePercent(node.mem, node.maxmem)"
              :color="getResourceColor(getResourcePercent(node.mem, node.maxmem))"
              :show-text="false"
              :stroke-width="6"
            />
          </div>
        </div>

        <!-- 统计信息 -->
        <div class="node-stats">
          <div class="stat-item">
            <span class="stat-label">运行时间</span>
            <span class="stat-value">{{ formatUptime(node.uptime || 0) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">磁盘使用</span>
            <span class="stat-value">{{ formatBytes(node.disk || 0) }} / {{ formatBytes(node.maxdisk || 0) }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="node-actions">
          <el-button type="primary" size="small" text @click.stop="goToNodeDetail(node.node)">
            <el-icon><View /></el-icon>
            查看详情
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Refresh,
  Monitor,
  Cpu,
  Memo,
  Search,
  View,
} from '@element-plus/icons-vue'
import { getClusterResources } from '@/api/cluster'
import type { ClusterResource } from '@/api/types'
import { formatBytes, formatUptime } from '@/utils/format'

const router = useRouter()

// ============================================================
// 状态管理
// ============================================================

/** 节点数据列表 */
const nodes = ref<ClusterResource[]>([])
/** 加载状态 */
const loading = ref(false)
/** 状态筛选 */
const statusFilter = ref<string>('all')
/** 搜索关键词 */
const searchQuery = ref('')
/** 自动刷新定时器 */
let refreshTimer: ReturnType<typeof setInterval> | null = null

// ============================================================
// 计算属性
// ============================================================

/** 过滤后的节点列表 */
const filteredNodes = computed(() => {
  let result = nodes.value.filter((n) => n.type === 'node')

  // 按状态筛选
  if (statusFilter.value !== 'all') {
    result = result.filter((n) => n.status === statusFilter.value)
  }

  // 按搜索关键词筛选
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.trim().toLowerCase()
    result = result.filter((n) => (n.node || '').toLowerCase().includes(query))
  }

  return result
})

/** 汇总统计项 */
const summaryItems = computed(() => {
  const allNodes = nodes.value.filter((n) => n.type === 'node')
  const onlineCount = allNodes.filter((n) => n.status === 'online').length
  const warningCount = allNodes.filter((n) => n.status === 'warning').length
  const offlineCount = allNodes.filter((n) => n.status === 'offline').length

  return [
    { label: '总节点数', value: allNodes.length, icon: Monitor, bgColor: '#e8f3ff', color: '#1677ff' },
    { label: '在线节点', value: onlineCount, icon: Monitor, bgColor: '#f6ffed', color: '#52c41a' },
    { label: '告警节点', value: warningCount, icon: Monitor, bgColor: '#fffbe6', color: '#faad14' },
    { label: '离线节点', value: offlineCount, icon: Monitor, bgColor: '#fff2f0', color: '#f5222d' },
  ]
})

// ============================================================
// 工具函数
// ============================================================

/**
 * 计算资源使用百分比
 */
function getResourcePercent(used?: number, total?: number): number {
  if (!total || total === 0) return 0
  return Math.round((used || 0) / total * 1000) / 10
}

/**
 * 根据使用率返回进度条颜色
 */
function getResourceColor(percent: number): string {
  if (percent < 50) return '#52c41a'
  if (percent < 75) return '#faad14'
  return '#f5222d'
}

/**
 * 获取节点状态对应的 CSS 类名
 */
function getNodeStatusClass(status?: string): string {
  return `node-status-${status || 'unknown'}`
}

/**
 * 获取节点状态对应的颜色
 */
function getNodeStatusColor(status?: string): string {
  const map: Record<string, string> = {
    online: '#52c41a',
    offline: '#8c8c8c',
    warning: '#faad14',
  }
  return map[status || ''] || '#8c8c8c'
}

/**
 * 获取节点状态文本
 */
function getNodeStatusText(status?: string): string {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    warning: '告警',
  }
  return map[status || ''] || '未知'
}

/**
 * 获取节点状态对应的标签类型
 */
function getNodeTagType(status?: string): 'success' | 'danger' | 'warning' | 'info' {
  const map: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
    online: 'success',
    offline: 'danger',
    warning: 'warning',
  }
  return map[status || ''] || 'info'
}

// ============================================================
// 事件处理
// ============================================================

/**
 * 从 API 获取节点数据
 */
async function fetchNodes(): Promise<void> {
  loading.value = true
  try {
    const rawResources = await getClusterResources()
    const resources = Array.isArray(rawResources) ? rawResources : (Array.isArray(rawResources?.data) ? rawResources.data : [])
    nodes.value = resources
  } catch (error) {
    console.error('获取节点数据失败:', error)
    ElMessage.error('获取节点数据失败')
  } finally {
    loading.value = false
  }
}

/**
 * 刷新节点数据
 */
async function handleRefresh(): Promise<void> {
  await fetchNodes()
  ElMessage.success('数据刷新成功')
}

/**
 * 跳转到节点详情页
 */
function goToNodeDetail(nodeName: string): void {
  router.push({ name: 'NodeDetail', params: { nodeName } })
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  fetchNodes()
  // 每 30 秒自动刷新
  refreshTimer = setInterval(fetchNodes, 30_000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.node-list-page {
  padding: $spacing-6;
  min-height: 100%;
  overflow: auto;
}

// 页面头部
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-6;
  gap: $spacing-4;

  @media (max-width: $breakpoint-sm) {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    .page-title {
      font-size: $font-size-3xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      margin: 0 0 $spacing-1;
    }

    .page-description {
      color: $color-text-secondary;
      font-size: $font-size-base;
      margin: 0;
    }
  }

  .header-right {
    display: flex;
    gap: $spacing-3;
  }
}

// 汇总统计条
.summary-bar {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: $spacing-4;
  margin-bottom: $spacing-6;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  padding: $spacing-4;
  background: $color-bg-container;
  border-radius: $radius-base;
  box-shadow: $shadow-card;
  transition: $transition-fast;

  &:hover {
    transform: translateY(-2px);
    box-shadow: $shadow-card-hover;
  }

  .summary-icon {
    width: 44px;
    height: 44px;
    border-radius: $radius-base;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .summary-info {
    display: flex;
    flex-direction: column;

    .summary-value {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      line-height: 1.2;
    }

    .summary-label {
      font-size: $font-size-xs;
      color: $color-text-secondary;
    }
  }
}

// 筛选条
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-4 $spacing-5;
  background: $color-bg-container;
  border-radius: $radius-base;
  box-shadow: $shadow-card;
  margin-bottom: $spacing-6;
  gap: $spacing-4;

  @media (max-width: $breakpoint-md) {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input {
    max-width: 300px;
  }
}

// 加载状态
.loading-state {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-6;
  box-shadow: $shadow-card;
}

// 节点网格
.node-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: $spacing-5;

  @media (max-width: $breakpoint-lg) {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }

  @media (max-width: $breakpoint-sm) {
    grid-template-columns: 1fr;
  }
}

// 节点卡片
.node-card {
  border-radius: $radius-base;
  cursor: pointer;
  transition: $transition-base;

  &:hover {
    transform: translateY(-2px);
    box-shadow: $shadow-card-hover;
  }

  &.node-status-online {
    border-top: 3px solid $success-6;
  }
  &.node-status-offline {
    border-top: 3px solid $color-text-disabled;
    opacity: 0.7;
  }
  &.node-status-warning {
    border-top: 3px solid $warning-6;
  }
  &.node-status-unknown {
    border-top: 3px solid $color-text-placeholder;
  }

  :deep(.el-card__header) {
    padding: $spacing-4 $spacing-5;
    border-bottom: none;
  }
}

.card-header {
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

      .node-status {
        font-size: $font-size-xs;
        color: $color-text-secondary;
      }
    }
  }
}

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

    .resource-value {
      margin-left: auto;
      font-weight: $font-weight-medium;
      color: $color-text-regular;
    }
  }
}

.node-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
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
  }
}

.node-actions {
  margin-top: $spacing-4;
  padding-top: $spacing-3;
  border-top: 1px solid $color-border-lighter;
  display: flex;
  justify-content: flex-end;
}
</style>
