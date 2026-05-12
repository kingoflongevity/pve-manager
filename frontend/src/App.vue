<template>
  <router-view />
  <FloatingPet v-if="isLoggedIn" />
</template>

<script setup lang="ts">
import { ref, provide, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { initTheme } from '@/composables/useTheme'
import { useAuthStore } from '@/stores/auth'
import FloatingPet from '@/components/ai/FloatingPet.vue'

const authStore = useAuthStore()
const router = useRouter()

const SIDEBAR_COLLAPSED_KEY = 'pve_sidebar_collapsed'
const collapsed = ref(localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true')

const isLoggedIn = computed(() => authStore.isLoggedIn)

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
