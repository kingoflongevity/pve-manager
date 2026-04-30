<template>
  <!-- 批量操作栏 - 当有选中项时显示 -->
  <Transition name="batch-bar">
    <div v-if="hasSelection" class="batch-operation-bar">
      <div class="batch-bar-content">
        <!-- 左侧：选中计数 -->
        <div class="batch-bar-left">
          <el-checkbox
            :model-value="true"
            disabled
            class="batch-checkbox"
          />
          <span class="batch-count">
            已选择 <strong>{{ selectedCount }}</strong> 项
          </span>
          <el-button
            link
            type="primary"
            size="small"
            @click="handleClearSelection"
          >
            取消全选
          </el-button>
        </div>

        <!-- 中间：操作按钮 -->
        <div class="batch-bar-center">
          <el-tooltip content="全部启动" placement="top">
            <el-button
              type="success"
              size="small"
              :disabled="isBatchProcessing"
              @click="handleAction('start')"
            >
              <el-icon><VideoPlay /></el-icon>
              全部启动
            </el-button>
          </el-tooltip>

          <el-tooltip content="全部停止" placement="top">
            <el-button
              type="danger"
              size="small"
              :disabled="isBatchProcessing"
              @click="handleAction('stop')"
            >
              <el-icon><VideoPause /></el-icon>
              全部停止
            </el-button>
          </el-tooltip>

          <el-tooltip content="全部关机" placement="top">
            <el-button
              type="warning"
              size="small"
              :disabled="isBatchProcessing"
              @click="handleAction('shutdown')"
            >
              <el-icon><SwitchButton /></el-icon>
              全部关机
            </el-button>
          </el-tooltip>

          <el-tooltip content="批量迁移" placement="top">
            <el-button
              type="primary"
              size="small"
              :disabled="isBatchProcessing"
              @click="handleAction('migrate')"
            >
              <el-icon><Position /></el-icon>
              迁移
            </el-button>
          </el-tooltip>

          <el-tooltip content="批量备份" placement="top">
            <el-button
              type="info"
              size="small"
              :disabled="isBatchProcessing"
              @click="handleAction('backup')"
            >
              <el-icon><FolderOpened /></el-icon>
              备份
            </el-button>
          </el-tooltip>
        </div>

        <!-- 右侧：关闭按钮 -->
        <div class="batch-bar-right">
          <el-button
            link
            type="info"
            size="small"
            @click="handleClearSelection"
          >
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>

      <!-- 进度指示器 -->
      <div v-if="isBatchProcessing" class="batch-progress-indicator">
        <el-progress
          :percentage="overallProgress"
          :stroke-width="3"
          :show-text="false"
          class="batch-progress-bar"
        />
        <span class="batch-progress-text">
          处理中: {{ processingCount }} / {{ selectedCount }}
        </span>
      </div>
    </div>
  </Transition>

  <!-- 批量操作进度弹窗 -->
  <BatchProgress
    v-model:visible="progressVisible"
    :action="currentAction"
  />
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  VideoPlay,
  VideoPause,
  SwitchButton,
  Position,
  FolderOpened,
  Close,
} from '@element-plus/icons-vue'
import { useBatchStore, type BatchActionType } from '@/stores/batch'
import BatchProgress from './BatchProgress.vue'

const batchStore = useBatchStore()

// ============================================================
// 状态
// ============================================================

/** 是否显示进度弹窗 */
const progressVisible = ref(false)

/** 当前执行的操作类型 */
const currentAction = ref<BatchActionType>('start')

// ============================================================
// 计算属性 (从 store 获取)
// ============================================================

const hasSelection = computed(() => batchStore.hasSelection)
const selectedCount = computed(() => batchStore.selectedCount)
const isBatchProcessing = computed(() => batchStore.isBatchProcessing)
const overallProgress = computed(() => batchStore.overallProgress)
const processingCount = computed(() => batchStore.processingCount)

// ============================================================
// 操作映射
// ============================================================

/** 操作类型对应的中文名称 */
const actionLabels: Record<BatchActionType, string> = {
  start: '启动',
  stop: '停止',
  shutdown: '关机',
  migrate: '迁移',
  backup: '备份',
  reboot: '重启',
}

/** 操作类型对应的确认提示 */
const actionConfirmMessages: Record<BatchActionType, string> = {
  start: `确认批量启动 ${selectedCount.value} 个虚拟机/容器？`,
  stop: `确认批量停止 ${selectedCount.value} 个虚拟机/容器？`,
  shutdown: `确认批量关机 ${selectedCount.value} 个虚拟机/容器？`,
  migrate: `确认批量迁移 ${selectedCount.value} 个虚拟机/容器？`,
  backup: `确认批量备份 ${selectedCount.value} 个虚拟机/容器？`,
  reboot: `确认批量重启 ${selectedCount.value} 个虚拟机/容器？`,
}

