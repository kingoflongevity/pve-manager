package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"github.com/kingoflongevity/pve-manager/backend/internal/repository"
	"go.uber.org/zap"
)

// PVEContext 从 JWT token 中提取的 PVE 连接上下文
type PVEContext struct {
	Username string
	Host     string
	Port     int
	Password string
	Realm    string
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Realm    string `json:"realm"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// AuthService 认证服务
// 负责用户登录、JWT 生成、密码加密、会话管理
type AuthService struct {
	logger      *zap.Logger
	configRepo  *repository.PVEConfigRepo
	sessionRepo *repository.SessionRepo
	auditRepo   *repository.AuditLogRepo
}

// NewAuthService 创建认证服务实例
func NewAuthService(logger *zap.Logger, configRepo *repository.PVEConfigRepo, sessionRepo *repository.SessionRepo, auditRepo *repository.AuditLogRepo) *AuthService {
	return &AuthService{
		logger:      logger,
		configRepo:  configRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
	}
}

// Login 执行 PVE 登录流程
// 验证凭据，生成 JWT，持久化会话，记录审计日志
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	if req.Realm == "" {
		req.Realm = "pam"
	}

	// 创建 PVE 客户端并验证凭据
	baseURL := fmt.Sprintf("https://%s:%d/api2/json", req.Host, req.Port)
	pveCfg := config.PVEConfig{BaseURL: baseURL, VerifyTLS: false}
	client, err := pve.NewClient(pveCfg, s.logger)
	if err != nil {
		s.logger.Error("创建 PVE 客户端失败", zap.Error(err))
		s.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "创建 PVE 客户端失败", clientIP, "failed")
		return nil, fmt.Errorf("创建 PVE 客户端失败")
	}

	loginCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = client.Login(loginCtx, req.Username, req.Password, req.Realm)
	if err != nil {
		fullUsername := fmt.Sprintf("%s@%s", req.Username, req.Realm)
		s.logger.Warn("PVE 登录失败", zap.String("host", req.Host), zap.String("username", fullUsername), zap.Error(err))
		s.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), fmt.Sprintf("登录失败: %s", err.Error()), clientIP, "failed")
		return nil, fmt.Errorf("认证失败：用户 %q 的密码不正确，请检查用户名和密码", fullUsername)
	}

	// 加密密码并生成 JWT
	encryptedPwd, err := EncryptPassword(req.Password)
	if err != nil {
		s.logger.Error("加密密码失败", zap.Error(err))
		s.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "加密密码失败", clientIP, "failed")
		return nil, fmt.Errorf("服务器内部错误")
	}

	token, expiresIn, err := GenerateJWT(req.Username, req.Host, req.Port, req.Realm, encryptedPwd)
	if err != nil {
		s.logger.Error("生成 JWT token 失败", zap.Error(err))
		s.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "生成 JWT 失败", clientIP, "failed")
		return nil, fmt.Errorf("服务器内部错误")
	}

	// 持久化 PVE 配置
	s.savePVEConfig(req, encryptedPwd)

	// 持久化用户会话
	s.saveUserSession(req, token, expiresIn, clientIP, userAgent)

	// 记录审计日志
	s.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "登录成功", clientIP, "success")

	s.logger.Info("用户登录成功", zap.String("host", req.Host), zap.String("username", req.Username), zap.String("ip", clientIP))

	return &LoginResponse{Token: token, ExpiresIn: expiresIn}, nil
}

// GetPVEContext 从 JWT token 中提取 PVE 连接上下文
func (s *AuthService) GetPVEContext(tokenString string) (*PVEContext, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("解析 JWT 失败: %w", err)
	}
	password, err := DecryptPassword(claims.Creds)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败: %w", err)
	}
	return &PVEContext{
		Host:     claims.Host,
		Port:     claims.Port,
		Username: claims.Username,
		Password: password,
		Realm:    claims.Realm,
	}, nil
}

// BuildPVEClientFromToken 从 JWT token 构建已认证的 PVE 客户端
func (s *AuthService) BuildPVEClientFromToken(tokenString string) (*pve.Client, error) {
	pveCtx, err := s.GetPVEContext(tokenString)
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("https://%s:%d/api2/json", pveCtx.Host, pveCtx.Port)
	cfg := config.PVEConfig{BaseURL: baseURL, VerifyTLS: false}
	client, err := pve.NewClient(cfg, s.logger)
	if err != nil {
		return nil, fmt.Errorf("创建 PVE 客户端失败: %w", err)
	}
	loginCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.Login(loginCtx, pveCtx.Username, pveCtx.Password, pveCtx.Realm)
	if err != nil {
		return nil, fmt.Errorf("PVE 登录失败: %w", err)
	}
	return client, nil
}

// GetPVEConfigs 获取所有保存的 PVE 配置
func (s *AuthService) GetPVEConfigs() ([]model.PVEConfig, error) {
	if s.configRepo == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return s.configRepo.GetAll()
}

// GetAuditLogs 获取审计日志（支持分页和过滤）
func (s *AuthService) GetAuditLogs(page, pageSize int, userID, action string) ([]model.AuditLog, int64, error) {
	if s.auditRepo == nil {
		return nil, 0, fmt.Errorf("数据库未初始化")
	}
	return s.auditRepo.List(page, pageSize, userID, action)
}

// --- 内部辅助方法 ---

func (s *AuthService) savePVEConfig(req *LoginRequest, encryptedPwd string) {
	if s.configRepo == nil {
		return
	}
	cfg, err := s.configRepo.GetByHost(req.Host, req.Port)
	if err != nil {
		newCfg := &model.PVEConfig{
			Host:      req.Host,
			Port:      req.Port,
			Realm:     req.Realm,
			Username:  req.Username,
			Password:  encryptedPwd,
			Name:      fmt.Sprintf("%s:%d", req.Host, req.Port),
			IsDefault: false,
			VerifyTLS: false,
		}
		if err := s.configRepo.Create(newCfg); err != nil {
			s.logger.Warn("保存 PVE 配置到数据库失败", zap.Error(err))
		}
	} else {
		cfg.Username = req.Username
		cfg.Password = encryptedPwd
		cfg.Realm = req.Realm
		cfg.UpdatedAt = time.Now()
		if err := s.configRepo.Update(cfg); err != nil {
			s.logger.Warn("更新 PVE 配置到数据库失败", zap.Error(err))
		}
	}
}

func (s *AuthService) saveUserSession(req *LoginRequest, token string, expiresIn int64, clientIP, userAgent string) {
	if s.sessionRepo == nil {
		return
	}
	session := &model.UserSession{
		UserID:     fmt.Sprintf("%s@%s:%d", req.Username, req.Host, req.Port),
		Username:   req.Username,
		Host:       req.Host,
		Port:       req.Port,
		Token:      token,
		IP:         clientIP,
		UserAgent:  userAgent,
		ExpiresAt:  time.Now().Add(time.Duration(expiresIn) * time.Second),
		LastActive: time.Now(),
	}
	if err := s.sessionRepo.Create(session); err != nil {
		s.logger.Warn("保存用户会话到数据库失败", zap.Error(err))
	}
}

func (s *AuthService) recordAuditLog(username, action, resource, detail, ip, status string) {
	if s.auditRepo == nil {
		return
	}
	log := &model.AuditLog{
		UserID:   username,
		Username: username,
		Action:   action,
		Resource: resource,
		Detail:   detail,
		IP:       ip,
		Status:   status,
	}
	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Warn("记录审计日志失败", zap.Error(err))
	}
}
