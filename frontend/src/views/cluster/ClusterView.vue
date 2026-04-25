<template>
  <div class="cluster-view-page">
    <!-- 页面头部 -->
    <div class="cluster-header">
      <div class="header-left">
        <h1 class="page-title">集群概览</h1>
        <p class="page-description">统一管理所有 PVE 节点与集群资源</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
      </div>
    </div>

    <!-- 集群汇总条 -->
    <ClusterSummaryBar
      :total-nodes="clusterSummary.totalNodes"
      :online-nodes="clusterSummary.onlineNodes"
      :total-vms="clusterSummary.totalVMs"
      :running-vms="clusterSummary.runningVMs"
      :total-cts="clusterSummary.totalCTs"
      :running-cts="clusterSummary.runningCTs"
      :total-storages="clusterSummary.totalStorages"
    />

    <!-- 筛选和批量操作 -->
    <div class="cluster-controls">
      <div class="filter-section">
        <span class="filter-label">节点状态:</span>
        <el-radio-group v-model="activeFilter" size="default">
          <el-radio-button value="all">全部</el-radio-button>
          <el-radio-button value="online">在线</el-radio-button>
          <el-radio-button value="warning">告警</el-radio-button>
          <el-radio-button value="offline">离线</el-radio-button>
        </el-radio-group>
      </div>
      <div class="batch-actions">
        <el-button size="default" @click="handleBatchAction('startAll')">
          <el-icon><VideoPlay /></el-icon>
          启动全部
        </el-button>
        <el-button size="default" type="warning" @click="handleBatchAction('stopAll')">
          <el-icon><VideoPause /></el-icon>
          停止全部
        </el-button>
        <el-button size="default" @click="handleBatchAction('migrate')">
          <el-icon><Switch /></el-icon>
          迁移任务
        </el-button>
      </div>
    </div>

    <!-- 节点网格 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="filteredNodes.length === 0" class="empty-state">
      <el-empty description="暂无节点数据" />
    </div>

    <div v-else class="node-grid">
      <ClusterNodeCard
        v-for="node in filteredNodes"
        :key="node.name"
        :node="node"
        @action="handleNodeAction"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  VideoPlay,
  VideoPause,
  Switch,
} from '@element-plus/icons-vue'
import ClusterSummaryBar from '@/components/cluster/ClusterSummaryBar.vue'
import ClusterNodeCard from '@/components/cluster/ClusterNodeCard.vue'
import type { ClusterNode, ClusterSummary } from '@/api/taskTypes'

// ===== 筛选状态 =====

const activeFilter = ref<string>('all')

// ===== 集群汇总数据 =====

const clusterSummary = ref<ClusterSummary>({
  totalNodes: 0,
  onlineNodes: 0,
  totalVMs: 0,
  runningVMs: 0,
  totalCTs: 0,
  runningCTs: 0,
  totalStorages: 0,
})

// ===== 节点列表 =====

const nodes = ref<ClusterNode[]>([])
const loading = ref(false)

/** 根据筛选条件过滤后的节点列表 */
const filteredNodes = computed(() => {
  if (activeFilter.value === 'all') return nodes.value
  return nodes.value.filter((n) => n.status === activeFilter.value)
})

// ===== 数据拉取 =====

/**
 * 从 PVE API 拉取集群节点数据
 * 开发阶段使用模拟数据
 */
async function fetchClusterData() {
  loading.value = true
  try {
    // TODO: 实际调用 PVE API
    // const [nodesRes, summaryRes] = await Promise.all([
    //   get('/cluster/nodes'),
    //   get('/cluster/summary'),
    // ])
    // nodes.value = nodesRes.data
    // clusterSummary.value = summaryRes.data

    // 开发阶段模拟数据
    await new Promise((resolve) => setTimeout(resolve, 500))
    nodes.value = generateMockNodes()
    clusterSummary.value = computeSummary(nodes.value)
  } catch (error) {
    console.error('获取集群数据失败:', error)
    ElMessage.error('获取集群数据失败')
  } finally {
    loading.value = false
  }
}

/**
 * 刷新集群数据
 */
async function handleRefresh() {
  await fetchClusterData()
  ElMessage.success('数据刷新成功')
}

// ===== 节点操作 =====

/**
 * 处理节点操作命令
 */
function handleNodeAction(command: string, node: ClusterNode) {
  switch (command) {
    case 'detail':
      ElMessage.info(`查看节点详情: ${node.name}`)
      break
    case 'console':
      ElMessage.info(`打开终端: ${node.name}`)
      break
    case 'refresh':
      handleRefresh()
      break
  }
}

/**
 * 处理批量操作
 */
