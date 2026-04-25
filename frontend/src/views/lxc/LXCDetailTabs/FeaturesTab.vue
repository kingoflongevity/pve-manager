<template>
  <div class="features-tab">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="card-title">容器特性</span>
          <el-button type="primary" size="small" @click="saveFeatures">保存更改</el-button>
        </div>
      </template>
      <el-form label-width="160px">
        <el-form-item label="Nesting (嵌套容器)">
          <el-switch v-model="features.nesting" />
          <span class="form-hint">允许在容器内运行其他容器 (LXC-in-LXC)</span>
        </el-form-item>
        <el-form-item label="Keyctl">
          <el-switch v-model="features.keyctl" />
          <span class="form-hint">允许使用 keyctl 系统调用</span>
        </el-form-item>
        <el-form-item label="FUSE">
          <el-switch v-model="features.fuse" />
          <span class="form-hint">允许使用 FUSE 文件系统</span>
        </el-form-item>
        <el-form-item label="Mount (NFS/CIFS)">
          <el-switch v-model="features.mount" />
          <span class="form-hint">允许挂载 NFS/CIFS 文件系统</span>
        </el-form-item>
        <el-form-item label="CIFS 支持">
          <el-switch v-model="features.cifs" />
          <span class="form-hint">允许使用 CIFS/SMB 共享</span>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 特性说明 -->
    <el-card>
      <template #header><span class="card-title">说明</span></template>
      <el-alert
        title="容器特性安全提示"
        description="启用某些特性（如 Nesting、FUSE）可能会降低容器的隔离性和安全性。请根据实际需求谨慎开启。"
        type="info"
        show-icon
        :closable="false"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getLXCConfig, setLXCConfig } from '@/api/lxc'
import type { LXCConfig } from '@/api/types'

interface Props {
  config: LXCConfig
  node: string
  vmid: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ refresh: [] }>()

/**
 * 特性开关状态
 */
const features = ref({
  nesting: false,
  keyctl: false,
  fuse: false,
  mount: false,
  cifs: false,
})

/**
 * 解析当前特性配置
 */
function parseFeatures() {
  const raw = props.config.features || ''
  const parts = raw.split(',')
  for (const part of parts) {
    const [key, value] = part.split('=')
    if (key && value !== undefined && key in features.value) {
      (features.value as Record<string, boolean>)[key] = value === '1'
    }
  }
}

/**
 * 保存特性更改
 */
async function saveFeatures() {
  const featureStr = Object.entries(features.value)
    .filter(([, v]) => v)
    .map(([k]) => `${k}=1`)
    .join(',')

  try {
    await setLXCConfig(props.node, props.vmid, { features: featureStr || '0' })
    ElMessage.success('特性配置已保存')
    emit('refresh')
  } catch (error) {
    console.error('保存特性失败:', error)
  }
}

onMounted(() => {
  parseFeatures()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;
.features-tab { display: flex; flex-direction: column; gap: $spacing-6; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title { font-weight: $font-weight-semibold; font-size: $font-size-md; }
.form-hint { font-size: $font-size-xs; color: $color-text-secondary; margin-left: $spacing-3; }
</style>
