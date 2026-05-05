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

      <!-- LXC 终端专用功能 -->
      <template v-if="vmType === 'lxc'">
        <el-button text size="small" @click="showUploadDialog">
          <el-icon><Upload /></el-icon>
          上传文件
        </el-button>

        <el-dropdown @command="handleQuickCommand">
          <el-button text size="small">
            <el-icon><SetUp /></el-icon>
            快捷命令
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="top">top</el-dropdown-item>
              <el-dropdown-item command="htop">htop</el-dropdown-item>
              <el-dropdown-item command="df -h">df -h</el-dropdown-item>
              <el-dropdown-item command="free -h">free -h</el-dropdown-item>
              <el-dropdown-item command="ifconfig">ifconfig</el-dropdown-item>
              <el-dropdown-item command="ip addr">ip addr</el-dropdown-item>
              <el-dropdown-item command="ps aux">ps aux</el-dropdown-item>
              <el-dropdown-item command="uptime">uptime</el-dropdown-item>
              <el-dropdown-item command="uname -a">uname -a</el-dropdown-item>
              <el-dropdown-item command="cat /etc/os-release">cat /etc/os-release</el-dropdown-item>
              <el-dropdown-item divided command="custom">自定义命令...</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <el-button text size="small" @click="toggleTerminalTheme">
          <el-icon><Sunny v-if="isDarkTheme" /><Moon v-else /></el-icon>
          {{ isDarkTheme ? '浅色' : '深色' }}
        </el-button>
      </template>

      <!-- 通用功能 -->
      <el-button text size="small" @click="$emit('toggle-fullscreen')">
        <el-icon><FullScreen /></el-icon>
        {{ isFullscreen ? '退出全屏' : '全屏' }}
      </el-button>

      <el-button text size="small" @click="$emit('reconnect')">
        <el-icon><RefreshRight /></el-icon>
        重连
      </el-button>
    </div>

    <!-- 剪贴板对话框（QEMU） -->
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

    <!-- 文件上传对话框（LXC） -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传文件到容器"
      width="500px"
      :append-to-body="true"
    >
      <el-form label-position="top">
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            drag
          >
            <el-icon class="el-icon--upload"><Upload /></el-icon>
            <div class="el-upload__text">拖拽文件到此处或 <em>点击上传</em></div>
          </el-upload>
        </el-form-item>
        <el-form-item label="目标路径">
          <el-input v-model="uploadTargetPath" placeholder="/root/" />
        </el-form-item>
        <el-alert
          v-if="uploadFile"
          :title="`文件大小: ${formatFileSize(uploadFile.size)}`"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 12px"
        />
        <el-alert
          v-if="uploadFile && uploadFile.size > 512 * 1024"
          title="大文件上传可能较慢，建议使用 SCP/SFTP 传输大文件"
          type="warning"
          :closable="false"
          show-icon
        />
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!uploadFile"
          :loading="uploading"
          @click="handleUpload"
        >
          {{ uploading ? '上传中...' : '上传' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 自定义命令对话框 -->
    <el-dialog
      v-model="customCommandVisible"
      title="发送自定义命令"
      width="400px"
      :append-to-body="true"
    >
      <el-input
        v-model="customCommand"
        placeholder="输入要执行的命令..."
        @keyup.enter="sendCustomCommand"
      />
      <template #footer>
        <el-button @click="customCommandVisible = false">取消</el-button>
        <el-button type="primary" @click="sendCustomCommand">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ArrowLeft, SetUp, FullScreen, DocumentCopy, RefreshRight,
  Upload, Sunny, Moon,
} from '@element-plus/icons-vue'
import type { UploadFile as ElUploadFile } from 'element-plus'

interface Props {
  vmName: string
  vmid: number
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'error'
  error?: string
  vmType?: 'qemu' | 'lxc'
  isFullscreen?: boolean
  isDarkTheme?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  error: '',
  vmType: 'qemu',
  isFullscreen: false,
  isDarkTheme: true,
})

const emit = defineEmits<{
  'send-combo': [combo: 'ctrl-alt-del' | 'ctrl-alt-backspace']
  'reconnect': []
  'zoom-change': [level: 'auto' | '100' | '50' | '25']
  'clipboard': [text: string]
  'toggle-fullscreen': []
  'toggle-theme': []
  'send-command': [command: string]
  'upload-file': [file: File, targetPath: string]
}>()

const clipboardVisible = ref(false)
const clipboardText = ref('')
const uploadDialogVisible = ref(false)
const customCommandVisible = ref(false)
const customCommand = ref('')
const uploading = ref(false)
const uploadFile = ref<File | null>(null)
const uploadTargetPath = ref('/root/')

const statusTagType = computed(() => {
  switch (props.connectionStatus) {
    case 'connected': return 'success'
    case 'connecting': return 'warning'
    case 'error': return 'danger'
    default: return 'info'
  }
})

const statusLabel = computed(() => {
  switch (props.connectionStatus) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中'
    case 'error': return '连接失败'
    default: return '未连接'
  }
})

function handleCombo(command: string) {
  if (command === 'ctrl-alt-del' || command === 'ctrl-alt-backspace') {
    emit('send-combo', command)
  }
}

function handleZoom(command: string) {
  if (['auto', '100', '50', '25'].includes(command)) {
    emit('zoom-change', command as 'auto' | '100' | '50' | '25')
  }
}

function showClipboard() {
  clipboardVisible.value = true
}

function sendClipboard() {
  if (clipboardText.value) {
    emit('clipboard', clipboardText.value)
  }
  clipboardVisible.value = false
}

function showUploadDialog() {
  uploadFile.value = null
  uploadTargetPath.value = '/root/'
  uploadDialogVisible.value = true
}

function handleFileChange(file: ElUploadFile) {
  if (file.raw) {
    uploadFile.value = file.raw
  }
}

function handleFileRemove() {
  uploadFile.value = null
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

async function handleUpload() {
  if (!uploadFile.value) return
  uploading.value = true
  try {
    emit('upload-file', uploadFile.value, uploadTargetPath.value)
  } finally {
    uploading.value = false
    uploadDialogVisible.value = false
  }
}

function handleQuickCommand(command: string) {
  if (command === 'custom') {
    customCommand.value = ''
    customCommandVisible.value = true
  } else {
    emit('send-command', command + '\n')
  }
}

function sendCustomCommand() {
  if (customCommand.value.trim()) {
    emit('send-command', customCommand.value.trim() + '\n')
  }
  customCommandVisible.value = false
}

function toggleTerminalTheme() {
  emit('toggle-theme')
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
