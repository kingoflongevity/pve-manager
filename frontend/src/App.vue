<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { initTheme } from '@/composables/useTheme'

// 初始化主题（应用启动时仅调用一次）
onMounted(() => {
  initTheme()
})
</script>
