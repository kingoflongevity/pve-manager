<template>
  <div class="login-page">
    <!-- 背景装饰层 -->
    <div class="login-background">
      <div class="bg-gradient"></div>
      <div class="bg-circle circle-1"></div>
      <div class="bg-circle circle-2"></div>
      <div class="bg-circle circle-3"></div>
      <div class="bg-grid"></div>
    </div>

    <!-- 主登录容器 -->
    <div class="login-container">
      <!-- Logo 和品牌信息 -->
      <div class="login-header">
        <div class="logo-wrapper">
          <div class="logo-icon">
            <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="32" cy="32" r="28" stroke="url(#logo-gradient)" stroke-width="4" />
              <path d="M24 20L40 32L24 44V20Z" fill="url(#logo-gradient)" />
              <defs>
                <linearGradient id="logo-gradient" x1="0" y1="0" x2="64" y2="64">
                  <stop offset="0%" stop-color="#667eea" />
                  <stop offset="100%" stop-color="#764ba2" />
                </linearGradient>
              </defs>
            </svg>
          </div>
          <div class="brand-text">
            <h1>PVE 运维管理中心</h1>
            <p class="brand-subtitle">Enterprise Virtualization Management Platform</p>
          </div>
        </div>
      </div>

      <!-- 步骤条 -->
      <el-steps :active="currentStep" finish-status="success" class="wizard-steps" align-center>
        <el-step title="连接目标" description="配置 PVE 服务器" />
        <el-step title="身份验证" description="输入管理员凭据" />
        <el-step title="连接确认" description="验证系统状态" />
      </el-steps>

      <!-- 步骤内容 -->
      <transition name="step-fade" mode="out-in">
        <div v-if="currentStep === 0" class="step-panel" key="step0">
          <el-form
            ref="formRef"
            :model="form"
            :rules="stepRules[0]"
            label-position="top"
            class="login-form"
          >
            <el-form-item label="服务器地址" prop="host">
              <el-input
                v-model="form.host"
                placeholder="例: 192.168.1.100 或 pve.example.com"
                size="large"
                prefix-icon="Monitor"
                clearable
              />
            </el-form-item>

            <div class="form-row">
              <el-form-item label="API 端口" prop="port" class="flex-1">
                <el-input-number
                  v-model="form.port"
                  :min="1"
                  :max="65535"
                  :controls="false"
                  class="full-width"
                  size="large"
                />
              </el-form-item>

              <el-form-item label="验证方式" prop="realm" class="flex-2">
                <el-radio-group v-model="form.realm" size="large" class="radio-group">
                  <el-radio-button value="pam">PAM (Linux)</el-radio-button>
                  <el-radio-button value="pve">PVE</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </div>

            <el-alert
              v-if="formErrors"
              :title="formErrors"
              type="error"
              show-icon
              closable
              class="error-alert"
              @close="formErrors = ''"
            />

            <el-button
              type="primary"
              size="large"
              class="next-btn"
              :disabled="!canProceed"
              :loading="checking"
              @click="nextStep"
            >
              下一步
              <el-icon class="arrow-icon"><ArrowRight /></el-icon>
            </el-button>
          </el-form>
        </div>

        <div v-else-if="currentStep === 1" class="step-panel" key="step1">
          <el-form
            ref="formRef"
            :model="form"
            :rules="stepRules[1]"
            label-position="top"
            class="login-form"
          >
            <div class="server-info">
              <el-tag size="small" type="info" class="info-tag">
                <el-icon><Monitor /></el-icon>
                {{ form.host }}:{{ form.port }}
              </el-tag>
              <el-tag size="small" class="info-tag">
                <el-icon><User /></el-icon>
                {{ form.realm === 'pam' ? 'PAM (Linux)' : 'PVE' }}
              </el-tag>
            </div>

            <el-form-item label="用户名" prop="username">
              <el-input
                v-model="form.username"
                placeholder="root"
                size="large"
                prefix-icon="User"
                clearable
              />
            </el-form-item>

            <el-form-item label="密码" prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="输入管理员密码"
                size="large"
                prefix-icon="Lock"
                show-password
              />
            </el-form-item>

            <el-alert
              v-if="formErrors"
              :title="formErrors"
              type="error"
              show-icon
              closable
              class="error-alert"
              @close="formErrors = ''"
            />

            <div class="step-actions">
              <el-button size="large" class="back-btn" @click="prevStep">
                <el-icon><ArrowLeft /></el-icon>
                上一步
              </el-button>
              <el-button
                type="primary"
                size="large"
                class="next-btn"
                :disabled="!canProceed"
                :loading="checking"
                @click="nextStep"
              >
                下一步
                <el-icon class="arrow-icon"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </el-form>
        </div>

        <div v-else class="step-panel" key="step2">
          <div class="verification-status">
            <div class="status-header">
              <el-icon v-if="verificationDone" :size="48" class="success-icon">
                <CircleCheckFilled />
              </el-icon>
              <div v-else class="loading-spinner">
                <div class="spinner-ring"></div>
              </div>
              <h3 v-if="verificationDone">验证成功</h3>
              <h3 v-else>正在验证连接...</h3>
            </div>

            <div class="check-items">
              <div class="check-item" v-for="item in checks" :key="item.label">
                <el-icon :class="['check-icon', item.status]" :size="20">
                  <Loading v-if="item.status === 'loading'" />
                  <CircleCheckFilled v-else-if="item.status === 'success'" />
                  <CircleCloseFilled v-else-if="item.status === 'error'" />
                  <span v-else class="pending-dot">○</span>
                </el-icon>
                <span class="check-label">{{ item.label }}</span>
                <span class="check-result" :class="item.status">{{ item.result }}</span>
              </div>
            </div>

            <div class="step-actions">
              <el-button size="large" class="back-btn" :disabled="checking" @click="prevStep">
                <el-icon><ArrowLeft /></el-icon>
                上一步
              </el-button>
              <el-button
                type="success"
                size="large"
                class="login-btn"
                :disabled="!canLogin"
                @click="handleLogin"
              >
                登录系统
                <el-icon><DArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </transition>

      <!-- 底部信息 -->
      <div class="login-footer">
        <span>© 2026 PVE Manager</span>
        <span class="footer-dot">•</span>
        <span>Enterprise Edition</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import {
  ArrowRight,
  ArrowLeft,
  DArrowRight,
  CircleCheckFilled,
  CircleCloseFilled,
  Loading,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

interface CheckItem {
  label: string
  status: 'pending' | 'loading' | 'success' | 'error'
  result: string
}

interface LoginForm {
  host: string
  port: number
  username: string
  password: string
  realm: string
}

const formRef = ref<FormInstance>()
const currentStep = ref(0)
const checking = ref(false)
const verificationDone = ref(false)
const formErrors = ref('')

const form = ref<LoginForm>({
  host: '',
  port: 8006,
  username: '',
  password: '',
  realm: 'pam',
})

const checks = ref<CheckItem[]>([
  { label: '网络连接', status: 'pending', result: '等待检测' },
  { label: 'API 认证', status: 'pending', result: '等待验证' },
  { label: '权限检查', status: 'pending', result: '等待检查' },
])

const stepRules: FormRules[] = [
  {
    host: [
      { required: true, message: '请输入服务器地址', trigger: 'blur' },
      { pattern: /^[\w.-]+$/, message: '地址格式不正确', trigger: 'blur' },
    ],
    port: [
      { required: true, message: '请输入端口号', trigger: 'blur' },
    ],
  },
  {
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
    ],
    password: [
      { required: true, message: '请输入密码', trigger: 'blur' },
      { min: 4, message: '密码至少 4 位', trigger: 'blur' },
    ],
  },
]

