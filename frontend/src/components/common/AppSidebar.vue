<template>
  <aside :class="['app-sidebar', { collapsed }]">
    <!-- Logo区域 -->
    <div class="sidebar-logo">
      <el-icon :size="24" class="logo-icon"><Monitor /></el-icon>
      <transition name="fade">
        <span v-show="!collapsed" class="logo-text">PVE Cloud</span>
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
.app-sidebar {
  width: 280px;
  min-width: 280px;
  height: 100vh;
  background: #111827;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
  border-right: 1px solid rgba(59, 130, 246, 0.1);

  &.collapsed {
    width: 64px;
    min-width: 64px;

    .node-select {
      width: 48px;
    }
  }
}

.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  padding: 0 16px;
  gap: 8px;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
  color: #fff;
  font-weight: 600;
  font-size: 15px;
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.2);
  font-family: 'JetBrains Mono', 'Consolas', monospace;

  .logo-icon {
    color: #3b82f6;
    flex-shrink: 0;
  }

  .logo-text {
    letter-spacing: 1px;
    transition: opacity 0.3s ease;
  }
}

.node-selector {
  padding: 12px;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);

  :deep(.node-select) {
    .el-input__wrapper {
      background: rgba(255, 255, 255, 0.03);
      box-shadow: none;
      border: 1px solid rgba(59, 130, 246, 0.15);
      border-radius: 6px;

      &:hover {
        box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.3) inset;
      }

      &.is-focus {
        box-shadow: 0 0 0 1px #3b82f6 inset;
      }
    }

    .el-input__inner {
      color: #e5e7eb;
      font-family: 'JetBrains Mono', 'Consolas', monospace;
    }

    .el-input__placeholder {
      color: #6b7280;
    }
  }
}

.resource-tree-wrapper {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.compact-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 16px;
  gap: 8px;
}

.compact-nav-item {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0;
  color: #6b7280;
  position: relative;

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
    color: #e5e7eb;
    background: rgba(59, 130, 246, 0.1);
  }

  &.active {
    color: #3b82f6;
    background: rgba(59, 130, 246, 0.15);

    &::before {
      background: #3b82f6;
    }
  }
}

.sidebar-footer {
  padding: 8px;
  border-top: 1px solid rgba(59, 130, 246, 0.1);
  background: rgba(0, 0, 0, 0.2);

  .collapse-btn {
    width: 100%;
    height: 40px;
    color: #6b7280;
    justify-content: center;
    border-radius: 6px;

    &:hover {
      color: #e5e7eb;
      background: rgba(59, 130, 246, 0.1);
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
