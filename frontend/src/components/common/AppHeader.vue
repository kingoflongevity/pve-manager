<template>
  <header class="app-header" height="64px">
    <div class="header-left">
      <div class="breadcrumb-wrapper">
        <span class="breadcrumb-title">首页</span>
        <span v-if="$route.path !== '/dashboard'" class="breadcrumb-separator">/</span>
        <span v-if="$route.path !== '/dashboard'" class="breadcrumb-current">
          {{ $route.meta.title || '页面' }}
        </span>
      </div>
    </div>

    <div class="header-right">
      <div class="search-box">
        <el-input placeholder="搜索功能..." size="small" clearable class="search-input">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <kbd class="search-shortcut">⌘K</kbd>
      </div>

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
import {
  Bell,
  Setting,
  User,
  SwitchButton,
  Search,
  ArrowDown,
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

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
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;

  .breadcrumb-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;

    .breadcrumb-title {
      color: #94a3b8;
      font-weight: 500;
    }

    .breadcrumb-separator {
      color: #475569;
    }

    .breadcrumb-current {
      color: #f8fafc;
      font-weight: 600;
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;

  .search-box {
    position: relative;
    margin-right: 8px;

    .search-input {
      width: 200px;

      :deep(.el-input__wrapper) {
        background: rgba(30, 41, 59, 0.4);
        border: 1px solid rgba(148, 163, 184, 0.1);
        border-radius: 8px;
      }

      :deep(.el-input__inner) {
        color: #e2e8f0;
        font-size: 13px;
      }
    }

    .search-shortcut {
      position: absolute;
      right: 8px;
      top: 50%;
      transform: translateY(-50%);
      padding: 2px 6px;
      background: rgba(51, 65, 85, 0.6);
      border: 1px solid rgba(148, 163, 184, 0.1);
      border-radius: 4px;
      font-family: 'Fira Code', monospace;
      font-size: 10px;
      color: #64748b;
      pointer-events: none;
    }
  }

  .header-icon-btn {
    padding: 8px;
    color: #64748b;
    border-radius: 8px;

    &:hover {
      color: #a5b4fc;
      background: rgba(102, 126, 234, 0.1);
    }
  }

  .header-divider {
    margin: 0 4px;
    background: rgba(148, 163, 184, 0.1);
  }

  .user-dropdown {
    margin-left: 4px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: rgba(102, 126, 234, 0.1);
    }

    .user-avatar {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      font-weight: 600;
      font-size: 13px;
      border: 2px solid rgba(255, 255, 255, 0.1);
    }

    .user-details {
      display: flex;
      flex-direction: column;

      .user-name {
        color: #f8fafc;
        font-size: 13px;
        font-weight: 500;
      }

      .user-role {
        color: #64748b;
        font-size: 11px;
      }
    }

    .dropdown-arrow {
      color: #64748b;
      font-size: 12px;
      transition: transform 0.2s;
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
