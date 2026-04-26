/**
 * 主题管理 composable
 * 支持暗色/亮色主题切换，持久化到 localStorage
 */
import { ref, computed } from 'vue'

export type ThemeMode = 'dark' | 'light'

const THEME_STORAGE_KEY = 'pve_theme'

// 全局主题状态
const currentTheme = ref<ThemeMode>('dark')

/**
 * 初始化主题（仅在应用启动时调用）
 */
export function initTheme(): ThemeMode {
  const saved = localStorage.getItem(THEME_STORAGE_KEY) as ThemeMode | null
  if (saved && (saved === 'dark' || saved === 'light')) {
    currentTheme.value = saved
  }
  applyTheme(currentTheme.value)
  return currentTheme.value
}

/**
 * 主题切换 hook
 */
export function useTheme() {
  /**
   * 切换主题
   */
  function toggleTheme(): void {
    currentTheme.value = currentTheme.value === 'dark' ? 'light' : 'dark'
    localStorage.setItem(THEME_STORAGE_KEY, currentTheme.value)
    applyTheme(currentTheme.value)
  }

  /**
   * 设置为指定主题
   */
  function setTheme(theme: ThemeMode): void {
    currentTheme.value = theme
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    applyTheme(theme)
  }

  /**
   * 是否为亮色主题
   */
  const isLight = computed(() => currentTheme.value === 'light')

  /**
   * 是否为暗色主题
   */
  const isDark = computed(() => currentTheme.value === 'dark')

  return {
    theme: currentTheme,
    toggleTheme,
    setTheme,
    isLight,
    isDark,
  }
}

/**
 * 应用主题到 DOM
 */
function applyTheme(theme: ThemeMode): void {
  const root = document.documentElement
  if (theme === 'light') {
    root.classList.add('theme-light')
    root.classList.remove('theme-dark')
  } else {
    root.classList.add('theme-dark')
    root.classList.remove('theme-light')
  }
}
