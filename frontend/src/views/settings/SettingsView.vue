<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <span>{{ t('settings.title') }}</span>
      </template>

      <el-divider>{{ t('settings.general') }}</el-divider>

      <el-form label-width="140px" style="max-width: 500px">
        <el-form-item :label="t('settings.language')">
          <el-select v-model="language" style="width: 200px">
            <el-option label="简体中文" value="zh-CN" />
            <el-option label="English" value="en-US" disabled />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('settings.theme')">
          <el-radio-group v-model="theme">
            <el-radio-button value="light">{{ t('settings.light') }}</el-radio-button>
            <el-radio-button value="dark">{{ t('settings.dark') }}</el-radio-button>
            <el-radio-button value="auto">{{ t('settings.auto') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="t('settings.refreshInterval')">
          <el-input-number v-model="refreshInterval" :min="5" :max="300" :step="5" style="width: 200px" />
          <span class="ml-8">{{ t('settings.seconds') }}</span>
        </el-form-item>
      </el-form>

      <el-divider>{{ t('settings.nodes') }}</el-divider>

      <el-table :data="savedNodes" style="width: 100%; max-width: 600px" border>
        <el-table-column prop="name" label="节点名称" />
        <el-table-column prop="host" label="地址" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="danger" size="small" @click="removeNode(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-16">
        <el-button type="primary" @click="addNode">{{ t('settings.addNode') }}</el-button>
      </div>

      <el-divider>{{ t('settings.about') }}</el-divider>
      <el-descriptions :column="1" border style="max-width: 500px">
        <el-descriptions-item :label="t('settings.version')">v0.1.0</el-descriptions-item>
        <el-descriptions-item label="前端框架">Vue 3 + TypeScript + Vite</el-descriptions-item>
        <el-descriptions-item label="UI 库">Element Plus</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const authStore = useAuthStore()

const language = ref('zh-CN')
const theme = ref('light')
const refreshInterval = ref(30)

/**
 * 获取已保存的节点列表
 */
const savedNodes = computed(() => authStore.savedNodes)

function addNode() {
  ElMessage.info('添加节点功能开发中...')
}

/**
 * 移除已保存的节点配置
 */
function removeNode(node: { host: string; port: number }) {
  authStore.removeNode(node.host, node.port)
  ElMessage.success('已移除节点配置')
}
</script>

<style lang="scss" scoped>
.ml-8 {
  margin-left: 8px;
}
.mt-16 {
  margin-top: 16px;
}
</style>
