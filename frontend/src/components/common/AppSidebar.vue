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
    <div class="node-selector">
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

    <!-- 导航菜单 -->
    <el-menu
      :default-active="currentRoute"
      :collapse="collapsed"
      :collapse-transition="true"
      router
      class="sidebar-menu"
    >
      <el-menu-item index="/">
        <el-icon><DataBoard /></el-icon>
        <template #title>{{ t('layout.dashboard') }}</template>
      </el-menu-item>

      <el-sub-menu index="compute">
        <template #title>
          <el-icon><Cpu /></el-icon>
          <span>{{ t('layout.compute') }}</span>
        </template>
        <el-menu-item index="/qemu">
          <el-icon><Monitor /></el-icon>
          <template #title>{{ t('layout.qemu') }}</template>
        </el-menu-item>
        <el-menu-item index="/lxc">
          <el-icon><Box /></el-icon>
          <template #title>{{ t('layout.lxc') }}</template>
        </el-menu-item>
      </el-sub-menu>

      <el-menu-item index="/storage">
        <el-icon><FolderOpened /></el-icon>
        <template #title>{{ t('layout.storage') }}</template>
      </el-menu-item>

      <el-menu-item index="/network" :class="{ 'coming-soon': true }">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('layout.network') }}</template>
      </el-menu-item>

      <el-menu-item index="/monitor" :class="{ 'coming-soon': true }">
        <el-icon><TrendCharts /></el-icon>
        <template #title>{{ t('layout.monitor') }}</template>
      </el-menu-item>

      <el-menu-item index="/tasks" :class="{ 'coming-soon': true }">
        <el-icon><List /></el-icon>
        <template #title>{{ t('layout.tasks') }}</template>
      </el-menu-item>

      <el-menu-item index="/settings">
        <el-icon><Setting /></el-icon>
        <template #title>{{ t('layout.settings') }}</template>
      </el-menu-item>
    </el-menu>

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
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import {
  Monitor,
  DataBoard,
  Cpu,
  Box,
  FolderOpened,
  Connection,
  TrendCharts,
  List,
  Setting,
  DArrowLeft,
  DArrowRight,
} from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()

const route = useRoute()
const { t } = useI18n()
const authStore = useAuthStore()

// 当前路由路径（用于菜单高亮）
const currentRoute = computed(() => route.path)

// 节点选择
const savedNodes = computed(() => authStore.savedNodes)
const selectedNode = ref(authStore.currentNode?.host || '')

// 监听节点切换
function handleNodeChange(host: string) {
  const node = savedNodes.value.find(n => n.host === host)
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

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  background: transparent;
  border-right: none;
  padding: $spacing-2 0;

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    margin: $spacing-1 $spacing-2;
    border-radius: $radius-sm;
    height: 44px;

    &:hover {
      background: rgba(255, 255, 255, 0.08) !important;
    }

    &.is-active {
      background: $color-primary !important;
      color: #fff !important;
    }

    .el-icon {
      margin-right: $spacing-3;
    }
  }

  // 未开发功能样式
  :deep(.coming-soon) {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

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

// 过渡动画
.fade-enter-active,
.fade-leave-active {
  transition: opacity $duration-normal $ease-base;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
