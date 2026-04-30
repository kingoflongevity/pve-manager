<template>
  <header class="app-header">
    <div class="header-left">
      <el-button text class="menu-toggle-btn" @click="$emit('toggle')">
        <el-icon :size="18"><Fold /></el-icon>
      </el-button>
      <div class="breadcrumb-wrapper">
        <span class="breadcrumb-title">首页</span>
        <span v-if="$route.path !== '/'" class="breadcrumb-separator">/</span>
        <span v-if="$route.path !== '/'" class="breadcrumb-current">
          {{ $route.meta.title || '页面' }}
        </span>
      </div>
    </div>

    <div class="header-right">
      <el-tooltip :content="isLight ? '暗色主题' : '亮色主题'" placement="bottom">
        <el-button text class="header-icon-btn" @click="toggleTheme">
          <el-icon :size="18">
            <Sunny v-if="isDark" />
            <Moon v-else />
          </el-icon>
        </el-button>
      </el-tooltip>

      <el-tooltip content="通知" placement="bottom">
        <el-button text class="header-icon-btn">
          <el-icon :size="18"><Bell /></el-icon>
        </el-button>
      </el-tooltip>

      <el-tooltip content="设置" placement="bottom">
        <el-button text class="header-icon-btn">
          <el-icon :size="18"><Setting /></el-icon>
        </el-button>
      </el-tooltip>

      <el-divider direction="vertical" class="header-divider" />

      <el-dropdown trigger="click" @command="handleCommand" class="user-dropdown">
        <div class="user-info">
          <el-avatar :size="32" class="user-avatar">
            A
          </el-avatar>
          <div class="user-details">
            <span class="user-name">管理员</span>
            <span class="user-role">Admin</span>
          </div>
          <el-icon class="dropdown-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">
              <el-icon><User /></el-icon>
              个人信息
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              系统设置
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import {
  Bell,
  Setting,
  User,
  SwitchButton,
  ArrowDown,
  Fold,
  Sunny,
  Moon,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()
const { toggleTheme, isLight, isDark } = useTheme()

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
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

<style scoped lang="scss">
@use '@/assets/styles/variables' as *;

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 $spacing-6;
  background: $color-bg-container;
  border-bottom: 1px solid $color-border-light;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-4;

  .menu-toggle-btn {
    padding: $spacing-2;
    color: $color-text-secondary;
    border-radius: $radius-base;

    &:hover {
      color: $green-400;
      background: $color-bg-hover;
    }
  }

  .breadcrumb-wrapper {
    display: flex;
    align-items: center;
    gap: $spacing-2;
    font-size: $font-size-base;

    .breadcrumb-title {
      color: $color-text-secondary;
      font-weight: $font-weight-medium;
    }

    .breadcrumb-separator {
      color: $color-text-disabled;
    }

    .breadcrumb-current {
      color: $color-text-primary;
      font-weight: $font-weight-semibold;
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: $spacing-2;

  .header-icon-btn {
    padding: $spacing-2;
    color: $color-text-secondary;
    border-radius: $radius-base;

    &:hover {
      color: $green-400;
      background: $color-bg-hover;
    }
  }

  .header-divider {
    margin: 0 4px;
    background: $color-border-light;
  }

  .user-dropdown {
    margin-left: 4px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: $spacing-3;
    padding: $spacing-2 $spacing-3;
    border-radius: $radius-md;
    cursor: pointer;
    transition: $transition-base;

    &:hover {
      background: $color-bg-hover;
    }

    .user-avatar {
      background: linear-gradient(135deg, $green-500 0%, $emerald-500 100%);
      color: #fff;
      font-weight: $font-weight-semibold;
      font-size: $font-size-sm;
      border: 2px solid $color-border-light;
    }

    .user-details {
      display: flex;
      flex-direction: column;

      .user-name {
        color: $color-text-primary;
        font-size: $font-size-sm;
        font-weight: $font-weight-medium;
      }

      .user-role {
        color: $color-text-secondary;
        font-size: $font-size-xs;
      }
    }

    .dropdown-arrow {
      color: $color-text-secondary;
      font-size: $font-size-xs;
      transition: transform $transition-base;
    }
  }
}

@media (max-width: 768px) {
  .search-box {
    display: none;
  }

  .user-details {
    display: none;
  }
}
</style>
