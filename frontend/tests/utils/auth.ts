import { Page } from '@playwright/test'

/**
 * 模拟认证令牌（用于测试环境）
 * 该令牌用于绕过实际 API 调用，直接模拟已登录状态
 */
const MOCK_AUTH_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.mock-token.e2e-test'

/**
 * 在页面加载前设置模拟认证信息到 localStorage
 *
 * 使用场景:
 * - 需要直接访问需要认证的页面（如仪表盘、VM列表）
 * - 避免实际调用 PVE API 登录接口
 *
 * 注意:
 * - 必须在 page.goto() 之前调用
 * - 需要配合 page.addInitScript() 使用
 * - 脚本会在页面任何代码执行前同步执行，确保路由守卫能读取到 token
 *
 * @param page - Playwright Page 实例
 */
export async function mockAuth(page: Page): Promise<void> {
  await page.addInitScript((token) => {
    // 在页面任何代码之前同步设置 localStorage
    // 这样 Vue Router 的守卫可以正确读取到 token
    localStorage.setItem('pve_token', token)
  }, MOCK_AUTH_TOKEN)
}

/**
 * 模拟认证并导航到指定页面
 *
 * 这是 mockAuth + page.goto() 的便捷组合方法
 *
 * @param page - Playwright Page 实例
 * @param url - 要导航到的页面路径（相对路径，如 '/' 或 '/qemu'）
 */
export async function gotoAuthenticatedPage(page: Page, url: string = '/'): Promise<void> {
  await mockAuth(page)
  await page.goto(url)
  // 等待页面加载完成
  await page.waitForLoadState('networkidle')
}

/**
 * 清除模拟认证信息
 *
 * @param page - Playwright Page 实例
 */
export async function clearMockAuth(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('pve_token')
  })
}
