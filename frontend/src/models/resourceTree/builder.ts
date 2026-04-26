/**
 * 资源树构建器
 * 
 * 核心职责:
 * 1. 将扁平资源数据构建成树形结构
 * 2. 支持资源分组 (虚拟机、容器、存储、网络)
 * 3. 计算状态徽章和统计信息
 * 4. 提供搜索和过滤能力
 * 
 * 设计参考: 阿里云/腾讯云控制台资源导航模式
 */

import {
  ResourceTypeEnum,
  ResourceStatusEnum,
  ResourceTypeName,
  ResourceTypePluralName,
  AbstractTreeNode,
  ResourceGroupNode,
  LeafResourceNode,
  StatusBadge,
  QuickAction,
  TreeStatistics,
  SearchOptions,
  FilterOptions,
  TreeBuilderConfig,
  DEFAULT_TREE_CONFIG,
  SearchScope,
} from './types'

import type {
  PVENode,
  PVEVM,
  PVECT,
  PVEStorage,
  PVENetwork,
  ResourceStatus,
  ResourceType,
} from '@/api/types'

// ============================================================
// 资源树构建器类
// ============================================================

/**
 * 抽象资源树构建器
 * 
 * 使用方式:
 * ```ts
 * const builder = new ResourceTreeBuilder(config)
 * const tree = builder.buildTree(nodes, vms, containers, storages, networks)
 * const filtered = builder.filterTree(tree, { query: 'vm', types: ['vm'] })
 * const stats = builder.computeStatistics(tree)
 * ```
 */
export class ResourceTreeBuilder {
  private config: TreeBuilderConfig

  constructor(config?: Partial<TreeBuilderConfig>) {
    this.config = { ...DEFAULT_TREE_CONFIG, ...config }
  }

  /**
   * 构建完整的资源树
   * 
   * @param nodes 物理节点列表
   * @param vms 虚拟机列表
   * @param containers 容器列表
   * @param storages 存储列表
   * @param networks 网络列表
   * @returns 树形结构数组
   */
  buildTree(
    nodes: PVENode[],
    vms: PVEVM[],
    containers: PVECT[],
    storages: PVEStorage[],
    networks: PVENetwork[],
  ): AbstractTreeNode[] {
    const dcName = this.config.dataCenterName || '数据中心'

    // 构建数据中心根节点
    const dataCenterNode: AbstractTreeNode = {
      id: 'datacenter-root',
      name: dcName,
      type: ResourceTypeEnum.DataCenter,
      status: ResourceStatusEnum.Running,
      parentId: null,
      depth: 0,
      expandable: true,
      icon: 'Coin',
      children: nodes.map((node) => this.buildNodeTree(node, vms, containers, storages, networks)),
    }

    // 计算数据中心级别的统计徽章
    dataCenterNode.badges = this.computeDataCenterBadges(nodes, vms, containers, storages, networks)

    return [dataCenterNode]
  }

  /**
   * 构建单个节点及其子资源
   */
  private buildNodeTree(
    node: PVENode,
    vms: PVEVM[],
    containers: PVECT[],
    storages: PVEStorage[],
    networks: PVENetwork[],
  ): AbstractTreeNode {
    // 筛选属于该节点的资源
    const nodeVMs = vms.filter((vm) => vm.node === node.name)
    const nodeCTs = containers.filter((ct) => ct.node === node.name)
    const nodeStorages = storages.filter((s) => s.node === node.name)
    const nodeNetworks = networks.filter((n) => n.node === node.name)

    const children: AbstractTreeNode[] = []

    if (this.config.enableGrouping) {
      // 按资源类型分组模式
      children.push(
        ...this.createResourceGroups(node.name, nodeVMs, nodeCTs, nodeStorages, nodeNetworks),
      )
    } else {
      // 扁平模式 - 直接列出所有资源
      children.push(
        ...nodeVMs.map((vm) => this.createLeafNode(vm, 'vm', node.name)),
        ...nodeCTs.map((ct) => this.createLeafNode(ct, 'ct', node.name)),
        ...nodeStorages.map((s) => this.createLeafNode(s, 'storage', node.name)),
        ...nodeNetworks.map((n) => this.createLeafNode(n, 'network', node.name)),
      )
    }

    // 推断节点状态
    const nodeStatus = this.inferNodeStatus(node, nodeVMs, nodeCTs)

    return {
      id: node.id,
      name: node.name,
      type: this.mapResourceType(node.type),
      status: nodeStatus,
      parentId: 'datacenter-root',
      depth: 1,
      expandable: children.length > 0,
      icon: 'Server',
      children,
      badges: this.computeNodeBadges(nodeVMs, nodeCTs, nodeStorages, nodeNetworks),
      meta: {
        cpuUsage: node.cpuUsage,
        memoryUsage: node.memoryUsage,
        diskUsage: node.diskUsage,
        uptime: node.uptime,
        node: node.name,
      },
      actions: this.getNodeActions(nodeStatus),
    }
  }

