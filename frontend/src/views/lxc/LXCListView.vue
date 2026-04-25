<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="flex-between">
          <span>{{ t('lxc.title') }}</span>
          <el-button type="primary" @click="handleCreate">{{ t('lxc.createCT') }}</el-button>
        </div>
      </template>
      <el-table :data="ctList" style="width: 100%" stripe>
        <el-table-column prop="ctid" :label="t('lxc.ctid')" width="100" />
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
            <el-button link type="primary" size="small">{{ t('lxc.start') }}</el-button>
            <el-button link type="danger" size="small">{{ t('lxc.stop') }}</el-button>
            <el-button link type="warning" size="small">{{ t('lxc.console') }}</el-button>
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
 * 容器模拟数据（开发阶段占位）
 */
const ctList = ref([
  { ctid: 200, name: 'nginx-proxy', status: 'running', cpu: 1, memory: '512 MB' },
  { ctid: 201, name: 'redis-cache', status: 'running', cpu: 1, memory: '1 GB' },
])

function handleCreate() {
  ElMessage.info('容器创建向导开发中...')
}
</script>
