<template>
  <div class="node-detail-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1 class="page-title">节点详情: {{ nodeName }}</h1>
      </div>
      <div class="header-right">
        <el-button type="danger" size="small" @click="handleShutdown">
          <el-icon><SwitchButton /></el-icon>
          关机
        </el-button>
        <el-button type="warning" size="small" @click="handleReboot">
          <el-icon><RefreshRight /></el-icon>
          重启
        </el-button>
      </div>
    </div>

    <!-- Tab 页签 -->
    <el-card>
      <el-tabs v-model="activeTab" type="border-card">
        <!-- ========== Tab 1: 概览 ========== -->
        <el-tab-pane label="概览" name="overview">
          <div v-loading="overviewLoading" class="tab-content">
            <el-row :gutter="20">
              <!-- 节点信息 -->
              <el-col :xs="24" :md="12">
                <el-card shadow="never" class="info-card">
                  <template #header>
                    <span class="card-title">节点信息</span>
                  </template>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item label="主机名">{{ nodeStatus?.node || '--' }}</el-descriptions-item>
                    <el-descriptions-item label="PVE 版本">{{ nodeVersion?.version || '--' }}</el-descriptions-item>
                    <el-descriptions-item label="内核版本">{{ nodeStatus?.kversion || '--' }}</el-descriptions-item>
                    <el-descriptions-item label="运行时间">{{ formatUptime(nodeStatus?.uptime || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="CPU 核心数">{{ nodeStatus?.cpus || 0 }} 核</el-descriptions-item>
                    <el-descriptions-item label="系统负载">
                      {{ nodeStatus?.loadavg?.join(', ') || '--' }}
                    </el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-col>

              <!-- 内存/磁盘信息 -->
              <el-col :xs="24" :md="12">
                <el-card shadow="never" class="info-card">
                  <template #header>
                    <span class="card-title">资源信息</span>
                  </template>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item label="内存总量">{{ formatBytes(nodeStatus?.maxmem || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="已用内存">{{ formatBytes(nodeStatus?.mem || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="交换分区">{{ formatBytes(nodeStatus?.swap || 0) }} / {{ formatBytes(nodeStatus?.maxswap || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="根分区总量">{{ formatBytes(nodeStatus?.rootfs?.total || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="根分区已用">{{ formatBytes(nodeStatus?.rootfs?.used || 0) }}</el-descriptions-item>
                    <el-descriptions-item label="根分区可用">{{ formatBytes(nodeStatus?.rootfs?.avail || 0) }}</el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-col>
            </el-row>

            <!-- 资源仪表盘 -->
            <el-row :gutter="20" class="gauge-row">
              <el-col :xs="12" :sm="6">
                <div class="gauge-card">
                  <div class="gauge-title">CPU</div>
                  <el-progress
                    type="dashboard"
                    :percentage="cpuPercent"
                    :color="getGaugeColor(cpuPercent)"
                    :width="100"
                  />
                  <div class="gauge-value">{{ cpuPercent.toFixed(1) }}%</div>
                </div>
              </el-col>
              <el-col :xs="12" :sm="6">
                <div class="gauge-card">
                  <div class="gauge-title">内存</div>
                  <el-progress
                    type="dashboard"
                    :percentage="memPercent"
                    :color="getGaugeColor(memPercent)"
                    :width="100"
                  />
                  <div class="gauge-value">{{ memPercent.toFixed(1) }}%</div>
                </div>
              </el-col>
              <el-col :xs="12" :sm="6">
                <div class="gauge-card">
                  <div class="gauge-title">磁盘</div>
                  <el-progress
                    type="dashboard"
                    :percentage="diskPercent"
                    :color="getGaugeColor(diskPercent)"
                    :width="100"
                  />
                  <div class="gauge-value">{{ diskPercent.toFixed(1) }}%</div>
                </div>
              </el-col>
              <el-col :xs="12" :sm="6">
                <div class="gauge-card">
                  <div class="gauge-title">Swap</div>
                  <el-progress
                    type="dashboard"
                    :percentage="swapPercent"
                    :color="getGaugeColor(swapPercent)"
                    :width="100"
                  />
                  <div class="gauge-value">{{ swapPercent.toFixed(1) }}%</div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>

        <!-- ========== Tab 2: 网络 ========== -->
        <el-tab-pane label="网络" name="network">
          <div v-loading="networkLoading" class="tab-content">
            <!-- 网络操作栏 -->
            <div class="toolbar">
              <el-button type="primary" size="small" @click="showNetworkDialog">
                <el-icon><Plus /></el-icon>
                创建接口
              </el-button>
              <el-button type="success" size="small" @click="handleApplyNetwork" :disabled="!networkHasPending">
                <el-icon><Check /></el-icon>
                应用变更
              </el-button>
            </div>

            <!-- 网络接口表格 -->
            <el-table :data="networkInterfaces" border stripe size="small" class="network-table">
              <el-table-column prop="iface" label="接口名称" width="120" />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="getNetTypeTag(row.type)" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="IP 地址" width="180">
                <template #default="{ row }">
                  <span v-if="row.address">{{ row.address }}/{{ getPrefixLength(row.netmask) }}</span>
                  <span v-else class="text-muted">--</span>
                </template>
              </el-table-column>
              <el-table-column prop="gateway" label="网关" width="150">
                <template #default="{ row }">
                  <span v-if="row.gateway">{{ row.gateway }}</span>
                  <span v-else class="text-muted">--</span>
                </template>
              </el-table-column>
              <el-table-column prop="mtu" label="MTU" width="80">
                <template #default="{ row }">
                  {{ row.mtu || '1500' }}
                </template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.active ? 'success' : 'info'" size="small">
                    {{ row.active ? '活跃' : '非活跃' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" size="small" @click="editNetwork(row)">编辑</el-button>
                  <el-popconfirm
                    title="确定删除该网络接口吗？"
                    @confirm="handleDeleteNetwork(row.iface)"
                  >
                    <template #reference>
                      <el-button link type="danger" size="small">删除</el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
            </el-table>

            <!-- 网络拓扑图 -->
            <NetworkTopology :interfaces="networkInterfaces" class="topology-section" />
          </div>
        </el-tab-pane>

        <!-- ========== Tab 3: 系统日志 ========== -->
        <el-tab-pane label="系统日志" name="syslog">
          <div class="tab-content syslog-content">
            <!-- 日志工具栏 -->
            <div class="toolbar">
              <div class="toolbar-left">
                <el-input
                  v-model="syslogSearch"
                  placeholder="搜索日志内容..."
                  :prefix-icon="Search"
                  clearable
                  size="small"
                  style="width: 250px;"
                />
                <el-select v-model="syslogPriority" placeholder="优先级" size="small" style="width: 120px; margin-left: 12px;">
                  <el-option label="全部" value="" />
                  <el-option label="Info" value="info" />
                  <el-option label="Warning" value="warning" />
                  <el-option label="Error" value="error" />
                </el-select>
              </div>
              <div class="toolbar-right">
                <el-switch
                  v-model="syslogAutoScroll"
                  active-text="自动滚动"
                  size="small"
                  style="margin-right: 12px;"
                />
                <el-button size="small" @click="fetchSyslog" :loading="syslogLoading">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </div>
            </div>

            <!-- 日志内容区 -->
            <div ref="syslogContainerRef" class="log-container">
              <div v-if="filteredSyslog.length === 0" class="log-empty">
                <el-empty description="暂无日志数据" />
              </div>
              <div v-else class="log-list">
                <div
                  v-for="(entry, index) in filteredSyslog"
                  :key="index"
                  class="log-entry"
                  :class="`log-${entry.priority}`"
                >
                  <span class="log-time">{{ formatTimestamp(entry.timestamp) }}</span>
                  <el-tag :type="getPriorityTagType(entry.priority)" size="small" class="log-priority">
                    {{ entry.priority }}
                  </el-tag>
                  <span class="log-message">{{ entry.message }}</span>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- ========== Tab 4: 服务 ========== -->
        <el-tab-pane label="服务" name="services">
          <div v-loading="servicesLoading" class="tab-content">
            <el-table :data="services" border stripe size="small">
              <el-table-column prop="name" label="服务名称" width="200" />
              <el-table-column prop="desc" label="描述" min-width="200" show-overflow-tooltip />
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.active ? 'success' : 'danger'" size="small" effect="dark">
                    {{ row.active ? '运行中' : '已停止' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="sub_state" label="子状态" width="120" />
              <el-table-column label="操作" width="220" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-if="!row.active"
                    link
                    type="success"
                    size="small"
                    @click="handleServiceAction('start', row)"
                  >
                    启动
                  </el-button>
                  <el-button
                    v-if="row.active"
                    link
                    type="warning"
                    size="small"
                    @click="handleServiceAction('stop', row)"
                  >
                    停止
                  </el-button>
                  <el-button
                    v-if="row.active"
                    link
                    type="primary"
                    size="small"
                    @click="handleServiceAction('restart', row)"
                  >
                    重启
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- ========== Tab 5: 系统更新 ========== -->
        <el-tab-pane label="系统更新" name="updates">
          <div v-loading="updatesLoading" class="tab-content">
            <div class="toolbar">
              <span class="updates-count">共 {{ updates.length }} 个可用更新</span>
              <div>
                <el-button size="small" @click="fetchUpdates" :loading="updatesLoading">
                  <el-icon><Refresh /></el-icon>
                  检查更新
                </el-button>
                <el-button type="primary" size="small" @click="handleUpgradeAll" :disabled="updates.length === 0">
                  <el-icon><Download /></el-icon>
                  升级全部
                </el-button>
              </div>
            </div>

            <el-table v-if="updates.length > 0" :data="updates" border stripe size="small">
              <el-table-column prop="Package" label="软件包" width="200" show-overflow-tooltip />
              <el-table-column prop="Title" label="描述" min-width="200" show-overflow-tooltip />
              <el-table-column prop="Version" label="新版本" width="150" show-overflow-tooltip />
              <el-table-column prop="Priority" label="优先级" width="100" />
              <el-table-column label="大小" width="100" />
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" size="small" @click="handleUpgradePackage(row.Package)">
                    升级
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="当前系统已是最新版本" />
          </div>
        </el-tab-pane>

        <!-- ========== Tab 6: DNS/时间 ========== -->
        <el-tab-pane label="DNS/时间" name="dns-time">
          <div v-loading="dnsLoading" class="tab-content">
            <el-row :gutter="20">
              <!-- DNS 配置 -->
              <el-col :xs="24" :md="12">
                <el-card shadow="never" class="info-card">
                  <template #header>
                    <div class="card-header-flex">
                      <span class="card-title">DNS 配置</span>
                      <el-button type="primary" size="small" @click="handleSaveDNS">保存</el-button>
                    </div>
                  </template>
                  <el-form :model="dnsForm" label-width="100px" size="default">
                    <el-form-item label="DNS 服务器 1">
                      <el-input v-model="dnsForm.dns1" placeholder="例如: 8.8.8.8" />
                    </el-form-item>
                    <el-form-item label="DNS 服务器 2">
                      <el-input v-model="dnsForm.dns2" placeholder="例如: 8.8.4.4" />
                    </el-form-item>
                    <el-form-item label="DNS 服务器 3">
                      <el-input v-model="dnsForm.dns3" placeholder="可选" />
                    </el-form-item>
                    <el-form-item label="搜索域">
                      <el-input v-model="dnsForm.search" placeholder="例如: example.com" />
                    </el-form-item>
                  </el-form>
                </el-card>
              </el-col>

              <!-- 时间配置 -->
              <el-col :xs="24" :md="12">
                <el-card shadow="never" class="info-card">
                  <template #header>
                    <span class="card-title">时间配置</span>
                  </template>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item label="当前时间">
                      <span class="mono-font">{{ currentTimeStr }}</span>
                    </el-descriptions-item>
                    <el-descriptions-item label="时区">
                      {{ nodeTime?.timezone || '--' }}
                    </el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 网络接口创建/编辑对话框 -->
    <el-dialog
      v-model="networkDialogVisible"
      :title="isEditingNetwork ? '编辑网络接口' : '创建网络接口'"
      width="600px"
      destroy-on-close
    >
      <el-form ref="networkFormRef" :model="networkForm" label-width="100px" size="default">
        <el-form-item label="接口名称" prop="iface" :rules="[{ required: true, message: '请输入接口名称' }]">
          <el-input v-model="networkForm.iface" :disabled="isEditingNetwork" placeholder="例如: vmbr1" />
        </el-form-item>
        <el-form-item label="类型" prop="type" :rules="[{ required: true, message: '请选择类型' }]">
          <el-select v-model="networkForm.type" placeholder="选择接口类型">
            <el-option label="Bridge (网桥)" value="bridge" />
            <el-option label="Bond (绑定)" value="bond" />
            <el-option label="VLAN" value="vlan" />
            <el-option label="Ethernet (以太网)" value="eth" />
          </el-select>
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-input v-model="networkForm.address" placeholder="例如: 192.168.1.10" />
        </el-form-item>
        <el-form-item label="子网掩码">
          <el-input v-model="networkForm.netmask" placeholder="例如: 255.255.255.0" />
        </el-form-item>
        <el-form-item label="网关">
          <el-input v-model="networkForm.gateway" placeholder="例如: 192.168.1.1" />
        </el-form-item>
        <el-form-item label="自动启动">
          <el-switch v-model="networkForm.autostart" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="networkDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveNetwork">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  ArrowLeft,
  Refresh,
  RefreshRight,
  SwitchButton,
  Search,
  Plus,
  Check,
  Download,
} from '@element-plus/icons-vue'
import {
  getNodeStatus,
  getNodeVersion,
  getNodeServices,
  getNodeSyslog,
  getNetworkInterfaces,
  createNetworkInterface,
  updateNetworkInterface,
  deleteNetworkInterface,
  applyNetworkChanges,
  getAPTUpdate,
  updatePackages,
  getDNS,
  setDNS,
  getTime,
  restartService,
  startService,
  stopService,
} from '@/api/node'
import NetworkTopology from '@/components/node/NetworkTopology.vue'
import type {
  NodeStatus,
  NodeService,
  NodeSyslog,
  NetInterface,
  DNSConfig,
  APTUpdate,
} from '@/api/types'
import { formatBytes, formatUptime } from '@/utils/format'

const router = useRouter()
const route = useRoute()

// ============================================================
// 路由参数
// ============================================================

/** 当前节点名称 */
const nodeName = computed(() => (route.params.nodeName as string) || '')

// ============================================================
// Tab 状态
// ============================================================

/** 当前激活的 Tab */
const activeTab = ref('overview')

// ============================================================
// Tab 1: 概览
// ============================================================

/** 节点状态数据 */
const nodeStatus = ref<NodeStatus | null>(null)
/** 节点版本信息 */
const nodeVersion = ref<{ version: string; release: string; repoid: string } | null>(null)
/** 概览加载状态 */
const overviewLoading = ref(false)

/** CPU 使用百分比 */
const cpuPercent = computed(() => {
  if (!nodeStatus.value || !nodeStatus.value.maxcpu) return 0
  return Math.round((nodeStatus.value.cpu || 0) / nodeStatus.value.maxcpu * 1000) / 10
})

/** 内存使用百分比 */
const memPercent = computed(() => {
  if (!nodeStatus.value || !nodeStatus.value.maxmem) return 0
  return Math.round((nodeStatus.value.mem || 0) / nodeStatus.value.maxmem * 1000) / 10
})

/** 磁盘使用百分比 */
const diskPercent = computed(() => {
  if (!nodeStatus.value || !nodeStatus.value.maxdisk) return 0
  return Math.round((nodeStatus.value.disk || 0) / nodeStatus.value.maxdisk * 1000) / 10
})

/** Swap 使用百分比 */
const swapPercent = computed(() => {
  if (!nodeStatus.value || !nodeStatus.value.maxswap) return 0
  return Math.round((nodeStatus.value.swap || 0) / nodeStatus.value.maxswap * 1000) / 10
})

/**
 * 获取仪表盘颜色
 */
function getGaugeColor(percent: number): string {
  if (percent < 50) return '#52c41a'
  if (percent < 75) return '#faad14'
  return '#f5222d'
}

/**
 * 获取节点概览数据
 */
async function fetchOverview(): Promise<void> {
  overviewLoading.value = true
  try {
    const [status, version] = await Promise.all([
      getNodeStatus(nodeName.value),
      getNodeVersion(nodeName.value),
    ])
    nodeStatus.value = status
    nodeVersion.value = version
  } catch (error) {
    console.error('获取节点概览数据失败:', error)
  } finally {
    overviewLoading.value = false
  }
}

// ============================================================
// Tab 2: 网络
// ============================================================

/** 网络接口列表 */
const networkInterfaces = ref<NetInterface[]>([])
/** 网络加载状态 */
const networkLoading = ref(false)
/** 网络对话框可见性 */
const networkDialogVisible = ref(false)
/** 是否处于编辑模式 */
const isEditingNetwork = ref(false)
/** 网络表单引用 */
const networkFormRef = ref<FormInstance>()

/** 网络表单数据 */
const networkForm = ref({
  iface: '',
  type: 'bridge',
  address: '',
  netmask: '',
  gateway: '',
  autostart: true,
})

/** 检查是否有待处理的网络变更 */
const networkHasPending = ref(false)

/**
 * 获取网络接口列表
 */
async function fetchNetwork(): Promise<void> {
  networkLoading.value = true
  try {
    const interfaces = await getNetworkInterfaces(nodeName.value)
    networkInterfaces.value = interfaces
  } catch (error) {
    console.error('获取网络接口失败:', error)
  } finally {
    networkLoading.value = false
  }
}

/**
 * 获取网络掩码对应的前缀长度
 */
function getPrefixLength(netmask?: string): string {
  if (!netmask) return ''
  const maskParts = netmask.split('.')
  let bits = 0
  for (const part of maskParts) {
    const num = parseInt(part, 10)
    bits += num.toString(2).split('1').length - 1
  }
  return String(bits)
}

/**
 * 获取网络类型对应的标签类型
 */
function getNetTypeTag(type: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  const map: Record<string, 'success' | 'warning' | 'info' | 'primary' | 'danger'> = {
    bridge: 'primary',
    bond: 'success',
    vlan: 'warning',
    eth: 'info',
  }
  return map[type] || 'info'
}

/**
 * 显示创建网络对话框
 */
function showNetworkDialog(): void {
  isEditingNetwork.value = false
  networkForm.value = {
    iface: '',
    type: 'bridge',
    address: '',
    netmask: '',
    gateway: '',
    autostart: true,
  }
  networkDialogVisible.value = true
}

/**
 * 显示编辑网络对话框
 */
function editNetwork(iface: NetInterface): void {
  isEditingNetwork.value = true
  networkForm.value = {
    iface: iface.iface,
    type: iface.type || 'eth',
    address: iface.address || '',
    netmask: iface.netmask || '',
    gateway: iface.gateway || '',
    autostart: !!iface.autostart,
  }
  networkDialogVisible.value = true
}

/**
 * 保存网络配置
 */
async function handleSaveNetwork(): Promise<void> {
  try {
    const params: Record<string, unknown> = {
      iface: networkForm.value.iface,
      type: networkForm.value.type,
      autostart: networkForm.value.autostart ? 1 : 0,
    }
    if (networkForm.value.address) params.address = networkForm.value.address
    if (networkForm.value.netmask) params.netmask = networkForm.value.netmask
    if (networkForm.value.gateway) params.gateway = networkForm.value.gateway

    if (isEditingNetwork.value) {
      await updateNetworkInterface(nodeName.value, networkForm.value.iface, params)
      ElMessage.success('网络接口更新成功')
    } else {
      await createNetworkInterface(nodeName.value, params)
      ElMessage.success('网络接口创建成功')
    }
    networkDialogVisible.value = false
    await fetchNetwork()
  } catch (error) {
    console.error('保存网络配置失败:', error)
    ElMessage.error('保存网络配置失败')
  }
}

/**
 * 删除网络接口
 */
async function handleDeleteNetwork(iface: string): Promise<void> {
  try {
    await deleteNetworkInterface(nodeName.value, iface)
    ElMessage.success('网络接口删除成功')
    await fetchNetwork()
  } catch (error) {
    console.error('删除网络接口失败:', error)
    ElMessage.error('删除网络接口失败')
  }
}

/**
 * 应用网络变更
 */
async function handleApplyNetwork(): Promise<void> {
  try {
    await applyNetworkChanges(nodeName.value)
    ElMessage.success('网络配置已应用')
    await fetchNetwork()
  } catch (error) {
    console.error('应用网络变更失败:', error)
    ElMessage.error('应用网络变更失败')
  }
}

// ============================================================
// Tab 3: 系统日志
// ============================================================

/** 系统日志数据 */
const syslogEntries = ref<NodeSyslog[]>([])
/** 日志加载状态 */
const syslogLoading = ref(false)
/** 日志搜索关键词 */
const syslogSearch = ref('')
/** 日志优先级筛选 */
const syslogPriority = ref('')
/** 是否自动滚动 */
const syslogAutoScroll = ref(true)
/** 日志容器引用 */
const syslogContainerRef = ref<HTMLElement | null>(null)
/** 日志轮询定时器 */
let syslogTimer: ReturnType<typeof setInterval> | null = null

/** 过滤后的日志 */
const filteredSyslog = computed(() => {
  let result = syslogEntries.value

  // 按优先级筛选
  if (syslogPriority.value) {
    result = result.filter((e) => e.priority === syslogPriority.value)
  }

  // 按搜索关键词筛选
  if (syslogSearch.value.trim()) {
    const query = syslogSearch.value.trim().toLowerCase()
    result = result.filter((e) => e.message.toLowerCase().includes(query))
  }

  return result
})

/**
 * 获取系统日志
 */
async function fetchSyslog(): Promise<void> {
  syslogLoading.value = true
  try {
    const entries = await getNodeSyslog(nodeName.value, { limit: 200 })
    syslogEntries.value = entries
    // 自动滚动到底部
    if (syslogAutoScroll.value) {
      await nextTick()
      scrollToBottom()
    }
  } catch (error) {
    console.error('获取系统日志失败:', error)
  } finally {
    syslogLoading.value = false
  }
}

/**
 * 滚动日志容器到底部
 */
function scrollToBottom(): void {
  if (syslogContainerRef.value) {
    syslogContainerRef.value.scrollTop = syslogContainerRef.value.scrollHeight
  }
}

/**
 * 格式化时间戳
 */
function formatTimestamp(ts: number): string {
  const date = new Date(ts * 1000)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

/**
 * 获取优先级对应的标签类型
 */
function getPriorityTagType(priority: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    info: 'info',
    notice: 'info',
    warning: 'warning',
    err: 'danger',
    error: 'danger',
    crit: 'danger',
    alert: 'danger',
    emerg: 'danger',
  }
  return map[priority] || 'info'
}

// ============================================================
// Tab 4: 服务
// ============================================================

/** 服务列表 */
const services = ref<NodeService[]>([])
/** 服务加载状态 */
const servicesLoading = ref(false)

/**
 * 获取服务列表
 */
async function fetchServices(): Promise<void> {
  servicesLoading.value = true
  try {
    const result = await getNodeServices(nodeName.value)
    services.value = result
  } catch (error) {
    console.error('获取服务列表失败:', error)
  } finally {
    servicesLoading.value = false
  }
}

/**
 * 处理服务操作
 */
async function handleServiceAction(
  action: 'start' | 'stop' | 'restart',
  service: NodeService,
): Promise<void> {
  try {
    const actionMap = { start: startService, stop: stopService, restart: restartService }
    const actionFn = actionMap[action]
    await actionFn(nodeName.value, service.name)
    const actionTextMap = { start: '启动', stop: '停止', restart: '重启' }
    ElMessage.success(`${actionTextMap[action]}服务 ${service.name} 成功`)
    await fetchServices()
  } catch (error) {
    console.error(`${action}服务失败:`, error)
    ElMessage.error(`${action}服务失败`)
  }
}

// ============================================================
// Tab 5: 系统更新
// ============================================================

/** 可用更新列表 */
const updates = ref<APTUpdate[]>([])
/** 更新加载状态 */
const updatesLoading = ref(false)

/**
 * 获取可用更新
 */
async function fetchUpdates(): Promise<void> {
  updatesLoading.value = true
  try {
    const result = await getAPTUpdate(nodeName.value)
    updates.value = result
  } catch (error) {
    console.error('获取更新列表失败:', error)
  } finally {
    updatesLoading.value = false
  }
}

/**
 * 升级全部包
 */
async function handleUpgradeAll(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定要升级全部 ${updates.value.length} 个软件包吗？此操作可能需要较长时间。`,
      '升级全部',
      { type: 'warning' },
    )
    const packageNames = updates.value.map((u) => u.Package)
    await updatePackages(nodeName.value, packageNames)
    ElMessage.success('升级任务已提交')
    await fetchUpdates()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('升级全部失败:', error)
      ElMessage.error('升级全部失败')
    }
  }
}

/**
 * 升级单个包
 */
async function handleUpgradePackage(packageName: string): Promise<void> {
  try {
    await updatePackages(nodeName.value, [packageName])
    ElMessage.success(`升级任务已提交: ${packageName}`)
    await fetchUpdates()
  } catch (error) {
    console.error('升级软件包失败:', error)
    ElMessage.error('升级软件包失败')
  }
}

// ============================================================
// Tab 6: DNS/时间
// ============================================================

/** DNS 表单 */
const dnsForm = ref<DNSConfig>({})
/** 节点时间信息 */
const nodeTime = ref<{ time: number; timezone: string } | null>(null)
/** DNS/时间加载状态 */
const dnsLoading = ref(false)

/** 当前时间字符串（实时更新） */
const currentTimeStr = ref('')
/** 时间刷新定时器 */
let timeTimer: ReturnType<typeof setInterval> | null = null

/**
 * 获取 DNS 配置
 */
async function fetchDNS(): Promise<void> {
  try {
    const config = await getDNS(nodeName.value)
    dnsForm.value = { ...config }
  } catch (error) {
    console.error('获取 DNS 配置失败:', error)
  }
}

/**
 * 获取时间信息
 */
async function fetchTime(): Promise<void> {
  try {
    const timeInfo = await getTime(nodeName.value)
    nodeTime.value = timeInfo
  } catch (error) {
    console.error('获取时间信息失败:', error)
  }
}

/**
 * 保存 DNS 配置
 */
async function handleSaveDNS(): Promise<void> {
  try {
    await setDNS(nodeName.value, dnsForm.value)
    ElMessage.success('DNS 配置保存成功')
  } catch (error) {
    console.error('保存 DNS 配置失败:', error)
    ElMessage.error('保存 DNS 配置失败')
  }
}

/**
 * 更新当前时间显示
 */
function updateTimeDisplay(): void {
  if (nodeTime.value) {
    const date = new Date(nodeTime.value.time * 1000)
    currentTimeStr.value = date.toLocaleString('zh-CN')
  }
}

// ============================================================
// 节点操作
// ============================================================

/**
 * 重启节点
 */
async function handleReboot(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定要重启节点 ${nodeName.value} 吗？这将中断所有正在运行的服务。`,
      '重启节点',
      { type: 'warning' },
    )
    ElMessage.success('重启命令已发送')
    // TODO: 调用重启 API
  } catch {
    // 用户取消
  }
}

/**
 * 关闭节点
 */
async function handleShutdown(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定要关闭节点 ${nodeName.value} 吗？这将关闭所有虚拟机和服务。`,
      '关闭节点',
      { type: 'error' },
    )
    ElMessage.warning('关机命令已发送')
    // TODO: 调用关机 API
  } catch {
    // 用户取消
  }
}

// ============================================================
// 导航
// ============================================================

/**
 * 返回节点列表
 */
function goBack(): void {
  router.push({ name: 'NodeList' })
}

// ============================================================
// Tab 切换监听
// ============================================================

watch(activeTab, (tab) => {
  switch (tab) {
    case 'network':
      fetchNetwork()
      break
    case 'syslog':
      fetchSyslog()
      break
    case 'services':
      fetchServices()
      break
    case 'updates':
      fetchUpdates()
      break
    case 'dns-time':
      dnsLoading.value = true
      Promise.all([fetchDNS(), fetchTime()]).finally(() => {
        dnsLoading.value = false
      })
      break
  }
})

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  fetchOverview()

  // 启动时间显示更新
  timeTimer = setInterval(() => {
    updateTimeDisplay()
  }, 1000)

  // 启动日志轮询
  syslogTimer = setInterval(() => {
    if (activeTab.value === 'syslog') {
      fetchSyslog()
    }
  }, 10_000)
})

onUnmounted(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
    timeTimer = null
  }
  if (syslogTimer) {
    clearInterval(syslogTimer)
    syslogTimer = null
  }
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.node-detail-page {
  padding: $spacing-6;
  min-height: 100%;
  overflow: auto;
}

// 页面头部
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-6;
  gap: $spacing-4;

  @media (max-width: $breakpoint-sm) {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: $spacing-4;

    .page-title {
      font-size: $font-size-2xl;
      font-weight: $font-weight-bold;
      color: $color-text-primary;
      margin: 0;
    }
  }

  .header-right {
    display: flex;
    gap: $spacing-3;
  }
}

// 通用工具栏
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-4;

  .toolbar-left {
    display: flex;
    align-items: center;
  }

  .toolbar-right {
    display: flex;
    align-items: center;
  }

  .updates-count {
    font-size: $font-size-sm;
    color: $color-text-secondary;
  }
}

// Tab 内容
.tab-content {
  padding: $spacing-4 0;
}

// 信息卡片
.info-card {
  margin-bottom: $spacing-5;

  .card-title {
    font-size: $font-size-base;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
  }

  .card-header-flex {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}

// 仪表盘行
.gauge-row {
  margin-top: $spacing-5;
}

.gauge-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: $spacing-5;
  background: $color-bg-container;
  border-radius: $radius-base;
  box-shadow: $shadow-card;

  .gauge-title {
    font-size: $font-size-sm;
    font-weight: $font-weight-semibold;
    color: $color-text-secondary;
    margin-bottom: $spacing-3;
  }

  .gauge-value {
    font-size: $font-size-lg;
    font-weight: $font-weight-bold;
    color: $color-text-primary;
    margin-top: $spacing-3;
  }
}

// 网络表格
.network-table {
  margin-bottom: $spacing-5;
}

.topology-section {
  margin-top: $spacing-5;
}

.text-muted {
  color: $color-text-placeholder;
}

// 日志
.syslog-content {
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.log-container {
  flex: 1;
  background: $gray-11;
  border-radius: $radius-sm;
  padding: $spacing-3;
  overflow-y: auto;
  max-height: 600px;
  min-height: 300px;
  font-family: $font-family-code;
  font-size: $font-size-xs;
  line-height: 1.8;
}

.log-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  gap: $spacing-2;
  padding: 2px $spacing-2;
  border-radius: $radius-xs;
  transition: $transition-fast;

  &:hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .log-time {
    color: $gray-5;
    flex-shrink: 0;
    min-width: 130px;
  }

  .log-priority {
    flex-shrink: 0;
    min-width: 60px;
    text-align: center;
  }

  .log-message {
    color: $gray-2;
    word-break: break-all;
  }

  &.log-warning {
    .log-message {
      color: $orange-400;
    }
  }

  &.log-error,
  &.log-err,
  &.log-crit,
  &.log-alert,
  &.log-emerg {
    .log-message {
      color: $red-400;
    }
  }
}

// 服务表格操作列
:deep(.el-table__row) {
  .el-button + .el-button {
    margin-left: $spacing-2;
  }
}

// 更新表格
.updates-table {
  margin-bottom: $spacing-4;
}

// DNS/时间
.mono-font {
  font-family: $font-family-code;
}

// 描述列表样式
:deep(.el-descriptions) {
  .el-descriptions__label {
    font-weight: $font-weight-medium;
    color: $color-text-secondary;
    width: 120px;
  }

  .el-descriptions__content {
    color: $color-text-primary;
    font-weight: $font-weight-medium;
  }
}

// Tab 页签样式覆盖
:deep(.el-tabs--border-card) {
  border: none;
  box-shadow: none;

  > .el-tabs__header {
    background: transparent;
    border-bottom: 1px solid $color-border-lighter;
  }
}
</style>
