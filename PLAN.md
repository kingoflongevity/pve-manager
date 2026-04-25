# PVE Web 管理面板开发计划

> 面向中国用户的 Proxmox VE 现代化 Web 管理面板

---

## 一、项目概述

### 1.1 项目背景

Proxmox VE (PVE) 原生的 Web UI 基于 ExtJS 7.x，存在以下问题：
- 界面风格老旧，不符合现代审美
- 移动端适配差
- 中文本地化不完善
- 缺乏轻量级部署方案
- 对新手用户学习成本高

本项目旨在构建一个现代化、轻量化、面向中国用户的 PVE Web 管理面板。

### 1.2 目标用户

- **个人用户/家庭实验室**: 需要简单易用的虚拟机/容器管理界面
- **中小企业运维**: 需要快速查看节点状态、资源使用情况
- **PVE 新手**: 需要更友好的引导和中文界面

### 1.3 核心价值

- 现代化 UI/UX，提升使用体验
- 完善的中文本地化
- 轻量级部署，低资源占用
- 常用功能快速访问（80/20 原则）

---

## 二、技术架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        浏览器 (SPA)                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  Frontend (Vue 3 + TS)                │  │
│  │  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌────────┐  │  │
│  │  |  页面   |  |  状态   |  |  API 层  |  | noVNC  |  │  │
│  │  └────┬────┘  └────┬────┘  └────┬─────┘  └───┬────┘  │  │
│  └───────┼────────────┼────────────┼──────────────┼───────┘  │
│          │            │            │              │          │
│          └────────────┴────────────┴──────────────┘          │
│                              │                               │
│                    HTTP / WebSocket                          │
│                              │                               │
│  ┌───────────────────────────┼───────────────────────────┐  │
│  │              Backend (Go + Gin)                        │  │
│  │  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌────────┐  │  │
│  │  |  API    |  |  认证   |  |  缓存    |  |  代理  |  │  │
│  │  |  代理   |  |  管理   |  |  层      |  |  层    |  │  │
│  │  └────┬────┘  └────┬────┘  └────┬─────┘  └───┬────┘  │  │
│  └───────┼────────────┼────────────┼──────────────┼───────┘  │
│          │            │            │              │          │
│          └────────────┴────────────┴──────────────┘          │
│                              │                               │
│              HTTPS (PVE API - localhost:8006)                │
│                              │                               │
└──────────────────────────────┼───────────────────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │   Proxmox VE Node   │
                    │   PVE API Server    │
                    └─────────────────────┘