  /**
   * 创建资源分组节点
   * 
   * 例如: 节点 -> 虚拟机(3) -> [VM1, VM2, VM3]
   */
  private createResourceGroups(
    nodeName: string,
    vms: PVEVM[],
    containers: PVECT[],
    storages: PVEStorage[],
    networks: PVENetwork[],
  ): AbstractTreeNode[] {
    const groups: AbstractTreeNode[] = []

    // 虚拟机组
    if (vms.length > 0 || this.config.includeEmptyGroups) {
      groups.push(this.createGroupNode(ResourceTypeEnum.VM, nodeName, vms))
    }

    // 容器组
    if (containers.length > 0 || this.config.includeEmptyGroups) {
      groups.push(this.createGroupNode(ResourceTypeEnum.Container, nodeName, containers))
    }

    // 存储组
    if (storages.length > 0 || this.config.includeEmptyGroups) {
      groups.push(this.createGroupNode(ResourceTypeEnum.Storage, nodeName, storages))
    }

    // 网络组
    if (networks.length > 0 || this.config.includeEmptyGroups) {
      groups.push(this.createGroupNode(ResourceTypeEnum.Network, nodeName, networks))
    }

    return groups
  }

  /**
   * 创建单个资源分组
   */
  private createGroupNode(
    type: ResourceTypeEnum,
    nodeName: string,
    resources: Array<PVEVM | PVECT | PVEStorage | PVENetwork>,
  ): ResourceGroupNode {
    const groupId = `group-${nodeName}-${type}`

    // 计算状态统计
    const statusCounts: Partial<Record<ResourceStatusEnum, number>> = {}
    for (const res of resources) {
      const status = this.mapStatus(res.status)
      statusCounts[status] = (statusCounts[status] || 0) + 1
    }

    // 构建子节点
    const children = resources.map((res) => this.createLeafNode(res, type, nodeName))

    // 计算分组状态 (有 running 则为 running, 全 stopped 则为 stopped)
    const groupStatus = this.computeGroupStatus(statusCounts)

    const groupNode: ResourceGroupNode = {
      id: groupId,
      name: ResourceTypePluralName[type],
      type,
      status: groupStatus,
      parentId: `node-${nodeName}`,
      depth: 2,
      expandable: resources.length > 0,
      totalCount: resources.length,
      statusCounts,
      children,
      icon: this.getGroupIcon(type),
      badges: this.computeGroupBadges(statusCounts),
    }

    return groupNode
  }

  /**
   * 创建叶子资源节点
   */
  private createLeafNode(
    resource: PVEVM | PVECT | PVEStorage | PVENetwork,
    type: ResourceType,
    nodeName: string,
  ): LeafResourceNode {
    const mappedType = this.mapResourceType(type)

    // 获取资源 ID
    let resourceId: number | string
    if ('vmid' in resource) resourceId = resource.vmid
    else if ('ctid' in resource) resourceId = resource.ctid
    else resourceId = resource.name

    return {
      id: resource.id,
      name: resource.name,
      type: mappedType,
      status: this.mapStatus(resource.status),
      parentId: `group-${nodeName}-${mappedType}`,
      depth: 3,
      expandable: false,
      resourceId,
      nodeName,
      icon: this.getResourceIcon(mappedType),
      meta: {
        cpuUsage: 'cpuUsage' in resource ? resource.cpuUsage : undefined,
        memoryUsage: 'memoryUsage' in resource ? undefined : undefined,
        uptime: 'uptime' in resource ? resource.uptime : undefined,
        node: nodeName,
        tags: 'tags' in resource && resource.tags ? resource.tags.split(',') : undefined,
      },
      actions: this.getResourceActions(mappedType, this.mapStatus(resource.status)),
    }
  }

  // ============================================================
  // 搜索与过滤
  // ============================================================