const canProceed = computed(() => {
  if (currentStep.value === 0) {
    return form.value.host && form.value.port
  }
  return form.value.username && form.value.password.length >= 4
})

const canLogin = computed(() => {
  return verificationDone.value && checks.value.every(c => c.status === 'success')
})

async function nextStep() {
  if (!canProceed.value) return
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  formErrors.value = ''

  if (currentStep.value === 0) {
    currentStep.value = 1
  } else if (currentStep.value === 1) {
    currentStep.value = 2
    await runChecks()
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
    verificationDone.value = false
    checks.value = [
      { label: '网络连接', status: 'pending', result: '等待检测' },
      { label: 'API 认证', status: 'pending', result: '等待验证' },
      { label: '权限检查', status: 'pending', result: '等待检查' },
    ]
  }
}

async function runChecks() {
  checking.value = true
  verificationDone.value = false

  checks.value = [
    { label: '网络连接', status: 'loading', result: '检测中...' },
    { label: 'API 认证', status: 'pending', result: '等待验证' },
    { label: '权限检查', status: 'pending', result: '等待检查' },
  ]

  try {
    const success = await authStore.login({
      host: form.value.host,
      port: form.value.port,
      username: form.value.username,
      password: form.value.password,
      realm: form.value.realm,
    })

    if (success) {
      checks.value[0] = { label: '网络连接', status: 'success', result: '已连接' }
      checks.value[1] = { label: 'API 认证', status: 'success', result: '已验证' }
      checks.value[2] = { label: '权限检查', status: 'success', result: '管理员权限' }
      verificationDone.value = true
    } else {
      checks.value[0] = { label: '网络连接', status: 'error', result: '连接失败' }
      checks.value[1] = { label: 'API 认证', status: 'pending', result: '未验证' }
      checks.value[2] = { label: '权限检查', status: 'pending', result: '未检查' }
      formErrors.value = '登录失败，请检查凭据是否正确'
    }
  } catch (e) {
    checks.value[0] = { label: '网络连接', status: 'error', result: '连接失败' }
    checks.value[1] = { label: 'API 认证', status: 'pending', result: '未验证' }
    checks.value[2] = { label: '权限检查', status: 'pending', result: '未检查' }
    formErrors.value = '连接异常，请检查网络或服务器状态'
  } finally {
    checking.value = false
  }
}