```

### 2.2 技术栈选型

#### 前端技术栈

| 技术 | 版本 | 选择理由 |
|------|------|----------|
| **Vue 3** | 3.4+ | 中国开发者生态完善，学习曲线平缓，Composition API 灵活 |
| **TypeScript** | 5.x | 类型安全，提升大型项目可维护性 |
| **Vite** | 5.x | 极速开发体验，HMR 优秀 |
| **Element Plus** | 2.x | 国内最流行的 Vue 3 UI 库，组件丰富，文档完善，中文友好 |
| **Pinia** | 2.x | Vue 官方推荐状态管理，轻量易用 |
| **Vue Router** | 4.x | 官方路由方案 |
| **Axios** | 1.x | 成熟的 HTTP 客户端 |
| **noVNC** | 1.x | VNC 远程控制台支持 |
| **ECharts** | 5.x | 图表可视化，适合监控数据展示 |

#### 后端技术栈

| 技术 | 版本 | 选择理由 |
|------|------|----------|
| **Go** | 1.21+ | 编译为单文件，部署极简，并发性能优秀，适合代理场景 |
| **Gin** | 1.x | 轻量高性能 HTTP 框架，生态成熟 |
| **go-pveproxy** | - | PVE API 客户端封装（可自建或集成现有库） |
| **Redis** | 7.x (可选) | 缓存层，减少 PVE API 调用频率 |

#### 技术选型对比分析

**为什么选择 Go 而非 Python/Node.js 作为后端？**

| 维度 | Go | Python (FastAPI) | Node.js (Express) |
|------|-----|------------------|-------------------|
| 部署复杂度 | 单二进制文件 | 需 Python 环境 | 需 Node 环境 |
| 并发处理 | 原生 Goroutine | asyncio | 事件循环 |
| 内存占用 | ~10MB | ~30MB | ~20MB |
| 适合场景 | API 代理/网关 | 数据处理 | 前后端同栈 |
| 学习成本 | 中等 | 低 | 中等 |

**为什么选择 Element Plus 而非 Ant Design Vue / Arco Design？**

- Element Plus 社区最活跃，Star 数最高，问题解答多
- 组件覆盖全面，表格、表单、对话框等常用组件完善
- 中文文档最完整
- 主题定制简单

### 2.3 架构模式

#### 后端：API 代理模式

后端不作为数据源，而是作为 PVE REST API 的代理层：

1. **认证代理**: 处理 PVE 的 ticket/token 认证，转发认证请求
2. **请求代理**: 将前端请求转发至 PVE API
3. **响应转换**: 对 PVE 返回数据进行格式化处理（时间、单位等）
4. **缓存层**: 对低频变化数据（节点配置、存储信息）进行缓存
5. **WebSocket 代理**: 代理 VNC 控制台连接

#### 前端：SPA 模式

- 单页应用，路由按需加载
- API 层统一封装，支持拦截器处理认证、错误
- 响应式设计，适配桌面端和平板

---

## 三、核心功能模块

### 3.1 PVE API 全量功能映射表

基于 Proxmox VE REST API 完整端点结构，结合阿里云、腾讯云、vSphere 等云平台设计模式，制定以下功能映射表。

#### 优先级定义

| 优先级 | 含义 | 说明 |
|--------|------|------|
| P0 - Must | 核心必备 | 产品可用性的基础，缺失则产品无法使用 |
| P1 - Should | 重要功能 | 提升用户体验的关键功能，短期可用替代方案 |
| P2 - Could | 锦上添花 | 增强竞争力的功能，可后续迭代 |
| P3 - Won't | 暂不实现 | 低频场景或高风险功能，后续版本考虑 |

#### 状态定义

| 状态 | 含义 |
|------|------|
| 📋 Planned | 已规划，待开发 |
| 🚧 In Progress | 开发中 |
| ✅ Done | 已完成 |

#### 3.1.1 认证与访问控制 (/access)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| A01 | `POST /access/ticket` | 用户登录认证 | P0 | 📋 | 阿里云 RAM 登录 / 腾讯云 CAM 登录 |
| A02 | `GET/POST /access/users` | 用户管理（增删改查） | P2 | 📋 | 阿里云 RAM 用户管理 / vSphere SSO 用户 |
| A03 | `GET/POST /access/groups` | 用户组管理 | P2 | 📋 | 阿里云 RAM 用户组 / vSphere 组 |
| A04 | `GET/POST /access/roles` | 角色与权限模板 | P2 | 📋 | 阿里云 RAM 策略 / vSphere 角色 |
| A05 | `GET/PUT/DELETE /access/acl` | ACL 权限分配 | P2 | 📋 | 阿里云 RAM 授权 / vSphere 权限管理 |
| A06 | `GET/POST /access/domains` | 认证域管理（PAM/LDAP/AD） | P2 | 📋 | 阿里云 SSO / vSphere Identity Sources |
| A07 | `GET /access/password` | 修改密码 | P1 | 📋 | 通用云平台密码管理 |

#### 3.1.2 节点管理 (/nodes/{node})

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| N01 | `GET /nodes/{node}/status` | 节点状态概览（CPU/内存/磁盘/运行时间） | P0 | 📋 | 阿里云 ECS 实例概览 / vSphere Host Summary |
| N02 | `GET /nodes/{node}/version` | 节点版本信息 | P1 | 📋 | 云平台版本信息 |
| N03 | `GET /nodes/{node}/dns` | DNS 配置查看与修改 | P2 | 📋 | 阿里云 VPC DNS / vSphere DNS 设置 |
| N04 | `GET/POST/PUT/DELETE /nodes/{node}/net/network` | 网络接口管理 | P1 | 📋 | 阿里云 ENI / vSphere vSwitch |
| N05 | `GET /nodes/{node}/apt/update` | 系统更新检查 | P2 | 📋 | 云平台系统补丁管理 |
| N06 | `GET/PUT /nodes/{node}/services` | 系统服务状态管理 | P2 | 📋 | vSphere 服务管理 |
| N07 | `GET /nodes/{node}/syslog` | 系统日志查看 | P1 | 📋 | 阿里云 CloudLog / vSphere syslog |
| N08 | `GET /nodes/{node}/tasks` | 节点任务列表 | P1 | 📋 | 阿里云任务中心 / vSphere Tasks |
| N09 | `GET /nodes/{node}/tasks/{upid}/log` | 任务日志详情 | P1 | 📋 | 阿里云任务日志 / vSphere Task Details |
| N10 | `POST /nodes/{node}/execute` | 节点命令执行 | P3 | 📋 | 阿里云 RunCommand / vSphere ESXi Shell |
| N11 | `GET/PUT /nodes/{node}/time` | 时间与时区配置 | P2 | 📋 | 云平台 NTP 设置 |
| N12 | `GET /nodes/{node}/journal` | 系统日志流（实时） | P2 | 📋 | 阿里云 SLS 日志服务 / vSphere Journal |

#### 3.1.3 QEMU 虚拟机管理 (/nodes/{node}/qemu)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| Q01 | `GET /nodes/{node}/qemu` | 虚拟机列表 | P0 | 📋 | 阿里云 ECS 列表 / 腾讯云 CVM 列表 / vSphere VM Inventory |
| Q02 | `GET /nodes/{node}/qemu/{vmid}` | 虚拟机状态详情 | P0 | 📋 | 阿里云 ECS 详情 / vSphere VM Summary |
| Q03 | `POST /nodes/{node}/qemu` | 创建虚拟机 | P0 | 📋 | 阿里云创建 ECS / vSphere New VM Wizard |
| Q04 | `POST /nodes/{node}/qemu/{vmid}/status/start` | 启动虚拟机 | P0 | 📋 | 阿里云启动实例 |
| Q05 | `POST /nodes/{node}/qemu/{vmid}/status/stop` | 停止虚拟机 | P0 | 📋 | 阿里云停止实例 |
| Q06 | `POST /nodes/{node}/qemu/{vmid}/status/reboot` | 重启虚拟机 | P0 | 📋 | 阿里云重启实例 |
| Q07 | `POST /nodes/{node}/qemu/{vmid}/status/shutdown` | 优雅关机 | P0 | 📋 | 阿里云关机（ACPI） |
| Q08 | `POST /nodes/{node}/qemu/{vmid}/status/suspend` | 暂停虚拟机 | P1 | 📋 | 阿里云挂起实例 |
| Q09 | `POST /nodes/{node}/qemu/{vmid}/status/resume` | 恢复虚拟机 | P1 | 📋 | 阿里云恢复实例 |
| Q10 | `POST /nodes/{node}/qemu/{vmid}/status/reset` | 强制重置 | P1 | 📋 | 阿里云强制重启 |
| Q11 | `GET/PUT /nodes/{node}/qemu/{vmid}/config` | 虚拟机配置读写 | P0 | 📋 | 阿里云修改配置 / vSphere Edit Settings |
| Q12 | `POST /nodes/{node}/qemu/{vmid}/config` | 在线修改配置 | P1 | 📋 | 阿里云变配 / vSphere Hot Add |
| Q13 | `GET/POST /nodes/{node}/qemu/{vmid}/snapshot` | 快照管理（创建/列表/删除/回滚） | P1 | 📋 | 阿里云快照 / 腾讯云快照 / vSphere Snapshots |
| Q14 | `POST /nodes/{node}/qemu/{vmid}/clone` | 虚拟机克隆 | P1 | 📋 | 阿里云自定义镜像创建 / vSphere Clone |
| Q15 | `POST /nodes/{node}/qemu/{vmid}/move_disk` | 磁盘迁移 | P2 | 📋 | 阿里云磁盘变更存储 / vSphere Storage vMotion |
| Q16 | `POST /nodes/{node}/qemu/{vmid}/resize` | 磁盘扩容 | P1 | 📋 | 阿里云磁盘扩容 / vSphere Extend Disk |
| Q17 | `POST /nodes/{node}/qemu/{vmid}/vncproxy` | VNC 代理连接 | P0 | 📋 | 阿里云 VNC 控制台 / vSphere Remote Console |
| Q18 | `GET /nodes/{node}/qemu/{vmid}/vncwebsocket` | VNC WebSocket 连接 | P0 | 📋 | 阿里云 WebSocket 控制台 |
| Q19 | `GET /nodes/{node}/qemu/{vmid}/rrd` | 虚拟机监控数据 | P1 | 📋 | 阿里云 CloudMonitor / 腾讯云监控 |
| Q20 | `POST /nodes/{node}/qemu/{vmid}/agent/exec` | QEMU Agent 命令执行 | P2 | 📋 | 阿里云云助手 / vSphere Guest Operations |
| Q21 | `POST /nodes/{node}/qemu/{vmid}/migrate` | 虚拟机迁移（节点间） | P2 | 📋 | 阿里云迁移实例 / vSphere vMotion |
| Q22 | `GET /nodes/{node}/qemu/{vmid}/pending` | 待生效配置 | P2 | 📋 | vSphere Pending Changes |
| Q23 | `GET/PUT /nodes/{node}/qemu/{vmid}/firewall` | 虚拟机防火墙 | P2 | 📋 | 阿里云安全组 / vSphere Distributed Firewall |

#### 3.1.4 LXC 容器管理 (/nodes/{node}/lxc)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| L01 | `GET /nodes/{node}/lxc` | 容器列表 | P0 | 📋 | 阿里云 ECS（轻量）/ Docker 容器列表 |
| L02 | `POST /nodes/{node}/lxc` | 创建容器 | P0 | 📋 | 容器创建向导 |
| L03 | `POST /nodes/{node}/lxc/{vmid}/status/start` | 启动容器 | P0 | 📋 | 容器启动 |
| L04 | `POST /nodes/{node}/lxc/{vmid}/status/stop` | 停止容器 | P0 | 📋 | 容器停止 |
| L05 | `POST /nodes/{node}/lxc/{vmid}/status/reboot` | 重启容器 | P0 | 📋 | 容器重启 |
| L06 | `POST /nodes/{node}/lxc/{vmid}/status/shutdown` | 优雅关闭容器 | P1 | 📋 | 容器优雅关闭 |
| L07 | `POST /nodes/{node}/lxc/{vmid}/status/freeze` | 冻结容器 | P2 | 📋 | Docker Pause |
| L08 | `POST /nodes/{node}/lxc/{vmid}/status/unfreeze` | 解冻容器 | P2 | 📋 | Docker Unpause |
| L09 | `GET/PUT /nodes/{node}/lxc/{vmid}/config` | 容器配置读写 | P0 | 📋 | 容器配置管理 |
| L10 | `GET/POST /nodes/{node}/lxc/{vmid}/snapshot` | 容器快照管理 | P1 | 📋 | 容器快照 / vSphere CT Snapshots |
| L11 | `POST /nodes/{node}/lxc/{vmid}/clone` | 容器克隆 | P1 | 📋 | 容器克隆 |
| L12 | `POST /nodes/{node}/lxc/{vmid}/vncproxy` | 容器 VNC 控制台 | P1 | 📋 | 容器 Console |
| L13 | `GET /nodes/{node}/lxc/{vmid}/rrd` | 容器监控数据 | P1 | 📋 | 容器监控 |
| L14 | `GET /nodes/{node}/lxc/{vmid}/pending` | 待生效配置 | P2 | 📋 | vSphere Pending Changes |
| L15 | `GET/PUT /nodes/{node}/lxc/{vmid}/firewall` | 容器防火墙 | P2 | 📋 | 阿里云安全组 |
| L16 | `GET /nodes/{node}/lxc/{vmid}/features` | LXC 特性管理 | P2 | 📋 | 容器高级特性 |

#### 3.1.5 存储管理 (/nodes/{node}/storage, /storage)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| S01 | `GET /nodes/{node}/storage` | 节点存储列表 | P1 | 📋 | 阿里云磁盘列表 / vSphere Datastore Browser |
| S02 | `GET /nodes/{node}/storage/{storage}/content` | 存储内容（ISO/模板/VZDump） | P1 | 📋 | 阿里云镜像列表 / vSphere Datastore Content |
| S03 | `GET /nodes/{node}/storage/{storage}/status` | 存储状态 | P1 | 📋 | 阿里云磁盘状态 / vSphere Datastore Status |
| S04 | `GET/POST /storage` | 全局存储配置 | P2 | 📋 | 阿里云存储管理 / vSphere Storage Management |
| S05 | `GET /storage/{storage}/content` | 全局存储内容列表 | P2 | 📋 | 阿里云镜像管理 |
| S06 | `POST /storage/{storage}/content` | 上传镜像/模板 | P1 | 📋 | 阿里云导入镜像 / vSphere Upload OVF |

#### 3.1.6 集群管理 (/cluster)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| C01 | `GET /cluster/resources` | 集群资源总览 | P0 | 📋 | 阿里云资源编排概览 / vSphere vCenter Inventory |
| C02 | `GET /cluster/tasks` | 集群任务列表 | P1 | 📋 | 阿里云任务中心 / vSphere Recent Tasks |
| C03 | `GET /cluster/config` | 集群配置信息 | P2 | 📋 | vSphere Cluster Configuration |
| C04 | `GET/POST /cluster/jobs` | 定时任务管理 | P2 | 📋 | 阿里云定时任务 / vSphere Scheduled Tasks |
| C05 | `GET/POST /cluster/ha` | 高可用管理（HA Groups/Resources） | P2 | 📋 | 阿里云 HA / vSphere HA |
| C06 | `GET/POST /cluster/backup` | 备份管理 | P1 | 📋 | 阿里云自动快照策略 / vSphere Backup |
| C07 | `GET /cluster/metrics` | 指标服务器状态 | P2 | 📋 | 阿里云 CloudMonitor / vSphere vStats |
| C08 | `GET/PUT /cluster/firewall` | 集群防火墙 | P2 | 📋 | 阿里云安全组 / vSphere NSX Firewall |

#### 3.1.7 资源池管理 (/pools)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| R01 | `GET /pools` | 资源池列表 | P2 | 📋 | 阿里云资源组 / vSphere Resource Pool |
| R02 | `GET /pools/{poolid}` | 资源池详情 | P2 | 📋 | 阿里云资源组详情 / vSphere Pool Summary |
| R03 | `POST /pools` | 创建资源池 | P2 | 📋 | 阿里云创建资源组 / vSphere New Pool |
| R04 | `PUT/DELETE /pools/{poolid}` | 修改/删除资源池 | P2 | 📋 | 阿里云资源组管理 |

#### 3.1.8 复制管理 (/cluster/replication)

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| X01 | `GET /cluster/replication` | 复制任务列表 | P2 | 📋 | 阿里云跨地域复制 / vSphere Replication |
| X02 | `POST /cluster/replication` | 创建复制任务 | P2 | 📋 | 阿里云复制策略 / vSphere New Replication |

#### 3.1.9 软件定义网络 (/cluster/sdn) - PVE 8+

| # | API 端点 | 功能名称 | 优先级 | 状态 | 云平台对标 |
|---|----------|----------|--------|------|------------|
| D01 | `GET/POST /cluster/sdn/zones` | SDN 区域管理 | P3 | 📋 | 阿里云 VPC / vSphere NSX-T Segment |
| D02 | `GET/POST /cluster/sdn/vnets` | 虚拟网络管理 | P3 | 📋 | 阿里云 VSwitch / vSphere Logical Switch |
| D03 | `GET/POST /cluster/sdn/subnets` | 子网管理 | P3 | 📋 | 阿里云子网 / vSphere Subnet |
| D04 | `GET/POST /cluster/sdn/ips` | IP 地址管理 | P3 | 📋 | 阿里云 EIP / vSphere IP Pool |

### 3.2 功能统计与优先级分布

| 优先级 | 功能数量 | 占比 | 涉及模块 |
|--------|----------|------|----------|
| P0 - Must | 18 | 17.5% | 认证、虚拟机/容器基础操作、控制台、集群资源 |
| P1 - Should | 28 | 27.2% | 监控、快照、克隆、网络、存储、日志、备份 |
| P2 - Could | 44 | 42.7% | 用户权限、系统配置、迁移、高级功能 |
| P3 - Won't | 5 | 4.9% | SDN、远程命令执行 |
| **总计** | **95** | **100%** | 9 大模块 |

### 3.3 云平台设计模式对标分析

#### 3.3.1 阿里云模式参考

| PVE 功能 | 阿里云对标产品 | 设计借鉴 |
|----------|----------------|----------|
| 虚拟机管理 | ECS 控制台 | 列表筛选、生命周期操作、变配流程 |
| 镜像管理 | ECS 镜像 | 自定义镜像、共享镜像、镜像市场 |
| 快照管理 | 快照策略 | 自动快照、快照链、回滚操作 |
| 网络管理 | VPC 控制台 | 网络拓扑可视化、安全组规则 |
| 监控告警 | CloudMonitor | 时序图表、阈值告警、告警通知 |
| 用户权限 | RAM 控制台 | 用户-角色-权限三层模型 |
| 资源组 | 资源管理 | 按资源组筛选、资源池化管理 |

#### 3.3.2 腾讯云模式参考

| PVE 功能 | 腾讯云对标产品 | 设计借鉴 |
|----------|----------------|----------|
| 虚拟机管理 | CVM 控制台 | 快速创建向导、标签管理 |
| 容器管理 | TKE 控制台 | 容器列表、资源配置 |
| 存储管理 | CBS 控制台 | 磁盘列表、扩容操作 |
| 日志管理 | CLS 日志服务 | 日志检索、实时日志流 |

#### 3.3.3 vSphere 模式参考

| PVE 功能 | vSphere 对标功能 | 设计借鉴 |
|----------|------------------|----------|
| 集群管理 | vCenter Inventory | 资源树导航、数据中心层级 |
| 虚拟机管理 | VM Inventory | 文件夹组织、模板部署 |
| 存储管理 | Datastore Browser | 存储浏览器、文件管理 |
| 快照管理 | Snapshot Manager | 快照树、快照管理器 |
| 网络管理 | Distributed Switch | 网络拓扑、端口组管理 |
| 主机管理 | Host Summary | 主机硬件信息、服务状态 |

#### 3.3.4 本产品设计原则

基于以上云平台对标分析，本产品设计遵循以下原则：

1. **列表优先**: 采用阿里云/腾讯云的列表式设计，支持筛选、排序、搜索
2. **向导式创建**: 参考阿里云 ECS 创建流程，分步引导用户
3. **详情标签页**: 采用 vSphere 的标签页设计，基本信息、配置、监控、日志分开
4. **资源树导航**: 参考 vSphere 的左侧资源树，支持快速定位
5. **操作审计**: 参考阿里云操作审计，记录所有管理操作

---

## 四、开发阶段与里程碑

### 4.1 总体规划

总开发周期: **约 14-16 周** (单人开发, 覆盖 95 个功能点)

```
Phase 1              Phase 2              Phase 3              Phase 4              Phase 5
基础架构 + 认证      虚拟机/容器核心       控制台 + 存储        节点 + 网络 + 监控   扩展功能 + 优化
(3 周)                (4 周)                (3 周)                (3 周)                (1-3 周)
```

### 4.2 Phase 1: 基础架构 + 认证 (第 1-3 周)

**覆盖功能**: A01, A07, C01, N01

| 任务 | 工期 | 产出 | 对应功能 |
|------|------|------|----------|
| 项目初始化 | 1 天 | 前后端项目骨架 | - |
| 后端 API 代理层 | 2 天 | Go + Gin 代理框架 | - |
| 前端框架搭建 | 1 天 | Vue 3 项目 + 路由 + 状态管理 | - |
| UI 组件库集成 | 1 天 | Element Plus 集成 + 主题定制 | - |
| 认证模块 (A01) | 2 天 | 登录页、Ticket 管理、Token 认证 | A01 |
| 密码修改 (A07) | 1 天 | 修改密码页面 | A07 |
| API 层封装 | 1 天 | Axios 封装、拦截器、错误处理 | - |
| 布局框架 | 2 天 | 侧边栏资源树、顶栏、内容区 | - |
| 集群资源总览 (C01) | 1 天 | 资源概览卡片 | C01 |
| 仪表盘基础版 (N01) | 2 天 | 节点状态、资源使用图表 | N01 |

**里程碑 1 (v0.1)**: 用户能登录系统，查看节点状态和资源总览

### 4.3 Phase 2: 虚拟机/容器核心 (第 4-7 周)

**覆盖功能**: Q01-Q12, L01-L09

| 任务 | 工期 | 产出 | 对应功能 |
|------|------|------|----------|
| 虚拟机列表 (Q01) | 1 天 | 列表页、筛选、排序、搜索 | Q01 |
| 虚拟机详情 (Q02) | 1 天 | 基本信息、硬件配置标签页 | Q02 |
| 创建虚拟机 (Q03) | 3 天 | 分步向导表单 | Q03 |
| 生命周期操作 (Q04-Q07) | 1.5 天 | 启动/关机/重启/优雅关机 | Q04, Q05, Q06, Q07 |
| 高级操作 (Q08-Q10) | 1 天 | 暂停/恢复/强制重置 | Q08, Q09, Q10 |
| 配置读写 (Q11-Q12) | 2 天 | 配置编辑、在线修改 | Q11, Q12 |
| 容器列表 (L01) | 0.5 天 | 复用 VM 列表组件 | L01 |
| 创建容器 (L02) | 2 天 | 容器创建向导、模板选择 | L02 |
| 容器生命周期 (L03-L06) | 1 天 | 启动/停止/重启/优雅关闭 | L03-L06 |
| 容器配置 (L07-L09) | 1 天 | 配置查看/编辑、冻结/解冻 | L07-L09 |
| 任务状态跟踪 | 1.5 天 | TaskTracker、异步操作轮询 | 通用 |
| 批量操作 | 1 天 | 多选、批量启停 | Q01, L01 |
| 错误处理完善 | 1 天 | 统一错误提示、重试机制 | 通用 |

**里程碑 2 (v0.2)**: 能完成虚拟机/容器的创建、配置管理、生命周期操作完整流程

### 4.4 Phase 3: 控制台 + 存储 + 快照克隆 (第 8-10 周)

**覆盖功能**: Q13-Q19, L10-L13, S01-S06, Q17-Q18, L12

| 任务 | 工期 | 产出 | 对应功能 |
|------|------|------|----------|
| noVNC 控制台 (Q17-Q18) | 3 天 | VNC 代理、WebSocket 连接 | Q17, Q18 |
| 容器控制台 (L12) | 1 天 | 容器 VNC 支持 | L12 |
| 快照管理 (Q13) | 1.5 天 | 快照列表/创建/删除/回滚 | Q13 |
| 容器快照 (L10) | 0.5 天 | 复用快照组件 | L10 |
| 克隆功能 (Q14) | 1.5 天 | 完整克隆、链接克隆 | Q14 |
| 容器克隆 (L11) | 0.5 天 | 容器克隆 | L11 |
| 磁盘扩容 (Q16) | 1 天 | 磁盘大小调整 | Q16 |
| 虚拟机监控 (Q19) | 1 天 | RRD 图表 | Q19 |
| 容器监控 (L13) | 0.5 天 | 容器 RRD 图表 | L13 |
| 存储列表 (S01) | 1 天 | 存储类型、状态、使用量 | S01 |
| 存储内容 (S02) | 1.5 天 | ISO/模板/VZDump 管理 | S02 |
| 存储状态 (S03) | 0.5 天 | 存储空间详情 | S03 |
| 存储配置 (S04) | 2 天 | 添加存储、编辑、删除 | S04 |
| 上传镜像 (S06) | 1.5 天 | 文件上传、进度显示 | S06 |

**里程碑 3 (v0.3)**: 具备远程控制台、快照克隆、存储管理完整能力

### 4.5 Phase 4: 节点 + 网络 + 监控 + 任务 (第 11-13 周)

**覆盖功能**: N02-N09, N11-N12, C02, C06, N04

| 任务 | 工期 | 产出 | 对应功能 |
|------|------|------|----------|
| 节点信息 (N02) | 0.5 天 | 版本信息、运行时间 | N02 |
| 系统更新 (N05) | 1 天 | apt 更新检查 | N05 |
| 服务管理 (N06) | 1 天 | 服务状态列表、启停 | N06 |
| 系统日志 (N07) | 1.5 天 | 日志查看、筛选、搜索 | N07 |
| 实时日志流 (N12) | 1.5 天 | 日志实时滚动 | N12 |
| 任务列表 (N08, C02) | 1.5 天 | 节点/集群任务列表 | N08, C02 |
| 任务日志 (N09) | 1 天 | 任务状态、日志详情 | N09 |
| 时间配置 (N11) | 0.5 天 | NTP 设置、时区 | N11 |
| 网络接口 (N04) | 3 天 | 网络拓扑、接口配置 | N04 |
| 备份管理 (C06) | 2 天 | 备份列表、创建、恢复 | C06 |
| DNS 配置 (N03) | 1 天 | DNS 查看与修改 | N03 |

**里程碑 4 (v0.4)**: 具备完整的系统管理和运维能力

### 4.6 Phase 5: 扩展功能 + 优化 (第 14-16 周)

**覆盖功能**: A02-A06, C03-C05, C07-C08, Q20-Q23, L14-L16, R01-R04, X01-X02

| 任务 | 工期 | 产出 | 对应功能 |
|------|------|------|----------|
| 用户管理 (A02) | 1 天 | 用户增删改查 | A02 |
| 组管理 (A03) | 0.5 天 | 用户组管理 | A03 |
| 角色管理 (A04) | 1 天 | 角色与权限模板 | A04 |
| ACL 管理 (A05) | 1.5 天 | 权限分配 | A05 |
| 认证域 (A06) | 1 天 | PAM/LDAP/AD 配置 | A06 |
| 集群配置 (C03) | 1 天 | 集群信息查看 | C03 |
| 定时任务 (C04) | 1 天 | 定时任务管理 | C04 |
| 高可用 (C05) | 2 天 | HA 组/资源管理 | C05 |
| 指标服务 (C07) | 0.5 天 | Metrics 状态 | C07 |
| 集群防火墙 (C08) | 1 天 | 集群级防火墙 | C08 |
| Agent 命令 (Q20) | 1.5 天 | QEMU Agent 执行 | Q20 |
| 虚拟机迁移 (Q21) | 1.5 天 | 跨节点迁移 | Q21 |
| 磁盘迁移 (Q15) | 1 天 | 存储迁移 | Q15 |
| 待生效配置 (Q22, L14) | 0.5 天 | Pending 配置展示 | Q22, L14 |
| 防火墙 (Q23, L15) | 1.5 天 | VM/CT 级防火墙 | Q23, L15 |
| LXC 特性 (L16) | 0.5 天 | 高级特性管理 | L16 |
| 资源池 (R01-R04) | 1.5 天 | 资源池管理 | R01-R04 |
| 复制任务 (X01-X02) | 1.5 天 | 复制管理 | X01, X02 |
| 国际化 | 2 天 | 中英文切换 | - |
| 性能优化 | 2 天 | 缓存、懒加载、打包优化 | - |
| Bug 修复 | 持续 | - | - |

**里程碑 5 (v0.5)**: 达到可发布状态, 覆盖全部 P0/P1 功能

### 4.7 版本规划

| 版本 | 内容 | 覆盖功能 | 预计时间 |
|------|------|----------|----------|
| v0.1 | 基础架构 + 认证 + 仪表盘 | A01, A07, C01, N01 | 第 3 周 |
| v0.2 | 虚拟机/容器核心管理 | Q01-Q12, L01-L09 | 第 7 周 |
| v0.3 | 控制台 + 快照 + 克隆 + 存储 | Q13-Q19, L10-L13, S01-S06 | 第 10 周 |
| v0.4 | 节点管理 + 网络 + 任务 + 备份 | N02-N12, C02, C06, N04 | 第 13 周 |
| v0.5 | 用户权限 + 集群 + 高级功能 | A02-A06, C03-C08, Q20-Q23 | 第 16 周 |
| v1.0 | 正式发布 + SDN 预留 | D01-D04 (P3) | 第 18 周 |

### 4.8 版本特性矩阵

| 版本 | P0 功能 | P1 功能 | P2 功能 | 用户价值 |
|------|---------|---------|---------|----------|
| v0.1 | 8/18 | 0/28 | 0/44 | 可查看、可认证 |
| v0.2 | 15/18 | 4/28 | 0/44 | 可创建、可管理虚拟机 |
| v0.3 | 18/18 | 18/28 | 0/44 | 可远程操作、有快照备份 |
| v0.4 | 18/18 | 26/28 | 2/44 | 可运维管理 |
| v0.5 | 18/18 | 28/28 | 30/44 | 完整功能 |
| v1.0 | 18/18 | 28/28 | 44/44 | 全面覆盖 |

---

## 五、关键技术挑战与解决方案

### 5.1 认证与安全

**挑战**: PVE 认证机制复杂，Ticket 有效期短

**解决方案**:
```
┌─────────────────────────────────────────────────────┐
│                    认证流程                           │
├─────────────────────────────────────────────────────┤
│                                                     │
│  前端                          后端              PVE  │
│   │                            │                 │   │
│   │── 登录请求 ───────────────>│                 │   │
│   │                            │── ticket 请求 ──>│   │
│   │                            │<── ticket ───────│   │
│   │                            │                  │   │
│   │                            │── 缓存 ticket    │   │
│   │                            │   (内存/Redis)   │   │
│   │<── session token ──────────│                  │   │
│   │                            │                  │   │
│   │── API 请求 + token ───────>│                 │   │
│   │                            │── 验证 token     │   │
│   │                            │── 转发 + ticket >│   │
│   │                            │<── 响应 ─────────│   │
│   │<── 响应 ───────────────────│                  │   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

