/**
 * 认证工具函数（E2E 测试用）
 * 
 * 使用 page.evaluate + page.reload() 的组合方式确保 localStorage 被正确设置。
 */
import { Page } from '@playwright/test'

/**
 * 模拟认证并导航到指定页面
 */
export async function gotoAuthenticatedPage(page: Page, url: string = '/'): Promise<void> {
  // 先导航到目标 URL
  await page.goto(url)
  await page.waitForLoadState('networkidle')
  
  // 在页面上设置 localStorage
  await page.evaluate(() => {
    localStorage.setItem('pve_token', 'test-token-e2e')
  })
  
  // 验证设置成功
  const token = await page.evaluate(() => localStorage.getItem('pve_token'))
  if (token !== 'test-token-e2e') {
    throw new Error('Failed to set auth token in localStorage')
  }
  
  // 使用 page.evaluate 调用 Vue Router 跳转到目标页面
  // 这样不会触发完整的页面刷新，token 会被保留
  await page.evaluate((targetUrl) => {
    // 如果已经在正确的页面，直接返回
    if (window.location.pathname === targetUrl || (targetUrl === '/' && window.location.pathname === '/')) {
      return
    }
    // 使用 window.location 跳转到目标页面（会保留 localStorage）
    window.location.href = targetUrl === '/' ? '/' : `/${targetUrl.replace(/^\//, '')}`
  }, url)
  
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(500)
}

/**
 * 清除模拟认证
 */
export async function clearMockAuth(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
}
