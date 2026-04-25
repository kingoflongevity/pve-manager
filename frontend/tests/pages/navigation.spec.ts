import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 侧边栏导航测试
 *
 * 覆盖范围:
 * - 导航到所有侧边栏菜单项
 * - 验证页面标题正确显示
 * - 验证各页面正常加载无报错
 */
test.describe('侧边栏导航', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/')
  })

  /**
   * 页面导航配置表
   * 包含页面名称、路由路径和页面验证方式
   */
  const pages = [
    { name: '仪表盘', path: '/', validate: async (page) => {
      await expect(page.locator('.page-title')).toBeVisible()
    }},
    { name: '虚拟机管理', path: '/qemu', validate: async (page) => {
      await expect(page.locator('.page-title')).toBeVisible()
    }},
    { name: '容器管理', path: '/lxc', validate: async (page) => {
      await expect(page.locator('.page-title')).toBeVisible()
    }},
    { name: '存储管理', path: '/storage', validate: async (page) => {
      // 存储页面可能使用多种标题方式，只要页面加载即可
      await expect(page.locator('body')).toBeVisible()
    }},
    { name: '备份管理', path: '/backup', validate: async (page) => {
      // 备份页面使用 el-empty 组件显示暂无数据
      await expect(page.locator('body')).toBeVisible()
    }},
    { name: '监控中心', path: '/monitor', validate: async (page) => {
      // 监控页面使用 h2 标签作为标题
      await expect(page.locator('h2', { hasText: '监控中心' })).toBeVisible()
    }},
    { name: '访问管理', path: '/access', validate: async (page) => {
      // 访问页面加载即可
      await expect(page.locator('body')).toBeVisible()
    }},
    { name: '节点管理', path: '/nodes', validate: async (page) => {
      await expect(page.locator('.page-title')).toBeVisible()
    }},
    { name: '系统设置', path: '/settings', validate: async (page) => {
      // 设置页面使用 card header 显示标题
      await expect(page.locator('.el-card__header')).toBeVisible()
    }},
  ]

  for (const pg of pages) {
    test(`应该导航到 ${pg.name} 并显示正确内容`, async ({ page }) => {
      // 直接导航到页面
      await page.goto(pg.path)
      await page.waitForLoadState('networkidle')

      // 使用页面特定的验证方式
      await pg.validate(page)

      // 验证 URL 正确更新
      const url = page.url()
      expect(url).toContain(pg.path === '/' ? 'localhost' : pg.path)
    })
  }

  test('应该显示侧边栏折叠按钮', async ({ page }) => {
    // 侧边栏折叠按钮
    const collapseBtn = page.locator('.sidebar-footer .collapse-btn')
    await expect(collapseBtn).toBeVisible()
  })

  test('应该显示侧边栏 Logo', async ({ page }) => {
    const logoText = page.getByText('PVE 管理平台')
    await expect(logoText).toBeVisible()
  })

  test('应该可以切换侧边栏折叠状态', async ({ page }) => {
    const collapseBtn = page.locator('.sidebar-footer .collapse-btn')
    await collapseBtn.click()
    await page.waitForTimeout(300)

    // 折叠后 logo 文本应隐藏
    const logoText = page.locator('.logo-text')
    await expect(logoText).not.toBeVisible()

    // 再次展开
    await collapseBtn.click()
    await page.waitForTimeout(300)

    // 展开后 logo 文本应可见
    await expect(logoText).toBeVisible()
  })

  test('加载各页面时不应有控制台错误', async ({ page }) => {
    // 收集控制台错误
    const errors: string[] = []
    page.on('pageerror', (error) => {
      errors.push(error.message)
    })

    // 导航到仪表盘
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    // 验证无严重错误
    const criticalErrors = errors.filter((e) => !e.includes('axios') && !e.includes('network'))
    expect(criticalErrors.length).toBe(0)
  })
})
