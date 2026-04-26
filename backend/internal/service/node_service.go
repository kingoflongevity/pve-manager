package service

import (
	"context"
	"time"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// NodeService 节点管理服务
// 负责节点状态、服务、网络、DNS、时间、软件包更新等节点级操作
type NodeService struct {
	logger *zap.Logger
}

// NewNodeService 创建节点服务实例
func NewNodeService(logger *zap.Logger) *NodeService {
	return &NodeService{logger: logger}
}

// GetStatus 获取节点状态
func (s *NodeService) GetStatus(ctx context.Context, client *pve.Client, node string) (*pve.NodeStatus, error) {
	return client.GetNodeStatus(ctx, node)
}

// GetVersion 获取节点 PVE 版本
func (s *NodeService) GetVersion(ctx context.Context, client *pve.Client, node string) (*pve.VersionInfo, error) {
	return client.GetNodeVersion(ctx, node)
}

// GetServices 获取节点服务列表
func (s *NodeService) GetServices(ctx context.Context, client *pve.Client, node string) ([]pve.Service, error) {
	return client.GetNodeServices(ctx, node)
}

// GetSyslog 获取节点系统日志
func (s *NodeService) GetSyslog(ctx context.Context, client *pve.Client, node string, limit, start int) ([]pve.LogEntry, error) {
	return client.GetNodeSyslog(ctx, node, limit, start)
}

// GetTasks 获取节点任务列表
func (s *NodeService) GetTasks(ctx context.Context, client *pve.Client, node string) ([]pve.Task, error) {
	return client.GetNodeTasks(ctx, node)
}

// GetTaskStatus 获取任务状态
func (s *NodeService) GetTaskStatus(ctx context.Context, client *pve.Client, node, upid string) (*pve.TaskStatus, error) {
	return client.GetNodeTaskStatus(ctx, node, upid)
}

// GetTaskLog 获取任务日志
func (s *NodeService) GetTaskLog(ctx context.Context, client *pve.Client, node, upid string) ([]pve.TaskLogLine, error) {
	return client.GetNodeTaskLog(ctx, node, upid)
}

// WaitForTask 等待任务完成（轮询）
func (s *NodeService) WaitForTask(ctx context.Context, client *pve.Client, node, upid string, timeout time.Duration) (*pve.TaskStatus, error) {
	return client.WaitForTask(ctx, node, upid, timeout)
}

// GetNetwork 获取节点网络接口列表
func (s *NodeService) GetNetwork(ctx context.Context, client *pve.Client, node string) ([]pve.NetInterface, error) {
	return client.GetNodeNetwork(ctx, node)
}

// GetDNS 获取节点 DNS 配置
func (s *NodeService) GetDNS(ctx context.Context, client *pve.Client, node string) (*pve.DNSConfig, error) {
	return client.GetNodeDNS(ctx, node)
}

// GetTime 获取节点时间信息
func (s *NodeService) GetTime(ctx context.Context, client *pve.Client, node string) (*pve.TimeInfo, error) {
	return client.GetNodeTime(ctx, node)
}

// GetAPTUpdate 获取可更新的软件包列表
func (s *NodeService) GetAPTUpdate(ctx context.Context, client *pve.Client, node string) ([]pve.PackageUpdate, error) {
	return client.GetNodeAPTUpdate(ctx, node)
}

// GetRRD 获取节点 RRD 性能数据
func (s *NodeService) GetRRD(ctx context.Context, client *pve.Client, node string, timeframe, dataset string) ([]pve.RRDPoint, error) {
	return client.GetNodeRRD(ctx, node, timeframe, dataset)
}

// GetList 获取所有节点列表
func (s *NodeService) GetList(ctx context.Context, client *pve.Client) ([]pve.NodeInfo, error) {
	return client.GetNodes(ctx)
}
