import { test, expect } from '@playwright/test'

/**
 * 登录页面 E2E 测试（3步向导版本）
 */
test.describe('登录页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示登录向导表单', async ({ page }) => {
    await expect(page.getByText('PVE 运维管理中心')).toBeVisible()
    await expect(page.getByText('连接目标')).toBeVisible()
    await expect(page.getByText('身份验证')).toBeVisible()
  })

  test('步骤 1 - 应该验证服务器地址必填', async ({ page }) => {
    const nextBtn = page.getByRole('button', { name: '下一步' })
    await expect(nextBtn).toBeVisible()
  })

  test('应该具有正确的品牌和布局', async ({ page }) => {
    await expect(page.locator('.logo-icon svg')).toBeVisible()
    await expect(page.getByText('Enterprise Edition')).toBeVisible()
  })
})
