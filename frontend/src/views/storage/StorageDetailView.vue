<template>
  <div class="page-container" v-loading="loading">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h2 class="page-title">
          {{ t('storage.storageDetail') }} - {{ storageName }}
        </h2>
      </div>
      <div class="header-right">
        <el-button @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          {{ t('storage.refresh') }}
        </el-button>
        <el-upload
          :action="uploadUrl"
          :headers="uploadHeaders"
          :data="uploadData"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :before-upload="beforeUpload"
          :show-file-list="false"
          :accept="acceptedFileTypes"
        >
          <el-button type="primary">
            <el-icon><Upload /></el-icon>
            {{ t('storage.uploadFile') }}
          </el-button>
        </el-upload>
      </div>
    </div>

    <!-- 存储信息卡片 -->
    <el-row :gutter="16" class="info-row">
      <!-- 基本信息 -->
      <el-col :span="16">
        <el-card shadow="never" class="info-card">
          <template #header>
            <div class="card-header">
              <span>{{ t('storage.storageInfo') }}</span>
              <el-tag
                :type="storageDetail?.status === 'active' ? 'success' : 'danger'"
                effect="dark"
                size="small"
              >
                {{ storageDetail?.status === 'active' ? t('storage.active') : t('storage.inactive') }}
              </el-tag>
            </div>
          </template>

          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('storage.storageId')">
              {{ storageDetail?.storage }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('storage.type')">
              <el-tag size="small">{{ getStorageTypeLabel(storageDetail?.type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('storage.node')">
              {{ storageDetail?.node || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('storage.shared')">
              {{ storageDetail?.shared ? '是' : '否' }}
            </el-descriptions-item>
            <el-descriptions-item v-if="storageDetail?.path" :label="t('storage.storagePath')">
              <el-text type="info" size="small">{{ storageDetail.path }}</el-text>
            </el-descriptions-item>
            <el-descriptions-item v-if="storageDetail?.server" :label="t('storage.server')">
              {{ storageDetail.server }}
            </el-descriptions-item>
            <el-descriptions-item v-if="storageDetail?.pool" :label="t('storage.pool')">
              {{ storageDetail.pool }}
            </el-descriptions-item>
            <el-descriptions-item v-if="storageDetail?.vgname" :label="t('storage.vgname')">
              {{ storageDetail.vgname }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- 内容类型 -->
          <div class="content-types">
            <span class="label">{{ t('storage.content') }}：</span>
            <el-tag
              v-for="c in storageDetail?.content"
              :key="c"
              size="small"
              class="content-tag"
            >
              {{ getContentTypeLabel(c) }}
            </el-tag>
          </div>
        </el-card>
      </el-col>

      <!-- 使用率饼图 -->
      <el-col :span="8">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <span>{{ t('storage.usageChart') }}</span>
          </template>
          <div class="chart-container">
            <div class="pie-chart" :style="pieChartStyle">
              <div class="pie-center">
                <span class="pie-percent">{{ storageDetail?.usage || 0 }}%</span>
              </div>
            </div>
          </div>
          <div class="chart-legend">
            <div class="legend-item">
              <span class="legend-dot used"></span>
              <span>{{ t('storage.usedSpace') }}: {{ formatBytes(storageDetail?.used || 0) }}</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot available"></span>
              <span>{{ t('storage.availableSpace') }}: {{ formatBytes(storageDetail?.available || 0) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 内容浏览器 -->
    <el-card shadow="never" class="content-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('storage.contentBrowser') }}</span>
          <el-select
            v-model="selectedContentType"
            :placeholder="t('storage.selectContentType')"
            style="width: 160px"
            @change="handleContentTypeChange"
          >
            <el-option :label="t('storage.allContent')" value="" />
            <el-option
              v-for="c in storageDetail?.content"
              :key="c"
              :label="getContentTypeLabel(c)"
              :value="c"
            />
          </el-select>
        </div>
      </template>

      <!-- 内容表格 -->
      <el-table
        :data="contentList"
        style="width: 100%"
        stripe
        :empty-text="t('storage.noContent')"
      >
        <el-table-column prop="filename" :label="t('storage.fileName')" min-width="250">
          <template #default="{ row }">
            <div class="file-name">
              <el-icon class="file-icon">
                <Document v-if="isFileType(row.filename)" />
                <Picture v-else-if="isImageFile(row.filename)" />
                <Folder v-else />
              </el-icon>
              <span>{{ row.filename }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="content" :label="t('storage.type')" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ getContentTypeLabel(row.content) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="size" :label="t('storage.size')" width="120" sortable>
          <template #default="{ row }">
            {{ formatBytes(row.size || 0) }}
          </template>
        </el-table-column>

        <el-table-column prop="format" :label="t('storage.format')" width="100">
          <template #default="{ row }">
            {{ row.format || '--' }}
          </template>
        </el-table-column>

        <el-table-column prop="ctime" :label="t('storage.createdTime')" width="150" sortable>
          <template #default="{ row }">
            {{ row.ctime ? formatDateTime(new Date(row.ctime * 1000).toISOString()) : '--' }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleDownload(row)">
              {{ t('storage.download') }}
            </el-button>
            <el-popconfirm
              :title="`确认删除 ${row.filename}？`"
              confirm-button-text="确认"
              cancel-button-text="取消"
              @confirm="handleDeleteContent(row)"
            >
              <template #reference>
                <el-button link type="danger" size="small">
                  {{ t('common.delete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft,
  Refresh,
  Upload,
  Document,
  Picture,
  Folder,
} from '@element-plus/icons-vue'
import type { StorageDetail, StorageContentItem } from '@/types/storage'
import { StorageBackendTypeLabel, StorageContentTypeLabel } from '@/types/storage'
import {
  getStorageDetail,
  getStorageContents,
  downloadStorageFile,
} from '@/api/storage'
import { formatBytes, formatDateTime } from '@/utils/format'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

// ============================================================
// 状态管理
// ============================================================

/** 加载状态 */
const loading = ref(false)
/** 存储详情 */
const storageDetail = ref<StorageDetail | null>(null)
/** 存储内容列表 */
const contentList = ref<StorageContentItem[]>([])
/** 选中的内容类型 */
const selectedContentType = ref('')

// ============================================================
// 计算属性
// ============================================================

/** 存储名称 */
const storageName = computed(() => {
  return (route.params.storage as string) || ''
})

/** 节点名称 */
const nodeName = computed(() => {
  return (route.params.node as string) || ''
})

/** 饼图样式（使用 conic-gradient 实现） */
const pieChartStyle = computed(() => {
  const usage = storageDetail.value?.usage || 0
  const color = usage > 80 ? '#f5222d' : '#faad14'
  return {
    background: `conic-gradient(${color} 0% ${usage}%, #73d13d ${usage}% 100%)`,
  }
})

/** 上传 URL */
const uploadUrl = computed(() => {
  return `/api/nodes/${nodeName.value}/storage/${storageName.value}/upload`
})

/** 上传请求头 */
const uploadHeaders = computed(() => {
  const token = localStorage.getItem('pve_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
})

/** 上传附加数据 */
const uploadData = computed(() => {
  return {
    content: selectedContentType.value || 'iso',
  }
})

/** 接受的文件类型 */
const acceptedFileTypes = computed(() => {
  if (selectedContentType.value === 'iso') return '.iso'
  if (selectedContentType.value === 'backup') return '.vma,.vma.zst,.tar,.tar.lzo,.tar.gz'
  if (selectedContentType.value === 'vztmpl') return '.tar.gz'
  return '*'
})

// ============================================================
// 工具函数
// ============================================================

/** 获取存储类型中文标签 */
function getStorageTypeLabel(type?: string): string {
  if (!type) return '--'
  return StorageBackendTypeLabel[type] || type
}

/** 获取内容类型中文标签 */
function getContentTypeLabel(type: string): string {
  return StorageContentTypeLabel[type] || type
}

/** 判断是否为普通文件 */
function isFileType(filename: string): boolean {
  const ext = filename.split('.').pop()?.toLowerCase()
  return ['iso', 'qcow2', 'raw', 'vmdk', 'vma', 'tar', 'gz', 'zst'].includes(ext || '')
}

/** 判断是否为图片文件 */
function isImageFile(filename: string): boolean {
  const ext = filename.split('.').pop()?.toLowerCase()
  return ['png', 'jpg', 'jpeg', 'gif', 'svg'].includes(ext || '')
}

// ============================================================
// 数据获取
// ============================================================

/**
 * 获取存储详情
 */
async function fetchStorageDetail(): Promise<void> {
  loading.value = true
  try {
    const response = await getStorageDetail(nodeName.value, storageName.value)
    storageDetail.value = response.data
  } catch (error) {
    console.error('获取存储详情失败:', error)
    ElMessage.error('获取存储详情失败')
  } finally {
    loading.value = false
  }
}

/**
 * 获取存储内容列表
 */
async function fetchContents(): Promise<void> {
  try {
    const response = await getStorageContents(
      nodeName.value,
      storageName.value,
      selectedContentType.value || undefined,
    )
    contentList.value = response.data || []
  } catch (error) {
    console.error('获取存储内容失败:', error)
    ElMessage.error('获取存储内容失败')
  }
}

// ============================================================
// 事件处理
// ============================================================

/** 返回存储列表 */
function handleBack(): void {
  router.push({ name: 'StorageList' })
}

/** 刷新 */
function handleRefresh(): void {
  fetchStorageDetail()
  fetchContents()
}

/** 内容类型变更 */
function handleContentTypeChange(): void {
  fetchContents()
}

/** 上传前校验 */
function beforeUpload(file: File): boolean {
  const maxSize = 50 * 1024 * 1024 * 1024 // 50GB
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过 50GB')
    return false
  }
  return true
}

/** 上传成功 */
function handleUploadSuccess(): void {
  ElMessage.success('上传成功')
  fetchContents()
}

/** 上传失败 */
function handleUploadError(): void {
  ElMessage.error('上传失败')
}

/** 下载文件 */
function handleDownload(row: StorageContentItem): void {
  const url = downloadStorageFile(nodeName.value, storageName.value, row.volid)
  window.open(url, '_blank')
}

/** 删除内容 */
async function handleDeleteContent(row: StorageContentItem): Promise<void> {
  try {
    // 注意：删除内容可能需要通过任务 API 执行
    // 这里使用通用 deleteStorage 作为示例
    ElMessage.info('删除功能需要后端支持卷删除接口')
  } catch (error) {
    console.error('删除内容失败:', error)
    ElMessage.error('删除内容失败')
  }
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  fetchStorageDetail()
  fetchContents()
})

onBeforeUnmount(() => {
  // 清理逻辑
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

  .header-left {
    display: flex;
    align-items: center;
    gap: $spacing-3;
  }

  .page-title {
    font-size: $font-size-2xl;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin: 0;
  }

  .header-right {
    display: flex;
    gap: $spacing-3;
  }
}

// 信息行
.info-row {
  margin-bottom: $spacing-6;
}

.info-card {
  border-radius: $radius-md;
}

.chart-card {
  border-radius: $radius-md;

  :deep(.el-card__body) {
    padding: $spacing-4;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

// 内容类型标签行
.content-types {
  margin-top: $spacing-4;
  padding-top: $spacing-4;
  border-top: 1px solid $color-border-light;

  .label {
    font-weight: $font-weight-medium;
    color: $color-text-regular;
  }

  .content-tag {
    margin-left: $spacing-2;
  }
}

// 图表容器
.chart-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: $spacing-4 0;
}

// 饼图（使用 CSS conic-gradient 实现）
.pie-chart {
  width: 180px;
  height: 180px;
  border-radius: 50%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;

  .pie-center {
    width: 120px;
    height: 120px;
    background: $color-bg-container;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    .pie-percent {
      font-size: $font-size-3xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
    }
  }
}

.chart-legend {
  margin-top: $spacing-4;
  display: flex;
  flex-direction: column;
  gap: $spacing-2;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  font-size: $font-size-sm;
  color: $color-text-regular;

  .legend-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;

    &.used {
      background: #faad14;
    }

    &.available {
      background: #73d13d;
    }
  }
}

// 内容浏览器
.content-card {
  border-radius: $radius-md;
}

// 文件名单元格
.file-name {
  display: flex;
  align-items: center;
  gap: $spacing-2;

  .file-icon {
    color: $color-text-secondary;
  }
}

// 响应式
@media (max-width: $breakpoint-lg) {
  .info-row {
    .el-col {
      margin-bottom: $spacing-4;
    }
  }
}

@media (max-width: $breakpoint-md) {
  .page-container {
    padding: $spacing-4;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-3;

    .header-right {
      width: 100%;
      justify-content: flex-start;
    }
  }
}
</style>
