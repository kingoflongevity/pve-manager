<template>
  <div class="app-store">
    <div class="page-header">
      <h2>应用商店</h2>
      <div class="header-actions">
        <el-input v-model="searchKey" prefix-icon="Search" placeholder="搜索应用..." size="small" style="width: 240px; margin-right: 8px" clearable />
        <el-button size="small" @click="showImportDialog">导入模板</el-button>
        <el-button size="small" @click="syncTemplates">同步</el-button>
      </div>
    </div>

    <div class="category-bar">
      <el-radio-group v-model="currentCategory" @change="fetchTemplates">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button v-for="cat in categories" :key="cat" :label="cat" />
      </el-radio-group>
    </div>

    <div class="app-grid">
      <el-card v-for="app in filteredApps" :key="app.id" class="app-card" shadow="hover" @click="viewDetail(app)">
        <div class="app-icon">
          <el-icon :size="32">
              <component :is="getCategoryIcon(app.category)" />
            </el-icon>
        </div>
        <div class="app-info">
          <h3>{{ app.name }}</h3>
          <p>{{ app.description }}</p>
          <div class="app-meta">
            <span>v{{ app.version }}</span>
            <div class="app-tags">
              <el-tag size="small" type="info">{{ categoryLabel(app.category) }}</el-tag>
              <el-tag size="small" :type="app.type === 'lxc' ? 'success' : 'warning'">{{ app.type === 'lxc' ? 'LXC' : 'QEMU' }}</el-tag>
            </div>
          </div>
        </div>
      </el-card>
      <el-empty v-if="!filteredApps.length" description="暂无应用" />
    </div>

    <el-dialog v-model="detailDialog" :title="detailApp?.name || ''" width="640px">
      <div v-if="detailApp">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="版本">{{ detailApp.version }}</el-descriptions-item>
          <el-descriptions-item label="作者">{{ detailApp.author || '-' }}</el-descriptions-item>
          <el-descriptions-item label="部署类型">
            <el-tag :type="detailApp.type === 'lxc' ? 'success' : 'warning'" size="small">{{ detailApp.type === 'lxc' ? 'LXC 容器' : 'QEMU 虚拟机' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="OS 模板">{{ detailApp.os_template || '-' }}</el-descriptions-item>
          <el-descriptions-item label="最低 CPU">{{ detailApp.min_cpu }} 核</el-descriptions-item>
          <el-descriptions-item label="最低内存">{{ detailApp.min_memory_mb }} MB</el-descriptions-item>
          <el-descriptions-item label="最低磁盘">{{ detailApp.min_disk_gb }} GB</el-descriptions-item>
          <el-descriptions-item label="内置">
            <el-tag :type="detailApp.is_built_in ? 'success' : 'info'" size="small">{{ detailApp.is_built_in ? '是' : '否' }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <p class="app-desc">{{ detailApp.description }}</p>

        <div v-if="setupStepList.length" class="setup-steps">
          <h4>部署步骤</h4>
          <el-steps direction="vertical" :active="0">
            <el-step v-for="(step, idx) in setupStepList" :key="idx" :title="`步骤 ${idx + 1}`" :description="step" />
          </el-steps>
        </div>

        <el-form :model="deployForm" label-width="100px" style="margin-top: 20px">
          <el-form-item label="实例名称" required>
            <el-input v-model="deployForm.name" placeholder="例如: my-nginx" />
          </el-form-item>
          <el-form-item label="目标节点" required>
            <el-select v-model="deployForm.node" style="width: 100%" placeholder="选择 PVE 节点">
              <el-option v-for="n in nodes" :key="n" :label="n" :value="n" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="detailDialog = false">取消</el-button>
        <el-button type="primary" :loading="deploying" @click="doDeploy">部署应用</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importDialog" title="导入模板" width="500px">
      <el-input v-model="importYAML" type="textarea" :rows="12" placeholder="粘贴 YAML 模板内容..." />
      <template #footer>
        <el-button @click="importDialog = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="doImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Monitor, Coin, Connection, Tools, VideoCamera, Files, Lock, Menu } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getAppTemplates, getAppCategories, deployApp, importAppTemplate, syncAppTemplates } from '@/api/appStore'
import { getClusterResources } from '@/api/cluster'
import type { AppTemplate } from '@/api/appStore'

const router = useRouter()
const searchKey = ref('')
const currentCategory = ref('')
const categories = ref<string[]>([])
const templates = ref<AppTemplate[]>([])
const detailDialog = ref(false)
const detailApp = ref<AppTemplate | null>(null)
const importDialog = ref(false)
const importYAML = ref('')
const deploying = ref(false)
const importing = ref(false)
const nodes = ref<string[]>([])
const deployForm = ref({ name: '', node: '' })


const setupStepList = computed(() => {
  if (!detailApp.value?.setup_steps) return []
  try {
    const raw = JSON.parse(detailApp.value.setup_steps)
    return raw.map((s: { step: string; desc: string }) => s.desc)
  } catch { return [] }
})

const categoryIcons: Record<string, any> = {
  infrastructure: Monitor,
  database: Coin,
  networking: Connection,
  monitoring: Monitor,
  devops: Tools,
  media: VideoCamera,
  storage: Files,
  security: Lock,
}

const categoryLabels: Record<string, string> = {
  'Web Server': 'Web 服务', 'Database': '数据库', 'CMS': '内容管理',
  'DevOps': '开发运维', 'Monitor': '监控运维', 'Storage': '存储服务',
  'Network': '网络服务', 'Messaging': '消息队列', 'Search': '搜索服务',
}

function getCategoryIcon(cat: string) { return categoryIcons[cat] || 'Menu' }
function categoryLabel(cat: string) { return categoryLabels[cat] || cat }

const filteredApps = computed(() => {
  let list = templates.value
  if (currentCategory.value) list = list.filter(a => a.category === currentCategory.value)
  if (searchKey.value) {
    const key = searchKey.value.toLowerCase()
    list = list.filter(a => a.name.toLowerCase().includes(key) || a.description?.toLowerCase().includes(key))
  }
  return list
})

onMounted(() => {
  fetchTemplates()
  fetchCategories()
  fetchNodes()
})

async function fetchTemplates() {
  try {
    templates.value = await getAppTemplates(currentCategory.value || undefined)
  } catch { ElMessage.error('获取应用列表失败') }
}

async function fetchCategories() {
  try { categories.value = await getAppCategories() } catch {}
}

async function fetchNodes() {
  try {
    const resources = await getClusterResources()
    const nodeSet = new Set<string>()
    for (const r of resources) {
      if (r.node) nodeSet.add(r.node)
    }
    nodes.value = nodeSet.size > 0 ? [...nodeSet].sort() : ['pve']
  } catch {
    nodes.value = ['pve']
  }
}

async function viewDetail(app: AppTemplate) {
  detailApp.value = app
  deployForm.value = { name: app.name.toLowerCase().replace(/\s+/g, '-'), node: nodes.value[0] || '' }
  detailDialog.value = true
}

async function doDeploy() {
  if (!detailApp.value) return
  if (!deployForm.value.name.trim()) { ElMessage.warning('请输入实例名称'); return }
  if (!deployForm.value.node) { ElMessage.warning('请选择目标节点'); return }
  deploying.value = true
  try {
    const resp = await deployApp({
      template_id: detailApp.value.id,
      name: deployForm.value.name,
      node: deployForm.value.node,
    })
    ElMessage.success(resp.message || '部署任务已提交')
    detailDialog.value = false
    router.push('/apps/deployments')
  } catch { ElMessage.error('部署失败') }
  finally { deploying.value = false }
}

function showImportDialog() {
  importYAML.value = ''
  importDialog.value = true
}

async function doImport() {
  if (!importYAML.value.trim()) { ElMessage.warning('请输入模板内容'); return }
  importing.value = true
  try {
    await importAppTemplate(importYAML.value)
    ElMessage.success('导入成功')
    importDialog.value = false
    fetchTemplates()
  } catch { ElMessage.error('导入失败') }
  finally { importing.value = false }
}

async function syncTemplates() {
  try {
    const resp = await syncAppTemplates('https://example.com/templates')
    ElMessage.success(`同步完成，新增 ${resp.synced_count} 个模板`)
    fetchTemplates()
  } catch { ElMessage.error('同步失败') }
}
</script>

<style scoped>
.app-store { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; font-size: 20px; }
.header-actions { display: flex; align-items: center; }
.category-bar { margin-bottom: 20px; }
.app-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }
.app-card { cursor: pointer; transition: transform 0.2s; }
.app-card:hover { transform: translateY(-2px); }
.app-icon { width: 48px; height: 48px; display: flex; align-items: center; justify-content: center; background: #f0f5ff; border-radius: 12px; color: #409eff; margin-bottom: 12px; }
.app-info h3 { margin: 0 0 8px; font-size: 16px; }
.app-info p { font-size: 13px; color: #909399; margin: 0 0 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.app-meta { display: flex; justify-content: space-between; align-items: center; font-size: 12px; color: #c0c4cc; }
.app-tags { display: flex; gap: 4px; }
.app-desc { margin-top: 12px; color: #606266; font-size: 14px; line-height: 1.6; }
.setup-steps { margin-top: 16px; padding: 12px; background: #f5f7fa; border-radius: 8px; }
.setup-steps h4 { margin: 0 0 8px; font-size: 14px; color: #303133; }
</style>
