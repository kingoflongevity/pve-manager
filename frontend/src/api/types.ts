/**
 * PVE WebUI API 层完整类型定义
 * 定义与 Proxmox VE API 交互的数据结构
 */

// ============================================================
// 通用响应类型
// ============================================================

/**
 * 后端统一响应格式
 * @template T 响应数据类型
 */
export interface APIResponse<T = unknown> {
  code: number
  data: T
  message: string
}

// ============================================================
// QEMU 虚拟机类型
// ============================================================

/** QEMU 虚拟机状态 */
export type QEMUStatus = 'running' | 'stopped' | 'paused' | 'prelaunch' | 'migrate' | 'suspended' | 'unknown'

/** QEMU 虚拟机基本信息 */
export interface QEMUVM {
  vmid: number
  name: string
  node: string
  status: QEMUStatus
  cpu: number
  maxcpu: number
  mem: number
  maxmem: number
  disk: number
  maxdisk: number
  uptime: number
  tags?: string
}

/** QEMU 虚拟机配置 */
export interface QEMUConfig {
  vmid: number
  name?: string
  description?: string
  memory: number
  balloon?: number
  cores: number
  sockets: number
  cpu?: string
  numac?: number
  boot?: string
  bootdisk?: string
  ostype?: string
  agent?: string
  net?: string[]
  scsi?: string[]
  ide?: string[]
  virtio?: string[]
  sata?: string[]
  onboot?: number
  autostart?: number
  startup?: string
  tablet?: number
  vga?: string
  usb?: string[]
  tags?: string
  protection?: number
  template?: number
}

/** QEMU 创建参数 */
export interface QEMUCreateParams {
  vmid: number
  name?: string
  description?: string
  memory?: number
  cores?: number
  sockets?: number
  ostype?: string
  ide?: string[]
  scsi?: string[]
  net?: string[]
  boot?: string
  onboot?: number
  autostart?: number
  startup?: string
  agent?: number
  vga?: string
  usb?: string[]
  cpu?: string
  numac?: number
  balloon?: number
  pool?: string
  tags?: string
  protection?: number
}

/** QEMU 克隆参数 */
export interface QEMUCloneParams {
  newid: number
  name?: string
  description?: string
  full?: number
  target?: string
  storage?: string
  pool?: string
  snapshot?: string
  format?: string
}

/** QEMU 迁移参数 */
export interface QEMUMigrateParams {
  target: string
  online?: number
  migration_type?: string
  with_local_disks?: number
  restart?: number
}

/** QEMU 快照信息 */
export interface QEMUSnapshot {
  name: string
  vmid: number
  snaptime: number
  description?: string
  parent?: string
  vmstate: number
}

// ============================================================
// LXC 容器类型
// ============================================================

/** LXC 容器状态 */
export type LXCStatus = 'running' | 'stopped' | 'frozen' | 'mounting' | 'unknown'

/** LXC 容器基本信息 */
export interface LXCContainer {
  vmid: number
  name: string
  node: string
  status: LXCStatus
  cpu: number
  cpus: number
  mem: number
  maxmem: number
  disk: number
  maxdisk: number
  uptime: number
  tags?: string
  swap?: number
  maxswap?: number
  net?: string[]
  hwm?: number
  ha?: string
}

/** LXC 容器配置 */
export interface LXCConfig {
  vmid: number
  name?: string
  description?: string
  memory: number
  swap: number
  cores: number
  ostype?: string
  rootfs?: string
  mp?: string[]
  net?: string[]
  onboot?: number
  startup?: string
  arch?: string
  unprivileged?: number
  protection?: number
  template?: number
  tags?: string
  hostname?: string
  console?: number
  tty?: number
}

/** LXC 创建参数 */
export interface LXCCreateParams {
  vmid: number
  ostemplate?: string
  password?: string
  hostname?: string
  memory?: number
  swap?: number
  cores?: number
  storage?: string
  rootfs?: string
  net?: string[]
  ostype?: string
  arch?: string
  unprivileged?: number
  onboot?: number
  startup?: string
  tags?: string
  description?: string
  pool?: string
}

/** LXC 克隆参数 */
export interface LXCCloneParams {
  newid: number
  name?: string
  description?: string
  target?: string
  storage?: string
  pool?: string
  snapshot?: string
  full?: number
}

/** LXC 迁移参数 */
export interface LXCMigrateParams {
  target: string
  online?: number
  restart?: number
}

/** LXC 快照信息 */
export interface LXCSnapshot {
  name: string
  vmid: number
  snaptime: number
  description?: string
  parent?: string
}

// ============================================================
// 快照通用类型
// ============================================================

/** 快照信息 (通用) */
export interface Snapshot {
  name: string
  vmid: number
  snaptime: number
  description?: string
  parent?: string
  vmstate?: number
}

/** 创建快照参数 */
export interface CreateSnapshotParams {
  snapname: string
  description?: string
  vmstate?: number
}

