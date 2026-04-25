import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * PVE 认证状态存储
 * 管理用户登录状态、token、节点配置等信息
 */
export interface PveNodeConfig {
  /** 节点地址 */
  host: string
  /** 节点端口 */
  port: number
  /** 节点名称（用户自定义） */
  name: string
}

export interface UserInfo {
  /** 用户名 */
  username: string
  /** 认证域 */
  realm: string
  /** 权限列表 */
  permissions: string[]
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string>(localStorage.getItem('pve_token') || '')
  const currentNode = ref<PveNodeConfig | null>(null)
  const userInfo = ref<UserInfo | null>(null)
  const savedNodes = ref<PveNodeConfig[]>(
    (() => {
      try {
        const saved = localStorage.getItem('pve_nodes')
        return saved ? JSON.parse(saved) : []
      } catch {
        return []
      }
    })(),
  )

  // Getters
  const isLoggedIn = computed(() => !!token.value)
  const currentHost = computed(() => currentNode.value?.host || '')

  /**
   * 用户登录
   * @param data 登录凭据（用户名/密码 或 API Token）
   */
  async function login(data: {
    host: string
    port: number
    username: string
    password?: string
    apiToken?: string
  }): Promise<boolean> {
    try {
      // TODO: 实际调用后端认证 API
      // const res = await post('/auth/login', data)
      // token.value = res.token
      // currentUserNode.value = { host: data.host, port: data.port, name: data.host }

      // 开发阶段模拟登录
      token.value = 'mock_token_' + Date.now()
      localStorage.setItem('pve_token', token.value)
      currentNode.value = {
        host: data.host,
        port: data.port,
        name: data.host,
      }
      userInfo.value = {
        username: data.username,
        realm: 'pam',
        permissions: [],
      }
      return true
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  /**
   * 退出登录，清除所有本地状态
   */
  function logout() {
    token.value = ''
    currentNode.value = null
    userInfo.value = null
    localStorage.removeItem('pve_token')
  }

  /**
   * 保存节点配置到本地存储
   */
  function saveNode(node: PveNodeConfig) {
    const exists = savedNodes.value.findIndex((n) => n.host === node.host && n.port === node.port)
    if (exists === -1) {
      savedNodes.value.push(node)
    } else {
      savedNodes.value[exists] = node
    }
    localStorage.setItem('pve_nodes', JSON.stringify(savedNodes.value))
  }

  /**
   * 删除已保存的节点配置
   */
  function removeNode(host: string, port: number) {
    savedNodes.value = savedNodes.value.filter(
      (n) => !(n.host === host && n.port === port),
    )
    localStorage.setItem('pve_nodes', JSON.stringify(savedNodes.value))
  }

  return {
    token,
    currentNode,
    userInfo,
    savedNodes,
    isLoggedIn,
    currentHost,
    login,
    logout,
    saveNode,
    removeNode,
  }
})
