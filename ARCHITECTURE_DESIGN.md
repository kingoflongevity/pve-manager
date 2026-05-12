# PVE WebUI 管理系统 - 功能扩展架构设计方案

> 文档版本: v1.0  
> 创建日期: 2026-05-05  
> 作者: 系统架构师

---

## 目录

1. [现有系统分析](#1-现有系统分析)
2. [功能一：文件夹浏览功能设计](#2-功能一文件夹浏览功能设计)
3. [功能二：AI 智能体集成设计](#3-功能二ai-智能体集成设计)
4. [功能三：应用商店设计方案](#4-功能三应用商店设计方案)
5. [数据库设计](#5-数据库设计)
6. [API 接口设计](#6-api-接口设计)
7. [前端组件设计](#7-前端组件设计)
8. [安全性考虑](#8-安全性考虑)
9. [实现优先级和阶段规划](#9-实现优先级和阶段规划)
10. [部署架构](#10-部署架构)

---

## 1. 现有系统分析

### 1.1 技术栈概览

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 后端语言 | Go 1.25 | 高性能、并发友好 |
| Web 框架 | Gin | 轻量级 HTTP 框架 |
| 数据库 | SQLite (GORM) | 嵌入式、轻量 |
| 日志 | Zap | 结构化日志 |
| 前端框架 | Vue 3 + TypeScript | Composition API |
| UI 组件库 | Element Plus | 企业级组件库 |
| 状态管理 | Pinia | Vue 3 推荐状态管理 |
| 构建工具 | Vite 6 | 快速开发体验 |

### 1.2 现有架构分层

```
backend/
├── cmd/server/main.go          # 入口文件
├── internal/
│   ├── client/pve/             # PVE API 客户端层
│   ├── handler/                # HTTP 处理器层
│   ├── service/                # 业务逻辑层
│   ├── repository/             # 数据访问层
│   ├── model/                  # 数据模型
│   └── database/               # 数据库初始化
├── pkg/crypto/                 # 加密工具包
└── data/                       # SQLite 数据文件

frontend/
├── src/
│   ├── api/                    # API 请求封装
│   ├── components/             # 可复用组件
│   ├── views/                  # 页面视图
│   ├── stores/                 # Pinia 状态管理
│   ├── router/                 # Vue Router 路由
│   └── composables/            # 组合式函数
```

### 1.3 现有 PVE API 覆盖范围

当前已实现的 PVE API 代理:
- 集群管理: 资源、任务、HA、SDN、Pool
- 节点操作: 状态、服务、日志、网络、DNS
- 存储管理: CRUD、内容浏览、ISO 下载
- QEMU 虚拟机: 全生命周期管理、快照、VNC
- LXC 容器: 全生命周期管理、快照、终端
- 访问控制: 用户、组、角色、ACL

---

## 2. 功能一：文件夹浏览功能设计

### 2.1 PVE 文件系统 API 分析

Proxmox VE 提供以下文件系统相关 API:

| API 端点 | 方法 | 说明 |
|----------|------|------|
| `/nodes/{node}/storage/{storage}/content` | GET | 获取存储内容列表 |
| `/nodes/{node}/storage/{storage}/download` | POST | 上传文件到存储 |
| `/nodes/{node}/download-file` | GET | 下载文件（需要 ticket） |
| `/nodes/{node}/lxc/{vmid}/download-file/directory` | GET | 下载 LXC 容器内目录（需 guest-agent） |
| `/nodes/{node}/lxc/{vmid}/upload` | POST | 上传文件到 LXC 容器 |

**关键限制**:
- PVE 原生不支持直接浏览 LXC/QEMU 内部文件系统
- 需要通过 **QEMU Guest Agent** 或 **LXC exec** 实现
- 文件上传/下载通过 PVE API 代理实现

### 2.2 架构设计

#### 2.2.1 总体方案

```
┌─────────────────────────────────────────────────────────────┐
│                        前端 (Vue 3)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ FileBrowser  │  │ FileViewer   │  │ FileOperationBar │   │
│  │ Component    │  │ (代码/图片)  │  │ (上传/下载/删除) │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
│         └─────────────────┼──────────────────┬─┘             │
└───────────────────────────┼──────────────────┼───────────────┘
                            │                  │
┌───────────────────────────┼──────────────────┼───────────────┐
│                    后端 (Go + Gin)            │               │
│  ┌──────────────┐  ┌─────┴──────┐  ┌────────┴─────────┐     │
│  │FileHandler   │  │FileService │  │  PVE Client      │     │
│  │(路由/鉴权)   │  │(业务逻辑)  │  │  (API 代理)      │     │
│  └──────────────┘  └────────────┘  └──────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                    Proxmox VE API                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ Storage API  │  │ LXC Exec API │  │ QEMU Guest Agent │   │
│  │ (存储内容)    │  │ (容器内操作)  │  │ (虚拟机内操作)    │   │
│  └──────────────┘  └──────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 实现方案

#### 方案 A：存储内容浏览（推荐首选）

浏览 PVE 存储（local、nfs、ceph 等）中的文件：

```
GET /api/pve/nodes/:node/storage/:storage/content?type=images
GET /api/pve/nodes/:node/storage/:storage/download?volume=xxx
```

**优点**: 
- PVE 原生支持，实现简单
- 适用于 ISO、VZDump、容器模板等
- 支持上传/下载/删除

**覆盖场景**:
- ISO 镜像管理
- 备份文件浏览
- 容器模板管理
- 磁盘镜像查看

#### 方案 B：LXC 容器内部文件浏览

通过 LXC exec 执行 `ls`、`cat` 等命令：

```
POST /nodes/{node}/lxc/{vmid}/exec
  { "command": "ls -la /path/to/dir" }
```

**优点**: 可以浏览容器内任意目录  
**缺点**: 需要 root 权限，安全风险较高

#### 方案 C：QEMU 虚拟机文件浏览

通过 QEMU Guest Agent 的 `guest-file-*` 命令：

```
POST /nodes/{node}/qemu/{vmid}/agent
  { "command": "guest-file-open", "arguments": { "path": "/etc/passwd" } }
POST /nodes/{node}/qemu/{vmid}/agent
  { "command": "guest-file-read", "arguments": { "handle": 1 } }
```

**前提条件**: 
- 虚拟机需安装 QEMU Guest Agent
- Guest Agent 服务需运行中

### 2.4 文件操作权限矩阵

| 操作 | 存储内容 | LXC 内部 | QEMU 内部 |
|------|---------|---------|----------|
| 浏览目录 | ✅ | ✅ (需 root) | ✅ (需 guest-agent) |
| 查看文件 | ✅ (文本) | ✅ | ✅ (文本) |
| 下载文件 | ✅ | ✅ | ✅ |
| 上传文件 | ✅ | ✅ | ❌ (不支持) |
| 删除文件 | ✅ | ✅ | ❌ (不支持) |
| 重命名 | ❌ | ✅ | ❌ |
| 新建目录 | ❌ | ✅ | ❌ |

---

## 3. 功能二：AI 智能体集成设计

### 3.1 架构总览

```
┌─────────────────────────────────────────────────────────────────┐
│                        前端 (Vue 3)                              │
│  ┌──────────┐  ┌──────────────┐  ┌──────────┐  ┌────────────┐  │
│  │ AI Chat  │  │ AI Dashboard │  │ Report   │  │ AI Config  │  │
│  │ Panel    │  │ (监控面板)    │  │ Viewer   │  │ (模型配置)  │  │
│  └────┬─────┘  └──────┬───────┘  └────┬─────┘  └─────┬──────┘  │
│       └────────────────┼───────────────┼──────────────┼─────────┘
└────────────────────────┼───────────────┼──────────────┼─────────┘
                         │               │              │
┌────────────────────────┼───────────────┼──────────────┼─────────┐
│                  后端 (Go + Gin)         │              │         │
│  ┌────────────┐  ┌─────┴──────┐  ┌─────┴─────┐  ┌────┴───────┐ │
│  │ AI Handler │  │ AI Service │  │ AI Agent  │  │ Report     │ │
│  │            │  │ (对话管理)  │  │ (工具链)  │  │ Service    │ │
│  └─────┬──────┘  └─────┬──────┘  └─────┬─────┘  └────┬───────┘ │
│        └───────────────┼───────────────┼──────────────┼─────────┘
└────────────────────────┼───────────────┼──────────────┼─────────┘
                         │               │              │
┌────────────────────────┼───────────────┼──────────────┼─────────┐
│              AI 服务提供商 (多模型适配)   │              │         │
│  ┌──────────┐  ┌───────┴──────┐  ┌─────┴──────┐  ┌────┴──────┐ │
│  │ OpenAI   │  │ Claude       │  │ 通义千问   │  │ 本地部署  │ │
│  │ API      │  │ API          │  │ API        │  │ Ollama    │ │
│  └──────────┘  └──────────────┘  └────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────────────────┘
                         │
┌────────────────────────┼─────────────────────────────────────────┐
│         PVE 数据收集层 (AI Agent 工具)        │                   │
│  ┌──────────┐  ┌───────┴──────┐  ┌────────────┐  ┌────────────┐ │
│  │ 资源采集 │  │ 性能指标采集 │  │ 日志采集   │  │ 告警采集   │ │
│  │ 器       │  │ 器           │  │ 器         │  │ 器         │ │
│  └──────────┘  └──────────────┘  └────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 大模型适配层设计

#### 3.2.1 统一接口定义

```go
// 定义统一的 LLM 接口
type LLMProvider interface {
    // 流式对话
    StreamChat(ctx context.Context, messages []Message, tools []Tool) (<-chan ChatResponse, error)
    // 非流式对话
    Chat(ctx context.Context, messages []Message, tools []Tool) (*ChatResponse, error)
    // 获取模型信息
    GetModelInfo() ModelInfo
}

// 统一消息格式
type Message struct {
    Role    string `json:"role"`    // system, user, assistant, tool
    Content string `json:"content"`
    Name    string `json:"name,omitempty"`
}

// 工具定义（Function Calling）
type Tool struct {
    Type     string     `json:"type"`
    Function ToolFunction `json:"function"`
}

type ToolFunction struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}
```

#### 3.2.2 支持的 LLM 提供商

| 提供商 | API 兼容 | 实现方式 | 配置项 |
|--------|---------|---------|--------|
| OpenAI | OpenAI | 标准 SDK | api_key, base_url, model |
| Azure OpenAI | OpenAI | 标准 SDK | api_key, endpoint, deployment, api_version |
| Claude | Anthropic | 专用 SDK | api_key, model |
| 通义千问 | OpenAI 兼容 | 标准 SDK | api_key, base_url, model |
| 智谱 AI | OpenAI 兼容 | 标准 SDK | api_key, base_url, model |
| 文心一言 | 专用 | 专用 SDK | api_key, secret_key, model |
| Ollama (本地) | OpenAI 兼容 | 标准 SDK | base_url, model |
| 自定义 | OpenAI 兼容 | 标准 SDK | api_key, base_url, model |

**策略**: 尽可能使用 OpenAI 兼容接口，减少适配代码。

#### 3.2.3 模型配置数据模型

```go
type AIModelConfig struct {
    ID           uint      `gorm:"primarykey" json:"id"`
    Name         string    `gorm:"size:100;not null;uniqueIndex" json:"name"`  // 显示名称
    Provider     string    `gorm:"size:50;not null" json:"provider"`           // openai/claude/qwen/ollama
    Model        string    `gorm:"size:100;not null" json:"model"`             // 模型名称
    BaseURL      string    `gorm:"size:500" json:"base_url"`                   // API 地址
    APIKey       string    `gorm:"size:500" json:"-"`                          // 加密存储
    APISecret    string    `gorm:"size:500" json:"-"`                          // 文心一言等需要
    IsDefault    bool      `gorm:"default:false" json:"is_default"`
    Enabled      bool      `gorm:"default:true" json:"enabled"`
    MaxTokens    int       `gorm:"default:4096" json:"max_tokens"`
    Temperature  float64   `gorm:"default:0.7" json:"temperature"`
    Timeout      int       `gorm:"default:60" json:"timeout"`                  // 秒
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### 3.3 AI 运维功能设计

#### 3.3.1 系统故障诊断

**工作流程**:
1. 收集系统状态数据（CPU、内存、磁盘、网络）
2. 收集最近错误日志
3. 构建诊断 Prompt
4. 调用 LLM 分析
5. 返回诊断结果和建议

**数据收集器**:
```go
type DiagnosticData struct {
    NodeStatus    map[string]NodeMetrics  `json:"node_status"`
    VMStatus      map[string]VMMetrics    `json:"vm_status"`
    StorageStatus map[string]StorageInfo  `json:"storage_status"`
    RecentErrors  []LogEntry              `json:"recent_errors"`
    TaskFailures  []TaskInfo              `json:"task_failures"`
    NetworkInfo   NetworkInfo             `json:"network_info"`
}
```

**Prompt 模板**:
```
你是一个专业的 Proxmox VE 运维专家。请分析以下系统状态数据，找出潜在问题并提供解决方案。

## 系统概览
- 节点数量: {node_count}
- 虚拟机总数: {vm_count}
- 存储池: {storage_count}

## 异常指标
{anomaly_metrics}

## 最近错误日志
{recent_errors}

## 请提供：
1. 问题诊断（按严重程度排序）
2. 根因分析
3. 解决建议（包含具体命令）
4. 预防措施
```

#### 3.3.2 配置建议

**场景**:
- 虚拟机资源配置优化（CPU、内存不足/过剩）
- 存储配置优化建议
- 网络配置建议
- 高可用配置检查

**实现方式**:
- 定期采集资源使用率
- 对比 PVE 最佳实践
- 生成优化建议报告

#### 3.3.3 问题排查助手

**AI Chat 交互**:
- 用户描述问题
- AI 自动收集相关数据
- 调用 PVE API 获取实时状态
- 提供排查步骤和解决方案

**工具链（Function Calling）**:
```go
var AITools = []Tool{
    {
        Name: "get_node_status",
        Description: "获取指定节点的状态信息（CPU、内存、磁盘、网络）",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "node": map[string]interface{}{
                    "type": "string",
                    "description": "节点名称",
                },
            },
            "required": []string{"node"},
        },
    },
    {
        Name: "get_vm_config",
        Description: "获取虚拟机的详细配置",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "node": map[string]interface{}{"type": "string"},
                "vmid": map[string]interface{}{"type": "integer"},
                "vm_type": map[string]interface{}{
                    "type": "string",
                    "enum": []string{"qemu", "lxc"},
                },
            },
            "required": []string{"node", "vmid", "vm_type"},
        },
    },
    {
        Name: "get_recent_logs",
        Description: "获取最近的系统日志",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "node": map[string]interface{}{"type": "string"},
                "limit": map[string]interface{}{"type": "integer", "default": 50},
                "level": map[string]interface{}{"type": "string", "enum": ["error", "warning", "info"]},
            },
        },
    },
    {
        Name: "get_storage_usage",
        Description: "获取存储使用情况和剩余空间",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "node": map[string]interface{}{"type": "string"},
                "storage": map[string]interface{}{"type": "string"},
            },
        },
    },
    {
        Name: "list_running_tasks",
        Description: "列出当前正在执行的任务",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "node": map[string]interface{}{"type": "string"},
            },
        },
    },
}
```

### 3.4 AI 监控功能设计

#### 3.4.1 异常检测

**检测维度**:
1. **资源异常**: CPU/内存/磁盘使用率突增或突降
2. **性能异常**: 响应时间、IOPS、网络延迟异常
3. **可用性异常**: 服务宕机、虚拟机异常重启
4. **安全异常**: 异常登录、未授权访问尝试

**实现方案**:

```go
// 异常检测规则
type AnomalyRule struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Name        string    `gorm:"size:100;not null" json:"name"`
    Metric      string    `json:"metric"`           // cpu_usage, memory_usage, disk_io, etc.
    Condition   string    `json:"condition"`        // >, <, >=, <=, ==
    Threshold   float64   `json:"threshold"`
    Duration    int       `json:"duration"`         // 持续时间（秒）
    Severity    string    `gorm:"size:20" json:"severity"`  // critical, warning, info
    Enabled     bool      `gorm:"default:true" json:"enabled"`
    CreatedAt   time.Time `json:"created_at"`
}

