<template>
  <div class="lxc-list-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">{{ t('lxc.title') }}</h1>
        <p class="page-description">管理和监控所有 LXC 容器实例</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateWizard = true">
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
        <el-dropdown :disabled="selectedRows.length === 0" @command="handleBatchCommand">
          <el-button :disabled="selectedRows.length === 0">
            批量操作
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="start"><el-icon><VideoPlay /></el-icon> 批量启动</el-dropdown-item>
              <el-dropdown-item command="stop"><el-icon><VideoPause /></el-icon> 批量停止</el-dropdown-item>
              <el-dropdown-item command="reboot"><el-icon><RefreshRight /></el-icon> 批量重启</el-dropdown-item>
              <el-dropdown-item command="delete" divided><el-icon><Delete /></el-icon> 批量删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <span v-if="selectedRows.length > 0" class="selected-count">已选择 {{ selectedRows.length }} 项</span>
      </div>

      <div class="toolbar-right">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 120px" @change="handleFilter">
          <el-option label="运行中" value="running" />
          <el-option label="已停止" value="stopped" />
          <el-option label="已暂停" value="paused" />
          <el-option label="错误" value="error" />
        </el-select>

        <el-input v-model="searchQuery" placeholder="搜索容器名称或ID" clearable style="width: 240px" @input="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

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
        <el-table-column prop="vmid" :label="t('lxc.ctid')" width="90" sortable />

        <el-table-column prop="name" :label="t('common.name')" min-width="180">
          <template #default="{ row }">
            <div class="ct-name">
              <span class="name-text">{{ row.name || `CT ${row.vmid}` }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="status" :label="t('common.status')" width="110">
          <template #default="{ row }">
            <VMStatusBadge :status="normalizeStatus(row.status)" />
          </template>
        </el-table-column>

        <el-table-column prop="node" label="节点" width="120" />

        <el-table-column label="CPU" width="100" sortable :sort-method="(a: LXCVM, b: LXCVM) => a.cpus - b.cpus">
          <template #default="{ row }">
            <span class="resource-cell">{{ row.cpus }} 核</span>
          </template>
        </el-table-column>

        <el-table-column label="内存" width="120" sortable :sort-method="(a: LXCVM, b: LXCVM) => a.maxmem - b.maxmem">
          <template #default="{ row }">
            <span class="resource-cell">{{ formatBytes(row.maxmem) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="磁盘" width="120" sortable :sort-method="(a: LXCVM, b: LXCVM) => a.maxdisk - b.maxdisk">
          <template #default="{ row }">
            <span class="resource-cell">{{ formatBytes(row.maxdisk) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="CPU 使用率" width="110" sortable :sort-method="(a: LXCVM, b: LXCVM) => a.cpu - b.cpu">
          <template #default="{ row }">
            <span class="resource-cell">{{ (row.cpu * 100).toFixed(1) }}%</span>
          </template>
        </el-table-column>

        <el-table-column label="运行时间" width="120" sortable :sort-method="(a: LXCVM, b: LXCVM) => a.uptime - b.uptime">
          <template #default="{ row }">
            <span class="resource-cell">{{ row.uptime > 0 ? formatUptime(row.uptime) : '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button v-if="row.status === 'stopped'" link type="success" size="small" @click.stop="handleStart(row)">启动</el-button>
              <el-button v-else link type="danger" size="small" @click.stop="handleStop(row)">停止</el-button>
              <el-button link type="primary" size="small" @click.stop="openDetail(row)">详情</el-button>
              <el-button link type="primary" size="small" @click.stop="handleConsole(row)">控制台</el-button>
              <el-dropdown trigger="click" @command="handleRowCommand($event, row)">
                <el-button link type="primary" size="small">更多<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="reboot"><el-icon><RefreshRight /></el-icon> 重启</el-dropdown-item>
                    <el-dropdown-item command="delete" divided><el-icon><Delete /></el-icon> 删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="ctList.length"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 创建容器向导 -->
    <LXCCreateWizard v-model="showCreateWizard" @created="handleCreated" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Search, Refresh, VideoPlay, VideoPause, RefreshRight,
  Delete, ArrowDown,
} from '@element-plus/icons-vue'
import VMStatusBadge from '@/components/vm/VMStatusBadge.vue'
import BatchOperationBar from '@/components/batch/BatchOperationBar.vue'
import LXCCreateWizard from '@/components/lxc/LXCCreateWizard.vue'
import { useBatchStore } from '@/stores/batch'
import { fetchLXCList, startLXC, stopLXC, rebootLXC } from '@/api/lxc'
import { getClusterResources } from '@/api/cluster'
import { formatBytes, formatUptime } from '@/utils/format'
import type { LXCVM } from '@/api/types'

const router = useRouter()
const { t } = useI18n()
const batchStore = useBatchStore()

const loading = ref(false)
const searchQuery = ref('')
const filterStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const selectedRows = ref<any[]>([])
const ctList = ref<LXCVM[]>([])
const showCreateWizard = ref(false)

/**
 * 加载容器列表（从 API）
 */
async function loadCTList() {
  loading.value = true
  try {
    const rawResources = await getClusterResources()
    const resources = Array.isArray(rawResources) ? rawResources : (Array.isArray(rawResources?.data) ? rawResources.data : [])
    const nodeNames = resources
      .filter((r: any) => r.type === 'node')
      .map((r: any) => r.node || r.name || r.id)
      .filter(Boolean)

    const allVMs: LXCVM[] = []
    for (const node of nodeNames) {
      try {
        const vms = await fetchLXCList(node)
        const list = Array.isArray(vms) ? vms : (Array.isArray((vms as any)?.data) ? (vms as any).data : [])
        for (const vm of list) {
          vm.node = node
        }
        allVMs.push(...list)
      } catch (e) {
        console.warn(`获取节点 ${node} 容器列表失败:`, e)
      }
    }
    ctList.value = allVMs
  } catch (error) {
    console.error('获取容器列表失败:', error)
  } finally {
    loading.value = false
  }
}

/**
 * 规范化状态
 */
function normalizeStatus(status: string): 'running' | 'stopped' | 'error' | 'paused' | 'unknown' {
  const map: Record<string, 'running' | 'stopped' | 'error' | 'paused' | 'unknown'> = {
    running: 'running', stopped: 'stopped', paused: 'paused',
    prelaunch: 'stopped', suspended: 'paused', migrate: 'running', unknown: 'unknown',
  }
  return map[status] || 'unknown'
}

/**
 * 筛选后的列表
 */
const filteredCTList = computed(() => {
  let list = ctList.value
  if (filterStatus.value) {
    list = list.filter(vm => normalizeStatus(vm.status) === filterStatus.value)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    list = list.filter(vm =>
      (vm.name || '').toLowerCase().includes(query) || vm.vmid.toString().includes(query)
    )
  }
  return list
})

function handleCreated(_vmid: number) { loadCTList() }
function handleBatchCommand(command: string) {
  if (selectedRows.value.length === 0) return
  ElMessage.info(`批量操作: ${command} (${selectedRows.value.length} 台)`)
}

function handleRowCommand(command: string, row: LXCVM) {
  switch (command) {
    case 'start': handleStart(row); break
    case 'stop': handleStop(row); break
    case 'reboot': handleReboot(row); break
    case 'delete': handleDelete(row); break
  }
}

async function handleStart(row: LXCVM) {
  ElMessageBox.confirm(`确认启动容器 ${row.name || row.vmid}？`, '提示', {
    confirmButtonText: '确认', cancelButtonText: '取消', type: 'info',
  }).then(async () => {
    try {
      await startLXC(row.node, row.vmid)
      ElMessage.success('启动命令已发送')
      setTimeout(() => loadCTList(), 2000)
    } catch (e) { console.error('启动失败:', e) }
  }).catch(() => {})
}

async function handleStop(row: LXCVM) {
  ElMessageBox.confirm(`确认停止容器 ${row.name || row.vmid}？`, '警告', {
    confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning',
  }).then(async () => {
    try {
      await stopLXC(row.node, row.vmid)
      ElMessage.success('停止命令已发送')
      setTimeout(() => loadCTList(), 2000)
    } catch (e) { console.error('停止失败:', e) }
  }).catch(() => {})
}

async function handleReboot(row: LXCVM) {
  ElMessageBox.confirm(`确认重启容器 ${row.name || row.vmid}？`, '警告', {
    confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning',
  }).then(async () => {
    try {
      await rebootLXC(row.node, row.vmid)
      ElMessage.success('重启命令已发送')
      setTimeout(() => loadCTList(), 2000)
    } catch (e) { console.error('重启失败:', e) }
  }).catch(() => {})
}

function handleDelete(row: LXCVM) {
  ElMessageBox.confirm(`确认删除容器 ${row.name || row.vmid}？此操作不可恢复！`, '危险操作', {
    confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'error'
  }).then(() => { ElMessage.info('删除功能需要后端支持') }).catch(() => {})
}

function openDetail(row: LXCVM) {
  router.push({ name: 'LXCDetail', params: { node: row.node, vmid: row.vmid.toString() } })
}

function handleConsole(row: LXCVM) {
  ElMessage.info(`${row.name || row.vmid} 控制台开发中`)
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
  batchStore.clearSelection()
  batchStore.selectMultiple(rows.map(row => ({
    id: `ct-${row.vmid}`, name: row.name || `CT ${row.vmid}`, type: 'ct' as const, vmid: row.vmid,
  })))
}

function handleRowClick(row: LXCVM) { openDetail(row) }
function handleSearch() { currentPage.value = 1 }
function handleFilter() { currentPage.value = 1 }

async function handleRefresh() {
  await loadCTList()
  ElMessage.success('数据刷新成功')
}

function handleSizeChange(size: number) { pageSize.value = size; currentPage.value = 1 }
function handlePageChange(page: number) { currentPage.value = page }

onMounted(() => { loadCTList() })
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.lxc-list-page { padding: $spacing-6; min-height: 100%; overflow: auto; }

.page-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-6; gap: $spacing-4;
  @media (max-width: $breakpoint-sm) { flex-direction: column; align-items: flex-start; }
  .header-left {
    .page-title { font-size: $font-size-3xl; font-weight: $font-weight-bold; color: $color-text-primary; margin-bottom: $spacing-1; }
    .page-description { color: $color-text-secondary; font-size: $font-size-base; }
  }
}

.toolbar {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: $spacing-4; gap: $spacing-4; flex-wrap: wrap;
  .toolbar-left { display: flex; align-items: center; gap: $spacing-4; .selected-count { color: $color-text-secondary; font-size: $font-size-sm; } }
  .toolbar-right { display: flex; align-items: center; gap: $spacing-3; }
}

.table-card { :deep(.el-card__body) { padding: 0; } }

:deep(.el-table) {
  .el-table__header th { background: $color-bg-base; }
  .el-table__row { cursor: pointer; &:hover { background: $primary-1; } }
}

.ct-name { display: flex; align-items: center; gap: $spacing-2; .name-text { font-weight: $font-weight-medium; color: $color-text-primary; } }
.resource-cell { font-size: $font-size-sm; color: $color-text-regular; }

.action-buttons { display: flex; align-items: center; gap: $spacing-2; }

.pagination-wrapper {
  display: flex; align-items: center; justify-content: flex-end; padding: $spacing-4 $spacing-6;
  border-top: 1px solid $color-border-light;
}
</style>
