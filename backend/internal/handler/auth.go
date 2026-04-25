package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"github.com/kingoflongevity/pve-manager/backend/internal/model"
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
	"github.com/kingoflongevity/pve-manager/backend/internal/repository"
	"go.uber.org/zap"
)

var jwtSecret = []byte("pve-manager-jwt-secret-change-in-production")

type LoginRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Realm    string `json:"realm"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

type PVEConfigResponse struct {
	ID        uint   `json:"id"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Realm     string `json:"realm"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

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

type PaginationResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Items interface{} `json:"items"`
}

type Claims struct {
	Username string `json:"username"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Realm    string `json:"realm"`
	Creds    string `json:"creds"`
	jwt.RegisteredClaims
}

type PVEContext struct {
	Host     string
	Port     int
	Username string
	Password string
	Realm    string
}

type AuthHandler struct {
	logger      *zap.Logger
	configRepo  *repository.PVEConfigRepo
	sessionRepo *repository.SessionRepo
	auditRepo   *repository.AuditLogRepo
}

func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{logger: logger}
}

func NewAuthHandlerWithDB(logger *zap.Logger, configRepo *repository.PVEConfigRepo, sessionRepo *repository.SessionRepo, auditRepo *repository.AuditLogRepo) *AuthHandler {
	return &AuthHandler{
		logger:      logger,
		configRepo:  configRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}
	if req.Realm == "" {
		req.Realm = "pam"
	}

	baseURL := fmt.Sprintf("https://%s:%d/api2/json", req.Host, req.Port)
	pveCfg := config.PVEConfig{BaseURL: baseURL, VerifyTLS: false}
	client, err := pve.NewClient(pveCfg, h.logger)
	if err != nil {
		h.logger.Error("创建 PVE 客户端失败", zap.Error(err))
		h.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "创建 PVE 客户端失败", c.ClientIP(), "failed")
		c.JSON(500, gin.H{"code": 500, "message": "创建 PVE 客户端失败"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Login(ctx, req.Username, req.Password, req.Realm)
	if err != nil {
		fullUsername := fmt.Sprintf("%s@%s", req.Username, req.Realm)
		h.logger.Warn("PVE 登录失败", zap.String("host", req.Host), zap.String("username", fullUsername), zap.String("ip", c.ClientIP()), zap.Error(err))
		h.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), fmt.Sprintf("登录失败: %s", err.Error()), c.ClientIP(), "failed")
		c.JSON(401, gin.H{"code": 401, "message": fmt.Sprintf("认证失败：用户 %q 的密码不正确，请检查用户名和密码", fullUsername)})
		return
	}

	encryptedPwd, err := encryptPassword(req.Password)
	if err != nil {
		h.logger.Error("加密密码失败", zap.Error(err))
		h.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "加密密码失败", c.ClientIP(), "failed")
		c.JSON(500, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	token, expiresIn, err := generateJWT(req.Username, req.Host, req.Port, req.Realm, encryptedPwd)
	if err != nil {
		h.logger.Error("生成 JWT token 失败", zap.Error(err))
		h.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "生成 JWT 失败", c.ClientIP(), "failed")
		c.JSON(500, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	// 持久化 PVE 配置到数据库
	if h.configRepo != nil {
		cfg, err := h.configRepo.GetByHost(req.Host, req.Port)
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
			if err := h.configRepo.Create(newCfg); err != nil {
				h.logger.Warn("保存 PVE 配置到数据库失败", zap.Error(err))
			}
		} else {
			cfg.Username = req.Username
			cfg.Password = encryptedPwd
			cfg.Realm = req.Realm
			cfg.UpdatedAt = time.Now()
			if err := h.configRepo.Update(cfg); err != nil {
				h.logger.Warn("更新 PVE 配置到数据库失败", zap.Error(err))
			}
		}
	}

	// 持久化用户会话到数据库
	if h.sessionRepo != nil {
		session := &model.UserSession{
			UserID:     fmt.Sprintf("%s@%s:%d", req.Username, req.Host, req.Port),
			Username:   req.Username,
			Host:       req.Host,
			Port:       req.Port,
			Token:      token,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			ExpiresAt:  time.Now().Add(time.Duration(expiresIn) * time.Second),
			LastActive: time.Now(),
		}
		if err := h.sessionRepo.Create(session); err != nil {
			h.logger.Warn("保存用户会话到数据库失败", zap.Error(err))
		}
	}

	// 记录审计日志
	h.recordAuditLog(req.Username, "login", fmt.Sprintf("%s:%d", req.Host, req.Port), "登录成功", c.ClientIP(), "success")

	h.logger.Info("用户登录成功", zap.String("host", req.Host), zap.String("username", req.Username), zap.String("ip", c.ClientIP()))
	c.JSON(200, gin.H{"code": 0, "message": "登录成功", "data": LoginResponse{Token: token, ExpiresIn: expiresIn}})
}

func (h *AuthHandler) recordAuditLog(username, action, resource, detail, ip, status string) {
	if h.auditRepo == nil {
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
	if err := h.auditRepo.Create(log); err != nil {
		h.logger.Warn("记录审计日志失败", zap.Error(err))
	}
}

func (h *AuthHandler) GetPVEConfigs(c *gin.Context) {
	if h.configRepo == nil {
		c.JSON(500, gin.H{"code": 500, "message": "数据库未初始化"})
		return
	}
	configs, err := h.configRepo.GetAll()
	if err != nil {
		h.logger.Error("获取 PVE 配置列表失败", zap.Error(err))
		c.JSON(500, gin.H{"code": 500, "message": "获取配置列表失败"})
		return
	}
	resp := make([]PVEConfigResponse, 0, len(configs))
	for _, cfg := range configs {
		resp = append(resp, PVEConfigResponse{
			ID: cfg.ID, Host: cfg.Host, Port: cfg.Port, Realm: cfg.Realm,
			Username: cfg.Username, Name: cfg.Name, IsDefault: cfg.IsDefault,
		})
	}
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": resp})
}

func (h *AuthHandler) GetAuditLogs(c *gin.Context) {
	if h.auditRepo == nil {
		c.JSON(500, gin.H{"code": 500, "message": "数据库未初始化"})
		return
	}
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

	logs, total, err := h.auditRepo.List(page, pageSize, userID, action)
	if err != nil {
		h.logger.Error("获取审计日志失败", zap.Error(err))
		c.JSON(500, gin.H{"code": 500, "message": "获取审计日志失败"})
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
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": PaginationResponse{
		Total: total, Page: page, Size: pageSize, Items: resp,
	}})
}

