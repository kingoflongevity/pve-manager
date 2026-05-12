const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });
  await page.goto('http://localhost:8088/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
  const html = await page.content();
  console.log(html.substring(0, 3000));
  await page.screenshot({ path: 'scripts/screenshots/login_debug.png', fullPage: true });
  await browser.close();
})();
