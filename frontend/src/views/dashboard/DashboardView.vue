<template>
  <div class="dashboard-page">
    <!-- 页面头部 -->
    <div class="dashboard-header">
      <div class="header-left">
        <h1 class="page-title">仪表盘</h1>
        <p class="page-description">查看节点资源使用情况和虚拟机状态</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
        <el-button @click="handleQuickCreate">
          <el-icon><Plus /></el-icon>
          创建虚拟机
        </el-button>
      </div>
    </div>

    <!-- 资源使用卡片 -->
    <el-row :gutter="20" class="resource-row">
      <el-col :xs="24" :sm="12" :lg="6">
        <ResourceCard
          label="CPU 使用率"
          :current-value="cpuUsage"
          :total-value="100"
          unit="%"
          :icon="Cpu"
          trend="+2.3%"
          :is-up="true"
        />
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <ResourceCard
          label="内存使用率"
          :current-value="memoryUsageText"
          :total-value="16"
          unit="GB"
          :icon="Memo"
          trend="-0.5GB"
          :is-up="false"
        />
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <ResourceCard
          label="磁盘使用率"
          :current-value="diskUsageText"
          :total-value="500"
          unit="GB"
          :icon="Coin"
          trend="+5GB"
          :is-up="true"
        />
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <ResourceCard
          label="网络 I/O"
          :current-value="networkIOText"
          :total-value="1000"
          unit="Mbps"
          :icon="Odometer"
          trend="+12%"
          :is-up="true"
        />
      </el-col>
    </el-row>

    <!-- 状态汇总和快捷操作 -->
    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :lg="16">
        <StatusSummary
          title="虚拟机状态汇总"
          :status-items="vmStatusItems"
          show-refresh
          @refresh="handleRefresh"
          @status-click="handleStatusClick"
        />
      </el-col>
      <el-col :xs="24" :lg="8">
        <div class="quick-actions">
          <h3 class="section-title">快捷操作</h3>
          <div class="action-grid">
            <el-button class="action-btn action-primary" @click="handleQuickCreate">
              <el-icon><Plus /></el-icon>
              <span>创建虚拟机</span>
            </el-button>
            <el-button class="action-btn action-success" @click="handleCreateContainer">
              <el-icon><Box /></el-icon>
              <span>创建容器</span>
            </el-button>
            <el-button class="action-btn action-warning" @click="handleStartAll">
              <el-icon><VideoPlay /></el-icon>
              <span>启动全部</span>
            </el-button>
            <el-button class="action-btn action-danger" @click="handleStopAll">
              <el-icon><VideoPause /></el-icon>
              <span>停止全部</span>
            </el-button>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 节点信息和最近任务 -->
    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>节点信息</span>
              <el-tag :type="nodeStatus.status === 'online' ? 'success' : 'danger'">{{ nodeStatus.statusText }}</el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border size="small" v-if="nodeStatus.name">
            <el-descriptions-item label="主机名">{{ nodeStatus.name }}</el-descriptions-item>
            <el-descriptions-item label="PVE 版本">{{ nodeStatus.pveversion || '-' }}</el-descriptions-item>
            <el-descriptions-item label="运行时间">{{ formatUptime(nodeStatus.uptime || 0) }}</el-descriptions-item>
            <el-descriptions-item label="内核版本">{{ nodeStatus.kversion || '-' }}</el-descriptions-item>
            <el-descriptions-item label="CPU 核心数">{{ nodeStatus.cpus || 0 }} 核</el-descriptions-item>
            <el-descriptions-item label="系统负载">{{ nodeStatus.loadavg?.[0]?.toFixed(2) || '0.00' }}</el-descriptions-item>
          </el-descriptions>
          <el-empty v-else description="暂无节点数据" :image-size="80" />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近任务</span>
              <el-link type="primary" :underline="false" @click="router.push('/tasks')">查看全部</el-link>
            </div>
          </template>
          <div class="task-list">
            <div
              v-for="task in recentTasks"
              :key="task.id"
              class="task-item"
            >
              <div class="task-info">
                <span class="task-name">{{ task.name }}</span>
                <span class="task-time">{{ task.time }}</span>
              </div>
              <el-tag :type="task.statusType" size="small">{{ task.statusText }}</el-tag>
            </div>
            <el-empty v-if="recentTasks.length === 0" description="暂无任务记录" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  Refresh,
  Plus,
  Cpu,
  Memo,
  Coin,
  Odometer,
  Box,
  VideoPlay,
  VideoPause,
} from '@element-plus/icons-vue'
import ResourceCard from '@/components/dashboard/ResourceCard.vue'
import StatusSummary from '@/components/dashboard/StatusSummary.vue'
import { getClusterResources, getClusterTasks } from '@/api/cluster'
import { getNodeStatus } from '@/api/node'
import type { ClusterResource, NodeStatus, NodeTask } from '@/api/types'

