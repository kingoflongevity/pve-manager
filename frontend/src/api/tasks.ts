/**
 * 任务相关 API 接口
 * 封装与 PVE 任务相关的后端交互
 */
import { get } from './request'
import type { Task } from './taskTypes'

/**
 * 获取集群所有任务列表
 * 注意: request.ts 响应拦截器已提取 res.data，此处直接返回数组
 */
export function fetchTasks(): Promise<Task[]> {
  return get('/pve/cluster/tasks')
}

/**
 * 获取单个任务的日志输出
 * @param upid - 任务 UPID
 */
export function fetchTaskLog(upid: string): Promise<string[]> {
  return get(`/pve/nodes/${extractNodeFromUPID(upid)}/tasks/${encodeURIComponent(upid)}/log`)
}

/**
 * 获取指定节点的最近任务列表
 * @param node - 节点名称
 * @param limit - 返回条数限制
 */
export function fetchNodeTasks(node: string, limit = 50): Promise<Task[]> {
  return get(`/pve/nodes/${node}/tasks`, { limit })
}

/**
 * 从 UPID 字符串中提取节点名称
 * UPID 格式: UPID:node:0001:...
 */
function extractNodeFromUPID(upid: string): string {
  const parts = upid.split(':')
  return parts.length >= 2 ? parts[1] : 'unknown'
}
