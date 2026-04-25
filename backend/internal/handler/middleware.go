package handler

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"go.uber.org/zap"
)

// JWTAuthMiddleware JWT 认证中间件
// 验证请求中的 Authorization header，解析 JWT token，将完整 Claims 存入上下文
func JWTAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 提取 Bearer token
		tokenString := authHeader[7:] // 去掉 "Bearer " 前缀

		claims, err := ParseJWT(tokenString)
		if err != nil {
			logger.Warn("JWT 验证失败",
				zap.String("ip", c.ClientIP()),
				zap.Error(err),
			)
			c.JSON(401, gin.H{
				"code":    401,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将完整 Claims 存入上下文，后续 handler 可从中提取 PVE 连接信息
		c.Set("claims", claims)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// RequestLogMiddleware 请求日志中间件
// 记录每个请求的方法、路径、IP 和响应时间
func RequestLogMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("请求完成",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}

// CORSMiddleware CORS 跨域中间件
// 根据配置允许跨域请求，支持预检请求处理
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(corsConfig)
}
