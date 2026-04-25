package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWT 声明
var jwtSecret = []byte("pve-manager-jwt-secret-change-in-production")

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// Claims JWT 负载声明
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthHandler 认证处理器
// 处理用户登录、token 生成和验证
type AuthHandler struct {
	logger *zap.Logger
}

// NewAuthHandler 创建认证处理器实例
// 接收 logger 用于记录认证相关日志
func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{logger: logger}
}

// Login 处理用户登录请求
// POST /api/auth/login
// 验证用户凭据，生成 JWT token 返回给客户端
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// TODO: 实际项目中应该验证 PVE 凭据
	// 这里仅作 scaffold 演示，简单返回 token

	// 生成 JWT token
	token, expiresIn, err := generateJWT(req.Username)
	if err != nil {
		h.logger.Error("生成 JWT token 失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
		return
	}

	h.logger.Info("用户登录成功",
		zap.String("username", req.Username),
		zap.String("ip", c.ClientIP()),
	)

	c.JSON(200, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": LoginResponse{
			Token:     token,
			ExpiresIn: expiresIn,
		},
	})
}

// generateJWT 生成 JWT token
// 包含用户名和过期时间，返回 token 字符串和过期秒数
func generateJWT(username string) (string, int64, error) {
	expiresIn := int64(24 * 3600) // 24 小时
	now := time.Now()

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "pve-manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn, nil
}

// ParseJWT 解析并验证 JWT token
// 返回 claims 和验证结果
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
