package pve

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/config"
	"go.uber.org/zap"
)

// Client PVE API 客户端
// 负责与 Proxmox VE API 进行通信，处理认证、请求重试和 ticket 自动刷新
type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	ticket     string
	csrfToken  string
	logger     *zap.Logger

	// 登录凭据（用于自动刷新 ticket）
	username string
	password string
	realm    string

	// 登录相关配置
	loginTimeout    time.Duration // 登录超时时间
	ticketExpiry    time.Duration // ticket 过期时间
	autoRefresh     bool          // 是否自动刷新 ticket
	refreshInterval time.Duration // 刷新间隔
}

// NewClient 创建新的 PVE API 客户端
// 根据配置初始化 HTTP 客户端、连接池、超时设置和重试策略
func NewClient(cfg config.PVEConfig, logger *zap.Logger) (*Client, error) {
	transport := &http.Transport{
		// 连接池配置
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,

		// TLS 配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifyTLS, //nolint:gosec // 用户可配置是否验证 TLS
		},

		// 超时配置
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 移除末尾斜杠
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		logger:          logger,
		loginTimeout:    30 * time.Second,
		ticketExpiry:    2 * time.Hour, // PVE ticket 默认 2 小时过期
		autoRefresh:     true,
		refreshInterval: 1 * time.Hour, // 每小时刷新一次
	}, nil
}

// Login 向 PVE 进行身份认证
// 使用用户名密码获取 ticket 和 CSRF token
// 返回认证响应和错误信息
func (c *Client) Login(ctx context.Context, username, password, realm string) (*TicketResponse, error) {
	c.username = username
	c.password = password
	c.realm = realm

	// 构建登录请求体
	formData := url.Values{}
	formData.Set("username", username)
	formData.Set("password", password)

	reqURL := fmt.Sprintf("%s/access/ticket", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建登录请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("登录失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w", err)
	}

	// 解析 ticket 数据
	ticketData, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("登录响应数据格式错误")
	}

	// 转换为结构化数据
	ticketJSON, _ := json.Marshal(ticketData)
	var ticket TicketResponse
	if err := json.Unmarshal(ticketJSON, &ticket); err != nil {
		return nil, fmt.Errorf("解析 ticket 失败: %w", err)
	}

	// 保存认证信息用于后续请求
	c.mu.Lock()
	c.ticket = ticket.Ticket
	c.csrfToken = ticket.CSRFToken
	c.mu.Unlock()

	// 启动后台 ticket 刷新协程
	if c.autoRefresh {
		go c.autoRefreshTicket()
	}

	return &ticket, nil
}

// LoginWithToken 使用 API Token 进行认证
// tokenID 格式: USER@REALM!TOKENID
// tokenSecret: 令牌密钥
func (c *Client) LoginWithToken(ctx context.Context, tokenID, tokenSecret string) error {
	// 使用 Token 认证时，不需要 ticket，直接在请求头中添加 Authorization
	c.mu.Lock()
	c.ticket = "" // 清空 ticket
	c.csrfToken = ""
	c.mu.Unlock()

	// 保存 Token 用于后续请求
	c.mu.Lock()
	defer c.mu.Unlock()

	// 验证 Token 格式
	reqURL := fmt.Sprintf("%s/access/token", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("创建 Token 验证请求失败: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s:%s", tokenID, tokenSecret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Token 验证请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Token 无效 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Token 认证成功，保存 Token
	c.ticket = fmt.Sprintf("PVEAPIToken=%s:%s", tokenID, tokenSecret)
	return nil
}

// SetTicket 手动设置 ticket 和 CSRF token
// 用于外部认证或 token 刷新场景
func (c *Client) SetTicket(ticket, csrfToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ticket = ticket
	c.csrfToken = csrfToken
}

// GetTicket 获取当前 ticket
func (c *Client) GetTicket() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ticket
}

// GetCSRFToken 获取当前 CSRF token
func (c *Client) GetCSRFToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.csrfToken
}

