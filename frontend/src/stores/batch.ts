import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/** 批量操作状态 */
export type BatchItemStatus = 'pending' | 'processing' | 'success' | 'error'

/** 单个批量操作项 */
export interface BatchOperationItem {
  /** 资源 ID */
  id: string
  /** 资源名称 */
  name: string
  /** 资源类型 (vm/ct) */
  type: 'vm' | 'ct'
  /** 资源 VMID/CTID */
  vmid: number
  /** 当前操作状态 */
  status: BatchItemStatus
  /** 进度百分比 (0-100) */
  progress: number
  /** 错误信息 */
  error?: string
}

/** 批量操作类型 */
export type BatchActionType = 'start' | 'stop' | 'shutdown' | 'migrate' | 'backup' | 'reboot'

/** 批量操作执行结果 */
export interface BatchOperationResult {
  /** 操作类型 */
  action: BatchActionType
  /** 总项目数 */
  total: number
  /** 成功数 */
  successCount: number
  /** 失败数 */
  errorCount: number
}

/**
 * 批量操作状态管理
 *
 * 职责:
 * 1. 管理用户选中的虚拟机/容器列表
 * 2. 执行批量操作 (启动/停止/关机等)
 * 3. 追踪每个项目的操作进度
 * 4. 支持取消和重试失败项目
 */
