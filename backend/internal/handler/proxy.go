package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

// ProxyHandler PVE API 代理 HTTP 处理器
// 仅负责参数提取、验证和响应格式化，业务逻辑委托给各 Service
type ProxyHandler struct {
	authService       *service.AuthService
	clusterService    *service.ClusterService
	vmService         *service.VMService
	containerService  *service.ContainerService
	storageService    *service.StorageService
	nodeService       *service.NodeService
	logger            *zap.Logger
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(
	logger *zap.Logger,
	authService *service.AuthService,
	clusterService *service.ClusterService,
	vmService *service.VMService,
	containerService *service.ContainerService,
	storageService *service.StorageService,
	nodeService *service.NodeService,
) *ProxyHandler {
	return &ProxyHandler{
		authService:      authService,
		clusterService:   clusterService,
		vmService:        vmService,
		containerService: containerService,
		storageService:   storageService,
		nodeService:      nodeService,
		logger:           logger,
	}
}

// buildClient 从请求 header 中提取 JWT 并构建已认证的 PVE 客户端
func (h *ProxyHandler) buildClient(c *gin.Context) (*pve.Client, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("未提供认证令牌")
	}
	tokenString := authHeader[7:]
	client, err := h.authService.BuildPVEClientFromToken(tokenString)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ==================== 集群管理 ====================

func (h *ProxyHandler) GetClusterResources(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	resourceType := c.Query("type")
	var data interface{}
	if resourceType != "" {
		data, err = h.clusterService.GetResourcesByType(c.Request.Context(), client, resourceType)
	} else {
		data, err = h.clusterService.GetResources(c.Request.Context(), client)
	}
	if err != nil {
		h.serverError(c, "获取集群资源失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetClusterTasks(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetTasks(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取集群任务失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNextID(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetNextID(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取下一个 VM ID 失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetHAConfig(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetHAConfig(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取 HA 配置失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetSDNZones(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetSDNZones(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取 SDN 区域失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetSDNVNETs(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetSDNVNETs(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取 SDN 虚拟网络失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetPoolList(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.ListPools(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取资源池列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetPool(c *gin.Context) {
	poolid := c.Param("poolid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetPool(c.Request.Context(), client, poolid)
	if err != nil {
		h.serverError(c, "获取资源池详情失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== 访问控制 ====================

func (h *ProxyHandler) GetUsers(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.ListUsers(c.Request.Context())
	if err != nil {
		h.serverError(c, "获取用户列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetGroups(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.ListGroups(c.Request.Context())
	if err != nil {
		h.serverError(c, "获取组列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetRoles(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.ListRoles(c.Request.Context())
	if err != nil {
		h.serverError(c, "获取角色列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetACLs(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.ListACLs(c.Request.Context())
	if err != nil {
		h.serverError(c, "获取 ACL 列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetDomains(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.ListDomains(c.Request.Context())
	if err != nil {
		h.serverError(c, "获取认证域列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== 节点操作 ====================

func (h *ProxyHandler) GetNodeStatus(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetStatus(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取节点状态失败: "+err.Error())
		return
	}
	flattened := flattenNodeStatus(data)
	flattened["node"] = node
	h.success(c, flattened)
}

/**
 * 将 PVE 9.x 嵌套的节点状态数据展平为前端期望的格式
 */
func flattenNodeStatus(data interface{}) map[string]interface{} {
	raw, ok := data.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"raw": data}
	}
	result := make(map[string]interface{})
	for k, v := range raw {
		result[k] = v
	}
	if mem, ok := raw["memory"].(map[string]interface{}); ok {
		if _, has := result["maxmem"]; !has {
			result["maxmem"] = mem["total"]
			result["mem"] = mem["used"]
		}
	}
	if swap, ok := raw["swap"].(map[string]interface{}); ok {
		if _, has := result["maxswap"]; !has {
			result["maxswap"] = swap["total"]
			result["swap"] = swap["used"]
		}
	}
	if cpuinfo, ok := raw["cpuinfo"].(map[string]interface{}); ok {
		if _, has := result["cpus"]; !has {
			result["cpus"] = cpuinfo["cpus"]
			result["maxcpu"] = cpuinfo["cpus"]
		}
	}
	if rootfs, ok := raw["rootfs"].(map[string]interface{}); ok {
		if _, has := result["maxdisk"]; !has {
			result["maxdisk"] = rootfs["total"]
			result["disk"] = rootfs["used"]
		}
	}
	return result
}

func (h *ProxyHandler) GetNodeVersion(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetVersion(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取节点版本失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeServices(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetServices(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取节点服务列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeSyslog(c *gin.Context) {
	node := c.Param("node")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetSyslog(c.Request.Context(), client, node, limit, start)
	if err != nil {
		h.serverError(c, "获取系统日志失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeTasks(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetTasks(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取任务列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeTaskStatus(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetTaskStatus(c.Request.Context(), client, node, upid)
	if err != nil {
		h.serverError(c, "获取任务状态失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeTaskLog(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetTaskLog(c.Request.Context(), client, node, upid)
	if err != nil {
		h.serverError(c, "获取任务日志失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) WaitForTask(c *gin.Context) {
	node := c.Param("node")
	upid := c.Param("upid")
	timeoutSec, _ := strconv.Atoi(c.DefaultQuery("timeout", "60"))
	timeout := time.Duration(timeoutSec) * time.Second
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.WaitForTask(c.Request.Context(), client, node, upid, timeout)
	if err != nil {
		h.serverError(c, "等待任务超时: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeNetwork(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetNetwork(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取网络接口列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeDNS(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetDNS(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取 DNS 配置失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeTime(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetTime(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取时间信息失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeAPTUpdate(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.nodeService.GetAPTUpdate(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取软件包更新列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetNodeRRD(c *gin.Context) {
	node := c.Param("node")
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.DefaultQuery("ds", "cpu")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	params := map[string]interface{}{
		"timeframe": timeframe,
		"ds":        dataset,
	}
	resp, err := client.Do(c.Request.Context(), "GET", fmt.Sprintf("nodes/%s/rrd", node), params)
	if err != nil {
		h.success(c, []interface{}{})
		return
	}
	h.success(c, resp.Data)
}

// ==================== 存储管理 ====================

func (h *ProxyHandler) GetStorageList(c *gin.Context) {
	node := c.Param("node")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.ListStorage(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取存储列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetStorageStatus(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.GetStorageStatus(c.Request.Context(), client, node, storage)
	if err != nil {
		h.serverError(c, "获取存储状态失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) GetStorageContent(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	contentType := c.Query("content")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.GetStorageContent(c.Request.Context(), client, node, storage, contentType)
	if err != nil {
		h.serverError(c, "获取存储内容失败: "+err.Error())
		return
	}
	h.success(c, data)
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.DownloadISO(c.Request.Context(), client, node, storage, req.URL, req.Filename)
	if err != nil {
		h.serverError(c, "下载 ISO 失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "ISO 下载任务已提交"})
}

// ==================== QEMU 虚拟机 ====================

func (h *ProxyHandler) GetQEMUList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.ListVMs(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取虚拟机列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) CreateQEMU(c *gin.Context) {
	node := c.Param("node")
	var params pve.QEMUCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.CreateVM(c.Request.Context(), client, node, &params)
	if err != nil {
		h.serverError(c, "创建虚拟机失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机创建任务已提交"})
}

func (h *ProxyHandler) GetQEMUConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.GetVMConfig(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "获取虚拟机配置失败: "+err.Error())
		return
	}
	h.success(c, data)
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.SetVMConfig(c.Request.Context(), client, node, vmid, config)
	if err != nil {
		h.serverError(c, "更新虚拟机配置失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机配置更新任务已提交"})
}

func (h *ProxyHandler) QEMUAction(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	action := c.Param("action")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.VMAction(c.Request.Context(), client, node, vmid, action)
	if err != nil {
		h.serverError(c, fmt.Sprintf("虚拟机 %s 操作失败: %s", action, err.Error()))
		return
	}
	if data == "" {
		h.badRequest(c, "不支持的操作: "+action)
		return
	}
	h.success(c, gin.H{"upid": data, "message": fmt.Sprintf("虚拟机 %s 任务已提交", action)})
}

func (h *ProxyHandler) DeleteQEMU(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.DeleteVM(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "删除虚拟机失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机删除任务已提交"})
}

func (h *ProxyHandler) GetQEMUSnapshots(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.GetSnapshots(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "获取虚拟机快照失败: "+err.Error())
		return
	}
	h.success(c, data)
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.CreateSnapshot(c.Request.Context(), client, node, vmid, req.Name, req.Description)
	if err != nil {
		h.serverError(c, "创建虚拟机快照失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机快照创建任务已提交"})
}

func (h *ProxyHandler) DeleteQEMUSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	snapname := c.Param("snapname")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.DeleteSnapshot(c.Request.Context(), client, node, vmid, snapname)
	if err != nil {
		h.serverError(c, "删除虚拟机快照失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机快照删除任务已提交"})
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.CloneVM(c.Request.Context(), client, node, vmid, &params)
	if err != nil {
		h.serverError(c, "克隆虚拟机失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机克隆任务已提交"})
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.MigrateVM(c.Request.Context(), client, node, vmid, &params)
	if err != nil {
		h.serverError(c, "迁移虚拟机失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "虚拟机迁移任务已提交"})
}

func (h *ProxyHandler) GetQEMURRD(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.DefaultQuery("ds", "cpu")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	params := map[string]interface{}{
		"timeframe": timeframe,
		"ds":        dataset,
	}
	resp, err := client.Do(c.Request.Context(), "GET", fmt.Sprintf("nodes/%s/qemu/%d/rrd", node, vmid), params)
	if err != nil {
		h.success(c, []interface{}{})
		return
	}
	h.success(c, resp.Data)
}

func (h *ProxyHandler) GetQEMUPending(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	path := fmt.Sprintf("nodes/%s/qemu/%d/pending", node, vmid)
	if err := client.Get(c.Request.Context(), path, &result); err != nil {
		h.serverError(c, "获取待处理配置失败: "+err.Error())
		return
	}
	h.success(c, result)
}

func (h *ProxyHandler) VNCProxy(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "虚拟机 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.VNCProxy(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "获取 VNC 代理失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== LXC 容器 ====================

func (h *ProxyHandler) GetLXCList(c *gin.Context) {
	node := c.Param("node")
	if node == "" {
		h.badRequest(c, "节点名称不能为空")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.ListContainers(c.Request.Context(), client, node)
	if err != nil {
		h.serverError(c, "获取容器列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) CreateLXC(c *gin.Context) {
	node := c.Param("node")
	var params pve.LXCCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.CreateContainer(c.Request.Context(), client, node, &params)
	if err != nil {
		h.serverError(c, "创建容器失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "容器创建任务已提交"})
}

func (h *ProxyHandler) GetLXCConfig(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.GetContainerConfig(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "获取容器配置失败: "+err.Error())
		return
	}
	h.success(c, data)
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.SetContainerConfig(c.Request.Context(), client, node, vmid, config)
	if err != nil {
		h.serverError(c, "更新容器配置失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "容器配置更新任务已提交"})
}

func (h *ProxyHandler) LXCAction(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	action := c.Param("action")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.ContainerAction(c.Request.Context(), client, node, vmid, action)
	if err != nil {
		h.serverError(c, fmt.Sprintf("容器 %s 操作失败: %s", action, err.Error()))
		return
	}
	if data == "" {
		h.badRequest(c, "不支持的操作: "+action)
		return
	}
	h.success(c, gin.H{"upid": data, "message": fmt.Sprintf("容器 %s 任务已提交", action)})
}

func (h *ProxyHandler) DeleteLXC(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.DeleteContainer(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "删除容器失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "容器删除任务已提交"})
}

func (h *ProxyHandler) GetLXCSnapshots(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.GetSnapshots(c.Request.Context(), client, node, vmid)
	if err != nil {
		h.serverError(c, "获取容器快照失败: "+err.Error())
		return
	}
	h.success(c, data)
}

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
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.CreateSnapshot(c.Request.Context(), client, node, vmid, req.Name, req.Description)
	if err != nil {
		h.serverError(c, "创建容器快照失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "容器快照创建任务已提交"})
}

func (h *ProxyHandler) DeleteLXCSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	snapname := c.Param("snapname")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.DeleteSnapshot(c.Request.Context(), client, node, vmid, snapname)
	if err != nil {
		h.serverError(c, "删除容器快照失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"upid": data, "message": "容器快照删除任务已提交"})
}

// ==================== 集群管理（扩展方法） ====================

// GetClusterStorage 获取集群级存储列表
func (h *ProxyHandler) GetClusterStorage(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	if err := client.Get(c.Request.Context(), "cluster/storage", &result); err != nil {
		h.serverError(c, "获取集群存储列表失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// GetClusterConfig 获取数据中心配置
func (h *ProxyHandler) GetClusterConfig(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	if err := client.Get(c.Request.Context(), "cluster/config", &result); err != nil {
		h.serverError(c, "获取数据中心配置失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// GetClusterLog 获取集群日志
func (h *ProxyHandler) GetClusterLog(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetClusterLog(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取集群日志失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// GetReplicationJobs 获取复制任务列表
func (h *ProxyHandler) GetReplicationJobs(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	if err := client.Get(c.Request.Context(), "cluster/replication", &result); err != nil {
		h.serverError(c, "获取复制任务列表失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// GetHAGroups 获取 HA 组列表
func (h *ProxyHandler) GetHAGroups(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetHAGroups(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取 HA 组列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// GetHAResources 获取 HA 资源列表
func (h *ProxyHandler) GetHAResources(c *gin.Context) {
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.clusterService.GetHAResources(c.Request.Context(), client)
	if err != nil {
		h.serverError(c, "获取 HA 资源列表失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// CreatePool 创建资源池
func (h *ProxyHandler) CreatePool(c *gin.Context) {
	var req struct {
		PoolID  string `json:"poolid" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.CreatePool(c.Request.Context(), req.PoolID, req.Comment)
	if err != nil {
		h.serverError(c, "创建资源池失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== 访问控制（扩展方法） ====================

// GetUser 获取单个用户信息
func (h *ProxyHandler) GetUser(c *gin.Context) {
	userid := c.Param("userid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.GetUser(c.Request.Context(), userid)
	if err != nil {
		h.serverError(c, "获取用户信息失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// UpdateUser 更新用户信息
func (h *ProxyHandler) UpdateUser(c *gin.Context) {
	userid := c.Param("userid")
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.UpdateUser(c.Request.Context(), userid, params)
	if err != nil {
		h.serverError(c, "更新用户信息失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// UpdateUserPassword 修改用户密码
func (h *ProxyHandler) UpdateUserPassword(c *gin.Context) {
	userid := c.Param("userid")
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.SetUserPassword(c.Request.Context(), userid, req.Password)
	if err != nil {
		h.serverError(c, "修改用户密码失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// GetGroup 获取单个用户组信息
func (h *ProxyHandler) GetGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	if err := client.Get(c.Request.Context(), fmt.Sprintf("access/groups/%s", groupid), &result); err != nil {
		h.serverError(c, "获取用户组信息失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// CreateGroup 创建用户组
func (h *ProxyHandler) CreateGroup(c *gin.Context) {
	var params pve.GroupCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.CreateGroup(c.Request.Context(), &params)
	if err != nil {
		h.serverError(c, "创建用户组失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// UpdateGroup 更新用户组信息
func (h *ProxyHandler) UpdateGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.UpdateGroup(c.Request.Context(), groupid, params)
	if err != nil {
		h.serverError(c, "更新用户组失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// DeleteGroup 删除用户组
func (h *ProxyHandler) DeleteGroup(c *gin.Context) {
	groupid := c.Param("groupid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.DeleteGroup(c.Request.Context(), groupid)
	if err != nil {
		h.serverError(c, "删除用户组失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// CreateRole 创建角色
func (h *ProxyHandler) CreateRole(c *gin.Context) {
	var params pve.RoleCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.CreateRole(c.Request.Context(), &params)
	if err != nil {
		h.serverError(c, "创建角色失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// UpdateRole 更新角色权限
func (h *ProxyHandler) UpdateRole(c *gin.Context) {
	roleid := c.Param("roleid")
	var req struct {
		Privs string `json:"privs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.UpdateRole(c.Request.Context(), roleid, req.Privs)
	if err != nil {
		h.serverError(c, "更新角色失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// DeleteRole 删除角色
func (h *ProxyHandler) DeleteRole(c *gin.Context) {
	roleid := c.Param("roleid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.DeleteRole(c.Request.Context(), roleid)
	if err != nil {
		h.serverError(c, "删除角色失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// GetDomain 获取单个认证域信息
func (h *ProxyHandler) GetDomain(c *gin.Context) {
	realm := c.Param("realm")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := client.GetDomain(c.Request.Context(), realm)
	if err != nil {
		h.serverError(c, "获取认证域信息失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// LXCVNCProxy 获取 LXC 容器 VNC 代理
func (h *ProxyHandler) LXCVNCProxy(c *gin.Context) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		h.badRequest(c, "容器 ID 格式错误")
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/vncproxy", node, vmid)
	if err := client.Post(c.Request.Context(), path, nil, &result); err != nil {
		h.serverError(c, "获取 LXC VNC 代理失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// ==================== 辅助方法 ====================

// success 返回成功响应
// ==================== LXC 容器（扩展方法） ====================

func (h *ProxyHandler) GetLXCRRD(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	timeframe := c.DefaultQuery("timeframe", "hour")
	dataset := c.DefaultQuery("ds", "cpu")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	params := map[string]interface{}{
		"timeframe": timeframe,
		"ds":        dataset,
	}
	resp, err := client.Do(c.Request.Context(), "GET", fmt.Sprintf("nodes/%s/lxc/%d/rrd", node, vmid), params)
	if err != nil {
		h.success(c, []interface{}{})
		return
	}
	h.success(c, resp.Data)
}

func (h *ProxyHandler) GetLXCPending(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/pending", node, vmid)
	if err := client.Get(c.Request.Context(), path, &result); err != nil {
		h.serverError(c, "获取 LXC 待处理配置失败: "+err.Error())
		return
	}
	h.success(c, result)
}

func (h *ProxyHandler) CloneLXC(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	var params pve.LXCCloneParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.CloneContainer(c.Request.Context(), client, node, vmid, &params)
	if err != nil {
		h.serverError(c, "克隆 LXC 容器失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) MigrateLXC(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	var params pve.LXCMigrateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.MigrateContainer(c.Request.Context(), client, node, vmid, &params)
	if err != nil {
		h.serverError(c, "迁移 LXC 容器失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) RollbackLXCSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	snapname := c.Param("snapname")
	var req struct {
		Start string `json:"start"`
	}
	c.ShouldBindJSON(&req)
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.containerService.ContainerAction(c.Request.Context(), client, node, vmid, "rollback-"+snapname)
	if err != nil {
		h.serverError(c, "回滚 LXC 快照失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== QEMU 虚拟机（扩展方法） ====================

func (h *ProxyHandler) RollbackQEMUSnapshot(c *gin.Context) {
	node := c.Param("node")
	vmid, _ := strconv.Atoi(c.Param("vmid"))
	snapname := c.Param("snapname")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.vmService.VMAction(c.Request.Context(), client, node, vmid, "rollback-"+snapname)
	if err != nil {
		h.serverError(c, "回滚 QEMU 快照失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== 存储管理（扩展方法） ====================

func (h *ProxyHandler) GetStorageDetail(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.GetStorageStatus(c.Request.Context(), client, node, storage)
	if err != nil {
		h.serverError(c, "获取存储详情失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) CreateStorage(c *gin.Context) {
	node := c.Param("node")
	var params pve.StorageCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.CreateStorage(c.Request.Context(), client, node, params)
	if err != nil {
		h.serverError(c, "创建存储失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) UpdateStorage(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	var params pve.StorageUpdateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.UpdateStorage(c.Request.Context(), client, node, storage, params)
	if err != nil {
		h.serverError(c, "更新存储失败: "+err.Error())
		return
	}
	h.success(c, data)
}

func (h *ProxyHandler) DeleteStorage(c *gin.Context) {
	node := c.Param("node")
	storage := c.Param("storage")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	data, err := h.storageService.DeleteStorage(c.Request.Context(), client, node, storage)
	if err != nil {
		h.serverError(c, "删除存储失败: "+err.Error())
		return
	}
	h.success(c, data)
}

// ==================== 节点管理（扩展方法） ====================

func (h *ProxyHandler) ActionService(c *gin.Context) {
	node := c.Param("node")
	service := c.Param("service")
	action := c.Param("action")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	path := fmt.Sprintf("nodes/%s/services/%s/state", node, service)
	body := map[string]string{"command": action}
	var result interface{}
	if err := client.Post(c.Request.Context(), path, body, &result); err != nil {
		h.serverError(c, "服务操作失败: "+err.Error())
		return
	}
	h.success(c, result)
}

// ==================== 访问控制（扩展方法） ====================

func (h *ProxyHandler) CreateUser(c *gin.Context) {
	var params pve.UserCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	if _, err := client.CreateUser(c.Request.Context(), &params); err != nil {
		h.serverError(c, "创建用户失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

func (h *ProxyHandler) DeleteUser(c *gin.Context) {
	userid := c.Param("userid")
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	if _, err := client.DeleteUser(c.Request.Context(), userid); err != nil {
		h.serverError(c, "删除用户失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

func (h *ProxyHandler) SetACL(c *gin.Context) {
	var params pve.ACLParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.badRequest(c, "参数错误: "+err.Error())
		return
	}
	client, err := h.buildClient(c)
	if err != nil {
		h.serverError(c, "获取 PVE 客户端失败: "+err.Error())
		return
	}
	if _, err := client.SetACL(c.Request.Context(), &params); err != nil {
		h.serverError(c, "设置 ACL 失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

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
	h.logger.Error("服务错误", zap.String("error", message), zap.String("path", c.Request.URL.Path))
	statusCode := http.StatusInternalServerError
	errCode := 500
	if strings.Contains(message, "HTTP 404") || strings.Contains(message, "does not exist") {
		statusCode = http.StatusNotFound
		errCode = 404
	} else if strings.Contains(message, "HTTP 403") || strings.Contains(message, "Permission") {
		statusCode = http.StatusForbidden
		errCode = 403
	} else if strings.Contains(message, "HTTP 401") {
		statusCode = http.StatusUnauthorized
		errCode = 401
	}
	c.JSON(statusCode, gin.H{
		"code":    errCode,
		"message": message,
	})
}
