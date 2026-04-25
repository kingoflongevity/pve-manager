<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? t('storage.editStorage') : t('storage.createStorage')"
    width="680px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      label-position="right"
    >
      <!-- 基本信息 -->
      <el-divider content-position="left">{{ t('storage.storageInfo') }}</el-divider>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item :label="t('storage.storageId')" prop="storage">
            <el-input
              v-model="formData.storage"
              :placeholder="t('storage.storageIdPlaceholder')"
              :disabled="isEdit"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('storage.type')" prop="type">
            <el-select
              v-model="formData.type"
              :placeholder="t('storage.selectType')"
              style="width: 100%"
              :disabled="isEdit"
              @change="handleTypeChange"
            >
              <el-option
                v-for="item in storageTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item :label="t('storage.enable')">
            <el-switch v-model="formData.enable" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('storage.isShared')">
            <el-switch v-model="formData.shared" />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 动态表单：根据存储类型显示不同字段 -->
      <el-divider content-position="left">{{ t('storage.type') }}配置</el-divider>

      <!-- Directory 类型 -->
      <template v-if="formData.type === 'dir'">
        <el-form-item :label="t('storage.path')" prop="path">
          <el-input
            v-model="formData.path"
            :placeholder="t('storage.pathPlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- NFS 类型 -->
      <template v-if="formData.type === 'nfs'">
        <el-form-item :label="t('storage.server')" prop="server">
          <el-input
            v-model="formData.server"
            :placeholder="t('storage.serverPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('storage.exportPath')" prop="export">
          <el-input
            v-model="formData.export"
            :placeholder="t('storage.exportPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('storage.options')">
          <el-input
            v-model="formData.options"
            :placeholder="t('storage.optionsPlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- LVM 类型 -->
      <template v-if="formData.type === 'lvm'">
        <el-form-item :label="t('storage.vgname')" prop="vgname">
          <el-input
            v-model="formData.vgname"
            :placeholder="t('storage.vgnamePlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- LVM-Thin 类型 -->
      <template v-if="formData.type === 'lvmthin'">
        <el-form-item :label="t('storage.vgname')" prop="vgname">
          <el-input
            v-model="formData.vgname"
            :placeholder="t('storage.vgnamePlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('storage.thinpool')" prop="thinpool">
          <el-input
            v-model="formData.thinpool"
            :placeholder="t('storage.thinpoolPlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- ZFS 类型 -->
      <template v-if="formData.type === 'zfs'">
        <el-form-item :label="t('storage.pool')" prop="pool">
          <el-input
            v-model="formData.pool"
            :placeholder="t('storage.poolPlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- Ceph 类型 -->
      <template v-if="formData.type === 'rbd'">
        <el-form-item :label="t('storage.monitors')" prop="monitors">
          <el-input
            v-model="formData.monitors"
            :placeholder="t('storage.monitorsPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('storage.cephUsername')">
          <el-input
            v-model="formData.username"
            :placeholder="t('storage.cephUsernamePlaceholder')"
          />
        </el-form-item>
      </template>

      <!-- 内容类型 -->
      <el-divider content-position="left">{{ t('storage.content') }}</el-divider>

      <el-form-item>
        <el-checkbox-group v-model="selectedContentTypes">
          <el-checkbox label="images">{{ t('storage.diskImage') }}</el-checkbox>
          <el-checkbox label="rootdir">{{ t('storage.container') }}</el-checkbox>
          <el-checkbox label="iso">{{ t('storage.iso') }}</el-checkbox>
          <el-checkbox label="backup">{{ t('storage.backup') }}</el-checkbox>
          <el-checkbox label="snippets">{{ t('storage.snippets') }}</el-checkbox>
          <el-checkbox label="vztmpl">{{ t('storage.vztmpl') }}</el-checkbox>
        </el-checkbox-group>
      </el-form-item>

      <!-- 备份存储专属配置 -->
      <template v-if="selectedContentTypes.includes('backup')">
        <el-form-item :label="t('storage.maxBackups')">
          <el-input-number
            v-model="formData.maxBackups"
            :min="0"
            :max="999"
            :placeholder="t('storage.maxBackups')"
          />
        </el-form-item>
      </template>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { CreateStorageParams, StorageBackendType } from '@/types/storage'
import { createStorage, updateStorage } from '@/api/storage'

const { t } = useI18n()

// ============================================================
// Props & Emits
// ============================================================

const props = defineProps<{
  /** 节点名称 */
  node: string
  /** 是否为编辑模式 */
  isEdit?: boolean
  /** 编辑时的存储配置 */
  storageData?: Partial<CreateStorageParams>
}>()

const emit = defineEmits<{
  /** 提交成功后触发 */
  success: []
  /** 关闭对话框 */
  close: []
}>()

// ============================================================
// 状态管理
// ============================================================

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

// 内容类型选项
const selectedContentTypes = ref<string[]>(['images', 'rootdir', 'iso', 'backup', 'snippets'])

