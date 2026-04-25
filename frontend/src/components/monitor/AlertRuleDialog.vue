<template>
  <el-dialog
    :model-value="visible"
    :title="isEdit ? '编辑告警规则' : '添加告警规则'"
    width="560px"
    @update:model-value="$emit('update:visible', $event)"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="right"
    >
      <!-- 规则名称 -->
      <el-form-item label="规则名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入规则名称" />
      </el-form-item>

      <!-- 监控指标 -->
      <el-form-item label="监控指标" prop="metric">
        <el-select v-model="form.metric" placeholder="请选择监控指标" style="width: 100%">
          <el-option label="CPU 使用率" value="cpu" />
          <el-option label="内存使用率" value="memory" />
          <el-option label="磁盘使用率" value="disk" />
          <el-option label="网络流量" value="network" />
          <el-option label="节点在线状态" value="node_status" />
        </el-select>
      </el-form-item>

      <!-- 条件运算符 -->
      <el-form-item label="条件" prop="condition">
        <el-select v-model="form.condition" placeholder="请选择条件" style="width: 100%">
          <el-option label="大于" value="greater_than" />
          <el-option label="小于" value="less_than" />
          <el-option label="等于" value="equal" />
          <el-option label="不等于" value="not_equal" />
        </el-select>
      </el-form-item>

      <!-- 阈值 -->
      <el-form-item label="阈值" prop="threshold">
        <el-input-number
          v-model="form.threshold"
          :min="0"
          :max="100"
          :step="1"
          :precision="1"
          style="width: 100%"
        />
        <span v-if="form.metric === 'node_status'" style="margin-left: 8px; color: #909399; font-size: 12px">
          节点离线触发
        </span>
      </el-form-item>

      <!-- 持续时间 -->
      <el-form-item label="持续时间" prop="duration">
        <el-input-number
          v-model="form.duration"
          :min="1"
          :max="60"
          :step="1"
          style="width: 120px"
        />
        <span style="margin-left: 8px; color: #909399">分钟</span>
      </el-form-item>

      <!-- 通知方式 -->
      <el-form-item label="通知方式" prop="notifyType">
        <el-select v-model="form.notifyType" placeholder="请选择通知方式" style="width: 100%">
          <el-option label="邮件通知" value="email" />
          <el-option label="Webhook" value="webhook" />
        </el-select>
      </el-form-item>

      <!-- 邮件地址 -->
      <el-form-item v-if="form.notifyType === 'email'" label="收件邮箱" prop="notifyTarget">
        <el-input
          v-model="form.notifyTarget"
          placeholder="多个邮箱用逗号分隔"
        />
      </el-form-item>

      <!-- Webhook URL -->
      <el-form-item v-if="form.notifyType === 'webhook'" label="Webhook URL" prop="notifyTarget">
        <el-input
          v-model="form.notifyTarget"
          placeholder="https://example.com/webhook"
        />
      </el-form-item>

      <!-- 启用状态 -->
      <el-form-item label="启用规则">
        <el-switch v-model="form.enabled" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
/**
 * AlertRuleDialog - 告警规则创建/编辑对话框
 * 
 * 支持配置监控指标、条件阈值、持续时间和通知渠道。
 */
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

export interface AlertRule {
  id: string
  name: string
  metric: 'cpu' | 'memory' | 'disk' | 'network' | 'node_status'
  condition: 'greater_than' | 'less_than' | 'equal' | 'not_equal'
  threshold: number
  duration: number
  notifyType: 'email' | 'webhook'
  notifyTarget: string
  enabled: boolean
}

const props = defineProps<{
  visible: boolean
  rule?: AlertRule | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: [rule: Omit<AlertRule, 'id'> & { id?: string }]
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

/** 默认表单 */
const defaultForm = (): Omit<AlertRule, 'id'> => ({
  name: '',
  metric: 'cpu',
  condition: 'greater_than',
  threshold: 80,
  duration: 5,
  notifyType: 'email',
  notifyTarget: '',
  enabled: true,
})

const form = ref<Omit<AlertRule, 'id'>>(defaultForm())

/** 是否为编辑模式 */
const isEdit = computed(() => !!props.rule)

/** 表单校验规则 */
const rules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  metric: [{ required: true, message: '请选择监控指标', trigger: 'change' }],
  condition: [{ required: true, message: '请选择条件', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
  duration: [{ required: true, message: '请输入持续时间', trigger: 'blur' }],
  notifyType: [{ required: true, message: '请选择通知方式', trigger: 'change' }],
  notifyTarget: [{ required: true, message: '请输入通知目标', trigger: 'blur' }],
}

/** 监听 rule 变化，填充表单 */
watch(
  () => props.rule,
  (rule) => {
    if (rule) {
      form.value = {
        name: rule.name,
        metric: rule.metric,
        condition: rule.condition,
        threshold: rule.threshold,
        duration: rule.duration,
        notifyType: rule.notifyType,
        notifyTarget: rule.notifyTarget,
        enabled: rule.enabled,
      }
    } else {
      form.value = defaultForm()
    }
  },
  { immediate: true },
)

/** 关闭时重置表单 */
function handleClose() {
  formRef.value?.resetFields()
  form.value = defaultForm()
}

/** 提交表单 */
async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    emit('submit', {
      ...(props.rule?.id ? { id: props.rule.id } : {}),
      ...form.value,
    })
    emit('update:visible', false)
  } finally {
    submitting.value = false
  }
}
</script>
