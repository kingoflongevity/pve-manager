/**
 * Node 节点管理 API 端点
 * 提供节点状态、服务、任务、网络、系统配置等管理功能
 */
import { get, post, put } from './request'
import type {
  NodeStatus,
  NodeService,
  NodeSyslog,
  NodeTask,
  TaskLogEntry,
  NetInterface,
  DNSConfig,
  APTUpdate,
  RRDDataPoint,
  QueryOptions,
} from './types'

/**
 * 获取节点状态信息
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getNodeStatus(
  node: string,
  options?: QueryOptions,
): Promise<NodeStatus> {
  return get<NodeStatus>(`/pve/nodes/${node}/status`, undefined, options)
}

/**
 * 获取节点版本信息
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getNodeVersion(
  node: string,
  options?: QueryOptions,
): Promise<{ version: string; release: string; repoid: string }> {
  return get<{ version: string; release: string; repoid: string }>(
    `/pve/nodes/${node}/version`,
    undefined,
    options,
  )
}

/**
 * 获取节点服务列表
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getNodeServices(
  node: string,
  options?: QueryOptions,
): Promise<NodeService[]> {
  return get<NodeService[]>(`/pve/nodes/${node}/services`, undefined, options)
}

/**
 * 获取节点系统日志
 * @param node 节点名称
 * @param params 查询参数 (start, limit, since, until 等)
 * @param options 查询选项
 */
export async function getNodeSyslog(
  node: string,
  params?: { start?: number; limit?: number; since?: number; until?: number; service?: string },
  options?: QueryOptions,
): Promise<NodeSyslog[]> {
  return get<NodeSyslog[]>(`/pve/nodes/${node}/syslog`, params, options)
}

/**
 * 获取节点任务列表
 * @param node 节点名称
 * @param params 查询参数 (start, limit, type, status 等)
 * @param options 查询选项
 */
export async function getNodeTasks(
  node: string,
  params?: { start?: number; limit?: number; type?: string; status?: string },
  options?: QueryOptions,
): Promise<NodeTask[]> {
  return get<NodeTask[]>(`/pve/nodes/${node}/tasks`, params, options)
}

/**
 * 获取任务状态
 * @param node 节点名称
 * @param upid 唯一进程标识符
 * @param options 查询选项
 */
export async function getTaskStatus(
  node: string,
  upid: string,
  options?: QueryOptions,
): Promise<{ upid: string; status: string; exitstatus: string }> {
  return get<{ upid: string; status: string; exitstatus: string }>(
    `/pve/nodes/${node}/tasks/${encodeURIComponent(upid)}/status`,
    undefined,
    options,
  )
}

/**
 * 获取任务日志
 * @param node 节点名称
 * @param upid 唯一进程标识符
 * @param params 查询参数 (start, limit)
 * @param options 查询选项
 */
export async function getTaskLog(
  node: string,
  upid: string,
  params?: { start?: number; limit?: number },
  options?: QueryOptions,
): Promise<TaskLogEntry[]> {
  return get<TaskLogEntry[]>(
    `/pve/nodes/${node}/tasks/${encodeURIComponent(upid)}/log`,
    params,
    options,
  )
}

/**
 * 获取网络接口列表
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getNetworkInterfaces(
  node: string,
  options?: QueryOptions,
): Promise<NetInterface[]> {
  return get<NetInterface[]>(`/pve/nodes/${node}/network`, undefined, options)
}

/**
 * 创建网络接口
 * @param node 节点名称
 * @param params 网络接口参数 (iface, type, address, netmask, gateway 等)
 * @param options 查询选项
 */
