/**
 * API 层类型定义
 * 定义与后端 API 交互的数据结构
 */

/** 后端统一响应格式 */
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
}

/** PVE 节点状态信息 */
export interface NodeStatus {
  /** 节点主机名 */
  node: string
  /** CPU 使用率 (0-1) */
  cpu: number
  /** 总内存字节数 */
  maxmem: number
  /** 已使用内存字节数 */
  mem: number
  /** 总磁盘字节数 */
  maxdisk: number
  /** 已使用磁盘字节数 */
  disk: number
  /** 系统运行时间（秒） */
  uptime: number
  /** PVE 版本 */
  version: string
  /** CPU 核心数 */
  cpus: number
}

/** 虚拟机资源信息 */
export interface VMResource {
  /** 资源类型: qemu / lxc / storage / node */
  type: string
  /** 虚拟机/容器 ID */
  vmid: number
  /** 名称 */
  name: string
  /** 节点名称 */
  node: string
  /** 运行状态: running / stopped */
  status: string
  /** CPU 使用率 */
  cpu: number
  /** 最大内存 */
  maxmem: number
  /** 已使用内存 */
  mem: number
  /** 最大磁盘 */
  maxdisk: number
  /** 已使用磁盘 */
  disk: number
}

/** 虚拟机配置信息 */
export interface VMConfig {
  /** 虚拟机 ID */
  vmid: number
  /** 名称 */
  name: string
  /** CPU 核心数 */
  cores: number
  /** 内存大小 */
  memory: number
  /** 磁盘大小 */
  disk: number
  /** 网络配置 */
  net: string[]
  /** 操作系统类型 */
  ostype: string
  /** 启动顺序 */
  boot: string
}

/** 存储信息 */
export interface StorageInfo {
  /** 存储 ID */
  storage: string
  /** 存储类型: dir / nfs / lvm / zfs 等 */
  type: string
  /** 内容类型: images,rootdir,iso,vztmpl,backup */
  content: string[]
  /** 总容量 */
  total: number
  /** 已使用容量 */
  used: number
  /** 是否激活 */
  active: boolean
  /** 是否共享存储 */
  shared: boolean
}
