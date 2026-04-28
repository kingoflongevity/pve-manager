package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/database"
	"github.com/kingoflongevity/pve-manager/backend/internal/handler"
	"github.com/kingoflongevity/pve-manager/backend/internal/repository"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

// appConfig 应用运行时配置
type appConfig struct {
	Port  int
	Debug bool
	DBPath string
}

// loadAppConfig 加载应用配置（不依赖 PVE 配置文件）
func loadAppConfig() *appConfig {
	port := 8080
	debug := false

	if p := os.Getenv("PVE_MANAGER_SERVER_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	if os.Getenv("PVE_MANAGER_SERVER_MODE") == "debug" {
		debug = true
	}

	dbPath := os.Getenv("PVE_MANAGER_DB_PATH")
	if dbPath == "" {
		dbPath = "data/pve-manager.db"
	}

	return &appConfig{
		Port:  port,
		Debug: debug,
		DBPath: dbPath,
	}
}

// setupRoutes 注册所有 API 路由
// 按功能模块分组：认证、集群、节点、存储、QEMU、LXC、访问控制
func setupRoutes(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	proxyHandler *handler.ProxyHandler,
	vncHandler *handler.VNCProxyHandler,
) {
	r.Use(handler.CORS())

	// 认证路由（无需 JWT）
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// 受保护的路由（需要 JWT）
	api := r.Group("/api/pve")
	api.Use(handler.AuthMiddleware(nil))
	{
		// 集群管理
		cluster := api.Group("/cluster")
		{
			cluster.GET("/resources", proxyHandler.GetClusterResources)
			cluster.GET("/tasks", proxyHandler.GetClusterTasks)
			cluster.GET("/nextid", proxyHandler.GetNextID)
			cluster.GET("/ha", proxyHandler.GetHAConfig)
			cluster.GET("/ha/groups", proxyHandler.GetHAGroups)
			cluster.GET("/ha/resources", proxyHandler.GetHAResources)
			cluster.GET("/sdn/zones", proxyHandler.GetSDNZones)
			cluster.GET("/sdn/vnets", proxyHandler.GetSDNVNETs)
			cluster.GET("/pools", proxyHandler.GetPoolList)
			cluster.POST("/pools", proxyHandler.CreatePool)
			cluster.GET("/pools/:poolid", proxyHandler.GetPool)
			cluster.GET("/storage", proxyHandler.GetClusterStorage)
			cluster.GET("/config", proxyHandler.GetClusterConfig)
			cluster.GET("/log", proxyHandler.GetClusterLog)
			cluster.GET("/replication", proxyHandler.GetReplicationJobs)
		}

		// 节点操作
		nodes := api.Group("/nodes/:node")
		{
			nodes.GET("/status", proxyHandler.GetNodeStatus)
			nodes.GET("/version", proxyHandler.GetNodeVersion)
			nodes.GET("/services", proxyHandler.GetNodeServices)
			nodes.GET("/syslog", proxyHandler.GetNodeSyslog)
			nodes.GET("/tasks", proxyHandler.GetNodeTasks)
			nodes.GET("/tasks/:upid/status", proxyHandler.GetNodeTaskStatus)
			nodes.GET("/tasks/:upid/log", proxyHandler.GetNodeTaskLog)
			nodes.GET("/tasks/:upid/wait", proxyHandler.WaitForTask)
			nodes.GET("/network", proxyHandler.GetNodeNetwork)
			nodes.GET("/dns", proxyHandler.GetNodeDNS)
			nodes.GET("/time", proxyHandler.GetNodeTime)
			nodes.GET("/apt/update", proxyHandler.GetNodeAPTUpdate)
			nodes.GET("/rrd", proxyHandler.GetNodeRRD)
			nodes.POST("/services/:service/:action", proxyHandler.ActionService)
		}

		// 存储管理
		storage := api.Group("/nodes/:node/storage")
		{
			storage.GET("", proxyHandler.GetStorageList)
			storage.GET("/:storage/status", proxyHandler.GetStorageStatus)
			storage.GET("/:storage/content", proxyHandler.GetStorageContent)
			storage.POST("/:storage/download", proxyHandler.DownloadISO)
			storage.GET("/:storage", proxyHandler.GetStorageDetail)
			storage.POST("", proxyHandler.CreateStorage)
			storage.POST("/:storage", proxyHandler.UpdateStorage)
			storage.DELETE("/:storage", proxyHandler.DeleteStorage)
		}

		// QEMU 虚拟机
		qemu := api.Group("/nodes/:node/qemu")
		{
			qemu.GET("", proxyHandler.GetQEMUList)
			qemu.POST("", proxyHandler.CreateQEMU)
			qemu.GET("/:vmid/config", proxyHandler.GetQEMUConfig)
			qemu.POST("/:vmid/config", proxyHandler.SetQEMUConfig)
			qemu.POST("/:vmid/status/:action", proxyHandler.QEMUAction)
			qemu.DELETE("/:vmid", proxyHandler.DeleteQEMU)
			qemu.GET("/:vmid/snapshot", proxyHandler.GetQEMUSnapshots)
			qemu.POST("/:vmid/snapshot", proxyHandler.CreateQEMUSnapshot)
			qemu.DELETE("/:vmid/snapshot/:snapname", proxyHandler.DeleteQEMUSnapshot)
			qemu.POST("/:vmid/clone", proxyHandler.CloneQEMU)
			qemu.POST("/:vmid/migrate", proxyHandler.MigrateQEMU)
			qemu.GET("/:vmid/rrd", proxyHandler.GetQEMURRD)
			qemu.GET("/:vmid/pending", proxyHandler.GetQEMUPending)
			qemu.POST("/:vmid/vncproxy", proxyHandler.VNCProxy)
			qemu.POST("/:vmid/snapshot/:snapname/rollback", proxyHandler.RollbackQEMUSnapshot)
		}

		// LXC 容器
		lxc := api.Group("/nodes/:node/lxc")
		{
			lxc.GET("", proxyHandler.GetLXCList)
			lxc.POST("", proxyHandler.CreateLXC)
			lxc.GET("/:vmid/config", proxyHandler.GetLXCConfig)
			lxc.POST("/:vmid/config", proxyHandler.SetLXCConfig)
			lxc.POST("/:vmid/status/:action", proxyHandler.LXCAction)
			lxc.DELETE("/:vmid", proxyHandler.DeleteLXC)
			lxc.GET("/:vmid/snapshot", proxyHandler.GetLXCSnapshots)
			lxc.POST("/:vmid/snapshot", proxyHandler.CreateLXCSnapshot)
			lxc.DELETE("/:vmid/snapshot/:snapname", proxyHandler.DeleteLXCSnapshot)
			lxc.POST("/:vmid/snapshot/:snapname/rollback", proxyHandler.RollbackLXCSnapshot)
			lxc.POST("/:vmid/clone", proxyHandler.CloneLXC)
			lxc.POST("/:vmid/migrate", proxyHandler.MigrateLXC)
			lxc.GET("/:vmid/rrd", proxyHandler.GetLXCRRD)
			lxc.GET("/:vmid/pending", proxyHandler.GetLXCPending)
			lxc.POST("/:vmid/vncproxy", proxyHandler.LXCVNCProxy)
		}

		// 访问控制
		access := api.Group("/access")
		{
			access.GET("/users", proxyHandler.GetUsers)
			access.POST("/users", proxyHandler.CreateUser)
			access.GET("/users/:userid", proxyHandler.GetUser)
			access.POST("/users/:userid", proxyHandler.UpdateUser)
			access.DELETE("/users/:userid", proxyHandler.DeleteUser)
			access.POST("/users/:userid/password", proxyHandler.UpdateUserPassword)
			access.GET("/groups", proxyHandler.GetGroups)
			access.POST("/groups", proxyHandler.CreateGroup)
			access.GET("/groups/:groupid", proxyHandler.GetGroup)
			access.POST("/groups/:groupid", proxyHandler.UpdateGroup)
			access.DELETE("/groups/:groupid", proxyHandler.DeleteGroup)
			access.GET("/roles", proxyHandler.GetRoles)
			access.POST("/roles", proxyHandler.CreateRole)
			access.POST("/roles/:roleid", proxyHandler.UpdateRole)
			access.DELETE("/roles/:roleid", proxyHandler.DeleteRole)
			access.GET("/acl", proxyHandler.GetACLs)
			access.POST("/acl", proxyHandler.SetACL)
			access.GET("/domains", proxyHandler.GetDomains)
			access.GET("/domains/:realm", proxyHandler.GetDomain)
		}

		// WebSocket VNC
		r.GET("/api/pve/vnc/websocket", vncHandler.HandleVNC)

		// 管理接口
		admin := api.Group("/admin")
		{
			admin.GET("/pve-configs", authHandler.GetPVEConfigs)
			admin.GET("/audit-logs", authHandler.GetAuditLogs)
		}
	}
}

func main() {
	// 初始化 Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 加载应用配置
	appCfg := loadAppConfig()

	// 初始化数据库
	_, err := database.InitDatabase(logger)
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}

	// 初始化 Repositories（使用全局 DB 实例）
	configRepo := repository.NewPVEConfigRepo()
	sessionRepo := repository.NewSessionRepo()
	auditRepo := repository.NewAuditLogRepo()

	// 初始化 Services
	authService := service.NewAuthService(logger, configRepo, sessionRepo, auditRepo)
	clusterService := service.NewClusterService(logger)
	vmService := service.NewVMService(logger)
	containerService := service.NewContainerService(logger)
	storageService := service.NewStorageService(logger)
	nodeService := service.NewNodeService(logger)

	// 初始化 Handlers
	authHandler := handler.NewAuthHandler(logger, authService)
	proxyHandler := handler.NewProxyHandler(
		logger, authService, clusterService, vmService, containerService, storageService, nodeService,
	)
	vncHandler := handler.NewVNCProxyHandler(authService)

	// 设置 Gin 模式
	if appCfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由引擎
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 注册路由
	setupRoutes(r, authHandler, proxyHandler, vncHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%d", appCfg.Port)
	logger.Info("服务器启动成功", zap.String("addr", addr))

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
