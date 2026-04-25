package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kingoflongevity/pve-manager/backend/internal/pve"
	"go.uber.org/zap"
)

// VNCHandler VNC WebSocket 代理处理器
// 负责将前端的 WebSocket 连接代理到 PVE 的 vncwebsocket 端点
// 实现双向 WebSocket 流量转发
type VNCHandler struct {
	pveClient *pve.Client
	logger    *zap.Logger
	upgrader  websocket.Upgrader
}

// NewVNCHandler 创建 VNC WebSocket 代理处理器
// upgrader 配置允许跨域 WebSocket 连接
func NewVNCHandler(pveClient *pve.Client, logger *zap.Logger) *VNCHandler {
	return &VNCHandler{
		pveClient: pveClient,
		logger:    logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			// 允许所有来源的 WebSocket 连接（已通过 JWT 中间件鉴权）
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// vncTicketResponse VNC 代理票据响应结构
type vncTicketResponse struct {
	Port   int    `json:"port"`
	Ticket string `json:"ticket"`
	Cert   string `json:"cert,omitempty"`
	UPID   string `json:"upid"`
}

/**
 * ProxyVNCWebSocket VNC WebSocket 代理处理函数
 *
 * 流程：
 * 1. 升级 HTTP 连接为 WebSocket
 * 2. 从 PVE 获取 VNC 代理票据（port + ticket）
 * 3. 连接到 PVE 的 vncwebsocket 端点
 * 4. 双向转发 WebSocket 流量
 *
 * 路由: GET /api/ws/vnc/:node/:vmid/:vmType
 */
func (h *VNCHandler) ProxyVNCWebSocket(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	vmType := c.Param("vmType")

	if node == "" || vmid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必要的路径参数",
		})
		return
	}

	// 验证 vmType
	if vmType != "qemu" && vmType != "lxc" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的虚拟机类型，仅支持 qemu 或 lxc",
		})
		return
	}

	// 检查认证状态
	if !h.pveClient.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "PVE 客户端未认证",
		})
		return
	}

	// 升级 HTTP 连接为 WebSocket
	clientConn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer clientConn.Close()

	// 获取 VNC 代理票据
	ticketResp, err := h.getVNCTicket(node, vmid, vmType)
	if err != nil {
		h.logger.Error("获取 VNC 票据失败",
			zap.String("node", node),
			zap.String("vmid", vmid),
			zap.String("vmType", vmType),
			zap.Error(err),
		)
		h.sendErrorMessage(clientConn, "获取 VNC 票据失败: "+err.Error())
		return
	}

	// 连接到 PVE 的 vncwebsocket 端点
	pveConn, err := h.connectToPVEVNC(node, vmid, vmType, ticketResp)
	if err != nil {
		h.logger.Error("连接 PVE VNC 失败",
			zap.String("node", node),
			zap.String("vmid", vmid),
			zap.Error(err),
		)
		h.sendErrorMessage(clientConn, "连接 PVE VNC 失败: "+err.Error())
		return
	}
	defer pveConn.Close()

	h.logger.Info("VNC WebSocket 代理建立成功",
		zap.String("node", node),
		zap.String("vmid", vmid),
		zap.String("vmType", vmType),
	)

	// 双向转发 WebSocket 流量
	h.bidirectionalForward(clientConn, pveConn)
}

/**
 * getVNCTicket 调用 PVE API 获取 VNC 代理票据
 *
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param vmType 虚拟机类型 (qemu | lxc)
 * @return 包含 port, ticket, upid 的票据信息
 */
func (h *VNCHandler) getVNCTicket(node, vmid, vmType string) (*vncTicketResponse, error) {
	// 构建 vncproxy API 路径
	apiPath := fmt.Sprintf("nodes/%s/%s/%s/vncproxy", node, vmType, vmid)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 PVE client 的 Do 方法获取票据
	resp, err := h.pveClient.Do(ctx, "POST", apiPath, map[string]interface{}{
		"websocket": 1,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 vncproxy 失败: %w", err)
	}

	// 解析响应数据
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("vncproxy 响应数据格式错误")
	}

	var ticketResp vncTicketResponse
	dataJSON, _ := json.Marshal(data)
	if err := json.Unmarshal(dataJSON, &ticketResp); err != nil {
		return nil, fmt.Errorf("解析 vncproxy 响应失败: %w", err)
	}

	if ticketResp.Port == 0 || ticketResp.Ticket == "" {
		return nil, fmt.Errorf("无效的 VNC 票据：缺少 port 或 ticket")
	}

	return &ticketResp, nil
}

/**
 * connectToPVEVNC 连接到 PVE 的 vncwebsocket 端点
 *
 * PVE vncwebsocket 协议：
 * 1. 首先发送一个 JSON 握手消息包含 ticket
 * 2. 等待 PVE 返回 "OK" 响应
 * 3. 之后开始转发 VNC 帧数据
 *
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @param vmType 虚拟机类型
 * @param ticketResp VNC 票据信息
 * @return WebSocket 连接
 */
