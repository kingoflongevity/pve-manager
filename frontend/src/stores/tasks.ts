/**
 * 任务状态存储
 * 管理 PVE 任务列表、轮询刷新、状态筛选等
 */
import { defineStore } from 'pinia'
import { ref, computed, shallowRef } from 'vue'
import { fetchTasks } from '@/api/tasks'
import type { Task, TaskStatusFilter } from '@/api/taskTypes'

/** 轮询间隔（毫秒） */
const POLL_INTERVAL = 5000

export const useTaskStore = defineStore('tasks', () => {
  // ===== State =====

  /** 全部任务列表 - 使用 shallowRef 优化大型数据集性能 */
  const tasks = shallowRef<Task[]>([])

  /** 加载状态 */
  const loading = ref(false)

  /** 最后一次刷新时间 */
  const lastRefresh = ref<number>(0)

  /** 轮询定时器 ID */
  let pollTimer: ReturnType<typeof setInterval> | null = null

  /** 是否启用自动轮询 */
  const polling = ref(false)

  // ===== Getters =====

  /** 运行中的任务 */
  const runningTasks = computed(() =>
    tasks.value.filter((t) => t.status === 'running'),
  )

  /** 运行中任务数量 */
  const taskCount = computed(() => runningTasks.value.length)

  /** 已完成的任务 */
  const completedTasks = computed(() =>
    tasks.value.filter((t) => t.status !== 'running'),
  )

  /** 根据筛选条件过滤任务 */
  function getTasksByFilter(filter: TaskStatusFilter): Task[] {
    if (filter === 'all') return tasks.value
    if (filter === 'running') return runningTasks.value
    return tasks.value.filter((t) => t.status === filter)
  }

  /** 是否存在正在运行的任务 */
  const hasRunningTasks = computed(() => runningTasks.value.length > 0)

  // ===== Actions =====

  /**
   * 拉取最新任务列表
   */
  async function refreshTasks() {
    loading.value = true
    try {
      const res = await fetchTasks()
      if (res.data) {
        tasks.value = res.data.map(normalizeTask)
      }
      lastRefresh.value = Date.now()
    } catch (error) {
      console.error('刷新任务列表失败:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 清除已完成的任务
   */
  function clearCompleted() {
    tasks.value = tasks.value.filter((t) => t.status === 'running')
  }

  /**
   * 清除所有任务
   */
  function clearAll() {
    tasks.value = []
    stopPolling()
  }

  /**
   * 开启自动轮询
   */
  function startPolling() {
    if (pollTimer) return
    polling.value = true
    // 立即拉取一次
    refreshTasks()
    pollTimer = setInterval(() => {
      // 仅当有运行中任务时继续轮询
      if (runningTasks.value.length === 0) {
        stopPolling()
      } else {
        refreshTasks()
      }
    }, POLL_INTERVAL)
  }

  /**
   * 停止自动轮询
   */
  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    polling.value = false
  }

  /**
   * 组件卸载时清理
   */
  function $reset() {
    stopPolling()
    tasks.value = []
  }

  return {
    tasks,
    loading,
    lastRefresh,
    polling,
    runningTasks,
    taskCount,
    completedTasks,
    hasRunningTasks,
    getTasksByFilter,
    refreshTasks,
    clearCompleted,
    clearAll,
    startPolling,
    stopPolling,
    $reset,
  }
})

/**
 * 将后端返回的原始任务数据标准化为前端使用的格式
 */
function normalizeTask(raw: Record<string, unknown>): Task {
  const upid = (raw.upid as string) || ''
  const status = normalizeStatus(raw.status as string, raw.exitstatus as string)
  const progress = calcProgress(status, raw as Record<string, unknown>)

  return {
    upid,
    id: upid.split(':').pop() || upid,
    node: (raw.node as string) || extractNodeFromUpid(upid),
    type: (raw.type as Task['type']) || 'unknown',
    vmid: Number(raw.vmid) || undefined,
    description: generateDescription(raw),
    status,
    progress,
    starttime: Number(raw.starttime) || 0,
    endtime: Number(raw.endtime) || 0,
    exitstatus: raw.exitstatus as string | undefined,
    user: raw.user as string | undefined,
  }
}

/**
 * 从 UPID 字符串中提取节点名称
 * UPID 格式: UPID:node:pid:starttime:type:vmid:action:user
 */
function extractNodeFromUpid(upid: string): string {
  const parts = upid.split(':')
  return parts.length > 1 ? parts[1] : ''
}

/**
 * 标准化任务状态
 */
function normalizeStatus(
  rawStatus: string | undefined,
  exitStatus: string | undefined,
): Task['status'] {
  if (rawStatus === 'running') return 'running'
  if (exitStatus === 'OK') return 'success'
  if (exitStatus && exitStatus !== 'OK') return 'error'
  if (rawStatus === 'stopped') return 'stopped'
  // 如果有 endtime 但没有明确状态，根据 exitstatus 判断
  if (exitStatus) return exitStatus === 'OK' ? 'success' : 'error'
  return 'running'
}

/**
 * 计算任务进度
 */
function calcProgress(
  status: Task['status'],
  raw: Record<string, unknown>,
): number {
  if (status === 'success') return 100
  if (status === 'error' || status === 'stopped') {
    return Number(raw.progress) || 0
  }
  // 运行中任务：如果有进度则用之，否则显示 50% 占位
  return Number(raw.progress) || 50
}

/**
 * 根据任务类型和参数生成可读的描述文本
 */
function generateDescription(raw: Record<string, unknown>): string {
  const type = (raw.type as string) || ''
  const vmid = raw.vmid ? ` (VM ${raw.vmid})` : ''
  const node = raw.node || 'unknown'

  const typeMap: Record<string, string> = {
    qmstart: '启动虚拟机',
    qmstop: '停止虚拟机',
    qmshutdown: '关闭虚拟机',
    qmreset: '重启虚拟机',
    qmreboot: '重启虚拟机',
    qmresume: '恢复虚拟机',
    qmmigrate: '迁移虚拟机',
    qmbackup: '备份虚拟机',
    vzstart: '启动容器',
    vzstop: '停止容器',
    vzshutdown: '关闭容器',
    vzresume: '恢复容器',
    vzmigrate: '迁移容器',
    vzbackup: '备份容器',
    vzrestore: '恢复容器',
    aptupdate: '更新软件包',
    upgrade: '系统升级',
  }

  const label = typeMap[type] || type || '未知任务'
  return `${label}${vmid} [${node}]`
}
