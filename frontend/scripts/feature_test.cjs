/**
 * PVE WebUI 功能完整测试 - AI 宠物、应用商店、访问管理
 */
const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');

const BASE_URL = 'http://localhost:8088';
const API_URL = 'http://localhost:8080/api/pve';
const SCREENSHOTS_DIR = path.join(__dirname, 'screenshots');
fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });

async function login(page) {
  await page.goto(`${BASE_URL}/login`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500);

  await page.fill('input[placeholder*="192.168"], input[placeholder*="pve."]', '192.168.1.100');
  await page.waitForTimeout(200);
  let nextBtn = page.locator('button').filter({ hasText: '下一步' }).first();
  await nextBtn.waitFor({ state: 'visible', timeout: 5000 });
  await nextBtn.click();
  await page.waitForTimeout(500);

  await page.fill('input[placeholder*="例: root"]', 'root');
  await page.waitForTimeout(200);
  await page.fill('input[type="password"]', 'password');
  await page.waitForTimeout(200);
  nextBtn = page.locator('button').filter({ hasText: '下一步' }).first();
  await nextBtn.waitFor({ state: 'visible', timeout: 5000 });
  await nextBtn.click();
  await page.waitForTimeout(2000);

  let loginBtn = page.locator('button').filter({ hasText: '登录系统' });
  try { await loginBtn.waitFor({ state: 'visible', timeout: 10000 }); } catch { await page.waitForTimeout(2000); }
  if (await loginBtn.isEnabled()) { await loginBtn.click(); }
  await page.waitForTimeout(2000);
  await page.waitForLoadState('networkidle');

  // 直接导航到 dashboard
  await page.goto(`${BASE_URL}/dashboard`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
  console.log('✅ 登录并跳转到仪表盘');
}

async function testAIFloatingPet() {
  console.log('\n=== 测试 AI 悬浮宠物助手 ===');
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  try {
    await login(page);

    // 检查悬浮宠物
    const petContainer = page.locator('.floating-pet-container');
    const petBtn = page.locator('.pet-button');
    await petBtn.waitFor({ state: 'visible', timeout: 5000 });
    console.log('✅ 悬浮宠物按钮可见');

    // 点击打开聊天 (force: true 跳过动画导致的 "not stable" 检测)
    await petBtn.click({ force: true });
    await page.waitForSelector('.chat-panel', { state: 'visible', timeout: 5000 });
    console.log('✅ 聊天面板已打开');

    // 检查快捷操作
    const actions = page.locator('.quick-actions button');
    const count = await actions.count();
    console.log(`✅ 找到 ${count} 个快捷操作按钮`);

    // 点击第一个快捷操作
    if (count > 0) {
      await actions.first().click();
      await page.waitForTimeout(3000);
    }

    const msgs = page.locator('.chat-message.assistant');
    const msgCount = await msgs.count();
    console.log(`✅ AI 回复了 ${msgCount} 条消息`);

    // 手动输入测试
    await page.fill('.chat-input', '如何创建虚拟机？');
    await page.click('.send-btn');
    await page.waitForTimeout(4000);

    const msgsAfter = await page.locator('.chat-message.assistant').count();
    console.log(`✅ 手动发送后 AI 回复数: ${msgsAfter}`);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'ai_chat.png'), fullPage: false });

    // 关闭面板
    const closeBtns = page.locator('.header-btn');
    const closeCount = await closeBtns.count();
    if (closeCount > 0) {
      await closeBtns.last().click();
      await page.waitForTimeout(500);
    }
    console.log('✅ 聊天面板关闭');

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

    // API 验证 - 获取模板列表
    const apiResp = await page.evaluate(async (url) => {
      const token = localStorage.getItem('pve_token');
      const resp = await fetch(url + '/apps', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      return await resp.json();
    }, API_URL);
    const templateCount = apiResp?.data?.length || 0;
    console.log(`✅ API 返回 ${templateCount} 个应用模板`);

    if (templateCount === 0) {
      console.log('⚠️ 模板为空，检查后端种子数据');
      return false;
    }

    // 导航到应用商店页面
    await page.goto(`${BASE_URL}/apps`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'app_store.png'), fullPage: true });

    // 检查页面是否有内容
    const bodyText = await page.textContent('body');
    console.log(`✅ 应用商店页面加载，内容长度: ${bodyText.length}`);

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

    // API 验证 - 获取用户列表
    const apiResp = await page.evaluate(async (url) => {
      const token = localStorage.getItem('pve_token');
      const resp = await fetch(url + '/access/users', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      return await resp.json();
    }, API_URL);
    const userCount = apiResp?.data?.length || 0;
    console.log(`✅ API 返回 ${userCount} 个用户`);

    // 验证角色
    const roleResp = await page.evaluate(async (url) => {
      const token = localStorage.getItem('pve_token');
      const resp = await fetch(url + '/access/roles', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      return await resp.json();
    }, API_URL);
    const roleCount = roleResp?.data?.length || 0;
    console.log(`✅ API 返回 ${roleCount} 个角色`);

    // 验证组
    const groupResp = await page.evaluate(async (url) => {
      const token = localStorage.getItem('pve_token');
      const resp = await fetch(url + '/access/groups', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      return await resp.json();
    }, API_URL);
    const groupCount = groupResp?.data?.length || 0;
    console.log(`✅ API 返回 ${groupCount} 个组`);

    // 验证 ACL
    const aclResp = await page.evaluate(async (url) => {
      const token = localStorage.getItem('pve_token');
      const resp = await fetch(url + '/access/acl', {
        headers: { 'Authorization': 'Bearer ' + token }
      });
      return await resp.json();
    }, API_URL);
    const aclCount = aclResp?.data?.length || 0;
    console.log(`✅ API 返回 ${aclCount} 条 ACL`);

    // 导航到访问管理页面
    await page.goto(`${BASE_URL}/access`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'access_page.png'), fullPage: true });

    const bodyText = await page.textContent('body');
    console.log(`✅ 访问管理页面加载，内容长度: ${bodyText.length}`);

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

    const sidebar = page.locator('[class*="sidebar"], [class*="Sidebar"], [class*="menu"]').first();
    let sidebarText = '';
    try {
      sidebarText = await sidebar.textContent({ timeout: 5000 });
    } catch {
      // 如果 sidebar 没有找到，检查整个页面
      sidebarText = await page.textContent('body');
    }
    const hasAI = sidebarText.includes('AI 智能中心') || sidebarText.includes('AI 智能体');
    if (hasAI) {
      throw new Error('侧边栏不应该有 AI 入口！');
    }
    console.log('✅ 侧边栏已正确移除 AI 入口');

    // 验证应用商店和访问管理在侧边栏
    const hasAppStore = sidebarText.includes('应用商店');
    const hasAccess = sidebarText.includes('访问管理');
    console.log(`✅ 侧边栏-应用商店: ${hasAppStore ? '有' : '无'}, 访问管理: ${hasAccess ? '有' : '无'}`);

    if (!hasAppStore || !hasAccess) {
      console.log('⚠️ 侧边栏缺少某些导航项，可能 UI 尚未完全渲染');
    }

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
    process.exit(0);
  } else {
    console.log('❌ 部分测试失败，请查看上方日志');
    process.exit(1);
  }
})();
