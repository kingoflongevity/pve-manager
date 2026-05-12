const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');

const SCREENSHOTS_DIR = path.join(__dirname, 'screenshots');
fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });

async function login(page) {
  await page.goto('http://localhost:8088/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);

  await page.fill('input[placeholder*="192.168"], input[placeholder*="pve."]', '192.168.1.100');
  await page.waitForTimeout(300);
  let nextBtn = page.locator('button').filter({ hasText: '下一步' }).first();
  await nextBtn.waitFor({ state: 'visible', timeout: 5000 });
  await nextBtn.click();
  await page.waitForTimeout(800);

  await page.fill('input[placeholder*="例: root"]', 'root');
  await page.waitForTimeout(300);
  await page.fill('input[type="password"]', 'password');
  await page.waitForTimeout(300);
  nextBtn = page.locator('button').filter({ hasText: '下一步' }).first();
  await nextBtn.waitFor({ state: 'visible', timeout: 5000 });
  await nextBtn.click();
  await page.waitForTimeout(2000);

  let loginBtn = page.locator('button').filter({ hasText: '登录系统' });
  try { await loginBtn.waitFor({ state: 'visible', timeout: 8000 }); } catch { await page.waitForTimeout(2000); }
  if (await loginBtn.isEnabled()) { await loginBtn.click(); }

  await page.waitForTimeout(3000);
  await page.waitForLoadState('networkidle');
  console.log('URL:', page.url());
}

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  // 监听 console
  page.on('console', msg => {
    if (msg.type() === 'error') console.log('CONSOLE ERROR:', msg.text());
    if (msg.type() === 'log') console.log('CONSOLE LOG:', msg.text().substring(0, 200));
  });

  await login(page);

  // 检查 body 内容
  const body = await page.textContent('body');
  console.log('BODY start:', body.substring(0, 500));

  // 检查所有按钮和链接
  const links = page.locator('a, li');
  const count = await links.count();
  console.log(`Links/li count: ${count}`);
  for (let i = 0; i < Math.min(count, 20); i++) {
    const text = await links.nth(i).textContent();
    if (text.trim()) console.log(`  [${i}]: "${text.trim().substring(0, 50)}"`);
  }

  // 检查 FloatingPet
  const pet = page.locator('.floating-pet-container, .pet-button');
  const petCount = await pet.count();
  console.log(`FloatingPet elements: ${petCount}`);

  await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'debug_page.png'), fullPage: true });
  console.log('Screenshot saved');

  await browser.close();
})();
