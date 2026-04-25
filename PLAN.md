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

### 3.1 模块优先级矩阵

采用 MoSCoW 优先级框架：

| 模块 | 优先级 | 用户价值 | 开发难度 | 预计工期 |
|------|--------|----------|----------|----------|
| M1: 认证与连接 | Must | 高 | 中 | 3 天 |
| M2: 仪表盘概览 | Must | 高 | 低 | 4 天 |
| M3: 虚拟机管理 | Must | 高 | 高 | 7 天 |
| M4: 容器管理 | Must | 高 | 高 | 5 天 |
| M5: 节点管理 | Should | 中 | 中 | 4 天 |
| M6: 存储管理 | Should | 中 | 中 | 4 天 |
| M7: noVNC 控制台 | Must | 高 | 高 | 5 天 |
| M8: 网络管理 | Should | 中 | 高 | 5 天 |
| M9: 监控与告警 | Should | 中 | 中 | 4 天 |
| M10: 用户与权限 | Could | 低 | 中 | 3 天 |
| M11: 任务日志 | Could | 低 | 低 | 2 天 |
| M12: 系统设置 | Could | 低 | 低 | 2 天 |

### 3.2 功能模块详细设计

#### M1: 认证与连接 (Must)

**功能描述**:
- 支持 PVE 用户名/密码认证（获取 ticket）
- 支持 API Token 认证（推荐，更安全）
- 多 PVE 节点配置管理
- 认证状态保持与自动刷新
- 连接测试

**API 依赖**:
- `POST /api2/json/access/ticket` - 获取认证 ticket
- `GET /api2/json/access/ticket` - 验证 ticket 有效性

**技术要点**:
- Ticket 有效期 2 小时，需自动续期或提醒重新登录
- API Token 格式: `USER@REALM!TOKENID=UUID`
- 敏感信息加密存储（localStorage 中使用 AES 加密）

**验收标准**:
- 用户可通过用户名/密码或 API Token 登录
- 登录成功后能正确调用 PVE API
- Token 过期后能正确提示并跳转登录
- 支持保存多个 PVE 节点连接配置

---

#### M2: 仪表盘概览 (Must)

**功能描述**:
- 节点基础信息（主机名、PVE 版本、运行时间）
- 资源使用概览（CPU、内存、存储、网络）
- 虚拟机/容器状态汇总（运行中、已停止、异常）
- 快捷操作入口（创建 VM/CT、重启、关机）
- 实时数据刷新（可配置间隔）

**API 依赖**:
- `GET /api2/json/nodes/{node}/status` - 节点状态
- `GET /api2/json/cluster/resources` - 集群资源列表
- `GET /api2/json/nodes/{node}/qemu` - 虚拟机列表
- `GET /api2/json/nodes/{node}/lxc` - 容器列表

**技术要点**:
- 使用 ECharts 绘制资源使用图表
- 定时轮询 + WebSocket 实时更新
- 数据缓存减少 API 调用

**验收标准**:
- 页面加载后 3 秒内展示完整概览
- 数据每 30 秒自动刷新
- 资源使用率可视化展示
- 支持手动刷新

---

#### M3: 虚拟机管理 (Must)

**功能描述**:
- 虚拟机列表（支持筛选、排序、搜索）
- 生命周期操作：启动、关机、重启、强制停止、暂停、恢复
- 创建虚拟机（向导式，支持 ISO/模板）
- 编辑虚拟机配置（CPU、内存、磁盘、网络）
- 虚拟机详情（基本信息、硬件配置、云初始化）
- 快照管理（创建、删除、回滚）
- 克隆/迁移
- 批量操作