// 异常检测结果
type AnomalyResult struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    RuleID    uint      `json:"rule_id"`
    Node      string    `json:"node"`
    Resource  string    `json:"resource"`         // 资源标识 (vmid/container_id)
    Metric    string    `json:"metric"`
    Value     float64   `json:"value"`
    Threshold float64   `json:"threshold"`
    Severity  string    `json:"severity"`
    Message   string    `gorm:"type:text" json:"message"`  // AI 生成的分析
    AITips    string    `gorm:"type:text" json:"ai_tips"`  // AI 建议
    DetectedAt time.Time `json:"detected_at"`
    Acked     bool      `gorm:"default:false" json:"acked"`
}
```

#### 3.4.2 趋势预测

**预测场景**:
- 磁盘容量预测（预计多久存满）
- 资源使用趋势预测
- 性能瓶颈预测

**实现方式**:
- 基于历史 RRD 数据
- 使用 LLM 的趋势分析能力
- 可选：集成简单的时间序列预测算法

#### 3.4.3 智能告警

**告警流程**:
```
数据采集 → 规则匹配 → 异常检测 → AI 分析 → 告警通知 → 用户处理
```

**通知渠道**:
- Webhook（钉钉、企业微信、飞书）
- 邮件
- 站内消息

**告警降噪**:
- 告警聚合（相同问题合并）
- 告警抑制（维护期间不告警）
- 告警升级（未处理自动升级）

### 3.5 报告生成功能

#### 3.5.1 报告类型

| 报告类型 | 生成周期 | 内容 |
|---------|---------|------|
| 运维日报 | 每天 | 资源使用概况、异常事件、任务完成情况 |
| 性能周报 | 每周 | 性能趋势、瓶颈分析、优化建议 |
| 安全月报 | 每月 | 安全事件、访问审计、漏洞扫描 |
| 容量规划 | 每月 | 容量趋势、扩容建议、成本分析 |
| 自定义报告 | 按需 | 用户指定维度的分析报告 |

#### 3.5.2 报告数据结构

```go
type AIReport struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Name        string    `gorm:"size:200;not null" json:"name"`
    Type        string    `gorm:"size:50;not null" json:"type"`  // daily/weekly/monthly/custom
    Status      string    `gorm:"size:20;default:pending" json:"status"`
    Content     string    `gorm:"type:longtext" json:"content"`   // Markdown 格式
    DataRange   string    `json:"data_range"`                     // 2024-01-01 ~ 2024-01-31
    GeneratedAt time.Time `json:"generated_at"`
    CreatedAt   time.Time `json:"created_at"`
}

