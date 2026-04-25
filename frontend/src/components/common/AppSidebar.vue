<template>
  <aside :class="['app-sidebar', { collapsed }]">
    <!-- Logo区域 -->
    <div class="sidebar-logo">
      <el-icon :size="28" class="logo-icon"><Monitor /></el-icon>
      <transition name="fade">
        <span v-show="!collapsed" class="logo-text">PVE 管理平台</span>
      </transition>
    </div>

    <!-- 节点选择器 -->
    <transition name="fade">
      <div v-show="!collapsed" class="node-selector">
        <el-select
          v-model="selectedNode"
          :placeholder="t('layout.selectNode')"
          size="small"
          class="node-select"
          @change="handleNodeChange"
        >
          <el-option
            v-for="node in savedNodes"
            :key="node.host"
            :label="node.name"
            :value="node.host"
          />
        </el-select>
      </div>
    </transition>

    <!-- 资源树 -->
    <transition name="fade">
      <div v-show="!collapsed" class="resource-tree-wrapper">
        <ResourceTree />
      </div>
    </transition>

    <!-- 折叠状态下的快捷导航 -->
    <div v-show="collapsed" class="compact-nav">
      <el-tooltip :content="t('layout.dashboard')" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/' }]"
          @click="router.push('/')"
        >
          <el-icon><DataBoard /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="'备份管理'" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/backup' }]"
          @click="router.push('/backup')"
        >
          <el-icon><Files /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="'监控中心'" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/monitor' }]"
          @click="router.push('/monitor')"
        >
          <el-icon><Odometer /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="'访问管理'" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/access' }]"
          @click="router.push('/access')"
        >
          <el-icon><Key /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="'节点管理'" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/nodes' }]"
          @click="router.push('/nodes')"
        >
          <el-icon><Monitor /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip :content="t('layout.settings')" placement="right">
        <el-button
          text
          :class="['compact-nav-item', { active: currentRoute === '/settings' }]"
          @click="router.push('/settings')"
        >
          <el-icon><Setting /></el-icon>
        </el-button>
      </el-tooltip>
    </div>

    <!-- 底部操作区 -->
    <div class="sidebar-footer">
      <el-button
        text
        class="collapse-btn"
        @click="$emit('toggle')"
      >
        <el-icon>
          <DArrowLeft v-if="!collapsed" />
          <DArrowRight v-else />
        </el-icon>
        <span v-show="!collapsed">{{ collapsed ? '展开' : '收起' }}</span>
      </el-button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import ResourceTree from '@/components/tree/ResourceTree.vue'
import {
  Monitor,
  DataBoard,
  Setting,
  DArrowLeft,
  DArrowRight,
  Connection,
  Files,
  Odometer,
  Key,
} from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()

// 当前路由路径（用于菜单高亮）
const currentRoute = computed(() => router.currentRoute.value.path)

// 节点选择
const savedNodes = computed(() => authStore.savedNodes)
const selectedNode = ref(authStore.currentNode?.host || '')

/** 处理节点切换 */
function handleNodeChange(host: string) {
  const node = savedNodes.value.find((n) => n.host === host)
  if (node) {
    authStore.currentNode = node
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.app-sidebar {
  width: $sidebar-width;
  min-width: $sidebar-width;
  background: $sidebar-bg;
  display: flex;
  flex-direction: column;
  transition: all $duration-slow $ease-in-out;
  overflow: hidden;

  &.collapsed {
    width: $sidebar-collapsed-width;
    min-width: $sidebar-collapsed-width;

    .node-select {
      width: 48px;
    }
  }
}

// ============================================================
// Logo 区域
// ============================================================

.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: $header-height;
  padding: 0 $spacing-4;
  gap: $spacing-3;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  color: #fff;
  font-weight: $font-weight-semibold;
  font-size: $font-size-lg;
  white-space: nowrap;

  .logo-icon {
    color: $color-primary;
    flex-shrink: 0;
  }

  .logo-text {
    transition: opacity $duration-normal $ease-base;
  }
}

// ============================================================
// 节点选择器
// ============================================================

.node-selector {
  padding: $spacing-4 $spacing-3;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);

  :deep(.node-select) {
    .el-input__wrapper {
      background: rgba(255, 255, 255, 0.06);
      box-shadow: none;
      border: 1px solid rgba(255, 255, 255, 0.1);

      &:hover {
        box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.2) inset;
      }

      &.is-focus {
        box-shadow: 0 0 0 1px $color-primary inset;
      }
    }

    .el-input__inner {
      color: rgba(255, 255, 255, 0.85);
    }

    .el-input__placeholder {
      color: rgba(255, 255, 255, 0.45);
    }
  }
}

// ============================================================
// 资源树容器
// ============================================================

.resource-tree-wrapper {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

// ============================================================
// 折叠状态下的快捷导航
// ============================================================

.compact-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: $spacing-4;
  gap: $spacing-2;
}

.compact-nav-item {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: $radius-sm;
  color: rgba(255, 255, 255, 0.55);

  &:hover {
    color: rgba(255, 255, 255, 0.85);
    background: rgba(255, 255, 255, 0.08);
  }

  &.active {
    color: #fff;
    background: $color-primary;
  }
}

// ============================================================
// 底部操作区
// ============================================================

.sidebar-footer {
  padding: $spacing-3;
  border-top: 1px solid rgba(255, 255, 255, 0.08);

  .collapse-btn {
    width: 100%;
    color: rgba(255, 255, 255, 0.65);
    justify-content: center;
    gap: $spacing-2;

    &:hover {
      color: rgba(255, 255, 255, 0.85);
      background: rgba(255, 255, 255, 0.08);
    }
  }
}

// ============================================================
// 过渡动画
// ============================================================

.fade-enter-active,
.fade-leave-active {
  transition: opacity $duration-normal $ease-base;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
