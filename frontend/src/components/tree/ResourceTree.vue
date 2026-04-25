<template>
  <div class="resource-tree">
    <!-- 搜索框 -->
    <div class="tree-search">
      <el-input
        v-model="searchQuery"
        :placeholder="t('tree.searchPlaceholder')"
        :prefix-icon="Search"
        clearable
        size="small"
        class="search-input"
      />
    </div>

    <!-- 树工具栏 -->
    <div class="tree-toolbar">
      <el-button
        :icon="Expand"
        text
        size="small"
        :title="t('tree.expandAll')"
        @click="resourceStore.expandAll()"
      />
      <el-button
        :icon="Fold"
        text
        size="small"
        :title="t('tree.collapseAll')"
        @click="resourceStore.collapseAll()"
      />
      <el-button
        :icon="Refresh"
        text
        size="small"
        :title="t('tree.refresh')"
        :loading="resourceStore.loading"
        @click="resourceStore.refresh()"
      />
    </div>

    <!-- 资源树主体 -->
    <div class="tree-content">
      <el-scrollbar v-loading="isInitialLoading">
        <el-tree
          ref="treeRef"
          :data="resourceTree"
          :props="treeProps"
          :expand-on-click-node="false"
          node-key="id"
          :default-expanded-keys="resourceStore.expandedKeys"
          :current-node-key="resourceStore.selectedNodeId"
          highlight-current
          class="resource-tree-el"
          @node-click="handleNodeClick"
        >
          <template #default="{ node, data }">
            <div class="tree-node-content" :data-type="data.type">
              <!-- 状态指示器 -->
              <span
                class="status-indicator"
                :class="`status-${data.status}`"
              />

              <!-- 类型图标 -->
              <el-icon class="type-icon" :size="16">
                <component :is="getTypeIcon(data.type)" />
              </el-icon>

              <!-- 节点名称 -->
              <span class="node-label" :title="node.label">
                {{ node.label }}
              </span>

              <!-- 附加信息标签 -->
              <span v-if="data.type === 'vm' || data.type === 'ct'" class="node-id-tag">
                {{ data.type === 'vm' ? 'VM' : 'CT' }}
              </span>
            </div>
          </template>
        </el-tree>

        <!-- 空状态 -->
        <el-empty
          v-if="showEmptyState"
          :description="t('tree.noResources')"
          :image-size="80"
        />
      </el-scrollbar>
    </div>

    <!-- 最后刷新时间 -->
    <div v-if="resourceStore.lastRefreshedAt" class="tree-footer">
      <span class="refresh-time">
        {{ t('tree.lastRefreshed') }}: {{ formatTime(resourceStore.lastRefreshedAt) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useResourceStore } from '@/stores/resources'
import { Search, Expand, Fold, Refresh } from '@element-plus/icons-vue'
import type { ElTree } from 'element-plus'
import type { TreeResourceData } from '@/types/resources'

/**
 * ResourceTree - PVE 资源树导航组件
 *
 * 功能:
 * - 三级树结构: 数据中心 -> 节点 -> VM/CT/Storage/Network
 * - 展开/收起动画
 * - 点击导航到详情页
 * - 状态指示器 (运行/停止/错误)
 * - 搜索过滤
 * - 深色侧边栏主题
 */
const router = useRouter()
const { t } = useI18n()
const resourceStore = useResourceStore()

// Tree 组件引用
const treeRef = ref<InstanceType<typeof ElTree>>()

// Tree 数据源配置
const treeProps = {
  children: 'children',
  label: 'name',
  disabled: 'disabled',
}

// 搜索查询绑定
const searchQuery = computed({
  get: () => resourceStore.searchQuery,
  set: (val: string) => resourceStore.setSearchQuery(val),
})

// 计算属性
const resourceTree = computed(() => resourceStore.resourceTree)

/** 是否显示空状态 */
const showEmptyState = computed(() => {
  return !resourceStore.loading && resourceTree.value.length === 0
})

/** 是否处于初始加载状态 */
const isInitialLoading = computed(() => {
  return resourceStore.loading && resourceStore.lastRefreshedAt === null
})

/**
 * 根据资源类型获取对应图标组件名
 */
function getTypeIcon(type: string): string {
  const iconMap: Record<string, string> = {
    datacenter: 'Coin',
    node: 'Server',
    vm: 'Monitor',
    ct: 'Box',
    storage: 'FolderOpened',
    network: 'Connection',
  }
  return iconMap[type] || 'Document'
}

/**
 * 格式化时间显示
 */
function formatTime(date: Date): string {
  const now = new Date()
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diff < 5) return t('tree.justNow')
  if (diff < 60) return `${diff}${t('tree.secondsAgo')}`
  if (diff < 3600) return `${Math.floor(diff / 60)}${t('tree.minutesAgo')}`

  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

/**
 * 处理节点点击事件 - 导航到详情页
 */
function handleNodeClick(data: TreeResourceData): void {
  resourceStore.selectNode(data.id)
  navigateToResource(data)
}

/**
 * 根据资源类型导航到对应路由
 */
function navigateToResource(data: TreeResourceData): void {
  const routeMap: Record<string, string> = {
    datacenter: '/',
    node: `/nodes/${data.name}`,
    vm: `/qemu/${data.id}`,
    ct: `/lxc/${data.id}`,
    storage: `/storage/${data.name}`,
    network: `/network/${data.name}`,
  }

  const route = routeMap[data.type]
  if (route) {
    router.push(route).catch(() => {
      // 路由不存在时忽略错误（如未实现的功能）
    })
  }
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  // 初始化加载资源数据
  resourceStore.fetchResources()
  // 启动 30 秒轮询
  resourceStore.startPolling()
})

