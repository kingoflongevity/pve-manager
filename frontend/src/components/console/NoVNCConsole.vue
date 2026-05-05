<template>
  <div class="novnc-console" :class="{ fullscreen: isFullscreen }">
    <!-- 连接状态指示器 -->
    <div v-if="connectionStatus !== 'connected'" class="status-overlay">
      <div class="status-content">
        <el-icon v-if="connectionStatus === 'connecting'" class="status-icon spinning">
          <Loading />
        </el-icon>
        <el-icon v-else-if="connectionStatus === 'error'" class="status-icon error">
          <CircleClose />
        </el-icon>
        <el-icon v-else class="status-icon">
          <Monitor />
        </el-icon>
        <p class="status-text">{{ statusText }}</p>
        <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
        <el-button
          v-if="connectionStatus === 'error'"
          type="primary"
          size="small"
          @click="reconnect"
        >
          重新连接
        </el-button>
      </div>
    </div>

    <!-- noVNC 渲染区域 -->
    <div ref="screenContainer" class="screen-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Loading, CircleClose, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import RFB from '@novnc/novnc/core/rfb'
import { getQEMUVNCTicket, getLXCVNCTicket } from '@/api/console'

/** 控制台组件属性 */
interface Props {
  /** 节点名称 */
  node: string
  /** 虚拟机/容器 ID */
  vmid: number
  /** 虚拟机类型: QEMU 或 LXC */
  vmType: 'qemu' | 'lxc'
}

const props = defineProps<Props>()

/** noVNC 实例 */
let rfbInstance: RFB | null = null

/** WebSocket 连接状态 */
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatus>('disconnected')

/** 错误信息 */
const errorMessage = ref('')

/** 容器 DOM */
const screenContainer = ref<HTMLElement | null>(null)

/** 是否全屏 */
const isFullscreen = ref(false)

/**
 * 获取状态显示文本
 */
const statusText = computed(() => {
  const map: Record<ConnectionStatus, string> = {
    disconnected: '未连接',
    connecting: '正在连接...',
    connected: '',
    error: '连接失败',
  }
  return map[connectionStatus.value]
})

/**
 * 构建 WebSocket URL
 * 通过后端代理转发 WebSocket 连接到 PVE
 * JWT token 通过 URL 参数传递（因为 noVNC 不支持自定义 WebSocket 头）
 * VNC 端口和票据也通过 URL 参数传递给后端代理
 */
function buildWebSocketUrl(vncPort: number, vncTicket: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const token = localStorage.getItem('pve_token') || ''

  const params = new URLSearchParams({
    vncport: String(vncPort),
    vncticket: vncTicket,
    token: token,
  })

  return `${protocol}//${host}/api/pve/vnc/websocket?${params.toString()}`
}

/**
 * 连接 VNC 控制台
 * 流程：
 * 1. 调用后端 API 获取 VNC 票据（port + ticket）
 * 2. 构建 WebSocket URL 并附加 JWT token、VNC 端口和票据
 * 3. 初始化 RFB (Remote Frame Buffer) 连接
 */
async function connect() {
  if (!screenContainer.value) {
    ElMessage.error('控制台容器未就绪')
    return
  }

  // 清理旧连接
  disconnect()

  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  try {
    // 获取 VNC 代理票据
    let ticket: { port: number; ticket: string }
    if (props.vmType === 'qemu') {
      ticket = await getQEMUVNCTicket(props.node, props.vmid)
    } else {
      ticket = await getLXCVNCTicket(props.node, props.vmid)
    }

    if (!ticket.ticket || !ticket.port) {
      throw new Error('无效的 VNC 票据')
    }

    // 构建 WebSocket URL（通过后端代理转发到 PVE VNC WebSocket）
    const wsUrl = buildWebSocketUrl(ticket.port, ticket.ticket)

    // 初始化 noVNC RFB 连接
    rfbInstance = new RFB(screenContainer.value, wsUrl, {
      credentials: { password: ticket.ticket },
      clipViewport: false,
      viewOnly: false,
      dragViewport: false,
      scaleViewport: false,
    })

    // 注册事件监听
    rfbInstance.addEventListener('connect', onConnect)
    rfbInstance.addEventListener('disconnect', onDisconnect)
    rfbInstance.addEventListener('credentialsrequired', onCredentialsRequired)
    rfbInstance.addEventListener('securityfailure', onSecurityFailure)
    rfbInstance.addEventListener('desktopname', onDesktopName)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : '未知错误'
    connectionStatus.value = 'error'
    errorMessage.value = message
    ElMessage.error(`VNC 连接失败: ${message}`)
  }
}

/**
 * 断开 VNC 连接
 */
