<template>
  <el-header class="app-header" height="56px">
    <div class="header-left">
      <el-button text class="collapse-btn" @click="$emit('toggle')">
        <el-icon :size="18">
          <Fold v-if="!collapsed" />
          <Expand v-else />
        </el-icon>
      </el-button>

      <div class="breadcrumb-wrapper">
        <span class="breadcrumb-title">{{ t('layout.dashboard') }}</span>
        <span v-if="pageTitle" class="breadcrumb-separator">/</span>
        <span v-if="pageTitle" class="breadcrumb-current">{{ pageTitle }}</span>
      </div>
    </div>

    <div class="header-right">
      <el-badge :value="notificationCount" :hidden="notificationCount === 0" class="notification-badge">
        <el-button text class="header-icon-btn" @click="handleNotification">
          <el-icon :size="18"><Bell /></el-icon>
        </el-button>
      </el-badge>

      <el-badge :value="taskStore.taskCount" :hidden="taskStore.taskCount === 0" class="task-badge">
        <el-button text class="header-icon-btn" @click="handleOpenTaskCenter">
          <el-icon :size="18"><List /></el-icon>
        </el-button>
      </el-badge>

      <div class="system-status">
        <span class="status-dot online"></span>
        <span class="status-text">ONLINE</span>
      </div>

      <el-dropdown trigger="click" @command="handleCommand" class="user-dropdown">
        <div class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userInitial }}
          </el-avatar>
          <span class="user-name">{{ userInfo?.username || 'Admin' }}</span>
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
  </el-header>

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
  Bell,
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

const pageTitle = computed(() => (route.meta.title as string) || '')

const notificationCount = ref(3)

function handleNotification() {
  console.log('打开通知')
}

const userInfo = authStore.userInfo

const userInitial = computed(() => {
  const username = userInfo?.username || 'Admin'
  return username.charAt(0).toUpperCase()
})

const taskCenterVisible = ref(false)

function handleOpenTaskCenter() {
  taskCenterVisible.value = true
}

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
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
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #111827;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;

  .collapse-btn {
    padding: 4px;
    color: #9ca3af;

    &:hover {
      color: #3b82f6;
      background: rgba(59, 130, 246, 0.1);
    }
  }

  .breadcrumb-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    font-size: 13px;

    .breadcrumb-title {
      color: #6b7280;
    }

    .breadcrumb-separator {
      color: #4b5563;
    }

    .breadcrumb-current {
      color: #e5e7eb;
      font-weight: 500;
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;

  .header-icon-btn {
    padding: 8px;
    color: #9ca3af;

    &:hover {
      color: #3b82f6;
      background: rgba(59, 130, 246, 0.1);
    }
  }

  .notification-badge {
    :deep(.el-badge__content) {
      transform: translateY(-2px) translateX(4px);
    }
  }

  .system-status {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    background: rgba(16, 185, 129, 0.1);
    border: 1px solid rgba(16, 185, 129, 0.2);
    border-radius: 12px;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    font-size: 11px;

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #10b981;
      box-shadow: 0 0 6px #10b981;
      animation: pulse 2s infinite;
    }

    .status-text {
      color: #10b981;
      letter-spacing: 0.5px;
    }
  }

  .user-dropdown {
    margin-left: 8px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 4px 10px;
    border-radius: 6px;
    transition: all 0.2s;

    &:hover {
      background: rgba(255, 255, 255, 0.05);
    }

    .user-avatar {
      background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
      color: #fff;
      font-weight: 500;
      font-size: 12px;
    }

    .user-name {
      color: #e5e7eb;
      font-size: 13px;
      font-weight: 500;
      font-family: 'JetBrains Mono', 'Consolas', monospace;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@media (max-width: 768px) {
  .user-name {
    display: none;
  }
  
  .system-status {
    display: none;
  }
}
</style>