**API 依赖**:
- `GET /api2/json/nodes/{node}/qemu` - 虚拟机列表
- `POST /api2/json/nodes/{node}/qemu` - 创建虚拟机
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/start` - 启动
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/stop` - 关机
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/reboot` - 重启
- `GET/PUT /api2/json/nodes/{node}/qemu/{vmid}/config` - 查看/修改配置
- `POST /api2/json/nodes/{node}/qemu/{vmid}/snapshot` - 创建快照
- `GET /api2/json/nodes/{node}/qemu/{vmid}/snapshot` - 快照列表

**技术要点**:
- 创建虚拟机使用分步表单（向导模式）
- 异步操作使用 Task 轮询获取状态
- 配置修改使用差异对比展示

**验收标准**:
- 列表支持分页、筛选、搜索
- 所有生命周期操作可用
- 能正确展示操作结果和错误信息
- 创建虚拟机流程完整

---

#### M4: 容器管理 (Must)

**功能描述**:
- 容器列表（类似虚拟机列表）
- 生命周期操作：启动、停止、重启
- 创建容器（支持模板选择）
- 编辑容器配置
- 容器详情
- 快照管理
- 克隆

**API 依赖**:
- `GET /api2/json/nodes/{node}/lxc` - 容器列表
- `POST /api2/json/nodes/{node}/lxc` - 创建容器
- `POST /api2/json/nodes/{node}/lxc/{vmid}/status/start` - 启动
- `GET/PUT /api2/json/nodes/{node}/lxc/{vmid}/config` - 查看/修改配置
- `POST /api2/json/nodes/{node}/lxc/{vmid}/snapshot` - 创建快照

**技术要点**:
- 与虚拟机管理复用大部分组件
- CT 和 QEMU 的差异通过策略模式处理

**验收标准**:
- 功能与虚拟机管理一致
- CT 特有功能（模板选择、unprivileged 选项）可用

---

#### M5: 节点管理 (Should)

**功能描述**:
- 节点信息（系统信息、版本、订阅状态）
- 系统更新（apt 更新检查）
- 服务状态（pve-cluster, pvedaemon 等）
- 系统日志查看
- DNS/时间/时区配置

**API 依赖**:
- `GET /api2/json/nodes/{node}/status` - 节点状态
- `GET /api2/json/nodes/{node}/version` - 版本信息
- `GET /api2/json/nodes/{node}/apt/update` - 可用更新
- `GET /api2/json/nodes/{node}/services` - 服务列表
- `GET /api2/json/nodes/{node}/syslog` - 系统日志

**验收标准**:
- 能查看节点完整信息
- 服务状态实时显示
- 日志支持筛选和搜索

---

#### M6: 存储管理 (Should)

**功能描述**:
- 存储列表（类型、状态、使用量）
- 存储详情（内容类型、空间使用）
- 添加存储（目录、NFS、LVM、ZFS 等）
- 编辑/删除存储
- ISO/模板镜像管理
- 磁盘使用可视化

**API 依赖**:
- `GET /api2/json/storage` - 存储列表
- `POST /api2/json/storage` - 添加存储
- `GET /api2/json/nodes/{node}/storage/{storage}/content` - 存储内容
- `GET /api2/json/nodes/{node}/storage/{storage}/status` - 存储状态

**验收标准**:
- 支持常见存储类型配置
- 空间使用可视化展示
- 镜像文件管理可用

---

#### M7: noVNC 控制台 (Must)

**功能描述**:
- 集成 noVNC 实现虚拟机/容器远程控制台
- 全屏模式支持
- 剪贴板同步
- 键盘快捷键（Ctrl+Alt+Del 等）
- 多标签控制台

**API 依赖**:
- `POST /api2/json/nodes/{node}/qemu/{vmid}/vncproxy` - 创建 VNC 代理
- WebSocket 连接: `wss://<host>:8006/api2/json/nodes/{node}/qemu/{vmid}/vncwebsocket`

**技术要点**:
- 需要先调用 vncproxy 获取 ticket 和 port
- WebSocket 连接需要携带认证 cookie
- 后端需要代理 WebSocket 连接
- 处理跨域 WebSocket 连接

**验收标准**:
- 能正常连接并显示控制台画面
- 键盘输入正常响应
- 全屏切换正常
- 连接断开后能重新连接

---

#### M8: 网络管理 (Should)

**功能描述**:
- 网络接口列表和拓扑图
- 接口配置（IP、网关、VLAN、Bond）
- 创建/编辑网络接口
- 网络配置预览（应用前）
- 网络流量监控

**API 依赖**:
- `GET /api2/json/nodes/{node}/network` - 网络接口列表
- `POST /api2/json/nodes/{node}/network` - 创建接口
- `PUT /api2/json/nodes/{node}/network/{iface}` - 修改接口

**技术要点**:
- 网络拓扑图使用 ECharts Graph 或自定义 SVG
- 网络配置需要应用+重启网络才能生效

**验收标准**:
- 能查看网络拓扑
- 能创建和修改网络接口
- 配置预览功能可用

---

#### M9: 监控与告警 (Should)

**功能描述**:
- CPU/内存/存储/网络历史图表
- 自定义监控时间范围
- 节点间资源对比
- 告警规则配置（阈值告警）
- 告警通知（邮件、Webhook）

**API 依赖**:
- `GET /api2/json/nodes/{node}/rrd` - RRD 监控数据
- `GET /api2/json/nodes/{node}/qemu/{vmid}/rrd` - VM 监控数据
- `GET /api2/json/nodes/{node}/lxc/{vmid}/rrd` - CT 监控数据

**技术要点**:
- RRD 数据格式转换
- ECharts 时间轴图表
- 告警规则本地存储+定时检查

**验收标准**:
- 历史数据图表展示正常
- 支持多种时间范围（1h, 6h, 24h, 7d, 30d, 1y）
- 告警功能可用

---

#### M10: 用户与权限 (Could)

**功能描述**:
- 用户管理（创建、编辑、删除）
- 组管理
- 角色与权限管理
- 资源池管理
- 认证域配置

**API 依赖**:
- `GET/POST /api2/json/access/users` - 用户管理
- `GET/POST /api2/json/access/groups` - 组管理
- `GET/POST /api2/json/access/acl` - ACL 管理
- `GET/POST /api2/json/pools` - 资源池管理

**验收标准**:
- 完整用户管理功能
- ACL 配置可用