async function handleLogin() {
  if (!canLogin.value) return
  ElMessage.success('登录成功，正在跳转...')
  router.push('/dashboard')
}
</script>

<style scoped lang="scss">
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f0c29 0%, #1a1a2e 50%, #16213e 100%);
  padding: 20px;
  font-family: 'Fira Sans', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  overflow: hidden;
}

// 背景装饰层
.login-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
  overflow: hidden;

  .bg-gradient {
    position: absolute;
    width: 100%;
    height: 100%;
    background: 
      radial-gradient(circle at 20% 50%, rgba(102, 126, 234, 0.12) 0%, transparent 50%),
      radial-gradient(circle at 80% 20%, rgba(118, 75, 162, 0.1) 0%, transparent 50%),
      radial-gradient(circle at 40% 80%, rgba(99, 102, 241, 0.08) 0%, transparent 50%);
  }

  .bg-circle {
    position: absolute;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.03);
    backdrop-filter: blur(1px);
    border: 1px solid rgba(255, 255, 255, 0.05);

    &.circle-1 {
      width: 600px;
      height: 600px;
      top: -200px;
      right: -100px;
      animation: float 20s ease-in-out infinite;
    }

    &.circle-2 {
      width: 400px;
      height: 400px;
      bottom: -100px;
      left: -50px;
      animation: float 15s ease-in-out infinite reverse;
    }

    &.circle-3 {
      width: 200px;
      height: 200px;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      animation: pulse 8s ease-in-out infinite;
    }
  }

  .bg-grid {
    position: absolute;
    width: 100%;
    height: 100%;
    background-image: 
      linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
    background-size: 60px 60px;
  }
}

// 主容器
.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 480px;
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  padding: 40px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  animation: slideUp 0.6s ease-out;
}

