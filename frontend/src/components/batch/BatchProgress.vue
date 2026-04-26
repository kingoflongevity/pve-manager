<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="680px"
    :close-on-click-modal="false"
    :close-on-press-escape="!isProcessing"
    :show-close="!isProcessing"
    @close="handleClose"
  >
    <!-- 总体进度 -->
    <div class="batch-progress-header">
      <div class="progress-summary">
        <div class="summary-item success">
          <el-icon class="summary-icon"><CircleCheckFilled /></el-icon>
          <span>成功 {{ successCount }}</span>
        </div>
        <div class="summary-item error">
          <el-icon class="summary-icon"><CircleCloseFilled /></el-icon>
          <span>失败 {{ errorCount }}</span>
        </div>
        <div v-if="processingCount > 0" class="summary-item processing">
          <el-icon class="summary-icon"><Loading /></el-icon>
          <span>处理中 {{ processingCount }}</span>
        </div>
      </div>

      <el-progress
        :percentage="overallProgress"
        :status="progressStatus"
        :stroke-width="8"
        class="overall-progress"
      />

      <!-- 选项 -->
      <div class="progress-options">
        <el-checkbox
          v-model="continueOnError"
          :disabled="isProcessing"
          @change="batchStore.toggleContinueOnError"
        >
          遇到错误时继续
        </el-checkbox>
      </div>
    </div>

    <!-- 项目列表 -->
    <div class="batch-progress-list">
      <div
        v-for="item in batchProgress"
        :key="item.id"
        class="progress-item"
        :class="`status-${item.status}`"
      >
        <!-- 状态图标 -->
        <div class="item-status">
          <!-- 加载中 - 旋转图标 -->
          <el-icon v-if="item.status === 'processing'" class="icon-spinning">
            <Loading />
          </el-icon>
          <!-- 成功 -->
          <el-icon v-else-if="item.status === 'success'" class="icon-success">
            <CircleCheckFilled />
          </el-icon>
          <!-- 错误 -->
          <el-icon v-else-if="item.status === 'error'" class="icon-error">
            <CircleCloseFilled />
          </el-icon>
          <!-- 等待中 -->
          <el-icon v-else class="icon-pending">
            <Clock />
          </el-icon>
        </div>

        <!-- 项目信息 -->
        <div class="item-info">
          <div class="item-header">
            <span class="item-name">{{ item.name }}</span>
            <span class="item-type">{{ item.type === 'vm' ? '虚拟机' : '容器' }}</span>
            <span class="item-id">#{{ item.vmid }}</span>
          </div>

          <!-- 错误详情 (可展开) -->
          <div v-if="item.status === 'error' && item.error" class="item-error">
            <el-button
              link
              type="danger"
              size="small"
              @click="toggleErrorDetail(item.id)"
            >
              <el-icon><WarningFilled /></el-icon>
              {{ expandedErrors[item.id] ? '收起详情' : '查看详情' }}
            </el-button>
            <Transition name="error-detail">
              <div v-if="expandedErrors[item.id]" class="error-detail">
                {{ item.error }}
              </div>
            </Transition>
          </div>
        </div>

        <!-- 项目进度 -->
        <div class="item-progress">
          <el-progress
            :percentage="item.progress"
            :status="itemStatusMap[item.status]"
            :stroke-width="4"
            :show-text="false"
            class="item-progress-bar"
          />
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <template #footer>
      <div class="batch-progress-footer">
        <div class="footer-left">
          <el-button
            v-if="hasFailedItems"
            type="warning"
            size="small"
            :disabled="isProcessing"
            @click="handleRetryFailed"
          >
            <el-icon><RefreshRight /></el-icon>
            重试失败项
          </el-button>
        </div>
        <div class="footer-right">
          <el-button
            v-if="isProcessing"
            type="danger"
            size="default"
            @click="handleCancel"
          >
            <el-icon><CircleClose /></el-icon>
            取消操作
          </el-button>
          <el-button
            v-if="isComplete"
            type="primary"
            size="default"
            @click="handleClose"
          >
            完成
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  CircleCheckFilled,
  CircleCloseFilled,
  Loading,
  Clock,
  WarningFilled,
  RefreshRight,
  CircleClose,
} from '@element-plus/icons-vue'
import { useBatchStore, type BatchActionType, type BatchOperationItem } from '@/stores/batch'

const batchStore = useBatchStore()

// ============================================================
// Props & Emits
// ============================================================

