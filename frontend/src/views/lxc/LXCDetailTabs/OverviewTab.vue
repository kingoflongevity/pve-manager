<template>
  <div class="lxc-overview-tab">
    <!-- 基本信息 -->
    <el-card>
      <template #header><span class="card-title">基本信息</span></template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="CT ID">{{ config.vmid }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ config.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusType" size="small">{{ statusText }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="运行时间">{{ formatUptime(uptime) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ config.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="非特权">
          <el-tag :type="config.unprivileged ? 'success' : 'info'" size="small">
            {{ config.unprivileged ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="架构">{{ config.arch || 'amd64' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 资源仪表盘 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="card-title">资源使用</span>
          <el-button text size="small"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </template>
      <div class="gauges-grid">
        <ResourceGauge :value="cpuUsage" :max="100" label="CPU 使用率" />
        <ResourceGauge :value="memUsage" :max="100" label="内存使用率" />
        <ResourceGauge :value="diskUsage" :max="100" label="RootFS 使用" />
        <ResourceGauge :value="0" :max="100" label="Swap 使用" unit="MB" />
      </div>
    </el-card>

    <!-- 操作按钮 -->
    <el-card>
      <template #header><span class="card-title">电源操作</span></template>
      <div class="actions-grid">
        <el-button type="success" :disabled="status === 'running'" @click="$emit('action', 'start')">
          <el-icon><VideoPlay /></el-icon> 启动
        </el-button>
        <el-button type="danger" :disabled="status !== 'running'" @click="$emit('action', 'stop')">
          <el-icon><VideoPause /></el-icon> 停止
        </el-button>
        <el-button type="warning" :disabled="status !== 'running'" @click="$emit('action', 'reboot')">
          <el-icon><RefreshRight /></el-icon> 重启
        </el-button>
        <el-button type="warning" :disabled="status !== 'running'" plain @click="$emit('action', 'freeze')">
          <el-icon><Snow /></el-icon> 冻结
        </el-button>
        <el-button type="success" :disabled="status !== 'frozen'" plain @click="$emit('action', 'unfreeze')">
          <el-icon><CaretRight /></el-icon> 解冻
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Refresh, VideoPlay, VideoPause, RefreshRight, Snow, CaretRight } from '@element-plus/icons-vue'
import ResourceGauge from '@/components/common/ResourceGauge.vue'
import { formatUptime } from '@/utils/format'
import type { LXCConfig } from '@/api/types'

interface Props {
  config: LXCConfig
  status: string
  uptime: number
  cpuUsage: number
  memUsage: number
  diskUsage: number
}

const props = defineProps<Props>()
defineEmits<{ action: [action: string] }>()

const statusText = computed(() => {
  const map: Record<string, string> = { running: '运行中', stopped: '已停止', frozen: '已冻结', unknown: '未知' }
  return map[props.status] || '未知'
})

const statusType = computed(() => {
  const map: Record<string, string> = { running: 'success', stopped: 'info', frozen: 'warning', unknown: '' }
  return map[props.status] || ''
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;
.lxc-overview-tab { display: flex; flex-direction: column; gap: $spacing-6; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title { font-weight: $font-weight-semibold; font-size: $font-size-md; }
.gauges-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: $spacing-6; padding: $spacing-4 0; }
.actions-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: $spacing-4; padding: $spacing-4 0; }
</style>
