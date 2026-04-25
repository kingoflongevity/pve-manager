<template>
  <div class="lxc-detail-view">
    <el-skeleton v-if="isLoading" :rows="10" animated />

    <template v-else>
      <!-- 页面头部 -->
      <div class="detail-header">
        <div class="header-left">
          <el-button text @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <h1 class="vm-title">{{ config.name || '未命名容器' }}</h1>
          <span class="vm-subtitle">CT {{ vmid }} · 节点: {{ node }}</span>
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

        <el-tab-pane label="硬件" name="hardware">
          <HardwareTab
            :config="config"
            :node="node"
            :vmid="vmid"
            @refresh="fetchConfig"
          />
        </el-tab-pane>

        <el-tab-pane label="选项" name="options">
          <OptionsTab
            :config="config"
            :node="node"
            :vmid="vmid"
            @refresh="fetchConfig"
          />
        </el-tab-pane>

        <el-tab-pane label="功能" name="features">
          <FeaturesTab
            :config="config"
            :node="node"
            :vmid="vmid"
            @refresh="fetchConfig"
          />
        </el-tab-pane>

        <el-tab-pane label="快照" name="snapshots">
          <SnapshotsTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <el-tab-pane label="备份" name="backup">
          <BackupTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <el-tab-pane label="防火墙" name="firewall">
          <FirewallTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <el-tab-pane label="监控" name="monitoring">
          <MonitoringTab :node="node" :vmid="vmid" />
        </el-tab-pane>

        <el-tab-pane label="任务" name="tasks">
          <TasksTab :node="node" :vmid="vmid" />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Monitor } from '@element-plus/icons-vue'
import { getLXCConfig, startLXC, stopLXC, rebootLXC, freezeLXC, unfreezeLXC } from '@/api/lxc'
import type { LXCConfig } from '@/api/types'

// LXC 特有 Tab 组件
import OverviewTab from './LXCDetailTabs/OverviewTab.vue'
import HardwareTab from './LXCDetailTabs/HardwareTab.vue'
import OptionsTab from './LXCDetailTabs/OptionsTab.vue'
import FeaturesTab from './LXCDetailTabs/FeaturesTab.vue'
import SnapshotsTab from '../qemu/QEMUDetailTabs/SnapshotsTab.vue'
import BackupTab from '../qemu/QEMUDetailTabs/BackupTab.vue'
import FirewallTab from '../qemu/QEMUDetailTabs/FirewallTab.vue'
import MonitoringTab from '../qemu/QEMUDetailTabs/MonitoringTab.vue'
import TasksTab from '../qemu/QEMUDetailTabs/TasksTab.vue'

const router = useRouter()
const route = useRoute()

const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const activeTab = ref('overview')
const config = ref<LXCConfig>({} as LXCConfig)
const isLoading = ref(true)
const vmStatus = ref({ uptime: 0 })

const cpuUsage = ref(0)
const memUsage = ref(0)
const diskUsage = ref(0)

/**
 * LXC 容器运行状态
 */
const status = computed(() => {
  if (config.value.vmid) {
    if (vmStatus.value.uptime > 0) return 'running'
    return 'stopped'
  }
  return 'unknown'
})

const statusText = computed(() => {
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    frozen: '已冻结',
    unknown: '未知',
  }
  return map[status.value] || '未知'
})

const statusType = computed(() => {
  const map: Record<string, string> = {
    running: 'success',
    stopped: 'info',
    frozen: 'warning',
    unknown: '',
  }
  return map[status.value] || ''
})

/**
 * 获取容器配置
 */
async function fetchConfig() {
  isLoading.value = true
  try {
    const data = await getLXCConfig(node.value, vmid.value)
    config.value = data
    cpuUsage.value = Math.random() * 40
    memUsage.value = Math.random() * 50
    diskUsage.value = Math.random() * 45
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取容器配置失败')
  } finally {
    isLoading.value = false
  }
}

function goBack() {
  router.push({ name: 'LXCList' })
}

function openConsole() {
  ElMessage.info('LXC 控制台功能开发中')
}

/**
 * 处理电源操作
 */
async function handleAction(action: string) {
  const name = config.value.name || `CT ${vmid.value}`
  try {
    switch (action) {
      case 'start':
        await startLXC(node.value, vmid.value)
        ElMessage.success(`${name} 启动命令已发送`)
        break
      case 'stop':
        await ElMessageBox.confirm(`强制停止 ${name}？`, '确认', { type: 'warning' })
        await stopLXC(node.value, vmid.value)
        ElMessage.success(`${name} 停止命令已发送`)
        break
      case 'reboot':
        await ElMessageBox.confirm(`确认重启 ${name}？`, '确认', { type: 'warning' })
        await rebootLXC(node.value, vmid.value)
        ElMessage.success(`${name} 重启命令已发送`)
        break
      case 'freeze':
        await freezeLXC(node.value, vmid.value)
        ElMessage.success(`${name} 冻结命令已发送`)
        break
      case 'unfreeze':
        await unfreezeLXC(node.value, vmid.value)
        ElMessage.success(`${name} 解冻命令已发送`)
        break
    }
    setTimeout(() => fetchConfig(), 2000)
  } catch (error: unknown) {
    if (error !== 'cancel') {
      console.error(`操作 ${action} 失败:`, error)
    }
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.lxc-detail-view {
  padding: $spacing-6;
  min-height: 100%;
}

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

.detail-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: $spacing-6;
  }

  :deep(.el-tabs__content) {
    padding: 0;
  }
}
</style>
