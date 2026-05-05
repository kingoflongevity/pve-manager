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
	"go.uber.org/zap"
)

// VNCProxyHandler WebSocket VNC 代理
// 通过 HTTP 升级建立与 PVE VNC WebSocket 的双向代理
type VNCProxyHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewVNCProxyHandler 创建 VNC 代理处理器
func NewVNCProxyHandler(authService *service.AuthService, logger *zap.Logger) *VNCProxyHandler {
	return &VNCProxyHandler{
		authService: authService,
		logger:      logger,
	}
}

// HandleVNC 处理 WebSocket VNC 连接
// 流程：
// 1. 从 query 参数获取节点名、VM ID、VNC 类型、端口、票据和 PVEAuthCookie
// 2. 使用 PVEAuthCookie 作为认证凭证连接 PVE VNC WebSocket
// 3. 启动双向代理（浏览器 ↔ PVE）
//
// 重要：PVEAuthCookie 必须与 vncticket 来自同一个 PVE 会话，
// 否则 WebSocket 握手会失败（bad handshake）。
// 因此前端在调用 vncproxy API 时会同时获取 PVEAuthCookie 并传递给此处理器。
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

	// 标记连接是否已关闭，避免重复关闭
	wsClosed := false
	closeWS := func() {
		if !wsClosed {
			wsClosed = true
			wsConn.Close()
		}
	}
	defer closeWS()

	// 从 query 参数获取 VNC 连接信息
	vncPort := c.Query("vncport")
	vncTicket := c.Query("vncticket")
	node := c.Query("node")
	vmid := c.Query("vmid")
	vmType := c.Query("vmtype")
	pveAuthCookie := c.Query("pveauthcookie")
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
		closeWS()
		return
	}

	if vncPort == "" || vncTicket == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少 VNC 连接参数: vncport, vncticket"))
		closeWS()
		return
	}

	if node == "" || vmid == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少 VNC 连接参数: node, vmid"))
		closeWS()
		return
	}

	if pveAuthCookie == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少 PVEAuthCookie"))
		closeWS()
		return
	}

	// 获取 PVE 连接信息（主机和端口）
	pveCtx, err := h.authService.GetPVEContext(tokenString)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("获取 PVE 连接信息失败: "+err.Error()))
		closeWS()
		return
	}

	// 构建 PVE VNC WebSocket URL
	// PVE VNC WebSocket 正确路径格式: /api2/json/nodes/{node}/{vmtype}/{vmid}/vncwebsocket?port={port}&vncticket={ticket}
	if vmType == "" {
		vmType = "qemu"
	}
	// vncticket 包含特殊字符（+, /, :, =），必须 URL 编码
	// gorilla/websocket.Dial 不会自动编码 query 参数
	targetURL := fmt.Sprintf("wss://%s:%d/api2/json/nodes/%s/%s/%s/vncwebsocket?port=%s&vncticket=%s",
		pveCtx.Host, pveCtx.Port, node, vmType, vmid, vncPort, url.QueryEscape(vncTicket))

	// 连接 PVE VNC WebSocket
	// 使用前端传入的 PVEAuthCookie（与 vncticket 来自同一个 PVE 会话）
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // PVE 通常使用自签名证书
		},
	}
	header := http.Header{}
	// PVEAuthCookie 是 PVE 登录认证的 ticket，必须与创建 vncticket 时使用的 ticket 一致
	header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pveAuthCookie))

	pveConn, _, err := dialer.Dial(targetURL, header)
	if err != nil {
		// 格式化 headers 用于日志
		headerStrs := []string{}
		for k, v := range header {
			headerStrs = append(headerStrs, fmt.Sprintf("%s: %s", k, v))
		}

		h.logger.Error("VNC Proxy: 连接 PVE VNC WebSocket 失败",
			zap.String("URL", targetURL),
			zap.Strings("Headers", headerStrs),
			zap.Error(err))

		if err == websocket.ErrBadHandshake {
			h.logger.Error("WebSocket 握手失败，可能是 PVEAuthCookie 与 vncticket 不匹配")
		}

		wsConn.WriteMessage(websocket.TextMessage, []byte("连接 PVE VNC 失败: "+err.Error()))
		closeWS()
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