// GetBaseURL 获取 PVE 基础 URL
// 用于构建 WebSocket 连接地址
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// autoRefreshTicket 后台自动刷新 ticket
// 在 ticket 过期前定期重新登录以维持认证状态
func (c *Client) autoRefreshTicket() {
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		if c.username == "" || c.password == "" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), c.loginTimeout)
		c.logger.Info("正在自动刷新 PVE ticket")
		_, err := c.Login(ctx, c.username, c.password, c.realm)
		cancel()

		if err != nil {
			c.logger.Error("自动刷新 PVE ticket 失败", zap.Error(err))
		} else {
			c.logger.Info("PVE ticket 自动刷新成功")
		}
	}
}

// IsAuthenticated 检查客户端是否已认证
func (c *Client) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ticket != ""
}

// Do 执行 HTTP 请求到 PVE API
// 自动附加认证 header，处理响应解析
// method: HTTP 方法 (GET, POST, PUT, DELETE)
// path: API 路径 (不含 baseURL)
// params: 请求参数 (GET 时为 query params, POST/PUT 时为 JSON body)
func (c *Client) Do(ctx context.Context, method, path string, params map[string]interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	var queryURL string

	// 处理 GET 请求的 query 参数
	if method == http.MethodGet && len(params) > 0 {
		queryParams := url.Values{}
		for k, v := range params {
			queryParams.Set(k, fmt.Sprintf("%v", v))
		}
		queryURL = fmt.Sprintf("%s/%s?%s", c.baseURL, strings.TrimPrefix(path, "/"), queryParams.Encode())
	} else {
		queryURL = fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(path, "/"))
	}

	// 处理 POST/PUT 请求的 body
	if (method == http.MethodPost || method == http.MethodPut) && params != nil {
		// 检查是否是表单数据
		if contentType, ok := params["contentType"].(string); ok && contentType == "form" {
			formData := url.Values{}
			for k, v := range params {
				if k == "contentType" {
					continue
				}
				formData.Set(k, fmt.Sprintf("%v", v))
			}
			reqBody = bytes.NewBufferString(formData.Encode())
		} else {
			// 默认使用 JSON
			jsonBody, err := json.Marshal(params)
			if err != nil {
				return nil, fmt.Errorf("序列化请求体失败: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, queryURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证 headers
	c.mu.RLock()
	if c.ticket != "" {
		// 判断是 API Token 还是普通 ticket
		if strings.HasPrefix(c.ticket, "PVEAPIToken=") {
			req.Header.Set("Authorization", c.ticket)
		} else {
			req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.ticket))
		}
	}
	// POST/PUT/DELETE 需要 CSRF token
	if c.csrfToken != "" && (method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete) {
		req.Header.Set("CSRFPreventionToken", c.csrfToken)
	}
	c.mu.RUnlock()

	// 设置 Content-Type（如果不是表单数据）
	if params == nil || (params != nil && params["contentType"] != "form") {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理错误响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("请求失败 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &apiResp, nil
}

// Get 发送 GET 请求到 PVE API
// path 为 API 路径（不含 baseURL）
// result 为响应数据要解析到的目标结构
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// GetWithParams 发送带查询参数的 GET 请求到 PVE API
func (c *Client) GetWithParams(ctx context.Context, path string, params url.Values, result interface{}) error {
	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", strings.TrimPrefix(path, "/"), params.Encode())
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post 发送 POST 请求到 PVE API
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Put 发送 PUT 请求到 PVE API
func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// Delete 发送 DELETE 请求到 PVE API
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, result)
}

// doRequest 执行 HTTP 请求的通用方法
// 自动附加认证 header，处理响应解析
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		// 判断是否是 url.Values
		if formData, ok := body.(url.Values); ok {
			reqBody = bytes.NewBufferString(formData.Encode())
		} else {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("序列化请求体失败: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(path, "/"))
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证 headers
	c.mu.RLock()
	if c.ticket != "" {
		// 判断是 API Token 还是普通 ticket
		if strings.HasPrefix(c.ticket, "PVEAPIToken=") {
			req.Header.Set("Authorization", c.ticket)
		} else {
			req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.ticket))
		}
	}
	// POST/PUT/DELETE 需要 CSRF token
	if c.csrfToken != "" && (method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete) {
		req.Header.Set("CSRFPreventionToken", c.csrfToken)
	}
	c.mu.RUnlock()

	// 设置 Content-Type（如果不是表单数据）
	if _, ok := body.(url.Values); !ok {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}

		// 将 data 字段转换为目标类型
		dataJSON, _ := json.Marshal(apiResp.Data)
		if err := json.Unmarshal(dataJSON, result); err != nil {
			return fmt.Errorf("转换响应数据失败: %w", err)
		}
	}

	return nil
}

