<template>
  <div class="ai-chat">
    <div class="chat-header">
      <div class="header-left">
        <h2>AI 智能对话</h2>
        <el-select v-model="currentScene" size="small" style="width: 160px; margin-left: 12px">
          <el-option label="运维故障排查" value="ops_troubleshoot" />
          <el-option label="配置建议" value="config_advice" />
          <el-option label="监控分析" value="monitoring" />
          <el-option label="自由对话" value="general" />
        </el-select>
      </div>
      <div class="header-right">
        <el-button size="small" @click="clearConversation">新建对话</el-button>
        <el-select v-model="selectedModel" size="small" style="width: 200px; margin-left: 8px">
          <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
        </el-select>
      </div>
    </div>

    <div class="chat-messages" ref="messagesRef">
      <div v-if="!messages.length" class="empty-state">
        <el-icon :size="48"><ChatLineSquare /></el-icon>
        <p>您好，我是 AI 运维助手，有什么可以帮助您的？</p>
        <div class="quick-actions">
          <el-button size="small" @click="sendQuickMessage('请帮我检查系统运行状态')">检查系统状态</el-button>
          <el-button size="small" @click="sendQuickMessage('最近有哪些异常日志？')">查看异常日志</el-button>
          <el-button size="small" @click="sendQuickMessage('如何优化虚拟机资源配置？')">优化建议</el-button>
        </div>
      </div>
      <div v-for="(msg, idx) in messages" :key="idx" :class="['message', msg.role]">
        <div class="message-avatar">
          <el-icon v-if="msg.role === 'user'"><User /></el-icon>
          <el-icon v-else><ChatDotRound /></el-icon>
        </div>
        <div class="message-content">
          <div class="message-bubble" v-html="renderMarkdown(msg.content)"></div>
          <div class="message-time">{{ formatTime(msg.created_at) }}</div>
        </div>
      </div>
      <div v-if="loading" class="message assistant">
        <div class="message-avatar"><el-icon><ChatDotRound /></el-icon></div>
        <div class="message-content">
          <div class="message-bubble typing">
            <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
          </div>
        </div>
      </div>
    </div>

    <div class="chat-input">
      <el-input
        v-model="inputMessage"
        type="textarea"
        :rows="3"
        placeholder="输入您的问题，例如：检查节点 pve1 的状态"
        :disabled="loading"
        @keydown.enter.ctrl="sendMessage"
      />
      <div class="input-actions">
        <span class="input-hint">Ctrl + Enter 发送</span>
        <el-button type="primary" :loading="loading" @click="sendMessage">发送</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { ChatLineSquare, ChatDotRound, User } from '@element-plus/icons-vue'
import { getConversations, createConversation, getAIModels } from '@/api/ai'
import type { AIModelConfig } from '@/api/ai'

const messagesRef = ref<HTMLElement>()
const inputMessage = ref('')
const messages = ref<Array<{ role: string; content: string; created_at: string }>>([])
const loading = ref(false)
const currentScene = ref('ops_troubleshoot')
const selectedModel = ref<number>(0)
const models = ref<AIModelConfig[]>([])

onMounted(async () => {
  try {
    models.value = await getAIModels()
    if (models.value.length > 0) {
      const defaultModel = models.value.find(m => m.is_default) || models.value[0]
      selectedModel.value = defaultModel.id
    }
  } catch (e) {
    console.error('获取模型列表失败', e)
  }
})

async function sendMessage() {
  const content = inputMessage.value.trim()
  if (!content || loading.value) return

  messages.value.push({ role: 'user', content, created_at: new Date().toISOString() })
  inputMessage.value = ''
  loading.value = true

  try {
    const resp = await createConversation({
      scene: currentScene.value,
      model_config_id: selectedModel.value || undefined,
      message: content,
    })
    messages.value.push({
      role: 'assistant',
      content: resp.content || '处理完成',
      created_at: new Date().toISOString(),
    })
  } catch (e: any) {
    messages.value.push({
      role: 'assistant',
      content: '请求失败: ' + (e.message || '未知错误'),
      created_at: new Date().toISOString(),
    })
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

function sendQuickMessage(msg: string) {
  inputMessage.value = msg
  sendMessage()
}

function clearConversation() {
  messages.value = []
}

function renderMarkdown(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\n/g, '<br>')
    .replace(/```(\w*)\n?([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
}

function formatTime(t: string): string {
  if (!t) return ''
  return new Date(t).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}
</script>

<style scoped>
.ai-chat { display: flex; flex-direction: column; height: 100%; background: #fff; border-radius: 8px; }
.chat-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #ebeef5; }
.header-left, .header-right { display: flex; align-items: center; }
.header-left h2 { font-size: 18px; font-weight: 600; margin: 0; }
.chat-messages { flex: 1; overflow-y: auto; padding: 20px; }
.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: #909399; gap: 16px; }
.empty-state p { font-size: 16px; margin: 0; }
.quick-actions { display: flex; gap: 8px; }
.message { display: flex; gap: 12px; margin-bottom: 20px; }
.message.user { flex-direction: row-reverse; }
.message-avatar { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; background: #e6f7ff; flex-shrink: 0; }
.message.user .message-avatar { background: #e8f5e9; }
.message-content { max-width: 70%; }
.message-bubble { padding: 12px 16px; border-radius: 12px; font-size: 14px; line-height: 1.6; background: #f5f7fa; }
.message.user .message-bubble { background: #409eff; color: #fff; border-bottom-right-radius: 4px; }
.message.assistant .message-bubble { background: #f5f7fa; border-bottom-left-radius: 4px; }
.message-time { font-size: 12px; color: #c0c4cc; margin-top: 4px; text-align: right; }
.typing .dot { animation: blink 1.4s infinite; font-size: 24px; line-height: 0; }
.typing .dot:nth-child(2) { animation-delay: 0.2s; }
.typing .dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes blink { 0%, 80%, 100% { opacity: 0; } 40% { opacity: 1; } }
.chat-input { padding: 16px 20px; border-top: 1px solid #ebeef5; }
.input-actions { display: flex; justify-content: space-between; align-items: center; margin-top: 8px; }
.input-hint { font-size: 12px; color: #c0c4cc; }
pre { background: #282c34; color: #abb2bf; padding: 12px; border-radius: 6px; overflow-x: auto; margin: 8px 0; }
code { background: #f0f2f5; padding: 2px 6px; border-radius: 4px; font-size: 13px; }
pre code { background: none; padding: 0; }
</style>
