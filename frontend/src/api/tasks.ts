/**
 * 任务相关 API 接口
 * 封装与 PVE 任务相关的后端交互
 */
import { get } from './request'
import type { Task } from './taskTypes'

/**
 * 获取集群所有任务列表
 */
export function fetchTasks(): Promise<{ data: Task[] }> {
  return get('/cluster/tasks')
}

/**
 * 获取单个任务的日志输出
 * @param upid - 任务 UPID
 */
export function fetchTaskLog(upid: string): Promise<{ data: string }> {
  return get(`/cluster/tasks/${encodeURIComponent(upid)}/log`)
}

/**
 * 获取指定节点的最近任务列表
 * @param node - 节点名称
 * @param limit - 返回条数限制
 */
export function fetchNodeTasks(node: string, limit = 50): Promise<{ data: Task[] }> {
  return get(`/nodes/${node}/tasks`, { limit })
}