- 后端维护 PVE ticket 池，避免频繁认证
- 前端使用 session token 与后端通信
- Ticket 过期前自动刷新
- API Token 方式作为备选（更适合自动化场景）
- 敏感数据使用 AES 加密存储

### 5.2 noVNC 控制台

**挑战**: VNC WebSocket 连接需要特殊处理，跨域问题

**解决方案**:
- 后端作为 WebSocket 代理，转发 VNC 流量
- 使用 `gorilla/websocket` 库实现双向代理
- 前端使用 noVNC 库连接后端代理端口
- 连接参数通过 API 获取（ticket、port）

### 5.3 实时数据更新

**挑战**: 频繁轮询增加 PVE API 压力

**解决方案**:
- 分级刷新策略:
  - 高频数据（CPU/内存）: 10 秒轮询
  - 中频数据（状态列表）: 30 秒轮询
  - 低频数据（配置信息）: 按需加载
- 后端缓存层: 使用 Redis 缓存低频变化数据
- ETag/If-None-Match 减少无效传输
- 支持 WebSocket 推送（PVE 6.3+ 支持部分事件）

### 5.4 异步任务处理

**挑战**: PVE 的创建、迁移等操作是异步的，需要轮询状态

**解决方案**:
- 封装 TaskTracker 统一处理异步任务
- 创建操作后立即获取 UPID
- 轮询 `/api2/json/nodes/{node}/tasks/{upid}/status`
- 使用指数退避策略减少请求频率
- 操作完成后自动刷新相关数据

