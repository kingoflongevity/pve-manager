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
        path: 'storage/:node/:storage',
        name: 'StorageDetail',
        component: () => import('@/views/storage/StorageDetailView.vue'),
        meta: { title: '存储详情' },
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
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => import('@/views/monitor/MonitorView.vue'),
        meta: { title: '监控中心' },
      },
      {
        path: 'access',
        name: 'Access',
        component: () => import('@/views/access/AccessView.vue'),
        meta: { title: '访问管理' },
      },
      {
        path: 'nodes',
        name: 'NodeList',
        component: () => import('@/views/node/NodeListView.vue'),
        meta: { title: '节点管理' },
      },
      {
        path: 'nodes/:nodeName',
        name: 'NodeDetail',
        component: () => import('@/views/node/NodeDetailView.vue'),
        meta: { title: '节点详情' },
      },
      {
        path: 'qemu/:node/:vmid',
        name: 'QEMUDetail',
        component: () => import('@/views/qemu/QEMUDetailView.vue'),
        meta: { title: '虚拟机详情' },
      },
      {
        path: 'lxc/:node/:vmid',
        name: 'LXCDetail',
        component: () => import('@/views/lxc/LXCDetailView.vue'),
        meta: { title: '容器详情' },
      },
    ],
  },
  // 全屏控制台（不需要 AppLayout 布局）
  {
    path: '/qemu/:node/:vmid/console',
    name: 'QEMUConsole',
    component: () => import('@/components/qemu/QEMUConsoleView.vue'),
    meta: { title: '虚拟机控制台', requiresAuth: true },
  },
  // 统一控制台页面（支持 QEMU 和 LXC）
  {
    path: '/console/:node/:vmid/:vmType',
    name: 'ConsoleView',
    component: () => import('@/views/console/ConsoleView.vue'),
    meta: { title: '远程控制', requiresAuth: true },
    props: true,
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
