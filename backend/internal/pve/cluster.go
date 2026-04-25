package pve

import (
	"context"
	"fmt"
	"net/url"
)

// GetClusterResources 获取集群所有资源
// 返回集群中所有资源（节点、VM、容器、存储）的列表
func (c *Client) GetClusterResources(ctx context.Context) ([]ClusterResource, error) {
	var resources []ClusterResource
	if err := c.Get(ctx, "cluster/resources", &resources); err != nil {
		return nil, fmt.Errorf("获取集群资源失败: %w", err)
	}
	return resources, nil
}

// GetClusterResourcesByType 按类型获取集群资源
// resourceType: 资源类型 (vm, storage, node)
// 返回指定类型的资源列表
func (c *Client) GetClusterResourcesByType(ctx context.Context, resourceType string) ([]ClusterResource, error) {
	var resources []ClusterResource
	params := url.Values{}
	if resourceType != "" {
		params.Set("type", resourceType)
	}
	if err := c.GetWithParams(ctx, "cluster/resources", params, &resources); err != nil {
		return nil, fmt.Errorf("获取集群资源失败: %w", err)
	}
	return resources, nil
}

// GetClusterTasks 获取集群任务列表
// 返回集群中所有节点的任务
func (c *Client) GetClusterTasks(ctx context.Context) ([]ClusterTask, error) {
	var tasks []ClusterTask
	if err := c.Get(ctx, "cluster/tasks", &tasks); err != nil {
		return nil, fmt.Errorf("获取集群任务失败: %w", err)
	}
	return tasks, nil
}

// GetNextID 获取下一个可用的 VM ID
// 返回可用于创建新虚拟机或容器的 ID
func (c *Client) GetNextID(ctx context.Context) (*NextVMID, error) {
	var nextID NextVMID
	if err := c.Get(ctx, "cluster/nextid", &nextID); err != nil {
		return nil, fmt.Errorf("获取下一个 VM ID 失败: %w", err)
	}
	return &nextID, nil
}

// GetHAConfig 获取 HA 配置
// 返回高可用配置（组、资源、隔离设备）
func (c *Client) GetHAConfig(ctx context.Context) (*HAConfig, error) {
	var config HAConfig
	if err := c.Get(ctx, "cluster/ha/resources", &config); err != nil {
		return nil, fmt.Errorf("获取 HA 配置失败: %w", err)
	}
	return &config, nil
}

// GetHAGroups 获取 HA 组列表
// 返回所有高可用组配置
func (c *Client) GetHAGroups(ctx context.Context) ([]HAGroup, error) {
	var groups []HAGroup
	if err := c.Get(ctx, "cluster/ha/groups", &groups); err != nil {
		return nil, fmt.Errorf("获取 HA 组列表失败: %w", err)
	}
	return groups, nil
}

// GetHAResources 获取 HA 资源列表
// 返回所有高可用资源配置
func (c *Client) GetHAResources(ctx context.Context) ([]HAResource, error) {
	var resources []HAResource
	if err := c.Get(ctx, "cluster/ha/resources", &resources); err != nil {
		return nil, fmt.Errorf("获取 HA 资源列表失败: %w", err)
	}
	return resources, nil
}

// GetSDNZones 获取 SDN 区域列表
// 返回所有软件定义网络区域
func (c *Client) GetSDNZones(ctx context.Context) ([]SDNZone, error) {
	var zones []SDNZone
	if err := c.Get(ctx, "cluster/sdn/zones", &zones); err != nil {
		return nil, fmt.Errorf("获取 SDN 区域失败: %w", err)
	}
	return zones, nil
}

// GetSDNVNETs 获取 SDN 虚拟网络列表
// 返回所有 SDN 虚拟网络
func (c *Client) GetSDNVNETs(ctx context.Context) ([]SDNVNET, error) {
	var vnets []SDNVNET
	if err := c.Get(ctx, "cluster/sdn/vnets", &vnets); err != nil {
		return nil, fmt.Errorf("获取 SDN 虚拟网络失败: %w", err)
	}
	return vnets, nil
}

// ListPools 获取所有资源池列表
// 返回所有资源池的简要信息
func (c *Client) ListPools(ctx context.Context) ([]Pool, error) {
	var pools []Pool
	if err := c.Get(ctx, "pools", &pools); err != nil {
		return nil, fmt.Errorf("获取资源池列表失败: %w", err)
	}
	return pools, nil
}

// GetPool 获取指定资源池的详细信息
// poolid: 资源池 ID
// 返回资源池的成员和配置
func (c *Client) GetPool(ctx context.Context, poolid string) (*PoolDetail, error) {
	var detail PoolDetail
	path := fmt.Sprintf("pools/%s", poolid)
	if err := c.Get(ctx, path, &detail); err != nil {
		return nil, fmt.Errorf("获取资源池详情失败: %w", err)
	}
	return &detail, nil
}

// CreatePool 创建资源池
// poolid: 资源池 ID, comment: 描述
// 返回异步任务 ID (UPID)
func (c *Client) CreatePool(ctx context.Context, poolid, comment string) (string, error) {
	var upid string
	params := map[string]interface{}{"poolid": poolid}
	if comment != "" {
		params["comment"] = comment
	}
	if err := c.Post(ctx, "pools", params, &upid); err != nil {
		return "", fmt.Errorf("创建资源池失败: %w", err)
	}
	return upid, nil
}

// UpdatePool 更新资源池
// poolid: 资源池 ID, params: 更新参数
// 返回异步任务 ID (UPID)
func (c *Client) UpdatePool(ctx context.Context, poolid string, params map[string]interface{}) (string, error) {
	var upid string
	path := fmt.Sprintf("pools/%s", poolid)
	if err := c.Put(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("更新资源池失败: %w", err)
	}
	return upid, nil
}

// DeletePool 删除资源池
// poolid: 资源池 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeletePool(ctx context.Context, poolid string) (string, error) {
	var upid string
	path := fmt.Sprintf("pools/%s", poolid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除资源池失败: %w", err)
	}
	return upid, nil
}

// GetClusterLog 获取集群日志
// 返回集群级别的日志
func (c *Client) GetClusterLog(ctx context.Context) ([]LogEntry, error) {
	var logs []LogEntry
	if err := c.Get(ctx, "cluster/log", &logs); err != nil {
		return nil, fmt.Errorf("获取集群日志失败: %w", err)
	}
	return logs, nil
}

// GetClusterOptions 获取集群选项配置
// 返回集群的全局配置选项
func (c *Client) GetClusterOptions(ctx context.Context) (map[string]interface{}, error) {
	var options map[string]interface{}
	if err := c.Get(ctx, "cluster/config", &options); err != nil {
		return nil, fmt.Errorf("获取集群配置失败: %w", err)
	}
	return options, nil
}
