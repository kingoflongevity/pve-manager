import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 仪表盘页面 E2E 测试
 *
 * 覆盖范围:
 * - 仪表盘页面加载（模拟认证）
 * - 资源卡片显示（CPU、内存、磁盘、网络）
 * - 状态汇总组件
 * - 快捷操作区域
 * - 任务列表
 */
test.describe('仪表盘页面', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/')
  })

  test('应该在认证后加载仪表盘页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('.page-title')).toBeVisible()

    // 验证页面描述
    await expect(page.getByText('查看节点资源使用情况和虚拟机状态')).toBeVisible()
  })

  test('应该显示资源卡片（CPU、内存、磁盘、网络）', async ({ page }) => {
    // CPU 使用率卡片
    await expect(page.getByText('CPU 使用率')).toBeVisible()
    // 内存使用率卡片
    await expect(page.getByText('内存使用率')).toBeVisible()
    // 磁盘使用率卡片
    await expect(page.getByText('磁盘使用率')).toBeVisible()
    // 网络 I/O 卡片
    await expect(page.getByText('网络 I/O')).toBeVisible()
  })

  test('应该显示状态汇总', async ({ page }) => {
    // 验证状态汇总标题 - 使用 CSS 选择器
    await expect(page.locator('.summary-title')).toBeVisible()

    // 验证状态项
    await expect(page.locator('.status-label', { hasText: '运行中' })).toBeVisible()
    await expect(page.locator('.status-label', { hasText: '已停止' })).toBeVisible()
    await expect(page.locator('.status-label', { hasText: '错误' })).toBeVisible()
    await expect(page.locator('.status-label', { hasText: '已暂停' })).toBeVisible()
  })

  test('应该显示快捷操作', async ({ page }) => {
    // 验证快捷操作区域存在
    await expect(page.locator('.quick-actions')).toBeVisible()

    // 验证有操作按钮（使用 first() 避免 strict mode violation）
    const actionBtns = page.locator('.action-btn')
    expect(await actionBtns.count()).toBeGreaterThanOrEqual(1)
    await expect(actionBtns.first()).toBeVisible()
  })

  test('应该显示任务列表', async ({ page }) => {
    // 验证任务卡片标题
    await expect(page.locator('.card-header').filter({ hasText: '最近任务' })).toBeVisible()
    await expect(page.getByText('查看全部')).toBeVisible()

    // 验证任务项存在
    const taskItems = page.locator('.task-item')
    await expect(taskItems).toHaveCount(5)
  })

  test('应该显示节点信息', async ({ page }) => {
    // 验证节点信息卡片
    await expect(page.locator('.card-header').filter({ hasText: '节点信息' })).toBeVisible()
    await expect(page.getByText('在线')).toBeVisible()

    // 验证节点详情字段
    await expect(page.getByText('pve-node-01')).toBeVisible()
    await expect(page.getByText('8.1.4')).toBeVisible()
  })

  test('刷新数据按钮应该可点击', async ({ page }) => {
    const refreshBtn = page.getByRole('button', { name: '刷新数据' })
    await expect(refreshBtn).toBeVisible()
    await expect(refreshBtn).toBeEnabled()
    await refreshBtn.click()
  })
})
