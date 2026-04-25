package pve

import (
	"context"
	"fmt"
	"net/url"
)

// GetNodeStatus 获取节点状态
// node: 节点名称
// 返回节点的详细信息（CPU、内存、磁盘、版本等）
func (c *Client) GetNodeStatus(ctx context.Context, node string) (*NodeStatus, error) {
	var status NodeStatus
	path := fmt.Sprintf("nodes/%s/status", node)
	if err := c.Get(ctx, path, &status); err != nil {
		return nil, fmt.Errorf("获取节点状态失败: %w", err)
	}
	return &status, nil
}

// GetNodeVersion 获取节点 PVE 版本信息
// node: 节点名称
// 返回 PVE 版本号、发行版等信息
func (c *Client) GetNodeVersion(ctx context.Context, node string) (*VersionInfo, error) {
	var version VersionInfo
	path := fmt.Sprintf("nodes/%s/version", node)
	if err := c.Get(ctx, path, &version); err != nil {
		return nil, fmt.Errorf("获取节点版本失败: %w", err)
	}
	return &version, nil
}

// GetNodeServices 获取节点服务列表
// node: 节点名称
// 返回节点上所有服务的状态信息
func (c *Client) GetNodeServices(ctx context.Context, node string) ([]Service, error) {
	var services []Service
	path := fmt.Sprintf("nodes/%s/services", node)
	if err := c.Get(ctx, path, &services); err != nil {
		return nil, fmt.Errorf("获取节点服务列表失败: %w", err)
	}
	return services, nil
}

// GetNodeSyslog 获取节点系统日志
// node: 节点名称, limit: 日志行数限制, start: 起始行号
// 返回系统日志条目列表
func (c *Client) GetNodeSyslog(ctx context.Context, node string, limit, start int) ([]LogEntry, error) {
	var logs []LogEntry
	path := fmt.Sprintf("nodes/%s/syslog", node)
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if start > 0 {
		params.Set("start", fmt.Sprintf("%d", start))
	}
	if err := c.GetWithParams(ctx, path, params, &logs); err != nil {
		return nil, fmt.Errorf("获取节点系统日志失败: %w", err)
	}
	return logs, nil
}

// GetNodeTasks 获取节点任务列表
// node: 节点名称
// 返回节点上的任务列表
func (c *Client) GetNodeTasks(ctx context.Context, node string) ([]Task, error) {
	var tasks []Task
	path := fmt.Sprintf("nodes/%s/tasks", node)
	if err := c.Get(ctx, path, &tasks); err != nil {
		return nil, fmt.Errorf("获取节点任务列表失败: %w", err)
	}
	return tasks, nil
}

// GetNodeTaskStatus 获取指定任务状态
// node: 节点名称, upid: 任务 UPID
// 返回任务的当前状态和退出码
func (c *Client) GetNodeTaskStatus(ctx context.Context, node, upid string) (*TaskStatus, error) {
	var status TaskStatus
	path := fmt.Sprintf("nodes/%s/tasks/%s/status", node, upid)
	if err := c.Get(ctx, path, &status); err != nil {
		return nil, fmt.Errorf("获取任务状态失败: %w", err)
	}
	return &status, nil
}

// GetNodeTaskLog 获取指定任务日志
// node: 节点名称, upid: 任务 UPID
// 返回任务的日志行列表
func (c *Client) GetNodeTaskLog(ctx context.Context, node, upid string) ([]TaskLogLine, error) {
	var logs []TaskLogLine
	path := fmt.Sprintf("nodes/%s/tasks/%s/log", node, upid)
	if err := c.Get(ctx, path, &logs); err != nil {
		return nil, fmt.Errorf("获取任务日志失败: %w", err)
	}
	return logs, nil
}

// GetNodeNetwork 获取节点网络接口列表
// node: 节点名称
// 返回所有网络接口的配置信息
func (c *Client) GetNodeNetwork(ctx context.Context, node string) ([]NetInterface, error) {
	var interfaces []NetInterface
	path := fmt.Sprintf("nodes/%s/network", node)
	if err := c.Get(ctx, path, &interfaces); err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %w", err)
	}
	return interfaces, nil
}

// SetNodeNetworkInterface 配置网络接口
// node: 节点名称, iface: 接口名称, config: 接口配置参数
// 返回异步任务 ID (UPID)
func (c *Client) SetNodeNetworkInterface(ctx context.Context, node, iface string, config NetInterfaceConfig) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/network/%s", node, iface)
	if err := c.Put(ctx, path, config, &upid); err != nil {
		return "", fmt.Errorf("配置网络接口失败: %w", err)
	}
	return upid, nil
}

// CreateNodeNetworkInterface 创建网络接口
// node: 节点名称, iface: 接口名称, config: 接口配置参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateNodeNetworkInterface(ctx context.Context, node, iface string, config NetInterfaceConfig) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/network", node)
	params := make(NetInterfaceConfig)
	params["iface"] = iface
	for k, v := range config {
		params[k] = v
	}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建网络接口失败: %w", err)
	}
	return upid, nil
}