type ReportSchedule struct {
    ID         uint      `gorm:"primarykey" json:"id"`
    Name       string    `gorm:"size:100;not null" json:"name"`
    Type       string    `json:"type"`
    CronExpr   string    `gorm:"size:50" json:"cron_expr"`    // cron 表达式
    Enabled    bool      `gorm:"default:true" json:"enabled"`
    Recipients []string  `json:"recipients"`                  // 邮件/Webhook 地址
    CreatedAt  time.Time `json:"created_at"`
}
```

#### 3.5.3 报告生成 Prompt

```
你是 Proxmox VE 运维专家，请根据以下数据生成专业的运维报告。

## 报告周期
{start_date} 至 {end_date}

## 集群概况
- 节点数: {node_count}
- 虚拟机数: {vm_count}
- 容器数: {lxc_count}

## 资源使用情况
{resource_metrics}

## 事件统计
- 任务总数: {task_count}
- 成功: {success_count}
- 失败: {failed_count}

## 异常事件
{anomaly_events}

## 请生成以下章节：
1. 执行摘要（1-2段话）
2. 资源使用分析（含图表描述）
3. 异常事件分析
4. 性能趋势
5. 优化建议
6. 下周重点关注

报告要求：专业、数据驱动、 actionable 建议。
```

---

## 4. 功能三：应用商店设计方案

### 4.1 市场竞品分析

#### 4.1.1 现有方案

| 方案 | 特点 | 适用场景 |
|------|------|---------|
| Proxmox VE 官方 | 无官方应用商店 | - |
| TurnKey Linux | 提供预配置 LXC 模板 | 快速部署常见应用 |
| LXC Image Server | 官方 LXC 镜像仓库 | 基础模板 |
| Proxmox VE Helper Scripts | 社区脚本集合 | 一键安装应用 |
| Cloud-Init | 虚拟机初始化 | 自定义部署 |

#### 4.1.2 设计思路

结合以上方案优点，设计内部应用商店:
- **模板管理**: 提供预配置的 LXC/QEMU 模板
- **一键部署**: 简化部署流程
- **配置管理**: 支持自定义参数
- **版本管理**: 模板版本控制

### 4.2 应用商店架构

```
┌──────────────────────────────────────────────────────────────┐
│                       前端 (Vue 3)                            │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────────────┐   │
│  │ App Store  │  │ App Detail   │  │ Deploy Wizard       │   │
│  │ (应用列表) │  │ (应用详情)    │  │ (部署向导)           │   │
│  └─────┬──────┘  └──────┬───────┘  └──────────┬──────────┘   │
│        └────────────────┼─────────────────────┬─┘             │
└─────────────────────────┼─────────────────────┼───────────────┘
                          │                     │
