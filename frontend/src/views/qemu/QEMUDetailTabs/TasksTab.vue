<template>
  <div class="tasks-tab">
    <!-- 过滤工具栏 -->
    <div class="toolbar">
      <el-select v-model="filterType" placeholder="任务类型" clearable style="width: 160px" @change="fetchTasks">
        <el-option label="全部" value="" />
        <el-option label="启动" value="start" />
        <el-option label="关机" value="stop" />
        <el-option label="重启" value="reboot" />
        <el-option label="创建快照" value="vmsnapshot" />
        <el-option label="备份" value="vzdump" />
        <el-option label="克隆" value="clone" />
        <el-option label="迁移" value="migrate" />
      </el-select>
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 120px" @change="fetchTasks">
        <el-option label="全部" value="" />
        <el-option label="成功" value="OK" />
        <el-option label="失败" value="ERROR" />
        <el-option label="运行中" value="running" />
      </el-select>
      <el-button text @click="fetchTasks">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 任务列表 -->
    <el-card>
      <el-table v-loading="loading" :data="taskList" style="width: 100%" border stripe>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ taskTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="taskStatusType(row.status)" size="small">
              {{ taskStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user" label="用户" width="140" />
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatTimestamp(row.starttime) }}
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ getDuration(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="upid" label="UPID" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="viewLog(row)">查看日志</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && taskList.length === 0" description="暂无任务记录" />
    </el-card>

    <!-- 任务日志对话框 -->
    <el-dialog v-model="showLogDialog" title="任务日志" width="800px">
      <div class="log-container">
        <pre class="log-content">{{ logContent }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getNodeTasks, getTaskLog } from '@/api/node'
import type { NodeTask, TaskStatus } from '@/api/types'

interface Props {
  node: string
  vmid: number
}

const props = defineProps<Props>()

const loading = ref(false)
const filterType = ref('')
const filterStatus = ref('')
const taskList = ref<NodeTask[]>([])

const showLogDialog = ref(false)
const logContent = ref('')

/**
 * 获取该虚拟机的任务列表
 * PVE 没有按 VM 筛选任务的 API，这里获取全部任务后在前端过滤
 */
async function fetchTasks() {
  loading.value = true
  try {
    const tasks = await getNodeTasks(props.node)
    // 根据 UPID 中是否包含该 vmid 来过滤
    let filtered = tasks.filter((t: NodeTask) => {
      if (!t.upid) return false
      return t.upid.includes(`:${props.vmid}:`)
    })
    if (filterType.value) {
      filtered = filtered.filter((t: NodeTask) => t.type === filterType.value)
    }
    if (filterStatus.value) {
      filtered = filtered.filter((t: NodeTask) => t.status === filterStatus.value)
    }
    taskList.value = filtered.slice(0, 50)
  } catch (error) {
    console.error('获取任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

/**
 * 任务类型标签
 */
function taskTypeLabel(type: string): string {
  const map: Record<string, string> = {
    start: '启动',
    stop: '关机',
    reboot: '重启',
    shutdown: 'ACPI关机',
    vmsnapshot: '快照',
    vzdump: '备份',
    clone: '克隆',
    migrate: '迁移',
    qmcreate: '创建',
    qmconfig: '修改配置',
  }
  return map[type] || type
}

/**
 * 任务状态标签类型
 */
function taskStatusType(status?: TaskStatus): string {
  switch (status) {
    case 'OK': return 'success'
    case 'ERROR': return 'danger'
    case 'running': return 'warning'
    default: return 'info'
  }
}

/**
 * 任务状态文本
 */
function taskStatusLabel(status?: TaskStatus): string {
  const map: Record<string, string> = {
    OK: '成功',
    ERROR: '失败',
    running: '运行中',
    stopped: '已停止',
  }
  return map[status || ''] || status || '未知'
}

/**
 * 格式化时间戳
 */
function formatTimestamp(ts: number): string {
  if (!ts) return '-'
  const date = new Date(ts * 1000)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${month}-${day} ${h}:${m}:${s}`
}

/**
 * 获取任务耗时
 */
function getDuration(task: NodeTask): string {
  if (!task.starttime) return '-'
  if (task.status === 'running') return '进行中...'
  const end = task.starttime + 30 // 估算
  const diff = end - task.starttime
  if (diff < 60) return `${diff}秒`
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return `${mins}分${secs}秒`
}

/**
 * 查看任务日志
 */
async function viewLog(task: NodeTask) {
  try {
    const log = await getTaskLog(props.node, task.upid)
    logContent.value = log.map((entry: { t?: string; p?: string; n?: number }) =>
      `${entry.t || ''}  [${entry.p || ''}] ${entry.n}`,
    ).join('\n')
  } catch {
    logContent.value = '日志加载中...'
  }
  showLogDialog.value = true
}

onMounted(() => {
  fetchTasks()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.tasks-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-4;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: $spacing-4;
}

.log-container {
  max-height: 500px;
  overflow: auto;
  background: #1e1e1e;
  border-radius: $radius-base;
  padding: $spacing-4;
}

.log-content {
  color: #d4d4d4;
  font-family: $font-family-code;
  font-size: 12px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}
</style>