// DeleteNodeNetworkInterface 删除网络接口
// node: 节点名称, iface: 接口名称
// 返回异步任务 ID (UPID)
func (c *Client) DeleteNodeNetworkInterface(ctx context.Context, node, iface string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/network/%s", node, iface)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除网络接口失败: %w", err)
	}
	return upid, nil
}

// ApplyNodeNetworkChanges 应用网络配置变更
// node: 节点名称
// 返回异步任务 ID (UPID)
func (c *Client) ApplyNodeNetworkChanges(ctx context.Context, node string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/network", node)
	params := map[string]interface{}{"restart": 1}
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("应用网络变更失败: %w", err)
	}
	return upid, nil
}

// GetNodeAPTUpdate 获取可更新的软件包列表
// node: 节点名称
// 返回待更新的软件包列表
func (c *Client) GetNodeAPTUpdate(ctx context.Context, node string) ([]PackageUpdate, error) {
	var updates []PackageUpdate
	path := fmt.Sprintf("nodes/%s/apt/update", node)
	if err := c.Get(ctx, path, &updates); err != nil {
		return nil, fmt.Errorf("获取软件包更新列表失败: %w", err)
	}
	return updates, nil
}

// NodeUpdatePackages 更新软件包
// node: 节点名称, packages: 要更新的包列表（逗号分隔）
// 返回异步任务 ID (UPID)
func (c *Client) NodeUpdatePackages(ctx context.Context, node string, packages string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/apt/update", node)
	params := map[string]interface{}{"packages": packages}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新软件包失败: %w", err)
	}
	return upid, nil
}

// GetNodeDNS 获取节点 DNS 配置
// node: 节点名称
// 返回 DNS 搜索域和名称服务器配置
func (c *Client) GetNodeDNS(ctx context.Context, node string) (*DNSConfig, error) {
	var dns DNSConfig
	path := fmt.Sprintf("nodes/%s/dns", node)
	if err := c.Get(ctx, path, &dns); err != nil {
		return nil, fmt.Errorf("获取 DNS 配置失败: %w", err)
	}
	return &dns, nil
}

// SetNodeDNS 更新节点 DNS 配置
// node: 节点名称, config: DNS 配置
// 返回异步任务 ID (UPID)
func (c *Client) SetNodeDNS(ctx context.Context, node string, config DNSConfig) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/dns", node)
	if err := c.Put(ctx, path, config, &upid); err != nil {
		return "", fmt.Errorf("更新 DNS 配置失败: %w", err)
	}
	return upid, nil
}

// GetNodeTime 获取节点时间信息
// node: 节点名称
// 返回节点的时区和时间信息
func (c *Client) GetNodeTime(ctx context.Context, node string) (*TimeInfo, error) {
	var timeInfo TimeInfo
	path := fmt.Sprintf("nodes/%s/time", node)
	if err := c.Get(ctx, path, &timeInfo); err != nil {
		return nil, fmt.Errorf("获取时间信息失败: %w", err)
	}
	return &timeInfo, nil
}

// SetNodeTimezone 设置节点时区
// node: 节点名称, timezone: 时区（如 Asia/Shanghai）
// 返回异步任务 ID (UPID)
func (c *Client) SetNodeTimezone(ctx context.Context, node, timezone string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/time", node)
	params := map[string]interface{}{"timezone": timezone}
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("设置时区失败: %w", err)
	}
	return upid, nil
}

// GetNodeRRD 获取节点 RRD 性能数据
// node: 节点名称, timeframe: 时间范围 (hour, day, week, month, year)
// dataset: 数据集 (cpu, memory, network, disk, loadavg)
func (c *Client) GetNodeRRD(ctx context.Context, node string, timeframe, dataset string) ([]RRDPoint, error) {
	var data []RRDPoint
	path := fmt.Sprintf("nodes/%s/rrd", node)
	params := url.Values{}
	params.Set("timeframe", timeframe)
	if dataset != "" {
		params.Set("ds", dataset)
	}
	if err := c.GetWithParams(ctx, path, params, &data); err != nil {
		return nil, fmt.Errorf("获取节点 RRD 数据失败: %w", err)
	}
	return data, nil
}

// GetNodeReport 获取节点诊断报告
// node: 节点名称
// 返回节点诊断报告数据
func (c *Client) GetNodeReport(ctx context.Context, node string) ([]byte, error) {
	resp, err := c.Do(ctx, "GET", fmt.Sprintf("nodes/%s/report", node), nil)
	if err != nil {
		return nil, fmt.Errorf("获取节点报告失败: %w", err)
	}
	data, ok := resp.Data.(string)
	if !ok {
		return nil, fmt.Errorf("节点报告数据格式错误")
	}
	return []byte(data), nil
}

// RebootNode 重启节点
// node: 节点名称
// 返回异步任务 ID (UPID)
func (c *Client) RebootNode(ctx context.Context, node string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/reboot", node)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("重启节点失败: %w", err)
	}
	return upid, nil
}

// ShutdownNode 关闭节点
// node: 节点名称
// 返回异步任务 ID (UPID)
func (c *Client) ShutdownNode(ctx context.Context, node string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/shutdown", node)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("关闭节点失败: %w", err)
	}
	return upid, nil
}
