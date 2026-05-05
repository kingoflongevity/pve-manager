<template>
  <div class="xterm-console" :class="{ fullscreen: isFullscreen }">
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

    <!-- xterm.js 终端容器 -->
    <div ref="terminalContainer" class="terminal-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Loading, CircleClose, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { getLXCTermTicket } from '@/api/console'

/** 终端组件属性 */
interface Props {
  /** 节点名称 */
  node: string
  /** 容器 ID */
  vmid: number
}

const props = defineProps<Props>()

/** xterm.js 终端实例 */
let termInstance: Terminal | null = null

/** FitAddon 实例 */
let fitAddon: FitAddon | null = null

/** WebSocket 连接 */
let wsConn: WebSocket | null = null

/** 心跳定时器 */
let pingTimer: number | null = null

/** WebSocket 连接状态 */
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatus>('disconnected')

/** 错误信息 */
const errorMessage = ref('')

/** 容器 DOM */
const terminalContainer = ref<HTMLElement | null>(null)

/** 是否全屏 */
const isFullscreen = ref(false)

/** ResizeObserver */
let resizeObserver: ResizeObserver | null = null

/** 终端票据（用于二次认证） */
let termTicket = ''

/** PVE 用户名（用于二次认证） */
let pveUserName = ''

/**
 * 获取状态显示文本
 */
const statusText = computed(() => {
  const map: Record<ConnectionStatus, string> = {
    disconnected: '未连接',
    connecting: '正在连接终端...',
    connected: '',
    error: '终端连接失败',
  }
  return map[connectionStatus.value]
})

/**
 * 构建 WebSocket URL
 * 通过后端代理转发 WebSocket 连接到 PVE termproxy
 */
function buildWebSocketUrl(termPort: number, termTicket: string, pveAuthCookie: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const token = localStorage.getItem('pve_token') || ''

  const params = new URLSearchParams({
    termport: String(termPort),
    termticket: termTicket,
    token: token,
    node: props.node,
    vmid: String(props.vmid),
    pveauthcookie: pveAuthCookie,
  })

  return `${protocol}//${host}/api/pve/term/websocket?${params.toString()}`
}

/**
 * 将 ArrayBuffer 转换为字符串
 */
function arrayBufferToString(buf: ArrayBuffer): string {
  const arr = new Uint8Array(buf)
  let str = ''
  for (let i = 0; i < arr.length; i++) {
    str += String.fromCharCode(arr[i])
  }
  return str
}

/**
 * 连接终端
 * 流程：
 * 1. 调用后端 API 获取 termproxy 票据和 PVEAuthCookie
 * 2. 构建 WebSocket URL 并连接后端代理
 * 3. WebSocket 连接建立后发送认证信息（username:ticket\n）
 * 4. 等待 "OK" 响应表示认证成功
 * 5. 初始化 xterm.js 终端并绑定数据收发
 *
 * PVE termproxy 协议：
 * - 认证：发送 "username:ticket\n"，等待 "OK" 响应
 * - 终端输入：发送 "0:<length>:<data>"
 * - 终端尺寸：发送 "1:<cols>:<rows>:"
 * - 心跳：发送 "2"（每30秒）
 * - 终端输出：直接写入 xterm.js
 */
