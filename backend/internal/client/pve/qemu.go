package pve

import (
	"context"
	"fmt"
	"net/url"
)

// ListQEMU 获取指定节点的 QEMU 虚拟机列表
// node: 节点名称
// 返回该节点上所有虚拟机的信息
func (c *Client) ListQEMU(ctx context.Context, node string) ([]QEMUVM, error) {
	var vms []QEMUVM
	path := fmt.Sprintf("nodes/%s/qemu", node)
	if err := c.Get(ctx, path, &vms); err != nil {
		return nil, fmt.Errorf("获取 QEMU 列表失败: %w", err)
	}
	return vms, nil
}

// GetQEMUConfig 获取指定虚拟机的完整配置
// node: 节点名称, vmid: 虚拟机 ID
// 返回虚拟机的详细配置参数
func (c *Client) GetQEMUConfig(ctx context.Context, node string, vmid int) (map[string]interface{}, error) {
	var config map[string]interface{}
	path := fmt.Sprintf("nodes/%s/qemu/%d/config", node, vmid)
	if err := c.Get(ctx, path, &config); err != nil {
		return nil, fmt.Errorf("获取 QEMU 配置失败: %w", err)
	}
	return config, nil
}

// CreateQEMU 创建新的 QEMU 虚拟机
// node: 节点名称, params: 创建参数
// 返回异步任务 ID (UPID)
func (c *Client) CreateQEMU(ctx context.Context, node string, params *QEMUCreateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu", node)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建 QEMU 失败: %w", err)
	}
	return upid, nil
}

// StartQEMU 启动虚拟机
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) StartQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/start", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("启动 QEMU 失败: %w", err)
	}
	return upid, nil
}

// StopQEMU 停止虚拟机（强制关机）
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) StopQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/stop", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("停止 QEMU 失败: %w", err)
	}
	return upid, nil
}

// ShutdownQEMU 关闭虚拟机（优雅关机，需要 QEMU guest agent）
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) ShutdownQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/shutdown", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("关闭 QEMU 失败: %w", err)
	}
	return upid, nil
}

// RebootQEMU 重启虚拟机
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) RebootQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/reboot", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("重启 QEMU 失败: %w", err)
	}
	return upid, nil
}

// SuspendQEMU 挂起虚拟机（保存状态到磁盘）
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) SuspendQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/suspend", node, vmid)
	params := map[string]interface{}{"todisk": 1}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("挂起 QEMU 失败: %w", err)
	}
	return upid, nil
}

// ResumeQEMU 恢复虚拟机（从挂起状态恢复）
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) ResumeQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/resume", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("恢复 QEMU 失败: %w", err)
	}
	return upid, nil
}

// ResetQEMU 重置虚拟机（相当于断电重启）
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) ResetQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/reset", node, vmid)
	if err := c.Post(ctx, path, nil, &upid); err != nil {
		return "", fmt.Errorf("重置 QEMU 失败: %w", err)
	}
	return upid, nil
}

// SetQEMUConfig 更新虚拟机配置
// node: 节点名称, vmid: 虚拟机 ID, config: 配置参数映射
// 返回异步任务 ID (UPID)
func (c *Client) SetQEMUConfig(ctx context.Context, node string, vmid int, config QEMUConfigParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/config", node, vmid)
	if err := c.Put(ctx, path, config, &upid); err != nil {
		return "", fmt.Errorf("更新 QEMU 配置失败: %w", err)
	}
	return upid, nil
}

// ListQEMUSnapshots 获取虚拟机快照列表
// node: 节点名称, vmid: 虚拟机 ID
// 返回快照信息列表
func (c *Client) ListQEMUSnapshots(ctx context.Context, node string, vmid int) ([]Snapshot, error) {
	var snapshots []Snapshot
	path := fmt.Sprintf("nodes/%s/qemu/%d/snapshot", node, vmid)
	if err := c.Get(ctx, path, &snapshots); err != nil {
		return nil, fmt.Errorf("获取 QEMU 快照列表失败: %w", err)
	}
	return snapshots, nil
}