const props = defineProps<{
  visible: boolean
  action: BatchActionType
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

// ============================================================
// 操作映射
// ============================================================

/** 操作类型中文名称 */
const actionLabels: Record<BatchActionType, string> = {
  start: '批量启动',
  stop: '批量停止',
  shutdown: '批量关机',
  migrate: '批量迁移',
  backup: '批量备份',
  reboot: '批量重启',
}

/** 项目状态对应的进度条状态 */
const itemStatusMap: Record<string, 'success' | 'exception' | undefined> = {
  success: 'success',
  error: 'exception',
  processing: undefined,
  pending: undefined,
}

// ============================================================
// 状态
// ============================================================

/** 展开的错误详情 ID 集合 */
const expandedErrors = ref<Record<string, boolean>>({})

// ============================================================
// 计算属性
// ============================================================

/** 对话框可见性 */
const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

/** 对话框标题 */
const dialogTitle = computed(() => actionLabels[props.action])

/** 批量操作进度列表 */
const batchProgress = computed<BatchOperationItem[]>(() => batchStore.batchProgress)

/** 是否正在处理 */
const isProcessing = computed(() => batchStore.isBatchProcessing)

/** 整体进度 */
const overallProgress = computed(() => batchStore.overallProgress)

/** 成功数 */
const successCount = computed(() => batchStore.successCount)

/** 失败数 */
const errorCount = computed(() => batchStore.errorCount)

/** 处理中数 */
const processingCount = computed(() => batchStore.processingCount)

/** 是否全部完成 */
const isComplete = computed(() => batchStore.isComplete)

/** 是否有失败项 */
const hasFailedItems = computed(() => batchStore.errorCount > 0)

/** 遇到错误时是否继续 */
const continueOnError = computed(() => batchStore.continueOnError)

/** 进度条状态 */
const progressStatus = computed<'success' | 'exception' | undefined>(() => {
  if (isComplete.value) {
    return errorCount.value > 0 ? 'exception' : 'success'
  }
  return undefined
})

// ============================================================
// 方法
// ============================================================

/**
 * 切换错误详情展开/收起
 * @param id 项目 ID
 */
function toggleErrorDetail(id: string): void {
  expandedErrors.value[id] = !expandedErrors.value[id]
}

/**
 * 取消操作
 */
async function handleCancel(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      '确认取消当前批量操作？已处理的项目不会回滚。',
      '取消操作',
      {
        confirmButtonText: '确认取消',
        cancelButtonText: '继续执行',
        type: 'warning',
      },
    )
    batchStore.cancelOperation()
    ElMessage.warning('批量操作已取消')
  } catch {
    // 用户选择继续执行
  }
}

/**
 * 重试失败项
 */
async function handleRetryFailed(): Promise<void> {
  ElMessage.info('正在重试失败的项目...')
  // TODO: 根据 action 类型实现重试逻辑
}

/**
 * 关闭对话框
 */
function handleClose(): void {
  dialogVisible.value = false
  // 清理展开的错误详情
  expandedErrors.value = {}
  // 重置 store 状态
  batchStore.reset()
}

// ============================================================
// 监听器
// ============================================================

/** 监听对话框关闭，清理状态 */
watch(
  () => props.visible,
  (val) => {
    if (!val) {
      expandedErrors.value = {}
    }
  },
)
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

// ============================================================
// 总体进度区域
// ============================================================

.batch-progress-header {
  margin-bottom: $spacing-4;
}

.progress-summary {
  display: flex;
  align-items: center;
  gap: $spacing-4;
  margin-bottom: $spacing-3;

  .summary-item {
    display: flex;
    align-items: center;
    gap: $spacing-1;
    font-size: $font-size-sm;
    font-weight: $font-weight-medium;

    .summary-icon {
      font-size: 16px;
    }

    &.success {
      color: $color-success;
    }

    &.error {
      color: $color-danger;
    }

    &.processing {
      color: $color-primary;
    }
  }
}

.overall-progress {
  margin-bottom: $spacing-3;
}

.progress-options {
  :deep(.el-checkbox) {
    .el-checkbox__label {
      font-size: $font-size-sm;
      color: $color-text-secondary;
    }
  }
}

// ============================================================
// 项目列表
// ============================================================

.batch-progress-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid $color-border-light;
  border-radius: $radius-base;
  background: $gray-2;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: $slate-600;
    border-radius: $radius-full;
  }
}

.progress-item {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  padding: $spacing-3 $spacing-4;
  background: $color-bg-container;
  border-bottom: 1px solid $color-border-lighter;
  transition: $transition-base;

  &:last-child {
    border-bottom: none;
  }

  &:hover {
    background: $gray-2;
  }

  &.status-error {
    background: $color-danger-bg;
  }

  &.status-success {
    background: $color-success-bg;
  }
}

// 状态图标
.item-status {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;

  .el-icon {
    font-size: 20px;
  }

  .icon-spinning {
    color: $color-primary;
    animation: spin 1s linear infinite;
  }

  .icon-success {
    color: $color-success;
  }

  .icon-error {
    color: $color-danger;
  }

  .icon-pending {
    color: $color-text-placeholder;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 项目信息
.item-info {
  flex: 1;
  min-width: 0;
}

.item-header {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  margin-bottom: $spacing-1;

  .item-name {
    font-size: $font-size-base;
    font-weight: $font-weight-medium;
    color: $color-text-primary;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .item-type {
    font-size: $font-size-xs;
    padding: 1px $spacing-1;
    border-radius: $radius-xs;
    background: $color-info-bg;
    color: $color-info;
    flex-shrink: 0;
  }

  .item-id {
    font-size: $font-size-xs;
    color: $color-text-secondary;
    font-family: $font-family-code;
    flex-shrink: 0;
  }
}

// 错误详情
.item-error {
  margin-top: $spacing-1;

  :deep(.el-button) {
    padding: 0;
    font-size: $font-size-xs;
  }
}

.error-detail {
  margin-top: $spacing-1;
  padding: $spacing-2;
  background: $gray-3;
  border-radius: $radius-xs;
  font-size: $font-size-xs;
  color: $color-danger;
  font-family: $font-family-code;
  word-break: break-all;
}

// 项目进度
.item-progress {
  flex-shrink: 0;
  width: 80px;

  .item-progress-bar {
    :deep(.el-progress-bar__outer) {
      background: $color-border-lighter;
    }
  }
}

// ============================================================
// 底部操作栏
// ============================================================

.batch-progress-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-3;

  .footer-left {
    display: flex;
    align-items: center;
  }

  .footer-right {
    display: flex;
    align-items: center;
    gap: $spacing-3;
  }
}

// ============================================================
// 动画
// ============================================================

.error-detail-enter-active,
.error-detail-leave-active {
  transition: all $duration-normal $ease-base;
}

.error-detail-enter-from,
.error-detail-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}
</style>
