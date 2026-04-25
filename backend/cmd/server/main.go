package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/handler"
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	// 加载配置
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		// 如果配置文件不存在，生成默认配置并提示
		if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
			sugar.Infof("配置文件不存在，生成默认配置: %s", cfgPath)
			if genErr := config.GenerateDefaultConfig(cfgPath); genErr != nil {
				sugar.Fatalf("生成默认配置失败: %v", genErr)
			}
			sugar.Infof("请编辑 %s 配置后重新启动服务", cfgPath)
			os.Exit(1)
		}
		sugar.Fatalf("加载配置失败: %v", err)
	}

	// 创建 PVE 客户端
	pveClient, err := pve.NewClient(cfg.PVE, sugar.Desugar())
	if err != nil {
		sugar.Fatalf("创建 PVE 客户端失败: %v", err)
	}

	// 后台自动登录 PVE
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		sugar.Infof("正在登录 PVE: %s", cfg.PVE.BaseURL)
		_, err := pveClient.Login(ctx, cfg.PVE.Username, cfg.PVE.Password, cfg.PVE.Realm)
		if err != nil {
			sugar.Errorf("PVE 登录失败: %v", err)
			return
		}
		sugar.Info("PVE 登录成功")
	}()

	// 初始化 Gin 路由
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// 注册中间件
	r.Use(gin.Recovery())
	r.Use(handler.RequestLogMiddleware(logger))
	r.Use(handler.CORSMiddleware(cfg.CORS))

	// 初始化处理器
	authHandler := handler.NewAuthHandler(logger)
	proxyHandler := handler.NewProxyHandler(pveClient, logger)
	vncHandler := handler.NewVNCHandler(pveClient, logger)

	// 注册路由
	setupRoutes(r, authHandler, proxyHandler, vncHandler, logger)

	// 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动服务器（优雅关闭）
	go func() {
		sugar.Infof("服务器启动在 %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalf("服务器关闭失败: %v", err)
	}

	sugar.Info("服务器已关闭")
}