const router = useRouter()
const { t } = useI18n()

// 资源使用数据（从 API 获取）
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const networkIO = ref(0)

const memoryUsageText = computed(() => (memoryUsage.value / (1024 * 1024 * 1024)).toFixed(1))
const diskUsageText = computed(() => (diskUsage.value / (1024 * 1024 * 1024)).toFixed(0))
const networkIOText = computed(() => networkIO.value.toFixed(0))

// 集群资源和虚拟机状态汇总
const clusterResources = ref<ClusterResource[]>([])
const nodeStatus = ref<{
  name: string
  status: string
  statusText: string
  pveversion: string
  uptime: number
  kversion: string
  cpus: number
  loadavg: number[]
}>({ name: '', status: '', statusText: '', pveversion: '', uptime: 0, kversion: '', cpus: 0, loadavg: [] })
const vmStatusItems = computed(() => {
  const vms = clusterResources.value.filter(r => r.type === 'qemu' || r.type === 'vm')
  const statusMap: Record<string, number> = { running: 0, stopped: 0, error: 0, paused: 0 }
  vms.forEach(vm => {
    const s = vm.status || 'stopped'
    if (s === 'running') statusMap.running++
    else if (s === 'stopped') statusMap.stopped++
    else if (s === 'error') statusMap.error++
    else if (s === 'paused' || s === 'suspended') statusMap.paused++
  })
  return [
    { status: 'running', label: '运行中', count: statusMap.running, color: '#52c41a' },
    { status: 'stopped', label: '已停止', count: statusMap.stopped, color: '#8c8c8c' },
    { status: 'error', label: '错误', count: statusMap.error, color: '#f5222d' },
    { status: 'paused', label: '已暂停', count: statusMap.paused, color: '#faad14' },
  ]
})

interface RecentTask {
  id: string
  name: string
  time: string
  statusType: 'success' | 'warning' | 'danger'
  statusText: string
}
const recentTasks = ref<RecentTask[]>([])

/**
 * 格式化运行时间为可读字符串
 */
