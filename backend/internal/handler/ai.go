package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

type AIHandler struct {
	aiService *service.AIService
	logger    *zap.Logger
}

func NewAIHandler(aiService *service.AIService, logger *zap.Logger) *AIHandler {
	return &AIHandler{aiService: aiService, logger: logger}
}

// ==================== 模型配置管理 ====================

func (h *AIHandler) GetModelConfigs(c *gin.Context) {
	configs, err := h.aiService.GetModelConfigs()
	if err != nil {
		h.serverError(c, "获取模型配置失败: "+err.Error())
		return
	}
	h.success(c, configs)
}

func (h *AIHandler) CreateModelConfig(c *gin.Context) {
	var req model.AIModelConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.aiService.CreateModelConfig(&req); err != nil {
		h.serverError(c, "创建模型配置失败: "+err.Error())
		return
	}
	h.success(c, req)
}

func (h *AIHandler) UpdateModelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.aiService.UpdateModelConfig(uint(id), updates); err != nil {
		h.serverError(c, "更新模型配置失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

func (h *AIHandler) DeleteModelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	if err := h.aiService.DeleteModelConfig(uint(id)); err != nil {
		h.serverError(c, "删除模型配置失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

func (h *AIHandler) SetDefaultModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	if err := h.aiService.SetDefaultModel(uint(id)); err != nil {
		h.serverError(c, "设置默认模型失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

func (h *AIHandler) TestModelConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	cfg, err := h.aiService.GetModelConfigByID(uint(id))
	if err != nil {
		h.serverError(c, "获取模型配置失败: "+err.Error())
		return
	}
	success, msg := h.aiService.TestModelConnection(c.Request.Context(), cfg)
	if !success {
		h.badRequest(c, "连接测试失败: "+msg)
		return
	}
	h.success(c, gin.H{"message": "连接测试成功"})
}

// ==================== 对话管理 ====================

func (h *AIHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	conversations, err := h.aiService.GetConversations(userID, limit)
	if err != nil {
		h.serverError(c, "获取对话列表失败: "+err.Error())
		return
	}
	h.success(c, conversations)
}

func (h *AIHandler) GetConversationDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	conv, err := h.aiService.GetConversationByID(uint(id))
	if err != nil {
		h.serverError(c, "获取对话详情失败: "+err.Error())
		return
	}
	h.success(c, conv)
}

// CreateConversation 创建对话并返回 AI 回复
// 使用普通 JSON 响应格式，内置 PVEA 知识库智能回复
func (h *AIHandler) CreateConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}

	var req struct {
		Title         string `json:"title"`
		Scene         string `json:"scene" binding:"required"`
		ModelConfigID uint   `json:"model_config_id"`
		Message       string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Title == "" {
		req.Title = req.Message
		if len(req.Title) > 50 {
			req.Title = req.Title[:50]
		}
	}

	// 获取模型配置
	var modelConfigID uint
	cfg, err := h.aiService.GetDefaultModel()
	if err == nil {
		modelConfigID = cfg.ID
	}

	conv := &model.AIConversation{
		Title:         req.Title,
		Scene:         req.Scene,
		ModelConfigID: modelConfigID,
		UserID:        userID,
	}
	if err := h.aiService.CreateConversation(conv); err != nil {
		h.serverError(c, "创建对话失败: "+err.Error())
		return
	}

	userMsg := &model.AIMessage{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        req.Message,
	}
	if err := h.aiService.AddMessage(userMsg); err != nil {
		h.serverError(c, "保存消息失败: "+err.Error())
		return
	}

	// 生成 AI 回复
	aiReply := h.aiService.GenerateChatResponse(req.Message, nil)

	aiMsg := &model.AIMessage{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        aiReply,
	}
	if err := h.aiService.AddMessage(aiMsg); err != nil {
		h.serverError(c, "保存 AI 回复失败: "+err.Error())
		return
	}

	// 返回完整对话
	convDetail, err := h.aiService.GetConversationByID(conv.ID)
	if err != nil {
		h.serverError(c, "获取对话失败: "+err.Error())
		return
	}

	h.success(c, convDetail)
}

// SendMessage 在已有对话中发送消息
func (h *AIHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}

	conv, err := h.aiService.GetConversationByID(uint(id))
	if err != nil {
		h.serverError(c, "获取对话失败: "+err.Error())
		return
	}
	if conv.UserID != userID {
		h.badRequest(c, "无权访问此对话")
		return
	}

	userMsg := &model.AIMessage{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        req.Message,
	}
	if err := h.aiService.AddMessage(userMsg); err != nil {
		h.serverError(c, "保存消息失败: "+err.Error())
		return
	}

	// 获取历史消息用于上下文
	history, _ := h.aiService.GetConversationMessages(uint(id))
	aiReply := h.aiService.GenerateChatResponse(req.Message, history)

	aiMsg := &model.AIMessage{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        aiReply,
	}
	if err := h.aiService.AddMessage(aiMsg); err != nil {
		h.serverError(c, "保存 AI 回复失败: "+err.Error())
		return
	}

	convDetail, err := h.aiService.GetConversationByID(conv.ID)
	if err != nil {
		h.serverError(c, "获取对话失败: "+err.Error())
		return
	}

	h.success(c, convDetail)
}

func (h *AIHandler) DeleteConversation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	if err := h.aiService.DeleteConversation(uint(id)); err != nil {
		h.serverError(c, "删除对话失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

// ==================== 报告管理 ====================

func (h *AIHandler) GetReports(c *gin.Context) {
	reportType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	reports, err := h.aiService.GetReports(reportType, limit)
	if err != nil {
		h.serverError(c, "获取报告列表失败: "+err.Error())
		return
	}
	h.success(c, reports)
}

func (h *AIHandler) GetReportDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	report, err := h.aiService.GetReportByID(uint(id))
	if err != nil {
		h.serverError(c, "获取报告详情失败: "+err.Error())
		return
	}
	h.success(c, report)
}

func (h *AIHandler) GenerateReport(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}
	var req struct {
		Title         string `json:"title" binding:"required"`
		Type          string `json:"type" binding:"required"`
		ModelConfigID uint   `json:"model_config_id"`
		Content       string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	report := &model.AIReport{
		Title:         req.Title,
		Type:          req.Type,
		Content:       req.Content,
		ModelConfigID: req.ModelConfigID,
		UserID:        userID,
	}
	if err := h.aiService.CreateReport(report); err != nil {
		h.serverError(c, "创建报告失败: "+err.Error())
		return
	}
	h.success(c, report)
}

func (h *AIHandler) DeleteReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.badRequest(c, "无效的 ID")
		return
	}
	if err := h.aiService.DeleteReport(uint(id)); err != nil {
		h.serverError(c, "删除报告失败: "+err.Error())
		return
	}
	h.success(c, nil)
}

// ==================== 诊断和建议 ====================

func (h *AIHandler) Diagnose(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		h.badRequest(c, "未登录")
		return
	}
	var req struct {
		Node    string `json:"node" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	diagnosticData, err := h.aiService.GetPVEDiagnosticData(c.Request.Context(), req.Node)
	if err != nil {
		h.serverError(c, "获取诊断数据失败: "+err.Error())
		return
	}
	h.success(c, gin.H{
		"diagnostic_data": diagnosticData,
		"message":         "系统诊断完成",
	})
}

func (h *AIHandler) GetSuggestion(c *gin.Context) {
	var req struct {
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   string `json:"resource_id" binding:"required"`
		Message      string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "请求参数错误: "+err.Error())
		return
	}
	h.success(c, gin.H{
		"message":         "建议生成成功",
		"resource_type":   req.ResourceType,
		"resource_id":     req.ResourceID,
		"suggestion":      "建议在生产环境中为关键 VM 配置 HA（高可用），并定期进行快照备份。",
	})
}

// ==================== 辅助方法 ====================

func (h *AIHandler) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func (h *AIHandler) badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    400,
		"message": msg,
	})
}

func (h *AIHandler) serverError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    500,
		"message": msg,
	})
}

type StreamWriter struct {
	Ctx     *gin.Context
	Flusher http.Flusher
}

func (w *StreamWriter) Write(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Ctx.Writer.WriteString("data: " + string(jsonData) + "\n\n")
	if err != nil {
		return err
	}
	w.Flusher.Flush()
	return nil
}
