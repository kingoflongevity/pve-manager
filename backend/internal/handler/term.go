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

// TermProxyHandler WebSocket 终端代理
// 代理浏览器与 PVE termproxy WebSocket 之间的双向通信
// LXC 容器使用 termproxy（而非 vncproxy）提供终端访问
type TermProxyHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewTermProxyHandler 创建终端代理处理器
func NewTermProxyHandler(authService *service.AuthService, logger *zap.Logger) *TermProxyHandler {
	return &TermProxyHandler{
		authService: authService,
		logger:      logger,
	}
}

// HandleTermProxy 处理 WebSocket 终端连接
// 流程：
// 1. 从 query 参数获取节点名、VM ID、终端端口、票据和 PVEAuthCookie
// 2. 使用 PVEAuthCookie 和 Referer header（包含 xtermjs=1）连接 PVE termproxy WebSocket
// 3. 启动双向代理（浏览器 ↔ PVE）
//
// 重要：
// - PVEAuthCookie 必须与终端票据来自同一个 PVE 会话
// - termproxy 需要在 Referer header 中包含 xtermjs=1 参数
// - PVE xterm.js 使用 'binary' 子协议和 ArrayBuffer 传输数据
// - 连接建立后需要发送认证信息：username:ticket\n
func (h *TermProxyHandler) HandleTermProxy(c *gin.Context) {
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

	wsClosed := false
	closeWS := func() {
		if !wsClosed {
			wsClosed = true
			wsConn.Close()
		}
	}
	defer closeWS()

	termPort := c.Query("termport")
	termTicket := c.Query("termticket")
	node := c.Query("node")
	vmid := c.Query("vmid")
	pveAuthCookie := c.Query("pveauthcookie")
	tokenString := c.Query("token")

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

	if termPort == "" || termTicket == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少终端连接参数: termport, termticket"))
		closeWS()
		return
	}

	if node == "" || vmid == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少终端连接参数: node, vmid"))
		closeWS()
		return
	}

	if pveAuthCookie == "" {
		wsConn.WriteMessage(websocket.TextMessage, []byte("缺少 PVEAuthCookie"))
		closeWS()
		return
	}

	pveCtx, err := h.authService.GetPVEContext(tokenString)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte("获取 PVE 连接信息失败: "+err.Error()))
		closeWS()
		return
	}

	// PVE termproxy WebSocket 路径与 vncwebsocket 相同
	// 格式: /api2/json/nodes/{node}/lxc/{vmid}/vncwebsocket?port={port}&vncticket={ticket}
	targetURL := fmt.Sprintf("wss://%s:%d/api2/json/nodes/%s/lxc/%s/vncwebsocket?port=%s&vncticket=%s",
		pveCtx.Host, pveCtx.Port, node, vmid, termPort, url.QueryEscape(termTicket))

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Subprotocols: []string{"binary"},
	}
	header := http.Header{}
	header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pveAuthCookie))
	// termproxy 需要在 Referer 中包含 xtermjs=1 参数和 console 类型
	header.Set("Referer", fmt.Sprintf("https://%s:%d/?console=lxc&xtermjs=1&vmid=%s&node=%s",
		pveCtx.Host, pveCtx.Port, vmid, node))

	pveConn, _, err := dialer.Dial(targetURL, header)
	if err != nil {
		headerStrs := []string{}
		for k, v := range header {
			headerStrs = append(headerStrs, fmt.Sprintf("%s: %s", k, v))
		}

		h.logger.Error("TermProxy: 连接 PVE termproxy WebSocket 失败",
			zap.String("URL", targetURL),
			zap.Strings("Headers", headerStrs),
			zap.Error(err))

		if err == websocket.ErrBadHandshake {
			h.logger.Error("WebSocket 握手失败，可能是 PVEAuthCookie 与票据不匹配")
		}

		wsConn.WriteMessage(websocket.TextMessage, []byte("连接 PVE 终端失败: "+err.Error()))
		closeWS()
		return
	}
	defer pveConn.Close()

	done := make(chan struct{}, 2)
	go proxyTermPVEToWS(wsConn, pveConn, done, h.logger)
	go proxyTermWSToPVE(wsConn, pveConn, done, h.logger)
	<-done
}

// proxyTermPVEToWS 将 PVE 终端数据转发到浏览器 WebSocket
func proxyTermPVEToWS(wsConn *websocket.Conn, pveConn *websocket.Conn, done chan struct{}, logger *zap.Logger) {
	defer func() { done <- struct{}{} }()
	for {
		msgType, msg, err := pveConn.ReadMessage()
		if err != nil {
			logger.Debug("TermProxy: PVE→Browser read error", zap.Error(err))
			return
		}
		if err := wsConn.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}

// proxyTermWSToPVE 将浏览器 WebSocket 数据转发到 PVE 终端
func proxyTermWSToPVE(wsConn *websocket.Conn, pveConn *websocket.Conn, done chan struct{}, logger *zap.Logger) {
	defer func() { done <- struct{}{} }()
	for {
		msgType, msg, err := wsConn.ReadMessage()
		if err != nil {
			logger.Debug("TermProxy: Browser→PVE read error", zap.Error(err))
			return
		}
		if err := pveConn.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}
