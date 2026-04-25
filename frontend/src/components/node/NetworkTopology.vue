<template>
  <div class="network-topology">
    <div class="topology-header">
      <h3 class="topology-title">网络拓扑图</h3>
      <div class="topology-legend">
        <span v-for="item in legendItems" :key="item.type" class="legend-item">
          <span class="legend-dot" :style="{ background: item.color }"></span>
          {{ item.label }}
        </span>
        <span class="legend-item">
          <span class="legend-dot legend-dot-active"></span>
          活跃
        </span>
        <span class="legend-item">
          <span class="legend-dot legend-dot-inactive"></span>
          非活跃
        </span>
      </div>
    </div>

    <div class="topology-container" ref="containerRef">
      <svg
        :width="svgWidth"
        :height="svgHeight"
        :viewBox="`0 0 ${svgWidth} ${svgHeight}`"
        class="topology-svg"
      >
        <!-- 连接线 -->
        <g class="topology-connections">
          <line
            v-for="(conn, index) in connections"
            :key="`conn-${index}`"
            :x1="conn.x1"
            :y1="conn.y1"
            :x2="conn.x2"
            :y2="conn.y2"
            :stroke="conn.color"
            :stroke-width="2"
            stroke-dasharray="5,5"
          />
        </g>

        <!-- 网络接口节点 -->
        <g class="topology-nodes">
          <g
            v-for="iface in positionedInterfaces"
            :key="iface.iface"
            :transform="`translate(${iface._x}, ${iface._y})`"
            class="topology-node"
            @click="handleNodeClick(iface)"
          >
            <!-- 接口形状 -->
            <rect
              :width="nodeWidth"
              :height="nodeHeight"
              :rx="6"
              :fill="getNodeTypeColor(iface.type)"
              :stroke="iface.active ? '#52c41a' : '#d9d9d9'"
              :stroke-width="2"
              class="node-shape"
            />
            <!-- 接口名称 -->
            <text
              x="0"
              y="-2"
              text-anchor="middle"
              dominant-baseline="middle"
              class="node-label"
            >
              {{ iface.iface }}
            </text>
            <!-- 接口类型 -->
            <text
              x="0"
              y="14"
              text-anchor="middle"
              dominant-baseline="middle"
              class="node-type-label"
            >
              {{ iface.type }}
            </text>
            <!-- 状态指示圆点 -->
            <circle
              :cx="nodeWidth / 2 - 5"
              :cy="-nodeHeight / 2 + 5"
              r="4"
              :fill="iface.active ? '#52c41a' : '#8c8c8c'"
              class="status-dot"
            />
          </g>
        </g>
      </svg>
    </div>

    <!-- 空状态 -->
    <el-empty
      v-if="positionedInterfaces.length === 0"
      description="暂无网络接口数据"
      :image-size="80"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import type { NetInterface } from '@/api/types'

