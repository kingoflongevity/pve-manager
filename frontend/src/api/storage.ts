/**
 * 存储管理 API 端点
 * 提供存储的创建、查询、更新、删除及内容管理功能
 */
import { get, post, put, del } from './request'
import type {
  Storage,
  StorageStatus,
  StorageContent,
  StorageCreateParams,
  StorageUpdateParams,
  QueryOptions,
} from './types'

/**
 * 获取指定节点的存储列表
 * @param node 节点名称
 * @param options 查询选项
 */
export async function fetchStorageList(
  node: string,
  options?: QueryOptions,
): Promise<Storage[]> {
  return get<Storage[]>(`/pve/nodes/${node}/storage`, undefined, options)
}

/**
 * 获取存储状态（包含容量信息）
 * @param node 节点名称
 * @param storage 存储名称
 * @param options 查询选项
 */
export async function getStorageStatus(
  node: string,
  storage: string,
  options?: QueryOptions,
): Promise<StorageStatus> {
  return get<StorageStatus>(`/pve/nodes/${node}/storage/${storage}/status`, undefined, options)
}

/**
 * 获取存储内容列表
 * @param node 节点名称
 * @param storage 存储名称
 * @param params 过滤参数 (content, vmid, enabled 等)
 * @param options 查询选项
 */
export async function getStorageContent(
  node: string,
  storage: string,
  params?: { content?: string; vmid?: number; enabled?: number },
  options?: QueryOptions,
): Promise<StorageContent[]> {
  return get<StorageContent[]>(`/pve/nodes/${node}/storage/${storage}/content`, params, options)
}

/**
 * 创建存储
 * @param node 节点名称（某些存储类型需要）
 * @param params 存储创建参数
 * @param options 查询选项
 */
export async function createStorage(
  node: string,
  params: StorageCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(`/pve/nodes/${node}/storage`, params as Record<string, unknown>, options)
}

/**
 * 更新存储配置
 * @param node 节点名称
 * @param storage 存储名称
 * @param params 更新参数
 * @param options 查询选项
 */
export async function updateStorage(
  node: string,
  storage: string,
  params: StorageUpdateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/nodes/${node}/storage/${storage}`,
    params as Record<string, unknown>,
    options,
  )
}

/**
 * 删除存储
 * @param node 节点名称
 * @param storage 存储名称
 * @param options 查询选项
 */
export async function deleteStorage(
  node: string,
  storage: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/nodes/${node}/storage/${storage}`, options)
}
