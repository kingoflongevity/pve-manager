<template>
  <div class="app-deployments">
    <div class="page-header">
      <h2>部署管理</h2>
      <el-button size="small" @click="fetchDeployments" :loading="loading">刷新</el-button>
    </div>

    <el-table :data="deployments" stripe>
      <el-table-column prop="name" label="实例名称" min-width="150" />
      <el-table-column label="类型" width="80">
        <template #default="{ row }">
          <el-tag :type="row.type === 'lxc' ? 'success' : 'warning'" size="small">{{ row.type === 'lxc' ? 'LXC' : 'QEMU' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="node" label="节点" width="100" />
      <el-table-column label="进度" min-width="200">
        <template #default="{ row }">
          <div class="progress-cell">
            <el-progress :percentage="row.progress || 0" :status="progressStatus(row.status)" :stroke-width="16" />
            <span class="step-info">{{ row.step_info || '-' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="VMID" width="80" align="center">
        <template #default="{ row }">{{ row.vmid || '-' }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="170" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
          <el-button v-if="row.status === 'running' || row.status === 'pending'" size="small" type="danger" @click="cancelDeployment(row)">取消</el-button>
          <el-button v-else-if="row.status === 'completed'" size="small" type="danger" @click="uninstallDeployment(row)">卸载</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="detailDialog" title="部署详情" width="600px">
      <div v-if="currentDeploy">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="实例名称">{{ currentDeploy.name }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentDeploy.type === 'lxc' ? 'LXC 容器' : 'QEMU 虚拟机' }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ currentDeploy.node }}</el-descriptions-item>
          <el-descriptions-item label="VMID">{{ currentDeploy.vmid || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentDeploy.status)" size="small">{{ statusLabel(currentDeploy.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="进度">{{ currentDeploy.progress || 0 }}%</el-descriptions-item>
          <el-descriptions-item label="当前步骤" :span="2">{{ currentDeploy.step_info || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentDeploy.created_at }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ currentDeploy.started_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="完成时间" :span="2">{{ currentDeploy.completed_at || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="currentDeploy.error_msg" label="错误信息" :span="2">
            <span style="color: #f56c6c">{{ currentDeploy.error_msg }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="currentDeploy.status === 'running' || currentDeploy.status === 'pending'" class="deploy-progress">
          <el-progress :percentage="currentDeploy.progress || 0" :status="currentDeploy.progress === 100 ? 'success' : ''" />
          <p class="step-text">{{ currentDeploy.step_info || '准备中...' }}</p>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAppDeployments, deleteAppDeployment } from '@/api/appStore'
import type { AppDeployment } from '@/api/appStore'

const deployments = ref<AppDeployment[]>([])
const detailDialog = ref(false)
const currentDeploy = ref<AppDeployment | null>(null)
const loading = ref(false)
let refreshTimer: number | null = null

function statusTag(s: string) {
  const m: Record<string, string> = { pending: 'info', running: 'warning', completed: 'success', failed: 'danger', cancelled: 'info', uninstalled: 'info' }
  return m[s] || 'info'
}
function statusLabel(s: string) {
  const m: Record<string, string> = { pending: '等待中', running: '部署中', completed: '已完成', failed: '失败', cancelled: '已取消', uninstalled: '已卸载' }
  return m[s] || s
}
function progressStatus(s: string) {
  if (s === 'running' || s === 'pending') return undefined
  if (s === 'completed') return 'success'
  if (s === 'failed' || s === 'cancelled') return 'exception'
  return undefined
}

onMounted(() => {
  fetchDeployments()
  refreshTimer = window.setInterval(() => {
    const hasRunning = deployments.value.some(d => d.status === 'running' || d.status === 'pending')
    if (hasRunning) fetchDeployments()
  }, 3000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

async function fetchDeployments() {
  loading.value = true
  try {
    deployments.value = await getAppDeployments()
  } catch { ElMessage.error('获取部署列表失败') }
  finally { loading.value = false }
}

async function cancelDeployment(row: AppDeployment) {
  ElMessageBox.confirm('确定取消该部署任务？', '提示').then(async () => {
    try {
      await deleteAppDeployment(row.id)
      ElMessage.success('已取消')
      fetchDeployments()
    } catch { ElMessage.error('取消失败') }
  }).catch(() => {})
}

async function uninstallDeployment(row: AppDeployment) {
  ElMessageBox.confirm(`确定卸载实例 "${row.name}"？此操作不可恢复`, '警告', { type: 'warning' }).then(async () => {
    try {
      await deleteAppDeployment(row.id)
      ElMessage.success('已卸载')
      fetchDeployments()
    } catch { ElMessage.error('卸载失败') }
  }).catch(() => {})
}

function viewDetail(row: AppDeployment) {
  currentDeploy.value = row
  detailDialog.value = true
}
</script>

<style scoped>
.app-deployments { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
.progress-cell { display: flex; flex-direction: column; gap: 4px; }
.progress-cell .el-progress { margin-bottom: 2px; }
.step-info { font-size: 12px; color: #909399; }
.deploy-progress { margin-top: 20px; }
.step-text { margin-top: 8px; font-size: 14px; color: #606266; text-align: center; }
</style>