  /**
   * 搜索并过滤树节点
   * 
   * @param tree 原始树数据
   * @param searchOptions 搜索选项
   * @param filterOptions 过滤选项
   * @returns 过滤后的树数据
   */
  filterTree(
    tree: AbstractTreeNode[],
    searchOptions?: SearchOptions,
    filterOptions?: FilterOptions,
  ): AbstractTreeNode[] {
    let result = tree

    // 应用搜索
    if (searchOptions?.query) {
      result = this.searchTree(result, searchOptions)
    }

    // 应用过滤
    if (filterOptions) {
      result = this.applyFilters(result, filterOptions)
    }

    return result
  }

  /**
   * 在树中搜索匹配的节点
   */
  private searchTree(tree: AbstractTreeNode[], options: SearchOptions): AbstractTreeNode[] {
    const query = options.query.toLowerCase()
    const scope = options.scope || SearchScope.All

    return tree.reduce<AbstractTreeNode[]>((result, node) => {
      const matches = this.nodeMatchesSearch(node, query, scope)
      const filteredChildren = this.searchTree(node.children, options)

      // 自身匹配或有匹配的子节点则保留
      if (matches || filteredChildren.length > 0) {
        result.push({
          ...node,
          children: filteredChildren,
        })
      }

      return result
    }, [])
  }

  /**
   * 检查节点是否匹配搜索条件
   */
  private nodeMatchesSearch(
    node: AbstractTreeNode,
    query: string,
    scope: SearchScope,
  ): boolean {
    switch (scope) {
      case SearchScope.Name:
        return node.name.toLowerCase().includes(query)
      case SearchScope.NameAndTags:
        const nameMatch = node.name.toLowerCase().includes(query)
        const tagMatch = node.meta?.tags?.some((tag) => tag.toLowerCase().includes(query)) ?? false
        return nameMatch || tagMatch
      default:
        // All: 搜索名称、类型、状态等
        const typeName = ResourceTypeName[node.type]?.toLowerCase() ?? ''
        const statusName = node.status.toLowerCase()
        return (
          node.name.toLowerCase().includes(query) ||
          typeName.includes(query) ||
          statusName.includes(query)
        )
    }
  }

  /**
   * 应用过滤条件
   */
  private applyFilters(
    tree: AbstractTreeNode[],
    options: FilterOptions,
  ): AbstractTreeNode[] {
    return tree.reduce<AbstractTreeNode[]>((result, node) => {
      // 递归过滤子节点
      const filteredChildren = this.applyFilters(node.children, options)

      // 检查当前节点是否满足过滤条件
      const passesFilter = this.nodePassesFilter(node, options)

      // 自身满足或有满足的子节点则保留
      if (passesFilter || filteredChildren.length > 0) {
        result.push({
          ...node,
          children: filteredChildren,
        })
      }

      return result
    }, [])
  }

  /**
   * 检查节点是否通过过滤条件
   */
  private nodePassesFilter(node: AbstractTreeNode, options: FilterOptions): boolean {
    // 类型过滤
    if (options.types?.length && !options.types.includes(node.type)) {
      return false
    }

    // 状态过滤
    if (options.statuses?.length && !options.statuses.includes(node.status)) {
      return false
    }

    // 节点过滤
    if (options.nodes?.length) {
      const nodeMatch = node.meta?.node && options.nodes.includes(node.meta.node)
      const nameMatch = options.nodes.includes(node.name)
      if (!nodeMatch && !nameMatch) return false
    }

    // 隐藏已停止资源
    if (options.hideStopped && node.status === ResourceStatusEnum.Stopped) {
      return false
    }

    return true
  }

  // ============================================================
  // 统计信息
  // ============================================================

  /**
   * 计算资源树统计摘要
   */
  computeStatistics(tree: AbstractTreeNode[]): TreeStatistics {
    const stats: TreeStatistics = {
      totalResources: 0,
      byType: {},
      byStatus: {},
      nodeCount: 0,
      dataCenterCount: 0,
    }

    this.traverseTree(tree, (node) => {
      if (node.type === ResourceTypeEnum.DataCenter) {
        stats.dataCenterCount++
      } else if (node.type === ResourceTypeEnum.Node) {
        stats.nodeCount++
      } else if (node.depth >= 3) {
        // 叶子资源节点
        stats.totalResources++
        stats.byType[node.type] = (stats.byType[node.type] || 0) + 1
        stats.byStatus[node.status] = (stats.byStatus[node.status] || 0) + 1
      }
    })

    return stats
  }