┌─────────────────────────┼─────────────────────┼───────────────┐
│                   后端 (Go + Gin)             │               │
│  ┌────────────┐  ┌──────┴──────┐  ┌──────────┴──────────┐    │
│  │App Handler │  │ App Service │  │ Template Manager    │    │
│  │            │  │             │  │ (模板管理)           │    │
│  └─────┬──────┘  └──────┬──────┘  └──────────┬──────────┘    │
│        └────────────────┼────────────────────┬─┘              │
└─────────────────────────┼────────────────────┼────────────────┘
                          │                     │
┌─────────────────────────┼─────────────────────┼────────────────┐
│                   应用模板仓库                 │                │
│  ┌────────────┐  ┌──────┴──────┐  ┌──────────┴──────────┐    │
│  │ 内置模板   │  │ 远程仓库    │  │ 自定义模板           │    │
│  │ (预置应用)  │  │ (Git/S3)   │  │ (用户上传)           │    │
│  └────────────┘  └─────────────┘  └─────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

### 4.3 应用模板设计

#### 4.3.1 模板定义格式 (YAML)

```yaml
# 应用模板定义
app:
  id: "nginx-proxy-manager"
  name: "Nginx Proxy Manager"
  version: "2.10.4"
  description: "可视化的 Nginx 反向代理管理工具"
  category: "networking"
  tags: ["proxy", "nginx", "web"]
  icon: "https://example.com/icons/npm.png"
  
  # 部署类型
  type: "lxc"  # lxc | qemu
  
  # 资源配置
  resources:
    cpu: 2
    memory: 2048    # MB
    disk: 10        # GB
    swap: 512       # MB
  
  # 网络配置
  network:
    ports:
      - container: 80
        host: 80
        protocol: tcp
      - container: 443
        host: 443
        protocol: tcp
      - container: 81
        host: 81
        protocol: tcp
  
  # 环境变量/配置参数
  config:
    - name: "TZ"
      label: "时区"
      type: "string"
      default: "Asia/Shanghai"
      required: false
    - name: "DB_PASSWORD"
      label: "数据库密码"
      type: "password"
      default: ""
      required: true
    - name: "DISK_SIZE"
      label: "数据盘大小 (GB)"
      type: "number"
      default: 10
      required: false
  
  # 存储配置
  storage:
    template: "local:vztmpl/ubuntu-22.04-standard_22.04-1_amd64.tar.zst"
    rootfs: "local-lvm"
  
  # 部署后脚本
  post_deploy: |
    #!/bin/bash
    apt-get update
    apt-get install -y curl
    curl -sL https://npm-install.example.com/setup.sh | bash
  
  # 依赖
  dependencies:
    - "docker"
  
  # 作者信息
  author: "Internal Team"
  homepage: "https://github.com/NginxProxyManager/nginx-proxy-manager"
  license: "MIT"
```