onUnmounted(() => {
  // 组件销毁时停止轮询，防止内存泄漏
  resourceStore.stopPolling()
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.resource-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: transparent;
}

// ============================================================
// 搜索框
// ============================================================

.tree-search {
  padding: $spacing-3;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);

  :deep(.search-input) {
    .el-input__wrapper {
      background: rgba(255, 255, 255, 0.06);
      box-shadow: none;
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: $radius-sm;

      &:hover {
        box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.2) inset;
      }

      &.is-focus {
        box-shadow: 0 0 0 1px $color-primary inset;
      }
    }

    .el-input__inner {
      color: rgba(255, 255, 255, 0.85);
      font-size: $font-size-sm;

      &::placeholder {
        color: rgba(255, 255, 255, 0.35);
      }
    }

    .el-input__prefix {
      color: rgba(255, 255, 255, 0.45);
    }

    .el-input__suffix {
      .el-icon {
        color: rgba(255, 255, 255, 0.45);
      }
    }
  }
}

// ============================================================
// 工具栏
// ============================================================

.tree-toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: $spacing-2 $spacing-3;
  gap: $spacing-1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);

  :deep(.el-button) {
    color: rgba(255, 255, 255, 0.55);
    padding: $spacing-1;
    min-height: 24px;

    &:hover {
      color: rgba(255, 255, 255, 0.85);
      background: rgba(255, 255, 255, 0.08);
    }

    &.is-loading {
      color: $color-primary-light;
    }
  }
}

// ============================================================
// 树内容区
// ============================================================

.tree-content {
  flex: 1;
  overflow: hidden;
  padding: $spacing-2 0;

  :deep(.el-scrollbar__view) {
    height: 100%;
  }
}

// ============================================================
// Element Plus Tree 样式覆盖
// ============================================================

.resource-tree-el {
  background: transparent;
  padding: 0 $spacing-2;

  // 树节点行
  :deep(.el-tree-node__content) {
    height: 36px;
    border-radius: $radius-sm;
    margin: 1px 0;
    padding: 0 $spacing-2;

    &:hover {
      background: rgba(255, 255, 255, 0.06);
    }
  }

  // 选中状态
  :deep(.el-tree-node.is-current > .el-tree-node__content) {
    background: rgba($color-primary, 0.25);
    color: #fff;
  }

  // 展开箭头
  :deep(.el-tree-node__expand-icon) {
    color: rgba(255, 255, 255, 0.45);
    font-size: 12px;
    transition: transform $duration-normal $ease-base;

    &.is-leaf {
      color: transparent;
    }
  }

  // 高亮指示条
  :deep(.el-tree-node__content.is-current) {
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 3px;
      height: 20px;
      background: $color-primary;
      border-radius: 0 $radius-xs $radius-xs 0;
    }
  }
}

// ============================================================
// 树节点内容
// ============================================================

.tree-node-content {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  flex: 1;
  min-width: 0;
  padding-right: $spacing-2;
}

// 状态指示器
.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: $radius-full;
  flex-shrink: 0;
  transition: background-color $duration-normal $ease-base;

  &.status-running {
    background: $color-success;
    box-shadow: 0 0 4px rgba($color-success, 0.5);
  }

  &.status-stopped {
    background: $gray-6;
  }

  &.status-error {
    background: $color-danger;
    box-shadow: 0 0 4px rgba($color-danger, 0.5);
    animation: pulse-error 2s ease-in-out infinite;
  }

  &.status-unknown {
    background: $warning-6;
  }
}

// 类型图标
.type-icon {
  color: rgba(255, 255, 255, 0.65);
  flex-shrink: 0;
}

// 节点名称
.node-label {
  flex: 1;
  min-width: 0;
  color: rgba(255, 255, 255, 0.85);
  font-size: $font-size-sm;
  line-height: $line-height-sm;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

// 资源 ID 标签
.node-id-tag {
  font-size: $font-size-xs;
  color: rgba(255, 255, 255, 0.45);
  background: rgba(255, 255, 255, 0.08);
  padding: 1px $spacing-2;
  border-radius: $radius-xs;
  flex-shrink: 0;
  font-weight: $font-weight-medium;
}

// ============================================================
// 底部信息
// ============================================================

.tree-footer {
  padding: $spacing-3;
  border-top: 1px solid rgba(255, 255, 255, 0.06);

  .refresh-time {
    color: rgba(255, 255, 255, 0.35);
    font-size: $font-size-xs;
    line-height: $line-height-xs;
  }
}

// ============================================================
// 动画
// ============================================================

@keyframes pulse-error {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 4px rgba($color-danger, 0.5);
  }
  50% {
    opacity: 0.6;
    box-shadow: 0 0 8px rgba($color-danger, 0.8);
  }
}
</style>
