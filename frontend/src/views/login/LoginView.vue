<template>
  <div class="login-page">
    <!-- 背景装饰 -->
    <div class="login-background">
      <div class="bg-shape bg-shape-1"></div>
      <div class="bg-shape bg-shape-2"></div>
      <div class="bg-shape bg-shape-3"></div>
      <div class="bg-grid"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-container animate-slide-up">
      <div class="login-card">
        <!-- Logo和标题 -->
        <div class="login-header">
          <div class="logo-wrapper">
            <el-icon :size="36" class="logo-icon"><Monitor /></el-icon>
          </div>
          <h1 class="login-title">PVE Cloud</h1>
          <p class="login-subtitle">本地私有云虚拟化管理平台</p>
        </div>

        <!-- 登录表单 -->
        <el-form ref="formRef" :model="loginForm" :rules="rules" label-position="top" @submit.prevent="handleLogin">
          <!-- 登录方式切换 -->
          <el-form-item>
            <el-segmented
              v-model="loginMethod"
              :options="[
                { label: '用户名密码', value: 'password' },
                { label: 'API Token', value: 'token' },
              ]"
              class="login-segmented"
            />
          </el-form-item>

          <!-- 节点地址 -->
          <el-form-item :label="t('auth.nodeAddress')" prop="host">
            <el-input
              v-model="loginForm.host"
              :placeholder="t('auth.nodeAddressPlaceholder')"
              size="large"
              clearable
            >
              <template #prefix>
                <el-icon><Link /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <!-- 端口 -->
          <el-form-item :label="t('auth.nodePort')" prop="port">
            <el-input-number
              v-model="loginForm.port"
              :min="1"
              :max="65535"
              size="large"
              class="port-input"
            />
          </el-form-item>

          <!-- 用户名/密码 -->
          <template v-if="loginMethod === 'password'">
            <el-form-item :label="t('auth.username')" prop="username">
              <el-input
                v-model="loginForm.username"
                :placeholder="t('auth.usernamePlaceholder')"
                size="large"
                clearable
              >
                <template #prefix>
                  <el-icon><User /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item :label="t('auth.password')" prop="password">
              <el-input
                v-model="loginForm.password"
                type="password"
                show-password
                :placeholder="t('auth.passwordPlaceholder')"
                size="large"
              >
                <template #prefix>
                  <el-icon><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </template>

          <!-- API Token -->
          <el-form-item v-else :label="t('auth.apiToken')" prop="apiToken">
            <el-input
              v-model="loginForm.apiToken"
              type="password"
              show-password
              :placeholder="t('auth.tokenPlaceholder')"
              size="large"
            >
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <!-- 记住我 -->
          <el-form-item>
            <div class="form-options">
              <el-checkbox v-model="rememberMe">记住节点配置</el-checkbox>
              <el-link type="primary" :underline="false">忘记配置？</el-link>
            </div>
          </el-form-item>

          <!-- 登录按钮 -->
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-btn"
              @click="handleLogin"
            >
              {{ loading ? '登录中...' : t('common.login') }}
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 已保存节点 -->
        <div v-if="savedNodes.length > 0" class="saved-nodes">
          <div class="saved-nodes-title">最近使用的节点</div>
          <div class="saved-nodes-list">
            <el-tag
              v-for="node in savedNodes.slice(0, 3)"
              :key="node.host"
              class="saved-node-tag"
              @click="quickLogin(node)"
            >
              {{ node.name }}:{{ node.port }}
            </el-tag>
          </div>
        </div>

        <!-- 底部信息 -->
        <div class="login-footer">
          <span>Proxmox VE Web UI v0.1.0</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Monitor, Link, User, Lock, Key } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const loginMethod = ref<'password' | 'token'>('password')
const rememberMe = ref(true)

// 已保存的节点
const savedNodes = computed(() => authStore.savedNodes)

/**
 * 登录表单数据
 */
const loginForm = reactive({
  host: '192.168.1.100',
  port: 8006,
  username: 'root',
  password: '',
  apiToken: '',
})

/**
 * 表单校验规则
 */
const rules: FormRules = {
  host: [{ required: true, message: t('auth.connectionFailed'), trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: t('auth.usernamePlaceholder'), trigger: 'blur' }],
  password: [{ required: true, message: t('auth.passwordPlaceholder'), trigger: 'blur' }],
  apiToken: [{ required: true, message: t('auth.tokenPlaceholder'), trigger: 'blur' }],
}