#### 4.3.2 应用分类

```go
var AppCategories = []AppCategory{
    {ID: "infrastructure", Name: "基础设施", Icon: "server", Color: "#409EFF"},
    {ID: "database", Name: "数据库", Icon: "coin", Color: "#67C23A"},
    {ID: "networking", Name: "网络服务", Icon: "connection", Color: "#E6A23C"},
    {ID: "monitoring", Name: "监控运维", Icon: "monitor", Color: "#F56C6C"},
    {ID: "devops", Name: "开发运维", Icon: "tools", Color: "#909399"},
    {ID: "media", Name: "媒体服务", Icon: "video-camera", Color: "#E6A23C"},
    {ID: "storage", Name: "存储服务", Icon: "files", Color: "#409EFF"},
    {ID: "security", Name: "安全服务", Icon: "lock", Color: "#F56C6C"},
    {ID: "other", Name: "其他", Icon: "menu", Color: "#909399"},
}
```

### 4.4 应用商店数据模型

```go
type AppTemplate struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    AppID       string    `gorm:"size:100;not null;uniqueIndex" json:"app_id"`
    Name        string    `gorm:"size:200;not null" json:"name"`
    Version     string    `gorm:"size:50;not null" json:"version"`
    Description string    `gorm:"type:text" json:"description"`
    Category    string    `gorm:"size:50" json:"category"`
    Tags        string    `gorm:"type:text" json:"tags"`            // JSON 数组
    Icon        string    `gorm:"size:500" json:"icon"`
    Type        string    `gorm:"size:20;not null" json:"type"`     // lxc/qemu
    Template    string    `gorm:"type:text;not null" json:"template"` // YAML 模板内容
    Source      string    `gorm:"size:50;default:built-in" json:"source"` // built-in/remote/custom
    Status      string    `gorm:"size:20;default:active" json:"status"`
    DeployCount int       `gorm:"default:0" json:"deploy_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AppDeployment struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    AppID       string    `gorm:"size:100;not null;index" json:"app_id"`
    Node        string    `gorm:"size:100;not null" json:"node"`
    VMID        int       `json:"vmid"`
    VMType      string    `gorm:"size:10" json:"vm_type"`
    Status      string    `gorm:"size:20;default:deploying" json:"status"`
    Config      string    `gorm:"type:text" json:"config"`          // 部署时配置 (JSON)
    UPID        string    `gorm:"size:100" json:"upid"`
    Error       string    `gorm:"type:text" json:"error"`
    DeployedAt  time.Time `json:"deployed_at"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 4.5 部署流程

```
1. 用户选择应用 → 2. 配置参数 → 3. 选择节点/存储 → 4. 提交部署
                                      ↓
5. 后端验证参数 → 6. 生成 PVE API 请求 → 7. 调用 PVE 创建容器/虚拟机
                                      ↓
8. 执行 post_deploy 脚本 → 9. 更新状态 → 10. 通知用户
```

---

## 5. 数据库设计

### 5.1 新增数据表

```sql
-- AI 模型配置表
CREATE TABLE ai_model_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,           -- 显示名称
    provider TEXT NOT NULL,              -- 提供商: openai/claude/qwen/ollama/wenxin
    model TEXT NOT NULL,                 -- 模型名称
    base_url TEXT,                       -- API 地址
    api_key TEXT NOT NULL,               -- 加密存储
    api_secret TEXT,                     -- 额外密钥（文心一言等）
    is_default INTEGER DEFAULT 0,        -- 是否默认
    enabled INTEGER DEFAULT 1,           -- 是否启用
    max_tokens INTEGER DEFAULT 4096,
    temperature REAL DEFAULT 0.7,
    timeout INTEGER DEFAULT 60,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI 对话记录表
CREATE TABLE ai_conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,               -- 用户 ID
    title TEXT,                          -- 对话标题
    scenario TEXT,                       -- 场景: ops_troubleshoot/config_advice/monitoring
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI 对话消息表
CREATE TABLE ai_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role TEXT NOT NULL,                  -- user/assistant/system/tool
    content TEXT NOT NULL,               -- 消息内容
    tool_calls TEXT,                     -- 工具调用 (JSON)
    tool_results TEXT,                   -- 工具结果 (JSON)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id)
);

