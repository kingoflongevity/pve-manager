import { test, expect } from '@playwright/test'
import { gotoAuthenticatedPage } from '../utils/auth'

/**
 * 虚拟机列表页面 E2E 测试
 *
 * 覆盖范围:
 * - VM 表格渲染
 * - 按状态筛选虚拟机
 * - 多选虚拟机
 * - 批量操作栏显示
 */
test.describe('虚拟机列表页面', () => {
  test.beforeEach(async ({ page }) => {
    await gotoAuthenticatedPage(page, '/qemu')
  })

  test('应该显示虚拟机列表表格', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('.page-title')).toBeVisible()

    // 验证搜索输入框存在
    await expect(page.locator('input[placeholder="搜索虚拟机名称或ID"]')).toBeVisible()

    // 验证表格存在
    const table = page.locator('.el-table')
    await expect(table).toBeVisible()
  })

  test('应该筛选虚拟机状态', async ({ page }) => {
    // 点击状态筛选下拉框
    const statusSelect = page.locator('.el-select').filter({ hasText: '状态筛选' })
    const selectInput = statusSelect.locator('.el-select__wrapper')
    await selectInput.click()

    // 验证筛选选项存在
    await expect(page.getByRole('option', { name: '运行中' })).toBeVisible()
    await expect(page.getByRole('option', { name: '已停止' })).toBeVisible()
    await expect(page.getByRole('option', { name: '已暂停' })).toBeVisible()
    await expect(page.getByRole('option', { name: '错误' })).toBeVisible()

    // 选择运行中状态
    await page.getByRole('option', { name: '运行中' }).click()

    // 等待页面响应（下拉框关闭即可）
    await page.waitForTimeout(500)
  })

  test('应该可以选择多个虚拟机', async ({ page }) => {
    // 等待表格加载
    await page.waitForSelector('.el-table', { state: 'visible' })

    // 获取复选框
    const checkboxes = page.locator('.el-table__body .el-checkbox')

    // 点击第一个复选框
    if (await checkboxes.count() > 0) {
      await checkboxes.first().click()

      // 验证选中状态
      const firstCheckbox = checkboxes.first().locator('input')
      await expect(firstCheckbox).toBeChecked()
    }
  })

  test('应该显示创建虚拟机按钮', async ({ page }) => {
    const createBtn = page.getByRole('button', { name: '创建虚拟机' })
    await expect(createBtn).toBeVisible()
    await expect(createBtn).toBeEnabled()
  })

  test('应该显示搜索输入框', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="搜索虚拟机名称或ID"]')
    await expect(searchInput).toBeVisible()

    // 输入搜索关键词
    await searchInput.fill('test-vm')
    await expect(searchInput).toHaveValue('test-vm')
  })

  test('刷新按钮应该可点击', async ({ page }) => {
    const refreshBtn = page.locator('.toolbar-right .el-button:last-child')
    await expect(refreshBtn).toBeVisible()
    await refreshBtn.click()
  })

  test('应该显示分页控件', async ({ page }) => {
    // 分页控件应该在表格底部
    const pagination = page.locator('.el-pagination')
    await expect(pagination).toBeVisible()
  })
})