// ============================================================
// 方法
// ============================================================

/**
 * 清空选择
 */
function handleClearSelection(): void {
  batchStore.clearSelection()
}

/**
 * 执行批量操作
 * 显示确认对话框后执行对应的批量操作
 * @param action 操作类型
 */
async function handleAction(action: BatchActionType): Promise<void> {
  const label = actionLabels[action]
  const message = actionConfirmMessages[action]

  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      `${message}\n\n此操作将依次对每个资源执行${label}操作，执行过程中可随时取消。`,
      `确认批量${label}`,
      {
        confirmButtonText: `确认${label}`,
        cancelButtonText: '取消',
        type: action === 'start' ? 'info' : 'warning',
        distinguishCancelAndClose: true,
      },
    )

    // 记录当前操作类型
    currentAction.value = action

    // 显示进度弹窗
    progressVisible.value = true

    // 执行批量操作
    const result = await batchStore.executeBatch(action, async (item) => {
      // TODO: 替换为真实的 API 调用
      // 模拟 API 调用延迟
      await simulateApiCall(item, action)
    })

    // 显示结果摘要
    if (result.errorCount === 0) {
      ElMessage.success(
        `批量${label}完成，成功 ${result.successCount} 项`,
      )
    } else {
      ElMessage.warning(
        `批量${label}完成，成功 ${result.successCount} 项，失败 ${result.errorCount} 项`,
      )
    }
  } catch (error) {
    // 用户取消或关闭对话框
    if (error === 'cancel' || error === 'close') {
      return
    }
    console.error('批量操作失败:', error)
    ElMessage.error('批量操作执行失败')
  }
}

/**
 * 模拟 API 调用 (开发阶段使用)
 * @param item 资源项
 * @param action 操作类型
 */
async function simulateApiCall(
  item: { id: string; name: string; type: string; vmid: number },
  action: BatchActionType,
): Promise<void> {
  // 模拟网络延迟 500-2000ms
  const delay = 500 + Math.random() * 1500
  await new Promise((resolve) => setTimeout(resolve, delay))

  // 模拟 10% 失败率
  if (Math.random() < 0.1) {
    throw new Error(`模拟失败: ${item.name}`)
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.batch-operation-bar {
  position: sticky;
  top: 0;
  z-index: 100;
  margin-bottom: $spacing-4;
  background: linear-gradient(135deg, $primary-7 0%, $primary-9 100%);
  border-radius: $radius-base;
  box-shadow: $shadow-3;
  overflow: hidden;
}

.batch-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-3 $spacing-4;
  gap: $spacing-4;
  min-height: 48px;

  @media (max-width: $breakpoint-md) {
    flex-wrap: wrap;
    padding: $spacing-2 $spacing-3;
  }
}

.batch-bar-left {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  flex-shrink: 0;

  .batch-checkbox {
    :deep(.el-checkbox__inner) {
      background-color: $primary-4;
      border-color: $primary-4;
    }
  }

  .batch-count {
    color: $color-text-primary;
    font-size: $font-size-base;
    white-space: nowrap;

    strong {
      color: $primary-2;
      font-weight: $font-weight-bold;
    }
  }

  :deep(.el-button) {
    color: $primary-2;

    &:hover {
      color: $color-text-primary;
    }
  }
}

.batch-bar-center {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  flex-wrap: wrap;

  @media (max-width: $breakpoint-md) {
    order: 3;
    width: 100%;
    justify-content: center;
    padding-top: $spacing-2;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }

  :deep(.el-button) {
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.1);
    color: $color-text-primary;
    transition: $transition-base;

    &:hover:not(:disabled) {
      background: rgba(255, 255, 255, 0.2);
      border-color: rgba(255, 255, 255, 0.4);
    }

    &:disabled {
      opacity: 0.5;
    }

    .el-icon {
      margin-right: $spacing-1;
    }
  }
}

.batch-bar-right {
  flex-shrink: 0;

  :deep(.el-button) {
    color: rgba(255, 255, 255, 0.7);

    &:hover {
      color: $color-text-primary;
    }
  }
}

.batch-progress-indicator {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  padding: $spacing-2 $spacing-4;
  background: rgba(0, 0, 0, 0.15);

  .batch-progress-bar {
    flex: 1;

    :deep(.el-progress-bar__outer) {
      background: rgba(255, 255, 255, 0.1);
    }

    :deep(.el-progress-bar__inner) {
      background: $primary-4;
      transition: width $duration-slow $ease-out;
    }
  }

  .batch-progress-text {
    color: rgba(255, 255, 255, 0.8);
    font-size: $font-size-xs;
    white-space: nowrap;
  }
}

// 动画
.batch-bar-enter-active {
  transition: all $duration-slow $ease-out;
}

.batch-bar-leave-active {
  transition: all $duration-normal $ease-in;
}

.batch-bar-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.batch-bar-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
