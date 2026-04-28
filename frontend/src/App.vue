<template>
  <router-view />
</template>

<script setup lang="ts">
import { ref, provide, onMounted } from 'vue'
import { initTheme } from '@/composables/useTheme'

const SIDEBAR_COLLAPSED_KEY = 'pve_sidebar_collapsed'

const collapsed = ref(localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true')

function toggleCollapsed() {
  collapsed.value = !collapsed.value
  localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(collapsed.value))
}

provide('sidebarCollapsed', collapsed)
provide('toggleSidebar', toggleCollapsed)

onMounted(() => {
  initTheme()
})
</script>