// setupRoutes 配置所有 API 路由
// 将路由注册集中管理，便于维护和扩展
func setupRoutes(r *gin.Engine, authHandler *handler.AuthHandler, proxyHandler *handler.ProxyHandler, vncHandler *handler.VNCHandler, logger *zap.Logger) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ok",
			"data":    gin.H{"status": "running"},
		})
	})

	// 认证相关路由（不需要 JWT）
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login)
	}

	// PVE API 代理路由（需要 JWT 认证）
	pveGroup := r.Group("/api/pve")
	pveGroup.Use(handler.JWTAuthMiddleware(logger))
	{
		// ==================== 集群管理 ====================
		clusterGroup := pveGroup.Group("/cluster")
		{
			clusterGroup.GET("/resources", proxyHandler.GetClusterResources)
			clusterGroup.GET("/tasks", proxyHandler.GetClusterTasks)
			clusterGroup.GET("/nextid", proxyHandler.GetNextID)
			clusterGroup.GET("/ha/config", proxyHandler.GetHAConfig)
			clusterGroup.GET("/sdn/zones", proxyHandler.GetSDNZones)
			clusterGroup.GET("/sdn/vnets", proxyHandler.GetSDNVNETs)
		}

		// ==================== 资源池管理 ====================
		poolsGroup := pveGroup.Group("/pools")
		{
			poolsGroup.GET("", proxyHandler.GetPoolList)
			poolsGroup.GET("/:poolid", proxyHandler.GetPool)
		}

		// ==================== 访问控制 ====================
		accessGroup := pveGroup.Group("/access")
		{
			accessGroup.GET("/users", proxyHandler.GetUsers)
			accessGroup.GET("/groups", proxyHandler.GetGroups)
			accessGroup.GET("/roles", proxyHandler.GetRoles)
			accessGroup.GET("/acl", proxyHandler.GetACLs)
			accessGroup.GET("/domains", proxyHandler.GetDomains)
		}

		// ==================== 节点操作 ====================
		nodeGroup := pveGroup.Group("/nodes/:node")
		{
			// 节点状态
			nodeGroup.GET("/status", proxyHandler.GetNodeStatus)
			nodeGroup.GET("/version", proxyHandler.GetNodeVersion)
			nodeGroup.GET("/services", proxyHandler.GetNodeServices)
			nodeGroup.GET("/syslog", proxyHandler.GetNodeSyslog)
			nodeGroup.GET("/tasks", proxyHandler.GetNodeTasks)
			nodeGroup.GET("/tasks/:upid/status", proxyHandler.GetNodeTaskStatus)
			nodeGroup.GET("/tasks/:upid/log", proxyHandler.GetNodeTaskLog)
			nodeGroup.GET("/tasks/:upid/wait", proxyHandler.WaitForTask)
			nodeGroup.GET("/network", proxyHandler.GetNodeNetwork)
			nodeGroup.GET("/dns", proxyHandler.GetNodeDNS)
			nodeGroup.GET("/time", proxyHandler.GetNodeTime)
			nodeGroup.GET("/apt/update", proxyHandler.GetNodeAPTUpdate)
			nodeGroup.GET("/rrd", proxyHandler.GetNodeRRD)

			// ==================== 存储管理 ====================
			nodeGroup.GET("/storage", proxyHandler.GetStorageList)
			nodeGroup.GET("/storage/:storage/status", proxyHandler.GetStorageStatus)
			nodeGroup.GET("/storage/:storage/content", proxyHandler.GetStorageContent)
			nodeGroup.POST("/storage/:storage/download-url", proxyHandler.DownloadISO)

			// ==================== QEMU 虚拟机 ====================
			nodeGroup.GET("/qemu", proxyHandler.GetQEMUList)
			nodeGroup.POST("/qemu", proxyHandler.CreateQEMU)
			nodeGroup.GET("/qemu/:vmid/config", proxyHandler.GetQEMUConfig)
			nodeGroup.PUT("/qemu/:vmid/config", proxyHandler.SetQEMUConfig)
			nodeGroup.POST("/qemu/:vmid/status/:action", proxyHandler.QEMUAction)
			nodeGroup.DELETE("/qemu/:vmid", proxyHandler.DeleteQEMU)
			nodeGroup.GET("/qemu/:vmid/snapshot", proxyHandler.GetQEMUSnapshots)
			nodeGroup.POST("/qemu/:vmid/snapshot", proxyHandler.CreateQEMUSnapshot)
			nodeGroup.DELETE("/qemu/:vmid/snapshot/:snapname", proxyHandler.DeleteQEMUSnapshot)
			nodeGroup.POST("/qemu/:vmid/clone", proxyHandler.CloneQEMU)
			nodeGroup.POST("/qemu/:vmid/migrate", proxyHandler.MigrateQEMU)
			nodeGroup.GET("/qemu/:vmid/rrd", proxyHandler.GetQEMURRD)
			nodeGroup.GET("/qemu/:vmid/pending", proxyHandler.GetQEMUPending)
			nodeGroup.POST("/qemu/:vmid/vncproxy", proxyHandler.VNCProxy)

			// ==================== LXC 容器 ====================
			nodeGroup.GET("/lxc", proxyHandler.GetLXCList)
			nodeGroup.POST("/lxc", proxyHandler.CreateLXC)
			nodeGroup.GET("/lxc/:vmid/config", proxyHandler.GetLXCConfig)
			nodeGroup.PUT("/lxc/:vmid/config", proxyHandler.SetLXCConfig)
			nodeGroup.POST("/lxc/:vmid/status/:action", proxyHandler.LXCAction)
			nodeGroup.DELETE("/lxc/:vmid", proxyHandler.DeleteLXC)
			nodeGroup.GET("/lxc/:vmid/snapshot", proxyHandler.GetLXCSnapshots)
			nodeGroup.POST("/lxc/:vmid/snapshot", proxyHandler.CreateLXCSnapshot)
			nodeGroup.DELETE("/lxc/:vmid/snapshot/:snapname", proxyHandler.DeleteLXCSnapshot)
		}

		// ==================== VNC 控制台 ====================
		// VNC 票据查询（HTTP）
		pveGroup.GET("/nodes/:node/:vmType/:vmid/vnc-ticket", vncHandler.VNCProxyTicket)
	}

	// VNC WebSocket 代理路由（需要 JWT 认证）
	// 使用 /api/ws 前缀区分普通 HTTP API 和 WebSocket 端点
	wsGroup := r.Group("/api/ws")
	wsGroup.Use(handler.JWTAuthMiddleware(logger))
	{
		wsGroup.GET("/vnc/:node/:vmid/:vmType", vncHandler.ProxyVNCWebSocket)
	}
}
