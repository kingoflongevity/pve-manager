<template>
  <div class="login-page">
    <div class="login-background">
      <div class="scan-line"></div>
      <div class="grid-overlay"></div>
      <div class="noise"></div>
    </div>

    <div class="login-container">
      <div class="login-panel">
        <div class="panel-header">
          <div class="brand">
            <div class="brand-icon">
              <el-icon :size="28"><Monitor /></el-icon>
            </div>
            <div class="brand-text">
              <h1>PVE Cloud</h1>
              <span class="brand-version">Operations Console v0.2.0</span>
            </div>
          </div>
          <div class="status-bar">
            <span class="status-dot"></span>
            <span class="status-text">SYSTEM READY</span>
          </div>
        </div>

        <div class="panel-body">
          <el-steps :active="currentStep" finish-status="success" class="wizard-steps">
            <el-step title="连接目标" />
            <el-step title="身份验证" />
            <el-step title="连接确认" />
          </el-steps>

          <div class="step-content">
            <transition name="step-fade" mode="out-in">
              <div v-if="currentStep === 0" class="step-panel" key="step0">
                <div class="step-title">
                  <el-icon :size="20" class="step-icon"><Link /></el-icon>
                  <h3>PVE 服务器地址</h3>
                </div>
                <p class="step-desc">输入 Proxmox VE 集群的管理地址和端口号</p>

                <div class="input-group">
                  <el-form-item label="服务器地址" class="field">
                    <el-input
                      v-model="form.host"
                      placeholder="例: 192.168.1.10"
                      size="large"
                      @keyup.enter="nextStep"
                    >
                      <template #prefix><el-icon><Link /></el-icon></template>
                    </el-input>
                  </el-form-item>

                  <el-form-item label="API 端口" class="field">
                    <el-input-number
                      v-model="form.port"
                      :min="1"
                      :max="65535"
                      size="large"
                      class="port-field"
                    />
                  </el-form-item>

                  <el-form-item label="验证方式" class="field">
                    <el-radio-group v-model="form.realm" size="large">
                      <el-radio-button value="pam">PAM (Linux)</el-radio-button>
                      <el-radio-button value="pve">PVE</el-radio-button>
                    </el-radio-group>
                  </el-form-item>
                </div>
              </div>

              <div v-else-if="currentStep === 1" class="step-panel" key="step1">
                <div class="step-title">
                  <el-icon :size="20" class="step-icon"><Key /></el-icon>
                  <h3>管理员凭据</h3>
                </div>
                <p class="step-desc">输入具有 root 或 Administrator 权限的账户信息</p>

                <div class="input-group">
                  <el-form-item label="用户名" class="field">
                    <el-input
                      v-model="form.username"
                      placeholder="root"
                      size="large"
                      @keyup.enter="nextStep"
                    >
                      <template #prefix><el-icon><User /></el-icon></template>
                    </el-input>
                  </el-form-item>

                  <el-form-item label="密码" class="field">
                    <el-input
                      v-model="form.password"
                      type="password"
                      show-password
                      placeholder="输入 PVE root 密码"
                      size="large"
                      @keyup.enter="handleLogin"
                    >
                      <template #prefix><el-icon><Lock /></el-icon></template>
                    </el-input>
                  </el-form-item>
                </div>
              </div>

              <div v-else class="step-panel" key="step2">
                <div class="step-title">
                  <el-icon :size="20" class="step-icon"><CircleCheck /></el-icon>
                  <h3>连接验证中...</h3>
                </div>

                <div class="verification-status">
                  <div class="check-item" v-for="item in checks" :key="item.label">
                    <el-icon :size="18" :class="['check-icon', item.status]">
                      <Loading v-if="item.status === 'loading'" />
                      <CircleCheck v-else-if="item.status === 'success'" />
                      <CircleClose v-else-if="item.status === 'error'" />
                    </el-icon>
                    <span class="check-label">{{ item.label }}</span>
                    <span class="check-result">{{ item.result }}</span>
                  </div>
                </div>

                <div v-if="loginError" class="error-box">
                  <el-icon :size="16"><WarningFilled /></el-icon>
                  <span>{{ loginError }}</span>
                </div>
              </div>
            </transition>
          </div>

          <div class="panel-actions">
            <el-button
              v-if="currentStep > 0 && currentStep < 2"
              size="large"
              @click="prevStep"
              :disabled="loading"
            >
              <el-icon><Back /></el-icon>
              上一步
            </el-button>
            <el-button
              v-if="currentStep < 2"
              type="primary"
              size="large"
              @click="nextStep"
              :disabled="!canProceed"
              class="btn-next"
            >
              下一步
              <el-icon><Right /></el-icon>
            </el-button>
            <el-button
              v-if="currentStep === 2"
              type="primary"
              size="large"
              @click="handleLogin"
              :loading="loading"
              :disabled="!canLogin"
              class="btn-connect"
            >
              {{ loading ? '连接中...' : '连接控制台' }}
            </el-button>
          </div>
        </div>

        <div class="panel-footer">
          <span>Proxmox VE Web Management Console</span>
          <span class="divider">|</span>
          <span>Powered by Go + Vue3</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Monitor, Link, User, Lock, Key,
  CircleCheck, CircleClose, Loading, WarningFilled,
  Back, Right,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const currentStep = ref(0)