-- 异常检测规则表
CREATE TABLE anomaly_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    metric TEXT NOT NULL,                -- cpu_usage/memory_usage/disk_io/etc.
    condition TEXT NOT NULL,             -- >/</>=/<=/==
    threshold REAL NOT NULL,
    duration INTEGER DEFAULT 60,         -- 持续时间（秒）
    severity TEXT DEFAULT 'warning',     -- critical/warning/info
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 异常检测结果表
CREATE TABLE anomaly_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    node TEXT NOT NULL,
    resource TEXT,                       -- 资源标识
    metric TEXT NOT NULL,
    value REAL NOT NULL,
    threshold REAL NOT NULL,
    severity TEXT NOT NULL,
    message TEXT,                        -- AI 分析结果
    ai_tips TEXT,                        -- AI 建议
    detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    acked INTEGER DEFAULT 0,
    FOREIGN KEY (rule_id) REFERENCES anomaly_rules(id)
);

-- AI 报告表
CREATE TABLE ai_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,                  -- daily/weekly/monthly/custom
    status TEXT DEFAULT 'pending',       -- pending/generating/completed/failed
    content TEXT,                        -- Markdown 内容
    data_range TEXT,                     -- 数据范围
    generated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 报告定时任务表
CREATE TABLE report_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    cron_expr TEXT NOT NULL,             -- cron 表达式
    enabled INTEGER DEFAULT 1,
    recipients TEXT,                     -- JSON 数组
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 应用模板表
CREATE TABLE app_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    category TEXT,
    tags TEXT,                           -- JSON 数组
    icon TEXT,
    type TEXT NOT NULL,                  -- lxc/qemu
    template TEXT NOT NULL,              -- YAML 模板内容
    source TEXT DEFAULT 'built-in',      -- built-in/remote/custom
    status TEXT DEFAULT 'active',
    deploy_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 应用部署记录表
CREATE TABLE app_deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id TEXT NOT NULL,
    node TEXT NOT NULL,
    vmid INTEGER,
    vm_type TEXT,
    status TEXT DEFAULT 'deploying',
    config TEXT,                         -- JSON 配置
    upid TEXT,
    error TEXT,
    deployed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES app_templates(app_id)
);
```

### 5.2 数据模型 (Go)

完整的 Go 数据模型见 `models.go` 文件更新方案。

---

## 6. API 接口设计

### 6.1 文件夹浏览 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/pve/nodes/:node/storage/:storage/content` | 获取存储内容列表 |
| GET | `/api/pve/nodes/:node/storage/:storage/download` | 下载存储中的文件 |
| POST | `/api/pve/nodes/:node/storage/:storage/upload` | 上传文件到存储 |
| DELETE | `/api/pve/nodes/:node/storage/:storage/delete` | 删除存储中的文件 |
| POST | `/api/pve/nodes/:node/lxc/:vmid/exec` | 在 LXC 内执行命令 |
| POST | `/api/pve/nodes/:node/qemu/:vmid/agent` | 通过 guest-agent 操作文件 |

### 6.2 AI 功能 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/ai/models` | 获取模型配置列表 |
| POST | `/api/ai/models` | 创建模型配置 |
| PUT | `/api/ai/models/:id` | 更新模型配置 |
| DELETE | `/api/ai/models/:id` | 删除模型配置 |
| POST | `/api/ai/models/:id/test` | 测试模型连接 |
| GET | `/api/ai/models/:id/models` | 获取可用模型列表 |
| POST | `/api/ai/chat` | 发起 AI 对话 |
| POST | `/api/ai/chat/stream` | 流式 AI 对话 (SSE) |
| GET | `/api/ai/conversations` | 获取对话列表 |
| GET | `/api/ai/conversations/:id` | 获取对话详情 |
| DELETE | `/api/ai/conversations/:id` | 删除对话 |
| POST | `/api/ai/diagnose` | 系统故障诊断 |
| POST | `/api/ai/suggest` | 获取配置建议 |
| GET | `/api/ai/anomalies` | 获取异常检测结果 |
| POST | `/api/ai/anomalies/:id/ack` | 确认异常告警 |
| GET | `/api/ai/reports` | 获取报告列表 |
| POST | `/api/ai/reports/generate` | 生成报告 |
| GET | `/api/ai/reports/:id` | 获取报告详情 |
| GET | `/api/ai/reports/schedules` | 获取报告定时任务 |
| POST | `/api/ai/reports/schedules` | 创建报告定时任务 |

### 6.3 应用商店 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/apps` | 获取应用列表 |
| GET | `/api/apps/:appId` | 获取应用详情 |
| GET | `/api/apps/categories` | 获取应用分类 |
| POST | `/api/apps/:appId/deploy` | 部署应用 |
| GET | `/api/apps/deployments` | 获取部署列表 |
| GET | `/api/apps/deployments/:id` | 获取部署详情 |
| DELETE | `/api/apps/deployments/:id` | 卸载应用 |
| POST | `/api/apps/import` | 导入自定义模板 |
| GET | `/api/apps/templates/sync` | 同步远程模板 |

---

## 7. 前端组件设计

### 7.1 文件夹浏览组件

```
frontend/src/components/file-browser/
├── FileBrowser.vue              # 主组件（文件浏览器）
├── FileBreadcrumb.vue           # 面包屑导航
├── FileList.vue                 # 文件列表（表格/网格视图切换）
├── FileItem.vue                 # 单个文件/目录项
├── FileViewer.vue               # 文件查看器（代码高亮/图片预览）
├── FileUploadDialog.vue         # 上传对话框
├── FileActions.vue              # 文件操作栏（下载/删除/重命名）
└── FileContextMenu.vue          # 右键菜单
```

