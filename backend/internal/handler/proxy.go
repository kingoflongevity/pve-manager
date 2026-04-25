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
// 提供类型化的 API 端点，进行请求/响应转换和错误处理
type ProxyHandler struct {
	client *pve.Client
	logger *zap.Logger
}

// NewProxyHandler 创建代理处理器实例
// 接收 PVE 客户端和 logger 用于代理请求和日志记录
func NewProxyHandler(client *pve.Client, logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		client: client,
		logger: logger,
	}
}

// ============================================================
// 通用代理方法
// ============================================================

// GetNodes 获取节点列表
// GET /api/pve/nodes
// 返回集群中所有节点的信息
func (h *ProxyHandler) GetNodes(c *gin.Context) {
	nodes, err := h.client.GetNodes(c.Request.Context())
	if err != nil {
		h.logger.Error("获取节点列表失败", zap.Error(err))
		h.serverError(c, "获取节点列表失败: "+err.Error())
		return
	}

	h.success(c, nodes)
}

// Proxy 通用代理请求到 PVE API
// POST/GET /api/pve/*
// 自动将请求路径和方法转发到 PVE，返回原始响应
func (h *ProxyHandler) Proxy(c *gin.Context) {
	proxyPath := c.Param("proxyPath")
	if proxyPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "代理路径不能为空",
		})
		return
	}

	if !h.client.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "PVE 客户端未认证",
		})
		return
	}

	var reqBody io.Reader
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
		reqBody = c.Request.Body
	}

	respBody, statusCode, err := h.client.ProxyRequest(
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

	vms, err := h.client.ListQEMU(c.Request.Context(), node)
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

	config, err := h.client.GetQEMUConfig(c.Request.Context(), node, vmid)
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

	upid, err := h.client.CreateQEMU(c.Request.Context(), node, &params)
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

	var upid string
	ctx := c.Request.Context()

	switch action {
	case "start":
		upid, err = h.client.StartQEMU(ctx, node, vmid)
	case "stop":
		upid, err = h.client.StopQEMU(ctx, node, vmid)
	case "shutdown":
		upid, err = h.client.ShutdownQEMU(ctx, node, vmid)
	case "reboot":
		upid, err = h.client.RebootQEMU(ctx, node, vmid)
	case "suspend":
		upid, err = h.client.SuspendQEMU(ctx, node, vmid)
	case "resume":
		upid, err = h.client.ResumeQEMU(ctx, node, vmid)
	case "reset":
		upid, err = h.client.ResetQEMU(ctx, node, vmid)
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

	upid, err := h.client.SetQEMUConfig(c.Request.Context(), node, vmid, config)
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

	snapshots, err := h.client.ListQEMUSnapshots(c.Request.Context(), node, vmid)
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

	upid, err := h.client.CreateQEMUSnapshot(c.Request.Context(), node, vmid, req.Name, req.Description)
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

	upid, err := h.client.DeleteQEMUSnapshot(c.Request.Context(), node, vmid, snapname)
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

	upid, err := h.client.CloneQEMU(c.Request.Context(), node, vmid, &params)
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

	upid, err := h.client.MigrateQEMU(c.Request.Context(), node, vmid, &params)
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

	upid, err := h.client.DeleteQEMU(c.Request.Context(), node, vmid)
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

	data, err := h.client.GetQEMURRD(c.Request.Context(), node, vmid, timeframe, dataset)
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

	pending, err := h.client.GetQEMUPending(c.Request.Context(), node, vmid)
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
// GET /api/pve/nodes/:node/lxc
func (h *ProxyHandler) GetLXCList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}

	containers, err := h.client.ListLXC(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 LXC 列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取容器列表失败: "+err.Error())
		return
	}

	h.success(c, containers)
}

// GetLXCConfig 获取容器完整配置
// GET /api/pve/nodes/:node/lxc/:vmid/config
func (h *ProxyHandler) GetLXCConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	config, err := h.client.GetLXCConfig(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 LXC 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取容器配置失败: "+err.Error())
		return
	}

	h.success(c, config)
}