interface Props {
  /** 网络接口列表 */
  interfaces: NetInterface[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'node-click': [iface: NetInterface]
}>()

// ============================================================
// 常量
// ============================================================

/** 节点宽度 */
const nodeWidth = 110
/** 节点高度 */
const nodeHeight = 40
/** 节点水平间距 */
const nodeGapX = 30
/** 节点垂直间距 */
const nodeGapY = 80
/** 画布内边距 */
const padding = 40

// ============================================================
// 图例项
// ============================================================

const legendItems = [
  { type: 'bridge', label: '网桥', color: '#1677ff' },
  { type: 'bond', label: '绑定', color: '#52c41a' },
  { type: 'vlan', label: 'VLAN', color: '#faad14' },
  { type: 'eth', label: '以太网', color: '#8c8c8c' },
]

// ============================================================
// 布局计算
// ============================================================

/** SVG 容器引用 */
const containerRef = ref<HTMLElement | null>(null)
/** SVG 宽度 */
const svgWidth = ref(800)
/** SVG 高度 */
const svgHeight = ref(300)

/**
 * 将网络接口按类型分组并排序
 */
const groupedInterfaces = computed(() => {
  const groups: Record<string, NetInterface[]> = {
    bridge: [],
    bond: [],
    vlan: [],
    eth: [],
    other: [],
  }

  for (const iface of props.interfaces) {
    const type = iface.type || 'other'
    if (groups[type]) {
      groups[type].push(iface)
    } else {
      groups.other.push(iface)
    }
  }

  // 过滤空组
  return Object.entries(groups).filter(([, items]) => items.length > 0)
})

/**
 * 计算每个接口的位置
 */
const positionedInterfaces = computed(() => {
  const result: (NetInterface & { _x: number; _y: number })[] = []
  let yOffset = padding

  for (const [type, items] of groupedInterfaces.value) {
    const totalWidth = items.length * nodeWidth + (items.length - 1) * nodeGapX
    const startX = (svgWidth.value - totalWidth) / 2

    items.forEach((item, index) => {
      result.push({
        ...item,
        _x: startX + index * (nodeWidth + nodeGapX) + nodeWidth / 2,
        _y: yOffset + nodeHeight / 2,
      })
    })

    yOffset += nodeHeight + nodeGapY
  }

  // 更新 SVG 高度
  svgHeight.value = yOffset + padding

  return result
})

/**
 * 计算连接线
 */
const connections = computed(() => {
  const conns: { x1: number; y1: number; x2: number; y2: number; color: string }[] = []

  // 从网桥/绑定指向其成员端口
  for (const bridge of positionedInterfaces.value.filter((i) => i.type === 'bridge')) {
    for (const eth of positionedInterfaces.value.filter((i) => i.type === 'eth')) {
      // 简单的启发式连接：网桥连接到物理接口
      if (bridge.bridge_ports && bridge.bridge_ports.includes(eth.iface)) {
        conns.push({
          x1: bridge._x,
          y1: bridge._y + nodeHeight / 2,
          x2: eth._x,
          y2: eth._y - nodeHeight / 2,
          color: '#52c41a',
        })
      }
    }
  }

  return conns
})

// ============================================================
// 工具函数
// ============================================================

/**
 * 获取接口类型对应的颜色
 */
function getNodeTypeColor(type: string): string {
  const map: Record<string, string> = {
    bridge: '#e8f3ff',
    bond: '#f6ffed',
    vlan: '#fffbe6',
    eth: '#f5f5f5',
  }
  return map[type] || '#f5f5f5'
}

/**
 * 处理节点点击
 */
function handleNodeClick(iface: NetInterface): void {
  emit('node-click', iface)
}

// ============================================================
// 响应式调整
// ============================================================

/**
 * 根据容器宽度调整 SVG 尺寸
 */
function updateSize(): void {
  if (containerRef.value) {
    svgWidth.value = containerRef.value.clientWidth
  }
}

onMounted(() => {
  updateSize()
  window.addEventListener('resize', updateSize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateSize)
})
</script>

<style lang="scss" scoped>
@use '@/assets/styles/variables' as *;

.network-topology {
  background: $color-bg-container;
  border-radius: $radius-base;
  padding: $spacing-5;
  box-shadow: $shadow-card;
  margin-top: $spacing-5;
}

.topology-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-5;
  flex-wrap: wrap;
  gap: $spacing-3;

  .topology-title {
    font-size: $font-size-base;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin: 0;
  }
}

.topology-legend {
  display: flex;
  align-items: center;
  gap: $spacing-4;
  flex-wrap: wrap;

  .legend-item {
    display: flex;
    align-items: center;
    gap: $spacing-1;
    font-size: $font-size-xs;
    color: $color-text-secondary;

    .legend-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
    }

    .legend-dot-active {
      background: #52c41a;
    }

    .legend-dot-inactive {
      background: #8c8c8c;
    }
  }
}

.topology-container {
  width: 100%;
  overflow-x: auto;
  min-height: 150px;
}

.topology-svg {
  display: block;
  margin: 0 auto;
}

.topology-node {
  cursor: pointer;
  transition: $transition-fast;

  &:hover {
    transform: scale(1.05);
  }

  .node-shape {
    transition: $transition-fast;
  }

  &:hover .node-shape {
    filter: brightness(0.95);
    stroke-width: 3;
  }

  .node-label {
    font-size: 13px;
    font-weight: 600;
    fill: $color-text-primary;
    font-family: $font-family-code;
  }

  .node-type-label {
    font-size: 10px;
    fill: $color-text-secondary;
    font-family: $font-family-code;
  }

  .status-dot {
    transition: $transition-fast;
  }
}
</style>