### 5.5 错误处理

**挑战**: PVE API 错误信息不够友好

**解决方案**:
- 后端统一拦截 PVE 错误，转换为中文友好提示
- 错误码映射表
- 前端统一错误提示组件
- 网络错误自动重试机制

### 5.6 大数据量处理

**挑战**: 大量虚拟机/容器时列表加载慢

**解决方案**:
- 虚拟列表 (Virtual Scroll) 优化长列表
- 分页加载
- 搜索和筛选在后端执行
- 按需加载详情数据

### 5.7 多节点管理

**挑战**: 一个面板管理多个 PVE 节点

**解决方案**:
- 连接配置持久化
- 节点切换时保留上下文
- 跨节点操作明确提示目标节点
- 集群模式下自动识别节点拓扑

---

## 六、目录结构

### 6.1 整体目录结构

```
pve_webui/
├── backend/                    # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # 程序入口
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   │   ├── config.go
│   │   │   └── config_test.go
│   │   ├── handler/           # HTTP 处理器
│   │   │   ├── auth.go
│   │   │   ├── proxy.go
│   │   │   ├── vnc.go
│   │   │   └── middleware.go
│   │   ├── pve/               # PVE API 客户端
│   │   │   ├── client.go
│   │   │   ├── auth.go
│   │   │   ├── qemu.go
│   │   │   ├── lxc.go
│   │   │   ├── node.go
│   │   │   ├── storage.go
│   │   │   └── types.go
│   │   ├── cache/             # 缓存层
│   │   │   └── cache.go
│   │   └── logger/            # 日志
│   │       └── logger.go
│   ├── pkg/                   # 可复用包
│   │   └── crypto/
│   │       └── aes.go
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile
│   └── config.yaml.example    # 配置示例
│
├── frontend/                   # Vue 3 前端
│   ├── public/
│   │   └── favicon.ico
│   ├── src/
│   │   ├── main.ts            # 入口文件
│   │   ├── App.vue            # 根组件
│   │   ├── api/               # API 层
│   │   │   ├── request.ts     # Axios 封装
│   │   │   ├── auth.ts
│   │   │   ├── qemu.ts
│   │   │   ├── lxc.ts
│   │   │   ├── node.ts
│   │   │   ├── storage.ts
│   │   │   └── types.ts       # API 类型定义
│   │   ├── assets/            # 静态资源
│   │   │   ├── images/
│   │   │   └── styles/
│   │   │       ├── variables.scss
│   │   │       └── global.scss
│   │   ├── components/        # 通用组件
│   │   │   ├── common/
│   │   │   │   ├── AppHeader.vue
│   │   │   │   ├── AppSidebar.vue
│   │   │   │   ├── AppLayout.vue
│   │   │   │   ├── LoadingOverlay.vue
│   │   │   │   └── ErrorBoundary.vue
│   │   │   ├── vm/
│   │   │   │   ├── VMStatusTag.vue
│   │   │   │   ├── VMActionMenu.vue
│   │   │   │   └── VMConfigForm.vue
│   │   │   └── monitor/
│   │   │       ├── ResourceChart.vue
│   │   │       └── StatusCard.vue
│   │   ├── composables/       # 组合式函数
│   │   │   ├── useAuth.ts
│   │   │   ├── useNodes.ts
│   │   │   ├── useTaskTracker.ts
│   │   │   └── useWebSocket.ts
│   │   ├── stores/            # Pinia 状态管理
│   │   │   ├── auth.ts
│   │   │   ├── nodes.ts
│   │   │   └── settings.ts
│   │   ├── views/             # 页面组件
│   │   │   ├── login/
│   │   │   │   ├── LoginView.vue
│   │   │   │   └── NodeSelector.vue
│   │   │   ├── dashboard/
│   │   │   │   └── DashboardView.vue
│   │   │   ├── qemu/
│   │   │   │   ├── QEMUListView.vue
│   │   │   │   ├── QEMUDetailView.vue
│   │   │   │   ├── QEMUCreateWizard.vue
│   │   │   │   └── QEMUConsoleView.vue
│   │   │   ├── lxc/
│   │   │   │   ├── LXCListView.vue
│   │   │   │   ├── LXCDetailView.vue
│   │   │   │   └── LXCCreateWizard.vue
│   │   │   ├── node/
│   │   │   │   ├── NodeListView.vue
│   │   │   │   └── NodeDetailView.vue
│   │   │   ├── storage/
│   │   │   │   ├── StorageListView.vue
│   │   │   │   └── StorageDetailView.vue
│   │   │   ├── network/
│   │   │   │   └── NetworkView.vue
│   │   │   ├── monitor/
│   │   │   │   └── MonitorView.vue
│   │   │   └── settings/
│   │   │       └── SettingsView.vue
│   │   ├── router/            # 路由配置
│   │   │   ├── index.ts
│   │   │   └── guards.ts      # 路由守卫
│   │   ├── utils/             # 工具函数
│   │   │   ├── format.ts      # 格式化函数
│   │   │   ├── crypto.ts      # 加密工具
│   │   │   └── constants.ts   # 常量定义
│   │   ├── locales/           # 国际化
│   │   │   ├── zh-CN.ts
│   │   │   └── en-US.ts
│   │   └── types/             # TypeScript 类型
│   │       └── index.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── pnpm-lock.yaml
│
├── docs/                       # 文档
│   ├── deployment.md           # 部署文档
│   ├── development.md          # 开发指南
│   └── api-reference.md        # API 参考
│
├── docker/                     # Docker 相关
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── nginx.conf
│
├── scripts/                    # 脚本
│   ├── build.sh
│   └── dev.sh
│
├── .gitignore
├── README.md
├── LICENSE
└── PLAN.md                     # 本文件
```

