<template>
  <div class="backup-tab">
    <div class="tab-header">
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Download /></el-icon>
        创建备份
      </el-button>
    </div>

    <!-- 备份历史 -->
    <el-card>
      <el-table v-loading="loading" :data="backupList" style="width: 100%" border stripe>
        <el-table-column prop="volid" label="备份文件" min-width="200" show-overflow-tooltip />
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="ctime" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatRelativeTime(row.ctime) }}
          </template>
        </el-table-column>
        <el-table-column prop="protection" label="保护" width="80" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.protection"><Lock /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleRestore(row)">恢复</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && backupList.length === 0" description="暂无备份" />
    </el-card>

    <!-- 创建备份对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建备份" width="500px">
      <el-form :model="backupForm" label-width="100px">
        <el-form-item label="存储">
          <el-select v-model="backupForm.storage" placeholder="选择备份存储" style="width: 100%">
            <el-option label="local" value="local" />
            <el-option label="nfs-backup" value="nfs-backup" />
          </el-select>
        </el-form-item>
        <el-form-item label="模式">
          <el-select v-model="backupForm.mode" style="width: 100%">
            <el-option label="快照 (推荐)" value="snapshot" />
            <el-option label="暂停" value="suspend" />
            <el-option label="停止" value="stop" />
          </el-select>
        </el-form-item>
        <el-form-item label="压缩">
          <el-switch v-model="backupForm.compress" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmCreate">开始备份</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Lock } from '@element-plus/icons-vue'
import { formatBytes, formatRelativeTime } from '@/utils/format'

// 备份占位数据 - 实际应从 PVE 备份 API 获取
interface BackupItem {
  volid: string
  size: number
  ctime: number
  protection?: number
}

defineProps<{
  node: string
  vmid: number
}>()

const loading = ref(false)
const showCreateDialog = ref(false)
const backupList = ref<BackupItem[]>([])
const backupForm = ref({
  storage: 'local',
  mode: 'snapshot',
  compress: true,
})

function confirmCreate() {
  ElMessage.info('备份功能需要后端支持，占位实现')
  showCreateDialog.value = false
}

function handleRestore(row: BackupItem) {
  ElMessageBox.confirm(
    `确认从备份 "${row.volid}" 恢复？当前数据将被覆盖。`,
    '确认恢复',
    { type: 'warning' }
  ).then(() => {
    ElMessage.info('恢复功能需要后端支持，占位实现')
  }).catch(() => {})
}

function handleDelete(row: BackupItem) {
  ElMessageBox.confirm(
    `确认删除备份 "${row.volid}"？`,
    '确认删除',
    { type: 'warning' }
  ).then(() => {
    ElMessage.info('删除备份功能需要后端支持，占位实现')
  }).catch(() => {})
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.backup-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-4;
}

.tab-header {
  display: flex;
  justify-content: flex-start;
}

.action-buttons {
  display: flex;
  gap: $spacing-2;
}
</style>
