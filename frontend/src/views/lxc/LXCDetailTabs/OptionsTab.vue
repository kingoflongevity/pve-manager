<template>
  <div class="lxc-options-tab">
    <!-- 启动/关机行为 -->
    <el-card>
      <template #header><span class="card-title">启动/关机行为</span></template>
      <el-form label-width="120px">
        <el-form-item label="开机自启">
          <el-switch
            :model-value="!!config.onboot"
            @change="(v: boolean) => updateConfig({ onboot: v ? 1 : 0 })"
          />
        </el-form-item>
        <el-form-item label="启动顺序">
          <el-input
            :model-value="config.startup || ''"
            placeholder="例如: order=2,up=60"
            style="width: 300px"
            @change="(v: string) => updateConfig({ startup: v })"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 非特权容器 -->
    <el-card>
      <template #header><span class="card-title">容器类型</span></template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="非特权">
          <el-tag :type="config.unprivileged ? 'success' : 'info'" size="small">
            {{ config.unprivileged ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="说明">
          非特权容器以非 root 用户运行，安全性更高
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 保护模式 -->
    <el-card>
      <template #header><span class="card-title">保护模式</span></template>
      <el-form label-width="120px">
        <el-form-item label="启用保护">
          <el-switch
            :model-value="!!config.protection"
            @change="(v: boolean) => updateConfig({ protection: v ? 1 : 0 })"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- DNS 配置 -->
    <el-card>
      <template #header><span class="card-title">DNS</span></template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="主机名">{{ config.hostname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="控制台">{{ config.console ? '启用' : '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { setLXCConfig } from '@/api/lxc'
import type { LXCConfig } from '@/api/types'

interface Props {
  config: LXCConfig
  node: string
  vmid: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ refresh: [] }>()

async function updateConfig(newConfig: Record<string, unknown>) {
  try {
    await setLXCConfig(props.node, props.vmid, newConfig)
    ElMessage.success('配置已更新')
    emit('refresh')
  } catch (error) {
    console.error('更新配置失败:', error)
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;
.lxc-options-tab { display: flex; flex-direction: column; gap: $spacing-6; }
.card-title { font-weight: $font-weight-semibold; font-size: $font-size-md; }
</style>