  /**
   * 遍历树节点
   */
  private traverseTree(
    tree: AbstractTreeNode[] | undefined,
    callback: (node: AbstractTreeNode) => void,
  ): void {
    if (!tree || !Array.isArray(tree)) return
    for (const node of tree) {
      callback(node)
      this.traverseTree(node.children, callback)
    }
  }

  // ============================================================
  // 树遍历辅助方法
  // ============================================================

  /**
   * 查找指定 ID 的节点
   */
  findNodeById(tree: AbstractTreeNode[], nodeId: string): AbstractTreeNode | null {
    for (const node of tree) {
      if (node.id === nodeId) return node
      const found = this.findNodeById(node.children, nodeId)
      if (found) return found
    }
    return null
  }

  /**
   * 收集所有节点 ID
   */
  collectAllNodeIds(tree: AbstractTreeNode[]): string[] {
    const ids: string[] = []
    this.traverseTree(tree, (node) => ids.push(node.id))
    return ids
  }

  /**
   * 获取默认展开的节点 ID
   */
  getDefaultExpandedKeys(tree: AbstractTreeNode[]): string[] {
    const keys: string[] = []
    this.traverseTree(tree, (node) => {
      if (node.depth < this.config.defaultExpandDepth) {
        keys.push(node.id)
      }
    })
    return keys
  }

  // ============================================================
  // 状态与图标映射
  // ============================================================

  /** 映射 PVE 状态到统一状态枚举 */
  private mapStatus(status: ResourceStatus): ResourceStatusEnum {
    switch (status) {
      case 'running':
        return ResourceStatusEnum.Running
      case 'stopped':
        return ResourceStatusEnum.Stopped
      case 'paused':
        return ResourceStatusEnum.Paused
      case 'frozen':
        return ResourceStatusEnum.Frozen
      case 'error':
        return ResourceStatusEnum.Error
      case 'unknown':
        return ResourceStatusEnum.Unknown
      default:
        return ResourceStatusEnum.Unknown
    }
  }

  /** 映射旧类型到新枚举 */
  private mapResourceType(type: ResourceType): ResourceTypeEnum {
    const typeMap: Record<string, ResourceTypeEnum> = {
      datacenter: ResourceTypeEnum.DataCenter,
      node: ResourceTypeEnum.Node,
      vm: ResourceTypeEnum.VM,
      ct: ResourceTypeEnum.Container,
      storage: ResourceTypeEnum.Storage,
      network: ResourceTypeEnum.Network,
      pool: ResourceTypeEnum.Pool,
    }
    return typeMap[type] || ResourceTypeEnum.Node
  }

  /** 推断节点整体状态 */
  private inferNodeStatus(
    node: PVENode,
    vms: PVEVM[],
    containers: PVECT[],
  ): ResourceStatusEnum {
    // 有高 CPU 使用率表示错误
    if (node.cpuUsage !== undefined && node.cpuUsage > 95) {
      return ResourceStatusEnum.Error
    }

    // 有任何运行中的资源表示节点运行中
    const hasRunning = [
      ...vms,
      ...containers,
    ].some((res) => res.status === 'running')

    if (hasRunning || node.cpuUsage !== undefined || node.memoryUsage !== undefined) {
      return ResourceStatusEnum.Running
    }

    return ResourceStatusEnum.Unknown
  }

  /** 计算分组状态 */
  private computeGroupStatus(
    statusCounts: Partial<Record<ResourceStatusEnum, number>>,
  ): ResourceStatusEnum {
    if (statusCounts[ResourceStatusEnum.Running]) {
      return ResourceStatusEnum.Running
    }
    if (statusCounts[ResourceStatusEnum.Error]) {
      return ResourceStatusEnum.Error
    }
    if (statusCounts[ResourceStatusEnum.Paused] || statusCounts[ResourceStatusEnum.Frozen]) {
      return ResourceStatusEnum.Paused
    }
    return ResourceStatusEnum.Stopped
  }

  /** 获取分组图标 */
  private getGroupIcon(type: ResourceTypeEnum): string {
    const iconMap: Record<ResourceTypeEnum, string> = {
      [ResourceTypeEnum.DataCenter]: 'Coin',
      [ResourceTypeEnum.Node]: 'Server',
      [ResourceTypeEnum.VM]: 'Monitor',
      [ResourceTypeEnum.Container]: 'Box',
      [ResourceTypeEnum.Storage]: 'FolderOpened',
      [ResourceTypeEnum.Network]: 'Connection',
      [ResourceTypeEnum.Pool]: 'Collection',
    }
    return iconMap[type] || 'Document'
  }

