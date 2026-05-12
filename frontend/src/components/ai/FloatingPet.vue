<template>
  <div class="floating-pet-container">
    <transition name="pet-bounce">
      <div
        v-if="!chatOpen"
        class="pet-button"
        @click="openChat"
        @mousedown="startDrag"
        :style="petPosition"
        :class="{ dragging: isDragging }"
        :title="petTooltip"
      >
        <!-- 猫咪身体 -->
        <div class="cat-container">
          <!-- 尾巴 -->
          <div class="tail"></div>
          <!-- 身体 -->
          <div class="body">
            <div class="belly"></div>
          </div>
          <!-- 头 -->
          <div class="head">
            <!-- 耳朵 -->
            <div class="ear left-ear">
              <div class="ear-inner"></div>
            </div>
            <div class="ear right-ear">
              <div class="ear-inner"></div>
            </div>
            <!-- 脸部 -->
            <div class="face">
              <!-- 眼睛 -->
              <div class="eyes">
                <div class="eye left-eye">
                  <div class="eye-white">
                    <div class="pupil" :class="{ blink: isBlinking }"></div>
                  </div>
                </div>
                <div class="eye right-eye">
                  <div class="eye-white">
                    <div class="pupil" :class="{ blink: isBlinking }"></div>
                  </div>
                </div>
              </div>
              <!-- 腮红 -->
              <div class="blush left-blush"></div>
              <div class="blush right-blush"></div>
              <!-- 嘴巴 -->
              <div class="mouth" :class="{ happy: isHappy }">
                <div class="mouth-shape"></div>
              </div>
              <!-- 胡须 -->
              <div class="whiskers">
                <div class="whisker left-w1"></div>
                <div class="whisker left-w2"></div>
                <div class="whisker left-w3"></div>
                <div class="whisker right-w1"></div>
                <div class="whisker right-w2"></div>
                <div class="whisker right-w3"></div>
              </div>
            </div>
          </div>
          <!-- 爪子 -->
          <div class="paws">
            <div class="paw left-paw"></div>
            <div class="paw right-paw"></div>
          </div>
        </div>
        <div class="pet-badge" v-if="unreadCount > 0">{{ unreadCount }}</div>
      </div>
    </transition>

    <!-- 聊天面板 -->
    <transition name="chat-slide">
      <div v-if="chatOpen" class="chat-panel">
        <div class="chat-header">
          <div class="chat-header-left">
            <span class="chat-pet-icon">🐱</span>
            <span class="chat-title">AI 智能助手</span>
          </div>
          <div class="chat-header-actions">
            <button class="header-btn" @click="startNewChat" title="新对话">+</button>
            <button class="header-btn" @click="chatOpen = false" title="关闭">✕</button>
          </div>
        </div>

        <div class="chat-messages" ref="messagesContainer">
          <div v-if="messages.length === 0" class="chat-welcome">
            <div class="welcome-icon">🐱</div>
            <p>你好！我是 PVE Manager 的 AI 智能助手</p>
            <p class="welcome-sub">试试问我 PVE 运维相关问题吧</p>
            <div class="quick-actions">
              <button @click="askQuestion('如何创建虚拟机？')">🖥️ 创建虚拟机</button>
              <button @click="askQuestion('常用命令有哪些？')">📋 常用命令</button>
              <button @click="askQuestion('系统性能如何优化？')">⚡ 性能优化</button>
              <button @click="askQuestion('遇到故障怎么排查？')">🔍 故障排查</button>
            </div>
          </div>
          <div
            v-for="(msg, idx) in messages"
            :key="idx"
            class="chat-message"
            :class="msg.role"
          >
            <div class="msg-avatar">{{ msg.role === 'assistant' ? '🐱' : '👤' }}</div>
            <div class="msg-content" v-html="formatContent(msg.content)"></div>
          </div>
          <div v-if="loading" class="chat-message assistant">
            <div class="msg-avatar">🐱</div>
            <div class="msg-content typing">
              <span></span><span></span><span></span>
            </div>
          </div>
        </div>

        <div class="chat-input-area">
          <input
            v-model="inputText"
            class="chat-input"
            placeholder="输入你的问题..."
            @keyup.enter="sendMessage"
            :disabled="loading"
          />
          <button
            class="send-btn"
            @click="sendMessage"
            :disabled="!inputText.trim() || loading"
          >
            ➤
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { createConversation, sendConversationMessage } from '@/api/ai'

const chatOpen = ref(false)
const inputText = ref('')
const loading = ref(false)
const isBlinking = ref(false)
const isHappy = ref(false)
const isDragging = ref(false)
const unreadCount = ref(0)
const messagesContainer = ref<HTMLElement | null>(null)
const moodState = ref<'idle' | 'happy' | 'excited' | 'sleepy'>('idle')