function disconnect() {
  if (rfbInstance) {
    try {
      rfbInstance.removeEventListener('connect', onConnect)
      rfbInstance.removeEventListener('disconnect', onDisconnect)
      rfbInstance.removeEventListener('credentialsrequired', onCredentialsRequired)
      rfbInstance.removeEventListener('securityfailure', onSecurityFailure)
      rfbInstance.removeEventListener('desktopname', onDesktopName)
      rfbInstance.disconnect()
    } catch {
      // 忽略断开连接时的错误
    }
    rfbInstance = null
  }
  connectionStatus.value = 'disconnected'
}

/**
 * 重新连接
 */
function reconnect() {
  connect()
}

/**
 * 连接成功回调
 */
function onConnect() {
  connectionStatus.value = 'connected'
  errorMessage.value = ''
}

/**
 * 断开连接回调
 */
function onDisconnect(e: CustomEvent) {
  connectionStatus.value = 'disconnected'
  // 获取断开原因
  const detail = e.detail
  if (detail && detail.clean === false) {
    connectionStatus.value = 'error'
    errorMessage.value = detail.reason || '连接异常断开'
  }
}

/**
 * 需要凭证回调
 */
function onCredentialsRequired() {
  // noVNC 会自动处理凭证
}

/**
 * 安全失败回调
 */
function onSecurityFailure(e: CustomEvent) {
  connectionStatus.value = 'error'
  errorMessage.value = e.detail?.reason || 'VNC 安全认证失败'
  ElMessage.error('VNC 安全认证失败')
}

/**
 * 桌面名称变更回调
 */
function onDesktopName() {
  // 桌面名称变更时记录日志
}

/**
 * 发送按键组合
 * @param keys X11 按键码数组
 */
function sendKeyCombo(keys: number[]) {
  if (!rfbInstance) return
  // 按下所有键
  for (const code of keys) {
    rfbInstance.sendKey(code, '', true)
  }
  // 释放所有键（逆序）
  for (let i = keys.length - 1; i >= 0; i--) {
    rfbInstance.sendKey(keys[i], '', false)
  }
}

/**
 * 发送 Ctrl+Alt+Del
 * X11 按键码: Ctrl=0xffe3, Alt=0xffe9, Delete=0xffff
 */
function sendCtrlAltDel() {
  sendKeyCombo([0xffe3, 0xffe9, 0xffff])
  ElMessage.info('已发送 Ctrl+Alt+Del')
}

/**
 * 发送 Ctrl+Alt+Backspace
 * X11 按键码: Ctrl=0xffe3, Alt=0xffe9, Backspace=0xff08
 */
function sendCtrlAltBackspace() {
  sendKeyCombo([0xffe3, 0xffe9, 0xff08])
  ElMessage.info('已发送 Ctrl+Alt+Backspace')
}

/**
 * 切换缩放模式
 * @param enabled 是否启用视口缩放
 */
function setScaleViewport(enabled: boolean) {
  if (rfbInstance) {
    rfbInstance.scaleViewport = enabled
  }
}

/**
 * 发送剪贴板文本到虚拟机
 * @param text 要发送的文本内容
 */
function sendClipboardText(text: string) {
  if (rfbInstance) {
    rfbInstance.clipboardPasteFrom(text)
  }
}

/**
 * 获取 RFB 实例（供父组件调用）
 */
defineExpose({
  rfbInstance: computed(() => rfbInstance),
  connect,
  disconnect,
  reconnect,
  sendCtrlAltDel,
  sendCtrlAltBackspace,
  setScaleViewport,
  sendClipboardText,
  connectionStatus,
})

// 生命周期钩子
onMounted(() => {
  // 组件挂载后自动连接
  connect()
})

onUnmounted(() => {
  // 组件销毁时断开连接
  disconnect()
})
</script>

<style lang="scss" scoped>
.novnc-console {
  position: relative;
  width: 100%;
  height: 100%;
  background: #000;
  overflow: hidden;

  &.fullscreen {
    position: fixed;
    inset: 0;
    z-index: 9999;
  }
}

.screen-container {
  width: 100%;
  height: 100%;

  // noVNC 渲染的 canvas 样式
  :deep(canvas) {
    display: block;
  }

  // 隐藏 noVNC 默认光标相关元素
  :deep(.noVNC_cursor) {
    display: none;
  }
}

// 状态覆盖层
.status-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.85);
  z-index: 10;
}

.status-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #fff;
  text-align: center;
}

.status-icon {
  font-size: 48px;
  color: #666;

  &.spinning {
    color: #409eff;
    animation: spin 1s linear infinite;
  }

  &.error {
    color: #f56c6c;
  }
}

.status-text {
  font-size: 16px;
  margin: 0;
}

.error-message {
  font-size: 12px;
  color: #f56c6c;
  max-width: 400px;
  margin: 0;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
