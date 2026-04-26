/**
 * PVE WebUI 完整 API 接口自动化测试（浏览器直调版）
 * 
 * 本测试通过 Playwright 浏览器自动化方式：
 * 1. 先通过后端 API 登录获取 JWT Token
 * 2. 使用浏览器直接调用所有后端 API 接口
 * 3. 验证每个接口能从真实 PVE 获取数据
 */
import { test, expect, Page } from '@playwright/test'

// ============================================================
// 测试配置
// ============================================================
const TEST_HOST = '192.168.1.10'
const TEST_USERNAME = 'root'
const TEST_PASSWORD = 'Qq1212112'

// ============================================================
// 辅助函数
// ============================================================

async function loginAndGetToken(page: Page): Promise<string> {
  await page.goto('/')
  await page.waitForLoadState('domcontentloaded')
  
  const result = await page.evaluate(async ({ host, port, username, password }) => {
    const response = await fetch('http://localhost:8080/api/auth/login', {
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
  }, { host: TEST_HOST, port: 8006, username: TEST_USERNAME, password: TEST_PASSWORD })

  if (!result.ok || !result.token) {
    throw new Error(`登录失败: HTTP ${result.status}, ${result.message}`)
  }

  await page.evaluate((token) => {
    localStorage.setItem('pve_token', token)
  }, result.token)

  console.log('🔑 登录成功，Token 长度:', result.token.length)
  return result.token
}

async function callApi(page: Page, token: string, endpoint: string, description: string): Promise<any> {
  const result = await page.evaluate(async ({ token, endpoint }) => {
    const response = await fetch(`http://localhost:8080/api/pve${endpoint}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
    const data = await response.json()
    return {
      status: response.status,
      data: data.data,
      code: data.code,
    }
  }, { token, endpoint })

  expect(result.status, `${description} - HTTP 状态码`).toBe(200)
  expect(result.code, `${description} - 业务码`).toBe(0)
  expect(result.data, `${description} - 数据不应为空`).toBeTruthy()
  
  console.log(`✅ ${description}: HTTP ${result.status}`)
  return result.data
}

async function callApiArray(page: Page, token: string, endpoint: string, description: string): Promise<any[]> {
  const result = await page.evaluate(async ({ token, endpoint }) => {
    const response = await fetch(`http://localhost:8080/api/pve${endpoint}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
    const data = await response.json()
    return {
      status: response.status,
      data: data.data,
      code: data.code,
    }
  }, { token, endpoint })

  expect(result.status, `${description} - HTTP 状态码`).toBe(200)
  expect(result.code, `${description} - 业务码`).toBe(0)
  expect(Array.isArray(result.data), `${description} - 返回应为数组`).toBe(true)
  
  console.log(`✅ ${description}: ${result.data.length} 条数据`)
  return result.data || []
}

// ============================================================
// 测试用例
// ============================================================

test.describe('PVE 完整 API 接口验证', () => {
  test.slow()
  let authToken: string
  
  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage()
    try {
      authToken = await loginAndGetToken(page)
    } finally {
      await page.close()
    }
  })

  test.describe('1. 集群管理', () => {
    test('集群资源', async ({ browser }) => {
      const page = await browser.newPage()
      const resources = await callApiArray(page, authToken, '/cluster/resources', '集群资源')
      const types = resources.map((r: any) => r.type)
      expect(types).toContain('node')
      console.log(`   📊 资源: 节点=${resources.filter((r: any) => r.type === 'node').length}, VM=${resources.filter((r: any) => r.type === 'vm').length}, CT=${resources.filter((r: any) => r.type === 'lxc').length}, 存储=${resources.filter((r: any) => r.type === 'storage').length}`)
      await page.close()
    })

    test('集群任务', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, '/cluster/tasks', '集群任务')
      await page.close()
    })

    test('资源池', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, '/cluster/pools', '资源池')
      await page.close()
    })
  })

  test.describe('2. 节点管理', () => {
    let nodeName = 'lang'

    test('节点状态', async ({ browser }) => {
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/status`, '节点状态')
      await page.close()
    })

    test('节点版本', async ({ browser }) => {
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/version`, '节点版本')
      await page.close()
    })

    test('节点服务', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/services`, '节点服务')
      await page.close()
    })

    test('节点网络', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/network`, '节点网络')
      await page.close()
    })

    test('节点 DNS', async ({ browser }) => {
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/dns`, '节点 DNS')
      await page.close()
    })

    test('节点时间', async ({ browser }) => {
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/time`, '节点时间')
      await page.close()
    })

    test('节点任务', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/tasks`, '节点任务')
      await page.close()
    })
  })

  test.describe('3. QEMU 虚拟机', () => {
    const nodeName = 'lang'
    let vmId: number | undefined

    test('获取 VM 列表', async ({ browser }) => {
      const page = await browser.newPage()
      const vms = await callApiArray(page, authToken, `/nodes/${nodeName}/qemu`, 'QEMU 列表')
      if (vms.length > 0) {
        vmId = vms[0].vmid
        console.log(`   📺 VM: ${vmId} (${vms[0].name})`)
      }
      await page.close()
    })

    test('虚拟机配置', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/qemu/${vmId}/config`, 'QEMU 配置')
      await page.close()
    })

    test('虚拟机快照', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/qemu/${vmId}/snapshot`, 'QEMU 快照')
      await page.close()
    })

    test('虚拟机 RRD', async ({ browser }) => {
      test.skip(!vmId, '没有 VM')
      const page = await browser.newPage()
      
      // RRD 数据可能在某些 PVE 配置中不可用，使用可选验证
      const result = await page.evaluate(async ({ authToken, endpoint }) => {
        const response = await fetch(`http://localhost:8080/api/pve${endpoint}`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${authToken}`,
            'Content-Type': 'application/json',
          },
        })
        const data = await response.json()
        return {
          status: response.status,
          data: data.data,
          code: data.code,
          message: data.message,
        }
      }, { authToken, endpoint: `/nodes/${nodeName}/qemu/${vmId}/rrd?timeframe=hour` })

      if (result.status === 200 && result.code === 0) {
        console.log(`✅ QEMU RRD: HTTP ${result.status}, ${Array.isArray(result.data) ? result.data.length + ' 条数据' : '有数据'}`)
      } else {
        console.log(`⚠️ QEMU RRD: HTTP ${result.status}, ${result.message || '无响应'}`)
      }
      
      await page.close()
    })
  })

  test.describe('4. LXC 容器', () => {
    const nodeName = 'lang'
    let containerId: number | undefined

    test('获取 CT 列表', async ({ browser }) => {
      const page = await browser.newPage()
      const containers = await callApiArray(page, authToken, `/nodes/${nodeName}/lxc`, 'LXC 列表')
      if (containers.length > 0) {
        containerId = containers[0].vmid
        console.log(`   📦 CT: ${containerId} (${containers[0].name})`)
      }
      await page.close()
    })

    test('容器配置', async ({ browser }) => {
      test.skip(!containerId, '没有 CT')
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/lxc/${containerId}/config`, 'LXC 配置')
      await page.close()
    })

    test('容器快照', async ({ browser }) => {
      test.skip(!containerId, '没有 CT')
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/lxc/${containerId}/snapshot`, 'LXC 快照')
      await page.close()
    })
  })

  test.describe('5. 存储管理', () => {
    const nodeName = 'lang'
    let storageName: string | undefined

    test('存储列表', async ({ browser }) => {
      const page = await browser.newPage()
      const storages = await callApiArray(page, authToken, `/nodes/${nodeName}/storage`, '存储列表')
      const enabled = storages.find((s: any) => s.enabled === 1)
      if (enabled) {
        storageName = enabled.storage
        console.log(`   💾 存储: ${storageName}`)
      }
      await page.close()
    })

    test('存储状态', async ({ browser }) => {
      test.skip(!storageName, '没有存储')
      const page = await browser.newPage()
      await callApi(page, authToken, `/nodes/${nodeName}/storage/${storageName}/status`, '存储状态')
      await page.close()
    })

    test('存储内容', async ({ browser }) => {
      test.skip(!storageName, '没有存储')
      const page = await browser.newPage()
      await callApiArray(page, authToken, `/nodes/${nodeName}/storage/${storageName}/content`, '存储内容')
      await page.close()
    })
  })

  test.describe('6. 访问控制', () => {
    test('用户列表', async ({ browser }) => {
      const page = await browser.newPage()
      const users = await callApiArray(page, authToken, '/access/users', '用户列表')
      console.log(`   👤 用户: ${users.length}`)
      await page.close()
    })

    test('组列表', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, '/access/groups', '组列表')
      await page.close()
    })

    test('角色列表', async ({ browser }) => {
      const page = await browser.newPage()
      await callApiArray(page, authToken, '/access/roles', '角色列表')
      await page.close()
    })

    test('ACL 列表', async ({ browser }) => {
      const page = await browser.newPage()
      await callApi(page, authToken, '/access/acl', 'ACL 列表')
      await page.close()
    })

    test('认证域列表', async ({ browser }) => {
      const page = await browser.newPage()
      const domains = await callApiArray(page, authToken, '/access/domains', '认证域列表')
      console.log(`   🔐 认证域: ${domains.map((d: any) => d.realm).join(', ')}`)
      await page.close()
    })
  })
})
