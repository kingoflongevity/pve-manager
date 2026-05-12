<template>
  <div class="ai-settings">
    <div class="page-header">
      <h2>AI 模型配置</h2>
      <el-button type="primary" @click="showAddDialog">添加模型</el-button>
    </div>

    <el-table :data="models" stripe style="width: 100%">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="provider" label="提供商" width="120">
        <template #default="{ row }">
          <el-tag :type="providerTag(row.provider)">{{ row.provider }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="model" label="模型" width="200" />
      <el-table-column prop="base_url" label="API 地址" min-width="250" show-overflow-tooltip />
      <el-table-column label="参数" width="160">
        <template #default="{ row }">
          <span>Max: {{ row.max_tokens }} | T: {{ row.temperature }}</span>
        </template>
      </el-table-column>
      <el-table-column label="默认" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.is_enabled" @change="toggleEnable(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="testConnection(row)">测试</el-button>
          <el-button v-if="!row.is_default" size="small" @click="setDefault(row)">设为默认</el-button>
          <el-button size="small" type="danger" @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模型' : '添加模型'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="例如: GPT-4o" />
        </el-form-item>
        <el-form-item label="提供商" required>
          <el-select v-model="form.provider" style="width: 100%">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Azure OpenAI" value="azure" />
            <el-option label="通义千问 (Qwen)" value="qwen" />
            <el-option label="智谱 AI (GLM)" value="zhipu" />
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="Ollama (本地)" value="ollama" />
            <el-option label="自定义 (OpenAI 兼容)" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 地址" required>
          <el-input v-model="form.base_url" placeholder="例如: https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="API Key" required>
          <el-input v-model="form.api_key" type="password" show-password placeholder="sk-..." />
        </el-form-item>
        <el-form-item label="模型名" required>
          <el-input v-model="form.model" placeholder="例如: gpt-4o" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Max Tokens">
              <el-input-number v-model="form.max_tokens" :min="512" :max="128000" :step="512" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Temperature">
              <el-slider v-model="form.temperature" :min="0" :max="2" :step="0.1" style="width: 180px" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="form.timeout" :min="10" :max="300" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveModel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAIModels, createAIModel, updateAIModel, deleteAIModel, setDefaultModel, testModelConnection } from '@/api/ai'
import type { AIModelConfig } from '@/api/ai'

const models = ref<AIModelConfig[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)

const form = ref({
  name: '', provider: 'openai', base_url: 'https://api.openai.com/v1',
  api_key: '', model: 'gpt-4o', max_tokens: 4096, temperature: 0.7, timeout: 60,
})

onMounted(() => fetchModels())

async function fetchModels() {
  try {
    models.value = await getAIModels()
  } catch { ElMessage.error('获取模型配置失败') }
}

function providerTag(p: string) {
  const m: Record<string, string> = { openai: 'primary', qwen: 'success', zhipu: 'warning', deepseek: 'info', ollama: 'danger' }
  return m[p] || ''
}

function showAddDialog() {
  isEdit.value = false
  Object.assign(form.value, { name: '', provider: 'openai', base_url: 'https://api.openai.com/v1', api_key: '', model: 'gpt-4o', max_tokens: 4096, temperature: 0.7, timeout: 60 })
  dialogVisible.value = true
}

async function saveModel() {
  saving.value = true
  try {
    await createAIModel(form.value as any)
    ElMessage.success('创建成功')
    dialogVisible.value = false
    fetchModels()
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

async function toggleEnable(row: AIModelConfig) {
  try {
    await updateAIModel(row.id, { is_enabled: !row.is_enabled })
    row.is_enabled = !row.is_enabled
  } catch { ElMessage.error('更新失败') }
}

async function setDefault(row: AIModelConfig) {
  try {
    await setDefaultModel(row.id)
    ElMessage.success('已设为默认')
    fetchModels()
  } catch { ElMessage.error('设置失败') }
}

async function testConnection(row: AIModelConfig) {
  try {
    await testModelConnection(row.id)
    ElMessage.success('连接测试成功')
  } catch { ElMessage.error('连接测试失败，请检查配置') }
}

function confirmDelete(row: AIModelConfig) {
  ElMessageBox.confirm(`确定删除模型 "${row.name}"？`, '提示').then(async () => {
    try {
      await deleteAIModel(row.id)
      ElMessage.success('已删除')
      fetchModels()
    } catch { ElMessage.error('删除失败') }
  }).catch(() => {})
}
</script>

<style scoped>
.ai-settings { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
</style>
