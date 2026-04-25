/**
 * 集群管理 API 端点
 * 提供集群资源、任务、池、HA、SDN、复制等管理功能
 */
import { get, post } from './request'
import type {
  ClusterResource,
  Pool,
  HAConfig,
  HAGroup,
  HAResource,
  SDNZone,
  SDNVNET,
  ReplicationJob,
  NodeTask,
  QueryOptions,
} from './types'

/**
 * 获取集群所有资源（节点、VM、CT、存储等）
 * @param options 查询选项
 */
export async function getClusterResources(
  options?: QueryOptions,
): Promise<ClusterResource[]> {
  return get<ClusterResource[]>('/pve/cluster/resources', undefined, options)
}

/**
 * 获取集群任务列表
 * @param options 查询选项
 */
export async function getClusterTasks(
  options?: QueryOptions,
): Promise<NodeTask[]> {
  return get<NodeTask[]>('/pve/cluster/tasks', undefined, options)
}

/**
 * 获取下一个可用的 VM/CT ID
 * @param options 查询选项
 */
export async function getNextID(
  options?: QueryOptions,
): Promise<{ data: number }> {
  return get<{ data: number }>('/pve/cluster/nextid', undefined, options)
}

/**
 * 获取资源池列表
 * @param options 查询选项
 */
export async function getPools(
  options?: QueryOptions,
): Promise<{ poolid: string; comment?: string }[]> {
  return get<{ poolid: string; comment?: string }[]>('/pve/pools', undefined, options)
}

/**
 * 获取指定资源池详情
 * @param poolid 池 ID
 * @param options 查询选项
 */
export async function getPool(
  poolid: string,
  options?: QueryOptions,
): Promise<Pool> {
  return get<Pool>(`/pve/pools/${poolid}`, undefined, options)
}

/**
 * 创建资源池
 * @param params 创建参数 (poolid, comment)
 * @param options 查询选项
 */
export async function createPool(
  params: { poolid: string; comment?: string },
  options?: QueryOptions,
): Promise<string> {
  return post<string>('/pve/pools', params as Record<string, unknown>, options)
}

/**
 * 获取 HA 配置
 * @param options 查询选项
 */
export async function getHAConfig(
  options?: QueryOptions,
): Promise<HAConfig> {
  return get<HAConfig>('/pve/cluster/ha', undefined, options)
}

/**
 * 获取 HA 组列表
 * @param options 查询选项
 */
export async function getHAGroups(
  options?: QueryOptions,
): Promise<HAGroup[]> {
  return get<HAGroup[]>('/pve/cluster/ha/groups', undefined, options)
}

/**
 * 获取 HA 资源列表
 * @param options 查询选项
 */
export async function getHAResources(
  options?: QueryOptions,
): Promise<HAResource[]> {
  return get<HAResource[]>('/pve/cluster/ha/resources', undefined, options)
}

/**
 * 获取 SDN Zone 列表
 * @param options 查询选项
 */
export async function getSDNZones(
  options?: QueryOptions,
): Promise<SDNZone[]> {
  return get<SDNZone[]>('/pve/cluster/sdn/zones', undefined, options)
}

/**
 * 获取 SDN VNET 列表
 * @param options 查询选项
 */
export async function getSDNVNETs(
  options?: QueryOptions,
): Promise<SDNVNET[]> {
  return get<SDNVNET[]>('/pve/cluster/sdn/vnets', undefined, options)
}

/**
 * 获取复制任务列表
 * @param options 查询选项
 */
export async function getReplicationJobs(
  options?: QueryOptions,
): Promise<ReplicationJob[]> {
  return get<ReplicationJob[]>('/pve/cluster/replication', undefined, options)
}

/**
 * 获取数据中心配置
 * @param options 查询选项
 */
export async function getDatacenterConfig(
  options?: QueryOptions,
): Promise<Record<string, unknown>> {
  return get<Record<string, unknown>>('/pve/cluster/config', undefined, options)
}

/**
 * 获取数据中心日志
 * @param params 查询参数
 * @param options 查询选项
 */
export async function getDatacenterLog(
  params?: { start?: number; limit?: number },
  options?: QueryOptions,
): Promise<unknown[]> {
  return get<unknown[]>('/pve/cluster/log', params, options)
}