---

#### M11: 任务日志 (Could)

**功能描述**:
- 任务列表（所有异步任务）
- 任务状态跟踪
- 任务日志查看
- 任务筛选（按类型、状态、时间）

**API 依赖**:
- `GET /api2/json/nodes/{node}/tasks` - 任务列表
- `GET /api2/json/nodes/{node}/tasks/{upid}/status` - 任务状态
- `GET /api2/json/nodes/{node}/tasks/{upid}/log` - 任务日志

**验收标准**:
- 任务列表实时刷新
- 能查看完整任务日志

---

#### M12: 系统设置 (Could)

**功能描述**:
- 界面主题（明暗模式）
- 语言切换（中文/英文）
- 刷新间隔配置
- 连接配置管理
- 关于信息

**验收标准**:
- 主题切换正常
- 配置持久化
- 国际化切换正常

---

## 四、开发阶段与里程碑

### 4.1 总体规划

总开发周期: **约 8-10 周** (单人开发)

```
Phase 1          Phase 2           Phase 3           Phase 4
基础架构          核心功能           扩展功能           完善优化
(2 周)            (3 周)             (2-3 周)           (1-2 周)
```

### 4.2 Phase 1: 基础架构 (第 1-2 周)

**目标**: 建立项目基础架构，实现认证与连接

| 任务 | 工期 | 产出 |
|------|------|------|
| 项目初始化 | 1 天 | 前后端项目骨架 |
| 后端 API 代理层 | 2 天 | Go + Gin 代理框架 |
| 前端框架搭建 | 1 天 | Vue 3 项目 + 路由 + 状态管理 |
| UI 组件库集成 | 1 天 | Element Plus 集成 + 主题定制 |
| 认证模块 | 2 天 | 登录页、Token 管理 |
| API 层封装 | 1 天 | Axios 封装、拦截器 |
| 布局框架 | 1 天 | 侧边栏、顶栏、内容区 |
| 仪表盘基础版 | 1 天 | 节点状态展示 |

**里程碑 1**: 用户能登录到系统，看到仪表盘基础信息

### 4.3 Phase 2: 核心功能 (第 3-5 周)

**目标**: 实现虚拟机/容器管理核心功能

| 任务 | 工期 | 产出 |
|------|------|------|
| 虚拟机列表 | 1 天 | 列表页、筛选、分页 |
| 虚拟机生命周期 | 1.5 天 | 启动/关机/重启等操作 |
| 虚拟机详情与配置 | 2 天 | 详情页、配置编辑 |
| 创建虚拟机向导 | 2 天 | 分步表单 |
| 容器管理 | 2 天 | 复用 VM 组件，适配 CT |
| 快照管理 | 1 天 | 快照列表、创建、回滚 |
| noVNC 控制台 | 3 天 | VNC 代理、WebSocket |
| 任务状态跟踪 | 1 天 | 异步操作状态轮询 |
| 批量操作 | 1 天 | 多选、批量启停 |
| 错误处理完善 | 1 天 | 统一错误提示 |

**里程碑 2**: 能完成虚拟机的创建、管理、远程控制台完整流程

### 4.4 Phase 3: 扩展功能 (第 6-8 周)

**目标**: 实现节点、存储、网络、监控等扩展功能

| 任务 | 工期 | 产出 |
|------|------|------|
| 节点管理 | 2 天 | 节点信息、服务状态、日志 |
| 存储管理 | 3 天 | 存储列表、添加、镜像管理 |
| 网络管理 | 3 天 | 网络拓扑、接口配置 |
| 监控图表 | 2 天 | 历史数据可视化 |
| 告警配置 | 2 天 | 阈值设置、通知 |
| 多标签支持 | 1 天 | 标签页导航 |

**里程碑 3**: 具备完整的 PVE 管理能力

### 4.5 Phase 4: 完善优化 (第 9-10 周)

**目标**: 完善用户体验，性能优化，文档

| 任务 | 工期 | 产出 |
|------|------|------|
| 国际化 | 2 天 | 中英文切换 |
| 主题切换 | 1 天 | 明暗模式 |
| 性能优化 | 2 天 | 懒加载、缓存、打包优化 |
| 移动端适配 | 2 天 | 响应式布局 |
| 系统设置 | 1 天 | 配置管理页 |
| 文档编写 | 2 天 | 部署文档、使用手册 |
| Bug 修复 | 持续 | - |

**里程碑 4**: 达到可发布状态

### 4.6 版本规划

| 版本 | 内容 | 预计时间 |
|------|------|----------|
| v0.1 | 基础架构 + 认证 | 第 2 周 |
| v0.2 | 虚拟机/容器管理 | 第 5 周 |
| v0.3 | noVNC + 扩展功能 | 第 8 周 |
| v0.4 | 完善优化 | 第 10 周 |
| v1.0 | 正式发布 | 第 12 周 |

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

> **文档版本**: v1.0
> **创建日期**: 2026-04-25
> **最后更新**: 2026-04-25
> **作者**: 产品需求专家