const messages = ref<Array<{ role: string; content: string }>>([])
const currentConversationId = ref<number | null>(null)

// 拖拽相关
const petPosition = reactive({
  position: 'fixed' as const,
  bottom: '100px',
  right: '30px',
  left: 'auto',
  top: 'auto',
})
let dragStart = { x: 0, y: 0 }
let posStart = { right: 30, bottom: 100 }

const petTooltip = computed(() => {
  const tooltips = {
    idle: 'AI 助手 - 点击聊天',
    happy: '😊 点击和我聊天！',
    excited: '🎉 快来问我问题吧！',
    sleepy: '💤 戳我一下试试？',
  }
  return tooltips[moodState.value]
})

// 眨眼动画
let blinkTimer: ReturnType<typeof setInterval> | null = null
let moodTimer: ReturnType<typeof setTimeout> | null = null
let idleTimer: ReturnType<typeof setInterval> | null = null

function startBlink() {
  blinkTimer = setInterval(() => {
    isBlinking.value = true
    setTimeout(() => {
      isBlinking.value = false
    }, 150)
  }, 3000 + Math.random() * 2000)
}

function setMood(mood: 'idle' | 'happy' | 'excited' | 'sleepy', duration = 3000) {
  moodState.value = mood
  if (moodTimer) clearTimeout(moodTimer)
  moodTimer = setTimeout(() => {
    moodState.value = 'idle'
  }, duration)
}

