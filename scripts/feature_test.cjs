/**
 * PVE WebUI 功能完整测试 - AI 宠物、应用商店、访问管理
 * 使用已安装的 Playwright (npx playwright)
 */
const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');

const BASE_URL = 'http://localhost:8088';
const SCREENSHOTS_DIR = path.join(__dirname, '..', 'scripts', 'screenshots');
fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });

async function login(page) {
  await page.goto(`${BASE_URL}/login`);
  await page.waitForLoadState('networkidle');
  await page.fill('input[placeholder*="用户名"], input[name="username"]', 'root');
  await page.fill('input[type="password"]', 'password');
  const loginBtn = page.locator('button').filter({ hasText: /登录|Login/i }).first();
  await loginBtn.click();
  await page.waitForURL('**/dashboard**', { timeout: 15000 });
  await page.waitForLoadState('networkidle');
  console.log('✅ 登录成功');
}

async function testAIFloatingPet() {
  console.log('\n=== 测试 AI 悬浮宠物助手 ===');
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  try {
    await login(page);

    // 检查悬浮宠物
    const petBtn = page.locator('.pet-button');
    await petBtn.waitFor({ state: 'visible', timeout: 5000 });
    console.log('✅ 悬浮宠物按钮可见');

    // 点击打开聊天
    await petBtn.click();
    await page.waitForSelector('.chat-panel', { state: 'visible', timeout: 5000 });
    console.log('✅ 聊天面板已打开');

    // 检查快捷操作
    const actions = page.locator('.quick-actions button');
    const count = await actions.count();
    console.log(`✅ 找到 ${count} 个快捷操作按钮`);

    // 点击第一个
    await actions.first().click();
    await page.waitForTimeout(3000);

    const msgs = page.locator('.chat-message.assistant');
    const msgCount = await msgs.count();
    console.log(`✅ AI 回复了 ${msgCount} 条消息`);

    // 手动输入测试
    await page.fill('.chat-input', '你好');
    await page.click('.send-btn');
    await page.waitForTimeout(3000);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'ai_chat.png'), fullPage: false });
    console.log('✅ 手动输入测试通过');

    // 关闭面板
    await page.locator('.header-btn').last().click();
    await page.waitForSelector('.chat-panel', { state: 'hidden', timeout: 3000 });
    console.log('✅ 聊天面板关闭成功');
    return true;
  } catch (e) {
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'ai_error.png') });
    console.error('❌ AI 宠物测试失败:', e.message);
    return false;
  } finally {
    await browser.close();
  }
}

async function testAppStore() {
  console.log('\n=== 测试应用商店 ===');
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  try {
    await login(page);

    // 导航到应用商店
    const sidebarLink = page.locator('a[href="/apps"], li').filter({ hasText: '应用商店' }).first();
    await sidebarLink.click();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    console.log('✅ 进入应用商店');

    // 检查页面内容
    const content = await page.textContent('body');
    console.log(`✅ 页面包含内容: ${content.substring(0, 200)}...`);

    // 检查卡片
    const cards = page.locator('.template-card, .app-card, .el-card, [class*="template"]');
    const cardCount = await cards.count();
    console.log(`✅ 找到 ${cardCount} 个应用模板`);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'app_store.png'), fullPage: true });
    return true;
  } catch (e) {
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'app_store_error.png') });
    console.error('❌ 应用商店测试失败:', e.message);
    return false;
  } finally {
    await browser.close();
  }
}

async function testAccessManagement() {
  console.log('\n=== 测试访问管理 ===');
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  try {
    await login(page);

    // 导航到访问管理
    const sidebarLink = page.locator('a[href="/access"], li').filter({ hasText: /访问|权限/i }).first();
    await sidebarLink.click();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    console.log('✅ 进入访问管理');

    // 检查页面内容
    const content = await page.textContent('body');
    console.log(`✅ 页面包含内容: ${content.substring(0, 200)}...`);

    // 检查标签页
    const tabs = page.locator('.el-tabs__item, [role="tab"]');
    const tabCount = await tabs.count();
    console.log(`✅ 找到 ${tabCount} 个标签页`);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'access_users.png'), fullPage: true });
    return true;
  } catch (e) {
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'access_error.png') });
    console.error('❌ 访问管理测试失败:', e.message);
    return false;
  } finally {
    await browser.close();
  }
}

async function testSidebarNoAI() {
  console.log('\n=== 验证侧边栏无 AI 入口 ===');
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  try {
    await login(page);
    await page.waitForTimeout(1000);

    const sidebarText = await page.locator('.sidebar, .app-sidebar, aside, nav').first().textContent();
    const hasAI = sidebarText.includes('AI 智能中心') || sidebarText.includes('AI 智能体');
    if (hasAI) {
      throw new Error('侧边栏不应该有 AI 入口！');
    }
    console.log('✅ 侧边栏已正确移除 AI 入口');

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'sidebar_no_ai.png') });
    return true;
  } catch (e) {
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'sidebar_error.png') });
    console.error('❌ 侧边栏验证失败:', e.message);
    return false;
  } finally {
    await browser.close();
  }
}

(async () => {
  let passed = 0;
  let failed = 0;

  const tests = [
    ['AI 悬浮宠物助手', testAIFloatingPet],
    ['应用商店', testAppStore],
    ['访问管理', testAccessManagement],
    ['侧边栏无 AI 入口', testSidebarNoAI],
  ];

  for (const [name, fn] of tests) {
    const result = await fn();
    if (result) passed++;
    else failed++;
  }

  console.log('\n' + '='.repeat(50));
  console.log(`通过: ${passed}/${tests.length}, 失败: ${failed}/${tests.length}`);
  if (failed === 0) {
    console.log('🎉 全部测试通过！');
  } else {
    console.log('❌ 部分测试失败，请查看上方日志');
    process.exit(1);
  }
})();
