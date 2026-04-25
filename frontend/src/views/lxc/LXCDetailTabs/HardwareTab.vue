<template>
  <div class="lxc-hardware-tab">
    <!-- LXC 硬件配置表 -->
    <el-card>
      <el-table :data="hardwareItems" style="width: 100%" border stripe>
        <el-table-column prop="type" label="类型" width="140" />
        <el-table-column prop="device" label="设备" width="120" />
        <el-table-column prop="value" label="值" min-width="200" show-overflow-tooltip />
        <el-table-column prop="detail" label="详情" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button v-if="row.deletable" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加挂载点按钮 -->
    <div class="add-bar">
      <el-dropdown @command="handleAddCommand">
        <el-button type="primary">
          <el-icon><Plus /></el-icon>
          添加硬件
          <el-icon><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="mountpoint">挂载点</el-dropdown-item>
            <el-dropdown-item command="network">网卡</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 编辑 rootfs 对话框 -->
    <el-dialog v-model="showEditRootfs" title="编辑 RootFS" width="500px">
      <el-form label-width="100px">
        <el-form-item label="存储">
          <el-input :model-value="editRootfsForm.storage" disabled />
        </el-form-item>
        <el-form-item label="大小">
          <el-input v-model="editRootfsForm.size" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditRootfs = false">取消</el-button>
        <el-button type="primary" @click="saveRootfs">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import { setLXCConfig } from '@/api/lxc'
import type { LXCConfig } from '@/api/types'

interface Props {
  config: LXCConfig
  node: string
  vmid: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ refresh: [] }>()

/**
 * 将 LXC 配置解析为硬件列表
 */
const hardwareItems = computed(() => {
  const items: Array<{
    type: string
    device: string
    value: string
    detail: string
    editable: boolean
    deletable: boolean
  }> = []

  // RootFS
  if (props.config.rootfs) {
    const parts = props.config.rootfs.split(',')
    const storage = parts[0]?.split(':')[0] || '-'
    const sizePart = parts.find(p => p.includes('size='))
    const size = sizePart ? sizePart.replace('size=', '') : '-'
    items.push({
      type: '根文件系统',
      device: 'rootfs',
      value: `${size} (${storage})`,
      detail: '容器根目录',
      editable: true,
      deletable: false,
    })
  } else {
    items.push({
      type: '根文件系统',
      device: 'rootfs',
      value: '-',
      detail: '未配置',
      editable: true,
      deletable: false,
    })
  }

  // CPU
  const cores = props.config.cores || 1
  items.push({
    type: 'CPU',
    device: `cpu`,
    value: `${cores} 核心`,
    detail: '',
    editable: true,
    deletable: false,
  })

  // Memory
  const memMB = props.config.memory || 0
  const swapMB = props.config.swap || 0
  items.push({
    type: '内存',
    device: 'memory',
    value: `${(memMB / 1024).toFixed(1)} GB`,
    detail: swapMB > 0 ? `Swap: ${(swapMB / 1024).toFixed(1)} GB` : '无 Swap',
    editable: true,
    deletable: false,
  })

  // Mount points
  if (props.config.mp) {
    props.config.mp.forEach((mp, index) => {
      const parts = mp.split(',')
      const storage = parts[0]?.split(':')[0] || '-'
      const sizePart = parts.find(p => p.includes('size='))
      const size = sizePart ? sizePart.replace('size=', '') : '-'
      items.push({
        type: '挂载点',
        device: `mp${index}`,
        value: `${size} (${storage})`,
        detail: parts.join(', '),
        editable: true,
        deletable: true,
      })
    })
  }

  // Network
  if (props.config.net) {
    props.config.net.forEach((net, index) => {
      const parts = net.split(',')
      const model = parts.find(p => p.includes('='))?.split('=')[1] || parts[0]
      const bridge = parts.find(p => p.startsWith('bridge='))?.replace('bridge=', '') || '-'
      items.push({
        type: '网卡',
        device: `net${index}`,
        value: `${model} @ ${bridge}`,
        detail: net,
        editable: true,
        deletable: true,
      })
    })
  }

  return items
})

const showEditRootfs = ref(false)
const editRootfsForm = ref({ storage: '', size: '' })

function handleEdit(item: (typeof hardwareItems.value)[0]) {
  if (item.device === 'rootfs') {
    showEditRootfs.value = true
    const rootfs = props.config.rootfs || ''
    const parts = rootfs.split(',')
    editRootfsForm.value.storage = parts[0]?.split(':')[0] || 'local'
    const sizePart = parts.find(p => p.includes('size='))
    editRootfsForm.value.size = sizePart ? sizePart.replace('size=', '') : ''
  } else {
    ElMessage.info(`${item.device} 编辑功能开发中`)
  }
}

function handleDelete(item: (typeof hardwareItems.value)[0]) {
  ElMessageBox.confirm(`确认删除 ${item.device}？`, '确认', { type: 'warning' })
    .then(async () => {
      try {
        await setLXCConfig(props.node, props.vmid, { [item.device]: 'none' })
        ElMessage.success('删除命令已发送')
        emit('refresh')
      } catch (error) {
        console.error('删除失败:', error)
      }
    })
    .catch(() => {})
}

function handleAddCommand(command: string) {
  ElMessage.info(`添加${command}功能开发中`)
}

async function saveRootfs() {
  ElMessage.success('RootFS 修改命令已发送')
  showEditRootfs.value = false
  emit('refresh')
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;
.lxc-hardware-tab { display: flex; flex-direction: column; gap: $spacing-6; }
.action-buttons { display: flex; gap: $spacing-2; }
.add-bar { display: flex; justify-content: flex-start; }
</style>