const loading = ref(false)
const loginError = ref('')

const form = reactive({
  host: '',
  port: 8006,
  username: 'root',
  password: '',
  realm: 'pam',
})

const canProceed = computed(() => {
  if (currentStep.value === 0) return form.host.trim() !== ''
  if (currentStep.value === 1) return form.password.trim() !== ''
  return false
})

const canLogin = computed(() => {
  return checks.value.every(c => c.status === 'success')
})

interface CheckItem {
  label: string
  status: 'loading' | 'success' | 'error' | 'pending'
  result: string
}

const checks = ref<CheckItem[]>([
  { label: '网络连接', status: 'pending', result: '等待检测' },
  { label: 'API 认证', status: 'pending', result: '等待验证' },
  { label: '权限检查', status: 'pending', result: '等待检查' },
])

function nextStep() {
  if (!canProceed.value) return
  if (currentStep.value === 0) {
    currentStep.value = 1
  } else if (currentStep.value === 1) {
    currentStep.value = 2
    runChecks()
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
    loginError.value = ''
  }
}

async function runChecks() {
  checks.value = [
    { label: '网络连接', status: 'loading', result: '检测中...' },
    { label: 'API 认证', status: 'pending', result: '等待验证' },
    { label: '权限检查', status: 'pending', result: '等待检查' },
  ]

  loginError.value = ''

  try {
    const success = await authStore.login({
      host: form.host,
      port: form.port,
      username: form.username,
      password: form.password,
      realm: form.realm,
    })

    if (success) {
      checks.value[0] = { label: '网络连接', status: 'success', result: '已连接' }
      checks.value[1] = { label: 'API 认证', status: 'success', result: '已验证' }
      checks.value[2] = { label: '权限检查', status: 'success', result: '管理员权限' }
      ElMessage.success('认证成功')
      const redirect = (route.query.redirect as string) || '/'
      router.push(redirect)
    } else {
      checks.value[1] = { label: 'API 认证', status: 'error', result: '认证失败' }
      loginError.value = 'PVE 认证失败，请检查用户名和密码'
    }
  } catch (error: any) {
    const msg = error?.message || error?.response?.data?.message || '连接失败'
    checks.value[0] = { label: '网络连接', status: 'error', result: '无法连接' }
    loginError.value = msg
  }
}
</script>

<style lang="scss" scoped>
@import '@/assets/styles/variables';

.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0e17;
  position: relative;
  overflow: hidden;
}

.login-background {
  position: absolute;
  inset: 0;
  z-index: 0;

  .scan-line {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, #3b82f6 50%, transparent);
    animation: scan 4s ease-in-out infinite;
    opacity: 0.3;
  }

  .grid-overlay {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(59, 130, 246, 0.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(59, 130, 246, 0.05) 1px, transparent 1px);
    background-size: 40px 40px;
  }

  .noise {
    position: absolute;
    inset: 0;
    opacity: 0.02;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='1'/%3E%3C/svg%3E");
  }
}

@keyframes scan {
  0%, 100% { top: 0; }
  50% { top: 100%; }
}

