package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
	"go.uber.org/zap"
)

// ProxyHandler API 代理处理器
// 将前端请求代理到 PVE API，处理路径映射和认证传递
type ProxyHandler struct {
	logger *zap.Logger
}

// NewProxyHandler 创建代理处理器实例
func NewProxyHandler(logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		logger: logger,
	}
}

// getPVEClient 从请求上下文中提取 JWT token，构建 PVE 客户端
// 每个请求使用独立的 PVE 客户端，避免共享状态问题
func (h *ProxyHandler) getPVEClient(c *gin.Context) (*pve.Client, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("未提供认证令牌")
	}
	tokenString := authHeader[7:]
	client, err := BuildPVEClient(tokenString, h.logger)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ============================================================
// 通用代理方法
// ============================================================

// GetNodes 获取节点列表
func (h *ProxyHandler) GetNodes(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.logger.Error("获取 PVE 客户端失败", zap.Error(err))
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	nodes, err := client.GetNodes(c.Request.Context())
	if err != nil {
		h.logger.Error("获取节点列表失败", zap.Error(err))
		h.serverError(c, "获取节点列表失败: "+err.Error())
		return
	}

	h.success(c, nodes)
}

// Proxy 通用代理请求到 PVE API
func (h *ProxyHandler) Proxy(c *gin.Context) {
	proxyPath := c.Param("proxyPath")
	if proxyPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "代理路径不能为空",
		})
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.logger.Error("获取 PVE 客户端失败", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "获取 PVE 客户端失败: " + err.Error(),
		})
		return
	}

	var reqBody io.Reader
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
		reqBody = c.Request.Body
	}

	respBody, statusCode, err := client.ProxyRequest(
		c.Request.Context(),
		c.Request.Method,
		proxyPath,
		reqBody,
	)
	if err != nil {
		h.logger.Error("代理请求失败",
			zap.String("path", proxyPath),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": "代理请求失败: " + err.Error(),
		})
		return
	}

	c.Data(statusCode, "application/json", respBody)
}

// ============================================================
// QEMU 虚拟机处理
// ============================================================

// GetQEMUList 获取节点上所有 QEMU 虚拟机列表
// GET /api/pve/nodes/:node/qemu
func (h *ProxyHandler) GetQEMUList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	vms, err := client.ListQEMU(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 QEMU 列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取虚拟机列表失败: "+err.Error())
		return
	}

	h.success(c, vms)
}

// GetQEMUConfig 获取虚拟机完整配置
// GET /api/pve/nodes/:node/qemu/:vmid/config
func (h *ProxyHandler) GetQEMUConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	config, err := client.GetQEMUConfig(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 QEMU 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取虚拟机配置失败: "+err.Error())
		return
	}

	h.success(c, config)
}

