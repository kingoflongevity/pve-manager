<template>
  <div class="app-layout">
    <!-- 侧边栏导航 -->
    <AppSidebar :collapsed="isCollapsed" @toggle="isCollapsed = !isCollapsed" />

    <!-- 主体区域 -->
    <el-container class="main-container">
      <!-- 顶部导航 -->
      <AppHeader :collapsed="isCollapsed" @toggle="isCollapsed = !isCollapsed" />

      <!-- 页面内容 -->
      <el-main class="content">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

/**
 * 侧边栏折叠状态
 */
const isCollapsed = ref(false)
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.content {
  flex: 1;
  background: $color-bg-layout;
  overflow: auto;
  padding: 0;
}
</style>
