<template>
  <el-header class="app-header" height="56px">
    <!-- 左侧：折叠按钮 + 面包屑 -->
    <div class="header-left">
      <el-button text class="collapse-btn" @click="$emit('toggle')">
        <el-icon :size="18">
          <Fold v-if="!collapsed" />
          <Expand v-else />
        </el-icon>
      </el-button>

      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/' }">
          {{ t('layout.dashboard') }}
        </el-breadcrumb-item>
        <el-breadcrumb-item v-if="pageTitle">
          {{ pageTitle }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 右侧：搜索 + 通知 + 用户菜单 -->
    <div class="header-right">
      <!-- 全局搜索 -->
      <el-input
        v-model="searchQuery"
        :placeholder="t('common.search')"
        class="header-search"
        clearable
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <!-- 通知中心 -->
      <el-badge :value="notificationCount" :hidden="notificationCount === 0" class="notification-badge">
        <el-button text class="header-icon-btn" @click="handleNotification">
          <el-icon :size="18"><Bell /></el-icon>
        </el-button>
      </el-badge>

      <!-- 任务中心 -->
      <el-badge :value="taskStore.taskCount" :hidden="taskStore.taskCount === 0" class="task-badge">
        <el-button text class="header-icon-btn" @click="handleOpenTaskCenter">
          <el-icon :size="18"><List /></el-icon>
        </el-button>
      </el-badge>

      <!-- 用户菜单 -->
      <el-dropdown trigger="click" @command="handleCommand" class="user-dropdown">
        <div class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userInitial }}
          </el-avatar>
          <span class="user-name">{{ userInfo?.username || '用户' }}</span>
          <el-icon :size="12"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">
              <el-icon><User /></el-icon>
              个人信息
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              {{ t('layout.settings') }}
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              {{ t('common.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>

  <!-- 任务中心面板 -->
  <TaskCenter v-model="taskCenterVisible" />
</template>
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useTaskStore } from '@/stores/tasks'
import {
  Fold,
  Expand,
  Search,
  Bell,
  ArrowDown,
  User,
  Setting,
  SwitchButton,
  List,
} from '@element-plus/icons-vue'
import TaskCenter from './TaskCenter.vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const taskStore = useTaskStore()

// 页面标题
const pageTitle = computed(() => (route.meta.title as string) || '')

// 搜索
const searchQuery = ref('')

function handleSearch() {
  // TODO: 实现全局搜索功能
  console.log('搜索:', searchQuery.value)
}

// 通知（模拟数据）
const notificationCount = ref(3)

function handleNotification() {
  // TODO: 打开通知面板
  console.log('打开通知')
}

// 用户信息
const userInfo = authStore.userInfo

const userInitial = computed(() => {
  const username = userInfo?.username || '用户'
  return username.charAt(0).toUpperCase()
})

// 任务中心
const taskCenterVisible = ref(false)

function handleOpenTaskCenter() {
  taskCenterVisible.value = true
}

/**
 * 处理用户下拉菜单命令
 */
function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      // TODO: 跳转到个人信息页面
      console.log('查看个人信息')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      authStore.logout()
      router.push('/login')
      break
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 $spacing-6;
  background: $header-bg;
  border-bottom: 1px solid $color-border-light;
  box-shadow: $shadow-header;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-4;

  .collapse-btn {
    padding: $spacing-1;
    color: $color-text-regular;

    &:hover {
      color: $color-primary;
      background: $primary-1;
    }
  }

  :deep(.el-breadcrumb) {
    .el-breadcrumb__item {
      .el-breadcrumb__inner {
        color: $color-text-secondary;

        &:hover {
          color: $color-primary;
        }

        &.is-link {
          font-weight: $font-weight-regular;
        }
      }

      &:last-child .el-breadcrumb__inner {
        color: $color-text-primary;
        font-weight: $font-weight-medium;
      }
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: $spacing-4;

  .header-search {
    width: 240px;

    :deep(.el-input__wrapper) {
      background: $gray-2;
      box-shadow: none;

      &:hover {
        box-shadow: 0 0 0 1px $color-border-base inset;
      }

      &.is-focus {
        box-shadow: 0 0 0 1px $color-primary inset;
      }
    }
  }

  .header-icon-btn {
    padding: $spacing-2;
    color: $color-text-regular;
    position: relative;

    &:hover {
      color: $color-primary;
      background: $primary-1;
    }
  }

  .notification-badge {
    :deep(.el-badge__content) {
      transform: translateY(-2px) translateX(4px);
    }
  }

  .user-dropdown {
    margin-left: $spacing-2;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: $spacing-2;
    cursor: pointer;
    padding: $spacing-1 $spacing-3;
    border-radius: $radius-full;
    transition: $transition-fast;

    &:hover {
      background: $gray-2;
    }

    .user-avatar {
      background: $gradient-primary;
      color: #fff;
      font-weight: $font-weight-medium;
    }

    .user-name {
      color: $color-text-primary;
      font-size: $font-size-sm;
      font-weight: $font-weight-medium;
    }
  }
}

// 响应式
@media (max-width: $breakpoint-md) {
  .header-search {
    display: none;
  }

  .user-name {
    display: none;
  }
}
</style>
