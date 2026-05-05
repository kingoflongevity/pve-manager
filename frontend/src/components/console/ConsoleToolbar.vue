<template>
  <div class="console-toolbar">
    <div class="toolbar-left">
      <el-button text @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <span class="vm-name">{{ vmName }}</span>
      <el-tag
        :type="statusTagType"
        size="small"
        class="status-tag"
      >
        {{ statusLabel }}
      </el-tag>
    </div>

    <div class="toolbar-right">
      <!-- QEMU VNC 专用功能 -->
      <template v-if="vmType === 'qemu'">
        <el-dropdown @command="handleCombo">
          <el-button text size="small">
            <el-icon><SetUp /></el-icon>
            快捷键
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="ctrl-alt-del">Ctrl+Alt+Del</el-dropdown-item>
              <el-dropdown-item command="ctrl-alt-backspace">Ctrl+Alt+Backspace</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <el-dropdown @command="handleZoom">
          <el-button text size="small">
            <el-icon><FullScreen /></el-icon>
            缩放
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="auto">自适应</el-dropdown-item>
              <el-dropdown-item command="100">100%</el-dropdown-item>
              <el-dropdown-item command="50">50%</el-dropdown-item>
              <el-dropdown-item command="25">25%</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <el-button text size="small" @click="showClipboard">
          <el-icon><DocumentCopy /></el-icon>
          剪贴板
        </el-button>
      </template>

      <!-- 通用功能 -->
      <el-button text size="small" @click="$emit('reconnect')">
        <el-icon><RefreshRight /></el-icon>
        重连
      </el-button>
    </div>

    <!-- 剪贴板对话框 -->
    <el-dialog
      v-model="clipboardVisible"
      title="剪贴板"
      width="400px"
      :append-to-body="true"
    >
      <el-input
        v-model="clipboardText"
        type="textarea"
        :rows="6"
        placeholder="输入文本发送到虚拟机剪贴板..."
      />
      <template #footer>
        <el-button @click="clipboardVisible = false">取消</el-button>
        <el-button type="primary" @click="sendClipboard">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ArrowLeft, SetUp, FullScreen, DocumentCopy, RefreshRight } from '@element-plus/icons-vue'

/** 工具栏属性 */
interface Props {
  /** 虚拟机/容器名称 */
  vmName: string
  /** 虚拟机/容器 ID */
  vmid: number
  /** 连接状态 */
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'error'
  /** 错误信息 */
  error?: string
  /** 虚拟机类型 */
  vmType?: 'qemu' | 'lxc'
}

const props = withDefaults(defineProps<Props>(), {
  error: '',
  vmType: 'qemu',
})

const emit = defineEmits<{
  'send-combo': [combo: 'ctrl-alt-del' | 'ctrl-alt-backspace']
  'reconnect': []
  'zoom-change': [level: 'auto' | '100' | '50' | '25']
  'clipboard': [text: string]
}>()

/** 剪贴板对话框可见性 */
const clipboardVisible = ref(false)

/** 剪贴板文本 */
const clipboardText = ref('')

/**
 * 状态标签类型
 */
const statusTagType = computed(() => {
  switch (props.connectionStatus) {
    case 'connected':
      return 'success'
    case 'connecting':
      return 'warning'
    case 'error':
      return 'danger'
    default:
      return 'info'
  }
})

/**
 * 状态标签文本
 */
const statusLabel = computed(() => {
  switch (props.connectionStatus) {
    case 'connected':
      return '已连接'
    case 'connecting':
      return '连接中'
    case 'error':
      return '连接失败'
    default:
      return '未连接'
  }
})

/**
 * 处理快捷键命令
 */
function handleCombo(command: string) {
  if (command === 'ctrl-alt-del' || command === 'ctrl-alt-backspace') {
    emit('send-combo', command)
  }
}

/**
 * 处理缩放命令
 */
function handleZoom(command: string) {
  if (['auto', '100', '50', '25'].includes(command)) {
    emit('zoom-change', command as 'auto' | '100' | '50' | '25')
  }
}

/**
 * 显示剪贴板对话框
 */
function showClipboard() {
  clipboardVisible.value = true
}

/**
 * 发送剪贴板内容
 */
function sendClipboard() {
  if (clipboardText.value) {
    emit('clipboard', clipboardText.value)
  }
  clipboardVisible.value = false
}
</script>

<style lang="scss" scoped>
.console-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 12px;
  background: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  min-height: 40px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.vm-name {
  color: #e0e0e0;
  font-size: 14px;
  font-weight: 500;
}

.status-tag {
  margin-left: 4px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
