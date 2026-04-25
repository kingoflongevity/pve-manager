<template>
  <div class="page-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h2 class="page-title">{{ t('storage.title') }}</h2>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        {{ t('storage.addStorage') }}
      </el-button>
    </div>

    <!-- 概览统计 -->
    <el-row :gutter="16" class="summary-bar">
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-label">{{ t('storage.totalStorage') }}</div>
            <div class="summary-value">{{ storageList.length }}</div>
          </div>
          <div class="summary-icon" style="color: var(--el-color-primary)">
            <el-icon :size="32"><FolderOpened /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-label">{{ t('storage.totalSpace') }}</div>
            <div class="summary-value">{{ formatBytes(totalSpace) }}</div>
          </div>
          <div class="summary-icon" style="color: var(--el-color-info)">
            <el-icon :size="32"><DataLine /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-label">{{ t('storage.usedSpace') }}</div>
            <div class="summary-value">{{ formatBytes(usedSpace) }}</div>
          </div>
          <div class="summary-icon" style="color: var(--el-color-warning)">
            <el-icon :size="32"><TrendCharts /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <div class="summary-label">{{ t('storage.availableSpace') }}</div>
            <div class="summary-value">{{ formatBytes(availableSpace) }}</div>
          </div>
          <div class="summary-icon" style="color: var(--el-color-success)">
            <el-icon :size="32"><CircleCheck /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
          <el-input
            v-model="searchQuery"
            :placeholder="t('common.search')"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select
            v-model="filterType"
            :placeholder="t('storage.filterByType')"
            clearable
            @change="handleFilter"
          >
            <el-option :label="t('storage.allTypes')" value="" />
            <el-option
              v-for="item in storageTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select
            v-model="filterStatus"
            :placeholder="t('storage.filterByStatus')"
            clearable
            @change="handleFilter"
          >
            <el-option :label="t('storage.allStatus')" value="" />
            <el-option :label="t('storage.active')" value="active" />
            <el-option :label="t('storage.inactive')" value="inactive" />
          </el-select>
        </el-col>
        <el-col :span="4" style="text-align: right">
          <el-button @click="handleRefresh" :loading="loading">
            <el-icon><Refresh /></el-icon>
            {{ t('storage.refresh') }}
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 数据表格 -->
    <el-card shadow="never" class="table-card">
      <el-table
        :data="filteredStorageList"
        style="width: 100%"
        stripe
        v-loading="loading"
        :empty-text="t('common.noData')"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="storage" :label="t('storage.storageId')" width="150" sortable="custom">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
              {{ row.storage }}
            </el-link>
          </template>
        </el-table-column>

        <el-table-column prop="type" :label="t('storage.type')" width="120" sortable="custom">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ getStorageTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('storage.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'active' ? 'success' : 'danger'"
              size="small"
              effect="dark"
            >
              {{ row.status === 'active' ? t('storage.active') : t('storage.inactive') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="total" :label="t('storage.total')" width="120" sortable="custom">
          <template #default="{ row }">
            {{ formatBytes(row.total) }}
          </template>
        </el-table-column>

        <el-table-column prop="used" :label="t('storage.used')" width="120" sortable="custom">
          <template #default="{ row }">
            {{ formatBytes(row.used) }}
          </template>
        </el-table-column>

        <el-table-column prop="available" :label="t('storage.available')" width="120" sortable="custom">
          <template #default="{ row }">
            {{ formatBytes(row.available) }}
          </template>
        </el-table-column>

        <el-table-column prop="usage" :label="t('dashboard.diskUsage')" width="160" sortable="custom">
          <template #default="{ row }">
            <div class="usage-cell">
              <el-progress
                :percentage="row.usage"
                :stroke-width="8"
                :color="getProgressColor(row.usage)"
                :show-text="false"
              />
              <span class="usage-text">{{ row.usage }}%</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="content" :label="t('storage.content')" min-width="180">
          <template #default="{ row }">
            <div class="content-tags">
              <el-tag
                v-for="c in row.content"
                :key="c"
                size="small"
                class="content-tag"
              >
                {{ getContentTypeLabel(c) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleViewDetail(row)">
              {{ t('storage.detail') }}
            </el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">
              {{ t('storage.edit') }}
            </el-button>
            <el-popconfirm
              :title="t('storage.confirmDelete', { name: row.storage })"
              confirm-button-text="确认"
              cancel-button-text="取消"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button link type="danger" size="small">
                  {{ t('storage.delete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="totalFiltered"
          layout="total, sizes, prev, pager, next"
          background
        />
      </div>
    </el-card>

    <!-- 创建/编辑存储对话框 -->
    <CreateStorageDialog
      ref="createDialogRef"
      :node="currentNode"
      :is-edit="isEditMode"
      :storage-data="editingStorage"
      @success="handleRefresh"
      @close="editingStorage = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Plus,
  FolderOpened,
  DataLine,
  TrendCharts,
  CircleCheck,
  Search,
  Refresh,
} from '@element-plus/icons-vue'
import type { StorageItem } from '@/types/storage'
import { StorageBackendTypeLabel, StorageContentTypeLabel } from '@/types/storage'
import { getClusterStorages, deleteStorage } from '@/api/storage'
import { formatBytes } from '@/utils/format'
import CreateStorageDialog from '@/components/storage/CreateStorageDialog.vue'

const { t } = useI18n()
const router = useRouter()

// ============================================================
// 状态管理
// ============================================================

/** 存储列表数据 */
const storageList = ref<StorageItem[]>([])
/** 加载状态 */
const loading = ref(false)
/** 搜索关键词 */
const searchQuery = ref('')
/** 筛选：存储类型 */
const filterType = ref('')
/** 筛选：状态 */
const filterStatus = ref('')
/** 当前页码 */
const currentPage = ref(1)
/** 每页条数 */
const pageSize = ref(20)
/** 当前节点（从路由获取） */
const currentNode = ref('')
/** 创建/编辑对话框引用 */
const createDialogRef = ref<InstanceType<typeof CreateStorageDialog>>()
/** 是否为编辑模式 */
const isEditMode = ref(false)
/** 正在编辑的存储配置 */
const editingStorage = ref<Partial<StorageItem> | null>(null)

// ============================================================
// 存储类型选项
// ============================================================

const storageTypeOptions = Object.entries(StorageBackendTypeLabel).map(([value, label]) => ({
  value,
  label,
}))

// ============================================================
// 计算属性
// ============================================================

/**
 * 总空间（所有存储的容量之和）
 */
const totalSpace = computed<number>(() => {
  return storageList.value.reduce((sum, s) => sum + (s.total || 0), 0)
})

/**
 * 已使用空间
 */
const usedSpace = computed<number>(() => {
  return storageList.value.reduce((sum, s) => sum + (s.used || 0), 0)
})

/**
 * 可用空间
 */
const availableSpace = computed<number>(() => {
  return storageList.value.reduce((sum, s) => sum + (s.available || 0), 0)
})

/**
 * 根据搜索和筛选条件过滤后的存储列表
 */
const filteredStorageList = computed<StorageItem[]>(() => {
  let result = [...storageList.value]

  // 按搜索关键词过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(
      (s) =>
        s.storage.toLowerCase().includes(query) ||
        s.type.toLowerCase().includes(query) ||
        s.node?.toLowerCase().includes(query),
    )
  }

  // 按类型过滤
  if (filterType.value) {
    result = result.filter((s) => s.type === filterType.value)
  }

  // 按状态过滤
  if (filterStatus.value) {
    result = result.filter((s) => s.status === filterStatus.value)
  }

  return result
})

/**
 * 过滤后的总数
 */
const totalFiltered = computed<number>(() => filteredStorageList.value.length)

// ============================================================
// 工具函数
// ============================================================

/**
 * 获取存储类型中文标签
 */
function getStorageTypeLabel(type: string): string {
  return StorageBackendTypeLabel[type] || type
}

/**
 * 获取内容类型中文标签
 */
function getContentTypeLabel(type: string): string {
  return StorageContentTypeLabel[type] || type
}

/**
 * 根据使用率获取进度条颜色
 */
function getProgressColor(usage: number): string {
  if (usage < 60) return '#67c23a'
  if (usage < 80) return '#e6a23c'
  return '#f56c6c'
}

// ============================================================
// 数据获取
// ============================================================

/**
 * 获取存储列表
 */
async function fetchStorages(): Promise<void> {
  loading.value = true
  try {
    const response = await getClusterStorages()
    storageList.value = response.data || []
  } catch (error) {
    console.error('获取存储列表失败:', error)
    ElMessage.error('获取存储列表失败')
  } finally {
    loading.value = false
  }
}

// ============================================================
// 事件处理
// ============================================================

/**
 * 刷新存储列表
 */
function handleRefresh(): void {
  fetchStorages()
}

/**
 * 搜索处理
 */
function handleSearch(): void {
  currentPage.value = 1
}

/**
 * 筛选处理
 */
function handleFilter(): void {
  currentPage.value = 1
}

/**
 * 排序处理
 */
function handleSortChange(): void {
  // 可根据需要实现排序逻辑
}

/**
 * 添加存储
 */
function handleAdd(): void {
  isEditMode.value = false
  editingStorage.value = null
  currentNode.value = currentNode.value || 'local'
  createDialogRef.value?.openDialog()
}

/**
 * 编辑存储
 */
function handleEdit(row: StorageItem): void {
  isEditMode.value = true
  editingStorage.value = {
    storage: row.storage,
    type: row.type,
    enable: row.active,
    shared: row.shared,
    content: row.content.join(','),
    path: row.path,
    server: row.server,
    export: row.export,
    pool: row.pool,
    vgname: row.vgname,
  }
  currentNode.value = row.node || 'local'
  createDialogRef.value?.openDialog()
}

/**
 * 删除存储
 */
async function handleDelete(row: StorageItem): Promise<void> {
  try {
    await deleteStorage(row.node || 'local', row.storage)
    ElMessage.success(t('storage.deleteSuccess'))
    handleRefresh()
  } catch (error) {
    console.error('删除存储失败:', error)
    ElMessage.error(t('storage.deleteFailed'))
  }
}

/**
 * 查看存储详情
 */
function handleViewDetail(row: StorageItem): void {
  router.push({
    name: 'StorageDetail',
    params: {
      node: row.node || 'local',
      storage: row.storage,
    },
  })
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  fetchStorages()
})
</script>

<style scoped lang="scss">
@use '@/assets/styles/variables' as *;

.page-container {
  padding: $spacing-6;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $spacing-6;

  .page-title {
    font-size: $font-size-3xl;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin: 0;
  }
}

// 概览统计卡片
.summary-bar {
  margin-bottom: $spacing-6;
}

.summary-card {
  border-radius: $radius-md;
  transition: transform $duration-normal $ease-base;

  &:hover {
    transform: translateY(-2px);
  }

  :deep(.el-card__body) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $spacing-5;
  }
}

.summary-content {
  flex: 1;
}

.summary-label {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  margin-bottom: $spacing-1;
}

.summary-value {
  font-size: $font-size-2xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.summary-icon {
  opacity: 0.6;
}

// 筛选栏
.filter-card {
  margin-bottom: $spacing-4;

  :deep(.el-card__body) {
    padding: $spacing-4;
  }
}

// 表格卡片
.table-card {
  :deep(.el-card__body) {
    padding: $spacing-4;
  }
}

// 使用率单元格
.usage-cell {
  display: flex;
  align-items: center;
  gap: $spacing-2;

  .el-progress {
    flex: 1;
  }

  .usage-text {
    font-size: $font-size-xs;
    color: $color-text-regular;
    min-width: 40px;
    text-align: right;
  }
}

// 内容类型标签
.content-tags {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-1;

  .content-tag {
    font-size: $font-size-xs;
  }
}

// 分页
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: $spacing-4;
}

// 响应式
@media (max-width: $breakpoint-md) {
  .page-container {
    padding: $spacing-4;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-3;
  }

  .summary-card {
    margin-bottom: $spacing-3;
  }
}
</style>