// ============================================================
// 节点管理类型
// ============================================================

/** 节点状态 */
export interface NodeStatus {
  node: string
  status: string
  cpu: number
  maxcpu: number
  mem: number
  maxmem: number
  disk: number
  maxdisk: number
  uptime: number
  version: string
  kversion: string
  pveversion: string
  cpus: number
  idle: number
  loadavg: number[]
  swap: number
  maxswap: number
  rootfs: {
    total: number
    used: number
    avail: number
    free: number
  }
}

/** 节点服务信息 */
export interface NodeService {
  name: string
  desc: string
  state: string
  loaded: string
  active: boolean
  sub_state: string
}

/** 节点系统日志 */
export interface NodeSyslog {
  message: string
  priority: string
  timestamp: number
}

/** 网络接口信息 */
export interface NetInterface {
  iface: string
  active: number
  method: string
  address?: string
  netmask?: string
  gateway?: string
  type: string
  cidr?: string
  priority?: number
  autostart?: number
  bridge_ports?: string
  bridge_stp?: number
  bridge_fd?: number
}

/** DNS 配置 */
export interface DNSConfig {
  dns1?: string
  dns2?: string
  dns3?: string
  search?: string
}

/** APT 更新信息 */
export interface APTUpdate {
  Package: string
  Title: string
  Version: string
  Priority: string
  Info?: string
}

// ============================================================
// 任务管理类型
// ============================================================

/** 任务状态 */
export type TaskStatus = 'running' | 'OK' | 'ERROR' | 'stopped' | 'unknown'

/** 任务信息 */
export interface NodeTask {
  pid: number
  pstart: number
  starttime: number
  type: string
  user: string
  upid: string
  id?: string
  status?: TaskStatus
}

/** 任务日志条目 */
export interface TaskLogEntry {
  n: number
  t: string
  p: string
}

/** 待处理配置 */
export interface PendingConfig {
  key: string
  old: string
  new: string
  pending: number
}

// ============================================================
// 存储管理类型
// ============================================================

/** 存储类型 */
export type StorageType = 'dir' | 'nfs' | 'cifs' | 'iscsi' | 'lvm' | 'lvmthin' | 'zfspool' | 'btrfs' | 'pbs' | 'rbd' | 'glusterfs'

/** 存储基本信息 */
export interface Storage {
  storage: string
  type: StorageType
  content: string
  active: number
  enabled: number
  shared: number
  node?: string
}

/** 存储状态 */
export interface StorageStatus {
  storage: string
  type: StorageType
  content: string
  active: number
  enabled: number
  shared: number
  total: number
  used: number
  avail: number
}

/** 存储内容条目 */
export interface StorageContent {
  volid: string
  content: string
  format?: string
  size: number
  parent?: string
  ctime?: number
  vmid?: number
  nodes?: string
  verified?: number
  backupid?: string
  protection?: number
}

/** 存储创建参数 */
export interface StorageCreateParams {
  storage: string
  type: StorageType
  content: string
  path?: string
  server?: string
  export?: string
  options?: string
  username?: string
  password?: string
  domain?: string
  pool?: string
  monhost?: string
  krbd?: number
  maxfiles?: number
  prune_backups?: string
  nodes?: string
  shared?: number
  disable?: number
}

/** 存储更新参数 */
export interface StorageUpdateParams {
  content?: string
  path?: string
  server?: string
  export?: string
  options?: string
  username?: string
  password?: string
  domain?: string
  pool?: string
  monhost?: string
  krbd?: number
  maxfiles?: number
  prune_backups?: string
  nodes?: string
  shared?: number
  disable?: number
}

// ============================================================
// 集群管理类型
// ============================================================

/** 集群资源类型 */
export type ClusterResourceType = 'node' | 'vm' | 'lxc' | 'storage' | 'pool'

/** 集群资源 */
export interface ClusterResource {
  id: string
  type: ClusterResourceType
  vmid?: number
  node?: string
  name?: string
  status?: string
  mem?: number
  maxmem?: number
  disk?: number
  maxdisk?: number
  cpu?: number
  maxcpu?: number
  uptime?: number
  tags?: string
  pool?: string
  level?: number
  storage?: string
  content?: string
  shared?: number
}

/** 资源池 */
export interface Pool {
  poolid: string
  comment?: string
  members?: PoolMember[]
}

/** 资源池成员 */
export interface PoolMember {
  id: string
  type: string
  vmid?: number
  node?: string
}

/** HA 配置 */
export interface HAConfig {
  groups?: HAGroup[]
  resources?: HAResource[]
  fencing?: FencingConfig
}

/** HA 组 */
export interface HAGroup {
  group: string
  comment?: string
  nodes?: string
  restricted?: number
  nofailback?: number
  nodename?: string
}

/** HA 资源 */
export interface HAResource {
  type: string
  sid: string
  group?: string
  state: string
  comment?: string
  max_relocate?: number
  max_restart?: number
  prefered_node?: string
}