**核心组件设计**:

```vue
<!-- FileBrowser.vue 核心结构 -->
<template>
  <div class="file-browser">
    <FileBreadcrumb :path="currentPath" @navigate="navigateTo" />
    
    <div class="file-toolbar">
      <el-button @click="showUploadDialog">上传</el-button>
      <el-button @click="refresh">刷新</el-button>
      <el-radio-group v-model="viewMode">
        <el-radio-button value="list">列表</el-radio-button>
        <el-radio-button value="grid">网格</el-radio-button>
      </el-radio-group>
    </div>
    
    <FileList
      :files="files"
      :view-mode="viewMode"
      :loading="loading"
      @open="openFileOrDir"
      @download="downloadFile"
      @delete="deleteFile"
      @contextmenu="showContextMenu"
    />
    
    <FileViewer
      v-if="previewFile"
      :file="previewFile"
      @close="closePreview"
    />
  </div>
</template>
```

### 7.2 AI 功能组件

```
frontend/src/components/ai/
├── AIChatPanel.vue              # AI 对话面板（侧边栏/弹窗）
├── AIChatMessage.vue            # 单条消息组件
├── AIChatInput.vue              # 输入框（支持快捷命令）
├── AIToolCall.vue               # 工具调用展示
├── AIMarkdown.vue               # Markdown 渲染
├── AIDiagnoseDialog.vue         # 故障诊断对话框
├── AIReportViewer.vue           # 报告查看器
├── AIModelConfig.vue            # 模型配置管理
├── AIAnomalyList.vue            # 异常检测列表
└── AIScenarioSelector.vue      # 场景选择器
```

**路由设计**:

```typescript
{
  path: 'ai',
  name: 'AICenter',
  component: () => import('@/views/ai/AICenterView.vue'),
  meta: { title: 'AI 智能中心' },
  children: [
    { path: 'chat', name: 'AIChat', component: () => import('@/views/ai/ChatView.vue') },
    { path: 'diagnose', name: 'AIDiagnose', component: () => import('@/views/ai/DiagnoseView.vue') },
    { path: 'monitoring', name: 'AIMonitoring', component: () => import('@/views/ai/MonitoringView.vue') },
    { path: 'reports', name: 'AIReports', component: () => import('@/views/ai/ReportsView.vue') },
    { path: 'settings', name: 'AISettings', component: () => import('@/views/ai/SettingsView.vue') },
  ]
}
```

### 7.3 应用商店组件

```
frontend/src/components/app-store/
├── AppStore.vue                 # 应用商店主页
├── AppCard.vue                  # 应用卡片
├── AppDetail.vue                # 应用详情
├── AppCategoryFilter.vue        # 分类筛选
├── AppSearchBar.vue             # 搜索栏
├── DeployWizard.vue             # 部署向导（多步骤表单）
├── DeploymentList.vue           # 部署列表
└── ImportTemplateDialog.vue     # 导入模板对话框
```

**部署向导步骤**:

```
Step 1: 选择应用（已完成）→ Step 2: 配置参数 → Step 3: 选择节点/存储 → Step 4: 确认部署
```

### 7.4 新增视图文件

```
frontend/src/views/
├── ai/
│   ├── AICenterView.vue         # AI 中心主页面
│   ├── ChatView.vue             # AI 对话
│   ├── DiagnoseView.vue         # 故障诊断
│   ├── MonitoringView.vue       # AI 监控
│   ├── ReportsView.vue          # 报告中心
│   └── SettingsView.vue         # AI 设置
├── file-browser/
│   ├── StorageBrowserView.vue   # 存储浏览器
│   └── LXCBrowserView.vue       # LXC 文件浏览器
└── app-store/
    ├── AppStoreView.vue         # 应用商店
    └── DeploymentsView.vue      # 部署管理
```

---

## 8. 安全性考虑

### 8.1 文件系统安全

| 风险 | 防护措施 |
|------|---------|
| 目录遍历攻击 | 严格校验路径，不允许 `..` 跳转 |
| 未授权访问 | JWT 鉴权 + PVE 权限检查 |
| 敏感文件泄露 | 黑名单过滤 `/etc/shadow` 等敏感文件 |
| 大文件上传 DoS | 限制上传大小（默认 500MB） |
| 命令注入 | LXC exec 命令白名单校验 |

### 8.2 AI 功能安全

| 风险 | 防护措施 |
|------|---------|
| API Key 泄露 | 加密存储，不在前端返回完整 Key |
| Prompt 注入 | 输入过滤，系统 Prompt 保护 |
| 敏感数据泄露 | 数据脱敏后再发送给 LLM |
| 滥用/超额 | 请求频率限制，Token 用量统计 |
| 模型返回有害建议 | AI 输出免责声明，人工确认机制 |

### 8.3 应用商店安全

| 风险 | 防护措施 |
|------|---------|
| 恶意模板 | 模板签名验证，来源审查 |
| 资源耗尽 | 部署配额限制 |
| 端口冲突 | 自动检测端口可用性 |
| 后部署脚本安全 | 沙箱执行，权限限制 |

### 8.4 通用安全措施

```go
// 请求频率限制中间件
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
    // 基于 IP 和用户的滑动窗口限流
}

// 操作审计日志
func AuditLogger() gin.HandlerFunc {
    // 记录所有写操作
}

// 数据脱敏
func SanitizeForAI(data interface{}) interface{} {
    // 移除密码、密钥等敏感字段
}
```

---

## 9. 实现优先级和阶段规划

### Phase 1: 基础功能（2-3 周）

