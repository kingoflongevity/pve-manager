<template>
  <el-dialog
    v-model="visible"
    title="创建虚拟机"
    width="720px"
    :close-on-click-modal="false"
    destroy-on-close
    @close="handleClose"
  >
    <!-- 步骤条 -->
    <el-steps :active="currentStep" finish-status="success" class="wizard-steps">
      <el-step title="基本信息" />
      <el-step title="操作系统" />
      <el-step title="系统" />
      <el-step title="磁盘" />
      <el-step title="CPU" />
      <el-step title="内存" />
      <el-step title="网络" />
      <el-step title="确认" />
    </el-steps>

    <!-- Step 1: 基本信息 -->
    <div v-show="currentStep === 0" class="wizard-step-content">
      <el-form :model="form" label-width="100px" ref="formRef">
        <el-form-item label="节点" required>
          <el-select v-model="form.node" placeholder="选择节点" style="width: 100%">
            <el-option label="pve-node-01" value="pve-node-01" />
            <el-option label="pve-node-02" value="pve-node-02" />
          </el-select>
        </el-form-item>
        <el-form-item label="VM ID" required>
          <el-input-number v-model="form.vmid" :min="100" :max="999999999" style="width: 100%">
            <template #append>
              <el-button text @click="suggestVmid">自动</el-button>
            </template>
          </el-input-number>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="虚拟机名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选描述" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 2: 操作系统 -->
    <div v-show="currentStep === 1" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="安装方式">
          <el-radio-group v-model="form.osMode">
            <el-radio value="iso">使用 ISO 镜像</el-radio>
            <el-radio value="template">使用模板</el-radio>
            <el-radio value="none">不使用介质</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.osMode === 'iso'">
          <el-form-item label="存储">
            <el-select v-model="form.isoStorage" placeholder="选择 ISO 存储" style="width: 100%">
              <el-option label="local" value="local" />
              <el-option label="nfs-storage" value="nfs-storage" />
            </el-select>
          </el-form-item>
          <el-form-item label="ISO 文件">
            <el-select v-model="form.isoFile" placeholder="选择 ISO 文件" style="width: 100%">
              <el-option label="ubuntu-24.04-live-server-amd64.iso" value="ubuntu-24.04-live-server-amd64.iso" />
              <el-option label="debian-12.5.0-amd64-netinst.iso" value="debian-12.5.0-amd64-netinst.iso" />
              <el-option label="CentOS-Stream-9-x86_64-boot.iso" value="CentOS-Stream-9-x86_64-boot.iso" />
            </el-select>
          </el-form-item>
        </template>
      </el-form>
    </div>

    <!-- Step 3: 系统 -->
    <div v-show="currentStep === 2" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="BIOS">
          <el-select v-model="form.bios" style="width: 100%">
            <el-option label="SeaBIOS (传统)" value="seabios" />
            <el-option label="OVMF (UEFI)" value="ovmf" />
          </el-select>
        </el-form-item>
        <el-form-item label="机器类型">
          <el-select v-model="form.machine" style="width: 100%">
            <el-option label="i440fx (默认)" value="pc" />
            <el-option label="q35" value="q35" />
          </el-select>
        </el-form-item>
        <el-form-item label="SCSI 控制器">
          <el-select v-model="form.scsiController" style="width: 100%">
            <el-option label="LSI 53C895A (默认)" value="lsi" />
            <el-option label="VirtIO SCSI" value="virtio-scsi-pci" />
            <el-option label="VMware PVSCSI" value="vmw_pvscsi" />
          </el-select>
        </el-form-item>
        <el-form-item label="QEMU Agent">
          <el-switch v-model="form.agent" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 4: 磁盘 -->
    <div v-show="currentStep === 3" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="总线类型">
          <el-select v-model="form.diskBus" style="width: 100%">
            <el-option label="SCSI" value="scsi" />
            <el-option label="VirtIO Block" value="virtio" />
            <el-option label="SATA" value="sata" />
            <el-option label="IDE" value="ide" />
          </el-select>
        </el-form-item>
        <el-form-item label="存储">
          <el-select v-model="form.diskStorage" style="width: 100%">
            <el-option label="local" value="local" />
            <el-option label="local-lvm" value="local-lvm" />
          </el-select>
        </el-form-item>
        <el-form-item label="磁盘大小" required>
          <el-input v-model="form.diskSize" style="width: 100%">
            <template #append>GB</template>
          </el-input>
        </el-form-item>
        <el-form-item label="格式">
          <el-select v-model="form.diskFormat" style="width: 100%">
            <el-option label="Raw" value="raw" />
            <el-option label="Qcow2" value="qcow2" />
          </el-select>
        </el-form-item>
        <el-form-item label="SSD 模拟">
          <el-switch v-model="form.diskSsd" />
        </el-form-item>
        <el-form-item label="Discard">
          <el-switch v-model="form.diskDiscard" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 5: CPU -->
    <div v-show="currentStep === 4" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="Sockets">
          <el-input-number v-model="form.sockets" :min="1" :max="16" style="width: 100%" />
        </el-form-item>
        <el-form-item label="每 Socket 核心数">
          <el-input-number v-model="form.cores" :min="1" :max="64" style="width: 100%" />
        </el-form-item>
        <el-form-item label="CPU 类型">
          <el-select v-model="form.cpuType" style="width: 100%">
            <el-option label="host (推荐)" value="host" />
            <el-option label="kvm64" value="kvm64" />
            <el-option label="x86-64-v2-AES" value="x86-64-v2-AES" />
            <el-option label="x86-64-v3" value="x86-64-v3" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用 NUMA">
          <el-switch v-model="form.numa" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 6: 内存 -->
    <div v-show="currentStep === 5" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="内存" required>
          <el-input v-model="form.memory" style="width: 100%">
            <template #append>MB</template>
          </el-input>
        </el-form-item>
        <el-form-item label="最小内存 (Ballooning)">
          <el-input v-model="form.balloon" style="width: 100%">
            <template #append>MB</template>
          </el-input>
          <span class="form-hint">设为 0 则不启用 Ballooning</span>
        </el-form-item>
        <el-form-item label="Swap">
          <el-input v-model="form.swap" style="width: 100%">
            <template #append>MB</template>
          </el-input>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 7: 网络 -->
    <div v-show="currentStep === 6" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="桥接">
          <el-select v-model="form.netBridge" style="width: 100%">
            <el-option label="vmbr0" value="vmbr0" />
            <el-option label="vmbr1" value="vmbr1" />
            <el-option label="vmbr2" value="vmbr2" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型">
          <el-select v-model="form.netModel" style="width: 100%">
            <el-option label="VirtIO (推荐)" value="virtio" />
            <el-option label="Intel e1000" value="e1000" />
            <el-option label="Realtek RTL8139" value="rtl8139" />
          </el-select>
        </el-form-item>
        <el-form-item label="MAC 地址">
          <el-input v-model="form.netMac" placeholder="留空自动生成">
            <template #append>
              <el-button text @click="generateMac">生成</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="VLAN Tag">
          <el-input-number v-model="form.netVlan" :min="0" :max="4094" placeholder="可选" style="width: 100%" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 8: 确认 -->
    <div v-show="currentStep === 7" class="wizard-step-content">
      <el-card>
        <template #header><span class="card-title">配置摘要</span></template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="VM ID">{{ form.vmid }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ form.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ form.node }}</el-descriptions-item>
          <el-descriptions-item label="BIOS">{{ form.bios === 'ovmf' ? 'OVMF (UEFI)' : 'SeaBIOS' }}</el-descriptions-item>
          <el-descriptions-item label="CPU">
            {{ form.sockets }}S × {{ form.cores }}C ({{ form.cpuType }})
          </el-descriptions-item>
          <el-descriptions-item label="内存">{{ form.memory }} MB</el-descriptions-item>
          <el-descriptions-item label="磁盘">
            {{ form.diskSize }} GB ({{ form.diskBus }}, {{ form.diskFormat }})
          </el-descriptions-item>
          <el-descriptions-item label="网络">
            {{ form.netModel }} @ {{ form.netBridge }}
          </el-descriptions-item>
          <el-descriptions-item label="QEMU Agent">{{ form.agent ? '启用' : '禁用' }}</el-descriptions-item>
          <el-descriptions-item label="安装方式">{{ osModeLabel }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-form-item style="margin-top: 24px">
        <el-checkbox v-model="form.startAfterCreate">创建后启动</el-checkbox>
      </el-form-item>
    </div>

    <!-- 底部操作栏 -->
    <template #footer>
      <div class="wizard-footer">
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <div style="flex: 1" />
        <el-button @click="handleClose">取消</el-button>
        <el-button v-if="currentStep < 7" type="primary" @click="currentStep++">下一步</el-button>
        <el-button v-else type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { createQEMU } from '@/api/qemu'

interface QEMUCreateForm {
  node: string
  vmid: number
  name: string
  description: string
  osMode: 'iso' | 'template' | 'none'
  isoStorage: string
  isoFile: string
  bios: string
  machine: string
  scsiController: string
  agent: boolean
  diskBus: string
  diskStorage: string
  diskSize: string
  diskFormat: string
  diskSsd: boolean
  diskDiscard: boolean
  sockets: number
  cores: number
  cpuType: string
  numa: boolean
  memory: string
  balloon: string
  swap: string
  netBridge: string
  netModel: string
  netMac: string
  netVlan: number
  startAfterCreate: boolean
}

const props = defineProps<{
  modelValue: boolean
  defaultNode?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  created: [vmid: number]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v: boolean) => emit('update:modelValue', v),
})

