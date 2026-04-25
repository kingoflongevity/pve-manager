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
              <el-tag type="success">在线</el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="主机名">pve-node-01</el-descriptions-item>
            <el-descriptions-item label="PVE 版本">8.1.4</el-descriptions-item>
            <el-descriptions-item label="运行时间">15 天 8 小时</el-descriptions-item>
            <el-descriptions-item label="内核版本">6.5.11</el-descriptions-item>
            <el-descriptions-item label="CPU 型号">Intel Xeon E5-2680</el-descriptions-item>
            <el-descriptions-item label="CPU 核心数">16 核 32 线程</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近任务</span>
              <el-link type="primary" :underline="false">查看全部</el-link>
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
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
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

const router = useRouter()
const { t } = useI18n()

// 资源使用数据（模拟）
const cpuUsage = ref(25)
const memoryUsage = ref(4.2)
const diskUsage = ref(120)
const networkIO = ref(256)

const memoryUsageText = computed(() => memoryUsage.value.toFixed(1))
const diskUsageText = computed(() => diskUsage.value.toFixed(0))
const networkIOText = computed(() => networkIO.value.toFixed(0))

// VM 状态汇总
const vmStatusItems = ref([
  { status: 'running', label: '运行中', count: 3, color: '#52c41a' },
  { status: 'stopped', label: '已停止', count: 1, color: '#8c8c8c' },
  { status: 'error', label: '错误', count: 0, color: '#f5222d' },
  { status: 'paused', label: '已暂停', count: 0, color: '#faad14' },
])

// 最近任务
const recentTasks = ref([
  { id: 1, name: '启动 VM web-server-01', time: '2 分钟前', statusType: 'success', statusText: '成功' },
  { id: 2, name: '快照 VM db-server-01', time: '15 分钟前', statusType: 'success', statusText: '成功' },
  { id: 3, name: '备份存储 local', time: '1 小时前', statusType: 'success', statusText: '成功' },
  { id: 4, name: '迁移 VM test-vm', time: '2 小时前', statusType: 'warning', statusText: '进行中' },
  { id: 5, name: '停止 VM old-server', time: '3 小时前', statusType: 'danger', statusText: '失败' },
])

/**
 * 刷新数据
 */
function handleRefresh() {
  ElMessage.success('数据刷新成功')
  // TODO: 实际调用 API 刷新数据
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
  ElMessage.info(`筛选状态: ${item.label}`)
  // TODO: 跳转到对应状态列表
}
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
