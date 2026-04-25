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
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
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
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{logger: logger}
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
		c.JSON(500, gin.H{"code": 500, "message": "创建 PVE 客户端失败"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Login(ctx, req.Username, req.Password, req.Realm)
	if err != nil {
		h.logger.Warn("PVE 登录失败", zap.String("host", req.Host), zap.String("username", req.Username), zap.String("ip", c.ClientIP()), zap.Error(err))
		c.JSON(401, gin.H{"code": 401, "message": fmt.Sprintf("PVE 认证失败: %s", err.Error())})
		return
	}

	encryptedPwd, err := encryptPassword(req.Password)
	if err != nil {
		h.logger.Error("加密密码失败", zap.Error(err))
		c.JSON(500, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	token, expiresIn, err := generateJWT(req.Username, req.Host, req.Port, req.Realm, encryptedPwd)
	if err != nil {
		h.logger.Error("生成 JWT token 失败", zap.Error(err))
		c.JSON(500, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	h.logger.Info("用户登录成功", zap.String("host", req.Host), zap.String("username", req.Username), zap.String("ip", c.ClientIP()))
	c.JSON(200, gin.H{"code": 0, "message": "登录成功", "data": LoginResponse{Token: token, ExpiresIn: expiresIn}})
}

func generateJWT(username, host string, port int, realm, creds string) (string, int64, error) {
	expiresIn := int64(24 * 3600)
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
