<template>
  <div class="lxc-list-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">{{ t('lxc.title') }}</h1>
        <p class="page-description">管理和监控所有 LXC 容器实例</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建容器
        </el-button>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <BatchOperationBar />

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <!-- 批量操作 -->
        <el-dropdown :disabled="selectedRows.length === 0" @command="handleBatchCommand">
          <el-button :disabled="selectedRows.length === 0">
            批量操作
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="start">
                <el-icon><VideoPlay /></el-icon>
                批量启动
              </el-dropdown-item>
              <el-dropdown-item command="stop">
                <el-icon><VideoPause /></el-icon>
                批量停止
              </el-dropdown-item>
              <el-dropdown-item command="reboot">
                <el-icon><RefreshRight /></el-icon>
                批量重启
              </el-dropdown-item>
              <el-dropdown-item command="delete" divided>
                <el-icon><Delete /></el-icon>
                批量删除
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <span v-if="selectedRows.length > 0" class="selected-count">
          已选择 {{ selectedRows.length }} 项
        </span>
      </div>

      <div class="toolbar-right">
        <!-- 状态筛选 -->
        <el-select
          v-model="filterStatus"
          placeholder="状态筛选"
          clearable
          style="width: 120px"
          @change="handleFilter"
        >
          <el-option label="运行中" value="running" />
          <el-option label="已停止" value="stopped" />
          <el-option label="错误" value="error" />
        </el-select>

        <!-- 搜索 -->
        <el-input
          v-model="searchQuery"
          placeholder="搜索容器名称或ID"
          clearable
          style="width: 240px"
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <!-- 刷新 -->
        <el-button text @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="filteredCTList"
        style="width: 100%"
        stripe
        border
        @selection-change="handleSelectionChange"
        @row-click="handleRowClick"
      >
        <el-table-column type="selection" width="50" />

        <el-table-column prop="ctid" :label="t('lxc.ctid')" width="90" sortable />

        <el-table-column prop="name" :label="t('common.name')" min-width="180">
          <template #default="{ row }">
            <div class="ct-name">
              <span class="name-text">{{ row.name }}</span>
              <el-tag v-if="row.template" size="small" type="warning">{{ row.template }}</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="status" :label="t('common.status')" width="110">
          <template #default="{ row }">
            <VMStatusBadge :status="row.status" />
          </template>
        </el-table-column>

        <el-table-column prop="node" label="节点" width="120" />

        <el-table-column label="CPU" width="100" sortable prop="cpu">
          <template #default="{ row }">
            <div class="resource-cell">
              <span>{{ row.cpu }} 核</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="内存" width="120" sortable prop="memory">
          <template #default="{ row }">
            <div class="resource-cell">
              <span>{{ row.memory }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="磁盘" width="120">
          <template #default="{ row }">
            <div class="resource-cell">
              <span>{{ row.disk }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="IP 地址" width="150">
          <template #default="{ row }">
            <span class="ip-text">{{ row.ip || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="uptime" label="运行时间" width="120" />

        <el-table-column :label="t('common.actions')" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                v-if="row.status === 'stopped'"
                link
                type="success"
                size="small"
                @click.stop="handleStart(row)"
              >
                启动
              </el-button>
              <el-button
                v-else
                link
                type="danger"
                size="small"
                @click.stop="handleStop(row)"
              >
                停止
              </el-button>
              <el-button
                link
                type="primary"
                size="small"
                @click.stop="handleConsole(row)"
              >
                控制台
              </el-button>
              <el-dropdown trigger="click" @command="handleRowCommand($event, row)">
                <el-button link type="primary" size="small">
                  更多
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="reboot">
                      <el-icon><RefreshRight /></el-icon>
                      重启
                    </el-dropdown-item>
                    <el-dropdown-item command="snapshot">
                      <el-icon><Camera /></el-icon>
                      快照
                    </el-dropdown-item>
                    <el-dropdown-item command="clone">
                      <el-icon><CopyDocument /></el-icon>
                      克隆
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <el-icon><Delete /></el-icon>
                      删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Refresh,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Delete,
  Camera,
  CopyDocument,
} from '@element-plus/icons-vue'
import VMStatusBadge from '@/components/vm/VMStatusBadge.vue'
import BatchOperationBar from '@/components/batch/BatchOperationBar.vue'
import { useBatchStore } from '@/stores/batch'

const { t } = useI18n()
const batchStore = useBatchStore()

const loading = ref(false)
const searchQuery = ref('')
const filterStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(2)
const selectedRows = ref<any[]>([])

/**
 * 容器列表数据（模拟）
 */
const ctList = ref([
  {
    ctid: 200,
    name: 'nginx-proxy',
    status: 'running' as const,
    node: 'pve-node-01',
    template: 'Ubuntu 22.04',
    cpu: 1,
    memory: '512 MB',
    disk: '8 GB',
    ip: '192.168.1.201',
    uptime: '10 天',
  },
  {
    ctid: 201,
    name: 'redis-cache',
    status: 'running' as const,
    node: 'pve-node-01',
    template: 'Debian 12',
    cpu: 1,
    memory: '1 GB',
    disk: '10 GB',
    ip: '192.168.1.202',
    uptime: '8 天',
  },
])

// 筛选后的列表
const filteredCTList = computed(() => {
  let list = ctList.value

  // 状态筛选
  if (filterStatus.value) {
    list = list.filter(ct => ct.status === filterStatus.value)
  }

  // 搜索
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    list = list.filter(ct =>
      ct.name.toLowerCase().includes(query) ||
      ct.ctid.toString().includes(query)
    )
  }

  return list
})

/**
 * 创建容器
 */
function handleCreate() {
  ElMessage.info('容器创建向导开发中...')
}

/**
 * 批量操作命令
 */
function handleBatchCommand(command: string) {
  if (selectedRows.value.length === 0) return

  ElMessage.info(`批量操作: ${command} (${selectedRows.value.length} 台)`)
  // TODO: 实现批量操作逻辑
}

/**
 * 行命令
 */
function handleRowCommand(command: string, row: any) {
  switch (command) {
    case 'start':
      handleStart(row)
      break
    case 'stop':
      handleStop(row)
      break
    case 'reboot':
      handleReboot(row)
      break
    case 'snapshot':
      ElMessage.info('快照功能开发中...')
      break
    case 'clone':
      ElMessage.info('克隆功能开发中...')
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

/**
 * 启动容器
 */
function handleStart(row: any) {
  ElMessageBox.confirm(`确认启动容器 ${row.name}？`, '提示', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'info',
  }).then(() => {
    ElMessage.success('启动命令已发送')
    // TODO: 调用 API 启动容器
  }).catch(() => {})
}

/**
 * 停止容器
 */
function handleStop(row: any) {
  ElMessageBox.confirm(`确认停止容器 ${row.name}？`, '警告', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(() => {
    ElMessage.success('停止命令已发送')
    // TODO: 调用 API 停止容器
  }).catch(() => {})
}

/**
 * 重启容器
 */
function handleReboot(row: any) {
  ElMessageBox.confirm(`确认重启容器 ${row.name}？`, '警告', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(() => {
    ElMessage.success('重启命令已发送')
    // TODO: 调用 API 重启容器
  }).catch(() => {})
}

/**
 * 删除容器
 */
function handleDelete(row: any) {
  ElMessageBox.confirm(
    `确认删除容器 ${row.name}？此操作不可恢复！`,
    '危险操作',
    {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'error',
    }
  ).then(() => {
    ElMessage.success('删除命令已发送')
    // TODO: 调用 API 删除容器
  }).catch(() => {})
}

/**
 * 打开控制台
 */
function handleConsole(row: any) {
  ElMessage.info(`打开 ${row.name} 控制台 (开发中...)`)
  // TODO: 实现容器控制台
}

/**
 * 选择变化 - 同步到批量操作 store
 */
function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
  // 同步到批量操作 store
  batchStore.clearSelection()
  batchStore.selectMultiple(
    rows.map(row => ({
      id: `ct-${row.ctid}`,
      name: row.name,
      type: 'ct' as const,
      vmid: row.ctid,
    }))
  )
}

/**
 * 行点击
 */
function handleRowClick(row: any) {
  // TODO: 跳转到详情页面
  console.log('点击行:', row)
}

/**
 * 搜索
 */
function handleSearch() {
  currentPage.value = 1
}

/**
 * 筛选
 */
function handleFilter() {
  currentPage.value = 1
}

/**
 * 刷新
 */
function handleRefresh() {
  loading.value = true
  setTimeout(() => {
    loading.value = false
    ElMessage.success('数据刷新成功')
  }, 500)
}

/**
 * 分页变化
 */
function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.lxc-list-page {
  padding: $spacing-6;
  min-height: 100%;
  overflow: auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-6;
  gap: $spacing-4;

  @media (max-width: $breakpoint-sm) {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    .page-title {
      font-size: $font-size-3xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      margin-bottom: $spacing-1;
    }

    .page-description {
      color: $color-text-secondary;
      font-size: $font-size-base;
    }
  }
}

// 工具栏
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-4;
  gap: $spacing-4;
  flex-wrap: wrap;

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: $spacing-4;

    .selected-count {
      color: $color-text-secondary;
      font-size: $font-size-sm;
    }
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: $spacing-3;
  }
}

// 表格卡片
.table-card {
  :deep(.el-card__body) {
    padding: 0;
  }
}

// 表格样式
:deep(.el-table) {
  .el-table__header th {
    background: $gray-2;
  }

  .el-table__row {
    cursor: pointer;

    &:hover {
      background: $primary-1;
    }
  }
}

// 容器名称
.ct-name {
  display: flex;
  align-items: center;
  gap: $spacing-2;

  .name-text {
    font-weight: $font-weight-medium;
    color: $color-text-primary;
  }
}

// 资源单元格
.resource-cell {
  font-size: $font-size-sm;
  color: $color-text-regular;
}

// IP 文本
.ip-text {
  font-family: $font-family-code;
  font-size: $font-size-sm;
  color: $color-text-regular;
}

// 操作按钮
.action-buttons {
  display: flex;
  align-items: center;
  gap: $spacing-2;
}

// 分页
.pagination-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: $spacing-4 $spacing-6;
  border-top: 1px solid $color-border-light;
}
</style>