// Logo 和品牌
.login-header {
  text-align: center;
  margin-bottom: 32px;

  .logo-wrapper {
    display: inline-flex;
    align-items: center;
    gap: 16px;

    .logo-icon {
      width: 56px;
      height: 56px;
      flex-shrink: 0;

      svg {
        width: 100%;
        height: 100%;
        filter: drop-shadow(0 4px 12px rgba(102, 126, 234, 0.4));
      }
    }

    .brand-text {
      text-align: left;

      h1 {
        font-size: 24px;
        font-weight: 700;
        color: #f8fafc;
        margin: 0;
        letter-spacing: 0.5px;
      }

      .brand-subtitle {
        font-size: 12px;
        color: #94a3b8;
        margin: 4px 0 0;
        letter-spacing: 0.3px;
        font-family: 'Fira Code', monospace;
      }
    }
  }
}

// 步骤条
.wizard-steps {
  margin-bottom: 32px;

  :deep(.el-step__title) {
    font-size: 14px;
    font-weight: 600;
    color: #e2e8f0;
  }

  :deep(.el-step__description) {
    font-size: 12px;
    color: #64748b;
  }

  :deep(.el-step__head.is-finish) {
    color: #10b981;
    border-color: #10b981;
  }

  :deep(.el-step__head.is-process) {
    color: #667eea;
    border-color: #667eea;
  }

  :deep(.el-step__head.is-wait) {
    color: #475569;
    border-color: #334155;
  }

  :deep(.el-step__line) {
    background: #334155;
  }
}

// 表单
.login-form {
  .form-row {
    display: flex;
    gap: 16px;

    .flex-1 {
      flex: 1;
    }

    .flex-2 {
      flex: 2;
    }
  }

  :deep(.el-form-item__label) {
    font-weight: 500;
    color: #cbd5e1;
    font-size: 14px;
  }

  :deep(.el-input__wrapper) {
    background: rgba(30, 41, 59, 0.6);
    border: 1px solid rgba(148, 163, 184, 0.15);
    box-shadow: none;
    border-radius: 12px;
    transition: all 0.2s;

    &:hover {
      border-color: rgba(148, 163, 184, 0.25);
    }

    &.is-focus {
      border-color: #667eea;
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
    }
  }

  :deep(.el-input__inner) {
    color: #f8fafc;
    font-size: 15px;

    &::placeholder {
      color: #475569;
    }
  }

  :deep(.el-input-number) {
    width: 100%;

    .el-input__wrapper {
      background: rgba(30, 41, 59, 0.6);
      border: 1px solid rgba(148, 163, 184, 0.15);
    }

    .el-input__inner {
      color: #f8fafc;
    }
  }

  :deep(.el-radio-group) {
    display: flex;
    width: 100%;

    .el-radio-button__inner {
      background: rgba(30, 41, 59, 0.6);
      border-color: rgba(148, 163, 184, 0.15);
      color: #94a3b8;
      border-radius: 10px;
      padding: 10px 20px;
      width: 100%;
      font-weight: 500;

      &:hover {
        color: #a5b4fc;
      }
    }

    .is-active .el-radio-button__inner {
      background: linear-gradient(135deg, #667eea, #764ba2);
      border-color: transparent;
      color: #fff;
      box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
    }
  }
}

// 按钮
.next-btn, .login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  margin-top: 8px;
  transition: all 0.3s;

  .arrow-icon {
    margin-left: 8px;
    transition: transform 0.2s;
  }

  &:hover .arrow-icon {
    transform: translateX(4px);
  }
}

.next-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, #818cf8 0%, #8b5cf6 100%);
    box-shadow: 0 8px 24px rgba(102, 126, 234, 0.35);
    transform: translateY(-2px);
  }

  &:active:not(:disabled) {
    transform: translateY(0);
  }
}

.login-btn {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
    box-shadow: 0 8px 24px rgba(16, 185, 129, 0.35);
    transform: translateY(-2px);
  }

  &:active:not(:disabled) {
    transform: translateY(0);
  }
}

