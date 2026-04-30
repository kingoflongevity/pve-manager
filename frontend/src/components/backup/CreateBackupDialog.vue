<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useBackupStore, type CreateBackupForm } from '@/stores/backup'
import type { TargetResource } from '@/stores/backup'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submitted: []
}>()

const backupStore = useBackupStore()

/** 表单引用 */
const formRef = ref()

/** 表单数据 */
const form = reactive<CreateBackupForm>({
  targetId: null,
  mode: 'snapshot',
  storageTarget: '',
  compression: 'gzip',
  scheduleType: 'daily',
  cronExpression: '0 2 * * *',
  retentionCount: 7,
  notifyEmail: false,
})

/** 表单验证规则 */
const rules = {
  targetId: [{ required: true, message: '请选择目标资源', trigger: 'change' }],
  storageTarget: [{ required: true, message: '请选择存储目标', trigger: 'change' }],
  retentionCount: [
    { required: true, message: '请输入保留数量', trigger: 'blur' },
    { type: 'number', min: 1, max: 100, message: '保留数量需在 1-100 之间', trigger: 'blur' },
  ],
  cronExpression: [
    {
      required: true,
      message: '请输入 Cron 表达式',
      trigger: 'blur',
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (form.scheduleType === 'custom' && !value.trim()) {
          callback(new Error('请输入 Cron 表达式'))
        } else {
          callback()
        }
      },
    },
  ],
}

/** 表单是否提交中 */
const submitting = ref(false)

/** 目标资源选项（按 VM / CT 分组） */
const targetOptions = computed(() => {
  const allTargets: TargetResource[] = backupStore.jobs.reduce<TargetResource[]>((acc, job) => {
    const exists = acc.some((t) => t.id === job.targetId)
    if (!exists) {
      acc.push({
        id: job.targetId,
        name: job.targetName,
        type: job.targetType,
        node: '',
      })
    }
    return acc
  }, [])

  // 补充额外的 mock targets
  const mockTargets: TargetResource[] = [
    { id: 100, name: 'web-server-01', type: 'vm', node: 'pve-node-01' },
    { id: 101, name: 'db-server-01', type: 'vm', node: 'pve-node-01' },
    { id: 102, name: 'test-vm-01', type: 'vm', node: 'pve-node-02' },
    { id: 103, name: 'api-gateway', type: 'vm', node: 'pve-node-02' },
    { id: 200, name: 'redis-cache', type: 'ct', node: 'pve-node-01' },
    { id: 201, name: 'nginx-proxy', type: 'ct', node: 'pve-node-02' },
  ]

  const merged = [...allTargets]
  for (const mt of mockTargets) {
    if (!merged.some((t) => t.id === mt.id)) {
      merged.push(mt)
    }
  }

  return merged
})

/** VM 目标列表 */
const vmTargets = computed(() => targetOptions.value.filter((t) => t.type === 'vm'))

/** CT 目标列表 */
const ctTargets = computed(() => targetOptions.value.filter((t) => t.type === 'ct'))

/** 存储目标选项 */
const storageOptions = computed(() => backupStore.storageTargets)

