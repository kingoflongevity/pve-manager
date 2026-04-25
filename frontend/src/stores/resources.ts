import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  PVENode,
  PVEVM,
  PVECT,
  PVEStorage,
  PVENetwork,
  TreeResourceData,
  ResourceStatus,
  ResourceType,
  ResourceListResponse,
} from '@/types/resources'

/** 轮询间隔 (30 秒) */
const POLL_INTERVAL = 30_000

/**
 * PVE 资源状态管理
 *
 * 职责:
 * 1. 从 PVE API 获取节点/虚拟机/容器/存储/网络资源
 * 2. 构建树形数据结构供 Tree 组件消费
 * 3. 支持 30 秒定时轮询刷新
 * 4. 暴露 resourceTree、selectedNode、expandedKeys 等计算属性
 */
export const useResourceStore = defineStore('resources', () => {
  // ============================================================
  // State
  // ============================================================

  /** 节点列表 */
  const nodes = ref<PVENode[]>([])
  /** 虚拟机列表 */
  const vms = ref<PVEVM[]>([])
  /** 容器列表 */
  const containers = ref<PVECT[]>([])
  /** 存储列表 */
  const storages = ref<PVEStorage[]>([])
  /** 网络列表 */
  const networks = ref<PVENetwork[]>([])

  /** 树形展开的节点 ID 集合 */
  const expandedKeys = ref<string[]>(['datacenter-root'])
  /** 当前选中的树节点 ID */
  const selectedNodeId = ref<string | null>(null)
  /** 搜索关键词 */
  const searchQuery = ref('')

  /** 加载状态 */
  const loading = ref(false)
  /** 最后刷新时间 */
  const lastRefreshedAt = ref<Date | null>(null)
  /** 轮询定时器 */
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // ============================================================
  // Getters
  // ============================================================

  /**
   * 构建完整的树形数据结构
   * 层级: 数据中心 -> 节点 -> (虚拟机/容器/存储/网络)
   */
  const resourceTree = computed<TreeResourceData[]>(() => {
    const dcChildren = nodes.value.map((node) =>
      buildNodeTree(node, vms.value, containers.value, storages.value, networks.value),
    )

    const tree: TreeResourceData[] = [
      {
        id: 'datacenter-root',
        name: '数据中心',
        type: 'datacenter' as ResourceType,
        status: 'running' as ResourceStatus,
        icon: 'DataCenter',
        children: dcChildren,
      },
    ]

    // 如果有搜索关键词，过滤树节点
    if (searchQuery.value.trim()) {
      return filterTree(tree, searchQuery.value.trim().toLowerCase())
    }

    return tree
  })

  /**
   * 当前选中的资源节点
   */
  const selectedNode = computed<TreeResourceData | null>(() => {
    if (!selectedNodeId.value) return null
    return findNodeInTree(resourceTree.value, selectedNodeId.value)
  })

  // ============================================================
  // Actions
  // ============================================================

  /**
   * 从 PVE API 获取所有资源数据
   * 实际项目中替换为真实的 API 调用
   */
  async function fetchResources(): Promise<void> {
    loading.value = true
    try {
      // TODO: 替换为真实 API
      // const res = await get<ResourceListResponse>('/cluster/resources')
      // nodes.value = res.nodes
      // vms.value = res.vms
      // containers.value = res.containers
      // storages.value = res.storages
      // networks.value = res.networks

      // 模拟数据（开发阶段）
      const mockData = getMockResources()
      nodes.value = mockData.nodes
      vms.value = mockData.vms
      containers.value = mockData.containers
      storages.value = mockData.storages
      networks.value = mockData.networks

      lastRefreshedAt.value = new Date()
    } catch (error) {
      console.error('获取资源数据失败:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 启动定时轮询刷新
   */
  function startPolling(): void {
    stopPolling()
    pollTimer = setInterval(() => {
      fetchResources()
    }, POLL_INTERVAL)
  }

  /**
   * 停止定时轮询
   */
  function stopPolling(): void {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  /**
   * 设置搜索关键词
   */
  function setSearchQuery(query: string): void {
    searchQuery.value = query
  }

  /**
   * 选中树节点
   */
  function selectNode(nodeId: string): void {
    selectedNodeId.value = nodeId
  }

  /**
   * 展开所有树节点
   */
  function expandAll(): void {
    expandedKeys.value = collectAllNodeIds(resourceTree.value)
  }

  /**
   * 收起所有树节点
   */
  function collapseAll(): void {
    expandedKeys.value = []
  }

  /**
   * 刷新资源数据（手动触发）
   */
  async function refresh(): Promise<void> {
    await fetchResources()
  }

  return {
    // State
    nodes,
    vms,
    containers,
    storages,
    networks,
    expandedKeys,
    selectedNodeId,
    searchQuery,
    loading,
    lastRefreshedAt,
    // Getters
    resourceTree,
    selectedNode,
    // Actions
    fetchResources,
    startPolling,
    stopPolling,
    setSearchQuery,
    selectNode,
    expandAll,
    collapseAll,
    refresh,
  }
})

// ============================================================
// 内部工具函数（纯函数，通过参数接收数据）
// ============================================================

/**
 * 将节点及其子资源构建为树形结构
 */
function buildNodeTree(
  node: PVENode,
  vms: PVEVM[],
  containers: PVECT[],
  storages: PVEStorage[],
  networks: PVENetwork[],
): TreeResourceData {
  const status = getNodeStatus(node)

  const children: TreeResourceData[] = [
    ...node.children || [],
    ...vms
      .filter((vm) => vm.node === node.name)
      .map((vm) => mapVMToTreeNode(vm)),
    ...containers
      .filter((ct) => ct.node === node.name)
      .map((ct) => mapCTToTreeNode(ct)),
    ...storages
      .filter((s) => s.node === node.name)
      .map((s) => mapStorageToTreeNode(s)),
    ...networks
      .filter((n) => n.node === node.name)
      .map((n) => mapNetworkToTreeNode(n)),
  ]

  return {
    id: `node-${node.name}`,
    name: node.name,
    type: node.type,
    status,
    icon: 'Server',
    children: children.length > 0 ? children : undefined,
  }
}

/**
 * 根据节点资源使用情况推断状态
 */
function getNodeStatus(node: PVENode): ResourceStatus {
  if (node.cpuUsage !== undefined && node.cpuUsage > 95) {
    return 'error' as ResourceStatus
  }
  if (node.cpuUsage !== undefined || node.memoryUsage !== undefined) {
    return 'running' as ResourceStatus
  }
  return 'unknown' as ResourceStatus
}

/** 将虚拟机映射为树节点 */
function mapVMToTreeNode(vm: PVEVM): TreeResourceData {
  return {
    id: `vm-${vm.vmid}`,
    name: vm.name || `VM ${vm.vmid}`,
    type: vm.type,
    status: vm.status,
    icon: 'Monitor',
  }
}

/** 将容器映射为树节点 */
function mapCTToTreeNode(ct: PVECT): TreeResourceData {
  return {
    id: `ct-${ct.ctid}`,
    name: ct.name || `CT ${ct.ctid}`,
    type: ct.type,
    status: ct.status,
    icon: 'Box',
  }
}

/** 将存储映射为树节点 */
function mapStorageToTreeNode(storage: PVEStorage): TreeResourceData {
  return {
    id: `storage-${storage.name}`,
    name: storage.name,
    type: storage.type,
    status: storage.active ? 'running' as ResourceStatus : 'stopped' as ResourceStatus,
    icon: 'FolderOpened',
  }
}

/** 将网络映射为树节点 */
function mapNetworkToTreeNode(network: PVENetwork): TreeResourceData {
  return {
    id: `network-${network.name}`,
    name: network.name,
    type: network.type,
    status: network.active ? 'running' as ResourceStatus : 'stopped' as ResourceStatus,
    icon: 'Connection',
  }
}

/**
 * 递归过滤树节点：匹配名称或子节点中有匹配项的节点保留
 */
function filterTree(
  tree: TreeResourceData[],
  query: string,
): TreeResourceData[] {
  return tree.reduce<TreeResourceData[]>((result, node) => {
    const nameMatch = node.name.toLowerCase().includes(query)
    const filteredChildren = node.children
      ? filterTree(node.children, query)
      : undefined

    // 自身匹配或有匹配的子节点则保留
    if (nameMatch || (filteredChildren && filteredChildren.length > 0)) {
      result.push({
        ...node,
        children: filteredChildren,
      })
    }

    return result
  }, [])
}

/** 在树中查找指定 ID 的节点 */
function findNodeInTree(
  tree: TreeResourceData[],
  nodeId: string,
): TreeResourceData | null {
  for (const node of tree) {
    if (node.id === nodeId) {
      return node
    }
    if (node.children) {
      const found = findNodeInTree(node.children, nodeId)
      if (found) return found
    }
  }
  return null
}

/** 收集树中所有节点的 ID */
function collectAllNodeIds(tree: TreeResourceData[]): string[] {
  return tree.reduce<string[]>((ids, node) => {
    ids.push(node.id)
    if (node.children) {
      ids.push(...collectAllNodeIds(node.children))
    }
    return ids
  }, [])
}

/** 获取模拟资源数据（开发阶段使用） */
function getMockResources(): ResourceListResponse {
  return {
    nodes: [
      {
        id: 'node-pve-01',
        name: 'pve-node-01',
        type: 'node' as ResourceType,
        status: 'running' as ResourceStatus,
        cpuUsage: 23.5,
        memoryUsage: 45.2,
        diskUsage: 62.0,
        uptime: 864000,
        version: '8.1.4',
        children: [],
      },
      {
        id: 'node-pve-02',
        name: 'pve-node-02',
        type: 'node' as ResourceType,
        status: 'running' as ResourceStatus,
        cpuUsage: 78.3,
        memoryUsage: 82.1,
        diskUsage: 45.0,
        uptime: 432000,
        version: '8.1.4',
        children: [],
      },
    ],
    vms: [
      {
        id: 'vm-100',
        name: 'web-server-01',
        type: 'vm' as ResourceType,
        status: 'running' as ResourceStatus,
        vmid: 100,
        cpus: 4,
        memoryMB: 8192,
        cpuUsage: 15.3,
        diskUsage: 32.0,
        uptime: 86400,
        node: 'pve-node-01',
      },
      {
        id: 'vm-101',
        name: 'db-server-01',
        type: 'vm' as ResourceType,
        status: 'running' as ResourceStatus,
        vmid: 101,
        cpus: 8,
        memoryMB: 16384,
        cpuUsage: 45.2,
        diskUsage: 67.0,
        uptime: 172800,
        node: 'pve-node-01',
      },
      {
        id: 'vm-102',
        name: 'test-vm-01',
        type: 'vm' as ResourceType,
        status: 'stopped' as ResourceStatus,
        vmid: 102,
        cpus: 2,
        memoryMB: 4096,
        node: 'pve-node-02',
      },
      {
        id: 'vm-103',
        name: 'api-gateway',
        type: 'vm' as ResourceType,
        status: 'error' as ResourceStatus,
        vmid: 103,
        cpus: 4,
        memoryMB: 8192,
        node: 'pve-node-02',
      },
    ],
    containers: [
      {
        id: 'ct-200',
        name: 'redis-cache',
        type: 'ct' as ResourceType,
        status: 'running' as ResourceStatus,
        ctid: 200,
        cpus: 2,
        memoryMB: 2048,
        cpuUsage: 5.1,
        diskUsage: 12.0,
        uptime: 259200,
        node: 'pve-node-01',
      },
      {
        id: 'ct-201',
        name: 'nginx-proxy',
        type: 'ct' as ResourceType,
        status: 'running' as ResourceStatus,
        ctid: 201,
        cpus: 1,
        memoryMB: 512,
        cpuUsage: 2.3,
        diskUsage: 8.0,
        uptime: 259200,
        node: 'pve-node-02',
      },
    ],
    storages: [
      {
        id: 'storage-local',
        name: 'local',
        type: 'storage' as ResourceType,
        status: 'running' as ResourceStatus,
        storageType: 'dir',
        total: 500_000_000_000,
        used: 310_000_000_000,
        available: 190_000_000_000,
        usage: 62,
        active: true,
        node: 'pve-node-01',
      },
      {
        id: 'storage-nfs-01',
        name: 'nfs-backup',
        type: 'storage' as ResourceType,
        status: 'running' as ResourceStatus,
        storageType: 'nfs',
        total: 2_000_000_000_000,
        used: 800_000_000_000,
        available: 1_200_000_000_000,
        usage: 40,
        active: true,
        node: 'pve-node-01',
      },
      {
        id: 'storage-local-02',
        name: 'local',
        type: 'storage' as ResourceType,
        status: 'running' as ResourceStatus,
        storageType: 'dir',
        total: 1_000_000_000_000,
        used: 450_000_000_000,
        available: 550_000_000_000,
        usage: 45,
        active: true,
        node: 'pve-node-02',
      },
    ],
    networks: [
      {
        id: 'network-vmbr0',
        name: 'vmbr0',
        type: 'network' as ResourceType,
        status: 'running' as ResourceStatus,
        interfaceType: 'bridge',
        address: '192.168.1.100',
        netmask: '255.255.255.0',
        gateway: '192.168.1.1',
        active: true,
        node: 'pve-node-01',
      },
      {
        id: 'network-eth0',
        name: 'eth0',
        type: 'network' as ResourceType,
        status: 'running' as ResourceStatus,
        interfaceType: 'eth',
        active: true,
        node: 'pve-node-01',
      },
    ],
  }
}
