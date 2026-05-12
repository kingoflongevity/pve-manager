<template>
  <div class="ai-reports">
    <div class="page-header">
      <h2>AI 报告中心</h2>
      <div class="header-actions">
        <el-select v-model="reportType" size="small" style="width: 140px; margin-right: 8px" @change="fetchReports">
          <el-option label="全部类型" value="" />
          <el-option label="运维日报" value="daily" />
          <el-option label="性能周报" value="weekly" />
          <el-option label="安全月报" value="monthly" />
          <el-option label="自定义" value="custom" />
        </el-select>
        <el-button type="primary" @click="showGenerateDialog">生成报告</el-button>
      </div>
    </div>

    <el-table :data="reports" stripe @row-dblclick="viewDetail">
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column label="类型" width="120">
        <template #default="{ row }">
          <el-tag :type="reportTag(row.type)" size="small">{{ reportLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="生成时间" width="180" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">查看</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="reportDialog" title="报告详情" width="800px">
      <div class="report-content" v-html="renderMarkdown(currentReport?.content || '')"></div>
    </el-dialog>

    <el-dialog v-model="genDialog" title="生成报告" width="500px">
      <el-form :model="genForm" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="genForm.title" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="genForm.type" style="width: 100%">
            <el-option label="运维日报" value="daily" />
            <el-option label="性能周报" value="weekly" />
            <el-option label="安全月报" value="monthly" />
            <el-option label="自定义报告" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="genForm.content" type="textarea" :rows="4" placeholder="可选：描述报告重点关注内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="genDialog = false">取消</el-button>
        <el-button type="primary" :loading="genLoading" @click="doGenerate">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getReports, generateReport, getReport } from '@/api/ai'
import type { AIReport } from '@/api/ai'

const reports = ref<AIReport[]>([])
const reportType = ref('')
const reportDialog = ref(false)
const currentReport = ref<AIReport | null>(null)
const genDialog = ref(false)
const genLoading = ref(false)
const genForm = ref({ title: '', type: 'daily', content: '' })

onMounted(() => fetchReports())

async function fetchReports() {
  try {
    reports.value = await getReports(reportType.value || undefined)
  } catch { ElMessage.error('获取报告列表失败') }
}

function reportTag(t: string) {
  const m: Record<string, string> = { daily: 'primary', weekly: 'success', monthly: 'warning', custom: 'info' }
  return m[t] || ''
}

function reportLabel(t: string) {
  const m: Record<string, string> = { daily: '运维日报', weekly: '性能周报', monthly: '安全月报', custom: '自定义' }
  return m[t] || t
}

async function viewDetail(row: AIReport) {
  try {
    currentReport.value = row.id ? await getReport(row.id) : row
    reportDialog.value = true
  } catch { ElMessage.error('获取报告详情失败') }
}

function showGenerateDialog() {
  Object.assign(genForm.value, { title: '', type: 'daily', content: '' })
  genDialog.value = true
}

async function doGenerate() {
  genLoading.value = true
  try {
    await generateReport(genForm.value)
    ElMessage.success('报告生成任务已提交')
    genDialog.value = false
    fetchReports()
  } catch { ElMessage.error('生成失败') }
  finally { genLoading.value = false }
}

function renderMarkdown(text: string): string {
  return text
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/\n/g, '<br>')
    .replace(/```(\w*)\n?([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/^## (.+)$/gm, '<h3>$1</h3>')
    .replace(/^### (.+)$/gm, '<h4>$1</h4>')
}
</script>

<style scoped>
.ai-reports { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 20px; }
.header-actions { display: flex; align-items: center; }
.report-content { padding: 20px; line-height: 1.8; font-size: 14px; }
.report-content :deep(h3) { margin: 16px 0 8px; color: #303133; }
.report-content :deep(h4) { margin: 12px 0 6px; color: #606266; }
.report-content :deep(pre) { background: #282c34; color: #abb2bf; padding: 12px; border-radius: 6px; overflow-x: auto; }
.report-content :deep(code) { background: #f0f2f5; padding: 2px 6px; border-radius: 4px; }
</style>