/**
 * 处理登录
 * 1. 校验表单
 * 2. 调用认证 store 的 login 方法
 * 3. 登录成功后保存节点配置并跳转
 */
async function handleLogin() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const success = await authStore.login({
        host: loginForm.host,
        port: loginForm.port,
        username: loginForm.username,
        password: loginMethod.value === 'password' ? loginForm.password : undefined,
        apiToken: loginMethod.value === 'token' ? loginForm.apiToken : undefined,
      })

      if (success) {
        ElMessage.success(t('auth.loginSuccess'))

        // 保存节点配置
        if (rememberMe.value) {
          authStore.saveNode({
            host: loginForm.host,
            port: loginForm.port,
            name: loginForm.host,
          })
        }

        // 跳转到目标页面
        const redirect = (route.query.redirect as string) || '/'
        router.push(redirect)
      } else {
        ElMessage.error(t('auth.loginFailed'))
      }
    } finally {
      loading.value = false
    }
  })
}

/**
 * 快速登录已保存的节点
 */
function quickLogin(node: { host: string; port: number }) {
  loginForm.host = node.host
  loginForm.port = node.port
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $gradient-login-bg;
  position: relative;
  overflow: hidden;
  padding: $spacing-8;
}

// 背景装饰 - 私有云风格
.login-background {
  position: absolute;
  inset: 0;
  overflow: hidden;

  .bg-shape {
    position: absolute;
    border-radius: 50%;
    opacity: 0.08;
    background: #fff;

    &-1 {
      width: 600px;
      height: 600px;
      top: -200px;
      right: -100px;
    }

    &-2 {
      width: 400px;
      height: 400px;
      bottom: -100px;
      left: -100px;
    }

    &-3 {
      width: 300px;
      height: 300px;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
    }
  }

  .bg-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
    background-size: 50px 50px;
  }
}

.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
}

.login-card {
  background: $gradient-login-card;
  backdrop-filter: blur(10px);
  border-radius: $radius-lg;
  padding: $spacing-10;
  box-shadow: $shadow-modal;
  border: 1px solid rgba(255, 255, 255, 0.3);

  @media (max-width: $breakpoint-sm) {
    padding: $spacing-6;
  }
}

.login-header {
  text-align: center;
  margin-bottom: $spacing-8;

  .logo-wrapper {
    width: 72px;
    height: 72px;
    margin: 0 auto $spacing-4;
    background: $gradient-primary;
    border-radius: $radius-md;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 8px 24px rgba(22, 119, 255, 0.3);

    .logo-icon {
      color: #fff;
    }
  }

  .login-title {
    font-size: $font-size-3xl;
    font-weight: $font-weight-bold;
    color: $color-text-primary;
    margin-bottom: $spacing-2;
  }

  .login-subtitle {
    color: $color-text-secondary;
    font-size: $font-size-sm;
  }
}

// 分段控制器
:deep(.login-segmented) {
  .el-segmented__item {
    padding: $spacing-2 $spacing-4;
  }
}

// 表单样式
:deep(.el-form-item) {
  margin-bottom: $spacing-5;

  .el-form-item__label {
    font-weight: $font-weight-medium;
    color: $color-text-regular;
  }
}

.port-input {
  width: 100%;
}

.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  border-radius: $radius-sm;
}

// 已保存节点
.saved-nodes {
  margin-top: $spacing-6;
  padding-top: $spacing-6;
  border-top: 1px solid $color-border-light;

  .saved-nodes-title {
    font-size: $font-size-sm;
    color: $color-text-secondary;
    margin-bottom: $spacing-3;
  }

  .saved-nodes-list {
    display: flex;
    flex-wrap: wrap;
    gap: $spacing-2;
  }

  .saved-node-tag {
    cursor: pointer;
    transition: $transition-fast;

    &:hover {
      background: $primary-1;
      color: $color-primary;
      border-color: $color-primary-light;
    }
  }
}

.login-footer {
  margin-top: $spacing-8;
  text-align: center;
  color: $color-text-placeholder;
  font-size: $font-size-xs;
}

// 标签样式覆盖
:deep(.el-tag) {
  border-radius: $radius-xs;
}
</style>