  /** 获取资源图标 */
  private getResourceIcon(type: ResourceTypeEnum): string {
    return this.getGroupIcon(type)
  }

  // ============================================================
  // 徽章计算
  // ============================================================

  /** 计算数据中心级别徽章 */
  private computeDataCenterBadges(
    nodes: PVENode[],
    vms: PVEVM[],
    containers: PVECT[],
    storages: PVEStorage[],
    networks: PVENetwork[],
  ): StatusBadge[] {
    const totalCount = vms.length + containers.length + storages.length + networks.length

    const runningCount = [
      ...vms.filter((vm) => vm.status === 'running'),
      ...containers.filter((ct) => ct.status === 'running'),
    ].length

    return [
      { status: ResourceStatusEnum.Running, count: totalCount },
    ]
  }

  /** 计算节点级别徽章 */
  private computeNodeBadges(
    vms: PVEVM[],
    containers: PVECT[],
    storages: PVEStorage[],
    networks: PVENetwork[],
  ): StatusBadge[] {
    const running = [
      ...vms.filter((vm) => vm.status === 'running'),
      ...containers.filter((ct) => ct.status === 'running'),
    ].length

    const stopped = [
      ...vms.filter((vm) => vm.status === 'stopped'),
      ...containers.filter((ct) => ct.status === 'stopped'),
    ].length

    const badges: StatusBadge[] = []
    if (running > 0) badges.push({ status: ResourceStatusEnum.Running, count: running })
    if (stopped > 0) badges.push({ status: ResourceStatusEnum.Stopped, count: stopped })

    return badges
  }

  /** 计算分组级别徽章 */
  private computeGroupBadges(
    statusCounts: Partial<Record<ResourceStatusEnum, number>>,
  ): StatusBadge[] {
    return Object.entries(statusCounts).map(([status, count]) => ({
      status: status as ResourceStatusEnum,
      count,
    }))
  }

  // ============================================================
  // 快捷操作
  // ============================================================

  /** 获取节点级快捷操作 */
  private getNodeActions(status: ResourceStatusEnum): QuickAction[] {
    const actions: QuickAction[] = []

    if (status === ResourceStatusEnum.Running) {
      actions.push({
        id: 'node-console',
        label: '控制台',
        icon: 'Monitor',
        enabled: true,
        type: 'primary',
      })
    }

    actions.push(
      {
        id: 'node-refresh',
        label: '刷新',
        icon: 'Refresh',
        enabled: true,
        type: 'info',
      },
      {
        id: 'node-summary',
        label: '概览',
        icon: 'DataAnalysis',
        enabled: true,
        type: 'info',
      },
    )

    return actions
  }

  /** 获取资源级快捷操作 */
  private getResourceActions(
    type: ResourceTypeEnum,
    status: ResourceStatusEnum,
  ): QuickAction[] {
    const actions: QuickAction[] = []

    switch (type) {
      case ResourceTypeEnum.VM:
      case ResourceTypeEnum.Container:
        if (status === ResourceStatusEnum.Running) {
          actions.push({
            id: 'resource-console',
            label: '控制台',
            icon: 'Monitor',
            enabled: true,
            type: 'primary',
          })
        }
        if (status === ResourceStatusEnum.Stopped) {
          actions.push({
            id: 'resource-start',
            label: '启动',
            icon: 'CaretRight',
            enabled: true,
            type: 'success',
          })
        }
        if (status === ResourceStatusEnum.Running) {
          actions.push({
            id: 'resource-stop',
            label: '停止',
            icon: 'VideoPause',
            enabled: true,
            type: 'danger',
          })
        }
        actions.push({
          id: 'resource-reboot',
          label: '重启',
          icon: 'RefreshLeft',
          enabled: status === ResourceStatusEnum.Running,
          type: 'warning',
        })
        break

      case ResourceTypeEnum.Storage:
        actions.push({
          id: 'storage-browse',
          label: '浏览',
          icon: 'Files',
          enabled: true,
          type: 'info',
        })
        break

      case ResourceTypeEnum.Network:
        actions.push({
          id: 'network-config',
          label: '配置',
          icon: 'Setting',
          enabled: true,
          type: 'info',
        })
        break
    }

    return actions
  }
}
