import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

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
  lastRunDuration: number | null // 秒
  lastRunSize: number | null // 字节
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
  duration: number // 秒
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
  const jobs = ref<BackupJob[]>(getMockBackupJobs())
  const history = ref<BackupHistory[]>(getMockBackupHistory())
  const storageTargets = ref<StorageTarget[]>(getMockStorageTargets())
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

  /** 创建备份任务 */
  async function createBackup(form: CreateBackupForm): Promise<boolean> {
    loading.value = true
    try {
      const target = findTarget(form.targetId)
      if (!target) {
        ElMessage.error('未找到目标资源')
        return false
      }

      const newJob: BackupJob = {
        id: `job-${Date.now()}`,
        name: `${target.name}-备份`,
        targetId: form.targetId!,
        targetName: target.name,
        targetType: target.type,
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

  /** 立即执行备份任务 */
  async function runBackup(jobId: string): Promise<boolean> {
    const job = jobs.value.find((j) => j.id === jobId)
    if (!job) {
      ElMessage.error('未找到备份任务')
      return false
    }

    loading.value = true
    try {
      // 模拟备份执行
      job.status = 'running'

      // 模拟异步执行过程
      await new Promise((resolve) => setTimeout(resolve, 2000))

      const success = Math.random() > 0.2 // 80% 成功率
      const size = Math.floor(Math.random() * 10_000_000_000) + 1_000_000_000
      const duration = Math.floor(Math.random() * 600) + 30

      job.status = success ? 'success' : 'failed'
      job.lastRunAt = new Date().toISOString()
      job.lastRunDuration = duration
      job.lastRunSize = size

      // 添加历史记录
      history.value.unshift({
        id: `history-${Date.now()}`,
        jobId: job.id,
        targetId: job.targetId,
        targetName: job.targetName,
        targetType: job.targetType,
        timestamp: new Date().toISOString(),
        size,
        status: job.status,
        duration,
        storageTarget: job.storageTarget,
        restoreAvailable: success,
      })

      ElMessage.success(success ? '备份执行成功' : '备份执行失败')
      return success
    } catch (error) {
      console.error('执行备份失败:', error)
      ElMessage.error('执行备份失败')
      job.status = 'failed'
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

  /** 从备份恢复 */
  async function restoreBackup(historyId: string): Promise<boolean> {
    loading.value = true
    try {
      const record = history.value.find((h) => h.id === historyId)
      if (!record || !record.restoreAvailable) {
        ElMessage.error('该备份不可用于恢复')
        return false
      }

      // 模拟恢复过程
      await new Promise((resolve) => setTimeout(resolve, 3000))

      ElMessage.success(`已恢复 ${record.targetName} 的备份`)
      return true
    } catch (error) {
      console.error('恢复备份失败:', error)
      ElMessage.error('恢复备份失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    jobs,
    history,
    storageTargets,
    loading,
    summary,
    createBackup,
    runBackup,
    deleteBackup,
    restoreBackup,
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

/** 查找目标资源 */
function findTarget(id: number | null): TargetResource | undefined {
  return mockTargets.find((t) => t.id === id)
}

// ============================================================
// Mock 数据
// ============================================================

const mockTargets: TargetResource[] = [
  { id: 100, name: 'web-server-01', type: 'vm', node: 'pve-node-01' },
  { id: 101, name: 'db-server-01', type: 'vm', node: 'pve-node-01' },
  { id: 102, name: 'test-vm-01', type: 'vm', node: 'pve-node-02' },
  { id: 103, name: 'api-gateway', type: 'vm', node: 'pve-node-02' },
  { id: 200, name: 'redis-cache', type: 'ct', node: 'pve-node-01' },
  { id: 201, name: 'nginx-proxy', type: 'ct', node: 'pve-node-02' },
]

function getMockStorageTargets(): StorageTarget[] {
  return [
    {
      id: 'storage-local',
      name: 'local',
      type: 'local',
      path: '/var/lib/vz/dump',
      totalBytes: 500_000_000_000,
      usedBytes: 120_000_000_000,
      active: true,
    },
    {
      id: 'storage-nfs-backup',
      name: 'nfs-backup',
      type: 'nfs',
      path: '/mnt/pve/nfs-backup',
      totalBytes: 2_000_000_000_000,
      usedBytes: 800_000_000_000,
      active: true,
    },
    {
      id: 'storage-pbs-01',
      name: 'pbs-01',
      type: 'pbs',
      path: 'pbs-01:8007:backup',
      totalBytes: 5_000_000_000_000,
      usedBytes: 1_500_000_000_000,
      active: true,
    },
  ]
}

function getMockBackupJobs(): BackupJob[] {
  const now = new Date()
  return [
    {
      id: 'job-001',
      name: 'web-server-01 每日备份',
      targetId: 100,
      targetName: 'web-server-01',
      targetType: 'vm',
      mode: 'snapshot',
      scheduleType: 'daily',
      cronExpression: '0 2 * * *',
      storageTarget: 'nfs-backup',
      compression: 'zstd',
      retentionCount: 7,
      notifyEmail: true,
      status: 'success',
      lastRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate(), 2, 15).toISOString(),
      nextRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1, 2, 0).toISOString(),
      lastRunDuration: 245,
      lastRunSize: 4_500_000_000,
    },
    {
      id: 'job-002',
      name: 'db-server-01 每日备份',
      targetId: 101,
      targetName: 'db-server-01',
      targetType: 'vm',
      mode: 'snapshot',
      scheduleType: 'daily',
      cronExpression: '0 1 * * *',
      storageTarget: 'pbs-01',
      compression: 'zstd',
      retentionCount: 14,
      notifyEmail: true,
      status: 'success',
      lastRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate(), 1, 30).toISOString(),
      nextRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1, 1, 0).toISOString(),
      lastRunDuration: 520,
      lastRunSize: 12_800_000_000,
    },
    {
      id: 'job-003',
      name: 'redis-cache 每周备份',
      targetId: 200,
      targetName: 'redis-cache',
      targetType: 'ct',
      mode: 'snapshot',
      scheduleType: 'weekly',
      cronExpression: '0 3 * * 0',
      storageTarget: 'local',
      compression: 'gzip',
      retentionCount: 4,
      notifyEmail: false,
      status: 'failed',
      lastRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate() - 7, 3, 5).toISOString(),
      nextRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate() + (7 - now.getDay()), 3, 0).toISOString(),
      lastRunDuration: null,
      lastRunSize: null,
    },
    {
      id: 'job-004',
      name: 'api-gateway 一次性备份',
      targetId: 103,
      targetName: 'api-gateway',
      targetType: 'vm',
      mode: 'stop',
      scheduleType: 'once',
      cronExpression: '',
      storageTarget: 'nfs-backup',
      compression: 'none',
      retentionCount: 3,
      notifyEmail: true,
      status: 'pending',
      lastRunAt: null,
      nextRunAt: null,
      lastRunDuration: null,
      lastRunSize: null,
    },
    {
      id: 'job-005',
      name: 'nginx-proxy 自定义调度',
      targetId: 201,
      targetName: 'nginx-proxy',
      targetType: 'ct',
      mode: 'suspend',
      scheduleType: 'custom',
      cronExpression: '0 */6 * * *',
      storageTarget: 'pbs-01',
      compression: 'lz4',
      retentionCount: 10,
      notifyEmail: false,
      status: 'success',
      lastRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate(), 6, 10).toISOString(),
      nextRunAt: new Date(now.getFullYear(), now.getMonth(), now.getDate(), 12, 0).toISOString(),
      lastRunDuration: 85,
      lastRunSize: 890_000_000,
    },
  ]
}

function getMockBackupHistory(): BackupHistory[] {
  const now = new Date()
  const today = now.toISOString().slice(0, 10)

  return [
    {
      id: 'history-001',
      jobId: 'job-002',
      targetId: 101,
      targetName: 'db-server-01',
      targetType: 'vm',
      timestamp: `${today}T01:30:00Z`,
      size: 12_800_000_000,
      status: 'success',
      duration: 520,
      storageTarget: 'pbs-01',
      restoreAvailable: true,
    },
    {
      id: 'history-002',
      jobId: 'job-001',
      targetId: 100,
      targetName: 'web-server-01',
      targetType: 'vm',
      timestamp: `${today}T02:15:00Z`,
      size: 4_500_000_000,
      status: 'success',
      duration: 245,
      storageTarget: 'nfs-backup',
      restoreAvailable: true,
    },
    {
      id: 'history-003',
      jobId: 'job-005',
      targetId: 201,
      targetName: 'nginx-proxy',
      targetType: 'ct',
      timestamp: `${today}T06:10:00Z`,
      size: 890_000_000,
      status: 'success',
      duration: 85,
      storageTarget: 'pbs-01',
      restoreAvailable: true,
    },
    {
      id: 'history-004',
      jobId: 'job-003',
      targetId: 200,
      targetName: 'redis-cache',
      targetType: 'ct',
      timestamp: new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - 7,
        3,
        5,
      ).toISOString(),
      size: 0,
      status: 'failed',
      duration: 0,
      storageTarget: 'local',
      restoreAvailable: false,
    },
    {
      id: 'history-005',
      jobId: 'job-001',
      targetId: 100,
      targetName: 'web-server-01',
      targetType: 'vm',
      timestamp: new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - 1,
        2,
        10,
      ).toISOString(),
      size: 4_200_000_000,
      status: 'success',
      duration: 230,
      storageTarget: 'nfs-backup',
      restoreAvailable: true,
    },
    {
      id: 'history-006',
      jobId: 'job-002',
      targetId: 101,
      targetName: 'db-server-01',
      targetType: 'vm',
      timestamp: new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - 1,
        1,
        25,
      ).toISOString(),
      size: 12_500_000_000,
      status: 'success',
      duration: 510,
      storageTarget: 'pbs-01',
      restoreAvailable: true,
    },
  ]
}