async function connect() {
  if (!terminalContainer.value) {
    ElMessage.error('终端容器未就绪')
    return
  }

  disconnect()

  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  try {
    // 获取终端代理票据
    const proxyResponse = await getLXCTermTicket(props.node, props.vmid)
    const termData = proxyResponse.term

    if (!termData.ticket || !termData.port) {
      throw new Error('无效的终端票据')
    }

    if (!proxyResponse.PVEAuthCookie) {
      throw new Error('缺少 PVEAuthCookie')
    }

    // 保存票据用于二次认证
    termTicket = termData.ticket

    // 从 PVEAuthCookie 中提取用户名
    // PVEAuthCookie 格式: PVE:user@realm:timestamp::hash
    const cookieStr = decodeURIComponent(proxyResponse.PVEAuthCookie)
    const cookieMatch = cookieStr.match(/^PVE:([^:]+):/)
    pveUserName = cookieMatch ? cookieMatch[1] : 'root@pam'

    // 初始化 xterm.js 终端
    termInstance = new Terminal({
      cursorBlink: true,
      cursorStyle: 'block',
      fontSize: 14,
      fontFamily: '"Cascadia Code", "Fira Code", "JetBrains Mono", Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        selectionBackground: '#264f78',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#ffffff',
      },
      allowProposedApi: true,
      scrollback: 1000,
    })

    fitAddon = new FitAddon()
    termInstance.loadAddon(fitAddon)
    termInstance.loadAddon(new WebLinksAddon())

    termInstance.open(terminalContainer.value)

    setTimeout(() => {
      fitAddon?.fit()
    }, 100)

    // 构建 WebSocket URL 并连接
    const wsUrl = buildWebSocketUrl(termData.port, termData.ticket, proxyResponse.PVEAuthCookie)
    wsConn = new WebSocket(wsUrl)
    wsConn.binaryType = 'arraybuffer'

    wsConn.onopen = () => {
      connectionStatus.value = 'connecting'
      // PVE termproxy 二次认证：发送 "username:ticket\n"
      if (wsConn && wsConn.readyState === WebSocket.OPEN) {
        wsConn.send(pveUserName + ':' + termTicket + '\n')
      }
    }

    wsConn.onmessage = (event) => {
      if (!termInstance) return

      const data = arrayBufferToString(event.data)

      if (connectionStatus.value === 'connecting') {
        // 等待 "OK" 响应表示认证成功
        if (data.slice(0, 2) === 'OK') {
          connectionStatus.value = 'connected'
          errorMessage.value = ''
          // 写入 "OK" 之后的数据
          termInstance.write(data.slice(2))
          termInstance.focus()

          // 认证成功后立即发送终端尺寸
          if (fitAddon) {
            fitAddon.fit()
          }
          if (termInstance && wsConn && wsConn.readyState === WebSocket.OPEN) {
            wsConn.send('1:' + termInstance.cols + ':' + termInstance.rows + ':')
            // 发送回车触发登录提示
            const enterData = '\r'
            const encoded = unescape(encodeURIComponent(enterData))
            wsConn.send('0:' + encoded.length.toString() + ':' + enterData)
          }

          // 启动心跳
          pingTimer = window.setInterval(() => {
            if (wsConn && wsConn.readyState === WebSocket.OPEN) {
              wsConn.send('2')
            }
          }, 30000)
        } else {
          // 认证失败
          connectionStatus.value = 'error'
          errorMessage.value = '终端认证失败: ' + data
        }
      } else if (connectionStatus.value === 'connected') {
        // 正常终端输出
        termInstance.write(data)
      }
    }

    wsConn.onclose = (event) => {
      if (pingTimer !== null) {
        clearInterval(pingTimer)
        pingTimer = null
      }

      if (connectionStatus.value === 'connected') {
        connectionStatus.value = 'disconnected'
      } else {
        connectionStatus.value = 'error'
        errorMessage.value = '终端连接异常断开'
      }
    }

    wsConn.onerror = () => {
      connectionStatus.value = 'error'
      errorMessage.value = '终端连接失败'
    }

    // 终端输入转发到 WebSocket
    // PVE termproxy 格式: "0:<length>:<data>"
    termInstance.onData((data: string) => {
      if (wsConn && wsConn.readyState === WebSocket.OPEN && connectionStatus.value === 'connected') {
        const encoded = unescape(encodeURIComponent(data))
        wsConn.send('0:' + encoded.length.toString() + ':' + data)
      }
    })

    // 终端尺寸变更
    // PVE termproxy 格式: "1:<cols>:<rows>:"
    termInstance.onResize(({ cols, rows }) => {
      if (wsConn && wsConn.readyState === WebSocket.OPEN && connectionStatus.value === 'connected') {
        wsConn.send('1:' + cols + ':' + rows + ':')
      }
    })
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : '未知错误'
    connectionStatus.value = 'error'
    errorMessage.value = message
    ElMessage.error(`终端连接失败: ${message}`)
  }
}

/**
 * 断开终端连接
 */
function disconnect() {
  if (pingTimer !== null) {
    clearInterval(pingTimer)
    pingTimer = null
  }

  if (wsConn) {
    try {
      wsConn.close()
    } catch {
      // 忽略关闭错误
    }
    wsConn = null
  }

  if (termInstance) {
    try {
      termInstance.dispose()
    } catch {
      // 忽略销毁错误
    }
    termInstance = null
    fitAddon = null
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
 * 聚焦终端
 */
function focus() {
  termInstance?.focus()
}

/**
 * 适配终端尺寸
 */
function fit() {
  fitAddon?.fit()
}

/**
 * 获取终端实例（供父组件调用）
 */
defineExpose({
  termInstance: computed(() => termInstance),
  connectionStatus,
  connect,
  disconnect,
  reconnect,
  focus,
  fit,
})

onMounted(() => {
  connect()

  if (terminalContainer.value) {
    resizeObserver = new ResizeObserver(() => {
      fitAddon?.fit()
    })
    resizeObserver.observe(terminalContainer.value)
  }
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  disconnect()
})
</script>

<style lang="scss" scoped>
.xterm-console {
  position: relative;
  width: 100%;
  height: 100%;
  background: #1e1e1e;
  overflow: hidden;

  &.fullscreen {
    position: fixed;
    inset: 0;
    z-index: 9999;
  }
}

.terminal-container {
  width: 100%;
  height: 100%;
  padding: 4px;

  :deep(.xterm) {
    height: 100%;
  }

  :deep(.xterm-viewport) {
    overflow-y: auto !important;
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
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
