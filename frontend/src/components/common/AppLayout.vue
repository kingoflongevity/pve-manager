<template>
  <div class="app-layout">
    <!-- 侧边栏导航 -->
    <el-aside :width="isCollapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <el-icon :size="28"><Monitor /></el-icon>
        <span v-show="!isCollapsed" class="logo-text">PVE 管理面板</span>
      </div>
      <el-menu
        :default-active="currentRoute"
        :collapse="isCollapsed"
        router
        background-color="#001529"
        text-color="#ffffffb3"
        active-text-color="#409eff"
      >
        <el-menu-item index="/">
          <el-icon><DataBoard /></el-icon>
          <template #title>{{ t('layout.dashboard') }}</template>
        </el-menu-item>
        <el-menu-item index="/qemu">
          <el-icon><Cpu /></el-icon>
          <template #title>{{ t('layout.qemu') }}</template>
        </el-menu-item>
        <el-menu-item index="/lxc">
          <el-icon><Box /></el-icon>
          <template #title>{{ t('layout.lxc') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage">
          <el-icon><FolderOpened /></el-icon>
          <template #title>{{ t('layout.storage') }}</template>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('layout.settings') }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主体区域 -->
    <el-container class="main-container">
      <!-- 顶部栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapsed = !isCollapsed">
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <span class="node-info">
            {{ currentNode?.host || '' }}:{{ currentNode?.port || '' }}
          </span>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <el-icon><UserFilled /></el-icon>
              <span>{{ userInfo?.username || '用户' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  {{ t('common.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 页面内容 -->
      <el-main class="content">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

// 导航菜单图标
import {
  Monitor,
  DataBoard,
  Cpu,
  Box,
  FolderOpened,
  Setting,
  Fold,
  Expand,
  UserFilled,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()

// 侧边栏折叠状态
const isCollapsed = ref(false)

// 当前路由路径（用于菜单高亮）
const currentRoute = computed(() => route.path)

// 页面标题
const pageTitle = computed(() => (route.meta.title as string) || '')

// 当前节点和用戶信息
const currentNode = authStore.currentNode
const userInfo = authStore.userInfo

/**
 * 处理用户下拉菜单命令
 */
function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<style lang="scss" scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
}

.sidebar {
  background-color: #001529;
  transition: width 0.3s ease;
  overflow: hidden;

  .logo {
    display: flex;
    align-items: center;
  justify-content: center;
    height: 56px;
    color: #fff;
    font-size: 16px;
    font-weight: 600;
    gap: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .logo-text {
      white-space: nowrap;
    }
  }

  :deep(.el-menu) {
    border-right: none;
  }
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 20px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      color: #606266;

      &:hover {
        color: #409eff;
      }
    }

    .page-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .node-info {
      color: #909399;
      font-size: 13px;
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 6px;
      cursor: pointer;
      color: #606266;
    }
  }
}

.content {
  flex: 1;
  background: #f5f7fa;
  overflow: auto;
}
</style>
