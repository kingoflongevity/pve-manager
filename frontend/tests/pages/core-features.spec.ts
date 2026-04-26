import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * PVE WebUI 核心功能 E2E 测试
 */
test.describe('PVE WebUI 核心功能', () => {
  
  test.describe('认证和路由', () => {
    test('未认证用户访问仪表盘应跳转到登录页', async ({ page }) => {
      await page.goto('/')
      await page.waitForLoadState('domcontentloaded')
      await expect(page).toHaveURL(/.*login/)
    })

    test('认证用户应该能访问仪表盘', async ({ page }) => {
      await gotoAuthenticatedPage(page, '/')
      await expect(page.locator('.page-title')).toBeVisible()
    })

    test('登录页面应该正确渲染', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await expect(page.getByText('PVE 运维管理中心')).toBeVisible()
    })
  })

  test.describe('侧边栏导航', () => {
    test.beforeEach(async ({ page }) => {
      await gotoAuthenticatedPage(page, '/')
    })

    test('侧边栏应该可见', async ({ page }) => {
      await expect(page.locator('.app-sidebar')).toBeVisible()
    })

    test('应该显示折叠按钮', async ({ page }) => {
      await expect(page.locator('.collapse-btn').first()).toBeVisible()
    })
  })

  test.describe('仪表盘', () => {
    test.beforeEach(async ({ page }) => {
      await gotoAuthenticatedPage(page, '/')
    })

    test('应该显示页面标题', async ({ page }) => {
      await expect(page.locator('.page-title')).toBeVisible()
    })

    test('刷新数据按钮应该可用', async ({ page }) => {
      const refreshBtn = page.getByRole('button', { name: '刷新数据' })
      await expect(refreshBtn).toBeVisible()
    })
  })

  test.describe('UI 响应式', () => {
    test('应该在小屏幕上适配', async ({ page }) => {
      await gotoAuthenticatedPage(page, '/')
      await page.setViewportSize({ width: 375, height: 667 })
      await expect(page.locator('.page-title')).toBeVisible()
    })

    test('头部应该在所有页面显示', async ({ page }) => {
      await gotoAuthenticatedPage(page, '/')
      await expect(page.locator('.app-header')).toBeVisible()
    })
  })
})