// CreateQEMU 创建虚拟机
// POST /api/pve/nodes/:node/qemu
func (h *ProxyHandler) CreateQEMU(c *gin.Context) {
	node := c.Param("node")
	var params pve.QEMUCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.CreateQEMU(c.Request.Context(), node, &params)
	if err != nil {
		h.logger.Error("创建 QEMU 失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "创建虚拟机失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机创建任务已提交"})
}

// QEMUAction 执行虚拟机操作
// POST /api/pve/nodes/:node/qemu/:vmid/status/:action
func (h *ProxyHandler) QEMUAction(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	action := c.Param("action")

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	var upid string
	ctx := c.Request.Context()

	switch action {
	case "start":
		upid, err = client.StartQEMU(ctx, node, vmid)
	case "stop":
		upid, err = client.StopQEMU(ctx, node, vmid)
	case "shutdown":
		upid, err = client.ShutdownQEMU(ctx, node, vmid)
	case "reboot":
		upid, err = client.RebootQEMU(ctx, node, vmid)
	case "suspend":
		upid, err = client.SuspendQEMU(ctx, node, vmid)
	case "resume":
		upid, err = client.ResumeQEMU(ctx, node, vmid)
	case "reset":
		upid, err = client.ResetQEMU(ctx, node, vmid)
	default:
		h.badRequest(c, "不支持的操作: "+action)
		return
	}

	if err != nil {
		h.logger.Error("QEMU 操作失败",
			zap.String("node", node), zap.Int("vmid", vmid),
			zap.String("action", action), zap.Error(err))
		h.serverError(c, fmt.Sprintf("虚拟机 %s 操作失败: %s", action, err.Error()))
		return
	}

	h.success(c, gin.H{"upid": upid, "message": fmt.Sprintf("虚拟机 %s 任务已提交", action)})
}

// SetQEMUConfig 更新虚拟机配置
// PUT /api/pve/nodes/:node/qemu/:vmid/config
func (h *ProxyHandler) SetQEMUConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	var config pve.QEMUConfigParams
	if err := c.ShouldBindJSON(&config); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.SetQEMUConfig(c.Request.Context(), node, vmid, config)
	if err != nil {
		h.logger.Error("更新 QEMU 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "更新虚拟机配置失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机配置更新任务已提交"})
}

// GetQEMUSnapshots 获取虚拟机快照列表
// GET /api/pve/nodes/:node/qemu/:vmid/snapshot
func (h *ProxyHandler) GetQEMUSnapshots(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	snapshots, err := client.ListQEMUSnapshots(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 QEMU 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取虚拟机快照失败: "+err.Error())
		return
	}

	h.success(c, snapshots)
}

// CreateQEMUSnapshot 创建虚拟机快照
// POST /api/pve/nodes/:node/qemu/:vmid/snapshot
func (h *ProxyHandler) CreateQEMUSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.CreateQEMUSnapshot(c.Request.Context(), node, vmid, req.Name, req.Description)
	if err != nil {
		h.logger.Error("创建 QEMU 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "创建虚拟机快照失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机快照创建任务已提交"})
}

// DeleteQEMUSnapshot 删除虚拟机快照
// DELETE /api/pve/nodes/:node/qemu/:vmid/snapshot/:snapname
func (h *ProxyHandler) DeleteQEMUSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	snapname := c.Param("snapname")

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.DeleteQEMUSnapshot(c.Request.Context(), node, vmid, snapname)
	if err != nil {
		h.logger.Error("删除 QEMU 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid),
			zap.String("snapname", snapname), zap.Error(err))
		h.serverError(c, "删除虚拟机快照失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机快照删除任务已提交"})
}

// CloneQEMU 克隆虚拟机
// POST /api/pve/nodes/:node/qemu/:vmid/clone
func (h *ProxyHandler) CloneQEMU(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	var params pve.QEMUCloneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.CloneQEMU(c.Request.Context(), node, vmid, &params)
	if err != nil {
		h.logger.Error("克隆 QEMU 失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "克隆虚拟机失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机克隆任务已提交"})
}

// MigrateQEMU 迁移虚拟机
// POST /api/pve/nodes/:node/qemu/:vmid/migrate
func (h *ProxyHandler) MigrateQEMU(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	var params pve.QEMUMigrateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.MigrateQEMU(c.Request.Context(), node, vmid, &params)
	if err != nil {
		h.logger.Error("迁移 QEMU 失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "迁移虚拟机失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机迁移任务已提交"})
}

// DeleteQEMU 删除虚拟机
// DELETE /api/pve/nodes/:node/qemu/:vmid
func (h *ProxyHandler) DeleteQEMU(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.DeleteQEMU(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("删除 QEMU 失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "删除虚拟机失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "虚拟机删除任务已提交"})
}

// GetQEMURRD 获取虚拟机 RRD 性能数据
// GET /api/pve/nodes/:node/qemu/:vmid/rrd
func (h *ProxyHandler) GetQEMURRD(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.Query("dataset")

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	data, err := client.GetQEMURRD(c.Request.Context(), node, vmid, timeframe, dataset)
	if err != nil {
		h.logger.Error("获取 QEMU RRD 数据失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取性能数据失败: "+err.Error())
		return
	}

	h.success(c, data)
}

// GetQEMUPending 获取虚拟机待处理配置
// GET /api/pve/nodes/:node/qemu/:vmid/pending
func (h *ProxyHandler) GetQEMUPending(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	pending, err := client.GetQEMUPending(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 QEMU 待处理配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取待处理配置失败: "+err.Error())
		return
	}

	h.success(c, pending)
}

// ============================================================
// LXC 容器处理
// ============================================================

// GetLXCList 获取节点上所有 LXC 容器列表
func (h *ProxyHandler) GetLXCList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	containers, err := client.ListLXC(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 LXC 列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取容器列表失败: "+err.Error())
		return
	}

	h.success(c, containers)
}

// GetLXCConfig 获取容器完整配置
func (h *ProxyHandler) GetLXCConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	config, err := client.GetLXCConfig(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 LXC 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取容器配置失败: "+err.Error())
		return
	}

	h.success(c, config)
}

// CreateLXC 创建容器
func (h *ProxyHandler) CreateLXC(c *gin.Context) {
	node := c.Param("node")
	var params pve.LXCCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.CreateLXC(c.Request.Context(), node, &params)
	if err != nil {
		h.logger.Error("创建 LXC 失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "创建容器失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器创建任务已提交"})
}

// LXCAction 执行容器操作
func (h *ProxyHandler) LXCAction(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	action := c.Param("action")

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	var upid string
	ctx := c.Request.Context()

	switch action {
	case "start":
		upid, err = client.StartLXC(ctx, node, vmid)
	case "stop":
		upid, err = client.StopLXC(ctx, node, vmid)
	case "shutdown":
		upid, err = client.ShutdownLXC(ctx, node, vmid)
	case "reboot":
		upid, err = client.RebootLXC(ctx, node, vmid)
	case "freeze":
		upid, err = client.FreezeLXC(ctx, node, vmid)
	case "unfreeze":
		upid, err = client.UnfreezeLXC(ctx, node, vmid)
	default:
		h.badRequest(c, "不支持的操作: "+action)
		return
	}

	if err != nil {
		h.logger.Error("LXC 操作失败",
			zap.String("node", node), zap.Int("vmid", vmid),
			zap.String("action", action), zap.Error(err))
		h.serverError(c, fmt.Sprintf("容器 %s 操作失败: %s", action, err.Error()))
		return
	}

	h.success(c, gin.H{"upid": upid, "message": fmt.Sprintf("容器 %s 任务已提交", action)})
}

// SetLXCConfig 更新容器配置
func (h *ProxyHandler) SetLXCConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	var config pve.LXCConfigParams
	if err := c.ShouldBindJSON(&config); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.SetLXCConfig(c.Request.Context(), node, vmid, config)
	if err != nil {
		h.logger.Error("更新 LXC 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "更新容器配置失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器配置更新任务已提交"})
}

// DeleteLXC 删除容器
func (h *ProxyHandler) DeleteLXC(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.DeleteLXC(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("删除 LXC 失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "删除容器失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器删除任务已提交"})
}

// GetLXCSnapshots 获取容器快照列表
func (h *ProxyHandler) GetLXCSnapshots(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	snapshots, err := client.ListLXCSnapshots(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 LXC 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取容器快照失败: "+err.Error())
		return
	}

	h.success(c, snapshots)
}

// CreateLXCSnapshot 创建容器快照
func (h *ProxyHandler) CreateLXCSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.CreateLXCSnapshot(c.Request.Context(), node, vmid, req.Name, req.Description)
	if err != nil {
		h.logger.Error("创建 LXC 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "创建容器快照失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器快照创建任务已提交"})
}

// DeleteLXCSnapshot 删除容器快照
func (h *ProxyHandler) DeleteLXCSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	snapname := c.Param("snapname")

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	upid, err := client.DeleteLXCSnapshot(c.Request.Context(), node, vmid, snapname)
	if err != nil {
		h.logger.Error("删除 LXC 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid),
			zap.String("snapname", snapname), zap.Error(err))
		h.serverError(c, "删除容器快照失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器快照删除任务已提交"})
}

// ============================================================
// 节点管理处理
// ============================================================

// GetNodeStatus 获取节点状态
func (h *ProxyHandler) GetNodeStatus(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	status, err := client.GetNodeStatus(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点状态失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点状态失败: "+err.Error())
		return
	}
	h.success(c, status)
}

// GetNodeVersion 获取节点版本
func (h *ProxyHandler) GetNodeVersion(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	version, err := client.GetNodeVersion(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点版本失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点版本失败: "+err.Error())
		return
	}
	h.success(c, version)
}

// GetNodeServices 获取节点服务列表
func (h *ProxyHandler) GetNodeServices(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	services, err := client.GetNodeServices(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点服务列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点服务列表失败: "+err.Error())
		return
	}
	h.success(c, services)
}

// GetNodeSyslog 获取节点系统日志
func (h *ProxyHandler) GetNodeSyslog(c *gin.Context) {
	node := c.Param("node")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	logs, err := client.GetNodeSyslog(c.Request.Context(), node, limit, start)
	if err != nil {
		h.logger.Error("获取系统日志失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取系统日志失败: "+err.Error())
		return
	}
	h.success(c, logs)
}

// GetNodeTasks 获取节点任务列表
func (h *ProxyHandler) GetNodeTasks(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	tasks, err := client.GetNodeTasks(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点任务列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取任务列表失败: "+err.Error())
		return
	}
	h.success(c, tasks)
}

// GetNodeTaskStatus 获取任务状态
func (h *ProxyHandler) GetNodeTaskStatus(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	status, err := client.GetNodeTaskStatus(c.Request.Context(), node, upid)
	if err != nil {
		h.logger.Error("获取任务状态失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "获取任务状态失败: "+err.Error())
		return
	}
	h.success(c, status)
}

// GetNodeTaskLog 获取任务日志
func (h *ProxyHandler) GetNodeTaskLog(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	logs, err := client.GetNodeTaskLog(c.Request.Context(), node, upid)
	if err != nil {
		h.logger.Error("获取任务日志失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "获取任务日志失败: "+err.Error())
		return
	}
	h.success(c, logs)
}

// WaitForTask 等待任务完成（轮询）
func (h *ProxyHandler) WaitForTask(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	timeoutSec, _ := strconv.Atoi(c.DefaultQuery("timeout", "60"))
	timeout := time.Duration(timeoutSec) * time.Second
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	status, err := client.WaitForTask(c.Request.Context(), node, upid, timeout)
	if err != nil {
		h.logger.Error("等待任务完成失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "等待任务超时: "+err.Error())
		return
	}
	h.success(c, status)
}

// GetNodeNetwork 获取网络接口列表
func (h *ProxyHandler) GetNodeNetwork(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	interfaces, err := client.GetNodeNetwork(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取网络接口列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取网络接口列表失败: "+err.Error())
		return
	}
	h.success(c, interfaces)
}

// GetNodeDNS 获取 DNS 配置
func (h *ProxyHandler) GetNodeDNS(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	dns, err := client.GetNodeDNS(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 DNS 配置失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取 DNS 配置失败: "+err.Error())
		return
	}
	h.success(c, dns)
}

// GetNodeTime 获取时间信息
func (h *ProxyHandler) GetNodeTime(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	timeInfo, err := client.GetNodeTime(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取时间信息失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取时间信息失败: "+err.Error())
		return
	}
	h.success(c, timeInfo)
}

// GetNodeAPTUpdate 获取可更新软件包
func (h *ProxyHandler) GetNodeAPTUpdate(c *gin.Context) {
	node := c.Param("node")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	updates, err := client.GetNodeAPTUpdate(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取软件包更新列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取软件包更新列表失败: "+err.Error())
		return
	}
	h.success(c, updates)
}

// GetNodeRRD 获取节点 RRD 性能数据
func (h *ProxyHandler) GetNodeRRD(c *gin.Context) {
	node := c.Param("node")
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.Query("dataset")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.GetNodeRRD(c.Request.Context(), node, timeframe, dataset)
	if err != nil {
		h.logger.Error("获取节点 RRD 数据失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取性能数据失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ============================================================
// 存储管理处理
// ============================================================

// GetStorageList 获取节点存储列表
func (h *ProxyHandler) GetStorageList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	storages, err := client.ListStorage(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取存储列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取存储列表失败: "+err.Error())
		return
	}
	h.success(c, storages)
}

// GetStorageStatus 获取存储状态
func (h *ProxyHandler) GetStorageStatus(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	status, err := client.GetStorageStatus(c.Request.Context(), node, storage)
	if err != nil {
		h.logger.Error("获取存储状态失败",
			zap.String("node", node), zap.String("storage", storage), zap.Error(err))
		h.serverError(c, "获取存储状态失败: "+err.Error())
		return
	}
	h.success(c, status)
}

// GetStorageContent 获取存储内容
func (h *ProxyHandler) GetStorageContent(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	contentType := c.Query("content")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var content []pve.StorageContent
	if contentType != "" {
		content, err = client.GetStorageContentByType(c.Request.Context(), node, storage, contentType)
	} else {
		content, err = client.GetStorageContent(c.Request.Context(), node, storage)
	}
	if err != nil {
		h.logger.Error("获取存储内容失败",
			zap.String("node", node), zap.String("storage", storage), zap.Error(err))
		h.serverError(c, "获取存储内容失败: "+err.Error())
		return
	}
	h.success(c, content)
}

// DownloadISO 从 URL 下载 ISO
func (h *ProxyHandler) DownloadISO(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	var req struct {
		URL      string `json:"url" binding:"required"`
		Filename string `json:"filename" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	upid, err := client.DownloadISO(c.Request.Context(), node, storage, req.URL, req.Filename)
	if err != nil {
		h.logger.Error("下载 ISO 失败",
			zap.String("node", node), zap.String("storage", storage), zap.Error(err))
		h.serverError(c, "下载 ISO 失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": upid, "message": "ISO 下载任务已提交"})
}

// ============================================================
// 集群管理处理
// ============================================================

// GetClusterResources 获取集群资源列表
func (h *ProxyHandler) GetClusterResources(c *gin.Context) {
	resourceType := c.Query("type")
	var resources []pve.ClusterResource
	var err error
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	if resourceType != "" {
		resources, err = client.GetClusterResourcesByType(c.Request.Context(), resourceType)
	} else {
		resources, err = client.GetClusterResources(c.Request.Context())
	}
	if err != nil {
		h.logger.Error("获取集群资源失败", zap.Error(err))
		h.serverError(c, "获取集群资源失败: "+err.Error())
		return
	}
	h.success(c, resources)
}

// GetClusterTasks 获取集群任务
func (h *ProxyHandler) GetClusterTasks(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	tasks, err := client.GetClusterTasks(c.Request.Context())
	if err != nil {
		h.logger.Error("获取集群任务失败", zap.Error(err))
		h.serverError(c, "获取集群任务失败: "+err.Error())
		return
	}
	h.success(c, tasks)
}

// GetNextID 获取下一个 VM ID
func (h *ProxyHandler) GetNextID(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	nextID, err := client.GetNextID(c.Request.Context())
	if err != nil {
		h.logger.Error("获取下一个 VM ID 失败", zap.Error(err))
		h.serverError(c, "获取下一个 VM ID 失败: "+err.Error())
		return
	}
	h.success(c, nextID)
}

// GetHAConfig 获取 HA 配置
func (h *ProxyHandler) GetHAConfig(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	config, err := client.GetHAConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 HA 配置失败", zap.Error(err))
		h.serverError(c, "获取 HA 配置失败: "+err.Error())
		return
	}
	h.success(c, config)
}

// GetSDNZones 获取 SDN 区域
func (h *ProxyHandler) GetSDNZones(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	zones, err := client.GetSDNZones(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 SDN 区域失败", zap.Error(err))
		h.serverError(c, "获取 SDN 区域失败: "+err.Error())
		return
	}
	h.success(c, zones)
}

// GetSDNVNETs 获取 SDN 虚拟网络
func (h *ProxyHandler) GetSDNVNETs(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	vnets, err := client.GetSDNVNETs(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 SDN 虚拟网络失败", zap.Error(err))
		h.serverError(c, "获取 SDN 虚拟网络失败: "+err.Error())
		return
	}
	h.success(c, vnets)
}

// GetPoolList 获取资源池列表
func (h *ProxyHandler) GetPoolList(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	pools, err := client.ListPools(c.Request.Context())
	if err != nil {
		h.logger.Error("获取资源池列表失败", zap.Error(err))
		h.serverError(c, "获取资源池列表失败: "+err.Error())
		return
	}
	h.success(c, pools)
}

// GetPool 获取资源池详情
func (h *ProxyHandler) GetPool(c *gin.Context) {
	poolid := c.Param("poolid")
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	detail, err := client.GetPool(c.Request.Context(), poolid)
	if err != nil {
		h.logger.Error("获取资源池详情失败", zap.String("poolid", poolid), zap.Error(err))
		h.serverError(c, "获取资源池详情失败: "+err.Error())
		return
	}
	h.success(c, detail)
}

// ============================================================
// 访问控制处理
// ============================================================

// GetUsers 获取用户列表
func (h *ProxyHandler) GetUsers(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	users, err := client.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("获取用户列表失败", zap.Error(err))
		h.serverError(c, "获取用户列表失败: "+err.Error())
		return
	}
	h.success(c, users)
}

// GetGroups 获取组列表
func (h *ProxyHandler) GetGroups(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	groups, err := client.ListGroups(c.Request.Context())
	if err != nil {
		h.logger.Error("获取组列表失败", zap.Error(err))
		h.serverError(c, "获取组列表失败: "+err.Error())
		return
	}
	h.success(c, groups)
}

// GetRoles 获取角色列表
func (h *ProxyHandler) GetRoles(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	roles, err := client.ListRoles(c.Request.Context())
	if err != nil {
		h.logger.Error("获取角色列表失败", zap.Error(err))
		h.serverError(c, "获取角色列表失败: "+err.Error())
		return
	}
	h.success(c, roles)
}

// GetACLs 获取 ACL 列表
func (h *ProxyHandler) GetACLs(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	acls, err := client.ListACLs(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 ACL 列表失败", zap.Error(err))
		h.serverError(c, "获取 ACL 列表失败: "+err.Error())
		return
	}
	h.success(c, acls)
}

// GetDomains 获取认证域列表
func (h *ProxyHandler) GetDomains(c *gin.Context) {
	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	domains, err := client.ListDomains(c.Request.Context())
	if err != nil {
		h.logger.Error("获取认证域列表失败", zap.Error(err))
		h.serverError(c, "获取认证域列表失败: "+err.Error())
		return
	}
	h.success(c, domains)
}

// ============================================================
// 辅助方法
// ============================================================

// success 返回成功响应
func (h *ProxyHandler) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// badRequest 返回请求错误
func (h *ProxyHandler) badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    400,
		"message": message,
	})
}

// serverError 返回服务器错误
func (h *ProxyHandler) serverError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    500,
		"message": message,
	})
}

// VNCProxy 处理 VNC 代理 WebSocket 连接
func (h *ProxyHandler) VNCProxy(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	client, err := h.getPVEClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}

	path := fmt.Sprintf("nodes/%s/qemu/%d/vncproxy", node, vmid)
	params := map[string]interface{}{
		"websocket": 1,
	}

	resp, err := client.Do(context.Background(), "POST", path, params)
	if err != nil {
		h.logger.Error("获取 VNC 代理失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取 VNC 代理失败: "+err.Error())
		return
	}
	h.success(c, resp.Data)
}
