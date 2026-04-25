# ============================================================
# PVE Web Management Panel - Multi-Stage Dockerfile
# ============================================================
# 构建说明:
#   构建完整镜像: docker build -t pve-webui:latest .
#   指定构建目标: docker build --target builder -t pve-webui:builder .
#   使用 buildkit: DOCKER_BUILDKIT=1 docker build -t pve-webui:latest .
# ============================================================

# ============================================================
# 阶段一: 构建前端静态资源
# ============================================================
FROM node:18-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app/frontend

# 复制前端依赖文件并安装（利用 Docker 层缓存）
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --prefer-offline --no-audit --no-fund

# 复制前端源代码并执行构建
COPY frontend/ ./
RUN npm run build

# ============================================================
# 阶段二: 构建后端 Go 二进制文件
# ============================================================
FROM golang:1.21-alpine AS backend-builder

# 安装构建依赖（CGO、git 等）
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /app/backend

# 复制 Go 模块文件并下载依赖（利用 Docker 层缓存）
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源代码并编译
# -ldflags 参数: 去除调试信息、压缩体积
# -trimpath 参数: 移除编译路径信息，增强安全性
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=0.1.0" \
    -trimpath \
    -o /app/server \
    ./cmd/server

# ============================================================
# 阶段三: 最终生产镜像
# ============================================================
FROM alpine:3.18 AS production

# 安装运行时依赖
# ca-certificates: HTTPS 证书校验
# tzdata: 时区支持
# nginx: 反向代理（可选，如需反向代理可启用）
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    dumb-init

# 设置时区为上海（可根据需要修改）
ENV TZ=Asia/Shanghai

# 创建非 root 用户运行应用（安全最佳实践）
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup -h /app -s /sbin/nologin

# 设置工作目录
WORKDIR /app

# 从阶段二复制编译好的后端二进制文件
COPY --from=backend-builder /app/server /app/server

# 从阶段一复制构建好的前端静态文件
COPY --from=frontend-builder /app/frontend/dist /app/static

# 复制入口脚本（如果存在）
# COPY docker/entrypoint.sh /app/entrypoint.sh
# RUN chmod +x /app/entrypoint.sh

# 创建必要的目录并设置权限
RUN mkdir -p /app/logs /app/config && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露应用端口（从 main.go 中可以看到默认端口为 3000）
EXPOSE 3000

# 健康检查（使用应用自带的 /health 端点）
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# 启动命令（使用 dumb-init 作为 PID 1，正确处理信号）
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/app/server"]
