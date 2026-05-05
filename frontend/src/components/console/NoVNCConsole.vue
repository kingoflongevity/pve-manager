<template>
  <div class="novnc-console" :class="{ fullscreen: isFullscreen }">
    <ConsoleToolbar
      :vm-name="`VM #${vmid}`"
      :vmid="vmid"
      :connection-status="connectionStatus"
      :error="errorMessage"
      vm-type="qemu"
      :is-fullscreen="isFullscreen"
      @send-combo="handleCombo"
      @reconnect="reconnect"
      @zoom-change="handleZoomChange"
      @clipboard="handleClipboard"
      @toggle-fullscreen="toggleFullscreen"
    />

    <div class="screen-wrapper">
      <div v-if="connectionStatus !== 'connected'" class="status-overlay">
        <div class="status-content">
          <el-icon v-if="connectionStatus === 'connecting'" class="status-icon spinning">
            <Loading />
          </el-icon>
          <el-icon v-else-if="connectionStatus === 'error'" class="status-icon error">
            <CircleClose />
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
      <div ref="screenContainer" class="screen-container" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Loading, CircleClose } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import RFB from '@novnc/novnc/core/rfb'
import ConsoleToolbar from './ConsoleToolbar.vue'
import { getQEMUVNCTicket } from '@/api/console'

interface Props {
  node: string
  vmid: number
  vmType: 'qemu' | 'lxc'
}

const props = defineProps<Props>()

let rfbInstance: RFB | null = null

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatus>('disconnected')
const errorMessage = ref('')
const screenContainer = ref<HTMLElement | null>(null)
const isFullscreen = ref(false)

const statusText = computed(() => {
  const map: Record<ConnectionStatus, string> = {
    disconnected: '未连接',
    connecting: '正在连接...',
    connected: '',
    error: '连接失败',
  }
  return map[connectionStatus.value]
})

function buildWebSocketUrl(vncPort: number, vncTicket: string, pveAuthCookie: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const token = localStorage.getItem('pve_token') || ''

  const params = new URLSearchParams({
    vncport: String(vncPort),
    vncticket: vncTicket,
    token: token,
    node: props.node,
    vmid: String(props.vmid),
    vmtype: props.vmType,
    pveauthcookie: pveAuthCookie,
  })

  return `${protocol}//${host}/api/pve/vnc/websocket?${params.toString()}`
}

async function connect() {
  if (!screenContainer.value) {
    ElMessage.error('控制台容器未就绪')
    return
  }

  disconnect()

  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  try {
    const proxyResponse = await getQEMUVNCTicket(props.node, props.vmid)
    const vncData = proxyResponse.vnc

    if (!vncData.ticket || !vncData.port) {
      throw new Error('无效的 VNC 票据')
    }

    if (!proxyResponse.PVEAuthCookie) {
      throw new Error('缺少 PVEAuthCookie')
    }

    const wsUrl = buildWebSocketUrl(vncData.port, vncData.ticket, proxyResponse.PVEAuthCookie)

    rfbInstance = new RFB(screenContainer.value, wsUrl, {
      credentials: { password: vncData.ticket },
      clipViewport: false,
      viewOnly: false,
      dragViewport: false,
      scaleViewport: true,
    })

    rfbInstance.addEventListener('connect', onConnect)
    rfbInstance.addEventListener('disconnect', onDisconnect)
    rfbInstance.addEventListener('credentialsrequired', onCredentialsRequired)
    rfbInstance.addEventListener('securityfailure', onSecurityFailure)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : '未知错误'
    connectionStatus.value = 'error'
    errorMessage.value = message
    ElMessage.error(`VNC 连接失败: ${message}`)
  }
}

function disconnect() {
  if (rfbInstance) {
    try {
      rfbInstance.removeEventListener('connect', onConnect)
      rfbInstance.removeEventListener('disconnect', onDisconnect)
      rfbInstance.removeEventListener('credentialsrequired', onCredentialsRequired)
      rfbInstance.removeEventListener('securityfailure', onSecurityFailure)
      rfbInstance.disconnect()
    } catch {
      // ignore
    }
    rfbInstance = null
  }
  connectionStatus.value = 'disconnected'
}

function reconnect() {
  connect()
}

function onConnect() {
  connectionStatus.value = 'connected'
  errorMessage.value = ''
}

function onDisconnect(e: CustomEvent) {
  connectionStatus.value = 'disconnected'
  const detail = e.detail
  if (detail && detail.clean === false) {
    connectionStatus.value = 'error'
    errorMessage.value = detail.reason || '连接异常断开'
  }
}

function onCredentialsRequired() {
  // noVNC 会自动处理凭证
}

function onSecurityFailure(e: CustomEvent) {
  connectionStatus.value = 'error'
  errorMessage.value = e.detail?.reason || 'VNC 安全认证失败'
  ElMessage.error('VNC 安全认证失败')
}

function sendKeyCombo(keys: number[]) {
  if (!rfbInstance) return
  for (const code of keys) {
    rfbInstance.sendKey(code, '', true)
  }
  for (let i = keys.length - 1; i >= 0; i--) {
    rfbInstance.sendKey(keys[i], '', false)
  }
}

function handleCombo(combo: 'ctrl-alt-del' | 'ctrl-alt-backspace') {
  if (combo === 'ctrl-alt-del') {
    sendKeyCombo([0xffe3, 0xffe9, 0xffff])
    ElMessage.info('已发送 Ctrl+Alt+Del')
  } else if (combo === 'ctrl-alt-backspace') {
    sendKeyCombo([0xffe3, 0xffe9, 0xff08])
    ElMessage.info('已发送 Ctrl+Alt+Backspace')
  }
}

function handleZoomChange(level: 'auto' | '100' | '50' | '25') {
  if (!rfbInstance) return
  if (level === 'auto') {
    rfbInstance.scaleViewport = true
  } else {
    rfbInstance.scaleViewport = false
    const scaleMap: Record<string, number> = { '100': 1, '50': 0.5, '25': 0.25 }
    const scale = scaleMap[level]
    if (scale !== undefined && rfbInstance._screen) {
      rfbInstance._screen.style.transform = `scale(${scale})`
      rfbInstance._screen.style.transformOrigin = 'top left'
    }
  }
}

function handleClipboard(text: string) {
  if (rfbInstance) {
    try {
      rfbInstance.clipboardPasteFrom(text)
    } catch (err) {
      console.error('剪贴板同步失败:', err)
      ElMessage.error('剪贴板同步失败')
    }
  }
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
}

defineExpose({
  rfbInstance: computed(() => rfbInstance),
  connect,
  disconnect,
  reconnect,
  connectionStatus,
})

onMounted(() => {
  connect()
})

onUnmounted(() => {
  disconnect()
})
</script>

<style lang="scss" scoped>
.novnc-console {
  display: flex;
  flex-direction: column;
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

.screen-wrapper {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.screen-container {
  width: 100%;
  height: 100%;

  :deep(canvas) {
    display: block;
  }

  :deep(.noVNC_cursor) {
    display: none;
  }
}

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
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
