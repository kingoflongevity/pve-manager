<template>
  <div class="qemu-detail-view">
    <el-skeleton v-if="isLoading" :rows="10" animated />

    <template v-else>
      <!-- 页面头部 -->
      <div class="detail-header">
        <div class="header-left">
          <el-button text @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <h1 class="vm-title">{{ config.name || '未命名虚拟机' }}</h1>
          <span class="vm-subtitle">VM {{ vmid }} · 节点: {{ node }}</span>
        </div>
        <div class="header-right">
          <el-tag :type="statusType" effect="dark" size="large" class="status-badge">
            <span class="status-dot"></span>
            {{ statusText }}
          </el-tag>
          <el-button type="primary" @click="openConsole">
            <el-icon><Monitor /></el-icon>
            控制台
          </el-button>
        </div>
      </div>

      <!-- Tab 导航 -->
      <el-tabs v-model="activeTab" type="border-card" class="detail-tabs">
        <!-- Tab 1: 概览 -->
        <el-tab-pane label="概览" name="overview">
          <OverviewTab
            :config="config"
            :status="status"
            :uptime="vmStatus.uptime"
            :cpu-usage="cpuUsage"
            :mem-usage="memUsage"
            :disk-usage="diskUsage"
            @action="handleAction"
          />
        </el-tab-pane>

        <!-- Tab 2: 硬件 -->
        <el-tab-pane label="硬件" name="hardware">
          <HardwareTab
            :config="config"
            :node="node"
            :vmid="vmid"
            @refresh="fetchConfig"
          />
        </el-tab-pane>

        <!-- Tab 3: 选项 -->
        <el-tab-pane label="选项" name="options">
          <OptionsTab
            :config="config"
            :node="node"
            :vmid="vmid"
            @refresh="fetchConfig"
          />
        </el-tab-pane>

        <!-- Tab 4: 快照 -->
        <el-tab-pane label="快照" name="snapshots">
          <SnapshotsTab
            :node="node"
            :vmid="vmid"
          />
        </el-tab-pane>

        <!-- Tab 5: 备份 -->
        <el-tab-pane label="备份" name="backup">
          <BackupTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <!-- Tab 6: 防火墙 -->
        <el-tab-pane label="防火墙" name="firewall">
          <FirewallTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <!-- Tab 7: 监控 -->
        <el-tab-pane label="监控" name="monitoring">
          <MonitoringTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <!-- Tab 8: 任务 -->
        <el-tab-pane label="任务" name="tasks">
          <TasksTab :node="node" :vmid="vmid" />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Monitor } from '@element-plus/icons-vue'
import { getQEMUConfig, startQEMU, stopQEMU, rebootQEMU, shutdownQEMU, suspendQEMU, resumeQEMU } from '@/api/qemu'
import type { QEMUConfig } from '@/api/types'

// 导入 Tab 子组件
import OverviewTab from './QEMUDetailTabs/OverviewTab.vue'
import HardwareTab from './QEMUDetailTabs/HardwareTab.vue'
import OptionsTab from './QEMUDetailTabs/OptionsTab.vue'
import SnapshotsTab from './QEMUDetailTabs/SnapshotsTab.vue'
import BackupTab from './QEMUDetailTabs/BackupTab.vue'
import FirewallTab from './QEMUDetailTabs/FirewallTab.vue'
import MonitoringTab from './QEMUDetailTabs/MonitoringTab.vue'
import TasksTab from './QEMUDetailTabs/TasksTab.vue'

const router = useRouter()
const route = useRoute()

const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const activeTab = ref('overview')
const config = ref<QEMUConfig>({} as QEMUConfig)
const isLoading = ref(true)
const vmStatus = ref({ uptime: 0 })

// 当前资源使用率（模拟，实际应从 API 获取实时数据）
const cpuUsage = ref(0)
const memUsage = ref(0)
const diskUsage = ref(0)

/**
 * 虚拟机运行状态
 */
