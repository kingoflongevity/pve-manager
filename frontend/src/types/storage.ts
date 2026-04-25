/**
 * 存储管理类型定义
 * 定义存储相关的所有数据结构和 API 交互类型
 */

// ============================================================
// 存储类型枚举
// ============================================================

/** 存储后端类型 */
export enum StorageBackendType {
  /** 本地目录 */
  DIRECTORY = 'dir',
  /** NFS 网络存储 */
  NFS = 'nfs',
  /** LVM 逻辑卷 */
  LVM = 'lvm',
  /** LVM-Thin 精简卷 */
  LVM_THIN = 'lvmthin',
  /** ZFS 文件系统 */
  ZFS = 'zfs',
  /** Ceph RADOS 块存储 */
  CEPH = 'rbd',
  /** GlusterFS 分布式文件系统 */
  GLUSTERFS = 'glusterfs',
  /** iSCSI 目标 */
  ISCSI = 'iscsi',
  /** SMB/CIFS 共享 */
  CIFS = 'cifs',
  /** Proxmox Backup Server */
  PBS = 'pbs',
}

/** 存储类型中文映射 */
export const StorageBackendTypeLabel: Record<string, string> = {
  dir: '本地目录',
  nfs: 'NFS',
  lvm: 'LVM',
  lvmthin: 'LVM-Thin',
  zfs: 'ZFS',
  rbd: 'Ceph RBD',
  glusterfs: 'GlusterFS',
  iscsi: 'iSCSI',
  cifs: 'SMB/CIFS',
  pbs: 'PBS',
}

// ============================================================
// 存储内容类型
// ============================================================

/** 存储可存储的内容类型 */
export enum StorageContentType {
  /** 磁盘镜像（虚拟机） */
  DISK_IMAGE = 'images',
  /** 容器模板 */
  CONTAINER = 'rootdir',
  /** ISO 镜像 */
  ISO = 'iso',
  /** 备份文件 */
  BACKUP = 'backup',
  /** 代码片段/模板文件 */
  SNIPPETS = 'snippets',
  /** 容器模板（LXC） */
  VZTMPL = 'vztmpl',
}

/** 存储内容类型中文映射 */
export const StorageContentTypeLabel: Record<string, string> = {
  images: '磁盘镜像',
  rootdir: '容器模板',
  iso: 'ISO 镜像',
  backup: '备份',
  snippets: '代码片段',
  vztmpl: '容器模板',
}

// ============================================================
// 存储列表项
// ============================================================

/** 存储列表项（简要信息） */
export interface StorageItem {
  /** 存储 ID */
  storage: string
  /** 所属节点 */
  node: string
  /** 存储类型 */
  type: string
  /** 存储状态: active / inactive */
  status: 'active' | 'inactive'
  /** 总容量（字节） */
  total: number
  /** 已使用容量（字节） */
  used: number
  /** 可用容量（字节） */
  available: number
  /** 使用率（百分比 0-100） */
  usage: number
  /** 支持的内容类型 */
  content: string[]
  /** 是否激活 */
  active: boolean
  /** 是否共享存储 */
  shared: boolean
  /** 存储路径（本地存储） */
  path?: string
  /** NFS 服务器地址 */
  server?: string
  /** NFS 导出路径 */
  export?: string
  /** 共享名称（CIFS） */
  share?: string
  /** 存储池名称（ZFS/LVM） */
  pool?: string
  /** 卷组名称（LVM） */
  vgname?: string
}

/** 存储列表 API 响应 */
export interface StorageListResponse {
  /** 存储列表 */
  data: StorageItem[]
}

// ============================================================
// 存储详情
// ============================================================

/** 存储详情（完整信息） */
export interface StorageDetail extends StorageItem {
  /** 存储完整配置 */
  config: Record<string, unknown>
  /** 内容类型列表 */
  contentTypes: string[]
  /** 最大备份数 */
  maxBackup?: number
  /** 修剪配置 */
  pruneBackups?: string
  /** NFS 挂载选项 */
  options?: string
  /** 是否启用 */
  enabled: boolean
}

// ============================================================
// 存储内容项
// ============================================================

/** 存储中的文件/卷信息 */
export interface StorageContentItem {
  /** 卷路径（如 local:iso/ubuntu.iso） */
  volid: string
  /** 文件名 */
  filename: string
  /** 内容类型 */
  content: string
  /** 文件大小（字节） */
  size: number
  /** 创建时间戳 */
  ctime?: number
  /** 父卷（快照相关） */
  parent?: string
  /** 格式（raw, qcow2 等） */
  format?: string
  /** VM ID */
  vmid?: number
  /** 卷大小 */
  volsize?: number
}

/** 存储内容 API 响应 */
export interface StorageContentResponse {
  /** 内容列表 */
  data: StorageContentItem[]
}

// ============================================================
// 创建/更新存储参数
// ============================================================

/** 创建存储参数 */
export interface CreateStorageParams {
  /** 存储 ID */
  storage: string
  /** 存储类型 */
  type: StorageBackendType | string
  /** 是否启用 */
  enable?: boolean
  /** 是否共享存储 */
  shared?: boolean
  /** 内容类型（逗号分隔） */
  content?: string
  /** 目录路径（Directory 类型） */
  path?: string
  /** NFS 服务器（NFS 类型） */
  server?: string
  /** NFS 导出路径（NFS 类型） */
  export?: string
  /** NFS 挂载选项（NFS 类型） */
  options?: string
  /** 卷组名称（LVM 类型） */
  vgname?: string
  /** 精简池名称（LVM-Thin 类型） */
  thinpool?: string
  /** ZFS 池名称（ZFS 类型） */
  pool?: string
  /** Ceph 监控节点（Ceph 类型） */
  monitors?: string
  /** Ceph 用户名（Ceph 类型） */
  username?: string
  /** GlusterFS 服务器地址 */
  serverName?: string
  /** GlusterFS 卷名 */
  volumeName?: string
  /** iSCSI 门户 */
  portal?: string
  /** iSCSI 目标 */
  target?: string
  /** SMB/CIFS 服务器 */
  smbServer?: string
  /** SMB/CIFS 共享名称 */
  smbShare?: string
  /** SMB/CIFS 用户名 */
  smbUsername?: string
  /** SMB 密码 */
  smbPassword?: string
  /** SMB 域名 */
  smbDomain?: string
  /** PBS 服务器地址 */
  serverAddr?: string
  /** PBS 用户名 */
  pbsUsername?: string
  /** PBS 密码 */
  pbsPassword?: string
  /** PBS 指纹 */
  fingerprint?: string
  /** 数据存储名称（PBS） */
  datastore?: string
  /** 最大备份数 */
  maxBackups?: number
  /** 修剪配置 */
  pruneBackups?: string
}

/** 更新存储参数（可选部分字段） */
export type UpdateStorageParams = Partial<CreateStorageParams> & {
  /** 新存储 ID（用于重命名） */
  newStorageId?: string
}