export async function createNetworkInterface(
  node: string,
  params: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/network`, params, options)
}

/**
 * 更新网络接口配置
 * @param node 节点名称
 * @param iface 接口名称
 * @param params 更新参数
 * @param options 查询选项
 */
export async function updateNetworkInterface(
  node: string,
  iface: string,
  params: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return put<string>(`/pve/nodes/${node}/network/${iface}`, params, options)
}

/**
 * 删除网络接口
 * @param node 节点名称
 * @param iface 接口名称
 * @param options 查询选项
 */
export async function deleteNetworkInterface(
  node: string,
  iface: string,
  options?: QueryOptions,
): Promise<string> {
  return get<string>(`/pve/nodes/${node}/network/${iface}`, { method: 'delete' } as Record<string, unknown>, options)
}

/**
 * 应用网络配置变更
 * @param node 节点名称
 * @param options 查询选项
 */
export async function applyNetworkChanges(
  node: string,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/network`, { method: 'apply' } as Record<string, unknown>, options)
}

/**
 * 获取 APT 可更新包列表
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getAPTUpdate(
  node: string,
  options?: QueryOptions,
): Promise<APTUpdate[]> {
  return get<APTUpdate[]>(`/pve/nodes/${node}/apt/update`, undefined, options)
}

/**
 * 更新软件包
 * @param node 节点名称
 * @param packages 包名称列表
 * @param options 查询选项
 */
export async function updatePackages(
  node: string,
  packages: string[],
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/apt/update`,
    { packages: packages.join(' ') } as Record<string, unknown>,
    options,
  )
}

/**
 * 获取 DNS 配置
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getDNS(
  node: string,
  options?: QueryOptions,
): Promise<DNSConfig> {
  return get<DNSConfig>(`/pve/nodes/${node}/dns`, undefined, options)
}

/**
 * 设置 DNS 配置
 * @param node 节点名称
 * @param config DNS 配置 (dns1, dns2, dns3, search)
 * @param options 查询选项
 */
export async function setDNS(
  node: string,
  config: DNSConfig,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/dns`, config as Record<string, unknown>, options)
}

/**
 * 获取节点时间
 * @param node 节点名称
 * @param options 查询选项
 */
export async function getTime(
  node: string,
  options?: QueryOptions,
): Promise<{ time: number; timezone: string }> {
  return get<{ time: number; timezone: string }>(`/pve/nodes/${node}/time`, undefined, options)
}

/**
 * 操作指定服务（启动/停止/重启）
 * @param node 节点名称
 * @param service 服务名称
 * @param action 操作类型 (start|stop|restart)
 * @param options 查询选项
 */
export async function actionService(
  node: string,
  service: string,
  action: string,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/services/${service}/${action}`, undefined, options)
}

/**
 * 重启指定服务
 * @param node 节点名称
 * @param service 服务名称
 * @param options 查询选项
 */
export async function restartService(
  node: string,
  service: string,
  options?: QueryOptions,
): Promise<string> {
  return actionService(node, service, 'restart', options)
}

/**
 * 启动指定服务
 * @param node 节点名称
 * @param service 服务名称
 * @param options 查询选项
 */
export async function startService(
  node: string,
  service: string,
  options?: QueryOptions,
): Promise<string> {
  return actionService(node, service, 'start', options)
}

/**
 * 停止指定服务
 * @param node 节点名称
 * @param service 服务名称
 * @param options 查询选项
 */
export async function stopService(
  node: string,
  service: string,
  options?: QueryOptions,
): Promise<string> {
  return actionService(node, service, 'stop', options)
}

// ============================================================
// 节点 RRD 监控数据
// ============================================================

/**
 * 获取节点 RRD 性能监控数据
 * @param node 节点名称
 * @param timeframe 时间范围 (hour|day|week|month|year)
 * @param dataset 数据集 (cpu|mem|net|disk|system)
 * @param options 查询选项
 */
export async function getNodeRRD(
  node: string,
  timeframe: string,
  dataset: string,
  options?: QueryOptions,
): Promise<RRDDataPoint[]> {
  return get<RRDDataPoint[]>(
    `/pve/nodes/${node}/rrd`,
    { timeframe, ds: dataset },
    options,
  )
}