### 6.2 设计说明

**后端目录设计原则**:
- `cmd/`: 程序入口，保持精简
- `internal/`: 业务逻辑，不对外暴露
- `pkg/`: 可复用的工具包
- `handler/`: HTTP 路由处理
- `pve/`: PVE API 封装，与 PVE 文档结构对齐

**前端目录设计原则**:
- `api/`: 与后端 API 对应，统一出口
- `components/`: 按功能模块组织
- `views/`: 页面级组件，对应路由
- `composables/`: 可复用逻辑
- `stores/`: 全局状态
- `utils/`: 纯函数工具

---

## 七、开发规范

### 7.1 代码规范

**Go 后端**:
- 遵循 Go 官方代码规范
- 使用 `golangci-lint` 进行代码检查
- 函数必须有注释（导出函数必须有文档注释）
- 错误处理使用 `fmt.Errorf` + `%w` 包装
- 结构体命名使用名词，函数命名使用动词开头

**Vue 前端**:
- 使用 Composition API + `<script setup>` 语法
- 使用 TypeScript 严格模式
- 组件命名使用 PascalCase
- 文件命名使用 kebab-case（页面）或 PascalCase（组件）
- Props 必须定义类型和默认值

### 7.2 提交规范

使用 Conventional Commits 规范:

