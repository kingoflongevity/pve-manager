package handler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kingoflongevity/pve-manager/backend/internal/service"
)

// VNCProxyHandler WebSocket VNC 代理
// 通过 HTTP 升级建立与 PVE VNC WebSocket 的双向代理
type VNCProxyHandler struct {
	authService *service.AuthService
}

// NewVNCProxyHandler 创建 VNC 代理处理器
func NewVNCProxyHandler(authService *service.AuthService) *VNCProxyHandler {
	return &VNCProxyHandler{
		authService: authService,
	}
}

// HandleVNC 处理 WebSocket VNC 连接
// 流程：
// 1. 从 query 参数获取 JWT token、VNC 端口和票据
// 2. 使用 JWT 构建 PVE 客户端获取认证 cookie
// 3. 连接 PVE VNC WebSocket 端点
// 4. 启动双向代理（浏览器 ↔ PVE）
func (h *VNCProxyHandler) HandleVNC(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "WebSocket 升级失败: " + err.Error()})
		return
	}
	defer wsConn.Close()

	// 从 query 参数获取 VNC 连接信息
	vncPort := c.Query("vncport")
	vncTicket := c.Query("vncticket")
	tokenString := c.Query("token")

	// 尝试从 Authorization header 获取 token（备用）
	if tokenString == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenString == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少认证 token"))
		return
	}

	if vncPort == "" || vncTicket == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少 VNC 连接参数: vncport, vncticket"))
		return
	}

	// 使用 JWT 构建 PVE 客户端（同时完成 PVE 认证）
	pveClient, err := h.authService.BuildPVEClientFromToken(tokenString)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("PVE 认证失败: "+err.Error()))
		return
	}

	// 获取 PVE 连接信息
	pveCtx, err := h.authService.GetPVEContext(tokenString)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("获取 PVE 连接信息失败: "+err.Error()))
		return
	}

	// 构建 PVE VNC WebSocket URL
	targetURL := fmt.Sprintf("wss://%s:%d/api2/json/vncwebsocket?port=%s&vncticket=%s",
		pveCtx.Host, pveCtx.Port, vncPort, url.QueryEscape(vncTicket))

	// 获取 PVE 认证 ticket 用于 cookie
	pveTicket := pveClient.GetTicket()

	// 连接 PVE VNC WebSocket（携带认证 cookie）
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // PVE 通常使用自签名证书
		},
	}
	header := http.Header{}
	if pveTicket != "" {
		header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pveTicket))
	}

	pveConn, _, err := dialer.Dial(targetURL, header)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("连接 PVE VNC 失败: "+err.Error()))
		return
	}
	defer pveConn.Close()

	// 启动双向代理
	done := make(chan struct{}, 2)
	go proxyPVEToWS(wsConn, pveConn, done)
	go proxyWSToPVE(wsConn, pveConn, done)

	// 等待任一连接关闭
	<-done
}

// proxyPVEToWS 将 PVE 数据转发到浏览器 WebSocket
func proxyPVEToWS(wsConn *websocket.Conn, pveConn *websocket.Conn, done chan struct{}) {
	defer func() { done <- struct{}{} }()
	for {
		_, msg, err := pveConn.ReadMessage()
		if err != nil {
			return
		}
		if err := wsConn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			return
		}
	}
}

// proxyWSToPVE 将浏览器 WebSocket 数据转发到 PVE
func proxyWSToPVE(wsConn *websocket.Conn, pveConn *websocket.Conn, done chan struct{}) {
	defer func() { done <- struct{}{} }()
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			return
		}
		if err := pveConn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			return
		}
	}
}
