<template>
  <!-- 遮罩层 -->
  <Transition name="mask-fade">
    <div v-if="visible" class="task-center-mask" @click="handleClose" />
  </Transition>

  <!-- 侧滑面板 -->
  <Transition name="drawer-slide">
    <div v-if="visible" class="task-center" :style="{ width: panelWidth + 'px' }">
      <!-- 面板头部 -->
      <div class="task-center-header">
        <div class="header-title">
          <el-icon :size="20" class="title-icon"><List /></el-icon>
          <h2>任务中心</h2>
          <el-badge
            v-if="taskStore.taskCount > 0"
            :value="taskStore.taskCount"
            :max="99"
            class="running-badge"
          />
        </div>
        <div class="header-actions">
          <el-tooltip content="刷新任务" placement="bottom">
            <el-button
              text
              :loading="taskStore.loading"
              @click="handleRefresh"
            >
              <el-icon><Refresh /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="清除已完成" placement="bottom">
            <el-button
              text
              :disabled="taskStore.completedTasks.length === 0"
              @click="handleClearCompleted"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="关闭面板" placement="bottom">
            <el-button text @click="handleClose">
              <el-icon><Close /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>

      <!-- 状态筛选 -->
      <div class="task-filter">
        <el-radio-group v-model="activeFilter" size="small">
          <el-radio-button value="all">
            全部 ({{ taskStore.tasks.length }})
          </el-radio-button>
          <el-radio-button value="running">
            <el-icon class="filter-icon"><Loading /></el-icon>
            运行中 ({{ taskStore.runningTasks.length }})
          </el-radio-button>
          <el-radio-button value="success">成功</el-radio-button>
          <el-radio-button value="error">失败</el-radio-button>
        </el-radio-group>
      </div>

      <!-- 任务列表 -->
      <div class="task-list-container">
        <div v-if="filteredTasks.length === 0" class="empty-state">
          <el-empty :description="emptyDescription" :image-size="80">
            <template #image>
              <el-icon :size="60" class="empty-icon"><DocumentChecked /></el-icon>
            </template>
          </el-empty>
        </div>

        <div v-else class="task-list">
          <div
            v-for="task in filteredTasks"
            :key="task.upid"
            class="task-item"
            :class="`task-${task.status}`"
          >
            <!-- 状态图标 -->
            <div class="task-status-icon">
              <el-icon v-if="task.status === 'running'" :size="20" class="status-running">
                <Loading />
              </el-icon>
              <el-icon v-else-if="task.status === 'success'" :size="20" class="status-success">
                <CircleCheckFilled />
              </el-icon>
              <el-icon v-else-if="task.status === 'error'" :size="20" class="status-error">
                <CircleCloseFilled />
              </el-icon>
              <el-icon v-else :size="20" class="status-stopped">
                <VideoPause />
              </el-icon>
            </div>

            <!-- 任务信息 -->
            <div class="task-content">
              <div class="task-description">
                <span class="desc-text" :title="task.description">{{ task.description }}</span>
                <el-tag :type="statusTagType(task.status)" size="small" class="status-tag">
                  {{ statusText(task.status) }}
                </el-tag>
              </div>

              <!-- 进度条 -->
              <el-progress
                :percentage="task.progress"
                :status="progressStatus(task.status)"
                :stroke-width="6"
                :show-text="false"
                class="task-progress"
              />

              <!-- 时间和耗时 -->
              <div class="task-meta">
                <span class="task-id" title="任务 UPID">{{ task.id }}</span>
                <span class="meta-divider">|</span>
                <span class="task-time">开始: {{ formatTime(task.starttime) }}</span>
                <span class="meta-divider">|</span>
                <span class="task-duration">{{ formatDuration(task.starttime, task.endtime) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 面板底部 -->
      <div class="task-center-footer">
        <div class="auto-refresh-indicator">
          <el-switch
            v-model="autoRefresh"
            size="small"
            @change="handleAutoRefreshChange"
          />
          <span>自动刷新</span>
        </div>
        <span class="refresh-time">
          上次刷新: {{ formatRefreshTime(taskStore.lastRefresh) }}
        </span>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useTaskStore } from '@/stores/tasks'
import { formatBytes, formatRelativeTime } from '@/utils/format'
import type { Task } from '@/api/taskTypes'
import type { TaskStatusFilter, TaskStatus } from '@/api/taskTypes'
import {
  List,
  Refresh,
  Delete,
  Close,
  Loading,
  CircleCheckFilled,
  CircleCloseFilled,
  VideoPause,
  DocumentChecked,
} from '@element-plus/icons-vue'

interface Props {
  /** 面板可见性 */
  modelValue: boolean
  /** 面板宽度（像素） */
  width?: number
}

const props = withDefaults(defineProps<Props>(), {
  width: 480,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const taskStore = useTaskStore()

// ===== 面板状态 =====

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const panelWidth = computed(() => props.width)

// ===== 筛选 =====

const activeFilter = ref<TaskStatusFilter>('all')

const filteredTasks = computed(() => taskStore.getTasksByFilter(activeFilter.value))

const emptyDescription = computed(() => {
  if (activeFilter.value === 'all') return '暂无任务记录'
  return `暂无${statusText(activeFilter.value as TaskStatus)}的任务`
})

// ===== 自动刷新 =====

const autoRefresh = ref(false)

function handleAutoRefreshChange(val: boolean) {
  if (val) {
    taskStore.startPolling()
  } else {
    taskStore.stopPolling()
  }
}

// 面板打开时自动开启轮询
watch(visible, (val) => {
  if (val && autoRefresh.value) {
    taskStore.startPolling()
  } else if (!val) {
    taskStore.stopPolling()
  }
})

// ===== 事件处理 =====

function handleClose() {
  visible.value = false
}

async function handleRefresh() {
  await taskStore.refreshTasks()
}

function handleClearCompleted() {
  taskStore.clearCompleted()
}

// ===== 格式化辅助 =====

/**
 * 将时间戳格式化为本地时间字符串
 */
function formatTime(timestamp: number): string {
  if (!timestamp) return '--'
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  // 1 分钟内
  if (diff < 60000) {
    return '刚刚'
  }
  // 1 小时内
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)} 分钟前`
  }
  // 今天
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  // 其他日期
  return date.toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * 计算任务耗时
 */
function formatDuration(starttime: number, endtime: number): string {
  if (!starttime) return '--'
  const end = endtime || Date.now() / 1000
  const diff = Math.floor(end - starttime)

  if (diff < 60) return `${diff}s`
  if (diff < 3600) {
    const m = Math.floor(diff / 60)
    const s = diff % 60
    return `${m}m ${s}s`
  }
  const h = Math.floor(diff / 3600)
  const m = Math.floor((diff % 3600) / 60)
  return `${h}h ${m}m`
}

/**
 * 格式化上次刷新时间
 */
function formatRefreshTime(timestamp: number): string {
  if (!timestamp) return '--'
  return new Date(timestamp).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

/**
 * 获取状态对应的 Tag 颜色类型
 */
function statusTagType(status: TaskStatus): 'success' | 'danger' | 'warning' | 'info' {
  const map: Record<TaskStatus, 'success' | 'danger' | 'warning' | 'info'> = {
    running: 'warning',
    success: 'success',
    error: 'danger',
    stopped: 'info',
  }
  return map[status] || 'info'
}

/**
 * 获取状态对应的中文文本
 */
function statusText(status: string): string {
  const map: Record<string, string> = {
    running: '运行中',
    success: '成功',
    error: '失败',
    stopped: '已停止',
    all: '全部',
  }
  return map[status] || status
}

/**
 * 获取进度条状态类型
 */
function progressStatus(status: TaskStatus): 'success' | 'exception' | '' {
  if (status === 'success') return 'success'
  if (status === 'error') return 'exception'
  return ''
}

// 组件卸载时清理轮询
onUnmounted(() => {
  taskStore.stopPolling()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

// ===== 遮罩层 =====
.task-center-mask {
  position: fixed;
  inset: 0;
  background: $color-bg-mask;
  z-index: $z-index-drawer;
}

// ===== 面板主体 =====
.task-center {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  background: $color-bg-container;
  box-shadow: $shadow-4;
  z-index: $z-index-modal;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

// ===== 面板头部 =====
.task-center-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-5 $spacing-6;
  border-bottom: 1px solid $color-border-light;
  background: $color-bg-container;
  flex-shrink: 0;

  .header-title {
    display: flex;
    align-items: center;
    gap: $spacing-3;

    h2 {
      font-size: $font-size-lg;
      font-weight: $font-weight-semibold;
      color: $color-text-primary;
      margin: 0;
    }

    .title-icon {
      color: $color-primary;
    }

    .running-badge {
      margin-left: $spacing-1;
    }
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: $spacing-1;
  }
}

// ===== 筛选栏 =====
.task-filter {
  padding: $spacing-3 $spacing-6;
  border-bottom: 1px solid $color-border-lighter;
  background: $gray-2;
  flex-shrink: 0;

  .filter-icon {
    animation: spin 1s linear infinite;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// ===== 任务列表 =====
.task-list-container {
  flex: 1;
  overflow-y: auto;
  padding: $spacing-3 $spacing-4;

  &::-webkit-scrollbar {
    width: 6px;
  }
  &::-webkit-scrollbar-thumb {
    background: $gray-5;
    border-radius: $radius-full;
    &:hover { background: $gray-6; }
  }
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 200px;

  .empty-icon {
    color: $gray-5;
  }
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-3;
}

.task-item {
  display: flex;
  gap: $spacing-3;
  padding: $spacing-4;
  border-radius: $radius-base;
  border: 1px solid $color-border-lighter;
  background: $color-bg-container;
  transition: $transition-fast;

  &:hover {
    border-color: $color-border-base;
    box-shadow: $shadow-1;
  }

  &.task-running {
    border-left: 3px solid $color-warning;
  }
  &.task-success {
    border-left: 3px solid $color-success;
  }
  &.task-error {
    border-left: 3px solid $color-danger;
  }
  &.task-stopped {
    border-left: 3px solid $gray-5;
  }
}

.task-status-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: $radius-full;

  .status-running {
    color: $color-warning;
    animation: spin 1.5s linear infinite;
  }
  .status-success {
    color: $color-success;
  }
  .status-error {
    color: $color-danger;
  }
  .status-stopped {
    color: $gray-5;
  }
}

.task-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: $spacing-2;
}

.task-description {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-2;

  .desc-text {
    font-size: $font-size-sm;
    font-weight: $font-weight-medium;
    color: $color-text-primary;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-tag {
    flex-shrink: 0;
  }
}

.task-progress {
  margin-top: $spacing-1;
}

.task-meta {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  font-size: $font-size-xs;
  color: $color-text-secondary;
  flex-wrap: wrap;

  .task-id {
    font-family: $font-family-code;
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta-divider {
    color: $gray-5;
  }
}

// ===== 面板底部 =====
.task-center-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-3 $spacing-6;
  border-top: 1px solid $color-border-lighter;
  background: $gray-2;
  flex-shrink: 0;
  font-size: $font-size-xs;
  color: $color-text-secondary;

  .auto-refresh-indicator {
    display: flex;
    align-items: center;
    gap: $spacing-2;
  }
}

// ===== 过渡动画 =====
.mask-fade-enter-active,
.mask-fade-leave-active {
  transition: opacity $duration-slow $ease-base;
}
.mask-fade-enter-from,
.mask-fade-leave-to {
  opacity: 0;
}

.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: transform $duration-slow $ease-base;
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateX(100%);
}
</style>
