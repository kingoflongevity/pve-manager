<template>
  <div class="qemu-console-view">
    <!-- 控制台工具栏 -->
    <div class="console-toolbar">
      <div class="toolbar-left">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回详情
        </el-button>
        <span class="divider">|</span>
        <span class="vm-info">
          <el-tag size="small">{{ vmName }}</el-tag>
          <span class="vm-id">VM {{ vmid }}</span>
        </span>
      </div>

      <div class="toolbar-right">
        <!-- 连接状态指示器 -->
        <div class="connection-status" :class="statusClass">
          <span class="status-dot"></span>
          <span class="status-text">{{ statusText }}</span>
        </div>

        <el-divider direction="vertical" />

        <!-- 发送 Ctrl+Alt+Del -->
        <el-tooltip content="发送 Ctrl+Alt+Del" placement="bottom">
          <el-button text :disabled="!isConnected" @click="sendCtrlAltDel">
            <el-icon><Key /></el-icon>
          </el-button>
        </el-tooltip>

        <!-- 全屏切换 -->
        <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
          <el-button text @click="toggleFullscreen">
            <el-icon><FullScreen /></el-icon>
          </el-button>
        </el-tooltip>

        <!-- 剪贴板 -->
        <el-tooltip content="剪贴板" placement="bottom">
          <el-button text :disabled="!isConnected" @click="handleClipboard">
            <el-icon><CopyDocument /></el-icon>
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- VNC 控制台容器 -->
    <div ref="consoleContainer" class="console-container">
      <div v-if="!isConnected" class="console-placeholder">
        <el-icon class="placeholder-icon"><Monitor /></el-icon>
        <p class="placeholder-text">VNC 控制台加载中...</p>
        <p class="placeholder-hint">
          实际 noVNC 集成将在此处渲染远程桌面
        </p>
        <el-button type="primary" @click="connectConsole">
          <el-icon><Connection /></el-icon>
          连接控制台
        </el-button>
      </div>

      <!-- noVNC 渲染区域 -->
      <div v-show="isConnected" class="vnc-canvas"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft,
  Key,
  FullScreen,
  CopyDocument,
  Monitor,
  Connection,
} from '@element-plus/icons-vue'
import { getQEMUConfig } from '@/api/qemu'

const router = useRouter()
const route = useRoute()

const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const vmName = ref('')
const isConnected = ref(false)
const consoleContainer = ref<HTMLElement | null>(null)

// 连接状态（预留，用于实际 noVNC 集成）
type ConnectionStatusType = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatusType>('disconnected')

/**
 * 状态 CSS 类
 */
const statusClass = computed(() => {
  const map: Record<string, string> = {
    disconnected: 'status-disconnected',
    connecting: 'status-connecting',
    connected: 'status-connected',
    error: 'status-error',
  }
  return map[connectionStatus.value] || 'status-disconnected'
})

/**
 * 状态文字
 */
const statusText = computed(() => {
  const map: Record<string, string> = {
    disconnected: '未连接',
    connecting: '连接中...',
    connected: '已连接',
    error: '连接失败',
  }
  return map[connectionStatus.value] || '未连接'
})

/** 是否全屏 */
const isFullscreen = ref(false)

/**
 * 返回详情页
 */
function goBack() {
  router.push({ name: 'QEMUDetail', params: { node: node.value, vmid: vmid.value.toString() } })
}

/**
 * 加载虚拟机名称
 */
async function loadVMInfo() {
  try {
    const config = await getQEMUConfig(node.value, vmid.value)
    vmName.value = config.name || `VM ${vmid.value}`
  } catch {
    vmName.value = `VM ${vmid.value}`
  }
}

/**
 * 连接控制台（占位，实际 noVNC 集成在后续任务中实现）
 */
function connectConsole() {
  connectionStatus.value = 'connecting'
  // TODO: 实际 noVNC 集成
  // 1. 从后端获取 VNC proxy URL
  // 2. 使用 noVNC 库连接到 VNC proxy
  // 3. 渲染到 consoleContainer 中
  setTimeout(() => {
    connectionStatus.value = 'error'
    ElMessage.warning('VNC 控制台集成将在后续版本实现')
  }, 1500)
}

/**
 * 发送 Ctrl+Alt+Del（占位）
 */
function sendCtrlAltDel() {
  ElMessage.info('Ctrl+Alt+Del 将在 noVNC 集成后实现')
}

/**
 * 切换全屏
 */
function toggleFullscreen() {
  if (!consoleContainer.value) return

  if (!isFullscreen.value) {
    consoleContainer.value.requestFullscreen?.()
    isFullscreen.value = true
  } else {
    document.exitFullscreen?.()
    isFullscreen.value = false
  }
}

/**
 * 处理剪贴板（占位）
 */
function handleClipboard() {
  ElMessage.info('剪贴板功能将在 noVNC 集成后实现')
}

/**
 * 监听全屏变化事件
 */
function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  loadVMInfo()
  document.addEventListener('fullscreenchange', onFullscreenChange)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.qemu-console-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1e1e1e;
}

// 工具栏
.console-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 $spacing-4;
  height: 48px;
  background: #2d2d2d;
  border-bottom: 1px solid #3e3e3e;
  flex-shrink: 0;

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: $spacing-3;
    color: #d4d4d4;

    .divider {
      color: #555;
    }

    .vm-info {
      display: flex;
      align-items: center;
      gap: $spacing-2;

      .vm-id {
        font-size: $font-size-xs;
        color: #888;
      }
    }
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: $spacing-2;

    :deep(.el-button) {
      color: #d4d4d4;

      &:hover {
        color: #fff;
        background: rgba(255, 255, 255, 0.1);
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
    font-size: $font-size-xs;
    white-space: nowrap;
  }
}

.status-connected {
  .status-dot {
    background: #52c41a;
  }
  .status-text {
    color: #52c41a;
  }
}

.status-connecting {
  .status-dot {
    background: #faad14;
    animation: pulse 1s ease-in-out infinite;
  }
  .status-text {
    color: #faad14;
  }
}

.status-disconnected,
.status-error {
  .status-dot {
    background: #f5222d;
  }
  .status-text {
    color: #f5222d;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

// 控制台容器
.console-container {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: #000;
}

.console-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;

  .placeholder-icon {
    font-size: 64px;
    margin-bottom: $spacing-6;
    color: #444;
  }

  .placeholder-text {
    font-size: $font-size-lg;
    color: #999;
    margin-bottom: $spacing-2;
  }

  .placeholder-hint {
    font-size: $font-size-sm;
    color: #666;
    margin-bottom: $spacing-8;
  }
}

.vnc-canvas {
  width: 100%;
  height: 100%;
}
</style>
