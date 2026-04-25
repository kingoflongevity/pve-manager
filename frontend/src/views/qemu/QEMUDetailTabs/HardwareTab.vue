<template>
  <div class="hardware-tab">
    <el-alert
      v-if="hasPendingConfig"
      title="存在待生效配置"
      description="部分配置更改将在下次重启后生效"
      type="warning"
      show-icon
      :closable="false"
      class="pending-alert"
    />

    <!-- 硬件配置列表 -->
    <el-card>
      <el-table :data="hardwareItems" style="width: 100%" border stripe>
        <el-table-column prop="type" label="类型" width="140" />
        <el-table-column prop="device" label="设备" width="100" />
        <el-table-column prop="value" label="值" min-width="200" show-overflow-tooltip />
        <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button
                v-if="row.deletable"
                link
                type="danger"
                size="small"
                @click="handleDelete(row)"
              >删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加硬件按钮 -->
    <div class="add-hardware-bar">
      <el-dropdown @command="handleAddCommand">
        <el-button type="primary">
          <el-icon><Plus /></el-icon>
          添加硬件
          <el-icon><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="disk">硬盘</el-dropdown-item>
            <el-dropdown-item command="network">网卡</el-dropdown-item>
            <el-dropdown-item command="cdrom">光驱</el-dropdown-item>
            <el-dropdown-item command="usb">USB 设备</el-dropdown-item>
            <el-dropdown-item command="memory">内存</el-dropdown-item>
            <el-dropdown-item command="cpu">CPU</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 磁盘调整大小对话框 -->
    <el-dialog v-model="showResizeDialog" title="调整磁盘大小" width="480px">
      <el-form :model="resizeForm" label-width="100px">
        <el-form-item label="磁盘">
          <el-input :model-value="resizeForm.disk" disabled />
        </el-form-item>
        <el-form-item label="当前大小">
          <el-input :model-value="resizeForm.currentSize" disabled />
        </el-form-item>
        <el-form-item label="扩容大小">
          <el-input v-model="resizeForm.expandSize" placeholder="例如: +10G">
            <template #append>GB</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResizeDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmResize">确认</el-button>
      </template>
    </el-dialog>

    <!-- 添加磁盘对话框 -->
    <el-dialog v-model="showAddDiskDialog" title="添加硬盘" width="520px">
      <el-form :model="addDiskForm" label-width="100px">
        <el-form-item label="存储">
          <el-select v-model="addDiskForm.storage" placeholder="选择存储" style="width: 100%">
            <el-option label="local" value="local" />
            <el-option label="local-lvm" value="local-lvm" />
            <el-option label="nfs-storage" value="nfs-storage" />
          </el-select>
        </el-form-item>
        <el-form-item label="大小">
          <el-input v-model="addDiskForm.size" placeholder="例如: 50">
            <template #append>GB</template>
          </el-input>
        </el-form-item>
        <el-form-item label="总线">
          <el-select v-model="addDiskForm.bus" style="width: 100%">
            <el-option label="SCSI" value="scsi" />
            <el-option label="VirtIO Block" value="virtio" />
            <el-option label="SATA" value="sata" />
            <el-option label="IDE" value="ide" />
          </el-select>
        </el-form-item>
        <el-form-item label="格式">
          <el-select v-model="addDiskForm.format" style="width: 100%">
            <el-option label="Raw" value="raw" />
            <el-option label="Qcow2" value="qcow2" />
          </el-select>
        </el-form-item>
        <el-form-item label="SSD 模拟">
          <el-switch v-model="addDiskForm.ssd" />
        </el-form-item>
        <el-form-item label="Discard">
          <el-switch v-model="addDiskForm.discard" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDiskDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddDisk">添加</el-button>
      </template>
    </el-dialog>

    <!-- 添加网卡对话框 -->
    <el-dialog v-model="showAddNetworkDialog" title="添加网卡" width="520px">
      <el-form :model="addNetworkForm" label-width="100px">
        <el-form-item label="模型">
          <el-select v-model="addNetworkForm.model" style="width: 100%">
            <el-option label="VirtIO (paravirtualized)" value="virtio" />
            <el-option label="Intel e1000" value="e1000" />
            <el-option label="Realtek RTL8139" value="rtl8139" />
            <el-option label="VMware vmxnet3" value="vmxnet3" />
          </el-select>
        </el-form-item>
        <el-form-item label="桥接">
          <el-select v-model="addNetworkForm.bridge" style="width: 100%">
            <el-option label="vmbr0" value="vmbr0" />
            <el-option label="vmbr1" value="vmbr1" />
            <el-option label="vmbr2" value="vmbr2" />
          </el-select>
        </el-form-item>
        <el-form-item label="MAC 地址">
          <el-input v-model="addNetworkForm.mac" placeholder="留空自动生成">
            <template #append>
              <el-button text @click="generateMac">生成</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="VLAN Tag">
          <el-input-number v-model="addNetworkForm.vlan" :min="0" :max="4094" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddNetworkDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddNetwork">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import { setQEMUConfig, resizeQEMUDisk } from '@/api/qemu'
