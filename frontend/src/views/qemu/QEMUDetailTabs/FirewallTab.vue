<template>
  <div class="firewall-tab">
    <!-- 防火墙全局开关 -->
    <el-card>
      <template #header>
        <span class="card-title">防火墙</span>
      </template>
      <el-form label-width="120px">
        <el-form-item label="启用防火墙">
          <el-switch v-model="firewallEnabled" @change="handleToggleFirewall" />
          <span class="form-hint">启用后所有流量将被防火墙规则过滤</span>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 规则列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="card-title">防火墙规则</span>
          <el-button type="primary" size="small" @click="showAddRule = true">
            <el-icon><Plus /></el-icon>
            添加规则
          </el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="rules" style="width: 100%" border stripe>
        <el-table-column prop="pos" label="#" width="60" />
        <el-table-column prop="action" label="动作" width="90">
          <template #default="{ row }">
            <el-tag :type="actionType(row.action)" size="small">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="90" />
        <el-table-column prop="source" label="源地址" min-width="140" />
        <el-table-column prop="dest" label="目标地址" min-width="140" />
        <el-table-column prop="proto" label="协议" width="80" />
        <el-table-column prop="dport" label="目标端口" width="100" />
        <el-table-column prop="enable" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enable ? 'success' : 'info'" size="small">
              {{ row.enable ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleEditRule(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDeleteRule(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && rules.length === 0" description="暂无防火墙规则" />
    </el-card>

    <!-- IPSet 管理 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="card-title">IPSet (IP 集合)</span>
          <el-button type="primary" size="small">
            <el-icon><Plus /></el-icon>
            创建 IPSet
          </el-button>
        </div>
      </template>
      <el-empty description="暂无 IPSet" />
    </el-card>

    <!-- 添加/编辑规则对话框 -->
    <el-dialog v-model="showAddRule" :title="editingRule ? '编辑规则' : '添加规则'" width="560px">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="动作" required>
          <el-select v-model="ruleForm.action" style="width: 100%">
            <el-option label="ACCEPT (放行)" value="ACCEPT" />
            <el-option label="DROP (丢弃)" value="DROP" />
            <el-option label="REJECT (拒绝)" value="REJECT" />
          </el-select>
        </el-form-item>
        <el-form-item label="方向">
          <el-select v-model="ruleForm.type" style="width: 100%">
            <el-option label="IN (入站)" value="in" />
            <el-option label="OUT (出站)" value="out" />
          </el-select>
        </el-form-item>
        <el-form-item label="源地址">
          <el-input v-model="ruleForm.source" placeholder="留空表示任意 / 可指定 CIDR" />
        </el-form-item>
        <el-form-item label="目标地址">
          <el-input v-model="ruleForm.dest" placeholder="留空表示任意" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="ruleForm.proto" style="width: 100%">
            <el-option label="TCP" value="tcp" />
            <el-option label="UDP" value="udp" />
            <el-option label="ICMP" value="icmp" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标端口">
          <el-input v-model="ruleForm.dport" placeholder="例如: 80 或 80:443" />
        </el-form-item>
        <el-form-item label="注释">
          <el-input v-model="ruleForm.comment" placeholder="可选注释" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enable" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddRule = false">取消</el-button>
        <el-button type="primary" @click="confirmSaveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

// 防火墙占位数据 - 实际应通过 PVE firewall API 获取
interface FirewallRule {
  pos: number
  action: 'ACCEPT' | 'DROP' | 'REJECT'
  type: 'in' | 'out'
  source?: string
  dest?: string
  proto?: string
  dport?: string
  enable: boolean
  comment?: string
}

defineProps<{
  node: string
  vmid: number
}>()

const loading = ref(false)
const firewallEnabled = ref(false)
const showAddRule = ref(false)
const editingRule = ref<FirewallRule | null>(null)

// 示例规则
const rules = ref<FirewallRule[]>([
  {
    pos: 0,
    action: 'ACCEPT',
    type: 'in',
    source: '10.0.0.0/8',
    proto: 'tcp',
    dport: '22',
    enable: true,
    comment: '允许内网 SSH',
  },
  {
    pos: 1,
    action: 'ACCEPT',
    type: 'in',
    proto: 'tcp',
    dport: '80',
    enable: true,
    comment: '允许 HTTP',
  },
])

const ruleForm = ref({
  action: 'ACCEPT' as const,
  type: 'in' as const,
  source: '',
  dest: '',
  proto: 'tcp',
  dport: '',
  comment: '',
  enable: true,
})

function actionType(action: string): string {
  switch (action) {
    case 'ACCEPT': return 'success'
    case 'DROP': return 'danger'
    case 'REJECT': return 'warning'
    default: return ''
  }
}

function handleToggleFirewall(val: boolean) {
  ElMessage.info(`防火墙${val ? '启用' : '禁用'}命令已发送`)
}

function handleEditRule(row: FirewallRule) {
  editingRule.value = row
  ruleForm.value = {
    action: row.action,
    type: row.type,
    source: row.source || '',
    dest: row.dest || '',
    proto: row.proto || 'tcp',
    dport: row.dport || '',
    comment: row.comment || '',
    enable: row.enable,
  }
  showAddRule.value = true
}

function handleDeleteRule(row: FirewallRule) {
  ElMessageBox.confirm('确认删除该规则？', '确认', { type: 'warning' })
    .then(() => {
      ElMessage.info('删除规则功能需要后端支持')
    })
    .catch(() => {})
}

function confirmSaveRule() {
  ElMessage.success('规则保存命令已发送')
  showAddRule.value = false
  editingRule.value = null
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.firewall-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-6;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
}

.form-hint {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  margin-left: $spacing-3;
}

.action-buttons {
  display: flex;
  gap: $spacing-2;
}
</style>