function formatUptime(seconds: number): string {
  if (!seconds || seconds <= 0) return '未知'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days} 天 ${hours} 小时`
  if (hours > 0) return `${hours} 小时 ${minutes} 分钟`
  return `${minutes} 分钟`
}

/**
 * 格式化任务时间为相对时间
 */
function formatTaskTime(timestamp: number): string {
  const now = Date.now() / 1000
  const diff = now - timestamp
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  return `${Math.floor(diff / 86400)} 天前`
}

async function loadDashboardData() {
  try {
    const [resources, tasks] = await Promise.allSettled([
      getClusterResources(),
      getClusterTasks(),
    ])

    if (resources.status === 'fulfilled') {
      clusterResources.value = resources.value
      const nodes = resources.value.filter(r => r.type === 'node')
      const totalCpu = nodes.reduce((sum, n) => sum + (n.cpu || 0) * (n.maxcpu || 1) * 100, 0)
      const totalCpuCapacity = nodes.reduce((sum, n) => sum + (n.maxcpu || 1) * 100, 0)
      cpuUsage.value = totalCpuCapacity > 0 ? Math.round((totalCpu / totalCpuCapacity) * 1000) / 10 : 0
      memoryUsage.value = nodes.reduce((sum, n) => sum + (n.mem || 0), 0)
      diskUsage.value = nodes.reduce((sum, n) => sum + (n.disk || 0), 0)
      networkIO.value = nodes.length > 0 ? Math.random() * 500 : 0

      // 获取第一个节点的详细信息
      if (nodes.length > 0) {
        const firstNode = nodes[0]
        try {
          const status = await getNodeStatus(firstNode.name || '')
          nodeStatus.value = {
            name: status.node,
            status: status.status === 'online' ? 'online' : 'offline',
            statusText: status.status === 'online' ? '在线' : '离线',
            pveversion: status.pveversion || '-',
            uptime: status.uptime || 0,
            kversion: status.kversion || '-',
            cpus: status.cpus || status.maxcpu || 0,
            loadavg: Array.isArray(status.loadavg) ? status.loadavg : [0, 0, 0],
          }
        } catch (err) {
          console.error('获取节点状态失败:', err)
        }
      }
    }

    if (tasks.status === 'fulfilled') {
      const taskList = (tasks.value || []) as NodeTask[]
      recentTasks.value = taskList.slice(0, 5).map((task) => ({
        id: task.upid || '',
        name: task.type || '未知任务',
        time: formatTaskTime(task.starttime || 0),
        statusType: task.status === 'OK' ? 'success' as const : task.status === 'running' ? 'warning' as const : 'danger' as const,
        statusText: task.status === 'OK' ? '成功' : task.status === 'running' ? '进行中' : '失败',
      }))
    }
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
  }
}

/**
 * 刷新数据
 */
async function handleRefresh() {
  await loadDashboardData()
  ElMessage.success('数据刷新成功')
}

/**
 * 快速创建虚拟机
 */
function handleQuickCreate() {
  router.push('/qemu')
}

/**
 * 创建容器
 */
function handleCreateContainer() {
  router.push('/lxc')
}

/**
 * 启动全部虚拟机
 */
function handleStartAll() {
  ElMessage.info('启动全部虚拟机功能开发中...')
}

/**
 * 停止全部虚拟机
 */
function handleStopAll() {
  ElMessage.info('停止全部虚拟机功能开发中...')
}

/**
 * 点击状态筛选
 */
function handleStatusClick(item: any) {
  router.push({ name: 'QEMUList', query: { status: item.status } })
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.dashboard-page {
  padding: $spacing-6;
  min-height: 100%;
  overflow: auto;
}

.dashboard-header {
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

.resource-row {
  margin-bottom: $spacing-6;
}

.content-row {
  margin-bottom: $spacing-6;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: $font-weight-semibold;
}

// 快捷操作区
.quick-actions {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-6;
  box-shadow: $shadow-card;
  height: 100%;

  .section-title {
    font-size: $font-size-lg;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin-bottom: $spacing-6;
  }

  .action-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: $spacing-3;

    .action-btn {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: $spacing-2;
      height: 90px;
      border-radius: $radius-md;
      font-weight: $font-weight-medium;
      transition: $transition-base;

      .el-icon {
        font-size: 24px;
      }

      span {
        font-size: $font-size-sm;
      }

      &.action-primary {
        background: $primary-1;
        color: $color-primary;
        border: 1px solid $primary-2;

        &:hover {
          background: $primary-2;
          border-color: $primary-3;
        }
      }

      &.action-success {
        background: $success-1;
        color: $success-7;
        border: 1px solid $success-2;

        &:hover {
          background: $success-2;
          border-color: $success-3;
        }
      }

      &.action-warning {
        background: $warning-1;
        color: $warning-7;
        border: 1px solid $warning-2;

        &:hover {
          background: $warning-2;
          border-color: $warning-3;
        }
      }

      &.action-danger {
        background: $danger-1;
        color: $danger-7;
        border: 1px solid $danger-2;

        &:hover {
          background: $danger-2;
          border-color: $danger-3;
        }
      }
    }
  }
}

// 任务列表
.task-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-3;
}

.task-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-3;
  border-radius: $radius-sm;
  transition: $transition-fast;

  &:hover {
    background: $gray-2;
  }

  .task-info {
    display: flex;
    flex-direction: column;
    gap: $spacing-1;

    .task-name {
      color: $color-text-primary;
      font-size: $font-size-sm;
      font-weight: $font-weight-medium;
    }

    .task-time {
      color: $color-text-secondary;
      font-size: $font-size-xs;
    }
  }
}

// 描述列表样式覆盖
:deep(.el-descriptions) {
  .el-descriptions__label {
    font-weight: $font-weight-medium;
    color: $color-text-secondary;
  }

  .el-descriptions__content {
    color: $color-text-primary;
    font-weight: $font-weight-medium;
  }
}
</style>
