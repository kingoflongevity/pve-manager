<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="flex-between">
          <span>{{ t('storage.title') }}</span>
          <el-button type="primary" @click="handleAdd">{{ t('storage.addStorage') }}</el-button>
        </div>
      </template>
      <el-table :data="storageList" style="width: 100%" stripe>
        <el-table-column prop="storage" label="ID" width="150" />
        <el-table-column prop="type" :label="t('storage.type')" width="120" />
        <el-table-column prop="content" :label="t('storage.content')" />
        <el-table-column prop="used" :label="t('storage.used')" width="150" />
        <el-table-column prop="total" :label="t('storage.total')" width="150" />
        <el-table-column prop="active" :label="t('storage.shared')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.shared ? 'success' : 'info'">
              {{ row.shared ? '是' : '否' }}
            </el-tag>
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
 * 存储模拟数据（开发阶段占位）
 */
const storageList = ref([
  { storage: 'local', type: 'dir', content: 'images,rootdir,iso,vztmpl,backup', used: '120 GB', total: '500 GB', shared: false, active: true },
  { storage: 'local-lvm', type: 'lvm', content: 'images,rootdir', used: '80 GB', total: '200 GB', shared: false, active: true },
  { storage: 'nfs-backup', type: 'nfs', content: 'backup,iso', used: '250 GB', total: '1 TB', shared: true, active: true },
])

function handleAdd() {
  ElMessage.info('添加存储功能开发中...')
}
</script>