const currentStep = ref(0)
const creating = ref(false)
const formRef = ref()

const form = ref<QEMUCreateForm>({
  node: props.defaultNode || 'pve-node-01',
  vmid: 100,
  name: '',
  description: '',
  osMode: 'iso',
  isoStorage: 'local',
  isoFile: '',
  bios: 'seabios',
  machine: 'pc',
  scsiController: 'lsi',
  agent: true,
  diskBus: 'scsi',
  diskStorage: 'local-lvm',
  diskSize: '32',
  diskFormat: 'raw',
  diskSsd: false,
  diskDiscard: false,
  sockets: 1,
  cores: 2,
  cpuType: 'host',
  numa: false,
  memory: '2048',
  balloon: '',
  swap: '',
  netBridge: 'vmbr0',
  netModel: 'virtio',
  netMac: '',
  netVlan: 0,
  startAfterCreate: false,
})

/**
 * OS 安装方式标签
 */
const osModeLabel = computed(() => {
  const map = { iso: 'ISO 镜像', template: '模板', none: '不使用介质' }
  return map[form.value.osMode] || '-'
})

/**
 * 自动建议 VM ID
 */
function suggestVmid() {
  form.value.vmid = 100 + Math.floor(Math.random() * 900)
}

/**
 * 生成随机 MAC 地址
 */
