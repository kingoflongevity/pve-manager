<script setup lang="ts">
import { ref, computed } from 'vue'
import { useBackupStore, type BackupJobStatus } from '@/stores/backup'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder,
  CircleCheck,
  CircleClose,
  Loading,
  Clock,
  RefreshRight,
  Delete,
  View,
  Plus,
  Clock as ClockIcon,
  Monitor,
  Box,
} from '@element-plus/icons-vue'
import CreateBackupDialog from '@/components/backup/CreateBackupDialog.vue'
import { formatBytes, formatDuration, formatDateTime } from '@/utils/format'

const backupStore = useBackupStore()

// ============================================================
// 状态管理
// ============================================================

/** 状态筛选 */
const statusFilter = ref<BackupJobStatus | 'all'>('all')

/** 创建备份对话框可见性 */
const createDialogVisible = ref(false)

// ============================================================
// 计算属性
// ============================================================

/** 筛选后的任务列表 */
const filteredJobs = computed(() => {
  if (statusFilter.value === 'all') {
    return backupStore.jobs
  }
  return backupStore.jobs.filter((job) => job.status === statusFilter.value)
})

/** 状态选项 */
const statusOptions = computed(() => [
  { label: '全部', value: 'all' as const, count: backupStore.jobs.length },
  {
    label: '成功',
    value: 'success' as const,
    count: backupStore.jobs.filter((j) => j.status === 'success').length,
  },
  {
    label: '失败',
    value: 'failed' as const,
    count: backupStore.jobs.filter((j) => j.status === 'failed').length,
  },
  {
    label: '运行中',
    value: 'running' as const,
    count: backupStore.jobs.filter((j) => j.status === 'running').length,
  },
])

/** 备份模式映射 */
const modeMap: Record<string, string> = {
  snapshot: '快照',
  stop: '停机',
  suspend: '挂起',
}

/** 调度类型映射 */
const scheduleMap: Record<string, string> = {
  once: '一次性',
  daily: '每天',
  weekly: '每周',
  monthly: '每月',
  custom: '自定义',
}

// ============================================================
// 操作方法
// ============================================================

/** 立即执行备份 */
async function handleRunNow(jobId: string) {
  await backupStore.runBackup(jobId)
}

/** 编辑备份任务（简单提示，后续可扩展完整编辑功能） */
function handleEdit(_jobId: string) {
  ElMessage.info('编辑功能开发中')
}