/** 调度类型选项 */
const scheduleOptions = [
  { label: '一次性', value: 'once' },
  { label: '每天', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' },
  { label: '自定义 Cron', value: 'custom' },
] as const

/** 备份模式选项 */
const modeOptions = [
  { label: '快照', value: 'snapshot', desc: '在线备份，不影响业务' },
  { label: '停机', value: 'stop', desc: '关闭虚拟机后备份，数据一致性最佳' },
  { label: '挂起', value: 'suspend', desc: '挂起虚拟机后备份，兼顾速度与一致性' },
] as const

/** 压缩选项 */
const compressionOptions = [
  { label: '无', value: 'none', desc: '不压缩，速度最快' },
  { label: 'gzip', value: 'gzip', desc: '通用压缩，兼容性好' },
  { label: 'lz4', value: 'lz4', desc: '极快解压，适合高性能场景' },
  { label: 'zstd', value: 'zstd', desc: '高压缩比，推荐' },
] as const

/** 重置表单 */
function resetForm() {
  form.targetId = null
  form.mode = 'snapshot'
  form.storageTarget = ''
  form.compression = 'gzip'
  form.scheduleType = 'daily'
  form.cronExpression = '0 2 * * *'
  form.retentionCount = 7
  form.notifyEmail = false
  formRef.value?.resetFields()
}

/** 对话框打开时重置表单 */
function handleOpen() {
  resetForm()
}

/** 对话框关闭时触发 */
function handleClose() {
  emit('update:visible', false)
}

/** 提交表单 */
async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const success = await backupStore.createBackup({ ...form })
    if (success) {
      emit('submitted')
      emit('update:visible', false)
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    title="创建备份任务"
    width="600px"
    :close-on-click-modal="false"
    @open="handleOpen"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      label-position="right"
    >
      <!-- 目标资源 -->
      <el-form-item label="目标资源" prop="targetId">
        <el-select
          v-model="form.targetId"
          placeholder="请选择虚拟机或容器"
          style="width: 100%"
          filterable
        >
          <el-option-group label="虚拟机 (QEMU)">
            <el-option
              v-for="vm in vmTargets"
              :key="vm.id"
              :label="`${vm.name} (VM ${vm.id})`"
              :value="vm.id"
            />
          </el-option-group>
          <el-option-group label="容器 (LXC)">
            <el-option
              v-for="ct in ctTargets"
              :key="ct.id"
              :label="`${ct.name} (CT ${ct.id})`"
              :value="ct.id"
            />
          </el-option-group>
        </el-select>
      </el-form-item>

      <!-- 备份模式 -->
      <el-form-item label="备份模式">
        <el-radio-group v-model="form.mode" class="mode-group">
          <el-radio
            v-for="mode in modeOptions"
            :key="mode.value"
            :value="mode.value"
            border
            class="mode-radio"
          >
            <div class="mode-content">
              <div class="mode-label">{{ mode.label }}</div>
              <div class="mode-desc">{{ mode.desc }}</div>
            </div>
          </el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- 存储目标 -->
      <el-form-item label="存储目标" prop="storageTarget">
        <el-select
          v-model="form.storageTarget"
          placeholder="请选择存储目标"
          style="width: 100%"
        >
          <el-option
            v-for="storage in storageOptions"
            :key="storage.id"
            :label="storage.name"
            :value="storage.id"
          >
            <div class="storage-option">
              <span>{{ storage.name }}</span>
              <span class="storage-type">{{ storage.type }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- 压缩方式 -->
      <el-form-item label="压缩方式">
        <el-select v-model="form.compression" style="width: 100%">
          <el-option
            v-for="comp in compressionOptions"
            :key="comp.value"
            :label="comp.label"
            :value="comp.value"
          >
            <div class="compression-option">
              <span>{{ comp.label }}</span>
              <span class="compression-desc">{{ comp.desc }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- 调度策略 -->
      <el-form-item label="调度策略">
        <el-select v-model="form.scheduleType" style="width: 100%">
          <el-option
            v-for="schedule in scheduleOptions"
            :key="schedule.value"
            :label="schedule.label"
            :value="schedule.value"
          />
        </el-select>
      </el-form-item>

      <!-- Cron 表达式（仅自定义时显示） -->
      <el-form-item v-if="form.scheduleType === 'custom'" label="Cron 表达式" prop="cronExpression">
        <el-input
          v-model="form.cronExpression"
          placeholder="例如: 0 */6 * * *"
        />
        <div class="form-tip">
          格式: 分 时 日 月 星期，<a href="https://crontab.guru/" target="_blank">在线生成工具</a>
        </div>
      </el-form-item>

      <!-- 保留数量 -->
      <el-form-item label="保留数量" prop="retentionCount">
        <el-input-number
          v-model="form.retentionCount"
          :min="1"
          :max="100"
          style="width: 100%"
        />
        <div class="form-tip">保留最近的 N 个备份，超出后自动删除旧备份</div>
      </el-form-item>

      <!-- 邮件通知 -->
      <el-form-item label="邮件通知">
        <el-switch
          v-model="form.notifyEmail"
          active-text="完成后发送通知"
        />
      </el-form-item>
    </el-form>

    <!-- 底部按钮 -->
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        创建备份
      </el-button>
    </template>
  </el-dialog>
</template>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

// ============================================================
// 备份模式选择
// ============================================================

.mode-group {
  display: flex;
  flex-direction: column;
  gap: $spacing-2;
  width: 100%;
}

.mode-radio {
  margin-right: 0 !important;
  width: 100%;

  :deep(.el-radio__label) {
    width: 100%;
    padding-left: $spacing-2;
  }
}

.mode-content {
  display: flex;
  flex-direction: column;
  gap: $spacing-1;

  .mode-label {
    font-weight: $font-weight-medium;
    font-size: $font-size-base;
  }

  .mode-desc {
    font-size: $font-size-xs;
    color: $color-text-secondary;
  }
}

// ============================================================
// 下拉选项样式
// ============================================================

.storage-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;

  .storage-type {
    font-size: $font-size-xs;
    color: $color-text-secondary;
    background: $color-bg-elevated;
    padding: $spacing-1 $spacing-2;
    border-radius: $radius-xs;
    text-transform: uppercase;
  }
}

.compression-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;

  .compression-desc {
    font-size: $font-size-xs;
    color: $color-text-secondary;
  }
}

// ============================================================
// 提示文字
// ============================================================

.form-tip {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  margin-top: $spacing-2;

  a {
    color: $color-primary;
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}
</style>
