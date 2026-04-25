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
	"sync"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/config"
)

// Client PVE API 客户端
// 负责与 Proxmox VE API 进行通信，处理认证和请求
type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	ticket     string
	csrfToken  string
}

// NewClient 创建新的 PVE API 客户端
// 根据配置初始化 HTTP 客户端和连接参数
func NewClient(cfg config.PVEConfig) (*Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifyTLS, //nolint:gosec // 用户可配置是否验证 TLS
		},
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}, nil
}

// Login 向 PVE 进行身份认证
// 使用用户名密码获取 ticket 和 CSRF token
// 返回认证响应和错误信息
func (c *Client) Login(ctx context.Context, username, password, realm string) (*TicketResponse, error) {
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

	return &ticket, nil
}

// Get 发送 GET 请求到 PVE API
// path 为 API 路径（不含 baseURL）
// result 为响应数据要解析到的目标结构
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
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
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证 headers
	c.mu.RLock()
	if c.ticket != "" {
		req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.ticket))
	}
	if c.csrfToken != "" && (method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete) {
		req.Header.Set("CSRFPreventionToken", c.csrfToken)
	}
	c.mu.RUnlock()

	req.Header.Set("Content-Type", "application/json")
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

// IsAuthenticated 检查客户端是否已认证
func (c *Client) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ticket != ""
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

// ProxyRequest 代理请求到 PVE API
// 用于处理前端的直接代理请求，返回原始响应
func (c *Client) ProxyRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, 0, fmt.Errorf("创建代理请求失败: %w", err)
	}

	// 复制认证 headers
	c.mu.RLock()
	if c.ticket != "" {
		req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", c.ticket))
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
