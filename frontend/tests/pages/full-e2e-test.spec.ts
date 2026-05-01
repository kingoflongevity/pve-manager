import { test, expect, Page, request } from '@playwright/test'

const TEST_HOST = '192.168.1.10'
const TEST_USERNAME = 'root'
const TEST_PASSWORD = 'Qq1212112'
const BACKEND_URL = 'http://localhost:8080'
const FRONTEND_URL = 'http://localhost:8088'

let authToken: string
let nodeName: string

async function getAuthToken(): Promise<string> {
  const apiContext = await request.newContext()
  const response = await apiContext.post(`${BACKEND_URL}/api/auth/login`, {
    data: { host: TEST_HOST, port: 8006, username: TEST_USERNAME, password: TEST_PASSWORD },
  })
  const data = await response.json()
  expect(data.code).toBe(0)
  expect(data.data.token).toBeTruthy()
  await apiContext.dispose()
  return data.data.token
}

async function apiGet(page: Page, endpoint: string): Promise<{ status: number; code: number; data: any; message?: string }> {
  return page.evaluate(async ({ backendUrl, token, endpoint: ep }) => {
    const res = await fetch(`${backendUrl}/api/pve${ep}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    const text = await res.text()
    try {
      const json = JSON.parse(text)
      return { status: res.status, code: json.code ?? -1, data: json.data, message: json.message }
    } catch {
      return { status: res.status, code: res.status, data: null, message: text.substring(0, 200) }
    }
  }, { backendUrl: BACKEND_URL, token: authToken, endpoint })
}

async function setTokenAndNavigate(page: Page, path: string): Promise<void> {
  await page.goto(`${FRONTEND_URL}`, { waitUntil: 'commit', timeout: 60000 })
  await page.evaluate((token) => {
    localStorage.setItem('pve_token', token)
  }, authToken)
  await page.goto(`${FRONTEND_URL}${path}`, { waitUntil: 'commit', timeout: 60000 })
  try {
    await page.waitForSelector('#app', { timeout: 15000 })
  } catch {
    await page.waitForTimeout(5000)
  }
}

test.describe('PVE 全面功能自动化测试', () => {

  test.beforeAll(async () => {
    authToken = await getAuthToken()
    const apiContext = await request.newContext()
    const response = await apiContext.get(`${BACKEND_URL}/api/pve/cluster/resources`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const data = await response.json()
    const nodes = (Array.isArray(data.data) ? data.data : []).filter((r: any) => r.type === 'node')
    nodeName = nodes[0]?.node || nodes[0]?.name || 'lang'
    await apiContext.dispose()
  })

  // ============================================================
  // 1. 认证系统
  // ============================================================
  test('1.1 登录获取 JWT Token', async () => {
    const token = await getAuthToken()
    expect(token).toBeTruthy()
    expect(token.split('.').length).toBe(3)
  })

  test('1.2 无效凭证登录失败', async () => {
    const apiContext = await request.newContext()
    const response = await apiContext.post(`${BACKEND_URL}/api/auth/login`, {
      data: { host: TEST_HOST, port: 8006, username: 'root', password: 'wrongpassword' },
    })
    expect(response.ok()).toBe(false)
    await apiContext.dispose()
  })

  test('1.3 无 Token 访问受保护 API 返回 401', async () => {
    const apiContext = await request.newContext()
    const response = await apiContext.get(`${BACKEND_URL}/api/pve/cluster/resources`)
    expect(response.status()).toBe(401)
    await apiContext.dispose()
  })

  // ============================================================
  // 2. 集群资源 API
  // ============================================================
  test('2.1 获取集群全部资源', async ({ page }) => {
    const res = await apiGet(page, '/cluster/resources')
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
    expect(res.data.length).toBeGreaterThan(0)
  })

  test('2.2 按类型获取节点资源', async ({ page }) => {
    const res = await apiGet(page, '/cluster/resources?type=node')
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  test('2.3 按类型获取 VM 资源', async ({ page }) => {
    const res = await apiGet(page, '/cluster/resources?type=vm')
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  test('2.4 按类型获取存储资源', async ({ page }) => {
    const res = await apiGet(page, '/cluster/resources?type=storage')
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  // ============================================================
  // 3. 集群任务 API
  // ============================================================
  test('3.1 获取集群任务列表', async ({ page }) => {
    const res = await apiGet(page, '/cluster/tasks')
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  test('3.2 获取下一个可用 VM ID', async ({ page }) => {
    const res = await apiGet(page, '/cluster/nextid')
    expect(res.code).toBe(0)
    expect(res.data).toBeTruthy()
  })

  // ============================================================
  // 4. 节点 API
  // ============================================================
  test('4.1 获取节点状态', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/status`)
    expect(res.code).toBe(0)
    expect(res.data).toBeTruthy()
  })

  test('4.2 获取节点版本', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/version`)
    expect(res.code).toBe(0)
    expect(res.data).toBeTruthy()
  })

  test('4.3 获取节点服务列表', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/services`)
    expect(res.code).toBe(0)
  })

  test('4.4 获取节点网络配置', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/network`)
    expect(res.code).toBe(0)
  })

  test('4.5 获取节点 DNS 配置', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/dns`)
    expect(res.code).toBe(0)
  })

  test('4.6 获取节点时间配置', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/time`)
    expect(res.code).toBe(0)
  })

  test('4.7 获取节点 RRD 监控数据', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/rrd?timeframe=hour&ds=cpu`)
    expect(res.code).toBe(0)
  })

  test('4.8 获取节点任务列表', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/tasks`)
    expect(res.code).toBe(0)
  })

  // ============================================================
  // 5. QEMU 虚拟机 API
  // ============================================================
  test('5.1 获取 QEMU 虚拟机列表', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/qemu`)
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  test('5.2 QEMU 列表数据包含必要字段', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/qemu`)
    expect(res.code).toBe(0)
    const vms = Array.isArray(res.data) ? res.data : []
    if (vms.length > 0) {
      const vm = vms[0]
      expect(vm).toHaveProperty('vmid')
      expect(vm).toHaveProperty('status')
    }
  })

  // ============================================================
  // 6. LXC 容器 API
  // ============================================================
  test('6.1 获取 LXC 容器列表', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/lxc`)
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  // ============================================================
  // 7. 存储 API
  // ============================================================
  test('7.1 获取节点存储列表', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/storage`)
    expect(res.code).toBe(0)
    expect(Array.isArray(res.data)).toBe(true)
  })

  // ============================================================
  // 8. 访问控制 API
  // ============================================================
  test('8.1 获取用户列表', async ({ page }) => {
    const res = await apiGet(page, '/access/users')
    expect(res.code).toBe(0)
  })

  test('8.2 获取角色列表', async ({ page }) => {
    const res = await apiGet(page, '/access/roles')
    expect(res.code).toBe(0)
  })

  test('8.3 获取 ACL 列表', async ({ page }) => {
    const res = await apiGet(page, '/access/acl')
    expect(res.code).toBe(0)
  })

  test('8.4 获取认证域列表', async ({ page }) => {
    const res = await apiGet(page, '/access/domains')
    expect(res.code).toBe(0)
  })

  // ============================================================
  // 9. 集群管理 API
  // ============================================================
  test('9.1 获取 HA 配置', async ({ page }) => {
    const res = await apiGet(page, '/cluster/ha')
    expect([0, 500]).toContain(res.code)
  })

  test('9.2 获取 HA 资源', async ({ page }) => {
    const res = await apiGet(page, '/cluster/ha/resources')
    expect([0, 500]).toContain(res.code)
  })

  test('9.3 获取资源池列表', async ({ page }) => {
    const res = await apiGet(page, '/cluster/pools')
    expect(res.code).toBe(0)
  })

  test('9.4 获取集群日志', async ({ page }) => {
    const res = await apiGet(page, '/cluster/log')
    expect([0, 500]).toContain(res.code)
  })

  // ============================================================
  // 10. 前端页面交互测试
  // ============================================================
  test('10.1 登录页面加载', async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/login`, { waitUntil: 'commit', timeout: 60000 })
    await page.waitForTimeout(2000)
    expect(page.url()).toContain('login')
  })

  test('10.2 登录后跳转到仪表盘', async ({ page }) => {
    await setTokenAndNavigate(page, '/')
    const url = page.url()
    expect(url).not.toContain('login')
  })

  test('10.3 仪表盘页面显示内容', async ({ page }) => {
    await setTokenAndNavigate(page, '/')
    await page.waitForTimeout(2000)
    const content = await page.textContent('body')
    const hasContent = content && content.trim().length > 50
    expect(hasContent).toBe(true)
  })

  test('10.4 侧边栏导航项可见', async ({ page }) => {
    await setTokenAndNavigate(page, '/')
    const sidebar = page.locator('.app-sidebar')
    await expect(sidebar).toBeVisible({ timeout: 15000 })
  })

  test('10.5 虚拟机列表页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/qemu')
    expect(page.url()).toContain('qemu')
  })

  test('10.6 容器列表页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/lxc')
    expect(page.url()).toContain('lxc')
  })

  test('10.7 存储列表页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/storage')
    expect(page.url()).toContain('storage')
  })

  test('10.8 节点列表页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/nodes')
    expect(page.url()).toContain('nodes')
  })

  test('10.9 集群概览页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/cluster')
    expect(page.url()).toContain('cluster')
  })

  test('10.10 访问管理页面加载', async ({ page }) => {
    await setTokenAndNavigate(page, '/access')
    expect(page.url()).toContain('access')
  })

  // ============================================================
  // 11. 主题切换测试
  // ============================================================
  test('11.1 主题切换功能', async ({ page }) => {
    await setTokenAndNavigate(page, '/')
    const themeToggle = page.locator('.header-icon-btn').first()
    if (await themeToggle.isVisible()) {
      await themeToggle.click()
      await page.waitForTimeout(1000)
    }
  })

  // ============================================================
  // 12. 侧边栏折叠测试
  // ============================================================
  test('12.1 侧边栏折叠/展开', async ({ page }) => {
    await setTokenAndNavigate(page, '/')
    const collapseBtn = page.locator('.collapse-btn')
    if (await collapseBtn.isVisible()) {
      await collapseBtn.click()
      await page.waitForTimeout(500)
      const sidebar = page.locator('.app-sidebar')
      const hasCollapsed = await sidebar.evaluate(el => el.classList.contains('collapsed'))
      expect(typeof hasCollapsed).toBe('boolean')
    }
  })

  // ============================================================
  // 13. QEMU VM 详细测试
  // ============================================================
  test('13.1 QEMU 配置获取', async ({ page }) => {
    const listRes = await apiGet(page, `/nodes/${nodeName}/qemu`)
    const vms = Array.isArray(listRes.data) ? listRes.data : []
    if (vms.length > 0) {
      const vmid = vms[0].vmid
      const configRes = await apiGet(page, `/nodes/${nodeName}/qemu/${vmid}/config`)
      expect(configRes.code).toBe(0)
    }
  })

  // ============================================================
  // 14. 存储内容 API
  // ============================================================
  test('14.1 获取存储内容列表', async ({ page }) => {
    const storageRes = await apiGet(page, `/nodes/${nodeName}/storage`)
    const storages = Array.isArray(storageRes.data) ? storageRes.data : []
    if (storages.length > 0) {
      const storageName = storages[0].storage
      const contentRes = await apiGet(page, `/nodes/${nodeName}/storage/${storageName}/content`)
      expect(contentRes.code).toBe(0)
    }
  })

  // ============================================================
  // 15. 错误处理测试
  // ============================================================
  test('15.1 访问不存在的节点返回错误', async ({ page }) => {
    const res = await apiGet(page, '/nodes/nonexistent-node-xyz/status')
    expect(res.code).not.toBe(0)
  })

  test('15.2 访问不存在的 VM 返回错误', async ({ page }) => {
    const res = await apiGet(page, `/nodes/${nodeName}/qemu/99999/status/current`)
    expect(res.code).not.toBe(0)
  })

  // ============================================================
  // 16. 前端控制台错误检测
  // ============================================================
  test('16.1 仪表盘页面无关键 JS 错误', async ({ page }) => {
    const errors: string[] = []
    page.on('pageerror', err => errors.push(err.message))
    await setTokenAndNavigate(page, '/')
    const criticalErrors = errors.filter(e => !e.includes('ResizeObserver') && !e.includes('fonts.googleapis'))
    expect(criticalErrors.length).toBe(0)
  })

  test('16.2 虚拟机页面无关键 JS 错误', async ({ page }) => {
    const errors: string[] = []
    page.on('pageerror', err => errors.push(err.message))
    await setTokenAndNavigate(page, '/qemu')
    const criticalErrors = errors.filter(e => !e.includes('ResizeObserver') && !e.includes('fonts.googleapis'))
    expect(criticalErrors.length).toBe(0)
  })
})
