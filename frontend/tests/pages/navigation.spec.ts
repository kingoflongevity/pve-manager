import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 侧边栏导航测试
 */
test.describe('侧边栏导航', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/')
  })

  /** 页面导航配置表 */
  const pages = [
    { name: '仪表盘', path: '/' },
    { name: '虚拟机管理', path: '/qemu' },
    { name: '容器管理', path: '/lxc' },
    { name: '存储管理', path: '/storage' },
    { name: '备份管理', path: '/backup' },
    { name: '监控中心', path: '/monitor' },
    { name: '访问管理', path: '/access' },
    { name: '节点管理', path: '/nodes' },
    { name: '系统设置', path: '/settings' },
  ]

  for (const pg of pages) {
    test(`应该导航到 ${pg.name}`, async ({ page }) => {
      await page.goto(pg.path)
      await page.waitForLoadState('domcontentloaded')
      // 验证 URL 正确
      const url = page.url()
      expect(url).toContain(pg.path === '/' ? 'localhost' : pg.path)
    })
  }

  test('应该显示侧边栏折叠按钮', async ({ page }) => {
    await expect(page.locator('.collapse-btn').first()).toBeVisible()
  })
})
