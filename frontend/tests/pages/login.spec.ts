import { test, expect } from '@playwright/test'

/**
 * 登录页面 E2E 测试
 *
 * 覆盖范围:
 * - 登录表单渲染验证
 * - 表单校验逻辑
 * - 登录方式切换（密码 / API Token）
 * - 品牌和布局检查
 */
test.describe('登录页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示登录表单', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('.login-title')).toBeVisible()

    // 验证副标题
    await expect(page.getByText('本地私有云虚拟化管理平台')).toBeVisible()

    // 验证登录按钮
    await expect(page.getByRole('button', { name: /登录/ })).toBeVisible()
  })

  test('应该为空白字段显示验证错误', async ({ page }) => {
    // 清空默认节点地址
    const hostInput = page.locator('.login-page .el-input__inner').first()
    await hostInput.fill('')

    // 点击登录按钮触发表单校验
    await page.getByRole('button', { name: /登录/ }).click()

    // 等待错误信息出现
    await page.waitForTimeout(500)

    // 验证错误提示至少有一个
    const errorMessages = page.locator('.el-form-item__error')
    const count = await errorMessages.count()
    expect(count).toBeGreaterThanOrEqual(1)
  })

  test('应该在用户名密码和 API Token 之间切换', async ({ page }) => {
    // 默认应显示密码模式
    await expect(page.getByText('用户名密码')).toBeVisible()
    await expect(page.getByText('API Token')).toBeVisible()

    // 切换到 API Token 模式 - 点击分段控制器的第二个选项
    await page.locator('.el-segmented__item').nth(1).click()
    await page.waitForTimeout(300)

    // 验证 API Token 输入框出现（使用 CSS 选择器定位）
    const apiTokenInput = page.locator('.el-form-item').filter({ hasText: 'API Token' }).locator('input[type="password"]')
    await expect(apiTokenInput).toBeVisible()

    // 切换回密码模式
    await page.locator('.el-segmented__item').nth(0).click()
    await page.waitForTimeout(300)

    // 验证用户名输入框恢复（使用 textInput 过滤掉 radio buttons）
    const usernameInput = page.locator('input[type="text"][placeholder="请输入用户名"]')
    await expect(usernameInput).toBeVisible()
  })

  test('应该具有正确的品牌和布局', async ({ page }) => {
    // 验证 Logo 图标
    const logoIcon = page.locator('.logo-icon')
    await expect(logoIcon).toBeVisible()

    // 验证底部版本信息
    await expect(page.getByText('Proxmox VE Web UI v0.1.0')).toBeVisible()

    // 验证记住我复选框
    await expect(page.getByText('记住节点配置')).toBeVisible()
  })

  test('应该正确显示端口输入框', async ({ page }) => {
    // 端口输入框
    const portInput = page.locator('.port-input .el-input__inner')
    await expect(portInput).toBeVisible()
    await expect(portInput).toHaveValue('8006')
  })

  test('API Token 模式下应该显示 Token 输入校验', async ({ page }) => {
    // 切换到 Token 模式
    await page.locator('.el-segmented__item').nth(1).click()
    await page.waitForTimeout(300)

    // 清空节点地址
    const hostInput = page.locator('.login-page .el-input__inner').first()
    await hostInput.fill('')

    // 点击登录
    await page.getByRole('button', { name: /登录/ }).click()
    await page.waitForTimeout(500)

    // 应该显示校验错误
    const errorMessages = page.locator('.el-form-item__error')
    const count = await errorMessages.count()
    expect(count).toBeGreaterThanOrEqual(1)
  })
})
