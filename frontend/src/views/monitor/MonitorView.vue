<template>
  <div class="monitor-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>监控中心</h2>
        <span class="header-subtitle">集群资源监控与告警管理</span>
      </div>
      <div class="header-right">
        <el-button :icon="Refresh" :loading="loading" @click="refreshAll">刷新</el-button>
      </div>
    </div>

    <!-- 控制栏 -->
    <el-card class="control-bar" shadow="never">
      <div class="control-row">
        <!-- 节点选择 -->
        <div class="control-item">
          <span class="control-label">节点：</span>
          <el-select v-model="selectedNode" placeholder="选择节点" style="width: 200px" @change="onNodeChange">
            <el-option label="全部节点" value="all" />
            <el-option v-for="node in nodeList" :key="node.name" :label="node.name" :value="node.name" />
          </el-select>
        </div>

        <!-- 时间范围 -->
        <div class="control-item">
          <span class="control-label">时间范围：</span>
          <el-button-group>
            <el-button
              v-for="tf in timeframes"
              :key="tf.value"
              :type="timeframe === tf.value ? 'primary' : ''"
              size="small"
              @click="timeframe = tf.value"
            >
              {{ tf.label }}
            </el-button>
          </el-button-group>
        </div>
      </div>
    </el-card>

    <!-- 图表区域 -->
    <div class="chart-grid">
      <!-- CPU 使用率 -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>CPU 使用率</span>
          </div>
        </template>
        <RRDChart
          v-if="activeNode"
          :key="`cpu-${activeNode}-${timeframe}`"
          :node="activeNode"
          :timeframe="timeframe"
          dataset="cpu"
          chart-type="line"
          height="280px"
          :yAxisMax="100"
          unit="%"
        />
        <el-empty v-else description="请选择节点" :image-size="80" />
      </el-card>

      <!-- 内存使用 -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>内存使用</span>
          </div>
        </template>
        <RRDChart
          v-if="activeNode"
          :key="`mem-${activeNode}-${timeframe}`"
          :node="activeNode"
          :timeframe="timeframe"
          dataset="mem"
          chart-type="area"
          height="280px"
          unit="MB"
        />
        <el-empty v-else description="请选择节点" :image-size="80" />
      </el-card>

      <!-- 磁盘 I/O -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>磁盘 I/O</span>
          </div>
        </template>
        <RRDChart
          v-if="activeNode"
          :key="`disk-${activeNode}-${timeframe}`"
          :node="activeNode"
          :timeframe="timeframe"
          dataset="disk"
          chart-type="bar"
          height="280px"
          unit="B/s"
        />
        <el-empty v-else description="请选择节点" :image-size="80" />
      </el-card>

      <!-- 网络 I/O -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>网络 I/O</span>
          </div>
        </template>
        <RRDChart
          v-if="activeNode"
          :key="`net-${activeNode}-${timeframe}`"
          :node="activeNode"
          :timeframe="timeframe"
          dataset="net"
          chart-type="area"
          height="280px"
          unit="B/s"
        />
        <el-empty v-else description="请选择节点" :image-size="80" />
      </el-card>
    </div>

    <!-- 资源对比表 -->
    <el-card class="comparison-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>节点资源对比</span>
        </div>
      </template>
      <el-table :data="comparisonData" stripe border v-loading="loading">
        <el-table-column prop="name" label="节点名称" width="150" fixed />
        <el-table-column label="CPU 使用率" width="180">
          <template #default="{ row }">
            <el-progress
              :percentage="row.cpuUsage"
              :color="getProgressColor(row.cpuUsage)"
              :stroke-width="14"
              :format="() => row.cpuUsage + '%'"
            />
          </template>
        </el-table-column>
        <el-table-column label="内存使用" width="200">
          <template #default="{ row }">
            <el-progress
              :percentage="row.memoryUsage"
              :color="getProgressColor(row.memoryUsage)"
              :stroke-width="14"
              :format="() => row.memoryUsage + '%'"
            />
          </template>
        </el-table-column>
        <el-table-column label="磁盘使用" width="200">
          <template #default="{ row }">
            <el-progress
              :percentage="row.diskUsage"
              :color="getProgressColor(row.diskUsage)"
              :stroke-width="14"
              :format="() => row.diskUsage + '%'"
            />
          </template>
        </el-table-column>
        <el-table-column label="虚拟机数" width="100" align="center">
          <template #default="{ row }">{{ row.vmCount }}</template>
        </el-table-column>
        <el-table-column label="容器数" width="100" align="center">
          <template #default="{ row }">{{ row.ctCount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 告警规则 -->
    <el-card class="alert-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>告警规则</span>
          <el-button type="primary" size="small" :icon="Plus" @click="handleAddRule">添加规则</el-button>
        </div>
      </template>

      <el-table :data="alertRules" stripe border>
        <el-table-column prop="name" label="规则名称" width="160" />
        <el-table-column label="监控指标" width="120">
          <template #default="{ row }">{{ metricLabel(row.metric) }}</template>
        </el-table-column>
        <el-table-column label="条件" width="100">
          <template #default="{ row }">{{ conditionLabel(row.condition) }}</template>
        </el-table-column>
        <el-table-column label="阈值" width="100">
          <template #default="{ row }">
            {{ row.threshold }}{{ row.metric === 'node_status' ? '' : '%' }}
          </template>
        </el-table-column>
        <el-table-column label="持续时间" width="100">
          <template #default="{ row }">{{ row.duration }} 分钟</template>
        </el-table-column>
        <el-table-column label="通知方式" width="150">
          <template #default="{ row }">
            <el-tag size="small" :type="row.notifyType === 'email' ? '' : 'warning'">
              {{ row.notifyType === 'email' ? '邮件' : 'Webhook' }}
            </el-tag>
            <span class="notify-target">{{ row.notifyTarget }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="handleToggleRule(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEditRule(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="alertRules.length === 0" description="暂无告警规则" :image-size="80" />
    </el-card>

    <!-- 告警规则对话框 -->
    <AlertRuleDialog
      v-model:visible="ruleDialogVisible"
      :rule="editingRule"
      @submit="handleRuleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * MonitorView - 监控中心页面
 * 
 * 展示集群级资源监控图表（CPU、内存、磁盘 I/O、网络 I/O），
 * 节点资源对比表，以及告警规则的增删改查。
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RRDChart from '@/components/monitor/RRDChart.vue'
import AlertRuleDialog, { type AlertRule } from '@/components/monitor/AlertRuleDialog.vue'
import { getClusterResources } from '@/api/cluster'
import type { ClusterResource } from '@/api/types'
import { useResourceStore } from '@/stores/resources'

// ============================================================
// 状态管理
// ============================================================

const resourceStore = useResourceStore()

const loading = ref(false)
const selectedNode = ref<string>('all')
const timeframe = ref<'hour' | 'day' | 'week' | 'month' | 'year'>('hour')
const clusterResources = ref<ClusterResource[]>([])

// 时间范围选项
const timeframes = [
  { label: '1小时', value: 'hour' as const },
  { label: '1天', value: 'day' as const },
  { label: '1周', value: 'week' as const },
  { label: '1月', value: 'month' as const },
  { label: '1年', value: 'year' as const },
]

// 告警规则
const alertRules = ref<AlertRule[]>([])
const ruleDialogVisible = ref(false)
const editingRule = ref<AlertRule | null>(null)

// ============================================================
// 计算属性
// ============================================================

/** 节点列表 */
const nodeList = computed(() => {
  return clusterResources.value
    .filter(r => r.type === 'node')
    .map(r => ({
      name: r.node || r.id,
      status: r.status,
      cpu: r.cpu ? Math.round(r.cpu * 1000) / 10 : 0,
      mem: r.mem || 0,
      maxmem: r.maxmem || 0,
      disk: r.disk || 0,
      maxdisk: r.maxdisk || 0,
    }))
})

/** 当前活动的节点（用于图表） */
const activeNode = computed(() => {
  if (selectedNode.value === 'all') {
    // 默认展示第一个在线节点
    const onlineNode = nodeList.value.find(n => n.status === 'online')
    return onlineNode?.name || ''
  }
  return selectedNode.value
})

/** 资源对比表数据 */
const comparisonData = computed(() => {
  const nodes = clusterResources.value.filter(r => r.type === 'node')
  const vms = clusterResources.value.filter(r => r.type === 'vm')
  const lxc = clusterResources.value.filter(r => r.type === 'lxc')

  return nodes.map(node => {
    const nodeName = node.node || node.id
    const vmCount = vms.filter(vm => vm.node === nodeName).length
    const ctCount = lxc.filter(ct => ct.node === nodeName).length
    const cpuUsage = node.cpu ? Math.round(node.cpu * 1000) / 10 : 0
    const memoryUsage = node.maxmem ? Math.round(((node.mem || 0) / node.maxmem) * 1000) / 10 : 0
    const diskUsage = node.maxdisk ? Math.round(((node.disk || 0) / node.maxdisk) * 1000) / 10 : 0

    return {
      name: nodeName,
      status: node.status,
      cpuUsage,
      memoryUsage,
      diskUsage,
      vmCount,
      ctCount,
    }
  })
})

// ============================================================
// 数据加载
// ============================================================

/** 加载集群资源 */
async function loadResources() {
  loading.value = true
  try {
    const resources = await getClusterResources()
    clusterResources.value = resources
  } catch (error) {
    console.error('获取集群资源失败:', error)
  } finally {
    loading.value = false
  }
}

/** 刷新所有数据 */
async function refreshAll() {
  await Promise.all([loadResources()])
  ElMessage.success('刷新成功')
}

// ============================================================
// 告警规则管理
// ============================================================

/** 指标标签映射 */
function metricLabel(metric: string): string {
  const map: Record<string, string> = {
    cpu: 'CPU 使用率',
    memory: '内存使用率',
    disk: '磁盘使用率',
    network: '网络流量',
    node_status: '节点状态',
  }
  return map[metric] || metric
}

/** 条件标签映射 */
function conditionLabel(condition: string): string {
  const map: Record<string, string> = {
    greater_than: '大于',
    less_than: '小于',
    equal: '等于',
    not_equal: '不等于',
  }
  return map[condition] || condition
}

/** 添加规则 */
function handleAddRule() {
  editingRule.value = null
  ruleDialogVisible.value = true
}

/** 编辑规则 */
function handleEditRule(rule: AlertRule) {
  editingRule.value = { ...rule }
  ruleDialogVisible.value = true
}

/** 删除规则 */
async function handleDeleteRule(rule: AlertRule) {
  try {
    await ElMessageBox.confirm(`确定要删除告警规则"${rule.name}"吗？`, '确认删除', {
      type: 'warning',
    })
    alertRules.value = alertRules.value.filter(r => r.id !== rule.id)
    ElMessage.success('删除成功')
  } catch {
    // 用户取消
  }
}

/** 切换规则状态 */
function handleToggleRule(rule: AlertRule) {
  ElMessage.success(`规则"${rule.name}"已${rule.enabled ? '启用' : '禁用'}`)
}

/** 提交规则 */
function handleRuleSubmit(ruleData: Omit<AlertRule, 'id'> & { id?: string }) {
  if (ruleData.id) {
    // 编辑模式
    const index = alertRules.value.findIndex(r => r.id === ruleData.id)
    if (index !== -1) {
      alertRules.value[index] = { ...ruleData, id: ruleData.id } as AlertRule
    }
    ElMessage.success('规则已更新')
  } else {
    // 新建模式
    const newRule: AlertRule = {
      ...ruleData,
      id: Date.now().toString(),
    }
    alertRules.value.push(newRule)
    ElMessage.success('规则已创建')
  }
}

/** 进度条颜色 */
function getProgressColor(value: number): string {
  if (value < 60) return '#67C23A'
  if (value < 80) return '#E6A23C'
  return '#F56C6C'
}

/** 节点切换时刷新资源 */
function onNodeChange() {
  // 节点切换时，RRDChart 组件会通过 key 重新加载数据
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  loadResources()
})

