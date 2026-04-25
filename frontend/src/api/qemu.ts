/**
 * QEMU 虚拟机 API 端点
 * 提供 QEMU 虚拟机的完整生命周期管理功能
 */
import { get, post, put, del } from './request'
import type {
  QEMUVM,
  QEMUConfig,
  QEMUCreateParams,
  QEMUCloneParams,
  QEMUMigrateParams,
  QEMUSnapshot,
  CreateSnapshotParams,
  PendingConfig,
  RRDDataPoint,
  QueryOptions,
} from './types'

/**
 * 获取指定节点的 QEMU 虚拟机列表
 * @param node 节点名称
 * @param options 查询选项（包含 abort signal）
 */
export async function fetchQEMUList(
  node: string,
  options?: QueryOptions,
): Promise<QEMUVM[]> {
  return get<QEMUVM[]>(`/pve/nodes/${node}/qemu`, undefined, options)
}

/**
 * 获取 QEMU 虚拟机配置
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function getQEMUConfig(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<QEMUConfig> {
  return get<QEMUConfig>(`/pve/nodes/${node}/qemu/${vmid}/config`, undefined, options)
}

/**
 * 创建 QEMU 虚拟机
 * @param node 节点名称
 * @param params 创建参数
 * @param options 查询选项
 */
export async function createQEMU(
  node: string,
  params: QEMUCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu`, params as Record<string, unknown>, options)
}

/**
 * 设置 QEMU 虚拟机配置
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param config 配置参数（键值对形式）
 * @param options 查询选项
 */
export async function setQEMUConfig(
  node: string,
  vmid: number,
  config: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/config`, config, options)
}

/**
 * 启动 QEMU 虚拟机
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function startQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/start`, undefined, options)
}

/**
 * 停止 QEMU 虚拟机（强制断电）
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function stopQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/stop`, undefined, options)
}

/**
 * 重启 QEMU 虚拟机
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function rebootQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/reboot`, undefined, options)
}

/**
 * 关闭 QEMU 虚拟机（通过 ACPI 优雅关机）
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function shutdownQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/shutdown`, undefined, options)
}

/**
 * 挂起 QEMU 虚拟机
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function suspendQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/suspend`, undefined, options)
}

/**
 * 恢复 QEMU 虚拟机（从挂起状态）
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function resumeQEMU(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/qemu/${vmid}/status/resume`, undefined, options)
}

/**
 * 获取 QEMU 虚拟机快照列表
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function listSnapshots(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<QEMUSnapshot[]> {
  return get<QEMUSnapshot[]>(`/pve/nodes/${node}/qemu/${vmid}/snapshot`, undefined, options)
}

/**
 * 创建 QEMU 虚拟机快照
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param params 快照参数
 * @param options 查询选项
 */
export async function createSnapshot(
  node: string,
  vmid: number,
  params: CreateSnapshotParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/qemu/${vmid}/snapshot`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 删除 QEMU 虚拟机快照
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param snapname 快照名称
 * @param options 查询选项
 */
export async function deleteSnapshot(
  node: string,
  vmid: number,
  snapname: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/nodes/${node}/qemu/${vmid}/snapshot/${snapname}`, options)
}

/**
 * 回滚到指定快照
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param snapname 快照名称
 * @param options 查询选项
 */
export async function rollbackSnapshot(
  node: string,
  vmid: number,
  snapname: string,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/qemu/${vmid}/snapshot/${snapname}/rollback`,
    undefined,
    options,
  )
}

/**
 * 克隆 QEMU 虚拟机
 * @param node 节点名称
 * @param vmid 源虚拟机 ID
 * @param params 克隆参数
 * @param options 查询选项
 */
export async function cloneQEMU(
  node: string,
  vmid: number,
  params: QEMUCloneParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/qemu/${vmid}/clone`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 迁移 QEMU 虚拟机到另一个节点
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param params 迁移参数
 * @param options 查询选项
 */
export async function migrateQEMU(
  node: string,
  vmid: number,
  params: QEMUMigrateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/qemu/${vmid}/migrate`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 获取 QEMU 虚拟机 RRD 监控数据
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param timeframe 时间范围 (hour|day|week|month|year)
 * @param dataset 数据集 (cpu|memory|network|disk|system)
 * @param options 查询选项
 */
export async function getQEMURRD(
  node: string,
  vmid: number,
  timeframe: string,
  dataset: string,
  options?: QueryOptions,
): Promise<RRDDataPoint[]> {
  return get<RRDDataPoint[]>(
    `/pve/nodes/${node}/qemu/${vmid}/rrd`,
    { timeframe, datasource: dataset },
    options,
  )
}

/**
 * 获取 QEMU 虚拟机待处理配置
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param options 查询选项
 */
export async function getQEMUPending(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<PendingConfig[]> {
  return get<PendingConfig[]>(
    `/pve/nodes/${node}/qemu/${vmid}/pending`,
    undefined,
    options,
  )
}

/**
 * 扩容 QEMU 虚拟机磁盘
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param disk 磁盘标识（如 scsi0、virtio0）
 * @param size 扩容大小（格式如 +10G）
 * @param options 查询选项
 */
export async function resizeQEMUDisk(
  node: string,
  vmid: number,
  disk: string,
  size: string,
  options?: QueryOptions,
): Promise<string> {
  return put<string>(
    `/pve/nodes/${node}/qemu/${vmid}/resize`,
    { disk, size } as Record<string, unknown>,
    options,
  )
}
