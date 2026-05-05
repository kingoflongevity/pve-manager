<template>
  <div class="console-toolbar">
    <!-- 左侧: 虚拟机名称和连接状态 -->
    <div class="toolbar-left">
      <span class="vm-name">
        <el-tag size="small" effect="dark">{{ vmName }}</el-tag>
        VM {{ vmid }}
      </span>

      <el-divider direction="vertical" />

      <!-- 连接状态指示器 -->
      <div class="connection-status" :class="statusClass">
        <span class="status-dot"></span>
        <span class="status-text">{{ statusText }}</span>
      </div>

      <!-- 错误信息提示 -->
      <span v-if="connectionStatus === 'error' && error" class="error-hint">
        {{ error }}
      </span>
    </div>

    <!-- 右侧: 操作按钮 -->
    <div class="toolbar-right">
      <!-- 发送快捷键 -->
      <el-dropdown trigger="click" :disabled="connectionStatus !== 'connected'">
        <el-button size="small" :disabled="connectionStatus !== 'connected'">
          <el-icon><Key /></el-icon>
          发送快捷键
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="emit('send-combo', 'ctrl-alt-del')">
              Ctrl+Alt+Del
            </el-dropdown-item>
            <el-dropdown-item @click="emit('send-combo', 'ctrl-alt-backspace')">
              Ctrl+Alt+Backspace
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 剪贴板同步 -->
      <el-tooltip content="同步剪贴板（本地 → 虚拟机）" placement="bottom">
        <el-button
          size="small"
          :disabled="connectionStatus !== 'connected'"
          @click="handleClipboard"
        >
          <el-icon><CopyDocument /></el-icon>
          剪贴板
        </el-button>
      </el-tooltip>

      <el-divider direction="vertical" />

      <!-- 缩放控制 -->
      <el-dropdown trigger="click">
        <el-button size="small">
          <el-icon><ZoomIn /></el-icon>
          {{ zoomLabel }}
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :class="{ active: zoom === 'auto' }" @click="setZoom('auto')">
              自适应
            </el-dropdown-item>
            <el-dropdown-item :class="{ active: zoom === '100' }" @click="setZoom('100')">
              100%
            </el-dropdown-item>
            <el-dropdown-item :class="{ active: zoom === '50' }" @click="setZoom('50')">
              50%
            </el-dropdown-item>
            <el-dropdown-item :class="{ active: zoom === '25' }" @click="setZoom('25')">
              25%
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-divider direction="vertical" />

      <!-- 全屏切换 -->
      <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
        <el-button size="small" @click="toggleFullscreen">
          <el-icon><FullScreen /></el-icon>
        </el-button>
      </el-tooltip>

      <!-- 重新连接 -->
      <el-tooltip content="重新连接" placement="bottom">
        <el-button
          size="small"
          type="warning"
          :disabled="connectionStatus === 'connecting'"
          @click="emit('reconnect')"
        >
          <el-icon><RefreshRight /></el-icon>
        </el-button>
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  Key,
  ArrowDown,
  CopyDocument,
  ZoomIn,
  FullScreen,
  RefreshRight,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

/** 组件属性 */
interface Props {
  /** 虚拟机名称 */
  vmName: string
  /** 虚拟机 ID */
  vmid: number
  /** 连接状态 */
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'error'
  /** 错误信息 */
  error?: string
}

/** 事件定义 */
interface Emits {
  /** 发送按键组合 */
  (e: 'send-combo', combo: 'ctrl-alt-del' | 'ctrl-alt-backspace'): void
  /** 重新连接 */
  (e: 'reconnect'): void
  /** 缩放级别变更 */
  (e: 'zoom-change', level: 'auto' | '100' | '50' | '25'): void
  /** 剪贴板内容 */
  (e: 'clipboard', text: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

/** 是否全屏 */
const isFullscreen = ref(false)

/** 缩放级别: 'auto' | '100' | '50' | '25' */
const zoom = ref<'auto' | '100' | '50' | '25'>('auto')

/**
 * 缩放级别显示文本
 */
const zoomLabel = computed(() => {
  const map: Record<string, string> = {
    auto: '自适应',
    '100': '100%',
    '50': '50%',
    '25': '25%',
  }
  return map[zoom.value] || '自适应'
})

/**
 * 连接状态 CSS 类
 */
const statusClass = computed(() => {
  return `status-${props.connectionStatus}`
})

/**
 * 连接状态文本
 */
const statusText = computed(() => {
  const map: Record<string, string> = {
    disconnected: '未连接',
    connecting: '连接中...',
    connected: '已连接',
    error: '连接失败',
  }
  return map[props.connectionStatus] || '未连接'
})

/**
 * 切换全屏状态
 */
function toggleFullscreen() {
  if (!isFullscreen.value) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

/**
 * 设置缩放级别
 */
function setZoom(level: 'auto' | '100' | '50' | '25') {
  zoom.value = level
  emit('zoom-change', level)
}

/**
 * 处理剪贴板同步
 * 将本地剪贴板内容发送到虚拟机
 */
async function handleClipboard() {
  try {
    const text = await navigator.clipboard.readText()
    if (text) {
      emit('clipboard', text)
      ElMessage.success('剪贴板内容已发送到虚拟机')
    } else {
      ElMessage.warning('剪贴板为空')
    }
  } catch {
    ElMessage.error('无法读取剪贴板，请检查浏览器权限')
  }
}

/**
 * 全屏状态变化监听
 */
function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  document.addEventListener('fullscreenchange', onFullscreenChange)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})
</script>

<style lang="scss" scoped>
.console-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 40px;
  background: #2d2d2d;
  border-bottom: 1px solid #3e3e3e;
  flex-shrink: 0;

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 12px;

    .vm-name {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #d4d4d4;
      font-size: 13px;
    }

    .error-hint {
      font-size: 12px;
      color: #f56c6c;
      max-width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;

    :deep(.el-button) {
      color: #d4d4d4;

      &:hover {
        color: #fff;
        background: rgba(255, 255, 255, 0.1);
      }

      &:disabled {
        color: #555;
      }
    }
  }
}

// 连接状态
.connection-status {
  display: flex;
  align-items: center;
  gap: 4px;

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-text {
    font-size: 12px;
    white-space: nowrap;
  }
}

.status-disconnected,
.status-error {
  .status-dot {
    background: #f56c6c;
  }
  .status-text {
    color: #f56c6c;
  }
}

.status-connecting {
  .status-dot {
    background: #e6a23c;
    animation: pulse 1s ease-in-out infinite;
  }
  .status-text {
    color: #e6a23c;
  }
}

.status-connected {
  .status-dot {
    background: #67c23a;
  }
  .status-text {
    color: #67c23a;
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.3;
  }
}

// 缩放选项激活状态
:deep(.el-dropdown-menu__item.active) {
  color: #409eff;
  font-weight: bold;
}
</style>