```
<type>(<scope>): <description>

type: feat | fix | docs | style | refactor | test | chore
scope: backend | frontend | auth | qemu | lxc | etc.
```

示例:
```
feat(qemu): 添加虚拟机创建向导
fix(auth): 修复 Ticket 过期处理逻辑
docs: 更新部署文档
```

### 7.3 分支策略

```
main (稳定版本)
├── develop (开发分支)
│   ├── feature/auth        (功能分支)
│   ├── feature/qemu-list
│   ├── fix/vnc-connection
│   └── refactor/api-layer
```

- `main`: 仅包含稳定版本
- `develop`: 日常开发分支
- `feature/*`: 功能开发，完成后合并到 develop
- 每个里程碑创建 Tag

---

## 八、测试策略

### 8.1 后端测试

| 测试类型 | 覆盖范围 | 工具 |
|----------|----------|------|
| 单元测试 | PVE API 客户端、工具函数 | Go testing + testify |
| 集成测试 | API 代理层、认证流程 | Go testing + httptest |
| Mock PVE | 模拟 PVE API 响应 | testify/mock |

### 8.2 前端测试

| 测试类型 | 覆盖范围 | 工具 |
|----------|----------|------|
| 单元测试 | Composables、Utils | Vitest |
| 组件测试 | 通用组件 | Vue Test Utils |
| E2E 测试 | 核心流程 | Playwright |