const status = computed(() => {
  if (config.value.vmid) {
    // 有配置说明存在，通过 uptime 判断
    if (vmStatus.value.uptime > 0) return 'running'
    return 'stopped'
  }
  return 'unknown'
})

/**
 * 状态文本
 */
const statusText = computed(() => {
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    paused: '已暂停',
    suspended: '已挂起',
    unknown: '未知',
  }
  return map[status.value] || '未知'
})

/**
 * 状态 Tag 类型
 */
const statusType = computed(() => {
  const map: Record<string, string> = {
    running: 'success',
    stopped: 'info',
    paused: 'warning',
    suspended: 'warning',
    unknown: '',
  }
  return map[status.value] || ''
})

/**
 * 获取虚拟机配置
 */
async function fetchConfig() {
  isLoading.value = true
  try {
    const data = await getQEMUConfig(node.value, vmid.value)
    config.value = data
    vmStatus.value.uptime = 0 // 实际需要从 status 接口获取
    // 模拟当前使用率数据
    cpuUsage.value = Math.random() * 60
    memUsage.value = Math.random() * 70
    diskUsage.value = Math.random() * 50
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取虚拟机配置失败')
  } finally {
    isLoading.value = false
  }
}

/**
 * 返回虚拟机列表
 */
function goBack() {
  router.push({ name: 'QEMUList' })
}

/**
 * 打开控制台
 */
function openConsole() {
  router.push({
    name: 'QEMUConsole',
    params: { node: node.value, vmid: vmid.value.toString() },
  })
}

/**
 * 处理电源操作
 */
async function handleAction(action: string) {
  const name = config.value.name || `VM ${vmid.value}`

  try {
    switch (action) {
      case 'start':
        await startQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 启动命令已发送`)
        break
      case 'stop':
        await ElMessageBox.confirm(`强制关闭 ${name}？可能导致数据丢失。`, '确认', {
          confirmButtonText: '确认关机',
          cancelButtonText: '取消',
          type: 'warning',
        })
        await stopQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 关机命令已发送`)
        break
      case 'reboot':
        await ElMessageBox.confirm(`确认重启 ${name}？`, '确认', {
          confirmButtonText: '确认重启',
          cancelButtonText: '取消',
          type: 'warning',
        })
        await rebootQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 重启命令已发送`)
        break
      case 'shutdown':
        await ElMessageBox.confirm(`通过 ACPI 关机 ${name}？`, '确认', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          type: 'info',
        })
        await shutdownQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 关机命令已发送`)
        break
      case 'suspend':
        await suspendQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 挂起命令已发送`)
        break
      case 'resume':
        await resumeQEMU(node.value, vmid.value)
        ElMessage.success(`${name} 恢复命令已发送`)
        break
      default:
        break
    }
    // 操作成功后刷新配置
    setTimeout(() => fetchConfig(), 2000)
  } catch (error: unknown) {
    // 用户取消或 API 错误
    if (error !== 'cancel') {
      console.error(`操作 ${action} 失败:`, error)
    }
  }
}

onMounted(() => {
  fetchConfig()
})

onUnmounted(() => {
  // 清理资源
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.qemu-detail-view {
  padding: $spacing-6;
  min-height: 100%;
}

// 页面头部
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-6;

  .header-left {
    display: flex;
    align-items: center;
    gap: $spacing-4;

    .vm-title {
      font-size: $font-size-2xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      margin: 0;
    }

    .vm-subtitle {
      font-size: $font-size-sm;
      color: $color-text-secondary;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: $spacing-4;
  }
}

// 状态 Badge
.status-badge {
  display: flex;
  align-items: center;
  gap: 6px;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }

  &.el-tag--success .status-dot {
    animation: pulse 2s ease-in-out infinite;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(1.3); }
}

// Tab 样式
.detail-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: $spacing-6;
  }

  :deep(.el-tabs__content) {
    padding: 0;
  }
}
</style>
