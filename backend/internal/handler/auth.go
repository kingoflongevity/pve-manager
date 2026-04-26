package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

// AuthHandler 认证相关 HTTP 处理器
// 仅负责参数提取、验证和响应格式化，业务逻辑委托给 AuthService
type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(logger *zap.Logger, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login 处理用户登录请求
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		// 根据错误类型返回不同状态码
		if err.Error() != "" && (len(err.Error()) > 10 && err.Error()[:10] == "认证失败") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
			return
		}
		h.logger.Error("登录处理失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "登录成功", "data": resp})
}

// GetPVEConfigs 获取 PVE 配置列表
// GET /api/admin/pve-configs
func (h *AuthHandler) GetPVEConfigs(c *gin.Context) {
	configs, err := h.authService.GetPVEConfigs()
	if err != nil {
		h.logger.Error("获取 PVE 配置列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取配置列表失败"})
		return
	}

	resp := make([]PVEConfigResponse, 0, len(configs))
	for _, cfg := range configs {
		resp = append(resp, PVEConfigResponse{
			ID: cfg.ID, Host: cfg.Host, Port: cfg.Port, Realm: cfg.Realm,
			Username: cfg.Username, Name: cfg.Name, IsDefault: cfg.IsDefault,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": resp})
}

// GetAuditLogs 获取审计日志（分页）
// GET /api/admin/audit-logs
func (h *AuthHandler) GetAuditLogs(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	userID := c.Query("user_id")
	action := c.Query("action")

	logs, total, err := h.authService.GetAuditLogs(page, pageSize, userID, action)
	if err != nil {
		h.logger.Error("获取审计日志失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取审计日志失败"})
		return
	}

	resp := make([]AuditLogResponse, 0, len(logs))
	for _, l := range logs {
		resp = append(resp, AuditLogResponse{
			ID: l.ID, Username: l.Username, Action: l.Action,
			Resource: l.Resource, Detail: l.Detail, IP: l.IP,
			Status: l.Status, CreatedAt: l.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": PaginationResponse{
		Total: total, Page: page, Size: pageSize, Items: resp,
	}})
}

// PVEConfigResponse PVE 配置响应结构
type PVEConfigResponse struct {
	ID        uint   `json:"id"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Realm     string `json:"realm"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// AuditLogResponse 审计日志响应结构
type AuditLogResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Items interface{} `json:"items"`
}
