import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 仪表盘页面 E2E 测试
 */
test.describe('仪表盘页面', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/')
  })

  test('应该在认证后加载仪表盘页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('.page-title')).toBeVisible()
  })

  test('应该显示刷新数据按钮', async ({ page }) => {
    const refreshBtn = page.getByRole('button', { name: '刷新数据' })
    await expect(refreshBtn).toBeVisible()
  })
})
