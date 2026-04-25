import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright E2E 测试配置
 * 用于 PVE Web Management Panel 的浏览器自动化测试
 *
 * 主要特性:
 * - 使用 Chromium 浏览器
 * - 失败时自动截图
 * - 重试时录制视频
 * - 顺序执行测试用例（避免 UI 状态冲突）
 */
export default defineConfig({
  // 测试目录
  testDir: './tests',

  // 完全并行执行测试文件，但每个文件内顺序执行
  fullyParallel: true,

  // 失败时中断
  forbidOnly: !!process.env.CI,

  // 重试次数
  retries: process.env.CI ? 2 : 1,

  // 工作进程数（UI 测试建议顺序执行）
  workers: 1,

  // 报告器
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],

  // 共享配置
  use: {
    // 基础 URL
    baseURL: 'http://localhost:3000',

    // 失败时截图
    screenshot: 'only-on-failure',

    // 重试时录制视频
    video: 'retain-on-failure',

    // 视口大小
    viewport: { width: 1280, height: 720 },

    // 操作超时
    actionTimeout: 10_000,

    // 导航超时
    navigationTimeout: 15_000,

    // 测试超时（单个测试用例）
    timeout: 30_000,

    // 追踪（失败时保留）
    trace: 'on-first-retry',
  },

  // 项目配置
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // 本地开发服务器配置（可选，如果需要自动启动）
  // webServer: {
  //   command: 'npm run dev',
  //   url: 'http://localhost:3000',
  //   reuseExistingServer: !process.env.CI,
  // },
})
