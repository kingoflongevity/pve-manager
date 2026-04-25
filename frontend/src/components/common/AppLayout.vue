<template>
  <div class="app-layout">
    <el-aside :width="collapsed ? '64px' : '260px'" class="app-sidebar" :class="{ collapsed }">
      <div class="sidebar-header">
        <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg" class="logo-icon">
          <circle cx="32" cy="32" r="28" stroke="url(#sidebar-logo-gradient)" stroke-width="4" />
          <path d="M24 20L40 32L24 44V20Z" fill="url(#sidebar-logo-gradient)" />
          <defs>
            <linearGradient id="sidebar-logo-gradient" x1="0" y1="0" x2="64" y2="64">
              <stop offset="0%" stop-color="#667eea" />
              <stop offset="100%" stop-color="#764ba2" />
            </linearGradient>
          </defs>
        </svg>
        <transition name="fade">
          <span v-if="!collapsed" class="logo-text">PVE 管理中心</span>
        </transition>
      </div>

      <transition name="fade">
        <div v-if="!collapsed" class="sidebar-menu">
          <el-menu
            :default-active="$route.path"
            :collapse="collapsed"
            router
            class="sidebar-el-menu"
          >
            <el-menu-item index="/dashboard">
              <el-icon><Monitor /></el-icon>
              <template #title>仪表盘</template>
            </el-menu-item>
            <el-menu-item index="/nodes">
              <el-icon><Connection /></el-icon>
              <template #title>节点管理</template>
            </el-menu-item>
            <el-menu-item index="/vms">
              <el-icon><Cpu /></el-icon>
              <template #title>虚拟机</template>
            </el-menu-item>
            <el-menu-item index="/storage">
              <el-icon><Folder /></el-icon>
              <template #title>存储管理</template>
            </el-menu-item>
            <el-menu-item index="/network">
              <el-icon><Switch /></el-icon>
              <template #title>网络管理</template>
            </el-menu-item>
          </el-menu>
        </div>
      </transition>

      <div class="sidebar-footer">
        <el-button text class="collapse-btn" @click="$emit('toggle')">
          <el-icon><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
        </el-button>
      </div>
    </el-aside>

    <div class="main-container">
      <AppHeader :collapsed="collapsed" @toggle="$emit('toggle')" />
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import { Monitor, Connection, Cpu, Folder, Switch, Fold, Expand } from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()
</script>

<style scoped lang="scss">
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: linear-gradient(135deg, #0f0c29 0%, #1a1a2e 50%, #16213e 100%);
}

.app-sidebar {
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(24px);
  border-right: 1px solid rgba(148, 163, 184, 0.1);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;

  &.collapsed {
    .logo-icon {
      width: 36px;
      height: 36px;
    }
  }
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(0, 0, 0, 0.2);

  .logo-icon {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    filter: drop-shadow(0 2px 8px rgba(102, 126, 234, 0.4));
  }

  .logo-text {
    font-size: 18px;
    font-weight: 700;
    color: #f8fafc;
    letter-spacing: 0.5px;
    white-space: nowrap;
  }
}

.sidebar-menu {
  flex: 1;
  overflow: auto;
  padding: 16px 8px;

  .sidebar-el-menu {
    background: transparent;
    border: none;

    :deep(.el-menu-item) {
      color: #94a3b8;
      border-radius: 10px;
      margin: 4px 0;
      transition: all 0.2s;

      &:hover {
        background: rgba(102, 126, 234, 0.1);
        color: #a5b4fc;
      }

      &.is-active {
        background: linear-gradient(135deg, rgba(102, 126, 234, 0.15), rgba(118, 75, 162, 0.15));
        color: #fff;
        font-weight: 500;

        &::before {
          content: '';
          position: absolute;
          left: 0;
          top: 50%;
          transform: translateY(-50%);
          width: 3px;
          height: 24px;
          background: linear-gradient(180deg, #667eea, #764ba2);
          border-radius: 0 3px 3px 0;
        }
      }
    }
  }
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(0, 0, 0, 0.15);

  .collapse-btn {
    width: 100%;
    height: 40px;
    color: #64748b;
    justify-content: center;
    border-radius: 8px;

    &:hover {
      color: #a5b4fc;
      background: rgba(102, 126, 234, 0.1);
    }
  }
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
  background: rgba(15, 23, 42, 0.4);
}

.content {
  flex: 1;
  overflow: auto;
  padding: 24px;
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