onUnmounted(() => {
  // 清理
})
</script>

<style scoped lang="scss">
.monitor-view {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .header-left {
    h2 {
      margin: 0;
      font-size: 20px;
      font-weight: 600;
      color: #262626;
    }

    .header-subtitle {
      font-size: 13px;
      color: #8c8c8c;
      margin-top: 4px;
    }
  }
}

.control-bar {
  margin-bottom: 16px;

  :deep(.el-card__body) {
    padding: 12px 16px;
  }

  .control-row {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    align-items: center;
  }

  .control-item {
    display: flex;
    align-items: center;
    gap: 8px;

    .control-label {
      font-size: 14px;
      color: #595959;
      white-space: nowrap;
    }
  }
}

.chart-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.chart-card {
  :deep(.el-card__header) {
    padding: 12px 16px;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.el-card__body) {
    padding: 16px;
  }
}

.comparison-card,
.alert-card {
  margin-bottom: 16px;

  :deep(.el-card__header) {
    padding: 12px 16px;
    border-bottom: 1px solid #f0f0f0;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 15px;
  font-weight: 600;
  color: #262626;
}

.notify-target {
  margin-left: 8px;
  font-size: 12px;
  color: #8c8c8c;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: middle;
}

@media (max-width: 1200px) {
  .chart-grid {
    grid-template-columns: 1fr;
  }
}
</style>
