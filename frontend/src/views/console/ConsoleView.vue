<template>
  <div class="console-view">
    <!-- 控制台工具栏 -->
    <ConsoleToolbar
      :vm-name="vmName"
      :vmid="vmid"
      :connection-status="connectionStatus"
      :error="errorMessage"
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

    <!-- noVNC 控制台 -->
    <div v-show="!isLoading" class="console-body">
      <NoVNCConsole
        ref="novncRef"
        :node="node"
        :vmid="vmid"
        :vm-type="vmType"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import NoVNCConsole from '@/components/console/NoVNCConsole.vue'
import ConsoleToolbar from '@/components/console/ConsoleToolbar.vue'
import { getQEMUConfig } from '@/api/qemu'
import { getLXCConfig } from '@/api/lxc'

const router = useRouter()
const route = useRoute()

/** 路由参数 */
const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const vmType = computed(() => (route.params.vmType as 'qemu' | 'lxc') || 'qemu')

/** noVNC 组件引用 */
const novncRef = ref<InstanceType<typeof NoVNCConsole> | null>(null)

/** 虚拟机名称 */
const vmName = ref('')

/** 是否加载中 */
const isLoading = ref(true)

/** 错误信息 */
const errorMessage = ref('')

/** 连接状态 */
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'
const connectionStatus = ref<ConnectionStatus>('connecting')

/** 当前缩放级别 */
const currentZoom = ref<'auto' | '100' | '50' | '25'>('auto')

/**
 * 加载虚拟机信息
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
 * 处理发送快捷键
 */
function handleSendCombo(combo: 'ctrl-alt-del' | 'ctrl-alt-backspace') {
  if (!novncRef.value) return

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
  if (!novncRef.value) return
  connectionStatus.value = 'connecting'
  errorMessage.value = ''
  novncRef.value.reconnect()
}

/**
 * 处理缩放级别变更
 */
function handleZoomChange(level: 'auto' | '100' | '50' | '25') {
  currentZoom.value = level

  if (!novncRef.value?.rfbInstance) return

  // 自适应模式启用缩放
  if (level === 'auto') {
    novncRef.value.setScaleViewport(true)
  } else {
    novncRef.value.setScaleViewport(false)
    // TODO: 设置具体缩放比例（noVNC 通过 CSS transform 实现）
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
 * 将文本内容发送到虚拟机剪贴板
 */
function handleClipboard(text: string) {
  if (!novncRef.value?.rfbInstance) return

  try {
    // 通过 noVNC 的剪贴板 API 发送文本
    novncRef.value.rfbInstance.clipboardPasteFrom(text)
  } catch (err) {
    console.error('剪贴板同步失败:', err)
    ElMessage.error('剪贴板同步失败')
  }
}

/**
 * 同步 noVNC 连接状态到父组件
 */
function syncConnectionStatus() {
  if (novncRef.value) {
    connectionStatus.value = novncRef.value.connectionStatus
    errorMessage.value = ''
  }
}

/** 轮询同步状态 */
let statusSyncTimer: number | null = null

onMounted(async () => {
  await loadVMInfo()

  // 定时同步连接状态
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

// 加载覆盖层
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