import type { QEMUConfig } from '@/api/types'

interface Props {
  config: QEMUConfig
  node: string
  vmid: number
}

const props = defineProps<Props>()
const emit = defineEmits<{ refresh: [] }>()

/**
 * 判断是否有待生效配置
 */
const hasPendingConfig = ref(false)

/**
 * 解析硬件配置并展平为表格数据
 */
const hardwareItems = computed(() => {
  const items: Array<{
    type: string
    device: string
    value: string
    detail: string
    editable: boolean
    deletable: boolean
    rawKey?: string
  }> = []

  // BIOS / Boot Order
  items.push({
    type: 'BIOS/启动',
    device: 'boot',
    value: props.config.boot || 'order=cdn',
    detail: '启动顺序设置',
    editable: true,
    deletable: false,
  })

  // CPU
  const sockets = props.config.sockets || 1
  const cores = props.config.cores || 1
  const totalCores = sockets * cores
  items.push({
    type: 'CPU',
    device: `cpu${props.config.numac ? ' (NUMA)' : ''}`,
    value: `${totalCores} 核心 (${sockets}S × ${cores}C)`,
    detail: props.config.cpu ? `类型: ${props.config.cpu}` : '默认 (kvm64)',
    editable: true,
    deletable: false,
  })

  // Memory
  const memoryMB = props.config.memory || 0
  const balloon = props.config.ballooning || 0
  items.push({
    type: '内存',
    device: 'memory',
    value: `${(memoryMB / 1024).toFixed(1)} GB`,
    detail: balloon > 0 ? `Ballooning: ${(balloon / 1024).toFixed(1)} GB` : '无 Ballooning',
    editable: true,
    deletable: false,
  })

  // SCSI disks
  if (props.config.scsi) {
    props.config.scsi.forEach((disk, index) => {
      const parts = disk.split(',')
      const storage = parts[0]?.split(':')[0] || '-'
      const sizePart = parts.find(p => p.includes('size='))
      const size = sizePart ? sizePart.replace('size=', '') : '-'
      const formatPart = parts.find(p => p.includes('format='))
      const format = formatPart ? formatPart.replace('format=', '') : 'raw'
      items.push({
        type: '硬盘 (SCSI)',
        device: `scsi${index}`,
        value: `${size} (${storage})`,
        detail: `格式: ${format}`,
        editable: true,
        deletable: true,
      })
    })
  }

  // VirtIO disks
  if (props.config.virtio) {
    props.config.virtio.forEach((disk, index) => {
      const parts = disk.split(',')
      const storage = parts[0]?.split(':')[0] || '-'
      const sizePart = parts.find(p => p.includes('size='))
      const size = sizePart ? sizePart.replace('size=', '') : '-'
      items.push({
        type: '硬盘 (VirtIO)',
        device: `virtio${index}`,
        value: `${size} (${storage})`,
        detail: '',
        editable: true,
        deletable: true,
      })
    })
  }

  // Network adapters
  if (props.config.net) {
    props.config.net.forEach((net, index) => {
      const parts = net.split(',')
      const model = parts[0]?.split('=')[1] || parts[0] || '-'
      const bridge = parts.find(p => p.startsWith('bridge='))?.replace('bridge=', '') || '-'
      const mac = parts.find(p => p.includes('='))?.split('=')[1] || '-'
      items.push({
        type: '网卡',
        device: `net${index}`,
        value: `${model} @ ${bridge}`,
        detail: `MAC: ${mac}`,
        editable: true,
        deletable: true,
      })
    })
  }

  // CD/DVD
  if (props.config.ide) {
    props.config.ide.forEach((drive, index) => {
      if (drive.includes('media=cdrom') || drive.includes('.iso')) {
        items.push({
          type: '光驱',
          device: `ide${index}`,
          value: drive.includes('none') ? '未挂载' : drive,
          detail: 'ISO 镜像',
          editable: true,
          deletable: false,
        })
      }
    })
  }

  // USB
  if (props.config.usb && props.config.usb.length > 0) {
    props.config.usb.forEach((usb, index) => {
      items.push({
        type: 'USB 设备',
        device: `usb${index}`,
        value: usb,
        detail: '',
        editable: true,
        deletable: true,
      })
    })
  }

  return items
})