func generateJWT(username, host string, port int, realm, creds string) (string, int64, error) {
	expiresIn := int64(7 * 24 * 3600) // 7天
	now := time.Now()
	claims := Claims{
		Username: username, Host: host, Port: port, Realm: realm, Creds: creds,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiresIn) * time.Second)), IssuedAt: jwt.NewNumericDate(now), Issuer: "pve-manager"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}
	return signedToken, expiresIn, nil
}

func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func extractPVEContext(tokenString string) (*PVEContext, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("解析 JWT 失败: %w", err)
	}
	password, err := decryptPassword(claims.Creds)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败: %w", err)
	}
	return &PVEContext{Host: claims.Host, Port: claims.Port, Username: claims.Username, Password: password, Realm: claims.Realm}, nil
}

func BuildPVEClient(tokenString string, logger *zap.Logger) (*pve.Client, error) {
	pveCtx, err := extractPVEContext(tokenString)
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("https://%s:%d/api2/json", pveCtx.Host, pveCtx.Port)
	cfg := config.PVEConfig{BaseURL: baseURL, VerifyTLS: false}
	client, err := pve.NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("创建 PVE 客户端失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.Login(ctx, pveCtx.Username, pveCtx.Password, pveCtx.Realm)
	if err != nil {
		return nil, fmt.Errorf("PVE 登录失败: %w", err)
	}
	return client, nil
}

var aesKey = []byte("pve-manager-aes-key-32bytes!!")

func encryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptPassword(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文过短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