.back-btn {
  height: 48px;
  padding: 0 32px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 12px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.15);
  color: #94a3b8;

  &:hover {
    background: rgba(51, 65, 85, 0.6);
    border-color: rgba(148, 163, 184, 0.25);
    color: #e2e8f0;
  }
}

.step-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;

  .back-btn {
    flex-shrink: 0;
  }

  .next-btn, .login-btn {
    flex: 1;
  }
}

// 服务器信息标签
.server-info {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  padding: 12px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);

  .info-tag {
    background: rgba(51, 65, 85, 0.6);
    border-color: rgba(148, 163, 184, 0.1);
    color: #cbd5e1;
    display: flex;
    align-items: center;
    gap: 4px;
  }
}

// 错误提示
.error-alert {
  margin-top: 16px;
  border-radius: 10px;
  border: none;
  background: rgba(239, 68, 68, 0.1);

  :deep(.el-alert__content) {
    color: #fca5a5;
  }
}

// 验证状态
.verification-status {
  text-align: center;
  padding: 20px 0;

  .status-header {
    margin-bottom: 32px;

    .success-icon {
      color: #10b981;
      margin-bottom: 12px;
      filter: drop-shadow(0 4px 12px rgba(16, 185, 129, 0.4));
    }

    .loading-spinner {
      width: 48px;
      height: 48px;
      margin: 0 auto 12px;
      position: relative;

      .spinner-ring {
        position: absolute;
        width: 100%;
        height: 100%;
        border: 3px solid transparent;
        border-top-color: #667eea;
        border-radius: 50%;
        animation: spin 1s linear infinite;

        &::before {
          content: '';
          position: absolute;
          top: 4px;
          left: 4px;
          right: 4px;
          bottom: 4px;
          border: 3px solid transparent;
          border-top-color: #764ba2;
          border-radius: 50%;
          animation: spin 0.8s linear infinite reverse;
        }
      }
    }

    h3 {
      font-size: 18px;
      font-weight: 600;
      color: #f8fafc;
      margin: 0;
    }
  }

  .check-items {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 24px;

    .check-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 14px 16px;
      background: rgba(30, 41, 59, 0.4);
      border-radius: 10px;
      border: 1px solid rgba(148, 163, 184, 0.1);
      transition: all 0.2s;

      .check-icon {
        flex-shrink: 0;

        &.loading {
          color: #667eea;
          animation: spin 1s linear infinite;
        }

        &.success {
          color: #10b981;
        }

        &.error {
          color: #ef4444;
        }

        .pending-dot {
          color: #475569;
          font-size: 16px;
        }
      }

      .check-label {
        flex: 1;
        text-align: left;
        font-weight: 500;
        color: #cbd5e1;
        font-size: 14px;
      }

      .check-result {
        font-family: 'Fira Code', monospace;
        font-size: 12px;
        padding: 4px 10px;
        border-radius: 6px;
        background: rgba(51, 65, 85, 0.4);
        color: #64748b;

        &.success {
          background: rgba(16, 185, 129, 0.15);
          color: #10b981;
        }

        &.error {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
        }
      }
    }
  }
}

// 底部信息
.login-footer {
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 12px;
  color: #475569;

  .footer-dot {
    color: #334155;
  }
}

// 过渡动画
.step-fade-enter-active {
  transition: all 0.3s ease-out;
}

.step-fade-leave-active {
  transition: all 0.2s ease-in;
}

.step-fade-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.step-fade-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

// 关键帧动画
@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-20px) rotate(3deg);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.5;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.1);
    opacity: 0.3;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

// 响应式
@media (max-width: 480px) {
  .login-container {
    padding: 24px;
    border-radius: 20px;
  }

  .login-header .logo-wrapper {
    flex-direction: column;
    text-align: center;

    .brand-text {
      text-align: center;
    }
  }

  .form-row {
    flex-direction: column;
  }
}
</style>
