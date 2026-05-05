<template>
  <div class="console-view">
    <!-- 控制台工具栏 -->
    <ConsoleToolbar
      :vm-name="vmName"
      :vmid="vmid"
      :connection-status="connectionStatus"
      :error="errorMessage"
      :vm-type="vmType"
      @send-combo="handleSendCombo"
      @reconnect="handleReconnect"
      @zoom-change="handleZoomChange"
      @clipboard="handleClipboard"
    />

    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-overlay">
      <el-icon class="loading-icon is-loading"><Loading /></el-icon>
      <p>正在加载控制台...</p>
    </div>

    <!-- QEMU 虚拟机使用 noVNC 远程桌面 -->
    <div v-if="!isLoading && vmType === 'qemu'" class="console-body">
      <NoVNCConsole
        ref="novncRef"
        :node="node"
        :vmid="vmid"
        :vm-type="vmType"
      />
    </div>

    <!-- LXC 容器使用 xterm.js 终端 -->
    <div v-if="!isLoading && vmType === 'lxc'" class="console-body">
      <XTermConsole
        ref="xtermRef"
        :node="node"
        :vmid="vmid"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import NoVNCConsole from '@/components/console/NoVNCConsole.vue'
import XTermConsole from '@/components/console/XTermConsole.vue'
import ConsoleToolbar from '@/components/console/ConsoleToolbar.vue'
import { getQEMUConfig } from '@/api/qemu'
import { getLXCConfig } from '@/api/lxc'

const route = useRoute()

/** 路由参数 */
const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const vmType = computed(() => (route.params.vmType as 'qemu' | 'lxc') || 'qemu')

/** noVNC 组件引用 */
const novncRef = ref<InstanceType<typeof NoVNCConsole> | null>(null)

/** xterm 组件引用 */
const xtermRef = ref<InstanceType<typeof XTermConsole> | null>(null)

/** 虚拟机名称 */
const vmName = ref('')

/** 是否加载中 */
const isLoading = ref(true)

/** 错误信息 */
const errorMessage = ref('')

/** 连接状态 */
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatus>('connecting')

/**
 * 加载虚拟机/容器信息
 */
async function loadVMInfo() {
  try {
    if (vmType.value === 'qemu') {
      const config = await getQEMUConfig(node.value, vmid.value)
      vmName.value = config.name || `VM ${vmid.value}`
    } else {
      const config = await getLXCConfig(node.value, vmid.value)
      vmName.value = config.hostname || `CT ${vmid.value}`
    }
  } catch {
    vmName.value = `${vmType.value === 'qemu' ? 'VM' : 'CT'} ${vmid.value}`
  } finally {
    isLoading.value = false
  }
}

/**
 * 处理发送快捷键（仅 QEMU VNC 可用）
 */
function handleSendCombo(combo: 'ctrl-alt-del' | 'ctrl-alt-backspace') {
  if (vmType.value !== 'qemu' || !novncRef.value) return

  if (combo === 'ctrl-alt-del') {
    novncRef.value.sendCtrlAltDel()
  } else if (combo === 'ctrl-alt-backspace') {
    novncRef.value.sendCtrlAltBackspace()
  }
}

/**
 * 处理重新连接
 */
function handleReconnect() {
  connectionStatus.value = 'connecting'
  errorMessage.value = ''

  if (vmType.value === 'qemu' && novncRef.value) {
    novncRef.value.reconnect()
  } else if (vmType.value === 'lxc' && xtermRef.value) {
    xtermRef.value.reconnect()
  }
}

/**
 * 处理缩放级别变更（仅 QEMU VNC 可用）
 */
function handleZoomChange(level: 'auto' | '100' | '50' | '25') {
  if (vmType.value !== 'qemu' || !novncRef.value?.rfbInstance) return

  if (level === 'auto') {
    novncRef.value.setScaleViewport(true)
  } else {
    novncRef.value.setScaleViewport(false)
    const canvas = novncRef.value.rfbInstance?._screen
    if (canvas) {
      const scaleMap: Record<string, number> = {
        '100': 1,
        '50': 0.5,
        '25': 0.25,
      }
      const scale = scaleMap[level]
      if (scale !== undefined) {
        canvas.style.transform = `scale(${scale})`
        canvas.style.transformOrigin = 'top left'
      }
    }
  }
}

/**
 * 处理剪贴板同步
 */
function handleClipboard(text: string) {
  if (vmType.value === 'qemu' && novncRef.value?.rfbInstance) {
    try {
      novncRef.value.rfbInstance.clipboardPasteFrom(text)
    } catch (err) {
      console.error('剪贴板同步失败:', err)
      ElMessage.error('剪贴板同步失败')
    }
  }
}

/**
 * 同步连接状态
 */
function syncConnectionStatus() {
  if (vmType.value === 'qemu' && novncRef.value) {
    connectionStatus.value = novncRef.value.connectionStatus
  } else if (vmType.value === 'lxc' && xtermRef.value) {
    connectionStatus.value = xtermRef.value.connectionStatus
  }
  errorMessage.value = ''
}

/** 轮询同步状态 */
let statusSyncTimer: number | null = null

onMounted(async () => {
  await loadVMInfo()
  statusSyncTimer = window.setInterval(syncConnectionStatus, 500)
})

onUnmounted(() => {
  if (statusSyncTimer !== null) {
    clearInterval(statusSyncTimer)
  }
})
</script>

<style lang="scss" scoped>
.console-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1e1e1e;
  overflow: hidden;
}

.console-body {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: #000;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.85);
  color: #fff;
  z-index: 10;

  .loading-icon {
    font-size: 48px;
    margin-bottom: 16px;
    color: #409eff;
  }

  p {
    font-size: 14px;
    margin: 0;
  }
}
</style>
