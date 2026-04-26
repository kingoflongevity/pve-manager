import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 虚拟机列表页面 E2E 测试
 */
test.describe('虚拟机列表页面', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/qemu')
  })

  test('应该加载虚拟机页面', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded')
    await expect(page).toHaveURL(/.*qemu/)
  })
})
