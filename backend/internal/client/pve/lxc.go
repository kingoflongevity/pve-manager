package pve

import (
	"context"
	"fmt"
	"net/url"
)

// ListLXC 获取指定节点的 LXC 容器列表
// node: 节点名称
// 返回该节点上所有 LXC 容器的信息
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListLXC(ctx context.Context, node string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc", node)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取 LXC 列表失败: %w", err)
	}
	return result, nil
}

// GetLXCConfig 获取指定容器的完整配置
// node: 节点名称, vmid: 容器 ID
// 返回容器的详细配置参数
func (c *Client) GetLXCConfig(ctx context.Context, node string, vmid int) (map[string]interface{}, error) {
	var config map[string]interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/config", node, vmid)
	if err := c.Get(ctx, path, &config); err != nil {
		return nil, fmt.Errorf("获取 LXC 配置失败: %w", err)
	}
	return config, nil
}

// CreateLXC 创建新的 LXC 容器
// node: 节点名称, params: 创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateLXC(ctx context.Context, node string, params *LXCCreateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc", node)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建 LXC 失败: %w", err)
	}
	return upid, nil
}

// StartLXC 启动容器
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) StartLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/start", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("启动 LXC 失败: %w", err)
	}
	return upid, nil
}

// StopLXC 停止容器（强制关机）
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) StopLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/stop", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("停止 LXC 失败: %w", err)
	}
	return upid, nil
}

// ShutdownLXC 关闭容器（优雅关机）
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) ShutdownLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/shutdown", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("关闭 LXC 失败: %w", err)
	}
	return upid, nil
}

// RebootLXC 重启容器
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) RebootLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/reboot", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("重启 LXC 失败: %w", err)
	}
	return upid, nil
}

// FreezeLXC 冻结容器（暂停所有进程）
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) FreezeLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/freeze", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("冻结 LXC 失败: %w", err)
	}
	return upid, nil
}

// UnfreezeLXC 解冻容器（恢复所有进程）
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) UnfreezeLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/unfreeze", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("解冻 LXC 失败: %w", err)
	}
	return upid, nil
}

// SetLXCConfig 更新容器配置
// node: 节点名称, vmid: 容器 ID, config: 配置参数映射
// 返回异步任务 ID (UPID)
func (c *Client) SetLXCConfig(ctx context.Context, node string, vmid int, config LXCConfigParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/config", node, vmid)
	if err := c.Put(ctx, path, config, &upid); err != nil {
		return "", fmt.Errorf("更新 LXC 配置失败: %w", err)
	}
	return upid, nil
}

// ListLXCSnapshots 获取容器快照列表
// node: 节点名称, vmid: 容器 ID
// 返回快照信息列表
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListLXCSnapshots(ctx context.Context, node string, vmid int) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/snapshot", node, vmid)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取 LXC 快照列表失败: %w", err)
	}
	return result, nil
}

// CreateLXCSnapshot 创建容器快照
// node: 节点名称, vmid: 容器 ID, name: 快照名称, description: 描述
// 返回异步任务 ID (UPID)
func (c *Client) CreateLXCSnapshot(ctx context.Context, node string, vmid int, name, description string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/snapshot", node, vmid)
	params := url.Values{}
	params.Set("snapname", name)
	if description != "" {
		params.Set("description", description)
	}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建 LXC 快照失败: %w", err)
	}
	return upid, nil
}

// DeleteLXCSnapshot 删除容器快照
// node: 节点名称, vmid: 容器 ID, snapname: 快照名称
// 返回异步任务 ID (UPID)
func (c *Client) DeleteLXCSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/snapshot/%s", node, vmid, snapname)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除 LXC 快照失败: %w", err)
	}
	return upid, nil
}

// CloneLXC 克隆容器
// node: 节点名称, vmid: 容器 ID, params: 克隆参数
// 返回异步任务 ID (UPID)
func (c *Client) CloneLXC(ctx context.Context, node string, vmid int, params *LXCCloneParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/clone", node, vmid)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("克隆 LXC 失败: %w", err)
	}
	return upid, nil
}

// MigrateLXC 迁移容器到其他节点
// node: 节点名称, vmid: 容器 ID, params: 迁移参数
// 返回异步任务 ID (UPID)
func (c *Client) MigrateLXC(ctx context.Context, node string, vmid int, params *LXCMigrateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d/migrate", node, vmid)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("迁移 LXC 失败: %w", err)
	}
	return upid, nil
}

// GetLXCURRD 获取容器 RRD 性能数据
// node: 节点名称, vmid: 容器 ID, timeframe: 时间范围 (hour, day, week, month, year)
// dataset: 数据集 (cpu, memory, network, disk)
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetLXCURRD(ctx context.Context, node string, vmid int, timeframe, dataset string) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/rrd", node, vmid)
	params := url.Values{}
	params.Set("timeframe", timeframe)
	if dataset != "" {
		params.Set("ds", dataset)
	}
	if err := c.GetWithParams(ctx, path, params, &result); err != nil {
		return nil, fmt.Errorf("获取 LXC RRD 数据失败: %w", err)
	}
	return result, nil
}

// DeleteLXC 删除容器
// node: 节点名称, vmid: 容器 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteLXC(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/lxc/%d", node, vmid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除 LXC 失败: %w", err)
	}
	return upid, nil
}

// GetLXCCurrent 获取容器当前状态
// node: 节点名称, vmid: 容器 ID
// 返回容器当前运行状态
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) GetLXCCurrent(ctx context.Context, node string, vmid int) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/status/current", node, vmid)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取 LXC 当前状态失败: %w", err)
	}
	return result, nil
}

// ListLXCPending 获取容器待处理配置
// node: 节点名称, vmid: 容器 ID
// 返回等待重启后生效的配置变更
// 使用 interface{} 返回值防止 PVE 9.x 返回意外数据格式时 json.Unmarshal 失败
func (c *Client) ListLXCPending(ctx context.Context, node string, vmid int) (interface{}, error) {
	var result interface{}
	path := fmt.Sprintf("nodes/%s/lxc/%d/pending", node, vmid)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("获取 LXC 待处理配置失败: %w", err)
	}
	return result, nil
}