.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 520px;
  padding: 16px;
}

.login-panel {
  background: rgba(17, 24, 39, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 0 40px rgba(59, 130, 246, 0.08), 0 20px 60px rgba(0, 0, 0, 0.5);
}

.panel-header {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
  background: rgba(15, 23, 42, 0.5);

  .brand {
    display: flex;
    align-items: center;
    gap: 14px;

    .brand-icon {
      width: 44px;
      height: 44px;
      background: linear-gradient(135deg, #3b82f6, #1d4ed8);
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
    }

    .brand-text h1 {
      font-size: 18px;
      font-weight: 700;
      color: #f1f5f9;
      margin: 0;
      letter-spacing: -0.5px;
    }

    .brand-version {
      font-size: 11px;
      color: #64748b;
      font-family: 'JetBrains Mono', 'Consolas', monospace;
    }
  }

  .status-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 10px;
    color: #10b981;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    letter-spacing: 1px;

    .status-dot {
      width: 6px;
      height: 6px;
      background: #10b981;
      border-radius: 50%;
      animation: pulse 2s ease-in-out infinite;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.panel-body {
  padding: 24px;
}

.wizard-steps {
  margin-bottom: 24px;

  :deep(.el-step__title) {
    font-size: 12px;
    color: #94a3b8;
  }

  :deep(.el-step__line) {
    background-color: rgba(59, 130, 246, 0.2);
  }

  :deep(.el-step__head.is-process) {
    .el-step__icon {
      border-color: #3b82f6;
    }
  }

  :deep(.el-step__icon-inner) {
    font-size: 12px;
    font-weight: 700;
  }
}

.step-content {
  min-height: 260px;
}

.step-panel {
  .step-title {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;

    .step-icon {
      color: #3b82f6;
    }

    h3 {
      font-size: 15px;
      font-weight: 600;
      color: #e2e8f0;
      margin: 0;
    }
  }

  .step-desc {
    font-size: 12px;
    color: #64748b;
    margin: 0 0 20px 0;
    padding-left: 30px;
  }
}

.input-group {
  .field {
    margin-bottom: 16px;

    :deep(.el-form-item__label) {
      font-size: 12px;
      color: #94a3b8;
      font-weight: 500;
      margin-bottom: 6px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
  }

  .port-field {
    width: 100%;
  }
}

:deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(59, 130, 246, 0.15);
  box-shadow: none;

  &:hover {
    border-color: rgba(59, 130, 246, 0.3);
  }

  &.is-focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
  }

  .el-input__inner {
    color: #e2e8f0;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
  }

  .el-input__prefix-inner {
    color: #3b82f6;
  }
}

.verification-status {
  margin-top: 16px;
  padding: 16px;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(59, 130, 246, 0.1);
  border-radius: 4px;
}

.check-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid rgba(59, 130, 246, 0.05);

  &:last-child {
    border-bottom: none;
  }

  .check-icon {
    &.loading { color: #3b82f6; animation: spin 1s linear infinite; }
    &.success { color: #10b981; }
    &.error { color: #ef4444; }
  }

  .check-label {
    flex: 1;
    font-size: 13px;
    color: #cbd5e1;
  }

  .check-result {
    font-size: 11px;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    color: #64748b;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.error-box {
  margin-top: 16px;
  padding: 12px 16px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fca5a5;
  font-size: 13px;
}

.panel-actions {
  margin-top: 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;

  .btn-next, .btn-connect {
    min-width: 120px;
    background: linear-gradient(135deg, #3b82f6, #1d4ed8);
    border: none;
    font-weight: 600;
  }

  :deep(.el-button:not(.btn-next):not(.btn-connect)) {
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(59, 130, 246, 0.2);
    color: #94a3b8;
  }
}

.panel-footer {
  padding: 12px 24px;
  border-top: 1px solid rgba(59, 130, 246, 0.1);
  background: rgba(15, 23, 42, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 11px;
  color: #475569;
  font-family: 'JetBrains Mono', 'Consolas', monospace;
}

.step-fade-enter-active,
.step-fade-leave-active {
  transition: all 0.25s ease;
}

.step-fade-enter-from {
  opacity: 0;
  transform: translateX(10px);
}

.step-fade-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}
</style>
