<template>
  <aside :class="['app-sidebar', { collapsed }]">
    <div class="sidebar-logo">
      <el-icon :size="24" class="logo-icon"><Monitor /></el-icon>
      <transition name="fade">
        <span v-show="!collapsed" class="logo-text">PVE Cloud</span>
      </transition>
    </div>

    <!-- 展开状态：导航菜单 -->
    <nav v-show="!collapsed" class="sidebar-nav">
      <div
        v-for="item in navItems"
        :key="item.path"
        :class="['nav-item', { active: matchRoute(item.path) }]"
        @click="navigateTo(item.path)"
      >
        <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
        <span class="nav-label">{{ item.label }}</span>
      </div>
    </nav>

    <!-- 收缩状态：图标导航 -->
    <nav v-show="collapsed" class="compact-nav">
      <el-tooltip
        v-for="item in navItems"
        :key="item.path"
        :content="item.label"
        placement="right"
      >
        <div
          :class="['compact-nav-item', { active: matchRoute(item.path) }]"
          @click="navigateTo(item.path)"
        >
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
      </el-tooltip>
    </nav>

    <div class="sidebar-footer">
      <el-button text class="collapse-btn" @click="toggle">
        <el-icon>
          <DArrowLeft v-if="!collapsed" />
          <DArrowRight v-else />
        </el-icon>
      </el-button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Monitor,
  DataBoard,
  Setting,
  DArrowLeft,
  DArrowRight,
  Files,
  Odometer,
  Key,
  Box,
  FolderOpened,
  Connection,
  OfficeBuilding,
  Grid,
} from '@element-plus/icons-vue'

interface NavItem {
  path: string
  label: string
  icon: Component
}

defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  toggle: []
}>()

const router = useRouter()
const { t } = useI18n()

const navItems = computed<NavItem[]>(() => [
  { path: '/', label: t('layout.dashboard'), icon: DataBoard },
  { path: '/qemu', label: '虚拟机管理', icon: Monitor },
  { path: '/lxc', label: '容器管理', icon: Box },
  { path: '/storage', label: '存储管理', icon: FolderOpened },
  { path: '/nodes', label: '节点管理', icon: OfficeBuilding },
  { path: '/cluster', label: '集群概览', icon: Connection },
  { path: '/backup', label: '备份管理', icon: Files },
  { path: '/monitor', label: '监控中心', icon: Odometer },
  { path: '/apps', label: '应用商店', icon: Grid },
  { path: '/access', label: '访问管理', icon: Key },
  { path: '/settings', label: t('layout.settings'), icon: Setting },
])

function matchRoute(path: string): boolean {
  if (path === '/') {
    return router.currentRoute.value.path === '/'
  }
  return router.currentRoute.value.path.startsWith(path)
}

function navigateTo(path: string): void {
  router.push(path)
}

function toggle(): void {
  emit('toggle')
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.app-sidebar {
  width: 220px;
  min-width: 220px;
  height: 100vh;
  background: $color-bg-container;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
  border-right: 1px solid $color-border-light;

  &.collapsed {
    width: 64px;
    min-width: 64px;
  }
}

.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  padding: 0 16px;
  gap: 8px;
  border-bottom: 1px solid $color-border-light;
  color: $color-text-primary;
  font-weight: 600;
  font-size: 15px;
  white-space: nowrap;
  background: $color-bg-elevated;
  font-family: 'Fira Code', 'Consolas', monospace;
  flex-shrink: 0;

  .logo-icon {
    color: $green-500;
    flex-shrink: 0;
  }

  .logo-text {
    letter-spacing: 1px;
    transition: opacity 0.3s ease;
  }
}

// ============================================================
// 展开状态：导航菜单
// ============================================================

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 20px;
  gap: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: $color-text-regular;
  position: relative;
  white-space: nowrap;
  user-select: none;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: transparent;
    transition: background 0.2s;
    border-radius: 0 2px 2px 0;
  }

  &:hover {
    color: $color-text-primary;
    background: $color-bg-hover;
  }

  &.active {
    color: $green-500;
    background: $color-bg-active;
    font-weight: 500;

    &::before {
      background: $green-500;
    }

    .nav-icon {
      color: $green-500;
    }
  }

  .nav-icon {
    font-size: 18px;
    flex-shrink: 0;
    color: inherit;
  }

  .nav-label {
    font-size: 14px;
    line-height: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

// ============================================================
// 收缩状态：图标导航
// ============================================================

.compact-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 8px;
  gap: 4px;
}

.compact-nav-item {
  width: 48px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0;
  color: $color-text-secondary;
  position: relative;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: transparent;
    transition: background 0.2s;
  }

  &:hover {
    color: $color-text-regular;
    background: $color-bg-hover;
  }

  &.active {
    color: $green-500;
    background: $color-bg-active;

    &::before {
      background: $green-500;
    }
  }

  .el-icon {
    font-size: 18px;
  }
}

// ============================================================
// 底部折叠按钮
// ============================================================

.sidebar-footer {
  padding: 8px;
  border-top: 1px solid $color-border-light;
  background: $color-bg-elevated;
  flex-shrink: 0;

  .collapse-btn {
    width: 100%;
    height: 40px;
    color: $color-text-secondary;
    justify-content: center;
    border-radius: 6px;

    &:hover {
      color: $color-text-regular;
      background: $color-bg-hover;
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
