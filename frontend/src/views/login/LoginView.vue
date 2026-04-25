<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <el-icon :size="40" color="#409eff"><Monitor /></el-icon>
        <h2>{{ t('auth.loginTitle') }}</h2>
      </div>

      <el-form ref="formRef" :model="loginForm" :rules="rules" label-position="top" @submit.prevent="handleLogin">
        <!-- 登录方式切换 -->
        <el-form-item :label="t('auth.loginMethod')">
          <el-radio-group v-model="loginMethod">
            <el-radio-button value="password">{{ t('auth.passwordLogin') }}</el-radio-button>
            <el-radio-button value="token">{{ t('auth.tokenLogin') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <!-- 节点地址 -->
        <el-form-item :label="t('auth.nodeAddress')" prop="host">
          <el-input v-model="loginForm.host" :placeholder="t('auth.nodeAddressPlaceholder')" />
        </el-form-item>

        <!-- 端口 -->
        <el-form-item :label="t('auth.nodePort')" prop="port">
          <el-input-number v-model="loginForm.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>

        <!-- 用户名/密码 -->
        <template v-if="loginMethod === 'password'">
          <el-form-item :label="t('auth.username')" prop="username">
            <el-input v-model="loginForm.username" :placeholder="t('auth.usernamePlaceholder')" />
          </el-form-item>
          <el-form-item :label="t('auth.password')" prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              show-password
              :placeholder="t('auth.passwordPlaceholder')"
            />
          </el-form-item>
        </template>

        <!-- API Token -->
        <el-form-item v-else :label="t('auth.apiToken')" prop="apiToken">
          <el-input
            v-model="loginForm.apiToken"
            type="password"
            show-password
            :placeholder="t('auth.tokenPlaceholder')"
          />
        </el-form-item>

        <!-- 登录按钮 -->
        <el-form-item>
          <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin">
            {{ t('common.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Monitor } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const loginMethod = ref<'password' | 'token'>('password')

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
 * 处理登录表单提交
 * 1. 校验表单
 * 2. 调用认证 store 的 login 方法
 * 3. 登录成功后跳转到目标页面或首页
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
        authStore.saveNode({
          host: loginForm.host,
          port: loginForm.port,
          name: loginForm.host,
        })
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
</script>

<style lang="scss" scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;

  h2 {
    margin-top: 12px;
    font-size: 22px;
    font-weight: 600;
    color: #303133;
  }
}

.login-btn {
  width: 100%;
}
</style>