async function handleBatchAction(action: string) {
  try {
    switch (action) {
      case 'startAll':
        await ElMessageBox.confirm(
          '确定要启动所有节点的虚拟机吗？此操作将影响当前在线节点下的全部虚拟机。',
          '启动全部',
          { type: 'warning' },
        )
        ElMessage.success('正在启动全部虚拟机...')
        break

      case 'stopAll':
        await ElMessageBox.confirm(
          '确定要停止所有节点的虚拟机吗？此操作将影响当前在线节点下的全部虚拟机。',
          '停止全部',
          { type: 'warning' },
        )
        ElMessage.warning('正在停止全部虚拟机...')
        break

      case 'migrate':
        ElMessage.info('迁移任务功能开发中...')
        break
    }
  } catch {
    // 用户取消操作
  }
}

// ===== 生命周期 =====

let refreshTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  fetchClusterData()
  // 每 30 秒自动刷新一次集群数据
  refreshTimer = setInterval(fetchClusterData, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

// ===== 模拟数据 =====

/**
 * 生成模拟节点数据用于开发测试
 */
function generateMockNodes(): ClusterNode[] {
  return [
    {
      name: 'pve-node-01',
      ip: '192.168.1.101',
      status: 'online',
      cpu: 0.35,
      maxmem: 34359738368,
      mem: 12025908428,
      maxdisk: 107374182400,
      disk: 42949672960,
      cpus: 16,
      uptime: 1296000,
      type: 'pve',
      level: '',
      vmCount: 8,
      vmTotal: 12,
      netin: 2048576,
      netout: 1048576,
    },
    {
      name: 'pve-node-02',
      ip: '192.168.1.102',
      status: 'online',
      cpu: 0.62,
      maxmem: 68719476736,
      mem: 42949672960,
      maxdisk: 214748364800,
      disk: 107374182400,
      cpus: 32,
      uptime: 864000,
      type: 'pve',
      level: '',
      vmCount: 15,
      vmTotal: 20,
      netin: 5242880,
      netout: 3145728,
    },
    {
      name: 'pve-node-03',
      ip: '192.168.1.103',
      status: 'warning',
      cpu: 0.88,
      maxmem: 34359738368,
      mem: 30064771072,
      maxdisk: 53687091200,
      disk: 48318382080,
      cpus: 8,
      uptime: 432000,
      type: 'pve',
      level: '',
      vmCount: 5,
      vmTotal: 6,
      netin: 1048576,
      netout: 524288,
    },
    {
      name: 'pve-node-04',
      ip: '192.168.1.104',
      status: 'offline',
      cpu: 0,
      maxmem: 34359738368,
      mem: 0,
      maxdisk: 107374182400,
      disk: 21474836480,
      cpus: 16,
      uptime: 0,
      type: 'pve',
      level: '',
      vmCount: 0,
      vmTotal: 4,
      netin: 0,
      netout: 0,
    },
  ]
}

/**
 * 从节点列表计算集群汇总数据
 */
function computeSummary(nodes: ClusterNode[]): ClusterSummary {
  const totalNodes = nodes.length
  const onlineNodes = nodes.filter((n) => n.status === 'online').length
  const totalVMs = nodes.reduce((sum, n) => sum + n.vmTotal, 0)
  const runningVMs = nodes.reduce((sum, n) => sum + n.vmCount, 0)

  return {
    totalNodes,
    onlineNodes,
    totalVMs,
    runningVMs,
    totalCTs: Math.floor(totalVMs * 0.3), // 模拟数据
    runningCTs: Math.floor(runningVMs * 0.3),
    totalStorages: totalNodes * 2,
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.cluster-view-page {
  padding: $spacing-6;
  min-height: 100%;
  overflow: auto;
}

// 页面头部
.cluster-header {
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
      margin-bottom: $spacing-1;
    }

    .page-description {
      color: $color-text-secondary;
      font-size: $font-size-base;
    }
  }

  .header-right {
    display: flex;
    gap: $spacing-3;

    @media (max-width: $breakpoint-sm) {
      width: 100%;

      .el-button {
        flex: 1;
      }
    }
  }
}

// 筛选和批量操作
.cluster-controls {
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

  .filter-section {
    display: flex;
    align-items: center;
    gap: $spacing-3;

    .filter-label {
      font-size: $font-size-sm;
      color: $color-text-secondary;
      font-weight: $font-weight-medium;
    }
  }

  .batch-actions {
    display: flex;
    gap: $spacing-3;

    @media (max-width: $breakpoint-sm) {
      .el-button {
        flex: 1;
      }
    }
  }
}

// 加载状态
.loading-state {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-6;
  box-shadow: $shadow-card;
}

// 空状态
.empty-state {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-12;
  box-shadow: $shadow-card;
  display: flex;
  align-items: center;
  justify-content: center;
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
</style>
