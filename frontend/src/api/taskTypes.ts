/**
 * Task API 类型定义
 * 定义与 PVE 任务相关的数据结构
 */

/** 任务状态类型 */
export type TaskStatus = 'running' | 'success' | 'error' | 'stopped'

/** PVE 任务类型 */
export type TaskType = 'qmstart' | 'qmstop' | 'qmshutdown' | 'qmreset' | 'qmreboot' |
  'qmresume' | 'qmmigrate' | 'qmbackup' | 'vzstart' | 'vzstop' | 'vzshutdown' |
  'vzresume' | 'vzmigrate' | 'vzbackup' | 'vzrestore' | 'aptupdate' | 'upgrade' | 'unknown'

/** 单个任务信息 */
export interface Task {
  /** 任务 ID (UPID:node:pid:starttime:type:vmid:action:user) */
  upid: string
  /** 任务 ID (简化) */
  id: string
  /** 节点名称 */
  node: string
  /** 任务类型 */
  type: TaskType
  /** 目标 VM/CT ID */
  vmid?: number
  /** 任务描述 */
  description: string
  /** 任务状态 */
  status: TaskStatus
  /** 进度 (0-100) */
  progress: number
  /** 任务开始时间戳 (秒) */
  starttime: number
  /** 任务结束时间戳 (秒)，0 表示未结束 */
  endtime: number
  /** 任务退出状态 */
  exitstatus?: string
  /** 执行用户 */
  user?: string
}

/** 集群概览信息 */
export interface ClusterSummary {
  /** 总虚拟机数 */
  totalVMs: number
  /** 运行中虚拟机数 */
  runningVMs: number
  /** 总容器数 */
  totalCTs: number
  /** 运行中容器数 */
  runningCTs: number
  /** 总存储数 */
  totalStorages: number
  /** 总节点数 */
  totalNodes: number
  /** 在线节点数 */
  onlineNodes: number
}

/** 集群节点信息 */
export interface ClusterNode {
  /** 节点名称 */
  name: string
  /** 节点 IP 地址 */
  ip: string
  /** 节点状态: online / offline / warning */
  status: 'online' | 'offline' | 'warning'
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
  /** CPU 核心数 */
  cpus: number
  /** 系统运行时间（秒） */
  uptime: number
  /** 节点类型 */
  type: string
  /** 节点级别 */
  level: string
  /** 运行中的 VM/CT 数量 */
  vmCount: number
  /** 总 VM/CT 数量 */
  vmTotal: number
  /** 网络入站速率 (bytes/s) */
  netin: number
  /** 网络出站速率 (bytes/s) */
  netout: number
}

/** 任务状态筛选选项 */
export type TaskStatusFilter = 'all' | TaskStatus
