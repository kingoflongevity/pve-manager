<template>
  <div class="snapshots-tab">
    <div class="tab-header">
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Camera /></el-icon>
        创建快照
      </el-button>
    </div>

    <!-- 快照列表 -->
    <el-card>
      <el-table
        v-loading="loading"
        :data="snapshotList"
        style="width: 100%"
        border
        stripe
        row-key="name"
        default-expand-all
      >
        <el-table-column prop="name" label="快照名称" min-width="160">
          <template #default="{ row }">
            <div class="snapshot-name">
              <el-icon v-if="row.parent" class="parent-icon"><Connection /></el-icon>
              <span class="name-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatSnapTime(row.snaptime) }}
          </template>
        </el-table-column>
        <el-table-column label="VM 状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.vmstate ? 'success' : 'info'" size="small">
              {{ row.vmstate ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="包含 RAM" width="90" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.vmstate"><Select /></el-icon>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleRollback(row)">
                回滚
              </el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && snapshotList.length === 0" description="暂无快照" />
    </el-card>

    <!-- 创建快照对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建快照" width="500px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="快照名称" required>
          <el-input v-model="createForm.snapname" placeholder="输入快照名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="可选描述信息"
          />
        </el-form-item>
        <el-form-item label="包含 RAM">
          <el-switch v-model="includeRam" />
          <span class="form-hint">保存当前内存状态，回滚后可直接恢复工作状态</span>
        </el-form-item>
        <el-form-item label="包含磁盘">
          <el-switch v-model="includeDisk" :disabled="true" />
          <span class="form-hint">磁盘状态始终会包含</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="confirmCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Camera, Connection, Select } from '@element-plus/icons-vue'
import { listSnapshots, createSnapshot, rollbackSnapshot, deleteSnapshot } from '@/api/qemu'
import type { QEMUSnapshot } from '@/api/types'

interface Props {
  node: string
  vmid: number
}

const props = defineProps<Props>()

const loading = ref(false)
const creating = ref(false)
const snapshotList = ref<QEMUSnapshot[]>([])
const showCreateDialog = ref(false)
const includeRam = ref(false)
const includeDisk = ref(true)
const createForm = ref({
  snapname: '',
  description: '',
})

/**
 * 获取快照列表
 */
async function fetchSnapshots() {
  loading.value = true
  try {
    const data = await listSnapshots(props.node, props.vmid)
    snapshotList.value = data || []
  } catch (error) {
    console.error('获取快照列表失败:', error)
  } finally {
    loading.value = false
  }
}

/**
 * 格式化快照时间
 */
function formatSnapTime(snaptime: number): string {
  if (!snaptime) return '-'
  const date = new Date(snaptime * 1000)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${month}-${day} ${hours}:${minutes}`
}

/**
 * 确认创建快照
 */
async function confirmCreate() {
  if (!createForm.value.snapname.trim()) {
    ElMessage.warning('请输入快照名称')
    return
  }
  creating.value = true
  try {
    await createSnapshot(props.node, props.vmid, {
      snapname: createForm.value.snapname.trim(),
      description: createForm.value.description.trim(),
      vmstate: includeRam.value ? 1 : 0,
    })
    ElMessage.success('快照创建命令已发送')
    showCreateDialog.value = false
    createForm.value = { snapname: '', description: '' }
    includeRam.value = false
    // 快照创建是异步任务，延迟刷新
    setTimeout(() => fetchSnapshots(), 3000)
  } catch (error) {
    console.error('创建快照失败:', error)
  } finally {
    creating.value = false
  }
}

/**
 * 回滚到指定快照
 */
function handleRollback(row: QEMUSnapshot) {
  ElMessageBox.confirm(
    `确认回滚到快照 "${row.name}"？当前状态将被丢弃。`,
    '确认回滚',
    {
      confirmButtonText: '确认回滚',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await rollbackSnapshot(props.node, props.vmid, row.name)
      ElMessage.success('回滚命令已发送')
      setTimeout(() => fetchSnapshots(), 3000)
    } catch (error) {
      console.error('回滚失败:', error)
    }
  }).catch(() => {})
}

/**
 * 删除快照
 */
function handleDelete(row: QEMUSnapshot) {
  ElMessageBox.confirm(
    `确认删除快照 "${row.name}"？`,
    '确认删除',
    {
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteSnapshot(props.node, props.vmid, row.name)
      ElMessage.success('快照删除命令已发送')
      setTimeout(() => fetchSnapshots(), 3000)
    } catch (error) {
      console.error('删除快照失败:', error)
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchSnapshots()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.snapshots-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-4;
}

.tab-header {
  display: flex;
  justify-content: flex-start;
}

.snapshot-name {
  display: flex;
  align-items: center;
  gap: $spacing-2;

  .parent-icon {
    color: $color-text-secondary;
    font-size: 12px;
  }

  .name-text {
    font-weight: $font-weight-medium;
  }
}

.action-buttons {
  display: flex;
  gap: $spacing-2;
}

.form-hint {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  margin-left: $spacing-3;
}
</style>
