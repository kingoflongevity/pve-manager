package handler

import (
	"fmt"
	"net/http"
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

	// 获取 PVE 连接信息
	host := c.Query("host")
	port := c.Query("port")
	vncTicket := c.Query("ticket")
	vnode := c.Query("vncticket")

		// 如果没有传入 host/port，尝试从 JWT 中获取
	if host == "" || port == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			pveCtx, pErr := h.authService.GetPVEContext(tokenString)
			if pErr == nil {
				host = pveCtx.Host
				port = fmt.Sprintf("%d", pveCtx.Port)
			}
		}
	}

	if host == "" || port == "" || vncTicket == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少必要的连接参数: host, port, ticket"))
		return
	}

	// 构建 PVE WebSocket URL
	targetURL := buildPVEWebSocketURL(host, port, vnode, vncTicket)

	// 连接 PVE VNC WebSocket
	dialer := websocket.Dialer{}
	pveConn, _, err := dialer.Dial(targetURL, nil)
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

// buildPVEWebSocketURL 构建 PVE VNC WebSocket URL
func buildPVEWebSocketURL(host, port, vnode, vncTicket string) string {
	if vnode == "" {
		vnode = "1"
	}
	return "wss://" + host + ":" + port + "/api2/json/vncwebsocket?port=5900&vncticket=" + vncTicket
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
