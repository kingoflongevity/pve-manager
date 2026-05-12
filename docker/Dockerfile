# ============================================================
# PVE Manager - Multi-Stage Dockerfile
# ============================================================
# 构建: docker build -t pve-manager:latest .
# 运行: docker run -d -p 8080:8080 -v pve-data:/app/data pve-manager:latest
# ============================================================

# ============================================================
# 阶段一: 构建前端静态资源
# ============================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --prefer-offline --no-audit --no-fund

COPY frontend/ ./
RUN npm run build

# ============================================================
# 阶段二: 构建后端 Go 二进制文件
# ============================================================
FROM golang:1.22-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/server \
    ./cmd/server

# ============================================================
# 阶段三: 最终生产镜像
# ============================================================
FROM alpine:3.20 AS production

RUN apk add --no-cache ca-certificates tzdata dumb-init

ENV TZ=Asia/Shanghai

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup -h /app -s /sbin/nologin

WORKDIR /app

COPY --from=backend-builder /app/server /app/server
COPY --from=frontend-builder /app/frontend/dist /app/static

RUN mkdir -p /app/data /app/config /app/logs && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/app/server"]