### 8.3 测试优先级

1. **P1**: 认证流程、API 代理
2. **P2**: 虚拟机生命周期操作
3. **P3**: noVNC 连接
4. **P4**: 其他功能

---

## 九、部署方案

### 9.1 开发环境

```bash
# 后端 (热重载)
cd backend && air

# 前端 (开发服务器)
cd frontend && pnpm dev
```

### 9.2 生产部署

**方案一: 单二进制 + 静态文件**
```
编译前端 -> 产出 dist/ -> 嵌入 Go 二进制 -> 单个可执行文件
```

**方案二: Docker 部署**
```yaml
services:
  pve-webui:
    build: .
    ports:
      - "3000:3000"
    environment:
      - PVE_HOST=192.168.1.100
      - PVE_PORT=8006
```

**方案三: 传统部署**
```
Nginx (静态文件 + 反向代理)
├── /api/* -> Go 后端 (:8080)
├── /ws/*  -> Go 后端 WebSocket
└── /*     -> 前端静态文件
```

### 9.3 部署要求

- Go 1.21+
- Node.js 18+ (仅构建)
- Redis (可选，用于缓存)
- 最低服务器配置: 1 核 512MB

---

## 十、风险评估与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| PVE API 变更 | 高 | 低 | 封装 API 层，隔离变更影响 |
| noVNC 兼容性问题 | 高 | 中 | 提前验证，准备替代方案 |
| 安全漏洞 | 高 | 中 | 定期安全审计，使用成熟加密库 |
| 性能瓶颈 | 中 | 低 | 压测验证，缓存优化 |
| 开发延期 | 中 | 中 | 分阶段发布，核心功能优先 |

