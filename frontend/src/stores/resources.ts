import { defineStore } from 'pinia'
import { ref, computed, shallowRef } from 'vue'
import { getClusterResources } from '@/api/cluster'
import type {
  PVENode,
  PVEVM,
  PVECT,
  PVEStorage,
  PVENetwork,
  ResourceStatus,
  ResourceType,
  ResourceListResponse,
  ClusterResource,
} from '@/api/types'
import {
  ResourceTreeBuilder,
  ResourceTypeEnum,
  ResourceStatusEnum,
  AbstractTreeNode,
  TreeStatistics,
  SearchOptions,
  FilterOptions,
  SearchScope,
} from '@/models/resourceTree'

/** 轮询间隔 (30 秒) */
const POLL_INTERVAL = 30_000

/**
 * PVE 资源状态管理
 *
 * 职责:
 * 1. 从 PVE API 获取节点/虚拟机/容器/存储/网络资源
 * 2. 使用抽象资源树构建器构建分组树结构
 * 3. 支持 30 秒定时轮询刷新
 * 4. 提供搜索、过滤、统计等能力
 */
export const useResourceStore = defineStore('resources', () => {
  // ============================================================
  // State
  // ============================================================

  /** 节点列表 */
  const nodes = shallowRef<PVENode[]>([])
  /** 虚拟机列表 */
  const vms = shallowRef<PVEVM[]>([])
  /** 容器列表 */
  const containers = shallowRef<PVECT[]>([])
  /** 存储列表 */
  const storages = shallowRef<PVEStorage[]>([])
  /** 网络列表 */
  const networks = shallowRef<PVENetwork[]>([])

  /** 树形展开的节点 ID 列表 */
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

  /** 资源树构建器实例 */
  const treeBuilder = new ResourceTreeBuilder()

  // ============================================================
  // Getters
  // ============================================================

  /**
   * 构建完整的树形数据结构 (使用抽象模型)
   * 新层级: 数据中心 -> 节点 -> 资源分组 -> 具体资源
   */
  const resourceTree = computed<AbstractTreeNode[]>(() => {
    const tree = treeBuilder.buildTree(
      nodes.value,
      vms.value,
      containers.value,
      storages.value,
      networks.value,
    )

    // 应用搜索和过滤
    const searchOpts = buildSearchOptions(searchQuery.value)
    return treeBuilder.filterTree(tree, searchOpts)
  })

  /**
   * 当前选中的资源节点
   */
  const selectedNode = computed<AbstractTreeNode | null>(() => {
    if (!selectedNodeId.value) return null
    return treeBuilder.findNodeById(resourceTree.value, selectedNodeId.value)
  })

  /**
   * 资源统计摘要
   */
  const statistics = computed<TreeStatistics>(() => {
    return treeBuilder.computeStatistics(resourceTree.value)
  })

  /**
   * 默认展开的节点 ID
   */
  const defaultExpandedKeys = computed<string[]>(() => {
    return treeBuilder.getDefaultExpandedKeys(resourceTree.value)
  })

  // ============================================================
  // Actions
  // ============================================================

  /**
   * 从 PVE API 获取所有资源数据
   * 通过 /cluster/resources 接口获取集群资源，然后按类型分类
   */
  async function fetchResources(): Promise<void> {
    loading.value = true
    try {
      const resources = await getClusterResources()
      const parsed = parseClusterResources(resources)
      nodes.value = parsed.nodes
      vms.value = parsed.vms
      containers.value = parsed.containers
      storages.value = parsed.storages
      networks.value = parsed.networks

      // 首次加载时设置默认展开
      if (lastRefreshedAt.value === null) {
        expandedKeys.value = defaultExpandedKeys.value
      }

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
    expandedKeys.value = treeBuilder.collectAllNodeIds(resourceTree.value)
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

  /**
   * 设置展开的节点
   */
  function setExpandedKeys(keys: string[]): void {
    expandedKeys.value = keys
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
    statistics,
    defaultExpandedKeys,
    // Actions
    fetchResources,
    startPolling,
    stopPolling,
    setSearchQuery,
    selectNode,
    expandAll,
    collapseAll,
    refresh,
    setExpandedKeys,
  }
})

// ============================================================
// 内部工具函数
// ============================================================

/**
 * 构建搜索选项
 */
function buildSearchOptions(query: string): SearchOptions | undefined {
  if (!query.trim()) return undefined
  return {
    query: query.trim(),
    fuzzy: true,
    scope: SearchScope.All,
  }
}

/**
 * 解析集群资源数据，按类型分类
 */
function parseClusterResources(resources: ClusterResource[]): ResourceListResponse {
  const nodes: PVENode[] = []
  const vms: PVEVM[] = []
  const containers: PVECT[] = []
  const storages: PVEStorage[] = []
  const networks: PVENetwork[] = []

  for (const res of resources) {
    switch (res.type) {
      case 'node':
        nodes.push({
          id: `node-${res.node || res.id}`,
          name: res.node || res.id,
          type: 'node' as ResourceType,
          status: normalizeNodeStatus(res.status),
          cpuUsage: res.cpu ? Math.round(res.cpu * 1000) / 10 : undefined,
          memoryUsage: res.maxmem ? Math.round(((res.mem || 0) / res.maxmem) * 1000) / 10 : undefined,
          uptime: res.uptime || 0,
          version: '',
        })
        break
      case 'vm':
        vms.push({
          id: `vm-${res.vmid}`,
          name: res.name || `VM ${res.vmid}`,
          type: 'vm' as ResourceType,
          status: normalizeVMStatus(res.status),
          vmid: res.vmid || 0,
          cpus: res.maxcpu || 1,
          memoryMB: res.maxmem ? Math.round(res.maxmem / (1024 * 1024)) : 0,
          cpuUsage: res.cpu ? Math.round(res.cpu * 1000) / 10 : undefined,
          uptime: res.uptime || 0,
          node: res.node || '',
        })
        break
      case 'lxc':
        containers.push({
          id: `ct-${res.vmid}`,
          name: res.name || `CT ${res.vmid}`,
          type: 'ct' as ResourceType,
          status: normalizeVMStatus(res.status),
          ctid: res.vmid || 0,
          cpus: res.maxcpu || 1,
          memoryMB: res.maxmem ? Math.round(res.maxmem / (1024 * 1024)) : 0,
          cpuUsage: res.cpu ? Math.round(res.cpu * 1000) / 10 : undefined,
          uptime: res.uptime || 0,
          node: res.node || '',
        })
        break
      case 'storage':
        storages.push({
          id: `storage-${res.storage || res.id}`,
          name: res.storage || res.id,
          type: 'storage' as ResourceType,
          status: (res.shared ? 'running' : 'running') as ResourceStatus,
          storageType: 'dir',
          total: res.maxdisk || 0,
          used: res.disk || 0,
          available: (res.maxdisk || 0) - (res.disk || 0),
          usage: res.maxdisk ? Math.round(((res.disk || 0) / res.maxdisk) * 100) : 0,
          active: true,
          node: res.node || '',
        })
        break
    }
  }

  return { nodes, vms, containers, storages, networks }
}

/**
 * 规范化节点状态
 */
function normalizeNodeStatus(status?: string): ResourceStatus {
  switch (status) {
    case 'online':
      return 'running'
    case 'offline':
      return 'stopped'
    case 'unknown':
      return 'unknown'
    default:
      return 'running'
  }
}

/**
 * 规范化 VM/CT 状态
 */
function normalizeVMStatus(status?: string): ResourceStatus {
  switch (status) {
    case 'running':
      return 'running'
    case 'stopped':
      return 'stopped'
    case 'paused':
      return 'paused'
    default:
      return 'unknown'
  }
}