// CreateLXC 创建容器
// POST /api/pve/nodes/:node/lxc
func (h *ProxyHandler) CreateLXC(c *gin.Context) {
	node := c.Param("node")
	var params pve.LXCCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	upid, err := h.client.CreateLXC(c.Request.Context(), node, &params)
	if err != nil {
		h.logger.Error("创建 LXC 失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "创建容器失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器创建任务已提交"})
}

// LXCAction 执行容器操作
// POST /api/pve/nodes/:node/lxc/:vmid/status/:action
func (h *ProxyHandler) LXCAction(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	action := c.Param("action")

	var upid string
	ctx := c.Request.Context()

	switch action {
	case "start":
		upid, err = h.client.StartLXC(ctx, node, vmid)
	case "stop":
		upid, err = h.client.StopLXC(ctx, node, vmid)
	case "shutdown":
		upid, err = h.client.ShutdownLXC(ctx, node, vmid)
	case "reboot":
		upid, err = h.client.RebootLXC(ctx, node, vmid)
	case "freeze":
		upid, err = h.client.FreezeLXC(ctx, node, vmid)
	case "unfreeze":
		upid, err = h.client.UnfreezeLXC(ctx, node, vmid)
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
// PUT /api/pve/nodes/:node/lxc/:vmid/config
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

	upid, err := h.client.SetLXCConfig(c.Request.Context(), node, vmid, config)
	if err != nil {
		h.logger.Error("更新 LXC 配置失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "更新容器配置失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器配置更新任务已提交"})
}

// DeleteLXC 删除容器
// DELETE /api/pve/nodes/:node/lxc/:vmid
func (h *ProxyHandler) DeleteLXC(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	upid, err := h.client.DeleteLXC(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("删除 LXC 失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "删除容器失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器删除任务已提交"})
}

// GetLXCSnapshots 获取容器快照列表
// GET /api/pve/nodes/:node/lxc/:vmid/snapshot
func (h *ProxyHandler) GetLXCSnapshots(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}

	snapshots, err := h.client.ListLXCSnapshots(c.Request.Context(), node, vmid)
	if err != nil {
		h.logger.Error("获取 LXC 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取容器快照失败: "+err.Error())
		return
	}

	h.success(c, snapshots)
}

// CreateLXCSnapshot 创建容器快照
// POST /api/pve/nodes/:node/lxc/:vmid/snapshot
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

	upid, err := h.client.CreateLXCSnapshot(c.Request.Context(), node, vmid, req.Name, req.Description)
	if err != nil {
		h.logger.Error("创建 LXC 快照失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "创建容器快照失败: "+err.Error())
		return
	}

	h.success(c, gin.H{"upid": upid, "message": "容器快照创建任务已提交"})
}

// DeleteLXCSnapshot 删除容器快照
// DELETE /api/pve/nodes/:node/lxc/:vmid/snapshot/:snapname
func (h *ProxyHandler) DeleteLXCSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	snapname := c.Param("snapname")

	upid, err := h.client.DeleteLXCSnapshot(c.Request.Context(), node, vmid, snapname)
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
// GET /api/pve/nodes/:node/status
func (h *ProxyHandler) GetNodeStatus(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}

	status, err := h.client.GetNodeStatus(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点状态失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点状态失败: "+err.Error())
		return
	}

	h.success(c, status)
}

// GetNodeVersion 获取节点版本
// GET /api/pve/nodes/:node/version
func (h *ProxyHandler) GetNodeVersion(c *gin.Context) {
	node := c.Param("node")
	version, err := h.client.GetNodeVersion(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点版本失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点版本失败: "+err.Error())
		return
	}

	h.success(c, version)
}

// GetNodeServices 获取节点服务列表
// GET /api/pve/nodes/:node/services
func (h *ProxyHandler) GetNodeServices(c *gin.Context) {
	node := c.Param("node")
	services, err := h.client.GetNodeServices(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点服务列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取节点服务列表失败: "+err.Error())
		return
	}

	h.success(c, services)
}

// GetNodeSyslog 获取节点系统日志
// GET /api/pve/nodes/:node/syslog
func (h *ProxyHandler) GetNodeSyslog(c *gin.Context) {
	node := c.Param("node")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))

	logs, err := h.client.GetNodeSyslog(c.Request.Context(), node, limit, start)
	if err != nil {
		h.logger.Error("获取系统日志失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取系统日志失败: "+err.Error())
		return
	}

	h.success(c, logs)
}

// GetNodeTasks 获取节点任务列表
// GET /api/pve/nodes/:node/tasks
func (h *ProxyHandler) GetNodeTasks(c *gin.Context) {
	node := c.Param("node")
	tasks, err := h.client.GetNodeTasks(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取节点任务列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取任务列表失败: "+err.Error())
		return
	}

	h.success(c, tasks)
}

// GetNodeTaskStatus 获取任务状态
// GET /api/pve/nodes/:node/tasks/:upid/status
func (h *ProxyHandler) GetNodeTaskStatus(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")

	status, err := h.client.GetNodeTaskStatus(c.Request.Context(), node, upid)
	if err != nil {
		h.logger.Error("获取任务状态失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "获取任务状态失败: "+err.Error())
		return
	}

	h.success(c, status)
}

// GetNodeTaskLog 获取任务日志
// GET /api/pve/nodes/:node/tasks/:upid/log
func (h *ProxyHandler) GetNodeTaskLog(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")

	logs, err := h.client.GetNodeTaskLog(c.Request.Context(), node, upid)
	if err != nil {
		h.logger.Error("获取任务日志失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "获取任务日志失败: "+err.Error())
		return
	}

	h.success(c, logs)
}

// WaitForTask 等待任务完成（轮询）
// GET /api/pve/nodes/:node/tasks/:upid/wait
func (h *ProxyHandler) WaitForTask(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	timeoutSec, _ := strconv.Atoi(c.DefaultQuery("timeout", "60"))
	timeout := time.Duration(timeoutSec) * time.Second

	status, err := h.client.WaitForTask(c.Request.Context(), node, upid, timeout)
	if err != nil {
		h.logger.Error("等待任务完成失败",
			zap.String("node", node), zap.String("upid", upid), zap.Error(err))
		h.serverError(c, "等待任务超时: "+err.Error())
		return
	}

	h.success(c, status)
}

// GetNodeNetwork 获取网络接口列表
// GET /api/pve/nodes/:node/network
func (h *ProxyHandler) GetNodeNetwork(c *gin.Context) {
	node := c.Param("node")
	interfaces, err := h.client.GetNodeNetwork(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取网络接口列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取网络接口列表失败: "+err.Error())
		return
	}

	h.success(c, interfaces)
}

// GetNodeDNS 获取 DNS 配置
// GET /api/pve/nodes/:node/dns
func (h *ProxyHandler) GetNodeDNS(c *gin.Context) {
	node := c.Param("node")
	dns, err := h.client.GetNodeDNS(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 DNS 配置失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取 DNS 配置失败: "+err.Error())
		return
	}

	h.success(c, dns)
}

// GetNodeTime 获取时间信息
// GET /api/pve/nodes/:node/time
func (h *ProxyHandler) GetNodeTime(c *gin.Context) {
	node := c.Param("node")
	timeInfo, err := h.client.GetNodeTime(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取时间信息失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取时间信息失败: "+err.Error())
		return
	}

	h.success(c, timeInfo)
}

// GetNodeAPTUpdate 获取可更新软件包
// GET /api/pve/nodes/:node/apt/update
func (h *ProxyHandler) GetNodeAPTUpdate(c *gin.Context) {
	node := c.Param("node")
	updates, err := h.client.GetNodeAPTUpdate(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取软件包更新列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取软件包更新列表失败: "+err.Error())
		return
	}

	h.success(c, updates)
}

// GetNodeRRD 获取节点 RRD 性能数据
// GET /api/pve/nodes/:node/rrd
func (h *ProxyHandler) GetNodeRRD(c *gin.Context) {
	node := c.Param("node")
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.Query("dataset")

	data, err := h.client.GetNodeRRD(c.Request.Context(), node, timeframe, dataset)
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
// GET /api/pve/nodes/:node/storage
func (h *ProxyHandler) GetStorageList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}

	storages, err := h.client.ListStorage(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取存储列表失败", zap.String("node", node), zap.Error(err))
		h.serverError(c, "获取存储列表失败: "+err.Error())
		return
	}

	h.success(c, storages)
}

// GetStorageStatus 获取存储状态
// GET /api/pve/nodes/:node/storage/:storage/status
func (h *ProxyHandler) GetStorageStatus(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")

	status, err := h.client.GetStorageStatus(c.Request.Context(), node, storage)
	if err != nil {
		h.logger.Error("获取存储状态失败",
			zap.String("node", node), zap.String("storage", storage), zap.Error(err))
		h.serverError(c, "获取存储状态失败: "+err.Error())
		return
	}

	h.success(c, status)
}

// GetStorageContent 获取存储内容
// GET /api/pve/nodes/:node/storage/:storage/content
func (h *ProxyHandler) GetStorageContent(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	contentType := c.Query("content")

	var content []pve.StorageContent
	var err error

	if contentType != "" {
		content, err = h.client.GetStorageContentByType(c.Request.Context(), node, storage, contentType)
	} else {
		content, err = h.client.GetStorageContent(c.Request.Context(), node, storage)
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
// POST /api/pve/nodes/:node/storage/:storage/download-url
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

	upid, err := h.client.DownloadISO(c.Request.Context(), node, storage, req.URL, req.Filename)
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
// GET /api/pve/cluster/resources
func (h *ProxyHandler) GetClusterResources(c *gin.Context) {
	resourceType := c.Query("type")
	var resources []pve.ClusterResource
	var err error

	if resourceType != "" {
		resources, err = h.client.GetClusterResourcesByType(c.Request.Context(), resourceType)
	} else {
		resources, err = h.client.GetClusterResources(c.Request.Context())
	}

	if err != nil {
		h.logger.Error("获取集群资源失败", zap.Error(err))
		h.serverError(c, "获取集群资源失败: "+err.Error())
		return
	}

	h.success(c, resources)
}

// GetClusterTasks 获取集群任务
// GET /api/pve/cluster/tasks
func (h *ProxyHandler) GetClusterTasks(c *gin.Context) {
	tasks, err := h.client.GetClusterTasks(c.Request.Context())
	if err != nil {
		h.logger.Error("获取集群任务失败", zap.Error(err))
		h.serverError(c, "获取集群任务失败: "+err.Error())
		return
	}

	h.success(c, tasks)
}

// GetNextID 获取下一个 VM ID
// GET /api/pve/cluster/nextid
func (h *ProxyHandler) GetNextID(c *gin.Context) {
	nextID, err := h.client.GetNextID(c.Request.Context())
	if err != nil {
		h.logger.Error("获取下一个 VM ID 失败", zap.Error(err))
		h.serverError(c, "获取下一个 VM ID 失败: "+err.Error())
		return
	}

	h.success(c, nextID)
}

// GetHAConfig 获取 HA 配置
// GET /api/pve/cluster/ha/config
func (h *ProxyHandler) GetHAConfig(c *gin.Context) {
	config, err := h.client.GetHAConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 HA 配置失败", zap.Error(err))
		h.serverError(c, "获取 HA 配置失败: "+err.Error())
		return
	}

	h.success(c, config)
}

// GetSDNZones 获取 SDN 区域
// GET /api/pve/cluster/sdn/zones
func (h *ProxyHandler) GetSDNZones(c *gin.Context) {
	zones, err := h.client.GetSDNZones(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 SDN 区域失败", zap.Error(err))
		h.serverError(c, "获取 SDN 区域失败: "+err.Error())
		return
	}

	h.success(c, zones)
}

// GetSDNVNETs 获取 SDN 虚拟网络
// GET /api/pve/cluster/sdn/vnets
func (h *ProxyHandler) GetSDNVNETs(c *gin.Context) {
	vnets, err := h.client.GetSDNVNETs(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 SDN 虚拟网络失败", zap.Error(err))
		h.serverError(c, "获取 SDN 虚拟网络失败: "+err.Error())
		return
	}

	h.success(c, vnets)
}

// GetPoolList 获取资源池列表
// GET /api/pve/pools
func (h *ProxyHandler) GetPoolList(c *gin.Context) {
	pools, err := h.client.ListPools(c.Request.Context())
	if err != nil {
		h.logger.Error("获取资源池列表失败", zap.Error(err))
		h.serverError(c, "获取资源池列表失败: "+err.Error())
		return
	}

	h.success(c, pools)
}

// GetPool 获取资源池详情
// GET /api/pve/pools/:poolid
func (h *ProxyHandler) GetPool(c *gin.Context) {
	poolid := c.Param("poolid")
	detail, err := h.client.GetPool(c.Request.Context(), poolid)
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
// GET /api/pve/access/users
func (h *ProxyHandler) GetUsers(c *gin.Context) {
	users, err := h.client.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("获取用户列表失败", zap.Error(err))
		h.serverError(c, "获取用户列表失败: "+err.Error())
		return
	}

	h.success(c, users)
}

// GetGroups 获取组列表
// GET /api/pve/access/groups
func (h *ProxyHandler) GetGroups(c *gin.Context) {
	groups, err := h.client.ListGroups(c.Request.Context())
	if err != nil {
		h.logger.Error("获取组列表失败", zap.Error(err))
		h.serverError(c, "获取组列表失败: "+err.Error())
		return
	}

	h.success(c, groups)
}

// GetRoles 获取角色列表
// GET /api/pve/access/roles
func (h *ProxyHandler) GetRoles(c *gin.Context) {
	roles, err := h.client.ListRoles(c.Request.Context())
	if err != nil {
		h.logger.Error("获取角色列表失败", zap.Error(err))
		h.serverError(c, "获取角色列表失败: "+err.Error())
		return
	}

	h.success(c, roles)
}

// GetACLs 获取 ACL 列表
// GET /api/pve/access/acl
func (h *ProxyHandler) GetACLs(c *gin.Context) {
	acls, err := h.client.ListACLs(c.Request.Context())
	if err != nil {
		h.logger.Error("获取 ACL 列表失败", zap.Error(err))
		h.serverError(c, "获取 ACL 列表失败: "+err.Error())
		return
	}

	h.success(c, acls)
}

// GetDomains 获取认证域列表
// GET /api/pve/access/domains
func (h *ProxyHandler) GetDomains(c *gin.Context) {
	domains, err := h.client.ListDomains(c.Request.Context())
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
// code: 0, message: success, data: 数据
func (h *ProxyHandler) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// badRequest 返回请求错误
// code: 400, message: 错误信息
func (h *ProxyHandler) badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    400,
		"message": message,
	})
}

// serverError 返回服务器错误
// code: 500, message: 错误信息
func (h *ProxyHandler) serverError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    500,
		"message": message,
	})
}

// VNCProxy 处理 VNC 代理 WebSocket 连接
// GET /api/pve/nodes/:node/qemu/:vmid/vncproxy
// 此方法仅获取 VNC ticket，WebSocket 连接需要单独处理
func (h *ProxyHandler) VNCProxy(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}

	path := fmt.Sprintf("nodes/%s/qemu/%d/vncproxy", node, vmid)
	params := map[string]interface{}{
		"websocket": 1,
	}

	resp, err := h.client.Do(context.Background(), "POST", path, params)
	if err != nil {
		h.logger.Error("获取 VNC 代理失败",
			zap.String("node", node), zap.Int("vmid", vmid), zap.Error(err))
		h.serverError(c, "获取 VNC 代理失败: "+err.Error())
		return
	}

	h.success(c, resp.Data)
}