// CreateQEMUSnapshot 创建虚拟机快照
// node: 节点名称, vmid: 虚拟机 ID, name: 快照名称, description: 描述
// 返回异步任务 ID (UPID)
func (c *Client) CreateQEMUSnapshot(ctx context.Context, node string, vmid int, name, description string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/snapshot", node, vmid)
	params := url.Values{}
	params.Set("snapname", name)
	if description != "" {
		params.Set("description", description)
	}
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("创建 QEMU 快照失败: %w", err)
	}
	return upid, nil
}

// DeleteQEMUSnapshot 删除虚拟机快照
// node: 节点名称, vmid: 虚拟机 ID, snapname: 快照名称
// 返回异步任务 ID (UPID)
func (c *Client) DeleteQEMUSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/snapshot/%s", node, vmid, snapname)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除 QEMU 快照失败: %w", err)
	}
	return upid, nil
}

// CloneQEMU 克隆虚拟机
// node: 节点名称, vmid: 虚拟机 ID, params: 克隆参数
// 返回异步任务 ID (UPID)
func (c *Client) CloneQEMU(ctx context.Context, node string, vmid int, params *QEMUCloneParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/clone", node, vmid)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("克隆 QEMU 失败: %w", err)
	}
	return upid, nil
}

// MigrateQEMU 迁移虚拟机到其他节点
// node: 节点名称, vmid: 虚拟机 ID, target: 目标节点名称
// 返回异步任务 ID (UPID)
func (c *Client) MigrateQEMU(ctx context.Context, node string, vmid int, params *QEMUMigrateParams) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d/migrate", node, vmid)
	if err := c.Post(ctx, path, params, &upid); err != nil {
		return "", fmt.Errorf("迁移 QEMU 失败: %w", err)
	}
	return upid, nil
}

// GetQEMURRD 获取虚拟机 RRD 性能数据
// node: 节点名称, vmid: 虚拟机 ID, timeframe: 时间范围 (hour, day, week, month, year)
// dataset: 数据集 (cpu, memory, network, disk)
func (c *Client) GetQEMURRD(ctx context.Context, node string, vmid int, timeframe, dataset string) ([]RRDPoint, error) {
	var data []RRDPoint
	path := fmt.Sprintf("nodes/%s/qemu/%d/rrd", node, vmid)
	params := url.Values{}
	params.Set("timeframe", timeframe)
	if dataset != "" {
		params.Set("ds", dataset)
	}
	if err := c.GetWithParams(ctx, path, params, &data); err != nil {
		return nil, fmt.Errorf("获取 QEMU RRD 数据失败: %w", err)
	}
	return data, nil
}

// GetQEMUPending 获取虚拟机待处理配置
// node: 节点名称, vmid: 虚拟机 ID
// 返回等待重启后生效的配置变更
func (c *Client) GetQEMUPending(ctx context.Context, node string, vmid int) ([]PendingConfig, error) {
	var pending []PendingConfig
	path := fmt.Sprintf("nodes/%s/qemu/%d/pending", node, vmid)
	if err := c.Get(ctx, path, &pending); err != nil {
		return nil, fmt.Errorf("获取 QEMU 待处理配置失败: %w", err)
	}
	return pending, nil
}

// DeleteQEMU 删除虚拟机
// node: 节点名称, vmid: 虚拟机 ID
// 返回异步任务 ID (UPID)
func (c *Client) DeleteQEMU(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := fmt.Sprintf("nodes/%s/qemu/%d", node, vmid)
	if err := c.Delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("删除 QEMU 失败: %w", err)
	}
	return upid, nil
}

// GetQEMUCurrent 获取虚拟机当前状态
// node: 节点名称, vmid: 虚拟机 ID
// 返回虚拟机当前运行状态
func (c *Client) GetQEMUCurrent(ctx context.Context, node string, vmid int) (*QEMUVM, error) {
	var vm QEMUVM
	path := fmt.Sprintf("nodes/%s/qemu/%d/status/current", node, vmid)
	if err := c.Get(ctx, path, &vm); err != nil {
		return nil, fmt.Errorf("获取 QEMU 当前状态失败: %w", err)
	}
	return &vm, nil
}
