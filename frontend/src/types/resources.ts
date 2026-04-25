/**
 * PVE 资源类型定义
 * 定义 Proxmox VE 资源树中所有资源节点的类型
 */

// ============================================================
// 资源状态枚举
// ============================================================

/** 资源运行状态 */
export enum ResourceStatus {
  /** 运行中 */
  RUNNING = 'running',
  /** 已停止 */
  STOPPED = 'stopped',
  /** 错误 */
  ERROR = 'error',
  /** 未知 */
  UNKNOWN = 'unknown',
}

// ============================================================
// 资源类型枚举
// ============================================================

/** 资源类型 */
export enum ResourceType {
  /** 数据中心 */
  DATACENTER = 'datacenter',
  /** 节点 */
  NODE = 'node',
  /** QEMU 虚拟机 */
  VM = 'vm',
  /** LXC 容器 */
  CT = 'ct',
  /** 存储 */
  STORAGE = 'storage',
  /** 网络 */
  NETWORK = 'network',
}

// ============================================================
// 基础资源节点接口
// ============================================================

/** PVE 资源节点基础接口 */
export interface PVEResourceNode {
  /** 节点唯一标识 */
  id: string
  /** 节点名称 */
  name: string
  /** 资源类型 */
  type: ResourceType
  /** 运行状态 */
  status: ResourceStatus
  /** 父节点 ID */
  parentId?: string
  /** 图标标识 */
  icon?: string
}

// ============================================================
// 虚拟机 (QEMU VM)
// ============================================================

/** QEMU 虚拟机 */
export interface PVEVM extends PVEResourceNode {
  type: ResourceType.VM
  /** 虚拟机 ID (VMID) */
  vmid: number
  /** CPU 核心数 */
  cpus?: number
  /** 内存大小 (MB) */
  memoryMB?: number
  /** 磁盘使用率 (%) */
  diskUsage?: number
  /** CPU 使用率 (%) */
  cpuUsage?: number
  /** 运行时间 (秒) */
  uptime?: number
  /** 节点名称 */
  node: string
}

// ============================================================
// 容器 (LXC CT)
// ============================================================

/** LXC 容器 */
export interface PVECT extends PVEResourceNode {
  type: ResourceType.CT
  /** 容器 ID (CTID) */
  ctid: number
  /** CPU 核心数 */
  cpus?: number
  /** 内存大小 (MB) */
  memoryMB?: number
  /** 磁盘使用率 (%) */
  diskUsage?: number
  /** CPU 使用率 (%) */
  cpuUsage?: number
  /** 运行时间 (秒) */
  uptime?: number
  /** 节点名称 */
  node: string
}

// ============================================================
// 存储 (Storage)
// ============================================================

/** PVE 存储 */
export interface PVEStorage extends PVEResourceNode {
  type: ResourceType.STORAGE
  /** 存储类型 (local, nfs, lvm, zfs, ceph 等) */
  storageType: string
  /** 总容量 (字节) */
  total?: number
  /** 已使用容量 (字节) */
  used?: number
  /** 可用容量 (字节) */
  available?: number
  /** 使用率 (%) */
  usage?: number
  /** 是否激活 */
  active?: boolean
  /** 节点名称 */
  node: string
}

// ============================================================
// 网络 (Network)
// ============================================================

/** PVE 网络接口 */
export interface PVENetwork extends PVEResourceNode {
  type: ResourceType.NETWORK
  /** 接口类型 (eth, bond, bridge, vlan) */
  interfaceType: string
  /** IP 地址 */
  address?: string
  /** 子网掩码 */
  netmask?: string
  /** 网关 */
  gateway?: string
  /** 是否激活 */
  active?: boolean
  /** 节点名称 */
  node: string
}

// ============================================================
// 节点 (Node)
// ============================================================

/** PVE 节点 */
export interface PVENode extends PVEResourceNode {
  type: ResourceType.NODE
  /** CPU 使用率 (%) */
  cpuUsage?: number
  /** 内存使用率 (%) */
  memoryUsage?: number
  /** 磁盘使用率 (%) */
  diskUsage?: number
  /** 运行时间 (秒) */
  uptime?: number
  /** PVE 版本 */
  version?: string
  /** 子资源列表 */
  children?: (PVEVM | PVECT | PVEStorage | PVENetwork)[]
}

// ============================================================
// 数据中心 (Datacenter)
// ============================================================

/** PVE 数据中心 */
export interface PVEDatacenter extends PVEResourceNode {
  type: ResourceType.DATACENTER
  /** 子节点列表 */
  children?: PVENode[]
}

// ============================================================
// 树节点包装接口 (用于 Element Plus Tree)
// ============================================================

/** Element Plus Tree 节点数据 */
export interface TreeResourceData extends PVEResourceNode {
  /** 子节点 (递归) */
  children?: TreeResourceData[]
  /** 是否禁用选择 */
  disabled?: boolean
  /** 自定义类名 */
  customClass?: string
}

// ============================================================
// API 响应类型
// ============================================================

/** 资源列表 API 响应 */
export interface ResourceListResponse {
  /** 节点列表 */
  nodes: PVENode[]
  /** 虚拟机列表 */
  vms: PVEVM[]
  /** 容器列表 */
  containers: PVECT[]
  /** 存储列表 */
  storages: PVEStorage[]
  /** 网络列表 */
  networks: PVENetwork[]
}
