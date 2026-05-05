<template>
  <div class="xterm-console" :class="{ 'fullscreen': isFullscreen }">
    <ConsoleToolbar
      :vm-name="`LXC #${vmid}`"
      :vmid="vmid"
      :connection-status="connectionStatus"
      :error="errorMessage"
      vm-type="lxc"
      :is-fullscreen="isFullscreen"
      :is-dark-theme="isDarkTheme"
      @reconnect="reconnect"
      @toggle-fullscreen="toggleFullscreen"
      @toggle-theme="toggleTheme"
      @send-command="sendCommand"
      @upload-file="uploadFile"
    />
    <div ref="terminalRef" class="terminal-container" />
    <div v-if="connectionStatus === 'connecting'" class="terminal-overlay">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>正在连接终端...</span>
    </div>
    <div v-if="connectionStatus === 'error'" class="terminal-overlay error">
      <el-icon><WarningFilled /></el-icon>
      <span>{{ errorMessage || '连接失败' }}</span>
      <el-button type="primary" size="small" @click="reconnect">重新连接</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Loading, WarningFilled } from '@element-plus/icons-vue'
import ConsoleToolbar from './ConsoleToolbar.vue'
import { getLXCTermTicket } from '@/api/console'
import '@xterm/xterm/css/xterm.css'

interface Props {
  node: string
  vmid: number
}

const props = defineProps<Props>()

const terminalRef = ref<HTMLElement>()
const connectionStatus = ref<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected')
const errorMessage = ref('')
const isFullscreen = ref(false)
const isDarkTheme = ref(true)

let termInstance: Terminal | null = null
let fitAddon: FitAddon | null = null
let wsConn: WebSocket | null = null
let pingTimer: number | null = null
let resizeObserver: ResizeObserver | null = null

const lightTheme = {
  background: '#ffffff',
  foreground: '#1e1e1e',
  cursor: '#333333',
  selectionBackground: '#add6ff',
  black: '#000000',
  red: '#cd3131',
  green: '#00bc00',
  yellow: '#949800',
  blue: '#0451a5',
  magenta: '#bc05bc',
  cyan: '#0598bc',
  white: '#555555',
  brightBlack: '#666666',
  brightRed: '#cd3131',
  brightGreen: '#14ce14',
  brightYellow: '#b5ba00',
  brightBlue: '#0451a5',
  brightMagenta: '#bc05bc',
  brightCyan: '#0598bc',
  brightWhite: '#a5a5a5',
}

const darkTheme = {
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
  brightWhite: '#e5e5e5',
}

function arrayBufferToString(buffer: ArrayBuffer): string {
  return new TextDecoder().decode(buffer)
}

function wsDataToString(data: ArrayBuffer | string): string {
  if (typeof data === 'string') {
    return data
  }
  return arrayBufferToString(data)
}

