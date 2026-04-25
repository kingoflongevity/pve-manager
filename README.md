# PVE Manager

<p align="center">
  <b>面向中国用户的 Proxmox VE 现代化 Web 管理面板</b>
</p>

<p align="center">
  <a href="https://vuejs.org/" target="_blank">
    <img src="https://img.shields.io/badge/Frontend-Vue3-42b883.svg?style=flat-square&logo=vue.js" alt="Vue3">
  </a>
  <a href="https://go.dev/" target="_blank">
    <img src="https://img.shields.io/badge/Backend-Go-00ADD8.svg?style=flat-square&logo=go" alt="Go">
  </a>
  <a href="https://www.proxmox.com/en/proxmox-virtual-environment/overview" target="_blank">
    <img src="https://img.shields.io/badge/PVE-API-E57000.svg?style=flat-square&logo=proxmox" alt="PVE">
  </a>
  <a href="https://github.com/kingoflongevity/pve-manager/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" alt="License">
  </a>
</p>

---

## 项目简介

Proxmox VE (PVE) 原生的 Web UI 基于 ExtJS，界面老旧且对中文用户不够友好。本项目旨在构建一个现代化、轻量化、面向中国用户的 PVE Web 管理面板。

### 核心价值

- 现代化 UI/UX，提升使用体验
- 完善的中文本地化
- 轻量级部署，低资源占用
- 常用功能快速访问（80/20 原则）

## 功能特性

### 已实现

- 用户认证（用户名/密码 + API Token）
- 仪表盘概览（CPU/内存/存储/网络）
- 虚拟机/容器管理（启停、创建、配置编辑）
- 快照管理
- noVNC 远程控制台
- 任务状态跟踪

### 规划中

- 存储管理（NFS、LVM、ZFS）
- 网络管理（VLAN、Bond）
- 监控与告警（历史图表、阈值告警）
- 用户与权限管理（RBAC）
- 多节点统一管理

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + TypeScript + Element Plus + Vite + Pinia |
| 后端 | Go + Gin + Viper + Zap |
| 远程控制台 | noVNC + WebSocket 代理 |
| 图表 | ECharts |
| 缓存 | Redis（可选） |
| 部署 | Docker + Nginx |

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Git

### 开发环境启动

**后端：**

```bash
cd backend
go run cmd/server/main.go
```

后端服务默认运行在 `http://localhost:8080`

**前端：**

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:3000`

### Docker 部署

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

访问 `http://localhost:3000` 即可。

### 配置说明

首次运行后端会自动生成 `config.yaml` 配置文件：

```yaml
server:
  port: 8080
  mode: debug

pve:
  host: 192.168.1.100
  port: 8006
  verify_ssl: false
```

## 项目结构

```
pve_webui/
├── backend/          # Go 后端
│   ├── cmd/server/   # 程序入口
│   ├── internal/     # 业务逻辑
│   └── pkg/          # 可复用工具包
├── frontend/         # Vue 3 前端
│   ├── src/
│   │   ├── api/      # API 层
│   │   ├── views/    # 页面组件
│   │   ├── stores/   # Pinia 状态管理
│   │   └── components/ # 通用组件
│   └── ...
├── docker/           # Docker 部署配置
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── nginx.conf
└── docs/             # 文档
    └── PLAN.md       # 开发计划
```

## 开发计划

详细开发计划请参阅 [PLAN.md](PLAN.md)

| 版本 | 内容 | 状态 |
|------|------|------|
| v0.1 | 基础架构 + 认证 | ✅ 完成 |
| v0.2 | 虚拟机/容器管理 | 🚧 开发中 |
| v0.3 | noVNC + 扩展功能 | ⏳ 待开始 |
| v1.0 | 正式发布 | ⏳ 待开始 |

## 许可证

[MIT](LICENSE)