// ProxyRequest 代理请求到 PVE API
// 用于处理前端的直接代理请求，返回原始响应
func (c *Client) ProxyRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(path, "/"))
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, 0, fmt.Errorf("创建代理请求失败: %w", err)
	}

	// 复制认证 headers
	c.mu.RLock()
	if c.ticket != "" {
		// 判断是 API Token 还是普通 ticket
		if strings.HasPrefix(c.ticket, "PVEAPIToken=") {
			req.Header.Set("Authorization", c.ticket)
		} else {
			req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.ticket))
		}
	}
	if c.csrfToken != "" {
		req.Header.Set("CSRFPreventionToken", c.csrfToken)
	}
	c.mu.RUnlock()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("代理请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取代理响应失败: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// GetClusterStatus 获取集群状态
// 返回集群基本信息
func (c *Client) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	var status ClusterStatus
	if err := c.Get(ctx, "cluster/status", &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetNodes 获取所有节点列表
// 返回集群中所有节点的信息
func (c *Client) GetNodes(ctx context.Context) ([]NodeInfo, error) {
	var nodes []NodeInfo
	if err := c.Get(ctx, "nodes", &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetVMs 获取指定节点的虚拟机列表
// node 为节点名称
func (c *Client) GetVMs(ctx context.Context, node string) ([]VMInfo, error) {
	var vms []VMInfo
	path := fmt.Sprintf("nodes/%s/qemu", node)
	if err := c.Get(ctx, path, &vms); err != nil {
		return nil, err
	}
	return vms, nil
}

// GetLXCs 获取指定节点的 LXC 容器列表
// node 为节点名称
func (c *Client) GetLXCs(ctx context.Context, node string) ([]LXCInfo, error) {
	var lxcs []LXCInfo
	path := fmt.Sprintf("nodes/%s/lxc", node)
	if err := c.Get(ctx, path, &lxcs); err != nil {
		return nil, err
	}
	return lxcs, nil
}

// GetStorages 获取指定节点的存储列表
// node 为节点名称
func (c *Client) GetStorages(ctx context.Context, node string) ([]StorageInfo, error) {
	var storages []StorageInfo
	path := fmt.Sprintf("nodes/%s/storage", node)
	if err := c.Get(ctx, path, &storages); err != nil {
		return nil, err
	}
	return storages, nil
}

// GetTaskStatus 获取任务状态
// upid 为任务 ID
func (c *Client) GetTaskStatus(ctx context.Context, node, upid string) (*TaskStatus, error) {
	var status TaskStatus
	path := fmt.Sprintf("nodes/%s/tasks/%s/status", node, upid)
	if err := c.Get(ctx, path, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// WaitForTask 等待任务完成
// 轮询任务状态直到完成或超时
func (c *Client) WaitForTask(ctx context.Context, node, upid string, timeout time.Duration) (*TaskStatus, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("等待任务超时: %s", upid)
		case <-ticker.C:
			status, err := c.GetTaskStatus(timeoutCtx, node, upid)
			if err != nil {
				return nil, err
			}
			if status.ExitCode != "" && status.ExitCode != "task is running" {
				return status, nil
			}
		}
	}
}