async function connect() {
  if (wsConn) {
    wsConn.close()
    wsConn = null
  }
  if (pingTimer) {
    clearInterval(pingTimer)
    pingTimer = null
  }

  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  try {
    const res = await getLXCTermTicket(props.node, props.vmid)
    const termData = res.term as Record<string, unknown>
    const termTicket = termData.ticket as string
    const termPort = termData.port as number
    const pveAuthCookie = res.PVEAuthCookie as string

    const pveUserName = localStorage.getItem('pve_username') || 'root@pam'
    const pveToken = localStorage.getItem('pve_token') || ''

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/pve/term/websocket?node=${encodeURIComponent(props.node)}&vmid=${props.vmid}&termport=${termPort}&termticket=${encodeURIComponent(termTicket)}&pveauthcookie=${encodeURIComponent(pveAuthCookie)}&username=${encodeURIComponent(pveUserName)}&token=${encodeURIComponent(pveToken)}`

    wsConn = new WebSocket(wsUrl)
    wsConn.binaryType = 'arraybuffer'

    wsConn.onopen = () => {
      if (wsConn) {
        wsConn.send(pveUserName + ':' + termTicket + '\n')
      }
    }

    wsConn.onmessage = (event) => {
      const data = wsDataToString(event.data)
      if (connectionStatus.value === 'connecting') {
        if (data.slice(0, 2) === 'OK') {
          connectionStatus.value = 'connected'
          if (termInstance) {
            termInstance.write(data.slice(2))
            termInstance.focus()
          }
          nextTick(() => {
            if (fitAddon) fitAddon.fit()
            sendResize()
            const enterData = '\r'
            if (wsConn && wsConn.readyState === WebSocket.OPEN) {
              wsConn.send('0:' + enterData.length.toString() + ':' + enterData)
            }
          })
          pingTimer = window.setInterval(() => {
            if (wsConn && wsConn.readyState === WebSocket.OPEN) {
              wsConn.send('2')
            }
          }, 30000)
        }
      } else if (connectionStatus.value === 'connected') {
        if (termInstance) {
          termInstance.write(data)
        }
      }
    }

    wsConn.onerror = () => {
      if (connectionStatus.value === 'connecting') {
        connectionStatus.value = 'error'
        errorMessage.value = 'WebSocket 连接失败'
      }
    }

    wsConn.onclose = () => {
      if (connectionStatus.value === 'connected') {
        connectionStatus.value = 'disconnected'
        if (termInstance) {
          termInstance.write('\r\n\x1b[31m--- 连接已断开 ---\x1b[0m\r\n')
        }
      }
    }
  } catch (err: unknown) {
    connectionStatus.value = 'error'
    errorMessage.value = err instanceof Error ? err.message : '获取终端票据失败'
  }
}

function sendResize() {
  if (wsConn && wsConn.readyState === WebSocket.OPEN && termInstance && connectionStatus.value === 'connected') {
    wsConn.send('1:' + termInstance.cols + ':' + termInstance.rows + ':')
  }
}

function sendCommand(command: string) {
  if (wsConn && wsConn.readyState === WebSocket.OPEN && connectionStatus.value === 'connected') {
    const encoded = unescape(encodeURIComponent(command))
    wsConn.send('0:' + encoded.length.toString() + ':' + command)
    if (termInstance) termInstance.focus()
  }
}

async function uploadFile(file: File, targetPath: string) {
  if (wsConn && wsConn.readyState === WebSocket.OPEN && connectionStatus.value === 'connected') {
    const fileName = file.name
    const targetFilePath = targetPath.endsWith('/')
      ? targetPath + fileName
      : targetPath + '/' + fileName

    if (termInstance) {
      termInstance.write('\r\n\x1b[33m--- 正在上传文件: ' + fileName + ' ---\x1b[0m\r\n')
    }

    const reader = new FileReader()
    reader.onload = async () => {
      const base64Content = (reader.result as string).split(',')[1]
      const chunkSize = 4096
      const totalChunks = Math.ceil(base64Content.length / chunkSize)

      const cmd = `mkdir -p $(dirname '${targetFilePath}') && echo "" > '${targetFilePath}'`
      sendCommand(cmd + '\n')

      await new Promise(resolve => setTimeout(resolve, 500))

      for (let i = 0; i < totalChunks; i++) {
        const chunk = base64Content.slice(i * chunkSize, (i + 1) * chunkSize)
        const appendCmd = `echo '${chunk}' | base64 -d >> '${targetFilePath}'`
        sendCommand(appendCmd + '\n')
        await new Promise(resolve => setTimeout(resolve, 100))
      }

      const verifyCmd = `ls -la '${targetFilePath}'`
      sendCommand(verifyCmd + '\n')

      if (termInstance) {
        termInstance.write('\x1b[33m--- 文件上传完成: ' + targetFilePath + ' ---\x1b[0m\r\n')
      }
    }
    reader.readAsDataURL(file)
  }
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    if (fitAddon) fitAddon.fit()
    sendResize()
  })
}

function toggleTheme() {
  isDarkTheme.value = !isDarkTheme.value
  if (termInstance) {
    termInstance.options.theme = isDarkTheme.value ? darkTheme : lightTheme
  }
}

function reconnect() {
  connect()
}

function initTerminal() {
  if (!terminalRef.value) return

  termInstance = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: "'Cascadia Code', 'Fira Code', 'JetBrains Mono', 'Consolas', monospace",
    theme: isDarkTheme.value ? darkTheme : lightTheme,
    allowProposedApi: true,
    scrollback: 5000,
  })

  fitAddon = new FitAddon()
  termInstance.loadAddon(fitAddon)
  termInstance.loadAddon(new WebLinksAddon())

  termInstance.open(terminalRef.value)
  fitAddon.fit()

  termInstance.onData((data: string) => {
    if (wsConn && wsConn.readyState === WebSocket.OPEN && connectionStatus.value === 'connected') {
      const encoded = unescape(encodeURIComponent(data))
      wsConn.send('0:' + encoded.length.toString() + ':' + data)
    }
  })

  termInstance.onResize(() => {
    sendResize()
  })

  resizeObserver = new ResizeObserver(() => {
    if (fitAddon) fitAddon.fit()
  })
  resizeObserver.observe(terminalRef.value)
}

onMounted(() => {
  initTerminal()
  connect()
})

onBeforeUnmount(() => {
  if (pingTimer) {
    clearInterval(pingTimer)
    pingTimer = null
  }
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
  if (wsConn) {
    wsConn.close()
    wsConn = null
  }
  if (termInstance) {
    termInstance.dispose()
    termInstance = null
  }
})

watch(() => [props.node, props.vmid], () => {
  connect()
})

defineExpose({
  reconnect,
  sendCommand,
  uploadFile,
})
</script>

<style lang="scss" scoped>
.xterm-console {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  overflow: hidden;

  &.fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 9999;
  }
}

.terminal-container {
  flex: 1;
  padding: 4px;
  overflow: hidden;

  :deep(.xterm) {
    height: 100%;
  }
}

.terminal-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #d4d4d4;
  font-size: 14px;

  &.error {
    color: #f14c4c;
  }

  .el-icon {
    font-size: 32px;
  }
}
</style>