| 任务 | 预计时间 | 依赖 |
|------|---------|------|
| **P1-1**: 存储内容浏览 API + 前端组件 | 3 天 | 现有存储 API 扩展 |
| **P1-2**: 文件上传/下载功能 | 2 天 | P1-1 |
| **P1-3**: AI 模型配置管理（CRUD） | 2 天 | 独立 |
| **P1-4**: AI Chat 基础对话（OpenAI 兼容） | 3 天 | P1-3 |
| **P1-5**: 应用商店基础框架 + 内置模板 | 4 天 | 独立 |

**交付物**:
- 存储文件浏览器
- AI 对话基础功能（接入 OpenAI）
- 应用商店页面 + 模板管理

### Phase 2: AI 增强（2-3 周）

| 任务 | 预计时间 | 依赖 |
|------|---------|------|
| **P2-1**: AI Agent 工具链（PVE API 调用） | 4 天 | P1-4 |
| **P2-2**: 故障诊断功能 | 3 天 | P2-1 |
| **P2-3**: 多模型适配（Claude/通义/智谱） | 2 天 | P1-3 |
| **P2-4**: 异常检测规则 + 告警 | 3 天 | 独立 |
| **P2-5**: AI 报告生成 | 3 天 | P2-1 |

**交付物**:
- AI 运维助手（带工具调用）
- 多模型支持
- 异常检测和报告

### Phase 3: 高级功能（2-3 周）

| 任务 | 预计时间 | 依赖 |
|------|---------|------|
| **P3-1**: LXC 容器内部文件浏览 | 3 天 | 独立 |
| **P3-2**: QEMU guest-agent 文件操作 | 4 天 | 独立 |
| **P3-3**: 应用一键部署 | 3 天 | P1-5 |
| **P3-4**: 报告定时任务 + 通知 | 2 天 | P2-5 |
| **P3-5**: 趋势预测功能 | 3 天 | P2-4 |

**交付物**:
- 完整文件浏览（含 LXC/QEMU 内部）
- 应用商店部署功能
- 定时报告

### Phase 4: 优化和完善（1-2 周）

| 任务 | 预计时间 |
|------|---------|
| 性能优化（缓存、分页） | 2 天 |
| 安全加固 | 2 天 |
| E2E 测试 | 2 天 |
| 文档完善 | 1 天 |

---

## 10. 部署架构

### 10.1 目录结构扩展

```
backend/
├── internal/
│   ├── client/
│   │   ├── pve/
│   │   └── llm/                    # LLM 客户端（新增）
│   │       ├── provider.go         # 统一接口
│   │       ├── openai.go           # OpenAI 实现
│   │       ├── claude.go           # Claude 实现
│   │       └── wenxin.go           # 文心一言实现
│   ├── handler/
│   │   ├── file_browser.go         # 文件浏览 Handler（新增）
│   │   ├── ai_chat.go              # AI 对话 Handler（新增）
│   │   ├── ai_config.go            # AI 配置 Handler（新增）
│   │   ├── ai_report.go            # AI 报告 Handler（新增）
│   │   └── app_store.go            # 应用商店 Handler（新增）
│   ├── service/
│   │   ├── file_browser_service.go # 文件浏览 Service（新增）
│   │   ├── ai_service.go           # AI Service（新增）
│   │   ├── ai_agent.go             # AI Agent 工具链（新增）
│   │   ├── report_service.go       # 报告 Service（新增）
│   │   ├── anomaly_service.go      # 异常检测 Service（新增）
│   │   └── app_store_service.go    # 应用商店 Service（新增）
│   ├── model/
│   │   └── models.go               # 扩展数据模型
│   └── scheduler/
│       └── report_scheduler.go     # 定时任务（新增）

frontend/src/
├── api/
│   ├── fileBrowser.ts              # 文件浏览 API
│   ├── ai.ts                       # AI API
│   └── appStore.ts                 # 应用商店 API
├── components/
│   ├── file-browser/               # 文件浏览组件
│   ├── ai/                         # AI 组件
│   └── app-store/                  # 应用商店组件
└── views/
    ├── ai/                         # AI 页面
    ├── file-browser/               # 文件浏览页面
    └── app-store/                  # 应用商店页面
```

### 10.2 新增 Go 依赖

```go
require (
    // LLM 客户端
    github.com/sashabaranov/go-openai v1.30.0    // OpenAI 兼容 SDK
    github.com/liushuangls/go-anthropic/v2 v2.0.0 // Claude SDK
    
    // 定时任务
    github.com/robfig/cron/v3 v3.0.1              // Cron 调度
    
    // YAML 解析
    gopkg.in/yaml.v3 v3.0.1                       // 已有
    
    // Markdown 处理（后端生成报告用）
    github.com/yuin/goldmark v1.7.0               // Markdown 解析
)
```

---

## 附录

### A. PVE 文件系统 API 详细文档

详见 Proxmox VE API 文档: https://pve.proxmox.com/pve-docs/api-viewer/

### B. 推荐 LLM 模型

| 场景 | 推荐模型 | 原因 |
|------|---------|------|
| 日常对话 | GPT-4o-mini / Claude Sonnet | 性价比 |
| 深度分析 | GPT-4o / Claude Opus | 分析能力强 |
| 中文场景 | 通义千问-Max / 智谱 GLM-4 | 中文理解 |
| 本地部署 | Llama 3.1 / Qwen 2.5 (Ollama) | 数据不出网 |

### C. 应用模板示例

推荐内置以下应用模板:
1. **Nginx Proxy Manager** - 反向代理
2. **Portainer** - Docker 管理
3. **Grafana + Prometheus** - 监控
4. **GitLab CE** - 代码托管
5. **Jenkins** - CI/CD
6. **Nextcloud** - 云存储
7. **Gitea** - 轻量 Git 服务
8. **Vaultwarden** - 密码管理
9. **Home Assistant** - 智能家居
10. **MinIO** - 对象存储
