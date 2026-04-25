import { Page } from '@playwright/test'

/**
 * 截图配置常量
 */
const SCREENSHOT_OPTIONS = {
  fullPage: false,
  type: 'png' as const,
}

/**
 * 对当前页面进行截图
 *
 * 使用统一的视口和截图配置，确保截图的一致性
 * 可用于视觉回归对比或失败调试
 *
 * @param page - Playwright Page 实例
 * @param name - 截图文件名称（不含路径和扩展名）
 * @param fullPage - 是否截取完整页面（默认 false，只截取视口）
 * @returns 截图缓冲区
 */
export async function takeScreenshot(
  page: Page,
  name: string,
  fullPage: boolean = false,
): Promise<Buffer> {
  return page.screenshot({
    path: `test-results/screenshots/${name}.png`,
    fullPage,
    type: 'png',
  })
}

/**
 * 对指定元素进行截图
 *
 * @param page - Playwright Page 实例
 * @param selector - CSS 选择器
 * @param name - 截图文件名称
 */
export async function takeElementScreenshot(
  page: Page,
  selector: string,
  name: string,
): Promise<void> {
  const element = page.locator(selector)
  await element.screenshot({
    path: `test-results/screenshots/${name}.png`,
    type: 'png',
  })
}
