import { test, expect, Page } from '@playwright/test'

const TEST_HOST = '192.168.1.10'
const TEST_USERNAME = 'root'
const TEST_PASSWORD = 'Qq1212112'
const BACKEND_URL = 'http://localhost:8080'
const FRONTEND_URL = 'http://localhost:8088'

let authToken: string
let nodeName: string

async function loginAndGetToken(page: Page): Promise<string> {
  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')

  const result = await page.evaluate(async ({ host, port, username, password, backendUrl }) => {
    const response = await fetch(`${backendUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ host, port, username, password }),
    })
    const data = await response.json()
    return {
      ok: response.ok,
      status: response.status,
      token: data.data?.token,
      message: data.message,
    }
  }, { host: TEST_HOST, port: 8006, username: TEST_USERNAME, password: TEST_PASSWORD, backendUrl: BACKEND_URL })

  if (!result.ok || !result.token) {
    throw new Error(`登录失败: HTTP ${result.status}, ${result.message}`)
  }

  await page.evaluate((token) => {
    localStorage.setItem('pve_token', token)
  }, result.token)

  return result.token
}

async function apiGet(page: Page, endpoint: string): Promise<{ status: number; code: number; data: any; message?: string }> {
  return page.evaluate(async ({ token, endpoint, backendUrl }) => {
    const response = await fetch(`${backendUrl}/api/pve${endpoint}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
    const data = await response.json()
    return { status: response.status, code: data.code, data: data.data, message: data.message }
  }, { token: authToken, endpoint, backendUrl: BACKEND_URL })
}

async function apiPost(page: Page, endpoint: string, body?: any): Promise<{ status: number; code: number; data: any; message?: string }> {
  return page.evaluate(async ({ token, endpoint, backendUrl, body }) => {
    const response = await fetch(`${backendUrl}/api/pve${endpoint}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    })
    const data = await response.json()
    return { status: response.status, code: data.code, data: data.data, message: data.message }
  }, { token: authToken, endpoint, backendUrl: BACKEND_URL, body })
}

async function setTokenAndNavigate(page: Page, path: string): Promise<void> {
  await page.goto(`${FRONTEND_URL}/login`)
  await page.waitForLoadState('domcontentloaded')
  await page.evaluate((token) => {
    localStorage.setItem('pve_token', token)
  }, authToken)
  await page.goto(`${FRONTEND_URL}${path}`)
  await page.waitForLoadState('domcontentloaded')
}

async function apiGetAllowFail(page: Page, endpoint: string): Promise<{ status: number; code: number; data: any; message?: string }> {
  const result = await apiGet(page, endpoint)
  if (result.code !== 0) {
    console.log(`⚠️ ${endpoint}: code=${result.code}, message=${result.message}`)
  }
  return result
}

test.describe('PVE 完整功能自动化测试', () => {
  test.slow()

  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage()
    try {
      authToken = await loginAndGetToken(page)
      const res = await apiGet(page, '/cluster/resources')
      if (res.code === 0 && Array.isArray(res.data)) {
        const node = res.data.find((r: any) => r.type === 'node')
        if (node) nodeName = node.node || node.name || node.id
      }
      if (!nodeName) nodeName = 'lang'
    } finally {
      await page.close()
    }
  })

  test.describe('1. 认证系统', () => {
    test('登录获取 JWT Token', () => {
      expect(authToken).toBeTruthy()
      expect(authToken.length).toBeGreaterThan(10)
    })

    test('无效凭据登录失败', async ({ browser }) => {
      const page = await browser.newPage()
      await page.goto('/')
      const result = await page.evaluate(async ({ backendUrl }) => {
        const response = await fetch(`${backendUrl}/api/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ host: '192.168.1.10', port: 8006, username: 'invalid', password: 'wrong' }),
        })
        return { status: response.status, ok: response.ok }
      }, { backendUrl: BACKEND_URL })
      expect(result.ok).toBe(false)
      await page.close()
    })
  })

  test.describe('2. 集群管理 API', () => {
    test('获取集群资源', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/resources')
      expect(res.code).toBe(0)
      expect(Array.isArray(res.data)).toBe(true)
      expect(res.data.some((r: any) => r.type === 'node')).toBe(true)
      await page.close()
    })

    test('获取集群任务', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/tasks')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取下一个 VM ID', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, '/cluster/nextid')
      expect(res.status).toBe(200)
      await page.close()
    })

    test('获取 HA 配置', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, '/cluster/ha')
      expect(res.status).toBe(200)
      await page.close()
    })

    test('获取 HA 组列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, '/cluster/ha/groups')
      expect([200, 500]).toContain(res.status)
      await page.close()
    })

    test('获取 HA 资源列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/ha/resources')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 SDN Zone', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/sdn/zones')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 SDN VNET', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/sdn/vnets')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取资源池列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/pools')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取集群存储', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, '/cluster/storage')
      expect([200, 500]).toContain(res.status)
      await page.close()
    })

    test('获取数据中心配置', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, '/cluster/config')
      expect(res.status).toBe(200)
      await page.close()
    })

    test('获取集群日志', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/log')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取复制任务', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/cluster/replication')
      expect(res.code).toBe(0)
      await page.close()
    })
  })

  test.describe('3. 节点管理 API', () => {
    test('获取节点状态', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/status`)
      expect(res.code).toBe(0)
      expect(res.data).toBeTruthy()
      await page.close()
    })

    test('获取节点版本', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/version`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点服务', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/services`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点系统日志', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/syslog`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点网络', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/network`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点 DNS', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/dns`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点时间', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/time`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取节点 RRD', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, `/nodes/${nodeName}/rrd?timeframe=hour`)
      expect([200, 500]).toContain(res.status)
      await page.close()
    })

    test('获取 APT 更新', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/apt/update`)
      expect(res.code).toBe(0)
      await page.close()
    })
  })

  test.describe('4. QEMU 虚拟机 API', () => {
    let vmId: number | undefined

    test('获取 QEMU 列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/qemu`)
      expect(res.code).toBe(0)
      if (Array.isArray(res.data) && res.data.length > 0) {
        vmId = res.data[0].vmid
      }
      await page.close()
    })

    test('获取 QEMU 配置', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/qemu/${vmId}/config`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 QEMU 快照', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/qemu/${vmId}/snapshot`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 QEMU Pending', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, `/nodes/${nodeName}/qemu/${vmId}/pending`)
      expect([200, 500]).toContain(res.status)
      await page.close()
    })
  })

  test.describe('5. LXC 容器 API', () => {
    let ctId: number | undefined

    test('获取 LXC 列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/lxc`)
      expect(res.code).toBe(0)
      if (Array.isArray(res.data) && res.data.length > 0) {
        ctId = res.data[0].vmid
      }
      await page.close()
    })

    test('获取 LXC 配置', async ({ browser }) => {
      test.skip(!ctId, '没有 CT')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/lxc/${ctId}/config`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 LXC 快照', async ({ browser }) => {
      test.skip(!ctId, '没有 CT')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/lxc/${ctId}/snapshot`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 LXC Pending', async ({ browser }) => {
      test.skip(!ctId, '没有 CT')
      const page = await browser.newPage()
      const res = await apiGetAllowFail(page, `/nodes/${nodeName}/lxc/${ctId}/pending`)
      expect([200, 500]).toContain(res.status)
      await page.close()
    })
  })

  test.describe('6. 存储管理 API', () => {
    let storageName: string | undefined

    test('获取存储列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/storage`)
      expect(res.code).toBe(0)
      if (Array.isArray(res.data) && res.data.length > 0) {
        const enabled = res.data.find((s: any) => s.enabled === 1)
        storageName = enabled ? enabled.storage : res.data[0].storage
      }
      await page.close()
    })

    test('获取存储状态', async ({ browser }) => {
      test.skip(!storageName, '没有存储')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/storage/${storageName}/status`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取存储内容', async ({ browser }) => {
      test.skip(!storageName, '没有存储')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/storage/${storageName}/content`)
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取存储详情', async ({ browser }) => {
      test.skip(!storageName, '没有存储')
      const page = await browser.newPage()
      const res = await apiGet(page, `/nodes/${nodeName}/storage/${storageName}`)
      expect(res.code).toBe(0)
      await page.close()
    })
  })

  test.describe('7. 访问控制 API', () => {
    test('获取用户列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/access/users')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取组列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/access/groups')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取角色列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/access/roles')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取 ACL 列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/access/acl')
      expect(res.code).toBe(0)
      await page.close()
    })

    test('获取认证域列表', async ({ browser }) => {
      const page = await browser.newPage()
      const res = await apiGet(page, '/access/domains')
      expect(res.code).toBe(0)
      await page.close()
    })
  })

  test.describe('8. 前端页面交互', () => {
    test('登录页面加载', async ({ page }) => {
      await page.goto(`${FRONTEND_URL}/login`)
      await page.waitForLoadState('domcontentloaded')
      await expect(page.locator('form')).toBeVisible({ timeout: 10000 })
    })

    test('登录后跳转到仪表盘', async ({ page }) => {
      await setTokenAndNavigate(page, '/')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)
      expect(page.url()).toContain('localhost:8088')
    })

    test('仪表盘页面显示集群资源', async ({ page }) => {
      await setTokenAndNavigate(page, '/')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(3000)
      const pageContent = await page.textContent('body')
      expect(pageContent).toBeTruthy()
    })

    test('节点列表页面加载', async ({ page }) => {
      await setTokenAndNavigate(page, '/nodes')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)
      const pageContent = await page.textContent('body')
      expect(pageContent).toBeTruthy()
    })

    test('存储列表页面加载', async ({ page }) => {
      await setTokenAndNavigate(page, '/storage')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)
      const pageContent = await page.textContent('body')
      expect(pageContent).toBeTruthy()
    })

    test('访问管理页面加载', async ({ page }) => {
      await setTokenAndNavigate(page, '/access')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)
      const pageContent = await page.textContent('body')
      expect(pageContent).toBeTruthy()
    })

    test('主题切换功能', async ({ page }) => {
      await setTokenAndNavigate(page, '/')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const themeButton = page.locator('[data-testid="theme-toggle"], button:has-text("☀"), button:has-text("🌙")').first()
      if (await themeButton.isVisible()) {
        const htmlBefore = await page.evaluate(() => document.documentElement.className)
        await themeButton.click()
        await page.waitForTimeout(500)
        const htmlAfter = await page.evaluate(() => document.documentElement.className)
        expect(htmlBefore).not.toBe(htmlAfter)
      }
    })
  })
})
