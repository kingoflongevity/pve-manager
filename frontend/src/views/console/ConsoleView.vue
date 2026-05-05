<template>
  <div class="console-view">
    <!-- 加载状态 -->
    <div v-if="isLoading" class="loading-overlay">
      <el-icon class="loading-icon is-loading"><Loading /></el-icon>
      <p>正在加载控制台...</p>
    </div>

    <!-- QEMU 虚拟机使用 noVNC 远程桌面 -->
    <div v-if="!isLoading && vmType === 'qemu'" class="console-body">
      <NoVNCConsole
        ref="novncRef"
        :node="node"
        :vmid="vmid"
        :vm-type="vmType"
      />
    </div>

    <!-- LXC 容器使用 xterm.js 终端 -->
    <div v-if="!isLoading && vmType === 'lxc'" class="console-body">
      <XTermConsole
        ref="xtermRef"
        :node="node"
        :vmid="vmid"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'
import NoVNCConsole from '@/components/console/NoVNCConsole.vue'
import XTermConsole from '@/components/console/XTermConsole.vue'
import { getQEMUConfig } from '@/api/qemu'
import { getLXCConfig } from '@/api/lxc'

const route = useRoute()

const node = computed(() => route.params.node as string)
const vmid = computed(() => Number(route.params.vmid))
const vmType = computed(() => (route.params.vmType as 'qemu' | 'lxc') || 'qemu')

const novncRef = ref<InstanceType<typeof NoVNCConsole> | null>(null)
const xtermRef = ref<InstanceType<typeof XTermConsole> | null>(null)
const isLoading = ref(true)

async function loadVMInfo() {
  try {
    if (vmType.value === 'qemu') {
      await getQEMUConfig(node.value, vmid.value)
    } else {
      await getLXCConfig(node.value, vmid.value)
    }
  } catch {
    // ignore
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  await loadVMInfo()
})
</script>

<style lang="scss" scoped>
.console-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1e1e1e;
  overflow: hidden;
}

.console-body {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: #000;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.85);
  color: #fff;
  z-index: 10;

  .loading-icon {
    font-size: 48px;
    margin-bottom: 16px;
    color: #409eff;
  }

  p {
    font-size: 14px;
    margin: 0;
  }
}
</style>