/** HA Fencing 配置 */
export interface FencingConfig {
  type: string
  params?: string
}

/** SDN Zone */
export interface SDNZone {
  zone: string
  type: string
  nodes?: string
  ipam?: string
  dns?: string
  dnsname?: string
}

/** SDN VNET */
export interface SDNVNET {
  vnet: string
  zone: string
  alias?: string
  tag?: number
  firewall?: number
  bridge?: string
  mtu?: number
}

/** 复制任务 */
export interface ReplicationJob {
  id: string
  type: string
  source: string
  target: string
  guest: number
  schedule?: string
  disable?: number
  rate?: number
  jobnum?: number
  fail_count?: number
  last_sync?: number
  last_try?: number
  duration?: number
}

// ============================================================
// 访问控制类型
// ============================================================

/** 用户信息 */
export interface User {
  userid: string
  username: string
  realm: string
  firstname?: string
  lastname?: string
  email?: string
  enabled: number
  expire: number
  tokens?: string[]
}

/** 用户创建参数 */
export interface UserCreateParams {
  userid: string
  password?: string
  firstname?: string
  lastname?: string
  email?: string
  enabled?: number
  expire?: number
  groups?: string[]
  comment?: string
}

/** 用户组 */
export interface Group {
  groupid: string
  comment?: string
  users?: string[]
}

/** 用户组创建参数 */
export interface GroupCreateParams {
  groupid: string
  comment?: string
  users?: string[]
}

/** 角色 */
export interface Role {
  roleid: string
  privs: string
}

/** 角色创建参数 */
export interface RoleCreateParams {
  roleid: string
  privs: string
}

/** ACL 条目 */
export interface ACL {
  path: string
  roleid: string
  propagate?: number
  auth_id?: string
  type: string
}

/** ACL 设置参数 */
export interface ACLSetParams {
  path: string
  roles: string
  auth_id?: string
  users?: string
  groups?: string
  delete?: number
  propagate?: number
}

// ============================================================
// RRD 监控数据类型
// ============================================================

/** RRD 时间范围 */
export type RRDTimeframe = 'hour' | 'day' | 'week' | 'month' | 'year'

/** RRD 数据集 */
export type RRDDataSet = 'cpu' | 'memory' | 'network' | 'disk' | 'system'

/** RRD 数据点 */
export interface RRDDataPoint {
  time: number
  [key: string]: number | undefined
}

/** RRD 数据响应 */
export interface RRDData {
  timeframe: RRDTimeframe
  dataset: RRDDataSet
  data: RRDDataPoint[]
}

// ============================================================
// 通用请求/响应类型
// ============================================================

/** 分页查询参数 */
export interface PaginationParams {
  offset?: number
  limit?: number
}

/** 排序参数 */
export interface SortParams {
  sort_by?: string
  sort_descending?: number
}

/** 查询选项 */
export interface QueryOptions {
  signal?: AbortSignal
}

/** 任务结果 */
export interface TaskResult {
  upid: string
  status: TaskStatus
  exitstatus: string
  starttime: number
  endtime?: number
}

// ============================================================
// 数据中心资源类型（用于 tree 组件 - 向后兼容）
// ============================================================

/** 资源类型 */
export type ResourceType = 'datacenter' | 'node' | 'vm' | 'ct' | 'storage' | 'network' | 'pool'

/** 资源状态 */
export type ResourceStatus = 'running' | 'stopped' | 'paused' | 'error' | 'unknown' | 'frozen'

/** 树形资源数据 (旧版 - 保留兼容) */
export interface TreeResourceData {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  icon?: string
  children?: TreeResourceData[]
}

/** 节点信息 */
export interface PVENode {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  cpuUsage?: number
  memoryUsage?: number
  diskUsage?: number
  uptime?: number
  version?: string
  children?: TreeResourceData[]
}

/** 虚拟机信息 */
export interface PVEVM {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  vmid: number
  cpus: number
  memoryMB: number
  cpuUsage?: number
  diskUsage?: number
  uptime?: number
  node: string
}

/** 容器信息 */
export interface PVECT {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  ctid: number
  cpus: number
  memoryMB: number
  cpuUsage?: number
  diskUsage?: number
  uptime?: number
  node: string
}

/** 存储信息 */
export interface PVEStorage {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  storageType: string
  total: number
  used: number
  available: number
  usage: number
  active: boolean
  node: string
}

/** 网络信息 */
export interface PVENetwork {
  id: string
  name: string
  type: ResourceType
  status: ResourceStatus
  interfaceType: string
  address?: string
  netmask?: string
  gateway?: string
  active: boolean
  node: string
}

/** 资源列表响应 */
export interface ResourceListResponse {
  nodes: PVENode[]
  vms: PVEVM[]
  containers: PVECT[]
  storages: PVEStorage[]
  networks: PVENetwork[]
}
