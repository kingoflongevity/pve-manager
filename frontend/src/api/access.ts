/**
 * 访问控制 API 端点
 * 提供用户、组、角色、权限（ACL）等访问控制管理功能
 */
import { get, post, put, del } from './request'
import type {
  User,
  UserCreateParams,
  Group,
  GroupCreateParams,
  Role,
  RoleCreateParams,
  ACL,
  ACLSetParams,
  QueryOptions,
} from './types'

// ============================================================
// 用户管理
// ============================================================

/**
 * 获取用户列表
 * @param params 查询参数 (enabled, realm)
 * @param options 查询选项
 */
export async function listUsers(
  params?: { enabled?: number; realm?: string },
  options?: QueryOptions,
): Promise<User[]> {
  return get<User[]>('/pve/access/users', params, options)
}

/**
 * 获取单个用户信息
 * @param userid 用户 ID (格式: username@realm)
 * @param options 查询选项
 */
export async function getUser(
  userid: string,
  options?: QueryOptions,
): Promise<User> {
  return get<User>(`/pve/access/users/${encodeURIComponent(userid)}`, undefined, options)
}

/**
 * 创建用户
 * @param params 用户创建参数
 * @param options 查询选项
 */
export async function createUser(
  params: UserCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>('/pve/access/users', params as Record<string, unknown>, options)
}

/**
 * 更新用户信息
 * @param userid 用户 ID
 * @param params 更新参数
 * @param options 查询选项
 */
export async function updateUser(
  userid: string,
  params: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/access/users/${encodeURIComponent(userid)}`,
    params,
    options,
  )
}

/**
 * 删除用户
 * @param userid 用户 ID
 * @param options 查询选项
 */
export async function deleteUser(
  userid: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/access/users/${encodeURIComponent(userid)}`, options)
}

/**
 * 修改用户密码
 * @param userid 用户 ID
 * @param password 新密码
 * @param options 查询选项
 */
export async function updateUserPassword(
  userid: string,
  password: string,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/access/users/${encodeURIComponent(userid)}/password`,
    { userid, password } as Record<string, unknown>,
    options,
  )
}

// ============================================================
// 用户组管理
// ============================================================

/**
 * 获取用户组列表
 * @param options 查询选项
 */
export async function listGroups(
  options?: QueryOptions,
): Promise<Group[]> {
  return get<Group[]>('/pve/access/groups', undefined, options)
}

/**
 * 获取单个用户组信息
 * @param groupid 组 ID
 * @param options 查询选项
 */
export async function getGroup(
  groupid: string,
  options?: QueryOptions,
): Promise<Group> {
  return get<Group>(`/pve/access/groups/${encodeURIComponent(groupid)}`, undefined, options)
}

/**
 * 创建用户组
 * @param params 组创建参数
 * @param options 查询选项
 */
export async function createGroup(
  params: GroupCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>('/pve/access/groups', params as Record<string, unknown>, options)
}

/**
 * 更新用户组信息
 * @param groupid 组 ID
 * @param params 更新参数
 * @param options 查询选项
 */
export async function updateGroup(
  groupid: string,
  params: Record<string, unknown>,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/access/groups/${encodeURIComponent(groupid)}`,
    params,
    options,
  )
}

/**
 * 删除用户组
 * @param groupid 组 ID
 * @param options 查询选项
 */
export async function deleteGroup(
  groupid: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/access/groups/${encodeURIComponent(groupid)}`, options)
}

// ============================================================
// 角色管理
// ============================================================

/**
 * 获取角色列表
 * @param options 查询选项
 */
export async function listRoles(
  options?: QueryOptions,
): Promise<Role[]> {
  return get<Role[]>('/pve/access/roles', undefined, options)
}

/**
 * 创建角色
 * @param params 角色创建参数
 * @param options 查询选项
 */
export async function createRole(
  params: RoleCreateParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>('/pve/access/roles', params as Record<string, unknown>, options)
}

/**
 * 更新角色权限
 * @param roleid 角色 ID
 * @param privs 权限字符串（逗号分隔）
 * @param options 查询选项
 */
export async function updateRole(
  roleid: string,
  privs: string,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    `/pve/access/roles/${encodeURIComponent(roleid)}`,
    { roleid, privs } as Record<string, unknown>,
    options,
  )
}

/**
 * 删除角色
 * @param roleid 角色 ID
 * @param options 查询选项
 */
export async function deleteRole(
  roleid: string,
  options?: QueryOptions,
): Promise<string> {
  return del<string>(`/pve/access/roles/${encodeURIComponent(roleid)}`, options)
}

// ============================================================
// ACL 权限管理
// ============================================================

/**
 * 获取 ACL 列表
 * @param options 查询选项
 */
export async function listACLs(
  options?: QueryOptions,
): Promise<ACL[]> {
  return get<ACL[]>('/pve/access/acl', undefined, options)
}

/**
 * 设置 ACL 权限
 * @param params ACL 设置参数
 * @param options 查询选项
 */
export async function setACL(
  params: ACLSetParams,
  options?: QueryOptions,
): Promise<string> {
  return post<string>(
    '/pve/access/acl',
    params as Record<string, unknown>,
    options,
  )
}

// ============================================================
// 认证域管理
// ============================================================

/**
 * 获取认证域列表
 * @param options 查询选项
 */
export async function listDomains(
  options?: QueryOptions,
): Promise<{ realm: string; type: string; comment?: string; default: number }[]> {
  return get<{ realm: string; type: string; comment?: string; default: number }[]>(
    '/pve/access/domains',
    undefined,
    options,
  )
}

/**
 * 获取单个认证域信息
 * @param realm 认证域名称
 * @param options 查询选项
 */
export async function getDomain(
  realm: string,
  options?: QueryOptions,
): Promise<Record<string, unknown>> {
  return get<Record<string, unknown>>(
    `/pve/access/domains/${encodeURIComponent(realm)}`,
    undefined,
    options,
  )
}
