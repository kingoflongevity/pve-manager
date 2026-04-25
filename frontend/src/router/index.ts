import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import AppLayout from '@/components/common/AppLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录', requiresAuth: false },
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'qemu',
        name: 'QEMUList',
        component: () => import('@/views/qemu/QEMUListView.vue'),
        meta: { title: '虚拟机管理' },
      },
      {
        path: 'lxc',
        name: 'LXCList',
        component: () => import('@/views/lxc/LXCListView.vue'),
        meta: { title: '容器管理' },
      },
      {
        path: 'storage',
        name: 'StorageList',
        component: () => import('@/views/storage/StorageListView.vue'),
        meta: { title: '存储管理' },
      },
      {
        path: 'cluster',
        name: 'ClusterView',
        component: () => import('@/views/cluster/ClusterView.vue'),
        meta: { title: '集群概览' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/SettingsView.vue'),
        meta: { title: '系统设置' },
      },
      {
        path: 'backup',
        name: 'Backup',
        component: () => import('@/views/backup/BackupView.vue'),
        meta: { title: '备份管理' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

/**
 * 全局路由守卫：检查认证状态
 * 未登录用户访问需要认证的页面时，重定向到登录页
 */
router.beforeEach((to, _from, next) => {
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth !== false)
  const token = localStorage.getItem('pve_token')

  if (requiresAuth && !token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && token) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
