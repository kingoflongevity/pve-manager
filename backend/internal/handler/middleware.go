package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
	"go.uber.org/zap"
)

var devModeAuth = os.Getenv("PVE_DEV_MODE") == "true"

// AuthMiddleware 认证中间件
// 验证 JWT token 并从 token 中恢复 PVE 连接凭据
// 开发模式下自动跳过认证
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if devModeAuth {
			c.Set("username", "dev_admin")
			c.Set("user_id", "dev_admin")
			c.Set("host", "localhost")
			c.Set("port", 8006)
			c.Set("realm", "pam")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证令牌"})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌格式错误"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		claims, err := service.ParseJWT(tokenString)
		if err != nil || claims == nil {
			logger.Warn("JWT token 无效或已过期", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌已过期，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存入 context，供后续 handler 使用
		c.Set("username", claims.Username)
		c.Set("user_id", claims.Username)
		c.Set("host", claims.Host)
		c.Set("port", claims.Port)
		c.Set("realm", claims.Realm)

		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