function startDrag(e: MouseEvent) {
  if (chatOpen.value) return
  isDragging.value = true
  dragStart = { x: e.clientX, y: e.clientY }
  posStart = {
    right: parseInt(petPosition.right as string) || 30,
    bottom: parseInt(petPosition.bottom as string) || 100,
  }
  setMood('excited', 2000)
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

function onDrag(e: MouseEvent) {
  const dx = dragStart.x - e.clientX
  const dy = dragStart.y - e.clientY
  petPosition.right = Math.max(10, Math.min(window.innerWidth - 80, posStart.right + dx)) + 'px'
  petPosition.bottom = Math.max(10, Math.min(window.innerHeight - 80, posStart.bottom + dy)) + 'px'
}

function stopDrag() {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

async function openChat() {
  if (isDragging.value) return
  chatOpen.value = true
  setMood('happy', 5000)
}

async function askQuestion(question: string) {
  await sendMessageInternal(question)
}

async function sendMessage() {
  const text = inputText.value.trim()
  if (!text || loading.value) return
  inputText.value = ''
  await sendMessageInternal(text)
}

async function sendMessageInternal(text: string) {
  loading.value = true
  messages.value.push({ role: 'user', content: text })
  await scrollToBottom()

  try {
    if (currentConversationId.value) {
      const resp = await sendConversationMessage(currentConversationId.value, text)
      const msgs = resp.data?.messages || resp.messages || []
      messages.value = msgs.map((m: { role: string; content: string }) => ({
        role: m.role,
        content: m.content,
      }))
    } else {
      const resp = await createConversation({
        scene: 'chat',
        message: text,
        title: text.slice(0, 50),
      })
      currentConversationId.value = resp.data?.id || resp.id
      const msgs = resp.data?.messages || resp.messages || []
      messages.value = msgs.map((m: { role: string; content: string }) => ({
        role: m.role,
        content: m.content,
      }))
    }
    setMood('excited', 3000)
  } catch (err) {
    messages.value.push({
      role: 'assistant',
      content: '抱歉，出了点小问题。请稍后再试 🐾',
    })
    setMood('sleepy', 3000)
  } finally {
    loading.value = false
    await scrollToBottom()
  }
}

function startNewChat() {
  messages.value = []
  currentConversationId.value = null
  setMood('idle')
}

function formatContent(content: string): string {
  return content
    .replace(/```(\w*)\n?/g, '<pre><code class="language-$1">')
    .replace(/```/g, '</code></pre>')
    .replace(/\n/g, '<br>')
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    .replace(/(https?:\/\/[^\s]+)/g, '<a href="$1" target="_blank">$1</a>')
}

async function scrollToBottom() {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  startBlink()
  idleTimer = setInterval(() => {
    if (moodState.value === 'idle' && Math.random() > 0.7) {
      setMood('sleepy', 4000)
    }
  }, 8000)
})

onUnmounted(() => {
  if (blinkTimer) clearInterval(blinkTimer)
  if (moodTimer) clearTimeout(moodTimer)
  if (idleTimer) clearInterval(idleTimer)
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
})
</script>

<style scoped>
.floating-pet-container {
  position: fixed;
  z-index: 9999;
  pointer-events: none;
}

.pet-button {
  pointer-events: all;
  width: 80px;
  height: 80px;
  border-radius: 40px;
  background: transparent;
  cursor: grab;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s ease;
  user-select: none;
  position: fixed;
  animation: pet-idle 2s ease-in-out infinite;
  overflow: visible;
}

.pet-button:hover {
  transform: scale(1.1);
}

.pet-button:hover .head {
  transform: translateY(-2px);
}

.pet-button.dragging {
  cursor: grabbing;
  animation: none;
}

/* ===== 猫咪容器 ===== */
.cat-container {
  position: relative;
  width: 60px;
  height: 70px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* ===== 尾巴 ===== */
.tail {
  position: absolute;
  right: -8px;
  bottom: 15px;
  width: 25px;
  height: 20px;
  border-top-right-radius: 15px;
  border-right: 5px solid #f4a460;
  border-top: 5px solid #f4a460;
  animation: tail-wag 2s ease-in-out infinite;
  z-index: 1;
}

@keyframes tail-wag {
  0%, 100% { transform: rotate(-5deg); }
  50% { transform: rotate(10deg); }
}

/* ===== 身体 ===== */
.body {
  position: absolute;
  bottom: 0;
  width: 45px;
  height: 35px;
  background: linear-gradient(180deg, #ffb347 0%, #f4a460 100%);
  border-radius: 50% 50% 45% 45%;
  z-index: 2;
  box-shadow: 0 2px 8px rgba(244, 164, 96, 0.3);
}

.belly {
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 25px;
  height: 18px;
  background: #fff5e6;
  border-radius: 50%;
}

/* ===== 头部 ===== */
.head {
  position: relative;
  width: 50px;
  height: 45px;
  background: linear-gradient(180deg, #ffc87c 0%, #ffb347 100%);
  border-radius: 50% 50% 45% 45%;
  margin-bottom: 5px;
  z-index: 3;
  box-shadow: 0 2px 6px rgba(255, 179, 71, 0.3);
  transition: transform 0.3s ease;
}

/* ===== 耳朵 ===== */
.ear {
  position: absolute;
  top: -8px;
  width: 16px;
  height: 16px;
  background: #f4a460;
  border-radius: 3px 50% 0 0;
  z-index: 4;
}

.left-ear {
  left: 6px;
  transform: rotate(-15deg);
}

.right-ear {
  right: 6px;
  transform: rotate(15deg) scaleX(-1);
}

.ear-inner {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 10px;
  height: 10px;
  background: #ffb6c1;
  border-radius: 50%;
  opacity: 0.8;
}

/* ===== 脸部 ===== */
.face {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 30px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* ===== 眼睛 ===== */
.eyes {
  display: flex;
  gap: 14px;
  margin-top: 2px;
}

.eye {
  position: relative;
  width: 12px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.eye-white {
  width: 12px;
  height: 12px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 2px rgba(0,0,0,0.1);
}

.pupil {
  width: 6px;
  height: 8px;
  background: #2c2c2c;
  border-radius: 50%;
  position: relative;
  transition: all 0.15s ease;
}

.pupil.blink {
  height: 2px;
  border-radius: 1px;
}

.pupil::after {
  content: '';
  position: absolute;
  top: 1px;
  right: 1px;
  width: 3px;
  height: 3px;
  background: white;
  border-radius: 50%;
}

/* ===== 腮红 ===== */
.blush {
  position: absolute;
  width: 8px;
  height: 5px;
  background: #ffb6c1;
  border-radius: 50%;
  top: 16px;
  opacity: 0.7;
}

.left-blush {
  left: 3px;
}

.right-blush {
  right: 3px;
}

/* ===== 嘴巴 ===== */
.mouth {
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 6px;
}

.mouth-shape {
  width: 10px;
  height: 5px;
  border-bottom: 2px solid #c17a5a;
  border-radius: 0 0 50% 50%;
  transition: all 0.3s ease;
}

.mouth.happy .mouth-shape {
  width: 14px;
  height: 8px;
  border-bottom-width: 2.5px;
}

/* ===== 胡须 ===== */
.whiskers {
  position: absolute;
  width: 100%;
  top: 12px;
}

.whisker {
  position: absolute;
  width: 10px;
  height: 1.5px;
  background: #d4a574;
  border-radius: 1px;
}

.left-w1 { left: -8px; top: 0; transform: rotate(-5deg); }
.left-w2 { left: -9px; top: 4px; transform: rotate(0deg); }
.left-w3 { left: -8px; top: 8px; transform: rotate(5deg); }
.right-w1 { right: -8px; top: 0; transform: rotate(5deg); }
.right-w2 { right: -9px; top: 4px; transform: rotate(0deg); }
.right-w3 { right: -8px; top: 8px; transform: rotate(-5deg); }

/* ===== 爪子 ===== */
.paws {
  position: absolute;
  bottom: -2px;
  width: 50px;
  display: flex;
  justify-content: space-between;
  z-index: 4;
  padding: 0 5px;
}

.paw {
  width: 14px;
  height: 10px;
  background: #fff5e6;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

/* ===== 徽章 ===== */
.pet-badge {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 18px;
  height: 18px;
  background: #f56c6c;
  color: white;
  border-radius: 50%;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  border: 2px solid white;
  z-index: 10;
  box-shadow: 0 1px 3px rgba(245, 108, 108, 0.5);
}

/* ===== 空闲动画 ===== */
@keyframes pet-idle {
  0%, 100% { transform: translateY(0) rotate(-2deg); }
  25% { transform: translateY(-3px) rotate(0deg); }
  50% { transform: translateY(0) rotate(2deg); }
  75% { transform: translateY(-2px) rotate(0deg); }
}

/* ===== 聊天面板 ===== */
.chat-panel {
  pointer-events: all;
  position: fixed;
  bottom: 100px;
  right: 30px;
  width: 380px;
  height: 520px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e8eaed;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: linear-gradient(135deg, #f5a623 0%, #e8921a 100%);
  color: white;
  flex-shrink: 0;
}

.chat-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chat-pet-icon {
  font-size: 22px;
  animation: pet-bounce 1s ease-in-out infinite alternate;
}

@keyframes pet-bounce {
  from { transform: translateY(0); }
  to { transform: translateY(-3px); }
}

.chat-title {
  font-size: 15px;
  font-weight: 600;
}

.chat-header-actions {
  display: flex;
  gap: 6px;
}

.header-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.35);
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f7f8fa;
}

.chat-welcome {
  text-align: center;
  padding: 30px 10px;
}

.welcome-icon {
  font-size: 48px;
  margin-bottom: 12px;
  animation: pet-bounce 1s ease-in-out infinite alternate;
}

.welcome-sub {
  color: #999;
  font-size: 13px;
  margin-top: 4px;
}

.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 16px;
}

.quick-actions button {
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  background: white;
  border-radius: 10px;
  cursor: pointer;
  font-size: 12px;
  color: #555;
  transition: all 0.2s;
  text-align: left;
}

.quick-actions button:hover {
  border-color: #f5a623;
  color: #e8921a;
  background: #fff8f0;
}

.chat-message {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.chat-message.user {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
  background: #e8eaed;
}

.chat-message.user .msg-avatar {
  background: linear-gradient(135deg, #f5a623 0%, #e8921a 100%);
  color: white;
}

.msg-content {
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.6;
  max-width: 260px;
  word-break: break-word;
}

.chat-message.user .msg-content {
  background: linear-gradient(135deg, #f5a623 0%, #e8921a 100%);
  color: white;
  border-bottom-right-radius: 4px;
}

.chat-message.assistant .msg-content {
  background: white;
  color: #333;
  border-bottom-left-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.msg-content :deep(pre) {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 10px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
  font-size: 12px;
}

.msg-content :deep(code) {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
}

.msg-content :deep(.inline-code) {
  background: rgba(0, 0, 0, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.chat-message.user .msg-content :deep(.inline-code) {
  background: rgba(255, 255, 255, 0.2);
}

.typing {
  display: flex;
  gap: 4px;
  padding: 14px 14px !important;
  align-items: center;
  min-height: 32px;
}

.typing span {
  width: 6px;
  height: 6px;
  background: #999;
  border-radius: 50%;
  animation: typing-dot 1.4s ease-in-out infinite;
}

.typing span:nth-child(2) { animation-delay: 0.2s; }
.typing span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing-dot {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-6px); opacity: 1; }
}

.chat-input-area {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #e8eaed;
  flex-shrink: 0;
  background: white;
}

.chat-input {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 20px;
  padding: 10px 16px;
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
  background: #f7f8fa;
}

.chat-input:focus {
  border-color: #f5a623;
  background: white;
}

.send-btn {
  width: 40px;
  height: 40px;
  border: none;
  background: linear-gradient(135deg, #f5a623 0%, #e8921a 100%);
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s, opacity 0.2s;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  transform: scale(1.05);
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Vue transitions */
.pet-bounce-enter-active,
.pet-bounce-leave-active {
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.pet-bounce-enter-from {
  opacity: 0;
  transform: scale(0) translateY(20px);
}

.chat-slide-enter-active {
  transition: all 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.chat-slide-leave-active {
  transition: all 0.25s ease-in;
}

.chat-slide-enter-from {
  opacity: 0;
  transform: translateY(30px) scale(0.9);
}

.chat-slide-leave-to {
  opacity: 0;
  transform: translateY(30px) scale(0.9);
}

@media (max-width: 420px) {
  .chat-panel {
    width: calc(100vw - 20px);
    height: 60vh;
    right: 10px;
    bottom: 90px;
  }
}
</style>