/** 删除备份任务 */
async function handleDelete(jobId: string) {
  try {
    await ElMessageBox.confirm('确认删除该备份任务？此操作不可恢复', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await backupStore.deleteBackup(jobId)
  } catch {
    // 用户取消
  }
}

/** 查看备份历史 */
function handleHistory(_jobId: string) {
  // 滚动到历史记录区域
  document.getElementById('backup-history')?.scrollIntoView({ behavior: 'smooth' })
}

/** 恢复备份 */
async function handleRestore(historyId: string) {
  await backupStore.restoreBackup(historyId)
}

/** 创建备份对话框已提交 */
function handleCreateSubmitted() {
  // 刷新数据（当前使用 mock 数据，无需额外操作）
}

/** 打开创建对话框 */
function handleCreateBackup() {
  createDialogVisible.value = true
}

// ============================================================
// 工具函数
// ============================================================

/** 获取状态标签类型 */
function getStatusType(status: BackupJobStatus): string {
  const map: Record<BackupJobStatus, string> = {
    success: 'success',
    failed: 'danger',
    running: 'primary',
    pending: 'info',
  }
  return map[status]
}

/** 获取状态显示文字 */
function getStatusLabel(status: BackupJobStatus): string {
  const map: Record<BackupJobStatus, string> = {
    success: '成功',
    failed: '失败',
    running: '运行中',
    pending: '等待中',
  }
  return map[status]
}

/** 获取时间线颜色 */
function getTimelineColor(status: BackupJobStatus): string {
  const map: Record<BackupJobStatus, string> = {
    success: '#52c41a',
    failed: '#f5222d',
    running: '#1677ff',
    pending: '#bfbfbf',
  }
  return map[status]
}
</script>

<template>
  <div class="page-container backup-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">备份管理</h1>
        <p class="page-description">集中管理所有备份任务，监控执行状态，快速恢复数据</p>
      </div>
      <el-button type="primary" :icon="Plus" size="large" @click="handleCreateBackup">
        创建备份
      </el-button>
    </div>

    <!-- 摘要栏 -->
    <el-row :gutter="16" class="summary-row">
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-icon primary">
              <el-icon :size="24"><Folder /></el-icon>
            </div>
            <div class="summary-text">
              <div class="summary-label">备份任务</div>
              <div class="summary-value">{{ backupStore.summary.totalJobs }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-icon success">
              <el-icon :size="24"><CircleCheck /></el-icon>
            </div>
            <div class="summary-text">
              <div class="summary-label">今日成功</div>
              <div class="summary-value success">{{ backupStore.summary.successToday }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-icon danger">
              <el-icon :size="24"><CircleClose /></el-icon>
            </div>
            <div class="summary-text">
              <div class="summary-label">今日失败</div>
              <div class="summary-value danger">{{ backupStore.summary.failedToday }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-icon info">
              <el-icon :size="24"><Folder /></el-icon>
            </div>
            <div class="summary-text">
              <div class="summary-label">存储使用</div>
              <div class="summary-value">{{ formatBytes(backupStore.summary.totalStorageUsed) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 备份任务表格 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">备份任务</span>
          <div class="header-actions">
            <!-- 状态筛选 -->
            <el-segmented v-model="statusFilter" :options="statusOptions as any">
              <template #default="{ item }: { item: { label: string; value: string; count: number } }">
                <span>{{ item.label }}</span>
                <el-badge
                  :value="item.count"
                  :max="99"
                  :type="item.value === 'all' ? 'info' : item.value === 'success' ? 'success' : item.value === 'failed' ? 'danger' : 'primary'"
                  class="count-badge"
                />
              </template>
            </el-segmented>
          </div>
        </div>
      </template>

      <el-table
        :data="filteredJobs"
        v-loading="backupStore.loading"
        stripe
        style="width: 100%"
        row-key="id"
      >
        <!-- 任务名称 -->
        <el-table-column label="任务名称" min-width="180">
          <template #default="{ row }">
            <div class="job-name-cell">
              <el-icon class="target-icon">
                <Monitor v-if="row.targetType === 'vm'" />
                <Box v-else />
              </el-icon>
              <div class="job-info">
                <span class="job-name">{{ row.name }}</span>
                <span class="job-target">{{ row.targetName }} ({{ row.targetType === 'vm' ? 'VM' : 'CT' }} {{ row.targetId }})</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- 备份模式 -->
        <el-table-column label="模式" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info" effect="plain">
              {{ modeMap[row.mode] }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 调度策略 -->
        <el-table-column label="调度" width="120">
          <template #default="{ row }">
            <div class="schedule-cell">
              <span class="schedule-type">{{ scheduleMap[row.scheduleType] }}</span>
              <span v-if="row.cronExpression" class="cron-expr">{{ row.cronExpression }}</span>
            </div>
          </template>
        </el-table-column>

        <!-- 存储目标 -->
        <el-table-column label="存储目标" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.storageTarget }}</el-tag>
          </template>
        </el-table-column>

        <!-- 上次运行 -->
        <el-table-column label="上次运行" width="140">
          <template #default="{ row }">
            <div v-if="row.lastRunAt" class="last-run-cell">
              <span class="run-time">{{ formatDateTime(row.lastRunAt) }}</span>
              <span v-if="row.lastRunDuration" class="run-duration">{{ formatDuration(row.lastRunDuration) }}</span>
            </div>
            <span v-else class="no-data">--</span>
          </template>
        </el-table-column>

        <!-- 下次运行 -->
        <el-table-column label="下次运行" width="140">
          <template #default="{ row }">
            <div v-if="row.nextRunAt" class="next-run-cell">
              <el-icon :size="14"><ClockIcon /></el-icon>
              <span>{{ formatDateTime(row.nextRunAt) }}</span>
            </div>
            <span v-else class="no-data">--</span>
          </template>
        </el-table-column>

        <!-- 状态 -->
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" effect="light" size="small">
              <el-icon :size="12" class="status-icon">
                <CircleCheck v-if="row.status === 'success'" />
                <CircleClose v-else-if="row.status === 'failed'" />
                <Loading v-else-if="row.status === 'running'" />
                <Clock v-else />
              </el-icon>
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 操作 -->
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                link
                size="small"
                :icon="RefreshRight"
                :disabled="row.status === 'running'"
                @click="handleRunNow(row.id)"
              >
                立即执行
              </el-button>
              <el-button type="primary" link size="small" @click="handleEdit(row.id)">
                编辑
              </el-button>
              <el-button type="primary" link size="small" :icon="View" @click="handleHistory(row.id)">
                历史
              </el-button>
              <el-button type="danger" link size="small" :icon="Delete" @click="handleDelete(row.id)">
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 备份历史记录 -->
    <el-card id="backup-history" class="history-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">
            <el-icon :size="18"><Clock /></el-icon>
            备份历史
          </span>
        </div>
      </template>

      <div v-if="backupStore.history.length === 0" class="empty-history">
        <el-empty description="暂无备份历史" />
      </div>

      <div v-else class="timeline-wrapper">
        <el-timeline>
          <el-timeline-item
            v-for="record in backupStore.history"
            :key="record.id"
            :color="getTimelineColor(record.status)"
            :timestamp="formatDateTime(record.timestamp)"
            placement="top"
          >
            <div class="history-item">
              <div class="history-main">
                <div class="history-header">
                  <el-icon class="target-icon">
                    <Monitor v-if="record.targetType === 'vm'" />
                    <Box v-else />
                  </el-icon>
                  <span class="history-target">{{ record.targetName }}</span>
                  <el-tag
                    :type="getStatusType(record.status)"
                    size="small"
                    effect="plain"
                    class="history-status"
                  >
                    {{ getStatusLabel(record.status) }}
                  </el-tag>
                </div>
                <div class="history-details">
                  <span>大小: {{ formatBytes(record.size) }}</span>
                  <span v-if="record.duration > 0">耗时: {{ formatDuration(record.duration) }}</span>
                  <span>存储: {{ record.storageTarget }}</span>
                </div>
              </div>
              <div class="history-actions">
                <el-button
                  v-if="record.restoreAvailable"
                  type="primary"
                  size="small"
                  plain
                  @click="handleRestore(record.id)"
                >
                  恢复
                </el-button>
              </div>
            </div>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-card>

    <!-- 创建备份对话框 -->
    <CreateBackupDialog
      v-model:visible="createDialogVisible"
      @submitted="handleCreateSubmitted"
    />
  </div>
</template>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

// ============================================================
// 摘要卡片
// ============================================================

.summary-row {
  margin-bottom: $spacing-6;
}

.summary-card {
  border-radius: $radius-base;

  :deep(.el-card__body) {
    padding: $spacing-4;
  }
}

.summary-content {
  display: flex;
  align-items: center;
  gap: $spacing-4;
}

.summary-icon {
  width: 48px;
  height: 48px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;

  &.primary {
    background: $primary-bg;
    color: $primary-color;
  }

  &.success {
    background: $success-bg;
    color: $success-color;
  }

  &.danger {
    background: $danger-bg;
    color: $danger-color;
  }

  &.info {
    background: $info-bg;
    color: $info-color;
  }
}

.summary-text {
  flex: 1;

  .summary-label {
    font-size: $font-size-sm;
    color: $color-text-secondary;
    margin-bottom: $spacing-1;
  }

  .summary-value {
    font-size: $font-size-2xl;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;

    &.success {
      color: $success-color;
    }

    &.danger {
      color: $danger-color;
    }
  }
}

// ============================================================
// 表格卡片头部
// ============================================================

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .card-title {
    font-size: $font-size-lg;
    font-weight: $font-weight-semibold;
    display: flex;
    align-items: center;
    gap: $spacing-2;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: $spacing-4;
}

.count-badge {
  margin-left: $spacing-1;
}

// ============================================================
// 表格单元格
// ============================================================

.job-name-cell {
  display: flex;
  align-items: center;
  gap: $spacing-3;

  .target-icon {
    font-size: 18px;
    color: $primary-color;
    flex-shrink: 0;
  }

  .job-info {
    display: flex;
    flex-direction: column;
    gap: $spacing-1;

    .job-name {
      font-weight: $font-weight-medium;
      color: $color-text-primary;
    }

    .job-target {
      font-size: $font-size-xs;
      color: $color-text-secondary;
    }
  }
}

.schedule-cell {
  display: flex;
  flex-direction: column;
  gap: $spacing-1;

  .schedule-type {
    font-weight: $font-weight-medium;
    font-size: $font-size-sm;
  }

  .cron-expr {
    font-family: $font-family-code;
    font-size: $font-size-xs;
    color: $color-text-secondary;
    background: $color-bg-elevated;
    padding: $spacing-1 $spacing-2;
    border-radius: $radius-xs;
  }
}

.last-run-cell,
.next-run-cell {
  display: flex;
  flex-direction: column;
  gap: $spacing-1;
}

.run-time {
  font-size: $font-size-sm;
}

.run-duration {
  font-size: $font-size-xs;
  color: $color-text-secondary;
}

.no-data {
  color: $color-text-placeholder;
}

.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-1;
  flex-wrap: wrap;
}

// ============================================================
// 状态标签
// ============================================================

.status-icon {
  margin-right: $spacing-1;

  &.is-loading {
    animation: spin 1s linear infinite;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// ============================================================
// 历史记录
// ============================================================

.history-card {
  margin-top: $spacing-6;
}

.empty-history {
  padding: $spacing-8 0;
}

.timeline-wrapper {
  padding: $spacing-2 0;
}

.history-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  background: $color-bg-hover;
  border-radius: $radius-sm;
  padding: $spacing-4;
  gap: $spacing-4;
  transition: $transition-fast;

  &:hover {
    background: $primary-1;
  }
}

.history-main {
  flex: 1;
  min-width: 0;
}

.history-header {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  margin-bottom: $spacing-2;

  .target-icon {
    font-size: 16px;
    color: $primary-color;
  }

  .history-target {
    font-weight: $font-weight-medium;
  }

  .history-status {
    margin-left: auto;
  }
}

.history-details {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-4;
  font-size: $font-size-sm;
  color: $color-text-secondary;
}

.history-actions {
  flex-shrink: 0;
}
</style>
