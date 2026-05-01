/**
 * LXC 容器 API 端点
 * 提供 LXC 容器的完整生命周期管理功能
 */
import { get, post, put, del } from './request'
import type {
  LXCContainer,
  LXCConfig,
  LXCCreateParams,
  LXCCloneParams,
  LXCMigrateParams,
  LXCSnapshot,
  CreateSnapshotParams,
  PendingConfig,
  RRDDataPoint,
  QueryOptions,
} from './types'

/**
 * 获取指定节点的 LXC 容器列表
 * @param node 节点名称
 * @param options 查询选项（包含 abort signal）
 */
export async function fetchLXCList(
  node: string,
  options?: QueryOptions,
): Promise<LXCContainer[]> {
  return get<LXCContainer[]>(`/pve/nodes/${node}/lxc`, undefined, options)
}

/**
 * 获取 LXC 容器配置
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function getLXCConfig(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<LXCConfig> {
  return get<LXCConfig>(`/pve/nodes/${node}/lxc/${vmid}/config`, undefined, options)
}

/**
 * 创建 LXC 容器
 * @param node 节点名称
 * @param params 创建参数
 * @param options 查询选项
 */
export async function createLXC(
  node: string,
  params: LXCCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc`, params as Record<string, unknown>, options)
}

/**
 * 启动 LXC 容器
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function startLXC(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/status/start`, undefined, options)
}

/**
 * 停止 LXC 容器
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function stopLXC(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/status/stop`, undefined, options)
}

/**
 * 重启 LXC 容器
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function rebootLXC(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/status/reboot`, undefined, options)
}

/**
 * 冻结 LXC 容器（暂停所有进程）
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function freezeLXC(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/status/freeze`, undefined, options)
}

/**
 * 解冻 LXC 容器（恢复所有进程）
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function unfreezeLXC(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/status/unfreeze`, undefined, options)
}

/**
 * 设置 LXC 容器配置
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param config 配置参数（键值对形式）
 * @param options 查询选项
 */
export async function setLXCConfig(
  node: string,
  vmid: number,
  config: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/lxc/${vmid}/config`, config, options)
}

/**
 * 获取 LXC 容器快照列表
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function listSnapshots(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<LXCSnapshot[]> {
  return get<LXCSnapshot[]>(`/pve/nodes/${node}/lxc/${vmid}/snapshot`, undefined, options)
}

/**
 * 创建 LXC 容器快照
 * @param node 节点名称
 * @param vmid 容器 ID
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
    `/pve/nodes/${node}/lxc/${vmid}/snapshot`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 删除 LXC 容器快照
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param snapname 快照名称
 * @param options 查询选项
 */
export async function deleteSnapshot(
  node: string,
  vmid: number,
  snapname: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/nodes/${node}/lxc/${vmid}/snapshot/${snapname}`, options)
}

/**
 * 回滚到指定快照
 * @param node 节点名称
 * @param vmid 容器 ID
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
    `/pve/nodes/${node}/lxc/${vmid}/snapshot/${snapname}/rollback`,
    undefined,
    options,
  )
}

/**
 * 克隆 LXC 容器
 * @param node 节点名称
 * @param vmid 源容器 ID
 * @param params 克隆参数
 * @param options 查询选项
 */
export async function cloneLXC(
  node: string,
  vmid: number,
  params: LXCCloneParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/lxc/${vmid}/clone`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 迁移 LXC 容器到另一个节点
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param params 迁移参数
 * @param options 查询选项
 */
export async function migrateLXC(
  node: string,
  vmid: number,
  params: LXCMigrateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/lxc/${vmid}/migrate`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 获取 LXC 容器 RRD 监控数据
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param timeframe 时间范围 (hour|day|week|month|year)
 * @param dataset 数据集 (cpu|memory|network|disk|system)
 * @param options 查询选项
 */
export async function getLXCRRD(
  node: string,
  vmid: number,
  timeframe: string,
  dataset: string,
  options?: QueryOptions,
): Promise<RRDDataPoint[]> {
  return get<RRDDataPoint[]>(
    `/pve/nodes/${node}/lxc/${vmid}/rrd`,
    { timeframe, ds: dataset },
    options,
  )
}

/**
 * 获取 LXC 容器待处理配置
 * @param node 节点名称
 * @param vmid 容器 ID
 * @param options 查询选项
 */
export async function getLXCPending(
  node: string,
  vmid: number,
  options?: QueryOptions,
): Promise<PendingConfig[]> {
  return get<PendingConfig[]>(
    `/pve/nodes/${node}/lxc/${vmid}/pending`,
    undefined,
    options,
  )
}
