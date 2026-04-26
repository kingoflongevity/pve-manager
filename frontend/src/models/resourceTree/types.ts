/**
 * PVE 抽象资源树模型 - 类型定义
 * 
 * 设计参考: 阿里云/腾讯云控制台资源导航模式
 * 树结构: 数据中心 -> 节点 -> 资源分组 -> 具体资源
 * 
 * 特性:
 * - 按资源类型分组 (虚拟机、容器、存储、网络)
 * - 状态徽章与数量统计
 * - 快捷操作支持
 * - 可复用的抽象数据模型
 */

// ============================================================
// 核心枚举与常量
// ============================================================

/** 资源类型枚举 */
export enum ResourceTypeEnum {
  /** 数据中心 */
  DataCenter = 'datacenter',
  /** 物理节点 */
  Node = 'node',
  /** 虚拟机 (QEMU) */
  VM = 'vm',
  /** 容器 (LXC) */
  Container = 'container',
  /** 存储 */
  Storage = 'storage',
  /** 网络 */
  Network = 'network',
  /** 资源池 */
  Pool = 'pool',
}

/** 资源类型中文映射 */
export const ResourceTypeName: Record<ResourceTypeEnum, string> = {
  [ResourceTypeEnum.DataCenter]: '数据中心',
  [ResourceTypeEnum.Node]: '节点',
  [ResourceTypeEnum.VM]: '虚拟机',
  [ResourceTypeEnum.Container]: '容器',
  [ResourceTypeEnum.Storage]: '存储',
  [ResourceTypeEnum.Network]: '网络',
  [ResourceTypeEnum.Pool]: '资源池',
}

/** 资源类型复数名称 (用于分组显示) */
export const ResourceTypePluralName: Record<ResourceTypeEnum, string> = {
  [ResourceTypeEnum.DataCenter]: '数据中心',
  [ResourceTypeEnum.Node]: '节点',
  [ResourceTypeEnum.VM]: '虚拟机',
  [ResourceTypeEnum.Container]: '容器',
  [ResourceTypeEnum.Storage]: '存储',
  [ResourceTypeEnum.Network]: '网络',
  [ResourceTypeEnum.Pool]: '资源池',
}

/** 资源状态枚举 */
export enum ResourceStatusEnum {
  /** 运行中 */
  Running = 'running',
  /** 已停止 */
  Stopped = 'stopped',
  /** 已暂停 */
  Paused = 'paused',
  /** 冻结 (LXC) */
  Frozen = 'frozen',
  /** 错误 */
  Error = 'error',
  /** 未知 */
  Unknown = 'unknown',
  /** 迁移中 */
  Migrating = 'migrating',
  /** 离线 */
  Offline = 'offline',
}

/** 资源状态中文映射 */
export const ResourceStatusName: Record<ResourceStatusEnum, string> = {
  [ResourceStatusEnum.Running]: '运行中',
  [ResourceStatusEnum.Stopped]: '已停止',
  [ResourceStatusEnum.Paused]: '已暂停',
  [ResourceStatusEnum.Frozen]: '已冻结',
  [ResourceStatusEnum.Error]: '错误',
  [ResourceStatusEnum.Unknown]: '未知',
  [ResourceStatusEnum.Migrating]: '迁移中',
  [ResourceStatusEnum.Offline]: '离线',
}

// ============================================================
// 抽象树节点接口
// ============================================================

/** 状态徽章信息 */
export interface StatusBadge {
  /** 状态类型 */
  status: ResourceStatusEnum
  /** 数量 */
  count: number
}

/** 快捷操作定义 */
export interface QuickAction {
  /** 操作唯一标识 */
  id: string
  /** 操作名称 */
  label: string
  /** 操作图标 (Element Plus icon name) */
  icon: string
  /** 是否可用 */
  enabled: boolean
  /** 操作类型 */
  type: 'primary' | 'success' | 'warning' | 'danger' | 'info'
}

/** 资源元数据 */
export interface ResourceMeta {
  /** CPU 使用率 (0-100) */
  cpuUsage?: number
  /** 内存使用率 (0-100) */
  memoryUsage?: number
  /** 磁盘使用率 (0-100) */
  diskUsage?: number
  /** 运行时间 (秒) */
  uptime?: number
  /** 所属节点 */
  node?: string
  /** 标签 */
  tags?: string[]
  /** 扩展数据 */
  [key: string]: unknown
}

/**
 * 抽象资源树节点 - 核心接口
 * 
 * 所有资源节点都实现此接口, 确保树结构的统一性和可扩展性
 */