function generateMac() {
  const parts = Array.from({ length: 6 }, () =>
    Math.floor(Math.random() * 256).toString(16).padStart(2, '0').toUpperCase()
  )
  form.value.netMac = parts.join(':')
}

/**
 * 关闭向导
 */
function handleClose() {
  visible.value = false
  currentStep.value = 0
}

/**
 * 创建虚拟机
 */
async function handleCreate() {
  if (!form.value.node || !form.value.vmid) {
    ElMessage.warning('请填写必填字段')
    return
  }

  creating.value = true
  try {
    // 构建创建参数
    const params: Record<string, unknown> = {
      vmid: form.value.vmid,
      name: form.value.name,
      description: form.value.description,
      memory: Number(form.value.memory),
      cores: form.value.cores,
      sockets: form.value.sockets,
      bios: form.value.bios,
      agent: form.value.agent ? 1 : 0,
      net0: `${form.value.netModel}=bridge=${form.value.netBridge}${form.value.netMac ? `,mac=${form.value.netMac}` : ''}${form.value.netVlan > 0 ? `,tag=${form.value.netVlan}` : ''}`,
      scsi0: `${form.value.diskStorage}:${form.value.diskSize},format=${form.value.diskFormat}${form.value.diskSsd ? ',ssd=1' : ''}${form.value.diskDiscard ? ',discard=on' : ''}`,
      cpu: form.value.cpuType,
    }

    if (form.value.balloon) params.balloon = Number(form.value.balloon)
    if (form.value.numa) params.numac = 1
    if (form.value.osMode === 'iso' && form.value.isoFile) {
      params.ide2 = `${form.value.isoStorage}:iso/${form.value.isoFile},media=cdrom`
    }

    await createQEMU(form.value.node, params as unknown as import('@/api/types').QEMUCreateParams)
    ElMessage.success('虚拟机创建命令已发送')
    emit('created', form.value.vmid)
    handleClose()
  } catch (error) {
    console.error('创建虚拟机失败:', error)
  } finally {
    creating.value = false
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.wizard-steps {
  margin-bottom: $spacing-8;
  padding: 0 $spacing-4;
}

.wizard-step-content {
  padding: $spacing-4;
  min-height: 280px;
}

.form-hint {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  margin-left: $spacing-3;
}

.card-title {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
}

.wizard-footer {
  display: flex;
  align-items: center;
  gap: $spacing-3;
}
</style>
