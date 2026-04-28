<template>
  <div class="app-layout">
    <AppSidebar :collapsed="collapsed" @toggle="toggle" />

    <div class="main-container">
      <AppHeader :collapsed="collapsed" @toggle="toggle" />
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject } from 'vue'
import type { Ref } from 'vue'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'

const collapsed = inject<Ref<boolean>>('sidebarCollapsed')!
const toggle = inject<() => void>('toggleSidebar')!
</script>

<style scoped lang="scss">
@use '@/assets/styles/variables' as *;

.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: $color-bg-base;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
  background: $color-bg-base;
}

.content {
  flex: 1;
  overflow: auto;
  padding: $spacing-6;
}
</style>
