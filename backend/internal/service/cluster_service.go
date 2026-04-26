package service

import (
	"context"

	"github.com/kingoflongevity/pve-manager/backend/internal/client/pve"
	"go.uber.org/zap"
)

// ClusterService 集群服务
// 负责集群资源、任务、HA、SDN、资源池等集群级操作
type ClusterService struct {
	logger *zap.Logger
}

// NewClusterService 创建集群服务实例
func NewClusterService(logger *zap.Logger) *ClusterService {
	return &ClusterService{logger: logger}
}

// GetResources 获取集群所有资源
func (s *ClusterService) GetResources(ctx context.Context, client *pve.Client) ([]pve.ClusterResource, error) {
	return client.GetClusterResources(ctx)
}

// GetResourcesByType 按类型获取集群资源
func (s *ClusterService) GetResourcesByType(ctx context.Context, client *pve.Client, resourceType string) ([]pve.ClusterResource, error) {
	return client.GetClusterResourcesByType(ctx, resourceType)
}

// GetTasks 获取集群任务列表
func (s *ClusterService) GetTasks(ctx context.Context, client *pve.Client) ([]pve.ClusterTask, error) {
	return client.GetClusterTasks(ctx)
}

// GetNextID 获取下一个可用的 VM ID
func (s *ClusterService) GetNextID(ctx context.Context, client *pve.Client) (*pve.NextVMID, error) {
	return client.GetNextID(ctx)
}

// GetHAConfig 获取 HA 配置
func (s *ClusterService) GetHAConfig(ctx context.Context, client *pve.Client) (*pve.HAConfig, error) {
	return client.GetHAConfig(ctx)
}

// GetSDNZones 获取 SDN 区域列表
func (s *ClusterService) GetSDNZones(ctx context.Context, client *pve.Client) ([]pve.SDNZone, error) {
	return client.GetSDNZones(ctx)
}

// GetSDNVNETs 获取 SDN 虚拟网络列表
func (s *ClusterService) GetSDNVNETs(ctx context.Context, client *pve.Client) ([]pve.SDNVNET, error) {
	return client.GetSDNVNETs(ctx)
}

// ListPools 获取资源池列表
func (s *ClusterService) ListPools(ctx context.Context, client *pve.Client) ([]pve.Pool, error) {
	return client.ListPools(ctx)
}

// GetPool 获取资源池详情
func (s *ClusterService) GetPool(ctx context.Context, client *pve.Client, poolid string) (*pve.PoolDetail, error) {
	return client.GetPool(ctx, poolid)
}
