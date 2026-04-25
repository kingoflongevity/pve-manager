<template>
  <div class="options-tab">
    <!-- 启动/关机行为 -->
    <el-card>
      <template #header>
        <span class="card-title">启动/关机行为</span>
      </template>
      <el-form label-width="120px" @submit.prevent>
        <el-form-item label="开机自启">
          <el-switch
            :model-value="!!config.onboot"
            @change="(v: boolean) => updateConfig({ onboot: v ? 1 : 0 })"
          />
          <span class="form-hint">节点重启后自动启动此虚拟机</span>
        </el-form-item>
        <el-form-item label="启动顺序">
          <el-input
            :model-value="parseStartupOrder()"
            placeholder="例如: order=cdn, 等待 5 秒"
            style="width: 300px"
            @change="(v: string) => updateConfig({ startup: v })"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- QEMU Agent -->
    <el-card>
      <template #header>
        <span class="card-title">QEMU Agent</span>
      </template>
      <el-form label-width="120px">
        <el-form-item label="启用 Agent">
          <el-switch
            :model-value="parseAgentEnabled()"
            @change="(v: boolean) => updateConfig({ agent: v ? '1' : '0' })"
          />
          <span class="form-hint">启用 QEMU Guest Agent 以获取更多信息和执行命令</span>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 保护模式 -->
    <el-card>
      <template #header>
        <span class="card-title">保护模式</span>
      </template>
      <el-form label-width="120px">
        <el-form-item label="启用保护">
          <el-switch
            :model-value="!!config.protection"
            @change="(v: boolean) => updateConfig({ protection: v ? 1 : 0 })"
          />
          <span class="form-hint">保护虚拟机不被意外删除或覆盖</span>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Cloud-init -->
    <el-card>
      <template #header>
        <span class="card-title">Cloud-init</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="状态">{{ hasCloudInit ? '已配置' : '未配置' }}</el-descriptions-item>
        <el-descriptions-item label="用户">
          {{ config.description?.includes('ciuser') ? '已设置' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="IP 配置" :span="2">
          {{ hasCloudInit ? '已配置网络' : '-' }}
        </el-descriptions-item>
      </el-descriptions>
      <div style="margin-top: 16px">
        <el-button size="small">编辑 Cloud-init 配置</el-button>
      </div>
    </el-card>

    <!-- Watchdog -->
    <el-card>
      <template #header>
        <span class="card-title">Watchdog</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="状态">未配置</el-descriptions-item>
        <el-descriptions-item label="动作">-</el-descriptions-item>
      </el-descriptions>
      <div style="margin-top: 16px">
        <el-button size="small">配置 Watchdog</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { setQEMUConfig } from '@/api/qemu'
import type { QEMUConfig } from '@/api/types'

interface Props {
  config: QEMUConfig
  node: string
  vmid: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ refresh: [] }>()

/**
 * 是否配置了 Cloud-init
 */
const hasCloudInit = computed(() => {
  // 检查是否有 IDE/SCSI 的 Cloud-init drive
  const allDrives = [
    ...(props.config.ide || []),
    ...(props.config.scsi || []),
  ]
  return allDrives.some(d => d.includes('cloudinit') || d.includes('type=cloudinit'))
})

/**
 * 解析 Agent 是否启用
 */
function parseAgentEnabled(): boolean {
  if (!props.config.agent) return false
  return props.config.agent.startsWith('1')
}

/**
 * 解析启动顺序
 */
function parseStartupOrder(): string {
  return props.config.startup || props.config.boot || ''
}

/**
 * 更新配置
 */
async function updateConfig(newConfig: Record<string, unknown>) {
  try {
    await setQEMUConfig(props.node, props.vmid, newConfig)
    ElMessage.success('配置已更新')
    emit('refresh')
  } catch (error) {
    console.error('更新配置失败:', error)
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.options-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-6;
}

.card-title {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
}

.form-hint {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  margin-left: $spacing-3;
}
</style>