// 磁盘调整
const showResizeDialog = ref(false)
const resizeForm = ref({ disk: '', currentSize: '', expandSize: '10' })

function handleEdit(item: (typeof hardwareItems.value)[0]) {
  if (item.type.startsWith('硬盘')) {
    resizeForm.value = {
      disk: item.device,
      currentSize: item.value,
      expandSize: '10',
    }
    showResizeDialog.value = true
  } else {
    ElMessage.info(`${item.device} 编辑功能开发中`)
  }
}

async function confirmResize() {
  if (!resizeForm.value.expandSize) {
    ElMessage.warning('请输入扩容大小')
    return
  }
  try {
    const size = resizeForm.value.expandSize.startsWith('+')
      ? resizeForm.value.expandSize
      : `+${resizeForm.value.expandSize}G`
    await resizeQEMUDisk(props.node, props.vmid, resizeForm.value.disk, size)
    ElMessage.success('磁盘扩容命令已发送')
    showResizeDialog.value = false
    emit('refresh')
  } catch (error) {
    console.error('磁盘扩容失败:', error)
  }
}

function handleDelete(item: (typeof hardwareItems.value)[0]) {
  ElMessageBox.confirm(`确认删除 ${item.device}？`, '确认', {
    type: 'warning',
  }).then(async () => {
    try {
      await setQEMUConfig(props.node, props.vmid, { [item.device]: 'none' })
      ElMessage.success(`${item.device} 已标记为删除`)
      emit('refresh')
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// 添加硬件
const showAddDiskDialog = ref(false)
const showAddNetworkDialog = ref(false)
const addDiskForm = ref({
  storage: 'local',
  size: '50',
  bus: 'scsi',
  format: 'raw',
  ssd: false,
  discard: false,
})
const addNetworkForm = ref({
  model: 'virtio',
  bridge: 'vmbr0',
  mac: '',
  vlan: 0,
})

function handleAddCommand(command: string) {
  switch (command) {
    case 'disk':
      showAddDiskDialog.value = true
      break
    case 'network':
      showAddNetworkDialog.value = true
      break
    default:
      ElMessage.info(`添加${command}功能开发中`)
  }
}

function generateMac() {
  const parts = ['AA', 'AA', 'AA', 'AA', 'AA', 'AA']
  for (let i = 0; i < 6; i++) {
    parts[i] = Math.floor(Math.random() * 256).toString(16).padStart(2, '0').toUpperCase()
  }
  addNetworkForm.value.mac = parts.join(':')
}

async function confirmAddDisk() {
  const { storage, size, bus, format, ssd, discard } = addDiskForm.value
  const config: Record<string, unknown> = {}

  // 找到下一个可用的 disk slot
  const existingDisks = hardwareItems.value.filter(item => item.type.startsWith('硬盘'))
  const nextIndex = existingDisks.length

  const diskValue = `${storage}:${size},format=${format}${ssd ? ',ssd=1' : ''}${discard ? ',discard=on' : ''}`
  config[`${bus}${nextIndex}`] = diskValue

  try {
    await setQEMUConfig(props.node, props.vmid, config)
    ElMessage.success('硬盘添加成功')
    showAddDiskDialog.value = false
    emit('refresh')
  } catch (error) {
    console.error('添加硬盘失败:', error)
  }
}

async function confirmAddNetwork() {
  const { model, bridge, mac, vlan } = addNetworkForm.value
  const config: Record<string, unknown> = {}

  const existingNets = hardwareItems.value.filter(item => item.type === '网卡')
  const nextIndex = existingNets.length

  let netValue = `${model}=bridge=${bridge}`
  if (mac) netValue += `,mac=${mac}`
  if (vlan > 0) netValue += `,tag=${vlan}`

  config[`net${nextIndex}`] = netValue

  try {
    await setQEMUConfig(props.node, props.vmid, config)
    ElMessage.success('网卡添加成功')
    showAddNetworkDialog.value = false
    emit('refresh')
  } catch (error) {
    console.error('添加网卡失败:', error)
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.hardware-tab {
  display: flex;
  flex-direction: column;
  gap: $spacing-6;
}

.pending-alert {
  border-radius: $radius-base;
}

.action-buttons {
  display: flex;
  gap: $spacing-2;
}

.add-hardware-bar {
  display: flex;
  justify-content: flex-start;
}
</style>