---

## 十一、参考资源

### 11.1 官方文档
- [PVE API 文档](https://pve.proxmox.com/pve-docs/api-viewer/index.html)
- [noVNC 文档](https://novnc.com/info.html)
- [PVE 管理员指南](https://pve.proxmox.com/pve-docs/pve-admin-guide.html)

### 11.2 类似项目
- [pve-ui](https://github.com/example/pve-ui) - Django + Vue3 + Arco Design
- [proxmox-dashboard](https://github.com/example/proxmox-dashboard) - Node.js + 原生 JS

### 11.3 技术文档
- [Vue 3 官方文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/zh-CN/)
- [Gin 框架文档](https://gin-gonic.com/zh-cn/docs/)
- [ECharts 文档](https://echarts.apache.org/zh/index.html)

---

## 十二、下一步行动

### 立即执行 (本周)

1. [ ] 创建 Git 仓库，初始化前后端项目
2. [ ] 搭建后端基础框架 (Go + Gin)
3. [ ] 搭建前端基础框架 (Vue 3 + Vite)
4. [ ] 实现 PVE 认证流程
5. [ ] 完成项目目录结构搭建

### 下周计划

1. [ ] 完成 API 代理层
2. [ ] 实现前端布局框架
3. [ ] 开发仪表盘基础页面
4. [ ] 开始虚拟机列表功能

---

> **文档版本**: v2.0
> **创建日期**: 2026-04-25
> **最后更新**: 2026-04-25
> **作者**: 产品需求专家
> **变更说明**: 基于 PVE 完整 REST API 更新功能映射表，增加云平台对标分析，重新规划开发阶段
