<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="flex-between">
          <span>{{ t('qemu.title') }}</span>
          <el-button type="primary" @click="handleCreate">{{ t('qemu.createVM') }}</el-button>
        </div>
      </template>
      <el-table :data="vmList" style="width: 100%" stripe>
        <el-table-column prop="vmid" :label="t('qemu.vmid')" width="100" />
        <el-table-column prop="name" :label="t('common.name')" />
        <el-table-column prop="status" :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : 'danger'">
              {{ row.status === 'running' ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cpu" label="CPU" width="100" />
        <el-table-column prop="memory" :label="t('qemu.memory')" width="120" />
        <el-table-column :label="t('common.actions')" width="200">
          <template #default>
            <el-button link type="primary" size="small">{{ t('qemu.start') }}</el-button>
            <el-button link type="danger" size="small">{{ t('qemu.stop') }}</el-button>
            <el-button link type="warning" size="small">{{ t('qemu.console') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

/**
 * 虚拟机模拟数据（开发阶段占位）
 */
const vmList = ref([
  { vmid: 100, name: 'web-server-01', status: 'running', cpu: 4, memory: '8 GB' },
  { vmid: 101, name: 'db-server-01', status: 'running', cpu: 8, memory: '16 GB' },
  { vmid: 102, name: 'test-vm', status: 'stopped', cpu: 2, memory: '4 GB' },
])

function handleCreate() {
  ElMessage.info('虚拟机创建向导开发中...')
}
</script>
