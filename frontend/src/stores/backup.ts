import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchStorageList, getStorageContent } from '@/api/storage'
import type { Storage, StorageContent } from '@/api/types'

// ============================================================
// 类型定义
// ============================================================

/** 备份任务状态 */
export type BackupJobStatus = 'success' | 'failed' | 'running' | 'pending'

/** 备份模式 */
export type BackupMode = 'snapshot' | 'stop' | 'suspend'

/** 压缩算法 */
export type CompressionType = 'none' | 'gzip' | 'lz4' | 'zstd'

/** 调度策略 */
export type ScheduleType = 'once' | 'daily' | 'weekly' | 'monthly' | 'custom'

/** 目标资源类型 */
export type TargetType = 'vm' | 'ct'

/** 存储目标 */
export interface StorageTarget {
  id: string
  name: string
  type: 'local' | 'nfs' | 'pbs' | 'smb' | 'cifs'
  path: string
  totalBytes: number
  usedBytes: number
  active: boolean
}

/** 目标资源（VM 或 CT） */
export interface TargetResource {
  id: number
  name: string
  type: TargetType
  node: string
}

/** 备份任务配置 */
export interface BackupJob {
  id: string
  name: string
  targetId: number
  targetName: string
  targetType: TargetType
  mode: BackupMode
  scheduleType: ScheduleType
  cronExpression: string
  storageTarget: string
  compression: CompressionType
  retentionCount: number
  notifyEmail: boolean
  status: BackupJobStatus
  lastRunAt: string | null
  nextRunAt: string | null
  lastRunDuration: number | null
  lastRunSize: number | null
}

/** 备份历史记录 */
export interface BackupHistory {
  id: string
  jobId: string
  targetId: number
  targetName: string
  targetType: TargetType
  timestamp: string
  size: number
  status: BackupJobStatus
  duration: number
  storageTarget: string
  restoreAvailable: boolean
}

/** 备份摘要统计 */
export interface BackupSummary {
  totalJobs: number
  successToday: number
  failedToday: number
  totalStorageUsed: number
}

/** 创建备份任务表单数据 */
export interface CreateBackupForm {
  targetId: number | null
  mode: BackupMode
  storageTarget: string
  compression: CompressionType
  scheduleType: ScheduleType
  cronExpression: string
  retentionCount: number
  notifyEmail: boolean
}

// ============================================================
// Store 定义
// ============================================================

export const useBackupStore = defineStore('backup', () => {
  // State
  const jobs = ref<BackupJob[]>([])
  const history = ref<BackupHistory[]>([])
  const storageTargets = ref<Storage[]>([])
  const backupFiles = ref<StorageContent[]>([])
  const loading = ref(false)

  // Getters
  const summary = computed<BackupSummary>(() => {
    const today = new Date().toISOString().slice(0, 10)
    const todayHistory = history.value.filter((h) =>
      h.timestamp.startsWith(today),
    )

    return {
      totalJobs: jobs.value.length,
      successToday: todayHistory.filter((h) => h.status === 'success').length,
      failedToday: todayHistory.filter((h) => h.status === 'failed').length,
      totalStorageUsed: history.value
        .filter((h) => h.status === 'success')
        .reduce((sum, h) => sum + h.size, 0),
    }
  })

  // Actions

  /**
   * 从后端加载存储列表（可用于备份目标选择）
   */
  async function loadStorageTargets(node: string): Promise<void> {
    loading.value = true
    try {
      const storages = await fetchStorageList(node)
      storageTargets.value = storages.filter((s) =>
        s.content.includes('backup'),
      )
    } catch (error) {
      console.error('加载存储列表失败:', error)
      ElMessage.error('加载存储列表失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 加载备份文件列表
   */
  async function loadBackupFiles(node: string, storage: string): Promise<void> {
    loading.value = true
    try {
      const files = await getStorageContent(node, storage, { content: 'backup' })
      backupFiles.value = files
    } catch (error) {
      console.error('加载备份文件失败:', error)
      ElMessage.error('加载备份文件失败')
    } finally {
      loading.value = false
    }
  }

  /** 创建备份任务 */
  async function createBackup(form: CreateBackupForm): Promise<boolean> {
    loading.value = true
    try {
      const newJob: BackupJob = {
        id: `job-${Date.now()}`,
        name: `备份任务-${form.targetId}`,
        targetId: form.targetId!,
        targetName: `VM/CT ${form.targetId}`,
        targetType: 'vm',
        mode: form.mode,
        scheduleType: form.scheduleType,
        cronExpression: form.scheduleType === 'custom' ? form.cronExpression : getCronFromSchedule(form.scheduleType),
        storageTarget: form.storageTarget,
        compression: form.compression,
        retentionCount: form.retentionCount,
        notifyEmail: form.notifyEmail,
        status: 'pending' as BackupJobStatus,
        lastRunAt: null,
        nextRunAt: getNextRunTime(form.scheduleType).toISOString(),
        lastRunDuration: null,
        lastRunSize: null,
      }

      jobs.value.push(newJob)
      ElMessage.success('备份任务创建成功')
      return true
    } catch (error) {
      console.error('创建备份任务失败:', error)
      ElMessage.error('创建备份任务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /** 删除备份任务 */
  async function deleteBackup(jobId: string): Promise<boolean> {
    loading.value = true
    try {
      const index = jobs.value.findIndex((j) => j.id === jobId)
      if (index === -1) {
        ElMessage.error('未找到备份任务')
        return false
      }

      jobs.value.splice(index, 1)
      ElMessage.success('备份任务已删除')
      return true
    } catch (error) {
      console.error('删除备份任务失败:', error)
      ElMessage.error('删除备份任务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /** 添加备份历史记录 */
  function addHistoryRecord(record: BackupHistory): void {
    history.value.unshift(record)
  }

  /** 清空历史记录 */
  function clearHistory(): void {
    history.value = []
  }

  return {
    jobs,
    history,
    storageTargets,
    backupFiles,
    loading,
    summary,
    loadStorageTargets,
    loadBackupFiles,
    createBackup,
    deleteBackup,
    addHistoryRecord,
    clearHistory,
  }
})

// ============================================================
// 内部工具函数
// ============================================================

/** 根据调度类型生成 cron 表达式 */
function getCronFromSchedule(type: ScheduleType): string {
  switch (type) {
    case 'daily':
      return '0 2 * * *'
    case 'weekly':
      return '0 2 * * 0'
    case 'monthly':
      return '0 2 1 * *'
    case 'once':
    case 'custom':
    default:
      return ''
  }
}

/** 根据调度类型估算下次运行时间 */
function getNextRunTime(type: ScheduleType): Date {
  const now = new Date()
  switch (type) {
    case 'daily':
      now.setDate(now.getDate() + 1)
      now.setHours(2, 0, 0, 0)
      return now
    case 'weekly':
      now.setDate(now.getDate() + (7 - now.getDay()))
      now.setHours(2, 0, 0, 0)
      return now
    case 'monthly':
      now.setMonth(now.getMonth() + 1)
      now.setDate(1)
      now.setHours(2, 0, 0, 0)
      return now
    case 'once':
      return now
    default:
      return now
  }
}
