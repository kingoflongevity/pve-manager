package pve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
)

// ListStorage 获取指定节点的存储列表
// node: 节点名称
// 返回该节点上所有存储的信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListStorage(ctx context.Context, node string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/storage", node)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取存储列表失败: %w", err)
	}
	return result, nil
}

// GetStorageStatus 获取指定存储的状态
// node: 节点名称, storage: 存储名称
// 返回存储的详细状态信息（总容量、已用空间等）
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetStorageStatus(ctx context.Context, node, storage string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/storage/%s/status", node, storage)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取存储状态失败: %w", err)
	}
	return result, nil
}

// GetStorageContent 获取存储内容列表
// node: 节点名称, storage: 存储名称
// 返回存储中的所有文件（ISO、VZDump、容器模板等）
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetStorageContent(ctx context.Context, node, storage string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/storage/%s/content", node, storage)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取存储内容失败: %w", err)
	}
	return result, nil
}

// GetStorageContentByType 按类型获取存储内容
// node: 节点名称, storage: 存储名称, contentType: 内容类型 (images, iso, vztmpl, backup, snippets)
// 返回指定类型的内容列表
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetStorageContentByType(ctx context.Context, node, storage, contentType string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/storage/%s/content", node, storage)
	params := url.Values{}
	if contentType != "" {
		params.Set("content", contentType)
	}
	if err := c.GetWithParams(ctx, path, params, &result); err != nil {
		return nil, fmt.Errorf("获取存储内容失败: %w", err)
	}
	return result, nil
}

// CreateStorage 创建存储
// node: 节点名称, params: 存储创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateStorage(ctx context.Context, node string, params StorageCreateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage", node)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建存储失败: %w", err)
	}
	return upid, nil
}

// UpdateStorage 更新存储配置
// node: 节点名称, storage: 存储名称, params: 更新参数
// 返回异步任务 ID (UPID)
func (c *Client) UpdateStorage(ctx context.Context, node, storage string, params StorageUpdateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage/%s", node, storage)
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新存储失败: %w", err)
	}
	return upid, nil
}

// DeleteStorage 删除存储
// node: 节点名称, storage: 存储名称
// 返回异步任务 ID (UPID)
func (c *Client) DeleteStorage(ctx context.Context, node, storage string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage/%s", node, storage)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除存储失败: %w", err)
	}
	return upid, nil
}

// DownloadISO 从 URL 下载 ISO 到存储
// node: 节点名称, storage: 存储名称, url: 下载 URL, filename: 文件名
// 返回异步任务 ID (UPID)
func (c *Client) DownloadISO(ctx context.Context, node, storage, fileURL, filename string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage/%s/download-url", node, storage)
	params := map[string]interface{}{
		"url":      fileURL,
		"filename": filename,
	}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("下载 ISO 失败: %w", err)
	}
	return upid, nil
}

// UploadStorageContent 上传文件到存储
// node: 节点名称, storage: 存储名称, contentType: 内容类型, filename: 文件名
// 注意：此方法需要 multipart/form-data 请求，前端应直接上传到 PVE
func (c *Client) UploadStorageContent(ctx context.Context, node, storage, contentType, filename string) error {
	// 返回上传 URL 信息供前端使用
	// 实际上传应由前端直接发送到 PVE 的 /nodes/{node}/storage/{storage}/upload 端点
	return nil
}

// UploadFileToStorage 通过后端代理上传文件到 PVE 存储
// node: 节点名称, storage: 存储名称, contentType: 内容类型
// fileData: 文件数据, filename: 文件名
// 返回异步任务 ID (UPID)
func (c *Client) UploadFileToStorage(ctx context.Context, node, storage, contentType, filename string, fileData []byte) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	// 添加 content 字段
	contentField, err := writer.CreateFormField("content")
	if err != nil {
		return "", fmt.Errorf("创建 content 字段失败: %w", err)
	}
	if _, err := contentField.Write([]byte(contentType)); err != nil {
		return "", fmt.Errorf("写入 content 值失败: %w", err)
	}
	
	// 添加文件字段
	part, err := writer.CreateFormFile("filename", filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭 multipart writer 失败: %w", err)
	}
	
	// 直接执行 HTTP 请求
	path := fmt.Sprintf("nodes/%s/storage/%s/upload", node, storage)
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, path)
	
	req, err := c.newRequest(ctx, "POST", reqURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("上传失败 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	
	// 解析 UPID
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	
	if upid, ok := apiResp.Data.(string); ok {
		return upid, nil
	}
	
	return "", fmt.Errorf("上传响应数据格式错误")
}

// AllocateDiskImage 分配磁盘镜像
// node: 节点名称, storage: 存储名称, params: 分配参数
// 返回异步任务 ID (UPID)
func (c *Client) AllocateDiskImage(ctx context.Context, node, storage string, params map[string]interface{}) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage/%s/allocate", node, storage)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("分配磁盘镜像失败: %w", err)
	}
	return upid, nil
}

// DeleteStorageContent 删除存储中的文件
// node: 节点名称, storage: 存储名称, volume: 卷 ID (如: local:iso/ubuntu-22.04.iso)
// 返回异步任务 ID (UPID)
func (c *Client) DeleteStorageContent(ctx context.Context, node, storage, volume string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/storage/%s/content/%s", node, storage, url.PathEscape(volume))
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除存储文件失败: %w", err)
	}
	return upid, nil
}