export interface AbstractTreeNode {
  /** 节点唯一标识 */
  id: string
  /** 节点名称 */
  name: string
  /** 资源类型 */
  type: ResourceTypeEnum
  /** 资源状态 */
  status: ResourceStatusEnum
  /** 父节点 ID (根节点为 null) */
  parentId: string | null
  /** 子节点列表 */
  children: AbstractTreeNode[]
  /** 层级深度 (数据中心=0, 节点=1, 分组=2, 资源=3) */
  depth: number
  /** 是否可展开 */
  expandable: boolean
  /** 是否禁用选择 */
  disabled?: boolean
  
  // 以下字段用于 UI 渲染
  /** 状态徽章列表 */
  badges?: StatusBadge[]
  /** 快捷操作列表 */
  actions?: QuickAction[]
  /** 资源元数据 */
  meta?: ResourceMeta
  /** 图标名称 */
  icon?: string
  /** 显示标签 (可选, 默认使用 name) */
  displayLabel?: string
  /** 描述信息 */
  description?: string
}

// ============================================================
// 资源分组相关接口
// ============================================================

/**
 * 资源分组节点
 * 
 * 用于在节点下按类型分组显示资源
 * 例如: 节点 -> [虚拟机(3), 容器(2), 存储(1), 网络(2)]
 */
export interface ResourceGroupNode extends AbstractTreeNode {
  type: ResourceTypeEnum
  /** 组内资源总数 */
  totalCount: number
  /** 各状态资源数量统计 */
  statusCounts: Partial<Record<ResourceStatusEnum, number>>
}

/**
 * 叶子资源节点
 * 
 * 具体的资源实体 (VM/CT/Storage/Network)
 */
export interface LeafResourceNode extends AbstractTreeNode {
  /** 资源 ID (PVE vmid 等) */
  resourceId: number | string
  /** 所属节点名称 */
  nodeName: string
}

// ============================================================
// 树构建配置接口
// ============================================================

/** 树构建器配置 */
export interface TreeBuilderConfig {
  /** 数据中心显示名称 */
  dataCenterName?: string
  /** 是否启用资源分组 */
  enableGrouping: boolean
  /** 需要分组的资源类型 */
  groupByTypes: ResourceTypeEnum[]
  /** 是否包含空分组 */
  includeEmptyGroups: boolean
  /** 默认展开层级 */
  defaultExpandDepth: number
}

/** 默认树构建配置 */
export const DEFAULT_TREE_CONFIG: TreeBuilderConfig = {
  dataCenterName: '数据中心',
  enableGrouping: true,
  groupByTypes: [
    ResourceTypeEnum.VM,
    ResourceTypeEnum.Container,
    ResourceTypeEnum.Storage,
    ResourceTypeEnum.Network,
  ],
  includeEmptyGroups: false,
  defaultExpandDepth: 1,
}

// ============================================================
// 搜索与过滤相关接口
// ============================================================

/** 搜索选项 */
export interface SearchOptions {
  /** 搜索关键词 */
  query: string
  /** 是否模糊匹配 */
  fuzzy?: boolean
  /** 搜索范围 */
  scope?: SearchScope
}

/** 搜索范围 */
export enum SearchScope {
  /** 搜索所有字段 */
  All = 'all',
  /** 仅搜索名称 */
  Name = 'name',
  /** 搜索名称和标签 */
  NameAndTags = 'nameAndTags',
}

/** 过滤条件 */
export interface FilterOptions {
  /** 按资源类型过滤 */
  types?: ResourceTypeEnum[]
  /** 按状态过滤 */
  statuses?: ResourceStatusEnum[]
  /** 按节点过滤 */
  nodes?: string[]
  /** 按标签过滤 */
  tags?: string[]
  /** 是否隐藏已停止资源 */
  hideStopped?: boolean
}

/** 树操作选项 */
export interface TreeActionOptions {
  /** 展开的节点 ID 列表 */
  expandedKeys: string[]
  /** 当前选中的节点 ID */
  selectedKey: string | null
  /** 搜索选项 */
  search?: SearchOptions
  /** 过滤选项 */
  filter?: FilterOptions
}

// ============================================================
// 树统计信息接口
// ============================================================

/** 资源树统计摘要 */
export interface TreeStatistics {
  /** 资源总数 */
  totalResources: number
  /** 按类型统计 */
  byType: Partial<Record<ResourceTypeEnum, number>>
  /** 按状态统计 */
  byStatus: Partial<Record<ResourceStatusEnum, number>>
  /** 节点数量 */
  nodeCount: number
  /** 数据中心数量 */
  dataCenterCount: number
}

// ============================================================
// 类型守卫
// ============================================================

/** 判断是否为资源分组节点 */
export function isResourceGroupNode(node: AbstractTreeNode): node is ResourceGroupNode {
  return 'totalCount' in node && node.depth === 2
}

/** 判断是否为叶子资源节点 */
export function isLeafResourceNode(node: AbstractTreeNode): node is LeafResourceNode {
  return 'resourceId' in node && node.children.length === 0
}
