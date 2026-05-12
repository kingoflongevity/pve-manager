package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

// AppStoreHandler 应用商店 HTTP 处理器
type AppStoreHandler struct {
	appStoreService *service.AppStoreService
	logger          *zap.Logger
}

// NewAppStoreHandler 创建应用商店处理器
func NewAppStoreHandler(appStoreService *service.AppStoreService, logger *zap.Logger) *AppStoreHandler {
	return &AppStoreHandler{
		appStoreService: appStoreService,
		logger:          logger,
	}
}

// GetAppTemplates 获取应用模板列表
func (h *AppStoreHandler) GetAppTemplates(c *gin.Context) {
	category := c.Query("category")
	templates, err := h.appStoreService.GetAppTemplates(category)
	if err != nil {
		h.serverError(c, "获取应用模板失败: "+err.Error())
		return
	}
	h.success(c, templates)
}

// GetAppTemplateDetail 获取应用模板详情
func (h *AppStoreHandler) GetAppTemplateDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}

	tpl, err := h.appStoreService.GetAppTemplateByID(uint(id))
	if err != nil {
		h.serverError(c, "获取应用模板详情失败: "+err.Error())
		return
	}
	h.success(c, tpl)
}

// GetAppCategories 获取应用分类
func (h *AppStoreHandler) GetAppCategories(c *gin.Context) {
	categories, err := h.appStoreService.GetAppCategories()
	if err != nil {
		h.serverError(c, "获取应用分类失败: "+err.Error())
		return
	}
	h.success(c, categories)
}

// DeployApp 部署应用到 PVE（真实部署引擎）
func (h *AppStoreHandler) DeployApp(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}

	var req struct {
		TemplateID uint              `json:"template_id" binding:"required"`
		Name       string            `json:"name" binding:"required"`
		Node       string            `json:"node" binding:"required"`
		Config     map[string]string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	template, err := h.appStoreService.GetAppTemplateByID(req.TemplateID)
	if err != nil {
		h.serverError(c, "获取应用模板失败: "+err.Error())
		return
	}

	deployment, err := h.appStoreService.DeployAppReal(template, req.Name, req.Node, req.Config, userID)
	if err != nil {
		h.serverError(c, "创建部署任务失败: "+err.Error())
		return
	}

	h.success(c, gin.H{
		"deployment_id": deployment.ID,
		"message":       "应用部署任务已提交",
	})
}

// GetAppDeployments 获取应用部署列表
func (h *AppStoreHandler) GetAppDeployments(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}

	deployments, err := h.appStoreService.GetAppDeployments(userID)
	if err != nil {
		h.serverError(c, "获取部署列表失败: "+err.Error())
		return
	}
	h.success(c, deployments)
}

// GetAppDeploymentDetail 获取应用部署详情
func (h *AppStoreHandler) GetAppDeploymentDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}

	deployment, err := h.appStoreService.GetAppDeploymentByID(uint(id))
	if err != nil {
		h.serverError(c, "获取部署详情失败: "+err.Error())
		return
	}
	h.success(c, deployment)
}

// DeleteAppDeployment 卸载/取消应用部署
func (h *AppStoreHandler) DeleteAppDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}

	if err := h.appStoreService.UninstallDeployment(uint(id)); err != nil {
		h.serverError(c, "卸载应用失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

// ImportTemplate 导入应用模板
func (h *AppStoreHandler) ImportTemplate(c *gin.Context) {
	var req struct {
		YAML string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tpl, err := h.appStoreService.ImportTemplate(req.YAML)
	if err != nil {
		h.serverError(c, "导入模板失败: "+err.Error())
		return
	}
	h.success(c, tpl)
}

// SyncTemplates 同步远程模板
func (h *AppStoreHandler) SyncTemplates(c *gin.Context) {
	var req struct {
		RemoteURL string `json:"remote_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	count, err := h.appStoreService.SyncRemoteTemplates(req.RemoteURL)
	if err != nil {
		h.serverError(c, "同步模板失败: "+err.Error())
		return
	}
	h.success(c, gin.H{"synced_count": count})
}

// ==================== 辅助方法 ====================

func (h *AppStoreHandler) success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func (h *AppStoreHandler) badRequest(c *gin.Context, msg string) {
	c.JSON(400, gin.H{
		"code":    400,
		"message": msg,
	})
}

func (h *AppStoreHandler) serverError(c *gin.Context, msg string) {
	c.JSON(500, gin.H{
		"code":    500,
		"message": msg,
	})
}
