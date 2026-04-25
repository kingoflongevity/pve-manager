<template>
  <div class="page-container">
    <h2>{{ t('dashboard.title') }}</h2>

    <!-- 节点信息 -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon><Cpu /></el-icon>
              <span>{{ t('dashboard.cpuUsage') }}</span>
            </div>
          </template>
          <div class="stat-value">25%</div>
          <el-progress :percentage="25" :color="progressColor" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon><Memo /></el-icon>
              <span>{{ t('dashboard.memoryUsage') }}</span>
            </div>
          </template>
          <div class="stat-value">4.2 / 16 GB</div>
          <el-progress :percentage="26" :color="progressColor" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon><Coin /></el-icon>
              <span>{{ t('dashboard.diskUsage') }}</span>
            </div>
          </template>
          <div class="stat-value">120 / 500 GB</div>
          <el-progress :percentage="24" :color="progressColor" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon><Timer /></el-icon>
              <span>{{ t('dashboard.uptime') }}</span>
            </div>
          </template>
          <div class="stat-value">15 天 8 小时</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 虚拟机/容器状态汇总 -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>{{ t('dashboard.vmStatusSummary') }}</span>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('dashboard.runningVMs')">3</el-descriptions-item>
            <el-descriptions-item :label="t('dashboard.stoppedVMs')">1</el-descriptions-item>
            <el-descriptions-item :label="t('dashboard.runningCTs')">2</el-descriptions-item>
            <el-descriptions-item :label="t('dashboard.stoppedCTs')">0</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>{{ t('dashboard.quickActions') }}</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary">{{ t('dashboard.createVM') }}</el-button>
            <el-button type="success">{{ t('dashboard.createCT') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 节点基础信息 -->
    <el-card class="mt-20">
      <template #header>
        <span>{{ t('dashboard.nodeInfo') }}</span>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="主机名">pve-node-01</el-descriptions-item>
        <el-descriptions-item :label="t('dashboard.version')">8.1.4</el-descriptions-item>
        <el-descriptions-item :label="t('dashboard.uptime')">15d 8h 32m</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Cpu, Memo, Coin, Timer } from '@element-plus/icons-vue'

const { t } = useI18n()

/**
 * 进度条颜色计算函数
 * 根据使用率返回不同的颜色
 */
function progressColor(percentage: number) {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}
</script>

<style lang="scss" scoped>
.mt-20 {
  margin-top: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.quick-actions {
  display: flex;
  gap: 12px;
}
</style>
