package handler

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
	"go.uber.org/zap"
)

// ProxyHandler API 代理处理器
// 将前端请求代理到 PVE API，处理路径映射和认证传递
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

// Proxy 代理请求到 PVE API
// POST/GET /api/pve/*
// 自动将请求路径和方法转发到 PVE，返回原始响应
func (h *ProxyHandler) Proxy(c *gin.Context) {
	// 获取代理路径
	proxyPath := c.Param("proxyPath")
	if proxyPath == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "代理路径不能为空",
		})
		return
	}

	// 检查 PVE 客户端是否已认证
	if !h.client.IsAuthenticated() {
		c.JSON(401, gin.H{
			"code":    401,
			"message": "PVE 客户端未认证",
		})
		return
	}

	// 执行代理请求
	var reqBody io.Reader
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
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
		c.JSON(502, gin.H{
			"code":    502,
			"message": "代理请求失败: " + err.Error(),
		})
		return
	}

	// 返回原始响应
	c.Data(statusCode, "application/json", respBody)
}

// GetNodes 获取节点列表
// GET /api/pve/nodes
// 返回集群中所有节点的详细信息
func (h *ProxyHandler) GetNodes(c *gin.Context) {
	nodes, err := h.client.GetNodes(c.Request.Context())
	if err != nil {
		h.logger.Error("获取节点列表失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取节点列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    nodes,
	})
}

// GetVMs 获取虚拟机列表
// GET /api/pve/nodes/:node/qemu
// 返回指定节点上所有虚拟机的信息
func (h *ProxyHandler) GetVMs(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "节点名称不能为空",
		})
		return
	}

	vms, err := h.client.GetVMs(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取虚拟机列表失败",
			zap.String("node", node),
			zap.Error(err),
		)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取虚拟机列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    vms,
	})
}

// GetLXCs 获取 LXC 容器列表
// GET /api/pve/nodes/:node/lxc
// 返回指定节点上所有 LXC 容器的信息
func (h *ProxyHandler) GetLXCs(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "节点名称不能为空",
		})
		return
	}

	lxcs, err := h.client.GetLXCs(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取 LXC 容器列表失败",
			zap.String("node", node),
			zap.Error(err),
		)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取 LXC 容器列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    lxcs,
	})
}

// GetStorages 获取存储列表
// GET /api/pve/nodes/:node/storage
// 返回指定节点上所有存储的信息
func (h *ProxyHandler) GetStorages(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "节点名称不能为空",
		})
		return
	}

	storages, err := h.client.GetStorages(c.Request.Context(), node)
	if err != nil {
		h.logger.Error("获取存储列表失败",
			zap.String("node", node),
			zap.Error(err),
		)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取存储列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    storages,
	})
}