func (h *VNCHandler) connectToPVEVNC(node, vmid, vmType string, ticketResp *vncTicketResponse) (*websocket.Conn, error) {
	// 获取 PVE 基础 URL 并解析
	baseURL := h.pveClient.GetBaseURL()
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 baseURL 失败: %w", err)
	}

	// 构建 WebSocket URL
	// PVE vncwebsocket 路径: /api2/json/nodes/{node}/{vmType}/{vmid}/vncwebsocket/{port}/{ticket}
	wsScheme := "wss"
	if parsedURL.Scheme == "http" {
		wsScheme = "ws"
	}

	wsPath := fmt.Sprintf("/api2/json/nodes/%s/%s/%s/vncwebsocket/%d/%s",
		node, vmType, vmid, ticketResp.Port, ticketResp.Ticket)

	wsURL := url.URL{
		Scheme: wsScheme,
		Host:   parsedURL.Host,
		Path:   wsPath,
	}

	// 构建 WebSocket 请求头（复用 PVE 认证）
	requestHeader := http.Header{}
	ticket := h.pveClient.GetTicket()
	if ticket != "" {
		if len(ticket) > 12 && ticket[:12] == "PVEAPIToken=" {
			requestHeader.Set("Authorization", ticket)
		} else {
			requestHeader.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", ticket))
		}
	}

	// 连接到 PVE vncwebsocket
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}

	pveConn, _, err := dialer.Dial(wsURL.String(), requestHeader)
	if err != nil {
		return nil, fmt.Errorf("WebSocket 连接 PVE 失败: %w", err)
	}

	// 发送握手消息（PVE vncwebsocket 协议要求）
	handshakeMsg := struct {
		Host string `json:"host"`
	}{
		Host: parsedURL.Hostname(),
	}
	handshakeBytes, err := json.Marshal(handshakeMsg)
	if err != nil {
		pveConn.Close()
		return nil, fmt.Errorf("序列化握手消息失败: %w", err)
	}

	if err := pveConn.WriteMessage(websocket.BinaryMessage, handshakeBytes); err != nil {
		pveConn.Close()
		return nil, fmt.Errorf("发送握手消息失败: %w", err)
	}

	// 等待 PVE 确认
	_, _, err = pveConn.ReadMessage()
	if err != nil {
		pveConn.Close()
		return nil, fmt.Errorf("读取 PVE 握手确认失败: %w", err)
	}

	return pveConn, nil
}

/**
 * bidirectionalForward 双向 WebSocket 流量转发
 *
 * 使用两个 goroutine 分别处理两个方向的数据流：
 * 1. 前端 -> PVE
 * 2. PVE -> 前端
 *
 * 当任意一端断开连接时，另一端也会被关闭
 *
 * @param clientConn 前端 WebSocket 连接
 * @param pveConn PVE WebSocket 连接
 */
func (h *VNCHandler) bidirectionalForward(clientConn, pveConn *websocket.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	errCh := make(chan error, 2)

	// 前端 -> PVE
	go func() {
		defer wg.Done()
		for {
			msgType, message, err := clientConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					h.logger.Debug("前端 WebSocket 读取异常", zap.Error(err))
					select {
					case errCh <- err:
					default:
					}
				}
				return
			}

			if err := pveConn.WriteMessage(msgType, message); err != nil {
				h.logger.Error("写入 PVE WebSocket 失败", zap.Error(err))
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}()

	// PVE -> 前端
	go func() {
		defer wg.Done()
		for {
			msgType, message, err := pveConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					h.logger.Debug("PVE WebSocket 读取异常", zap.Error(err))
					select {
					case errCh <- err:
					default:
					}
				}
				return
			}

			if err := clientConn.WriteMessage(msgType, message); err != nil {
				h.logger.Error("写入前端 WebSocket 失败", zap.Error(err))
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}()

	// 等待转发完成
	wg.Wait()

	// 关闭错误通道
	close(errCh)
}

/**
 * sendErrorMessage 向前端发送错误消息
 */
func (h *VNCHandler) sendErrorMessage(conn *websocket.Conn, message string) {
	errMsg := map[string]string{
		"error": message,
	}
	data, _ := json.Marshal(errMsg)
	_ = conn.WriteMessage(websocket.TextMessage, data)
}

// VNCProxyTicket 获取 VNC 代理票据（HTTP API，非 WebSocket）
// GET /api/pve/nodes/:node/:vmType/:vmid/vnc-ticket
// 用于前端在连接 WebSocket 前获取票据信息
func (h *VNCHandler) VNCProxyTicket(c *gin.Context) {
	node := c.Param("node")
	vmidStr := c.Param("vmid")
	vmType := c.Param("vmType")

	if node == "" || vmidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必要的路径参数",
		})
		return
	}

	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "虚拟机 ID 格式错误",
		})
		return
	}

	if vmType != "qemu" && vmType != "lxc" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的虚拟机类型，仅支持 qemu 或 lxc",
		})
		return
	}

	// 检查认证状态
	if !h.pveClient.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "PVE 客户端未认证",
		})
		return
	}

	// 获取 VNC 票据
	ticketResp, err := h.getVNCTicket(node, vmidStr, vmType)
	if err != nil {
		h.logger.Error("获取 VNC 票据失败",
			zap.String("node", node),
			zap.Int("vmid", vmid),
			zap.String("vmType", vmType),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取 VNC 票据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"port":   ticketResp.Port,
			"ticket": ticketResp.Ticket,
			"cert":   ticketResp.Cert,
			"upid":   ticketResp.UPID,
		},
	})
}