export const useBatchStore = defineStore('batch', () => {
  // ============================================================
  // State
  // ============================================================

  /** 已选中的资源项列表 */
  const selectedItems = ref<Map<string, { id: string; name: string; type: 'vm' | 'ct'; vmid: number }>>(
    new Map(),
  )

  /** 当前批量操作进度列表 */
  const batchProgress = ref<BatchOperationItem[]>([])

  /** 是否正在执行批量操作 */
  const isBatchProcessing = ref(false)

  /** 遇到错误时是否继续执行 */
  const continueOnError = ref(true)

  /** 当前批量操作的 AbortController，用于取消操作 */
  let abortController: AbortController | null = null

  // ============================================================
  // Getters
  // ============================================================

  /** 选中项数组形式 (便于遍历) */
  const selectedItemsArray = computed(() => Array.from(selectedItems.value.values()))

  /** 选中项数量 */
  const selectedCount = computed(() => selectedItems.value.size)

  /** 是否有选中项 */
  const hasSelection = computed(() => selectedItems.value.size > 0)

  /** 整体进度百分比 */
  const overallProgress = computed(() => {
    if (batchProgress.value.length === 0) return 0
    const total = batchProgress.value.reduce((sum, item) => sum + item.progress, 0)
    return Math.round(total / batchProgress.value.length)
  })

  /** 成功数量 */
  const successCount = computed(
    () => batchProgress.value.filter((item) => item.status === 'success').length,
  )

  /** 失败数量 */
  const errorCount = computed(
    () => batchProgress.value.filter((item) => item.status === 'error').length,
  )

  /** 处理中数量 */
  const processingCount = computed(
    () => batchProgress.value.filter((item) => item.status === 'processing').length,
  )

  /** 是否全部完成 */
  const isComplete = computed(() => {
    if (batchProgress.value.length === 0) return false
    return batchProgress.value.every(
      (item) => item.status === 'success' || item.status === 'error',
    )
  })

  // ============================================================
  // Actions
  // ============================================================

  /**
   * 选中单个资源项
   * @param item 资源项信息
   */
  function selectItem(item: { id: string; name: string; type: 'vm' | 'ct'; vmid: number }): void {
    selectedItems.value.set(item.id, item)
  }

  /**
   * 取消选中单个资源项
   * @param id 资源 ID
   */
  function deselectItem(id: string): void {
    selectedItems.value.delete(id)
  }

  /**
   * 清空所有选中项
   */
  function clearSelection(): void {
    selectedItems.value.clear()
  }

  /**
   * 批量选中多个资源项
   * @param items 资源项列表
   */
  function selectMultiple(items: { id: string; name: string; type: 'vm' | 'ct'; vmid: number }[]): void {
    for (const item of items) {
      selectedItems.value.set(item.id, item)
    }
  }

  /**
   * 批量取消选中多个资源项
   * @param ids 资源 ID 列表
   */
  function deselectMultiple(ids: string[]): void {
    for (const id of ids) {
      selectedItems.value.delete(id)
    }
  }

  /**
   * 初始化批量操作进度追踪
   * @param action 操作类型
   */
  function initBatchProgress(action: BatchActionType): void {
    batchProgress.value = selectedItemsArray.value.map((item) => ({
      id: item.id,
      name: item.name,
      type: item.type,
      vmid: item.vmid,
      status: 'pending' as BatchItemStatus,
      progress: 0,
    }))
  }

  /**
   * 更新单个项目进度
   * @param id 资源 ID
   * @param status 状态
   * @param progress 进度
   * @param error 错误信息
   */
  function updateItemProgress(
    id: string,
    status: BatchItemStatus,
    progress: number,
    error?: string,
  ): void {
    const item = batchProgress.value.find((i) => i.id === id)
    if (item) {
      item.status = status
      item.progress = progress
      if (error) {
        item.error = error
      }
    }
  }

  /**
   * 取消当前批量操作
   */
  function cancelOperation(): void {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    isBatchProcessing.value = false

    // 将处理中的项目标记为错误
    for (const item of batchProgress.value) {
      if (item.status === 'processing') {
        item.status = 'error'
        item.error = '操作已取消'
      }
    }
  }

  /**
   * 重试失败的项目
   * @param action 操作类型
   * @param executeFn 执行函数
   */
  async function retryFailed(
    action: BatchActionType,
    executeFn: (item: { id: string; name: string; type: 'vm' | 'ct'; vmid: number }) => Promise<void>,
  ): Promise<void> {
    const failedItems = batchProgress.value.filter((item) => item.status === 'error')

    if (failedItems.length === 0) return

    // 重置失败项目的状态
    for (const item of failedItems) {
      item.status = 'pending'
      item.progress = 0
      item.error = undefined
    }

    isBatchProcessing.value = true
    abortController = new AbortController()

    for (const failedItem of failedItems) {
      if (abortController.signal.aborted) break

      // 如果遇到错误且不允许继续，则停止
      if (errorCount.value > 0 && !continueOnError.value) {
        break
      }

      try {
        updateItemProgress(failedItem.id, 'processing', 30)
        await executeFn(failedItem)
        updateItemProgress(failedItem.id, 'success', 100)
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : '未知错误'
        updateItemProgress(failedItem.id, 'error', 0, errorMsg)

        if (!continueOnError.value) {
          break
        }
      }
    }

    isBatchProcessing.value = false
    abortController = null
  }

  /**
   * 执行批量操作
   * @param action 操作类型
   * @param executeFn 执行单个项目的函数
   */
  async function executeBatch(
    action: BatchActionType,
    executeFn: (item: { id: string; name: string; type: 'vm' | 'ct'; vmid: number }) => Promise<void>,
  ): Promise<BatchOperationResult> {
    if (selectedItems.value.size === 0) {
      return { action, total: 0, successCount: 0, errorCount: 0 }
    }

    isBatchProcessing.value = true
    abortController = new AbortController()
    initBatchProgress(action)

    const items = selectedItemsArray.value

    for (const item of items) {
      // 检查是否被取消
      if (abortController.signal.aborted) {
        break
      }

      // 如果遇到错误且不允许继续，则停止
      if (errorCount.value > 0 && !continueOnError.value) {
        break
      }

      try {
        // 标记为处理中
        updateItemProgress(item.id, 'processing', 30)

        // 执行单个操作
        await executeFn(item)

        // 标记为成功
        updateItemProgress(item.id, 'success', 100)
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : '未知错误'
        updateItemProgress(item.id, 'error', 0, errorMsg)

        // 如果配置为遇到错误时停止，则终止循环
        if (!continueOnError.value) {
          break
        }
      }
    }

    isBatchProcessing.value = false
    abortController = null

    return {
      action,
      total: items.length,
      successCount: successCount.value,
      errorCount: errorCount.value,
    }
  }

  /**
   * 切换"遇到错误时是否继续"选项
   */
  function toggleContinueOnError(): void {
    continueOnError.value = !continueOnError.value
  }

  /**
   * 重置批量操作状态
   */
  function reset(): void {
    isBatchProcessing.value = false
    batchProgress.value = []
    if (abortController) {
      abortController.abort()
      abortController = null
    }
  }

  return {
    // State
    selectedItems,
    batchProgress,
    isBatchProcessing,
    continueOnError,
    // Getters
    selectedItemsArray,
    selectedCount,
    hasSelection,
    overallProgress,
    successCount,
    errorCount,
    processingCount,
    isComplete,
    // Actions
    selectItem,
    deselectItem,
    clearSelection,
    selectMultiple,
    deselectMultiple,
    cancelOperation,
    retryFailed,
    executeBatch,
    toggleContinueOnError,
    reset,
  }
})
