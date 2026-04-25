<template>
  <el-dialog
    v-model="visible"
    title="创建容器"
    width="720px"
    :close-on-click-modal="false"
    destroy-on-close
    @close="handleClose"
  >
    <!-- 步骤条 -->
    <el-steps :active="currentStep" finish-status="success" class="wizard-steps">
      <el-step title="基本信息" />
      <el-step title="模板" />
      <el-step title="根磁盘" />
      <el-step title="CPU" />
      <el-step title="内存" />
      <el-step title="网络" />
      <el-step title="DNS" />
      <el-step title="功能" />
      <el-step title="确认" />
    </el-steps>

    <!-- Step 1: 基本信息 -->
    <div v-show="currentStep === 0" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="节点" required>
          <el-select v-model="form.node" placeholder="选择节点" style="width: 100%">
            <el-option label="pve-node-01" value="pve-node-01" />
            <el-option label="pve-node-02" value="pve-node-02" />
          </el-select>
        </el-form-item>
        <el-form-item label="CT ID" required>
          <el-input-number v-model="form.vmid" :min="100" :max="999999999" style="width: 100%">
            <template #append><el-button text @click="suggestVmid">自动</el-button></template>
          </el-input-number>
        </el-form-item>
        <el-form-item label="主机名">
          <el-input v-model="form.hostname" placeholder="容器主机名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="root 用户密码" show-password />
        </el-form-item>
        <el-form-item label="SSH 公钥">
          <el-input v-model="form.pubkey" type="textarea" :rows="2" placeholder="可选 SSH 公钥" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 2: 模板 -->
    <div v-show="currentStep === 1" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="模板">
          <el-select v-model="form.template" placeholder="选择容器模板" style="width: 100%" filterable>
            <el-option label="ubuntu-24.04-standard_24.04-1_amd64.tar.zst" value="local:vztmpl/ubuntu-24.04-standard_24.04-1_amd64.tar.zst" />
            <el-option label="debian-12-standard_12.6-1_amd64.tar.zst" value="local:vztmpl/debian-12-standard_12.6-1_amd64.tar.zst" />
            <el-option label="alpine-3.19-default_3.19.1-1_amd64.tar.zst" value="local:vztmpl/alpine-3.19-default_3.19.1-1_amd64.tar.zst" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板存储">
          <el-select v-model="form.templateStorage" style="width: 100%">
            <el-option label="local" value="local" />
          </el-select>
        </el-form-item>
        <el-form-item label="非特权容器">
          <el-switch v-model="form.unprivileged" />
          <span class="form-hint">推荐启用，提高安全性</span>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 3: Root 磁盘 -->
    <div v-show="currentStep === 2" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="存储" required>
          <el-select v-model="form.rootfsStorage" style="width: 100%">
            <el-option label="local" value="local" />
            <el-option label="local-lvm" value="local-lvm" />
          </el-select>
        </el-form-item>
        <el-form-item label="磁盘大小" required>
          <el-input v-model="form.rootfsSize" style="width: 100%">
            <template #append>GB</template>
          </el-input>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 4: CPU -->
    <div v-show="currentStep === 3" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="核心数">
          <el-input-number v-model="form.cores" :min="1" :max="64" style="width: 100%" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 5: 内存和 Swap -->
    <div v-show="currentStep === 4" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="内存" required>
          <el-input v-model="form.memory" style="width: 100%">
            <template #append>MB</template>
          </el-input>
        </el-form-item>
        <el-form-item label="Swap">
          <el-input v-model="form.swap" style="width: 100%">
            <template #append>MB</template>
          </el-input>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 6: 网络 -->
    <div v-show="currentStep === 5" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="桥接">
          <el-select v-model="form.netBridge" style="width: 100%">
            <el-option label="vmbr0" value="vmbr0" />
            <el-option label="vmbr1" value="vmbr1" />
            <el-option label="vmbr2" value="vmbr2" />
          </el-select>
        </el-form-item>
        <el-form-item label="IPv4">
          <el-radio-group v-model="form.ipv4Mode">
            <el-radio value="dhcp">DHCP</el-radio>
            <el-radio value="static">静态</el-radio>
            <el-radio value="none">无</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.ipv4Mode === 'static'" label="IPv4 地址">
          <el-input v-model="form.ipv4Address" placeholder="例如: 192.168.1.100/24" />
        </el-form-item>
        <el-form-item v-if="form.ipv4Mode === 'static'" label="网关">
          <el-input v-model="form.ipv4Gateway" placeholder="例如: 192.168.1.1" />
        </el-form-item>
        <el-form-item label="IPv6">
          <el-radio-group v-model="form.ipv6Mode">
            <el-radio value="dhcp">DHCP</el-radio>
            <el-radio value="static">静态</el-radio>
            <el-radio value="none">无</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 7: DNS -->
    <div v-show="currentStep === 6" class="wizard-step-content">
      <el-form :model="form" label-width="100px">
        <el-form-item label="DNS 域名">
          <el-input v-model="form.dnsDomain" placeholder="例如: example.com" />
        </el-form-item>
        <el-form-item label="DNS 服务器">
          <el-input v-model="form.dnsServer" placeholder="例如: 8.8.8.8" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 8: 功能 -->
    <div v-show="currentStep === 7" class="wizard-step-content">
      <el-form label-width="160px">
        <el-form-item label="Nesting">
          <el-switch v-model="form.featureNesting" />
          <span class="form-hint">允许嵌套虚拟化</span>
        </el-form-item>
        <el-form-item label="Keyctl">
          <el-switch v-model="form.featureKeyctl" />
          <span class="form-hint">允许 keyctl 系统调用</span>
        </el-form-item>
      </el-form>
    </div>

    <!-- Step 9: 确认 -->
    <div v-show="currentStep === 8" class="wizard-step-content">
      <el-card>
        <template #header><span class="card-title">配置摘要</span></template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="CT ID">{{ form.vmid }}</el-descriptions-item>
          <el-descriptions-item label="主机名">{{ form.hostname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ form.node }}</el-descriptions-item>
          <el-descriptions-item label="非特权">{{ form.unprivileged ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="根磁盘">{{ form.rootfsSize }} GB ({{ form.rootfsStorage }})</el-descriptions-item>
          <el-descriptions-item label="CPU">{{ form.cores }} 核心</el-descriptions-item>
          <el-descriptions-item label="内存">{{ form.memory }} MB</el-descriptions-item>
          <el-descriptions-item label="Swap">{{ form.swap }} MB</el-descriptions-item>
          <el-descriptions-item label="网络">{{ form.netBridge }}</el-descriptions-item>
          <el-descriptions-item label="IPv4">{{ form.ipv4Mode }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-form-item style="margin-top: 24px">
        <el-checkbox v-model="form.startAfterCreate">创建后启动</el-checkbox>
      </el-form-item>
    </div>

    <!-- 底部操作 -->
    <template #footer>
      <div class="wizard-footer">
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <div style="flex: 1" />
        <el-button @click="handleClose">取消</el-button>
        <el-button v-if="currentStep < 8" type="primary" @click="currentStep++">下一步</el-button>
        <el-button v-else type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { createLXC } from '@/api/lxc'

interface LXCCreateForm {
  node: string
  vmid: number
  hostname: string
  password: string
  pubkey: string
  template: string
  templateStorage: string
  unprivileged: boolean
  rootfsStorage: string
  rootfsSize: string
  cores: number
  memory: string
  swap: string
  netBridge: string
  ipv4Mode: string
  ipv4Address: string
  ipv4Gateway: string
  ipv6Mode: string
  dnsDomain: string
  dnsServer: string
  featureNesting: boolean
  featureKeyctl: boolean
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

const form = ref<LXCCreateForm>({
  node: props.defaultNode || 'pve-node-01',
  vmid: 100,
  hostname: '',
  password: '',
  pubkey: '',
  template: '',
  templateStorage: 'local',
  unprivileged: true,
  rootfsStorage: 'local-lvm',
  rootfsSize: '8',
  cores: 1,
  memory: '512',
  swap: '0',
  netBridge: 'vmbr0',
  ipv4Mode: 'dhcp',
  ipv4Address: '',
  ipv4Gateway: '',
  ipv6Mode: 'none',
  dnsDomain: '',
  dnsServer: '',
  featureNesting: false,
  featureKeyctl: false,
  startAfterCreate: false,
})

function suggestVmid() {
  form.value.vmid = 100 + Math.floor(Math.random() * 900)
}

function handleClose() {
  visible.value = false
  currentStep.value = 0
}

/**
 * 创建 LXC 容器
 */
async function handleCreate() {
  if (!form.value.node || !form.value.vmid || !form.value.template) {
    ElMessage.warning('请填写必填字段')
    return
  }

  creating.value = true
  try {
    const params: Record<string, unknown> = {
      vmid: form.value.vmid,
      hostname: form.value.hostname,
      ostemplate: form.value.template,
      memory: Number(form.value.memory),
      swap: Number(form.value.swap),
      cores: form.value.cores,
      storage: form.value.rootfsStorage,
      rootfs: `size=${form.value.rootfsSize}G`,
      net0: `${form.value.netBridge},bridge=${form.value.netBridge}`,
      unprivileged: form.value.unprivileged ? 1 : 0,
    }

    if (form.value.password) params.password = form.value.password
    if (form.value.pubkey) params.pubkey = form.value.pubkey
    if (form.value.ipv4Mode === 'static') {
      params.net0 = `${form.value.netBridge},ip=${form.value.ipv4Address},gw=${form.value.ipv4Gateway}`
    }
    if (form.value.dnsServer) params.nameserver = form.value.dnsServer
    if (form.value.dnsDomain) params.searchdomain = form.value.dnsDomain

    const features = []
    if (form.value.featureNesting) features.push('nesting=1')
    if (form.value.featureKeyctl) features.push('keyctl=1')
    if (features.length > 0) params.features = features.join(',')

    await createLXC(form.value.node, params as import('@/api/types').LXCCreateParams)
    ElMessage.success('容器创建命令已发送')
    emit('created', form.value.vmid)
    handleClose()
  } catch (error) {
    console.error('创建容器失败:', error)
  } finally {
    creating.value = false
  }
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.wizard-steps { margin-bottom: $spacing-8; padding: 0 $spacing-4; }
.wizard-step-content { padding: $spacing-4; min-height: 280px; }
.form-hint { font-size: $font-size-xs; color: $color-text-secondary; margin-left: $spacing-3; }
.card-title { font-weight: $font-weight-semibold; font-size: $font-size-md; }
.wizard-footer { display: flex; align-items: center; gap: $spacing-3; }
</style>