// 存储类型选项
const storageTypeOptions = [
  { label: '本地目录 (Directory)', value: 'dir' },
  { label: 'NFS', value: 'nfs' },
  { label: 'LVM', value: 'lvm' },
  { label: 'LVM-Thin', value: 'lvmthin' },
  { label: 'ZFS', value: 'zfs' },
  { label: 'Ceph RBD', value: 'rbd' },
  { label: 'GlusterFS', value: 'glusterfs' },
  { label: 'iSCSI', value: 'iscsi' },
  { label: 'SMB/CIFS', value: 'cifs' },
  { label: 'PBS', value: 'pbs' },
]

// 表单数据
const defaultFormData: CreateStorageParams = {
  storage: '',
  type: 'dir',
  enable: true,
  shared: false,
  path: '',
  server: '',
  export: '',
  options: '',
  vgname: '',
  thinpool: '',
  pool: '',
  monitors: '',
  username: '',
  maxBackups: 1,
  content: 'images,rootdir,iso,backup,snippets',
}

const formData = ref<CreateStorageParams>({ ...defaultFormData })

// ============================================================
// 表单校验规则
// ============================================================

const formRules = computed<FormRules>(() => {
  const rules: FormRules = {
    storage: [
      { required: true, message: '请输入存储 ID', trigger: 'blur' },
      {
        pattern: /^[a-zA-Z0-9_-]+$/,
        message: '存储 ID 只能包含字母、数字、下划线和连字符',
        trigger: 'blur',
      },
    ],
    type: [
      { required: true, message: '请选择存储类型', trigger: 'change' },
    ],
  }

  // 根据存储类型添加动态校验
  switch (formData.value.type) {
    case 'dir':
      rules.path = [
        { required: true, message: '请输入存储路径', trigger: 'blur' },
      ]
      break
    case 'nfs':
      rules.server = [
        { required: true, message: '请输入 NFS 服务器地址', trigger: 'blur' },
      ]
      rules.export = [
        { required: true, message: '请输入 NFS 导出路径', trigger: 'blur' },
      ]
      break
    case 'lvm':
    case 'lvmthin':
      rules.vgname = [
        { required: true, message: '请输入 LVM 卷组名称', trigger: 'blur' },
      ]
      if (formData.value.type === 'lvmthin') {
        rules.thinpool = [
          { required: true, message: '请输入 LVM-Thin 精简池名称', trigger: 'blur' },
        ]
      }
      break
    case 'zfs':
      rules.pool = [
        { required: true, message: '请输入 ZFS 存储池名称', trigger: 'blur' },
      ]
      break
    case 'rbd':
      rules.monitors = [
        { required: true, message: '请输入 Ceph 监控节点', trigger: 'blur' },
      ]
      break
  }

  return rules
})

// 是否为编辑模式
const isEdit = computed(() => props.isEdit ?? false)

// ============================================================
// 监听器
// ============================================================

// 监听 props 变化，初始化表单
watch(
  () => props.storageData,
  (newData) => {
    if (newData && isEdit.value) {
      formData.value = { ...defaultFormData, ...newData }
      // 解析内容类型
      if (newData.content) {
        selectedContentTypes.value = newData.content.split(',')
      }
    }
  },
  { immediate: true },
)

// ============================================================
// 方法
// ============================================================

/**
 * 打开对话框
 */
function openDialog(): void {
  visible.value = true
}

/**
 * 关闭对话框并重置表单
 */
function handleClose(): void {
  visible.value = false
  resetForm()
  emit('close')
}

/**
 * 重置表单
 */
function resetForm(): void {
  formData.value = { ...defaultFormData }
  selectedContentTypes.value = ['images', 'rootdir', 'iso', 'backup', 'snippets']
  formRef.value?.resetFields()
}

/**
 * 存储类型变更时清空不相关的字段
 */
function handleTypeChange(): void {
  // 清空所有动态字段
  formData.value.path = ''
  formData.value.server = ''
  formData.value.export = ''
  formData.value.options = ''
  formData.value.vgname = ''
  formData.value.thinpool = ''
  formData.value.pool = ''
  formData.value.monitors = ''
  formData.value.username = ''
  // 清除校验
  formRef.value?.clearValidate()
}

/**
 * 提交表单
 */
async function handleSubmit(): Promise<void> {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      // 构建提交数据
      const submitData: CreateStorageParams = {
        ...formData.value,
        content: selectedContentTypes.value.join(','),
      }

      // 根据模式调用不同 API
      if (isEdit.value && props.storageData?.storage) {
        await updateStorage(props.node, props.storageData.storage, submitData)
        ElMessage.success('存储配置已更新')
      } else {
        await createStorage(props.node, submitData)
        ElMessage.success('存储创建成功')
      }

      handleClose()
      emit('success')
    } catch (error) {
      console.error('存储操作失败:', error)
      ElMessage.error(isEdit.value ? '更新存储失败' : '创建存储失败')
    } finally {
      submitting.value = false
    }
  })
}

// 暴露 openDialog 方法供父组件调用
defineExpose({
  openDialog,
})
</script>

<style scoped lang="scss">
.el-divider {
  margin: 16px 0;
}

.el-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
</style>
